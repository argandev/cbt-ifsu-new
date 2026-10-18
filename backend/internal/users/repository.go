package users

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- DTO --------------------------------------------------------------------

// penggunaDTO adalah bentuk data akun staf yang dikembalikan API.
// (Dipindah dari users.go tanpa perubahan.)
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

// StafFilter adalah filter daftar akun staf. Nilai nil = filter tidak dipakai
// (dipetakan ke NULL pada parameter SQL, sama seperti pola any sebelumnya).
type StafFilter struct {
	Q        *string
	Role     *string
	IsActive *bool
}

// --- Repository -------------------------------------------------------------

// Repository adalah lapisan akses data paket users (semua SQL ada di sini).
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository membuat Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// uniqueConstraint melaporkan nama constraint bila err adalah pelanggaran
// unique PostgreSQL (23505). (Dipindah dari users.go tanpa perubahan.)
func uniqueConstraint(err error) (string, bool) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return pgErr.ConstraintName, true
	}
	return "", false
}

// GetStaf memuat satu akun STAF. Akun siswa / tidak ada -> pgx.ErrNoRows
// sehingga pemanggil mengembalikan 404. (Dulu: getStaf.)
func (r *Repository) GetStaf(ctx context.Context, id string) (*penggunaDTO, error) {
	var d penggunaDTO
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, nama, email, role, is_active, must_change_password, terakhir_masuk, created_at
		 FROM users WHERE id = $1 AND role <> 'siswa'`, id).
		Scan(&d.ID, &d.Username, &d.Nama, &d.Email, &d.Role,
			&d.IsActive, &d.MustChangePassword, &d.TerakhirMasuk, &d.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// CountActiveAdminsExcept menghitung admin AKTIF selain id yang dikecualikan.
func (r *Repository) CountActiveAdminsExcept(ctx context.Context, excludeID string) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM users WHERE role = 'admin' AND is_active = TRUE AND id <> $1`, excludeID).Scan(&n)
	return n, err
}

func nilOr[T any](p *T) any {
	if p == nil {
		return nil
	}
	return *p
}

// CountStaf menghitung total akun staf sesuai filter.
func (r *Repository) CountStaf(ctx context.Context, f StafFilter) (int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM users
		 WHERE role <> 'siswa'
		   AND ($1::text IS NULL OR nama ILIKE '%' || $1 || '%' OR COALESCE(username, '') ILIKE '%' || $1 || '%' OR COALESCE(email, '') ILIKE '%' || $1 || '%')
		   AND ($2::text IS NULL OR role = $2)
		   AND ($3::boolean IS NULL OR is_active = $3)`,
		nilOr(f.Q), nilOr(f.Role), nilOr(f.IsActive)).Scan(&total)
	return total, err
}

// ListStaf memuat halaman akun staf sesuai filter, terurut nama.
func (r *Repository) ListStaf(ctx context.Context, f StafFilter, limit, offset int) ([]penggunaDTO, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, username, nama, email, role, is_active, must_change_password, terakhir_masuk, created_at
		 FROM users
		 WHERE role <> 'siswa'
		   AND ($1::text IS NULL OR nama ILIKE '%' || $1 || '%' OR COALESCE(username, '') ILIKE '%' || $1 || '%' OR COALESCE(email, '') ILIKE '%' || $1 || '%')
		   AND ($2::text IS NULL OR role = $2)
		   AND ($3::boolean IS NULL OR is_active = $3)
		 ORDER BY nama ASC LIMIT $4 OFFSET $5`,
		nilOr(f.Q), nilOr(f.Role), nilOr(f.IsActive), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []penggunaDTO{}
	for rows.Next() {
		var d penggunaDTO
		if err := rows.Scan(&d.ID, &d.Username, &d.Nama, &d.Email, &d.Role,
			&d.IsActive, &d.MustChangePassword, &d.TerakhirMasuk, &d.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// UsernameExists melaporkan apakah username (case-insensitive) sudah dipakai.
func (r *Repository) UsernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE lower(username) = lower($1))`, username).Scan(&exists)
	return exists, err
}

// EmailExists melaporkan apakah email (case-insensitive) sudah dipakai.
func (r *Repository) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE lower(email) = lower($1))`, email).Scan(&exists)
	return exists, err
}

// EmailExistsExcept seperti EmailExists tetapi mengecualikan satu id.
func (r *Repository) EmailExistsExcept(ctx context.Context, email, excludeID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE lower(email) = lower($1) AND id <> $2)`, email, excludeID).Scan(&exists)
	return exists, err
}

// CreateStaf menyisipkan akun staf baru (must_change_password=TRUE).
// Error duplikat dikembalikan apa adanya agar pemanggil bisa memetakan
// nama constraint via uniqueConstraint.
func (r *Repository) CreateStaf(ctx context.Context, username, nama string, emailArg any, role, hash string) (*penggunaDTO, error) {
	var d penggunaDTO
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (username, nama, email, role, password_hash, must_change_password, is_active)
		 VALUES ($1, $2, $3, $4, $5, TRUE, TRUE)
		 RETURNING id, username, nama, email, role, is_active, must_change_password, terakhir_masuk, created_at`,
		username, nama, emailArg, role, hash).
		Scan(&d.ID, &d.Username, &d.Nama, &d.Email, &d.Role,
			&d.IsActive, &d.MustChangePassword, &d.TerakhirMasuk, &d.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// UpdateStaf mengubah nama/email/role akun staf.
func (r *Repository) UpdateStaf(ctx context.Context, id, nama string, emailArg any, role string) (*penggunaDTO, error) {
	var d penggunaDTO
	err := r.pool.QueryRow(ctx,
		`UPDATE users SET nama = $2, email = $3, role = $4, updated_at = now()
		 WHERE id = $1
		 RETURNING id, username, nama, email, role, is_active, must_change_password, terakhir_masuk, created_at`,
		id, nama, emailArg, role).
		Scan(&d.ID, &d.Username, &d.Nama, &d.Email, &d.Role,
			&d.IsActive, &d.MustChangePassword, &d.TerakhirMasuk, &d.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// UpdatePassword mengganti hash password + menandai wajib ganti password.
func (r *Repository) UpdatePassword(ctx context.Context, id, hash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET password_hash = $1, must_change_password = TRUE, updated_at = now() WHERE id = $2`,
		hash, id)
	return err
}

// SetAktif mengubah status aktif akun staf.
func (r *Repository) SetAktif(ctx context.Context, id string, isActive bool) (*penggunaDTO, error) {
	var d penggunaDTO
	err := r.pool.QueryRow(ctx,
		`UPDATE users SET is_active = $2, updated_at = now()
		 WHERE id = $1
		 RETURNING id, username, nama, email, role, is_active, must_change_password, terakhir_masuk, created_at`,
		id, isActive).
		Scan(&d.ID, &d.Username, &d.Nama, &d.Email, &d.Role,
			&d.IsActive, &d.MustChangePassword, &d.TerakhirMasuk, &d.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}
