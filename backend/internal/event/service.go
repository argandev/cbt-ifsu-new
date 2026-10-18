package event

import (
	"context"


	"github.com/jackc/pgx/v5/pgxpool"
)


// Service adalah lapisan logika bisnis paket event.
// repository adalah kontrak penyimpanan yang dipakai Service.
// (DIP: Service bergantung pada abstraksi, bukan *Repository konkrit.)
type repository interface {
	AssignEventRuangan(ctx context.Context, jadwalID string, eventArg, ruanganArg any) error
	CountEvent(ctx context.Context, jenis, search any) (int, error)
	CountRuangan(ctx context.Context, search any) (int, error)
	CountSiswaAktifKelas(ctx context.Context, kelasID string) (int, error)
	DeleteEvent(ctx context.Context, id string) error
	DeleteRuangan(ctx context.Context, id string) error
	GetJadwalKelas(ctx context.Context, jadwalID string) (kelasID, kelasNama string, err error)
	InsertEvent(ctx context.Context, in eventInput, mulai, selesai any) (*eventRow, error)
	InsertRuangan(ctx context.Context, in ruanganInput) (*ruanganRow, error)
	ListEvent(ctx context.Context, jenis, search any, limit, offset int) ([]eventRow, error)
	ListRuangan(ctx context.Context, search any, limit, offset int) ([]ruanganRow, error)
	RuanganJadwalPenuh(ctx context.Context, ruanganID string, kapasitasBaru int) (string, error)
	UpdateEvent(ctx context.Context, id string, in eventInput, mulai, selesai any) error
	UpdateRuangan(ctx context.Context, id string, in ruanganInput) error
}

type Service struct {
	repo repository
	pool *pgxpool.Pool
}

// NewService membuat Service.
func NewService(repo repository, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, pool: pool}
}

