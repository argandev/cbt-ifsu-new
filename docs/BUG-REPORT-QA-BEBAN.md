# BUG REPORT — QA + Uji Beban (Tim Tambahan 4)

**Proyek:** CBT SMK v2 — `~/workspace/cbt-smk-v2`
**Tanggal:** 16 Okt 2026
**Penguji:** Tim Tambahan 4 (QA + Uji Beban)
**Lingkungan uji:** server scratch port **29988** + DB scratch **`cbt_v2_beban`**
(BUKAN server/DB tim lain; port 9988 CBT-Go dan 9989 server tim lain tidak disentuh).
**Cara reproduksi umum:** `go run ./test/e2e` (33 langkah) dan
`go run ./test/load -siswa 300 -soal 10` — lihat `docs/UJI-BEBAN-300.md`.
**Aturan:** file ini hanya MELAPORKAN. Tidak ada file sumber aplikasi
yang diubah oleh tim ini.

---

## BUG-01 (SEDANG) — Permission `analisis.*`: 4 sumber kebenaran tidak sinkron; grant migrasi 0008 diam-diam tidak jalan

**Judul:** `analisis.read` / `analisis.export` ada di DB (61 permission) tapi
tidak ada di `KONTRAK.md` (59) dan `rbac.go` (59); grant migrasi 0008 ke
proktor+guru tidak pernah berlaku pada alur standar (migrate → seed).

**File:baris:**
- `backend/migrations/0008_analisis.up.sql:8-22` — INSERT permission +
  grant `role_permissions` ke `admin, proktor, guru`.
- `backend/db/seed/seed.sql:136-138` — admin: `CROSS JOIN permissions`
  (dinamis = SEMUA permission yang ada di tabel).
- `backend/db/seed/seed.sql:144-205` — proktor/guru/pengawas/siswa/
  penguji_ukk: daftar tetap (fixed list) TANPA `analisis.*`.
- `backend/internal/rbac/rbac.go:30-63` — `AllPermissions` = 59, tanpa
  `analisis.*`.
- `docs/KONTRAK.md:178` — "Total **59 permission**".
- `backend/internal/analisis/analisis.go:860,882` — endpoint menegakkan
  `analisis.read` / `analisis.export`.
- Koroborasi sisi frontend: `docs/AUDIT-UI.md` butir C.2 (tim lain) —
  `hasPermission("analisis.read")` selalu `false`.

**Langkah reproduksi:**
1. DB kosong → jalankan server scratch (migrasi auto-up, termasuk 0008)
   → jalankan `make seed` (alur standar `migrate-up && seed`).
2. `SELECT count(*) FROM permissions;` → **61** (59 seed + 2 dari 0008).
3. `SELECT r.code, count(*) FROM role_permissions rp JOIN roles r ...` →
   admin **61**, proktor **28**, guru **25** (angka kontrak, TANPA analisis).
4. Login sebagai proktor/guru → `GET /api/v1/analisis/<jadwalId>` → **403**.

**Bukti:**
- `permissions` = 61 baris; `analisis.read` + `analisis.export` hanya
  berasal dari `0008_analisis.up.sql:9-10` (grep `analisis` di
  `seed.sql` = nihil).
- Matriks hasil seed: `admin 61, guru 25, pengawas 4, penguji_ukk 4,
  proktor 28, siswa 2`.
- Akar masalah: saat migrasi 0008 berjalan, tabel `roles` masih KOSONG
  (roles diisi oleh seed, BUKAN migrasi). Akibatnya
  `INSERT INTO role_permissions SELECT ... FROM roles r JOIN permissions p
  ... WHERE r.code IN ('admin','proktor','guru')` menghasilkan **0 baris**
  tanpa error — grant yang dijanjikan komentar "Grant: admin + proktor +
  guru" diam-diam tidak terjadi untuk proktor/guru. Admin tetap dapat
  kedua permission hanya karena grant seed bersifat dinamis ("semua").
- `rbac.go` (matriks statis fallback, dipakai bila DB belum di-seed)
  juga tidak memuat `analisis.*`, sehingga pada jalur fallback BAHKAN
  admin ditolak 403 di endpoint analisis.

**Dampak:**
- Proktor & guru TIDAK BISA membuka Analisis Butir Soal / Nilai per
  Kompetensi (403) padahal migrasi 0008 dirancang memberi mereka akses.
- Kontrak (§5 "SATU-SATUNYA sumber kebenaran") dilanggar: 3 sumber
  (DB/seed-dinamis, migrasi, rbac) masing-masing bilang lain.
- Frontend menyembunyikan menu Analisis untuk non-admin (AUDIT-UI C.2).

**Saran perbaikan (pilih SATU, lalu selaraskan ke-4 sumber):**
- Opsi A (bila proktor+guru memang boleh melihat analisis): tambahkan
  `analisis.read`, `analisis.export` ke `KONTRAK.md` §5 (total menjadi
  **61**), ke `AllPermissions` + matriks `rbac.go`, dan ke daftar tetap
  proktor/guru di `seed.sql`; lalu buat migrasi baru (mis. 0012) yang
  meng-grant kedua permission ke proktor+guru (migrasi 0008 tidak bisa
  diandalkan karena urutan migrate-seed), atau dokumentasikan bahwa
  seed harus dijalankan ulang setelah 0008.
- Opsi B (bila analisis hanya untuk admin): ubah komentar 0008
  (hapus klaim grant proktor+guru), tambahkan `analisis.*` ke kontrak
  sebagai admin-only, dan selaraskan gating frontend.
- Apa pun opsinya: JANGAN biarkan migrasi grant yang bergantung pada
  tabel `roles` yang belum di-seed tanpa pemeriksaan (mis. `RAISE
  WARNING` bila 0 baris ter-grant).

---

## BUG-02 (TINGGI) — Import siswa besar: klien menerima EOF (WriteTimeout 30 dtk Echo), padahal server commit — import "gagal" yang sebenarnya berhasil

**Judul:** `POST /api/v1/master/siswa/import` 300 baris memakan ~50 dtk
(300x hash Argon2id sekuensial) > `WriteTimeout` 30 dtk Echo v5, sehingga
server menutup koneksi; klien mendapat `EOF` dan mengira import gagal,
padahal transaksi SUDAH commit. Retry buta → import kedua melaporkan
300x "NISN sudah terdaftar".

**File:baris:**
- `backend/cmd/server/main.go` (fungsi `runServer`, pemanggilan
  `e.Start(cfg.Addr)`) — memakai timeout default Echo, TIDAK ada
  override `ReadTimeout`/`WriteTimeout`.
- `backend/internal/master/master.go` (handler `importSiswa`, sekitar
  baris 1476-1730) — satu transaksi + hash Argon2id per baris
  sekuensial (`createStudentAccountTx`), tanpa batas ukuran batch /
  tanpa idempotency.
- Penyebab di dependensi: `echo/v5@v5.0.0/server.go:106-107` —
  `ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second`
  (default `StartConfig.start`, demi Gosec G112).

**Langkah reproduksi:**
1. Server scratch port 29988 + DB scratch (lihat kepala dokumen ini).
2. Import CSV 300 siswa lewat API (atau UI Import Siswa).
3. Klien menunggu ~50 dtk lalu menerima error transport:
   `Post "http://127.0.0.1:29988/api/v1/master/siswa/import": EOF`.
4. Cek DB: 300 siswa + 300 akun `users` SUDAH ADA (commit terjadi).
5. Kirim import yang sama lagi → respons 200 `sukses=0 gagal=300`,
   semua baris "NISN sudah terdaftar."

**Bukti (16 Okt 2026, server scratch port 29988):**
- Log server, run QB1: `POST /api/v1/master/siswa/import status 200
  latency 52.717s`, lalu 0,5 dtk kemudian request kedua (retry klien)
  `status 200 latency 21ms` dengan hasil `sukses=0 gagal=300`.
- Log server, run QB2 (tanpa retry klien): `POST .../import status 200
  latency 49.636s`, tetapi klien menerima `EOF`; cek DB:
  `SELECT count(*) FROM siswa WHERE nisn LIKE '98%'` = **600**
  (300 QB1 + 300 QB2) — kedua import ter-commit walau klien QB2
  hanya melihat EOF.
- Per-siswa: ~175 ms/hash Argon2id (52,7 dtk / 300). Batch 100 siswa
  ≈ 16,5 dtk < 30 dtk (terverifikasi lolos pada run QB3 dan QB4,
  masing-masing 3 batch × 100 sukses 100/100 — lihat
  `docs/UJI-BEBAN-300.md`).

**Dampak:**
- Operator/admin mengira import besar GAGAL (UI menampilkan error),
  padahal data sudah masuk. Bila ia mengulang import, ia mendapat
  300x "NISN sudah terdaftar" dan tidak tahu data mana yang sudah masuk.
- Pola yang sama mengancam endpoint lambat lain (>30 dtk): backup/restore
  besar, export rekap besar, dsb. — respons hilang walau pekerjaan
  selesai.
- Klien yang me-retry otomatis request non-idempotent (seperti harness
  uji beban versi awal) akan mengeksekusi ulang import.

**Saran perbaikan:**
1. (Wajib) Di `main.go`, set timeout server secara eksplisit sebelum
   `e.Start`, mis. `e.Server.ReadTimeout = 60s`,
   `e.Server.WriteTimeout = 120s`, `e.Server.IdleTimeout = 120s`
   — atau jalankan import sebagai pekerjaan latar (job + polling
   progres) agar tidak terikat timeout HTTP sama sekali.
2. Buat import idempotent terhadap retry: terima `Idempotency-Key`
   (atau tolak duplikat berbasis hash isi file + dedupe per baris
   yang sudah benar — saat ini baris duplikat dilaporkan gagal,
   sehingga retry total selalu "gagal").
3. Pertimbangkan batas batch per request di server (mis. maks 200
   baris) dengan pesan jelas, atau paralelkan hash password
   (Argon2id per baris independen — worker pool) agar import 300
   baris jauh di bawah 30 dtk.
4. Workaround sampai diperbaiki: import maksimal **100 siswa per file**
   (≈17 dtk), lalu ulangi file berikutnya.

---

## BUG-03 (SEDANG) — Rate limit 100 req/menit/IP + frontend tanpa retry 429: jawaban siswa berisiko hilang bila banyak siswa berbagi satu IP (NAT/proxy)

**Judul:** Pada beban 300 siswa yang berbagi 30 IP, server menjawab
**1800× 429 `rate_limited`**; harness uji pulih 100% karena me-retry 429,
tetapi **frontend produksi tidak me-retry 429 sama sekali** — satu 429
pada `POST jawaban`/`submit` di jaringan NAT berarti jawaban gagal
tersimpan tanpa pemulihan otomatis.

**File:baris:**
- `backend/internal/middleware/middleware.go:156-159` —
  `RateLimit() = NewRateLimiter(100, time.Minute)` (per-IP, fixed window,
  tanpa header `Retry-After`).
- `frontend/src/lib/api.ts` — grep `retry`/`429` = **nihil** (tidak ada
  retry, tidak ada penanganan khusus 429).
- `backend/internal/middleware/middleware.go:126-154` — penolakan 429
  terjadi di middleware, sebelum handler (retry 429 aman secara
  semantik — klienlah yang tidak melakukannya).

**Langkah reproduksi:**
1. Jalankan `go run ./test/load -siswa 300 -soal 10 -srcips 30`
   (30 IP → 10 siswa/IP × 13 request = 130 request/IP dalam ~41 dtk >
   anggaran 100/menit).
2. Amati: **1800 respons 429** pada fase ujian; dengan retry harness
   semua 300 siswa lulus (3000/3000 jawaban).
3. Tanpa retry (perilaku frontend saat ini): request jawaban yang
   terkena 429 gagal; siswa harus menekan simpan ulang secara manual —
   bila ia sudah pindah soal / submit, jawaban itu hilang.

**Bukti (run QB4, 16 Okt 2026):**
- `Total request: 6331 | respons 429 (di-retry): 1800 | 5xx: 0`.
- `jawaban n=4800` = 3000 sukses + 1800 percobaan 429.
- Server tidak pernah mencatat 429 sebagai error handler (wajar —
  ditolak middleware); semua 300 siswa tetap lulus HANYA karena harness
  retry.
- Di LAN sekolah yang tiap perangkat punya IP sendiri, beban per IP
  hanya ~13 request/ujian → di bawah batas (aman). Risiko nyata: NAT /
  proxy / hotspot / satu lab di belakang satu router.

**Dampak:**
- Skenario terburuk: jawaban siswa hilang diam-diam pada jaringan
  berbagi IP; proktor melihat attempt "tidak lengkap" tanpa jejak
  bahwa penyebabnya 429.
- Fixed window memperburuk: setelah IP mencapai 100 dalam satu menit,
  SEMUA request sisa di menit itu ditolak (bukan sekadar diperlambat).

**Saran perbaikan (pilih kombinasi):**
1. Frontend `apiClient`: retry-with-backoff (mis. 3×: 300 ms, 800 ms,
   2 dtk) untuk 429 (dan 5xx) pada request yang aman diulang —
   `jawaban` adalah upsert (aman); `submit` yang gagal transport/429
   harus diikuti `GET .../status` untuk verifikasi (pola yang dipakai
   harness QB4), bukan submit buta ulang.
2. Server: kembalikan header `Retry-After` pada respons 429 agar klien
   tahu kapan mencoba lagi.
3. Server: rate limit berbasis sesi/pengguna (`session_id`/`user_id`)
   untuk `/api/v1/ujian/*`, atau anggaran lebih besar untuk endpoint
   ujian (mis. 300/menit/IP), karena pola akses ujian adalah burst
   singkat yang sah, bukan serangan.
4. Pertimbangkan sliding-window / token-bucket sebagai pengganti fixed
   window agar tidak ada "tebing" penolakan total di akhir menit.

---

## Bukan bug (dugaan yang diperiksa dan TERBANTAH)

### NB-01 — Hasil ujian tersembunyi walau paket `tampil_hasil=true`
**Hasil:** BUKAN bug — perilaku sesuai desain.
Bukti: `backend/internal/exam/exam.go:207-212`
`tampilHasilEfektif = paket.tampil_hasil AND settings.tampil_hasil_ke_siswa`
(default settings `false`, lihat `backend/db/seed/seed.sql:214`).
E2E awal mengira ini bug (langkah 27 gagal); setelah membaca kode,
program E2E diperbaiki: menguji hasil TERSEMBUNYI pada settings default,
lalu mengaktifkan settings `tampil_hasil_ke_siswa=true` via
`PUT /api/v1/settings` dan menguji hasil tampil nilai 90 — keduanya PASS.
Kontrak §8.6 menegaskan nilai disembunyikan sampai final
(`tampil_hasil_ke_siswa=false` default).

### NB-02 — Cleanup E2E/load mengembalikan 409 pada banyak DELETE
**Hasil:** BUKAN bug — guard yang benar.
Bukti: `DELETE jadwal/paket/bank-soal/siswa/mapel/kelas/jurusan` setelah
ujian mengembalikan 409 (`conflict` / `bank_in_use`) dengan pesan yang
tepat, mis. "Jadwal sudah punya attempt peserta. Tidak boleh dihapus
(hasil ujian tidak pernah dihapus fisik)." / "Siswa sudah punya 1
attempt ujian. Nonaktifkan saja, jangan hapus." Cara bersih tuntas
data tes adalah `DROP DATABASE` scratch (dicetak program di akhir).

---

## Cakupan yang TIDAK menemukan bug

- Alur E2E penuh 33 langkah (`backend/test/e2e/e2e.go`): jurusan →
  kelas → mapel → guru → siswa → bank soal (PG + uraian) → paket →
  jadwal → token → ujian siswa (jawab + ragu + submit) → koreksi →
  nilai → rekap → finalisasi → berita acara → export CSV — **33/33 PASS**
  (prefix QAE2E3, 16 Okt 2026). Nilai terverifikasi: PG benar (10 poin)
  + uraian skor 80 (8 poin) = **90.00**.
- Uji beban 300 siswa, run final QB4 (`backend/test/load/loadtest.go`,
  lihat `docs/UJI-BEBAN-300.md`): 300/300 mulai, 300 attempt,
  **3000/3000 jawaban** terverifikasi di DB (attempt lengkap 300,
  `nilai_akhir` 300), **0 respons 5xx**, **0 timeout**, **0 deadlock**
  teramati, **0 dari 300** respons attempt mengandung `"kunci"`
  (kunci jawaban tidak bocor ke siswa). 1800 respons 429 terjadi dan
  pulih via retry harness — dicatat sebagai BUG-03 (celah ketahanan,
  bukan kegagalan server).
- 7 kegagalan aktivasi klien pada run QB3 (server mencatat 303/303
  ganti-password 200, nol 429) TIDAK dinyatakan sebagai bug aplikasi:
  penyebab pasti tidak teridentifikasi dan tidak terulang pada QB4
  (aktivasi gagal 0). Dicatat di `docs/UJI-BEBAN-300.md` §3.5 butir 3
  sebagai transien yang perlu bukti tambahan bila terulang.
