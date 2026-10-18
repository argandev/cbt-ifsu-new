package master

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// ---------------------------------------------------------------------
// Siswa
// ---------------------------------------------------------------------

// CountSiswa menghitung total siswa sesuai filter.
func (r *Repository) CountSiswa(ctx context.Context, kelasID, jurusanID *string, isActive *bool, q *string) (int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM siswa
		 WHERE ($1::uuid IS NULL OR kelas_id = $1)
		   AND ($2::uuid IS NULL OR jurusan_id = $2)
		   AND ($3::boolean IS NULL OR is_active = $3)
		   AND ($4::text IS NULL OR nama ILIKE '%' || $4 || '%' OR nisn ILIKE '%' || $4 || '%' OR COALESCE(nis, '') ILIKE '%' || $4 || '%')`,
		kelasID, jurusanID, isActive, q).Scan(&total)
	return total, err
}

// ListSiswa memuat halaman siswa, terurut nama.
func (r *Repository) ListSiswa(ctx context.Context, kelasID, jurusanID *string, isActive *bool, q *string, limit, offset int) ([]siswaDTO, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, nisn, nis, nama, kelas_id, jurusan_id, tahun_masuk, is_active, created_at, updated_at FROM siswa
		 WHERE ($1::uuid IS NULL OR kelas_id = $1)
		   AND ($2::uuid IS NULL OR jurusan_id = $2)
		   AND ($3::boolean IS NULL OR is_active = $3)
		   AND ($4::text IS NULL OR nama ILIKE '%' || $4 || '%' OR nisn ILIKE '%' || $4 || '%' OR COALESCE(nis, '') ILIKE '%' || $4 || '%')
		 ORDER BY nama ASC LIMIT $5 OFFSET $6`,
		kelasID, jurusanID, isActive, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []siswaDTO{}
	for rows.Next() {
		var d siswaDTO
		var userIDV, nisV *string
		var tahunMasukV *int32
		if err := rows.Scan(&d.ID, &userIDV, &d.NISN, &nisV, &d.Nama, &d.KelasID, &d.JurusanID, &tahunMasukV, &d.IsActive, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		d.UserID, d.NIS, d.TahunMasuk = userIDV, nisV, tahunMasukV
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// GetSiswa memuat satu siswa.
func (r *Repository) GetSiswa(ctx context.Context, id string) (*siswaDTO, error) {
	var d siswaDTO
	var userIDV, nisV *string
	var tahunMasukV *int32
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, nisn, nis, nama, kelas_id, jurusan_id, tahun_masuk, is_active, created_at, updated_at FROM siswa WHERE id = $1`, id).
		Scan(&d.ID, &userIDV, &d.NISN, &nisV, &d.Nama, &d.KelasID, &d.JurusanID, &tahunMasukV, &d.IsActive, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	d.UserID, d.NIS, d.TahunMasuk = userIDV, nisV, tahunMasukV
	return &d, nil
}

// InsertSiswaTx menyisipkan siswa baru di dalam transaksi.
func (r *Repository) InsertSiswaTx(ctx context.Context, tx pgx.Tx, userIDArg any, nisn string, nis *string, nama, kelasID, jurusanID string, tahunMasuk *int32, isActive bool) (*siswaDTO, error) {
	var d siswaDTO
	var userIDV, nisV *string
	var tahunMasukV *int32
	err := tx.QueryRow(ctx,
		`INSERT INTO siswa (user_id, nisn, nis, nama, kelas_id, jurusan_id, tahun_masuk, is_active)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, user_id, nisn, nis, nama, kelas_id, jurusan_id, tahun_masuk, is_active, created_at, updated_at`,
		userIDArg, nisn, nis, nama, kelasID, jurusanID, tahunMasuk, isActive).
		Scan(&d.ID, &userIDV, &d.NISN, &nisV, &d.Nama, &d.KelasID, &d.JurusanID, &tahunMasukV, &d.IsActive, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	d.UserID, d.NIS, d.TahunMasuk = userIDV, nisV, tahunMasukV
	return &d, nil
}

// UpdateSiswa mengubah siswa.
func (r *Repository) UpdateSiswa(ctx context.Context, id string, userIDArg any, nisn string, nis *string, nama, kelasID, jurusanID string, tahunMasuk *int32, isActive bool) (*siswaDTO, error) {
	var d siswaDTO
	var userIDV, nisV *string
	var tahunMasukV *int32
	err := r.pool.QueryRow(ctx,
		`UPDATE siswa SET user_id = $2, nisn = $3, nis = $4, nama = $5, kelas_id = $6, jurusan_id = $7,
		   tahun_masuk = $8, is_active = $9, updated_at = now()
		 WHERE id = $1
		 RETURNING id, user_id, nisn, nis, nama, kelas_id, jurusan_id, tahun_masuk, is_active, created_at, updated_at`,
		id, userIDArg, nisn, nis, nama, kelasID, jurusanID, tahunMasuk, isActive).
		Scan(&d.ID, &userIDV, &d.NISN, &nisV, &d.Nama, &d.KelasID, &d.JurusanID, &tahunMasukV, &d.IsActive, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	d.UserID, d.NIS, d.TahunMasuk = userIDV, nisV, tahunMasukV
	return &d, nil
}

// GetSiswaUserID mengembalikan user_id siswa (nil bila tidak ada).
// Error diabaikan seperti kode asal (_ = ...) — operasi berikutnya yang
// menentukan 404 bila baris tidak ada.
func (r *Repository) GetSiswaUserID(ctx context.Context, id string) *string {
	var userID *string
	_ = r.pool.QueryRow(ctx, `SELECT user_id FROM siswa WHERE id = $1`, id).Scan(&userID)
	return userID
}

// DeleteSiswa menghapus siswa; mengembalikan baris terpengaruh.
func (r *Repository) DeleteSiswa(ctx context.Context, id string) (int64, error) {
	ct, err := r.pool.Exec(ctx, `DELETE FROM siswa WHERE id = $1`, id)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}
