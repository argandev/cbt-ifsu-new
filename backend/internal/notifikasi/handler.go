package notifikasi

import (
	"net/http"
	"strconv"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/config"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// --- Handler ----------------------------------------------------------------

// Handler adalah lapisan HTTP paket notifikasi.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// actorFrom membangun RequestActor dari echo.Context.
func actorFrom(c *echo.Context) RequestActor {
	userID, _, _, _, _ := audit.CurrentUser(c)
	return RequestActor{UserID: userID, RealIP: c.RealIP()}
}

// Register mendaftarkan endpoint notifikasi.
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q
	_ = cfg

	repo := NewRepository(pool)
	svc := NewService(repo, pool)
	h := NewHandler(svc)

	// GET /api/v1/notifikasi — milik sendiri, pagination limit/offset.
	e.GET("/api/v1/notifikasi", h.List)

	// POST /api/v1/notifikasi/:id/baca — tandai dibaca; bukan milik -> 404.
	e.POST("/api/v1/notifikasi/:id/baca", h.TandaiBaca)
}

// List menangani GET /api/v1/notifikasi.
func (h *Handler) List(c *echo.Context) error {
	userID, _, _, _, ok := audit.CurrentUser(c)
	if !ok {
		return platform.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
	}
	// Gaya paginasi notifikasi: default 50, maks 200 — dipertahankan apa
	// adanya (platform.ParseLimit memakai maks 100, tidak cocok di sini).
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

	items, appErr := h.svc.List(c.Request().Context(), userID, limit, offset)
	if appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, map[string]any{"items": items, "limit": limit, "offset": offset})
}

// TandaiBaca menangani POST /api/v1/notifikasi/:id/baca.
func (h *Handler) TandaiBaca(c *echo.Context) error {
	if _, _, _, _, ok := audit.CurrentUser(c); !ok {
		return platform.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
	}
	id := c.Param("id")
	if appErr := h.svc.TandaiBaca(c.Request().Context(), actorFrom(c), id); appErr != nil {
		return platform.WriteAppError(c, appErr)
	}
	return platform.WriteOK(c, map[string]any{"dibaca": true, "id": id})
}
