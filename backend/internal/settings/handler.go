// Package settings — pengaturan sekolah (key/value JSONB).
//
// Tabel settings dibuat di migrasi 0001 (Tim 1): settings(key PK, value jsonb, updated_at).
// Daftar key yang dikenal + default ada di DefaultSettings. Key tak dikenal
// ditolak pada PUT agar tidak ada sampah konfigurasi.
// remedial_mode HANYA "tertinggi" (nilai remedial vs awal dipakai yang tertinggi).
//
// Struktur layer (hasil refactor arsitektur):
//   - handler.go:    Echo handler + Register
//   - service.go:    DefaultSettings + validasi per key + logika simpan
//   - repository.go: baca + upsert transaksional
//   - platform:      envelope JSON (menggantikan helper lokal)
//
// Perilaku (route, status, bentuk JSON, pesan error) IDENTIK dengan
// sebelum refactor.
package settings

import (
	"encoding/json"
	"net/http"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/config"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// --- Handler ----------------------------------------------------------------

// Handler adalah lapisan HTTP paket settings.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// actorFrom membangun RequestActor dari echo.Context.
func actorFrom(c *echo.Context) RequestActor {
	userID, _, _, _, _ := audit.CurrentUser(c)
	return RequestActor{UserID: userID, RealIP: c.RealIP()}
}

// Register mendaftarkan endpoint settings.
//
//	GET /api/v1/settings
//	PUT /api/v1/settings
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q
	_ = cfg

	repo := NewRepository(pool)
	svc := NewService(repo, pool)
	h := NewHandler(svc)

	e.GET("/api/v1/settings", h.Get)
	e.PUT("/api/v1/settings", h.Put)
}

// Get menangani GET /api/v1/settings.
func (h *Handler) Get(c *echo.Context) error {
	if !audit.HasPermission(c, "settings.read") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat settings.")
	}
	result, appErr := h.svc.GetAll(c.Request().Context())
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, result)
}

// Put menangani PUT /api/v1/settings.
func (h *Handler) Put(c *echo.Context) error {
	if !audit.HasPermission(c, "settings.update") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mengubah settings.")
	}
	// Terima body {"settings": {...}} maupun flat {"key": value}.
	var generic map[string]json.RawMessage
	if err := c.Bind(&generic); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	updates := map[string]json.RawMessage{}
	if rawSettings, hasWrapper := generic["settings"]; hasWrapper {
		if err := json.Unmarshal(rawSettings, &updates); err != nil {
			return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Field settings tidak valid.")
		}
	} else {
		updates = generic
	}
	body := struct {
		Settings map[string]json.RawMessage `json:"settings"`
	}{Settings: updates}
	if len(body.Settings) == 0 {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Tidak ada setting yang dikirim.")
	}

	data, appErr := h.svc.Update(c.Request().Context(), actorFrom(c), body.Settings)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}
