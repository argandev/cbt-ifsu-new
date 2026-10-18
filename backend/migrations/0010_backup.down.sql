-- Rollback migrasi 0010 (Tim Fase 2 — Backup / Restore).
-- Hanya menghapus tabel metadata. File dump di ./storage/backup/
-- TIDAK dihapus oleh rollback ini (hapus manual bila perlu).
DROP TABLE IF EXISTS backup_meta;
