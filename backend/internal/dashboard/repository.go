package dashboard

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Repository -------------------------------------------------------------

// Repository adalah lapisan akses data paket dashboard.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository membuat Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Count menghitung jumlah baris satu tabel agregat dashboard.
// Nama tabel berasal dari konstanta pemanggil, bukan input user.
// (Dulu: closure hitung inline di handler.)
func (r *Repository) Count(ctx context.Context, tabel string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM `+tabel).Scan(&n)
	return n, err
}
