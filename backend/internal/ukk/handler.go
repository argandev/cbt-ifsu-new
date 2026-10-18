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
//	GET    /ukk/:id/peserta                       (ukk.read; D5: daftar siswa
//	                                              kelas UKK ini untuk dropdown
//	                                              halaman nilai — tanpa izin
//	                                              master, hanya field yang
//	                                              dibutuhkan halaman nilai)
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
//
// Struktur layer (hasil refactor arsitektur):
//   - ukk.go:              dokumentasi paket
//   - handler.go:          Echo handler + Register + helper auth (file ini)
//   - handler_nilai.go:    handler nilai + peserta
//   - handler_berkas.go:   handler berkas praktik
//   - handler_sertifikat.go: handler sertifikat
//   - service.go:          validasi + aturan hitung + logika bisnis + audit
//   - repository.go:       DTO + semua query SQL
//   - platform:            envelope JSON + paginasi (menggantikan helper lokal)
//
// Perilaku (route, status, bentuk JSON, pesan error) IDENTIK dengan
// sebelum refactor.
package ukk

import (
	"net/http"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/config"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// --- Handler ----------------------------------------------------------------

// Handler adalah lapisan HTTP paket ukk.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// requirePerm: 401 bila belum login (context auth kosong), 403 bila tidak
// punya permission. Role admin lolos semua permission (audit.HasPermission).
// (Dipindah dari ukk.go tanpa perubahan.)
func requirePerm(c *echo.Context, perm string) (userID, role string, ok bool) {
	uid, r, _, _, ok := audit.CurrentUser(c)
	if !ok {
		_ = platform.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
		return "", "", false
	}
	if !audit.HasPermission(c, perm) {
		_ = platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin: "+perm+".")
		return "", "", false
	}
	return uid, r, true
}

// requireLogin hanya memastikan ada sesi, tanpa permission tertentu.
// Dipakai jalur yang mengizinkan role siswa untuk dirinya sendiri.
// (Dipindah dari ukk.go tanpa perubahan.)
func requireLogin(c *echo.Context) (userID, role string, ok bool) {
	uid, r, _, _, ok := audit.CurrentUser(c)
	if !ok {
		_ = platform.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
		return "", "", false
	}
	return uid, r, true
}

func clientIP(c *echo.Context) string {
	if ip := c.RealIP(); ip != "" {
		return ip
	}
	return c.Request().RemoteAddr
}

// actorFrom membangun platform.RequestActor dari echo.Context.
func actorFrom(c *echo.Context) platform.RequestActor {
	userID, role, _, _, _ := audit.CurrentUser(c)
	return platform.RequestActor{UserID: userID, Role: role, RealIP: clientIP(c)}
}

// Register mendaftarkan endpoint UKK.
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q

	repo := NewRepository(pool)
	svc := NewService(repo, pool, cfg.StorageDir)
	h := NewHandler(svc)

	e.GET("/api/v1/ukk", h.ListUkk)
	e.POST("/api/v1/ukk", h.CreateUkk)
	e.GET("/api/v1/ukk/:id", h.GetUkk)
	e.PUT("/api/v1/ukk/:id", h.UpdateUkk)
	e.DELETE("/api/v1/ukk/:id", h.DeleteUkk)

	e.GET("/api/v1/ukk/:id/nilai", h.ListNilai)
	e.GET("/api/v1/ukk/:id/peserta", h.ListPeserta)
	e.POST("/api/v1/ukk/:id/nilai", func(c *echo.Context) error { return h.UpsertNilai(c, "") })
	e.PUT("/api/v1/ukk/:id/nilai/:nilaiId", func(c *echo.Context) error {
		return h.UpsertNilai(c, c.Param("nilaiId"))
	})

	e.GET("/api/v1/ukk/:id/berkas", h.ListBerkas)
	e.POST("/api/v1/ukk/:id/berkas", h.UploadBerkas)
	e.GET("/api/v1/ukk/:id/berkas/:berkasId/download", h.DownloadBerkas)

	e.POST("/api/v1/ukk/:id/sertifikat/:siswaId", h.CreateSertifikat)
	e.GET("/api/v1/ukk/:id/sertifikat/:siswaId", h.GetSertifikat)
}

// ListUkk menangani GET /api/v1/ukk.
func (h *Handler) ListUkk(c *echo.Context) error {
	if _, _, ok := requirePerm(c, "ukk.read"); !ok {
		return nil
	}
	// Gaya paginasi ukk: limit/offset, default 50, maks 100.
	limit, offset := platform.ParseLimit(c, 50)
	jurusanID := strPtr(c.QueryParam("jurusan_id"))
	kelasID := strPtr(c.QueryParam("kelas_id"))
	status := strPtr(c.QueryParam("status"))
	search := strPtr(c.QueryParam("q"))

	items, total, appErr := h.svc.ListUkk(c.Request().Context(), jurusanID, kelasID, status, search, limit, offset)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteLimitList(c, items, limit, offset, total)
}

// CreateUkk menangani POST /api/v1/ukk. -> 201
func (h *Handler) CreateUkk(c *echo.Context) error {
	if _, _, ok := requirePerm(c, "ukk.create"); !ok {
		return nil
	}
	var in ukkInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	u, appErr := h.svc.CreateUkk(c.Request().Context(), actorFrom(c), in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteStatus(c, http.StatusCreated, u)
}

// GetUkk menangani GET /api/v1/ukk/:id.
func (h *Handler) GetUkk(c *echo.Context) error {
	if _, _, ok := requirePerm(c, "ukk.read"); !ok {
		return nil
	}
	id := c.Param("id")
	if !isValidUUID(id) {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "ID UKK tidak valid.")
	}
	u, appErr := h.svc.GetUkk(c.Request().Context(), id)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, u)
}

// UpdateUkk menangani PUT /api/v1/ukk/:id.
func (h *Handler) UpdateUkk(c *echo.Context) error {
	if _, _, ok := requirePerm(c, "ukk.update"); !ok {
		return nil
	}
	id := c.Param("id")
	if !isValidUUID(id) {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "ID UKK tidak valid.")
	}
	var in ukkInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	u, appErr := h.svc.UpdateUkk(c.Request().Context(), actorFrom(c), id, in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, u)
}

// DeleteUkk menangani DELETE /api/v1/ukk/:id.
func (h *Handler) DeleteUkk(c *echo.Context) error {
	if _, _, ok := requirePerm(c, "ukk.update"); !ok {
		return nil
	}
	id := c.Param("id")
	if !isValidUUID(id) {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "ID UKK tidak valid.")
	}
	if appErr := h.svc.DeleteUkk(c.Request().Context(), actorFrom(c), id); appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, map[string]any{"deleted": true, "id": id})
}
