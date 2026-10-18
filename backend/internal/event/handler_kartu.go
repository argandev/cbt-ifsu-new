package event

import (
	"net/http"

	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/labstack/echo/v5"
)

// KartuJadwal menangani GET /api/v1/kartu/:jadwalId (JSON, LEGACY).
// Staf admin/proktor/guru saja (gate requireKartuStaf).
func (h *Handler) KartuJadwal(c *echo.Context) error {
	if _, _, ok := requireKartuStaf(c); !ok {
		return nil
	}
	out, appErr := h.svc.KartuJadwal(c.Request().Context(), c.Param("jadwalId"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, out)
}

// KartuSiswa menangani GET /api/v1/kartu/:jadwalId/:siswaId (JSON, LEGACY).
// Siswa hanya boleh melihat kartunya sendiri (bukan pemilik -> 404,
// anti-IDOR); staf admin/proktor/guru boleh semua (gate requireKartuStaf).
func (h *Handler) KartuSiswa(c *echo.Context) error {
	uid, role, _, _, ok := audit.CurrentUser(c)
	if !ok {
		return platform.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
	}
	if role != "siswa" {
		if _, _, ok := requireKartuStaf(c); !ok {
			return nil
		}
	}
	k, appErr := h.svc.KartuSiswa(c.Request().Context(), uid, role, c.Param("jadwalId"), c.Param("siswaId"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, k)
}
