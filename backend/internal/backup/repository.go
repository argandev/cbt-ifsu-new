package backup

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// --- DTO --------------------------------------------------------------------

// backupMeta adalah satu baris backup_meta (+ nama pembuat untuk list).
// (Dipindah dari backup.go tanpa perubahan.)
type backupMeta struct {
	ID          string    `json:"id"`
	NamaFile    string    `json:"nama_file"`
	UkuranBytes int64     `json:"ukuran_bytes"`
	Jenis       string    `json:"jenis"`
	DibuatOleh  *string   `json:"dibuat_oleh"`
	PembuatNama *string   `json:"pembuat_nama,omitempty"`
	Keterangan  *string   `json:"keterangan"`
	CreatedAt   time.Time `json:"created_at"`
}

// --- Repository -------------------------------------------------------------

// Repository adalah lapisan akses data paket backup.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository membuat Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func strOrNil(s *string) any {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil
	}
	return *s
}

// GetMeta membaca satu baris backup_meta. Tidak ditemukan -> (nil, nil).
// (Dulu: getMeta.)
func (r *Repository) GetMeta(ctx context.Context, id string) (*backupMeta, error) {
	var m backupMeta
	err := r.pool.QueryRow(ctx,
		`SELECT bm.id, bm.nama_file, bm.ukuran_bytes, bm.jenis, bm.dibuat_oleh,
		        bm.keterangan, bm.created_at, u.nama
		 FROM backup_meta bm
		 LEFT JOIN users u ON u.id = bm.dibuat_oleh
		 WHERE bm.id = $1`, id).
		Scan(&m.ID, &m.NamaFile, &m.UkuranBytes, &m.Jenis, &m.DibuatOleh,
			&m.Keterangan, &m.CreatedAt, &m.PembuatNama)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

// InsertMeta mencatat satu baris backup_meta. (Dulu: insertMeta.)
func (r *Repository) InsertMeta(ctx context.Context, m *backupMeta) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO backup_meta (id, nama_file, ukuran_bytes, jenis, dibuat_oleh, keterangan, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		m.ID, m.NamaFile, m.UkuranBytes, m.Jenis, strOrNil(m.DibuatOleh),
		strOrNil(m.Keterangan), m.CreatedAt)
	return err
}

// UpsertMeta mengembalikan baris meta setelah restore (lihat catatan
// paket: pg_restore --clean menghapus baris yang ditulis sebelum
// restore). ON CONFLICT DO NOTHING: bila dump sudah berisi baris yang
// sama (mis. dump dibuat setelah meta itu tercatat), jangan duplikat.
// (Dulu: upsertMeta.)
func (r *Repository) UpsertMeta(ctx context.Context, m *backupMeta) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO backup_meta (id, nama_file, ukuran_bytes, jenis, dibuat_oleh, keterangan, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (id) DO NOTHING`,
		m.ID, m.NamaFile, m.UkuranBytes, m.Jenis, strOrNil(m.DibuatOleh),
		strOrNil(m.Keterangan), m.CreatedAt)
	return err
}

// ListMeta membaca halaman metadata backup, terbaru dulu. (Dulu: query
// inline di handler GET /api/v1/backup.)
func (r *Repository) ListMeta(ctx context.Context, limit, offset int) ([]backupMeta, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT bm.id, bm.nama_file, bm.ukuran_bytes, bm.jenis, bm.dibuat_oleh,
		        bm.keterangan, bm.created_at, u.nama
		 FROM backup_meta bm
		 LEFT JOIN users u ON u.id = bm.dibuat_oleh
		 ORDER BY bm.created_at DESC
		 LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []backupMeta{}
	for rows.Next() {
		var m backupMeta
		if err := rows.Scan(&m.ID, &m.NamaFile, &m.UkuranBytes, &m.Jenis,
			&m.DibuatOleh, &m.Keterangan, &m.CreatedAt, &m.PembuatNama); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// CountMeta menghitung total baris backup_meta.
func (r *Repository) CountMeta(ctx context.Context) (int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM backup_meta`).Scan(&total)
	return total, err
}

// DeleteMeta menghapus satu baris backup_meta.
func (r *Repository) DeleteMeta(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM backup_meta WHERE id = $1`, id)
	return err
}

// ExamInProgress memeriksa gerbang restore: jadwal aktif/pause atau
// attempt berjalan. Mengembalikan pesan penolakan bila ada yang
// menghalangi, string kosong bila aman. (Dulu: examInProgress.)
func (r *Repository) ExamInProgress(ctx context.Context) (string, error) {
	var jadwalAktif int
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM jadwal_ujian WHERE status IN ('aktif', 'pause')`).
		Scan(&jadwalAktif); err != nil {
		return "", err
	}
	if jadwalAktif > 0 {
		return fmt.Sprintf("Restore ditolak: ada %d jadwal ujian berstatus aktif/pause. "+
			"Selesaikan atau ubah status jadwal tersebut terlebih dahulu, lalu ulangi restore.", jadwalAktif), nil
	}
	var attemptBerjalan int
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM attempt WHERE status = 'berjalan'`).
		Scan(&attemptBerjalan); err != nil {
		return "", err
	}
	if attemptBerjalan > 0 {
		return fmt.Sprintf("Restore ditolak: ada %d attempt ujian berstatus berjalan. "+
			"Tunggu sampai semua peserta selesai / diblokir, lalu ulangi restore.", attemptBerjalan), nil
	}
	return "", nil
}
