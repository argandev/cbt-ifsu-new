-- Migrasi 0007 (Tim Fase 2 — UKK).
-- Entitas UKK sendiri yang merujuk mapel + kelas + jurusan (keputusan
-- tugas Fase 2). Nilai teori BOLEH ditarik dari attempt ujian teori bila
-- ada, tetapi tidak dipaksakan: nilai_teori diisi penguji secara manual
-- lewat endpoint nilai (opsional).
--
-- Aturan hitung (FINAL, keputusan user; ditegakkan di backend, bukan di
-- CHECK database, agar pesan error dapat dibaca user):
--   nilai_praktik = nilai_proses*0.3 + nilai_produk*0.5 + nilai_sikap*0.2
--   nilai_akhir   = nilai_teori*0.4 + nilai_praktik*0.6
--   lulus: nilai_akhir >= 70
--   predikat: <70 Belum Tuntas, 70-79 Cukup, 80-90 Baik, 91-100 Sangat Baik
-- Sertifikat HANYA bagi yang lulus; nomor format
-- UKK/<tahun>/<kode-jurusan>/<nomor-urut 4 digit> dibuat backend.
-- Sertifikat di CBT adalah DATA + draf (bukan penerbitan hukum BNSP/LSP);
-- penerbitan resmi oleh lembaga berwenang (ANALISIS Bab 12.4).

CREATE TABLE IF NOT EXISTS ukk_ujian (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nama TEXT NOT NULL,
    mapel_id UUID NOT NULL REFERENCES mapel (id) ON DELETE RESTRICT,
    kelas_id UUID NOT NULL REFERENCES kelas (id) ON DELETE RESTRICT,
    jurusan_id UUID NOT NULL REFERENCES jurusan (id) ON DELETE RESTRICT,
    tanggal DATE,
    penguji_internal TEXT,
    penguji_eksternal TEXT,
    status TEXT NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'berjalan', 'selesai')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_ukk_ujian_jurusan ON ukk_ujian (jurusan_id);
CREATE INDEX IF NOT EXISTS idx_ukk_ujian_kelas ON ukk_ujian (kelas_id);
CREATE INDEX IF NOT EXISTS idx_ukk_ujian_status ON ukk_ujian (status);

CREATE TABLE IF NOT EXISTS ukk_nilai (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ukk_id UUID NOT NULL REFERENCES ukk_ujian (id) ON DELETE CASCADE,
    siswa_id UUID NOT NULL REFERENCES siswa (id) ON DELETE RESTRICT,
    nilai_teori NUMERIC(5, 2),
    nilai_proses NUMERIC(5, 2),
    nilai_produk NUMERIC(5, 2),
    nilai_sikap NUMERIC(5, 2),
    nilai_praktik NUMERIC(5, 2),
    nilai_akhir NUMERIC(5, 2),
    lulus BOOLEAN,
    predikat TEXT,
    catatan TEXT,
    dinilai_oleh UUID REFERENCES users (id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (ukk_id, siswa_id),
    CHECK (nilai_teori IS NULL OR (nilai_teori >= 0 AND nilai_teori <= 100)),
    CHECK (nilai_proses IS NULL OR (nilai_proses >= 0 AND nilai_proses <= 100)),
    CHECK (nilai_produk IS NULL OR (nilai_produk >= 0 AND nilai_produk <= 100)),
    CHECK (nilai_sikap IS NULL OR (nilai_sikap >= 0 AND nilai_sikap <= 100)),
    CHECK (nilai_praktik IS NULL OR (nilai_praktik >= 0 AND nilai_praktik <= 100)),
    CHECK (nilai_akhir IS NULL OR (nilai_akhir >= 0 AND nilai_akhir <= 100))
);

CREATE INDEX IF NOT EXISTS idx_ukk_nilai_ukk ON ukk_nilai (ukk_id);
CREATE INDEX IF NOT EXISTS idx_ukk_nilai_siswa ON ukk_nilai (siswa_id);

CREATE TABLE IF NOT EXISTS ukk_berkas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ukk_id UUID NOT NULL REFERENCES ukk_ujian (id) ON DELETE CASCADE,
    siswa_id UUID NOT NULL REFERENCES siswa (id) ON DELETE RESTRICT,
    nama_file TEXT NOT NULL,
    path TEXT NOT NULL,
    ukuran INT NOT NULL CHECK (ukuran >= 0),
    diupload_oleh UUID REFERENCES users (id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_ukk_berkas_ukk_siswa ON ukk_berkas (ukk_id, siswa_id);

CREATE TABLE IF NOT EXISTS ukk_sertifikat (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ukk_id UUID NOT NULL REFERENCES ukk_ujian (id) ON DELETE CASCADE,
    siswa_id UUID NOT NULL REFERENCES siswa (id) ON DELETE RESTRICT,
    nomor_sertifikat TEXT NOT NULL UNIQUE,
    diterbitkan_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (ukk_id, siswa_id)
);

CREATE INDEX IF NOT EXISTS idx_ukk_sertifikat_ukk ON ukk_sertifikat (ukk_id);
