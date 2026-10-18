-- Kueri sqlc Tim 2: Master Data (jurusan, kelas, mapel, guru, siswa,
-- kompetensi, kisi_kisi) + pembuatan akun siswa untuk import CSV.
-- Dijalankan via `sqlc generate` (lihat backend/sqlc.yaml).
-- Handler (backend/internal/master/master.go) memakai SQL yang sama
-- persis seperti di berkas ini lewat pgxpool (pola Tim 5), sampai
-- kode hasil sqlc generate menggantikan pemakaian langsung.
--
-- Konvensi filter list: parameter NULL = tanpa filter. Pencarian q
-- mencocokkan nama/kode (ILIKE). Pagination: LIMIT + OFFSET.
-- Tabel users dibuat di migrasi 0001 (Tim 1) dengan kolom role tunggal.

-- =====================================================================
-- Jurusan
-- =====================================================================

-- name: ListJurusan :many
SELECT id, kode, nama, created_at, updated_at FROM jurusan
WHERE ($1::text IS NULL OR nama ILIKE '%' || $1 || '%' OR kode ILIKE '%' || $1 || '%')
ORDER BY nama ASC
LIMIT $2 OFFSET $3;

-- name: CountJurusan :one
SELECT COUNT(*) FROM jurusan
WHERE ($1::text IS NULL OR nama ILIKE '%' || $1 || '%' OR kode ILIKE '%' || $1 || '%');

-- name: GetJurusanByID :one
SELECT id, kode, nama, created_at, updated_at FROM jurusan WHERE id = $1;

-- name: GetJurusanByKode :one
SELECT id, kode, nama, created_at, updated_at FROM jurusan WHERE kode = $1;

-- name: CreateJurusan :one
INSERT INTO jurusan (kode, nama) VALUES ($1, $2) RETURNING *;

-- name: UpdateJurusan :one
UPDATE jurusan SET kode = $2, nama = $3, updated_at = now()
WHERE id = $1 RETURNING *;

-- name: DeleteJurusan :exec
DELETE FROM jurusan WHERE id = $1;

-- name: CountKelasByJurusan :one
SELECT COUNT(*) FROM kelas WHERE jurusan_id = $1;

-- name: CountMapelByJurusan :one
SELECT COUNT(*) FROM mapel WHERE jurusan_id = $1;

-- name: CountSiswaByJurusan :one
SELECT COUNT(*) FROM siswa WHERE jurusan_id = $1;

-- =====================================================================
-- Kelas / Rombel
-- =====================================================================

-- name: ListKelas :many
SELECT id, jurusan_id, tingkat, nama, tahun_ajaran, created_at, updated_at FROM kelas
WHERE ($1::uuid IS NULL OR jurusan_id = $1)
  AND ($2::text IS NULL OR nama ILIKE '%' || $2 || '%' OR tahun_ajaran ILIKE '%' || $2 || '%')
ORDER BY tahun_ajaran DESC, tingkat ASC, nama ASC
LIMIT $3 OFFSET $4;

-- name: CountKelas :one
SELECT COUNT(*) FROM kelas
WHERE ($1::uuid IS NULL OR jurusan_id = $1)
  AND ($2::text IS NULL OR nama ILIKE '%' || $2 || '%' OR tahun_ajaran ILIKE '%' || $2 || '%');

-- name: GetKelasByID :one
SELECT id, jurusan_id, tingkat, nama, tahun_ajaran, created_at, updated_at FROM kelas WHERE id = $1;

-- name: GetKelasByNama :one
SELECT id, jurusan_id, tingkat, nama, tahun_ajaran, created_at, updated_at FROM kelas
WHERE LOWER(nama) = LOWER($1)
ORDER BY created_at DESC LIMIT 1;

-- name: CreateKelas :one
INSERT INTO kelas (jurusan_id, tingkat, nama, tahun_ajaran)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpdateKelas :one
UPDATE kelas SET jurusan_id = $2, tingkat = $3, nama = $4, tahun_ajaran = $5, updated_at = now()
WHERE id = $1 RETURNING *;

-- name: DeleteKelas :exec
DELETE FROM kelas WHERE id = $1;

-- name: CountSiswaByKelas :one
SELECT COUNT(*) FROM siswa WHERE kelas_id = $1;

-- =====================================================================
-- Mapel
-- =====================================================================

-- name: ListMapel :many
SELECT id, kode, nama, jenis, jurusan_id, kelas_tingkat, created_at, updated_at FROM mapel
WHERE ($1::text IS NULL OR jenis = $1)
  AND ($2::uuid IS NULL OR jurusan_id = $2)
  AND ($3::text IS NULL OR nama ILIKE '%' || $3 || '%' OR kode ILIKE '%' || $3 || '%')
ORDER BY nama ASC
LIMIT $4 OFFSET $5;

-- name: CountMapel :one
SELECT COUNT(*) FROM mapel
WHERE ($1::text IS NULL OR jenis = $1)
  AND ($2::uuid IS NULL OR jurusan_id = $2)
  AND ($3::text IS NULL OR nama ILIKE '%' || $3 || '%' OR kode ILIKE '%' || $3 || '%');

-- name: GetMapelByID :one
SELECT id, kode, nama, jenis, jurusan_id, kelas_tingkat, created_at, updated_at FROM mapel WHERE id = $1;

-- name: GetMapelByKode :one
SELECT id, kode, nama, jenis, jurusan_id, kelas_tingkat, created_at, updated_at FROM mapel WHERE kode = $1;

-- name: CreateMapel :one
INSERT INTO mapel (kode, nama, jenis, jurusan_id, kelas_tingkat)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: UpdateMapel :one
UPDATE mapel SET kode = $2, nama = $3, jenis = $4, jurusan_id = $5, kelas_tingkat = $6, updated_at = now()
WHERE id = $1 RETURNING *;

-- name: DeleteMapel :exec
DELETE FROM mapel WHERE id = $1;

-- name: CountKompetensiByMapel :one
SELECT COUNT(*) FROM kompetensi WHERE mapel_id = $1;

-- name: CountKisiKisiByMapel :one
SELECT COUNT(*) FROM kisi_kisi WHERE mapel_id = $1;

-- name: CountBankSoalByMapel :one
SELECT COUNT(*) FROM bank_soal WHERE mapel_id = $1;

-- =====================================================================
-- Guru
-- =====================================================================

-- name: ListGuru :many
SELECT id, user_id, nip, nama, mapel_ids, created_at, updated_at FROM guru
WHERE ($1::text IS NULL OR nama ILIKE '%' || $1 || '%' OR COALESCE(nip, '') ILIKE '%' || $1 || '%')
ORDER BY nama ASC
LIMIT $2 OFFSET $3;

-- name: CountGuru :one
SELECT COUNT(*) FROM guru
WHERE ($1::text IS NULL OR nama ILIKE '%' || $1 || '%' OR COALESCE(nip, '') ILIKE '%' || $1 || '%');

-- name: GetGuruByID :one
SELECT id, user_id, nip, nama, mapel_ids, created_at, updated_at FROM guru WHERE id = $1;

-- name: CreateGuru :one
INSERT INTO guru (user_id, nip, nama, mapel_ids) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpdateGuru :one
UPDATE guru SET user_id = $2, nip = $3, nama = $4, mapel_ids = $5, updated_at = now()
WHERE id = $1 RETURNING *;

-- name: DeleteGuru :exec
DELETE FROM guru WHERE id = $1;

-- name: CountBankSoalByGuru :one
SELECT COUNT(*) FROM bank_soal WHERE guru_id = $1;

-- =====================================================================
-- Siswa
-- =====================================================================

-- name: ListSiswa :many
SELECT id, user_id, nisn, nis, nama, kelas_id, jurusan_id, tahun_masuk, is_active, created_at, updated_at FROM siswa
WHERE ($1::uuid IS NULL OR kelas_id = $1)
  AND ($2::uuid IS NULL OR jurusan_id = $2)
  AND ($3::boolean IS NULL OR is_active = $3)
  AND ($4::text IS NULL OR nama ILIKE '%' || $4 || '%' OR nisn ILIKE '%' || $4 || '%' OR COALESCE(nis, '') ILIKE '%' || $4 || '%')
ORDER BY nama ASC
LIMIT $5 OFFSET $6;

-- name: CountSiswa :one
SELECT COUNT(*) FROM siswa
WHERE ($1::uuid IS NULL OR kelas_id = $1)
  AND ($2::uuid IS NULL OR jurusan_id = $2)
  AND ($3::boolean IS NULL OR is_active = $3)
  AND ($4::text IS NULL OR nama ILIKE '%' || $4 || '%' OR nisn ILIKE '%' || $4 || '%' OR COALESCE(nis, '') ILIKE '%' || $4 || '%');

-- name: GetSiswaByID :one
SELECT id, user_id, nisn, nis, nama, kelas_id, jurusan_id, tahun_masuk, is_active, created_at, updated_at FROM siswa WHERE id = $1;

-- name: GetSiswaByNISN :one
SELECT id, user_id, nisn, nis, nama, kelas_id, jurusan_id, tahun_masuk, is_active, created_at, updated_at FROM siswa WHERE nisn = $1;

-- name: CreateSiswa :one
INSERT INTO siswa (user_id, nisn, nis, nama, kelas_id, jurusan_id, tahun_masuk, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *;

-- name: UpdateSiswa :one
UPDATE siswa SET user_id = $2, nisn = $3, nis = $4, nama = $5, kelas_id = $6, jurusan_id = $7,
  tahun_masuk = $8, is_active = $9, updated_at = now()
WHERE id = $1 RETURNING *;

-- name: DeleteSiswa :exec
DELETE FROM siswa WHERE id = $1;

-- =====================================================================
-- Kompetensi
-- =====================================================================

-- name: ListKompetensi :many
SELECT id, mapel_id, kode, deskripsi, created_at, updated_at FROM kompetensi
WHERE ($1::uuid IS NULL OR mapel_id = $1)
  AND ($2::text IS NULL OR kode ILIKE '%' || $2 || '%' OR deskripsi ILIKE '%' || $2 || '%')
ORDER BY kode ASC
LIMIT $3 OFFSET $4;

-- name: CountKompetensi :one
SELECT COUNT(*) FROM kompetensi
WHERE ($1::uuid IS NULL OR mapel_id = $1)
  AND ($2::text IS NULL OR kode ILIKE '%' || $2 || '%' OR deskripsi ILIKE '%' || $2 || '%');

-- name: GetKompetensiByID :one
SELECT id, mapel_id, kode, deskripsi, created_at, updated_at FROM kompetensi WHERE id = $1;

-- name: CreateKompetensi :one
INSERT INTO kompetensi (mapel_id, kode, deskripsi) VALUES ($1, $2, $3) RETURNING *;

-- name: UpdateKompetensi :one
UPDATE kompetensi SET mapel_id = $2, kode = $3, deskripsi = $4, updated_at = now()
WHERE id = $1 RETURNING *;

-- name: DeleteKompetensi :exec
DELETE FROM kompetensi WHERE id = $1;

-- name: CountKisiKisiByKompetensi :one
SELECT COUNT(*) FROM kisi_kisi WHERE kompetensi_id = $1;

-- name: CountSoalByKompetensi :one
SELECT COUNT(*) FROM soal WHERE kompetensi_id = $1;

-- =====================================================================
-- Kisi-kisi
-- =====================================================================

-- name: ListKisiKisi :many
SELECT id, mapel_id, kelas_tingkat, kelompok, kompetensi_id, jumlah_soal, created_at, updated_at FROM kisi_kisi
WHERE ($1::uuid IS NULL OR mapel_id = $1)
  AND ($2::int IS NULL OR kelas_tingkat = $2)
  AND ($3::text IS NULL OR kelompok ILIKE '%' || $3 || '%')
ORDER BY kelas_tingkat ASC, kelompok ASC
LIMIT $4 OFFSET $5;

-- name: CountKisiKisi :one
SELECT COUNT(*) FROM kisi_kisi
WHERE ($1::uuid IS NULL OR mapel_id = $1)
  AND ($2::int IS NULL OR kelas_tingkat = $2)
  AND ($3::text IS NULL OR kelompok ILIKE '%' || $3 || '%');

-- name: GetKisiKisiByID :one
SELECT id, mapel_id, kelas_tingkat, kelompok, kompetensi_id, jumlah_soal, created_at, updated_at FROM kisi_kisi WHERE id = $1;

-- name: CreateKisiKisi :one
INSERT INTO kisi_kisi (mapel_id, kelas_tingkat, kelompok, kompetensi_id, jumlah_soal)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: UpdateKisiKisi :one
UPDATE kisi_kisi SET mapel_id = $2, kelas_tingkat = $3, kelompok = $4, kompetensi_id = $5,
  jumlah_soal = $6, updated_at = now()
WHERE id = $1 RETURNING *;

-- name: DeleteKisiKisi :exec
DELETE FROM kisi_kisi WHERE id = $1;

-- name: CountSoalByKisiKisi :one
SELECT COUNT(*) FROM soal WHERE kisi_kisi_id = $1;

-- =====================================================================
-- Akun siswa (import CSV). Role siswa disimpan di kolom users.role
-- (satu role per user, migrasi 0001 Tim 1 — tidak ada user_roles).
-- Kontrak kolom users: nama, username UNIQUE, password_hash, role,
-- must_change_password, is_active. Password awal = NISN, di-hash
-- Argon2id (format PHC) di handler; siswa WAJIB ganti saat login pertama.
-- =====================================================================

-- name: CreateUserForSiswa :one
INSERT INTO users (nama, username, password_hash, role, must_change_password, is_active)
VALUES ($1, $2, $3, 'siswa', TRUE, TRUE)
ON CONFLICT (username) DO NOTHING
RETURNING id;

-- name: GetUserIDByUsername :one
SELECT id FROM users WHERE username = $1;
