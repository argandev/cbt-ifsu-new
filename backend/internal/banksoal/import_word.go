// Package banksoal — import soal format Word (.docx).
//
// FORMAT WORD KITA (desain sendiri — BUKAN format Exo S:/A:/JAWAB: /
// TS/KD/KJ/ABS, dan bukan dialek lain). Satu file .docx berisi satu
// atau lebih blok soal. Tiap blok diawali baris:
//
//	SOAL 1
//
// lalu label-label berikut, masing-masing satu baris, diakhiri titik
// dua. Urutan label bebas, tetapi urutan yang dipakai template adalah
// urutan yang disarankan (opsi ditulis SETELAH pertanyaan dan SEBELUM
// label KUNCI):
//
//	SOAL 1
//	TIPE: pg
//	STIMULUS: <teks bacaan/kasus; wajib untuk studi_kasus>
//	PERTANYAAN: <teks soal; wajib>
//	A. <opsi A>
//	B. <opsi B>
//	C. <opsi C>
//	D. <opsi D>
//	E. <opsi E>
//	KUNCI: <kunci, format per tipe — lihat di bawah>
//	BOBOT: 1
//	LEVEL: sedang
//	PEMBAHASAN: <penjelasan jawaban>
//	KISI-KISI: <kode kisi-kisi bank ini>
//	KOMPETENSI: <kode kompetensi bank ini>
//	MEDIA: <satu URL gambar https://...>
//	ALT MEDIA: <teks alternatif gambar di MEDIA>
//
// Contoh blok lengkap (dari template yang dihasilkan
// buildTemplateDocx):
//
//	SOAL 1
//	TIPE: pg
//	STIMULUS:
//	PERTANYAAN: Ibu kota provinsi Jawa Barat adalah …
//	A. Bandung
//	B. Semarang
//	C. Medan
//	D. Surabaya
//	KUNCI: A
//	BOBOT: 1
//	LEVEL: sedang
//	PEMBAHASAN: Jawa Barat beribu kota di Bandung.
//	KISI-KISI:
//	KOMPETENSI:
//	MEDIA:
//	ALT MEDIA:
//
//	SOAL 2
//	TIPE: menjodohkan
//	PERTANYAAN: Pasangkan kota dengan provinsinya.
//	A. Bandung::Jawa Barat
//	B. Semarang::Jawa Tengah
//	C. Surabaya::Jawa Timur
//	D. Medan::Sumatera Utara
//	KUNCI: Bandung::Jawa Barat;Semarang::Jawa Tengah;Surabaya::Jawa Timur;Medan::Sumatera Utara
//	BOBOT: 1
//	LEVEL: sedang
//	PEMBAHASAN: Pasangan kota–provinsi yang benar sesuai peta Indonesia.
//
//	SOAL 3
//	TIPE: uraian
//	PERTANYAAN: Jelaskan perbedaan antara RAM dan ROM beserta contoh
//	penggunaannya dalam kehidupan sehari-hari.
//	KUNCI: Rubrik: pengertian RAM (30), pengertian ROM (30),
//	perbedaan (20), contoh (20).
//	BOBOT: 2
//	LEVEL: sulit
//
// Aturan parsing (kontrak ketat — dokumen yang tidak mengikuti
// kontrak ditolak dengan laporan kesalahan, bukan ditebak mesin):
//
//   - Label dikenali case-insensitive dan wajib diakhiri titik dua
//     ("TIPE: pg"). Alias yang diterima: "KISI KISI:" / "KISI_KISI:"
//     untuk KISI-KISI, "ALT_MEDIA:" / "MEDIA ALT:" untuk ALT MEDIA.
//   - TIPE wajib kode (mis. pg), bukan label UI ("Pilihan Ganda"
//     ditolak). Daftar 16 kode sama seperti import Excel/CSV.
//   - Baris biasa yang bukan label dan bukan opsi adalah LANJUTAN
//     field terakhir (pertanyaan / stimulus / pembahasan multi-
//     paragraf didukung; baris liar sebelum label apa pun dianggap
//     awal PERTANYAAN).
//   - Opsi dikenali dari pola "A. ...", "A) ...", "A: ...", "A- ..."
//     (huruf A–E, besar/kecil) dan HANYA bila field aktif adalah
//     PERTANYAAN / TIPE / belum ada field, dan label KUNCI belum
//     ditemui. Konsekuensi yang didokumentasikan: (a) paragraf
//     stimulus yang diawali "A. ..." TIDAK dimakan sebagai opsi
//     karena stimulus ditulis sebelum pertanyaan; (b) teks pertanyaan
//     multi-paragraf harus ditulis SEBELUM baris opsi pertama —
//     baris biasa sesudah opsi dimulai ditempel ke opsi terakhir
//     (opsi multi-paragraf); (c) opsi harus ditulis sebelum label
//     KUNCI, sesuai urutan template.
//   - Label bernilai tunggal (TIPE, KUNCI, BOBOT, LEVEL, MEDIA,
//     ALT MEDIA, KISI-KISI, KOMPETENSI) yang ditulis dua kali dalam
//     satu blok → error. Label multi-baris (STIMULUS, PERTANYAAN,
//     PEMBAHASAN) yang diulang → nilainya digabung dengan spasi.
//   - Label yang tidak dikenal ("CATATAN: ...") diperlakukan sebagai
//     baris lanjutan, bukan error.
//   - Tabel Word (.docx) dibaca sebagai baris teks juga: tiap baris
//     tabel menjadi satu baris "sel1 sel2 ...", sehingga guru yang
//     menulis "TIPE:" di kolom kiri dan "pg" di kolom kanan tetap
//     ter-parse. Teks yang dihapus (track changes / w:del) dan
//     instruksi field (w:instrText) diabaikan.
//   - Tanpa label "SOAL <n>", dokumen dipecah dengan garis pemisah
//     ("---", "===", "***", ... — satu baris berisi >= 3 karakter
//     pemisah yang sama) dan blok dinomori berurutan 1..N. Tanpa
//     label SOAL DAN tanpa garis pemisah, seluruh dokumen dianggap
//     satu blok (Soal ke-1) — dokumen satu soal tanpa header tetap
//     bisa diimpor; dokumen multi-soal tanpa header/pemisah akan
//     gagal dengan error label ganda, bukan diam-diam tergabung.
//   - Teks sebelum label "SOAL 1" pertama dianggap petunjuk dan
//     diabaikan (template memakai ini untuk menaruh cara pengisian
//     di dalam dokumen).
//
// Validasi per soal — sama seperti import CSV/Excel (reuse
// buildCSVSoal + ValidateSoal, jadi semantik kunci per tipe, pesan
// error, dan aturan penilaiannya identik):
//
//   - tipe dikenal; pertanyaan tidak kosong.
//   - kunci WAJIB untuk 12 tipe yang dinilai otomatis (error bila
//     kosong: "kunci wajib diisi"); kunci boleh kosong untuk tipe
//     rubrik (uraian, upload_file, praktik_singkat) → WARNING
//     "rubrik belum diisi"; kunci DIABAIKAN untuk pernyataan_sikap
//     (info bila terisi).
//   - bobot bilangan bulat > 0 (default 1); level mudah/sedang/sulit
//     (default sedang); stimulus wajib untuk studi_kasus; jumlah
//     rumpang ___ harus sama dengan jumlah jawaban kunci untuk
//     teks_rumpang.
//   - pilihan_gambar: tiap opsi yang terisi harus URL http(s) yang
//     valid, bila tidak → error "opsi <LABEL> harus URL gambar
//     http(s)".
//   - MEDIA: bila terisi harus URL http(s) yang valid, bila tidak →
//     error `media_url "<nilai>" bukan URL http(s) yang valid`.
//   - ALT MEDIA: teks bebas (teks alternatif gambar di MEDIA,
//     sejajar media_alt di Excel); tidak ada validasi bentuk.
//
// Bentuk kembalian parseDocxImport: ([]csvSoal, []map[string]any).
// Karena signature ini hanya punya satu slot untuk temuan, error /
// warning / info digabung dalam slice kedua dengan pembeda field
// "tingkat" ("error" | "warning" | "info") — pemanggil WAJIB
// memfilter tingkat=="error" untuk menghitung error (entri CSV lama
// tidak punya field "tingkat"; anggap tanpa tingkat = error).
// Setiap temuan memakai "baris" = NOMOR SOAL (label "SOAL <n>" atau
// nomor urut blok bila tanpa label), BUKAN nomor baris dokumen —
// dan pesannya diawali "Soal ke-N: ...". Entri level file memakai
// baris 0. Invarian: bila sebuah blok punya temuan tingkat error,
// soal blok itu TIDAK dikembalikan (pemanggil memutuskan
// all_or_nothing vs skip_errors seperti pada CSV).
//
// Duplikat: TIDAK dideteksi di sini — pemanggil (endpoint
// preview/simpan Tim Excel) yang mendeteksi, karena duplikat-dengan-
// bank butuh konteks bank + database, dan satu definisi duplikat
// (normalisasi trim + lowercase + rapatkan spasi, §2.7 spesifikasi)
// harus berlaku seragam untuk Excel dan Word.
//
// KETERBATASAN YANG JUJUR (baca sebelum mengklaim "sudah didukung"):
//   - csvSoal (didefinisikan di banksoal.go) TIDAK punya field
//     media_url / media_alt / kisi_kisi / kompetensi, sehingga nilai
//     MEDIA / ALT MEDIA / KISI-KISI / KOMPETENSI yang ter-parse
//     BELUM BISA disimpan lewat tipe ini. Parser memvalidasi MEDIA
//     (URL) dan mengeluarkan WARNING eksplisit per soal bila salah
//     satu dari keempat label itu terisi — tidak pernah diam-diam
//     dibuang. Agar keempat nilai itu ikut tersimpan, koordinator
//     perlu memperluas struct csvSoal (atau menambah struct hasil
//     parse baru) + mengisi kolomnya di endpoint simpan; pemetaan
//     KISI-KISI / KOMPETENSI dari kode ke ID juga harus dilakukan
//     di endpoint simpan (parser ini tidak punya konteks bank/DB,
//     jadi tidak bisa melakukan lookup E30/E31 spesifikasi).
//   - Gambar INLINE yang ditempel di dalam .docx TIDAK ikut
//     terimpor (tidak ada ekstraksi word/media/ + pemetaan ke soal).
//     Alasan: gambar inline butuh jalur storage + pemetaan
//     gambar-ke-soal yang belum ada di import manapun; model yang
//     didukung adalah satu URL gambar per soal lewat label MEDIA:
//     (sama seperti kolom media_url di Excel). Guru yang menempel
//     gambar langsung akan melihat soal masuk tanpa gambar —
//     template menegaskan hal ini di petunjuknya.
//   - Format teks kaya (bold/italic/warna), rumus Word (OMML),
//     catatan kaki, dan kotak teks TIDAK dipertahankan — teks
//     dibaca sebagai teks polos (paragraf di dalam kotak teks ikut
//     terbaca sebagai paragraf biasa). Rumus tetap lewat editor
//     soal, bukan lewat import.
//   - Cek ekstensi ".docx" TIDAK dilakukan di parseDocxImport
//     (ia hanya menerima []byte) — endpoint upload Tim Excel WAJIB
//     memeriksa ekstensi nama file sebelum memanggil parser ini.
//     Yang diperiksa di sini: ukuran maks 5 MB + magic bytes ZIP
//     "PK\x03\x04" + struktur word/document.xml.
package banksoal

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

const (
	// maxDocxImportBytes: batas ukuran file .docx (sama seperti Excel: 5 MB).
	maxDocxImportBytes = 5 * 1024 * 1024
	// maxDocxSoal: batas jumlah soal per file (sama seperti CSV: 500).
	maxDocxSoal = 500
	// maxDocxXMLBytes: batas ukuran word/document.xml setelah diekstrak
	// (pertahanan terhadap zip bomb — file .docx sendiri sudah dibatasi
	// 5 MB, tetapi hasil dekompresinya bisa jauh lebih besar).
	maxDocxXMLBytes = 10 * 1024 * 1024

	docxContentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
)

// RegisterImportWord mendaftarkan endpoint template Word.
//
//	GET /api/v1/bank-soal/import/template.docx  (banksoal.import)
//
// Mengunduh dokumen Word FORMAT KITA (lihat komentar paket di atas):
// petunjuk singkat + 3 contoh soal lengkap. Template ini sendiri
// lolos parseDocxImport (3 soal, tanpa error/warning) — diuji di
// import_word_test.go. Pool tidak dipakai (template tidak menyentuh
// database) tetapi dipertahankan di signature agar sejajar dengan
// Register dan siap bila template suatu hari butuh data DB.
func RegisterImportWord(e *echo.Echo, pool *pgxpool.Pool) {
	_ = pool
	e.GET("/api/v1/bank-soal/import/template.docx", func(c *echo.Context) error {
		_, role, ok := requirePerm(c, "banksoal.import")
		if !ok {
			return nil
		}
		if !forbidSoalUntukProktor(c, role) {
			return nil
		}
		data, err := buildTemplateDocx()
		if err != nil {
			return platform.WriteError(c, http.StatusInternalServerError, "internal_error", "Gagal membuat template Word.")
		}
		c.Response().Header().Set("Content-Disposition", `attachment; filename="template-import-soal-cbt-smk-v2.docx"`)
		return c.Blob(http.StatusOK, docxContentType, data)
	})
}

// --- Temuan (error / warning / info) ---

func docxFinding(baris int, tingkat, format string, args ...any) map[string]any {
	return map[string]any{
		"baris":   baris,
		"pesan":   fmt.Sprintf(format, args...),
		"tingkat": tingkat,
	}
}

// docxErr membuat temuan error per soal ("Soal ke-N: ...").
func docxErr(nomor int, format string, args ...any) map[string]any {
	return docxFinding(nomor, "error", "Soal ke-%d: %s", nomor, fmt.Sprintf(format, args...))
}

// docxErrKode membuat temuan error dengan kode katalog (mis. E30/E31)
// agar gabungkanDocx bisa meneruskan kodenya ke preview (paritas .xlsx).
func docxErrKode(nomor int, kode, format string, args ...any) map[string]any {
	m := docxErr(nomor, format, args...)
	m["kode"] = kode
	return m
}

func docxWarn(nomor int, format string, args ...any) map[string]any {
	return docxFinding(nomor, "warning", "Soal ke-%d: %s", nomor, fmt.Sprintf(format, args...))
}

func docxInfo(nomor int, format string, args ...any) map[string]any {
	return docxFinding(nomor, "info", "Soal ke-%d: %s", nomor, fmt.Sprintf(format, args...))
}

// --- parseDocxImport ---

// parseDocxImport mem-parsing isi file .docx menjadi soal yang sudah
// divalidasi. Bentuk kembalian sama dengan parseCSVImport, kecuali:
// (1) "baris" pada temuan = NOMOR SOAL, bukan nomor baris file;
// (2) slice temuan memuat error + warning + info yang dibedakan
// field "tingkat" (lihat komentar paket); (3) duplikat tidak
// dideteksi di sini (tugas pemanggil/endpoint).
func parseDocxImport(data []byte) ([]csvSoal, []map[string]any) {
	if len(data) == 0 {
		return nil, []map[string]any{
			docxFinding(0, "error", "File .docx kosong."),
		}
	}
	if len(data) > maxDocxImportBytes {
		return nil, []map[string]any{
			docxFinding(0, "error", "File melebihi batas 5 MB."),
		}
	}
	if len(data) < 4 || !bytes.Equal(data[:4], []byte{'P', 'K', 0x03, 0x04}) {
		return nil, []map[string]any{
			docxFinding(0, "error", "File bukan .docx yang valid (harus diawali magic bytes ZIP PK\\x03\\x04)."),
		}
	}
	lines, err := extractDocxLines(data)
	if err != nil {
		return nil, []map[string]any{
			docxFinding(0, "error", "File .docx tidak bisa dibaca: %s", err.Error()),
		}
	}
	blocks, splitFindings := splitDocxBlocks(lines)
	if len(blocks) == 0 {
		// splitFindings berisi error level file (tidak ada blok soal).
		return nil, splitFindings
	}
	if len(blocks) > maxDocxSoal {
		return nil, []map[string]any{
			docxFinding(0, "error", "File melebihi batas 500 baris soal."),
		}
	}
	var out []csvSoal
	findings := append([]map[string]any(nil), splitFindings...)
	for _, b := range blocks {
		soal, blockFindings := parseDocxBlock(b)
		findings = append(findings, blockFindings...)
		if soal != nil {
			out = append(out, *soal)
		}
	}
	return out, findings
}

// --- Ekstraksi teks dari word/document.xml ---

// extractDocxLines membaca word/document.xml dari arsip .docx dan
// mengembalikan satu string per paragraf. Baris tabel menjadi satu
// baris "sel1 sel2 ...". Teks terhapus (w:del/w:delText) dan instruksi
// field (w:instrText) diabaikan; tab / jeda baris menjadi spasi.
func extractDocxLines(data []byte) ([]string, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	var docXML []byte
	for _, f := range zr.File {
		if f.Name != "word/document.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		b, err := io.ReadAll(io.LimitReader(rc, maxDocxXMLBytes+1))
		_ = rc.Close()
		if err != nil {
			return nil, err
		}
		if len(b) > maxDocxXMLBytes {
			return nil, fmt.Errorf("word/document.xml melebihi batas %d byte", maxDocxXMLBytes)
		}
		docXML = b
		break
	}
	if docXML == nil {
		return nil, fmt.Errorf("word/document.xml tidak ditemukan di dalam .docx")
	}

	dec := xml.NewDecoder(bytes.NewReader(docXML))
	var lines []string
	var para strings.Builder
	inPara := false
	inTable := 0
	var rowCells []string
	var cell strings.Builder
	inCell := false
	skipDepth := 0
	skipTags := map[string]bool{"instrText": true, "delText": true, "del": true}

	flushPara := func() {
		if !inPara {
			return
		}
		t := strings.TrimSpace(para.String())
		para.Reset()
		inPara = false
		if t == "" {
			return
		}
		if inCell {
			cell.WriteString(t)
			cell.WriteString(" ")
			return
		}
		lines = append(lines, t)
	}

	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			name := t.Name.Local
			if skipTags[name] {
				skipDepth++
			}
			if skipDepth > 0 {
				continue
			}
			switch name {
			case "tbl":
				inTable++
			case "tr":
				if inTable > 0 {
					rowCells = nil
				}
			case "tc":
				if inTable > 0 {
					inCell = true
					cell.Reset()
				}
			case "p":
				inPara = true
				para.Reset()
			case "tab", "br", "cr":
				if inPara {
					para.WriteString(" ")
				}
			}
		case xml.EndElement:
			name := t.Name.Local
			if skipTags[name] {
				if skipDepth > 0 {
					skipDepth--
				}
				continue
			}
			if skipDepth > 0 {
				continue
			}
			switch name {
			case "p":
				flushPara()
			case "tc":
				if inCell {
					rowCells = append(rowCells, strings.TrimSpace(cell.String()))
					cell.Reset()
					inCell = false
				}
			case "tr":
				if inTable > 0 {
					var parts []string
					for _, c := range rowCells {
						if c != "" {
							parts = append(parts, c)
						}
					}
					if len(parts) > 0 {
						lines = append(lines, strings.Join(parts, " "))
					}
					rowCells = nil
				}
			case "tbl":
				if inTable > 0 {
					inTable--
				}
			}
		case xml.CharData:
			if skipDepth == 0 && inPara {
				para.WriteString(string(t))
			}
		}
	}
	flushPara()
	return lines, nil
}

// --- Pemecahan dokumen menjadi blok soal ---

type docxBlock struct {
	nomor int // nomor SOAL (atau nomor urut blok bila tanpa label)
	lines []string
}

var (
	// "SOAL 1", "SOAL 1:", "Soal 12. Judul ..." — grup 1 = nomor,
	// grup 2 = sisa teks di baris yang sama (diperlakukan sebagai
	// baris isi pertama blok).
	docxSoalRe = regexp.MustCompile(`(?i)^SOAL\s+(\d+)\s*[:.]?\s*(.*)$`)
	// Garis pemisah: satu baris berisi >= 3 karakter pemisah yang sama.
	docxSepRe = regexp.MustCompile(`^[-=_*#~]{3,}$`)
)

// splitDocxBlocks memecah baris dokumen menjadi blok per soal.
// Mengembalikan blok + temuan (error nomor SOAL duplikat / tidak
// valid, atau error level file bila tidak ada blok sama sekali).
func splitDocxBlocks(lines []string) ([]docxBlock, []map[string]any) {
	type hit struct {
		idx   int
		nomor int
		rest  string
	}
	var hits []hit
	for i, ln := range lines {
		if m := docxSoalRe.FindStringSubmatch(ln); m != nil {
			n, _ := strconv.Atoi(m[1])
			hits = append(hits, hit{idx: i, nomor: n, rest: strings.TrimSpace(m[2])})
		}
	}
	if len(hits) > 0 {
		var blocks []docxBlock
		var findings []map[string]any
		seen := map[int]bool{}
		for h, cur := range hits {
			end := len(lines)
			if h+1 < len(hits) {
				end = hits[h+1].idx
			}
			if cur.nomor < 1 {
				findings = append(findings, docxFinding(0, "error",
					"Label SOAL %d tidak valid — nomor soal mulai dari 1.", cur.nomor))
				continue
			}
			if seen[cur.nomor] {
				findings = append(findings, docxErr(cur.nomor,
					"nomor SOAL duplikat — blok kedua dan seterusnya diabaikan."))
				continue
			}
			seen[cur.nomor] = true
			var bl []string
			if cur.rest != "" {
				bl = append(bl, cur.rest)
			}
			for _, ln := range lines[cur.idx+1 : end] {
				if docxSepRe.MatchString(ln) {
					continue // garis pemisah visual di dalam blok diabaikan
				}
				bl = append(bl, ln)
			}
			blocks = append(blocks, docxBlock{nomor: cur.nomor, lines: bl})
		}
		return blocks, findings
	}

	// Tanpa label SOAL: pecah dengan garis pemisah.
	var blocks []docxBlock
	var cur []string
	flush := func() {
		nonEmpty := false
		for _, ln := range cur {
			if strings.TrimSpace(ln) != "" {
				nonEmpty = true
				break
			}
		}
		if nonEmpty {
			blocks = append(blocks, docxBlock{
				nomor: len(blocks) + 1,
				lines: append([]string(nil), cur...),
			})
		}
		cur = nil
	}
	for _, ln := range lines {
		if docxSepRe.MatchString(ln) {
			flush()
			continue
		}
		cur = append(cur, ln)
	}
	flush()
	if len(blocks) == 0 {
		return nil, []map[string]any{docxFinding(0, "error",
			`Tidak ditemukan label SOAL 1, SOAL 2, ... — tiap soal harus diawali baris "SOAL <nomor>" (atau pisahkan antar-soal dengan garis --- bila tanpa label).`)}
	}
	return blocks, nil
}

// --- Parsing satu blok soal ---

var (
	// Label: kata (huruf/spasi/underscore/strip) + titik dua + nilai.
	// Contoh: "TIPE: pg", "KISI-KISI: K1", "ALT MEDIA: teks".
	docxLabelRe = regexp.MustCompile(`^([A-Za-z][A-Za-z _\-]*)\s*:\s*(.*)$`)
	// Opsi: "A. ...", "A) ...", "A: ...", "A- ..." (A–E, besar/kecil).
	docxOpsiRe = regexp.MustCompile(`^([A-Ea-e])\s*[.)\-:]\s*(.*)$`)
)

// normDocxLabel menormalkan label ("KISI-KISI" -> "KISIKISI",
// "ALT MEDIA" -> "ALTMEDIA") agar alias ejaan tidak perlu dicabang.
func normDocxLabel(s string) string {
	s = strings.ToUpper(s)
	s = strings.NewReplacer(" ", "", "_", "", "-", "").Replace(s)
	return s
}

// docxFieldForLabel memetakan label ternormalisasi ke nama field.
// "" = label tidak dikenal (diperlakukan sebagai baris lanjutan).
func docxFieldForLabel(norm string) string {
	switch norm {
	case "TIPE":
		return "tipe"
	case "STIMULUS":
		return "stimulus"
	case "PERTANYAAN":
		return "pertanyaan"
	case "KUNCI":
		return "kunci"
	case "BOBOT":
		return "bobot"
	case "LEVEL":
		return "level"
	case "PEMBAHASAN":
		return "pembahasan"
	case "KISIKISI":
		return "kisi_kisi"
	case "KOMPETENSI":
		return "kompetensi"
	case "MEDIA":
		return "media"
	case "ALTMEDIA", "MEDIAALT":
		// "ALT MEDIA:" / "ALT_MEDIA:" / "MEDIA ALT:" — teks alternatif
		// gambar di MEDIA (sejajar media_alt di Excel).
		return "alt_media"
	}
	return ""
}

// isHTTPURL melaporkan apakah s adalah URL http(s) absolut dengan host.
func isHTTPURL(s string) bool {
	u, err := url.Parse(strings.TrimSpace(s))
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return u.Host != ""
}

// parseDocxBlock mem-parsing satu blok soal menjadi csvSoal.
// Mengembalikan (nil, temuan) bila blok punya temuan tingkat error
// (invarian: soal dengan error tidak pernah dikembalikan).
func parseDocxBlock(b docxBlock) (*csvSoal, []map[string]any) {
	nomor := b.nomor
	var findings []map[string]any
	if len(b.lines) == 0 {
		return nil, []map[string]any{docxErr(nomor, "blok soal kosong.")}
	}

	fields := map[string]string{}
	seenSingle := map[string]bool{}
	// Field bernilai tunggal: ditulis dua kali dalam satu blok = error.
	singleVal := map[string]bool{
		"tipe": true, "kunci": true, "bobot": true, "level": true,
		"media": true, "alt_media": true, "kisi_kisi": true, "kompetensi": true,
	}
	var opsiText [5]string
	var opsiSeen [5]bool
	cur := "" // field aktif: "" = belum ada label
	kunciSeen := false
	opsiMode := false // true setelah baris opsi pertama, sampai label baru
	lastOpsi := -1

	appendVal := func(field, val string) {
		val = strings.TrimSpace(val)
		if val == "" {
			return
		}
		if existing := fields[field]; existing == "" {
			fields[field] = val
		} else {
			fields[field] = existing + " " + val
		}
	}

	for _, rawLn := range b.lines {
		ln := strings.TrimSpace(rawLn)
		if ln == "" {
			continue
		}
		// 1. Label dikenal?
		if m := docxLabelRe.FindStringSubmatch(ln); m != nil {
			if field := docxFieldForLabel(normDocxLabel(m[1])); field != "" {
				if singleVal[field] && seenSingle[field] {
					findings = append(findings, docxErr(nomor,
						"label %s ditulis dua kali.", strings.ToUpper(strings.TrimSpace(m[1]))))
					continue
				}
				seenSingle[field] = true
				cur = field
				if field == "kunci" {
					kunciSeen = true
				}
				opsiMode = false
				lastOpsi = -1
				appendVal(field, m[2])
				continue
			}
			// Label tak dikenal → jatuh ke bawah sebagai baris lanjutan.
		}
		// 2. Baris opsi? Hanya bila field aktif PERTANYAAN / TIPE /
		// belum ada field, dan label KUNCI belum ditemui.
		if m := docxOpsiRe.FindStringSubmatch(ln); m != nil && !kunciSeen &&
			(cur == "pertanyaan" || cur == "tipe" || cur == "") {
			idx := int(strings.ToUpper(m[1])[0] - 'A')
			if opsiSeen[idx] {
				findings = append(findings, docxErr(nomor,
					"opsi %s ditulis dua kali.", strings.ToUpper(m[1])))
				continue
			}
			opsiSeen[idx] = true
			opsiText[idx] = strings.TrimSpace(m[2])
			opsiMode = true
			lastOpsi = idx
			continue
		}
		// 3. Baris lanjutan.
		if opsiMode && lastOpsi >= 0 && cur == "pertanyaan" {
			// Opsi multi-paragraf: tempel ke opsi terakhir.
			if opsiText[lastOpsi] == "" {
				opsiText[lastOpsi] = ln
			} else {
				opsiText[lastOpsi] += " " + ln
			}
			continue
		}
		if cur == "" {
			cur = "pertanyaan" // baris liar sebelum label apa pun = pertanyaan
		}
		appendVal(cur, ln)
	}

	tipe := strings.ToLower(strings.TrimSpace(fields["tipe"]))
	pertanyaan := strings.TrimSpace(fields["pertanyaan"])
	stimulus := strings.TrimSpace(fields["stimulus"])
	pembahasan := strings.TrimSpace(fields["pembahasan"])
	kunciStr := strings.TrimSpace(fields["kunci"])
	media := strings.TrimSpace(fields["media"])
	altMedia := strings.TrimSpace(fields["alt_media"])
	kisi := strings.TrimSpace(fields["kisi_kisi"])
	kompetensi := strings.TrimSpace(fields["kompetensi"])

	fail := func(format string, args ...any) (*csvSoal, []map[string]any) {
		return nil, append(findings, docxErr(nomor, format, args...))
	}

	if tipe == "" {
		return fail("tipe wajib diisi")
	}
	if !ValidTipe(tipe) {
		return fail("tipe %q tidak dikenal", tipe)
	}
	if pertanyaan == "" {
		return fail("pertanyaan wajib diisi")
	}
	// Aturan kunci kosong (sama seperti Excel): error untuk 12 tipe
	// dinilai otomatis; warning untuk tipe rubrik; diabaikan untuk
	// pernyataan_sikap.
	switch tipe {
	case "pernyataan_sikap":
		if kunciStr != "" {
			findings = append(findings, docxInfo(nomor,
				"kunci diabaikan untuk pernyataan_sikap (skala 1–4, tidak dinilai)."))
		}
		kunciStr = ""
	case "uraian", "upload_file", "praktik_singkat":
		if kunciStr == "" {
			findings = append(findings, docxWarn(nomor,
				"kunci kosong — rubrik belum diisi. Rubrik bisa dilengkapi di editor sesudah import."))
		}
	default:
		if kunciStr == "" {
			return fail("kunci wajib diisi")
		}
	}
	bobot := 1
	if raw := strings.TrimSpace(fields["bobot"]); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			return fail("bobot %q harus bilangan bulat > 0", raw)
		}
		bobot = n
	}
	level := NormalLevel(fields["level"])
	if level == "" {
		level = "sedang"
	}
	if !ValidLevel(level) {
		return fail("level %q harus mudah / sedang / sulit", fields["level"])
	}

	// Bangun jawaban/kunci lewat konstruktor yang sama dengan CSV,
	// agar semantik kunci per tipe identik (pemakaian get dari map).
	vals := map[string]string{
		"tipe": tipe, "pertanyaan": pertanyaan, "stimulus": stimulus,
		"pembahasan": pembahasan, "kunci": kunciStr,
		"bobot": fields["bobot"], "level": fields["level"],
	}
	for i, lbl := range []string{"a", "b", "c", "d", "e"} {
		if opsiSeen[i] {
			vals["opsi_"+lbl] = opsiText[i]
		}
	}
	get := func(_ []string, col string) string { return vals[col] }
	jawaban, kunci, err := buildCSVSoal(tipe, nil, get, kunciStr, stimulus)
	if err != nil {
		return fail("%s", err.Error())
	}
	if err := ValidateSoal(tipe, jawaban, kunci); err != nil {
		return fail("%s", err.Error())
	}
	if tipe == "teks_rumpang" {
		var k map[string]any
		_ = json.Unmarshal(kunci, &k)
		if ks, ok := asStringSlice(k["jawaban"]); ok && strings.Count(pertanyaan, "___") != len(ks) {
			return fail("jumlah rumpang ___ harus sama dengan jumlah jawaban kunci (pisah dengan |)")
		}
	}
	if tipe == "studi_kasus" && stimulus == "" {
		return fail("stimulus wajib diisi untuk tipe studi_kasus")
	}
	if tipe == "pilihan_gambar" {
		labels := []string{"A", "B", "C", "D", "E"}
		for i := 0; i < 5; i++ {
			if opsiSeen[i] && !isHTTPURL(opsiText[i]) {
				return fail("opsi %s harus URL gambar http(s)", labels[i])
			}
		}
	}
	if media != "" && !isHTTPURL(media) {
		return fail("media_url %q bukan URL http(s) yang valid", media)
	}
	// MEDIA / ALT MEDIA kini ikut tersimpan (csvSoal membawa fieldnya;
	// gabungkanDocx memindahkannya ke excelSoal + memberi W2/W7).
	// KOMPETENSI dibawa mentah; pemetaan ke ID via lookup di
	// gabungkanDocx (E31 bila kode tidak ditemukan — paritas .xlsx).
	// KISI-KISI: E30 seperti jalur .xlsx (tabel kisi_kisi tanpa kolom
	// kode, jadi tidak ada yang bisa dicocokkan).
	if kisi != "" {
		findings = append(findings, docxErrKode(nomor, "E30",
			"kisi_kisi %q tidak ditemukan di bank ini", kisi))
	}
	// Invarian: temuan tingkat error di blok ini (mis. label/opsi
	// ganda) → soal tidak dikembalikan.
	for _, f := range findings {
		if f["tingkat"] == "error" {
			return nil, findings
		}
	}
	return &csvSoal{
		baris: nomor, tipe: tipe, pertanyaan: pertanyaan, stimulus: stimulus,
		jawaban: jawaban, kunci: kunci, pembahasan: pembahasan,
		bobot: bobot, level: level,
		mediaURL: media, mediaAlt: altMedia, kisiKisi: kisi, kompetensi: kompetensi,
	}, findings
}

// --- Pembuatan template .docx (ZIP + document.xml manual) ---

// docxPara: satu paragraf template. Size dalam half-point
// (22 = 11pt); 0 = default 22.
type docxPara struct {
	Text string
	Bold bool
	Size int
}

// buildTemplateDocx menghasilkan dokumen Word FORMAT KITA: petunjuk
// singkat + 3 contoh soal lengkap (pg, menjodohkan, uraian). .docx
// dirakit manual sebagai ZIP (stdlib archive/zip) tanpa dependensi
// tambahan — hanya 4 part minimal yang diterima Word / LibreOffice /
// Google Docs: [Content_Types].xml, _rels/.rels, word/document.xml,
// word/styles.xml.
func buildTemplateDocx() ([]byte, error) {
	paras := []docxPara{
		{Text: "Template Import Soal — CBT SMK v2 (Format Word)", Bold: true, Size: 32},
		{},
		{Text: "CARA MENGISI (baca dulu — 1 menit):", Bold: true},
		{Text: "1. Satu blok soal diawali baris SOAL 1, SOAL 2, dan seterusnya. Teks sebelum SOAL 1 dianggap petunjuk dan diabaikan."},
		{Text: "2. Di dalam tiap blok, tulis label persis seperti contoh, diakhiri titik dua: TIPE: STIMULUS: PERTANYAAN: KUNCI: BOBOT: LEVEL: PEMBAHASAN: KISI-KISI: KOMPETENSI: MEDIA: ALT MEDIA:"},
		{Text: "3. TIPE: wajib diisi dengan KODE (bukan nama panjang). Contoh benar: pg. Contoh salah: Pilihan Ganda. Daftar kode 16 tipe: pg, pg_kompleks, menjodohkan, isian, uraian, benar_salah, pg_multi, urutkan, isi_ganda, teks_rumpang, peta_konsep, studi_kasus, pernyataan_sikap, upload_file, praktik_singkat, pilihan_gambar."},
		{Text: "4. Opsi ditulis satu per baris: A. ..., B. ..., sampai E. Tulis opsi SETELAH pertanyaan dan SEBELUM label KUNCI."},
		{Text: "5. KUNCI: formatnya beda per tipe — pg: satu huruf (A); pg_kompleks / pg_multi: huruf dipisah koma (A,C); benar_salah: B/S dipisah koma (B,S,B); menjodohkan / peta_konsep: pasangan kiri::kanan dipisah titik koma; isian / isi_ganda / teks_rumpang: jawaban dipisah garis tegak |; uraian / upload_file / praktik_singkat: tulis RUBRIK (boleh kosong, nanti dilengkapi di editor); pernyataan_sikap: KUNCI diabaikan (skala 1–4, tidak dinilai)."},
		{Text: "6. STIMULUS: wajib untuk studi_kasus. PERTANYAAN: wajib. BOBOT: angka bulat lebih dari 0 (kosong = 1). LEVEL: mudah / sedang / sulit (kosong = sedang)."},
		{Text: "7. MEDIA: satu URL gambar (https://...). Gambar yang ditempel langsung di dokumen TIDAK ikut terimpor — harus berupa tautan. ALT MEDIA: teks alternatif gambar tersebut (untuk pembaca layar)."},
		{Text: "8. KISI-KISI / KOMPETENSI: isi dengan KODE yang tampil di halaman Kisi-kisi / Kompetensi bank ini. Kosongkan bila belum dipetakan."},
		{Text: "9. Batas: maks 500 soal per file, maks 5 MB. Soal hasil import berstatus DRAFT — wajib ditelaah sebelum dipakai."},
		{},
		{Text: "CONTOH (hapus contoh ini dan isi dengan soal Anda, atau salin polanya):", Bold: true},
		{},
		{Text: "SOAL 1", Bold: true},
		{Text: "TIPE: pg"},
		{Text: "STIMULUS:"},
		{Text: "PERTANYAAN: Ibu kota provinsi Jawa Barat adalah …"},
		{Text: "A. Bandung"},
		{Text: "B. Semarang"},
		{Text: "C. Medan"},
		{Text: "D. Surabaya"},
		{Text: "KUNCI: A"},
		{Text: "BOBOT: 1"},
		{Text: "LEVEL: sedang"},
		{Text: "PEMBAHASAN: Jawa Barat beribu kota di Bandung."},
		{Text: "KISI-KISI:"},
		{Text: "KOMPETENSI:"},
		{Text: "MEDIA:"},
		{Text: "ALT MEDIA:"},
		{},
		{Text: "SOAL 2", Bold: true},
		{Text: "TIPE: menjodohkan"},
		{Text: "STIMULUS:"},
		{Text: "PERTANYAAN: Pasangkan kota dengan provinsinya."},
		{Text: "A. Bandung::Jawa Barat"},
		{Text: "B. Semarang::Jawa Tengah"},
		{Text: "C. Surabaya::Jawa Timur"},
		{Text: "D. Medan::Sumatera Utara"},
		{Text: "KUNCI: Bandung::Jawa Barat;Semarang::Jawa Tengah;Surabaya::Jawa Timur;Medan::Sumatera Utara"},
		{Text: "BOBOT: 1"},
		{Text: "LEVEL: sedang"},
		{Text: "PEMBAHASAN: Pasangan kota–provinsi yang benar sesuai peta Indonesia."},
		{Text: "KISI-KISI:"},
		{Text: "KOMPETENSI:"},
		{Text: "MEDIA:"},
		{Text: "ALT MEDIA:"},
		{},
		{Text: "SOAL 3", Bold: true},
		{Text: "TIPE: uraian"},
		{Text: "STIMULUS:"},
		{Text: "PERTANYAAN: Jelaskan perbedaan antara RAM dan ROM beserta contoh penggunaannya dalam kehidupan sehari-hari."},
		{Text: "KUNCI: Rubrik: pengertian RAM (30), pengertian ROM (30), perbedaan (20), contoh (20)."},
		{Text: "BOBOT: 2"},
		{Text: "LEVEL: sulit"},
		{Text: "PEMBAHASAN:"},
		{Text: "KISI-KISI:"},
		{Text: "KOMPETENSI:"},
		{Text: "MEDIA:"},
		{Text: "ALT MEDIA:"},
	}
	return docxFromParagraphs(paras)
}

// docxFromParagraphs merakit .docx minimal dari daftar paragraf.
func docxFromParagraphs(paras []docxPara) ([]byte, error) {
	var doc bytes.Buffer
	doc.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n")
	doc.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>` + "\n")
	for _, p := range paras {
		if strings.TrimSpace(p.Text) == "" {
			doc.WriteString("<w:p/>\n")
			continue
		}
		sz := p.Size
		if sz == 0 {
			sz = 22
		}
		doc.WriteString("<w:p><w:r><w:rPr>")
		if p.Bold {
			doc.WriteString("<w:b/>")
		}
		doc.WriteString(`<w:sz w:val="` + strconv.Itoa(sz) + `"/><w:szCs w:val="` + strconv.Itoa(sz) + `"/>`)
		doc.WriteString(`</w:rPr><w:t xml:space="preserve">`)
		_ = xml.EscapeText(&doc, []byte(p.Text))
		doc.WriteString("</w:t></w:r></w:p>\n")
	}
	doc.WriteString("</w:body></w:document>")

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	writeEntry := func(name, content string) error {
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(content))
		return err
	}
	writeBytes := func(name string, content []byte) error {
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		_, err = w.Write(content)
		return err
	}
	parts := []struct{ name, content string }{
		{"[Content_Types].xml", docxContentTypesXML},
		{"_rels/.rels", docxRelsXML},
		{"word/styles.xml", docxStylesXML},
	}
	for _, part := range parts {
		if err := writeEntry(part.name, part.content); err != nil {
			return nil, err
		}
	}
	if err := writeBytes("word/document.xml", doc.Bytes()); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

const docxContentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">` +
	`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>` +
	`<Default Extension="xml" ContentType="application/xml"/>` +
	`<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>` +
	`<Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>` +
	`</Types>`

const docxRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
	`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>` +
	`</Relationships>`

const docxStylesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">` +
	`<w:docDefaults><w:rPrDefault><w:rPr><w:rFonts w:ascii="Calibri" w:hAnsi="Calibri"/><w:sz w:val="22"/><w:szCs w:val="22"/><w:lang w:val="id-ID"/></w:rPr></w:rPrDefault>` +
	`<w:pPrDefault><w:pPr><w:spacing w:after="120" w:line="276" w:lineRule="auto"/></w:pPr></w:pPrDefault></w:docDefaults>` +
	`</w:styles>`
