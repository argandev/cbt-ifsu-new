package notifikasi

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Repository -------------------------------------------------------------

// Repository adalah lapisan akses data paket notifikasi.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository membuat Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// ListByPenerima membaca notifikasi milik satu penerima, terbaru dulu.
// (Dulu: query inline di handler GET /api/v1/notifikasi.)
func (r *Repository) ListByPenerima(ctx context.Context, penerimaUserID string, limit, offset int) ([]map[string]any, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, penerima_user_id, judul, isi, dibaca, created_at
		 FROM notifikasi WHERE penerima_user_id = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		penerimaUserID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, penerima, judul, isi string
		var dibaca bool
		var createdAt any
		if err := rows.Scan(&id, &penerima, &judul, &isi, &dibaca, &createdAt); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{
			"id": id, "penerima_user_id": penerima, "judul": judul,
			"isi": isi, "dibaca": dibaca, "created_at": createdAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// TandaiDibaca menandai satu notifikasi milik penerima sebagai dibaca.
// Mengembalikan true bila ada baris yang terpengaruh (404 bila false).
// (Dulu: Exec inline di handler POST /api/v1/notifikasi/:id/baca.)
func (r *Repository) TandaiDibaca(ctx context.Context, id, penerimaUserID string) (bool, error) {
	ct, err := r.pool.Exec(ctx,
		`UPDATE notifikasi SET dibaca = TRUE WHERE id = $1 AND penerima_user_id = $2`,
		id, penerimaUserID)
	if err != nil {
		return false, err
	}
	return ct.RowsAffected() > 0, nil
}
