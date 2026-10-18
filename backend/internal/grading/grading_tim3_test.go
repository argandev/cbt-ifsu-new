// Tes bentuk kunci Tim 3 (frontend/src/lib/banksoal.ts): kunci {"opsi":...},
// {"nilai":[...]}, {"pasangan":[...]}, {"jawaban":[...]}, {"urutan":[...]}.
package grading_test

import (
	"testing"

	"cbt-smk-v2/backend/internal/grading"
)

func TestHitungPoinBentukTim3(t *testing.T) {
	// PG: kunci {"opsi":"B"}.
	if poin, _ := grading.HitungPoin("pg", []byte(`"B"`), []byte(`{"opsi":"B"}`), 10); poin != 10 {
		t.Fatalf("PG Tim3: poin=%v, mau 10", poin)
	}
	// PG Kompleks: kunci {"opsi":["A","C"]}.
	if poin, _ := grading.HitungPoin("pg_kompleks", []byte(`["A"]`), []byte(`{"opsi":["A","C"]}`), 20); poin != 10 {
		t.Fatalf("PGK Tim3 1/2: poin=%v, mau 10", poin)
	}
	// pg_multi = PG Kompleks.
	if poin, _ := grading.HitungPoin("pg_multi", []byte(`["A","C"]`), []byte(`{"opsi":["A","C"]}`), 20); poin != 20 {
		t.Fatalf("pg_multi Tim3: poin=%v, mau 20", poin)
	}
	// Benar-Salah: kunci {"nilai":[true,false]}; jawaban [true,true] -> 1/2.
	if poin, _ := grading.HitungPoin("benar_salah", []byte(`[true,true]`), []byte(`{"nilai":[true,false]}`), 10); poin != 5 {
		t.Fatalf("Benar-Salah Tim3: poin=%v, mau 5", poin)
	}
	// Menjodohkan: kunci {"pasangan":[...]}; jawaban map.
	if poin, _ := grading.HitungPoin("menjodohkan",
		[]byte(`{"Jakarta":"Indonesia","Tokyo":"Jepang"}`),
		[]byte(`{"pasangan":[{"kiri":"Jakarta","kanan":"Indonesia"},{"kiri":"Tokyo","kanan":"Jepang"}]}`), 20); poin != 20 {
		t.Fatalf("Menjodohkan Tim3: poin=%v, mau 20", poin)
	}
	// Peta konsep = menjodohkan.
	if poin, _ := grading.HitungPoin("peta_konsep",
		[]byte(`[{"kiri":"A","kanan":"1"}]`),
		[]byte(`{"pasangan":[{"kiri":"A","kanan":"1"}]}`), 10); poin != 10 {
		t.Fatalf("Peta konsep Tim3: poin=%v, mau 10", poin)
	}
	// Isian (kode Tim 3 "isian"): kunci {"jawaban":[...]}.
	if poin, _ := grading.HitungPoin("isian", []byte(`"jakarta"`), []byte(`{"jawaban":["Jakarta","DKI Jakarta"]}`), 10); poin != 10 {
		t.Fatalf("Isian Tim3: poin=%v, mau 10", poin)
	}
	// Isi ganda: proporsional per kolom.
	if poin, _ := grading.HitungPoin("isi_ganda", []byte(`["a","X"]`), []byte(`{"jawaban":["a","b"]}`), 10); poin != 5 {
		t.Fatalf("Isi ganda Tim3: poin=%v, mau 5", poin)
	}
	// Teks rumpang: proporsional per rumpang.
	if poin, _ := grading.HitungPoin("teks_rumpang", []byte(`["x","y"]`), []byte(`{"jawaban":["x","y"]}`), 10); poin != 10 {
		t.Fatalf("Teks rumpang Tim3: poin=%v, mau 10", poin)
	}
	// Urutkan: kunci {"urutan":[2,0,1]}; jawaban indeks [2,0,1] -> penuh.
	if poin, _ := grading.HitungPoin("urutkan", []byte(`[2,0,1]`), []byte(`{"urutan":[2,0,1]}`), 30); poin != 30 {
		t.Fatalf("Urutkan indeks Tim3: poin=%v, mau 30", poin)
	}
	// Urutkan: jawaban string + kunci diperkaya "item".
	if poin, _ := grading.HitungPoin("urutkan", []byte(`["C","A","B"]`),
		[]byte(`{"urutan":[2,0,1],"item":["A","B","C"]}`), 30); poin != 30 {
		t.Fatalf("Urutkan string Tim3: poin=%v, mau 30", poin)
	}
	// Pernyataan sikap tidak dinilai.
	if grading.Dinilai("pernyataan_sikap", nil) {
		t.Fatal("pernyataan_sikap harus Dinilai=false")
	}
	// Upload file / praktik singkat = manual.
	if _, auto := grading.HitungPoin("upload_file", []byte(`"file.pdf"`), nil, 10); auto {
		t.Fatal("upload_file harus manual (auto=false)")
	}
	// Pilihan gambar = PG.
	if poin, _ := grading.HitungPoin("pilihan_gambar", []byte(`"C"`), []byte(`{"opsi":"C"}`), 10); poin != 10 {
		t.Fatalf("Pilihan gambar Tim3: poin=%v, mau 10", poin)
	}
}
