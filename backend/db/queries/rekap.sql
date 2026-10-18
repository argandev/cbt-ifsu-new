-- Kueri sqlc Tim 5: rekap nilai, berita acara, remedial.
-- Dijalankan via `sqlc generate` (lihat backend/sqlc.yaml).
-- Handler Go memakai SQL yang sama melalui pgxpool; berkas ini sumber kebenaran kueri.

-- name: ListNilaiByJadwal :many
SELECT
    na.id,
    na.attempt_id,
    na.jadwal_id,
    na.siswa_id,
    na.nilai,
    na.status,
    na.predikat,
    s.nisn,
    s.nama AS nama_siswa,
    k.nama AS nama_kelas,
    (SELECT COUNT(*) FROM jawaban j WHERE j.attempt_id = na.attempt_id) AS jumlah_dijawab
FROM nilai_akhir na
JOIN siswa s ON s.id = na.siswa_id
LEFT JOIN kelas k ON k.id = s.kelas_id
WHERE na.jadwal_id = $1
ORDER BY s.nama ASC;

-- name: ListNilaiByJadwalForSiswa :many
SELECT
    na.id,
    na.attempt_id,
    na.jadwal_id,
    na.siswa_id,
    na.nilai,
    na.status,
    na.predikat,
    s.nisn,
    s.nama AS nama_siswa,
    k.nama AS nama_kelas,
    (SELECT COUNT(*) FROM jawaban j WHERE j.attempt_id = na.attempt_id) AS jumlah_dijawab
FROM nilai_akhir na
JOIN siswa s ON s.id = na.siswa_id
LEFT JOIN kelas k ON k.id = s.kelas_id
WHERE na.jadwal_id = $1
  AND na.siswa_id = $2
  AND na.status = 'final'
ORDER BY s.nama ASC;

-- name: GetNilaiByAttempt :one
SELECT id, attempt_id, jadwal_id, siswa_id, nilai, status, predikat, created_at, updated_at
FROM nilai_akhir
WHERE attempt_id = $1;

-- name: UpsertNilaiAkhir :one
INSERT INTO nilai_akhir (attempt_id, jadwal_id, siswa_id, nilai, status, predikat, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, now())
ON CONFLICT (attempt_id) DO UPDATE SET
    nilai = EXCLUDED.nilai,
    status = EXCLUDED.status,
    predikat = EXCLUDED.predikat,
    updated_at = now()
RETURNING *;

-- name: FinalisasiNilaiByJadwal :many
UPDATE nilai_akhir
SET status = 'final',
    predikat = CASE
        WHEN nilai < 70 THEN 'Belum Tuntas'
        WHEN nilai < 80 THEN 'Cukup'
        WHEN nilai <= 90 THEN 'Baik'
        ELSE 'Sangat Baik'
    END,
    updated_at = now()
WHERE jadwal_id = $1
RETURNING *;

-- name: GetBeritaAcaraByJadwal :one
SELECT id, jadwal_id, isi, dibuat_oleh, created_at
FROM berita_acara
WHERE jadwal_id = $1;

-- name: UpsertBeritaAcara :one
INSERT INTO berita_acara (jadwal_id, isi, dibuat_oleh)
VALUES ($1, $2, $3)
ON CONFLICT (jadwal_id) DO UPDATE SET
    isi = EXCLUDED.isi,
    dibuat_oleh = EXCLUDED.dibuat_oleh
RETURNING *;

-- name: CreateRemedial :one
INSERT INTO remedial (jadwal_id, siswa_id, nilai)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListRemedialByJadwal :many
SELECT id, jadwal_id, siswa_id, nilai, created_at
FROM remedial
WHERE jadwal_id = $1
ORDER BY created_at DESC;
