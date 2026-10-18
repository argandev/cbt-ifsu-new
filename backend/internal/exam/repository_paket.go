package exam

import (
	"context"
)

// paketSoalItem adalah satu baris susunan soal paket.
type paketSoalItem struct {
	SoalID     string `json:"soal_id"`
	Urutan     int    `json:"urutan"`
	Tipe       string `json:"tipe"`
	Pertanyaan string `json:"pertanyaan"`
	Bobot      int    `json:"bobot"`
}

// ListPaket memuat daftar paket ujian (filter jenis + mapel opsional).
func (r *Repository) ListPaket(ctx context.Context, jenis, mapelID string) ([]paket, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT p.id, p.kode, p.nama, p.mapel_id, COALESCE(m.nama, ''), p.jenis,
		        p.durasi_menit, p.jumlah_soal, p.acak_soal, p.acak_jawaban,
		        p.tampil_hasil, p.created_at,
		        (SELECT COUNT(*) FROM paket_soal ps WHERE ps.paket_id = p.id)
		 FROM paket_ujian p LEFT JOIN mapel m ON m.id = p.mapel_id
		 WHERE ($1::text IS NULL OR p.jenis = $1)
		   AND ($2::uuid IS NULL OR p.mapel_id = $2)
		 ORDER BY p.created_at DESC`, nilOrStr(jenis), nilOrStr(mapelID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []paket{}
	for rows.Next() {
		var p paket
		if err := rows.Scan(&p.ID, &p.Kode, &p.Nama, &p.MapelID, &p.MapelNama, &p.Jenis,
			&p.DurasiMenit, &p.JumlahSoal, &p.AcakSoal, &p.AcakJawaban, &p.TampilHasil,
			&p.CreatedAt, &p.JumlahTersusun); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// MapelExists melaporkan apakah mapel ada.
func (r *Repository) MapelExists(ctx context.Context, mapelID string) (bool, error) {
	var ada bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM mapel WHERE id = $1)`, mapelID).Scan(&ada)
	return ada, err
}

// InsertPaket menyisipkan paket baru, mengembalikan id-nya.
func (r *Repository) InsertPaket(ctx context.Context, p paket) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO paket_ujian (kode, nama, mapel_id, jenis, durasi_menit, jumlah_soal,
		                           acak_soal, acak_jawaban, tampil_hasil)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`,
		p.Kode, p.Nama, p.MapelID, p.Jenis, p.DurasiMenit, p.JumlahSoal,
		p.AcakSoal, p.AcakJawaban, p.TampilHasil).Scan(&id)
	return id, err
}

// GetPaketDetail memuat satu paket + jumlah tersusun.
func (r *Repository) GetPaketDetail(ctx context.Context, id string) (paket, error) {
	var p paket
	err := r.pool.QueryRow(ctx,
		`SELECT p.id, p.kode, p.nama, p.mapel_id, COALESCE(m.nama, ''), p.jenis,
		        p.durasi_menit, p.jumlah_soal, p.acak_soal, p.acak_jawaban,
		        p.tampil_hasil, p.created_at,
		        (SELECT COUNT(*) FROM paket_soal ps WHERE ps.paket_id = p.id)
		 FROM paket_ujian p LEFT JOIN mapel m ON m.id = p.mapel_id
		 WHERE p.id = $1`, id).
		Scan(&p.ID, &p.Kode, &p.Nama, &p.MapelID, &p.MapelNama, &p.Jenis,
			&p.DurasiMenit, &p.JumlahSoal, &p.AcakSoal, &p.AcakJawaban,
			&p.TampilHasil, &p.CreatedAt, &p.JumlahTersusun)
	return p, err
}

// PaketMapelID mengambil mapel_id paket (untuk updatePaket).
func (r *Repository) PaketMapelID(ctx context.Context, paketID string) (string, error) {
	var mapelID string
	err := r.pool.QueryRow(ctx, `SELECT mapel_id FROM paket_ujian WHERE id = $1`, paketID).Scan(&mapelID)
	return mapelID, err
}

// CountSoalMapelLain menghitung soal paket yang bukan dari mapel baru.
func (r *Repository) CountSoalMapelLain(ctx context.Context, paketID, mapelID string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM paket_soal ps
		 JOIN soal s ON s.id = ps.soal_id
		 JOIN bank_soal bs ON bs.id = s.bank_id
		 WHERE ps.paket_id = $1 AND bs.mapel_id <> $2`, paketID, mapelID).Scan(&n)
	return n, err
}

// UpdatePaket mengubah paket; mengembalikan baris terpengaruh.
func (r *Repository) UpdatePaket(ctx context.Context, id string, p paket) (int64, error) {
	ct, err := r.pool.Exec(ctx,
		`UPDATE paket_ujian SET kode=$2, nama=$3, mapel_id=$4, jenis=$5, durasi_menit=$6,
		        jumlah_soal=$7, acak_soal=$8, acak_jawaban=$9, tampil_hasil=$10
		 WHERE id=$1`, id, p.Kode, p.Nama, p.MapelID, p.Jenis, p.DurasiMenit,
		p.JumlahSoal, p.AcakSoal, p.AcakJawaban, p.TampilHasil)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}

// CountJadwalPaket menghitung jadwal yang memakai paket.
func (r *Repository) CountJadwalPaket(ctx context.Context, paketID string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM jadwal_ujian WHERE paket_id = $1`, paketID).Scan(&n)
	return n, err
}

// DeletePaket menghapus paket; mengembalikan baris terpengaruh.
func (r *Repository) DeletePaket(ctx context.Context, paketID string) (int64, error) {
	ct, err := r.pool.Exec(ctx, `DELETE FROM paket_ujian WHERE id = $1`, paketID)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}

// ListPaketSoal memuat susunan soal paket terurut.
func (r *Repository) ListPaketSoal(ctx context.Context, paketID string) ([]paketSoalItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT ps.soal_id, ps.urutan, s.tipe, s.pertanyaan, s.bobot
		 FROM paket_soal ps JOIN soal s ON s.id = ps.soal_id
		 WHERE ps.paket_id = $1 ORDER BY ps.urutan ASC`, paketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []paketSoalItem{}
	for rows.Next() {
		var it paketSoalItem
		if err := rows.Scan(&it.SoalID, &it.Urutan, &it.Tipe, &it.Pertanyaan, &it.Bobot); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// SoalMapelID mengambil mapel_id soal lewat bank_soal.
func (r *Repository) SoalMapelID(ctx context.Context, soalID string) (string, error) {
	var mapelID string
	err := r.pool.QueryRow(ctx,
		`SELECT bs.mapel_id FROM soal s JOIN bank_soal bs ON bs.id = s.bank_id WHERE s.id = $1`,
		soalID).Scan(&mapelID)
	return mapelID, err
}

// MaxUrutanPaketSoal mengambil urutan maksimum susunan paket.
func (r *Repository) MaxUrutanPaketSoal(ctx context.Context, paketID string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COALESCE(MAX(urutan), 0) FROM paket_soal WHERE paket_id = $1`, paketID).Scan(&n)
	return n, err
}

// InsertPaketSoal menambah soal ke paket (idempotent via ON CONFLICT).
func (r *Repository) InsertPaketSoal(ctx context.Context, paketID, soalID string, urutan int) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO paket_soal (paket_id, soal_id, urutan) VALUES ($1, $2, $3)
		 ON CONFLICT (paket_id, soal_id) DO NOTHING`, paketID, soalID, urutan)
	return err
}

// DeletePaketSoal menghapus soal dari paket; mengembalikan baris terpengaruh.
func (r *Repository) DeletePaketSoal(ctx context.Context, paketID, soalID string) (int64, error) {
	ct, err := r.pool.Exec(ctx, `DELETE FROM paket_soal WHERE paket_id = $1 AND soal_id = $2`, paketID, soalID)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}

// SinkronJumlahSoal menyamakan paket.jumlah_soal dengan jumlah susunan aktual.
func (r *Repository) SinkronJumlahSoal(ctx context.Context, paketID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE paket_ujian SET jumlah_soal = (SELECT COUNT(*) FROM paket_soal WHERE paket_id = $1)
		 WHERE id = $1`, paketID)
	return err
}

// ListSoalAcak mengambil id soal aktif acak dari mapel yang belum ada di paket.
func (r *Repository) ListSoalAcak(ctx context.Context, mapelID, paketID string, jumlah int) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT s.id FROM soal s JOIN bank_soal bs ON bs.id = s.bank_id
		 WHERE bs.mapel_id = $1 AND s.status = 'aktif'
		   AND s.id NOT IN (SELECT soal_id FROM paket_soal WHERE paket_id = $2)
		 ORDER BY random() LIMIT $3`, mapelID, paketID, jumlah)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}
