package exam

import (
	"net/http"
	"strings"

	"cbt-smk-v2/backend/internal/platform"

	"github.com/labstack/echo/v5"
)

// ---------------------------------------------------------------------------
// Jadwal ujian
// ---------------------------------------------------------------------------

// ListJadwal menangani GET /api/v1/jadwal.
func (h *Handler) ListJadwal(c *echo.Context) error {
	// penguji_ukk boleh BACA jadwal (Rekap butuh daftar jadwal; tanpa ini
	// menu Rekap yang tampil untuk penguji_ukk selalu 403). Tulis tetap
	// admin/proktor/guru/pengawas (create/update/delete tidak berubah).
	if _, _, _, ok := requireRole(c, "admin", "proktor", "guru", "pengawas", "penguji_ukk"); !ok {
		return nil
	}
	status := strings.TrimSpace(c.QueryParam("status"))
	kelasID := strings.TrimSpace(c.QueryParam("kelas_id"))
	tanggal := strings.TrimSpace(c.QueryParam("tanggal"))
	data, appErr := h.svc.ListJadwal(c.Request().Context(), actorFrom(c), status, kelasID, tanggal)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// CreateJadwal menangani POST /api/v1/jadwal. -> 201
func (h *Handler) CreateJadwal(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "proktor", "guru"); !ok {
		return nil
	}
	var body struct {
		PaketID    string `json:"paket_id"`
		KelasID    string `json:"kelas_id"`
		Tanggal    string `json:"tanggal"`
		JamMulai   string `json:"jam_mulai"`
		JamSelesai string `json:"jam_selesai"`
		Status     string `json:"status"`
	}
	if err := c.Bind(&body); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	j, appErr := h.svc.CreateJadwal(c.Request().Context(), actorFrom(c), JadwalInput{
		PaketID: body.PaketID, KelasID: body.KelasID, Tanggal: body.Tanggal,
		JamMulai: body.JamMulai, JamSelesai: body.JamSelesai, Status: body.Status,
	})
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteCreated(c, j)
}

// GetJadwalDetail menangani GET /api/v1/jadwal/:id.
func (h *Handler) GetJadwalDetail(c *echo.Context) error {
	// penguji_ukk boleh BACA detail jadwal (tanpa token_hash — tidak pernah
	// dikembalikan endpoint ini). Tulis tetap admin/proktor/guru/pengawas.
	if _, _, _, ok := requireRole(c, "admin", "proktor", "guru", "pengawas", "penguji_ukk"); !ok {
		return nil
	}
	id := c.Param("id")
	if !isUUID(id) {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "ID jadwal tidak valid.")
	}
	data, appErr := h.svc.GetJadwalDetail(c.Request().Context(), actorFrom(c), id)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// UpdateJadwal menangani PUT /api/v1/jadwal/:id.
func (h *Handler) UpdateJadwal(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "proktor", "guru"); !ok {
		return nil
	}
	id := c.Param("id")
	if !isUUID(id) {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "ID jadwal tidak valid.")
	}
	var body struct {
		PaketID    string `json:"paket_id"`
		KelasID    string `json:"kelas_id"`
		Tanggal    string `json:"tanggal"`
		JamMulai   string `json:"jam_mulai"`
		JamSelesai string `json:"jam_selesai"`
		Status     string `json:"status"`
	}
	if err := c.Bind(&body); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	data, appErr := h.svc.UpdateJadwal(c.Request().Context(), actorFrom(c), id, JadwalInput{
		PaketID: body.PaketID, KelasID: body.KelasID, Tanggal: body.Tanggal,
		JamMulai: body.JamMulai, JamSelesai: body.JamSelesai, Status: body.Status,
	})
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// DeleteJadwal menangani DELETE /api/v1/jadwal/:id.
func (h *Handler) DeleteJadwal(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "proktor", "guru"); !ok {
		return nil
	}
	id := c.Param("id")
	if !isUUID(id) {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "ID jadwal tidak valid.")
	}
	data, appErr := h.svc.DeleteJadwal(c.Request().Context(), actorFrom(c), id)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// --- Pause / resume ---

// PauseJadwal menangani POST /api/v1/jadwal/:id/pause.
func (h *Handler) PauseJadwal(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "proktor"); !ok {
		return nil
	}
	data, appErr := h.svc.PauseJadwal(c.Request().Context(), actorFrom(c), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// ResumeJadwal menangani POST /api/v1/jadwal/:id/resume.
func (h *Handler) ResumeJadwal(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "proktor"); !ok {
		return nil
	}
	data, appErr := h.svc.ResumeJadwal(c.Request().Context(), actorFrom(c), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// --- Token ---

// GenerateToken menangani POST /api/v1/jadwal/:id/token.
func (h *Handler) GenerateToken(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "proktor"); !ok {
		return nil
	}
	data, appErr := h.svc.GenerateToken(c.Request().Context(), actorFrom(c), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// TokenStatus menangani GET /api/v1/jadwal/:id/token-status.
func (h *Handler) TokenStatus(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "proktor"); !ok {
		return nil
	}
	data, appErr := h.svc.TokenStatus(c.Request().Context(), actorFrom(c), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}
