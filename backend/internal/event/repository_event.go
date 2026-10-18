package event

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Model JSON ---

type eventRow struct {
	ID             string  `json:"id"`
	Kode           string  `json:"kode"`
	Nama           string  `json:"nama"`
	Jenis          string  `json:"jenis"`
	TanggalMulai   *string `json:"tanggal_mulai"`
	TanggalSelesai *string `json:"tanggal_selesai"`
	Deskripsi      *string `json:"deskripsi"`
	CreatedAt      any     `json:"created_at"`
}

type ruanganRow struct {
	ID         string  `json:"id"`
	Kode       string  `json:"kode"`
	Nama       string  `json:"nama"`
	Kapasitas  int     `json:"kapasitas"`
	Lokasi     *string `json:"lokasi"`
	Keterangan *string `json:"keterangan"`
	CreatedAt  any     `json:"created_at"`
}

// --- Repository -------------------------------------------------------------

// Repository adalah lapisan akses data paket event (semua SQL ada di sini).
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository membuat Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func isUniqueViolation(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "unique"))
}

// --- Load helpers (dipakai service + kartu_pdf.go) ---

func scanEvent(row pgx.Row) (*eventRow, error) {
	var e eventRow
	err := row.Scan(&e.ID, &e.Kode, &e.Nama, &e.Jenis, &e.TanggalMulai,
		&e.TanggalSelesai, &e.Deskripsi, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func loadEvent(ctx context.Context, pool *pgxpool.Pool, id string) (*eventRow, bool, error) {
	if !isValidUUID(id) {
		return nil, false, nil
	}
	e, err := scanEvent(pool.QueryRow(ctx,
		`SELECT id, kode, nama, jenis, tanggal_mulai::text, tanggal_selesai::text,
		        deskripsi, created_at FROM event WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return e, true, nil
}

func scanRuangan(row pgx.Row) (*ruanganRow, error) {
	var r ruanganRow
	err := row.Scan(&r.ID, &r.Kode, &r.Nama, &r.Kapasitas, &r.Lokasi, &r.Keterangan, &r.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func loadRuangan(ctx context.Context, pool *pgxpool.Pool, id string) (*ruanganRow, bool, error) {
	if !isValidUUID(id) {
		return nil, false, nil
	}
	r, err := scanRuangan(pool.QueryRow(ctx,
		`SELECT id, kode, nama, kapasitas, lokasi, keterangan, created_at
		 FROM ruangan WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return r, true, nil
}

// kodeDipakai: true bila kode sudah dipakai baris lain (excludeID boleh "").
func kodeDipakai(ctx context.Context, pool *pgxpool.Pool, tabel, kode, excludeID string) (bool, error) {
	var n int
	q := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE kode = $1`, tabel)
	args := []any{kode}
	if excludeID != "" {
		q += ` AND id <> $2`
		args = append(args, excludeID)
	}
	if err := pool.QueryRow(ctx, q, args...).Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}

func jadwalMemakai(ctx context.Context, pool *pgxpool.Pool, kolom, id string) (int, error) {
	var n int
	q := fmt.Sprintf(`SELECT COUNT(*) FROM jadwal_ujian WHERE %s = $1`, kolom)
	if err := pool.QueryRow(ctx, q, id).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// --- Event: list/count/insert/update/delete ---

// CountEvent menghitung total event sesuai filter.
func (r *Repository) CountEvent(ctx context.Context, jenis, search any) (int, error) {
	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM event
		 WHERE ($1::text IS NULL OR jenis = $1)
		   AND ($2::text IS NULL OR nama ILIKE '%' || $2 || '%' OR kode ILIKE '%' || $2 || '%')`,
		jenis, search).Scan(&total)
	return total, err
}

// ListEvent memuat halaman event sesuai filter.
func (r *Repository) ListEvent(ctx context.Context, jenis, search any, limit, offset int) ([]eventRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, kode, nama, jenis, tanggal_mulai::text, tanggal_selesai::text,
		        deskripsi, created_at FROM event
		 WHERE ($1::text IS NULL OR jenis = $1)
		   AND ($2::text IS NULL OR nama ILIKE '%' || $2 || '%' OR kode ILIKE '%' || $2 || '%')
		 ORDER BY tanggal_mulai DESC NULLS LAST, nama ASC
		 LIMIT $3 OFFSET $4`,
		jenis, search, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []eventRow{}
	for rows.Next() {
		var ev eventRow
		if err := rows.Scan(&ev.ID, &ev.Kode, &ev.Nama, &ev.Jenis, &ev.TanggalMulai,
			&ev.TanggalSelesai, &ev.Deskripsi, &ev.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, ev)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// InsertEvent menyisipkan event baru.
func (r *Repository) InsertEvent(ctx context.Context, in eventInput, mulai, selesai any) (*eventRow, error) {
	return scanEvent(r.pool.QueryRow(ctx,
		`INSERT INTO event (kode, nama, jenis, tanggal_mulai, tanggal_selesai, deskripsi)
		 VALUES ($1, $2, $3, $4::date, $5::date, $6)
		 RETURNING id, kode, nama, jenis, tanggal_mulai::text, tanggal_selesai::text,
		           deskripsi, created_at`,
		in.Kode, in.Nama, in.Jenis, mulai, selesai, optionalText(in.Deskripsi)))
}

// UpdateEvent mengubah event.
func (r *Repository) UpdateEvent(ctx context.Context, id string, in eventInput, mulai, selesai any) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE event SET kode = $2, nama = $3, jenis = $4,
		 tanggal_mulai = $5::date, tanggal_selesai = $6::date, deskripsi = $7
		 WHERE id = $1`,
		id, in.Kode, in.Nama, in.Jenis, mulai, selesai, optionalText(in.Deskripsi))
	return err
}

// DeleteEvent menghapus event.
func (r *Repository) DeleteEvent(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM event WHERE id = $1`, id)
	return err
}

// --- Ruangan: list/count/insert/update/delete ---

// CountRuangan menghitung total ruangan sesuai pencarian.
func (r *Repository) CountRuangan(ctx context.Context, search any) (int, error) {
	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM ruangan
		 WHERE ($1::text IS NULL OR nama ILIKE '%' || $1 || '%' OR kode ILIKE '%' || $1 || '%'
		        OR COALESCE(lokasi, '') ILIKE '%' || $1 || '%')`,
		search).Scan(&total)
	return total, err
}

// ListRuangan memuat halaman ruangan, terurut nama.
func (r *Repository) ListRuangan(ctx context.Context, search any, limit, offset int) ([]ruanganRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, kode, nama, kapasitas, lokasi, keterangan, created_at FROM ruangan
		 WHERE ($1::text IS NULL OR nama ILIKE '%' || $1 || '%' OR kode ILIKE '%' || $1 || '%'
		        OR COALESCE(lokasi, '') ILIKE '%' || $1 || '%')
		 ORDER BY nama ASC LIMIT $2 OFFSET $3`,
		search, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ruanganRow{}
	for rows.Next() {
		var ru ruanganRow
		if err := rows.Scan(&ru.ID, &ru.Kode, &ru.Nama, &ru.Kapasitas, &ru.Lokasi, &ru.Keterangan, &ru.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, ru)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// InsertRuangan menyisipkan ruangan baru.
func (r *Repository) InsertRuangan(ctx context.Context, in ruanganInput) (*ruanganRow, error) {
	return scanRuangan(r.pool.QueryRow(ctx,
		`INSERT INTO ruangan (kode, nama, kapasitas, lokasi, keterangan)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, kode, nama, kapasitas, lokasi, keterangan, created_at`,
		in.Kode, in.Nama, in.Kapasitas, optionalText(in.Lokasi), optionalText(in.Keterangan)))
}

// UpdateRuangan mengubah ruangan.
func (r *Repository) UpdateRuangan(ctx context.Context, id string, in ruanganInput) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE ruangan SET kode = $2, nama = $3, kapasitas = $4, lokasi = $5, keterangan = $6
		 WHERE id = $1`,
		id, in.Kode, in.Nama, in.Kapasitas, optionalText(in.Lokasi), optionalText(in.Keterangan))
	return err
}

// DeleteRuangan menghapus ruangan.
func (r *Repository) DeleteRuangan(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM ruangan WHERE id = $1`, id)
	return err
}

// RuanganJadwalPenuh mengembalikan "nama kelas (tanggal)" jadwal yang
// jumlah siswa aktif kelasnya melebihi kapasitas baru ("") bila tidak ada.
// Error query diabaikan pemanggil seperti kode asal (_ = ...).
func (r *Repository) RuanganJadwalPenuh(ctx context.Context, ruanganID string, kapasitasBaru int) (string, error) {
	var jadwalPenuh string
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(MAX(k.nama || ' (' || j.tanggal::text || ')'), '') FROM jadwal_ujian j
		 JOIN kelas k ON k.id = j.kelas_id
		 WHERE j.ruangan_id = $1
		   AND (SELECT COUNT(*) FROM siswa s WHERE s.kelas_id = j.kelas_id AND s.is_active = TRUE) > $2`,
		ruanganID, kapasitasBaru).Scan(&jadwalPenuh)
	return jadwalPenuh, err
}

// --- Tugaskan event + ruangan ke jadwal ---

// GetJadwalKelas mengembalikan (kelasID, kelasNama) untuk satu jadwal.
// Baris tidak ada -> pgx.ErrNoRows.
func (r *Repository) GetJadwalKelas(ctx context.Context, jadwalID string) (kelasID, kelasNama string, err error) {
	err = r.pool.QueryRow(ctx,
		`SELECT j.kelas_id, COALESCE(k.nama, '') FROM jadwal_ujian j
		 LEFT JOIN kelas k ON k.id = j.kelas_id WHERE j.id = $1`, jadwalID).
		Scan(&kelasID, &kelasNama)
	return kelasID, kelasNama, err
}

// CountSiswaAktifKelas menghitung jumlah siswa aktif satu kelas.
func (r *Repository) CountSiswaAktifKelas(ctx context.Context, kelasID string) (int, error) {
	var jumlahSiswa int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM siswa WHERE kelas_id = $1 AND is_active = TRUE`,
		kelasID).Scan(&jumlahSiswa)
	return jumlahSiswa, err
}

// AssignEventRuangan menugaskan/melepas event + ruangan pada jadwal.
func (r *Repository) AssignEventRuangan(ctx context.Context, jadwalID string, eventArg, ruanganArg any) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE jadwal_ujian SET event_id = $2, ruangan_id = $3 WHERE id = $1`,
		jadwalID, eventArg, ruanganArg)
	return err
}
