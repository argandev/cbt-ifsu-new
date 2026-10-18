// Package webdist men-embed hasil build frontend (backend/dist).
//
// `make build` (root) menyalin frontend/dist -> backend/dist sebelum
// `go build`, sehingga satu binary Go menyajikan SPA + API.
// Bila backend/dist hanya berisi placeholder (belum pernah `make build`),
// server tetap jalan dan menampilkan halaman placeholder.
package webdist

import "embed"

// Dist berisi seluruh isi backend/dist.
var (
	//go:embed dist
	Dist embed.FS
)
