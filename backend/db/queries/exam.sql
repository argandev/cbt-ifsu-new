-- Kueri sqlc Tim 4: paket, jadwal, attempt, jawaban, pelanggaran, unblock, koreksi.
-- Dijalankan via `sqlc generate` (lihat backend/sqlc.yaml).
-- Handler Go memakai SQL yang sama melalui pgxpool (pola Tim 5);
-- berkas ini sumber kebenaran kueri.
-- Nama tabel mengikuti spesifikasi Tim 4: paket_ujian, jadwal_ujian,
-- attempt (singular). Lihat catatan konflik di 0004_exam.up.sql.

-- ============================== PAKET ==============================

-- name: ListPaket :many
SELECT id, kode, nama, mapel_id, jenis, durasi_menit, jumlah_soal,
       acak_soal, acak_jawaban, tampil_hasil, created_at
FROM paket_ujian
ORDER BY created_at DESC;

-- name: GetPaket :one
SELECT id, kode, nama, mapel_id, jenis, durasi_menit, jumlah_soal,
       acak_soal, acak_jawaban, tampil_hasil, created_at
FROM paket_ujian
WHERE id = $1;

-- name: CreatePaket :one
INSERT INTO paket_ujian (kode, nama, mapel_id, jenis, durasi_menit, jumlah_soal,
                         acak_soal, acak_jawaban, tampil_hasil)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdatePaket :exec
UPDATE paket_ujian
SET kode = $2, nama = $3, mapel_id = $4, jenis = $5, durasi_menit = $6,
    jumlah_soal = $7, acak_soal = $8, acak_jawaban = $9, tampil_hasil = $10
WHERE id = $1;

-- name: DeletePaket :exec
DELETE FROM paket_ujian WHERE id = $1;

-- name: ListPaketSoal :many
SELECT ps.paket_id, ps.soal_id, ps.urutan,
       s.tipe, s.pertanyaan, s.bobot
FROM paket_soal ps
JOIN soal s ON s.id = ps.soal_id
WHERE ps.paket_id = $1
ORDER BY ps.urutan ASC;

-- name: AddPaketSoal :exec
INSERT INTO paket_soal (paket_id, soal_id, urutan)
VALUES ($1, $2, $3)
ON CONFLICT (paket_id, soal_id) DO NOTHING;

-- name: DeletePaketSoal :exec
DELETE FROM paket_soal WHERE paket_id = $1 AND soal_id = $2;

-- name: MaxUrutanPaketSoal :one
SELECT COALESCE(MAX(urutan), 0) FROM paket_soal WHERE paket_id = $1;

-- name: ListSoalByMapelAcak :many
-- Soal Tim 3 tidak punya mapel_id langsung: mapel via bank_soal.bank_id.
-- Hanya soal aktif yang boleh masuk paket.
SELECT s.id FROM soal s
JOIN bank_soal bs ON bs.id = s.bank_id
WHERE bs.mapel_id = $1 AND s.status = 'aktif'
  AND s.id NOT IN (SELECT soal_id FROM paket_soal WHERE paket_id = $2)
ORDER BY random()
LIMIT $3;

-- name: GetSoalMapel :one
SELECT bs.mapel_id FROM soal s
JOIN bank_soal bs ON bs.id = s.bank_id
WHERE s.id = $1;

-- ============================== JADWAL ==============================

-- name: ListJadwal :many
SELECT j.id, j.paket_id, j.kelas_id, j.tanggal, j.jam_mulai, j.jam_selesai,
       j.status, j.token_expires_at, j.created_at,
       p.kode AS paket_kode, p.nama AS paket_nama, p.jenis AS paket_jenis,
       COALESCE(k.nama, '') AS kelas_nama
FROM jadwal_ujian j
JOIN paket_ujian p ON p.id = j.paket_id
LEFT JOIN kelas k ON k.id = j.kelas_id
ORDER BY j.tanggal DESC, j.jam_mulai DESC;

-- name: GetJadwal :one
SELECT j.id, j.paket_id, j.kelas_id, j.tanggal, j.jam_mulai, j.jam_selesai,
       j.status, j.token_hash, j.token_expires_at, j.created_at,
       p.kode AS paket_kode, p.nama AS paket_nama, p.jenis AS paket_jenis,
       p.durasi_menit, p.tampil_hasil
FROM jadwal_ujian j
JOIN paket_ujian p ON p.id = j.paket_id
WHERE j.id = $1;

-- name: CreateJadwal :one
INSERT INTO jadwal_ujian (paket_id, kelas_id, tanggal, jam_mulai, jam_selesai, status)
VALUES ($1, $2, $3, $4, $5, COALESCE($6, 'draft'))
RETURNING *;

-- name: UpdateJadwal :exec
UPDATE jadwal_ujian
SET paket_id = $2, kelas_id = $3, tanggal = $4, jam_mulai = $5, jam_selesai = $6,
    status = $7
WHERE id = $1;

-- name: DeleteJadwal :exec
DELETE FROM jadwal_ujian WHERE id = $1;

-- name: CountAttemptByJadwal :one
SELECT COUNT(*) FROM attempt WHERE jadwal_id = $1;

-- name: UpdateJadwalStatus :exec
UPDATE jadwal_ujian SET status = $2 WHERE id = $1;

-- name: UpdateJadwalToken :exec
UPDATE jadwal_ujian SET token_hash = $2, token_expires_at = $3 WHERE id = $1;

-- name: FreezeAttemptsOnPause :exec
UPDATE attempt
SET sisa_waktu_beku = GREATEST(0, EXTRACT(EPOCH FROM (deadline - now()))::INT)
WHERE jadwal_id = $1 AND status = 'berjalan';

-- name: ResumeAttemptsOnResume :exec
UPDATE attempt
SET deadline = now() + make_interval(secs => COALESCE(sisa_waktu_beku, 0)),
    sisa_waktu_beku = NULL
WHERE jadwal_id = $1 AND status = 'berjalan' AND sisa_waktu_beku IS NOT NULL;

-- ============================== ATTEMPT / UJIAN ==============================

-- name: GetAttempt :one
SELECT id, jadwal_id, siswa_id, status, mulai_at, deadline,
       sisa_waktu_beku, urutan_soal, urutan_opsi, submit_at, nilai_akhir, created_at
FROM attempt
WHERE id = $1;

-- name: GetAttemptBerjalan :one
SELECT id, jadwal_id, siswa_id, status, mulai_at, deadline,
       sisa_waktu_beku, urutan_soal, urutan_opsi, submit_at, nilai_akhir, created_at
FROM attempt
WHERE jadwal_id = $1 AND siswa_id = $2 AND status = 'berjalan'
ORDER BY created_at DESC
LIMIT 1;

-- name: GetAttemptByJadwalSiswa :many
SELECT id, jadwal_id, siswa_id, status, mulai_at, deadline,
       sisa_waktu_beku, urutan_soal, urutan_opsi, submit_at, nilai_akhir, created_at
FROM attempt
WHERE jadwal_id = $1 AND siswa_id = $2
ORDER BY created_at DESC;

-- name: CreateAttempt :one
INSERT INTO attempt (jadwal_id, siswa_id, status, mulai_at, deadline,
                     urutan_soal, urutan_opsi)
VALUES ($1, $2, 'berjalan', now(), $3, $4, $5)
RETURNING *;

-- name: UpdateAttemptStatus :exec
UPDATE attempt SET status = $2 WHERE id = $1;

-- name: SubmitAttempt :exec
UPDATE attempt
SET status = 'selesai', submit_at = now(), nilai_akhir = $2
WHERE id = $1;

-- name: UnblockAttempt :exec
UPDATE attempt
SET status = 'berjalan',
    deadline = now() + make_interval(secs => $2),
    sisa_waktu_beku = NULL
WHERE id = $1;

-- name: AutoSubmitExpiredAttempt :exec
UPDATE attempt
SET status = 'selesai', submit_at = now()
WHERE id = $1 AND status = 'berjalan' AND deadline <= now();

-- name: ListAttemptsByJadwal :many
SELECT a.id, a.jadwal_id, a.siswa_id, a.status, a.mulai_at, a.deadline,
       a.sisa_waktu_beku, a.submit_at, a.nilai_akhir, a.created_at,
       COALESCE(s.nama, '') AS nama_siswa, COALESCE(s.nisn, '') AS nisn
FROM attempt a
LEFT JOIN siswa s ON s.id = a.siswa_id
WHERE a.jadwal_id = $1
ORDER BY a.created_at ASC;

-- ============================== SOAL UJIAN ==============================
-- KUNCI JAWABAN TIDAK PERNAH dikirim ke browser: kueri untuk endpoint
-- ujian siswa (GetSoalUjian) TIDAK memilih kolom kunci/pembahasan.

-- name: ListSoalPaketForUjian :many
-- KUNCI JAWABAN TIDAK PERNAH dipilih di sini. s.jawaban = data tampil
-- Tim 3 (opsi/pernyataan/pasangan/item), BUKAN jawaban siswa.
SELECT s.id, s.tipe, s.pertanyaan, s.jawaban, s.bobot, ps.urutan,
       s.stimulus, s.media_url
FROM paket_soal ps
JOIN soal s ON s.id = ps.soal_id
WHERE ps.paket_id = $1
ORDER BY ps.urutan ASC;

-- name: GetSoalForGrading :one
SELECT id, tipe, kunci, bobot FROM soal WHERE id = $1;

-- ============================== JAWABAN ==============================

-- name: ListJawabanByAttempt :many
SELECT id, attempt_id, soal_id, jawaban, ragu, poin, dikoreksi, updated_at
FROM jawaban
WHERE attempt_id = $1;

-- name: UpsertJawaban :one
INSERT INTO jawaban (attempt_id, soal_id, jawaban, ragu, poin, dikoreksi, updated_at)
VALUES ($1, $2, $3, COALESCE($4, FALSE), $5, COALESCE($6, FALSE), now())
ON CONFLICT (attempt_id, soal_id) DO UPDATE SET
    jawaban = COALESCE(EXCLUDED.jawaban, jawaban.jawaban),
    ragu = EXCLUDED.ragu,
    poin = EXCLUDED.poin,
    dikoreksi = EXCLUDED.dikoreksi,
    updated_at = now()
RETURNING *;

-- name: UpdateJawabanRagu :exec
UPDATE jawaban SET ragu = $3, updated_at = now()
WHERE attempt_id = $1 AND soal_id = $2;

-- name: EnsureJawabanRow :exec
INSERT INTO jawaban (attempt_id, soal_id, jawaban, ragu)
VALUES ($1, $2, NULL, $3)
ON CONFLICT (attempt_id, soal_id) DO UPDATE SET ragu = EXCLUDED.ragu, updated_at = now();

-- name: GetJawaban :one
SELECT id, attempt_id, soal_id, jawaban, ragu, poin, dikoreksi, updated_at
FROM jawaban
WHERE id = $1;

-- ============================== PELANGGARAN ==============================

-- name: CreatePelanggaran :one
INSERT INTO pelanggaran (attempt_id, jenis, detail) VALUES ($1, $2, $3) RETURNING *;

-- name: CountPelanggaran :one
SELECT COUNT(*) FROM pelanggaran WHERE attempt_id = $1;

-- name: LastPelanggaran :one
SELECT id, attempt_id, jenis, detail, created_at
FROM pelanggaran WHERE attempt_id = $1
ORDER BY created_at DESC LIMIT 1;

-- name: ListPelanggaranByAttempt :many
SELECT id, attempt_id, jenis, detail, created_at
FROM pelanggaran WHERE attempt_id = $1 ORDER BY created_at ASC;

-- ============================== USULAN UNBLOCK ==============================

-- name: CreateUsulanUnblock :one
INSERT INTO usulan_unblock (attempt_id, pengawas_id, alasan, status)
VALUES ($1, $2, $3, 'menunggu') RETURNING *;

-- name: GetPendingUsulan :one
SELECT id, attempt_id, pengawas_id, alasan, status, diproses_oleh, diproses_at, created_at
FROM usulan_unblock
WHERE attempt_id = $1 AND status = 'menunggu'
ORDER BY created_at DESC LIMIT 1;

-- name: ListUsulanByJadwal :many
SELECT u.id, u.attempt_id, u.pengawas_id, u.alasan, u.status,
       u.diproses_oleh, u.diproses_at, u.created_at,
       COALESCE(s.nama, '') AS nama_siswa
FROM usulan_unblock u
JOIN attempt a ON a.id = u.attempt_id
LEFT JOIN siswa s ON s.id = a.siswa_id
WHERE a.jadwal_id = $1
ORDER BY u.created_at DESC;

-- name: UpdateUsulanStatus :exec
UPDATE usulan_unblock
SET status = $2, diproses_oleh = $3, diproses_at = now()
WHERE id = $1;

-- name: ListHasilByAttempt :many
SELECT jw.soal_id, jw.jawaban, jw.poin, jw.dikoreksi, jw.ragu,
       so.tipe, so.pertanyaan, so.jawaban, so.kunci, so.bobot
FROM jawaban jw JOIN soal so ON so.id = jw.soal_id
WHERE jw.attempt_id = $1;

-- ============================== KOREKSI ==============================

-- name: ListKoreksiByJadwal :many
SELECT jw.id AS jawaban_id, jw.attempt_id, jw.soal_id, jw.jawaban, jw.ragu,
       jw.poin, jw.dikoreksi, jw.updated_at,
       a.siswa_id, a.status AS attempt_status,
       COALESCE(s.nama, '') AS nama_siswa,
       so.tipe, so.pertanyaan, so.bobot
FROM jawaban jw
JOIN attempt a ON a.id = jw.attempt_id
JOIN soal so ON so.id = jw.soal_id
LEFT JOIN siswa s ON s.id = a.siswa_id
WHERE a.jadwal_id = $1
  AND jw.dikoreksi = FALSE
  AND jw.jawaban IS NOT NULL
  AND so.tipe IN ('uraian', 'upload_file', 'praktik_singkat')
ORDER BY s.nama ASC, so.pertanyaan ASC;

-- name: UpdateJawabanKoreksi :exec
UPDATE jawaban SET poin = $2, dikoreksi = TRUE, updated_at = now() WHERE id = $1;

-- name: CountBelumKoreksi :one
SELECT COUNT(*)
FROM jawaban jw
JOIN attempt a ON a.id = jw.attempt_id
JOIN soal so ON so.id = jw.soal_id
WHERE a.id = $1 AND jw.dikoreksi = FALSE
  AND so.tipe IN ('uraian', 'upload_file', 'praktik_singkat');
