// Package migrations men-embed berkas migrasi SQL agar golang-migrate
// dapat menjalankannya dari dalam binary (auto-up on boot).
//
// Berkas di sini milik tim masing-masing (0001 = Tim 1, 0003-0005 = Tim 3,
// dst). Jangan ubah migrasi yang sudah diterapkan di environment manapun —
// buat migrasi baru berversi.
package migrations

import "embed"

// FS berisi seluruh berkas *.sql di direktori ini.
var (
	//go:embed *.sql
	FS embed.FS
)
