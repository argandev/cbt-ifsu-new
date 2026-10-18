package notifikasi

import (
	"context"

	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Service ----------------------------------------------------------------

// RequestActor adalah identitas pemanggil untuk audit log.
// Dibangun handler dari echo.Context agar service bebas dari HTTP.
type RequestActor struct {
	UserID string
	RealIP string
}

// Service adalah lapisan logika bisnis paket notifikasi.
type Service struct {
	repo *Repository
	pool *pgxpool.Pool
}

// NewService membuat Service.
func NewService(repo *Repository, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, pool: pool}
}

// catat menulis audit log. (Dulu: audit.Catat di handler.)
func (s *Service) catat(ctx context.Context, actor RequestActor, aksi, entitas, entitasID string, detail any) {
	_ = audit.Catat(ctx, s.pool, actor.UserID, aksi, entitas, entitasID, detail, actor.RealIP)
}

// List mengembalikan notifikasi milik sendiri.
func (s *Service) List(ctx context.Context, userID string, limit, offset int) ([]map[string]any, *platform.AppError) {
	items, err := s.repo.ListByPenerima(ctx, userID, limit, offset)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca notifikasi.")
	}
	return items, nil
}

// TandaiBaca menandai satu notifikasi milik sendiri sebagai dibaca.
// Bukan milik / tidak ada -> 404. (Pesan dipertahankan.)
func (s *Service) TandaiBaca(ctx context.Context, actor RequestActor, id string) *platform.AppError {
	ada, err := s.repo.TandaiDibaca(ctx, id, actor.UserID)
	if err != nil {
		return platform.ErrInternal("internal_error", "Gagal menandai notifikasi dibaca.")
	}
	if !ada {
		return platform.ErrNotFound("not_found", "Notifikasi tidak ditemukan.")
	}
	// Audit: tandai-dibaca adalah mutasi (Fase 3, audit completeness).
	s.catat(ctx, actor, "notifikasi.baca", "notifikasi", id, nil)
	return nil
}
