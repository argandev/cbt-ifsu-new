package exam

import (
	"net/http"

	"cbt-smk-v2/backend/internal/platform"

	"github.com/labstack/echo/v5"
)

// ---------------------------------------------------------------------------
// Usulan unblock / unblock
// ---------------------------------------------------------------------------

// ListUsulanUnblock menangani GET /api/v1/jadwal/:id/usulan-unblock.
func (h *Handler) ListUsulanUnblock(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "proktor", "pengawas"); !ok {
		return nil
	}
	data, appErr := h.svc.ListUsulanUnblock(c.Request().Context(), actorFrom(c), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// CreateUsulanUnblock menangani POST /api/v1/jadwal/:id/usulan-unblock. -> 201
func (h *Handler) CreateUsulanUnblock(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "pengawas", "proktor", "admin"); !ok {
		return nil
	}
	var body struct {
		AttemptID string `json:"attempt_id"`
		Alasan    string `json:"alasan"`
	}
	if err := c.Bind(&body); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	data, appErr := h.svc.CreateUsulanUnblock(c.Request().Context(), actorFrom(c), c.Param("id"), body.AttemptID, body.Alasan)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteCreated(c, data)
}

// UnblockAttempt menangani POST /api/v1/jadwal/:id/unblock.
func (h *Handler) UnblockAttempt(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "proktor", "admin"); !ok {
		return nil
	}
	var body struct {
		AttemptID string `json:"attempt_id"`
		Alasan    string `json:"alasan"`
	}
	if err := c.Bind(&body); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	data, appErr := h.svc.UnblockAttempt(c.Request().Context(), actorFrom(c), c.Param("id"), body.AttemptID, body.Alasan)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}
