# Panduan Deploy — Server Lokal Sekolah (CBT SMK v2)

Deploy utama v2: **satu komputer/server di sekolah, diakses lewat
LAN/WiFi sekolah (semi-daring)**. Ujian jalan walau internet putus.
**cPanel shared hosting TIDAK dipakai** untuk aplikasi CBT
(tidak ada proses persisten, port kustom, PostgreSQL, atau systemd).

Dokumen ini milik Tim Fase 3 (Dokumen Final). Kontrak teknis (stack,
permission, aturan ujian) ada di `docs/KONTRAK.md`. Keamanan ada di
`docs/KEAMANAN.md`. Cara jalan singkat: `CARA-JALAN.txt`.

---

## 1. Spesifikasi server

### Minimal (ujian kecil, < 100 siswa serentak)

| Komponen | Minimal |
|----------|---------|
| CPU | 2 core |
| RAM | 4 GB |
| Disk | SSD 50 GB |
| OS | Linux (Ubuntu/Debian/AlmaLinux) |
| Jaringan | IP statis di LAN sekolah, kabel LAN ke switch |

### Disarankan (300 siswa serentak)

| Komponen | Disarankan | Alasan |
|----------|-----------|--------|
| CPU | **4 core** | Request HTTP + penilaian + monitor real-time |
| RAM | **8 GB** | PostgreSQL + Go + cache + 300 sesi |
| Disk | **SSD 100 GB** | Database + upload UKK/media soal + backup lokal |
| OS | Linux LTS | systemd, update keamanan |
| Jaringan | IP statis + kabel LAN (bukan hanya WiFi) | Latensi < 5 ms, stabil |
| Daya | **UPS** | Listrik padam di tengah ujian = sesi terganggu |
| Cadangan | Komputer cadangan / rencana pinjam lab lain | Satu titik kegagalan |

### Hasil uji beban (terukur, bukan target)

Uji beban 300 siswa, run final QB4, 16 Okt 2026
(metode + tabel lengkap di `docs/UJI-BEBAN-300.md`):

| Metrik | Angka |
|--------|-------|
| Siswa serentak / soal per siswa | 300 / 10 |
| Submit OK | **300/300** |
| Jawaban tersimpan (terverifikasi di DB) | **3000/3000** |
| Respons 5xx / timeout / deadlock | **0 / 0 / tidak teramati** |
| Kebocoran kunci jawaban ke siswa | **0** |
| **Throughput fase ujian** | **~95 req/detik** (95,62 req/dtk; 3900 request sukses / 40,8 dtk) |
| Throughput keseluruhan (termasuk setup + aktivasi) | 33,73 req/dtk |

Catatan jujur: angka diukur di VM bersama memakai loopback, bukan
LAN sekolah yang sebenarnya. Angka ini **bukan jaminan** untuk
1000+ siswa atau hardware berbeda — uji ulang 600–1000 siswa di
server sekolah yang sebenarnya sebelum hari-H, dengan jaringan LAN.
Pada run QB4 terjadi 1800 respons 429 `rate_limited` (30 IP berbagi,
10 siswa/IP) yang semuanya pulih via retry; Fase 3 sudah menaikkan
batas ujian ke 300 req/menit/IP + retry 429 di frontend
(lihat `docs/BUG-REPORT-QA-BEBAN.md` BUG-03).

Perangkat siswa: browser modern (Chrome/Edge/Firefox terbaru), HP /
tablet / laptop. Tidak perlu install aplikasi di perangkat siswa.

---

## 2. Instalasi prasyarat (server)

Contoh untuk Ubuntu/Debian (sesuaikan dengan distro sekolah):

```bash
sudo apt update && sudo apt install -y postgresql-16 curl git ca-certificates

# Go 1.25+ — pakai versi sesuai go.mod (go 1.26; unduh dari go.dev bila repo distro tertinggal)
# Node.js LTS — untuk build frontend (npm run build)
# Contoh via NodeSource / nvm; pastikan `node --version` dan `npm --version` jalan.
```

PostgreSQL:

```bash
sudo systemctl enable --now postgresql
sudo -u postgres psql -c "CREATE USER cbt WITH PASSWORD 'ganti-password-kuat';"
sudo -u postgres psql -c "CREATE DATABASE cbt_smk_v2 OWNER cbt;"
```

> Password database di atas hanya contoh. Pakai password kuat dan
> simpan di `.env` / environment service — jangan di-commit, jangan
> ditulis di dokumen bersama.

---

## 3. Build binary

```bash
git clone <repo-cbt-smk-v2> /opt/cbt-smk-v2
cd /opt/cbt-smk-v2

cp .env.example .env
# Isi .env: DATABASE_URL, SESSION_SECRET (acak, min 32 karakter),
# APP_ENV=production, PORT=9988

# Build: frontend (Vite) -> salin dist ke backend/dist -> binary Go satu file.
# Modul Go ada di backend/ (backend/go.mod); `make build` mengurusnya.
make build

# Migrasi + seed produksi
make migrate-up
make seed
```

`make seed` mengisi 6 role + 62 permission + matriks + settings.
**Seed tidak membuat akun apa pun** — tidak ada akun demo.
Akun pertama:

```bash
cd backend && go run ./cmd/server create-admin --email <email> --password <password> --nama "Nama Admin"
# create-admin HANYA bila tabel users kosong.
# Login, lalu WAJIB ganti password saat login pertama.
```

Atau jalankan dengan skrip (build + jalankan):

```bash
./deploy/start.sh
```

> JANGAN menjalankan `go build ./cmd/server` dari root repo — path
> itu tidak ada (yang ada `backend/cmd/server`). Pakai `make build`
> atau build dari dalam `backend/`:
> `(cd backend && go build -o ../cbt-smk-v2 ./cmd/server)`.

---

## 4. Menjalankan sebagai service (systemd)

Contoh unit lengkap ada di `deploy/cbt-v2.service`. Garis besarnya:

```bash
sudo cp deploy/cbt-v2.service /etc/systemd/system/cbt-v2.service
sudo systemctl daemon-reload
sudo systemctl enable --now cbt-v2
sudo systemctl status cbt-v2
sudo journalctl -u cbt-v2 -f   # melihat log
```

Poin penting unit service:

- `User` khusus `cbt` (bukan root), `WorkingDirectory` = `/opt/cbt-smk-v2`.
- `EnvironmentFile=/etc/cbt-smk-v2/env` (permission 600, milik root)
  untuk `DATABASE_URL` & `SESSION_SECRET` + `APP_ENV=production` +
  `PORT=9988` — produksi memakai env yang disuntikkan sistem,
  bukan `.env` yang ikut ter-commit.
- `After=network.target postgresql.service`,
  `Restart=on-failure`, `RestartSec=5`.
- `ProtectSystem=strict`, `PrivateTmp=true`, `NoNewPrivileges=true`.
- Aplikasi gagal start (fail-fast) bila `SESSION_SECRET` kosong di
  prod atau konfigurasi tidak valid — itu perilaku yang benar,
  bukan bug.

---

## 5. Reverse proxy opsional (LAN)

Akses langsung `http://<IP-server>:9988` sudah cukup untuk ujian di LAN.
Bila sekolah ingin nama host yang mudah diingat (mis. `http://cbt`
atau `http://cbt.sekolah.local`), pakai reverse proxy di server yang
sama. Contoh Caddy ada di `deploy/Caddyfile.contoh`:

```
# deploy/Caddyfile.contoh — jalankan `caddy run` / `caddy adapt` sesuai
# instalasi Caddy di server. Ganti cbt.sekolah.local dengan nama host LAN.
http://cbt.sekolah.local {
    reverse_proxy 127.0.0.1:9988
}
```

Catatan:

- Proxy hanya meneruskan ke `127.0.0.1:9988`; aplikasi tidak perlu
  diekspos langsung ke internet.
- PostgreSQL port 5432 **hanya localhost** — jangan dibuka ke LAN.
- Bila sekolah memakai HTTPS (sertifikat internal / domain resmi),
  terminasi TLS dilakukan di proxy; cookie `Secure` diaktifkan bila
  akses sudah HTTPS (lihat `docs/KEAMANAN.md`).

---

## 6. Konfigurasi produksi yang perlu dicek

| Variabel | Produksi |
|----------|----------|
| `APP_ENV` | `production` (menonaktifkan mode debug/verbose) |
| `PORT` | `9988` |
| `SESSION_SECRET` | Acak panjang, unik per server, **tidak sama** dengan contoh |
| `DATABASE_URL` | User khusus `cbt`, host `localhost` |
| Pool DB | `MaxConns = 30` (terverifikasi di `backend/cmd/server/main.go`). Uji beban merekomendasikan 50–100 sebagai tuning lanjutan (sesuaikan `max_connections` Postgres) — bukan nilai saat ini. |
| Upload | Batas ukuran via env, bukan hardcode |
| Backup | UI (`/adm/backup`) + cron `deploy/backup.sh` aktif (lihat §7) |

---

## 7. Backup & restore

### Backup — dua jalur (keduanya tersedia)

**a) Lewat UI (`/adm/backup`, butuh permission `backup.create`):**

- Buat backup manual: `POST /api/v1/backup` (jenis `manual`).
- Daftar + unduh file backup: `GET /api/v1/backup`,
  `GET /api/v1/backup/:id/download`.
- File backup: dump PostgreSQL + metadata (`backup_meta`:
  nama file, ukuran, jenis, waktu). Audit mencatat `backup.create`.

**b) Lewat cron (`deploy/backup.sh`, `pg_dump`):**

`deploy/backup.sh` membuat dump PostgreSQL berformat arsip
(`pg_dump -F c`) + timestamp dan mengarsipkan file upload.
Jalankan tiap malam + sebelum hari ujian:

```bash
# crontab -e (sebagai user yang menjalankan aplikasi)
0 2 * * * /opt/cbt-smk-v2/deploy/backup.sh >> /var/log/cbt-backup.log 2>&1
```

Konfigurasi via env (dengan default): `CBT_DB_NAME` (default
`cbt_smk_v2`), `CBT_DB_USER` (default `cbt`),
`CBT_BACKUP_DIR` (default `<repo>/backups`),
`CBT_UPLOAD_DIR` (default `<repo>/uploads`),
`CBT_BACKUP_KEEP_DAYS` (default **30** — file backup lebih lama
dihapus otomatis).

Yang di-backup: (1) database penuh, (2) file upload
(media soal, lampiran, berkas UKK — database hanya menyimpan metadata),
(3) konfigurasi server (tanpa password plaintext),
(4) versi aplikasi + versi migrasi saat backup diambil.
Yang tidak perlu: hasil build (binary bisa dibangun ulang), cache, file
sementara.

Simpan satu salinan backup **di luar server** (flashdisk / komputer
lain / drive sekolah) — backup yang hanya ada di server yang sama
tidak menyelamatkan dari disk rusak / server hilang.

Retensi: default skrip 30 hari. Audit log dipertahankan minimal
1 tahun ajaran.

### Restore (urutan yang benar)

Restore dilakukan lewat UI (`POST /api/v1/backup/:id/restore`,
butuh permission `backup.restore`) atau manual dengan `pg_restore`.
Perilaku yang dijamin server:

1. **Restore ditolak bila ada attempt ujian berstatus `berjalan`.**
   Jangan restore di tengah ujian — selesaikan / hentikan dulu
   semua ujian yang berjalan.
2. **Backup pre-restore otomatis.** Sebelum menimpa database,
   server membuat backup pengaman jenis `pre-restore` + mencatatnya
   di `backup_meta` + audit, sehingga restore yang salah masih bisa
   dikembalikan.

Langkah restore manual (bila tidak lewat UI):

1. Hentikan service: `sudo systemctl stop cbt-v2` (tidak ada yang
   menulis ke DB selama restore).
2. Pastikan tidak ada attempt berjalan
   (`SELECT count(*) FROM attempt WHERE status = 'berjalan';`
   harus 0) — dan buat backup pengaman dulu
   (`./deploy/backup.sh`).
3. Siapkan database tujuan: database kosong baru (atau database yang
   akan ditimpa — pastikan backup terbarunya sudah aman).
4. Restore database: `pg_restore` (format arsip) atau `psql`
   (format teks), sesuai format backup.
5. Kembalikan file upload ke direktori yang sama (atau direktori baru
   bila konfigurasi diubah), periksa izin berkas.
6. Kembalikan konfigurasi (koneksi DB, port, direktori file),
   sesuaikan bila server tujuan berbeda.
7. Cek versi migrasi di DB hasil restore; jalankan migrasi yang
   tertinggal bila aplikasi tujuan lebih baru. Bila aplikasi tujuan
   **lebih lama** dari backup, jangan memaksakan restore — pakai
   versi aplikasi yang sama dengan backup.
8. Uji kesehatan: aplikasi start, login admin berhasil, satu bank
   soal terbuka, satu jadwal terbaca, satu hasil ujian contoh cocok.
9. Baru buka layanan untuk pengguna setelah uji kesehatan lulus.

---

## 8. Update binary

1. Backup dulu (database + file upload).
2. Stop service: `sudo systemctl stop cbt-v2`.
3. Build binary baru dari tag/versi yang dituju; ganti file binary.
4. Jalankan migrasi: `make migrate-up` (migrasi dirancang maju saja;
   jangan jalankan migrasi turun di produksi tanpa rencana).
5. Start service, cek log + uji kesehatan (§7 langkah 8).
6. Bila gagal: stop, kembalikan binary lama + restore backup bila
   migrasi sudah mengubah DB, lalu catat insiden.

**Jangan update mendadak H-1 / hari-H ujian.** Bekukan versi minimal
seminggu sebelum ujian besar; update hanya bila ada perbaikan yang
diuji di server cadangan.

---

## 9. Troubleshooting

| Gejala | Periksa / tindakan |
|--------|-------------------|
| Aplikasi tidak start | `journalctl -u cbt-v2 -n 100`; cek `SESSION_SECRET` kosong? `DATABASE_URL` salah? port bentrok? |
| Port 9988 bentrok | `ss -tlnp \| grep 9988` — cari proses lama (mis. CBT lain / instance ganda) lalu stop yang tidak dipakai. Jangan jalankan dua CBT di port yang sama. |
| DB tidak jalan | `systemctl status postgresql`; cek PostgreSQL jalan, user/password & nama DB di `DATABASE_URL` benar, `pg_isready`. |
| Login admin gagal | Pastikan akun dibuat via `create-admin`; caps-lock; cek akun tidak terkunci (5x gagal / 15 menit → 429 `login_locked`, tunggu 15 menit). |
| **Lupa password admin** | Reset via CLI (dari `backend/`): `go run ./cmd/server reset-password --email <email> --password <password-baru>`. Bila tabel users kosong, buat admin baru via `create-admin`. **Jangan** edit hash password manual di database. |
| Siswa tidak bisa akses | IP server berubah? (pakai IP statis). Firewall server memblokir 9988? Satu WiFi/LAN yang sama? |
| Lambat saat ujian | Cek CPU/RAM (`top`), pool DB, log lambat; pastikan server memakai kabel LAN + SSD, bukan WiFi + HDD. Periksa juga apakah banyak siswa berbagi satu IP (NAT) — lihat BUG-03 di `docs/BUG-REPORT-QA-BEBAN.md`. |
| Sesi hilang setelah restart | Sesi tersimpan server-side di DB — restart server **tidak** membuat logout massal (kontrak §2.1). Bila sesi hilang, periksa apakah database ikut ter-reset / `DATABASE_URL` menunjuk DB yang salah, bukan restart service-nya. |
| Upload gagal | Cek batas ukuran env + ruang disk (`df -h`) + izin direktori upload. |
| Waktu / timer aneh | Samakan zona waktu server ke Asia/Jakarta (`timedatectl`); timer dihitung server, bukan browser. |
| **Migrasi dirty** (`schema_migrations` dirty) | Jangan jalankan `migrate-down` di produksi tanpa rencana. Identifikasi migrasi yang gagal, perbaiki penyebabnya (koneksi / SQL), lalu ulangi `make migrate-up`. Bila database tidak bisa dipulihkan, restore dari backup terakhir yang sehat (§7) — jangan mengedit tabel `schema_migrations` manual tanpa backup. |
| Import siswa tampak "gagal" tapi data masuk | Pola BUG-02 (commit-tanpa-respons) sudah diperbaiki Fase 3 (WriteTimeout 120 dtk + import idempotent `sudah_ada`). Bila terjadi, JANGAN retry buta: cek dulu jumlah siswa di DB / hasil import, lalu import ulang file yang sama — baris yang sudah ada dilaporkan `sudah_ada`, bukan gagal. |

---

## 10. Checklist go-live

- [ ] `SESSION_SECRET` diganti (acak, unik, bukan nilai contoh).
- [ ] `APP_ENV=production`; mode debug/verbose nonaktif.
- [ ] **Akun demo tidak ada**: seed tidak membuat user apa pun;
      verifikasi `SELECT count(*) FROM users;` = 0 sebelum akun
      pertama dibuat, lalu buat admin pertama via `create-admin`.
- [ ] Akun admin pertama dibuat via `create-admin`; **wajib ganti
      password** saat login pertama.
- [ ] `nama_sekolah` diisi di Settings (bukan nilai placeholder).
- [ ] KKTP per mapel/TP diisi (bukan hardcode KKM=75).
- [ ] Settings dicek: token 900 detik, toleransi telat 15 menit,
      max pelanggaran 2, remedial tertinggi, tampil hasil ke siswa
      sesuai kebijakan sekolah.
- [ ] Tes 10 siswa serentak: login → token → kerjakan → submit →
      nilai muncul di rekap; monitor proktor jalan; satu skenario
      pause/resume + satu unblock diuji.
- [ ] **Backup pertama berhasil** (lewat UI `/adm/backup` atau
      `./deploy/backup.sh`) + satu salinan disimpan DI LUAR server;
      backup cron aktif + satu restore diuji di server cadangan.
- [ ] UPS terpasang & dites; IP server statis; PostgreSQL hanya localhost.
- [ ] Akun guru/proktor/pengawas dibuat per orang (tanpa akun bersama).
      Catatan: pembuatan akun staf tambahan belum ada jalur aplikasi
      (lihat "Keterbatasan yang diketahui" di `README.md`) — siapkan
      jalurnya sebelum go-live.
- [ ] Versi aplikasi + migrasi dicatat; rencana update dibekukan H-7.
