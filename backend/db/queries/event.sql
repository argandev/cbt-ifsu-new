-- Kueri sqlc Tim Fase 2: Event + Ruangan + penugasan event/ruangan ke
-- jadwal + kartu peserta.
-- Dijalankan via `sqlc generate` (lihat backend/sqlc.yaml).
-- Handler (backend/internal/event/event.go) memakai SQL yang sama
-- persis seperti di berkas ini lewat pgxpool (pola Tim 5), sampai
-- kode hasil sqlc generate menggantikan pemakaian langsung.
--
-- Konvensi filter list: parameter NULL = tanpa filter. Pencarian q
-- mencocokkan nama/kode (ILIKE). Pagination: LIMIT + OFFSET.

-- =====================================================================
-- Event
-- =====================================================================

-- name: ListEvent :many
SELECT id, kode, nama, jenis, tanggal_mulai, tanggal_selesai, deskripsi, created_at FROM event
WHERE ($1::text IS NULL OR jenis = $1)
  AND ($2::text IS NULL OR nama ILIKE '%' || $2 || '%' OR kode ILIKE '%' || $2 || '%')
ORDER BY tanggal_mulai DESC NULLS LAST, nama ASC
LIMIT $3 OFFSET $4;

-- name: CountEvent :one
SELECT COUNT(*) FROM event
WHERE ($1::text IS NULL OR jenis = $1)
  AND ($2::text IS NULL OR nama ILIKE '%' || $2 || '%' OR kode ILIKE '%' || $2 || '%');

-- name: GetEventByID :one
SELECT id, kode, nama, jenis, tanggal_mulai, tanggal_selesai, deskripsi, created_at FROM event WHERE id = $1;

-- name: GetEventByKode :one
SELECT id, kode, nama, jenis, tanggal_mulai, tanggal_selesai, deskripsi, created_at FROM event WHERE kode = $1;

-- name: CreateEvent :one
INSERT INTO event (kode, nama, jenis, tanggal_mulai, tanggal_selesai, deskripsi)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: UpdateEvent :one
UPDATE event SET kode = $2, nama = $3, jenis = $4, tanggal_mulai = $5,
  tanggal_selesai = $6, deskripsi = $7
WHERE id = $1 RETURNING *;

-- name: DeleteEvent :exec
DELETE FROM event WHERE id = $1;

-- name: CountJadwalByEvent :one
SELECT COUNT(*) FROM jadwal_ujian WHERE event_id = $1;

-- =====================================================================
-- Ruangan
-- =====================================================================

-- name: ListRuangan :many
SELECT id, kode, nama, kapasitas, lokasi, keterangan, created_at FROM ruangan
WHERE ($1::text IS NULL OR nama ILIKE '%' || $1 || '%' OR kode ILIKE '%' || $1 || '%' OR COALESCE(lokasi, '') ILIKE '%' || $1 || '%')
ORDER BY nama ASC
LIMIT $2 OFFSET $3;

-- name: CountRuangan :one
SELECT COUNT(*) FROM ruangan
WHERE ($1::text IS NULL OR nama ILIKE '%' || $1 || '%' OR kode ILIKE '%' || $1 || '%' OR COALESCE(lokasi, '') ILIKE '%' || $1 || '%');

-- name: GetRuanganByID :one
SELECT id, kode, nama, kapasitas, lokasi, keterangan, created_at FROM ruangan WHERE id = $1;

-- name: GetRuanganByKode :one
SELECT id, kode, nama, kapasitas, lokasi, keterangan, created_at FROM ruangan WHERE kode = $1;

-- name: CreateRuangan :one
INSERT INTO ruangan (kode, nama, kapasitas, lokasi, keterangan)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: UpdateRuangan :one
UPDATE ruangan SET kode = $2, nama = $3, kapasitas = $4, lokasi = $5, keterangan = $6
WHERE id = $1 RETURNING *;

-- name: DeleteRuangan :exec
DELETE FROM ruangan WHERE id = $1;

-- name: CountJadwalByRuangan :one
SELECT COUNT(*) FROM jadwal_ujian WHERE ruangan_id = $1;

-- =====================================================================
-- Penugasan event + ruangan ke jadwal (migrasi 0009)
-- =====================================================================

-- name: GetJadwalEventRuangan :one
SELECT j.id, j.paket_id, j.kelas_id, j.tanggal, j.jam_mulai, j.jam_selesai,
  j.status, j.event_id, j.ruangan_id,
  e.kode AS event_kode, e.nama AS event_nama,
  r.kode AS ruangan_kode, r.nama AS ruangan_nama, r.kapasitas AS ruangan_kapasitas
FROM jadwal_ujian j
LEFT JOIN event e ON e.id = j.event_id
LEFT JOIN ruangan r ON r.id = j.ruangan_id
WHERE j.id = $1;

-- name: UpdateJadwalEventRuangan :exec
UPDATE jadwal_ujian SET event_id = $2, ruangan_id = $3 WHERE id = $1;

-- name: CountSiswaAktifByKelas :one
SELECT COUNT(*) FROM siswa WHERE kelas_id = $1 AND is_active = TRUE;

-- =====================================================================
-- Kartu peserta: satu baris per siswa aktif di kelas jadwal.
-- Nomor peserta dihitung di handler (per jadwal, urut nama ASC):
--   <kode-event atau US>/<tahun tanggal jadwal>/<nomor urut 3 digit>
-- Kartu TIDAK berisi token ujian (token hanya dari proktor saat ujian).
-- =====================================================================

-- name: GetJadwalUntukKartu :one
SELECT j.id, j.kelas_id, j.tanggal, j.jam_mulai, j.jam_selesai, j.status,
  p.nama AS paket_nama, p.jenis AS paket_jenis,
  m.nama AS mapel_nama,
  k.nama AS kelas_nama, ju.nama AS jurusan_nama,
  j.event_id, e.kode AS event_kode, e.nama AS event_nama,
  j.ruangan_id, r.nama AS ruangan_nama, r.lokasi AS ruangan_lokasi
FROM jadwal_ujian j
JOIN paket_ujian p ON p.id = j.paket_id
JOIN mapel m ON m.id = p.mapel_id
JOIN kelas k ON k.id = j.kelas_id
JOIN jurusan ju ON ju.id = k.jurusan_id
LEFT JOIN event e ON e.id = j.event_id
LEFT JOIN ruangan r ON r.id = j.ruangan_id
WHERE j.id = $1;

-- name: ListSiswaUntukKartu :many
SELECT s.id, s.nisn, s.nama, s.kelas_id, s.user_id
FROM siswa s
WHERE s.kelas_id = $1 AND s.is_active = TRUE
ORDER BY s.nama ASC;

-- name: GetSiswaUntukKartu :one
SELECT s.id, s.nisn, s.nama, s.kelas_id, s.user_id
FROM siswa s
WHERE s.id = $1;

-- name: GetNamaSekolah :one
SELECT value FROM settings WHERE key = 'nama_sekolah';
