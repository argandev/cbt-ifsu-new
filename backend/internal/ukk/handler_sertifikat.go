package ukk

import (
	"net/http"

	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/labstack/echo/v5"
)

// GetSertifikat menangani GET /api/v1/ukk/:id/sertifikat/:siswaId.
// Siswa: hanya miliknya sendiri.
func (h *Handler) GetSertifikat(c *echo.Context) error {
	if _, _, ok := requireLogin(c); !ok {
		return nil
	}
	ukkID := c.Param("id")
	siswaID := c.Param("siswaId")
	if !isValidUUID(ukkID) || !isValidUUID(siswaID) {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "ID tidak valid.")
	}
	// Izin staf dihitung di sini (butuh echo.Context); service memakai
	// hasilnya setelah pemeriksaan 404 — urutan dipertahankan.
	bolehStaf := audit.HasPermission(c, "ukk.read")
	s, appErr := h.svc.GetSertifikat(c.Request().Context(), actorFrom(c), ukkID, siswaID, bolehStaf)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, s)
}

// CreateSertifikat menangani POST /api/v1/ukk/:id/sertifikat/:siswaId.
// Sertifikat HANYA bagi yang lulus.
func (h *Handler) CreateSertifikat(c *echo.Context) error {
	if _, _, ok := requirePerm(c, "ukk.update"); !ok {
		return nil
	}
	ukkID := c.Param("id")
	siswaID := c.Param("siswaId")
	if !isValidUUID(ukkID) || !isValidUUID(siswaID) {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "ID tidak valid.")
	}
	s, baru, appErr := h.svc.CreateSertifikat(c.Request().Context(), actorFrom(c), ukkID, siswaID)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	// Idempoten: sudah ada -> 200; baru diterbitkan -> 201.
	if !baru {
		return platform.WriteOK(c, s)
	}
	return platform.WriteStatus(c, http.StatusCreated, s)
}
