package auth

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"cbt-smk-v2/backend/internal/platform"
	"cbt-smk-v2/backend/internal/session"
)

// Panjang password minimal.
const minPasswordLength = 8

// --- rate limit login (anti brute force) -----------------------------------
// Aturan (ANALISIS Bab 7, T13): maks 5 gagal per akun per 15 menit,
// percobaan ke-6 -> 429 + kunci 15 menit. In-memory per instance.
// (Dipindah dari auth.go tanpa perubahan.)
type loginAttempt struct {
	failures   int
	firstFail  time.Time
	lockedTill time.Time
}

var (
	loginAttemptsMu sync.Mutex
	loginAttempts   = make(map[string]*loginAttempt)
)

const (
	maxLoginFailures = 5
	loginLockWindow  = 15 * time.Minute
)

func loginLocked(key string) bool {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()
	a, ok := loginAttempts[key]
	return ok && time.Now().Before(a.lockedTill)
}

func loginRecordFailure(key string) {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()
	now := time.Now()
	a, ok := loginAttempts[key]
	if !ok || now.Sub(a.firstFail) > loginLockWindow {
		a = &loginAttempt{firstFail: now}
		loginAttempts[key] = a
	}
	a.failures++
	if a.failures >= maxLoginFailures {
		a.lockedTill = now.Add(loginLockWindow)
	}
}

func loginReset(key string) {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()
	delete(loginAttempts, key)
}

// --- Service ----------------------------------------------------------------


// Service adalah lapisan logika bisnis paket auth.
// repository adalah kontrak penyimpanan yang dipakai Service.
// (DIP: Service bergantung pada abstraksi, bukan *Repository konkrit.)
type repository interface {
	ByID(ctx context.Context, id uuid.UUID) (*userRow, error)
	ByLogin(ctx context.Context, login string) (*userRow, error)
	PermissionsForRole(ctx context.Context, role string) []string
	UpdatePasswordGanti(ctx context.Context, id uuid.UUID, hash string) error
	UpdateTerakhirMasuk(ctx context.Context, id uuid.UUID)
}

type Service struct {
	repo repository
	pool *pgxpool.Pool
}

// NewService membuat Service.
func NewService(repo repository, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, pool: pool}
}


// Login memproses login: validasi -> rate limit -> verifikasi password ->
// guard aktif/wajib-ganti -> buat session. Mengembalikan sessionID +
// csrfToken + payload respons (user + permissions).
// (Dipindah dari closure Register tanpa perubahan perilaku.)
func (s *Service) Login(ctx context.Context, actor platform.RequestActor, login, email, username, password string) (sessionID, csrfToken string, data map[string]any, appErr *platform.AppError) {
	if login == "" {
		login = email
	}
	if login == "" {
		login = username
	}
	if strings.TrimSpace(login) == "" || password == "" {
		return "", "", nil, platform.ErrBadRequest("validation_error", "Email/username dan password wajib diisi.")
	}
	key := strings.ToLower(strings.TrimSpace(login))
	if loginLocked(key) {
		return "", "", nil, platform.NewAppError("login_locked", "Akun terkunci sementara karena terlalu banyak percobaan login gagal. Coba lagi dalam 15 menit.", 429)
	}
	u, err := s.repo.ByLogin(ctx, login)
	if err != nil {
		loginRecordFailure(key)
		platform.Catat(ctx, s.pool, actor, "auth.login.gagal", "users", "",
			map[string]any{"login": login, "alasan": "akun tidak ditemukan"})
		return "", "", nil, platform.ErrUnauthorized("invalid_credentials", "Email/username atau password salah.")
	}
	ok, err := VerifyPassword(password, u.PasswordHash)
	if err != nil || !ok {
		loginRecordFailure(key)
		platform.Catat(ctx, s.pool, platform.RequestActor{UserID: u.ID.String(), RealIP: actor.RealIP}, "auth.login.gagal", "users", u.ID.String(),
			map[string]any{"alasan": "password salah"})
		return "", "", nil, platform.ErrUnauthorized("invalid_credentials", "Email/username atau password salah.")
	}
	if !u.IsActive {
		return "", "", nil, platform.ErrForbidden("account_inactive", "Akun tidak aktif. Hubungi admin.")
	}
	if u.MustChangePassword {
		// TIDAK dibuatkan session: pengguna harus ganti password dulu
		// lewat POST /auth/ganti-password (mode kredensial).
		return "", "", nil, platform.ErrForbidden("PASSWORD_CHANGE_REQUIRED", "Password harus diganti sebelum login.")
	}
	loginReset(key)
	sessionID, csrfToken, err = session.Create(ctx, s.pool, u.ID, u.Role, actor.RealIP, actor.UserAgent)
	if err != nil {
		return "", "", nil, platform.ErrInternal("internal_error", "Gagal membuat sesi.")
	}
	// UC20: catat waktu login terakhir. Kegagalan UPDATE diabaikan
	// (tidak boleh menggagalkan login yang sudah sukses).
	s.repo.UpdateTerakhirMasuk(ctx, u.ID)
	platform.Catat(ctx, s.pool, platform.RequestActor{UserID: u.ID.String(), RealIP: actor.RealIP}, "auth.login", "users", u.ID.String(),
		map[string]any{"role": u.Role})
	return sessionID, csrfToken, map[string]any{
		"user":        userPublic(u),
		"permissions": s.repo.PermissionsForRole(ctx, u.Role),
	}, nil
}

// Logout menghancurkan session (bila ada) + mencatat audit.
func (s *Service) Logout(ctx context.Context, actor platform.RequestActor, sessionCookie string) {
	if sessionCookie != "" {
		_ = session.Destroy(ctx, s.pool, sessionCookie)
	}
	platform.Catat(ctx, s.pool, actor, "auth.logout", "users", actor.UserID, nil)
}

// Me mengembalikan user + permissions untuk session aktif.
func (s *Service) Me(ctx context.Context, actor platform.RequestActor) (map[string]any, *platform.AppError) {
	id, err := uuid.Parse(actor.UserID)
	if err != nil {
		return nil, platform.ErrUnauthorized("unauthenticated", "Harus login dulu.")
	}
	u, err := s.repo.ByID(ctx, id)
	if err != nil || !u.IsActive {
		return nil, platform.ErrUnauthorized("unauthenticated", "Akun tidak ditemukan / tidak aktif.")
	}
	return map[string]any{
		"user":        userPublic(u),
		"permissions": s.repo.PermissionsForRole(ctx, u.Role),
	}, nil
}

// GantiPassword memproses ganti password dalam dua mode:
// (a) Mode session: actor.UserID terisi (cookie valid) — cukup
//
//	password_lama + password_baru.
//
// (b) Mode kredensial: tanpa session — login + password_lama +
//
//	password_baru. Setelah berhasil, SEMUA session user dihapus.
func (s *Service) GantiPassword(ctx context.Context, actor platform.RequestActor, login, passwordLama, passwordBaru string) *platform.AppError {
	if len(passwordBaru) < minPasswordLength {
		return platform.ErrBadRequest("validation_error", "Password baru minimal 8 karakter.")
	}
	if passwordBaru == passwordLama {
		return platform.ErrBadRequest("validation_error", "Password baru harus berbeda dari password lama.")
	}
	var u *userRow
	// Mode (a): session aktif.
	if actor.UserID != "" {
		if id, err := uuid.Parse(actor.UserID); err == nil {
			u, _ = s.repo.ByID(ctx, id)
		}
	}
	// Mode (b): kredensial.
	if u == nil {
		if strings.TrimSpace(login) == "" {
			return platform.ErrUnauthorized("unauthenticated", "Harus login dulu atau sertakan email/username.")
		}
		var err error
		u, err = s.repo.ByLogin(ctx, login)
		if err != nil {
			return platform.ErrUnauthorized("invalid_credentials", "Email/username atau password lama salah.")
		}
	}
	ok, err := VerifyPassword(passwordLama, u.PasswordHash)
	if err != nil || !ok {
		return platform.ErrUnauthorized("invalid_credentials", "Password lama salah.")
	}
	hash, err := HashPassword(passwordBaru)
	if err != nil {
		return platform.ErrInternal("internal_error", "Gagal meng-hash password.")
	}
	if err := s.repo.UpdatePasswordGanti(ctx, u.ID, hash); err != nil {
		return platform.ErrInternal("internal_error", "Gagal menyimpan password baru.")
	}
	_ = session.DestroyAllForUser(ctx, s.pool, u.ID)
	platform.Catat(ctx, s.pool, platform.RequestActor{UserID: u.ID.String(), RealIP: actor.RealIP}, "auth.ganti-password", "users", u.ID.String(), nil)
	return nil
}

// --- CLI akun admin ---------------------------------------------------------
// (Dipindah dari auth.go tanpa perubahan perilaku; signature EXPORT
// dipertahankan: CreateAdmin, ResetPassword, ErrUsersNotEmpty.)

// ErrUsersNotEmpty berarti tabel users sudah berisi akun (create-admin
// hanya diizinkan saat tabel users kosong).
var ErrUsersNotEmpty = errors.New("tabel users sudah berisi akun — create-admin hanya diizinkan saat users kosong")

// CreateAdmin membuat akun admin pertama. Hanya bila tabel users kosong.
// must_change_password=true agar admin wajib ganti password saat login pertama.
func CreateAdmin(ctx context.Context, pool *pgxpool.Pool, email, username, nama, password string) error {
	repo := NewRepository(pool)
	count, err := repo.CountUsers(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrUsersNotEmpty
	}
	if strings.TrimSpace(nama) == "" {
		return errors.New("nama wajib diisi")
	}
	if len(password) < minPasswordLength {
		return errors.New("password minimal 8 karakter")
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	var emailVal, usernameVal any
	if strings.TrimSpace(email) != "" {
		emailVal = strings.TrimSpace(email)
	}
	if strings.TrimSpace(username) != "" {
		usernameVal = strings.TrimSpace(username)
	}
	if emailVal == nil && usernameVal == nil {
		return errors.New("email atau username wajib diisi")
	}
	return repo.InsertAdmin(ctx, nama, emailVal, usernameVal, hash)
}

// ResetPassword mengganti password akun (berdasarkan email/username),
// mengaktifkan akun, menandai must_change_password=true, dan menghapus
// semua session akun tersebut.
func ResetPassword(ctx context.Context, pool *pgxpool.Pool, login, newPassword string) error {
	repo := NewRepository(pool)
	if len(newPassword) < minPasswordLength {
		return errors.New("password minimal 8 karakter")
	}
	u, err := repo.ByLogin(ctx, login)
	if err != nil {
		return errors.New("akun tidak ditemukan")
	}
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	if err := repo.UpdatePasswordReset(ctx, u.ID, hash); err != nil {
		return err
	}
	return session.DestroyAllForUser(ctx, pool, u.ID)
}
