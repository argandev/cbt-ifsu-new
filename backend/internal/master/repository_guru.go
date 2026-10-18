package master

import (
	"context"
	"encoding/json"
)

// ---------------------------------------------------------------------
// Guru
// ---------------------------------------------------------------------

// CountGuru menghitung total guru sesuai pencarian.
func (r *Repository) CountGuru(ctx context.Context, q *string) (int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM guru
		 WHERE ($1::text IS NULL OR nama ILIKE '%' || $1 || '%' OR COALESCE(nip, '') ILIKE '%' || $1 || '%')`, q).
		Scan(&total)
	return total, err
}

// ListGuru memuat halaman guru, terurut nama.
func (r *Repository) ListGuru(ctx context.Context, q *string, limit, offset int) ([]guruDTO, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, nip, nama, mapel_ids, created_at, updated_at FROM guru
		 WHERE ($1::text IS NULL OR nama ILIKE '%' || $1 || '%' OR COALESCE(nip, '') ILIKE '%' || $1 || '%')
		 ORDER BY nama ASC LIMIT $2 OFFSET $3`, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []guruDTO{}
	for rows.Next() {
		var d guruDTO
		var userIDV, nipV *string
		var mapelIDs json.RawMessage
		if err := rows.Scan(&d.ID, &userIDV, &nipV, &d.Nama, &mapelIDs, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		d.UserID, d.NIP, d.MapelIDs = userIDV, nipV, mapelIDs
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// GetGuru memuat satu guru.
func (r *Repository) GetGuru(ctx context.Context, id string) (*guruDTO, error) {
	var d guruDTO
	var userIDV, nipV *string
	var mapelIDs json.RawMessage
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, nip, nama, mapel_ids, created_at, updated_at FROM guru WHERE id = $1`, id).
		Scan(&d.ID, &userIDV, &nipV, &d.Nama, &mapelIDs, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	d.UserID, d.NIP, d.MapelIDs = userIDV, nipV, mapelIDs
	return &d, nil
}

// InsertGuru menyisipkan guru baru.
func (r *Repository) InsertGuru(ctx context.Context, userID, nip any, nama string, mapelIDs json.RawMessage) (*guruDTO, error) {
	var d guruDTO
	var userIDV, nipV *string
	var mapelIDsV json.RawMessage
	err := r.pool.QueryRow(ctx,
		`INSERT INTO guru (user_id, nip, nama, mapel_ids) VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, nip, nama, mapel_ids, created_at, updated_at`,
		userID, nip, nama, mapelIDs).
		Scan(&d.ID, &userIDV, &nipV, &d.Nama, &mapelIDsV, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	d.UserID, d.NIP, d.MapelIDs = userIDV, nipV, mapelIDsV
	return &d, nil
}

// UpdateGuru mengubah guru.
func (r *Repository) UpdateGuru(ctx context.Context, id string, userID, nip any, nama string, mapelIDs json.RawMessage) (*guruDTO, error) {
	var d guruDTO
	var userIDV, nipV *string
	var mapelIDsV json.RawMessage
	err := r.pool.QueryRow(ctx,
		`UPDATE guru SET user_id = $2, nip = $3, nama = $4, mapel_ids = $5, updated_at = now()
		 WHERE id = $1
		 RETURNING id, user_id, nip, nama, mapel_ids, created_at, updated_at`,
		id, userID, nip, nama, mapelIDs).
		Scan(&d.ID, &userIDV, &nipV, &d.Nama, &mapelIDsV, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	d.UserID, d.NIP, d.MapelIDs = userIDV, nipV, mapelIDsV
	return &d, nil
}

// GetGuruUserID mengembalikan user_id guru (nil bila tidak ada).
// Error diabaikan seperti kode asal (_ = ...) — operasi berikutnya yang
// menentukan 404 bila baris tidak ada.
func (r *Repository) GetGuruUserID(ctx context.Context, id string) *string {
	var userID *string
	_ = r.pool.QueryRow(ctx, `SELECT user_id FROM guru WHERE id = $1`, id).Scan(&userID)
	return userID
}

// DeleteGuru menghapus guru; mengembalikan baris terpengaruh.
func (r *Repository) DeleteGuru(ctx context.Context, id string) (int64, error) {
	ct, err := r.pool.Exec(ctx, `DELETE FROM guru WHERE id = $1`, id)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}
