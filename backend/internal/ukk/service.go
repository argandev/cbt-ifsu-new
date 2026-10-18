package ukk

import (
	"context"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"cbt-smk-v2/backend/internal/audit"
	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ---------------------------------------------------------------------
// Helper umum (murni)
// (Dipindah dari ukk.go tanpa perubahan.)
// ---------------------------------------------------------------------

func isValidUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, ch := range s {
		switch i {
		case 8, 13, 18, 23:
			if ch != '-' {
				return false
			}
		default:
			if !(ch >= '0' && ch <= '9' || ch >= 'a' && ch <= 'f' || ch >= 'A' && ch <= 'F') {
				return false
			}
		}
	}
	return true
}

func strPtr(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// strPtrOrNil mengubah *string menjadi argumen SQL (nil bila kosong).
func strPtrOrNil(s *string) any {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil
	}
	return *s
}

// ---------------------------------------------------------------------
// Aturan hitung UKK (satu-satunya tempat rumus ini ditulis)
// ---------------------------------------------------------------------

// round2 membulatkan ke 2 desimal agar nilai tersimpan konsisten.
func round2(v float64) float64 { return math.Round(v*100) / 100 }

// hitungPraktik: proses*0.3 + produk*0.5 + sikap*0.2.
// Mengembalikan nil bila ada satu komponen rubrik yang belum diisi.
func hitungPraktik(proses, produk, sikap *float64) *float64 {
	if proses == nil || produk == nil || sikap == nil {
		return nil
	}
	v := round2(*proses*0.3 + *produk*0.5 + *sikap*0.2)
	return &v
}

// hitungAkhir: teori*0.4 + praktik*0.6. Nil bila teori/praktik belum lengkap.
func hitungAkhir(teori, praktik *float64) *float64 {
	if teori == nil || praktik == nil {
		return nil
	}
	v := round2(*teori*0.4 + *praktik*0.6)
	return &v
}

// Predikat dari nilai akhir 0-100 (sama dengan rekap: <70 Belum Tuntas,
// 70-79 Cukup, 80-90 Baik, 91-100 Sangat Baik).
func Predikat(nilai float64) string {
	switch {
	case nilai < 70:
		return "Belum Tuntas"
	case nilai < 80:
		return "Cukup"
	case nilai <= 90:
		return "Baik"
	default:
		return "Sangat Baik"
	}
}

// validNilai: nilai 0-100, bukan NaN/Inf.
func validNilai(v *float64) bool {
	if v == nil {
		return true
	}
	return !math.IsNaN(*v) && !math.IsInf(*v, 0) && *v >= 0 && *v <= 100
}

// ---------------------------------------------------------------------
// Input
// ---------------------------------------------------------------------

type ukkInput struct {
	Nama             string  `json:"nama"`
	MapelID          string  `json:"mapel_id"`
	KelasID          string  `json:"kelas_id"`
	JurusanID        string  `json:"jurusan_id"`
	Tanggal          *string `json:"tanggal"`
	PengujiInternal  *string `json:"penguji_internal"`
	PengujiEksternal *string `json:"penguji_eksternal"`
	Status           *string `json:"status"`
}

func (in ukkInput) validate() string {
	if strings.TrimSpace(in.Nama) == "" {
		return "Nama UKK wajib diisi."
	}
	if !isValidUUID(in.MapelID) {
		return "mapel_id tidak valid."
	}
	if !isValidUUID(in.KelasID) {
		return "kelas_id tidak valid."
	}
	if !isValidUUID(in.JurusanID) {
		return "jurusan_id tidak valid."
	}
	if in.Status != nil && *in.Status != "draft" && *in.Status != "berjalan" && *in.Status != "selesai" {
		return "status harus draft, berjalan, atau selesai."
	}
	if in.Tanggal != nil && strings.TrimSpace(*in.Tanggal) != "" {
		if _, err := time.Parse("2006-01-02", strings.TrimSpace(*in.Tanggal)); err != nil {
			return "tanggal harus format YYYY-MM-DD."
		}
	}
	return ""
}

func (in ukkInput) tanggalArg() any {
	if in.Tanggal == nil || strings.TrimSpace(*in.Tanggal) == "" {
		return nil
	}
	t, _ := time.Parse("2006-01-02", strings.TrimSpace(*in.Tanggal))
	return t
}

type nilaiInput struct {
	SiswaID     string   `json:"siswa_id"`
	NilaiTeori  *float64 `json:"nilai_teori"`
	NilaiProses *float64 `json:"nilai_proses"`
	NilaiProduk *float64 `json:"nilai_produk"`
	NilaiSikap  *float64 `json:"nilai_sikap"`
	Catatan     *string  `json:"catatan"`
}

func (in nilaiInput) validate() string {
	if !validNilai(in.NilaiTeori) || !validNilai(in.NilaiProses) ||
		!validNilai(in.NilaiProduk) || !validNilai(in.NilaiSikap) {
		return "Semua nilai harus 0-100."
	}
	return ""
}

// hitungNilaiInput menghitung nilai turunan dari input mentah.
func hitungNilaiInput(in nilaiInput) (praktik, akhir *float64, lulus *bool, predikat *string) {
	praktik = hitungPraktik(in.NilaiProses, in.NilaiProduk, in.NilaiSikap)
	akhir = hitungAkhir(in.NilaiTeori, praktik)
	if akhir != nil {
		l := *akhir >= 70
		lulus = &l
		p := Predikat(*akhir)
		predikat = &p
	}
	return praktik, akhir, lulus, predikat
}

var unsafeNameChars = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func sanitizeFileName(name string) string {
	base := filepath.Base(name)
	base = unsafeNameChars.ReplaceAllString(base, "_")
	base = strings.Trim(base, "._")
	if base == "" {
		base = "berkas"
	}
	if len(base) > 120 {
		base = base[:120]
	}
	return base
}

// ---------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------

// Batas upload berkas praktik UKK.
const maxBerkasBytes = 10 << 20 // 10 MB

// Ekstensi berkas praktik yang diizinkan (jpeg = alias jpg).
var allowedExt = map[string]bool{
	".pdf": true, ".doc": true, ".docx": true,
	".jpg": true, ".jpeg": true, ".png": true, ".zip": true,
}

// RequestActor adalah identitas pemanggil untuk guard + audit log.
// Dibangun handler dari echo.Context agar service bebas dari HTTP.
type RequestActor struct {
	UserID string
	Role   string
	RealIP string
}

// Service adalah lapisan logika bisnis paket ukk.
type Service struct {
	repo *Repository
	pool *pgxpool.Pool
	// storageDir adalah akar direktori penyimpanan berkas
	// (dari config.StorageDir; default "storage").
	storageDir string
}

// NewService membuat Service.
func NewService(repo *Repository, pool *pgxpool.Pool, storageDir string) *Service {
	if storageDir == "" {
		storageDir = "storage"
	}
	return &Service{repo: repo, pool: pool, storageDir: storageDir}
}

// catat menulis audit log. (Dulu: audit.Catat di handler.)
func (s *Service) catat(ctx context.Context, actor RequestActor, aksi, entitas, entitasID string, detail any) {
	_ = audit.Catat(ctx, s.pool, actor.UserID, aksi, entitas, entitasID, detail, actor.RealIP)
}

// ---------------------------------------------------------------------
// CRUD ukk_ujian
// ---------------------------------------------------------------------

// ListUkk mengembalikan halaman UKK sesuai filter.
func (s *Service) ListUkk(ctx context.Context, jurusanID, kelasID, status, search any, limit, offset int) ([]ukkRow, int64, *platform.AppError) {
	total, err := s.repo.CountUkk(ctx, jurusanID, kelasID, status, search)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca daftar UKK.")
	}
	items, err := s.repo.ListUkkRows(ctx, jurusanID, kelasID, status, search, limit, offset)
	if err != nil {
		return nil, 0, platform.ErrInternal("internal_error", "Gagal membaca daftar UKK.")
	}
	return items, total, nil
}

// CreateUkk membuat UKK baru. Mengembalikan baris lengkap (201 di handler).
func (s *Service) CreateUkk(ctx context.Context, actor RequestActor, in ukkInput) (*ukkRow, *platform.AppError) {
	if msg := in.validate(); msg != "" {
		return nil, platform.ErrBadRequest("validation_error", msg)
	}
	if !s.repo.ExistsID(ctx, "mapel", in.MapelID) {
		return nil, platform.ErrBadRequest("validation_error", "Mapel tidak ditemukan.")
	}
	if !s.repo.ExistsID(ctx, "kelas", in.KelasID) {
		return nil, platform.ErrBadRequest("validation_error", "Kelas tidak ditemukan.")
	}
	if !s.repo.ExistsID(ctx, "jurusan", in.JurusanID) {
		return nil, platform.ErrBadRequest("validation_error", "Jurusan tidak ditemukan.")
	}
	status := "draft"
	if in.Status != nil && *in.Status != "" {
		status = *in.Status
	}
	id, err := s.repo.InsertUkk(ctx, strings.TrimSpace(in.Nama), in.MapelID, in.KelasID, in.JurusanID,
		in.tanggalArg(), strPtrOrNil(in.PengujiInternal), strPtrOrNil(in.PengujiEksternal), status)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membuat UKK.")
	}
	s.catat(ctx, actor, "ukk.buat", "ukk_ujian", id,
		map[string]any{"nama": strings.TrimSpace(in.Nama), "status": status})
	u, err := s.repo.GetUkkRow(ctx, id)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca UKK baru.")
	}
	return u, nil
}

// GetUkk mengembalikan satu UKK. Tidak ada -> 404.
func (s *Service) GetUkk(ctx context.Context, id string) (*ukkRow, *platform.AppError) {
	u, err := s.repo.GetUkkRow(ctx, id)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "UKK tidak ditemukan.")
	}
	return u, nil
}

// UpdateUkk mengubah satu UKK.
func (s *Service) UpdateUkk(ctx context.Context, actor RequestActor, id string, in ukkInput) (*ukkRow, *platform.AppError) {
	if msg := in.validate(); msg != "" {
		return nil, platform.ErrBadRequest("validation_error", msg)
	}
	old, err := s.repo.GetUkkRow(ctx, id)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "UKK tidak ditemukan.")
	}
	if !s.repo.ExistsID(ctx, "mapel", in.MapelID) {
		return nil, platform.ErrBadRequest("validation_error", "Mapel tidak ditemukan.")
	}
	if !s.repo.ExistsID(ctx, "kelas", in.KelasID) {
		return nil, platform.ErrBadRequest("validation_error", "Kelas tidak ditemukan.")
	}
	if !s.repo.ExistsID(ctx, "jurusan", in.JurusanID) {
		return nil, platform.ErrBadRequest("validation_error", "Jurusan tidak ditemukan.")
	}
	status := old.Status
	if in.Status != nil && *in.Status != "" {
		status = *in.Status
	}
	if err := s.repo.UpdateUkk(ctx, id, strings.TrimSpace(in.Nama), in.MapelID, in.KelasID, in.JurusanID,
		in.tanggalArg(), strPtrOrNil(in.PengujiInternal), strPtrOrNil(in.PengujiEksternal), status); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal mengubah UKK.")
	}
	s.catat(ctx, actor, "ukk.ubah", "ukk_ujian", id,
		map[string]any{"nama": strings.TrimSpace(in.Nama), "status_lama": old.Status, "status_baru": status})
	u, err := s.repo.GetUkkRow(ctx, id)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca UKK.")
	}
	return u, nil
}

// DeleteUkk menghapus UKK yang belum punya nilai/berkas/sertifikat.
func (s *Service) DeleteUkk(ctx context.Context, actor RequestActor, id string) *platform.AppError {
	if _, err := s.repo.GetUkkRow(ctx, id); err != nil {
		return platform.ErrNotFound("not_found", "UKK tidak ditemukan.")
	}
	nNilai, nBerkas, nSert := s.repo.CountAnak(ctx, id)
	if nNilai+nBerkas+nSert > 0 {
		return platform.ErrConflict("conflict",
			"UKK tidak dapat dihapus karena sudah memiliki nilai/berkas/sertifikat.")
	}
	if err := s.repo.DeleteUkk(ctx, id); err != nil {
		return platform.ErrInternal("internal_error", "Gagal menghapus UKK.")
	}
	s.catat(ctx, actor, "ukk.hapus", "ukk_ujian", id, nil)
	return nil
}

// ---------------------------------------------------------------------
// Nilai UKK
// ---------------------------------------------------------------------

// ListNilai mengembalikan semua nilai satu UKK.
func (s *Service) ListNilai(ctx context.Context, ukkID string) ([]nilaiRow, *platform.AppError) {
	if _, err := s.repo.GetUkkRow(ctx, ukkID); err != nil {
		return nil, platform.ErrNotFound("not_found", "UKK tidak ditemukan.")
	}
	items, err := s.repo.ListNilaiRows(ctx, ukkID)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca nilai UKK.")
	}
	return items, nil
}

// ListPeserta mengembalikan daftar siswa kelas UKK (dropdown halaman nilai).
func (s *Service) ListPeserta(ctx context.Context, ukkID string) ([]pesertaRow, *platform.AppError) {
	ukk, err := s.repo.GetUkkRow(ctx, ukkID)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "UKK tidak ditemukan.")
	}
	items, err := s.repo.ListPesertaRows(ctx, ukk.KelasID)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca peserta UKK.")
	}
	return items, nil
}

// UpsertNilai menangani POST /ukk/:id/nilai (upsert per siswa, nilaiID
// kosong) dan PUT /ukk/:id/nilai/:nilaiId (per nilai id).
// Mengembalikan baris + apakah baris baru dibuat (untuk status 201).
func (s *Service) UpsertNilai(ctx context.Context, actor RequestActor, ukkID, nilaiID string, in nilaiInput) (*nilaiRow, bool, *platform.AppError) {
	if msg := in.validate(); msg != "" {
		return nil, false, platform.ErrBadRequest("validation_error", msg)
	}
	ukk, err := s.repo.GetUkkRow(ctx, ukkID)
	if err != nil {
		return nil, false, platform.ErrNotFound("not_found", "UKK tidak ditemukan.")
	}
	// Nilai ditolak bila UKK sudah selesai, kecuali admin (+ audit override).
	if ukk.Status == "selesai" && actor.Role != "admin" {
		return nil, false, platform.ErrForbidden("forbidden",
			"UKK sudah selesai: nilai tidak dapat diubah kecuali oleh admin.")
	}
	override := ukk.Status == "selesai" && actor.Role == "admin"

	praktik, akhir, lulus, predikat := hitungNilaiInput(in)
	var row *nilaiRow
	var created bool
	if nilaiID != "" {
		// PUT per nilai id: siswa_id diambil dari baris yang ada.
		existingSiswaID, err := s.repo.SiswaIDByNilaiID(ctx, nilaiID, ukkID)
		if err != nil {
			return nil, false, platform.ErrNotFound("not_found", "Nilai UKK tidak ditemukan.")
		}
		in.SiswaID = existingSiswaID
		row, err = s.repo.UpdateNilaiRow(ctx, nilaiID, in, praktik, akhir, lulus, predikat, actor.UserID)
		if err != nil {
			return nil, false, platform.ErrInternal("internal_error", "Gagal menyimpan nilai UKK.")
		}
	} else {
		// POST: upsert per (ukk_id, siswa_id).
		if !isValidUUID(in.SiswaID) {
			return nil, false, platform.ErrBadRequest("validation_error", "siswa_id tidak valid.")
		}
		if !s.repo.ExistsID(ctx, "siswa", in.SiswaID) {
			return nil, false, platform.ErrBadRequest("validation_error", "Siswa tidak ditemukan.")
		}
		created = !s.repo.NilaiExists(ctx, ukkID, in.SiswaID)
		row, err = s.repo.UpsertNilaiRow(ctx, ukkID, in.SiswaID, in, praktik, akhir, lulus, predikat, actor.UserID)
		if err != nil {
			return nil, false, platform.ErrInternal("internal_error", "Gagal menyimpan nilai UKK.")
		}
	}
	// Lengkapi info siswa untuk respons.
	row.SiswaNISN, row.SiswaNama = s.repo.SiswaNISNNama(ctx, row.SiswaID)
	detail := map[string]any{"siswa_id": row.SiswaID, "nilai_akhir": row.NilaiAkhir, "lulus": row.Lulus}
	if override {
		detail["override_ukk_selesai"] = true
	}
	s.catat(ctx, actor, "ukk.nilai", "ukk_nilai", row.ID, detail)
	return row, created, nil
}

// ResolveSiswaID memetakan user login ke siswa.id.
func (s *Service) ResolveSiswaID(ctx context.Context, userID string) (string, bool) {
	return s.repo.ResolveSiswaID(ctx, userID)
}

// SiswaExists melaporkan apakah siswa ada.
func (s *Service) SiswaExists(ctx context.Context, siswaID string) bool {
	return s.repo.ExistsID(ctx, "siswa", siswaID)
}

// ---------------------------------------------------------------------
// Berkas praktik
// ---------------------------------------------------------------------

// ListBerkas mengembalikan berkas satu UKK, opsional filter siswa.
func (s *Service) ListBerkas(ctx context.Context, ukkID string, filterSiswa *string) ([]berkasRow, *platform.AppError) {
	if _, err := s.repo.GetUkkRow(ctx, ukkID); err != nil {
		return nil, platform.ErrNotFound("not_found", "UKK tidak ditemukan.")
	}
	items, err := s.repo.ListBerkasRows(ctx, ukkID, filterSiswa)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca berkas UKK.")
	}
	return items, nil
}

// UploadBerkas menyimpan file fisik + mencatat di DB. (Dulu: bagian
// akhir uploadBerkas setelah validasi input di handler.)
func (s *Service) UploadBerkas(ctx context.Context, actor RequestActor, ukkID, siswaID string, fileHeader *multipart.FileHeader) (*berkasRow, *platform.AppError) {
	src, err := fileHeader.Open()
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca berkas upload.")
	}
	defer src.Close()

	safeName := sanitizeFileName(fileHeader.Filename)
	// Prefix unik: cegah tabrakan nama + nama file tidak dapat ditebak.
	diskName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), safeName)
	relDir := filepath.Join(s.storageDir, "ukk", ukkID, siswaID)
	absDir, _ := filepath.Abs(relDir)
	if err := os.MkdirAll(absDir, 0o750); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menyiapkan penyimpanan berkas.")
	}
	relPath := filepath.Join(relDir, diskName)
	absPath, _ := filepath.Abs(relPath)
	dst, err := os.Create(absPath)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menyimpan berkas.")
	}
	written, copyErr := io.Copy(dst, src)
	_ = dst.Close()
	if copyErr != nil {
		_ = os.Remove(absPath)
		return nil, platform.ErrInternal("internal_error", "Gagal menyimpan berkas.")
	}
	if written > maxBerkasBytes {
		_ = os.Remove(absPath)
		return nil, platform.ErrBadRequest("validation_error", "Ukuran berkas melebihi 10 MB.")
	}

	b, err := s.repo.InsertBerkas(ctx, ukkID, siswaID, fileHeader.Filename, relPath, written, actor.UserID)
	if err != nil {
		_ = os.Remove(absPath)
		return nil, platform.ErrInternal("internal_error", "Gagal mencatat berkas UKK.")
	}
	s.catat(ctx, actor, "ukk.berkas.upload", "ukk_berkas", b.ID,
		map[string]any{"ukk_id": ukkID, "siswa_id": siswaID, "nama_file": fileHeader.Filename, "ukuran": written})
	return b, nil
}

// BerkasUnduhan adalah file berkas siap di-attachment.
type BerkasUnduhan struct {
	AbsPath  string
	NamaFile string
}

// DownloadBerkas memeriksa akses + menyiapkan file untuk diunduh.
// bolehStaf dihitung handler via audit.HasPermission (butuh echo.Context).
func (s *Service) DownloadBerkas(ctx context.Context, actor RequestActor, ukkID, berkasID string, bolehStaf bool) (*BerkasUnduhan, *platform.AppError) {
	if _, err := s.repo.GetUkkRow(ctx, ukkID); err != nil {
		return nil, platform.ErrNotFound("not_found", "UKK tidak ditemukan.")
	}
	b, storedPath, err := s.repo.GetBerkasRow(ctx, berkasID, ukkID)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Berkas tidak ditemukan.")
	}
	// Siswa hanya boleh mengunduh berkas miliknya sendiri.
	if actor.Role == "siswa" {
		sid, found := s.repo.ResolveSiswaID(ctx, actor.UserID)
		if !found || sid != b.SiswaID {
			return nil, platform.ErrNotFound("not_found", "Berkas tidak ditemukan.")
		}
	} else if !bolehStaf {
		return nil, platform.ErrForbidden("forbidden", "Tidak punya izin: ukk.read.")
	}
	// Keamanan path: file harus berada di dalam <storageDir>/ukk/.
	absRoot, _ := filepath.Abs(filepath.Join(s.storageDir, "ukk"))
	absPath, _ := filepath.Abs(storedPath)
	if absPath != absRoot && !strings.HasPrefix(absPath, absRoot+string(os.PathSeparator)) {
		return nil, platform.ErrNotFound("not_found", "Berkas tidak ditemukan.")
	}
	if _, err := os.Stat(absPath); err != nil {
		return nil, platform.ErrNotFound("not_found", "File berkas tidak ditemukan di penyimpanan.")
	}
	s.catat(ctx, actor, "ukk.berkas.download", "ukk_berkas", b.ID,
		map[string]any{"ukk_id": ukkID, "siswa_id": b.SiswaID, "nama_file": b.NamaFile})
	return &BerkasUnduhan{AbsPath: absPath, NamaFile: b.NamaFile}, nil
}

// ---------------------------------------------------------------------
// Sertifikat
// ---------------------------------------------------------------------

// GetSertifikat mengembalikan sertifikat satu siswa pada satu UKK.
func (s *Service) GetSertifikat(ctx context.Context, actor RequestActor, ukkID, siswaID string, bolehStaf bool) (*sertifikatRow, *platform.AppError) {
	if _, err := s.repo.GetUkkRow(ctx, ukkID); err != nil {
		return nil, platform.ErrNotFound("not_found", "UKK tidak ditemukan.")
	}
	if actor.Role == "siswa" {
		sid, found := s.repo.ResolveSiswaID(ctx, actor.UserID)
		if !found || sid != siswaID {
			return nil, platform.ErrNotFound("not_found", "Sertifikat tidak ditemukan.")
		}
	} else if !bolehStaf {
		return nil, platform.ErrForbidden("forbidden", "Tidak punya izin: ukk.read.")
	}
	srt, err := s.repo.GetSertifikatRow(ctx, ukkID, siswaID)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Sertifikat tidak ditemukan.")
	}
	return srt, nil
}

// CreateSertifikat menerbitkan sertifikat (idempoten) bagi siswa yang lulus.
// Mengembalikan baris + apakah sertifikat baru diterbitkan (untuk 201)
// atau sudah ada sebelumnya (untuk 200).
func (s *Service) CreateSertifikat(ctx context.Context, actor RequestActor, ukkID, siswaID string) (*sertifikatRow, bool, *platform.AppError) {
	ukk, err := s.repo.GetUkkRow(ctx, ukkID)
	if err != nil {
		return nil, false, platform.ErrNotFound("not_found", "UKK tidak ditemukan.")
	}
	if !s.repo.ExistsID(ctx, "siswa", siswaID) {
		return nil, false, platform.ErrBadRequest("validation_error", "Siswa tidak ditemukan.")
	}
	// Idempotensi: sertifikat yang sudah ada dikembalikan apa adanya.
	if existing, err := s.repo.GetSertifikatRow(ctx, ukkID, siswaID); err == nil {
		return existing, false, nil
	}
	// Sertifikat HANYA bagi yang lulus -> ditolak 409 bila tidak lulus.
	lulus, err := s.repo.GetLulus(ctx, ukkID, siswaID)
	if err != nil || lulus == nil || !*lulus {
		return nil, false, platform.ErrConflict("conflict",
			"Sertifikat hanya dapat diterbitkan bagi siswa yang lulus UKK (nilai akhir >= 70).")
	}
	// Nomor: UKK/<tahun>/<kode-jurusan>/<nomor-urut 4 digit>.
	// Tahun dari tanggal UKK bila ada, bila tidak tahun berjalan.
	tahun := time.Now().Year()
	if ukk.Tanggal != nil {
		tahun = ukk.Tanggal.Year()
	}
	jurusanKode := ""
	if ukk.JurusanKode != nil {
		jurusanKode = *ukk.JurusanKode
	}
	// Nomor urut per jurusan + tahun. Coba hingga 5x bila bentrok UNIQUE
	// (perlombaan dua penerbitan bersamaan).
	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		count, err := s.repo.CountSertifikatUrut(ctx, ukk.JurusanID, tahun)
		if err != nil {
			return nil, false, platform.ErrInternal("internal_error", "Gagal menerbitkan sertifikat.")
		}
		nomor := fmt.Sprintf("UKK/%d/%s/%04d", tahun, jurusanKode, count+1+int64(attempt))
		lastErr = s.repo.InsertSertifikat(ctx, ukkID, siswaID, nomor)
		if lastErr == nil {
			s.catat(ctx, actor, "ukk.sertifikat.terbit", "ukk_sertifikat", siswaID,
				map[string]any{"ukk_id": ukkID, "siswa_id": siswaID, "nomor_sertifikat": nomor})
			srt, err := s.repo.GetSertifikatRow(ctx, ukkID, siswaID)
			if err != nil {
				return nil, false, platform.ErrInternal("internal_error", "Gagal membaca sertifikat.")
			}
			return srt, true, nil
		}
		// Bukan bentrok nomor unik -> jangan ulangi.
		if !strings.Contains(lastErr.Error(), "duplicate key") && !strings.Contains(lastErr.Error(), "unique") {
			break
		}
	}
	_ = lastErr
	return nil, false, platform.ErrInternal("internal_error", "Gagal menerbitkan sertifikat.")
}
