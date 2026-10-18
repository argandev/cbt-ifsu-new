-- Rollback migrasi 0012: hapus tabel notifikasi + sod_overrides,
-- cabut grant lalu hapus permission sod.override.
DROP TABLE IF EXISTS notifikasi;
DROP TABLE IF EXISTS sod_overrides;

DELETE FROM role_permissions
WHERE permission_id IN (
  SELECT id FROM permissions WHERE code = 'sod.override'
);
DELETE FROM permissions
WHERE code = 'sod.override';
