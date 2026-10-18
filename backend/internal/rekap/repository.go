package rekap

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- DTO --------------------------------------------------------------------

// nilaiRow adalah satu baris nilai akhir (+ identitas siswa).
// (Dipindah dari rekap.go tanpa perubahan.)
type nilaiRow struct {
	ID            string  `json:"id"`
	AttemptID     string  `json:"attempt_id"`
	JadwalID      string  `json:"jadwal_id"`
	SiswaID       string  `json:"siswa_id"`
	Nilai         float64 `json:"nilai"`
	Status        string  `json:"status"`
	Predikat      *string `json:"predikat"`
	NISN          string  `json:"nisn"`
	NamaSiswa     string  `json:"nama"`
	NamaKelas     string  `json:"kelas"`
	JumlahDijawab int     `json:"jumlah_dijawab"`
}

// jadwalInfo adalah info jadwal untuk berita acara.
type jadwalInfo struct {
	ID         string
	MapelNama  string
	KelasNama  string
	Tanggal    any
	JamMulai   any
	JamSelesai any
}

// --- Repository -------------------------------------------------------------

// Repository adalah lapisan akses data paket rekap.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository membuat Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// listNilaiSQL adalah query dasar daftar nilai. (Dipindah verbatim.)
const listNilaiSQL = `
	SELECT na.id, na.attempt_id, na.jadwal_id, na.siswa_id, na.nilai, na.status, na.predikat,
	       s.nisn, s.nama, COALESCE(k.nama, ''), 
	       -- D3: hanya jawaban yang benar-benar diisi (jawaban IS NOT NULL).
	       -- Baris jawaban dengan jawaban NULL dibuat saat siswa menandai
	       -- ragu-ragu saja — bukan jawaban — dan tidak boleh terhitung
	       -- sebagai "Dijawab" (samakan dengan Monitor).
	       (SELECT COUNT(*) FROM jawaban j WHERE j.attempt_id = na.attempt_id AND j.jawaban IS NOT NULL)
	FROM nilai_akhir na
	JOIN siswa s ON s.id = na.siswa_id
	LEFT JOIN kelas k ON k.id = s.kelas_id
	WHERE na.jadwal_id = $1`

// ListNilai membaca nilai akhir satu jadwal, opsional filter siswa dan
// status final. (Dulu: queryNilai.)
func (r *Repository) ListNilai(ctx context.Context, jadwalID string, siswaID *string, onlyFinal bool) ([]nilaiRow, error) {
	query := listNilaiSQL
	args := []any{jadwalID}
	if siswaID != nil {
		query += ` AND na.siswa_id = $2`
		args = append(args, *siswaID)
	}
	if onlyFinal {
		query += ` AND na.status = 'final'`
	}
	query += ` ORDER BY s.nama ASC`
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []nilaiRow{}
	for rows.Next() {
		var r nilaiRow
		var predikat *string
		if err := rows.Scan(&r.ID, &r.AttemptID, &r.JadwalID, &r.SiswaID, &r.Nilai, &r.Status, &predikat,
			&r.NISN, &r.NamaSiswa, &r.NamaKelas, &r.JumlahDijawab); err != nil {
			return nil, err
		}
		r.Predikat = predikat
		out = append(out, r)
	}
	return out, rows.Err()
}

// TampilHasilKeSiswa membaca setting tampil_hasil_ke_siswa (default false).
// (Dulu: tampilHasilKeSiswa.)
func (r *Repository) TampilHasilKeSiswa(ctx context.Context) bool {
	var raw json.RawMessage
	err := r.pool.QueryRow(ctx,
		`SELECT value FROM settings WHERE key = 'tampil_hasil_ke_siswa'`).Scan(&raw)
	if err != nil {
		return false
	}
	var v bool
	if err := json.Unmarshal(raw, &v); err != nil {
		return false
	}
	return v
}

// GetJadwalInfo membaca info jadwal untuk berita acara, dengan fallback
// ke info minimal bila query utama gagal (mis. jadwal tanpa mapel).
// Tidak ditemukan sama sekali -> pgx.ErrNoRows.
func (r *Repository) GetJadwalInfo(ctx context.Context, jadwalID string) (jadwalInfo, error) {
	var ji jadwalInfo
	err := r.pool.QueryRow(ctx,
		// Tanggal/jam diambil sebagai TEXT: pgtype.Time/Date yang di-scan
		// ke `any` terserialisasi jadi objek {"Microseconds":...,"Valid":...}
		// sehingga frontend crash ("Objects are not valid as a React child")
		// saat merender jam_mulai/jam_selesai di BeritaAcaraPage.
		`SELECT j.id, COALESCE(m.nama, ''), COALESCE(k.nama, ''), j.tanggal::text, j.jam_mulai::text, j.jam_selesai::text
		 FROM jadwal_ujian j
		 JOIN paket_ujian p ON p.id = j.paket_id
		 LEFT JOIN mapel m ON m.id = p.mapel_id
		 LEFT JOIN kelas k ON k.id = j.kelas_id
		 WHERE j.id = $1`, jadwalID).
		Scan(&ji.ID, &ji.MapelNama, &ji.KelasNama,
			&ji.Tanggal, &ji.JamMulai, &ji.JamSelesai)
	if err == nil {
		return ji, nil
	}
	// Fallback ke info minimal (jadwal tanpa mapel).
	var fb jadwalInfo
	err2 := r.pool.QueryRow(ctx,
		`SELECT j.id, COALESCE(m.nama, '') FROM jadwal_ujian j
		 JOIN paket_ujian p ON p.id = j.paket_id
		 LEFT JOIN mapel m ON m.id = p.mapel_id WHERE j.id = $1`, jadwalID).
		Scan(&fb.ID, &fb.MapelNama)
	if err2 != nil {
		return jadwalInfo{}, pgx.ErrNoRows
	}
	return fb, nil
}

// HitungKehadiran menghitung total attempt + yang hadir
// (status 'selesai'/'diblokir') pada satu jadwal.
func (r *Repository) HitungKehadiran(ctx context.Context, jadwalID string) (total, hadir int, err error) {
	err = r.pool.QueryRow(ctx,
		`SELECT COUNT(*), COUNT(*) FILTER (WHERE status IN ('selesai','diblokir'))
		 FROM attempt WHERE jadwal_id = $1`, jadwalID).Scan(&total, &hadir)
	return total, hadir, err
}

// --- Transaksi finalisasi & remedial ----------------------------------------

// BeginTx memulai transaksi. (Dipisah agar service bisa memilih pesan
// error "Gagal memulai ..." yang berbeda dari kegagalan di dalamnya.)
func (r *Repository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

// FinalisasiNilaiTx menandai semua nilai jadwal sebagai final + mengisi
// predikat; mengembalikan jumlah baris yang difinalisasi.
func (r *Repository) FinalisasiNilaiTx(ctx context.Context, tx pgx.Tx, jadwalID string) (int, error) {
	rows, err := tx.Query(ctx,
		`UPDATE nilai_akhir SET status = 'final',
			predikat = CASE WHEN nilai < 70 THEN 'Belum Tuntas'
				WHEN nilai < 80 THEN 'Cukup'
				WHEN nilai <= 90 THEN 'Baik'
				ELSE 'Sangat Baik' END,
			updated_at = now()
		 WHERE jadwal_id = $1 RETURNING id, nilai`, jadwalID)
	if err != nil {
		return 0, err
	}
	count := 0
	for rows.Next() {
		var id string
		var nilai float64
		if err := rows.Scan(&id, &nilai); err != nil {
			rows.Close()
			return 0, err
		}
		count++
	}
	rows.Close()
	return count, nil
}

// InsertRemedialTx mencatat satu nilai remedial.
func (r *Repository) InsertRemedialTx(ctx context.Context, tx pgx.Tx, jadwalID, siswaID string, nilaiBaru float64) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO remedial (jadwal_id, siswa_id, nilai) VALUES ($1, $2, $3)`,
		jadwalID, siswaID, nilaiBaru)
	return err
}

// GetNilaiSiswaTx membaca id + nilai akhir siswa pada jadwal.
// Tidak ada -> pgx.ErrNoRows.
func (r *Repository) GetNilaiSiswaTx(ctx context.Context, tx pgx.Tx, jadwalID, siswaID string) (string, float64, error) {
	var nilaiAkhirID string
	var nilaiAwal float64
	err := tx.QueryRow(ctx,
		`SELECT id, nilai FROM nilai_akhir WHERE jadwal_id = $1 AND siswa_id = $2`,
		jadwalID, siswaID).Scan(&nilaiAkhirID, &nilaiAwal)
	return nilaiAkhirID, nilaiAwal, err
}

// UpdateNilaiAkhirTx memperbarui nilai + predikat satu baris nilai_akhir.
func (r *Repository) UpdateNilaiAkhirTx(ctx context.Context, tx pgx.Tx, nilaiAkhirID string, nilai float64, predikat string) error {
	_, err := tx.Exec(ctx,
		`UPDATE nilai_akhir SET nilai = $1, predikat = $2, updated_at = now() WHERE id = $3`,
		nilai, predikat, nilaiAkhirID)
	return err
}
