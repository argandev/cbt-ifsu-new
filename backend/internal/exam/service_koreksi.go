package exam

import (
	"context"

	"cbt-smk-v2/backend/internal/platform"
)

// ListKoreksi mengembalikan jawaban manual yang belum dikoreksi
// untuk satu jadwal.
func (s *Service) ListKoreksi(ctx context.Context, actor RequestActor, jadwalID string) (any, *platform.AppError) {
	if _, err := s.repo.GetJadwal(ctx, jadwalID); err != nil {
		return nil, platform.ErrNotFound("not_found", "Jadwal tidak ditemukan.")
	}
	items, err := s.repo.ListKoreksi(ctx, jadwalID)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca daftar koreksi.")
	}
	return map[string]any{"jadwal_id": jadwalID, "items": items}, nil
}

// KoreksiJawaban menyimpan skor koreksi manual (skor 0–100);
// poin = skor/100 × bobot. Nilai akhir dihitung ulang bila attempt
// sudah selesai.
func (s *Service) KoreksiJawaban(ctx context.Context, actor RequestActor, jawabanID string, skor float64) (map[string]any, *platform.AppError) {
	attemptID, soalID, bobot, err := s.repo.GetJawabanKoreksi(ctx, jawabanID)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Jawaban tidak ditemukan.")
	}
	// poin = skor/100 × bobot.
	poin := skor / 100 * bobot
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal memulai koreksi.")
	}
	defer tx.Rollback(ctx)
	if err := s.repo.UpdateJawabanPoin(ctx, tx, jawabanID, poin); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menyimpan koreksi.")
	}
	attemptStatus, _ := s.repo.AttemptStatusJadwal(ctx, tx, attemptID)
	var nilaiBaru *float64
	menunggu := true
	if attemptStatus == "selesai" {
		// Hitung ulang nilai akhir attempt.
		if err := tx.Commit(ctx); err != nil {
			return nil, platform.ErrInternal("internal_error", "Gagal menyimpan koreksi.")
		}
		nilai, m, err := s.hitungNilai(ctx, attemptID)
		if err != nil {
			return nil, platform.ErrInternal("internal_error", "Gagal menghitung ulang nilai.")
		}
		menunggu = m
		_ = s.repo.UpdateNilaiAkhir(ctx, attemptID, nilai)
		nilaiBaru = &nilai
	} else {
		if err := tx.Commit(ctx); err != nil {
			return nil, platform.ErrInternal("internal_error", "Gagal menyimpan koreksi.")
		}
		// Attempt masih berjalan: nilai akhir dihitung saat submit.
		_, menunggu, _ = s.hitungNilai(ctx, attemptID)
	}
	s.catat(ctx, actor, "koreksi.update", "jawaban", jawabanID,
		map[string]any{"attempt_id": attemptID, "soal_id": soalID, "skor": skor, "poin": poin})
	a, _ := s.repo.GetAttempt(ctx, attemptID)
	if a.ID != "" && nilaiBaru != nil {
		s.syncNilaiAkhir(ctx, a, *nilaiBaru)
	}
	s.broadcastAttempt(ctx, a)
	resp := map[string]any{
		"jawaban_id": jawabanID, "attempt_id": attemptID, "poin": poin,
		"menunggu_koreksi": menunggu,
	}
	if nilaiBaru != nil {
		resp["nilai_akhir_baru"] = *nilaiBaru
		resp["final"] = !menunggu
	}
	return resp, nil
}
