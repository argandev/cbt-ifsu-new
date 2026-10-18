// Package exam — paket, jadwal, token, attempt, grading trigger, koreksi.
//
// Kontrak (tugas Tim 4): Register(e *echo.Echo, pool *pgxpool.Pool,
// q *db.Queries, cfg *config.Config). Repository memakai pgxpool langsung
// dengan SQL yang sama seperti backend/db/queries/exam.sql (pola Tim 5);
// q diterima agar signature sesuai kontrak bersama.
//
// Aturan ujian FINAL yang ditegakkan di sini (jangan diubah diam-diam):
//   - deadline attempt = jadwal_selesai (tanggal + jam_selesai).
//   - Pause MEMBEKUKAN timer (sisa_waktu_beku); resume: deadline = now + sisa.
//   - Unblock = LANJUT sisa waktu, jawaban tersimpan, TIDAK reset;
//     hanya Proktor/Admin eksekusi; Pengawas hanya mengusulkan; wajib alasan.
//   - Token kedua = token pertama HANGUS. Token expiry default 900 dtk
//     dari WAKTU GENERATE (settings.token_expiry_detik) — token baru
//     selalu berlaku 15 menit sejak dibuat.
//   - Toleransi telat default 15 menit (settings.toleransi_telat_menit):
//     siswa boleh MULAI hingga mulai+15 menit; sisa = jadwal_selesai - masuk.
//   - Tryout boleh tanpa token + pembahasan langsung.
//   - Nilai esai disembunyikan sampai koreksi final: nilai tampil ke siswa
//     hanya bila paket.tampil_hasil AND settings.tampil_hasil_ke_siswa.
//   - KUNCI JAWABAN TIDAK PERNAH dikirim ke browser (kueri soal ujian
//     tidak memilih kolom kunci/pembahasan).
//   - Timer berjangkar server: server menolak save/submit setelah deadline.
//   - Anti-curang: pelanggaran 1x = peringatan, 2x = auto-block PER-ATTEMPT;
//     deduplikasi event berulang dalam 5 detik.
//   - Nilai akhir = Σ poin / Σ bobot × 100 (grading.HitungNilaiAkhir).
//
// Permission (matriks ANALISIS Bab 11 + tugas Tim 4; ditegakkan via ROLE):
//   - paket: admin, guru. Proktor TIDAK (tidak boleh lihat isi soal).
//   - jadwal CRUD: admin, proktor, guru (guru "lihat/buat jika diizinkan").
//     Pengawas read-only. Token: admin/proktor saja (pengawas TIDAK
//     boleh melihat token — termasuk token-status).
//   - pause/resume/unblock: admin/proktor. Usulan-unblock: pengawas
//     (proktor/admin boleh mengusulkan juga).
//   - ujian.*: siswa, hanya attempt milik sendiri.
//   - koreksi: admin/guru. Monitor: lihat paket monitor.
//
// CATATAN KONFLIK SEED (Tim 6): seed memberi proktor paket.* dan TIDAK
// memberi pengawas jadwal.usulan-unblock — keduanya bertentangan dengan
// matriks ANALISIS Bab 11. Handler ini menegakkan MATRIKS (role), bukan
// matriks seed. Bila koordinator memutuskan seed yang benar, ubah di sini.
//
// Struktur layer (hasil refactor arsitektur):
//   - handler.go:          Echo handler + Register + guard role (file ini)
//   - handler_paket.go:    endpoint paket + susunan soal
//   - handler_jadwal.go:   endpoint jadwal + token + pause/resume
//   - handler_usulan.go:   endpoint usulan unblock + unblock
//   - handler_ujian.go:    endpoint ujian siswa
//   - handler_koreksi.go:  endpoint koreksi
//   - service*.go:         validasi + logika bisnis + audit + broadcast
//   - repository*.go:      DTO + semua query SQL (verbatim)
//   - platform:            envelope JSON + AppError (menggantikan helper lokal)
//
// Perilaku (route, status, bentuk JSON, pesan error) IDENTIK dengan
// sebelum refactor.
package exam

import (
	"net/http"
	"strings"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/config"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// Handler adalah lapisan HTTP paket exam.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// actorFrom membangun RequestActor dari echo.Context.
func actorFrom(c *echo.Context) RequestActor {
	uid, _, _, _ := currentUser(c)
	return RequestActor{UserID: uid, RealIP: c.RealIP()}
}

// ---------------------------------------------------------------------------
// Helpers umum
// ---------------------------------------------------------------------------
// (Dipindah dari exam.go tanpa perubahan perilaku; writeError/writeOK/
// writeCreated lokal diganti platform.* yang terbukti identik.)

// Kunci context yang diisi middleware auth Tim 1 (sama seperti yang
// dibaca paket audit Tim 5): auth_user_id, auth_role, auth_siswa_id,
// auth_permissions.
const (
	ctxUserID  = "auth_user_id"
	ctxRole    = "auth_role"
	ctxSiswaID = "auth_siswa_id"
)

// currentUser membaca identitas dari context (middleware auth Tim 1).
// Mengembalikan userID, role, siswaID (default = userID), ok.
// TIDAK memakai paket audit Tim 5 di sini: audit.go masih memakai
// signature Echo v4 (echo.Context) yang tidak compile di Echo v5
// (*echo.Context). Bila Tim 5 sudah memperbaikinya, fungsi ini boleh
// dialihkan ke audit.CurrentUser.
func currentUser(c *echo.Context) (string, string, string, bool) {
	uid, _ := c.Get(ctxUserID).(string)
	role, _ := c.Get(ctxRole).(string)
	siswaID, _ := c.Get(ctxSiswaID).(string)
	if siswaID == "" {
		siswaID = uid
	}
	if uid == "" {
		return "", "", "", false
	}
	return uid, role, siswaID, true
}

func requireLogin(c *echo.Context) (uid, role, siswaID string, ok bool) {
	uid, role, siswaID, ok = currentUser(c)
	if !ok {
		_ = platform.WriteError(c, http.StatusUnauthorized, "unauthorized", "Harus login.")
		return "", "", "", false
	}
	return uid, role, siswaID, true
}

// requireRole menolak bila role pemanggil tidak ada di daftar.
func requireRole(c *echo.Context, roles ...string) (uid, role, siswaID string, ok bool) {
	uid, role, siswaID, ok = requireLogin(c)
	if !ok {
		return "", "", "", false
	}
	for _, r := range roles {
		if role == r {
			return uid, role, siswaID, true
		}
	}
	_ = platform.WriteError(c, http.StatusForbidden, "forbidden", "Role "+role+" tidak punya izin untuk aksi ini.")
	return "", "", "", false
}

func isUUID(s string) bool {
	return len(s) == 36 && strings.Count(s, "-") == 4
}

// ---------------------------------------------------------------------------
// Register
// ---------------------------------------------------------------------------

// Register mendaftarkan seluruh endpoint Tim 4.
// q/cfg diterima sesuai kontrak.
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q
	_ = cfg

	repo := NewRepository(pool)
	svc := NewService(repo, pool)
	h := NewHandler(svc)

	// Paket.
	e.GET("/api/v1/paket", h.ListPaket)
	e.POST("/api/v1/paket", h.CreatePaket)
	e.GET("/api/v1/paket/:id", h.GetPaket)
	e.PUT("/api/v1/paket/:id", h.UpdatePaket)
	e.DELETE("/api/v1/paket/:id", h.DeletePaket)
	e.GET("/api/v1/paket/:id/soal", h.ListPaketSoal)
	e.POST("/api/v1/paket/:id/soal", h.AddPaketSoal)
	e.DELETE("/api/v1/paket/:id/soal", h.DeletePaketSoal)
	e.POST("/api/v1/paket/:id/generate-acak", h.GenerateAcak)

	// Jadwal + token + pause/resume + unblock.
	e.GET("/api/v1/jadwal", h.ListJadwal)
	e.POST("/api/v1/jadwal", h.CreateJadwal)
	e.GET("/api/v1/jadwal/:id", h.GetJadwalDetail)
	e.PUT("/api/v1/jadwal/:id", h.UpdateJadwal)
	e.DELETE("/api/v1/jadwal/:id", h.DeleteJadwal)
	e.POST("/api/v1/jadwal/:id/pause", h.PauseJadwal)
	e.POST("/api/v1/jadwal/:id/resume", h.ResumeJadwal)
	e.POST("/api/v1/jadwal/:id/token", h.GenerateToken)
	e.GET("/api/v1/jadwal/:id/token-status", h.TokenStatus)
	e.GET("/api/v1/jadwal/:id/usulan-unblock", h.ListUsulanUnblock)
	e.POST("/api/v1/jadwal/:id/usulan-unblock", h.CreateUsulanUnblock)
	e.POST("/api/v1/jadwal/:id/unblock", h.UnblockAttempt)

	// Ujian siswa.
	e.GET("/api/v1/ujian/aktif", h.UjianAktif)
	e.POST("/api/v1/ujian/mulai", h.MulaiUjian)
	e.GET("/api/v1/ujian/attempt/:id", h.GetAttemptDetail)
	e.POST("/api/v1/ujian/attempt/:id/jawaban", h.SimpanJawaban)
	e.POST("/api/v1/ujian/attempt/:id/ragu", h.TandaiRagu)
	e.POST("/api/v1/ujian/attempt/:id/pelanggaran", h.CatatPelanggaran)
	e.POST("/api/v1/ujian/attempt/:id/submit", h.SubmitUjian)
	e.GET("/api/v1/ujian/attempt/:id/status", h.StatusAttempt)
	e.GET("/api/v1/ujian/attempt/:id/hasil", h.HasilAttempt)

	// Koreksi.
	e.GET("/api/v1/koreksi/:jadwalId", h.ListKoreksi)
	e.POST("/api/v1/koreksi/:jawabanId", h.KoreksiJawaban)
}
