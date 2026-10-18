// Package rekap — rekap nilai, berita acara, export CSV, finalisasi, remedial.
//
// Asumsi skema tim lain (migrasi 0001-0004):
//
//	jadwal_ujian(id, paket_id, kelas_id, ...), mapel(id, nama),
//	siswa(id, nisn, nama, kelas_id), kelas(id, nama),
//	attempt(id, jadwal_id, siswa_id, status),
//	jawaban(id, attempt_id, ...), settings(key, value jsonb).
//
// Status attempt yang dihitung "hadir" di berita acara: 'selesai', 'diblokir'.
// Predikat (ANALISIS Bab 8 / UKK): <70 Belum Tuntas, 70-79 Cukup,
// 80-90 Baik, 91-100 Sangat Baik. Ambang lulus rekap: nilai >= 70.
//
// Struktur layer (hasil refactor arsitektur):
//   - handler.go:    Echo handler + Register
//   - service.go:    Predikat + statistik + logika rekap/berita-acara/
//     finalisasi/remedial + audit log
//   - repository.go: DTO + semua query SQL
//   - platform:      envelope JSON (menggantikan helper lokal)
//
// Perilaku (route, status, bentuk JSON, pesan error) IDENTIK dengan
// sebelum refactor.
package rekap

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/config"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// --- Handler ----------------------------------------------------------------

// Handler adalah lapisan HTTP paket rekap.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// actorFrom membangun platform.RequestActor dari echo.Context.
func actorFrom(c *echo.Context) platform.RequestActor {
	userID, _, _, _, _ := audit.CurrentUser(c)
	return platform.RequestActor{UserID: userID, RealIP: c.RealIP()}
}

// Register mendaftarkan endpoint rekap.
//
//	GET  /api/v1/rekap?jadwal_id=
//	GET  /api/v1/rekap/:jadwalId/export.csv
//	GET  /api/v1/rekap/:jadwalId/berita-acara
//	POST /api/v1/rekap/:jadwalId/finalisasi
//	POST /api/v1/rekap/:jadwalId/remedial
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q
	_ = cfg

	repo := NewRepository(pool)
	svc := NewService(repo, pool)
	h := NewHandler(svc)

	// GET /api/v1/rekap?jadwal_id=
	e.GET("/api/v1/rekap", h.Rekap)

	// GET /api/v1/rekap/:jadwalId/export.csv
	e.GET("/api/v1/rekap/:jadwalId/export.csv", h.ExportCSV)

	// GET /api/v1/rekap/:jadwalId/berita-acara
	e.GET("/api/v1/rekap/:jadwalId/berita-acara", h.BeritaAcara)

	// POST /api/v1/rekap/:jadwalId/finalisasi
	e.POST("/api/v1/rekap/:jadwalId/finalisasi", h.Finalisasi)

	// POST /api/v1/rekap/:jadwalId/remedial {siswa_id, nilai_baru}
	e.POST("/api/v1/rekap/:jadwalId/remedial", h.Remedial)
}

// Rekap menangani GET /api/v1/rekap?jadwal_id=.
func (h *Handler) Rekap(c *echo.Context) error {
	jadwalID := strings.TrimSpace(c.QueryParam("jadwal_id"))
	if jadwalID == "" {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Parameter jadwal_id wajib diisi.")
	}
	_, role, _, siswaID, ok := audit.CurrentUser(c)
	if !ok {
		return platform.WriteError(c, http.StatusUnauthorized, "unauthorized", "Harus login.")
	}
	ctx := c.Request().Context()
	if role == "siswa" {
		// Siswa hanya melihat nilainya sendiri, hanya jika final + tampil_hasil.
		rows, stats, appErr := h.svc.RekapSiswa(ctx, jadwalID, siswaID)
		if appErr != nil {
			return platform.WriteAppError(c, appErr)
		}
		return platform.WriteOK(c, map[string]any{"items": rows, "statistik": stats})
	}
	if !audit.HasPermission(c, "rekap.read") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat rekap.")
	}
	rows, stats, appErr := h.svc.RekapStaf(ctx, jadwalID)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, map[string]any{"items": rows, "statistik": stats})
}

// ExportCSV menangani GET /api/v1/rekap/:jadwalId/export.csv.
func (h *Handler) ExportCSV(c *echo.Context) error {
	jadwalID := c.Param("jadwalId")
	if !audit.HasPermission(c, "rekap.export") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin export rekap.")
	}
	rows, appErr := h.svc.ExportData(c.Request().Context(), actorFrom(c), jadwalID)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}

	filename := fmt.Sprintf("rekap-%s-%s.csv", jadwalID, time.Now().Format("2006-01-02"))
	c.Response().Header().Set("Content-Type", "text/csv; charset=utf-8")
	c.Response().Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w := csv.NewWriter(c.Response())
	// BOM UTF-8 agar Excel membuka file dengan benar.
	_, _ = c.Response().Write([]byte{0xEF, 0xBB, 0xBF})
	_ = w.Write([]string{"nisn", "nama", "kelas", "nilai", "predikat", "status"})
	for _, r := range rows {
		pred := ""
		if r.Predikat != nil {
			pred = *r.Predikat
		}
		_ = w.Write([]string{r.NISN, r.NamaSiswa, r.NamaKelas,
			strconv.FormatFloat(r.Nilai, 'f', 2, 64), pred, r.Status})
	}
	w.Flush()
	return nil
}

// BeritaAcara menangani GET /api/v1/rekap/:jadwalId/berita-acara.
func (h *Handler) BeritaAcara(c *echo.Context) error {
	jadwalID := c.Param("jadwalId")
	if !audit.HasPermission(c, "rekap.read") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat berita acara.")
	}
	data, appErr := h.svc.BeritaAcara(c.Request().Context(), jadwalID)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}

// Finalisasi menangani POST /api/v1/rekap/:jadwalId/finalisasi.
func (h *Handler) Finalisasi(c *echo.Context) error {
	jadwalID := c.Param("jadwalId")
	if !audit.HasPermission(c, "rekap.finalisasi") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin finalisasi nilai.")
	}
	count, appErr := h.svc.Finalisasi(c.Request().Context(), actorFrom(c), jadwalID)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, map[string]any{"jadwal_id": jadwalID, "jumlah_difinalisasi": count, "status": "final"})
}

// Remedial menangani POST /api/v1/rekap/:jadwalId/remedial.
func (h *Handler) Remedial(c *echo.Context) error {
	jadwalID := c.Param("jadwalId")
	if !audit.HasPermission(c, "rekap.remedial") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin remedial.")
	}
	var body struct {
		SiswaID   string  `json:"siswa_id"`
		NilaiBaru float64 `json:"nilai_baru"`
	}
	if err := c.Bind(&body); err != nil {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	if strings.TrimSpace(body.SiswaID) == "" {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "siswa_id wajib diisi.")
	}
	if body.NilaiBaru < 0 || body.NilaiBaru > 100 {
		return platform.WriteError(c, http.StatusBadRequest, "validation_error", "nilai_baru harus 0-100.")
	}
	data, appErr := h.svc.Remedial(c.Request().Context(), actorFrom(c), jadwalID, body.SiswaID, body.NilaiBaru)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, data)
}
