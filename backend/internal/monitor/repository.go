package monitor

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// --- DTO --------------------------------------------------------------------

// PesertaSnapshot adalah satu baris peserta di snapshot monitor.
type PesertaSnapshot struct {
	AttemptID   string   `json:"attempt_id"`
	SiswaID     string   `json:"siswa_id"`
	NamaSiswa   string   `json:"nama_siswa"`
	NISN        string   `json:"nisn"`
	KelasNama   string   `json:"kelas_nama"`
	Status      string   `json:"status"`
	MulaiAt     any      `json:"mulai_at"`
	Deadline    any      `json:"deadline"`
	SisaWaktu   *int     `json:"sisa_waktu_detik"`
	NilaiAkhir  *float64 `json:"nilai_akhir"`
	Dijawab     int      `json:"dijawab"`
	Ragu        int      `json:"ragu"`
	Pelanggaran int      `json:"pelanggaran"`
	TotalSoal   int      `json:"total_soal"`
}

// Snapshot adalah isi pesan snapshot/update monitor.
type Snapshot struct {
	JadwalID  string            `json:"jadwal_id"`
	Generated string            `json:"generated_at"`
	Peserta   []PesertaSnapshot `json:"peserta"`
}

// --- Repository -------------------------------------------------------------

// Repository adalah lapisan akses data paket monitor.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository membuat Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// JadwalExists melaporkan apakah jadwal ada.
func (r *Repository) JadwalExists(ctx context.Context, jadwalID string) (bool, error) {
	var ada bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM jadwal_ujian WHERE id = $1)`, jadwalID).Scan(&ada)
	return ada, err
}

// SnapshotPeserta membaca snapshot peserta satu jadwal dari DB.
// (Dipindah dari monitor.go tanpa perubahan.)
func (r *Repository) SnapshotPeserta(ctx context.Context, jadwalID string) (Snapshot, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT a.id, a.siswa_id, a.status, a.mulai_at, a.deadline,
		        a.sisa_waktu_beku, a.nilai_akhir,
		        COALESCE(s.nama, ''), COALESCE(s.nisn, ''), COALESCE(k.nama, ''),
		        (SELECT COUNT(*) FROM jawaban j
		          WHERE j.attempt_id = a.id AND j.jawaban IS NOT NULL),
		        (SELECT COUNT(*) FROM jawaban j
		          WHERE j.attempt_id = a.id AND j.ragu = TRUE),
		        (SELECT COUNT(*) FROM pelanggaran p WHERE p.attempt_id = a.id),
		        (SELECT COUNT(*) FROM paket_soal ps
		          JOIN jadwal_ujian ju ON ju.paket_id = ps.paket_id
		          WHERE ju.id = a.jadwal_id)
		 FROM attempt a
		 -- attempt.siswa_id merujuk users(id) (migrasi 0004), BUKAN
		 -- siswa(id): JOIN ke siswa HARUS lewat siswa.user_id
		 -- (migrasi 0002). JOIN lama (s.id = a.siswa_id) tidak
		 -- pernah cocok sehingga nama/NISN/kelas selalu kosong.
		 LEFT JOIN siswa s ON s.user_id = a.siswa_id
		 LEFT JOIN kelas k ON k.id = s.kelas_id
		 WHERE a.jadwal_id = $1
		 ORDER BY s.nama ASC`, jadwalID)
	if err != nil {
		return Snapshot{}, err
	}
	defer rows.Close()

	snap := Snapshot{JadwalID: jadwalID, Generated: time.Now().Format(time.RFC3339), Peserta: []PesertaSnapshot{}}
	for rows.Next() {
		var p PesertaSnapshot
		var sisaBeku *int
		var nilai *float64
		var deadline any
		if err := rows.Scan(&p.AttemptID, &p.SiswaID, &p.Status, &p.MulaiAt, &deadline,
			&sisaBeku, &nilai, &p.NamaSiswa, &p.NISN, &p.KelasNama,
			&p.Dijawab, &p.Ragu, &p.Pelanggaran, &p.TotalSoal); err != nil {
			return Snapshot{}, err
		}
		p.Deadline = deadline
		p.NilaiAkhir = nilai
		p.SisaWaktu = hitungSisa(deadline, sisaBeku)
		snap.Peserta = append(snap.Peserta, p)
	}
	return snap, rows.Err()
}

// hitungSisa menghitung sisa waktu (detik) dari deadline.
// Bila sisa dibekukan (pause), kembalikan nilai beku itu.
// (Dipindah dari monitor.go tanpa perubahan.)
func hitungSisa(deadline any, sisaBeku *int) *int {
	if sisaBeku != nil {
		v := *sisaBeku
		if v < 0 {
			v = 0
		}
		return &v
	}
	t, ok := deadline.(time.Time)
	if !ok {
		return nil
	}
	sisa := int(time.Until(t).Seconds())
	if sisa < 0 {
		sisa = 0
	}
	return &sisa
}
