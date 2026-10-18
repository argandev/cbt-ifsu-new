// Package backup — Backup / Restore database (Tim Fase 2).
//
// Definisi (kontrak tugas Fase 2):
//
//	Backup  = dump PostgreSQL via pg_dump -Fc (format custom) + metadata
//	          di tabel backup_meta. Password DB TIDAK PERNAH ditulis ke
//	          log / respons — hanya diteruskan sebagai env PGPASSWORD
//	          ke proses anak pg_dump / pg_restore.
//	Restore = mengembalikan satu dump ke database yang sama via
//	          pg_restore --clean --if-exists.
//
// Aturan keamanan keras yang diimplementasikan di sini:
//  1. Restore HANYA untuk permission backup.restore (admin + proktor
//     menurut seed/matrik RBAC; penegakan di server, bukan hanya di UI).
//  2. Sebelum restore, sistem OTOMATIS membuat backup pengaman
//     (jenis 'pre-restore') dan mencatatnya di backup_meta + audit.
//  3. Restore mencatat audit lengkap (siapa, kapan, file mana, IP).
//  4. Restore DITOLAK (409) bila ada jadwal berstatus aktif/pause atau
//     attempt berstatus berjalan — cegah restore saat ujian berlangsung.
//  5. File backup di ./storage/backup/, nama
//     backup-<YYYYMMDD-HHMMSS>-<8 karakter acak>.dump; file lama tidak
//     pernah ditimpa (nama selalu unik, kolom nama_file UNIQUE).
//  6. Download backup HANYA via endpoint dengan permission check —
//     direktori storage TIDAK disajikan sebagai static file server.
//  7. TIDAK ada scheduler di dalam aplikasi: backup otomatis harian
//     TIDAK dibuat; backup manual via UI + skrip cron deploy/backup.sh
//     yang sudah ada.
//  8. TIDAK ADA perintah shell dari input user: nama file selalu
//     diambil dari baris backup_meta di DB (berdasarkan :id UUID),
//     bukan dari parameter / body. Satu-satunya input user adalah
//     keterangan bebas (teks) dan :id.
//  9. Timeout pg_dump / pg_restore: 10 menit (context timeout).
//
// Subtleti restore yang penting: pg_restore --clean menghapus dan
// membuat ulang SEMUA tabel dari isi dump, termasuk backup_meta dan
// audit_log. Akibatnya baris metadata pre-restore backup + audit yang
// ditulis SEBELUM pg_restore akan ikut terhapus oleh restore itu
// sendiri. Karena itu, setelah pg_restore sukses, handler mengembalikan
// (upsert) baris metadata pre-restore backup + baris metadata backup
// target + audit backup.create (pre-restore) dan backup.restore —
// sehingga jejak "restore dari file X, pengaman file Y" tetap ada di
// database hasil restore. Bila pg_restore gagal, pre-restore backup
// tetap tercatat (restore belum mengubah apa pun secara utuh) dan
// respons 500 menyebut nama file pengamannya agar admin bisa
// memulihkan manual.
//
// Struktur layer (hasil refactor arsitektur):
//   - backup.go:     dokumentasi paket
//   - handler.go:    Echo handler + Register + cek permission
//   - service.go:    logika backup/restore/hapus + audit log
//   - repository.go: DTO + query backup_meta + gerbang ujian
//   - platform:      envelope JSON + paginasi (menggantikan helper lokal)
//
// Perilaku (route, status, bentuk JSON, pesan error) IDENTIK dengan
// sebelum refactor.
package backup

import (
	"net/http"
	"strconv"
	"strings"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/config"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// --- Handler ----------------------------------------------------------------

// Handler adalah lapisan HTTP paket backup.
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

// Register mendaftarkan endpoint backup / restore.
//
//	POST   /api/v1/backup                 (backup.create)
//	GET    /api/v1/backup                 (backup.create ATAU backup.restore)
//	GET    /api/v1/backup/:id/download    (backup.create ATAU backup.restore)
//	POST   /api/v1/backup/:id/restore     (backup.restore)
//	DELETE /api/v1/backup/:id             (backup.restore)
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q

	repo := NewRepository(pool)
	svc := NewService(repo, pool, cfg)
	h := NewHandler(svc)

	// POST /api/v1/backup — buat backup manual.
	e.POST("/api/v1/backup", h.Buat)

	// GET /api/v1/backup — daftar metadata + pagination.
	e.GET("/api/v1/backup", h.Daftar)

	// GET /api/v1/backup/:id/download — unduh file via permission check
	// + audit. Streaming; file TIDAK disajikan lewat static file server.
	e.GET("/api/v1/backup/:id/download", h.Unduh)

	// POST /api/v1/backup/:id/restore — pulihkan dump ke DB yang sama.
	e.POST("/api/v1/backup/:id/restore", h.Restore)

	// DELETE /api/v1/backup/:id — hapus file + meta + audit. Pre-restore
	// backup berusia < 7 hari DITOLAK (pengaman restore terakhir harus
	// tetap tersedia) dan penolakan dicatat alasannya di audit.
	e.DELETE("/api/v1/backup/:id", h.Hapus)
}

// Buat menangani POST /api/v1/backup.
func (h *Handler) Buat(c *echo.Context) error {
	if !audit.HasPermission(c, "backup.create") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin membuat backup.")
	}
	var body struct {
		Keterangan string `json:"keterangan"`
	}
	_ = c.Bind(&body) // keterangan opsional; body kosong tetap sah
	if len(body.Keterangan) > 500 {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Keterangan maksimal 500 karakter.")
	}

	data, appErr := h.svc.Buat(c.Request().Context(), actorFrom(c), strings.TrimSpace(body.Keterangan))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// Daftar menangani GET /api/v1/backup.
func (h *Handler) Daftar(c *echo.Context) error {
	if !audit.HasPermission(c, "backup.create") && !audit.HasPermission(c, "backup.restore") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat daftar backup.")
	}
	// Gaya paginasi backup: default 50, maks 200 — dipertahankan apa adanya
	// (platform.ParseLimit memakai maks 100, tidak cocok di sini).
	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset, err := strconv.Atoi(c.QueryParam("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	items, total, appErr := h.svc.Daftar(c.Request().Context(), limit, offset)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteLimitList(c, items, limit, offset, total)
}

// Unduh menangani GET /api/v1/backup/:id/download.
func (h *Handler) Unduh(c *echo.Context) error {
	if !audit.HasPermission(c, "backup.create") && !audit.HasPermission(c, "backup.restore") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mengunduh backup.")
	}
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		return platform.WriteError(c, http.StatusNotFound, "not_found", "Backup tidak ditemukan.")
	}
	b, appErr := h.svc.Unduh(c.Request().Context(), actorFrom(c), id)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	defer b.File.Close()
	resp := c.Response()
	resp.Header().Set("Content-Disposition", `attachment; filename="`+b.Meta.NamaFile+`"`)
	resp.Header().Set("Content-Length", strconv.FormatInt(b.Size, 10))
	return c.Stream(http.StatusOK, "application/octet-stream", b.File)
}

// Restore menangani POST /api/v1/backup/:id/restore.
func (h *Handler) Restore(c *echo.Context) error {
	if !audit.HasPermission(c, "backup.restore") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melakukan restore. Restore hanya untuk admin dan proktor.")
	}
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		return platform.WriteError(c, http.StatusNotFound, "not_found", "Backup tidak ditemukan.")
	}
	data, appErr := h.svc.Restore(c.Request().Context(), actorFrom(c), id)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// Hapus menangani DELETE /api/v1/backup/:id.
func (h *Handler) Hapus(c *echo.Context) error {
	if !audit.HasPermission(c, "backup.restore") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menghapus backup.")
	}
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		return platform.WriteError(c, http.StatusNotFound, "not_found", "Backup tidak ditemukan.")
	}
	data, appErr := h.svc.Hapus(c.Request().Context(), actorFrom(c), id)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}
