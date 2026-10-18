package master

import (
	"encoding/json"
	"strings"
	"testing"
)

// Tes import idempotent (Fase 3, BUG-02): respons import membawa field
// sudah_ada — NISN yang sudah ada dihitung terpisah dari gagal.
// (Alur DB penuh: import 300 siswa 2x -> sukses 0, sudah_ada 300,
// gagal 0 — diuji di tes live server scratch.)
func TestImportResultMemuatSudahAda(t *testing.T) {
	r := importResult{Sukses: 0, Gagal: 0, SudahAda: 300, Errors: []importRowError{}}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"sudah_ada":300`) {
		t.Fatalf("respons import harus memuat sudah_ada, dapat %s", b)
	}
	var kembali importResult
	if err := json.Unmarshal(b, &kembali); err != nil {
		t.Fatal(err)
	}
	if kembali.SudahAda != 300 || kembali.Sukses != 0 || kembali.Gagal != 0 {
		t.Fatalf("roundtrip importResult salah: %+v", kembali)
	}
}
