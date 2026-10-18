package master

import (
	"context"
)

// ---------------------------------------------------------------------
// Kompetensi
// ---------------------------------------------------------------------

// CountKompetensi menghitung total kompetensi sesuai filter.
func (r *Repository) CountKompetensi(ctx context.Context, mapelID, q *string) (int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM kompetensi
		 WHERE ($1::uuid IS NULL OR mapel_id = $1)
		   AND ($2::text IS NULL OR kode ILIKE '%' || $2 || '%' OR deskripsi ILIKE '%' || $2 || '%')`,
		mapelID, q).Scan(&total)
	return total, err
}

// ListKompetensi memuat halaman kompetensi, terurut kode.
func (r *Repository) ListKompetensi(ctx context.Context, mapelID, q *string, limit, offset int) ([]kompetensiDTO, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, mapel_id, kode, deskripsi, created_at, updated_at FROM kompetensi
		 WHERE ($1::uuid IS NULL OR mapel_id = $1)
		   AND ($2::text IS NULL OR kode ILIKE '%' || $2 || '%' OR deskripsi ILIKE '%' || $2 || '%')
		 ORDER BY kode ASC LIMIT $3 OFFSET $4`, mapelID, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []kompetensiDTO{}
	for rows.Next() {
		var d kompetensiDTO
		if err := rows.Scan(&d.ID, &d.MapelID, &d.Kode, &d.Deskripsi, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// GetKompetensi memuat satu kompetensi.
func (r *Repository) GetKompetensi(ctx context.Context, id string) (*kompetensiDTO, error) {
	var d kompetensiDTO
	err := r.pool.QueryRow(ctx,
		`SELECT id, mapel_id, kode, deskripsi, created_at, updated_at FROM kompetensi WHERE id = $1`, id).
		Scan(&d.ID, &d.MapelID, &d.Kode, &d.Deskripsi, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// InsertKompetensi menyisipkan kompetensi baru.
func (r *Repository) InsertKompetensi(ctx context.Context, mapelID, kode, deskripsi string) (*kompetensiDTO, error) {
	var d kompetensiDTO
	err := r.pool.QueryRow(ctx,
		`INSERT INTO kompetensi (mapel_id, kode, deskripsi) VALUES ($1, $2, $3)
		 RETURNING id, mapel_id, kode, deskripsi, created_at, updated_at`,
		mapelID, kode, deskripsi).
		Scan(&d.ID, &d.MapelID, &d.Kode, &d.Deskripsi, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// UpdateKompetensi mengubah kompetensi.
func (r *Repository) UpdateKompetensi(ctx context.Context, id, mapelID, kode, deskripsi string) (*kompetensiDTO, error) {
	var d kompetensiDTO
	err := r.pool.QueryRow(ctx,
		`UPDATE kompetensi SET mapel_id = $2, kode = $3, deskripsi = $4, updated_at = now()
		 WHERE id = $1
		 RETURNING id, mapel_id, kode, deskripsi, created_at, updated_at`,
		id, mapelID, kode, deskripsi).
		Scan(&d.ID, &d.MapelID, &d.Kode, &d.Deskripsi, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// DeleteKompetensi menghapus kompetensi; mengembalikan baris terpengaruh.
func (r *Repository) DeleteKompetensi(ctx context.Context, id string) (int64, error) {
	ct, err := r.pool.Exec(ctx, `DELETE FROM kompetensi WHERE id = $1`, id)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}

// ---------------------------------------------------------------------
// Kisi-kisi
// ---------------------------------------------------------------------

// CountKisiKisi menghitung total kisi-kisi sesuai filter.
func (r *Repository) CountKisiKisi(ctx context.Context, mapelID *string, tingkat *int, q *string) (int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM kisi_kisi
		 WHERE ($1::uuid IS NULL OR mapel_id = $1)
		   AND ($2::int IS NULL OR kelas_tingkat = $2)
		   AND ($3::text IS NULL OR kelompok ILIKE '%' || $3 || '%')`,
		mapelID, tingkat, q).Scan(&total)
	return total, err
}

// ListKisiKisi memuat halaman kisi-kisi, terurut tingkat/kelompok.
func (r *Repository) ListKisiKisi(ctx context.Context, mapelID *string, tingkat *int, q *string, limit, offset int) ([]kisiKisiDTO, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, mapel_id, kelas_tingkat, kelompok, kompetensi_id, jumlah_soal, created_at, updated_at FROM kisi_kisi
		 WHERE ($1::uuid IS NULL OR mapel_id = $1)
		   AND ($2::int IS NULL OR kelas_tingkat = $2)
		   AND ($3::text IS NULL OR kelompok ILIKE '%' || $3 || '%')
		 ORDER BY kelas_tingkat ASC, kelompok ASC LIMIT $4 OFFSET $5`,
		mapelID, tingkat, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []kisiKisiDTO{}
	for rows.Next() {
		var d kisiKisiDTO
		var kompetensiIDV *string
		if err := rows.Scan(&d.ID, &d.MapelID, &d.KelasTingkat, &d.Kelompok, &kompetensiIDV, &d.JumlahSoal, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		d.KompetensiID = kompetensiIDV
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// GetKisiKisi memuat satu kisi-kisi.
func (r *Repository) GetKisiKisi(ctx context.Context, id string) (*kisiKisiDTO, error) {
	var d kisiKisiDTO
	var kompetensiIDV *string
	err := r.pool.QueryRow(ctx,
		`SELECT id, mapel_id, kelas_tingkat, kelompok, kompetensi_id, jumlah_soal, created_at, updated_at FROM kisi_kisi WHERE id = $1`, id).
		Scan(&d.ID, &d.MapelID, &d.KelasTingkat, &d.Kelompok, &kompetensiIDV, &d.JumlahSoal, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	d.KompetensiID = kompetensiIDV
	return &d, nil
}

// InsertKisiKisi menyisipkan kisi-kisi baru.
func (r *Repository) InsertKisiKisi(ctx context.Context, mapelID string, kelasTingkat int32, kelompok string, kompetensiID *string, jumlahSoal int32) (*kisiKisiDTO, error) {
	var d kisiKisiDTO
	var kompetensiIDV *string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO kisi_kisi (mapel_id, kelas_tingkat, kelompok, kompetensi_id, jumlah_soal)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, mapel_id, kelas_tingkat, kelompok, kompetensi_id, jumlah_soal, created_at, updated_at`,
		mapelID, kelasTingkat, kelompok, kompetensiID, jumlahSoal).
		Scan(&d.ID, &d.MapelID, &d.KelasTingkat, &d.Kelompok, &kompetensiIDV, &d.JumlahSoal, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	d.KompetensiID = kompetensiIDV
	return &d, nil
}

// UpdateKisiKisi mengubah kisi-kisi.
func (r *Repository) UpdateKisiKisi(ctx context.Context, id, mapelID string, kelasTingkat int32, kelompok string, kompetensiID *string, jumlahSoal int32) (*kisiKisiDTO, error) {
	var d kisiKisiDTO
	var kompetensiIDV *string
	err := r.pool.QueryRow(ctx,
		`UPDATE kisi_kisi SET mapel_id = $2, kelas_tingkat = $3, kelompok = $4, kompetensi_id = $5,
		   jumlah_soal = $6, updated_at = now()
		 WHERE id = $1
		 RETURNING id, mapel_id, kelas_tingkat, kelompok, kompetensi_id, jumlah_soal, created_at, updated_at`,
		id, mapelID, kelasTingkat, kelompok, kompetensiID, jumlahSoal).
		Scan(&d.ID, &d.MapelID, &d.KelasTingkat, &d.Kelompok, &kompetensiIDV, &d.JumlahSoal, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	d.KompetensiID = kompetensiIDV
	return &d, nil
}

// DeleteKisiKisi menghapus kisi-kisi; mengembalikan baris terpengaruh.
func (r *Repository) DeleteKisiKisi(ctx context.Context, id string) (int64, error) {
	ct, err := r.pool.Exec(ctx, `DELETE FROM kisi_kisi WHERE id = $1`, id)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}
