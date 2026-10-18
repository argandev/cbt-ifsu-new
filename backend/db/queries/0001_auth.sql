-- Kueri sqlc (Tim 1: Auth + RBAC + Audit + Settings).
-- Dijalankan dengan `make sqlc` (butuh sqlc terinstall). Sampai sqlc
-- generate dijalankan, handler memakai pgxpool langsung dengan SQL yang
-- sama seperti di bawah (lihat backend/db/db.go — placeholder Tim 5).
-- CATATAN: `make sqlc` baru bisa sukses setelah migrasi 0002 (Tim 2,
-- tabel mapel dkk) ada, karena sqlc mem-parse SELURUH skema di
-- backend/migrations dan 0003 merujuk tabel mapel.

-- name: GetUserByLogin :one
SELECT id, nama, email, username, password_hash, role, must_change_password, is_active, created_at
FROM users
WHERE lower(email) = lower($1) OR lower(username) = lower($1);

-- name: GetUserByID :one
SELECT id, nama, email, username, password_hash, role, must_change_password, is_active, created_at
FROM users
WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (nama, email, username, password_hash, role, must_change_password, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, must_change_password = $3, updated_at = now()
WHERE id = $1;

-- name: GetSessionByID :one
SELECT id, user_id, role, csrf_token, ip, user_agent, created_at, expires_at, last_seen_at
FROM sessions
WHERE id = $1;

-- name: CreateSession :one
INSERT INTO sessions (user_id, role, csrf_token, ip, user_agent, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: TouchSession :exec
UPDATE sessions SET last_seen_at = now() WHERE id = $1;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = $1;

-- name: DeleteSessionsByUser :exec
DELETE FROM sessions WHERE user_id = $1;

-- name: ListPermissionsByRole :many
SELECT p.code
FROM role_permissions rp
JOIN permissions p ON p.id = rp.permission_id
JOIN roles r ON r.id = rp.role_id
WHERE r.code = $1
ORDER BY p.code;

-- name: CreateAuditLog :exec
INSERT INTO audit_log (actor_id, aksi, entitas, entitas_id, detail, ip)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ListAuditLog :many
SELECT id, actor_id, aksi, entitas, entitas_id, detail, ip, created_at
FROM audit_log
WHERE ($1::text IS NULL OR aksi = $1)
  AND ($2::text IS NULL OR entitas = $2)
  AND ($3::uuid IS NULL OR actor_id = $3)
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: GetSetting :one
SELECT key, value, updated_at FROM settings WHERE key = $1;

-- name: UpsertSetting :exec
INSERT INTO settings (key, value, updated_at) VALUES ($1, $2, now())
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now();
