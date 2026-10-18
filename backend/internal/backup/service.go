package backup

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/config"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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
// (Dipindah dari backup.go tanpa perubahan.)
var dumpMu sync.Mutex

// dbConn memegang parameter koneksi yang diteruskan ke pg_dump/pg_restore.
// (Dipindah dari backup.go tanpa perubahan.)
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
		return "", fmt.Errorf("nama file backup di luar direktori backup")
	}
	p := filepath.Join(absDir, namaFile)
	if !strings.HasPrefix(p, absDir+string(os.PathSeparator)) {
		return "", fmt.Errorf("nama file backup di luar direktori backup")
	}
	return p, nil
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

// --- Service ----------------------------------------------------------------

// RequestActor adalah identitas pemanggil untuk audit log.
// Dibangun handler dari echo.Context agar service bebas dari HTTP.
type RequestActor struct {
	UserID string
	RealIP string
}

// Service adalah lapisan logika bisnis paket backup.
type Service struct {
	repo *Repository
	pool *pgxpool.Pool
	cfg  *config.Config
}

// NewService membuat Service.
func NewService(repo *Repository, pool *pgxpool.Pool, cfg *config.Config) *Service {
	return &Service{repo: repo, pool: pool, cfg: cfg}
}

// catat menulis audit log. (Dulu: audit.Catat di handler.)
func (s *Service) catat(ctx context.Context, actor RequestActor, aksi, entitas, entitasID string, detail any) {
	_ = audit.Catat(ctx, s.pool, actor.UserID, aksi, entitas, entitasID, detail, actor.RealIP)
}

// doBackup membuat satu file dump + baris meta-nya. Dipakai POST /backup
// (jenis 'manual') dan sebagai pre-restore backup (jenis 'pre-restore').
// Keterangan hanya untuk 'manual'; pre-restore memakai keterangan baku.
// (Dulu: fungsi paket doBackup; pesan error pemanggil dipertahankan.)
func (s *Service) doBackup(ctx context.Context, jenis, dibuatOleh, keterangan string) (*backupMeta, error) {
	if jenis != "manual" && jenis != "pre-restore" {
		return nil, fmt.Errorf("jenis backup tidak valid: %q", jenis)
	}
	conn, err := parseDBConn(s.cfg)
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
	if err := s.repo.InsertMeta(ctx, m); err != nil {
		_ = os.Remove(destPath)
		return nil, err
	}
	return m, nil
}

// Buat membuat backup manual + mencatat audit.
func (s *Service) Buat(ctx context.Context, actor RequestActor, keterangan string) (map[string]any, *platform.AppError) {
	dumpMu.Lock()
	m, err := s.doBackup(ctx, "manual", actor.UserID, keterangan)
	dumpMu.Unlock()
	if err != nil {
		if strings.Contains(err.Error(), "pg_dump tidak tersedia") {
			return nil, platform.ErrInternal("internal_error", "pg_dump tidak tersedia di server — backup tidak dapat dibuat. Pasang paket PostgreSQL client di server.")
		}
		return nil, platform.ErrInternal("internal_error", "Gagal membuat backup: "+err.Error())
	}
	s.catat(ctx, actor, "backup.create", "backup", m.ID,
		map[string]any{"nama_file": m.NamaFile, "ukuran_bytes": m.UkuranBytes, "jenis": m.Jenis})
	return map[string]any{
		"id": m.ID, "nama_file": m.NamaFile, "ukuran_bytes": m.UkuranBytes,
		"jenis": m.Jenis, "created_at": m.CreatedAt,
	}, nil
}

// Daftar mengembalikan halaman metadata backup + total.
func (s *Service) Daftar(ctx context.Context, limit, offset int) ([]backupMeta, int64, *platform.AppError) {
	items, err := s.repo.ListMeta(ctx, limit, offset)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca daftar backup.")
	}
	total, err := s.repo.CountMeta(ctx)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca daftar backup.")
	}
	return items, total, nil
}

// BerkasUnduhan adalah file backup siap di-stream + info audit.
type BerkasUnduhan struct {
	Meta *backupMeta
	File *os.File
	Size int64
}

// Unduh menyiapkan file backup untuk diunduh + mencatat audit download.
// File HARUS ditutup pemanggil setelah selesai di-stream.
func (s *Service) Unduh(ctx context.Context, actor RequestActor, id string) (*BerkasUnduhan, *platform.AppError) {
	m, err := s.repo.GetMeta(ctx, id)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca metadata backup.")
	}
	if m == nil {
		return nil, platform.ErrNotFound("not_found", "Backup tidak ditemukan.")
	}
	path, err := safeBackupPath(m.NamaFile)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Nama file backup tidak valid.")
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, platform.ErrNotFound("not_found", "File backup tidak ditemukan di server (metadata ada, file hilang).")
		}
		return nil, platform.ErrInternal("internal_error", "Gagal membuka file backup.")
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, platform.ErrInternal("internal_error", "Gagal membaca file backup.")
	}
	s.catat(ctx, actor, "backup.download", "backup", m.ID,
		map[string]any{"nama_file": m.NamaFile, "ukuran_bytes": m.UkuranBytes})
	return &BerkasUnduhan{Meta: m, File: f, Size: info.Size()}, nil
}

// Restore memulihkan satu dump ke database yang sama, diawali backup
// pengaman otomatis. Lihat catatan paket soal subtleti pg_restore
// --clean yang menghapus baris meta/audit sebelum restore.
func (s *Service) Restore(ctx context.Context, actor RequestActor, id string) (map[string]any, *platform.AppError) {
	dumpMu.Lock()
	defer dumpMu.Unlock()

	target, err := s.repo.GetMeta(ctx, id)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca metadata backup.")
	}
	if target == nil {
		return nil, platform.ErrNotFound("not_found", "Backup tidak ditemukan.")
	}
	targetPath, err := safeBackupPath(target.NamaFile)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Nama file backup tidak valid.")
	}
	if _, err := os.Stat(targetPath); err != nil {
		if os.IsNotExist(err) {
			return nil, platform.ErrNotFound("not_found", "File backup tidak ditemukan di server (metadata ada, file hilang).")
		}
		return nil, platform.ErrInternal("internal_error", "Gagal membaca file backup.")
	}
	if _, err := exec.LookPath("pg_restore"); err != nil {
		return nil, platform.ErrInternal("internal_error", "pg_restore tidak tersedia di server — restore tidak dapat dilakukan. Pasang paket PostgreSQL client di server.")
	}

	// Gerbang (4): tolak bila ujian sedang berlangsung.
	if blocked, err := s.repo.ExamInProgress(ctx); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal memeriksa status ujian sebelum restore.")
	} else if blocked != "" {
		s.catat(ctx, actor, "backup.restore.ditolak", "backup", target.ID,
			map[string]any{"nama_file": target.NamaFile, "alasan": blocked})
		return nil, platform.ErrConflict("conflict", blocked)
	}

	// Gerbang (2): backup pengaman otomatis SEBELUM restore.
	pre, err := s.doBackup(ctx, "pre-restore", actor.UserID, "Backup pengaman otomatis sebelum restore dari "+target.NamaFile)
	if err != nil {
		if strings.Contains(err.Error(), "pg_dump tidak tersedia") {
			return nil, platform.ErrInternal("internal_error", "pg_dump tidak tersedia di server — backup pengaman tidak dapat dibuat, restore dibatalkan demi keamanan data.")
		}
		return nil, platform.ErrInternal("internal_error", "Gagal membuat backup pengaman sebelum restore: "+err.Error())
	}
	s.catat(ctx, actor, "backup.create", "backup", pre.ID,
		map[string]any{"nama_file": pre.NamaFile, "ukuran_bytes": pre.UkuranBytes, "jenis": "pre-restore"})

	// Jalankan pg_restore --clean --if-exists ke DB yang sama.
	// --no-owner/--no-privileges: restore tidak gagal bila owner/
	// privilege di dump berbeda dengan user koneksi saat ini.
	conn, err := parseDBConn(s.cfg)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca konfigurasi database: "+err.Error())
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
		s.catat(ctx, actor, "backup.restore.gagal", "backup", target.ID,
			map[string]any{"nama_file": target.NamaFile, "pre_restore_backup": pre.NamaFile, "error": strings.TrimSpace(stderr.String())})
		if restoreCtx.Err() == context.DeadlineExceeded {
			return nil, platform.ErrInternal("internal_error", "pg_restore timeout setelah 10 menit. Data saat ini mungkin sebagian terpulihkan. Backup pengaman tersedia di file "+pre.NamaFile+" — pulihkan manual bila perlu.")
		}
		return nil, platform.ErrInternal("internal_error", "Restore gagal: "+strings.TrimSpace(stderr.String())+". Backup pengaman tersedia di file "+pre.NamaFile+" — pulihkan manual bila perlu.")
	}

	// pg_restore --clean telah menghapus baris meta + audit yang
	// ditulis sebelum restore: kembalikan jejaknya ke DB hasil restore.
	_ = s.repo.UpsertMeta(ctx, pre)
	_ = s.repo.UpsertMeta(ctx, target)
	s.catat(ctx, actor, "backup.create", "backup", pre.ID,
		map[string]any{"nama_file": pre.NamaFile, "ukuran_bytes": pre.UkuranBytes, "jenis": "pre-restore",
			"keterangan": "Backup pengaman otomatis sebelum restore dari " + target.NamaFile})
	s.catat(ctx, actor, "backup.restore", "backup", target.ID,
		map[string]any{"nama_file": target.NamaFile, "dipulihkan_dari": target.NamaFile,
			"pre_restore_backup": pre.NamaFile, "pre_restore_backup_id": pre.ID})
	return map[string]any{
		"dipulihkan_dari": target.NamaFile,
		"pre_restore_backup": map[string]any{
			"id": pre.ID, "nama_file": pre.NamaFile, "ukuran_bytes": pre.UkuranBytes,
			"jenis": pre.Jenis, "created_at": pre.CreatedAt,
		},
	}, nil
}

// Hapus menghapus file + metadata backup. Pre-restore backup berusia
// < 7 hari DITOLAK (pengaman restore terakhir harus tetap tersedia)
// dan penolakan dicatat alasannya di audit.
func (s *Service) Hapus(ctx context.Context, actor RequestActor, id string) (map[string]any, *platform.AppError) {
	dumpMu.Lock()
	defer dumpMu.Unlock()

	m, err := s.repo.GetMeta(ctx, id)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca metadata backup.")
	}
	if m == nil {
		return nil, platform.ErrNotFound("not_found", "Backup tidak ditemukan.")
	}
	if m.Jenis == "pre-restore" && time.Since(m.CreatedAt) < preRestoreMinAge {
		alasan := fmt.Sprintf("Pre-restore backup berusia %s (kurang dari 7 hari) tidak boleh dihapus — pengaman restore terakhir harus tetap tersedia.", time.Since(m.CreatedAt).Round(time.Minute))
		s.catat(ctx, actor, "backup.delete.ditolak", "backup", m.ID,
			map[string]any{"nama_file": m.NamaFile, "jenis": m.Jenis, "alasan": alasan})
		return nil, platform.ErrConflict("conflict", alasan)
	}
	path, err := safeBackupPath(m.NamaFile)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Nama file backup tidak valid.")
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return nil, platform.ErrInternal("internal_error", "Gagal menghapus file backup.")
	}
	if err := s.repo.DeleteMeta(ctx, m.ID); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menghapus metadata backup.")
	}
	s.catat(ctx, actor, "backup.delete", "backup", m.ID,
		map[string]any{"nama_file": m.NamaFile, "jenis": m.Jenis, "ukuran_bytes": m.UkuranBytes})
	return map[string]any{"deleted": true, "id": m.ID, "nama_file": m.NamaFile}, nil
}
