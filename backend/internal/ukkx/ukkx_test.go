package ukkx

import (
	"math"
	"strings"
	"testing"
)

// ---------- HitungPraktik ----------

func TestHitungPraktik(t *testing.T) {
	tests := []struct {
		nama           string
		proses, produk float64
		sikap          float64
		ingin          float64
	}{
		{"contoh: proses 75, produk 85, sikap 90 -> 83.0", 75, 85, 90, 83.0},
		{"semua nol", 0, 0, 0, 0},
		{"semua 100", 100, 100, 100, 100},
		{"hanya proses", 100, 0, 0, 30},
		{"hanya produk", 0, 100, 0, 50},
		{"hanya sikap", 0, 0, 100, 20},
		// 10*0.3 + 20*0.5 + 30*0.2 = 3 + 10 + 6 = 19
		{"campuran", 10, 20, 30, 19},
		// Pembulatan half-up: 33.33*0.3 + 33.33*0.5 + 33.33*0.2 = 33.33
		{"pembulatan", 33.33, 33.33, 33.33, 33.33},
	}
	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			if got := HitungPraktik(tc.proses, tc.produk, tc.sikap); got != tc.ingin {
				t.Errorf("HitungPraktik(%v, %v, %v) = %v, ingin %v", tc.proses, tc.produk, tc.sikap, got, tc.ingin)
			}
		})
	}
}

func TestHitungPraktik_PembulatanHalfUp(t *testing.T) {
	// Nilai yang mentah-nya 10.005 harus dibulatkan ke 10.01 (half-up),
	// bukan 10.00 (banker's rounding). 33.35*0.3 = 10.005.
	// Catatan: 33.35 tidak persis dalam biner, jadi uji via input yang
	// menghasilkan representasi biner di atas ambang .005:
	// 0.125*... -> pakai kasus langsung: proses=33.35, produk=0, sikap=0.
	// Hasil mentah = 10.004999... dalam float64? Verifikasi perilaku
	// pembulatan lewat nilai yang terbukti: 2.345 -> 2.35 tidak bisa
	// diuji langsung karena HitungPraktik memakai bobot. Uji bobot
	// yang menghasilkan 1.005 persis-ish: proses=3.35 -> 3.35*0.3.
	got := HitungPraktik(3.35, 0, 0) // 3.35*0.3 = 1.005 (float) -> 1.01 half-up
	if got != 1.01 {
		t.Errorf("HitungPraktik(3.35, 0, 0) = %v, ingin 1.01 (half-up)", got)
	}
}

// ---------- HitungAkhir ----------

func TestHitungAkhir(t *testing.T) {
	tests := []struct {
		nama           string
		teori, praktik float64
		ingin          float64
	}{
		{"contoh: teori 80, praktik 83.0 -> 81.8", 80, 83.0, 81.8},
		{"semua nol", 0, 0, 0},
		{"semua 100", 100, 100, 100},
		{"hanya teori", 100, 0, 40},
		{"hanya praktik", 0, 100, 60},
		{"teori 70 praktik 70 -> 70", 70, 70, 70},
		{"batas lulus dari bawah: 69.98", 69.95, 70, 69.98},
	}
	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			if got := HitungAkhir(tc.teori, tc.praktik); got != tc.ingin {
				t.Errorf("HitungAkhir(%v, %v) = %v, ingin %v", tc.teori, tc.praktik, got, tc.ingin)
			}
		})
	}
}

func TestHitungAkhir_RantaiPenuh(t *testing.T) {
	// Alur lengkap: teori 80, proses 75, produk 85, sikap 90.
	// praktik = 75*0.3 + 85*0.5 + 90*0.2 = 22.5 + 42.5 + 18 = 83.0
	// akhir   = 80*0.4 + 83.0*0.6 = 32 + 49.8 = 81.8 -> lulus, Baik.
	praktik := HitungPraktik(75, 85, 90)
	if praktik != 83.0 {
		t.Fatalf("praktik = %v, ingin 83.0", praktik)
	}
	akhir := HitungAkhir(80, praktik)
	if akhir != 81.8 {
		t.Fatalf("akhir = %v, ingin 81.8", akhir)
	}
	if !Lulus(akhir) {
		t.Errorf("akhir %v seharusnya lulus", akhir)
	}
	if got := Predikat(akhir); got != PredikatBaik {
		t.Errorf("predikat = %q, ingin %q", got, PredikatBaik)
	}
}

// ---------- Lulus ----------

func TestLulus_Batas(t *testing.T) {
	tests := []struct {
		nilai float64
		ingin bool
	}{
		{0, false},
		{69.99, false}, // kasus batas: 69.99 TIDAK lulus
		{69.999, false},
		{70, true}, // kasus batas: tepat 70 lulus
		{70.01, true},
		{100, true},
		{math.NaN(), false},
		{math.Inf(1), false},
		{math.Inf(-1), false},
	}
	for _, tc := range tests {
		if got := Lulus(tc.nilai); got != tc.ingin {
			t.Errorf("Lulus(%v) = %v, ingin %v", tc.nilai, got, tc.ingin)
		}
	}
}

// ---------- Predikat ----------

func TestPredikat_Batas(t *testing.T) {
	tests := []struct {
		nilai float64
		ingin string
	}{
		{0, PredikatBelumTuntas},
		{69.99, PredikatBelumTuntas},
		{70, PredikatCukup},      // batas bawah Cukup
		{79.99, PredikatCukup},   // batas atas Cukup
		{80, PredikatBaik},       // batas bawah Baik
		{90.99, PredikatBaik},    // batas atas Baik
		{91, PredikatSangatBaik}, // batas bawah Sangat Baik
		{100, PredikatSangatBaik},
		{math.NaN(), PredikatBelumTuntas}, // NaN tidak boleh jadi predikat lulus
	}
	for _, tc := range tests {
		if got := Predikat(tc.nilai); got != tc.ingin {
			t.Errorf("Predikat(%v) = %q, ingin %q", tc.nilai, got, tc.ingin)
		}
	}
}

// ---------- ValidasiNilai ----------

func TestValidasiNilai(t *testing.T) {
	valid := []float64{0, 0.01, 50, 69.99, 70, 99.99, 100, -0.0}
	for _, v := range valid {
		if err := ValidasiNilai(v); err != nil {
			t.Errorf("ValidasiNilai(%v) = %v, ingin nil", v, err)
		}
	}
	invalid := []float64{-0.01, -1, 100.01, 101, math.NaN(), math.Inf(1), math.Inf(-1)}
	for _, v := range invalid {
		if err := ValidasiNilai(v); err == nil {
			t.Errorf("ValidasiNilai(%v) = nil, ingin error", v)
		}
	}
}

// ---------- NomorSertifikat ----------

func TestNomorSertifikat(t *testing.T) {
	tests := []struct {
		nama  string
		tahun int
		kode  string
		urut  int
		ingin string
	}{
		{"urut 1", 2026, "RPL", 1, "UKK/2026/RPL/0001"},
		{"urut 9999", 2026, "RPL", 9999, "UKK/2026/RPL/9999"},
		{"urut 42 zero-pad", 2026, "TKJ", 42, "UKK/2026/TKJ/0042"},
		{"kode huruf kecil dinormalisasi", 2026, "rpl", 1, "UKK/2026/RPL/0001"},
		{"kode 2 huruf", 2025, "MM", 7, "UKK/2025/MM/0007"},
		{"kode 10 karakter", 2026, "ABCDEFGHIJ", 1, "UKK/2026/ABCDEFGHIJ/0001"},
		{"tahun bawah", 2000, "RPL", 1, "UKK/2000/RPL/0001"},
		{"tahun atas", 2100, "RPL", 1, "UKK/2100/RPL/0001"},
		// Input tidak valid -> string kosong.
		{"tahun 1999", 1999, "RPL", 1, ""},
		{"tahun 2101", 2101, "RPL", 1, ""},
		{"kode 1 huruf", 2026, "R", 1, ""},
		{"kode 11 karakter", 2026, "ABCDEFGHIJK", 1, ""},
		{"kode ada strip", 2026, "RP-L", 1, ""},
		{"kode ada spasi", 2026, "RP L", 1, ""},
		{"kode kosong", 2026, "", 1, ""},
		{"urut 0", 2026, "RPL", 0, ""},
		{"urut 10000", 2026, "RPL", 10000, ""},
		{"urut negatif", 2026, "RPL", -1, ""},
	}
	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			if got := NomorSertifikat(tc.tahun, tc.kode, tc.urut); got != tc.ingin {
				t.Errorf("NomorSertifikat(%d, %q, %d) = %q, ingin %q", tc.tahun, tc.kode, tc.urut, got, tc.ingin)
			}
		})
	}
}

// ---------- ParseNomorSertifikat ----------

func TestParseNomorSertifikat(t *testing.T) {
	tahun, kode, urut, err := ParseNomorSertifikat("UKK/2026/RPL/0001")
	if err != nil || tahun != 2026 || kode != "RPL" || urut != 1 {
		t.Errorf("Parse(UKK/2026/RPL/0001) = (%d, %q, %d, %v)", tahun, kode, urut, err)
	}
	// Round-trip: NomorSertifikat -> Parse harus kembali ke komponen awal.
	for _, urutAwal := range []int{1, 42, 9999} {
		nomor := NomorSertifikat(2026, "TKJ", urutAwal)
		th, kd, u, err := ParseNomorSertifikat(nomor)
		if err != nil || th != 2026 || kd != "TKJ" || u != urutAwal {
			t.Errorf("round-trip %q = (%d, %q, %d, %v)", nomor, th, kd, u, err)
		}
	}

	invalid := []string{
		"",
		"UKK/2026/RPL/1",      // urut tidak 4 digit
		"UKK/2026/RPL/0000",   // urut 0
		"UKK/2026/RPL/10000",  // urut 5 digit
		"UKK/26/RPL/0001",     // tahun tidak 4 digit
		"UKK/1999/RPL/0001",   // tahun di luar rentang
		"UKK/2026/R/0001",     // kode 1 huruf
		"UKK/2026/RP-L/0001",  // kode ada karakter non-alfanumerik
		"UKK/2026/RPL/000A",   // urut bukan angka
		"UKK/2026/RPL",        // kurang segmen
		"XXX/2026/RPL/0001",   // prefix salah
		"UKK/2026/RPL/0001/x", // segmen berlebih
		"UKK/2026/RPL/0001 ",  // spasi trailing
	}
	for _, s := range invalid {
		if _, _, _, err := ParseNomorSertifikat(s); err == nil {
			t.Errorf("ParseNomorSertifikat(%q) = nil, ingin error", s)
		}
	}
}

// ---------- BolehTerbit ----------

func TestBolehTerbit(t *testing.T) {
	tests := []struct {
		nama  string
		nilai float64
		lulus bool
		ingin bool
	}{
		{"lulus + 82.1", 82.1, true, true},
		{"lulus + tepat 70", 70, true, true},
		{"lulus + 69.99", 69.99, true, false},
		{"tidak lulus + 82.1 (nilai belum final)", 82.1, false, false},
		{"tidak lulus + 50", 50, false, false},
		{"lulus + NaN", math.NaN(), true, false},
	}
	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			if got := BolehTerbit(tc.nilai, tc.lulus); got != tc.ingin {
				t.Errorf("BolehTerbit(%v, %v) = %v, ingin %v", tc.nilai, tc.lulus, got, tc.ingin)
			}
		})
	}
}

// ---------- ValidasiBerkas ----------

func TestValidasiBerkas_OK(t *testing.T) {
	boleh := []struct {
		nama   string
		ukuran int64
	}{
		{"laporan.pdf", 1024},
		{"tugas.DOCX", 1024}, // ekstensi case-insensitive
		{"dokumen.Doc", 0},   // file kosong diizinkan di lapisan ini
		{"foto.JPG", 1024},
		{"foto.jpeg", 1024},
		{"gambar.png", 1024},
		{"arsip.zip", 1024},
		{"laporan.final.pdf", 1024}, // titik ganda, ekstensi terakhir yang dipakai
		{"tepat-10mb.pdf", UkuranMaksBerkasUKK},
		{"nama ada spasi.pdf", 1024},
		{"nama-dengan_strip.dan.titik.pdf", 1024},
	}
	for _, tc := range boleh {
		if err := ValidasiBerkas(tc.nama, tc.ukuran); err != nil {
			t.Errorf("ValidasiBerkas(%q, %d) = %v, ingin nil", tc.nama, tc.ukuran, err)
		}
	}
}

func TestValidasiBerkas_Ditolak(t *testing.T) {
	ditolak := []struct {
		nama   string
		ukuran int64
		alasan string
	}{
		{"../../etc/passwd", 100, "traversal + tanpa ekstensi diizinkan"},
		{"..\\..\\windows\\x.pdf", 100, "backslash"},
		{"folder/file.pdf", 100, "slash"},
		{"file\x00.pdf", 100, "karakter NUL/kontrol"},
		{"file\n.pdf", 100, "karakter kontrol newline"},
		{"..", 100, `".."`},
		{"", 100, "nama kosong"},
		{"tanpa-ekstensi", 100, "tanpa ekstensi"},
		{"file.", 100, "titik trailing tanpa ekstensi"},
		{"program.exe", 100, "ekstensi .exe"},
		{"script.sh", 100, "ekstensi .sh"},
		{"page.html", 100, "ekstensi .html"},
		{"app.js", 100, "ekstensi .js"},
		{"shell.php", 100, "ekstensi .php"},
		{"macro.xlsm", 100, "ekstensi .xlsm (macro)"},
		{"laporan.pdf.exe", 100, "ekstensi ganda berakhir .exe"},
		{"besar.pdf", UkuranMaksBerkasUKK + 1, "10 MB + 1 byte"},
		{"besar.pdf", UkuranMaksBerkasUKK * 2, "20 MB"},
		{"negatif.pdf", -1, "ukuran negatif"},
	}
	for _, tc := range ditolak {
		t.Run(tc.alasan+": "+tc.nama, func(t *testing.T) {
			if err := ValidasiBerkas(tc.nama, tc.ukuran); err == nil {
				t.Errorf("ValidasiBerkas(%q, %d) = nil, ingin error (%s)", tc.nama, tc.ukuran, tc.alasan)
			}
		})
	}
}

// ---------- PathAman ----------

func TestPathAman(t *testing.T) {
	tests := []struct {
		nama           string
		ukkID, siswaID string
		namaFile       string
		ingin          string
	}{
		{"normal", "ukk-1", "siswa-1", "laporan.pdf", "storage/ukk/ukk-1/siswa-1/laporan.pdf"},
		{"traversal nama file", "ukk-1", "siswa-1", "../../etc/passwd", "storage/ukk/ukk-1/siswa-1/___.._etc_passwd"},
		{"backslash", "ukk-1", "siswa-1", "..\\..\\x.pdf", "storage/ukk/ukk-1/siswa-1/___.._x.pdf"},
		{"spasi dan karakter aneh", "ukk-1", "siswa-1", "laporan akhir (1).pdf", "storage/ukk/ukk-1/siswa-1/laporan_akhir__1_.pdf"},
		{"traversal via ID", "../../etc", "..", "a.pdf", "storage/ukk/___.._etc/__/a.pdf"},
		{"nama kosong", "ukk-1", "siswa-1", "", "storage/ukk/ukk-1/siswa-1/_"},
		{"nama hanya titik", "ukk-1", "siswa-1", "...", "storage/ukk/ukk-1/siswa-1/___"},
	}
	for _, tc := range tests {
		t.Run(tc.nama, func(t *testing.T) {
			if got := PathAman(tc.ukkID, tc.siswaID, tc.namaFile); got != tc.ingin {
				t.Errorf("PathAman(%q, %q, %q) = %q, ingin %q", tc.ukkID, tc.siswaID, tc.namaFile, got, tc.ingin)
			}
		})
	}
}

func TestPathAman_TidakPernahKeluarDariStorageUKK(t *testing.T) {
	// Properti: output selalu berawalan "storage/ukk/" dan tidak
	// mengandung segmen ".." sebagai komponen path.
	jahat := []string{"../../etc/passwd", "..\\..\\x", "/abs/path.pdf", "....//....//x.pdf"}
	for _, nama := range jahat {
		got := PathAman("u1", "s1", nama)
		if !strings.HasPrefix(got, "storage/ukk/") {
			t.Errorf("PathAman(%q) = %q, tidak berawalan storage/ukk/", nama, got)
		}
		for _, seg := range strings.Split(got, "/") {
			if seg == ".." {
				t.Errorf("PathAman(%q) = %q, mengandung segmen ..", nama, got)
			}
		}
	}
}
