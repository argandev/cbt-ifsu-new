package analisis

import "testing"

// Tes SoD proktor vs analisis (Fase 3): respons GET analisis untuk
// proktor MENYEMBUNYIKAN field pertanyaan, statistik tetap tampil.
// Guru/admin tetap melihat pertanyaan.
func TestSembunyikanPertanyaanUntukProktor(t *testing.T) {
	hasil := analisisResult{
		JadwalID: "jadwal-1",
		Items: []analisisItem{
			{SoalID: "s1", Pertanyaan: "Ibu kota Indonesia?", JumlahBenar: 3, JumlahPenjab: 5},
			{SoalID: "s2", Pertanyaan: "2 + 2?", JumlahBenar: 4, JumlahPenjab: 5},
		},
	}
	proktor := SembunyikanPertanyaanUntukProktor(hasil, "proktor")
	for _, it := range proktor.Items {
		if it.Pertanyaan != "" {
			t.Fatalf("proktor tidak boleh melihat pertanyaan, dapat %q", it.Pertanyaan)
		}
		if it.JumlahBenar == 0 || it.JumlahPenjab == 0 {
			t.Fatal("statistik proktor harus tetap tampil")
		}
	}
	guru := SembunyikanPertanyaanUntukProktor(hasil, "guru")
	if guru.Items[0].Pertanyaan != "Ibu kota Indonesia?" {
		t.Fatalf("guru harus tetap melihat pertanyaan, dapat %q", guru.Items[0].Pertanyaan)
	}
	admin := SembunyikanPertanyaanUntukProktor(hasil, "admin")
	if admin.Items[1].Pertanyaan != "2 + 2?" {
		t.Fatalf("admin harus tetap melihat pertanyaan, dapat %q", admin.Items[1].Pertanyaan)
	}
}
