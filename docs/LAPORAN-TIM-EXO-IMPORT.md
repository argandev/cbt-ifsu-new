# LAPORAN TIM EXO IMPORT — Mekanisme Import Soal Extraordinary CBT

**Tim:** TIM EXO IMPORT (15 orang, kerja mandiri — tanpa subagent)
**Tanggal:** 16 Oktober 2026
**Ruang lingkup:** HANYA mekanisme import soal. Arsitektur umum, role/permission (kecuali izin import itu sendiri), timer, anti-curang, grading, lisensi, dll. TIDAK dibahas.
**Metode:** Binary TIDAK PERNAH dijalankan. Analisis hanya dari: dokumen dalam paket unduhan resmi (README, template di `storage/exo-templates/`, `.env` — hanya nama key, bukan nilai), skema `exo-dump-master.sql` (dibaca saja, tidak dijalankan ke DB mana pun), situs resmi ekstraordinary.com (homepage, halaman download, blog), panduan pihak ketiga (DewaHoster), repo issue resmi developer (shellrean-dev), dan dokumen template komunitas sebagai koroborasi sekunder. Tidak ada login, crack, decompile, atau kredensial yang dipakai.

## Legenda label bukti

- **TERBUKTI** — ada di dokumen/skema/situs resmi yang dikutip; sumber disebut.
- **DISIMPULKAN** — inferensi logis dari bukti yang disebut; alasannya ditulis.
- **TIDAK DIKETAHUI** — tidak ditemukan buktinya di sumber yang diperiksa; jangan dianggap ada.

---

## 0. RINGKASAN EKSEKUTIF

1. **Klaim lama "import Google Docs / Word / CSV" perlu dikoreksi.** Yang TERBUKTI: import **Word (.docx)** dalam beberapa varian format + **HTML Zip** + **Excel (.xlsx, terbatas)**. **Google Docs TERBUKTI hanya sebagai klaim marketing + konfigurasi env** — prosedur impornya TIDAK DIKETAHUI. **CSV untuk soal TIDAK ADA BUKTINYA** (template CSV/XLSX yang ada hanya untuk peserta/matpel/user/grup/sesi/ruangan, bukan soal).
2. **Format utama adalah "Format 2"**: dokumen Word berisi **satu tabel 2 kolom per soal** dengan baris marker `TS / KD / KJ / ABS` + nomor soal + opsi. Format ini mendokumentasikan **9 kode tipe soal**.
3. **"16 tipe soal" tidak terbukti sebagai 16 parser import.** Bukti terkuat: komentar template kompetensi resmi mencantumkan **13 kode TS**; template Format 2 hanya mencontohkan **9**; 3 sisanya TIDAK DIKETAHUI. Diduga angka 16 mencampur tipe soal dengan slot kompetensi `code01–code16` di `banksoals`.
4. **Tidak ada field pembahasan, bobot per soal, level, atau stimulus per soal** di template import maupun di tabel `soals`. Bobot/poin hidup di `competence_items`, stimulus/lampiran hidup di level bank soal ("banksoal support").
5. **Tidak ditemukan bukti** adanya pratinjau pra-simpan, pesan error import, batas jumlah soal, atau penanganan duplikat. Yang ada adalah tombol **Preview Bank Soal sesudah import** (verifikasi pasca-import).

---

## 1. Format yang diterima

| # | Format | Status bukti | Sumber |
|---|---|---|---|
| 1 | Word `.docx` **Format 1** — gaya baris `S:` / `A:)` / `JAWAB:` | TERBUKTI (template bawaan paket) | `storage/exo-templates/format-input-soal-doc.docx` |
| 2 | Word `.docx` **Format 2** — tabel 2 kolom per soal, marker `TS/KD/KJ/ABS` | TERBUKTI (template bawaan paket + panduan DewaHoster menyebut "format tipe file 2") | `storage/exo-templates/format-input-soal-doc-2.docx`; DewaHoster KB 141 |
| 3 | **HTML Zip Format 2** — hasil Save As HTML dari Format 2 + folder gambar, di-zip | TERBUKTI (panduan DewaHoster: pilihan dialog "[HTML Zip] Format 2") | DewaHoster KB 141 (pihak ketiga, bukan situs resmi) |
| 4 | Excel `.xlsx` — 8 kolom, kode tipe numerik | TERBUKTI (template bawaan paket), tapi hanya PG + Esai yang dicontohkan | `storage/exo-templates/format-input-soal-exc.xlsx` |
| 5 | Word `.docx` gaya `##...##` + pemisah `***` (PG) | TERBUKTI sebagai file template bawaan; pemakaian di dialog import TIDAK DIKETAHUI | `storage/exo-templates/format-input-soal-pg.docx` |
| 6 | Word `.docx` gaya pemisah `***` (Esai) | TERBUKTI sebagai file template bawaan; pemakaian di dialog import TIDAK DIKETAHUI | `storage/exo-templates/format-input-soal-esay.docx` |
| 7 | **Google Docs** | TERBUKTI sebagai klaim + konfigurasi; mekanisme TIDAK DIKETAHUI | Homepage ekstraordinary.com; key `.env` `GOOGLE_DOC_ACTIVE`, `GOOGLE_CREDENTIAL` (nama key saja) |
| 8 | **CSV** (.csv) untuk soal | TIDAK DIKETAHUI / tidak ada bukti | Tidak ada template soal berekstensi .csv di paket; template .xlsx yang ada untuk master data, bukan soal |
| 9 | PDF / TXT / lainnya untuk soal | TIDAK DIKETAHUI / tidak ada bukti | — |

**Catatan penamaan versi:** Release note resmi v3.17.0 (repo `shellrean-dev/exo-cbt-client`) menyebut perbaikan urutan saat import **"docx 2, docx 1 and html"** — TERBUKTI ada tiga jalur import dokumen bernama docx 1, docx 2, dan html. Ini cocok dengan Format 1, Format 2, dan HTML Zip Format 2 di atas. Status jalur Excel dan `##/***` terhadap dialog import TIDAK DIKETAHUI (bisa jadi contoh/legacy, bukan pilihan dialog).

**Koreksi klaim lama (dari `PELAJARI-EXTRAORDINARY-CBT.md`):**

- "Import soal: Google Docs / Word (.docx), termasuk rumus" → **Word TERBUKTI; Google Docs hanya klaim; "termasuk rumus" TIDAK TERBUKTI** (lihat §7). Satu-satunya bukti rumus adalah judul video di README ("Cara import soal dengan rumus") — isi videonya tidak diverifikasi dalam kajian ini.
- Klaim "import … / CSV" di kajian lama → **tidak ada bukti untuk soal**; jangan dipertahankan.

**Akses menu import:** Permission `0010K` "Import banksoal" di `role_menus` hanya dimiliki role **EVAL** dan **EVIL** (OPRA tidak punya) — TERBUKTI dari dump. Artinya import adalah kewenangan guru/super admin, bukan proktor.

---

## 2. Struktur format / template

### 2.1 Template tersedia di mana?

- **TERBUKTI:** Paket unduhan resmi menyertakan folder `storage/exo-templates/` berisi file-file template (lihat daftar §1). Template soal yang ada: `format-input-soal-doc.docx`, `format-input-soal-doc-2.docx`, `format-input-soal-pg.docx`, `format-input-soal-esay.docx`, `format-input-soal-exc.xlsx`, plus `format-template-competence.xlsx` (untuk pemetaan kompetensi).
- **TERBUKTI (pihak ketiga):** Dari dialog Import Soal, pengguna bisa mengunduh contoh penulisan ("Download format tipe file 2") — DewaHoster KB 141.
- **TIDAK DIKETAHUI:** Apakah template bisa diunduh langsung dari situs web publik (tanpa login ke aplikasi). Tidak ditemukan tautannya di ekstraordinary.com.

### 2.2 Format 2 — tabel per soal (format utama, paling lengkap)

Struktur hasil ekstraksi `format-input-soal-doc-2.docx` (10 tabel contoh):

- Setiap soal = **satu tabel 2 kolom**.
- **4 baris pertama SELALU** (urutan tetap): `TS` | `KD` | `KJ` | `ABS`.
- Baris berikutnya: **nomor soal** (kolom 1, mis. `1.`) = teks pertanyaan; lalu **baris opsi** (kolom 1 = label opsi, kolom 2 = teks opsi).
- Di atas tabel-tabel contoh, template memuat blok "Perhatian" berisi: larangan membuat tabel selain format Exo; larangan menyimpan floating image; daftar 9 kode TS; dan satu paragraf penjelasan per tipe soal.
- **DISIMPULKAN:** Parser membaca per tabel secara berurutan; urutan tabel di dokumen = urutan soal. Alasan: release note v3.17.0 menyebut perbaikan "order when import soal docx 2, docx 1 and html" — urutan import adalah perilaku yang diperhatikan dan pernah bermasalah.

### 2.3 Format 1 — gaya baris

Struktur hasil ekstraksi `format-input-soal-doc.docx` (6 blok contoh):

- Satu soal = blok paragraf: `S:<nomor>) <teks soal>` → baris-baris opsi `A:) <teks>` … `E:) <teks>` → `JAWAB: <kunci>`.
- Tidak ada marker TS/KD/KJ/ABS; tidak ada tabel.
- Varian kunci yang tercermin di contoh: kunci tunggal (`JAWAB: E`) dan kunci ganda (`JAWAB: B,C`).
- Satu contoh soal tidak memiliki kunci (blok terakhir hanya berisi pertanyaan) — **DISIMPULKAN** contoh itu tidak lengkap / ilustrasi soal tanpa kunci (mis. esai), bukan bukti aturan.
- Format ini juga dipakai di dokumen-dokumen soal beredar di web (judulnya menyebut "FORMAT CBT CANDY ATAU CBT ekstraordinary") — koroborasi sekunder bahwa Format 1 adalah format lama yang kompatibel dengan ekosistem CBT Candy/Exo.

### 2.4 Gaya `##...##` + `***` (PG) dan `***` (Esai)

- `format-input-soal-pg.docx`: teks soal, lalu tiap opsi diawali `##`; kunci ditandai dengan menulis label jawaban di dalam `##...##` pada baris kunci (mis. baris `##A##` sebelum teks opsi A = A benar; pada contoh kedua, kunci D ditulis sebagai `##D##` sebelum opsi A — posisi marker kunci tidak terikat opsi). Blok soal dipisah `***`.
- `format-input-soal-esay.docx`: hanya teks soal per blok, dipisah `***`.
- **TIDAK DIKETAHUI:** Apakah parser import saat ini masih menerima gaya ini, atau file-file ini sekadar contoh lama. Tidak ada rujukan ke `##`/`***` di panduan DewaHoster, changelog, maupun release note yang diperiksa.

### 2.5 Excel `.xlsx`

- Satu sheet, tabel `A1:H51` (header + 50 baris slot).
- Header (8 kolom): `SOAL | Tipe Soal | Kunci Jawaban | Opsi 1 | Opsi 2 | Opsi 3 | Opsi 4 | Opsi 5`.
- Hanya **3 baris contoh terisi** (2 PG + 1 esai), sisanya kosong.
- Kode tipe di contoh memakai **angka** (`1` untuk PG, `2` untuk esai) — bukan kode huruf TS.
- **DISIMPULKAN:** Template Excel dirancang terutama untuk PG + Esai; tipe lain (menjodohkan, mengurutkan, dsb.) tidak dicontohkan dan pemetaannya TIDAK DIKETAHUI.

### 2.6 Kompetensi (file pendamping)

`format-template-competence.xlsx` — struktur:

- Baris atas: `Total part` + `Kode Kompetensi` + `Nama`.
- Tabel item (header): `Kode | Tipe soal | Total soal | Point | Judul | Deskripsi | Part | Point dijawab salah`.
- Contoh item memakai `Kode 1.1.0`, `Tipe soal PG`, `Point 10`, `Point dijawab salah 0`.
- Komentar sel pada file ini (ditulis "developer") mencantumkan daftar kode TS — lihat §9.
- **DISIMPULKAN:** File ini adalah cara guru mendefinisikan pemetaan kompetensi (bobot/poin per tipe soal per kompetensi) yang kemudian dirujuk dari `KD` di Format 2. Alasan: kolom-kolomnya cocok 1:1 dengan kolom `competence_items` di skema (`code`, `tipe_soal`, `total_soal`, `point`, `name`, `description`, `part`, `wrong_point`) dan `competences` (`code`, `name`, `total_part`).

---

## 3. Pemetaan kolom → field soal

### 3.1 Format 2 (tabel)

| Marker / baris | Arti (TERBUKTI dari template) | Tujuan di DB (DISIMPULKAN, kecuali disebut) |
|---|---|---|
| `TS` | Kode tipe soal (PG, PGX, ESY, SKT, JD, JDX, URT, STJ, BNR) | `soals.tipe_soal` (integer; angka pemetaannya TIDAK DIKETAHUI — dump tidak berisi satu pun baris soal) |
| `KD` | Kode kompetensi, mis. `1.0.1`; boleh kosong | `soals.competence_item_id` via `competence_items.code`. **DISIMPULKAN** dari kecocokan format kode (`1.0.1` / `1.1.0`) dengan contoh file kompetensi |
| `KJ` | Kunci jawaban; formatnya bergantung tipe (lihat §3.3) | `jawaban_soals.correct` / `text_jawaban` — **DISIMPULKAN**: opsi yang ditandai kunci disimpan sebagai baris `jawaban_soals` dengan `correct = true` |
| `ABS` | Flag; berisi `Y` atau kosong di contoh | **DISIMPULKAN** = `soals.absolute_eval` (boolean; `Y` = true). Alasan: nama kolom cocok, dan nilai `Y`/kosong cocok dengan boolean. **Belum TERBUKTI** karena tidak ada dokumentasi yang mendefinisikan ABS |
| Nomor soal (`1.`) | Teks pertanyaan | `soals.pertanyaan` |
| Baris opsi (`A`, `B`, … / `A1`, `A2` … / `Y`, `N`) | Teks opsi / pasangan / pernyataan | `jawaban_soals.text_jawaban` (+ `label_mark` — **DISIMPULKAN** untuk label seperti `A1`, `Y`) |
| — | Urutan tabel dalam dokumen | `soals.sequence` — **DISIMPULKAN** (lihat §2.2) |

**Anomali ABS yang wajib dicatat:** Pada dokumen template komunitas yang beredar (Scribd), ada contoh blok soal yang menulis huruf jawaban (mis. `ABS C`) pada posisi yang di template resmi ditempati `KJ`. Ini **tidak konsisten** dengan template resmi (di sana huruf kunci ada di `KJ`, `ABS` kosong/`Y`). **DISIMPULKAN:** Kemungkinan besar itu artefak ekstraksi teks (baris `KJ` yang kosong lalu `ABS` tergabung dengan huruf dari baris lain), BUKAN aturan bahwa ABS menampung kunci. **Jangan mengimplementasikan "ABS = kunci"** tanpa verifikasi di aplikasi asli. Arti ABS tetap berstatus DISIMPULKAN (bukan TERBUKTI).

### 3.2 Format 1

| Elemen | Arti | Tujuan (DISIMPULKAN) |
|---|---|---|
| `S:<n>)` | Teks soal | `soals.pertanyaan` |
| `A:)` … `E:)` | Opsi | `jawaban_soals.text_jawaban` |
| `JAWAB: <huruf>` / `JAWAB: B,C` | Kunci tunggal / ganda | `jawaban_soals.correct` |

Tidak ada pemetaan KD/kompetensi, ABS, pembahasan, atau bobot di Format 1.

### 3.3 Excel

| Kolom | Arti | Tujuan (DISIMPULKAN) |
|---|---|---|
| `SOAL` | Teks pertanyaan | `soals.pertanyaan` |
| `Tipe Soal` | Angka (`1` = PG, `2` = esai di contoh) | `soals.tipe_soal` |
| `Kunci Jawaban` | Huruf kunci (mis. `A`) | `jawaban_soals.correct` |
| `Opsi 1`–`Opsi 5` | Teks opsi | `jawaban_soals.text_jawaban` |

### 3.4 Semantik `KJ` per tipe (TERBUKTI dari penjelasan di template Format 2)

| Tipe | Isi `KJ` | Catatan template |
|---|---|---|
| PG | Satu huruf opsi (mis. `E`) | KJ harus tersedia pada opsi |
| PGX | Daftar huruf dipisah koma (mis. `A,B,C,D`) | Peserta menjawab 1 atau lebih jawaban |
| ESY | Teks jawaban rujukan (panjang) | **Tidak** mengoreksi otomatis; hanya rujukan pengoreksi; boleh dikosongkan (di dokumen komunitas; di template resmi contohnya terisi teks rujukan) |
| SKT | Teks jawaban eksak (mis. `1944`) | Jawaban peserta harus sama persis dengan KJ |
| JD | Pasangan (mis. `A1-A2, B1-B2, C1-D1`) | `A1` = sisi asal/domain, `A2` = sisi tujuan/kodomain |
| JDX | Pasangan (mis. `A1-A2, B1-B2, C1-C2`) | Seperti JD tetapi ada opsi blank (opsi tanpa daerah asal) |
| URT | Kosong | "Tidak perlu menuliskan KJ karena opsinya adalah KJ itu sendiri"; opsi ditampilkan acak ke peserta |
| STJ | Kosong | Dikoreksi manual; peserta menulis pendapat atas argumen |
| BNR | Kosong | Kunci melekat pada opsi: tiap pernyataan diawali `Y` (benar) atau `N` (salah) |

### 3.5 Field yang TIDAK ADA di import (TERBUKTI dari skema + template)

| Field di DB | Bukti ketiadaan di import |
|---|---|
| `soals` **tidak punya** kolom pembahasan / bobot per soal / level kesulitan | `CREATE TABLE soals` hanya berisi: `id, banksoal_id, tipe_soal, pertanyaan, layout, competence_item_id, absolute_eval, case_sensitive, supports, sequence, paper_answer`. **Tidak ada kolom pembahasan, bobot, atau level.** |
| Bobot / poin | Hidup di `competence_items.point` + `wrong_point`, dan plafon di `banksoals.code01_max … code16_max`. Diatur lewat file kompetensi (§2.6), bukan per soal di template |
| `soals.case_sensitive` | Ada di skema, tapi **tidak ada marker-nya** di template mana pun → cara mengaturnya lewat import TIDAK DIKETAHUI (nilai default di skema: `true`) |
| `soals.layout` | Ada di skema, tapi **tidak ada marker-nya** di template → TIDAK DIKETAHUI |
| Stimulus per soal | **Tidak ada** baris stimulus di template. Stimulus/lampiran ada di level bank soal: `banksoals.supports`, `soals.supports` (kolom teks), dan menu `0010L` "Tambah banksoal support" / `0010M` "Hapus banksoal support". **DISIMPULKAN:** stimulus di-attach ke bank soal, bukan ditulis di dalam file import soal |

---

## 4. Validasi

**TERBUKTI dari template Format 2 (aturan tertulis di dalam file):**

1. Dilarang membuat tabel selain format Extraordinary CBT.
2. Dilarang menyimpan floating image (gambar harus inline/anchored).
3. `TS` wajib diisi (dengan salah satu kode yang didaftar).
4. Untuk PG: `KJ` harus tersedia pada opsi (kunci harus merujuk ke opsi yang ada).

**TERBUKTI dari penjelasan tipe:**

5. SKT: jawaban peserta harus sama persis dengan `KJ` (implikasi: `KJ` kosong / tidak cocok = soal tidak bisa dinilai benar secara otomatis).

**TERBUKTI dari riwayat bug (repo issue resmi developer):**

6. Import ZIP pernah gagal merekam jawaban pengecoh pada soal menjodohkan ke banyak jawaban (JDX) — dilaporkan sebagai bug, dinyatakan selesai di versi 4.1.3-rosetta. Ini bukti bahwa **pengecoh (distractor) adalah data yang harus terekam** pada JDX, dan pernah hilang saat import ZIP.
7. Import `.docx` pernah gagal memasukkan gambar ke filemedia — diperbaiki di 2.0.0-rc.2.
8. Urutan soal hasil import (docx 2, docx 1, html) pernah salah — diperbaiki di v3.17.0.

**TIDAK DIKETAHUI:**

- Teks pesan error import yang sesungguhnya (tidak ditemukan satu pun contoh pesan error di sumber yang diperiksa).
- Perilaku saat validasi gagal: apakah seluruh file ditolak (all-or-nothing), baris bermasalah dilewati, atau soal gagal disimpan sebagian. Tidak ada bukti.
- Apakah ada validasi jumlah opsi minimum, duplikasi label opsi, atau `KD` yang tidak terdaftar di kompetensi.
- Apakah `KJ` yang tidak cocok dengan opsi ditolak atau disimpan begitu saja.

---

## 5. Preview / konfirmasi sebelum simpan

- **TERBUKTI (pihak ketiga):** Setelah upload berhasil, pengguna bisa memeriksa tampilan soal lewat tombol **"Preview Bank Soal"** — DewaHoster KB 141. Ini adalah **pratinjau pasca-import**, bukan konfirmasi pra-simpan.
- **TERBUKTI (riwayat):** Release note v3.17.0 menyebut perbaikan "preview soal menjodohkan in admin" — pratinjau soal di admin memang ada sebagai fitur.
- **TIDAK DIKETAHUI:** Apakah dialog Import menampilkan pratinjau / ringkasan / konfirmasi ("N soal akan disimpan") sebelum commit. Tidak ada bukti langkah konfirmasi di sumber yang diperiksa. **DISIMPULKAN:** Alur import tampaknya langsung (pilih format → Upload → tersimpan), dengan verifikasi dilakukan sesudahnya lewat Preview Bank Soal. Verifikasi langsung di aplikasi diperlukan untuk memastikan.

---

## 6. Batas jumlah soal per import

- **TIDAK DIKETAHUI.** Tidak ada angka batas maksimum yang ditemukan di situs resmi, template, changelog, panduan, maupun skema.
- **Jangan keliru:** Template Excel memakai tabel 50 baris slot (header + 50). Itu adalah ukuran template contoh, **BUKAN bukti** batas engine 50 soal. Template Word Format 2 berisi 10 soal contoh — juga bukan batas.
- **DISIMPULKAN lemah:** Karena tidak ada batas yang didokumentasikan, batas praktis (jika ada) kemungkinan berasal dari ukuran file / timeout upload, bukan dari jumlah soal — tetapi ini inferensi tanpa bukti langsung.

---

## 7. Penanganan gambar / media dalam soal yang di-import

**TERBUKTI dari template:**

| Jalur | Cara gambar dibawa | Bukti |
|---|---|---|
| Format 1 (`.docx`) | Gambar inline di dalam dokumen: satu gambar stimulus sebelum teks pertanyaan (contoh S:2), dan satu gambar per opsi (contoh S:3 — lima opsi bergambar) | Ekstraksi `format-input-soal-doc.docx`: 6 file media, drawing menempel pada paragraf stimulus dan pada tiap baris opsi |
| Format 2 (`.docx`) | Gambar inline di dalam sel tabel soal (contoh: gambar di dalam sel pertanyaan PG soal ke-2) | Ekstraksi `format-input-soal-doc-2.docx`: 1 PNG di sel pertanyaan; file juga mengandung 1 SVG yang penempatannya tidak terlacak dari ekstraksi teks |
| Gaya `##` (PG) | Gambar inline pada paragraf kosong yang disisipkan di antara teks soal dan opsi | Ekstraksi `format-input-soal-pg.docx`: 4 PNG pada paragraf kosong |
| HTML Zip Format 2 | File HTML + folder gambar di-zip; ZIP diunggah manual lewat file manager ke folder storage (`…/public/storage/public/<id-soal>/…`), diekstrak, lalu dipilih di dialog import | DewaHoster KB 141 (pihak ketiga) |
| Larangan | Floating image dilarang (template Format 2) | Template Format 2 |

**Audio / video dalam import:**

- **Tidak ada bukti** audio/video bisa dibawa lewat file import soal. Template tidak mencontohkannya.
- **Koroborasi sekunder (dokumen soal komunitas):** Untuk soal berisi suara/video, penulis diminta menulis deskripsi lalu **mengedit kembali di menu Daftar Soal** dan mengunggah audio (mp3) / video (mp4) lalu menyisipkannya. **DISIMPULKAN:** Media audio/video adalah langkah pasca-import manual, bukan bagian dari file import.
- Skema mendukung lampiran di level bank soal (`banksoals.supports`, `soals.supports`, menu banksoal support) — **DISIMPULKAN** sebagai jalur lampiran stimulus.

**Rumus / equation:**

- **TIDAK DIKETAHUI** bagaimana rumus Word (Equation Editor / OMML) ditangani parser. Template tidak berisi satu pun contoh rumus; satu-satunya contoh aritmetika di Format 1 ditulis sebagai teks biasa.
- **DISIMPULKAN lemah:** Pada jalur HTML Zip, rumus Word yang di-Save-As-HTML umumnya berubah menjadi gambar, sehingga rumus kemungkinan ikut sebagai gambar — tetapi ini inferensi dari perilaku Word, bukan bukti perilaku Exo. Judul video resmi "Cara import soal dengan rumus" belum diverifikasi isinya.

---

## 8. Penanganan duplikat

- **TERBUKTI dari skema:** Tabel `soals` hanya memiliki `PRIMARY KEY (id)`; **tidak ada constraint UNIQUE** pada teks pertanyaan atau kombinasi apa pun. (Untuk perbandingan: `pesertas.no_ujian` dan `settings.code` punya UNIQUE — `soals` tidak.)
- **TIDAK DIKETAHUI:** Apakah aplikasi memeriksa duplikat saat import (mis. membandingkan teks pertanyaan yang sudah ada di bank soal yang sama).
- **DISIMPULKAN:** Tanpa constraint UNIQUE dan tanpa bukti pemeriksaan di aplikasi, import ulang file yang sama kemungkinan besar membuat soal ganda (duplikat baris baru dengan id baru). Ini inferensi dari skema, bukan perilaku yang teramati. Rekomendasi untuk v2: implementasikan deteksi duplikat eksplisit (lihat §11).

---

## 9. Pemetaan 16 kode soal dari format import

### 9.1 Klaim "16"

- **TERBUKTI:** Homepage + blog resmi mengklaim "16 Tipe Soal" (blog menyebut variannya: pilihan ganda tunggal & kompleks, benar-salah, menjodohkan, isian singkat, esai, "dan masih banyak lagi" — tanpa daftar lengkap).
- **TERBUKTI:** `banksoals` memiliki 16 kolom plafon `code01_max … code16_max`.

### 9.2 Kode yang TERBUKTI bisa ditulis di format import

| Kode | Nama | Format 2 (tabel) | Format 1 | Excel | Komentar kompetensi |
|---|---|---|---|---|---|
| PG | Pilihan ganda | ✅ contoh | ✅ (`JAWAB:`) | ✅ (`Tipe 1`) | ✅ |
| PGX | Pilihan ganda kompleks | ✅ contoh | ✅ (`JAWAB: B,C`) | ❌ | ✅ |
| ESY | Esai / uraian | ✅ contoh | ❓ (blok tanpa kunci, tidak berlabel) | ✅ (`Tipe 2`) | ✅ |
| SKT | Isian singkat | ✅ contoh | ❌ | ❌ | ✅ |
| JD | Menjodohkan 1-1 | ✅ contoh | ❌ | ❌ | ✅ |
| JDX | Menjodohkan 1-X | ✅ contoh | ❌ | ❌ | ✅ |
| URT | Mengurutkan | ✅ contoh | ❌ | ❌ | ✅ |
| STJ | Setuju / tidak setuju | ✅ contoh | ❌ | ❌ | ✅ |
| BNR | Benar / salah | ✅ contoh | ❌ | ❌ | ✅ |
| MCK | Matrix checkbox | ❌ | ❌ | ❌ | ✅ (hanya tercantum di komentar file kompetensi) |
| MFT | Matrix free text | ❌ | ❌ | ❌ | ✅ (idem) |
| GRP | Mengelompokkan | ❌ | ❌ | ❌ | ✅ (idem) |
| BNG | Blank space | ❌ | ❌ | ❌ | ✅ (idem) |

### 9.3 Yang TIDAK DIKETAHUI

- **3 kode sisanya** dari klaim "16" tidak ditemukan di sumber mana pun yang diperiksa. Kemungkinan (DISIMPULKAN, belum terbukti): angka 16 menghitung varian turunan (mis. PG tunggal vs PG kompleks sebagai dua, JD vs JDX sebagai dua, dst.), atau ada tipe yang hanya bisa dibuat lewat input manual / AI generate, atau `code01–code16` adalah slot kompetensi (per kompetensi per tipe), bukan 16 parser.
- **MCK / MFT / GRP / BNG:** tercantum di komentar file kompetensi resmi, tetapi **tidak ada contoh penulisannya** di template import mana pun → cara meng-import keempat tipe ini (atau keharusan input manual) TIDAK DIKETAHUI.
- **Pemetaan kode TS → `soals.tipe_soal` (integer):** TIDAK DIKETAHUI. Satu-satunya petunjuk numerik adalah Excel (`1` = PG, `2` = esai), yang belum tentu sama dengan enum internal.
- **Pernyataan jujur untuk v2:** Yang bisa diadopsi dengan percaya diri sebagai "format import Exo" adalah **9 tipe Format 2** + **PG/PGX Format 1** + **PG/Esai Excel**. Sisanya adalah klaim yang belum terverifikasi — jangan menulis di dokumen v2 bahwa "Exo mendukung import 16 tipe".

---

## 10. Panduan pengisian yang disediakan Exo untuk guru

| Bentuk panduan | Status | Keterangan |
|---|---|---|
| Instruksi di dalam template Format 2 (blok Perhatian + penjelasan per tipe) | TERBUKTI | Paling lengkap & paling konkret; mencakup larangan, daftar TS, dan semantik KJ per tipe |
| Komentar "developer" di file template kompetensi (daftar 13 kode TS) | TERBUKTI | Singkat; hanya daftar kode |
| Header kolom template Excel | TERBUKTI | 8 kolom berlabel jelas; contoh minimal |
| Video resmi di README: "Cara import soal" + "Cara import soal dengan rumus" + "Cara input soal" | TERBUKTI keberadaan tautannya; isi TIDAK DIVERIFIKASI | Hanya judul + URL di README; video tidak ditonton dalam kajian ini |
| Tutorial audio in-app (`feature_infos` AH001–AH004 bank soal, AG001–AG005 kompetensi) | TERBUKTI keberadaannya (file .mp3 di `storage/tutorial-in-app/`) | Isi audio tidak ditranskripsikan dalam kajian ini |
| Panduan instalasi / blog resmi | TIDAK ADA panduan import | Blog resmi yang diperiksa hanya memuat intro, panduan per jenjang, update, troubleshooting login/app — tidak ada artikel import soal |
| Panduan pihak ketiga (DewaHoster) | TERBUKTI (bukan resmi) | Satu-satunya panduan langkah-demi-langkah import (khusus HTML Zip Format 2) yang ditemukan |

**Penilaian (DISIMPULKAN):** Panduan pengisian Exo bersifat **tersebar dan bertingkat**: instruksi terbaik justru tertanam di dalam file template itu sendiri, ditambah video + tutorial audio. Tidak ada satu halaman manual import resmi yang lengkap dan terpusat. Guru yang hanya membaca situs web tidak akan menemukan spesifikasi format import — ia harus membuka template dari dalam aplikasi/paket.

---

## 11. Implikasi & rekomendasi untuk CBT SMK v2

1. **Adopsi Format 2 sebagai format import utama v2** (tabel 2 kolom per soal, marker TS/KD/KJ/ABS) — format ini yang paling terdokumentasi dan paling ekspresif (9 tipe). Definisikan ABS secara eksplisit di dokumen v2 (kami putuskan: `ABS = Y` → penilaian absolut / semua-benar-atau-salah; kosong = proporsional) — jangan meniru ambiguitas Exo.
2. **Adopsi Format 1 sebagai format legacy** (S:/JAWAB:) untuk kompatibilitas dengan naskah soal lama sekolah — hanya untuk PG/PGX.
3. **Adopsi Excel 8 kolom** untuk PG + Esai sebagai jalur cepat guru yang terbiasa spreadsheet.
4. **Lampaui Exo** pada empat titik yang di Exo TIDAK DIKETAHUI / tidak ada buktinya: (a) pratinjau pra-simpan dengan ringkasan (jumlah soal per tipe, daftar soal yang akan dibuat); (b) laporan error per baris dengan pesan yang jelas + mode tolak-semua vs lewati-baris-gagal yang dipilih pengguna; (c) deteksi duplikat eksplisit (teks pertanyaan normalisasi + opsi) dengan pilihan lewati / timpa / simpan-ganda; (d) batas yang didokumentasikan (ukuran file, jumlah soal per file).
5. **Pembahasan, bobot per soal, level kesulitan, dan stimulus per soal** tidak ada di Exo — v2 sudah memutuskan mendukungnya (lihat KONTRAK v2); jangan menghilangkannya dengan alasan "mengikuti Exo".
6. **Jangan klaim "16 tipe import"** di dokumen v2. Tulis jujur: "9 tipe import Format-2 terverifikasi dari template Exo + PG/PGX Format-1 + PG/Esai Excel; tipe matriks/pengelompokan/blank-space tercantum di dokumen kompetensi Exo tanpa contoh penulisan import."
7. **Rumus:** v2 perlu keputusan eksplisit (MathJax/KaTeX inline `$...$` sudah diputuskan di audit UI/UX) karena Exo tidak memberikan spesifikasi rumus yang dapat diverifikasi.

---

## 12. Daftar pertanyaan yang MASIH TERBUKA (perlu verifikasi langsung di aplikasi Exo)

1. Teks pesan error import dan perilaku gagal (tolak semua vs lewati baris).
2. Ada/tidak pratinjau & konfirmasi pra-simpan; isi dialog import (pilihan format apa saja yang tampil).
3. Batas jumlah soal / ukuran file per import.
4. Penanganan duplikat (apakah ada pemeriksaan di aplikasi).
5. Pemetaan kode TS → `tipe_soal` integer; 3 tipe "16" yang belum teridentifikasi; cara import MCK/MFT/GRP/BNG.
6. Arti pasti `ABS` dan cara mengatur `case_sensitive` / `layout` lewat import.
7. Mekanisme import Google Docs (apakah via API dengan service account, format dokumen apa yang dibaca, apakah memakai marker Format 1/2 yang sama).
8. Cara rumus Word ditangani parser (teks / gambar / gagal).
9. Apakah `##/***` dan Excel masih diterima parser import versi 4.7.0, atau hanya contoh lama.
10. Cara kerja "banksoal support" (stimulus) dan hubungannya dengan `soals.supports`.

---

## Sumber

**Dokumen paket resmi (di `~/workspace/extraordinary-cbt/`):**

- `app/README.txt` — tautan video "Cara import soal", "Cara import soal dengan rumus", "Cara input soal".
- `app/exo-dump-master.sql` — skema `soals`, `jawaban_soals`, `banksoals` (`code01_max…code16_max`, `supports`), `competence_items`, `competences`, `menus` (`0010K` Import banksoal, `0010L/M` banksoal support), `role_menus` (EVAL+EVIL untuk `0010K`), `feature_infos` (tutorial audio AH/AG), `settings`; dump tidak berisi baris soal (0 baris `soals`/`jawaban_soals`/`banksoals`).
- `app/storage/exo-templates/` — `format-input-soal-doc.docx`, `format-input-soal-doc-2.docx`, `format-input-soal-pg.docx`, `format-input-soal-esay.docx`, `format-input-soal-exc.xlsx`, `format-template-competence.xlsx` (diekstrak struktur XML-nya; isi teks contoh tidak disalin ke laporan ini).
- `app/.env` — hanya nama key `GOOGLE_DOC_ACTIVE`, `GOOGLE_CREDENTIAL` (nilai tidak dibaca/dilaporkan).
- `download.html` — snapshot halaman download resmi; changelog v3.17.0 (import menjodohkan/mengurutkan/benar-salah/setuju-tidak via docx Format 2 & Format 2 HTML; perbaikan urutan import) dan v4.7.0.

**Situs resmi:**

- https://ekstraordinary.com — klaim "16 Tipe Soal & Impor Google Docs" (Word/Google Docs "tanpa pusing memformat ulang").
- https://ekstraordinary.com/blog — tidak ditemukan artikel import soal.
- https://ekstraordinary.com/blog/pengenalan/intro — daftar varian tipe soal (tanpa daftar lengkap 16).

**Pihak ketiga / komunitas (koroborasi, bukan sumber primer):**

- DewaHoster KB 141, "Cara Upload Bank Soal dengan Tipe Dokumen HTML ke Aplikasi Extraordinary CBT" — https://www.dewahoster.com/knowledgebase/141/Cara-Upload-Bank-Soal-dengan-Tipe-Dokumen-HTML-ke-Aplikasi-Extraordinary-CBT.html — prosedur HTML Zip Format 2 + tombol Preview Bank Soal.
- GitHub `shellrean-dev/exo-cbt-client` release v3.17.0 — jalur import "docx 2, docx 1 and html"; perbaikan preview menjodohkan.
- GitHub `shellrean-dev/exo-cbt-4-rosetta-issue` issue #1 — bug pengecoh JDX pada import ZIP (selesai di 4.1.3-rosetta); `shellrean-dev/exo-cbt-service` 2.0.0-rc.2 — perbaikan gambar docx tidak masuk filemedia.
- Dokumen template komunitas berlabel "Extraordinary CBT Template by shellrean" (Scribd) — dipakai hanya sebagai koroborasi struktur Format 2; isinya konsisten dengan template resmi `format-input-soal-doc-2.docx`.
- Dokumen soal komunitas berformat `S:/JAWAB:` (assyifakarangsari.sch.id) — koroborasi Format 1.

**Yang sengaja TIDAK dilakukan:** menjalankan binary `main-amd64`/`main-arm64`; menjalankan dump ke database; mengakses akun/ekosistem; menonton video tutorial (hanya judul/URL dari README yang dikutip); menyalin isi soal contoh dari template.
