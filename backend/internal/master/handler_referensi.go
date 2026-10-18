package master

import (
	"net/http"

	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/labstack/echo/v5"
)

// ---------------------------------------------------------------------
// Kompetensi (guard: master.mapel.*)
// ---------------------------------------------------------------------

// ListKompetensi menangani GET /api/v1/master/kompetensi.
func (h *Handler) ListKompetensi(c *echo.Context) error {
	if !audit.HasPermission(c, "master.mapel.read") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat kompetensi.")
	}
	page, perPage, offset := platform.ParsePage(c)
	items, total, appErr := h.svc.ListKompetensi(c.Request().Context(), optUUID(c, "mapel_id"), searchParam(c), perPage, offset)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WritePageList(c, items, page, perPage, total)
}

// GetKompetensi menangani GET /api/v1/master/kompetensi/:id.
func (h *Handler) GetKompetensi(c *echo.Context) error {
	if !audit.HasPermission(c, "master.mapel.read") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat kompetensi.")
	}
	d, appErr := h.svc.GetKompetensi(c.Request().Context(), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, d)
}

// CreateKompetensi menangani POST /api/v1/master/kompetensi.
func (h *Handler) CreateKompetensi(c *echo.Context) error {
	if !audit.HasPermission(c, "master.mapel.create") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menambah kompetensi.")
	}
	var in kompetensiInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	d, appErr := h.svc.CreateKompetensi(c.Request().Context(), actorFrom(c), in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteCreated(c, d)
}

// UpdateKompetensi menangani PUT /api/v1/master/kompetensi/:id.
func (h *Handler) UpdateKompetensi(c *echo.Context) error {
	if !audit.HasPermission(c, "master.mapel.update") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mengubah kompetensi.")
	}
	var in kompetensiInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	d, appErr := h.svc.UpdateKompetensi(c.Request().Context(), actorFrom(c), c.Param("id"), in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, d)
}

// DeleteKompetensi menangani DELETE /api/v1/master/kompetensi/:id.
func (h *Handler) DeleteKompetensi(c *echo.Context) error {
	if !audit.HasPermission(c, "master.mapel.delete") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menghapus kompetensi.")
	}
	out, appErr := h.svc.DeleteKompetensi(c.Request().Context(), actorFrom(c), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, out)
}

// ---------------------------------------------------------------------
// Kisi-kisi (guard: master.mapel.*)
// ---------------------------------------------------------------------

// ListKisiKisi menangani GET /api/v1/master/kisi-kisi.
func (h *Handler) ListKisiKisi(c *echo.Context) error {
	if !audit.HasPermission(c, "master.mapel.read") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat kisi-kisi.")
	}
	page, perPage, offset := platform.ParsePage(c)
	items, total, appErr := h.svc.ListKisiKisi(c.Request().Context(), optUUID(c, "mapel_id"), optInt(c, "kelas_tingkat"), searchParam(c), perPage, offset)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WritePageList(c, items, page, perPage, total)
}

// GetKisiKisi menangani GET /api/v1/master/kisi-kisi/:id.
func (h *Handler) GetKisiKisi(c *echo.Context) error {
	if !audit.HasPermission(c, "master.mapel.read") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat kisi-kisi.")
	}
	d, appErr := h.svc.GetKisiKisi(c.Request().Context(), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, d)
}

// CreateKisiKisi menangani POST /api/v1/master/kisi-kisi.
func (h *Handler) CreateKisiKisi(c *echo.Context) error {
	if !audit.HasPermission(c, "master.mapel.create") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menambah kisi-kisi.")
	}
	var in kisiKisiInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	d, appErr := h.svc.CreateKisiKisi(c.Request().Context(), actorFrom(c), in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteCreated(c, d)
}

// UpdateKisiKisi menangani PUT /api/v1/master/kisi-kisi/:id.
func (h *Handler) UpdateKisiKisi(c *echo.Context) error {
	if !audit.HasPermission(c, "master.mapel.update") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mengubah kisi-kisi.")
	}
	var in kisiKisiInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	d, appErr := h.svc.UpdateKisiKisi(c.Request().Context(), actorFrom(c), c.Param("id"), in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, d)
}

// DeleteKisiKisi menangani DELETE /api/v1/master/kisi-kisi/:id.
func (h *Handler) DeleteKisiKisi(c *echo.Context) error {
	if !audit.HasPermission(c, "master.mapel.delete") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menghapus kisi-kisi.")
	}
	out, appErr := h.svc.DeleteKisiKisi(c.Request().Context(), actorFrom(c), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, out)
}
