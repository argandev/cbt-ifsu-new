package platform

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
)

// Aplikasi ini memakai DUA gaya paginasi (keduanya dipertahankan apa adanya
// agar bentuk JSON response tidak berubah):
//
//  1. Gaya page/per_page — paket master, users.
//     Query: ?page=1&per_page=20 ; meta: {page, per_page, total, total_pages}
//  2. Gaya limit/offset — paket banksoal, event, ukk.
//     Query: ?limit=20&offset=0 ; meta: {limit, offset, total}

// PageMeta — bentuk meta untuk gaya page/per_page. IDENTIK dengan
// pageMeta yang sebelumnya didefinisikan di paket master & users.
type PageMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// WritePageList menulis daftar gaya page/per_page — IDENTIK dengan
// writeList di paket master & users.
func WritePageList(c *echo.Context, items any, page, perPage int, total int64) error {
	totalPages := 0
	if perPage > 0 {
		totalPages = int((total + int64(perPage) - 1) / int64(perPage))
	}
	return c.JSON(http.StatusOK, map[string]any{
		"data": items,
		"meta": PageMeta{Page: page, PerPage: perPage, Total: total, TotalPages: totalPages},
	})
}

// ParsePage mem-parsing ?page=&per_page= — IDENTIK dengan parsePagination
// di paket master & users.
func ParsePage(c *echo.Context) (page, perPage, offset int) {
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page < 1 {
		page = 1
	}
	perPage, err = strconv.Atoi(c.QueryParam("per_page"))
	if err != nil || perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	return page, perPage, (page - 1) * perPage
}

// WriteLimitList menulis daftar gaya limit/offset — IDENTIK dengan
// writeList di paket ukk.
func WriteLimitList(c *echo.Context, items any, limit, offset int, total int64) error {
	return c.JSON(http.StatusOK, map[string]any{
		"data": items,
		"meta": map[string]any{"limit": limit, "offset": offset, "total": total},
	})
}

// ParseLimit mem-parsing ?limit=&offset= — IDENTIK dengan parsePagination
// di paket banksoal, event (default 20), dan ukk (default 50).
// defaultLimit diisi sesuai paket pemanggil agar perilaku tidak berubah.
func ParseLimit(c *echo.Context, defaultLimit int) (limit, offset int) {
	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || limit <= 0 {
		limit = defaultLimit
	}
	if limit > 100 {
		limit = 100
	}
	offset, err = strconv.Atoi(c.QueryParam("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}
	return limit, offset
}
