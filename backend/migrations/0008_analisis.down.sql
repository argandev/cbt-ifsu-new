-- Rollback migrasi 0008: cabut grant lalu hapus 2 permission analisis.
DELETE FROM role_permissions
WHERE permission_id IN (
  SELECT id FROM permissions WHERE code IN ('analisis.read', 'analisis.export')
);
DELETE FROM permissions
WHERE code IN ('analisis.read', 'analisis.export');
