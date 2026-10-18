// Package grading — penskoran murni (tanpa DB, tanpa HTTP) untuk mesin ujian.
//
// Kontrak (tugas Tim 4):
//
//	func HitungPoin(tipe string, jawaban, kunci jsonb, bobot int) (poin float64, auto bool)
//	func HitungNilaiAkhir(jawabans) (nilai, menungguKoreksi)
//
// Aturan penilaian (FINAL, dari tugas Tim 4 + ANALISIS Bab 6):
//   - PG: 1/0 (benar = bobot penuh, salah/kosong = 0, tanpa penalti).
//   - PG Kompleks: proporsional (benar_dipilih/total_kunci × bobot, tanpa penalti).
//   - Menjodohkan: proporsional per pasangan benar.
//   - Isian: exact-match setelah normalisasi (huruf kecil, spasi diringkas).
//   - Uraian/manual: TIDAK auto (auto=false, poin 0 sementara, menunggu koreksi).
//   - Varian Tim 3: benar-salah proporsional per pernyataan;
//     mengurutkan proporsional per posisi benar; setuju-tidak-setuju
//     seperti benar-salah bila punya kunci; skala sikap TIDAK dinilai.
//   - Nilai akhir = Σ poin / Σ bobot × 100 (soal yang tidak dinilai
//     tidak masuk pembilang maupun penyebut).
//
// Format jawaban & kunci (jsonb) — toleran terhadap beberapa bentuk:
//   - PG: jawaban "B" atau {"pilihan":"B"}; kunci "B" atau {"jawaban":"B"}.
//   - PG Kompleks: jawaban ["A","C"]; kunci ["A","C"].
//   - Benar-Salah: jawaban {"p1":true,"p2":false}; kunci {"p1":true,"p2":false}.
//   - Menjodohkan: jawaban {"kiri1":"kanan2"}; kunci {"kiri1":"kanan2"}.
//   - Isian: jawaban "Jakarta"; kunci "Jakarta" atau ["Jakarta","DKI Jakarta"].
//   - Mengurutkan: jawaban ["s1","s2","s3"]; kunci ["s1","s2","s3"].
//   - Kunci kosong/nil = soal tidak punya kunci terstruktur:
//     untuk tipe auto -> poin 0 (bukan error); untuk tipe survei/sikap
//     -> tidak dinilai (lihat Dinilai).
package grading

import (
	"encoding/json"
	"math"
	"strings"
)

// JawabanNilai adalah satu baris jawaban untuk HitungNilaiAkhir.
type JawabanNilai struct {
	// Poin yang diperoleh untuk soal ini.
	Poin float64
	// Bobot soal ini (penyebut).
	Bobot float64
	// MenungguKoreksi true bila soal manual belum dikoreksi.
	MenungguKoreksi bool
	// Dinilai false bila soal tidak masuk nilai (mis. skala sikap,
	// studi kasus sebagai wadah). Default: true — pemanggil yang
	// tahu tipe soal mengisinya lewat Dinilai().
	Dinilai bool
}

// NormalisasiTipe menyamakan berbagai penulisan tipe soal Tim 3
// (snake_case, label Indonesia, singkatan) ke kunci kanonis.
func NormalisasiTipe(tipe string) string {
	t := strings.ToLower(strings.TrimSpace(tipe))
	t = strings.ReplaceAll(t, "-", "_")
	t = strings.ReplaceAll(t, " ", "_")
	switch t {
	case "pg", "pilihan_ganda", "pilihan_ganda_tunggal", "pg_tunggal",
		"pg_stimulus_gambar", "pg_berbasis_tabel", "pg_berbasis_grafik",
		"pg_gambar", "pg_tabel", "pg_grafik",
		"pilihan_gambar", "studi_kasus":
		// Tim 3: pilihan_gambar & studi_kasus dinilai seperti PG (1/0).
		return "pg"
	case "pg_kompleks", "pilihan_ganda_kompleks", "pgk", "pg_kompleks_multi_benar",
		"pgk_multi_benar", "pg_kompleks_pernyataan", "mcma",
		"pg_multi":
		// Tim 3: pg_multi = PG multi-jawaban, dinilai seperti PG Kompleks.
		return "pg_kompleks"
	case "benar_salah", "benar_salah_pernyataan", "bs", "true_false", "pgk_benar_salah":
		return "benar_salah"
	case "menjodohkan", "menjodohkan_gambar", "matching", "pasangan",
		"peta_konsep":
		// Tim 3: peta_konsep = pasangan simpul, dinilai seperti menjodohkan.
		return "menjodohkan"
	case "isian_singkat", "isian", "jawaban_singkat", "isian_singkat_jawaban":
		// Tim 3 memakai kode "isian".
		return "isian_singkat"
	case "uraian", "esai", "essay", "uraian_esai", "isian_panjang", "isi_panjang",
		"esai_singkat", "uraian_singkat",
		"upload_file", "praktik_singkat":
		// Tim 3: upload_file & praktik_singkat dinilai manual seperti uraian.
		return "uraian"
	case "mengurutkan", "urutan", "ordering", "sequence", "urutkan":
		// Tim 3 memakai kode "urutkan".
		return "mengurutkan"
	case "isi_ganda":
		// Tim 3: isian multi-kolom, proporsional per kolom.
		return "isi_ganda"
	case "teks_rumpang":
		// Tim 3: isian dalam paragraf, proporsional per rumpang.
		return "teks_rumpang"
	case "setuju_tidak_setuju", "setuju_tidak", "sts", "likert_setuju":
		return "setuju_tidak_setuju"
	case "skala_sikap", "skala", "survei", "survey", "angket", "pernyataan_sikap":
		// Tim 3 memakai kode "pernyataan_sikap".
		return "skala_sikap"
	case "studi_kasus_wadah", "kasus":
		// Wadah studi kasus (bukan soal PG-nya) — tidak dinilai langsung.
		return "studi_kasus"
	default:
		return t
	}
}

// Dinilai melaporkan apakah soal bertipe `tipe` dengan `kunci` masuk
// dalam perhitungan nilai akhir. Skala sikap / survei dan wadah studi
// kasus TIDAK dinilai. Setuju-tidak-setuju hanya dinilai bila punya kunci.
func Dinilai(tipe string, kunci []byte) bool {
	switch NormalisasiTipe(tipe) {
	case "skala_sikap", "studi_kasus":
		return false
	case "setuju_tidak_setuju":
		return len(bytesTrim(kunci)) > 0 && string(bytesTrim(kunci)) != "null" &&
			string(bytesTrim(kunci)) != "{}" && string(bytesTrim(kunci)) != "[]"
	default:
		return true
	}
}

func bytesTrim(b []byte) []byte {
	return []byte(strings.TrimSpace(string(b)))
}

// Manual melaporkan apakah tipe soal dinilai manual oleh guru.
func Manual(tipe string) bool {
	return NormalisasiTipe(tipe) == "uraian"
}

// NormalisasiIsian: huruf kecil, spasi diringkas, trim — untuk
// pencocokan isian singkat (default: tidak case-sensitive).
func NormalisasiIsian(s string) string {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(s)))
	return strings.Join(fields, " ")
}

func parseString(raw []byte) (string, bool) {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s, true
	}
	return "", false
}

func parseStringSlice(raw []byte) ([]string, bool) {
	var arr []string
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr, true
	}
	// Bentuk tunggal ("A") dianggap list satu elemen.
	if s, ok := parseString(raw); ok && s != "" {
		return []string{s}, true
	}
	return nil, false
}

func parseBoolMap(raw []byte) (map[string]bool, bool) {
	var m map[string]bool
	if err := json.Unmarshal(raw, &m); err == nil {
		return m, true
	}
	// Bentuk list boolean [true, false] -> map "0","1",...
	var arr []bool
	if err := json.Unmarshal(raw, &arr); err == nil {
		m = map[string]bool{}
		for i, v := range arr {
			m[itoa(i)] = v
		}
		return m, true
	}
	return nil, false
}

// kunciPNLengkap true bila kunci adalah map warisan {"p1":bool,...,"pN":bool}
// yang lengkap tanpa lubang. Dipakai untuk menyejajarkan jawaban array
// [bool,...] dengan label p1..pN per posisi (p1=[0], p2=[1], ...).
func kunciPNLengkap(m map[string]bool) bool {
	if len(m) == 0 {
		return false
	}
	for i := 1; i <= len(m); i++ {
		if _, ada := m["p"+itoa(i)]; !ada {
			return false
		}
	}
	return true
}

func parseStringMap(raw []byte) (map[string]string, bool) {
	var m map[string]string
	if err := json.Unmarshal(raw, &m); err == nil {
		return m, true
	}
	return nil, false
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [12]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

// kunciString mengambil kunci bentuk string dari berbagai bentuk kunci:
// "B", {"jawaban":"B"}, {"pilihan_benar":"B"}, {"benar":"B"}, {"kunci":"B"}.
func kunciString(kunci []byte) (string, bool) {
	if s, ok := parseString(kunci); ok {
		return s, true
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(kunci, &obj); err != nil {
		return "", false
	}
	for _, key := range []string{"jawaban", "pilihan_benar", "benar", "kunci", "pilihan"} {
		if raw, ada := obj[key]; ada {
			if s, ok := parseString(raw); ok {
				return s, true
			}
		}
	}
	return "", false
}

// jawabanString mengambil jawaban PG dari bentuk "B" / {"pilihan":"B"} /
// {"jawaban":"B"}.
func jawabanString(jawaban []byte) (string, bool) {
	if s, ok := parseString(jawaban); ok {
		return s, true
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(jawaban, &obj); err != nil {
		return "", false
	}
	for _, key := range []string{"pilihan", "jawaban", "pilihan_benar", "kunci"} {
		if raw, ada := obj[key]; ada {
			if s, ok := parseString(raw); ok {
				return s, true
			}
		}
	}
	return "", false
}

// kunciStringSlice mengambil kunci list dari bentuk ["A","C"] atau
// {"jawaban":["A","C"]} / {"kunci":[...]} / {"pilihan_benar":[...]}.
func kunciStringSlice(kunci []byte) ([]string, bool) {
	if arr, ok := parseStringSlice(kunci); ok {
		return arr, true
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(kunci, &obj); err != nil {
		return nil, false
	}
	for _, key := range []string{"jawaban", "kunci", "pilihan_benar", "benar"} {
		if raw, ada := obj[key]; ada {
			if arr, ok := parseStringSlice(raw); ok {
				return arr, true
			}
		}
	}
	return nil, false
}

// --- Helper kunci bentuk Tim 3 (frontend/src/lib/banksoal.ts) ---
// Soal Tim 3: jawaban = data tampil ke peserta, kunci = guru saja.
// Bentuk kunci per tipe:
//   pg/pilihan_gambar/studi_kasus: {"opsi":"B"}
//   pg_kompleks/pg_multi:          {"opsi":["A","C"]}
//   benar_salah:                   {"nilai":[true,false,...]}
//   menjodohkan/peta_konsep:       {"pasangan":[{"kiri","kanan"}]}
//   isian/isi_ganda/teks_rumpang:  {"jawaban":["alt1","alt2",...]}
//   urutkan:                       {"urutan":[indeks,...]} (indeks ke jawaban.item)
//   uraian/upload_file/praktik:    {"rubrik":"..."}
// Semua helper juga menerima bentuk sederhana ("B", ["A","C"], {...})
// agar kompatibel dengan kunci lama.

// kunciObj mengurai kunci menjadi map (bila berbentuk objek).
func kunciObj(kunci []byte) (map[string]json.RawMessage, bool) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(kunci, &obj); err != nil {
		return nil, false
	}
	return obj, true
}

// kunciOpsi mengambil daftar label opsi dari kunci {"opsi": ...} / bentuk sederhana.
func kunciOpsi(kunci []byte) ([]string, bool) {
	if obj, ok := kunciObj(kunci); ok {
		if raw, ada := obj["opsi"]; ada {
			if arr, ok := parseStringSlice(raw); ok {
				return arr, true
			}
		}
	}
	return kunciStringSlice(kunci)
}

// kunciNilaiBool mengambil array boolean dari kunci {"nilai": [...]}.
func kunciNilaiBool(kunci []byte) ([]bool, bool) {
	if obj, ok := kunciObj(kunci); ok {
		if raw, ada := obj["nilai"]; ada {
			var arr []bool
			if err := json.Unmarshal(raw, &arr); err == nil {
				return arr, true
			}
		}
	}
	var arr []bool
	if err := json.Unmarshal(kunci, &arr); err == nil {
		return arr, true
	}
	return nil, false
}

// kunciPasangan mengambil peta kiri->kanan dari kunci
// {"pasangan":[{"kiri","kanan"}]} / {"kiri":"kanan"}.
func kunciPasangan(kunci []byte) (map[string]string, bool) {
	if m, ok := parseStringMap(kunci); ok && len(m) > 0 {
		// Pastikan bukan objek kunci Tim 3 (punya field "pasangan").
		if _, adaPasangan := m["pasangan"]; !adaPasangan {
			return m, true
		}
	}
	if obj, ok := kunciObj(kunci); ok {
		if raw, ada := obj["pasangan"]; ada {
			var arr []struct {
				Kiri  string `json:"kiri"`
				Kanan string `json:"kanan"`
			}
			if err := json.Unmarshal(raw, &arr); err == nil && len(arr) > 0 {
				m := map[string]string{}
				for _, p := range arr {
					m[p.Kiri] = p.Kanan
				}
				return m, true
			}
		}
	}
	return nil, false
}

// jawabanPasangan menormalkan jawaban menjodohkan menjadi peta kiri->kanan.
// Menerima map {"kiri":"kanan"} atau array [{"kiri","kanan"}].
func jawabanPasangan(jawaban []byte) (map[string]string, bool) {
	if m, ok := parseStringMap(jawaban); ok && len(m) > 0 {
		return m, true
	}
	var arr []struct {
		Kiri  string `json:"kiri"`
		Kanan string `json:"kanan"`
	}
	if err := json.Unmarshal(jawaban, &arr); err == nil && len(arr) > 0 {
		m := map[string]string{}
		for _, p := range arr {
			m[p.Kiri] = p.Kanan
		}
		return m, true
	}
	return nil, false
}

// kunciJawabanList mengambil daftar jawaban dari kunci {"jawaban": [...]}.
func kunciJawabanList(kunci []byte) ([]string, bool) {
	if obj, ok := kunciObj(kunci); ok {
		if raw, ada := obj["jawaban"]; ada {
			if arr, ok := parseStringSlice(raw); ok {
				return arr, true
			}
		}
	}
	return kunciStringSlice(kunci)
}

// parseFloatSlice mengurai array angka JSON (untuk jawaban urutkan
// berbentuk indeks).
func parseFloatSlice(raw []byte) ([]float64, bool) {
	var arr []float64
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr, true
	}
	return nil, false
}

// kunciUrutan mengambil {"urutan":[indeks]} + daftar item (bila
// dikirim exam.go sebagai {"urutan":[...],"item":[...]}) untuk urutkan.
func kunciUrutan(kunci []byte) (urutan []int, item []string, ok bool) {
	if obj, okObj := kunciObj(kunci); okObj {
		if raw, ada := obj["urutan"]; ada {
			var nums []float64
			if err := json.Unmarshal(raw, &nums); err == nil {
				for _, n := range nums {
					urutan = append(urutan, int(n))
				}
			}
		}
		if raw, ada := obj["item"]; ada {
			item, _ = parseStringSlice(raw)
		}
		if raw, ada := obj["items"]; ada && len(item) == 0 {
			item, _ = parseStringSlice(raw)
		}
	}
	if len(urutan) == 0 {
		// Bentuk sederhana: kunci = [indeks, ...].
		if nums, ok2 := parseFloatSlice(kunci); ok2 {
			for _, n := range nums {
				urutan = append(urutan, int(n))
			}
		}
	}
	return urutan, item, len(urutan) > 0
}

// HitungPoin menghitung poin satu soal.
// jawaban & kunci adalah JSONB (nil/kosong = tidak dijawab / tanpa kunci).
// bobot <= 0 diperlakukan sebagai 1 (hindari pembagian nol di pemanggil;
// pemanggil seharusnya selalu mengirim bobot soal yang valid).
// Kembali: (poin, auto). auto=false berarti soal manual — poin 0
// sementara dan menunggu koreksi guru.
func HitungPoin(tipe string, jawaban, kunci []byte, bobot int) (float64, bool) {
	b := float64(bobot)
	if b <= 0 {
		b = 1
	}
	if len(bytesTrim(jawaban)) == 0 || string(bytesTrim(jawaban)) == "null" {
		// Soal kosong = 0, dicatat sebagai "kosong" (bukan "salah").
		// Untuk soal manual, kosong tetap menunggu koreksi? Tidak:
		// jawaban kosong tidak bisa dikoreksi -> auto, poin 0.
		return 0, true
	}

	switch NormalisasiTipe(tipe) {
	case "uraian":
		// Manual 0–100 × bobot (skor diisi guru lewat koreksi).
		return 0, false

	case "skala_sikap", "studi_kasus":
		// Tidak dinilai.
		return 0, true

	case "pg":
		// Tim 3: kunci {"opsi":"B"}; jawaban "B".
		opsi, ok := kunciOpsi(kunci)
		if !ok || len(opsi) == 0 {
			return 0, true
		}
		jwb, ok := jawabanString(jawaban)
		if !ok {
			return 0, true
		}
		if jwb == opsi[0] {
			return b, true
		}
		return 0, true

	case "pg_kompleks":
		// Tim 3: kunci {"opsi":["A","C"]}; jawaban ["A","C"].
		kunciArr, ok := kunciOpsi(kunci)
		if !ok || len(kunciArr) == 0 {
			return 0, true
		}
		jwbArr, ok := parseStringSlice(jawaban)
		if !ok {
			return 0, true
		}
		benarDipilih := 0
		kunciSet := map[string]bool{}
		for _, k := range kunciArr {
			kunciSet[k] = true
		}
		for _, j := range jwbArr {
			if kunciSet[j] {
				benarDipilih++
			}
		}
		// Proporsional, tanpa penalti jawaban salah.
		return b * float64(benarDipilih) / float64(len(kunciArr)), true

	case "benar_salah", "setuju_tidak_setuju":
		// Tim 3: kunci {"nilai":[bool,...]}; jawaban [bool,...] sejajar
		// pernyataan. Bentuk map {"p1":bool} juga didukung.
		if kunciArr, ok := kunciNilaiBool(kunci); ok && len(kunciArr) > 0 {
			if jwbArr, ok := func() ([]bool, bool) {
				var arr []bool
				if err := json.Unmarshal(jawaban, &arr); err == nil {
					return arr, true
				}
				return nil, false
			}(); ok {
				benar := 0
				for i, v := range kunciArr {
					if i < len(jwbArr) && jwbArr[i] == v {
						benar++
					}
				}
				return b * float64(benar) / float64(len(kunciArr)), true
			}
			return 0, true
		}
		kunciMap, ok := parseBoolMap(kunci)
		if !ok || len(kunciMap) == 0 {
			return 0, true
		}
		// Kunci warisan {"p1":..,"p2":..} + jawaban array [..]: sejajarkan
		// per posisi (p1=[0], p2=[1], ...). Tanpa cabang ini nilai selalu 0
		// karena parseBoolMap menormalisasi array jawaban menjadi
		// {"0","1",...} yang tidak pernah cocok dengan label p1..pN.
		if kunciPNLengkap(kunciMap) {
			var jwbArr []bool
			if err := json.Unmarshal(jawaban, &jwbArr); err == nil && len(jwbArr) > 0 {
				benar := 0
				for i := 1; i <= len(kunciMap); i++ {
					if i-1 < len(jwbArr) && jwbArr[i-1] == kunciMap["p"+itoa(i)] {
						benar++
					}
				}
				return b * float64(benar) / float64(len(kunciMap)), true
			}
		}
		jwbMap, ok := parseBoolMap(jawaban)
		if !ok {
			return 0, true
		}
		benar := 0
		for k, v := range kunciMap {
			if jwb, ada := jwbMap[k]; ada && jwb == v {
				benar++
			}
		}
		return b * float64(benar) / float64(len(kunciMap)), true

	case "menjodohkan":
		// Tim 3: kunci {"pasangan":[{"kiri","kanan"}]};
		// jawaban map {"kiri":"kanan"} / array [{"kiri","kanan"}].
		kunciMap, ok := kunciPasangan(kunci)
		if !ok || len(kunciMap) == 0 {
			return 0, true
		}
		jwbMap, ok := jawabanPasangan(jawaban)
		if !ok {
			return 0, true
		}
		benar := 0
		for k, v := range kunciMap {
			if jwb, ada := jwbMap[k]; ada && jwb == v {
				benar++
			}
		}
		return b * float64(benar) / float64(len(kunciMap)), true

	case "isian_singkat":
		jwb, ok := jawabanString(jawaban)
		if !ok {
			// Jawaban bisa tersimpan sebagai teks mentah tanpa kutip JSON.
			jwb = strings.TrimSpace(string(jawaban))
		}
		jwbNorm := NormalisasiIsian(jwb)
		if jwbNorm == "" {
			return 0, true
		}
		// Kunci: string tunggal, {"jawaban": [...]}, atau daftar alternatif.
		if arr, ok := kunciJawabanList(kunci); ok {
			for _, alt := range arr {
				if NormalisasiIsian(alt) == jwbNorm {
					return b, true
				}
			}
			return 0, true
		}
		if ks, ok := kunciString(kunci); ok {
			if NormalisasiIsian(ks) == jwbNorm {
				return b, true
			}
		}
		return 0, true

	case "isi_ganda", "teks_rumpang":
		// Tim 3: kunci {"jawaban":["kolom1","kolom2",...]}; jawaban
		// ["isi1","isi2",...] sejajar posisi. Proporsional per posisi.
		kunciArr, ok := kunciJawabanList(kunci)
		if !ok || len(kunciArr) == 0 {
			return 0, true
		}
		jwbArr, ok := parseStringSlice(jawaban)
		if !ok {
			return 0, true
		}
		benar := 0
		for i, k := range kunciArr {
			if i < len(jwbArr) && NormalisasiIsian(jwbArr[i]) == NormalisasiIsian(k) {
				benar++
			}
		}
		return b * float64(benar) / float64(len(kunciArr)), true

	case "mengurutkan":
		// Tim 3: kunci {"urutan":[indeks,...]} (indeks ke jawaban.item).
		// Jawaban: array indeks [2,0,1] (dikirim frontend) ATAU array
		// string item dalam urutan siswa (bila exam.go memperkaya kunci
		// dengan "item"). Proporsional per posisi benar.
		urutan, item, ok := kunciUrutan(kunci)
		if !ok || len(urutan) == 0 {
			// Fallback lama: kunci = urutan item yang benar (string).
			kunciArr, ok2 := kunciStringSlice(kunci)
			if !ok2 || len(kunciArr) == 0 {
				return 0, true
			}
			jwbArr, ok2 := parseStringSlice(jawaban)
			if !ok2 {
				return 0, true
			}
			benarPosisi := 0
			for i, k := range kunciArr {
				if i < len(jwbArr) && jwbArr[i] == k {
					benarPosisi++
				}
			}
			return b * float64(benarPosisi) / float64(len(kunciArr)), true
		}
		if idxArr, okIdx := parseFloatSlice(jawaban); okIdx {
			benarPosisi := 0
			for i, u := range urutan {
				if i < len(idxArr) && int(idxArr[i]) == u {
					benarPosisi++
				}
			}
			return b * float64(benarPosisi) / float64(len(urutan)), true
		}
		if len(item) > 0 {
			// Bangun urutan benar sebagai string, bandingkan posisi.
			benar := make([]string, 0, len(urutan))
			for _, u := range urutan {
				if u >= 0 && u < len(item) {
					benar = append(benar, item[u])
				}
			}
			jwbArr, okStr := parseStringSlice(jawaban)
			if !okStr {
				return 0, true
			}
			benarPosisi := 0
			for i, k := range benar {
				if i < len(jwbArr) && jwbArr[i] == k {
					benarPosisi++
				}
			}
			return b * float64(benarPosisi) / float64(len(benar)), true
		}
		return 0, true

	default:
		// Tipe tak dikenal: jangan memberi nilai palsu.
		// Perlakukan sebagai manual agar guru memeriksa (aman).
		return 0, false
	}
}

// HitungNilaiAkhir menghitung nilai akhir 0–100 dari daftar jawaban.
// nilai = Σ poin / Σ bobot × 100 (hanya soal yang Dinilai=true).
// menungguKoreksi = ada soal manual yang belum dikoreksi.
// Bila tidak ada soal yang dinilai, nilai = 0 dan menungguKoreksi
// mengikuti input (tidak ada yang menunggu bila list kosong).
func HitungNilaiAkhir(jawabans []JawabanNilai) (nilai float64, menungguKoreksi bool) {
	var totalPoin, totalBobot float64
	for _, j := range jawabans {
		if !j.Dinilai {
			continue
		}
		totalPoin += j.Poin
		totalBobot += j.Bobot
		if j.MenungguKoreksi {
			menungguKoreksi = true
		}
	}
	if totalBobot <= 0 {
		return 0, menungguKoreksi
	}
	nilai = totalPoin / totalBobot * 100
	// Jepit ke [0,100] untuk keamanan numerik.
	nilai = math.Max(0, math.Min(100, nilai))
	return nilai, menungguKoreksi
}
