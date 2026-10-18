package event

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5"
)

// --- Input + validasi ---

var jenisEventValid = map[string]bool{
	"STS": true, "SAS": true, "USBK": true, "UKK": true, "Tryout": true, "Lainnya": true,
}

type eventInput struct {
	Kode           string  `json:"kode"`
	Nama           string  `json:"nama"`
	Jenis          string  `json:"jenis"`
	TanggalMulai   *string `json:"tanggal_mulai"`
	TanggalSelesai *string `json:"tanggal_selesai"`
	Deskripsi      *string `json:"deskripsi"`
}

// parseTanggalValidasi: "" / null -> nil; selain itu harus YYYY-MM-DD.
func parseTanggalValidasi(v *string, field string) (any, error) {
	if v == nil || strings.TrimSpace(*v) == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", strings.TrimSpace(*v))
	if err != nil {
		return nil, fmt.Errorf("%s harus format tanggal YYYY-MM-DD", field)
	}
	return t.Format("2006-01-02"), nil
}

func (in *eventInput) validate() (mulai, selesai any, err error) {
	in.Kode = strings.TrimSpace(in.Kode)
	in.Nama = strings.TrimSpace(in.Nama)
	in.Jenis = strings.TrimSpace(in.Jenis)
	if in.Kode == "" || in.Nama == "" {
		return nil, nil, fmt.Errorf("kode dan nama event wajib diisi")
	}
	if in.Jenis == "" {
		in.Jenis = "Lainnya"
	}
	if !jenisEventValid[in.Jenis] {
		return nil, nil, fmt.Errorf("jenis harus salah satu: STS, SAS, USBK, UKK, Tryout, Lainnya")
	}
	mulai, err = parseTanggalValidasi(in.TanggalMulai, "tanggal_mulai")
	if err != nil {
		return nil, nil, err
	}
	selesai, err = parseTanggalValidasi(in.TanggalSelesai, "tanggal_selesai")
	if err != nil {
		return nil, nil, err
	}
	if mulai != nil && selesai != nil && selesai.(string) < mulai.(string) {
		return nil, nil, fmt.Errorf("tanggal_selesai tidak boleh sebelum tanggal_mulai")
	}
	return mulai, selesai, nil
}

type ruanganInput struct {
	Kode       string  `json:"kode"`
	Nama       string  `json:"nama"`
	Kapasitas  int     `json:"kapasitas"`
	Lokasi     *string `json:"lokasi"`
	Keterangan *string `json:"keterangan"`
}

func (in *ruanganInput) validate() error {
	in.Kode = strings.TrimSpace(in.Kode)
	in.Nama = strings.TrimSpace(in.Nama)
	if in.Kode == "" || in.Nama == "" {
		return fmt.Errorf("kode dan nama ruangan wajib diisi")
	}
	if in.Kapasitas <= 0 {
		return fmt.Errorf("kapasitas harus > 0")
	}
	return nil
}

func optionalText(v *string) any {
	if v == nil || strings.TrimSpace(*v) == "" {
		return nil
	}
	return strings.TrimSpace(*v)
}

// ---------------------------------------------------------------------
// Event
// ---------------------------------------------------------------------

// ListEvent mengembalikan halaman event sesuai filter.
func (s *Service) ListEvent(ctx context.Context, jenis, search any, limit, offset int) ([]eventRow, int, *platform.AppError) {
	items, err := s.repo.ListEvent(ctx, jenis, search, limit, offset)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca event.")
	}
	// Error COUNT diabaikan seperti kode asal (_ = ...) — total 0 bila gagal.
	total, _ := s.repo.CountEvent(ctx, jenis, search)
	return items, total, nil
}

// CreateEvent membuat event baru.
func (s *Service) CreateEvent(ctx context.Context, actor RequestActor, in eventInput) (*eventRow, *platform.AppError) {
	mulai, selesai, err := in.validate()
	if err != nil {
		return nil, platform.ErrBadRequest("validation_error", err.Error())
	}
	dipakai, err := kodeDipakai(ctx, s.pool, "event", in.Kode, "")
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal memeriksa kode event.")
	}
	if dipakai {
		return nil, platform.ErrConflict("duplicate_kode", "Kode event sudah dipakai.")
	}
	ev, err := s.repo.InsertEvent(ctx, in, mulai, selesai)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, platform.ErrConflict("duplicate_kode", "Kode event sudah dipakai.")
		}
		return nil, platform.ErrInternal("internal_error", "Gagal membuat event.")
	}
	s.catat(ctx, actor, "event.create", "event", ev.ID,
		map[string]any{"kode": ev.Kode, "nama": ev.Nama})
	return ev, nil
}

// GetEvent mengembalikan satu event.
func (s *Service) GetEvent(ctx context.Context, id string) (*eventRow, *platform.AppError) {
	ev, found, err := loadEvent(ctx, s.pool, id)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca event.")
	}
	if !found {
		return nil, platform.ErrNotFound("not_found", "Event tidak ditemukan.")
	}
	return ev, nil
}

// UpdateEvent mengubah event.
// LoadEventChecked memuat event (404/500). Dipanggil handler SEBELUM
// bind body agar urutan status identik dengan kode asal.
func (s *Service) LoadEventChecked(ctx context.Context, id string) (*eventRow, *platform.AppError) {
	ev, found, err := loadEvent(ctx, s.pool, id)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca event.")
	}
	if !found {
		return nil, platform.ErrNotFound("not_found", "Event tidak ditemukan.")
	}
	return ev, nil
}

// ApplyEventUpdate memvalidasi input + menyimpan perubahan event.
// Dipanggil handler SETELAH bind body (agar 400 bind mendahului validasi,
// sesuai urutan kode asal).
func (s *Service) ApplyEventUpdate(ctx context.Context, actor RequestActor, ev *eventRow, in eventInput) (*eventRow, *platform.AppError) {
	mulai, selesai, err := in.validate()
	if err != nil {
		return nil, platform.ErrBadRequest("validation_error", err.Error())
	}
	dipakai, err := kodeDipakai(ctx, s.pool, "event", in.Kode, ev.ID)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal memeriksa kode event.")
	}
	if dipakai {
		return nil, platform.ErrConflict("duplicate_kode", "Kode event sudah dipakai event lain.")
	}
	if err := s.repo.UpdateEvent(ctx, ev.ID, in, mulai, selesai); err != nil {
		if isUniqueViolation(err) {
			return nil, platform.ErrConflict("duplicate_kode", "Kode event sudah dipakai event lain.")
		}
		return nil, platform.ErrInternal("internal_error", "Gagal mengubah event.")
	}
	s.catat(ctx, actor, "event.update", "event", ev.ID,
		map[string]any{"kode": in.Kode, "nama": in.Nama})
	updated, _, _ := loadEvent(ctx, s.pool, ev.ID)
	return updated, nil
}

// DeleteEvent menghapus event. Ditolak 409 bila masih dipakai jadwal ujian.
func (s *Service) DeleteEvent(ctx context.Context, actor RequestActor, id string) (map[string]any, *platform.AppError) {
	ev, found, err := loadEvent(ctx, s.pool, id)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca event.")
	}
	if !found {
		return nil, platform.ErrNotFound("not_found", "Event tidak ditemukan.")
	}
	n, err := jadwalMemakai(ctx, s.pool, "event_id", ev.ID)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal memeriksa pemakaian event.")
	}
	if n > 0 {
		return nil, platform.ErrConflict("event_in_use",
			fmt.Sprintf("Event masih dipakai %d jadwal ujian — lepas event dari jadwal dulu.", n))
	}
	if err := s.repo.DeleteEvent(ctx, ev.ID); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menghapus event.")
	}
	s.catat(ctx, actor, "event.delete", "event", ev.ID,
		map[string]any{"kode": ev.Kode, "nama": ev.Nama})
	return map[string]any{"deleted": true, "id": ev.ID}, nil
}

// ---------------------------------------------------------------------
// Ruangan
// ---------------------------------------------------------------------

// ListRuangan mengembalikan halaman ruangan sesuai pencarian.
func (s *Service) ListRuangan(ctx context.Context, search any, limit, offset int) ([]ruanganRow, int, *platform.AppError) {
	items, err := s.repo.ListRuangan(ctx, search, limit, offset)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca ruangan.")
	}
	// Error COUNT diabaikan seperti kode asal (_ = ...) — total 0 bila gagal.
	total, _ := s.repo.CountRuangan(ctx, search)
	return items, total, nil
}

// CreateRuangan membuat ruangan baru.
func (s *Service) CreateRuangan(ctx context.Context, actor RequestActor, in ruanganInput) (*ruanganRow, *platform.AppError) {
	if err := in.validate(); err != nil {
		return nil, platform.ErrBadRequest("validation_error", err.Error())
	}
	dipakai, err := kodeDipakai(ctx, s.pool, "ruangan", in.Kode, "")
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal memeriksa kode ruangan.")
	}
	if dipakai {
		return nil, platform.ErrConflict("duplicate_kode", "Kode ruangan sudah dipakai.")
	}
	r, err := s.repo.InsertRuangan(ctx, in)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, platform.ErrConflict("duplicate_kode", "Kode ruangan sudah dipakai.")
		}
		return nil, platform.ErrInternal("internal_error", "Gagal membuat ruangan.")
	}
	s.catat(ctx, actor, "event.create", "ruangan", r.ID,
		map[string]any{"kode": r.Kode, "nama": r.Nama})
	return r, nil
}

// GetRuangan mengembalikan satu ruangan.
func (s *Service) GetRuangan(ctx context.Context, id string) (*ruanganRow, *platform.AppError) {
	r, found, err := loadRuangan(ctx, s.pool, id)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca ruangan.")
	}
	if !found {
		return nil, platform.ErrNotFound("not_found", "Ruangan tidak ditemukan.")
	}
	return r, nil
}

// UpdateRuangan mengubah ruangan. Kapasitas baru tidak boleh di bawah
// jumlah siswa kelas jadwal yang sudah memakai ruangan ini.
// LoadRuanganChecked memuat ruangan (404/500). Dipanggil handler SEBELUM
// bind body agar urutan status identik dengan kode asal.
func (s *Service) LoadRuanganChecked(ctx context.Context, id string) (*ruanganRow, *platform.AppError) {
	r, found, err := loadRuangan(ctx, s.pool, id)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca ruangan.")
	}
	if !found {
		return nil, platform.ErrNotFound("not_found", "Ruangan tidak ditemukan.")
	}
	return r, nil
}

// ApplyRuanganUpdate memvalidasi input + menyimpan perubahan ruangan.
// Dipanggil handler SETELAH bind body (agar 400 bind mendahului validasi,
// sesuai urutan kode asal).
func (s *Service) ApplyRuanganUpdate(ctx context.Context, actor RequestActor, r *ruanganRow, in ruanganInput) (*ruanganRow, *platform.AppError) {
	if err := in.validate(); err != nil {
		return nil, platform.ErrBadRequest("validation_error", err.Error())
	}
	dipakai, err := kodeDipakai(ctx, s.pool, "ruangan", in.Kode, r.ID)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal memeriksa kode ruangan.")
	}
	if dipakai {
		return nil, platform.ErrConflict("duplicate_kode", "Kode ruangan sudah dipakai ruangan lain.")
	}
	// Kapasitas baru tidak boleh di bawah jumlah siswa kelas jadwal
	// yang sudah memakai ruangan ini. Pesan 409 menyebut jadwalnya.
	// Error query diabaikan seperti kode asal (_ = ...).
	jadwalPenuh, _ := s.repo.RuanganJadwalPenuh(ctx, r.ID, in.Kapasitas)
	if jadwalPenuh != "" {
		return nil, platform.ErrConflict("kapasitas_kurang",
			fmt.Sprintf("Kapasitas baru (%d kursi) kurang dari jumlah siswa kelas jadwal %s yang memakai ruangan ini. Turunkan kapasitas setelah jadwal dipindah.", in.Kapasitas, jadwalPenuh))
	}
	if err := s.repo.UpdateRuangan(ctx, r.ID, in); err != nil {
		if isUniqueViolation(err) {
			return nil, platform.ErrConflict("duplicate_kode", "Kode ruangan sudah dipakai ruangan lain.")
		}
		return nil, platform.ErrInternal("internal_error", "Gagal mengubah ruangan.")
	}
	s.catat(ctx, actor, "event.update", "ruangan", r.ID,
		map[string]any{"kode": in.Kode, "nama": in.Nama})
	updated, _, _ := loadRuangan(ctx, s.pool, r.ID)
	return updated, nil
}

// DeleteRuangan menghapus ruangan. Ditolak 409 bila masih dipakai jadwal ujian.
func (s *Service) DeleteRuangan(ctx context.Context, actor RequestActor, id string) (map[string]any, *platform.AppError) {
	r, found, err := loadRuangan(ctx, s.pool, id)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca ruangan.")
	}
	if !found {
		return nil, platform.ErrNotFound("not_found", "Ruangan tidak ditemukan.")
	}
	n, err := jadwalMemakai(ctx, s.pool, "ruangan_id", r.ID)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal memeriksa pemakaian ruangan.")
	}
	if n > 0 {
		return nil, platform.ErrConflict("ruangan_in_use",
			fmt.Sprintf("Ruangan masih dipakai %d jadwal ujian — lepas ruangan dari jadwal dulu.", n))
	}
	if err := s.repo.DeleteRuangan(ctx, r.ID); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menghapus ruangan.")
	}
	s.catat(ctx, actor, "event.delete", "ruangan", r.ID,
		map[string]any{"kode": r.Kode, "nama": r.Nama})
	return map[string]any{"deleted": true, "id": r.ID}, nil
}

// ---------------------------------------------------------------------
// Tugaskan event + ruangan ke jadwal
// ---------------------------------------------------------------------

// AssignEventRuangan menugaskan/melepas event + ruangan pada jadwal ujian.
// Field tidak dikirim / string kosong = lepas penugasan (NULL).
// Ditolak 409 bila kapasitas ruangan < jumlah siswa aktif kelas jadwal.
// LoadJadwalKelasChecked memuat kelas jadwal (404/500). Dipanggil handler
// SEBELUM bind body agar urutan status identik dengan kode asal.
func (s *Service) LoadJadwalKelasChecked(ctx context.Context, jadwalID string) (kelasID, kelasNama string, appErr *platform.AppError) {
	kelasID, kelasNama, err := s.repo.GetJadwalKelas(ctx, jadwalID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", platform.ErrNotFound("not_found", "Jadwal tidak ditemukan.")
		}
		return "", "", platform.ErrInternal("internal_error", "Gagal membaca jadwal.")
	}
	return kelasID, kelasNama, nil
}

// ApplyAssign menetapkan/melepas event & ruangan pada jadwal.
// Dipanggil handler SETELAH bind body (agar 400 bind mendahului validasi,
// sesuai urutan kode asal).
func (s *Service) ApplyAssign(ctx context.Context, actor RequestActor, jadwalID, kelasID, kelasNama string, eventID, ruanganID *string) (map[string]any, *platform.AppError) {
	// Field tidak dikirim / string kosong = lepas penugasan (NULL).
	var eventArg, ruanganArg any
	if eventID != nil && strings.TrimSpace(*eventID) != "" {
		evID := strings.TrimSpace(*eventID)
		if !isValidUUID(evID) {
			return nil, platform.ErrBadRequest("validation_error", "event_id harus UUID event yang valid.")
		}
		if _, found, err := loadEvent(ctx, s.pool, evID); err != nil {
			return nil, platform.ErrInternal("internal_error", "Gagal memeriksa event.")
		} else if !found {
			return nil, platform.ErrBadRequest("validation_error", "event_id tidak ditemukan.")
		}
		eventArg = evID
	}
	if ruanganID != nil && strings.TrimSpace(*ruanganID) != "" {
		rgID := strings.TrimSpace(*ruanganID)
		if !isValidUUID(rgID) {
			return nil, platform.ErrBadRequest("validation_error", "ruangan_id harus UUID ruangan yang valid.")
		}
		rg, found, err := loadRuangan(ctx, s.pool, rgID)
		if err != nil {
			return nil, platform.ErrInternal("internal_error", "Gagal memeriksa ruangan.")
		}
		if !found {
			return nil, platform.ErrBadRequest("validation_error", "ruangan_id tidak ditemukan.")
		}
		// Cek kapasitas ruangan >= jumlah siswa AKTIF kelas jadwal.
		jumlahSiswa, err := s.repo.CountSiswaAktifKelas(ctx, kelasID)
		if err != nil {
			return nil, platform.ErrInternal("internal_error", "Gagal menghitung siswa kelas.")
		}
		if rg.Kapasitas < jumlahSiswa {
			return nil, platform.ErrConflict("kapasitas_kurang",
				fmt.Sprintf("Ruangan %s berkapasitas %d kursi, kurang dari jumlah siswa kelas %s (%d siswa). Pilih ruangan yang lebih besar.", rg.Nama, rg.Kapasitas, kelasNama, jumlahSiswa))
		}
		ruanganArg = rgID
	}
	if err := s.repo.AssignEventRuangan(ctx, jadwalID, eventArg, ruanganArg); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menugaskan event/ruangan ke jadwal.")
	}
	s.catat(ctx, actor, "event.assign", "jadwal_ujian", jadwalID,
		map[string]any{"event_id": eventArg, "ruangan_id": ruanganArg})
	j, _, _ := loadJadwalKartu(ctx, s.pool, jadwalID)
	var evOut, rgOut any
	if j != nil {
		if j.EventNama != nil {
			evOut = map[string]any{"id": j.EventID, "kode": j.EventKode, "nama": j.EventNama}
		}
		if j.RuanganNama != nil {
			rgOut = map[string]any{"id": j.RuanganID, "nama": j.RuanganNama, "lokasi": j.RuanganLokasi}
		}
	}
	return map[string]any{
		"jadwal_id": jadwalID, "event_id": eventArg, "ruangan_id": ruanganArg,
		"event": evOut, "ruangan": rgOut,
	}, nil
}
