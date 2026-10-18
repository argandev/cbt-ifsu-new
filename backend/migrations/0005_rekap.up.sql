-- Migrasi 0005 (Tim 5: Rekap + Berita Acara + Remedial)
-- Bergantung pada migrasi Tim 1/2/3/4 (0001-0004) yang mendefinisikan:
--   users(id uuid), jadwal_ujian(id uuid, paket_id, kelas_id),
--   paket_ujian(id uuid, mapel_id), siswa(id uuid, nisn, nama, kelas_id),
--   kelas(id uuid, nama), mapel(id uuid, nama), attempt(id uuid, jadwal_id, siswa_id, status),
--   jawaban(id, attempt_id), settings(key, value), audit_log(id, actor_id, aksi, entitas, entitas_id, detail, ip, created_at)
-- NAMA KANONIS (keputusan integrasi Fase 1): tabel ujian adalah
-- `attempt` (singular) dan `jadwal_ujian` (dari 0004). `jadwal_ujian`
-- TIDAK punya mapel_id langsung — mapel didapat via JOIN paket_ujian.
-- Nilai akhir TIDAK PERNAH dihapus fisik (aturan ANALISIS.md Bab 8): hanya UPDATE status/nilai + audit.

CREATE TABLE IF NOT EXISTS nilai_akhir (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attempt_id UUID NOT NULL UNIQUE REFERENCES attempt (id) ON DELETE RESTRICT,
    jadwal_id UUID NOT NULL REFERENCES jadwal_ujian (id) ON DELETE RESTRICT,
    siswa_id UUID NOT NULL REFERENCES siswa (id) ON DELETE RESTRICT,
    nilai NUMERIC(5, 2) NOT NULL DEFAULT 0 CHECK (nilai >= 0 AND nilai <= 100),
    status TEXT NOT NULL DEFAULT 'menunggu_koreksi'
        CHECK (status IN ('menunggu_koreksi', 'final')),
    predikat TEXT CHECK (predikat IS NULL OR predikat IN ('Belum Tuntas', 'Cukup', 'Baik', 'Sangat Baik')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_nilai_akhir_jadwal ON nilai_akhir (jadwal_id);
CREATE INDEX IF NOT EXISTS idx_nilai_akhir_siswa ON nilai_akhir (siswa_id);
CREATE INDEX IF NOT EXISTS idx_nilai_akhir_jadwal_status ON nilai_akhir (jadwal_id, status);

CREATE TABLE IF NOT EXISTS berita_acara (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    jadwal_id UUID NOT NULL UNIQUE REFERENCES jadwal_ujian (id) ON DELETE RESTRICT,
    isi JSONB NOT NULL DEFAULT '{}'::jsonb,
    dibuat_oleh UUID REFERENCES users (id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_berita_acara_jadwal ON berita_acara (jadwal_id);

CREATE TABLE IF NOT EXISTS remedial (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    jadwal_id UUID NOT NULL REFERENCES jadwal_ujian (id) ON DELETE RESTRICT,
    siswa_id UUID NOT NULL REFERENCES siswa (id) ON DELETE RESTRICT,
    nilai NUMERIC(5, 2) NOT NULL CHECK (nilai >= 0 AND nilai <= 100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_remedial_jadwal_siswa ON remedial (jadwal_id, siswa_id);
