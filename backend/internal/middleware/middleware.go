// Package middleware — middleware HTTP bersama (Tim 1).
//
// Urutan pemasangan di main.go (global, untuk semua route):
// Logger -> Recover -> RateLimit -> RequireAuth -> CSRF.
//
// Kontrak context (dipakai paket lain, mis. audit.CurrentUser):
// "auth_user_id", "auth_role", "auth_permissions" ([]string),
// "auth_siswa_id", "auth_csrf_token", "auth_session_id".
// Nilai-nilai ini HARUS sama persis dengan konstanta di paket audit.
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"cbt-smk-v2/backend/internal/rbac"
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

// RateLimiter membatasi request per IP: maxRequests per window.
// Endpoint ujian (/api/v1/ujian/attempt/*) mendapat anggaran lebih
// besar (examMaxRequests) karena pola akses ujian adalah burst singkat
// yang sah (autosave jawaban dkk.), bukan serangan — BUG-03: pada beban
// 300 siswa berbagi 30 IP, batas 100/menit/IP menghasilkan 1800x 429
// (pulih hanya karena harness me-retry; frontend tidak me-retry 429).
// Penolakan 429 mengembalikan header Retry-After (detik sampai jendela
// reset) agar klien tahu kapan mencoba lagi.
// Implementasi in-memory, fixed window. Cukup untuk satu instance
// (deploy utama = satu server sekolah). Bila kelak multi-instance,
// ganti dengan pembatas berbasis DB/Redis.
type RateLimiter struct {
	mu               sync.Mutex
	hits             map[string]*ipBucket
	maxRequests      int
	examMaxRequests  int
	window           time.Duration
}

type ipBucket struct {
	count     int
	windowEnd time.Time
}

// NewRateLimiter membuat pembatas baru (default: 100 req/menit per IP,
// berlaku juga untuk endpoint ujian).
func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	return NewRateLimiterWithExamLimits(maxRequests, maxRequests, window)
}

// NewRateLimiterWithExamLimits membuat pembatas dengan anggaran
// terpisah untuk endpoint ujian (/api/v1/ujian/attempt/*).
func NewRateLimiterWithExamLimits(maxRequests, examMaxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		hits:            make(map[string]*ipBucket),
		maxRequests:     maxRequests,
		examMaxRequests: examMaxRequests,
		window:          window,
	}
}

// isExamPath melaporkan apakah path memakai anggaran ujian.
func isExamPath(path string) bool {
	return strings.HasPrefix(path, "/api/v1/ujian/attempt/")
}

// Middleware mengembalikan middleware Echo untuk pembatas ini.
func (rl *RateLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// D8: aset statis + halaman SPA (HTML) TIDAK dihitung limiter —
			// limiter hanya untuk API (dan handshake WebSocket). Tanpa ini,
			// navigasi cepat menghabiskan kuota 100/menit/IP lalu browser
			// menampilkan JSON 429 mentah untuk file JS/CSS/HTML.
			path := c.Request().URL.Path
			if !strings.HasPrefix(path, "/api/") && !strings.HasPrefix(path, "/ws/") {
				return next(c)
			}
			ip := c.RealIP()
			limit := rl.maxRequests
			key := ip
			if isExamPath(c.Request().URL.Path) {
				limit = rl.examMaxRequests
				key = ip + "#ujian"
			}
			now := time.Now()
			rl.mu.Lock()
			b, ok := rl.hits[key]
			if !ok || now.After(b.windowEnd) {
				b = &ipBucket{count: 0, windowEnd: now.Add(rl.window)}
				rl.hits[key] = b
			}
			b.count++
			count := b.count
			retryAfter := int(time.Until(b.windowEnd).Seconds()) + 1
			if retryAfter < 1 {
				retryAfter = 1
			}
			// Pembersihan oportunistik agar map tidak tumbuh tanpa batas.
			if len(rl.hits) > 10000 {
				for k, v := range rl.hits {
					if now.After(v.windowEnd) {
						delete(rl.hits, k)
					}
				}
			}
			rl.mu.Unlock()
			if count > limit {
				c.Response().Header().Set("Retry-After", strconv.Itoa(retryAfter))
				return writeError(c, http.StatusTooManyRequests, "rate_limited", "Terlalu banyak permintaan. Coba lagi sebentar.")
			}
			return next(c)
		}
	}
}

// RateLimit adalah pembatas default: 100 request/menit per IP untuk
// endpoint umum, 300 request/menit per IP untuk endpoint ujian
// (/api/v1/ujian/attempt/*: jawaban/ragu/pelanggaran/submit/status).
// Rate limit login (5 gagal/15 menit -> kunci) TIDAK berubah: ia
// ditegakkan terpisah di paket auth (KONTRAK §4).
func RateLimit() echo.MiddlewareFunc {
	return NewRateLimiterWithExamLimits(100, 300, time.Minute).Middleware()
}

// --- RequireAuth ------------------------------------------------------------

// isExemptAuth melaporkan path yang TIDAK wajib login:
//   - /api/v1/auth/login dan /api/v1/auth/ganti-password (alur login /
//     ganti password pertama; ganti-password menegakkan auth sendiri:
//     session aktif ATAU password lama yang benar),
//   - path di luar /api (halaman SPA) dan /ws (auth via ticket, Tim 4).
func isExemptAuth(path string) bool {
	if !strings.HasPrefix(path, "/api/") {
		return true
	}
	if strings.HasPrefix(path, "/ws/") {
		return true
	}
	switch path {
	case "/api/v1/auth/login", "/api/v1/auth/ganti-password":
		return true
	}
	return false
}

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

// loadSessionIntoContext memuat session bila cookie ada; mengabaikan error
// (dipakai untuk path exempt — tidak adanya session bukan kesalahan).
func loadSessionIntoContext(c *echo.Context, pool *pgxpool.Pool) {
	cookie, err := c.Cookie(SessionCookieName)
	if err != nil || cookie.Value == "" {
		return
	}
	s, err := session.Get(c.Request().Context(), pool, cookie.Value)
	if err != nil {
		return
	}
	setSessionContextWithPool(c, s, pool)
}

// setSessionContextWithPool mengisi context Echo dari session + permissions.
// Bila pool != nil, permission diambil dari DB (role_permissions join
// permissions); bila DB belum di-seed / error, fallback ke matriks statis
// paket rbac agar auth tetap berfungsi.
func setSessionContextWithPool(c *echo.Context, s *session.Session, pool *pgxpool.Pool) {
	c.Set(CtxSessionID, s.ID.String())
	c.Set(CtxUserID, s.UserID.String())
	c.Set(CtxRole, s.Role)
	c.Set(CtxSiswaID, s.UserID.String()) // siswa_id = user_id sampai Tim 2 memetakan profil siswa
	c.Set(CtxCSRF, s.CSRFToken)
	c.Set(CtxSession, s)
	c.Set(CtxPermissions, permissionsForRole(c.Request().Context(), pool, s.Role))
}

// permissionsForRole mengambil permission role dari DB (role_permissions
// join permissions). Bila pool nil / query gagal, fallback ke matriks rbac.
func permissionsForRole(ctx context.Context, pool *pgxpool.Pool, role string) []string {
	if pool != nil {
		rows, err := pool.Query(ctx, `
			SELECT p.code FROM role_permissions rp
			JOIN permissions p ON p.id = rp.permission_id
			JOIN roles r ON r.id = rp.role_id
			WHERE r.code = $1`, role)
		if err == nil {
			defer rows.Close()
			var perms []string
			for rows.Next() {
				var code string
				if err := rows.Scan(&code); err == nil {
					perms = append(perms, code)
				}
			}
			if rows.Err() == nil && len(perms) > 0 {
				return perms
			}
		}
	}
	return rbac.PermissionsForRole(rbac.Role(role))
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

// --- CSRF -------------------------------------------------------------------

// isExemptCSRF melaporkan path yang dilewati pemeriksaan CSRF:
// - method aman (GET/HEAD/OPTIONS) tidak pernah dicek,
// - /api/v1/auth/login dan /api/v1/auth/csrf (belum ada session saat login).
func isExemptCSRF(c *echo.Context) bool {
	switch c.Request().Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	}
	switch c.Request().URL.Path {
	case "/api/v1/auth/login", "/api/v1/auth/csrf":
		return true
	}
	return false
}

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
