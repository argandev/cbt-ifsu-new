package event

import (
	"context"

	"cbt-smk-v2/backend/internal/audit"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RequestActor adalah identitas pemanggil untuk guard + audit log.
// Dibangun handler dari echo.Context agar service bebas dari HTTP.
type RequestActor struct {
	UserID string
	RealIP string
}

// Service adalah lapisan logika bisnis paket event.
type Service struct {
	repo *Repository
	pool *pgxpool.Pool
}

// NewService membuat Service.
func NewService(repo *Repository, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, pool: pool}
}

// catat menulis audit log. (Dulu: audit.Catat(ctx, pool, userID, ...)
// di handler.)
func (s *Service) catat(ctx context.Context, actor RequestActor, aksi, entitas, entitasID string, detail any) {
	_ = audit.Catat(ctx, s.pool, actor.UserID, aksi, entitas, entitasID, detail, actor.RealIP)
}
