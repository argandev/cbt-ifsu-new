// Package middleware — logika pembatas laju + helper pengecekan sesi/CSRF
// (layer service paket middleware; konstruktor middleware Echo ada di
// handler.go). Kode di sini dipindah dari middleware.go tanpa perubahan.
package middleware

import (
	"context"
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

// --- RateLimiter ------------------------------------------------------------

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
	mu              sync.Mutex
	hits            map[string]*ipBucket
	maxRequests     int
	examMaxRequests int
	window          time.Duration
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

// --- Helper pengecekan sesi/CSRF --------------------------------------------

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
