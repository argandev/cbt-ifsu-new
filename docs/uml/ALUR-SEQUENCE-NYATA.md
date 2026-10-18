# ALUR SEQUENCE NYATA — CBT SMK v2 (Telusur dari Kode, Fase 1)

> **Status:** Dokumen telusur Fase 1. Disusun HANYA dari membaca kode sumber (bukan dari diagram lama, bukan dari ingatan).
> **Tanggal telusur:** 18 Okt 2026.
> **Aturan penulisan:** Nama teknis kode (file, endpoint, tabel, kolom, setting) ditulis apa adanya di dokumen ini. Translasi ke bahasa desain (PIM) dilakukan Fase 2, bukan di sini.
> **Cakupan:** Alur A (login + ganti password + sesi + CSRF), Alur B (ujian), Alur C (pemantauan + anti-curang). Di luar itu (master data, bank soal, paket/jadwal CRUD, kartu, UKK, backup, analisis) TIDAK ditelusuri di sini, kecuali bila bersinggungan langsung (mis. token jadwal, remedial, koreksi esai).

## Daftar file kode yang dibaca

| # | File | Isi yang dipakai |
|---|------|------------------|
| 1 | `frontend/src/App.tsx` | Routing + guard `RequireAuth` + `STAFF_ROLES` + `fetchMe` saat mount |
| 2 | `frontend/src/pages/LoginPage.tsx` | Form login, cabang `PASSWORD_CHANGE_REQUIRED`, navigasi per role |
| 3 | `frontend/src/pages/GantiPasswordPage.tsx` | Form ganti password 2 mode, validasi client, navigasi sesudah berhasil |
| 4 | `frontend/src/pages/siswa/SiswaHomePage.tsx` | Daftar jadwal hari ini + `handleMasuk` (titik masuk ujian #1) |
| 5 | `frontend/src/pages/siswa/UjianPage.tsx` | Halaman ujian lengkap: daftar jadwal, mulai, soal, autosave, ragu, anti-curang, timer, submit, hasil |
| 6 | `frontend/src/pages/siswa/HasilPage.tsx` | Halaman hasil per-soal |
| 7 | `frontend/src/pages/admin/monitor/MonitorPage.tsx` | Monitor WebSocket + fallback polling + pause/resume + unblock + daftar usulan |
| 8 | `frontend/src/pages/admin/jadwal/JadwalPage.tsx` | Generate token + status token (hanya dibaca bagian token) |
| 9 | `frontend/src/components/exam/Timer.tsx` | Hitung mundur lokal 1 detik, sinkron dari server, `onExpire` sekali |
| 10 | `frontend/src/components/exam/SoalRenderer.tsx` | Format nilai jawaban per tipe (dibaca sekilas, pemetaan `onChange`) |
| 11 | `frontend/src/components/exam/NavigasiSoal.tsx` | Grid nomor soal: dijawab / ragu / belum / aktif |
| 12 | `frontend/src/components/exam/BannerPeringatan.tsx` | Banner peringatan ke-1 dan layar diblokir |
| 13 | `frontend/src/store/auth.ts` | `useAuthStore`: `fetchMe`, `login`, `logout` |
| 14 | `frontend/src/store/exam.ts` | `useExamStore` (dicatat: TIDAK dipakai alur UjianPage — lihat §5) |
| 15 | `frontend/src/lib/api.ts` | `apiClient`: cookie `credentials: include`, header `X-CSRF-Token` dari cookie `cbt_csrf`, retry 429 |
| 16 | `backend/internal/auth/auth.go` | Endpoint login/logout/me/ganti-password/csrf, rate limit login, cookie sesi |
| 17 | `backend/internal/auth/password.go` | Hash/verify Argon2id |
| 18 | `backend/internal/session/session.go` | `Create/Get/Touch/Destroy/DestroyAllForUser/VerifyCSRF`, TTL 12 jam, idle 2 jam |
| 19 | `backend/internal/middleware/middleware.go` | `RequireAuth`, `CSRF`, `RateLimit`, `isExemptAuth`, `isExemptCSRF` |
| 20 | `backend/internal/exam/exam.go` | Seluruh endpoint paket/jadwal/token/pause/resume/unblock/ujian/koreksi (2.832 baris, dibaca penuh) |
| 21 | `backend/internal/grading/grading.go` | `HitungPoin`, `HitungNilaiAkhir`, `Dinilai`, `NormalisasiTipe` |
| 22 | `backend/internal/monitor/monitor.go` | Ticket, Hub WebSocket, snapshot, broadcast |
| 23 | `backend/internal/rekap/rekap.go` | Endpoint remedial (hanya bagian remedial) + finalisasi (disinggung) |
| 24 | `docs/uml/DIAGNOSIS-SEQUENCE-JELAS.md` | Hanya dipakai sebagai pembanding "diagram lama" pada §6 (bukan sebagai sumber alur) |

**Konvensi respons API:** Semua respons sukses dibungkus `{"data": ...}` (kecuali `GET /api/v1/auth/csrf` yang mengembalikan `{"csrf_token": ...}` mentah). Respons gagal berbentuk `{"error": {"code": ..., "message": ...}}`. `apiClient` mengupas envelope `data` dan melempar `ApiError(code, message, status)`.

---

## 0. Prinsip arsitektur yang menentukan semua alur

1. **Tidak ada JWT.** Auth memakai **session cookie server-side**: cookie `cbt_session` (HttpOnly) membawa ID session; data session (`user_id`, `role`, `csrf_token`, `expires_at`, `last_seen_at`) tersimpan di tabel `sessions` di PostgreSQL. Restart server tidak membuat logout massal.
2. **Timer berjangkar server.** Frontend hanya menampilkan hitung mundur. Server menolak simpan/submit setelah deadline, dan server-lah yang memutuskan auto-submit (secara *lazy*, lihat §B.9).
3. **Kunci jawaban tidak pernah dikirim ke frontend selama ujian.** `muatSoalAttempt` hanya memilih kolom tampil (`soal.jawaban` = data tampil Tim 3, BUKAN kunci). Kunci (`soal.kunci`) hanya dibaca server saat grading dan hanya dirangkum di halaman hasil bila boleh tampil.
4. **Semua mutasi ujian dicatat audit** (`audit_log`), kecuali isi jawaban (hanya `attempt_id` + `soal_id` yang dicatat untuk `ujian.jawaban`).
5. **Rate limit global:** middleware `RateLimit()` = 100 req/menit/IP untuk API umum, 300 req/menit/IP untuk `/api/v1/ujian/attempt/*`; aset statis + halaman SPA tidak dihitung. Respons 429 berkode `rate_limited` di-retry otomatis oleh `apiClient` maks 3× (backoff 500 ms → 1 dtk → 2 dtk, atau mengikuti header `Retry-After`, maks 10 dtk) untuk GET/HEAD/POST. Rate limit login diatur terpisah di paket `auth` (lihat §A.3).

---

## Alur A — Login + wajib ganti password + sesi + CSRF

### A.0 Partisipan

| Partisipan | Peran |
|---|---|
| Pengguna (siswa / staf: admin, proktor, guru, pengawas, penguji_ukk) | Mengisi kredensial, mengisi form ganti password |
| `App.tsx` (SPA shell) | Saat mount memanggil `fetchMe`; guard rute `RequireAuth` |
| `LoginPage.tsx` | Form login; menangani cabang `PASSWORD_CHANGE_REQUIRED` |
| `GantiPasswordPage.tsx` | Form ganti password |
| `useAuthStore` (`store/auth.ts`) | Cache user + permissions; `login()` / `logout()` / `fetchMe()` |
| `apiClient` (`lib/api.ts`) | `fetch` dengan `credentials: "include"`; membaca cookie `cbt_csrf`; mengirim header `X-CSRF-Token` untuk semua mutasi |
| Backend `auth.go` | Endpoint `POST /auth/login`, `POST /auth/logout`, `GET /auth/me`, `POST /auth/ganti-password`, `GET /auth/csrf` |
| Backend `session.go` | Tabel `sessions`: create/get/touch/destroy |
| Backend `middleware.go` | `RequireAuth` (validasi cookie sesi per request), `CSRF` (validasi header) |
| PostgreSQL | Tabel `users`, `sessions`, `role_permissions` + `permissions` + `roles`, `audit_log` |

### A.1 Langkah demi langkah (alur normal, tanpa wajib ganti password)

1. **Aplikasi dibuka** → `App.tsx` mount → `useAuthStore.fetchMe()` → `GET /api/v1/auth/me`.
   - Bila cookie `cbt_session` valid → `RequireAuth` memuat session dari DB, `Touch` (perbarui `last_seen_at`), isi context (`user_id`, `role`, `permissions`, `csrf_token`) → `GET /me` mengembalikan `{user, permissions}` → store terisi → rute terlindungi langsung tampil.
   - Bila tidak ada/kedaluwarsa → `401` → store `user = null` → `RequireAuth` mengarahkan ke `/login` (dengan `state.from = pathname` — **nilai `from` ini tidak pernah dipakai untuk redirect balik; LoginPage selalu navigasi berdasarkan role**; lihat §5).
2. **Pengguna mengisi** Email/Nama pengguna + Kata sandi di `LoginPage`, submit → `useAuthStore.login(login, password)` → `POST /api/v1/auth/login` dengan body `{login, password}`.
   - Catatan: `apiClient` mengirim header `X-CSRF-Token` juga untuk request ini (bila cookie `cbt_csrf` ada), tetapi middleware `CSRF` **meng-exempt** path `/api/v1/auth/login`, jadi header itu diabaikan server pada langkah ini.
3. **Server — validasi login** (`auth.go`, handler login), urutan persis:
   1. Body tidak valid / login atau password kosong → `400 validation_error`.
   2. **Rate limit login:** kunci = login (lowercase, trim). Bila `loginLocked(key)` → `429 login_locked` ("Akun terkunci sementara… Coba lagi dalam 15 menit."). Aturan: maks **5 gagal per akun per 15 menit**, percobaan ke-6 mengunci 15 menit. Ditegakkan **in-memory per instance server** (bukan di DB). Hanya kegagalan "akun tidak ditemukan" dan "password salah" yang dicatat sebagai gagal; cabang nonaktif / wajib-ganti-password di bawah TIDAK menambah hitungan gagal.
   3. `SELECT … FROM users WHERE lower(email) = lower($1) OR lower(username) = lower($1)` — login boleh email ATAU username, case-insensitive, spasi dipangkas. Tidak ketemu → catat gagal + audit `auth.login.gagal` (alasan "akun tidak ditemukan") → `401 invalid_credentials`.
   4. `VerifyPassword` (Argon2id, parameter `m=65536, t=1, p=4`, salt 16 byte, format `$argon2id$v=19$…`, perbandingan constant-time). Password salah / format hash tidak valid → catat gagal + audit `auth.login.gagal` (alasan "password salah") → `401 invalid_credentials`.
   5. `is_active = false` → `403 account_inactive` ("Akun tidak aktif. Hubungi admin.").
   6. `must_change_password = true` → **`403 {error.code: "PASSWORD_CHANGE_REQUIRED"}` — TANPA membuat session.** (Ini cabang kunci alur A; lihat §A.2.)
   7. Sukses → `loginReset(key)` (hapus hitungan gagal) → `session.Create(...)` → `INSERT INTO sessions (id, user_id, role, csrf_token, ip, user_agent, created_at, expires_at = now+12 jam, last_seen_at = now)` dengan `csrf_token` acak 32 byte (hex) → set 2 cookie (lihat §A.4) → audit `auth.login` → `200 {data: {user, permissions}}` (`permissions` dari `role_permissions` JOIN `permissions`, fallback ke matriks statis `rbac` bila DB belum di-seed).
4. **Frontend — sesudah login sukses:** `useAuthStore` menyimpan `user` + `permissions` → `LoginPage.handleSubmit` membaca `role` dari store:
   - `role === "siswa"` → `navigate("/")` (SiswaHomePage).
   - role lain → `navigate("/adm")` (Dashboard admin).
   - Bila `user` sudah terisi saat membuka `/login`, form tidak ditampilkan; tampil teks "Sudah login sebagai …" + link "Lanjut ke beranda".
5. **Setiap request API berikutnya:** browser mengirim cookie `cbt_session` otomatis (`credentials: "include"`); `RequireAuth` memvalidasi ke DB tiap request (`session.Get`: `404` bila tidak ada; hapus + `ErrExpired` bila melewati `expires_at` **atau** tidak disentuh > 2 jam idle) → `Touch` perbarui `last_seen_at` → handler jalan. Bila session tidak valid → `401 unauthenticated`.
6. **Setiap mutasi (POST/PUT/PATCH/DELETE):** `apiClient` membaca cookie `cbt_csrf` (cookie non-HttpOnly) dan mengirim header `X-CSRF-Token`. Middleware `CSRF` membandingkan header dengan `csrf_token` session (constant-time). Tidak cocok → `403 csrf_token_invalid`. Pengecualian CSRF: method aman (GET/HEAD/OPTIONS); path `/api/v1/auth/login` dan `/api/v1/auth/csrf`; dan request tanpa cookie session (dilewati di sini — `RequireAuth` yang akan menolak 401, kecuali `ganti-password` mode kredensial yang memang dirancang tanpa session).

### A.2 Cabang "wajib ganti password"

Pemicu: `must_change_password = true` di tabel `users` (akun baru dari `CreateAdmin` — hanya saat tabel `users` kosong — dan dari `ResetPassword` CLI; keduanya set flag `TRUE`).

1. Login → server menjawab `403 PASSWORD_CHANGE_REQUIRED` (session TIDAK dibuat).
2. `LoginPage` menangkap `ApiError` dengan `code === "PASSWORD_CHANGE_REQUIRED"` → `navigate("/ganti-password", {state: {login}})` (login ikut terbawa sebagai nilai awal field).
3. `GantiPasswordPage` tampil dengan keterangan "Kata sandi awal harus diganti sebelum login." Validasi **client** (di handler, bukan di `rules` antd): password baru ≥ 8 karakter; konfirmasi sama dengan password baru; password baru ≠ password lama. Gagal validasi → Alert error, request TIDAK dikirim.
4. Submit → `POST /api/v1/auth/ganti-password` body `{login, password_lama, password_baru}`.
5. **Server — dua mode dalam satu endpoint** (endpoint ini exempt dari `RequireAuth` global justru agar mode kredensial bisa jalan; tetapi `RequireAuth` tetap memuat session ke context bila cookie ada):
   - **Mode (a) session aktif:** bila context berisi `user_id` (cookie session valid), user diambil dari session; field `login` diabaikan.
   - **Mode (b) kredensial:** tanpa session, user dicari lewat `login` (email/username). `login` kosong → `401 unauthenticated`.
   - Validasi server: password baru < 8 → `400`; password baru == password lama → `400`.
   - `VerifyPassword(password_lama, hash tersimpan)` gagal → `401 invalid_credentials` ("Password lama salah.").
   - Sukses → `UPDATE users SET password_hash = hash baru (Argon2id), must_change_password = FALSE, updated_at = now()` → **`session.DestroyAllForUser(user_id)` — SEMUA session user dihapus** → `clearSessionCookies` (kedua cookie dihapus, `MaxAge: -1`) → audit `auth.ganti-password` → `200 {data: {password_changed: true}}`.
6. **Frontend:** `navigate("/login", {state: {pesan: "Kata sandi berhasil diganti. Silakan login dengan kata sandi baru."}})` → pengguna **wajib login ulang** dengan password baru (tidak ada auto-login sesudah ganti password). Login ulang berjalan seperti §A.1 dan kini lolos karena flag sudah `FALSE`; session + cookie baru dibuat.
7. **Cabang terkait — ganti password dari dalam sesi login:** halaman yang sama dipakai; mode (a) berlaku; sesudah berhasil, semua session (termasuk session yang sedang berjalan) dihapus sehingga pengguna tetap wajib login ulang.

### A.3 Guard dan kondisi/cabang tambahan Alur A

| Guard / kondisi | Efek |
|---|---|
| 5 login gagal / akun / 15 menit (in-memory) | Percobaan ke-6 → `429 login_locked`, kunci 15 menit |
| Login sukses | Hitungan gagal akun di-reset |
| `is_active = false` | `403 account_inactive`, tanpa session |
| `must_change_password = true` | `403 PASSWORD_CHANGE_REQUIRED`, tanpa session |
| Session melewati `expires_at` (12 jam) atau idle > 2 jam | `401 unauthenticated`, session dihapus dari DB |
| Header `X-CSRF-Token` ≠ `csrf_token` session (mutasi) | `403 csrf_token_invalid` |
| Rute `/` , `/ujian/:attemptId`, `/hasil/:attemptId`, `/kartu` | `RequireAuth` (semua role yang login); tanpa user → redirect `/login` |
| Rute `/adm/*` | `RequireAuth staffOnly`; role bukan `admin/proktor/guru/pengawas/penguji_ukk` (mis. siswa) → redirect `/403` |
| `POST /auth/logout` | `session.Destroy` (session dari cookie) + hapus kedua cookie + audit `auth.logout` → `200 {logged_out: true}`; frontend (`useAuthStore.logout`) mengosongkan store walau request gagal (`finally`) |

### A.4 Keadaan sesi/CSRF per langkah (ringkasan)

| Langkah | Cookie `cbt_session` | Cookie `cbt_csrf` | Baris `sessions` di DB |
|---|---|---|---|
| Sebelum login | — | — (atau sisa lama) | — |
| Login ditolak `PASSWORD_CHANGE_REQUIRED` | tidak dibuat | tidak dibuat | tidak ada baris baru |
| Login sukses | dibuat (HttpOnly, SameSite Lax, Secure hanya bila prod, MaxAge 12 jam) | dibuat (non-HttpOnly, nilai = `csrf_token` session, MaxAge 12 jam) | 1 baris baru (`expires_at = now + 12 jam`) |
| Request terautentikasi | dikirim otomatis | dibaca frontend → header `X-CSRF-Token` | `last_seen_at` diperbarui (`Touch`) |
| Ganti password sukses | dihapus (MaxAge −1) | dihapus (MaxAge −1) | **semua** session user dihapus |
| Logout | dihapus (MaxAge −1) | dihapus (MaxAge −1) | session cookie itu dihapus |
| Idle > 2 jam / umur > 12 jam | ditolak server (401) | tidak berlaku | baris dihapus saat `Get` mendeteksi kedaluwarsa |

> Catatan: endpoint `GET /api/v1/auth/csrf` (mengembalikan `{"csrf_token": ...}` mentah, tanpa envelope `data`) **ada di backend tetapi tidak pernah dipanggil oleh frontend** — frontend selalu membaca token CSRF dari cookie `cbt_csrf` (`lib/api.ts: getCsrfToken()`).

---

## Alur B — Ujian

### B.0 Partisipan

| Partisipan | Peran |
|---|---|
| Siswa | Memasukkan token, menjawab soal, menandai ragu, submit |
| `SiswaHomePage.tsx` | Daftar jadwal hari ini + titik masuk ujian #1 (`handleMasuk`) |
| `UjianPage.tsx` | Titik masuk ujian #2 (daftar jadwal + mulai) + seluruh interaksi selama ujian |
| `Timer.tsx` | Hitung mundur lokal |
| `SoalRenderer.tsx` | Input jawaban per tipe; memetakan nilai kosong ke format backend |
| `NavigasiSoal.tsx` | Pindah soal; status dijawab/ragu/belum |
| `BannerPeringatan.tsx` | Tampilan peringatan / diblokir |
| Backend `exam.go` | Endpoint `ujian/aktif`, `ujian/mulai`, `attempt/:id`, `.../jawaban`, `.../ragu`, `.../pelanggaran`, `.../submit`, `.../status`, `.../hasil`, `koreksi/*`, jadwal/token/pause/resume/unblock |
| Backend `grading.go` | `HitungPoin` per soal (saat simpan) + `HitungNilaiAkhir` (saat finalisasi) |
| Backend `monitor.go` | `Broadcast` perubahan attempt ke room monitor |
| PostgreSQL | `jadwal_ujian`, `paket_ujian`, `paket_soal`, `soal`, `attempt`, `jawaban`, `pelanggaran`, `nilai_akhir`, `siswa`, `kelas`, `settings`, `audit_log` |

### B.1 Daftar jadwal + input token

1. Siswa (sudah login) membuka `/` (SiswaHomePage) atau `/ujian` tanpa `attemptId` (UjianPage) → `GET /api/v1/ujian/aktif` (hanya role `siswa`; UjianPage me-refetch tiap 30 detik, SiswaHomePage sekali tanpa retry).
2. Server `ujianAktif`: cari `kelas_id` siswa dari `siswa` (cocokkan `siswa.id = siswaID` — yang dalam konteks auth adalah `user_id` — lalu fallback `siswa.user_id = siswaID`). Kelas tidak ditemukan → `200 {items: [], kelas_tidak_ditemukan: true, pesan: …}`.
3. Server mengembalikan jadwal **hari ini** (tanggal = `now()` zona Asia/Jakarta) untuk kelas itu dengan status jadwal `draft / aktif / pause`, JOIN attempt siswa yang berstatus `berjalan / diblokir / selesai` (bila ada). Field per item: `jadwal_id, paket_id, tanggal, jam_mulai, jam_selesai, status, paket_kode, paket_nama, paket_jenis, durasi_menit, kelas_nama, attempt_id?, attempt_status?`.
4. Frontend menampilkan per jadwal:
   - `attempt_status === "berjalan"` → tombol **Lanjutkan Ujian** (SiswaHome: langsung `navigate(/ujian/:attemptId)` tanpa POST; UjianPage: `setAttemptId` — juga tanpa POST).
   - `attempt_status === "diblokir"` → Alert "Attempt Anda diblokir. Hubungi proktor untuk unblock." (tidak ada tombol mulai).
   - `attempt_status === "selesai"` → teks "Anda sudah menyelesaikan ujian ini."
   - `status === "pause"` → Alert "Ujian sedang dijeda proktor."
   - Selain itu → input token (kecuali `paket_jenis === "Tryout"`) + tombol **Mulai Ujian**.
   - Input token: UjianPage meng-`uppercase` + `maxLength 6`; SiswaHomePage tidak meng-uppercase (perbedaan kecil antar dua titik masuk; server membandingkan hash persis — token yang digenerate hanya berisi huruf/angka uppercase, lihat §B.2).
5. **Pembuatan token (di luar alur siswa, oleh admin/proktor):** `POST /api/v1/jadwal/:id/token` (JadwalPage) → server generate 6 karakter dari alfabet `ABCDEFGHJKLMNPQRSTUVWXYZ23456789` (tanpa `0/1/I/O`), simpan **hanya hash SHA-256** (`token_hash`) + `token_expires_at = waktu generate + token_expiry_detik` (setting, default **900 detik = 15 menit**). Token plaintext dikembalikan **sekali** di respons dan tidak bisa dilihat lagi (`GET …/token-status` hanya mengembalikan `ada_token`, `token_expires_at`, `kedaluwarsa_dalam_detik`, `kedaluwarsa` — tanpa nilai token). **Token kedua menimpa token pertama: token pertama langsung hangus.** Jadwal `selesai` tidak bisa dibuatkan token (`409`).

### B.2 Mulai ujian — `POST /api/v1/ujian/mulai` `{jadwal_id, token}`

Urutan validasi server (`mulaiUjian`, persis seperti di kode — urutan ini penting karena menentukan cabang mana yang menang):

| # | Pemeriksaan | Bila gagal |
|---|---|---|
| 1 | Role harus `siswa` | 401/403 dari `requireRole` |
| 2 | `jadwal_id` kosong | `400 validation_error` |
| 3 | Jadwal tidak ditemukan | `404 not_found` |
| 4 | Jadwal bukan untuk kelas siswa (`kelasSiswa` vs `jadwal.kelas_id`) | `403 forbidden` |
| 5 | Jadwal `selesai` | `409 conflict` |
| 6 | Jadwal `pause` | `409 conflict` ("Ujian sedang dijeda…") |
| 7 | **Attempt `berjalan` sudah ada** | **Dikembalikan apa adanya** (`200`, `sudah_ada: true`, `attempt_id` lama, `deadline`, `sisa_waktu_detik`) — idempotent, resume tanpa token baru |
| 8 | Attempt `diblokir` sudah ada | `403 forbidden` ("Attempt Anda diblokir…") |
| 9 | Attempt `selesai` sudah ada | `409 conflict` ("Anda sudah menyelesaikan ujian ini.") |
| 10 | `now < jam_mulai` (zona Asia/Jakarta) | `409 conflict` ("Ujian belum dimulai.") |
| 11 | `now > jam_selesai` | `409 conflict` ("Ujian sudah berakhir.") |
| 12 | **Toleransi telat:** `now > jam_mulai + toleransi_telat_menit` (setting, default **15 menit**) — hanya untuk attempt BARU | `403 forbidden` ("Batas toleransi keterlambatan habis…") |
| 13 | **Token** (dilewati bila `paket_jenis === "Tryout"`): token kosong ATAU jadwal belum punya `token_hash` → `401 token_invalid`; `SHA256(token) ≠ token_hash` → `401 token_invalid`; `now > token_expires_at` → `409 token_expired` | sesuai kolom |
| 14 | Paket tanpa soal | `409 conflict` |

Bila lolos semua:
- Susun soal paket (`paket_soal` urut `urutan`). Bila `paket.acak_soal`, urutan dikocok dengan seed deterministik dari `attempt_id` (FNV-1a). Bila `paket.acak_jawaban`, data tampil per soal dikocok (PG/PG Kompleks: array `opsi`; Menjodohkan: nilai `kanan` antar pasangan, `kiri` tetap; tipe lain tidak dikocok) dan disimpan per attempt di `attempt.urutan_opsi` agar stabil.
- Transaksi: bila jadwal `draft` → ubah jadi `aktif` (peserta pertama mengaktifkan jadwal). `INSERT INTO attempt (status='berjalan', mulai_at=now(), deadline = jam_selesai jadwal, urutan_soal, urutan_opsi)`.
  - **Deadline = jam selesai jadwal, BUKAN `mulai + durasi_menit`.** Kolom `paket.durasi_menit` hanya ditampilkan di UI, tidak dipakai menghitung deadline.
- Audit `ujian.mulai`; broadcast monitor `{type: "ujian_mulai", jadwal_id, attempt_id}`.
- Respons `201 {attempt_id, jadwal_id, status: "berjalan", deadline, sisa_waktu_detik, jumlah_soal}` → frontend `navigate(/ujian/:attemptId)` (SiswaHome) atau `setAttemptId` + `pushState /ujian/:attemptId` (UjianPage).
- **Perbedaan dua titik masuk:** `UjianPage.mulai()` lebih dulu meminta **fullscreen** (`document.documentElement.requestFullscreen()`; bila ditolak browser, ujian tetap jalan tanpa penalti) lalu POST; `SiswaHomePage.handleMasuk()` **tidak** meminta fullscreen. Keluar fullscreen selama ujian = pelanggaran (lihat §C.2).

### B.3 Muat soal — `GET /api/v1/ujian/attempt/:id`

1. Frontend (UjianPage, `attemptQuery`, tanpa `refetchInterval` — dimuat sekali) → `GET /api/v1/ujian/attempt/:id`.
2. Server: attempt harus milik siswa sendiri (bukan milik sendiri → `404 not_found`, sengaja disamarkan). Lalu `autoSubmitJikaLewat` (lihat §B.9) → baca jadwal → `muatSoalAttempt`:
   - Soal dari `paket_soal` JOIN `soal`, diurutkan mengikuti `attempt.urutan_soal` (soal baru di luar urutan ditambahkan di akhir).
   - Per soal: `{soal_id, nomor, tipe, pertanyaan, stimulus, media_url, tampilan (dari `urutan_opsi` attempt, fallback tampilan asli), bobot, jawaban (tersimpan), ragu (tersimpan)}`. **Kunci dan pembahasan tidak dipilih.**
   - Respons: `{attempt_id, jadwal_id, status, deadline, mulai_at, sisa_waktu_detik, jadwal_status, paket_jenis, server_now, soal[]}` + bila `status === "selesai"` dan tampil-hasil efektif: `{nilai_akhir, menunggu_koreksi}`.
3. Frontend menginisialisasi `jawabanMap` + `raguMap` dari server; bila `status === "diblokir"` → tampil layar diblokir; bila `status === "selesai"` → tampil kartu hasil submit.
4. Paralel: `GET /api/v1/kartu/:jadwalId/:userId` untuk bilah identitas (nama/NISN/kelas/mapel); gagal (404) → pakai nama akun login dan "—", ujian tetap jalan.

### B.4 Menjawab + autosave

1. Siswa mengubah jawaban di `SoalRenderer` → `UjianPage.ubahJawaban(soalId, nilai)`:
   - State lokal langsung diperbarui; indikator status simpan → `"menyimpan"`.
   - **Debounce 500 ms**: timer sebelumnya dibatalkan (`clearTimeout`); hanya nilai terakhir yang dikirim setelah 500 ms tanpa perubahan.
2. `POST /api/v1/ujian/attempt/:id/jawaban` `{soal_id, jawaban}`.
3. Server `simpanJawaban`, urutan:
   1. Attempt milik sendiri (bukan → 404); `diblokir` → `403`; `selesai` → `409` ("…Jawaban tidak bisa diubah.").
   2. `autoSubmitJikaLewat` — bila deadline lewat, attempt difinalisasi dulu, lalu simpan ditolak `409` ("Batas waktu ujian habis.").
   3. Jadwal `pause` → `409` ("Ujian sedang dijeda…").
   4. Soal harus anggota paket jadwal ini (bukan → `400`).
   5. Baca `tipe, kunci, bobot, tampilan` soal; **hitung poin saat itu juga** (`grading.HitungPoin`, lihat §B.8):
      - Tipe `uraian` (termasuk esai/essay/isian panjang/upload_file/praktik_singkat) → `poin = NULL`, `dikoreksi = false` (menunggu koreksi guru).
      - Tipe tidak dinilai (`skala_sikap`, `studi_kasus`, `setuju_tidak_setuju` tanpa kunci) → `poin = 0`, `dikoreksi = true`.
      - Lainnya → `poin = HitungPoin(...)`; bila `HitungPoin` mengembalikan `auto = false` (tipe tak dikenal) → `poin = NULL`, `dikoreksi = false`.
   6. `INSERT INTO jawaban … ON CONFLICT (attempt_id, soal_id) DO UPDATE SET jawaban, poin, dikoreksi, updated_at` — **nilai `ragu` lama dipertahankan** (`raguLama` dibaca dulu; upsert jawaban tidak menyentuh `ragu`).
   7. Audit `ujian.jawaban` (tanpa isi jawaban) + `broadcastAttempt` (update ringan ke monitor).
   8. Respons `200 {tersimpan: true, soal_id, server_now}`.
4. Frontend: sukses → status `"tersimpan"`; gagal → `"gagal"` ("Gagal — coba lagi"). **Tidak ada antrean retry autosave di frontend** — jawaban yang gagal tersimpan tidak dikirim ulang otomatis (siswa harus mengubah jawaban lagi untuk memicu simpan baru; nilai terakhir tetap di state lokal).
5. Catatan penting: **penilaian objektif terjadi per soal pada saat simpan**, bukan hanya saat submit. Nilai akhir saat submit/finalisasi adalah agregasi dari `poin` tersimpan + soal tanpa jawaban (poin 0).

### B.5 Tandai ragu-ragu

1. Checkbox "Tandai ragu-ragu" → `UjianPage.ubahRagu(soalId, ragu)` → state lokal + langsung `POST /api/v1/ujian/attempt/:id/ragu` `{soal_id, ragu}` (**tanpa debounce**; kegagalan diabaikan diam-diam `.catch(() => undefined)`).
2. Server `tandaiRagu`: milik sendiri; `status ≠ berjalan` → `409`; `autoSubmitJikaLewat`; `INSERT INTO jawaban (attempt_id, soal_id, jawaban NULL, ragu) ON CONFLICT … DO UPDATE SET ragu` — **ragu dapat ditandai pada soal yang belum dijawab** (baris jawaban dibuat dengan `jawaban = NULL`). Audit `ujian.ragu` + broadcast.
3. `NavigasiSoal` menampilkan kotak: aktif = primary; ragu = latar/border warning + ikon Flag; dijawab = hijau sukses; belum = default. **Ragu-only (ragu tanpa jawaban) dihitung sebagai BELUM dijawab** di penghitung frontend (`dijawab()` memeriksa isi jawaban), tetapi di backend `statusAttempt` menghitung `dijawab` sebagai `COUNT jawaban WHERE jawaban IS NOT NULL` — baris ragu-only (jawaban NULL) tidak terhitung dijawab di kedua sisi (konsisten).

### B.6 Polling timer server — `GET /api/v1/ujian/attempt/:id/status` tiap 30 detik

1. `statusQuery` di UjianPage: `refetchInterval: 30000`, aktif selama `attemptId` ada dan belum `hasilSubmit`.
2. Server `statusAttempt`: milik sendiri → `autoSubmitJikaLewat` → hitung `dijawab` (`jawaban IS NOT NULL`), `ragu` (`ragu = TRUE`), `total_soal` (dari `paket_soal`) → respons:
   `{attempt_id, jadwal_id, status, deadline, jadwal_status, sisa_waktu_detik, dijawab, ragu, total_soal, belum_dijawab, server_now}` + bila `status === "selesai"`: `nilai_akhir` (hanya bila tampil-hasil efektif dan tersimpan, selain itu `null`) dan `menunggu_koreksi`.
3. `sisa_waktu_detik` (server, `sisaWaktu`): bila jadwal `pause` dan ada `sisa_waktu_beku` → nilai beku; bila deadline lewat → 0; selain itu `deadline − now` (detik, dibulatkan ke bawah, min 0).
4. Frontend: `Timer` menerima `sisaServer = statusQuery.data?.sisa_waktu_detik ?? attempt.sisa_waktu_detik`; setiap perubahan prop → `Timer` menyetel ulang hitung mundur lokal (lihat §B.7). Efek polling: `status === "diblokir"` → tampil layar diblokir; `status === "selesai"` (mis. dijeda/difinalisasi dari sisi server atau auto-submit lewat request lain) → tampil kartu "Waktu ujian habis — jawaban tersimpan otomatis." dengan `nilai: null`.

### B.7 Timer lokal — `Timer.tsx`

- `setInterval` 1 detik: `sisa − 1` tiap tik; format `MM:SS` atau `HH:MM:SS` bila ≥ 1 jam; merah (`Tag color error`) bila sisa ≤ 300 detik (5 menit).
- Bila prop `sisaDetik` berubah (sinkron dari polling 30 detik), state lokal di-reset ke nilai server dan guard `sudahExpire` di-reset bila sisa > 0.
- Saat mencapai 0 → `onExpire()` **sekali** (dijaga `sudahExpire` ref) → `UjianPage.onTimerHabis` → `doSubmit()` (submit normal lewat endpoint submit; lihat §B.10). Jadi auto-submit di sisi tampilan = submit biasa yang dipicu timer; auto-submit di sisi server = finalisasi lazy pada request berikutnya (§B.9). Keduanya bertemu di `finalisasiAttempt` yang sama.

### B.8 Penilaian

**Per soal, saat simpan** — `grading.HitungPoin(tipe, jawaban, kunci, bobot) → (poin, auto)` (di `exam.go`, kunci diperkaya dulu untuk tipe `mengurutkan` dengan daftar `item` dari tampilan):

| Tipe (ternormalisasi) | Aturan poin |
|---|---|
| `pg` (termasuk pilihan_gambar, studi_kasus sebagai soal) | Benar semua / 0. Kunci `{"opsi": "B"}`, jawaban string label |
| `pg_kompleks` | Proporsional: `bobot × benarDipilih / jumlahKunci`, **tanpa penalti** jawaban salah. Kunci `{"opsi": [...]}`, jawaban array label |
| `benar_salah`, `setuju_tidak_setuju` | Proporsional per pernyataan: `bobot × benar / total`. Kunci `{"nilai": [bool…]}` + jawaban array bool sejajar, atau bentuk map `{"p1": bool}` (keduanya didukung) |
| `menjodohkan` | Proporsional per pasangan benar: `bobot × benar / total`. Kunci `{"pasangan": [{kiri, kanan}]}`, jawaban map `{kiri: kanan}` atau array pasangan |
| `isian_singkat` | Benar semua / 0, pencocokan ternormalisasi (huruf kecil, spasi diringkas, trim; tidak case-sensitive). Kunci: string tunggal, `{"jawaban": [...]}`, atau daftar alternatif |
| `isi_ganda`, `teks_rumpang` | Proporsional per posisi/kolom: `bobot × benar / total` |
| `mengurutkan` | Proporsional per posisi benar. Kunci `{"urutan": [indeks…]}`; jawaban array indeks ATAU array string item; fallback kunci = urutan string |
| `uraian` (esai, essay, isian panjang, upload_file, praktik_singkat) | Manual: `(0, auto=false)` → `poin = NULL`, menunggu koreksi guru |
| `skala_sikap`, `studi_kasus` (wadah) | `(0, auto=true)`, `Dinilai = false` — tidak masuk penyebut nilai |
| `setuju_tidak_setuju` | `Dinilai` hanya bila kunci ada dan tidak kosong |
| Tipe tak dikenal | `(0, auto=false)` — diperlakukan manual agar guru memeriksa, bukan diberi nilai palsu |
| Jawaban kosong / `null` | `(0, auto=true)` |

**Nilai akhir** — `grading.HitungNilaiAkhir`: `nilai = Σ poin / Σ bobot × 100`, **hanya soal dengan `Dinilai = true`**; `menunggu_koreksi = true` bila ada soal manual belum dikoreksi; hasil dijepit ke `[0, 100]`; bila tidak ada soal dinilai → `0`.

**Agregasi server** (`exam.go: hitungNilai`): baca semua baris `jawaban` attempt (jawaban `NULL` → poin 0; uraian belum dikoreksi → poin 0 sementara + `menunggu = true`) **ditambah** soal paket tanpa baris jawaban sama sekali (poin 0, masuk penyebut bila `Dinilai`).

**Sinkronisasi nilai** (`syncNilaiAkhir`, best-effort, gagal sync tidak menggagalkan ujian): salin nilai attempt ke tabel `nilai_akhir` (`attempt_id, jadwal_id, siswa_id` — dipetakan lewat `siswa.user_id = attempt.siswa_id` karena `attempt.siswa_id` menyimpan `users.id` — dengan status `menunggu_koreksi`). Status `nilai_akhir` baru menjadi `final` lewat finalisasi rekap (di luar Alur B), dan sync tidak pernah menurunkan status `final` yang sudah ada.

**Koreksi esai (guru/admin)** — `GET /api/v1/koreksi/:jadwalId` (role admin/guru): daftar jawaban dengan `dikoreksi = FALSE`, `jawaban IS NOT NULL`, tipe manual (`uraian, upload_file, praktik_singkat, esai, essay, isian_panjang, …`). `POST /api/v1/koreksi/:jawabanId` `{skor: 0–100}` → `poin = skor/100 × bobot`, `dikoreksi = TRUE`; bila attempt sudah `selesai` → hitung ulang nilai akhir + update `attempt.nilai_akhir` + sync; bila attempt masih berjalan → nilai akhir dihitung saat submit. Respons menyertakan `nilai_akhir_baru` + `final` bila ada.

**Remedial (staf, via rekap)** — `POST /api/v1/rekap/:jadwalId/remedial` `{siswa_id, nilai_baru}` (permission `rekap.remedial`): `INSERT INTO remedial (jadwal_id, siswa_id, nilai)` lalu `nilai_akhir = MAX(nilai_awal, nilai_baru)` + predikat diperbarui. **Remedial di kode adalah entri nilai manual (nilai tertinggi yang menang), BUKAN ujian ulang** — tidak ada attempt baru, tidak ada soal, tidak ada token.

### B.9 Auto-submit saat deadline lewat (lazy — tanpa job latar)

**Tidak ada scheduler/cron/job latar di kode untuk auto-submit.** Finalisasi otomatis terjadi secara *lazy* lewat fungsi `autoSubmitJikaLewat(ctx, pool, attempt, jadwalStatus)`, yang dipanggil di awal handler: `getAttemptDetail`, `simpanJawaban`, `tandaiRagu`, `statusAttempt` (dan di dalam `simpanJawaban`/`tandaiRagu`/`statusAttempt` setelah cek status awal).

Logika `autoSubmitJikaLewat`:
- Guard: `attempt.status ≠ "berjalan"` → tidak ada apa-apa. **`jadwal_status === "pause"` → tidak auto-submit** (timer beku; deadline lama yang sudah lewat diabaikan).
- Bila `now ≥ deadline` → `finalisasiAttempt`: `hitungNilai` → `UPDATE attempt SET status='selesai', submit_at=now(), nilai_akhir=nilai` → `syncNilaiAkhir` → kembalikan attempt terbaru + `true`.
- Konsekuensi: **bila siswa menutup browser tepat saat deadline lewat dan tidak pernah mengirim request lagi, attempt tetap berstatus `berjalan` di DB** sampai ada request berikutnya yang menyentuhnya (atau sampai proktor/rekap memprosesnya). Tidak ada proses server yang menyapu attempt kedaluwarsa sendiri.

### B.10 Submit manual + tampil hasil

**Submit manual:**
1. Siswa klik **Selesai Ujian** → Modal konfirmasi "Kumpulkan Jawaban?" (menampilkan jumlah dijawab/belum; soal belum dijawab dinyatakan "dihitung 0"; peringatan "Setelah dikumpulkan, jawaban tidak bisa diubah.") → **Ya, Kumpulkan** → `doSubmit()`.
2. `doSubmit` dijaga `sudahSubmitRef` (klik ganda tidak mengirim dua submit); gagal → guard dilepas + pesan error.
3. `POST /api/v1/ujian/attempt/:id/submit` `{}`. Server `submitUjian`:
   - Milik sendiri; `diblokir` → `403`; `selesai` → **idempotent**: kembalikan hasil tersimpan (nilai hanya bila tampil-hasil efektif).
   - Selain itu → `finalisasiAttempt` (sama seperti auto-submit) → audit `ujian.submit` → broadcast → respons `{attempt_id, status: "selesai", menunggu_koreksi, nilai_akhir}` dengan **`nilai_akhir = null` + `pesan: "Nilai disembunyikan sampai koreksi final."` bila tampil-hasil TIDAK efektif**.
   - Tampil-hasil efektif = `paket.tampil_hasil AND settings.tampil_hasil_ke_siswa` (setting default `false` — jadi defaultnya nilai disembunyikan kecuali kedua saklar aktif).
4. Frontend menampilkan kartu "Ujian Selesai / Jawaban Anda sudah tersimpan." + nilai (bila tidak null) atau teks "Nilai disembunyikan sampai koreksi final." (+ "(Masih ada soal uraian yang menunggu koreksi guru.)" bila `menunggu`) + tombol **Lihat Hasil** (`/hasil/:attemptId`) dan **Kembali ke Beranda**.

**Halaman hasil** — `GET /api/v1/ujian/attempt/:id/hasil` (hanya milik sendiri, hanya bila `status === "selesai"`, selain itu `409`):
- `tampil = tampilHasilEfektif(paket) OR paket_jenis === "Tryout"` — **Tryout selalu menampilkan hasil lengkap** (kunci + pembahasan).
- Per soal (diurutkan sesuai `urutan_soal` attempt): `jawaban_siswa`, `poin`, `bobot`, `ragu`, `status` (`benar` bila poin ≥ bobot; `sebagian` bila 0 < poin < bobot; `salah` bila poin = 0; `menunggu_koreksi` bila poin NULL; `disembunyikan` bila tidak tampil). Bila tampil: `kunci_tampil` (ringkasan kunci per tipe, mis. "Jawaban benar: B — …", "Pasangan benar: …", "Dinilai manual oleh guru (rubrik).", "Tidak dinilai (survei).") + `pembahasan` (kolom `soal.pembahasan`).
- Soal paket **tanpa baris jawaban** (tidak dijawab sama sekali) tetap ditampilkan: status `salah`, poin 0 (bila tampil) atau `disembunyikan` (bila tidak tampil).
- `nilai_akhir` hanya diisi bila tampil dan tersimpan; selain itu `null`.
- `HasilPage` menampilkan nilai (bila tidak null) atau "Nilai disembunyikan sampai koreksi final.", lalu per soal: pertanyaan + jawaban siswa (map Menjodohkan dirender per pasangan `kiri → kanan`) + poin + kunci + pembahasan — tiga terakhir hanya bila `tampil_hasil` dari server true.

### B.11 Daftar cabang / guard Alur B (ringkas)

| Kondisi | Efek |
|---|---|
| Kelas siswa tidak ditemukan | Daftar ujian kosong + pesan "Hubungi admin/proktor" |
| Jadwal belum mulai / sudah berakhir | `409`, tidak bisa mulai |
| Mulai lewat toleransi telat (default 15 mnt) | `403`, hanya untuk attempt baru |
| Token kosong / hash tidak cocok | `401 token_invalid` |
| Token kedaluwarsa (default 15 mnt sejak generate) | `409 token_expired` |
| Paket tanpa soal | `409` |
| Attempt berjalan sudah ada | Kembalikan attempt lama (idempotent, tanpa token) |
| Attempt diblokir / selesai | `403` / `409`, tidak bisa mulai |
| Jadwal `pause` | Mulai ditolak; simpan jawaban ditolak; auto-submit ditunda; polling mengembalikan `jadwal_status: "pause"` |
| Simpan jawaban pada attempt `selesai` / `diblokir` | `409` / `403` |
| Soal bukan anggota paket | `400` |
| Deadline lewat saat simpan/ragu/status/detail | Auto-finalisasi dulu, lalu request ditolak / status terbaru dikembalikan |
| Submit pada attempt `selesai` | Idempotent, kembalikan hasil tersimpan |
| Submit pada attempt `diblokir` | `403` |
| Hasil diminta sebelum selesai | `409` |
| Nilai tampil | Hanya bila `paket.tampil_hasil AND settings.tampil_hasil_ke_siswa`, atau Tryout (hasil lengkap) |

---

## Alur C — Pemantauan + anti-curang

### C.0 Partisipan

| Partisipan | Peran |
|---|---|
| Proktor / admin | Membuka monitor, pause/resume jadwal, unblock peserta |
| Pengawas | Memantau (role boleh monitor); **mengusulkan** unblock (hanya lewat API — lihat §C.5) |
| Guru | Memantau (role boleh monitor) |
| Siswa | Sumber event pelanggaran (pindah tab / blur / keluar fullscreen) |
| `MonitorPage.tsx` | Klien WebSocket + fallback polling + aksi pause/resume/unblock |
| Backend `monitor.go` | Ticket sekali pakai, Hub room per jadwal, snapshot, broadcast, heartbeat |
| Backend `exam.go` | `catatPelanggaran`, `pauseJadwal`, `resumeJadwal`, `createUsulanUnblock`, `listUsulanUnblock`, `unblockAttempt`, `broadcastAttempt` |
| PostgreSQL | `pelanggaran`, `usulan_unblock`, `attempt`, `jadwal_ujian`, `jawaban` |

### C.1 Koneksi monitor — ticket + WebSocket

1. Staf membuka `/adm/monitor/:jadwalId` (atau `/adm/monitor` lalu memilih jadwal) → `MonitorPage` langsung `muatSnapshot()` (REST) + memulai `sambung()`.
2. `POST /api/v1/monitor/ticket` `{jadwal_id}` — guard: role harus `admin / proktor / guru / pengawas` atau permission `ujian.monitor`; jadwal harus ada (bila tidak → `404`). Server membuat **ticket sekali pakai, berlaku 30 detik** (token acak 24 byte hex, disimpan in-memory bersama `jadwal_id + user_id + role`) → respons `{ticket, jadwal_id, expires_in: 30, expires_at}`.
3. Frontend membuka `new WebSocket((wss|ws)://host/ws/monitor?ticket=…)`.
4. Server `GET /ws/monitor?ticket=`: `tickets.pakai(token)` — **validasi + konsumsi atomik**: ticket tidak ada / sudah dipakai / kedaluwarsa → `401`; role di ticket bukan pemantau → `403`. Lalu `websocket.Accept` (coder/websocket, `OriginPatterns: ["*"]`), subscribe klien ke room `jadwal_id` di `Hub` global (buffer kirim per klien 32 pesan; klien lambat dilewati agar broadcast tidak macet).
5. **Snapshot awal:** server langsung mengirim `{type: "snapshot", data: {jadwal_id, generated_at, peserta[]}}` lewat koneksi WS. Isi per peserta: `attempt_id, siswa_id, nama_siswa, nisn, kelas_nama, status, mulai_at, deadline, sisa_waktu_detik (beku bila pause, atau deadline − now), nilai_akhir, dijawab (jawaban IS NOT NULL), ragu (ragu = TRUE), pelanggaran (COUNT), total_soal`.
6. **Heartbeat server:** goroutine penulis mengirim `{"type":"ping"}` tiap **30 detik** + meneruskan pesan antrean `send`.
7. **Loop baca server:** mendeteksi disconnect (read error → return → unsubscribe + `conn.Close`); bila klien mengirim pesan `{type: "snapshot"}` → server mengirim snapshot baru. **Frontend `MonitorPage` tidak pernah mengirim pesan `snapshot`** — yang ada hanya reconnect (lihat §C.2).
8. **Reconnect + fallback (frontend):** `ws.onclose` → `setWsAktif(false)` → `setTimeout(sambung, 5000)` (minta ticket BARU — ticket lama sudah terpakai/hangus); `ws.onerror` → `ws.close()` (memicu jalur reconnect yang sama). **Bila WS tidak aktif, fallback polling REST `GET /api/v1/monitor/:jadwalId` tiap 5 detik** (interval hanya jalan saat `!wsAktif`); label UI berubah "Real-time (WebSocket)" ↔ "Fallback polling 5 dtk". Unmount → batalkan reconnect + tutup WS.
9. **Usulan unblock** di-poll terpisah tiap **10 detik** (`GET /api/v1/jadwal/:jadwalId/usulan-unblock`).

### C.2 Siaran status (broadcast) — dari ujian ke monitor

`monitor.Broadcast(jadwalID, msg)` mengirim JSON ke semua klien di room jadwal itu. Pemancar (di `exam.go`):

| Pemicu | Pesan broadcast |
|---|---|
| Jawaban tersimpan (`simpanJawaban`) | `{type: "update", jadwal_id, attempt_id, status, dijawab, ragu, pelanggaran, at}` via `broadcastAttempt` |
| Ragu diubah (`tandaiRagu`) | `update` (sama) |
| Pelanggaran dicatat (`catatPelanggaran`) | `update` (sama) |
| Submit / finalisasi (`submitUjian`) | `update` (sama) |
| Koreksi disimpan (`koreksiJawaban`) | `update` (sama) |
| Ujian dimulai (`mulaiUjian`) | `{type: "ujian_mulai", jadwal_id, attempt_id}` |
| Jadwal di-pause | `{type: "jadwal_pause", jadwal_id}` |
| Jadwal di-resume | `{type: "jadwal_resume", jadwal_id}` |
| Usulan unblock dibuat | `{type: "usulan_unblock", jadwal_id, attempt_id}` |
| Unblock berhasil | `update` (via `broadcastAttempt` pada attempt terbaru) |

Perilaku frontend atas pesan WS (`MonitorPage.ws.onmessage`):
- `snapshot` → ganti daftar peserta dari `msg.data.peserta`.
- `update` → **muat ulang snapshot lewat REST** (`muatSnapshot()`) — update ringan tidak dipakai langsung untuk memperbarui baris.
- `jadwal_pause` / `jadwal_resume` → `jadwalQuery.refetch()` + `muatSnapshot()`.
- `ping` / non-JSON → diabaikan.
- **Tidak ada WS ke arah siswa.** Siswa tidak menerima siaran apa pun; status blokir/pause/selesai diketahui siswa lewat polling `/status` 30 detik atau lewat request berikutnya (simpan/ragu/detail).

### C.3 Anti-curang — pindah tab / blur / keluar fullscreen

**Sisi siswa (`UjianPage`):** listener aktif selama ada `attemptId` dan belum `hasilSubmit`:
- `document.visibilitychange` + `document.hidden` → lapor `"pindah_tab"`.
- `window.blur` → lapor `"blur"`.
- `document.fullscreenchange` + `!document.fullscreenElement` + `fullscreenAktifRef` → lapor `"fullscreen_exit"` (hanya bila fullscreen sempat aktif).
- `laporPelanggaran(jenis)`: guard — tidak ada attempt / sudah submit → batal; **dedupe client 5 detik** (`terakhirLaporRef`: laporan dalam 5 detik sejak laporan sebelumnya dibuang) → `POST /api/v1/ujian/attempt/:id/pelanggaran` `{jenis}`.
- Respons: `diblokir: true` → tampil layar diblokir; `peringatan: true` → tampil banner peringatan (`jumlah_pelanggaran`); error jaringan → diabaikan diam-diam.
- **Peristiwa `copy/paste/cut` pada kolom soal dicegah** (`onCopy/onPaste/onCut → preventDefault`) — pencegahan di sisi tampilan saja, bukan pelanggaran tercatat.

**Sisi server (`catatPelanggaran`):**
1. Role siswa; attempt milik sendiri (bukan → 404); `selesai` → `409`; `diblokir` → `200 {diblokir: true, pesan: "Attempt sudah diblokir…"}` (idempotent).
2. **Deduplikasi server 5 detik:** bila pelanggaran terakhir attempt ini < 5 detik lalu → `200 {diabaikan_duplikat: true, jumlah_pelanggaran: jumlah_lama, peringatan: (jumlah == 1), diblokir: false}` — tanpa INSERT baru. (Alasan di kode: `blur` + `visibilitychange` sering berapi hampir bersamaan.)
3. Selain itu: `INSERT INTO pelanggaran (attempt_id, jenis, detail)`; `jumlah++`; audit `ujian.pelanggaran`.
4. **Batas:** `max_pelanggaran` dari settings (default **2**; bila setting < 1 → 2).
   - `jumlah < batas` → broadcast `update` → `200 {peringatan: true, diblokir: false, jumlah_pelanggaran, pesan: "Peringatan: … Pelanggaran berikutnya akan memblokir ujian Anda."}`. Untuk pelanggaran pertama, `jumlah = 1`.
   - `jumlah ≥ batas` → **auto-block**: `UPDATE attempt SET status = 'diblokir', sisa_waktu_beku = GREATEST(0, deadline − now())` (sisa waktu dibekukan agar unblock dapat melanjutkan dari sisa yang sama) → broadcast → `200 {diblokir: true, jumlah_pelanggaran, pesan: "Anda diblokir karena pelanggaran berulang. Hubungi proktor."}`.
5. **Tampilan:** banner peringatan (`BannerPeringatan mode="peringatan"`, satu baris, tombol "Mengerti" menutup banner) dan layar diblokir (`mode="diblokir"`, teks: "Jawaban Anda tersimpan. Silakan hubungi proktor / pengawas untuk meminta unblock — ujian akan dilanjutkan dari sisa waktu Anda."). UjianPage meng-hardcode `jumlahPelanggaran={2}` pada tampilan diblokir (bukan dari server — server mengembalikan jumlah sebenarnya, tetapi tampilan tidak memakainya).

### C.4 Pause jadwal — timer beku

**Aksi:** `MonitorPage` tombol Pause/Resume → `POST /api/v1/jadwal/:id/pause` atau `/resume` (juga tersedia di `JadwalPage`).
**Server:**
- `pauseJadwal` (role admin/proktor): jadwal harus `aktif` (selain itu `409`). Transaksi: `UPDATE jadwal_ujian SET status='pause'` + `UPDATE attempt SET sisa_waktu_beku = GREATEST(0, deadline − now()) WHERE jadwal_id AND status='berjalan'`. Audit `jadwal.pause` + broadcast `jadwal_pause`.
- `resumeJadwal` (role admin/proktor): jadwal harus `pause` (selain itu `409`). Transaksi: `UPDATE jadwal_ujian SET status='aktif'` + `UPDATE attempt SET deadline = now() + sisa_waktu_beku, sisa_waktu_beku = NULL WHERE jadwal_id AND status='berjalan' AND sisa_waktu_beku IS NOT NULL`. Audit `jadwal.resume` + broadcast `jadwal_resume`.
- **Efek selama pause (sudah diverifikasi di kode):** `sisaWaktu()` mengembalikan nilai beku; `autoSubmitJikaLewat` dilewati; `simpanJawaban` ditolak `409`; `mulaiUjian` ditolak `409`; `UjianPage` menampilkan Alert "Ujian dijeda proktor — timer dibekukan. Tunggu hingga dilanjutkan." (dari `jadwal_status` polling/detail); `MonitorPage` menonaktifkan tombol Pause bila status ≠ `aktif` dan tombol Resume bila status ≠ `pause`.
- **Catatan:** deadline attempt yang `diblokir` saat pause TIDAK ikut dibekukan/di-resume oleh pause/resume (klausa `status='berjalan'`); sisa waktu attempt diblokir dibekukan tersendiri saat auto-block (§C.3) dan dipakai saat unblock (§C.5).

### C.5 Usul unblock + unblock — resume sisa waktu

**Usulan (pengawas/proktor/admin)** — `POST /api/v1/jadwal/:id/usulan-unblock` `{attempt_id, alasan}`:
- Role `pengawas / proktor / admin`; attempt harus dari jadwal itu (`400` bila bukan); attempt harus `diblokir` (selain itu `409`); tidak boleh ada usulan `menunggu` yang sama untuk attempt itu (`409`); `alasan` wajib non-kosong (`400`).
- `INSERT INTO usulan_unblock (attempt_id, pengawas_id = user pemohon, alasan, status='menunggu')` → audit `jadwal.usulan_unblock` → broadcast `usulan_unblock` → `201 {id, attempt_id, status: "menunggu"}`.
- **Daftar usulan:** `GET /api/v1/jadwal/:id/usulan-unblock` (role admin/proktor/pengawas) → semua usulan jadwal itu (urut `created_at DESC`), termasuk `status, diproses_oleh, diproses_at, nama_siswa`. `MonitorPage` menampilkan hanya yang `status === "menunggu"` (di-poll 10 detik), sebagai teks "nama — alasan" tanpa tombol aksi.

**Unblock (proktor/admin)** — `POST /api/v1/jadwal/:id/unblock` `{attempt_id, alasan}` (dari `MonitorPage`: kolom Aksi pada baris peserta `diblokir` — input alasan wajib diisi di client, lalu POST):
- Role `proktor / admin` (pengawas TIDAK bisa unblock langsung — hanya mengusulkan).
- Attempt harus dari jadwal itu; harus `diblokir` (selain itu `409`); `alasan` wajib.
- **Sisa waktu lanjutan:** bila `attempt.sisa_waktu_beku` ada → pakai itu; bila tidak → `sisa = jam_selesai_jadwal − now`. Bila `sisa ≤ 0` → `409` ("Ujian sudah berakhir… Unblock tidak bisa dilakukan.").
- Transaksi: `UPDATE attempt SET status='berjalan', deadline = now() + sisa, sisa_waktu_beku = NULL` — **jawaban tersimpan TIDAK di-reset** — lalu: bila ada usulan `menunggu` terbaru → `status='disetujui', diproses_oleh, diproses_at=now()`; bila tidak ada usulan (proktor unblock langsung) → dibuat baris `usulan_unblock` dengan `status='disetujui'` sebagai jejak audit.
- Audit `jadwal.unblock` (dengan `sisa_waktu_detik`) → broadcast `update` → `200 {attempt_id, status: "berjalan", sisa_waktu_detik, deadline}`.
- **Sisi siswa:** unblock diketahui lewat polling `/status` berikutnya (status kembali `berjalan`, `diblokir=false`) atau lewat memuat ulang detail attempt; tidak ada notifikasi push ke siswa. Setelah unblock, siswa melanjutkan dari soal terakhir (jawaban + ragu tersimpan utuh).

### C.6 Daftar cabang / guard Alur C (ringkas)

| Kondisi | Efek |
|---|---|
| Ticket monitor tidak ada / sudah dipakai / kedaluwarsa (> 30 dtk) | `401` WS |
| Role ticket bukan pemantau | `403` WS |
| WS putus | Reconnect 5 dtk dengan ticket baru; fallback polling REST 5 dtk selama WS mati |
| Pelanggaran < 5 dtk sejak yang sebelumnya (client & server) | Dibuang (dedupe) |
| Pelanggaran ke-1 (di bawah batas) | Peringatan, ujian jalan |
| Pelanggaran ke-2 (≥ batas default 2) | Auto-block + sisa waktu dibekukan |
| Pelanggaran pada attempt selesai / diblokir | `409` / `200 diblokir: true` (tanpa INSERT) |
| Pause jadwal bukan `aktif` / resume bukan `pause` | `409` |
| Usulan unblock: attempt tidak diblokir / sudah ada usulan menunggu / alasan kosong | `409` / `409` / `400` |
| Unblock: attempt tidak diblokir / sisa waktu ≤ 0 / alasan kosong / role pengawas | `409` / `409` / `400` / `403` (pengawas hanya boleh mengusulkan) |

---

## 5. Hal yang TIDAK ditemukan di kode (jangan dianggap ada)

1. **Tidak ada background job/scheduler/cron** untuk auto-submit, pembersihan session, atau penyapuan attempt kedaluwarsa di paket yang dibaca (`exam.go`, `monitor.go`, `session.go` — tidak ada ticker/goroutine periodik kecuali heartbeat WS 30 detik di `monitor.go`). Auto-submit murni lazy per-request (§B.9).
2. **Tidak ada WebSocket / push ke siswa.** Satu-satunya WS adalah `/ws/monitor` untuk pemantau. Siswa mengetahui blokir/pause/selesai hanya lewat polling `/status` 30 detik atau request berikutnya.
3. **Tidak ada UI frontend untuk membuat usulan unblock.** Endpoint `POST …/usulan-unblock` ada di backend, tetapi tidak ada satu pun `POST …/usulan-unblock` di kode frontend; `MonitorPage` hanya `GET` daftar usulan (tampil sebagai teks, tanpa tombol setuju/tolak). Pengawas tidak punya tombol "usulkan unblock" di UI yang dibaca. (Proktor/admin melakukan unblock langsung dari `MonitorPage`.)
4. **Tidak ada endpoint/tombol "minta unblock" untuk siswa.** Siswa yang diblokir hanya melihat pesan "hubungi proktor/pengawas" — tidak ada aksi yang bisa dikirim dari halaman ujian.
5. **Tidak ada retry autosave.** Jawaban yang gagal tersimpan (jaringan putus) tidak dikirim ulang otomatis; `statusSimpan` hanya menampilkan "gagal".
6. **Tidak ada pemakaian `GET /api/v1/auth/csrf` oleh frontend.** Token CSRF selalu dibaca dari cookie `cbt_csrf`.
7. **Tidak ada pemakaian `state.from` untuk redirect balik sesudah login.** `RequireAuth` menyimpan `from`, tetapi `LoginPage` mengabaikannya dan selalu navigasi berdasarkan role (`siswa → /`, staf → `/adm`).
8. **Pesan sukses ganti password tidak ditampilkan.** `GantiPasswordPage` mengirim `state.pesan` ke `/login`, tetapi `LoginPage` tidak membaca `location.state` — pesan itu hilang.
9. **`useExamStore` (`store/exam.ts`) tidak dipakai oleh `UjianPage`.** State ujian (jawabanMap, raguMap, attemptId) dikelola lokal di `UjianPage` dengan `useState`; store tersebut tampak mati (dead code) untuk alur ini.
10. **Tidak ada token ulang / kirim ulang token ke siswa.** Token hanya tampil sekali di `JadwalPage` sesudah generate; bila hilang, satu-satunya jalan adalah generate token baru (yang menghanguskan token lama).
11. **Tidak ada pengecekan CSRF pada `ganti-password` mode kredensial** (disengaja: tanpa cookie session, middleware CSRF melewatkan request; keamanannya mengandalkan verifikasi `password_lama`).
12. **Tidak ada batas jumlah peserta / validasi `durasi_menit` terhadap deadline** di `mulaiUjian` — deadline selalu jam selesai jadwal.
13. **Tidak ada `DELETE`/pencabutan ticket monitor** selain konsumsi sekali pakai + kedaluwarsa 30 detik.
14. **Tidak ada pembedaan `detail` pelanggaran di sisi client** — `UjianPage` selalu mengirim `{jenis}` tanpa `detail` (field `detail` backend tetap ada, diisi `NULL`).

## 6. Perbedaan terhadap diagram Sequence lama (berdasarkan `DIAGNOSIS-SEQUENCE-JELAS.md` + kode)

Diagram lama yang dibandingkan: Sequence 3a (login + ganti password) dan 3b (autosave + timer + auto-submit). Catatan: perbandingan ini hanya soal **isi alur**, bukan soal render StarUML (penomoran ekspor, self-message melayang) yang sudah didiagnosis di dokumen tersebut.

| # | Yang digambarkan diagram lama | Yang sebenarnya terjadi di kode | Dampak untuk perbaikan diagram |
|---|---|---|---|
| 1 | 3a: endpoint ganti password ditulis `POST /api/v1/auth/change-password` | Endpoint sebenarnya `POST /api/v1/auth/ganti-password` (body `login, password_lama, password_baru`) | Nama endpoint salah — perbaiki label |
| 2 | 3a: tidak ada cabang rate-limit / akun nonaktif / kredensial salah | Ada: 5 gagal → kunci 15 mnt (`429 login_locked`); `is_active=false` → `403 account_inactive`; kredensial salah → `401 invalid_credentials`; ketiganya mendahului cek `must_change_password` | Tambah cabang-cabang ini (atau frag `alt`) |
| 3 | 3a: alur berhenti di "redirect ke dashboard/ujian" sesudah `200 + session cookie + CSRF` | Sesudah login masih ada: simpan `user + permissions` ke store → navigasi **berdasarkan role** (`siswa → /`, staf → `/adm`); dan saat aplikasi dibuka, `fetchMe` (`GET /auth/me`) mendahului segalanya; guard rute `RequireAuth` + `/403` untuk siswa ke `/adm` | Tambah langkah store + navigasi per role + `fetchMe` awal |
| 4 | 3a: `INSERT sessions` digambarkan hanya pada login ulang sesudah ganti password | `INSERT sessions` terjadi pada **setiap** login sukses (termasuk yang pertama); sesudah ganti password, yang terjadi adalah **penghapusan SEMUA session** (`DestroyAllForUser`) + hapus cookie, lalu login ulang (dengan INSERT baru) | Perbaiki: ganti-password → DELETE semua session + clear cookie → login ulang → INSERT session |
| 5 | 3a: tidak ada CSRF | Ada: cookie `cbt_csrf` + header `X-CSRF-Token` untuk mutasi; `403 csrf_token_invalid` bila tidak cocok; login + csrf exempt | Tambah partisipan/langkah CSRF |
| 6 | 3a: tidak ada TTL/idle session | Session TTL 12 jam, idle timeout 2 jam (`last_seen_at`, `Touch` per request terautentikasi) | Tambah catatan guard kedaluwarsa |
| 7 | 3b: dimulai langsung dari "pilih/ubah jawaban" (autosave) | Sebelum autosave ada alur lengkap yang hilang di diagram: `GET ujian/aktif` → token → `POST ujian/mulai` (dengan 14 pemeriksaan berurutan, termasuk toleransi telat + token + idempotent attempt berjalan) → `GET attempt/:id` (muat soal tanpa kunci) | Diagram ujian perlu didahului frag mulai/token/muat-soal, atau dibuat diagram terpisah |
| 8 | 3b: `hitungNilai: HitungPoin per soal` digambarkan terjadi saat auto-submit | `HitungPoin` per soal terjadi **saat simpan jawaban** (autosave); saat finalisasi (submit/auto-submit) yang terjadi adalah **agregasi** `hitungNilai` atas `poin` tersimpan + soal tanpa jawaban | Pindahkan `HitungPoin` ke langkah autosave; finalisasi hanya agregasi + `UPDATE attempt` + `syncNilaiAkhir` |
| 9 | 3b: `cek deadline lewat (autoSubmitJikaLewat)` digambarkan sebagai satu langkah sesudah polling status | `autoSubmitJikaLewat` dipanggil di awal **empat** handler (detail, simpan, ragu, status) — bukan hanya sesudah polling; dan dilewati bila jadwal `pause` | Gambarkan sebagai guard berulang / frag `opt` di tiap handler, dengan guard `[jadwal_status ≠ pause]` |
| 10 | 3b: auto-submit digambarkan sebagai aksi proaktif server | Auto-submit **lazy**: tidak ada job latar; attempt tetap `berjalan` sampai ada request berikutnya | Tambah catatan "tidak ada scheduler; finalisasi terjadi pada request berikutnya" |
| 11 | 3b: `200 hasil + nilai (final bila tanpa esai)` sebagai balasan langsung alur timer | Balasan status-polling saat selesai hanya `{status, nilai_akhir (bila tampil), menunggu_koreksi}`; halaman hasil adalah request terpisah `GET …/hasil` dengan aturan tampil sendiri (Tryout selalu tampil; esai → `menunggu_koreksi`; nilai disembunyikan bila `tampil_hasil` tidak efektif) | Pisahkan "status polling" dari "halaman hasil"; tambah aturan tampil-hasil |
| 12 | 3b: `broadcastAttempt (WS: status peserta berubah)` digambarkan sebagai pengiriman langsung ke "Monitor WebSocket" | Broadcast adalah pesan `update` ringan ke room monitor; `MonitorPage` menanggapinya dengan **memuat ulang snapshot lewat REST**; pesan `snapshot` penuh hanya dikirim saat koneksi awal (dan bila klien mengirim `{type:"snapshot"}`, yang tidak pernah dilakukan frontend) | Perbaiki arah/isi pesan: `update` ringan → monitor refetch REST |
| 13 | 3b: tidak ada ragu-ragu, pelanggaran, blokir, pause, unblock | Semua ada sebagai alur sendiri (lihat §B.5, §C.3–C.5): endpoint `ragu` terpisah; pelanggaran dedupe 5 dtk (client + server), peringatan ke-1, auto-block ke-2 + bekukan sisa; pause beku + tolak simpan; unblock lanjut sisa waktu | Tambah alur-alur ini (atau diagram terpisah Alur C) |
| 14 | 3a/3b: deadline digambarkan seolah dari durasi / timer attempt | Deadline = **jam selesai jadwal**; `durasi_menit` tidak dipakai menghitung deadline | Perbaiki sumber deadline |
| 15 | 3b: tidak ada `syncNilaiAkhir` ke tabel `nilai_akhir` / remedial / koreksi | Ada: `syncNilaiAkhir` (status `menunggu_koreksi` sampai finalisasi rekap); koreksi esai `POST /koreksi/:jawabanId` (skor 0–100 → poin); remedial = entri nilai manual `MAX(awal, baru)`, bukan ujian ulang | Tambah bila diagram mencakup pasca-submit |

## 7. Hal yang tidak pasti / perlu konfirmasi (tidak ditebak)

1. **Urutan middleware global di `main`/bootstrap server** (apakah `RequireAuth`/`CSRF`/`RateLimit` dipasang dalam urutan apa, dan apakah ada pengecualian tambahan di level bootstrap) **tidak ditelusuri** — hanya isi masing-masing middleware yang dibaca. Urutan eksekusi antar-middleware tidak diklaim di dokumen ini.
2. **Isi `docs/KONTRAK.md` / `docs/KEAMANAN.md`** tidak dipakai sebagai sumber — bila ada perbedaan antara kontrak tertulis dan kode, yang ditulis di sini adalah **kode**.
3. **Perilaku `SiswaHomePage` untuk staf** yang membuka `/` (rute `/` mengizinkan semua role yang login) tidak ditelusuri sampai ke render per-role; hanya guard rutenya yang dicatat.
4. **Migrasi SQL** (nama tabel/kolom persis, constraint, index) tidak dibaca — nama tabel/kolom di dokumen ini berasal dari query di kode Go, bukan dari file migrasi.
5. **Halaman `JadwalPage`** hanya dibaca bagian token; CRUD jadwal, pause/resume dari halaman itu, dan validasi form jadwal tidak ditelusuri.
6. **Nilai `settings` di database produksi** (token_expiry_detik, toleransi_telat_menit, max_pelanggaran, tampil_hasil_ke_siswa) tidak diperiksa — yang ditulis adalah **nilai default di kode** bila setting tidak ada.

---

# PERLUASAN FASE 1b — Telusur SEMUA Use Case (18 Okt 2026)

> **Status:** Perluasan dokumen telusur Fase 1. Bagian Alur A/B/C di atas TIDAK diubah. Bagian ini menambahkan Alur D–R untuk UC-UC di luar A/B/C.
> **Aturan yang sama:** nama teknis kode ditulis apa adanya; translasi ke bahasa desain (PIM) dilakukan fase bangun diagram, bukan di sini.
> **Pemetaan ke Alur A/B/C:** UC01 → Alur A; UC08 → Alur B; UC07 + UC18 → Alur C. UC-UC itu TIDAK ditelusuri ulang di sini kecuali penambahan yang dinyatakan eksplisit (UC09, UC10, UC19 — yang di Alur B hanya tercakup sebagian).

## 8. Inventaris 19 Use Case (dari diagram Use Case di `cbt-smk-v2.mdj`, read-only)

Dokumentasi (`documentation`) tiap UC disalin dari `.mdj`. Aktor diambil dari asosiasi di diagram Use Case yang sama.

| Kode | Nama UC (persis `.mdj`) | Aktor di diagram | Dokumentasi `.mdj` (ringkas) | Ditelusuri di |
|---|---|---|---|---|
| UC01 | Masuk & Wajib Ganti Kata Sandi | Admin, Proktor, Guru, Pengawas, Siswa, Penguji UKK | Masuk dengan sesi; bila wajib ganti kata sandi, sistem menolak masuk dan menuntut ganti kata sandi dulu. | Alur A |
| UC02 | Kelola Master Data | Admin | CRUD Jurusan, Kelas, Mapel, Guru, Siswa (+ Impor Siswa). Hapus permanen, ditolak bila data masih dipakai. | Alur D (baru) |
| UC03 | Kelola Bank Soal | Admin, Guru | CRUD Bank Soal + Soal (16 tipe). Bank terkunci tidak bisa diubah. Proktor DILARANG lihat isi soal. | Alur E (baru) |
| UC04 | Impor Soal | Admin, Guru | Impor Excel & Word via pratinjau tanpa menyimpan, templat, validasi E1–E32/W1–W7, mode semua-atau-tidak-sama-sekali / lewati-yang-salah. | Alur F (baru) |
| UC05 | Susun Paket Ujian | Admin, Proktor, Guru | CRUD Paket (STS/SAS/USBK/UKK/Tryout) + tambah/hapus soal ke paket, pembangkitan acak, sinkron jumlah soal. | Alur G (baru) |
| UC06 | Kelola Jadwal Ujian | Admin, Proktor | CRUD Jadwal, jeda/lanjut (jeda membekukan pewaktu), buat token (token baru menghanguskan yang lama, berlaku 900 detik). Pengawas DILARANG lihat token. | Alur H (baru; pause/resume + token sudah di B/C — dirujuk) |
| UC07 | Pantau Ujian Langsung | Admin, Proktor, Pengawas | Pantau peserta per jadwal secara langsung, status pengerjaan, pelanggaran, buka blokir. | Alur C |
| UC08 | Ikuti Ujian | Siswa | Lihat ujian aktif, masukkan token, kerjakan soal, simpan otomatis, tandai ragu, kirim / kirim otomatis saat batas waktu. | Alur B |
| UC09 | Validasi Token | (tidak ada asosiasi aktor di diagram — UC tanpa aktor) | Validasi token jadwal: token cocok, belum kedaluwarsa (900 detik), dalam toleransi telat 15 menit. | Alur I (baru; langkah di dalam `mulaiUjian`, §B.2) |
| UC10 | Koreksi Esai / Manual | Guru | Guru mengoreksi jawaban uraian/manual (daftar koreksi, koreksi jawaban), poin manual, status dikoreksi. | Alur J (baru; ringkasan di §B.8) |
| UC11 | Rekap Nilai & Berita Acara | Admin, Proktor | Rekap nilai akhir per jadwal, predikat, remedial (nilai tertinggi), berita acara. | Alur K (baru; remedial sudah di §B.8) |
| UC12 | Analisis Butir Soal | Admin, Proktor, Guru | Analisis daya beda + pengecoh per soal, ekspor berkas. | Alur L (baru) |
| UC13 | Nilai per Kompetensi | Admin, Guru | Nilai per kompetensi (soal terhubung ke kompetensi) per siswa/jadwal. | Alur M (baru) |
| UC14 | Kelola UKK | Admin, Penguji UKK | CRUD UKK, nilai rubrik (Proses 30% / Produk 50% / Sikap 20%, teori 40% + praktik 60%, lulus ≥70), berkas praktik, sertifikat (hanya lulus, nomor UKK/tahun/kode-jurusan/urut). | Alur N (baru) |
| UC15 | Kelola Kegiatan, Ruangan & Kartu Peserta | Admin, Proktor, Siswa | CRUD Kegiatan & Ruangan (jadwal terhubung ke kegiatan dan ruangan), kartu peserta per jadwal; siswa melihat Kartu Saya. | Alur O (baru) |
| UC16 | Cadangkan & Pulihkan | Admin | Buat cadangan, pulihkan dengan pra-pemulihan otomatis, hapus cadangan (pra-pemulihan <7 hari dilindungi). | Alur P (baru) |
| UC17 | Kelola Pengaturan, Audit & Notifikasi | Admin | Pengaturan pasangan kunci-nilai, jejak audit hanya-tambah, notifikasi dalam aplikasi, pengecualian pemisahan tugas (alasan tertulis). | Alur Q (baru) |
| UC18 | Usul & Proses Buka Blokir | Admin, Proktor, Pengawas | Pengawas mengusulkan buka blokir (alasan); Proktor/Admin menyetujui/menolak; buka blokir = lanjut sisa waktu. | Alur C (§C.5) + catatan selisih di §10.1 |
| UC19 | Lihat Hasil & Pembahasan | Siswa | Siswa melihat hasil bila tampil hasil aktif (pengaturan paket DAN pengaturan sistem); pembahasan hanya setelah kirim / Tryout. | Alur R (baru; mekanisme di §B.10) |

Catatan inventaris:
- UC09 tidak punya asosiasi aktor di diagram Use Case `.mdj` (tidak ada garis aktor→UC09 yang ditemukan saat parse). Di kode, validasi token bukan endpoint/UI sendiri — melainkan langkah di dalam `POST /api/v1/ujian/mulai` (Alur I).
- UC18 di `.mdj` menyebut "menyetujui/menolak", tetapi kode tidak punya endpoint setuju/tolak usulan (lihat §10.1).

## 8.1 Daftar file kode yang dibaca (tambahan Fase 1b)

| # | File | Isi yang dipakai |
|---|---|---|
| F1 | `backend/cmd/server/main.go` | Daftar `Register(...)` semua paket (termasuk yang TIDAK terdaftar: `ukkx`, `analisisx` — lihat §10) |
| F2 | `backend/internal/master/master.go` | Seluruh CRUD Jurusan/Kelas/Mapel/Guru/Siswa/Kompetensi/Kisi-kisi + import siswa CSV |
| F3 | `backend/internal/banksoal/banksoal.go` | CRUD bank + soal + lock/unlock + import CSV legacy |
| F4 | `backend/internal/banksoal/import_excel.go` | Template .xlsx, preview, simpan .xlsx/.docx, parser .xlsx, duplikat |
| F5 | `backend/internal/banksoal/import_word.go` | Template .docx, parser .docx |
| F6 | `backend/internal/exam/exam.go` | Bagian paket, paket_soal, jadwal, token, koreksi (bagian ujian sudah di Fase 1) |
| F7 | `backend/internal/rekap/rekap.go` | Rekap, export CSV, berita acara, finalisasi, remedial (dibaca penuh) |
| F8 | `backend/internal/analisis/analisis.go` | Analisis butir + nilai per kompetensi (dibaca penuh) |
| F9 | `backend/internal/ukk/ukk.go` | CRUD UKK, nilai, berkas, sertifikat (dibaca penuh) |
| F10 | `backend/internal/event/event.go` | Event, ruangan, penugasan event/ruangan, kartu JSON legacy |
| F11 | `backend/internal/event/kartu_pdf.go` | Kartu PDF per siswa + gabungan (gofpdf) |
| F12 | `backend/internal/backup/backup.go` | Backup (pg_dump), restore (pg_restore), hapus, unduh (dibaca penuh) |
| F13 | `backend/internal/settings/settings.go` | Default + validasi + GET/PUT settings |
| F14 | `backend/internal/audit/audit.go` | `Catat`, `HasPermission`, GET audit |
| F15 | `backend/internal/notifikasi/notifikasi.go` | GET notifikasi + tandai dibaca + `Buat` internal |
| F16 | `backend/internal/sod/sod.go` | POST sod/override |
| F17 | `backend/internal/upload/upload.go` | Upload + sajikan gambar (pendukung, bukan UC) |
| F18 | `backend/internal/dashboard/dashboard.go` | GET dashboard/ringkasan (pendukung, bukan UC) |
| F19 | `frontend/src/App.tsx` | Daftar rute (dipakai memastikan ada/tidaknya halaman per UC) |
| F20 | `frontend/src/lib/api.ts` | `normalizePath`: path `/x` → `/api/v1/x`; FormData diteruskan tanpa Content-Type JSON |
| F21 | Halaman admin: `master/*.tsx`, `banksoal/*.tsx`, `paket/PaketPage.tsx`, `jadwal/JadwalPage.tsx`, `rekap/*.tsx`, `analisis/*.tsx`, `ukk/*.tsx`, `event/*.tsx`, `kartu/KartuPesertaPage.tsx`, `backup/BackupPage.tsx`, `settings/SettingsPage.tsx` | Panggilan API per halaman (dipakai memastikan alur frontend↔backend) |
| F22 | `frontend/src/pages/siswa/KartuSayaPage.tsx`, `HasilPage.tsx`; `components/common/NotifikasiBell.tsx` | Kartu siswa (link PDF), hasil, lonceng notifikasi |

---

## Alur D — UC02 Kelola Master Data

### D.0 Partisipan

| Partisipan | Peran |
|---|---|
| Admin (via permission `master.<entitas>.<aksi>`) | Membuka halaman master, isi form, hapus |
| Halaman `JurusanPage / KelasPage / MapelPage / GuruPage / SiswaPage / ImportSiswaPage` | Daftar + form modal + hapus + import siswa |
| Backend `master.go` | Seluruh endpoint `/api/v1/master/*` |
| Backend `audit.go` | `catat` → `audit_log` |
| PostgreSQL | `jurusan, kelas, mapel, guru, siswa, kompetensi, kisi_kisi, users, audit_log` |

### D.1 Pola umum semua entitas master

1. **Izin:** tiap handler memeriksa `audit.HasPermission(c, "master.<entitas>.<aksi>")` dulu — gagal → `403 forbidden`. (Pengecualian: Kompetensi + Kisi-kisi memakai izin `master.mapel.*`, bukan izin sendiri.)
2. **ID tidak valid / tidak ditemukan:** `GET/PUT/DELETE /:id` dengan ID bukan UUID → `404 not_found` (bukan 400). ID UUID tapi tidak ada → `404`.
3. **Daftar:** `?page=` (default 1) + `?per_page=` (default 20, maks 100) + `?q=` (ILIKE nama/kode). Filter spesifik per entitas (mis. `mapel_id` di Kompetensi) memakai helper `optUUID/optInt/optBool` — filter tiap entitas TIDAK ditelusuri satu per satu di fase ini.
4. **Buat/ubah:** body tidak valid → `400`; validasi bisnis → `422 validation_error`; pelanggaran unik → `409 conflict`; FK tidak ditemukan → `422` (pesan "…tidak ditemukan.").
5. **Hapus:** pola "guard hitung pemakaian dulu, baru DELETE": tiap guard yang > 0 → `409 conflict` dengan pesan berisi jumlah pemakaian. Berhasil → `200 {deleted: true, id}` + audit `delete_<entitas>`.
6. **Audit:** setiap create/update/delete/import mencatat `audit_log` (aksi `create_/update_/delete_/import_<entitas>`).

### D.2 Jurusan — `POST/GET/PUT/DELETE /api/v1/master/jurusan[/:id]`

- Validasi: `kode` wajib, `nama` wajib.
- Hapus ditolak `409` bila: dipakai > 0 `kelas`, atau > 0 `mapel` (pesan: "mapel produktif"), atau > 0 `siswa`. Urutan cek: kelas → mapel → siswa.

### D.3 Kelas — `POST/GET/PUT/DELETE /api/v1/master/kelas[/:id]`

- Validasi: `jurusan_id` wajib UUID; `tingkat` harus 10/11/12; `nama` wajib; `tahun_ajaran` wajib (contoh `2026/2027`).
- Buat/ubah dengan `jurusan_id` yang tidak ada di DB → `422` ("Jurusan tidak ditemukan.", via pelanggaran FK).
- Hapus ditolak `409` bila: dipakai > 0 `siswa`, atau > 0 `jadwal_ujian`.

### D.4 Mapel — `POST/GET/PUT/DELETE /api/v1/master/mapel[/:id]`

- Validasi: `kode` + `nama` wajib; `jenis` harus `umum` / `produktif`; `produktif` wajib `jurusan_id` UUID; `umum` DILARANG punya `jurusan_id` (harus kosong); `kelas_tingkat` bila diisi harus 10/11/12.
- Hapus ditolak `409` bila dipakai: `kompetensi`, `kisi_kisi`, `bank_soal`, `paket_ujian`, atau `jadwal_ujian` (via paket). Urutan cek sesuai daftar itu.

### D.5 Guru — `POST/GET/PUT/DELETE /api/v1/master/guru[/:id]`

- Validasi: `nama` wajib; `user_id` bila diisi harus UUID; `mapel_ids` array berisi UUID semua.
- `user_id` opsional: guru bisa dibuat TANPA akun login (akun dikaitkan belakangan lewat update). `user_id` yang tidak ada di `users` → `422` (pelanggaran FK).
- Hapus ditolak `409` bila `guru.user_id` adalah pemilik > 0 `bank_soal` — pesan: "Nonaktifkan akunnya, jangan hapus." (Alasan di kode: `bank_soal.guru_id` merujuk `users.id`, bukan `guru.id`.) Bila tidak punya bank soal → baris `guru` dihapus; baris `users` TIDAK ikut dihapus (tidak ada DELETE users di handler ini).

### D.6 Siswa — `POST/GET/PUT/DELETE /api/v1/master/siswa[/:id]`

- Validasi: `nisn` + `nama` wajib; `kelas_id` + `jurusan_id` wajib UUID; `user_id` bila diisi harus UUID; `tahun_masuk` bila diisi harus 1900–2100. `is_active` default `true` bila tidak dikirim.
- **Buat (transaksi):** bila `user_id` tidak dikirim → `createStudentAccountTx`: `INSERT INTO users (nama, username = NISN, password_hash = Argon2id(NISN), role = 'siswa', must_change_password = TRUE, is_active = TRUE) ON CONFLICT (username) DO NOTHING`; bila username sudah ada → pakai akun lama + paksa `must_change_password = TRUE` lagi. Lalu `INSERT INTO siswa (…)` dalam transaksi yang sama. (Parameter hash di `master.go`: Argon2id `m=65536, t=3, p=2`, salt 16 byte, hash 32 byte — format PHC yang sama dengan `auth.go`, yang memakai `t=1, p=4`; parameter terbaca dari format PHC saat verifikasi.)
- **Ubah:** `user_id` tidak dikirim → pertahankan yang lama; `user_id` string kosong → set NULL.
- **Hapus ditolak `409`** bila: punya > 0 `attempt` (dicek via `attempt.siswa_id = siswa.user_id`, karena `attempt.siswa_id` menyimpan `users.id`), atau > 0 `nilai_akhir`, atau > 0 `remedial` — pesan: "Nonaktifkan saja, jangan hapus." Bila lolos → baris `siswa` dihapus; **akun login (`users`) SENGAJA tidak dihapus** (komentar kode: riwayat + username tidak dipakai ulang untuk orang lain).

### D.7 Kompetensi + Kisi-kisi — `POST/GET/PUT/DELETE /api/v1/master/kompetensi[/:id]`, `/kisi-kisi[/:id]`

- Izin memakai `master.mapel.*` (bukan izin kompetensi/kisi-kisi sendiri).
- Kompetensi: `mapel_id` wajib UUID; `kode` wajib; `deskripsi` opsional. Hapus ditolak `409` bila dipakai > 0 `kisi_kisi` (via `kompetensi_id`) atau > 0 `soal`.
- Kisi-kisi: `mapel_id` wajib UUID; `kelas_tingkat` 10/11/12; `kelompok` wajib; `kompetensi_id` opsional UUID; `jumlah_soal` > 0. Hapus ditolak `409` bila dipakai > 0 `soal` (via `kisi_kisi_id`).
- **Tidak ada halaman frontend** untuk Kompetensi / Kisi-kisi (tidak ada rute `/adm/master/kompetensi` atau `/kisi-kisi` di `App.tsx`; tidak ada panggilan `api/v1/master/kompetensi|kisi-kisi` di frontend). Kedua entitas hanya bisa dikelola lewat API langsung.

### D.8 Impor Siswa (CSV) — bagian dari UC02

1. Frontend `ImportSiswaPage`: template CSV dibuat **di sisi client** (blob, nama `template-import-siswa.csv`, header `nama,nisn,nis,kelas_kode`); file dipilih (accept `.csv`) → `FormData {file}` → `apiClient.post("/api/v1/master/siswa/import", form)` (FormData diteruskan apa adanya, tanpa header Content-Type JSON). Respons ditampilkan sebagai ringkasan sukses/gagal/sudah_ada + daftar error per baris.
2. Backend `POST /api/v1/master/siswa/import` (izin `master.siswa.import`):
   - `ParseMultipartForm` maks 5 MB; field `file` wajib; baca CSV (maks 5 MB + 1), `FieldsPerRecord = -1`, `TrimLeadingSpace = true`.
   - Header dinormalisasi (huruf kecil, trim, buang BOM). Kolom: `nama` (alias `nama_lengkap`, `nama_siswa`), `nisn`, `nis` (opsional), kelas (alias `kelas_id`, `kelas_kode`, `kelas`, `kode_kelas`). Salah satu dari nama/nisn/kelas hilang → `400`.
   - > 2000 baris data → `400`.
3. Satu transaksi untuk seluruh file; per baris (baris = nomor baris file, header = baris 1):
   - `nama` / `nisn` / kelas kosong → `gagal` + error per baris; lanjut baris berikutnya.
   - NISN duplikat **di dalam file** → `gagal` + "NISN duplikat di dalam file."
   - NISN sudah ada di DB → `sudah_ada` (dilewati, TANPA error, TANPA membuat user baru) — impor idempotent; impor ulang file yang sama → `sukses 0, sudah_ada N, gagal 0`.
   - Resolusi kelas: bila nilai kelas adalah UUID → cari by `id`; bila bukan → cari by `LOWER(nama)` (ambil yang `created_at` terbaru). Tidak ketemu → `gagal` + "Kelas … tidak ditemukan." (jurusan diambil dari kelas yang ketemu — `jurusan_id` siswa = jurusan kelas).
   - Buat akun siswa via `createStudentAccountTx` (sama seperti D.6) + `INSERT INTO siswa (…, is_active = TRUE)`. Balapan import paralel yang menabrak unique (`siswa_nisn_key` / `users_username_key`) → dihitung `sudah_ada` (jaring pengaman).
4. Commit → audit `import_siswa` (sukses/gagal/sudah_ada) → `200 {sukses, gagal, sudah_ada, errors: [{baris, pesan}]}`.
5. **Perbedaan vs impor soal (Alur F):** impor siswa TIDAK punya langkah pratinjau (preview) dan TIDAK punya mode all_or_nothing — baris gagal dilewati satu per satu, baris lain tetap tersimpan dalam satu commit.

### D.9 Cabang/guard Alur D (ringkas)

| Kondisi | Efek |
|---|---|
| Tanpa izin `master.<entitas>.<aksi>` | `403` |
| ID bukan UUID / tidak ada | `404` |
| Validasi bisnis gagal | `422` (master) |
| Kode/NISN/username duplikat | `409` |
| Hapus entitas yang masih dipakai (guard per entitas) | `409` + jumlah pemakaian |
| Siswa baru tanpa `user_id` | Akun login dibuat otomatis (username = password awal = NISN, wajib ganti password) |
| Impor siswa: NISN sudah ada | `sudah_ada`, dilewati tanpa error |
| Impor siswa: > 2000 baris / > 5 MB / header tidak dikenali | `400` |

---

## Alur E — UC03 Kelola Bank Soal

### E.0 Partisipan

| Partisipan | Peran |
|---|---|
| Admin / Guru | Buat/ubah/hapus bank + soal; kunci/buka kunci bank (dengan batas SoD) |
| Proktor | Hanya metadata bank (lihat daftar bank) — DILARANG lihat isi soal |
| `BankSoalListPage / BankSoalFormPage / BankSoalDetailPage` + `SoalEditor` | Daftar bank, form bank, detail bank + editor soal |
| Backend `banksoal.go` | Endpoint `/api/v1/bank-soal/*` |
| PostgreSQL | `bank_soal, soal, paket_soal, audit_log` |

### E.1 Bank soal — CRUD

1. Daftar `GET /api/v1/bank-soal` (izin `banksoal.read`): filter `?q=` (kode/nama), `?mapel_id=`, `?kelas_tingkat=`, `?is_locked=`; tiap baris menyertakan `jumlah_soal` (COUNT soal). Proktor boleh memanggil endpoint ini (metadata saja).
2. Buat `POST /api/v1/bank-soal` (izin `banksoal.create`): `kode` + `nama` wajib; `mapel_id` wajib UUID; `kelas_tingkat` 10/11/12. Kepemilikan: bila role `guru` → `guru_id = user_id` sendiri (input diabaikan); bila bukan guru → `guru_id` dari input bila ada. Duplikat `kode` → `409 duplicate_kode`; mapel/guru tidak ada → `400`. Audit `banksoal.create` → `201`.
3. Ubah `PUT /api/v1/bank-soal/:id` (izin `banksoal.update`): cek `checkGuruScope` — guru hanya boleh mengubah bank miliknya sendiri (bukan miliknya → `403` "Bank soal ini milik guru lain."). `guru_id` hanya boleh diubah admin; guru tidak bisa memindahkan kepemilikan. Duplikat kode → `409`.
4. Hapus `DELETE /api/v1/bank-soal/:id` (izin `banksoal.delete`): ditolak `409 bank_locked` bila bank terkunci; ditolak `409 bank_in_use` bila ada soal bank ini yang tersusun di paket manapun (`paket_soal JOIN soal`). Lolos → DELETE + audit.

### E.2 Kunci / buka kunci bank (SoD)

1. `POST /api/v1/bank-soal/:id/lock` / `/unlock` (izin `banksoal.lock`).
2. **SoD-1:** bila role `guru` DAN `bank.guru_id == user_id` (pembuat bank sendiri) → `403 sod_violation` ("Pembuat bank tidak boleh mengunci/membuka bank buatannya sendiri. Minta guru lain atau admin.").
3. Bila bank sudah dalam keadaan yang diminta (`is_locked == locked`) → kembalikan bank apa adanya (`200`, tanpa UPDATE, tanpa audit).
4. Selain itu → `UPDATE bank_soal SET is_locked` + audit `banksoal.lock` / `banksoal.unlock`.

### E.3 Soal — CRUD di dalam bank

1. **Gerbang isi soal** (`forbidSoalUntukProktor`, dipanggil di SEMUA endpoint soal): role `proktor` → `403` "Proktor tidak boleh melihat isi soal." — proktor tidak bisa list/get/create/update/delete soal maupun import. Role tanpa izin `banksoal.*` (pengawas, siswa, penguji UKK) → `403` lewat `requirePerm`.
2. Daftar `GET /api/v1/bank-soal/:bankId/soal` + detail `GET …/:soalId` (izin `banksoal.read`, guru/admin saja): bank tidak ada → `404`; soal tidak ada → `404`.
3. Tambah `POST …/soal` / ubah `PUT …/soal/:soalId`:
   - Bank tidak ada → `404`; guru bukan pemilik bank → `403` (scope); **bank terkunci → `409 bank_locked`** ("…soal tidak bisa ditambah/diubah.").
   - Validasi `soalInput.validate()`: `tipe` harus dikenal (`ValidTipe`); `pertanyaan` wajib; `studi_kasus` wajib `stimulus`; `bobot` > 0; `level` bila diisi harus `mudah / sedang / sulit`; `status` bila diisi harus `draft / aktif / nonaktif / perlu_revisi / diarsipkan`; `teks_rumpang`: jumlah `___` di pertanyaan harus sama dengan jumlah jawaban kunci; lalu `ValidateSoal(tipe, jawaban, kunci)` per tipe. Gagal → `400 validation_error`.
   - Simpan: INSERT / UPDATE `soal (bank_id, tipe, kisi_kisi_id, kompetensi_id, stimulus, pertanyaan, media_url, jawaban, kunci, pembahasan, bobot, level, status)` + audit.
4. Hapus `DELETE …/soal/:soalId`: bank terkunci → `409 bank_locked`; soal tidak ada → `404`; lalu `DELETE FROM soal WHERE id AND bank_id`. **Tidak ada guard "soal dipakai paket" di handler hapus soal** (berbeda dengan hapus bank) — apakah FK/trigger DB menolaknya tidak ditelusuri (migrasi tidak dibaca).

### E.4 Impor CSV legacy (backend saja)

- `POST /api/v1/bank-soal/import` (izin `banksoal.import`, proktor dilarang): body JSON `{bank_id, csv_text|csv, mode}`; `mode` `all_or_nothing` (default) / `skip_errors`; bank terkunci → `409`; CSV kosong → `400`. **Endpoint ini TIDAK dipakai frontend** — panel impor memakai endpoint .xlsx/.docx (Alur F). Perilaku parse CSV-nya (`parseCSVImport`) tidak ditelusuri detail di fase ini.

### E.5 Cabang/guard Alur E (ringkas)

| Kondisi | Efek |
|---|---|
| Proktor memanggil endpoint isi soal | `403` (semua endpoint soal) |
| Guru mengubah bank milik guru lain | `403` |
| Bank terkunci + mutasi soal/import/hapus bank | `409 bank_locked` |
| Guru pembuat mengunci/membuka bank sendiri | `403 sod_violation` |
| Hapus bank yang dipakai paket | `409 bank_in_use` |
| Validasi soal gagal | `400 validation_error` |

---

## Alur F — UC04 Impor Soal (Excel / Word)

### F.0 Partisipan

| Partisipan | Peran |
|---|---|
| Admin / Guru | Unduh template, unggah file, periksa pratinjau, konfirmasi simpan |
| `ImportSoalPanel` (+ `PanduanImportPage`) | 3 langkah: template → unggah/pratinjau → konfirmasi |
| Backend `import_excel.go` + `import_word.go` | Template, preview, simpan |
| PostgreSQL | `bank_soal, soal, kompetensi, audit_log` |

### F.1 Langkah 1 — Unduh template

1. `GET /api/v1/bank-soal/import/template.xlsx` (juga alias `/template`) dan `GET …/template.docx` (izin `banksoal.import`, proktor dilarang) → file template sebagai attachment (`template-import-soal-cbt-smk-v2.xlsx/.docx`).
2. Template .xlsx (dibangun `buildTemplateXLSX`): 4 sheet — `Petunjuk`, `Soal` (satu-satunya yang dibaca parser; header = 16 kolom kanonik: `tipe, pertanyaan, stimulus, opsi_a, opsi_b, opsi_c, opsi_d, opsi_e, kunci, bobot, level, pembahasan, media_url, media_alt, kisi_kisi, kompetensi`), `Contoh` (16 baris), `Referensi Tipe` (16 tipe + format kuncinya). Guru mendarat di sheet `Soal` (sheet aktif = 1).
3. Template .docx (`buildTemplateDocx`): format blok per soal — `SOAL N`, label `TIPE:`, `STIMULUS:`, `PERTANYAAN:`, opsi `A.–E.`, `KUNCI:`, `BOBOT:`, `LEVEL:`, `PEMBAHASAN:`, `KISI-KISI:`, `KOMPETENSI:`, `MEDIA:`, `ALT MEDIA:`.

### F.2 Langkah 2 — Pratinjau (dry-run, TIDAK menyimpan)

1. Frontend: pilih format (xlsx/docx), pilih file (validasi client: ekstensi harus cocok dengan format yang dipilih), `fetch` POST multipart `{file, bank_id}` ke `POST /api/v1/bank-soal/import/preview` (fetch langsung + `credentials: include`, BUKAN `apiClient.post`). Panel menampilkan Statistik (total/valid/error/warning/duplikat/total bobot valid) + tabel pratinjau + daftar error/warning.
2. Backend `bacaFileImport`: multipart maks 5 MB (+1 MB ruang header); field `file` wajib; file kosong ditolak; ekstensi hanya `.xlsx` / `.docx` (selain itu `400`); baca isi maks 5 MB + 1 (klaim ukuran header tidak bisa menipu); **magic bytes harus arsip ZIP (`PK…`)**; isi arsip harus punya `xl/workbook.xml` (xlsx) / `word/document.xml` (docx).
3. Bank: `muatBankImport` — `bank_id` (form) wajib UUID; bank tidak ada → `404`; guru hanya bank miliknya (scope, seperti Alur E); **bank terkunci → `409 bank_locked`**.
4. `parseImportFile` → dispatch by format → respons preview: `{bank_id, total_baris, valid, error, warning, duplikat, ringkasan_tipe, total_bobot_valid, errors, warnings, duplikat_detail, preview[]}`. Tiap baris preview memuat kunci yang sudah terpetakan ke teks (`kunci_terpetakan`: label + teks opsi / pasangan `kiri::kanan` / jawaban per kolom) agar guru bisa verifikasi.
5. Validasi .xlsx (`parseXLSXImport`): sheet `Soal` wajib ada (hilang → error level file); header wajib memuat ke-16 kolom kanonik (kolom hilang → error per kolom, baris 1); header tak dikenal → warning `W3` (baris 0); **> 500 baris data → error level file**; baris kosong dilewati. Per baris (baris = nomor baris sheet, header = baris 1):
   - `E1` tipe kosong; `E2` tipe tak dikenal; `E3` pertanyaan kosong; `E4` kunci wajib diisi (kecuali tipe yang kuncinya boleh kosong → warning `W1` "rubrik belum diisi"); `E5` bobot bukan bilangan bulat > 0; `E6` level bukan mudah/sedang/sulit; `E28` `media_url` bukan URL http(s); `E29` opsi gambar bukan URL http(s); `E25` jumlah `___` ≠ jumlah jawaban kunci (teks_rumpang); `E26` stimulus wajib untuk studi_kasus; `ValidateSoal` sebagai jaring pengaman (error tanpa kode katalog bila gagal).
   - `E30` kisi_kisi: **kolom `kisi_kisi` yang diisi SELALU gagal E30** (`kisiKisiID` selalu `""` di kode ini); `kisi_kisi` kosong → NULL, lolos.
   - `E31` kompetensi: lookup kode (case-insensitive, trim) di `kompetensi WHERE mapel_id = mapel bank`; tidak ketemu → error; ketemu → `kompetensi_id` terisi.
   - `W2`: `media_url` diisi tapi `media_alt` kosong.
6. Validasi .docx (`parseDocxImport` + `gabungkanDocx`): pecah blok `SOAL N`; > 500 soal → error; temuan ber-`tingkat` error/warning/info; `baris` pada temuan = NOMOR SOAL (bukan nomor baris file); `E30` (kisi_kisi) + `E31` (kompetensi, lookup sama) diterapkan di `gabungkanDocx`; `W2/W7` media paritas .xlsx.
7. Duplikat (`tandaiDuplikat`, kedua format): pertanyaan dinormalisasi (`NormalisasiIsian`); duplikat di dalam file → warning `W4` (+ `duplikatBaris`); pertanyaan sudah ada di bank ini → warning `W5` (+ `duplikatSoalID`). Duplikat tetap berstatus "valid" (warning, bukan error).

### F.3 Langkah 3 — Konfirmasi simpan

1. Frontend: pilih `mode` (`all_or_nothing` / `skip_errors`) + `duplikat` (`lewati` / `simpan_ganda`), kirim ulang file yang SAMA + `bank_id + mode + duplikat` via fetch multipart ke `POST /api/v1/bank-soal/import/xlsx`. Tombol konfirmasi **dinonaktifkan** bila `mode == all_or_nothing` dan preview punya error > 0. **Nama endpoint `/xlsx` dipakai untuk file Excel MAUPUN Word.**
2. Backend: validasi `mode` + `duplikat` (`400` bila di luar daftar); `bacaFileImport` + bank + lock-check diulang (sama seperti preview); parse ulang file.
   - `mode == all_or_nothing` dan ada error → `200 {sukses: 0, gagal: len(errors), dilewati_duplikat: 0, disimpan: false, …}` — **tidak ada baris yang disimpan**.
   - Selain itu: tiap baris valid di-INSERT (`INSERT INTO soal (…, status = 'draft')`); baris duplikat dilewati bila `duplikat == lewati` (dihitung `dilewati_duplikat`); gagal INSERT per baris → ditambah ke `errors` dengan kode `E32`. Audit `banksoal.import` (sukses/gagal/dilewati/mode/duplikat/format) → `200 {sukses, gagal, dilewati_duplikat, warnings, errors, mode, duplikat, disimpan: sukses > 0}`.
3. **Soal hasil impor selalu berstatus `draft`** — wajib ditelaah sebelum dipakai (dipakai `generate-acak` hanya soal `status = 'aktif'`, Alur G).

### F.4 Yang TIDAK ada di impor soal

1. Tidak ada endpoint impor CSV di panel (CSV hanya di endpoint legacy Alur E.4 yang tidak dipakai frontend).
2. Tidak ada impor gambar inline .docx yang ditelusuri — jalur media adalah kolom `MEDIA:`/URL, bukan parsing gambar tertanam (tidak diklaim ada/tidak; tidak dibaca).
3. Tidak ada penguncian bank otomatis sesudah impor; tidak ada pengiriman soal impor ke paket otomatis.

---

## Alur G — UC05 Susun Paket Ujian

### G.0 Partisipan

| Partisipan | Peran |
|---|---|
| Admin / Guru (`requireRole(admin, guru)`) | Buat/ubah/hapus paket, susun soal, generate acak |
| `PaketPage` (+ komponen `PaketFormPage` / `PaketDetailPage` sebagai mode, BUKAN rute terpisah) | Daftar + form + detail susunan soal |
| Backend `exam.go` (bagian paket) | Endpoint `/api/v1/paket/*` |
| PostgreSQL | `paket_ujian, paket_soal, soal, bank_soal, mapel, jadwal_ujian, audit_log` |

### G.1 Paket — CRUD

1. Daftar `GET /api/v1/paket` (filter `?jenis=`, `?mapel_id=`): tiap paket menyertakan `jumlah_tersusun` (COUNT `paket_soal`) di samping `jumlah_soal` tersimpan.
2. Buat `POST /api/v1/paket` / ubah `PUT /api/v1/paket/:id` — validasi `bindPaket`: `kode`, `nama`, `mapel_id` wajib; `jenis` harus `STS / SAS / USBK / UKK / Tryout`; `durasi_menit` > 0; `jumlah_soal` ≥ 0. Default bila tidak dikirim: `acak_soal = true`, `acak_jawaban = true`, `tampil_hasil = false`.
   - Buat: `mapel_id` harus ada di `mapel` (tidak ada → `400`); duplikat `kode` → `409`.
   - Ubah: ID bukan UUID → `400`; paket tidak ada → `404`; **ganti mapel ditolak `422`** bila paket masih berisi soal dari mapel lain ("Kosongkan susunan soal dulu sebelum ganti mapel."); duplikat kode → `409`; audit `paket.update`.
3. Hapus `DELETE /api/v1/paket/:id`: ditolak `409` bila dipakai > 0 `jadwal_ujian` ("Hapus jadwal dulu."); lolos → DELETE (baris `paket_soal` ikut terhapus via FK/aturan DB — mekanisme cascade tidak ditelusuri) + audit `paket.delete`.

### G.2 Susun soal ke paket

1. Lihat susunan `GET /api/v1/paket/:id/soal`: JOIN `paket_soal + soal`, urut `urutan ASC`; per item `{soal_id, urutan, tipe, pertanyaan, bobot}`.
2. Tambah `POST /api/v1/paket/:id/soal {soal_id}`: paket tidak ada → `404`; soal tidak ada → `400`; **mapel soal (via bank soal) harus sama dengan mapel paket** → beda → `422`. `urutan = MAX(urutan) + 1`; INSERT `ON CONFLICT (paket_id, soal_id) DO NOTHING` (tambah soal yang sudah ada = idempotent, tanpa error — tetapi respons tetap mengembalikan `urutan = MAX+1`, yang belum tentu urutan baris lama yang sudah ada). Lalu `sinkronJumlahSoal`: `paket.jumlah_soal = COUNT(paket_soal)`. Audit `paket.soal.add` → `201`.
3. Hapus dari paket `DELETE /api/v1/paket/:id/soal` (`soal_id` via `?soal_id=` atau body): tidak ada di paket → `404`; lalu DELETE + `sinkronJumlahSoal` + audit.
4. **Tidak ada endpoint ubah urutan soal dalam paket** (tidak ada PUT/reorder `paket_soal` di `Register` exam.go). Urutan hanya ditentukan saat tambah/generate-acak.

### G.3 Generate acak — `POST /api/v1/paket/:id/generate-acak {jumlah}`

1. `jumlah` harus > 0 (`400` bila tidak); paket tidak ada → `404`.
2. Ambil soal: `soal JOIN bank_soal WHERE bank.mapel_id = mapel paket AND soal.status = 'aktif' AND soal.id NOT IN (SELECT soal_id FROM paket_soal WHERE paket_id)` — `ORDER BY random() LIMIT jumlah`. Artinya: soal `draft` (termasuk hasil impor, Alur F) TIDAK pernah terpilih; soal yang sudah ada di paket tidak dipilih ulang.
3. Sisipkan dengan `urutan` lanjutan (MAX+1, +2, …; `ON CONFLICT DO NOTHING`) + `sinkronJumlahSoal` + audit `paket.generate_acak` (diminta/ditambahkan).
4. Respons `{paket_id, diminta, ditambahkan}` + `peringatan` ("Soal tersedia di bank mapel ini kurang dari yang diminta.") bila `ditambahkan < diminta`.

### G.4 Catatan untuk diagram

- `paket.jumlah_soal` adalah kolom tersimpan yang disinkronkan ke COUNT susunan — bukan input yang dipercaya. Nilai awal dari form (`jumlah_soal` di `bindPaket`) akan tertimpa sinkronisasi begitu soal ditambah/dihapus/generate.
- `paket.durasi_menit` hanya ditampilkan; deadline ujian = jam selesai jadwal (Alur B).

---

## Alur H — UC06 Kelola Jadwal Ujian

### H.0 Partisipan

| Partisipan | Peran |
|---|---|
| Admin / Proktor / Guru | Buat/ubah/hapus jadwal (pengawas TIDAK bisa) |
| Pengawas | Hanya lihat daftar/detail jadwal (termasuk `token_ada` — lihat H.4) |
| `JadwalPage` (+ komponen `JadwalFormPage` sebagai mode, BUKAN rute terpisah) | Daftar + form + token + pause/resume |
| Backend `exam.go` (bagian jadwal) | Endpoint `/api/v1/jadwal/*` |
| PostgreSQL | `jadwal_ujian, paket_ujian, kelas, attempt, audit_log` |

### H.1 Jadwal — CRUD

1. Daftar `GET /api/v1/jadwal` (role admin/proktor/guru/pengawas): filter tidak ditelusuri detail di fase ini (handler `listJadwal` hanya dibaca sekilas untuk peran).
2. Buat `POST /api/v1/jadwal` (role admin/proktor/guru) — validasi `bindJadwal`: `paket_id`, `kelas_id`, `tanggal`, `jam_mulai`, `jam_selesai` wajib; `jam_selesai` harus sesudah `jam_mulai` (tanggal yang sama — tidak ada jadwal lintas hari); `status` default `draft`, harus `draft / aktif / pause / selesai`. `paket_id` / `kelas_id` tidak ada → `400`. Audit `jadwal.create` → `201`.
3. Ubah `PUT /api/v1/jadwal/:id` (role admin/proktor/guru): jadwal `selesai` → `409` ("Jadwal yang sudah selesai tidak boleh diubah."). **Update TIDAK memeriksa ulang keberadaan paket/kelas** (berbeda dengan create). Audit `jadwal.update`.
4. Hapus `DELETE /api/v1/jadwal/:id` (role admin/proktor/guru): ditolak `409` bila sudah ada > 0 `attempt` ("…hasil ujian tidak pernah dihapus fisik."). Lolos → DELETE + audit.
5. Detail `GET /api/v1/jadwal/:id` (4 role staf): mengembalikan `token_ada` (boolean) + `token_expires_at`, **TIDAK PERNAH `token_hash`**.

### H.2 Pause / resume jadwal — lihat Alur C §C.4

`POST /api/v1/jadwal/:id/pause` dan `/resume` (role admin/proktor). Seluruh mekanisme (bekukan `sisa_waktu_beku`, deadline baru = now + sisa beku, broadcast) sudah ditelusuri di §C.4 — tidak diulang di sini.

### H.3 Generate token — lihat Alur B §B.1/§B.2 + Alur I

`POST /api/v1/jadwal/:id/token` (role admin/proktor): 6 karakter alfabet `ABCDEFGHJKLMNPQRSTUVWXYZ23456789`; simpan hanya hash SHA-256 + `token_expires_at = now + token_expiry_detik` (setting, default 900); token plaintext dikembalikan SEKALI; token kedua menimpa token pertama; jadwal `selesai` → `409`; audit `jadwal.token.generate` TANPA token plaintext.

### H.4 Status token + visibilitas token untuk pengawas

1. `GET /api/v1/jadwal/:id/token-status` (role admin/proktor SAJA): `{jadwal_id, ada_token, token_expires_at, kedaluwarsa_dalam_detik, kedaluwarsa}` — tanpa nilai token.
2. **Selisih vs dokumentasi UC06** ("Pengawas DILARANG lihat token"): nilai token memang tidak pernah dikembalikan ke siapapun. Namun `GET /jadwal/:id` (yang boleh diakses pengawas) mengembalikan `token_ada` + `token_expires_at`. Jadi pengawas bisa melihat ADA/TIDAKNYA token dan KAPAN kedaluwarsanya, tetapi tidak nilainya. Ini dicatat apa adanya dari kode; keputusan apakah ini pelanggaran UC06 diserahkan ke fase bangun diagram.

---

## Alur I — UC09 Validasi Token

> UC09 TIDAK punya endpoint atau UI sendiri. Seluruh validasi terjadi sebagai langkah di dalam `POST /api/v1/ujian/mulai` (Alur B §B.2, langkah 7–13). Bila fase bangun diagram membuat diagram UC09 terpisah, isinya adalah fragmen langkah-langkah berikut; bila digabung ke UC08, UC09 = fragmen `alt/opt` di dalam diagram ujian.

### I.1 Langkah validasi (urutan persis di `mulaiUjian`)

| # | Pemeriksaan | Bila gagal |
|---|---|---|
| 1 | Attempt `berjalan` sudah ada | Dikembalikan apa adanya (`200`, `sudah_ada: true`) — **validasi token DILEWATI** (resume tanpa token) |
| 2 | Attempt `diblokir` / `selesai` sudah ada | `403` / `409` — token tidak diperiksa |
| 3 | `now < jam_mulai` / `now > jam_selesai` | `409` — token tidak diperiksa |
| 4 | **Toleransi telat:** `now > jam_mulai + toleransi_telat_menit` (setting, default 15) — hanya attempt BARU | `403` — token tidak diperiksa |
| 5 | **Token** (dilewati bila `paket_jenis == "Tryout"`): token kosong ATAU jadwal belum punya `token_hash` → `401 token_invalid`; `SHA256(token) ≠ token_hash` → `401 token_invalid`; `now > token_expires_at` → `409 token_expired` | sesuai kolom |

### I.2 Sisi frontend

- `UjianPage`: input token di-`uppercase` + `maxLength 6`; `SiswaHomePage`: input token TIDAK di-uppercase (perbedaan kecil, §B.1). Server membandingkan hash persis — token yang digenerate hanya berisi huruf/angka uppercase, jadi token huruf kecil dari SiswaHomePage akan gagal hash.
- Tidak ada tombol "kirim ulang token" / "lihat token lagi": token yang hilang hanya bisa diganti dengan generate token baru (yang menghanguskan token lama).

---

## Alur J — UC10 Koreksi Esai / Manual

### J.0 Partisipan

| Partisipan | Peran |
|---|---|
| Guru / Admin (`requireRole(admin, guru)`) | Melihat daftar koreksi, memberi skor |
| Backend `exam.go` (listKoreksi, koreksiJawaban) | Endpoint `/api/v1/koreksi/*` |
| PostgreSQL | `jawaban, attempt, soal, siswa, nilai_akhir` |

### J.1 Daftar koreksi — `GET /api/v1/koreksi/:jadwalId`

1. Jadwal tidak ada / ID bukan UUID → `404` / `400`.
2. Kueri: `jawaban JOIN attempt JOIN soal LEFT JOIN siswa` dengan filter `attempt.jadwal_id = :jadwalId AND jawaban.dikoreksi = FALSE AND jawaban.jawaban IS NOT NULL AND soal.tipe IN ('uraian','upload_file','praktik_singkat','esai','essay','isian_panjang','isi_panjang','esai_singkat','uraian_singkat','uraian_esai')`, urut nama siswa + `jawaban.updated_at`.
3. Per item: `{jawaban_id, attempt_id, soal_id, jawaban, ragu, poin (NULL), dikoreksi: false, updated_at, siswa_id (= attempt.siswa_id), attempt_status, nama_siswa, nisn, tipe, pertanyaan, bobot}`.
4. **Observasi join nama:** `LEFT JOIN siswa s ON s.id = a.siswa_id` — sedangkan di bagian lain kode (`syncNilaiAkhir`, §B.8; `deleteSiswa`, §D.6) `attempt.siswa_id` menyimpan `users.id`, BUKAN `siswa.id`. Bila itu benar, `nama_siswa`/`nisn` di daftar koreksi akan kosong (`COALESCE('', '')`). Ini TIDAK dikonfirmasi ke data nyata — dicatat sebagai ketidakpastian (lihat §10.4), bukan sebagai fakta.

### J.2 Koreksi satu jawaban — `POST /api/v1/koreksi/:jawabanId {skor}`

1. `skor` harus 0–100 (`400` bila di luar); jawaban tidak ada → `404`.
2. `poin = skor / 100 × bobot` soal. Transaksi: `UPDATE jawaban SET poin, dikoreksi = TRUE, updated_at = now()`.
3. Bila attempt `selesai` → hitung ulang nilai akhir (`hitungNilai`) → `UPDATE attempt SET nilai_akhir` → `syncNilaiAkhir` → respons menyertakan `nilai_akhir_baru` + `final = !menunggu_koreksi`. Bila attempt masih berjalan → nilai akhir dihitung saat submit; respons hanya `{jawaban_id, attempt_id, poin, menunggu_koreksi}`.
4. Audit `koreksi.update` (attempt_id, soal_id, skor, poin) + `broadcastAttempt` (monitor ikut ter-refresh).
5. **Tidak ada guard "sudah dikoreksi":** POST ulang ke `jawabanId` yang sama akan menimpa `poin`/`dikoreksi` tanpa penolakan — koreksi ulang dimungkinkan lewat endpoint yang sama (tidak ada status "terkunci" sesudah koreksi).
6. **Tidak ada rubrik di koreksi CBT:** skor adalah satu angka 0–100 bebas dari guru. Rubrik Proses/Produk/Sikap hanya ada di UKK (Alur N), bukan di sini. Paket `ukkx/rubrik.go` tidak terdaftar di `main.go` (lihat §10.5).
7. **Tidak ada halaman frontend koreksi:** tidak ada rute `/adm/koreksi*` di `App.tsx` dan tidak ada panggilan `/api/v1/koreksi` di frontend. Endpoint hanya bisa dipakai lewat API langsung.

### J.3 Hubungan dengan Alur B

§B.8 sudah mencatat: esai yang tersimpan saat autosave mendapat `poin = NULL, dikoreksi = false`; `menunggu_koreksi = true` bila ada soal manual belum dikoreksi; nilai esai disembunyikan sampai final bila tampil-hasil tidak efektif. Alur J adalah sisi staf dari mekanisme yang sama.

---

## Alur K — UC11 Rekap Nilai & Berita Acara

### K.0 Partisipan

| Partisipan | Peran |
|---|---|
| Admin / Proktor (izin `rekap.read / rekap.export / rekap.finalisasi / rekap.remedial`) | Lihat rekap, export, finalisasi, remedial |
| Siswa | Hanya melihat nilainya sendiri (bila final + tampil hasil) |
| `RekapPage / BeritaAcaraPage` | Dropdown jadwal + tabel + remedial + finalisasi + export + berita acara |
| Backend `rekap.go` | Endpoint `/api/v1/rekap*` |
| PostgreSQL | `nilai_akhir, siswa, kelas, attempt, jadwal_ujian, paket_ujian, mapel, jawaban, remedial, settings, audit_log` |

### K.1 Lihat rekap — `GET /api/v1/rekap?jadwal_id=`

1. `jadwal_id` kosong → `400`; belum login → `401`.
2. **Siswa:** bila setting `tampil_hasil_ke_siswa` false → `200 {items: [], statistik: statistik(nil)}` (kosong, bukan error). Bila true → baris miliknya sendiri SAJA (`queryNilai(…, siswaID, onlyFinal = true)`) — siswa hanya melihat nilai yang sudah `final`.
3. **Staf** (izin `rekap.read`): semua baris jadwal itu (`onlyFinal = false`), urut nama siswa ASC.
4. Isi per baris (`queryNilai`): `{id, attempt_id, jadwal_id, siswa_id, nilai, status, predikat, nisn, nama, kelas, jumlah_dijawab}` — `jumlah_dijawab = COUNT(jawaban WHERE jawaban IS NOT NULL)` (ragu-only tidak terhitung, sama seperti §B.5/D3).
5. Statistik: `jumlah_peserta, rata_rata` (dibulatkan 2 desimal), `tertinggi, terendah, jumlah_lulus` (nilai ≥ 70), `jumlah_menunggu_koreksi` (status `menunggu_koreksi`). Predikat: `< 70 Belum Tuntas; 70–79 Cukup; 80–90 Baik; 91–100 Sangat Baik` (fungsi `Predikat`, sama di UKK).

### K.2 Export CSV — `GET /api/v1/rekap/:jadwalId/export.csv`

1. Izin `rekap.export`; audit `rekap.export` (jadwal_id, jumlah_baris).
2. File `rekap-<jadwalId>-<tanggal>.csv`, BOM UTF-8 (agar Excel benar), header `nisn, nama, kelas, nilai, predikat, status`; nilai diformat 2 desimal. **Tidak ada filter final** — export memakai `queryNilai(…, onlyFinal = false)`, jadi nilai belum final ikut ter-export.

### K.3 Berita acara — `GET /api/v1/rekap/:jadwalId/berita-acara`

1. Izin `rekap.read`; jadwal tidak ada → `404` (dengan fallback kueri minimal bila JOIN mapel/kelas gagal).
2. Isi: `{jadwal_id, mapel, kelas, tanggal, jam_mulai, jam_selesai, jumlah_peserta, hadir, tidak_hadir, rata_rata, lulus, catatan: "", generated_at}`.
   - `jumlah_peserta = COUNT(attempt)` jadwal itu; `hadir = COUNT(attempt WHERE status IN ('selesai','diblokir'))`; `tidak_hadir = jumlah_peserta − hadir`.
   - **Konsekuensi:** siswa tanpa attempt sama sekali (tidak pernah mulai) TIDAK terhitung di jumlah_peserta maupun tidak_hadir; attempt `berjalan` terhitung "tidak hadir" meski sedang mengerjakan. `catatan` selalu string kosong — tidak ada input catatan di endpoint ini.
   - Tanggal/jam diambil sebagai TEXT (`::text`) agar frontend tidak crash render objek tanggal.
3. Frontend `BeritaAcaraPage`: mengambil data + menampilkan; tombol cetak/print tidak diverifikasi di fase ini.

### K.4 Finalisasi — `POST /api/v1/rekap/:jadwalId/finalisasi`

1. Izin `rekap.finalisasi`. Transaksi: `UPDATE nilai_akhir SET status = 'final', predikat = CASE (skala Predikat), updated_at = now() WHERE jadwal_id = …` — **tanpa filter status awal**: baris `menunggu_koreksi` ikut difinalisasi (tidak ada guard "semua esai harus sudah dikoreksi"). Audit `rekap.finalisasi` → `200 {jadwal_id, jumlah_difinalisasi, status: "final"}`.
2. Sesudah finalisasi, `syncNilaiAkhir` tidak akan menurunkan status `final` (Alur B §B.8).

### K.5 Remedial — `POST /api/v1/rekap/:jadwalId/remedial {siswa_id, nilai_baru}`

1. Izin `rekap.remedial`; `siswa_id` wajib; `nilai_baru` 0–100 (`400` bila di luar).
2. Transaksi: `INSERT INTO remedial (jadwal_id, siswa_id, nilai)` → baca `nilai_akhir` siswa di jadwal itu (tidak ada → `404`) → `nilai_akhir_baru = MAX(nilai_awal, nilai_baru)` → `UPDATE nilai_akhir SET nilai, predikat` (predikat dihitung ulang). Audit `rekap.remedial` (nilai_awal, nilai_remedial, nilai_akhir).
3. **Remedial adalah entri nilai manual, BUKAN ujian ulang** (§B.8): tidak ada attempt baru, soal, atau token. Setting `remedial_mode` hanya mengenal `"tertinggi"` (validasi settings menolak nilai lain) — konsisten dengan MAX di handler.
4. Frontend `RekapPage`: remedial lewat modal (siswa + nilai baru), finalisasi lewat tombol, export lewat `apiClient.download`.

---

## Alur L — UC12 Analisis Butir Soal

### L.0 Partisipan

| Partisipan | Peran |
|---|---|
| Admin / Proktor / Guru (izin `analisis.read / analisis.export`) | Melihat analisis + export |
| `AnalisisButirPage` (+ `DayaBedaBadge / DistribusiChart / PengecohTable`) | Dropdown jadwal + tabel + grafik + unduh CSV |
| Backend `analisis.go` | `GET /api/v1/analisis/:jadwalId[/export.csv]` |
| PostgreSQL | `jadwal_ujian, paket_soal, soal, attempt, jawaban, nilai_akhir, kompetensi` |

### L.1 Muat data — `loadDasar`

1. Jadwal tidak ada → `404` (error `errNotFound`).
2. Soal = soal paket jadwal itu (via `paket_soal`, lengkap dengan `kompetensi_id/kode/nama` + array `opsi`).
3. Attempt = attempt jadwal itu yang "selesai" menurut filter: `status = 'selesai' OR EXISTS (nilai_akhir final untuk attempt itu)` — attempt `diblokir`/`berjalan` TANPA nilai final tidak masuk analisis.
4. Jawaban = jawaban attempt-attempt tersebut. Jawaban untuk soal di luar paket diabaikan (`computeAnalisis`).

### L.2 Hitung analisis — `computeAnalisis`

1. **Pemeringkatan:** tiap attempt dinilai `nilaiAttempt` = `attempt.nilai_akhir` bila ada, else `Σ poin / Σ bobot paket × 100` (jawaban tanpa poin = 0). Urut menurun (seri → `attempt.id` ASC).
2. **Kelompok atas/bawah:** `n ≥ 10` → 27% atas/bawah (`metode = "27%"`); `2 ≤ n ≤ 9` → median split 50% + `sampel_kecil = true` + catatan "daya beda tidak dapat dipercaya sepenuhnya"; `n < 2` → daya beda tidak dapat dihitung (`sampel_kecil = true` + catatan).
3. **Per soal:**
   - Penjawab (`jumlah_penjab`): jawaban non-NULL; soal manual yang belum dikoreksi (`poin = NULL`) dihitung di `belum_dikoreksi` dan TIDAK masuk penjawab maupun benar. `jumlah_tidak_dijawab = n − dijawab` (dijawab = jawaban non-NULL, termasuk manual belum dikoreksi — jadi soal manual belum dikoreksi terhitung dijawab di penyebut tidak-dijawab, tapi tidak di penjawab).
   - Benar (`isBenar`): jawaban kosong/`"null"` → salah; `poin = NULL` → salah; tipe manual → benar bila `poin ≥ 0.6 × bobot`; tipe lain → benar bila `poin == bobot` (toleransi 1e-9). Artinya soal sebagian (mis. PG Kompleks 50%) terhitung SALAH di analisis, meski mendapat poin sebagian di nilai.
   - Kesukaran `p = jumlah_benar / jumlah_penjab` (bila penjawab > 0; else `kesukaran = null`, label "Belum ada jawaban"): `p < 0.3` Sukar; `p ≤ 0.7` Sedang; else Mudah. Label dihitung dari `p` mentah (bukan dari pembulatan).
   - Daya beda `d = (benarAtas − benarBawah) / ukuranKelompok` (bila ukuran kelompok > 0): `≥ 0.40` Sangat Baik; `≥ 0.30` Baik; `≥ 0.20` Cukup; `≥ 0` Buruk; `< 0` Sangat Buruk. Interpretasi dari `d` mentah; nilai tampil dibulatkan 2 desimal (`round2`).
   - Pengecoh: hanya tipe `pg` / `pg_kompleks` (`opsiBased`; `pengecoh_status = "ada"`, selain itu `"n/a"`). Per opsi: `pemilih` = jumlah attempt yang memilih label itu; `persen = pemilih / dijawab × 100`; `berfungsi = persen ≥ 5`.
   - `pertanyaan` diringkas maks 120 rune (`ringkas`) di output.
4. Respons: `{jadwal_id, jumlah_peserta: n, jumlah_soal, metode_kelompok, sampel_kecil, catatan, items[]}`.

### L.3 SoD proktor + export

1. **SoD:** `SembunyikanPertanyaanUntukProktor` diterapkan pada JSON maupun CSV: proktor melihat statistik TANPA teks pertanyaan (agar tidak mengelak lewat export).
2. Export `GET /api/v1/analisis/:jadwalId/export.csv` (izin `analisis.export`): audit `analisis.export`; nama file `analisis-butir-<jadwalId>-<tanggal>.csv`; `Content-Type: text/csv; charset=utf-8`.
3. **Tidak ada penyimpanan hasil analisis:** analisis dihitung ulang tiap request dari data mentah (tidak ada tabel hasil analisis, tidak ada endpoint simpan).

---

## Alur M — UC13 Nilai per Kompetensi

### M.0 Partisipan

| Partisipan | Peran |
|---|---|
| Admin / Guru (lihat M.3 soal izin) | Melihat nilai per kompetensi + export |
| `KompetensiPage` (+ `KompetensiHeatmap`) | Dropdown jadwal + heatmap + unduh CSV |
| Backend `analisis.go` (computeKompetensi) | `GET /api/v1/kompetensi/:jadwalId[/export.csv]` |
| PostgreSQL | Sama seperti Alur L + `siswa, kelas` (via `loadSiswa`) |

### M.1 Kelompok soal per kompetensi — `computeKompetensi`

1. Soal dikelompokkan by `kompetensi_id`; soal tanpa kompetensi → grup `"Tanpa Kompetensi"` (kode `"-"`, diurutkan TERAKHIR; grup lain urut kode ASC).
2. Data dasar sama seperti Alur L (`loadDasar`: attempt selesai/final + jawaban), ditambah `loadSiswa` (nama/NISN/kelas per attempt).

### M.2 Nilai satu siswa satu kompetensi — `nilaiKompetensi`

`nilai = Σ poin / Σ bobot × 100` dengan aturan penyebut:
- Soal tidak dijawab → poin 0, bobot TETAP masuk penyebut.
- Soal manual sudah dijawab tapi belum dikoreksi → dikeluarkan dari pembilang DAN penyebut ("belum dinilai", bukan 0).
- Soal ternilai → poin + bobot masuk.
- Hasil: `dijawabN == 0` → `nilai = null`, flag `"tidak ada soal"`, `tidakAdaSoal = true` (perhatikan: flag ini berarti "tidak ada soal yang dijawab", bukan "grup tanpa soal"); semua yang dijawab belum dikoreksi → `nilai = null`, flag `"belum dikoreksi"`; sebagian belum dikoreksi → nilai terhitung + flag `"sebagian belum dikoreksi"` + `belum_dikoreksi` > 0; selain itu flag `""`.
- Nilai dibulatkan 2 desimal (`round2`).
- `rata_rata` siswa = `nilaiKompetensi` atas SEMUA soal (aturan sama). Agregat kelas per kompetensi = rata-rata nilai non-null siswa + `jumlah_siswa` (yang punya nilai). `rata_rata_kelas` = rata-rata `rata_rata` siswa non-null.
- Siswa diurutkan nama ASC (seri → `siswa_id`).

### M.3 Izin + export (observasi)

1. `GET /api/v1/kompetensi/:jadwalId` memakai permission **`rekap.read`** (teramati di kode — bukan `kompetensi.*` atau `analisis.read`). Export `GET …/export.csv` memakai **`rekap.export`** (bukan `analisis.export`). Penamaan izin ini tidak konsisten dengan fitur; dicatat apa adanya.
2. Export: audit `rekap.export` (entitas `kompetensi`); nama file `nilai-kompetensi-<jadwalId>-<tanggal>.csv`.
3. Sama seperti Alur L: dihitung ulang tiap request, tidak disimpan.

---

## Alur N — UC14 Kelola UKK

### N.0 Partisipan

| Partisipan | Peran |
|---|---|
| Admin / Penguji UKK (izin `ukk.*`) | Buat/ubah/hapus UKK, nilai, berkas, terbitkan sertifikat |
| Siswa | Upload berkas praktik miliknya sendiri; lihat berkas/sertifikat miliknya |
| `UkkListPage / UkkDetailPage / UkkNilaiPage / UkkSertifikatPage` | Daftar, detail + berkas + sertifikat, nilai |
| Backend `ukk.go` | Endpoint `/api/v1/ukk/*` |
| PostgreSQL | `ukk_ujian, ukk_nilai, ukk_berkas, ukk_sertifikat, siswa, mapel, kelas, jurusan, audit_log` + disk `./storage/ukk/` |

### N.1 UKK — CRUD

1. Daftar `GET /api/v1/ukk` (izin `ukk.read`): filter `jurusan_id, kelas_id, status, q` (nama); tiap baris menyertakan `jumlah_peserta` (COUNT `ukk_nilai`).
2. Buat `POST /api/v1/ukk` (izin `ukk.create`): validasi `nama` wajib; `mapel_id / kelas_id / jurusan_id` wajib UUID; `status` bila diisi harus `draft / berjalan / selesai` (default `draft`); `tanggal` bila diisi harus `YYYY-MM-DD`; `penguji_internal / penguji_eksternal` opsional. Mapel/kelas/jurusan dicek keberadaannya (`existsID`) → tidak ada → `400`. Audit `ukk.buat` → `201`.
3. Ubah `PUT /api/v1/ukk/:id` (izin `ukk.update` — teramati dari rute; isi validasi sama dengan buat; `status` tidak dikirim → pertahankan lama) + audit (nama, status_lama, status_baru). Tidak ada guard status khusus di update yang ditelusuri.
4. Hapus `DELETE /api/v1/ukk/:id`: ditolak `409` bila masih punya > 0 `ukk_nilai` / `ukk_berkas` / `ukk_sertifikat`. Lolos → DELETE + audit. (Berkas di disk TIDAK disebutkan dihapus di handler yang dibaca — lihat §10.6.)

### N.2 Nilai UKK — rubrik

1. Daftar `GET /api/v1/ukk/:id/nilai` (izin `ukk.read`): UKK tidak ada → `404`; baris JOIN `siswa`, urut nama.
2. Simpan `POST /api/v1/ukk/:id/nilai` (upsert per `(ukk_id, siswa_id)`) / `PUT /api/v1/ukk/:id/nilai/:nilaiId` (izin `ukk.update`): tiap komponen (`nilai_teori, nilai_proses, nilai_produk, nilai_sikap`) harus 0–100 atau NULL (`validNilai`: NULL = belum diisi, sah).
3. Hitung server (bukan client):
   - `nilai_praktik = proses × 0.3 + produk × 0.5 + sikap × 0.2` — NULL bila ada satu komponen rubrik yang belum diisi.
   - `nilai_akhir = teori × 0.4 + praktik × 0.6` — NULL bila teori/praktik belum lengkap.
   - `lulus = nilai_akhir ≥ 70` (NULL bila nilai_akhir NULL); `predikat` skala sama seperti rekap (Belum Tuntas / Cukup / Baik / Sangat Baik).
   - Semua nilai dibulatkan 2 desimal (`round2`).
4. **Guard UKK selesai:** bila `ukk.status == "selesai"` dan role bukan `admin` → `403` ("…nilai tidak dapat diubah kecuali oleh admin."). Admin BOLEH mengubah + audit mencantumkan `override_ukk_selesai: true`. Ini override terpisah dari `sod_overrides` (Alur Q) — tidak menulis baris `sod_overrides`.
5. PUT per `nilaiId` mengambil `siswa_id` dari baris yang ada (tidak dari body); POST tanpa `siswa_id` valid → `400`; siswa tidak ada → `400`. Respons: `201` bila baris baru (POST), `200` bila update.
6. Frontend `UkkNilaiPage`: form per siswa → POST/PUT di atas.

### N.3 Berkas praktik

1. Daftar `GET /api/v1/ukk/:id/berkas` (login wajib): siswa hanya melihat berkas MILIKNYA (filter `siswa_id` dari profilnya; profil tidak ketemu → `403`); staf wajib izin `ukk.read`.
2. Upload `POST /api/v1/ukk/:id/berkas` (login wajib; multipart): siswa → `siswa_id` dipaksa ke dirinya sendiri (mengirim `siswa_id` orang lain → `403`); staf → wajib izin `ukk.create` + `siswa_id` valid + siswa ada. Validasi: `ParseMultipartForm` maks 10 MB; field `file` wajib; ukuran file > 10 MB → `400`; ekstensi hanya `pdf, doc, docx, jpg, png, zip` → selain itu `400`.
3. Simpan: nama file disanitasi (`sanitizeFileName`: `filepath.Base`, karakter tidak aman → `_`, maks 120 karakter, kosong → `"berkas"`); nama disk `<nanotime>_<namaAman>` (unik + tidak dapat ditebak); direktori `./storage/ukk/<ukk_id>/<siswa_id>/` (mode `0750`); `INSERT INTO ukk_berkas (ukk_id, siswa_id, nama_file = nama ASLI, path, ukuran, diupload_oleh)`; bila INSERT gagal → file di disk dihapus lagi. Audit `ukk.berkas.upload` → `201`.
4. Unduh `GET /api/v1/ukk/:id/berkas/:berkasId/download`: siswa hanya miliknya (bukan miliknya → `404` disamarkan); staf wajib `ukk.read`; path file harus di dalam `./storage/ukk/` (cek prefix, anti path traversal); file tidak ada di disk → `404`; audit `ukk.berkas.download`; dikirim sebagai attachment dengan nama asli.
5. Frontend `UkkDetailPage`: upload via `fetch` + FormData langsung (BUKAN apiClient — karena apiClient hanya JSON) + CSRF; unduh via `apiClient.download`.

### N.4 Sertifikat

1. Lihat `GET /api/v1/ukk/:id/sertifikat/:siswaId`: siswa hanya miliknya (bukan → `404` disamarkan); staf wajib `ukk.read`. Data dari JOIN `ukk_sertifikat + siswa + ukk_ujian + jurusan + ukk_nilai` (nilai teori/praktik/akhir + predikat).
2. Terbitkan `POST /api/v1/ukk/:id/sertifikat/:siswaId` (izin `ukk.update`):
   - **Idempotent:** sertifikat sudah ada → dikembalikan apa adanya (`200`, tanpa INSERT baru).
   - **Hanya lulus:** `ukk_nilai.lulus` tidak ada / NULL / false → `409` ("Sertifikat hanya dapat diterbitkan bagi siswa yang lulus UKK (nilai akhir >= 70).").
   - Nomor: `UKK/<tahun>/<kode-jurusan>/<urut 4 digit>`; tahun dari `tanggal` UKK bila ada, else tahun berjalan; urut = COUNT sertifikat jurusan itu di tahun `diterbitkan_at` + 1; bentrok UNIQUE → coba lagi maks 5× (nomor +1 tiap percobaan).
   - Audit `ukk.sertifikat.terbit` → `201` (sertifikat baru).
3. **Sertifikat adalah data JSON, bukan PDF:** tidak ada endpoint PDF sertifikat di `ukk.go`; frontend `UkkSertifikatPage` hanya GET + menampilkan data. (Beda dengan kartu peserta yang punya PDF, Alur O.)

### N.5 Yang TIDAK ada di UKK

- Paket `ukkx` (`rubrik.go, sertifikat.go, upload.go`) **tidak didaftarkan** di `main.go` — kode pendukung/cross-check yang tidak melayani request. Semua perilaku UKK di atas berasal dari `ukk.go`.
- Tidak ada CRUD rubrik terpisah: rubrik adalah rumus tetap 30/50/20 + 40/60 di kode, bukan data yang bisa diubah lewat API.

---

## Alur O — UC15 Kelola Kegiatan (Event), Ruangan & Kartu Peserta

### O.0 Partisipan

| Partisipan | Peran |
|---|---|
| Admin / Proktor (izin `event.*`) | CRUD event/ruangan, tugaskan event/ruangan ke jadwal, unduh kartu |
| Guru | Lihat kartu (via `requireKartuStaf`) |
| Siswa | Lihat/unduh Kartu Saya |
| `EventPage / RuanganPage / KartuPesertaPage / KartuSayaPage` | Daftar event/ruangan, kartu + penugasan, kartu siswa |
| Backend `event.go` + `kartu_pdf.go` | Endpoint `/api/v1/event*`, `/ruangan*`, `/jadwal/:id/event-ruangan`, `/kartu/*` |
| PostgreSQL | `event, ruangan, jadwal_ujian, siswa, kelas, jurusan, settings, audit_log` |

### O.1 Event — CRUD

- Validasi: `kode` + `nama` wajib; `jenis` default `Lainnya`, harus `STS / SAS / USBK / UKK / Tryout / Lainnya`; `tanggal_mulai / tanggal_selesai` bila diisi harus `YYYY-MM-DD`; `tanggal_selesai` tidak boleh sebelum `tanggal_mulai`.
- Duplikat `kode` (di luar baris sendiri saat update) → `409`.
- Hapus: ditolak `409 event_in_use` bila dipakai > 0 `jadwal_ujian` (`event_id`).
- Izin: `event.read / create / update / delete` untuk semua operasi event.

### O.2 Ruangan — CRUD

- Validasi: `kode` + `nama` wajib; `kapasitas` > 0.
- Duplikat `kode` → `409`. Hapus: ditolak `409 ruangan_in_use` bila dipakai > 0 `jadwal_ujian` (`ruangan_id`).
- **Izin memakai `event.*` juga** (tidak ada izin `ruangan.*` sendiri): daftar/buat/ubah/hapus ruangan memakai `event.read/create/update/delete`.

### O.3 Tugaskan event + ruangan ke jadwal — `PUT /api/v1/jadwal/:id/event-ruangan`

1. Izin `event.update`; jadwal tidak ada → `404`; body JSON `{event_id, ruangan_id}`.
2. Field tidak dikirim / string kosong = **lepas penugasan (NULL)** — endpoint ini juga dipakai untuk melepas.
3. `event_id` diisi → harus UUID + event ada (tidak ada → `400`). `ruangan_id` diisi → harus UUID + ruangan ada (tidak ada → `400`).
4. **Cek kapasitas:** `ruangan.kapasitas ≥ COUNT(siswa WHERE kelas_id = kelas jadwal AND is_active = TRUE)` — kurang → `409 kapasitas_kurang` (pesan menyebut kapasitas, nama kelas, jumlah siswa).
5. `UPDATE jadwal_ujian SET event_id, ruangan_id` + audit `event.assign` → respons berisi event/ruangan yang ditugaskan (atau null).

### O.4 Kartu peserta — JSON legacy

1. `GET /api/v1/kartu/:jadwalId` (staf admin/proktor/guru via `requireKartuStaf`): jadwal tidak ada → `404`; respons `{jadwal: {…}, items: […], total}`.
2. `GET /api/v1/kartu/:jadwalId/:siswaId`: siswa → hanya miliknya sendiri (cocok `siswa.user_id == uid` ATAU `siswa.id == uid`, dan kelas siswa == kelas jadwal; selain itu `404` disamarkan). Staf → siswa harus sekelas dengan jadwal (`404` bila tidak).
3. `buildKartuList`: siswa AKTIF kelas jadwal, urut nama A–Z. **Nomor peserta dihitung per request, TIDAK disimpan:** `nomor = <kodeEvent atau "US"> / <tahun tanggal jadwal> / <urut 3 digit>` (urut = posisi alfabetis, mulai 001). Konsekuensi: perubahan nama/event/tahun mengubah nomor peserta yang tampil.
4. Isi kartu: nama sekolah (dari settings `nama_sekolah`), nama event, nama siswa, NISN, kelas, jurusan, mapel, paket, tanggal, jam, ruangan (+ lokasi), `username = NISN`, `catatan = "Password awal: NISN (wajib diganti saat login pertama)"`. **Tidak ada token di kartu** (JSON maupun PDF).

### O.5 Kartu peserta — PDF backend (gofpdf)

1. `GET /api/v1/kartu/siswa/:siswaId/pdf`: siswa hanya miliknya (bukan → `404` disamarkan; `siswaId` dicari dulu by `siswa.id`, fallback `siswa.user_id` — karena `KartuSayaPage` mengirim `user.id` akun login); staf admin/proktor/guru boleh semua. Siswa TIDAK AKTIF → `404` (tidak punya kartu).
2. `GET /api/v1/kartu/pdf?kelas_id=&event_id=` (staf admin/proktor/guru saja): `kelas_id`/`event_id` bila diisi harus UUID + harus ada di DB (`404` bila tidak). Kumpulkan kartu semua siswa aktif (urut nama A–Z; filter event = hanya siswa yang kelasnya punya jadwal di event itu). **> 500 siswa → `400 too_many_siswa`** ("gunakan filter kelas_id / event_id"); 0 kartu → `404`. Respons PDF gabungan (`kartu-peserta-gabungan.pdf`), 1 halaman per siswa.
3. Isi satu kartu PDF (`loadKartuPDFData`): identitas SEKALI (nama, NISN, kelas, jurusan, username) + **tabel SEMUA jadwal kelas itu** (urut tanggal + jam mulai; tanpa filter status — jadwal `draft` ikut tampil; bila `event_id` filter diisi, hanya jadwal event itu). Per baris jadwal: nomor peserta (via `buildKartuList`, sama seperti O.4), mapel, paket, tanggal (diformat), jam, ruangan (nama saja), nama event. Ditambah kop (nama sekolah), kotak login, tanda tangan. PDF per siswa 1 kartu (bukan 1 kartu per jadwal).
4. Frontend: `KartuSayaPage` memakai `<a href>` langsung ke PDF milik sendiri (`/api/v1/kartu/siswa/<user.id>/pdf`, cookie session ikut); `KartuPesertaPage` memakai link per siswa + link gabungan (dengan filter kelas/event), pratinjau JSON (`GET /kartu/:jadwalId`), dan form penugasan (O.3). `window.print()` tidak dipakai untuk kartu (0 kemunculan — PDF di-backend adalah satu-satunya jalur cetak).
5. **Rute `/kartu/:jadwalId` di `App.tsx`** ada, tetapi perilaku redirect-nya tidak ditelusuri di fase ini (tidak dibaca sampai render).

---

## Alur P — UC16 Cadangkan & Pulihkan (Backup / Restore)

### P.0 Partisipan

| Partisipan | Peran |
|---|---|
| Admin / Proktor (izin `backup.create / backup.restore`) | Buat, unduh, pulihkan, hapus backup |
| `BackupPage` | Daftar + buat (keterangan) + unduh + konfirmasi restore/hapus |
| Backend `backup.go` | Endpoint `/api/v1/backup*` |
| Sistem | Proses anak `pg_dump` / `pg_restore`; file di direktori backup server; tabel `backup_meta` |

### P.1 Buat backup — `POST /api/v1/backup`

1. Izin `backup.create`; body `{keterangan}` opsional, maks 500 karakter (`400` bila lebih); body kosong tetap sah.
2. Mutex `dumpMu` (dump/restore/hapus diserikan — tidak boleh berjalan bersamaan). `doBackup` (jenis `manual`): `pg_dump -Fc` (format custom) via proses anak, timeout 10 menit; bila `pg_dump` tidak tersedia → `500` ("Pasang paket PostgreSQL client di server."). Nama file unik (sufiks acak).
3. Metadata `backup_meta` (id, nama_file, ukuran_bytes, jenis, dibuat_oleh, keterangan, created_at) + audit `backup.create` → `200 {id, nama_file, ukuran_bytes, jenis, created_at}`.
4. **Tidak ada backup otomatis/terjadwal di kode Go** (komentar kode: backup otomatis lewat skrip cron `deploy/backup.sh` di luar aplikasi — skrip itu tidak ditelusuri; tidak ada scheduler/ticker backup di backend).

### P.2 Daftar + unduh

1. `GET /api/v1/backup` (izin `backup.create` ATAU `backup.restore`): `?limit=` default 50 maks 200 + `?offset=`; urut `created_at` DESC; respons `{data: items, meta: {limit, offset, total}}` (bentuk meta berbeda dari pola `writeList` master).
2. `GET /api/v1/backup/:id/download` (izin sama): ID bukan UUID → `404`; meta tidak ada → `404`; **meta ada tapi file hilang di disk → `404`** ("File backup tidak ditemukan di server (metadata ada, file hilang)."); file disajikan via streaming (bukan static file server) + audit `backup.download`.

### P.3 Restore — `POST /api/v1/backup/:id/restore`

1. Izin `backup.restore`; ID bukan UUID / meta tidak ada / file hilang → `404`; `pg_restore` tidak tersedia → `500`.
2. **Gerbang ujian (`examInProgress`):** ditolak `409` + audit `backup.restore.ditolak` bila ada > 0 jadwal `aktif`/`pause` ATAU > 0 attempt `berjalan`.
3. **Backup pengaman otomatis:** `doBackup` jenis `pre-restore` ("Backup pengaman otomatis sebelum restore dari <file>") DULU; bila `pg_dump` tidak tersedia / pengaman gagal → restore DIBATALKAN (tidak lanjut ke `pg_restore`). Audit `backup.create` untuk pengaman.
4. Jalankan `pg_restore --clean --if-exists --no-owner --no-privileges -d <db> <file target>`, timeout 10 menit. Gagal → audit `backup.restore.gagal` + `500` (pesan menyebut nama file pengaman agar admin bisa pulih manual; timeout → pesan "Data saat ini mungkin sebagian terpulihkan.").
5. Sukses → **kembalikan jejak:** karena `--clean` menghapus baris meta + audit yang ditulis sebelum restore, handler meng-`upsert` kembali meta pengaman + meta target, lalu audit `backup.create` (pengaman) + `backup.restore` (target + `pre_restore_backup`). Respons `{dipulihkan_dari, pre_restore_backup: {id, nama_file, ukuran_bytes, jenis, created_at}}`.
6. Frontend: `Modal.confirm` dengan peringatan keras ("Restore menimpa SEMUA data saat ini — backup pengaman dibuat otomatis…") sebelum POST.

### P.4 Hapus — `DELETE /api/v1/backup/:id`

1. Izin `backup.restore` (bukan `backup.create`).
2. **Pre-restore < 7 hari DITOLAK:** `m.jenis == "pre-restore" AND umur < 7×24 jam` → `409` + audit `backup.delete.ditolak` ("pengaman restore terakhir harus tetap tersedia").
3. Selain itu: hapus file (file sudah hilang = tetap lanjut, `IsNotExist` ditoleransi) + `DELETE FROM backup_meta` + audit `backup.delete` → `200 {deleted: true, id, nama_file}`.
4. **Tidak ada endpoint upload file backup:** restore hanya dari file backup yang sudah ada di server (meta + file). File `.dump` dari luar tidak bisa diunggah lewat API.

---

## Alur Q — UC17 Kelola Pengaturan, Audit & Notifikasi

### Q.0 Partisipan

| Partisipan | Peran |
|---|---|
| Admin (izin `settings.read/update`, `audit.read`, `sod.override`) | Ubah pengaturan, lihat audit, catat override SoD |
| Semua pengguna login | Lihat notifikasi miliknya, tandai dibaca |
| `SettingsPage`, `NotifikasiBell` | Form pengaturan; lonceng notifikasi |
| Backend `settings.go / audit.go / notifikasi.go / sod.go` | Endpoint `/api/v1/settings`, `/audit`, `/notifikasi/*`, `/sod/override` |
| PostgreSQL | `settings, audit_log, notifikasi, sod_overrides` |

### Q.1 Pengaturan — `GET / PUT /api/v1/settings`

1. GET (izin `settings.read`): DB rows digabung di atas `DefaultSettings()`:
   `{tampil_hasil_ke_siswa: false, token_expiry_detik: 900, toleransi_telat_menit: 15, max_pelanggaran: 2, durasi_default: 90, nama_sekolah: "", logo_url: "", kktp_per_mapel: {}, remedial_mode: "tertinggi"}`.
2. PUT (izin `settings.update`): body `{"settings": {…}}` ATAU flat `{key: value}`; body kosong → `400`.
   - Key tak dikenal → `400` ("Key tidak dikenal: …") — tidak ada sampah konfigurasi.
   - Validasi per key: `tampil_hasil_ke_siswa` boolean; 4 key angka harus bilangan bulat ≥ 0 + batas: token 60–86400, toleransi 0–120, max_pelanggaran 1–10, durasi 5–600; `nama_sekolah/logo_url` string; `kktp_per_mapel` objek (KKTP per mapel/TP — BUKAN hardcode KKM=75; tiap mapel/TP punya ambang + deskripsi sendiri); `remedial_mode` hanya `"tertinggi"`.
   - Simpan: transaksi upsert per key (`ON CONFLICT (key) DO UPDATE`); audit `settings.update` (map `updated`) → `200 {updated}`.
3. Frontend `SettingsPage`: semua 9 key dalam satu form (angka via InputNumber min/max sesuai validasi backend; `kktp_per_mapel` via textarea JSON yang divalidasi `JSON.parse` di client; `remedial_mode` ditampilkan tetap `"tertinggi"`); simpan via `PUT {settings: payload}`.
4. **Konsumen `durasi_default` tidak ditelusuri** — tidak ditemukan pemakaian setting ini di handler yang dibaca (pembuatan jadwal memakai jam mulai/selesai eksplisit; deadline ujian memakai jam selesai jadwal). Dicatat sebagai ketidakpastian (§10.7), bukan sebagai "tidak dipakai".

### Q.2 Audit log — `GET /api/v1/audit`

1. Izin `audit.read`; filter `?aksi= &entitas= &actor_id=`; `?limit=` default 50 maks 200 + `?offset=`; urut `created_at` DESC. Kolom: `id, actor_id, aksi, entitas, entitas_id, detail, ip, created_at`.
2. **Audit hanya-tambah (append-only) dari sisi API:** tidak ada endpoint POST/PUT/DELETE audit. Baris audit hanya ditulis lewat `audit.Catat` dari handler lain (semua alur D–P mencatat aksinya masing-masing).
3. **Tidak ada halaman frontend audit:** tidak ada rute `/adm/audit` di `App.tsx`; `SettingsPage` hanya menyebut "Perubahan tercatat di audit log." sebagai teks.

### Q.3 Notifikasi — `GET /api/v1/notifikasi`, `POST /api/v1/notifikasi/:id/baca`

1. GET: milik sendiri SAJA (`WHERE penerima_user_id = user login`); `?limit=` default 50 maks 200 + `?offset=`; urut `created_at` DESC. Kolom: `id, penerima_user_id, judul, isi, dibaca, created_at`.
2. Tandai dibaca: `UPDATE notifikasi SET dibaca = TRUE WHERE id AND penerima_user_id = user` — bukan miliknya → `404` (disamarkan); audit `notifikasi.baca` → `200 {dibaca: true, id}`.
3. **Satu-satunya produsen notifikasi di kode adalah SoD override** (`sod.go` → `notifikasi.Buat` ke semua admin aktif kecuali actor). Tidak ada notifikasi untuk peristiwa lain (ujian mulai/selesai, pelanggaran, unblock, backup, dst.) — tidak ada pemanggilan `notifikasi.Buat` lain di backend.
4. **Tidak ada endpoint** hapus notifikasi / tandai semua dibaca / kirim notifikasi manual. Frontend `NotifikasiBell`: GET daftar + POST baca per item.
5. Catatan: tidak ada pengiriman notifikasi real-time (WS/email/push) — notifikasi hanya dibaca lewat polling/GET saat lonceng dibuka (perilaku polling lonceng tidak ditelusuri detail).

### Q.4 Pengecualian SoD (override) — `POST /api/v1/sod/override`

1. Login wajib + izin `sod.override`; body `{aksi, entitas, entitas_id, alasan}`; `aksi`/`entitas` wajib; `alasan` wajib minimal `PanjangAlasanMinimum = 20` karakter (rune, sesudah trim) → kurang → `400`.
2. `INSERT INTO sod_overrides (actor_id, aksi, entitas, entitas_id, alasan)` → `201 {id, actor_id, aksi, entitas, entitas_id, alasan}`.
3. Audit `sod.override` (detail berisi aksi + alasan + entitas + flag `SOD_OVERRIDE`); gagal catat audit TIDAK menggagalkan override yang sudah tersimpan (hanya log server).
4. Notifikasi "Pengecualian SoD" ke SEMUA user `role = 'admin' AND is_active = TRUE` kecuali actor.
5. **Override ini HANYA pencatatan, bukan bypass:** tidak ada handler yang memeriksa `sod_overrides` untuk mengizinkan suatu aksi. Contoh: SoD-1 lock bank (Alur E.2) tetap `403` meski ada override tercatat; override UKK-selesai (Alur N.2) adalah logika `role == admin` terpisah, bukan via `sod_overrides`.
6. **Tidak ada halaman frontend SoD override** (tidak ada rute/panggilan `/api/v1/sod` di frontend) — hanya API.

### Q.5 Pendukung (bukan UC): upload gambar + dashboard

- `POST /api/v1/upload/image` (login + salah satu izin `settings.update / banksoal.create / banksoal.update / ukk.create`): multipart maks 5 MB (+1 MB header); file kosong ditolak; ukuran riil > 5 MB ditolak; **tipe ditentukan dari magic bytes** (PNG/JPEG/WebP/GIF saja — selain itu `400`); nama disk UUID acak + ekstensi turunan magic bytes. `GET /api/v1/upload/image/:nama` (login): nama tidak aman / file tidak ada / magic bytes tidak cocok → `404`; `Cache-Control: private, max-age=3600`. Dipakai `SettingsPage` (logo) + editor soal (media) — alur detail pemakaian per halaman tidak ditelusuri.
- `GET /api/v1/dashboard/ringkasan` (Dipakai `DashboardPage`): ringkasan angka — bukan bagian UC manapun; isi respons tidak ditelusuri di fase ini.

---

## Alur R — UC19 Lihat Hasil & Pembahasan

> Mekanisme inti sudah ditelusuri di §B.10 (submit + halaman hasil). Alur R menambahkan: (a) jalur rekap siswa sebagai cara kedua melihat nilai, (b) batasan yang berlaku, (c) apa yang TIDAK bisa dilihat siswa.

### R.1 Jalur 1 — Halaman hasil (`HasilPage`)

1. `GET /api/v1/ujian/attempt/:id/hasil` (milik sendiri; attempt belum `selesai` → `409`).
2. `tampil = (paket.tampil_hasil AND settings.tampil_hasil_ke_siswa) OR paket_jenis == "Tryout"` — **Tryout selalu menampilkan hasil lengkap** (kunci + pembahasan), tanpa perlu kedua saklar.
3. Per soal (urut `urutan_soal` attempt): `jawaban_siswa, poin, bobot, ragu, status` (`benar` bila poin ≥ bobot; `sebagian` bila 0 < poin < bobot; `salah` bila poin = 0; `menunggu_koreksi` bila poin NULL; `disembunyikan` bila tidak tampil). Bila tampil: `kunci_tampil` per tipe + `pembahasan` (`soal.pembahasan`). Soal tanpa baris jawaban tetap ditampilkan (`salah` / `disembunyikan`).
4. `nilai_akhir` hanya diisi bila tampil dan tersimpan; selain itu `null` + frontend menampilkan "Nilai disembunyikan sampai koreksi final."

### R.2 Jalur 2 — Rekap siswa (`GET /api/v1/rekap?jadwal_id=`)

Siswa hanya melihat nilainya sendiri dan HANYA bila nilai sudah `final` DAN `tampil_hasil_ke_siswa` true (Alur K.1). Bila belum final atau saklar mati → daftar kosong. Jalur ini TIDAK menampilkan per-soal / kunci / pembahasan — hanya nilai akhir + predikat + status.

### R.3 Yang TIDAK bisa dilihat siswa

1. Kunci jawaban selama ujian (tidak pernah dikirim ke frontend — §B.3).
2. Kunci/pembahasan soal yang `tampil = false` (non-Tryout, salah satu saklar mati) — meski attempt sudah selesai dan sudah dikoreksi.
3. Nilai per soal / nilai per kompetensi milik siswa lain (endpoint kompetensi tidak punya cabang siswa; `GET /kompetensi` memakai izin staf `rekap.read`).
4. Kartu peserta orang lain (404 disamarkan — Alur O.4/O.5).

---

## 9. Usulan klasifikasi UTAMA vs LAMPIRAN per UC

Aturan user: **UTAMA** = alur inti untuk badan skripsi (login, ujian, penilaian/koreksi, impor, rekap/pantau); **LAMPIRAN** = sisanya. Usulan ini dipakai fase bangun diagram; keputusan final di sana.

| UC | Usulan | Alasan singkat |
|---|---|---|
| UC01 Masuk & Wajib Ganti Kata Sandi | **UTAMA** | Login — gerbang semua alur; sudah Alur A. |
| UC02 Kelola Master Data | LAMPIRAN | Prasyarat administratif (data referensi), bukan alur ujian itu sendiri. |
| UC03 Kelola Bank Soal | LAMPIRAN | Persiapan konten soal; alur ujian bisa dipahami tanpa detail CRUD bank. (Catatan: lock/SoD bank adalah aturan keamanan, bukan alur ujian.) |
| UC04 Impor Soal | **UTAMA** | Disebut eksplisit sebagai impor; jalur utama pengisian bank soal (template → preview → simpan). |
| UC05 Susun Paket Ujian | LAMPIRAN | Persiapan paket sebelum ujian; bukan pelaksanaan ujian. (Generate-acak + susun soal adalah setup.) |
| UC06 Kelola Jadwal Ujian | LAMPIRAN | CRUD jadwal adalah setup. Pengecualian: fragmen token + pause/resume SUDAH masuk diagram UTAMA (B/C) lewat rujukan — diagram UC06 lampiran cukup memuat CRUD + rujuk ke B/C untuk token/pause. |
| UC07 Pantau Ujian Langsung | **UTAMA** | Pantau — disebut eksplisit; sudah Alur C. |
| UC08 Ikuti Ujian | **UTAMA** | Ujian — inti skripsi; sudah Alur B. |
| UC09 Validasi Token | **UTAMA** | Sub-alur masuk ujian yang tak terpisahkan dari UC08; diagramnya fragmen di dalam / pendamping diagram UC08 (tidak berdiri sebagai diagram besar sendiri). |
| UC10 Koreksi Esai / Manual | **UTAMA** | Penilaian/koreksi — disebut eksplisit; penentu nilai akhir + `menunggu_koreksi`. |
| UC11 Rekap Nilai & Berita Acara | **UTAMA** | Rekap — disebut eksplisit; penutup siklus ujian (finalisasi + remedial). |
| UC12 Analisis Butir Soal | LAMPIRAN | Analitik pasca-ujian (daya beda/pengecoh); bukan penentu kelangsungan ujian. |
| UC13 Nilai per Kompetensi | LAMPIRAN | Analitik pasca-ujian per kompetensi; turunan dari data yang sama dengan UC12. |
| UC14 Kelola UKK | LAMPIRAN | Jalur ujian terpisah (UKK + sertifikat) di luar ujian CBT reguler; besar tapi bukan inti CBT. |
| UC15 Kegiatan, Ruangan & Kartu Peserta | LAMPIRAN | Dukungan administratif + kartu; alur ujian bisa dipahami tanpa detail kartu PDF. |
| UC16 Cadangkan & Pulihkan | LAMPIRAN | Operasional/DevOps (backup/restore), bukan alur fungsional ujian. |
| UC17 Pengaturan, Audit & Notifikasi | LAMPIRAN | Konfigurasi + jejak audit + notifikasi; tidak ada alur pengguna inti (audit/SoD bahkan tanpa UI). |
| UC18 Usul & Proses Buka Blokir | **UTAMA** | Bagian dari loop pantau + anti-curang (Alur C §C.5); tanpa ini, blokir tidak bisa pulih. Digabung ke diagram UC07/C, bukan diagram sendiri. |
| UC19 Lihat Hasil & Pembahasan | **UTAMA** | Keluaran ujian bagi siswa; penutup alur UC08 (aturan tampil-hasil adalah satu kesatuan dengan submit). Digabung/didampingi diagram UC08. |

Ringkasan usulan: **UTAMA = UC01, UC04, UC07, UC08, UC09, UC10, UC11, UC18, UC19 (9 UC)**; **LAMPIRAN = UC02, UC03, UC05, UC06, UC12, UC13, UC14, UC15, UC16, UC17 (10 UC)**.

---

## 10. Hal yang TIDAK ditemukan / tidak pasti (tambahan Fase 1b)

### 10.1 Selisih UC18: tidak ada setuju/tolak usulan unblock

Dokumentasi UC18 di `.mdj` menyebut "Proktor/Admin menyetujui/menolak". Di kode: **tidak ada endpoint setuju/tolak usulan** (`POST …/usulan-unblock/:id/setuju|tolak` tidak ada di `exam.go`). Satu-satunya aksi adalah `POST …/unblock` (proktor/admin), yang menandai usulan `menunggu` terbaru menjadi `disetujui` sebagai efek samping — tidak ada penolakan, tidak ada tombol setuju/tolak di `MonitorPage` (hanya teks usulan menunggu). "Menolak usulan" tidak punya wujud di kode.

### 10.2 Selisih UC06: pengawas bisa melihat `token_ada` + `token_expires_at`

Lihat Alur H.4. Nilai token tidak pernah terekspos; tetapi metadata token terekspos ke pengawas lewat `GET /jadwal/:id`.

### 10.3 Tidak ada UI frontend untuk: Kompetensi/Kisi-kisi (D.7), Koreksi (J.2), Audit (Q.2), SoD override (Q.4)

Keempatnya hanya bisa dipakai lewat API langsung. Diagram Sequence untuk UC-UC tersebut akan berisi aktor staf → backend → DB saja (tanpa halaman), atau perlu diputuskan di fase bangun diagram apakah tetap digambar dengan partisipan frontend.

### 10.4 Join nama siswa di daftar koreksi tampak tidak cocok (belum dikonfirmasi)

`listKoreksi` memakai `LEFT JOIN siswa s ON s.id = a.siswa_id`, sementara bagian lain kode menyatakan `attempt.siswa_id` menyimpan `users.id`. Bila benar, `nama_siswa`/`nisn` kosong di daftar koreksi. Tidak dikonfirmasi ke data nyata — jangan digambar sebagai fakta sebelum dicek.

### 10.5 Paket `ukkx` dan `analisisx` tidak terdaftar di `main.go`

`main.go` hanya me-`Register` paket: auth, audit, master, banksoal (+import Excel/Word), exam, event, monitor, rekap, analisis, settings, backup, ukk, upload, dashboard, sod, notifikasi. `ukkx/*` dan `analisisx/statistik.go` tidak dipanggil — fungsinya (statistik murni / pendukung UKK) tidak melayani request. `computeAnalisis`/`computeKompetensi` di `analisis.go` TIDAK memanggil `analisisx` (rumus ditulis langsung di `analisis.go`).

### 10.6 Hapus UKK: berkas di disk tidak disebutkan dihapus

Handler `deleteUkk` yang dibaca menghitung `ukk_nilai / ukk_berkas / ukk_sertifikat` dan menolak hapus bila ada isinya — sehingga dalam praktik berkas tidak pernah yatim lewat jalur ini. Apakah ada pembersihan file terpisah (cron/manual) tidak ditelusuri.

### 10.7 Konsumen `durasi_default` tidak ditemukan di handler yang dibaca

Setting ada + bisa diubah + tervalidasi, tetapi tidak terlihat dipakai untuk menghitung apa pun di handler yang dibaca (jadwal memakai jam eksplisit; deadline memakai jam selesai). Jangan klaim "tidak dipakai" — mungkin dipakai di tempat yang tidak dibaca (seed/dokumen).

### 10.8 Tidak ditelusuri detail di Fase 1b (bukan berarti tidak ada)

1. Filter spesifik tiap endpoint daftar (selain yang disebut eksplisit) — mis. filter `listJadwal`, `listKelas/Mapel/Guru/Siswa` per kolom, dan filter `listUkk` di luar yang disebut.
2. Isi `frontend` per halaman di luar panggilan API (validasi form client, teks, konfirmasi modal) — kecuali yang disebut eksplisit (mis. konfirmasi restore BackupPage, disable tombol konfirmasi impor).
3. Perilaku `DELETE` berantai di DB (cascade FK) — migrasi SQL tidak dibaca; yang dicatat hanya guard di handler + DELETE itu sendiri.
4. Isi skrip `deploy/backup.sh` (backup otomatis di luar aplikasi) — tidak dibaca.
5. Perilaku redirect rute `/kartu/:jadwalId` di `App.tsx` — tidak dibaca sampai render.
6. Detail polling lonceng notifikasi (`NotifikasiBell`) — hanya endpoint GET/baca yang dicatat.
7. Detail `DashboardPage` / `GET /dashboard/ringkasan` — bukan UC; tidak ditelusuri.
8. Isi `docs/KONTRAK.md` / `ANALISIS.md` tidak dipakai sebagai sumber — yang ditulis di sini adalah **kode** (+ dokumentasi `.mdj` hanya untuk inventaris nama/deskripsi UC dan aktor).

