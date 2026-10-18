package master

import (
	"net/http"

	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/labstack/echo/v5"
)

// ---------------------------------------------------------------------
// Guru
// ---------------------------------------------------------------------

// ListGuru menangani GET /api/v1/master/guru.
func (h *Handler) ListGuru(c *echo.Context) error {
	if !audit.HasPermission(c, "master.guru.read") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat guru.")
	}
	page, perPage, offset := platform.ParsePage(c)
	items, total, appErr := h.svc.ListGuru(c.Request().Context(), searchParam(c), perPage, offset)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WritePageList(c, items, page, perPage, total)
}

// GetGuru menangani GET /api/v1/master/guru/:id.
func (h *Handler) GetGuru(c *echo.Context) error {
	if !audit.HasPermission(c, "master.guru.read") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat guru.")
	}
	d, appErr := h.svc.GetGuru(c.Request().Context(), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, d)
}

// CreateGuru menangani POST /api/v1/master/guru.
func (h *Handler) CreateGuru(c *echo.Context) error {
	if !audit.HasPermission(c, "master.guru.create") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menambah guru.")
	}
	var in guruInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	d, appErr := h.svc.CreateGuru(c.Request().Context(), actorFrom(c), in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteCreated(c, d)
}

// UpdateGuru menangani PUT /api/v1/master/guru/:id.
func (h *Handler) UpdateGuru(c *echo.Context) error {
	if !audit.HasPermission(c, "master.guru.update") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mengubah guru.")
	}
	var in guruInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	d, appErr := h.svc.UpdateGuru(c.Request().Context(), actorFrom(c), c.Param("id"), in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, d)
}

// DeleteGuru menangani DELETE /api/v1/master/guru/:id.
func (h *Handler) DeleteGuru(c *echo.Context) error {
	if !audit.HasPermission(c, "master.guru.delete") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menghapus guru.")
	}
	out, appErr := h.svc.DeleteGuru(c.Request().Context(), actorFrom(c), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, out)
}
