package master

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"cbt-smk-v2/backend/internal/platform"

	"github.com/jackc/pgx/v5"
)

// ImportSiswa memproses file CSV (multipart yang sudah dibuka handler)
// menjadi baris siswa. Baris yang gagal dilewati + dilaporkan per baris;
// baris yang valid disimpan dalam satu transaksi.
//
// IDEMPOTEN (Fase 3, BUG-02): NISN yang sudah ada di DB dihitung sebagai
// `sudah_ada` (bukan error, bukan duplikat, tidak buat user baru).
func (s *Service) ImportSiswa(ctx context.Context, actor RequestActor, file multipart.File) (*importResult, *platform.AppError) {
	reader := csv.NewReader(io.LimitReader(file, (maxImportFileMB<<20)+1))
	reader.FieldsPerRecord = -1 // baris pendek/panjang ditangani manual per kolom
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, platform.ErrBadRequest("validation_error", "File bukan CSV yang valid.")
	}
	if len(records) == 0 {
		return nil, platform.ErrBadRequest("validation_error", "File CSV kosong.")
	}
	// Normalisasi header: huruf kecil, tanpa spasi, tanpa BOM.
	header := make(map[string]int, len(records[0]))
	for i, h := range records[0] {
		h = strings.ToLower(strings.TrimSpace(h))
		h = strings.TrimPrefix(h, "\ufeff")
		header[h] = i
	}
	col := func(names ...string) (int, bool) {
		for _, n := range names {
			if idx, ok := header[n]; ok {
				return idx, true
			}
		}
		return -1, false
	}
	namaCol, okNama := col("nama", "nama_lengkap", "nama_siswa")
	nisnCol, okNISN := col("nisn")
	nisCol, _ := col("nis")
	kelasCol, okKelas := col("kelas_id", "kelas_kode", "kelas", "kode_kelas")
	if !okNama || !okNISN || !okKelas {
		return nil, platform.ErrBadRequest("validation_error",
			"Header CSV tidak dikenali. Harus ada kolom: nama, nisn, dan kelas_id/kelas_kode (opsional: nis).")
	}
	dataRows := records[1:]
	if len(dataRows) > maxImportRows {
		return nil, platform.ErrBadRequest("validation_error",
			fmt.Sprintf("Maksimal %d baris per file import.", maxImportRows))
	}
	cell := func(row []string, idx int) string {
		if idx < 0 || idx >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[idx])
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal memulai import.")
	}
	defer tx.Rollback(ctx)
	result := &importResult{Errors: []importRowError{}}
	seenNISN := map[string]bool{}
	for i, row := range dataRows {
		baris := i + 2 // baris 1 = header
		nama := cell(row, namaCol)
		nisn := cell(row, nisnCol)
		nis := cell(row, nisCol)
		kelasVal := cell(row, kelasCol)
		fail := func(pesan string) {
			result.Gagal++
			result.Errors = append(result.Errors, importRowError{Baris: baris, Pesan: pesan})
		}
		if nama == "" {
			fail("Nama wajib diisi.")
			continue
		}
		if nisn == "" {
			fail("NISN wajib diisi.")
			continue
		}
		if kelasVal == "" {
			fail("Kelas (kelas_id/kelas_kode) wajib diisi.")
			continue
		}
		if seenNISN[nisn] {
			fail("NISN duplikat di dalam file.")
			continue
		}
		seenNISN[nisn] = true
		exists, err := s.repo.SiswaNISNExistsTx(ctx, tx, nisn)
		if err != nil {
			return nil, platform.ErrInternal("internal_error", "Gagal memeriksa NISN.")
		}
		if exists {
			// Idempotent: NISN sudah ada -> lewati tanpa error,
			// tanpa membuat user baru, tanpa baris errors.
			result.SudahAda++
			continue
		}
		// Resolusi kelas: UUID -> id; selain itu -> nama kelas (tidak peka huruf).
		kelasID, jurusanID, err := s.repo.LookupKelasTx(ctx, tx, kelasVal)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				fail(fmt.Sprintf("Kelas %q tidak ditemukan.", kelasVal))
				continue
			}
			return nil, platform.ErrInternal("internal_error", "Gagal mencari kelas.")
		}
		userID, err := s.createStudentAccountTx(ctx, tx, nisn, nama)
		if err != nil {
			return nil, platform.ErrInternal("internal_error", "Gagal membuat akun siswa.")
		}
		var nisArg any
		if nis != "" {
			nisArg = nis
		}
		if err := s.repo.InsertSiswaImportTx(ctx, tx, userID, nisn, nisArg, nama, kelasID, jurusanID); err != nil {
			if constraint, ok := uniqueConstraint(err); ok {
				// Race dengan import paralel: NISN ternyata sudah
				// masuk di antara cek EXISTS dan INSERT -> idempotent,
				// hitung sudah_ada (jalur utama adalah cek EXISTS di
				// atas; jalur ini hanya jaring pengaman race).
				if constraint == "siswa_nisn_key" || constraint == "users_username_key" {
					result.SudahAda++
					continue
				}
				fail(conflictMessage(constraint))
				continue
			}
			return nil, platform.ErrInternal("internal_error", "Gagal menyimpan siswa.")
		}
		result.Sukses++
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, platform.ErrInternal("internal_error", "Gagal menyelesaikan import.")
	}
	s.catat(ctx, actor, "import_siswa", "siswa", "", map[string]any{"sukses": result.Sukses, "gagal": result.Gagal, "sudah_ada": result.SudahAda})
	return result, nil
}
