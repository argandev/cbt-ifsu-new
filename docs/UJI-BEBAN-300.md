# UJI BEBAN 300 SISWA — CBT SMK v2 (Tim QA + Uji Beban)

**Tanggal:** 16 Okt 2026
**Program uji:** `backend/test/load/loadtest.go` (dijalankan ulang:
`go run ./test/load -siswa 300 -soal 10 -prefix QB4 -srcips 30 -actconc 30 -batch 100`)
**Lingkungan:** server scratch port **29988** + DB scratch **`cbt_v2_beban`**
(PostgreSQL 16, VM bersama — server lain tetap berjalan: CBT-Go :9988,
server tim lain :9989; angka latensi memuat kontensi VM bersama).
**Hasil akhir: LULUS** — 300/300 submit, 3000/3000 jawaban tersimpan,
0 error siswa, 0 respons 5xx, 0 kebocoran kunci jawaban.

---

## 1. Metode

| Tahap | Isi |
|---|---|
| Setup via API | 1 jurusan, 1 kelas, 1 mapel (produktif), 1 bank soal 10 soal PG (bobot 10), 1 paket PTS (durasi 60 mnt, `acak_soal=false`, `acak_jawaban=false`, `tampil_hasil=true`), 1 jadwal aktif, 1 token 6 karakter, 300 siswa via import CSV |
| Import | CSV dipecah **3 batch × 100 siswa** (alasan: BUG-02 — satu import 300 siswa > WriteTimeout 30 dtk) |
| Aktivasi | Per siswa: `ganti-password` (NISN → password baru) + login. Konkurensi 30. Pass pemulihan: bila pass 1 gagal, coba login dengan password baru dulu (menangani commit-tanpa-respons, pola BUG-02) |
| Fase ujian | **300 goroutine serentak**, tiap siswa: `mulai` (token) → `GET attempt` → 10× `jawaban` (autosave, jawaban bergantian benar/salah agar nilai bervariasi) → `submit` |
| Rate limit | Siswa disimulasikan dari **30 alamat IP loopback** (`127.0.1.1–30`, round-robin) karena di LAN sekolah tiap perangkat punya IP sendiri. Rate limit server **100 req/menit/IP** tetap berlaku per IP. Harness me-retry HANYA respons **429** (ditolak sebelum handler → aman); error transport TIDAK di-retry (request mungkin sudah commit) — kecuali pemulihan terverifikasi (login ulang / cek status) |
| Verifikasi | Hitung langsung di DB: `attempt`, `jawaban` (join attempt), attempt lengkap 10 jawaban, `nilai_akhir`; cek respons attempt tidak mengandung `"kunci"` |
| Resource | `SERVER_PID` → `/proc/<pid>/status` (VmRSS) + `/proc/<pid>/stat` (utime/stime) sebelum/sesudah; `/proc/loadavg`, `/proc/meminfo` |

## 2. Hasil (run QB4 — run final yang sah)

| Metrik | Angka |
|---|---|
| Siswa / soal per siswa / jawaban diharapkan | 300 / 10 / **3000** |
| Setup (s.d. import) | **49,2 dtk** (3 batch × ~16,4 dtk) |
| Aktivasi (300× ganti-password + login, konkurensi 30) | **97,5 dtk**, gagal 0 |
| Fase ujian (300 serentak) | **40,8 dtk** |
| TOTAL | **187,7 dtk** |
| Mulai OK | **300/300** |
| Attempt OK (GET detail) | **300** |
| Jawaban OK | **3000/3000** |
| Submit OK | **300/300** |
| Error siswa | **0** |
| Respons attempt mengandung `"kunci"` | **0** (harus 0) |
| Verifikasi DB: attempt / jawaban / attempt lengkap / nilai_akhir | **300 / 3000 / 300 / 300** |
| Jawaban hilang | **0** |
| Respons 429 `rate_limited` (di-retry harness, semua pulih) | **1800** |
| Respons 5xx | **0** |
| Timeout / error transport | **0** |
| Deadlock | Tidak teramati (semua request selesai; tidak ada request menggantung) |
| Throughput keseluruhan | **33,73 req/detik** (6331 request / 187,7 dtk) |
| Throughput fase ujian | **95,62 req/detik** (3900 request sukses / 40,8 dtk) |

### Latensi per endpoint (ms) — run QB4

| Endpoint | n | p50 | p95 | p99 | max |
|---|---|---|---|---|---|
| setup-jurusan/kelas/mapel/bank/paket/token | 1–10 | ~2–3 | — | — | ~4 |
| setup-import (per batch 100 siswa) | 3 | 16524 | 16719 | 16719 | 16719 |
| aktivasi-ganti-password | 300 | 6503 | 12860 | 14055 | 14231 |
| login | 301 | 1778 | 3254 | 4041 | 5689 |
| mulai | 300 | 1254 | 1525 | 1539 | 1551 |
| jawaban (termasuk percobaan 429 yang cepat) | 4800 | 425 | 778 | 860 | 902 |
| submit | 300 | 757 | 1016 | 1046 | 1075 |

Catatan baca tabel: `jawaban n=4800` = 3000 percobaan sukses + 1800
percobaan yang dijawab 429 (di-retry). `login n=301` = 300 siswa + 1 admin.
Latensi diukur per percobaan (tidak termasuk jeda backoff antar-retry).

### Resource server (via `/proc`, PID server scratch)

| Metrik | Sebelum | Sesudah | Delta |
|---|---|---|---|
| `load1` sistem | 2,72 | 4,34 | +1,62 |
| VmRSS proses server | 1.518.952 KB (~1,45 GB) | 1.193.948 KB (~1,14 GB) | **−325.004 KB** |
| CPU server utime (jiffies) | 30751 | 48560 | +17809 (≈178 dtk CPU / 187 dtk wall ≈ 0,95 core rata-rata) |
| CPU server stime (jiffies) | 1445 | 2640 | +1195 |
| MemAvailable sistem | 2.538.680 KB | 3.059.604 KB | +520.924 KB |

## 3. Analisis

### 3.1 Tidak ada kehilangan data / deadlock / 500 — tetapi 429 masif
Semua 300 siswa menyelesaikan ujian dan semua 3000 jawaban terverifikasi
di DB. Namun **1800 respons 429** terjadi pada fase ujian. Mekanismenya:
tiap IP sumber melayani 10 siswa × 13 request = **130 request dalam
~41 dtk**, melebihi anggaran **100 req/menit/IP** — setelah jendela
habis, semua request sisa di menit itu ditolak sampai jendela reset.
Harness pulih 100% karena me-retry 429 dengan backoff. **Frontend
produksi TIDAK me-retry 429** (`frontend/src/lib/api.ts`: nihil hasil
grep `retry`/`429`) — lihat BUG-03. Di LAN sekolah yang tiap perangkat
punya IP sendiri, beban per IP hanya ~13 request/ujian (aman). Risiko
nyata muncul bila banyak siswa berbagi satu IP (NAT sekolah, proxy,
hotspot, satu lab di belakang satu router).

### 3.2 Bottleneck #1: Argon2id mendominasi aktivasi + import
- Import 100 siswa ≈ **16,5 dtk** → ~165 ms/hash. Import 300 sekaligus
  ≈ 50 dtk → menabrak WriteTimeout 30 dtk (BUG-02).
- Aktivasi: `ganti-password` p50 **6,5 dtk** (verify lama + hash baru =
  2× Argon2id ≈ 350 ms ideal, membengkak di bawah konkurensi 30 dan
  kontensi CPU VM bersama); `login` p50 **1,8 dtk** (1× verify).
- Aktivasi 97,5 dtk = **52% total waktu uji**. Ini biaya sekali-per-siswa
  (ganti password pertama), bukan biaya tiap ujian — tetapi pada hari-H
  ujian, bila ratusan siswa baru pertama kali login, antrean aktivasi
  ini akan dirasakan sebagai "server lambat".

### 3.3 Bottleneck #2: latensi `mulai` (p50 1,25 dtk) dan `submit` (p50 0,76 dtk)
`mulai` tidak memakai Argon2id, tetapi tetap >1 dtk pada p50 di bawah
300 konkurensi. Penyebab yang teridentifikasi dari pembacaan kode
(belum dibuktikan profiling):
- Tiap request melewati `RequireAuth` (SELECT session + UPDATE
  `last_seen_at`) + `CSRF` (SELECT session lagi) + `audit.Catat`
  (INSERT `audit_log`) — mutasi ujian ≈ 3 kueri sesi/audit di luar
  kueri handler.
- Pool DB `MaxConns = 30` (`backend/cmd/server/main.go`) untuk 300
  goroutine → antrean koneksi.
- `submit` menjalankan finalisasi + hitung nilai + sinkronisasi
  `nilai_akhir`.
CPU server hanya ~0,95 core rata-rata — latensi lebih disebabkan antrean
(koneksi DB, lock baris sesi/audit) + kontensi VM bersama daripada
kejenuhan CPU proses ini.

### 3.4 RSS server ~1,1–1,5 GB — perlu profiling, bukan vonis bocor
VmRSS tercatat **turun** selama run (1,45 → 1,14 GB), jadi bukan bocor
monotonik. Tetapi angka absolutnya besar untuk server API Go pada beban
ini. Rekomendasi: jalankan `pprof` (heap + goroutine) pada beban serupa
sebelum menyatakan normal.

### 3.5 Insiden selama pengujian (jujur dicatat)
1. **Run QB1** — harness versi awal me-retry error transport pada import
   non-idempotent → import tereksekusi 2× (yang kedua: 300x "NISN sudah
   terdaftar"). Harness diperbaiki (retry HANYA 429) — lihat §1.
2. **Run QB2** — import 300 sekaligus: klien `EOF` setelah ~52 dtk,
   server mencatat 200 + commit 300 siswa. Akar: BUG-02 (WriteTimeout).
   Diperbaiki dengan import per batch 100.
3. **Run QB3** — 7 dari 300 aktivasi gagal di sisi klien (pass 1),
   sementara log server mencatat **303/303 ganti-password status 200**
   dan nol 429. Penyebab pasti **tidak teridentifikasi** (diduga race
   penutupan koneksi keep-alive di sisi klien; tidak terulang pada QB4).
   Ditangani dengan pass-2 pemulihan (login dengan password baru) —
   pada QB4 aktivasi gagal 0. Bila terulang di produksi, kumpulkan
   pesan error klien + `tcpdump`/log koneksi sebelum menyimpulkan.

## 4. Rekomendasi tuning

| # | Rekomendasi | Alasan / target |
|---|---|---|
| 1 | **Naikkan / pisahkan timeout HTTP** (`main.go`: `ReadTimeout 60 dtk`, `WriteTimeout 120 dtk`) atau jadikan import pekerjaan latar + polling progres | BUG-02: tidak ada lagi commit-tanpa-respons pada import besar |
| 2 | **Rate limit per-sesi/pengguna, bukan hanya per-IP**, untuk endpoint ujian; atau naikkan batas untuk `/api/v1/ujian/*`; kembalikan header `Retry-After` pada 429 | 1800× 429 pada 30 IP bersama; NAT sekolah adalah kasus nyata |
| 3 | **Retry-with-backoff di frontend `apiClient`** untuk 429 (dan 5xx) pada `jawaban` (upsert → aman) + verifikasi status setelah `submit` gagal | Tanpa ini, satu 429 di jaringan NAT = jawaban berisiko hilang |
| 4 | **Buat import idempotent** (Idempotency-Key / dedupe berbasis hash isi file) + batas batch per request (mis. 200 baris) | Mencegah eksekusi ganda saat klien retry |
| 5 | **Paralelkan hash Argon2id import** (worker pool — tiap baris independen) dan/atau tunda pembuatan akun login siswa ke aktivasi pertama | Import 300 siswa 50 dtk → <10 dtk; aktivasi 97 dtk turun drastis bila hash tidak diulang |
| 6 | **Kurangi kueri sesi per request**: gabungkan cek sesi CSRF + auth (1 SELECT, bukan 2); `last_seen_at` perbarui secara sampled/batched, bukan tiap request | Mengurangi antrean pada 300 konkurensi (mulai p50 1,25 dtk) |
| 7 | **Naikkan `MaxConns` pool DB** (30 → 50–100, sesuaikan `max_connections` Postgres) + indeks: pastikan indeks pada `sessions(user_id)`, `jawaban(attempt_id, soal_id)`, `attempt(jadwal_id, siswa_id)`, `audit_log` sesuai pola kueri | Antrean koneksi pada 300 goroutine |
| 8 | **Profiling `pprof`** (heap, goroutine, mutex) pada beban 300 siswa | RSS 1,1–1,5 GB belum dijelaskan; verifikasi sebelum produksi |
| 9 | **Dokumentasikan kapasitas teruji**: 300 siswa serentak LULUS pada VM bersama ini (95,6 req/dtk fase ujian). Uji ulang 600–1000 siswa di server sekolah yang sebenarnya sebelum hari-H, dengan jaringan LAN (bukan loopback) | Angka ini bukan jaminan untuk 1000+ siswa / hardware berbeda |

## 5. Cara mengulang

```bash
# Terminal 1 — server scratch (JANGAN server tim lain)
DATABASE_URL=postgres://<user>:<pass>@127.0.0.1:5432/cbt_v2_beban \
  PORT=29988 go run ./cmd/server serve
psql "$DATABASE_URL" -f backend/db/seed/seed.sql
go run ./cmd/server create-admin --email <ADMIN_LOGIN> \
  --password <ADMIN_INITIAL_PASSWORD> --nama "Admin QA"

# Terminal 2 — uji beban
BASE_URL=http://127.0.0.1:29988 ADMIN_LOGIN=... \
  ADMIN_INITIAL_PASSWORD=... ADMIN_PASSWORD=... \
  LOAD_DATABASE_URL=postgres://.../cbt_v2_beban \
  SERVER_PID=<pid server scratch> \
  go run ./test/load -siswa 300 -soal 10 [-batch 100] [-srcips 30]

# Bersihkan tuntas: DROP DATABASE cbt_v2_beban;
# (cleanup API best-effort; guard 409 melarang hapus data ber-attempt)
```
