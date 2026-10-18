// Package dashboard — angka ringkasan untuk kartu Dashboard /adm.
//
// Sebelumnya kartu Dashboard (Bank Soal, Jadwal Ujian, Peserta, Rekap)
// di-hardcode "-" di frontend karena belum ada endpoint agregat — terlihat
// seperti data rusak. Endpoint ini mengembalikan hitungan nyata dari
// database. Hanya staf (role apa pun kecuali "siswa") yang boleh membaca:
// angka agregat seluruh sekolah bukan data milik siswa sendiri.
//
// Struktur layer (hasil refactor arsitektur):
//   - dashboard.go:  dokumentasi paket
//   - handler.go:    Echo handler + Register (signature Register
//     DIPERTAHANKAN: tanpa q, cfg)
//   - service.go:    agregat ringkasan
//   - repository.go: COUNT(*) per tabel
//   - platform:      envelope JSON (menggantikan helper lokal)
//
// Perilaku (route, status, bentuk JSON, pesan error) IDENTIK dengan
// sebelum refactor.
package dashboard

import (
	"net/http"

	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// --- Handler ----------------------------------------------------------------

// Handler adalah lapisan HTTP paket dashboard.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register mendaftarkan endpoint ringkasan dashboard.
// Signature DIPERTAHANKAN (tanpa q, cfg) — dipakai cmd/server/main.go.
func Register(e *echo.Echo, pool *pgxpool.Pool) {
	repo := NewRepository(pool)
	svc := NewService(repo, pool)
	h := NewHandler(svc)

	// GET /api/v1/dashboard/ringkasan
	e.GET("/api/v1/dashboard/ringkasan", h.Ringkasan)
}

// Ringkasan menangani GET /api/v1/dashboard/ringkasan.
func (h *Handler) Ringkasan(c *echo.Context) error {
	_, role, _, _, ok := audit.CurrentUser(c)
	if !ok {
		return platform.WriteError(c, http.StatusUnauthorized, "unauthorized", "Silakan login dulu.")
	}
	if role == "siswa" {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Ringkasan dashboard hanya untuk staf.")
	}
	data, appErr := h.svc.Ringkasan(c.Request().Context())
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}
