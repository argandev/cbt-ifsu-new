package exam

import (
	"encoding/json"
	"net/http"
	"strings"

	"cbt-smk-v2/backend/internal/platform"

	"github.com/labstack/echo/v5"
)

// ---------------------------------------------------------------------------
// Ujian siswa
// ---------------------------------------------------------------------------

// UjianAktif menangani GET /api/v1/ujian/aktif.
func (h *Handler) UjianAktif(c *echo.Context) error {
	_, _, siswaID, ok := requireRole(c, "siswa")
	if !ok {
		return nil
	}
	data, appErr := h.svc.UjianAktif(c.Request().Context(), actorFrom(c), siswaID)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// MulaiUjian menangani POST /api/v1/ujian/mulai.
// Attempt baru -> 201; attempt berjalan yang sudah ada -> 200 (idempotent).
func (h *Handler) MulaiUjian(c *echo.Context) error {
	_, _, siswaID, ok := requireRole(c, "siswa")
	if !ok {
		return nil
	}
	var body struct {
		JadwalID string `json:"jadwal_id"`
		Token    string `json:"token"`
	}
	if err := c.Bind(&body); err != nil || body.JadwalID == "" {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "jadwal_id wajib diisi.")
	}
	data, created, appErr := h.svc.MulaiUjian(c.Request().Context(), actorFrom(c), siswaID, body.JadwalID, body.Token)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	if created {
		return platform.WriteCreated(c, data)
	}
	return platform.WriteOK(c, data)
}

// GetAttemptDetail menangani GET /api/v1/ujian/attempt/:id.
func (h *Handler) GetAttemptDetail(c *echo.Context) error {
	_, _, siswaID, ok := requireRole(c, "siswa")
	if !ok {
		return nil
	}
	data, appErr := h.svc.GetAttemptDetail(c.Request().Context(), actorFrom(c), siswaID, c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// SimpanJawaban menangani POST /api/v1/ujian/attempt/:id/jawaban.
func (h *Handler) SimpanJawaban(c *echo.Context) error {
	_, _, siswaID, ok := requireRole(c, "siswa")
	if !ok {
		return nil
	}
	var body struct {
		SoalID  string          `json:"soal_id"`
		Jawaban json.RawMessage `json:"jawaban"`
	}
	if err := c.Bind(&body); err != nil || body.SoalID == "" {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "soal_id dan jawaban wajib diisi.")
	}
	data, appErr := h.svc.SimpanJawaban(c.Request().Context(), actorFrom(c), siswaID, c.Param("id"), body.SoalID, body.Jawaban)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// TandaiRagu menangani POST /api/v1/ujian/attempt/:id/ragu.
func (h *Handler) TandaiRagu(c *echo.Context) error {
	_, _, siswaID, ok := requireRole(c, "siswa")
	if !ok {
		return nil
	}
	var body struct {
		SoalID string `json:"soal_id"`
		Ragu   bool   `json:"ragu"`
	}
	if err := c.Bind(&body); err != nil || body.SoalID == "" {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "soal_id dan ragu wajib diisi.")
	}
	data, appErr := h.svc.TandaiRagu(c.Request().Context(), actorFrom(c), siswaID, c.Param("id"), body.SoalID, body.Ragu)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// CatatPelanggaran menangani POST /api/v1/ujian/attempt/:id/pelanggaran.
func (h *Handler) CatatPelanggaran(c *echo.Context) error {
	_, _, siswaID, ok := requireRole(c, "siswa")
	if !ok {
		return nil
	}
	var body struct {
		Jenis  string `json:"jenis"`
		Detail string `json:"detail"`
	}
	if err := c.Bind(&body); err != nil || strings.TrimSpace(body.Jenis) == "" {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "jenis pelanggaran wajib diisi.")
	}
	data, appErr := h.svc.CatatPelanggaran(c.Request().Context(), actorFrom(c), siswaID, c.Param("id"), body.Jenis, body.Detail)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// SubmitUjian menangani POST /api/v1/ujian/attempt/:id/submit.
func (h *Handler) SubmitUjian(c *echo.Context) error {
	_, _, siswaID, ok := requireRole(c, "siswa")
	if !ok {
		return nil
	}
	data, appErr := h.svc.SubmitUjian(c.Request().Context(), actorFrom(c), siswaID, c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// StatusAttempt menangani GET /api/v1/ujian/attempt/:id/status.
func (h *Handler) StatusAttempt(c *echo.Context) error {
	_, _, siswaID, ok := requireRole(c, "siswa")
	if !ok {
		return nil
	}
	data, appErr := h.svc.StatusAttempt(c.Request().Context(), actorFrom(c), siswaID, c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// HasilAttempt menangani GET /api/v1/ujian/attempt/:id/hasil.
func (h *Handler) HasilAttempt(c *echo.Context) error {
	_, _, siswaID, ok := requireRole(c, "siswa")
	if !ok {
		return nil
	}
	data, appErr := h.svc.HasilAttempt(c.Request().Context(), actorFrom(c), siswaID, c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}
