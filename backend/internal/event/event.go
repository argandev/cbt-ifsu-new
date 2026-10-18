// Package event — Event + Ruangan + Kartu Peserta (Tim Fase 2).
//
// Event = pengelompokan ujian (mis. "STS Ganjil 2026/2027", "SAS Genap",
// "USBK 2027", "UKK 2027"). Jadwal ujian dapat merujuk ke satu event
// (opsional). Ruangan = ruang fisik ujian (nama, kapasitas, lokasi,
// keterangan). Jadwal dapat ditugaskan ke satu ruangan.
//
// Kartu peserta = SATU kartu per siswa untuk SEMUA ujian (model 1-KARTU):
// identitas siswa sekali + tabel SEMUA jadwal kelasnya. Nomor peserta unik PER
// JADWAL (siswa yang sama di jadwal berbeda boleh bernomor berbeda),
// format: <kode-event-atau-US>/<tahun tanggal jadwal>/<nomor urut 3 digit>
// (contoh: "STS-GANJIL-26/2026/001", tanpa event: "US/2026/001").
// Nomor urut dihitung dari urutan nama siswa A-Z. Kartu TIDAK berisi
// token ujian — token hanya dari proktor saat ujian.
//
// Izin (sesuai seed + kontrak §5):
//   - Event/ruangan CRUD: event.read (GET), event.create/update/delete
//     (mutasi). Guru TIDAK punya event.* di matriks kanonis, jadi guru
//     mendapat 403 pada CRUD event/ruangan — itu sesuai seed.
//   - Kartu (daftar + satu): staf admin/proktor/guru. Karena guru tidak
//     punya event.read di seed, kartu menerima event.read ATAU
//     rekap.read untuk staf — pengawas/penguji_ukk yang juga punya
//     rekap.read tetap ditolak karena gate role staf di atas.
//   - Siswa hanya boleh melihat kartunya sendiri (satu kartu); daftar
//     kartu bukan haknya (anti-IDOR: 404, bukan 403).
//
// Handler memakai pool langsung dengan SQL yang sama seperti di
// backend/db/queries/event.sql (pola Tim 5), sampai sqlc generate.
package event

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

func writeError(c *echo.Context, status int, code, message string) error {
	return c.JSON(status, map[string]any{
		"error": map[string]any{"code": code, "message": message},
	})
}

func writeOK(c *echo.Context, data any) error {
	return c.JSON(http.StatusOK, map[string]any{"data": data})
}

func writeCreated(c *echo.Context, data any) error {
	return c.JSON(http.StatusCreated, map[string]any{"data": data})
}

// requirePerm: 401 bila belum login (context auth Tim 1 kosong),
// 403 bila tidak punya permission. Role admin lolos semua permission
// (lihat audit.HasPermission).
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

// requireKartuStaf: gate kartu untuk staf admin/proktor/guru (lihat
// komentar paket). Role lain -> 404 (anti-IDOR: resource bukan haknya).
func requireKartuStaf(c *echo.Context) (userID, role string, ok bool) {
	uid, r, _, _, ok := audit.CurrentUser(c)
	if !ok {
		_ = writeError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
		return "", "", false
	}
	if r != "admin" && r != "proktor" && r != "guru" {
		_ = writeError(c, http.StatusNotFound, "not_found", "Kartu peserta tidak ditemukan.")
		return "", "", false
	}
	if !audit.HasPermission(c, "event.read") && !audit.HasPermission(c, "rekap.read") {
		_ = writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat kartu peserta.")
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

func parsePagination(c *echo.Context) (limit, offset int) {
	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || limit <= 0 {
		limit = 20
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

func strPtr(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func isUniqueViolation(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "unique"))
}

// --- Model JSON ---

type eventRow struct {
	ID             string  `json:"id"`
	Kode           string  `json:"kode"`
	Nama           string  `json:"nama"`
	Jenis          string  `json:"jenis"`
	TanggalMulai   *string `json:"tanggal_mulai"`
	TanggalSelesai *string `json:"tanggal_selesai"`
	Deskripsi      *string `json:"deskripsi"`
	CreatedAt      any     `json:"created_at"`
}

type ruanganRow struct {
	ID         string  `json:"id"`
	Kode       string  `json:"kode"`
	Nama       string  `json:"nama"`
	Kapasitas  int     `json:"kapasitas"`
	Lokasi     *string `json:"lokasi"`
	Keterangan *string `json:"keterangan"`
	CreatedAt  any     `json:"created_at"`
}

const catatanPasswordAwal = "Password awal: NISN (wajib diganti saat login pertama)"

type kartuRow struct {
	JadwalID      string  `json:"jadwal_id"`
	SiswaID       string  `json:"siswa_id"`
	NomorPeserta  string  `json:"nomor_peserta"`
	NamaSekolah   string  `json:"nama_sekolah"`
	NamaEvent     *string `json:"nama_event"`
	NamaSiswa     string  `json:"nama_siswa"`
	NISN          string  `json:"nisn"`
	Kelas         string  `json:"kelas"`
	Jurusan       string  `json:"jurusan"`
	Mapel         string  `json:"mapel"`
	Paket         string  `json:"paket"`
	Tanggal       string  `json:"tanggal"`
	JamMulai      string  `json:"jam_mulai"`
	JamSelesai    string  `json:"jam_selesai"`
	Ruangan       *string `json:"ruangan"`
	RuanganLokasi *string `json:"ruangan_lokasi"`
	Username      string  `json:"username"`
	Catatan       string  `json:"catatan"`
}

// jadwalKartu adalah data jadwal + konteksnya untuk kartu peserta.
type jadwalKartu struct {
	ID            string
	KelasID       string
	Tanggal       string // YYYY-MM-DD
	JamMulai      string // HH:MM
	JamSelesai    string // HH:MM
	Status        string
	PaketNama     string
	PaketJenis    string
	MapelNama     string
	KelasNama     string
	JurusanNama   string
	EventID       *string
	EventKode     *string
	EventNama     *string
	RuanganID     *string
	RuanganNama   *string
	RuanganLokasi *string
}

type siswaKartu struct {
	ID      string
	NISN    string
	Nama    string
	KelasID string
	UserID  *string
}

// --- Load helpers ---

func scanEvent(row pgx.Row) (*eventRow, error) {
	var e eventRow
	err := row.Scan(&e.ID, &e.Kode, &e.Nama, &e.Jenis, &e.TanggalMulai,
		&e.TanggalSelesai, &e.Deskripsi, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func loadEvent(ctx context.Context, pool *pgxpool.Pool, id string) (*eventRow, bool, error) {
	if !isValidUUID(id) {
		return nil, false, nil
	}
	e, err := scanEvent(pool.QueryRow(ctx,
		`SELECT id, kode, nama, jenis, tanggal_mulai::text, tanggal_selesai::text,
		        deskripsi, created_at FROM event WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return e, true, nil
}

func scanRuangan(row pgx.Row) (*ruanganRow, error) {
	var r ruanganRow
	err := row.Scan(&r.ID, &r.Kode, &r.Nama, &r.Kapasitas, &r.Lokasi, &r.Keterangan, &r.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func loadRuangan(ctx context.Context, pool *pgxpool.Pool, id string) (*ruanganRow, bool, error) {
	if !isValidUUID(id) {
		return nil, false, nil
	}
	r, err := scanRuangan(pool.QueryRow(ctx,
		`SELECT id, kode, nama, kapasitas, lokasi, keterangan, created_at
		 FROM ruangan WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return r, true, nil
}

func loadJadwalKartu(ctx context.Context, pool *pgxpool.Pool, jadwalID string) (*jadwalKartu, bool, error) {
	if !isValidUUID(jadwalID) {
		return nil, false, nil
	}
	var j jadwalKartu
	err := pool.QueryRow(ctx,
		`SELECT j.id, j.kelas_id, to_char(j.tanggal, 'YYYY-MM-DD'),
		        to_char(j.jam_mulai, 'HH24:MI'), to_char(j.jam_selesai, 'HH24:MI'), j.status,
		        p.nama, p.jenis, m.nama, k.nama, ju.nama,
		        j.event_id, e.kode, e.nama, j.ruangan_id, r.nama, r.lokasi
		 FROM jadwal_ujian j
		 JOIN paket_ujian p ON p.id = j.paket_id
		 JOIN mapel m ON m.id = p.mapel_id
		 JOIN kelas k ON k.id = j.kelas_id
		 JOIN jurusan ju ON ju.id = k.jurusan_id
		 LEFT JOIN event e ON e.id = j.event_id
		 LEFT JOIN ruangan r ON r.id = j.ruangan_id
		 WHERE j.id = $1`, jadwalID).
		Scan(&j.ID, &j.KelasID, &j.Tanggal, &j.JamMulai, &j.JamSelesai, &j.Status,
			&j.PaketNama, &j.PaketJenis, &j.MapelNama, &j.KelasNama, &j.JurusanNama,
			&j.EventID, &j.EventKode, &j.EventNama, &j.RuanganID, &j.RuanganNama, &j.RuanganLokasi)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &j, true, nil
}

func loadNamaSekolah(ctx context.Context, pool *pgxpool.Pool) string {
	var raw json.RawMessage
	if err := pool.QueryRow(ctx,
		`SELECT value FROM settings WHERE key = 'nama_sekolah'`).Scan(&raw); err != nil {
		return ""
	}
	var nama string
	if err := json.Unmarshal(raw, &nama); err != nil {
		return ""
	}
	return nama
}

// buildKartuList menyusun kartu semua siswa aktif di kelas jadwal,
// urut nama A-Z. Nomor peserta unik per jadwal.
func buildKartuList(ctx context.Context, pool *pgxpool.Pool, j *jadwalKartu) ([]kartuRow, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, nisn, nama, kelas_id, user_id FROM siswa
		 WHERE kelas_id = $1 AND is_active = TRUE ORDER BY nama ASC`, j.KelasID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	namaSekolah := loadNamaSekolah(ctx, pool)
	kode := "US"
	if j.EventKode != nil && strings.TrimSpace(*j.EventKode) != "" {
		kode = *j.EventKode
	}
	tahun := ""
	if len(j.Tanggal) >= 4 {
		tahun = j.Tanggal[:4]
	}
	items := []kartuRow{}
	for rows.Next() {
		var s siswaKartu
		if err := rows.Scan(&s.ID, &s.NISN, &s.Nama, &s.KelasID, &s.UserID); err != nil {
			return nil, err
		}
		nomor := fmt.Sprintf("%s/%s/%03d", kode, tahun, len(items)+1)
		items = append(items, kartuRow{
			JadwalID:      j.ID,
			SiswaID:       s.ID,
			NomorPeserta:  nomor,
			NamaSekolah:   namaSekolah,
			NamaEvent:     j.EventNama,
			NamaSiswa:     s.Nama,
			NISN:          s.NISN,
			Kelas:         j.KelasNama,
			Jurusan:       j.JurusanNama,
			Mapel:         j.MapelNama,
			Paket:         j.PaketNama,
			Tanggal:       j.Tanggal,
			JamMulai:      j.JamMulai,
			JamSelesai:    j.JamSelesai,
			Ruangan:       j.RuanganNama,
			RuanganLokasi: j.RuanganLokasi,
			Username:      s.NISN,
			Catatan:       catatanPasswordAwal,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func loadSiswaKartu(ctx context.Context, pool *pgxpool.Pool, siswaID string) (*siswaKartu, bool, error) {
	if !isValidUUID(siswaID) {
		return nil, false, nil
	}
	// Cari siswa.id dulu; bila tidak ketemu, fallback ke siswa.user_id.
	// Frontend KartuSayaPage mengirim siswaId = user.id (ID akun login),
	// bukan siswa.id — tanpa fallback, kartu sendiri selalu 404.
	// Pola sama seperti kelasSiswa di exam.go.
	var s siswaKartu
	err := pool.QueryRow(ctx,
		`SELECT id, nisn, nama, kelas_id, user_id FROM siswa WHERE id = $1`, siswaID).
		Scan(&s.ID, &s.NISN, &s.Nama, &s.KelasID, &s.UserID)
	if err == nil {
		return &s, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, err
	}
	err = pool.QueryRow(ctx,
		`SELECT id, nisn, nama, kelas_id, user_id FROM siswa WHERE user_id = $1`, siswaID).
		Scan(&s.ID, &s.NISN, &s.Nama, &s.KelasID, &s.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &s, true, nil
}

// --- Input + validasi ---

var jenisEventValid = map[string]bool{
	"STS": true, "SAS": true, "USBK": true, "UKK": true, "Tryout": true, "Lainnya": true,
}

type eventInput struct {
	Kode           string  `json:"kode"`
	Nama           string  `json:"nama"`
	Jenis          string  `json:"jenis"`
	TanggalMulai   *string `json:"tanggal_mulai"`
	TanggalSelesai *string `json:"tanggal_selesai"`
	Deskripsi      *string `json:"deskripsi"`
}

// parseTanggalValidasi: "" / null -> nil; selain itu harus YYYY-MM-DD.
func parseTanggalValidasi(v *string, field string) (any, error) {
	if v == nil || strings.TrimSpace(*v) == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", strings.TrimSpace(*v))
	if err != nil {
		return nil, fmt.Errorf("%s harus format tanggal YYYY-MM-DD", field)
	}
	return t.Format("2006-01-02"), nil
}

func (in *eventInput) validate() (mulai, selesai any, err error) {
	in.Kode = strings.TrimSpace(in.Kode)
	in.Nama = strings.TrimSpace(in.Nama)
	in.Jenis = strings.TrimSpace(in.Jenis)
	if in.Kode == "" || in.Nama == "" {
		return nil, nil, fmt.Errorf("kode dan nama event wajib diisi")
	}
	if in.Jenis == "" {
		in.Jenis = "Lainnya"
	}
	if !jenisEventValid[in.Jenis] {
		return nil, nil, fmt.Errorf("jenis harus salah satu: STS, SAS, USBK, UKK, Tryout, Lainnya")
	}
	mulai, err = parseTanggalValidasi(in.TanggalMulai, "tanggal_mulai")
	if err != nil {
		return nil, nil, err
	}
	selesai, err = parseTanggalValidasi(in.TanggalSelesai, "tanggal_selesai")
	if err != nil {
		return nil, nil, err
	}
	if mulai != nil && selesai != nil && selesai.(string) < mulai.(string) {
		return nil, nil, fmt.Errorf("tanggal_selesai tidak boleh sebelum tanggal_mulai")
	}
	return mulai, selesai, nil
}

type ruanganInput struct {
	Kode       string  `json:"kode"`
	Nama       string  `json:"nama"`
	Kapasitas  int     `json:"kapasitas"`
	Lokasi     *string `json:"lokasi"`
	Keterangan *string `json:"keterangan"`
}

func (in *ruanganInput) validate() error {
	in.Kode = strings.TrimSpace(in.Kode)
	in.Nama = strings.TrimSpace(in.Nama)
	if in.Kode == "" || in.Nama == "" {
		return fmt.Errorf("kode dan nama ruangan wajib diisi")
	}
	if in.Kapasitas <= 0 {
		return fmt.Errorf("kapasitas harus > 0")
	}
	return nil
}

func optionalText(v *string) any {
	if v == nil || strings.TrimSpace(*v) == "" {
		return nil
	}
	return strings.TrimSpace(*v)
}

// kodeDipakai: true bila kode sudah dipakai baris lain (excludeID boleh "").
func kodeDipakai(ctx context.Context, pool *pgxpool.Pool, tabel, kode, excludeID string) (bool, error) {
	var n int
	q := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE kode = $1`, tabel)
	args := []any{kode}
	if excludeID != "" {
		q += ` AND id <> $2`
		args = append(args, excludeID)
	}
	if err := pool.QueryRow(ctx, q, args...).Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}

func jadwalMemakai(ctx context.Context, pool *pgxpool.Pool, kolom, id string) (int, error) {
	var n int
	q := fmt.Sprintf(`SELECT COUNT(*) FROM jadwal_ujian WHERE %s = $1`, kolom)
	if err := pool.QueryRow(ctx, q, id).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// Register mendaftarkan endpoint event + ruangan + kartu peserta.
//
//	GET    /api/v1/event                        (event.read)
//	POST   /api/v1/event                        (event.create)
//	GET    /api/v1/event/:id                    (event.read)
//	PUT    /api/v1/event/:id                    (event.update)
//	DELETE /api/v1/event/:id                    (event.delete)
//	GET    /api/v1/ruangan                      (event.read)
//	POST   /api/v1/ruangan                      (event.create)
//	GET    /api/v1/ruangan/:id                  (event.read)
//	PUT    /api/v1/ruangan/:id                  (event.update)
//	DELETE /api/v1/ruangan/:id                  (event.delete)
//	PUT    /api/v1/jadwal/:id/event-ruangan     (event.update)
//	GET    /api/v1/kartu/pdf                  (staf admin/proktor/guru, PDF gabungan)
//	GET    /api/v1/kartu/siswa/:siswaId/pdf  (staf / siswa pemilik kartu, PDF 1 kartu)
//	GET    /api/v1/kartu/:jadwalId              (staf admin/proktor/guru, JSON legacy)
//	GET    /api/v1/kartu/:jadwalId/:siswaId    (staf / siswa pemilik kartu, JSON legacy)
//
// Aturan kunci:
//   - Hapus event/ruangan -> 409 bila masih dipakai jadwal ujian.
//   - Kode event/ruangan unik -> 409 bila duplikat.
//   - Tugaskan ruangan ke jadwal -> 409 bila kapasitas ruangan < jumlah
//     siswa aktif kelas jadwal.
//   - Kartu TIDAK berisi token ujian.
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q
	_ = cfg

	// ---- Event ----

	e.GET("/api/v1/event", func(c *echo.Context) error {
		if _, _, ok := requirePerm(c, "event.read"); !ok {
			return nil
		}
		limit, offset := parsePagination(c)
		jenis := strPtr(c.QueryParam("jenis"))
		search := strPtr(c.QueryParam("q"))
		rows, err := pool.Query(c.Request().Context(),
			`SELECT id, kode, nama, jenis, tanggal_mulai::text, tanggal_selesai::text,
			        deskripsi, created_at FROM event
			 WHERE ($1::text IS NULL OR jenis = $1)
			   AND ($2::text IS NULL OR nama ILIKE '%' || $2 || '%' OR kode ILIKE '%' || $2 || '%')
			 ORDER BY tanggal_mulai DESC NULLS LAST, nama ASC
			 LIMIT $3 OFFSET $4`,
			jenis, search, limit, offset)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca event.")
		}
		defer rows.Close()
		items := []eventRow{}
		for rows.Next() {
			var ev eventRow
			if err := rows.Scan(&ev.ID, &ev.Kode, &ev.Nama, &ev.Jenis, &ev.TanggalMulai,
				&ev.TanggalSelesai, &ev.Deskripsi, &ev.CreatedAt); err != nil {
				return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca event.")
			}
			items = append(items, ev)
		}
		var total int
		_ = pool.QueryRow(c.Request().Context(),
			`SELECT COUNT(*) FROM event
			 WHERE ($1::text IS NULL OR jenis = $1)
			   AND ($2::text IS NULL OR nama ILIKE '%' || $2 || '%' OR kode ILIKE '%' || $2 || '%')`,
			jenis, search).Scan(&total)
		return c.JSON(http.StatusOK, map[string]any{
			"data": map[string]any{"items": items, "total": total, "limit": limit, "offset": offset},
		})
	})

	e.POST("/api/v1/event", func(c *echo.Context) error {
		userID, _, ok := requirePerm(c, "event.create")
		if !ok {
			return nil
		}
		var in eventInput
		if err := c.Bind(&in); err != nil {
			return writeError(c, http.StatusBadRequest, "validation_error", "Body JSON tidak valid.")
		}
		mulai, selesai, err := in.validate()
		if err != nil {
			return writeError(c, http.StatusBadRequest, "validation_error", err.Error())
		}
		ctx := c.Request().Context()
		dipakai, err := kodeDipakai(ctx, pool, "event", in.Kode, "")
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa kode event.")
		}
		if dipakai {
			return writeError(c, http.StatusConflict, "duplicate_kode", "Kode event sudah dipakai.")
		}
		ev, err := scanEvent(pool.QueryRow(ctx,
			`INSERT INTO event (kode, nama, jenis, tanggal_mulai, tanggal_selesai, deskripsi)
			 VALUES ($1, $2, $3, $4::date, $5::date, $6)
			 RETURNING id, kode, nama, jenis, tanggal_mulai::text, tanggal_selesai::text,
			           deskripsi, created_at`,
			in.Kode, in.Nama, in.Jenis, mulai, selesai, optionalText(in.Deskripsi)))
		if err != nil {
			if isUniqueViolation(err) {
				return writeError(c, http.StatusConflict, "duplicate_kode", "Kode event sudah dipakai.")
			}
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membuat event.")
		}
		_ = audit.Catat(ctx, pool, userID, "event.create", "event", ev.ID,
			map[string]any{"kode": ev.Kode, "nama": ev.Nama}, c.RealIP())
		return writeCreated(c, ev)
	})

	e.GET("/api/v1/event/:id", func(c *echo.Context) error {
		if _, _, ok := requirePerm(c, "event.read"); !ok {
			return nil
		}
		ev, found, err := loadEvent(c.Request().Context(), pool, c.Param("id"))
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca event.")
		}
		if !found {
			return writeError(c, http.StatusNotFound, "not_found", "Event tidak ditemukan.")
		}
		return writeOK(c, ev)
	})

	e.PUT("/api/v1/event/:id", func(c *echo.Context) error {
		userID, _, ok := requirePerm(c, "event.update")
		if !ok {
			return nil
		}
		ctx := c.Request().Context()
		ev, found, err := loadEvent(ctx, pool, c.Param("id"))
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca event.")
		}
		if !found {
			return writeError(c, http.StatusNotFound, "not_found", "Event tidak ditemukan.")
		}
		var in eventInput
		if err := c.Bind(&in); err != nil {
			return writeError(c, http.StatusBadRequest, "validation_error", "Body JSON tidak valid.")
		}
		mulai, selesai, err := in.validate()
		if err != nil {
			return writeError(c, http.StatusBadRequest, "validation_error", err.Error())
		}
		dipakai, err := kodeDipakai(ctx, pool, "event", in.Kode, ev.ID)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa kode event.")
		}
		if dipakai {
			return writeError(c, http.StatusConflict, "duplicate_kode", "Kode event sudah dipakai event lain.")
		}
		_, err = pool.Exec(ctx,
			`UPDATE event SET kode = $2, nama = $3, jenis = $4,
			 tanggal_mulai = $5::date, tanggal_selesai = $6::date, deskripsi = $7
			 WHERE id = $1`,
			ev.ID, in.Kode, in.Nama, in.Jenis, mulai, selesai, optionalText(in.Deskripsi))
		if err != nil {
			if isUniqueViolation(err) {
				return writeError(c, http.StatusConflict, "duplicate_kode", "Kode event sudah dipakai event lain.")
			}
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mengubah event.")
		}
		_ = audit.Catat(ctx, pool, userID, "event.update", "event", ev.ID,
			map[string]any{"kode": in.Kode, "nama": in.Nama}, c.RealIP())
		updated, _, _ := loadEvent(ctx, pool, ev.ID)
		return writeOK(c, updated)
	})

	e.DELETE("/api/v1/event/:id", func(c *echo.Context) error {
		userID, _, ok := requirePerm(c, "event.delete")
		if !ok {
			return nil
		}
		ctx := c.Request().Context()
		ev, found, err := loadEvent(ctx, pool, c.Param("id"))
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca event.")
		}
		if !found {
			return writeError(c, http.StatusNotFound, "not_found", "Event tidak ditemukan.")
		}
		n, err := jadwalMemakai(ctx, pool, "event_id", ev.ID)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa pemakaian event.")
		}
		if n > 0 {
			return writeError(c, http.StatusConflict, "event_in_use",
				fmt.Sprintf("Event masih dipakai %d jadwal ujian — lepas event dari jadwal dulu.", n))
		}
		if _, err := pool.Exec(ctx, `DELETE FROM event WHERE id = $1`, ev.ID); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghapus event.")
		}
		_ = audit.Catat(ctx, pool, userID, "event.delete", "event", ev.ID,
			map[string]any{"kode": ev.Kode, "nama": ev.Nama}, c.RealIP())
		return writeOK(c, map[string]any{"deleted": true, "id": ev.ID})
	})

	// ---- Ruangan ----

	e.GET("/api/v1/ruangan", func(c *echo.Context) error {
		if _, _, ok := requirePerm(c, "event.read"); !ok {
			return nil
		}
		limit, offset := parsePagination(c)
		search := strPtr(c.QueryParam("q"))
		rows, err := pool.Query(c.Request().Context(),
			`SELECT id, kode, nama, kapasitas, lokasi, keterangan, created_at FROM ruangan
			 WHERE ($1::text IS NULL OR nama ILIKE '%' || $1 || '%' OR kode ILIKE '%' || $1 || '%'
			        OR COALESCE(lokasi, '') ILIKE '%' || $1 || '%')
			 ORDER BY nama ASC LIMIT $2 OFFSET $3`,
			search, limit, offset)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca ruangan.")
		}
		defer rows.Close()
		items := []ruanganRow{}
		for rows.Next() {
			var r ruanganRow
			if err := rows.Scan(&r.ID, &r.Kode, &r.Nama, &r.Kapasitas, &r.Lokasi, &r.Keterangan, &r.CreatedAt); err != nil {
				return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca ruangan.")
			}
			items = append(items, r)
		}
		var total int
		_ = pool.QueryRow(c.Request().Context(),
			`SELECT COUNT(*) FROM ruangan
			 WHERE ($1::text IS NULL OR nama ILIKE '%' || $1 || '%' OR kode ILIKE '%' || $1 || '%'
			        OR COALESCE(lokasi, '') ILIKE '%' || $1 || '%')`,
			search).Scan(&total)
		return c.JSON(http.StatusOK, map[string]any{
			"data": map[string]any{"items": items, "total": total, "limit": limit, "offset": offset},
		})
	})

	e.POST("/api/v1/ruangan", func(c *echo.Context) error {
		userID, _, ok := requirePerm(c, "event.create")
		if !ok {
			return nil
		}
		var in ruanganInput
		if err := c.Bind(&in); err != nil {
			return writeError(c, http.StatusBadRequest, "validation_error", "Body JSON tidak valid.")
		}
		if err := in.validate(); err != nil {
			return writeError(c, http.StatusBadRequest, "validation_error", err.Error())
		}
		ctx := c.Request().Context()
		dipakai, err := kodeDipakai(ctx, pool, "ruangan", in.Kode, "")
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa kode ruangan.")
		}
		if dipakai {
			return writeError(c, http.StatusConflict, "duplicate_kode", "Kode ruangan sudah dipakai.")
		}
		r, err := scanRuangan(pool.QueryRow(ctx,
			`INSERT INTO ruangan (kode, nama, kapasitas, lokasi, keterangan)
			 VALUES ($1, $2, $3, $4, $5)
			 RETURNING id, kode, nama, kapasitas, lokasi, keterangan, created_at`,
			in.Kode, in.Nama, in.Kapasitas, optionalText(in.Lokasi), optionalText(in.Keterangan)))
		if err != nil {
			if isUniqueViolation(err) {
				return writeError(c, http.StatusConflict, "duplicate_kode", "Kode ruangan sudah dipakai.")
			}
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membuat ruangan.")
		}
		_ = audit.Catat(ctx, pool, userID, "event.create", "ruangan", r.ID,
			map[string]any{"kode": r.Kode, "nama": r.Nama}, c.RealIP())
		return writeCreated(c, r)
	})

	e.GET("/api/v1/ruangan/:id", func(c *echo.Context) error {
		if _, _, ok := requirePerm(c, "event.read"); !ok {
			return nil
		}
		r, found, err := loadRuangan(c.Request().Context(), pool, c.Param("id"))
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca ruangan.")
		}
		if !found {
			return writeError(c, http.StatusNotFound, "not_found", "Ruangan tidak ditemukan.")
		}
		return writeOK(c, r)
	})

	e.PUT("/api/v1/ruangan/:id", func(c *echo.Context) error {
		userID, _, ok := requirePerm(c, "event.update")
		if !ok {
			return nil
		}
		ctx := c.Request().Context()
		r, found, err := loadRuangan(ctx, pool, c.Param("id"))
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca ruangan.")
		}
		if !found {
			return writeError(c, http.StatusNotFound, "not_found", "Ruangan tidak ditemukan.")
		}
		var in ruanganInput
		if err := c.Bind(&in); err != nil {
			return writeError(c, http.StatusBadRequest, "validation_error", "Body JSON tidak valid.")
		}
		if err := in.validate(); err != nil {
			return writeError(c, http.StatusBadRequest, "validation_error", err.Error())
		}
		dipakai, err := kodeDipakai(ctx, pool, "ruangan", in.Kode, r.ID)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa kode ruangan.")
		}
		if dipakai {
			return writeError(c, http.StatusConflict, "duplicate_kode", "Kode ruangan sudah dipakai ruangan lain.")
		}
		// Kapasitas baru tidak boleh di bawah jumlah siswa kelas jadwal
		// yang sudah memakai ruangan ini. Pesan 409 menyebut jadwalnya.
		var jadwalPenuh string
		_ = pool.QueryRow(ctx,
			`SELECT COALESCE(MAX(k.nama || ' (' || j.tanggal::text || ')'), '') FROM jadwal_ujian j
			 JOIN kelas k ON k.id = j.kelas_id
			 WHERE j.ruangan_id = $1
			   AND (SELECT COUNT(*) FROM siswa s WHERE s.kelas_id = j.kelas_id AND s.is_active = TRUE) > $2`,
			r.ID, in.Kapasitas).Scan(&jadwalPenuh)
		if jadwalPenuh != "" {
			return writeError(c, http.StatusConflict, "kapasitas_kurang",
				fmt.Sprintf("Kapasitas baru (%d kursi) kurang dari jumlah siswa kelas jadwal %s yang memakai ruangan ini. Turunkan kapasitas setelah jadwal dipindah.", in.Kapasitas, jadwalPenuh))
		}
		_, err = pool.Exec(ctx,
			`UPDATE ruangan SET kode = $2, nama = $3, kapasitas = $4, lokasi = $5, keterangan = $6
			 WHERE id = $1`,
			r.ID, in.Kode, in.Nama, in.Kapasitas, optionalText(in.Lokasi), optionalText(in.Keterangan))
		if err != nil {
			if isUniqueViolation(err) {
				return writeError(c, http.StatusConflict, "duplicate_kode", "Kode ruangan sudah dipakai ruangan lain.")
			}
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mengubah ruangan.")
		}
		_ = audit.Catat(ctx, pool, userID, "event.update", "ruangan", r.ID,
			map[string]any{"kode": in.Kode, "nama": in.Nama}, c.RealIP())
		updated, _, _ := loadRuangan(ctx, pool, r.ID)
		return writeOK(c, updated)
	})

	e.DELETE("/api/v1/ruangan/:id", func(c *echo.Context) error {
		userID, _, ok := requirePerm(c, "event.delete")
		if !ok {
			return nil
		}
		ctx := c.Request().Context()
		r, found, err := loadRuangan(ctx, pool, c.Param("id"))
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca ruangan.")
		}
		if !found {
			return writeError(c, http.StatusNotFound, "not_found", "Ruangan tidak ditemukan.")
		}
		n, err := jadwalMemakai(ctx, pool, "ruangan_id", r.ID)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa pemakaian ruangan.")
		}
		if n > 0 {
			return writeError(c, http.StatusConflict, "ruangan_in_use",
				fmt.Sprintf("Ruangan masih dipakai %d jadwal ujian — lepas ruangan dari jadwal dulu.", n))
		}
		if _, err := pool.Exec(ctx, `DELETE FROM ruangan WHERE id = $1`, r.ID); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghapus ruangan.")
		}
		_ = audit.Catat(ctx, pool, userID, "event.delete", "ruangan", r.ID,
			map[string]any{"kode": r.Kode, "nama": r.Nama}, c.RealIP())
		return writeOK(c, map[string]any{"deleted": true, "id": r.ID})
	})

	// ---- Tugaskan event + ruangan ke jadwal ----

	e.PUT("/api/v1/jadwal/:id/event-ruangan", func(c *echo.Context) error {
		userID, _, ok := requirePerm(c, "event.update")
		if !ok {
			return nil
		}
		ctx := c.Request().Context()
		jadwalID := c.Param("id")
		var kelasID string
		var kelasNama string
		if err := pool.QueryRow(ctx,
			`SELECT j.kelas_id, COALESCE(k.nama, '') FROM jadwal_ujian j
			 LEFT JOIN kelas k ON k.id = j.kelas_id WHERE j.id = $1`, jadwalID).
			Scan(&kelasID, &kelasNama); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
			}
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca jadwal.")
		}
		var body struct {
			EventID   *string `json:"event_id"`
			RuanganID *string `json:"ruangan_id"`
		}
		if err := c.Bind(&body); err != nil {
			return writeError(c, http.StatusBadRequest, "validation_error", "Body JSON tidak valid.")
		}
		// Field tidak dikirim / string kosong = lepas penugasan (NULL).
		var eventArg, ruanganArg any
		if body.EventID != nil && strings.TrimSpace(*body.EventID) != "" {
			evID := strings.TrimSpace(*body.EventID)
			if !isValidUUID(evID) {
				return writeError(c, http.StatusBadRequest, "validation_error", "event_id harus UUID event yang valid.")
			}
			if _, found, err := loadEvent(ctx, pool, evID); err != nil {
				return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa event.")
			} else if !found {
				return writeError(c, http.StatusBadRequest, "validation_error", "event_id tidak ditemukan.")
			}
			eventArg = evID
		}
		if body.RuanganID != nil && strings.TrimSpace(*body.RuanganID) != "" {
			rgID := strings.TrimSpace(*body.RuanganID)
			if !isValidUUID(rgID) {
				return writeError(c, http.StatusBadRequest, "validation_error", "ruangan_id harus UUID ruangan yang valid.")
			}
			rg, found, err := loadRuangan(ctx, pool, rgID)
			if err != nil {
				return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa ruangan.")
			}
			if !found {
				return writeError(c, http.StatusBadRequest, "validation_error", "ruangan_id tidak ditemukan.")
			}
			// Cek kapasitas ruangan >= jumlah siswa AKTIF kelas jadwal.
			var jumlahSiswa int
			if err := pool.QueryRow(ctx,
				`SELECT COUNT(*) FROM siswa WHERE kelas_id = $1 AND is_active = TRUE`,
				kelasID).Scan(&jumlahSiswa); err != nil {
				return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghitung siswa kelas.")
			}
			if rg.Kapasitas < jumlahSiswa {
				return writeError(c, http.StatusConflict, "kapasitas_kurang",
					fmt.Sprintf("Ruangan %s berkapasitas %d kursi, kurang dari jumlah siswa kelas %s (%d siswa). Pilih ruangan yang lebih besar.", rg.Nama, rg.Kapasitas, kelasNama, jumlahSiswa))
			}
			ruanganArg = rgID
		}
		if _, err := pool.Exec(ctx,
			`UPDATE jadwal_ujian SET event_id = $2, ruangan_id = $3 WHERE id = $1`,
			jadwalID, eventArg, ruanganArg); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menugaskan event/ruangan ke jadwal.")
		}
		_ = audit.Catat(ctx, pool, userID, "event.assign", "jadwal_ujian", jadwalID,
			map[string]any{"event_id": eventArg, "ruangan_id": ruanganArg}, c.RealIP())
		j, _, _ := loadJadwalKartu(ctx, pool, jadwalID)
		var evOut, rgOut any
		if j != nil {
			if j.EventNama != nil {
				evOut = map[string]any{"id": j.EventID, "kode": j.EventKode, "nama": j.EventNama}
			}
			if j.RuanganNama != nil {
				rgOut = map[string]any{"id": j.RuanganID, "nama": j.RuanganNama, "lokasi": j.RuanganLokasi}
			}
		}
		return writeOK(c, map[string]any{
			"jadwal_id": jadwalID, "event_id": eventArg, "ruangan_id": ruanganArg,
			"event": evOut, "ruangan": rgOut,
		})
	})

	// ---- Kartu peserta PDF (server-side, generator di kartu_pdf.go) ----
	// Route PDF didaftarkan SEBELUM route JSON legacy agar segmen statis
	// "pdf" / "siswa" menang atas parameter :jadwalId.

	e.GET("/api/v1/kartu/pdf", handleKartuBulkPDF(pool))
	e.GET("/api/v1/kartu/siswa/:siswaId/pdf", handleKartuSiswaPDF(pool))

	// ---- Kartu peserta (JSON, LEGACY — tetap dipertahankan) ----

	e.GET("/api/v1/kartu/:jadwalId", func(c *echo.Context) error {
		if _, _, ok := requireKartuStaf(c); !ok {
			return nil
		}
		ctx := c.Request().Context()
		j, found, err := loadJadwalKartu(ctx, pool, c.Param("jadwalId"))
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca jadwal.")
		}
		if !found {
			return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
		}
		items, err := buildKartuList(ctx, pool, j)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menyusun kartu peserta.")
		}
		var evOut, rgOut any
		if j.EventNama != nil {
			evOut = map[string]any{"id": j.EventID, "kode": j.EventKode, "nama": j.EventNama}
		}
		if j.RuanganNama != nil {
			rgOut = map[string]any{"id": j.RuanganID, "nama": j.RuanganNama, "lokasi": j.RuanganLokasi}
		}
		return writeOK(c, map[string]any{
			"jadwal": map[string]any{
				"id": j.ID, "tanggal": j.Tanggal, "jam_mulai": j.JamMulai, "jam_selesai": j.JamSelesai,
				"status": j.Status, "paket_nama": j.PaketNama, "paket_jenis": j.PaketJenis,
				"mapel_nama": j.MapelNama, "kelas_nama": j.KelasNama, "jurusan_nama": j.JurusanNama,
				"event": evOut, "ruangan": rgOut,
			},
			"items": items,
			"total": len(items),
		})
	})

	e.GET("/api/v1/kartu/:jadwalId/:siswaId", func(c *echo.Context) error {
		ctx := c.Request().Context()
		uid, role, _, _, ok := audit.CurrentUser(c)
		if !ok {
			return writeError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
		}
		if role == "siswa" {
			// Siswa hanya boleh melihat kartunya sendiri.
			s, found, err := loadSiswaKartu(ctx, pool, c.Param("siswaId"))
			if err != nil {
				return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca kartu peserta.")
			}
			if !found {
				return writeError(c, http.StatusNotFound, "not_found", "Kartu peserta tidak ditemukan.")
			}
			milikSendiri := (s.UserID != nil && *s.UserID == uid) || s.ID == uid
			if !milikSendiri {
				return writeError(c, http.StatusNotFound, "not_found", "Kartu peserta tidak ditemukan.")
			}
			j, found, err := loadJadwalKartu(ctx, pool, c.Param("jadwalId"))
			if err != nil {
				return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca jadwal.")
			}
			if !found || j.KelasID != s.KelasID {
				return writeError(c, http.StatusNotFound, "not_found", "Kartu peserta tidak ditemukan.")
			}
			items, err := buildKartuList(ctx, pool, j)
			if err != nil {
				return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menyusun kartu peserta.")
			}
			for _, k := range items {
				if k.SiswaID == s.ID {
					return writeOK(c, k)
				}
			}
			return writeError(c, http.StatusNotFound, "not_found", "Kartu peserta tidak ditemukan.")
		}
		if _, _, ok := requireKartuStaf(c); !ok {
			return nil
		}
		j, found, err := loadJadwalKartu(ctx, pool, c.Param("jadwalId"))
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca jadwal.")
		}
		if !found {
			return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
		}
		s, found, err := loadSiswaKartu(ctx, pool, c.Param("siswaId"))
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca kartu peserta.")
		}
		if !found || s.KelasID != j.KelasID {
			return writeError(c, http.StatusNotFound, "not_found", "Kartu peserta tidak ditemukan.")
		}
		items, err := buildKartuList(ctx, pool, j)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menyusun kartu peserta.")
		}
		for _, k := range items {
			if k.SiswaID == s.ID {
				return writeOK(c, k)
			}
		}
		return writeError(c, http.StatusNotFound, "not_found", "Kartu peserta tidak ditemukan.")
	})
}
