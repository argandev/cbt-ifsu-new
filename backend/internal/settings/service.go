package settings

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Service ----------------------------------------------------------------

// DefaultSettings adalah nilai default tiap key (dipakai bila baris belum ada).
func DefaultSettings() map[string]any {
	return map[string]any{
		"tampil_hasil_ke_siswa": false,
		"token_expiry_detik":    900,
		"toleransi_telat_menit": 15,
		"max_pelanggaran":       2,
		"durasi_default":        90,
		"nama_sekolah":          "",
		"logo_url":              "",
		"kktp_per_mapel":        map[string]any{},
		"remedial_mode":         "tertinggi",
	}
}

// validasi per key; mengembalikan nilai yang sudah dinormalisasi + ok.
// (Dipindah dari settings.go tanpa perubahan.)
func validateSetting(key string, raw json.RawMessage) (any, bool) {
	switch key {
	case "tampil_hasil_ke_siswa":
		var v bool
		if json.Unmarshal(raw, &v) != nil {
			return nil, false
		}
		return v, true
	case "token_expiry_detik", "toleransi_telat_menit", "max_pelanggaran", "durasi_default":
		var v float64
		if json.Unmarshal(raw, &v) != nil || v < 0 || v != float64(int(v)) {
			return nil, false
		}
		// Batas wajar: token expiry 60-86400 detik, toleransi 0-120 menit,
		// max pelanggaran 1-10, durasi 5-600 menit.
		switch key {
		case "token_expiry_detik":
			if v < 60 || v > 86400 {
				return nil, false
			}
		case "toleransi_telat_menit":
			if v > 120 {
				return nil, false
			}
		case "max_pelanggaran":
			if v < 1 || v > 10 {
				return nil, false
			}
		case "durasi_default":
			if v < 5 || v > 600 {
				return nil, false
			}
		}
		return int(v), true
	case "nama_sekolah", "logo_url":
		var v string
		if json.Unmarshal(raw, &v) != nil {
			return nil, false
		}
		return v, true
	case "kktp_per_mapel":
		// KKTP per mapel/TP: objek { "<kode_mapel>": {deskripsi, rubrik, interval, ...} }.
		// BUKAN hardcode KKM=75 — tiap mapel/TP punya ambang & deskripsinya sendiri.
		var v map[string]any
		if json.Unmarshal(raw, &v) != nil {
			return nil, false
		}
		return v, true
	case "remedial_mode":
		var v string
		if json.Unmarshal(raw, &v) != nil || v != "tertinggi" {
			return nil, false
		}
		return v, true
	}
	return nil, false
}

// RequestActor adalah identitas pemanggil untuk audit log.
// Dibangun handler dari echo.Context agar service bebas dari HTTP.
type RequestActor struct {
	UserID string
	RealIP string
}

// Service adalah lapisan logika bisnis paket settings.
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

// GetAll mengembalikan gabungan default + nilai tersimpan.
func (s *Service) GetAll(ctx context.Context) (map[string]any, *platform.AppError) {
	stored, err := s.repo.BacaSemua(ctx)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca settings.")
	}
	result := DefaultSettings()
	for key, raw := range stored {
		var v any
		if err := json.Unmarshal(raw, &v); err == nil {
			result[key] = v
		}
	}
	return result, nil
}

// Update memvalidasi + menyimpan batch settings. updates tidak boleh kosong
// (dicek handler). Pesan error dipertahankan.
func (s *Service) Update(ctx context.Context, actor RequestActor, updates map[string]json.RawMessage) (map[string]any, *platform.AppError) {
	known := DefaultSettings()
	normalized := map[string]any{}
	for key, raw := range updates {
		key = strings.TrimSpace(key)
		if _, ada := known[key]; !ada {
			return nil, platform.ErrBadRequest("validation_error", "Key tidak dikenal: "+key)
		}
		v, ok := validateSetting(key, raw)
		if !ok {
			return nil, platform.ErrBadRequest("validation_error", "Nilai tidak valid untuk key: "+key)
		}
		normalized[key] = v
	}
	if err := s.repo.Simpan(ctx, normalized); err != nil {
		if errors.Is(err, ErrMulaiTx) {
			return nil, platform.ErrInternal("internal_error", "Gagal memulai penyimpanan settings.")
		}
		return nil, platform.ErrInternal("internal_error", "Gagal menyimpan settings.")
	}
	s.catat(ctx, actor, "settings.update", "settings", "",
		map[string]any{"updated": normalized})
	return map[string]any{"updated": normalized}, nil
}
