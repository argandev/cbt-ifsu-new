package sod

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgxpool"

	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/notifikasi"
	"cbt-smk-v2/backend/internal/platform"
)

// PanjangAlasanMinimum adalah panjang alasan tertulis minimum (karakter).
const PanjangAlasanMinimum = 20

// ValidasiAlasan melaporkan pesan error bila alasan < 20 karakter;
// string kosong berarti valid. Dipakai handler + tes.
func ValidasiAlasan(alasan string) string {
	if utf8.RuneCountInString(strings.TrimSpace(alasan)) < PanjangAlasanMinimum {
		return fmt.Sprintf("alasan wajib minimal %d karakter", PanjangAlasanMinimum)
	}
	return ""
}

// DetailAudit membangun detail JSON audit_log untuk satu override:
// berisi aksi + alasan + entitas + flag SOD_OVERRIDE.
func DetailAudit(aksi, entitas, entitasID, alasan string) map[string]any {
	return map[string]any{
		"flag":         "SOD_OVERRIDE",
		"sod_override": true,
		"aksi":         aksi,
		"entitas":      entitas,
		"entitas_id":   entitasID,
		"alasan":       alasan,
	}
}

// IsiNotifikasi membangun isi notifikasi untuk admin lain.
func IsiNotifikasi(aksi, entitas, entitasID, alasan, namaActor string) string {
	return fmt.Sprintf("Pengecualian SoD oleh %s: %s pada %s (%s). Alasan: %s",
		namaActor, aksi, entitas, entitasID, alasan)
}

// OverrideInput adalah input pencatatan pengecualian SoD.
type OverrideInput struct {
	Aksi      string
	Entitas   string
	EntitasID string
	Alasan    string
}

// RequestActor adalah identitas pemanggil untuk audit log.
// Dibangun handler dari echo.Context agar service bebas dari HTTP
// (pola paket users).
type RequestActor struct {
	UserID string
	RealIP string
}

// Service adalah lapisan logika bisnis paket sod.
type Service struct {
	repo *Repository
	pool *pgxpool.Pool
	log  *slog.Logger
}

// NewService membuat Service.
func NewService(repo *Repository, pool *pgxpool.Pool, log *slog.Logger) *Service {
	return &Service{repo: repo, pool: pool, log: log}
}

// catat menulis audit log. (Dulu: audit.Catat langsung di handler.)
func (s *Service) catat(ctx context.Context, actor RequestActor, aksi, entitas, entitasID string, detail any) error {
	return audit.Catat(ctx, s.pool, actor.UserID, aksi, entitas, entitasID, detail, actor.RealIP)
}

// Override mencatat satu pengecualian SoD darurat: insert override ->
// audit berflag SOD_OVERRIDE -> notifikasi semua admin lain.
// (Dipindah dari closure Register tanpa perubahan perilaku.)
func (s *Service) Override(ctx context.Context, actor RequestActor, in OverrideInput) (map[string]any, *platform.AppError) {
	if actor.UserID == "" {
		return nil, platform.ErrUnauthorized("unauthenticated", "Harus login dulu.")
	}
	aksi := strings.TrimSpace(in.Aksi)
	entitas := strings.TrimSpace(in.Entitas)
	entitasID := strings.TrimSpace(in.EntitasID)
	alasan := strings.TrimSpace(in.Alasan)
	if aksi == "" || entitas == "" {
		return nil, platform.ErrBadRequest("validation_error", "aksi dan entitas wajib diisi.")
	}
	if msg := ValidasiAlasan(in.Alasan); msg != "" {
		return nil, platform.ErrBadRequest("validation_error", msg)
	}
	id, err := s.repo.InsertOverride(ctx, actor.UserID, aksi, entitas, entitasID, alasan)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal mencatat pengecualian SoD.")
	}
	// Audit dengan flag SOD_OVERRIDE (detail JSON berisi aksi +
	// alasan + entitas). Gagal catat tidak menggagalkan override
	// yang sudah tersimpan, tapi dicatat ke log server.
	if err := s.catat(ctx, actor, "sod.override", "sod_override", id,
		DetailAudit(aksi, entitas, entitasID, alasan)); err != nil {
		s.log.Error("sod: gagal mencatat audit SOD_OVERRIDE", "error", err)
	}
	// Notifikasi untuk SEMUA user role admin kecuali actor.
	namaActor := s.repo.NamaActor(ctx, actor.UserID)
	if namaActor == "" {
		namaActor = actor.UserID
	}
	for _, adminID := range s.repo.AdminLainAktif(ctx, actor.UserID) {
		_ = notifikasi.Buat(ctx, s.pool, adminID, "Pengecualian SoD",
			IsiNotifikasi(aksi, entitas, entitasID, alasan, namaActor))
	}
	return map[string]any{
		"id": id, "actor_id": actor.UserID, "aksi": aksi, "entitas": entitas,
		"entitas_id": entitasID, "alasan": alasan,
	}, nil
}
