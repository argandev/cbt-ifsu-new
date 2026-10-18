package banksoal

import (
	"encoding/json"
	"testing"
)

// Tes pembahasan (Fase 3): soalInput/soalRow membawa field pembahasan
// (persist via INSERT/SELECT/UPDATE di handler; di sini diuji
// serialisasi + validasi tidak menolak pembahasan).
func TestSoalInputPembahasan(t *testing.T) {
	in := soalInput{
		Tipe: "pg", Pertanyaan: "Ibu kota Indonesia?", Bobot: 10,
		Jawaban: raw(map[string]any{"opsi": []any{
			map[string]any{"label": "A", "teks": "Jakarta"},
			map[string]any{"label": "B", "teks": "Bandung"},
		}}),
		Kunci:      raw(map[string]any{"opsi": "A"}),
		Pembahasan: "Jakarta adalah ibu kota sejak 1945.",
	}
	if err := in.validate(); err != nil {
		t.Fatalf("soal dengan pembahasan harus valid: %v", err)
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var kembali soalInput
	if err := json.Unmarshal(b, &kembali); err != nil {
		t.Fatal(err)
	}
	if kembali.Pembahasan != "Jakarta adalah ibu kota sejak 1945." {
		t.Fatalf("pembahasan hilang dalam roundtrip JSON: %q", kembali.Pembahasan)
	}
	var row soalRow
	row.Pembahasan = "Penjelasan guru."
	b, _ = json.Marshal(row)
	var kembaliRow soalRow
	if err := json.Unmarshal(b, &kembaliRow); err != nil {
		t.Fatal(err)
	}
	if kembaliRow.Pembahasan != "Penjelasan guru." {
		t.Fatalf("soalRow kehilangan pembahasan: %q", kembaliRow.Pembahasan)
	}
}
