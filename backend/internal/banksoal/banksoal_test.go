package banksoal

import (
	"encoding/json"
	"testing"
)

func raw(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func TestTipeSoalCount(t *testing.T) {
	if len(TipeSoal) != 16 {
		t.Fatalf("harus 16 tipe, dapat %d", len(TipeSoal))
	}
	inti, varian := 0, 0
	for _, info := range TipeSoal {
		switch info.Kelompok {
		case KelompokInti:
			inti++
		case KelompokVarian:
			varian++
		default:
			t.Fatalf("tipe %s kelompok tidak valid: %s", info.Kode, info.Kelompok)
		}
	}
	if inti != 5 || varian != 11 {
		t.Fatalf("harus 5 inti + 11 varian, dapat %d + %d", inti, varian)
	}
	if TipeSoal["pernyataan_sikap"].Dinilai {
		t.Fatal("pernyataan_sikap harus Dinilai=false (non-nilai)")
	}
}

func TestValidateSoalPG(t *testing.T) {
	jawaban := raw(map[string]any{"opsi": []any{
		map[string]any{"label": "A", "teks": "Satu"},
		map[string]any{"label": "B", "teks": "Dua"},
	}})
	if err := ValidateSoal("pg", jawaban, raw(map[string]any{"opsi": "B"})); err != nil {
		t.Fatal(err)
	}
	if err := ValidateSoal("pg", jawaban, raw(map[string]any{"opsi": "Z"})); err == nil {
		t.Fatal("kunci di luar opsi harus ditolak")
	}
	if err := ValidateSoal("pg", jawaban, raw(map[string]any{"opsi": "A"})); err != nil {
		t.Fatal(err)
	}
}

func TestValidateSoalBenarSalah(t *testing.T) {
	jawaban := raw(map[string]any{"pernyataan": []any{"P1", "P2"}})
	if err := ValidateSoal("benar_salah", jawaban, raw(map[string]any{"nilai": []any{true, false}})); err != nil {
		t.Fatal(err)
	}
	if err := ValidateSoal("benar_salah", jawaban, raw(map[string]any{"nilai": []any{true}})); err == nil {
		t.Fatal("jumlah nilai != pernyataan harus ditolak")
	}
}

func TestValidateSoalTeksRumpang(t *testing.T) {
	if err := ValidateSoal("teks_rumpang", raw(map[string]any{}),
		raw(map[string]any{"jawaban": []any{"a", "b"}})); err != nil {
		t.Fatal(err)
	}
}

func TestScoreAutoPGKompleksProporsional(t *testing.T) {
	// Kunci A,C (2 benar). Siswa pilih A saja -> 1/2 bobot. Tanpa penalti:
	// siswa pilih A + B(salah) -> tetap 1/2 bobot.
	kunci := raw(map[string]any{"opsi": []any{"A", "C"}})
	p, auto, err := ScoreAuto("pg_kompleks", kunci, raw(map[string]any{"opsi": []any{"A"}}), 4)
	if err != nil || !auto || p != 2 {
		t.Fatalf("harap poin 2, dapat %v auto=%v err=%v", p, auto, err)
	}
	p, _, _ = ScoreAuto("pg_kompleks", kunci, raw(map[string]any{"opsi": []any{"A", "B"}}), 4)
	if p != 2 {
		t.Fatalf("tanpa penalti: harap poin 2, dapat %v", p)
	}
}

func TestScoreAutoIsianNormalisasi(t *testing.T) {
	kunci := raw(map[string]any{"jawaban": []any{"Jakarta"}})
	p, _, _ := ScoreAuto("isian", kunci, raw(map[string]any{"jawaban": "  JAKARTA "}), 1)
	if p != 1 {
		t.Fatalf("normalisasi trim+lowercase harus cocok, dapat %v", p)
	}
	p, _, _ = ScoreAuto("isian", kunci, raw(map[string]any{"jawaban": "Bandung"}), 1)
	if p != 0 {
		t.Fatalf("jawaban salah harus 0, dapat %v", p)
	}
}

func TestScoreAutoManualBukanAuto(t *testing.T) {
	for _, tipe := range []string{"uraian", "upload_file", "praktik_singkat", "pernyataan_sikap"} {
		_, auto, err := ScoreAuto(tipe, raw(map[string]any{}), raw(map[string]any{}), 1)
		if err != nil || auto {
			t.Fatalf("%s harus isAuto=false, dapat auto=%v err=%v", tipe, auto, err)
		}
	}
}

func TestScoreAutoMenjodohkan(t *testing.T) {
	kunci := raw(map[string]any{"pasangan": []any{
		map[string]any{"kiri": "K1", "kanan": "V1"},
		map[string]any{"kiri": "K2", "kanan": "V2"},
	}})
	siswa := raw(map[string]any{"pasangan": []any{
		map[string]any{"kiri": "K1", "kanan": "V1"},
		map[string]any{"kiri": "K2", "kanan": "SALAH"},
	}})
	p, _, _ := ScoreAuto("menjodohkan", kunci, siswa, 2)
	if p != 1 {
		t.Fatalf("1 dari 2 pasangan benar, bobot 2 -> harap 1, dapat %v", p)
	}
}

func TestParseCSVImportPG(t *testing.T) {
	csvText := "tipe,pertanyaan,opsi_a,opsi_b,opsi_c,opsi_d,opsi_e,kunci,bobot,level,stimulus\n" +
		"pg,Berapa 1+1?,1,2,3,4,,B,1,sedang,\n" +
		"pg,Salah,,2,,,,Z,1,sedang,\n"
	rows, errs := parseCSVImport(csvText)
	if len(rows) != 1 {
		t.Fatalf("harap 1 baris valid, dapat %d", len(rows))
	}
	if len(errs) != 1 {
		t.Fatalf("harap 1 baris gagal, dapat %d: %v", len(errs), errs)
	}
	if rows[0].tipe != "pg" || rows[0].bobot != 1 {
		t.Fatalf("baris valid salah: %+v", rows[0])
	}
}

func TestParseCSVImportBenarSalah(t *testing.T) {
	csvText := "tipe,pertanyaan,opsi_a,opsi_b,opsi_c,opsi_d,opsi_e,kunci,bobot,level,stimulus\n" +
		"benar_salah,Tentukan benar/salah,Air basah,Api dingin,,,,\"B,S\",2,mudah,\n"
	rows, errs := parseCSVImport(csvText)
	if len(errs) != 0 {
		t.Fatalf("harap tanpa error, dapat %v", errs)
	}
	if len(rows) != 1 {
		t.Fatalf("harap 1 baris, dapat %d", len(rows))
	}
}

func TestParseCSVImportBatas500(t *testing.T) {
	csvText := "tipe,pertanyaan,opsi_a,opsi_b,opsi_c,opsi_d,opsi_e,kunci,bobot,level,stimulus\n"
	for i := 0; i < 501; i++ {
		csvText += "pg,Q?,1,2,,,,A,1,sedang,\n"
	}
	_, errs := parseCSVImport(csvText)
	if len(errs) == 0 {
		t.Fatal("501 baris harus ditolak (batas 500)")
	}
}
