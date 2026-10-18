-- Migrasi 0012 (Tim Fase 3 — Hardening Backend).
--
-- 1. Permission baru `sod.override` (pengecualian darurat SoD,
--    ANALISIS Bab 11.4 + KONTRAK §5.4). Grant ke admin BUKAN di sini:
--    seed (backend/db/seed/seed.sql) adalah sumber kebenaran grant —
--    admin mendapat SEMUA permission via CROSS JOIN, dan grant di
--    migrasi tidak dapat diandalkan karena tabel `roles` masih KOSONG
--    saat migrasi berjalan pada DB fresh (pelajaran BUG-01: grant
--    migrasi 0008 diam-diam no-op). INSERT role_permissions di bawah
--    ini dipertahankan sebagai jaring pengaman bila roles sudah ada
--    (mis. migrasi dijalankan pada DB yang sudah di-seed), dan
--    bersifat no-op bila roles kosong.
-- 2. Tabel `sod_overrides`: jejak tiap pengecualian SoD.
-- 3. Tabel `notifikasi`: pesan dalam aplikasi untuk user.
-- Semua pernyataan idempotent (IF NOT EXISTS / ON CONFLICT DO NOTHING).

INSERT INTO permissions (code, description, group_name) VALUES
  ('sod.override', 'Mencatat pengecualian darurat SoD (alasan tertulis + audit + notifikasi)', 'sod')
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p ON TRUE
WHERE r.code = 'admin'
  AND p.code = 'sod.override'
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS sod_overrides (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id UUID REFERENCES users (id) ON DELETE SET NULL,
    aksi TEXT NOT NULL,
    entitas TEXT NOT NULL,
    entitas_id TEXT NOT NULL DEFAULT '',
    alasan TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_sod_overrides_actor
    ON sod_overrides (actor_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_sod_overrides_entitas
    ON sod_overrides (entitas, entitas_id, created_at DESC);

CREATE TABLE IF NOT EXISTS notifikasi (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    penerima_user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    judul TEXT NOT NULL,
    isi TEXT NOT NULL DEFAULT '',
    dibaca BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_notifikasi_penerima
    ON notifikasi (penerima_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifikasi_penerima_dibaca
    ON notifikasi (penerima_user_id, dibaca, created_at DESC);
