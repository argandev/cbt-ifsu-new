package analisis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Repository -------------------------------------------------------------

// Repository adalah lapisan akses data paket analisis.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository membuat Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// selesaiFilter: attempt selesai = status 'selesai' ATAU punya nilai_akhir final.
// (Dipindah dari analisis.go tanpa perubahan.)
const selesaiFilter = `(a.status = 'selesai' OR EXISTS (SELECT 1 FROM nilai_akhir na2 WHERE na2.attempt_id = a.id AND na2.status = 'final'))`

var errNotFound = fmt.Errorf("jadwal tidak ditemukan")

// LoadDasar memuat soal paket + attempt selesai + jawaban satu jadwal.
// Jadwal tidak ada -> errNotFound. (Dulu: loadDasar.)
func (r *Repository) LoadDasar(ctx context.Context, jadwalID string) (soals []soalInput, attempts []attemptInput, jawabans []jawabanInput, err error) {
	var paketID string
	err = r.pool.QueryRow(ctx,
		`SELECT paket_id FROM jadwal_ujian WHERE id = $1`, jadwalID).Scan(&paketID)
	if err != nil {
		return nil, nil, nil, errNotFound
	}

	rows, err := r.pool.Query(ctx,
		`SELECT s.id, COALESCE(ps.urutan, 0), s.pertanyaan, s.tipe, s.bobot::float8,
		        s.kompetensi_id, COALESCE(k.kode, ''), COALESCE(k.deskripsi, ''), s.jawaban
		 FROM paket_soal ps
		 JOIN soal s ON s.id = ps.soal_id
		 LEFT JOIN kompetensi k ON k.id = s.kompetensi_id
		 WHERE ps.paket_id = $1
		 ORDER BY ps.urutan ASC, s.pertanyaan ASC`, paketID)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var s soalInput
		var jawabanRaw json.RawMessage
		var kompetensiID *string
		if err := rows.Scan(&s.ID, &s.Urutan, &s.Pertanyaan, &s.Tipe, &s.Bobot,
			&kompetensiID, &s.KompetensiKode, &s.KompetensiNama, &jawabanRaw); err != nil {
			return nil, nil, nil, err
		}
		s.KompetensiID = kompetensiID
		s.Opsi = parseOpsiSoal(jawabanRaw)
		soals = append(soals, s)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, nil, err
	}

	rows, err = r.pool.Query(ctx,
		`SELECT a.id, a.siswa_id, na.nilai
		 FROM attempt a LEFT JOIN nilai_akhir na ON na.attempt_id = a.id
		 WHERE a.jadwal_id = $1 AND `+selesaiFilter+`
		 ORDER BY a.id`, jadwalID)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var a attemptInput
		var nilai *float64
		if err := rows.Scan(&a.ID, &a.SiswaID, &nilai); err != nil {
			return nil, nil, nil, err
		}
		a.Nilai = nilai
		attempts = append(attempts, a)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, nil, err
	}

	rows, err = r.pool.Query(ctx,
		`SELECT j.attempt_id, j.soal_id, j.jawaban, j.poin::float8, j.dikoreksi
		 FROM jawaban j JOIN attempt a ON a.id = j.attempt_id
		 WHERE a.jadwal_id = $1 AND `+selesaiFilter, jadwalID)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var j jawabanInput
		var poin *float64
		var jawabanRaw *string
		// jawaban JSONB: baca sebagai teks lalu jadikan RawMessage (NULL -> nil).
		if err := rows.Scan(&j.AttemptID, &j.SoalID, &jawabanRaw, &poin, &j.Dikoreksi); err != nil {
			return nil, nil, nil, err
		}
		if jawabanRaw != nil {
			j.Jawaban = json.RawMessage(*jawabanRaw)
		}
		j.Poin = poin
		jawabans = append(jawabans, j)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, nil, err
	}
	return soals, attempts, jawabans, nil
}

// parseOpsiSoal mengambil {"opsi":[{"label","teks"}, ...]} dari soal.jawaban.
// (Dipindah dari analisis.go tanpa perubahan.)
func parseOpsiSoal(raw json.RawMessage) []opsiSoal {
	var obj struct {
		Opsi []struct {
			Label string `json:"label"`
			Teks  string `json:"teks"`
		} `json:"opsi"`
	}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil
	}
	out := make([]opsiSoal, 0, len(obj.Opsi))
	for _, o := range obj.Opsi {
		out = append(out, opsiSoal{Label: o.Label, Teks: o.Teks})
	}
	return out
}

// LoadSiswa memuat identitas siswa per attempt. (Dulu: loadSiswa.)
func (r *Repository) LoadSiswa(ctx context.Context, jadwalID string) (map[string]siswaInput, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT a.id, a.siswa_id, COALESCE(s.nisn, ''), COALESCE(s.nama, u.nama, ''),
		        COALESCE(k.nama, '')
		 FROM attempt a
		 LEFT JOIN siswa s ON s.id = a.siswa_id OR s.user_id = a.siswa_id
		 LEFT JOIN users u ON u.id = a.siswa_id
		 LEFT JOIN kelas k ON k.id = s.kelas_id
		 WHERE a.jadwal_id = $1 AND `+selesaiFilter, jadwalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]siswaInput{}
	for rows.Next() {
		var info siswaInput
		if err := rows.Scan(&info.AttemptID, &info.SiswaID, &info.NISN, &info.Nama, &info.Kelas); err != nil {
			return nil, err
		}
		if _, ada := out[info.AttemptID]; !ada {
			out[info.AttemptID] = info
		}
	}
	return out, rows.Err()
}
