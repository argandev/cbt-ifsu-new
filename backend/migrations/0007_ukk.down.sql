-- Rollback migrasi 0007 (Tim Fase 2 — UKK).
-- Urutan: anak dulu (sertifikat, berkas, nilai), lalu ukk_ujian.
-- Berkas fisik di ./storage/ukk/ TIDAK dihapus otomatis oleh rollback;
-- hapus manual bila memang ingin membersihkan arsip.
DROP TABLE IF EXISTS ukk_sertifikat;
DROP TABLE IF EXISTS ukk_berkas;
DROP TABLE IF EXISTS ukk_nilai;
DROP TABLE IF EXISTS ukk_ujian;
