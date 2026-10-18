// Package audit — pencatatan jejak audit (append-only) + endpoint baca.
//
// Audit log TIDAK PERNAH diubah/dihapus lewat API: hanya ada Catat (INSERT)
// dan GET daftar. Kontrak kolom tabel audit_log (migrasi 0001, Tim 1):
// id, actor_id, aksi, entitas, entitas_id, detail (jsonb), ip, created_at.
package audit

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// Kunci context yang diisi middleware auth Tim 1.
const (
	CtxUserID      = "auth_user_id"
	CtxRole        = "auth_role"
	CtxPermissions = "auth_permissions"
	CtxSiswaID     = "auth_siswa_id"
)

// Catat menulis satu baris audit log. actorID boleh kosong (aksi sistem).
// detail di-marshal menjadi JSONB. Gagal catat TIDAK boleh menggagalkan
// operasi utama, jadi pemanggil mengabaikan error ini (atau log saja).
func Catat(ctx context.Context, pool *pgxpool.Pool, actorID, aksi, entitas, entitasID string, detail any, ip string) error {
	var detailJSON []byte
	var err error
	if detail == nil {
		detailJSON = []byte("{}")
	} else {
		detailJSON, err = json.Marshal(detail)
		if err != nil {
			detailJSON = []byte("{}")
		}
	}
	var actor any
	if strings.TrimSpace(actorID) == "" {
		actor = nil
	} else {
		actor = actorID
	}
	_, err = pool.Exec(ctx,
		`INSERT INTO audit_log (actor_id, aksi, entitas, entitas_id, detail, ip)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		actor, aksi, entitas, entitasID, detailJSON, ip)
	return err
}

// CurrentUser membaca identitas dari context (diisi middleware Tim 1).
func CurrentUser(c *echo.Context) (userID, role string, permissions []string, siswaID string, ok bool) {
	uid, _ := c.Get(CtxUserID).(string)
	role, _ = c.Get(CtxRole).(string)
	siswaID, _ = c.Get(CtxSiswaID).(string)
	if siswaID == "" {
		siswaID = uid
	}
	switch p := c.Get(CtxPermissions).(type) {
	case []string:
		permissions = p
	case string:
		if p != "" {
			permissions = strings.Split(p, ",")
		}
	}
	if uid == "" {
		return "", "", nil, "", false
	}
	return uid, role, permissions, siswaID, true
}

// HasPermission memeriksa permission; role "admin" (EVIL) lolos semua.
func HasPermission(c *echo.Context, perm string) bool {
	_, role, permissions, _, ok := CurrentUser(c)
	if !ok {
		return false
	}
	if role == "admin" {
		return true
	}
	for _, p := range permissions {
		if strings.TrimSpace(p) == perm {
			return true
		}
	}
	return false
}

func writeError(c *echo.Context, status int, code, message string) error {
	return c.JSON(status, map[string]any{
		"error": map[string]any{"code": code, "message": message},
	})
}

func writeOK(c *echo.Context, data any) error {
	return c.JSON(http.StatusOK, map[string]any{"data": data})
}

func strPtr(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// Register mendaftarkan endpoint audit log.
//
//	GET /api/v1/audit?aksi=&entitas=&actor_id=&limit=&offset=
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q
	_ = cfg

	e.GET("/api/v1/audit", func(c *echo.Context) error {
		if !HasPermission(c, "audit.read") {
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat audit log.")
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
		aksi := c.QueryParam("aksi")
		entitas := c.QueryParam("entitas")
		actorID := c.QueryParam("actor_id")

		rows, err := pool.Query(c.Request().Context(),
			`SELECT id, actor_id, aksi, entitas, entitas_id, detail, ip, created_at
			 FROM audit_log
			 WHERE ($1::text IS NULL OR aksi = $1)
			   AND ($2::text IS NULL OR entitas = $2)
			   AND ($3::uuid IS NULL OR actor_id = $3)
			 ORDER BY created_at DESC
			 LIMIT $4 OFFSET $5`,
			strPtr(aksi), strPtr(entitas), strPtr(actorID), limit, offset)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca audit log.")
		}
		defer rows.Close()

		items := []map[string]any{}
		for rows.Next() {
			var id, aksiV, entitasV string
			var actorIDV, entitasIDV, ipV *string
			var detail json.RawMessage
			var createdAt any
			if err := rows.Scan(&id, &actorIDV, &aksiV, &entitasV, &entitasIDV, &detail, &ipV, &createdAt); err != nil {
				return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca audit log.")
			}
			var detailAny any
			if len(detail) > 0 {
				_ = json.Unmarshal(detail, &detailAny)
			}
			items = append(items, map[string]any{
				"id": id, "actor_id": actorIDV, "aksi": aksiV, "entitas": entitasV,
				"entitas_id": entitasIDV, "detail": detailAny, "ip": ipV, "created_at": createdAt,
			})
		}
		if err := rows.Err(); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca audit log.")
		}
		return writeOK(c, map[string]any{"items": items, "limit": limit, "offset": offset})
	})
}
