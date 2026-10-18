// Package upload — upload gambar umum (Tim D).
//
// Satu endpoint untuk SEMUA kebutuhan gambar aplikasi, agar tidak ada lagi
// input URL gambar ketik manual (logo sekolah di Settings, media soal di
// SoalEditor, gambar opsi pilihan_gambar — lihat docs/AUDIT-UI-UX.md §4.4.1):
//
//	POST /api/v1/upload/image        (multipart, field "file")
//	GET  /api/v1/upload/image/:nama  (menyajikan file yang sudah diunggah)
//
// Keamanan (berlapis, jangan kurangi satu pun tanpa kontrak baru):
//   - Auth: route /api/* sudah dilewati middleware global RequireAuth —
//     tanpa session login, request ditolak 401 sebelum sampai ke handler.
//   - CSRF: middleware global CSRF mewajibkan header X-CSRF-Token ==
//     session.csrf_token untuk POST (lihat middleware/handler.go).
//   - Otorisasi upload: hanya role yang memang mengelola konten bergambar
//     (settings.update = logo, banksoal.create/update = media soal,
//     ukk.create). Siswa TIDAK punya ketiganya sehingga tidak bisa upload.
//     Admin (role "admin") lolos semua via audit.HasPermission.
//   - GET file: boleh dibaca semua user yang login (siswa pun perlu melihat
//     gambar soal saat ujian) — tetapi nama file tidak dapat ditebak
//     (UUID) dan path divalidasi ketat.
//   - Magic bytes: tipe file ditentukan dari isi file (signature), BUKAN
//     dari ekstensi nama file atau header Content-Type kiriman client.
//     File "jahat.png" yang isinya PHP/HTML ditolak.
//   - Nama file di disk: UUID acak + ekstensi aman turunan magic bytes
//     (bukan nama asli kiriman client) — tidak ada eksekusi / traversal.
//   - Ukuran: maks 5 MB per file (ditegakkan dari header DAN isi sebenarnya).
//   - Hash SHA-256: dihitung dari isi file, dicatat di audit log
//     (aksi "upload.image") dan dikembalikan di respons.
//   - Penyimpanan: storage/upload/image/ — folder storage yang sama dengan
//     berkas UKK (storage/ukk/...). File TIDAK pernah disajikan sebagai
//     file statis webroot; hanya lewat handler GET di bawah ini yang
//     mengirim Content-Type dari tipe terdeteksi + nosniff.
//
// Respons POST (201): {"data": {"url", "sha256", "ukuran", "content_type"}}
// dengan url = URL internal "/api/v1/upload/image/<uuid>.<ext>" (BUKAN URL
// eksternal/CDN). Frontend menyimpan url ini apa adanya (mis. settings
// logo_url, media soal) — JANGAN pernah menyimpan URL eksternal.
//
// Struktur layer (hasil refactor arsitektur):
//   - upload.go:     dokumentasi paket
//   - handler.go:    Echo handler + Register + cek permission
//   - service.go:    validasi magic bytes + logika simpan/saji + audit
//   - repository.go: penyimpanan file di disk (tanpa DB di paket ini)
//   - platform:      envelope JSON (menggantikan helper lokal)
//
// Perilaku (route, status, bentuk JSON, pesan error) IDENTIK dengan
// sebelum refactor.
package upload

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

// Handler adalah lapisan HTTP paket upload.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// actorFrom membangun RequestActor dari echo.Context.
func actorFrom(c *echo.Context) RequestActor {
	return RequestActor{UserID: currentUserID(c), RealIP: clientIP(c)}
}

func currentUserID(c *echo.Context) string {
	userID, _, _, _, _ := audit.CurrentUser(c)
	return userID
}

func clientIP(c *echo.Context) string {
	if ip := c.RealIP(); ip != "" {
		return ip
	}
	return c.Request().RemoteAddr
}

// bolehUpload melaporkan apakah user yang login boleh mengunggah gambar:
// punya salah satu permission pengelola konten bergambar. Admin lolos
// semua via audit.HasPermission. (Dipindah dari upload.go tanpa perubahan.)
func bolehUpload(c *echo.Context) bool {
	for _, perm := range []string{"settings.update", "banksoal.create", "banksoal.update", "ukk.create"} {
		if audit.HasPermission(c, perm) {
			return true
		}
	}
	return false
}

// Register mendaftarkan endpoint upload gambar.
//
//	POST /api/v1/upload/image       (upload; wajib login + CSRF + permission)
//	GET  /api/v1/upload/image/:nama (sajikan file; wajib login)
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q

	repo := NewRepository(cfg.StorageDir)
	svc := NewService(repo, pool)
	h := NewHandler(svc)

	e.POST("/api/v1/upload/image", h.Upload)
	e.GET("/api/v1/upload/image/:nama", h.Sajikan)
}

// Upload menangani POST /api/v1/upload/image.
func (h *Handler) Upload(c *echo.Context) error {
	if _, _, _, _, ok := audit.CurrentUser(c); !ok {
		return platform.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
	}
	if !bolehUpload(c) {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mengunggah gambar.")
	}
	// ParseMultipartForm dengan batas 5 MB (+1 MB ruang header multipart).
	// File yang lebih besar ditolak di bawah lewat Size + isi sebenarnya.
	if err := c.Request().ParseMultipartForm(maxGambarBytes + (1 << 20)); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body multipart tidak valid (maks. 5 MB).")
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Field file wajib diisi.")
	}
	data, appErr := h.svc.Simpan(c.Request().Context(), actorFrom(c), fileHeader)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteStatus(c, http.StatusCreated, data)
}

// Sajikan menangani GET /api/v1/upload/image/:nama.
func (h *Handler) Sajikan(c *echo.Context) error {
	if _, _, _, _, ok := audit.CurrentUser(c); !ok {
		return platform.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
	}
	g, appErr := h.svc.Sajikan(c.Param("nama"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	c.Response().Header().Set("Content-Type", g.ContentType)
	c.Response().Header().Set("X-Content-Type-Options", "nosniff")
	c.Response().Header().Set("Cache-Control", "private, max-age=3600")
	c.Response().Header().Set("Content-Disposition", "inline")
	return c.Blob(http.StatusOK, g.ContentType, g.Data)
}
