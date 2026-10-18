-- Migrasi 0002 (Tim 2: Master Data)
-- Bergantung pada migrasi 0001 (Tim 1) yang mendefinisikan:
--   users(id uuid)
-- Tabel ini dipakai migrasi lain:
--   0003 (Tim 3, Bank Soal): mapel(id), soal.kisi_kisi_id / soal.kompetensi_id
--     (kolom di soal, TANPA FK — FK ditambahkan menyusul bila disepakati),
--     bank_soal.mapel_id, bank_soal.guru_id -> users(id)
--   0005 (Tim 5, Rekap): siswa(id, nisn, nama, kelas_id), kelas(id, nama)
--
-- Kontrak import siswa (dipakai handler import):
--   users memakai kolom tunggal role TEXT ('siswa' untuk akun siswa):
--     nama TEXT NOT NULL, username TEXT UNIQUE NOT NULL,
--     password_hash TEXT NOT NULL, role TEXT NOT NULL,
--     must_change_password BOOLEAN NOT NULL DEFAULT TRUE,
--     is_active BOOLEAN NOT NULL DEFAULT TRUE.
--   SATU role per user — tidak ada tabel junction user_roles.
--
-- Aturan dari ANALISIS.md Bab 5.1:
--   - Mapel Produktif WAJIB punya jurusan_id; Mapel Umum WAJIB jurusan_id kosong.
--   - Satu siswa aktif di satu rombel (kelas) per tahun ajaran.
-- Soft-delete TIDAK dipakai: delete hard + ditolak (409) bila masih dipakai.
-- Password siswa baru di-hash Argon2id (format PHC standar) di handler.

CREATE TABLE IF NOT EXISTS jurusan (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kode TEXT NOT NULL UNIQUE,
    nama TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS kelas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    jurusan_id UUID NOT NULL REFERENCES jurusan (id) ON DELETE RESTRICT,
    tingkat INT NOT NULL CHECK (tingkat IN (10, 11, 12)),
    nama TEXT NOT NULL,
    tahun_ajaran TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (jurusan_id, tingkat, nama, tahun_ajaran)
);

CREATE INDEX IF NOT EXISTS idx_kelas_jurusan ON kelas (jurusan_id);

CREATE TABLE IF NOT EXISTS mapel (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kode TEXT NOT NULL UNIQUE,
    nama TEXT NOT NULL,
    jenis TEXT NOT NULL CHECK (jenis IN ('umum', 'produktif')),
    jurusan_id UUID REFERENCES jurusan (id) ON DELETE RESTRICT,
    kelas_tingkat INT CHECK (kelas_tingkat IN (10, 11, 12)),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- Produktif terikat satu jurusan; Umum berlaku semua jurusan (null).
    CHECK ((jenis = 'produktif' AND jurusan_id IS NOT NULL) OR
           (jenis = 'umum' AND jurusan_id IS NULL))
);

CREATE INDEX IF NOT EXISTS idx_mapel_jurusan ON mapel (jurusan_id);
CREATE INDEX IF NOT EXISTS idx_mapel_jenis ON mapel (jenis);

CREATE TABLE IF NOT EXISTS guru (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users (id) ON DELETE SET NULL,
    nip TEXT UNIQUE,
    nama TEXT NOT NULL,
    mapel_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_guru_user ON guru (user_id);

CREATE TABLE IF NOT EXISTS siswa (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users (id) ON DELETE SET NULL,
    nisn TEXT NOT NULL UNIQUE,
    nis TEXT,
    nama TEXT NOT NULL,
    kelas_id UUID NOT NULL REFERENCES kelas (id) ON DELETE RESTRICT,
    jurusan_id UUID NOT NULL REFERENCES jurusan (id) ON DELETE RESTRICT,
    tahun_masuk INT CHECK (tahun_masuk IS NULL OR (tahun_masuk BETWEEN 1900 AND 2100)),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_siswa_kelas ON siswa (kelas_id);
CREATE INDEX IF NOT EXISTS idx_siswa_jurusan ON siswa (jurusan_id);
CREATE INDEX IF NOT EXISTS idx_siswa_user ON siswa (user_id);
CREATE INDEX IF NOT EXISTS idx_siswa_nisn ON siswa (nisn);

-- Kompetensi (CP/TP): tabel referensi, BUKAN teks bebas (ANALISIS Bab 5.2).
-- Satu soal -> satu kompetensi utama (dihitung ke nilai per kompetensi).
CREATE TABLE IF NOT EXISTS kompetensi (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mapel_id UUID NOT NULL REFERENCES mapel (id) ON DELETE RESTRICT,
    kode TEXT NOT NULL,
    deskripsi TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (mapel_id, kode)
);

CREATE INDEX IF NOT EXISTS idx_kompetensi_mapel ON kompetensi (mapel_id);

-- Kisi-kisi (ANALISIS Bab 5.3): satu baris = satu kelompok soal sejenis
-- untuk satu mapel + satu tingkat. Soal wajib merujuk ke satu baris
-- kisi-kisi + satu kompetensi utama.
CREATE TABLE IF NOT EXISTS kisi_kisi (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mapel_id UUID NOT NULL REFERENCES mapel (id) ON DELETE RESTRICT,
    kelas_tingkat INT NOT NULL CHECK (kelas_tingkat IN (10, 11, 12)),
    kelompok TEXT NOT NULL,
    kompetensi_id UUID REFERENCES kompetensi (id) ON DELETE SET NULL,
    jumlah_soal INT NOT NULL CHECK (jumlah_soal > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_kisi_kisi_mapel ON kisi_kisi (mapel_id);
CREATE INDEX IF NOT EXISTS idx_kisi_kisi_kompetensi ON kisi_kisi (kompetensi_id);
CREATE INDEX IF NOT EXISTS idx_kisi_kisi_mapel_tingkat ON kisi_kisi (mapel_id, kelas_tingkat);
