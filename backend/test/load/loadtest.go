// Program uji beban CBT SMK v2 (Tim QA + Uji Beban).
//
// Skenario: setup via API (1 jurusan, 1 kelas, 1 mapel, 1 bank soal
// N soal PG, 1 paket, 1 jadwal, S siswa via import CSV), lalu S
// goroutine siswa serentak: ganti-password pertama -> login ->
// mulai ujian (token) -> jawab N soal (autosave) -> submit.
//
// Pengukuran: total waktu, throughput (req/detik), latensi
// p50/p95/p99 per endpoint, jumlah error/timeout/5xx, pemakaian
// CPU/mem server (bila SERVER_PID diisi), serta verifikasi jawaban
// tersimpan vs S*N (bila LOAD_DATABASE_URL diisi).
//
// CATATAN RATE LIMIT: server membatasi 100 req/menit PER IP. Di LAN
// sekolah tiap perangkat siswa punya IP sendiri, jadi simulasi ini
// memakai -srcips alamat loopback berbeda (127.0.1.1..N) — satu per
// siswa, round-robin — agar rate limit per-IP tetap berlaku seperti
// di produksi. Request 429 tetap dihitung + di-retry (backoff).
//
// Prasyarat (server + DB scratch milik penguji, BUKAN server tim lain):
//
//	DATABASE_URL=postgres://.../cbt_v2_beban PORT=29988 go run ./cmd/server serve
//	psql $DATABASE_URL -f db/seed/seed.sql
//	go run ./cmd/server create-admin --email <ADMIN_LOGIN> --password <ADMIN_INITIAL_PASSWORD> --nama "Admin QA"
//
// Pemakaian:
//
//	BASE_URL=http://127.0.0.1:29988 ADMIN_LOGIN=... ADMIN_INITIAL_PASSWORD=... \
//	  ADMIN_PASSWORD=... LOAD_DATABASE_URL=postgres://.../cbt_v2_beban \
//	  SERVER_PID=<pid server scratch> \
//	  go run ./test/load [-siswa 300] [-soal 10] [-prefix QB...] [-srcips 30]
//
// Prefix unik per run (default QB+timestamp) agar rerun tidak bentrok
// pada constraint UNIQUE. Cleanup best-effort via API; cara bersih
// tuntas adalah DROP DATABASE scratch (program mencetak perintahnya).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	base      = envOr("BASE_URL", "http://127.0.0.1:29988")
	adminLogin = os.Getenv("ADMIN_LOGIN")
	adminInit  = os.Getenv("ADMIN_INITIAL_PASSWORD")
	adminPass  = os.Getenv("ADMIN_PASSWORD")
	dbURL      = os.Getenv("LOAD_DATABASE_URL")
	serverPID  = os.Getenv("SERVER_PID")

	nSiswa   = flag.Int("siswa", 300, "Jumlah siswa serentak")
	nSoal    = flag.Int("soal", 10, "Jumlah soal PG per paket")
	prefix   = flag.String("prefix", "", "Prefix unik data tes (default: QB+YYYYMMDDHHMMSS)")
	srcIPs   = flag.Int("srcips", 30, "Jumlah alamat IP sumber loopback (127.0.1.1..N) untuk simulasi perangkat berbeda")
	actConc  = flag.Int("actconc", 30, "Konkurensi fase aktivasi (ganti-password + login)")
	doCleanup = flag.Bool("cleanup", true, "Coba hapus data tes via API di akhir (best-effort)")
	maxRetry = flag.Int("maxretry", 8, "Maksimal retry saat 429 rate_limited")
	batch    = flag.Int("batch", 100, "Ukuran batch import CSV siswa (lihat catatan WriteTimeout di bawah)")
)

// stats mengumpulkan latensi per endpoint + penghitung error.
type stats struct {
	mu          sync.Mutex
	lat         map[string][]float64 // label -> latensi ms
	reqTotal    int64
	rateLimited int64 // jumlah respons 429 (sebelum retry berhasil)
	err5xx      int64
	errOther    int64
	timeout     int64
}

func (s *stats) record(label string, ms float64) {
	s.mu.Lock()
	s.lat[label] = append(s.lat[label], ms)
	s.mu.Unlock()
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(p / 100 * float64(len(sorted)))
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	if idx < 0 {
		idx = 0
	}
	return sorted[idx]
}

func (s *stats) report() {
	fmt.Println("\n===== LATENSI PER ENDPOINT (ms) =====")
	fmt.Printf("%-28s %6s %10s %10s %10s %10s\n", "endpoint", "n", "p50", "p95", "p99", "max")
	labels := make([]string, 0, len(s.lat))
	for l := range s.lat {
		labels = append(labels, l)
	}
	sort.Strings(labels)
	for _, l := range labels {
		v := append([]float64(nil), s.lat[l]...)
		sort.Float64s(v)
		max := v[len(v)-1]
		fmt.Printf("%-28s %6d %10.1f %10.1f %10.1f %10.1f\n", l, len(v),
			percentile(v, 50), percentile(v, 95), percentile(v, 99), max)
	}
	fmt.Printf("\nTotal request (tercatat): %d | respons 429 (di-retry): %d | 5xx: %d | error lain: %d | timeout: %d\n",
		s.reqTotal, s.rateLimited, s.err5xx, s.errOther, s.timeout)
}

var st = &stats{lat: map[string][]float64{}}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// --- Klien HTTP (jar + CSRF + retry 429 + ukur latensi) ---

type Client struct {
	http *http.Client
	csrf string
}

func newClient(srcIP string) *Client {
	jar, _ := cookiejar.New(nil)
	tr := &http.Transport{Proxy: http.ProxyFromEnvironment}
	if srcIP != "" {
		local := &net.TCPAddr{IP: net.ParseIP(srcIP)}
		tr.DialContext = (&net.Dialer{LocalAddr: local, Timeout: 30 * time.Second}).DialContext
	}
	return &Client{http: &http.Client{Jar: jar, Timeout: 300 * time.Second, Transport: tr}}
}

// do mengirim request JSON (label untuk statistik). HANYA 429 yang
// di-retry (middleware rate-limit menolak SEBELUM handler berjalan,
// jadi retry 429 aman). Error transport TIDAK di-retry: request
// mungkin sudah diproses server, retry buta bisa mengeksekusi
// ulang aksi non-idempotent (mulai/submit/ganti-password).
func (c *Client) do(label, method, path string, body any, csrf bool) (int, map[string]any) {
	var lastStatus int
	var lastParsed map[string]any
	for attempt := 0; attempt <= *maxRetry; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(500*(1<<uint(attempt-1))) * time.Millisecond
			if backoff > 20*time.Second {
				backoff = 20 * time.Second
			}
			time.Sleep(backoff + time.Duration(rand.Intn(200))*time.Millisecond)
		}
		start := time.Now()
		stCode, raw := c.doOnce(method, path, body, csrf)
		elapsed := float64(time.Since(start).Microseconds()) / 1000.0
		atomic.AddInt64(&st.reqTotal, 1)
		st.record(label, elapsed)
		var parsed map[string]any
		_ = json.Unmarshal(raw, &parsed)
		if parsed == nil {
			parsed = map[string]any{"_raw": string(raw)}
		}
		lastStatus, lastParsed = stCode, parsed
		if stCode == 429 {
			atomic.AddInt64(&st.rateLimited, 1)
			continue
		}
		if stCode >= 500 {
			atomic.AddInt64(&st.err5xx, 1)
		}
		if stCode == 0 {
			atomic.AddInt64(&st.timeout, 1)
			// Jangan retry: request mungkin sudah dieksekusi server.
			return stCode, parsed
		}
		return stCode, parsed
	}
	atomic.AddInt64(&st.errOther, 1)
	return lastStatus, lastParsed
}

func (c *Client) doOnce(method, path string, body any, csrf bool) (int, []byte) {
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

// doMultipart mengirim CSV import siswa (field "file").
// HANYA 429 yang di-retry (ditolak sebelum handler). Error transport
// TIDAK di-retry: import non-idempotent — retry buta mengeksekusi
// ulang import (terbukti pada run QB1: import kedua melaporkan
// 300x "NISN sudah terdaftar" karena import pertama sudah commit).
func (c *Client) doMultipart(label, path, csv string) (int, map[string]any) {
	var lastStatus int
	var lastParsed map[string]any
	for attempt := 0; attempt <= *maxRetry; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(500*(1<<uint(attempt-1))) * time.Millisecond)
		}
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		fw, err := w.CreateFormFile("file", "siswa.csv")
		if err != nil {
			return 0, map[string]any{"_raw": err.Error()}
		}
		_, _ = fw.Write([]byte(csv))
		_ = w.Close()
		start := time.Now()
		req, _ := http.NewRequest("POST", base+path, &buf)
		req.Header.Set("Content-Type", w.FormDataContentType())
		if c.csrf != "" {
			req.Header.Set("X-CSRF-Token", c.csrf)
		}
		resp, err := c.http.Do(req)
		elapsed := float64(time.Since(start).Microseconds()) / 1000.0
		atomic.AddInt64(&st.reqTotal, 1)
		st.record(label, elapsed)
		if err != nil {
			atomic.AddInt64(&st.timeout, 1)
			// Jangan retry: import mungkin sudah commit di server.
			return 0, map[string]any{"_raw": "transport error (tidak di-retry): " + err.Error()}
		}
		raw, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		var parsed map[string]any
		_ = json.Unmarshal(raw, &parsed)
		if parsed == nil {
			parsed = map[string]any{"_raw": string(raw)}
		}
		lastStatus, lastParsed = resp.StatusCode, parsed
		if resp.StatusCode == 429 {
			atomic.AddInt64(&st.rateLimited, 1)
			continue
		}
		if resp.StatusCode >= 500 {
			atomic.AddInt64(&st.err5xx, 1)
		}
		return resp.StatusCode, parsed
	}
	return lastStatus, lastParsed
}

func (c *Client) ensureLogin(login, initial, target string) error {
	stCode, body := c.do("login", "POST", "/api/v1/auth/login", map[string]any{
		"login": login, "password": target,
	}, false)
	if stCode == 200 {
		return c.fetchCSRF()
	}
	// Server memverifikasi password dulu baru flag must_change_password,
	// jadi percobaan gagal pertama menghasilkan 401, bukan 403.
	st0, body0 := c.do("login", "POST", "/api/v1/auth/login", map[string]any{
		"login": login, "password": initial,
	}, false)
	if st0 == 200 {
		return c.fetchCSRF()
	}
	if st0 == 403 && getStr(body0, "error", "code") == "PASSWORD_CHANGE_REQUIRED" {
		st2, body2 := c.do("ganti-password", "POST", "/api/v1/auth/ganti-password", map[string]any{
			"login": login, "password_lama": initial, "password_baru": target,
		}, false)
		if st2 != 200 {
			return fmt.Errorf("ganti-password: status %d, body %v", st2, body2)
		}
		st3, body3 := c.do("login", "POST", "/api/v1/auth/login", map[string]any{
			"login": login, "password": target,
		}, false)
		if st3 != 200 {
			return fmt.Errorf("login setelah ganti password: status %d, body %v", st3, body3)
		}
		return c.fetchCSRF()
	}
	return fmt.Errorf("login: target status %d body %v; initial status %d body %v", stCode, body, st0, body0)
}

func (c *Client) fetchCSRF() error {
	stCode, raw := c.doOnce("GET", "/api/v1/auth/csrf", nil, false)
	if stCode != 200 {
		return fmt.Errorf("GET csrf: status %d, body %s", stCode, raw)
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

// --- Util ---

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

func must(cond bool, format string, args ...any) {
	if !cond {
		fmt.Printf("SETUP GAGAL: "+format+"\n", args...)
		os.Exit(1)
	}
}

// --- Pengukuran resource server (opsional, via /proc) ---

type procSnap struct {
	rssKB  int64
	utime  int64
	stime  int64
	load1  string
	memAvailKB int64
}

func snapProc() procSnap {
	var s procSnap
	if serverPID != "" {
		if b, err := os.ReadFile("/proc/" + serverPID + "/status"); err == nil {
			for _, line := range strings.Split(string(b), "\n") {
				if strings.HasPrefix(line, "VmRSS:") {
					f := strings.Fields(line)
					if len(f) >= 2 {
						s.rssKB, _ = strconv.ParseInt(f[1], 10, 64)
					}
				}
			}
		}
		if b, err := os.ReadFile("/proc/" + serverPID + "/stat"); err == nil {
			f := strings.Fields(string(b))
			if len(f) >= 15 {
				s.utime, _ = strconv.ParseInt(f[13], 10, 64)
				s.stime, _ = strconv.ParseInt(f[14], 10, 64)
			}
		}
	}
	if b, err := os.ReadFile("/proc/loadavg"); err == nil {
		s.load1 = strings.Fields(string(b))[0]
	}
	if b, err := os.ReadFile("/proc/meminfo"); err == nil {
		for _, line := range strings.Split(string(b), "\n") {
			if strings.HasPrefix(line, "MemAvailable:") {
				f := strings.Fields(line)
				if len(f) >= 2 {
					s.memAvailKB, _ = strconv.ParseInt(f[1], 10, 64)
				}
			}
		}
	}
	return s
}

// --- Main ---

func main() {
	flag.Parse()
	runPrefix := *prefix
	if runPrefix == "" {
		runPrefix = "QB" + time.Now().Format("20060102150405")
	}
	if adminLogin == "" || adminInit == "" || adminPass == "" {
		fmt.Fprintln(os.Stderr, "LOAD: env ADMIN_LOGIN, ADMIN_INITIAL_PASSWORD, ADMIN_PASSWORD wajib diisi.")
		os.Exit(2)
	}
	runID := time.Now().Format("060102150405") // 12 digit untuk NISN unik
	fmt.Printf("Uji beban: %d siswa, %d soal, prefix %s, base %s, srcips %d\n",
		*nSiswa, *nSoal, runPrefix, base, *srcIPs)

	totalStart := time.Now()
	before := snapProc()
	admin := newClient("")
	must(admin.ensureLogin(adminLogin, adminInit, adminPass) == nil, "login admin gagal")
	fmt.Println("[setup] admin login OK")

	// --- Setup via API ---
	var jurusanID, kelasID, mapelID, bankID, paketID, jadwalID, token string
	kelasNama := "XI BEBAN " + runPrefix
	stCode, body := admin.do("setup-jurusan", "POST", "/api/v1/master/jurusan",
		map[string]any{"kode": runPrefix + "-J", "nama": "Jurusan Beban " + runPrefix}, true)
	must(stCode == 201, "buat jurusan: %d %v", stCode, body)
	jurusanID = getStr(body, "data", "id")

	stCode, body = admin.do("setup-kelas", "POST", "/api/v1/master/kelas",
		map[string]any{"jurusan_id": jurusanID, "tingkat": 11, "nama": kelasNama, "tahun_ajaran": "2026/2027"}, true)
	must(stCode == 201, "buat kelas: %d %v", stCode, body)
	kelasID = getStr(body, "data", "id")

	stCode, body = admin.do("setup-mapel", "POST", "/api/v1/master/mapel",
		map[string]any{"kode": runPrefix + "-M", "nama": "Mapel Beban " + runPrefix, "jenis": "produktif", "jurusan_id": jurusanID}, true)
	must(stCode == 201, "buat mapel: %d %v", stCode, body)
	mapelID = getStr(body, "data", "id")

	stCode, body = admin.do("setup-bank", "POST", "/api/v1/bank-soal",
		map[string]any{"kode": runPrefix + "-B", "nama": "Bank Beban " + runPrefix, "mapel_id": mapelID, "kelas_tingkat": 11}, true)
	must(stCode == 201, "buat bank: %d %v", stCode, body)
	bankID = getStr(body, "data", "id")

	labels := []string{"A", "B", "C", "D"}
	soalIDs := make([]string, 0, *nSoal)
	for i := 0; i < *nSoal; i++ {
		benar := labels[i%4]
		stCode, body = admin.do("setup-soal", "POST", "/api/v1/bank-soal/"+bankID+"/soal", map[string]any{
			"tipe": "pg", "pertanyaan": fmt.Sprintf("Soal beban #%d: 2 + %d = ?", i+1, i),
			"bobot": 10, "level": "sedang", "status": "aktif",
			"jawaban": map[string]any{"opsi": []any{
				map[string]any{"label": "A", "teks": "Opsi A"},
				map[string]any{"label": "B", "teks": "Opsi B"},
				map[string]any{"label": "C", "teks": "Opsi C"},
				map[string]any{"label": "D", "teks": "Opsi D"},
			}},
			"kunci": map[string]any{"opsi": benar},
		}, true)
		must(stCode == 201, "buat soal %d: %d %v", i+1, stCode, body)
		soalIDs = append(soalIDs, getStr(body, "data", "id"))
	}
	fmt.Printf("[setup] bank %d soal OK (%.1fs)\n", *nSoal, time.Since(totalStart).Seconds())

	stCode, body = admin.do("setup-paket", "POST", "/api/v1/paket", map[string]any{
		"kode": runPrefix + "-P", "nama": "Paket Beban " + runPrefix, "mapel_id": mapelID,
		"jenis": "PTS", "durasi_menit": 60, "jumlah_soal": *nSoal,
		"acak_soal": false, "acak_jawaban": false, "tampil_hasil": true,
	}, true)
	must(stCode == 201, "buat paket: %d %v", stCode, body)
	paketID = getStr(body, "data", "id")
	for _, sid := range soalIDs {
		stCode, body = admin.do("setup-paket-soal", "POST", "/api/v1/paket/"+paketID+"/soal", map[string]any{"soal_id": sid}, true)
		must(stCode == 201, "tambah soal paket: %d %v", stCode, body)
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	stCode, body = admin.do("setup-jadwal", "POST", "/api/v1/jadwal", map[string]any{
		"paket_id": paketID, "kelas_id": kelasID,
		"tanggal": now.Format("2006-01-02"),
		"jam_mulai": now.Add(-5 * time.Minute).Format("15:04:05"),
		"jam_selesai": now.Add(3 * time.Hour).Format("15:04:05"),
		"status": "aktif",
	}, true)
	must(stCode == 201, "buat jadwal: %d %v", stCode, body)
	jadwalID = getStr(body, "data", "id")

	stCode, body = admin.do("setup-token", "POST", "/api/v1/jadwal/"+jadwalID+"/token", map[string]any{}, true)
	must(stCode == 200, "generate token: %d %v", stCode, body)
	token = getStr(body, "data", "token")
	must(len(token) == 6, "token bukan 6 karakter: %q", token)

	// --- Import CSV siswa (PER BATCH) ---
	// Satu import 300 siswa memakan ~50 dtk (300x Argon2id sekuensial)
	// dan MELEBIHI WriteTimeout 30 dtk Echo v5 -> server menutup
	// koneksi, klien menerima EOF, padahal transaksi sudah commit
	// (BUG-02 di docs/BUG-REPORT-QA-BEBAN.md). Import dipecah per
	// batch (default 100 siswa ~17 dtk < 30 dtk) agar setiap respons
	// sampai ke klien. Ini juga workaround yang didokumentasikan
	// untuk operator sampai BUG-02 diperbaiki.
	var csvRows []string
	csvRows = append(csvRows, "nama,nisn,nis,kelas_kode")
	nisns := make([]string, *nSiswa)
	for i := 0; i < *nSiswa; i++ {
		nisn := "98" + runID + fmt.Sprintf("%03d", i+1)
		nisns[i] = nisn
		csvRows = append(csvRows, fmt.Sprintf("Siswa Beban %03d,%s,%03d,%s", i+1, nisn, i+1, kelasNama))
	}
	importStart := time.Now()
	totalSukses := 0
	for b := 0; b < len(csvRows)-1; b += *batch {
		end := b + *batch
		if end > len(csvRows)-1 {
			end = len(csvRows) - 1
		}
		chunk := "nama,nisn,nis,kelas_kode\n" + strings.Join(csvRows[1+b:1+end], "\n") + "\n"
		stCode, body = admin.doMultipart("setup-import", "/api/v1/master/siswa/import", chunk)
		if stCode != 200 {
			fmt.Printf("SETUP GAGAL: import batch %d-%d: %d %v\n", b+1, end, stCode, body)
			os.Exit(1)
		}
		s, _ := getPath(body, "data", "sukses").(float64)
		totalSukses += int(s)
		fmt.Printf("[setup] import batch %d-%d: sukses=%v gagal=%v\n", b+1, end,
			getPath(body, "data", "sukses"), getPath(body, "data", "gagal"))
		if int(s) != end-b {
			fmt.Printf("SETUP GAGAL: import batch %d-%d sukses %v, mau %d. 3 error pertama: ", b+1, end, s, end-b)
			if errs, ok := getPath(body, "data", "errors").([]any); ok {
				for i, e := range errs {
					if i >= 3 {
						break
					}
					fmt.Printf("%v ", e)
				}
			}
			fmt.Println()
			os.Exit(1)
		}
	}
	importDur := time.Since(importStart)
	fmt.Printf("[setup] import siswa total: sukses=%d/%d (%.1fs)\n", totalSukses, *nSiswa, importDur.Seconds())

	// --- Fase aktivasi: ganti-password + login per siswa ---
	// Pass 1: ganti-password (NISN -> password baru) + login.
	// Pass 2 (pemulihan): untuk siswa yang gagal di pass 1, coba
	// login dengan password baru dulu — bila server sudah commit
	// ganti-password tetapi klien hanya melihat EOF (fenomena yang
	// sama seperti BUG-02), login berhasil dan siswa pulih tanpa
	// perlu ganti-password ulang (yang justru akan gagal 401
	// karena password lama sudah tidak berlaku).
	fmt.Printf("[aktivasi] %d siswa, konkurensi %d ...\n", *nSiswa, *actConc)
	actStart := time.Now()
	clients := make([]*Client, *nSiswa)
	var actFail []int
	var actFailMu sync.Mutex
	activate := func(i int, pass int) bool {
		srcIP := fmt.Sprintf("127.0.1.%d", i%*srcIPs+1)
		cl := clients[i]
		if cl == nil {
			cl = newClient(srcIP)
		}
		pw := fmt.Sprintf("Beban%04d!", i+1)
		if pass == 2 {
			if err := cl.ensureLogin(nisns[i], pw, pw); err == nil {
				clients[i] = cl
				return true
			}
			// Login gagal: password mungkin belum terganti — coba
			// ganti-password lagi dari password awal.
		}
		stCode, body := cl.do("aktivasi-ganti-password", "POST", "/api/v1/auth/ganti-password",
			map[string]any{"login": nisns[i], "password_lama": nisns[i], "password_baru": pw}, false)
		if stCode != 200 {
			fmt.Printf("  aktivasi siswa %d pass %d GAGAL ganti-password: status=%d body=%.200v\n", i+1, pass, stCode, body)
			return false
		}
		if err := cl.ensureLogin(nisns[i], pw, pw); err != nil {
			fmt.Printf("  aktivasi siswa %d pass %d GAGAL login: %.200v\n", i+1, pass, err)
			return false
		}
		clients[i] = cl
		return true
	}
	{
		var actWg sync.WaitGroup
		sem := make(chan struct{}, *actConc)
		for i := 0; i < *nSiswa; i++ {
			actWg.Add(1)
			sem <- struct{}{}
			go func(i int) {
				defer actWg.Done()
				defer func() { <-sem }()
				if !activate(i, 1) {
					actFailMu.Lock()
					actFail = append(actFail, i)
					actFailMu.Unlock()
				}
			}(i)
		}
		actWg.Wait()
	}
	if len(actFail) > 0 {
		fmt.Printf("[aktivasi] pass 2 (pemulihan) untuk %d siswa ...\n", len(actFail))
		var stillFail []int
		var actWg2 sync.WaitGroup
		sem := make(chan struct{}, *actConc)
		for _, i := range actFail {
			actWg2.Add(1)
			sem <- struct{}{}
			go func(i int) {
				defer actWg2.Done()
				defer func() { <-sem }()
				if !activate(i, 2) {
					actFailMu.Lock()
					stillFail = append(stillFail, i)
					actFailMu.Unlock()
				}
			}(i)
		}
		actWg2.Wait()
		actFail = stillFail
	}
	actDur := time.Since(actStart)
	fmt.Printf("[aktivasi] selesai %.1fs, gagal %d\n", actDur.Seconds(), len(actFail))
	must(len(actFail) == 0, "aktivasi gagal %d siswa", len(actFail))

	// --- Fase ujian: 300 goroutine serentak ---
	fmt.Printf("[ujian] %d siswa serentak: mulai -> attempt -> %d jawaban -> submit ...\n", *nSiswa, *nSoal)
	examStart := time.Now()
	var (
		mulaiOK, attemptOK, submitOK int64
		jawabanOK                    int64
		kunciLeak                    int64
		examErr                      int64
		firstErr                     sync.Mutex
		firstErrMsg                  string
	)
	var examWg sync.WaitGroup
	for i := 0; i < *nSiswa; i++ {
		examWg.Add(1)
		go func(i int) {
			defer examWg.Done()
			cl := clients[i]
			fail := func(format string, args ...any) {
				atomic.AddInt64(&examErr, 1)
				firstErr.Lock()
				if firstErrMsg == "" {
					firstErrMsg = fmt.Sprintf("siswa %d: "+format, append([]any{i + 1}, args...)...)
				}
				firstErr.Unlock()
			}
			stCode, body := cl.do("mulai", "POST", "/api/v1/ujian/mulai",
				map[string]any{"jadwal_id": jadwalID, "token": token}, true)
			if stCode == 0 {
				// Error transport: mulai bersifat idempotent (attempt
				// berjalan yang sudah ada dikembalikan), aman retry 1x.
				stCode, body = cl.do("mulai-retry", "POST", "/api/v1/ujian/mulai",
					map[string]any{"jadwal_id": jadwalID, "token": token}, true)
			}
			if stCode != 201 {
				fail("mulai: %d %v", stCode, body)
				return
			}
			atomic.AddInt64(&mulaiOK, 1)
			attemptID := getStr(body, "data", "attempt_id")
			stCode, raw := cl.doOnce("GET", "/api/v1/ujian/attempt/"+attemptID, nil, false)
			atomic.AddInt64(&st.reqTotal, 1)
			if stCode != 200 {
				fail("attempt: %d %s", stCode, raw)
				return
			}
			if strings.Contains(string(raw), `"kunci"`) {
				atomic.AddInt64(&kunciLeak, 1)
			}
			atomic.AddInt64(&attemptOK, 1)
			for j := 0; j < *nSoal; j++ {
				// Jawab bergantian benar/salah agar nilai bervariasi.
				var jawab string
				benar := labels[j%4]
				if (i+j)%2 == 0 {
					jawab = benar
				} else {
					jawab = labels[(j+1)%4]
				}
				stCode, body := cl.do("jawaban", "POST", "/api/v1/ujian/attempt/"+attemptID+"/jawaban",
					map[string]any{"soal_id": soalIDs[j], "jawaban": jawab}, true)
				if stCode == 0 {
					// Jawaban adalah upsert — aman retry 1x.
					stCode, body = cl.do("jawaban-retry", "POST", "/api/v1/ujian/attempt/"+attemptID+"/jawaban",
						map[string]any{"soal_id": soalIDs[j], "jawaban": jawab}, true)
				}
				if stCode != 200 {
					fail("jawaban soal %d: %d %v", j+1, stCode, body)
					return
				}
				atomic.AddInt64(&jawabanOK, 1)
			}
			stCode, body = cl.do("submit", "POST", "/api/v1/ujian/attempt/"+attemptID+"/submit", map[string]any{}, true)
			if stCode == 0 {
				// Error transport: submit mungkin sudah commit di
				// server (fenomena BUG-02). Verifikasi via status
				// sebelum menyatakan gagal.
				stCode2, body2 := cl.do("submit-verify", "GET", "/api/v1/ujian/attempt/"+attemptID+"/status", nil, false)
				if stCode2 == 200 && getStr(body2, "data", "status") == "selesai" {
					atomic.AddInt64(&submitOK, 1)
					return
				}
				stCode, body = cl.do("submit-retry", "POST", "/api/v1/ujian/attempt/"+attemptID+"/submit", map[string]any{}, true)
				if stCode == 0 {
					stCode2, body2 := cl.do("submit-verify2", "GET", "/api/v1/ujian/attempt/"+attemptID+"/status", nil, false)
					if stCode2 == 200 && getStr(body2, "data", "status") == "selesai" {
						atomic.AddInt64(&submitOK, 1)
						return
					}
				}
			}
			if stCode != 200 {
				fail("submit: %d %v", stCode, body)
				return
			}
			atomic.AddInt64(&submitOK, 1)
		}(i)
	}
	examWg.Wait()
	examDur := time.Since(examStart)
	after := snapProc()

	// --- Verifikasi jawaban tersimpan (via DB langsung, bila diisi) ---
	var dbAttempt, dbJawaban, dbFull, dbNilai int64 = -1, -1, -1, -1
	if dbURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		pool, err := pgxpool.New(ctx, dbURL)
		if err == nil {
			defer pool.Close()
			_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM attempt WHERE jadwal_id = $1`, jadwalID).Scan(&dbAttempt)
			_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM jawaban j JOIN attempt a ON a.id = j.attempt_id WHERE a.jadwal_id = $1`, jadwalID).Scan(&dbJawaban)
			_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM (SELECT attempt_id FROM jawaban j JOIN attempt a ON a.id = j.attempt_id WHERE a.jadwal_id = $1 GROUP BY attempt_id HAVING COUNT(*) = $2) x`, jadwalID, *nSoal).Scan(&dbFull)
			_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM nilai_akhir WHERE jadwal_id = $1`, jadwalID).Scan(&dbNilai)
		} else {
			fmt.Printf("verifikasi DB gagal (pool): %v\n", err)
		}
	}

	totalDur := time.Since(totalStart)
	fmt.Println("\n===== HASIL UJI BEBAN =====")
	fmt.Printf("Siswa: %d | Soal/siswa: %d | Jawaban diharapkan: %d\n", *nSiswa, *nSoal, *nSiswa**nSoal)
	fmt.Printf("Setup (s.d. import): %.1fs | Aktivasi: %.1fs | Fase ujian: %.1fs | TOTAL: %.1fs\n",
		importDur.Seconds(), actDur.Seconds(), examDur.Seconds(), totalDur.Seconds())
	fmt.Printf("Mulai OK: %d/%d | Attempt OK: %d | Jawaban OK: %d/%d | Submit OK: %d/%d | Error siswa: %d\n",
		mulaiOK, *nSiswa, attemptOK, jawabanOK, *nSiswa**nSoal, submitOK, *nSiswa, examErr)
	if firstErrMsg != "" {
		fmt.Printf("Error pertama: %s\n", firstErrMsg)
	}
	fmt.Printf("Respons attempt yang mengandung \"kunci\": %d (harus 0)\n", kunciLeak)
	if dbURL != "" {
		fmt.Printf("Verifikasi DB: attempt=%d (mau %d) | jawaban=%d (mau %d) | attempt lengkap %d soal=%d | nilai_akhir=%d\n",
			dbAttempt, *nSiswa, dbJawaban, *nSiswa**nSoal, dbFull, *nSiswa, dbNilai)
		if dbJawaban != int64(*nSiswa**nSoal) {
			fmt.Printf("!!! JAWABAN HILANG: tersimpan %d dari %d\n", dbJawaban, *nSiswa**nSoal)
		}
	} else {
		fmt.Println("Verifikasi DB dilewati (LOAD_DATABASE_URL kosong).")
	}
	throughput := float64(st.reqTotal) / totalDur.Seconds()
	fmt.Printf("Throughput keseluruhan: %.2f req/detik (%d request / %.1fs)\n", throughput, st.reqTotal, totalDur.Seconds())
	examThroughput := float64(mulaiOK+attemptOK+jawabanOK+submitOK) / examDur.Seconds()
	fmt.Printf("Throughput fase ujian: %.2f req/detik\n", examThroughput)
	st.report()
	fmt.Println("\n--- Resource server (via /proc, bila SERVER_PID diisi) ---")
	fmt.Printf("SERVER_PID=%q | load1 sebelum=%s sesudah=%s | VmRSS sebelum=%d KB sesudah=%d KB (delta %+d KB)\n",
		serverPID, before.load1, after.load1, before.rssKB, after.rssKB, after.rssKB-before.rssKB)
	fmt.Printf("CPU server (jiffies): utime %d->%d (delta %d) | stime %d->%d (delta %d) | MemAvailable %d->%d KB\n",
		before.utime, after.utime, after.utime-before.utime,
		before.stime, after.stime, after.stime-before.stime,
		before.memAvailKB, after.memAvailKB)

	if *doCleanup {
		fmt.Println("\n--- Cleanup best-effort (kegagalan 409 = guard yang benar) ---")
		del := func(label, path string) {
			stCode, body := admin.do("cleanup-"+label, "DELETE", path, nil, true)
			fmt.Printf("  DELETE %s -> %d %v\n", label, stCode, body)
		}
		del("jadwal", "/api/v1/jadwal/"+jadwalID)
		del("paket", "/api/v1/paket/"+paketID)
		del("bank-soal", "/api/v1/bank-soal/"+bankID)
		for i := 0; i < *nSiswa; i++ {
			// Siswa tidak dihapus satu per satu (409 bila punya attempt) —
			// dicatat saja; bersih tuntas via DROP DATABASE.
			_ = i
			break
		}
		del("mapel", "/api/v1/master/mapel/"+mapelID)
		del("kelas", "/api/v1/master/kelas/"+kelasID)
		del("jurusan", "/api/v1/master/jurusan/"+jurusanID)
	}
	fmt.Printf("\nBersihkan tuntas DB scratch: DROP DATABASE <nama-db-scratch>;  (data prefix %s)\n", runPrefix)

	ok := submitOK == int64(*nSiswa) && jawabanOK == int64(*nSiswa**nSoal) && kunciLeak == 0
	if dbURL != "" && dbJawaban != int64(*nSiswa**nSoal) {
		ok = false
	}
	if !ok {
		fmt.Println("UJI BEBAN: TIDAK LULUS — lihat angka di atas.")
		os.Exit(1)
	}
	fmt.Println("UJI BEBAN: LULUS")
}
