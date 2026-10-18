package banksoal

import (
	"net/http"
	"strings"

	"cbt-smk-v2/backend/internal/platform"

	"github.com/labstack/echo/v5"
)

// ListSoal menangani GET /api/v1/bank-soal/:bankId/soal.
func (h *Handler) ListSoal(c *echo.Context) error {
	_, role, ok := requirePerm(c, "banksoal.read")
	if !ok {
		return nil
	}
	if !forbidSoalUntukProktor(c, role) {
		return nil
	}
	limit, offset := platform.ParseLimit(c, 20)
	items, total, appErr := h.svc.ListSoal(c.Request().Context(), c.Param("bankId"), soalFilterArgs(c), limit, offset)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return writeItemsList(c, items, total, limit, offset)
}

// CreateSoal menangani POST /api/v1/bank-soal/:bankId/soal.
func (h *Handler) CreateSoal(c *echo.Context) error {
	userID, role, ok := requirePerm(c, "banksoal.create")
	if !ok {
		return nil
	}
	if !forbidSoalUntukProktor(c, role) {
		return nil
	}
	var in soalInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "invalid_body", "Body JSON tidak valid.")
	}
	s, appErr := h.svc.CreateSoal(c.Request().Context(), RequestActor{UserID: userID, RealIP: c.RealIP()}, role, c.Param("bankId"), in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteCreated(c, s)
}

// GetSoal menangani GET /api/v1/bank-soal/:bankId/soal/:soalId.
func (h *Handler) GetSoal(c *echo.Context) error {
	_, role, ok := requirePerm(c, "banksoal.read")
	if !ok {
		return nil
	}
	if !forbidSoalUntukProktor(c, role) {
		return nil
	}
	s, appErr := h.svc.GetSoal(c.Request().Context(), c.Param("bankId"), c.Param("soalId"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, s)
}

// UpdateSoal menangani PUT /api/v1/bank-soal/:bankId/soal/:soalId.
func (h *Handler) UpdateSoal(c *echo.Context) error {
	userID, role, ok := requirePerm(c, "banksoal.update")
	if !ok {
		return nil
	}
	if !forbidSoalUntukProktor(c, role) {
		return nil
	}
	// Urutan sesuai kode asal: load bank/soal (404) + cek scope/kunci dulu, baru bind (400).
	b, sr, appErr := h.svc.LoadSoalChecked(c.Request().Context(), role, userID, c.Param("bankId"), c.Param("soalId"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	var in soalInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "invalid_body", "Body JSON tidak valid.")
	}
	updated, appErr := h.svc.ApplySoalUpdate(c.Request().Context(), RequestActor{UserID: userID, RealIP: c.RealIP()}, b, sr, in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, updated)
}

// DeleteSoal menangani DELETE /api/v1/bank-soal/:bankId/soal/:soalId.
func (h *Handler) DeleteSoal(c *echo.Context) error {
	userID, role, ok := requirePerm(c, "banksoal.delete")
	if !ok {
		return nil
	}
	if !forbidSoalUntukProktor(c, role) {
		return nil
	}
	out, appErr := h.svc.DeleteSoal(c.Request().Context(), RequestActor{UserID: userID, RealIP: c.RealIP()}, role, c.Param("bankId"), c.Param("soalId"))
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, out)
}

// ImportCSV menangani POST /api/v1/bank-soal/import.
func (h *Handler) ImportCSV(c *echo.Context) error {
	userID, role, ok := requirePerm(c, "banksoal.import")
	if !ok {
		return nil
	}
	if !forbidSoalUntukProktor(c, role) {
		return nil
	}
	var body struct {
		BankID  string `json:"bank_id"`
		CSVText string `json:"csv_text"`
		CSV     string `json:"csv"`
		Mode    string `json:"mode"` // all_or_nothing (default) | skip_errors
	}
	if err := c.Bind(&body); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "invalid_body", "Body JSON tidak valid.")
	}
	csvText := body.CSVText
	if csvText == "" {
		csvText = body.CSV
	}
	if !isValidUUID(body.BankID) {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "bank_id wajib UUID bank yang valid.")
	}
	if strings.TrimSpace(csvText) == "" {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "csv_text wajib diisi (teks CSV).")
	}
	mode := strings.ToLower(strings.TrimSpace(body.Mode))
	if mode == "" {
		mode = "all_or_nothing"
	}
	if mode != "all_or_nothing" && mode != "skip_errors" {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "mode harus all_or_nothing atau skip_errors.")
	}
	out, appErr := h.svc.ImportCSV(c.Request().Context(), RequestActor{UserID: userID, RealIP: c.RealIP()}, role, body.BankID, csvText, mode)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, out)
}
