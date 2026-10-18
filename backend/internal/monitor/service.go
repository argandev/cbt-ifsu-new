package monitor

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	"cbt-smk-v2/backend/internal/platform"

	"github.com/coder/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Hub --------------------------------------------------------------------
// (Dipindah dari monitor.go tanpa perubahan.)

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
// Signature DIPERTAHANKAN — dipakai paket exam.
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
// (Dipindah dari monitor.go tanpa perubahan.)

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

// --- Service ----------------------------------------------------------------

// Service adalah lapisan logika bisnis paket monitor.
type Service struct {
	repo *Repository
	pool *pgxpool.Pool
}

// NewService membuat Service.
func NewService(repo *Repository, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, pool: pool}
}

// BuatTicket menerbitkan ticket monitor 30 detik untuk satu jadwal.
// Jadwal tidak ada -> 404. (Dulu: logika inline di handler ticket.)
func (s *Service) BuatTicket(ctx context.Context, jadwalID, userID, role string) (token string, exp time.Time, appErr *platform.AppError) {
	ada, err := s.repo.JadwalExists(ctx, jadwalID)
	if err != nil || !ada {
		return "", time.Time{}, platform.ErrNotFound("not_found", "Jadwal tidak ditemukan.")
	}
	token, exp = tickets.buat(jadwalID, userID, role)
	return token, exp, nil
}

// PakaiTicket memvalidasi + mengonsumsi ticket sekali pakai.
func (s *Service) PakaiTicket(token string) (ticketInfo, bool) {
	return tickets.pakai(token)
}

// Snapshot membaca snapshot peserta satu jadwal.
func (s *Service) Snapshot(ctx context.Context, jadwalID string) (Snapshot, *platform.AppError) {
	snap, err := s.repo.SnapshotPeserta(ctx, jadwalID)
	if err != nil {
		return Snapshot{}, platform.ErrInternal("internal_error", "Gagal membaca snapshot monitor.")
	}
	return snap, nil
}
