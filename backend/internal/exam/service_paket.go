package exam

import (
	"context"
	"strings"

	"cbt-smk-v2/backend/internal/platform"
)

// PaketInput adalah input buat/ubah paket (dipetakan dari body JSON).
type PaketInput struct {
	Kode        string
	Nama        string
	MapelID     string
	Jenis       string
	DurasiMenit int
	JumlahSoal  int
	AcakSoal    *bool
	AcakJawaban *bool
	TampilHasil *bool
}

func validJenis(jenis string) bool {
	switch jenis {
	case "STS", "SAS", "USBK", "UKK", "Tryout":
		return true
	}
	return false
}

// validatePaket memvalidasi input paket (dulu: bindPaket) + menerapkan
// nilai default.
func validatePaket(in PaketInput) (paket, *platform.AppError) {
	kode := strings.TrimSpace(in.Kode)
	nama := strings.TrimSpace(in.Nama)
	jenis := strings.TrimSpace(in.Jenis)
	if kode == "" || nama == "" || in.MapelID == "" {
		return paket{}, platform.ErrBadRequest("validation_error", "kode, nama, dan mapel_id wajib diisi.")
	}
	if !validJenis(jenis) {
		return paket{}, platform.ErrBadRequest("validation_error", "jenis harus STS/SAS/USBK/UKK/Tryout.")
	}
	if in.DurasiMenit <= 0 {
		return paket{}, platform.ErrBadRequest("validation_error", "durasi_menit harus > 0.")
	}
	if in.JumlahSoal < 0 {
		return paket{}, platform.ErrBadRequest("validation_error", "jumlah_soal tidak boleh negatif.")
	}
	p := paket{Kode: kode, Nama: nama, MapelID: in.MapelID, Jenis: jenis,
		DurasiMenit: in.DurasiMenit, JumlahSoal: in.JumlahSoal,
		AcakSoal: true, AcakJawaban: true, TampilHasil: false}
	if in.AcakSoal != nil {
		p.AcakSoal = *in.AcakSoal
	}
	if in.AcakJawaban != nil {
		p.AcakJawaban = *in.AcakJawaban
	}
	if in.TampilHasil != nil {
		p.TampilHasil = *in.TampilHasil
	}
	return p, nil
}

// ListPaket mengembalikan daftar paket ujian.
func (s *Service) ListPaket(ctx context.Context, actor platform.RequestActor, jenis, mapelID string) (any, *platform.AppError) {
	items, err := s.repo.ListPaket(ctx, jenis, mapelID)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca paket.")
	}
	return map[string]any{"items": items}, nil
}

// CreatePaket membuat paket ujian baru.
func (s *Service) CreatePaket(ctx context.Context, actor platform.RequestActor, in PaketInput) (*paket, *platform.AppError) {
	p, appErr := validatePaket(in)
	if appErr != nil {
		return nil, appErr
	}
	mapelAda, err := s.repo.MapelExists(ctx, p.MapelID)
	if err != nil || !mapelAda {
		return nil, platform.ErrBadRequest("validation_error", "mapel_id tidak ditemukan.")
	}
	id, err := s.repo.InsertPaket(ctx, p)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, platform.ErrConflict("conflict", "Kode paket sudah dipakai.")
		}
		return nil, platform.ErrInternal("internal_error", "Gagal membuat paket.")
	}
	p.ID = id
	platform.Catat(ctx, s.pool, actor, "paket.create", "paket_ujian", id,
		map[string]any{"kode": p.Kode, "nama": p.Nama})
	return &p, nil
}

// GetPaket mengembalikan detail satu paket.
func (s *Service) GetPaket(ctx context.Context, actor platform.RequestActor, id string) (*paket, *platform.AppError) {
	p, err := s.repo.GetPaketDetail(ctx, id)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Paket tidak ditemukan.")
	}
	return &p, nil
}

// UpdatePaket mengubah paket ujian.
func (s *Service) UpdatePaket(ctx context.Context, actor platform.RequestActor, id string, in PaketInput) (map[string]any, *platform.AppError) {
	p, appErr := validatePaket(in)
	if appErr != nil {
		return nil, appErr
	}
	lamaMapel, err := s.repo.PaketMapelID(ctx, id)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Paket tidak ditemukan.")
	}
	if lamaMapel != p.MapelID {
		jumlahSoalLain, _ := s.repo.CountSoalMapelLain(ctx, id, p.MapelID)
		if jumlahSoalLain > 0 {
			return nil, platform.NewAppError("validation_error",
				"Paket masih berisi soal dari mapel lain. Kosongkan susunan soal dulu sebelum ganti mapel.", 422)
		}
	}
	n, err := s.repo.UpdatePaket(ctx, id, p)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, platform.ErrConflict("conflict", "Kode paket sudah dipakai.")
		}
		return nil, platform.ErrInternal("internal_error", "Gagal mengubah paket.")
	}
	if n == 0 {
		return nil, platform.ErrNotFound("not_found", "Paket tidak ditemukan.")
	}
	platform.Catat(ctx, s.pool, actor, "paket.update", "paket_ujian", id,
		map[string]any{"kode": p.Kode, "nama": p.Nama})
	return map[string]any{"id": id, "updated": true}, nil
}

// DeletePaket menghapus paket (ditolak bila masih dipakai jadwal).
func (s *Service) DeletePaket(ctx context.Context, actor platform.RequestActor, id string) (map[string]any, *platform.AppError) {
	dipakai, _ := s.repo.CountJadwalPaket(ctx, id)
	if dipakai > 0 {
		return nil, platform.ErrConflict("conflict", "Paket masih dipakai jadwal ujian. Hapus jadwal dulu.")
	}
	n, err := s.repo.DeletePaket(ctx, id)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menghapus paket.")
	}
	if n == 0 {
		return nil, platform.ErrNotFound("not_found", "Paket tidak ditemukan.")
	}
	platform.Catat(ctx, s.pool, actor, "paket.delete", "paket_ujian", id, nil)
	return map[string]any{"id": id, "deleted": true}, nil
}

// ListPaketSoal mengembalikan susunan soal paket.
func (s *Service) ListPaketSoal(ctx context.Context, actor platform.RequestActor, paketID string) (any, *platform.AppError) {
	items, err := s.repo.ListPaketSoal(ctx, paketID)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca susunan soal.")
	}
	return map[string]any{"paket_id": paketID, "items": items}, nil
}

// AddPaketSoal menambah satu soal ke paket (soal harus se-mapel).
func (s *Service) AddPaketSoal(ctx context.Context, actor platform.RequestActor, paketID, soalID string) (map[string]any, *platform.AppError) {
	paketMapel, err := s.repo.PaketMapelID(ctx, paketID)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Paket tidak ditemukan.")
	}
	soalMapel, err := s.repo.SoalMapelID(ctx, soalID)
	if err != nil {
		return nil, platform.ErrBadRequest("validation_error", "soal_id tidak ditemukan.")
	}
	if soalMapel != paketMapel {
		return nil, platform.NewAppError("validation_error", "Soal bukan dari mapel paket ini.", 422)
	}
	maxUrutan, _ := s.repo.MaxUrutanPaketSoal(ctx, paketID)
	if err := s.repo.InsertPaketSoal(ctx, paketID, soalID, maxUrutan+1); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menambah soal ke paket.")
	}
	_ = s.repo.SinkronJumlahSoal(ctx, paketID)
	platform.Catat(ctx, s.pool, actor, "paket.soal.add", "paket_ujian", paketID,
		map[string]any{"soal_id": soalID})
	return map[string]any{"paket_id": paketID, "soal_id": soalID, "urutan": maxUrutan + 1}, nil
}

// DeletePaketSoal menghapus soal dari paket.
func (s *Service) DeletePaketSoal(ctx context.Context, actor platform.RequestActor, paketID, soalID string) (map[string]any, *platform.AppError) {
	n, err := s.repo.DeletePaketSoal(ctx, paketID, soalID)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menghapus soal dari paket.")
	}
	if n == 0 {
		return nil, platform.ErrNotFound("not_found", "Soal tidak ada di paket ini.")
	}
	_ = s.repo.SinkronJumlahSoal(ctx, paketID)
	platform.Catat(ctx, s.pool, actor, "paket.soal.delete", "paket_ujian", paketID,
		map[string]any{"soal_id": soalID})
	return map[string]any{"paket_id": paketID, "soal_id": soalID, "deleted": true}, nil
}

// GenerateAcak menambah soal acak dari bank mapel ke paket.
func (s *Service) GenerateAcak(ctx context.Context, actor platform.RequestActor, paketID string, jumlah int) (map[string]any, *platform.AppError) {
	paketMapel, err := s.repo.PaketMapelID(ctx, paketID)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Paket tidak ditemukan.")
	}
	ids, err := s.repo.ListSoalAcak(ctx, paketMapel, paketID, jumlah)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal mengambil soal acak.")
	}
	maxUrutan, _ := s.repo.MaxUrutanPaketSoal(ctx, paketID)
	for i, soalID := range ids {
		_ = s.repo.InsertPaketSoal(ctx, paketID, soalID, maxUrutan+i+1)
	}
	_ = s.repo.SinkronJumlahSoal(ctx, paketID)
	platform.Catat(ctx, s.pool, actor, "paket.generate_acak", "paket_ujian", paketID,
		map[string]any{"diminta": jumlah, "ditambahkan": len(ids)})
	resp := map[string]any{"paket_id": paketID, "diminta": jumlah, "ditambahkan": len(ids)}
	if len(ids) < jumlah {
		resp["peringatan"] = "Soal tersedia di bank mapel ini kurang dari yang diminta."
	}
	return resp, nil
}
