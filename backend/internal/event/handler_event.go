package event

import (
	"net/http"

	"cbt-smk-v2/backend/internal/platform"

	"github.com/labstack/echo/v5"
)

// ListEvent menangani GET /api/v1/event.
func (h *Handler) ListEvent(c *echo.Context) error {
	if _, _, ok := requirePerm(c, "event.read"); !ok {
		return nil
	}
	limit, offset := platform.ParseLimit(c, 20)
	items, total, appErr := h.svc.ListEvent(c.Request().Context(), strPtr(c.QueryParam("jenis")), strPtr(c.QueryParam("q")), limit, offset)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return writeItemsList(c, items, total, limit, offset)
}

// CreateEvent menangani POST /api/v1/event.
func (h *Handler) CreateEvent(c *echo.Context) error {
	userID, _, ok := requirePerm(c, "event.create")
	if !ok {
		return nil
	}
	var in eventInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body JSON tidak valid.")
	}
	ev, appErr := h.svc.CreateEvent(c.Request().Context(), RequestActor{UserID: userID, RealIP: c.RealIP()}, in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteCreated(c, ev)
}

// GetEvent menangani GET /api/v1/event/:id.
func (h *Handler) GetEvent(c *echo.Context) error {
	if _, _, ok := requirePerm(c, "event.read"); !ok {
		return nil
	}
	ev, appErr := h.svc.GetEvent(c.Request().Context(), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, ev)
}

// UpdateEvent menangani PUT /api/v1/event/:id.
func (h *Handler) UpdateEvent(c *echo.Context) error {
	userID, _, ok := requirePerm(c, "event.update")
	if !ok {
		return nil
	}
	// Urutan sesuai kode asal: load (404) dulu, baru bind (400).
	ev, appErr := h.svc.LoadEventChecked(c.Request().Context(), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	var in eventInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body JSON tidak valid.")
	}
	updated, appErr := h.svc.ApplyEventUpdate(c.Request().Context(), RequestActor{UserID: userID, RealIP: c.RealIP()}, ev, in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, updated)
}

// DeleteEvent menangani DELETE /api/v1/event/:id.
func (h *Handler) DeleteEvent(c *echo.Context) error {
	userID, _, ok := requirePerm(c, "event.delete")
	if !ok {
		return nil
	}
	out, appErr := h.svc.DeleteEvent(c.Request().Context(), RequestActor{UserID: userID, RealIP: c.RealIP()}, c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, out)
}

// ListRuangan menangani GET /api/v1/ruangan.
func (h *Handler) ListRuangan(c *echo.Context) error {
	if _, _, ok := requirePerm(c, "event.read"); !ok {
		return nil
	}
	limit, offset := platform.ParseLimit(c, 20)
	items, total, appErr := h.svc.ListRuangan(c.Request().Context(), strPtr(c.QueryParam("q")), limit, offset)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return writeItemsList(c, items, total, limit, offset)
}

// CreateRuangan menangani POST /api/v1/ruangan.
func (h *Handler) CreateRuangan(c *echo.Context) error {
	userID, _, ok := requirePerm(c, "event.create")
	if !ok {
		return nil
	}
	var in ruanganInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body JSON tidak valid.")
	}
	r, appErr := h.svc.CreateRuangan(c.Request().Context(), RequestActor{UserID: userID, RealIP: c.RealIP()}, in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteCreated(c, r)
}

// GetRuangan menangani GET /api/v1/ruangan/:id.
func (h *Handler) GetRuangan(c *echo.Context) error {
	if _, _, ok := requirePerm(c, "event.read"); !ok {
		return nil
	}
	r, appErr := h.svc.GetRuangan(c.Request().Context(), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, r)
}

// UpdateRuangan menangani PUT /api/v1/ruangan/:id.
func (h *Handler) UpdateRuangan(c *echo.Context) error {
	userID, _, ok := requirePerm(c, "event.update")
	if !ok {
		return nil
	}
	// Urutan sesuai kode asal: load (404) dulu, baru bind (400).
	r, appErr := h.svc.LoadRuanganChecked(c.Request().Context(), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	var in ruanganInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body JSON tidak valid.")
	}
	updated, appErr := h.svc.ApplyRuanganUpdate(c.Request().Context(), RequestActor{UserID: userID, RealIP: c.RealIP()}, r, in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, updated)
}

// DeleteRuangan menangani DELETE /api/v1/ruangan/:id.
func (h *Handler) DeleteRuangan(c *echo.Context) error {
	userID, _, ok := requirePerm(c, "event.delete")
	if !ok {
		return nil
	}
	out, appErr := h.svc.DeleteRuangan(c.Request().Context(), RequestActor{UserID: userID, RealIP: c.RealIP()}, c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, out)
}

// AssignEventRuangan menangani PUT /api/v1/jadwal/:id/event-ruangan.
// Field tidak dikirim / string kosong = lepas penugasan (NULL).
func (h *Handler) AssignEventRuangan(c *echo.Context) error {
	userID, _, ok := requirePerm(c, "event.update")
	if !ok {
		return nil
	}
	// Urutan sesuai kode asal: load jadwal (404) dulu, baru bind (400).
	kelasID, kelasNama, appErr := h.svc.LoadJadwalKelasChecked(c.Request().Context(), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	var body struct {
		EventID   *string `json:"event_id"`
		RuanganID *string `json:"ruangan_id"`
	}
	if err := c.Bind(&body); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body JSON tidak valid.")
	}
	out, appErr := h.svc.ApplyAssign(c.Request().Context(), RequestActor{UserID: userID, RealIP: c.RealIP()}, c.Param("id"), kelasID, kelasNama, body.EventID, body.RuanganID)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, out)
}
