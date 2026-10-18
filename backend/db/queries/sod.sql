-- Kueri sqlc Tim Fase 3 — SoD Override (tabel sod_overrides, migrasi 0012).
-- Dijalankan via `sqlc generate` (lihat backend/sqlc.yaml).
-- Handler Go memakai SQL yang sama melalui pgxpool; berkas ini sumber
-- kebenaran kueri. Pengecualian darurat SoD: ANALISIS Bab 11.4 +
-- docs/KONTRAK.md §5.4 (alasan tertulis min 20 karakter + audit dengan
-- flag SOD_OVERRIDE + notifikasi semua admin lain).

-- name: CreateSodOverride :one
INSERT INTO sod_overrides (actor_id, aksi, entitas, entitas_id, alasan)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListSodOverrides :many
SELECT id, actor_id, aksi, entitas, entitas_id, alasan, created_at
FROM sod_overrides
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
