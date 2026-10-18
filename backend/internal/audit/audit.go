// Package audit — pencatatan jejak audit (append-only) + endpoint baca.
//
// Audit log TIDAK PERNAH diubah/dihapus lewat API: hanya ada Catat (INSERT)
// dan GET daftar. Kontrak kolom tabel audit_log (migrasi 0001, Tim 1):
// id, actor_id, aksi, entitas, entitas_id, detail (jsonb), ip, created_at.
//
// Struktur layer (hasil refactor arsitektur):
//   - audit.go:      API publik paket (Catat, CurrentUser, HasPermission,
//     konstanta context) — dipakai banyak paket lain,
//     signature TIDAK BOLEH berubah.
//   - handler.go:    Echo handler + Register (endpoint GET /api/v1/audit)
//   - service.go:    logika baca audit log
//   - repository.go: SQL baca audit_log
//   - platform:      envelope JSON (menggantikan helper lokal)
//
// Perilaku (route, status, bentuk JSON, pesan error) IDENTIK dengan
// sebelum refactor.
package audit

import (
	"context"
	"encoding/json"
	"strings"

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
