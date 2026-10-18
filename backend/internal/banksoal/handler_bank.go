package banksoal

import (
	"net/http"

	"cbt-smk-v2/backend/internal/platform"

	"github.com/labstack/echo/v5"
)

// ListBank menangani GET /api/v1/bank-soal.
func (h *Handler) ListBank(c *echo.Context) error {
	if _, _, ok := requirePerm(c, "banksoal.read"); !ok {
		return nil
	}
	limit, offset := platform.ParseLimit(c, 20)
	items, total, appErr := h.svc.ListBank(c.Request().Context(), bankFilterArgs(c), limit, offset)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return writeItemsList(c, items, total, limit, offset)
}

// CreateBank menangani POST /api/v1/bank-soal.
func (h *Handler) CreateBank(c *echo.Context) error {
	userID, role, ok := requirePerm(c, "banksoal.create")
	if !ok {
		return nil
	}
	var in bankInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "invalid_body", "Body JSON tidak valid.")
	}
	b, appErr := h.svc.CreateBank(c.Request().Context(), RequestActor{UserID: userID, RealIP: c.RealIP()}, role, in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteCreated(c, b)
}

// GetBank menangani GET /api/v1/bank-soal/:id.
func (h *Handler) GetBank(c *echo.Context) error {
	if _, _, ok := requirePerm(c, "banksoal.read"); !ok {
		return nil
	}
	b, appErr := h.svc.GetBank(c.Request().Context(), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, b)
}

// UpdateBank menangani PUT /api/v1/bank-soal/:id.
func (h *Handler) UpdateBank(c *echo.Context) error {
	userID, role, ok := requirePerm(c, "banksoal.update")
	if !ok {
		return nil
	}
	// Urutan sesuai kode asal: load (404) + cek scope dulu, baru bind (400).
	b, appErr := h.svc.LoadBankChecked(c.Request().Context(), role, userID, c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	var in bankInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "invalid_body", "Body JSON tidak valid.")
	}
	updated, appErr := h.svc.ApplyBankUpdate(c.Request().Context(), RequestActor{UserID: userID, RealIP: c.RealIP()}, role, b, in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, updated)
}

// DeleteBank menangani DELETE /api/v1/bank-soal/:id.
func (h *Handler) DeleteBank(c *echo.Context) error {
	userID, role, ok := requirePerm(c, "banksoal.delete")
	if !ok {
		return nil
	}
	out, appErr := h.svc.DeleteBank(c.Request().Context(), RequestActor{UserID: userID, RealIP: c.RealIP()}, role, c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, out)
}

// LockBank menangani POST /api/v1/bank-soal/:id/lock.
func (h *Handler) LockBank(c *echo.Context) error {
	return h.setBankLock(c, true)
}

// UnlockBank menangani POST /api/v1/bank-soal/:id/unlock.
func (h *Handler) UnlockBank(c *echo.Context) error {
	return h.setBankLock(c, false)
}

// setBankLock menangani lock/unlock bank soal. SoD-1 (ANALISIS Bab 11.4):
// pembuat bank TIDAK BOLEH mengunci/membuka bank buatannya sendiri —
// pengunci harus orang kedua (guru lain) atau admin.
func (h *Handler) setBankLock(c *echo.Context, locked bool) error {
	userID, role, ok := requirePerm(c, "banksoal.lock")
	if !ok {
		return nil
	}
	b, appErr := h.svc.SetBankLock(c.Request().Context(), RequestActor{UserID: userID, RealIP: c.RealIP()}, role, c.Param("id"), locked)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, b)
}
