// Tes end-to-end Tim 4 dengan PostgreSQL nyata.
//
// Dijalankan hanya bila TEST_DATABASE_URL diset, mis.:
//
//	TEST_DATABASE_URL=postgres://tim4test:tim4test@127.0.0.1:5432/cbt_v2_tim4_test
//	go test ./internal/exam/ -run TestE2E -count=1 -v
//
// Tanpa env itu, tes di-skip (aman untuk `go test ./...` di CI).
// Skenario: paket -> jadwal -> token -> mulai -> jawaban -> pelanggaran
// (peringatan lalu auto-block) -> unblock -> submit -> koreksi ->
// hasil tersembunyi -> pause/resume -> token kedua hanguskan pertama
// -> monitor ticket + snapshot.
package exam_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"cbt-smk-v2/backend/internal/exam"
	"cbt-smk-v2/backend/internal/monitor"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

const (
	adminID    = "11111111-1111-1111-1111-111111111111"
	proktorID  = "22222222-2222-2222-2222-222222222222"
	guruID     = "33333333-3333-3333-3333-333333333333"
	pengawasID = "44444444-4444-4444-4444-444444444444"
	siswaID    = "55555555-5555-5555-5555-555555555555"
	mapelID    = "66666666-6666-6666-6666-666666666666"
	kelasID    = "77777777-7777-7777-7777-777777777777"
	bankID     = "88888888-8888-8888-8888-888888888888"
	soalPG1    = "99999999-9999-9999-9999-999999999991"
	soalPG2    = "99999999-9999-9999-9999-999999999992"
	soalUraian = "99999999-9999-9999-9999-999999999993"
)

func setupDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL tidak diset — skip tes E2E.")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("koneksi DB: %v", err)
	}
	// Bersihkan skema + buat tabel dependensi minimal (Tim 1/2/3).
	ddl := []string{
		`DROP TABLE IF EXISTS usulan_unblock, pelanggaran, jawaban, attempt, jadwal_ujian, paket_soal, paket_ujian CASCADE`,
		`DROP TABLE IF EXISTS soal, bank_soal, siswa, kelas, mapel, users, settings, audit_log CASCADE`,
		`CREATE TABLE users (id UUID PRIMARY KEY, username TEXT)`,
		`CREATE TABLE mapel (id UUID PRIMARY KEY, nama TEXT)`,
		`CREATE TABLE kelas (id UUID PRIMARY KEY, nama TEXT)`,
		`CREATE TABLE bank_soal (id UUID PRIMARY KEY, mapel_id UUID REFERENCES mapel(id), nama TEXT)`,
		`CREATE TABLE soal (id UUID PRIMARY KEY, bank_id UUID REFERENCES bank_soal(id), tipe TEXT, stimulus TEXT, pertanyaan TEXT, media_url TEXT, jawaban JSONB, kunci JSONB, bobot INT, level TEXT, status TEXT)`,
		`CREATE TABLE siswa (id UUID PRIMARY KEY, nisn TEXT, nama TEXT, kelas_id UUID REFERENCES kelas(id), user_id UUID)`,
		`CREATE TABLE settings (key TEXT PRIMARY KEY, value JSONB)`,
		`CREATE TABLE audit_log (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), actor_id UUID, aksi TEXT, entitas TEXT, entitas_id TEXT, detail JSONB, ip TEXT, created_at TIMESTAMPTZ DEFAULT now())`,
	}
	for _, q := range ddl {
		if _, err := pool.Exec(ctx, q); err != nil {
			t.Fatalf("DDL: %v", err)
		}
	}
	migrasi, err := os.ReadFile("../../migrations/0004_exam.up.sql")
	if err != nil {
		t.Fatalf("baca migrasi 0004: %v", err)
	}
	if _, err := pool.Exec(ctx, string(migrasi)); err != nil {
		t.Fatalf("migrasi 0004: %v", err)
	}
	seed := []string{
		fmt.Sprintf(`INSERT INTO users VALUES ('%s','admin'),('%s','proktor'),('%s','guru'),('%s','pengawas'),('%s','siswa')`,
			adminID, proktorID, guruID, pengawasID, siswaID),
		fmt.Sprintf(`INSERT INTO mapel VALUES ('%s','Informatika')`, mapelID),
		fmt.Sprintf(`INSERT INTO kelas VALUES ('%s','XI PPLG 1')`, kelasID),
		fmt.Sprintf(`INSERT INTO bank_soal VALUES ('%s','%s','Bank Informatika')`, bankID, mapelID),
		fmt.Sprintf(`INSERT INTO siswa VALUES ('%s','12345','Budi','%s','%s')`, siswaID, kelasID, siswaID),
		`INSERT INTO settings VALUES ('tampil_hasil_ke_siswa','false'),('token_expiry_detik','900'),('toleransi_telat_menit','15'),('max_pelanggaran','2')`,
		fmt.Sprintf(`INSERT INTO soal VALUES
			('%s','%s','pg','','Ibukota Indonesia?','', '{"opsi":[{"label":"A","teks":"Bandung"},{"label":"B","teks":"Jakarta"},{"label":"C","teks":"Surabaya"},{"label":"D","teks":"Medan"}]}', '{"opsi":"B"}', 10, 'sedang', 'aktif'),
			('%s','%s','pg','','2 + 2 = ?','', '{"opsi":[{"label":"A","teks":"3"},{"label":"B","teks":"5"},{"label":"C","teks":"4"},{"label":"D","teks":"6"}]}', '{"opsi":"C"}', 10, 'mudah', 'aktif'),
			('%s','%s','uraian','','Jelaskan pengertian algoritma.','', '{}', '{"rubrik":"..."}', 10, 'sedang', 'aktif')`,
			soalPG1, bankID, soalPG2, bankID, soalUraian, bankID),
	}
	for _, q := range seed {
		if _, err := pool.Exec(ctx, q); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	return pool
}

type client struct {
	e    *echo.Echo
	pool *pgxpool.Pool
	t    *testing.T
}

func (c *client) req(method, path, userID, role, body string) (int, map[string]any) {
	c.t.Helper()
	var rdr *strings.Reader
	if body == "" {
		rdr = strings.NewReader("")
	} else {
		rdr = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, rdr)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User", userID)
	req.Header.Set("X-Test-Role", role)
	req.Header.Set("X-Test-Siswa", userID)
	rec := httptest.NewRecorder()
	c.e.ServeHTTP(rec, req)
	var parsed map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &parsed)
	if rec.Code >= 500 {
		c.t.Logf("%s %s -> %d %s", method, path, rec.Code, rec.Body.String())
	}
	return rec.Code, parsed
}

func data(t *testing.T, status int, resp map[string]any) map[string]any {
	t.Helper()
	if errObj, ada := resp["error"]; ada {
		t.Fatalf("respons error (status %d): %v", status, errObj)
	}
	d, _ := resp["data"].(map[string]any)
	return d
}

func TestE2E(t *testing.T) {
	pool := setupDB(t)
	defer pool.Close()
	ctx := context.Background()

	e := echo.New()
	// Middleware auth palsu untuk tes: isi context dari header.
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Set("auth_user_id", c.Request().Header.Get("X-Test-User"))
			c.Set("auth_role", c.Request().Header.Get("X-Test-Role"))
			c.Set("auth_siswa_id", c.Request().Header.Get("X-Test-Siswa"))
			return next(c)
		}
	})
	exam.Register(e, pool, nil, nil)
	monitor.Register(e, pool, nil, nil)
	cl := &client{e: e, pool: pool, t: t}

	// 1. Paket (admin) + 3 soal.
	status, resp := cl.req("POST", "/api/v1/paket", adminID, "admin",
		fmt.Sprintf(`{"kode":"STS-1","nama":"STS Informatika","mapel_id":"%s","jenis":"STS","durasi_menit":60,"jumlah_soal":0,"acak_soal":false,"acak_jawaban":false,"tampil_hasil":false}`, mapelID))
	if status != 201 {
		t.Fatalf("buat paket: %d %v", status, resp)
	}
	paketID := data(t, status, resp)["id"].(string)
	for _, soalID := range []string{soalPG1, soalPG2, soalUraian} {
		status, _ = cl.req("POST", "/api/v1/paket/"+paketID+"/soal", adminID, "admin",
			fmt.Sprintf(`{"soal_id":"%s"}`, soalID))
		if status != 201 {
			t.Fatalf("tambah soal %s: %d", soalID, status)
		}
	}

	// 2. Jadwal (proktor): mulai 5 menit lalu, selesai 55 menit lagi.
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	tanggal := now.Format("2006-01-02")
	mulai := now.Add(-5 * time.Minute).Format("15:04:05")
	selesai := now.Add(55 * time.Minute).Format("15:04:05")
	status, resp = cl.req("POST", "/api/v1/jadwal", proktorID, "proktor",
		fmt.Sprintf(`{"paket_id":"%s","kelas_id":"%s","tanggal":"%s","jam_mulai":"%s","jam_selesai":"%s","status":"aktif"}`,
			paketID, kelasID, tanggal, mulai, selesai))
	if status != 201 {
		t.Fatalf("buat jadwal: %d %v", status, resp)
	}
	jadwalID := data(t, status, resp)["id"].(string)

	// 3. Token (proktor) — tangkap plaintext sekali.
	status, resp = cl.req("POST", "/api/v1/jadwal/"+jadwalID+"/token", proktorID, "proktor", `{}`)
	if status != 200 {
		t.Fatalf("generate token: %d %v", status, resp)
	}
	token := data(t, status, resp)["token"].(string)
	if len(token) != 6 {
		t.Fatalf("token harus 6 karakter, dapat %q", token)
	}

	// 4. Siswa mulai dengan token.
	status, resp = cl.req("POST", "/api/v1/ujian/mulai", siswaID, "siswa",
		fmt.Sprintf(`{"jadwal_id":"%s","token":"%s"}`, jadwalID, token))
	if status != 201 {
		t.Fatalf("mulai ujian: %d %v", status, resp)
	}
	attemptID := data(t, status, resp)["attempt_id"].(string)

	// 5. GET attempt: 3 soal, TANPA kunci.
	status, resp = cl.req("GET", "/api/v1/ujian/attempt/"+attemptID, siswaID, "siswa", "")
	if status != 200 {
		t.Fatalf("get attempt: %d %v", status, resp)
	}
	// Kunci jawaban tidak pernah dikirim ke browser: respons mentah
	// tidak boleh mengandung kunci "kunci" sama sekali.
	rawAttempt, _ := json.Marshal(resp)
	if strings.Contains(string(rawAttempt), "kunci") {
		t.Fatalf("respons attempt membocorkan kunci jawaban: %s", rawAttempt)
	}
	soalList, _ := data(t, status, resp)["soal"].([]any)
	if len(soalList) != 3 {
		t.Fatalf("jumlah soal: %d, mau 3", len(soalList))
	}

	// 6. Jawaban: PG1 benar (B), PG2 salah (A), uraian teks.
	for _, p := range []struct{ soal, jwb string }{
		{soalPG1, `"B"`}, {soalPG2, `"A"`}, {soalUraian, `"Algoritma adalah langkah-langkah."`},
	} {
		status, _ = cl.req("POST", "/api/v1/ujian/attempt/"+attemptID+"/jawaban", siswaID, "siswa",
			fmt.Sprintf(`{"soal_id":"%s","jawaban":%s}`, p.soal, p.jwb))
		if status != 200 {
			t.Fatalf("simpan jawaban %s: %d", p.soal, status)
		}
	}

	// 7. Pelanggaran 1x = peringatan.
	status, resp = cl.req("POST", "/api/v1/ujian/attempt/"+attemptID+"/pelanggaran", siswaID, "siswa",
		`{"jenis":"pindah_tab"}`)
	d := data(t, status, resp)
	if d["peringatan"] != true {
		t.Fatalf("pelanggaran 1: mau peringatan, dapat %v", d)
	}
	// Mundurkan created_at agar dedupe 5 detik tidak mengabaikan pelanggaran ke-2.
	_, _ = pool.Exec(ctx, `UPDATE pelanggaran SET created_at = now() - interval '10 seconds' WHERE attempt_id = $1`, attemptID)
	// Pelanggaran 2x = auto-block per-attempt.
	status, resp = cl.req("POST", "/api/v1/ujian/attempt/"+attemptID+"/pelanggaran", siswaID, "siswa",
		`{"jenis":"pindah_tab"}`)
	d = data(t, status, resp)
	if d["diblokir"] != true {
		t.Fatalf("pelanggaran 2: mau diblokir, dapat %v", d)
	}

	// 8. Unblock (proktor, wajib alasan) = lanjut sisa waktu.
	status, resp = cl.req("POST", "/api/v1/jadwal/"+jadwalID+"/unblock", proktorID, "proktor",
		fmt.Sprintf(`{"attempt_id":"%s","alasan":"Siswa mengaku tidak sengaja, diverifikasi pengawas."}`, attemptID))
	if status != 200 {
		t.Fatalf("unblock: %d %v", status, resp)
	}
	if data(t, status, resp)["status"] != "berjalan" {
		t.Fatalf("unblock: status bukan berjalan")
	}

	// 9. Submit: nilai disembunyikan (tampil_hasil=false), menunggu koreksi.
	status, resp = cl.req("POST", "/api/v1/ujian/attempt/"+attemptID+"/submit", siswaID, "siswa", `{}`)
	d = data(t, status, resp)
	if d["nilai_akhir"] != nil {
		t.Fatalf("submit: nilai harus disembunyikan (null), dapat %v", d["nilai_akhir"])
	}
	if d["menunggu_koreksi"] != true {
		t.Fatalf("submit: menunggu_koreksi harus true (uraian belum dikoreksi)")
	}

	// 10. Koreksi (guru): uraian skor 100 -> poin 10 -> nilai (10+0+10)/30*100 = 66.67.
	status, resp = cl.req("GET", "/api/v1/koreksi/"+jadwalID, guruID, "guru", "")
	items, _ := data(t, status, resp)["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("daftar koreksi: %d item, mau 1 (uraian)", len(items))
	}
	jawabanID := items[0].(map[string]any)["jawaban_id"].(string)
	status, resp = cl.req("POST", "/api/v1/koreksi/"+jawabanID, guruID, "guru", `{"skor":100}`)
	d = data(t, status, resp)
	if d["poin"] != 10.0 {
		t.Fatalf("koreksi: poin %v, mau 10", d["poin"])
	}
	nilaiBaru, _ := d["nilai_akhir_baru"].(float64)
	if nilaiBaru < 66.66 || nilaiBaru > 66.68 {
		t.Fatalf("koreksi: nilai_akhir_baru %v, mau ~66.67", nilaiBaru)
	}
	if d["final"] != true {
		t.Fatalf("koreksi: final harus true setelah semua terkoreksi")
	}

	// 11. Hasil siswa: nilai tetap disembunyikan (paket tampil_hasil=false).
	status, resp = cl.req("GET", "/api/v1/ujian/attempt/"+attemptID+"/hasil", siswaID, "siswa", "")
	d = data(t, status, resp)
	if d["nilai_akhir"] != nil {
		t.Fatalf("hasil: nilai harus disembunyikan, dapat %v", d["nilai_akhir"])
	}
	if len(d["items"].([]any)) != 3 {
		t.Fatalf("hasil: %d item, mau 3", len(d["items"].([]any)))
	}

	// 12. Token kedua menghanguskan pertama: hash berubah, token-status tanpa token.
	var hashLama string
	_ = pool.QueryRow(ctx, `SELECT token_hash FROM jadwal_ujian WHERE id = $1`, jadwalID).Scan(&hashLama)
	status, resp = cl.req("POST", "/api/v1/jadwal/"+jadwalID+"/token", proktorID, "proktor", `{}`)
	token2 := data(t, status, resp)["token"].(string)
	if token2 == token {
		t.Fatalf("token kedua harus berbeda dari pertama")
	}
	var hashBaru string
	_ = pool.QueryRow(ctx, `SELECT token_hash FROM jadwal_ujian WHERE id = $1`, jadwalID).Scan(&hashBaru)
	if hashLama == hashBaru {
		t.Fatalf("token_hash harus berubah setelah generate kedua")
	}
	status, resp = cl.req("GET", "/api/v1/jadwal/"+jadwalID+"/token-status", proktorID, "proktor", "")
	d = data(t, status, resp)
	if d["ada_token"] != true {
		t.Fatalf("token-status: ada_token harus true")
	}
	// Token-status TIDAK menampilkan token.
	raw, _ := json.Marshal(resp)
	if strings.Contains(strings.ToLower(string(raw)), token2) {
		t.Fatalf("token-status membocorkan nilai token!")
	}

	// 13. Pause membekukan timer + resume memulihkan deadline.
	// Buat jadwal kedua + attempt kedua untuk uji pause.
	status, resp = cl.req("POST", "/api/v1/jadwal", proktorID, "proktor",
		fmt.Sprintf(`{"paket_id":"%s","kelas_id":"%s","tanggal":"%s","jam_mulai":"%s","jam_selesai":"%s","status":"aktif"}`,
			paketID, kelasID, tanggal, mulai, selesai))
	jadwal2 := data(t, status, resp)["id"].(string)
	status, resp = cl.req("POST", "/api/v1/jadwal/"+jadwal2+"/token", proktorID, "proktor", `{}`)
	tokenJ2 := data(t, status, resp)["token"].(string)
	status, resp = cl.req("POST", "/api/v1/ujian/mulai", siswaID, "siswa",
		fmt.Sprintf(`{"jadwal_id":"%s","token":"%s"}`, jadwal2, tokenJ2))
	attempt2 := data(t, status, resp)["attempt_id"].(string)
	status, resp = cl.req("POST", "/api/v1/jadwal/"+jadwal2+"/pause", proktorID, "proktor", `{}`)
	if status != 200 {
		t.Fatalf("pause: %d %v", status, resp)
	}
	var sisaBeku *int
	_ = pool.QueryRow(ctx, `SELECT sisa_waktu_beku FROM attempt WHERE id = $1`, attempt2).Scan(&sisaBeku)
	if sisaBeku == nil || *sisaBeku <= 0 {
		t.Fatalf("pause: sisa_waktu_beku harus > 0, dapat %v", sisaBeku)
	}
	status, resp = cl.req("POST", "/api/v1/jadwal/"+jadwal2+"/resume", proktorID, "proktor", `{}`)
	if status != 200 {
		t.Fatalf("resume: %d %v", status, resp)
	}
	var deadline time.Time
	_ = pool.QueryRow(ctx, `SELECT deadline, sisa_waktu_beku FROM attempt WHERE id = $1`, attempt2).Scan(&deadline, &sisaBeku)
	if sisaBeku != nil {
		t.Fatalf("resume: sisa_waktu_beku harus NULL setelah resume")
	}
	if time.Until(deadline) <= 0 {
		t.Fatalf("resume: deadline baru harus di masa depan")
	}

	// 13b. Regresi: deadline lewat SAAT pause tidak boleh auto-submit
	// (timer beku). Buat jadwal ketiga, pause, mundurkan deadline,
	// lalu GET status harus tetap berjalan.
	status, resp = cl.req("POST", "/api/v1/jadwal", proktorID, "proktor",
		fmt.Sprintf(`{"paket_id":"%s","kelas_id":"%s","tanggal":"%s","jam_mulai":"%s","jam_selesai":"%s","status":"aktif"}`,
			paketID, kelasID, tanggal, mulai, selesai))
	jadwal3 := data(t, status, resp)["id"].(string)
	status, resp = cl.req("POST", "/api/v1/jadwal/"+jadwal3+"/token", proktorID, "proktor", `{}`)
	tokenJ3 := data(t, status, resp)["token"].(string)
	status, resp = cl.req("POST", "/api/v1/ujian/mulai", siswaID, "siswa",
		fmt.Sprintf(`{"jadwal_id":"%s","token":"%s"}`, jadwal3, tokenJ3))
	attempt3 := data(t, status, resp)["attempt_id"].(string)
	status, _ = cl.req("POST", "/api/v1/jadwal/"+jadwal3+"/pause", proktorID, "proktor", `{}`)
	if status != 200 {
		t.Fatalf("pause jadwal3: %d", status)
	}
	_, _ = pool.Exec(ctx, `UPDATE attempt SET deadline = now() - interval '1 minute' WHERE id = $1`, attempt3)
	status, resp = cl.req("GET", "/api/v1/ujian/attempt/"+attempt3+"/status", siswaID, "siswa", "")
	if data(t, status, resp)["status"] != "berjalan" {
		t.Fatalf("status saat pause + deadline lewat harus tetap berjalan, dapat %v", data(t, status, resp)["status"])
	}
	status, _ = cl.req("POST", "/api/v1/jadwal/"+jadwal3+"/resume", proktorID, "proktor", `{}`)
	if status != 200 {
		t.Fatalf("resume jadwal3: %d", status)
	}

	// 14. Monitor: ticket + snapshot REST.
	status, resp = cl.req("POST", "/api/v1/monitor/ticket", proktorID, "proktor",
		fmt.Sprintf(`{"jadwal_id":"%s"}`, jadwalID))
	if status != 200 || data(t, status, resp)["ticket"] == "" {
		t.Fatalf("monitor ticket: %d %v", status, resp)
	}
	status, resp = cl.req("GET", "/api/v1/monitor/"+jadwalID, proktorID, "proktor", "")
	peserta, _ := data(t, status, resp)["peserta"].([]any)
	if len(peserta) != 1 {
		t.Fatalf("monitor snapshot: %d peserta, mau 1", len(peserta))
	}
}
