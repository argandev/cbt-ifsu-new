package exam

import (
	"net/http"
	"strings"

	"cbt-smk-v2/backend/internal/platform"

	"github.com/labstack/echo/v5"
)

// ---------------------------------------------------------------------------
// Paket ujian
// ---------------------------------------------------------------------------

// ListPaket menangani GET /api/v1/paket.
func (h *Handler) ListPaket(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "guru"); !ok {
		return nil
	}
	jenis := strings.TrimSpace(c.QueryParam("jenis"))
	mapelID := strings.TrimSpace(c.QueryParam("mapel_id"))
	data, appErr := h.svc.ListPaket(c.Request().Context(), actorFrom(c), jenis, mapelID)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// CreatePaket menangani POST /api/v1/paket. -> 201
func (h *Handler) CreatePaket(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "guru"); !ok {
		return nil
	}
	var body struct {
		Kode        string `json:"kode"`
		Nama        string `json:"nama"`
		MapelID     string `json:"mapel_id"`
		Jenis       string `json:"jenis"`
		DurasiMenit int    `json:"durasi_menit"`
		JumlahSoal  int    `json:"jumlah_soal"`
		AcakSoal    *bool  `json:"acak_soal"`
		AcakJawaban *bool  `json:"acak_jawaban"`
		TampilHasil *bool  `json:"tampil_hasil"`
	}
	if err := c.Bind(&body); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	p, appErr := h.svc.CreatePaket(c.Request().Context(), actorFrom(c), PaketInput{
		Kode: body.Kode, Nama: body.Nama, MapelID: body.MapelID, Jenis: body.Jenis,
		DurasiMenit: body.DurasiMenit, JumlahSoal: body.JumlahSoal,
		AcakSoal: body.AcakSoal, AcakJawaban: body.AcakJawaban, TampilHasil: body.TampilHasil,
	})
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteCreated(c, p)
}

// GetPaket menangani GET /api/v1/paket/:id.
func (h *Handler) GetPaket(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "guru"); !ok {
		return nil
	}
	id := c.Param("id")
	if !isUUID(id) {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "ID paket tidak valid.")
	}
	p, appErr := h.svc.GetPaket(c.Request().Context(), actorFrom(c), id)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, p)
}

// UpdatePaket menangani PUT /api/v1/paket/:id.
func (h *Handler) UpdatePaket(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "guru"); !ok {
		return nil
	}
	id := c.Param("id")
	if !isUUID(id) {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "ID paket tidak valid.")
	}
	var body struct {
		Kode        string `json:"kode"`
		Nama        string `json:"nama"`
		MapelID     string `json:"mapel_id"`
		Jenis       string `json:"jenis"`
		DurasiMenit int    `json:"durasi_menit"`
		JumlahSoal  int    `json:"jumlah_soal"`
		AcakSoal    *bool  `json:"acak_soal"`
		AcakJawaban *bool  `json:"acak_jawaban"`
		TampilHasil *bool  `json:"tampil_hasil"`
	}
	if err := c.Bind(&body); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	data, appErr := h.svc.UpdatePaket(c.Request().Context(), actorFrom(c), id, PaketInput{
		Kode: body.Kode, Nama: body.Nama, MapelID: body.MapelID, Jenis: body.Jenis,
		DurasiMenit: body.DurasiMenit, JumlahSoal: body.JumlahSoal,
		AcakSoal: body.AcakSoal, AcakJawaban: body.AcakJawaban, TampilHasil: body.TampilHasil,
	})
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// DeletePaket menangani DELETE /api/v1/paket/:id.
func (h *Handler) DeletePaket(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "guru"); !ok {
		return nil
	}
	id := c.Param("id")
	if !isUUID(id) {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "ID paket tidak valid.")
	}
	data, appErr := h.svc.DeletePaket(c.Request().Context(), actorFrom(c), id)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// ListPaketSoal menangani GET /api/v1/paket/:id/soal.
func (h *Handler) ListPaketSoal(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "guru"); !ok {
		return nil
	}
	data, appErr := h.svc.ListPaketSoal(c.Request().Context(), actorFrom(c), c.Param("id"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// AddPaketSoal menangani POST /api/v1/paket/:id/soal. -> 201
func (h *Handler) AddPaketSoal(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "guru"); !ok {
		return nil
	}
	var body struct {
		SoalID string `json:"soal_id"`
	}
	if err := c.Bind(&body); err != nil || body.SoalID == "" {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "soal_id wajib diisi.")
	}
	data, appErr := h.svc.AddPaketSoal(c.Request().Context(), actorFrom(c), c.Param("id"), body.SoalID)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteCreated(c, data)
}

// DeletePaketSoal menangani DELETE /api/v1/paket/:id/soal.
func (h *Handler) DeletePaketSoal(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "guru"); !ok {
		return nil
	}
	paketID := c.Param("id")
	soalID := strings.TrimSpace(c.QueryParam("soal_id"))
	if soalID == "" {
		var body struct {
			SoalID string `json:"soal_id"`
		}
		_ = c.Bind(&body)
		soalID = body.SoalID
	}
	if soalID == "" {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "soal_id wajib diisi (body atau ?soal_id=).")
	}
	data, appErr := h.svc.DeletePaketSoal(c.Request().Context(), actorFrom(c), paketID, soalID)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// GenerateAcak menangani POST /api/v1/paket/:id/generate-acak.
func (h *Handler) GenerateAcak(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "guru"); !ok {
		return nil
	}
	var body struct {
		Jumlah int `json:"jumlah"`
	}
	if err := c.Bind(&body); err != nil || body.Jumlah <= 0 {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "jumlah harus > 0.")
	}
	data, appErr := h.svc.GenerateAcak(c.Request().Context(), actorFrom(c), c.Param("id"), body.Jumlah)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}
