package exam

import (
	"context"
	"time"
)

// --- DTO baris untuk kueri ujian -------------------------------------------

type ujianAktifItem struct {
	JadwalID      string  `json:"jadwal_id"`
	PaketID       string  `json:"paket_id"`
	Tanggal       string  `json:"tanggal"`
	JamMulai      string  `json:"jam_mulai"`
	JamSelesai    string  `json:"jam_selesai"`
	Status        string  `json:"status"`
	PaketKode     string  `json:"paket_kode"`
	PaketNama     string  `json:"paket_nama"`
	PaketJenis    string  `json:"paket_jenis"`
	DurasiMenit   int     `json:"durasi_menit"`
	TampilHasil   bool    `json:"tampil_hasil"`
	KelasNama     string  `json:"kelas_nama"`
	AttemptID     *string `json:"attempt_id"`
	AttemptStatus *string `json:"attempt_status"`
}

type soalPaketRow struct {
	ID      string
	Tipe    string
	Jawaban []byte
}

type soalDetailRow struct {
	ID         string
	Tipe       string
	Pertanyaan string
	Tampilan   []byte
	Bobot      int
	Stimulus   string
	MediaURL   string
}

type jawabanTersimpanRow struct {
	SoalID  string
	Jawaban []byte
	Ragu    bool
}

type jawabanNilaiRow struct {
	Poin       float64
	Bobot      float64
	Tipe       string
	Dikoreksi  bool
	AdaJawaban bool
}

type soalNilaiRow struct {
	Bobot float64
	Tipe  string
}

type hasilJawabanRow struct {
	SoalID     string
	Jawaban    []byte
	Poin       *float64
	Ragu       bool
	Tipe       string
	Pertanyaan string
	Tampilan   []byte
	Kunci      []byte
	Bobot      float64
	Pembahasan string
}

type soalHasilRow struct {
	ID         string
	Tipe       string
	Pertanyaan string
	Bobot      float64
	Pembahasan string
}

// KelasSiswa mencari kelas siswa dari tabel siswa (Tim 2).
// Fallback: Tim 2 mungkin mengaitkan siswa ke akun via user_id.
func (r *Repository) KelasSiswa(ctx context.Context, siswaID string) (string, bool) {
	var kelasID string
	if err := r.pool.QueryRow(ctx, `SELECT kelas_id FROM siswa WHERE id = $1`, siswaID).Scan(&kelasID); err == nil && kelasID != "" {
		return kelasID, true
	}
	if err := r.pool.QueryRow(ctx, `SELECT kelas_id FROM siswa WHERE user_id = $1`, siswaID).Scan(&kelasID); err == nil && kelasID != "" {
		return kelasID, true
	}
	return "", false
}

// ListUjianAktif memuat jadwal hari ini untuk kelas siswa + attempt-nya.
func (r *Repository) ListUjianAktif(ctx context.Context, siswaID, kelasID string) ([]ujianAktifItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT j.id, j.paket_id, j.tanggal::text, j.jam_mulai::text, j.jam_selesai::text,
		        j.status, p.kode, p.nama, p.jenis, p.durasi_menit, p.tampil_hasil,
		        COALESCE(k.nama, ''),
		        a.id, a.status
		 FROM jadwal_ujian j
		 JOIN paket_ujian p ON p.id = j.paket_id
		 LEFT JOIN kelas k ON k.id = j.kelas_id
		 LEFT JOIN attempt a ON a.jadwal_id = j.id AND a.siswa_id = $1 AND a.status IN ('berjalan','diblokir','selesai')
		 WHERE j.kelas_id = $2
		   AND j.tanggal = (now() AT TIME ZONE 'Asia/Jakarta')::date
		   AND j.status IN ('draft', 'aktif', 'pause')
		 ORDER BY j.jam_mulai ASC`, siswaID, kelasID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ujianAktifItem{}
	for rows.Next() {
		var it ujianAktifItem
		if err := rows.Scan(&it.JadwalID, &it.PaketID, &it.Tanggal, &it.JamMulai, &it.JamSelesai,
			&it.Status, &it.PaketKode, &it.PaketNama, &it.PaketJenis, &it.DurasiMenit,
			&it.TampilHasil,
			&it.KelasNama, &it.AttemptID, &it.AttemptStatus); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// FindAttemptByStatus mencari attempt terbaru siswa pada jadwal dengan
// status tertentu. Tidak ada -> error (pemanggil membedakan via err==nil).
func (r *Repository) FindAttemptByStatus(ctx context.Context, jadwalID, siswaID, status string) (attempt, error) {
	return scanAttempt(r.pool.QueryRow(ctx,
		`SELECT id, jadwal_id, siswa_id, status, mulai_at, deadline,
		        sisa_waktu_beku, urutan_soal, urutan_opsi, submit_at,
		        nilai_akhir, created_at
		 FROM attempt WHERE jadwal_id = $1 AND siswa_id = $2 AND status = $3
		 ORDER BY created_at DESC LIMIT 1`, jadwalID, siswaID, status))
}

// ListSoalPaket memuat id + tipe + tampilan soal paket (untuk mulaiUjian).
func (r *Repository) ListSoalPaket(ctx context.Context, paketID string) ([]soalPaketRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT s.id, s.tipe, s.jawaban FROM paket_soal ps JOIN soal s ON s.id = ps.soal_id
		 WHERE ps.paket_id = $1 ORDER BY ps.urutan ASC`, paketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []soalPaketRow{}
	for rows.Next() {
		var s soalPaketRow
		if err := rows.Scan(&s.ID, &s.Tipe, &s.Jawaban); err == nil {
			out = append(out, s)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ListSoalDetail memuat detail soal paket (untuk muatSoalAttempt).
func (r *Repository) ListSoalDetail(ctx context.Context, paketID string) ([]soalDetailRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT s.id, s.tipe, s.pertanyaan, s.jawaban, s.bobot, s.stimulus, s.media_url
		 FROM paket_soal ps JOIN soal s ON s.id = ps.soal_id
		 WHERE ps.paket_id = $1 ORDER BY ps.urutan ASC`, paketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []soalDetailRow{}
	for rows.Next() {
		var s soalDetailRow
		if err := rows.Scan(&s.ID, &s.Tipe, &s.Pertanyaan, &s.Tampilan, &s.Bobot, &s.Stimulus, &s.MediaURL); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ListJawabanTersimpan memuat jawaban + ragu tersimpan untuk attempt.
func (r *Repository) ListJawabanTersimpan(ctx context.Context, attemptID string) ([]jawabanTersimpanRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT soal_id, jawaban, ragu FROM jawaban WHERE attempt_id = $1`, attemptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []jawabanTersimpanRow{}
	for rows.Next() {
		var j jawabanTersimpanRow
		if err := rows.Scan(&j.SoalID, &j.Jawaban, &j.Ragu); err == nil {
			out = append(out, j)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// PaketSoalExists melaporkan apakah soal bagian dari paket.
func (r *Repository) PaketSoalExists(ctx context.Context, paketID, soalID string) (bool, error) {
	var ada bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM paket_soal WHERE paket_id = $1 AND soal_id = $2)`,
		paketID, soalID).Scan(&ada)
	return ada, err
}

// SoalUntukSkor mengambil tipe + kunci + bobot (text) + tampilan soal
// untuk penskoran otomatis.
func (r *Repository) SoalUntukSkor(ctx context.Context, soalID string) (tipe string, kunci, bobotRaw, tampilan []byte, err error) {
	err = r.pool.QueryRow(ctx,
		`SELECT tipe, kunci, bobot::text, jawaban FROM soal WHERE id = $1`, soalID).
		Scan(&tipe, &kunci, &bobotRaw, &tampilan)
	return tipe, kunci, bobotRaw, tampilan, err
}

// RaguLama mengambil flag ragu jawaban yang sudah ada.
func (r *Repository) RaguLama(ctx context.Context, attemptID, soalID string) bool {
	var ragu bool
	_ = r.pool.QueryRow(ctx,
		`SELECT ragu FROM jawaban WHERE attempt_id = $1 AND soal_id = $2`,
		attemptID, soalID).Scan(&ragu)
	return ragu
}

// UpsertJawaban menyimpan/memperbarui jawaban (poinArg nil = menunggu koreksi).
func (r *Repository) UpsertJawaban(ctx context.Context, attemptID, soalID string, jawaban []byte, ragu bool, poinArg any, dikoreksi bool) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO jawaban (attempt_id, soal_id, jawaban, ragu, poin, dikoreksi, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, now())
		 ON CONFLICT (attempt_id, soal_id) DO UPDATE SET
		     jawaban = EXCLUDED.jawaban, poin = EXCLUDED.poin,
		     dikoreksi = EXCLUDED.dikoreksi, updated_at = now()`,
		attemptID, soalID, jawaban, ragu, poinArg, dikoreksi)
	return err
}

// UpsertRagu menyimpan/memperbarui flag ragu-ragu.
func (r *Repository) UpsertRagu(ctx context.Context, attemptID, soalID string, ragu bool) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO jawaban (attempt_id, soal_id, jawaban, ragu)
		 VALUES ($1, $2, NULL, $3)
		 ON CONFLICT (attempt_id, soal_id) DO UPDATE SET ragu = EXCLUDED.ragu, updated_at = now()`,
		attemptID, soalID, ragu)
	return err
}

// PelanggaranTerakhir mengembalikan waktu pelanggaran terakhir + jumlah.
func (r *Repository) PelanggaranTerakhir(ctx context.Context, attemptID string) (*time.Time, int) {
	var terakhir *time.Time
	_ = r.pool.QueryRow(ctx,
		`SELECT created_at FROM pelanggaran WHERE attempt_id = $1
		 ORDER BY created_at DESC LIMIT 1`, attemptID).Scan(&terakhir)
	var jumlah int
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM pelanggaran WHERE attempt_id = $1`, attemptID).Scan(&jumlah)
	return terakhir, jumlah
}

// InsertPelanggaran mencatat satu pelanggaran (detail kosong -> NULL).
func (r *Repository) InsertPelanggaran(ctx context.Context, attemptID, jenis, detail string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO pelanggaran (attempt_id, jenis, detail) VALUES ($1, $2, NULLIF($3, ''))`,
		attemptID, jenis, detail)
	return err
}

// BlokirAttempt memblokir attempt + membekukan sisa waktu.
func (r *Repository) BlokirAttempt(ctx context.Context, attemptID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE attempt SET status = 'diblokir',
		        sisa_waktu_beku = GREATEST(0, EXTRACT(EPOCH FROM (deadline - now()))::INT)
		 WHERE id = $1`, attemptID)
	return err
}

// AttemptRingkasan menghitung dijawab/ragu/total untuk statusAttempt.
func (r *Repository) AttemptRingkasan(ctx context.Context, attemptID, paketID string) (dijawab, ragu, total int) {
	_ = r.pool.QueryRow(ctx,
		`SELECT (SELECT COUNT(*) FROM jawaban WHERE attempt_id = $1 AND jawaban IS NOT NULL),
		        (SELECT COUNT(*) FROM jawaban WHERE attempt_id = $1 AND ragu = TRUE),
		        (SELECT COUNT(*) FROM paket_soal WHERE paket_id = $2)`,
		attemptID, paketID).Scan(&dijawab, &ragu, &total)
	return dijawab, ragu, total
}

// AttemptMonitorCounts menghitung dijawab/ragu/pelanggaran untuk broadcast.
func (r *Repository) AttemptMonitorCounts(ctx context.Context, attemptID string) (dijawab, ragu, pelanggaran int) {
	_ = r.pool.QueryRow(ctx,
		`SELECT (SELECT COUNT(*) FROM jawaban WHERE attempt_id = $1 AND jawaban IS NOT NULL),
		        (SELECT COUNT(*) FROM jawaban WHERE attempt_id = $1 AND ragu = TRUE),
		        (SELECT COUNT(*) FROM pelanggaran WHERE attempt_id = $1)`,
		attemptID).Scan(&dijawab, &ragu, &pelanggaran)
	return dijawab, ragu, pelanggaran
}

// ListHasilJawaban memuat jawaban attempt + soal untuk halaman hasil.
func (r *Repository) ListHasilJawaban(ctx context.Context, attemptID string) ([]hasilJawabanRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT jw.soal_id, jw.jawaban, jw.poin, jw.dikoreksi, jw.ragu,
		        so.tipe, so.pertanyaan, so.jawaban, so.kunci, so.bobot, so.pembahasan
		 FROM jawaban jw JOIN soal so ON so.id = jw.soal_id
		 WHERE jw.attempt_id = $1`, attemptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []hasilJawabanRow{}
	for rows.Next() {
		var h hasilJawabanRow
		var dikoreksi bool
		if err := rows.Scan(&h.SoalID, &h.Jawaban, &h.Poin, &dikoreksi, &h.Ragu,
			&h.Tipe, &h.Pertanyaan, &h.Tampilan, &h.Kunci, &h.Bobot, &h.Pembahasan); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ListSoalHasilPaket memuat soal paket untuk melengkapi hasil (soal
// tanpa baris jawaban).
func (r *Repository) ListSoalHasilPaket(ctx context.Context, paketID string) ([]soalHasilRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT s.id, s.tipe, s.pertanyaan, s.bobot, s.pembahasan
		 FROM paket_soal ps JOIN soal s ON s.id = ps.soal_id
		 WHERE ps.paket_id = $1 ORDER BY ps.urutan ASC`, paketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []soalHasilRow{}
	for rows.Next() {
		var s soalHasilRow
		if err := rows.Scan(&s.ID, &s.Tipe, &s.Pertanyaan, &s.Bobot, &s.Pembahasan); err != nil {
			continue
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// TandaiSelesai menandai attempt selesai + menyimpan nilai akhir.
func (r *Repository) TandaiSelesai(ctx context.Context, attemptID string, nilai float64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE attempt SET status = 'selesai', submit_at = now(), nilai_akhir = $2 WHERE id = $1`,
		attemptID, nilai)
	return err
}

// UpdateNilaiAkhir menyimpan nilai akhir attempt (koreksiJawaban).
func (r *Repository) UpdateNilaiAkhir(ctx context.Context, attemptID string, nilai float64) error {
	_, err := r.pool.Exec(ctx, `UPDATE attempt SET nilai_akhir = $2 WHERE id = $1`, attemptID, nilai)
	return err
}

// ListJawabanNilai memuat baris jawaban untuk hitungNilai.
func (r *Repository) ListJawabanNilai(ctx context.Context, attemptID string) ([]jawabanNilaiRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT COALESCE(jw.poin, 0), so.bobot, so.tipe, jw.dikoreksi,
		        (jw.jawaban IS NOT NULL)
		 FROM jawaban jw
		 JOIN soal so ON so.id = jw.soal_id
		 WHERE jw.attempt_id = $1`, attemptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []jawabanNilaiRow{}
	for rows.Next() {
		var j jawabanNilaiRow
		if err := rows.Scan(&j.Poin, &j.Bobot, &j.Tipe, &j.Dikoreksi, &j.AdaJawaban); err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ListSoalNilaiKosong memuat soal paket yang BELUM punya baris jawaban
// (dihitung sebagai kosong, poin 0).
func (r *Repository) ListSoalNilaiKosong(ctx context.Context, attemptID string) ([]soalNilaiRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT so.bobot, so.tipe
		 FROM paket_soal ps
		 JOIN jadwal_ujian ju ON ju.paket_id = ps.paket_id
		 JOIN attempt a ON a.jadwal_id = ju.id
		 JOIN soal so ON so.id = ps.soal_id
		 WHERE a.id = $1
		   AND NOT EXISTS (SELECT 1 FROM jawaban jw
		                    WHERE jw.attempt_id = a.id AND jw.soal_id = ps.soal_id)`, attemptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []soalNilaiRow{}
	for rows.Next() {
		var s soalNilaiRow
		if err := rows.Scan(&s.Bobot, &s.Tipe); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// SiswaIDByUser memetakan users.id (auth_siswa_id) ke siswa.id.
func (r *Repository) SiswaIDByUser(ctx context.Context, userID string) (string, error) {
	var siswaID string
	err := r.pool.QueryRow(ctx, `SELECT id FROM siswa WHERE user_id = $1`, userID).Scan(&siswaID)
	return siswaID, err
}

// UpsertNilaiAkhir menyalin nilai attempt ke tabel nilai_akhir (Tim 5).
func (r *Repository) UpsertNilaiAkhir(ctx context.Context, attemptID, jadwalID, siswaID string, nilai float64) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO nilai_akhir (attempt_id, jadwal_id, siswa_id, nilai, status)
		 VALUES ($1, $2, $3, $4, 'menunggu_koreksi')
		 ON CONFLICT (attempt_id) DO UPDATE SET nilai = EXCLUDED.nilai, updated_at = now()`,
		attemptID, jadwalID, siswaID, nilai)
	return err
}
