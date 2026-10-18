// Tes live Master Data — HANYA jalan bila env CBT_TEST_DATABASE_URL diisi.
//
//	CBT_TEST_DATABASE_URL=postgres://user:pass@127.0.0.1:5432/cbt_master_test \
//	  go test ./internal/master/ -run TestMasterLive -v
//
// Tanpa env tersebut tes di-skip (aman untuk `go test ./...` di CI).
// Tes memakai database khusus (BUKAN database produksi) dan membersihkan
// tabelnya sendiri. Skema users/audit_log dibuat minimal di sini;
// users memakai kolom role tunggal (tidak ada user_roles).
package master

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

func TestMasterLive(t *testing.T) {
	dsn := os.Getenv("CBT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("CBT_TEST_DATABASE_URL tidak diisi — tes live dilewati")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	// Idempotent: bersihkan sisa run sebelumnya agar tes bisa diulang.
	_, _ = pool.Exec(ctx, `DROP TABLE IF EXISTS kisi_kisi, kompetensi, siswa, guru, mapel, kelas, jurusan, audit_log, users`)
	schema, err := os.ReadFile("../../migrations/0002_master.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			nama TEXT NOT NULL DEFAULT '',
			username TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'siswa'
				CHECK (role IN ('admin','proktor','guru','pengawas','siswa','penguji_ukk')),
			must_change_password BOOLEAN NOT NULL DEFAULT TRUE,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now());
		CREATE TABLE IF NOT EXISTS audit_log (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(), actor_id UUID, aksi TEXT,
			entitas TEXT, entitas_id TEXT, detail JSONB, ip TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now());
	`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, string(schema)); err != nil {
		t.Fatal(err)
	}
	// Tiruan minimal tabel Tim 3 (0003) — dibuat SETELAH 0002 agar FK ke
	// mapel/kompetensi/kisi_kisi valid. Dipakai menguji guard hapus yang
	// merujuk bank_soal/soal terhadap tabel nyata, bukan "does not exist".
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS bank_soal (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			mapel_id UUID REFERENCES mapel(id), guru_id UUID REFERENCES users(id));
		CREATE TABLE IF NOT EXISTS soal (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			kompetensi_id UUID REFERENCES kompetensi(id),
			kisi_kisi_id UUID REFERENCES kisi_kisi(id));
	`); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DROP TABLE IF EXISTS soal, bank_soal, kisi_kisi, kompetensi, siswa, guru, mapel, kelas, jurusan, audit_log, users`)
	})

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Set("auth_user_id", "00000000-0000-0000-0000-000000000001")
			c.Set("auth_role", "admin")
			return next(c)
		}
	})
	Register(e, pool, db.New(pool), &config.Config{})

	call := func(method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec
	}
	mustID := func(rec *httptest.ResponseRecorder) string {
		var env map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
			t.Fatalf("body bukan JSON: %v (%s)", err, rec.Body.String())
		}
		d, _ := env["data"].(map[string]any)
		id, _ := d["id"].(string)
		if id == "" {
			t.Fatalf("tidak ada id di respons: %s", rec.Body.String())
		}
		return id
	}

	// Jurusan + duplikat + validasi
	rec := call("POST", "/api/v1/master/jurusan", `{"kode":"RPL","nama":"Rekayasa Perangkat Lunak"}`)
	if rec.Code != 201 {
		t.Fatalf("create jurusan: %d %s", rec.Code, rec.Body.String())
	}
	jurusanID := mustID(rec)
	if rec := call("POST", "/api/v1/master/jurusan", `{"kode":"RPL","nama":"Duplikat"}`); rec.Code != 409 {
		t.Fatalf("duplikat kode jurusan harus 409, dapat %d %s", rec.Code, rec.Body.String())
	}
	if rec := call("POST", "/api/v1/master/jurusan", `{"kode":"","nama":"X"}`); rec.Code != 422 {
		t.Fatalf("kode kosong harus 422, dapat %d", rec.Code)
	}

	// Kelas + validasi tingkat
	rec = call("POST", "/api/v1/master/kelas",
		fmt.Sprintf(`{"jurusan_id":%q,"tingkat":11,"nama":"XI RPL 1","tahun_ajaran":"2026/2027"}`, jurusanID))
	if rec.Code != 201 {
		t.Fatalf("create kelas: %d %s", rec.Code, rec.Body.String())
	}
	kelasID := mustID(rec)
	if rec := call("POST", "/api/v1/master/kelas",
		fmt.Sprintf(`{"jurusan_id":%q,"tingkat":9,"nama":"IX X","tahun_ajaran":"2026/2027"}`, jurusanID)); rec.Code != 422 {
		t.Fatalf("tingkat 9 harus 422, dapat %d", rec.Code)
	}

	// Mapel umum + produktif (wajib jurusan) + filter jenis
	if rec := call("POST", "/api/v1/master/mapel", `{"kode":"MTK","nama":"Matematika","jenis":"umum"}`); rec.Code != 201 {
		t.Fatalf("create mapel umum: %d %s", rec.Code, rec.Body.String())
	}
	if rec := call("POST", "/api/v1/master/mapel", `{"kode":"PROG","nama":"Pemrograman","jenis":"produktif"}`); rec.Code != 422 {
		t.Fatalf("produktif tanpa jurusan harus 422, dapat %d %s", rec.Code, rec.Body.String())
	}
	rec = call("POST", "/api/v1/master/mapel",
		fmt.Sprintf(`{"kode":"PROG","nama":"Pemrograman","jenis":"produktif","jurusan_id":%q,"kelas_tingkat":11}`, jurusanID))
	if rec.Code != 201 {
		t.Fatalf("create mapel produktif: %d %s", rec.Code, rec.Body.String())
	}
	mapelID := mustID(rec)
	rec = call("GET", "/api/v1/master/mapel?jenis=produktif", "")
	var env map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &env)
	if env["meta"].(map[string]any)["total"].(float64) != 1 {
		t.Fatalf("filter jenis produktif harus 1, dapat %s", rec.Body.String())
	}

	// Guru
	rec = call("POST", "/api/v1/master/guru",
		fmt.Sprintf(`{"nip":"19800101","nama":"Budi Guru","mapel_ids":[%q]}`, mapelID))
	if rec.Code != 201 {
		t.Fatalf("create guru: %d %s", rec.Code, rec.Body.String())
	}

	// Siswa manual -> akun otomatis (Argon2id + role siswa + must_change)
	rec = call("POST", "/api/v1/master/siswa",
		fmt.Sprintf(`{"nisn":"1234567890","nis":"001","nama":"Andi","kelas_id":%q,"jurusan_id":%q,"tahun_masuk":2026}`, kelasID, jurusanID))
	if rec.Code != 201 {
		t.Fatalf("create siswa: %d %s", rec.Code, rec.Body.String())
	}
	siswaID := mustID(rec)
	var hash string
	var mustChange bool
	var userID *string
	if err := pool.QueryRow(ctx,
		`SELECT password_hash, must_change_password FROM users WHERE username='1234567890'`).Scan(&hash, &mustChange); err != nil {
		t.Fatalf("akun siswa tidak dibuat: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$") || !mustChange {
		t.Fatalf("hash harus Argon2id + must_change=true, dapat %q %v", hash, mustChange)
	}
	if err := pool.QueryRow(ctx, `SELECT user_id FROM siswa WHERE id=$1`, siswaID).Scan(&userID); err != nil || userID == nil || *userID == "" {
		t.Fatalf("siswa.user_id harus terisi: %v %v", err, userID)
	}
	var role string
	if err := pool.QueryRow(ctx,
		`SELECT role FROM users WHERE id=$1`, *userID).Scan(&role); err != nil || role != "siswa" {
		t.Fatalf("users.role siswa harus 'siswa': %v %q", err, role)
	}

	// Kompetensi + kisi-kisi + guard hapus
	rec = call("POST", "/api/v1/master/kompetensi",
		fmt.Sprintf(`{"mapel_id":%q,"kode":"KOMP-PROG-01","deskripsi":"Menerapkan pemrograman dasar"}`, mapelID))
	if rec.Code != 201 {
		t.Fatalf("create kompetensi: %d %s", rec.Code, rec.Body.String())
	}
	kompID := mustID(rec)
	rec = call("POST", "/api/v1/master/kisi-kisi",
		fmt.Sprintf(`{"mapel_id":%q,"kelas_tingkat":11,"kelompok":"Pemrograman dasar","kompetensi_id":%q,"jumlah_soal":5}`, mapelID, kompID))
	if rec.Code != 201 {
		t.Fatalf("create kisi-kisi: %d %s", rec.Code, rec.Body.String())
	}
	kisiID := mustID(rec)
	if rec := call("DELETE", "/api/v1/master/kompetensi/"+kompID, ""); rec.Code != 409 {
		t.Fatalf("hapus kompetensi yang dipakai kisi-kisi harus 409, dapat %d", rec.Code)
	}
	if rec := call("DELETE", "/api/v1/master/kisi-kisi/"+kisiID, ""); rec.Code != 200 {
		t.Fatalf("hapus kisi-kisi: %d %s", rec.Code, rec.Body.String())
	}
	if rec := call("DELETE", "/api/v1/master/kompetensi/"+kompID, ""); rec.Code != 200 {
		t.Fatalf("hapus kompetensi bebas: %d %s", rec.Code, rec.Body.String())
	}

	// Guard hapus terhadap bank_soal/soal (tabel Tim 3, nyata di tes ini).
	if _, err := pool.Exec(ctx, `INSERT INTO bank_soal (mapel_id) VALUES ($1)`, mapelID); err != nil {
		t.Fatal(err)
	}
	if rec := call("DELETE", "/api/v1/master/mapel/"+mapelID, ""); rec.Code != 409 {
		t.Fatalf("hapus mapel yang dipakai bank soal harus 409, dapat %d %s", rec.Code, rec.Body.String())
	}
	if _, err := pool.Exec(ctx, `DELETE FROM bank_soal WHERE mapel_id = $1`, mapelID); err != nil {
		t.Fatal(err)
	}
	rec = call("POST", "/api/v1/master/kompetensi",
		fmt.Sprintf(`{"mapel_id":%q,"kode":"KOMP-PROG-02","deskripsi":"Guard soal"}`, mapelID))
	if rec.Code != 201 {
		t.Fatalf("create kompetensi 2: %d %s", rec.Code, rec.Body.String())
	}
	komp2ID := mustID(rec)
	rec = call("POST", "/api/v1/master/kisi-kisi",
		fmt.Sprintf(`{"mapel_id":%q,"kelas_tingkat":11,"kelompok":"Guard soal","kompetensi_id":%q,"jumlah_soal":2}`, mapelID, komp2ID))
	if rec.Code != 201 {
		t.Fatalf("create kisi-kisi 2: %d %s", rec.Code, rec.Body.String())
	}
	kisi2ID := mustID(rec)
	if _, err := pool.Exec(ctx, `INSERT INTO soal (kompetensi_id, kisi_kisi_id) VALUES ($1, $2)`, komp2ID, kisi2ID); err != nil {
		t.Fatal(err)
	}
	if rec := call("DELETE", "/api/v1/master/kompetensi/"+komp2ID, ""); rec.Code != 409 {
		t.Fatalf("hapus kompetensi yang dipakai soal harus 409, dapat %d", rec.Code)
	}
	if rec := call("DELETE", "/api/v1/master/kisi-kisi/"+kisi2ID, ""); rec.Code != 409 {
		t.Fatalf("hapus kisi-kisi yang dipakai soal harus 409, dapat %d", rec.Code)
	}
	if _, err := pool.Exec(ctx, `DELETE FROM soal WHERE kompetensi_id = $1`, komp2ID); err != nil {
		t.Fatal(err)
	}
	if rec := call("DELETE", "/api/v1/master/kisi-kisi/"+kisi2ID, ""); rec.Code != 200 {
		t.Fatalf("hapus kisi-kisi bebas: %d %s", rec.Code, rec.Body.String())
	}
	if rec := call("DELETE", "/api/v1/master/kompetensi/"+komp2ID, ""); rec.Code != 200 {
		t.Fatalf("hapus kompetensi bebas 2: %d %s", rec.Code, rec.Body.String())
	}

	// Import CSV: 1 valid (kelas via nama), 1 NISN duplikat, 1 kelas tak ada, 1 nama kosong
	csvBody := "nama,nisn,nis,kelas_kode\nBella,2234567890,002,XI RPL 1\nCaca,1234567890,003,XI RPL 1\nDodi,3234567890,004,Kelas Fiktif\n,4234567890,005,XI RPL 1\n"
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, _ := w.CreateFormFile("file", "siswa.csv")
	_, _ = fw.Write([]byte(csvBody))
	_ = w.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/master/siswa/import", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req)
	if rec2.Code != 200 {
		t.Fatalf("import: %d %s", rec2.Code, rec2.Body.String())
	}
	var imp map[string]any
	_ = json.Unmarshal(rec2.Body.Bytes(), &imp)
	d := imp["data"].(map[string]any)
	if d["sukses"].(float64) != 1 || d["gagal"].(float64) != 3 {
		t.Fatalf("import harus sukses=1 gagal=3, dapat %s", rec2.Body.String())
	}
	errs := d["errors"].([]any)
	if len(errs) != 3 {
		t.Fatalf("errors harus 3 baris, dapat %d", len(errs))
	}
	if int(errs[0].(map[string]any)["baris"].(float64)) != 3 {
		t.Fatalf("baris error pertama harus 3: %s", rec2.Body.String())
	}

	// List + filter siswa (latihan SQL list dengan parameter).
	rec = call("GET", "/api/v1/master/siswa?kelas_id="+kelasID, "")
	_ = json.Unmarshal(rec.Body.Bytes(), &env)
	if env["meta"].(map[string]any)["total"].(float64) != 2 {
		t.Fatalf("filter siswa per kelas harus total 2, dapat %s", rec.Body.String())
	}
	rec = call("GET", "/api/v1/master/siswa?q=Bella", "")
	_ = json.Unmarshal(rec.Body.Bytes(), &env)
	if env["meta"].(map[string]any)["total"].(float64) != 1 {
		t.Fatalf("pencarian siswa Bella harus total 1, dapat %s", rec.Body.String())
	}

	// Guard hapus jurusan/kelas
	if rec := call("DELETE", "/api/v1/master/jurusan/"+jurusanID, ""); rec.Code != 409 {
		t.Fatalf("hapus jurusan dipakai harus 409, dapat %d", rec.Code)
	}
	if rec := call("DELETE", "/api/v1/master/kelas/"+kelasID, ""); rec.Code != 409 {
		t.Fatalf("hapus kelas dipakai siswa harus 409, dapat %d", rec.Code)
	}

	// PUT (edit) kelas + hapus siswa lalu kelas
	rec = call("PUT", "/api/v1/master/kelas/"+kelasID,
		fmt.Sprintf(`{"jurusan_id":%q,"tingkat":11,"nama":"XI RPL 1 (baru)","tahun_ajaran":"2026/2027"}`, jurusanID))
	if rec.Code != 200 {
		t.Fatalf("edit kelas: %d %s", rec.Code, rec.Body.String())
	}
	if rec := call("GET", "/api/v1/master/siswa/"+siswaID, ""); rec.Code != 200 {
		t.Fatalf("get siswa: %d %s", rec.Code, rec.Body.String())
	}
	if rec := call("DELETE", "/api/v1/master/siswa/"+siswaID, ""); rec.Code != 200 {
		t.Fatalf("hapus siswa: %d %s", rec.Code, rec.Body.String())
	}
	var impID string
	_ = pool.QueryRow(ctx, `SELECT id FROM siswa WHERE nisn='2234567890'`).Scan(&impID)
	if rec := call("DELETE", "/api/v1/master/siswa/"+impID, ""); rec.Code != 200 {
		t.Fatalf("hapus siswa import: %d %s", rec.Code, rec.Body.String())
	}
	if rec := call("DELETE", "/api/v1/master/kelas/"+kelasID, ""); rec.Code != 200 {
		t.Fatalf("hapus kelas bebas: %d %s", rec.Code, rec.Body.String())
	}
	if rec := call("GET", "/api/v1/master/jurusan/00000000-0000-0000-0000-000000000099", ""); rec.Code != 404 {
		t.Fatalf("jurusan tak ada harus 404, dapat %d", rec.Code)
	}
	var auditCount int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_log WHERE entitas IN ('jurusan','kelas','mapel','guru','siswa','kompetensi','kisi_kisi')`).Scan(&auditCount)
	if auditCount < 10 {
		t.Fatalf("audit log harus mencatat aksi master, hanya %d", auditCount)
	}
}
