package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
)

// Tes rate limit (Fase 3, BUG-03): 150 request jawaban dalam 1 menit
// dari 1 IP -> tidak ada 429 (anggaran ujian 300/menit/IP).
// Endpoint umum tetap 100/menit/IP, dan 429 mengembalikan Retry-After.
func TestRateLimitUjian150Tanpa429(t *testing.T) {
	e := echo.New()
	e.Use(RateLimit())
	e.POST("/api/v1/ujian/attempt/:id/jawaban", func(c *echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	for i := 0; i < 150; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ujian/attempt/abc/jawaban", nil)
		req.RemoteAddr = "10.0.0.7:1234"
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request jawaban ke-%d: status %d, mau 200 (tidak boleh 429)", i+1, rec.Code)
		}
	}
}

func TestRateLimitUmumTetap100DanRetryAfter(t *testing.T) {
	e := echo.New()
	e.Use(RateLimit())
	e.GET("/api/v1/settings", func(c *echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	var got429 int
	var retryAfter string
	for i := 0; i < 105; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
		req.RemoteAddr = "10.0.0.8:1234"
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code == http.StatusTooManyRequests {
			got429++
			if retryAfter == "" {
				retryAfter = rec.Header().Get("Retry-After")
			}
		} else if rec.Code != http.StatusOK {
			t.Fatalf("request ke-%d: status %d, mau 200/429", i+1, rec.Code)
		}
	}
	if got429 != 5 {
		t.Fatalf("request ke-101..105 harus 429 (5x), dapat %d", got429)
	}
	if retryAfter == "" {
		t.Fatal("respons 429 harus membawa header Retry-After")
	}
}

func TestRateLimitUjianBatas300(t *testing.T) {
	rl := NewRateLimiterWithExamLimits(100, 300, time.Minute)
	e := echo.New()
	e.Use(rl.Middleware())
	e.POST("/api/v1/ujian/attempt/:id/jawaban", func(c *echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	var got429 int
	for i := 0; i < 305; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ujian/attempt/abc/jawaban", nil)
		req.RemoteAddr = "10.0.0.9:1234"
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code == http.StatusTooManyRequests {
			got429++
		}
	}
	if got429 != 5 {
		t.Fatalf("request ke-301..305 harus 429 (5x), dapat %d", got429)
	}
}
