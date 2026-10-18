// Package banksoal — import soal Excel/Word (jalur baru).
//
// File ini HANYA berisi jalur import baru (Excel .xlsx + Word .docx).
// Endpoint JSON CSV lama di banksoal.go dipertahankan (deprecated) dan
// TIDAK diubah. Wiring ke server dilakukan koordinator dengan memanggil
// RegisterImportExcel dari main.go.
//
// Spesifikasi: docs/ANALISIS-IMPORT-SOAL.md Bab 2 (16 kolom kanonik,
// katalog error E1–E32, warning W1–W7, duplikat, preview dry-run,
// mode all_or_nothing/skip_errors + pilihan duplikat lewati/simpan_ganda).
//
// Perbedaan yang disengaja terhadap spesifikasi (dicatat jujur):
//   - Endpoint template: tugas ini memakai
//     GET /api/v1/bank-soal/import/template.xlsx; alias
//     GET /api/v1/bank-soal/import/template (dari spesifikasi) juga
//     didaftarkan agar kedua acuan jalan.
//   - Kolom kisi_kisi: tabel kisi_kisi TIDAK punya kolom `kode`
//     (Bab 4 Q2 spesifikasi, belum terjawab — tidak ada migrasi yang
//     menambahkannya di tugas ini). Maka nilai kisi_kisi yang terisi
//     selalu menghasilkan E30 `tidak ditemukan di bank ini` sampai
//     migrasi kolom kode tersedia. Kolom kompetensi BEKERJA karena
//     tabel kompetensi punya kolom kode (dicocokkan case-insensitive
//     dalam mapel bank tujuan).
//   - Kolom media_alt: tabel soal TIDAK punya kolom media_alt, jadi
//     nilai ini hanya tampil di preview, TIDAK disimpan ke database.
//   - Menjodohkan/peta_konsep: ValidateSoal TIDAK dipanggil untuk dua
//     tipe ini di jalur baru, sesuai §2.4 spesifikasi (pasangan kunci
//     boleh memuat pasangan yang tidak ada di opsi tampil — pengecoh
//     di sisi kanan didukung). Untuk 14 tipe lain ValidateSoal tetap
//     dipanggil sebagai jaring pengaman.
//   - Preview juga menerima .docx (via parseDocxImport Tim Word) agar
//     guru bisa meninjau file Word sebelum simpan; spesifikasi hanya
//     mewajibkan preview .xlsx, penerimaan .docx di preview adalah
//     perluasan yang tidak mengubah perilaku .xlsx.
//   - Error level file yang berasal dari ISI file (tidak bisa dibaca,
//     sheet "Soal" hilang, header wajib hilang, > 500 baris) dikembalikan
//     sebagai body 200 {errors: [...], disimpan: false} mengikuti pola
//     handler CSV lama. Error level REQUEST (tidak ada file, file
//     kosong, > 5 MB, ekstensi/magic bytes salah, bank tidak ditemukan,
//     bank terkunci, mode tidak valid) memakai HTTP error biasa.
//
// File .docx diparsing Tim Word lewat parseDocxImport (file
// import_word.go, paket yang sama). File ini hanya memanggilnya —
// lihat gabungkanDocx untuk kontrak temuannya (field "tingkat").
package banksoal

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/xuri/excelize/v2"
)

const (
	// Batas file import (spesifikasi §2.2): maks 500 baris soal, maks 5 MB.
	maxImportXLSXBaris = 500
	maxImportXLSXBytes = 5 << 20
)

// RegisterImportExcel mendaftarkan endpoint import soal jalur baru.
//
//	GET  /api/v1/bank-soal/import/template.xlsx  (banksoal.import)
//	GET  /api/v1/bank-soal/import/template       (alias, banksoal.import)
//	POST /api/v1/bank-soal/import/preview        (banksoal.import)
//	POST /api/v1/bank-soal/import/xlsx           (banksoal.import)
func RegisterImportExcel(e *echo.Echo, pool *pgxpool.Pool) {
	templateHandler := func(c *echo.Context) error {
		_, role, ok := requirePerm(c, "banksoal.import")
		if !ok {
			return nil
		}
		if !forbidSoalUntukProktor(c, role) {
			return nil
		}
		data, err := buildTemplateXLSX()
		if err != nil {
			return platform.WriteError(c, http.StatusInternalServerError, "internal_error", "Gagal membuat template.")
		}
		c.Response().Header().Set("Content-Disposition",
			`attachment; filename="template-import-soal-cbt-smk-v2.xlsx"`)
		return c.Blob(http.StatusOK,
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
	}
	e.GET("/api/v1/bank-soal/import/template.xlsx", templateHandler)
	e.GET("/api/v1/bank-soal/import/template", templateHandler)

	e.POST("/api/v1/bank-soal/import/preview", func(c *echo.Context) error {
		_, role, ok := requirePerm(c, "banksoal.import")
		if !ok {
			return nil
		}
		if !forbidSoalUntukProktor(c, role) {
			return nil
		}
		data, format, errMsg := bacaFileImport(c)
		if errMsg != "" {
			return platform.WriteError(c, http.StatusBadRequest, "validation_error", errMsg)
		}
		bank, okBank := muatBankImport(c, pool, "", role)
		if !okBank {
			return nil
		}
		if bank.IsLocked {
			return platform.WriteError(c, http.StatusConflict, "bank_locked", "Bank soal terkunci — import ditolak.")
		}
		res := parseImportFile(c.Request().Context(), pool, bank, format, data)
		return platform.WriteOK(c, map[string]any{
			"bank_id": bank.ID, "total_baris": res.total,
			"valid": len(res.valid), "error": len(res.errors),
			"warning": len(res.warnings), "duplikat": res.duplikat,
			"ringkasan_tipe": res.ringkasan, "total_bobot_valid": res.totalBobot,
			"errors": res.errors, "warnings": res.warnings,
			"duplikat_detail": res.dupDetail, "preview": res.preview,
		})
	})

	e.POST("/api/v1/bank-soal/import/xlsx", func(c *echo.Context) error {
		userID, role, ok := requirePerm(c, "banksoal.import")
		if !ok {
			return nil
		}
		if !forbidSoalUntukProktor(c, role) {
			return nil
		}
		mode := strings.ToLower(strings.TrimSpace(c.FormValue("mode")))
		if mode == "" {
			mode = "all_or_nothing"
		}
		if mode != "all_or_nothing" && mode != "skip_errors" {
			return platform.WriteError(c, http.StatusBadRequest, "validation_error", "mode harus all_or_nothing atau skip_errors.")
		}
		duplikat := strings.ToLower(strings.TrimSpace(c.FormValue("duplikat")))
		if duplikat == "" {
			duplikat = "lewati"
		}
		if duplikat != "lewati" && duplikat != "simpan_ganda" {
			return platform.WriteError(c, http.StatusBadRequest, "validation_error", "duplikat harus lewati atau simpan_ganda.")
		}
		data, format, errMsg := bacaFileImport(c)
		if errMsg != "" {
			return platform.WriteError(c, http.StatusBadRequest, "validation_error", errMsg)
		}
		bank, okBank := muatBankImport(c, pool, userID, role)
		if !okBank {
			return nil
		}
		if bank.IsLocked {
			return platform.WriteError(c, http.StatusConflict, "bank_locked", "Bank soal terkunci — import ditolak.")
		}
		res := parseImportFile(c.Request().Context(), pool, bank, format, data)
		errs := res.errors
		if mode == "all_or_nothing" && len(errs) > 0 {
			return platform.WriteOK(c, map[string]any{
				"sukses": 0, "gagal": len(errs), "dilewati_duplikat": 0,
				"warnings": res.warnings, "errors": errs,
				"mode": mode, "duplikat": duplikat, "disimpan": false,
			})
		}
		ctx := c.Request().Context()
		sukses := 0
		dilewati := 0
		for _, row := range res.valid {
			if row.duplikat && duplikat == "lewati" {
				dilewati++
				continue
			}
			var kisiArg, kompArg any
			if row.kisiKisiID != "" {
				kisiArg = row.kisiKisiID
			}
			if row.kompetensiID != "" {
				kompArg = row.kompetensiID
			}
			_, err := pool.Exec(ctx,
				`INSERT INTO soal (bank_id, tipe, kisi_kisi_id, kompetensi_id, stimulus,
				 pertanyaan, media_url, jawaban, kunci, pembahasan, bobot, level, status)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,'draft')`,
				bank.ID, row.tipe, kisiArg, kompArg, row.stimulus, row.pertanyaan,
				row.mediaURL, row.jawaban, row.kunci, row.pembahasan, row.bobot, row.level)
			if err != nil {
				errs = append(errs, map[string]any{
					"baris": row.baris, "kode": "E32", "pesan": "Gagal menyimpan ke database.",
				})
				continue
			}
			sukses++
		}
		_ = audit.Catat(ctx, pool, userID, "banksoal.import", "bank_soal", bank.ID,
			map[string]any{"sukses": sukses, "gagal": len(errs), "dilewati_duplikat": dilewati,
				"mode": mode, "duplikat": duplikat, "format": format}, c.RealIP())
		return platform.WriteOK(c, map[string]any{
			"sukses": sukses, "gagal": len(errs), "dilewati_duplikat": dilewati,
			"warnings": res.warnings, "errors": errs,
			"mode": mode, "duplikat": duplikat, "disimpan": sukses > 0,
		})
	})
}

// muatBankImport memvalidasi bank_id dari form multipart + scope guru.
// userID/role berasal dari requirePerm yang sudah dipanggil handler.
// Mengembalikan (bank, true) bila OK, atau (nil, false) bila respons
// error sudah ditulis (handler cukup `return nil`).
func muatBankImport(c *echo.Context, pool *pgxpool.Pool, userID, role string) (*bankRow, bool) {
	bankID := strings.TrimSpace(c.FormValue("bank_id"))
	if !isValidUUID(bankID) {
		_ = platform.WriteError(c, http.StatusBadRequest, "validation_error", "bank_id wajib UUID bank yang valid.")
		return nil, false
	}
	b, found, err := loadBank(c, pool, bankID)
	if err != nil {
		_ = platform.WriteError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca bank soal.")
		return nil, false
	}
	if !found {
		_ = platform.WriteError(c, http.StatusNotFound, "not_found", "Bank soal tidak ditemukan.")
		return nil, false
	}
	if !checkGuruScope(c, role, userID, b) {
		return nil, false
	}
	return b, true
}

// --- Baca + validasi file upload ---

// bacaFileImport membaca field multipart "file": cek ekstensi, ukuran,
// magic bytes ZIP, dan anggota arsip sesuai format (.xlsx wajib punya
// xl/workbook.xml, .docx wajib punya word/document.xml — kedua format
// sama-sama ZIP sehingga magic bytes saja tidak cukup membedakan).
// Mengembalikan (data, format, "") bila OK, atau ("", "", pesan) bila gagal.
func bacaFileImport(c *echo.Context) ([]byte, string, string) {
	if err := c.Request().ParseMultipartForm((maxImportXLSXBytes) + (1 << 20)); err != nil {
		return nil, "", "Body multipart tidak valid (maks. 5 MB)."
	}
	fh, err := c.FormFile("file")
	if err != nil {
		return nil, "", "Field file wajib diunggah."
	}
	if fh.Size <= 0 {
		return nil, "", "File kosong."
	}
	nama := strings.ToLower(fh.Filename)
	var format string
	switch {
	case strings.HasSuffix(nama, ".xlsx"):
		format = "xlsx"
	case strings.HasSuffix(nama, ".docx"):
		format = "docx"
	default:
		return nil, "", `Hanya file .xlsx dan .docx yang diterima. Unduh template dari tombol "Unduh Template" bila belum punya.`
	}
	src, err := fh.Open()
	if err != nil {
		return nil, "", "File tidak bisa dibuka."
	}
	defer src.Close()
	// Baca dengan batas 5 MB + 1 agar klaim ukuran di header tidak bisa menipu.
	data, err := io.ReadAll(io.LimitReader(src, maxImportXLSXBytes+1))
	if err != nil {
		return nil, "", "Gagal membaca berkas upload."
	}
	if int64(len(data)) > maxImportXLSXBytes {
		return nil, "", "File melebihi batas 5 MB."
	}
	if !importMagicZIP(data) {
		return nil, "", "File tidak bisa dibaca: bukan file .xlsx/.docx yang valid (magic bytes tidak cocok)."
	}
	anggota := "xl/workbook.xml"
	if format == "docx" {
		anggota = "word/document.xml"
	}
	if !zipMemiliki(data, anggota) {
		return nil, "", fmt.Sprintf("File tidak bisa dibaca: isi arsip tidak sesuai format .%s (tidak ditemukan %s).", format, anggota)
	}
	return data, format, ""
}

func importMagicZIP(data []byte) bool {
	if len(data) < 4 || data[0] != 'P' || data[1] != 'K' {
		return false
	}
	// Arsip ZIP: PK\x03\x04 (entri biasa), PK\x05\x06 (arsip kosong),
	// PK\x07\x08 (spanned).
	if (data[2] == 3 || data[2] == 5 || data[2] == 7) && (data[3] == 4 || data[3] == 6 || data[3] == 8) {
		return true
	}
	return false
}

func zipMemiliki(data []byte, nama string) bool {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return false
	}
	for _, f := range zr.File {
		if f.Name == nama {
			return true
		}
	}
	return false
}

func importValidHTTPURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	if u.Host == "" {
		return false
	}
	return true
}

// --- Model hasil parse ---

// excelSoal = satu soal hasil import jalur baru: csvSoal (tipe,
// pertanyaan, stimulus, jawaban, kunci, pembahasan, bobot, level)
// ditambah kolom baru jalur Excel.
type excelSoal struct {
	csvSoal
	mediaURL       string
	mediaAlt       string
	kisiKisi       string // kode mentah dari kolom kisi_kisi
	kompetensi     string // kode mentah dari kolom kompetensi
	kisiKisiID     string // UUID hasil lookup ("" = NULL; saat ini selalu "" — lihat catatan file)
	kompetensiID   string // UUID hasil lookup ("" = NULL)
	warnings       []map[string]any
	duplikat       bool
	duplikatBaris  int    // baris pertama di file dengan pertanyaan sama (W4)
	duplikatSoalID string // id soal yang sudah ada di bank (W5)
}

type importParseResult struct {
	valid      []excelSoal
	errors     []map[string]any // {baris, kode, pesan}
	warnings   []map[string]any // {baris, kode, pesan} — baris 0 = level file
	preview    []map[string]any
	total      int // baris data tak-kosong (valid + error)
	duplikat   int
	dupDetail  []map[string]any
	ringkasan  map[string]int
	totalBobot int
}

// parseImportFile mendispatch ke parser sesuai format.
func parseImportFile(ctx context.Context, pool *pgxpool.Pool, bank *bankRow, format string, data []byte) importParseResult {
	if format == "docx" {
		rows, errs := parseDocxImport(data)
		return gabungkanDocx(ctx, pool, bank, rows, errs)
	}
	return parseXLSXImport(ctx, pool, bank, data)
}

// --- Parser .xlsx ---

var importKolomKanonic = []string{
	"tipe", "pertanyaan", "stimulus",
	"opsi_a", "opsi_b", "opsi_c", "opsi_d", "opsi_e",
	"kunci", "bobot", "level", "pembahasan",
	"media_url", "media_alt", "kisi_kisi", "kompetensi",
}

var importKolomDikenal = func() map[string]bool {
	m := map[string]bool{}
	for _, c := range importKolomKanonic {
		m[c] = true
	}
	return m
}()

func parseXLSXImport(ctx context.Context, pool *pgxpool.Pool, bank *bankRow, data []byte) importParseResult {
	var res importParseResult
	res.ringkasan = map[string]int{}
	fileErr := func(pesan string) importParseResult {
		res.errors = append(res.errors, map[string]any{"baris": 0, "kode": "", "pesan": pesan})
		return res
	}
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return fileErr("File tidak bisa dibaca: " + err.Error())
	}
	defer f.Close()
	punyaSoal := false
	for _, s := range f.GetSheetList() {
		if s == "Soal" {
			punyaSoal = true
		}
	}
	if !punyaSoal {
		return fileErr(`Sheet "Soal" tidak ditemukan. Gunakan template resmi — jangan mengganti nama sheet.`)
	}
	rows, err := f.GetRows("Soal")
	if err != nil {
		return fileErr("File tidak bisa dibaca: " + err.Error())
	}
	if len(rows) < 2 {
		return fileErr("File harus punya header + minimal 1 baris soal.")
	}
	header := map[string]int{}
	for i, h := range rows[0] {
		key := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(h, "\ufeff")))
		if key != "" {
			header[key] = i
		}
	}
	for _, col := range []string{"tipe", "pertanyaan", "kunci"} {
		if _, ok := header[col]; !ok {
			res.errors = append(res.errors, map[string]any{"baris": 1, "kode": "", "pesan": "Kolom wajib hilang di header: " + col})
			return res
		}
	}
	// W3: header tak dikenal → warning level file (baris 0).
	for _, h := range rows[0] {
		key := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(h, "\ufeff")))
		if key != "" && !importKolomDikenal[key] {
			res.warnings = append(res.warnings, map[string]any{
				"baris": 0, "kode": "W3",
				"pesan": fmt.Sprintf("Kolom %q tidak dikenal — diabaikan.", strings.TrimSpace(h)),
			})
		}
	}
	dataRows := rows[1:]
	if len(dataRows) > maxImportXLSXBaris {
		return fileErr("File melebihi batas 500 baris soal.")
	}
	get := func(row []string, col string) string {
		idx, ok := header[col]
		if !ok || idx >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[idx])
	}
	// Cache lookup kompetensi: kode-lower -> id ("" bila tidak ditemukan).
	kompCache := map[string]string{}
	kompMiss := map[string]bool{}

	failRow := func(baris int, kode, pesan, tipe, pertanyaan, kunciStr string) {
		e := map[string]any{"baris": baris, "kode": kode, "pesan": pesan}
		res.errors = append(res.errors, e)
		res.preview = append(res.preview, map[string]any{
			"baris": baris, "tipe": tipe, "pertanyaan": pertanyaan,
			"kunci": kunciStr, "errors": []map[string]any{e},
		})
	}

	for i, row := range dataRows {
		baris := i + 2 // header = baris 1
		kosong := true
		for _, v := range row {
			if strings.TrimSpace(v) != "" {
				kosong = false
				break
			}
		}
		if kosong {
			continue
		}
		res.total++
		tipe := strings.ToLower(strings.TrimSpace(get(row, "tipe")))
		pertanyaan := get(row, "pertanyaan")
		kunciStr := get(row, "kunci")
		stimulus := get(row, "stimulus")
		pembahasan := get(row, "pembahasan")
		mediaURL := get(row, "media_url")
		mediaAlt := get(row, "media_alt")
		kisiKode := get(row, "kisi_kisi")
		kompKode := get(row, "kompetensi")
		var rowWarn []map[string]any
		warn := func(kode, pesan string) {
			rowWarn = append(rowWarn, map[string]any{"baris": baris, "kode": kode, "pesan": pesan})
		}

		// E1–E6: tipe / pertanyaan / kunci / bobot / level.
		if tipe == "" {
			failRow(baris, "E1", "tipe wajib diisi", tipe, pertanyaan, kunciStr)
			continue
		}
		if !ValidTipe(tipe) {
			failRow(baris, "E2", fmt.Sprintf("tipe %q tidak dikenal", tipe), tipe, pertanyaan, kunciStr)
			continue
		}
		if pertanyaan == "" {
			failRow(baris, "E3", "pertanyaan wajib diisi", tipe, pertanyaan, kunciStr)
			continue
		}
		// E4 / W1 / W6: aturan kunci kosong (§2.3 kolom 9).
		switch tipe {
		case "pernyataan_sikap":
			if kunciStr != "" {
				warn("W6", "kunci diabaikan untuk pernyataan_sikap (skala 1–4, tidak dinilai).")
			}
		case "uraian", "upload_file", "praktik_singkat":
			if kunciStr == "" {
				warn("W1", "kunci kosong — rubrik belum diisi. Rubrik bisa dilengkapi di editor sesudah import.")
			}
		default:
			if kunciStr == "" {
				failRow(baris, "E4", "kunci wajib diisi", tipe, pertanyaan, kunciStr)
				continue
			}
		}
		bobot := 1
		if raw := get(row, "bobot"); raw != "" {
			n, err := strconv.Atoi(raw)
			if err != nil || n <= 0 {
				failRow(baris, "E5", fmt.Sprintf("bobot %q harus bilangan bulat > 0", raw), tipe, pertanyaan, kunciStr)
				continue
			}
			bobot = n
		}
		level := NormalLevel(get(row, "level"))
		if level == "" {
			level = "sedang"
		}
		if !ValidLevel(level) {
			failRow(baris, "E6", fmt.Sprintf("level %q harus mudah / sedang / sulit", level), tipe, pertanyaan, kunciStr)
			continue
		}
		// E28: media_url harus URL http(s) valid.
		if mediaURL != "" && !importValidHTTPURL(mediaURL) {
			failRow(baris, "E28", fmt.Sprintf("media_url %q bukan URL http(s) yang valid", mediaURL), tipe, pertanyaan, kunciStr)
			continue
		}
		// Bangun jawaban/kunci JSON lewat builder CSV lama (pesan
		// E7–E24 dipertahankan verbatim; kode dipetakan dari pesan).
		getByName := func(_ []string, col string) string { return get(row, col) }
		jawaban, kunci, err := buildCSVSoal(tipe, nil, getByName, kunciStr, stimulus)
		if err != nil {
			failRow(baris, kodeErrorImport(tipe, err.Error()), err.Error(), tipe, pertanyaan, kunciStr)
			continue
		}
		// E25: jumlah rumpang = jumlah jawaban kunci.
		if tipe == "teks_rumpang" {
			var k map[string]any
			_ = json.Unmarshal(kunci, &k)
			if ks, ok := asStringSlice(k["jawaban"]); ok && strings.Count(pertanyaan, "___") != len(ks) {
				failRow(baris, "E25", "jumlah rumpang ___ harus sama dengan jumlah jawaban kunci (pisah dengan |)", tipe, pertanyaan, kunciStr)
				continue
			}
		}
		// E26: stimulus wajib untuk studi_kasus.
		if tipe == "studi_kasus" && stimulus == "" {
			failRow(baris, "E26", "stimulus wajib diisi untuk tipe studi_kasus", tipe, pertanyaan, kunciStr)
			continue
		}
		// E29: opsi pilihan_gambar harus URL http(s).
		if tipe == "pilihan_gambar" {
			if labelBuruk := opsiBukanURL(jawaban); labelBuruk != "" {
				failRow(baris, "E29", fmt.Sprintf("opsi %s harus URL gambar http(s)", labelBuruk), tipe, pertanyaan, kunciStr)
				continue
			}
		}
		// Jaring pengaman: ValidateSoal untuk semua tipe KECUALI
		// menjodohkan/peta_konsep (§2.4: kunci boleh memuat pasangan
		// yang tidak ada di opsi tampil — pengecoh sisi kanan).
		if tipe != "menjodohkan" && tipe != "peta_konsep" {
			if err := ValidateSoal(tipe, jawaban, kunci); err != nil {
				failRow(baris, "", err.Error(), tipe, pertanyaan, kunciStr)
				continue
			}
		}
		// E30: kisi_kisi — selalu gagal sampai tabel kisi_kisi punya
		// kolom kode (lihat catatan di kepala file, Bab 4 Q2).
		if kisiKode != "" {
			failRow(baris, "E30", fmt.Sprintf("kisi_kisi %q tidak ditemukan di bank ini", kisiKode), tipe, pertanyaan, kunciStr)
			continue
		}
		// E31: kompetensi — lookup kode (case-insensitive) di mapel bank.
		var kompID string
		if kompKode != "" {
			key := strings.ToLower(kompKode)
			switch {
			case kompMiss[key]:
				failRow(baris, "E31", fmt.Sprintf("kompetensi %q tidak ditemukan", kompKode), tipe, pertanyaan, kunciStr)
				continue
			case kompCache[key] != "":
				kompID = kompCache[key]
			default:
				var id string
				err := pool.QueryRow(ctx,
					`SELECT id FROM kompetensi WHERE mapel_id=$1 AND LOWER(TRIM(kode))=$2`,
					bank.MapelID, strings.ToLower(strings.TrimSpace(kompKode))).Scan(&id)
				if err != nil {
					kompMiss[key] = true
					failRow(baris, "E31", fmt.Sprintf("kompetensi %q tidak ditemukan", kompKode), tipe, pertanyaan, kunciStr)
					continue
				}
				kompCache[key] = id
				kompID = id
			}
		}
		// W2/W7: media_url terisi.
		if mediaURL != "" {
			if mediaAlt == "" {
				warn("W2", "media_url diisi tetapi media_alt kosong — gambar tanpa teks alternatif.")
			}
			warn("W7", "URL gambar tidak diperiksa keterjangkauannya — klik URL di preview untuk memastikan gambar bisa dibuka.")
		}
		res.valid = append(res.valid, excelSoal{
			csvSoal: csvSoal{
				baris: baris, tipe: tipe, pertanyaan: pertanyaan, stimulus: stimulus,
				jawaban: jawaban, kunci: kunci, pembahasan: pembahasan,
				bobot: bobot, level: level,
			},
			mediaURL: mediaURL, mediaAlt: mediaAlt,
			kisiKisi: kisiKode, kompetensi: kompKode, kompetensiID: kompID,
			warnings: rowWarn,
		})
	}
	dupDetail, dup := tandaiDuplikat(ctx, pool, bank.ID, res.valid)
	res.dupDetail = dupDetail
	res.duplikat = dup
	finalisasiImport(&res)
	return res
}

// opsiBukanURL mengembalikan label opsi pertama yang bukan URL http(s),
// atau "" bila semua opsi berupa URL valid (khusus pilihan_gambar).
func opsiBukanURL(jawaban json.RawMessage) string {
	var j map[string]any
	if err := json.Unmarshal(jawaban, &j); err != nil {
		return ""
	}
	arr, ok := j["opsi"].([]any)
	if !ok {
		return ""
	}
	for _, it := range arr {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		label, _ := m["label"].(string)
		teks, _ := m["teks"].(string)
		if !importValidHTTPURL(teks) {
			return label
		}
	}
	return ""
}

// kodeErrorImport memetakan pesan error builder CSV lama (verbatim dari
// spesifikasi §2.6.2) ke kode E7–E27. Pesan yang tidak dikenali -> "".
func kodeErrorImport(tipe, pesan string) string {
	switch {
	case strings.Contains(pesan, "minimal 2 opsi terisi"):
		switch tipe {
		case "pg", "studi_kasus", "pilihan_gambar":
			return "E7"
		case "pg_kompleks", "pg_multi":
			return "E9"
		}
	case strings.Contains(pesan, "tidak cocok dengan label opsi yang terisi"):
		switch tipe {
		case "pg", "studi_kasus", "pilihan_gambar":
			return "E8"
		case "pg_kompleks", "pg_multi":
			return "E10"
		case "urutkan":
			return "E21"
		}
	case strings.Contains(pesan, "kunci wajib minimal satu label opsi"):
		return "E11"
	case strings.Contains(pesan, "isi pernyataan (minimal 2)"):
		return "E12"
	case strings.Contains(pesan, "nilai B/S, dipisah koma"):
		return "E13"
	case strings.Contains(pesan, "harus B atau S"):
		return "E14"
	case strings.Contains(pesan, "isi pasangan format kiri::kanan"):
		return "E15"
	case strings.Contains(pesan, "harus format kiri::kanan"):
		return "E16"
	case strings.Contains(pesan, "kunci harus pasangan format kiri::kanan"):
		return "E17"
	case strings.Contains(pesan, "kunci isian wajib diisi"):
		return "E18"
	case strings.Contains(pesan, "isi butir yang diurutkan"):
		return "E19"
	case strings.Contains(pesan, "label opsi (urutan benar)"):
		return "E20"
	case strings.Contains(pesan, "duplikat"):
		return "E22"
	case strings.Contains(pesan, "kunci wajib jawaban per kolom"):
		return "E23"
	case strings.Contains(pesan, "kunci wajib jawaban per rumpang"):
		return "E24"
	case strings.Contains(pesan, "stimulus wajib diisi"):
		return "E26"
	case strings.Contains(pesan, "isi pernyataan sikap"):
		return "E27"
	}
	return ""
}

// --- Duplikat (§2.7) ---

// tandaiDuplikat memberi flag duplikat pada baris valid: W4 (duplikat
// dalam file — menyebut baris pertama) dan W5 (duplikat dengan soal
// yang sudah ada di bank). Normalisasi pertanyaan = trim + lowercase +
// rapatkan spasi (NormalisasiIsian). Duplikat adalah warning, bukan
// error: tidak pernah menggagalkan file dalam mode all_or_nothing.
func tandaiDuplikat(ctx context.Context, pool *pgxpool.Pool, bankID string, rows []excelSoal) ([]map[string]any, int) {
	sudahAda := map[string]string{}
	qrows, err := pool.Query(ctx, `SELECT id, pertanyaan FROM soal WHERE bank_id=$1`, bankID)
	if err == nil {
		defer qrows.Close()
		for qrows.Next() {
			var id, p string
			if err := qrows.Scan(&id, &p); err == nil {
				n := NormalisasiIsian(p)
				if _, ok := sudahAda[n]; !ok {
					sudahAda[n] = id
				}
			}
		}
	}
	terlihat := map[string]int{}
	var detail []map[string]any
	dup := 0
	for i := range rows {
		n := NormalisasiIsian(rows[i].pertanyaan)
		flag := false
		if pertama, ok := terlihat[n]; ok {
			rows[i].duplikat = true
			rows[i].duplikatBaris = pertama
			flag = true
			rows[i].warnings = append(rows[i].warnings, map[string]any{
				"baris": rows[i].baris, "kode": "W4",
				"pesan": fmt.Sprintf("pertanyaan duplikat dengan baris %d.", pertama),
			})
			detail = append(detail, map[string]any{
				"baris": rows[i].baris, "baris_duplikat": pertama,
				"pesan": fmt.Sprintf("pertanyaan duplikat dengan baris %d.", pertama),
			})
		} else {
			terlihat[n] = rows[i].baris
		}
		if soalID, ok := sudahAda[n]; ok {
			rows[i].duplikat = true
			rows[i].duplikatSoalID = soalID
			flag = true
			rows[i].warnings = append(rows[i].warnings, map[string]any{
				"baris": rows[i].baris, "kode": "W5",
				"pesan": "pertanyaan sudah ada di bank ini.",
			})
			detail = append(detail, map[string]any{
				"baris": rows[i].baris, "soal_id": soalID,
				"pesan": "pertanyaan sudah ada di bank ini.",
			})
		}
		if flag {
			dup++
		}
	}
	return detail, dup
}

// finalisasiImport menyusun warnings gabungan, preview baris valid,
// dan ringkasan — dipakai jalur .xlsx maupun .docx.
func finalisasiImport(res *importParseResult) {
	for _, row := range res.valid {
		res.warnings = append(res.warnings, row.warnings...)
	}
	sort.SliceStable(res.warnings, func(i, j int) bool {
		return intDariAny(res.warnings[i]["baris"]) < intDariAny(res.warnings[j]["baris"])
	})
	for _, row := range res.valid {
		res.preview = append(res.preview, previewBarisValid(row))
	}
	sort.SliceStable(res.preview, func(i, j int) bool {
		return intDariAny(res.preview[i]["baris"]) < intDariAny(res.preview[j]["baris"])
	})
	if res.ringkasan == nil {
		res.ringkasan = map[string]int{}
	}
	for _, row := range res.valid {
		res.ringkasan[row.tipe]++
		res.totalBobot += row.bobot
	}
}

// previewBarisValid membangun satu baris preview untuk soal valid,
// termasuk kunci yang terpetakan ke teks opsi (bukan hanya huruf).
func previewBarisValid(es excelSoal) map[string]any {
	info := TipeSoal[es.tipe]
	warns := es.warnings
	if warns == nil {
		warns = []map[string]any{}
	}
	var jawabanAny any
	_ = json.Unmarshal(es.jawaban, &jawabanAny)
	return map[string]any{
		"baris": es.baris, "tipe": es.tipe, "tipe_label": info.Label,
		"kelompok": info.Kelompok, "pertanyaan": es.pertanyaan,
		"stimulus": es.stimulus, "opsi": jawabanAny,
		"kunci_terpetakan": kunciTerpetakan(es.tipe, es.jawaban, es.kunci),
		"bobot":            es.bobot, "level": es.level, "pembahasan": es.pembahasan,
		"media_url": es.mediaURL, "media_alt": es.mediaAlt,
		"kisi_kisi": es.kisiKisi, "kompetensi": es.kompetensi,
		"duplikat": es.duplikat, "warnings": warns,
		"errors": []map[string]any{},
	}
}

// kunciTerpetakan memetakan kunci ke bentuk yang bisa diverifikasi
// guru di preview: label + teks opsi yang dirujuk, pasangan
// kiri::kanan, jawaban per kolom/rumpang, atau teks rubrik.
func kunciTerpetakan(tipe string, jawabanRaw, kunciRaw json.RawMessage) any {
	var j, k map[string]any
	_ = json.Unmarshal(jawabanRaw, &j)
	_ = json.Unmarshal(kunciRaw, &k)
	opsiMap := map[string]string{}
	if arr, ok := j["opsi"].([]any); ok {
		for _, it := range arr {
			if m, ok := it.(map[string]any); ok {
				label, _ := m["label"].(string)
				teks, _ := m["teks"].(string)
				opsiMap[label] = teks
			}
		}
	}
	switch tipe {
	case "pg", "studi_kasus", "pilihan_gambar":
		label, _ := k["opsi"].(string)
		return map[string]any{"label": label, "teks": opsiMap[label]}
	case "pg_kompleks", "pg_multi":
		ks, _ := asStringSlice(k["opsi"])
		out := []map[string]any{}
		for _, l := range ks {
			out = append(out, map[string]any{"label": l, "teks": opsiMap[l]})
		}
		return out
	case "benar_salah":
		per, _ := asStringSlice(j["pernyataan"])
		nilai, _ := asBoolSlice(k["nilai"])
		out := []map[string]any{}
		for i, p := range per {
			v := "S"
			if i < len(nilai) && nilai[i] {
				v = "B"
			}
			out = append(out, map[string]any{"pernyataan": p, "nilai": v})
		}
		return out
	case "menjodohkan", "peta_konsep":
		return k["pasangan"]
	case "isian", "teks_rumpang":
		return k["jawaban"]
	case "isi_ganda":
		return map[string]any{"kolom": j["kolom"], "jawaban": k["jawaban"]}
	case "uraian", "upload_file", "praktik_singkat":
		rubrik, _ := k["rubrik"].(string)
		return rubrik
	case "urutkan":
		items, _ := asStringSlice(j["item"])
		ord, _ := asFloatSlice(k["urutan"])
		labels := []string{"A", "B", "C", "D", "E"}
		var labs, teks []string
		for _, f := range ord {
			idx := int(f)
			if idx < 0 || idx >= len(items) {
				continue
			}
			labs = append(labs, labels[idx])
			teks = append(teks, items[idx])
		}
		return map[string]any{"urutan_label": labs, "urutan_teks": teks}
	case "pernyataan_sikap":
		return "tidak dinilai (skala 1-4)"
	}
	return nil
}

// --- Gabung hasil .docx (Tim Word) ---

// gabungkanDocx mengubah hasil parseDocxImport (Tim Word) menjadi
// importParseResult yang sama dengan jalur .xlsx: deteksi duplikat,
// preview, dan ringkasan. Baris .docx dianggap sudah divalidasi Tim
// Word (temuan mereka diteruskan apa adanya); validasi ulang tidak
// dilakukan di sini agar tidak ada dua sumber kebenaran validasi.
//
// Kontrak parseDocxImport (lihat import_word.go): temuan memakai field
// "tingkat" ("error" | "warning" | "info") — HANYA tingkat error yang
// dihitung sebagai error; "baris" pada temuan .docx adalah NOMOR SOAL
// (label "SOAL <n>"), BUKAN nomor baris dokumen; blok dengan temuan
// error tidak dikembalikan sebagai baris. csvSoal dari .docx membawa
// media_url / media_alt / kompetensi mentah (kisi_kisi selalu E30 di
// parser, paritas .xlsx); di sini media dipindahkan ke excelSoal
// (W2/W7 seperti .xlsx) dan kompetensi dipetakan ke ID via lookup
// kode (E31 bila tidak ditemukan — paritas .xlsx).
func gabungkanDocx(ctx context.Context, pool *pgxpool.Pool, bank *bankRow, rows []csvSoal, findings []map[string]any) importParseResult {
	var res importParseResult
	res.ringkasan = map[string]int{}
	errByNomor := map[int][]map[string]any{}
	warnByNomor := map[int][]map[string]any{}
	for _, e := range findings {
		tingkat := strDariAny(e["tingkat"])
		if tingkat == "" {
			tingkat = "error" // tanpa tingkat = error (kontrak Tim Word)
		}
		baris := intDariAny(e["baris"])
		pesan := strDariAny(e["pesan"])
		if pesan == "" {
			pesan = "Baris gagal divalidasi."
		}
		entri := map[string]any{"baris": baris, "kode": strDariAny(e["kode"]), "tingkat": tingkat, "pesan": pesan}
		if tingkat == "error" {
			errByNomor[baris] = append(errByNomor[baris], entri)
		} else {
			warnByNomor[baris] = append(warnByNomor[baris], entri)
		}
	}
	// Satu baris preview per nomor soal yang error (terurut).
	nomorError := make([]int, 0, len(errByNomor))
	for n := range errByNomor {
		nomorError = append(nomorError, n)
	}
	sort.Ints(nomorError)
	for _, n := range nomorError {
		errs := errByNomor[n]
		res.errors = append(res.errors, errs...)
		res.preview = append(res.preview, map[string]any{
			"baris": n, "tipe": "", "pertanyaan": "", "errors": errs,
		})
	}
	// Warning/info level file (baris 0) langsung ke warnings global.
	res.warnings = append(res.warnings, warnByNomor[0]...)
	kompCache := map[string]string{}
	kompMiss := map[string]bool{}
	for _, r := range rows {
		es := excelSoal{csvSoal: r}
		// Pindahkan nilai mentah .docx ke field excelSoal (paritas .xlsx).
		es.mediaURL = r.mediaURL
		es.mediaAlt = r.mediaAlt
		es.kisiKisi = r.kisiKisi
		es.kompetensi = r.kompetensi
		// Kompetensi: lookup kode → ID (cache per file, query sama
		// seperti jalur .xlsx). Gagal → E31, baris pindah ke error.
		if r.kompetensi != "" {
			key := strings.ToLower(r.kompetensi)
			switch {
			case kompMiss[key]:
				errEntri := map[string]any{"baris": r.baris, "kode": "E31", "tingkat": "error",
					"pesan": fmt.Sprintf("kompetensi %q tidak ditemukan", r.kompetensi)}
				res.errors = append(res.errors, errEntri)
				res.preview = append(res.preview, map[string]any{
					"baris": r.baris, "tipe": r.tipe, "pertanyaan": r.pertanyaan, "errors": []map[string]any{errEntri}})
				continue
			case kompCache[key] != "":
				es.kompetensiID = kompCache[key]
			default:
				var id string
				err := pool.QueryRow(ctx,
					`SELECT id FROM kompetensi WHERE mapel_id=$1 AND LOWER(TRIM(kode))=$2`,
					bank.MapelID, strings.ToLower(strings.TrimSpace(r.kompetensi))).Scan(&id)
				if err != nil {
					kompMiss[key] = true
					errEntri := map[string]any{"baris": r.baris, "kode": "E31", "tingkat": "error",
						"pesan": fmt.Sprintf("kompetensi %q tidak ditemukan", r.kompetensi)}
					res.errors = append(res.errors, errEntri)
					res.preview = append(res.preview, map[string]any{
						"baris": r.baris, "tipe": r.tipe, "pertanyaan": r.pertanyaan, "errors": []map[string]any{errEntri}})
					continue
				}
				kompCache[key] = id
				es.kompetensiID = id
			}
		}
		// W2/W7 media (paritas jalur .xlsx).
		if r.mediaURL != "" {
			if r.mediaAlt == "" {
				es.warnings = append(es.warnings, map[string]any{"baris": r.baris, "kode": "W2", "tingkat": "warning",
					"pesan": "media_url diisi tetapi media_alt kosong — gambar tanpa teks alternatif."})
			}
			es.warnings = append(es.warnings, map[string]any{"baris": r.baris, "kode": "W7", "tingkat": "warning",
				"pesan": "URL gambar tidak diperiksa keterjangkauannya — klik URL di preview untuk memastikan gambar bisa dibuka."})
		}
		es.warnings = append(es.warnings, warnByNomor[r.baris]...)
		res.valid = append(res.valid, es)
	}
	// total .docx = soal valid + blok soal yang error (distinct nomor
	// > 0; error level file baris 0 bukan soal).
	jumlahBlokError := 0
	for n := range errByNomor {
		if n > 0 {
			jumlahBlokError++
		}
	}
	res.total = len(rows) + jumlahBlokError
	dupDetail, dup := tandaiDuplikat(ctx, pool, bank.ID, res.valid)
	res.dupDetail = dupDetail
	res.duplikat = dup
	finalisasiImport(&res)
	return res
}

func intDariAny(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	case string:
		i, _ := strconv.Atoi(strings.TrimSpace(n))
		return i
	default:
		return 0
	}
}

func strDariAny(v any) string {
	switch s := v.(type) {
	case string:
		return s
	case fmt.Stringer:
		return s.String()
	default:
		return ""
	}
}

// --- Template .xlsx (4 sheet, spesifikasi §2.2, §2.5, §2.9) ---

func buildTemplateXLSX() ([]byte, error) {
	f := excelize.NewFile()
	headerStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	petunjukStyle, _ := f.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{WrapText: true, Vertical: "top"}})

	// Sheet 1 — Petunjuk (§2.9, teks spesifikasi).
	_ = f.SetSheetName("Sheet1", "Petunjuk")
	for i, baris := range petunjukTemplate {
		_ = f.SetCellValue("Petunjuk", fmt.Sprintf("A%d", i+1), baris)
	}
	_ = f.SetColWidth("Petunjuk", "A", "A", 140)
	_ = f.SetCellStyle("Petunjuk", "A1", fmt.Sprintf("A%d", len(petunjukTemplate)), petunjukStyle)

	// Sheet 2 — Soal (satu-satunya yang dibaca parser).
	if _, err := f.NewSheet("Soal"); err != nil {
		return nil, err
	}
	for i, col := range importKolomKanonic {
		_ = f.SetCellValue("Soal", fmt.Sprintf("%s1", kolomHuruf(i)), col)
	}
	_ = f.SetCellStyle("Soal", "A1", fmt.Sprintf("%s1", kolomHuruf(len(importKolomKanonic)-1)), headerStyle)
	lebarSoal := map[string]float64{"A": 18, "B": 60, "C": 40, "D": 30, "E": 30, "F": 30, "G": 30, "H": 30, "I": 40, "J": 8, "K": 10, "L": 40, "M": 40, "N": 30, "O": 15, "P": 15}
	for huruf, lebar := range lebarSoal {
		_ = f.SetColWidth("Soal", huruf, huruf, lebar)
	}

	// Sheet 3 — Contoh (16 baris, §2.5).
	if _, err := f.NewSheet("Contoh"); err != nil {
		return nil, err
	}
	for i, col := range importKolomKanonic {
		_ = f.SetCellValue("Contoh", fmt.Sprintf("%s1", kolomHuruf(i)), col)
	}
	_ = f.SetCellStyle("Contoh", "A1", fmt.Sprintf("%s1", kolomHuruf(len(importKolomKanonic)-1)), headerStyle)
	for r, baris := range contohTemplate {
		for c, nilai := range baris {
			_ = f.SetCellValue("Contoh", fmt.Sprintf("%s%d", kolomHuruf(c), r+2), nilai)
		}
	}
	for huruf, lebar := range lebarSoal {
		_ = f.SetColWidth("Contoh", huruf, huruf, lebar)
	}

	// Sheet 4 — Referensi Tipe (16 tipe + format kuncinya).
	if _, err := f.NewSheet("Referensi Tipe"); err != nil {
		return nil, err
	}
	headerRef := []string{"kode", "label", "kelompok", "format kunci", "pemisah", "kolom opsi yang dipakai"}
	for i, col := range headerRef {
		_ = f.SetCellValue("Referensi Tipe", fmt.Sprintf("%s1", kolomHuruf(i)), col)
	}
	_ = f.SetCellStyle("Referensi Tipe", "A1", fmt.Sprintf("%s1", kolomHuruf(len(headerRef)-1)), headerStyle)
	for r, baris := range referensiTipeTemplate {
		for c, nilai := range baris {
			_ = f.SetCellValue("Referensi Tipe", fmt.Sprintf("%s%d", kolomHuruf(c), r+2), nilai)
		}
	}
	for huruf, lebar := range map[string]float64{"A": 18, "B": 24, "C": 18, "D": 44, "E": 12, "F": 40} {
		_ = f.SetColWidth("Referensi Tipe", huruf, huruf, lebar)
	}

	// Guru mendarat langsung di sheet input.
	f.SetActiveSheet(1)
	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func kolomHuruf(i int) string {
	// 0 -> A, 25 -> Z, 26 -> AA. Template hanya memakai A–P.
	if i < 26 {
		return string(rune('A' + i))
	}
	return string(rune('A'+i/26-1)) + string(rune('A'+i%26))
}

// petunjukTemplate = isi sheet Petunjuk (spesifikasi §2.9).
var petunjukTemplate = []string{
	"CARA MENGISI TEMPLATE IMPORT SOAL — CBT SMK v2",
	"",
	`1. Isi HANYA sheet "Soal". Sheet "Contoh" dan "Referensi Tipe" hanya contoh. Jangan mengganti nama sheet dan jangan menghapus/mengganti nama kolom.`,
	"2. Satu baris = satu soal. Urutan baris = urutan soal masuk ke bank. Baris yang semua selnya kosong akan dilewati (tidak dihitung error).",
	`3. Kolom "tipe" wajib diisi dengan KODE, bukan nama panjang. Contoh benar: pg. Contoh salah: Pilihan Ganda. Daftar kode ada di sheet "Referensi Tipe" dan sheet "Contoh".`,
	`4. Kolom "pertanyaan" wajib diisi. Tulis satu pertanyaan yang jelas.`,
	`5. Kolom "kunci" — cara menulisnya BERBEDA per tipe (ini sumber salah yang paling sering terjadi, jadi baca baris tipe Anda di bawah):`,
	"   - pg / studi_kasus / pilihan_gambar: SATU huruf, mis. B",
	"   - pg_kompleks / pg_multi: huruf dipisah KOMA, mis. A,C",
	"   - benar_salah: B/S dipisah KOMA sebanyak pernyataan, mis. B,S,B",
	"   - menjodohkan / peta_konsep: pasangan kiri::kanan, antar-pasangan dipisah TITIK KOMA, mis. Bandung::Jawa Barat;Semarang::Jawa Tengah",
	"   - isian: jawaban, alternatif dipisah GARIS TEGAK |, mis. Jakarta|DKI Jakarta",
	"   - isi_ganda / teks_rumpang: jawaban dipisah GARIS TEGAK |, mis. 50|30",
	"   - uraian / upload_file / praktik_singkat: tulis RUBRIK di kolom kunci. Boleh kosong (nanti dilengkapi di editor), tetapi akan muncul peringatan \"rubrik belum diisi\".",
	"   - pernyataan_sikap: kolom kunci DIABAIKAN (skala 1-4, tidak dinilai).",
	`6. Kolom "opsi_a" s.d. "opsi_e":`,
	"   - Untuk pg / pg_kompleks / pg_multi / benar_salah / urutkan / pernyataan_sikap / studi_kasus: isi dengan teks opsi/pernyataan/butir.",
	"   - Untuk menjodohkan / peta_konsep: isi dengan pasangan kiri::kanan.",
	"   - Untuk isian / isi_ganda / teks_rumpang / uraian / upload_file / praktik_singkat: kolom opsi TIDAK dipakai, kosongkan saja.",
	"   - Untuk pilihan_gambar: isi dengan URL gambar (https://...), bukan teks. Satu opsi = satu gambar.",
	"   - Minimal 2 opsi terisi untuk tipe yang memakai opsi.",
	`7. Kolom "bobot": angka bulat lebih dari 0. Kosong = 1. Kolom "level": mudah / sedang / sulit. Kosong = sedang.`,
	`8. Kolom "stimulus": teks bacaan/kasus. WAJIB diisi untuk studi_kasus. Bila beberapa soal memakai stimulus yang sama, salin teksnya ke tiap baris soal tersebut.`,
	`9. Kolom "pembahasan": penjelasan jawaban. Opsional.`,
	`10. Kolom "media_url": SATU tautan gambar (https://...) untuk soal ini. Gambar yang ditempel langsung di file Excel TIDAK ikut terimpor — harus berupa tautan. Kolom "media_alt": teks alternatif gambar (untuk pembaca layar). Bila media_url diisi tetapi media_alt kosong, muncul peringatan (tidak menggagalkan import).`,
	`11. Kolom "kisi_kisi" dan "kompetensi": isi dengan KODE yang tampil di halaman Kisi-kisi / Kompetensi bank ini. Bila kode tidak ditemukan, baris tersebut ERROR (tidak disimpan diam-diam tanpa kisi-kisi). Kosongkan bila belum dipetakan — soal tetap masuk sebagai draft.`,
	`12. Teks rumpang ditulis dengan tiga garis bawah: ___ Contoh: "Ibu kota Indonesia adalah ___." Jawabannya di kolom kunci.`,
	`13. Jangan menulis rumus Excel / formula di sel mana pun. Tulis teks biasa. Rumus matematika dalam soal ditulis sebagai teks, mis. "2x + 3".`,
	"14. Batas: maks 500 baris soal per file, maks 5 MB. Bila soal Anda lebih dari 500, pecah menjadi beberapa file.",
	`15. Alur: unggah file → PERIKSA PRATINJAU (pastikan kunci terpetakan ke opsi yang benar — jangan hanya melihat "tidak ada error") → pilih mode (tolak semua bila ada error / lewati baris yang error) → pilih perlakuan duplikat (lewati / simpan ganda) → Simpan. Soal hasil import berstatus DRAFT — wajib ditelaah sebelum dipakai.`,
	"16. Bila ada error, perbaiki di file Excel Anda lalu unggah ulang. Daftar error menyebut NOMOR BARIS Excel (baris 1 = header).",
}

// contohTemplate = 16 baris contoh (spesifikasi §2.5), urutan kolom =
// importKolomKanonic: tipe, pertanyaan, stimulus, opsi_a..e, kunci,
// bobot, level, pembahasan, media_url, media_alt, kisi_kisi, kompetensi.
// Kolom yang tidak disebut di spesifikasi dibiarkan kosong (default).
var contohTemplate = [][]string{
	{"pg", "Ibu kota provinsi Jawa Barat adalah …", "", "Bandung", "Semarang", "Medan", "Surabaya", "", "A", "", "", "", "", "", "", ""},
	{"pg_kompleks", "Manakah yang termasuk perangkat keras komputer? (Pilih semua yang benar.)", "", "Monitor", "Sistem operasi", "Keyboard", "Peramban", "", "A,C", "", "", "", "", "", "", ""},
	{"menjodohkan", "Pasangkan kota dengan provinsinya.", "", "Bandung::Jawa Barat", "Semarang::Jawa Tengah", "Surabaya::Jawa Timur", "Medan::Sumatera Utara", "", "Bandung::Jawa Barat;Semarang::Jawa Tengah;Surabaya::Jawa Timur;Medan::Sumatera Utara", "", "", "", "", "", "", ""},
	{"isian", "Ibu kota negara Indonesia adalah ….", "", "", "", "", "", "", "Jakarta|DKI Jakarta", "", "", "", "", "", "", ""},
	{"uraian", "Jelaskan perbedaan antara RAM dan ROM beserta contoh penggunaannya dalam kehidupan sehari-hari.", "", "", "", "", "", "", "Rubrik: pengertian RAM (30), pengertian ROM (30), perbedaan (20), contoh (20).", "", "", "", "", "", "", ""},
	{"benar_salah", "Tentukan benar atau salah untuk setiap pernyataan berikut.", "", "Air mendidih pada suhu 100°C pada tekanan 1 atm.", "Bulan menghasilkan cahayanya sendiri.", "Fotosintesis menghasilkan oksigen.", "", "", "B,S,B", "", "", "", "", "", "", ""},
	{"pg_multi", "Manakah yang termasuk bilangan prima? (Pilih semua yang benar.)", "", "2", "4", "5", "9", "", "A,C", "", "", "", "", "", "", ""},
	{"urutkan", "Urutkan langkah mencuci tangan yang benar.", "", "Bilas tangan dengan air bersih.", "Basahi tangan dengan air mengalir.", "Gosok kedua telapak tangan dengan sabun.", "Keringkan tangan.", "", "B,C,A,D", "", "", "", "", "", "", ""},
	{"isi_ganda", "Sebuah persegi panjang memiliki panjang 10 dan lebar 5. Isi kolom berikut.", "", "", "", "", "", "", "50|30", "", "", "", "", "", "", ""},
	{"teks_rumpang", "Ibu kota Indonesia adalah ___ dan mata uangnya adalah ___.", "", "", "", "", "", "", "Jakarta|Rupiah", "", "", "", "", "", "", ""},
	{"peta_konsep", "Pasangkan konsep dengan pasangannya.", "", "Gula::Manis", "Jeruk::Asam", "Garam::Asin", "", "", "Gula::Manis;Jeruk::Asam;Garam::Asin", "", "", "", "", "", "", ""},
	{"studi_kasus", "Berdasarkan kasus di atas, tindakan pertama yang paling tepat adalah ….", "Sebuah bengkel menerima komputer yang tidak mau menyala sama sekali: tidak ada lampu indikator, tidak ada suara kipas.", "Langsung mengganti motherboard", "Memeriksa kabel daya dan stopkontak", "Menginstal ulang sistem operasi", "Mengganti RAM", "", "B", "", "", "", "", "", "", ""},
	{"pernyataan_sikap", "Beri tanggapanmu untuk setiap pernyataan berikut.", "", "Saya senang belajar dengan diskusi kelompok.", "Saya lebih paham bila belajar sambil praktik langsung.", "", "", "", "", "", "", "", "", "", "", ""},
	{"upload_file", "Buatlah laporan praktikum sesuai format yang diberikan, lalu unggah berkasnya.", "", "", "", "", "", "", "Rubrik: kelengkapan isi (40), format laporan (30), ketepatan data (30).", "", "", "", "", "", "", ""},
	{"praktik_singkat", "Praktikkan cara membuat kabel UTP straight dan jelaskan hasilnya.", "", "", "", "", "", "", "Rubrik: urutan warna benar (50), konektor terpasang rapi (30), penjelasan hasil (20).", "", "", "", "", "", "", ""},
	{"pilihan_gambar", "Manakah gambar yang menunjukkan kabel UTP?", "", "https://contoh.sekolah.id/gambar/utp.png", "https://contoh.sekolah.id/gambar/hdmi.png", "https://contoh.sekolah.id/gambar/vga.png", "https://contoh.sekolah.id/gambar/power.png", "", "A", "", "", "", "", "", "", ""},
}

// referensiTipeTemplate = sheet Referensi Tipe: kode, label, kelompok,
// format kunci, pemisah, kolom opsi yang dipakai.
var referensiTipeTemplate = [][]string{
	{"pg", "Pilihan Ganda", "[Inti AKM]", "Satu label, mis. B", "-", "opsi_a..e: teks opsi"},
	{"pg_kompleks", "PG Kompleks", "[Inti AKM]", "Label dipisah koma, mis. A,C", ",", "opsi_a..e: teks opsi"},
	{"menjodohkan", "Menjodohkan", "[Inti AKM]", "Pasangan kiri::kanan, antar-pasangan dipisah ;", "; dan ::", "opsi_a..e: pasangan kiri::kanan"},
	{"isian", "Isian / Jawaban Singkat", "[Inti AKM]", "Jawaban, alternatif dipisah |", "|", "(tidak dipakai)"},
	{"uraian", "Uraian / Esai", "[Inti AKM]", "Rubrik (boleh kosong + warning)", "-", "(tidak dipakai)"},
	{"benar_salah", "Benar-Salah", "[Varian aplikasi]", "Nilai B/S dipisah koma, mis. B,S,B", ",", "opsi_a..e: pernyataan"},
	{"pg_multi", "PG Multi-Jawaban", "[Varian aplikasi]", "Label dipisah koma, mis. A,C", ",", "opsi_a..e: teks opsi"},
	{"urutkan", "Mengurutkan", "[Varian aplikasi]", "Label urutan benar dipisah koma, mis. C,A,B,D", ",", "opsi_a..e: butir"},
	{"isi_ganda", "Isian Multi-Kolom", "[Varian aplikasi]", "Jawaban per kolom dipisah |", "|", "(tidak dipakai)"},
	{"teks_rumpang", "Teks Rumpang", "[Varian aplikasi]", "Jawaban per rumpang dipisah |", "|", "(tidak dipakai)"},
	{"peta_konsep", "Peta Konsep", "[Varian aplikasi]", "Pasangan kiri::kanan, antar-pasangan dipisah ;", "; dan ::", "opsi_a..e: pasangan kiri::kanan"},
	{"studi_kasus", "Studi Kasus + PG", "[Varian aplikasi]", "Satu label, mis. B (+ stimulus wajib)", "-", "opsi_a..e: teks opsi"},
	{"pernyataan_sikap", "Skala Sikap", "[Varian aplikasi]", "Diabaikan (skala 1-4, tidak dinilai)", "-", "opsi_a..e: pernyataan sikap"},
	{"upload_file", "Upload Berkas", "[Varian aplikasi]", "Rubrik (boleh kosong + warning)", "-", "(tidak dipakai)"},
	{"praktik_singkat", "Praktik Singkat", "[Varian aplikasi]", "Rubrik (boleh kosong + warning)", "-", "(tidak dipakai)"},
	{"pilihan_gambar", "Pilihan Gambar", "[Varian aplikasi]", "Satu label, mis. B", "-", "opsi_a..e: URL gambar"},
}
