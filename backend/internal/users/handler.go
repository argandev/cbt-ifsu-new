// Package users — Manajemen Pengguna (akun STAF) — UC20.
//
// Endpoint (prefix /api/v1/pengguna):
//
//	GET    /api/v1/pengguna                     (pengguna.read)
//	POST   /api/v1/pengguna                     (pengguna.create) -> 201
//	GET    /api/v1/pengguna/:id                 (pengguna.read)
//	PUT    /api/v1/pengguna/:id                 (pengguna.update)
//	POST   /api/v1/pengguna/:id/reset-password (pengguna.reset-password)
//	POST   /api/v1/pengguna/:id/toggle-aktif   (pengguna.toggle-aktif)
//
// Permission diperiksa per-handler via audit.HasPermission (pola paket
// master); CSRF + auth sesi ditangani middleware global. Hanya akun STAF
// yang dikelola di sini (role != 'siswa'); akun siswa tetap dikelola
// lewat halaman Siswa (paket master). Akun siswa yang ditarget lewat
// endpoint ini mendapat 404 (anti-IDOR, sama seperti pola siswa).
//
// Struktur layer (hasil refactor arsitektur):
//   - handler.go:    Echo handler + Register (file ini)
//   - service.go:    validasi + logika bisnis + guard
//   - repository.go: DTO + semua query SQL
//   - platform:      envelope JSON + paginasi (menggantikan helper lokal)
//
// Perilaku (route, status, bentuk JSON, pesan error) IDENTIK dengan
// sebelum refactor.
package users

import (
	"net/http"
	"strconv"
	"strings"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/config"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// Handler adalah lapisan HTTP paket users.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// actorFrom membangun platform.RequestActor dari echo.Context.
func actorFrom(c *echo.Context) platform.RequestActor {
	userID, _, _, _, _ := audit.CurrentUser(c)
	return platform.RequestActor{UserID: userID, RealIP: c.RealIP()}
}

// Register mendaftarkan endpoint Manajemen Pengguna (staf).
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q
	_ = cfg

	repo := NewRepository(pool)
	svc := NewService(repo, pool)
	h := NewHandler(svc)

	// GET /api/v1/pengguna — list akun STAF saja.
	// Filter: ?q= (nama/username/email), ?role= (salah satu role staf),
	// ?is_active= (true/false). Pagination ?page=&per_page= (pola master).
	e.GET("/api/v1/pengguna", h.List)

	// POST /api/v1/pengguna — buat akun staf baru. -> 201
	e.POST("/api/v1/pengguna", h.Create)

	// GET /api/v1/pengguna/:id — detail akun staf (siswa -> 404).
	e.GET("/api/v1/pengguna/:id", h.Get)

	// PUT /api/v1/pengguna/:id — ubah nama/email/role akun staf.
	// Username TIDAK bisa diubah (diabaikan bila dikirim). Field yang
	// tidak dikirim dipertahankan. Role baru tidak boleh 'siswa'.
	e.PUT("/api/v1/pengguna/:id", h.Update)

	// POST /api/v1/pengguna/:id/reset-password — reset password akun staf.
	// Hash baru + must_change_password=TRUE + hapus SEMUA sesi user itu.
	e.POST("/api/v1/pengguna/:id/reset-password", h.ResetPassword)

	// POST /api/v1/pengguna/:id/toggle-aktif — aktif/nonaktifkan akun staf.
	// DILARANG: menonaktifkan akun sendiri (403) dan menonaktifkan admin
	// terakhir yang masih aktif (409). TIDAK ADA hard delete.
	e.POST("/api/v1/pengguna/:id/toggle-aktif", h.ToggleAktif)
}

// List menangani GET /api/v1/pengguna.
func (h *Handler) List(c *echo.Context) error {
	if !audit.HasPermission(c, "pengguna.read") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat pengguna.")
	}
	page, perPage, offset := platform.ParsePage(c)

	var f StafFilter
	if q := strings.TrimSpace(c.QueryParam("q")); q != "" {
		f.Q = &q
	}
	if role := strings.TrimSpace(c.QueryParam("role")); role != "" {
		f.Role = &role
	}
	if v := strings.TrimSpace(c.QueryParam("is_active")); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			f.IsActive = &b
		}
	}

	items, total, appErr := h.svc.List(c.Request().Context(), f, perPage, offset)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WritePageList(c, items, page, perPage, total)
}

// Create menangani POST /api/v1/pengguna.
func (h *Handler) Create(c *echo.Context) error {
	if !audit.HasPermission(c, "pengguna.create") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menambah pengguna.")
	}
	var in struct {
		Username string `json:"username"`
		Nama     string `json:"nama"`
		Email    string `json:"email"`
		Role     string `json:"role"`
		Password string `json:"password"`
	}
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	d, appErr := h.svc.Create(c.Request().Context(), actorFrom(c), CreateInput{
		Username: in.Username,
		Nama:     in.Nama,
		Email:    in.Email,
		Role:     in.Role,
		Password: in.Password,
	})
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteCreated(c, d)
}

// Get menangani GET /api/v1/pengguna/:id.
func (h *Handler) Get(c *echo.Context) error {
	if !audit.HasPermission(c, "pengguna.read") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat pengguna.")
	}
	d, appErr := h.svc.Get(c.Request().Context(), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, d)
}

// Update menangani PUT /api/v1/pengguna/:id.
func (h *Handler) Update(c *echo.Context) error {
	if !audit.HasPermission(c, "pengguna.update") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mengubah pengguna.")
	}
	// Urutan sesuai kode asal: cek format UUID (404) dulu, baru bind (400).
	id := c.Param("id")
	if !validUUID(id) {
		return platform.WriteError(c, http.StatusNotFound, "not_found", "Pengguna tidak ditemukan.")
	}
	var in struct {
		Nama  *string `json:"nama"`
		Email *string `json:"email"`
		Role  *string `json:"role"`
	}
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	d, appErr := h.svc.Update(c.Request().Context(), actorFrom(c), id, UpdateInput{
		Nama:  in.Nama,
		Email: in.Email,
		Role:  in.Role,
	})
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, d)
}

// ResetPassword menangani POST /api/v1/pengguna/:id/reset-password.
func (h *Handler) ResetPassword(c *echo.Context) error {
	if !audit.HasPermission(c, "pengguna.reset-password") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mereset password pengguna.")
	}
	var in struct {
		PasswordBaru string `json:"password_baru"`
	}
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	data, appErr := h.svc.ResetPassword(c.Request().Context(), actorFrom(c), c.Param("id"), in.PasswordBaru)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// ToggleAktif menangani POST /api/v1/pengguna/:id/toggle-aktif.
func (h *Handler) ToggleAktif(c *echo.Context) error {
	if !audit.HasPermission(c, "pengguna.toggle-aktif") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mengubah status aktif pengguna.")
	}
	var in struct {
		IsActive *bool `json:"is_active"`
	}
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	d, appErr := h.svc.ToggleAktif(c.Request().Context(), actorFrom(c), c.Param("id"), in.IsActive)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, d)
}
