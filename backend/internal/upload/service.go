package upload

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"strings"

	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Konstanta & deteksi tipe ----------------------------------------------
// (Dipindah dari upload.go tanpa perubahan.)

// maxGambarBytes adalah ukuran maksimum satu file gambar: 5 MB.
const maxGambarBytes = 5 << 20 // 5 MiB = 5242880 byte

// subDirUpload adalah subfolder gambar di dalam direktori penyimpanan
// (storageDir/upload/image).
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

// --- Service ----------------------------------------------------------------

// RequestActor adalah identitas pemanggil untuk audit log.
// Dibangun handler dari echo.Context agar service bebas dari HTTP.
type RequestActor struct {
	UserID string
	RealIP string
}

// Service adalah lapisan logika bisnis paket upload.
type Service struct {
	repo *Repository
	pool *pgxpool.Pool
}

// NewService membuat Service.
func NewService(repo *Repository, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, pool: pool}
}

// catat menulis audit log. (Dulu: audit.Catat di handler.)
func (s *Service) catat(ctx context.Context, actor RequestActor, aksi, entitas, entitasID string, detail any) {
	_ = audit.Catat(ctx, s.pool, actor.UserID, aksi, entitas, entitasID, detail, actor.RealIP)
}

// Simpan memvalidasi + menyimpan satu file gambar upload.
// Mengembalikan payload respons 201. (Dulu: uploadGambar.)
func (s *Service) Simpan(ctx context.Context, actor RequestActor, fileHeader *multipart.FileHeader) (map[string]any, *platform.AppError) {
	if fileHeader.Size <= 0 {
		return nil, platform.ErrBadRequest("validation_error", "File kosong.")
	}
	if fileHeader.Size > maxGambarBytes {
		return nil, platform.ErrBadRequest("validation_error", "Ukuran gambar melebihi 5 MB.")
	}
	src, err := fileHeader.Open()
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca berkas upload.")
	}
	defer src.Close()
	// Baca isi sebenarnya dengan batas max+1 agar klaim ukuran di header
	// tidak bisa menipu: bila terbaca lebih dari 5 MB, tolak.
	isi, err := io.ReadAll(io.LimitReader(src, maxGambarBytes+1))
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca berkas upload.")
	}
	if int64(len(isi)) > maxGambarBytes {
		return nil, platform.ErrBadRequest("validation_error", "Ukuran gambar melebihi 5 MB.")
	}
	// MAGIC BYTES — satu-satunya sumber kebenaran tipe file. Ekstensi nama
	// file dan Content-Type kiriman client sengaja diabaikan.
	tipe, cocok := deteksiTipeGambar(isi)
	if !cocok {
		return nil, platform.ErrBadRequest("validation_error",
			"File bukan gambar yang diizinkan. Gunakan PNG, JPEG, WebP, atau GIF (maks. 5 MB).")
	}
	// SHA-256 dari isi file (dicatat di audit log + dikembalikan ke client).
	sum := sha256.Sum256(isi)
	hash := hex.EncodeToString(sum[:])

	// Nama file di disk: UUID acak + ekstensi aman turunan magic bytes.
	diskName := uuid.NewString() + "." + tipe.ext
	if _, err := s.repo.Simpan(diskName, isi); err != nil {
		if errors.Is(err, ErrSiapkanPenyimpanan) {
			return nil, platform.ErrInternal("internal_error", "Gagal menyiapkan penyimpanan berkas.")
		}
		return nil, platform.ErrInternal("internal_error", "Gagal menyimpan berkas.")
	}

	url := "/api/v1/upload/image/" + diskName
	s.catat(ctx, actor, "upload.image", "upload_image", diskName,
		map[string]any{"sha256": hash, "ukuran": len(isi), "content_type": tipe.contentType, "url": url})
	return map[string]any{
		"url":          url,
		"sha256":       hash,
		"ukuran":       len(isi),
		"content_type": tipe.contentType,
	}, nil
}

// GambarTersaji adalah hasil baca gambar untuk disajikan via GET.
type GambarTersaji struct {
	ContentType string
	Data        []byte
}

// Sajikan membaca + memvalidasi satu file gambar untuk endpoint GET.
// (Dulu: sajikanGambar.)
func (s *Service) Sajikan(nama string) (*GambarTersaji, *platform.AppError) {
	if !namaFileAman.MatchString(strings.ToLower(nama)) {
		return nil, platform.ErrNotFound("not_found", "Gambar tidak ditemukan.")
	}
	data, err := s.repo.Baca(nama)
	if err != nil {
		if errors.Is(err, ErrDiLuarRoot) {
			return nil, platform.ErrNotFound("not_found", "Gambar tidak ditemukan.")
		}
		return nil, platform.ErrNotFound("not_found", "Gambar tidak ditemukan di penyimpanan.")
	}
	ext := strings.ToLower(filepath.Ext(nama))
	var contentType string
	for _, t := range tipeGambarDiizinkan {
		if "."+t.ext == ext {
			contentType = t.contentType
			break
		}
	}
	if contentType == "" {
		return nil, platform.ErrNotFound("not_found", "Gambar tidak ditemukan.")
	}
	// Validasi ulang magic bytes file di disk (bela bila file di disk
	// diganti/diutak-atik): sajikan hanya bila isinya benar-benar gambar.
	if tipe, cocok := deteksiTipeGambar(data); !cocok || tipe.contentType != contentType {
		return nil, platform.ErrNotFound("not_found", "Gambar tidak ditemukan.")
	}
	// GET tidak dicatat di audit log (gambar soal dimuat berulang-ulang
	// tiap render soal; audit tiap GET akan membanjiri audit_log).
	return &GambarTersaji{ContentType: contentType, Data: data}, nil
}
