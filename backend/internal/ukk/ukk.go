// Package ukk — Uji Kompetensi Keahlian (Tim Fase 2).
//
// Entitas UKK sendiri (ukk_ujian) yang merujuk mapel + kelas + jurusan.
// Nilai teori BOLEH diisi manual lewat endpoint nilai; penarikan otomatis
// dari attempt ujian teori TIDAK dipaksakan (keputusan tugas Fase 2).
//
// Aturan hitung (FINAL, keputusan user — nilai SELALU dihitung backend,
// frontend tidak pernah menjadi sumber nilai):
//
//	nilai_praktik = nilai_proses*0.3 + nilai_produk*0.5 + nilai_sikap*0.2
//	nilai_akhir   = nilai_teori*0.4 + nilai_praktik*0.6
//	lulus: nilai_akhir >= 70
//	predikat: <70 Belum Tuntas, 70-79 Cukup, 80-90 Baik, 91-100 Sangat Baik
//
// Satu format rubrik + satu format sertifikat untuk semua jurusan.
// Rubrik praktik: Proses 30% / Produk 50% / Sikap 20%.
// Sertifikat HANYA bagi yang lulus; nomor format
// UKK/<tahun>/<kode-jurusan>/<nomor-urut 4 digit>.
// Sertifikat di CBT adalah DATA + draf, bukan penerbitan hukum BNSP/LSP.
//
// Endpoint (prefix /api/v1/ukk):
//
//	GET    /ukk                                   (ukk.read)
//	POST   /ukk                                   (ukk.create)
//	GET    /ukk/:id                               (ukk.read)
//	PUT    /ukk/:id                               (ukk.update)
//	DELETE /ukk/:id                               (ukk.update; tidak ada ukk.delete di seed)
//	GET    /ukk/:id/nilai                         (ukk.read)
//	POST   /ukk/:id/nilai                         (ukk.update, upsert per siswa)
//	PUT    /ukk/:id/nilai/:nilaiId               (ukk.update)
//	GET    /ukk/:id/berkas                        (ukk.read; siswa: hanya miliknya)
//	POST   /ukk/:id/berkas                        (ukk.create; siswa: hanya untuk dirinya)
//	GET    /ukk/:id/berkas/:berkasId/download     (ukk.read + audit; siswa: hanya miliknya)
//	POST   /ukk/:id/sertifikat/:siswaId           (ukk.update; hanya bila lulus)
//	GET    /ukk/:id/sertifikat/:siswaId           (ukk.read; siswa: hanya miliknya)
//
// Permission memakai ukk.{read,create,update} dari seed. Siswa TIDAK punya
// permission ukk.* di seed (hanya ujian.kerjakan + rekap.read), tetapi tugas
// Fase 2 menetapkan siswa boleh upload berkas praktik untuk dirinya
// sendiri: untuk endpoint upload (dan download/lihat sertifikat miliknya)
// role siswa diizinkan KHUSUS untuk dirinya sendiri tanpa ukk.create.
// Nilai 0-100; nilai ditolak bila UKK sudah selesai, kecuali admin
// (dengan catatan audit override).
//
// Berkas praktik: multipart field "file", maks 10 MB, ekstensi
// pdf/doc/docx/jpg/jpeg/png/zip, disimpan di
// ./storage/ukk/<ukk_id>/<siswa_id>/<uuid>_<nama-aman> + dicatat di DB.
// Nama file di disk memakai prefix UUID agar tidak bertabrakan dan tidak
// dapat ditebak; nama asli disimpan di kolom nama_file.
package ukk

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// Batas upload berkas praktik UKK.
const (
	maxBerkasBytes = 10 << 20 // 10 MB
	storageRoot    = "storage"
)

// Ekstensi berkas praktik yang diizinkan (jpeg = alias jpg).
var allowedExt = map[string]bool{
	".pdf": true, ".doc": true, ".docx": true,
	".jpg": true, ".jpeg": true, ".png": true, ".zip": true,
}

var unsafeNameChars = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

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

func writeList(c *echo.Context, items any, limit, offset int, total int64) error {
	return c.JSON(http.StatusOK, map[string]any{
		"data": items,
		"meta": map[string]any{"limit": limit, "offset": offset, "total": total},
	})
}

// requirePerm: 401 bila belum login (context auth kosong), 403 bila tidak
// punya permission. Role admin lolos semua permission (audit.HasPermission).
func requirePerm(c *echo.Context, perm string) (userID, role string, ok bool) {
	uid, r, _, _, ok := audit.CurrentUser(c)
	if !ok {
		_ = writeError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
		return "", "", false
	}
	if !audit.HasPermission(c, perm) {
		_ = writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin: "+perm+".")
		return "", "", false
	}
	return uid, r, true
}

// requireLogin hanya memastikan ada sesi, tanpa permission tertentu.
// Dipakai jalur yang mengizinkan role siswa untuk dirinya sendiri.
func requireLogin(c *echo.Context) (userID, role string, ok bool) {
	uid, r, _, _, ok := audit.CurrentUser(c)
	if !ok {
		_ = writeError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
		return "", "", false
	}
	return uid, r, true
}

func isValidUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, ch := range s {
		switch i {
		case 8, 13, 18, 23:
			if ch != '-' {
				return false
			}
		default:
			if !(ch >= '0' && ch <= '9' || ch >= 'a' && ch <= 'f' || ch >= 'A' && ch <= 'F') {
				return false
			}
		}
	}
	return true
}

func strPtr(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// strPtrOrNil mengubah *string menjadi argumen SQL (nil bila kosong).
func strPtrOrNil(s *string) any {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil
	}
	return *s
}

func parsePagination(c *echo.Context) (limit, offset int) {
	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	offset, err = strconv.Atoi(c.QueryParam("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}
	return limit, offset
}

func clientIP(c *echo.Context) string {
	if ip := c.RealIP(); ip != "" {
		return ip
	}
	return c.Request().RemoteAddr
}

// resolveSiswaID memetakan user login ke siswa.id via siswa.user_id.
// Mengembalikan ("", false) bila tidak ada profil siswa untuk user ini.
// (Kontrak: auth_siswa_id = auth_user_id sampai Tim 2 memetakan profil;
// pemetaan eksplisit lewat tabel siswa lebih tepat untuk UKK.)
func resolveSiswaID(ctx context.Context, pool *pgxpool.Pool, userID string) (string, bool) {
	var id string
	err := pool.QueryRow(ctx, `SELECT id FROM siswa WHERE user_id = $1`, userID).Scan(&id)
	if err != nil {
		return "", false
	}
	return id, true
}

// ---------------------------------------------------------------------
// Aturan hitung UKK (satu-satunya tempat rumus ini ditulis)
// ---------------------------------------------------------------------

// round2 membulatkan ke 2 desimal agar nilai tersimpan konsisten.
func round2(v float64) float64 { return math.Round(v*100) / 100 }

// hitungPraktik: proses*0.3 + produk*0.5 + sikap*0.2.
// Mengembalikan nil bila ada satu komponen rubrik yang belum diisi.
func hitungPraktik(proses, produk, sikap *float64) *float64 {
	if proses == nil || produk == nil || sikap == nil {
		return nil
	}
	v := round2(*proses*0.3 + *produk*0.5 + *sikap*0.2)
	return &v
}

// hitungAkhir: teori*0.4 + praktik*0.6. Nil bila teori/praktik belum lengkap.
func hitungAkhir(teori, praktik *float64) *float64 {
	if teori == nil || praktik == nil {
		return nil
	}
	v := round2(*teori*0.4 + *praktik*0.6)
	return &v
}

// Predikat dari nilai akhir 0-100 (sama dengan rekap: <70 Belum Tuntas,
// 70-79 Cukup, 80-90 Baik, 91-100 Sangat Baik).
func Predikat(nilai float64) string {
	switch {
	case nilai < 70:
		return "Belum Tuntas"
	case nilai < 80:
		return "Cukup"
	case nilai <= 90:
		return "Baik"
	default:
		return "Sangat Baik"
	}
}

// validNilai: nilai 0-100, bukan NaN/Inf.
func validNilai(v *float64) bool {
	if v == nil {
		return true
	}
	return !math.IsNaN(*v) && !math.IsInf(*v, 0) && *v >= 0 && *v <= 100
}

// ---------------------------------------------------------------------
// Model JSON
// ---------------------------------------------------------------------

type ukkRow struct {
	ID               string     `json:"id"`
	Nama             string     `json:"nama"`
	MapelID          string     `json:"mapel_id"`
	KelasID          string     `json:"kelas_id"`
	JurusanID        string     `json:"jurusan_id"`
	Tanggal          *time.Time `json:"tanggal"`
	PengujiInternal  *string    `json:"penguji_internal"`
	PengujiEksternal *string    `json:"penguji_eksternal"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	MapelNama        *string    `json:"mapel_nama,omitempty"`
	KelasNama        *string    `json:"kelas_nama,omitempty"`
	JurusanKode      *string    `json:"jurusan_kode,omitempty"`
	JurusanNama      *string    `json:"jurusan_nama,omitempty"`
	JumlahPeserta    *int64     `json:"jumlah_peserta,omitempty"`
}

type nilaiRow struct {
	ID           string    `json:"id"`
	UkkID        string    `json:"ukk_id"`
	SiswaID      string    `json:"siswa_id"`
	NilaiTeori   *float64  `json:"nilai_teori"`
	NilaiProses  *float64  `json:"nilai_proses"`
	NilaiProduk  *float64  `json:"nilai_produk"`
	NilaiSikap   *float64  `json:"nilai_sikap"`
	NilaiPraktik *float64  `json:"nilai_praktik"`
	NilaiAkhir   *float64  `json:"nilai_akhir"`
	Lulus        *bool     `json:"lulus"`
	Predikat     *string   `json:"predikat"`
	Catatan      *string   `json:"catatan"`
	DinilaiOleh  *string   `json:"dinilai_oleh"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	SiswaNISN    *string   `json:"siswa_nisn,omitempty"`
	SiswaNama    *string   `json:"siswa_nama,omitempty"`
}

type berkasRow struct {
	ID           string    `json:"id"`
	UkkID        string    `json:"ukk_id"`
	SiswaID      string    `json:"siswa_id"`
	NamaFile     string    `json:"nama_file"`
	Ukuran       int       `json:"ukuran"`
	DiuploadOleh *string   `json:"diupload_oleh"`
	CreatedAt    time.Time `json:"created_at"`
	SiswaNISN    *string   `json:"siswa_nisn,omitempty"`
	SiswaNama    *string   `json:"siswa_nama,omitempty"`
}

type sertifikatRow struct {
	ID               string     `json:"id"`
	UkkID            string     `json:"ukk_id"`
	SiswaID          string     `json:"siswa_id"`
	NomorSertifikat  string     `json:"nomor_sertifikat"`
	DiterbitkanAt    time.Time  `json:"diterbitkan_at"`
	SiswaNISN        *string    `json:"siswa_nisn,omitempty"`
	SiswaNama        *string    `json:"siswa_nama,omitempty"`
	UkkNama          *string    `json:"ukk_nama,omitempty"`
	UkkTanggal       *time.Time `json:"ukk_tanggal,omitempty"`
	PengujiInternal  *string    `json:"penguji_internal,omitempty"`
	PengujiEksternal *string    `json:"penguji_eksternal,omitempty"`
	JurusanKode      *string    `json:"jurusan_kode,omitempty"`
	JurusanNama      *string    `json:"jurusan_nama,omitempty"`
	NilaiTeori       *float64   `json:"nilai_teori,omitempty"`
	NilaiPraktik     *float64   `json:"nilai_praktik,omitempty"`
	NilaiAkhir       *float64   `json:"nilai_akhir,omitempty"`
	Predikat         *string    `json:"predikat,omitempty"`
}

// ---------------------------------------------------------------------
// Pembacaan UKK
// ---------------------------------------------------------------------

func getUkkRow(ctx context.Context, pool *pgxpool.Pool, id string) (*ukkRow, error) {
	var u ukkRow
	err := pool.QueryRow(ctx,
		`SELECT u.id, u.nama, u.mapel_id, u.kelas_id, u.jurusan_id, u.tanggal,
		        u.penguji_internal, u.penguji_eksternal, u.status, u.created_at,
		        m.nama, k.nama, j.kode, j.nama
		 FROM ukk_ujian u
		 JOIN mapel m ON m.id = u.mapel_id
		 JOIN kelas k ON k.id = u.kelas_id
		 JOIN jurusan j ON j.id = u.jurusan_id
		 WHERE u.id = $1`, id).Scan(
		&u.ID, &u.Nama, &u.MapelID, &u.KelasID, &u.JurusanID, &u.Tanggal,
		&u.PengujiInternal, &u.PengujiEksternal, &u.Status, &u.CreatedAt,
		&u.MapelNama, &u.KelasNama, &u.JurusanKode, &u.JurusanNama)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func existsID(ctx context.Context, pool *pgxpool.Pool, table, id string) bool {
	var n int
	// table berasal dari konstanta internal, bukan input user.
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM `+table+` WHERE id = $1`, id).Scan(&n)
	return n > 0
}

// ---------------------------------------------------------------------
// Register
// ---------------------------------------------------------------------

func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q
	_ = cfg

	e.GET("/api/v1/ukk", func(c *echo.Context) error { return listUkk(c, pool) })
	e.POST("/api/v1/ukk", func(c *echo.Context) error { return createUkk(c, pool) })
	e.GET("/api/v1/ukk/:id", func(c *echo.Context) error { return getUkk(c, pool) })
	e.PUT("/api/v1/ukk/:id", func(c *echo.Context) error { return updateUkk(c, pool) })
	e.DELETE("/api/v1/ukk/:id", func(c *echo.Context) error { return deleteUkk(c, pool) })

	e.GET("/api/v1/ukk/:id/nilai", func(c *echo.Context) error { return listNilai(c, pool) })
	e.POST("/api/v1/ukk/:id/nilai", func(c *echo.Context) error { return upsertNilai(c, pool, "") })
	e.PUT("/api/v1/ukk/:id/nilai/:nilaiId", func(c *echo.Context) error {
		return upsertNilai(c, pool, c.Param("nilaiId"))
	})

	e.GET("/api/v1/ukk/:id/berkas", func(c *echo.Context) error { return listBerkas(c, pool) })
	e.POST("/api/v1/ukk/:id/berkas", func(c *echo.Context) error { return uploadBerkas(c, pool) })
	e.GET("/api/v1/ukk/:id/berkas/:berkasId/download", func(c *echo.Context) error {
		return downloadBerkas(c, pool)
	})

	e.POST("/api/v1/ukk/:id/sertifikat/:siswaId", func(c *echo.Context) error {
		return createSertifikat(c, pool)
	})
	e.GET("/api/v1/ukk/:id/sertifikat/:siswaId", func(c *echo.Context) error {
		return getSertifikat(c, pool)
	})
}

// ---------------------------------------------------------------------
// CRUD ukk_ujian
// ---------------------------------------------------------------------

type ukkInput struct {
	Nama             string  `json:"nama"`
	MapelID          string  `json:"mapel_id"`
	KelasID          string  `json:"kelas_id"`
	JurusanID        string  `json:"jurusan_id"`
	Tanggal          *string `json:"tanggal"`
	PengujiInternal  *string `json:"penguji_internal"`
	PengujiEksternal *string `json:"penguji_eksternal"`
	Status           *string `json:"status"`
}

func (in ukkInput) validate() string {
	if strings.TrimSpace(in.Nama) == "" {
		return "Nama UKK wajib diisi."
	}
	if !isValidUUID(in.MapelID) {
		return "mapel_id tidak valid."
	}
	if !isValidUUID(in.KelasID) {
		return "kelas_id tidak valid."
	}
	if !isValidUUID(in.JurusanID) {
		return "jurusan_id tidak valid."
	}
	if in.Status != nil && *in.Status != "draft" && *in.Status != "berjalan" && *in.Status != "selesai" {
		return "status harus draft, berjalan, atau selesai."
	}
	if in.Tanggal != nil && strings.TrimSpace(*in.Tanggal) != "" {
		if _, err := time.Parse("2006-01-02", strings.TrimSpace(*in.Tanggal)); err != nil {
			return "tanggal harus format YYYY-MM-DD."
		}
	}
	return ""
}

func (in ukkInput) tanggalArg() any {
	if in.Tanggal == nil || strings.TrimSpace(*in.Tanggal) == "" {
		return nil
	}
	t, _ := time.Parse("2006-01-02", strings.TrimSpace(*in.Tanggal))
	return t
}

func listUkk(c *echo.Context, pool *pgxpool.Pool) error {
	if _, _, ok := requirePerm(c, "ukk.read"); !ok {
		return nil
	}
	limit, offset := parsePagination(c)
	jurusanID := strPtr(c.QueryParam("jurusan_id"))
	kelasID := strPtr(c.QueryParam("kelas_id"))
	status := strPtr(c.QueryParam("status"))
	search := strPtr(c.QueryParam("q"))
	ctx := c.Request().Context()

	var total int64
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM ukk_ujian u
		 WHERE ($1::uuid IS NULL OR u.jurusan_id = $1)
		   AND ($2::uuid IS NULL OR u.kelas_id = $2)
		   AND ($3::text IS NULL OR u.status = $3)
		   AND ($4::text IS NULL OR u.nama ILIKE '%' || $4 || '%')`,
		jurusanID, kelasID, status, search).Scan(&total); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca daftar UKK.")
	}
	rows, err := pool.Query(ctx,
		`SELECT u.id, u.nama, u.mapel_id, u.kelas_id, u.jurusan_id, u.tanggal,
		        u.penguji_internal, u.penguji_eksternal, u.status, u.created_at,
		        m.nama, k.nama, j.kode, j.nama,
		        (SELECT COUNT(*) FROM ukk_nilai n WHERE n.ukk_id = u.id)
		 FROM ukk_ujian u
		 JOIN mapel m ON m.id = u.mapel_id
		 JOIN kelas k ON k.id = u.kelas_id
		 JOIN jurusan j ON j.id = u.jurusan_id
		 WHERE ($1::uuid IS NULL OR u.jurusan_id = $1)
		   AND ($2::uuid IS NULL OR u.kelas_id = $2)
		   AND ($3::text IS NULL OR u.status = $3)
		   AND ($4::text IS NULL OR u.nama ILIKE '%' || $4 || '%')
		 ORDER BY u.created_at DESC
		 LIMIT $5 OFFSET $6`,
		jurusanID, kelasID, status, search, limit, offset)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca daftar UKK.")
	}
	defer rows.Close()
	items := []ukkRow{}
	for rows.Next() {
		var u ukkRow
		var jumlah int64
		if err := rows.Scan(&u.ID, &u.Nama, &u.MapelID, &u.KelasID, &u.JurusanID,
			&u.Tanggal, &u.PengujiInternal, &u.PengujiEksternal, &u.Status, &u.CreatedAt,
			&u.MapelNama, &u.KelasNama, &u.JurusanKode, &u.JurusanNama, &jumlah); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca daftar UKK.")
		}
		u.JumlahPeserta = &jumlah
		items = append(items, u)
	}
	if err := rows.Err(); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca daftar UKK.")
	}
	return writeList(c, items, limit, offset, total)
}

func createUkk(c *echo.Context, pool *pgxpool.Pool) error {
	userID, _, ok := requirePerm(c, "ukk.create")
	if !ok {
		return nil
	}
	var in ukkInput
	if err := c.Bind(&in); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	if msg := in.validate(); msg != "" {
		return writeError(c, http.StatusBadRequest, "validation_error", msg)
	}
	ctx := c.Request().Context()
	if !existsID(ctx, pool, "mapel", in.MapelID) {
		return writeError(c, http.StatusBadRequest, "validation_error", "Mapel tidak ditemukan.")
	}
	if !existsID(ctx, pool, "kelas", in.KelasID) {
		return writeError(c, http.StatusBadRequest, "validation_error", "Kelas tidak ditemukan.")
	}
	if !existsID(ctx, pool, "jurusan", in.JurusanID) {
		return writeError(c, http.StatusBadRequest, "validation_error", "Jurusan tidak ditemukan.")
	}
	status := "draft"
	if in.Status != nil && *in.Status != "" {
		status = *in.Status
	}
	var id string
	err := pool.QueryRow(ctx,
		`INSERT INTO ukk_ujian (nama, mapel_id, kelas_id, jurusan_id, tanggal, penguji_internal, penguji_eksternal, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		strings.TrimSpace(in.Nama), in.MapelID, in.KelasID, in.JurusanID,
		in.tanggalArg(), strPtrOrNil(in.PengujiInternal), strPtrOrNil(in.PengujiEksternal), status).Scan(&id)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membuat UKK.")
	}
	_ = audit.Catat(ctx, pool, userID, "ukk.buat", "ukk_ujian", id,
		map[string]any{"nama": strings.TrimSpace(in.Nama), "status": status}, clientIP(c))
	u, err := getUkkRow(ctx, pool, id)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca UKK baru.")
	}
	return writeOKStatus(c, http.StatusCreated, u)
}

func getUkk(c *echo.Context, pool *pgxpool.Pool) error {
	if _, _, ok := requirePerm(c, "ukk.read"); !ok {
		return nil
	}
	id := c.Param("id")
	if !isValidUUID(id) {
		return writeError(c, http.StatusBadRequest, "validation_error", "ID UKK tidak valid.")
	}
	u, err := getUkkRow(c.Request().Context(), pool, id)
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "UKK tidak ditemukan.")
	}
	return writeOK(c, u)
}

func updateUkk(c *echo.Context, pool *pgxpool.Pool) error {
	userID, _, ok := requirePerm(c, "ukk.update")
	if !ok {
		return nil
	}
	id := c.Param("id")
	if !isValidUUID(id) {
		return writeError(c, http.StatusBadRequest, "validation_error", "ID UKK tidak valid.")
	}
	var in ukkInput
	if err := c.Bind(&in); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	if msg := in.validate(); msg != "" {
		return writeError(c, http.StatusBadRequest, "validation_error", msg)
	}
	ctx := c.Request().Context()
	old, err := getUkkRow(ctx, pool, id)
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "UKK tidak ditemukan.")
	}
	if !existsID(ctx, pool, "mapel", in.MapelID) {
		return writeError(c, http.StatusBadRequest, "validation_error", "Mapel tidak ditemukan.")
	}
	if !existsID(ctx, pool, "kelas", in.KelasID) {
		return writeError(c, http.StatusBadRequest, "validation_error", "Kelas tidak ditemukan.")
	}
	if !existsID(ctx, pool, "jurusan", in.JurusanID) {
		return writeError(c, http.StatusBadRequest, "validation_error", "Jurusan tidak ditemukan.")
	}
	status := old.Status
	if in.Status != nil && *in.Status != "" {
		status = *in.Status
	}
	_, err = pool.Exec(ctx,
		`UPDATE ukk_ujian SET nama = $2, mapel_id = $3, kelas_id = $4, jurusan_id = $5,
		        tanggal = $6, penguji_internal = $7, penguji_eksternal = $8, status = $9
		 WHERE id = $1`,
		id, strings.TrimSpace(in.Nama), in.MapelID, in.KelasID, in.JurusanID,
		in.tanggalArg(), strPtrOrNil(in.PengujiInternal), strPtrOrNil(in.PengujiEksternal), status)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mengubah UKK.")
	}
	_ = audit.Catat(ctx, pool, userID, "ukk.ubah", "ukk_ujian", id,
		map[string]any{"nama": strings.TrimSpace(in.Nama), "status_lama": old.Status, "status_baru": status}, clientIP(c))
	u, err := getUkkRow(ctx, pool, id)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca UKK.")
	}
	return writeOK(c, u)
}

func deleteUkk(c *echo.Context, pool *pgxpool.Pool) error {
	userID, _, ok := requirePerm(c, "ukk.update")
	if !ok {
		return nil
	}
	id := c.Param("id")
	if !isValidUUID(id) {
		return writeError(c, http.StatusBadRequest, "validation_error", "ID UKK tidak valid.")
	}
	ctx := c.Request().Context()
	if _, err := getUkkRow(ctx, pool, id); err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "UKK tidak ditemukan.")
	}
	var nNilai, nBerkas, nSert int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM ukk_nilai WHERE ukk_id = $1`, id).Scan(&nNilai)
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM ukk_berkas WHERE ukk_id = $1`, id).Scan(&nBerkas)
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM ukk_sertifikat WHERE ukk_id = $1`, id).Scan(&nSert)
	if nNilai+nBerkas+nSert > 0 {
		return writeError(c, http.StatusConflict, "conflict",
			"UKK tidak dapat dihapus karena sudah memiliki nilai/berkas/sertifikat.")
	}
	if _, err := pool.Exec(ctx, `DELETE FROM ukk_ujian WHERE id = $1`, id); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghapus UKK.")
	}
	_ = audit.Catat(ctx, pool, userID, "ukk.hapus", "ukk_ujian", id, nil, clientIP(c))
	return writeOK(c, map[string]any{"deleted": true, "id": id})
}

// ---------------------------------------------------------------------
// Nilai UKK
// ---------------------------------------------------------------------

type nilaiInput struct {
	SiswaID     string   `json:"siswa_id"`
	NilaiTeori  *float64 `json:"nilai_teori"`
	NilaiProses *float64 `json:"nilai_proses"`
	NilaiProduk *float64 `json:"nilai_produk"`
	NilaiSikap  *float64 `json:"nilai_sikap"`
	Catatan     *string  `json:"catatan"`
}

func (in nilaiInput) validate() string {
	if !validNilai(in.NilaiTeori) || !validNilai(in.NilaiProses) ||
		!validNilai(in.NilaiProduk) || !validNilai(in.NilaiSikap) {
		return "Semua nilai harus 0-100."
	}
	return ""
}

// hitungNilaiInput menghitung nilai turunan dari input mentah.
func hitungNilaiInput(in nilaiInput) (praktik, akhir *float64, lulus *bool, predikat *string) {
	praktik = hitungPraktik(in.NilaiProses, in.NilaiProduk, in.NilaiSikap)
	akhir = hitungAkhir(in.NilaiTeori, praktik)
	if akhir != nil {
		l := *akhir >= 70
		lulus = &l
		p := Predikat(*akhir)
		predikat = &p
	}
	return praktik, akhir, lulus, predikat
}

func listNilai(c *echo.Context, pool *pgxpool.Pool) error {
	if _, _, ok := requirePerm(c, "ukk.read"); !ok {
		return nil
	}
	id := c.Param("id")
	if !isValidUUID(id) {
		return writeError(c, http.StatusBadRequest, "validation_error", "ID UKK tidak valid.")
	}
	ctx := c.Request().Context()
	if _, err := getUkkRow(ctx, pool, id); err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "UKK tidak ditemukan.")
	}
	rows, err := pool.Query(ctx,
		`SELECT n.id, n.ukk_id, n.siswa_id, n.nilai_teori, n.nilai_proses,
		        n.nilai_produk, n.nilai_sikap, n.nilai_praktik, n.nilai_akhir,
		        n.lulus, n.predikat, n.catatan, n.dinilai_oleh,
		        n.created_at, n.updated_at, s.nisn, s.nama
		 FROM ukk_nilai n JOIN siswa s ON s.id = n.siswa_id
		 WHERE n.ukk_id = $1 ORDER BY s.nama ASC`, id)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca nilai UKK.")
	}
	defer rows.Close()
	items := []nilaiRow{}
	for rows.Next() {
		var n nilaiRow
		if err := rows.Scan(&n.ID, &n.UkkID, &n.SiswaID, &n.NilaiTeori, &n.NilaiProses,
			&n.NilaiProduk, &n.NilaiSikap, &n.NilaiPraktik, &n.NilaiAkhir,
			&n.Lulus, &n.Predikat, &n.Catatan, &n.DinilaiOleh,
			&n.CreatedAt, &n.UpdatedAt, &n.SiswaNISN, &n.SiswaNama); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca nilai UKK.")
		}
		items = append(items, n)
	}
	if err := rows.Err(); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca nilai UKK.")
	}
	return writeOK(c, items)
}

// upsertNilai menangani POST /ukk/:id/nilai (nilaiID kosong, upsert per
// siswa) dan PUT /ukk/:id/nilai/:nilaiId (per nilai id).
func upsertNilai(c *echo.Context, pool *pgxpool.Pool, nilaiID string) error {
	userID, role, ok := requirePerm(c, "ukk.update")
	if !ok {
		return nil
	}
	ukkID := c.Param("id")
	if !isValidUUID(ukkID) {
		return writeError(c, http.StatusBadRequest, "validation_error", "ID UKK tidak valid.")
	}
	if nilaiID != "" && !isValidUUID(nilaiID) {
		return writeError(c, http.StatusBadRequest, "validation_error", "ID nilai tidak valid.")
	}
	var in nilaiInput
	if err := c.Bind(&in); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	if msg := in.validate(); msg != "" {
		return writeError(c, http.StatusBadRequest, "validation_error", msg)
	}
	ctx := c.Request().Context()
	ukk, err := getUkkRow(ctx, pool, ukkID)
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "UKK tidak ditemukan.")
	}
	// Nilai ditolak bila UKK sudah selesai, kecuali admin (+ audit override).
	if ukk.Status == "selesai" && role != "admin" {
		return writeError(c, http.StatusForbidden, "forbidden",
			"UKK sudah selesai: nilai tidak dapat diubah kecuali oleh admin.")
	}
	override := ukk.Status == "selesai" && role == "admin"

	praktik, akhir, lulus, predikat := hitungNilaiInput(in)
	var row nilaiRow
	var created bool
	if nilaiID != "" {
		// PUT per nilai id: siswa_id diambil dari baris yang ada.
		var existingSiswaID string
		err := pool.QueryRow(ctx,
			`SELECT siswa_id FROM ukk_nilai WHERE id = $1 AND ukk_id = $2`,
			nilaiID, ukkID).Scan(&existingSiswaID)
		if err != nil {
			return writeError(c, http.StatusNotFound, "not_found", "Nilai UKK tidak ditemukan.")
		}
		in.SiswaID = existingSiswaID
		err = pool.QueryRow(ctx,
			`UPDATE ukk_nilai SET nilai_teori = $2, nilai_proses = $3, nilai_produk = $4,
			        nilai_sikap = $5, nilai_praktik = $6, nilai_akhir = $7, lulus = $8,
			        predikat = $9, catatan = $10, dinilai_oleh = $11, updated_at = now()
			 WHERE id = $1 RETURNING id, ukk_id, siswa_id, nilai_teori, nilai_proses,
			        nilai_produk, nilai_sikap, nilai_praktik, nilai_akhir, lulus,
			        predikat, catatan, dinilai_oleh, created_at, updated_at`,
			nilaiID, in.NilaiTeori, in.NilaiProses, in.NilaiProduk, in.NilaiSikap,
			praktik, akhir, lulus, predikat, strPtrOrNil(in.Catatan), userID).Scan(
			&row.ID, &row.UkkID, &row.SiswaID, &row.NilaiTeori, &row.NilaiProses,
			&row.NilaiProduk, &row.NilaiSikap, &row.NilaiPraktik, &row.NilaiAkhir,
			&row.Lulus, &row.Predikat, &row.Catatan, &row.DinilaiOleh,
			&row.CreatedAt, &row.UpdatedAt)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menyimpan nilai UKK.")
		}
	} else {
		// POST: upsert per (ukk_id, siswa_id).
		if !isValidUUID(in.SiswaID) {
			return writeError(c, http.StatusBadRequest, "validation_error", "siswa_id tidak valid.")
		}
		if !existsID(ctx, pool, "siswa", in.SiswaID) {
			return writeError(c, http.StatusBadRequest, "validation_error", "Siswa tidak ditemukan.")
		}
		var existed bool
		_ = pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM ukk_nilai WHERE ukk_id = $1 AND siswa_id = $2)`,
			ukkID, in.SiswaID).Scan(&existed)
		created = !existed
		err = pool.QueryRow(ctx,
			`INSERT INTO ukk_nilai (ukk_id, siswa_id, nilai_teori, nilai_proses, nilai_produk,
			        nilai_sikap, nilai_praktik, nilai_akhir, lulus, predikat, catatan, dinilai_oleh)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			 ON CONFLICT (ukk_id, siswa_id) DO UPDATE
			 SET nilai_teori = EXCLUDED.nilai_teori, nilai_proses = EXCLUDED.nilai_proses,
			     nilai_produk = EXCLUDED.nilai_produk, nilai_sikap = EXCLUDED.nilai_sikap,
			     nilai_praktik = EXCLUDED.nilai_praktik, nilai_akhir = EXCLUDED.nilai_akhir,
			     lulus = EXCLUDED.lulus, predikat = EXCLUDED.predikat,
			     catatan = EXCLUDED.catatan, dinilai_oleh = EXCLUDED.dinilai_oleh,
			     updated_at = now()
			 RETURNING id, ukk_id, siswa_id, nilai_teori, nilai_proses,
			        nilai_produk, nilai_sikap, nilai_praktik, nilai_akhir, lulus,
			        predikat, catatan, dinilai_oleh, created_at, updated_at`,
			ukkID, in.SiswaID, in.NilaiTeori, in.NilaiProses, in.NilaiProduk, in.NilaiSikap,
			praktik, akhir, lulus, predikat, strPtrOrNil(in.Catatan), userID).Scan(
			&row.ID, &row.UkkID, &row.SiswaID, &row.NilaiTeori, &row.NilaiProses,
			&row.NilaiProduk, &row.NilaiSikap, &row.NilaiPraktik, &row.NilaiAkhir,
			&row.Lulus, &row.Predikat, &row.Catatan, &row.DinilaiOleh,
			&row.CreatedAt, &row.UpdatedAt)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menyimpan nilai UKK.")
		}
	}
	// Lengkapi info siswa untuk respons.
	_ = pool.QueryRow(ctx, `SELECT nisn, nama FROM siswa WHERE id = $1`, row.SiswaID).
		Scan(&row.SiswaNISN, &row.SiswaNama)
	detail := map[string]any{"siswa_id": row.SiswaID, "nilai_akhir": row.NilaiAkhir, "lulus": row.Lulus}
	if override {
		detail["override_ukk_selesai"] = true
	}
	_ = audit.Catat(ctx, pool, userID, "ukk.nilai", "ukk_nilai", row.ID, detail, clientIP(c))
	if nilaiID != "" {
		return writeOK(c, row)
	}
	if created {
		return writeOKStatus(c, http.StatusCreated, row)
	}
	return writeOK(c, row)
}

// ---------------------------------------------------------------------
// Berkas praktik
// ---------------------------------------------------------------------

func listBerkas(c *echo.Context, pool *pgxpool.Pool) error {
	userID, role, ok := requireLogin(c)
	if !ok {
		return nil
	}
	ukkID := c.Param("id")
	if !isValidUUID(ukkID) {
		return writeError(c, http.StatusBadRequest, "validation_error", "ID UKK tidak valid.")
	}
	ctx := c.Request().Context()
	if _, err := getUkkRow(ctx, pool, ukkID); err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "UKK tidak ditemukan.")
	}
	// Siswa hanya melihat berkas miliknya sendiri.
	var filterSiswa any
	if role == "siswa" {
		siswaID, found := resolveSiswaID(ctx, pool, userID)
		if !found {
			return writeError(c, http.StatusForbidden, "forbidden", "Profil siswa tidak ditemukan.")
		}
		filterSiswa = siswaID
	} else if !audit.HasPermission(c, "ukk.read") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin: ukk.read.")
	}
	rows, err := pool.Query(ctx,
		`SELECT b.id, b.ukk_id, b.siswa_id, b.nama_file, b.ukuran,
		        b.diupload_oleh, b.created_at, s.nisn, s.nama
		 FROM ukk_berkas b JOIN siswa s ON s.id = b.siswa_id
		 WHERE b.ukk_id = $1 AND ($2::uuid IS NULL OR b.siswa_id = $2)
		 ORDER BY b.created_at DESC`, ukkID, filterSiswa)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca berkas UKK.")
	}
	defer rows.Close()
	items := []berkasRow{}
	for rows.Next() {
		var b berkasRow
		if err := rows.Scan(&b.ID, &b.UkkID, &b.SiswaID, &b.NamaFile, &b.Ukuran,
			&b.DiuploadOleh, &b.CreatedAt, &b.SiswaNISN, &b.SiswaNama); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca berkas UKK.")
		}
		items = append(items, b)
	}
	if err := rows.Err(); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca berkas UKK.")
	}
	return writeOK(c, items)
}

func sanitizeFileName(name string) string {
	base := filepath.Base(name)
	base = unsafeNameChars.ReplaceAllString(base, "_")
	base = strings.Trim(base, "._")
	if base == "" {
		base = "berkas"
	}
	if len(base) > 120 {
		base = base[:120]
	}
	return base
}

func uploadBerkas(c *echo.Context, pool *pgxpool.Pool) error {
	userID, role, ok := requireLogin(c)
	if !ok {
		return nil
	}
	ukkID := c.Param("id")
	if !isValidUUID(ukkID) {
		return writeError(c, http.StatusBadRequest, "validation_error", "ID UKK tidak valid.")
	}
	ctx := c.Request().Context()
	if _, err := getUkkRow(ctx, pool, ukkID); err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "UKK tidak ditemukan.")
	}
	// Siswa: hanya untuk dirinya sendiri. Selain siswa: wajib ukk.create.
	var callerSiswaID string
	if role == "siswa" {
		sid, found := resolveSiswaID(ctx, pool, userID)
		if !found {
			return writeError(c, http.StatusForbidden, "forbidden", "Profil siswa tidak ditemukan.")
		}
		callerSiswaID = sid
	} else if !audit.HasPermission(c, "ukk.create") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin: ukk.create.")
	}
	if err := c.Request().ParseMultipartForm(maxBerkasBytes); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Body multipart tidak valid (maks. 10 MB).")
	}
	siswaID := strings.TrimSpace(c.FormValue("siswa_id"))
	if role == "siswa" {
		if siswaID == "" {
			siswaID = callerSiswaID
		}
		if siswaID != callerSiswaID {
			return writeError(c, http.StatusForbidden, "forbidden", "Siswa hanya boleh upload berkas untuk dirinya sendiri.")
		}
	} else if !isValidUUID(siswaID) {
		return writeError(c, http.StatusBadRequest, "validation_error", "siswa_id tidak valid.")
	}
	if !existsID(ctx, pool, "siswa", siswaID) {
		return writeError(c, http.StatusBadRequest, "validation_error", "Siswa tidak ditemukan.")
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Field file wajib diisi.")
	}
	if fileHeader.Size > maxBerkasBytes {
		return writeError(c, http.StatusBadRequest, "validation_error", "Ukuran berkas melebihi 10 MB.")
	}
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowedExt[ext] {
		return writeError(c, http.StatusBadRequest, "validation_error",
			"Ekstensi tidak diizinkan. Gunakan: pdf, doc, docx, jpg, png, zip.")
	}
	src, err := fileHeader.Open()
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca berkas upload.")
	}
	defer src.Close()

	safeName := sanitizeFileName(fileHeader.Filename)
	// Prefix unik: cegah tabrakan nama + nama file tidak dapat ditebak.
	diskName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), safeName)
	relDir := filepath.Join(storageRoot, "ukk", ukkID, siswaID)
	absDir, _ := filepath.Abs(relDir)
	if err := os.MkdirAll(absDir, 0o750); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menyiapkan penyimpanan berkas.")
	}
	relPath := filepath.Join(relDir, diskName)
	absPath, _ := filepath.Abs(relPath)
	dst, err := os.Create(absPath)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menyimpan berkas.")
	}
	written, copyErr := io.Copy(dst, src)
	_ = dst.Close()
	if copyErr != nil {
		_ = os.Remove(absPath)
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menyimpan berkas.")
	}
	if written > maxBerkasBytes {
		_ = os.Remove(absPath)
		return writeError(c, http.StatusBadRequest, "validation_error", "Ukuran berkas melebihi 10 MB.")
	}

	var b berkasRow
	err = pool.QueryRow(ctx,
		`INSERT INTO ukk_berkas (ukk_id, siswa_id, nama_file, path, ukuran, diupload_oleh)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, ukk_id, siswa_id, nama_file, ukuran, diupload_oleh, created_at`,
		ukkID, siswaID, fileHeader.Filename, relPath, written, userID).Scan(
		&b.ID, &b.UkkID, &b.SiswaID, &b.NamaFile, &b.Ukuran, &b.DiuploadOleh, &b.CreatedAt)
	if err != nil {
		_ = os.Remove(absPath)
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mencatat berkas UKK.")
	}
	_ = audit.Catat(ctx, pool, userID, "ukk.berkas.upload", "ukk_berkas", b.ID,
		map[string]any{"ukk_id": ukkID, "siswa_id": siswaID, "nama_file": fileHeader.Filename, "ukuran": written}, clientIP(c))
	return writeOKStatus(c, http.StatusCreated, b)
}

func downloadBerkas(c *echo.Context, pool *pgxpool.Pool) error {
	userID, role, ok := requireLogin(c)
	if !ok {
		return nil
	}
	ukkID := c.Param("id")
	berkasID := c.Param("berkasId")
	if !isValidUUID(ukkID) || !isValidUUID(berkasID) {
		return writeError(c, http.StatusBadRequest, "validation_error", "ID tidak valid.")
	}
	ctx := c.Request().Context()
	if _, err := getUkkRow(ctx, pool, ukkID); err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "UKK tidak ditemukan.")
	}
	var b berkasRow
	var storedPath string
	err := pool.QueryRow(ctx,
		`SELECT id, ukk_id, siswa_id, nama_file, ukuran, diupload_oleh, created_at, path
		 FROM ukk_berkas WHERE id = $1 AND ukk_id = $2`,
		berkasID, ukkID).Scan(&b.ID, &b.UkkID, &b.SiswaID, &b.NamaFile, &b.Ukuran,
		&b.DiuploadOleh, &b.CreatedAt, &storedPath)
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Berkas tidak ditemukan.")
	}
	// Siswa hanya boleh mengunduh berkas miliknya sendiri.
	if role == "siswa" {
		sid, found := resolveSiswaID(ctx, pool, userID)
		if !found || sid != b.SiswaID {
			return writeError(c, http.StatusNotFound, "not_found", "Berkas tidak ditemukan.")
		}
	} else if !audit.HasPermission(c, "ukk.read") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin: ukk.read.")
	}
	// Keamanan path: file harus berada di dalam ./storage/ukk/.
	absRoot, _ := filepath.Abs(filepath.Join(storageRoot, "ukk"))
	absPath, _ := filepath.Abs(storedPath)
	if absPath != absRoot && !strings.HasPrefix(absPath, absRoot+string(os.PathSeparator)) {
		return writeError(c, http.StatusNotFound, "not_found", "Berkas tidak ditemukan.")
	}
	if _, err := os.Stat(absPath); err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "File berkas tidak ditemukan di penyimpanan.")
	}
	_ = audit.Catat(ctx, pool, userID, "ukk.berkas.download", "ukk_berkas", b.ID,
		map[string]any{"ukk_id": ukkID, "siswa_id": b.SiswaID, "nama_file": b.NamaFile}, clientIP(c))
	return c.Attachment(absPath, b.NamaFile)
}

// ---------------------------------------------------------------------
// Sertifikat
// ---------------------------------------------------------------------

func getSertifikatRow(ctx context.Context, pool *pgxpool.Pool, ukkID, siswaID string) (*sertifikatRow, error) {
	var s sertifikatRow
	err := pool.QueryRow(ctx,
		`SELECT s.id, s.ukk_id, s.siswa_id, s.nomor_sertifikat, s.diterbitkan_at,
		        si.nisn, si.nama, u.nama, u.tanggal, u.penguji_internal, u.penguji_eksternal,
		        j.kode, j.nama, n.nilai_teori, n.nilai_praktik, n.nilai_akhir, n.predikat
		 FROM ukk_sertifikat s
		 JOIN siswa si ON si.id = s.siswa_id
		 JOIN ukk_ujian u ON u.id = s.ukk_id
		 JOIN jurusan j ON j.id = u.jurusan_id
		 LEFT JOIN ukk_nilai n ON n.ukk_id = s.ukk_id AND n.siswa_id = s.siswa_id
		 WHERE s.ukk_id = $1 AND s.siswa_id = $2`,
		ukkID, siswaID).Scan(
		&s.ID, &s.UkkID, &s.SiswaID, &s.NomorSertifikat, &s.DiterbitkanAt,
		&s.SiswaNISN, &s.SiswaNama, &s.UkkNama, &s.UkkTanggal, &s.PengujiInternal,
		&s.PengujiEksternal, &s.JurusanKode, &s.JurusanNama,
		&s.NilaiTeori, &s.NilaiPraktik, &s.NilaiAkhir, &s.Predikat)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func getSertifikat(c *echo.Context, pool *pgxpool.Pool) error {
	userID, role, ok := requireLogin(c)
	if !ok {
		return nil
	}
	ukkID := c.Param("id")
	siswaID := c.Param("siswaId")
	if !isValidUUID(ukkID) || !isValidUUID(siswaID) {
		return writeError(c, http.StatusBadRequest, "validation_error", "ID tidak valid.")
	}
	ctx := c.Request().Context()
	if _, err := getUkkRow(ctx, pool, ukkID); err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "UKK tidak ditemukan.")
	}
	if role == "siswa" {
		sid, found := resolveSiswaID(ctx, pool, userID)
		if !found || sid != siswaID {
			return writeError(c, http.StatusNotFound, "not_found", "Sertifikat tidak ditemukan.")
		}
	} else if !audit.HasPermission(c, "ukk.read") {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin: ukk.read.")
	}
	s, err := getSertifikatRow(ctx, pool, ukkID, siswaID)
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Sertifikat tidak ditemukan.")
	}
	return writeOK(c, s)
}

func createSertifikat(c *echo.Context, pool *pgxpool.Pool) error {
	userID, _, ok := requirePerm(c, "ukk.update")
	if !ok {
		return nil
	}
	ukkID := c.Param("id")
	siswaID := c.Param("siswaId")
	if !isValidUUID(ukkID) || !isValidUUID(siswaID) {
		return writeError(c, http.StatusBadRequest, "validation_error", "ID tidak valid.")
	}
	ctx := c.Request().Context()
	ukk, err := getUkkRow(ctx, pool, ukkID)
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "UKK tidak ditemukan.")
	}
	if !existsID(ctx, pool, "siswa", siswaID) {
		return writeError(c, http.StatusBadRequest, "validation_error", "Siswa tidak ditemukan.")
	}
	// Idempotensi: sertifikat yang sudah ada dikembalikan apa adanya.
	if existing, err := getSertifikatRow(ctx, pool, ukkID, siswaID); err == nil {
		return writeOK(c, existing)
	}
	// Sertifikat HANYA bagi yang lulus -> ditolak 409 bila tidak lulus.
	var lulus *bool
	err = pool.QueryRow(ctx,
		`SELECT lulus FROM ukk_nilai WHERE ukk_id = $1 AND siswa_id = $2`,
		ukkID, siswaID).Scan(&lulus)
	if err != nil || lulus == nil || !*lulus {
		return writeError(c, http.StatusConflict, "conflict",
			"Sertifikat hanya dapat diterbitkan bagi siswa yang lulus UKK (nilai akhir >= 70).")
	}
	// Nomor: UKK/<tahun>/<kode-jurusan>/<nomor-urut 4 digit>.
	// Tahun dari tanggal UKK bila ada, bila tidak tahun berjalan.
	tahun := time.Now().Year()
	if ukk.Tanggal != nil {
		tahun = ukk.Tanggal.Year()
	}
	jurusanKode := ""
	if ukk.JurusanKode != nil {
		jurusanKode = *ukk.JurusanKode
	}
	// Nomor urut per jurusan + tahun. Coba hingga 5x bila bentrok UNIQUE
	// (perlombaan dua penerbitan bersamaan).
	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		var count int64
		if err := pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM ukk_sertifikat s JOIN ukk_ujian u ON u.id = s.ukk_id
			 WHERE u.jurusan_id = $1 AND EXTRACT(YEAR FROM s.diterbitkan_at) = $2`,
			ukk.JurusanID, tahun).Scan(&count); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menerbitkan sertifikat.")
		}
		nomor := fmt.Sprintf("UKK/%d/%s/%04d", tahun, jurusanKode, count+1+int64(attempt))
		_, lastErr = pool.Exec(ctx,
			`INSERT INTO ukk_sertifikat (ukk_id, siswa_id, nomor_sertifikat) VALUES ($1, $2, $3)`,
			ukkID, siswaID, nomor)
		if lastErr == nil {
			_ = audit.Catat(ctx, pool, userID, "ukk.sertifikat.terbit", "ukk_sertifikat", siswaID,
				map[string]any{"ukk_id": ukkID, "siswa_id": siswaID, "nomor_sertifikat": nomor}, clientIP(c))
			s, err := getSertifikatRow(ctx, pool, ukkID, siswaID)
			if err != nil {
				return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca sertifikat.")
			}
			return writeOKStatus(c, http.StatusCreated, s)
		}
		// Bukan bentrok nomor unik -> jangan ulangi.
		if !strings.Contains(lastErr.Error(), "duplicate key") && !strings.Contains(lastErr.Error(), "unique") {
			break
		}
	}
	_ = lastErr
	return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menerbitkan sertifikat.")
}
