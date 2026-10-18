package settings

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrMulaiTx menandai kegagalan memulai transaksi (Begin), agar service
// bisa memakai pesan error yang berbeda dari kegagalan Exec/Commit.
var ErrMulaiTx = errors.New("gagal memulai transaksi")

// --- Repository -------------------------------------------------------------

// Repository adalah lapisan akses data paket settings.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository membuat Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// BacaSemua membaca seluruh baris settings sebagai map key -> raw JSON.
// (Dulu: query inline di handler GET /api/v1/settings.)
func (r *Repository) BacaSemua(ctx context.Context) (map[string]json.RawMessage, error) {
	rows, err := r.pool.Query(ctx, `SELECT key, value FROM settings ORDER BY key ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]json.RawMessage{}
	for rows.Next() {
		var key string
		var raw json.RawMessage
		if err := rows.Scan(&key, &raw); err != nil {
			return nil, err
		}
		out[key] = raw
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// Simpan menyimpan nilai yang sudah divalidasi dalam satu transaksi
// (upsert per key). (Dulu: tx inline di handler PUT /api/v1/settings.)
func (r *Repository) Simpan(ctx context.Context, normalized map[string]any) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return ErrMulaiTx
	}
	defer tx.Rollback(ctx)
	for key, v := range normalized {
		valueJSON, _ := json.Marshal(v)
		if _, err := tx.Exec(ctx,
			`INSERT INTO settings (key, value, updated_at) VALUES ($1, $2, now())
			 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()`,
			key, valueJSON); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
