// Tes paket users (Manajemen Pengguna staf, UC20).
//
// Tes validasi murni selalu jalan (tanpa DB). Tes endpoint live memakai
// PostgreSQL nyata dan di-skip bila TEST_DATABASE_URL tidak diset
// (pola yang sama seperti kartu_pdf_test.go):
//
//	TEST_DATABASE_URL=<redacted>
//	  go test ./internal/users/ -run TestPenggunaLive -v
package users

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"cbt-smk-v2/backend/internal/auth"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// --- tes validasi murni (tanpa DB) ------------------------------------------

func TestValidasiRoleStaf(t *testing.T) {
	for _, role := range []string{"admin", "proktor", "guru", "pengawas", "penguji_ukk"} {
		if msg := ValidasiRoleStaf(role); msg != "" {
			t.Errorf("role staf %q harus valid, dapat %q", role, msg)
		}
	}
	if msg := ValidasiRoleStaf("siswa"); msg == "" {
		t.Error("role siswa harus ditolak")
	}
	if msg := ValidasiRoleStaf("kepsek"); msg == "" {
		t.Error("role tak dikenal harus ditolak")
	}
	if msg := ValidasiRoleStaf(""); msg == "" {
		t.Error("role kosong harus ditolak")
	}
}

func TestValidasiEmail(t *testing.T) {
	if msg := ValidasiEmail(""); msg != "" {
		t.Errorf("email kosong (opsional) harus valid, dapat %q", msg)
	}
	if msg := ValidasiEmail("admin@sekolah.id"); msg != "" {
		t.Errorf("email valid ditolak: %q", msg)
	}
	for _, bad := range []string{"bukan-email", "a@b", "Admin <a@b.co>", "a @b.co", "a@b .co"} {
		if msg := ValidasiEmail(bad); msg == "" {
			t.Errorf("email %q harus ditolak", bad)
		}
	}
}

func TestValidasiUsernameDanPassword(t *testing.T) {
	if msg := ValidasiUsername(""); msg == "" {
		t.Error("username kosong harus ditolak")
	}
	if msg := ValidasiUsername("ada spasi"); msg == "" {
		t.Error("username ber-spasi harus ditolak")
	}
	if msg := ValidasiUsername("guru_01"); msg != "" {
		t.Errorf("username valid ditolak: %q", msg)
	}
	if msg := ValidasiPassword("pendek"); msg == "" {
		t.Error("password < 8 karakter harus ditolak")
	}
	if msg := ValidasiPassword("12345678"); msg != "" {
		t.Errorf("password 8 karakter harus lolos: %q", msg)
	}
}

// --- tes endpoint live (PostgreSQL nyata) ----------------------------------

const (
	penggunaAdmin1 = "11111111-1111-1111-1111-111111111111"
	penggunaAdmin2 = "22222222-2222-2222-2222-222222222222"
	penggunaGuru1  = "33333333-3333-3333-3333-333333333333"
	penggunaSiswa1 = "44444444-4444-4444-4444-444444444444"
)

func setupPenggunaDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL tidak diset — skip tes endpoint pengguna.")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("koneksi DB: %v", err)
	}
	ddl := []string{
		`DROP TABLE IF EXISTS sessions, audit_log, users CASCADE`,
		`CREATE TABLE users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(), nama TEXT NOT NULL, email TEXT UNIQUE, username TEXT UNIQUE,
			password_hash TEXT NOT NULL, role TEXT NOT NULL,
			must_change_password BOOLEAN NOT NULL DEFAULT TRUE,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			terakhir_masuk TIMESTAMPTZ NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			CHECK (email IS NOT NULL OR username IS NOT NULL))`,
		`CREATE TABLE audit_log (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(), actor_id UUID,
			aksi TEXT NOT NULL, entitas TEXT NOT NULL, entitas_id TEXT,
			detail JSONB NOT NULL DEFAULT '{}'::jsonb, ip TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now())`,
		`CREATE TABLE sessions (
			id UUID PRIMARY KEY, user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role TEXT NOT NULL, csrf_token TEXT NOT NULL, ip TEXT, user_agent TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			expires_at TIMESTAMPTZ NOT NULL, last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now())`,
	}
	for _, q := range ddl {
		if _, err := pool.Exec(ctx, q); err != nil {
			t.Fatalf("DDL: %v", err)
		}
	}
	hash, err := auth.HashPassword("PasswordAwal123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	seed := []string{
		fmt.Sprintf(`INSERT INTO users (id, nama, email, username, password_hash, role, must_change_password, is_active)
			VALUES ('%s','Admin Satu','admin1@sekolah.id','admin1','%s','admin',FALSE,TRUE),
			       ('%s','Admin Dua','admin2@sekolah.id','admin2','%s','admin',FALSE,TRUE),
			       ('%s','Guru Satu','guru1@sekolah.id','guru1','%s','guru',FALSE,TRUE),
			       ('%s','Siswa Satu','siswa1@sekolah.id','siswa1','%s','siswa',TRUE,TRUE)`,
			penggunaAdmin1, hash, penggunaAdmin2, hash, penggunaGuru1, hash, penggunaSiswa1, hash),
		// Satu sesi milik guru (untuk membuktikan reset menghapus sesi).
		fmt.Sprintf(`INSERT INTO sessions (id, user_id, role, csrf_token, expires_at)
			VALUES (gen_random_uuid(), '%s', 'guru', 'csrf-tes', now() + interval '1 hour')`, penggunaGuru1),
	}
	for _, q := range seed {
		if _, err := pool.Exec(ctx, q); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	return pool
}

// doPengguna memanggil endpoint dengan context auth palsu dari header
// (pola kartu_pdf_test.go): role admin lolos semua permission via
// audit.HasPermission; role lain hanya punya permission yang disebut
// di header X-Test-Perms.
func doPengguna(t *testing.T, e *echo.Echo, method, path, userID, role, perms string, body any) (int, map[string]any) {
	t.Helper()
	var reqBody *bytes.Reader
	if body == nil {
		reqBody = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reqBody = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User", userID)
	req.Header.Set("X-Test-Role", role)
	req.Header.Set("X-Test-Perms", perms)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	var parsed map[string]any
	if len(rec.Body.Bytes()) > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
			t.Fatalf("%s %s: body bukan JSON: %s", method, path, rec.Body.Bytes())
		}
	}
	return rec.Code, parsed
}

func errCode(body map[string]any) string {
	if e, ok := body["error"].(map[string]any); ok {
		if c, ok := e["code"].(string); ok {
			return c
		}
	}
	return ""
}

func auditCount(t *testing.T, pool *pgxpool.Pool, aksi string) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM audit_log WHERE aksi = $1`, aksi).Scan(&n); err != nil {
		t.Fatalf("audit count %s: %v", aksi, err)
	}
	return n
}

func TestPenggunaLive(t *testing.T) {
	pool := setupPenggunaDB(t)
	defer pool.Close()

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Set("auth_user_id", c.Request().Header.Get("X-Test-User"))
			c.Set("auth_role", c.Request().Header.Get("X-Test-Role"))
			perms := c.Request().Header.Get("X-Test-Perms")
			if perms == "" {
				c.Set("auth_permissions", []string{})
			} else {
				c.Set("auth_permissions", strings.Split(perms, ","))
			}
			return next(c)
		}
	})
	Register(e, pool, nil, nil)
	ctx := context.Background()

	// 1. List sebagai admin: hanya staf (siswa TIDAK tampil).
	status, body := doPengguna(t, e, "GET", "/api/v1/pengguna", penggunaAdmin1, "admin", "", nil)
	if status != http.StatusOK {
		t.Fatalf("list admin: status %d, body %v", status, body)
	}
	items, _ := body["data"].([]any)
	if len(items) != 3 {
		t.Fatalf("list harus 3 akun staf, dapat %d: %v", len(items), body)
	}
	for _, it := range items {
		m := it.(map[string]any)
		if m["role"] == "siswa" {
			t.Fatalf("list tidak boleh memuat siswa: %v", m)
		}
	}

	// 1b. Filter role=guru -> hanya guru.
	status, body = doPengguna(t, e, "GET", "/api/v1/pengguna?role=guru", penggunaAdmin1, "admin", "", nil)
	if status != http.StatusOK {
		t.Fatalf("list filter guru: status %d", status)
	}
	if got := len(body["data"].([]any)); got != 1 {
		t.Fatalf("filter guru harus 1, dapat %d", got)
	}

	// 2. Non-admin tanpa permission -> 403 (uji permission).
	status, _ = doPengguna(t, e, "GET", "/api/v1/pengguna", penggunaGuru1, "guru", "", nil)
	if status != http.StatusForbidden {
		t.Fatalf("guru tanpa pengguna.read harus 403, dapat %d", status)
	}
	status, _ = doPengguna(t, e, "POST", "/api/v1/pengguna", penggunaGuru1, "guru", "pengguna.read",
		map[string]any{"username": "x", "nama": "X", "role": "guru", "password": "12345678"})
	if status != http.StatusForbidden {
		t.Fatalf("guru tanpa pengguna.create harus 403, dapat %d", status)
	}

	// 3. Buat OK -> 201, must_change_password=true.
	status, body = doPengguna(t, e, "POST", "/api/v1/pengguna", penggunaAdmin1, "admin", "",
		map[string]any{"username": "proktor1", "nama": "Proktor Satu", "email": "proktor1@sekolah.id", "role": "proktor", "password": "ProktorKuat123"})
	if status != http.StatusCreated {
		t.Fatalf("create: status %d, body %v", status, body)
	}
	baru := body["data"].(map[string]any)
	baruID, _ := baru["id"].(string)
	if baru["must_change_password"] != true {
		t.Fatalf("akun baru harus must_change_password=true: %v", baru)
	}
	if baru["terakhir_masuk"] != nil {
		t.Fatalf("akun baru harus terakhir_masuk=null: %v", baru)
	}
	var hashBaru string
	if err := pool.QueryRow(ctx, `SELECT password_hash FROM users WHERE id = $1`, baruID).Scan(&hashBaru); err != nil {
		t.Fatalf("baca hash baru: %v", err)
	}
	if ok, _ := auth.VerifyPassword("ProktorKuat123", hashBaru); !ok {
		t.Fatal("password akun baru harus terverifikasi terhadap hash Argon2id")
	}
	if auditCount(t, pool, "pengguna.create") != 1 {
		t.Fatal("audit pengguna.create harus tercatat 1x")
	}

	// 4. Duplikat username (persis + beda huruf besar) -> 409.
	for _, dup := range []string{"proktor1", "PROKTOR1"} {
		status, body = doPengguna(t, e, "POST", "/api/v1/pengguna", penggunaAdmin1, "admin", "",
			map[string]any{"username": dup, "nama": "Duplikat", "role": "guru", "password": "12345678"})
		if status != http.StatusConflict {
			t.Fatalf("duplikat username %q harus 409, dapat %d: %v", dup, status, body)
		}
	}
	// 4b. Duplikat email -> 409.
	status, _ = doPengguna(t, e, "POST", "/api/v1/pengguna", penggunaAdmin1, "admin", "",
		map[string]any{"username": "lain1", "nama": "Lain", "email": "proktor1@sekolah.id", "role": "guru", "password": "12345678"})
	if status != http.StatusConflict {
		t.Fatalf("duplikat email harus 409, dapat %d", status)
	}

	// 5. Role siswa ditolak -> 400.
	status, body = doPengguna(t, e, "POST", "/api/v1/pengguna", penggunaAdmin1, "admin", "",
		map[string]any{"username": "siswa-baru", "nama": "Siswa Baru", "role": "siswa", "password": "12345678"})
	if status != http.StatusBadRequest {
		t.Fatalf("role siswa harus 400, dapat %d: %v", status, body)
	}

	// 6. Detail siswa -> 404 (anti-IDOR); detail staf -> 200.
	status, _ = doPengguna(t, e, "GET", "/api/v1/pengguna/"+penggunaSiswa1, penggunaAdmin1, "admin", "", nil)
	if status != http.StatusNotFound {
		t.Fatalf("detail siswa harus 404, dapat %d", status)
	}
	status, body = doPengguna(t, e, "GET", "/api/v1/pengguna/"+penggunaGuru1, penggunaAdmin1, "admin", "", nil)
	if status != http.StatusOK {
		t.Fatalf("detail guru harus 200, dapat %d", status)
	}

	// 7. PUT: username tidak bisa diubah (dikirim pun diabaikan).
	status, body = doPengguna(t, e, "PUT", "/api/v1/pengguna/"+penggunaGuru1, penggunaAdmin1, "admin", "",
		map[string]any{"username": "guru-diubah", "nama": "Guru Satu Baru", "email": "guru1baru@sekolah.id", "role": "guru"})
	if status != http.StatusOK {
		t.Fatalf("PUT guru: status %d, body %v", status, body)
	}
	upd := body["data"].(map[string]any)
	if upd["username"] != "guru1" {
		t.Fatalf("username harus tetap guru1, dapat %v", upd["username"])
	}
	if upd["nama"] != "Guru Satu Baru" {
		t.Fatalf("nama harus berubah: %v", upd)
	}
	if auditCount(t, pool, "pengguna.update") != 1 {
		t.Fatal("audit pengguna.update harus tercatat 1x")
	}

	// 8. Reset password guru -> must_change=true + sesi terhapus.
	status, body = doPengguna(t, e, "POST", "/api/v1/pengguna/"+penggunaGuru1+"/reset-password", penggunaAdmin1, "admin", "",
		map[string]any{"password_baru": "GuruBaru456"})
	if status != http.StatusOK {
		t.Fatalf("reset-password: status %d, body %v", status, body)
	}
	var mustChange bool
	if err := pool.QueryRow(ctx, `SELECT must_change_password FROM users WHERE id = $1`, penggunaGuru1).Scan(&mustChange); err != nil || !mustChange {
		t.Fatalf("reset harus must_change_password=true (err %v, val %v)", err, mustChange)
	}
	var sesi int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM sessions WHERE user_id = $1`, penggunaGuru1).Scan(&sesi); err != nil || sesi != 0 {
		t.Fatalf("reset harus menghapus semua sesi (sisa %d, err %v)", sesi, err)
	}
	var hashReset string
	_ = pool.QueryRow(ctx, `SELECT password_hash FROM users WHERE id = $1`, penggunaGuru1).Scan(&hashReset)
	if ok, _ := auth.VerifyPassword("GuruBaru456", hashReset); !ok {
		t.Fatal("password hasil reset harus terverifikasi")
	}
	if auditCount(t, pool, "pengguna.reset-password") != 1 {
		t.Fatal("audit pengguna.reset-password harus tercatat 1x")
	}

	// 9. Nonaktifkan diri ditolak -> 403.
	status, body = doPengguna(t, e, "POST", "/api/v1/pengguna/"+penggunaAdmin1+"/toggle-aktif", penggunaAdmin1, "admin", "",
		map[string]any{"is_active": false})
	if status != http.StatusForbidden {
		t.Fatalf("nonaktifkan diri harus 403, dapat %d: %v", status, body)
	}

	// 10. Demosi admin2 (masih ada admin1) -> 200; lalu demosi admin1
	// (admin terakhir) -> 409; lalu nonaktifkan admin1 -> 409.
	status, _ = doPengguna(t, e, "PUT", "/api/v1/pengguna/"+penggunaAdmin2, penggunaAdmin1, "admin", "",
		map[string]any{"nama": "Admin Dua", "email": "admin2@sekolah.id", "role": "proktor"})
	if status != http.StatusOK {
		t.Fatalf("demosi admin2 harus 200 (admin1 masih ada), dapat %d", status)
	}
	status, body = doPengguna(t, e, "PUT", "/api/v1/pengguna/"+penggunaAdmin1, penggunaAdmin1, "admin", "",
		map[string]any{"nama": "Admin Satu", "email": "admin1@sekolah.id", "role": "proktor"})
	if status != http.StatusConflict || errCode(body) != "admin_terakhir" {
		t.Fatalf("demosi admin terakhir harus 409 admin_terakhir, dapat %d: %v", status, body)
	}
	status, body = doPengguna(t, e, "POST", "/api/v1/pengguna/"+penggunaAdmin1+"/toggle-aktif", penggunaAdmin2, "proktor", "pengguna.toggle-aktif",
		map[string]any{"is_active": false})
	if status != http.StatusConflict || errCode(body) != "admin_terakhir" {
		t.Fatalf("nonaktifkan admin terakhir harus 409 admin_terakhir, dapat %d: %v", status, body)
	}
	// Actor proktor dengan permission eksplisit lolos guard permission
	// (dibuktikan status 409 dari guard bisnis, bukan 403).

	// 11. Toggle nonaktifkan + aktifkan kembali akun staf biasa -> 200.
	status, body = doPengguna(t, e, "POST", "/api/v1/pengguna/"+baruID+"/toggle-aktif", penggunaAdmin1, "admin", "",
		map[string]any{"is_active": false})
	if status != http.StatusOK {
		t.Fatalf("toggle nonaktif proktor harus 200, dapat %d: %v", status, body)
	}
	if body["data"].(map[string]any)["is_active"] != false {
		t.Fatalf("is_active harus false: %v", body)
	}
	status, _ = doPengguna(t, e, "POST", "/api/v1/pengguna/"+baruID+"/toggle-aktif", penggunaAdmin1, "admin", "",
		map[string]any{"is_active": true})
	if status != http.StatusOK {
		t.Fatalf("toggle aktif kembali harus 200, dapat %d", status)
	}
	if auditCount(t, pool, "pengguna.toggle-aktif") != 2 {
		t.Fatal("audit pengguna.toggle-aktif harus tercatat 2x")
	}
}
