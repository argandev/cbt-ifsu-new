-- Migrasi 0013 (Tim Backend — Manajemen Pengguna staf, UC20).
--
-- 1. Permission baru `pengguna.*` (5): read, create, update,
--    reset-password, toggle-aktif. Grant ke admin di bawah ini adalah
--    jaring pengaman bila roles sudah ada (DB yang sudah di-seed);
--    grant yang berlaku untuk fresh install berasal dari seed
--    (backend/db/seed/seed.sql, admin CROSS JOIN semua permission).
--    Pelajaran BUG-01 (migrasi 0008/0012): grant migrasi diam-diam
--    no-op bila tabel `roles` masih KOSONG saat migrasi berjalan.
-- 2. Kolom `users.terakhir_masuk` (TIMESTAMPTZ NULL): diisi handler
--    login saat login sukses; NULL = belum pernah login.
-- Semua pernyataan idempotent (IF NOT EXISTS / ON CONFLICT DO NOTHING).

INSERT INTO permissions (code, description, group_name) VALUES
  ('pengguna.read',           'Lihat daftar & detail akun staf',        'pengguna'),
  ('pengguna.create',         'Buat akun staf',                         'pengguna'),
  ('pengguna.update',         'Ubah nama/email/role akun staf',         'pengguna'),
  ('pengguna.reset-password', 'Reset password akun staf',               'pengguna'),
  ('pengguna.toggle-aktif',  'Aktif/nonaktifkan akun staf',            'pengguna')
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p ON TRUE
WHERE r.code = 'admin'
  AND p.code IN ('pengguna.read', 'pengguna.create', 'pengguna.update',
                 'pengguna.reset-password', 'pengguna.toggle-aktif')
ON CONFLICT DO NOTHING;

ALTER TABLE users ADD COLUMN IF NOT EXISTS terakhir_masuk TIMESTAMPTZ NULL;
