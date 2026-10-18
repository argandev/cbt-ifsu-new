package upload

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// --- Repository (file store) ------------------------------------------------
// Paket upload tidak memakai database (selain audit log): "repository"-nya
// adalah penyimpanan file di disk (storage/upload/image/).

// ErrSiapkanPenyimpanan menandai kegagalan menyiapkan direktori.
var ErrSiapkanPenyimpanan = errors.New("gagal menyiapkan penyimpanan")

// ErrDiLuarRoot menandai path yang lolos validasi nama tetapi berada di
// luar direktori upload (defense in depth).
var ErrDiLuarRoot = errors.New("path di luar direktori upload")

// Repository adalah lapisan penyimpanan file paket upload.
type Repository struct {
	// storageDir adalah akar direktori penyimpanan
	// (dari config.StorageDir; default "storage").
	storageDir string
}

// NewRepository membuat Repository.
func NewRepository(storageDir string) *Repository {
	if storageDir == "" {
		storageDir = "storage"
	}
	return &Repository{storageDir: storageDir}
}

// Simpan menulis isi ke <storageDir>/upload/image/<diskName> (file baru,
// O_EXCL). Mengembalikan path relatif untuk audit. (Dulu: blok tulis
// file inline di uploadGambar.)
func (r *Repository) Simpan(diskName string, isi []byte) (string, error) {
	relDir := filepath.Join(r.storageDir, subDirUpload)
	absDir, _ := filepath.Abs(relDir)
	if err := os.MkdirAll(absDir, 0o750); err != nil {
		return "", ErrSiapkanPenyimpanan
	}
	relPath := filepath.Join(relDir, diskName)
	absPath, _ := filepath.Abs(relPath)
	// Tulis ke file baru (O_EXCL): UUID acak tidak boleh menimpa file lama.
	dst, err := os.OpenFile(absPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return "", err
	}
	if _, err := dst.Write(isi); err != nil {
		_ = dst.Close()
		_ = os.Remove(absPath)
		return "", err
	}
	_ = dst.Close()
	return relPath, nil
}

// Baca membaca file gambar dari direktori upload dengan validasi path.
// (Dulu: blok baca + validasi path inline di sajikanGambar.)
func (r *Repository) Baca(nama string) ([]byte, error) {
	// Keamanan path: file harus berada di dalam <storageDir>/upload/image/.
	absRoot, _ := filepath.Abs(filepath.Join(r.storageDir, subDirUpload))
	absPath, _ := filepath.Abs(filepath.Join(absRoot, filepath.Base(nama)))
	if absPath != absRoot && !strings.HasPrefix(absPath, absRoot+string(os.PathSeparator)) {
		return nil, ErrDiLuarRoot
	}
	return os.ReadFile(absPath)
}
