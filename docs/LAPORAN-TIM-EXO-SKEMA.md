# LAPORAN TIM EXO — SKEMA DATA Extraordinary CBT v4.7.0

**Tim:** TIM EXO SKEMA DATA (analisis dokumen & dump seed)
**Tanggal:** 16 Oktober 2026
**Lokasi kerja:** `~/workspace/cbt-smk-v2` (laporan ini) — bahan di `~/workspace/extraordinary-cbt/`
**Status:** ANALISIS SAJA. Tidak ada kode yang diubah. Binary tidak pernah dijalankan.

---

## 0. Metode, Kepatuhan, dan Label Bukti

### 0.1 Bahan yang dibaca (satu-satunya sumber)

| # | Bahan | Cara baca |
|---|---|---|
| 1 | `~/workspace/extraordinary-cbt/PELAJARI-EXTRAORDINARY-CBT.md` | Dibaca penuh |
| 2 | `~/workspace/extraordinary-cbt/app/exo-dump-master.sql` (1.834 baris, 38 `CREATE TABLE`) | Dibaca penuh (DDL + `INSERT` seed + constraint + indeks). **Hanya dibaca, tidak dijalankan ke DB mana pun.** |
| 3 | `~/workspace/cbt-smk-v2/ANALISIS.md` | Dibaca (dipakai sebagai pembanding + koreksi yang sudah dicatat di sana, terutama Bab 6) |

### 0.2 Aturan keras yang dipatuhi

- **DILARANG menjalankan binary** Extraordinary CBT dalam bentuk apa pun — dipatuhi. Tidak ada proses binary yang dijalankan.
- **DILARANG membobol lisensi / crack / decompile / memakai kredensial** — dipatuhi. Tidak ada upaya apa pun terhadap `ecosystem_keys`, tidak ada nilai kunci/signature/token/password yang disalin ke laporan ini.
- **DILARANG mengarang fitur** — setiap klaim di laporan ini memakai salah satu label berikut.
- **DILARANG mengubah kode** — tidak ada file kode yang diubah. Hanya file laporan ini yang ditulis.
- **Nilai rahasia dari seed TIDAK direproduksi.** Laporan ini menyebut bentuk/keberadaan kolom dan hitungan baris seed, bukan isi kredensial: tidak ada hash password, token ujian, token API, payload/signature lisensi, maupun blob terenkode dari tabel `tools`.

### 0.3 Label bukti (berlaku untuk SEMUA klaim di laporan ini)

| Label | Arti |
|---|---|
| **TERBUKTI** | Terlihat langsung di dump (DDL / `INSERT` seed / constraint / indeks) atau di dokumen yang dikutip. Nomor baris/sumber disebut bila penting. |
| **DISIMPULKAN** | Tidak tertulis eksplisit di dump; diturunkan dari nama kolom/tabel, pola seed, judul menu di `menus`, atau dokumen pendamping. Alasan inferensi selalu ditulis. Bisa salah. |
| **TIDAK DIKETAHUI** | Tidak ada bukti di bahan yang dibaca. Tebakan dari nama saja dilarang ditegaskan — dicatat di Bab 7. |

### 0.4 Batasan paling penting dari bahan ini (baca dulu sebelum memakai laporan)

1. **Dump seed hampir kosong.** Dari 38 tabel, hanya 13 tabel yang punya baris seed, dan sebagian besar hanya 1–2 baris contoh. Semua tabel inti pelaksanaan ujian (`banksoals`, `soals`, `jadwals`, `siswa_ujians`, `jawaban_pesertas`, `hasil_ujians`, `pesertas`, …) berisi **0 baris seed**. Artinya: **perilaku runtime tidak bisa dibuktikan dari dump** — yang bisa dibuktikan hanyalah bentuk skema + segelintir nilai konfigurasi/contoh.
2. **Tidak ada logika di dump.** Dump hanya berisi `CREATE TABLE`, `INSERT`, `PRIMARY KEY`, `UNIQUE`, 1 `CHECK`, dan indeks. **Tidak ada `FOREIGN KEY`, `VIEW`, `TRIGGER`, atau `FUNCTION`** (grep: 0 kecocokan). Semua relasi antar tabel di laporan ini berstatus **DISIMPULKAN** dari penamaan kolom, bukan ditegakkan database.
3. **Nilai enum/status tidak ada di dump.** Semua kolom `status_ujian`, `status`, `tipe_soal`, `layout`, `sequence_type`, dsb. adalah integer/varchar tanpa tabel lookup dan tanpa `CHECK`. Nilai-nilainya **TIDAK DIKETAHUI** dari bahan ini.

---

## 1. Temuan Struktural Global (TERBUKTI, dari dump)

### 1.1 Inventaris

- **38 tabel** (TERBUKTI, hitungan `CREATE TABLE` = 38).
- **0 `FOREIGN KEY`** — tidak satu pun relasi dideklarasikan di database.
- **0 `VIEW`, 0 `TRIGGER`, 0 `FUNCTION`** — tidak ada logika di level DB.
- **1 `CHECK` constraint** — hanya di `tokens.status` (lihat §6.1).
- **5 `UNIQUE` constraint**: `jurusans.kode`, `personal_access_tokens.token`, `pesertas.no_ujian`, `settings.code`, `users.email`.
- **17 indeks** (selain PK/UNIQUE): tercantum per tabel di Bab 2 dan diringkas di Lampiran A.
- **1 tabel tanpa PRIMARY KEY**: `role_menus` (TERBUKTI — tidak ada constraint PK untuknya di dump). Semua tabel lain punya PK.

### 1.2 Ragam tipe ID (TERBUKTI — inkonsisten antar tabel)

| Pola ID | Tabel |
|---|---|
| `uuid` | Mayoritas (30 tabel), mis. `banksoals`, `soals`, `jadwals`, `pesertas`, `users` |
| `bigint` + sequence | `backups`, `feature_infos`, `personal_access_tokens` |
| `varchar(26)` / `char(26)` (mirip ULID) | `banksoal_chains.id`, `prompts.id`, `shortcuts.id`, `tools.id` |
| `varchar(100)` | `ecosystem_keys.id` |
| `varchar(255)` sebagai PK | `menus.code`, `settings.code` (natural key) |

**DISIMPULKAN:** skema ini hasil evolusi multi-versi / gabungan beberapa subsistem, bukan desain satu kali. Untuk v2, konsekuensi: samakan tipe ID (satu pola), jangan tiru campuran ini.

### 1.3 Ragam tipe waktu (TERBUKTI — semua tanpa zona waktu)

- Mayoritas `timestamp(0) without time zone`; sebagian `timestamp(6)` (`devices.last_verified_at`, `hasil_ujians`, `jawaban_pesertas`, `soals`, `room_members`, `prompts`, `jadwals.paused_at`, `siswa_ujians` tidak — lihat Bab 2).
- Beberapa kolom waktu justru **string**: `jadwals.tanggal` / `jadwals.mulai` (`varchar(50)`), `siswa_ujians.mulai_ujian` / `mulai_ujian_shadow` / `selesai_ujian` (`varchar(255)`), `hasil_ujians.duration` (`varchar(50)`), `backups.generated_date` (`varchar(255)`).
- **DISIMPULKAN:** perbandingan/urut waktu di aplikasi bergantung pada format string yang tidak terdokumentasi di dump → rawan salah urut / salah zona waktu. v2 wajib `timestamptz` untuk semua waktu kejadian.

### 1.4 Isi seed: 13 tabel berisi, 25 tabel kosong (TERBUKTI, hitungan `INSERT`)

| Tabel berisi seed | Baris |
|---|---|
| `role_menus` | 170 |
| `menus` | 99 |
| `feature_infos` | 20 |
| `settings` | 19 |
| `agamas` | 7 |
| `tools` | 4 |
| `ecosystem_keys` | 2 |
| `users` | 1 |
| `tokens` | 1 |
| `personal_access_tokens` | 1 |
| `medias` | 1 |
| `jurusans` | 1 |
| `devices` | 1 |

**25 tabel dengan 0 baris seed:** `backups`, `banksoal_chains`, `banksoal_summaries`, `banksoals`, `check_answers`, `competence_items`, `competences`, `events`, `group_members`, `groups`, `hasil_ujians`, `jadwals`, `jadwals_peserta`, `jawaban_pesertas`, `jawaban_soals`, `matpels`, `pesertas`, `prompts`, `room_members`, `rooms`, `sesi_schedules`, `shortcuts`, `siswa_ujians`, `soal_summaries`, `soals`.

**Konsekuensi:** semua inferensi perilaku ujian (Bab 4–6) bertumpu pada nama kolom + judul menu + dokumen pendamping, bukan pada data contoh. Itulah alasan banyak label **TIDAK DIKETAHUI** di laporan ini — dan itu disengaja.

### 1.5 Koreksi atas kajian sebelumnya (TERBUKTI dari skema + ANALISIS.md Bab 6)

1. **`code01–code16` adalah kode kelompok kompetensi / bagian penilaian, BUKAN daftar nama 16 tipe soal.** Bukti skema: `banksoals` punya `code01_max … code16_max`; `hasil_ujians` punya `code01_correct/wrong/point/void … code16_*`; `competence_items` punya `code`; `soals` terikat ke `competence_item_id` — dan **tidak ada** tabel/kolom pemetaan nama `tipe_soal` di dump. ANALISIS.md Bab 6 mencatat koreksi yang sama (situs resmi hanya mengklaim 6 label bentuk soal). Kajian lama yang menyebut "16 kode soal" adalah salah tafsir — **jangan dipakai lagi**.
2. **Jumlah menu seed = 99, bukan "100+".** PELAJARI menyebut "100+ permission". **TERBUKTI:** dump memuat 99 baris `menus`; `role_menus` untuk `EVIL` = 99 baris (sama dengan jumlah menu). Kemungkinan: versi dump lebih lama dari klaim situs, atau hitungan "100+" membulatkan/menghitung hal lain. **TIDAK DIKETAHUI** mana yang benar — yang pasti, di dump ini angkanya 99.
3. **Tidak ada tabel ledger / berita acara.** PELAJARI menyebut "ledger nilai" dan "berita acara". **TERBUKTI:** tidak ada tabel bernama `ledger`, `berita_acara`, atau sejenisnya. Yang ada hanya menu `0019A 'Daftar ledger peserta'` dan `0021F 'Kartu peserta event ujian'`. **DISIMPULKAN:** ledger/berita acara di Exo adalah hasil hitung/tampilan dari `hasil_ujians` (+ agregasi), bukan tabel tersendiri. Bukti perilaku (format BAU, kapan final) **TIDAK DIKETAHUI** dari dump.

---

## 2. Katalog 38 Tabel

Konvensi per tabel: **PK** = primary key (TERBUKTI). **FK→** = relasi **DISIMPULKAN** dari nama kolom (tidak ada FK di dump, kecuali dinyatakan lain). Kolom status/timestamp yang menggerakkan state machine ditandai dengan ⏱ / ⚙. Kolom yang maknanya hanya tebakan ditandai ❓ dan dirinci di Bab 7.

### 2.1 Master data & pengguna

#### 2.1.1 `agamas` — daftar agama (referensi)
- **Isi:** master agama. **Seed TERBUKTI:** 7 baris (Islam, Protestan, Katolik, Hindu, Budha, Kong Hu Cu, Kepercayaan Terhadap Tuhan Yang Maha Esa).
- **PK:** `id` uuid. **Kolom:** `kode` varchar(255) + indeks, `nama` varchar(255), `created_at` / `updated_at` timestamp(0).
- **Status/timestamp penggerak:** tidak ada kolom status.
- **FK→ (DISIMPULKAN):** diacu oleh `pesertas.agama_id`, `matpels.agama_id`.

#### 2.1.2 `jurusans` — daftar jurusan (referensi)
- **Isi:** master jurusan. **Seed TERBUKTI:** 1 baris generik (kode + nama untuk jurusan "UMUM" — nilai persis tidak direproduksi karena tidak relevan).
- **PK:** `id` uuid. **UNIQUE:** `kode`. **Kolom:** `kode` varchar(255) + indeks, `nama` varchar(255).
- **Catatan TERBUKTI:** **tidak punya kolom `created_at`/`updated_at`** — satu-satunya tabel referensi tanpa timestamp.
- **FK→ (DISIMPULKAN):** diacu oleh `pesertas.jurusan_id`, `matpels.jurusan_id` (yang bertipe `text` — ketidakcocokan tipe, lihat §2.1.4).

#### 2.1.3 `matpels` — mata pelajaran
- **Isi:** master mapel + target sasaran + tim pengoreksi.
- **PK:** `id` uuid. **Kolom:** `kode_mapel` varchar(100) + indeks, `jurusan_id` ❓ `text` NULLABLE, `agama_id` uuid NULLABLE, `correctors` ❓ `text` NULLABLE, `nama` varchar(255), `created_at`/`updated_at`.
- **FK→ (DISIMPULKAN):** `agama_id` → `agamas.id`; `jurusan_id` → `jurusans.id` (tipe `text` vs `uuid` target — **TERBUKTI** tidak cocok, **DISIMPULKAN** berisi satu/beberapa ID jurusan dalam bentuk teks; formatnya **TIDAK DIKETAHUI**); `correctors` **DISIMPULKAN** daftar ID pengguna pengoreksi, format **TIDAK DIKETAHUI**.
- **Status:** tidak ada kolom status. **Diacu oleh (DISIMPULKAN):** `banksoals.matpel_id`.

#### 2.1.4 `groups` — kelompok peserta (hierarkis)
- **Isi:** pengelompokan peserta. **Seed:** 0 baris.
- **PK:** `id` uuid. **Kolom:** `parent_id` uuid NULLABLE (⚙ hierarki), `name` varchar(255), `created_at`/`updated_at`. **Indeks:** `(parent_id, name)`.
- **FK→ (DISIMPULKAN):** `parent_id` → `groups.id` (self-reference, sub-kelompok).
- **TIDAK DIKETAHUI:** jenis kelompok (rombel? kelas? kelompok bebas?) — tidak ada kolom tipe/status. **DISIMPULKAN** dari `jadwals.group_ids` (teks): jadwal menunjuk ke kelompok via daftar ID teks.

#### 2.1.5 `group_members` — anggota kelompok
- **Isi:** keanggotaan peserta dalam kelompok (many-to-many). **Seed:** 0 baris.
- **PK:** `id` uuid. **Kolom:** `group_id` uuid, `student_id` uuid, `created_at`/`updated_at`. **Indeks:** `(group_id, student_id)`.
- **FK→ (DISIMPULKAN):** `group_id` → `groups.id`; `student_id` → `pesertas.id` (nama kolom `student_id`, tabel target bernama `pesertas` — pemetaan nama ini **DISIMPULKAN**).
- **Catatan TERBUKTI:** tidak ada `UNIQUE(group_id, student_id)` — duplikasi keanggotaan dimungkinkan di level DB.

#### 2.1.6 `pesertas` — peserta ujian (siswa)
- **Isi:** akun + identitas + status blokir peserta. **Seed:** 0 baris.
- **PK:** `id` uuid. **UNIQUE:** `no_ujian`. **Kolom:**
  - Identitas: `sesi` int NOT NULL, `no_ujian` varchar(50), `agama_id` uuid NOT NULL, `jurusan_id` uuid NOT NULL, `nama` varchar(255), `ava` ❓ text NULLABLE (**DISIMPULKAN** avatar/foto; format **TIDAK DIKETAHUI**).
  - Kredensial: `password` varchar(255) (format hash **TIDAK DIKETAHUI** dari skema), `api_token` ❓ varchar(255) NULLABLE + indeks `(no_ujian, api_token)` (tujuan token ini **TIDAK DIKETAHUI**; hubungannya dengan `personal_access_tokens` **TIDAK DIKETAHUI**).
  - ⚙ Status/blokir: `status` ❓ int DEFAULT 1 (**DISIMPULKAN** 1 = aktif, dari pasangan menu `0001G Aktifkan` / `0001H Nonaktifkan`; nilai lain **TIDAK DIKETAHUI**), `block_reason` varchar(255) NULLABLE, `antiblock` bool DEFAULT false (**DISIMPULKAN** pengecualian auto-block), `blocked_at` ⏱ timestamp(0) NULLABLE.
  - `created_at` / `updated_at` timestamp(0).
- **FK→ (DISIMPULKAN):** `agama_id` → `agamas.id`; `jurusan_id` → `jurusans.id`.
- **Diacu oleh (DISIMPULKAN):** `group_members.student_id`, `jadwals_peserta.peserta_id`, `sesi_schedules.student_id`, `room_members.student_id`, `devices.peserta_id`, `siswa_ujians.peserta_id`, `jawaban_pesertas.peserta_id`, `check_answers.peserta_id`, `hasil_ujians.peserta_id`.

#### 2.1.7 `users` — pengguna staf (admin/guru/proktor)
- **Isi:** akun staf. **Seed TERBUKTI:** 1 baris, role `EVIL`. (Nama/email/hash tidak direproduksi.)
- **PK:** `id` uuid. **UNIQUE:** `email`. **Kolom:** `name`, `email`, `password` varchar(255) (format hash **TIDAK DIKETAHUI** dari skema), `role` varchar(255) DEFAULT `'EVAL'`, `created_at`/`updated_at`. **Indeks:** `(email, role)`.
- **Nilai `role` TERBUKTI dari `role_menus`:** `EVIL`, `EVAL`, `OPRA`. **DISIMPULKAN** (dari PELAJARI + judul menu): EVIL = super admin, EVAL = evaluator/guru, OPRA = operator/proktor.
- **Catatan TERBUKTI:** tidak ada kolom status akun, `last_login_at`, atau penguncian akun.

### 2.2 Bank soal, soal, kompetensi

#### 2.2.1 `banksoals` — bank soal
- **Isi:** header bank soal + kunci + plafon poin per kode kompetensi. **Seed:** 0 baris.
- **PK:** `id` uuid. **Kolom:**
  - Identitas: `kode_banksoal` varchar(100), `matpel_id` uuid NOT NULL, `author` uuid NOT NULL, `competence_id` uuid NULLABLE.
  - ⚙ Kunci: `is_locked` bool DEFAULT false, `key_lock` ❓ varchar(255) NULLABLE, `lock_by` ❓ varchar(255) NULLABLE (**DISIMPULKAN**: `key_lock` = kunci/kode pengunci, `lock_by` = siapa yang mengunci; apakah `key_lock` berupa password/kode — **TIDAK DIKETAHUI**).
  - ❓ `persen` varchar(350) NOT NULL — **TIDAK DIKETAHUI** maknanya (hipotesis di Bab 7; judul menu `0010` / feature `AH003 'setting max persen'` hanya menegaskan ada pengaturan persen, bukan maknanya).
  - ⚙ Kompetensi: `code01_max … code16_max` int DEFAULT 0 (**DISIMPULKAN** plafon poin per kode kompetensi/bagian; lihat §1.5 — ini BUKAN tipe soal).
  - ❓ `supports` text NULLABLE (**DISIMPULKAN** lampiran/pendukung bank; format **TIDAK DIKETAHUI**), `last_sequence` int DEFAULT 0, `sequence_type` ❓ int DEFAULT 0 (nilai/makna **TIDAK DIKETAHUI**).
  - `created_at` / `updated_at`.
- **FK→ (DISIMPULKAN):** `matpel_id` → `matpels.id`; `author` → `users.id`; `competence_id` → `competences.id`.
- **Diacu oleh (DISIMPULKAN):** `soals.banksoal_id`, `jadwals.banksoal_id` (teks), `siswa_ujians.banksoal_id`, `jawaban_pesertas.banksoal_id`, `check_answers.banksoal_id`, `hasil_ujians.banksoal_id`, `banksoal_chains.banksoal_id`, `banksoal_summaries.banksoal_id`, `soal_summaries.banksoal_id`.

#### 2.2.2 `soals` — butir soal
- **Isi:** butir soal dalam satu bank. **Seed:** 0 baris.
- **PK:** `id` uuid. **Kolom:**
  - `banksoal_id` uuid, `tipe_soal` ❓ int NOT NULL (pemetaan integer → nama tipe **TIDAK DIKETAHUI** — tidak ada tabel lookup di dump; ANALISIS Bab 6: situs resmi hanya mengklaim 6 label).
  - `pertanyaan` text NOT NULL (teks soal).
  - `layout` ❓ int DEFAULT 1 (nilai/makna **TIDAK DIKETAHUI**).
  - `competence_item_id` uuid NULLABLE (FK→ `competence_items.id` **DISIMPULKAN** — jembatan soal → kompetensi).
  - ⚙ Penilaian: `absolute_eval` bool DEFAULT true, `case_sensitive` bool DEFAULT true (**DISIMPULKAN** flag ketelitian penilaian isian; ANALISIS Bab 4.3 mencatat opsi `case_sensitive` / `absolute_eval` per soal; perilaku persis tiap flag **TIDAK DIKETAHUI** dari dump).
  - `supports` ❓ text NULLABLE (paralel dengan `banksoals.supports`; format **TIDAK DIKETAHUI**).
  - `sequence` bigint DEFAULT 0 (**DISIMPULKAN** nomor urut soal dalam bank).
  - `paper_answer` bool DEFAULT false (**DISIMPULKAN** varian jawaban kertas; paralel dengan `jawaban_pesertas.paper_code`).
  - `created_at` / `updated_at` timestamp(6).
  - **Indeks:** `(banksoal_id, tipe_soal)`.
- **Catatan TERBUKTI:** tidak ada kolom status soal (draf/aktif/arsip), tidak ada kolom kunci jawaban di tabel ini (kunci ada di `jawaban_soals.correct`), tidak ada kolom pembahasan, tidak ada kolom bobot per soal (bobot hidup di `competence_items.point` / `banksoals.codeXX_max` — **DISIMPULKAN**).
- **Diacu oleh (DISIMPULKAN):** `jawaban_soals.soal_id`, `jawaban_pesertas.soal_id`, `soal_summaries.soal_id`.

#### 2.2.3 `jawaban_soals` — opsi / kunci jawaban soal
- **Isi:** opsi jawaban + penanda benar. **Seed:** 0 baris.
- **PK:** `id` uuid. **Kolom:** `soal_id` uuid, `text_jawaban` text NOT NULL, `correct` bool NOT NULL, `label_mark` ❓ varchar(255) NULLABLE (**DISIMPULKAN** label opsi mis. A/B/C/D; **TIDAK DIKETAHUI**), `created_at`/`updated_at` timestamp(6). **Indeks:** `(soal_id, correct)`.
- **FK→ (DISIMPULKAN):** `soal_id` → `soals.id`.
- **Catatan TERBUKTI:** tidak ada kolom urutan tampil opsi → mekanisme urutan opsi **TIDAK DIKETAHUI** (urutan insert? `label_mark`?).

#### 2.2.4 `competences` — kelompok kompetensi (header)
- **Isi:** definisi satu set kompetensi. **Seed:** 0 baris.
- **PK:** `id` uuid. **Kolom:** `code` varchar(255), `name` varchar(255) NULLABLE, `created_by` uuid DEFAULT zero-UUID (`00000000-…`), `total_part` ❓ int DEFAULT 1 (**DISIMPULKAN** jumlah bagian/part dalam set ini), `created_at`/`updated_at`.
- **Diacu oleh (DISIMPULKAN):** `competence_items.competence_id`, `banksoals.competence_id`.

#### 2.2.5 `competence_items` — item kompetensi (bagian penilaian)
- **Isi:** satu baris = satu bagian penilaian dalam satu kompetensi: tipe soal, jumlah soal, kode, poin. **Seed:** 0 baris.
- **PK:** `id` uuid. **Kolom:** `competence_id` uuid, `tipe_soal` ❓ int NOT NULL (domain integer yang sama dengan `soals.tipe_soal` — **DISIMPULKAN**; pemetaan nama **TIDAK DIKETAHUI**), `total_soal` int NOT NULL, `code` ❓ varchar(255) (**DISIMPULKAN** kode bagian, kemungkinan `code01…code16` — **TIDAK DIKETAHUI**, tidak ada seed/CHECK yang membuktikan), `point` int NOT NULL (**DISIMPULKAN** poin per soal benar pada item ini), `wrong_point` int DEFAULT 0 (**DISIMPULKAN** poin bila menjawab tapi salah — bisa 0/negatif; judul tutorial `AG005 'Tambah kompetensi jawab tapi salah'` mendukung bacaan ini), `name` varchar(255), `description` text NULLABLE, `part` ❓ int DEFAULT 1 (**DISIMPULKAN** nomor bagian), `index` ❓ int DEFAULT 0 (**DISIMPULKAN** urutan tampil), `created_at`/`updated_at`.
- **FK→ (DISIMPULKAN):** `competence_id` → `competences.id`.
- **Diacu oleh (DISIMPULKAN):** `soals.competence_item_id`.

#### 2.2.6 `banksoal_chains` — ❓ rantai bank soal
- **Isi:** **TIDAK DIKETAHUI.** **Seed:** 0 baris.
- **PK:** `id` varchar(26). **Kolom:** `banksoal_id` uuid, `chain` text NULLABLE. **Tidak punya timestamp.**
- **FK→ (DISIMPULKAN):** `banksoal_id` → `banksoals.id`.
- **Hipotesis (DISIMPULKAN, keyakinan rendah):** `chain` menyimpan rangkaian/relasi antar bank (mis. bank lanjutan, urutan pemakaian, atau rantai versi). Tidak ada bukti lain. Lihat Bab 7 — jangan dipakai sebagai dasar desain sebelum dikonfirmasi dari sumber resmi/dokumentasi produk.

### 2.3 Jadwal, sesi, attempt (pelaksanaan ujian)

#### 2.3.1 `jadwals` — jadwal ujian
- **Isi:** jadwal ujian + konfigurasi + status pause. **Seed:** 0 baris.
- **PK:** `id` uuid. **Kolom:**
  - ⚙ Rujukan (tipe janggal, TERBUKTI): `banksoal_id` **text** NOT NULL (bukan uuid seperti `banksoals.id` — **DISIMPULKAN** menampung satu/beberapa ID bank sebagai teks; format **TIDAK DIKETAHUI**), `group_ids` ❓ text NULLABLE (**DISIMPULKAN** daftar ID kelompok sebagai teks; format **TIDAK DIKETAHUI**), `event_id` uuid NULLABLE (FK→ `events.id` **DISIMPULKAN**).
  - ⚙ Waktu (string, TERBUKTI): `tanggal` varchar(50), `mulai` varchar(50), `lama` int NOT NULL (**DISIMPULKAN** durasi menit), `min_test` ❓ int DEFAULT 0 (**TIDAK DIKETAHUI** — hipotesis: waktu minimum / jumlah minimum; lihat Bab 7), `mulai_sesi` ❓ text NOT NULL (format/isi **TIDAK DIKETAHUI**).
  - ⚙ Status: `status_ujian` int NOT NULL (nilai enum **TIDAK DIKETAHUI** — tidak ada CHECK/lookup; menu `0011E Pause` / `0011F Resume` membuktikan minimal ada status dijeda, **DISIMPULKAN**), `paused_at` ⏱ timestamp(6) NULLABLE (**DISIMPULKAN** kapan dijeda), `transition_time` int DEFAULT 0 (**DISIMPULKAN** override per-jadwal atas `settings.TRANSITION_TIME`=60).
  - Lain: `alias` varchar(50) (**DISIMPULKAN** label/nama pendek jadwal), `sesi` int DEFAULT 1, `setting` ❓ text NOT NULL (isi/format **TIDAK DIKETAHUI** — **DISIMPULKAN** blob pengaturan terserialisasi), `view_result` bool DEFAULT false (**DISIMPULKAN** peserta boleh lihat hasil), `created_by` uuid DEFAULT zero-UUID, `participant` ❓ varchar(5) DEFAULT `'CP'` (makna **TIDAK DIKETAHUI**), `created_at`/`updated_at`.
  - **Indeks:** `(tanggal, status_ujian, event_id)`.
- **FK→ (DISIMPULKAN):** `event_id` → `events.id`; `created_by` → `users.id`.
- **Diacu oleh (DISIMPULKAN):** `jadwals_peserta.jadwal_id`, `sesi_schedules.jadwal_id`, `siswa_ujians.jadwal_id`, `jawaban_pesertas.jadwal_id`, `check_answers.jadwal_id`, `hasil_ujians.jadwal_id`.

#### 2.3.2 `jadwals_peserta` — peserta yang ditugaskan ke jadwal
- **Isi:** penugasan peserta → jadwal (join table murni). **Seed:** 0 baris.
- **PK:** `id` uuid. **Kolom:** hanya `jadwal_id` uuid, `peserta_id` uuid. **Tidak punya timestamp, tidak punya kolom status.**
- **FK→ (DISIMPULKAN):** `jadwal_id` → `jadwals.id`; `peserta_id` → `pesertas.id`.
- **Catatan TERBUKTI:** tidak ada `UNIQUE(jadwal_id, peserta_id)` — duplikasi penugasan dimungkinkan di level DB.

#### 2.3.3 `sesi_schedules` — sesi peserta per jadwal
- **Isi:** penugasan sesi (gelombang) peserta pada satu jadwal. **Seed:** 0 baris.
- **PK:** `id` uuid. **Kolom:** `jadwal_id` uuid, `sesi` int DEFAULT 1, `student_id` uuid, `created_at`/`updated_at`. **Indeks:** `(jadwal_id, sesi)`.
- **FK→ (DISIMPULKAN):** `jadwal_id` → `jadwals.id`; `student_id` → `pesertas.id`.
- **TIDAK DIKETAHUI:** beda fungsi dengan `jadwals_peserta`, `jadwals.sesi`, dan `pesertas.sesi` — dump memuat **empat** konsep "sesi" yang tumpang tindih (TERBUKTI ketiganya/keempatnya ada) tanpa dokumentasi pemisahannya. Menu `0021E–0021J` ("sesi peserta ujian": daftar/tambah/hapus/copy/upload) menunjukkan sesi peserta dikelola di bawah Event, bukan Jadwal — **DISIMPULKAN** `sesi_schedules` adalah penugasan sesi tingkat event/jadwal. Otoritas masing-masing kolom sesi **TIDAK DIKETAHUI**.

#### 2.3.4 `siswa_ujians` — attempt ujian peserta (tabel inti pelaksanaan)
- **Isi:** satu baris = satu percobaan mengerjakan satu peserta pada satu jadwal. **Seed:** 0 baris.
- **PK:** `id` uuid. **Kolom:**
  - Rujukan: `peserta_id` uuid, `jadwal_id` uuid, `banksoal_id` uuid NOT NULL (redundan dengan `jadwals.banksoal_id` — **DISIMPULKAN** denormalisasi).
  - ⏱ Waktu (string, TERBUKTI): `mulai_ujian` varchar(255) NOT NULL, `mulai_ujian_shadow` ❓ varchar(255) NOT NULL (**TIDAK DIKETAHUI** bedanya dengan `mulai_ujian`; hipotesis di Bab 7), `selesai_ujian` varchar(255) NULLABLE (**DISIMPULKAN** NULL = belum selesai).
  - ⚙ State machine: `status_ujian` int NOT NULL (nilai enum **TIDAK DIKETAHUI**), `sisa_waktu` int NOT NULL (**DISIMPULKAN** sisa waktu dalam detik — satuan tidak tertulis di skema; pola `TOKEN_EXPIRED`/`TRANSITION_TIME` yang memakai detik mendukung detik, tetap **DISIMPULKAN**), `out_ujian_counter` int DEFAULT 0 (**DISIMPULKAN** hitungan keluar ujian; berpasangan dengan `settings.MAX_COUNTER_OUT=1` + `AUTO_BLOCK=Y` — korelasi **DISIMPULKAN**, mekanisme persis **TIDAK DIKETAHUI**), `token` ❓ varchar(100) NULLABLE (**DISIMPULKAN** salinan token yang dipakai attempt ini; hubungannya dengan tabel `tokens` **TIDAK DIKETAHUI**).
  - ⚙ Part/transisi: `last_part` int DEFAULT 1 (**DISIMPULKAN** bagian terakhir yang dikerjakan), `transition_state` bool DEFAULT false + `transition_start` bigint DEFAULT 0 (**DISIMPULKAN** status & awal masa transisi antar-bagian; berpasangan dengan `jadwals.transition_time` / `settings.TRANSITION_TIME`).
  - `created_at` / `updated_at`. **Indeks:** `(peserta_id, jadwal_id, status_ujian)`.
- **FK→ (DISIMPULKAN):** `peserta_id` → `pesertas.id`; `jadwal_id` → `jadwals.id`; `banksoal_id` → `banksoals.id`.
- **Diacu oleh (DISIMPULKAN):** `hasil_ujians.ujian_id` (nama kolom `ujian_id` vs tabel `siswa_ujians` — pemetaan ini **DISIMPULKAN**, tidak ada kolom `siswa_ujian_id` yang eksplisit).
- **Catatan TERBUKTI:** tidak ada `UNIQUE(peserta_id, jadwal_id)` — percobaan ulang ganda dimungkinkan di level DB; tidak ada kolom perangkat di attempt (perangkat hanya di `devices`, terikat peserta, bukan attempt).

#### 2.3.5 `tokens` — token ujian (pool global)
- **Isi:** token ujian + status + kedaluwarsa. **Seed TERBUKTI:** 1 baris: `status = 0`, selisih `created_at → expired_at` tepat 15 menit (900 detik) — cocok dengan `settings.TOKEN_EXPIRED = '900'`. (Nilai tokennya tidak direproduksi.)
- **PK:** `id` uuid. **Kolom:** `token` varchar(255) NOT NULL (disimpan apa adanya — **TERBUKTI** bukan hash; bentuk seed: 4 karakter alfanumerik huruf besar), `status` int NOT NULL + **CHECK** `status::text IN ('0','1')` (satu-satunya CHECK di dump — **TERBUKTI**; pemetaan 0/1 ke arti **TIDAK DIKETAHUI**), `expired_at` ⏱ timestamp(0) NULLABLE, `created_at`/`updated_at`.
- **Catatan TERBUKTI:** tidak ada kolom `jadwal_id`/`banksoal_id`/`peserta_id` → token di tabel ini adalah **pool global**, bukan token per jadwal (berbeda dengan desain v2 di ANALISIS Bab 16 yang menaruh token di jadwal). Cara token mengikat ke jadwal tertentu **TIDAK DIKETAHUI** (kemungkinan lewat `siswa_ujians.token` — **DISIMPULKAN**).
- **Implikasi keamanan untuk v2 (DISIMPULKAN dari bentuk):** token pendek (4 karakter di seed) hanya aman bila dipasangi rate limit + masa berlaku singkat + revoke — ANALISIS Bab 7 sudah mencatat hal yang sama.

### 2.4 Jawaban, koreksi, hasil

#### 2.4.1 `jawaban_pesertas` — jawaban peserta per soal
- **Isi:** satu baris = jawaban satu peserta untuk satu soal dalam satu jadwal. **Seed:** 0 baris.
- **PK:** `id` uuid. **Kolom:**
  - Rujukan: `banksoal_id`, `soal_id`, `peserta_id`, `jadwal_id` (semua uuid; **catatan TERBUKTI:** tidak ada kolom `siswa_ujian_id`/attempt — kaitan ke attempt hanya lewat pasangan peserta+jadwal, **DISIMPULKAN**).
  - Isi jawaban: `jawab` text NULLABLE (**DISIMPULKAN** jawaban mentah peserta, format per tipe soal — format **TIDAK DIKETAHUI**), `answered` bool DEFAULT false (**DISIMPULKAN** sudah menjawab vs kosong), `paper_code` ❓ varchar(100) NULLABLE (**DISIMPULKAN** kode lembar kertas untuk varian ujian kertas; paralel `soals.paper_answer`).
  - ⚙ Penilaian: `checked` bool DEFAULT false (**DISIMPULKAN** sudah dinilai/dikoreksi), `point` numeric(19,2) DEFAULT 0 (poin yang diberikan), `iscorrect` bool DEFAULT false (**DISIMPULKAN** verdict benar/salah otomatis), `ai_opinion` ❓ text NULLABLE (**DISIMPULKAN** opini AI atas jawaban — PELAJARI menyebut opini AI di koreksi; isi/format **TIDAK DIKETAHUI**).
  - ⚙ Lain: `ragu_ragu` bool DEFAULT false (**TERBUKTI** ada flag ragu-ragu per jawaban), `part` int DEFAULT 1 (**DISIMPULKAN** bagian soal).
  - `created_at` / `updated_at` timestamp(6). **Indeks:** `(banksoal_id, soal_id, peserta_id, jadwal_id)`.
- **FK→ (DISIMPULKAN):** keempat kolom rujukan ke tabel masing-masing.
- **Catatan TERBUKTI:** tidak ada `UNIQUE(banksoal_id, soal_id, peserta_id, jadwal_id)` — jawaban ganda untuk satu soal dimungkinkan di level DB (autosave harus upsert di aplikasi, bukan dijamin DB).

#### 2.4.2 `check_answers` — koreksi manual jawaban
- **Isi:** jejak koreksi manual (penilaian manusia) per jawaban. **Seed:** 0 baris.
- **PK:** `id` uuid. **Kolom:** `banksoal_id`, `peserta_id`, `jawab_id` uuid (FK→ `jawaban_pesertas.id` **DISIMPULKAN** — nama kolomnya `jawab_id`, bukan `jawaban_id`), `corrected_by` uuid (FK→ `users.id` **DISIMPULKAN**), `point` numeric(19,2) NOT NULL (poin hasil koreksi), `jadwal_id` uuid, `ai_opinion` ❓ text NULLABLE (**DISIMPULKAN** opini AI yang dipakai/dicatat saat koreksi), `created_at`/`updated_at`. **Indeks:** `(banksoal_id, peserta_id, jawab_id)`.
- **Catatan TERBUKTI:** tidak ada kolom komentar koreksi / rubrik / status koreksi. Satu jawaban bisa punya banyak baris koreksi (tidak ada UNIQUE) — mana yang berlaku **TIDAK DIKETAHUI** (terakhir? agregat?).

#### 2.4.3 `hasil_ujians` — hasil agregat per peserta per jadwal
- **Isi:** agregat nilai satu peserta dalam satu jadwal + rincian per kode kompetensi. **Seed:** 0 baris.
- **PK:** `id` uuid. **Kolom:**
  - Rujukan: `ujian_id` ❓ uuid NULLABLE (target **DISIMPULKAN** `siswa_ujians.id`; nama tidak eksplisit — **TIDAK DIKETAHUI** pasti), `banksoal_id`, `peserta_id`, `jadwal_id` (uuid NOT NULL).
  - ⚙ Nilai total: `check_point` numeric(19,2) DEFAULT 0, `check_point_auto` numeric(19,2) DEFAULT 0 NULLABLE, `check_point_manual` numeric(19,2) DEFAULT 0 NOT NULL, `check_point_max` numeric(19,2) DEFAULT 0 (**DISIMPULKAN**: total / komponen otomatis / komponen manual / plafon).
  - ⚙ Per kode kompetensi (×16): `codeNN_correct` int, `codeNN_wrong` int, `codeNN_point` numeric(19,2), `codeNN_void` int — **DISIMPULKAN** benar/salah/poin/kosong per kode kompetensi (lihat §1.5: ini BUKAN per tipe soal). **Anomali TERBUKTI:** `code09_wrong` adalah satu-satunya kolom `codeXX_wrong` **tanpa** `NOT NULL`. Urutan kolom **TERBUKTI** janggal: `code15_*`/`code16_*` baru muncul setelah `check_point_max`/`duration` — tanda kolom ditambah belakangan.
  - ⏱ `duration` varchar(50) DEFAULT `'-'` (durasi pengerjaan sebagai string — format **TIDAK DIKETAHUI**).
  - `created_at` / `updated_at` timestamp(6). **Indeks:** `(banksoal_id, peserta_id, jadwal_id, ujian_id)`.
- **Catatan TERBUKTI:** tidak ada kolom status hasil (draf/final/publish), tidak ada kolom kelulusan, tidak ada `UNIQUE(peserta_id, jadwal_id)` — finalitas nilai **TIDAK DIKETAHUI** dari skema.

### 2.5 Analisis butir soal (agregat statistik)

#### 2.5.1 `banksoal_summaries` — ringkasan statistik per bank soal
- **Isi:** statistik agregat satu bank soal. **Seed:** 0 baris.
- **PK:** `id` uuid. **Kolom:** `banksoal_id` uuid, `total_done` int, `avg_duration` numeric(19,4), `total_schedule` int, `lowest_duration` / `highest_duration` int, `ds_duration` ❓ numeric(19,6) (**DISIMPULKAN** simpangan baku durasi — `ds` = deviasi standar, dari konteks kolom `avg_*`; **TIDAK DIKETAHUI** pasti), `avg_result` numeric(19,2), `ds_result` numeric(19,6), `avg_point` numeric(19,2), `ds_point` numeric(19,6), `lowest_result` / `highest_result` numeric(19,2), `max_point` numeric(19,2), `lowest_point` / `highest_point` numeric(19,2), `created_at` saja (tidak ada `updated_at` — **TERBUKTI**).
- **FK→ (DISIMPULKAN):** `banksoal_id` → `banksoals.id`.
- **TIDAK DIKETAHUI:** cakupan agregat (semua jadwal? per periode?), kapan dihitung ulang (menu `0018C Generate summary banksoal` membuktikan ada pemicu manual — **DISIMPULKAN** agregat dihitung atas perintah, bukan otomatis real-time).

#### 2.5.2 `soal_summaries` — analisis per butir soal
- **Isi:** statistik + psikometri per butir soal. **Seed:** 0 baris.
- **PK:** `id` uuid. **Kolom:** `banksoal_id`, `soal_id`, `total_void` / `total_correct` / `total_wrong` int, `highest_point` / `lowest_point` numeric(19,2), `avg_point` numeric(19,4), `ds_point` numeric(19,6), `addons` ❓ text NULLABLE (isi/format **TIDAK DIKETAHUI**), `created_at` saja (tanpa `updated_at`), `total_done` int DEFAULT 0, `discrimination` numeric(19,8) DEFAULT 0 (**DISIMPULKAN** daya beda soal), `distractor` ❓ text NULLABLE (**DISIMPULKAN** analisis pengecoh per opsi; format — JSON? teks? — **TIDAK DIKETAHUI**).
- **FK→ (DISIMPULKAN):** `banksoal_id` → `banksoals.id`; `soal_id` → `soals.id`.
- **Catatan TERBUKTI:** tidak ada kolom `jadwal_id` → cakupan agregat (per bank? lintas jadwal?) **TIDAK DIKETAHUI**. Tidak ada kolom tingkat kesukaran eksplisit — **DISIMPULKAN** diturunkan dari `total_correct / total_done`.

### 2.6 Event, ruangan, kartu peserta

#### 2.6.1 `events` — event ujian (gelombang)
- **Isi:** header event ujian. **Seed:** 0 baris.
- **PK:** `id` uuid. **Kolom:** hanya `name` varchar(255) + `created_at`/`updated_at`.
- **Catatan TERBUKTI:** tidak ada kolom tanggal/status/lokasi/deskripsi → event di sini murni label pengelompok. Menu `0021A–0021N` (tambah/edit/hapus event, sesi peserta, kartu peserta, ruangan) membuktikan event adalah wadah operasional (**DISIMPULKAN**).
- **Diacu oleh (DISIMPULKAN):** `jadwals.event_id`, `room_members.event_id`.

#### 2.6.2 `rooms` — ruangan
- **Isi:** daftar ruangan. **Seed:** 0 baris.
- **PK:** `id` uuid. **Kolom:** hanya `code` varchar(255), `name` varchar(255). **Tidak punya timestamp.**
- **Catatan TERBUKTI:** tidak ada kapasitas/lokasi/status → ruangan murni label kode+nama.
- **Diacu oleh (DISIMPULKAN):** `room_members.room_id`.

#### 2.6.3 `room_members` — anggota ruangan per event
- **Isi:** penempatan peserta ke ruangan dalam satu event. **Seed:** 0 baris.
- **PK:** `id` uuid. **Kolom:** `room_id`, `event_id`, `student_id` (uuid), `created_at` timestamp(6) saja (tanpa `updated_at`).
- **FK→ (DISIMPULKAN):** `room_id` → `rooms.id`; `event_id` → `events.id`; `student_id` → `pesertas.id`.
- **Catatan TERBUKTI:** tidak ada `UNIQUE(room_id, event_id, student_id)` — satu peserta bisa tercatat dua kali di ruangan/event yang sama di level DB.
- **Kartu peserta:** **TERBUKTI** tidak ada tabel kartu peserta. Menu `0021F 'Kartu peserta event ujian'` membuktikan fitur kartu ada di aplikasi; datanya **DISIMPULKAN** dirangkai dari `pesertas` + `events` + `room_members` (+ `no_ujian`) saat cetak. Format/QR/nomor kartu **TIDAK DIKETAHUI**.

### 2.7 Keamanan, perangkat, sistem

#### 2.7.1 `devices` — perangkat peserta terverifikasi
- **Isi:** catatan perangkat peserta. **Seed TERBUKTI:** 1 baris: `peserta_id` NULL, `out_count = 0`, `last_verified_at` NULL. (Nilai `device_id`/`device_os`/`device_name`/`ip_address` tidak direproduksi — `ip_address` seed kosong.)
- **PK:** `id` uuid. **Kolom:** `device_id` varchar(255), `device_os` varchar(255), `device_name` varchar(255), `ip_address` varchar(255), `created_at`, `peserta_id` uuid NULLABLE, `out_count` int NOT NULL (**DISIMPULKAN** hitungan keluar ujian pada perangkat ini; berpasangan dengan `siswa_ujians.out_ujian_counter` + `settings.MAX_COUNTER_OUT` — apakah keduanya dihitung terpisah atau salah satu turunan yang lain **TIDAK DIKETAHUI**), `last_verified_at` ⏱ timestamp(6) NULLABLE (**DISIMPULKAN** terakhir verifikasi perangkat; menu `0024A 'Verified device client'` mendukung bacaan ini).
- **FK→ (DISIMPULKAN):** `peserta_id` → `pesertas.id`.
- **Catatan TERBUKTI:** `peserta_id` NULLABLE + seed 1 baris tanpa peserta → perangkat bisa tercatat sebelum terikat peserta (atau baris seed sekadar contoh). Cara fingerprint perangkat dibuat **TIDAK DIKETAHUI** dari skema (tidak ada kolom user-agent/hash).

#### 2.7.2 `settings` — konfigurasi key-value
- **Isi:** seluruh konfigurasi aplikasi sebagai pasangan kunci → nilai teks. **Seed TERBUKTI:** 19 baris (daftar kode di bawah; nilai rahasia tidak direproduksi).
- **PK + UNIQUE:** `code` varchar(255). **Kolom:** `code`, `value` text NOT NULL. **Tidak punya timestamp / pengubah.**
- **Kode TERBUKTI ada (19):** `EXAM_THEME`, `EXAM_TEXT_WELCOME_PARSEDXX88`, `EXAM_TEXT_FINISH_PARSEDXX88`, `TRANSITION_TIME`, `TOKEN_EXPIRED`, `SECRET_EXO_BROWSER`, `BROWSER_ALLOW_TYPE`, `BROWSER_AGENT`, `LOGIN_ONCE_ATIME`, `SECRET_EXO_LOGOUT`, `SECRET_EXO_LOGOUT_ACTIVE`, `VERIFIED_DEVICE`, `AUTO_BLOCK`, `MAX_COUNTER_OUT`, `INSTANT_LOGO_1712LGX`, `INSTANT_KOP_1712LGX`, `INSTANT_NAME`, `INSTANT_CITY`, `INSTANT_ADDRESS`.
- **Nilai TERBUKTI (non-rahasia):** `TOKEN_EXPIRED = '900'` (detik — **DISIMPULKAN** satuannya detik, dikuatkan interval 900 detik pada seed `tokens`), `TRANSITION_TIME = '60'`, `LOGIN_ONCE_ATIME = 'Y'`, `AUTO_BLOCK = 'Y'`, `MAX_COUNTER_OUT = '1'`, `VERIFIED_DEVICE = 'N'`, `EXAM_THEME = '1'`, `EXAM_TEXT_*` kosong, `BROWSER_ALLOW_TYPE` berisi teks mirip JSON array, `INSTANT_*` berisi identitas sekolah + logo/kop (nilai persis tidak direproduksi).
- **Nilai DISEMBUNYIKAN (rahasia, tidak direproduksi):** `SECRET_EXO_BROWSER`, `SECRET_EXO_LOGOUT`, `SECRET_EXO_LOGOUT_ACTIVE`, `BROWSER_AGENT`.
- **DISIMPULKAN:** semua nilai bertipe teks — interpretasi (int/bool/Y-N/JSON) ada di kode aplikasi, bukan di skema. Tidak ada audit perubahan settings (tanpa `updated_at`/`updated_by`) — siapa mengubah settings **TIDAK DIKETAHUI** dari dump.

#### 2.7.3 `menus` — katalog menu / permission
- **Isi:** daftar aksi (permission) aplikasi. **Seed TERBUKTI:** 99 baris, kode `0001A … 0025C`, tiap baris = satu aksi (mis. `0010H 'Tambah soal'`, `0011E 'Pause jadwal ujian'`, `0011F 'Resume jadwal ujian'`, `0014C 'Force finish ujian peserta'`, `0019A 'Daftar ledger peserta'`, `0023A/B 'Backup/Restore master system'`).
- **PK:** `code` varchar(255). **Kolom:** `title` text, `"order"` int (**DISIMPULKAN** urutan tampil; nilai 11000–36001, naik per kelompok kode).
- **Catatan TERBUKTI:** tidak ada kolom parent/grup — pengelompokan hanya dari awalan kode (`0001` = peserta, `0010` = bank soal, `0011` = jadwal, `0012` = token, `0021` = event, dst. — **DISIMPULKAN**).

#### 2.7.4 `role_menus` — pemetaan role → permission
- **Isi:** role mana boleh melakukan aksi mana. **Seed TERBUKTI:** 170 baris — `EVIL` 99 baris (= seluruh menu), `EVAL` 32 baris, `OPRA` 39 baris.
- **PK:** **TIDAK ADA (TERBUKTI).** **Kolom:** hanya `role_code` varchar(255), `menu_code` varchar(255). **Tidak ada FK, tidak ada timestamp.**
- **FK→ (DISIMPULKAN):** `menu_code` → `menus.code`; `role_code` adalah string bebas (nilai yang terbukti: `EVIL`/`EVAL`/`OPRA`, sama dengan domain `users.role` — **DISIMPULKAN**).
- **Contoh isi TERBUKTI dari seed (pola, bukan daftar lengkap):** `EVIL` memegang semua aksi termasuk backup/restore, settings, pause/resume; `OPRA` memegang jadwal/token/monitor/event/ruangan tetapi **tidak** memegang bank soal; `EVAL` memegang bank soal/kompetensi/koreksi/hasil tetapi **tidak** memegang pause/resume jadwal. Rincian lengkap ada di seed dump; ringkasan matriks untuk v2 sebaiknya diturunkan dari ANALISIS Bab 11, bukan dari daftar mentah ini.
- **Catatan TERBUKTI:** tanpa PK, duplikat pasangan (role, menu) dimungkinkan di level DB.

#### 2.7.5 `personal_access_tokens` — token akses personal (polimorfik)
- **Isi:** token akses API per subjek. **Seed TERBUKTI:** 1 baris, `tokenable_type = 'User'`. (Nilai token tidak direproduksi.)
- **PK:** `id` bigint (sequence). **UNIQUE:** `token`. **Kolom:** `tokenable_type` varchar(255), `tokenable_id` uuid, `name` varchar(255), `token` varchar(64), `created_at`/`updated_at`. **Indeks:** `(tokenable_type, tokenable_id)`.
- **DISIMPULKAN:** pola polimorfik ala token akses framework web (tipe + ID subjek). Hubungan dengan `pesertas.api_token` dan `tokens` **TIDAK DIKETAHUI**.
- **Catatan TERBUKTI:** tidak ada kolom `expired_at`, `last_used_at`, atau `revoked` → masa berlaku/pencabutan **TIDAK DIKETAHUI**.

#### 2.7.6 `backups` — riwayat backup sistem
- **Isi:** catatan backup. **Seed:** 0 baris.
- **PK:** `id` bigint (sequence `exo_backups_id_seq`). **Kolom:** `filename`, `version`, `detail` text, `generated_date` ❓ varchar(255) (string, bukan timestamp — **TERBUKTI**; format **TIDAK DIKETAHUI**), `bak_type` varchar(255) DEFAULT `'BACKUP'`, `status` ❓ varchar(255) DEFAULT `'SUCCESS'` (teks bebas, tanpa CHECK — nilai lain **TIDAK DIKETAHUI**), `created_at`/`updated_at`.
- **Catatan:** menu `0023A/B` (backup/restore master system) + `0010N/O` (backup/restore bank soal) membuktikan dua cakupan backup (**DISIMPULKAN**); tabel ini hanya mencatat backup tingkat sistem — **DISIMPULKAN** dari nama + `bak_type`. Lokasi file backup / isi `detail` **TIDAK DIKETAHUI**.

#### 2.7.7 `ecosystem_keys` — kunci lisensi offline
- **Isi:** kunci lisensi ekosistem (satu baris = satu lisensi). **Seed TERBUKTI:** 2 baris. (Payload/signature tidak direproduksi.)
- **PK:** `id` varchar(100). **Kolom:** `payload` text, `public_key` text, `signature` text. **Tidak punya timestamp.**
- **TERBUKTI dari bentuk seed:** `payload` adalah teks JSON yang memuat identitas sekolah + masa berlaku (`expired_at` sebagai epoch); `public_key` kosong pada kedua baris seed; `signature` terisi (opaque).
- **Untuk v2: JANGAN DITIRU.** Mekanisme lisensi Exo tidak diadopsi (keputusan ANALISIS). Kolom ini dicatat hanya agar tim paham bahwa lisensi disimpan di DB, bukan di file config — dan agar tidak ada yang mencoba memakai/membuka kunci seed.

### 2.8 Utilitas & subsistem lain

#### 2.8.1 `medias` — manajer berkas/media (pohon folder)
- **Isi:** node berkas/folder generik. **Seed TERBUKTI:** 1 baris folder generik akar.
- **PK:** `id` uuid. **Kolom:** `name` varchar(255), `is_directory` bool DEFAULT false, `parent_id` uuid NULLABLE (self-reference **DISIMPULKAN**), `created_at`/`updated_at`.
- **Catatan TERBUKTI:** tidak ada kolom path/ukuran/mime/soal_id → cara media menempel ke soal **TIDAK DIKETAHUI** (kemungkinan lewat kolom `supports` — **DISIMPULKAN**, keyakinan rendah).

#### 2.8.2 `feature_infos` — tutorial in-app (audio)
- **Isi:** metadata tutorial per fitur. **Seed TERBUKTI:** 20 baris, kode `AA001 … AI001`, `content` berupa path `tutorial-in-app/XX###.mp3`.
- **PK:** `id` bigint (sequence). **Kolom:** `name`, `code`, `content` text, `created_at`/`updated_at`. **Indeks:** `(name)`.
- **DISIMPULKAN:** tiap baris = satu klip audio penjelasan fitur (judul tutorialnya dipakai sebagai bukti pendukung di beberapa inferensi laporan ini, mis. `AG005`, `AH003`, `AB002`).

#### 2.8.3 `prompts` — ❓ (tabel paling tidak pasti)
- **Isi:** **TIDAK DIKETAHUI.** **Seed:** 0 baris.
- **PK:** `id` char(26). **Kolom:** `user_id` varchar(36), `side` varchar(100), `secondary_id` varchar(36) NULLABLE, `content` text, `hide` bool NOT NULL, `created_at` timestamp(6) NOT NULL, `session_id` char(26) NOT NULL, `location_id` varchar(100) DEFAULT `''`.
- **Hipotesis (DISIMPULKAN, keyakinan rendah):** pola kolom (`side`, `session_id`, `content`, `hide`) menyerupai log percakapan/chat atau antrean prompt AI. Bisa juga prompt UI. **Tidak ada bukti yang cukup untuk memilih.** Lihat Bab 7 — jangan jadikan dasar desain.
- **Catatan TERBUKTI:** `user_id` varchar(36) tidak cocok tipe dengan `users.id` uuid; `session_id` char(26) sekeluarga dengan ID `tools`/`shortcuts`.

#### 2.8.4 `tools` — ❓ registry tool/fungsi (diduga untuk AI)
- **Isi:** definisi tool/fungsi yang bisa dipanggil. **Seed TERBUKTI:** 4 baris: `knowledge_base` (type `system`), `generate_question`, `detail_student`, `total_student` (type `function`).
- **PK:** `id` char(26). **Kolom:** `type` varchar(100) DEFAULT `'function'`, `name` varchar(100), `description` text NULLABLE, `parameters` text NOT NULL, `script` text NULLABLE, `required` text NULLABLE, `location` varchar(100) DEFAULT `''`. **Tidak punya timestamp.**
- **TERBUKTI:** isi `description`/`parameters`/`script`/`required` pada seed berupa blob teks terenkode (opaque) — **tidak direproduksi dan tidak didekode** dalam tugas ini (di luar ruang lingkup analisis skema; mendekodekannya tidak diperlukan untuk memetakan skema dan berisiko menyentuh konten non-skema).
- **DISIMPULKAN (keyakinan sedang, dari nama):** registry tool untuk asisten/AI (`generate_question` = generate soal; `detail_student`/`total_student` = kueri data siswa; `knowledge_base` = basis pengetahuan). Perilaku, cara eksekusi `script`, dan arti `location`/`required` **TIDAK DIKETAHUI**.

#### 2.8.5 `shortcuts` — pintasan per pengguna
- **Isi:** pintasan (shortcut) milik pengguna. **Seed:** 0 baris.
- **PK:** `id` char(26). **Kolom:** `user_id` char(36), `name` varchar(255), `destination` text, `icon` text NULLABLE. **Tidak punya timestamp.**
- **FK→ (DISIMPULKAN):** `user_id` → `users.id` (tipe char(36) vs uuid — ketidakcocokan **TERBUKTI**).
- **DISIMPULKAN** dari menu `0025A–C` (daftar/tambah/hapus shortcut): pintasan navigasi/aksi cepat per pengguna.
