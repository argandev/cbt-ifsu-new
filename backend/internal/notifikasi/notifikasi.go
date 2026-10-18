// Package notifikasi — notifikasi dalam aplikasi (Tim Fase 3).
//
// Tabel `notifikasi` (migrasi 0012): id, penerima_user_id, judul, isi,
// dibaca, created_at. Endpoint (Register):
//
//	GET  /api/v1/notifikasi                 (notifikasi milik sendiri, pagination)
//	POST /api/v1/notifikasi/:id/baca        (tandai dibaca; bukan milik -> 404)
//
// Helper Buat dipakai paket lain (mis. sod: notifikasi pengecualian SoD
// untuk semua admin kecuali actor). Notifikasi hanya dibaca pemiliknya
// (anti-IDOR, KONTRAK §5.3).
//
// Struktur layer (hasil refactor arsitektur):
//   - notifikasi.go: API publik (Buat — signature DIPERTAHANKAN, dipakai
//     paket lain) + dokumentasi paket
//   - handler.go:    Echo handler + Register
//   - service.go:    logika bisnis + audit log
//   - repository.go: semua query SQL
//   - platform:      envelope JSON (menggantikan helper lokal)
//
// Perilaku (route, status, bentuk JSON, pesan error) IDENTIK dengan
// sebelum refactor.
package notifikasi

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Buat menyisipkan satu notifikasi untuk penerima. Gagal buat TIDAK
// boleh menggagalkan operasi pemanggil (pemanggil mengabaikan error).
// Signature DIPERTAHANKAN — dipakai paket lain (mis. sod).
func Buat(ctx context.Context, pool *pgxpool.Pool, penerimaUserID, judul, isi string) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO notifikasi (penerima_user_id, judul, isi) VALUES ($1, $2, $3)`,
		penerimaUserID, judul, isi)
	return err
}
