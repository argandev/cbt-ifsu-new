package master

import (
	"net/http"

	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/labstack/echo/v5"
)

// ---------------------------------------------------------------------
// Siswa
// ---------------------------------------------------------------------

// ListSiswa menangani GET /api/v1/master/siswa.
func (h *Handler) ListSiswa(c *echo.Context) error {
	if !audit.HasPermission(c, "master.siswa.read") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat siswa.")
	}
	page, perPage, offset := platform.ParsePage(c)
	items, total, appErr := h.svc.ListSiswa(c.Request().Context(),
		optUUID(c, "kelas_id"), optUUID(c, "jurusan_id"), optBool(c, "is_active"), searchParam(c),
		perPage, offset)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WritePageList(c, items, page, perPage, total)
}

// GetSiswa menangani GET /api/v1/master/siswa/:id.
func (h *Handler) GetSiswa(c *echo.Context) error {
	if !audit.HasPermission(c, "master.siswa.read") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat siswa.")
	}
	d, appErr := h.svc.GetSiswa(c.Request().Context(), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, d)
}

// CreateSiswa menangani POST /api/v1/master/siswa.
func (h *Handler) CreateSiswa(c *echo.Context) error {
	if !audit.HasPermission(c, "master.siswa.create") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menambah siswa.")
	}
	var in siswaInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	d, appErr := h.svc.CreateSiswa(c.Request().Context(), actorFrom(c), in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteCreated(c, d)
}

// UpdateSiswa menangani PUT /api/v1/master/siswa/:id.
func (h *Handler) UpdateSiswa(c *echo.Context) error {
	if !audit.HasPermission(c, "master.siswa.update") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mengubah siswa.")
	}
	var in siswaInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	d, appErr := h.svc.UpdateSiswa(c.Request().Context(), actorFrom(c), c.Param("id"), in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, d)
}

// DeleteSiswa menangani DELETE /api/v1/master/siswa/:id.
func (h *Handler) DeleteSiswa(c *echo.Context) error {
	if !audit.HasPermission(c, "master.siswa.delete") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menghapus siswa.")
	}
	out, appErr := h.svc.DeleteSiswa(c.Request().Context(), actorFrom(c), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, out)
}
