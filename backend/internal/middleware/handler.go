// Package middleware — middleware HTTP bersama (Tim 1).
//
// Urutan pemasangan di main.go (global, untuk semua route):
// Logger -> Recover -> RateLimit -> RequireAuth -> CSRF.
//
// Kontrak context (dipakai paket lain, mis. audit.CurrentUser):
// "auth_user_id", "auth_role", "auth_permissions" ([]string),
// "auth_siswa_id", "auth_csrf_token", "auth_session_id".
// Nilai-nilai ini HARUS sama persis dengan konstanta di paket audit.
//
// Struktur layer (hasil refactor arsitektur):
//   - handler.go:  konstruktor middleware Echo = lapisan HTTP (file ini)
//   - service.go:  RateLimiter + helper pengecekan sesi/CSRF
//
// Perilaku (middleware, status, bentuk JSON, pesan error) IDENTIK dengan
// sebelum refactor.
package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"cbt-smk-v2/backend/internal/session"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// Kunci context auth (SAMA PERSIS dengan paket audit).
const (
	CtxUserID      = "auth_user_id"
	CtxRole        = "auth_role"
	CtxPermissions = "auth_permissions"
	CtxSiswaID     = "auth_siswa_id"
	CtxCSRF        = "auth_csrf_token"
	CtxSessionID   = "auth_session_id"
)

// Kunci context tambahan.
const (
	CtxSession = "auth_session"
)

// Nama cookie session + CSRF (lihat docs/KONTRAK.md).
const (
	SessionCookieName = "cbt_session"
	CSRFCookieName    = "cbt_csrf"
)

// writeError menulis respons error dalam format kontrak:
// {"error": {"code": "...", "message": "..."}}.
func writeError(c *echo.Context, status int, code, message string) error {
	return c.JSON(status, map[string]any{
		"error": map[string]any{"code": code, "message": message},
	})
}

// --- Logger -----------------------------------------------------------------

// Logger mencatat method, path, status, dan latensi tiap request.
// Echo v5: c.Response() adalah http.ResponseWriter yang implementasi
// konkretnya *echo.Response (punya field Status).
func Logger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()
			err := next(c)
			status := 0
			if r, ok := c.Response().(*echo.Response); ok {
				status = r.Status
			}
			c.Logger().Info("request",
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
				"status", status,
				"latency", time.Since(start).Round(time.Millisecond).String())
			return err
		}
	}
}

// --- Recover ----------------------------------------------------------------

// Recover menangkap panic di handler dan mengembalikan 500 JSON,
// bukan stack trace ke client.
func Recover() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					c.Logger().Error("panic", "recover", r)
					err = writeError(c, http.StatusInternalServerError, "internal_error", "Terjadi kesalahan internal.")
				}
			}()
			return next(c)
		}
	}
}

// --- RateLimit --------------------------------------------------------------

// RateLimit adalah pembatas default: 100 request/menit per IP untuk
// endpoint umum, 300 request/menit per IP untuk endpoint ujian
// (/api/v1/ujian/attempt/*: jawaban/ragu/pelanggaran/submit/status).
// Rate limit login (5 gagal/15 menit -> kunci) TIDAK berubah: ia
// ditegakkan terpisah di paket auth (KONTRAK §4).
func RateLimit() echo.MiddlewareFunc {
	return NewRateLimiterWithExamLimits(100, 300, time.Minute).Middleware()
}

// --- RequireAuth ------------------------------------------------------------

// RequireAuth membaca cookie `cbt_session`, memuat session dari DB,
// dan mengisi context auth. 401 bila tidak ada / kedaluwarsa.
// Session yang valid di-Touch (perbarui last_seen_at).
func RequireAuth(pool *pgxpool.Pool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if isExemptAuth(c.Request().URL.Path) {
				// Tetap muat session bila cookie ada (agar ganti-password
				// mode session bisa membaca user dari context).
				loadSessionIntoContext(c, pool)
				return next(c)
			}
			cookie, err := c.Cookie(SessionCookieName)
			if err != nil || cookie.Value == "" {
				return writeError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
			}
			s, err := session.Get(c.Request().Context(), pool, cookie.Value)
			if err != nil {
				return writeError(c, http.StatusUnauthorized, "unauthenticated", "Sesi tidak valid atau kedaluwarsa. Silakan login lagi.")
			}
			setSessionContextWithPool(c, s, pool)
			session.Touch(c.Request().Context(), pool, cookie.Value)
			return next(c)
		}
	}
}

// --- RequireRole ------------------------------------------------------------

// RequireRole menolak request bila role aktif tidak ada di daftar.
// Harus dipasang SETELAH RequireAuth (atau di dalam handler via context).
func RequireRole(roles ...string) echo.MiddlewareFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			role, _ := c.Get(CtxRole).(string)
			if role == "" {
				return writeError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
			}
			if !allowed[role] {
				return writeError(c, http.StatusForbidden, "forbidden", "Role tidak diizinkan untuk aksi ini.")
			}
			return next(c)
		}
	}
}

// --- RequirePermission ------------------------------------------------------

// RequirePermission menolak request bila role tidak punya permission.
// Pengecekan dilakukan ke DB (role_permissions join permissions) sesuai
// kontrak, dengan admin selalu lolos. Dipasang setelah RequireAuth.
func RequirePermission(pool *pgxpool.Pool, perm string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			role, _ := c.Get(CtxRole).(string)
			if role == "" {
				return writeError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
			}
			if role == "admin" {
				return next(c)
			}
			var exists bool
			err := pool.QueryRow(c.Request().Context(), `
				SELECT EXISTS (
					SELECT 1 FROM role_permissions rp
					JOIN permissions p ON p.id = rp.permission_id
					JOIN roles r ON r.id = rp.role_id
					WHERE r.code = $1 AND p.code = $2
				)`, role, perm).Scan(&exists)
			if err != nil || !exists {
				return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin: "+perm+".")
			}
			return next(c)
		}
	}
}

// --- CSRF -------------------------------------------------------------------

// CSRF mewajibkan header X-CSRF-Token == session.csrf_token untuk
// POST/PUT/PATCH/DELETE (constant-time compare). Bila request tidak membawa
// cookie session, pemeriksaan dilewati di sini — RequireAuth akan
// mengembalikan 401 (kecuali ganti-password mode kredensial yang memang
// dirancang tanpa session dan mengandalkan password lama).
func CSRF(pool *pgxpool.Pool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if isExemptCSRF(c) {
				return next(c)
			}
			cookie, err := c.Cookie(SessionCookieName)
			if err != nil || cookie.Value == "" {
				return next(c)
			}
			s, err := session.Get(c.Request().Context(), pool, cookie.Value)
			if err != nil {
				return next(c)
			}
			provided := c.Request().Header.Get("X-CSRF-Token")
			if !session.VerifyCSRF(s.CSRFToken, provided) {
				return writeError(c, http.StatusForbidden, "csrf_token_invalid", "Token CSRF tidak valid.")
			}
			return next(c)
		}
	}
}

// EnsureJSON dipakai handler bila perlu memastikan body JSON (opsional).
func EnsureJSON(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		if c.Request().Body != nil && c.Request().ContentLength > 0 {
			ct := c.Request().Header.Get("Content-Type")
			if ct != "" && !strings.HasPrefix(ct, "application/json") {
				return writeError(c, http.StatusUnsupportedMediaType, "unsupported_media_type", "Content-Type harus application/json.")
			}
		}
		return next(c)
	}
}

// DecodeJSON membaca body JSON ke v dengan pesan error kontrak.
func DecodeJSON(c *echo.Context, v any) error {
	dec := json.NewDecoder(c.Request().Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Body JSON tidak valid.")
	}
	return nil
}
