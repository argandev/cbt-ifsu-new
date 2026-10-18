// Package sod — Pengecualian darurat Pemisahan Tugas / Separation of
// Duties (Tim Fase 3).
//
// Aturan SoD (ANALISIS Bab 11.4, 7 aturan): SoD-1 pembuat/edit bank ≠
// pengunci; SoD-2 pembuat soal ≠ pemegang token; SoD-3 pengoreksi esai
// ≠ pengubah nilai final; SoD-4 penguji UKK ≠ finalisator nilai UKK;
// SoD-5 pengawas mengajukan, proktor mengeksekusi; SoD-6 pengelola user
// ≠ pengelola audit log; SoD-7 pengaju draft jadwal ≠ finalisator.
//
// Pengecualian darurat HANYA lewat endpoint di bawah, dengan:
// alasan tertulis (min 20 karakter) + audit dengan flag SOD_OVERRIDE +
// notifikasi semua user role admin kecuali actor. Pengecualian TIDAK
// PERNAH dipakai untuk ubah nilai tanpa jejak / hapus audit log —
// audit_log append-only, tidak ada endpoint ubah/hapus.
//
// Endpoint (Register):
//
//	POST /api/v1/sod/override  (permission sod.override — admin saja)
//
// Struktur layer (hasil refactor arsitektur):
//   - handler.go:    Echo handler + Register (file ini)
//   - service.go:    validasi + logika bisnis + notifikasi admin
//   - repository.go: DTO + semua query SQL
//   - platform:      envelope JSON (menggantikan helper lokal)
//
// Perilaku (route, status, bentuk JSON, pesan error) IDENTIK dengan
// sebelum refactor.
package sod

import (
	"net/http"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/config"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// Handler adalah lapisan HTTP paket sod.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// actorFrom membangun platform.RequestActor dari echo.Context. Pemanggil sudah
// dipastikan login (cek 401 di handler), jadi userID tidak kosong.
func actorFrom(c *echo.Context) platform.RequestActor {
	userID, _, _, _, _ := audit.CurrentUser(c)
	return platform.RequestActor{UserID: userID, RealIP: c.RealIP()}
}

// Register mendaftarkan endpoint SoD override.
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q
	_ = cfg

	repo := NewRepository(pool)
	svc := NewService(repo, pool, e.Logger)
	h := NewHandler(svc)

	e.POST("/api/v1/sod/override", h.Override)
}

// Override menangani POST /api/v1/sod/override.
func (h *Handler) Override(c *echo.Context) error {
	_, _, _, _, ok := audit.CurrentUser(c)
	if !ok {
		return platform.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
	}
	if !audit.HasPermission(c, "sod.override") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mencatat pengecualian SoD.")
	}
	var in struct {
		Aksi      string `json:"aksi"`
		Entitas   string `json:"entitas"`
		EntitasID string `json:"entitas_id"`
		Alasan    string `json:"alasan"`
	}
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body JSON tidak valid.")
	}
	data, appErr := h.svc.Override(c.Request().Context(), actorFrom(c), OverrideInput{
		Aksi:      in.Aksi,
		Entitas:   in.Entitas,
		EntitasID: in.EntitasID,
		Alasan:    in.Alasan,
	})
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteCreated(c, data)
}
