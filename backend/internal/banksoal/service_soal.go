package banksoal

import (
	"context"

	"cbt-smk-v2/backend/internal/platform"
)

// SoalFilter adalah filter daftar soal. Nilai nil = filter tidak dipakai
// (dipetakan ke NULL pada parameter SQL, sama seperti pola any sebelumnya).
type SoalFilter struct {
	Tipe   any
	Level  any
	Kisi   any
	Search any
}

// ListSoal mengembalikan halaman soal dalam satu bank.
func (s *Service) ListSoal(ctx context.Context, bankID string, f SoalFilter, limit, offset int) ([]soalRow, int, *platform.AppError) {
	b, found, err := s.repo.LoadBank(ctx, bankID)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca bank soal.")
	}
	if !found {
		return nil, 0, platform.ErrNotFound("not_found", "Bank soal tidak ditemukan.")
	}
	items, err := s.repo.ListSoal(ctx, b.ID, f, limit, offset)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca soal.")
	}
	// Error COUNT diabaikan seperti kode asal (_ = ...) — total 0 bila gagal.
	total, _ := s.repo.CountSoal(ctx, b.ID, f)
	return items, total, nil
}

// CreateSoal menambah soal ke bank (bank tidak boleh terkunci).
func (s *Service) CreateSoal(ctx context.Context, actor platform.RequestActor, role, bankID string, in soalInput) (*soalRow, *platform.AppError) {
	b, found, err := s.repo.LoadBank(ctx, bankID)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca bank soal.")
	}
	if !found {
		return nil, platform.ErrNotFound("not_found", "Bank soal tidak ditemukan.")
	}
	if appErr := s.checkGuruScope(role, actor.UserID, b); appErr != nil {
		return nil, appErr
	}
	if b.IsLocked {
		return nil, platform.ErrConflict("bank_locked", "Bank soal terkunci — soal tidak bisa ditambah.")
	}
	if err := in.validate(); err != nil {
		return nil, platform.ErrBadRequest("validation_error", err.Error())
	}
	level := NormalLevel(in.Level)
	if level == "" {
		level = "sedang"
	}
	status := in.Status
	if status == "" {
		status = "aktif"
	}
	sr, err := s.repo.InsertSoal(ctx, b.ID, in, level, status)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membuat soal.")
	}
	platform.Catat(ctx, s.pool, actor, "banksoal.create", "soal", sr.ID,
		map[string]any{"bank_id": b.ID, "tipe": sr.Tipe})
	return sr, nil
}

// GetSoal mengembalikan satu soal.
func (s *Service) GetSoal(ctx context.Context, bankID, soalID string) (*soalRow, *platform.AppError) {
	sr, found, err := s.repo.LoadSoal(ctx, bankID, soalID)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca soal.")
	}
	if !found {
		return nil, platform.ErrNotFound("not_found", "Soal tidak ditemukan.")
	}
	return sr, nil
}

// UpdateSoal mengubah soal (bank tidak boleh terkunci).
// LoadSoalChecked memuat bank + soal + cek scope guru + cek kunci bank
// (404/403/409/500). Dipanggil handler SEBELUM bind body agar urutan
// status identik dengan kode asal.
func (s *Service) LoadSoalChecked(ctx context.Context, role, userID, bankID, soalID string) (*bankRow, *soalRow, *platform.AppError) {
	b, found, err := s.repo.LoadBank(ctx, bankID)
	if err != nil {
		return nil, nil, platform.ErrInternal("internal_error", "Gagal membaca bank soal.")
	}
	if !found {
		return nil, nil, platform.ErrNotFound("not_found", "Bank soal tidak ditemukan.")
	}
	if appErr := s.checkGuruScope(role, userID, b); appErr != nil {
		return nil, nil, appErr
	}
	if b.IsLocked {
		return nil, nil, platform.ErrConflict("bank_locked", "Bank soal terkunci — soal tidak bisa diubah.")
	}
	sr, found, err := s.repo.LoadSoal(ctx, b.ID, soalID)
	if err != nil {
		return nil, nil, platform.ErrInternal("internal_error", "Gagal membaca soal.")
	}
	if !found {
		return nil, nil, platform.ErrNotFound("not_found", "Soal tidak ditemukan.")
	}
	return b, sr, nil
}

// ApplySoalUpdate memvalidasi input + menyimpan perubahan soal.
// Dipanggil handler SETELAH bind body (agar 400 bind mendahului validasi,
// sesuai urutan kode asal).
func (s *Service) ApplySoalUpdate(ctx context.Context, actor platform.RequestActor, b *bankRow, sr *soalRow, in soalInput) (*soalRow, *platform.AppError) {
	if err := in.validate(); err != nil {
		return nil, platform.ErrBadRequest("validation_error", err.Error())
	}
	level := NormalLevel(in.Level)
	if level == "" {
		level = "sedang"
	}
	status := in.Status
	if status == "" {
		status = sr.Status
	}
	if err := s.repo.UpdateSoal(ctx, sr.ID, b.ID, in, level, status); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal mengubah soal.")
	}
	platform.Catat(ctx, s.pool, actor, "banksoal.update", "soal", sr.ID,
		map[string]any{"bank_id": b.ID, "tipe": in.Tipe})
	updated, _, _ := s.repo.LoadSoal(ctx, b.ID, sr.ID)
	return updated, nil
}

// DeleteSoal menghapus soal (bank tidak boleh terkunci).
func (s *Service) DeleteSoal(ctx context.Context, actor platform.RequestActor, role, bankID, soalID string) (map[string]any, *platform.AppError) {
	b, found, err := s.repo.LoadBank(ctx, bankID)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca bank soal.")
	}
	if !found {
		return nil, platform.ErrNotFound("not_found", "Bank soal tidak ditemukan.")
	}
	if appErr := s.checkGuruScope(role, actor.UserID, b); appErr != nil {
		return nil, appErr
	}
	if b.IsLocked {
		return nil, platform.ErrConflict("bank_locked", "Bank soal terkunci — soal tidak bisa dihapus.")
	}
	sr, found, err := s.repo.LoadSoal(ctx, b.ID, soalID)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca soal.")
	}
	if !found {
		return nil, platform.ErrNotFound("not_found", "Soal tidak ditemukan.")
	}
	if err := s.repo.DeleteSoal(ctx, sr.ID, b.ID); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menghapus soal.")
	}
	platform.Catat(ctx, s.pool, actor, "banksoal.delete", "soal", sr.ID,
		map[string]any{"bank_id": b.ID, "tipe": sr.Tipe})
	return map[string]any{"deleted": true, "id": sr.ID}, nil
}

// ImportCSV mengimpor soal dari teks CSV ke bank (bank tidak boleh terkunci).
// Mode "all_or_nothing": ada baris gagal -> tidak ada yang disimpan
// (respons 200 membawa ringkasan). Mode "skip_errors": baris valid tetap
// disimpan, baris gagal dilaporkan.
func (s *Service) ImportCSV(ctx context.Context, actor platform.RequestActor, role, bankID, csvText, mode string) (map[string]any, *platform.AppError) {
	b, found, err := s.repo.LoadBank(ctx, bankID)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca bank soal.")
	}
	if !found {
		return nil, platform.ErrNotFound("not_found", "Bank soal tidak ditemukan.")
	}
	if appErr := s.checkGuruScope(role, actor.UserID, b); appErr != nil {
		return nil, appErr
	}
	if b.IsLocked {
		return nil, platform.ErrConflict("bank_locked", "Bank soal terkunci — import ditolak.")
	}
	parsed, importErrs := parseCSVImport(csvText)
	if len(parsed) == 0 && len(importErrs) == 0 {
		return nil, platform.ErrBadRequest("validation_error", "CSV tidak berisi baris soal.")
	}
	if mode == "all_or_nothing" && len(importErrs) > 0 {
		return map[string]any{
			"sukses": 0, "gagal": len(importErrs), "errors": importErrs,
			"mode": mode, "disimpan": false,
		}, nil
	}
	sukses := 0
	for _, row := range parsed {
		if err := s.repo.InsertSoalCSV(ctx, b.ID, row); err != nil {
			importErrs = append(importErrs, map[string]any{
				"baris": row.baris, "pesan": "Gagal menyimpan ke database.",
			})
			continue
		}
		sukses++
	}
	platform.Catat(ctx, s.pool, actor, "banksoal.import", "bank_soal", b.ID,
		map[string]any{"sukses": sukses, "gagal": len(importErrs), "mode": mode})
	return map[string]any{
		"sukses": sukses, "gagal": len(importErrs), "errors": importErrs,
		"mode": mode, "disimpan": sukses > 0,
	}, nil
}
