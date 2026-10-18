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
//
// Struktur layer (hasil refactor arsitektur):
//   - handler.go:           Echo handler + Register (file ini) + helper query
//   - handler_master.go:    handler Jurusan/Kelas/Mapel
//   - handler_guru.go:      handler Guru
//   - handler_siswa.go:     handler Siswa
//   - handler_import.go:    handler Import Siswa (info + upload CSV)
//   - handler_referensi.go: handler Kompetensi/Kisi-kisi
//   - service.go:           Service + RequestActor + catat + hashPassword
//   - service_master.go:    logika bisnis Jurusan/Kelas/Mapel
//   - service_guru.go:      logika bisnis Guru
//   - service_siswa.go:     logika bisnis Siswa (+ akun login siswa)
//   - service_import.go:    logika bisnis Import Siswa CSV
//   - service_referensi.go: logika bisnis Kompetensi/Kisi-kisi
//   - repository.go:        Repository + helper error SQL
//   - repository_master.go: query SQL Jurusan/Kelas/Mapel
//   - repository_guru.go:   query SQL Guru
//   - repository_siswa.go:  query SQL Siswa
//   - repository_import.go: query SQL Import Siswa
//   - repository_referensi.go: query SQL Kompetensi/Kisi-kisi
//   - dto.go:               DTO JSON + tipe import
//
// Perilaku (route, status, bentuk JSON, pesan error) IDENTIK dengan
// sebelum refactor.
package master

import (
	"strconv"
	"strings"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// Handler adalah lapisan HTTP paket master.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// actorFrom membangun RequestActor dari echo.Context.
func actorFrom(c *echo.Context) RequestActor {
	userID, _, _, _, _ := audit.CurrentUser(c)
	return RequestActor{UserID: userID, RealIP: c.RealIP()}
}

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

	repo := NewRepository(pool)
	svc := NewService(repo, pool)
	h := NewHandler(svc)

	e.GET("/api/v1/master/jurusan", h.ListJurusan)
	e.POST("/api/v1/master/jurusan", h.CreateJurusan)
	e.GET("/api/v1/master/jurusan/:id", h.GetJurusan)
	e.PUT("/api/v1/master/jurusan/:id", h.UpdateJurusan)
	e.DELETE("/api/v1/master/jurusan/:id", h.DeleteJurusan)

	e.GET("/api/v1/master/kelas", h.ListKelas)
	e.POST("/api/v1/master/kelas", h.CreateKelas)
	e.GET("/api/v1/master/kelas/:id", h.GetKelas)
	e.PUT("/api/v1/master/kelas/:id", h.UpdateKelas)
	e.DELETE("/api/v1/master/kelas/:id", h.DeleteKelas)

	e.GET("/api/v1/master/mapel", h.ListMapel)
	e.POST("/api/v1/master/mapel", h.CreateMapel)
	e.GET("/api/v1/master/mapel/:id", h.GetMapel)
	e.PUT("/api/v1/master/mapel/:id", h.UpdateMapel)
	e.DELETE("/api/v1/master/mapel/:id", h.DeleteMapel)

	e.GET("/api/v1/master/guru", h.ListGuru)
	e.POST("/api/v1/master/guru", h.CreateGuru)
	e.GET("/api/v1/master/guru/:id", h.GetGuru)
	e.PUT("/api/v1/master/guru/:id", h.UpdateGuru)
	e.DELETE("/api/v1/master/guru/:id", h.DeleteGuru)

	e.GET("/api/v1/master/siswa", h.ListSiswa)
	e.POST("/api/v1/master/siswa", h.CreateSiswa)
	// Route import (GET info + POST upload) SEBELUM route :id agar
	// segmen statis "import" tidak tertangkap param :id.
	e.GET("/api/v1/master/siswa/import", h.InfoImportSiswa)
	e.POST("/api/v1/master/siswa/import", h.ImportSiswa)
	e.GET("/api/v1/master/siswa/:id", h.GetSiswa)
	e.PUT("/api/v1/master/siswa/:id", h.UpdateSiswa)
	e.DELETE("/api/v1/master/siswa/:id", h.DeleteSiswa)

	e.GET("/api/v1/master/kompetensi", h.ListKompetensi)
	e.POST("/api/v1/master/kompetensi", h.CreateKompetensi)
	e.GET("/api/v1/master/kompetensi/:id", h.GetKompetensi)
	e.PUT("/api/v1/master/kompetensi/:id", h.UpdateKompetensi)
	e.DELETE("/api/v1/master/kompetensi/:id", h.DeleteKompetensi)

	e.GET("/api/v1/master/kisi-kisi", h.ListKisiKisi)
	e.POST("/api/v1/master/kisi-kisi", h.CreateKisiKisi)
	e.GET("/api/v1/master/kisi-kisi/:id", h.GetKisiKisi)
	e.PUT("/api/v1/master/kisi-kisi/:id", h.UpdateKisiKisi)
	e.DELETE("/api/v1/master/kisi-kisi/:id", h.DeleteKisiKisi)
}

// ---------------------------------------------------------------------
// Helper query param (parsing HTTP — tetap di layer handler)
// ---------------------------------------------------------------------

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
