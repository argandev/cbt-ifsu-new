package banksoal

import (
	"net/http"
	"strconv"
	"strings"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/config"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// Handler adalah lapisan HTTP paket banksoal.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register mendaftarkan endpoint bank soal (Tim 3).
//
//	GET    /api/v1/bank-soal                        (banksoal.read)
//	POST   /api/v1/bank-soal                        (banksoal.create)
//	GET    /api/v1/bank-soal/:id                    (banksoal.read)
//	PUT    /api/v1/bank-soal/:id                    (banksoal.update)
//	DELETE /api/v1/bank-soal/:id                    (banksoal.delete)
//	POST   /api/v1/bank-soal/:id/lock               (banksoal.lock)
//	POST   /api/v1/bank-soal/:id/unlock             (banksoal.lock)
//	POST   /api/v1/bank-soal/import                 (banksoal.import)
//	GET    /api/v1/bank-soal/:bankId/soal           (banksoal.read, guru/admin saja)
//	POST   /api/v1/bank-soal/:bankId/soal           (banksoal.create)
//	GET    /api/v1/bank-soal/:bankId/soal/:soalId   (banksoal.read, guru/admin saja)
//	PUT    /api/v1/bank-soal/:bankId/soal/:soalId   (banksoal.update)
//	DELETE /api/v1/bank-soal/:bankId/soal/:soalId   (banksoal.delete)
//
// Aturan kunci:
//   - Bank terkunci -> mutasi soal (tambah/ubah/hapus/import) = 409.
//   - Hapus bank -> 409 bila terkunci atau dipakai paket.
//   - Kunci jawaban hanya untuk guru/admin; proktor tidak melihat isi
//     soal; pengawas/siswa/penguji UKK tidak punya akses (403).
//
// Struktur layer (hasil refactor arsitektur):
//   - handler.go:        Echo handler + Register (file ini) + guard HTTP
//   - handler_bank.go:   handler CRUD bank + lock/unlock
//   - handler_soal.go:   handler CRUD soal + import CSV
//   - service.go:        Service + RequestActor + catat
//   - service_bank.go:   logika bisnis bank (+ guard scope guru)
//   - service_soal.go:   logika bisnis soal + import CSV
//   - repository_bank.go: DTO bankRow + semua query SQL bank
//   - repository_soal.go: DTO soalRow + semua query SQL soal
//   - validation.go:     validasi soalInput + helper import CSV
//   - types.go:          tipe soal + ValidTipe/ValidateSoal (tidak diubah)
//   - import_excel.go / import_word.go: jalur import xlsx/docx (struktur
//     tidak diubah; hanya helper writeError/writeOK -> platform)
//
// Perilaku (route, status, bentuk JSON, pesan error) IDENTIK dengan
// sebelum refactor.
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q
	_ = cfg

	repo := NewRepository(pool)
	svc := NewService(repo, pool)
	h := NewHandler(svc)

	// ---- Bank soal ----

	e.GET("/api/v1/bank-soal", h.ListBank)
	e.POST("/api/v1/bank-soal", h.CreateBank)
	e.GET("/api/v1/bank-soal/:id", h.GetBank)
	e.PUT("/api/v1/bank-soal/:id", h.UpdateBank)
	e.DELETE("/api/v1/bank-soal/:id", h.DeleteBank)
	e.POST("/api/v1/bank-soal/:id/lock", h.LockBank)
	e.POST("/api/v1/bank-soal/:id/unlock", h.UnlockBank)

	// ---- Import CSV ----
	// POST /api/v1/bank-soal/import (didaftarkan sebelum rute :id agar
	// segmen statis "import" tidak tertukar dengan parameter :id).
	e.POST("/api/v1/bank-soal/import", h.ImportCSV)

	// ---- Soal ----

	e.GET("/api/v1/bank-soal/:bankId/soal", h.ListSoal)
	e.POST("/api/v1/bank-soal/:bankId/soal", h.CreateSoal)
	e.GET("/api/v1/bank-soal/:bankId/soal/:soalId", h.GetSoal)
	e.PUT("/api/v1/bank-soal/:bankId/soal/:soalId", h.UpdateSoal)
	e.DELETE("/api/v1/bank-soal/:bankId/soal/:soalId", h.DeleteSoal)
}

// requirePerm: 401 bila belum login (context auth Tim 1 kosong),
// 403 bila tidak punya permission. Role admin lolos semua permission
// (lihat audit.HasPermission).
func requirePerm(c *echo.Context, perm string) (userID, role string, ok bool) {
	uid, r, _, _, ok := audit.CurrentUser(c)
	if !ok {
		_ = platform.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
		return "", "", false
	}
	if !audit.HasPermission(c, perm) {
		_ = platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin: "+perm+".")
		return "", "", false
	}
	return uid, r, true
}

// forbidSoalUntukProktor: proktor & pengawas TIDAK boleh melihat isi
// soal (README + seed: proktor hanya punya banksoal.read untuk metadata
// bank; pengawas tidak punya permission banksoal sama sekali).
// Dipanggil di SEMUA endpoint soal (list/get/create/update/delete/import).
func forbidSoalUntukProktor(c *echo.Context, role string) bool {
	if role == "proktor" {
		_ = platform.WriteError(c, http.StatusForbidden, "forbidden", "Proktor tidak boleh melihat isi soal.")
		return false
	}
	return true
}

func isValidUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, ch := range s {
		switch i {
		case 8, 13, 18, 23:
			if ch != '-' {
				return false
			}
		default:
			if !(ch >= '0' && ch <= '9' || ch >= 'a' && ch <= 'f' || ch >= 'A' && ch <= 'F') {
				return false
			}
		}
	}
	return true
}

func strPtr(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// checkGuruScope: guru hanya boleh memutasi bank miliknya (guru_id =
// dirinya) atau bank tanpa pemilik. Admin bebas.
// (Dipakai jalur import xlsx/docx; endpoint CRUD memakai versi service
// yang mengembalikan *platform.AppError.)
func checkGuruScope(c *echo.Context, role, userID string, b *bankRow) bool {
	if !guruScopeOK(role, userID, b) {
		_ = platform.WriteError(c, http.StatusForbidden, "forbidden", "Bank soal ini milik guru lain.")
		return false
	}
	return true
}

// writeItemsList menulis daftar gaya limit/offset khas banksoal:
//
//	{"data": {"items": ..., "total": ..., "limit": ..., "offset": ...}}
//
// Bentuk ini dipertahankan apa adanya (berbeda dari
// platform.WriteLimitList) agar kontrak JSON tidak berubah.
func writeItemsList(c *echo.Context, items any, total, limit, offset int) error {
	return c.JSON(http.StatusOK, map[string]any{
		"data": map[string]any{"items": items, "total": total, "limit": limit, "offset": offset},
	})
}

// bankFilterArgs merakit argumen filter query bank dari parameter HTTP.
func bankFilterArgs(c *echo.Context) BankFilter {
	var kelas *int
	if v := c.QueryParam("kelas_tingkat"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			kelas = &n
		}
	}
	var locked *bool
	if v := c.QueryParam("is_locked"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			locked = &b
		}
	}
	var kelasArg, lockedArg any
	if kelas != nil {
		kelasArg = *kelas
	}
	if locked != nil {
		lockedArg = *locked
	}
	return BankFilter{
		Search:       strPtr(c.QueryParam("q")),
		MapelID:      strPtr(c.QueryParam("mapel_id")),
		KelasTingkat: kelasArg,
		Locked:       lockedArg,
	}
}

// soalFilterArgs merakit argumen filter query soal dari parameter HTTP.
func soalFilterArgs(c *echo.Context) SoalFilter {
	return SoalFilter{
		Tipe:   strPtr(c.QueryParam("tipe")),
		Level:  strPtr(c.QueryParam("level")),
		Kisi:   strPtr(c.QueryParam("kisi_kisi_id")),
		Search: strPtr(c.QueryParam("q")),
	}
}
