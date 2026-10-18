package platform

import "net/http"

// AppError adalah error aplikasi terstruktur yang dibawa dari service/
// repository ke handler. Handler memetakannya ke envelope error JSON
// yang sudah ada via WriteAppError — tanpa mengubah bentuk response.
type AppError struct {
	Code       string // kode error untuk JSON, mis. "tidak_ditemukan"
	Message    string // pesan untuk pengguna (Bahasa Indonesia)
	HTTPStatus int    // status HTTP
}

func (e *AppError) Error() string { return e.Code + ": " + e.Message }

// NewAppError membuat AppError.
func NewAppError(code, message string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: httpStatus}
}

// Konstruktor umum — memetakan ke kode & status yang sudah dipakai
// di seluruh aplikasi.

// ErrBadRequest: 400 — input tidak valid.
func ErrBadRequest(code, message string) *AppError {
	return NewAppError(code, message, http.StatusBadRequest)
}

// ErrUnauthorized: 401 — belum login.
func ErrUnauthorized(code, message string) *AppError {
	return NewAppError(code, message, http.StatusUnauthorized)
}

// ErrForbidden: 403 — tidak punya izin.
func ErrForbidden(code, message string) *AppError {
	return NewAppError(code, message, http.StatusForbidden)
}

// ErrNotFound: 404 — data tidak ditemukan.
func ErrNotFound(code, message string) *AppError {
	return NewAppError(code, message, http.StatusNotFound)
}

// ErrConflict: 409 — konflik (mis. duplikat unik).
func ErrConflict(code, message string) *AppError {
	return NewAppError(code, message, http.StatusConflict)
}

// ErrInternal: 500 — kesalahan internal.
func ErrInternal(code, message string) *AppError {
	return NewAppError(code, message, http.StatusInternalServerError)
}
