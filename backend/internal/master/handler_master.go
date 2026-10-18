package master

import (
	"net/http"
	"strings"

	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/labstack/echo/v5"
)

// ---------------------------------------------------------------------
// Jurusan
// ---------------------------------------------------------------------

// ListJurusan menangani GET /api/v1/master/jurusan.
func (h *Handler) ListJurusan(c *echo.Context) error {
	if !audit.HasPermission(c, "master.jurusan.read") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat jurusan.")
	}
	page, perPage, offset := platform.ParsePage(c)
	items, total, appErr := h.svc.ListJurusan(c.Request().Context(), searchParam(c), perPage, offset)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WritePageList(c, items, page, perPage, total)
}

// GetJurusan menangani GET /api/v1/master/jurusan/:id.
func (h *Handler) GetJurusan(c *echo.Context) error {
	if !audit.HasPermission(c, "master.jurusan.read") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat jurusan.")
	}
	d, appErr := h.svc.GetJurusan(c.Request().Context(), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, d)
}

// CreateJurusan menangani POST /api/v1/master/jurusan.
func (h *Handler) CreateJurusan(c *echo.Context) error {
	if !audit.HasPermission(c, "master.jurusan.create") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menambah jurusan.")
	}
	var in jurusanInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	d, appErr := h.svc.CreateJurusan(c.Request().Context(), actorFrom(c), in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteCreated(c, d)
}

// UpdateJurusan menangani PUT /api/v1/master/jurusan/:id.
func (h *Handler) UpdateJurusan(c *echo.Context) error {
	if !audit.HasPermission(c, "master.jurusan.update") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mengubah jurusan.")
	}
	var in jurusanInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	d, appErr := h.svc.UpdateJurusan(c.Request().Context(), actorFrom(c), c.Param("id"), in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, d)
}

// DeleteJurusan menangani DELETE /api/v1/master/jurusan/:id.
func (h *Handler) DeleteJurusan(c *echo.Context) error {
	if !audit.HasPermission(c, "master.jurusan.delete") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menghapus jurusan.")
	}
	out, appErr := h.svc.DeleteJurusan(c.Request().Context(), actorFrom(c), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, out)
}

// ---------------------------------------------------------------------
// Kelas
// ---------------------------------------------------------------------

// ListKelas menangani GET /api/v1/master/kelas.
func (h *Handler) ListKelas(c *echo.Context) error {
	if !audit.HasPermission(c, "master.kelas.read") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat kelas.")
	}
	page, perPage, offset := platform.ParsePage(c)
	items, total, appErr := h.svc.ListKelas(c.Request().Context(), optUUID(c, "jurusan_id"), searchParam(c), perPage, offset)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WritePageList(c, items, page, perPage, total)
}

// GetKelas menangani GET /api/v1/master/kelas/:id.
func (h *Handler) GetKelas(c *echo.Context) error {
	if !audit.HasPermission(c, "master.kelas.read") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat kelas.")
	}
	d, appErr := h.svc.GetKelas(c.Request().Context(), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, d)
}

// CreateKelas menangani POST /api/v1/master/kelas.
func (h *Handler) CreateKelas(c *echo.Context) error {
	if !audit.HasPermission(c, "master.kelas.create") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menambah kelas.")
	}
	var in kelasInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	d, appErr := h.svc.CreateKelas(c.Request().Context(), actorFrom(c), in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteCreated(c, d)
}

// UpdateKelas menangani PUT /api/v1/master/kelas/:id.
func (h *Handler) UpdateKelas(c *echo.Context) error {
	if !audit.HasPermission(c, "master.kelas.update") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mengubah kelas.")
	}
	var in kelasInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	d, appErr := h.svc.UpdateKelas(c.Request().Context(), actorFrom(c), c.Param("id"), in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, d)
}

// DeleteKelas menangani DELETE /api/v1/master/kelas/:id.
func (h *Handler) DeleteKelas(c *echo.Context) error {
	if !audit.HasPermission(c, "master.kelas.delete") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menghapus kelas.")
	}
	out, appErr := h.svc.DeleteKelas(c.Request().Context(), actorFrom(c), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, out)
}

// ---------------------------------------------------------------------
// Mapel
// ---------------------------------------------------------------------

// optJenis membaca ?jenis= (hanya 'umum' / 'produktif' yang diterima).
func optJenis(c *echo.Context) *string {
	v := strings.TrimSpace(c.QueryParam("jenis"))
	if v != "umum" && v != "produktif" {
		return nil
	}
	return &v
}

// ListMapel menangani GET /api/v1/master/mapel.
func (h *Handler) ListMapel(c *echo.Context) error {
	if !audit.HasPermission(c, "master.mapel.read") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat mapel.")
	}
	page, perPage, offset := platform.ParsePage(c)
	items, total, appErr := h.svc.ListMapel(c.Request().Context(), optJenis(c), optUUID(c, "jurusan_id"), searchParam(c), perPage, offset)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WritePageList(c, items, page, perPage, total)
}

// GetMapel menangani GET /api/v1/master/mapel/:id.
func (h *Handler) GetMapel(c *echo.Context) error {
	if !audit.HasPermission(c, "master.mapel.read") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat mapel.")
	}
	d, appErr := h.svc.GetMapel(c.Request().Context(), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, d)
}

// CreateMapel menangani POST /api/v1/master/mapel.
func (h *Handler) CreateMapel(c *echo.Context) error {
	if !audit.HasPermission(c, "master.mapel.create") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menambah mapel.")
	}
	var in mapelInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	d, appErr := h.svc.CreateMapel(c.Request().Context(), actorFrom(c), in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteCreated(c, d)
}

// UpdateMapel menangani PUT /api/v1/master/mapel/:id.
func (h *Handler) UpdateMapel(c *echo.Context) error {
	if !audit.HasPermission(c, "master.mapel.update") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mengubah mapel.")
	}
	var in mapelInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	d, appErr := h.svc.UpdateMapel(c.Request().Context(), actorFrom(c), c.Param("id"), in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, d)
}

// DeleteMapel menangani DELETE /api/v1/master/mapel/:id.
func (h *Handler) DeleteMapel(c *echo.Context) error {
	if !audit.HasPermission(c, "master.mapel.delete") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menghapus mapel.")
	}
	out, appErr := h.svc.DeleteMapel(c.Request().Context(), actorFrom(c), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, out)
}
