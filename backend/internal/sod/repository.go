package sod

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Repository -------------------------------------------------------------

// Repository adalah lapisan akses data paket sod (semua SQL ada di sini).
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository membuat Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// InsertOverride mencatat satu pengecualian SoD, mengembalikan id-nya.
func (r *Repository) InsertOverride(ctx context.Context, actorID, aksi, entitas, entitasID, alasan string) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO sod_overrides (actor_id, aksi, entitas, entitas_id, alasan)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		actorID, aksi, entitas, entitasID, alasan).Scan(&id)
	return id, err
}

// NamaActor mengambil nama user; string kosong bila tidak ada/gagal.
func (r *Repository) NamaActor(ctx context.Context, userID string) string {
	var nama string
	_ = r.pool.QueryRow(ctx, `SELECT COALESCE(nama, '') FROM users WHERE id = $1`, userID).Scan(&nama)
	return nama
}

// AdminLainAktif mengembalikan id semua user role admin yang aktif
// kecuali actor.
func (r *Repository) AdminLainAktif(ctx context.Context, actorID string) []string {
	rows, err := r.pool.Query(ctx,
		`SELECT id FROM users WHERE role = 'admin' AND id <> $1 AND is_active = TRUE`, actorID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var adminID string
		if err := rows.Scan(&adminID); err != nil {
			continue
		}
		ids = append(ids, adminID)
	}
	return ids
}
