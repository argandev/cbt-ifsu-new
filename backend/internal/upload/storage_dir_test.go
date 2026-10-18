package upload

import (
	"os"
	"path/filepath"
	"testing"
)

// StorageDir kustom (mis. dari STORAGE_DIR) harus dipakai repository,
// bukan hardcoded "storage".
func TestRepository_SimpanPakaiStorageDirKustom(t *testing.T) {
	custom := filepath.Join(t.TempDir(), "dataku")
	r := NewRepository(custom)
	rel, err := r.Simpan("uji.png", []byte{0x89, 0x50, 0x4E, 0x47})
	if err != nil {
		t.Fatalf("Simpan gagal: %v", err)
	}
	abs, _ := filepath.Abs(rel)
	if _, err := os.Stat(abs); err != nil {
		t.Fatalf("file tidak ada di %s: %v", abs, err)
	}
	// File harus berada di dalam direktori kustom.
	absCustom, _ := filepath.Abs(custom)
	if len(abs) <= len(absCustom) || abs[:len(absCustom)] != absCustom {
		t.Fatalf("file %s di luar storageDir kustom %s", abs, absCustom)
	}
	// Baca kembali.
	isi, err := r.Baca("uji.png")
	if err != nil {
		t.Fatalf("Baca gagal: %v", err)
	}
	if len(isi) != 4 {
		t.Fatalf("isi terbaca %d byte, mau 4", len(isi))
	}
}

// Default tetap "storage" bila storageDir kosong (perilaku lama).
func TestRepository_DefaultStorageDir(t *testing.T) {
	r := NewRepository("")
	if r.storageDir != "storage" {
		t.Fatalf("default storageDir = %q, mau %q", r.storageDir, "storage")
	}
}
