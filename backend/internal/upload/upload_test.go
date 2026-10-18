package upload

import (
	"bytes"
	"testing"
)

// Gambar minimal yang sah per tipe (hanya header signature — cukup untuk
// deteksi magic bytes; bukan gambar utuh yang bisa dirender).
func TestDeteksiTipeGambar(t *testing.T) {
	png := append([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, bytes.Repeat([]byte{0}, 16)...)
	jpeg := append([]byte{0xFF, 0xD8, 0xFF, 0xE0}, bytes.Repeat([]byte{0}, 16)...)
	webp := append([]byte("RIFF"), 0x00, 0x00, 0x00, 0x00)
	webp = append(webp, []byte("WEBP")...)
	webp = append(webp, bytes.Repeat([]byte{0}, 8)...)
	gif := append([]byte("GIF89a"), bytes.Repeat([]byte{0}, 16)...)

	cases := []struct {
		nama string
		isi  []byte
		ext  string
	}{
		{"PNG", png, "png"},
		{"JPEG", jpeg, "jpg"},
		{"WebP", webp, "webp"},
		{"GIF89a", gif, "gif"},
	}
	for _, tc := range cases {
		tipe, ok := deteksiTipeGambar(tc.isi)
		if !ok || tipe.ext != tc.ext {
			t.Errorf("%s: mau ext %q, dapat %+v (ok=%v)", tc.nama, tc.ext, tipe, ok)
		}
	}
}

// File jahat: ekstensi .png + Content-Type image/png tidak boleh menolong —
// isinya PHP/HTML harus ditolak karena magic bytes tidak cocok.
func TestDeteksiTipeGambar_TolakFilePalsu(t *testing.T) {
	palsu := [][]byte{
		[]byte("<?php echo 'pwned';"),
		[]byte("<html><script>alert(1)</script></html>"),
		[]byte("%PDF-1.7 palsu"),
		{},
		[]byte{0x89, 0x50, 0x4E}, // signature PNG terpotong
		[]byte("RIFFxxxxXXXX"),   // RIFF tapi bukan WEBP
	}
	for i, isi := range palsu {
		if _, ok := deteksiTipeGambar(isi); ok {
			t.Errorf("kasus %d: file bukan gambar harus ditolak, malah diterima", i)
		}
	}
}

func TestNamaFileAman(t *testing.T) {
	baik := "123e4567-e89b-12d3-a456-426614174000.png"
	if !namaFileAman.MatchString(baik) {
		t.Errorf("nama UUID yang sah ditolak: %s", baik)
	}
	buruk := []string{
		"../../etc/passwd",
		"123e4567-e89b-12d3-a456-426614174000.php",
		"123e4567-e89b-12d3-a456-426614174000.png/..",
		"logo.png",
		"123e4567-e89b-12d3-a456-426614174000.PNG.exe",
		"",
		"123e4567-e89b-12d3-a456-426614174000.png\x00",
	}
	for _, n := range buruk {
		if namaFileAman.MatchString(n) {
			t.Errorf("nama berbahaya diterima: %q", n)
		}
	}
}
