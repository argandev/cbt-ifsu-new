package master

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5"
)

// ---------------------------------------------------------------------
// Siswa
// ---------------------------------------------------------------------

type siswaInput struct {
	UserID     *string `json:"user_id"`
	NISN       string  `json:"nisn"`
	NIS        *string `json:"nis"`
	Nama       string  `json:"nama"`
	KelasID    string  `json:"kelas_id"`
	JurusanID  string  `json:"jurusan_id"`
	TahunMasuk *int32  `json:"tahun_masuk"`
	IsActive   *bool   `json:"is_active"`
}

func (in siswaInput) validate() string {
	if strings.TrimSpace(in.NISN) == "" {
		return "NISN wajib diisi."
	}
	if strings.TrimSpace(in.Nama) == "" {
		return "Nama siswa wajib diisi."
	}
	if !validUUID(in.KelasID) {
		return "Kelas wajib dipilih."
	}
	if !validUUID(in.JurusanID) {
		return "Jurusan wajib dipilih."
	}
	if in.UserID != nil && strings.TrimSpace(*in.UserID) != "" && !validUUID(*in.UserID) {
		return "user_id tidak valid."
	}
	if in.TahunMasuk != nil && (*in.TahunMasuk < 1900 || *in.TahunMasuk > 2100) {
		return "Tahun masuk tidak valid."
	}
	return ""
}

// siswaNISArg mengubah NIS kosong menjadi nil.
func siswaNISArg(in siswaInput) *string {
	if in.NIS == nil {
		return nil
	}
	if strings.TrimSpace(*in.NIS) == "" {
		return nil
	}
	v := strings.TrimSpace(*in.NIS)
	return &v
}

// createStudentAccountTx membuat akun login siswa di dalam transaksi:
// username = NISN, password awal = NISN (Argon2id), role 'siswa' di kolom
// users.role (SATU role per user — tidak ada tabel user_roles),
// must_change_password = TRUE. Bila username sudah ada (akun yatim),
// akun lama dipakai kembali. Mengembalikan user_id.
func (s *Service) createStudentAccountTx(ctx context.Context, tx pgx.Tx, nisn, nama string) (string, error) {
	hash, err := hashPassword(nisn)
	if err != nil {
		return "", err
	}
	var userID string
	err = tx.QueryRow(ctx,
		`INSERT INTO users (nama, username, password_hash, role, must_change_password, is_active)
		 VALUES ($1, $2, $3, 'siswa', TRUE, TRUE)
		 ON CONFLICT (username) DO NOTHING
		 RETURNING id`, nama, nisn, hash).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		// Username sudah ada: pakai akun lama + paksa ganti password.
		if err := tx.QueryRow(ctx,
			`SELECT id FROM users WHERE username = $1`, nisn).Scan(&userID); err != nil {
			return "", err
		}
		_, _ = tx.Exec(ctx,
			`UPDATE users SET must_change_password = TRUE, role = 'siswa' WHERE id = $1`, userID)
	} else if err != nil {
		return "", err
	}
	return userID, nil
}

// ListSiswa mengembalikan halaman siswa sesuai filter.
func (s *Service) ListSiswa(ctx context.Context, kelasID, jurusanID *string, isActive *bool, q *string, limit, offset int) ([]siswaDTO, int64, *platform.AppError) {
	total, err := s.repo.CountSiswa(ctx, kelasID, jurusanID, isActive, q)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca siswa.")
	}
	items, err := s.repo.ListSiswa(ctx, kelasID, jurusanID, isActive, q, limit, offset)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca siswa.")
	}
	return items, total, nil
}

// GetSiswa mengembalikan satu siswa.
func (s *Service) GetSiswa(ctx context.Context, id string) (*siswaDTO, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Siswa tidak ditemukan.")
	}
	d, err := s.repo.GetSiswa(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, platform.ErrNotFound("not_found", "Siswa tidak ditemukan.")
	}
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca siswa.")
	}
	return d, nil
}

// CreateSiswa menambah siswa baru. Tanpa user_id: buatkan akun login
// otomatis (sama seperti import): username = NISN, password awal = NISN.
func (s *Service) CreateSiswa(ctx context.Context, actor RequestActor, in siswaInput) (*siswaDTO, *platform.AppError) {
	if msg := in.validate(); msg != "" {
		return nil, platform.NewAppError("validation_error", msg, http.StatusUnprocessableEntity)
	}
	isActive := true
	if in.IsActive != nil {
		isActive = *in.IsActive
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menambah siswa.")
	}
	defer tx.Rollback(ctx)
	userID := ""
	if in.UserID != nil && strings.TrimSpace(*in.UserID) != "" {
		userID = *in.UserID
	} else {
		// Tanpa user_id: buatkan akun login otomatis (sama seperti import).
		userID, err = s.createStudentAccountTx(ctx, tx, strings.TrimSpace(in.NISN), strings.TrimSpace(in.Nama))
		if err != nil {
			return nil, platform.ErrInternal("internal_error", "Gagal membuat akun siswa.")
		}
	}
	var userIDArg any
	if userID != "" {
		userIDArg = userID
	}
	d, err := s.repo.InsertSiswaTx(ctx, tx, userIDArg, strings.TrimSpace(in.NISN), siswaNISArg(in),
		strings.TrimSpace(in.Nama), in.KelasID, in.JurusanID, in.TahunMasuk, isActive)
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			return nil, platform.ErrConflict("conflict", conflictMessage(constraint))
		}
		if isFKViolation(err) {
			return nil, platform.NewAppError("validation_error", "Kelas/jurusan/akun pengguna tidak ditemukan.", http.StatusUnprocessableEntity)
		}
		return nil, platform.ErrInternal("internal_error", "Gagal menambah siswa.")
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menambah siswa.")
	}
	s.catat(ctx, actor, "create_siswa", "siswa", d.ID, map[string]any{"nisn": d.NISN, "nama": d.Nama})
	return d, nil
}

// UpdateSiswa mengubah siswa. user_id tidak dikirim = pertahankan yang lama.
func (s *Service) UpdateSiswa(ctx context.Context, actor RequestActor, id string, in siswaInput) (*siswaDTO, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Siswa tidak ditemukan.")
	}
	if msg := in.validate(); msg != "" {
		return nil, platform.NewAppError("validation_error", msg, http.StatusUnprocessableEntity)
	}
	// user_id tidak dikirim = pertahankan yang lama.
	userIDArg := any(nil)
	if in.UserID != nil {
		if strings.TrimSpace(*in.UserID) == "" {
			userIDArg = nil
		} else {
			userIDArg = *in.UserID
		}
	} else {
		if existing := s.repo.GetSiswaUserID(ctx, id); existing != nil {
			userIDArg = *existing
		}
	}
	isActive := true
	if in.IsActive != nil {
		isActive = *in.IsActive
	}
	d, err := s.repo.UpdateSiswa(ctx, id, userIDArg, strings.TrimSpace(in.NISN), siswaNISArg(in),
		strings.TrimSpace(in.Nama), in.KelasID, in.JurusanID, in.TahunMasuk, isActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, platform.ErrNotFound("not_found", "Siswa tidak ditemukan.")
	}
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			return nil, platform.ErrConflict("conflict", conflictMessage(constraint))
		}
		if isFKViolation(err) {
			return nil, platform.NewAppError("validation_error", "Kelas/jurusan/akun pengguna tidak ditemukan.", http.StatusUnprocessableEntity)
		}
		return nil, platform.ErrInternal("internal_error", "Gagal mengubah siswa.")
	}
	s.catat(ctx, actor, "update_siswa", "siswa", d.ID, map[string]any{"nisn": d.NISN, "nama": d.Nama})
	return d, nil
}

// DeleteSiswa menghapus siswa. Siswa yang sudah punya attempt / nilai akhir
// / remedial hanya dinonaktifkan (ANALISIS 16.2), tidak dihapus.
// Akun login (users) SENGAJA tidak dihapus: riwayat & username tidak
// dipakai ulang untuk orang lain.
func (s *Service) DeleteSiswa(ctx context.Context, actor RequestActor, id string) (map[string]any, *platform.AppError) {
	if !validUUID(id) {
		return nil, platform.ErrNotFound("not_found", "Siswa tidak ditemukan.")
	}
	// Guard: siswa yang sudah punya attempt / nilai akhir / remedial
	// hanya dinonaktifkan (ANALISIS 16.2), tidak dihapus.
	// attempt.siswa_id menyimpan auth_siswa_id (= users.id akun siswa),
	// sedangkan nilai_akhir/remedial.siswa_id merujuk siswa.id.
	if userID := s.repo.GetSiswaUserID(ctx, id); userID != nil && *userID != "" {
		if n := s.repo.countOrZero(ctx, `SELECT COUNT(*) FROM attempt WHERE siswa_id = $1`, *userID); n > 0 {
			return nil, platform.ErrConflict("conflict",
				fmt.Sprintf("Siswa sudah punya %d attempt ujian. Nonaktifkan saja, jangan hapus.", n))
		} else if n < 0 {
			return nil, platform.ErrInternal("internal_error", "Gagal memeriksa pemakaian siswa.")
		}
	}
	if n := s.repo.countOrZero(ctx, `SELECT COUNT(*) FROM nilai_akhir WHERE siswa_id = $1`, id); n > 0 {
		return nil, platform.ErrConflict("conflict",
			fmt.Sprintf("Siswa sudah punya %d nilai akhir. Nonaktifkan saja, jangan hapus.", n))
	} else if n < 0 {
		return nil, platform.ErrInternal("internal_error", "Gagal memeriksa pemakaian siswa.")
	}
	if n := s.repo.countOrZero(ctx, `SELECT COUNT(*) FROM remedial WHERE siswa_id = $1`, id); n > 0 {
		return nil, platform.ErrConflict("conflict",
			fmt.Sprintf("Siswa sudah punya %d nilai remedial. Nonaktifkan saja, jangan hapus.", n))
	} else if n < 0 {
		return nil, platform.ErrInternal("internal_error", "Gagal memeriksa pemakaian siswa.")
	}
	affected, err := s.repo.DeleteSiswa(ctx, id)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menghapus siswa.")
	}
	if affected == 0 {
		return nil, platform.ErrNotFound("not_found", "Siswa tidak ditemukan.")
	}
	s.catat(ctx, actor, "delete_siswa", "siswa", id, nil)
	return map[string]any{"deleted": true, "id": id}, nil
}
