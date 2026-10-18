package notifikasi

import (
	"context"

	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Service ----------------------------------------------------------------


// Service adalah lapisan logika bisnis paket notifikasi.
// repository adalah kontrak penyimpanan yang dipakai Service.
// (DIP: Service bergantung pada abstraksi, bukan *Repository konkrit.)
type repository interface {
	ListByPenerima(ctx context.Context, penerimaUserID string, limit, offset int) ([]map[string]any, error)
	TandaiDibaca(ctx context.Context, id, penerimaUserID string) (bool, error)
}

type Service struct {
	repo repository
	pool *pgxpool.Pool
}

// NewService membuat Service.
func NewService(repo repository, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, pool: pool}
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
func (s *Service) TandaiBaca(ctx context.Context, actor platform.RequestActor, id string) *platform.AppError {
	ada, err := s.repo.TandaiDibaca(ctx, id, actor.UserID)
	if err != nil {
		return platform.ErrInternal("internal_error", "Gagal menandai notifikasi dibaca.")
	}
	if !ada {
		return platform.ErrNotFound("not_found", "Notifikasi tidak ditemukan.")
	}
	// Audit: tandai-dibaca adalah mutasi (Fase 3, audit completeness).
	platform.Catat(ctx, s.pool, actor, "notifikasi.baca", "notifikasi", id, nil)
	return nil
}
