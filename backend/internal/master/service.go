package master

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"cbt-smk-v2/backend/internal/audit"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/argon2"
)

// RequestActor adalah identitas pemanggil untuk guard + audit log.
// Dibangun handler dari echo.Context agar service bebas dari HTTP.
type RequestActor struct {
	UserID string
	RealIP string
}

// Service adalah lapisan logika bisnis paket master.
type Service struct {
	repo *Repository
	pool *pgxpool.Pool
}

// NewService membuat Service.
func NewService(repo *Repository, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, pool: pool}
}

// catat mencatat audit log; error diabaikan (tidak boleh menggagalkan
// operasi utama).
func (s *Service) catat(ctx context.Context, actor RequestActor, aksi, entitas, entitasID string, detail any) {
	_ = audit.Catat(ctx, s.pool, actor.UserID, aksi, entitas, entitasID, detail, actor.RealIP)
}

// ---------------------------------------------------------------------
// Password siswa: Argon2id, format PHC standar.
// Tim 1 (auth) HARUS memverifikasi dengan parser PHC yang sama.
// Parameter: m=64 MiB, t=3, p=2, salt 16 byte, hash 32 byte.
// ---------------------------------------------------------------------

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 2, 32)
	b64 := base64.RawStdEncoding
	return fmt.Sprintf("$argon2id$v=19$m=65536,t=3,p=2$%s$%s",
		b64.EncodeToString(salt), b64.EncodeToString(hash)), nil
}
