package rekap

import (
	"context"
	"math"
	"strings"
	"time"

	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Service ----------------------------------------------------------------


// statistik menghitung ringkasan nilai. (Dipindah tanpa perubahan.)
func statistik(rows []nilaiRow) map[string]any {
	if len(rows) == 0 {
		return map[string]any{
			"jumlah_peserta": 0, "rata_rata": nil, "tertinggi": nil, "terendah": nil,
			"jumlah_lulus": 0, "jumlah_menunggu_koreksi": 0,
		}
	}
	sum := 0.0
	max := rows[0].Nilai
	min := rows[0].Nilai
	lulus := 0
	menunggu := 0
	for _, r := range rows {
		sum += r.Nilai
		max = math.Max(max, r.Nilai)
		min = math.Min(min, r.Nilai)
		if r.Nilai >= 70 {
			lulus++
		}
		if r.Status == "menunggu_koreksi" {
			menunggu++
		}
	}
	avg := math.Round(sum/float64(len(rows))*100) / 100
	return map[string]any{
		"jumlah_peserta": len(rows), "rata_rata": avg, "tertinggi": max, "terendah": min,
		"jumlah_lulus": lulus, "jumlah_menunggu_koreksi": menunggu,
	}
}


// Service adalah lapisan logika bisnis paket rekap.
// repository adalah kontrak penyimpanan yang dipakai Service.
// (DIP: Service bergantung pada abstraksi, bukan *Repository konkrit.)
type repository interface {
	BeginTx(ctx context.Context) (pgx.Tx, error)
	FinalisasiNilaiTx(ctx context.Context, tx pgx.Tx, jadwalID string) (int, error)
	GetJadwalInfo(ctx context.Context, jadwalID string) (jadwalInfo, error)
	GetNilaiSiswaTx(ctx context.Context, tx pgx.Tx, jadwalID, siswaID string) (string, float64, error)
	HitungKehadiran(ctx context.Context, jadwalID string) (total, hadir int, err error)
	InsertRemedialTx(ctx context.Context, tx pgx.Tx, jadwalID, siswaID string, nilaiBaru float64) error
	ListNilai(ctx context.Context, jadwalID string, siswaID *string, onlyFinal bool) ([]nilaiRow, error)
	TampilHasilKeSiswa(ctx context.Context) bool
	UpdateNilaiAkhirTx(ctx context.Context, tx pgx.Tx, nilaiAkhirID string, nilai float64, predikat string) error
}

type Service struct {
	repo repository
	pool *pgxpool.Pool
}

// NewService membuat Service.
func NewService(repo repository, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, pool: pool}
}


// RekapSiswa mengembalikan rekap untuk siswa: hanya nilainya sendiri,
// hanya bila final + setting tampil_hasil_ke_siswa aktif.
func (s *Service) RekapSiswa(ctx context.Context, jadwalID, siswaID string) ([]nilaiRow, map[string]any, *platform.AppError) {
	if !s.repo.TampilHasilKeSiswa(ctx) {
		return []nilaiRow{}, statistik(nil), nil
	}
	rows, err := s.repo.ListNilai(ctx, jadwalID, &siswaID, true)
	if err != nil {
		return nil, nil, platform.ErrInternal("internal_error", "Gagal membaca rekap.")
	}
	return rows, statistik(rows), nil
}

// RekapStaf mengembalikan rekap penuh satu jadwal.
func (s *Service) RekapStaf(ctx context.Context, jadwalID string) ([]nilaiRow, map[string]any, *platform.AppError) {
	rows, err := s.repo.ListNilai(ctx, jadwalID, nil, false)
	if err != nil {
		return nil, nil, platform.ErrInternal("internal_error", "Gagal membaca rekap.")
	}
	return rows, statistik(rows), nil
}

// ExportData membaca baris rekap untuk export CSV + mencatat audit.
func (s *Service) ExportData(ctx context.Context, actor platform.RequestActor, jadwalID string) ([]nilaiRow, *platform.AppError) {
	rows, err := s.repo.ListNilai(ctx, jadwalID, nil, false)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca rekap.")
	}
	platform.Catat(ctx, s.pool, actor, "rekap.export", "rekap", jadwalID,
		map[string]any{"jadwal_id": jadwalID, "jumlah_baris": len(rows)})
	return rows, nil
}

// BeritaAcara menyusun data berita acara ujian.
func (s *Service) BeritaAcara(ctx context.Context, jadwalID string) (map[string]any, *platform.AppError) {
	ji, err := s.repo.GetJadwalInfo(ctx, jadwalID)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Jadwal tidak ditemukan.")
	}
	rows, err := s.repo.ListNilai(ctx, jadwalID, nil, false)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca rekap.")
	}
	jumlahAttempt, jumlahHadir, err := s.repo.HitungKehadiran(ctx, jadwalID)
	if err != nil {
		jumlahAttempt = len(rows)
		jumlahHadir = len(rows)
	}
	stats := statistik(rows)
	lulus, _ := stats["jumlah_lulus"].(int)
	// Kelas: dari jadwal bila ada, bila kosong gabungkan kelas peserta.
	kelas := ji.KelasNama
	if kelas == "" {
		seen := map[string]bool{}
		names := []string{}
		for _, r := range rows {
			if r.NamaKelas != "" && !seen[r.NamaKelas] {
				seen[r.NamaKelas] = true
				names = append(names, r.NamaKelas)
			}
		}
		kelas = strings.Join(names, ", ")
	}
	return map[string]any{
		"jadwal_id":      jadwalID,
		"mapel":          ji.MapelNama,
		"kelas":          kelas,
		"tanggal":        ji.Tanggal,
		"jam_mulai":      ji.JamMulai,
		"jam_selesai":    ji.JamSelesai,
		"jumlah_peserta": jumlahAttempt,
		"hadir":          jumlahHadir,
		"tidak_hadir":    jumlahAttempt - jumlahHadir,
		"rata_rata":      stats["rata_rata"],
		"lulus":          lulus,
		"catatan":        "",
		"generated_at":   time.Now().Format(time.RFC3339),
	}, nil
}

// Finalisasi menandai semua nilai jadwal sebagai final + mencatat audit.
func (s *Service) Finalisasi(ctx context.Context, actor platform.RequestActor, jadwalID string) (int, *platform.AppError) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return 0, platform.ErrInternal("internal_error", "Gagal memulai finalisasi.")
	}
	defer tx.Rollback(ctx)
	count, err := s.repo.FinalisasiNilaiTx(ctx, tx, jadwalID)
	if err != nil {
		return 0, platform.ErrInternal("internal_error", "Gagal finalisasi nilai.")
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, platform.ErrInternal("internal_error", "Gagal finalisasi nilai.")
	}
	platform.Catat(ctx, s.pool, actor, "rekap.finalisasi", "rekap", jadwalID,
		map[string]any{"jadwal_id": jadwalID, "jumlah_difinalisasi": count})
	return count, nil
}

// Remedial mencatat nilai remedial (dipakai yang tertinggi) + audit.
func (s *Service) Remedial(ctx context.Context, actor platform.RequestActor, jadwalID, siswaID string, nilaiBaru float64) (map[string]any, *platform.AppError) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal memulai remedial.")
	}
	defer tx.Rollback(ctx)
	if err := s.repo.InsertRemedialTx(ctx, tx, jadwalID, siswaID, nilaiBaru); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menyimpan remedial.")
	}
	nilaiAkhirID, nilaiAwal, err := s.repo.GetNilaiSiswaTx(ctx, tx, jadwalID, siswaID)
	if err != nil {
		// Kontrak lama: kegagalan baca nilai (termasuk tidak ada baris)
		// -> 404 "Nilai akhir siswa pada jadwal ini tidak ditemukan."
		return nil, platform.ErrNotFound("not_found", "Nilai akhir siswa pada jadwal ini tidak ditemukan.")
	}
	nilaiAkhir := math.Max(nilaiAwal, nilaiBaru)
	if err := s.repo.UpdateNilaiAkhirTx(ctx, tx, nilaiAkhirID, nilaiAkhir, platform.Predikat(nilaiAkhir)); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal memperbarui nilai akhir.")
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menyimpan remedial.")
	}
	platform.Catat(ctx, s.pool, actor, "rekap.remedial", "nilai_akhir", nilaiAkhirID,
		map[string]any{"jadwal_id": jadwalID, "siswa_id": siswaID,
			"nilai_awal": nilaiAwal, "nilai_remedial": nilaiBaru, "nilai_akhir": nilaiAkhir})
	return map[string]any{
		"jadwal_id": jadwalID, "siswa_id": siswaID,
		"nilai_awal": nilaiAwal, "nilai_remedial": nilaiBaru,
		"nilai_akhir": nilaiAkhir, "predikat": platform.Predikat(nilaiAkhir),
	}, nil
}
