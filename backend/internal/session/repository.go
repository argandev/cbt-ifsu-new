// Package session — operasi DB session server-side (layer repository).
//
// TIDAK ADA kode JWT di proyek ini. Auth memakai session cookie:
// cookie `cbt_session` (HttpOnly) membawa ID session (UUID acak);
// data session (user_id, role, csrf_token, expiry) tinggal di server (DB).
// Restart server TIDAK membuat logout massal karena session ada di DB.
//
// Struktur layer (hasil refactor arsitektur):
//   - repository.go: Get, Create, Touch, Destroy, DestroyAllForUser — semua
//     sebagai fungsi package-level (bukan method) dengan signature IDENTIK,
//     karena dipanggil sebagai session.Get(ctx, pool, cookie) dari
//     middleware, auth/service.go, dan users/service.go (file ini)
//   - service.go:    tipe Session + VerifyCSRF
//
// Kode dipindah dari session.go tanpa perubahan.
package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TTL default session: 12 jam. Idle timeout diberlakukan via LastSeenAt
// oleh Get (sesi yang tidak disentuh > 2 jam dianggap kedaluwarsa).
const (
	DefaultTTL  = 12 * time.Hour
	IdleTimeout = 2 * time.Hour
)

var (
	// ErrNotFound berarti session tidak ada di DB.
	ErrNotFound = errors.New("session tidak ditemukan")
	// ErrExpired berarti session sudah kedaluwarsa / idle timeout.
	ErrExpired = errors.New("session kedaluwarsa")
)

// Create membuat session baru dan mengembalikan (sessionID, csrfToken, err).
// csrfToken adalah token CSRF acak yang terikat ke session ini.
func Create(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, role, ip, ua string) (string, string, error) {
	id := uuid.New()
	csrf, err := randomToken(32)
	if err != nil {
		return "", "", err
	}
	now := time.Now()
	_, err = pool.Exec(ctx, `
		INSERT INTO sessions (id, user_id, role, csrf_token, ip, user_agent, created_at, expires_at, last_seen_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		id, userID, role, csrf, ip, ua, now, now.Add(DefaultTTL), now)
	if err != nil {
		return "", "", err
	}
	return id.String(), csrf, nil
}

// Get memuat session berdasarkan ID string. Mengembalikan ErrNotFound bila
// tidak ada, ErrExpired bila melewati expires_at atau idle timeout.
// Get TIDAK memperbarui last_seen_at; pakai Touch untuk itu.
func Get(ctx context.Context, pool *pgxpool.Pool, sessionID string) (*Session, error) {
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, ErrNotFound
	}
	var s Session
	err = pool.QueryRow(ctx, `
		SELECT id, user_id, role, csrf_token, ip, user_agent, created_at, expires_at, last_seen_at
		FROM sessions WHERE id = $1`, id).
		Scan(&s.ID, &s.UserID, &s.Role, &s.CSRFToken, &s.IP, &s.UserAgent, &s.CreatedAt, &s.ExpiresAt, &s.LastSeenAt)
	if err != nil {
		return nil, ErrNotFound
	}
	now := time.Now()
	if now.After(s.ExpiresAt) || now.Sub(s.LastSeenAt) > IdleTimeout {
		// Bersihkan session yang kedaluwarsa agar tidak menumpuk.
		_ = Destroy(ctx, pool, sessionID)
		return nil, ErrExpired
	}
	return &s, nil
}

// Touch memperbarui last_seen_at (aktivitas). Dipanggil pada tiap request terautentikasi.
func Touch(ctx context.Context, pool *pgxpool.Pool, sessionID string) {
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return
	}
	_, _ = pool.Exec(ctx, `UPDATE sessions SET last_seen_at = now() WHERE id = $1`, id)
}

// Destroy menghapus satu session (logout).
func Destroy(ctx context.Context, pool *pgxpool.Pool, sessionID string) error {
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return ErrNotFound
	}
	_, err = pool.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	return err
}

// DestroyAllForUser menghapus SEMUA session milik user (dipakai saat
// pemblokiran akun / logout semua perangkat / reset keamanan).
func DestroyAllForUser(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) error {
	_, err := pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
