// Package event — Event + Ruangan + Kartu Peserta (Tim Fase 2).
//
// Event = pengelompokan ujian (mis. "STS Ganjil 2026/2027", "SAS Genap",
// "USBK 2027", "UKK 2027"). Jadwal ujian dapat merujuk ke satu event
// (opsional). Ruangan = ruang fisik ujian (nama, kapasitas, lokasi,
// keterangan). Jadwal dapat ditugaskan ke satu ruangan.
//
// Kartu peserta = SATU kartu per siswa untuk SEMUA ujian (model 1-KARTU):
// identitas siswa sekali + tabel SEMUA jadwal kelasnya. Nomor peserta unik PER
// JADWAL (siswa yang sama di jadwal berbeda boleh bernomor berbeda),
// format: <kode-event-atau-US>/<tahun tanggal jadwal>/<nomor urut 3 digit>
// (contoh: "STS-GANJIL-26/2026/001", tanpa event: "US/2026/001").
// Nomor urut dihitung dari urutan nama siswa A-Z. Kartu TIDAK berisi
// token ujian — token hanya dari proktor saat ujian.
//
// Izin (sesuai seed + kontrak §5):
//   - Event/ruangan CRUD: event.read (GET), event.create/update/delete
//     (mutasi). Guru TIDAK punya event.* di matriks kanonis, jadi guru
//     mendapat 403 pada CRUD event/ruangan — itu sesuai seed.
//   - Kartu (daftar + satu): staf admin/proktor/guru. Karena guru tidak
//     punya event.read di seed, kartu menerima event.read ATAU
//     rekap.read untuk staf — pengawas/penguji_ukk yang juga punya
//     rekap.read tetap ditolak karena gate role staf di atas.
//   - Siswa hanya boleh melihat kartunya sendiri (satu kartu); daftar
//     kartu bukan haknya (anti-IDOR: 404, bukan 403).
//
// Struktur layer (hasil refactor arsitektur):
//   - handler.go:        Echo handler + Register (file ini) + guard HTTP
//   - handler_event.go:  handler CRUD event/ruangan + tugaskan ke jadwal
//   - handler_kartu.go:  handler kartu peserta (JSON legacy)
//   - service.go:        Service + platform.RequestActor + catat
//   - service_event.go:  logika bisnis event/ruangan + validasi input
//   - service_kartu.go:  logika bisnis kartu + helper murni kartu
//   - repository_event.go: DTO eventRow/ruanganRow + semua query SQL event
//   - repository_kartu.go: DTO kartu + loader kartu (dipakai kartu_pdf.go)
//   - kartu_pdf.go:      generator PDF kartu (tidak diubah strukturnya;
//     hanya helper writeError -> platform + helper murni pindah ke service)
//
// Perilaku (route, status, bentuk JSON, pesan error) IDENTIK dengan
// sebelum refactor.
package event

import (
	"net/http"
	"strings"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/config"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// Handler adalah lapisan HTTP paket event.
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

// Register mendaftarkan endpoint event + ruangan + kartu peserta.
//
//	GET    /api/v1/event                        (event.read)
//	POST   /api/v1/event                        (event.create)
//	GET    /api/v1/event/:id                    (event.read)
//	PUT    /api/v1/event/:id                    (event.update)
//	DELETE /api/v1/event/:id                    (event.delete)
//	GET    /api/v1/ruangan                      (event.read)
//	POST   /api/v1/ruangan                      (event.create)
//	GET    /api/v1/ruangan/:id                  (event.read)
//	PUT    /api/v1/ruangan/:id                  (event.update)
//	DELETE /api/v1/ruangan/:id                  (event.delete)
//	PUT    /api/v1/jadwal/:id/event-ruangan     (event.update)
//	GET    /api/v1/kartu/pdf                  (staf admin/proktor/guru, PDF gabungan)
//	GET    /api/v1/kartu/siswa/:siswaId/pdf  (staf / siswa pemilik kartu, PDF 1 kartu)
//	GET    /api/v1/kartu/:jadwalId              (staf admin/proktor/guru, JSON legacy)
//	GET    /api/v1/kartu/:jadwalId/:siswaId    (staf / siswa pemilik kartu, JSON legacy)
//
// Aturan kunci:
//   - Hapus event/ruangan -> 409 bila masih dipakai jadwal ujian.
//   - Kode event/ruangan unik -> 409 bila duplikat.
//   - Tugaskan ruangan ke jadwal -> 409 bila kapasitas ruangan < jumlah
//     siswa aktif kelas jadwal.
//   - Kartu TIDAK berisi token ujian.
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q
	_ = cfg

	repo := NewRepository(pool)
	svc := NewService(repo, pool)
	h := NewHandler(svc)

	// ---- Event ----

	e.GET("/api/v1/event", h.ListEvent)
	e.POST("/api/v1/event", h.CreateEvent)
	e.GET("/api/v1/event/:id", h.GetEvent)
	e.PUT("/api/v1/event/:id", h.UpdateEvent)
	e.DELETE("/api/v1/event/:id", h.DeleteEvent)

	// ---- Ruangan ----

	e.GET("/api/v1/ruangan", h.ListRuangan)
	e.POST("/api/v1/ruangan", h.CreateRuangan)
	e.GET("/api/v1/ruangan/:id", h.GetRuangan)
	e.PUT("/api/v1/ruangan/:id", h.UpdateRuangan)
	e.DELETE("/api/v1/ruangan/:id", h.DeleteRuangan)

	// ---- Tugaskan event + ruangan ke jadwal ----

	e.PUT("/api/v1/jadwal/:id/event-ruangan", h.AssignEventRuangan)

	// ---- Kartu peserta PDF (server-side, generator di kartu_pdf.go) ----
	// Route PDF didaftarkan SEBELUM route JSON legacy agar segmen statis
	// "pdf" / "siswa" menang atas parameter :jadwalId.

	e.GET("/api/v1/kartu/pdf", handleKartuBulkPDF(pool))
	e.GET("/api/v1/kartu/siswa/:siswaId/pdf", handleKartuSiswaPDF(pool))

	// ---- Kartu peserta (JSON, LEGACY — tetap dipertahankan) ----

	e.GET("/api/v1/kartu/:jadwalId", h.KartuJadwal)
	e.GET("/api/v1/kartu/:jadwalId/:siswaId", h.KartuSiswa)
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

// requireKartuStaf: gate kartu untuk staf admin/proktor/guru (lihat
// komentar paket). Role lain -> 404 (anti-IDOR: resource bukan haknya).
func requireKartuStaf(c *echo.Context) (userID, role string, ok bool) {
	uid, r, _, _, ok := audit.CurrentUser(c)
	if !ok {
		_ = platform.WriteError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
		return "", "", false
	}
	if r != "admin" && r != "proktor" && r != "guru" {
		_ = platform.WriteError(c, http.StatusNotFound, "not_found", "Kartu peserta tidak ditemukan.")
		return "", "", false
	}
	if !audit.HasPermission(c, "event.read") && !audit.HasPermission(c, "rekap.read") {
		_ = platform.WriteError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat kartu peserta.")
		return "", "", false
	}
	return uid, r, true
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

// writeItemsList menulis daftar gaya limit/offset khas event:
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
