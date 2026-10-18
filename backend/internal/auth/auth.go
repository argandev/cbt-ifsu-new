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
package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/config"
	"cbt-smk-v2/backend/internal/middleware"
	"cbt-smk-v2/backend/internal/rbac"
	"cbt-smk-v2/backend/internal/session"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// Panjang password minimal.
const minPasswordLength = 8

// --- rate limit login (anti brute force) -----------------------------------
// Aturan (ANALISIS Bab 7, T13): maks 5 gagal per akun per 15 menit,
// percobaan ke-6 -> 429 + kunci 15 menit. In-memory per instance.
type loginAttempt struct {
	failures   int
	firstFail  time.Time
	lockedTill time.Time
}

var (
	loginAttemptsMu sync.Mutex
	loginAttempts   = make(map[string]*loginAttempt)
)

const (
	maxLoginFailures = 5
	loginLockWindow  = 15 * time.Minute
)

func loginLocked(key string) bool {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()
	a, ok := loginAttempts[key]
	return ok && time.Now().Before(a.lockedTill)
}

func loginRecordFailure(key string) {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()
	now := time.Now()
	a, ok := loginAttempts[key]
	if !ok || now.Sub(a.firstFail) > loginLockWindow {
		a = &loginAttempt{firstFail: now}
		loginAttempts[key] = a
	}
	a.failures++
	if a.failures >= maxLoginFailures {
		a.lockedTill = now.Add(loginLockWindow)
	}
}

func loginReset(key string) {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()
	delete(loginAttempts, key)
}

// --- model ------------------------------------------------------------------

type userRow struct {
	ID                 uuid.UUID
	Nama               string
	Email              *string
	Username           *string
	PasswordHash       string
	Role               string
	MustChangePassword bool
	IsActive           bool
	CreatedAt          time.Time
}

func userPublic(u *userRow) map[string]any {
	return map[string]any{
		"id": u.ID.String(), "nama": u.Nama, "email": u.Email,
		"username": u.Username, "role": u.Role,
		"must_change_password": u.MustChangePassword, "is_active": u.IsActive,
		"created_at": u.CreatedAt,
	}
}

// getUserByLogin mencari user berdasarkan email ATAU username
// (case-insensitive, spasi dipangkas).
func getUserByLogin(ctx context.Context, pool *pgxpool.Pool, login string) (*userRow, error) {
	login = strings.TrimSpace(login)
	var u userRow
	err := pool.QueryRow(ctx, `
		SELECT id, nama, email, username, password_hash, role, must_change_password, is_active, created_at
		FROM users WHERE lower(email) = lower($1) OR lower(username) = lower($1)`,
		login).Scan(&u.ID, &u.Nama, &u.Email, &u.Username, &u.PasswordHash, &u.Role,
		&u.MustChangePassword, &u.IsActive, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func getUserByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*userRow, error) {
	var u userRow
	err := pool.QueryRow(ctx, `
		SELECT id, nama, email, username, password_hash, role, must_change_password, is_active, created_at
		FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.Nama, &u.Email, &u.Username, &u.PasswordHash, &u.Role,
			&u.MustChangePassword, &u.IsActive, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// listPermissionsForRole mengambil permission role dari DB
// (role_permissions join permissions); fallback ke matriks rbac bila
// DB belum di-seed.
func listPermissionsForRole(ctx context.Context, pool *pgxpool.Pool, role string) []string {
	rows, err := pool.Query(ctx, `
		SELECT p.code FROM role_permissions rp
		JOIN permissions p ON p.id = rp.permission_id
		JOIN roles r ON r.id = rp.role_id
		WHERE r.code = $1 ORDER BY p.code`, role)
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
	return rbac.PermissionsForRole(rbac.Role(role))
}

// --- cookie -----------------------------------------------------------------

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

// --- error helpers ----------------------------------------------------------

func writeError(c *echo.Context, status int, code, message string) error {
	return c.JSON(status, map[string]any{
		"error": map[string]any{"code": code, "message": message},
	})
}

func writeOK(c *echo.Context, data any) error {
	return c.JSON(http.StatusOK, map[string]any{"data": data})
}

// --- Register ---------------------------------------------------------------

// Register mendaftarkan endpoint auth.
//
//	func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config)
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q // auth memakai pool langsung (konvensi tim lain); kueri sqlc ada di db/queries/0001_auth.sql

	// POST /api/v1/auth/login — TIDAK butuh session (exempt RequireAuth).
	// Body: {"login": "<email|username>", "password": "..."}
	// 403 {"error":{"code":"PASSWORD_CHANGE_REQUIRED"}} bila flag aktif.
	e.POST("/api/v1/auth/login", func(c *echo.Context) error {
		ctx := c.Request().Context()
		var body struct {
			Login    string `json:"login"`
			Email    string `json:"email"`
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.Bind(&body); err != nil {
			return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
		}
		login := body.Login
		if login == "" {
			login = body.Email
		}
		if login == "" {
			login = body.Username
		}
		if strings.TrimSpace(login) == "" || body.Password == "" {
			return writeError(c, http.StatusBadRequest, "validation_error", "Email/username dan password wajib diisi.")
		}
		key := strings.ToLower(strings.TrimSpace(login))
		if loginLocked(key) {
			return writeError(c, http.StatusTooManyRequests, "login_locked", "Akun terkunci sementara karena terlalu banyak percobaan login gagal. Coba lagi dalam 15 menit.")
		}
		u, err := getUserByLogin(ctx, pool, login)
		if err != nil {
			loginRecordFailure(key)
			_ = audit.Catat(ctx, pool, "", "auth.login.gagal", "users", "",
				map[string]any{"login": login, "alasan": "akun tidak ditemukan"}, c.RealIP())
			return writeError(c, http.StatusUnauthorized, "invalid_credentials", "Email/username atau password salah.")
		}
		ok, err := VerifyPassword(body.Password, u.PasswordHash)
		if err != nil || !ok {
			loginRecordFailure(key)
			_ = audit.Catat(ctx, pool, u.ID.String(), "auth.login.gagal", "users", u.ID.String(),
				map[string]any{"alasan": "password salah"}, c.RealIP())
			return writeError(c, http.StatusUnauthorized, "invalid_credentials", "Email/username atau password salah.")
		}
		if !u.IsActive {
			return writeError(c, http.StatusForbidden, "account_inactive", "Akun tidak aktif. Hubungi admin.")
		}
		if u.MustChangePassword {
			// TIDAK dibuatkan session: pengguna harus ganti password dulu
			// lewat POST /auth/ganti-password (mode kredensial).
			return writeError(c, http.StatusForbidden, "PASSWORD_CHANGE_REQUIRED", "Password harus diganti sebelum login.")
		}
		loginReset(key)
		sessionID, csrfToken, err := session.Create(ctx, pool, u.ID, u.Role, c.RealIP(), c.Request().UserAgent())
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membuat sesi.")
		}
		setSessionCookies(c, cfg, sessionID, csrfToken)
		// UC20: catat waktu login terakhir. Kegagalan UPDATE diabaikan
		// (tidak boleh menggagalkan login yang sudah sukses).
		_, _ = pool.Exec(ctx, `UPDATE users SET terakhir_masuk = now() WHERE id = $1`, u.ID)
		_ = audit.Catat(ctx, pool, u.ID.String(), "auth.login", "users", u.ID.String(),
			map[string]any{"role": u.Role}, c.RealIP())
		return writeOK(c, map[string]any{
			"user":        userPublic(u),
			"permissions": listPermissionsForRole(ctx, pool, u.Role),
		})
	})

	// POST /api/v1/auth/logout — hapus session + clear cookie.
	e.POST("/api/v1/auth/logout", func(c *echo.Context) error {
		ctx := c.Request().Context()
		uid, _ := c.Get(middleware.CtxUserID).(string)
		if cookie, err := c.Cookie(middleware.SessionCookieName); err == nil && cookie.Value != "" {
			_ = session.Destroy(ctx, pool, cookie.Value)
		}
		clearSessionCookies(c, cfg)
		_ = audit.Catat(ctx, pool, uid, "auth.logout", "users", uid, nil, c.RealIP())
		return writeOK(c, map[string]any{"logged_out": true})
	})

	// GET /api/v1/auth/me — user + role + permissions.
	e.GET("/api/v1/auth/me", func(c *echo.Context) error {
		ctx := c.Request().Context()
		uid, _ := c.Get(middleware.CtxUserID).(string)
		id, err := uuid.Parse(uid)
		if err != nil {
			return writeError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
		}
		u, err := getUserByID(ctx, pool, id)
		if err != nil || !u.IsActive {
			return writeError(c, http.StatusUnauthorized, "unauthenticated", "Akun tidak ditemukan / tidak aktif.")
		}
		return writeOK(c, map[string]any{
			"user":        userPublic(u),
			"permissions": listPermissionsForRole(ctx, pool, u.Role),
		})
	})

	// POST /api/v1/auth/ganti-password — BOLEH diakses saat flag
	// must_change_password aktif, dalam dua mode:
	// (a) Mode session: user sudah login (cookie valid) — body cukup
	//     {"password_lama", "password_baru"}.
	// (b) Mode kredensial: tanpa session (mis. login pertama yang ditolak
	//     403 PASSWORD_CHANGE_REQUIRED) — body {"login", "password_lama",
	//     "password_baru"}. Endpoint ini exempt dari RequireAuth global
	//     justru agar mode (b) bisa jalan.
	// Setelah berhasil, SEMUA session user dihapus (wajib login ulang).
	e.POST("/api/v1/auth/ganti-password", func(c *echo.Context) error {
		ctx := c.Request().Context()
		var body struct {
			Login        string `json:"login"`
			PasswordLama string `json:"password_lama"`
			PasswordBaru string `json:"password_baru"`
		}
		if err := c.Bind(&body); err != nil {
			return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
		}
		if len(body.PasswordBaru) < minPasswordLength {
			return writeError(c, http.StatusBadRequest, "validation_error", "Password baru minimal 8 karakter.")
		}
		if body.PasswordBaru == body.PasswordLama {
			return writeError(c, http.StatusBadRequest, "validation_error", "Password baru harus berbeda dari password lama.")
		}
		var u *userRow
		// Mode (a): session aktif.
		if uidStr, _ := c.Get(middleware.CtxUserID).(string); uidStr != "" {
			id, err := uuid.Parse(uidStr)
			if err == nil {
				u, _ = getUserByID(ctx, pool, id)
			}
		}
		// Mode (b): kredensial.
		if u == nil {
			if strings.TrimSpace(body.Login) == "" {
				return writeError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu atau sertakan email/username.")
			}
			var err error
			u, err = getUserByLogin(ctx, pool, body.Login)
			if err != nil {
				return writeError(c, http.StatusUnauthorized, "invalid_credentials", "Email/username atau password lama salah.")
			}
		}
		ok, err := VerifyPassword(body.PasswordLama, u.PasswordHash)
		if err != nil || !ok {
			return writeError(c, http.StatusUnauthorized, "invalid_credentials", "Password lama salah.")
		}
		hash, err := HashPassword(body.PasswordBaru)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal meng-hash password.")
		}
		_, err = pool.Exec(ctx, `
			UPDATE users SET password_hash = $1, must_change_password = FALSE, updated_at = now()
			WHERE id = $2`, hash, u.ID)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menyimpan password baru.")
		}
		_ = session.DestroyAllForUser(ctx, pool, u.ID)
		clearSessionCookies(c, cfg)
		_ = audit.Catat(ctx, pool, u.ID.String(), "auth.ganti-password", "users", u.ID.String(), nil, c.RealIP())
		return writeOK(c, map[string]any{"password_changed": true})
	})

	// GET /api/v1/auth/csrf — kembalikan csrf_token untuk session aktif.
	// PENANDA: endpoint ini mengembalikan {"csrf_token": ...} mentah
	// (tanpa envelope {"data": ...}) sesuai spesifikasi tugas Tim 1,
	// karena nilainya juga tersedia sebagai cookie `cbt_csrf`.
	e.GET("/api/v1/auth/csrf", func(c *echo.Context) error {
		csrf, _ := c.Get(middleware.CtxCSRF).(string)
		if csrf == "" {
			return writeError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
		}
		return c.JSON(http.StatusOK, map[string]any{"csrf_token": csrf})
	})
}

// --- CLI akun admin ---------------------------------------------------------

// ErrUsersNotEmpty berarti tabel users sudah berisi akun (create-admin
// hanya diizinkan saat tabel users kosong).
var ErrUsersNotEmpty = errors.New("tabel users sudah berisi akun — create-admin hanya diizinkan saat users kosong")

// CreateAdmin membuat akun admin pertama. Hanya bila tabel users kosong.
// must_change_password=true agar admin wajib ganti password saat login pertama.
func CreateAdmin(ctx context.Context, pool *pgxpool.Pool, email, username, nama, password string) error {
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return ErrUsersNotEmpty
	}
	if strings.TrimSpace(nama) == "" {
		return errors.New("nama wajib diisi")
	}
	if len(password) < minPasswordLength {
		return errors.New("password minimal 8 karakter")
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	var emailVal, usernameVal any
	if strings.TrimSpace(email) != "" {
		emailVal = strings.TrimSpace(email)
	}
	if strings.TrimSpace(username) != "" {
		usernameVal = strings.TrimSpace(username)
	}
	if emailVal == nil && usernameVal == nil {
		return errors.New("email atau username wajib diisi")
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO users (nama, email, username, password_hash, role, must_change_password, is_active)
		VALUES ($1, $2, $3, $4, 'admin', TRUE, TRUE)`,
		strings.TrimSpace(nama), emailVal, usernameVal, hash)
	return err
}

// ResetPassword mengganti password akun (berdasarkan email/username),
// mengaktifkan akun, menandai must_change_password=true, dan menghapus
// semua session akun tersebut.
func ResetPassword(ctx context.Context, pool *pgxpool.Pool, login, newPassword string) error {
	if len(newPassword) < minPasswordLength {
		return errors.New("password minimal 8 karakter")
	}
	u, err := getUserByLogin(ctx, pool, login)
	if err != nil {
		return errors.New("akun tidak ditemukan")
	}
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, `
		UPDATE users SET password_hash = $1, must_change_password = TRUE, is_active = TRUE, updated_at = now()
		WHERE id = $2`, hash, u.ID)
	if err != nil {
		return err
	}
	return session.DestroyAllForUser(ctx, pool, u.ID)
}
