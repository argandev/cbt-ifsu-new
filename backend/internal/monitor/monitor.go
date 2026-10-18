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
package monitor

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/config"

	"github.com/coder/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// PesertaSnapshot adalah satu baris peserta di snapshot monitor.
type PesertaSnapshot struct {
	AttemptID   string   `json:"attempt_id"`
	SiswaID     string   `json:"siswa_id"`
	NamaSiswa   string   `json:"nama_siswa"`
	NISN        string   `json:"nisn"`
	KelasNama   string   `json:"kelas_nama"`
	Status      string   `json:"status"`
	MulaiAt     any      `json:"mulai_at"`
	Deadline    any      `json:"deadline"`
	SisaWaktu   *int     `json:"sisa_waktu_detik"`
	NilaiAkhir  *float64 `json:"nilai_akhir"`
	Dijawab     int      `json:"dijawab"`
	Ragu        int      `json:"ragu"`
	Pelanggaran int      `json:"pelanggaran"`
	TotalSoal   int      `json:"total_soal"`
}

// Snapshot adalah isi pesan snapshot/update monitor.
type Snapshot struct {
	JadwalID  string            `json:"jadwal_id"`
	Generated string            `json:"generated_at"`
	Peserta   []PesertaSnapshot `json:"peserta"`
}

// --- Hub -------------------------------------------------------------------

// client adalah satu koneksi WebSocket.
type client struct {
	hub      *Hub
	jadwalID string
	conn     *websocket.Conn
	send     chan []byte
}

// Hub menyimpan room per jadwal_id.
type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[*client]bool
}

// NewHub membuat Hub kosong.
func NewHub() *Hub {
	return &Hub{rooms: map[string]map[*client]bool{}}
}

// defaultHub dipakai Broadcast() level paket.
var defaultHub = NewHub()

// Broadcast mengirim msg (di-marshal JSON) ke semua klien di room jadwalID.
// Aman dipanggil dari paket mana pun (mis. exam). Bila tidak ada klien,
// tidak melakukan apa-apa.
func Broadcast(jadwalID string, msg any) {
	defaultHub.Broadcast(jadwalID, msg)
}

// Broadcast mengirim pesan ke room jadwalID.
func (h *Hub) Broadcast(jadwalID string, msg any) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.mu.RLock()
	room := h.rooms[jadwalID]
	clients := make([]*client, 0, len(room))
	for c := range room {
		clients = append(clients, c)
	}
	h.mu.RUnlock()
	for _, c := range clients {
		select {
		case c.send <- data:
		default:
			// Buffer penuh: lewati klien lambat agar broadcast tidak macet.
		}
	}
}

func (h *Hub) subscribe(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[c.jadwalID] == nil {
		h.rooms[c.jadwalID] = map[*client]bool{}
	}
	h.rooms[c.jadwalID][c] = true
}

func (h *Hub) unsubscribe(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if room := h.rooms[c.jadwalID]; room != nil {
		delete(room, c)
		if len(room) == 0 {
			delete(h.rooms, c.jadwalID)
		}
	}
}

// --- Ticket -----------------------------------------------------------------

type ticketInfo struct {
	JadwalID  string
	UserID    string
	Role      string
	ExpiresAt time.Time
}

type ticketStore struct {
	mu      sync.Mutex
	tickets map[string]ticketInfo
}

var tickets = &ticketStore{tickets: map[string]ticketInfo{}}

// buatTicket membuat ticket sekali pakai, berlaku 30 detik.
func (s *ticketStore) buat(jadwalID, userID, role string) (string, time.Time) {
	buf := make([]byte, 24)
	_, _ = rand.Read(buf)
	token := hex.EncodeToString(buf)
	exp := time.Now().Add(30 * time.Second)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tickets[token] = ticketInfo{JadwalID: jadwalID, UserID: userID, Role: role, ExpiresAt: exp}
	return token, exp
}

// pakai memvalidasi + mengonsumsi ticket (sekali pakai).
func (s *ticketStore) pakai(token string) (ticketInfo, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	info, ada := s.tickets[token]
	if !ada {
		return ticketInfo{}, false
	}
	delete(s.tickets, token)
	if time.Now().After(info.ExpiresAt) {
		return ticketInfo{}, false
	}
	return info, true
}

// --- Snapshot ---------------------------------------------------------------

// SnapshotPeserta membaca snapshot peserta satu jadwal dari DB.
func SnapshotPeserta(ctx context.Context, pool *pgxpool.Pool, jadwalID string) (Snapshot, error) {
	rows, err := pool.Query(ctx,
		`SELECT a.id, a.siswa_id, a.status, a.mulai_at, a.deadline,
		        a.sisa_waktu_beku, a.nilai_akhir,
		        COALESCE(s.nama, ''), COALESCE(s.nisn, ''), COALESCE(k.nama, ''),
		        (SELECT COUNT(*) FROM jawaban j
		          WHERE j.attempt_id = a.id AND j.jawaban IS NOT NULL),
		        (SELECT COUNT(*) FROM jawaban j
		          WHERE j.attempt_id = a.id AND j.ragu = TRUE),
		        (SELECT COUNT(*) FROM pelanggaran p WHERE p.attempt_id = a.id),
		        (SELECT COUNT(*) FROM paket_soal ps
		          JOIN jadwal_ujian ju ON ju.paket_id = ps.paket_id
		          WHERE ju.id = a.jadwal_id)
		 FROM attempt a
		 -- attempt.siswa_id merujuk users(id) (migrasi 0004), BUKAN
		 -- siswa(id): JOIN ke siswa HARUS lewat siswa.user_id
		 -- (migrasi 0002). JOIN lama (s.id = a.siswa_id) tidak
		 -- pernah cocok sehingga nama/NISN/kelas selalu kosong.
		 LEFT JOIN siswa s ON s.user_id = a.siswa_id
		 LEFT JOIN kelas k ON k.id = s.kelas_id
		 WHERE a.jadwal_id = $1
		 ORDER BY s.nama ASC`, jadwalID)
	if err != nil {
		return Snapshot{}, err
	}
	defer rows.Close()

	snap := Snapshot{JadwalID: jadwalID, Generated: time.Now().Format(time.RFC3339), Peserta: []PesertaSnapshot{}}
	for rows.Next() {
		var p PesertaSnapshot
		var sisaBeku *int
		var nilai *float64
		var deadline any
		if err := rows.Scan(&p.AttemptID, &p.SiswaID, &p.Status, &p.MulaiAt, &deadline,
			&sisaBeku, &nilai, &p.NamaSiswa, &p.NISN, &p.KelasNama,
			&p.Dijawab, &p.Ragu, &p.Pelanggaran, &p.TotalSoal); err != nil {
			return Snapshot{}, err
		}
		p.Deadline = deadline
		p.NilaiAkhir = nilai
		p.SisaWaktu = hitungSisa(deadline, sisaBeku)
		snap.Peserta = append(snap.Peserta, p)
	}
	return snap, rows.Err()
}

// hitungSisa menghitung sisa waktu (detik) dari deadline.
// Bila sisa dibekukan (pause), kembalikan nilai beku itu.
func hitungSisa(deadline any, sisaBeku *int) *int {
	if sisaBeku != nil {
		v := *sisaBeku
		if v < 0 {
			v = 0
		}
		return &v
	}
	t, ok := deadline.(time.Time)
	if !ok {
		return nil
	}
	sisa := int(time.Until(t).Seconds())
	if sisa < 0 {
		sisa = 0
	}
	return &sisa
}

// --- HTTP helpers -----------------------------------------------------------

func writeError(c *echo.Context, status int, code, message string) error {
	return c.JSON(status, map[string]any{
		"error": map[string]any{"code": code, "message": message},
	})
}

func writeOK(c *echo.Context, data any) error {
	return c.JSON(http.StatusOK, map[string]any{"data": data})
}

// bolehMonitor: admin/proktor/guru/pengawas (matriks ANALISIS Bab 11)
// atau pemegang permission ujian.monitor.
// Membaca role langsung dari context (middleware auth Tim 1) — tidak
// memakai paket audit Tim 5 (signature Echo v4, belum compile di Echo v5).
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

	// POST /api/v1/monitor/ticket
	e.POST("/api/v1/monitor/ticket", func(c *echo.Context) error {
		if !bolehMonitor(c) {
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin memantau ujian.")
		}
		var body struct {
			JadwalID string `json:"jadwal_id"`
		}
		if err := c.Bind(&body); err != nil || body.JadwalID == "" {
			return writeError(c, http.StatusBadRequest, "validation_error", "jadwal_id wajib diisi.")
		}
		var ada bool
		err := pool.QueryRow(c.Request().Context(),
			`SELECT EXISTS(SELECT 1 FROM jadwal_ujian WHERE id = $1)`, body.JadwalID).Scan(&ada)
		if err != nil || !ada {
			return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
		}
		userID, _ := c.Get("auth_user_id").(string)
		role, _ := c.Get("auth_role").(string)
		token, exp := tickets.buat(body.JadwalID, userID, role)
		return writeOK(c, map[string]any{
			"ticket":     token,
			"jadwal_id":  body.JadwalID,
			"expires_in": 30,
			"expires_at": exp.Format(time.RFC3339),
		})
	})

	// GET /api/v1/monitor/:jadwalId — snapshot REST (fallback polling).
	e.GET("/api/v1/monitor/:jadwalId", func(c *echo.Context) error {
		if !bolehMonitor(c) {
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin memantau ujian.")
		}
		snap, err := SnapshotPeserta(c.Request().Context(), pool, c.Param("jadwalId"))
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca snapshot monitor.")
		}
		return writeOK(c, snap)
	})

	// WS /ws/monitor?ticket=
	e.GET("/ws/monitor", func(c *echo.Context) error {
		token := c.QueryParam("ticket")
		info, ok := tickets.pakai(token)
		if !ok {
			return writeError(c, http.StatusUnauthorized, "unauthorized", "Ticket monitor tidak valid atau kedaluwarsa.")
		}
		// Ticket sudah memvalidasi permission saat diterbitkan (oleh
		// pemegang sesi login). Peran dicatat di ticket; tolak bila
		// peran bukan pemantau.
		switch info.Role {
		case "admin", "proktor", "guru", "pengawas":
		default:
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin memantau ujian.")
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
		if snap, err := SnapshotPeserta(ctx, pool, info.JadwalID); err == nil {
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
				if snap, err := SnapshotPeserta(ctx, pool, info.JadwalID); err == nil {
					if out, err := json.Marshal(map[string]any{"type": "snapshot", "data": snap}); err == nil {
						_ = conn.Write(ctx, websocket.MessageText, out)
					}
				}
			}
		}
	})
}
