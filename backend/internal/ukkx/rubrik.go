// Package ukkx adalah paket pendukung UKK (cross-check + siap di-wire
// tim integrasi) untuk CBT SMK v2.
//
// Struktur layer (hasil refactor arsitektur):
//   - rubrik.go:     logika rubrik UKK (file ini)
//   - sertifikat.go: nomor sertifikat + syarat terbit
//   - upload.go:     validasi berkas + path aman storage
//
// Paket ini berisi fungsi MURNI tanpa akses database, tanpa HTTP, dan
// tanpa dependensi di luar standard library, sehingga bisa diuji
// independen dari handler UKK utama (paket ukk).
//
// Aturan UKK FINAL yang diimplementasikan di sini (satu format rubrik +
// satu format sertifikat untuk semua jurusan):
//
//   - Bobot nilai akhir: teori 40% + praktik 60%.
//   - Rubrik praktik: Proses 30% / Produk 50% / Sikap 20%.
//   - Lulus bila nilai akhir >= 70.
//   - Predikat: <70 Belum Tuntas, 70-79 Cukup, 80-90 Baik, 91-100 Sangat Baik.
//   - Sertifikat HANYA bagi yang lulus.
//   - Nomor sertifikat: UKK/<tahun>/<kode-jurusan>/<nomor-urut 4 digit>.
//
// Semua nilai dibulatkan ke 2 desimal (round half-up). Nilai dihitung
// backend — frontend TIDAK PERNAH menghitung nilai (KONTRAK §0).
package ukkx

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
)

// Bobot rubrik praktik dan nilai akhir UKK (FINAL, satu format semua jurusan).
const (
	BobotProses = 0.30
	BobotProduk = 0.50
	BobotSikap  = 0.20

	BobotTeori   = 0.40
	BobotPraktik = 0.60

	// AmbangLulus adalah nilai akhir minimum agar dinyatakan lulus UKK.
	AmbangLulus = 70.0
)

// Predikat nilai akhir UKK.
const (
	PredikatBelumTuntas = "Belum Tuntas"
	PredikatCukup       = "Cukup"
	PredikatBaik        = "Baik"
	PredikatSangatBaik  = "Sangat Baik"
)

// bulat2 membulatkan v ke 2 desimal, round half-up pada representasi
// desimalnya.
//
// Pembulatan TIDAK dilakukan langsung pada float64 mentah (mis.
// math.Round(v*100)/100), karena 3.35*0.3 tersimpan sebagai
// 1.0049999999999999 dalam biner dan akan salah dibulatkan ke 1.00.
// Sebagai gantinya, v diubah dulu ke representasi desimal terpendek
// yang round-trip ke float yang sama ("1.005"), lalu dibulatkan
// half-up memakai aritmetika rasional eksak (math/big.Rat).
// Untuk nilai >= 0 (seluruh domain nilai UKK) ini adalah half-up
// murni; untuk nilai negatif berlaku half away from zero.
func bulat2(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return v
	}
	r, ok := new(big.Rat).SetString(strconv.FormatFloat(v, 'f', -1, 64))
	if !ok {
		return math.Round(v*100) / 100
	}
	r.Mul(r, big.NewRat(100, 1)) // skala ke satuan sen
	negatif := r.Sign() < 0
	if negatif {
		r.Neg(r)
	}
	r.Add(r, big.NewRat(1, 2)) // +0.5 lalu floor = half-up
	bulat := new(big.Int).Quo(r.Num(), r.Denom())
	hasil := new(big.Rat).SetFrac(bulat, big.NewInt(100))
	if negatif {
		hasil.Neg(hasil)
	}
	f, _ := hasil.Float64()
	return f
}

// HitungPraktik menghitung nilai praktik dari tiga komponen rubrik:
//
//	praktik = proses*0.30 + produk*0.50 + sikap*0.20
//
// Hasil dibulatkan ke 2 desimal (round half-up). Pemanggil wajib
// memvalidasi tiap komponen dengan ValidasiNilai sebelum memanggil;
// fungsi ini sendiri tidak mengembalikan error agar tetap murni dan
// mudah dipakai sebagai cross-check terhadap handler UKK utama.
func HitungPraktik(proses, produk, sikap float64) float64 {
	return bulat2(proses*BobotProses + produk*BobotProduk + sikap*BobotSikap)
}

// HitungAkhir menghitung nilai akhir UKK:
//
//	akhir = teori*0.40 + praktik*0.60
//
// Hasil dibulatkan ke 2 desimal (round half-up). Parameter praktik
// adalah nilai praktik yang SUDAH dihitung (mis. dari HitungPraktik).
func HitungAkhir(teori, praktik float64) float64 {
	return bulat2(teori*BobotTeori + praktik*BobotPraktik)
}

// Lulus melaporkan apakah nilai akhir memenuhi ambang kelulusan (>= 70).
// NaN dan +Inf tidak pernah lulus (perbandingan dengan NaN bernilai false,
// +Inf ditolak secara eksplisit agar tidak lolos lewat jalur aneh).
func Lulus(nilai float64) bool {
	if math.IsNaN(nilai) || math.IsInf(nilai, 1) {
		return false
	}
	return nilai >= AmbangLulus
}

// Predikat mengembalikan predikat untuk nilai akhir:
//
//	<70        -> "Belum Tuntas"
//	70-79.99.. -> "Cukup"       (batas atas inklusif 79.99..; 80 masuk "Baik")
//	80-90.99.. -> "Baik"        (91 masuk "Sangat Baik")
//	91-100     -> "Sangat Baik"
//
// Batas dievaluasi pada nilai APA ADANYA (tanpa pembulatan ulang), karena
// pemanggil sudah menerima nilai yang dibulatkan 2 desimal dari
// HitungAkhir. Nilai NaN diperlakukan sebagai Belum Tuntas (input tidak
// valid tidak boleh menghasilkan predikat lulus). Nilai > 100 (di luar
// rentang valid) dipetakan ke "Sangat Baik"; pemanggil wajib menolak
// nilai semacam itu lewat ValidasiNilai terlebih dahulu.
func Predikat(nilai float64) string {
	if math.IsNaN(nilai) || nilai < AmbangLulus {
		return PredikatBelumTuntas
	}
	if nilai < 80 {
		return PredikatCukup
	}
	if nilai < 91 {
		return PredikatBaik
	}
	return PredikatSangatBaik
}

// ValidasiNilai menolak nilai di luar rentang 0-100 serta NaN dan Inf.
// Nilai -0 diterima (sama dengan 0).
func ValidasiNilai(v float64) error {
	if math.IsNaN(v) {
		return errors.New("nilai tidak boleh NaN")
	}
	if math.IsInf(v, 0) {
		return errors.New("nilai tidak boleh tak hingga")
	}
	if v < 0 || v > 100 {
		return fmt.Errorf("nilai %.2f di luar rentang 0-100", v)
	}
	return nil
}
