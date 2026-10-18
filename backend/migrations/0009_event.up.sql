-- =====================================================================
-- Migrasi 0009 (Tim Fase 2 — Event + Ruangan + Kartu Peserta)
-- Tabel: event, ruangan.
-- Kolom baru di jadwal_ujian (0004): event_id, ruangan_id (opsional,
-- ON DELETE SET NULL — menghapus event/ruangan TIDAK menghapus jadwal,
-- tetapi hapus via API tetap ditolak 409 bila masih dipakai jadwal).
-- =====================================================================

CREATE TABLE IF NOT EXISTS event (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kode TEXT NOT NULL UNIQUE,
    nama TEXT NOT NULL,
    jenis TEXT NOT NULL DEFAULT 'Lainnya'
        CHECK (jenis IN ('STS', 'SAS', 'USBK', 'UKK', 'Tryout', 'Lainnya')),
    tanggal_mulai DATE,
    tanggal_selesai DATE,
    deskripsi TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (tanggal_mulai IS NULL OR tanggal_selesai IS NULL OR tanggal_selesai >= tanggal_mulai)
);

CREATE INDEX IF NOT EXISTS idx_event_jenis ON event (jenis);
CREATE INDEX IF NOT EXISTS idx_event_tanggal_mulai ON event (tanggal_mulai);

CREATE TABLE IF NOT EXISTS ruangan (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kode TEXT NOT NULL UNIQUE,
    nama TEXT NOT NULL,
    kapasitas INT NOT NULL CHECK (kapasitas > 0),
    lokasi TEXT,
    keterangan TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_ruangan_kode ON ruangan (kode);

ALTER TABLE jadwal_ujian
    ADD COLUMN IF NOT EXISTS event_id UUID REFERENCES event (id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS ruangan_id UUID REFERENCES ruangan (id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_jadwal_ujian_event ON jadwal_ujian (event_id);
CREATE INDEX IF NOT EXISTS idx_jadwal_ujian_ruangan ON jadwal_ujian (ruangan_id);
