-- Rollback migrasi 0011: hapus kolom pembahasan dari tabel soal.
ALTER TABLE soal
    DROP COLUMN IF EXISTS pembahasan;
