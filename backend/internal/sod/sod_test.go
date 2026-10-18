package sod

import (
	"strings"
	"testing"
)

// Tes SoD override (Fase 3): alasan pendek -> 400 (validasi),
// sukses -> audit flag SOD_OVERRIDE + notifikasi (jalur DB diuji
// di tes live; di sini diuji fungsi murni pembangun payload-nya).
func TestValidasiAlasan(t *testing.T) {
	if got := ValidasiAlasan(""); got == "" {
		t.Fatal("alasan kosong harus ditolak")
	}
	if got := ValidasiAlasan("kurang dari 20"); got == "" {
		t.Fatal("alasan 13 karakter harus ditolak")
	}
	if got := ValidasiAlasan(strings.Repeat("a", 19)); got == "" {
		t.Fatal("alasan 19 karakter harus ditolak")
	}
	if got := ValidasiAlasan(strings.Repeat("a", 20)); got != "" {
		t.Fatalf("alasan 20 karakter harus lolos, dapat: %s", got)
	}
	if got := ValidasiAlasan("Guru pengganti sakit mendadak hari ini."); got != "" {
		t.Fatalf("alasan panjang harus lolos, dapat: %s", got)
	}
}

func TestDetailAuditMemuatFlagSODOverride(t *testing.T) {
	d := DetailAudit("banksoal.lock", "bank_soal", "id-1", "Pengunci utama sakit mendadak.")
	if d["flag"] != "SOD_OVERRIDE" {
		t.Fatalf("detail audit harus memuat flag SOD_OVERRIDE, dapat %v", d)
	}
	if d["sod_override"] != true {
		t.Fatalf("detail audit harus memuat sod_override=true, dapat %v", d)
	}
	if d["aksi"] != "banksoal.lock" || d["entitas"] != "bank_soal" || d["entitas_id"] != "id-1" {
		t.Fatalf("detail audit harus memuat aksi+entitas+entitas_id, dapat %v", d)
	}
	if d["alasan"] != "Pengunci utama sakit mendadak." {
		t.Fatalf("detail audit harus memuat alasan, dapat %v", d)
	}
}

func TestIsiNotifikasi(t *testing.T) {
	isi := IsiNotifikasi("banksoal.lock", "bank_soal", "id-1", "Pengunci utama sakit mendadak.", "Admin Satu")
	for _, want := range []string{"banksoal.lock", "bank_soal", "Pengunci utama sakit mendadak.", "Admin Satu"} {
		if !strings.Contains(isi, want) {
			t.Fatalf("isi notifikasi harus memuat %q, dapat %q", want, isi)
		}
	}
}
