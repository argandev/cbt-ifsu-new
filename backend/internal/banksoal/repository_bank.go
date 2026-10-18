package banksoal

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// --- Model JSON ---

type bankRow struct {
	ID           string  `json:"id"`
	Kode         string  `json:"kode"`
	Nama         string  `json:"nama"`
	MapelID      string  `json:"mapel_id"`
	KelasTingkat int     `json:"kelas_tingkat"`
	GuruID       *string `json:"guru_id"`
	IsLocked     bool    `json:"is_locked"`
	JumlahSoal   int     `json:"jumlah_soal"`
	CreatedAt    any     `json:"created_at"`
	UpdatedAt    any     `json:"updated_at"`
}

// --- Repository -------------------------------------------------------------

// Repository adalah lapisan akses data paket banksoal (semua SQL ada di sini).
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository membuat Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// loadBank mengambil bank; mengembalikan (bank, ditemukan, err).
// Shim kompatibilitas untuk jalur import xlsx/docx (import_excel.go)
// yang memanggil dengan echo.Context + pool; implementasi aslinya ada
// di (*Repository).LoadBank.
func loadBank(c *echo.Context, pool *pgxpool.Pool, id string) (*bankRow, bool, error) {
	return NewRepository(pool).LoadBank(c.Request().Context(), id)
}

// LoadBank mengambil bank; mengembalikan (bank, ditemukan, err).
func (r *Repository) LoadBank(ctx context.Context, id string) (*bankRow, bool, error) {
	if !isValidUUID(id) {
		return nil, false, nil
	}
	var b bankRow
	err := r.pool.QueryRow(ctx,
		`SELECT b.id, b.kode, b.nama, b.mapel_id, b.kelas_tingkat, b.guru_id,
		        b.is_locked, b.created_at, b.updated_at,
		        (SELECT COUNT(*) FROM soal s WHERE s.bank_id = b.id)
		 FROM bank_soal b WHERE b.id = $1`, id).
		Scan(&b.ID, &b.Kode, &b.Nama, &b.MapelID, &b.KelasTingkat, &b.GuruID,
			&b.IsLocked, &b.CreatedAt, &b.UpdatedAt, &b.JumlahSoal)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &b, true, nil
}

// BankDipakaiPaket: true bila ada soal dari bank ini yang tersusun di
// paket manapun (via paket_soal JOIN soal). Kueri langsung ke tabel
// kanonis 0004 — tanpa workaround information_schema.
func (r *Repository) BankDipakaiPaket(ctx context.Context, bankID string) (bool, error) {
	var n int
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM paket_soal ps
		 JOIN soal s ON s.id = ps.soal_id
		 WHERE s.bank_id = $1`, bankID).Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}

// CountBank menghitung total bank soal sesuai filter.
func (r *Repository) CountBank(ctx context.Context, f BankFilter) (int, error) {
	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM bank_soal b
		 WHERE ($1::text IS NULL OR b.kode ILIKE '%' || $1 || '%' OR b.nama ILIKE '%' || $1 || '%')
		   AND ($2::uuid IS NULL OR b.mapel_id = $2)
		   AND ($3::int IS NULL OR b.kelas_tingkat = $3)
		   AND ($4::boolean IS NULL OR b.is_locked = $4)`,
		f.Search, f.MapelID, f.KelasTingkat, f.Locked).Scan(&total)
	return total, err
}

// ListBank memuat halaman bank soal sesuai filter, terurut terbaru.
func (r *Repository) ListBank(ctx context.Context, f BankFilter, limit, offset int) ([]bankRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT b.id, b.kode, b.nama, b.mapel_id, b.kelas_tingkat, b.guru_id,
		        b.is_locked, b.created_at, b.updated_at,
		        (SELECT COUNT(*) FROM soal s WHERE s.bank_id = b.id)
		 FROM bank_soal b
		 WHERE ($1::text IS NULL OR b.kode ILIKE '%' || $1 || '%' OR b.nama ILIKE '%' || $1 || '%')
		   AND ($2::uuid IS NULL OR b.mapel_id = $2)
		   AND ($3::int IS NULL OR b.kelas_tingkat = $3)
		   AND ($4::boolean IS NULL OR b.is_locked = $4)
		 ORDER BY b.created_at DESC LIMIT $5 OFFSET $6`,
		f.Search, f.MapelID, f.KelasTingkat, f.Locked, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []bankRow{}
	for rows.Next() {
		var b bankRow
		if err := rows.Scan(&b.ID, &b.Kode, &b.Nama, &b.MapelID, &b.KelasTingkat,
			&b.GuruID, &b.IsLocked, &b.CreatedAt, &b.UpdatedAt, &b.JumlahSoal); err != nil {
			return nil, err
		}
		items = append(items, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// InsertBank menyisipkan bank soal baru (guruID nil = tanpa pemilik).
func (r *Repository) InsertBank(ctx context.Context, kode, nama, mapelID string, kelasTingkat int, guruID any) (*bankRow, error) {
	var b bankRow
	err := r.pool.QueryRow(ctx,
		`INSERT INTO bank_soal (kode, nama, mapel_id, kelas_tingkat, guru_id)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, kode, nama, mapel_id, kelas_tingkat, guru_id, is_locked, created_at, updated_at, 0`,
		kode, nama, mapelID, kelasTingkat, guruID).
		Scan(&b.ID, &b.Kode, &b.Nama, &b.MapelID, &b.KelasTingkat, &b.GuruID,
			&b.IsLocked, &b.CreatedAt, &b.UpdatedAt, &b.JumlahSoal)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// UpdateBank mengubah bank soal (guruID nil = lepas kepemilikan).
func (r *Repository) UpdateBank(ctx context.Context, id, kode, nama, mapelID string, kelasTingkat int, guruID *string) error {
	var guruArg any
	if guruID != nil {
		guruArg = *guruID
	}
	_, err := r.pool.Exec(ctx,
		`UPDATE bank_soal SET kode=$2, nama=$3, mapel_id=$4, kelas_tingkat=$5,
		 guru_id=$6, updated_at=now() WHERE id=$1`,
		id, kode, nama, mapelID, kelasTingkat, guruArg)
	return err
}

// DeleteBank menghapus bank soal.
func (r *Repository) DeleteBank(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM bank_soal WHERE id=$1`, id)
	return err
}

// SetBankLock mengunci/membuka bank soal.
func (r *Repository) SetBankLock(ctx context.Context, id string, locked bool) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE bank_soal SET is_locked=$2, updated_at=now() WHERE id=$1`, id, locked)
	return err
}
