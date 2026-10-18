// Tes kartu peserta PDF — Tim Backend Kartu PDF.
//
// Tes generator (RenderKartuPDF / RenderKartuBulkPDF) berjalan tanpa DB.
// Tes endpoint memakai PostgreSQL nyata dan di-skip bila
// TEST_DATABASE_URL tidak diset (pola yang sama seperti
// exam_e2e_test.go):
//
//	TEST_DATABASE_URL=<redacted>
//	go test ./internal/event/ -run TestKartuPDFEndpoint -count=1 -v
package event

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// ---------------------------------------------------------------------------
// Tes generator (tanpa DB).
// ---------------------------------------------------------------------------

func contohKartuPDF(nJadwal int) kartuPDFData {
	d := kartuPDFData{
		NamaSekolah: "SMK Contoh Bangsa",
		NamaEvent:   "STS Ganjil 2026/2027",
		NamaSiswa:   "Budi Santoso",
		NISN:        "0012345678",
		Kelas:       "XI PPLG 1",
		Jurusan:     "Pengembangan Perangkat Lunak dan Gim",
		Username:    "0012345678",
	}
	for i := 0; i < nJadwal; i++ {
		d.Jadwal = append(d.Jadwal, kartuPDFJadwal{
			NomorPeserta: fmt.Sprintf("STS-GANJIL-26/2026/%03d", i+1),
			Mapel:        fmt.Sprintf("Mapel %d", i+1),
			Paket:        fmt.Sprintf("Paket %d", i+1),
			Tanggal:      "17 Okt 2026",
			Jam:          "07:30 - 09:30",
			Ruangan:      "Ruang 1",
			EventNama:    "STS Ganjil 2026/2027",
		})
	}
	return d
}

func TestRenderKartuPDFDasar(t *testing.T) {
	b, err := RenderKartuPDF(contohKartuPDF(2))
	if err != nil {
		t.Fatalf("RenderKartuPDF: %v", err)
	}
	if !bytes.HasPrefix(b, []byte("%PDF")) {
		t.Fatalf("PDF harus diawali %%PDF, dapat %q", b[:min(20, len(b))])
	}
	// Kompresi dimatikan -> teks harus terbaca langsung di byte stream.
	for _, teks := range []string{"Budi Santoso", "Mapel 1", "Kartu Peserta Ujian", "STS Ganjil 2026/2027", "Token ujian diberikan proktor"} {
		if !bytes.Contains(b, []byte(teks)) {
			t.Fatalf("PDF tidak berisi teks %q", teks)
		}
	}
	if got := hitungHalamanPDF(b); got != 1 {
		t.Fatalf("2 jadwal harus 1 halaman, dapat %d", got)
	}
	// Nomor peserta TIDAK boleh di blok identitas: label "Nomor Peserta"
	// hanya muncul di header tabel (1x), bukan di identitas.
	if got := bytes.Count(b, []byte("Nomor Peserta")); got != 1 {
		t.Fatalf("label 'Nomor Peserta' harus 1x (header tabel saja), dapat %d", got)
	}
}

func TestRenderKartuPDFBanyakJadwalLanjutHalaman2(t *testing.T) {
	b, err := RenderKartuPDF(contohKartuPDF(30))
	if err != nil {
		t.Fatalf("RenderKartuPDF: %v", err)
	}
	if got := hitungHalamanPDF(b); got != 2 {
		t.Fatalf("30 jadwal harus 2 halaman, dapat %d", got)
	}
	// Header tabel diulang di halaman 2 ...
	if got := bytes.Count(b, []byte("Nomor Peserta")); got != 2 {
		t.Fatalf("header tabel harus diulang (2x), dapat %d", got)
	}
	// ... tetapi identitas TIDAK diulang.
	if got := bytes.Count(b, []byte("Jurusan")); got != 1 {
		t.Fatalf("blok identitas harus 1x (tidak diulang), dapat %d", got)
	}
	// Semua jadwal tetap ada.
	if !bytes.Contains(b, []byte("Mapel 30")) {
		t.Fatalf("PDF harus berisi jadwal ke-30")
	}
}

func TestRenderKartuPDFTanpaJadwal(t *testing.T) {
	d := contohKartuPDF(0)
	b, err := RenderKartuPDF(d)
	if err != nil {
		t.Fatalf("RenderKartuPDF: %v", err)
	}
	if !bytes.Contains(b, []byte("Belum ada jadwal")) {
		t.Fatalf("kartu tanpa jadwal harus menampilkan pesan 'Belum ada jadwal'")
	}
	if got := hitungHalamanPDF(b); got != 1 {
		t.Fatalf("kartu tanpa jadwal harus 1 halaman, dapat %d", got)
	}
}

func TestRenderKartuBulkPDF(t *testing.T) {
	kartu1 := contohKartuPDF(1)
	kartu2 := contohKartuPDF(1)
	kartu2.NamaSiswa = "Siti Rahma"
	kartu2.NISN = "0087654321"
	kartu2.Username = "0087654321"
	b, err := RenderKartuBulkPDF([]kartuPDFData{kartu1, kartu2})
	if err != nil {
		t.Fatalf("RenderKartuBulkPDF: %v", err)
	}
	if got := hitungHalamanPDF(b); got != 2 {
		t.Fatalf("bulk 2 siswa harus 2 halaman (1 halaman per siswa), dapat %d", got)
	}
	for _, nama := range []string{"Budi Santoso", "Siti Rahma"} {
		if !bytes.Contains(b, []byte(nama)) {
			t.Fatalf("PDF gabungan tidak berisi %q", nama)
		}
	}
	if _, err := RenderKartuBulkPDF(nil); err == nil {
		t.Fatalf("bulk kosong harus error")
	}
	if _, err := RenderKartuBulkPDF(make([]kartuPDFData, kartuPDFMaksSiswaBulk+1)); err == nil {
		t.Fatalf("bulk > %d siswa harus error", kartuPDFMaksSiswaBulk)
	}
}

func TestFormatTanggalKartu(t *testing.T) {
	if got := formatTanggalKartu("2026-10-17"); got != "17 Okt 2026" {
		t.Fatalf("formatTanggalKartu(2026-10-17) = %q", got)
	}
	if got := formatTanggalKartu("bukan-tanggal"); got != "bukan-tanggal" {
		t.Fatalf("tanggal tidak valid harus dikembalikan apa adanya, dapat %q", got)
	}
}

func TestNamaFileKartuPDF(t *testing.T) {
	if got := namaFileKartuPDF("0012345678"); got != "kartu-peserta-0012345678.pdf" {
		t.Fatalf("nama file = %q", got)
	}
	if got := namaFileKartuPDF("../../x"); got != "kartu-peserta-x.pdf" {
		t.Fatalf("nama file harus dibersihkan dari path traversal, dapat %q", got)
	}
}

func TestNamaEventGabungan(t *testing.T) {
	if got := namaEventGabungan([]kartuPDFJadwal{{EventNama: "STS"}, {EventNama: "STS"}}); got != "STS" {
		t.Fatalf("satu event sama harus dipakai, dapat %q", got)
	}
	if got := namaEventGabungan([]kartuPDFJadwal{{EventNama: "STS"}, {EventNama: "SAS"}}); got != "Ujian Sekolah" {
		t.Fatalf("event campur harus jadi 'Ujian Sekolah', dapat %q", got)
	}
	if got := namaEventGabungan(nil); got != "Ujian Sekolah" {
		t.Fatalf("tanpa jadwal harus jadi 'Ujian Sekolah', dapat %q", got)
	}
}

func TestKartuSanitize(t *testing.T) {
	if got := kartuSanitize("A – B “C” …"); got != "A - B \"C\" ..." {
		t.Fatalf("kartuSanitize = %q", got)
	}
}

// TestRenderKartuPDFLayoutPoles mengunci hasil poles Tim Format Kartu:
// urutan elemen, 8 jadwal tetap 1 halaman, nomor peserta tidak terpotong,
// dan tidak ada "..." pemotongan pada sampel bernilai pendek.
func TestRenderKartuPDFLayoutPoles(t *testing.T) {
	b, err := RenderKartuPDF(contohKartuPDF(8))
	if err != nil {
		t.Fatalf("RenderKartuPDF: %v", err)
	}
	if !bytes.HasPrefix(b, []byte("%PDF")) {
		t.Fatalf("PDF harus diawali %%PDF")
	}
	for _, teks := range []string{"Budi Santoso", "Mapel 1", "SMK CONTOH BANGSA",
		"Kartu Peserta Ujian", "Jadwal Ujian", "Login Ujian", "Kata sandi awal",
		"Peserta Ujian", "Panitia", "Token ujian diberikan proktor"} {
		if !bytes.Contains(b, []byte(teks)) {
			t.Fatalf("PDF hasil poles tidak berisi teks %q", teks)
		}
	}
	// Urutan elemen: kop -> judul -> identitas -> tabel -> login ->
	// tanda tangan -> footer.
	// ("Kartu Peserta Ujian" mengandung substring "Peserta Ujian", jadi
	// label tanda tangan dicek terpisah mulai dari posisi "Login Ujian".)
	urutan := []string{"SMK CONTOH BANGSA", "Kartu Peserta Ujian", "Nama",
		"Jadwal Ujian", "Login Ujian", "Panitia",
		"Token ujian diberikan proktor"}
	pos := -1
	for _, teks := range urutan {
		p := bytes.Index(b, []byte(teks))
		if p < 0 {
			t.Fatalf("teks %q tidak ditemukan", teks)
		}
		if p <= pos {
			t.Fatalf("teks %q keluar urutan (pos %d, sebelumnya %d)", teks, p, pos)
		}
		pos = p
	}
	// Label "Peserta Ujian" tanda tangan harus muncul SETELAH kotak login
	// (bukan hanya di dalam judul "Kartu Peserta Ujian").
	pLogin := bytes.Index(b, []byte("Login Ujian"))
	if bytes.Index(b[pLogin:], []byte("Peserta Ujian")) < 0 {
		t.Fatalf("label tanda tangan 'Peserta Ujian' harus muncul setelah kotak login")
	}
	// 8 jadwal tetap muat 1 halaman (paginasi dipertahankan).
	if got := hitungHalamanPDF(b); got != 1 {
		t.Fatalf("8 jadwal harus 1 halaman, dapat %d", got)
	}
	// Nomor peserta utuh (tidak terpotong "...") + header tabel 1x.
	if !bytes.Contains(b, []byte("STS-GANJIL-26/2026/001")) {
		t.Fatalf("nomor peserta harus tampil utuh, tanpa pemotongan")
	}
	if got := bytes.Count(b, []byte("Nomor Peserta")); got != 1 {
		t.Fatalf("label 'Nomor Peserta' harus 1x, dapat %d", got)
	}
	if bytes.Contains(b, []byte("...")) {
		t.Fatalf("sampel bernilai pendek tidak boleh mengandung pemotongan '...'")
	}
}

// ---------------------------------------------------------------------------
// Tes endpoint (PostgreSQL nyata, di-skip tanpa TEST_DATABASE_URL).
// ---------------------------------------------------------------------------

const (
	kartuAdminID = "11111111-1111-1111-1111-111111111111"
	kartuSiswa1  = "55555555-5555-5555-5555-555555555551"
	kartuSiswa2  = "55555555-5555-5555-5555-555555555552"
	kartuUserS1  = "55555555-5555-5555-5555-555555555501"
	kartuUserS2  = "55555555-5555-5555-5555-555555555502"
	kartuJurusan = "66666666-6666-6666-6666-666666666661"
	kartuKelas   = "77777777-7777-7777-7777-777777777771"
	kartuMapel   = "88888888-8888-8888-8888-888888888881"
	kartuPaket   = "99999999-9999-9999-9999-999999999991"
	kartuJadwal  = "99999999-9999-9999-9999-999999999992"
	kartuEvent   = "99999999-9999-9999-9999-999999999993"
	kartuRuangan = "99999999-9999-9999-9999-999999999994"
)

func setupKartuPDFDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL tidak diset — skip tes endpoint kartu PDF.")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("koneksi DB: %v", err)
	}
	ddl := []string{
		`DROP TABLE IF EXISTS jadwal_ujian, paket_ujian, siswa, kelas, jurusan, mapel, event, ruangan, users, settings CASCADE`,
		`CREATE TABLE users (id UUID PRIMARY KEY, nama TEXT, username TEXT UNIQUE, password_hash TEXT, role TEXT, is_active BOOLEAN DEFAULT TRUE)`,
		`CREATE TABLE jurusan (id UUID PRIMARY KEY, kode TEXT, nama TEXT)`,
		`CREATE TABLE kelas (id UUID PRIMARY KEY, jurusan_id UUID REFERENCES jurusan(id), tingkat INT, nama TEXT, tahun_ajaran TEXT)`,
		`CREATE TABLE mapel (id UUID PRIMARY KEY, kode TEXT, nama TEXT, jenis TEXT, jurusan_id UUID, kelas_tingkat INT)`,
		`CREATE TABLE event (id UUID PRIMARY KEY, kode TEXT, nama TEXT, jenis TEXT, tanggal_mulai DATE, tanggal_selesai DATE, deskripsi TEXT, created_at TIMESTAMPTZ DEFAULT now())`,
		`CREATE TABLE ruangan (id UUID PRIMARY KEY, kode TEXT, nama TEXT, kapasitas INT, lokasi TEXT, keterangan TEXT, created_at TIMESTAMPTZ DEFAULT now())`,
		`CREATE TABLE paket_ujian (id UUID PRIMARY KEY, kode TEXT, nama TEXT, mapel_id UUID REFERENCES mapel(id), jenis TEXT, durasi_menit INT)`,
		`CREATE TABLE jadwal_ujian (id UUID PRIMARY KEY, paket_id UUID REFERENCES paket_ujian(id), kelas_id UUID REFERENCES kelas(id), tanggal DATE, jam_mulai TIME, jam_selesai TIME, status TEXT, event_id UUID REFERENCES event(id), ruangan_id UUID REFERENCES ruangan(id))`,
		`CREATE TABLE siswa (id UUID PRIMARY KEY, user_id UUID REFERENCES users(id), nisn TEXT UNIQUE, nis TEXT, nama TEXT, kelas_id UUID REFERENCES kelas(id), jurusan_id UUID REFERENCES jurusan(id), is_active BOOLEAN DEFAULT TRUE)`,
		`CREATE TABLE settings (key TEXT PRIMARY KEY, value JSONB)`,
	}
	for _, q := range ddl {
		if _, err := pool.Exec(ctx, q); err != nil {
			t.Fatalf("DDL: %v", err)
		}
	}
	seed := []string{
		fmt.Sprintf(`INSERT INTO users VALUES ('%s','Admin','admin','x','admin',TRUE),('%s','Budi Santoso','0012345678','x','siswa',TRUE),('%s','Siti Rahma','0087654321','x','siswa',TRUE)`, kartuAdminID, kartuUserS1, kartuUserS2),
		fmt.Sprintf(`INSERT INTO jurusan VALUES ('%s','PPLG','Pengembangan Perangkat Lunak dan Gim')`, kartuJurusan),
		fmt.Sprintf(`INSERT INTO kelas VALUES ('%s','%s',11,'XI PPLG 1','2026/2027')`, kartuKelas, kartuJurusan),
		fmt.Sprintf(`INSERT INTO mapel VALUES ('%s','INF','Informatika','umum',NULL,NULL)`, kartuMapel),
		fmt.Sprintf(`INSERT INTO event VALUES ('%s','STS-GANJIL-26','STS Ganjil 2026/2027','STS',NULL,NULL,NULL,now())`, kartuEvent),
		fmt.Sprintf(`INSERT INTO ruangan VALUES ('%s','R1','Ruang 1',40,'Gedung A',NULL,now())`, kartuRuangan),
		fmt.Sprintf(`INSERT INTO paket_ujian VALUES ('%s','STS-INF','STS Informatika','%s','STS',60)`, kartuPaket, kartuMapel),
		fmt.Sprintf(`INSERT INTO jadwal_ujian VALUES ('%s','%s','%s','2026-10-19','07:30','09:30','aktif','%s','%s')`, kartuJadwal, kartuPaket, kartuKelas, kartuEvent, kartuRuangan),
		fmt.Sprintf(`INSERT INTO siswa VALUES ('%s','%s','0012345678',NULL,'Budi Santoso','%s','%s',TRUE)`, kartuSiswa1, kartuUserS1, kartuKelas, kartuJurusan),
		fmt.Sprintf(`INSERT INTO siswa VALUES ('%s','%s','0087654321',NULL,'Siti Rahma','%s','%s',TRUE)`, kartuSiswa2, kartuUserS2, kartuKelas, kartuJurusan),
		`INSERT INTO settings VALUES ('nama_sekolah','"SMK Contoh Bangsa"')`,
	}
	for _, q := range seed {
		if _, err := pool.Exec(ctx, q); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	return pool
}

// reqKartuPDF memanggil endpoint dengan context auth palsu dari header.
func reqKartuPDF(t *testing.T, e *echo.Echo, method, path, userID, role string) (int, http.Header, []byte) {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("X-Test-User", userID)
	req.Header.Set("X-Test-Role", role)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec.Code, rec.Header(), rec.Body.Bytes()
}

func TestKartuPDFEndpoint(t *testing.T) {
	pool := setupKartuPDFDB(t)
	defer pool.Close()

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Set("auth_user_id", c.Request().Header.Get("X-Test-User"))
			c.Set("auth_role", c.Request().Header.Get("X-Test-Role"))
			// Admin lolos semua permission via audit.HasPermission (role admin).
			return next(c)
		}
	})
	Register(e, pool, nil, nil)

	// 1. Siswa unduh kartunya sendiri -> 200 + PDF + nama file per NISN.
	status, header, body := reqKartuPDF(t, e, "GET", "/api/v1/kartu/siswa/"+kartuSiswa1+"/pdf", kartuUserS1, "siswa")
	if status != http.StatusOK {
		t.Fatalf("siswa unduh sendiri: status %d, body %s", status, body)
	}
	if ct := header.Get("Content-Type"); !strings.HasPrefix(ct, "application/pdf") {
		t.Fatalf("content-type harus application/pdf, dapat %q", ct)
	}
	if cd := header.Get("Content-Disposition"); cd != `attachment; filename="kartu-peserta-0012345678.pdf"` {
		t.Fatalf("content-disposition = %q", cd)
	}
	if !bytes.HasPrefix(body, []byte("%PDF")) {
		t.Fatalf("body harus PDF")
	}
	for _, teks := range []string{"Budi Santoso", "Informatika", "STS Ganjil 2026/2027", "Ruang 1", "STS-GANJIL-26/2026/001"} {
		if !bytes.Contains(body, []byte(teks)) {
			t.Fatalf("PDF siswa sendiri tidak berisi %q", teks)
		}
	}
	// Nomor peserta Siti (002) TIDAK boleh ada di kartu Budi.
	if bytes.Contains(body, []byte("STS-GANJIL-26/2026/002")) {
		t.Fatalf("kartu Budi tidak boleh berisi nomor peserta Siti")
	}

	// 1b. Siswa boleh memakai user.id akunnya sebagai siswaId (fallback loadSiswaKartu).
	status, _, body = reqKartuPDF(t, e, "GET", "/api/v1/kartu/siswa/"+kartuUserS1+"/pdf", kartuUserS1, "siswa")
	if status != http.StatusOK || !bytes.Contains(body, []byte("Budi Santoso")) {
		t.Fatalf("siswa via user.id: status %d", status)
	}

	// 2. Siswa lain -> 404 (anti-IDOR), bukan 403.
	status, _, _ = reqKartuPDF(t, e, "GET", "/api/v1/kartu/siswa/"+kartuSiswa2+"/pdf", kartuUserS1, "siswa")
	if status != http.StatusNotFound {
		t.Fatalf("siswa lain harus 404, dapat %d", status)
	}

	// 3. Tanpa login -> 401.
	status, _, _ = reqKartuPDF(t, e, "GET", "/api/v1/kartu/siswa/"+kartuSiswa1+"/pdf", "", "")
	if status != http.StatusUnauthorized {
		t.Fatalf("tanpa login harus 401, dapat %d", status)
	}

	// 4. Staf bulk -> 200 + jumlah halaman = jumlah siswa (2).
	status, header, body = reqKartuPDF(t, e, "GET", "/api/v1/kartu/pdf", kartuAdminID, "admin")
	if status != http.StatusOK {
		t.Fatalf("staf bulk: status %d, body %s", status, body)
	}
	if ct := header.Get("Content-Type"); !strings.HasPrefix(ct, "application/pdf") {
		t.Fatalf("bulk content-type harus application/pdf, dapat %q", ct)
	}
	if got := hitungHalamanPDF(body); got != 2 {
		t.Fatalf("bulk 2 siswa harus 2 halaman, dapat %d", got)
	}
	if !bytes.Contains(body, []byte("Siti Rahma")) || !bytes.Contains(body, []byte("Budi Santoso")) {
		t.Fatalf("bulk harus berisi kedua siswa")
	}

	// 4b. Bulk dengan filter kelas_id -> tetap 2 halaman.
	status, _, body = reqKartuPDF(t, e, "GET", "/api/v1/kartu/pdf?kelas_id="+kartuKelas, kartuAdminID, "admin")
	if status != http.StatusOK || hitungHalamanPDF(body) != 2 {
		t.Fatalf("bulk filter kelas: status %d", status)
	}

	// 4c. Bulk dengan filter event_id -> tetap 2 halaman (kedua siswa punya jadwal di event itu).
	status, _, body = reqKartuPDF(t, e, "GET", "/api/v1/kartu/pdf?event_id="+kartuEvent, kartuAdminID, "admin")
	if status != http.StatusOK || hitungHalamanPDF(body) != 2 {
		t.Fatalf("bulk filter event: status %d", status)
	}

	// 4d. Siswa TIDAK boleh bulk -> 404 (gate staf).
	status, _, _ = reqKartuPDF(t, e, "GET", "/api/v1/kartu/pdf", kartuUserS1, "siswa")
	if status != http.StatusNotFound {
		t.Fatalf("siswa bulk harus 404, dapat %d", status)
	}

	// 5. Endpoint JSON legacy TETAP jalan (tidak rusak oleh route PDF).
	status, _, body = reqKartuPDF(t, e, "GET", "/api/v1/kartu/"+kartuJadwal, kartuAdminID, "admin")
	if status != http.StatusOK || !bytes.Contains(body, []byte("Budi Santoso")) {
		t.Fatalf("JSON legacy daftar: status %d, body %s", status, body)
	}
	status, _, body = reqKartuPDF(t, e, "GET", "/api/v1/kartu/"+kartuJadwal+"/"+kartuSiswa1, kartuAdminID, "admin")
	if status != http.StatusOK || !bytes.Contains(body, []byte("0012345678")) {
		t.Fatalf("JSON legacy satu: status %d, body %s", status, body)
	}
}
