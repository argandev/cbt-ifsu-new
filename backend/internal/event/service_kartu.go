package event

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cbt-smk-v2/backend/internal/platform"
)

// ---------------------------------------------------------------------
// Helper murni kartu (dipindah dari kartu_pdf.go tanpa perubahan)
// ---------------------------------------------------------------------

// kartuSanitize membersihkan teks untuk font inti Helvetica (Latin-1).
// Tanda baca di luar Latin-1 (en dash, kutip lengkung, elipsis, dst.)
// diganti padanan ASCII; rune di luar Latin-1 -> '?'. Karakter Latin-1
// (mis. e aksen) dibiarkan apa adanya — gofpdf menerjemahkan UTF-8 ke
// Latin-1 untuk font inti.
func kartuSanitize(s string) string {
	s = strings.NewReplacer(
		"–", "-", "—", "-", "‐", "-",
		"“", `"`, "”", `"`, "„", `"`,
		"‘", "'", "’", "'", "‚", "'",
		"…", "...", " ", " ", "•", "-", "·", "-",
	).Replace(s)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '\n' || r == '\t':
			b.WriteByte(' ')
		case r < 32 || r > 255:
			b.WriteByte('?')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

var kartuNamaBulan = []string{"", "Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}

// formatTanggalKartu: "2026-10-17" -> "17 Okt 2026". Input yang gagal
// di-parse dikembalikan apa adanya (jangan pernah kosongkan tanggal).
func formatTanggalKartu(tanggal string) string {
	t, err := time.Parse("2006-01-02", strings.TrimSpace(tanggal))
	if err != nil {
		return tanggal
	}
	return fmt.Sprintf("%d %s %d", t.Day(), kartuNamaBulan[int(t.Month())], t.Year())
}

// namaFileKartuPDF: "kartu-peserta-<nisn>.pdf"; NISN dibersihkan ke
// karakter aman nama file.
func namaFileKartuPDF(nisn string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(nisn) {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "kartu-peserta.pdf"
	}
	return "kartu-peserta-" + b.String() + ".pdf"
}

// ---------------------------------------------------------------------
// Kartu peserta (JSON, LEGACY)
// ---------------------------------------------------------------------

// KartuJadwal menyusun kartu semua siswa aktif di kelas jadwal
// (staf admin/proktor/guru).
func (s *Service) KartuJadwal(ctx context.Context, jadwalID string) (map[string]any, *platform.AppError) {
	j, found, err := loadJadwalKartu(ctx, s.pool, jadwalID)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca jadwal.")
	}
	if !found {
		return nil, platform.ErrNotFound("not_found", "Jadwal tidak ditemukan.")
	}
	items, err := buildKartuList(ctx, s.pool, j)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menyusun kartu peserta.")
	}
	var evOut, rgOut any
	if j.EventNama != nil {
		evOut = map[string]any{"id": j.EventID, "kode": j.EventKode, "nama": j.EventNama}
	}
	if j.RuanganNama != nil {
		rgOut = map[string]any{"id": j.RuanganID, "nama": j.RuanganNama, "lokasi": j.RuanganLokasi}
	}
	return map[string]any{
		"jadwal": map[string]any{
			"id": j.ID, "tanggal": j.Tanggal, "jam_mulai": j.JamMulai, "jam_selesai": j.JamSelesai,
			"status": j.Status, "paket_nama": j.PaketNama, "paket_jenis": j.PaketJenis,
			"mapel_nama": j.MapelNama, "kelas_nama": j.KelasNama, "jurusan_nama": j.JurusanNama,
			"event": evOut, "ruangan": rgOut,
		},
		"items": items,
		"total": len(items),
	}, nil
}

// KartuSiswa mengembalikan satu kartu peserta.
// Siswa hanya kartunya sendiri (bukan pemilik -> 404, anti-IDOR);
// staf (gate requireKartuStaf sudah di handler) boleh semua kartu
// selama siswa sekelas dengan jadwal.
func (s *Service) KartuSiswa(ctx context.Context, uid, role, jadwalID, siswaID string) (*kartuRow, *platform.AppError) {
	if role == "siswa" {
		// Siswa hanya boleh melihat kartunya sendiri.
		sw, found, err := loadSiswaKartu(ctx, s.pool, siswaID)
		if err != nil {
			return nil, platform.ErrInternal("internal_error", "Gagal membaca kartu peserta.")
		}
		if !found {
			return nil, platform.ErrNotFound("not_found", "Kartu peserta tidak ditemukan.")
		}
		milikSendiri := (sw.UserID != nil && *sw.UserID == uid) || sw.ID == uid
		if !milikSendiri {
			return nil, platform.ErrNotFound("not_found", "Kartu peserta tidak ditemukan.")
		}
		j, found, err := loadJadwalKartu(ctx, s.pool, jadwalID)
		if err != nil {
			return nil, platform.ErrInternal("internal_error", "Gagal membaca jadwal.")
		}
		if !found || j.KelasID != sw.KelasID {
			return nil, platform.ErrNotFound("not_found", "Kartu peserta tidak ditemukan.")
		}
		items, err := buildKartuList(ctx, s.pool, j)
		if err != nil {
			return nil, platform.ErrInternal("internal_error", "Gagal menyusun kartu peserta.")
		}
		for _, k := range items {
			if k.SiswaID == sw.ID {
				kartu := k
				return &kartu, nil
			}
		}
		return nil, platform.ErrNotFound("not_found", "Kartu peserta tidak ditemukan.")
	}
	j, found, err := loadJadwalKartu(ctx, s.pool, jadwalID)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca jadwal.")
	}
	if !found {
		return nil, platform.ErrNotFound("not_found", "Jadwal tidak ditemukan.")
	}
	sw, found, err := loadSiswaKartu(ctx, s.pool, siswaID)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca kartu peserta.")
	}
	if !found || sw.KelasID != j.KelasID {
		return nil, platform.ErrNotFound("not_found", "Kartu peserta tidak ditemukan.")
	}
	items, err := buildKartuList(ctx, s.pool, j)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menyusun kartu peserta.")
	}
	for _, k := range items {
		if k.SiswaID == sw.ID {
			kartu := k
			return &kartu, nil
		}
	}
	return nil, platform.ErrNotFound("not_found", "Kartu peserta tidak ditemukan.")
}
