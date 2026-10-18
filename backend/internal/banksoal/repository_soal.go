package banksoal

import (
	"context"
	"encoding/json"
)

// --- Model JSON ---

type soalRow struct {
	ID           string          `json:"id"`
	BankID       string          `json:"bank_id"`
	Tipe         string          `json:"tipe"`
	KisiKisiID   *string         `json:"kisi_kisi_id"`
	KompetensiID *string         `json:"kompetensi_id"`
	Stimulus     string          `json:"stimulus"`
	Pertanyaan   string          `json:"pertanyaan"`
	MediaURL     string          `json:"media_url"`
	Jawaban      json.RawMessage `json:"jawaban"`
	Kunci        json.RawMessage `json:"kunci,omitempty"`
	// Pembahasan (Fase 3, kolom soal.pembahasan migrasi 0011):
	// penjelasan guru; HANYA untuk guru/admin via endpoint bank soal.
	// TIDAK PERNAH dikirim ke siswa sebelum submit (endpoint attempt
	// tidak memilih kolom ini; hasil hanya setelah submit + tampil_hasil).
	Pembahasan string `json:"pembahasan"`
	Bobot      int    `json:"bobot"`
	Level      string `json:"level"`
	Status     string `json:"status"`
	CreatedAt  any    `json:"created_at"`
	UpdatedAt  any    `json:"updated_at"`
}

// LoadSoal mengambil satu soal dalam bank; mengembalikan (soal, ditemukan, err).
func (r *Repository) LoadSoal(ctx context.Context, bankID, soalID string) (*soalRow, bool, error) {
	if !isValidUUID(soalID) {
		return nil, false, nil
	}
	var s soalRow
	err := r.pool.QueryRow(ctx,
		`SELECT id, bank_id, tipe, kisi_kisi_id, kompetensi_id, stimulus, pertanyaan,
		        media_url, jawaban, kunci, pembahasan, bobot, level, status, created_at, updated_at
		 FROM soal WHERE id=$1 AND bank_id=$2`, soalID, bankID).
		Scan(&s.ID, &s.BankID, &s.Tipe, &s.KisiKisiID, &s.KompetensiID, &s.Stimulus,
			&s.Pertanyaan, &s.MediaURL, &s.Jawaban, &s.Kunci, &s.Pembahasan, &s.Bobot, &s.Level,
			&s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &s, true, nil
}

// CountSoal menghitung total soal dalam bank sesuai filter.
func (r *Repository) CountSoal(ctx context.Context, bankID string, f SoalFilter) (int, error) {
	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM soal WHERE bank_id=$1
		   AND ($2::text IS NULL OR tipe=$2)
		   AND ($3::text IS NULL OR level=$3)
		   AND ($4::uuid IS NULL OR kisi_kisi_id=$4)
		   AND ($5::text IS NULL OR pertanyaan ILIKE '%' || $5 || '%')`,
		bankID, f.Tipe, f.Level, f.Kisi, f.Search).Scan(&total)
	return total, err
}

// ListSoal memuat halaman soal dalam bank sesuai filter, terurut dibuat.
func (r *Repository) ListSoal(ctx context.Context, bankID string, f SoalFilter, limit, offset int) ([]soalRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, bank_id, tipe, kisi_kisi_id, kompetensi_id, stimulus, pertanyaan,
		        media_url, jawaban, kunci, pembahasan, bobot, level, status, created_at, updated_at
		 FROM soal WHERE bank_id=$1
		   AND ($2::text IS NULL OR tipe=$2)
		   AND ($3::text IS NULL OR level=$3)
		   AND ($4::uuid IS NULL OR kisi_kisi_id=$4)
		   AND ($5::text IS NULL OR pertanyaan ILIKE '%' || $5 || '%')
		 ORDER BY created_at ASC LIMIT $6 OFFSET $7`,
		bankID, f.Tipe, f.Level, f.Kisi, f.Search, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []soalRow{}
	for rows.Next() {
		var s soalRow
		if err := rows.Scan(&s.ID, &s.BankID, &s.Tipe, &s.KisiKisiID, &s.KompetensiID,
			&s.Stimulus, &s.Pertanyaan, &s.MediaURL, &s.Jawaban, &s.Kunci, &s.Pembahasan, &s.Bobot,
			&s.Level, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// InsertSoal menyisipkan soal baru ke bank.
func (r *Repository) InsertSoal(ctx context.Context, bankID string, in soalInput, level, status string) (*soalRow, error) {
	var s soalRow
	err := r.pool.QueryRow(ctx,
		`INSERT INTO soal (bank_id, tipe, kisi_kisi_id, kompetensi_id, stimulus,
		 pertanyaan, media_url, jawaban, kunci, pembahasan, bobot, level, status)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		 RETURNING id, bank_id, tipe, kisi_kisi_id, kompetensi_id, stimulus,
		 pertanyaan, media_url, jawaban, kunci, pembahasan, bobot, level, status, created_at, updated_at`,
		bankID, in.Tipe, strPtr(ptrVal(in.KisiKisiID)), strPtr(ptrVal(in.KompetensiID)),
		in.Stimulus, in.Pertanyaan, in.MediaURL, in.Jawaban, in.Kunci, in.Pembahasan, in.Bobot, level, status).
		Scan(&s.ID, &s.BankID, &s.Tipe, &s.KisiKisiID, &s.KompetensiID, &s.Stimulus,
			&s.Pertanyaan, &s.MediaURL, &s.Jawaban, &s.Kunci, &s.Pembahasan, &s.Bobot, &s.Level,
			&s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// UpdateSoal mengubah soal.
func (r *Repository) UpdateSoal(ctx context.Context, soalID, bankID string, in soalInput, level, status string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE soal SET tipe=$3, kisi_kisi_id=$4, kompetensi_id=$5, stimulus=$6,
		 pertanyaan=$7, media_url=$8, jawaban=$9, kunci=$10, bobot=$11,
		 level=$12, status=$13, pembahasan=$14, updated_at=now() WHERE id=$1 AND bank_id=$2`,
		soalID, bankID, in.Tipe, strPtr(ptrVal(in.KisiKisiID)), strPtr(ptrVal(in.KompetensiID)),
		in.Stimulus, in.Pertanyaan, in.MediaURL, in.Jawaban, in.Kunci, in.Bobot, level, status, in.Pembahasan)
	return err
}

// DeleteSoal menghapus soal.
func (r *Repository) DeleteSoal(ctx context.Context, soalID, bankID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM soal WHERE id=$1 AND bank_id=$2`, soalID, bankID)
	return err
}

// InsertSoalCSV menyisipkan satu baris hasil parsing CSV (status 'draft').
func (r *Repository) InsertSoalCSV(ctx context.Context, bankID string, row csvSoal) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO soal (bank_id, tipe, stimulus, pertanyaan, media_url,
		 jawaban, kunci, pembahasan, bobot, level, status)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'draft')`,
		bankID, row.tipe, row.stimulus, row.pertanyaan, "",
		row.jawaban, row.kunci, row.pembahasan, row.bobot, row.level)
	return err
}
