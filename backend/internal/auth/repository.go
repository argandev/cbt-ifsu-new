package auth

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"cbt-smk-v2/backend/internal/rbac"
)

// --- model ------------------------------------------------------------------
// (Dipindah dari auth.go tanpa perubahan.)

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

// --- Repository -------------------------------------------------------------

// Repository adalah lapisan akses data paket auth (semua SQL ada di sini).
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository membuat Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// ByLogin mencari user berdasarkan email ATAU username
// (case-insensitive, spasi dipangkas). (Dulu: getUserByLogin.)
func (r *Repository) ByLogin(ctx context.Context, login string) (*userRow, error) {
	login = strings.TrimSpace(login)
	var u userRow
	err := r.pool.QueryRow(ctx, `
		SELECT id, nama, email, username, password_hash, role, must_change_password, is_active, created_at
		FROM users WHERE lower(email) = lower($1) OR lower(username) = lower($1)`,
		login).Scan(&u.ID, &u.Nama, &u.Email, &u.Username, &u.PasswordHash, &u.Role,
		&u.MustChangePassword, &u.IsActive, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ByID mencari user berdasarkan id. (Dulu: getUserByID.)
func (r *Repository) ByID(ctx context.Context, id uuid.UUID) (*userRow, error) {
	var u userRow
	err := r.pool.QueryRow(ctx, `
		SELECT id, nama, email, username, password_hash, role, must_change_password, is_active, created_at
		FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.Nama, &u.Email, &u.Username, &u.PasswordHash, &u.Role,
			&u.MustChangePassword, &u.IsActive, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// PermissionsForRole mengambil permission role dari DB
// (role_permissions join permissions); fallback ke matriks rbac bila
// DB belum di-seed. (Dulu: listPermissionsForRole.)
func (r *Repository) PermissionsForRole(ctx context.Context, role string) []string {
	rows, err := r.pool.Query(ctx, `
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

// CountUsers menghitung jumlah baris tabel users.
func (r *Repository) CountUsers(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&count)
	return count, err
}

// InsertAdmin menyisipkan akun admin pertama (must_change_password=TRUE).
func (r *Repository) InsertAdmin(ctx context.Context, nama string, emailVal, usernameVal any, hash string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users (nama, email, username, password_hash, role, must_change_password, is_active)
		VALUES ($1, $2, $3, $4, 'admin', TRUE, TRUE)`,
		strings.TrimSpace(nama), emailVal, usernameVal, hash)
	return err
}

// UpdatePasswordGanti menyimpan password baru hasil ganti-password
// (must_change_password=FALSE).
func (r *Repository) UpdatePasswordGanti(ctx context.Context, id uuid.UUID, hash string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET password_hash = $1, must_change_password = FALSE, updated_at = now()
		WHERE id = $2`, hash, id)
	return err
}

// UpdatePasswordReset menyimpan password baru hasil reset CLI
// (must_change_password=TRUE, akun diaktifkan).
func (r *Repository) UpdatePasswordReset(ctx context.Context, id uuid.UUID, hash string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET password_hash = $1, must_change_password = TRUE, is_active = TRUE, updated_at = now()
		WHERE id = $2`, hash, id)
	return err
}

// UpdateTerakhirMasuk mencatat waktu login terakhir (UC20).
// Kegagalan diabaikan pemanggil (tidak boleh menggagalkan login).
func (r *Repository) UpdateTerakhirMasuk(ctx context.Context, id uuid.UUID) {
	_, _ = r.pool.Exec(ctx, `UPDATE users SET terakhir_masuk = now() WHERE id = $1`, id)
}
