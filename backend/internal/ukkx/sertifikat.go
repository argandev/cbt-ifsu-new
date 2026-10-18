package ukkx

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// NomorSertifikat membentuk nomor sertifikat UKK dengan format:
//
//	UKK/<tahun>/<kode-jurusan>/<nomor-urut 4 digit>
//
// Contoh: NomorSertifikat(2026, "RPL", 1) == "UKK/2026/RPL/0001".
//
// Validasi: tahun 2000-2100, kode jurusan alfanumerik 2-10 karakter,
// nomor urut 1-9999. Kode jurusan dinormalisasi ke huruf besar agar
// format kanonis (mis. "rpl" -> "RPL").
//
// Karena signature fungsi ini mengembalikan string saja (sesuai kontrak
// tugas), input yang tidak valid menghasilkan string kosong "".
// Pemanggil yang butuh pesan error dapat memakai ParseNomorSertifikat
// untuk validasi, atau memeriksa hasil "" sebelum menyimpan.
func NomorSertifikat(tahun int, kodeJurusan string, urut int) string {
	kode := strings.ToUpper(strings.TrimSpace(kodeJurusan))
	if err := validasiKomponenSertifikat(tahun, kode, urut); err != nil {
		return ""
	}
	return fmt.Sprintf("UKK/%d/%s/%04d", tahun, kode, urut)
}

// ParseNomorSertifikat mengurai nomor sertifikat UKK dan mengembalikan
// komponennya (tahun, kode jurusan, nomor urut). Nomor urut harus
// persis 4 digit (mis. "0001", "9999"); "1" atau "0000" ditolak.
// Kode jurusan dikembalikan apa adanya seperti tertulis di nomor
// (tanpa normalisasi huruf besar).
func ParseNomorSertifikat(s string) (tahun int, kode string, urut int, err error) {
	bagian := strings.Split(s, "/")
	if len(bagian) != 4 || bagian[0] != "UKK" {
		return 0, "", 0, fmt.Errorf("nomor sertifikat %q bukan format UKK/<tahun>/<kode-jurusan>/<nomor-urut>", s)
	}

	tahunStr, kodeStr, urutStr := bagian[1], bagian[2], bagian[3]

	if len(tahunStr) != 4 {
		return 0, "", 0, fmt.Errorf("tahun %q harus 4 digit", tahunStr)
	}
	tahun, err = strconv.Atoi(tahunStr)
	if err != nil {
		return 0, "", 0, fmt.Errorf("tahun %q bukan angka: %w", tahunStr, err)
	}

	if len(urutStr) != 4 {
		return 0, "", 0, fmt.Errorf("nomor urut %q harus 4 digit", urutStr)
	}
	urut, err = strconv.Atoi(urutStr)
	if err != nil {
		return 0, "", 0, fmt.Errorf("nomor urut %q bukan angka: %w", urutStr, err)
	}

	if err := validasiKomponenSertifikat(tahun, kodeStr, urut); err != nil {
		return 0, "", 0, err
	}
	return tahun, kodeStr, urut, nil
}

// BolehTerbit melaporkan apakah sertifikat boleh diterbitkan untuk
// peserta: HANYA bila lulus == true DAN nilai >= 70. Kedua syarat
// harus terpenuhi — nilai >= 70 saja tanpa flag lulus (mis. nilai
// belum difinalisasi) TIDAK cukup, begitu pula sebaliknya.
func BolehTerbit(nilai float64, lulus bool) bool {
	return lulus && nilai >= AmbangLulus
}

func validasiKomponenSertifikat(tahun int, kode string, urut int) error {
	if tahun < 2000 || tahun > 2100 {
		return fmt.Errorf("tahun %d di luar rentang 2000-2100", tahun)
	}
	if len(kode) < 2 || len(kode) > 10 {
		return fmt.Errorf("kode jurusan %q harus 2-10 karakter", kode)
	}
	for _, r := range kode {
		if !(r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9') {
			return fmt.Errorf("kode jurusan %q harus alfanumerik", kode)
		}
	}
	if urut < 1 || urut > 9999 {
		return errors.New("nomor urut harus 1-9999")
	}
	return nil
}
