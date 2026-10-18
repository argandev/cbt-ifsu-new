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

// Analisis menangani GET /api/v1/analisis/:jadwalId.
func (h *Handler) Analisis(c *echo.Context) error {
	_, role, _, ok := requirePerm(c, "analisis.read")
	if !ok {
		if _, _, _, _, okUser := audit.CurrentUser(c); !okUser {
			return platform.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Harus login.")
		}
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat analisis butir soal.")
	}
	jadwalID, err := validJadwalID(c)
	if err != nil {
		return err
	}
	hasil, appErr := h.svc.Analisis(c.Request().Context(), jadwalID, role)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, hasil)
}

// AnalisisCSV menangani GET /api/v1/analisis/:jadwalId/export.csv.
func (h *Handler) AnalisisCSV(c *echo.Context) error {
	_, role, _, ok := requirePerm(c, "analisis.export")
	if !ok {
		if _, _, _, _, okUser := audit.CurrentUser(c); !okUser {
			return platform.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Harus login.")
		}
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin export analisis butir soal.")
	}
	jadwalID, err := validJadwalID(c)
	if err != nil {
		return err
	}
	hasil, appErr := h.svc.AnalisisExport(c.Request().Context(), actorFrom(c), jadwalID, role)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}

	filename := fmt.Sprintf("analisis-butir-%s-%s.csv", jadwalID, time.Now().Format("2006-01-02"))
	c.Response().Header().Set("Content-Type", "text/csv; charset=utf-8")
	c.Response().Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w := csv.NewWriter(c.Response())
	_, _ = c.Response().Write([]byte{0xEF, 0xBB, 0xBF}) // BOM UTF-8 agar Excel benar
	_ = w.Write([]string{"no", "soal_id", "pertanyaan", "tipe", "bobot", "kompetensi",
		"jumlah_penjab", "jumlah_benar", "kesukaran", "label_kesukaran",
		"daya_beda", "interpretasi_daya_beda", "opsi", "pemilih", "persen", "pengecoh_berfungsi"})
	for i, item := range hasil.Items {
		kesukaran := ""
		if item.Kesukaran != nil {
			kesukaran = fmt.Sprintf("%.2f", *item.Kesukaran)
		}
		dayaBeda := ""
		if item.DayaBeda != nil {
			dayaBeda = fmt.Sprintf("%.2f", *item.DayaBeda)
		}
		if item.PengecohStatus == "n/a" || len(item.Pengecoh) == 0 {
			_ = w.Write([]string{
				fmt.Sprint(i + 1), item.SoalID, item.Pertanyaan, item.Tipe,
				fmt.Sprintf("%g", item.Bobot), item.KompetensiKode,
				fmt.Sprint(item.JumlahPenjab), fmt.Sprint(item.JumlahBenar),
				kesukaran, item.LabelKesukaran, dayaBeda, item.InterpretasiDayaBeda,
				"N/A", "", "", "",
			})
			continue
		}
		for _, p := range item.Pengecoh {
			berfungsi := "tidak berfungsi"
			if p.Berfungsi {
				berfungsi = "berfungsi"
			}
			_ = w.Write([]string{
				fmt.Sprint(i + 1), item.SoalID, item.Pertanyaan, item.Tipe,
				fmt.Sprintf("%g", item.Bobot), item.KompetensiKode,
				fmt.Sprint(item.JumlahPenjab), fmt.Sprint(item.JumlahBenar),
				kesukaran, item.LabelKesukaran, dayaBeda, item.InterpretasiDayaBeda,
				p.Opsi, fmt.Sprint(p.Pemilih), fmt.Sprintf("%.2f", p.Persen), berfungsi,
			})
		}
	}
	w.Flush()
	return nil
}
