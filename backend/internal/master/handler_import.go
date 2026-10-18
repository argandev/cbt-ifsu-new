package master

import (
	"net/http"

	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/labstack/echo/v5"
)

// ---------------------------------------------------------------------
// Import Siswa dari CSV
// ---------------------------------------------------------------------

// InfoImportSiswa menangani GET /api/v1/master/siswa/import.
//
// Mengembalikan format CSV yang diterima POST /siswa/import. Route ini
// WAJIB didaftarkan SEBELUM GET /siswa/:id: segmen statis "import"
// bila tidak ada route-nya akan tertangkap param :id ("import" bukan
// UUID) dan getSiswa mengembalikan 404 "Siswa tidak ditemukan".
func (h *Handler) InfoImportSiswa(c *echo.Context) error {
	if !audit.HasPermission(c, "master.siswa.import") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin import siswa.")
	}
	return platform.WriteOK(c, map[string]any{
		"method":  "POST",
		"path":    "/api/v1/master/siswa/import",
		"body":    "multipart/form-data, field \"file\" berisi CSV UTF-8",
		"kolom":   []string{"nama", "nisn", "nis (opsional)", "kelas_id / kelas_kode / kelas / kode_kelas"},
		"catatan": "kelas_kode = id UUID kelas ATAU nama kelas. Akun login dibuat otomatis: username = NISN, password awal = NISN (wajib ganti saat login pertama). Import IDEMPOTEN: NISN yang sudah ada dihitung sudah_ada (bukan gagal). Rekomendasi: maks. 100 siswa per file.",
		"contoh":  "nama,nisn,nis,kelas_kode\nBudi Santoso,0012345678,1234,XI RPL 1",
	})
}

// ImportSiswa menangani POST /api/v1/master/siswa/import.
// Form multipart: field "file" berisi CSV UTF-8.
// Header yang diterima:
//
//	nama,nisn,nis,kelas_kode   (kelas_kode = id UUID kelas ATAU nama kelas)
//	nama,nisn,kelas_id         (alias: kelas, kode_kelas)
//
// Kolom nis opsional. Baris yang gagal dilewati + dilaporkan per baris;
// baris yang valid disimpan dalam satu transaksi.
//
// IDEMPOTEN (Fase 3, BUG-02): NISN yang sudah ada di DB dihitung sebagai
// `sudah_ada` (bukan error, bukan duplikat, tidak buat user baru) —
// retry import yang sama aman. Rekomendasi operasional: import
// maksimal 100 siswa per file (±17 dtk, jauh di bawah WriteTimeout
// server 120 dtk; lihat KONTRAK §10) lalu ulangi file berikutnya.
func (h *Handler) ImportSiswa(c *echo.Context) error {
	if !audit.HasPermission(c, "master.siswa.import") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin import siswa.")
	}
	if err := c.Request().ParseMultipartForm(maxImportFileMB << 20); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body multipart tidak valid (maks. 5 MB).")
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Field file (CSV) wajib diunggah.")
	}
	f, err := fileHeader.Open()
	if err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "File CSV tidak bisa dibuka.")
	}
	defer f.Close()
	result, appErr := h.svc.ImportSiswa(c.Request().Context(), actorFrom(c), f)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, result)
}
