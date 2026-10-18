# Keamanan — CBT SMK v2

Ringkasan keamanan untuk admin / proktor / tim sekolah. Dokumen ini
jujur tentang apa yang dilindungi dan apa yang **tidak bisa** dicegah
oleh aplikasi web. Milik Tim Fase 3 (Dokumen Final). Kontrak teknis
di `docs/KONTRAK.md`; deploy di
`docs/PANDUAN-DEPLOY-SEKOLAH.md`; riwayat versi di `docs/CHANGELOG.md`.

---

## 1. Autentikasi & sesi

- **Session cookie server-side** (bukan JWT). Sesi dibuat di server,
  disimpan di server (tabel `sessions` di PostgreSQL), dan bisa
  dicabut seketika (peserta yang baru diblokir berhenti pada request
  berikutnya — tidak menunggu token kedaluwarsa).
- Cookie sesi `cbt_session` — nilai = ID session (UUID):
  `HttpOnly` (tidak bisa dibaca JavaScript → kebal pencurian lewat
  XSS), `SameSite=Lax` (lapis kedua anti-CSRF), `Path=/`,
  `Secure` bila `APP_ENV=prod`, `MaxAge` 12 jam.
- **Sesi bertahan melewati restart server** (sesi ada di DB —
  restart TIDAK membuat logout massal). TTL session: 12 jam.
  Idle timeout: 2 jam tanpa request → session dianggap kedaluwarsa
  (dihapus saat diakses lagi). Berlaku sama untuk semua role.
- **Kunci tidak ke browser:** ID sesi, token CSRF rahasia, dan kunci
  apa pun tidak pernah dikirim ke JavaScript / `localStorage` /
  query string / log. Token ujian tidak muncul di URL.
- Logout menghapus sesi di server **dan** cookie di browser — bukan
  hanya menghapus di sisi klien.
- `must_change_password=true` → login ditolak 403
  `PASSWORD_CHANGE_REQUIRED` (tanpa session). Ganti password
  menghapus SEMUA session user (wajib login ulang).

## 2. CSRF

- Cookie `cbt_csrf` — nilai = `csrf_token` session. **Bukan**
  `HttpOnly` (agar JS bisa membacanya), `SameSite=Lax`, `Path=/`,
  `Secure` bila prod.
- Untuk `POST`/`PUT`/`PATCH`/`DELETE`, client WAJIB mengirim header
  `X-CSRF-Token` == `csrf_token` session (dibaca frontend dari
  cookie `cbt_csrf`). Perbandingan constant-time di server.
  Tanpa token yang valid → ditolak 403 `csrf_token_invalid`.
- Dilewati untuk: method aman (`GET`/`HEAD`/`OPTIONS`),
  `POST /api/v1/auth/login`, `GET /api/v1/auth/csrf`.
  `POST /api/v1/auth/ganti-password` mode kredensial (tanpa session)
  tidak memakai CSRF — ia mengandalkan password lama yang benar.
- `SameSite` pada cookie adalah pertahanan lapis kedua, bukan
  pengganti token CSRF.

## 3. Password

- Hash **Argon2id** (bukan bcrypt/MD5/SHA-1), format PHC:
  `$argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>`.
  Parameter: `time=1`, `memory=65536` KiB (64 MiB), `threads=4`,
  `keyLen=32`, salt acak 16 byte. Password asli tidak pernah
  disimpan, tidak pernah dicatat di log.
- Password baru minimal 8 karakter (login, ganti-password,
  create-admin, reset-password — semua jalur memakai batas yang sama).
- **Tidak ada akun demo / password bawaan.** Seed tidak membuat user
  apa pun. Akun pertama via CLI `create-admin` (hanya bila tabel
  `users` kosong), wajib ganti password saat login pertama.
- Akun per orang — akun bersama dilarang (log audit jadi tidak bermakna).
- **Lockout login:** 5x login gagal per akun per 15 menit → akun
  terkunci 15 menit; percobaan berikutnya 429 `login_locked`.
  Pesan kredensial salah generik (`invalid_credentials`, 401) —
  tidak membedakan "akun tidak ada" vs "password salah".
- Kegagalan login dicatat di audit log.

## 4. Otorisasi (RBAC)

- 6 role, **62 permission** per aksi (lihat `docs/KONTRAK.md` §5
  dan `backend/db/seed/seed.sql`). Penegakan di **server**
  (middleware + lapisan service + kepemilikan/scope), bukan hanya
  menyembunyikan tombol di frontend. Prinsip: deny-by-default;
  satu sesi = satu role aktif.
- Pengawas tidak melihat token & isi soal; BOLEH mengusulkan
  unblock (`jadwal.usulan-unblock`) — eksekusi unblock tetap
  proktor/admin. Proktor tidak melihat isi soal (di analisis butir,
  field `pertanyaan` disembunyikan untuk proktor).
  Guru hanya mapel + kelas yang ditugaskan (penegakan di server).
  Siswa hanya jadwal kelasnya + jendela waktu + token valid; tidak
  boleh melihat jadwal/nilai siswa lain walau menebak URL/ID
  (anti-IDOR: kembalikan 404, bukan 403).
- Kunci jawaban tidak pernah dikirim ke browser peserta. Penilaian
  dilakukan server — frontend TIDAK PERNAH menghitung nilai.
- **Pemisahan tugas (SoD)** ditegakkan sistem (bukan hanya
  disembunyikan): pembuat bank ≠ pengunci bank; pembuat soal ≠
  pemegang token; pengoreksi esai ≠ pengubah nilai final;
  penguji UKK ≠ finalisator nilai UKK; pengawas mengajukan,
  proktor mengeksekusi; pengelola user ≠ pengelola audit log;
  pengaju draft jadwal ≠ finalisator jadwal.

## 5. Integritas ujian

- **Timer server-side.** Waktu dihitung server; mengubah jam / waktu
  di perangkat siswa tidak menambah waktu. Auto-submit dari server
  saat waktu habis.
- **Pause membekukan timer.** Jadwal yang dijeda menyimpan sisa waktu
  saat pause dan melanjutkan dari nilai itu saat resume — bukan
  sekadar gerbang masuk. Attempt diblokir juga membekukan sisa waktu
  (`sisa_waktu_beku`) sehingga unblock lanjut dari sisa yang sama.
- **Auto-block per-attempt:** pelanggaran ke-1 = peringatan, ke-2 =
  auto-block (default `max_pelanggaran=2`). Blokir menempel pada
  attempt ujian itu, **bukan** mengubah status akun (siswa tetap bisa
  ikut ujian lain). Unblock = lanjut sisa waktu; hanya Proktor/Admin
  yang mengeksekusi.
- Pindah tab / keluar fullscreen dideduplikasi (event apa pun dalam
  5 detik setelah pelanggaran terakhir diabaikan — satu peristiwa =
  satu pelanggaran, bukan dihitung ganda karena `blur` +
  `visibilitychange` berapi bersamaan).
- Soal/jawaban diacak per peserta (urutan soal & opsi). Token ujian
  per-attempt, opaque, server-side; token kedua membuat token pertama
  hangus; kedaluwarsa default 900 detik.
- Autosave jawaban dengan antrean + retry; jawaban yang belum terkirim
  diselamatkan sebelum pengalihan apa pun (mis. sesi berakhir).
- Fullscreen enforcement: deteksi keluar fullscreen dicatat sebagai
  pelanggaran (bila diaktifkan sekolah). Fullscreen di browser tidak
  bisa dipaksakan 100% — lihat §8.

## 6. Data & audit

- **Audit log append-only:** login/logout, masuk ujian, submit,
  pelanggaran, block/unblock, token (generate/rilis/revoke), perubahan
  nilai, export, percobaan akses ditolak. Format: waktu
  (Asia/Jakarta + UTC) + aktor (ID + nama + role) + aksi + target +
  hasil + IP/perangkat + alasan (wajib untuk aksi sensitif) +
  sebelum/sesudah. **Tidak ada fitur ubah/hapus log dari aplikasi.**
- **Pengecualian darurat SoD (Fase 3)** diizinkan HANYA lewat
  `POST /api/v1/sod/override` (permission `sod.override`, admin
  saja): `alasan` wajib min **20 karakter** (kurang → 400). Efek:
  catat `sod_overrides` + `audit_log` dengan flag `SOD_OVERRIDE`
  (detail JSON berisi aksi + alasan + entitas) + buat `notifikasi`
  "Pengecualian SoD" untuk SEMUA user role admin kecuali actor.
  Pengecualian tidak pernah dipakai untuk ubah nilai tanpa jejak /
  hapus audit log.
- Password / token plaintext / kunci jawaban **tidak pernah** masuk log.
- Data ujian yang sudah dipakai adalah arsip: tidak dihapus fisik
  (soft delete / arsip). Penghapusan fisik hanya untuk data yang belum
  pernah dipakai dalam ujian.
- Retensi audit log: minimal 1 tahun ajaran (idealnya 2), lalu arsip.
- Upload: ekstensi whitelist, sniffing magic bytes (bukan percaya
  header client), nama file disimpan acak (UUID) + ekstensi tervalidasi,
  batas ukuran di 3 lapis (proxy, handler, validasi isi), file di luar
  webroot, dilayani via handler terautentikasi — tidak pernah dieksekusi.
- Isi soal/jawaban disanitasi dari HTML/JS berbahaya (XSS lewat isi
  soal adalah risiko nyata karena ditampilkan ke ratusan siswa).
- Header keamanan: CSP, `X-Content-Type-Options: nosniff`,
  `Referrer-Policy`, `frame-ancestors`, HSTS bila HTTPS.

## 7. Operasional

- **Hapus via POST, bukan GET.** Tautan (`<a href>`) tidak boleh
  menghapus / mengubah data — crawler / pratinjau tautan tidak boleh
  memicu aksi destruktif.
- **Rate limit** (per-IP, fixed window; penolakan 429 berkode
  `rate_limited` + header `Retry-After` dalam detik):
  - Endpoint umum: **100 req/menit/IP**.
  - Endpoint ujian (`/api/v1/ujian/attempt/*`:
    jawaban/ragu/pelanggaran/submit/status): **300 req/menit/IP**.
  - Login: 5x gagal per akun per 15 menit → kunci 15 menit,
    429 `login_locked` (terpisah dari rate limit per-IP).
  - Frontend me-retry 429 otomatis maks 3x (GET/HEAD + POST saja)
    dengan backoff 500 ms → 1 dtk → 2 dtk, mengikuti header
    `Retry-After` bila ada (dibatasi maks 10 detik). Retry aman:
    429 ditolak middleware SEBELUM handler, jadi request yang
    di-retry belum pernah dieksekusi.
- Timeout HTTP eksplisit: `ReadTimeout` 60 dtk, `WriteTimeout`
  120 dtk, `IdleTimeout` 120 dtk (Fase 3, BUG-02).
- Validasi berlapis: transport → struct → domain/bisnis → constraint DB.
  Decoder menolak field tak dikenal & body terlalu besar.
- Konfigurasi dari environment variables; startup **fail-fast** bila
  secret kosong / konfigurasi tidak valid. Secret tidak pernah di-log.
  `.env` hanya untuk dev, tidak pernah di-commit.
- Backup terenkripsi / disimpan aman; restore mengikuti prosedur di
  `docs/PANDUAN-DEPLOY-SEKOLAH.md` (pre-restore otomatis + tolak
  saat ujian berjalan → restore → uji kesehatan → buka layanan).

## 8. Yang TIDAK bisa dicegah aplikasi web — jujur

Ini bukan kekurangan implementasi, melainkan batas teknologi web.
Sekolah perlu mengimbanginya dengan pengawasan ruang:

| Risiko | Kenapa web tidak bisa mencegah | Mitigasi sekolah |
|--------|-------------------------------|------------------|
| **HP kedua** | Browser tidak bisa melihat perangkat lain di tangan siswa | Pengawas keliling, HP dikumpulkan / ditaruh di depan |
| **Foto layar** | Tidak ada cara andal melarang kamera fisik memotret layar | Pengawas, tata letak tempat duduk, larangan HP |
| **Orang lain mengerjakan** (joki) | Login hanya membuktikan kredensial, bukan identitas fisik | Cek kartu peserta / wajah saat masuk ruang, pengawas kenal siswa |
| **Internet / AI eksternal di perangkat lain** | Di luar jangkauan browser ujian | Sama seperti HP kedua |
| **Screenshot / rekam layar di OS** | Bisa dihambat sebagian (deteksi blur/fokus), tidak bisa dicegah total di semua OS | Kebijakan + pengawas + soal diacak per peserta |
| **Listrik / jaringan padam** | Fisik | UPS, server cadangan, rencana ujian susulan |

Yang **bisa** dicegah aplikasi: curang lewat tab lain di perangkat
yang sama (dideteksi + auto-block), mengubah waktu perangkat,
menyalin soal lewat copy/paste / unduh (copy dinonaktifkan + soal
tidak diekspos), memakai token bocor (expiry + hangus saat token
baru diterbitkan), dan mengubah nilai dari browser (penilaian server
+ audit log).
