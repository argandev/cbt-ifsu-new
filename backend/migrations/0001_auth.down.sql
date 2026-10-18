-- Migrasi 0001 down (Tim 1) — rollback fondasi auth.
-- PERINGATAN: menghapus tabel auth + seluruh data user/session/audit.
-- Hanya dipakai di lingkungan dev. Jangan jalankan di produksi berisi data.
DROP TABLE IF EXISTS settings CASCADE;
DROP TABLE IF EXISTS audit_log CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
