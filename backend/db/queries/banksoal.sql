-- Kueri sqlc Tim 3: bank soal + soal.
-- Dijalankan via `sqlc generate` (lihat backend/sqlc.yaml).
-- Handler Go memakai SQL yang sama melalui pgxpool; berkas ini sumber
-- kebenaran kueri. Kolom `kunci` (kunci jawaban) TIDAK PERNAH dikirim
-- ke frontend siswa — endpoint siswa (Tim 4) wajib menghilangkannya.

-- name: ListBankSoal :many
SELECT
    b.id, b.kode, b.nama, b.mapel_id, b.kelas_tingkat, b.guru_id,
    b.is_locked, b.created_at, b.updated_at,
    (SELECT COUNT(*) FROM soal s WHERE s.bank_id = b.id) AS jumlah_soal
FROM bank_soal b
WHERE ($1::text IS NULL OR b.kode ILIKE '%' || $1 || '%' OR b.nama ILIKE '%' || $1 || '%')
  AND ($2::uuid IS NULL OR b.mapel_id = $2)
  AND ($3::int IS NULL OR b.kelas_tingkat = $3)
  AND ($4::boolean IS NULL OR b.is_locked = $4)
ORDER BY b.created_at DESC
LIMIT $5 OFFSET $6;

-- name: CountBankSoal :one
SELECT COUNT(*)
FROM bank_soal b
WHERE ($1::text IS NULL OR b.kode ILIKE '%' || $1 || '%' OR b.nama ILIKE '%' || $1 || '%')
  AND ($2::uuid IS NULL OR b.mapel_id = $2)
  AND ($3::int IS NULL OR b.kelas_tingkat = $3)
  AND ($4::boolean IS NULL OR b.is_locked = $4);

-- name: GetBankSoal :one
SELECT
    b.id, b.kode, b.nama, b.mapel_id, b.kelas_tingkat, b.guru_id,
    b.is_locked, b.created_at, b.updated_at,
    (SELECT COUNT(*) FROM soal s WHERE s.bank_id = b.id) AS jumlah_soal
FROM bank_soal b
WHERE b.id = $1;

-- name: CreateBankSoal :one
INSERT INTO bank_soal (kode, nama, mapel_id, kelas_tingkat, guru_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateBankSoal :exec
UPDATE bank_soal
SET kode = $2, nama = $3, mapel_id = $4, kelas_tingkat = $5, guru_id = $6,
    updated_at = now()
WHERE id = $1;

-- name: DeleteBankSoal :exec
DELETE FROM bank_soal WHERE id = $1;

-- name: SetBankSoalLocked :exec
UPDATE bank_soal SET is_locked = $2, updated_at = now() WHERE id = $1;

-- name: ListSoalByBank :many
SELECT
    id, bank_id, tipe, kisi_kisi_id, kompetensi_id, stimulus, pertanyaan,
    media_url, jawaban, kunci, bobot, level, status, created_at, updated_at
FROM soal
WHERE bank_id = $1
  AND ($2::text IS NULL OR tipe = $2)
  AND ($3::text IS NULL OR level = $3)
  AND ($4::uuid IS NULL OR kisi_kisi_id = $4)
  AND ($5::text IS NULL OR pertanyaan ILIKE '%' || $5 || '%')
ORDER BY created_at ASC
LIMIT $6 OFFSET $7;

-- name: CountSoalByBank :one
SELECT COUNT(*)
FROM soal
WHERE bank_id = $1
  AND ($2::text IS NULL OR tipe = $2)
  AND ($3::text IS NULL OR level = $3)
  AND ($4::uuid IS NULL OR kisi_kisi_id = $4)
  AND ($5::text IS NULL OR pertanyaan ILIKE '%' || $5 || '%');

-- name: GetSoal :one
SELECT
    id, bank_id, tipe, kisi_kisi_id, kompetensi_id, stimulus, pertanyaan,
    media_url, jawaban, kunci, bobot, level, status, created_at, updated_at
FROM soal
WHERE id = $1 AND bank_id = $2;

-- name: CreateSoal :one
INSERT INTO soal
    (bank_id, tipe, kisi_kisi_id, kompetensi_id, stimulus, pertanyaan,
     media_url, jawaban, kunci, bobot, level, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: UpdateSoal :exec
UPDATE soal
SET tipe = $3, kisi_kisi_id = $4, kompetensi_id = $5, stimulus = $6,
    pertanyaan = $7, media_url = $8, jawaban = $9, kunci = $10,
    bobot = $11, level = $12, status = $13, updated_at = now()
WHERE id = $1 AND bank_id = $2;

-- name: DeleteSoal :exec
DELETE FROM soal WHERE id = $1 AND bank_id = $2;

-- name: CountPaketByBank :one
-- Jumlah susunan paket yang memakai soal dari bank ini (via paket_soal
-- JOIN soal). Dipakai guard hapus bank soal.
SELECT COUNT(*) FROM paket_soal ps
JOIN soal s ON s.id = ps.soal_id
WHERE s.bank_id = $1;
