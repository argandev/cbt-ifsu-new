package analisis

import (
	"net/http"
	"strings"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/config"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// --- Handler ----------------------------------------------------------------

// Handler adalah lapisan HTTP paket analisis.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// requirePerm: 401 bila belum login (context kosong), 403 bila tidak
// punya permission — sesuai KONTRAK.md §1.2.
// (Dipindah dari analisis.go tanpa perubahan; TIDAK menulis respons
// sendiri — handler yang memutuskan 401 vs 403.)
func requirePerm(c *echo.Context, perm string) (userID, role, siswaID string, ok bool) {
	var okUser bool
	userID, role, _, siswaID, okUser = audit.CurrentUser(c)
	if !okUser {
		return "", "", "", false
	}
	if !audit.HasPermission(c, perm) {
		return "", "", "", false
	}
	return userID, role, siswaID, true
}

func validJadwalID(c *echo.Context) (string, error) {
	id := strings.TrimSpace(c.Param("jadwalId"))
	if _, err := uuid.Parse(id); err != nil {
		return "", platform.WriteError(c, http.StatusBadRequest, "validation_error", "ID jadwal tidak valid.")
	}
	return id, nil
}

// actorFrom membangun RequestActor dari echo.Context.
func actorFrom(c *echo.Context) RequestActor {
	userID, _, _, _, _ := audit.CurrentUser(c)
	return RequestActor{UserID: userID, RealIP: c.RealIP()}
}

// Register mendaftarkan endpoint analisis butir soal + nilai per kompetensi.
//
//	GET /api/v1/analisis/:jadwalId
//	GET /api/v1/analisis/:jadwalId/export.csv
//	GET /api/v1/kompetensi/:jadwalId
//	GET /api/v1/kompetensi/:jadwalId/export.csv
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q
	_ = cfg

	repo := NewRepository(pool)
	svc := NewService(repo, pool)
	h := NewHandler(svc)

	// GET /api/v1/analisis/:jadwalId
	e.GET("/api/v1/analisis/:jadwalId", h.Analisis)

	// GET /api/v1/analisis/:jadwalId/export.csv
	e.GET("/api/v1/analisis/:jadwalId/export.csv", h.AnalisisCSV)

	// GET /api/v1/kompetensi/:jadwalId
	e.GET("/api/v1/kompetensi/:jadwalId", h.Kompetensi)

	// GET /api/v1/kompetensi/:jadwalId/export.csv
	e.GET("/api/v1/kompetensi/:jadwalId/export.csv", h.KompetensiCSV)
}
