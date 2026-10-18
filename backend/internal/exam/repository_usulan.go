package exam

import (
	"context"
)

// usulanItem adalah satu baris usulan unblock.
type usulanItem struct {
	ID           string  `json:"id"`
	AttemptID    string  `json:"attempt_id"`
	PengawasID   string  `json:"pengawas_id"`
	Alasan       string  `json:"alasan"`
	Status       string  `json:"status"`
	DiprosesOleh *string `json:"diproses_oleh"`
	DiprosesAt   any     `json:"diproses_at"`
	CreatedAt    any     `json:"created_at"`
	NamaSiswa    string  `json:"nama_siswa"`
}

// ListUsulanUnblock memuat usulan unblock satu jadwal (terbaru dulu).
func (r *Repository) ListUsulanUnblock(ctx context.Context, jadwalID string) ([]usulanItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT u.id, u.attempt_id, u.pengawas_id, u.alasan, u.status,
		        u.diproses_oleh, u.diproses_at, u.created_at,
		        COALESCE(s.nama, '')
		 FROM usulan_unblock u
		 JOIN attempt a ON a.id = u.attempt_id
		 LEFT JOIN siswa s ON s.user_id = a.siswa_id
		 WHERE a.jadwal_id = $1 ORDER BY u.created_at DESC`, jadwalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []usulanItem{}
	for rows.Next() {
		var it usulanItem
		if err := rows.Scan(&it.ID, &it.AttemptID, &it.PengawasID, &it.Alasan, &it.Status,
			&it.DiprosesOleh, &it.DiprosesAt, &it.CreatedAt, &it.NamaSiswa); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// CountUsulanMenunggu menghitung usulan berstatus menunggu untuk attempt.
func (r *Repository) CountUsulanMenunggu(ctx context.Context, attemptID string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM usulan_unblock WHERE attempt_id = $1 AND status = 'menunggu'`,
		attemptID).Scan(&n)
	return n, err
}

// InsertUsulan mencatat usulan unblock baru (status 'menunggu').
func (r *Repository) InsertUsulan(ctx context.Context, attemptID, pengawasID, alasan string) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO usulan_unblock (attempt_id, pengawas_id, alasan, status)
		 VALUES ($1, $2, $3, 'menunggu') RETURNING id`,
		attemptID, pengawasID, alasan).Scan(&id)
	return id, err
}

// LanjutkanAttempt meng-unblock attempt: LANJUT sisa waktu (jawaban
// tersimpan, TIDAK reset). q bisa pool atau tx.
func (r *Repository) LanjutkanAttempt(ctx context.Context, q querier, attemptID string, sisa int) error {
	_, err := q.Exec(ctx,
		`UPDATE attempt SET status = 'berjalan',
		        deadline = now() + make_interval(secs => $2),
		        sisa_waktu_beku = NULL WHERE id = $1`, attemptID, sisa)
	return err
}

// UsulanMenungguTerbaru mengambil usulan menunggu terbaru untuk attempt.
// q bisa pool atau tx.
func (r *Repository) UsulanMenungguTerbaru(ctx context.Context, q querier, attemptID string) (string, error) {
	var usulanID string
	err := q.QueryRow(ctx,
		`SELECT id FROM usulan_unblock WHERE attempt_id = $1 AND status = 'menunggu'
		 ORDER BY created_at DESC LIMIT 1`, attemptID).Scan(&usulanID)
	return usulanID, err
}

// SetujuiUsulan menandai usulan disetujui + pencatatnya. q bisa pool atau tx.
func (r *Repository) SetujuiUsulan(ctx context.Context, q querier, usulanID, uid string) {
	_, _ = q.Exec(ctx,
		`UPDATE usulan_unblock SET status = 'disetujui', diproses_oleh = $2, diproses_at = now()
		 WHERE id = $1`, usulanID, uid)
}

// InsertUsulanDisetujui mencatat usulan terproses (proktor unblock langsung
// tanpa usulan). q bisa pool atau tx.
func (r *Repository) InsertUsulanDisetujui(ctx context.Context, q querier, attemptID, uid, alasan string) {
	_, _ = q.Exec(ctx,
		`INSERT INTO usulan_unblock (attempt_id, pengawas_id, alasan, status, diproses_oleh, diproses_at)
		 VALUES ($1, $2, $3, 'disetujui', $2, now())`, attemptID, uid, alasan)
}
