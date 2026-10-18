// Package config — konfigurasi aplikasi.
//
// Diperkaya Tim 1 (dari scaffolding minimal Tim 5). Field lama (DatabaseURL,
// Addr) DIPERTAHANKAN agar kode tim lain tidak rusak; field baru menambah
// konfigurasi yang dibutuhkan auth/session (SESSION_SECRET, APP_ENV, dll).
// Satu-satunya sumber konfigurasi runtime. Tidak ada secret yang di-hardcode.
//
// Konfigurasi dibaca dari environment; sebagai kemudahan dev, bila ada file
// .env di working directory / backend/ / root repo, nilainya dipakai HANYA
// untuk variabel yang belum ter-set di environment.
package config

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config memegang konfigurasi runtime backend.
type Config struct {
	// DatabaseURL adalah DSN PostgreSQL, mis. postgres://user:pass@host:5432/cbt_smk_v2
	// (dari DATABASE_URL). Selalu sama dengan DBURL.
	DatabaseURL string `env:"DATABASE_URL"`
	// Addr adalah alamat listen HTTP, mis. ":9988" (diturunkan dari PORT).
	Addr string `env:"ADDR"`

	// DBURL adalah DSN PostgreSQL (alias DatabaseURL, dari DATABASE_URL).
	DBURL string
	// PORT adalah port HTTP server. Default 9988.
	PORT int
	// SessionSecret dipakai untuk menandatangani/mengamankan sesi.
	// Di produksi WAJIB diisi via SESSION_SECRET. Di dev, bila kosong,
	// Load() membangkitkan secret acak per proses (sesi tidak tahan restart)
	// dan mencatat peringatan.
	SessionSecret string
	// Env adalah "dev" atau "prod", dari APP_ENV. Default "dev".
	Env string
	// AdminBootstrap bila true mengizinkan bootstrap admin awal via
	// variabel ADMIN_EMAIL + ADMIN_PASSWORD saat tabel users masih kosong.
	// Di produksi, gunakan CLI `create-admin` sebagai gantinya.
	AdminBootstrap bool
	// AdminEmail / AdminPassword hanya dipakai saat AdminBootstrap true.
	AdminEmail    string
	AdminPassword string
	// StorageDir adalah direktori akar penyimpanan berkas (upload UKK,
	// gambar soal). Dari STORAGE_DIR, default "storage" (relatif terhadap
	// working directory proses). Di produksi set ke path absolut yang
	// writable, mis. /var/lib/cbt-smk-v2/storage.
	StorageDir string
}

// IsProd melaporkan apakah aplikasi berjalan dalam mode produksi.
func (c *Config) IsProd() bool {
	return strings.EqualFold(c.Env, "prod")
}

// Load membaca konfigurasi dari environment (+ file .env bila ada).
// Mengembalikan error bila variabel wajib untuk mode produksi tidak terisi.
func Load() (*Config, error) {
	loadDotEnv()

	cfg := &Config{
		Env:            firstNonEmpty(os.Getenv("APP_ENV"), "dev"),
		SessionSecret:  os.Getenv("SESSION_SECRET"),
		AdminBootstrap: os.Getenv("ADMIN_BOOTSTRAP") == "true" || os.Getenv("ADMIN_BOOTSTRAP") == "1",
		AdminEmail:     os.Getenv("ADMIN_EMAIL"),
		AdminPassword:  os.Getenv("ADMIN_PASSWORD"),
		StorageDir:     firstNonEmpty(os.Getenv("STORAGE_DIR"), "storage"),
	}
	cfg.DBURL = os.Getenv("DATABASE_URL")
	cfg.DatabaseURL = cfg.DBURL

	port := 9988
	if raw := os.Getenv("PORT"); raw != "" {
		p, err := strconv.Atoi(raw)
		if err != nil || p <= 0 || p > 65535 {
			return nil, fmt.Errorf("PORT tidak valid: %q", raw)
		}
		port = p
	}
	cfg.PORT = port
	cfg.Addr = ":" + strconv.Itoa(port)

	if cfg.DBURL == "" {
		return nil, fmt.Errorf("DATABASE_URL wajib diisi")
	}
	if cfg.SessionSecret == "" {
		if cfg.IsProd() {
			return nil, fmt.Errorf("SESSION_SECRET wajib diisi di mode prod (APP_ENV=prod)")
		}
		// Dev only: secret acak per proses. Sesi lama menjadi tidak valid
		// setelah restart — perilaku yang dapat diterima di dev.
		var b [32]byte
		if _, err := rand.Read(b[:]); err != nil {
			return nil, fmt.Errorf("gagal membangkitkan SESSION_SECRET dev: %w", err)
		}
		cfg.SessionSecret = hex.EncodeToString(b[:])
		fmt.Fprintln(os.Stderr, "PERINGATAN: SESSION_SECRET kosong — memakai secret acak per proses (dev only). Set SESSION_SECRET untuk sesi yang tahan restart.")
	}
	return cfg, nil
}

// loadDotEnv memuat .env dari cwd, ../.env (saat dijalankan dari backend/),
// atau ../../.env (saat dijalankan dari backend/cmd/server). Variabel yang
// sudah ada di environment TIDAK ditimpa. Format: KEY=VALUE, # komentar,
// nilai boleh dikutip "..." / '...'. Parser minimal — bukan pengganti
// pustaka dotenv untuk kasus kompleks.
func loadDotEnv() {
	candidates := []string{".env", "../.env", "../../.env"}
	for _, path := range candidates {
		f, err := os.Open(filepath.Clean(path))
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if strings.HasPrefix(line, "export ") {
				line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
			}
			key, value, found := strings.Cut(line, "=")
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			if !found || key == "" {
				continue
			}
			if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'')) {
				value = value[1 : len(value)-1]
			}
			if _, exists := os.LookupEnv(key); !exists {
				_ = os.Setenv(key, value)
			}
		}
		_ = f.Close()
		return // pakai file .env pertama yang ditemukan
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
