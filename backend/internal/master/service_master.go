package master

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5"
)

// ---------------------------------------------------------------------
// Jurusan
// ---------------------------------------------------------------------

type jurusanInput struct {
	Kode string `json:"kode"`
	Nama string `json:"nama"`
}

func (in jurusanInput) validate() string {
	if strings.TrimSpace(in.Kode) == "" {
		return "Kode jurusan wajib diisi."
	}
	if strings.TrimSpace(in.Nama) == "" {
		return "Nama jurusan wajib diisi."
	}
	return ""
}

// ListJurusan mengembalikan halaman jurusan sesuai pencarian.
func (s *Service) ListJurusan(ctx context.Context, q *string, limit, offset int) ([]jurusanDTO, int64, *platform.AppError) {
	total, err := s.repo.CountJurusan(ctx, q)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca jurusan.")
	}
	items, err := s.repo.ListJurusan(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca jurusan.")
	}
	return items, total, nil
}

// GetJurusan mengembalikan satu jurusan.
func (s *Service) GetJurusan(ctx context.Context, id string) (*jurusanDTO, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Jurusan tidak ditemukan.")
	}
	d, err := s.repo.GetJurusan(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, platform.ErrNotFound("not_found", "Jurusan tidak ditemukan.")
	}
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca jurusan.")
	}
	return d, nil
}

// CreateJurusan menambah jurusan baru.
func (s *Service) CreateJurusan(ctx context.Context, actor RequestActor, in jurusanInput) (*jurusanDTO, *platform.AppError) {
	if msg := in.validate(); msg != "" {
		return nil, platform.NewAppError("validation_error", msg, http.StatusUnprocessableEntity)
	}
	d, err := s.repo.InsertJurusan(ctx, strings.TrimSpace(in.Kode), strings.TrimSpace(in.Nama))
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			return nil, platform.ErrConflict("conflict", conflictMessage(constraint))
		}
		return nil, platform.ErrInternal("internal_error", "Gagal menambah jurusan.")
	}
	s.catat(ctx, actor, "create_jurusan", "jurusan", d.ID, map[string]any{"kode": d.Kode, "nama": d.Nama})
	return d, nil
}

// UpdateJurusan mengubah jurusan.
func (s *Service) UpdateJurusan(ctx context.Context, actor RequestActor, id string, in jurusanInput) (*jurusanDTO, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Jurusan tidak ditemukan.")
	}
	if msg := in.validate(); msg != "" {
		return nil, platform.NewAppError("validation_error", msg, http.StatusUnprocessableEntity)
	}
	d, err := s.repo.UpdateJurusan(ctx, id, strings.TrimSpace(in.Kode), strings.TrimSpace(in.Nama))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, platform.ErrNotFound("not_found", "Jurusan tidak ditemukan.")
	}
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			return nil, platform.ErrConflict("conflict", conflictMessage(constraint))
		}
		return nil, platform.ErrInternal("internal_error", "Gagal mengubah jurusan.")
	}
	s.catat(ctx, actor, "update_jurusan", "jurusan", d.ID, map[string]any{"kode": d.Kode, "nama": d.Nama})
	return d, nil
}

// DeleteJurusan menghapus jurusan. Ditolak 409 bila masih dipakai
// kelas, mapel produktif, atau siswa.
func (s *Service) DeleteJurusan(ctx context.Context, actor RequestActor, id string) (map[string]any, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Jurusan tidak ditemukan.")
	}
	if n := s.repo.countOrZero(ctx, `SELECT COUNT(*) FROM kelas WHERE jurusan_id = $1`, id); n > 0 {
		return nil, platform.ErrConflict("conflict",
			fmt.Sprintf("Jurusan masih dipakai %d kelas. Pindahkan/hapus kelas dulu.", n))
	} else if n < 0 {
		return nil, platform.ErrInternal("internal_error", "Gagal memeriksa pemakaian jurusan.")
	}
	if n := s.repo.countOrZero(ctx, `SELECT COUNT(*) FROM mapel WHERE jurusan_id = $1`, id); n > 0 {
		return nil, platform.ErrConflict("conflict",
			fmt.Sprintf("Jurusan masih dipakai %d mapel produktif. Hapus/ubah mapel dulu.", n))
	} else if n < 0 {
		return nil, platform.ErrInternal("internal_error", "Gagal memeriksa pemakaian jurusan.")
	}
	if n := s.repo.countOrZero(ctx, `SELECT COUNT(*) FROM siswa WHERE jurusan_id = $1`, id); n > 0 {
		return nil, platform.ErrConflict("conflict",
			fmt.Sprintf("Jurusan masih dipakai %d siswa. Pindahkan siswa dulu.", n))
	} else if n < 0 {
		return nil, platform.ErrInternal("internal_error", "Gagal memeriksa pemakaian jurusan.")
	}
	affected, err := s.repo.DeleteJurusan(ctx, id)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menghapus jurusan.")
	}
	if affected == 0 {
		return nil, platform.ErrNotFound("not_found", "Jurusan tidak ditemukan.")
	}
	s.catat(ctx, actor, "delete_jurusan", "jurusan", id, nil)
	return map[string]any{"deleted": true, "id": id}, nil
}

// ---------------------------------------------------------------------
// Kelas
// ---------------------------------------------------------------------

type kelasInput struct {
	JurusanID   string `json:"jurusan_id"`
	Tingkat     int32  `json:"tingkat"`
	Nama        string `json:"nama"`
	TahunAjaran string `json:"tahun_ajaran"`
}

func (in kelasInput) validate() string {
	if !validUUID(in.JurusanID) {
		return "Jurusan wajib dipilih."
	}
	if in.Tingkat != 10 && in.Tingkat != 11 && in.Tingkat != 12 {
		return "Tingkat harus 10, 11, atau 12."
	}
	if strings.TrimSpace(in.Nama) == "" {
		return "Nama kelas wajib diisi."
	}
	if strings.TrimSpace(in.TahunAjaran) == "" {
		return "Tahun ajaran wajib diisi (mis. 2026/2027)."
	}
	return ""
}

// ListKelas mengembalikan halaman kelas sesuai filter.
func (s *Service) ListKelas(ctx context.Context, jurusanID, q *string, limit, offset int) ([]kelasDTO, int64, *platform.AppError) {
	total, err := s.repo.CountKelas(ctx, jurusanID, q)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca kelas.")
	}
	items, err := s.repo.ListKelas(ctx, jurusanID, q, limit, offset)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca kelas.")
	}
	return items, total, nil
}

// GetKelas mengembalikan satu kelas.
func (s *Service) GetKelas(ctx context.Context, id string) (*kelasDTO, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Kelas tidak ditemukan.")
	}
	d, err := s.repo.GetKelas(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, platform.ErrNotFound("not_found", "Kelas tidak ditemukan.")
	}
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca kelas.")
	}
	return d, nil
}

// CreateKelas menambah kelas baru.
func (s *Service) CreateKelas(ctx context.Context, actor RequestActor, in kelasInput) (*kelasDTO, *platform.AppError) {
	if msg := in.validate(); msg != "" {
		return nil, platform.NewAppError("validation_error", msg, http.StatusUnprocessableEntity)
	}
	d, err := s.repo.InsertKelas(ctx, in.JurusanID, in.Tingkat, strings.TrimSpace(in.Nama), strings.TrimSpace(in.TahunAjaran))
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			return nil, platform.ErrConflict("conflict", conflictMessage(constraint))
		}
		if isFKViolation(err) {
			return nil, platform.NewAppError("validation_error", "Jurusan tidak ditemukan.", http.StatusUnprocessableEntity)
		}
		return nil, platform.ErrInternal("internal_error", "Gagal menambah kelas.")
	}
	s.catat(ctx, actor, "create_kelas", "kelas", d.ID, map[string]any{"nama": d.Nama})
	return d, nil
}

// UpdateKelas mengubah kelas.
func (s *Service) UpdateKelas(ctx context.Context, actor RequestActor, id string, in kelasInput) (*kelasDTO, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Kelas tidak ditemukan.")
	}
	if msg := in.validate(); msg != "" {
		return nil, platform.NewAppError("validation_error", msg, http.StatusUnprocessableEntity)
	}
	d, err := s.repo.UpdateKelas(ctx, id, in.JurusanID, in.Tingkat, strings.TrimSpace(in.Nama), strings.TrimSpace(in.TahunAjaran))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, platform.ErrNotFound("not_found", "Kelas tidak ditemukan.")
	}
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			return nil, platform.ErrConflict("conflict", conflictMessage(constraint))
		}
		if isFKViolation(err) {
			return nil, platform.NewAppError("validation_error", "Jurusan tidak ditemukan.", http.StatusUnprocessableEntity)
		}
		return nil, platform.ErrInternal("internal_error", "Gagal mengubah kelas.")
	}
	s.catat(ctx, actor, "update_kelas", "kelas", d.ID, map[string]any{"nama": d.Nama})
	return d, nil
}

// DeleteKelas menghapus kelas. Ditolak 409 bila masih dipakai siswa
// atau jadwal ujian.
func (s *Service) DeleteKelas(ctx context.Context, actor RequestActor, id string) (map[string]any, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Kelas tidak ditemukan.")
	}
	if n := s.repo.countOrZero(ctx, `SELECT COUNT(*) FROM siswa WHERE kelas_id = $1`, id); n > 0 {
		return nil, platform.ErrConflict("conflict",
			fmt.Sprintf("Kelas masih dipakai %d siswa. Pindahkan siswa dulu.", n))
	} else if n < 0 {
		return nil, platform.ErrInternal("internal_error", "Gagal memeriksa pemakaian kelas.")
	}
	// Guard jadwal: kelas yang sudah dipakai jadwal ujian tidak boleh
	// dihapus (jadwal_ujian.kelas_id ON DELETE RESTRICT + aturan 409).
	if n := s.repo.countOrZero(ctx, `SELECT COUNT(*) FROM jadwal_ujian WHERE kelas_id = $1`, id); n > 0 {
		return nil, platform.ErrConflict("conflict",
			fmt.Sprintf("Kelas masih dipakai %d jadwal ujian. Hapus jadwal dulu.", n))
	} else if n < 0 {
		return nil, platform.ErrInternal("internal_error", "Gagal memeriksa pemakaian kelas.")
	}
	affected, err := s.repo.DeleteKelas(ctx, id)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menghapus kelas.")
	}
	if affected == 0 {
		return nil, platform.ErrNotFound("not_found", "Kelas tidak ditemukan.")
	}
	s.catat(ctx, actor, "delete_kelas", "kelas", id, nil)
	return map[string]any{"deleted": true, "id": id}, nil
}

// ---------------------------------------------------------------------
// Mapel
// ---------------------------------------------------------------------

type mapelInput struct {
	Kode         string  `json:"kode"`
	Nama         string  `json:"nama"`
	Jenis        string  `json:"jenis"`
	JurusanID    *string `json:"jurusan_id"`
	KelasTingkat *int32  `json:"kelas_tingkat"`
}

func (in mapelInput) validate() string {
	if strings.TrimSpace(in.Kode) == "" {
		return "Kode mapel wajib diisi."
	}
	if strings.TrimSpace(in.Nama) == "" {
		return "Nama mapel wajib diisi."
	}
	if in.Jenis != "umum" && in.Jenis != "produktif" {
		return "Jenis mapel harus 'umum' atau 'produktif'."
	}
	if in.Jenis == "produktif" {
		if in.JurusanID == nil || !validUUID(*in.JurusanID) {
			return "Mapel produktif wajib terikat satu jurusan."
		}
	}
	if in.Jenis == "umum" && in.JurusanID != nil && strings.TrimSpace(*in.JurusanID) != "" {
		return "Mapel umum tidak terikat jurusan (jurusan_id harus kosong)."
	}
	if in.KelasTingkat != nil && *in.KelasTingkat != 10 && *in.KelasTingkat != 11 && *in.KelasTingkat != 12 {
		return "Kelas tingkat harus 10, 11, atau 12."
	}
	return ""
}

// mapelJurusanArg mengubah jurusan_id input menjadi argumen SQL (nil bila kosong).
func mapelJurusanArg(in mapelInput) *string {
	if in.JurusanID == nil {
		return nil
	}
	if strings.TrimSpace(*in.JurusanID) == "" {
		return nil
	}
	return in.JurusanID
}

// ListMapel mengembalikan halaman mapel sesuai filter.
func (s *Service) ListMapel(ctx context.Context, jenis, jurusanID, q *string, limit, offset int) ([]mapelDTO, int64, *platform.AppError) {
	total, err := s.repo.CountMapel(ctx, jenis, jurusanID, q)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca mapel.")
	}
	items, err := s.repo.ListMapel(ctx, jenis, jurusanID, q, limit, offset)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca mapel.")
	}
	return items, total, nil
}

// GetMapel mengembalikan satu mapel.
func (s *Service) GetMapel(ctx context.Context, id string) (*mapelDTO, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Mapel tidak ditemukan.")
	}
	d, err := s.repo.GetMapel(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, platform.ErrNotFound("not_found", "Mapel tidak ditemukan.")
	}
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca mapel.")
	}
	return d, nil
}

// CreateMapel menambah mapel baru.
func (s *Service) CreateMapel(ctx context.Context, actor RequestActor, in mapelInput) (*mapelDTO, *platform.AppError) {
	if msg := in.validate(); msg != "" {
		return nil, platform.NewAppError("validation_error", msg, http.StatusUnprocessableEntity)
	}
	d, err := s.repo.InsertMapel(ctx, strings.TrimSpace(in.Kode), strings.TrimSpace(in.Nama), in.Jenis, mapelJurusanArg(in), in.KelasTingkat)
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			return nil, platform.ErrConflict("conflict", conflictMessage(constraint))
		}
		if isFKViolation(err) {
			return nil, platform.NewAppError("validation_error", "Jurusan tidak ditemukan.", http.StatusUnprocessableEntity)
		}
		return nil, platform.ErrInternal("internal_error", "Gagal menambah mapel.")
	}
	s.catat(ctx, actor, "create_mapel", "mapel", d.ID, map[string]any{"kode": d.Kode, "nama": d.Nama})
	return d, nil
}

// UpdateMapel mengubah mapel.
func (s *Service) UpdateMapel(ctx context.Context, actor RequestActor, id string, in mapelInput) (*mapelDTO, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Mapel tidak ditemukan.")
	}
	if msg := in.validate(); msg != "" {
		return nil, platform.NewAppError("validation_error", msg, http.StatusUnprocessableEntity)
	}
	d, err := s.repo.UpdateMapel(ctx, id, strings.TrimSpace(in.Kode), strings.TrimSpace(in.Nama), in.Jenis, mapelJurusanArg(in), in.KelasTingkat)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, platform.ErrNotFound("not_found", "Mapel tidak ditemukan.")
	}
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			return nil, platform.ErrConflict("conflict", conflictMessage(constraint))
		}
		if isFKViolation(err) {
			return nil, platform.NewAppError("validation_error", "Jurusan tidak ditemukan.", http.StatusUnprocessableEntity)
		}
		return nil, platform.ErrInternal("internal_error", "Gagal mengubah mapel.")
	}
	s.catat(ctx, actor, "update_mapel", "mapel", d.ID, map[string]any{"kode": d.Kode, "nama": d.Nama})
	return d, nil
}

// DeleteMapel menghapus mapel. Ditolak 409 bila masih dipakai kompetensi,
// kisi-kisi, bank soal, paket ujian, atau jadwal ujian (via paket).
func (s *Service) DeleteMapel(ctx context.Context, actor RequestActor, id string) (map[string]any, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Mapel tidak ditemukan.")
	}
	guards := []struct {
		sql string
		msg string
	}{
		{`SELECT COUNT(*) FROM kompetensi WHERE mapel_id = $1`, "Mapel masih dipakai %d kompetensi. Hapus kompetensi dulu."},
		{`SELECT COUNT(*) FROM kisi_kisi WHERE mapel_id = $1`, "Mapel masih dipakai %d baris kisi-kisi. Hapus kisi-kisi dulu."},
		{`SELECT COUNT(*) FROM bank_soal WHERE mapel_id = $1`, "Mapel masih dipakai %d bank soal. Hapus/arsipkan bank soal dulu."},
		{`SELECT COUNT(*) FROM paket_ujian WHERE mapel_id = $1`, "Mapel masih dipakai %d paket ujian. Hapus paket dulu."},
		{`SELECT COUNT(*) FROM jadwal_ujian j JOIN paket_ujian p ON p.id = j.paket_id WHERE p.mapel_id = $1`, "Mapel masih dipakai %d jadwal ujian (via paket). Hapus jadwal dulu."},
	}
	for _, g := range guards {
		if n := s.repo.countOrZero(ctx, g.sql, id); n > 0 {
			return nil, platform.ErrConflict("conflict", fmt.Sprintf(g.msg, n))
		} else if n < 0 {
			return nil, platform.ErrInternal("internal_error", "Gagal memeriksa pemakaian mapel.")
		}
	}
	affected, err := s.repo.DeleteMapel(ctx, id)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menghapus mapel.")
	}
	if affected == 0 {
		return nil, platform.ErrNotFound("not_found", "Mapel tidak ditemukan.")
	}
	s.catat(ctx, actor, "delete_mapel", "mapel", id, nil)
	return map[string]any{"deleted": true, "id": id}, nil
}
