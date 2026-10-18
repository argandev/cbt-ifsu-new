package banksoal

import (
	"context"
	"strings"

	"cbt-smk-v2/backend/internal/platform"
)

// BankFilter adalah filter daftar bank. Nilai nil = filter tidak dipakai
// (dipetakan ke NULL pada parameter SQL, sama seperti pola any sebelumnya).
type BankFilter struct {
	Search       any
	MapelID      any
	KelasTingkat any
	Locked       any
}

// guruScopeOK adalah inti murni aturan scope guru: admin bebas; guru hanya
// boleh memutasi bank miliknya (guru_id = dirinya) atau bank tanpa pemilik;
// role lain lolos (izin sudah diperiksa requirePerm).
func guruScopeOK(role, userID string, b *bankRow) bool {
	if role == "admin" {
		return true
	}
	if role == "guru" {
		return b.GuruID == nil || *b.GuruID == "" || *b.GuruID == userID
	}
	return true
}

// checkGuruScope menolak (403) bila guru memutasi bank milik guru lain.
// (Versi service dari checkGuruScope di handler.go; dipakai endpoint CRUD.)
func (s *Service) checkGuruScope(role, userID string, b *bankRow) *platform.AppError {
	if !guruScopeOK(role, userID, b) {
		return platform.ErrForbidden("forbidden", "Bank soal ini milik guru lain.")
	}
	return nil
}

// ListBank mengembalikan halaman bank soal sesuai filter.
func (s *Service) ListBank(ctx context.Context, f BankFilter, limit, offset int) ([]bankRow, int, *platform.AppError) {
	items, err := s.repo.ListBank(ctx, f, limit, offset)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca bank soal.")
	}
	// Error COUNT diabaikan seperti kode asal (_ = ...) — total 0 bila gagal.
	total, _ := s.repo.CountBank(ctx, f)
	return items, total, nil
}

// CreateBank membuat bank soal baru.
// Guru yang membuat bank menjadi pemiliknya; admin boleh menetapkan guru_id.
func (s *Service) CreateBank(ctx context.Context, actor RequestActor, role string, in bankInput) (*bankRow, *platform.AppError) {
	if strings.TrimSpace(in.Kode) == "" || strings.TrimSpace(in.Nama) == "" {
		return nil, platform.ErrBadRequest("validation_error", "kode dan nama bank wajib diisi.")
	}
	if !isValidUUID(in.MapelID) {
		return nil, platform.ErrBadRequest("validation_error", "mapel_id wajib UUID mapel yang valid.")
	}
	if in.KelasTingkat != 10 && in.KelasTingkat != 11 && in.KelasTingkat != 12 {
		return nil, platform.ErrBadRequest("validation_error", "kelas_tingkat harus 10, 11, atau 12.")
	}
	var guruID any
	if role == "guru" {
		guruID = actor.UserID
	} else if in.GuruID != nil && *in.GuruID != "" {
		if !isValidUUID(*in.GuruID) {
			return nil, platform.ErrBadRequest("validation_error", "guru_id harus UUID yang valid.")
		}
		guruID = *in.GuruID
	}
	b, err := s.repo.InsertBank(ctx, in.Kode, in.Nama, in.MapelID, in.KelasTingkat, guruID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
			return nil, platform.ErrConflict("duplicate_kode", "kode bank soal sudah dipakai.")
		}
		if strings.Contains(err.Error(), "foreign key") || strings.Contains(err.Error(), "violates") {
			return nil, platform.ErrBadRequest("validation_error", "mapel_id / guru_id tidak ditemukan.")
		}
		return nil, platform.ErrInternal("internal_error", "Gagal membuat bank soal.")
	}
	s.catat(ctx, actor, "banksoal.create", "bank_soal", b.ID,
		map[string]any{"kode": b.Kode, "nama": b.Nama})
	return b, nil
}

// GetBank mengembalikan satu bank soal.
func (s *Service) GetBank(ctx context.Context, id string) (*bankRow, *platform.AppError) {
	b, found, err := s.repo.LoadBank(ctx, id)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca bank soal.")
	}
	if !found {
		return nil, platform.ErrNotFound("not_found", "Bank soal tidak ditemukan.")
	}
	return b, nil
}

// UpdateBank mengubah bank soal.
// guru_id hanya boleh diubah admin; guru tidak bisa memindahkan kepemilikan.
// LoadBankChecked memuat bank + cek scope guru (404/403/500).
// Dipanggil handler SEBELUM bind body agar urutan status (404 dulu,
// 400 kemudian) identik dengan kode asal.
func (s *Service) LoadBankChecked(ctx context.Context, role, userID, id string) (*bankRow, *platform.AppError) {
	b, found, err := s.repo.LoadBank(ctx, id)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca bank soal.")
	}
	if !found {
		return nil, platform.ErrNotFound("not_found", "Bank soal tidak ditemukan.")
	}
	if appErr := s.checkGuruScope(role, userID, b); appErr != nil {
		return nil, appErr
	}
	return b, nil
}

// ApplyBankUpdate memvalidasi input + menyimpan perubahan bank.
// Dipanggil handler SETELAH bind body (agar 400 bind mendahului validasi,
// sesuai urutan kode asal).
func (s *Service) ApplyBankUpdate(ctx context.Context, actor RequestActor, role string, b *bankRow, in bankInput) (*bankRow, *platform.AppError) {
	if strings.TrimSpace(in.Kode) == "" || strings.TrimSpace(in.Nama) == "" {
		return nil, platform.ErrBadRequest("validation_error", "kode dan nama bank wajib diisi.")
	}
	if !isValidUUID(in.MapelID) {
		return nil, platform.ErrBadRequest("validation_error", "mapel_id wajib UUID mapel yang valid.")
	}
	if in.KelasTingkat != 10 && in.KelasTingkat != 11 && in.KelasTingkat != 12 {
		return nil, platform.ErrBadRequest("validation_error", "kelas_tingkat harus 10, 11, atau 12.")
	}
	guruID := b.GuruID
	if role == "admin" && in.GuruID != nil {
		if *in.GuruID == "" {
			guruID = nil
		} else if isValidUUID(*in.GuruID) {
			guruID = in.GuruID
		} else {
			return nil, platform.ErrBadRequest("validation_error", "guru_id harus UUID yang valid.")
		}
	}
	if err := s.repo.UpdateBank(ctx, b.ID, in.Kode, in.Nama, in.MapelID, in.KelasTingkat, guruID); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
			return nil, platform.ErrConflict("duplicate_kode", "kode bank soal sudah dipakai.")
		}
		return nil, platform.ErrInternal("internal_error", "Gagal mengubah bank soal.")
	}
	s.catat(ctx, actor, "banksoal.update", "bank_soal", b.ID,
		map[string]any{"kode": in.Kode, "nama": in.Nama})
	updated, _, _ := s.repo.LoadBank(ctx, b.ID)
	return updated, nil
}

// DeleteBank menghapus bank soal. Ditolak 409 bila terkunci atau dipakai paket.
func (s *Service) DeleteBank(ctx context.Context, actor RequestActor, role, id string) (map[string]any, *platform.AppError) {
	b, found, err := s.repo.LoadBank(ctx, id)
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
		return nil, platform.ErrConflict("bank_locked", "Bank soal terkunci — tidak bisa dihapus. Buka kunci dulu (oleh admin / bukan pembuat).")
	}
	inUse, err := s.repo.BankDipakaiPaket(ctx, b.ID)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal memeriksa pemakaian bank.")
	}
	if inUse {
		return nil, platform.ErrConflict("bank_in_use", "Bank soal dipakai paket ujian — tidak bisa dihapus.")
	}
	if err := s.repo.DeleteBank(ctx, b.ID); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menghapus bank soal.")
	}
	s.catat(ctx, actor, "banksoal.delete", "bank_soal", b.ID,
		map[string]any{"kode": b.Kode, "nama": b.Nama})
	return map[string]any{"deleted": true, "id": b.ID}, nil
}

// SetBankLock mengunci/membuka bank soal. SoD-1 (ANALISIS Bab 11.4):
// pembuat bank TIDAK BOLEH mengunci/membuka bank buatannya sendiri.
func (s *Service) SetBankLock(ctx context.Context, actor RequestActor, role, id string, locked bool) (*bankRow, *platform.AppError) {
	b, found, err := s.repo.LoadBank(ctx, id)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca bank soal.")
	}
	if !found {
		return nil, platform.ErrNotFound("not_found", "Bank soal tidak ditemukan.")
	}
	if role == "guru" && b.GuruID != nil && *b.GuruID == actor.UserID {
		return nil, platform.ErrForbidden("sod_violation",
			"Pemisahan tugas (SoD-1): pembuat bank tidak boleh mengunci/membuka bank buatannya sendiri. Minta guru lain atau admin.")
	}
	if b.IsLocked == locked {
		return b, nil
	}
	if err := s.repo.SetBankLock(ctx, b.ID, locked); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal mengubah kunci bank.")
	}
	aksi := "banksoal.lock"
	if !locked {
		aksi = "banksoal.unlock"
	}
	s.catat(ctx, actor, aksi, "bank_soal", b.ID,
		map[string]any{"kode": b.Kode, "is_locked": locked})
	updated, _, _ := s.repo.LoadBank(ctx, b.ID)
	return updated, nil
}
