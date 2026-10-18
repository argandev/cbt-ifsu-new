// Package sod — Pengecualian darurat Pemisahan Tugas / Separation of
// Duties (Tim Fase 3).
//
// Aturan SoD (ANALISIS Bab 11.4, 7 aturan): SoD-1 pembuat/edit bank ≠
// pengunci; SoD-2 pembuat soal ≠ pemegang token; SoD-3 pengoreksi esai
// ≠ pengubah nilai final; SoD-4 penguji UKK ≠ finalisator nilai UKK;
// SoD-5 pengawas mengajukan, proktor mengeksekusi; SoD-6 pengelola user
// ≠ pengelola audit log; SoD-7 pengaju draft jadwal ≠ finalisator.
//
// Pengecualian darurat HANYA lewat endpoint di bawah, dengan:
// alasan tertulis (min 20 karakter) + audit dengan flag SOD_OVERRIDE +
// notifikasi semua user role admin kecuali actor. Pengecualian TIDAK
// PERNAH dipakai untuk ubah nilai tanpa jejak / hapus audit log —
// audit_log append-only, tidak ada endpoint ubah/hapus.
//
// Endpoint (Register):
//
//	POST /api/v1/sod/override  (permission sod.override — admin saja)
package sod

import (
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/config"
	"cbt-smk-v2/backend/internal/notifikasi"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// PanjangAlasanMinimum adalah panjang alasan tertulis minimum (karakter).
const PanjangAlasanMinimum = 20

func writeError(c *echo.Context, status int, code, message string) error {
	return c.JSON(status, map[string]any{
		"error": map[string]any{"code": code, "message": message},
	})
}

func writeCreated(c *echo.Context, data any) error {
	return c.JSON(http.StatusCreated, map[string]any{"data": data})
}

// ValidasiAlasan melaporkan pesan error bila alasan < 20 karakter;
// string kosong berarti valid. Dipakai handler + tes.
func ValidasiAlasan(alasan string) string {
	if utf8.RuneCountInString(strings.TrimSpace(alasan)) < PanjangAlasanMinimum {
		return fmt.Sprintf("alasan wajib minimal %d karakter", PanjangAlasanMinimum)
	}
	return ""
}

// DetailAudit membangun detail JSON audit_log untuk satu override:
// berisi aksi + alasan + entitas + flag SOD_OVERRIDE.
func DetailAudit(aksi, entitas, entitasID, alasan string) map[string]any {
	return map[string]any{
		"flag":       "SOD_OVERRIDE",
		"sod_override": true,
		"aksi":       aksi,
		"entitas":    entitas,
		"entitas_id": entitasID,
		"alasan":     alasan,
	}
}

// IsiNotifikasi membangun isi notifikasi untuk admin lain.
func IsiNotifikasi(aksi, entitas, entitasID, alasan, namaActor string) string {
	return fmt.Sprintf("Pengecualian SoD oleh %s: %s pada %s (%s). Alasan: %s",
		namaActor, aksi, entitas, entitasID, alasan)
}

type overrideInput struct {
	Aksi      string `json:"aksi"`
	Entitas   string `json:"entitas"`
	EntitasID string `json:"entitas_id"`
	Alasan    string `json:"alasan"`
}

// Register mendaftarkan endpoint SoD override.
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q
	_ = cfg

	e.POST("/api/v1/sod/override", func(c *echo.Context) error {
		userID, _, _, _, ok := audit.CurrentUser(c)
		if !ok {
			return writeError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
		}
		if !audit.HasPermission(c, "sod.override") {
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mencatat pengecualian SoD.")
		}
		var in overrideInput
		if err := c.Bind(&in); err != nil {
			return writeError(c, http.StatusBadRequest, "validation_error", "Body JSON tidak valid.")
		}
		in.Aksi = strings.TrimSpace(in.Aksi)
		in.Entitas = strings.TrimSpace(in.Entitas)
		in.EntitasID = strings.TrimSpace(in.EntitasID)
		if in.Aksi == "" || in.Entitas == "" {
			return writeError(c, http.StatusBadRequest, "validation_error", "aksi dan entitas wajib diisi.")
		}
		if msg := ValidasiAlasan(in.Alasan); msg != "" {
			return writeError(c, http.StatusBadRequest, "validation_error", msg)
		}
		ctx := c.Request().Context()
		var id string
		if err := pool.QueryRow(ctx,
			`INSERT INTO sod_overrides (actor_id, aksi, entitas, entitas_id, alasan)
			 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			userID, in.Aksi, in.Entitas, in.EntitasID, strings.TrimSpace(in.Alasan)).Scan(&id); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mencatat pengecualian SoD.")
		}
		// Audit dengan flag SOD_OVERRIDE (detail JSON berisi aksi +
		// alasan + entitas). Gagal catat tidak menggagalkan override
		// yang sudah tersimpan, tapi dicatat ke log server.
		if err := audit.Catat(ctx, pool, userID, "sod.override", "sod_override", id,
			DetailAudit(in.Aksi, in.Entitas, in.EntitasID, strings.TrimSpace(in.Alasan)), c.RealIP()); err != nil {
			c.Logger().Error("sod: gagal mencatat audit SOD_OVERRIDE", "error", err)
		}
		// Notifikasi untuk SEMUA user role admin kecuali actor.
		var namaActor string
		_ = pool.QueryRow(ctx, `SELECT COALESCE(nama, '') FROM users WHERE id = $1`, userID).Scan(&namaActor)
		if namaActor == "" {
			namaActor = userID
		}
		rows, err := pool.Query(ctx,
			`SELECT id FROM users WHERE role = 'admin' AND id <> $1 AND is_active = TRUE`, userID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var adminID string
				if err := rows.Scan(&adminID); err != nil {
					continue
				}
				_ = notifikasi.Buat(ctx, pool, adminID, "Pengecualian SoD",
					IsiNotifikasi(in.Aksi, in.Entitas, in.EntitasID, strings.TrimSpace(in.Alasan), namaActor))
			}
		}
		return writeCreated(c, map[string]any{
			"id": id, "actor_id": userID, "aksi": in.Aksi, "entitas": in.Entitas,
			"entitas_id": in.EntitasID, "alasan": strings.TrimSpace(in.Alasan),
		})
	})
}
