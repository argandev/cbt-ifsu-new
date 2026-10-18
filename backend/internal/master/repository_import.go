package master

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// ---------------------------------------------------------------------
// Import Siswa (operasi dalam transaksi import)
// ---------------------------------------------------------------------

// SiswaNISNExistsTx melaporkan apakah NISN sudah terdaftar (dalam tx).
func (r *Repository) SiswaNISNExistsTx(ctx context.Context, tx pgx.Tx, nisn string) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM siswa WHERE nisn = $1)`, nisn).Scan(&exists)
	return exists, err
}

// LookupKelasTx meresolusi kelas dari UUID id atau nama kelas (tidak peka
// huruf); mengembalikan (kelasID, jurusanID). Baris tidak ada -> pgx.ErrNoRows.
func (r *Repository) LookupKelasTx(ctx context.Context, tx pgx.Tx, kelasVal string) (kelasID, jurusanID string, err error) {
	lookupSQL := kelasLookupByNama
	if validUUID(kelasVal) {
		lookupSQL = kelasLookupByID
	}
	err = tx.QueryRow(ctx, lookupSQL, kelasVal).Scan(&kelasID, &jurusanID)
	return kelasID, jurusanID, err
}

// InsertSiswaImportTx menyisipkan satu baris siswa hasil import (dalam tx).
func (r *Repository) InsertSiswaImportTx(ctx context.Context, tx pgx.Tx, userID, nisn string, nisArg any, nama, kelasID, jurusanID string) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO siswa (user_id, nisn, nis, nama, kelas_id, jurusan_id, is_active)
		 VALUES ($1, $2, $3, $4, $5, $6, TRUE)`,
		userID, nisn, nisArg, nama, kelasID, jurusanID)
	return err
}
