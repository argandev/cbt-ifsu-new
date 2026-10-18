-- Migrasi 0008 (Tim Fase 2: Analisis Butir Soal + Nilai per Kompetensi).
-- HANYA menambah 2 permission baru + grant ke role. TIDAK mengubah
-- permission lain. Idempotent (ON CONFLICT DO NOTHING) sehingga aman
-- dijalankan ulang / setelah seed.
-- Sinkronisasi wajib (docs/KONTRAK.md §5): daftar yang sama harus ada di
-- backend/internal/rbac/rbac.go + backend/db/seed/seed.sql (di luar file
-- ini — koordinator menyelaraskan ketiga sumber kebenaran itu).
INSERT INTO permissions (code, description, group_name) VALUES
  ('analisis.read',   'Lihat analisis butir soal',       'analisis'),
  ('analisis.export', 'Export analisis butir soal ke CSV', 'analisis')
ON CONFLICT (code) DO NOTHING;

-- Grant: admin + proktor + guru.
-- (admin juga tercakup seed "semua permission"; baris ini eksplisit agar
-- migrasi ini berdiri sendiri tanpa bergantung pada urutan seed.)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p ON TRUE
WHERE r.code IN ('admin', 'proktor', 'guru')
  AND p.code IN ('analisis.read', 'analisis.export')
ON CONFLICT DO NOTHING;
