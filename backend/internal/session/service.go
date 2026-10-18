// Package session — tipe Session + verifikasi CSRF (layer service paket
// session; operasi DB ada di repository.go). Kode dipindah dari session.go
// tanpa perubahan.
package session

import (
	"crypto/subtle"
	"time"

	"github.com/google/uuid"
)

// Session adalah satu baris tabel sessions.
type Session struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Role       string
	CSRFToken  string
	IP         string
	UserAgent  string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	LastSeenAt time.Time
}

// VerifyCSRF membandingkan token CSRF dengan constant-time compare.
func VerifyCSRF(sessionCSRF, provided string) bool {
	if sessionCSRF == "" || provided == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(sessionCSRF), []byte(provided)) == 1
}
