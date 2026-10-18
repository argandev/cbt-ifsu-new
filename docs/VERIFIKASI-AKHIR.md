# VERIFIKASI AKHIR — CBT SMK v2 (Tim Fase 3)

**Tanggal:** 16 Okt 2026
**Pelaksana:** Tim Fase 3 — Verifikasi Akhir + Perbaikan Sisa (tim integrasi terakhir)
**Lingkungan verifikasi:** server scratch port **29988** + DB scratch **`cbt_v2_verif`**
(port 9988 CBT-Go dan 9989 server tim lain tidak disentuh), plus binary final
`make build` diuji di port **29989** + DB **`cbt_v2_final`**.
**Status: VERIFIKASI SELESAI — semua pemeriksaan 1–9 lulus.**

> Catatan lingkungan: di tengah pekerjaan, VM/sandbox ter-reset (Postgres kembali
> ke cluster kosong, proses server scratch mati, `/tmp` ter-prune sebagian).
> Seluruh rangkaian verifikasi diulang dari nol di lingkungan fresh setelah
> reset — angka di dokumen ini adalah hasil run ULANG tersebut (E2E prefix
> E2EFIN4, beban prefix QVFIN2, live prefix VF*, capture prefix CAP*), kecuali
> dinyatakan lain. Perbaikan kode (di `~/workspace`) selamat dari reset.

---

## 1. BUG `ujianAktif` (KRITIS) — DIPERBAIKI + TERVERIFIKASI LIVE ✅

**Temuan awal (benar):** `backend/internal/exam/exam.go` fungsi `ujianAktif`
memilih **14 kolom** (`j.id, j.paket_id, j.tanggal, j.jam_mulai, j.jam_selesai,
j.status, p.kode, p.nama, p.jenis, p.durasi_menit, p.tampil_hasil,
COALESCE(k.nama,''), a.id, a.status`) tetapi `Scan` hanya **13 tujuan**
(`p.tampil_hasil` tidak di-scan). Akibatnya: bila ADA jadwal hari ini,
pgx mengembalikan error jumlah tujuan dan handler menjawab **500**;
bila TIDAK ADA jadwal, loop tidak jalan sehingga bug tersembunyi (200 kosong).

**Perbaikan** (`backend/internal/exam/exam.go`, `ujianAktif`):
- Tambah field `TampilHasil bool \`json:"tampil_hasil"\`` ke struct `item`.
- Tambah `&it.TampilHasil` ke `rows.Scan` pada posisi yang sesuai urutan SELECT
  (setelah `DurasiMenit`, sebelum `KelasNama`). Jumlah kolom = tujuan Scan = 14.

**Verifikasi live** (4 siswa, jadwal aktif hari ini): `GET /api/v1/ujian/aktif`
→ **200** untuk keempat siswa, item jadwal benar, `tampil_hasil: true`,
`jadwal_id` terisi (tidak ada `/ujian/undefined` di frontend — frontend memakai
`item.jadwal_id`). Frontend `SiswaHomePage` sudah memakai field persis respons.

## 2. Kartu siswa lookup — DIPERBAIKI + TERVERIFIKASI LIVE ✅

**Temuan awal (benar):** `GET /api/v1/kartu/:jadwalId/:siswaId` — frontend
`KartuSayaPage` mengirim `siswaId = user.id` (ID akun login), tetapi backend
`loadSiswaKartu` (`backend/internal/event/event.go`) hanya mencari
`WHERE id = $1` (siswa.id) → kartu sendiri selalu **404**.

**Perbaikan** (`backend/internal/event/event.go`, `loadSiswaKartu`):
cari `siswa.id` dulu; bila tidak ketemu, fallback `siswa.user_id = $1`
(pola sama seperti `kelasSiswa` di `exam.go`). Cek kepemilikan siswa tetap
berlaku (`s.UserID == uid` atau `s.ID == uid`, selain itu 404 anti-IDOR).

**Verifikasi live** (4 siswa): kartu via `user.id` → **200** (NISN cocok),
kartu via `siswa.id` → **200** (regresi aman).

## 3. Daya beda numerik — TERVERIFIKASI LIVE ✅

Skenario: 1 jadwal, 1 soal PG (kunci B), **4 peserta**: 2 jawab benar (B),
2 jawab salah (A). `GET /api/v1/analisis/:jadwalId` →

| Metrik | Nilai |
|---|---|
| jumlah_peserta | 4 |
| jumlah_benar | 2 |
| kesukaran | **0.5** (Sedang) |
| **daya_beda** | **1** (numerik, bukan null; interpretasi "Sangat Baik") |

Hitungan tangan: kelompok atas (median split 50%) 2/2 benar = 1.00;
kelompok bawah 0/2 benar = 0.00; daya beda = 1.00 − 0.00 = **1.00** ✅ cocok.
Pengecoh: A 2 (50%), B 2 (50%), C/D 0 (tidak berfungsi).

## 4. Pembahasan end-to-end — TERVERIFIKASI LIVE (4/4 langkah) ✅

1. **POST soal dengan `pembahasan`** → 201.
2. **GET soal guru** (`GET /api/v1/bank-soal/:bankId/soal`) → mengembalikan
   `pembahasan` persis (`PEMBAHASAN-VF1: Jakarta adalah ibukota ...`) ✅.
3. **Attempt sebelum submit** (`GET /api/v1/ujian/attempt/:id`) → bebas kata
   `pembahasan` dan bebas `"kunci"` ✅ (4/4 siswa).
4. **Hasil setelah submit + tampil_hasil** (`GET .../hasil`, paket
   `tampil_hasil=true` + settings `tampil_hasil_ke_siswa=true`) →
   `pembahasan` ada ✅ (4/4 siswa; nilai benar 100, salah 0).

## 5. Re-run uji beban — LULUS ✅

`go run ./test/load -siswa 300 -soal 10 -prefix QVFIN2 -srcips 30 -actconc 30 -batch 100`
(server scratch 29988, DB scratch):

| Metrik | Run verifikasi akhir (QVFIN2) | Run QB4 (Tim QA, 16 Okt 2026) |
|---|---|---|
| Mulai OK | 300/300 | 300/300 |
| Jawaban OK | **3000/3000** | 3000/3000 |
| Submit OK | 300/300 | 300/300 |
| Error siswa | 0 | 0 |
| Respons attempt mengandung `"kunci"` | 0 | 0 |
| Verifikasi DB (attempt/jawaban/lengkap/nilai_akhir) | 300 / 3000 / 300 / 300 | 300 / 3000 / 300 / 300 |
| Respons **429** (di-retry) | **0** | 1800 |
| Respons 5xx / timeout | **0 / 0** | 0 / 0 |
| Import siswa | 300/300 (3 batch × 100, tanpa EOF) | 300/300 (3 batch × 100) |
| Throughput keseluruhan | 26,45 req/dtk (4531 req / 171,3 dtk) | 33,73 req/dtk (6331 req / 187,7 dtk) |
| **Throughput fase ujian** | **510,49 req/dtk** | **95,62 req/dtk** |
| Latensi jawaban p50 / p95 | 529 / 782 ms | 425 / 778 ms |
| Latensi mulai p50 | 613 ms | 1254 ms |
| Latensi submit p50 | 954 ms | 757 ms |

**BUG-02 beres:** import 300 siswa (3×100) selesai tanpa EOF; timeout server
eksplisit `ReadTimeout 60 dtk / WriteTimeout 120 dtk / IdleTimeout 120 dtk`
(`backend/cmd/server/main.go`). **Import kedua idempotent** (diuji live
terpisah): import-1 `{sukses: 3, gagal: 0, sudah_ada: 0}` → import-2 file sama
`{sukses: 0, gagal: 0, sudah_ada: 3}` ✅.
**BUG-03 beres:** 0× 429 pada run ini (vs 1800× pada QB4) — limit endpoint
ujian 300 req/menit/IP + retry 429 di `apiClient` berlaku; tidak ada jawaban
hilang (3000/3000 terverifikasi di DB). Throughput fase ujian naik ~5,3×
karena tidak ada lagi backoff retry 429.
Catatan jujur: statistik resource (`/proc`, VmRSS/CPU) pada run ini TIDAK VALID
(SERVER_PID tidak diisi harness) — tidak dikutip.

## 6. Re-run E2E — 33/33 PASS ✅

`go run ./test/e2e -prefix E2EFIN4` → **PASS 33 / 33 langkah**,
"E2E: SEMUA LANGKAH PASS" (nilai akhir terverifikasi 90: PG benar 10 poin +
uraian skor 80 → 8 poin).

Dua gangguan selama re-run yang BUKAN bug aplikasi (dicatat jujur):
- Run pertama gagal langkah 27 karena skrip verifikasi live (§4) telah mengubah
  settings global `tampil_hasil_ke_siswa=true`, sementara E2E langkah 27
  mengasumsikan settings default `false`. Setelah settings dikembalikan ke
  `false`, E2E 33/33 PASS. (Settings bersifat global — rangkaian uji yang
  mengubahnya harus mengembalikannya.)
- Satu run gagal langkah 05 dengan 429 `rate_limited` karena jendela
  100 req/menit/IP habis akibat re-run cepat berturut-turut dari IP yang sama;
  setelah menunggu ±70 dtk (reset jendela), run berikutnya 33/33 PASS.
  Harness E2E tidak me-retry 429 (fail-fast) — perilaku harness, bukan server.

## 7. Pembuatan akun staf — VERIFIKASI (tanpa modul baru) ✅

**Temuan — TIDAK ADA jalur pembuatan akun staf** (dicatat sebagai
KETERBATASAN YANG DIKETAHUI, bukan dibangun — di luar cakupan):
- CLI `create-admin` (`backend/cmd/server/main.go` → `auth.CreateAdmin`):
  flag hanya `--email/--username/--nama/--password`, **tanpa flag `--role`**,
  role di-hardcode `'admin'`, dan HANYA bila tabel `users` kosong
  (bila tidak: error `tabel users sudah berisi akun`).
- **Tidak ada** endpoint `POST /api/v1/users` (grep: nihil) dan **tidak ada**
  UI pembuatan akun staf (grep: nihil).
- Yang ADA: akun **siswa** dibuat otomatis via master siswa (create/import,
  username = NISN, password awal = NISN); `reset-password` CLI bisa me-reset
  password akun yang sudah ada (berdasarkan email/username); halaman Guru
  punya field "UUID akun pengguna" untuk MENGKAITKAN akun yang sudah ada ke
  record guru (bukan membuat akun).
- **Implikasi operasional:** akun staf (proktor/guru/pengawas/penguji_ukk)
  dibuat sekolah via **SQL langsung ke tabel `users`** (INSERT dengan
  `password_hash` Argon2id + `role` yang sesuai + `must_change_password=true`),
  atau dengan memakai `create-admin` saat DB masih kosong lalu mengubah
  `role` via SQL. Modul manajemen user TIDAK dibangun (sesuai instruksi).

## 8. Screenshot semua halaman utama ✅

Metode: headless Chrome `/opt/meta-chromium/chrome` + CDP. **Kendala sandbox:**
Chrome di sandbox ini TIDAK punya akses jaringan sama sekali (navigasi ke
localhost maupun situs publik diblokir `ERR_BLOCKED_BY_LOCAL_NETWORK_ACCESS_CHECKS`;
flag `--disable-features=...` + policy `LocalNetworkAccessAllowedForUrls` tidak
mengubahnya). Yang jalan: `file://`. Maka screenshot memakai **file:// + aset
inline** (bundle Vite hasil `npm run build` + CSS + font di-inline) dengan
**mock `fetch` berisi respons API REAL hasil capture dari server scratch**
(`window.fetch` di-override sebelum bundle jalan; route diatur via
`history.replaceState`; aksi isi form jadwal + pilih jadwal dilakukan via CDP).
UI yang dirender adalah 100% kode produksi; hanya transport yang di-stub karena
sandbox memblokir jaringan. Halaman diverifikasi tidak blank/error via cuplikan
teks DOM + inspeksi visual 5 halaman kunci.

Nomor dilanjutkan dari 07 (file 01–07 milik Fase 1 dipertahankan):

| File | Halaman | Isi terverifikasi |
|---|---|---|
| `screenshots/08-login.png` | Login | Form email/username + password |
| `screenshots/09-dashboard-admin.png` | Dashboard admin | Menu + kartu stat |
| `screenshots/10-master-jurusan.png` | Master Jurusan | Daftar jurusan |
| `screenshots/11-bank-soal.png` | Bank Soal | Daftar bank soal |
| `screenshots/12-paket.png` | Paket | Daftar paket |
| `screenshots/13-jadwal.png` | Jadwal | Daftar jadwal |
| `screenshots/14-ujian-siswa.png` | Halaman Ujian siswa | Timer 01:59:54, Soal 1 + opsi A–D |
| `screenshots/15-monitor.png` | Monitor (per jadwal, `/monitor/:jadwalId`) | 5 peserta (1 berjalan, 4 selesai), tombol Pause/Resume |
| `screenshots/16-rekap.png` | Rekap | Rekap per jadwal |
| `screenshots/17-berita-acara.png` | Berita Acara | Dokumen lengkap (mapel/kelas/tanggal/jam/peserta 5/hadir 4/rata-rata 50.00/lulus 2) |
| `screenshots/18-ukk-list.png` | UKK (list) | Daftar UKK |
| `screenshots/19-ukk-sertifikat.png` | UKK sertifikat | Sertifikat UKK siswa |
| `screenshots/20-analisis-butir.png` | Analisis Butir | Kesukaran 0.50, daya beda 1.00, pengecoh per opsi |
| `screenshots/21-kompetensi.png` | Kompetensi (nilai per kompetensi) | Nilai per kompetensi |
| `screenshots/22-event.png` | Event | Daftar event |
| `screenshots/23-ruangan.png` | Ruangan | Daftar ruangan |
| `screenshots/24-kartu-peserta.png` | Kartu Peserta (admin) | 5 kartu, nomor US/2026/001–004, penugasan event/ruangan |
| `screenshots/25-backup.png` | Backup | Daftar backup + aksi |
| `screenshots/26-settings.png` | Settings | Form settings |
| `screenshots/27-siswa-home.png` | SiswaHome (`/`) | Ujian Hari Ini + tombol Lanjut/Kartu Saya |
| `screenshots/28-kartu-saya.png` | Kartu Saya (`/kartu/:jadwalId`) | Kartu peserta milik siswa |

Catatan: tidak ada halaman khusus Notifikasi — notifikasi adalah **bel di header
admin** (`NotifikasiBell`, polling `GET /api/v1/notifikasi`, badge belum dibaca),
yang terlihat di semua screenshot admin. Monitor `/adm/monitor` tanpa jadwal
menampilkan petunjuk "Buka monitor dari halaman jadwal" (lihat §SISA/bug
diperbaiki untuk rute per-jadwal).

## 9. Build final ✅

| Perintah | Hasil |
|---|---|
| `go build ./...` | ✅ OK |
| `go vet ./...` | ✅ OK (tanpa temuan) |
| `go test ./...` | ✅ **12 paket OK, 71 fungsi tes PASS, 0 FAIL**, 2 SKIP by-design (`TestE2E`, `TestMasterLive` — butuh `CBT_TEST_DATABASE_URL`, dilewati bila tidak diisi) |
| `npx tsc --noEmit` | ✅ OK |
| `npm run build` | ✅ OK (bundle `index-BXJ0jWrI.js` 603 KB) |
| `make build` | ✅ OK — binary `cbt-smk-v2` (19,8 MB), `backend/dist` tersalin dari `frontend/dist` |
| Binary `make build` | ✅ jalan: `GET /` → 200, `GET /adm` → 200, `GET /login` → 200 (SPA) |

## Bug yang diperbaiki Tim Fase 3 (4 bug)

| # | Bug | File | Perbaikan |
|---|---|---|---|
| 1 | `ujianAktif`: SELECT 14 kolom, Scan 13 tujuan → 500 bila ada jadwal hari ini | `backend/internal/exam/exam.go` | Tambah field `TampilHasil` + `&it.TampilHasil` ke Scan (14 = 14) |
| 2 | Kartu siswa: `loadSiswaKartu` hanya `WHERE id` → 404 untuk `siswaId = user.id` | `backend/internal/event/event.go` | Fallback `siswa.user_id` bila `siswa.id` tidak ketemu |
| 3 | **Berita Acara crash (ditemukan via screenshot):** backend mengirim `jam_mulai`/`jam_selesai` sebagai objek pgtype `{"Microseconds":…,"Valid":…}` → React ErrorBoundary ("Objects are not valid as a React child") | `backend/internal/rekap/rekap.go` | SELECT memakai `j.tanggal::text, j.jam_mulai::text, j.jam_selesai::text` → string (`"19:37:39"`, `"2026-10-16"`); diverifikasi live + screenshot ulang |
| 4 | **Monitor per-jadwal tidak terjangkau (ditemukan via screenshot):** `JadwalPage` menaut ke `/monitor/:jadwalId`, tetapi tidak ada route yang cocok → NotFound | `frontend/src/App.tsx` | Tambah route `adm/monitor/:jadwalId`, `/monitor/:jadwalId`, `/pengawas/monitor/:jadwalId` → `MonitorPage`; diverifikasi screenshot monitor berdata (5 peserta) |

Tidak ada perubahan arsitektur. Tidak ada copy-paste dari `cbt-go`.
Tidak ada mekanisme lisensi Extraordinary yang disalin.

## SISA (jujur)

1. **Keterbatasan diketahui — pembuatan akun staf** (lihat §7): sekolah membuat
   akun proktor/guru/pengawas/penguji_ukk via SQL / CLI `create-admin` (saat
   users kosong) + ubah `role` via SQL. Modul manajemen user belum ada.
2. **`/adm/monitor` tanpa jadwal** menampilkan petunjuk "Buka monitor dari
   halaman jadwal" — perilaku by-design; monitor berdata diakses via
   `/monitor/:jadwalId` (dari JadwalPage) atau `/adm/monitor/:jadwalId`.
3. **E2E sensitif terhadap settings global + rate limit** (lihat §6): rangkaian
   uji yang mengubah `tampil_hasil_ke_siswa` harus mengembalikannya; re-run
   cepat dari satu IP bisa kena 429 (harness E2E fail-fast, bukan bug server).
4. **Screenshot memakai mock fetch** (bukan jaringan live) karena Chrome di
   sandbox ini diblokir total untuk jaringan (lihat §8) — transport di-stub,
   UI + data 100% kode/respons produksi. Bila perlu bukti jaringan-live,
   ambil screenshot dari perangkat yang bisa mencapai server sekolah.
5. **Uji beban 600–1000 siswa** di server sekolah yang sebenarnya (dengan
   jaringan LAN, bukan loopback) tetap disarankan sebelum hari-H, sesuai
   rekomendasi Tim QA (angka 300 siswa di sini bukan jaminan untuk 1000+).
6. **Pembersihan:** DB/server scratch di-drop/dimatikan setelah verifikasi
   (lihat bawah); skrip verifikasi sementara dihapus; file policy Chrome
   sementara di `/etc` dihapus.

---

**Bersihkan DB/server scratch setelah selesai:** server scratch port 29988
& 29989 dimatikan, database scratch `cbt_v2_verif` & `cbt_v2_final` di-DROP,
role scratch `root` di-DROP, direktori skrip sementara
`.scratch-bin/` dihapus, file policy Chrome sementara dihapus.
