-- Migrasi 0010 (Tim Fase 2 — Backup / Restore).
-- Tabel metadata backup. File dump-nya sendiri disimpan di
-- ./storage/backup/ (BUKAN di database, BUKAN di static file server);
-- tabel ini hanya mencatat metadata + jejak file yang ada di disk.
-- File backup TIDAK PERNAH ditimpa: nama file selalu unik
-- (backup-<YYYYMMDD-HHMMSS>-<8 karakter acak>.dump).
-- jenis: 'manual' = dibuat via tombol "Buat Backup" / POST /api/v1/backup;
--        'pre-restore' = backup pengaman otomatis yang dibuat sistem
--        SEBELUM setiap restore. Pre-restore < 7 hari tidak boleh dihapus.
CREATE TABLE IF NOT EXISTS backup_meta (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nama_file TEXT NOT NULL UNIQUE,
    ukuran_bytes BIGINT NOT NULL DEFAULT 0 CHECK (ukuran_bytes >= 0),
    jenis TEXT NOT NULL CHECK (jenis IN ('manual', 'pre-restore')),
    dibuat_oleh UUID REFERENCES users (id) ON DELETE SET NULL,
    keterangan TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_backup_meta_created_at ON backup_meta (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_backup_meta_jenis ON backup_meta (jenis);
