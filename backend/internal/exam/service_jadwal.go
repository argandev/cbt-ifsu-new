package exam

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"strings"
	"time"

	"cbt-smk-v2/backend/internal/monitor"
	"cbt-smk-v2/backend/internal/platform"
)

// JadwalInput adalah input buat/ubah jadwal (dipetakan dari body JSON).
type JadwalInput struct {
	PaketID    string
	KelasID    string
	Tanggal    string
	JamMulai   string
	JamSelesai string
	Status     string
}

// validateJadwal memvalidasi input jadwal (dulu: bindJadwal).
func validateJadwal(in JadwalInput) (jadwal, *platform.AppError) {
	tanggal := strings.TrimSpace(in.Tanggal)
	jamMulai := strings.TrimSpace(in.JamMulai)
	jamSelesai := strings.TrimSpace(in.JamSelesai)
	if in.PaketID == "" || in.KelasID == "" || tanggal == "" || jamMulai == "" || jamSelesai == "" {
		return jadwal{}, platform.ErrBadRequest("validation_error", "paket_id, kelas_id, tanggal, jam_mulai, jam_selesai wajib diisi.")
	}
	mulai, err1 := gabungTanggalJam(tanggal, jamMulai)
	selesai, err2 := gabungTanggalJam(tanggal, jamSelesai)
	if err1 != nil || err2 != nil || !selesai.After(mulai) {
		return jadwal{}, platform.ErrBadRequest("validation_error", "Tanggal/jam tidak valid; jam_selesai harus sesudah jam_mulai.")
	}
	status := strings.TrimSpace(in.Status)
	if status == "" {
		status = "draft"
	}
	if status != "draft" && status != "aktif" && status != "pause" && status != "selesai" {
		return jadwal{}, platform.ErrBadRequest("validation_error", "status harus draft/aktif/pause/selesai.")
	}
	return jadwal{PaketID: in.PaketID, KelasID: in.KelasID, Tanggal: tanggal,
		JamMulai: jamMulai, JamSelesai: jamSelesai, Status: status}, nil
}

// ListJadwal mengembalikan daftar jadwal ujian.
func (s *Service) ListJadwal(ctx context.Context, actor platform.RequestActor, status, kelasID, tanggal string) (any, *platform.AppError) {
	items, err := s.repo.ListJadwal(ctx, status, kelasID, tanggal)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca jadwal.")
	}
	return map[string]any{"items": items}, nil
}

// CreateJadwal membuat jadwal ujian baru.
func (s *Service) CreateJadwal(ctx context.Context, actor platform.RequestActor, in JadwalInput) (*jadwal, *platform.AppError) {
	j, appErr := validateJadwal(in)
	if appErr != nil {
		return nil, appErr
	}
	paketAda, _ := s.repo.PaketExists(ctx, j.PaketID)
	kelasAda, _ := s.repo.KelasExists(ctx, j.KelasID)
	if !paketAda {
		return nil, platform.ErrBadRequest("validation_error", "paket_id tidak ditemukan.")
	}
	if !kelasAda {
		return nil, platform.ErrBadRequest("validation_error", "kelas_id tidak ditemukan.")
	}
	id, err := s.repo.InsertJadwal(ctx, j)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membuat jadwal.")
	}
	j.ID = id
	platform.Catat(ctx, s.pool, actor, "jadwal.create", "jadwal_ujian", id,
		map[string]any{"paket_id": j.PaketID, "kelas_id": j.KelasID, "tanggal": j.Tanggal})
	return &j, nil
}

// GetJadwalDetail mengembalikan detail jadwal (TANPA token_hash).
func (s *Service) GetJadwalDetail(ctx context.Context, actor platform.RequestActor, id string) (map[string]any, *platform.AppError) {
	j, err := s.repo.GetJadwal(ctx, id)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Jadwal tidak ditemukan.")
	}
	// JANGAN pernah kembalikan token_hash ke client.
	return map[string]any{
		"id": j.ID, "paket_id": j.PaketID, "paket_kode": j.PaketKode,
		"paket_nama": j.PaketNama, "paket_jenis": j.PaketJenis,
		"kelas_id": j.KelasID, "kelas_nama": j.KelasNama,
		"tanggal": j.Tanggal, "jam_mulai": j.JamMulai, "jam_selesai": j.JamSelesai,
		"status": j.Status, "durasi_menit": j.DurasiMenit,
		"tampil_hasil": j.TampilHasil, "created_at": j.CreatedAt,
		"token_ada":        j.TokenHash != nil,
		"token_expires_at": j.TokenExpiresAt,
	}, nil
}

// UpdateJadwal mengubah jadwal ujian.
func (s *Service) UpdateJadwal(ctx context.Context, actor platform.RequestActor, id string, in JadwalInput) (map[string]any, *platform.AppError) {
	j, appErr := validateJadwal(in)
	if appErr != nil {
		return nil, appErr
	}
	lamaStatus, err := s.repo.JadwalStatusOnly(ctx, id)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Jadwal tidak ditemukan.")
	}
	if lamaStatus == "selesai" {
		return nil, platform.ErrConflict("conflict", "Jadwal yang sudah selesai tidak boleh diubah.")
	}
	n, err := s.repo.UpdateJadwal(ctx, id, j)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal mengubah jadwal.")
	}
	if n == 0 {
		return nil, platform.ErrNotFound("not_found", "Jadwal tidak ditemukan.")
	}
	platform.Catat(ctx, s.pool, actor, "jadwal.update", "jadwal_ujian", id,
		map[string]any{"status": j.Status})
	return map[string]any{"id": id, "updated": true}, nil
}

// DeleteJadwal menghapus jadwal (ditolak bila sudah punya attempt).
func (s *Service) DeleteJadwal(ctx context.Context, actor platform.RequestActor, id string) (map[string]any, *platform.AppError) {
	jumlahAttempt, _ := s.repo.CountAttemptJadwal(ctx, id)
	if jumlahAttempt > 0 {
		return nil, platform.ErrConflict("conflict", "Jadwal sudah punya attempt peserta. Tidak boleh dihapus (hasil ujian tidak pernah dihapus fisik).")
	}
	n, err := s.repo.DeleteJadwal(ctx, id)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menghapus jadwal.")
	}
	if n == 0 {
		return nil, platform.ErrNotFound("not_found", "Jadwal tidak ditemukan.")
	}
	platform.Catat(ctx, s.pool, actor, "jadwal.delete", "jadwal_ujian", id, nil)
	return map[string]any{"id": id, "deleted": true}, nil
}

// PauseJadwal menjeda jadwal + membekukan timer peserta.
func (s *Service) PauseJadwal(ctx context.Context, actor platform.RequestActor, id string) (map[string]any, *platform.AppError) {
	j, err := s.repo.GetJadwal(ctx, id)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Jadwal tidak ditemukan.")
	}
	if j.Status != "aktif" {
		return nil, platform.ErrConflict("conflict", "Hanya jadwal aktif yang bisa dijeda (status: "+j.Status+").")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal memulai pause.")
	}
	defer tx.Rollback(ctx)
	if err := s.repo.PauseJadwalStatus(ctx, tx, id); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menjeda jadwal.")
	}
	// Bekukan sisa waktu semua attempt berjalan: sisa = deadline - now.
	if err := s.repo.BekukanTimer(ctx, tx, id); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membekukan timer peserta.")
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menjeda jadwal.")
	}
	platform.Catat(ctx, s.pool, actor, "jadwal.pause", "jadwal_ujian", id, nil)
	monitor.Broadcast(id, map[string]any{"type": "jadwal_pause", "jadwal_id": id})
	return map[string]any{"id": id, "status": "pause"}, nil
}

// ResumeJadwal melanjutkan jadwal yang dijeda + memulihkan timer.
func (s *Service) ResumeJadwal(ctx context.Context, actor platform.RequestActor, id string) (map[string]any, *platform.AppError) {
	j, err := s.repo.GetJadwal(ctx, id)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Jadwal tidak ditemukan.")
	}
	if j.Status != "pause" {
		return nil, platform.ErrConflict("conflict", "Hanya jadwal yang dijeda yang bisa dilanjutkan (status: "+j.Status+").")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal memulai resume.")
	}
	defer tx.Rollback(ctx)
	if err := s.repo.AktifkanJadwal(ctx, tx, id); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal melanjutkan jadwal.")
	}
	// Deadline baru = now() + sisa_waktu_beku.
	if err := s.repo.PulihkanTimer(ctx, tx, id); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal memulihkan timer peserta.")
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal melanjutkan jadwal.")
	}
	platform.Catat(ctx, s.pool, actor, "jadwal.resume", "jadwal_ujian", id, nil)
	monitor.Broadcast(id, map[string]any{"type": "jadwal_resume", "jadwal_id": id})
	return map[string]any{"id": id, "status": "aktif"}, nil
}

// --- Token ---

const tokenAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 32 simbol, tanpa 0/1/I/O

func generateToken() (string, error) {
	var sb strings.Builder
	for i := 0; i < 6; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(tokenAlphabet))))
		if err != nil {
			return "", err
		}
		sb.WriteByte(tokenAlphabet[n.Int64()])
	}
	return sb.String(), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// GenerateToken membuat token baru untuk jadwal (token lama HANGUS).
// Token plaintext dikembalikan SEKALI di respons ini.
func (s *Service) GenerateToken(ctx context.Context, actor platform.RequestActor, id string) (map[string]any, *platform.AppError) {
	j, err := s.repo.GetJadwal(ctx, id)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Jadwal tidak ditemukan.")
	}
	if j.Status == "selesai" {
		return nil, platform.ErrConflict("conflict", "Jadwal sudah selesai; token tidak bisa dibuat.")
	}
	token, err := generateToken()
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membuat token.")
	}
	// Expiry = waktu GENERATE + token_expiry_detik (default 900 dtk = 15 menit).
	// Sebelumnya expiry = jam mulai jadwal + 900 dtk, sehingga token yang
	// di-generate setelah jam mulai + 15 menit langsung lahir KEDALUWARSA
	// (mis. jadwal mulai 18.43, generate jam 20.00 → expiry 18.58).
	// Token kedua tetap menimpa (HANGUSKAN) token pertama: hanya satu hash disimpan.
	expiry := time.Now().In(jakarta()).Add(time.Duration(s.repo.TokenExpiryDetik(ctx)) * time.Second)
	if err := s.repo.UpdateToken(ctx, id, hashToken(token), expiry); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menyimpan token.")
	}
	// Audit TIDAK BOLEH mencatat token plaintext (lihat docs/KEAMANAN.md).
	platform.Catat(ctx, s.pool, actor, "jadwal.token.generate", "jadwal_ujian", id,
		map[string]any{"token_expires_at": expiry.Format(time.RFC3339)})
	return map[string]any{
		"jadwal_id": id, "token": token,
		"token_expires_at": expiry.Format(time.RFC3339),
		"peringatan":       "Token hanya ditampilkan sekali. Salin sekarang — token tidak bisa dilihat lagi.",
	}, nil
}

// TokenStatus mengembalikan status token (TANPA menampilkan token).
func (s *Service) TokenStatus(ctx context.Context, actor platform.RequestActor, id string) (map[string]any, *platform.AppError) {
	j, err := s.repo.GetJadwal(ctx, id)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Jadwal tidak ditemukan.")
	}
	// TANPA menampilkan token (nilai token tidak pernah dikembalikan).
	resp := map[string]any{"jadwal_id": id, "ada_token": j.TokenHash != nil}
	if j.TokenExpiresAt != nil {
		if t, ok := (*j.TokenExpiresAt).(time.Time); ok {
			resp["token_expires_at"] = t.Format(time.RFC3339)
			sisa := int(time.Until(t).Seconds())
			if sisa < 0 {
				sisa = 0
			}
			resp["kedaluwarsa_dalam_detik"] = sisa
			resp["kedaluwarsa"] = sisa == 0
		}
	}
	return resp, nil
}
