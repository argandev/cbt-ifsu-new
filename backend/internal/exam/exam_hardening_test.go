package exam

import "testing"

// Tes HH:MM (Fase 3): gabungTanggalJam menerima jam "08:00"
// (dinormalisasi ke "08:00:00") selain "HH:MM:SS".
func TestGabungTanggalJamHHMM(t *testing.T) {
	got, err := gabungTanggalJam("2026-10-16", "08:00")
	if err != nil {
		t.Fatalf("jam HH:MM harus diterima: %v", err)
	}
	if got.Hour() != 8 || got.Minute() != 0 || got.Second() != 0 {
		t.Fatalf("jam 08:00 dinormalisasi salah: %v", got)
	}
	got2, err := gabungTanggalJam("2026-10-16", "08:00:30")
	if err != nil {
		t.Fatalf("jam HH:MM:SS harus tetap diterima: %v", err)
	}
	if got2.Hour() != 8 || got2.Minute() != 0 || got2.Second() != 30 {
		t.Fatalf("jam 08:00:30 salah: %v", got2)
	}
	if _, err := gabungTanggalJam("2026-10-16", "25:00"); err == nil {
		t.Fatal("jam 25:00 harus ditolak")
	}
	if _, err := gabungTanggalJam("2026-10-16", "8:00"); err == nil {
		t.Fatal("jam 8:00 (tanpa nol depan) harus ditolak — format HH:MM")
	}
}
