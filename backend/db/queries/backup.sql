-- Kueri sqlc Tim Fase 2 — Backup / Restore (tabel backup_meta, migrasi 0010).
-- Dijalankan via `sqlc generate` (lihat backend/sqlc.yaml).
-- Handler Go memakai SQL yang sama melalui pgxpool; berkas ini sumber
-- kebenaran kueri. Nama file dump TIDAK PERNAH diambil dari input user:
-- selalu dibaca dari baris backup_meta ini.

-- name: CreateBackupMeta :one
INSERT INTO backup_meta (nama_file, ukuran_bytes, jenis, dibuat_oleh, keterangan)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetBackupMeta :one
SELECT id, nama_file, ukuran_bytes, jenis, dibuat_oleh, keterangan, created_at
FROM backup_meta
WHERE id = $1;

-- name: ListBackupMeta :many
SELECT bm.id, bm.nama_file, bm.ukuran_bytes, bm.jenis, bm.dibuat_oleh,
       bm.keterangan, bm.created_at, u.nama AS pembuat_nama
FROM backup_meta bm
LEFT JOIN users u ON u.id = bm.dibuat_oleh
ORDER BY bm.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountBackupMeta :one
SELECT COUNT(*) FROM backup_meta;

-- name: DeleteBackupMeta :exec
DELETE FROM backup_meta WHERE id = $1;
