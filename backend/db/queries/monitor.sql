-- Kueri sqlc Tim 4: monitor real-time (snapshot peserta per jadwal).
-- Dijalankan via `sqlc generate` (lihat backend/sqlc.yaml).
-- Handler Go memakai SQL yang sama melalui pgxpool (pola Tim 5);
-- berkas ini sumber kebenaran kueri.

-- name: MonitorSnapshot :many
SELECT a.id AS attempt_id, a.siswa_id, a.status, a.mulai_at, a.deadline,
       a.sisa_waktu_beku, a.nilai_akhir,
       COALESCE(s.nama, '') AS nama_siswa, COALESCE(s.nisn, '') AS nisn,
       COALESCE(k.nama, '') AS kelas_nama,
       (SELECT COUNT(*) FROM jawaban j
         WHERE j.attempt_id = a.id AND j.jawaban IS NOT NULL) AS dijawab,
       (SELECT COUNT(*) FROM jawaban j
         WHERE j.attempt_id = a.id AND j.ragu = TRUE) AS ragu,
       (SELECT COUNT(*) FROM pelanggaran p WHERE p.attempt_id = a.id) AS pelanggaran,
       (SELECT COUNT(*) FROM paket_soal ps
         JOIN jadwal_ujian ju ON ju.paket_id = ps.paket_id
         WHERE ju.id = a.jadwal_id) AS total_soal
FROM attempt a
LEFT JOIN siswa s ON s.id = a.siswa_id
LEFT JOIN kelas k ON k.id = s.kelas_id
WHERE a.jadwal_id = $1
ORDER BY s.nama ASC;
