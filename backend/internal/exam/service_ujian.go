package exam

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"cbt-smk-v2/backend/internal/grading"
	"cbt-smk-v2/backend/internal/monitor"
	"cbt-smk-v2/backend/internal/platform"
)

// cekMilik menolak bila attempt bukan milik siswa (404, sama seperti
// milikSendiri dulu).
func cekMilik(a attempt, siswaID string) *platform.AppError {
	if a.SiswaID != siswaID {
		return platform.ErrNotFound("not_found", "Attempt tidak ditemukan.")
	}
	return nil
}

// UjianAktif mengembalikan jadwal hari ini untuk kelas siswa.
func (s *Service) UjianAktif(ctx context.Context, actor RequestActor, siswaID string) (any, *platform.AppError) {
	kelasID, ada := s.repo.KelasSiswa(ctx, siswaID)
	if !ada {
		return map[string]any{
			"items": []any{}, "kelas_tidak_ditemukan": true,
			"pesan": "Kelas siswa tidak ditemukan. Hubungi admin/proktor.",
		}, nil
	}
	items, err := s.repo.ListUjianAktif(ctx, siswaID, kelasID)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca jadwal aktif.")
	}
	return map[string]any{"items": items}, nil
}

// MulaiUjian memulai attempt baru (atau mengembalikan attempt berjalan
// yang sudah ada — idempotent). created=true -> respons 201.
func (s *Service) MulaiUjian(ctx context.Context, actor RequestActor, siswaID, jadwalID, token string) (data map[string]any, created bool, appErr *platform.AppError) {
	j, err := s.repo.GetJadwal(ctx, jadwalID)
	if err != nil {
		return nil, false, platform.ErrNotFound("not_found", "Jadwal tidak ditemukan.")
	}
	// Siswa hanya boleh ikut jadwal kelasnya.
	if kelasID, ada := s.repo.KelasSiswa(ctx, siswaID); ada && kelasID != j.KelasID {
		return nil, false, platform.ErrForbidden("forbidden", "Jadwal ini bukan untuk kelas Anda.")
	}
	if j.Status == "selesai" {
		return nil, false, platform.ErrConflict("conflict", "Jadwal sudah selesai.")
	}
	if j.Status == "pause" {
		return nil, false, platform.ErrConflict("conflict", "Ujian sedang dijeda oleh proktor. Tunggu hingga dilanjutkan.")
	}
	// Attempt berjalan yang sudah ada -> kembalikan (resume, idempotent).
	if lama, err := s.repo.FindAttemptByStatus(ctx, jadwalID, siswaID, "berjalan"); err == nil {
		return map[string]any{
			"attempt_id": lama.ID, "jadwal_id": lama.JadwalID, "status": lama.Status,
			"deadline": lama.Deadline, "sudah_ada": true,
			"sisa_waktu_detik": sisaWaktu(lama, j.Status, time.Now()),
		}, false, nil
	}
	if _, err := s.repo.FindAttemptByStatus(ctx, jadwalID, siswaID, "diblokir"); err == nil {
		return nil, false, platform.ErrForbidden("forbidden", "Attempt Anda diblokir. Hubungi proktor untuk unblock.")
	}
	if _, err := s.repo.FindAttemptByStatus(ctx, jadwalID, siswaID, "selesai"); err == nil {
		return nil, false, platform.ErrConflict("conflict", "Anda sudah menyelesaikan ujian ini.")
	}

	mulai, err := gabungTanggalJam(j.Tanggal, j.JamMulai)
	if err != nil {
		return nil, false, platform.ErrInternal("internal_error", "Tanggal/jam jadwal tidak valid.")
	}
	selesai, err := gabungTanggalJam(j.Tanggal, j.JamSelesai)
	if err != nil {
		return nil, false, platform.ErrInternal("internal_error", "Tanggal/jam jadwal tidak valid.")
	}
	now := time.Now()
	if now.Before(mulai) {
		return nil, false, platform.ErrConflict("conflict", "Ujian belum dimulai.")
	}
	if now.After(selesai) {
		return nil, false, platform.ErrConflict("conflict", "Ujian sudah berakhir.")
	}
	// Toleransi telat: hanya untuk MEMULAI (attempt baru).
	if now.After(mulai.Add(time.Duration(s.repo.ToleransiTelatMenit(ctx)) * time.Minute)) {
		return nil, false, platform.ErrForbidden("forbidden", "Batas toleransi keterlambatan habis. Anda tidak bisa memulai ujian ini.")
	}
	// Token (kecuali Tryout).
	if j.PaketJenis != "Tryout" {
		t := strings.TrimSpace(token)
		if t == "" || j.TokenHash == nil {
			return nil, false, platform.ErrUnauthorized("token_invalid", "Token ujian wajib diisi.")
		}
		if hashToken(t) != *j.TokenHash {
			return nil, false, platform.ErrUnauthorized("token_invalid", "Token ujian tidak cocok.")
		}
		if j.TokenExpiresAt != nil {
			if exp, ok := (*j.TokenExpiresAt).(time.Time); ok && now.After(exp) {
				return nil, false, platform.ErrConflict("token_expired", "Token ujian sudah kedaluwarsa.")
			}
		}
	}

	// Susun soal paket (urutan tampil per attempt).
	daftar, err := s.repo.ListSoalPaket(ctx, j.PaketID)
	if err != nil {
		return nil, false, platform.ErrInternal("internal_error", "Gagal membaca soal paket.")
	}
	soalIDs := []string{}
	tampilanPerSoal := map[string][]byte{}
	tipePerSoal := map[string]string{}
	for _, sd := range daftar {
		soalIDs = append(soalIDs, sd.ID)
		tampilanPerSoal[sd.ID] = sd.Jawaban
		tipePerSoal[sd.ID] = sd.Tipe
	}
	if len(soalIDs) == 0 {
		return nil, false, platform.ErrConflict("conflict", "Paket ujian belum berisi soal.")
	}

	attemptID := newUUID()
	seed := seedDariAttemptID(attemptID)
	urutanSoal := soalIDs
	if j.AcakSoal {
		urutanSoal = shuffleSeed(soalIDs, seed)
	}
	// urutan_opsi menyimpan TAMPILAN soal per attempt (soal.jawaban yang
	// sudah dikocok bila acak_jawaban aktif). Bila acak mati, tampilan
	// asli disimpan apa adanya agar GET attempt stabil.
	urutanOpsi := map[string]json.RawMessage{}
	for i, soalID := range urutanSoal {
		if j.AcakJawaban {
			urutanOpsi[soalID] = shuffleTampilan(tampilanPerSoal[soalID], tipePerSoal[soalID], seed+int64(i+1)*1009)
		} else {
			urutanOpsi[soalID] = tampilanPerSoal[soalID]
		}
	}
	urutanSoalJSON, _ := json.Marshal(urutanSoal)
	urutanOpsiJSON, _ := json.Marshal(urutanOpsi)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, false, platform.ErrInternal("internal_error", "Gagal memulai ujian.")
	}
	defer tx.Rollback(ctx)
	// Jadwal draft -> aktif saat peserta pertama mulai (sesuai tugas:
	// validasi "jadwal aktif/draft→aktif").
	if j.Status == "draft" {
		if err := s.repo.DraftKeAktif(ctx, tx, j.ID); err != nil {
			return nil, false, platform.ErrInternal("internal_error", "Gagal mengaktifkan jadwal.")
		}
	}
	// deadline = jadwal_selesai (ujian berakhir jam selesai jadwal).
	if err := s.repo.InsertAttempt(ctx, tx, attemptID, j.ID, siswaID, selesai, urutanSoalJSON, urutanOpsiJSON); err != nil {
		return nil, false, platform.ErrInternal("internal_error", "Gagal membuat attempt ujian.")
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, false, platform.ErrInternal("internal_error", "Gagal memulai ujian.")
	}
	s.catat(ctx, actor, "ujian.mulai", "attempt", attemptID,
		map[string]any{"jadwal_id": j.ID})
	monitor.Broadcast(j.ID, map[string]any{"type": "ujian_mulai", "jadwal_id": j.ID, "attempt_id": attemptID})
	return map[string]any{
		"attempt_id": attemptID, "jadwal_id": j.ID, "status": "berjalan",
		"deadline":         selesai.Format(time.RFC3339),
		"sisa_waktu_detik": int(time.Until(selesai).Seconds()),
		"jumlah_soal":      len(soalIDs),
	}, true, nil
}

// muatSoalAttempt mengembalikan soal paket untuk attempt, TANPA kunci &
// pembahasan, dalam urutan attempt (urutan_soal) + opsi dalam urutan
// attempt (urutan_opsi). Disertai jawaban tersimpan + ragu.
func (s *Service) muatSoalAttempt(ctx context.Context, a attempt, j jadwal) ([]map[string]any, error) {
	daftar, err := s.repo.ListSoalDetail(ctx, j.PaketID)
	if err != nil {
		return nil, err
	}
	byID := map[string]soalDetailRow{}
	var semua []string
	for _, sd := range daftar {
		byID[sd.ID] = sd
		semua = append(semua, sd.ID)
	}
	// Urutan tampil = urutan_soal attempt; soal baru (di luar urutan)
	// ditambahkan di akhir.
	urutan := []string{}
	terlihat := map[string]bool{}
	for _, id := range a.UrutanSoal {
		if _, ada := byID[id]; ada {
			urutan = append(urutan, id)
			terlihat[id] = true
		}
	}
	for _, id := range semua {
		if !terlihat[id] {
			urutan = append(urutan, id)
		}
	}
	// Jawaban tersimpan.
	tersimpan, err := s.repo.ListJawabanTersimpan(ctx, a.ID)
	if err != nil {
		return nil, err
	}
	jawabanMap := map[string]json.RawMessage{}
	raguMap := map[string]bool{}
	for _, t := range tersimpan {
		jawabanMap[t.SoalID] = t.Jawaban
		raguMap[t.SoalID] = t.Ragu
	}
	out := []map[string]any{}
	for i, id := range urutan {
		sd := byID[id]
		// Tampilan = soal.jawaban (data tampil Tim 3) dalam urutan
		// attempt; bila tidak tersimpan (attempt lama), pakai asli.
		// KUNCI TIDAK PERNAH disertakan (kolom kunci tidak dipilih).
		tampilan := sd.Tampilan
		if ts, ada := a.UrutanOpsi[id]; ada && len(ts) > 0 {
			tampilan = ts
		}
		var tampilanAny any
		if len(tampilan) > 0 {
			_ = json.Unmarshal(tampilan, &tampilanAny)
		}
		item := map[string]any{
			"soal_id": sd.ID, "nomor": i + 1, "tipe": sd.Tipe,
			"pertanyaan": sd.Pertanyaan, "stimulus": sd.Stimulus,
			"media_url": sd.MediaURL,
			"tampilan":  tampilanAny, "bobot": sd.Bobot,
			"jawaban": jawabanMap[id], "ragu": raguMap[id],
		}
		out = append(out, item)
	}
	return out, nil
}

// GetAttemptDetail mengembalikan detail attempt + soal untuk siswa.
func (s *Service) GetAttemptDetail(ctx context.Context, actor RequestActor, siswaID, attemptID string) (map[string]any, *platform.AppError) {
	a, err := s.repo.GetAttempt(ctx, attemptID)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Attempt tidak ditemukan.")
	}
	if appErr := cekMilik(a, siswaID); appErr != nil {
		return nil, appErr
	}
	a, _ = s.autoSubmitJikaLewat(ctx, a, s.repo.JadwalStatus(ctx, a.JadwalID))
	j, err := s.repo.GetJadwal(ctx, a.JadwalID)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Jadwal tidak ditemukan.")
	}
	soal, err := s.muatSoalAttempt(ctx, a, j)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca soal ujian.")
	}
	resp := map[string]any{
		"attempt_id": a.ID, "jadwal_id": a.JadwalID, "status": a.Status,
		"deadline": a.Deadline, "mulai_at": a.MulaiAt,
		"sisa_waktu_detik": sisaWaktu(a, j.Status, time.Now()),
		"jadwal_status":    j.Status, "paket_jenis": j.PaketJenis,
		"server_now": time.Now().Format(time.RFC3339),
		"soal":       soal,
	}
	if a.Status == "selesai" && s.repo.TampilHasilEfektif(ctx, j.TampilHasil) {
		nilai, menunggu, err := s.hitungNilai(ctx, a.ID)
		if err == nil {
			resp["nilai_akhir"] = nilai
			resp["menunggu_koreksi"] = menunggu
		}
	}
	return resp, nil
}

// SimpanJawaban menyimpan satu jawaban siswa (dengan penskoran otomatis).
func (s *Service) SimpanJawaban(ctx context.Context, actor RequestActor, siswaID, attemptID, soalID string, jawaban json.RawMessage) (map[string]any, *platform.AppError) {
	a, err := s.repo.GetAttempt(ctx, attemptID)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Attempt tidak ditemukan.")
	}
	if appErr := cekMilik(a, siswaID); appErr != nil {
		return nil, appErr
	}
	if a.Status == "diblokir" {
		return nil, platform.ErrForbidden("forbidden", "Attempt diblokir. Hubungi proktor.")
	}
	if a.Status == "selesai" {
		return nil, platform.ErrConflict("conflict", "Ujian sudah selesai (submit). Jawaban tidak bisa diubah.")
	}
	a, _ = s.autoSubmitJikaLewat(ctx, a, s.repo.JadwalStatus(ctx, a.JadwalID))
	if a.Status != "berjalan" {
		return nil, platform.ErrConflict("conflict", "Batas waktu ujian habis. Jawaban tidak bisa disimpan.")
	}
	j, err := s.repo.GetJadwal(ctx, a.JadwalID)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Jadwal tidak ditemukan.")
	}
	if j.Status == "pause" {
		return nil, platform.ErrConflict("conflict", "Ujian sedang dijeda. Jawaban tidak bisa disimpan sampai dilanjutkan.")
	}
	// Soal harus bagian dari paket jadwal ini.
	soalAda, _ := s.repo.PaketSoalExists(ctx, j.PaketID, soalID)
	if !soalAda {
		return nil, platform.ErrBadRequest("validation_error", "Soal bukan bagian dari paket ujian ini.")
	}
	// Ambil tipe + kunci + bobot + tampilan soal untuk penskoran otomatis.
	// (Tampilan dibutuhkan grading urutkan bila jawaban berupa string item.)
	tipe, kunci, bobotRaw, tampilan, err := s.repo.SoalUntukSkor(ctx, soalID)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca soal.")
	}
	var bobot int
	if f, err := strconv.ParseFloat(strings.TrimSpace(string(bobotRaw)), 64); err == nil {
		bobot = int(f)
	}
	raguLama := s.repo.RaguLama(ctx, a.ID, soalID)
	// Penskoran: otomatis kecuali uraian/manual.
	var poin *float64
	dikoreksi := false
	kanon := grading.NormalisasiTipe(tipe)
	if kanon == "uraian" {
		poin = nil // manual: menunggu koreksi guru.
	} else if !grading.Dinilai(tipe, kunci) {
		nol := 0.0
		poin = &nol
		dikoreksi = true // tidak dinilai -> tidak menunggu koreksi.
	} else {
		p, auto := grading.HitungPoin(tipe, jawaban, kunciUntukGrading(kunci, tampilan, tipe), bobot)
		if !auto {
			poin = nil
		} else {
			poin = &p
			dikoreksi = true // auto sudah final per soal.
		}
	}
	var poinArg any
	if poin != nil {
		poinArg = *poin
	}
	if err := s.repo.UpsertJawaban(ctx, a.ID, soalID, jawaban, raguLama, poinArg, dikoreksi); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menyimpan jawaban.")
	}
	// Audit (Fase 3): simpan jawaban adalah mutasi — sebelumnya tidak
	// dicatat. Detail TIDAK memuat isi jawaban (hemat audit log; isi
	// jawaban sudah tersimpan di tabel jawaban).
	s.catat(ctx, actor, "ujian.jawaban", "jawaban", soalID,
		map[string]any{"attempt_id": a.ID, "soal_id": soalID})
	s.broadcastAttempt(ctx, a)
	return map[string]any{
		"tersimpan": true, "soal_id": soalID,
		"server_now": time.Now().Format(time.RFC3339),
	}, nil
}

// TandaiRagu menandai/membatalkan flag ragu-ragu satu soal.
func (s *Service) TandaiRagu(ctx context.Context, actor RequestActor, siswaID, attemptID, soalID string, ragu bool) (map[string]any, *platform.AppError) {
	a, err := s.repo.GetAttempt(ctx, attemptID)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Attempt tidak ditemukan.")
	}
	if appErr := cekMilik(a, siswaID); appErr != nil {
		return nil, appErr
	}
	if a.Status != "berjalan" {
		return nil, platform.ErrConflict("conflict", "Attempt tidak sedang berjalan.")
	}
	a, _ = s.autoSubmitJikaLewat(ctx, a, s.repo.JadwalStatus(ctx, a.JadwalID))
	if a.Status != "berjalan" {
		return nil, platform.ErrConflict("conflict", "Batas waktu ujian habis.")
	}
	if err := s.repo.UpsertRagu(ctx, a.ID, soalID, ragu); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menandai ragu-ragu.")
	}
	// Audit (Fase 3): tandai ragu-ragu adalah mutasi — sebelumnya tidak dicatat.
	s.catat(ctx, actor, "ujian.ragu", "jawaban", soalID,
		map[string]any{"attempt_id": a.ID, "soal_id": soalID, "ragu": ragu})
	s.broadcastAttempt(ctx, a)
	return map[string]any{"soal_id": soalID, "ragu": ragu}, nil
}

// CatatPelanggaran mencatat pelanggaran anti-curang (1x peringatan,
// 2x auto-block per attempt; deduplikasi 5 detik).
func (s *Service) CatatPelanggaran(ctx context.Context, actor RequestActor, siswaID, attemptID, jenis, detail string) (map[string]any, *platform.AppError) {
	a, err := s.repo.GetAttempt(ctx, attemptID)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Attempt tidak ditemukan.")
	}
	if appErr := cekMilik(a, siswaID); appErr != nil {
		return nil, appErr
	}
	if a.Status == "selesai" {
		return nil, platform.ErrConflict("conflict", "Ujian sudah selesai.")
	}
	if a.Status == "diblokir" {
		return map[string]any{"diblokir": true, "pesan": "Attempt sudah diblokir. Hubungi proktor."}, nil
	}
	// Deduplikasi: abaikan event berulang dalam 5 detik (blur +
	// visibilitychange sering berapi hampir bersamaan).
	terakhir, jumlah := s.repo.PelanggaranTerakhir(ctx, a.ID)
	if terakhir != nil && time.Since(*terakhir) < 5*time.Second {
		return map[string]any{
			"diabaikan_duplikat": true, "jumlah_pelanggaran": jumlah,
			"peringatan": jumlah == 1, "diblokir": false,
		}, nil
	}
	if err := s.repo.InsertPelanggaran(ctx, a.ID, jenis, detail); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal mencatat pelanggaran.")
	}
	jumlah++
	s.catat(ctx, actor, "ujian.pelanggaran", "attempt", a.ID,
		map[string]any{"jenis": jenis, "jumlah": jumlah})
	batas := s.repo.MaxPelanggaran(ctx)
	if jumlah >= batas {
		// Auto-block PER-ATTEMPT. Bekukan sisa waktu agar unblock bisa
		// melanjutkan dari sisa yang sama.
		_ = s.repo.BlokirAttempt(ctx, a.ID)
		a.Status = "diblokir"
		s.broadcastAttempt(ctx, a)
		return map[string]any{
			"diblokir": true, "jumlah_pelanggaran": jumlah,
			"pesan": "Anda diblokir karena pelanggaran berulang. Hubungi proktor.",
		}, nil
	}
	s.broadcastAttempt(ctx, a)
	return map[string]any{
		"peringatan": true, "diblokir": false, "jumlah_pelanggaran": jumlah,
		"pesan": "Peringatan: terdeteksi meninggalkan halaman ujian. Pelanggaran berikutnya akan memblokir ujian Anda.",
	}, nil
}

// SubmitUjian memfinalisasi attempt (idempotent bila sudah selesai).
func (s *Service) SubmitUjian(ctx context.Context, actor RequestActor, siswaID, attemptID string) (map[string]any, *platform.AppError) {
	a, err := s.repo.GetAttempt(ctx, attemptID)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Attempt tidak ditemukan.")
	}
	if appErr := cekMilik(a, siswaID); appErr != nil {
		return nil, appErr
	}
	if a.Status == "diblokir" {
		return nil, platform.ErrForbidden("forbidden", "Attempt diblokir. Hubungi proktor.")
	}
	if a.Status == "selesai" {
		// Idempotent: kembalikan hasil tersimpan.
		j, _ := s.repo.GetJadwal(ctx, a.JadwalID)
		resp := map[string]any{"attempt_id": a.ID, "status": "selesai", "submit_at": a.SubmitAt}
		if s.repo.TampilHasilEfektif(ctx, j.TampilHasil) && a.NilaiAkhir != nil {
			resp["nilai_akhir"] = *a.NilaiAkhir
		}
		return resp, nil
	}
	nilai, menunggu, err := s.finalisasiAttempt(ctx, a)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal submit ujian.")
	}
	j, err := s.repo.GetJadwal(ctx, a.JadwalID)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Jadwal tidak ditemukan.")
	}
	s.catat(ctx, actor, "ujian.submit", "attempt", a.ID,
		map[string]any{"jadwal_id": a.JadwalID, "nilai_akhir": nilai, "menunggu_koreksi": menunggu})
	s.broadcastAttempt(ctx, a)
	resp := map[string]any{
		"attempt_id": a.ID, "status": "selesai", "menunggu_koreksi": menunggu,
	}
	// Nilai disembunyikan hingga final bila tampil_hasil false.
	if s.repo.TampilHasilEfektif(ctx, j.TampilHasil) {
		resp["nilai_akhir"] = nilai
	} else {
		resp["nilai_akhir"] = nil
		resp["pesan"] = "Nilai disembunyikan sampai koreksi final."
	}
	return resp, nil
}

// StatusAttempt mengembalikan status ringan attempt untuk polling client.
func (s *Service) StatusAttempt(ctx context.Context, actor RequestActor, siswaID, attemptID string) (map[string]any, *platform.AppError) {
	a, err := s.repo.GetAttempt(ctx, attemptID)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Attempt tidak ditemukan.")
	}
	if appErr := cekMilik(a, siswaID); appErr != nil {
		return nil, appErr
	}
	a, _ = s.autoSubmitJikaLewat(ctx, a, s.repo.JadwalStatus(ctx, a.JadwalID))
	j, err := s.repo.GetJadwal(ctx, a.JadwalID)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Jadwal tidak ditemukan.")
	}
	dijawab, ragu, total := s.repo.AttemptRingkasan(ctx, a.ID, j.PaketID)
	resp := map[string]any{
		"attempt_id": a.ID, "jadwal_id": a.JadwalID, "status": a.Status,
		"deadline": a.Deadline, "jadwal_status": j.Status,
		"sisa_waktu_detik": sisaWaktu(a, j.Status, time.Now()),
		"dijawab":          dijawab, "ragu": ragu, "total_soal": total,
		"belum_dijawab": total - dijawab,
		"server_now":    time.Now().Format(time.RFC3339),
	}
	if a.Status == "selesai" {
		if s.repo.TampilHasilEfektif(ctx, j.TampilHasil) && a.NilaiAkhir != nil {
			resp["nilai_akhir"] = *a.NilaiAkhir
		} else {
			resp["nilai_akhir"] = nil
		}
		_, menunggu, err := s.hitungNilai(ctx, a.ID)
		if err == nil {
			resp["menunggu_koreksi"] = menunggu
		}
	}
	return resp, nil
}

// hasilItem adalah satu baris rekap per-soal halaman hasil siswa.
type hasilItem struct {
	SoalID      string   `json:"soal_id"`
	Nomor       int      `json:"nomor"`
	Tipe        string   `json:"tipe"`
	Pertanyaan  string   `json:"pertanyaan"`
	Jawaban     any      `json:"jawaban_siswa"`
	Poin        *float64 `json:"poin"`
	Bobot       float64  `json:"bobot"`
	Ragu        bool     `json:"ragu"`
	Status      string   `json:"status"`
	KunciTampil string   `json:"kunci_tampil,omitempty"`
	// Pembahasan hanya diisi bila tampil (Tryout / tampil_hasil
	// efektif); omitempty menyembunyikannya bila kosong.
	Pembahasan string `json:"pembahasan,omitempty"`
}

// HasilAttempt mengembalikan rekap per-soal untuk halaman hasil siswa.
// Hanya attempt milik sendiri & sudah selesai. Pembahasan + kunci tampil
// hanya disertakan bila paket Tryout atau tampil_hasil efektif.
func (s *Service) HasilAttempt(ctx context.Context, actor RequestActor, siswaID, attemptID string) (map[string]any, *platform.AppError) {
	a, err := s.repo.GetAttempt(ctx, attemptID)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Attempt tidak ditemukan.")
	}
	if appErr := cekMilik(a, siswaID); appErr != nil {
		return nil, appErr
	}
	if a.Status != "selesai" {
		return nil, platform.ErrConflict("conflict", "Ujian belum selesai. Hasil tersedia setelah submit.")
	}
	j, err := s.repo.GetJadwal(ctx, a.JadwalID)
	if err != nil {
		return nil, platform.ErrNotFound("not_found", "Jadwal tidak ditemukan.")
	}
	tampil := s.repo.TampilHasilEfektif(ctx, j.TampilHasil) || j.PaketJenis == "Tryout"
	baris, err := s.repo.ListHasilJawaban(ctx, a.ID)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal membaca hasil ujian.")
	}
	bySoal := map[string]*hasilItem{}
	for _, h := range baris {
		it := &hasilItem{
			SoalID: h.SoalID, Ragu: h.Ragu, Tipe: h.Tipe,
			Pertanyaan: h.Pertanyaan, Bobot: h.Bobot,
		}
		it.Poin = h.Poin
		var jwbAny any
		if len(h.Jawaban) > 0 {
			_ = json.Unmarshal(h.Jawaban, &jwbAny)
		}
		it.Jawaban = jwbAny
		if tampil {
			it.KunciTampil = kunciTampil(h.Tipe, h.Kunci, h.Tampilan)
			it.Pembahasan = h.Pembahasan
			if it.Poin == nil {
				it.Status = "menunggu_koreksi"
			} else if *it.Poin >= it.Bobot {
				it.Status = "benar"
			} else if *it.Poin > 0 {
				it.Status = "sebagian"
			} else {
				it.Status = "salah"
			}
		} else {
			it.Poin = nil
			it.Status = "disembunyikan"
		}
		bySoal[h.SoalID] = it
	}
	// Urutkan sesuai urutan_soal attempt; soal tanpa jawaban ikut (kosong).
	items := []hasilItem{}
	terlihat := map[string]bool{}
	for i, soalID := range a.UrutanSoal {
		if it, ada := bySoal[soalID]; ada {
			it.Nomor = i + 1
			items = append(items, *it)
			terlihat[soalID] = true
		}
	}
	for soalID, it := range bySoal {
		if !terlihat[soalID] {
			it.Nomor = len(items) + 1
			items = append(items, *it)
		}
	}
	// Soal paket tanpa baris jawaban (tidak dijawab sama sekali):
	// lengkapi soal yang belum ada di jawaban (poin 0, jawaban null).
	sudah := map[string]bool{}
	for _, it := range items {
		sudah[it.SoalID] = true
	}
	if soalPaket, err := s.repo.ListSoalHasilPaket(ctx, j.PaketID); err == nil {
		nomor := len(items)
		for _, sp := range soalPaket {
			if sudah[sp.ID] {
				continue
			}
			nomor++
			it := hasilItem{SoalID: sp.ID, Nomor: nomor, Tipe: sp.Tipe, Pertanyaan: sp.Pertanyaan,
				Bobot: sp.Bobot, Status: "disembunyikan"}
			if tampil {
				it.Status = "salah"
				nol := 0.0
				it.Poin = &nol
				it.Pembahasan = sp.Pembahasan
			}
			items = append(items, it)
		}
	}
	_, menunggu, _ := s.hitungNilai(ctx, a.ID)
	resp := map[string]any{
		"attempt_id": a.ID, "jadwal_id": a.JadwalID, "paket_jenis": j.PaketJenis,
		"tampil_hasil": tampil, "menunggu_koreksi": menunggu, "items": items,
	}
	if tampil && a.NilaiAkhir != nil {
		resp["nilai_akhir"] = *a.NilaiAkhir
	} else {
		resp["nilai_akhir"] = nil
	}
	return resp, nil
}
