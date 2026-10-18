package master

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository adalah lapisan akses data paket master (semua SQL ada di sini).
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository membuat Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// uniqueConstraint melaporkan nama constraint bila err adalah pelanggaran
// unique PostgreSQL (23505).
func uniqueConstraint(err error) (string, bool) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return pgErr.ConstraintName, true
	}
	return "", false
}

// conflictMessage memetakan constraint UNIQUE ke pesan Bahasa Indonesia.
func conflictMessage(constraint string) string {
	switch constraint {
	case "jurusan_kode_key":
		return "Kode jurusan sudah dipakai."
	case "mapel_kode_key":
		return "Kode mapel sudah dipakai."
	case "siswa_nisn_key":
		return "NISN sudah terdaftar."
	case "guru_nip_key":
		return "NIP sudah dipakai guru lain."
	case "kompetensi_mapel_id_kode_key":
		return "Kode kompetensi sudah dipakai untuk mapel ini."
	case "kelas_jurusan_id_tingkat_nama_tahun_ajaran_key":
		return "Kelas dengan nama, tingkat, dan tahun ajaran tersebut sudah ada di jurusan ini."
	case "users_username_key":
		return "Username sudah dipakai."
	default:
		return "Data duplikat: nilai unik sudah dipakai."
	}
}

// isFKViolation melaporkan pelanggaran foreign key (23503).
func isFKViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

// countOrZero menghitung COUNT(*); bila tabel rujukan belum ada
// (migrasi tim lain belum di-apply) dianggap 0 agar guard tidak
// menggagalkan operasi. Dipakai HANYA untuk guard hapus lintas tim.
func (r *Repository) countOrZero(ctx context.Context, sql string, args ...any) int64 {
	var n int64
	if err := r.pool.QueryRow(ctx, sql, args...).Scan(&n); err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return 0
		}
		return -1 // penanda error selain tabel tidak ada
	}
	return n
}
