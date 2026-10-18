# KONTRAK BERSAMA — CBT SMK v2

**Pemilik:** Tim 1 (Scaffolding + Auth + RBAC).
**Status:** Kontrak ini MENGIKAT semua tim. Bila kontrak harus berubah, ubah
dokumen ini + kode terkait + seed dalam SATU PR, dan beri tahu semua tim.
**Sumber keputusan:** `ANALISIS.md` (Bab 4, 6, 7, 9, 11, 14, 16) + keputusan
final user 16 Okt 2026 ("Setuju semua").

---

## 0. Keputusan arsitektur yang tidak boleh dilanggar

| Keputusan | Isi |
|---|---|
| Monorepo | `backend/` (Go) + `frontend/` (React + TS) + `docs/` + `deploy/` |
| Modul Go | **SATU modul: `cbt-smk-v2/backend`** (file `backend/go.mod`). TIDAK ADA `go.mod` di root — root hanya berisi `Makefile`, `.env.example`, `.gitignore`, `docs/`. Perintah Go selalu dijalankan dari `backend/` (lihat `Makefile`). |
| Satu binary | Frontend hasil build di-embed (`backend/dist` disalin dari `frontend/dist` oleh `make build`) + route `/*` fallback ke `index.html`. `/api/*` dan `/ws/*` didahulukan. |
| Auth | **Session cookie server-side + CSRF. TIDAK ADA kode JWT.** |
| Password | Argon2id (lihat §4). |
| Nilai | SELALU dihitung backend. Frontend TIDAK PERNAH menghitung nilai. |
| Deploy utama | Server lokal sekolah (LAN, semi-daring). **cPanel DICORET.** |
| Label tipe soal | **"5 tipe standar AKM + 11 varian/ekstensi aplikasi"**. DILARANG mengklaim "16 tipe standar Kemendikbud". |
| UI | **NOL emoji** di UI — ikon `lucide-react`. Primary `#2563EB`, font Plus Jakarta Sans. |

---

## 1. Format respons API

### 1.1 Sukses

```json
{ "data": { "...": "..." } }
```

List dengan pagination:

```json
{ "data": [ "..." ], "meta": { "limit": 50, "offset": 0, "total": 123 } }
```

### 1.2 Error

```json
{ "error": { "code": "validation_error", "message": "Penjelasan untuk user." } }
```

Kode error yang dipakai bersama:

| Kode | HTTP | Arti |
|---|---|---|
| `validation_error` | 400 | Body / query tidak valid |
| `unauthenticated` | 401 | Belum login / sesi tidak valid / kedaluwarsa |
| `forbidden` | 403 | Login, tapi tidak punya izin (permission / scope) |
| `PASSWORD_CHANGE_REQUIRED` | 403 | Login benar, tapi `must_change_password=true` — wajib ganti password dulu, session TIDAK dibuat |
| `csrf_token_invalid` | 403 | Header `X-CSRF-Token` tidak cocok dengan session |
| `not_found` | 404 | Route / resource tidak ada (untuk resource yang tidak berhak: kembalikan 404, JANGAN 403, agar tidak membocorkan keberadaan — anti-IDOR) |
| `conflict` | 409 | Pelanggaran nilai unik (username/email duplikat) atau guard bisnis |
| `admin_terakhir` | 409 | Aksi ditolak karena akan menghilangkan admin aktif terakhir (UC20) |
| `account_inactive` | 403 | Akun nonaktif |
| `login_locked` | 429 | Akun terkunci sementara (5x login gagal / 15 menit) |
| `rate_limited` | 429 | Rate limit per-IP terlampaui |
| `internal_error` | 500 | Kesalahan server |
| `invalid_credentials` | 401 | Email/username atau password salah (pesan generik — jangan bedakan "akun tidak ada" vs "password salah" di respons) |

### 1.3 Aturan umum handler

- Hapus / aksi destruktif HANYA via `POST` (atau `DELETE`), TIDAK PERNAH via `GET`.
- Frontend TIDAK PERNAH menghitung nilai; nilai dihitung backend.
- Endpoint soal ujian siswa TIDAK PERNAH mengembalikan kunci jawaban, poin benar/salah, atau pembahasan (kecuali pembahasan tryout SETELAH submit, bila diizinkan settings). Pembahasan (`soal.pembahasan`, migrasi 0011) diisi guru lewat endpoint bank soal dan dikembalikan di endpoint guru + di `GET .../hasil` SETELAH submit bila tampil_hasil efektif (Tryout ATAU paket `tampil_hasil` AND settings `tampil_hasil_ke_siswa`) — JANGAN pernah di respons attempt sebelum submit.
- `Content-Type` request mutasi: `application/json` (kecuali upload multipart yang ditentukan).
- Semua handler yang butuh login memeriksa izin lewat `audit.CurrentUser` /
  `audit.HasPermission` (konteks diisi middleware Tim 1) — 401 bila context
  kosong, 403 bila tidak punya permission.

---

## 2. Session cookie + CSRF

### 2.1 Mekanisme

- Login sukses -> server membuat baris di tabel `sessions` (Postgres) dan
  mengirim 2 cookie:
  - `cbt_session` — nilai = ID session (UUID). `HttpOnly`, `SameSite=Lax`,
    `Path=/`, `Secure` bila `APP_ENV=prod`, `MaxAge` 12 jam.
  - `cbt_csrf` — nilai = `csrf_token` session. **Bukan** `HttpOnly`
    (agar JS bisa membacanya), `SameSite=Lax`, `Path=/`, `Secure` bila prod.
- Data session (user_id, role, csrf_token, expiry) tinggal di SERVER (DB).
  Restart server TIDAK membuat logout massal.
- TTL session: 12 jam. Idle timeout: 2 jam tanpa request -> session
  dianggap kedaluwarsa (dihapus saat diakses lagi).
- Logout -> baris session dihapus + kedua cookie di-clear (`MaxAge: -1`).

### 2.2 CSRF

- Untuk `POST`/`PUT`/`PATCH`/`DELETE`, client WAJIB mengirim header
  `X-CSRF-Token` == `csrf_token` session (dibaca frontend dari cookie
  `cbt_csrf`). Perbandingan constant-time di server.
- Dilewati untuk: method aman (`GET`/`HEAD`/`OPTIONS`),
  `POST /api/v1/auth/login`, `GET /api/v1/auth/csrf`.
- `POST /api/v1/auth/ganti-password` mode kredensial (tanpa session)
  tidak memakai CSRF — ia mengandalkan password lama yang benar.
- Middleware `CSRF` global (Tim 1) + wrapper frontend `apiClient`
  (membaca cookie `cbt_csrf` -> header `X-CSRF-Token` otomatis).

### 2.3 Context auth (backend)

Middleware `RequireAuth` mengisi context Echo (dipakai `audit.CurrentUser`):

| Kunci | Isi |
|---|---|
| `auth_user_id` | UUID user (string) |
| `auth_role` | Kode role (`admin`, `proktor`, `guru`, `pengawas`, `siswa`, `penguji_ukk`) |
| `auth_permissions` | `[]string` permission role (dari DB `role_permissions` JOIN `permissions`; fallback matriks statis `rbac` bila DB belum di-seed) |
| `auth_siswa_id` | = `auth_user_id` sampai Tim 2 memetakan profil siswa |
| `auth_csrf_token` | `csrf_token` session aktif |
| `auth_session_id` | ID session aktif |

---

## 3. Signature `Register` tiap paket (WAJIB SAMA)

```go
func Register(e *echo.Echo, pool *pgxpool.Pool, q *db.Queries, cfg *config.Config)
```

- `main.go` (`backend/cmd/server/main.go`) memanggil `Register` untuk:
  `auth`, `audit`, `master`, `banksoal`, `exam`, `monitor`, `rekap`,
  `settings`, `analisis`, `event`, `backup`, `ukk`, `sod`,
  `notifikasi`. **JANGAN hapus pemanggilan `Register`.**
- Paket STUB (`master`, `exam`, `monitor`) dibuat Tim 1 agar `go build`
  lolos — tim terkait mengganti isi STUB dengan implementasi penuh TANPA
  mengubah signature.
- Middleware global dipasang di `main.go` (Logger -> Recover -> RateLimit
  -> RequireAuth -> CSRF). Handler tiap paket TIDAK perlu memasang
  RequireAuth/CSRF lagi, tetapi WAJIB memeriksa permission sendiri
  (pola `requirePerm` di `banksoal.go`) karena penegakan ada di server,
  bukan hanya sembunyikan menu di frontend.
- `q *db.Queries` diterima semua `Register`, tetapi sampai `sqlc generate`
  berjalan, handler memakai `pool` langsung dengan SQL yang sama seperti
  di `backend/db/queries/*.sql` (konvensi Tim 5, lihat `backend/db/db.go`).

---

## 4. Password

- Algoritma: **Argon2id** via `golang.org/x/crypto/argon2`.
- Parameter: `time=1`, `memory=65536` KiB (64 MiB), `threads=4`,
  `keyLen=32`, salt acak 16 byte.
- Format simpan:
  `$argon2id$v=19$m=65536,t=1,p=4$<base64-salt-tanpa-padding>$<base64-hash-tanpa-padding>`
- Password baru minimal 8 karakter (login, ganti-password, create-admin,
  reset-password — semua jalur memakai batas yang sama).
- `must_change_password=true` -> login ditolak 403
  `PASSWORD_CHANGE_REQUIRED` (tanpa session). Akun baru via CLI
  `create-admin` selalu `must_change_password=true`.
- Ganti password menghapus SEMUA session user (wajib login ulang).
- Rate limit login: 5x gagal per akun per 15 menit -> kunci 15 menit,
  percobaan berikutnya 429 `login_locked`.

### CLI akun (Tim 1)

```
cbt-smk-v2 create-admin --email E --password P --nama N
cbt-smk-v2 create-admin --username U --nama N        # password diminta interaktif
cbt-smk-v2 reset-password --email E --password P
cbt-smk-v2 migrate-up | migrate-down [n]
```

- `create-admin` HANYA bila tabel `users` kosong (bila tidak: error).
- TIDAK ADA akun demo / akun bawaan di seed maupun migrasi.
- Lupa password admin -> buat admin baru via `create-admin`, atau
  `reset-password`. Jangan edit hash password manual di database.

---

## 5. RBAC — daftar permission (KANONIS) + matriks role

**Daftar ini adalah SATU-SATUNYA sumber kebenaran.** `backend/internal/rbac`
(kode), `docs/KONTRAK.md` (dokumen ini), dan
`backend/db/seed/seed.sql` (Tim 6) WAJIB memakai daftar yang sama persis.
Total **67 permission** (revisi integrasi Fase 1: +`rekap.finalisasi`,
+`rekap.remedial` = 59; Fase 3: +`analisis.read`, +`analisis.export`
(perbaikan BUG-01), +`sod.override` = 62; UC20: +`pengguna.*` (5) = 67).

**Peringatan BUG-01 (jangan diulangi):** `analisis.*` sempat HANYA ada di
migrasi 0008 — grant migrasi 0008 ke proktor+guru diam-diam no-op pada
alur standar `migrate → seed` karena tabel `roles` masih KOSONG saat
migrasi berjalan (roles diisi seed, BUKAN migrasi). Akibatnya proktor+guru
403 di endpoint analisis. Aturan tetap: **seed adalah sumber kebenaran
grant**; migrasi boleh menyisipkan baris `permissions` (ON CONFLICT DO
NOTHING) tetapi TIDAK BOLEH diandalkan untuk grant `role_permissions`.

### 5.1 Daftar permission

```
master.jurusan.{read,create,update,delete}            (4)
master.kelas.{read,create,update,delete}             (4)
master.mapel.{read,create,update,delete}             (4)
master.guru.{read,create,update,delete}              (4)
master.siswa.{read,create,update,delete,import}      (5)
banksoal.{read,create,update,delete,lock,import}     (6)
paket.{read,create,update,delete}                    (4)
jadwal.{read,create,update,delete,pause,resume,unblock,usulan-unblock} (8)
ujian.{kerjakan,monitor}                             (2)
koreksi.{read,update}                                (2)
rekap.{read,export,finalisasi,remedial}                    (4)
settings.{read,update}                               (2)
audit.{read}                                         (1)
ukk.{read,create,update}                             (3)
event.{read,create,update,delete}                    (4)
backup.{create,restore}                              (2)
analisis.{read,export}                               (2)
sod.override                                         (1, admin saja)
pengguna.{read,create,update,reset-password,toggle-aktif} (5, admin saja — UC20, §8.7)
```

### 5.2 Matriks role -> permission

| Role | Permission | Jumlah |
|---|---|---|
| **admin** | SEMUA (termasuk `sod.override` dan `pengguna.*`) | 67 |
| **proktor** | `master.*.read`, `banksoal.read`, `jadwal.*` (semua 8), `ujian.monitor`, `koreksi.read`, `rekap.*` (semua 4), `settings.read`, `audit.read`, `analisis.*` (statistik TANPA teks pertanyaan — lihat §5.3), `event.*`, `backup.*` | 30 |
| **guru** | `master.*.read`, `banksoal.*`, `paket.*`, `jadwal.{read,create,update,delete}`, `ujian.monitor`, `koreksi.*`, `rekap.read`, `rekap.remedial`, `settings.read`, `analisis.*` | 27 |
| **pengawas** | `ujian.monitor`, `jadwal.read`, `jadwal.usulan-unblock`, `rekap.read` | 4 |
| **siswa** | `ujian.kerjakan`, `rekap.read` | 2 |
| **penguji_ukk** | `ukk.*`, `rekap.read` | 4 |

### 5.3 Aturan scope & pemisahan tugas (SoD)

- **Guru** hanya mapel + kelas yang ditugaskan (penegakan di server).
- **Proktor** hanya jadwal yang ditugaskan; **TIDAK BOLEH melihat isi
  soal** (hanya metadata bank via `banksoal.read`). Konsekuensi untuk
  analisis (Fase 3): `GET /api/v1/analisis/:jadwalId` + export CSV-nya
  MENYEMBUNYIKAN field `pertanyaan` bila pemanggil proktor — statistik
  (kesukaran, daya beda, pengecoh) tetap tampil; guru/admin tetap
  melihat pertanyaan.
- **Pengawas** hanya ruangan/sesinya, read-only + catat insiden;
  **TIDAK BOLEH melihat token** dan **TIDAK BOLEH melihat isi soal**.
  Pengawas BOLEH mengusulkan unblock via `jadwal.usulan-unblock`;
  eksekusi unblock tetap proktor/admin.
- **Siswa** hanya jadwal kelasnya + jendela waktu + token valid; tidak
  boleh melihat jadwal/nilai siswa lain walau menebak URL/ID (anti-IDOR:
  kembalikan 404, bukan 403).
- **Penguji UKK** hanya skema + kompetensi + peserta yang ditugaskan.
- SoD (sistem harus MENOLAK, bukan hanya menyembunyikan): pembuat bank
  != pengunci bank; pembuat soal != pemegang token; pengoreksi esai !=
  pengubah nilai final; penguji UKK != finalisator nilai UKK; pengawas
  mengajukan, proktor mengeksekusi; pengelola user != pengelola audit log;
  pengaju draft jadwal != finalisator jadwal.
- Prinsip: deny-by-default; satu sesi = satu role aktif; route guard di
  frontend DAN backend (frontend hanya sembunyikan menu, backend menolak
  aksi terlarang: 401/403 + catat audit).

### 5.4 Pengecualian darurat SoD (Fase 3, ANALISIS Bab 11.4)

Pengecualian diizinkan HANYA lewat endpoint ini — tidak pernah lewat
jalur lain, tidak pernah untuk ubah nilai tanpa jejak / hapus audit log
(audit_log append-only, tanpa endpoint ubah/hapus):

| Method & Path | Keterangan |
|---|---|
| `POST /api/v1/sod/override` | Permission `sod.override` (admin saja). Body `{aksi, entitas, entitas_id, alasan}`. `alasan` wajib min **20 karakter** (kurang -> 400). Efek: catat `sod_overrides` + `audit_log` dengan flag `SOD_OVERRIDE` (detail JSON berisi aksi + alasan + entitas) + buat `notifikasi` "Pengecualian SoD" untuk SEMUA user role admin kecuali actor. |
| `GET /api/v1/notifikasi` | Notifikasi milik sendiri, pagination `limit`/`offset`. |
| `POST /api/v1/notifikasi/:id/baca` | Tandai dibaca; bukan milik -> 404 (anti-IDOR). |

Tabel (migrasi 0012): `sod_overrides(id, actor_id, aksi, entitas,
entitas_id, alasan, created_at)`, `notifikasi(id, penerima_user_id,
judul, isi, dibaca default false, created_at)`.

---

## 6. Skema database (migrasi 0001, Tim 1)

Migrasi: `backend/migrations/0001_auth.up.sql` / `.down.sql`.
Seed DATA (bukan DDL) ada di `backend/db/seed/seed.sql` (Tim 6).
Migrasi 0002 = Tim 2 (master: jurusan, kelas, mapel, guru, siswa, ...).
Migrasi 0003-0005 = Tim 3 (bank soal, exam, rekap) — sudah ada dan
merujuk `users(id UUID)` + `mapel(id UUID)` dari migrasi Tim 1/2.

| Tabel | Kolom |
|---|---|
| `roles` | `id BIGSERIAL PK`, `code TEXT UNIQUE`, `name`, `description`, `is_active BOOL DEFAULT TRUE` |
| `permissions` | `id BIGSERIAL PK`, `code TEXT UNIQUE`, `description`, `group_name` |
| `role_permissions` | `role_id BIGINT -> roles(id)`, `permission_id BIGINT -> permissions(id)`, `PK(role_id, permission_id)` |
| `users` | `id UUID PK DEFAULT gen_random_uuid()`, `nama`, `email TEXT UNIQUE NULL`, `username TEXT UNIQUE NULL`, `password_hash`, `role TEXT CHECK IN (6 role)`, `must_change_password BOOL DEFAULT TRUE`, `is_active BOOL DEFAULT TRUE`, `created_at`, `updated_at`, `terakhir_masuk TIMESTAMPTZ NULL` (diisi saat login sukses; NULL = belum pernah login — UC20, migrasi 0013), `CHECK (email IS NOT NULL OR username IS NOT NULL)` |
| `sessions` | `id UUID PK`, `user_id UUID -> users(id) CASCADE`, `role`, `csrf_token`, `ip`, `user_agent`, `created_at`, `expires_at`, `last_seen_at` |
| `audit_log` | `id UUID PK`, `actor_id UUID NULL -> users(id) SET NULL`, `aksi`, `entitas`, `entitas_id TEXT`, `detail JSONB DEFAULT '{}'`, `ip`, `created_at`. **Append-only**: tidak ada endpoint ubah/hapus. |
| `settings` | `key TEXT PK`, `value JSONB`, `group_name NULL`, `description NULL`, `updated_at` |

Indeks: `users(lower(email))`, `users(lower(username))`, `users(role)`,
`sessions(user_id)`, `sessions(expires_at)`,
`audit_log(actor_id, created_at)`, `audit_log(entitas, entitas_id, created_at)`,
`audit_log(aksi, created_at)`.

**PENTING — seed settings (Tim 6, wajib diperbaiki):** `settings.value`
adalah `JSONB`. INSERT seed untuk nilai string WAJIB memakai literal JSONB
valid: `'"tertinggi"'::jsonb`, BUKAN `'tertinggi'` (itu bukan JSON valid
dan INSERT akan gagal). Nilai boolean/angka (`false`, `900`, `15`, `2`)
sudah valid sebagai JSONB. Kolom `group_name`/`description` di `settings`
disediakan nullable agar INSERT seed yang menyebut kedua kolom itu jalan.

---

## 7. Konvensi query sqlc

- File kueri: `backend/db/queries/<nomor>_<domain>.sql`
  (mis. `0001_auth.sql`). Konfigurasi: `backend/sqlc.yaml`
  (sqlc v1, driver postgresql, package `db`, output `backend/db`).
- Penamaan (WAJIB):
  - `GetXxx :one` — satu baris,
  - `ListXxx :many` — banyak baris,
  - `CreateXxx :one` — INSERT dengan `RETURNING *`,
  - `UpdateXxx :exec` — UPDATE,
  - `DeleteXxx :exec` — DELETE.
- Sampai `sqlc generate` dijalankan, handler memakai `pool` langsung
  dengan SQL yang sama seperti di berkas kueri (placeholder `db.Queries`
  di `backend/db/db.go` — Tim 5).
- **`make sqlc` masih terblokir, kini oleh 0005** (lihat §11.7):
  sqlc mem-parse SELURUH skema di `backend/migrations` dan 0005 merujuk
  tabel `attempts`/`jadwal` yang tidak ada. Migrasi 0002 (Tim 2) SUDAH ada.

---

## 8. Daftar endpoint Fase 1

Awalan: `/api/v1`. Format respons §1. Tim pelaksana di kolom terakhir.

### 8.1 Auth (Tim 1 — SELESAI)

| Method & Path | Keterangan |
|---|---|
| `POST /api/v1/auth/login` | email/username + password; 403 `PASSWORD_CHANGE_REQUIRED` bila flag aktif |
| `POST /api/v1/auth/logout` | Hapus session + clear cookie |
| `GET /api/v1/auth/me` | User + role + permissions |
| `POST /api/v1/auth/ganti-password` | Password lama + baru (min 8); boleh saat flag aktif; hapus semua session |
| `GET /api/v1/auth/csrf` | `{"csrf_token": "..."}` mentah (tanpa envelope `data`) untuk session aktif |

### 8.2 Master (Tim Master Data)

| Method & Path |
|---|
| `GET /api/v1/master/{jurusan,kelas,mapel,guru,siswa}` |
| `POST /api/v1/master/{jurusan,kelas,mapel,guru,siswa}` |
| `GET /api/v1/master/{jurusan,kelas,mapel,guru,siswa}/:id` |
| `PUT /api/v1/master/{jurusan,kelas,mapel,guru,siswa}/:id` |
| `DELETE /api/v1/master/{jurusan,kelas,mapel,guru,siswa}/:id` |
| `POST /api/v1/master/siswa/import` (CSV multipart) | IDEMPOTEN (Fase 3, BUG-02): NISN yang sudah ada -> `sudah_ada` (bukan gagal). Respons `{sukses, gagal, sudah_ada, errors}`. Rekomendasi: maks. 100 siswa per file. |

### 8.3 Bank soal (Tim 3)

| Method & Path |
|---|
| CRUD `/api/v1/bank-soal` |
| `POST /api/v1/bank-soal/:id/lock` , `POST /api/v1/bank-soal/:id/unlock` |
| `POST /api/v1/bank-soal/import` |
| CRUD `/api/v1/bank-soal/:bankId/soal` |

### 8.4 Paket + jadwal + ujian + koreksi (Tim Ujian)

| Method & Path |
|---|
| CRUD `/api/v1/paket` |
| `GET /api/v1/paket/:id/soal`, `POST /api/v1/paket/:id/soal`, `DELETE /api/v1/paket/:id/soal` |
| CRUD `/api/v1/jadwal` |
| `POST /api/v1/jadwal/:id/pause` , `.../resume`, `.../usulan-unblock`, `.../unblock` |
| `POST /api/v1/jadwal/:id/token`, `GET /api/v1/jadwal/:id/token-status` |
| `GET /api/v1/ujian/aktif` |
| `POST /api/v1/ujian/mulai` `{jadwal_id, token?}` |
| `GET /api/v1/ujian/attempt/:id` (soal TANPA kunci) |
| `POST /api/v1/ujian/attempt/:id/jawaban` `{soal_id, jawaban}` |
| `POST /api/v1/ujian/attempt/:id/ragu` `{soal_id, ragu}` |
| `POST /api/v1/ujian/attempt/:id/pelanggaran` `{jenis}` |
| `POST /api/v1/ujian/attempt/:id/submit` |
| `GET /api/v1/ujian/attempt/:id/status` |
| `GET /api/v1/koreksi/:jadwalId` |
| `POST /api/v1/koreksi/:jawabanId` `{skor}` |

### 8.5 Rekap + monitor + settings + audit

| Method & Path | Tim |
|---|---|
| `GET /api/v1/rekap?jadwal_id=` | Tim Rekap |
| `GET /api/v1/rekap/:jadwalId/export.csv` | Tim Rekap |
| `GET /api/v1/rekap/:jadwalId/berita-acara` | Tim Rekap |
| `POST /api/v1/monitor/ticket` `{jadwal_id}` (ticket berlaku 30 detik) | Tim Monitor |
| `WS /ws/monitor?ticket=` | Tim Monitor |
| `GET /api/v1/monitor/:jadwalId` | Tim Monitor |
| `GET /api/v1/settings`, `PUT /api/v1/settings` | Tim Settings |
| `GET /api/v1/audit` `?aksi=&entitas=&actor_id=&limit=&offset=` | Tim Audit |
| `POST /api/v1/sod/override`, `GET /api/v1/notifikasi`, `POST /api/v1/notifikasi/:id/baca` | Fase 3 — lihat §5.4 |
| `GET /api/v1/analisis/:jadwalId`, `GET /api/v1/analisis/:jadwalId/export.csv` | Fase 2/3 — proktor tanpa `pertanyaan` (§5.3) |
| `GET /api/v1/pengguna`, `POST /api/v1/pengguna`, `GET /api/v1/pengguna/:id`, `PUT /api/v1/pengguna/:id`, `POST /api/v1/pengguna/:id/reset-password`, `POST /api/v1/pengguna/:id/toggle-aktif` | Manajemen Pengguna staf — UC20, lihat §8.7 |

### 8.6 Aturan ujian yang mengikat endpoint di atas

- Token kedua membuat token pertama hangus; token expiry default 900 detik
  (dapat diubah di Settings: `token_expiry_detik`).
- Toleransi telat default 15 menit (`toleransi_telat_menit`).
- Pause MEMBEKUKAN timer (simpan sisa waktu saat pause, lanjutkan dari
  nilai itu saat resume) — bukan sekadar gerbang masuk.
- Auto-block: pelanggaran ke-1 = peringatan, ke-2 = auto-block
  **per-attempt** (bukan mengubah status akun). Unblock = lanjut sisa
  waktu; hanya Proktor/Admin yang mengeksekusi.
- Tryout: tanpa token + pembahasan langsung.
- Nilai esai disembunyikan sampai final
  (`tampil_hasil_ke_siswa=false` default).
- KKTP per mapel/TP (bukan hardcode KKM=75). Remedial memakai nilai
  tertinggi. UKK: teori 40% + praktik 60%; rubrik Proses 30% / Produk
  50% / Sikap 20%; lulus >= 70.

### 8.7 Manajemen Pengguna staf (UC20)

Hanya akun **STAF** (role != `siswa`) yang dikelola di sini. Akun siswa
tetap dikelola lewat halaman Siswa (paket master). Akun siswa yang
ditarget lewat endpoint ini mendapat **404** (anti-IDOR).

| Method & Path | Permission | Keterangan |
|---|---|---|
| `GET /api/v1/pengguna?q=&role=&is_active=` | `pengguna.read` | List akun staf. `role` wajib salah satu role staf (`siswa` -> 400). Pagination `?page=&per_page=` (pola master). Field: `id, username, nama, email, role, is_active, must_change_password, terakhir_masuk, created_at`. |
| `POST /api/v1/pengguna` | `pengguna.create` | Body `{username, nama, email, role, password}` -> **201**. |
| `GET /api/v1/pengguna/:id` | `pengguna.read` | Detail; tidak ada / role siswa -> 404. |
| `PUT /api/v1/pengguna/:id` | `pengguna.update` | Body `{nama, email, role}` -> 200. Field yang tidak dikirim dipertahankan. |
| `POST /api/v1/pengguna/:id/reset-password` | `pengguna.reset-password` | Body `{password_baru}` -> 200. |
| `POST /api/v1/pengguna/:id/toggle-aktif` | `pengguna.toggle-aktif` | Body `{is_active}` -> 200. |

Aturan yang mengikat (lihat juga spesifikasi UC20 di
`docs/uml/UC20-KELOLA-PENGGUNA.md`):

- **Validasi buat:** username wajib + unik (cek case-insensitive;
  duplikat -> 409 `conflict`); nama wajib; email opsional, bila diisi
  harus valid dan unik (duplikat -> 409); role wajib salah satu dari
  5 role staf (`admin, proktor, guru, pengawas, penguji_ukk`);
  role `siswa` -> **400**; password minimal **8 karakter** (sama
  seperti kebijakan auth). Password disimpan Argon2id;
  akun baru `must_change_password=true`.
- **PUT:** username **TIDAK** bisa diubah (diabaikan bila dikirim).
  Role baru tidak boleh `siswa` (400). Bila role target diubah dari
  `admin` ke non-admin, pastikan minimal 1 admin aktif tersisa,
  bila tidak -> 409 `admin_terakhir`.
- **Reset password:** hash baru + `must_change_password=true` +
  **hapus SEMUA sesi** user itu agar sesi lama tidak bertahan.
  Target siswa -> 404.
- **Toggle aktif:** DILARANG menonaktifkan akun sendiri (-> 403
  `forbidden`); DILARANG menonaktifkan admin terakhir yang masih
  aktif (-> 409 `admin_terakhir`). **TIDAK ADA hard delete** —
  menonaktifkan saja.
- **Audit:** semua mutasi dicatat ke `audit_log` dengan aksi
  `pengguna.create / pengguna.update / pengguna.reset-password /
  pengguna.toggle-aktif`, entitas `users`.
- **Login terakhir:** `users.terakhir_masuk` diisi saat login sukses
  (migrasi 0013); NULL = belum pernah login.

---

## 9. Frontend — rute & kepemilikan halaman

> Catatan verifikasi Fase 3 (16 Okt 2026): §9 ini dicocokkan rute per
> rute terhadap `frontend/src/App.tsx` dan dinyatakan SESUAI —
> termasuk `/adm/rekap/:jadwalId/berita-acara` (param jadwalId),
> `/403` (tujuan redirect `RequireAuth staffOnly`, `App.tsx:57`),
> `/kartu` + `/kartu/:jadwalId`, dan catch-all `NotFoundPage`.
> §5 juga terverifikasi: total 67 permission + matriks
> (admin 67, proktor 30, guru 27, pengawas 4, siswa 2,
> penguji_ukk 4) sesuai `backend/db/seed/seed.sql`.

Router: React Router v7 (`BrowserRouter`). Layout: `AdminLayout`
(sidebar + header, hamburger di HP, responsif) untuk `/adm/*`;
`Layout` untuk halaman siswa.

| Route | Halaman | Pemilik |
|---|---|---|
| `/login` | `src/pages/LoginPage.tsx` | Tim 1 |
| `/ganti-password` | `src/pages/GantiPasswordPage.tsx` | Tim 1 |
| `/` | `src/pages/siswa/SiswaHomePage.tsx` | Tim 1 |
| `/adm` | `src/pages/admin/DashboardPage.tsx` | Tim 1 |
| `/adm/master/jurusan` | `.../master/JurusanPage.tsx` | Tim Master |
| `/adm/master/kelas` | `.../master/KelasPage.tsx` | Tim Master |
| `/adm/master/mapel` | `.../master/MapelPage.tsx` | Tim Master |
| `/adm/master/guru` | `.../master/GuruPage.tsx` | Tim Master |
| `/adm/master/siswa` | `.../master/SiswaPage.tsx` | Tim Master |
| `/adm/master/siswa/import` | `.../master/ImportSiswaPage.tsx` | Tim Master |
| `/adm/bank-soal` | `.../banksoal/BankSoalListPage.tsx` | Tim Bank Soal |
| `/adm/bank-soal/baru` | `.../banksoal/BankSoalFormPage.tsx` | Tim Bank Soal |
| `/adm/bank-soal/:id` | `.../banksoal/BankSoalDetailPage.tsx` | Tim Bank Soal |
| `/adm/paket` | `.../paket/PaketPage.tsx` | Tim Ujian |
| `/adm/jadwal` | `.../jadwal/JadwalPage.tsx` | Tim Ujian |
| `/adm/monitor` | `.../monitor/MonitorPage.tsx` | Tim Monitor |
| `/adm/rekap` | `.../rekap/RekapPage.tsx` | Tim Rekap (ada) |
| `/adm/rekap/:jadwalId/berita-acara` | `.../rekap/BeritaAcaraPage.tsx` | Tim Rekap (ada) |
| `/adm/analisis` | `.../analisis/AnalisisButirPage.tsx` | Tim Fase 2 |
| `/adm/kompetensi` | `.../analisis/KompetensiPage.tsx` | Tim Fase 2 |
| `/adm/event` | `.../event/EventPage.tsx` | Tim Fase 2 |
| `/adm/ruangan` | `.../event/RuanganPage.tsx` | Tim Fase 2 |
| `/adm/kartu` | `.../kartu/KartuPesertaPage.tsx` | Tim Fase 2 |
| `/adm/ukk` | `.../ukk/UkkListPage.tsx` | Tim Fase 2 |
| `/adm/ukk/:id` | `.../ukk/UkkDetailPage.tsx` | Tim Fase 2 |
| `/adm/ukk/:id/nilai` | `.../ukk/UkkNilaiPage.tsx` | Tim Fase 2 |
| `/adm/ukk/:id/sertifikat/:siswaId` | `.../ukk/UkkSertifikatPage.tsx` | Tim Fase 2 |
| `/adm/backup` | `.../backup/BackupPage.tsx` | Tim Fase 2 |
| `/adm/settings` | `.../settings/SettingsPage.tsx` | Tim Settings (ada) |
| `/ujian/:attemptId` | `src/pages/siswa/UjianPage.tsx` | Tim Ujian |
| `/hasil/:attemptId` | `src/pages/siswa/HasilPage.tsx` | Tim Ujian |
| `/kartu` | `src/pages/siswa/KartuSayaPage.tsx` | Tim Fase 3 (model 1-KARTU: satu kartu per siswa untuk semua ujian, unduh PDF dari backend) |
| `/kartu/:jadwalId` | redirect → `/kartu` (`<Navigate to="/kartu" replace />` di `App.tsx`) | Tim Kartu 1-Kartu (17 Okt 2026; sebelumnya `KartuSayaPage`, pilih jadwal → kartu sendiri) |
| `/403` | `src/pages/ForbiddenPage.tsx` | Tim Tambahan 3 (tujuan redirect `RequireAuth staffOnly`) |
| `/pengawas/monitor` | `src/pages/admin/monitor/MonitorPage.tsx` (dipakai ulang) | Tim Monitor |
| `*` (catch-all) | `src/pages/NotFoundPage.tsx` | Tim Tambahan 3 |
| `/adm/*` (tidak cocok) | `src/pages/NotFoundPage.tsx` | Tim Tambahan 3 |

- Halaman tim lain yang belum ada dibuat Tim 1 sebagai **STUB** bernama
  jelas (`STUB — ...`) agar `npm run build` lolos; tim terkait mengganti
  isi STUB tanpa mengubah path / nama export.
- `apiClient` (`src/lib/api.ts`, Tim 1): base `/api/v1` — pemanggil boleh
  memakai path pendek (`/settings`) atau lengkap (`/api/v1/settings`);
  keduanya dinormalisasi ke `/api/v1/...`. `credentials: "include"`,
  baca cookie `cbt_csrf` -> header `X-CSRF-Token` untuk mutasi, throw
  `ApiError` (dengan `code`, `status`, `message`) saat `!res.ok`.
  Retry 429 (Fase 3, BUG-03): respons 429 berkode `rate_limited` diulang
  otomatis maks 3x (GET/HEAD + POST saja) dengan backoff 500ms -> 1s ->
  2s, mengikuti header `Retry-After` bila ada (dibatasi maks 10 detik).
- State: `useAuthStore` (Zustand: user, login, logout, fetchMe),
  `useExamStore` (Zustand: attempt, answers, ragu, setAnswer, toggleRagu).
- Komponen UI bersama di `src/components/ui/` (Button, Input, Card,
  Dialog, Table, Badge, Select — gaya shadcn, Tailwind).
- CSS: `src/index.css` — Tailwind 4 (`@import "tailwindcss"`), CSS vars,
  primary `#2563EB`, font Plus Jakarta Sans, teks soal 16-17px, timer
  20-24px monospace, target sentuh min 44px, NOL emoji di UI.

---

## 10. Konfigurasi & cara jalan

```
cp .env.example .env     # isi DATABASE_URL & SESSION_SECRET (wajib di prod)
make migrate-up          # migrasi ter-embed via CLI server
make seed                # seed PRODUKSI (role + 67 permission + matriks + settings)
go run ./cmd/server create-admin --email ... --nama ...   # dari backend/
make build               # frontend + binary satu file
make test                # go test + vitest
```

Variabel env: `DATABASE_URL` (wajib), `PORT` (default 9988),
`SESSION_SECRET` (wajib di prod), `APP_ENV` (`dev`/`prod`),
`ADMIN_BOOTSTRAP` + `ADMIN_EMAIL` + `ADMIN_PASSWORD` (opsional, hanya
bila users kosong). File `.env` dicari di cwd, `../.env`, `../../.env`.

**Timeout HTTP server (Fase 3, BUG-02):** `ReadTimeout` 60 dtk,
`WriteTimeout` **120 dtk**, `IdleTimeout` 120 dtk (via
`echo.StartConfig.BeforeServeFunc` — Echo v5 tidak punya `e.Server`).
Import siswa tetap dianjurkan **maks. 100 siswa per file** (±17 dtk).

**Rate limit (Fase 3, BUG-03):** 100 req/menit/IP untuk endpoint umum,
**300 req/menit/IP** untuk `/api/v1/ujian/attempt/*`
(jawaban/ragu/pelanggaran/submit/status). Penolakan 429 berkode
`rate_limited` + header `Retry-After` (detik). Rate limit login TIDAK
berubah: 5x gagal/15 menit -> kunci 15 menit, 429 `login_locked`.

**Jam jadwal:** `gabungTanggalJam` menerima `HH:MM` (dinormalisasi ke
`HH:MM:SS`) selain `HH:MM:SS` — POST jadwal dengan jam `"08:00"` -> 201.

Target `Makefile`: `build`, `run`, `dev`, `migrate-up`, `migrate-down`,
`seed`, `sqlc`, `test`, `test-backend`, `test-frontend`, `clean`.

---

## 11. Catatan konflik antar-dokumen yang sudah diselesaikan Tim 1

1. **Lokasi `go.mod`**: tugas awal menyebut `go.mod` di root, tetapi
   `backend/go.mod` (modul `cbt-smk-v2/backend`) SUDAH ada dan seluruh
   kode tim lain mengimpor `cbt-smk-v2/backend/...`. Memakai dua modul
   akan memecah tree dan merusak impor tim lain. Keputusan: SATU modul
   `cbt-smk-v2/backend`; tidak ada `go.mod` di root.
2. **`deploy/start.sh` (Tim 6) menjalankan `go build ./cmd/server` dari
   root** — path itu tidak ada (yang ada `backend/cmd/server`). Tim 6
   perlu menyesuaikan skrip ke `make build` / build dari `backend/`.
   Sampai diperbaiki, cara jalan yang benar adalah §10 dokumen ini.
3. **`frontend/src/lib/api.ts`**: versi awal tim lain membaca cookie
   `csrf_token`. Kontrak ini menetapkan cookie **`cbt_csrf`** (yang
   ditulis backend Tim 1); `api.ts` sudah diselaraskan Tim 1 dan tetap
   kompatibel dengan pemanggil lama (path `/api/v1/...` maupun pendek).
4. **Seed settings (Tim 6)**: lihat §6 — nilai string WAJIB literal
   JSONB valid (`'"tertinggi"'`).
5. **`backend/internal/config/config.go`**: field lama Tim 5
   (`DatabaseURL`, `Addr`) DIPERTAHANKAN; Tim 1 menambahkan field baru +
   `Load()`. Jangan hapus field lama.
6. **`backend/internal/rbac/rbac.go` & `session/session.go`**: stub
   koordinator diganti implementasi penuh Tim 1 (tidak ada kode tim lain
   yang memakai API lama — diverifikasi via grep sebelum mengganti).
7. **BLOKER BOOT — migrasi 0005 (Tim Rekap) merujuk tabel yang tidak ada
   (16 Okt 2026, terverifikasi Tim 1):** `0005_rekap.up.sql` merujuk
   `attempts` (0004 membuat `attempt`), `jadwal` (0004 membuat
   `jadwal_ujian`), dan `siswa` (ada di 0002 — OK). Akibatnya
   `migrate-up` GAGAL di 0005 (`relation "attempts" does not exist`,
   schema_migrations jadi dirty) dan **server tidak bisa boot sama
   sekali** pada database mana pun (auto-up on boot berhenti).
   Perbaikan milik Tim Rekap + Tim Ujian (samakan nama tabel — rekomendasi:
   0005 memakai `attempt` + `jadwal_ujian` seperti 0004) dalam migrasi
   BARU atau perbaikan 0005 SEBELUM ada database produksi yang
   menerapkannya; jangan ubah 0005 setelah diterapkan di produksi.
   Sampai diperbaiki, `make migrate-up`, `make build` + jalan, dan
   `make sqlc` semuanya terblokir di titik ini. Tim 1 memverifikasi alur
   auth + SPA memakai binary test tanpa 0005 (lihat laporan Tim 1).
8. **File yang sudah dibuat tim lain SEBELUM / SELAMA kerja Tim 1**
   (Tim 1 TIDAK menulis ulang, hanya mengadopsi/menyesuaikan titik
   integrasi): `backend/go.mod` + `backend/go.sum`, `backend/db/db.go`,
   `backend/db/queries/{banksoal,exam,monitor,rekap,settings}.sql`,
   `backend/db/seed/seed.sql`, `backend/migrations/0002-0005`,
   `backend/internal/{audit,banksoal,grading,rekap,settings,master,exam,monitor}/`,
   `frontend/package.json`, `frontend/vite.config.ts`,
   `frontend/tsconfig.json`, `frontend/index.html`,
   `frontend/src/index.css`, `frontend/src/pages/**` (kecuali halaman
   Tim 1), `frontend/src/lib/{banksoal,tipeSoal,exam}.ts`,
   `frontend/src/components/{exam,soal}/`, `deploy/`, `README.md`,
   `CARA-JALAN.txt`, `docs/{KEAMANAN,PANDUAN-DEPLOY-SEKOLAH}.md`.
   Perbedaan terhadap spesifikasi tugas Tim 1 yang ditemukan di file
   tersebut (dicatat, bukan diubah sepihak): `package.json` memakai
   Vite 6 (bukan 8), tanpa `@fontsource/plus-jakarta-sans` (font via
   Google Fonts di `index.css`) dan tanpa Vitest; `deploy/start.sh`
   membangun `./cmd/server` dari root (lihat butir 2).
9. **Perbaikan integrasi Fase 1 (16 Okt 2026, Tim Integrasi):** (a) 0005
   memakai nama kanonis `attempt` + `jadwal_ujian` (bukan `attempts` /
   `jadwal`); kueri berita-acara rekap ikut diperbaiki (mapel via JOIN
   `paket_ujian`). (b) Seed: 59 permission (+`rekap.finalisasi`,
   +`rekap.remedial`); proktor TANPA `paket.*` (28), guru +`rekap.remedial`
   (25), pengawas +`jadwal.usulan-unblock` (4); settings string memakai
   literal JSONB (`'"tertinggi"'::jsonb`). `rbac.go`, tes RBAC, dan §5
   dokumen ini diselaraskan. (c) `nilai_akhir` kini DISINKRONKAN dari
   alur ujian (`syncNilaiAkhir` di `exam.go`: finalisasiAttempt +
   koreksiJawaban) — sebelumnya rekap selalu kosong karena nilai hanya
   tertulis di `attempt.nilai_akhir`. (d) `users.role` tunggal — tidak
   ada `user_roles`; `createStudentAccountTx` mengisi `nama` +
   `role='siswa'` (sebelumnya INSERT tanpa `nama`/`role` gagal NOT NULL).
   (e) Migrasi 0006: FK `soal -> kisi_kisi / kompetensi` (ON DELETE SET
   NULL). (f) Guard hapus: kelas/mapel/siswa ditolak 409 bila dipakai
   `jadwal_ujian` / paket / attempt / nilai akhir. (g) Cek "bank dipakai
   paket" memakai `paket_soal JOIN soal` (tanpa `information_schema`).
   (h) Font Plus Jakarta Sans via `@fontsource` (tanpa Google Fonts);
   plugin `@tailwindcss/vite` ditambahkan — tanpanya TIDAK ADA utility
   class yang dihasilkan dan aplikasi tampil tanpa styling.
