// Package users — Manajemen Pengguna (akun STAF) — UC20.
//
// Endpoint (prefix /api/v1/pengguna):
//
//	GET    /api/v1/pengguna                 (pengguna.read)
//	POST   /api/v1/pengguna                 (pengguna.create) -> 201
//	GET    /api/v1/pengguna/:id             (pengguna.read)
//	PUT    /api/v1/pengguna/:id             (pengguna.update)
//	POST   /api/v1/pengguna/:id/reset-password (pengguna.reset-password)
//	POST   /api/v1/pengguna/:id/toggle-aktif   (pengguna.toggle-aktif)
//
// Permission diperiksa per-handler via audit.HasPermission (pola paket
// master); CSRF + auth sesi ditangani middleware global. Hanya akun STAF
// yang dikelola di sini (role != 'siswa'); akun siswa tetap dikelola
// lewat halaman Siswa (paket master). Akun siswa yang ditarget lewat
// endpoint ini mendapat 404 (anti-IDOR, sama seperti pola siswa).
//
// Aturan yang ditegakkan:
//   - Username wajib + unik (cek case-insensitive; duplikat -> 409).
//   - Email opsional; bila diisi harus valid dan unik (duplikat -> 409).
//   - Role wajib salah satu dari 5 role staf; role 'siswa' -> 400.
//   - Password minimal 8 karakter (sama seperti kebijakan auth).
//   - Password disimpan Argon2id via auth.HashPassword;
//     akun baru / reset selalu must_change_password = TRUE.
//   - Username TIDAK bisa diubah (PUT mengabaikan username).
//   - Guard admin terakhir: demosi admin -> non-admin dan penonaktifan
//     admin DITOLAK 409 bila tidak ada admin aktif lain yang tersisa.
//   - Menonaktifkan akun sendiri DITOLAK 403.
//   - TIDAK ADA hard delete: menonaktifkan saja.
//   - Reset password menghapus SEMUA sesi user itu agar sesi lama
//     tidak bertahan.
//   - Semua mutasi dicatat ke audit_log (aksi pengguna.*).
//   - Envelope: sukses {"data":...} / error {"error":{"code","message"}}
//     (Bahasa Indonesia), sama seperti paket lain.
package users

import (
	"context"
	"errors"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/auth"
	"cbt-smk-v2/backend/internal/config"
	"cbt-smk-v2/backend/internal/session"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// Panjang password minimal — sama seperti kebijakan auth (auth.go).
const minPasswordLength = 8

// StaffRoles adalah 5 role staf yang boleh dikelola lewat endpoint ini.
// Role 'siswa' sengaja TIDAK ada: akun siswa dikelola lewat Siswa.
func StaffRoles() []string {
	return []string{"admin", "proktor", "guru", "pengawas", "penguji_ukk"}
}

// IsStaffRole melaporkan apakah role termasuk role staf.
func IsStaffRole(role string) bool {
	for _, r := range StaffRoles() {
		if r == role {
			return true
		}
	}
	return false
}

// --- validasi (fungsi murni, diuji tanpa DB) -------------------------------

// ValidasiRoleStaf mengembalikan pesan error bila role bukan role staf
// ("" = valid). Role 'siswa' ditolak dengan pesan yang menjelaskan
// bahwa akun siswa dikelola di halaman Siswa.
func ValidasiRoleStaf(role string) string {
	if strings.TrimSpace(role) == "" {
		return "Role wajib diisi."
	}
	if role == "siswa" {
		return "Role siswa tidak dikelola di sini; kelola akun siswa lewat halaman Siswa."
	}
	if !IsStaffRole(role) {
		return "Role harus salah satu dari: admin, proktor, guru, pengawas, penguji_ukk."
	}
	return ""
}

// ValidasiUsername mengembalikan pesan error bila username tidak valid.
func ValidasiUsername(username string) string {
	username = strings.TrimSpace(username)
	if username == "" {
		return "Username wajib diisi."
	}
	if strings.ContainsAny(username, " \t\r\n") {
		return "Username tidak boleh mengandung spasi."
	}
	return ""
}

// ValidasiEmail mengembalikan pesan error bila email tidak valid.
// String kosong = valid (email opsional; disimpan NULL).
func ValidasiEmail(email string) string {
	email = strings.TrimSpace(email)
	if email == "" {
		return ""
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return "Format email tidak valid."
	}
	// mail.ParseAddress menerima "a@b" (tanpa titik di domain) — tolak,
	// karena alamat seperti itu bukan email yang bisa dipakai sekolah.
	domain := addr.Address[strings.LastIndex(addr.Address, "@")+1:]
	if !strings.Contains(domain, ".") {
		return "Format email tidak valid."
	}
	return ""
}

// ValidasiPassword mengembalikan pesan error bila password terlalu pendek.
func ValidasiPassword(password string) string {
	if len(password) < minPasswordLength {
		return "Password minimal 8 karakter."
	}
	return ""
}

// --- DTO --------------------------------------------------------------------

type penggunaDTO struct {
	ID                 string     `json:"id"`
	Username           *string    `json:"username"`
	Nama               string     `json:"nama"`
	Email              *string    `json:"email"`
	Role               string     `json:"role"`
	IsActive           bool       `json:"is_active"`
	MustChangePassword bool       `json:"must_change_password"`
	TerakhirMasuk      *time.Time `json:"terakhir_masuk"`
	CreatedAt          time.Time  `json:"created_at"`
}

// --- helper umum (pola paket master) ---------------------------------------

func writeError(c *echo.Context, status int, code, message string) error {
	return c.JSON(status, map[string]any{
		"error": map[string]any{"code": code, "message": message},
	})
}

func writeOK(c *echo.Context, data any) error {
	return c.JSON(http.StatusOK, map[string]any{"data": data})
}

func writeOKStatus(c *echo.Context, status int, data any) error {
	return c.JSON(status, map[string]any{"data": data})
}

type pageMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func writeList(c *echo.Context, items any, page, perPage int, total int64) error {
	totalPages := 0
	if perPage > 0 {
		totalPages = int((total + int64(perPage) - 1) / int64(perPage))
	}
	return c.JSON(http.StatusOK, map[string]any{
		"data": items,
		"meta": pageMeta{Page: page, PerPage: perPage, Total: total, TotalPages: totalPages},
	})
}

func parsePagination(c *echo.Context) (page, perPage, offset int) {
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page < 1 {
		page = 1
	}
	perPage, err = strconv.Atoi(c.QueryParam("per_page"))
	if err != nil || perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	return page, perPage, (page - 1) * perPage
}

func validUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

func uniqueConstraint(err error) (string, bool) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return pgErr.ConstraintName, true
	}
	return "", false
}

func catat(c *echo.Context, pool *pgxpool.Pool, aksi, entitas, entitasID string, detail any) {
	userID, _, _, _, _ := audit.CurrentUser(c)
	_ = audit.Catat(c.Request().Context(), pool, userID, aksi, entitas, entitasID, detail, c.RealIP())
}

// --- akses data -------------------------------------------------------------

// getStaf memuat satu akun STAF. Akun siswa / tidak ada / UUID tidak
// valid -> (nil, ErrNoRows) sehingga handler mengembalikan 404.
func getStaf(ctx context.Context, pool *pgxpool.Pool, id string) (*penggunaDTO, error) {
	var d penggunaDTO
	err := pool.QueryRow(ctx,
		`SELECT id, username, nama, email, role, is_active, must_change_password, terakhir_masuk, created_at
		 FROM users WHERE id = $1 AND role <> 'siswa'`, id).
		Scan(&d.ID, &d.Username, &d.Nama, &d.Email, &d.Role,
			&d.IsActive, &d.MustChangePassword, &d.TerakhirMasuk, &d.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// countActiveAdminsExcept menghitung admin AKTIF selain id yang dikecualikan.
func countActiveAdminsExcept(ctx context.Context, pool *pgxpool.Pool, excludeID string) (int64, error) {
	var n int64
	err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM users WHERE role = 'admin' AND is_active = TRUE AND id <> $1`, excludeID).Scan(&n)
	return n, err
}

// --- Register ---------------------------------------------------------------

// Register mendaftarkan endpoint Manajemen Pengguna (staf).
//
//	func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config)
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q
	_ = cfg

	// GET /api/v1/pengguna — list akun STAF saja.
	// Filter: ?q= (nama/username/email), ?role= (salah satu role staf),
	// ?is_active= (true/false). Pagination ?page=&per_page= (pola master).
	e.GET("/api/v1/pengguna", func(c *echo.Context) error {
		if !audit.HasPermission(c, "pengguna.read") {
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat pengguna.")
		}
		page, perPage, offset := parsePagination(c)
		ctx := c.Request().Context()

		q := strings.TrimSpace(c.QueryParam("q"))
		var qArg any
		if q != "" {
			qArg = q
		}
		role := strings.TrimSpace(c.QueryParam("role"))
		if role != "" {
			// Kontrak: filter role 'siswa' -> 400 (akun siswa di halaman Siswa).
			if role == "siswa" {
				return writeError(c, http.StatusBadRequest, "validation_error", "Role siswa tidak dikelola di sini; kelola akun siswa lewat halaman Siswa.")
			}
			if msg := ValidasiRoleStaf(role); msg != "" {
				return writeError(c, http.StatusUnprocessableEntity, "validation_error", msg)
			}
		}
		var roleArg any
		if role != "" {
			roleArg = role
		}
		var isActiveArg any
		if v := strings.TrimSpace(c.QueryParam("is_active")); v != "" {
			if b, err := strconv.ParseBool(v); err == nil {
				isActiveArg = b
			}
		}

		var total int64
		if err := pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM users
			 WHERE role <> 'siswa'
			   AND ($1::text IS NULL OR nama ILIKE '%' || $1 || '%' OR COALESCE(username, '') ILIKE '%' || $1 || '%' OR COALESCE(email, '') ILIKE '%' || $1 || '%')
			   AND ($2::text IS NULL OR role = $2)
			   AND ($3::boolean IS NULL OR is_active = $3)`,
			qArg, roleArg, isActiveArg).Scan(&total); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca pengguna.")
		}
		rows, err := pool.Query(ctx,
			`SELECT id, username, nama, email, role, is_active, must_change_password, terakhir_masuk, created_at
			 FROM users
			 WHERE role <> 'siswa'
			   AND ($1::text IS NULL OR nama ILIKE '%' || $1 || '%' OR COALESCE(username, '') ILIKE '%' || $1 || '%' OR COALESCE(email, '') ILIKE '%' || $1 || '%')
			   AND ($2::text IS NULL OR role = $2)
			   AND ($3::boolean IS NULL OR is_active = $3)
			 ORDER BY nama ASC LIMIT $4 OFFSET $5`,
			qArg, roleArg, isActiveArg, perPage, offset)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca pengguna.")
		}
		defer rows.Close()
		items := []penggunaDTO{}
		for rows.Next() {
			var d penggunaDTO
			if err := rows.Scan(&d.ID, &d.Username, &d.Nama, &d.Email, &d.Role,
				&d.IsActive, &d.MustChangePassword, &d.TerakhirMasuk, &d.CreatedAt); err != nil {
				return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca pengguna.")
			}
			items = append(items, d)
		}
		if err := rows.Err(); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca pengguna.")
		}
		return writeList(c, items, page, perPage, total)
	})

	// POST /api/v1/pengguna — buat akun staf baru. -> 201
	e.POST("/api/v1/pengguna", func(c *echo.Context) error {
		if !audit.HasPermission(c, "pengguna.create") {
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menambah pengguna.")
		}
		var in struct {
			Username string `json:"username"`
			Nama     string `json:"nama"`
			Email    string `json:"email"`
			Role     string `json:"role"`
			Password string `json:"password"`
		}
		if err := c.Bind(&in); err != nil {
			return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
		}
		if msg := ValidasiUsername(in.Username); msg != "" {
			return writeError(c, http.StatusUnprocessableEntity, "validation_error", msg)
		}
		if strings.TrimSpace(in.Nama) == "" {
			return writeError(c, http.StatusUnprocessableEntity, "validation_error", "Nama wajib diisi.")
		}
		if msg := ValidasiEmail(in.Email); msg != "" {
			return writeError(c, http.StatusUnprocessableEntity, "validation_error", msg)
		}
		// Kontrak: role 'siswa' -> 400 (akun siswa di halaman Siswa).
		if strings.TrimSpace(in.Role) == "siswa" {
			return writeError(c, http.StatusBadRequest, "validation_error", "Role siswa tidak dikelola di sini; kelola akun siswa lewat halaman Siswa.")
		}
		if msg := ValidasiRoleStaf(in.Role); msg != "" {
			return writeError(c, http.StatusUnprocessableEntity, "validation_error", msg)
		}
		if msg := ValidasiPassword(in.Password); msg != "" {
			return writeError(c, http.StatusUnprocessableEntity, "validation_error", msg)
		}
		ctx := c.Request().Context()
		username := strings.TrimSpace(in.Username)
		email := strings.TrimSpace(in.Email)

		// Duplikat case-insensitive (UNIQUE DB case-sensitive, sedangkan
		// login case-insensitive — "Admin" vs "admin" tidak boleh ganda).
		var dupeUsername bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE lower(username) = lower($1))`, username).Scan(&dupeUsername); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa username.")
		}
		if dupeUsername {
			return writeError(c, http.StatusConflict, "conflict", "Username sudah dipakai.")
		}
		if email != "" {
			var dupeEmail bool
			if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE lower(email) = lower($1))`, email).Scan(&dupeEmail); err != nil {
				return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa email.")
			}
			if dupeEmail {
				return writeError(c, http.StatusConflict, "conflict", "Email sudah dipakai.")
			}
		}

		hash, err := auth.HashPassword(in.Password)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal meng-hash password.")
		}
		var emailArg any
		if email != "" {
			emailArg = email
		}
		var d penggunaDTO
		err = pool.QueryRow(ctx,
			`INSERT INTO users (username, nama, email, role, password_hash, must_change_password, is_active)
			 VALUES ($1, $2, $3, $4, $5, TRUE, TRUE)
			 RETURNING id, username, nama, email, role, is_active, must_change_password, terakhir_masuk, created_at`,
			username, strings.TrimSpace(in.Nama), emailArg, in.Role, hash).
			Scan(&d.ID, &d.Username, &d.Nama, &d.Email, &d.Role,
				&d.IsActive, &d.MustChangePassword, &d.TerakhirMasuk, &d.CreatedAt)
		if err != nil {
			if constraint, ok := uniqueConstraint(err); ok {
				switch constraint {
				case "users_username_key":
					return writeError(c, http.StatusConflict, "conflict", "Username sudah dipakai.")
				case "users_email_key":
					return writeError(c, http.StatusConflict, "conflict", "Email sudah dipakai.")
				}
				return writeError(c, http.StatusConflict, "conflict", "Data duplikat: nilai unik sudah dipakai.")
			}
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menambah pengguna.")
		}
		catat(c, pool, "pengguna.create", "users", d.ID,
			map[string]any{"username": username, "nama": d.Nama, "email": d.Email, "role": d.Role})
		return writeOKStatus(c, http.StatusCreated, d)
	})

	// GET /api/v1/pengguna/:id — detail akun staf (siswa -> 404).
	e.GET("/api/v1/pengguna/:id", func(c *echo.Context) error {
		if !audit.HasPermission(c, "pengguna.read") {
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat pengguna.")
		}
		id := c.Param("id")
		if !validUUID(id) {
			return writeError(c, http.StatusNotFound, "not_found", "Pengguna tidak ditemukan.")
		}
		d, err := getStaf(c.Request().Context(), pool, id)
		if errors.Is(err, pgx.ErrNoRows) {
			return writeError(c, http.StatusNotFound, "not_found", "Pengguna tidak ditemukan.")
		}
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca pengguna.")
		}
		return writeOK(c, d)
	})

	// PUT /api/v1/pengguna/:id — ubah nama/email/role akun staf.
	// Username TIDAK bisa diubah (diabaikan bila dikirim). Field yang
	// tidak dikirim dipertahankan. Role baru tidak boleh 'siswa'.
	e.PUT("/api/v1/pengguna/:id", func(c *echo.Context) error {
		if !audit.HasPermission(c, "pengguna.update") {
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mengubah pengguna.")
		}
		id := c.Param("id")
		if !validUUID(id) {
			return writeError(c, http.StatusNotFound, "not_found", "Pengguna tidak ditemukan.")
		}
		var in struct {
			Nama  *string `json:"nama"`
			Email *string `json:"email"`
			Role  *string `json:"role"`
		}
		if err := c.Bind(&in); err != nil {
			return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
		}
		ctx := c.Request().Context()
		cur, err := getStaf(ctx, pool, id)
		if errors.Is(err, pgx.ErrNoRows) {
			return writeError(c, http.StatusNotFound, "not_found", "Pengguna tidak ditemukan.")
		}
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca pengguna.")
		}

		nama := cur.Nama
		if in.Nama != nil {
			nama = strings.TrimSpace(*in.Nama)
			if nama == "" {
				return writeError(c, http.StatusUnprocessableEntity, "validation_error", "Nama wajib diisi.")
			}
		}
		var emailArg any = cur.Email
		if in.Email != nil {
			email := strings.TrimSpace(*in.Email)
			if msg := ValidasiEmail(email); msg != "" {
				return writeError(c, http.StatusUnprocessableEntity, "validation_error", msg)
			}
			if email == "" {
				emailArg = nil
			} else {
				var dupe bool
				if err := pool.QueryRow(ctx,
					`SELECT EXISTS(SELECT 1 FROM users WHERE lower(email) = lower($1) AND id <> $2)`, email, id).Scan(&dupe); err != nil {
					return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa email.")
				}
				if dupe {
					return writeError(c, http.StatusConflict, "conflict", "Email sudah dipakai.")
				}
				emailArg = email
			}
		}
		role := cur.Role
		if in.Role != nil {
			// Kontrak: role 'siswa' -> 400 (akun siswa di halaman Siswa).
			if strings.TrimSpace(*in.Role) == "siswa" {
				return writeError(c, http.StatusBadRequest, "validation_error", "Role siswa tidak dikelola di sini; kelola akun siswa lewat halaman Siswa.")
			}
			if msg := ValidasiRoleStaf(*in.Role); msg != "" {
				return writeError(c, http.StatusUnprocessableEntity, "validation_error", msg)
			}
			role = *in.Role
		}
		// Guard admin terakhir: demosi admin aktif -> non-admin ditolak
		// bila tidak ada admin aktif lain yang tersisa.
		if cur.Role == "admin" && role != "admin" && cur.IsActive {
			n, err := countActiveAdminsExcept(ctx, pool, id)
			if err != nil {
				return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa admin aktif.")
			}
			if n == 0 {
				return writeError(c, http.StatusConflict, "admin_terakhir", "Tidak dapat mengubah role admin terakhir yang masih aktif. Buat admin lain dulu.")
			}
		}

		var d penggunaDTO
		err = pool.QueryRow(ctx,
			`UPDATE users SET nama = $2, email = $3, role = $4, updated_at = now()
			 WHERE id = $1
			 RETURNING id, username, nama, email, role, is_active, must_change_password, terakhir_masuk, created_at`,
			id, nama, emailArg, role).
			Scan(&d.ID, &d.Username, &d.Nama, &d.Email, &d.Role,
				&d.IsActive, &d.MustChangePassword, &d.TerakhirMasuk, &d.CreatedAt)
		if err != nil {
			if constraint, ok := uniqueConstraint(err); ok && constraint == "users_email_key" {
				return writeError(c, http.StatusConflict, "conflict", "Email sudah dipakai.")
			}
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mengubah pengguna.")
		}
		catat(c, pool, "pengguna.update", "users", d.ID,
			map[string]any{"nama": d.Nama, "email": d.Email, "role": d.Role})
		return writeOK(c, d)
	})

	// POST /api/v1/pengguna/:id/reset-password — reset password akun staf.
	// Hash baru + must_change_password=TRUE + hapus SEMUA sesi user itu.
	e.POST("/api/v1/pengguna/:id/reset-password", func(c *echo.Context) error {
		if !audit.HasPermission(c, "pengguna.reset-password") {
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mereset password pengguna.")
		}
		id := c.Param("id")
		if !validUUID(id) {
			return writeError(c, http.StatusNotFound, "not_found", "Pengguna tidak ditemukan.")
		}
		var in struct {
			PasswordBaru string `json:"password_baru"`
		}
		if err := c.Bind(&in); err != nil {
			return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
		}
		if msg := ValidasiPassword(in.PasswordBaru); msg != "" {
			return writeError(c, http.StatusUnprocessableEntity, "validation_error", msg)
		}
		ctx := c.Request().Context()
		cur, err := getStaf(ctx, pool, id)
		if errors.Is(err, pgx.ErrNoRows) {
			return writeError(c, http.StatusNotFound, "not_found", "Pengguna tidak ditemukan.")
		}
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca pengguna.")
		}
		hash, err := auth.HashPassword(in.PasswordBaru)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal meng-hash password.")
		}
		_, err = pool.Exec(ctx,
			`UPDATE users SET password_hash = $1, must_change_password = TRUE, updated_at = now() WHERE id = $2`,
			hash, id)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mereset password.")
		}
		targetID, err := uuid.Parse(id)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mereset password.")
		}
		if err := session.DestroyAllForUser(ctx, pool, targetID); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghapus sesi pengguna.")
		}
		catat(c, pool, "pengguna.reset-password", "users", id,
			map[string]any{"username": cur.Username})
		return writeOK(c, map[string]any{"password_reset": true, "must_change_password": true, "id": id})
	})

	// POST /api/v1/pengguna/:id/toggle-aktif — aktif/nonaktifkan akun staf.
	// DILARANG: menonaktifkan akun sendiri (403) dan menonaktifkan admin
	// terakhir yang masih aktif (409). TIDAK ADA hard delete.
	e.POST("/api/v1/pengguna/:id/toggle-aktif", func(c *echo.Context) error {
		if !audit.HasPermission(c, "pengguna.toggle-aktif") {
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mengubah status aktif pengguna.")
		}
		id := c.Param("id")
		if !validUUID(id) {
			return writeError(c, http.StatusNotFound, "not_found", "Pengguna tidak ditemukan.")
		}
		var in struct {
			IsActive *bool `json:"is_active"`
		}
		if err := c.Bind(&in); err != nil {
			return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
		}
		if in.IsActive == nil {
			return writeError(c, http.StatusUnprocessableEntity, "validation_error", "Field is_active wajib diisi (true/false).")
		}
		ctx := c.Request().Context()
		cur, err := getStaf(ctx, pool, id)
		if errors.Is(err, pgx.ErrNoRows) {
			return writeError(c, http.StatusNotFound, "not_found", "Pengguna tidak ditemukan.")
		}
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca pengguna.")
		}
		// Guard: nonaktifkan akun sendiri — dicek dulu agar pesannya
		// jelas walau actor juga admin terakhir.
		actorID, _, _, _, _ := audit.CurrentUser(c)
		if !*in.IsActive && id == actorID {
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak dapat menonaktifkan akun sendiri.")
		}
		// Guard admin terakhir: hanya untuk transisi aktif -> nonaktif.
		if !*in.IsActive && cur.Role == "admin" && cur.IsActive {
			n, err := countActiveAdminsExcept(ctx, pool, id)
			if err != nil {
				return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa admin aktif.")
			}
			if n == 0 {
				return writeError(c, http.StatusConflict, "admin_terakhir", "Tidak dapat menonaktifkan admin terakhir yang masih aktif. Buat admin lain dulu.")
			}
		}
		var d penggunaDTO
		err = pool.QueryRow(ctx,
			`UPDATE users SET is_active = $2, updated_at = now()
			 WHERE id = $1
			 RETURNING id, username, nama, email, role, is_active, must_change_password, terakhir_masuk, created_at`,
			id, *in.IsActive).
			Scan(&d.ID, &d.Username, &d.Nama, &d.Email, &d.Role,
				&d.IsActive, &d.MustChangePassword, &d.TerakhirMasuk, &d.CreatedAt)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mengubah status pengguna.")
		}
		catat(c, pool, "pengguna.toggle-aktif", "users", d.ID,
			map[string]any{"username": d.Username, "is_active": d.IsActive})
		return writeOK(c, d)
	})
}
