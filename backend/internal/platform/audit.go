// Package platform — identitas pemanggil + pencatatan audit bersama.
// Dipindah dari 10 paket (auth, master, banksoal, rekap, settings, exam,
// backup, ukk, analisis, event) agar satu definisi; perilaku/SQL IDENTIK.
package platform

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RequestActor adalah identitas pemanggil untuk guard + audit log.
// Dibangun handler dari echo.Context agar service bebas dari HTTP.
// (Superset: UserAgent hanya dipakai alur login.)
type RequestActor struct {
	UserID    string
	RealIP    string
	UserAgent string
	Role      string
}

// Catat menulis satu baris audit log (append-only) ke tabel audit_log.
// detail di-marshal menjadi JSONB; actorID kosong -> NULL.
// Gagal catat TIDAK boleh menggagalkan operasi utama: pemanggil
// mengabaikan error ini (atau log saja).
func Catat(ctx context.Context, pool *pgxpool.Pool, actor RequestActor, aksi, entitas, entitasID string, detail any) error {
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
	var actorID any
	if strings.TrimSpace(actor.UserID) == "" {
		actorID = nil
	} else {
		actorID = actor.UserID
	}
	_, err = pool.Exec(ctx,
		`INSERT INTO audit_log (actor_id, aksi, entitas, entitas_id, detail, ip)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		actorID, aksi, entitas, entitasID, detailJSON, actor.RealIP)
	return err
}
