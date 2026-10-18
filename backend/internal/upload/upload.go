// Package upload — upload gambar umum (Tim D).
//
// Satu endpoint untuk SEMUA kebutuhan gambar aplikasi, agar tidak ada lagi
// input URL gambar ketik manual (logo sekolah di Settings, media soal di
// SoalEditor, gambar opsi pilihan_gambar — lihat docs/AUDIT-UI-UX.md §4.4.1):
//
//	POST /api/v1/upload/image        (multipart, field "file")
//	GET  /api/v1/upload/image/:nama  (menyajikan file yang sudah diunggah)
//
// Keamanan (berlapis, jangan kurangi satu pun tanpa kontrak baru):
//   - Auth: route /api/* sudah dilewati middleware global RequireAuth —
//     tanpa session login, request ditolak 401 sebelum sampai ke handler.
//   - CSRF: middleware global CSRF mewajibkan header X-CSRF-Token ==
//     session.csrf_token untuk POST (lihat middleware/middleware.go).
//   - Otorisasi upload: hanya role yang memang mengelola konten bergambar
//     (settings.update = logo, banksoal.create/update = media soal,
//     ukk.create). Siswa TIDAK punya ketiganya sehingga tidak bisa upload.
//     Admin (role "admin") lolos semua via audit.HasPermission.
//   - GET file: boleh dibaca semua user yang login (siswa pun perlu melihat
//     gambar soal saat ujian) — tetapi nama file tidak dapat ditebak
//     (UUID) dan path divalidasi ketat.
//   - Magic bytes: tipe file ditentukan dari isi file (signature), BUKAN
//     dari ekstensi nama file atau header Content-Type kiriman client.
//     File "jahat.png" yang isinya PHP/HTML ditolak.
//   - Nama file di disk: UUID acak + ekstensi aman turunan magic bytes
//     (bukan nama asli kiriman client) — tidak ada eksekusi / traversal.
//   - Ukuran: maks 5 MB per file (ditegakkan dari header DAN isi sebenarnya).
//   - Hash SHA-256: dihitung dari isi file, dicatat di audit log
//     (aksi "upload.image") dan dikembalikan di respons.
//   - Penyimpanan: storage/upload/image/ — folder storage yang sama dengan
//     berkas UKK (storage/ukk/...). File TIDAK pernah disajikan sebagai
//     file statis webroot; hanya lewat handler GET di bawah ini yang
//     mengirim Content-Type dari tipe terdeteksi + nosniff.
//
// Respons POST (201): {"data": {"url", "sha256", "ukuran", "content_type"}}
// dengan url = URL internal "/api/v1/upload/image/<uuid>.<ext>" (BUKAN URL
// eksternal/CDN). Frontend menyimpan url ini apa adanya (mis. settings
// logo_url, media soal) — JANGAN pernah menyimpan URL eksternal.
package upload

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/config"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

// maxGambarBytes adalah ukuran maksimum satu file gambar: 5 MB.
const maxGambarBytes = 5 << 20 // 5 MiB = 5242880 byte

// storageRoot adalah akar folder penyimpanan yang sama dengan UKK
// (storage/ukk/...). Di sini: storage/upload/image/<uuid>.<ext>.
const storageRoot = "storage"

// subDirUpload adalah subfolder gambar di dalam storageRoot.
const subDirUpload = "upload/image"

// tipeGambar mendeskripsikan satu tipe gambar yang diizinkan.
type tipeGambar struct {
	ext         string // ekstensi aman yang dipakai untuk nama file di disk
	contentType string
	// cek menerima seluruh isi file dan melaporkan apakah signature-nya cocok.
	cek func(b []byte) bool
}

func isPNG(b []byte) bool {
	return len(b) >= 8 && bytes.Equal(b[:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
}

func isJPEG(b []byte) bool {
	return len(b) >= 3 && b[0] == 0xFF && b[1] == 0xD8 && b[2] == 0xFF
}

func isGIF(b []byte) bool {
	return len(b) >= 6 && (bytes.Equal(b[:6], []byte("GIF87a")) || bytes.Equal(b[:6], []byte("GIF89a")))
}

func isWebP(b []byte) bool {
	return len(b) >= 12 && bytes.Equal(b[:4], []byte("RIFF")) && bytes.Equal(b[8:12], []byte("WEBP"))
}

// tipeGambarDiizinkan adalah whitelist FINAL: PNG, JPEG, WebP, GIF.
// Urutan penting untuk deteksi: tiap cek memakai signature yang unik.
var tipeGambarDiizinkan = []tipeGambar{
	{ext: "png", contentType: "image/png", cek: isPNG},
	{ext: "jpg", contentType: "image/jpeg", cek: isJPEG},
	{ext: "webp", contentType: "image/webp", cek: isWebP},
	{ext: "gif", contentType: "image/gif", cek: isGIF},
}

// deteksiTipeGambar menentukan tipe dari magic bytes isi file.
// Mengembalikan (tipe, true) bila cocok, atau (tipeGambar{}, false) bila
// isinya bukan gambar yang diizinkan — apa pun nama file / Content-Type
// yang dikirim client.
func deteksiTipeGambar(b []byte) (tipeGambar, bool) {
	for _, t := range tipeGambarDiizinkan {
		if t.cek(b) {
			return t, true
		}
	}
	return tipeGambar{}, false
}

// namaFileAman memvalidasi nama file hasil upload untuk endpoint GET:
// hanya "<uuid>.<ext>" dengan ext dari whitelist. Nama apa pun yang
// tidak cocok (termasuk "..", "/", "\") ditolak — cegah path traversal.
var namaFileAman = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\.(png|jpg|webp|gif)$`)

// bolehUpload melaporkan apakah user yang login boleh mengunggah gambar:
// punya salah satu permission pengelola konten bergambar. Admin lolos
// semua via audit.HasPermission.
func bolehUpload(c *echo.Context) bool {
	for _, perm := range []string{"settings.update", "banksoal.create", "banksoal.update", "ukk.create"} {
		if audit.HasPermission(c, perm) {
			return true
		}
	}
	return false
}

func writeError(c *echo.Context, status int, code, message string) error {
	return c.JSON(status, map[string]any{
		"error": map[string]any{"code": code, "message": message},
	})
}

func writeOKStatus(c *echo.Context, status int, data any) error {
	return c.JSON(status, map[string]any{"data": data})
}

func clientIP(c *echo.Context) string {
	if ip := c.RealIP(); ip != "" {
		return ip
	}
	return c.Request().RemoteAddr
}

// Register mendaftarkan endpoint upload gambar.
//
//	POST /api/v1/upload/image       (upload; wajib login + CSRF + permission)
//	GET  /api/v1/upload/image/:nama (sajikan file; wajib login)
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q
	_ = cfg

	e.POST("/api/v1/upload/image", func(c *echo.Context) error { return uploadGambar(c, pool) })
	e.GET("/api/v1/upload/image/:nama", func(c *echo.Context) error { return sajikanGambar(c) })
}

func uploadGambar(c *echo.Context, pool *pgxpool.Pool) error {
	userID, _, _, _, ok := audit.CurrentUser(c)
	if !ok {
		return writeError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
	}
	if !bolehUpload(c) {
		return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mengunggah gambar.")
	}
	// ParseMultipartForm dengan batas 5 MB (+1 MB ruang header multipart).
	// File yang lebih besar ditolak di bawah lewat Size + isi sebenarnya.
	if err := c.Request().ParseMultipartForm(maxGambarBytes + (1 << 20)); err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Body multipart tidak valid (maks. 5 MB).")
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return writeError(c, http.StatusBadRequest, "validation_error", "Field file wajib diisi.")
	}
	if fileHeader.Size <= 0 {
		return writeError(c, http.StatusBadRequest, "validation_error", "File kosong.")
	}
	if fileHeader.Size > maxGambarBytes {
		return writeError(c, http.StatusBadRequest, "validation_error", "Ukuran gambar melebihi 5 MB.")
	}
	src, err := fileHeader.Open()
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca berkas upload.")
	}
	defer src.Close()
	// Baca isi sebenarnya dengan batas max+1 agar klaim ukuran di header
	// tidak bisa menipu: bila terbaca lebih dari 5 MB, tolak.
	isi, err := io.ReadAll(io.LimitReader(src, maxGambarBytes+1))
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca berkas upload.")
	}
	if int64(len(isi)) > maxGambarBytes {
		return writeError(c, http.StatusBadRequest, "validation_error", "Ukuran gambar melebihi 5 MB.")
	}
	// MAGIC BYTES — satu-satunya sumber kebenaran tipe file. Ekstensi nama
	// file dan Content-Type kiriman client sengaja diabaikan.
	tipe, cocok := deteksiTipeGambar(isi)
	if !cocok {
		return writeError(c, http.StatusBadRequest, "validation_error",
			"File bukan gambar yang diizinkan. Gunakan PNG, JPEG, WebP, atau GIF (maks. 5 MB).")
	}
	// SHA-256 dari isi file (dicatat di audit log + dikembalikan ke client).
	sum := sha256.Sum256(isi)
	hash := hex.EncodeToString(sum[:])

	// Nama file di disk: UUID acak + ekstensi aman turunan magic bytes.
	diskName := uuid.NewString() + "." + tipe.ext
	relDir := filepath.Join(storageRoot, subDirUpload)
	absDir, _ := filepath.Abs(relDir)
	if err := os.MkdirAll(absDir, 0o750); err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menyiapkan penyimpanan berkas.")
	}
	relPath := filepath.Join(relDir, diskName)
	absPath, _ := filepath.Abs(relPath)
	// Tulis ke file baru (O_EXCL): UUID acak tidak boleh menimpa file lama.
	dst, err := os.OpenFile(absPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menyimpan berkas.")
	}
	if _, err := dst.Write(isi); err != nil {
		_ = dst.Close()
		_ = os.Remove(absPath)
		return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menyimpan berkas.")
	}
	_ = dst.Close()

	url := "/api/v1/upload/image/" + diskName
	_ = audit.Catat(c.Request().Context(), pool, userID, "upload.image", "upload_image", diskName,
		map[string]any{"sha256": hash, "ukuran": len(isi), "content_type": tipe.contentType, "url": url}, clientIP(c))
	return writeOKStatus(c, http.StatusCreated, map[string]any{
		"url":          url,
		"sha256":       hash,
		"ukuran":       len(isi),
		"content_type": tipe.contentType,
	})
}

func sajikanGambar(c *echo.Context) error {
	if _, _, _, _, ok := audit.CurrentUser(c); !ok {
		return writeError(c, http.StatusUnauthorized, "unauthenticated", "Harus login dulu.")
	}
	nama := c.Param("nama")
	if !namaFileAman.MatchString(strings.ToLower(nama)) {
		return writeError(c, http.StatusNotFound, "not_found", "Gambar tidak ditemukan.")
	}
	// Keamanan path: file harus berada di dalam ./storage/upload/image/.
	absRoot, _ := filepath.Abs(filepath.Join(storageRoot, subDirUpload))
	absPath, _ := filepath.Abs(filepath.Join(absRoot, filepath.Base(nama)))
	if absPath != absRoot && !strings.HasPrefix(absPath, absRoot+string(os.PathSeparator)) {
		return writeError(c, http.StatusNotFound, "not_found", "Gambar tidak ditemukan.")
	}
	ext := strings.ToLower(filepath.Ext(absPath))
	var contentType string
	for _, t := range tipeGambarDiizinkan {
		if "."+t.ext == ext {
			contentType = t.contentType
			break
		}
	}
	if contentType == "" {
		return writeError(c, http.StatusNotFound, "not_found", "Gambar tidak ditemukan.")
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return writeError(c, http.StatusNotFound, "not_found", "Gambar tidak ditemukan di penyimpanan.")
	}
	// Validasi ulang magic bytes file di disk (bela bila file di disk
	// diganti/diutak-atik): sajikan hanya bila isinya benar-benar gambar.
	if tipe, cocok := deteksiTipeGambar(data); !cocok || tipe.contentType != contentType {
		return writeError(c, http.StatusNotFound, "not_found", "Gambar tidak ditemukan.")
	}
	// GET tidak dicatat di audit log (gambar soal dimuat berulang-ulang
	// tiap render soal; audit tiap GET akan membanjiri audit_log).
	c.Response().Header().Set("Content-Type", contentType)
	c.Response().Header().Set("X-Content-Type-Options", "nosniff")
	c.Response().Header().Set("Cache-Control", "private, max-age=3600")
	c.Response().Header().Set("Content-Disposition", "inline")
	return c.Blob(http.StatusOK, contentType, data)
}
