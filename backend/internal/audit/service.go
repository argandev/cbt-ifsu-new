package audit

import (
	"context"

	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Service ----------------------------------------------------------------

// Service adalah lapisan logika bisnis paket audit.
// repository adalah kontrak penyimpanan yang dipakai Service.
// (DIP: Service bergantung pada abstraksi, bukan *Repository konkrit.)
type repository interface {
	ListLog(ctx context.Context, aksi, entitas, actorID string, limit, offset int) ([]map[string]any, error)
}

type Service struct {
	repo repository
	pool *pgxpool.Pool
}

// NewService membuat Service.
func NewService(repo repository, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, pool: pool}
}

// ListLog membaca audit log sesuai filter. Pesan error dipertahankan.
func (s *Service) ListLog(ctx context.Context, aksi, entitas, actorID string, limit, offset int) ([]map[string]any, *platform.AppError) {
	items, err := s.repo.ListLog(ctx, aksi, entitas, actorID, limit, offset)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca audit log.")
	}
	return items, nil
}
