-- =====================================================================
-- Migrasi 0004 (Tim 4: Mesin Ujian + Monitor)
-- Tabel: paket_ujian, paket_soal, jadwal_ujian, attempt, jawaban,
--        pelanggaran, usulan_unblock
-- =====================================================================
-- KONTRAK NAMA (dari tugas Tim 4 — WAJIB, jangan diubah diam-diam):
--   Paket/jadwal/attempt memakai nama SINGULAR sesuai spesifikasi:
--   paket_ujian, jadwal_ujian, attempt (BUKAN jadwal/attempts).
--
-- !!! KONFLIK ANTAR-TIM — PERLU KEPUTUSAN KOORDINATOR !!!
--   Migrasi 0005 (Tim 5, sudah ada) mengasumsikan tabel bernama
--   `jadwal` dan `attempts` (FK nilai_akhir -> attempts(id), dsb).
--   Nama itu TIDAK ADA di migrasi ini. Sebelum `migrate-up` penuh,
--   koordinator harus memilih SATU:
--     (a) Tim 5 mengubah 0005: jadwal -> jadwal_ujian, attempts -> attempt
--         (disarankan: hanya 1 berkas, 3 FK + 2 kueri), ATAU
--     (b) Tim 4 mengganti nama tabel di 0004 ini menjadi jadwal/attempts
--         (berarti menyimpang dari spesifikasi tugas Tim 4).
--   Sampai diputuskan, JANGAN menjalankan 0004 + 0005 bersamaan tanpa
--   penyesuaian — 0005 akan gagal karena tabel rujukannya tidak ada.
--
-- RUJUKAN KE TABEL TIM LAIN (asumsi, milik Tim 1/2/3):
--   users(id uuid)       — Tim 1, akun (siswa_id & pengawas_id merujuk ke sini
--                          bila profil siswa menyatu dengan akun; lihat catatan
--                          di bawah tentang siswa_id).
--   mapel(id uuid)       — Tim 2, master mapel.
--   kelas(id uuid)       — Tim 2, master kelas/rombel.
--   soal(id uuid, ...)   — Tim 3, bank soal. Kontrak kolom yang dibaca
--                          paket exam: id, tipe, pertanyaan, opsi (jsonb),
--                          kunci (jsonb), pembahasan, bobot (int).
--                          Bila Tim 3 memakai tabel opsi/pasangan terpisah,
--                          Tim 3 wajib menyediakan kolom-kolom tersebut
--                          (atau view) agar mesin ujian bisa membaca soal.
--   settings(key, value) — Tim 1/6, dibaca toleran (jsonb maupun text).
--   audit_log(...)       — Tim 1, ditulis via audit.Catat (Tim 5).
--
-- siswa_id: attempt.siswa_id menyimpan NILAI auth_siswa_id dari context
--   auth Tim 1 (audit.CurrentUser: siswa_id default = user_id). Bila
--   Tim 1 memakai siswa.id yang berbeda dari users.id, Tim 1 wajib
--   mengisi context auth_siswa_id dengan siswa.id yang benar.
--
-- ATURAN WAKTU (FINAL, dari tugas Tim 4):
--   - deadline attempt = jadwal_selesai (tanggal + jam_selesai jadwal).
--   - Pause MEMBEKUKAN timer: sisa_waktu_beku disimpan saat pause.
--   - Resume: deadline baru = now() + sisa_waktu_beku.
--   - Unblock = LANJUT sisa waktu (jawaban tersimpan, TIDAK reset).
--   - Token expiry default 900 detik dari jam mulai jadwal (baca
--     settings.token_expiry_detik bila ada; default 900).
--   - Toleransi telat default 15 menit (settings.toleransi_telat_menit).
--   - Auto-block: pelanggaran ke-1 = peringatan, ke-2 = auto-block
--     PER-ATTEMPT (settings.max_pelanggaran, default 2).
-- =====================================================================

CREATE TABLE IF NOT EXISTS paket_ujian (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kode TEXT NOT NULL UNIQUE,
    nama TEXT NOT NULL,
    mapel_id UUID NOT NULL REFERENCES mapel (id) ON DELETE RESTRICT,
    jenis TEXT NOT NULL CHECK (jenis IN ('STS', 'SAS', 'USBK', 'UKK', 'Tryout')),
    durasi_menit INT NOT NULL CHECK (durasi_menit > 0),
    jumlah_soal INT NOT NULL DEFAULT 0 CHECK (jumlah_soal >= 0),
    acak_soal BOOLEAN NOT NULL DEFAULT TRUE,
    acak_jawaban BOOLEAN NOT NULL DEFAULT TRUE,
    -- tampil_hasil = paket mengizinkan siswa melihat nilai.
    -- Nilai EFEKTIF yang tampil ke siswa = paket.tampil_hasil
    --   AND settings.tampil_hasil_ke_siswa (keduanya harus true).
    tampil_hasil BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_paket_ujian_mapel ON paket_ujian (mapel_id);
CREATE INDEX IF NOT EXISTS idx_paket_ujian_jenis ON paket_ujian (jenis);

CREATE TABLE IF NOT EXISTS paket_soal (
    paket_id UUID NOT NULL REFERENCES paket_ujian (id) ON DELETE CASCADE,
    soal_id UUID NOT NULL REFERENCES soal (id) ON DELETE RESTRICT,
    urutan INT NOT NULL DEFAULT 0,
    PRIMARY KEY (paket_id, soal_id)
);

CREATE INDEX IF NOT EXISTS idx_paket_soal_paket_urutan ON paket_soal (paket_id, urutan);

CREATE TABLE IF NOT EXISTS jadwal_ujian (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    paket_id UUID NOT NULL REFERENCES paket_ujian (id) ON DELETE RESTRICT,
    kelas_id UUID NOT NULL REFERENCES kelas (id) ON DELETE RESTRICT,
    tanggal DATE NOT NULL,
    jam_mulai TIME NOT NULL,
    jam_selesai TIME NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'aktif', 'pause', 'selesai')),
    token_hash TEXT,
    token_expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (jam_selesai > jam_mulai)
);

CREATE INDEX IF NOT EXISTS idx_jadwal_ujian_kelas_tanggal ON jadwal_ujian (kelas_id, tanggal);
CREATE INDEX IF NOT EXISTS idx_jadwal_ujian_paket ON jadwal_ujian (paket_id);
CREATE INDEX IF NOT EXISTS idx_jadwal_ujian_status ON jadwal_ujian (status);

CREATE TABLE IF NOT EXISTS attempt (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    jadwal_id UUID NOT NULL REFERENCES jadwal_ujian (id) ON DELETE RESTRICT,
    siswa_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    status TEXT NOT NULL DEFAULT 'berjalan'
        CHECK (status IN ('berjalan', 'selesai', 'diblokir')),
    mulai_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deadline TIMESTAMPTZ NOT NULL,
    -- Sisa waktu (detik) yang dibekukan saat pause / saat diblokir.
    -- NULL = timer tidak dibekukan.
    sisa_waktu_beku INT,
    -- Urutan soal per-attempt (stabil walau paket diacak):
    -- JSON array berisi soal_id sesuai urutan tampil siswa.
    urutan_soal JSONB NOT NULL DEFAULT '[]'::jsonb,
    -- Urutan opsi jawaban per soal (stabil per-attempt):
    -- JSON object { "<soal_id>": ["<opsi_id>", ...] }.
    urutan_opsi JSONB NOT NULL DEFAULT '{}'::jsonb,
    submit_at TIMESTAMPTZ,
    nilai_akhir NUMERIC(5, 2),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_attempt_jadwal_siswa ON attempt (jadwal_id, siswa_id);
CREATE INDEX IF NOT EXISTS idx_attempt_siswa_status ON attempt (siswa_id, status);
CREATE INDEX IF NOT EXISTS idx_attempt_jadwal_status ON attempt (jadwal_id, status);

CREATE TABLE IF NOT EXISTS jawaban (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attempt_id UUID NOT NULL REFERENCES attempt (id) ON DELETE CASCADE,
    soal_id UUID NOT NULL REFERENCES soal (id) ON DELETE RESTRICT,
    -- Isi jawaban terstruktur per tipe (lihat docs/TIPE-SOAL.md Tim 3):
    --   PG: "B" / {"pilihan":"B"} ; PG Kompleks: ["A","C"];
    --   Benar-Salah: {"p1": true, ...}; Menjodohkan: {"kiri1":"kanan2"};
    --   Isian: "teks"; Uraian: "teks panjang"; Mengurutkan: ["s1","s2"].
    -- NULL = soal belum dijawab (baris dibuat saat ragu-ragu ditandai).
    jawaban JSONB,
    ragu BOOLEAN NOT NULL DEFAULT FALSE,
    -- Poin otomatis (NULL untuk soal manual sebelum dikoreksi).
    poin NUMERIC(8, 2),
    dikoreksi BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (attempt_id, soal_id)
);

CREATE INDEX IF NOT EXISTS idx_jawaban_attempt ON jawaban (attempt_id);
CREATE INDEX IF NOT EXISTS idx_jawaban_attempt_dikoreksi ON jawaban (attempt_id, dikoreksi);

CREATE TABLE IF NOT EXISTS pelanggaran (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attempt_id UUID NOT NULL REFERENCES attempt (id) ON DELETE CASCADE,
    jenis TEXT NOT NULL,
    detail TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_pelanggaran_attempt_waktu ON pelanggaran (attempt_id, created_at DESC);

CREATE TABLE IF NOT EXISTS usulan_unblock (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attempt_id UUID NOT NULL REFERENCES attempt (id) ON DELETE CASCADE,
    pengawas_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    alasan TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'menunggu'
        CHECK (status IN ('menunggu', 'disetujui', 'ditolak')),
    diproses_oleh UUID REFERENCES users (id) ON DELETE SET NULL,
    diproses_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_usulan_unblock_attempt_status ON usulan_unblock (attempt_id, status);
