package ukkx

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"unicode"
)

// UkuranMaksBerkasUKK adalah ukuran maksimum satu berkas upload UKK: 10 MB.
const UkuranMaksBerkasUKK = 10 * 1024 * 1024 // 10 MiB = 10485760 byte

// ekstensiBerkasUKK adalah whitelist ekstensi upload UKK (huruf kecil,
// tanpa titik). Daftar ini FINAL untuk paket pendukung ini; bila tim
// integrasi perlu jenis lain (mis. video demo), tambahkan lewat
// perubahan kontrak, bukan lewat pengecualian di handler.
var ekstensiBerkasUKK = map[string]bool{
	"pdf":  true,
	"doc":  true,
	"docx": true,
	"jpg":  true,
	"jpeg": true,
	"png":  true,
	"zip":  true,
}

// ValidasiBerkas memeriksa nama file dan ukuran berkas upload UKK:
//   - nama tidak boleh kosong;
//   - nama ditolak bila mengandung '/', '\', "..", atau karakter kontrol
//     (termasuk NUL) — cegah path traversal dan nama aneh;
//   - ekstensi (setelah titik terakhir, case-insensitive) harus ada di
//     whitelist: pdf/doc/docx/jpg/jpeg/png/zip;
//   - ukuran 0 sampai tepat 10 MB diizinkan; di atas 10 MB atau negatif ditolak.
//
// Validasi ini adalah lapisan nama + ukuran saja. Tim integrasi TETAP
// wajib menambahkan lapisan yang tidak bisa dilakukan fungsi murni ini:
// sniffing magic bytes (jangan percaya ekstensi), batas entri/ukuran
// hasil ekstrak untuk ZIP (anti zip bomb, tolak symlink), penyimpanan
// di luar webroot + serve via handler terautentikasi, dan kuota per
// peserta (lihat ANALISIS Bab upload).
func ValidasiBerkas(namaFile string, ukuran int64) error {
	if namaFile == "" {
		return errors.New("nama file tidak boleh kosong")
	}
	for _, r := range namaFile {
		if r == '/' || r == '\\' || unicode.IsControl(r) {
			return fmt.Errorf("nama file %q mengandung karakter terlarang", namaFile)
		}
	}
	if strings.Contains(namaFile, "..") {
		return fmt.Errorf("nama file %q mengandung \"..\"", namaFile)
	}

	ekstensi := ekstensiFile(namaFile)
	if ekstensi == "" {
		return fmt.Errorf("file %q tidak memiliki ekstensi", namaFile)
	}
	if !ekstensiBerkasUKK[ekstensi] {
		return fmt.Errorf("ekstensi .%s tidak diizinkan (yang diizinkan: pdf, doc, docx, jpg, jpeg, png, zip)", ekstensi)
	}

	if ukuran < 0 {
		return fmt.Errorf("ukuran file %d tidak valid", ukuran)
	}
	if ukuran > UkuranMaksBerkasUKK {
		return fmt.Errorf("ukuran file %d byte melebihi batas 10 MB (%d byte)", ukuran, UkuranMaksBerkasUKK)
	}
	return nil
}

// ekstensiFile mengembalikan ekstensi huruf kecil tanpa titik, atau ""
// bila tidak ada titik / titik di posisi terakhir.
func ekstensiFile(namaFile string) string {
	i := strings.LastIndexByte(namaFile, '.')
	if i < 0 || i == len(namaFile)-1 {
		return ""
	}
	return strings.ToLower(namaFile[i+1:])
}

// bersihkanNama menyimpan hanya karakter [a-zA-Z0-9._-]; karakter lain
// (termasuk '/', '\', spasi, karakter non-ASCII) diganti '_'. Titik di
// awal nama diganti '_' agar nama seperti ".." atau ".hidden" tidak
// pernah menjadi komponen path khusus. Hasil kosong diganti "_".
func bersihkanNama(nama string) string {
	var b strings.Builder
	for _, r := range nama {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '.' || r == '_' || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	hasil := b.String()
	// Ganti titik-titik awal agar tidak membentuk "." / "..".
	i := 0
	for i < len(hasil) && hasil[i] == '.' {
		i++
	}
	if i > 0 {
		hasil = strings.Repeat("_", i) + hasil[i:]
	}
	if hasil == "" {
		return "_"
	}
	return hasil
}

// PathAman membangun path penyimpanan berkas UKK:
//
//	storage/ukk/<ukkID>/<siswaID>/<nama-bersih>
//
// Nama file dibersihkan: hanya [a-zA-Z0-9._-] yang dipertahankan,
// sisanya menjadi '_', sehingga traversal (mis. "../../etc/passwd")
// mustahil — separator path '/' dari input ikut menjadi '_'.
// ID UKK dan ID siswa ikut dibersihkan dengan aturan yang sama agar
// ID kiriman client tidak bisa menyisipkan traversal. Path yang
// dihasilkan selalu berada di bawah "storage/ukk/".
func PathAman(ukkID, siswaID, namaFile string) string {
	return path.Join("storage", "ukk", bersihkanNama(ukkID), bersihkanNama(siswaID), bersihkanNama(namaFile))
}
