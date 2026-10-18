// Package db — placeholder Queries untuk sqlc (Tim 5).
//
// Kode hasil `sqlc generate` (dari backend/db/queries/*.sql) akan
// menggantikan/memperkaya berkas ini. Sampai sqlc dijalankan, handler
// Tim 5 memakai pgxpool langsung dengan SQL yang sama seperti di
// berkas kueri, dan hanya memakai tipe Queries ini agar signature
// Register(...) sesuai kontrak bersama.
package db

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// Queries adalah pegangan kueri sqlc (placeholder sampai generate).
type Queries struct {
	Pool *pgxpool.Pool
}

// New membuat Queries dari pool.
func New(pool *pgxpool.Pool) *Queries {
	return &Queries{Pool: pool}
}
