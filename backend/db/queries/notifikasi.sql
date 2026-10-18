-- Kueri sqlc Tim Fase 3 — Notifikasi (tabel notifikasi, migrasi 0012).
-- Dijalankan via `sqlc generate` (lihat backend/sqlc.yaml).
-- Handler Go memakai SQL yang sama melalui pgxpool; berkas ini sumber
-- kebenaran kueri. Notifikasi hanya dibaca pemiliknya (anti-IDOR:
-- UPDATE dibaca-dibatasi penerima_user_id; bukan milik -> 404).

-- name: CreateNotifikasi :one
INSERT INTO notifikasi (penerima_user_id, judul, isi)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListNotifikasiMilik :many
SELECT id, penerima_user_id, judul, isi, dibaca, created_at
FROM notifikasi
WHERE penerima_user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: TandaiNotifikasiDibaca :exec
UPDATE notifikasi SET dibaca = TRUE
WHERE id = $1 AND penerima_user_id = $2;
