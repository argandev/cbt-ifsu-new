// Package master — Master Data: Jurusan, Kelas, Mapel, Guru, Siswa,
// Kompetensi, Kisi-kisi + Import Siswa dari CSV.
//
// Endpoint (prefix /api/v1/master):
//
//	Jurusan, Kelas, Mapel, Guru, Siswa: GET / POST, GET / PUT / DELETE /:id
//	Kompetensi, Kisi-kisi: GET / POST, GET / PUT / DELETE /:id
//	POST /siswa/import (multipart CSV)
//
// Permission (diperiksa per-handler via audit.HasPermission, pola Tim 5;
// CSRF + auth sesi ditangani middleware global Tim 1):
//
//	master.<entitas>.read / .create / .update / .delete, master.siswa.import
//	Kompetensi & Kisi-kisi memakai permission master.mapel.* (keduanya
//	sub-master mapel; tidak ada kode permission sendiri di seed 57).
//
// Aturan penting:
//   - Full CRUD termasuk EDIT (PUT). Soft-delete TIDAK dipakai: DELETE
//     hard, ditolak 409 bila data masih dipakai (kelas/mapel/siswa/
//     bank soal/soal; guard jadwal (Tim 4, migrasi 0004) menyusul).
//   - Mapel Produktif wajib punya jurusan_id; Mapel Umum wajib null
//     (ditegakkan juga oleh CHECK di migrasi 0002).
//   - Import siswa membuat akun login otomatis: username = NISN,
//     password awal = NISN (di-hash Argon2id format PHC), role siswa,
//     must_change_password = TRUE (siswa WAJIB ganti saat login pertama).
//     POST /siswa manual memakai alur akun yang sama bila user_id kosong.
//   - Akun siswa dibuat di tabel users dengan kolom role='siswa'
//     (satu role per user, migrasi 0001 Tim 1). Kontrak kolom ada di
//     backend/db/queries/master.sql.
//   - Envelope: sukses {"data":...} / {"data":...,"meta":{...}};
//     error {"error":{"code":"...","message":"..."}}.
//   - SQL di handler ini sama persis dengan backend/db/queries/master.sql
//     (pola Tim 5), sampai kode hasil `sqlc generate` dipakai.
package master

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"golang.org/x/crypto/argon2"
)

// ---------------------------------------------------------------------
// Helper umum
// ---------------------------------------------------------------------

func writeError(c *echo.Context, status int, code, message string) error {
	return c.JSON(status, map[string]any{
		"error": map[string]any{"code": code, "message": message},
	})
}

func writeOK(c *echo.Context, data any) error {
	return c.JSON(http.StatusOK, map[string]any{"data": data})
}

func writeOKStatus(c *echo.Context, status int, data any) error {
	return c.JSON(status, map[string]any{"data": data})
}

type pageMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func writeList(c *echo.Context, items any, page, perPage int, total int64) error {
	totalPages := 0
	if perPage > 0 {
		totalPages = int((total + int64(perPage) - 1) / int64(perPage))
	}
	return c.JSON(http.StatusOK, map[string]any{
		"data": items,
		"meta": pageMeta{Page: page, PerPage: perPage, Total: total, TotalPages: totalPages},
	})
}

// parsePagination membaca ?page= & ?per_page= (default 1 / 20, maks 100).
func parsePagination(c *echo.Context) (page, perPage, offset int) {
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page < 1 {
		page = 1
	}
	perPage, err = strconv.Atoi(c.QueryParam("per_page"))
	if err != nil || perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	return page, perPage, (page - 1) * perPage
}

// searchParam mengubah ?q= menjadi *string (nil = tanpa filter).
func searchParam(c *echo.Context) *string {
	q := strings.TrimSpace(c.QueryParam("q"))
	if q == "" {
		return nil
	}
	return &q
}

// optUUID mengubah param query kosong menjadi nil (tanpa filter).
func optUUID(c *echo.Context, name string) *string {
	v := strings.TrimSpace(c.QueryParam(name))
	if v == "" {
		return nil
	}
	return &v
}

// optInt mengubah param query kosong menjadi nil.
func optInt(c *echo.Context, name string) *int {
	v := strings.TrimSpace(c.QueryParam(name))
	if v == "" {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return nil
	}
	return &n
}

// optBool mengubah param query kosong menjadi nil.
func optBool(c *echo.Context, name string) *bool {
	v := strings.TrimSpace(c.QueryParam(name))
	if v == "" {
		return nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return nil
	}
	return &b
}

// validUUID memeriksa format UUID (36 karakter, hex + strip).
func validUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, r := range s {
		switch {
		case i == 8 || i == 13 || i == 18 || i == 23:
			if r != '-' {
				return false
			}
		default:
			if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f' || r >= 'A' && r <= 'F') {
				return false
			}
		}
	}
	return true
}

// isUniqueViolation melaporkan apakah err adalah pelanggaran UNIQUE (23505)
// beserta nama constraint-nya.
func uniqueConstraint(err error) (string, bool) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return pgErr.ConstraintName, true
	}
	return "", false
}

// conflictMessage memetakan constraint UNIQUE ke pesan Bahasa Indonesia.
func conflictMessage(constraint string) string {
	switch constraint {
	case "jurusan_kode_key":
		return "Kode jurusan sudah dipakai."
	case "mapel_kode_key":
		return "Kode mapel sudah dipakai."
	case "siswa_nisn_key":
		return "NISN sudah terdaftar."
	case "guru_nip_key":
		return "NIP sudah dipakai guru lain."
	case "kompetensi_mapel_id_kode_key":
		return "Kode kompetensi sudah dipakai untuk mapel ini."
	case "kelas_jurusan_id_tingkat_nama_tahun_ajaran_key":
		return "Kelas dengan nama, tingkat, dan tahun ajaran tersebut sudah ada di jurusan ini."
	case "users_username_key":
		return "Username sudah dipakai."
	default:
		return "Data duplikat: nilai unik sudah dipakai."
	}
}

// countOrZero menghitung COUNT(*); bila tabel rujukan belum ada
// (migrasi tim lain belum di-apply) dianggap 0 agar guard tidak
// menggagalkan operasi. Dipakai HANYA untuk guard hapus lintas tim.
func countOrZero(ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) int64 {
	var n int64
	if err := pool.QueryRow(ctx, sql, args...).Scan(&n); err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return 0
		}
		return -1 // penanda error selain tabel tidak ada
	}
	return n
}

// catat mencatat audit log; error diabaikan (tidak boleh menggagalkan
// operasi utama).
func catat(c *echo.Context, pool *pgxpool.Pool, aksi, entitas, entitasID string, detail any) {
	userID, _, _, _, _ := audit.CurrentUser(c)
	_ = audit.Catat(c.Request().Context(), pool, userID, aksi, entitas, entitasID, detail, c.RealIP())
}

// ---------------------------------------------------------------------
// Password siswa: Argon2id, format PHC standar.
// Tim 1 (auth) HARUS memverifikasi dengan parser PHC yang sama.
// Parameter: m=64 MiB, t=3, p=2, salt 16 byte, hash 32 byte.
// ---------------------------------------------------------------------

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 2, 32)
	b64 := base64.RawStdEncoding
	return fmt.Sprintf("$argon2id$v=19$m=65536,t=3,p=2$%s$%s",
		b64.EncodeToString(salt), b64.EncodeToString(hash)), nil
}

// ---------------------------------------------------------------------
// DTO
// ---------------------------------------------------------------------

type jurusanDTO struct {
	ID        string    `json:"id"`
	Kode      string    `json:"kode"`
	Nama      string    `json:"nama"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type kelasDTO struct {
	ID          string    `json:"id"`
	JurusanID   string    `json:"jurusan_id"`
	Tingkat     int32     `json:"tingkat"`
	Nama        string    `json:"nama"`
	TahunAjaran string    `json:"tahun_ajaran"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type mapelDTO struct {
	ID           string    `json:"id"`
	Kode         string    `json:"kode"`
	Nama         string    `json:"nama"`
	Jenis        string    `json:"jenis"`
	JurusanID    *string   `json:"jurusan_id"`
	KelasTingkat *int32    `json:"kelas_tingkat"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type guruDTO struct {
	ID        string          `json:"id"`
	UserID    *string         `json:"user_id"`
	NIP       *string         `json:"nip"`
	Nama      string          `json:"nama"`
	MapelIDs  json.RawMessage `json:"mapel_ids"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type siswaDTO struct {
	ID         string    `json:"id"`
	UserID     *string   `json:"user_id"`
	NISN       string    `json:"nisn"`
	NIS        *string   `json:"nis"`
	Nama       string    `json:"nama"`
	KelasID    string    `json:"kelas_id"`
	JurusanID  string    `json:"jurusan_id"`
	TahunMasuk *int32    `json:"tahun_masuk"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type kompetensiDTO struct {
	ID        string    `json:"id"`
	MapelID   string    `json:"mapel_id"`
	Kode      string    `json:"kode"`
	Deskripsi string    `json:"deskripsi"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type kisiKisiDTO struct {
	ID           string    `json:"id"`
	MapelID      string    `json:"mapel_id"`
	KelasTingkat int32     `json:"kelas_tingkat"`
	Kelompok     string    `json:"kelompok"`
	KompetensiID *string   `json:"kompetensi_id"`
	JumlahSoal   int32     `json:"jumlah_soal"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ---------------------------------------------------------------------
// Jurusan
// ---------------------------------------------------------------------

func listJurusan(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.jurusan.read") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat jurusan.")
	}
	page, perPage, offset := parsePagination(c)
	q := searchParam(c)
	ctx := c.Request().Context()
	var total int64
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM jurusan
		 WHERE ($1::text IS NULL OR nama ILIKE '%' || $1 || '%' OR kode ILIKE '%' || $1 || '%')`, q).Scan(&total); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca jurusan.")
	}
	rows, err := pool.Query(ctx,
		`SELECT id, kode, nama, created_at, updated_at FROM jurusan
		 WHERE ($1::text IS NULL OR nama ILIKE '%' || $1 || '%' OR kode ILIKE '%' || $1 || '%')
		 ORDER BY nama ASC LIMIT $2 OFFSET $3`, q, perPage, offset)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca jurusan.")
	}
	defer rows.Close()
	items := []jurusanDTO{}
	for rows.Next() {
		var d jurusanDTO
		if err := rows.Scan(&d.ID, &d.Kode, &d.Nama, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca jurusan.")
		}
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca jurusan.")
	}
	return writeList(c, items, page, perPage, total)
}

func getJurusan(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.jurusan.read") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat jurusan.")
	}
	id := c.Param("id")
	if !validUUID(id) {
		return writeError(c, http.StatusNotFound, "not_found", "Jurusan tidak ditemukan.")
	}
	var d jurusanDTO
	err := pool.QueryRow(c.Request().Context(),
		`SELECT id, kode, nama, created_at, updated_at FROM jurusan WHERE id = $1`, id).
		Scan(&d.ID, &d.Kode, &d.Nama, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return writeError(c, http.StatusNotFound, "not_found", "Jurusan tidak ditemukan.")
	}
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca jurusan.")
	}
	return writeOK(c, d)
}

type jurusanInput struct {
	Kode string `json:"kode"`
	Nama string `json:"nama"`
}

func (in jurusanInput) validate() string {
	if strings.TrimSpace(in.Kode) == "" {
		return "Kode jurusan wajib diisi."
	}
	if strings.TrimSpace(in.Nama) == "" {
		return "Nama jurusan wajib diisi."
	}
	return ""
}

func createJurusan(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.jurusan.create") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menambah jurusan.")
	}
	var in jurusanInput
	if err := c.Bind(&in); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	if msg := in.validate(); msg != "" {
		return writeError(c, http.StatusUnprocessableEntity, "validation_error", msg)
	}
	var d jurusanDTO
	err := pool.QueryRow(c.Request().Context(),
		`INSERT INTO jurusan (kode, nama) VALUES ($1, $2)
		 RETURNING id, kode, nama, created_at, updated_at`,
		strings.TrimSpace(in.Kode), strings.TrimSpace(in.Nama)).
		Scan(&d.ID, &d.Kode, &d.Nama, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			return writeError(c, http.StatusConflict, "conflict", conflictMessage(constraint))
		}
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menambah jurusan.")
	}
	catat(c, pool, "create_jurusan", "jurusan", d.ID, map[string]any{"kode": d.Kode, "nama": d.Nama})
	return writeOKStatus(c, http.StatusCreated, d)
}

func updateJurusan(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.jurusan.update") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mengubah jurusan.")
	}
	id := c.Param("id")
	if !validUUID(id) {
		return writeError(c, http.StatusNotFound, "not_found", "Jurusan tidak ditemukan.")
	}
	var in jurusanInput
	if err := c.Bind(&in); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	if msg := in.validate(); msg != "" {
		return writeError(c, http.StatusUnprocessableEntity, "validation_error", msg)
	}
	var d jurusanDTO
	err := pool.QueryRow(c.Request().Context(),
		`UPDATE jurusan SET kode = $2, nama = $3, updated_at = now()
		 WHERE id = $1 RETURNING id, kode, nama, created_at, updated_at`,
		id, strings.TrimSpace(in.Kode), strings.TrimSpace(in.Nama)).
		Scan(&d.ID, &d.Kode, &d.Nama, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return writeError(c, http.StatusNotFound, "not_found", "Jurusan tidak ditemukan.")
	}
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			return writeError(c, http.StatusConflict, "conflict", conflictMessage(constraint))
		}
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mengubah jurusan.")
	}
	catat(c, pool, "update_jurusan", "jurusan", d.ID, map[string]any{"kode": d.Kode, "nama": d.Nama})
	return writeOK(c, d)
}

func deleteJurusan(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.jurusan.delete") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menghapus jurusan.")
	}
	id := c.Param("id")
	if !validUUID(id) {
		return writeError(c, http.StatusNotFound, "not_found", "Jurusan tidak ditemukan.")
	}
	ctx := c.Request().Context()
	if n := countOrZero(ctx, pool, `SELECT COUNT(*) FROM kelas WHERE jurusan_id = $1`, id); n > 0 {
		return writeError(c, http.StatusConflict, "conflict",
			fmt.Sprintf("Jurusan masih dipakai %d kelas. Pindahkan/hapus kelas dulu.", n))
	} else if n < 0 {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa pemakaian jurusan.")
	}
	if n := countOrZero(ctx, pool, `SELECT COUNT(*) FROM mapel WHERE jurusan_id = $1`, id); n > 0 {
		return writeError(c, http.StatusConflict, "conflict",
			fmt.Sprintf("Jurusan masih dipakai %d mapel produktif. Hapus/ubah mapel dulu.", n))
	} else if n < 0 {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa pemakaian jurusan.")
	}
	if n := countOrZero(ctx, pool, `SELECT COUNT(*) FROM siswa WHERE jurusan_id = $1`, id); n > 0 {
		return writeError(c, http.StatusConflict, "conflict",
			fmt.Sprintf("Jurusan masih dipakai %d siswa. Pindahkan siswa dulu.", n))
	} else if n < 0 {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa pemakaian jurusan.")
	}
	ct, err := pool.Exec(ctx, `DELETE FROM jurusan WHERE id = $1`, id)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghapus jurusan.")
	}
	if ct.RowsAffected() == 0 {
		return writeError(c, http.StatusNotFound, "not_found", "Jurusan tidak ditemukan.")
	}
	catat(c, pool, "delete_jurusan", "jurusan", id, nil)
	return writeOK(c, map[string]any{"deleted": true, "id": id})
}

// ---------------------------------------------------------------------
// Kelas
// ---------------------------------------------------------------------

func listKelas(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.kelas.read") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat kelas.")
	}
	page, perPage, offset := parsePagination(c)
	jurusanID := optUUID(c, "jurusan_id")
	q := searchParam(c)
	ctx := c.Request().Context()
	var total int64
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM kelas
		 WHERE ($1::uuid IS NULL OR jurusan_id = $1)
		   AND ($2::text IS NULL OR nama ILIKE '%' || $2 || '%' OR tahun_ajaran ILIKE '%' || $2 || '%')`,
		jurusanID, q).Scan(&total); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca kelas.")
	}
	rows, err := pool.Query(ctx,
		`SELECT id, jurusan_id, tingkat, nama, tahun_ajaran, created_at, updated_at FROM kelas
		 WHERE ($1::uuid IS NULL OR jurusan_id = $1)
		   AND ($2::text IS NULL OR nama ILIKE '%' || $2 || '%' OR tahun_ajaran ILIKE '%' || $2 || '%')
		 ORDER BY tahun_ajaran DESC, tingkat ASC, nama ASC LIMIT $3 OFFSET $4`,
		jurusanID, q, perPage, offset)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca kelas.")
	}
	defer rows.Close()
	items := []kelasDTO{}
	for rows.Next() {
		var d kelasDTO
		if err := rows.Scan(&d.ID, &d.JurusanID, &d.Tingkat, &d.Nama, &d.TahunAjaran, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca kelas.")
		}
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca kelas.")
	}
	return writeList(c, items, page, perPage, total)
}

func getKelas(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.kelas.read") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat kelas.")
	}
	id := c.Param("id")
	if !validUUID(id) {
		return writeError(c, http.StatusNotFound, "not_found", "Kelas tidak ditemukan.")
	}
	var d kelasDTO
	err := pool.QueryRow(c.Request().Context(),
		`SELECT id, jurusan_id, tingkat, nama, tahun_ajaran, created_at, updated_at FROM kelas WHERE id = $1`, id).
		Scan(&d.ID, &d.JurusanID, &d.Tingkat, &d.Nama, &d.TahunAjaran, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return writeError(c, http.StatusNotFound, "not_found", "Kelas tidak ditemukan.")
	}
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca kelas.")
	}
	return writeOK(c, d)
}

type kelasInput struct {
	JurusanID   string `json:"jurusan_id"`
	Tingkat     int32  `json:"tingkat"`
	Nama        string `json:"nama"`
	TahunAjaran string `json:"tahun_ajaran"`
}

func (in kelasInput) validate() string {
	if !validUUID(in.JurusanID) {
		return "Jurusan wajib dipilih."
	}
	if in.Tingkat != 10 && in.Tingkat != 11 && in.Tingkat != 12 {
		return "Tingkat harus 10, 11, atau 12."
	}
	if strings.TrimSpace(in.Nama) == "" {
		return "Nama kelas wajib diisi."
	}
	if strings.TrimSpace(in.TahunAjaran) == "" {
		return "Tahun ajaran wajib diisi (mis. 2026/2027)."
	}
	return ""
}

func createKelas(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.kelas.create") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menambah kelas.")
	}
	var in kelasInput
	if err := c.Bind(&in); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	if msg := in.validate(); msg != "" {
		return writeError(c, http.StatusUnprocessableEntity, "validation_error", msg)
	}
	var d kelasDTO
	err := pool.QueryRow(c.Request().Context(),
		`INSERT INTO kelas (jurusan_id, tingkat, nama, tahun_ajaran) VALUES ($1, $2, $3, $4)
		 RETURNING id, jurusan_id, tingkat, nama, tahun_ajaran, created_at, updated_at`,
		in.JurusanID, in.Tingkat, strings.TrimSpace(in.Nama), strings.TrimSpace(in.TahunAjaran)).
		Scan(&d.ID, &d.JurusanID, &d.Tingkat, &d.Nama, &d.TahunAjaran, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			return writeError(c, http.StatusConflict, "conflict", conflictMessage(constraint))
		}
		if isFKViolation(err) {
			return writeError(c, http.StatusUnprocessableEntity, "validation_error", "Jurusan tidak ditemukan.")
		}
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menambah kelas.")
	}
	catat(c, pool, "create_kelas", "kelas", d.ID, map[string]any{"nama": d.Nama})
	return writeOKStatus(c, http.StatusCreated, d)
}

func updateKelas(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.kelas.update") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mengubah kelas.")
	}
	id := c.Param("id")
	if !validUUID(id) {
		return writeError(c, http.StatusNotFound, "not_found", "Kelas tidak ditemukan.")
	}
	var in kelasInput
	if err := c.Bind(&in); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	if msg := in.validate(); msg != "" {
		return writeError(c, http.StatusUnprocessableEntity, "validation_error", msg)
	}
	var d kelasDTO
	err := pool.QueryRow(c.Request().Context(),
		`UPDATE kelas SET jurusan_id = $2, tingkat = $3, nama = $4, tahun_ajaran = $5, updated_at = now()
		 WHERE id = $1
		 RETURNING id, jurusan_id, tingkat, nama, tahun_ajaran, created_at, updated_at`,
		id, in.JurusanID, in.Tingkat, strings.TrimSpace(in.Nama), strings.TrimSpace(in.TahunAjaran)).
		Scan(&d.ID, &d.JurusanID, &d.Tingkat, &d.Nama, &d.TahunAjaran, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return writeError(c, http.StatusNotFound, "not_found", "Kelas tidak ditemukan.")
	}
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			return writeError(c, http.StatusConflict, "conflict", conflictMessage(constraint))
		}
		if isFKViolation(err) {
			return writeError(c, http.StatusUnprocessableEntity, "validation_error", "Jurusan tidak ditemukan.")
		}
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mengubah kelas.")
	}
	catat(c, pool, "update_kelas", "kelas", d.ID, map[string]any{"nama": d.Nama})
	return writeOK(c, d)
}

func deleteKelas(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.kelas.delete") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menghapus kelas.")
	}
	id := c.Param("id")
	if !validUUID(id) {
		return writeError(c, http.StatusNotFound, "not_found", "Kelas tidak ditemukan.")
	}
	ctx := c.Request().Context()
	if n := countOrZero(ctx, pool, `SELECT COUNT(*) FROM siswa WHERE kelas_id = $1`, id); n > 0 {
		return writeError(c, http.StatusConflict, "conflict",
			fmt.Sprintf("Kelas masih dipakai %d siswa. Pindahkan siswa dulu.", n))
	} else if n < 0 {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa pemakaian kelas.")
	}
	// Guard jadwal: kelas yang sudah dipakai jadwal ujian tidak boleh
	// dihapus (jadwal_ujian.kelas_id ON DELETE RESTRICT + aturan 409).
	if n := countOrZero(ctx, pool, `SELECT COUNT(*) FROM jadwal_ujian WHERE kelas_id = $1`, id); n > 0 {
		return writeError(c, http.StatusConflict, "conflict",
			fmt.Sprintf("Kelas masih dipakai %d jadwal ujian. Hapus jadwal dulu.", n))
	} else if n < 0 {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa pemakaian kelas.")
	}
	ct, err := pool.Exec(ctx, `DELETE FROM kelas WHERE id = $1`, id)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghapus kelas.")
	}
	if ct.RowsAffected() == 0 {
		return writeError(c, http.StatusNotFound, "not_found", "Kelas tidak ditemukan.")
	}
	catat(c, pool, "delete_kelas", "kelas", id, nil)
	return writeOK(c, map[string]any{"deleted": true, "id": id})
}

// isFKViolation melaporkan pelanggaran foreign key (23503).
func isFKViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

// ---------------------------------------------------------------------
// Mapel
// ---------------------------------------------------------------------

func listMapel(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.mapel.read") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat mapel.")
	}
	page, perPage, offset := parsePagination(c)
	jenis := optJenis(c)
	jurusanID := optUUID(c, "jurusan_id")
	q := searchParam(c)
	ctx := c.Request().Context()
	var total int64
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM mapel
		 WHERE ($1::text IS NULL OR jenis = $1)
		   AND ($2::uuid IS NULL OR jurusan_id = $2)
		   AND ($3::text IS NULL OR nama ILIKE '%' || $3 || '%' OR kode ILIKE '%' || $3 || '%')`,
		jenis, jurusanID, q).Scan(&total); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca mapel.")
	}
	rows, err := pool.Query(ctx,
		`SELECT id, kode, nama, jenis, jurusan_id, kelas_tingkat, created_at, updated_at FROM mapel
		 WHERE ($1::text IS NULL OR jenis = $1)
		   AND ($2::uuid IS NULL OR jurusan_id = $2)
		   AND ($3::text IS NULL OR nama ILIKE '%' || $3 || '%' OR kode ILIKE '%' || $3 || '%')
		 ORDER BY nama ASC LIMIT $4 OFFSET $5`,
		jenis, jurusanID, q, perPage, offset)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca mapel.")
	}
	defer rows.Close()
	items := []mapelDTO{}
	for rows.Next() {
		var d mapelDTO
		var jurusanIDV *string
		var tingkatV *int32
		if err := rows.Scan(&d.ID, &d.Kode, &d.Nama, &d.Jenis, &jurusanIDV, &tingkatV, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca mapel.")
		}
		d.JurusanID, d.KelasTingkat = jurusanIDV, tingkatV
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca mapel.")
	}
	return writeList(c, items, page, perPage, total)
}

// optJenis membaca ?jenis= (hanya 'umum' / 'produktif' yang diterima).
func optJenis(c *echo.Context) *string {
	v := strings.TrimSpace(c.QueryParam("jenis"))
	if v != "umum" && v != "produktif" {
		return nil
	}
	return &v
}

func getMapel(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.mapel.read") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat mapel.")
	}
	id := c.Param("id")
	if !validUUID(id) {
		return writeError(c, http.StatusNotFound, "not_found", "Mapel tidak ditemukan.")
	}
	var d mapelDTO
	var jurusanIDV *string
	var tingkatV *int32
	err := pool.QueryRow(c.Request().Context(),
		`SELECT id, kode, nama, jenis, jurusan_id, kelas_tingkat, created_at, updated_at FROM mapel WHERE id = $1`, id).
		Scan(&d.ID, &d.Kode, &d.Nama, &d.Jenis, &jurusanIDV, &tingkatV, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return writeError(c, http.StatusNotFound, "not_found", "Mapel tidak ditemukan.")
	}
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca mapel.")
	}
	d.JurusanID, d.KelasTingkat = jurusanIDV, tingkatV
	return writeOK(c, d)
}

type mapelInput struct {
	Kode         string  `json:"kode"`
	Nama         string  `json:"nama"`
	Jenis        string  `json:"jenis"`
	JurusanID    *string `json:"jurusan_id"`
	KelasTingkat *int32  `json:"kelas_tingkat"`
}

func (in mapelInput) validate() string {
	if strings.TrimSpace(in.Kode) == "" {
		return "Kode mapel wajib diisi."
	}
	if strings.TrimSpace(in.Nama) == "" {
		return "Nama mapel wajib diisi."
	}
	if in.Jenis != "umum" && in.Jenis != "produktif" {
		return "Jenis mapel harus 'umum' atau 'produktif'."
	}
	if in.Jenis == "produktif" {
		if in.JurusanID == nil || !validUUID(*in.JurusanID) {
			return "Mapel produktif wajib terikat satu jurusan."
		}
	}
	if in.Jenis == "umum" && in.JurusanID != nil && strings.TrimSpace(*in.JurusanID) != "" {
		return "Mapel umum tidak terikat jurusan (jurusan_id harus kosong)."
	}
	if in.KelasTingkat != nil && *in.KelasTingkat != 10 && *in.KelasTingkat != 11 && *in.KelasTingkat != 12 {
		return "Kelas tingkat harus 10, 11, atau 12."
	}
	return ""
}

// mapelJurusanArg mengubah jurusan_id input menjadi argumen SQL (nil bila kosong).
func mapelJurusanArg(in mapelInput) *string {
	if in.JurusanID == nil {
		return nil
	}
	if strings.TrimSpace(*in.JurusanID) == "" {
		return nil
	}
	return in.JurusanID
}

func createMapel(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.mapel.create") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menambah mapel.")
	}
	var in mapelInput
	if err := c.Bind(&in); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	if msg := in.validate(); msg != "" {
		return writeError(c, http.StatusUnprocessableEntity, "validation_error", msg)
	}
	var d mapelDTO
	var jurusanIDV *string
	var tingkatV *int32
	err := pool.QueryRow(c.Request().Context(),
		`INSERT INTO mapel (kode, nama, jenis, jurusan_id, kelas_tingkat) VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, kode, nama, jenis, jurusan_id, kelas_tingkat, created_at, updated_at`,
		strings.TrimSpace(in.Kode), strings.TrimSpace(in.Nama), in.Jenis, mapelJurusanArg(in), in.KelasTingkat).
		Scan(&d.ID, &d.Kode, &d.Nama, &d.Jenis, &jurusanIDV, &tingkatV, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			return writeError(c, http.StatusConflict, "conflict", conflictMessage(constraint))
		}
		if isFKViolation(err) {
			return writeError(c, http.StatusUnprocessableEntity, "validation_error", "Jurusan tidak ditemukan.")
		}
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menambah mapel.")
	}
	d.JurusanID, d.KelasTingkat = jurusanIDV, tingkatV
	catat(c, pool, "create_mapel", "mapel", d.ID, map[string]any{"kode": d.Kode, "nama": d.Nama})
	return writeOKStatus(c, http.StatusCreated, d)
}

func updateMapel(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.mapel.update") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mengubah mapel.")
	}
	id := c.Param("id")
	if !validUUID(id) {
		return writeError(c, http.StatusNotFound, "not_found", "Mapel tidak ditemukan.")
	}
	var in mapelInput
	if err := c.Bind(&in); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	if msg := in.validate(); msg != "" {
		return writeError(c, http.StatusUnprocessableEntity, "validation_error", msg)
	}
	var d mapelDTO
	var jurusanIDV *string
	var tingkatV *int32
	err := pool.QueryRow(c.Request().Context(),
		`UPDATE mapel SET kode = $2, nama = $3, jenis = $4, jurusan_id = $5, kelas_tingkat = $6, updated_at = now()
		 WHERE id = $1
		 RETURNING id, kode, nama, jenis, jurusan_id, kelas_tingkat, created_at, updated_at`,
		id, strings.TrimSpace(in.Kode), strings.TrimSpace(in.Nama), in.Jenis, mapelJurusanArg(in), in.KelasTingkat).
		Scan(&d.ID, &d.Kode, &d.Nama, &d.Jenis, &jurusanIDV, &tingkatV, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return writeError(c, http.StatusNotFound, "not_found", "Mapel tidak ditemukan.")
	}
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			return writeError(c, http.StatusConflict, "conflict", conflictMessage(constraint))
		}
		if isFKViolation(err) {
			return writeError(c, http.StatusUnprocessableEntity, "validation_error", "Jurusan tidak ditemukan.")
		}
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mengubah mapel.")
	}
	d.JurusanID, d.KelasTingkat = jurusanIDV, tingkatV
	catat(c, pool, "update_mapel", "mapel", d.ID, map[string]any{"kode": d.Kode, "nama": d.Nama})
	return writeOK(c, d)
}

func deleteMapel(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.mapel.delete") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menghapus mapel.")
	}
	id := c.Param("id")
	if !validUUID(id) {
		return writeError(c, http.StatusNotFound, "not_found", "Mapel tidak ditemukan.")
	}
	ctx := c.Request().Context()
	guards := []struct {
		sql string
		msg string
	}{
		{`SELECT COUNT(*) FROM kompetensi WHERE mapel_id = $1`, "Mapel masih dipakai %d kompetensi. Hapus kompetensi dulu."},
		{`SELECT COUNT(*) FROM kisi_kisi WHERE mapel_id = $1`, "Mapel masih dipakai %d baris kisi-kisi. Hapus kisi-kisi dulu."},
		{`SELECT COUNT(*) FROM bank_soal WHERE mapel_id = $1`, "Mapel masih dipakai %d bank soal. Hapus/arsipkan bank soal dulu."},
		{`SELECT COUNT(*) FROM paket_ujian WHERE mapel_id = $1`, "Mapel masih dipakai %d paket ujian. Hapus paket dulu."},
		{`SELECT COUNT(*) FROM jadwal_ujian j JOIN paket_ujian p ON p.id = j.paket_id WHERE p.mapel_id = $1`, "Mapel masih dipakai %d jadwal ujian (via paket). Hapus jadwal dulu."},
	}
	for _, g := range guards {
		if n := countOrZero(ctx, pool, g.sql, id); n > 0 {
			return writeError(c, http.StatusConflict, "conflict", fmt.Sprintf(g.msg, n))
		} else if n < 0 {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa pemakaian mapel.")
		}
	}
	ct, err := pool.Exec(ctx, `DELETE FROM mapel WHERE id = $1`, id)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghapus mapel.")
	}
	if ct.RowsAffected() == 0 {
		return writeError(c, http.StatusNotFound, "not_found", "Mapel tidak ditemukan.")
	}
	catat(c, pool, "delete_mapel", "mapel", id, nil)
	return writeOK(c, map[string]any{"deleted": true, "id": id})
}

// ---------------------------------------------------------------------
// Guru
// ---------------------------------------------------------------------

func listGuru(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.guru.read") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat guru.")
	}
	page, perPage, offset := parsePagination(c)
	q := searchParam(c)
	ctx := c.Request().Context()
	var total int64
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM guru
		 WHERE ($1::text IS NULL OR nama ILIKE '%' || $1 || '%' OR COALESCE(nip, '') ILIKE '%' || $1 || '%')`, q).
		Scan(&total); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca guru.")
	}
	rows, err := pool.Query(ctx,
		`SELECT id, user_id, nip, nama, mapel_ids, created_at, updated_at FROM guru
		 WHERE ($1::text IS NULL OR nama ILIKE '%' || $1 || '%' OR COALESCE(nip, '') ILIKE '%' || $1 || '%')
		 ORDER BY nama ASC LIMIT $2 OFFSET $3`, q, perPage, offset)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca guru.")
	}
	defer rows.Close()
	items := []guruDTO{}
	for rows.Next() {
		var d guruDTO
		var userIDV, nipV *string
		var mapelIDs json.RawMessage
		if err := rows.Scan(&d.ID, &userIDV, &nipV, &d.Nama, &mapelIDs, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca guru.")
		}
		d.UserID, d.NIP, d.MapelIDs = userIDV, nipV, mapelIDs
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca guru.")
	}
	return writeList(c, items, page, perPage, total)
}

func getGuru(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.guru.read") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat guru.")
	}
	id := c.Param("id")
	if !validUUID(id) {
		return writeError(c, http.StatusNotFound, "not_found", "Guru tidak ditemukan.")
	}
	var d guruDTO
	var userIDV, nipV *string
	var mapelIDs json.RawMessage
	err := pool.QueryRow(c.Request().Context(),
		`SELECT id, user_id, nip, nama, mapel_ids, created_at, updated_at FROM guru WHERE id = $1`, id).
		Scan(&d.ID, &userIDV, &nipV, &d.Nama, &mapelIDs, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return writeError(c, http.StatusNotFound, "not_found", "Guru tidak ditemukan.")
	}
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca guru.")
	}
	d.UserID, d.NIP, d.MapelIDs = userIDV, nipV, mapelIDs
	return writeOK(c, d)
}

type guruInput struct {
	UserID   *string  `json:"user_id"`
	NIP      *string  `json:"nip"`
	Nama     string   `json:"nama"`
	MapelIDs []string `json:"mapel_ids"`
}

func (in guruInput) validate() string {
	if strings.TrimSpace(in.Nama) == "" {
		return "Nama guru wajib diisi."
	}
	if in.UserID != nil && strings.TrimSpace(*in.UserID) != "" && !validUUID(*in.UserID) {
		return "user_id tidak valid."
	}
	for _, m := range in.MapelIDs {
		if !validUUID(m) {
			return "Daftar mapel berisi ID yang tidak valid."
		}
	}
	return ""
}

// guruArgs mengubah input menjadi argumen SQL (nil untuk string kosong).
func guruArgs(in guruInput) (userID, nip any, mapelIDs json.RawMessage) {
	if in.UserID != nil && strings.TrimSpace(*in.UserID) != "" {
		userID = *in.UserID
	}
	if in.NIP != nil && strings.TrimSpace(*in.NIP) != "" {
		nip = strings.TrimSpace(*in.NIP)
	}
	ids := in.MapelIDs
	if ids == nil {
		ids = []string{}
	}
	raw, _ := json.Marshal(ids)
	return userID, nip, raw
}

func createGuru(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.guru.create") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menambah guru.")
	}
	var in guruInput
	if err := c.Bind(&in); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	if msg := in.validate(); msg != "" {
		return writeError(c, http.StatusUnprocessableEntity, "validation_error", msg)
	}
	userID, nip, mapelIDs := guruArgs(in)
	var d guruDTO
	var userIDV, nipV *string
	var mapelIDsV json.RawMessage
	err := pool.QueryRow(c.Request().Context(),
		`INSERT INTO guru (user_id, nip, nama, mapel_ids) VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, nip, nama, mapel_ids, created_at, updated_at`,
		userID, nip, strings.TrimSpace(in.Nama), mapelIDs).
		Scan(&d.ID, &userIDV, &nipV, &d.Nama, &mapelIDsV, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			return writeError(c, http.StatusConflict, "conflict", conflictMessage(constraint))
		}
		if isFKViolation(err) {
			return writeError(c, http.StatusUnprocessableEntity, "validation_error", "Akun pengguna (user_id) tidak ditemukan.")
		}
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menambah guru.")
	}
	d.UserID, d.NIP, d.MapelIDs = userIDV, nipV, mapelIDsV
	catat(c, pool, "create_guru", "guru", d.ID, map[string]any{"nama": d.Nama})
	return writeOKStatus(c, http.StatusCreated, d)
}

func updateGuru(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.guru.update") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mengubah guru.")
	}
	id := c.Param("id")
	if !validUUID(id) {
		return writeError(c, http.StatusNotFound, "not_found", "Guru tidak ditemukan.")
	}
	var in guruInput
	if err := c.Bind(&in); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	if msg := in.validate(); msg != "" {
		return writeError(c, http.StatusUnprocessableEntity, "validation_error", msg)
	}
	userID, nip, mapelIDs := guruArgs(in)
	var d guruDTO
	var userIDV, nipV *string
	var mapelIDsV json.RawMessage
	err := pool.QueryRow(c.Request().Context(),
		`UPDATE guru SET user_id = $2, nip = $3, nama = $4, mapel_ids = $5, updated_at = now()
		 WHERE id = $1
		 RETURNING id, user_id, nip, nama, mapel_ids, created_at, updated_at`,
		id, userID, nip, strings.TrimSpace(in.Nama), mapelIDs).
		Scan(&d.ID, &userIDV, &nipV, &d.Nama, &mapelIDsV, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return writeError(c, http.StatusNotFound, "not_found", "Guru tidak ditemukan.")
	}
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			return writeError(c, http.StatusConflict, "conflict", conflictMessage(constraint))
		}
		if isFKViolation(err) {
			return writeError(c, http.StatusUnprocessableEntity, "validation_error", "Akun pengguna (user_id) tidak ditemukan.")
		}
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mengubah guru.")
	}
	d.UserID, d.NIP, d.MapelIDs = userIDV, nipV, mapelIDsV
	catat(c, pool, "update_guru", "guru", d.ID, map[string]any{"nama": d.Nama})
	return writeOK(c, d)
}

func deleteGuru(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.guru.delete") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menghapus guru.")
	}
	id := c.Param("id")
	if !validUUID(id) {
		return writeError(c, http.StatusNotFound, "not_found", "Guru tidak ditemukan.")
	}
	ctx := c.Request().Context()
	// Guru yang jadi pemilik bank soal tidak boleh dihapus (ANALISIS 16.2:
	// staf yang pernah membuat bank soal hanya dinonaktifkan).
	// bank_soal.guru_id adalah rujukan ke users(id) (Tim 3), bukan ke
	// guru(id), jadi guard bank soal di sini memakai user_id guru.
	var userID *string
	_ = pool.QueryRow(ctx, `SELECT user_id FROM guru WHERE id = $1`, id).Scan(&userID)
	if userID != nil && *userID != "" {
		if n := countOrZero(ctx, pool, `SELECT COUNT(*) FROM bank_soal WHERE guru_id = $1`, *userID); n > 0 {
			return writeError(c, http.StatusConflict, "conflict",
				fmt.Sprintf("Guru masih pemilik %d bank soal. Nonaktifkan akunnya, jangan hapus.", n))
		} else if n < 0 {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa pemakaian guru.")
		}
	}
	ct, err := pool.Exec(ctx, `DELETE FROM guru WHERE id = $1`, id)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghapus guru.")
	}
	if ct.RowsAffected() == 0 {
		return writeError(c, http.StatusNotFound, "not_found", "Guru tidak ditemukan.")
	}
	catat(c, pool, "delete_guru", "guru", id, nil)
	return writeOK(c, map[string]any{"deleted": true, "id": id})
}

// ---------------------------------------------------------------------
// Siswa
// ---------------------------------------------------------------------

func listSiswa(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.siswa.read") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat siswa.")
	}
	page, perPage, offset := parsePagination(c)
	kelasID := optUUID(c, "kelas_id")
	jurusanID := optUUID(c, "jurusan_id")
	isActive := optBool(c, "is_active")
	q := searchParam(c)
	ctx := c.Request().Context()
	var total int64
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM siswa
		 WHERE ($1::uuid IS NULL OR kelas_id = $1)
		   AND ($2::uuid IS NULL OR jurusan_id = $2)
		   AND ($3::boolean IS NULL OR is_active = $3)
		   AND ($4::text IS NULL OR nama ILIKE '%' || $4 || '%' OR nisn ILIKE '%' || $4 || '%' OR COALESCE(nis, '') ILIKE '%' || $4 || '%')`,
		kelasID, jurusanID, isActive, q).Scan(&total); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca siswa.")
	}
	rows, err := pool.Query(ctx,
		`SELECT id, user_id, nisn, nis, nama, kelas_id, jurusan_id, tahun_masuk, is_active, created_at, updated_at FROM siswa
		 WHERE ($1::uuid IS NULL OR kelas_id = $1)
		   AND ($2::uuid IS NULL OR jurusan_id = $2)
		   AND ($3::boolean IS NULL OR is_active = $3)
		   AND ($4::text IS NULL OR nama ILIKE '%' || $4 || '%' OR nisn ILIKE '%' || $4 || '%' OR COALESCE(nis, '') ILIKE '%' || $4 || '%')
		 ORDER BY nama ASC LIMIT $5 OFFSET $6`,
		kelasID, jurusanID, isActive, q, perPage, offset)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca siswa.")
	}
	defer rows.Close()
	items := []siswaDTO{}
	for rows.Next() {
		var d siswaDTO
		var userIDV, nisV *string
		var tahunMasukV *int32
		if err := rows.Scan(&d.ID, &userIDV, &d.NISN, &nisV, &d.Nama, &d.KelasID, &d.JurusanID, &tahunMasukV, &d.IsActive, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca siswa.")
		}
		d.UserID, d.NIS, d.TahunMasuk = userIDV, nisV, tahunMasukV
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca siswa.")
	}
	return writeList(c, items, page, perPage, total)
}

func getSiswa(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.siswa.read") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat siswa.")
	}
	id := c.Param("id")
	if !validUUID(id) {
		return writeError(c, http.StatusNotFound, "not_found", "Siswa tidak ditemukan.")
	}
	var d siswaDTO
	var userIDV, nisV *string
	var tahunMasukV *int32
	err := pool.QueryRow(c.Request().Context(),
		`SELECT id, user_id, nisn, nis, nama, kelas_id, jurusan_id, tahun_masuk, is_active, created_at, updated_at FROM siswa WHERE id = $1`, id).
		Scan(&d.ID, &userIDV, &d.NISN, &nisV, &d.Nama, &d.KelasID, &d.JurusanID, &tahunMasukV, &d.IsActive, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return writeError(c, http.StatusNotFound, "not_found", "Siswa tidak ditemukan.")
	}
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca siswa.")
	}
	d.UserID, d.NIS, d.TahunMasuk = userIDV, nisV, tahunMasukV
	return writeOK(c, d)
}

type siswaInput struct {
	UserID     *string `json:"user_id"`
	NISN       string  `json:"nisn"`
	NIS        *string `json:"nis"`
	Nama       string  `json:"nama"`
	KelasID    string  `json:"kelas_id"`
	JurusanID  string  `json:"jurusan_id"`
	TahunMasuk *int32  `json:"tahun_masuk"`
	IsActive   *bool   `json:"is_active"`
}

func (in siswaInput) validate() string {
	if strings.TrimSpace(in.NISN) == "" {
		return "NISN wajib diisi."
	}
	if strings.TrimSpace(in.Nama) == "" {
		return "Nama siswa wajib diisi."
	}
	if !validUUID(in.KelasID) {
		return "Kelas wajib dipilih."
	}
	if !validUUID(in.JurusanID) {
		return "Jurusan wajib dipilih."
	}
	if in.UserID != nil && strings.TrimSpace(*in.UserID) != "" && !validUUID(*in.UserID) {
		return "user_id tidak valid."
	}
	if in.TahunMasuk != nil && (*in.TahunMasuk < 1900 || *in.TahunMasuk > 2100) {
		return "Tahun masuk tidak valid."
	}
	return ""
}

// siswaNISArg mengubah NIS kosong menjadi nil.
func siswaNISArg(in siswaInput) *string {
	if in.NIS == nil {
		return nil
	}
	if strings.TrimSpace(*in.NIS) == "" {
		return nil
	}
	v := strings.TrimSpace(*in.NIS)
	return &v
}

// createStudentAccountTx membuat akun login siswa di dalam transaksi:
// username = NISN, password awal = NISN (Argon2id), role 'siswa' di kolom
// users.role (SATU role per user — tidak ada tabel user_roles),
// must_change_password = TRUE. Bila username sudah ada (akun yatim),
// akun lama dipakai kembali. Mengembalikan user_id.
func createStudentAccountTx(ctx context.Context, tx pgx.Tx, nisn, nama string) (string, error) {
	hash, err := hashPassword(nisn)
	if err != nil {
		return "", err
	}
	var userID string
	err = tx.QueryRow(ctx,
		`INSERT INTO users (nama, username, password_hash, role, must_change_password, is_active)
		 VALUES ($1, $2, $3, 'siswa', TRUE, TRUE)
		 ON CONFLICT (username) DO NOTHING
		 RETURNING id`, nama, nisn, hash).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		// Username sudah ada: pakai akun lama + paksa ganti password.
		if err := tx.QueryRow(ctx,
			`SELECT id FROM users WHERE username = $1`, nisn).Scan(&userID); err != nil {
			return "", err
		}
		_, _ = tx.Exec(ctx,
			`UPDATE users SET must_change_password = TRUE, role = 'siswa' WHERE id = $1`, userID)
	} else if err != nil {
		return "", err
	}
	return userID, nil
}

func createSiswa(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.siswa.create") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menambah siswa.")
	}
	var in siswaInput
	if err := c.Bind(&in); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	if msg := in.validate(); msg != "" {
		return writeError(c, http.StatusUnprocessableEntity, "validation_error", msg)
	}
	ctx := c.Request().Context()
	isActive := true
	if in.IsActive != nil {
		isActive = *in.IsActive
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menambah siswa.")
	}
	defer tx.Rollback(ctx)
	userID := ""
	if in.UserID != nil && strings.TrimSpace(*in.UserID) != "" {
		userID = *in.UserID
	} else {
		// Tanpa user_id: buatkan akun login otomatis (sama seperti import).
		userID, err = createStudentAccountTx(ctx, tx, strings.TrimSpace(in.NISN), strings.TrimSpace(in.Nama))
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membuat akun siswa.")
		}
	}
	var userIDArg any
	if userID != "" {
		userIDArg = userID
	}
	var d siswaDTO
	var userIDV, nisV *string
	var tahunMasukV *int32
	err = tx.QueryRow(ctx,
		`INSERT INTO siswa (user_id, nisn, nis, nama, kelas_id, jurusan_id, tahun_masuk, is_active)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, user_id, nisn, nis, nama, kelas_id, jurusan_id, tahun_masuk, is_active, created_at, updated_at`,
		userIDArg, strings.TrimSpace(in.NISN), siswaNISArg(in), strings.TrimSpace(in.Nama),
		in.KelasID, in.JurusanID, in.TahunMasuk, isActive).
		Scan(&d.ID, &userIDV, &d.NISN, &nisV, &d.Nama, &d.KelasID, &d.JurusanID, &tahunMasukV, &d.IsActive, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			return writeError(c, http.StatusConflict, "conflict", conflictMessage(constraint))
		}
		if isFKViolation(err) {
			return writeError(c, http.StatusUnprocessableEntity, "validation_error", "Kelas/jurusan/akun pengguna tidak ditemukan.")
		}
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menambah siswa.")
	}
	if err := tx.Commit(ctx); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menambah siswa.")
	}
	d.UserID, d.NIS, d.TahunMasuk = userIDV, nisV, tahunMasukV
	catat(c, pool, "create_siswa", "siswa", d.ID, map[string]any{"nisn": d.NISN, "nama": d.Nama})
	return writeOKStatus(c, http.StatusCreated, d)
}

func updateSiswa(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.siswa.update") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mengubah siswa.")
	}
	id := c.Param("id")
	if !validUUID(id) {
		return writeError(c, http.StatusNotFound, "not_found", "Siswa tidak ditemukan.")
	}
	var in siswaInput
	if err := c.Bind(&in); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	if msg := in.validate(); msg != "" {
		return writeError(c, http.StatusUnprocessableEntity, "validation_error", msg)
	}
	ctx := c.Request().Context()
	// user_id tidak dikirim = pertahankan yang lama.
	userIDArg := any(nil)
	if in.UserID != nil {
		if strings.TrimSpace(*in.UserID) == "" {
			userIDArg = nil
		} else {
			userIDArg = *in.UserID
		}
	} else {
		var existing *string
		_ = pool.QueryRow(ctx, `SELECT user_id FROM siswa WHERE id = $1`, id).Scan(&existing)
		if existing != nil {
			userIDArg = *existing
		}
	}
	isActive := true
	if in.IsActive != nil {
		isActive = *in.IsActive
	}
	var d siswaDTO
	var userIDV, nisV *string
	var tahunMasukV *int32
	err := pool.QueryRow(ctx,
		`UPDATE siswa SET user_id = $2, nisn = $3, nis = $4, nama = $5, kelas_id = $6, jurusan_id = $7,
		   tahun_masuk = $8, is_active = $9, updated_at = now()
		 WHERE id = $1
		 RETURNING id, user_id, nisn, nis, nama, kelas_id, jurusan_id, tahun_masuk, is_active, created_at, updated_at`,
		id, userIDArg, strings.TrimSpace(in.NISN), siswaNISArg(in), strings.TrimSpace(in.Nama),
		in.KelasID, in.JurusanID, in.TahunMasuk, isActive).
		Scan(&d.ID, &userIDV, &d.NISN, &nisV, &d.Nama, &d.KelasID, &d.JurusanID, &tahunMasukV, &d.IsActive, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return writeError(c, http.StatusNotFound, "not_found", "Siswa tidak ditemukan.")
	}
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			return writeError(c, http.StatusConflict, "conflict", conflictMessage(constraint))
		}
		if isFKViolation(err) {
			return writeError(c, http.StatusUnprocessableEntity, "validation_error", "Kelas/jurusan/akun pengguna tidak ditemukan.")
		}
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mengubah siswa.")
	}
	d.UserID, d.NIS, d.TahunMasuk = userIDV, nisV, tahunMasukV
	catat(c, pool, "update_siswa", "siswa", d.ID, map[string]any{"nisn": d.NISN, "nama": d.Nama})
	return writeOK(c, d)
}

func deleteSiswa(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.siswa.delete") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menghapus siswa.")
	}
	id := c.Param("id")
	if !validUUID(id) {
		return writeError(c, http.StatusNotFound, "not_found", "Siswa tidak ditemukan.")
	}
	ctx := c.Request().Context()
	// Guard: siswa yang sudah punya attempt / nilai akhir / remedial
	// hanya dinonaktifkan (ANALISIS 16.2), tidak dihapus.
	// attempt.siswa_id menyimpan auth_siswa_id (= users.id akun siswa),
	// sedangkan nilai_akhir/remedial.siswa_id merujuk siswa.id.
	var userID *string
	_ = pool.QueryRow(ctx, `SELECT user_id FROM siswa WHERE id = $1`, id).Scan(&userID)
	if userID != nil && *userID != "" {
		if n := countOrZero(ctx, pool, `SELECT COUNT(*) FROM attempt WHERE siswa_id = $1`, *userID); n > 0 {
			return writeError(c, http.StatusConflict, "conflict",
				fmt.Sprintf("Siswa sudah punya %d attempt ujian. Nonaktifkan saja, jangan hapus.", n))
		} else if n < 0 {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa pemakaian siswa.")
		}
	}
	if n := countOrZero(ctx, pool, `SELECT COUNT(*) FROM nilai_akhir WHERE siswa_id = $1`, id); n > 0 {
		return writeError(c, http.StatusConflict, "conflict",
			fmt.Sprintf("Siswa sudah punya %d nilai akhir. Nonaktifkan saja, jangan hapus.", n))
	} else if n < 0 {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa pemakaian siswa.")
	}
	if n := countOrZero(ctx, pool, `SELECT COUNT(*) FROM remedial WHERE siswa_id = $1`, id); n > 0 {
		return writeError(c, http.StatusConflict, "conflict",
			fmt.Sprintf("Siswa sudah punya %d nilai remedial. Nonaktifkan saja, jangan hapus.", n))
	} else if n < 0 {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa pemakaian siswa.")
	}
	// Akun login (users) SENGAJA tidak dihapus: riwayat & username tidak
	// dipakai ulang untuk orang lain.
	ct, err := pool.Exec(ctx, `DELETE FROM siswa WHERE id = $1`, id)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghapus siswa.")
	}
	if ct.RowsAffected() == 0 {
		return writeError(c, http.StatusNotFound, "not_found", "Siswa tidak ditemukan.")
	}
	catat(c, pool, "delete_siswa", "siswa", id, nil)
	return writeOK(c, map[string]any{"deleted": true, "id": id})
}

// ---------------------------------------------------------------------
// Import Siswa dari CSV
// ---------------------------------------------------------------------

type importRowError struct {
	Baris int    `json:"baris"`
	Pesan string `json:"pesan"`
}

// importResult: SudahAda = baris yang NISN-nya sudah terdaftar di DB —
// dihitung TERPISAH dari gagal (Fase 3, BUG-02: import IDEMPOTEN.
// Retry import yang sama tidak lagi melaporkan "gagal" massal;
// import kedua 300 siswa yang sama -> sukses 0, sudah_ada 300, gagal 0).
type importResult struct {
	Sukses   int              `json:"sukses"`
	Gagal    int              `json:"gagal"`
	SudahAda int              `json:"sudah_ada"`
	Errors   []importRowError `json:"errors"`
	Catatan  string           `json:"catatan,omitempty"`
}

const (
	maxImportRows     = 2000
	maxImportFileMB   = 5
	kelasLookupByID   = `SELECT id, jurusan_id FROM kelas WHERE id = $1`
	kelasLookupByNama = `SELECT id, jurusan_id FROM kelas WHERE LOWER(nama) = LOWER($1) ORDER BY created_at DESC LIMIT 1`
)

// infoImportSiswa menangani GET /api/v1/master/siswa/import.
//
// Mengembalikan format CSV yang diterima POST /siswa/import. Route ini
// WAJIB didaftarkan SEBELUM GET /siswa/:id: segmen statis "import"
// bila tidak ada route-nya akan tertangkap param :id ("import" bukan
// UUID) dan getSiswa mengembalikan 404 "Siswa tidak ditemukan".
func infoImportSiswa(c *echo.Context, pool *pgxpool.Pool) error {
	_ = pool
	if !audit.HasPermission(c, "master.siswa.import") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin import siswa.")
	}
	return writeOK(c, map[string]any{
		"method":  "POST",
		"path":    "/api/v1/master/siswa/import",
		"body":    "multipart/form-data, field \"file\" berisi CSV UTF-8",
		"kolom":   []string{"nama", "nisn", "nis (opsional)", "kelas_id / kelas_kode / kelas / kode_kelas"},
		"catatan": "kelas_kode = id UUID kelas ATAU nama kelas. Akun login dibuat otomatis: username = NISN, password awal = NISN (wajib ganti saat login pertama). Import IDEMPOTEN: NISN yang sudah ada dihitung sudah_ada (bukan gagal). Rekomendasi: maks. 100 siswa per file.",
		"contoh":  "nama,nisn,nis,kelas_kode\nBudi Santoso,0012345678,1234,XI RPL 1",
	})
}

// importSiswa menangani POST /api/v1/master/siswa/import.
// Form multipart: field "file" berisi CSV UTF-8.
// Header yang diterima:
//
//	nama,nisn,nis,kelas_kode   (kelas_kode = id UUID kelas ATAU nama kelas)
//	nama,nisn,kelas_id         (alias: kelas, kode_kelas)
//
// Kolom nis opsional. Baris yang gagal dilewati + dilaporkan per baris;
// baris yang valid disimpan dalam satu transaksi.
//
// IDEMPOTEN (Fase 3, BUG-02): NISN yang sudah ada di DB dihitung sebagai
// `sudah_ada` (bukan error, bukan duplikat, tidak buat user baru) —
// retry import yang sama aman. Rekomendasi operasional: import
// maksimal 100 siswa per file (±17 dtk, jauh di bawah WriteTimeout
// server 120 dtk; lihat KONTRAK §10) lalu ulangi file berikutnya.
func importSiswa(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.siswa.import") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin import siswa.")
	}
	if err := c.Request().ParseMultipartForm(maxImportFileMB << 20); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Body multipart tidak valid (maks. 5 MB).")
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Field file (CSV) wajib diunggah.")
	}
	f, err := fileHeader.Open()
	if err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "File CSV tidak bisa dibuka.")
	}
	defer f.Close()
	reader := csv.NewReader(io.LimitReader(f, (maxImportFileMB<<20)+1))
	reader.FieldsPerRecord = -1 // baris pendek/panjang ditangani manual per kolom
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "File bukan CSV yang valid.")
	}
	if len(records) == 0 {
		return writeError(c, http.StatusBadRequest, "validation_error", "File CSV kosong.")
	}
	// Normalisasi header: huruf kecil, tanpa spasi, tanpa BOM.
	header := make(map[string]int, len(records[0]))
	for i, h := range records[0] {
		h = strings.ToLower(strings.TrimSpace(h))
		h = strings.TrimPrefix(h, "\ufeff")
		header[h] = i
	}
	col := func(names ...string) (int, bool) {
		for _, n := range names {
			if idx, ok := header[n]; ok {
				return idx, true
			}
		}
		return -1, false
	}
	namaCol, okNama := col("nama", "nama_lengkap", "nama_siswa")
	nisnCol, okNISN := col("nisn")
	nisCol, _ := col("nis")
	kelasCol, okKelas := col("kelas_id", "kelas_kode", "kelas", "kode_kelas")
	if !okNama || !okNISN || !okKelas {
		return writeError(c, http.StatusBadRequest, "validation_error",
			"Header CSV tidak dikenali. Harus ada kolom: nama, nisn, dan kelas_id/kelas_kode (opsional: nis).")
	}
	dataRows := records[1:]
	if len(dataRows) > maxImportRows {
		return writeError(c, http.StatusBadRequest, "validation_error",
			fmt.Sprintf("Maksimal %d baris per file import.", maxImportRows))
	}
	cell := func(row []string, idx int) string {
		if idx < 0 || idx >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[idx])
	}
	ctx := c.Request().Context()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memulai import.")
	}
	defer tx.Rollback(ctx)
	result := importResult{Errors: []importRowError{}}
	seenNISN := map[string]bool{}
	for i, row := range dataRows {
		baris := i + 2 // baris 1 = header
		nama := cell(row, namaCol)
		nisn := cell(row, nisnCol)
		nis := cell(row, nisCol)
		kelasVal := cell(row, kelasCol)
		fail := func(pesan string) {
			result.Gagal++
			result.Errors = append(result.Errors, importRowError{Baris: baris, Pesan: pesan})
		}
		if nama == "" {
			fail("Nama wajib diisi.")
			continue
		}
		if nisn == "" {
			fail("NISN wajib diisi.")
			continue
		}
		if kelasVal == "" {
			fail("Kelas (kelas_id/kelas_kode) wajib diisi.")
			continue
		}
		if seenNISN[nisn] {
			fail("NISN duplikat di dalam file.")
			continue
		}
		seenNISN[nisn] = true
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM siswa WHERE nisn = $1)`, nisn).Scan(&exists); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa NISN.")
		}
		if exists {
			// Idempotent: NISN sudah ada -> lewati tanpa error,
			// tanpa membuat user baru, tanpa baris errors.
			result.SudahAda++
			continue
		}
		// Resolusi kelas: UUID -> id; selain itu -> nama kelas (tidak peka huruf).
		var kelasID, jurusanID string
		var lookupSQL string
		if validUUID(kelasVal) {
			lookupSQL = kelasLookupByID
		} else {
			lookupSQL = kelasLookupByNama
		}
		if err := tx.QueryRow(ctx, lookupSQL, kelasVal).Scan(&kelasID, &jurusanID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				fail(fmt.Sprintf("Kelas %q tidak ditemukan.", kelasVal))
				continue
			}
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mencari kelas.")
		}
		userID, err := createStudentAccountTx(ctx, tx, nisn, nama)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membuat akun siswa.")
		}
		var nisArg any
		if nis != "" {
			nisArg = nis
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO siswa (user_id, nisn, nis, nama, kelas_id, jurusan_id, is_active)
			 VALUES ($1, $2, $3, $4, $5, $6, TRUE)`,
			userID, nisn, nisArg, nama, kelasID, jurusanID); err != nil {
			if constraint, ok := uniqueConstraint(err); ok {
				// Race dengan import paralel: NISN ternyata sudah
				// masuk di antara cek EXISTS dan INSERT -> idempotent,
				// hitung sudah_ada (jalur utama adalah cek EXISTS di
				// atas; jalur ini hanya jaring pengaman race).
				if constraint == "siswa_nisn_key" || constraint == "users_username_key" {
					result.SudahAda++
					continue
				}
				fail(conflictMessage(constraint))
				continue
			}
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menyimpan siswa.")
		}
		result.Sukses++
	}
	if err := tx.Commit(ctx); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menyelesaikan import.")
	}
	catat(c, pool, "import_siswa", "siswa", "", map[string]any{"sukses": result.Sukses, "gagal": result.Gagal, "sudah_ada": result.SudahAda})
	return writeOK(c, result)
}

// ---------------------------------------------------------------------
// Kompetensi (guard: master.mapel.*)
// ---------------------------------------------------------------------

func listKompetensi(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.mapel.read") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat kompetensi.")
	}
	page, perPage, offset := parsePagination(c)
	mapelID := optUUID(c, "mapel_id")
	q := searchParam(c)
	ctx := c.Request().Context()
	var total int64
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM kompetensi
		 WHERE ($1::uuid IS NULL OR mapel_id = $1)
		   AND ($2::text IS NULL OR kode ILIKE '%' || $2 || '%' OR deskripsi ILIKE '%' || $2 || '%')`,
		mapelID, q).Scan(&total); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca kompetensi.")
	}
	rows, err := pool.Query(ctx,
		`SELECT id, mapel_id, kode, deskripsi, created_at, updated_at FROM kompetensi
		 WHERE ($1::uuid IS NULL OR mapel_id = $1)
		   AND ($2::text IS NULL OR kode ILIKE '%' || $2 || '%' OR deskripsi ILIKE '%' || $2 || '%')
		 ORDER BY kode ASC LIMIT $3 OFFSET $4`, mapelID, q, perPage, offset)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca kompetensi.")
	}
	defer rows.Close()
	items := []kompetensiDTO{}
	for rows.Next() {
		var d kompetensiDTO
		if err := rows.Scan(&d.ID, &d.MapelID, &d.Kode, &d.Deskripsi, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca kompetensi.")
		}
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca kompetensi.")
	}
	return writeList(c, items, page, perPage, total)
}

func getKompetensi(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.mapel.read") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat kompetensi.")
	}
	id := c.Param("id")
	if !validUUID(id) {
		return writeError(c, http.StatusNotFound, "not_found", "Kompetensi tidak ditemukan.")
	}
	var d kompetensiDTO
	err := pool.QueryRow(c.Request().Context(),
		`SELECT id, mapel_id, kode, deskripsi, created_at, updated_at FROM kompetensi WHERE id = $1`, id).
		Scan(&d.ID, &d.MapelID, &d.Kode, &d.Deskripsi, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return writeError(c, http.StatusNotFound, "not_found", "Kompetensi tidak ditemukan.")
	}
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca kompetensi.")
	}
	return writeOK(c, d)
}

type kompetensiInput struct {
	MapelID   string `json:"mapel_id"`
	Kode      string `json:"kode"`
	Deskripsi string `json:"deskripsi"`
}

func (in kompetensiInput) validate() string {
	if !validUUID(in.MapelID) {
		return "Mapel wajib dipilih."
	}
	if strings.TrimSpace(in.Kode) == "" {
		return "Kode kompetensi wajib diisi."
	}
	return ""
}

func createKompetensi(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.mapel.create") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menambah kompetensi.")
	}
	var in kompetensiInput
	if err := c.Bind(&in); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	if msg := in.validate(); msg != "" {
		return writeError(c, http.StatusUnprocessableEntity, "validation_error", msg)
	}
	var d kompetensiDTO
	err := pool.QueryRow(c.Request().Context(),
		`INSERT INTO kompetensi (mapel_id, kode, deskripsi) VALUES ($1, $2, $3)
		 RETURNING id, mapel_id, kode, deskripsi, created_at, updated_at`,
		in.MapelID, strings.TrimSpace(in.Kode), in.Deskripsi).
		Scan(&d.ID, &d.MapelID, &d.Kode, &d.Deskripsi, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			return writeError(c, http.StatusConflict, "conflict", conflictMessage(constraint))
		}
		if isFKViolation(err) {
			return writeError(c, http.StatusUnprocessableEntity, "validation_error", "Mapel tidak ditemukan.")
		}
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menambah kompetensi.")
	}
	catat(c, pool, "create_kompetensi", "kompetensi", d.ID, map[string]any{"kode": d.Kode})
	return writeOKStatus(c, http.StatusCreated, d)
}

func updateKompetensi(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.mapel.update") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mengubah kompetensi.")
	}
	id := c.Param("id")
	if !validUUID(id) {
		return writeError(c, http.StatusNotFound, "not_found", "Kompetensi tidak ditemukan.")
	}
	var in kompetensiInput
	if err := c.Bind(&in); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	if msg := in.validate(); msg != "" {
		return writeError(c, http.StatusUnprocessableEntity, "validation_error", msg)
	}
	var d kompetensiDTO
	err := pool.QueryRow(c.Request().Context(),
		`UPDATE kompetensi SET mapel_id = $2, kode = $3, deskripsi = $4, updated_at = now()
		 WHERE id = $1
		 RETURNING id, mapel_id, kode, deskripsi, created_at, updated_at`,
		id, in.MapelID, strings.TrimSpace(in.Kode), in.Deskripsi).
		Scan(&d.ID, &d.MapelID, &d.Kode, &d.Deskripsi, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return writeError(c, http.StatusNotFound, "not_found", "Kompetensi tidak ditemukan.")
	}
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			return writeError(c, http.StatusConflict, "conflict", conflictMessage(constraint))
		}
		if isFKViolation(err) {
			return writeError(c, http.StatusUnprocessableEntity, "validation_error", "Mapel tidak ditemukan.")
		}
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mengubah kompetensi.")
	}
	catat(c, pool, "update_kompetensi", "kompetensi", d.ID, map[string]any{"kode": d.Kode})
	return writeOK(c, d)
}

func deleteKompetensi(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.mapel.delete") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menghapus kompetensi.")
	}
	id := c.Param("id")
	if !validUUID(id) {
		return writeError(c, http.StatusNotFound, "not_found", "Kompetensi tidak ditemukan.")
	}
	ctx := c.Request().Context()
	guards := []struct {
		sql string
		msg string
	}{
		{`SELECT COUNT(*) FROM kisi_kisi WHERE kompetensi_id = $1`, "Kompetensi masih dipakai %d baris kisi-kisi."},
		{`SELECT COUNT(*) FROM soal WHERE kompetensi_id = $1`, "Kompetensi masih dipakai %d soal."},
	}
	for _, g := range guards {
		if n := countOrZero(ctx, pool, g.sql, id); n > 0 {
			return writeError(c, http.StatusConflict, "conflict", fmt.Sprintf(g.msg, n))
		} else if n < 0 {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa pemakaian kompetensi.")
		}
	}
	ct, err := pool.Exec(ctx, `DELETE FROM kompetensi WHERE id = $1`, id)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghapus kompetensi.")
	}
	if ct.RowsAffected() == 0 {
		return writeError(c, http.StatusNotFound, "not_found", "Kompetensi tidak ditemukan.")
	}
	catat(c, pool, "delete_kompetensi", "kompetensi", id, nil)
	return writeOK(c, map[string]any{"deleted": true, "id": id})
}

// ---------------------------------------------------------------------
// Kisi-kisi (guard: master.mapel.*)
// ---------------------------------------------------------------------

func listKisiKisi(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.mapel.read") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat kisi-kisi.")
	}
	page, perPage, offset := parsePagination(c)
	mapelID := optUUID(c, "mapel_id")
	tingkat := optInt(c, "kelas_tingkat")
	q := searchParam(c)
	ctx := c.Request().Context()
	var total int64
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM kisi_kisi
		 WHERE ($1::uuid IS NULL OR mapel_id = $1)
		   AND ($2::int IS NULL OR kelas_tingkat = $2)
		   AND ($3::text IS NULL OR kelompok ILIKE '%' || $3 || '%')`,
		mapelID, tingkat, q).Scan(&total); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca kisi-kisi.")
	}
	rows, err := pool.Query(ctx,
		`SELECT id, mapel_id, kelas_tingkat, kelompok, kompetensi_id, jumlah_soal, created_at, updated_at FROM kisi_kisi
		 WHERE ($1::uuid IS NULL OR mapel_id = $1)
		   AND ($2::int IS NULL OR kelas_tingkat = $2)
		   AND ($3::text IS NULL OR kelompok ILIKE '%' || $3 || '%')
		 ORDER BY kelas_tingkat ASC, kelompok ASC LIMIT $4 OFFSET $5`,
		mapelID, tingkat, q, perPage, offset)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca kisi-kisi.")
	}
	defer rows.Close()
	items := []kisiKisiDTO{}
	for rows.Next() {
		var d kisiKisiDTO
		var kompetensiIDV *string
		if err := rows.Scan(&d.ID, &d.MapelID, &d.KelasTingkat, &d.Kelompok, &kompetensiIDV, &d.JumlahSoal, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca kisi-kisi.")
		}
		d.KompetensiID = kompetensiIDV
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca kisi-kisi.")
	}
	return writeList(c, items, page, perPage, total)
}

func getKisiKisi(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.mapel.read") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat kisi-kisi.")
	}
	id := c.Param("id")
	if !validUUID(id) {
		return writeError(c, http.StatusNotFound, "not_found", "Kisi-kisi tidak ditemukan.")
	}
	var d kisiKisiDTO
	var kompetensiIDV *string
	err := pool.QueryRow(c.Request().Context(),
		`SELECT id, mapel_id, kelas_tingkat, kelompok, kompetensi_id, jumlah_soal, created_at, updated_at FROM kisi_kisi WHERE id = $1`, id).
		Scan(&d.ID, &d.MapelID, &d.KelasTingkat, &d.Kelompok, &kompetensiIDV, &d.JumlahSoal, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return writeError(c, http.StatusNotFound, "not_found", "Kisi-kisi tidak ditemukan.")
	}
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca kisi-kisi.")
	}
	d.KompetensiID = kompetensiIDV
	return writeOK(c, d)
}

type kisiKisiInput struct {
	MapelID      string  `json:"mapel_id"`
	KelasTingkat int32   `json:"kelas_tingkat"`
	Kelompok     string  `json:"kelompok"`
	KompetensiID *string `json:"kompetensi_id"`
	JumlahSoal   int32   `json:"jumlah_soal"`
}

func (in kisiKisiInput) validate() string {
	if !validUUID(in.MapelID) {
		return "Mapel wajib dipilih."
	}
	if in.KelasTingkat != 10 && in.KelasTingkat != 11 && in.KelasTingkat != 12 {
		return "Kelas tingkat harus 10, 11, atau 12."
	}
	if strings.TrimSpace(in.Kelompok) == "" {
		return "Kelompok kisi-kisi wajib diisi."
	}
	if in.KompetensiID != nil && strings.TrimSpace(*in.KompetensiID) != "" && !validUUID(*in.KompetensiID) {
		return "kompetensi_id tidak valid."
	}
	if in.JumlahSoal <= 0 {
		return "Jumlah soal harus lebih dari 0."
	}
	return ""
}

func kisiKompetensiArg(in kisiKisiInput) *string {
	if in.KompetensiID == nil {
		return nil
	}
	if strings.TrimSpace(*in.KompetensiID) == "" {
		return nil
	}
	return in.KompetensiID
}

func createKisiKisi(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.mapel.create") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menambah kisi-kisi.")
	}
	var in kisiKisiInput
	if err := c.Bind(&in); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	if msg := in.validate(); msg != "" {
		return writeError(c, http.StatusUnprocessableEntity, "validation_error", msg)
	}
	var d kisiKisiDTO
	var kompetensiIDV *string
	err := pool.QueryRow(c.Request().Context(),
		`INSERT INTO kisi_kisi (mapel_id, kelas_tingkat, kelompok, kompetensi_id, jumlah_soal)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, mapel_id, kelas_tingkat, kelompok, kompetensi_id, jumlah_soal, created_at, updated_at`,
		in.MapelID, in.KelasTingkat, strings.TrimSpace(in.Kelompok), kisiKompetensiArg(in), in.JumlahSoal).
		Scan(&d.ID, &d.MapelID, &d.KelasTingkat, &d.Kelompok, &kompetensiIDV, &d.JumlahSoal, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if isFKViolation(err) {
			return writeError(c, http.StatusUnprocessableEntity, "validation_error", "Mapel/kompetensi tidak ditemukan.")
		}
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menambah kisi-kisi.")
	}
	d.KompetensiID = kompetensiIDV
	catat(c, pool, "create_kisi_kisi", "kisi_kisi", d.ID, map[string]any{"kelompok": d.Kelompok})
	return writeOKStatus(c, http.StatusCreated, d)
}

func updateKisiKisi(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.mapel.update") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mengubah kisi-kisi.")
	}
	id := c.Param("id")
	if !validUUID(id) {
		return writeError(c, http.StatusNotFound, "not_found", "Kisi-kisi tidak ditemukan.")
	}
	var in kisiKisiInput
	if err := c.Bind(&in); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	if msg := in.validate(); msg != "" {
		return writeError(c, http.StatusUnprocessableEntity, "validation_error", msg)
	}
	var d kisiKisiDTO
	var kompetensiIDV *string
	err := pool.QueryRow(c.Request().Context(),
		`UPDATE kisi_kisi SET mapel_id = $2, kelas_tingkat = $3, kelompok = $4, kompetensi_id = $5,
		   jumlah_soal = $6, updated_at = now()
		 WHERE id = $1
		 RETURNING id, mapel_id, kelas_tingkat, kelompok, kompetensi_id, jumlah_soal, created_at, updated_at`,
		id, in.MapelID, in.KelasTingkat, strings.TrimSpace(in.Kelompok), kisiKompetensiArg(in), in.JumlahSoal).
		Scan(&d.ID, &d.MapelID, &d.KelasTingkat, &d.Kelompok, &kompetensiIDV, &d.JumlahSoal, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return writeError(c, http.StatusNotFound, "not_found", "Kisi-kisi tidak ditemukan.")
	}
	if err != nil {
		if isFKViolation(err) {
			return writeError(c, http.StatusUnprocessableEntity, "validation_error", "Mapel/kompetensi tidak ditemukan.")
		}
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mengubah kisi-kisi.")
	}
	d.KompetensiID = kompetensiIDV
	catat(c, pool, "update_kisi_kisi", "kisi_kisi", d.ID, map[string]any{"kelompok": d.Kelompok})
	return writeOK(c, d)
}

func deleteKisiKisi(c *echo.Context, pool *pgxpool.Pool) error {
	if !audit.HasPermission(c, "master.mapel.delete") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menghapus kisi-kisi.")
	}
	id := c.Param("id")
	if !validUUID(id) {
		return writeError(c, http.StatusNotFound, "not_found", "Kisi-kisi tidak ditemukan.")
	}
	ctx := c.Request().Context()
	if n := countOrZero(ctx, pool, `SELECT COUNT(*) FROM soal WHERE kisi_kisi_id = $1`, id); n > 0 {
		return writeError(c, http.StatusConflict, "conflict",
			fmt.Sprintf("Kisi-kisi masih dipakai %d soal. Hapus/ubah soal dulu.", n))
	} else if n < 0 {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa pemakaian kisi-kisi.")
	}
	ct, err := pool.Exec(ctx, `DELETE FROM kisi_kisi WHERE id = $1`, id)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghapus kisi-kisi.")
	}
	if ct.RowsAffected() == 0 {
		return writeError(c, http.StatusNotFound, "not_found", "Kisi-kisi tidak ditemukan.")
	}
	catat(c, pool, "delete_kisi_kisi", "kisi_kisi", id, nil)
	return writeOK(c, map[string]any{"deleted": true, "id": id})
}

// ---------------------------------------------------------------------
// Register
// ---------------------------------------------------------------------

// Register mendaftarkan endpoint Master Data.
//
//	GET/POST /api/v1/master/jurusan, GET/PUT/DELETE /api/v1/master/jurusan/:id
//	GET/POST /api/v1/master/kelas,    GET/PUT/DELETE /api/v1/master/kelas/:id
//	GET/POST /api/v1/master/mapel,    GET/PUT/DELETE /api/v1/master/mapel/:id
//	GET/POST /api/v1/master/guru,     GET/PUT/DELETE /api/v1/master/guru/:id
//	GET/POST /api/v1/master/siswa,    GET/PUT/DELETE /api/v1/master/siswa/:id
//	GET      /api/v1/master/siswa/import (info format CSV; didaftarkan sebelum :id)
//	POST     /api/v1/master/siswa/import
//	GET/POST /api/v1/master/kompetensi, GET/PUT/DELETE /api/v1/master/kompetensi/:id
//	GET/POST /api/v1/master/kisi-kisi,  GET/PUT/DELETE /api/v1/master/kisi-kisi/:id
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q
	_ = cfg

	e.GET("/api/v1/master/jurusan", func(c *echo.Context) error { return listJurusan(c, pool) })
	e.POST("/api/v1/master/jurusan", func(c *echo.Context) error { return createJurusan(c, pool) })
	e.GET("/api/v1/master/jurusan/:id", func(c *echo.Context) error { return getJurusan(c, pool) })
	e.PUT("/api/v1/master/jurusan/:id", func(c *echo.Context) error { return updateJurusan(c, pool) })
	e.DELETE("/api/v1/master/jurusan/:id", func(c *echo.Context) error { return deleteJurusan(c, pool) })

	e.GET("/api/v1/master/kelas", func(c *echo.Context) error { return listKelas(c, pool) })
	e.POST("/api/v1/master/kelas", func(c *echo.Context) error { return createKelas(c, pool) })
	e.GET("/api/v1/master/kelas/:id", func(c *echo.Context) error { return getKelas(c, pool) })
	e.PUT("/api/v1/master/kelas/:id", func(c *echo.Context) error { return updateKelas(c, pool) })
	e.DELETE("/api/v1/master/kelas/:id", func(c *echo.Context) error { return deleteKelas(c, pool) })

	e.GET("/api/v1/master/mapel", func(c *echo.Context) error { return listMapel(c, pool) })
	e.POST("/api/v1/master/mapel", func(c *echo.Context) error { return createMapel(c, pool) })
	e.GET("/api/v1/master/mapel/:id", func(c *echo.Context) error { return getMapel(c, pool) })
	e.PUT("/api/v1/master/mapel/:id", func(c *echo.Context) error { return updateMapel(c, pool) })
	e.DELETE("/api/v1/master/mapel/:id", func(c *echo.Context) error { return deleteMapel(c, pool) })

	e.GET("/api/v1/master/guru", func(c *echo.Context) error { return listGuru(c, pool) })
	e.POST("/api/v1/master/guru", func(c *echo.Context) error { return createGuru(c, pool) })
	e.GET("/api/v1/master/guru/:id", func(c *echo.Context) error { return getGuru(c, pool) })
	e.PUT("/api/v1/master/guru/:id", func(c *echo.Context) error { return updateGuru(c, pool) })
	e.DELETE("/api/v1/master/guru/:id", func(c *echo.Context) error { return deleteGuru(c, pool) })

	e.GET("/api/v1/master/siswa", func(c *echo.Context) error { return listSiswa(c, pool) })
	e.POST("/api/v1/master/siswa", func(c *echo.Context) error { return createSiswa(c, pool) })
	// Route import (GET info + POST upload) SEBELUM route :id agar
	// segmen statis "import" tidak tertangkap param :id.
	e.GET("/api/v1/master/siswa/import", func(c *echo.Context) error { return infoImportSiswa(c, pool) })
	e.POST("/api/v1/master/siswa/import", func(c *echo.Context) error { return importSiswa(c, pool) })
	e.GET("/api/v1/master/siswa/:id", func(c *echo.Context) error { return getSiswa(c, pool) })
	e.PUT("/api/v1/master/siswa/:id", func(c *echo.Context) error { return updateSiswa(c, pool) })
	e.DELETE("/api/v1/master/siswa/:id", func(c *echo.Context) error { return deleteSiswa(c, pool) })

	e.GET("/api/v1/master/kompetensi", func(c *echo.Context) error { return listKompetensi(c, pool) })
	e.POST("/api/v1/master/kompetensi", func(c *echo.Context) error { return createKompetensi(c, pool) })
	e.GET("/api/v1/master/kompetensi/:id", func(c *echo.Context) error { return getKompetensi(c, pool) })
	e.PUT("/api/v1/master/kompetensi/:id", func(c *echo.Context) error { return updateKompetensi(c, pool) })
	e.DELETE("/api/v1/master/kompetensi/:id", func(c *echo.Context) error { return deleteKompetensi(c, pool) })

	e.GET("/api/v1/master/kisi-kisi", func(c *echo.Context) error { return listKisiKisi(c, pool) })
	e.POST("/api/v1/master/kisi-kisi", func(c *echo.Context) error { return createKisiKisi(c, pool) })
	e.GET("/api/v1/master/kisi-kisi/:id", func(c *echo.Context) error { return getKisiKisi(c, pool) })
	e.PUT("/api/v1/master/kisi-kisi/:id", func(c *echo.Context) error { return updateKisiKisi(c, pool) })
	e.DELETE("/api/v1/master/kisi-kisi/:id", func(c *echo.Context) error { return deleteKisiKisi(c, pool) })
}
