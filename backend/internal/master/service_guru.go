package master

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5"
)

// ---------------------------------------------------------------------
// Guru
// ---------------------------------------------------------------------

type guruInput struct {
	UserID   *string  `json:"user_id"`
	NIP      *string  `json:"nip"`
	Nama     string   `json:"nama"`
	MapelIDs []string `json:"mapel_ids"`
}

func (in guruInput) validate() string {
	if strings.TrimSpace(in.Nama) == "" {
		return "Nama guru wajib diisi."
	}
	if in.UserID != nil && strings.TrimSpace(*in.UserID) != "" && !validUUID(*in.UserID) {
		return "user_id tidak valid."
	}
	for _, m := range in.MapelIDs {
		if !validUUID(m) {
			return "Daftar mapel berisi ID yang tidak valid."
		}
	}
	return ""
}

// guruArgs mengubah input menjadi argumen SQL (nil untuk string kosong).
func guruArgs(in guruInput) (userID, nip any, mapelIDs json.RawMessage) {
	if in.UserID != nil && strings.TrimSpace(*in.UserID) != "" {
		userID = *in.UserID
	}
	if in.NIP != nil && strings.TrimSpace(*in.NIP) != "" {
		nip = strings.TrimSpace(*in.NIP)
	}
	ids := in.MapelIDs
	if ids == nil {
		ids = []string{}
	}
	raw, _ := json.Marshal(ids)
	return userID, nip, raw
}

// ListGuru mengembalikan halaman guru sesuai pencarian.
func (s *Service) ListGuru(ctx context.Context, q *string, limit, offset int) ([]guruDTO, int64, *platform.AppError) {
	total, err := s.repo.CountGuru(ctx, q)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca guru.")
	}
	items, err := s.repo.ListGuru(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca guru.")
	}
	return items, total, nil
}

// GetGuru mengembalikan satu guru.
func (s *Service) GetGuru(ctx context.Context, id string) (*guruDTO, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Guru tidak ditemukan.")
	}
	d, err := s.repo.GetGuru(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, platform.ErrNotFound("not_found", "Guru tidak ditemukan.")
	}
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca guru.")
	}
	return d, nil
}

// CreateGuru menambah guru baru.
func (s *Service) CreateGuru(ctx context.Context, actor RequestActor, in guruInput) (*guruDTO, *platform.AppError) {
	if msg := in.validate(); msg != "" {
		return nil, platform.NewAppError("validation_error", msg, http.StatusUnprocessableEntity)
	}
	userID, nip, mapelIDs := guruArgs(in)
	d, err := s.repo.InsertGuru(ctx, userID, nip, strings.TrimSpace(in.Nama), mapelIDs)
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			return nil, platform.ErrConflict("conflict", conflictMessage(constraint))
		}
		if isFKViolation(err) {
			return nil, platform.NewAppError("validation_error", "Akun pengguna (user_id) tidak ditemukan.", http.StatusUnprocessableEntity)
		}
		return nil, platform.ErrInternal("internal_error", "Gagal menambah guru.")
	}
	s.catat(ctx, actor, "create_guru", "guru", d.ID, map[string]any{"nama": d.Nama})
	return d, nil
}

// UpdateGuru mengubah guru.
func (s *Service) UpdateGuru(ctx context.Context, actor RequestActor, id string, in guruInput) (*guruDTO, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Guru tidak ditemukan.")
	}
	if msg := in.validate(); msg != "" {
		return nil, platform.NewAppError("validation_error", msg, http.StatusUnprocessableEntity)
	}
	userID, nip, mapelIDs := guruArgs(in)
	d, err := s.repo.UpdateGuru(ctx, id, userID, nip, strings.TrimSpace(in.Nama), mapelIDs)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, platform.ErrNotFound("not_found", "Guru tidak ditemukan.")
	}
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			return nil, platform.ErrConflict("conflict", conflictMessage(constraint))
		}
		if isFKViolation(err) {
			return nil, platform.NewAppError("validation_error", "Akun pengguna (user_id) tidak ditemukan.", http.StatusUnprocessableEntity)
		}
		return nil, platform.ErrInternal("internal_error", "Gagal mengubah guru.")
	}
	s.catat(ctx, actor, "update_guru", "guru", d.ID, map[string]any{"nama": d.Nama})
	return d, nil
}

// DeleteGuru menghapus guru. Guru yang jadi pemilik bank soal tidak boleh
// dihapus (ANALISIS 16.2: staf yang pernah membuat bank soal hanya
// dinonaktifkan).
func (s *Service) DeleteGuru(ctx context.Context, actor RequestActor, id string) (map[string]any, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Guru tidak ditemukan.")
	}
	// bank_soal.guru_id adalah rujukan ke users(id) (Tim 3), bukan ke
	// guru(id), jadi guard bank soal di sini memakai user_id guru.
	if userID := s.repo.GetGuruUserID(ctx, id); userID != nil && *userID != "" {
		if n := s.repo.countOrZero(ctx, `SELECT COUNT(*) FROM bank_soal WHERE guru_id = $1`, *userID); n > 0 {
			return nil, platform.ErrConflict("conflict",
				fmt.Sprintf("Guru masih pemilik %d bank soal. Nonaktifkan akunnya, jangan hapus.", n))
		} else if n < 0 {
			return nil, platform.ErrInternal("internal_error", "Gagal memeriksa pemakaian guru.")
		}
	}
	affected, err := s.repo.DeleteGuru(ctx, id)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menghapus guru.")
	}
	if affected == 0 {
		return nil, platform.ErrNotFound("not_found", "Guru tidak ditemukan.")
	}
	s.catat(ctx, actor, "delete_guru", "guru", id, nil)
	return map[string]any{"deleted": true, "id": id}, nil
}
