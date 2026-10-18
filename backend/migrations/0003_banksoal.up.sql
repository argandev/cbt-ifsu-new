-- Migrasi 0003 (Tim 3: Bank Soal)
-- Bergantung pada migrasi Tim 1/2 (0001-0002) yang mendefinisikan:
--   mapel(id uuid), users(id uuid)
-- Jika nama tabel rujukan dari tim lain berbeda, sesuaikan FK di bawah
-- sebelum migrate.
--
-- kisi_kisi_id / kompetensi_id SENGAJA tanpa FK: tabel kisi_kisi dan
-- kompetensi dibuat tim lain (belum ada saat migrasi ini ditulis).
-- FK ditambahkan menyusul setelah tabel tersebut ada. Kolom tetap
-- diindeks agar filter per kisi-kisi/kompetensi cepat.
--
-- KUNCI JAWABAN (kolom kunci) TIDAK PERNAH dikirim ke frontend siswa.
-- Endpoint siswa (Tim 4) wajib menghilangkannya; endpoint guru BOLEH
-- menampilkannya. Proktor TIDAK boleh melihat isi soal sama sekali
-- (README: "Proktor tidak melihat isi soal") — penegakan di handler.
--
-- Label jujur: 5 tipe inti AKM + 11 varian/ekstensi aplikasi.
-- TIDAK ada klaim "16 tipe standar Kemendikbud". Lihat docs/TIPE-SOAL.md.

CREATE TABLE IF NOT EXISTS bank_soal (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kode TEXT NOT NULL UNIQUE,
    nama TEXT NOT NULL,
    mapel_id UUID NOT NULL REFERENCES mapel (id) ON DELETE RESTRICT,
    kelas_tingkat INT NOT NULL CHECK (kelas_tingkat IN (10, 11, 12)),
    guru_id UUID REFERENCES users (id) ON DELETE SET NULL,
    is_locked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_bank_soal_mapel ON bank_soal (mapel_id);
CREATE INDEX IF NOT EXISTS idx_bank_soal_guru ON bank_soal (guru_id);
CREATE INDEX IF NOT EXISTS idx_bank_soal_locked ON bank_soal (is_locked);

CREATE TABLE IF NOT EXISTS soal (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bank_id UUID NOT NULL REFERENCES bank_soal (id) ON DELETE CASCADE,
    tipe TEXT NOT NULL CHECK (tipe IN (
        'pg', 'pg_kompleks', 'menjodohkan', 'isian', 'uraian',
        'benar_salah', 'pg_multi', 'urutkan', 'isi_ganda', 'teks_rumpang',
        'peta_konsep', 'studi_kasus', 'pernyataan_sikap', 'upload_file',
        'praktik_singkat', 'pilihan_gambar'
    )),
    kisi_kisi_id UUID,
    kompetensi_id UUID,
    stimulus TEXT NOT NULL DEFAULT '',
    pertanyaan TEXT NOT NULL,
    media_url TEXT NOT NULL DEFAULT '',
    jawaban JSONB NOT NULL,
    kunci JSONB NOT NULL,
    bobot INT NOT NULL DEFAULT 1 CHECK (bobot > 0),
    level TEXT NOT NULL DEFAULT 'sedang'
        CHECK (level IN ('mudah', 'sedang', 'sulit')),
    status TEXT NOT NULL DEFAULT 'aktif'
        CHECK (status IN ('draft', 'aktif', 'nonaktif', 'perlu_revisi', 'diarsipkan')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_soal_bank ON soal (bank_id);
CREATE INDEX IF NOT EXISTS idx_soal_bank_tipe ON soal (bank_id, tipe);
CREATE INDEX IF NOT EXISTS idx_soal_bank_level ON soal (bank_id, level);
CREATE INDEX IF NOT EXISTS idx_soal_kisi ON soal (kisi_kisi_id);
CREATE INDEX IF NOT EXISTS idx_soal_kompetensi ON soal (kompetensi_id);
