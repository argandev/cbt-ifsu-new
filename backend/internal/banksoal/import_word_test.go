package banksoal

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

// --- Helper pembangun .docx untuk tes ---

func testParaXML(text string) string {
	var buf bytes.Buffer
	_ = xml.EscapeText(&buf, []byte(text))
	return `<w:p><w:r><w:t xml:space="preserve">` + buf.String() + `</w:t></w:r></w:p>`
}

func testTableXML(rows [][]string) string {
	var sb strings.Builder
	sb.WriteString(`<w:tbl>`)
	for _, row := range rows {
		sb.WriteString(`<w:tr>`)
		for _, cell := range row {
			sb.WriteString(`<w:tc>` + testParaXML(cell) + `</w:tc>`)
		}
		sb.WriteString(`</w:tr>`)
	}
	sb.WriteString(`</w:tbl>`)
	return sb.String()
}

func testDocxFromBody(t *testing.T, body string) []byte {
	t.Helper()
	doc := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>` +
		body + `</w:body></w:document>`
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, part := range []struct{ name, content string }{
		{"[Content_Types].xml", docxContentTypesXML},
		{"_rels/.rels", docxRelsXML},
		{"word/styles.xml", docxStylesXML},
		{"word/document.xml", doc},
	} {
		w, err := zw.Create(part.name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(part.content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func testDocxFromLines(t *testing.T, lines []string) []byte {
	t.Helper()
	var body strings.Builder
	for _, ln := range lines {
		body.WriteString(testParaXML(ln))
	}
	return testDocxFromBody(t, body.String())
}

func countFindings(findings []map[string]any, tingkat string) int {
	n := 0
	for _, f := range findings {
		if f["tingkat"] == tingkat {
			n++
		}
	}
	return n
}

func TestParseDocxImportTemplateRoundTrip(t *testing.T) {
	data, err := buildTemplateDocx()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 4 || string(data[:2]) != "PK" {
		t.Fatal("template harus diawali magic bytes ZIP")
	}
	rows, findings := parseDocxImport(data)
	if len(findings) != 0 {
		t.Fatalf("template sendiri harus bersih, dapat temuan: %v", findings)
	}
	if len(rows) != 3 {
		t.Fatalf("harap 3 soal contoh, dapat %d", len(rows))
	}
	if rows[0].tipe != "pg" || rows[1].tipe != "menjodohkan" || rows[2].tipe != "uraian" {
		t.Fatalf("tipe contoh salah: %v %v %v", rows[0].tipe, rows[1].tipe, rows[2].tipe)
	}
	if rows[2].bobot != 2 {
		t.Fatalf("bobot uraian contoh harus 2, dapat %d", rows[2].bobot)
	}
	if rows[0].baris != 1 || rows[1].baris != 2 || rows[2].baris != 3 {
		t.Fatalf("baris = nomor soal, dapat %d %d %d", rows[0].baris, rows[1].baris, rows[2].baris)
	}
}

func TestParseDocxImportBasicPG(t *testing.T) {
	data := testDocxFromLines(t, []string{
		"SOAL 1",
		"TIPE: pg",
		"PERTANYAAN: Ibu kota Jawa Barat adalah",
		"kota yang terkenal dengan kulinernya.",
		"A. Bandung",
		"B. Semarang",
		"KUNCI: A",
		"BOBOT: 2",
		"LEVEL: Sedang",
		"PEMBAHASAN: Bandung adalah ibu kotanya.",
	})
	rows, findings := parseDocxImport(data)
	if countFindings(findings, "error") != 0 {
		t.Fatalf("harap tanpa error, dapat %v", findings)
	}
	if len(rows) != 1 {
		t.Fatalf("harap 1 soal, dapat %d", len(rows))
	}
	if !strings.Contains(rows[0].pertanyaan, "kuliner") {
		t.Fatalf("pertanyaan multi-paragraf harus digabung: %q", rows[0].pertanyaan)
	}
	if rows[0].bobot != 2 || rows[0].level != "sedang" {
		t.Fatalf("bobot/level salah: %+v", rows[0])
	}
}

func TestParseDocxImportSeparatorSplit(t *testing.T) {
	data := testDocxFromLines(t, []string{
		"TIPE: pg", "PERTANYAAN: Q1?", "A. 1", "B. 2", "KUNCI: A",
		"---",
		"TIPE: isian", "PERTANYAAN: Ibu kota Indonesia?", "KUNCI: Jakarta|DKI Jakarta",
	})
	rows, findings := parseDocxImport(data)
	if countFindings(findings, "error") != 0 {
		t.Fatalf("harap tanpa error, dapat %v", findings)
	}
	if len(rows) != 2 || rows[0].baris != 1 || rows[1].baris != 2 {
		t.Fatalf("harap 2 soal bernomor 1,2, dapat %+v", rows)
	}
}

func TestParseDocxImportNoLabelNoSeparator(t *testing.T) {
	// Tanpa label SOAL dan tanpa pemisah: seluruh dokumen = satu blok.
	data := testDocxFromLines(t, []string{
		"TIPE: pg", "PERTANYAAN: Q1?", "A. 1", "B. 2", "KUNCI: A",
	})
	rows, findings := parseDocxImport(data)
	if countFindings(findings, "error") != 0 {
		t.Fatalf("satu soal tanpa header harus tetap ter-parse, dapat %v", findings)
	}
	if len(rows) != 1 || rows[0].baris != 1 {
		t.Fatalf("harap 1 soal bernomor 1, dapat %+v", rows)
	}
	// Dokumen kosong total (tanpa paragraf) → error level file.
	_, findings = parseDocxImport(testDocxFromBody(t, ""))
	if countFindings(findings, "error") == 0 || findings[0]["baris"] != 0 {
		t.Fatalf("dokumen kosong harus error level file (baris 0), dapat %v", findings)
	}
}

func TestParseDocxImportKunciKosongAuto(t *testing.T) {
	data := testDocxFromLines(t, []string{
		"SOAL 1", "TIPE: pg", "PERTANYAAN: Q?", "A. 1", "B. 2",
	})
	rows, findings := parseDocxImport(data)
	if len(rows) != 0 {
		t.Fatalf("soal tanpa kunci (tipe auto) tidak boleh dikembalikan")
	}
	if countFindings(findings, "error") != 1 {
		t.Fatalf("harap 1 error kunci wajib, dapat %v", findings)
	}
	if !strings.Contains(findings[0]["pesan"].(string), "Soal ke-1:") {
		t.Fatalf("pesan harus berawalan 'Soal ke-1:', dapat %v", findings[0])
	}
}

func TestParseDocxImportRubrikWarning(t *testing.T) {
	data := testDocxFromLines(t, []string{
		"SOAL 1", "TIPE: uraian", "PERTANYAAN: Jelaskan X.",
	})
	rows, findings := parseDocxImport(data)
	if len(rows) != 1 {
		t.Fatalf("uraian tanpa rubrik tetap disimpan, dapat %d soal", len(rows))
	}
	if countFindings(findings, "warning") != 1 || countFindings(findings, "error") != 0 {
		t.Fatalf("harap 1 warning rubrik tanpa error, dapat %v", findings)
	}
}

func TestParseDocxImportSikapInfo(t *testing.T) {
	data := testDocxFromLines(t, []string{
		"SOAL 1", "TIPE: pernyataan_sikap",
		"PERTANYAAN: Beri tanggapanmu.",
		"A. Saya senang diskusi.", "B. Saya suka praktik.",
		"KUNCI: abaikan",
	})
	rows, findings := parseDocxImport(data)
	if len(rows) != 1 {
		t.Fatalf("pernyataan_sikap harus disimpan, dapat %d", len(rows))
	}
	if countFindings(findings, "info") != 1 || countFindings(findings, "error") != 0 {
		t.Fatalf("harap 1 info kunci-diabaikan, dapat %v", findings)
	}
}

func TestParseDocxImportMediaDanAltMedia(t *testing.T) {
	data := testDocxFromLines(t, []string{
		"SOAL 1", "TIPE: pg", "PERTANYAAN: Manakah gambar UTP?",
		"A. https://contoh.sekolah.id/a.png", "B. https://contoh.sekolah.id/b.png",
		"KUNCI: A",
		"MEDIA: https://contoh.sekolah.id/gambar/utp.png",
		"ALT MEDIA: Kabel UTP",
		"SOAL 2", "TIPE: pg", "PERTANYAAN: Q2?",
		"A. 1", "B. 2", "KUNCI: B",
		"MEDIA ALT: Gambar soal kedua",
	})
	rows, findings := parseDocxImport(data)
	if len(rows) != 2 {
		t.Fatalf("harap 2 soal, dapat %d (%v)", len(rows), findings)
	}
	// MEDIA / ALT MEDIA kini ikut tersimpan (csvSoal membawa fieldnya) —
	// tidak ada lagi warning "BELUM disimpan" dari parser.
	if countFindings(findings, "warning") != 0 || countFindings(findings, "error") != 0 {
		t.Fatalf("harap 0 warning/error dari parser, dapat %v", findings)
	}
	if rows[0].mediaURL != "https://contoh.sekolah.id/gambar/utp.png" || rows[0].mediaAlt != "Kabel UTP" {
		t.Fatalf("soal 1 harus membawa media_url + media_alt, dapat %+v", rows[0])
	}
	if rows[1].mediaAlt != "Gambar soal kedua" {
		t.Fatalf("soal 2 harus membawa media_alt (alias MEDIA ALT:), dapat %+v", rows[1])
	}
}

func TestParseDocxImportKisiKisiE30(t *testing.T) {
	data := testDocxFromLines(t, []string{
		"SOAL 1", "TIPE: pg", "PERTANYAAN: Q?",
		"A. 1", "B. 2", "KUNCI: A",
		"KISI-KISI: K1",
	})
	rows, findings := parseDocxImport(data)
	if len(rows) != 0 || countFindings(findings, "error") != 1 {
		t.Fatalf("KISI-KISI terisi harus E30 (blok gagal), dapat %d soal %v", len(rows), findings)
	}
	kode := ""
	for _, f := range findings {
		if s, ok := f["kode"].(string); ok && s != "" {
			kode = s
		}
	}
	if kode != "E30" {
		t.Fatalf("harap kode E30, dapat %q (%v)", kode, findings)
	}
}

func TestParseDocxImportKompetensiDibawa(t *testing.T) {
	data := testDocxFromLines(t, []string{
		"SOAL 1", "TIPE: pg", "PERTANYAAN: Q?",
		"A. 1", "B. 2", "KUNCI: A",
		"KOMPETENSI: KD-1",
	})
	rows, findings := parseDocxImport(data)
	if len(rows) != 1 || countFindings(findings, "error") != 0 {
		t.Fatalf("harap 1 soal tanpa error (lookup E31 terjadi di gabungkanDocx), dapat %d %v", len(rows), findings)
	}
	if rows[0].kompetensi != "KD-1" {
		t.Fatalf("soal harus membawa kode kompetensi mentah, dapat %+v", rows[0])
	}
}

func TestParseDocxImportMediaTidakValid(t *testing.T) {
	data := testDocxFromLines(t, []string{
		"SOAL 1", "TIPE: pg", "PERTANYAAN: Q?",
		"A. 1", "B. 2", "KUNCI: A",
		"MEDIA: bukan-url",
	})
	rows, findings := parseDocxImport(data)
	if len(rows) != 0 || countFindings(findings, "error") != 1 {
		t.Fatalf("MEDIA bukan URL harus error, dapat %d soal %v", len(rows), findings)
	}
}

func TestParseDocxImportPilihanGambarOpsiBukanURL(t *testing.T) {
	data := testDocxFromLines(t, []string{
		"SOAL 1", "TIPE: pilihan_gambar", "PERTANYAAN: Manakah kabel UTP?",
		"A. teks bukan url", "B. https://contoh.sekolah.id/b.png",
		"KUNCI: B",
	})
	rows, findings := parseDocxImport(data)
	if len(rows) != 0 || countFindings(findings, "error") != 1 {
		t.Fatalf("opsi pilihan_gambar bukan URL harus error, dapat %d soal %v", len(rows), findings)
	}
}

func TestParseDocxImportTabel(t *testing.T) {
	body := testParaXML("SOAL 1") + testTableXML([][]string{
		{"TIPE:", "pg"},
		{"PERTANYAAN:", "Berapa 1+1?"},
		{"A.", "1"},
		{"B.", "2"},
		{"KUNCI:", "B"},
	})
	data := testDocxFromBody(t, body)
	rows, findings := parseDocxImport(data)
	if countFindings(findings, "error") != 0 {
		t.Fatalf("tabel harus terbaca sebagai baris teks, dapat %v", findings)
	}
	if len(rows) != 1 || rows[0].tipe != "pg" {
		t.Fatalf("harap 1 soal pg dari tabel, dapat %+v", rows)
	}
}

func TestParseDocxImportStimulusTidakDimakanOpsi(t *testing.T) {
	data := testDocxFromLines(t, []string{
		"SOAL 1", "TIPE: pg",
		"STIMULUS: Bacalah teks berikut.",
		"A. Langkah pertama adalah membaca.",
		"PERTANYAAN: Apa langkah pertama?",
		"A. Membaca", "B. Menulis",
		"KUNCI: A",
	})
	rows, findings := parseDocxImport(data)
	if countFindings(findings, "error") != 0 {
		t.Fatalf("harap tanpa error, dapat %v", findings)
	}
	if len(rows) != 1 || !strings.Contains(rows[0].stimulus, "Langkah pertama") {
		t.Fatalf("paragraf stimulus 'A. ...' harus tetap jadi stimulus: %+v", rows)
	}
}

func TestParseDocxImportValidasiFile(t *testing.T) {
	if _, findings := parseDocxImport([]byte("bukan zip")); countFindings(findings, "error") == 0 {
		t.Fatal("bukan ZIP harus ditolak magic bytes")
	}
	besar := make([]byte, maxDocxImportBytes+1)
	copy(besar, []byte{'P', 'K', 0x03, 0x04})
	if _, findings := parseDocxImport(besar); countFindings(findings, "error") == 0 {
		t.Fatal("file > 5 MB harus ditolak")
	}
	// ZIP valid tetapi tanpa word/document.xml.
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create("word/lain.xml")
	_, _ = w.Write([]byte("<x/>"))
	_ = zw.Close()
	if _, findings := parseDocxImport(buf.Bytes()); countFindings(findings, "error") == 0 {
		t.Fatal("docx tanpa document.xml harus ditolak")
	}
}

func TestRegisterImportWordTemplateTolakTanpaLogin(t *testing.T) {
	e := echo.New()
	RegisterImportWord(e, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/bank-soal/import/template.docx", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("tanpa login harus 401, dapat %d", rec.Code)
	}
}
