package exam

import (
	"context"
	"time"
)

// jadwalItem adalah jadwal + flag token_ada untuk listJadwal.
type jadwalItem struct {
	jadwal
	TokenAda bool `json:"token_ada"`
}

// ListJadwal memuat daftar jadwal (filter status/kelas/tanggal opsional).
func (r *Repository) ListJadwal(ctx context.Context, status, kelasID, tanggal string) ([]jadwalItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT j.id, j.paket_id, j.kelas_id, j.tanggal::text, j.jam_mulai::text,
		        j.jam_selesai::text, j.status, j.token_expires_at, j.created_at,
		        p.kode, p.nama, p.jenis, COALESCE(k.nama, '')
		 FROM jadwal_ujian j
		 JOIN paket_ujian p ON p.id = j.paket_id
		 LEFT JOIN kelas k ON k.id = j.kelas_id
		 WHERE ($1::text IS NULL OR j.status = $1)
		   AND ($2::uuid IS NULL OR j.kelas_id = $2)
		   AND ($3::date IS NULL OR j.tanggal = $3)
		 ORDER BY j.tanggal DESC, j.jam_mulai DESC`,
		nilOrStr(status), nilOrStr(kelasID), nilOrStr(tanggal))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []jadwalItem{}
	for rows.Next() {
		var it jadwalItem
		var tokenExp any
		if err := rows.Scan(&it.ID, &it.PaketID, &it.KelasID, &it.Tanggal, &it.JamMulai,
			&it.JamSelesai, &it.Status, &tokenExp, &it.CreatedAt,
			&it.PaketKode, &it.PaketNama, &it.PaketJenis, &it.KelasNama); err != nil {
			return nil, err
		}
		it.TokenAda = tokenExp != nil
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// PaketExists melaporkan apakah paket ada.
func (r *Repository) PaketExists(ctx context.Context, paketID string) (bool, error) {
	var ada bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM paket_ujian WHERE id = $1)`, paketID).Scan(&ada)
	return ada, err
}

// KelasExists melaporkan apakah kelas ada.
func (r *Repository) KelasExists(ctx context.Context, kelasID string) (bool, error) {
	var ada bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM kelas WHERE id = $1)`, kelasID).Scan(&ada)
	return ada, err
}

// InsertJadwal menyisipkan jadwal baru, mengembalikan id-nya.
func (r *Repository) InsertJadwal(ctx context.Context, j jadwal) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO jadwal_ujian (paket_id, kelas_id, tanggal, jam_mulai, jam_selesai, status)
		 VALUES ($1, $2, $3::date, $4::time, $5::time, $6) RETURNING id`,
		j.PaketID, j.KelasID, j.Tanggal, j.JamMulai, j.JamSelesai, j.Status).Scan(&id)
	return id, err
}

// JadwalStatusOnly mengambil status jadwal (untuk updateJadwal).
func (r *Repository) JadwalStatusOnly(ctx context.Context, id string) (string, error) {
	var st string
	err := r.pool.QueryRow(ctx, `SELECT status FROM jadwal_ujian WHERE id = $1`, id).Scan(&st)
	return st, err
}

// UpdateJadwal mengubah jadwal; mengembalikan baris terpengaruh.
func (r *Repository) UpdateJadwal(ctx context.Context, id string, j jadwal) (int64, error) {
	ct, err := r.pool.Exec(ctx,
		`UPDATE jadwal_ujian SET paket_id=$2, kelas_id=$3, tanggal=$4::date,
		        jam_mulai=$5::time, jam_selesai=$6::time, status=$7 WHERE id=$1`,
		id, j.PaketID, j.KelasID, j.Tanggal, j.JamMulai, j.JamSelesai, j.Status)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}

// CountAttemptJadwal menghitung attempt pada jadwal.
func (r *Repository) CountAttemptJadwal(ctx context.Context, id string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM attempt WHERE jadwal_id = $1`, id).Scan(&n)
	return n, err
}

// DeleteJadwal menghapus jadwal; mengembalikan baris terpengaruh.
func (r *Repository) DeleteJadwal(ctx context.Context, id string) (int64, error) {
	ct, err := r.pool.Exec(ctx, `DELETE FROM jadwal_ujian WHERE id = $1`, id)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}

// PauseJadwalStatus mengubah status jadwal menjadi 'pause'.
// q bisa pool atau tx.
func (r *Repository) PauseJadwalStatus(ctx context.Context, q querier, id string) error {
	_, err := q.Exec(ctx, `UPDATE jadwal_ujian SET status = 'pause' WHERE id = $1`, id)
	return err
}

// BekukanTimer membekukan sisa waktu semua attempt berjalan:
// sisa = deadline - now. q bisa pool atau tx.
func (r *Repository) BekukanTimer(ctx context.Context, q querier, id string) error {
	_, err := q.Exec(ctx,
		`UPDATE attempt
		 SET sisa_waktu_beku = GREATEST(0, EXTRACT(EPOCH FROM (deadline - now()))::INT)
		 WHERE jadwal_id = $1 AND status = 'berjalan'`, id)
	return err
}

// AktifkanJadwal mengubah status jadwal menjadi 'aktif'.
// q bisa pool atau tx.
func (r *Repository) AktifkanJadwal(ctx context.Context, q querier, id string) error {
	_, err := q.Exec(ctx, `UPDATE jadwal_ujian SET status = 'aktif' WHERE id = $1`, id)
	return err
}

// PulihkanTimer memulihkan deadline attempt (deadline baru =
// now() + sisa_waktu_beku). q bisa pool atau tx.
func (r *Repository) PulihkanTimer(ctx context.Context, q querier, id string) error {
	_, err := q.Exec(ctx,
		`UPDATE attempt
		 SET deadline = now() + make_interval(secs => COALESCE(sisa_waktu_beku, 0)),
		     sisa_waktu_beku = NULL
		 WHERE jadwal_id = $1 AND status = 'berjalan' AND sisa_waktu_beku IS NOT NULL`, id)
	return err
}

// DraftKeAktif mengubah jadwal draft menjadi aktif (peserta pertama mulai).
// q bisa pool atau tx.
func (r *Repository) DraftKeAktif(ctx context.Context, q querier, jadwalID string) error {
	_, err := q.Exec(ctx, `UPDATE jadwal_ujian SET status = 'aktif' WHERE id = $1`, jadwalID)
	return err
}

// InsertAttempt menyisipkan attempt baru. q bisa pool atau tx.
func (r *Repository) InsertAttempt(ctx context.Context, q querier, attemptID, jadwalID, siswaID string, selesai time.Time, urutanSoalJSON, urutanOpsiJSON []byte) error {
	_, err := q.Exec(ctx,
		`INSERT INTO attempt (id, jadwal_id, siswa_id, status, mulai_at, deadline,
		                       urutan_soal, urutan_opsi)
		 VALUES ($1, $2, $3, 'berjalan', now(), $4, $5, $6)`,
		attemptID, jadwalID, siswaID, selesai, urutanSoalJSON, urutanOpsiJSON)
	return err
}

// UpdateToken menyimpan hash + expiry token baru (token lama HANGUS).
func (r *Repository) UpdateToken(ctx context.Context, id, hash string, expiry time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE jadwal_ujian SET token_hash = $2, token_expires_at = $3 WHERE id = $1`,
		id, hash, expiry)
	return err
}
