// Program tes end-to-end CBT SMK v2 (Tim QA + Uji Beban).
//
// Alur: jurusan -> kelas -> mapel -> guru -> siswa -> bank soal
// (PG + uraian) -> paket -> jadwal -> token -> siswa ujian
// (jawab + submit) -> koreksi uraian -> nilai -> rekap ->
// finalisasi -> berita acara -> export CSV.
//
// Setiap langkah meng-assert status HTTP + isi respons. Bila satu
// langkah gagal, program BERHENTI dan mencetak langkah yang gagal
// beserta respons mentahnya (fail-fast).
//
// Prasyarat (server + DB scratch milik penguji, BUKAN server tim lain):
//
//	DATABASE_URL=postgres://.../cbt_v2_beban  PORT=29988 go run ./cmd/server serve
//	psql $DATABASE_URL -f db/seed/seed.sql   # seed role + 59 permission + settings
//	go run ./cmd/server create-admin --email <ADMIN_LOGIN> --password <ADMIN_INITIAL_PASSWORD> --nama "Admin QA"
//
// Pemakaian:
//
//	BASE_URL=http://127.0.0.1:29988 ADMIN_LOGIN=... ADMIN_INITIAL_PASSWORD=... \
//	  ADMIN_PASSWORD=... go run ./test/e2e [-prefix QA...] [-cleanup=false]
//
// Prefix unik per run (default QA+timestamp) agar rerun tidak bentrok
// pada constraint UNIQUE (kode, NISN, username). Cleanup best-effort
// via API; karena guard 409 melarang hapus jadwal/siswa yang sudah
// punya attempt, cara bersih tuntas adalah DROP DATABASE scratch
// (program mencetak perintahnya di akhir).
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"time"
)

var (
	base       = envOr("BASE_URL", "http://127.0.0.1:29988")
	adminLogin = os.Getenv("ADMIN_LOGIN")
	adminInit  = os.Getenv("ADMIN_INITIAL_PASSWORD")
	adminPass  = os.Getenv("ADMIN_PASSWORD")
	prefix     string
	doCleanup  bool
	failed     bool
	steps      []stepResult
	// ID-ID yang dibuat selama run (untuk cleanup best-effort).
	ids = map[string]string{}
)

type stepResult struct {
	No     int
	Name   string
	Status string // PASS / FAIL
	Detail string
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	flag.StringVar(&prefix, "prefix", "", "Prefix unik data tes (default: QA+YYYYMMDDHHMMSS)")
	flag.BoolVar(&doCleanup, "cleanup", true, "Coba hapus data tes via API di akhir (best-effort)")
	flag.Parse()
	if prefix == "" {
		prefix = "QA" + time.Now().Format("20060102150405")
	}
	if adminLogin == "" || adminInit == "" || adminPass == "" {
		fmt.Fprintln(os.Stderr, "E2E: env ADMIN_LOGIN, ADMIN_INITIAL_PASSWORD, ADMIN_PASSWORD wajib diisi.")
		os.Exit(2)
	}

	admin := newClient()
	run("Admin login (ganti password pertama bila perlu)", func() error {
		return admin.ensureLogin(adminLogin, adminInit, adminPass)
	})
	run("Admin: GET /auth/me mengembalikan role admin", func() error {
		st, body := admin.do("GET", "/api/v1/auth/me", nil, false)
		if st != 200 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		if getStr(body, "data", "user", "role") != "admin" {
			return fmt.Errorf("role bukan admin: %s", body)
		}
		return nil
	})

	var jurusanID, kelasID, mapelID, guruID, siswaID, bankID, soalPG, soalUraian, paketID, jadwalID, token, attemptID, jawabanID string

	run("Buat jurusan", func() error {
		st, body := admin.do("POST", "/api/v1/master/jurusan", map[string]any{
			"kode": prefix + "-J", "nama": "Jurusan QA " + prefix,
		}, true)
		if st != 201 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		jurusanID = getStr(body, "data", "id")
		if jurusanID == "" {
			return fmt.Errorf("id jurusan kosong: %s", body)
		}
		ids["jurusan"] = jurusanID
		return nil
	})
	run("Buat kelas", func() error {
		st, body := admin.do("POST", "/api/v1/master/kelas", map[string]any{
			"jurusan_id": jurusanID, "tingkat": 11, "nama": "XI QA " + prefix, "tahun_ajaran": "2026/2027",
		}, true)
		if st != 201 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		kelasID = getStr(body, "data", "id")
		ids["kelas"] = kelasID
		return nil
	})
	run("Buat mapel (produktif, terikat jurusan)", func() error {
		st, body := admin.do("POST", "/api/v1/master/mapel", map[string]any{
			"kode": prefix + "-M", "nama": "Mapel QA " + prefix, "jenis": "produktif", "jurusan_id": jurusanID,
		}, true)
		if st != 201 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		mapelID = getStr(body, "data", "id")
		ids["mapel"] = mapelID
		return nil
	})
	run("Buat guru (mapel_ids berisi mapel QA)", func() error {
		st, body := admin.do("POST", "/api/v1/master/guru", map[string]any{
			"nama": "Guru QA " + prefix, "nip": "99" + time.Now().Format("060102150405") + "02", "mapel_ids": []string{mapelID},
		}, true)
		if st != 201 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		guruID = getStr(body, "data", "id")
		ids["guru"] = guruID
		return nil
	})
	nisn := "99" + time.Now().Format("060102150405") + "01" // 16 digit, unik per run
	run("Buat siswa (akun login dibuat otomatis: username=NISN)", func() error {
		st, body := admin.do("POST", "/api/v1/master/siswa", map[string]any{
			"nisn": nisn, "nama": "Siswa QA " + prefix, "kelas_id": kelasID, "jurusan_id": jurusanID,
		}, true)
		if st != 201 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		siswaID = getStr(body, "data", "id")
		ids["siswa"] = siswaID
		return nil
	})
	run("Buat bank soal", func() error {
		st, body := admin.do("POST", "/api/v1/bank-soal", map[string]any{
			"kode": prefix + "-B", "nama": "Bank QA " + prefix, "mapel_id": mapelID, "kelas_tingkat": 11,
		}, true)
		if st != 201 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		bankID = getStr(body, "data", "id")
		ids["bank"] = bankID
		return nil
	})
	run("Buat soal PG (kunci B, bobot 10)", func() error {
		st, body := admin.do("POST", "/api/v1/bank-soal/"+bankID+"/soal", map[string]any{
			"tipe": "pg", "pertanyaan": "Ibukota Indonesia?", "bobot": 10, "level": "sedang", "status": "aktif",
			"jawaban": map[string]any{"opsi": []any{
				map[string]any{"label": "A", "teks": "Bandung"},
				map[string]any{"label": "B", "teks": "Jakarta"},
				map[string]any{"label": "C", "teks": "Surabaya"},
				map[string]any{"label": "D", "teks": "Medan"},
			}},
			"kunci": map[string]any{"opsi": "B"},
		}, true)
		if st != 201 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		soalPG = getStr(body, "data", "id")
		return nil
	})
	run("Buat soal uraian (bobot 10)", func() error {
		st, body := admin.do("POST", "/api/v1/bank-soal/"+bankID+"/soal", map[string]any{
			"tipe": "uraian", "pertanyaan": "Jelaskan pengertian algoritma.", "bobot": 10,
			"level": "sedang", "status": "aktif",
			"jawaban": map[string]any{}, "kunci": map[string]any{"rubrik": "Pengertian + contoh"},
		}, true)
		if st != 201 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		soalUraian = getStr(body, "data", "id")
		return nil
	})
	run("Buat paket (STS, tampil_hasil=true agar nilai teruji)", func() error {
		st, body := admin.do("POST", "/api/v1/paket", map[string]any{
			"kode": prefix + "-P", "nama": "Paket QA " + prefix, "mapel_id": mapelID,
			"jenis": "STS", "durasi_menit": 60, "jumlah_soal": 2,
			"acak_soal": false, "acak_jawaban": false, "tampil_hasil": true,
		}, true)
		if st != 201 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		paketID = getStr(body, "data", "id")
		ids["paket"] = paketID
		return nil
	})
	for _, soalID := range []string{soalPG, soalUraian} {
		sid := soalID
		run("Tambah soal ke paket: "+sid[:8], func() error {
			st, body := admin.do("POST", "/api/v1/paket/"+paketID+"/soal", map[string]any{"soal_id": sid}, true)
			if st != 201 {
				return fmt.Errorf("status %d, body %v", st, body)
			}
			return nil
		})
	}
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	run("Buat jadwal aktif (mulai 5 mnt lalu, selesai 2 jam lagi)", func() error {
		st, body := admin.do("POST", "/api/v1/jadwal", map[string]any{
			"paket_id": paketID, "kelas_id": kelasID,
			"tanggal":   now.Format("2006-01-02"),
			"jam_mulai": now.Add(-5 * time.Minute).Format("15:04:05"),
			"jam_selesai": now.Add(2 * time.Hour).Format("15:04:05"),
			"status":      "aktif",
		}, true)
		if st != 201 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		jadwalID = getStr(body, "data", "id")
		ids["jadwal"] = jadwalID
		return nil
	})
	run("Generate token jadwal (6 karakter)", func() error {
		st, body := admin.do("POST", "/api/v1/jadwal/"+jadwalID+"/token", map[string]any{}, true)
		if st != 200 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		token = getStr(body, "data", "token")
		if len(token) != 6 {
			return fmt.Errorf("token harus 6 karakter, dapat %q (body %v)", token, body)
		}
		return nil
	})

	// --- Alur siswa ---
	siswa := newClient()
	siswaPass := "QaSiswa1!"
	run("Siswa: ganti password pertama (NISN -> password baru)", func() error {
		st, body := siswa.do("POST", "/api/v1/auth/ganti-password", map[string]any{
			"login": nisn, "password_lama": nisn, "password_baru": siswaPass,
		}, false)
		if st != 200 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		return nil
	})
	run("Siswa: login", func() error {
		return siswa.ensureLogin(nisn, siswaPass, siswaPass)
	})
	run("Siswa: mulai ujian dengan token", func() error {
		st, body := siswa.do("POST", "/api/v1/ujian/mulai", map[string]any{
			"jadwal_id": jadwalID, "token": token,
		}, true)
		if st != 201 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		attemptID = getStr(body, "data", "attempt_id")
		if attemptID == "" {
			return fmt.Errorf("attempt_id kosong: %s", body)
		}
		return nil
	})
	run("Siswa: GET attempt — 2 soal, TANPA kunci jawaban", func() error {
		st, raw := siswa.doRaw("GET", "/api/v1/ujian/attempt/"+attemptID, nil, false)
		if st != 200 {
			return fmt.Errorf("status %d, body %v", st, raw)
		}
		if strings.Contains(string(raw), `"kunci"`) {
			return fmt.Errorf("respons attempt membocorkan kunci jawaban")
		}
		var parsed map[string]any
		_ = json.Unmarshal(raw, &parsed)
		soal, _ := getPath(parsed, "data", "soal").([]any)
		if len(soal) != 2 {
			return fmt.Errorf("jumlah soal %d, mau 2: %s", len(soal), raw)
		}
		return nil
	})
	run("Siswa: jawab PG benar (B)", func() error {
		st, body := siswa.do("POST", "/api/v1/ujian/attempt/"+attemptID+"/jawaban", map[string]any{
			"soal_id": soalPG, "jawaban": "B",
		}, true)
		if st != 200 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		return nil
	})
	run("Siswa: jawab uraian (teks)", func() error {
		st, body := siswa.do("POST", "/api/v1/ujian/attempt/"+attemptID+"/jawaban", map[string]any{
			"soal_id": soalUraian, "jawaban": "Algoritma adalah langkah-langkah sistematis.",
		}, true)
		if st != 200 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		return nil
	})
	run("Siswa: tandai ragu-ragu pada soal uraian", func() error {
		st, body := siswa.do("POST", "/api/v1/ujian/attempt/"+attemptID+"/ragu", map[string]any{
			"soal_id": soalUraian, "ragu": true,
		}, true)
		if st != 200 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		return nil
	})
	run("Siswa: status attempt = berjalan, 2 jawaban", func() error {
		st, body := siswa.do("GET", "/api/v1/ujian/attempt/"+attemptID+"/status", nil, false)
		if st != 200 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		if getStr(body, "data", "status") != "berjalan" {
			return fmt.Errorf("status attempt bukan berjalan: %s", body)
		}
		return nil
	})
	run("Siswa: submit — menunggu koreksi (uraian belum dikoreksi)", func() error {
		st, body := siswa.do("POST", "/api/v1/ujian/attempt/"+attemptID+"/submit", map[string]any{}, true)
		if st != 200 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		if getPath(body, "data", "menunggu_koreksi") != true {
			return fmt.Errorf("menunggu_koreksi harus true: %s", body)
		}
		return nil
	})
	run("Admin: daftar koreksi berisi 1 uraian", func() error {
		st, body := admin.do("GET", "/api/v1/koreksi/"+jadwalID, nil, false)
		if st != 200 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		items, _ := getPath(body, "data", "items").([]any)
		if len(items) != 1 {
			return fmt.Errorf("item koreksi %d, mau 1: %s", len(items), body)
		}
		jawabanID, _ = items[0].(map[string]any)["jawaban_id"].(string)
		if jawabanID == "" {
			return fmt.Errorf("jawaban_id kosong: %s", body)
		}
		return nil
	})
	run("Admin: koreksi uraian skor 80 -> poin 8, nilai akhir 90, final", func() error {
		st, body := admin.do("POST", "/api/v1/koreksi/"+jawabanID, map[string]any{"skor": 80}, true)
		if st != 200 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		d, _ := getPath(body, "data").(map[string]any)
		if d == nil {
			return fmt.Errorf("data kosong: %v", body)
		}
		if f, _ := d["poin"].(float64); f != 8 {
			return fmt.Errorf("poin %v, mau 8: %s", d["poin"], body)
		}
		if f, _ := d["nilai_akhir_baru"].(float64); f < 89.99 || f > 90.01 {
			return fmt.Errorf("nilai_akhir_baru %v, mau 90: %s", d["nilai_akhir_baru"], body)
		}
		if d["final"] != true {
			return fmt.Errorf("final harus true: %s", body)
		}
		return nil
	})
	run("Siswa: hasil TERSEMBUNYI pada settings default (paket tampil AND settings=false)", func() error {
		st, body := siswa.do("GET", "/api/v1/ujian/attempt/"+attemptID+"/hasil", nil, false)
		if st != 200 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		if getPath(body, "data", "tampil_hasil") != false {
			return fmt.Errorf("tampil_hasil harus false pada settings default: %v", body)
		}
		if getPath(body, "data", "nilai_akhir") != nil {
			return fmt.Errorf("nilai harus disembunyikan pada settings default: %v", body)
		}
		return nil
	})
	run("Admin: aktifkan settings tampil_hasil_ke_siswa=true", func() error {
		st, body := admin.do("PUT", "/api/v1/settings", map[string]any{
			"tampil_hasil_ke_siswa": true,
		}, true)
		if st != 200 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		return nil
	})
	run("Siswa: hasil menampilkan nilai 90 (paket tampil_hasil=true)", func() error {
		st, body := siswa.do("GET", "/api/v1/ujian/attempt/"+attemptID+"/hasil", nil, false)
		if st != 200 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		if f, _ := getPath(body, "data", "nilai_akhir").(float64); f < 89.99 || f > 90.01 {
			return fmt.Errorf("nilai hasil %v, mau 90: %v", getPath(body, "data", "nilai_akhir"), body)
		}
		return nil
	})
	run("Admin: rekap berisi siswa dengan nilai 90", func() error {
		st, body := admin.do("GET", "/api/v1/rekap?jadwal_id="+jadwalID, nil, false)
		if st != 200 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		items, _ := getPath(body, "data", "items").([]any)
		if len(items) != 1 {
			return fmt.Errorf("item rekap %d, mau 1: %s", len(items), body)
		}
		it := items[0].(map[string]any)
		if f, _ := it["nilai"].(float64); f < 89.99 || f > 90.01 {
			return fmt.Errorf("nilai rekap %v, mau 90: %s", it["nilai"], body)
		}
		if it["nisn"] != nisn {
			return fmt.Errorf("nisn rekap %v, mau %s: %s", it["nisn"], nisn, body)
		}
		return nil
	})
	run("Admin: finalisasi rekap (1 difinalisasi)", func() error {
		st, body := admin.do("POST", "/api/v1/rekap/"+jadwalID+"/finalisasi", map[string]any{}, true)
		if st != 200 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		if f, _ := getPath(body, "data", "jumlah_difinalisasi").(float64); f != 1 {
			return fmt.Errorf("jumlah_difinalisasi %v, mau 1: %s", getPath(body, "data", "jumlah_difinalisasi"), body)
		}
		return nil
	})
	run("Admin: berita acara memuat mapel + 1 peserta", func() error {
		st, body := admin.do("GET", "/api/v1/rekap/"+jadwalID+"/berita-acara", nil, false)
		if st != 200 {
			return fmt.Errorf("status %d, body %v", st, body)
		}
		if getStr(body, "data", "mapel") == "" {
			return fmt.Errorf("mapel berita acara kosong: %s", body)
		}
		return nil
	})
	run("Admin: export CSV rekap (header + baris siswa)", func() error {
		st, raw := admin.doRaw("GET", "/api/v1/rekap/"+jadwalID+"/export.csv", nil, false)
		if st != 200 {
			return fmt.Errorf("status %d, body %v", st, raw)
		}
		if !strings.Contains(string(raw), "nisn") || !strings.Contains(string(raw), nisn) {
			return fmt.Errorf("CSV tidak memuat header nisn / baris siswa: %s", raw)
		}
		if !strings.Contains(string(raw), "90") {
			return fmt.Errorf("CSV tidak memuat nilai 90: %s", raw)
		}
		return nil
	})

	if doCleanup {
		cleanup(admin)
	}
	printSummary()
	fmt.Printf("\nBersihkan tuntas DB scratch: DROP DATABASE <nama-db-scratch>;  (data prefix %s)\n", prefix)
	if failed {
		os.Exit(1)
	}
	fmt.Println("E2E: SEMUA LANGKAH PASS")
}

// run mengeksekusi satu langkah; bila gagal, cetak dan keluar (fail-fast).
func run(name string, fn func() error) {
	no := len(steps) + 1
	if err := fn(); err != nil {
		steps = append(steps, stepResult{No: no, Name: name, Status: "FAIL", Detail: err.Error()})
		fmt.Printf("[%02d] FAIL %s\n     -> %v\n", no, name, err)
		printSummary()
		os.Exit(1)
	}
	steps = append(steps, stepResult{No: no, Name: name, Status: "PASS"})
	fmt.Printf("[%02d] PASS %s\n", no, name)
}

func printSummary() {
	fmt.Println("\n===== RINGKASAN E2E =====")
	pass := 0
	for _, s := range steps {
		if s.Status == "PASS" {
			pass++
		} else {
			failed = true
			fmt.Printf("[%02d] %s %s\n     %s\n", s.No, s.Status, s.Name, s.Detail)
		}
	}
	fmt.Printf("PASS %d / %d langkah (prefix %s)\n", pass, len(steps), prefix)
}

// cleanup best-effort via API (urutan terbalik). Guard 409 pada
// jadwal/siswa yang sudah punya attempt adalah perilaku yang benar —
// kegagalan di sini dicatat, bukan dianggap gagalnya tes.
func cleanup(admin *Client) {
	fmt.Println("\n--- Cleanup best-effort (kegagalan 409 = guard yang benar) ---")
	// Kembalikan settings yang diubah tes ini.
	st, body := admin.do("PUT", "/api/v1/settings", map[string]any{
		"tampil_hasil_ke_siswa": false,
	}, true)
	fmt.Printf("  PUT settings tampil_hasil_ke_siswa=false -> %d %v\n", st, body)
	del := func(label, path string) {
		st, body := admin.do("DELETE", path, nil, true)
		fmt.Printf("  DELETE %s -> %d %v\n", label, st, body)
	}
	if id := ids["jadwal"]; id != "" {
		del("jadwal", "/api/v1/jadwal/"+id)
	}
	if id := ids["paket"]; id != "" {
		del("paket", "/api/v1/paket/"+id)
	}
	if id := ids["bank"]; id != "" {
		del("bank-soal", "/api/v1/bank-soal/"+id)
	}
	if id := ids["siswa"]; id != "" {
		del("siswa", "/api/v1/master/siswa/"+id)
	}
	if id := ids["guru"]; id != "" {
		del("guru", "/api/v1/master/guru/"+id)
	}
	if id := ids["mapel"]; id != "" {
		del("mapel", "/api/v1/master/mapel/"+id)
	}
	if id := ids["kelas"]; id != "" {
		del("kelas", "/api/v1/master/kelas/"+id)
	}
	if id := ids["jurusan"]; id != "" {
		del("jurusan", "/api/v1/master/jurusan/"+id)
	}
}

// --- Klien HTTP minimal (cookie jar + CSRF) ---

type Client struct {
	http *http.Client
	csrf string
}

func newClient() *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{http: &http.Client{Jar: jar, Timeout: 60 * time.Second}}
}

// do mengirim request JSON; csrf=true menambahkan header X-CSRF-Token.
// Mengembalikan status + body ter-parse (map kosong bila bukan JSON).
func (c *Client) do(method, path string, body any, csrf bool) (int, map[string]any) {
	st, raw := c.doRaw(method, path, body, csrf)
	var parsed map[string]any
	_ = json.Unmarshal(raw, &parsed)
	if parsed == nil {
		parsed = map[string]any{"_raw": string(raw)}
	}
	return st, parsed
}

func (c *Client) doRaw(method, path string, body any, csrf bool) (int, []byte) {
	var rdr io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, base+path, rdr)
	if err != nil {
		return 0, []byte("build request: " + err.Error())
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if csrf && c.csrf != "" {
		req.Header.Set("X-CSRF-Token", c.csrf)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, []byte("http: " + err.Error())
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, raw
}

// ensureLogin: login dengan password target. Bila belum cocok, coba
// password awal: bila server menjawab PASSWORD_CHANGE_REQUIRED,
// ganti password (awal -> target) lalu login lagi.
// (Urutan cek server: verifikasi password dulu, baru flag
// must_change_password — jadi percobaan pertama yang gagal
// menghasilkan 401, bukan 403.)
func (c *Client) ensureLogin(login, initial, target string) error {
	st, body := c.do("POST", "/api/v1/auth/login", map[string]any{
		"login": login, "password": target,
	}, false)
	if st == 200 {
		return c.fetchCSRF()
	}
	st0, body0 := c.do("POST", "/api/v1/auth/login", map[string]any{
		"login": login, "password": initial,
	}, false)
	if st0 == 200 {
		return c.fetchCSRF()
	}
	if st0 == 403 && getStr(body0, "error", "code") == "PASSWORD_CHANGE_REQUIRED" {
		st2, body2 := c.do("POST", "/api/v1/auth/ganti-password", map[string]any{
			"login": login, "password_lama": initial, "password_baru": target,
		}, false)
		if st2 != 200 {
			return fmt.Errorf("ganti-password: status %d, body %v", st2, body2)
		}
		st3, body3 := c.do("POST", "/api/v1/auth/login", map[string]any{
			"login": login, "password": target,
		}, false)
		if st3 != 200 {
			return fmt.Errorf("login setelah ganti password: status %d, body %v", st3, body3)
		}
		return c.fetchCSRF()
	}
	return fmt.Errorf("login: target status %d body %v; initial status %d body %v", st, body, st0, body0)
}

func (c *Client) fetchCSRF() error {
	st, raw := c.doRaw("GET", "/api/v1/auth/csrf", nil, false)
	if st != 200 {
		return fmt.Errorf("GET csrf: status %d, body %v", st, raw)
	}
	var parsed map[string]any
	_ = json.Unmarshal(raw, &parsed)
	tok, _ := parsed["csrf_token"].(string)
	if tok == "" {
		return fmt.Errorf("csrf_token kosong: %s", raw)
	}
	c.csrf = tok
	return nil
}

// --- Util parsing ---

func getPath(m map[string]any, keys ...string) any {
	var cur any = m
	for _, k := range keys {
		mm, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = mm[k]
	}
	return cur
}

func getStr(m map[string]any, keys ...string) string {
	s, _ := getPath(m, keys...).(string)
	return s
}
