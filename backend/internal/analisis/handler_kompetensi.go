package analisis

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/labstack/echo/v5"
)

// Kompetensi menangani GET /api/v1/kompetensi/:jadwalId.
func (h *Handler) Kompetensi(c *echo.Context) error {
	_, role, siswaID, ok := requirePerm(c, "rekap.read")
	if !ok {
		if _, _, _, _, okUser := audit.CurrentUser(c); !okUser {
			return platform.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Harus login.")
		}
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat nilai per kompetensi.")
	}
	jadwalID, err := validJadwalID(c)
	if err != nil {
		return err
	}
	hasil, appErr := h.svc.Kompetensi(c.Request().Context(), jadwalID, role, siswaID)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, hasil)
}

// KompetensiCSV menangani GET /api/v1/kompetensi/:jadwalId/export.csv.
func (h *Handler) KompetensiCSV(c *echo.Context) error {
	_, role, siswaID, ok := requirePerm(c, "rekap.export")
	if !ok {
		if _, _, _, _, okUser := audit.CurrentUser(c); !okUser {
			return platform.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Harus login.")
		}
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin export nilai per kompetensi.")
	}
	jadwalID, err := validJadwalID(c)
	if err != nil {
		return err
	}
	hasil, appErr := h.svc.KompetensiExport(c.Request().Context(), actorFrom(c), jadwalID, role, siswaID)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}

	filename := fmt.Sprintf("nilai-kompetensi-%s-%s.csv", jadwalID, time.Now().Format("2006-01-02"))
	c.Response().Header().Set("Content-Type", "text/csv; charset=utf-8")
	c.Response().Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w := csv.NewWriter(c.Response())
	_, _ = c.Response().Write([]byte{0xEF, 0xBB, 0xBF})
	_ = w.Write([]string{"nisn", "nama", "kelas", "kompetensi_kode", "kompetensi",
		"nilai", "jumlah_soal", "dijawab", "flag", "rata_rata_siswa"})
	for _, s := range hasil.Siswa {
		rata := ""
		if s.RataRata != nil {
			rata = fmt.Sprintf("%.2f", *s.RataRata)
		}
		for _, k := range s.NilaiPerKompetensi {
			nilai := ""
			if k.Nilai != nil {
				nilai = fmt.Sprintf("%.2f", *k.Nilai)
			}
			_ = w.Write([]string{
				s.NISN, s.Nama, s.Kelas, k.KompetensiKode, k.Kompetensi,
				nilai, fmt.Sprint(k.JumlahSoal), fmt.Sprint(k.Dijawab), k.Flag, rata,
			})
		}
	}
	w.Flush()
	return nil
}
