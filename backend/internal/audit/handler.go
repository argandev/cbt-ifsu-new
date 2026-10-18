package audit

import (
	"net/http"
	"strconv"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/config"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// --- Handler ----------------------------------------------------------------

// Handler adalah lapisan HTTP paket audit.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register mendaftarkan endpoint audit log.
//
//	GET /api/v1/audit?aksi=&entitas=&actor_id=&limit=&offset=
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q
	_ = cfg

	repo := NewRepository(pool)
	svc := NewService(repo, pool)
	h := NewHandler(svc)

	// GET /api/v1/audit
	e.GET("/api/v1/audit", h.List)
}

// List menangani GET /api/v1/audit.
func (h *Handler) List(c *echo.Context) error {
	if !HasPermission(c, "audit.read") {
		return platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat audit log.")
	}
	// Gaya paginasi audit: default 50, maks 200 — dipertahankan apa adanya
	// (platform.ParseLimit memakai maks 100, tidak cocok di sini).
	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset, err := strconv.Atoi(c.QueryParam("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	items, appErr := h.svc.ListLog(c.Request().Context(),
		c.QueryParam("aksi"), c.QueryParam("entitas"), c.QueryParam("actor_id"),
		limit, offset)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, map[string]any{"items": items, "limit": limit, "offset": offset})
}
