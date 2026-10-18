# CBT SMK v2 — CARA JALAN DI WINDOWS

Dokumen ini khusus Windows. Cara jalan umum (Linux/macOS): `CARA-JALAN.txt`.
Detail deploy server sekolah: `docs/PANDUAN-DEPLOY-SEKOLAH.md`.

## Prasyarat

| Kebutuhan | Versi | Unduh |
|---|---|---|
| Go | 1.26+ (`go.mod`: `go 1.26.0`) | https://go.dev/dl/ |
| Node.js + npm | LTS | https://nodejs.org/ |
| PostgreSQL | 16 | https://www.postgresql.org/download/windows/ |
| Salah satu: Git Bash (membawa `make` + `cp`/`rm`) **atau** cmd.exe | — | https://git-scm.com/download/win |

Setelah install Go/Node, **tutup lalu buka lagi** terminal/cmd agar PATH baru terbaca.
Cek: `go version` → harus 1.26 atau lebih baru. `node --version` → harus jalan.

## ⚠️ PENTING — Matikan exe lama dulu sebelum build ulang

Windows **mengunci** file `.exe` yang sedang jalan. Bila `cbt-smk-v2.exe`
masih jalan saat build:

- `go build` gagal dengan pesan `Access is denied`, **atau**
- build terlihat sukses tapi exe yang jalan tetap versi lama (semua halaman
  terlihat lama, perbaikan tidak muncul).

**Cara benar:** tekan `Ctrl+C` di jendela server lama (atau tutup jendela
itu), pastikan proses `cbt-smk-v2.exe` sudah berhenti, **baru** build ulang.
Cara verifikasi build baru: form Bank Soal → Mapel harus dropdown, bukan
input UUID (lihat riwayat perbaikan di `docs/CHANGELOG.md` bila ada).

## 1) Konfigurasi (sekali saja)

Di root repo (folder yang berisi `Makefile` dan `build.bat`):

```bat
copy .env.example .env
```

Lalu isi `.env` (WAJIB, tanpa default untuk secret):

```ini
DATABASE_URL=postgres://<user>:<password>@localhost:5432/cbt_smk_v2
SESSION_SECRET=<acak panjang, min 32 karakter>
```

Buat database-nya (di `psql` atau pgAdmin):

```sql
CREATE DATABASE cbt_smk_v2;
```

`.env` JANGAN di-commit ke git.

## 2) Migrasi + seed (sekali saja, setelah konfigurasi)

```bat
make migrate-up
make seed
```

`make seed` memakai `psql` (termasuk dalam install PostgreSQL — pastikan
folder `bin` PostgreSQL ada di PATH) dan membaca `DATABASE_URL` dari `.env`.
Seed mengisi: 6 role, 62 permission, matriks role-permission, 6 settings
default. **TANPA** akun demo / user apa pun.

## 3) Akun pertama (via CLI)

```bat
cd backend
go run ./cmd/server create-admin --email admin@sekolah.id --password <password> --nama "Nama Admin"
cd ..
```

`create-admin` HANYA bila tabel users kosong. Setelah itu login dan **WAJIB**
ganti password saat login pertama (login ditolak `403 PASSWORD_CHANGE_REQUIRED`
sampai password diganti).

## 4) Build

### Opsi A — pakai `make` (Git Bash / mingw32-make) — disarankan

Buka **Git Bash** (bukan cmd.exe murni — recipe Makefile memakai
`cp`/`rm -rf`/`mkdir -p` yang disediakan Git Bash), dari root repo:

```bash
# Build untuk Windows (hasil: ./cbt-smk-v2.exe)
make build-windows

# Atau build untuk OS saat ini (di Windows = sama dengan build-windows)
make build

# Target lain yang tersedia:
make frontend       # build frontend saja
make run            # jalankan server (go run ./cmd/server serve)
make migrate-up     # jalankan migrasi yang tertunda
make migrate-down   # rollback 1 langkah (STEPS=2 untuk 2 langkah: make migrate-down STEPS=2)
make seed           # seed produksi via psql
make test           # go test ./... + typecheck frontend
make clean          # hapus binary + dist (node_modules dipertahankan)
make clean-all      # clean + hapus frontend/node_modules
make dist           # paket dist/ (binary + CARA-JALAN + .env.example + seed.sql)
make zip            # arsip zip dari dist/
make help           # daftar target
```

Bila `make` tidak ditemukan di Git Bash, install `mingw32-make` (ikut
paket MinGW) atau pakai Opsi B.

### Opsi B — tanpa `make`: `build.bat` (cmd.exe / PowerShell)

Dari root repo, di cmd.exe atau PowerShell:

```bat
build.bat
```

`build.bat` menjalankan langsung: cek `go` + `npm` → `npm install` →
`npm run build` → salin `frontend/dist` ke `backend/dist` → `go build`
(`GOOS=windows GOARCH=amd64`) → `.\cbt-smk-v2.exe`. Bila gagal, pesan
error ditampilkan dan jendela berhenti di `pause` agar bisa dibaca.
Khusus error `Access is denied`: itu tanda exe lama masih jalan — matikan
dulu (lihat ⚠️ di atas), lalu jalankan `build.bat` lagi.

## 5) Menjalankan server

```bat
cbt-smk-v2.exe
```

Lalu buka di browser: http://localhost:9988
Dari perangkat lain di jaringan sekolah: `http://<IP-server>:9988`.

Port default aplikasi: **9988**. PostgreSQL: **5432** (hanya localhost).

Lupa password admin → reset via CLI (dari `backend/`):

```bat
go run ./cmd/server reset-password --email <email> --password <password-baru>
```

Jangan edit hash password manual di database.

## Troubleshooting Windows

| Gejala | Penyebab / solusi |
|---|---|
| `make: command not found` | `make` tidak ada — pakai `build.bat`, atau install Git Bash / mingw32-make. |
| `Access is denied` saat build | Exe lama masih jalan → `Ctrl+C` / tutup jendela server lama, build ulang. |
| Halaman tetap lama setelah build | Exe lama masih jalan (lihat di atas) **atau** build gagal tapi tidak disadari — baca pesan error sampai akhir. |
| `go: go.mod requires go >= 1.26` | Go terlalu lama → install Go 1.26+ dan buka terminal baru. |
| `psql: command not found` | Folder `bin` PostgreSQL belum di PATH → tambahkan (mis. `C:\Program Files\PostgreSQL\16\bin`). |
| `npm: command not found` | Node.js belum terinstall / terminal belum dibuka ulang setelah install. |
| Port 9988 sudah dipakai | Server lama masih jalan (dua jendela) → matikan salah satu. |
