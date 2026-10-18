package ukk

import (
	"net/http"

	"cbt-smk-v2/backend/internal/platform"

	"github.com/labstack/echo/v5"
)

// ListNilai menangani GET /api/v1/ukk/:id/nilai.
func (h *Handler) ListNilai(c *echo.Context) error {
	if _, _, ok := requirePerm(c, "ukk.read"); !ok {
		return nil
	}
	id := c.Param("id")
	if !isValidUUID(id) {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "ID UKK tidak valid.")
	}
	items, appErr := h.svc.ListNilai(c.Request().Context(), id)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, items)
}

// ListPeserta menangani GET /api/v1/ukk/:id/peserta (izin ukk.read).
// D5: penguji_ukk WAJIB bisa tambah nilai siswa baru. Dropdown halaman
// nilai sebelumnya memanggil /master/siswa (izin master.siswa.read)
// sehingga penguji_ukk 403 dan tidak bisa tambah nilai baru. Endpoint
// khusus UKK ini memberi daftar siswa kelas UKK tersebut di bawah izin
// ukk.* yang sudah dimiliki penguji_ukk — TANPA menambah izin master
// apa pun dan TANPA mengubah endpoint master yang sudah ada.
func (h *Handler) ListPeserta(c *echo.Context) error {
	if _, _, ok := requirePerm(c, "ukk.read"); !ok {
		return nil
	}
	id := c.Param("id")
	if !isValidUUID(id) {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "ID UKK tidak valid.")
	}
	items, appErr := h.svc.ListPeserta(c.Request().Context(), id)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, items)
}

// UpsertNilai menangani POST /ukk/:id/nilai (nilaiID kosong, upsert per
// siswa) dan PUT /ukk/:id/nilai/:nilaiId (per nilai id).
func (h *Handler) UpsertNilai(c *echo.Context, nilaiID string) error {
	if _, _, ok := requirePerm(c, "ukk.update"); !ok {
		return nil
	}
	ukkID := c.Param("id")
	if !isValidUUID(ukkID) {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "ID UKK tidak valid.")
	}
	if nilaiID != "" && !isValidUUID(nilaiID) {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "ID nilai tidak valid.")
	}
	var in nilaiInput
	if err := c.Bind(&in); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	row, created, appErr := h.svc.UpsertNilai(c.Request().Context(), actorFrom(c), ukkID, nilaiID, in)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	if nilaiID != "" {
		return platform.WriteOK(c, row)
	}
	if created {
		return platform.WriteStatus(c, http.StatusCreated, row)
	}
	return platform.WriteOK(c, row)
}
