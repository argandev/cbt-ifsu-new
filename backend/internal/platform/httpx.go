package platform

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
)

// WriteError menulis envelope error — IDENTIK dengan writeError yang
// sebelumnya diduplikasi di tiap paket:
//
//	{"error": {"code": ..., "message": ...}}
func WriteError(c *echo.Context, status int, code, message string) error {
	return c.JSON(status, map[string]any{
		"error": map[string]any{"code": code, "message": message},
	})
}

// WriteAppError memetakan error ke envelope error yang sudah ada.
// *AppError -> status + code + message dari error itu sendiri.
// Error lain -> 500 {"error": {"code": "internal_error", ...}}.
func WriteAppError(c *echo.Context, err error) error {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return WriteError(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
	}
	return WriteError(c, http.StatusInternalServerError, "internal_error", "Terjadi kesalahan internal.")
}

// WriteOK menulis {"data": ...} dengan status 200 — IDENTIK dengan
// writeOK yang sebelumnya diduplikasi di tiap paket.
func WriteOK(c *echo.Context, data any) error {
	return c.JSON(http.StatusOK, map[string]any{"data": data})
}

// WriteStatus menulis {"data": ...} dengan status tertentu — IDENTIK
// dengan writeOKStatus (mis. dipakai untuk 201 Created).
func WriteStatus(c *echo.Context, status int, data any) error {
	return c.JSON(status, map[string]any{"data": data})
}

// WriteCreated menulis {"data": ...} dengan status 201.
func WriteCreated(c *echo.Context, data any) error {
	return WriteStatus(c, http.StatusCreated, data)
}
