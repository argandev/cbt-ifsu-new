package analisisx

import (
	"math"
	"reflect"
	"testing"
)

func hampirSama(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestKesukaran(t *testing.T) {
	tests := []struct {
		nama     string
		benar    int
		penjawab int
		wantP    float64
		wantLbl  string
	}{
		{"sedang biasa", 5, 10, 0.5, "Sedang"},
		{"sukar", 2, 10, 0.2, "Sukar"},
		{"mudah", 8, 10, 0.8, "Mudah"},
		// Batas inklusif kategori Sedang: 0,3 dan 0,7 masih Sedang.
		{"batas bawah 0,3", 3, 10, 0.3, "Sedang"},
		{"batas atas 0,7", 7, 10, 0.7, "Sedang"},
		{"semua benar", 10, 10, 1.0, "Mudah"},
		{"tidak ada yang benar", 0, 10, 0.0, "Sukar"},
		{"benar melebihi penjawab dijepit", 15, 10, 1.0, "Mudah"},
		{"benar negatif dijepit", -3, 10, 0.0, "Sukar"},
	}
	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			p, lbl := Kesukaran(tc.benar, tc.penjawab)
			hampirSama(t, p, tc.wantP)
			if lbl != tc.wantLbl {
				t.Fatalf("label = %q, want %q", lbl, tc.wantLbl)
			}
		})
	}
}

func TestKesukaranPenjawabNol(t *testing.T) {
	p, lbl := Kesukaran(0, 0)
	hampirSama(t, p, 0)
	if lbl != "tidak ada data" {
		t.Fatalf("label = %q, want %q", lbl, "tidak ada data")
	}
	p, lbl = Kesukaran(5, -2)
	hampirSama(t, p, 0)
	if lbl != "tidak ada data" {
		t.Fatalf("label = %q, want %q", lbl, "tidak ada data")
	}
}

func TestDayaBeda(t *testing.T) {
	tests := []struct {
		nama             string
		atas, bawah, n   int
		sampelKecil      bool
		wantD            float64
		wantInterpretasi string
		wantCatatan      string
	}{
		{"sangat baik", 9, 1, 10, false, 0.8, "Sangat Baik", ""},
		{"batas sangat baik 0,40", 4, 0, 10, false, 0.4, "Sangat Baik", ""},
		{"baik 0,35", 7, 0, 20, false, 0.35, "Baik", ""},
		{"batas baik 0,30", 3, 0, 10, false, 0.3, "Baik", ""},
		{"cukup", 5, 3, 10, false, 0.2, "Cukup", ""},
		{"batas cukup 0,29", 29, 0, 100, false, 0.29, "Cukup", ""},
		{"buruk nol", 5, 5, 10, false, 0.0, "Buruk", ""},
		{"buruk 0,10", 6, 5, 10, false, 0.1, "Buruk", ""},
		{"sangat buruk negatif", 2, 5, 10, false, -0.3, "Sangat Buruk", "nilai negatif: kunci jawaban dicurigai"},
		{"sangat buruk negatif + sampel kecil", 1, 3, 5, true, -0.4, "Sangat Buruk", "sampel kecil (n < 10): daya beda tidak dapat dipercaya; nilai negatif: kunci jawaban dicurigai"},
		{"sampel kecil tanpa peringatan lain", 4, 1, 5, true, 0.6, "Sangat Baik", "sampel kecil (n < 10): daya beda tidak dapat dipercaya"},
		{"benar negatif dijepit", -2, 0, 10, false, 0.0, "Buruk", ""},
	}
	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			d, interp, catatan := DayaBeda(tc.atas, tc.bawah, tc.n, tc.sampelKecil)
			hampirSama(t, d, tc.wantD)
			if interp != tc.wantInterpretasi {
				t.Fatalf("interpretasi = %q, want %q", interp, tc.wantInterpretasi)
			}
			if catatan != tc.wantCatatan {
				t.Fatalf("catatan = %q, want %q", catatan, tc.wantCatatan)
			}
		})
	}
}

func TestDayaBedaKelompokNol(t *testing.T) {
	d, interp, catatan := DayaBeda(1, 0, 0, false)
	hampirSama(t, d, 0)
	if interp != "tidak dapat dihitung" {
		t.Fatalf("interpretasi = %q", interp)
	}
	if catatan == "" {
		t.Fatal("catatan harus menjelaskan jumlah kelompok tidak valid")
	}
}

func TestBagiKelompokDuaPuluhTujuhPersen(t *testing.T) {
	// 100 peserta nilai 100..1 (sudah menurun): kelompok = 27 + 27.
	nilai := make([]float64, 100)
	for i := range nilai {
		nilai[i] = float64(100 - i)
	}
	atas, bawah := BagiKelompok(nilai)
	if len(atas) != 27 {
		t.Fatalf("len(atas) = %d, want 27", len(atas))
	}
	if len(bawah) != 27 {
		t.Fatalf("len(bawah) = %d, want 27", len(bawah))
	}
	// Indeks 0..26 = nilai tertinggi; indeks 99..73 = nilai terendah
	// (bawah terurut dari yang paling rendah).
	if atas[0] != 0 || atas[26] != 26 {
		t.Fatalf("atas = %v, want indeks 0..26", atas)
	}
	if bawah[0] != 99 || bawah[26] != 73 {
		t.Fatalf("bawah = %v, want indeks 99..73", bawah)
	}

	// n = 10 -> round(2,7) = 3 per kelompok.
	nilai10 := []float64{10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	atas, bawah = BagiKelompok(nilai10)
	if !reflect.DeepEqual(atas, []int{0, 1, 2}) {
		t.Fatalf("atas n=10 = %v, want [0 1 2]", atas)
	}
	if !reflect.DeepEqual(bawah, []int{9, 8, 7}) {
		t.Fatalf("bawah n=10 = %v, want [9 8 7]", bawah)
	}
}

func TestBagiKelompokMedianSplitSampelKecil(t *testing.T) {
	// n = 9 (ganjil): 4 atas + 4 bawah, peserta median (indeks 4) terbuang.
	nilai := []float64{90, 80, 70, 60, 50, 40, 30, 20, 10}
	atas, bawah := BagiKelompok(nilai)
	if !reflect.DeepEqual(atas, []int{0, 1, 2, 3}) {
		t.Fatalf("atas = %v, want [0 1 2 3]", atas)
	}
	if !reflect.DeepEqual(bawah, []int{8, 7, 6, 5}) {
		t.Fatalf("bawah = %v, want [8 7 6 5]", bawah)
	}

	// n = 8 (genap): 4 atas + 4 bawah, tidak ada yang terbuang.
	nilai8 := []float64{80, 70, 60, 50, 40, 30, 20, 10}
	atas, bawah = BagiKelompok(nilai8)
	if !reflect.DeepEqual(atas, []int{0, 1, 2, 3}) {
		t.Fatalf("atas n=8 = %v", atas)
	}
	if !reflect.DeepEqual(bawah, []int{7, 6, 5, 4}) {
		t.Fatalf("bawah n=8 = %v", bawah)
	}

	// Input tidak terurut: indeks harus mengikuti nilai, bukan posisi.
	acak := []float64{10, 90, 30, 80, 50}
	atas, bawah = BagiKelompok(acak)
	// Terurut menurun: 90(idx1), 80(idx3), 50(idx4), 30(idx2), 10(idx0).
	// Median split n=5 -> ukuran 2: atas [1 3], bawah [0 2] (terendah dulu).
	if !reflect.DeepEqual(atas, []int{1, 3}) {
		t.Fatalf("atas acak = %v, want [1 3]", atas)
	}
	if !reflect.DeepEqual(bawah, []int{0, 2}) {
		t.Fatalf("bawah acak = %v, want [0 2]", bawah)
	}
}

func TestBagiKelompokKasusBatas(t *testing.T) {
	atas, bawah := BagiKelompok(nil)
	if len(atas) != 0 || len(bawah) != 0 {
		t.Fatalf("nil -> atas=%v bawah=%v, want kosong", atas, bawah)
	}
	atas, bawah = BagiKelompok([]float64{75})
	if len(atas) != 0 || len(bawah) != 0 {
		t.Fatalf("n=1 -> atas=%v bawah=%v, want kosong", atas, bawah)
	}
	// n = 2: masing-masing 1 peserta.
	atas, bawah = BagiKelompok([]float64{40, 90})
	if !reflect.DeepEqual(atas, []int{1}) || !reflect.DeepEqual(bawah, []int{0}) {
		t.Fatalf("n=2 -> atas=%v bawah=%v", atas, bawah)
	}
}

func TestPengecoh(t *testing.T) {
	hasil := Pengecoh(map[string]int{"A": 600, "B": 250, "C": 100, "D": 50}, 1000)
	if len(hasil) != 4 {
		t.Fatalf("len = %d, want 4", len(hasil))
	}
	// Terurut berdasarkan label opsi.
	if hasil[0].Opsi != "A" || hasil[3].Opsi != "D" {
		t.Fatalf("urutan opsi = %v", hasil)
	}
	hampirSama(t, hasil[0].Persen, 60)
	hampirSama(t, hasil[3].Persen, 5)
	if !hasil[3].Berfungsi {
		t.Fatal("pengecoh tepat 5% harus Berfungsi")
	}
	if hasil[1].Pemilih != 250 {
		t.Fatalf("pemilih B = %d, want 250", hasil[1].Pemilih)
	}
}

func TestPengecohAmbangLimaPersen(t *testing.T) {
	// 49/1000 = 4,9% -> tidak berfungsi; 51/1000 = 5,1% -> berfungsi.
	hasil := Pengecoh(map[string]int{"A": 49, "B": 51}, 1000)
	hampirSama(t, hasil[0].Persen, 4.9)
	if hasil[0].Berfungsi {
		t.Fatal("pengecoh 4,9% harus TIDAK berfungsi")
	}
	hampirSama(t, hasil[1].Persen, 5.1)
	if !hasil[1].Berfungsi {
		t.Fatal("pengecoh 5,1% harus berfungsi")
	}
}

func TestPengecohKasusBatas(t *testing.T) {
	// totalPenjawab 0 -> semua persen 0, semua tidak berfungsi.
	hasil := Pengecoh(map[string]int{"A": 0, "B": 0}, 0)
	for _, h := range hasil {
		hampirSama(t, h.Persen, 0)
		if h.Berfungsi {
			t.Fatalf("opsi %s dengan total 0 harus tidak berfungsi", h.Opsi)
		}
	}
	// Pemilih negatif dijepit ke 0.
	hasil = Pengecoh(map[string]int{"A": -5}, 100)
	if hasil[0].Pemilih != 0 || hasil[0].Persen != 0 {
		t.Fatalf("pemilih negatif: %+v", hasil[0])
	}
	// Map kosong -> hasil kosong (bukan nil panic).
	if got := Pengecoh(map[string]int{}, 100); len(got) != 0 {
		t.Fatalf("map kosong -> %v", got)
	}
}

func TestNilaiKompetensi(t *testing.T) {
	// (8+6)/(10+10)*100 = 70.
	nilai, flag := NilaiKompetensi([]float64{8, 6}, []float64{10, 10}, []bool{true, true})
	if flag != "" {
		t.Fatalf("flag = %q, want kosong", flag)
	}
	if nilai == nil {
		t.Fatal("nilai nil, want 70")
	}
	hampirSama(t, *nilai, 70)

	// Soal yang tidak dijawab tidak menyumbang poin maupun bobot:
	// hanya soal pertama: 8/10*100 = 80 (bukan (8+0)/20*100 = 40).
	nilai, flag = NilaiKompetensi([]float64{8, 9}, []float64{10, 10}, []bool{true, false})
	if flag != "" || nilai == nil {
		t.Fatalf("nilai=%v flag=%q", nilai, flag)
	}
	hampirSama(t, *nilai, 80)

	// Nilai sempurna.
	nilai, _ = NilaiKompetensi([]float64{10}, []float64{10}, []bool{true})
	hampirSama(t, *nilai, 100)
}

func TestNilaiKompetensiTanpaJawaban(t *testing.T) {
	// Semua tidak dijawab -> null + "tidak ada soal".
	nilai, flag := NilaiKompetensi([]float64{8, 6}, []float64{10, 10}, []bool{false, false})
	if nilai != nil {
		t.Fatalf("nilai = %v, want nil", *nilai)
	}
	if flag != "tidak ada soal" {
		t.Fatalf("flag = %q, want %q", flag, "tidak ada soal")
	}
	// Slice kosong -> sama.
	nilai, flag = NilaiKompetensi(nil, nil, nil)
	if nilai != nil || flag != "tidak ada soal" {
		t.Fatalf("kosong -> nilai=%v flag=%q", nilai, flag)
	}
}

func TestNilaiKompetensiBobotNol(t *testing.T) {
	nilai, flag := NilaiKompetensi([]float64{5}, []float64{0}, []bool{true})
	if nilai != nil {
		t.Fatalf("nilai = %v, want nil", *nilai)
	}
	if flag != "bobot nol" {
		t.Fatalf("flag = %q, want %q", flag, "bobot nol")
	}
}

func TestNilaiKompetensiPanjangBerbeda(t *testing.T) {
	// Hanya panjang terpendek yang dibaca: 1 soal dijawab -> 8/10*100.
	nilai, flag := NilaiKompetensi([]float64{8, 6, 4}, []float64{10}, []bool{true, true, true})
	if flag != "" || nilai == nil {
		t.Fatalf("nilai=%v flag=%q", nilai, flag)
	}
	hampirSama(t, *nilai, 80)
}

func TestAmbangBenarUraian(t *testing.T) {
	hampirSama(t, AmbangBenarUraian(), 0.6)
}
