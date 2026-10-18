// Package monitor — monitor ujian real-time via WebSocket.
//
// Kontrak (tugas Tim 4):
//   - POST /api/v1/monitor/ticket {jadwal_id} -> ticket 30 detik
//     (disimpan di memori + jadwal_id + user).
//   - WS /ws/monitor?ticket= — validasi ticket + permission ujian.monitor,
//     subscribe room per jadwal_id, kirim snapshot awal, broadcast update
//     saat jawaban/pelanggaran/status berubah, heartbeat 30 dtk,
//     reconnect -> snapshot lagi.
//   - GET /api/v1/monitor/:jadwalId — snapshot REST (fallback polling).
//   - Broadcast(jadwalID, msg) — dipanggil paket exam saat ada perubahan.
//
// Pola Hub: satu Hub global (defaultHub) berisi room per jadwal_id.
// Klien yang disconnect dihapus dari room. Ticket sekali pakai
// (dihapus begitu dipakai) dan kedaluwarsa 30 detik.
//
// Struktur layer (hasil refactor arsitektur):
//   - monitor.go:    dokumentasi paket
//   - handler.go:    Echo handler + Register + cek permission
//   - service.go:    Hub + ticket store + logika ticket/snapshot
//   - repository.go: DTO + query snapshot
//   - platform:      envelope JSON (menggantikan helper lokal)
//
// Perilaku (route, status, bentuk JSON, pesan error) IDENTIK dengan
// sebelum refactor. Signature Broadcast DIPERTAHANKAN (dipakai exam).
package monitor

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/config"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/coder/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// --- Handler ----------------------------------------------------------------

// Handler adalah lapisan HTTP paket monitor.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// bolehMonitor: admin/proktor/guru/pengawas (matriks ANALISIS Bab 11)
// atau pemegang permission ujian.monitor.
// Membaca role langsung dari context (middleware auth Tim 1) — tidak
// memakai paket audit Tim 5 (signature Echo v4, belum compile di Echo v5).
// (Dipindah dari monitor.go tanpa perubahan.)
func bolehMonitor(c *echo.Context) bool {
	uid, _ := c.Get("auth_user_id").(string)
	if uid == "" {
		return false
	}
	role, _ := c.Get("auth_role").(string)
	switch role {
	case "admin", "proktor", "guru", "pengawas":
		return true
	}
	perms, _ := c.Get("auth_permissions").([]string)
	for _, p := range perms {
		if strings.TrimSpace(p) == "ujian.monitor" {
			return true
		}
	}
	return false
}

// Register mendaftarkan endpoint monitor.
//
//	POST /api/v1/monitor/ticket {jadwal_id}
//	GET  /api/v1/monitor/:jadwalId
//	WS   /ws/monitor?ticket=
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q
	_ = cfg

	repo := NewRepository(pool)
	svc := NewService(repo, pool)
	h := NewHandler(svc)

	// POST /api/v1/monitor/ticket
	e.POST("/api/v1/monitor/ticket", h.Ticket)

	// GET /api/v1/monitor/:jadwalId — snapshot REST (fallback polling).
	e.GET("/api/v1/monitor/:jadwalId", h.Snapshot)

	// WS /ws/monitor?ticket=
	e.GET("/ws/monitor", h.WS)
}

// Ticket menangani POST /api/v1/monitor/ticket.
func (h *Handler) Ticket(c *echo.Context) error {
	if !bolehMonitor(c) {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin memantau ujian.")
	}
	var body struct {
		JadwalID string `json:"jadwal_id"`
	}
	if err := c.Bind(&body); err != nil || body.JadwalID == "" {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "jadwal_id wajib diisi.")
	}
	userID, _ := c.Get("auth_user_id").(string)
	role, _ := c.Get("auth_role").(string)
	token, exp, appErr := h.svc.BuatTicket(c.Request().Context(), body.JadwalID, userID, role)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, map[string]any{
		"ticket":     token,
		"jadwal_id":  body.JadwalID,
		"expires_in": 30,
		"expires_at": exp.Format(time.RFC3339),
	})
}

// Snapshot menangani GET /api/v1/monitor/:jadwalId.
func (h *Handler) Snapshot(c *echo.Context) error {
	if !bolehMonitor(c) {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin memantau ujian.")
	}
	snap, appErr := h.svc.Snapshot(c.Request().Context(), c.Param("jadwalId"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, snap)
}

// WS menangani WS /ws/monitor?ticket=.
func (h *Handler) WS(c *echo.Context) error {
	token := c.QueryParam("ticket")
	info, ok := h.svc.PakaiTicket(token)
	if !ok {
		return platform.WriteError(c, http.StatusUnauthorized, "unauthorized", "Ticket monitor tidak valid atau kedaluwarsa.")
	}
	// Ticket sudah memvalidasi permission saat diterbitkan (oleh
	// pemegang sesi login). Peran dicatat di ticket; tolak bila
	// peran bukan pemantau.
	switch info.Role {
	case "admin", "proktor", "guru", "pengawas":
	default:
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin memantau ujian.")
	}

	conn, err := websocket.Accept(c.Response(), c.Request(), &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		return err
	}
	defer conn.Close(websocket.StatusNormalClosure, "selesai")

	ctx := c.Request().Context()
	cl := &client{hub: defaultHub, jadwalID: info.JadwalID, conn: conn, send: make(chan []byte, 32)}
	defaultHub.subscribe(cl)
	defer defaultHub.unsubscribe(cl)

	// Kirim snapshot awal.
	if snap, appErr := h.svc.Snapshot(ctx, info.JadwalID); appErr == nil {
		if data, err := json.Marshal(map[string]any{"type": "snapshot", "data": snap}); err == nil {
			_ = conn.Write(ctx, websocket.MessageText, data)
		}
	}

	// Goroutine penulis: antrean kirim + heartbeat 30 detik.
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case data, ok := <-cl.send:
				if !ok {
					return
				}
				_ = conn.Write(ctx, websocket.MessageText, data)
			case <-ticker.C:
				if err := conn.Write(ctx, websocket.MessageText, []byte(`{"type":"ping"}`)); err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Loop baca: deteksi disconnect + pesan client ("snapshot" = minta
	// snapshot ulang, mis. setelah reconnect).
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return nil
		}
		var msg struct {
			Type string `json:"type"`
		}
		if json.Unmarshal(data, &msg) == nil && msg.Type == "snapshot" {
			if snap, appErr := h.svc.Snapshot(ctx, info.JadwalID); appErr == nil {
				if out, err := json.Marshal(map[string]any{"type": "snapshot", "data": snap}); err == nil {
					_ = conn.Write(ctx, websocket.MessageText, out)
				}
			}
		}
	}
}
