package master

import (
	"context"
)

// ---------------------------------------------------------------------
// Jurusan
// ---------------------------------------------------------------------

// CountJurusan menghitung total jurusan sesuai pencarian.
func (r *Repository) CountJurusan(ctx context.Context, q *string) (int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM jurusan
		 WHERE ($1::text IS NULL OR nama ILIKE '%' || $1 || '%' OR kode ILIKE '%' || $1 || '%')`, q).Scan(&total)
	return total, err
}

// ListJurusan memuat halaman jurusan, terurut nama.
func (r *Repository) ListJurusan(ctx context.Context, q *string, limit, offset int) ([]jurusanDTO, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, kode, nama, created_at, updated_at FROM jurusan
		 WHERE ($1::text IS NULL OR nama ILIKE '%' || $1 || '%' OR kode ILIKE '%' || $1 || '%')
		 ORDER BY nama ASC LIMIT $2 OFFSET $3`, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []jurusanDTO{}
	for rows.Next() {
		var d jurusanDTO
		if err := rows.Scan(&d.ID, &d.Kode, &d.Nama, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// GetJurusan memuat satu jurusan.
func (r *Repository) GetJurusan(ctx context.Context, id string) (*jurusanDTO, error) {
	var d jurusanDTO
	err := r.pool.QueryRow(ctx,
		`SELECT id, kode, nama, created_at, updated_at FROM jurusan WHERE id = $1`, id).
		Scan(&d.ID, &d.Kode, &d.Nama, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// InsertJurusan menyisipkan jurusan baru.
func (r *Repository) InsertJurusan(ctx context.Context, kode, nama string) (*jurusanDTO, error) {
	var d jurusanDTO
	err := r.pool.QueryRow(ctx,
		`INSERT INTO jurusan (kode, nama) VALUES ($1, $2)
		 RETURNING id, kode, nama, created_at, updated_at`,
		kode, nama).
		Scan(&d.ID, &d.Kode, &d.Nama, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// UpdateJurusan mengubah jurusan.
func (r *Repository) UpdateJurusan(ctx context.Context, id, kode, nama string) (*jurusanDTO, error) {
	var d jurusanDTO
	err := r.pool.QueryRow(ctx,
		`UPDATE jurusan SET kode = $2, nama = $3, updated_at = now()
		 WHERE id = $1 RETURNING id, kode, nama, created_at, updated_at`,
		id, kode, nama).
		Scan(&d.ID, &d.Kode, &d.Nama, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// DeleteJurusan menghapus jurusan; mengembalikan baris terpengaruh.
func (r *Repository) DeleteJurusan(ctx context.Context, id string) (int64, error) {
	ct, err := r.pool.Exec(ctx, `DELETE FROM jurusan WHERE id = $1`, id)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}

// ---------------------------------------------------------------------
// Kelas
// ---------------------------------------------------------------------

// CountKelas menghitung total kelas sesuai filter.
func (r *Repository) CountKelas(ctx context.Context, jurusanID, q *string) (int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM kelas
		 WHERE ($1::uuid IS NULL OR jurusan_id = $1)
		   AND ($2::text IS NULL OR nama ILIKE '%' || $2 || '%' OR tahun_ajaran ILIKE '%' || $2 || '%')`,
		jurusanID, q).Scan(&total)
	return total, err
}

// ListKelas memuat halaman kelas, terurut tahun ajaran/tingkat/nama.
func (r *Repository) ListKelas(ctx context.Context, jurusanID, q *string, limit, offset int) ([]kelasDTO, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, jurusan_id, tingkat, nama, tahun_ajaran, created_at, updated_at FROM kelas
		 WHERE ($1::uuid IS NULL OR jurusan_id = $1)
		   AND ($2::text IS NULL OR nama ILIKE '%' || $2 || '%' OR tahun_ajaran ILIKE '%' || $2 || '%')
		 ORDER BY tahun_ajaran DESC, tingkat ASC, nama ASC LIMIT $3 OFFSET $4`,
		jurusanID, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []kelasDTO{}
	for rows.Next() {
		var d kelasDTO
		if err := rows.Scan(&d.ID, &d.JurusanID, &d.Tingkat, &d.Nama, &d.TahunAjaran, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// GetKelas memuat satu kelas.
func (r *Repository) GetKelas(ctx context.Context, id string) (*kelasDTO, error) {
	var d kelasDTO
	err := r.pool.QueryRow(ctx,
		`SELECT id, jurusan_id, tingkat, nama, tahun_ajaran, created_at, updated_at FROM kelas WHERE id = $1`, id).
		Scan(&d.ID, &d.JurusanID, &d.Tingkat, &d.Nama, &d.TahunAjaran, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// InsertKelas menyisipkan kelas baru.
func (r *Repository) InsertKelas(ctx context.Context, jurusanID string, tingkat int32, nama, tahunAjaran string) (*kelasDTO, error) {
	var d kelasDTO
	err := r.pool.QueryRow(ctx,
		`INSERT INTO kelas (jurusan_id, tingkat, nama, tahun_ajaran) VALUES ($1, $2, $3, $4)
		 RETURNING id, jurusan_id, tingkat, nama, tahun_ajaran, created_at, updated_at`,
		jurusanID, tingkat, nama, tahunAjaran).
		Scan(&d.ID, &d.JurusanID, &d.Tingkat, &d.Nama, &d.TahunAjaran, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// UpdateKelas mengubah kelas.
func (r *Repository) UpdateKelas(ctx context.Context, id, jurusanID string, tingkat int32, nama, tahunAjaran string) (*kelasDTO, error) {
	var d kelasDTO
	err := r.pool.QueryRow(ctx,
		`UPDATE kelas SET jurusan_id = $2, tingkat = $3, nama = $4, tahun_ajaran = $5, updated_at = now()
		 WHERE id = $1
		 RETURNING id, jurusan_id, tingkat, nama, tahun_ajaran, created_at, updated_at`,
		id, jurusanID, tingkat, nama, tahunAjaran).
		Scan(&d.ID, &d.JurusanID, &d.Tingkat, &d.Nama, &d.TahunAjaran, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// DeleteKelas menghapus kelas; mengembalikan baris terpengaruh.
func (r *Repository) DeleteKelas(ctx context.Context, id string) (int64, error) {
	ct, err := r.pool.Exec(ctx, `DELETE FROM kelas WHERE id = $1`, id)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}

// ---------------------------------------------------------------------
// Mapel
// ---------------------------------------------------------------------

// CountMapel menghitung total mapel sesuai filter.
func (r *Repository) CountMapel(ctx context.Context, jenis, jurusanID, q *string) (int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM mapel
		 WHERE ($1::text IS NULL OR jenis = $1)
		   AND ($2::uuid IS NULL OR jurusan_id = $2)
		   AND ($3::text IS NULL OR nama ILIKE '%' || $3 || '%' OR kode ILIKE '%' || $3 || '%')`,
		jenis, jurusanID, q).Scan(&total)
	return total, err
}

// ListMapel memuat halaman mapel, terurut nama.
func (r *Repository) ListMapel(ctx context.Context, jenis, jurusanID, q *string, limit, offset int) ([]mapelDTO, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, kode, nama, jenis, jurusan_id, kelas_tingkat, created_at, updated_at FROM mapel
		 WHERE ($1::text IS NULL OR jenis = $1)
		   AND ($2::uuid IS NULL OR jurusan_id = $2)
		   AND ($3::text IS NULL OR nama ILIKE '%' || $3 || '%' OR kode ILIKE '%' || $3 || '%')
		 ORDER BY nama ASC LIMIT $4 OFFSET $5`,
		jenis, jurusanID, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []mapelDTO{}
	for rows.Next() {
		var d mapelDTO
		var jurusanIDV *string
		var tingkatV *int32
		if err := rows.Scan(&d.ID, &d.Kode, &d.Nama, &d.Jenis, &jurusanIDV, &tingkatV, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		d.JurusanID, d.KelasTingkat = jurusanIDV, tingkatV
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// GetMapel memuat satu mapel.
func (r *Repository) GetMapel(ctx context.Context, id string) (*mapelDTO, error) {
	var d mapelDTO
	var jurusanIDV *string
	var tingkatV *int32
	err := r.pool.QueryRow(ctx,
		`SELECT id, kode, nama, jenis, jurusan_id, kelas_tingkat, created_at, updated_at FROM mapel WHERE id = $1`, id).
		Scan(&d.ID, &d.Kode, &d.Nama, &d.Jenis, &jurusanIDV, &tingkatV, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	d.JurusanID, d.KelasTingkat = jurusanIDV, tingkatV
	return &d, nil
}

// InsertMapel menyisipkan mapel baru.
func (r *Repository) InsertMapel(ctx context.Context, kode, nama, jenis string, jurusanID *string, kelasTingkat *int32) (*mapelDTO, error) {
	var d mapelDTO
	var jurusanIDV *string
	var tingkatV *int32
	err := r.pool.QueryRow(ctx,
		`INSERT INTO mapel (kode, nama, jenis, jurusan_id, kelas_tingkat) VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, kode, nama, jenis, jurusan_id, kelas_tingkat, created_at, updated_at`,
		kode, nama, jenis, jurusanID, kelasTingkat).
		Scan(&d.ID, &d.Kode, &d.Nama, &d.Jenis, &jurusanIDV, &tingkatV, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	d.JurusanID, d.KelasTingkat = jurusanIDV, tingkatV
	return &d, nil
}

// UpdateMapel mengubah mapel.
func (r *Repository) UpdateMapel(ctx context.Context, id, kode, nama, jenis string, jurusanID *string, kelasTingkat *int32) (*mapelDTO, error) {
	var d mapelDTO
	var jurusanIDV *string
	var tingkatV *int32
	err := r.pool.QueryRow(ctx,
		`UPDATE mapel SET kode = $2, nama = $3, jenis = $4, jurusan_id = $5, kelas_tingkat = $6, updated_at = now()
		 WHERE id = $1
		 RETURNING id, kode, nama, jenis, jurusan_id, kelas_tingkat, created_at, updated_at`,
		id, kode, nama, jenis, jurusanID, kelasTingkat).
		Scan(&d.ID, &d.Kode, &d.Nama, &d.Jenis, &jurusanIDV, &tingkatV, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	d.JurusanID, d.KelasTingkat = jurusanIDV, tingkatV
	return &d, nil
}

// DeleteMapel menghapus mapel; mengembalikan baris terpengaruh.
func (r *Repository) DeleteMapel(ctx context.Context, id string) (int64, error) {
	ct, err := r.pool.Exec(ctx, `DELETE FROM mapel WHERE id = $1`, id)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}
