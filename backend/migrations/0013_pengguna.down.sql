-- Rollback migrasi 0013: cabut grant lalu hapus permission
-- pengguna.*, dan hapus kolom users.terakhir_masuk.
DELETE FROM role_permissions
WHERE permission_id IN (
  SELECT id FROM permissions
  WHERE code IN ('pengguna.read', 'pengguna.create', 'pengguna.update',
                 'pengguna.reset-password', 'pengguna.toggle-aktif')
);
DELETE FROM permissions
WHERE code IN ('pengguna.read', 'pengguna.create', 'pengguna.update',
               'pengguna.reset-password', 'pengguna.toggle-aktif');

ALTER TABLE users DROP COLUMN IF EXISTS terakhir_masuk;
