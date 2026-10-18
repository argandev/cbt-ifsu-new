// Package settings — pengaturan sekolah (key/value JSONB).
//
// Tabel settings dibuat di migrasi 0001 (Tim 1): settings(key PK, value jsonb, updated_at).
// Daftar key yang dikenal + default ada di DefaultSettings. Key tak dikenal
// ditolak pada PUT agar tidak ada sampah konfigurasi.
// remedial_mode HANYA "tertinggi" (nilai remedial vs awal dipakai yang tertinggi).
package settings

import (
	"encoding/json"
	"net/http"
	"strings"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// DefaultSettings adalah nilai default tiap key (dipakai bila baris belum ada).
func DefaultSettings() map[string]any {
	return map[string]any{
		"tampil_hasil_ke_siswa": false,
		"token_expiry_detik":    900,
		"toleransi_telat_menit": 15,
		"max_pelanggaran":       2,
		"durasi_default":        90,
		"nama_sekolah":          "",
		"logo_url":              "",
		"kktp_per_mapel":        map[string]any{},
		"remedial_mode":         "tertinggi",
	}
}

// validasi per key; mengembalikan nilai yang sudah dinormalisasi + ok.
func validateSetting(key string, raw json.RawMessage) (any, bool) {
	switch key {
	case "tampil_hasil_ke_siswa":
		var v bool
		if json.Unmarshal(raw, &v) != nil {
			return nil, false
		}
		return v, true
	case "token_expiry_detik", "toleransi_telat_menit", "max_pelanggaran", "durasi_default":
		var v float64
		if json.Unmarshal(raw, &v) != nil || v < 0 || v != float64(int(v)) {
			return nil, false
		}
		// Batas wajar: token expiry 60-86400 detik, toleransi 0-120 menit,
		// max pelanggaran 1-10, durasi 5-600 menit.
		switch key {
		case "token_expiry_detik":
			if v < 60 || v > 86400 {
				return nil, false
			}
		case "toleransi_telat_menit":
			if v > 120 {
				return nil, false
			}
		case "max_pelanggaran":
			if v < 1 || v > 10 {
				return nil, false
			}
		case "durasi_default":
			if v < 5 || v > 600 {
				return nil, false
			}
		}
		return int(v), true
	case "nama_sekolah", "logo_url":
		var v string
		if json.Unmarshal(raw, &v) != nil {
			return nil, false
		}
		return v, true
	case "kktp_per_mapel":
		// KKTP per mapel/TP: objek { "<kode_mapel>": {deskripsi, rubrik, interval, ...} }.
		// BUKAN hardcode KKM=75 — tiap mapel/TP punya ambang & deskripsinya sendiri.
		var v map[string]any
		if json.Unmarshal(raw, &v) != nil {
			return nil, false
		}
		return v, true
	case "remedial_mode":
		var v string
		if json.Unmarshal(raw, &v) != nil || v != "tertinggi" {
			return nil, false
		}
		return v, true
	}
	return nil, false
}

func writeError(c *echo.Context, status int, code, message string) error {
	return c.JSON(status, map[string]any{
		"error": map[string]any{"code": code, "message": message},
	})
}

func writeOK(c *echo.Context, data any) error {
	return c.JSON(http.StatusOK, map[string]any{"data": data})
}

// Register mendaftarkan endpoint settings.
//
//	GET /api/v1/settings
//	PUT /api/v1/settings
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q
	_ = cfg

	e.GET("/api/v1/settings", func(c *echo.Context) error {
		if !audit.HasPermission(c, "settings.read") {
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat settings.")
		}
		ctx := c.Request().Context()
		rows, err := pool.Query(ctx, `SELECT key, value FROM settings ORDER BY key ASC`)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca settings.")
		}
		defer rows.Close()
		result := DefaultSettings()
		for rows.Next() {
			var key string
			var raw json.RawMessage
			if err := rows.Scan(&key, &raw); err != nil {
				return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca settings.")
			}
			var v any
			if err := json.Unmarshal(raw, &v); err == nil {
				result[key] = v
			}
		}
		if err := rows.Err(); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca settings.")
		}
		return writeOK(c, result)
	})

	e.PUT("/api/v1/settings", func(c *echo.Context) error {
		if !audit.HasPermission(c, "settings.update") {
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mengubah settings.")
		}
		// Terima body {"settings": {...}} maupun flat {"key": value}.
		var generic map[string]json.RawMessage
		if err := c.Bind(&generic); err != nil {
			return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
		}
		updates := map[string]json.RawMessage{}
		if rawSettings, hasWrapper := generic["settings"]; hasWrapper {
			if err := json.Unmarshal(rawSettings, &updates); err != nil {
				return writeError(c, http.StatusBadRequest, "validation_error", "Field settings tidak valid.")
			}
		} else {
			updates = generic
		}
		body := struct {
			Settings map[string]json.RawMessage `json:"settings"`
		}{Settings: updates}
		if len(body.Settings) == 0 {
			return writeError(c, http.StatusBadRequest, "validation_error", "Tidak ada setting yang dikirim.")
		}
		ctx := c.Request().Context()
		tx, err := pool.Begin(ctx)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memulai penyimpanan settings.")
		}
		defer tx.Rollback(ctx)
		updated := map[string]any{}
		for key, raw := range body.Settings {
			key = strings.TrimSpace(key)
			if _, known := DefaultSettings()[key]; !known {
				return writeError(c, http.StatusBadRequest, "validation_error", "Key tidak dikenal: "+key)
			}
			normalized, ok := validateSetting(key, raw)
			if !ok {
				return writeError(c, http.StatusBadRequest, "validation_error", "Nilai tidak valid untuk key: "+key)
			}
			valueJSON, _ := json.Marshal(normalized)
			if _, err := tx.Exec(ctx,
				`INSERT INTO settings (key, value, updated_at) VALUES ($1, $2, now())
				 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()`,
				key, valueJSON); err != nil {
				return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menyimpan settings.")
			}
			updated[key] = normalized
		}
		if err := tx.Commit(ctx); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menyimpan settings.")
		}
		userID, _, _, _, _ := audit.CurrentUser(c)
		_ = audit.Catat(ctx, pool, userID, "settings.update", "settings", "",
			map[string]any{"updated": updated}, c.RealIP())
		return writeOK(c, map[string]any{"updated": updated})
	})
}
