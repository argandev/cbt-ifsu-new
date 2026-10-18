# ARSITEKTUR Backend CBT SMK v2 (refactor)

Hasil refactor arsitektur: pola **layered** per paket — `handler → service → repository`.
Perilaku (endpoint, logika bisnis, skema DB, bentuk JSON) TIDAK berubah; yang berubah
hanya susunan kode.

## Diagram layer

```
┌─────────────────────────────────────────────────────────┐
│  cmd/server/main.go                                     │
│  (wiring: Register(e, pool, q, cfg) per paket — tetap)   │
└──────────────────────────┬──────────────────────────────┘
                           │ memanggil
┌──────────────────────────▼──────────────────────────────┐
│  internal/<paket>/handler*.go                           │
│  - Echo handler: func (h *Handler) Nama(c *echo.Context) │
│  - Tugas: parse input HTTP, panggil service, tulis JSON │
│  - TIDAK ada SQL di sini.                               │
│  - Response via platform.WriteOK / WriteError / dst.    │
└──────────────────────────┬──────────────────────────────┘
                           │ memanggil
┌──────────────────────────▼──────────────────────────────┐
│  internal/<paket>/service*.go                            │
│  - Logika bisnis + orkestrasi (dipindah ASLI dari kode   │
│    lama, bukan ditulis ulang).                           │
│  - Mengembalikan (hasil, *platform.AppError).           │
│  - Boleh memanggil service/repository paket lain ASALKAN │
│    tidak circular (lihat peta dependensi di bawah).      │
└──────────────────────────┬──────────────────────────────┘
                           │ memanggil
┌──────────────────────────▼──────────────────────────────┐
│  internal/<paket>/repository*.go                         │
│  - Akses DB via pgxpool (query dipindah ASLI, semantik   │
│    SQL identik).                                         │
│  - Mengembalikan (hasil, error) biasa.                   │
└─────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────┐
│  internal/platform/                                     │
│  - errors.go: AppError{Code, Message, HTTPStatus}        │
│  - httpx.go: WriteError/WriteOK/WriteStatus/WriteCreated │
│    (output JSON IDENTIK dengan helper lama tiap paket)   │
│  - pagination.go: 2 gaya paginasi (page & limit),       │
│    dipertahankan apa adanya per paket.                   │
└─────────────────────────────────────────────────────────┘
```

## Aturan refactor (keras)

1. **Endpoint identik**: route, method, path, status code, bentuk JSON response.
2. **Logika identik**: PINDAHKAN kode, jangan tulis ulang. Grading, auth,
   token, timer, anti-cheat, permission, semantik SQL tidak berubah.
3. **Skema DB tidak berubah**: `migrations/` dan `db/` tidak disentuh.
4. **Test hijau**: `go test ./...` harus lolos; asersi test tidak diubah.
5. **Tidak circular import**: arah dependensi antar-paket dipertahankan
   seperti sebelum refactor.

## Peta dependensi antar-paket (arah panah = "memakai")

```
config, rbac, session          (fondasi, tanpa dependensi)
audit → config
grading                        (murni, tanpa dependensi)
middleware → rbac, session
auth → audit, config, middleware, rbac, session
analisis → audit, config, grading
backup, banksoal, dashboard, event, master, rekap, settings,
notifikasi, ukk, upload → audit, config
monitor → config
exam → config, grading, monitor
sod → audit, config, notifikasi
users → audit, auth, config, session
analisisx, ukkx               (tanpa dependensi internal)
```

## Status refactor per paket

| Paket | Status | File baru |
|---|---|---|
| platform | ✅ baru | errors.go, httpx.go, pagination.go |
| users | ✅ referensi | handler.go, service.go, repository.go (+ perbaikan urutan 404-UUID sebelum bind) |
| auth, exam, sod | ✅ worker A | handler_*.go, service_*.go, repository_*.go per sub-domain |
| banksoal, master, event | ✅ worker B | 9 + 18 + 9 file (lihat laporan); PUT load-dulu dipertahankan via LoadXChecked/ApplyXUpdate |
| ukk, rekap, settings, backup, upload, dashboard, monitor, notifikasi, analisis, audit | ✅ worker C | handler.go/service.go/repository.go per paket |
| config, grading, middleware, rbac, session, analisisx, ukkx | ⏭️ tidak direfactor | (tanpa handler / sudah library murni) |

Legenda: ✅ selesai · 🔄 dikerjakan · ⬜ belum

## Aturan urutan cek (penting — ditemukan saat refactor)

Kode asal tidak konsisten soal urutan "load (404) vs bind (400)":
- Sebagian handler PUT memuat data DULU (404) baru bind body (400):
  `bank-soal/:id`, `bank-soal/:bankId/soal/:soalId`, `event/:id`,
  `ruangan/:id`, `jadwal/:id/event-ruangan`, `pengguna/:id` (cek format UUID).
- Sebagian bind DULU (400) baru load (404): `master/*`, `ukk/:id`,
  `paket/:id`, `jadwal/:id`, `ukk/:id/nilai/:nilaiId`.

Urutan ini MEMPENGARUHI status code untuk kasus tepi "body invalid + id tidak ada"
(404 vs 400), jadi HARUS dipertahankan apa adanya. Pola yang dipakai:
untuk kasus load-dulu, service dipecah jadi `LoadXChecked` (dipanggil handler
SEBELUM `c.Bind`) + `ApplyXUpdate` (dipanggil SETELAH bind). Jangan pernah
"merapikan" urutan ini menjadi bind-dulu demi estetika layering.
