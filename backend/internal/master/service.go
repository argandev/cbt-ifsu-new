package master

import (
	"context"
	"encoding/json"
	"crypto/rand"
	"encoding/base64"
	"fmt"


	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/argon2"
)


// Service adalah lapisan logika bisnis paket master.
// repository adalah kontrak penyimpanan yang dipakai Service.
// (DIP: Service bergantung pada abstraksi, bukan *Repository konkrit.)
type repository interface {
	CountGuru(ctx context.Context, q *string) (int64, error)
	CountJurusan(ctx context.Context, q *string) (int64, error)
	CountKelas(ctx context.Context, jurusanID, q *string) (int64, error)
	CountKisiKisi(ctx context.Context, mapelID *string, tingkat *int, q *string) (int64, error)
	CountKompetensi(ctx context.Context, mapelID, q *string) (int64, error)
	CountMapel(ctx context.Context, jenis, jurusanID, q *string) (int64, error)
	CountSiswa(ctx context.Context, kelasID, jurusanID *string, isActive *bool, q *string) (int64, error)
	DeleteGuru(ctx context.Context, id string) (int64, error)
	DeleteJurusan(ctx context.Context, id string) (int64, error)
	DeleteKelas(ctx context.Context, id string) (int64, error)
	DeleteKisiKisi(ctx context.Context, id string) (int64, error)
	DeleteKompetensi(ctx context.Context, id string) (int64, error)
	DeleteMapel(ctx context.Context, id string) (int64, error)
	DeleteSiswa(ctx context.Context, id string) (int64, error)
	GetGuru(ctx context.Context, id string) (*guruDTO, error)
	GetGuruUserID(ctx context.Context, id string) *string
	GetJurusan(ctx context.Context, id string) (*jurusanDTO, error)
	GetKelas(ctx context.Context, id string) (*kelasDTO, error)
	GetKisiKisi(ctx context.Context, id string) (*kisiKisiDTO, error)
	GetKompetensi(ctx context.Context, id string) (*kompetensiDTO, error)
	GetMapel(ctx context.Context, id string) (*mapelDTO, error)
	GetSiswa(ctx context.Context, id string) (*siswaDTO, error)
	GetSiswaUserID(ctx context.Context, id string) *string
	InsertGuru(ctx context.Context, userID, nip any, nama string, mapelIDs json.RawMessage) (*guruDTO, error)
	InsertJurusan(ctx context.Context, kode, nama string) (*jurusanDTO, error)
	InsertKelas(ctx context.Context, jurusanID string, tingkat int32, nama, tahunAjaran string) (*kelasDTO, error)
	InsertKisiKisi(ctx context.Context, mapelID string, kelasTingkat int32, kelompok string, kompetensiID *string, jumlahSoal int32) (*kisiKisiDTO, error)
	InsertKompetensi(ctx context.Context, mapelID, kode, deskripsi string) (*kompetensiDTO, error)
	InsertMapel(ctx context.Context, kode, nama, jenis string, jurusanID *string, kelasTingkat *int32) (*mapelDTO, error)
	InsertSiswaImportTx(ctx context.Context, tx pgx.Tx, userID, nisn string, nisArg any, nama, kelasID, jurusanID string) error
	InsertSiswaTx(ctx context.Context, tx pgx.Tx, userIDArg any, nisn string, nis *string, nama, kelasID, jurusanID string, tahunMasuk *int32, isActive bool) (*siswaDTO, error)
	ListGuru(ctx context.Context, q *string, limit, offset int) ([]guruDTO, error)
	ListJurusan(ctx context.Context, q *string, limit, offset int) ([]jurusanDTO, error)
	ListKelas(ctx context.Context, jurusanID, q *string, limit, offset int) ([]kelasDTO, error)
	ListKisiKisi(ctx context.Context, mapelID *string, tingkat *int, q *string, limit, offset int) ([]kisiKisiDTO, error)
	ListKompetensi(ctx context.Context, mapelID, q *string, limit, offset int) ([]kompetensiDTO, error)
	ListMapel(ctx context.Context, jenis, jurusanID, q *string, limit, offset int) ([]mapelDTO, error)
	ListSiswa(ctx context.Context, kelasID, jurusanID *string, isActive *bool, q *string, limit, offset int) ([]siswaDTO, error)
	LookupKelasTx(ctx context.Context, tx pgx.Tx, kelasVal string) (kelasID, jurusanID string, err error)
	SiswaNISNExistsTx(ctx context.Context, tx pgx.Tx, nisn string) (bool, error)
	UpdateGuru(ctx context.Context, id string, userID, nip any, nama string, mapelIDs json.RawMessage) (*guruDTO, error)
	UpdateJurusan(ctx context.Context, id, kode, nama string) (*jurusanDTO, error)
	UpdateKelas(ctx context.Context, id, jurusanID string, tingkat int32, nama, tahunAjaran string) (*kelasDTO, error)
	UpdateKisiKisi(ctx context.Context, id, mapelID string, kelasTingkat int32, kelompok string, kompetensiID *string, jumlahSoal int32) (*kisiKisiDTO, error)
	UpdateKompetensi(ctx context.Context, id, mapelID, kode, deskripsi string) (*kompetensiDTO, error)
	UpdateMapel(ctx context.Context, id, kode, nama, jenis string, jurusanID *string, kelasTingkat *int32) (*mapelDTO, error)
	UpdateSiswa(ctx context.Context, id string, userIDArg any, nisn string, nis *string, nama, kelasID, jurusanID string, tahunMasuk *int32, isActive bool) (*siswaDTO, error)
	countOrZero(ctx context.Context, sql string, args ...any) int64
}

type Service struct {
	repo repository
	pool *pgxpool.Pool
}

// NewService membuat Service.
func NewService(repo repository, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, pool: pool}
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
