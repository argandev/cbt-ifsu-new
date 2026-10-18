package exam

import (
	"context"
	"encoding/json"
)

// koreksiItem adalah satu baris daftar koreksi.
type koreksiItem struct {
	JawabanID     string   `json:"jawaban_id"`
	AttemptID     string   `json:"attempt_id"`
	SoalID        string   `json:"soal_id"`
	Jawaban       any      `json:"jawaban"`
	Ragu          bool     `json:"ragu"`
	Poin          *float64 `json:"poin"`
	Dikoreksi     bool     `json:"dikoreksi"`
	UpdatedAt     any      `json:"updated_at"`
	SiswaID       string   `json:"siswa_id"`
	AttemptStatus string   `json:"attempt_status"`
	NamaSiswa     string   `json:"nama_siswa"`
	NISN          string   `json:"nisn"`
	Tipe          string   `json:"tipe"`
	Pertanyaan    string   `json:"pertanyaan"`
	Bobot         int      `json:"bobot"`
}

// ListKoreksi memuat jawaban manual yang belum dikoreksi untuk satu jadwal.
func (r *Repository) ListKoreksi(ctx context.Context, jadwalID string) ([]koreksiItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT jw.id, jw.attempt_id, jw.soal_id, jw.jawaban, jw.ragu,
		        jw.poin, jw.dikoreksi, jw.updated_at,
		        a.siswa_id, a.status,
		        COALESCE(s.nama, ''), COALESCE(s.nisn, ''),
		        so.tipe, so.pertanyaan, so.bobot
		 FROM jawaban jw
		 JOIN attempt a ON a.id = jw.attempt_id
		 JOIN soal so ON so.id = jw.soal_id
		 LEFT JOIN siswa s ON s.user_id = a.siswa_id
		 WHERE a.jadwal_id = $1
		   AND jw.dikoreksi = FALSE
		   AND jw.jawaban IS NOT NULL
		   AND so.tipe IN ('uraian','upload_file','praktik_singkat',
		                    'esai','essay','isian_panjang','isi_panjang',
		                    'esai_singkat','uraian_singkat','uraian_esai')
		 ORDER BY s.nama ASC, jw.updated_at ASC`, jadwalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []koreksiItem{}
	for rows.Next() {
		var it koreksiItem
		var jwb []byte
		if err := rows.Scan(&it.JawabanID, &it.AttemptID, &it.SoalID, &jwb, &it.Ragu,
			&it.Poin, &it.Dikoreksi, &it.UpdatedAt, &it.SiswaID, &it.AttemptStatus,
			&it.NamaSiswa, &it.NISN, &it.Tipe, &it.Pertanyaan, &it.Bobot); err != nil {
			return nil, err
		}
		var jwbAny any
		if len(jwb) > 0 {
			_ = json.Unmarshal(jwb, &jwbAny)
		}
		it.Jawaban = jwbAny
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// GetJawabanKoreksi mengambil attempt/soal/bobot untuk satu jawaban.
func (r *Repository) GetJawabanKoreksi(ctx context.Context, jawabanID string) (attemptID, soalID string, bobot float64, err error) {
	err = r.pool.QueryRow(ctx,
		`SELECT jw.attempt_id, jw.soal_id, so.bobot
		 FROM jawaban jw JOIN soal so ON so.id = jw.soal_id
		 WHERE jw.id = $1`, jawabanID).Scan(&attemptID, &soalID, &bobot)
	return attemptID, soalID, bobot, err
}

// UpdateJawabanPoin menyimpan poin koreksi manual. q bisa pool atau tx.
func (r *Repository) UpdateJawabanPoin(ctx context.Context, q querier, jawabanID string, poin float64) error {
	_, err := q.Exec(ctx,
		`UPDATE jawaban SET poin = $2, dikoreksi = TRUE, updated_at = now() WHERE id = $1`,
		jawabanID, poin)
	return err
}

// AttemptStatusJadwal membaca status + jadwal_id attempt. q bisa pool atau tx.
func (r *Repository) AttemptStatusJadwal(ctx context.Context, q querier, attemptID string) (status, jadwalID string) {
	_ = q.QueryRow(ctx, `SELECT status, jadwal_id FROM attempt WHERE id = $1`, attemptID).
		Scan(&status, &jadwalID)
	return status, jadwalID
}
