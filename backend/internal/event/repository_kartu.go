package event

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Model kartu ---

const catatanPasswordAwal = "Password awal: NISN (wajib diganti saat login pertama)"

type kartuRow struct {
	JadwalID      string  `json:"jadwal_id"`
	SiswaID       string  `json:"siswa_id"`
	NomorPeserta  string  `json:"nomor_peserta"`
	NamaSekolah   string  `json:"nama_sekolah"`
	NamaEvent     *string `json:"nama_event"`
	NamaSiswa     string  `json:"nama_siswa"`
	NISN          string  `json:"nisn"`
	Kelas         string  `json:"kelas"`
	Jurusan       string  `json:"jurusan"`
	Mapel         string  `json:"mapel"`
	Paket         string  `json:"paket"`
	Tanggal       string  `json:"tanggal"`
	JamMulai      string  `json:"jam_mulai"`
	JamSelesai    string  `json:"jam_selesai"`
	Ruangan       *string `json:"ruangan"`
	RuanganLokasi *string `json:"ruangan_lokasi"`
	Username      string  `json:"username"`
	Catatan       string  `json:"catatan"`
}

// jadwalKartu adalah data jadwal + konteksnya untuk kartu peserta.
type jadwalKartu struct {
	ID            string
	KelasID       string
	Tanggal       string // YYYY-MM-DD
	JamMulai      string // HH:MM
	JamSelesai    string // HH:MM
	Status        string
	PaketNama     string
	PaketJenis    string
	MapelNama     string
	KelasNama     string
	JurusanNama   string
	EventID       *string
	EventKode     *string
	EventNama     *string
	RuanganID     *string
	RuanganNama   *string
	RuanganLokasi *string
}

type siswaKartu struct {
	ID      string
	NISN    string
	Nama    string
	KelasID string
	UserID  *string
}

// --- Load helpers kartu (dipakai service + kartu_pdf.go) ---

func loadJadwalKartu(ctx context.Context, pool *pgxpool.Pool, jadwalID string) (*jadwalKartu, bool, error) {
	if !isValidUUID(jadwalID) {
		return nil, false, nil
	}
	var j jadwalKartu
	err := pool.QueryRow(ctx,
		`SELECT j.id, j.kelas_id, to_char(j.tanggal, 'YYYY-MM-DD'),
		        to_char(j.jam_mulai, 'HH24:MI'), to_char(j.jam_selesai, 'HH24:MI'), j.status,
		        p.nama, p.jenis, m.nama, k.nama, ju.nama,
		        j.event_id, e.kode, e.nama, j.ruangan_id, r.nama, r.lokasi
		 FROM jadwal_ujian j
		 JOIN paket_ujian p ON p.id = j.paket_id
		 JOIN mapel m ON m.id = p.mapel_id
		 JOIN kelas k ON k.id = j.kelas_id
		 JOIN jurusan ju ON ju.id = k.jurusan_id
		 LEFT JOIN event e ON e.id = j.event_id
		 LEFT JOIN ruangan r ON r.id = j.ruangan_id
		 WHERE j.id = $1`, jadwalID).
		Scan(&j.ID, &j.KelasID, &j.Tanggal, &j.JamMulai, &j.JamSelesai, &j.Status,
			&j.PaketNama, &j.PaketJenis, &j.MapelNama, &j.KelasNama, &j.JurusanNama,
			&j.EventID, &j.EventKode, &j.EventNama, &j.RuanganID, &j.RuanganNama, &j.RuanganLokasi)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &j, true, nil
}

func loadNamaSekolah(ctx context.Context, pool *pgxpool.Pool) string {
	var raw json.RawMessage
	if err := pool.QueryRow(ctx,
		`SELECT value FROM settings WHERE key = 'nama_sekolah'`).Scan(&raw); err != nil {
		return ""
	}
	var nama string
	if err := json.Unmarshal(raw, &nama); err != nil {
		return ""
	}
	return nama
}

// buildKartuList menyusun kartu semua siswa aktif di kelas jadwal,
// urut nama A-Z. Nomor peserta unik per jadwal.
func buildKartuList(ctx context.Context, pool *pgxpool.Pool, j *jadwalKartu) ([]kartuRow, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, nisn, nama, kelas_id, user_id FROM siswa
		 WHERE kelas_id = $1 AND is_active = TRUE ORDER BY nama ASC`, j.KelasID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	namaSekolah := loadNamaSekolah(ctx, pool)
	kode := "US"
	if j.EventKode != nil && strings.TrimSpace(*j.EventKode) != "" {
		kode = *j.EventKode
	}
	tahun := ""
	if len(j.Tanggal) >= 4 {
		tahun = j.Tanggal[:4]
	}
	items := []kartuRow{}
	for rows.Next() {
		var s siswaKartu
		if err := rows.Scan(&s.ID, &s.NISN, &s.Nama, &s.KelasID, &s.UserID); err != nil {
			return nil, err
		}
		nomor := fmt.Sprintf("%s/%s/%03d", kode, tahun, len(items)+1)
		items = append(items, kartuRow{
			JadwalID:      j.ID,
			SiswaID:       s.ID,
			NomorPeserta:  nomor,
			NamaSekolah:   namaSekolah,
			NamaEvent:     j.EventNama,
			NamaSiswa:     s.Nama,
			NISN:          s.NISN,
			Kelas:         j.KelasNama,
			Jurusan:       j.JurusanNama,
			Mapel:         j.MapelNama,
			Paket:         j.PaketNama,
			Tanggal:       j.Tanggal,
			JamMulai:      j.JamMulai,
			JamSelesai:    j.JamSelesai,
			Ruangan:       j.RuanganNama,
			RuanganLokasi: j.RuanganLokasi,
			Username:      s.NISN,
			Catatan:       catatanPasswordAwal,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func loadSiswaKartu(ctx context.Context, pool *pgxpool.Pool, siswaID string) (*siswaKartu, bool, error) {
	if !isValidUUID(siswaID) {
		return nil, false, nil
	}
	// Cari siswa.id dulu; bila tidak ketemu, fallback ke siswa.user_id.
	// Frontend KartuSayaPage mengirim siswaId = user.id (ID akun login),
	// bukan siswa.id — tanpa fallback, kartu sendiri selalu 404.
	// Pola sama seperti kelasSiswa di exam.go.
	var s siswaKartu
	err := pool.QueryRow(ctx,
		`SELECT id, nisn, nama, kelas_id, user_id FROM siswa WHERE id = $1`, siswaID).
		Scan(&s.ID, &s.NISN, &s.Nama, &s.KelasID, &s.UserID)
	if err == nil {
		return &s, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, err
	}
	err = pool.QueryRow(ctx,
		`SELECT id, nisn, nama, kelas_id, user_id FROM siswa WHERE user_id = $1`, siswaID).
		Scan(&s.ID, &s.NISN, &s.Nama, &s.KelasID, &s.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &s, true, nil
}
