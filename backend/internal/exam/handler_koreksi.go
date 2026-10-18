package exam

import (
	"net/http"

	"cbt-smk-v2/backend/internal/platform"

	"github.com/labstack/echo/v5"
)

// ---------------------------------------------------------------------------
// Koreksi
// ---------------------------------------------------------------------------

// ListKoreksi menangani GET /api/v1/koreksi/:jadwalId.
func (h *Handler) ListKoreksi(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "guru"); !ok {
		return nil
	}
	jadwalID := c.Param("jadwalId")
	if !isUUID(jadwalID) {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "ID jadwal tidak valid.")
	}
	data, appErr := h.svc.ListKoreksi(c.Request().Context(), actorFrom(c), jadwalID)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// KoreksiJawaban menangani POST /api/v1/koreksi/:jawabanId.
func (h *Handler) KoreksiJawaban(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "guru"); !ok {
		return nil
	}
	var body struct {
		Skor float64 `json:"skor"`
	}
	if err := c.Bind(&body); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	if body.Skor < 0 || body.Skor > 100 {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "skor harus 0–100.")
	}
	data, appErr := h.svc.KoreksiJawaban(c.Request().Context(), actorFrom(c), c.Param("jawabanId"), body.Skor)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}
