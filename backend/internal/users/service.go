package users

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"cbt-smk-v2/backend/internal/auth"
	"cbt-smk-v2/backend/internal/platform"
	"cbt-smk-v2/backend/internal/session"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Panjang password minimal — sama seperti kebijakan auth (auth.go).
const minPasswordLength = 8

// --- validasi (fungsi murni, diuji tanpa DB) -------------------------------
// (Dipindah dari users.go tanpa perubahan.)

// StaffRoles adalah 5 role staf yang boleh dikelola lewat endpoint ini.
// Role 'siswa' sengaja TIDAK ada: akun siswa dikelola lewat Siswa.
func StaffRoles() []string {
	return []string{"admin", "proktor", "guru", "pengawas", "penguji_ukk"}
}

// IsStaffRole melaporkan apakah role termasuk role staf.
func IsStaffRole(role string) bool {
	for _, r := range StaffRoles() {
		if r == role {
			return true
		}
	}
	return false
}

// ValidasiRoleStaf mengembalikan pesan error bila role bukan role staf
// ("" = valid). Role 'siswa' ditolak dengan pesan yang menjelaskan
// bahwa akun siswa dikelola di halaman Siswa.
func ValidasiRoleStaf(role string) string {
	if strings.TrimSpace(role) == "" {
		return "Role wajib diisi."
	}
	if role == "siswa" {
		return "Role siswa tidak dikelola di sini; kelola akun siswa lewat halaman Siswa."
	}
	if !IsStaffRole(role) {
		return "Role harus salah satu dari: admin, proktor, guru, pengawas, penguji_ukk."
	}
	return ""
}

// ValidasiUsername mengembalikan pesan error bila username tidak valid.
func ValidasiUsername(username string) string {
	username = strings.TrimSpace(username)
	if username == "" {
		return "Username wajib diisi."
	}
	if strings.ContainsAny(username, " \t\r\n") {
		return "Username tidak boleh mengandung spasi."
	}
	return ""
}

// ValidasiEmail mengembalikan pesan error bila email tidak valid.
// String kosong = valid (email opsional; disimpan NULL).
func ValidasiEmail(email string) string {
	email = strings.TrimSpace(email)
	if email == "" {
		return ""
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return "Format email tidak valid."
	}
	// mail.ParseAddress menerima "a@b" (tanpa titik di domain) — tolak,
	// karena alamat seperti itu bukan email yang bisa dipakai sekolah.
	domain := addr.Address[strings.LastIndex(addr.Address, "@")+1:]
	if !strings.Contains(domain, ".") {
		return "Format email tidak valid."
	}
	return ""
}

// ValidasiPassword mengembalikan pesan error bila password terlalu pendek.
func ValidasiPassword(password string) string {
	if len(password) < minPasswordLength {
		return "Password minimal 8 karakter."
	}
	return ""
}

// validUUID melaporkan apakah s adalah UUID valid.
func validUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// --- Service ----------------------------------------------------------------


// CreateInput adalah input pembuatan akun staf.
type CreateInput struct {
	Username string
	Nama     string
	Email    string
	Role     string
	Password string
}

// UpdateInput adalah input pengubahan akun staf (nil = tidak diubah).
type UpdateInput struct {
	Nama  *string
	Email *string
	Role  *string
}

// Service adalah lapisan logika bisnis paket users.
// repository adalah kontrak penyimpanan yang dipakai Service.
// (DIP: Service bergantung pada abstraksi, bukan *Repository konkrit.)
type repository interface {
	CountActiveAdminsExcept(ctx context.Context, excludeID string) (int64, error)
	CountStaf(ctx context.Context, f StafFilter) (int64, error)
	CreateStaf(ctx context.Context, username, nama string, emailArg any, role, hash string) (*penggunaDTO, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	EmailExistsExcept(ctx context.Context, email, excludeID string) (bool, error)
	GetStaf(ctx context.Context, id string) (*penggunaDTO, error)
	ListStaf(ctx context.Context, f StafFilter, limit, offset int) ([]penggunaDTO, error)
	SetAktif(ctx context.Context, id string, isActive bool) (*penggunaDTO, error)
	UpdatePassword(ctx context.Context, id, hash string) error
	UpdateStaf(ctx context.Context, id, nama string, emailArg any, role string) (*penggunaDTO, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
}

type Service struct {
	repo repository
	pool *pgxpool.Pool
}

// NewService membuat Service.
func NewService(repo repository, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, pool: pool}
}


// List mengembalikan halaman akun staf sesuai filter.
func (s *Service) List(ctx context.Context, f StafFilter, perPage, offset int) ([]penggunaDTO, int64, *platform.AppError) {
	if f.Role != nil {
		// Kontrak: filter role 'siswa' -> 400 (akun siswa di halaman Siswa).
		if *f.Role == "siswa" {
			return nil, 0, platform.ErrBadRequest("validation_error", "Role siswa tidak dikelola di sini; kelola akun siswa lewat halaman Siswa.")
		}
		if msg := ValidasiRoleStaf(*f.Role); msg != "" {
			return nil, 0, platform.NewAppError("validation_error", msg, 422)
		}
	}
	total, err := s.repo.CountStaf(ctx, f)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca pengguna.")
	}
	items, err := s.repo.ListStaf(ctx, f, perPage, offset)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca pengguna.")
	}
	return items, total, nil
}

// Create membuat akun staf baru.
func (s *Service) Create(ctx context.Context, actor platform.RequestActor, in CreateInput) (*penggunaDTO, *platform.AppError) {
	if msg := ValidasiUsername(in.Username); msg != "" {
		return nil, platform.NewAppError("validation_error", msg, 422)
	}
	if strings.TrimSpace(in.Nama) == "" {
		return nil, platform.NewAppError("validation_error", "Nama wajib diisi.", 422)
	}
	if msg := ValidasiEmail(in.Email); msg != "" {
		return nil, platform.NewAppError("validation_error", msg, 422)
	}
	// Kontrak: role 'siswa' -> 400 (akun siswa di halaman Siswa).
	if strings.TrimSpace(in.Role) == "siswa" {
		return nil, platform.ErrBadRequest("validation_error", "Role siswa tidak dikelola di sini; kelola akun siswa lewat halaman Siswa.")
	}
	if msg := ValidasiRoleStaf(in.Role); msg != "" {
		return nil, platform.NewAppError("validation_error", msg, 422)
	}
	if msg := ValidasiPassword(in.Password); msg != "" {
		return nil, platform.NewAppError("validation_error", msg, 422)
	}
	username := strings.TrimSpace(in.Username)
	email := strings.TrimSpace(in.Email)

	// Duplikat case-insensitive (UNIQUE DB case-sensitive, sedangkan
	// login case-insensitive — "Admin" vs "admin" tidak boleh ganda).
	dupeUsername, err := s.repo.UsernameExists(ctx, username)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal memeriksa username.")
	}
	if dupeUsername {
		return nil, platform.ErrConflict("conflict", "Username sudah dipakai.")
	}
	if email != "" {
		dupeEmail, err := s.repo.EmailExists(ctx, email)
		if err != nil {
			return nil, platform.ErrInternal("internal_error", "Gagal memeriksa email.")
		}
		if dupeEmail {
			return nil, platform.ErrConflict("conflict", "Email sudah dipakai.")
		}
	}

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal meng-hash password.")
	}
	var emailArg any
	if email != "" {
		emailArg = email
	}
	d, err := s.repo.CreateStaf(ctx, username, strings.TrimSpace(in.Nama), emailArg, in.Role, hash)
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			switch constraint {
			case "users_username_key":
				return nil, platform.ErrConflict("conflict", "Username sudah dipakai.")
			case "users_email_key":
				return nil, platform.ErrConflict("conflict", "Email sudah dipakai.")
			}
			return nil, platform.ErrConflict("conflict", "Data duplikat: nilai unik sudah dipakai.")
		}
		return nil, platform.ErrInternal("internal_error", "Gagal menambah pengguna.")
	}
	platform.Catat(ctx, s.pool, actor, "pengguna.create", "users", d.ID,
		map[string]any{"username": username, "nama": d.Nama, "email": d.Email, "role": d.Role})
	return d, nil
}

// Get mengembalikan detail satu akun staf.
func (s *Service) Get(ctx context.Context, id string) (*penggunaDTO, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Pengguna tidak ditemukan.")
	}
	d, err := s.repo.GetStaf(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, platform.ErrNotFound("not_found", "Pengguna tidak ditemukan.")
	}
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca pengguna.")
	}
	return d, nil
}

// Update mengubah nama/email/role akun staf. Username TIDAK bisa diubah.
func (s *Service) Update(ctx context.Context, actor platform.RequestActor, id string, in UpdateInput) (*penggunaDTO, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Pengguna tidak ditemukan.")
	}
	cur, err := s.repo.GetStaf(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, platform.ErrNotFound("not_found", "Pengguna tidak ditemukan.")
	}
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca pengguna.")
	}

	nama := cur.Nama
	if in.Nama != nil {
		nama = strings.TrimSpace(*in.Nama)
		if nama == "" {
			return nil, platform.NewAppError("validation_error", "Nama wajib diisi.", 422)
		}
	}
	var emailArg any = cur.Email
	if in.Email != nil {
		email := strings.TrimSpace(*in.Email)
		if msg := ValidasiEmail(email); msg != "" {
			return nil, platform.NewAppError("validation_error", msg, 422)
		}
		if email == "" {
			emailArg = nil
		} else {
			dupe, err := s.repo.EmailExistsExcept(ctx, email, id)
			if err != nil {
				return nil, platform.ErrInternal("internal_error", "Gagal memeriksa email.")
			}
			if dupe {
				return nil, platform.ErrConflict("conflict", "Email sudah dipakai.")
			}
			emailArg = email
		}
	}
	role := cur.Role
	if in.Role != nil {
		// Kontrak: role 'siswa' -> 400 (akun siswa di halaman Siswa).
		if strings.TrimSpace(*in.Role) == "siswa" {
			return nil, platform.ErrBadRequest("validation_error", "Role siswa tidak dikelola di sini; kelola akun siswa lewat halaman Siswa.")
		}
		if msg := ValidasiRoleStaf(*in.Role); msg != "" {
			return nil, platform.NewAppError("validation_error", msg, 422)
		}
		role = *in.Role
	}
	// Guard admin terakhir: demosi admin aktif -> non-admin ditolak
	// bila tidak ada admin aktif lain yang tersisa.
	if cur.Role == "admin" && role != "admin" && cur.IsActive {
		n, err := s.repo.CountActiveAdminsExcept(ctx, id)
		if err != nil {
			return nil, platform.ErrInternal("internal_error", "Gagal memeriksa admin aktif.")
		}
		if n == 0 {
			return nil, platform.ErrConflict("admin_terakhir", "Tidak dapat mengubah role admin terakhir yang masih aktif. Buat admin lain dulu.")
		}
	}

	d, err := s.repo.UpdateStaf(ctx, id, nama, emailArg, role)
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok && constraint == "users_email_key" {
			return nil, platform.ErrConflict("conflict", "Email sudah dipakai.")
		}
		return nil, platform.ErrInternal("internal_error", "Gagal mengubah pengguna.")
	}
	platform.Catat(ctx, s.pool, actor, "pengguna.update", "users", d.ID,
		map[string]any{"nama": d.Nama, "email": d.Email, "role": d.Role})
	return d, nil
}

// ResetPassword mereset password akun staf + menghapus semua sesinya.
func (s *Service) ResetPassword(ctx context.Context, actor platform.RequestActor, id, passwordBaru string) (map[string]any, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Pengguna tidak ditemukan.")
	}
	if msg := ValidasiPassword(passwordBaru); msg != "" {
		return nil, platform.NewAppError("validation_error", msg, 422)
	}
	cur, err := s.repo.GetStaf(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, platform.ErrNotFound("not_found", "Pengguna tidak ditemukan.")
	}
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca pengguna.")
	}
	hash, err := auth.HashPassword(passwordBaru)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal meng-hash password.")
	}
	if err := s.repo.UpdatePassword(ctx, id, hash); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal mereset password.")
	}
	targetID, err := uuid.Parse(id)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal mereset password.")
	}
	if err := session.DestroyAllForUser(ctx, s.pool, targetID); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menghapus sesi pengguna.")
	}
	platform.Catat(ctx, s.pool, actor, "pengguna.reset-password", "users", id,
		map[string]any{"username": cur.Username})
	return map[string]any{"password_reset": true, "must_change_password": true, "id": id}, nil
}

// ToggleAktif mengaktifkan/menonaktifkan akun staf.
// DILARANG: menonaktifkan akun sendiri (403) dan menonaktifkan admin
// terakhir yang masih aktif (409). TIDAK ADA hard delete.
func (s *Service) ToggleAktif(ctx context.Context, actor platform.RequestActor, id string, isActive *bool) (*penggunaDTO, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Pengguna tidak ditemukan.")
	}
	if isActive == nil {
		return nil, platform.NewAppError("validation_error", "Field is_active wajib diisi (true/false).", 422)
	}
	cur, err := s.repo.GetStaf(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, platform.ErrNotFound("not_found", "Pengguna tidak ditemukan.")
	}
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca pengguna.")
	}
	// Guard: nonaktifkan akun sendiri — dicek dulu agar pesannya
	// jelas walau actor juga admin terakhir.
	if !*isActive && id == actor.UserID {
		return nil, platform.ErrForbidden("forbidden", "Tidak dapat menonaktifkan akun sendiri.")
	}
	// Guard admin terakhir: hanya untuk transisi aktif -> nonaktif.
	if !*isActive && cur.Role == "admin" && cur.IsActive {
		n, err := s.repo.CountActiveAdminsExcept(ctx, id)
		if err != nil {
			return nil, platform.ErrInternal("internal_error", "Gagal memeriksa admin aktif.")
		}
		if n == 0 {
			return nil, platform.ErrConflict("admin_terakhir", "Tidak dapat menonaktifkan admin terakhir yang masih aktif. Buat admin lain dulu.")
		}
	}
	d, err := s.repo.SetAktif(ctx, id, *isActive)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal mengubah status pengguna.")
	}
	platform.Catat(ctx, s.pool, actor, "pengguna.toggle-aktif", "users", d.ID,
		map[string]any{"username": d.Username, "is_active": d.IsActive})
	return d, nil
}
