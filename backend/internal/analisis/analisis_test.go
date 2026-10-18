// Tes fungsi murni analisis butir + nilai per kompetensi (tanpa DB).
//
// Skenario tes kecil (dihitung manual, angka contoh untuk verifikasi):
//
//	10 peserta, nilai akhir: 95, 90, 85, 80, 75, 70, 65, 60, 55, 50.
//	N = 10 >= 10 -> metode 27%, n = round(2.7) = 3.
//	Kelompok atas = {95, 90, 85}; kelompok bawah = {60, 55, 50}.
package analisis

import (
	"encoding/json"
	"testing"
)

func fptr(v float64) *float64 { return &v }

func jwb(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func TestComputeAnalisisContohKecil(t *testing.T) {
	nilai := []float64{95, 90, 85, 80, 75, 70, 65, 60, 55, 50}
	var attempts []attemptInput
	for i, n := range nilai {
		attempts = append(attempts, attemptInput{ID: string(rune('A' + i)), SiswaID: "siswa", Nilai: fptr(n)})
	}
	// attemptID: A=95, B=90, C=85, D=80, E=75, F=70, G=65, H=60, I=55, J=50.
	// Atas = A,B,C. Bawah = H,I,J.
	soals := []soalInput{
		{ID: "soal1", Urutan: 1, Pertanyaan: "Ibu kota Indonesia?", Tipe: "pg", Bobot: 1,
			Opsi: []opsiSoal{{"A", "Jakarta"}, {"B", "Bandung"}, {"C", "Surabaya"}, {"D", "Medan"}}},
		{ID: "soal2", Urutan: 2, Pertanyaan: "2 + 2?", Tipe: "pg", Bobot: 1,
			Opsi: []opsiSoal{{"A", "4"}, {"B", "5"}}},
		{ID: "soal3", Urutan: 3, Pertanyaan: "Jelaskan fotosintesis.", Tipe: "uraian", Bobot: 2},
	}
	// Soal 1: benar = atas 3/3, tengah 2/4, bawah 1/3 -> total benar 6/10.
	// Pengecoh: A 6 (60%), B 2 (20%), C 2 (20%), D 0 (0%, tidak berfungsi).
	// Soal 2: benar = atas 0/3, tengah 2/4, bawah 3/3 -> total 5/10, D = -1.
	// Soal 3 (uraian, ambang benar 0.6*2 = 1.2): dijawab 6, dikoreksi 4
	// (2.0, 1.5, 0.5, 1.2), belum dikoreksi 2 -> benar 3/4, P = 0.75.
	jawabans := []jawabanInput{
		// Soal 1
		{AttemptID: "A", SoalID: "soal1", Jawaban: jwb("A"), Poin: fptr(1)},
		{AttemptID: "B", SoalID: "soal1", Jawaban: jwb("A"), Poin: fptr(1)},
		{AttemptID: "C", SoalID: "soal1", Jawaban: jwb("A"), Poin: fptr(1)},
		{AttemptID: "D", SoalID: "soal1", Jawaban: jwb("A"), Poin: fptr(1)},
		{AttemptID: "E", SoalID: "soal1", Jawaban: jwb("B"), Poin: fptr(0)},
		{AttemptID: "F", SoalID: "soal1", Jawaban: jwb("A"), Poin: fptr(1)},
		{AttemptID: "G", SoalID: "soal1", Jawaban: jwb("C"), Poin: fptr(0)},
		{AttemptID: "H", SoalID: "soal1", Jawaban: jwb("B"), Poin: fptr(0)},
		{AttemptID: "I", SoalID: "soal1", Jawaban: jwb("A"), Poin: fptr(1)},
		{AttemptID: "J", SoalID: "soal1", Jawaban: jwb("C"), Poin: fptr(0)},
		// Soal 2
		{AttemptID: "A", SoalID: "soal2", Jawaban: jwb("B"), Poin: fptr(0)},
		{AttemptID: "B", SoalID: "soal2", Jawaban: jwb("B"), Poin: fptr(0)},
		{AttemptID: "C", SoalID: "soal2", Jawaban: jwb("B"), Poin: fptr(0)},
		{AttemptID: "D", SoalID: "soal2", Jawaban: jwb("A"), Poin: fptr(1)},
		{AttemptID: "E", SoalID: "soal2", Jawaban: jwb("B"), Poin: fptr(0)},
		{AttemptID: "F", SoalID: "soal2", Jawaban: jwb("B"), Poin: fptr(0)},
		{AttemptID: "G", SoalID: "soal2", Jawaban: jwb("A"), Poin: fptr(1)},
		{AttemptID: "H", SoalID: "soal2", Jawaban: jwb("A"), Poin: fptr(1)},
		{AttemptID: "I", SoalID: "soal2", Jawaban: jwb("A"), Poin: fptr(1)},
		{AttemptID: "J", SoalID: "soal2", Jawaban: jwb("A"), Poin: fptr(1)},
		// Soal 3 (uraian)
		{AttemptID: "A", SoalID: "soal3", Jawaban: jwb("teks"), Poin: fptr(2.0)},
		{AttemptID: "B", SoalID: "soal3", Jawaban: jwb("teks"), Poin: nil},
		{AttemptID: "C", SoalID: "soal3", Jawaban: jwb("teks"), Poin: fptr(1.5)},
		{AttemptID: "H", SoalID: "soal3", Jawaban: jwb("teks"), Poin: nil},
		{AttemptID: "I", SoalID: "soal3", Jawaban: jwb("teks"), Poin: fptr(0.5)},
		{AttemptID: "J", SoalID: "soal3", Jawaban: jwb("teks"), Poin: fptr(1.2)},
	}

	hasil := computeAnalisis("jadwal-test", soals, attempts, jawabans)
	if hasil.JumlahPeserta != 10 {
		t.Fatalf("JumlahPeserta = %d, mau 10", hasil.JumlahPeserta)
	}
	if hasil.MetodeKelompok != "27%" {
		t.Fatalf("MetodeKelompok = %q, mau 27%%", hasil.MetodeKelompok)
	}
	if hasil.SampelKecil {
		t.Fatalf("SampelKecil harus false untuk N=10")
	}
	byID := map[string]analisisItem{}
	for _, it := range hasil.Items {
		byID[it.SoalID] = it
	}

	// Soal 1: P = 6/10 = 0.60 Sedang; D = (3-1)/3 = 0.67 Sangat Baik.
	s1 := byID["soal1"]
	if s1.JumlahPenjab != 10 || s1.JumlahBenar != 6 {
		t.Errorf("soal1 penjab/benar = %d/%d, mau 10/6", s1.JumlahPenjab, s1.JumlahBenar)
	}
	if s1.Kesukaran == nil || *s1.Kesukaran != 0.60 {
		t.Errorf("soal1 kesukaran = %v, mau 0.60", s1.Kesukaran)
	}
	if s1.LabelKesukaran != "Sedang" {
		t.Errorf("soal1 label = %q, mau Sedang", s1.LabelKesukaran)
	}
	if s1.DayaBeda == nil || *s1.DayaBeda != 0.67 {
		t.Errorf("soal1 daya beda = %v, mau 0.67", s1.DayaBeda)
	}
	if s1.InterpretasiDayaBeda != "Sangat Baik" {
		t.Errorf("soal1 interpretasi = %q, mau Sangat Baik", s1.InterpretasiDayaBeda)
	}
	wantPengecoh := map[string][3]any{"A": {6, 60.0, true}, "B": {2, 20.0, true}, "C": {2, 20.0, true}, "D": {0, 0.0, false}}
	for _, p := range s1.Pengecoh {
		w := wantPengecoh[p.Opsi]
		if p.Pemilih != w[0] || p.Persen != w[1] || p.Berfungsi != w[2] {
			t.Errorf("soal1 pengecoh %s = (%d, %v, %v), mau (%v, %v, %v)",
				p.Opsi, p.Pemilih, p.Persen, p.Berfungsi, w[0], w[1], w[2])
		}
	}

	// Soal 2: P = 5/10 = 0.50 Sedang; D = (0-3)/3 = -1.00 Sangat Buruk.
	s2 := byID["soal2"]
	if s2.Kesukaran == nil || *s2.Kesukaran != 0.50 || s2.LabelKesukaran != "Sedang" {
		t.Errorf("soal2 kesukaran = %v (%q), mau 0.50 Sedang", s2.Kesukaran, s2.LabelKesukaran)
	}
	if s2.DayaBeda == nil || *s2.DayaBeda != -1.00 {
		t.Errorf("soal2 daya beda = %v, mau -1.00", s2.DayaBeda)
	}
	if s2.InterpretasiDayaBeda != "Sangat Buruk" {
		t.Errorf("soal2 interpretasi = %q, mau Sangat Buruk", s2.InterpretasiDayaBeda)
	}

	// Soal 3 (uraian): penjawab = hanya yang sudah dikoreksi (4); benar 3;
	// P = 3/4 = 0.75 Mudah; belum dikoreksi 2; pengecoh N/A.
	// Daya beda: atas benar = A (2.0) + C (1.5) = 2 (B belum dikoreksi);
	// bawah benar = J (1.2) = 1 (H belum dikoreksi, I 0.5 salah) -> D = 0.33 Baik.
	s3 := byID["soal3"]
	if s3.JumlahPenjab != 4 || s3.JumlahBenar != 3 || s3.BelumDikoreksi != 2 {
		t.Errorf("soal3 penjab/benar/belum = %d/%d/%d, mau 4/3/2",
			s3.JumlahPenjab, s3.JumlahBenar, s3.BelumDikoreksi)
	}
	if s3.Kesukaran == nil || *s3.Kesukaran != 0.75 || s3.LabelKesukaran != "Mudah" {
		t.Errorf("soal3 kesukaran = %v (%q), mau 0.75 Mudah", s3.Kesukaran, s3.LabelKesukaran)
	}
	if s3.DayaBeda == nil || *s3.DayaBeda != 0.33 || s3.InterpretasiDayaBeda != "Baik" {
		t.Errorf("soal3 daya beda = %v (%q), mau 0.33 Baik", s3.DayaBeda, s3.InterpretasiDayaBeda)
	}
	if s3.PengecohStatus != "n/a" {
		t.Errorf("soal3 pengecoh_status = %q, mau n/a", s3.PengecohStatus)
	}
}

func TestComputeAnalisisSampelKecil(t *testing.T) {
	// N = 4 < 10 -> median split 50%, n = 2, catatan sampel kecil.
	attempts := []attemptInput{
		{ID: "A", Nilai: fptr(90)},
		{ID: "B", Nilai: fptr(80)},
		{ID: "C", Nilai: fptr(60)},
		{ID: "D", Nilai: fptr(50)},
	}
	soals := []soalInput{{ID: "s1", Pertanyaan: "Q?", Tipe: "pg", Bobot: 1,
		Opsi: []opsiSoal{{"A", "x"}, {"B", "y"}}}}
	jawabans := []jawabanInput{
		{AttemptID: "A", SoalID: "s1", Jawaban: jwb("A"), Poin: fptr(1)},
		{AttemptID: "B", SoalID: "s1", Jawaban: jwb("A"), Poin: fptr(1)},
		{AttemptID: "C", SoalID: "s1", Jawaban: jwb("B"), Poin: fptr(0)},
		{AttemptID: "D", SoalID: "s1", Jawaban: jwb("B"), Poin: fptr(0)},
	}
	hasil := computeAnalisis("j", soals, attempts, jawabans)
	if !hasil.SampelKecil || hasil.MetodeKelompok != "median split 50%" {
		t.Fatalf("metode/sampel = %q/%v, mau 'median split 50%%' dan true", hasil.MetodeKelompok, hasil.SampelKecil)
	}
	it := hasil.Items[0]
	// D = (2-0)/2 = 1.00 Sangat Baik; P = 2/4 = 0.50 Sedang.
	if it.DayaBeda == nil || *it.DayaBeda != 1.00 {
		t.Errorf("daya beda = %v, mau 1.00", it.DayaBeda)
	}
	if it.Kesukaran == nil || *it.Kesukaran != 0.50 {
		t.Errorf("kesukaran = %v, mau 0.50", it.Kesukaran)
	}
}

func TestComputeKompetensi(t *testing.T) {
	k1 := "kompetensi-1"
	k2 := "kompetensi-2"
	soals := []soalInput{
		{ID: "s1", Pertanyaan: "Q1", Tipe: "pg", Bobot: 1, KompetensiID: &k1, KompetensiKode: "K1", KompetensiNama: "Kompetensi 1"},
		{ID: "s2", Pertanyaan: "Q2", Tipe: "pg", Bobot: 1, KompetensiID: &k1, KompetensiKode: "K1", KompetensiNama: "Kompetensi 1"},
		{ID: "s3", Pertanyaan: "Q3", Tipe: "pg", Bobot: 2, KompetensiID: &k2, KompetensiKode: "K2", KompetensiNama: "Kompetensi 2"},
	}
	attempts := []attemptInput{{ID: "att1", SiswaID: "sw1"}, {ID: "att2", SiswaID: "sw2"}}
	jawabans := []jawabanInput{
		// Siswa 1: K1: s1 benar (1), s2 tidak dijawab -> 1/(1+1) = 50.
		// K2: s3 benar (2) -> 2/2 = 100. Rata-rata: 3/(1+1+2) = 75.
		{AttemptID: "att1", SoalID: "s1", Jawaban: jwb("A"), Poin: fptr(1)},
		{AttemptID: "att1", SoalID: "s3", Jawaban: jwb("A"), Poin: fptr(2)},
		// Siswa 2: K1: tidak menjawab apa pun -> null + "tidak ada soal".
		// K2: dijawab tapi belum dikoreksi -> null + "belum dikoreksi".
		{AttemptID: "att2", SoalID: "s3", Jawaban: jwb("teks"), Poin: nil},
	}
	siswaByAttempt := map[string]siswaInput{
		"att1": {AttemptID: "att1", SiswaID: "sw1", NISN: "001", Nama: "Andi", Kelas: "XI RPL 1"},
		"att2": {AttemptID: "att2", SiswaID: "sw2", NISN: "002", Nama: "Budi", Kelas: "XI RPL 1"},
	}
	// Soal s3 dipakai dua peran di sini (PG untuk att1, manual-uncorrected
	// untuk att2) — untuk menguji cabang belum-dikoreksi, s3 harus
	// bertipe manual. Ubah tipe s3 menjadi uraian agar cabang manual aktif.
	soals[2].Tipe = "uraian"

	hasil := computeKompetensi("j", soals, attempts, jawabans, siswaByAttempt)
	if hasil.JumlahSiswa != 2 {
		t.Fatalf("JumlahSiswa = %d, mau 2", hasil.JumlahSiswa)
	}
	if hasil.RataRataKelas == nil || *hasil.RataRataKelas != 75 {
		t.Errorf("RataRataKelas = %v, mau 75", hasil.RataRataKelas)
	}
	if len(hasil.RataRataPerKompetensi) != 2 {
		t.Fatalf("agregat kompetensi = %d baris, mau 2", len(hasil.RataRataPerKompetensi))
	}
	// Agregat: K1 rata-rata 50 (N=1), K2 rata-rata 100 (N=1).
	if hasil.RataRataPerKompetensi[0].RataRata == nil || *hasil.RataRataPerKompetensi[0].RataRata != 50 ||
		hasil.RataRataPerKompetensi[0].JumlahSiswa != 1 {
		t.Errorf("agregat K1 = %v (N=%d), mau 50 (N=1)",
			hasil.RataRataPerKompetensi[0].RataRata, hasil.RataRataPerKompetensi[0].JumlahSiswa)
	}
	if hasil.RataRataPerKompetensi[1].RataRata == nil || *hasil.RataRataPerKompetensi[1].RataRata != 100 {
		t.Errorf("agregat K2 = %v, mau 100", hasil.RataRataPerKompetensi[1].RataRata)
	}
	byNama := map[string]kompetensiSiswa{}
	for _, s := range hasil.Siswa {
		byNama[s.Nama] = s
	}
	andi := byNama["Andi"]
	if andi.RataRata == nil || *andi.RataRata != 75 {
		t.Errorf("Andi rata-rata = %v, mau 75", andi.RataRata)
	}
	if andi.NilaiPerKompetensi[0].Nilai == nil || *andi.NilaiPerKompetensi[0].Nilai != 50 {
		t.Errorf("Andi K1 = %v, mau 50", andi.NilaiPerKompetensi[0].Nilai)
	}
	if andi.NilaiPerKompetensi[1].Nilai == nil || *andi.NilaiPerKompetensi[1].Nilai != 100 {
		t.Errorf("Andi K2 = %v, mau 100", andi.NilaiPerKompetensi[1].Nilai)
	}
	budi := byNama["Budi"]
	if budi.RataRata != nil {
		t.Errorf("Budi rata-rata = %v, mau null (tidak ada soal ternilai)", budi.RataRata)
	}
	if budi.NilaiPerKompetensi[0].Nilai != nil || !budi.NilaiPerKompetensi[0].TidakAdaSoal ||
		budi.NilaiPerKompetensi[0].Flag != "tidak ada soal" {
		t.Errorf("Budi K1 = %v (flag %q), mau null + 'tidak ada soal'",
			budi.NilaiPerKompetensi[0].Nilai, budi.NilaiPerKompetensi[0].Flag)
	}
	if budi.NilaiPerKompetensi[1].Nilai != nil || budi.NilaiPerKompetensi[1].Flag != "belum dikoreksi" {
		t.Errorf("Budi K2 = %v (flag %q), mau null + 'belum dikoreksi'",
			budi.NilaiPerKompetensi[1].Nilai, budi.NilaiPerKompetensi[1].Flag)
	}
}
