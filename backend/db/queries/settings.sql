-- Kueri sqlc Tim 5: settings + audit log.
-- Tabel settings & audit_log dibuat di migrasi 0001 (Tim 1).
-- Dijalankan via `sqlc generate` (lihat backend/sqlc.yaml).

-- name: ListSettings :many
SELECT key, value, updated_at FROM settings ORDER BY key ASC;

-- name: GetSetting :one
SELECT key, value, updated_at FROM settings WHERE key = $1;

-- name: UpsertSetting :one
INSERT INTO settings (key, value, updated_at)
VALUES ($1, $2, now())
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()
RETURNING *;

-- name: ListAudit :many
SELECT id, actor_id, aksi, entitas, entitas_id, detail, ip, created_at
FROM audit_log
WHERE ($1::text IS NULL OR aksi = $1)
  AND ($2::text IS NULL OR entitas = $2)
  AND ($3::uuid IS NULL OR actor_id = $3)
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: CreateAudit :one
INSERT INTO audit_log (actor_id, aksi, entitas, entitas_id, detail, ip)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
