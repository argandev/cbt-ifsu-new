-- Rollback migrasi 0003 (Tim 3: Bank Soal). Urutan: soal -> bank_soal.
DROP TABLE IF EXISTS soal;
DROP TABLE IF EXISTS bank_soal;
