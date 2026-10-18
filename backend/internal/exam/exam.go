// Package exam — paket, jadwal, token, attempt, grading trigger, koreksi.
//
// Kontrak (tugas Tim 4): Register(e *echo.Echo, pool *pgxpool.Pool,
// q *db.Queries, cfg *config.Config). Handler memakai pgxpool langsung
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
package exam

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math/big"
	mrand "math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/config"
	"cbt-smk-v2/backend/internal/grading"
	"cbt-smk-v2/backend/internal/monitor"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// ---------------------------------------------------------------------------
// Helpers umum
// ---------------------------------------------------------------------------

func writeError(c *echo.Context, status int, code, message string) error {
	return c.JSON(status, map[string]any{
		"error": map[string]any{"code": code, "message": message},
	})
}

func writeOK(c *echo.Context, data any) error {
	return c.JSON(http.StatusOK, map[string]any{"data": data})
}

func writeCreated(c *echo.Context, data any) error {
	return c.JSON(http.StatusCreated, map[string]any{"data": data})
}

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
		_ = writeError(c, http.StatusUnauthorized, "unauthorized", "Harus login.")
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
	_ = writeError(c, http.StatusForbidden, "forbidden", "Role "+role+" tidak punya izin untuk aksi ini.")
	return "", "", "", false
}

func isUUID(s string) bool {
	return len(s) == 36 && strings.Count(s, "-") == 4
}

func newUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // versi 4
	b[8] = (b[8] & 0x3f) | 0x80 // varian
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// --- Settings (dibaca toleran: value jsonb maupun text) ---

func settingRaw(ctx context.Context, pool *pgxpool.Pool, key string) (string, bool) {
	var raw []byte
	err := pool.QueryRow(ctx, `SELECT value FROM settings WHERE key = $1`, key).Scan(&raw)
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(raw)), true
}

func settingInt(ctx context.Context, pool *pgxpool.Pool, key string, def int) int {
	raw, ok := settingRaw(ctx, pool, key)
	if !ok {
		return def
	}
	// Bentuk JSON number ("900") atau text ("900") atau '"900"'.
	var f float64
	if err := json.Unmarshal([]byte(raw), &f); err == nil {
		return int(f)
	}
	if n, err := strconv.Atoi(strings.Trim(raw, `"`)); err == nil {
		return n
	}
	return def
}

func settingBool(ctx context.Context, pool *pgxpool.Pool, key string, def bool) bool {
	raw, ok := settingRaw(ctx, pool, key)
	if !ok {
		return def
	}
	var b bool
	if err := json.Unmarshal([]byte(raw), &b); err == nil {
		return b
	}
	if parsed, err := strconv.ParseBool(strings.Trim(raw, `"`)); err == nil {
		return parsed
	}
	return def
}

func tokenExpiryDetik(ctx context.Context, pool *pgxpool.Pool) int {
	return settingInt(ctx, pool, "token_expiry_detik", 900)
}

func toleransiTelatMenit(ctx context.Context, pool *pgxpool.Pool) int {
	return settingInt(ctx, pool, "toleransi_telat_menit", 15)
}

func maxPelanggaran(ctx context.Context, pool *pgxpool.Pool) int {
	if v := settingInt(ctx, pool, "max_pelanggaran", 2); v >= 1 {
		return v
	}
	return 2
}

// tampilHasilEfektif = paket.tampil_hasil AND settings.tampil_hasil_ke_siswa.
func tampilHasilEfektif(ctx context.Context, pool *pgxpool.Pool, paketTampil bool) bool {
	if !paketTampil {
		return false
	}
	return settingBool(ctx, pool, "tampil_hasil_ke_siswa", false)
}

// --- Waktu jadwal ---

// jakarta memuat zona Asia/Jakarta (fallback: Local bila gagal).
func jakarta() *time.Location {
	if loc, err := time.LoadLocation("Asia/Jakarta"); err == nil {
		return loc
	}
	return time.Local
}

// gabungTanggalJam menggabungkan tanggal "2006-01-02" + jam menjadi
// time.Time di zona sekolah (Asia/Jakarta). Jam diterima dalam format
// "15:04:05" maupun "15:04" (Fase 3: "HH:MM" dinormalisasi ke
// "HH:MM:SS" agar POST jadwal dengan jam "08:00" lolos). Fraksi detik
// pada jam (mis. "15:04:05.123") dipotong agar Parse tidak gagal.
func gabungTanggalJam(tanggal, jam string) (time.Time, error) {
	if i := strings.Index(jam, "."); i >= 0 {
		jam = jam[:i]
	}
	jam = strings.TrimSpace(jam)
	if len(jam) == 5 && jam[2] == ':' {
		jam += ":00"
	}
	return time.ParseInLocation("2006-01-02 15:04:05", tanggal+" "+jam, jakarta())
}

// --- Struktur DB ---

type paket struct {
	ID             string `json:"id"`
	Kode           string `json:"kode"`
	Nama           string `json:"nama"`
	MapelID        string `json:"mapel_id"`
	MapelNama      string `json:"mapel_nama,omitempty"`
	Jenis          string `json:"jenis"`
	DurasiMenit    int    `json:"durasi_menit"`
	JumlahSoal     int    `json:"jumlah_soal"`
	JumlahTersusun int    `json:"jumlah_soal_tersusun,omitempty"`
	AcakSoal       bool   `json:"acak_soal"`
	AcakJawaban    bool   `json:"acak_jawaban"`
	TampilHasil    bool   `json:"tampil_hasil"`
	CreatedAt      any    `json:"created_at"`
}

type jadwal struct {
	ID             string  `json:"id"`
	PaketID        string  `json:"paket_id"`
	PaketKode      string  `json:"paket_kode,omitempty"`
	PaketNama      string  `json:"paket_nama,omitempty"`
	PaketJenis     string  `json:"paket_jenis,omitempty"`
	KelasID        string  `json:"kelas_id"`
	KelasNama      string  `json:"kelas_nama,omitempty"`
	Tanggal        string  `json:"tanggal"`
	JamMulai       string  `json:"jam_mulai"`
	JamSelesai     string  `json:"jam_selesai"`
	Status         string  `json:"status"`
	TokenHash      *string `json:"-"`
	TokenExpiresAt *any    `json:"-"`
	DurasiMenit    int     `json:"durasi_menit,omitempty"`
	AcakSoal       bool    `json:"-"`
	AcakJawaban    bool    `json:"-"`
	TampilHasil    bool    `json:"tampil_hasil,omitempty"`
	CreatedAt      any     `json:"created_at"`
}

type attempt struct {
	ID            string                     `json:"id"`
	JadwalID      string                     `json:"jadwal_id"`
	SiswaID       string                     `json:"siswa_id"`
	Status        string                     `json:"status"`
	MulaiAt       any                        `json:"mulai_at"`
	Deadline      any                        `json:"deadline"`
	SisaWaktuBeku *int                       `json:"sisa_waktu_beku"`
	UrutanSoal    []string                   `json:"-"`
	UrutanOpsi    map[string]json.RawMessage `json:"-"`
	SubmitAt      any                        `json:"submit_at"`
	NilaiAkhir    *float64                   `json:"nilai_akhir"`
	CreatedAt     any                        `json:"created_at"`
}

func scanAttempt(rows pgx.Row) (attempt, error) {
	var a attempt
	var urutanSoal, urutanOpsi []byte
	err := rows.Scan(&a.ID, &a.JadwalID, &a.SiswaID, &a.Status, &a.MulaiAt,
		&a.Deadline, &a.SisaWaktuBeku, &urutanSoal, &urutanOpsi,
		&a.SubmitAt, &a.NilaiAkhir, &a.CreatedAt)
	if err != nil {
		return a, err
	}
	_ = json.Unmarshal(urutanSoal, &a.UrutanSoal)
	_ = json.Unmarshal(urutanOpsi, &a.UrutanOpsi)
	if a.UrutanOpsi == nil {
		a.UrutanOpsi = map[string]json.RawMessage{}
	}
	return a, nil
}

func getJadwal(ctx context.Context, pool *pgxpool.Pool, id string) (jadwal, error) {
	var j jadwal
	var tokenHash, tokenExp *string
	var tokenExpAny any
	err := pool.QueryRow(ctx,
		`SELECT j.id, j.paket_id, j.kelas_id, j.tanggal::text, j.jam_mulai::text,
		        j.jam_selesai::text, j.status, j.token_hash, j.token_expires_at, j.created_at,
		        p.kode, p.nama, p.jenis, p.durasi_menit, p.acak_soal, p.acak_jawaban, p.tampil_hasil,
		        COALESCE(k.nama, '')
		 FROM jadwal_ujian j
		 JOIN paket_ujian p ON p.id = j.paket_id
		 LEFT JOIN kelas k ON k.id = j.kelas_id
		 WHERE j.id = $1`, id).
		Scan(&j.ID, &j.PaketID, &j.KelasID, &j.Tanggal, &j.JamMulai, &j.JamSelesai,
			&j.Status, &tokenHash, &tokenExpAny, &j.CreatedAt,
			&j.PaketKode, &j.PaketNama, &j.PaketJenis, &j.DurasiMenit,
			&j.AcakSoal, &j.AcakJawaban, &j.TampilHasil, &j.KelasNama)
	if err != nil {
		return j, err
	}
	_ = tokenExp
	j.TokenHash = tokenHash
	if tokenExpAny != nil {
		j.TokenExpiresAt = &tokenExpAny
	}
	return j, nil
}

func getAttempt(ctx context.Context, pool *pgxpool.Pool, id string) (attempt, error) {
	return scanAttempt(pool.QueryRow(ctx,
		`SELECT id, jadwal_id, siswa_id, status, mulai_at, deadline,
		        sisa_waktu_beku, urutan_soal, urutan_opsi, submit_at,
		        nilai_akhir, created_at
		 FROM attempt WHERE id = $1`, id))
}

// sisaWaktu menghitung sisa waktu (detik) attempt dari sisi server.
// Bila jadwal dijeda -> sisa beku. Bila deadline lewat -> 0.
func sisaWaktu(a attempt, jadwalStatus string, now time.Time) int {
	if jadwalStatus == "pause" && a.SisaWaktuBeku != nil {
		if *a.SisaWaktuBeku < 0 {
			return 0
		}
		return *a.SisaWaktuBeku
	}
	if dl, ok := a.Deadline.(time.Time); ok {
		sisa := int(dl.Sub(now).Seconds())
		if sisa < 0 {
			return 0
		}
		return sisa
	}
	return 0
}

// shuffleSeed mengocok salinan slice dengan seed deterministik per attempt.
func shuffleSeed(items []string, seed int64) []string {
	out := append([]string(nil), items...)
	r := mrand.New(mrand.NewSource(seed))
	r.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

func seedDariAttemptID(id string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(id))
	return int64(h.Sum64())
}

// shuffleTampilan mengocok data tampil soal Tim 3 (kolom soal.jawaban)
// bila paket acak_jawaban aktif:
//   - PG / PG Kompleks / PG Multi / Pilihan Gambar / Studi Kasus:
//     kocok array jawaban.opsi.
//   - Menjodohkan / Peta Konsep: kocok nilai "kanan" antar pasangan
//     (urutan "kiri" tetap), agar siswa tidak menghafal posisi.
//   - Tipe lain: tidak dikocok.
func shuffleTampilan(tampilanRaw []byte, tipe string, seed int64) json.RawMessage {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(tampilanRaw, &obj); err != nil {
		return tampilanRaw
	}
	kanon := grading.NormalisasiTipe(tipe)
	r := mrand.New(mrand.NewSource(seed))
	switch kanon {
	case "pg", "pg_kompleks":
		var opsi []json.RawMessage
		if err := json.Unmarshal(obj["opsi"], &opsi); err != nil || len(opsi) < 2 {
			return tampilanRaw
		}
		r.Shuffle(len(opsi), func(i, j int) { opsi[i], opsi[j] = opsi[j], opsi[i] })
		obj["opsi"], _ = json.Marshal(opsi)
	case "menjodohkan":
		var pasangan []map[string]json.RawMessage
		if err := json.Unmarshal(obj["pasangan"], &pasangan); err != nil || len(pasangan) < 2 {
			return tampilanRaw
		}
		kanan := make([]json.RawMessage, 0, len(pasangan))
		for _, p := range pasangan {
			kanan = append(kanan, p["kanan"])
		}
		r.Shuffle(len(kanan), func(i, j int) { kanan[i], kanan[j] = kanan[j], kanan[i] })
		for i := range pasangan {
			pasangan[i]["kanan"] = kanan[i]
		}
		obj["pasangan"], _ = json.Marshal(pasangan)
	default:
		return tampilanRaw
	}
	out, err := json.Marshal(obj)
	if err != nil {
		return tampilanRaw
	}
	return out
}

// kunciUntukGrading memperkaya kunci untuk tipe mengurutkan: grading
// butuh daftar item asli (dari tampilan soal) untuk membandingkan
// jawaban string siswa. Frontend mengirim jawaban indeks, jadi ini
// hanya cadangan bila jawaban berupa string item.
func kunciUntukGrading(kunci, tampilan []byte, tipe string) []byte {
	if grading.NormalisasiTipe(tipe) != "mengurutkan" {
		return kunci
	}
	var kunciObj map[string]json.RawMessage
	if err := json.Unmarshal(kunci, &kunciObj); err != nil {
		kunciObj = map[string]json.RawMessage{}
	}
	var tampilanObj map[string]json.RawMessage
	if err := json.Unmarshal(tampilan, &tampilanObj); err == nil {
		if item, ada := tampilanObj["item"]; ada {
			kunciObj["item"] = item
		}
	}
	out, err := json.Marshal(kunciObj)
	if err != nil {
		return kunci
	}
	return out
}

// hitungNilai menghitung nilai akhir attempt dari tabel jawaban + soal.
// Dipanggil saat submit & saat koreksi (attempt selesai).
func hitungNilai(ctx context.Context, pool *pgxpool.Pool, attemptID string) (float64, bool, error) {
	rows, err := pool.Query(ctx,
		`SELECT COALESCE(jw.poin, 0), so.bobot, so.tipe, jw.dikoreksi,
		        (jw.jawaban IS NOT NULL)
		 FROM jawaban jw
		 JOIN soal so ON so.id = jw.soal_id
		 WHERE jw.attempt_id = $1`, attemptID)
	if err != nil {
		return 0, false, err
	}
	defer rows.Close()
	var list []grading.JawabanNilai
	for rows.Next() {
		var poin, bobot float64
		var tipe string
		var dikoreksi, adaJawaban bool
		if err := rows.Scan(&poin, &bobot, &tipe, &dikoreksi, &adaJawaban); err != nil {
			return 0, false, err
		}
		kanon := grading.NormalisasiTipe(tipe)
		dinilai := grading.Dinilai(tipe, nil)
		menunggu := false
		if kanon == "uraian" && !dikoreksi {
			// Soal manual belum dikoreksi -> poin 0 sementara, menunggu.
			menunggu = true
			poin = 0
		}
		if !adaJawaban {
			poin = 0
		}
		list = append(list, grading.JawabanNilai{
			Poin: poin, Bobot: bobot, MenungguKoreksi: menunggu, Dinilai: dinilai,
		})
	}
	if err := rows.Err(); err != nil {
		return 0, false, err
	}
	// Soal dalam paket yang BELUM punya baris jawaban = kosong (poin 0).
	// Hitung lewat paket agar penyebut lengkap.
	rows2, err := pool.Query(ctx,
		`SELECT so.bobot, so.tipe
		 FROM paket_soal ps
		 JOIN jadwal_ujian ju ON ju.paket_id = ps.paket_id
		 JOIN attempt a ON a.jadwal_id = ju.id
		 JOIN soal so ON so.id = ps.soal_id
		 WHERE a.id = $1
		   AND NOT EXISTS (SELECT 1 FROM jawaban jw
		                    WHERE jw.attempt_id = a.id AND jw.soal_id = ps.soal_id)`, attemptID)
	if err != nil {
		return 0, false, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var b float64
		var t string
		if err := rows2.Scan(&b, &t); err != nil {
			return 0, false, err
		}
		if grading.Dinilai(t, nil) {
			list = append(list, grading.JawabanNilai{Poin: 0, Bobot: b, Dinilai: true})
		}
	}
	if err := rows2.Err(); err != nil {
		return 0, false, err
	}
	nilai, menunggu := grading.HitungNilaiAkhir(list)
	return nilai, menunggu, nil
}

// syncNilaiAkhir menyalin nilai attempt ke tabel nilai_akhir (Tim 5) agar
// endpoint rekap membaca nilai yang sama dengan yang dihitung mesin
// ujian. Dipanggil dari finalisasiAttempt (submit/auto-submit) dan
// koreksiJawaban. status dibiarkan 'menunggu_koreksi' sampai proktor
// menjalankan finalisasi rekap (yang mengubahnya jadi 'final') —
// sync TIDAK PERNAH menurunkan status 'final' yang sudah ada.
// Best-effort: gagal sync tidak menggagalkan alur ujian.
func syncNilaiAkhir(ctx context.Context, pool *pgxpool.Pool, a attempt, nilai float64) {
	// nilai_akhir.siswa_id merujuk siswa.id, sedangkan attempt.siswa_id
	// menyimpan users.id (auth_siswa_id) — petakan lewat siswa.user_id.
	var siswaID string
	if err := pool.QueryRow(ctx, `SELECT id FROM siswa WHERE user_id = $1`, a.SiswaID).Scan(&siswaID); err != nil {
		return
	}
	_, _ = pool.Exec(ctx,
		`INSERT INTO nilai_akhir (attempt_id, jadwal_id, siswa_id, nilai, status)
		 VALUES ($1, $2, $3, $4, 'menunggu_koreksi')
		 ON CONFLICT (attempt_id) DO UPDATE SET nilai = EXCLUDED.nilai, updated_at = now()`,
		a.ID, a.JadwalID, siswaID, nilai)
}

// finalisasiAttempt menandai attempt selesai + menghitung nilai akhir.
func finalisasiAttempt(ctx context.Context, pool *pgxpool.Pool, a attempt) (float64, bool, error) {
	nilai, menunggu, err := hitungNilai(ctx, pool, a.ID)
	if err != nil {
		return 0, false, err
	}
	_, err = pool.Exec(ctx,
		`UPDATE attempt SET status = 'selesai', submit_at = now(), nilai_akhir = $2 WHERE id = $1`,
		a.ID, nilai)
	if err != nil {
		return 0, false, err
	}
	syncNilaiAkhir(ctx, pool, a, nilai)
	return nilai, menunggu, nil
}

// jadwalStatusUntuk membaca status jadwal sebuah attempt. Bila gagal
// dibaca, kembalikan "pause" (aman: auto-submit dilewati saat ragu).
func jadwalStatusUntuk(c *echo.Context, pool *pgxpool.Pool, a attempt) string {
	var st string
	if err := pool.QueryRow(c.Request().Context(),
		`SELECT status FROM jadwal_ujian WHERE id = $1`, a.JadwalID).Scan(&st); err != nil {
		return "pause"
	}
	return st
}

// autoSubmitJikaLewat: bila attempt berjalan & deadline lewat, finalisasi.
// Mengembalikan attempt terbaru + apakah baru saja disubmit otomatis.
// Bila jadwal dijeda (pause), timer BEKU -> JANGAN auto-submit walau
// deadline (yang belum disesuaikan) sudah lewat.
func autoSubmitJikaLewat(ctx context.Context, pool *pgxpool.Pool, a attempt, jadwalStatus string) (attempt, bool) {
	if a.Status != "berjalan" || jadwalStatus == "pause" {
		return a, false
	}
	dl, ok := a.Deadline.(time.Time)
	if !ok || time.Now().Before(dl) {
		return a, false
	}
	if _, _, err := finalisasiAttempt(ctx, pool, a); err != nil {
		return a, false
	}
	baru, err := getAttempt(ctx, pool, a.ID)
	if err != nil {
		return a, false
	}
	return baru, true
}

// broadcastAttempt mengirim update ringan monitor untuk satu attempt.
func broadcastAttempt(ctx context.Context, pool *pgxpool.Pool, a attempt) {
	var dijawab, ragu, pelanggaran int
	_ = pool.QueryRow(ctx,
		`SELECT (SELECT COUNT(*) FROM jawaban WHERE attempt_id = $1 AND jawaban IS NOT NULL),
		        (SELECT COUNT(*) FROM jawaban WHERE attempt_id = $1 AND ragu = TRUE),
		        (SELECT COUNT(*) FROM pelanggaran WHERE attempt_id = $1)`,
		a.ID).Scan(&dijawab, &ragu, &pelanggaran)
	monitor.Broadcast(a.JadwalID, map[string]any{
		"type": "update", "jadwal_id": a.JadwalID, "attempt_id": a.ID,
		"status": a.Status, "dijawab": dijawab, "ragu": ragu,
		"pelanggaran": pelanggaran, "at": time.Now().Format(time.RFC3339),
	})
}

// auditCatat menulis satu baris audit log. Gagal catat TIDAK menggagalkan
// operasi utama. Menulis langsung ke audit_log (bukan via paket audit
// Tim 5) karena paket itu belum compile di Echo v5 — lihat currentUser.
func auditCatat(ctx context.Context, pool *pgxpool.Pool, actorID, aksi, entitas, entitasID string, detail any, ip string) {
	detailJSON, err := json.Marshal(detail)
	if err != nil {
		detailJSON = []byte("{}")
	}
	var actor any
	if strings.TrimSpace(actorID) != "" {
		actor = actorID
	}
	_, _ = pool.Exec(ctx,
		`INSERT INTO audit_log (actor_id, aksi, entitas, entitas_id, detail, ip)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		actor, aksi, entitas, entitasID, detailJSON, ip)
}

// ---------------------------------------------------------------------------
// Paket ujian
// ---------------------------------------------------------------------------

func listPaket(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "guru"); !ok {
		return nil
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	jenis := strings.TrimSpace(c.QueryParam("jenis"))
	mapelID := strings.TrimSpace(c.QueryParam("mapel_id"))
	query := `SELECT p.id, p.kode, p.nama, p.mapel_id, COALESCE(m.nama, ''), p.jenis,
	                 p.durasi_menit, p.jumlah_soal, p.acak_soal, p.acak_jawaban,
	                 p.tampil_hasil, p.created_at,
	                 (SELECT COUNT(*) FROM paket_soal ps WHERE ps.paket_id = p.id)
	          FROM paket_ujian p LEFT JOIN mapel m ON m.id = p.mapel_id
	          WHERE ($1::text IS NULL OR p.jenis = $1)
	            AND ($2::uuid IS NULL OR p.mapel_id = $2)
	          ORDER BY p.created_at DESC`
	rows, err := pool.Query(ctx, query, strPtr(jenis), strPtr(mapelID))
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca paket.")
	}
	defer rows.Close()
	items := []paket{}
	for rows.Next() {
		var p paket
		if err := rows.Scan(&p.ID, &p.Kode, &p.Nama, &p.MapelID, &p.MapelNama, &p.Jenis,
			&p.DurasiMenit, &p.JumlahSoal, &p.AcakSoal, &p.AcakJawaban, &p.TampilHasil,
			&p.CreatedAt, &p.JumlahTersusun); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca paket.")
		}
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca paket.")
	}
	return writeOK(c, map[string]any{"items": items})
}

func strPtr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func validJenis(jenis string) bool {
	switch jenis {
	case "STS", "SAS", "USBK", "UKK", "Tryout":
		return true
	}
	return false
}

func bindPaket(c *echo.Context) (paket, bool) {
	var body struct {
		Kode        string `json:"kode"`
		Nama        string `json:"nama"`
		MapelID     string `json:"mapel_id"`
		Jenis       string `json:"jenis"`
		DurasiMenit int    `json:"durasi_menit"`
		JumlahSoal  int    `json:"jumlah_soal"`
		AcakSoal    *bool  `json:"acak_soal"`
		AcakJawaban *bool  `json:"acak_jawaban"`
		TampilHasil *bool  `json:"tampil_hasil"`
	}
	if err := c.Bind(&body); err != nil {
		_ = writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
		return paket{}, false
	}
	body.Kode = strings.TrimSpace(body.Kode)
	body.Nama = strings.TrimSpace(body.Nama)
	body.Jenis = strings.TrimSpace(body.Jenis)
	if body.Kode == "" || body.Nama == "" || body.MapelID == "" {
		_ = writeError(c, http.StatusBadRequest, "validation_error", "kode, nama, dan mapel_id wajib diisi.")
		return paket{}, false
	}
	if !validJenis(body.Jenis) {
		_ = writeError(c, http.StatusBadRequest, "validation_error", "jenis harus STS/SAS/USBK/UKK/Tryout.")
		return paket{}, false
	}
	if body.DurasiMenit <= 0 {
		_ = writeError(c, http.StatusBadRequest, "validation_error", "durasi_menit harus > 0.")
		return paket{}, false
	}
	if body.JumlahSoal < 0 {
		_ = writeError(c, http.StatusBadRequest, "validation_error", "jumlah_soal tidak boleh negatif.")
		return paket{}, false
	}
	p := paket{Kode: body.Kode, Nama: body.Nama, MapelID: body.MapelID, Jenis: body.Jenis,
		DurasiMenit: body.DurasiMenit, JumlahSoal: body.JumlahSoal,
		AcakSoal: true, AcakJawaban: true, TampilHasil: false}
	if body.AcakSoal != nil {
		p.AcakSoal = *body.AcakSoal
	}
	if body.AcakJawaban != nil {
		p.AcakJawaban = *body.AcakJawaban
	}
	if body.TampilHasil != nil {
		p.TampilHasil = *body.TampilHasil
	}
	return p, true
}

func createPaket(c *echo.Context) error {
	uid, _, _, ok := requireRole(c, "admin", "guru")
	if !ok {
		return nil
	}
	p, ok := bindPaket(c)
	if !ok {
		return nil
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	var mapelAda bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM mapel WHERE id = $1)`, p.MapelID).Scan(&mapelAda); err != nil || !mapelAda {
		return writeError(c, http.StatusBadRequest, "validation_error", "mapel_id tidak ditemukan.")
	}
	var id string
	err := pool.QueryRow(ctx,
		`INSERT INTO paket_ujian (kode, nama, mapel_id, jenis, durasi_menit, jumlah_soal,
		                           acak_soal, acak_jawaban, tampil_hasil)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`,
		p.Kode, p.Nama, p.MapelID, p.Jenis, p.DurasiMenit, p.JumlahSoal,
		p.AcakSoal, p.AcakJawaban, p.TampilHasil).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return writeError(c, http.StatusConflict, "conflict", "Kode paket sudah dipakai.")
		}
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membuat paket.")
	}
	p.ID = id
	auditCatat(ctx, pool, uid, "paket.create", "paket_ujian", id,
		map[string]any{"kode": p.Kode, "nama": p.Nama}, c.RealIP())
	return writeCreated(c, p)
}

func getPaket(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "guru"); !ok {
		return nil
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	id := c.Param("id")
	if !isUUID(id) {
		return writeError(c, http.StatusBadRequest, "validation_error", "ID paket tidak valid.")
	}
	var p paket
	err := pool.QueryRow(ctx,
		`SELECT p.id, p.kode, p.nama, p.mapel_id, COALESCE(m.nama, ''), p.jenis,
		        p.durasi_menit, p.jumlah_soal, p.acak_soal, p.acak_jawaban,
		        p.tampil_hasil, p.created_at,
		        (SELECT COUNT(*) FROM paket_soal ps WHERE ps.paket_id = p.id)
		 FROM paket_ujian p LEFT JOIN mapel m ON m.id = p.mapel_id
		 WHERE p.id = $1`, id).
		Scan(&p.ID, &p.Kode, &p.Nama, &p.MapelID, &p.MapelNama, &p.Jenis,
			&p.DurasiMenit, &p.JumlahSoal, &p.AcakSoal, &p.AcakJawaban,
			&p.TampilHasil, &p.CreatedAt, &p.JumlahTersusun)
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Paket tidak ditemukan.")
	}
	return writeOK(c, p)
}

func updatePaket(c *echo.Context) error {
	uid, _, _, ok := requireRole(c, "admin", "guru")
	if !ok {
		return nil
	}
	id := c.Param("id")
	if !isUUID(id) {
		return writeError(c, http.StatusBadRequest, "validation_error", "ID paket tidak valid.")
	}
	p, ok := bindPaket(c)
	if !ok {
		return nil
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	var lamaMapel string
	err := pool.QueryRow(ctx, `SELECT mapel_id FROM paket_ujian WHERE id = $1`, id).Scan(&lamaMapel)
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Paket tidak ditemukan.")
	}
	if lamaMapel != p.MapelID {
		var jumlahSoalLain int
		_ = pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM paket_soal ps
			 JOIN soal s ON s.id = ps.soal_id
			 JOIN bank_soal bs ON bs.id = s.bank_id
			 WHERE ps.paket_id = $1 AND bs.mapel_id <> $2`, id, p.MapelID).Scan(&jumlahSoalLain)
		if jumlahSoalLain > 0 {
			return writeError(c, http.StatusUnprocessableEntity, "validation_error",
				"Paket masih berisi soal dari mapel lain. Kosongkan susunan soal dulu sebelum ganti mapel.")
		}
	}
	ct, err := pool.Exec(ctx,
		`UPDATE paket_ujian SET kode=$2, nama=$3, mapel_id=$4, jenis=$5, durasi_menit=$6,
		        jumlah_soal=$7, acak_soal=$8, acak_jawaban=$9, tampil_hasil=$10
		 WHERE id=$1`, id, p.Kode, p.Nama, p.MapelID, p.Jenis, p.DurasiMenit,
		p.JumlahSoal, p.AcakSoal, p.AcakJawaban, p.TampilHasil)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return writeError(c, http.StatusConflict, "conflict", "Kode paket sudah dipakai.")
		}
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mengubah paket.")
	}
	if ct.RowsAffected() == 0 {
		return writeError(c, http.StatusNotFound, "not_found", "Paket tidak ditemukan.")
	}
	auditCatat(ctx, pool, uid, "paket.update", "paket_ujian", id,
		map[string]any{"kode": p.Kode, "nama": p.Nama}, c.RealIP())
	return writeOK(c, map[string]any{"id": id, "updated": true})
}

func deletePaket(c *echo.Context) error {
	uid, _, _, ok := requireRole(c, "admin", "guru")
	if !ok {
		return nil
	}
	id := c.Param("id")
	if !isUUID(id) {
		return writeError(c, http.StatusBadRequest, "validation_error", "ID paket tidak valid.")
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	var dipakai int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM jadwal_ujian WHERE paket_id = $1`, id).Scan(&dipakai)
	if dipakai > 0 {
		return writeError(c, http.StatusConflict, "conflict", "Paket masih dipakai jadwal ujian. Hapus jadwal dulu.")
	}
	ct, err := pool.Exec(ctx, `DELETE FROM paket_ujian WHERE id = $1`, id)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghapus paket.")
	}
	if ct.RowsAffected() == 0 {
		return writeError(c, http.StatusNotFound, "not_found", "Paket tidak ditemukan.")
	}
	auditCatat(ctx, pool, uid, "paket.delete", "paket_ujian", id, nil, c.RealIP())
	return writeOK(c, map[string]any{"id": id, "deleted": true})
}

// sinkronJumlahSoal menyamakan paket.jumlah_soal dengan jumlah susunan aktual.
func sinkronJumlahSoal(ctx context.Context, pool *pgxpool.Pool, paketID string) {
	_, _ = pool.Exec(ctx,
		`UPDATE paket_ujian SET jumlah_soal = (SELECT COUNT(*) FROM paket_soal WHERE paket_id = $1)
		 WHERE id = $1`, paketID)
}

func listPaketSoal(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "guru"); !ok {
		return nil
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	id := c.Param("id")
	rows, err := pool.Query(ctx,
		`SELECT ps.soal_id, ps.urutan, s.tipe, s.pertanyaan, s.bobot
		 FROM paket_soal ps JOIN soal s ON s.id = ps.soal_id
		 WHERE ps.paket_id = $1 ORDER BY ps.urutan ASC`, id)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca susunan soal.")
	}
	defer rows.Close()
	type item struct {
		SoalID     string `json:"soal_id"`
		Urutan     int    `json:"urutan"`
		Tipe       string `json:"tipe"`
		Pertanyaan string `json:"pertanyaan"`
		Bobot      int    `json:"bobot"`
	}
	items := []item{}
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.SoalID, &it.Urutan, &it.Tipe, &it.Pertanyaan, &it.Bobot); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca susunan soal.")
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca susunan soal.")
	}
	return writeOK(c, map[string]any{"paket_id": id, "items": items})
}

func addPaketSoal(c *echo.Context) error {
	uid, _, _, ok := requireRole(c, "admin", "guru")
	if !ok {
		return nil
	}
	paketID := c.Param("id")
	var body struct {
		SoalID string `json:"soal_id"`
	}
	if err := c.Bind(&body); err != nil || body.SoalID == "" {
		return writeError(c, http.StatusBadRequest, "validation_error", "soal_id wajib diisi.")
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	var paketMapel string
	if err := pool.QueryRow(ctx, `SELECT mapel_id FROM paket_ujian WHERE id = $1`, paketID).Scan(&paketMapel); err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Paket tidak ditemukan.")
	}
	var soalMapel string
	if err := pool.QueryRow(ctx,
		`SELECT bs.mapel_id FROM soal s JOIN bank_soal bs ON bs.id = s.bank_id WHERE s.id = $1`,
		body.SoalID).Scan(&soalMapel); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "soal_id tidak ditemukan.")
	}
	if soalMapel != paketMapel {
		return writeError(c, http.StatusUnprocessableEntity, "validation_error", "Soal bukan dari mapel paket ini.")
	}
	var maxUrutan int
	_ = pool.QueryRow(ctx, `SELECT COALESCE(MAX(urutan), 0) FROM paket_soal WHERE paket_id = $1`, paketID).Scan(&maxUrutan)
	_, err := pool.Exec(ctx,
		`INSERT INTO paket_soal (paket_id, soal_id, urutan) VALUES ($1, $2, $3)
		 ON CONFLICT (paket_id, soal_id) DO NOTHING`, paketID, body.SoalID, maxUrutan+1)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menambah soal ke paket.")
	}
	sinkronJumlahSoal(ctx, pool, paketID)
	auditCatat(ctx, pool, uid, "paket.soal.add", "paket_ujian", paketID,
		map[string]any{"soal_id": body.SoalID}, c.RealIP())
	return writeCreated(c, map[string]any{"paket_id": paketID, "soal_id": body.SoalID, "urutan": maxUrutan + 1})
}

func deletePaketSoal(c *echo.Context) error {
	uid, _, _, ok := requireRole(c, "admin", "guru")
	if !ok {
		return nil
	}
	paketID := c.Param("id")
	soalID := strings.TrimSpace(c.QueryParam("soal_id"))
	if soalID == "" {
		var body struct {
			SoalID string `json:"soal_id"`
		}
		_ = c.Bind(&body)
		soalID = body.SoalID
	}
	if soalID == "" {
		return writeError(c, http.StatusBadRequest, "validation_error", "soal_id wajib diisi (body atau ?soal_id=).")
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	ct, err := pool.Exec(ctx, `DELETE FROM paket_soal WHERE paket_id = $1 AND soal_id = $2`, paketID, soalID)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghapus soal dari paket.")
	}
	if ct.RowsAffected() == 0 {
		return writeError(c, http.StatusNotFound, "not_found", "Soal tidak ada di paket ini.")
	}
	sinkronJumlahSoal(ctx, pool, paketID)
	auditCatat(ctx, pool, uid, "paket.soal.delete", "paket_ujian", paketID,
		map[string]any{"soal_id": soalID}, c.RealIP())
	return writeOK(c, map[string]any{"paket_id": paketID, "soal_id": soalID, "deleted": true})
}

func generateAcak(c *echo.Context) error {
	uid, _, _, ok := requireRole(c, "admin", "guru")
	if !ok {
		return nil
	}
	paketID := c.Param("id")
	var body struct {
		Jumlah int `json:"jumlah"`
	}
	if err := c.Bind(&body); err != nil || body.Jumlah <= 0 {
		return writeError(c, http.StatusBadRequest, "validation_error", "jumlah harus > 0.")
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	var paketMapel string
	if err := pool.QueryRow(ctx, `SELECT mapel_id FROM paket_ujian WHERE id = $1`, paketID).Scan(&paketMapel); err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Paket tidak ditemukan.")
	}
	rows, err := pool.Query(ctx,
		`SELECT s.id FROM soal s JOIN bank_soal bs ON bs.id = s.bank_id
		 WHERE bs.mapel_id = $1 AND s.status = 'aktif'
		   AND s.id NOT IN (SELECT soal_id FROM paket_soal WHERE paket_id = $2)
		 ORDER BY random() LIMIT $3`, paketMapel, paketID, body.Jumlah)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mengambil soal acak.")
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mengambil soal acak.")
	}
	var maxUrutan int
	_ = pool.QueryRow(ctx, `SELECT COALESCE(MAX(urutan), 0) FROM paket_soal WHERE paket_id = $1`, paketID).Scan(&maxUrutan)
	for i, soalID := range ids {
		_, _ = pool.Exec(ctx,
			`INSERT INTO paket_soal (paket_id, soal_id, urutan) VALUES ($1, $2, $3)
			 ON CONFLICT (paket_id, soal_id) DO NOTHING`, paketID, soalID, maxUrutan+i+1)
	}
	sinkronJumlahSoal(ctx, pool, paketID)
	auditCatat(ctx, pool, uid, "paket.generate_acak", "paket_ujian", paketID,
		map[string]any{"diminta": body.Jumlah, "ditambahkan": len(ids)}, c.RealIP())
	resp := map[string]any{"paket_id": paketID, "diminta": body.Jumlah, "ditambahkan": len(ids)}
	if len(ids) < body.Jumlah {
		resp["peringatan"] = "Soal tersedia di bank mapel ini kurang dari yang diminta."
	}
	return writeOK(c, resp)
}

// ---------------------------------------------------------------------------
// Jadwal ujian
// ---------------------------------------------------------------------------

func listJadwal(c *echo.Context) error {
	// penguji_ukk boleh BACA jadwal (Rekap butuh daftar jadwal; tanpa ini
	// menu Rekap yang tampil untuk penguji_ukk selalu 403). Tulis tetap
	// admin/proktor/guru/pengawas (create/update/delete tidak berubah).
	if _, _, _, ok := requireRole(c, "admin", "proktor", "guru", "pengawas", "penguji_ukk"); !ok {
		return nil
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	status := strings.TrimSpace(c.QueryParam("status"))
	kelasID := strings.TrimSpace(c.QueryParam("kelas_id"))
	tanggal := strings.TrimSpace(c.QueryParam("tanggal"))
	rows, err := pool.Query(ctx,
		`SELECT j.id, j.paket_id, j.kelas_id, j.tanggal::text, j.jam_mulai::text,
		        j.jam_selesai::text, j.status, j.token_expires_at, j.created_at,
		        p.kode, p.nama, p.jenis, COALESCE(k.nama, '')
		 FROM jadwal_ujian j
		 JOIN paket_ujian p ON p.id = j.paket_id
		 LEFT JOIN kelas k ON k.id = j.kelas_id
		 WHERE ($1::text IS NULL OR j.status = $1)
		   AND ($2::uuid IS NULL OR j.kelas_id = $2)
		   AND ($3::date IS NULL OR j.tanggal = $3)
		 ORDER BY j.tanggal DESC, j.jam_mulai DESC`,
		strPtr(status), strPtr(kelasID), strPtr(tanggal))
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca jadwal.")
	}
	defer rows.Close()
	type item struct {
		jadwal
		TokenAda bool `json:"token_ada"`
	}
	items := []item{}
	for rows.Next() {
		var it item
		var tokenExp any
		if err := rows.Scan(&it.ID, &it.PaketID, &it.KelasID, &it.Tanggal, &it.JamMulai,
			&it.JamSelesai, &it.Status, &tokenExp, &it.CreatedAt,
			&it.PaketKode, &it.PaketNama, &it.PaketJenis, &it.KelasNama); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca jadwal.")
		}
		it.TokenAda = tokenExp != nil
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca jadwal.")
	}
	return writeOK(c, map[string]any{"items": items})
}

func bindJadwal(c *echo.Context) (jadwal, bool) {
	var body struct {
		PaketID    string `json:"paket_id"`
		KelasID    string `json:"kelas_id"`
		Tanggal    string `json:"tanggal"`
		JamMulai   string `json:"jam_mulai"`
		JamSelesai string `json:"jam_selesai"`
		Status     string `json:"status"`
	}
	if err := c.Bind(&body); err != nil {
		_ = writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
		return jadwal{}, false
	}
	body.Tanggal = strings.TrimSpace(body.Tanggal)
	body.JamMulai = strings.TrimSpace(body.JamMulai)
	body.JamSelesai = strings.TrimSpace(body.JamSelesai)
	if body.PaketID == "" || body.KelasID == "" || body.Tanggal == "" || body.JamMulai == "" || body.JamSelesai == "" {
		_ = writeError(c, http.StatusBadRequest, "validation_error", "paket_id, kelas_id, tanggal, jam_mulai, jam_selesai wajib diisi.")
		return jadwal{}, false
	}
	mulai, err1 := gabungTanggalJam(body.Tanggal, body.JamMulai)
	selesai, err2 := gabungTanggalJam(body.Tanggal, body.JamSelesai)
	if err1 != nil || err2 != nil || !selesai.After(mulai) {
		_ = writeError(c, http.StatusBadRequest, "validation_error", "Tanggal/jam tidak valid; jam_selesai harus sesudah jam_mulai.")
		return jadwal{}, false
	}
	status := strings.TrimSpace(body.Status)
	if status == "" {
		status = "draft"
	}
	if status != "draft" && status != "aktif" && status != "pause" && status != "selesai" {
		_ = writeError(c, http.StatusBadRequest, "validation_error", "status harus draft/aktif/pause/selesai.")
		return jadwal{}, false
	}
	return jadwal{PaketID: body.PaketID, KelasID: body.KelasID, Tanggal: body.Tanggal,
		JamMulai: body.JamMulai, JamSelesai: body.JamSelesai, Status: status}, true
}

func createJadwal(c *echo.Context) error {
	uid, _, _, ok := requireRole(c, "admin", "proktor", "guru")
	if !ok {
		return nil
	}
	j, ok := bindJadwal(c)
	if !ok {
		return nil
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	var paketAda, kelasAda bool
	_ = pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM paket_ujian WHERE id = $1)`, j.PaketID).Scan(&paketAda)
	_ = pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM kelas WHERE id = $1)`, j.KelasID).Scan(&kelasAda)
	if !paketAda {
		return writeError(c, http.StatusBadRequest, "validation_error", "paket_id tidak ditemukan.")
	}
	if !kelasAda {
		return writeError(c, http.StatusBadRequest, "validation_error", "kelas_id tidak ditemukan.")
	}
	var id string
	err := pool.QueryRow(ctx,
		`INSERT INTO jadwal_ujian (paket_id, kelas_id, tanggal, jam_mulai, jam_selesai, status)
		 VALUES ($1, $2, $3::date, $4::time, $5::time, $6) RETURNING id`,
		j.PaketID, j.KelasID, j.Tanggal, j.JamMulai, j.JamSelesai, j.Status).Scan(&id)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membuat jadwal.")
	}
	j.ID = id
	auditCatat(ctx, pool, uid, "jadwal.create", "jadwal_ujian", id,
		map[string]any{"paket_id": j.PaketID, "kelas_id": j.KelasID, "tanggal": j.Tanggal}, c.RealIP())
	return writeCreated(c, j)
}

func getJadwalDetail(c *echo.Context) error {
	// penguji_ukk boleh BACA detail jadwal (tanpa token_hash — tidak pernah
	// dikembalikan endpoint ini). Tulis tetap admin/proktor/guru/pengawas.
	if _, _, _, ok := requireRole(c, "admin", "proktor", "guru", "pengawas", "penguji_ukk"); !ok {
		return nil
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	id := c.Param("id")
	if !isUUID(id) {
		return writeError(c, http.StatusBadRequest, "validation_error", "ID jadwal tidak valid.")
	}
	j, err := getJadwal(ctx, pool, id)
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
	}
	// JANGAN pernah kembalikan token_hash ke client.
	return writeOK(c, map[string]any{
		"id": j.ID, "paket_id": j.PaketID, "paket_kode": j.PaketKode,
		"paket_nama": j.PaketNama, "paket_jenis": j.PaketJenis,
		"kelas_id": j.KelasID, "kelas_nama": j.KelasNama,
		"tanggal": j.Tanggal, "jam_mulai": j.JamMulai, "jam_selesai": j.JamSelesai,
		"status": j.Status, "durasi_menit": j.DurasiMenit,
		"tampil_hasil": j.TampilHasil, "created_at": j.CreatedAt,
		"token_ada":        j.TokenHash != nil,
		"token_expires_at": j.TokenExpiresAt,
	})
}

func updateJadwal(c *echo.Context) error {
	uid, _, _, ok := requireRole(c, "admin", "proktor", "guru")
	if !ok {
		return nil
	}
	id := c.Param("id")
	if !isUUID(id) {
		return writeError(c, http.StatusBadRequest, "validation_error", "ID jadwal tidak valid.")
	}
	j, ok := bindJadwal(c)
	if !ok {
		return nil
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	var lamaStatus string
	if err := pool.QueryRow(ctx, `SELECT status FROM jadwal_ujian WHERE id = $1`, id).Scan(&lamaStatus); err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
	}
	if lamaStatus == "selesai" {
		return writeError(c, http.StatusConflict, "conflict", "Jadwal yang sudah selesai tidak boleh diubah.")
	}
	ct, err := pool.Exec(ctx,
		`UPDATE jadwal_ujian SET paket_id=$2, kelas_id=$3, tanggal=$4::date,
		        jam_mulai=$5::time, jam_selesai=$6::time, status=$7 WHERE id=$1`,
		id, j.PaketID, j.KelasID, j.Tanggal, j.JamMulai, j.JamSelesai, j.Status)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mengubah jadwal.")
	}
	if ct.RowsAffected() == 0 {
		return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
	}
	auditCatat(ctx, pool, uid, "jadwal.update", "jadwal_ujian", id,
		map[string]any{"status": j.Status}, c.RealIP())
	return writeOK(c, map[string]any{"id": id, "updated": true})
}

func deleteJadwal(c *echo.Context) error {
	uid, _, _, ok := requireRole(c, "admin", "proktor", "guru")
	if !ok {
		return nil
	}
	id := c.Param("id")
	if !isUUID(id) {
		return writeError(c, http.StatusBadRequest, "validation_error", "ID jadwal tidak valid.")
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	var jumlahAttempt int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM attempt WHERE jadwal_id = $1`, id).Scan(&jumlahAttempt)
	if jumlahAttempt > 0 {
		return writeError(c, http.StatusConflict, "conflict", "Jadwal sudah punya attempt peserta. Tidak boleh dihapus (hasil ujian tidak pernah dihapus fisik).")
	}
	ct, err := pool.Exec(ctx, `DELETE FROM jadwal_ujian WHERE id = $1`, id)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghapus jadwal.")
	}
	if ct.RowsAffected() == 0 {
		return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
	}
	auditCatat(ctx, pool, uid, "jadwal.delete", "jadwal_ujian", id, nil, c.RealIP())
	return writeOK(c, map[string]any{"id": id, "deleted": true})
}

// --- Pause / resume ---

func pauseJadwal(c *echo.Context) error {
	uid, _, _, ok := requireRole(c, "admin", "proktor")
	if !ok {
		return nil
	}
	id := c.Param("id")
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	j, err := getJadwal(ctx, pool, id)
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
	}
	if j.Status != "aktif" {
		return writeError(c, http.StatusConflict, "conflict", "Hanya jadwal aktif yang bisa dijeda (status: "+j.Status+").")
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memulai pause.")
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `UPDATE jadwal_ujian SET status = 'pause' WHERE id = $1`, id); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menjeda jadwal.")
	}
	// Bekukan sisa waktu semua attempt berjalan: sisa = deadline - now.
	if _, err := tx.Exec(ctx,
		`UPDATE attempt
		 SET sisa_waktu_beku = GREATEST(0, EXTRACT(EPOCH FROM (deadline - now()))::INT)
		 WHERE jadwal_id = $1 AND status = 'berjalan'`, id); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membekukan timer peserta.")
	}
	if err := tx.Commit(ctx); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menjeda jadwal.")
	}
	auditCatat(ctx, pool, uid, "jadwal.pause", "jadwal_ujian", id, nil, c.RealIP())
	monitor.Broadcast(id, map[string]any{"type": "jadwal_pause", "jadwal_id": id})
	return writeOK(c, map[string]any{"id": id, "status": "pause"})
}

func resumeJadwal(c *echo.Context) error {
	uid, _, _, ok := requireRole(c, "admin", "proktor")
	if !ok {
		return nil
	}
	id := c.Param("id")
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	j, err := getJadwal(ctx, pool, id)
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
	}
	if j.Status != "pause" {
		return writeError(c, http.StatusConflict, "conflict", "Hanya jadwal yang dijeda yang bisa dilanjutkan (status: "+j.Status+").")
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memulai resume.")
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `UPDATE jadwal_ujian SET status = 'aktif' WHERE id = $1`, id); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal melanjutkan jadwal.")
	}
	// Deadline baru = now() + sisa_waktu_beku.
	if _, err := tx.Exec(ctx,
		`UPDATE attempt
		 SET deadline = now() + make_interval(secs => COALESCE(sisa_waktu_beku, 0)),
		     sisa_waktu_beku = NULL
		 WHERE jadwal_id = $1 AND status = 'berjalan' AND sisa_waktu_beku IS NOT NULL`, id); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memulihkan timer peserta.")
	}
	if err := tx.Commit(ctx); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal melanjutkan jadwal.")
	}
	auditCatat(ctx, pool, uid, "jadwal.resume", "jadwal_ujian", id, nil, c.RealIP())
	monitor.Broadcast(id, map[string]any{"type": "jadwal_resume", "jadwal_id": id})
	return writeOK(c, map[string]any{"id": id, "status": "aktif"})
}

// --- Token ---

const tokenAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 32 simbol, tanpa 0/1/I/O

func generateToken() (string, error) {
	var sb strings.Builder
	for i := 0; i < 6; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(tokenAlphabet))))
		if err != nil {
			return "", err
		}
		sb.WriteByte(tokenAlphabet[n.Int64()])
	}
	return sb.String(), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func generateJadwalToken(c *echo.Context) error {
	uid, _, _, ok := requireRole(c, "admin", "proktor")
	if !ok {
		return nil
	}
	id := c.Param("id")
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	j, err := getJadwal(ctx, pool, id)
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
	}
	if j.Status == "selesai" {
		return writeError(c, http.StatusConflict, "conflict", "Jadwal sudah selesai; token tidak bisa dibuat.")
	}
	token, err := generateToken()
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membuat token.")
	}
	// Expiry = waktu GENERATE + token_expiry_detik (default 900 dtk = 15 menit).
	// Sebelumnya expiry = jam mulai jadwal + 900 dtk, sehingga token yang
	// di-generate setelah jam mulai + 15 menit langsung lahir KEDALUWARSA
	// (mis. jadwal mulai 18.43, generate jam 20.00 → expiry 18.58).
	// Token kedua tetap menimpa (HANGUSKAN) token pertama: hanya satu hash disimpan.
	expiry := time.Now().In(jakarta()).Add(time.Duration(tokenExpiryDetik(ctx, pool)) * time.Second)
	if _, err := pool.Exec(ctx,
		`UPDATE jadwal_ujian SET token_hash = $2, token_expires_at = $3 WHERE id = $1`,
		id, hashToken(token), expiry); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menyimpan token.")
	}
	// Audit TIDAK BOLEH mencatat token plaintext (lihat docs/KEAMANAN.md).
	auditCatat(ctx, pool, uid, "jadwal.token.generate", "jadwal_ujian", id,
		map[string]any{"token_expires_at": expiry.Format(time.RFC3339)}, c.RealIP())
	// Token plaintext dikembalikan SEKALI di respons ini.
	return writeOK(c, map[string]any{
		"jadwal_id": id, "token": token,
		"token_expires_at": expiry.Format(time.RFC3339),
		"peringatan":       "Token hanya ditampilkan sekali. Salin sekarang — token tidak bisa dilihat lagi.",
	})
}

func tokenStatus(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "proktor"); !ok {
		return nil
	}
	id := c.Param("id")
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	j, err := getJadwal(ctx, pool, id)
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
	}
	// TANPA menampilkan token (nilai token tidak pernah dikembalikan).
	resp := map[string]any{"jadwal_id": id, "ada_token": j.TokenHash != nil}
	if j.TokenExpiresAt != nil {
		if t, ok := (*j.TokenExpiresAt).(time.Time); ok {
			resp["token_expires_at"] = t.Format(time.RFC3339)
			sisa := int(time.Until(t).Seconds())
			if sisa < 0 {
				sisa = 0
			}
			resp["kedaluwarsa_dalam_detik"] = sisa
			resp["kedaluwarsa"] = sisa == 0
		}
	}
	return writeOK(c, resp)
}

// --- Usulan unblock / unblock ---

func listUsulanUnblock(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "proktor", "pengawas"); !ok {
		return nil
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	id := c.Param("id")
	rows, err := pool.Query(ctx,
		`SELECT u.id, u.attempt_id, u.pengawas_id, u.alasan, u.status,
		        u.diproses_oleh, u.diproses_at, u.created_at,
		        COALESCE(s.nama, '')
		 FROM usulan_unblock u
		 JOIN attempt a ON a.id = u.attempt_id
		 LEFT JOIN siswa s ON s.user_id = a.siswa_id
		 WHERE a.jadwal_id = $1 ORDER BY u.created_at DESC`, id)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca usulan unblock.")
	}
	defer rows.Close()
	type item struct {
		ID           string  `json:"id"`
		AttemptID    string  `json:"attempt_id"`
		PengawasID   string  `json:"pengawas_id"`
		Alasan       string  `json:"alasan"`
		Status       string  `json:"status"`
		DiprosesOleh *string `json:"diproses_oleh"`
		DiprosesAt   any     `json:"diproses_at"`
		CreatedAt    any     `json:"created_at"`
		NamaSiswa    string  `json:"nama_siswa"`
	}
	items := []item{}
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.ID, &it.AttemptID, &it.PengawasID, &it.Alasan, &it.Status,
			&it.DiprosesOleh, &it.DiprosesAt, &it.CreatedAt, &it.NamaSiswa); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca usulan unblock.")
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca usulan unblock.")
	}
	return writeOK(c, map[string]any{"jadwal_id": id, "items": items})
}

func createUsulanUnblock(c *echo.Context) error {
	uid, _, _, ok := requireRole(c, "pengawas", "proktor", "admin")
	if !ok {
		return nil
	}
	jadwalID := c.Param("id")
	var body struct {
		AttemptID string `json:"attempt_id"`
		Alasan    string `json:"alasan"`
	}
	if err := c.Bind(&body); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	body.Alasan = strings.TrimSpace(body.Alasan)
	if body.AttemptID == "" || body.Alasan == "" {
		return writeError(c, http.StatusBadRequest, "validation_error", "attempt_id dan alasan wajib diisi.")
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	a, err := getAttempt(ctx, pool, body.AttemptID)
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Attempt tidak ditemukan.")
	}
	if a.JadwalID != jadwalID {
		return writeError(c, http.StatusBadRequest, "validation_error", "Attempt bukan dari jadwal ini.")
	}
	if a.Status != "diblokir" {
		return writeError(c, http.StatusConflict, "conflict", "Hanya attempt yang diblokir yang bisa diusulkan unblock (status: "+a.Status+").")
	}
	var pending int
	_ = pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM usulan_unblock WHERE attempt_id = $1 AND status = 'menunggu'`,
		body.AttemptID).Scan(&pending)
	if pending > 0 {
		return writeError(c, http.StatusConflict, "conflict", "Sudah ada usulan unblock yang menunggu untuk attempt ini.")
	}
	var id string
	if err := pool.QueryRow(ctx,
		`INSERT INTO usulan_unblock (attempt_id, pengawas_id, alasan, status)
		 VALUES ($1, $2, $3, 'menunggu') RETURNING id`,
		body.AttemptID, uid, body.Alasan).Scan(&id); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menyimpan usulan unblock.")
	}
	auditCatat(ctx, pool, uid, "jadwal.usulan_unblock", "attempt", body.AttemptID,
		map[string]any{"jadwal_id": jadwalID, "alasan": body.Alasan}, c.RealIP())
	monitor.Broadcast(jadwalID, map[string]any{"type": "usulan_unblock", "jadwal_id": jadwalID, "attempt_id": body.AttemptID})
	return writeCreated(c, map[string]any{"id": id, "attempt_id": body.AttemptID, "status": "menunggu"})
}

func unblockAttempt(c *echo.Context) error {
	uid, _, _, ok := requireRole(c, "proktor", "admin")
	if !ok {
		return nil
	}
	jadwalID := c.Param("id")
	var body struct {
		AttemptID string `json:"attempt_id"`
		Alasan    string `json:"alasan"`
	}
	if err := c.Bind(&body); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	body.Alasan = strings.TrimSpace(body.Alasan)
	if body.AttemptID == "" || body.Alasan == "" {
		return writeError(c, http.StatusBadRequest, "validation_error", "attempt_id dan alasan wajib diisi.")
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	a, err := getAttempt(ctx, pool, body.AttemptID)
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Attempt tidak ditemukan.")
	}
	if a.JadwalID != jadwalID {
		return writeError(c, http.StatusBadRequest, "validation_error", "Attempt bukan dari jadwal ini.")
	}
	if a.Status != "diblokir" {
		return writeError(c, http.StatusConflict, "conflict", "Hanya attempt yang diblokir yang bisa di-unblock (status: "+a.Status+").")
	}
	j, err := getJadwal(ctx, pool, jadwalID)
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
	}
	// Sisa waktu untuk dilanjutkan: sisa beku bila ada, bila tidak
	// sisa = jadwal_selesai - now. Bila sudah lewat -> tidak bisa unblock.
	sisa := 0
	if a.SisaWaktuBeku != nil {
		sisa = *a.SisaWaktuBeku
	} else {
		selesai, err := gabungTanggalJam(j.Tanggal, j.JamSelesai)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Tanggal/jam jadwal tidak valid.")
		}
		sisa = int(time.Until(selesai).Seconds())
	}
	if sisa <= 0 {
		return writeError(c, http.StatusConflict, "conflict", "Ujian sudah berakhir (tidak ada sisa waktu). Unblock tidak bisa dilakukan.")
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memulai unblock.")
	}
	defer tx.Rollback(ctx)
	// LANJUT sisa waktu: jawaban tersimpan, TIDAK reset.
	if _, err := tx.Exec(ctx,
		`UPDATE attempt SET status = 'berjalan',
		        deadline = now() + make_interval(secs => $2),
		        sisa_waktu_beku = NULL WHERE id = $1`, body.AttemptID, sisa); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal unblock attempt.")
	}
	// Tutup usulan yang menunggu (bila ada); bila tidak ada usulan
	// (proktor unblock langsung), buat baris audit usulan terproses.
	var usulanID string
	err = tx.QueryRow(ctx,
		`SELECT id FROM usulan_unblock WHERE attempt_id = $1 AND status = 'menunggu'
		 ORDER BY created_at DESC LIMIT 1`, body.AttemptID).Scan(&usulanID)
	if err == nil {
		_, _ = tx.Exec(ctx,
			`UPDATE usulan_unblock SET status = 'disetujui', diproses_oleh = $2, diproses_at = now()
			 WHERE id = $1`, usulanID, uid)
	} else {
		_, _ = tx.Exec(ctx,
			`INSERT INTO usulan_unblock (attempt_id, pengawas_id, alasan, status, diproses_oleh, diproses_at)
			 VALUES ($1, $2, $3, 'disetujui', $2, now())`, body.AttemptID, uid, body.Alasan)
	}
	if err := tx.Commit(ctx); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal unblock attempt.")
	}
	auditCatat(ctx, pool, uid, "jadwal.unblock", "attempt", body.AttemptID,
		map[string]any{"jadwal_id": jadwalID, "alasan": body.Alasan, "sisa_waktu_detik": sisa}, c.RealIP())
	baru, _ := getAttempt(ctx, pool, body.AttemptID)
	broadcastAttempt(ctx, pool, baru)
	return writeOK(c, map[string]any{
		"attempt_id": body.AttemptID, "status": "berjalan",
		"sisa_waktu_detik": sisa, "deadline": baru.Deadline,
	})
}

// ---------------------------------------------------------------------------
// Ujian siswa
// ---------------------------------------------------------------------------

// kelasSiswa mencari kelas siswa dari tabel siswa (Tim 2).
func kelasSiswa(ctx context.Context, pool *pgxpool.Pool, siswaID string) (string, bool) {
	var kelasID string
	if err := pool.QueryRow(ctx, `SELECT kelas_id FROM siswa WHERE id = $1`, siswaID).Scan(&kelasID); err == nil && kelasID != "" {
		return kelasID, true
	}
	// Fallback: Tim 2 mungkin mengaitkan siswa ke akun via user_id.
	if err := pool.QueryRow(ctx, `SELECT kelas_id FROM siswa WHERE user_id = $1`, siswaID).Scan(&kelasID); err == nil && kelasID != "" {
		return kelasID, true
	}
	return "", false
}

func ujianAktif(c *echo.Context) error {
	_, _, siswaID, ok := requireRole(c, "siswa")
	if !ok {
		return nil
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	kelasID, ada := kelasSiswa(ctx, pool, siswaID)
	if !ada {
		return writeOK(c, map[string]any{
			"items": []any{}, "kelas_tidak_ditemukan": true,
			"pesan": "Kelas siswa tidak ditemukan. Hubungi admin/proktor.",
		})
	}
	rows, err := pool.Query(ctx,
		`SELECT j.id, j.paket_id, j.tanggal::text, j.jam_mulai::text, j.jam_selesai::text,
		        j.status, p.kode, p.nama, p.jenis, p.durasi_menit, p.tampil_hasil,
		        COALESCE(k.nama, ''),
		        a.id, a.status
		 FROM jadwal_ujian j
		 JOIN paket_ujian p ON p.id = j.paket_id
		 LEFT JOIN kelas k ON k.id = j.kelas_id
		 LEFT JOIN attempt a ON a.jadwal_id = j.id AND a.siswa_id = $1 AND a.status IN ('berjalan','diblokir','selesai')
		 WHERE j.kelas_id = $2
		   AND j.tanggal = (now() AT TIME ZONE 'Asia/Jakarta')::date
		   AND j.status IN ('draft', 'aktif', 'pause')
		 ORDER BY j.jam_mulai ASC`, siswaID, kelasID)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca jadwal aktif.")
	}
	defer rows.Close()
	type item struct {
		JadwalID      string  `json:"jadwal_id"`
		PaketID       string  `json:"paket_id"`
		Tanggal       string  `json:"tanggal"`
		JamMulai      string  `json:"jam_mulai"`
		JamSelesai    string  `json:"jam_selesai"`
		Status        string  `json:"status"`
		PaketKode     string  `json:"paket_kode"`
		PaketNama     string  `json:"paket_nama"`
		PaketJenis    string  `json:"paket_jenis"`
		DurasiMenit   int     `json:"durasi_menit"`
		TampilHasil   bool    `json:"tampil_hasil"`
		KelasNama     string  `json:"kelas_nama"`
		AttemptID     *string `json:"attempt_id"`
		AttemptStatus *string `json:"attempt_status"`
	}
	items := []item{}
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.JadwalID, &it.PaketID, &it.Tanggal, &it.JamMulai, &it.JamSelesai,
			&it.Status, &it.PaketKode, &it.PaketNama, &it.PaketJenis, &it.DurasiMenit,
			&it.TampilHasil,
			&it.KelasNama, &it.AttemptID, &it.AttemptStatus); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca jadwal aktif.")
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca jadwal aktif.")
	}
	return writeOK(c, map[string]any{"items": items})
}

func mulaiUjian(c *echo.Context) error {
	uid, _, siswaID, ok := requireRole(c, "siswa")
	if !ok {
		return nil
	}
	var body struct {
		JadwalID string `json:"jadwal_id"`
		Token    string `json:"token"`
	}
	if err := c.Bind(&body); err != nil || body.JadwalID == "" {
		return writeError(c, http.StatusBadRequest, "validation_error", "jadwal_id wajib diisi.")
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	j, err := getJadwal(ctx, pool, body.JadwalID)
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
	}
	// Siswa hanya boleh ikut jadwal kelasnya.
	if kelasID, ada := kelasSiswa(ctx, pool, siswaID); ada && kelasID != j.KelasID {
		return writeError(c, http.StatusForbidden, "forbidden", "Jadwal ini bukan untuk kelas Anda.")
	}
	if j.Status == "selesai" {
		return writeError(c, http.StatusConflict, "conflict", "Jadwal sudah selesai.")
	}
	if j.Status == "pause" {
		return writeError(c, http.StatusConflict, "conflict", "Ujian sedang dijeda oleh proktor. Tunggu hingga dilanjutkan.")
	}
	// Attempt berjalan yang sudah ada -> kembalikan (resume, idempotent).
	if lama, err := scanAttempt(pool.QueryRow(ctx,
		`SELECT id, jadwal_id, siswa_id, status, mulai_at, deadline,
		        sisa_waktu_beku, urutan_soal, urutan_opsi, submit_at,
		        nilai_akhir, created_at
		 FROM attempt WHERE jadwal_id = $1 AND siswa_id = $2 AND status = 'berjalan'
		 ORDER BY created_at DESC LIMIT 1`, body.JadwalID, siswaID)); err == nil {
		return writeOK(c, map[string]any{
			"attempt_id": lama.ID, "jadwal_id": lama.JadwalID, "status": lama.Status,
			"deadline": lama.Deadline, "sudah_ada": true,
			"sisa_waktu_detik": sisaWaktu(lama, j.Status, time.Now()),
		})
	}
	if ada, err := scanAttempt(pool.QueryRow(ctx,
		`SELECT id, jadwal_id, siswa_id, status, mulai_at, deadline,
		        sisa_waktu_beku, urutan_soal, urutan_opsi, submit_at,
		        nilai_akhir, created_at
		 FROM attempt WHERE jadwal_id = $1 AND siswa_id = $2 AND status = 'diblokir'
		 ORDER BY created_at DESC LIMIT 1`, body.JadwalID, siswaID)); err == nil {
		_ = ada
		return writeError(c, http.StatusForbidden, "forbidden", "Attempt Anda diblokir. Hubungi proktor untuk unblock.")
	}
	if _, err := scanAttempt(pool.QueryRow(ctx,
		`SELECT id, jadwal_id, siswa_id, status, mulai_at, deadline,
		        sisa_waktu_beku, urutan_soal, urutan_opsi, submit_at,
		        nilai_akhir, created_at
		 FROM attempt WHERE jadwal_id = $1 AND siswa_id = $2 AND status = 'selesai'
		 ORDER BY created_at DESC LIMIT 1`, body.JadwalID, siswaID)); err == nil {
		return writeError(c, http.StatusConflict, "conflict", "Anda sudah menyelesaikan ujian ini.")
	}

	mulai, err := gabungTanggalJam(j.Tanggal, j.JamMulai)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Tanggal/jam jadwal tidak valid.")
	}
	selesai, err := gabungTanggalJam(j.Tanggal, j.JamSelesai)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Tanggal/jam jadwal tidak valid.")
	}
	now := time.Now()
	if now.Before(mulai) {
		return writeError(c, http.StatusConflict, "conflict", "Ujian belum dimulai.")
	}
	if now.After(selesai) {
		return writeError(c, http.StatusConflict, "conflict", "Ujian sudah berakhir.")
	}
	// Toleransi telat: hanya untuk MEMULAI (attempt baru).
	if now.After(mulai.Add(time.Duration(toleransiTelatMenit(ctx, pool)) * time.Minute)) {
		return writeError(c, http.StatusForbidden, "forbidden", "Batas toleransi keterlambatan habis. Anda tidak bisa memulai ujian ini.")
	}
	// Token (kecuali Tryout).
	if j.PaketJenis != "Tryout" {
		token := strings.TrimSpace(body.Token)
		if token == "" || j.TokenHash == nil {
			return writeError(c, http.StatusUnauthorized, "token_invalid", "Token ujian wajib diisi.")
		}
		if hashToken(token) != *j.TokenHash {
			return writeError(c, http.StatusUnauthorized, "token_invalid", "Token ujian tidak cocok.")
		}
		if j.TokenExpiresAt != nil {
			if exp, ok := (*j.TokenExpiresAt).(time.Time); ok && now.After(exp) {
				return writeError(c, http.StatusConflict, "token_expired", "Token ujian sudah kedaluwarsa.")
			}
		}
	}

	// Susun soal paket (urutan tampil per attempt).
	rows, err := pool.Query(ctx,
		`SELECT s.id, s.tipe, s.jawaban FROM paket_soal ps JOIN soal s ON s.id = ps.soal_id
		 WHERE ps.paket_id = $1 ORDER BY ps.urutan ASC`, j.PaketID)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca soal paket.")
	}
	soalIDs := []string{}
	tampilanPerSoal := map[string][]byte{}
	tipePerSoal := map[string]string{}
	for rows.Next() {
		var id, tipe string
		var tampilan []byte
		if err := rows.Scan(&id, &tipe, &tampilan); err == nil {
			soalIDs = append(soalIDs, id)
			tampilanPerSoal[id] = tampilan
			tipePerSoal[id] = tipe
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca soal paket.")
	}
	if len(soalIDs) == 0 {
		return writeError(c, http.StatusConflict, "conflict", "Paket ujian belum berisi soal.")
	}

	attemptID := newUUID()
	seed := seedDariAttemptID(attemptID)
	urutanSoal := soalIDs
	if j.AcakSoal {
		urutanSoal = shuffleSeed(soalIDs, seed)
	}
	// urutan_opsi menyimpan TAMPILAN soal per attempt (soal.jawaban yang
	// sudah dikocok bila acak_jawaban aktif). Bila acak mati, tampilan
	// asli disimpan apa adanya agar GET attempt stabil.
	urutanOpsi := map[string]json.RawMessage{}
	for i, soalID := range urutanSoal {
		if j.AcakJawaban {
			urutanOpsi[soalID] = shuffleTampilan(tampilanPerSoal[soalID], tipePerSoal[soalID], seed+int64(i+1)*1009)
		} else {
			urutanOpsi[soalID] = tampilanPerSoal[soalID]
		}
	}
	urutanSoalJSON, _ := json.Marshal(urutanSoal)
	urutanOpsiJSON, _ := json.Marshal(urutanOpsi)

	tx, err := pool.Begin(ctx)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memulai ujian.")
	}
	defer tx.Rollback(ctx)
	// Jadwal draft -> aktif saat peserta pertama mulai (sesuai tugas:
	// validasi "jadwal aktif/draft→aktif").
	if j.Status == "draft" {
		if _, err := tx.Exec(ctx, `UPDATE jadwal_ujian SET status = 'aktif' WHERE id = $1`, j.ID); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mengaktifkan jadwal.")
		}
	}
	// deadline = jadwal_selesai (ujian berakhir jam selesai jadwal).
	if _, err := tx.Exec(ctx,
		`INSERT INTO attempt (id, jadwal_id, siswa_id, status, mulai_at, deadline,
		                       urutan_soal, urutan_opsi)
		 VALUES ($1, $2, $3, 'berjalan', now(), $4, $5, $6)`,
		attemptID, j.ID, siswaID, selesai, urutanSoalJSON, urutanOpsiJSON); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membuat attempt ujian.")
	}
	if err := tx.Commit(ctx); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memulai ujian.")
	}
	auditCatat(ctx, pool, uid, "ujian.mulai", "attempt", attemptID,
		map[string]any{"jadwal_id": j.ID}, c.RealIP())
	monitor.Broadcast(j.ID, map[string]any{"type": "ujian_mulai", "jadwal_id": j.ID, "attempt_id": attemptID})
	return writeCreated(c, map[string]any{
		"attempt_id": attemptID, "jadwal_id": j.ID, "status": "berjalan",
		"deadline":         selesai.Format(time.RFC3339),
		"sisa_waktu_detik": int(time.Until(selesai).Seconds()),
		"jumlah_soal":      len(soalIDs),
	})
}

// muatSoalAttempt mengembalikan soal paket untuk attempt, TANPA kunci &
// pembahasan, dalam urutan attempt (urutan_soal) + opsi dalam urutan
// attempt (urutan_opsi). Disertai jawaban tersimpan + ragu.
func muatSoalAttempt(ctx context.Context, pool *pgxpool.Pool, a attempt, j jadwal) ([]map[string]any, error) {
	rows, err := pool.Query(ctx,
		`SELECT s.id, s.tipe, s.pertanyaan, s.jawaban, s.bobot, s.stimulus, s.media_url
		 FROM paket_soal ps JOIN soal s ON s.id = ps.soal_id
		 WHERE ps.paket_id = $1 ORDER BY ps.urutan ASC`, j.PaketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type soalRow struct {
		ID         string
		Tipe       string
		Pertanyaan string
		Tampilan   []byte
		Bobot      int
		Stimulus   string
		MediaURL   string
	}
	byID := map[string]soalRow{}
	var semua []string
	for rows.Next() {
		var s soalRow
		if err := rows.Scan(&s.ID, &s.Tipe, &s.Pertanyaan, &s.Tampilan, &s.Bobot, &s.Stimulus, &s.MediaURL); err != nil {
			return nil, err
		}
		byID[s.ID] = s
		semua = append(semua, s.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Urutan tampil = urutan_soal attempt; soal baru (di luar urutan)
	// ditambahkan di akhir.
	urutan := []string{}
	terlihat := map[string]bool{}
	for _, id := range a.UrutanSoal {
		if _, ada := byID[id]; ada {
			urutan = append(urutan, id)
			terlihat[id] = true
		}
	}
	for _, id := range semua {
		if !terlihat[id] {
			urutan = append(urutan, id)
		}
	}
	// Jawaban tersimpan.
	jawabanRows, err := pool.Query(ctx,
		`SELECT soal_id, jawaban, ragu FROM jawaban WHERE attempt_id = $1`, a.ID)
	if err != nil {
		return nil, err
	}
	defer jawabanRows.Close()
	jawabanMap := map[string]json.RawMessage{}
	raguMap := map[string]bool{}
	for jawabanRows.Next() {
		var soalID string
		var jwb []byte
		var ragu bool
		if err := jawabanRows.Scan(&soalID, &jwb, &ragu); err == nil {
			jawabanMap[soalID] = jwb
			raguMap[soalID] = ragu
		}
	}
	if err := jawabanRows.Err(); err != nil {
		return nil, err
	}
	out := []map[string]any{}
	for i, id := range urutan {
		s := byID[id]
		// Tampilan = soal.jawaban (data tampil Tim 3) dalam urutan
		// attempt; bila tidak tersimpan (attempt lama), pakai asli.
		// KUNCI TIDAK PERNAH disertakan (kolom kunci tidak dipilih).
		tampilan := s.Tampilan
		if tersimpan, ada := a.UrutanOpsi[id]; ada && len(tersimpan) > 0 {
			tampilan = tersimpan
		}
		var tampilanAny any
		if len(tampilan) > 0 {
			_ = json.Unmarshal(tampilan, &tampilanAny)
		}
		item := map[string]any{
			"soal_id": s.ID, "nomor": i + 1, "tipe": s.Tipe,
			"pertanyaan": s.Pertanyaan, "stimulus": s.Stimulus,
			"media_url": s.MediaURL,
			"tampilan":  tampilanAny, "bobot": s.Bobot,
			"jawaban": jawabanMap[id], "ragu": raguMap[id],
		}
		out = append(out, item)
	}
	return out, nil
}

func milikSendiri(c *echo.Context, a attempt, siswaID string) bool {
	if a.SiswaID != siswaID {
		_ = writeError(c, http.StatusNotFound, "not_found", "Attempt tidak ditemukan.")
		return false
	}
	return true
}

func getAttemptDetail(c *echo.Context) error {
	_, _, siswaID, ok := requireRole(c, "siswa")
	if !ok {
		return nil
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	a, err := getAttempt(ctx, pool, c.Param("id"))
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Attempt tidak ditemukan.")
	}
	if !milikSendiri(c, a, siswaID) {
		return nil
	}
	a, _ = autoSubmitJikaLewat(ctx, pool, a, jadwalStatusUntuk(c, pool, a))
	j, err := getJadwal(ctx, pool, a.JadwalID)
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
	}
	soal, err := muatSoalAttempt(ctx, pool, a, j)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca soal ujian.")
	}
	resp := map[string]any{
		"attempt_id": a.ID, "jadwal_id": a.JadwalID, "status": a.Status,
		"deadline": a.Deadline, "mulai_at": a.MulaiAt,
		"sisa_waktu_detik": sisaWaktu(a, j.Status, time.Now()),
		"jadwal_status":    j.Status, "paket_jenis": j.PaketJenis,
		"server_now": time.Now().Format(time.RFC3339),
		"soal":       soal,
	}
	if a.Status == "selesai" && tampilHasilEfektif(ctx, pool, j.TampilHasil) {
		nilai, menunggu, err := hitungNilai(ctx, pool, a.ID)
		if err == nil {
			resp["nilai_akhir"] = nilai
			resp["menunggu_koreksi"] = menunggu
		}
	}
	return writeOK(c, resp)
}

func simpanJawaban(c *echo.Context) error {
	uid, _, siswaID, ok := requireRole(c, "siswa")
	if !ok {
		return nil
	}
	var body struct {
		SoalID  string          `json:"soal_id"`
		Jawaban json.RawMessage `json:"jawaban"`
	}
	if err := c.Bind(&body); err != nil || body.SoalID == "" {
		return writeError(c, http.StatusBadRequest, "validation_error", "soal_id dan jawaban wajib diisi.")
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	a, err := getAttempt(ctx, pool, c.Param("id"))
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Attempt tidak ditemukan.")
	}
	if !milikSendiri(c, a, siswaID) {
		return nil
	}
	if a.Status == "diblokir" {
		return writeError(c, http.StatusForbidden, "forbidden", "Attempt diblokir. Hubungi proktor.")
	}
	if a.Status == "selesai" {
		return writeError(c, http.StatusConflict, "conflict", "Ujian sudah selesai (submit). Jawaban tidak bisa diubah.")
	}
	a, _ = autoSubmitJikaLewat(ctx, pool, a, jadwalStatusUntuk(c, pool, a))
	if a.Status != "berjalan" {
		return writeError(c, http.StatusConflict, "conflict", "Batas waktu ujian habis. Jawaban tidak bisa disimpan.")
	}
	j, err := getJadwal(ctx, pool, a.JadwalID)
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
	}
	if j.Status == "pause" {
		return writeError(c, http.StatusConflict, "conflict", "Ujian sedang dijeda. Jawaban tidak bisa disimpan sampai dilanjutkan.")
	}
	// Soal harus bagian dari paket jadwal ini.
	var soalAda bool
	_ = pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM paket_soal WHERE paket_id = $1 AND soal_id = $2)`,
		j.PaketID, body.SoalID).Scan(&soalAda)
	if !soalAda {
		return writeError(c, http.StatusBadRequest, "validation_error", "Soal bukan bagian dari paket ujian ini.")
	}
	// Ambil tipe + kunci + bobot + tampilan soal untuk penskoran otomatis.
	// (Tampilan dibutuhkan grading urutkan bila jawaban berupa string item.)
	var tipe string
	var kunci, bobotRaw, tampilan []byte
	var bobot int
	if err := pool.QueryRow(ctx,
		`SELECT tipe, kunci, bobot::text, jawaban FROM soal WHERE id = $1`, body.SoalID).
		Scan(&tipe, &kunci, &bobotRaw, &tampilan); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca soal.")
	}
	if f, err := strconv.ParseFloat(strings.TrimSpace(string(bobotRaw)), 64); err == nil {
		bobot = int(f)
	}
	var raguLama bool
	_ = pool.QueryRow(ctx,
		`SELECT ragu FROM jawaban WHERE attempt_id = $1 AND soal_id = $2`,
		a.ID, body.SoalID).Scan(&raguLama)
	// Penskoran: otomatis kecuali uraian/manual.
	var poin *float64
	dikoreksi := false
	kanon := grading.NormalisasiTipe(tipe)
	if kanon == "uraian" {
		poin = nil // manual: menunggu koreksi guru.
	} else if !grading.Dinilai(tipe, kunci) {
		nol := 0.0
		poin = &nol
		dikoreksi = true // tidak dinilai -> tidak menunggu koreksi.
	} else {
		p, auto := grading.HitungPoin(tipe, body.Jawaban, kunciUntukGrading(kunci, tampilan, tipe), bobot)
		if !auto {
			poin = nil
		} else {
			poin = &p
			dikoreksi = true // auto sudah final per soal.
		}
	}
	var poinArg any
	if poin != nil {
		poinArg = *poin
	}
	_, err = pool.Exec(ctx,
		`INSERT INTO jawaban (attempt_id, soal_id, jawaban, ragu, poin, dikoreksi, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, now())
		 ON CONFLICT (attempt_id, soal_id) DO UPDATE SET
		     jawaban = EXCLUDED.jawaban, poin = EXCLUDED.poin,
		     dikoreksi = EXCLUDED.dikoreksi, updated_at = now()`,
		a.ID, body.SoalID, body.Jawaban, raguLama, poinArg, dikoreksi)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menyimpan jawaban.")
	}
	// Audit (Fase 3): simpan jawaban adalah mutasi — sebelumnya tidak
	// dicatat. Detail TIDAK memuat isi jawaban (hemat audit log; isi
	// jawaban sudah tersimpan di tabel jawaban).
	auditCatat(ctx, pool, uid, "ujian.jawaban", "jawaban", body.SoalID,
		map[string]any{"attempt_id": a.ID, "soal_id": body.SoalID}, c.RealIP())
	broadcastAttempt(ctx, pool, a)
	return writeOK(c, map[string]any{
		"tersimpan": true, "soal_id": body.SoalID,
		"server_now": time.Now().Format(time.RFC3339),
	})
}

func tandaiRagu(c *echo.Context) error {
	uid, _, siswaID, ok := requireRole(c, "siswa")
	if !ok {
		return nil
	}
	var body struct {
		SoalID string `json:"soal_id"`
		Ragu   bool   `json:"ragu"`
	}
	if err := c.Bind(&body); err != nil || body.SoalID == "" {
		return writeError(c, http.StatusBadRequest, "validation_error", "soal_id dan ragu wajib diisi.")
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	a, err := getAttempt(ctx, pool, c.Param("id"))
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Attempt tidak ditemukan.")
	}
	if !milikSendiri(c, a, siswaID) {
		return nil
	}
	if a.Status != "berjalan" {
		return writeError(c, http.StatusConflict, "conflict", "Attempt tidak sedang berjalan.")
	}
	a, _ = autoSubmitJikaLewat(ctx, pool, a, jadwalStatusUntuk(c, pool, a))
	if a.Status != "berjalan" {
		return writeError(c, http.StatusConflict, "conflict", "Batas waktu ujian habis.")
	}
	_, err = pool.Exec(ctx,
		`INSERT INTO jawaban (attempt_id, soal_id, jawaban, ragu)
		 VALUES ($1, $2, NULL, $3)
		 ON CONFLICT (attempt_id, soal_id) DO UPDATE SET ragu = EXCLUDED.ragu, updated_at = now()`,
		a.ID, body.SoalID, body.Ragu)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menandai ragu-ragu.")
	}
	// Audit (Fase 3): tandai ragu-ragu adalah mutasi — sebelumnya tidak dicatat.
	auditCatat(ctx, pool, uid, "ujian.ragu", "jawaban", body.SoalID,
		map[string]any{"attempt_id": a.ID, "soal_id": body.SoalID, "ragu": body.Ragu}, c.RealIP())
	broadcastAttempt(ctx, pool, a)
	return writeOK(c, map[string]any{"soal_id": body.SoalID, "ragu": body.Ragu})
}

func catatPelanggaran(c *echo.Context) error {
	uid, _, siswaID, ok := requireRole(c, "siswa")
	if !ok {
		return nil
	}
	var body struct {
		Jenis  string `json:"jenis"`
		Detail string `json:"detail"`
	}
	if err := c.Bind(&body); err != nil || strings.TrimSpace(body.Jenis) == "" {
		return writeError(c, http.StatusBadRequest, "validation_error", "jenis pelanggaran wajib diisi.")
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	a, err := getAttempt(ctx, pool, c.Param("id"))
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Attempt tidak ditemukan.")
	}
	if !milikSendiri(c, a, siswaID) {
		return nil
	}
	if a.Status == "selesai" {
		return writeError(c, http.StatusConflict, "conflict", "Ujian sudah selesai.")
	}
	if a.Status == "diblokir" {
		return writeOK(c, map[string]any{"diblokir": true, "pesan": "Attempt sudah diblokir. Hubungi proktor."})
	}
	// Deduplikasi: abaikan event berulang dalam 5 detik (blur +
	// visibilitychange sering berapi hampir bersamaan).
	var terakhir *time.Time
	_ = pool.QueryRow(ctx,
		`SELECT created_at FROM pelanggaran WHERE attempt_id = $1
		 ORDER BY created_at DESC LIMIT 1`, a.ID).Scan(&terakhir)
	var jumlah int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM pelanggaran WHERE attempt_id = $1`, a.ID).Scan(&jumlah)
	if terakhir != nil && time.Since(*terakhir) < 5*time.Second {
		return writeOK(c, map[string]any{
			"diabaikan_duplikat": true, "jumlah_pelanggaran": jumlah,
			"peringatan": jumlah == 1, "diblokir": false,
		})
	}
	_, err = pool.Exec(ctx,
		`INSERT INTO pelanggaran (attempt_id, jenis, detail) VALUES ($1, $2, NULLIF($3, ''))`,
		a.ID, body.Jenis, body.Detail)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal mencatat pelanggaran.")
	}
	jumlah++
	auditCatat(ctx, pool, uid, "ujian.pelanggaran", "attempt", a.ID,
		map[string]any{"jenis": body.Jenis, "jumlah": jumlah}, c.RealIP())
	batas := maxPelanggaran(ctx, pool)
	if jumlah >= batas {
		// Auto-block PER-ATTEMPT. Bekukan sisa waktu agar unblock bisa
		// melanjutkan dari sisa yang sama.
		_, _ = pool.Exec(ctx,
			`UPDATE attempt SET status = 'diblokir',
			        sisa_waktu_beku = GREATEST(0, EXTRACT(EPOCH FROM (deadline - now()))::INT)
			 WHERE id = $1`, a.ID)
		a.Status = "diblokir"
		broadcastAttempt(ctx, pool, a)
		return writeOK(c, map[string]any{
			"diblokir": true, "jumlah_pelanggaran": jumlah,
			"pesan": "Anda diblokir karena pelanggaran berulang. Hubungi proktor.",
		})
	}
	broadcastAttempt(ctx, pool, a)
	return writeOK(c, map[string]any{
		"peringatan": true, "diblokir": false, "jumlah_pelanggaran": jumlah,
		"pesan": "Peringatan: terdeteksi meninggalkan halaman ujian. Pelanggaran berikutnya akan memblokir ujian Anda.",
	})
}

func submitUjian(c *echo.Context) error {
	uid, _, siswaID, ok := requireRole(c, "siswa")
	if !ok {
		return nil
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	a, err := getAttempt(ctx, pool, c.Param("id"))
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Attempt tidak ditemukan.")
	}
	if !milikSendiri(c, a, siswaID) {
		return nil
	}
	if a.Status == "diblokir" {
		return writeError(c, http.StatusForbidden, "forbidden", "Attempt diblokir. Hubungi proktor.")
	}
	if a.Status == "selesai" {
		// Idempotent: kembalikan hasil tersimpan.
		j, _ := getJadwal(ctx, pool, a.JadwalID)
		resp := map[string]any{"attempt_id": a.ID, "status": "selesai", "submit_at": a.SubmitAt}
		if tampilHasilEfektif(ctx, pool, j.TampilHasil) && a.NilaiAkhir != nil {
			resp["nilai_akhir"] = *a.NilaiAkhir
		}
		return writeOK(c, resp)
	}
	nilai, menunggu, err := finalisasiAttempt(ctx, pool, a)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal submit ujian.")
	}
	j, err := getJadwal(ctx, pool, a.JadwalID)
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
	}
	auditCatat(ctx, pool, uid, "ujian.submit", "attempt", a.ID,
		map[string]any{"jadwal_id": a.JadwalID, "nilai_akhir": nilai, "menunggu_koreksi": menunggu}, c.RealIP())
	broadcastAttempt(ctx, pool, a)
	resp := map[string]any{
		"attempt_id": a.ID, "status": "selesai", "menunggu_koreksi": menunggu,
	}
	// Nilai disembunyikan hingga final bila tampil_hasil false.
	if tampilHasilEfektif(ctx, pool, j.TampilHasil) {
		resp["nilai_akhir"] = nilai
	} else {
		resp["nilai_akhir"] = nil
		resp["pesan"] = "Nilai disembunyikan sampai koreksi final."
	}
	return writeOK(c, resp)
}

func statusAttempt(c *echo.Context) error {
	_, _, siswaID, ok := requireRole(c, "siswa")
	if !ok {
		return nil
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	a, err := getAttempt(ctx, pool, c.Param("id"))
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Attempt tidak ditemukan.")
	}
	if !milikSendiri(c, a, siswaID) {
		return nil
	}
	a, _ = autoSubmitJikaLewat(ctx, pool, a, jadwalStatusUntuk(c, pool, a))
	j, err := getJadwal(ctx, pool, a.JadwalID)
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
	}
	var dijawab, ragu, total int
	_ = pool.QueryRow(ctx,
		`SELECT (SELECT COUNT(*) FROM jawaban WHERE attempt_id = $1 AND jawaban IS NOT NULL),
		        (SELECT COUNT(*) FROM jawaban WHERE attempt_id = $1 AND ragu = TRUE),
		        (SELECT COUNT(*) FROM paket_soal WHERE paket_id = $2)`,
		a.ID, j.PaketID).Scan(&dijawab, &ragu, &total)
	resp := map[string]any{
		"attempt_id": a.ID, "jadwal_id": a.JadwalID, "status": a.Status,
		"deadline": a.Deadline, "jadwal_status": j.Status,
		"sisa_waktu_detik": sisaWaktu(a, j.Status, time.Now()),
		"dijawab":          dijawab, "ragu": ragu, "total_soal": total,
		"belum_dijawab": total - dijawab,
		"server_now":    time.Now().Format(time.RFC3339),
	}
	if a.Status == "selesai" {
		if tampilHasilEfektif(ctx, pool, j.TampilHasil) && a.NilaiAkhir != nil {
			resp["nilai_akhir"] = *a.NilaiAkhir
		} else {
			resp["nilai_akhir"] = nil
		}
		_, menunggu, err := hitungNilai(ctx, pool, a.ID)
		if err == nil {
			resp["menunggu_koreksi"] = menunggu
		}
	}
	return writeOK(c, resp)
}

// kunciTampil merangkum kunci benar untuk halaman hasil (Tryout /
// tampil_hasil). Hanya dipakai SETELAH ujian selesai, tidak pernah
// sebelum/selama ujian. Mengembalikan teks ringkas, bukan kunci mentah.
func kunciTampil(tipe string, kunci, tampilan []byte) string {
	kanon := grading.NormalisasiTipe(tipe)
	switch kanon {
	case "pg":
		if opsi, ok := func() ([]string, bool) {
			var obj map[string]json.RawMessage
			if err := json.Unmarshal(kunci, &obj); err == nil {
				if raw, ada := obj["opsi"]; ada {
					var s string
					if err := json.Unmarshal(raw, &s); err == nil {
						return []string{s}, true
					}
				}
			}
			return nil, false
		}(); ok && len(opsi) > 0 {
			// Cari teks opsi dari tampilan.
			var tamp map[string]json.RawMessage
			if err := json.Unmarshal(tampilan, &tamp); err == nil {
				var opsiList []struct {
					Label string `json:"label"`
					Teks  string `json:"teks"`
				}
				if err := json.Unmarshal(tamp["opsi"], &opsiList); err == nil {
					for _, o := range opsiList {
						if o.Label == opsi[0] {
							return "Jawaban benar: " + o.Label + " — " + o.Teks
						}
					}
				}
			}
			return "Jawaban benar: " + opsi[0]
		}
		return ""
	case "pg_kompleks":
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(kunci, &obj); err == nil {
			var arr []string
			if err := json.Unmarshal(obj["opsi"], &arr); err == nil {
				return "Jawaban benar: " + strings.Join(arr, ", ")
			}
		}
		return ""
	case "benar_salah":
		var obj map[string]json.RawMessage
		var nilai []bool
		if err := json.Unmarshal(kunci, &obj); err == nil {
			_ = json.Unmarshal(obj["nilai"], &nilai)
		}
		var tamp map[string]json.RawMessage
		var pernyataan []string
		if err := json.Unmarshal(tampilan, &tamp); err == nil {
			_ = json.Unmarshal(tamp["pernyataan"], &pernyataan)
		}
		parts := []string{}
		for i, v := range nilai {
			label := ""
			if i < len(pernyataan) {
				label = pernyataan[i]
			}
			bs := "Salah"
			if v {
				bs = "Benar"
			}
			parts = append(parts, label+" = "+bs)
		}
		return strings.Join(parts, "; ")
	case "menjodohkan":
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(kunci, &obj); err == nil {
			var pasangan []struct {
				Kiri  string `json:"kiri"`
				Kanan string `json:"kanan"`
			}
			if err := json.Unmarshal(obj["pasangan"], &pasangan); err == nil {
				parts := []string{}
				for _, p := range pasangan {
					parts = append(parts, p.Kiri+" → "+p.Kanan)
				}
				return "Pasangan benar: " + strings.Join(parts, "; ")
			}
		}
		return ""
	case "isian_singkat", "isi_ganda", "teks_rumpang":
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(kunci, &obj); err == nil {
			var arr []string
			if err := json.Unmarshal(obj["jawaban"], &arr); err == nil {
				return "Jawaban yang diterima: " + strings.Join(arr, " | ")
			}
		}
		return ""
	case "mengurutkan":
		var obj map[string]json.RawMessage
		var urutan []int
		if err := json.Unmarshal(kunci, &obj); err == nil {
			var nums []float64
			if err := json.Unmarshal(obj["urutan"], &nums); err == nil {
				for _, n := range nums {
					urutan = append(urutan, int(n))
				}
			}
		}
		var tamp map[string]json.RawMessage
		var item []string
		if err := json.Unmarshal(tampilan, &tamp); err == nil {
			_ = json.Unmarshal(tamp["item"], &item)
		}
		if len(urutan) > 0 && len(item) > 0 {
			parts := []string{}
			for _, u := range urutan {
				if u >= 0 && u < len(item) {
					parts = append(parts, item[u])
				}
			}
			return "Urutan benar: " + strings.Join(parts, " → ")
		}
		return ""
	case "uraian":
		return "Dinilai manual oleh guru (rubrik)."
	case "skala_sikap":
		return "Tidak dinilai (survei)."
	default:
		return ""
	}
}

// hasilAttempt mengembalikan rekap per-soal untuk halaman hasil siswa.
//
//	GET /api/v1/ujian/attempt/:id/hasil
//
// Hanya attempt milik sendiri & sudah selesai. Pembahasan (kolom
// soal.pembahasan, migrasi 0011) + kunci tampil hanya disertakan bila
// paket Tryout atau tampil_hasil efektif; bila tidak, hanya pertanyaan
// + jawaban siswa (tanpa benar/salah/poin/pembahasan). Pembahasan TIDAK
// PERNAH dikirim di respons attempt sebelum submit (muatSoalAttempt
// tidak memilih kolom pembahasan).
func hasilAttempt(c *echo.Context) error {
	_, _, siswaID, ok := requireRole(c, "siswa")
	if !ok {
		return nil
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	a, err := getAttempt(ctx, pool, c.Param("id"))
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Attempt tidak ditemukan.")
	}
	if !milikSendiri(c, a, siswaID) {
		return nil
	}
	if a.Status != "selesai" {
		return writeError(c, http.StatusConflict, "conflict", "Ujian belum selesai. Hasil tersedia setelah submit.")
	}
	j, err := getJadwal(ctx, pool, a.JadwalID)
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
	}
	tampil := tampilHasilEfektif(ctx, pool, j.TampilHasil) || j.PaketJenis == "Tryout"
	rows, err := pool.Query(ctx,
		`SELECT jw.soal_id, jw.jawaban, jw.poin, jw.dikoreksi, jw.ragu,
		        so.tipe, so.pertanyaan, so.jawaban, so.kunci, so.bobot, so.pembahasan
		 FROM jawaban jw JOIN soal so ON so.id = jw.soal_id
		 WHERE jw.attempt_id = $1`, a.ID)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca hasil ujian.")
	}
	defer rows.Close()
	type item struct {
		SoalID      string   `json:"soal_id"`
		Nomor       int      `json:"nomor"`
		Tipe        string   `json:"tipe"`
		Pertanyaan  string   `json:"pertanyaan"`
		Jawaban     any      `json:"jawaban_siswa"`
		Poin        *float64 `json:"poin"`
		Bobot       float64  `json:"bobot"`
		Ragu        bool     `json:"ragu"`
		Status      string   `json:"status"`
		KunciTampil string   `json:"kunci_tampil,omitempty"`
		// Pembahasan hanya diisi bila tampil (Tryout / tampil_hasil
		// efektif); omitempty menyembunyikannya bila kosong.
		Pembahasan string `json:"pembahasan,omitempty"`
	}
	bySoal := map[string]*item{}
	for rows.Next() {
		var it item
		var jwb, tampilan, kunci []byte
		var tipe, pembahasan string
		if err := rows.Scan(&it.SoalID, &jwb, &it.Poin, new(bool), &it.Ragu,
			&tipe, &it.Pertanyaan, &tampilan, &kunci, &it.Bobot, &pembahasan); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca hasil ujian.")
		}
		// Kolom dikoreksi tidak dipakai di sini (dihitung via poin nil).
		var jwbAny any
		if len(jwb) > 0 {
			_ = json.Unmarshal(jwb, &jwbAny)
		}
		it.Jawaban = jwbAny
		it.Tipe = tipe
		if tampil {
			it.KunciTampil = kunciTampil(tipe, kunci, tampilan)
			it.Pembahasan = pembahasan
			if it.Poin == nil {
				it.Status = "menunggu_koreksi"
			} else if *it.Poin >= it.Bobot {
				it.Status = "benar"
			} else if *it.Poin > 0 {
				it.Status = "sebagian"
			} else {
				it.Status = "salah"
			}
		} else {
			it.Poin = nil
			it.Status = "disembunyikan"
		}
		bySoal[it.SoalID] = &it
	}
	if err := rows.Err(); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca hasil ujian.")
	}
	// Urutkan sesuai urutan_soal attempt; soal tanpa jawaban ikut (kosong).
	items := []item{}
	terlihat := map[string]bool{}
	for i, soalID := range a.UrutanSoal {
		if it, ada := bySoal[soalID]; ada {
			it.Nomor = i + 1
			items = append(items, *it)
			terlihat[soalID] = true
		}
	}
	for soalID, it := range bySoal {
		if !terlihat[soalID] {
			it.Nomor = len(items) + 1
			items = append(items, *it)
		}
	}
	// Soal paket tanpa baris jawaban (tidak dijawab sama sekali):
	// lengkapi soal yang belum ada di jawaban (poin 0, jawaban null).
	sudah := map[string]bool{}
	for _, it := range items {
		sudah[it.SoalID] = true
	}
	rows2, err := pool.Query(ctx,
		`SELECT s.id, s.tipe, s.pertanyaan, s.bobot, s.pembahasan
		 FROM paket_soal ps JOIN soal s ON s.id = ps.soal_id
		 WHERE ps.paket_id = $1 ORDER BY ps.urutan ASC`, j.PaketID)
	if err == nil {
		defer rows2.Close()
		nomor := len(items)
		for rows2.Next() {
			var id, tipe, pertanyaan, pembahasan string
			var bobot float64
			if err := rows2.Scan(&id, &tipe, &pertanyaan, &bobot, &pembahasan); err != nil {
				continue
			}
			if sudah[id] {
				continue
			}
			nomor++
			it := item{SoalID: id, Nomor: nomor, Tipe: tipe, Pertanyaan: pertanyaan,
				Bobot: bobot, Status: "disembunyikan"}
			if tampil {
				it.Status = "salah"
				nol := 0.0
				it.Poin = &nol
				it.Pembahasan = pembahasan
			}
			items = append(items, it)
		}
	}
	_, menunggu, _ := hitungNilai(ctx, pool, a.ID)
	resp := map[string]any{
		"attempt_id": a.ID, "jadwal_id": a.JadwalID, "paket_jenis": j.PaketJenis,
		"tampil_hasil": tampil, "menunggu_koreksi": menunggu, "items": items,
	}
	if tampil && a.NilaiAkhir != nil {
		resp["nilai_akhir"] = *a.NilaiAkhir
	} else {
		resp["nilai_akhir"] = nil
	}
	return writeOK(c, resp)
}

// ---------------------------------------------------------------------------
// Koreksi
// ---------------------------------------------------------------------------

func listKoreksi(c *echo.Context) error {
	if _, _, _, ok := requireRole(c, "admin", "guru"); !ok {
		return nil
	}
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	jadwalID := c.Param("jadwalId")
	if !isUUID(jadwalID) {
		return writeError(c, http.StatusBadRequest, "validation_error", "ID jadwal tidak valid.")
	}
	if _, err := getJadwal(ctx, pool, jadwalID); err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Jadwal tidak ditemukan.")
	}
	rows, err := pool.Query(ctx,
		`SELECT jw.id, jw.attempt_id, jw.soal_id, jw.jawaban, jw.ragu,
		        jw.poin, jw.dikoreksi, jw.updated_at,
		        a.siswa_id, a.status,
		        COALESCE(s.nama, ''), COALESCE(s.nisn, ''),
		        so.tipe, so.pertanyaan, so.bobot
		 FROM jawaban jw
		 JOIN attempt a ON a.id = jw.attempt_id
		 JOIN soal so ON so.id = jw.soal_id
		 LEFT JOIN siswa s ON s.user_id = a.siswa_id
		 WHERE a.jadwal_id = $1
		   AND jw.dikoreksi = FALSE
		   AND jw.jawaban IS NOT NULL
		   AND so.tipe IN ('uraian','upload_file','praktik_singkat',
		                    'esai','essay','isian_panjang','isi_panjang',
		                    'esai_singkat','uraian_singkat','uraian_esai')
		 ORDER BY s.nama ASC, jw.updated_at ASC`, jadwalID)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca daftar koreksi.")
	}
	defer rows.Close()
	type item struct {
		JawabanID     string   `json:"jawaban_id"`
		AttemptID     string   `json:"attempt_id"`
		SoalID        string   `json:"soal_id"`
		Jawaban       any      `json:"jawaban"`
		Ragu          bool     `json:"ragu"`
		Poin          *float64 `json:"poin"`
		Dikoreksi     bool     `json:"dikoreksi"`
		UpdatedAt     any      `json:"updated_at"`
		SiswaID       string   `json:"siswa_id"`
		AttemptStatus string   `json:"attempt_status"`
		NamaSiswa     string   `json:"nama_siswa"`
		NISN          string   `json:"nisn"`
		Tipe          string   `json:"tipe"`
		Pertanyaan    string   `json:"pertanyaan"`
		Bobot         int      `json:"bobot"`
	}
	items := []item{}
	for rows.Next() {
		var it item
		var jwb []byte
		if err := rows.Scan(&it.JawabanID, &it.AttemptID, &it.SoalID, &jwb, &it.Ragu,
			&it.Poin, &it.Dikoreksi, &it.UpdatedAt, &it.SiswaID, &it.AttemptStatus,
			&it.NamaSiswa, &it.NISN, &it.Tipe, &it.Pertanyaan, &it.Bobot); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca daftar koreksi.")
		}
		var jwbAny any
		if len(jwb) > 0 {
			_ = json.Unmarshal(jwb, &jwbAny)
		}
		it.Jawaban = jwbAny
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca daftar koreksi.")
	}
	return writeOK(c, map[string]any{"jadwal_id": jadwalID, "items": items})
}

func koreksiJawaban(c *echo.Context) error {
	uid, _, _, ok := requireRole(c, "admin", "guru")
	if !ok {
		return nil
	}
	var body struct {
		Skor float64 `json:"skor"`
	}
	if err := c.Bind(&body); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Body tidak valid.")
	}
	if body.Skor < 0 || body.Skor > 100 {
		return writeError(c, http.StatusBadRequest, "validation_error", "skor harus 0–100.")
	}
	jawabanID := c.Param("jawabanId")
	pool := c.Get("pool").(*pgxpool.Pool)
	ctx := c.Request().Context()
	var attemptID, soalID string
	var bobot float64
	err := pool.QueryRow(ctx,
		`SELECT jw.attempt_id, jw.soal_id, so.bobot
		 FROM jawaban jw JOIN soal so ON so.id = jw.soal_id
		 WHERE jw.id = $1`, jawabanID).Scan(&attemptID, &soalID, &bobot)
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Jawaban tidak ditemukan.")
	}
	// poin = skor/100 × bobot.
	poin := body.Skor / 100 * bobot
	tx, err := pool.Begin(ctx)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memulai koreksi.")
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx,
		`UPDATE jawaban SET poin = $2, dikoreksi = TRUE, updated_at = now() WHERE id = $1`,
		jawabanID, poin); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menyimpan koreksi.")
	}
	var attemptStatus string
	var jadwalID string
	_ = tx.QueryRow(ctx, `SELECT status, jadwal_id FROM attempt WHERE id = $1`, attemptID).
		Scan(&attemptStatus, &jadwalID)
	var nilaiBaru *float64
	menunggu := true
	if attemptStatus == "selesai" {
		// Hitung ulang nilai akhir attempt.
		if err := tx.Commit(ctx); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menyimpan koreksi.")
		}
		nilai, m, err := hitungNilai(ctx, pool, attemptID)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghitung ulang nilai.")
		}
		menunggu = m
		_, _ = pool.Exec(ctx, `UPDATE attempt SET nilai_akhir = $2 WHERE id = $1`, attemptID, nilai)
		nilaiBaru = &nilai
	} else {
		if err := tx.Commit(ctx); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menyimpan koreksi.")
		}
		// Attempt masih berjalan: nilai akhir dihitung saat submit.
		_, menunggu, _ = hitungNilai(ctx, pool, attemptID)
	}
	auditCatat(ctx, pool, uid, "koreksi.update", "jawaban", jawabanID,
		map[string]any{"attempt_id": attemptID, "soal_id": soalID, "skor": body.Skor, "poin": poin}, c.RealIP())
	a, _ := getAttempt(ctx, pool, attemptID)
	if a.ID != "" && nilaiBaru != nil {
		syncNilaiAkhir(ctx, pool, a, *nilaiBaru)
	}
	broadcastAttempt(ctx, pool, a)
	resp := map[string]any{
		"jawaban_id": jawabanID, "attempt_id": attemptID, "poin": poin,
		"menunggu_koreksi": menunggu,
	}
	if nilaiBaru != nil {
		resp["nilai_akhir_baru"] = *nilaiBaru
		resp["final"] = !menunggu
	}
	return writeOK(c, resp)
}

// ---------------------------------------------------------------------------
// Register
// ---------------------------------------------------------------------------

// Register mendaftarkan seluruh endpoint Tim 4.
// pool disimpan di context tiap request (c.Set("pool", ...)) agar handler
// tidak bergantung pada variabel global; q/cfg diterima sesuai kontrak.
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q
	_ = cfg

	// Middleware pool hanya untuk grup route Tim 4 (tidak menyentuh
	// route tim lain yang terdaftar di Echo yang sama).
	g := e.Group("", func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Set("pool", pool)
			return next(c)
		}
	})

	// Paket.
	g.GET("/api/v1/paket", listPaket)
	g.POST("/api/v1/paket", createPaket)
	g.GET("/api/v1/paket/:id", getPaket)
	g.PUT("/api/v1/paket/:id", updatePaket)
	g.DELETE("/api/v1/paket/:id", deletePaket)
	g.GET("/api/v1/paket/:id/soal", listPaketSoal)
	g.POST("/api/v1/paket/:id/soal", addPaketSoal)
	g.DELETE("/api/v1/paket/:id/soal", deletePaketSoal)
	g.POST("/api/v1/paket/:id/generate-acak", generateAcak)

	// Jadwal + token + pause/resume + unblock.
	g.GET("/api/v1/jadwal", listJadwal)
	g.POST("/api/v1/jadwal", createJadwal)
	g.GET("/api/v1/jadwal/:id", getJadwalDetail)
	g.PUT("/api/v1/jadwal/:id", updateJadwal)
	g.DELETE("/api/v1/jadwal/:id", deleteJadwal)
	g.POST("/api/v1/jadwal/:id/pause", pauseJadwal)
	g.POST("/api/v1/jadwal/:id/resume", resumeJadwal)
	g.POST("/api/v1/jadwal/:id/token", generateJadwalToken)
	g.GET("/api/v1/jadwal/:id/token-status", tokenStatus)
	g.GET("/api/v1/jadwal/:id/usulan-unblock", listUsulanUnblock)
	g.POST("/api/v1/jadwal/:id/usulan-unblock", createUsulanUnblock)
	g.POST("/api/v1/jadwal/:id/unblock", unblockAttempt)

	// Ujian siswa.
	g.GET("/api/v1/ujian/aktif", ujianAktif)
	g.POST("/api/v1/ujian/mulai", mulaiUjian)
	g.GET("/api/v1/ujian/attempt/:id", getAttemptDetail)
	g.POST("/api/v1/ujian/attempt/:id/jawaban", simpanJawaban)
	g.POST("/api/v1/ujian/attempt/:id/ragu", tandaiRagu)
	g.POST("/api/v1/ujian/attempt/:id/pelanggaran", catatPelanggaran)
	g.POST("/api/v1/ujian/attempt/:id/submit", submitUjian)
	g.GET("/api/v1/ujian/attempt/:id/status", statusAttempt)
	g.GET("/api/v1/ujian/attempt/:id/hasil", hasilAttempt)

	// Koreksi.
	g.GET("/api/v1/koreksi/:jadwalId", listKoreksi)
	g.POST("/api/v1/koreksi/:jawabanId", koreksiJawaban)
}
