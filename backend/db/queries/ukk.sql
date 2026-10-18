-- Kueri sqlc Tim Fase 2 — UKK (ukk_ujian, ukk_nilai, ukk_berkas, ukk_sertifikat).
-- Dijalankan via `sqlc generate` (lihat backend/sqlc.yaml).
-- Handler (backend/internal/ukk/ukk.go) memakai SQL yang sama persis
-- seperti di berkas ini lewat pgxpool (pola Tim 5), sampai kode hasil
-- `sqlc generate` dipakai.
--
-- Konvensi filter list: parameter NULL = tanpa filter.
-- Pagination: LIMIT + OFFSET.

-- =====================================================================
-- ukk_ujian
-- =====================================================================

-- name: ListUkk :many
SELECT u.id, u.nama, u.mapel_id, u.kelas_id, u.jurusan_id, u.tanggal,
       u.penguji_internal, u.penguji_eksternal, u.status, u.created_at,
       m.nama AS mapel_nama, k.nama AS kelas_nama, j.kode AS jurusan_kode,
       j.nama AS jurusan_nama,
       (SELECT COUNT(*) FROM ukk_nilai n WHERE n.ukk_id = u.id) AS jumlah_peserta
FROM ukk_ujian u
JOIN mapel m ON m.id = u.mapel_id
JOIN kelas k ON k.id = u.kelas_id
JOIN jurusan j ON j.id = u.jurusan_id
WHERE ($1::uuid IS NULL OR u.jurusan_id = $1)
  AND ($2::uuid IS NULL OR u.kelas_id = $2)
  AND ($3::text IS NULL OR u.status = $3)
  AND ($4::text IS NULL OR u.nama ILIKE '%' || $4 || '%')
ORDER BY u.created_at DESC
LIMIT $5 OFFSET $6;

-- name: CountUkk :one
SELECT COUNT(*) FROM ukk_ujian u
WHERE ($1::uuid IS NULL OR u.jurusan_id = $1)
  AND ($2::uuid IS NULL OR u.kelas_id = $2)
  AND ($3::text IS NULL OR u.status = $3)
  AND ($4::text IS NULL OR u.nama ILIKE '%' || $4 || '%');

-- name: GetUkkByID :one
SELECT u.id, u.nama, u.mapel_id, u.kelas_id, u.jurusan_id, u.tanggal,
       u.penguji_internal, u.penguji_eksternal, u.status, u.created_at,
       m.nama AS mapel_nama, k.nama AS kelas_nama, j.kode AS jurusan_kode,
       j.nama AS jurusan_nama
FROM ukk_ujian u
JOIN mapel m ON m.id = u.mapel_id
JOIN kelas k ON k.id = u.kelas_id
JOIN jurusan j ON j.id = u.jurusan_id
WHERE u.id = $1;

-- name: CreateUkk :one
INSERT INTO ukk_ujian (nama, mapel_id, kelas_id, jurusan_id, tanggal, penguji_internal, penguji_eksternal, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, COALESCE($8, 'draft'))
RETURNING *;

-- name: UpdateUkk :one
UPDATE ukk_ujian
SET nama = $2, mapel_id = $3, kelas_id = $4, jurusan_id = $5, tanggal = $6,
    penguji_internal = $7, penguji_eksternal = $8, status = $9
WHERE id = $1
RETURNING *;

-- name: DeleteUkk :exec
DELETE FROM ukk_ujian WHERE id = $1;

-- name: CountUkkNilaiByUkk :one
SELECT COUNT(*) FROM ukk_nilai WHERE ukk_id = $1;

-- name: CountUkkBerkasByUkk :one
SELECT COUNT(*) FROM ukk_berkas WHERE ukk_id = $1;

-- name: CountUkkSertifikatByUkk :one
SELECT COUNT(*) FROM ukk_sertifikat WHERE ukk_id = $1;

-- =====================================================================
-- ukk_nilai
-- =====================================================================

-- name: ListUkkNilai :many
SELECT n.id, n.ukk_id, n.siswa_id, n.nilai_teori, n.nilai_proses,
       n.nilai_produk, n.nilai_sikap, n.nilai_praktik, n.nilai_akhir,
       n.lulus, n.predikat, n.catatan, n.dinilai_oleh,
       n.created_at, n.updated_at,
       s.nisn AS siswa_nisn, s.nama AS siswa_nama
FROM ukk_nilai n
JOIN siswa s ON s.id = n.siswa_id
WHERE n.ukk_id = $1
ORDER BY s.nama ASC;

-- name: GetUkkNilaiByID :one
SELECT * FROM ukk_nilai WHERE id = $1;

-- name: GetUkkNilaiByUkkSiswa :one
SELECT * FROM ukk_nilai WHERE ukk_id = $1 AND siswa_id = $2;

-- name: UpsertUkkNilai :one
INSERT INTO ukk_nilai (ukk_id, siswa_id, nilai_teori, nilai_proses, nilai_produk,
                       nilai_sikap, nilai_praktik, nilai_akhir, lulus, predikat,
                       catatan, dinilai_oleh)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (ukk_id, siswa_id) DO UPDATE
SET nilai_teori = EXCLUDED.nilai_teori,
    nilai_proses = EXCLUDED.nilai_proses,
    nilai_produk = EXCLUDED.nilai_produk,
    nilai_sikap = EXCLUDED.nilai_sikap,
    nilai_praktik = EXCLUDED.nilai_praktik,
    nilai_akhir = EXCLUDED.nilai_akhir,
    lulus = EXCLUDED.lulus,
    predikat = EXCLUDED.predikat,
    catatan = EXCLUDED.catatan,
    dinilai_oleh = EXCLUDED.dinilai_oleh,
    updated_at = now()
RETURNING *;

-- name: UpdateUkkNilaiByID :one
UPDATE ukk_nilai
SET nilai_teori = $2, nilai_proses = $3, nilai_produk = $4, nilai_sikap = $5,
    nilai_praktik = $6, nilai_akhir = $7, lulus = $8, predikat = $9,
    catatan = $10, dinilai_oleh = $11, updated_at = now()
WHERE id = $1
RETURNING *;

-- =====================================================================
-- ukk_berkas
-- =====================================================================

-- name: ListUkkBerkas :many
SELECT b.id, b.ukk_id, b.siswa_id, b.nama_file, b.path, b.ukuran,
       b.diupload_oleh, b.created_at,
       s.nisn AS siswa_nisn, s.nama AS siswa_nama
FROM ukk_berkas b
JOIN siswa s ON s.id = b.siswa_id
WHERE b.ukk_id = $1
  AND ($2::uuid IS NULL OR b.siswa_id = $2)
ORDER BY b.created_at DESC;

-- name: GetUkkBerkasByID :one
SELECT * FROM ukk_berkas WHERE id = $1;

-- name: CreateUkkBerkas :one
INSERT INTO ukk_berkas (ukk_id, siswa_id, nama_file, path, ukuran, diupload_oleh)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- =====================================================================
-- ukk_sertifikat
-- =====================================================================

-- name: GetUkkSertifikatByUkkSiswa :one
SELECT s.id, s.ukk_id, s.siswa_id, s.nomor_sertifikat, s.diterbitkan_at,
       si.nisn AS siswa_nisn, si.nama AS siswa_nama,
       u.nama AS ukk_nama, u.tanggal AS ukk_tanggal,
       u.penguji_internal, u.penguji_eksternal,
       j.kode AS jurusan_kode, j.nama AS jurusan_nama,
       n.nilai_teori, n.nilai_praktik, n.nilai_akhir, n.predikat
FROM ukk_sertifikat s
JOIN siswa si ON si.id = s.siswa_id
JOIN ukk_ujian u ON u.id = s.ukk_id
JOIN jurusan j ON j.id = u.jurusan_id
LEFT JOIN ukk_nilai n ON n.ukk_id = s.ukk_id AND n.siswa_id = s.siswa_id
WHERE s.ukk_id = $1 AND s.siswa_id = $2;

-- name: CountUkkSertifikatByJurusanTahun :one
SELECT COUNT(*) FROM ukk_sertifikat s
JOIN ukk_ujian u ON u.id = s.ukk_id
WHERE u.jurusan_id = $1
  AND EXTRACT(YEAR FROM s.diterbitkan_at) = $2;

-- name: CreateUkkSertifikat :one
INSERT INTO ukk_sertifikat (ukk_id, siswa_id, nomor_sertifikat)
VALUES ($1, $2, $3)
RETURNING *;
