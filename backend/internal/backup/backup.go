// Package backup — Backup / Restore database (Tim Fase 2).
//
// Definisi (kontrak tugas Fase 2):
//
//	Backup  = dump PostgreSQL via pg_dump -Fc (format custom) + metadata
//	          di tabel backup_meta. Password DB TIDAK PERNAH ditulis ke
//	          log / respons — hanya diteruskan sebagai env PGPASSWORD
//	          ke proses anak pg_dump / pg_restore.
//	Restore = mengembalikan satu dump ke database yang sama via
//	          pg_restore --clean --if-exists.
//
// Aturan keamanan keras yang diimplementasikan di sini:
//  1. Restore HANYA untuk permission backup.restore (admin + proktor
//     menurut seed/matrik RBAC; penegakan di server, bukan hanya di UI).
//  2. Sebelum restore, sistem OTOMATIS membuat backup pengaman
//     (jenis 'pre-restore') dan mencatatnya di backup_meta + audit.
//  3. Restore mencatat audit lengkap (siapa, kapan, file mana, IP).
//  4. Restore DITOLAK (409) bila ada jadwal berstatus aktif/pause atau
//     attempt berstatus berjalan — cegah restore saat ujian berlangsung.
//  5. File backup di ./storage/backup/, nama
//     backup-<YYYYMMDD-HHMMSS>-<8 karakter acak>.dump; file lama tidak
//     pernah ditimpa (nama selalu unik, kolom nama_file UNIQUE).
//  6. Download backup HANYA via endpoint dengan permission check —
//     direktori storage TIDAK disajikan sebagai static file server.
//  7. TIDAK ada scheduler di dalam aplikasi: backup otomatis harian
//     TIDAK dibuat; backup manual via UI + skrip cron deploy/backup.sh
//     yang sudah ada.
//  8. TIDAK ADA perintah shell dari input user: nama file selalu
//     diambil dari baris backup_meta di DB (berdasarkan :id UUID),
//     bukan dari parameter / body. Satu-satunya input user adalah
//     keterangan bebas (teks) dan :id.
//  9. Timeout pg_dump / pg_restore: 10 menit (context timeout).
//
// Subtleti restore yang penting: pg_restore --clean menghapus dan
// membuat ulang SEMUA tabel dari isi dump, termasuk backup_meta dan
// audit_log. Akibatnya baris metadata pre-restore backup + audit yang
// ditulis SEBELUM pg_restore akan ikut terhapus oleh restore itu
// sendiri. Karena itu, setelah pg_restore sukses, handler mengembalikan
// (upsert) baris metadata pre-restore backup + baris metadata backup
// target + audit backup.create (pre-restore) dan backup.restore —
// sehingga jejak "restore dari file X, pengaman file Y" tetap ada di
// database hasil restore. Bila pg_restore gagal, pre-restore backup
// tetap tercatat (restore belum mengubah apa pun secara utuh) dan
// respons 500 menyebut nama file pengamannya agar admin bisa
// memulihkan manual.
package backup

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"cbt-smk-v2/backend/db"
	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/config"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

const (
	// backupDir adalah direktori penyimpanan file dump, relatif terhadap
	// working directory proses server (lihat panduan deploy: jalankan
	// binary dari root repo agar menjadi ./storage/backup/).
	backupDir = "storage/backup"
	// dumpTimeout adalah timeout pg_dump / pg_restore.
	dumpTimeout = 10 * time.Minute
	// preRestoreMinAge adalah umur minimum pre-restore backup agar boleh
	// dihapus (7 hari). Di bawah itu hapus ditolak + dicatat alasannya.
	preRestoreMinAge = 7 * 24 * time.Hour
)

// dumpMu menyerikan operasi dump/restore/hapus: pg_restore tidak boleh
// berjalan bersamaan dengan pg_dump atau restore lain, dan penghapusan
// file tidak boleh balapan dengan restore yang memakai file itu.
var dumpMu sync.Mutex

// backupMeta adalah satu baris backup_meta (+ nama pembuat untuk list).
type backupMeta struct {
	ID          string    `json:"id"`
	NamaFile    string    `json:"nama_file"`
	UkuranBytes int64     `json:"ukuran_bytes"`
	Jenis       string    `json:"jenis"`
	DibuatOleh  *string   `json:"dibuat_oleh"`
	PembuatNama *string   `json:"pembuat_nama,omitempty"`
	Keterangan  *string   `json:"keterangan"`
	CreatedAt   time.Time `json:"created_at"`
}

// dbConn memegang parameter koneksi yang diteruskan ke pg_dump/pg_restore.
type dbConn struct {
	host     string
	port     string
	database string
	user     string
	password string
}

// parseDBConn membaca host/port/db/user/password dari DATABASE_URL di
// config (pgxpool.ParseConfig). Password HANYA dipakai sebagai env
// PGPASSWORD proses anak — tidak pernah di-log atau dikembalikan.
func parseDBConn(cfg *config.Config) (dbConn, error) {
	dsn := cfg.DBURL
	if dsn == "" {
		dsn = cfg.DatabaseURL
	}
	if dsn == "" {
		return dbConn{}, fmt.Errorf("DATABASE_URL tidak diisi")
	}
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return dbConn{}, fmt.Errorf("DATABASE_URL tidak valid: %w", err)
	}
	cc := poolCfg.ConnConfig
	return dbConn{
		host:     cc.Host,
		port:     strconv.Itoa(int(cc.Port)),
		database: cc.Database,
		user:     cc.User,
		password: cc.Password,
	}, nil
}

// connArgs mengubah dbConn menjadi argumen -h/-p/-U/-d pg_dump/pg_restore.
// Flag -w (--no-password) mencegah pg_dump/pg_restore berhenti menunggu
// password di stdin bila PGPASSWORD tidak tersedia.
func (c dbConn) connArgs() []string {
	args := []string{"-w"}
	if c.host != "" {
		args = append(args, "-h", c.host)
	}
	if c.port != "" {
		args = append(args, "-p", c.port)
	}
	if c.user != "" {
		args = append(args, "-U", c.user)
	}
	if c.database != "" {
		args = append(args, "-d", c.database)
	}
	return args
}

// cmdEnv meneruskan environment proses + PGPASSWORD (bila ada password).
func (c dbConn) cmdEnv() []string {
	env := os.Environ()
	if c.password != "" {
		env = append(env, "PGPASSWORD="+c.password)
	}
	return env
}

// randomSuffix menghasilkan 8 karakter acak [a-z0-9] untuk nama file.
func randomSuffix() (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b[:]), nil
}

// backupFileName menghasilkan nama file unik sesuai aturan penamaan.
func backupFileName() (string, error) {
	suffix, err := randomSuffix()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("backup-%s-%s.dump",
		time.Now().Format("20060102-150405"), suffix), nil
}

// safeBackupPath mengubah nama file dari DB menjadi path absolut di
// dalam backupDir. Menolak nama yang mengandung pemisah path / traversal
// walau sumbernya DB (defense in depth — input user tidak pernah dipakai
// sebagai nama file).
func safeBackupPath(namaFile string) (string, error) {
	if namaFile == "" || filepath.Base(namaFile) != namaFile ||
		strings.ContainsAny(namaFile, `/\`) {
		return "", fmt.Errorf("nama file backup tidak valid")
	}
	absDir, err := filepath.Abs(backupDir)
	if err != nil {
		return "", err
	}
	p := filepath.Join(absDir, namaFile)
	if !strings.HasPrefix(p, absDir+string(os.PathSeparator)) {
		return "", fmt.Errorf("nama file backup di luar direktori backup")
	}
	return p, nil
}

// getMeta membaca satu baris backup_meta. Tidak ditemukan -> (nil, nil).
func getMeta(ctx context.Context, pool *pgxpool.Pool, id string) (*backupMeta, error) {
	var m backupMeta
	err := pool.QueryRow(ctx,
		`SELECT bm.id, bm.nama_file, bm.ukuran_bytes, bm.jenis, bm.dibuat_oleh,
		        bm.keterangan, bm.created_at, u.nama
		 FROM backup_meta bm
		 LEFT JOIN users u ON u.id = bm.dibuat_oleh
		 WHERE bm.id = $1`, id).
		Scan(&m.ID, &m.NamaFile, &m.UkuranBytes, &m.Jenis, &m.DibuatOleh,
			&m.Keterangan, &m.CreatedAt, &m.PembuatNama)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

// insertMeta mencatat satu baris backup_meta.
func insertMeta(ctx context.Context, pool *pgxpool.Pool, m *backupMeta) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO backup_meta (id, nama_file, ukuran_bytes, jenis, dibuat_oleh, keterangan, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		m.ID, m.NamaFile, m.UkuranBytes, m.Jenis, strOrNil(m.DibuatOleh),
		strOrNil(m.Keterangan), m.CreatedAt)
	return err
}

func strOrNil(s *string) any {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil
	}
	return *s
}

// upsertMeta mengembalikan baris meta setelah restore (lihat catatan
// paket: pg_restore --clean menghapus baris yang ditulis sebelum
// restore). ON CONFLICT DO NOTHING: bila dump sudah berisi baris yang
// sama (mis. dump dibuat setelah meta itu tercatat), jangan duplikat.
func upsertMeta(ctx context.Context, pool *pgxpool.Pool, m *backupMeta) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO backup_meta (id, nama_file, ukuran_bytes, jenis, dibuat_oleh, keterangan, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (id) DO NOTHING`,
		m.ID, m.NamaFile, m.UkuranBytes, m.Jenis, strOrNil(m.DibuatOleh),
		strOrNil(m.Keterangan), m.CreatedAt)
	return err
}

// runDump menjalankan pg_dump -Fc ke path tujuan. Mengembalikan ukuran
// file. Bila pg_dump tidak ditemukan -> error berpesan jelas.
func runDump(ctx context.Context, conn dbConn, destPath string) (int64, error) {
	if _, err := exec.LookPath("pg_dump"); err != nil {
		return 0, fmt.Errorf("pg_dump tidak tersedia di server")
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o750); err != nil {
		return 0, err
	}
	dumpCtx, cancel := context.WithTimeout(ctx, dumpTimeout)
	defer cancel()
	args := append(conn.connArgs(), "-Fc", "-f", destPath)
	cmd := exec.CommandContext(dumpCtx, "pg_dump", args...)
	cmd.Env = conn.cmdEnv()
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		_ = os.Remove(destPath) // jangan sisakan dump parsial
		if dumpCtx.Err() == context.DeadlineExceeded {
			return 0, fmt.Errorf("pg_dump timeout setelah 10 menit")
		}
		return 0, fmt.Errorf("pg_dump gagal: %s", strings.TrimSpace(stderr.String()))
	}
	info, err := os.Stat(destPath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// doBackup membuat satu file dump + baris meta-nya. Dipakai POST /backup
// (jenis 'manual') dan sebagai pre-restore backup (jenis 'pre-restore').
// Keterangan hanya untuk 'manual'; pre-restore memakai keterangan baku.
func doBackup(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config, jenis, dibuatOleh, keterangan string) (*backupMeta, error) {
	if jenis != "manual" && jenis != "pre-restore" {
		return nil, fmt.Errorf("jenis backup tidak valid: %q", jenis)
	}
	conn, err := parseDBConn(cfg)
	if err != nil {
		return nil, err
	}
	namaFile, err := backupFileName()
	if err != nil {
		return nil, err
	}
	destPath, err := safeBackupPath(namaFile)
	if err != nil {
		return nil, err
	}
	size, err := runDump(ctx, conn, destPath)
	if err != nil {
		return nil, err
	}
	m := &backupMeta{
		ID:          uuid.NewString(),
		NamaFile:    namaFile,
		UkuranBytes: size,
		Jenis:       jenis,
		CreatedAt:   time.Now(),
	}
	if dibuatOleh != "" {
		v := dibuatOleh
		m.DibuatOleh = &v
	}
	if keterangan != "" {
		v := keterangan
		m.Keterangan = &v
	}
	if err := insertMeta(ctx, pool, m); err != nil {
		_ = os.Remove(destPath)
		return nil, err
	}
	return m, nil
}

// examInProgress memeriksa gerbang restore: jadwal aktif/pause atau
// attempt berjalan. Mengembalikan pesan penolakan bila ada yang
// menghalangi, string kosong bila aman.
func examInProgress(ctx context.Context, pool *pgxpool.Pool) (string, error) {
	var jadwalAktif int
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM jadwal_ujian WHERE status IN ('aktif', 'pause')`).
		Scan(&jadwalAktif); err != nil {
		return "", err
	}
	if jadwalAktif > 0 {
		return fmt.Sprintf("Restore ditolak: ada %d jadwal ujian berstatus aktif/pause. "+
			"Selesaikan atau ubah status jadwal tersebut terlebih dahulu, lalu ulangi restore.", jadwalAktif), nil
	}
	var attemptBerjalan int
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM attempt WHERE status = 'berjalan'`).
		Scan(&attemptBerjalan); err != nil {
		return "", err
	}
	if attemptBerjalan > 0 {
		return fmt.Sprintf("Restore ditolak: ada %d attempt ujian berstatus berjalan. "+
			"Tunggu sampai semua peserta selesai / diblokir, lalu ulangi restore.", attemptBerjalan), nil
	}
	return "", nil
}

func writeError(c *echo.Context, status int, code, message string) error {
	return c.JSON(status, map[string]any{
		"error": map[string]any{"code": code, "message": message},
	})
}

func writeOK(c *echo.Context, data any) error {
	return c.JSON(http.StatusOK, map[string]any{"data": data})
}

// Register mendaftarkan endpoint backup / restore.
//
//	POST   /api/v1/backup                 (backup.create)
//	GET    /api/v1/backup                 (backup.create ATAU backup.restore)
//	GET    /api/v1/backup/:id/download    (backup.create ATAU backup.restore)
//	POST   /api/v1/backup/:id/restore     (backup.restore)
//	DELETE /api/v1/backup/:id             (backup.restore)
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config) {
	_ = q

	// POST /api/v1/backup — buat backup manual.
	e.POST("/api/v1/backup", func(c *echo.Context) error {
		if !audit.HasPermission(c, "backup.create") {
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin membuat backup.")
		}
		var body struct {
			Keterangan string `json:"keterangan"`
		}
		_ = c.Bind(&body) // keterangan opsional; body kosong tetap sah
		if len(body.Keterangan) > 500 {
			return writeError(c, http.StatusBadRequest, "validation_error", "Keterangan maksimal 500 karakter.")
		}
		userID, _, _, _, _ := audit.CurrentUser(c)
		ctx := c.Request().Context()

		dumpMu.Lock()
		m, err := doBackup(ctx, pool, cfg, "manual", userID, strings.TrimSpace(body.Keterangan))
		dumpMu.Unlock()
		if err != nil {
			if strings.Contains(err.Error(), "pg_dump tidak tersedia") {
				return writeError(c, http.StatusInternalServerError, "internal_error", "pg_dump tidak tersedia di server — backup tidak dapat dibuat. Pasang paket PostgreSQL client di server.")
			}
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membuat backup: "+err.Error())
		}
		_ = audit.Catat(ctx, pool, userID, "backup.create", "backup", m.ID,
			map[string]any{"nama_file": m.NamaFile, "ukuran_bytes": m.UkuranBytes, "jenis": m.Jenis}, c.RealIP())
		return writeOK(c, map[string]any{
			"id": m.ID, "nama_file": m.NamaFile, "ukuran_bytes": m.UkuranBytes,
			"jenis": m.Jenis, "created_at": m.CreatedAt,
		})
	})

	// GET /api/v1/backup — daftar metadata + pagination.
	e.GET("/api/v1/backup", func(c *echo.Context) error {
		if !audit.HasPermission(c, "backup.create") && !audit.HasPermission(c, "backup.restore") {
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melihat daftar backup.")
		}
		limit, err := strconv.Atoi(c.QueryParam("limit"))
		if err != nil || limit <= 0 {
			limit = 50
		}
		if limit > 200 {
			limit = 200
		}
		offset, err := strconv.Atoi(c.QueryParam("offset"))
		if err != nil || offset < 0 {
			offset = 0
		}
		ctx := c.Request().Context()
		rows, err := pool.Query(ctx,
			`SELECT bm.id, bm.nama_file, bm.ukuran_bytes, bm.jenis, bm.dibuat_oleh,
			        bm.keterangan, bm.created_at, u.nama
			 FROM backup_meta bm
			 LEFT JOIN users u ON u.id = bm.dibuat_oleh
			 ORDER BY bm.created_at DESC
			 LIMIT $1 OFFSET $2`, limit, offset)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca daftar backup.")
		}
		defer rows.Close()
		items := []backupMeta{}
		for rows.Next() {
			var m backupMeta
			if err := rows.Scan(&m.ID, &m.NamaFile, &m.UkuranBytes, &m.Jenis,
				&m.DibuatOleh, &m.Keterangan, &m.CreatedAt, &m.PembuatNama); err != nil {
				return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca daftar backup.")
			}
			items = append(items, m)
		}
		if err := rows.Err(); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca daftar backup.")
		}
		var total int64
		if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM backup_meta`).Scan(&total); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca daftar backup.")
		}
		return c.JSON(http.StatusOK, map[string]any{
			"data": items,
			"meta": map[string]any{"limit": limit, "offset": offset, "total": total},
		})
	})

	// GET /api/v1/backup/:id/download — unduh file via permission check
	// + audit. Streaming; file TIDAK disajikan lewat static file server.
	e.GET("/api/v1/backup/:id/download", func(c *echo.Context) error {
		if !audit.HasPermission(c, "backup.create") && !audit.HasPermission(c, "backup.restore") {
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin mengunduh backup.")
		}
		id := c.Param("id")
		if _, err := uuid.Parse(id); err != nil {
			return writeError(c, http.StatusNotFound, "not_found", "Backup tidak ditemukan.")
		}
		m, err := getMeta(c.Request().Context(), pool, id)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca metadata backup.")
		}
		if m == nil {
			return writeError(c, http.StatusNotFound, "not_found", "Backup tidak ditemukan.")
		}
		path, err := safeBackupPath(m.NamaFile)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Nama file backup tidak valid.")
		}
		f, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				return writeError(c, http.StatusNotFound, "not_found", "File backup tidak ditemukan di server (metadata ada, file hilang).")
			}
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membuka file backup.")
		}
		defer f.Close()
		info, err := f.Stat()
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca file backup.")
		}
		userID, _, _, _, _ := audit.CurrentUser(c)
		_ = audit.Catat(c.Request().Context(), pool, userID, "backup.download", "backup", m.ID,
			map[string]any{"nama_file": m.NamaFile, "ukuran_bytes": m.UkuranBytes}, c.RealIP())
		resp := c.Response()
		resp.Header().Set("Content-Disposition", `attachment; filename="`+m.NamaFile+`"`)
		resp.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
		return c.Stream(http.StatusOK, "application/octet-stream", f)
	})

	// POST /api/v1/backup/:id/restore — pulihkan dump ke DB yang sama.
	e.POST("/api/v1/backup/:id/restore", func(c *echo.Context) error {
		if !audit.HasPermission(c, "backup.restore") {
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin melakukan restore. Restore hanya untuk admin dan proktor.")
		}
		id := c.Param("id")
		if _, err := uuid.Parse(id); err != nil {
			return writeError(c, http.StatusNotFound, "not_found", "Backup tidak ditemukan.")
		}
		userID, _, _, _, _ := audit.CurrentUser(c)
		ctx := c.Request().Context()
		ip := c.RealIP()

		dumpMu.Lock()
		defer dumpMu.Unlock()

		target, err := getMeta(ctx, pool, id)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca metadata backup.")
		}
		if target == nil {
			return writeError(c, http.StatusNotFound, "not_found", "Backup tidak ditemukan.")
		}
		targetPath, err := safeBackupPath(target.NamaFile)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Nama file backup tidak valid.")
		}
		if _, err := os.Stat(targetPath); err != nil {
			if os.IsNotExist(err) {
				return writeError(c, http.StatusNotFound, "not_found", "File backup tidak ditemukan di server (metadata ada, file hilang).")
			}
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca file backup.")
		}
		if _, err := exec.LookPath("pg_restore"); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "pg_restore tidak tersedia di server — restore tidak dapat dilakukan. Pasang paket PostgreSQL client di server.")
		}

		// Gerbang (4): tolak bila ujian sedang berlangsung.
		if blocked, err := examInProgress(ctx, pool); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal memeriksa status ujian sebelum restore.")
		} else if blocked != "" {
			_ = audit.Catat(ctx, pool, userID, "backup.restore.ditolak", "backup", target.ID,
				map[string]any{"nama_file": target.NamaFile, "alasan": blocked}, ip)
			return writeError(c, http.StatusConflict, "conflict", blocked)
		}

		// Gerbang (2): backup pengaman otomatis SEBELUM restore.
		pre, err := doBackup(ctx, pool, cfg, "pre-restore", userID, "Backup pengaman otomatis sebelum restore dari "+target.NamaFile)
		if err != nil {
			if strings.Contains(err.Error(), "pg_dump tidak tersedia") {
				return writeError(c, http.StatusInternalServerError, "internal_error", "pg_dump tidak tersedia di server — backup pengaman tidak dapat dibuat, restore dibatalkan demi keamanan data.")
			}
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membuat backup pengaman sebelum restore: "+err.Error())
		}
		_ = audit.Catat(ctx, pool, userID, "backup.create", "backup", pre.ID,
			map[string]any{"nama_file": pre.NamaFile, "ukuran_bytes": pre.UkuranBytes, "jenis": "pre-restore"}, ip)

		// Jalankan pg_restore --clean --if-exists ke DB yang sama.
		// --no-owner/--no-privileges: restore tidak gagal bila owner/
		// privilege di dump berbeda dengan user koneksi saat ini.
		conn, err := parseDBConn(cfg)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca konfigurasi database: "+err.Error())
		}
		restoreCtx, cancel := context.WithTimeout(ctx, dumpTimeout)
		defer cancel()
		args := append(conn.connArgs(), "--clean", "--if-exists",
			"--no-owner", "--no-privileges", "-d", conn.database, targetPath)
		cmd := exec.CommandContext(restoreCtx, "pg_restore", args...)
		cmd.Env = conn.cmdEnv()
		var stderr strings.Builder
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			// Pre-restore backup tetap tercatat (restore belum berhasil
			// menimpa data) — sebutkan namanya agar admin bisa pulih manual.
			_ = audit.Catat(ctx, pool, userID, "backup.restore.gagal", "backup", target.ID,
				map[string]any{"nama_file": target.NamaFile, "pre_restore_backup": pre.NamaFile, "error": strings.TrimSpace(stderr.String())}, ip)
			if restoreCtx.Err() == context.DeadlineExceeded {
				return writeError(c, http.StatusInternalServerError, "internal_error", "pg_restore timeout setelah 10 menit. Data saat ini mungkin sebagian terpulihkan. Backup pengaman tersedia di file "+pre.NamaFile+" — pulihkan manual bila perlu.")
			}
			return writeError(c, http.StatusInternalServerError, "internal_error", "Restore gagal: "+strings.TrimSpace(stderr.String())+". Backup pengaman tersedia di file "+pre.NamaFile+" — pulihkan manual bila perlu.")
		}

		// pg_restore --clean telah menghapus baris meta + audit yang
		// ditulis sebelum restore: kembalikan jejaknya ke DB hasil restore.
		_ = upsertMeta(ctx, pool, pre)
		_ = upsertMeta(ctx, pool, target)
		_ = audit.Catat(ctx, pool, userID, "backup.create", "backup", pre.ID,
			map[string]any{"nama_file": pre.NamaFile, "ukuran_bytes": pre.UkuranBytes, "jenis": "pre-restore",
				"keterangan": "Backup pengaman otomatis sebelum restore dari " + target.NamaFile}, ip)
		_ = audit.Catat(ctx, pool, userID, "backup.restore", "backup", target.ID,
			map[string]any{"nama_file": target.NamaFile, "dipulihkan_dari": target.NamaFile,
				"pre_restore_backup": pre.NamaFile, "pre_restore_backup_id": pre.ID}, ip)
		return writeOK(c, map[string]any{
			"dipulihkan_dari": target.NamaFile,
			"pre_restore_backup": map[string]any{
				"id": pre.ID, "nama_file": pre.NamaFile, "ukuran_bytes": pre.UkuranBytes,
				"jenis": pre.Jenis, "created_at": pre.CreatedAt,
			},
		})
	})

	// DELETE /api/v1/backup/:id — hapus file + meta + audit. Pre-restore
	// backup berusia < 7 hari DITOLAK (pengaman restore terakhir harus
	// tetap tersedia) dan penolakan dicatat alasannya di audit.
	e.DELETE("/api/v1/backup/:id", func(c *echo.Context) error {
		if !audit.HasPermission(c, "backup.restore") {
			return writeError(c, http.StatusForbidden, "forbidden", "Tidak punya izin menghapus backup.")
		}
		id := c.Param("id")
		if _, err := uuid.Parse(id); err != nil {
			return writeError(c, http.StatusNotFound, "not_found", "Backup tidak ditemukan.")
		}
		userID, _, _, _, _ := audit.CurrentUser(c)
		ctx := c.Request().Context()
		ip := c.RealIP()

		dumpMu.Lock()
		defer dumpMu.Unlock()

		m, err := getMeta(ctx, pool, id)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal membaca metadata backup.")
		}
		if m == nil {
			return writeError(c, http.StatusNotFound, "not_found", "Backup tidak ditemukan.")
		}
		if m.Jenis == "pre-restore" && time.Since(m.CreatedAt) < preRestoreMinAge {
			alasan := fmt.Sprintf("Pre-restore backup berusia %s (kurang dari 7 hari) tidak boleh dihapus — pengaman restore terakhir harus tetap tersedia.", time.Since(m.CreatedAt).Round(time.Minute))
			_ = audit.Catat(ctx, pool, userID, "backup.delete.ditolak", "backup", m.ID,
				map[string]any{"nama_file": m.NamaFile, "jenis": m.Jenis, "alasan": alasan}, ip)
			return writeError(c, http.StatusConflict, "conflict", alasan)
		}
		path, err := safeBackupPath(m.NamaFile)
		if err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Nama file backup tidak valid.")
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghapus file backup.")
		}
		if _, err := pool.Exec(ctx, `DELETE FROM backup_meta WHERE id = $1`, m.ID); err != nil {
			return writeError(c, http.StatusInternalServerError, "internal_error", "Gagal menghapus metadata backup.")
		}
		_ = audit.Catat(ctx, pool, userID, "backup.delete", "backup", m.ID,
			map[string]any{"nama_file": m.NamaFile, "jenis": m.Jenis, "ukuran_bytes": m.UkuranBytes}, ip)
		return writeOK(c, map[string]any{"deleted": true, "id": m.ID, "nama_file": m.NamaFile})
	})
}
