// Package auth — endpoint autentikasi + CLI akun admin (Tim 1).
//
// Auth memakai SESSION COOKIE server-side + CSRF. TIDAK ADA kode JWT di
// proyek ini. Cookie `cbt_session` (HttpOnly) membawa ID session; data
// session (user, role, csrf_token) tinggal di server (Postgres).
//
// Endpoint (lihat docs/KONTRAK.md):
// POST /api/v1/auth/login, POST /api/v1/auth/logout,
// GET  /api/v1/auth/me, POST /api/v1/auth/ganti-password,
// GET  /api/v1/auth/csrf
//
// Struktur layer (hasil refactor arsitektur):
//   - handler.go:    Echo handler + Register + cookie (file ini)
//   - service.go:    validasi + logika bisnis + rate limit login + CLI admin
//   - repository.go: DTO + semua query SQL
//   - password.go:   HashPassword/VerifyPassword (murni, tidak diubah)
//   - platform:      envelope JSON (menggantikan helper lokal)
//
// Perilaku (route, status, bentuk JSON, pesan error) IDENTIK dengan
// sebelum refactor.
package auth

import (
	"net/http"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/config"
	"cbt-smk-v2/backend/internal/middleware"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// Handler adalah lapisan HTTP paket auth.
type Handler struct {
	svc *Service
	cfg *config.Config
}

// NewHandler membuat Handler.
func NewHandler(svc *Service, cfg *config.Config) *Handler {
	return &Handler{svc: svc, cfg: cfg}
}

// actorFrom membangun platform.RequestActor dari echo.Context.
func actorFrom(c *echo.Context) platform.RequestActor {
	userID, _ := c.Get(middleware.CtxUserID).(string)
	return platform.RequestActor{UserID: userID, RealIP: c.RealIP(), UserAgent: c.Request().UserAgent()}
}

// --- cookie -----------------------------------------------------------------
// (Dipindah dari auth.go tanpa perubahan.)

func setSessionCookies(c *echo.Context, cfg *config.Config, sessionID, csrfToken string) {
	secure := cfg.IsProd()
	maxAge := 12 * 60 * 60 // 12 jam, sama dengan session.DefaultTTL
	c.SetCookie(&http.Cookie{
		Name: middleware.SessionCookieName, Value: sessionID, Path: "/",
		HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode,
		MaxAge: maxAge,
	})
	c.SetCookie(&http.Cookie{
		Name: middleware.CSRFCookieName, Value: csrfToken, Path: "/",
		HttpOnly: false, Secure: secure, SameSite: http.SameSiteLaxMode,
		MaxAge: maxAge,
	})
}

func clearSessionCookies(c *echo.Context, cfg *config.Config) {
	secure := cfg.IsProd()
	for _, name := range []string{middleware.SessionCookieName, middleware.CSRFCookieName} {
		c.SetCookie(&http.Cookie{
			Name: name, Value: "", Path: "/",
			HttpOnly: name == middleware.SessionCookieName,
			Secure:   secure, SameSite: http.SameSiteLaxMode, MaxAge: -1,
		})
	}
}

// --- Register ---------------------------------------------------------------

// Register mendaftarkan endpoint auth.
//
//	func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config)
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q // auth memakai pool langsung (konvensi tim lain); kueri sqlc ada di db/queries/0001_auth.sql

	repo := NewRepository(pool)
	svc := NewService(repo, pool)
	h := NewHandler(svc, cfg)

	// POST /api/v1/auth/login — TIDAK butuh session (exempt RequireAuth).
	// Body: {"login": "<email|username>", "password": "..."}
	// 403 {"error":{"code":"PASSWORD_CHANGE_REQUIRED"}} bila flag aktif.
	e.POST("/api/v1/auth/login", h.Login)

	// POST /api/v1/auth/logout — hapus session + clear cookie.
	e.POST("/api/v1/auth/logout", h.Logout)

	// GET /api/v1/auth/me — user + role + permissions.
	e.GET("/api/v1/auth/me", h.Me)

	// POST /api/v1/auth/ganti-password — BOLEH diakses saat flag
	// must_change_password aktif, dalam dua mode:
	// (a) Mode session: user sudah login (cookie valid) — body cukup
	//     {"password_lama", "password_baru"}.
	// (b) Mode kredensial: tanpa session (mis. login pertama yang ditolak
	//     403 PASSWORD_CHANGE_REQUIRED) — body {"login", "password_lama",
	//     "password_baru"}. Endpoint ini exempt dari RequireAuth global
	//     justru agar mode (b) bisa jalan.
	// Setelah berhasil, SEMUA session user dihapus (wajib login ulang).
	e.POST("/api/v1/auth/ganti-password", h.GantiPassword)

	// GET /api/v1/auth/csrf — kembalikan csrf_token untuk session aktif.
	// PENANDA: endpoint ini mengembalikan {"csrf_token": ...} mentah
	// (tanpa envelope {"data": ...}) sesuai spesifikasi tugas Tim 1,
	// karena nilainya juga tersedia sebagai cookie `cbt_csrf`.
	e.GET("/api/v1/auth/csrf", h.Csrf)
}

// Login menangani POST /api/v1/auth/login.
func (h *Handler) Login(c *echo.Context) error {
	var body struct {
		Login    string `json:"login"`
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.Bind(&body); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	actor := actorFrom(c)
	sessionID, csrfToken, data, appErr := h.svc.Login(c.Request().Context(), actor,
		body.Login, body.Email, body.Username, body.Password)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	setSessionCookies(c, h.cfg, sessionID, csrfToken)
	return platform.WriteOK(c, data)
}

// Logout menangani POST /api/v1/auth/logout.
func (h *Handler) Logout(c *echo.Context) error {
	actor := actorFrom(c)
	sessionCookie := ""
	if cookie, err := c.Cookie(middleware.SessionCookieName); err == nil && cookie.Value != "" {
		sessionCookie = cookie.Value
	}
	h.svc.Logout(c.Request().Context(), actor, sessionCookie)
	clearSessionCookies(c, h.cfg)
	return platform.WriteOK(c, map[string]any{"logged_out": true})
}

// Me menangani GET /api/v1/auth/me.
func (h *Handler) Me(c *echo.Context) error {
	data, appErr := h.svc.Me(c.Request().Context(), actorFrom(c))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// GantiPassword menangani POST /api/v1/auth/ganti-password.
func (h *Handler) GantiPassword(c *echo.Context) error {
	var body struct {
		Login        string `json:"login"`
		PasswordLama string `json:"password_lama"`
		PasswordBaru string `json:"password_baru"`
	}
	if err := c.Bind(&body); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	if appErr := h.svc.GantiPassword(c.Request().Context(), actorFrom(c),
		body.Login, body.PasswordLama, body.PasswordBaru); appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	clearSessionCookies(c, h.cfg)
	return platform.WriteOK(c, map[string]any{"password_changed": true})
}

// Csrf menangani GET /api/v1/auth/csrf.
func (h *Handler) Csrf(c *echo.Context) error {
	csrf, _ := c.Get(middleware.CtxCSRF).(string)
	if csrf == "" {
		return platform.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
	}
	return c.JSON(http.StatusOK, map[string]any{"csrf_token": csrf})
}
