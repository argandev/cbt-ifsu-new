# RELEASE v2.1.0 — CBT SMK v2 (Paket Final)

**Versi:** v2.1.0 (Fase 3: Hardening)
**Tanggal paket:** 16 Okt 2026
**Status:** Build final + verifikasi paket lulus (lihat "Verifikasi paket" di bawah).

## Isi paket (`cbt-smk-v2-final.zip`)

| Item | Keterangan |
|---|---|
| `cbt-smk-v2` | Binary Go hasil `make build` (Linux x86-64, ~19,8 MB). Frontend React sudah ter-embed di dalam binary — tidak perlu `node_modules` / server frontend terpisah. |
| `backend/` | Source backend Go lengkap (Echo v5, pgx, sqlc, migrasi ter-embed) — **TANPA** `backend/dist/` (hasil build, tidak perlu di paket). |
| `frontend/` | Source frontend React + TypeScript + Vite lengkap — **TANPA** `node_modules/` dan **TANPA** `frontend/dist/`. |
| `docs/` | Seluruh dokumentasi: `VERIFIKASI-AKHIR.md`, `KONTRAK.md`, `KEAMANAN.md`, `PANDUAN-DEPLOY-SEKOLAH.md`, `UJI-BEBAN-300.md`, `BUG-REPORT-QA-BEBAN.md`, `CHANGELOG.md`, dll. |
| `deploy/` | `start.sh`, `backup.sh`, `cbt-v2.service` (systemd), `Caddyfile.contoh`. |
| `screenshots/` | 29 file bukti visual (halaman 01–28 + 1 tangkapan awal). |
| `README.md`, `CARA-JALAN.txt`, `ANALISIS.md`, `Makefile`, `.env.example`, `.gitignore` | Dokumen + build di root paket. |
| `RELEASE-v2.1.0.md` | Dokumen ini (salinan di root zip; aslinya `docs/RELEASE-v2.1.0.md`). |

**Yang SENGAJA tidak ada di paket:** `.env` (berisi secret lokal — jangan
pernah dibagikan), `node_modules/`, `frontend/dist/`, `backend/dist/`,
`storage/` (upload + backup = data sekolah, bukan bagian paket), `.git/`,
`*.dump`, `*.log`, file test scratch.

## Cara jalan dari zip

Prasyarat: Go 1.25+ (hanya bila ingin build ulang / pakai CLI dari source),
PostgreSQL 16, `psql`. Untuk jalan cepat, binary di paket sudah cukup +
`psql` untuk seed.

```bash
unzip cbt-smk-v2-final.zip
cd cbt-smk-v2-final          # atau nama folder hasil ekstrak

cp .env.example .env
# Isi di .env (WAJIB):
#   DATABASE_URL=postgres://<user>:<password>@localhost:5432/cbt_smk_v2
#   SESSION_SECRET=<acak panjang, min 32 karakter — contoh: openssl rand -hex 32>

# Buat database dulu (bila belum ada):
createdb cbt_smk_v2

# Migrasi (binary sudah berisi migrasi ter-embed):
./cbt-smk-v2 migrate-up

# Seed produksi (6 role, 62 permission, matriks role-permission,
# 6 settings default; TANPA akun demo / TANPA user apa pun):
# Catatan: binary TIDAK punya subcommand `seed` — seed lewat psql:
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f backend/db/seed/seed.sql
# (atau `make seed` bila Go + Makefile dipakai)

# Akun admin pertama (HANYA bila tabel users masih kosong):
./cbt-smk-v2 create-admin --email <email> --username <username> \
  --nama "Nama Admin" --password <password-kuat>
# Setelah itu login dan WAJIB ganti password saat login pertama
# (login ditolak 403 PASSWORD_CHANGE_REQUIRED sampai password diganti).

# Jalankan:
./cbt-smk-v2
# Buka dari perangkat lain di jaringan sekolah: http://<IP-server>:9988
```

**Port default: 9988** (diubah via `PORT` di `.env`).
Detail deploy server sekolah: `docs/PANDUAN-DEPLOY-SEKOLAH.md`.
Detail keamanan: `docs/KEAMANAN.md`. Kontrak teknis: `docs/KONTRAK.md`.

## Keterbatasan yang diketahui (baca sebelum go-live)

1. **Akun staf via SQL.** Tidak ada modul pembuatan akun staf di UI/API dan
   CLI `create-admin` hanya membuat role `admin` saat tabel `users` kosong.
   Akun proktor / guru / pengawas / penguji_ukk dibuat sekolah via SQL
   langsung ke tabel `users` (`INSERT` dengan `password_hash` Argon2id +
   `role` yang sesuai + `must_change_password=true`), atau memakai
   `create-admin` saat DB masih kosong lalu mengubah `role` via SQL.
   Akun **siswa** dibuat otomatis via master siswa (create/import,
   username = NISN, password awal = NISN). Reset password akun yang sudah
   ada: `./cbt-smk-v2 reset-password --email <email> --password <baru>`.
   Detail: `docs/VERIFIKASI-AKHIR.md` §7.
2. **Uji beban di LAN sekolah sebelum hari-H.** Angka uji beban (300 siswa,
   throughput fase ujian 510,49 req/dtk pada run verifikasi akhir, 0 respons
   5xx, 3000/3000 jawaban terverifikasi) diukur di VM bersama memakai
   loopback, bukan LAN sekolah yang sebenarnya. Uji ulang 600–1000 siswa di
   server sekolah yang sebenarnya sebelum hari-H. Detail:
   `docs/UJI-BEBAN-300.md`, `docs/VERIFIKASI-AKHIR.md` §5.
3. **Insiden transien QB3 (belum terjelaskan).** Pada run uji beban QB3, 7 dari
   300 aktivasi gagal di sisi klien, sementara log server mencatat 303/303
   ganti-password status 200 dan nol 429. Penyebab pasti tidak
   teridentifikasi (diduga race penutupan koneksi keep-alive di sisi klien)
   dan **tidak terulang pada QB4** (aktivasi gagal 0). Bila terulang di
   produksi, kumpulkan pesan error klien + log koneksi sebelum menyimpulkan.
   Detail: `docs/UJI-BEBAN-300.md` §3.5 butir 3.
4. **Import siswa: anjuran maks 100 siswa per file** (~17 detik).
5. **Batas teknologi web:** HP kedua, foto layar, joki, internet/AI di
   perangkat lain, screenshot OS, dan listrik/jaringan padam TIDAK bisa
   dicegah aplikasi web — lihat tabel jujur di `docs/KEAMANAN.md` §8.

## Verifikasi paket

- Build final: `make build` sukses; binary `cbt-smk-v2` jalan
  (`GET /` → 200, `GET /adm` → 200, `GET /login` → 200) pada port scratch
  dengan DB scratch (migrasi + seed + create-admin jalan).
- Isi zip diperiksa: tidak ada `.env`, `node_modules/`, `storage/`, `.git/`,
  `*.dump`.
- Pemindaian kredensial: tidak ada password/secret nyata di dalam zip.
  Pengecualian yang didokumentasikan (bukan kredensial yang berlaku):
  `ANALISIS.md` menyebut akun demo `admin123` dkk. dalam konteks
  dokumentasi desain/akun demo — seed paket ini membuat **nol user**,
  tidak ada akun demo yang bisa dipakai login.
- Uji ekstrak: zip diekstrak ke direktori sementara, binary hasil ekstrak
  dijalankan dengan DB scratch + migrasi + seed, `GET /` → 200, dan
  `create-admin` jalan.

Verifikasi akhir menyeluruh (E2E 33/33, beban, analisis butir, pembahasan):
`docs/VERIFIKASI-AKHIR.md`.
