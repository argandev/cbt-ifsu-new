package ukk

import (
	"testing"

	"cbt-smk-v2/backend/internal/platform"
)

func f(v float64) *float64 { return &v }

func TestHitungPraktik(t *testing.T) {
	// 80*0.3 + 90*0.5 + 70*0.2 = 24 + 45 + 14 = 83
	p := hitungPraktik(f(80), f(90), f(70))
	if p == nil || *p != 83 {
		t.Fatalf("harap praktik 83, dapat %v", p)
	}
	if hitungPraktik(f(80), nil, f(70)) != nil {
		t.Fatal("rubrik belum lengkap harus menghasilkan nil")
	}
}

func TestHitungAkhirDanPredikat(t *testing.T) {
	// teori 80, praktik 90 -> 80*0.4 + 90*0.6 = 32 + 54 = 86 -> Baik, lulus
	praktik, akhir, lulus, predikat := hitungNilaiInput(nilaiInput{
		NilaiTeori: f(80), NilaiProses: f(90), NilaiProduk: f(90), NilaiSikap: f(90),
	})
	if praktik == nil || *praktik != 90 {
		t.Fatalf("harap praktik 90, dapat %v", praktik)
	}
	if akhir == nil || *akhir != 86 {
		t.Fatalf("harap akhir 86, dapat %v", akhir)
	}
	if lulus == nil || !*lulus {
		t.Fatalf("harap lulus, dapat %v", lulus)
	}
	if predikat == nil || *predikat != "Baik" {
		t.Fatalf("harap predikat Baik, dapat %v", predikat)
	}
	// Batas bawah lulus: 70 persis -> lulus, Cukup.
	_, akhir, lulus, predikat = hitungNilaiInput(nilaiInput{
		NilaiTeori: f(70), NilaiProses: f(70), NilaiProduk: f(70), NilaiSikap: f(70),
	})
	if akhir == nil || *akhir != 70 || lulus == nil || !*lulus || predikat == nil || *predikat != "Cukup" {
		t.Fatalf("batas 70: harap akhir 70/lulus/Cukup, dapat %v/%v/%v", akhir, lulus, predikat)
	}
	// Di bawah 70 -> tidak lulus, Belum Tuntas.
	_, akhir, lulus, predikat = hitungNilaiInput(nilaiInput{
		NilaiTeori: f(60), NilaiProses: f(60), NilaiProduk: f(60), NilaiSikap: f(60),
	})
	if akhir == nil || *akhir != 60 || lulus == nil || *lulus || predikat == nil || *predikat != "Belum Tuntas" {
		t.Fatalf("bawah 70: harap akhir 60/tidak lulus/Belum Tuntas, dapat %v/%v/%v", akhir, lulus, predikat)
	}
	// Hanya teori, tanpa rubrik -> akhir nil (belum lengkap).
	_, akhir, _, _ = hitungNilaiInput(nilaiInput{NilaiTeori: f(90)})
	if akhir != nil {
		t.Fatalf("tanpa rubrik, akhir harus nil, dapat %v", akhir)
	}
}

func TestPredikatBatas(t *testing.T) {
	cases := []struct {
		nilai float64
		want  string
	}{
		{0, "Belum Tuntas"}, {69.99, "Belum Tuntas"}, {70, "Cukup"},
		{79.99, "Cukup"}, {80, "Baik"}, {90, "Baik"},
		{90.01, "Sangat Baik"}, {100, "Sangat Baik"},
	}
	for _, tc := range cases {
		if got := platform.Predikat(tc.nilai); got != tc.want {
			t.Fatalf("platform.Predikat(%v) = %q, harap %q", tc.nilai, got, tc.want)
		}
	}
}
