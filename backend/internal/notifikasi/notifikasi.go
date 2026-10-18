// Package notifikasi — notifikasi dalam aplikasi (Tim Fase 3).
//
// Tabel `notifikasi` (migrasi 0012): id, penerima_user_id, judul, isi,
// dibaca, created_at. Endpoint (Register):
//
//	GET  /api/v1/notifikasi                 (notifikasi milik sendiri, pagination)
//	POST /api/v1/notifikasi/:id/baca        (tandai dibaca; bukan milik -> 404)
//
// Helper Buat dipakai paket lain (mis. sod: notifikasi pengecualian SoD
// untuk semua admin kecuali actor). Notifikasi hanya dibaca pemiliknya
// (anti-IDOR, KONTRAK §5.3).
package notifikasi

import (
	"context"
	"net/http"
	"strconv"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

func writeError(c *echo.Context, status int, code, message string) error {
	return c.JSON(status, map[string]any{
		"error": map[string]any{"code": code, "message": message},
	})
}

func writeOK(c *echo.Context, data any) error {
	return c.JSON(http.StatusOK, map[string]any{"data": data})
}

// Buat menyisipkan satu notifikasi untuk penerima. Gagal buat TIDAK
// boleh menggagalkan operasi pemanggil (pemanggil mengabaikan error).
func Buat(ctx context.Context, pool *pgxpool.Pool, penerimaUserID, judul, isi string) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO notifikasi (penerima_user_id, judul, isi) VALUES ($1, $2, $3)`,
		penerimaUserID, judul, isi)
	return err
}

// Register mendaftarkan endpoint notifikasi.
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q
	_ = cfg

	// GET /api/v1/notifikasi — milik sendiri, pagination limit/offset.
	e.GET("/api/v1/notifikasi", func(c *echo.Context) error {
		userID, _, _, _, ok := audit.CurrentUser(c)
		if !ok {
			return writeError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
		}
		limit, err := strconv.Atoi(c.QueryParam("limit"))
		if err != nil || limit <= 0 {
			limit = 50
		}
		if limit > 200 {
			limit = 200
		}
		offset, err := strconv.Atoi(c.QueryParam("offset"))
		if err != nil || offset < 0 {
			offset = 0
		}
		rows, err := pool.Query(c.Request().Context(),
			`SELECT id, penerima_user_id, judul, isi, dibaca, created_at
			 FROM notifikasi WHERE penerima_user_id = $1
			 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
			userID, limit, offset)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca notifikasi.")
		}
		defer rows.Close()
		items := []map[string]any{}
		for rows.Next() {
			var id, penerima, judul, isi string
			var dibaca bool
			var createdAt any
			if err := rows.Scan(&id, &penerima, &judul, &isi, &dibaca, &createdAt); err != nil {
				return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca notifikasi.")
			}
			items = append(items, map[string]any{
				"id": id, "penerima_user_id": penerima, "judul": judul,
				"isi": isi, "dibaca": dibaca, "created_at": createdAt,
			})
		}
		if err := rows.Err(); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca notifikasi.")
		}
		return writeOK(c, map[string]any{"items": items, "limit": limit, "offset": offset})
	})

	// POST /api/v1/notifikasi/:id/baca — tandai dibaca; bukan milik -> 404.
	e.POST("/api/v1/notifikasi/:id/baca", func(c *echo.Context) error {
		userID, _, _, _, ok := audit.CurrentUser(c)
		if !ok {
			return writeError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
		}
		ct, err := pool.Exec(c.Request().Context(),
			`UPDATE notifikasi SET dibaca = TRUE WHERE id = $1 AND penerima_user_id = $2`,
			c.Param("id"), userID)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menandai notifikasi dibaca.")
		}
		if ct.RowsAffected() == 0 {
			return writeError(c, http.StatusNotFound, "not_found", "Notifikasi tidak ditemukan.")
		}
		// Audit: tandai-dibaca adalah mutasi (Fase 3, audit completeness).
		_ = audit.Catat(c.Request().Context(), pool, userID, "notifikasi.baca", "notifikasi", c.Param("id"), nil, c.RealIP())
		return writeOK(c, map[string]any{"dibaca": true, "id": c.Param("id")})
	})
}
