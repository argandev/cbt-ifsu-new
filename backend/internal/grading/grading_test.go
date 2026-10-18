// Unit test penskoran (table-driven). Murni: tanpa DB, tanpa HTTP.
package grading_test

import (
	"testing"

	"cbt-smk-v2/backend/internal/grading"
)

func TestHitungPoinPG(t *testing.T) {
	poin, auto := grading.HitungPoin("pg", []byte(`"B"`), []byte(`"B"`), 10)
	if !auto || poin != 10 {
		t.Fatalf("PG benar: poin=%v auto=%v, mau 10 true", poin, auto)
	}
	poin, _ = grading.HitungPoin("pg", []byte(`"A"`), []byte(`"B"`), 10)
	if poin != 0 {
		t.Fatalf("PG salah: poin=%v, mau 0", poin)
	}
	poin, _ = grading.HitungPoin("pg", []byte(`{"pilihan":"B"}`), []byte(`{"jawaban":"B"}`), 5)
	if poin != 5 {
		t.Fatalf("PG bentuk objek: poin=%v, mau 5", poin)
	}
	poin, _ = grading.HitungPoin("pg", nil, []byte(`"B"`), 10)
	if poin != 0 {
		t.Fatalf("PG kosong: poin=%v, mau 0", poin)
	}
}

func TestHitungPoinPGKompleks(t *testing.T) {
	// 2 dari 3 kunci benar -> 10 * 2/3.
	poin, auto := grading.HitungPoin("pg_kompleks", []byte(`["A","C"]`), []byte(`["A","B","C"]`), 30)
	if !auto || poin != 20 {
		t.Fatalf("PGK 2/3: poin=%v auto=%v, mau 20 true", poin, auto)
	}
	// Jawaban salah tidak mengurangi (tanpa penalti): A benar + D salah -> 1/3.
	poin, _ = grading.HitungPoin("pg_kompleks", []byte(`["A","D"]`), []byte(`["A","B","C"]`), 30)
	if poin != 10 {
		t.Fatalf("PGK tanpa penalti: poin=%v, mau 10", poin)
	}
}

func TestHitungPoinBenarSalah(t *testing.T) {
	poin, _ := grading.HitungPoin("benar_salah",
		[]byte(`{"p1":true,"p2":false,"p3":true}`),
		[]byte(`{"p1":true,"p2":true,"p3":true}`), 9)
	if poin != 6 {
		t.Fatalf("Benar-Salah 2/3: poin=%v, mau 6", poin)
	}
}

func TestHitungPoinBenarSalahKunciWarisanPN(t *testing.T) {
	// Kunci warisan {"p1":..} + jawaban array [..] disejajarkan per posisi.
	poin, _ := grading.HitungPoin("benar_salah",
		[]byte(`[true,true]`),
		[]byte(`{"p1":true,"p2":true}`), 10)
	if poin != 10 {
		t.Fatalf("Benar-Salah pN+array semua benar: poin=%v, mau 10", poin)
	}
	poin, _ = grading.HitungPoin("benar_salah",
		[]byte(`[true,false]`),
		[]byte(`{"p1":true,"p2":true}`), 10)
	if poin != 5 {
		t.Fatalf("Benar-Salah pN+array 1/2: poin=%v, mau 5", poin)
	}
	// Map-map tetap seperti semula.
	poin, _ = grading.HitungPoin("benar_salah",
		[]byte(`{"p1":true,"p2":true}`),
		[]byte(`{"p1":true,"p2":false}`), 10)
	if poin != 5 {
		t.Fatalf("Benar-Salah map-map 1/2: poin=%v, mau 5", poin)
	}
}

func TestHitungPoinMenjodohkan(t *testing.T) {
	poin, _ := grading.HitungPoin("menjodohkan",
		[]byte(`{"k1":"r2","k2":"r1"}`),
		[]byte(`{"k1":"r1","k2":"r2"}`), 20)
	if poin != 0 {
		t.Fatalf("Menjodohkan semua salah: poin=%v, mau 0", poin)
	}
	poin, _ = grading.HitungPoin("menjodohkan",
		[]byte(`{"k1":"r1","k2":"rX"}`),
		[]byte(`{"k1":"r1","k2":"r2"}`), 20)
	if poin != 10 {
		t.Fatalf("Menjodohkan 1/2: poin=%v, mau 10", poin)
	}
}

func TestHitungPoinIsian(t *testing.T) {
	poin, auto := grading.HitungPoin("isian_singkat", []byte(`"  JAKARTA "`), []byte(`"jakarta"`), 10)
	if !auto || poin != 10 {
		t.Fatalf("Isian normalisasi: poin=%v auto=%v, mau 10 true", poin, auto)
	}
	poin, _ = grading.HitungPoin("isian_singkat", []byte(`"Bandung"`), []byte(`["Jakarta","DKI Jakarta"]`), 10)
	if poin != 0 {
		t.Fatalf("Isian tidak cocok: poin=%v, mau 0", poin)
	}
	poin, _ = grading.HitungPoin("isian_singkat", []byte(`"DKI   Jakarta"`), []byte(`["Jakarta","DKI Jakarta"]`), 10)
	if poin != 10 {
		t.Fatalf("Isian alternatif + spasi: poin=%v, mau 10", poin)
	}
}

func TestHitungPoinUraianManual(t *testing.T) {
	poin, auto := grading.HitungPoin("uraian", []byte(`"jawaban panjang"`), []byte(`"rubrik"`), 25)
	if auto || poin != 0 {
		t.Fatalf("Uraian: poin=%v auto=%v, mau 0 false", poin, auto)
	}
	poin, auto = grading.HitungPoin("esai_singkat", []byte(`"x"`), nil, 10)
	if auto || poin != 0 {
		t.Fatalf("Esai singkat: poin=%v auto=%v, mau 0 false", poin, auto)
	}
}

func TestHitungPoinMengurutkan(t *testing.T) {
	poin, _ := grading.HitungPoin("mengurutkan",
		[]byte(`["s1","s3","s2"]`), []byte(`["s1","s2","s3"]`), 30)
	if poin != 10 {
		t.Fatalf("Mengurutkan 1/3 posisi: poin=%v, mau 10", poin)
	}
}

func TestHitungPoinSkalaSikapTidakDinilai(t *testing.T) {
	poin, auto := grading.HitungPoin("skala_sikap", []byte(`"sangat setuju"`), nil, 10)
	if !auto || poin != 0 {
		t.Fatalf("Skala sikap: poin=%v auto=%v, mau 0 true", poin, auto)
	}
	if grading.Dinilai("skala_sikap", nil) {
		t.Fatal("Skala sikap harus Dinilai=false")
	}
	if grading.Dinilai("setuju_tidak_setuju", nil) {
		t.Fatal("Setuju-tidak-setuju tanpa kunci harus Dinilai=false")
	}
	if !grading.Dinilai("setuju_tidak_setuju", []byte(`{"p1":true}`)) {
		t.Fatal("Setuju-tidak-setuju berkunci harus Dinilai=true")
	}
	if !grading.Dinilai("pg", []byte(`"A"`)) {
		t.Fatal("PG harus Dinilai=true")
	}
}

func TestHitungNilaiAkhir(t *testing.T) {
	nilai, menunggu := grading.HitungNilaiAkhir([]grading.JawabanNilai{
		{Poin: 10, Bobot: 10, Dinilai: true},
		{Poin: 5, Bobot: 10, Dinilai: true},
		{Poin: 0, Bobot: 10, Dinilai: true, MenungguKoreksi: true},
		{Poin: 0, Bobot: 10, Dinilai: false}, // skala sikap: tidak masuk
	})
	// (10+5+0)/(10+10+10)*100 = 50.
	if nilai != 50 {
		t.Fatalf("Nilai akhir: %v, mau 50", nilai)
	}
	if !menunggu {
		t.Fatal("menungguKoreksi harus true")
	}

	nilai, menunggu = grading.HitungNilaiAkhir(nil)
	if nilai != 0 || menunggu {
		t.Fatalf("List kosong: nilai=%v menunggu=%v, mau 0 false", nilai, menunggu)
	}
}

func TestNormalisasiTipe(t *testing.T) {
	cases := map[string]string{
		"PG": "pg", "Pilihan Ganda": "pg", "pilihan_ganda_kompleks": "pg_kompleks",
		"Benar-Salah": "benar_salah", "Esai": "uraian", "Isian Panjang": "uraian",
		"Mengurutkan": "mengurutkan",
		// Tim 3: studi kasus (soal PG berstimulus) dinilai seperti PG.
		"Studi Kasus": "pg",
	}
	for in, mau := range cases {
		if got := grading.NormalisasiTipe(in); got != mau {
			t.Fatalf("NormalisasiTipe(%q)=%q, mau %q", in, got, mau)
		}
	}
}
