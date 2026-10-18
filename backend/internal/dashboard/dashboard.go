// Package dashboard — angka ringkasan untuk kartu Dashboard /adm.
//
// Sebelumnya kartu Dashboard (Bank Soal, Jadwal Ujian, Peserta, Rekap)
// di-hardcode "-" di frontend karena belum ada endpoint agregat — terlihat
// seperti data rusak. Endpoint ini mengembalikan hitungan nyata dari
// database. Hanya staf (role apa pun kecuali "siswa") yang boleh membaca:
// angka agregat seluruh sekolah bukan data milik siswa sendiri.
package dashboard

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"

	"cbt-smk-v2/backend/internal/audit"
)

func writeError(c *echo.Context, status int, code, message string) error {
	return c.JSON(status, map[string]any{
		"error": map[string]any{"code": code, "message": message},
	})
}

func writeOK(c *echo.Context, data any) error {
	return c.JSON(http.StatusOK, map[string]any{"data": data})
}

// Register mendaftarkan endpoint ringkasan dashboard.
func Register(e *echo.Echo, pool *pgxpool.Pool) {
	// GET /api/v1/dashboard/ringkasan
	e.GET("/api/v1/dashboard/ringkasan", func(c *echo.Context) error {
		_, role, _, _, ok := audit.CurrentUser(c)
		if !ok {
			return writeError(c, http.StatusUnauthorized, "unauthorized", "Silakan login dulu.")
		}
		if role == "siswa" {
			return writeError(c, http.StatusForbidden, "forbidden", "Ringkasan dashboard hanya untuk staf.")
		}
		ctx := c.Request().Context()
		hitung := func(tabel string) (int, error) {
			var n int
			// Nama tabel berasal dari konstanta di bawah, bukan input user.
			err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM `+tabel).Scan(&n)
			return n, err
		}
		bankSoal, err := hitung("bank_soal")
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghitung bank soal.")
		}
		jadwal, err := hitung("jadwal_ujian")
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghitung jadwal ujian.")
		}
		peserta, err := hitung("siswa")
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghitung peserta.")
		}
		rekap, err := hitung("nilai_akhir")
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghitung rekap.")
		}
		return writeOK(c, map[string]any{
			// Rekap = jumlah hasil ujian yang sudah direkap (baris nilai_akhir).
			"bank_soal":    bankSoal,
			"jadwal_ujian": jadwal,
			"peserta":      peserta,
			"rekap":        rekap,
		})
	})
}
