-- Rollback migrasi 0004 (Tim 4: Mesin Ujian + Monitor).
-- Urutan: anak -> induk. Jalankan HANYA di luar jam ujian.
DROP TABLE IF EXISTS usulan_unblock;
DROP TABLE IF EXISTS pelanggaran;
DROP TABLE IF EXISTS jawaban;
DROP TABLE IF EXISTS attempt;
DROP TABLE IF EXISTS jadwal_ujian;
DROP TABLE IF EXISTS paket_soal;
DROP TABLE IF EXISTS paket_ujian;
