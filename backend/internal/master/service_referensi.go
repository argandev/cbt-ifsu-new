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
// Kompetensi (guard: master.mapel.*)
// ---------------------------------------------------------------------

type kompetensiInput struct {
	MapelID   string `json:"mapel_id"`
	Kode      string `json:"kode"`
	Deskripsi string `json:"deskripsi"`
}

func (in kompetensiInput) validate() string {
	if !validUUID(in.MapelID) {
		return "Mapel wajib dipilih."
	}
	if strings.TrimSpace(in.Kode) == "" {
		return "Kode kompetensi wajib diisi."
	}
	return ""
}

// ListKompetensi mengembalikan halaman kompetensi sesuai filter.
func (s *Service) ListKompetensi(ctx context.Context, mapelID, q *string, limit, offset int) ([]kompetensiDTO, int64, *platform.AppError) {
	total, err := s.repo.CountKompetensi(ctx, mapelID, q)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca kompetensi.")
	}
	items, err := s.repo.ListKompetensi(ctx, mapelID, q, limit, offset)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca kompetensi.")
	}
	return items, total, nil
}

// GetKompetensi mengembalikan satu kompetensi.
func (s *Service) GetKompetensi(ctx context.Context, id string) (*kompetensiDTO, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Kompetensi tidak ditemukan.")
	}
	d, err := s.repo.GetKompetensi(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, platform.ErrNotFound("not_found", "Kompetensi tidak ditemukan.")
	}
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca kompetensi.")
	}
	return d, nil
}

// CreateKompetensi menambah kompetensi baru.
func (s *Service) CreateKompetensi(ctx context.Context, actor RequestActor, in kompetensiInput) (*kompetensiDTO, *platform.AppError) {
	if msg := in.validate(); msg != "" {
		return nil, platform.NewAppError("validation_error", msg, http.StatusUnprocessableEntity)
	}
	d, err := s.repo.InsertKompetensi(ctx, in.MapelID, strings.TrimSpace(in.Kode), in.Deskripsi)
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			return nil, platform.ErrConflict("conflict", conflictMessage(constraint))
		}
		if isFKViolation(err) {
			return nil, platform.NewAppError("validation_error", "Mapel tidak ditemukan.", http.StatusUnprocessableEntity)
		}
		return nil, platform.ErrInternal("internal_error", "Gagal menambah kompetensi.")
	}
	s.catat(ctx, actor, "create_kompetensi", "kompetensi", d.ID, map[string]any{"kode": d.Kode})
	return d, nil
}

// UpdateKompetensi mengubah kompetensi.
func (s *Service) UpdateKompetensi(ctx context.Context, actor RequestActor, id string, in kompetensiInput) (*kompetensiDTO, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Kompetensi tidak ditemukan.")
	}
	if msg := in.validate(); msg != "" {
		return nil, platform.NewAppError("validation_error", msg, http.StatusUnprocessableEntity)
	}
	d, err := s.repo.UpdateKompetensi(ctx, id, in.MapelID, strings.TrimSpace(in.Kode), in.Deskripsi)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, platform.ErrNotFound("not_found", "Kompetensi tidak ditemukan.")
	}
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			return nil, platform.ErrConflict("conflict", conflictMessage(constraint))
		}
		if isFKViolation(err) {
			return nil, platform.NewAppError("validation_error", "Mapel tidak ditemukan.", http.StatusUnprocessableEntity)
		}
		return nil, platform.ErrInternal("internal_error", "Gagal mengubah kompetensi.")
	}
	s.catat(ctx, actor, "update_kompetensi", "kompetensi", d.ID, map[string]any{"kode": d.Kode})
	return d, nil
}

// DeleteKompetensi menghapus kompetensi. Ditolak 409 bila masih dipakai
// kisi-kisi atau soal.
func (s *Service) DeleteKompetensi(ctx context.Context, actor RequestActor, id string) (map[string]any, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Kompetensi tidak ditemukan.")
	}
	guards := []struct {
		sql string
		msg string
	}{
		{`SELECT COUNT(*) FROM kisi_kisi WHERE kompetensi_id = $1`, "Kompetensi masih dipakai %d baris kisi-kisi."},
		{`SELECT COUNT(*) FROM soal WHERE kompetensi_id = $1`, "Kompetensi masih dipakai %d soal."},
	}
	for _, g := range guards {
		if n := s.repo.countOrZero(ctx, g.sql, id); n > 0 {
			return nil, platform.ErrConflict("conflict", fmt.Sprintf(g.msg, n))
		} else if n < 0 {
			return nil, platform.ErrInternal("internal_error", "Gagal memeriksa pemakaian kompetensi.")
		}
	}
	affected, err := s.repo.DeleteKompetensi(ctx, id)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menghapus kompetensi.")
	}
	if affected == 0 {
		return nil, platform.ErrNotFound("not_found", "Kompetensi tidak ditemukan.")
	}
	s.catat(ctx, actor, "delete_kompetensi", "kompetensi", id, nil)
	return map[string]any{"deleted": true, "id": id}, nil
}

// ---------------------------------------------------------------------
// Kisi-kisi (guard: master.mapel.*)
// ---------------------------------------------------------------------

type kisiKisiInput struct {
	MapelID      string  `json:"mapel_id"`
	KelasTingkat int32   `json:"kelas_tingkat"`
	Kelompok     string  `json:"kelompok"`
	KompetensiID *string `json:"kompetensi_id"`
	JumlahSoal   int32   `json:"jumlah_soal"`
}

func (in kisiKisiInput) validate() string {
	if !validUUID(in.MapelID) {
		return "Mapel wajib dipilih."
	}
	if in.KelasTingkat != 10 && in.KelasTingkat != 11 && in.KelasTingkat != 12 {
		return "Kelas tingkat harus 10, 11, atau 12."
	}
	if strings.TrimSpace(in.Kelompok) == "" {
		return "Kelompok kisi-kisi wajib diisi."
	}
	if in.KompetensiID != nil && strings.TrimSpace(*in.KompetensiID) != "" && !validUUID(*in.KompetensiID) {
		return "kompetensi_id tidak valid."
	}
	if in.JumlahSoal <= 0 {
		return "Jumlah soal harus lebih dari 0."
	}
	return ""
}

func kisiKompetensiArg(in kisiKisiInput) *string {
	if in.KompetensiID == nil {
		return nil
	}
	if strings.TrimSpace(*in.KompetensiID) == "" {
		return nil
	}
	return in.KompetensiID
}

// ListKisiKisi mengembalikan halaman kisi-kisi sesuai filter.
func (s *Service) ListKisiKisi(ctx context.Context, mapelID *string, tingkat *int, q *string, limit, offset int) ([]kisiKisiDTO, int64, *platform.AppError) {
	total, err := s.repo.CountKisiKisi(ctx, mapelID, tingkat, q)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca kisi-kisi.")
	}
	items, err := s.repo.ListKisiKisi(ctx, mapelID, tingkat, q, limit, offset)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca kisi-kisi.")
	}
	return items, total, nil
}

// GetKisiKisi mengembalikan satu kisi-kisi.
func (s *Service) GetKisiKisi(ctx context.Context, id string) (*kisiKisiDTO, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Kisi-kisi tidak ditemukan.")
	}
	d, err := s.repo.GetKisiKisi(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, platform.ErrNotFound("not_found", "Kisi-kisi tidak ditemukan.")
	}
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca kisi-kisi.")
	}
	return d, nil
}

// CreateKisiKisi menambah kisi-kisi baru.
func (s *Service) CreateKisiKisi(ctx context.Context, actor RequestActor, in kisiKisiInput) (*kisiKisiDTO, *platform.AppError) {
	if msg := in.validate(); msg != "" {
		return nil, platform.NewAppError("validation_error", msg, http.StatusUnprocessableEntity)
	}
	d, err := s.repo.InsertKisiKisi(ctx, in.MapelID, in.KelasTingkat, strings.TrimSpace(in.Kelompok), kisiKompetensiArg(in), in.JumlahSoal)
	if err != nil {
		if isFKViolation(err) {
			return nil, platform.NewAppError("validation_error", "Mapel/kompetensi tidak ditemukan.", http.StatusUnprocessableEntity)
		}
		return nil, platform.ErrInternal("internal_error", "Gagal menambah kisi-kisi.")
	}
	s.catat(ctx, actor, "create_kisi_kisi", "kisi_kisi", d.ID, map[string]any{"kelompok": d.Kelompok})
	return d, nil
}

// UpdateKisiKisi mengubah kisi-kisi.
func (s *Service) UpdateKisiKisi(ctx context.Context, actor RequestActor, id string, in kisiKisiInput) (*kisiKisiDTO, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Kisi-kisi tidak ditemukan.")
	}
	if msg := in.validate(); msg != "" {
		return nil, platform.NewAppError("validation_error", msg, http.StatusUnprocessableEntity)
	}
	d, err := s.repo.UpdateKisiKisi(ctx, id, in.MapelID, in.KelasTingkat, strings.TrimSpace(in.Kelompok), kisiKompetensiArg(in), in.JumlahSoal)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, platform.ErrNotFound("not_found", "Kisi-kisi tidak ditemukan.")
	}
	if err != nil {
		if isFKViolation(err) {
			return nil, platform.NewAppError("validation_error", "Mapel/kompetensi tidak ditemukan.", http.StatusUnprocessableEntity)
		}
		return nil, platform.ErrInternal("internal_error", "Gagal mengubah kisi-kisi.")
	}
	s.catat(ctx, actor, "update_kisi_kisi", "kisi_kisi", d.ID, map[string]any{"kelompok": d.Kelompok})
	return d, nil
}

// DeleteKisiKisi menghapus kisi-kisi. Ditolak 409 bila masih dipakai soal.
func (s *Service) DeleteKisiKisi(ctx context.Context, actor RequestActor, id string) (map[string]any, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Kisi-kisi tidak ditemukan.")
	}
	if n := s.repo.countOrZero(ctx, `SELECT COUNT(*) FROM soal WHERE kisi_kisi_id = $1`, id); n > 0 {
		return nil, platform.ErrConflict("conflict",
			fmt.Sprintf("Kisi-kisi masih dipakai %d soal. Hapus/ubah soal dulu.", n))
	} else if n < 0 {
		return nil, platform.ErrInternal("internal_error", "Gagal memeriksa pemakaian kisi-kisi.")
	}
	affected, err := s.repo.DeleteKisiKisi(ctx, id)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menghapus kisi-kisi.")
	}
	if affected == 0 {
		return nil, platform.ErrNotFound("not_found", "Kisi-kisi tidak ditemukan.")
	}
	s.catat(ctx, actor, "delete_kisi_kisi", "kisi_kisi", id, nil)
	return map[string]any{"deleted": true, "id": id}, nil
}
