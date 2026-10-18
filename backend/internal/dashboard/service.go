package dashboard

import (
	"context"

	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Service ----------------------------------------------------------------

// Service adalah lapisan logika bisnis paket dashboard.
type Service struct {
	repo *Repository
	pool *pgxpool.Pool
}

// NewService membuat Service.
func NewService(repo *Repository, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, pool: pool}
}

// Ringkasan mengembalikan hitungan agregat untuk kartu Dashboard.
// Pesan error dipertahankan (per tabel, sesuai urutan semula).
func (s *Service) Ringkasan(ctx context.Context) (map[string]any, *platform.AppError) {
	bankSoal, err := s.repo.Count(ctx, "bank_soal")
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menghitung bank soal.")
	}
	jadwal, err := s.repo.Count(ctx, "jadwal_ujian")
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menghitung jadwal ujian.")
	}
	peserta, err := s.repo.Count(ctx, "siswa")
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menghitung peserta.")
	}
	rekap, err := s.repo.Count(ctx, "nilai_akhir")
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menghitung rekap.")
	}
	return map[string]any{
		// Rekap = jumlah hasil ujian yang sudah direkap (baris nilai_akhir).
		"bank_soal":    bankSoal,
		"jadwal_ujian": jadwal,
		"peserta":      peserta,
		"rekap":        rekap,
	}, nil
}
