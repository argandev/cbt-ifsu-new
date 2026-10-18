// Package platform — helper teks bersama (dipindah dari banksoal/grading/rekap/ukk/ukkx
// agar satu definisi; perilaku IDENTIK).
package platform

import "strings"

// NormalisasiIsian menyeragamkan jawaban isian/jawaban singkat:
// trim + lowercase + rapatkan spasi. Dipakai validasi kunci dan penilaian.
func NormalisasiIsian(s string) string {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(s)))
	return strings.Join(fields, " ")
}

// Predikat mengembalikan predikat kelulusan skala 0-100:
// <70 Belum Tuntas, 70-79 Cukup, 80-90 Baik, 91-100 Sangat Baik.
func Predikat(nilai float64) string {
	switch {
	case nilai < 70:
		return "Belum Tuntas"
	case nilai < 80:
		return "Cukup"
	case nilai <= 90:
		return "Baik"
	default:
		return "Sangat Baik"
	}
}
