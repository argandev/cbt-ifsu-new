package banksoal

import (
	"context"


	"github.com/jackc/pgx/v5/pgxpool"
)


// Service adalah lapisan logika bisnis paket banksoal.
// repository adalah kontrak penyimpanan yang dipakai Service.
// (DIP: Service bergantung pada abstraksi, bukan *Repository konkrit.)
type repository interface {
	BankDipakaiPaket(ctx context.Context, bankID string) (bool, error)
	CountBank(ctx context.Context, f BankFilter) (int, error)
	CountSoal(ctx context.Context, bankID string, f SoalFilter) (int, error)
	DeleteBank(ctx context.Context, id string) error
	DeleteSoal(ctx context.Context, soalID, bankID string) error
	InsertBank(ctx context.Context, kode, nama, mapelID string, kelasTingkat int, guruID any) (*bankRow, error)
	InsertSoal(ctx context.Context, bankID string, in soalInput, level, status string) (*soalRow, error)
	InsertSoalCSV(ctx context.Context, bankID string, row csvSoal) error
	ListBank(ctx context.Context, f BankFilter, limit, offset int) ([]bankRow, error)
	ListSoal(ctx context.Context, bankID string, f SoalFilter, limit, offset int) ([]soalRow, error)
	LoadBank(ctx context.Context, id string) (*bankRow, bool, error)
	LoadSoal(ctx context.Context, bankID, soalID string) (*soalRow, bool, error)
	SetBankLock(ctx context.Context, id string, locked bool) error
	UpdateBank(ctx context.Context, id, kode, nama, mapelID string, kelasTingkat int, guruID *string) error
	UpdateSoal(ctx context.Context, soalID, bankID string, in soalInput, level, status string) error
}

type Service struct {
	repo repository
	pool *pgxpool.Pool
}

// NewService membuat Service.
func NewService(repo repository, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, pool: pool}
}

