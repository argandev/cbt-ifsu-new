// Package analisisx menyediakan fungsi statistik MURNI (tanpa DB, tanpa
// HTTP, tanpa dependensi paket lain) untuk analisis butir soal dan nilai
// per kompetensi CBT SMK v2.
//
// Paket ini sengaja independen dari backend/internal/analisis (milik Tim
// Analisis utama): tim utama memakai paket ini — atau tidak — lewat
// pemanggilan fungsi biasa, dan halaman Analisis memanggil komponen
// visualisasi di frontend/src/components/analisis/ yang menampilkan
// hasil perhitungan backend.
//
// Definisi yang dipakai adalah definisi FINAL dari penugasan (mengikat,
// mengesampingkan ambang lama di ANALISIS.md Bab 8.4 bila berbeda):
//
//   - Kesukaran P = benar / penjawab; P < 0,3 Sukar, 0,3–0,7 Sedang,
//     P > 0,7 Mudah. Per soal dari attempt SELESAI pada satu jadwal.
//   - Daya beda D = (benar atas − benar bawah) / jumlah kelompok.
//     Kelompok atas/bawah = 27% peserta nilai tertinggi/terendah.
//     Bila peserta < 10: median split 50% + flag "sampel kecil".
//     Interpretasi: D >= 0,40 Sangat Baik; 0,30–0,39 Baik; 0,20–0,29
//     Cukup; 0,00–0,19 Buruk; D < 0,00 Sangat Buruk.
//   - Pengecoh: per opsi, jumlah pemilih + persen. Pengecoh "tidak
//     berfungsi" bila dipilih < 5% peserta. Soal uraian/manual = N/A
//     (bukan dihitung lewat fungsi ini).
//   - Nilai per kompetensi per siswa per jadwal:
//     Σ poin soal kompetensi itu / Σ bobot soal kompetensi itu × 100.
//     Bila siswa tidak menjawab satu pun soal kompetensi itu -> nilai
//     nil + flag "tidak ada soal".
//   - "Benar" untuk soal auto = poin == bobot (poin penuh); untuk soal
//     uraian/manual = poin >= 0,6 × bobot. Penentuan benar/salah ini
//     dilakukan pemanggil (yang tahu tipe soal); fungsi di sini hanya
//     menerima hitungan benar yang sudah ditentukan.
//
// Semua fungsi di sini deterministik dan bebas efek samping sehingga
// mudah diuji dan dipakai ulang.
package analisisx

import (
	"math"
	"sort"
	"strings"
)

// ambangKesukaranSukar adalah batas bawah kategori Sedang (inklusif).
const ambangKesukaranSukar = 0.3

// ambangKesukaranMudah adalah batas atas kategori Sedang (inklusif).
const ambangKesukaranMudah = 0.7

// ambangPengecohBerfungsi adalah persen minimal agar pengecoh dianggap
// berfungsi (5%). Pengecoh dipilih < 5% peserta = tidak berfungsi.
const ambangPengecohBerfungsi = 5.0

// ambangBenarUraian adalah fraksi bobot minimal agar jawaban soal
// uraian/manual dianggap benar (0,6 × bobot).
const ambangBenarUraian = 0.6

// AmbangBenarUraian mengembalikan fraksi bobot minimal agar jawaban soal
// uraian/manual dianggap benar (0,6). Diekspos agar pemanggil tidak
// meng-hardcode angka yang sama di tempat lain.
func AmbangBenarUraian() float64 { return ambangBenarUraian }

// Kesukaran menghitung tingkat kesukaran soal P = benar / penjawab
// (0–1) beserta labelnya: "Sukar" (P < 0,3), "Sedang" (0,3–0,7),
// "Mudah" (P > 0,7).
//
// Kasus batas: bila penjawab <= 0, tidak ada data untuk dihitung —
// mengembalikan (0, "tidak ada data"). Nilai benar dijepit ke
// [0, penjawab] agar P selalu dalam [0, 1] walau input tidak konsisten.
func Kesukaran(benar, penjawab int) (float64, string) {
	if penjawab <= 0 {
		return 0, "tidak ada data"
	}
	if benar < 0 {
		benar = 0
	}
	if benar > penjawab {
		benar = penjawab
	}
	p := float64(benar) / float64(penjawab)
	switch {
	case p < ambangKesukaranSukar:
		return p, "Sukar"
	case p > ambangKesukaranMudah:
		return p, "Mudah"
	default:
		return p, "Sedang"
	}
}

// DayaBeda menghitung daya beda D = (benarAtas − benarBawah) /
// jumlahKelompok beserta interpretasi dan catatannya.
//
// Interpretasi: D >= 0,40 "Sangat Baik"; 0,30–0,39 "Baik"; 0,20–0,29
// "Cukup"; 0,00–0,19 "Buruk"; D < 0,00 "Sangat Buruk".
//
// Catatan berisi peringatan yang perlu ditampilkan ke guru:
//   - bila sampelKecil (n < 10, kelompok dibentuk via median split):
//     "sampel kecil (n < 10): daya beda tidak dapat dipercaya";
//   - bila D negatif: "nilai negatif: kunci jawaban dicurigai".
//
// Bila tidak ada peringatan, catatan berupa string kosong.
//
// Kasus batas: bila jumlahKelompok <= 0, mengembalikan
// (0, "tidak dapat dihitung", "jumlah kelompok harus > 0").
func DayaBeda(benarAtas, benarBawah, jumlahKelompok int, sampelKecil bool) (float64, string, string) {
	if jumlahKelompok <= 0 {
		return 0, "tidak dapat dihitung", "jumlah kelompok harus > 0"
	}
	if benarAtas < 0 {
		benarAtas = 0
	}
	if benarBawah < 0 {
		benarBawah = 0
	}
	d := float64(benarAtas-benarBawah) / float64(jumlahKelompok)

	var interpretasi string
	switch {
	case d >= 0.40:
		interpretasi = "Sangat Baik"
	case d >= 0.30:
		interpretasi = "Baik"
	case d >= 0.20:
		interpretasi = "Cukup"
	case d >= 0.00:
		interpretasi = "Buruk"
	default:
		interpretasi = "Sangat Buruk"
	}

	var catatan []string
	if sampelKecil {
		catatan = append(catatan, "sampel kecil (n < 10): daya beda tidak dapat dipercaya")
	}
	if d < 0 {
		catatan = append(catatan, "nilai negatif: kunci jawaban dicurigai")
	}
	return d, interpretasi, strings.Join(catatan, "; ")
}

// BagiKelompok membagi peserta menjadi kelompok atas dan kelompok bawah
// dan mengembalikan INDEKS peserta (posisi dalam slice nilai, bukan
// nilainya) untuk masing-masing kelompok, terurut dari nilai tertinggi
// (atas) / terendah (bawah).
//
//   - n >= 10: masing-masing kelompok berisi pembulatan 27% × n peserta
//     (minimal 1), diambil dari ujung atas / bawah setelah peserta
//     diurutkan menurun berdasarkan nilai. Ini metode 27% klasik.
//   - n < 10 (sampel kecil): median split — kelompok atas = separuh atas,
//     kelompok bawah = separuh bawah; bila n ganjil, peserta median
//     tengah TIDAK masuk kelompok mana pun. Pemanggil menentukan flag
//     sampelKecil dari len(nilai) < 10 dan meneruskannya ke DayaBeda.
//   - n <= 1: kedua kelompok kosong (tidak ada pembanding yang sah).
//
// Pengurutan stabil: peserta dengan nilai sama mempertahankan urutan
// aslinya, sehingga hasil deterministik. Keterbatasan yang dicatat:
// peserta yang nilainya sama persis di batas kelompok dipisah secara
// arbitrer oleh urutan stabil ini — pemanggil yang butuh perlakuan
// khusus seri (ties) harus menanganinya sebelum memanggil.
func BagiKelompok(nilai []float64) (atas, bawah []int) {
	n := len(nilai)
	if n <= 1 {
		return nil, nil
	}

	// Indeks terurut menurun berdasarkan nilai; stabil demi determinisme.
	idx := make([]int, n)
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool {
		return nilai[idx[a]] > nilai[idx[b]]
	})

	var ukuran int
	if n < 10 {
		// Median split: separuh bawah dibulatkan ke bawah (n/2);
		// sisa tengah (bila n ganjil) tidak dipakai.
		ukuran = n / 2
	} else {
		// Metode 27%: pembulatan terdekat, minimal 1.
		ukuran = int(math.Round(0.27 * float64(n)))
		if ukuran < 1 {
			ukuran = 1
		}
	}
	if ukuran <= 0 {
		return nil, nil
	}
	if ukuran > n {
		ukuran = n
	}

	atas = append([]int(nil), idx[:ukuran]...)
	bawah = append([]int(nil), idx[n-ukuran:]...)
	// Kelompok bawah diurutkan dari nilai terendah ke yang lebih tinggi
	// agar konsisten dengan cara baca "kelompok bawah".
	for i, j := 0, len(bawah)-1; i < j; i, j = i+1, j-1 {
		bawah[i], bawah[j] = bawah[j], bawah[i]
	}
	return atas, bawah
}

// PengecohItem adalah hasil analisis satu opsi jawaban (pengecoh atau
// kunci) pada soal berbasis opsi.
//
// Catatan penamaan: tipe ini bernama PengecohItem (bukan Pengecoh)
// karena Go melarang tipe dan fungsi memakai nama yang sama dalam satu
// paket, sedangkan fungsi penghitungnya wajib bernama Pengecoh sesuai
// spesifikasi penugasan.
type PengecohItem struct {
	// Opsi adalah label opsi, mis. "A", "B", "C", "D".
	Opsi string
	// Pemilih adalah jumlah peserta yang memilih opsi ini.
	Pemilih int
	// Persen adalah Pemilih / totalPenjawab × 100.
	Persen float64
	// Berfungsi bernilai true bila Persen >= 5. Pengecoh dengan
	// Persen < 5% dianggap "tidak berfungsi". Untuk KUNCI jawaban,
	// field ini tetap dihitung dengan aturan yang sama tetapi tidak
	// boleh dibaca sebagai penilaian kunci — pemanggil membedakan kunci
	// vs pengecoh dari konteks soal.
	Berfungsi bool
}

// Pengecoh menghitung, untuk tiap opsi dalam pemilihPerOpsi, jumlah
// pemilih dan persentasenya terhadap totalPenjawab, serta apakah opsi
// itu berfungsi (dipilih >= 5% peserta).
//
// Hasil diurutkan berdasarkan label opsi (A, B, C, ...) agar
// deterministik — map Go tidak memiliki urutan. Jumlah pemilih negatif
// dijepit ke 0. Bila totalPenjawab <= 0, semua persen = 0 (dan semua
// opsi tidak berfungsi).
//
// Catatan: fungsi ini hanya untuk soal berbasis opsi (PG, PG Kompleks,
// Benar-Salah, Menjodohkan per pasangan). Soal uraian/manual tidak
// memiliki opsi sehingga pengecohnya N/A — pemanggil tidak memanggil
// fungsi ini untuk soal tersebut.
func Pengecoh(pemilihPerOpsi map[string]int, totalPenjawab int) []PengecohItem {
	opsi := make([]string, 0, len(pemilihPerOpsi))
	for k := range pemilihPerOpsi {
		opsi = append(opsi, k)
	}
	sort.Strings(opsi)

	hasil := make([]PengecohItem, 0, len(opsi))
	for _, o := range opsi {
		pemilih := pemilihPerOpsi[o]
		if pemilih < 0 {
			pemilih = 0
		}
		var persen float64
		if totalPenjawab > 0 {
			persen = float64(pemilih) / float64(totalPenjawab) * 100
		}
		hasil = append(hasil, PengecohItem{
			Opsi:      o,
			Pemilih:   pemilih,
			Persen:    persen,
			Berfungsi: persen >= ambangPengecohBerfungsi,
		})
	}
	return hasil
}

// NilaiKompetensi menghitung nilai per kompetensi untuk satu siswa pada
// satu jadwal: Σ poin soal kompetensi itu / Σ bobot soal kompetensi
// itu × 100, hanya dari soal yang dijawab (dijawab[i] == true).
//
// Ketiga slice dibaca sampai panjang terpendeknya; elemen di luar itu
// diabaikan. Soal yang tidak dijawab TIDAK menyumbang poin maupun bobot
// (soal kosong tidak menekan nilai kompetensi — sesuai definisi: siswa
// yang tidak menjawab soal kompetensi itu sama sekali mendapat null).
//
// Mengembalikan (nil, "tidak ada soal") bila tidak ada satu pun soal
// yang dijawab. Mengembalikan (nil, "bobot nol") bila ada soal yang
// dijawab tetapi total bobotnya 0 (pembagian tak terdefinisi).
// Bila berhasil, flag berupa string kosong.
func NilaiKompetensi(poinPerSoal []float64, bobotPerSoal []float64, dijawab []bool) (nilai *float64, flag string) {
	n := len(poinPerSoal)
	if len(bobotPerSoal) < n {
		n = len(bobotPerSoal)
	}
	if len(dijawab) < n {
		n = len(dijawab)
	}

	var totalPoin, totalBobot float64
	adaJawaban := false
	for i := 0; i < n; i++ {
		if !dijawab[i] {
			continue
		}
		adaJawaban = true
		totalPoin += poinPerSoal[i]
		totalBobot += bobotPerSoal[i]
	}

	if !adaJawaban {
		return nil, "tidak ada soal"
	}
	if totalBobot == 0 {
		return nil, "bobot nol"
	}
	hasil := totalPoin / totalBobot * 100
	return &hasil, ""
}
