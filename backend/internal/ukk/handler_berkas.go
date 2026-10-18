package ukk

import (
	"net/http"
	"path/filepath"
	"strings"

	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/labstack/echo/v5"
)

// ListBerkas menangani GET /api/v1/ukk/:id/berkas.
// Siswa hanya melihat berkas miliknya sendiri.
func (h *Handler) ListBerkas(c *echo.Context) error {
	userID, role, ok := requireLogin(c)
	if !ok {
		return nil
	}
	ukkID := c.Param("id")
	if !isValidUUID(ukkID) {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "ID UKK tidak valid.")
	}
	ctx := c.Request().Context()
	if _, appErr := h.svc.GetUkk(ctx, ukkID); appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	// Siswa hanya melihat berkas miliknya sendiri.
	var filterSiswa *string
	if role == "siswa" {
		siswaID, found := h.svc.ResolveSiswaID(ctx, userID)
		if !found {
			return platform.WriteError(c, http.StatusForbidden, "forbidden", "Profil siswa tidak ditemukan.")
		}
		filterSiswa = &siswaID
	} else if !audit.HasPermission(c, "ukk.read") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin: ukk.read.")
	}
	items, appErr := h.svc.ListBerkas(ctx, ukkID, filterSiswa)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, items)
}

// UploadBerkas menangani POST /api/v1/ukk/:id/berkas.
// Siswa: hanya untuk dirinya sendiri. Selain siswa: wajib ukk.create.
func (h *Handler) UploadBerkas(c *echo.Context) error {
	userID, role, ok := requireLogin(c)
	if !ok {
		return nil
	}
	ukkID := c.Param("id")
	if !isValidUUID(ukkID) {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "ID UKK tidak valid.")
	}
	ctx := c.Request().Context()
	if _, appErr := h.svc.GetUkk(ctx, ukkID); appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	// Siswa: hanya untuk dirinya sendiri. Selain siswa: wajib ukk.create.
	var callerSiswaID string
	if role == "siswa" {
		sid, found := h.svc.ResolveSiswaID(ctx, userID)
		if !found {
			return platform.WriteError(c, http.StatusForbidden, "forbidden", "Profil siswa tidak ditemukan.")
		}
		callerSiswaID = sid
	} else if !audit.HasPermission(c, "ukk.create") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin: ukk.create.")
	}
	if err := c.Request().ParseMultipartForm(maxBerkasBytes); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body multipart tidak valid (maks. 10 MB).")
	}
	siswaID := strings.TrimSpace(c.FormValue("siswa_id"))
	if role == "siswa" {
		if siswaID == "" {
			siswaID = callerSiswaID
		}
		if siswaID != callerSiswaID {
			return platform.WriteError(c, http.StatusForbidden, "forbidden", "Siswa hanya boleh upload berkas untuk dirinya sendiri.")
		}
	} else if !isValidUUID(siswaID) {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "siswa_id tidak valid.")
	}
	if !h.svc.SiswaExists(ctx, siswaID) {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Siswa tidak ditemukan.")
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Field file wajib diisi.")
	}
	if fileHeader.Size > maxBerkasBytes {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Ukuran berkas melebihi 10 MB.")
	}
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowedExt[ext] {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error",
			"Ekstensi tidak diizinkan. Gunakan: pdf, doc, docx, jpg, png, zip.")
	}
	b, appErr := h.svc.UploadBerkas(ctx, actorFrom(c), ukkID, siswaID, fileHeader)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteStatus(c, http.StatusCreated, b)
}

// DownloadBerkas menangani GET /api/v1/ukk/:id/berkas/:berkasId/download.
// Siswa hanya boleh mengunduh berkas miliknya sendiri.
func (h *Handler) DownloadBerkas(c *echo.Context) error {
	if _, _, ok := requireLogin(c); !ok {
		return nil
	}
	ukkID := c.Param("id")
	berkasID := c.Param("berkasId")
	if !isValidUUID(ukkID) || !isValidUUID(berkasID) {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "ID tidak valid.")
	}
	// Izin staf dihitung di sini (butuh echo.Context); service memakai
	// hasilnya setelah pemeriksaan 404 — urutan dipertahankan.
	bolehStaf := audit.HasPermission(c, "ukk.read")
	b, appErr := h.svc.DownloadBerkas(c.Request().Context(), actorFrom(c), ukkID, berkasID, bolehStaf)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return c.Attachment(b.AbsPath, b.NamaFile)
}
