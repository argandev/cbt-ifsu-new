// Package banksoal — bank soal (Tim 3).
//
// Label JUJUR: v2 mendukung "5 tipe standar AKM + 11 varian/ekstensi
// aplikasi". TIDAK PERNAH mengklaim "16 tipe standar Kemendikbud" —
// klaim tersebut tidak ditemukan di dokumen resmi Kemendikbud
// (Framework AKM, BSKAP 2023, hanya menetapkan 5 bentuk soal).
// Rincian tiap tipe: docs/TIPE-SOAL.md.
//
// Keamanan kunci: kolom `kunci` TIDAK PERNAH dikirim ke frontend siswa.
// Endpoint guru BOLEH menampilkan kunci. Proktor TIDAK boleh melihat
// isi soal sama sekali (README). Penegakan di server, bukan hanya
// dengan menyembunyikan tombol di frontend.
//
// Auth: identitas dibaca dari context yang diisi middleware auth Tim 1
// (lihat audit.CurrentUser: auth_user_id, auth_role, auth_permissions).
// Permission yang dipakai: banksoal.read / .create / .update / .delete /
// .lock / .import (seed Tim 6). Role admin lolos semua permission.
// CSRF ditegakkan middleware global Tim 1 (semua mutasi wajib token CSRF).
package banksoal

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Kelompok tipe soal: "inti" = bentuk resmi AKM (Framework AKM 2023),
// "varian" = varian/ekstensi aplikasi, BUKAN standar Kemendikbud.
const (
	KelompokInti   = "inti"
	KelompokVarian = "varian"
)

// Penilaian: "auto" = dinilai otomatis backend; "manual" = dikoreksi
// guru/penguji; "non-nilai" = survei/refleksi, tidak masuk nilai akhir.
const (
	NilaiAuto     = "auto"
	NilaiManual   = "manual"
	NilaiNonNilai = "non-nilai"
)

// TipeInfo mendeskripsikan satu tipe soal.
type TipeInfo struct {
	Kode      string `json:"kode"`
	Label     string `json:"label"`
	Kelompok  string `json:"kelompok"`  // inti | varian
	Penilaian string `json:"penilaian"` // auto | manual | non-nilai
	Dinilai   bool   `json:"dinilai"`   // false => tidak masuk nilai akhir
	Deskripsi string `json:"deskripsi"`
}

// TipeSoal adalah sumber kebenaran metadata 16 tipe soal. Frontend
// mengimpor cerminan tabel ini dari frontend/src/lib/tipeSoal.ts —
// bila menambah/mengubah tipe, ubah KEDUA tempat + docs/TIPE-SOAL.md.
var TipeSoal = map[string]TipeInfo{
	// ---- 5 INTI AKM (Framework AKM, BSKAP Kemendikbudristek 2023) ----
	"pg": {
		Kode: "pg", Label: "Pilihan Ganda", Kelompok: KelompokInti,
		Penilaian: NilaiAuto, Dinilai: true,
		Deskripsi: "Pilih 1 dari beberapa opsi (A–E). Nilai otomatis 1/0.",
	},
	"pg_kompleks": {
		Kode: "pg_kompleks", Label: "PG Kompleks", Kelompok: KelompokInti,
		Penilaian: NilaiAuto, Dinilai: true,
		Deskripsi: "Pilih semua opsi yang benar. Nilai otomatis PROPORSIONAL = benar_dipilih/total_kunci × bobot, tanpa penalti.",
	},
	"menjodohkan": {
		Kode: "menjodohkan", Label: "Menjodohkan", Kelompok: KelompokInti,
		Penilaian: NilaiAuto, Dinilai: true,
		Deskripsi: "Pasangkan premis (kiri) dengan respons (kanan). Nilai otomatis PROPORSIONAL per pasangan benar.",
	},
	"isian": {
		Kode: "isian", Label: "Isian / Jawaban Singkat", Kelompok: KelompokInti,
		Penilaian: NilaiAuto, Dinilai: true,
		Deskripsi: "Ketik jawaban singkat. Nilai otomatis exact-match setelah normalisasi (trim + lowercase + rapatkan spasi).",
	},
	"uraian": {
		Kode: "uraian", Label: "Uraian / Esai", Kelompok: KelompokInti,
		Penilaian: NilaiManual, Dinilai: true,
		Deskripsi: "Jawaban bebas terstruktur. Nilai MANUAL 0–100 × bobot, dikoreksi guru, disembunyikan sampai final.",
	},
	// ---- 11 VARIAN / EKSTENSI APLIKASI (bukan standar Kemendikbud) ----
	"benar_salah": {
		Kode: "benar_salah", Label: "Benar-Salah", Kelompok: KelompokVarian,
		Penilaian: NilaiAuto, Dinilai: true,
		Deskripsi: "Tentukan Benar/Salah per pernyataan. Nilai otomatis PROPORSIONAL per pernyataan benar.",
	},
	"pg_multi": {
		Kode: "pg_multi", Label: "PG Multi-Jawaban", Kelompok: KelompokVarian,
		Penilaian: NilaiAuto, Dinilai: true,
		Deskripsi: "Seperti PG Kompleks tapi opsi bisa >1 dan label beda. Nilai otomatis PROPORSIONAL, tanpa penalti.",
	},
	"urutkan": {
		Kode: "urutkan", Label: "Mengurutkan", Kelompok: KelompokVarian,
		Penilaian: NilaiAuto, Dinilai: true,
		Deskripsi: "Urutkan langkah/peristiwa. Nilai otomatis PROPORSIONAL per posisi benar.",
	},
	"isi_ganda": {
		Kode: "isi_ganda", Label: "Isian Multi-Kolom", Kelompok: KelompokVarian,
		Penilaian: NilaiAuto, Dinilai: true,
		Deskripsi: "Beberapa kolom isian dalam satu soal. Nilai otomatis PROPORSIONAL per kolom benar.",
	},
	"teks_rumpang": {
		Kode: "teks_rumpang", Label: "Teks Rumpang", Kelompok: KelompokVarian,
		Penilaian: NilaiAuto, Dinilai: true,
		Deskripsi: "Isian dalam paragraf (tanda ___). Nilai otomatis PROPORSIONAL per rumpang benar.",
	},
	"peta_konsep": {
		Kode: "peta_konsep", Label: "Peta Konsep", Kelompok: KelompokVarian,
		Penilaian: NilaiAuto, Dinilai: true,
		Deskripsi: "Pasangkan simpul-simpul konsep. Nilai otomatis PROPORSIONAL per pasangan benar.",
	},
	"studi_kasus": {
		Kode: "studi_kasus", Label: "Studi Kasus + PG", Kelompok: KelompokVarian,
		Penilaian: NilaiAuto, Dinilai: true,
		Deskripsi: "Stimulus panjang (kasus) + soal PG. Nilai otomatis 1/0. Stimulus wajib diisi.",
	},
	"pernyataan_sikap": {
		Kode: "pernyataan_sikap", Label: "Skala Sikap", Kelompok: KelompokVarian,
		Penilaian: NilaiNonNilai, Dinilai: false,
		Deskripsi: "Skala 1–4 per pernyataan. NON-NILAI — hanya survei/refleksi, tidak masuk nilai akhir. Tulis jelas di UI: tidak dinilai.",
	},
	"upload_file": {
		Kode: "upload_file", Label: "Upload Berkas", Kelompok: KelompokVarian,
		Penilaian: NilaiManual, Dinilai: true,
		Deskripsi: "Jawaban berupa berkas yang diunggah. Nilai MANUAL oleh guru/penguji.",
	},
	"praktik_singkat": {
		Kode: "praktik_singkat", Label: "Praktik Singkat", Kelompok: KelompokVarian,
		Penilaian: NilaiManual, Dinilai: true,
		Deskripsi: "Jawaban teks + berkas opsional. Nilai MANUAL dengan rubrik sederhana.",
	},
	"pilihan_gambar": {
		Kode: "pilihan_gambar", Label: "Pilihan Gambar", Kelompok: KelompokVarian,
		Penilaian: NilaiAuto, Dinilai: true,
		Deskripsi: "Opsi berupa gambar. Nilai otomatis 1/0.",
	},
}

// ValidTipe melaporkan apakah kode tipe dikenal.
func ValidTipe(tipe string) bool {
	_, ok := TipeSoal[tipe]
	return ok
}

// ValidLevel melaporkan apakah level valid (mudah/sedang/sulit).
func ValidLevel(level string) bool {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "mudah", "sedang", "sulit":
		return true
	}
	return false
}

// NormalLevel menormalkan level ke huruf kecil ("Sedang" -> "sedang").
func NormalLevel(level string) string {
	return strings.ToLower(strings.TrimSpace(level))
}

// NormalisasiIsian menormalkan jawaban isian: trim, lowercase,
// rapatkan spasi berlebih. Dipakai isian, isi_ganda, teks_rumpang.
func NormalisasiIsian(s string) string {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(s)))
	return strings.Join(fields, " ")
}

// --- Bentuk jawaban/kunci per tipe (dipakai validasi + penilaian) ---

type opsi struct {
	Label string `json:"label"`
	Teks  string `json:"teks"`
}

type pasangan struct {
	Kiri  string `json:"kiri"`
	Kanan string `json:"kanan"`
}

// decode mengubah json.RawMessage menjadi map untuk validasi.
func decode(raw json.RawMessage) (map[string]any, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("jawaban/kunci kosong")
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("jawaban/kunci bukan objek JSON: %v", err)
	}
	if m == nil {
		m = map[string]any{}
	}
	return m, nil
}

func asStringSlice(v any) ([]string, bool) {
	arr, ok := v.([]any)
	if !ok {
		return nil, false
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		s, ok := item.(string)
		if !ok {
			return nil, false
		}
		out = append(out, s)
	}
	return out, true
}

func asBoolSlice(v any) ([]bool, bool) {
	arr, ok := v.([]any)
	if !ok {
		return nil, false
	}
	out := make([]bool, 0, len(arr))
	for _, item := range arr {
		b, ok := item.(bool)
		if !ok {
			return nil, false
		}
		out = append(out, b)
	}
	return out, true
}

func asFloatSlice(v any) ([]float64, bool) {
	arr, ok := v.([]any)
	if !ok {
		return nil, false
	}
	out := make([]float64, 0, len(arr))
	for _, item := range arr {
		f, ok := item.(float64)
		if !ok {
			return nil, false
		}
		out = append(out, f)
	}
	return out, true
}

// opsiList mengambil jawaban.opsi sebagai []opsi.
func opsiList(jawaban map[string]any) ([]opsi, error) {
	raw, ok := jawaban["opsi"]
	if !ok {
		return nil, fmt.Errorf("jawaban.opsi wajib ada")
	}
	arr, ok := raw.([]any)
	if !ok || len(arr) == 0 {
		return nil, fmt.Errorf("jawaban.opsi wajib array tidak kosong")
	}
	out := make([]opsi, 0, len(arr))
	seen := map[string]bool{}
	for i, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("jawaban.opsi[%d] harus objek {label, teks}", i)
		}
		label, _ := obj["label"].(string)
		teks, _ := obj["teks"].(string)
		if strings.TrimSpace(label) == "" {
			return nil, fmt.Errorf("jawaban.opsi[%d].label wajib diisi", i)
		}
		if seen[label] {
			return nil, fmt.Errorf("jawaban.opsi label %q duplikat", label)
		}
		seen[label] = true
		out = append(out, opsi{Label: label, Teks: teks})
	}
	return out, nil
}

// pasanganList mengambil field pasangan dari map (jawaban/kunci).
func pasanganList(m map[string]any, field string) ([]pasangan, error) {
	raw, ok := m[field]
	if !ok {
		return nil, fmt.Errorf("%s wajib ada", field)
	}
	arr, ok := raw.([]any)
	if !ok || len(arr) == 0 {
		return nil, fmt.Errorf("%s wajib array tidak kosong", field)
	}
	out := make([]pasangan, 0, len(arr))
	for i, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s[%d] harus objek {kiri, kanan}", field, i)
		}
		kiri, _ := obj["kiri"].(string)
		kanan, _ := obj["kanan"].(string)
		if strings.TrimSpace(kiri) == "" || strings.TrimSpace(kanan) == "" {
			return nil, fmt.Errorf("%s[%d].kiri dan .kanan wajib diisi", field, i)
		}
		out = append(out, pasangan{Kiri: kiri, Kanan: kanan})
	}
	return out, nil
}

// ValidateSoal memvalidasi struktur jawaban/kunci untuk satu tipe.
// Dipakai endpoint JSON maupun import CSV (CSV membangun jawaban/kunci
// dulu, lalu divalidasi lewat fungsi ini).
func ValidateSoal(tipe string, jawabanRaw, kunciRaw json.RawMessage) error {
	info, ok := TipeSoal[tipe]
	if !ok {
		return fmt.Errorf("tipe %q tidak dikenal", tipe)
	}
	_ = info
	jawaban, err := decode(jawabanRaw)
	if err != nil {
		return fmt.Errorf("jawaban: %v", err)
	}
	kunci, err := decode(kunciRaw)
	if err != nil {
		return fmt.Errorf("kunci: %v", err)
	}
	switch tipe {
	case "pg", "studi_kasus", "pilihan_gambar":
		opsi, err := opsiList(jawaban)
		if err != nil {
			return err
		}
		if tipe == "studi_kasus" {
			// Stimulus dicek terpisah di handler (perlu field stimulus).
			_ = opsi
		}
		if tipe == "pilihan_gambar" {
			for i, o := range opsi {
				if strings.TrimSpace(o.Teks) == "" {
					return fmt.Errorf("jawaban.opsi[%d].teks (URL gambar) wajib diisi", i)
				}
			}
		}
		k, _ := kunci["opsi"].(string)
		if strings.TrimSpace(k) == "" {
			return fmt.Errorf("kunci.opsi wajib satu label opsi (mis. \"B\")")
		}
		found := false
		for _, o := range opsi {
			if o.Label == k {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("kunci.opsi %q tidak ada di jawaban.opsi", k)
		}
	case "pg_kompleks", "pg_multi":
		opsi, err := opsiList(jawaban)
		if err != nil {
			return err
		}
		ks, ok := asStringSlice(kunci["opsi"])
		if !ok || len(ks) == 0 {
			return fmt.Errorf("kunci.opsi wajib array label tidak kosong")
		}
		labels := map[string]bool{}
		for _, o := range opsi {
			labels[o.Label] = true
		}
		for _, k := range ks {
			if !labels[k] {
				return fmt.Errorf("kunci.opsi %q tidak ada di jawaban.opsi", k)
			}
		}
	case "benar_salah", "pernyataan_sikap":
		pernyataan, ok := asStringSlice(jawaban["pernyataan"])
		if !ok || len(pernyataan) == 0 {
			return fmt.Errorf("jawaban.pernyataan wajib array teks tidak kosong")
		}
		if tipe == "benar_salah" {
			nilai, ok := asBoolSlice(kunci["nilai"])
			if !ok {
				return fmt.Errorf("kunci.nilai wajib array boolean per pernyataan")
			}
			if len(nilai) != len(pernyataan) {
				return fmt.Errorf("kunci.nilai (%d) harus sama banyak dengan pernyataan (%d)", len(nilai), len(pernyataan))
			}
		} else {
			skala, ok := asFloatSlice(kunci["skala"])
			if !ok || len(skala) == 0 {
				return fmt.Errorf("kunci.skala wajib array angka (mis. [1,2,3,4])")
			}
		}
	case "menjodohkan", "peta_konsep":
		jPas, err := pasanganList(jawaban, "pasangan")
		if err != nil {
			return err
		}
		kPas, err := pasanganList(kunci, "pasangan")
		if err != nil {
			return err
		}
		jSet := map[pasangan]bool{}
		for _, p := range jPas {
			jSet[p] = true
		}
		for _, p := range kPas {
			if !jSet[p] {
				return fmt.Errorf("kunci.pasangan {%q -> %q} tidak ada di jawaban.pasangan", p.Kiri, p.Kanan)
			}
		}
	case "isian":
		ks, ok := asStringSlice(kunci["jawaban"])
		if !ok || len(ks) == 0 {
			return fmt.Errorf("kunci.jawaban wajib array teks (satu atau lebih jawaban alternatif)")
		}
		for _, k := range ks {
			if strings.TrimSpace(k) == "" {
				return fmt.Errorf("kunci.jawaban tidak boleh berisi teks kosong")
			}
		}
	case "uraian", "upload_file", "praktik_singkat":
		// Manual: kunci berisi rubrik (boleh kosong untuk draf, tapi
		// sangat disarankan). Rubrik kosong diterima agar guru bisa
		// menyimpan draf lalu melengkapi rubrik sebelum bank dikunci.
		if rubrik, _ := kunci["rubrik"].(string); rubrik != "" {
			_ = rubrik
		}
	case "urutkan":
		items, ok := asStringSlice(jawaban["item"])
		if !ok || len(items) < 2 {
			return fmt.Errorf("jawaban.item wajib array minimal 2 butir")
		}
		urutan, ok := asFloatSlice(kunci["urutan"])
		if !ok || len(urutan) != len(items) {
			return fmt.Errorf("kunci.urutan wajib array indeks sepanjang jawaban.item (%d)", len(items))
		}
		seen := map[int]bool{}
		for _, f := range urutan {
			idx := int(f)
			if float64(idx) != f || idx < 0 || idx >= len(items) {
				return fmt.Errorf("kunci.urutan berisi indeks %v di luar rentang", f)
			}
			if seen[idx] {
				return fmt.Errorf("kunci.urutan berisi indeks %d duplikat", idx)
			}
			seen[idx] = true
		}
	case "isi_ganda":
		kolom, ok := asStringSlice(jawaban["kolom"])
		if !ok || len(kolom) == 0 {
			return fmt.Errorf("jawaban.kolom wajib array label kolom tidak kosong")
		}
		ks, ok := asStringSlice(kunci["jawaban"])
		if !ok || len(ks) != len(kolom) {
			return fmt.Errorf("kunci.jawaban (%d) harus sama banyak dengan jawaban.kolom (%d)", len(ks), len(kolom))
		}
	case "teks_rumpang":
		ks, ok := asStringSlice(kunci["jawaban"])
		if !ok || len(ks) == 0 {
			return fmt.Errorf("kunci.jawaban wajib array jawaban per rumpang")
		}
		// Jumlah rumpang (___) dicek di handler karena butuh pertanyaan.
		_ = ks
	}
	return nil
}

// ScoreAuto menghitung poin otomatis (0..bobot) untuk satu soal.
// Mengembalikan (poin, isAuto, err). isAuto=false untuk tipe manual /
// non-nilai (uraian, upload_file, praktik_singkat, pernyataan_sikap):
// nilai ditentukan guru, bukan fungsi ini. jawabanSiswa adalah objek
// JSON jawaban peserta, bentuknya per tipe (lihat docs/TIPE-SOAL.md).
// Nilai TIDAK PERNAH dihitung frontend — fungsi ini dipakai backend
// (Tim 4/5: penilaian + rekap).
func ScoreAuto(tipe string, kunciRaw json.RawMessage, jawabanSiswaRaw json.RawMessage, bobot int) (float64, bool, error) {
	info, ok := TipeSoal[tipe]
	if !ok {
		return 0, false, fmt.Errorf("tipe %q tidak dikenal", tipe)
	}
	if info.Penilaian != NilaiAuto {
		return 0, false, nil
	}
	if bobot <= 0 {
		return 0, false, fmt.Errorf("bobot harus > 0")
	}
	kunci, err := decode(kunciRaw)
	if err != nil {
		return 0, false, fmt.Errorf("kunci: %v", err)
	}
	siswa, err := decode(jawabanSiswaRaw)
	if err != nil {
		return 0, false, fmt.Errorf("jawaban siswa: %v", err)
	}
	b := float64(bobot)
	switch tipe {
	case "pg", "studi_kasus", "pilihan_gambar":
		k, _ := kunci["opsi"].(string)
		s, _ := siswa["opsi"].(string)
		if s != "" && s == k {
			return b, true, nil
		}
		return 0, true, nil
	case "pg_kompleks", "pg_multi":
		ks, _ := asStringSlice(kunci["opsi"])
		ss, _ := asStringSlice(siswa["opsi"])
		if len(ks) == 0 {
			return 0, false, fmt.Errorf("kunci.opsi kosong")
		}
		kSet := map[string]bool{}
		for _, k := range ks {
			kSet[k] = true
		}
		benar := 0
		for _, s := range ss {
			if kSet[s] {
				benar++
			}
		}
		// Proporsional, TANPA penalti jawaban salah (keputusan kontrak).
		return b * float64(benar) / float64(len(ks)), true, nil
	case "benar_salah":
		ks, _ := asBoolSlice(kunci["nilai"])
		ss, _ := asBoolSlice(siswa["nilai"])
		if len(ks) == 0 {
			return 0, false, fmt.Errorf("kunci.nilai kosong")
		}
		benar := 0
		for i, k := range ks {
			if i < len(ss) && ss[i] == k {
				benar++
			}
		}
		return b * float64(benar) / float64(len(ks)), true, nil
	case "menjodohkan", "peta_konsep":
		kPas, err := pasanganList(kunci, "pasangan")
		if err != nil {
			return 0, false, err
		}
		sPas, _ := pasanganList(siswa, "pasangan")
		sMap := map[string]string{}
		for _, p := range sPas {
			sMap[p.Kiri] = p.Kanan
		}
		benar := 0
		for _, p := range kPas {
			if sMap[p.Kiri] == p.Kanan {
				benar++
			}
		}
		return b * float64(benar) / float64(len(kPas)), true, nil
	case "isian":
		ks, _ := asStringSlice(kunci["jawaban"])
		s, _ := siswa["jawaban"].(string)
		if s == "" {
			s, _ = siswa["teks"].(string)
		}
		norm := NormalisasiIsian(s)
		for _, k := range ks {
			if norm == NormalisasiIsian(k) {
				return b, true, nil
			}
		}
		return 0, true, nil
	case "urutkan":
		ks, _ := asFloatSlice(kunci["urutan"])
		ss, _ := asFloatSlice(siswa["urutan"])
		if len(ks) == 0 {
			return 0, false, fmt.Errorf("kunci.urutan kosong")
		}
		benar := 0
		for i, k := range ks {
			if i < len(ss) && ss[i] == k {
				benar++
			}
		}
		return b * float64(benar) / float64(len(ks)), true, nil
	case "isi_ganda", "teks_rumpang":
		ks, _ := asStringSlice(kunci["jawaban"])
		ss, _ := asStringSlice(siswa["jawaban"])
		if len(ks) == 0 {
			return 0, false, fmt.Errorf("kunci.jawaban kosong")
		}
		benar := 0
		for i, k := range ks {
			if i < len(ss) && NormalisasiIsian(ss[i]) == NormalisasiIsian(k) {
				benar++
			}
		}
		return b * float64(benar) / float64(len(ks)), true, nil
	}
	return 0, false, fmt.Errorf("tipe %q belum punya aturan skor otomatis", tipe)
}
