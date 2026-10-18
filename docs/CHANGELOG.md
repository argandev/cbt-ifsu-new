# CHANGELOG — CBT SMK v2

Format: versi mengikuti fase proyek (Fase 1 → v1.0.0, Fase 2 → v2.0.0,
Fase 3 → v2.1.0). Tanggal memakai tanggal penyelesaian fase
(16 Okt 2026 untuk ketiganya, sesuai laporan tim).
Dokumen ini milik Tim Fase 3 (Dokumen Final).
Kontrak teknis: `docs/KONTRAK.md`. Bukti uji beban:
`docs/UJI-BEBAN-300.md` + `docs/BUG-REPORT-QA-BEBAN.md`.

---

## v2.1.0 — Fase 3: Hardening (16 Okt 2026)

### Pemisahan tugas (SoD) + notifikasi

- Endpoint pengecualian darurat SoD: `POST /api/v1/sod/override`
  (permission `sod.override`, admin saja). Body
  `{aksi, entitas, entitas_id, alasan}`; `alasan` wajib min
  **20 karakter** (kurang → 400). Efek: catat `sod_overrides` +
  `audit_log` dengan flag `SOD_OVERRIDE` (detail JSON berisi aksi +
  alasan + entitas) + buat notifikasi "Pengecualian SoD" untuk SEMUA
  user role admin kecuali actor.
- Notifikasi: `GET /api/v1/notifikasi` (milik sendiri, pagination
  `limit`/`offset`), `POST /api/v1/notifikasi/:id/baca`
  (bukan milik → 404, anti-IDOR). Bel notifikasi di frontend.
- Tabel baru (migrasi 0012): `sod_overrides`, `notifikasi`.
- Audit log tetap append-only: tidak ada endpoint ubah/hapus log;
  pengecualian SoD tidak pernah dipakai untuk ubah nilai tanpa jejak
  / hapus audit log.

### Perbaikan BUG-01 / BUG-02 / BUG-03

- **BUG-01 (permission `analisis.*`):** `analisis.read` +
  `analisis.export` diselaraskan ke 4 sumber kebenaran —
  `docs/KONTRAK.md` §5 (total **62 permission**),
  `backend/db/seed/seed.sql` (grant proktor + guru),
  `backend/internal/rbac/rbac.go` (`AllPermissions` + matriks),
  migrasi 0008. Grant migrasi 0008 ke proktor/guru sebelumnya
  diam-diam no-op (tabel `roles` masih kosong saat migrasi berjalan);
  aturan tetap: **seed adalah sumber kebenaran grant**.
  Proktor melihat analisis TANPA teks pertanyaan (§5.3 kontrak).
- **BUG-02 (import siswa terpotong timeout):** timeout HTTP eksplisit
  — `ReadTimeout` 60 dtk, `WriteTimeout` 120 dtk, `IdleTimeout`
  120 dtk (via `StartConfig.BeforeServeFunc`, Echo v5 tidak punya
  `e.Server`). Import siswa dibuat **idempotent**: NISN yang sudah
  ada → `sudah_ada` (bukan gagal); respons
  `{sukses, gagal, sudah_ada, errors}`. Anjuran operasional tetap
  maks. 100 siswa per file.
- **BUG-03 (429 tanpa retry):** rate limit ujian dinaikkan —
  100 req/menit/IP untuk endpoint umum,
  **300 req/menit/IP** untuk `/api/v1/ujian/attempt/*`;
  penolakan 429 berkode `rate_limited` + header `Retry-After`
  (detik). Frontend `apiClient` me-retry 429 otomatis maks 3x
  (GET/HEAD + POST saja) dengan backoff 500 ms → 1 dtk → 2 dtk,
  mengikuti header `Retry-After` bila ada (dibatasi maks 10 detik).
  Rate limit login TIDAK berubah: 5x gagal/15 menit → kunci
  15 menit, 429 `login_locked`.

### Pembahasan end-to-end

- Guru mengisi `soal.pembahasan` lewat endpoint bank soal
  (migrasi 0011). Pembahasan dikembalikan di endpoint guru +
  di `GET .../hasil` SETELAH submit bila tampil_hasil efektif
  (Tryout ATAU paket `tampil_hasil` AND settings
  `tampil_hasil_ke_siswa`). Pembahasan TIDAK PERNAH dikirim ke
  siswa sebelum submit (endpoint attempt tidak memilih kolom
  kunci/pembahasan).

### Hardening frontend / backend

- Hapus / aksi destruktif HANYA via `POST` (atau `DELETE`), tidak
  pernah via `GET`.
- Anti-IDOR: resource yang tidak berhak → 404, bukan 403.
- Kunci jawaban tidak pernah dikirim ke browser siswa
  (terverifikasi 0 kebocoran pada uji beban 300 siswa).
- `gabungTanggalJam` menerima `HH:MM` (dinormalisasi ke `HH:MM:SS`)
  selain `HH:MM:SS` — POST jadwal dengan jam `"08:00"` → 201.
- Temuan audit UI Fase 2 yang sudah ditindaklanjuti: route
  `/adm/rekap/:jadwalId/berita-acara` (param), route `/403`
  (redirect `RequireAuth staffOnly`), halaman `NotFoundPage` untuk
  catch-all, `print.css` + `ErrorBoundary` diimpor di `main.tsx`,
  `ConfirmDialog` menggantikan `window.confirm`,
  alur token siswa di beranda siswa (field respons backend +
  input token + lanjut attempt berjalan).
  Detail audit: `docs/AUDIT-UI.md`.

### Verifikasi Fase 3

- E2E penuh 33 langkah: 33/33 PASS (nilai terverifikasi:
  PG benar 10 poin + uraian skor 80 = **90.00**).
- Uji beban 300 siswa (run QB4): 300/300 submit, 3000/3000 jawaban,
  0 respons 5xx, 0 timeout, 0 deadlock, 0 kebocoran kunci;
  throughput fase ujian 95,62 req/dtk.
- Yang masih berstatus "menunggu verifikasi akhir": dokumen
  `docs/VERIFIKASI-AKHIR.md` belum ada; `make sqlc`
  pasca-perbaikan migrasi 0005 dan pengujian e2e pasca-Fase-3 di
  database bersih belum tercatat dalam laporan verifikasi akhir.

---

## v2.0.0 — Fase 2: UKK, analisis butir, kompetensi, event/ruangan/kartu, backup (16 Okt 2026)

### UKK

- Skema + kompetensi + peserta UKK (CRUD, `ukk.*` permission).
- Penilaian: teori 40% + praktik 60%; rubrik Proses 30% /
  Produk 50% / Sikap 20%; lulus ≥ 70.
- Sertifikat UKK (cetak, `print.css`); halaman
  `/adm/ukk`, `/adm/ukk/:id`, `/adm/ukk/:id/nilai`,
  `/adm/ukk/:id/sertifikat/:siswaId`.
- Rubrik + format sertifikat: `docs/UKK-RUBRIK-SERTIFIKAT.md`.
  Paket `ukkx` sebagai cross-check independen fungsi murni.
- SoD UKK: penguji UKK ≠ finalisator nilai UKK.

### Analisis butir soal + nilai per kompetensi

- Endpoint `GET /api/v1/analisis/:jadwalId` +
  `GET /api/v1/analisis/:jadwalId/export.csv`
  (permission `analisis.read` / `analisis.export`).
- Statistik per soal dari attempt SELESAI pada satu jadwal:
  tingkat kesukaran (P = B/N), daya beda, pengecoh; jawaban
  kosong dibedakan dari salah. Rumus final:
  `docs/ANALISIS-RUMUS.md`.
- Nilai per kompetensi: agregasi nilai per kompetensi
  (`/adm/kompetensi`).
- Proktor melihat statistik TANPA teks pertanyaan.

### Event + ruangan + kartu peserta

- Event ujian + ruangan (CRUD, `event.*` permission):
  `/adm/event`, `/adm/ruangan`.
- Kartu peserta admin (`/adm/kartu`) + kartu sendiri siswa
  (`/kartu`, `/kartu/:jadwalId`).

### Backup / restore

- Endpoint backup: `POST /api/v1/backup` (buat, `backup.create`),
  `GET /api/v1/backup` (daftar), `GET /api/v1/backup/:id/download`,
  `POST /api/v1/backup/:id/restore` (`backup.restore`).
- Halaman `/adm/backup`. Cron `deploy/backup.sh` (`pg_dump`
  format arsip + arsip upload, retensi default 30 hari).
- Restore membuat backup pre-restore otomatis dan ditolak bila ada
  attempt berstatus berjalan (migrasi 0010).

### Frontend extras

- Halaman 403/404 (`ForbiddenPage`, `NotFoundPage`), error boundary.
- Komponen bersama: `ConfirmDialog`, `StatCard`, `PageHeader`,
  `EmptyState`, `Pagination`, `SearchInput`, `StatusBadge`,
  `PrintLayout`, `Sidebar`/`Topbar`/`Breadcrumbs`, `print.css`.
- Bel notifikasi (`NotifikasiBell`).
- Alur token siswa di beranda siswa: input token (kecuali Tryout),
  lanjut attempt berjalan, lihat hasil attempt selesai.

---

## v1.0.0 — Fase 1: MVP (16 Okt 2026)

### Auth + RBAC

- Session cookie server-side + CSRF (bukan JWT): cookie
  `cbt_session` (`HttpOnly`, `SameSite=Lax`, `Secure` bila prod,
  TTL 12 jam, idle 2 jam) + cookie `cbt_csrf` (dibaca JS untuk
  header `X-CSRF-Token`); sesi tersimpan di DB, restart server
  tidak membuat logout massal.
- Password Argon2id (format PHC). `must_change_password=true` →
  login ditolak 403 `PASSWORD_CHANGE_REQUIRED` (tanpa session);
  ganti password menghapus SEMUA session user.
- CLI: `create-admin` (hanya bila tabel `users` kosong),
  `reset-password`, `migrate-up` / `migrate-down`.
- RBAC deny-by-default: 6 role, 62 permission (Fase 1: 59;
  +`rekap.finalisasi`, +`rekap.remedial` pada integrasi Fase 1;
  +`analisis.*`, +`sod.override` pada Fase 3 — total akhir 62).
  Matriks: `docs/KONTRAK.md` §5.

### Master data

- CRUD jurusan, kelas, mapel, guru, siswa; guard hapus 409 bila
  dipakai jadwal/paket/attempt/nilai akhir.
- Import siswa CSV (multipart, idempotent pada Fase 3);
  akun login siswa dibuat otomatis (username = NISN, password
  awal = NISN, wajib ganti saat login pertama).
- Scope guru: hanya mapel + kelas yang ditugaskan
  (penegakan di server).

### Bank soal + paket + jadwal + ujian

- Bank soal: CRUD, lock/unlock (pembuat bank ≠ pengunci bank),
  import; **5 tipe standar AKM + 11 varian/ekstensi aplikasi**
  (`docs/TIPE-SOAL.md`); soal kosong = skor 0.
- Paket ujian (PTS/PAS/USBK/UKK/Tryout) + susunan soal.
- Jadwal: token (token kedua menghanguskan pertama, expiry
  900 detik), toleransi telat 15 menit, pause membekukan timer,
  usulan-unblock (pengawas) / unblock (proktor/admin).
- Ujian siswa: timer server-side, auto-submit, autosave jawaban,
  tandai ragu-ragu, pelanggaran (deduplikasi 5 detik),
  auto-block per-attempt (ke-1 peringatan, ke-2 block),
  unblock = lanjut sisa waktu. Tryout tanpa token + pembahasan
  langsung. Nilai esai disembunyikan sampai final.
- Koreksi esai/uraian guru; nilai akhir = Σ poin / Σ bobot × 100,
  dihitung backend; `nilai_akhir` disinkronkan dari alur ujian
  ke rekap.

### Rekap + monitor + settings + audit

- Rekap nilai + export CSV + berita acara; finalisasi +
  remedial (nilai tertinggi); KKTP per mapel/TP.
- Monitor peserta real-time: `POST /api/v1/monitor/ticket`
  (ticket 30 detik) + `WS /ws/monitor`.
- Settings (token expiry, toleransi telat, max pelanggaran,
  remedial, tampil hasil, nama sekolah).
- Audit log append-only (login/logout, masuk ujian, submit,
  pelanggaran, block/unblock, token, perubahan nilai, export,
  akses ditolak) — tanpa endpoint ubah/hapus.

### Perbaikan integrasi Fase 1 (tercatat di kontrak §11.9)

- Migrasi 0005 memakai nama kanonis `attempt` + `jadwal_ujian`.
- Seed 59 permission (+`rekap.finalisasi`, +`rekap.remedial`);
  proktor TANPA `paket.*`; settings string memakai literal JSONB.
- Migrasi 0006: FK `soal → kisi_kisi / kompetensi`
  (ON DELETE SET NULL).
- Font Plus Jakarta Sans via `@fontsource`
  (tanpa Google Fonts); plugin `@tailwindcss/vite`.

---

## Keterbatasan yang diketahui

- **Pembuatan akun staf belum ada jalur aplikasinya.**
  Satu-satunya pembuatan akun di kode aplikasi adalah `create-admin`
  (hanya bila tabel `users` kosong) dan akun siswa otomatis
  (import / buat siswa). Tidak ada endpoint / UI / CLI untuk membuat
  akun staf tambahan (guru / proktor / pengawas / penguji UKK);
  profil guru wajib merujuk `user_id` yang sudah ada. Jalur
  pembuatan akun staf perlu disiapkan sebelum go-live.
- **Angka uji beban bukan jaminan untuk hardware lain.** Uji beban
  300 siswa diukur di VM bersama memakai loopback, bukan LAN
  sekolah yang sebenarnya. Uji ulang di server sekolah sebelum
  hari-H (lihat `docs/UJI-BEBAN-300.md` §4 butir 9).
- **Insiden transien QB3 belum terjelaskan.** 7 dari 300 aktivasi
  gagal di sisi klien pada run QB3 sementara log server mencatat
  303/303 ganti-password 200. Penyebab pasti tidak teridentifikasi
  dan tidak terulang pada QB4 (aktivasi gagal 0).
- **RSS server saat uji beban belum dijelaskan.** VmRSS 1,1–1,5 GB
  dan TURUN selama run (bukan bocor monotonik), tetapi angka
  absolutnya besar untuk server API Go. Perlu profiling `pprof`
  sebelum menyatakan normal.
- **Import siswa: anjuran maks 100 siswa per file** (~17 dtk);
  import 300 baris (~50 dtk) kini muat dalam `WriteTimeout`
  120 dtk, tetapi batch 100/file tetap anjuran operasional.
- **Verifikasi akhir masih menunggu.** `docs/VERIFIKASI-AKHIR.md`
  belum ada; `make sqlc` pasca-perbaikan migrasi 0005 dan pengujian
  e2e pasca-Fase-3 di database bersih belum tercatat dalam laporan
  verifikasi akhir.
- **Batas teknologi web:** HP kedua, foto layar, joki, internet/AI
  di perangkat lain, screenshot OS, dan listrik/jaringan padam
  TIDAK bisa dicegah aplikasi web — lihat tabel jujur di
  `docs/KEAMANAN.md` §8.
