package exam

import (
	"context"
	"strings"
	"time"

	"cbt-smk-v2/backend/internal/monitor"
	"cbt-smk-v2/backend/internal/platform"
)

// ListUsulanUnblock mengembalikan usulan unblock satu jadwal.
func (s *Service) ListUsulanUnblock(ctx context.Context, actor RequestActor, jadwalID string) (any, *platform.AppError) {
	items, err := s.repo.ListUsulanUnblock(ctx, jadwalID)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca usulan unblock.")
	}
	return map[string]any{"jadwal_id": jadwalID, "items": items}, nil
}

// CreateUsulanUnblock mencatat usulan unblock dari pengawas/proktor/admin.
func (s *Service) CreateUsulanUnblock(ctx context.Context, actor RequestActor, jadwalID, attemptID, alasan string) (map[string]any, *platform.AppError) {
	alasan = strings.TrimSpace(alasan)
	if attemptID == "" || alasan == "" {
		return nil, platform.ErrBadRequest("validation_error", "attempt_id dan alasan wajib diisi.")
	}
	a, err := s.repo.GetAttempt(ctx, attemptID)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Attempt tidak ditemukan.")
	}
	if a.JadwalID != jadwalID {
		return nil, platform.ErrBadRequest("validation_error", "Attempt bukan dari jadwal ini.")
	}
	if a.Status != "diblokir" {
		return nil, platform.ErrConflict("conflict", "Hanya attempt yang diblokir yang bisa diusulkan unblock (status: "+a.Status+").")
	}
	pending, _ := s.repo.CountUsulanMenunggu(ctx, attemptID)
	if pending > 0 {
		return nil, platform.ErrConflict("conflict", "Sudah ada usulan unblock yang menunggu untuk attempt ini.")
	}
	id, err := s.repo.InsertUsulan(ctx, attemptID, actor.UserID, alasan)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menyimpan usulan unblock.")
	}
	s.catat(ctx, actor, "jadwal.usulan_unblock", "attempt", attemptID,
		map[string]any{"jadwal_id": jadwalID, "alasan": alasan})
	monitor.Broadcast(jadwalID, map[string]any{"type": "usulan_unblock", "jadwal_id": jadwalID, "attempt_id": attemptID})
	return map[string]any{"id": id, "attempt_id": attemptID, "status": "menunggu"}, nil
}

// UnblockAttempt meng-unblock attempt (proktor/admin): LANJUT sisa waktu.
func (s *Service) UnblockAttempt(ctx context.Context, actor RequestActor, jadwalID, attemptID, alasan string) (map[string]any, *platform.AppError) {
	alasan = strings.TrimSpace(alasan)
	if attemptID == "" || alasan == "" {
		return nil, platform.ErrBadRequest("validation_error", "attempt_id dan alasan wajib diisi.")
	}
	a, err := s.repo.GetAttempt(ctx, attemptID)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Attempt tidak ditemukan.")
	}
	if a.JadwalID != jadwalID {
		return nil, platform.ErrBadRequest("validation_error", "Attempt bukan dari jadwal ini.")
	}
	if a.Status != "diblokir" {
		return nil, platform.ErrConflict("conflict", "Hanya attempt yang diblokir yang bisa di-unblock (status: "+a.Status+").")
	}
	j, err := s.repo.GetJadwal(ctx, jadwalID)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Jadwal tidak ditemukan.")
	}
	// Sisa waktu untuk dilanjutkan: sisa beku bila ada, bila tidak
	// sisa = jadwal_selesai - now. Bila sudah lewat -> tidak bisa unblock.
	sisa := 0
	if a.SisaWaktuBeku != nil {
		sisa = *a.SisaWaktuBeku
	} else {
		selesai, err := gabungTanggalJam(j.Tanggal, j.JamSelesai)
		if err != nil {
			return nil, platform.ErrInternal("internal_error", "Tanggal/jam jadwal tidak valid.")
		}
		sisa = int(time.Until(selesai).Seconds())
	}
	if sisa <= 0 {
		return nil, platform.ErrConflict("conflict", "Ujian sudah berakhir (tidak ada sisa waktu). Unblock tidak bisa dilakukan.")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal memulai unblock.")
	}
	defer tx.Rollback(ctx)
	// LANJUT sisa waktu: jawaban tersimpan, TIDAK reset.
	if err := s.repo.LanjutkanAttempt(ctx, tx, attemptID, sisa); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal unblock attempt.")
	}
	// Tutup usulan yang menunggu (bila ada); bila tidak ada usulan
	// (proktor unblock langsung), buat baris audit usulan terproses.
	if usulanID, err := s.repo.UsulanMenungguTerbaru(ctx, tx, attemptID); err == nil {
		s.repo.SetujuiUsulan(ctx, tx, usulanID, actor.UserID)
	} else {
		s.repo.InsertUsulanDisetujui(ctx, tx, attemptID, actor.UserID, alasan)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal unblock attempt.")
	}
	s.catat(ctx, actor, "jadwal.unblock", "attempt", attemptID,
		map[string]any{"jadwal_id": jadwalID, "alasan": alasan, "sisa_waktu_detik": sisa})
	baru, _ := s.repo.GetAttempt(ctx, attemptID)
	s.broadcastAttempt(ctx, baru)
	return map[string]any{
		"attempt_id": attemptID, "status": "berjalan",
		"sisa_waktu_detik": sisa, "deadline": baru.Deadline,
	}, nil
}
