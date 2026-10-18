package banksoal

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// --- Input ---

type bankInput struct {
	Kode         string  `json:"kode"`
	Nama         string  `json:"nama"`
	MapelID      string  `json:"mapel_id"`
	KelasTingkat int     `json:"kelas_tingkat"`
	GuruID       *string `json:"guru_id"`
}

type soalInput struct {
	Tipe         string          `json:"tipe"`
	KisiKisiID   *string         `json:"kisi_kisi_id"`
	KompetensiID *string         `json:"kompetensi_id"`
	Stimulus     string          `json:"stimulus"`
	Pertanyaan   string          `json:"pertanyaan"`
	MediaURL     string          `json:"media_url"`
	Jawaban      json.RawMessage `json:"jawaban"`
	Kunci        json.RawMessage `json:"kunci"`
	Pembahasan   string          `json:"pembahasan"`
	Bobot        int             `json:"bobot"`
	Level        string          `json:"level"`
	Status       string          `json:"status"`
}

func (in *soalInput) validate() error {
	if !ValidTipe(in.Tipe) {
		return fmt.Errorf("tipe %q tidak dikenal", in.Tipe)
	}
	if strings.TrimSpace(in.Pertanyaan) == "" {
		return fmt.Errorf("pertanyaan wajib diisi")
	}
	if in.Tipe == "studi_kasus" && strings.TrimSpace(in.Stimulus) == "" {
		return fmt.Errorf("stimulus wajib diisi untuk tipe studi_kasus")
	}
	if in.Bobot <= 0 {
		return fmt.Errorf("bobot harus > 0")
	}
	if in.Level != "" && !ValidLevel(in.Level) {
		return fmt.Errorf("level harus mudah / sedang / sulit")
	}
	if in.Status != "" {
		switch in.Status {
		case "draft", "aktif", "nonaktif", "perlu_revisi", "diarsipkan":
		default:
			return fmt.Errorf("status %q tidak dikenal", in.Status)
		}
	}
	if in.Tipe == "teks_rumpang" {
		// Jumlah rumpang (___) harus sama dengan jumlah jawaban kunci.
		var kunci map[string]any
		if err := json.Unmarshal(in.Kunci, &kunci); err == nil {
			if ks, ok := asStringSlice(kunci["jawaban"]); ok {
				if strings.Count(in.Pertanyaan, "___") != len(ks) {
					return fmt.Errorf("jumlah rumpang ___ (%d) harus sama dengan jumlah jawaban kunci (%d)",
						strings.Count(in.Pertanyaan, "___"), len(ks))
				}
			}
		}
	}
	return ValidateSoal(in.Tipe, in.Jawaban, in.Kunci)
}

func ptrVal(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// --- Import CSV ---

type csvSoal struct {
	baris      int
	tipe       string
	pertanyaan string
	stimulus   string
	jawaban    json.RawMessage
	kunci      json.RawMessage
	pembahasan string
	bobot      int
	level      string
	// Nilai mentah dari parser Word (.docx) — gabungkanDocx memindahkannya
	// ke excelSoal: media_url tersimpan, kompetensi dipetakan ke ID via
	// lookup kode, kisi_kisi ditolak E30 (sama seperti jalur .xlsx).
	// Jalur CSV tidak mengisi keempatnya (CSV tidak punya kolomnya).
	// Catatan: excelSoal juga punya field bernama sama — selektor
	// es.mediaURL dll. menunjuk field excelSoal (depth 0 menaungi
	// field embedded depth 1), jadi tidak ada ambiguitas.
	mediaURL   string
	mediaAlt   string
	kisiKisi   string
	kompetensi string
}

var csvColumns = []string{
	"tipe", "pertanyaan", "opsi_a", "opsi_b", "opsi_c", "opsi_d", "opsi_e",
	"kunci", "bobot", "level", "stimulus",
}

// parseCSVImport mem-parsing teks CSV menjadi soal yang sudah divalidasi.
// Kolom: tipe,pertanyaan,opsi_a,opsi_b,opsi_c,opsi_d,opsi_e,kunci,bobot,level,stimulus
// Maks 500 baris data. Baris yang gagal dikembalikan di errors dan TIDAK
// disimpan (pemanggil memutuskan all_or_nothing vs skip_errors).
func parseCSVImport(csvText string) ([]csvSoal, []map[string]any) {
	r := csv.NewReader(strings.NewReader(csvText))
	r.TrimLeadingSpace = true
	r.FieldsPerRecord = -1
	records, err := r.ReadAll()
	if err != nil {
		return nil, []map[string]any{{"baris": 0, "pesan": "CSV tidak bisa dibaca: " + err.Error()}}
	}
	if len(records) < 2 {
		return nil, []map[string]any{{"baris": 0, "pesan": "CSV harus punya header + minimal 1 baris soal."}}
	}
	header := map[string]int{}
	for i, h := range records[0] {
		header[strings.ToLower(strings.TrimSpace(h))] = i
	}
	var errors []map[string]any
	fail := func(baris int, msg string) {
		errors = append(errors, map[string]any{"baris": baris, "pesan": msg})
	}
	for _, col := range []string{"tipe", "pertanyaan", "kunci"} {
		if _, ok := header[col]; !ok {
			return nil, []map[string]any{{"baris": 1, "pesan": "Kolom wajib hilang di header: " + col}}
		}
	}
	dataRows := records[1:]
	if len(dataRows) > 500 {
		return nil, []map[string]any{{"baris": 0, "pesan": "CSV melebihi batas 500 baris soal."}}
	}
	get := func(rec []string, col string) string {
		idx, ok := header[col]
		if !ok || idx >= len(rec) {
			return ""
		}
		return strings.TrimSpace(rec[idx])
	}
	var out []csvSoal
	for i, rec := range dataRows {
		baris := i + 2 // header = baris 1
		// Lewati baris kosong.
		allEmpty := true
		for _, v := range rec {
			if strings.TrimSpace(v) != "" {
				allEmpty = false
			}
		}
		if allEmpty {
			continue
		}
		tipe := strings.ToLower(strings.TrimSpace(get(rec, "tipe")))
		pertanyaan := strings.TrimSpace(get(rec, "pertanyaan"))
		kunciStr := strings.TrimSpace(get(rec, "kunci"))
		stimulus := strings.TrimSpace(get(rec, "stimulus"))
		if !ValidTipe(tipe) {
			fail(baris, fmt.Sprintf("tipe %q tidak dikenal", tipe))
			continue
		}
		if pertanyaan == "" {
			fail(baris, "pertanyaan wajib diisi")
			continue
		}
		if kunciStr == "" {
			fail(baris, "kunci wajib diisi")
			continue
		}
		bobot := 1
		if raw := get(rec, "bobot"); raw != "" {
			n, err := strconv.Atoi(raw)
			if err != nil || n <= 0 {
				fail(baris, fmt.Sprintf("bobot %q harus bilangan bulat > 0", raw))
				continue
			}
			bobot = n
		}
		level := NormalLevel(get(rec, "level"))
		if level == "" {
			level = "sedang"
		}
		if !ValidLevel(level) {
			fail(baris, fmt.Sprintf("level %q harus mudah / sedang / sulit", level))
			continue
		}
		jawaban, kunci, err := buildCSVSoal(tipe, rec, get, kunciStr, stimulus)
		if err != nil {
			fail(baris, err.Error())
			continue
		}
		if err := ValidateSoal(tipe, jawaban, kunci); err != nil {
			fail(baris, err.Error())
			continue
		}
		if tipe == "teks_rumpang" {
			var k map[string]any
			_ = json.Unmarshal(kunci, &k)
			if ks, ok := asStringSlice(k["jawaban"]); ok && strings.Count(pertanyaan, "___") != len(ks) {
				fail(baris, "jumlah rumpang ___ harus sama dengan jumlah jawaban kunci (pisah dengan |)")
				continue
			}
		}
		if tipe == "studi_kasus" && stimulus == "" {
			fail(baris, "stimulus wajib diisi untuk tipe studi_kasus")
			continue
		}
		out = append(out, csvSoal{
			baris: baris, tipe: tipe, pertanyaan: pertanyaan, stimulus: stimulus,
			jawaban: jawaban, kunci: kunci, pembahasan: strings.TrimSpace(get(rec, "pembahasan")),
			bobot: bobot, level: level,
		})
	}
	return out, errors
}

func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func opsiFromCSV(get func([]string, string) string, rec []string) []map[string]any {
	labels := []string{"A", "B", "C", "D", "E"}
	var out []map[string]any
	for i, lbl := range labels {
		teks := strings.TrimSpace(get(rec, "opsi_"+strings.ToLower(lbl)))
		_ = i
		if teks != "" {
			out = append(out, map[string]any{"label": lbl, "teks": teks})
		}
	}
	return out
}

func splitAndTrim(s, sep string) []string {
	var out []string
	for _, part := range strings.Split(s, sep) {
		if t := strings.TrimSpace(part); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// buildCSVSoal membangun jawaban/kunci JSON dari satu baris CSV.
func buildCSVSoal(tipe string, rec []string, get func([]string, string) string, kunciStr, stimulus string) (json.RawMessage, json.RawMessage, error) {
	switch tipe {
	case "pg", "studi_kasus", "pilihan_gambar":
		opsi := opsiFromCSV(get, rec)
		if len(opsi) < 2 {
			return nil, nil, fmt.Errorf("opsi_a..opsi_e: minimal 2 opsi terisi")
		}
		k := strings.ToUpper(kunciStr)
		found := false
		for _, o := range opsi {
			if o["label"] == k {
				found = true
			}
		}
		if !found {
			return nil, nil, fmt.Errorf("kunci %q tidak cocok dengan label opsi yang terisi", kunciStr)
		}
		return mustJSON(map[string]any{"opsi": opsi}),
			mustJSON(map[string]any{"opsi": k}), nil
	case "pg_kompleks", "pg_multi":
		opsi := opsiFromCSV(get, rec)
		if len(opsi) < 2 {
			return nil, nil, fmt.Errorf("opsi_a..opsi_e: minimal 2 opsi terisi")
		}
		ks := splitAndTrim(strings.ToUpper(kunciStr), ",")
		labels := map[string]bool{}
		for _, o := range opsi {
			labels[o["label"].(string)] = true
		}
		for _, k := range ks {
			if !labels[k] {
				return nil, nil, fmt.Errorf("kunci %q tidak cocok dengan label opsi yang terisi", k)
			}
		}
		if len(ks) == 0 {
			return nil, nil, fmt.Errorf("kunci wajib minimal satu label opsi")
		}
		return mustJSON(map[string]any{"opsi": opsi}),
			mustJSON(map[string]any{"opsi": ks}), nil
	case "benar_salah":
		opsi := opsiFromCSV(get, rec)
		if len(opsi) < 2 {
			return nil, nil, fmt.Errorf("opsi_a..opsi_e: isi pernyataan (minimal 2)")
		}
		pernyataan := make([]string, 0, len(opsi))
		for _, o := range opsi {
			pernyataan = append(pernyataan, o["teks"].(string))
		}
		parts := splitAndTrim(kunciStr, ",")
		if len(parts) != len(pernyataan) {
			return nil, nil, fmt.Errorf("kunci harus %d nilai B/S, dipisah koma", len(pernyataan))
		}
		var nilai []bool
		for _, p := range parts {
			switch strings.ToLower(p) {
			case "b", "benar", "true", "1":
				nilai = append(nilai, true)
			case "s", "salah", "false", "0":
				nilai = append(nilai, false)
			default:
				return nil, nil, fmt.Errorf("kunci %q harus B atau S", p)
			}
		}
		return mustJSON(map[string]any{"pernyataan": pernyataan}),
			mustJSON(map[string]any{"nilai": nilai}), nil
	case "menjodohkan", "peta_konsep":
		opsi := opsiFromCSV(get, rec)
		if len(opsi) < 2 {
			return nil, nil, fmt.Errorf("opsi_a..opsi_e: isi pasangan format kiri::kanan (minimal 2)")
		}
		var pas []map[string]any
		for _, o := range opsi {
			kv := strings.SplitN(o["teks"].(string), "::", 2)
			if len(kv) != 2 || strings.TrimSpace(kv[0]) == "" || strings.TrimSpace(kv[1]) == "" {
				return nil, nil, fmt.Errorf("opsi %s harus format kiri::kanan", o["label"])
			}
			pas = append(pas, map[string]any{"kiri": strings.TrimSpace(kv[0]), "kanan": strings.TrimSpace(kv[1])})
		}
		var kPas []map[string]any
		for _, part := range strings.Split(kunciStr, ";") {
			kv := strings.SplitN(strings.TrimSpace(part), "::", 2)
			if len(kv) != 2 || strings.TrimSpace(kv[0]) == "" || strings.TrimSpace(kv[1]) == "" {
				return nil, nil, fmt.Errorf("kunci harus pasangan format kiri::kanan dipisah ;")
			}
			kPas = append(kPas, map[string]any{"kiri": strings.TrimSpace(kv[0]), "kanan": strings.TrimSpace(kv[1])})
		}
		return mustJSON(map[string]any{"pasangan": pas}),
			mustJSON(map[string]any{"pasangan": kPas}), nil
	case "isian":
		alts := splitAndTrim(kunciStr, "|")
		if len(alts) == 0 {
			return nil, nil, fmt.Errorf("kunci isian wajib diisi (alternatif dipisah |)")
		}
		return mustJSON(map[string]any{}),
			mustJSON(map[string]any{"jawaban": alts}), nil
	case "uraian", "upload_file", "praktik_singkat":
		// Kolom kunci CSV = rubrik.
		return mustJSON(map[string]any{}),
			mustJSON(map[string]any{"rubrik": kunciStr}), nil
	case "urutkan":
		opsi := opsiFromCSV(get, rec)
		if len(opsi) < 2 {
			return nil, nil, fmt.Errorf("opsi_a..opsi_e: isi butir yang diurutkan (minimal 2)")
		}
		items := make([]string, 0, len(opsi))
		labelToIdx := map[string]int{}
		for i, o := range opsi {
			items = append(items, o["teks"].(string))
			labelToIdx[o["label"].(string)] = i
		}
		ks := splitAndTrim(strings.ToUpper(kunciStr), ",")
		if len(ks) != len(items) {
			return nil, nil, fmt.Errorf("kunci harus %d label opsi (urutan benar), dipisah koma", len(items))
		}
		var urutan []float64
		seen := map[string]bool{}
		for _, k := range ks {
			idx, ok := labelToIdx[k]
			if !ok {
				return nil, nil, fmt.Errorf("kunci %q tidak cocok dengan label opsi yang terisi", k)
			}
			if seen[k] {
				return nil, nil, fmt.Errorf("kunci berisi label %q duplikat", k)
			}
			seen[k] = true
			urutan = append(urutan, float64(idx))
		}
		return mustJSON(map[string]any{"item": items}),
			mustJSON(map[string]any{"urutan": urutan}), nil
	case "isi_ganda":
		alts := splitAndTrim(kunciStr, "|")
		if len(alts) == 0 {
			return nil, nil, fmt.Errorf("kunci wajib jawaban per kolom, dipisah |")
		}
		kolom := make([]string, 0, len(alts))
		for i := range alts {
			kolom = append(kolom, fmt.Sprintf("Kolom %d", i+1))
		}
		return mustJSON(map[string]any{"kolom": kolom}),
			mustJSON(map[string]any{"jawaban": alts}), nil
	case "teks_rumpang":
		alts := splitAndTrim(kunciStr, "|")
		if len(alts) == 0 {
			return nil, nil, fmt.Errorf("kunci wajib jawaban per rumpang, dipisah |")
		}
		return mustJSON(map[string]any{}),
			mustJSON(map[string]any{"jawaban": alts}), nil
	case "pernyataan_sikap":
		opsi := opsiFromCSV(get, rec)
		if len(opsi) < 2 {
			return nil, nil, fmt.Errorf("opsi_a..opsi_e: isi pernyataan sikap (minimal 2)")
		}
		pernyataan := make([]string, 0, len(opsi))
		for _, o := range opsi {
			pernyataan = append(pernyataan, o["teks"].(string))
		}
		skala := []float64{1, 2, 3, 4}
		return mustJSON(map[string]any{"pernyataan": pernyataan, "skala": skala}),
			mustJSON(map[string]any{"skala": skala}), nil
	}
	return nil, nil, fmt.Errorf("tipe %q belum didukung import CSV", tipe)
}
