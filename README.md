# CBT SMK v2

Web ujian online untuk SMK — STS / SAS / USBK / UKK / Tryout — dengan
bank soal, paket ujian, jadwal + token, koreksi esai, rekap nilai,
analisis butir soal, nilai per kompetensi, event + ruangan + kartu
peserta, backup/restore, dan modul UKK (teori + praktik).

> **Label jujur tentang tipe soal:** v2 mendukung
> **5 tipe standar AKM + 11 varian/ekstensi aplikasi**.
> Kami TIDAK mengklaim "16 tipe standar Kemendikbud" — klaim tersebut
> tidak ditemukan di dokumen resmi Kemendikbud. Dokumen resmi paling
> otoritatif, *Framework AKM* (Pusat Asesmen Pendidikan, BSKAP,
> Kemendikbudristek, edisi revisi 2023, Lampiran 1 hlm. 147–149),
> menetapkan **hanya 5 bentuk soal**: Pilihan Ganda, Pilihan Ganda
> Kompleks, Menjodohkan, Isian/Jawaban Singkat, Uraian/Esai.
> Rincian tiap tipe (kode, cara jawab, aturan nilai) ada di
> `docs/TIPE-SOAL.md`. Di UI, tipe inti berlabel `[Inti AKM]`,
> varian berlabel `[Varian aplikasi]`.
>
> 5 inti: `pg`, `pg_kompleks`, `menjodohkan`, `isian`, `uraian`.
> 11 varian: `benar_salah`, `pg_multi`, `urutkan`, `isi_ganda`,
> `teks_rumpang`, `peta_konsep`, `studi_kasus`, `pernyataan_sikap`
> (non-nilai, survei/refleksi), `upload_file`, `praktik_singkat`,
> `pilihan_gambar`.

## Stack

| Lapisan   | Teknologi |
|-----------|-----------|
| Backend   | Go (go.mod: `go 1.26`, prasyarat Go 1.25+), modul `cbt-smk-v2/backend`, Echo v5, pgx v5 + sqlc |
| Database  | PostgreSQL 16 (migrasi SQL ter-embed, seed idempotent) |
| Frontend  | React 19 + TypeScript + Vite 6 + Tailwind CSS 4 + lucide-react + Zustand + TanStack Query |
| Auth      | Session cookie server-side (`HttpOnly`, `SameSite`) + token CSRF — **bukan JWT** |
| Password  | Argon2id (format PHC) |
| Deploy    | Satu binary Go, frontend ter-embed (`go:embed`) — server lokal sekolah (LAN, semi-daring) |

Aturan tampilan: **nol emoji** di UI — ikon memakai `lucide-react`.
Warna primary `#2563EB` (indigo), font **Plus Jakarta Sans**
(via `@fontsource/plus-jakarta-sans`).

Deploy utama: **server lokal sekolah (LAN, semi-daring)**.
**cPanel shared hosting TIDAK dipakai** untuk aplikasi CBT
(tidak ada proses persisten, port kustom, PostgreSQL, atau systemd).

## Fitur per fase

### Fase 1 — MVP

- Auth: session server-side + CSRF, wajib ganti password saat login
  pertama, lockout login (5x gagal / 15 menit).
- RBAC: 6 role (`admin`, `proktor`, `guru`, `pengawas`, `siswa`,
  `penguji_ukk`), **62 permission** per aksi
  (matriks lengkap di `docs/KONTRAK.md` §5).
- Master data: jurusan, kelas, mapel, guru, siswa (CRUD penuh +
  import siswa CSV, idempotent).
- Bank soal: CRUD, lock/unlock, import, 5 + 11 tipe soal.
- Paket ujian + jadwal: token (token kedua menghanguskan pertama,
  expiry 900 detik default), toleransi telat 15 menit,
  pause yang **membekukan** timer.
- Ujian siswa: timer server-side, auto-submit, autosave jawaban,
  tandai ragu-ragu, pelaporan pelanggaran, auto-block per-attempt
  (pelanggaran ke-1 = peringatan, ke-2 = auto-block),
  unblock = lanjut sisa waktu (hanya Proktor/Admin mengeksekusi).
- Koreksi esai/uraian guru; nilai akhir dihitung backend
  (nilai = Σ poin / Σ bobot × 100).
- Rekap + export CSV + berita acara; monitor peserta real-time
  (WebSocket, ticket 30 detik); settings; audit log append-only.
- Tryout: tanpa token + pembahasan langsung.

### Fase 2 — UKK / analisis / event / backup

- **UKK:** skema + kompetensi + peserta, penilaian teori 40% +
  praktik 60%, rubrik Proses 30% / Produk 50% / Sikap 20%,
  lulus ≥ 70, sertifikat (cetak, `print.css`).
- **Analisis butir soal:** tingkat kesukaran (P), daya beda,
  pengecoh, jawaban kosong dibedakan dari salah
  (rumus di `docs/ANALISIS-RUMUS.md`); export CSV.
  Proktor melihat statistik **tanpa** teks pertanyaan.
- **Nilai per kompetensi:** agregasi nilai per kompetensi dari
  attempt selesai pada satu jadwal.
- **Event + ruangan + kartu peserta:** event ujian, ruangan,
  kartu peserta admin (`/adm/kartu`) + kartu sendiri siswa
  (`/kartu`, `/kartu/:jadwalId`).
- **Backup/restore:** lewat UI (`/adm/backup`: buat, daftar, unduh,
  restore) + cron `deploy/backup.sh` (`pg_dump` format arsip +
  arsip upload). Restore membuat backup pre-restore otomatis dan
  **ditolak bila ada attempt berstatus berjalan**.
- **Frontend extras:** halaman 403/404, error boundary,
  `ConfirmDialog` (pengganti `window.confirm`), `StatCard`,
  `PageHeader`, `EmptyState`, `Pagination`, `SearchInput`,
  `StatusBadge`, `PrintLayout`, `Sidebar`/`Topbar`/`Breadcrumbs`,
  bel notifikasi, alur token siswa di beranda siswa
  (input token, lanjut attempt berjalan, lihat hasil).

### Fase 3 — Hardening

- **SoD override:** pengecualian darurat pemisahan tugas HANYA lewat
  `POST /api/v1/sod/override` (admin saja, `sod.override`), alasan
  tertulis min 20 karakter, dicatat di `sod_overrides` + `audit_log`
  dengan flag `SOD_OVERRIDE`, + notifikasi ke semua admin lain.
- **Perbaikan BUG-01/02/03** (laporan lengkap di
  `docs/BUG-REPORT-QA-BEBAN.md`): permission `analisis.*`
  diselaraskan ke 4 sumber (kontrak + seed + `rbac.go` + migrasi);
  timeout HTTP eksplisit (Read 60 dtk, Write 120 dtk, Idle 120 dtk);
  import siswa idempotent (`sudah_ada`, bukan gagal);
  rate limit ujian 300 req/menit/IP + header `Retry-After` +
  retry 429 di frontend (maks 3x, backoff 500 ms → 1 dtk → 2 dtk,
  mengikuti `Retry-After` maks 10 dtk).
- **Pembahasan end-to-end:** guru mengisi `soal.pembahasan` lewat
  bank soal; pembahasan tampil di hasil bila tampil_hasil efektif
  (Tryout ATAU paket `tampil_hasil` AND settings
  `tampil_hasil_ke_siswa`). Pembahasan TIDAK PERNAH dikirim sebelum
  submit.
- **Hardening:** hapus via POST (bukan GET), anti-IDOR (resource
  yang tidak berhak → 404, bukan 403), kunci jawaban tidak pernah
  ke browser siswa (terverifikasi 0 kebocoran pada uji beban).

## Struktur monorepo

```
cbt-smk-v2/
├── backend/
│   ├── cmd/server/        # entrypoint: serve, migrate-up/down, create-admin, reset-password
│   ├── internal/          # config, auth, session, rbac, audit, master, banksoal,
│   │                      # exam, grading, monitor, rekap, settings, analisis,
│   │                      # event, backup, ukk, sod, notifikasi, middleware
│   ├── db/
│   │   ├── migrations/    # migrasi SQL 0001–0012 (ter-embed ke binary)
│   │   ├── queries/       # kueri sqlc
│   │   └── seed/seed.sql  # seed PRODUKSI: 6 role + 62 permission + matriks + settings
│   └── dist/              # hasil build frontend (disalin `make build`, di-embed)
├── frontend/              # React 19 + TS + Vite 6 + Tailwind 4 (src/pages, src/components, src/lib)
├── docs/
│   ├── KONTRAK.md                 # kontrak bersama (sumber kebenaran: 62 permission, session+CSRF, endpoint, route)
│   ├── TIPE-SOAL.md              # 5 + 11 tipe soal
│   ├── ANALISIS-RUMUS.md         # rumus analisis butir + nilai per kompetensi
│   ├── UKK-RUBRIK-SERTIFIKAT.md # rubrik + sertifikat UKK
│   ├── PANDUAN-DEPLOY-SEKOLAH.md # deploy server lokal sekolah
│   ├── KEAMANAN.md               # ringkasan keamanan (jujur: termasuk yang TIDAK bisa dicegah web)
│   ├── CHANGELOG.md              # riwayat versi
│   ├── BUG-REPORT-QA-BEBAN.md    # BUG-01/02/03 + bukti uji beban
│   └── UJI-BEBAN-300.md          # metode + hasil uji beban 300 siswa
├── deploy/
│   ├── cbt-v2.service     # contoh unit systemd
│   ├── start.sh           # build + jalankan (manual / tanpa systemd)
│   ├── backup.sh          # backup PostgreSQL (pg_dump) + upload, retensi default 30 hari
│   └── Caddyfile.contoh   # reverse proxy opsional (LAN)
├── README.md
├── CARA-JALAN.txt
└── ANALISIS.md            # analisis & keputusan desain v2
```

## Cara jalan (ringkas)

Langkah lengkap ada di `CARA-JALAN.txt`. Intinya:

```bash
# 1. Prasyarat: Go 1.25+, Node.js + npm, PostgreSQL 16. Buat database.
createdb cbt_smk_v2

# 2. Konfigurasi
cp .env.example .env   # lalu isi DATABASE_URL & SESSION_SECRET (wajib!)

# 3. Migrasi + seed produksi (6 role, 62 permission, settings — TANPA akun demo)
make migrate-up
make seed

# 4. Akun pertama (TIDAK ada akun demo / bawaan)
cd backend && go run ./cmd/server create-admin --email <email> --password <password> --nama "Nama Admin"

# 5a. Development (dari root repo)
cd backend && go run ./cmd/server &   # backend :9988 (default: serve)
npm --prefix frontend run dev          # frontend Vite

# 5b. Produksi (satu binary, frontend ter-embed)
make build          # -> binary ./cbt-smk-v2
./cbt-smk-v2        # buka http://<IP-server>:9988
```

Login pertama **wajib ganti password** (`must_change_password=true`;
login ditolak 403 `PASSWORD_CHANGE_REQUIRED` sampai diganti).
Port default aplikasi: **9988**. PostgreSQL: 5432 (hanya localhost).

## Akun pertama & keamanan (ringkas)

- **Tidak ada akun demo.** Seed tidak membuat user apa pun.
- Akun pertama dibuat lewat CLI `create-admin` (hanya bila tabel
  `users` kosong). Akun tersebut wajib ganti password saat login
  pertama; sampai diganti, sesi tidak dibuat.
- Password di-hash Argon2id, tidak pernah disimpan plaintext.
- Sesi disimpan server-side (DB); cookie `cbt_session` `HttpOnly` +
  `SameSite=Lax` (+ `Secure` bila prod), TTL 12 jam, idle 2 jam.
  Restart server **tidak** membuat logout massal (sesi ada di DB).
- Kunci sesi / token **tidak pernah** dikirim ke browser / JavaScript /
  `localStorage` / query string / log. Token ujian tidak muncul di URL.
- Setiap request yang mengubah state wajib membawa token CSRF
  (header `X-CSRF-Token` == token sesi, dibaca dari cookie
  `cbt_csrf`; perbandingan constant-time).
- Hapus memakai `POST` (bukan `GET`) — tautan tidak boleh menghapus data.
- Detail lengkap: `docs/KEAMANAN.md`. Deploy sekolah:
  `docs/PANDUAN-DEPLOY-SEKOLAH.md`.

## Peran & permission

6 role: `admin`, `proktor`, `guru`, `pengawas`, `siswa`, `penguji_ukk`.
**62 permission** per aksi. Jumlah per role (matriks lengkap di
`docs/KONTRAK.md` §5): admin 62, proktor 30, guru 27, pengawas 4,
siswa 2, penguji_ukk 4.

Pengawas **tidak** melihat token ujian dan **tidak** melihat isi soal.
Proktor **tidak** melihat isi soal (di analisis butir, field
`pertanyaan` disembunyikan untuk proktor). Pengawas BOLEH mengusulkan
unblock; eksekusi unblock tetap proktor/admin. Penegakan ada di server,
bukan hanya dengan menyembunyikan tombol di frontend.

## Aturan ujian (ringkas)

- Token kedua membuat token pertama hangus; token kedaluwarsa 900 detik
  (default, dapat diubah di Settings).
- Toleransi telat 15 menit (default). Pause membekukan timer.
- Auto-block: pelanggaran ke-1 = peringatan, ke-2 = auto-block
  **per-attempt** (bukan mengubah status akun). Unblock = lanjut sisa
  waktu; hanya Proktor/Admin yang mengeksekusi unblock.
- Tryout: tanpa token + pembahasan langsung.
- Nilai esai disembunyikan sampai final (default
  `tampil_hasil_ke_siswa=false`).
- KKTP per mapel/TP (bukan hardcode KKM=75). Remedial memakai nilai
  tertinggi. UKK: teori 40% + praktik 60%; rubrik Proses 30% /
  Produk 50% / Sikap 20%; lulus ≥ 70.

## Keterbatasan yang diketahui (jujur)

- **Akun staf dibuat lewat halaman Manajemen Pengguna oleh Admin.**
  Menu Sistem → Pengguna (`/adm/pengguna`, endpoint `/api/v1/pengguna`):
  buat akun staf (username unik, nama, email opsional, role Admin /
  Proktor / Guru / Pengawas / Penguji UKK, kata sandi minimal 8
  karakter), ubah nama / email / role (username tidak dapat diubah),
  reset kata sandi (pengguna wajib ganti password saat login
  berikutnya), dan nonaktifkan / aktifkan akun (menonaktifkan akun
  sendiri atau admin terakhir ditolak). Selain itu, pembuatan akun di
  kode aplikasi adalah `create-admin` (hanya bila tabel `users` kosong)
  dan akun siswa otomatis (import / buat siswa: username = NISN,
  password awal = NISN, wajib ganti saat login pertama); profil guru
  (`master/guru`) wajib merujuk `user_id` yang sudah ada.
- **Angka uji beban bukan jaminan untuk hardware lain.** Uji beban
  300 siswa (300/300 submit, 3000/3000 jawaban, 0 respons 5xx,
  throughput fase ujian 95,62 req/dtk) diukur di VM bersama memakai
  loopback, bukan LAN sekolah yang sebenarnya. Uji ulang di server
  sekolah sebelum hari-H (lihat `docs/UJI-BEBAN-300.md` §4 butir 9).
- **Insiden transien QB3 belum terjelaskan.** Pada run QB3, 7 dari 300
  aktivasi gagal di sisi klien sementara log server mencatat 303/303
  ganti-password 200. Penyebab pasti tidak teridentifikasi dan tidak
  terulang pada QB4 (aktivasi gagal 0). Bila terulang di produksi,
  kumpulkan pesan error klien + log koneksi sebelum menyimpulkan.
- **RSS server saat uji beban belum dijelaskan.** VmRSS tercatat
  1,1–1,5 GB dan TURUN selama run (bukan bocor monotonik), tetapi angka
  absolutnya besar untuk server API Go. Perlu profiling `pprof`
  (heap + goroutine) sebelum menyatakan normal.
- **Import siswa: anjuran maks 100 siswa per file.** Fase 3 menaikkan
  `WriteTimeout` ke 120 dtk sehingga import 300 baris (~50 dtk) tidak
  lagi terpotong, tetapi batch 100/file (~17 dtk) tetap anjuran
  operasional agar respons tetap cepat.
- **Verifikasi akhir masih menunggu.** Dokumen
  `docs/VERIFIKASI-AKHIR.md` belum ada saat dokumen ini ditulis.
  Yang masih berstatus "menunggu verifikasi akhir": `make sqlc`
  pasca-perbaikan migrasi 0005, dan pengujian e2e pasca-Fase-3 di
  database bersih (migrate → seed → create-admin → alur penuh).
- **Batas teknologi web (bukan kekurangan implementasi):** HP kedua,
  foto layar, joki, internet/AI di perangkat lain, screenshot OS,
  dan listrik/jaringan padam TIDAK bisa dicegah aplikasi web —
  lihat tabel jujur di `docs/KEAMANAN.md` §8.

## Lisensi & asal kode

Kode v2 ditulis dari nol untuk proyek ini (berdasarkan ANALISIS.md).
Tidak menyalin kode mentah dari proyek lain dan tidak menyalin
mekanisme lisensi produk mana pun.
