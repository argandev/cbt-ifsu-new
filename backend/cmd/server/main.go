// Command server — entrypoint CBT SMK v2 (Tim 1).
//
// Pemakaian:
// cbt-smk-v2 serve                          # jalankan server (default, tanpa argumen juga serve)
// cbt-smk-v2 migrate-up                     # jalankan migrasi yang tertunda
// cbt-smk-v2 migrate-down [n]               # rollback n langkah (default 1)
// cbt-smk-v2 create-admin --email E --password P --nama N
// cbt-smk-v2 reset-password --email E|--username U --password P
//
// Wiring: config -> pgxpool -> migrasi auto-up on boot -> sqlc Queries ->
// middleware global -> Register(...) tiap paket -> SPA fallback.
// Route /api/* dan /ws/* didahulukan; /* fallback ke index.html (SPA).
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	webdist "cbt-smk-v2/backend"
	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/analisis"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/auth"
	"cbt-smk-v2/backend/internal/backup"
	"cbt-smk-v2/backend/internal/banksoal"
	"cbt-smk-v2/backend/internal/config"
	"cbt-smk-v2/backend/internal/dashboard"
	"cbt-smk-v2/backend/internal/event"
	"cbt-smk-v2/backend/internal/exam"
	"cbt-smk-v2/backend/internal/master"
	"cbt-smk-v2/backend/internal/middleware"
	"cbt-smk-v2/backend/internal/monitor"
	"cbt-smk-v2/backend/internal/notifikasi"
	"cbt-smk-v2/backend/internal/rekap"
	"cbt-smk-v2/backend/internal/settings"
	"cbt-smk-v2/backend/internal/sod"
	"cbt-smk-v2/backend/internal/ukk"
	"cbt-smk-v2/backend/internal/upload"
	"cbt-smk-v2/backend/internal/users"
	"cbt-smk-v2/backend/migrations"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"golang.org/x/term"

	_ "github.com/jackc/pgx/v5/stdlib" // driver database/sql "pgx" untuk golang-migrate
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	args := os.Args[1:]
	cmd := "serve"
	if len(args) > 0 {
		cmd = args[0]
	}
	switch cmd {
	case "serve":
		if err := runServer(); err != nil {
			log.Fatalf("server: %v", err)
		}
	case "migrate-up":
		if err := runMigrateUp(); err != nil {
			log.Fatalf("migrate-up: %v", err)
		}
	case "migrate-down":
		steps := 1
		if len(args) > 1 {
			if n, err := strconv.Atoi(args[1]); err == nil && n > 0 {
				steps = n
			}
		}
		if err := runMigrateDown(steps); err != nil {
			log.Fatalf("migrate-down: %v", err)
		}
	case "create-admin":
		if err := runCreateAdmin(args[1:]); err != nil {
			log.Fatalf("create-admin: %v", err)
		}
	case "reset-password":
		if err := runResetPassword(args[1:]); err != nil {
			log.Fatalf("reset-password: %v", err)
		}
	default:
		fmt.Fprintf(os.Stderr, "Perintah tidak dikenal: %q\n", cmd)
		fmt.Fprintln(os.Stderr, "Perintah: serve | migrate-up | migrate-down [n] | create-admin | reset-password")
		os.Exit(2)
	}
}

// --- migrasi ----------------------------------------------------------------

// newMigrator membuat instance golang-migrate dari migrasi ter-embed.
func newMigrator(cfg *config.Config) (*migrate.Migrate, error) {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return nil, err
	}
	sqlDB, err := sql.Open("pgx", cfg.DBURL)
	if err != nil {
		return nil, err
	}
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return nil, err
	}
	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// migrateUp menjalankan migrasi yang tertunda. ErrNoChange bukan error.
func migrateUp(cfg *config.Config) error {
	m, err := newMigrator(cfg)
	if err != nil {
		return err
	}
	defer func() { _, _ = m.Close() }()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func runMigrateUp() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := migrateUp(cfg); err != nil {
		return err
	}
	log.Println("migrate-up: OK")
	return nil
}

func runMigrateDown(steps int) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	m, err := newMigrator(cfg)
	if err != nil {
		return err
	}
	defer func() { _, _ = m.Close() }()
	if err := m.Steps(-steps); err != nil && err != migrate.ErrNoChange {
		return err
	}
	log.Printf("migrate-down %d: OK", steps)
	return nil
}

// --- pool -------------------------------------------------------------------

func newPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DBURL)
	if err != nil {
		return nil, err
	}
	// Pool lebih besar dari CBT-Go (MaxOpenConns 10 terlalu kecil untuk
	// 300 siswa serentak — ANALISIS Bab 16.5 / Bab 2).
	poolCfg.MaxConns = 30
	poolCfg.MinConns = 2
	poolCfg.MaxConnLifetime = time.Hour
	poolCfg.HealthCheckPeriod = time.Minute
	return pgxpool.NewWithConfig(ctx, poolCfg)
}

// --- server -----------------------------------------------------------------

func runServer() error {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	pool, err := newPool(ctx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("koneksi database gagal: %w", err)
	}

	// Migrasi auto-up on boot. Gagal migrasi -> startup berhenti dengan
	// pesan jelas (jangan jalankan aplikasi dengan skema setengah jadi).
	if err := migrateUp(cfg); err != nil {
		return fmt.Errorf("migrasi gagal: %w", err)
	}

	// Bootstrap admin via env (opsional; hanya bila users masih kosong).
	if cfg.AdminBootstrap && cfg.AdminEmail != "" && cfg.AdminPassword != "" {
		if err := auth.CreateAdmin(ctx, pool, cfg.AdminEmail, "", "Administrator", cfg.AdminPassword); err != nil {
			if err != auth.ErrUsersNotEmpty {
				log.Printf("PERINGATAN: bootstrap admin gagal: %v", err)
			}
		} else {
			log.Println("Bootstrap admin dibuat dari ADMIN_EMAIL (wajib ganti password saat login pertama).")
		}
	}

	q := db.New(pool)

	e := echo.New()
	// Echo v5 tidak punya HideBanner (banner dihapus dari v5).
	e.HTTPErrorHandler = httpErrorHandler

	// Middleware global (urutan penting):
	// Logger -> Recover -> RateLimit -> RequireAuth -> CSRF.
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RateLimit())
	e.Use(middleware.RequireAuth(pool))
	e.Use(middleware.CSRF(pool))

	// Register paket-paket (signature kontrak: Register(e, pool, q, cfg)).
	// JANGAN hapus pemanggilan Register — paket STUB (master, exam,
	// monitor) diganti tim terkait tanpa mengubah signature.
	auth.Register(e, pool, q, cfg)
	audit.Register(e, pool, q, cfg)
	master.Register(e, pool, q, cfg)
	banksoal.Register(e, pool, q, cfg)
	banksoal.RegisterImportExcel(e, pool)
	banksoal.RegisterImportWord(e, pool)
	exam.Register(e, pool, q, cfg)
	event.Register(e, pool, q, cfg)
	monitor.Register(e, pool, q, cfg)
	rekap.Register(e, pool, q, cfg)
	analisis.Register(e, pool, q, cfg)
	settings.Register(e, pool, q, cfg)
	backup.Register(e, pool, q, cfg)
	ukk.Register(e, pool, q, cfg)
	upload.Register(e, pool, q, cfg)
	dashboard.Register(e, pool)
	users.Register(e, pool, q, cfg)
	sod.Register(e, pool, q, cfg)
	notifikasi.Register(e, pool, q, cfg)

	// SPA fallback — didaftarkan TERAKHIR agar /api/* dan /ws/* didahulukan.
	// Echo memprioritaskan route statis (/api/...) di atas wildcard /*,
	// tetapi pendaftaran terakhir membuat maksudnya eksplisit.
	e.GET("/*", serveSPA)

	// Timeout HTTP eksplisit (Fase 3, BUG-02): default Echo v5
	// WriteTimeout 30 dtk membuat import 300 siswa (~50 dtk) terpotong
	// di tengah (klien EOF padahal server commit). Server LAN sekolah:
	// WriteTimeout 120 dtk + ReadTimeout 60 dtk aman. Import tetap
	// dianjurkan per batch 100 siswa (KONTRAK §10).
	// Echo v5 TIDAK punya field e.Server — timeout diatur lewat
	// StartConfig.BeforeServeFunc (dipanggil sebelum Serve).
	log.Printf("CBT SMK v2 berjalan di http://localhost:%d (ENV=%s)", cfg.PORT, cfg.Env)
	sc := echo.StartConfig{
		Address: cfg.Addr,
		BeforeServeFunc: func(s *http.Server) error {
			s.ReadTimeout = 60 * time.Second
			s.WriteTimeout = 120 * time.Second
			s.IdleTimeout = 120 * time.Second
			return nil
		},
	}
	ctxSrv, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := sc.Start(ctxSrv, e); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// httpErrorHandler mengembalikan error dalam format kontrak
// {"error": {"code": "...", "message": "..."}} untuk error Echo
// (404, 405, dsb) yang tidak ditangani handler.
// Echo v5: HTTPErrorHandler func(c *Context, err error); HTTPError.Message
// adalah string; c.Response() konkretnya *echo.Response (punya Committed).
func httpErrorHandler(c *echo.Context, err error) {
	if r, ok := c.Response().(*echo.Response); ok && r.Committed {
		return
	}
	code := http.StatusInternalServerError
	message := "Terjadi kesalahan internal."
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if he.Message != "" {
			message = he.Message
		}
	}
	switch code {
	case http.StatusNotFound:
		message = "Route tidak ditemukan."
	case http.StatusMethodNotAllowed:
		message = "Method tidak diizinkan untuk route ini."
	}
	_ = c.JSON(code, map[string]any{
		"error": map[string]any{
			"code":    errorCodeForStatus(code),
			"message": message,
		},
	})
}

func errorCodeForStatus(code int) string {
	switch code {
	case http.StatusNotFound:
		return "not_found"
	case http.StatusMethodNotAllowed:
		return "method_not_allowed"
	case http.StatusUnauthorized:
		return "unauthenticated"
	case http.StatusForbidden:
		return "forbidden"
	default:
		return "internal_error"
	}
}

// serveSPA menyajikan frontend ter-embed (backend/dist, disalin dari
// frontend/dist oleh `make build`). File statis disajikan langsung;
// path lain fallback ke index.html (SPA routing).
func serveSPA(c *echo.Context) error {
	sub, err := fs.Sub(webdistFS(), "dist")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Frontend tidak tersedia.")
	}
	reqPath := strings.TrimPrefix(c.Request().URL.Path, "/")
	if reqPath == "" {
		reqPath = "index.html"
	}
	// Cegah path traversal.
	reqPath = path.Clean("/" + reqPath)[1:]
	data, err := fs.ReadFile(sub, reqPath)
	if err != nil {
		// Fallback SPA: semua path non-file -> index.html.
		data, err = fs.ReadFile(sub, "index.html")
		if err != nil {
			return c.String(http.StatusNotFound, "Frontend tidak tersedia. Jalankan: make build")
		}
		reqPath = "index.html"
	}
	ctype := mime.TypeByExtension(path.Ext(reqPath))
	if ctype == "" {
		ctype = "application/octet-stream"
		if reqPath == "index.html" {
			ctype = "text/html; charset=utf-8"
		}
	}
	return c.Blob(http.StatusOK, ctype, data)
}

// webdistFS mengembalikan FS frontend ter-embed.
func webdistFS() fs.FS {
	return webdist.Dist
}

// --- CLI: create-admin / reset-password -------------------------------------

func promptPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	return string(pw), nil
}

func runCreateAdmin(args []string) error {
	fs := flag.NewFlagSet("create-admin", flag.ExitOnError)
	email := fs.String("email", "", "Email admin (atau --username)")
	username := fs.String("username", "", "Username admin (atau --email)")
	nama := fs.String("nama", "", "Nama lengkap admin")
	password := fs.String("password", "", "Password (bila kosong, diminta interaktif)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *password == "" {
		pw, err := promptPassword("Password admin baru (min 8 karakter): ")
		if err != nil {
			return err
		}
		*password = pw
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx := context.Background()
	pool, err := newPool(ctx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := auth.CreateAdmin(ctx, pool, *email, *username, *nama, *password); err != nil {
		return err
	}
	log.Println("Admin dibuat. Akun WAJIB ganti password saat login pertama.")
	return nil
}

func runResetPassword(args []string) error {
	fs := flag.NewFlagSet("reset-password", flag.ExitOnError)
	email := fs.String("email", "", "Email akun")
	username := fs.String("username", "", "Username akun")
	password := fs.String("password", "", "Password baru (bila kosong, diminta interaktif)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	login := *email
	if login == "" {
		login = *username
	}
	if login == "" {
		return fmt.Errorf("--email atau --username wajib diisi")
	}
	if *password == "" {
		pw, err := promptPassword("Password baru (min 8 karakter): ")
		if err != nil {
			return err
		}
		*password = pw
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx := context.Background()
	pool, err := newPool(ctx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := auth.ResetPassword(ctx, pool, login, *password); err != nil {
		return err
	}
	log.Println("Password direset. Akun WAJIB ganti password saat login berikutnya; semua session lama dihapus.")
	return nil
}
