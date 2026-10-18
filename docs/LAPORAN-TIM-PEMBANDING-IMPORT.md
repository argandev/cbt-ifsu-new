# LAPORAN TIM PEMBANDING IMPORT — Mekanisme Import Soal Aplikasi Lain vs CBT SMK v2

**Tanggal:** 16 Oktober 2026
**Penyusun:** Tim Pembanding Import (kerja mandiri — tanpa subagent)
**Status:** HANYA ANALISIS. Tidak ada kode yang diubah.
**Bahan awal yang diperdalam:** `docs/LAPORAN-TIM-PEMBANDING-LUAR.md` (laporan pembanding sebelumnya) — laporan ini TIDAK mengulang dimensi non-import; ia memperdalam dimensi import yang di sana masih berlabel TIDAK DIKETAHUI (Quizizz, Kahoot, Exam.net) dan menambahkan detail sintaks / kolom / validasi / preview / media untuk semua pembanding.

## 0. Cara membaca laporan ini (aturan bukti)

Setiap klaim memakai salah satu label:

- **TERBUKTI** — ada di dokumentasi / halaman resmi / dokumen publik yang disebut sumbernya (nama + URL).
- **DISIMPULKAN** — penalaran yang masuk akal dari bukti yang dikutip, tetapi tidak dinyatakan eksplisit oleh sumber.
- **TIDAK DIKETAHUI** — tidak ditemukan bukti publik yang cukup dalam riset ini. Tidak ditebak.

Aturan tambahan yang dipatuhi:

- Dokumen ini boleh MENYEBUT nama aplikasi lain untuk perbandingan faktual. Tidak ada aset, logo, branding, atau teks berhak cipta pihak lain yang disalin ke sini. Contoh sintaks yang ditampilkan adalah parafrase/ringkasan format yang didokumentasikan publik, bukan salinan halaman vendor.
- Kolom CBT SMK v2 diisi dari kode yang benar-benar ada di repo (`backend/internal/banksoal/banksoal.go`, `frontend/src/pages/admin/banksoal/BankSoalDetailPage.tsx`), BUKAN dari spesifikasi ANALISIS.md semata. Bila spesifikasi dan kode berbeda, yang dipakai adalah kode — dan perbedaannya dicatat eksplisit (lihat Bab 1 dan Bab 10).
- Untuk Extraordinary CBT (Exo), kolom import diisi HANYA dari kajian `~/workspace/extraordinary-cbt/PELAJARI-EXTRAORDINARY-CBT.md` + ANALISIS.md Bab 13, seperti pada laporan sebelumnya — artinya "terdapat di menu/skema menurut kajian", bukan "terverifikasi jalan live".

### Keterbatasan riset ini

- Riset memakai pencarian web publik + pembacaan halaman resmi. Beberapa halaman resmi hanya terbaca sebagian lewat alat baca teks (mis. gambar contoh file ProProfs hanya terbaca sebagai keterangan gambar, isi sel contoh tidak terbaca) — bagian yang tidak terbaca ditulis TIDAK DIKETAHUI, bukan ditebak dari gambar.
- Beberapa detail (mis. pesan error persis per baris di Quizizz/Wayground, batas jumlah soal di Moodle) tidak dinyatakan eksplisit di dokumentasi resmi yang dibaca — ditulis TIDAK DIKETAHUI.
- Harga / paket tidak dibahas di laporan ini (di luar fokus import), kecuali bila batas import terikat paket (mis. ProProfs Free maks 8 soal) karena itu adalah batas import, bukan harga.

---

## 1. Pembanding internal: import soal CBT SMK v2 saat ini (baseline yang akan diganti)

Sumber: kode di repo (bukan spesifikasi). Endpoint `POST /api/v1/bank-soal/import`, parser `parseCSVImport` di `backend/internal/banksoal/banksoal.go`, UI di `frontend/src/pages/admin/banksoal/BankSoalDetailPage.tsx`.

| Dimensi | CBT SMK v2 (kode yang ada) |
|---|---|
| Format file | **TERBUKTI (kode):** CSV teks yang dikirim sebagai JSON (`csv_text` / `csv` + `bank_id` + `mode`). Frontend menerima file `.csv` lalu membaca isinya sebagai teks. Bukan upload file biner, bukan `.xlsx`, bukan `.docx`. |
| Template yang disediakan | **TERBUKTI (kode UI):** TIDAK ada tombol unduh template. Yang ada hanya (a) satu baris teks nama kolom di atas panel import, (b) placeholder textarea berisi header + 1 baris contoh PG. Guru harus mengetik/menyalin header sendiri. |
| Panduan pengisian | **TERBUKTI (kode UI):** Satu kalimat ("Format kunci per tipe: lihat docs/TIPE-SOAL.md") + placeholder 1 baris. Tidak ada panduan per tipe di dalam UI import; tidak ada contoh baris untuk 15+ tipe lain. |
| Pemetaan kolom → field soal | **TERBUKTI (kode):** Header CSV dipetakan by-name (case-insensitive, trim): wajib `tipe`, `pertanyaan`, `kunci`; opsional `opsi_a`…`opsi_e`, `bobot` (default 1, bulat > 0), `level` (default `sedang`), `stimulus`, `pembahasan`. Satu baris = satu soal. Tidak ada kolom untuk kisi-kisi/kompetensi, kategori, nama soal, feedback per opsi, atau media. |
| Format kunci per tipe (pemetaan paling rawan salah) | **TERBUKTI (kode `buildCSVSoal`):** tiap tipe punya dialek kunci sendiri — PG: satu huruf (`B`); PG Kompleks/PG Multi: huruf dipisah koma (`A,C`); Benar-Salah: nilai `B/S` (juga diterima `benar/salah`, `true/false`, `1/0`) dipisah koma, jumlahnya harus sama dengan jumlah pernyataan; Menjodohkan/Peta Konsep: pasangan `kiri::kanan`, kunci pasangan dipisah `;`, sementara opsi diisi `kiri::kanan` juga; Isian: alternatif dipisah `\|`; Urutkan: label urutan benar dipisah koma; Isi Ganda / Teks Rumpang: jawaban per kolom/rumpang dipisah `\|` (rumpang: jumlah `___` di pertanyaan harus sama dengan jumlah jawaban kunci); Pernyataan Sikap: (kode lanjutan, format kunci diatur `ValidateSoal`). Tiga pemisah berbeda (`,`, `;`, `\|`) + satu pemisah pasangan (`::`) dalam satu format file. |
| Validasi + pesan error | **TERBUKTI (kode):** Validasi per baris: tipe dikenal, pertanyaan terisi, kunci terisi, bobot bulat > 0, level valid, aturan per tipe (mis. minimal 2 opsi, kunci cocok dengan label opsi yang terisi, pasangan berformat `kiri::kanan`, jumlah rumpang = jumlah kunci), plus `ValidateSoal`. Setiap error dikembalikan sebagai `{baris, pesan}` dengan nomor baris file (header = baris 1, data mulai baris 2). Error level file: CSV tak terbaca, tanpa header + 1 baris, kolom wajib hilang, > 500 baris. Header tidak sensitif huruf besar/kecil. Baris kosong dilewati (tidak dihitung error). |
| Mode gagal | **TERBUKTI (kode):** Dua mode eksplisit. `all_or_nothing` (default): bila ada error apa pun → `sukses: 0`, `disimpan: false`, tidak ada soal tersimpan. `skip_errors`: baris valid disimpan, baris gagal dilaporkan. Mode tidak valid → 400. Bank terkunci → 409, import ditolak. |
| Preview sebelum simpan | **TERBUKTI (kode):** TIDAK ada preview per soal sebelum simpan. Alur: kirim → server validasi → langsung simpan (mode `skip_errors`) atau tolak semua (mode `all_or_nothing`). Yang ditampilkan sesudahnya hanya ringkasan `Sukses / Gagal / Disimpan` + daftar error per baris. Guru TIDAK melihat soal hasil parsing (teks opsi, kunci terpetakan) sebelum soal masuk bank. Satu-satunya "pratinjau" adalah mode `all_or_nothing` yang berfungsi sebagai dry-run validasi, bukan pratinjau isi. |
| Batas jumlah | **TERBUKTI (kode):** Maks 500 baris data per berkas (dihitung sebelum validasi per baris; bila > 500, seluruh file ditolak dengan satu error `baris: 0`). Tidak ada batas ukuran byte eksplisit di parser. |
| Penanganan gambar/media | **TERBUKTI (kode):** TIDAK ADA. `INSERT` import selalu mengisi `media_url` dengan string kosong (`""`), tidak ada kolom CSV untuk media, tidak ada upload file pendamping, tidak ada kolom URL gambar. Soal bergambar tidak bisa diimpor; gambar harus ditambahkan lewat jalur lain (edit soal / upload gambar per soal, bila tersedia) setelah import. |
| Status hasil import | **TERBUKTI (kode):** Soal hasil import berstatus `draft`. |
| Keamanan / otorisasi | **TERBUKTI (kode):** Membutuhkan permission `banksoal.import`; proktor dilarang mutasi soal; guru dibatasi scope banknya (`checkGuruScope`); setiap import dicatat di audit log (sukses/gagal/mode). |

**DISIMPULKAN — apa yang bikin guru gampang vs bingung di import v2 saat ini:**

- Gampang: error bernomor baris + mode `all_or_nothing` (aman: gagal = tidak ada yang setengah masuk); header tidak sensitif kapital; baris kosong dilewati; bobot/level punya default.
- Bingung (alasan konkret): (1) **Tidak ada file template unduhan** — guru harus menyalin header dari kalimat UI / placeholder; satu huruf header salah ketik = kolom dianggap hilang/tidak terbaca. (2) **Tiga pemisah kunci + satu pemisah pasangan** (`,`, `;`, `\|`, `::`) yang berbeda per tipe, tanpa contoh per tipe di UI — risiko salah pemisah tinggi, dan pesan error baru muncul setelah kirim. (3) **Tidak ada preview isi** — guru tidak bisa memastikan kunci terpetakan benar sebelum soal tersimpan (mis. kunci `B` yang ternyata memetakan ke opsi yang salah urutan). (4) **Tidak ada kolom media** — soal bergambar harus dikerjakan dua tahap. (5) **Tidak ada kolom kisi-kisi/kompetensi/kategori** — pemetaan ke kurikulum harus dilakukan manual per soal setelah import, sehingga import massal belum benar-benar selesai dalam satu langkah.

**PERBEDAAN PENTING — spesifikasi vs kode (TERBUKTI dari kode):** ANALISIS.md / laporan sebelumnya menyebut v2 juga punya "Import Word (.docx): 1 dokumen = 1 bank; kontrak format ketat". Di kode yang ada, **TIDAK ditemukan endpoint import .docx** — yang diimplementasikan hanya CSV. Maka dalam laporan ini, baseline v2 = CSV saja; klaim .docx diperlakukan sebagai rencana yang belum terwujud, bukan fitur yang ada.

---

## 2. Moodle (GIFT, Aiken, Moodle XML, + format lain)

Sumber utama resmi: MoodleDocs — GIFT format (https://docs.moodle.org/34/en/GIFT_format), Aiken format (https://docs.moodle.org/36/en/Aiken_format), Moodle XML format (https://docs.moodle.org/39/en/Moodle_XML_format). Sumber pendukung: panduan import institusi (Clarkson, Ulethbridge, Bethel, Portsmouth) dan artikel perbandingan format (arXiv 2010.15577 — artikel akademik, bukan dokumen vendor).

### 2.1 Gambaran umum jalur import

- **TERBUKTI (MoodleDocs + panduan institusi):** Import dilakukan dari Question Bank → Import, dengan memilih format file, kategori tujuan, lalu unggah file. Soal masuk ke bank soal (kategori), bukan langsung ke kuis — kuis kemudian mengambil soal dari bank. Ini memisahkan "impor soal" dari "susun ujian", mirip v2 (bank → paket → jadwal).
- **TERBUKTI (MoodleDocs GIFT):** GIFT juga tersedia sebagai format ekspor dari Question Bank, dan dapat dipakai untuk bulk loading ke kategori atau ke Lesson.
- **DISIMPULKAN:** Karena soal hasil import mendarat di bank + dapat diedit per soal setelah import, kesalahan import masih bisa diperbaiki satu per satu tanpa mengulang seluruh file — pola yang sama dengan v2 (soal import berstatus draft).

### 2.2 Format 1 — GIFT (General Import Format Technology)

- **Format file — TERBUKTI (MoodleDocs):** File teks biasa (`.txt`), encoding **wajib UTF-8** (ANSI hanya aman bila tanpa karakter khusus; UTF-16/“Unicode” tidak bisa dipakai). Satu soal dipisahkan oleh **minimal satu baris kosong**.
- **Template yang disediakan — TERBUKTI (MoodleDocs):** Tidak ada file template unduhan resmi di halaman GIFT; yang disediakan adalah halaman dokumentasi dengan contoh-contoh soal per tipe + tautan ke "GIFT examples" yang dapat diimpor/dijadikan template kasar. Guru menyalin contoh dari dokumentasi, bukan mengunduh `.txt` siap isi.
- **Panduan pengisian — TERBUKTI (MoodleDocs):** Sangat lengkap: tabel simbol + penjelasan tiap simbol + contoh per tipe soal (PG, PG multi-jawaban benar, Benar-Salah, Isian Singkat, Menjodohkan, Kata Hilang, Numerik, Esai, Deskripsi) + opsi lanjutan (komentar, nama soal, feedback, bobot persen).
- **Pemetaan (sintaks → field soal) — TERBUKTI (MoodleDocs), diringkas dalam bahasa sendiri:**
  - Judul soal opsional diapit `::` (mis. `::Judul::`); bila tidak diisi, seluruh teks soal dipakai sebagai judul.
  - Jawaban diapit `{ }`; `=` = benar, `~` = salah, `#` = feedback untuk jawaban sebelumnya, `####` = feedback umum, `%n%` = bobot persen per jawaban.
  - Benar-Salah: `{T}` / `{F}` (atau `{TRUE}` / `{FALSE}`).
  - Isian Singkat: semua jawaban berawalan `=` dianggap benar (alternatif jawaban benar).
  - Menjodohkan: pasangan ditulis `=kiri -> kanan`; **minimal 3 pasang**; tidak mendukung feedback/bobot persen.
  - Kata Hilang (missing word): opsi ditulis di tengah kalimat; garis isian disisipkan otomatis.
  - Numerik: diawali `{#`, toleransi ditulis setelah `:` (mis. jawaban 2 toleransi 0,5), atau rentang `min..max`; beberapa jawaban numerik dengan bobot persen didukung lewat `=` per jawaban.
  - Esai: `{}` kosong (tidak ada apa pun di antara kurung kurawal).
  - Deskripsi: soal tanpa bagian jawaban sama sekali.
  - Komentar baris diawali `//` tidak diimpor. Format teks per potongan dapat dinyatakan (`[html]`, `[moodle]`, `[plain]`, `[markdown]`). Karakter khusus (`~`, `=`, `#`, `{`, `}`, `:`) harus di-escape dengan backslash (`\`); baris baru di dalam soal ditulis `\n` karena baris kosong adalah pemisah soal.
- **Validasi + pesan error — TIDAK DIKETAHUI secara rinci:** MoodleDocs GIFT tidak merinci daftar pesan error per baris dalam halaman yang dibaca. Panduan institusi (Clarkson) menyebut alur "review for errors … click Continue karena semua soal sudah ter-parse; bila tidak, perbaiki file teks dan coba lagi" — DISIMPULKAN: ada tahap parse + tampil error sebelum Continue, tetapi format pesan error persisnya (apakah bernomor baris/soal) TIDAK DIKETAHUI dari sumber yang dibaca.
- **Preview sebelum simpan — DISIMPULKAN (dari panduan Clarkson):** Ada tahap peninjauan hasil parse sebelum Continue — tetapi apakah preview menampilkan isi soal per soal (teks + kunci) atau hanya hitungan soal ter-parse, TIDAK DIKETAHUI dari sumber yang dibaca. Jangan samakan dengan preview isi penuh.
- **Batas jumlah — TIDAK DIKETAHUI:** Tidak ditemukan batas jumlah soal per file di halaman GIFT yang dibaca. Jangan klaim ada batas 500 seperti v2.
- **Gambar/media — TERBUKTI (MoodleDocs GIFT + artikel akademik + plugin pihak ketiga):** GIFT inti (teks) TIDAK membawa file gambar. Ada dua jalur yang didokumentasikan: (a) menulis tag `<img>` dengan referensi `@@PLUGINFILE@@/...` dan mengimpor sebagai arsip ZIP berisi file teks + folder media (disebut "GIFT with media" — implementasi ini tersedia sebagai plugin format impor pihak ketiga, bukan format inti GIFT); (b) artikel akademik menyebut cara yang sama (folder gambar + tag `@@PLUGINFILE@@` + ZIP). **DISIMPULKAN:** Untuk guru, media di GIFT adalah jalur lanjutan yang rawan salah (path case-sensitive, escape `=` di HTML, struktur ZIP) — bukan jalur default.
- **Gampang vs bingung (DISIMPULKAN dari desain format):**
  - Gampang: satu file teks untuk banyak tipe soal sekaligus; bobot persen + feedback per opsi dalam satu tempat; komentar `//` untuk anotasi; ekspor-impor simetris (GIFT ↔ bank soal) sehingga guru bisa "melihat" format yang benar dengan mengekspor soal contoh.
  - Bingung: (1) **karakter khusus bermakna ganda** — `=`, `~`, `#`, `{`, `}` adalah sintaks, sehingga soal yang secara alami mengandung karakter itu (mis. rumus `1+1=2`, tanda `#`) wajib di-escape; lupa escape = soal rusak diam-diam atau parse gagal. (2) **Baris kosong adalah pemisah** — soal multi-paragraf/stimulus panjang tidak bisa ditulis dengan baris kosong natural; harus memakai `\n` atau pindah format. (3) **Satu kesalahan sintaks dapat menggagalkan satu blok soal** tanpa penanda visual di editor teks (tidak ada pewarnaan sintaks bawaan di Notepad). (4) **Tanpa template unduhan resmi**, guru pemula harus merakit file dari contoh dokumentasi.

### 2.3 Format 2 — Aiken

- **Format file — TERBUKTI (MoodleDocs Aiken):** File teks biasa (`.txt`), UTF-8. Format paling sederhana di Moodle.
- **Template — TERBUKTI (MoodleDocs):** Tidak ada template unduhan; dokumentasi memberi dua contoh soal lengkap.
- **Panduan — TERBUKTI (MoodleDocs):** Sangat singkat (satu halaman): aturan + contoh + catatan penting + tips.
- **Pemetaan — TERBUKTI (MoodleDocs):** Soal harus **semuanya dalam satu baris** (tidak boleh ada baris baru di dalam soal). Setiap jawaban diawali **satu huruf kapital** + `.` atau `)` + spasi. Baris jawaban harus segera diikuti baris `ANSWER: ` + huruf jawaban (huruf kapital, ada spasi setelah titik dua). Contoh struktur: baris soal, baris `A. …`, `B. …`, dst., lalu `ANSWER: D`; antar soal dipisah baris kosong. Hanya untuk PG satu jawaban benar (per artikel akademik: hanya Multiple Choice satu jawaban benar).
- **Validasi + error — TERBUKTI (MoodleDocs + panduan institusi):** Sangat ketat dan gagal total bila dilanggar: huruf jawaban dan kata `ANSWER` **harus kapital** — bila tidak, import gagal. Karakter non-ASCII seperti kutip melengkung ("smart quotes") dapat menyebabkan error import; MoodleDocs menyarankan file UTF-8 + hindari karakter eksotis (elipsis `…`, dsb.). Panduan institusi menekankan "formatting is very particular and must be exactly as shown". Opsi import yang ada: `Get category from file`, `Get context from file`, `Match grades`, `Stop on error [Yes/No]` — artinya ada opsi "berhenti saat error" vs lanjut (mirip all_or_nothing vs skip_errors di v2), tetapi pesan error per barisnya TIDAK DIKETAHUI dari sumber yang dibaca.
- **Preview — TIDAK DIKETAHUI** apakah Aiken punya preview isi per soal; panduan Clarkson menyebut tahap "review for errors … Continue" setelah parse (lihat 2.2) — berlaku umum untuk import, bukan spesifik Aiken.
- **Batas jumlah — TIDAK DIKETAHUI** dari MoodleDocs Aiken. Artikel akademik menyebut jumlah alternatif jawaban maks 10 (A–J) — itu batas opsi, bukan batas soal.
- **Gambar/media — TERBUKTI (panduan institusi Ulethbridge + Clarkson):** File `.txt` **tidak bisa membawa gambar**; gambar ditambahkan manual per soal di Moodle setelah import.
- **Gampang vs bingung (DISIMPULKAN):**
  - Gampang: format paling mudah dibaca manusia; dapat ditulis di Word lalu Save As `.txt`; satu soal = satu blok pendek yang mudah diperiksa mata.
  - Bingung: (1) **Satu baris per soal** — soal stimulus/paragraf panjang tidak muat; forum Moodle menyebut soal panjang terpotong sampai baris pertama dan menyarankan pindah ke GIFT. (2) **Case-sensitive diam-diam** — `answer: d` (huruf kecil) gagal, dan penyebabnya tidak terlihat tanpa membaca aturan kapital. (3) **Tidak ada pembeda visual antara soal dan jawaban** selain posisi baris — satu baris kosong yang hilang/menumpuk menggeser parsing seluruh file sesudahnya. (4) **Hanya PG satu kunci** — guru yang butuh PGK/menjodohkan/esai harus belajar format kedua.

### 2.4 Format 3 — Moodle XML

- **Format file — TERBUKTI (MoodleDocs XML):** File XML (`.xml`), UTF-8, dibungkus `<quiz>…</quiz>`; tag `<?xml …?>` **harus benar-benar di baris pertama** (baris kosong pertama atau tag tambahan di baris pertama membingungkan parser). Setiap soal `<question type="…">` wajib punya `<name>` dan `<questiontext>`; tipe yang dicantumkan: `multichoice|truefalse|shortanswer|matching|cloze|essay|numerical|description`.
- **Template — TERBUKTI (MoodleDocs):** Tidak ada template unduhan; dokumentasi menyatakan import-ekspor seimbang — cara memahami format adalah **membuat soal contoh lalu mengekspornya** dan melihat XML hasil ekspor. Ini adalah "template by example" lewat ekspor, bukan file template.
- **Panduan — TERBUKTI (MoodleDocs):** Lengkap per tipe: struktur umum, tag bersama semua tipe (tags, penalty, generalfeedback, defaultgrade, hidden), lalu contoh per tipe (PG dengan `<answer fraction="100|0">` + feedback + `<single>`, `<shuffleanswers>`, `<answernumbering>`; Benar-Salah; Isian Singkat dengan alternatif jawaban + `<usecase>`; Numerik dengan `<tolerance>` + `<unit>`; Menjodohkan dengan `<subquestion>`; Esai; Cloze; Deskripsi). HTML di dalam teks harus dibungkus CDATA.
- **Pemetaan — TERBUKTI (MoodleDocs):** Kunci dinyatakan sebagai bobot `fraction` (100 = benar, 0 = salah, nilai antara = kredit parsial); kategori dapat dinyatakan lewat soal dummy `<question type="category"><category><text>$course$/NamaKategori</text></category></question>` dan berlaku bila opsi "Get category from file" dicentang; beberapa kategori dapat dinyatakan dalam satu file (soal setelah dummy kategori masuk kategori itu). Nilai default (`defaultgrade`), penalti, feedback umum, dan tag kata kunci ikut terbawa.
- **Validasi + error — TERBUKTI (MoodleDocs):** Parser XML **mengasumsikan file well-formed dan tidak mendeteksi/melaporkan error** — bila tidak well-formed, hasilnya adalah "unexpected errors". MoodleDocs secara eksplisit menyarankan memvalidasi XML lewat validator eksternal (mis. membuka file di browser) sebelum import. **DISIMPULKAN:** Ini model validasi terlemah dari semua pembanding — tidak ada pesan error bernomor baris dari importer; guru harus memvalidasi sendiri di luar sistem.
- **Preview — TIDAK DIKETAHUI** dari halaman MoodleDocs XML yang dibaca (tidak menyebut preview isi per soal).
- **Batas jumlah — TIDAK DIKETAHUI** dari halaman yang dibaca.
- **Gambar/media — TERBUKTI (MoodleDocs):** Tag `<image>` berisi URL gambar; di dalamnya dapat ada `<image_base64>` berisi data gambar ter-encode base64. Ekspor Moodle 2.x/3.x menyertakan gambar sebagai base64, sedangkan ekspor 1.9 hanya URL relatif (gambar tidak ikut). **DISIMPULKAN:** Media paling lengkap di antara format Moodle, tetapi file membengkak (base64) dan tidak ramah ditulis tangan — jalur realistisnya adalah ekspor, bukan ketik manual.
- **Gampang vs bingung (DISIMPULKAN):**
  - Gampang: paling ekspresif (bobot, penalti, feedback, tag, kategori, shuffle, penomoran jawaban); round-trip ekspor-impor menjaga struktur; kategori dalam file menghemat langkah pemilahan.
  - Bingung: (1) **Tidak untuk ditulis tangan** — satu soal = puluhan baris XML; satu tag tak tertutup = seluruh file gagal tanpa pesan jelas. (2) **Validator eksternal wajib** — sistem tidak memberi tahu letak kesalahan. (3) **Perilaku kategori bergantung pada checkbox** ("Get category from file") — bila tidak dicentang, kategori di file diabaikan dan soal masuk kategori pilihan UI; mudah salah sangka.

### 2.5 Format Moodle lain (ringkas)

- **TERBUKTI (MoodleDocs XML "See also" + halaman GIFT):** Moodle memiliki format impor lain (daftar format di halaman Import questions), termasuk Blackboard, WebCT, dan format lain — tetapi riset ini TIDAK memperdalam masing-masing karena dokumentasi per formatnya tidak dibaca penuh. **TIDAK DIKETAHUI** detail kolom/validasi format-format tersebut dari riset ini.
- **DISIMPULKAN:** Strategi Moodle adalah "banyak format impor, satu bank soal" — kebalikan dari v2 saat ini ("satu format CSV"). Konsekuensinya: guru Moodle dapat memilih format sesuai sumber soal yang sudah dimiliki (teks sederhana → Aiken; kaya fitur → XML; tengah-tengah → GIFT), tetapi harus belajar aturan tiap format.

---

## 3. Quizizz (kini Wayground) — import spreadsheet

Sumber utama: panduan pihak ketiga yang membaca template resmi Wayground (VidQuiz, diperbarui 13 Okt 2026 — menyatakan spesifikasinya dibaca dari file template resmi Wayground, Agustus 2026) + panduan Monsha + halaman harga Wayground (dari laporan sebelumnya). **Catatan bukti:** tidak ditemukan halaman Help Center resmi Wayground/Quizizz yang merinci kolom template dalam riset ini; rincian kolom di bawah berasal dari VidQuiz yang mengklaim membaca file template resmi — sehingga labelnya TERBUKTI (dari template resmi via VidQuiz), bukan "halaman dokumentasi resmi".

- **Format file — TERBUKTI (VidQuiz):** `.xlsx` (template resmi diterbitkan sebagai Google Sheet; guru menyalinnya, mengisi, lalu unduh sebagai `.xlsx`). **CSV TIDAK dibaca importer** — bila sumbernya CSV, guru harus menempelkan kolom ke template lalu simpan sebagai `.xlsx`.
- **Template yang disediakan — TERBUKTI (VidQuiz + Monsha):** YA, ada template resmi (Google Sheet) + tombol unduh template di dalam alur import. Alur: Create → Assessment → Start from scratch → scroll ke bawah ke "Import existing files" → Spreadsheet → unggah `.xlsx` → review → Publish.
- **Panduan pengisian — TERBUKTI (VidQuiz):** Template berisi kolom + (menurut VidQuiz) contoh baris. Panduan kolomnya ringkas (satu tabel 11 kolom).
- **Pemetaan kolom → field soal — TERBUKTI (VidQuiz), 11 kolom, satu baris = satu soal:**

  | Kolom | Wajib? | Pemetaan |
  |---|---|---|
  | Question Text | Wajib | Teks soal |
  | Question Type | Opsional (kosong = Multiple Choice) | Multiple Choice, Checkbox, Fill-in-the-Blank, Open-Ended, Poll, Draw |
  | Option 1 | Wajib kecuali Open-Ended/Draw | Opsi 1 |
  | Option 2 | Wajib kecuali Open-Ended/Draw | Opsi 2 |
  | Option 3, 4, 5 | Opsional | Opsi 3–5 (maks 5 opsi) |
  | Correct Answer | Tergantung tipe | **NOMOR opsi 1–5**, bukan huruf. Checkbox: beberapa nomor dipisah koma (`1,2,3`). Open-Ended: kosong. |
  | Time in seconds | Opsional (kosong = 30 detik) | Timer per soal |
  | Image Link | Opsional | URL gambar (bukan file tertanam) |
  | Answer explanation | Opsional | Penjelasan jawaban benar |

  Khusus Fill-in-the-Blank: kolom Option berisi jawaban-jawaban yang diterima (bukan pilihan), dan Correct Answer dibiarkan kosong.
- **Validasi + pesan error — TIDAK DIKETAHUI secara rinci:** VidQuiz tidak merinci daftar pesan error / apakah error bernomor baris. Yang TERBUKTI: ada tahap review setelah unggah (guru meninjau soal yang terbaca lalu Publish). **DISIMPULKAN (alasan konkret):** Kolom Correct Answer adalah jebakan paling berbahaya — VidQuiz secara eksplisit memperingatkan bahwa spreadsheet berisi huruf A/B/C/D **terunggah tanpa error** tetapi menghasilkan kunci yang salah pada semua soal ("fails quietly"). Ini adalah kegagalan validasi diam-diam: importer tidak menolak nilai huruf yang tidak valid sebagai nomor, sehingga guru baru tahu saat siswa memprotes / rata-rata kelas aneh. Ini kebalikan dari v2 (kunci huruf yang tidak cocok opsi → error per baris).
- **Preview sebelum simpan — TERBUKTI (VidQuiz + Monsha):** YA — setelah unggah, soal yang terbaca ditampilkan untuk ditinjau/diedit, baru Publish. Ini preview isi (bukan hanya hitungan), lebih kuat dari v2.
- **Batas jumlah — TIDAK DIKETAHUI** dari sumber yang dibaca (tidak ada angka batas soal per file di VidQuiz/Monsha). Batas yang disebut VidQuiz adalah batas produk lain (sesi live Basic 30 peserta) — bukan batas import.
- **Gambar/media — TERBUKTI (VidQuiz):** Satu gambar per soal, lewat kolom Image Link berupa URL (gambar harus sudah ter-host di tempat yang dapat dijangkau platform). Hanya teks polos yang diimpor — bold/italic/warna di-strip. Rumus/format persamaan yang bergantung format harus dibangun ulang manual di creator. **DISIMPULKAN:** Media "setengah jalan" — lebih baik dari v2 (yang nol), tetapi guru tanpa hosting gambar tidak dapat memakai kolom ini, dan gambar tertanam di file `.xlsx` tidak ikut.
- **Jalur import lain — TERBUKTI (VidQuiz):** (a) Import Google Forms — soal masuk sebagai tipe Poll (tidak dinilai default); bila skor penting, guru harus mengubah tiap soal ke tipe dinilai satu per satu. (b) Rute dokumen/tempel via Wayground AI (PDF, PPT, PPTX, DOC, DOCX, JPG, PNG) — hanya 16.000 karakter pertama file / 10.000 karakter tempelan per unggahan, satu file per unggahan; AI menginferensi kunci (bukan ditegaskan guru). **DISIMPULKAN:** Rute AI adalah "impor tebakan" — kunci hasil inferensi wajib ditinjau; rute spreadsheet tetap satu-satunya jalur yang kuncinya ditegaskan guru.
- **Gampang vs bingung (DISIMPULKAN):**
  - Gampang: template resmi + 11 kolom bermakna (ada kolom waktu per soal, penjelasan, link gambar); tipe soal ikut dalam satu file; preview sebelum Publish; format kolom mendekati cara guru menyusun soal di spreadsheet.
  - Bingung: (1) **Correct Answer berupa nomor, bukan huruf** — bertentangan dengan konvensi hampir semua alat soal lain (termasuk v2, Moodle Aiken, Respondus) dan gagal diam-diam bila diisi huruf. (2) **Menu import tersembunyi** — opsi import berada di bawah semua kategori tipe soal, bukan di atas; guru baru mudah tidak menemukannya. (3) **Semantik kolom berubah per tipe** — kolom Option berarti "pilihan" untuk PG tetapi "jawaban diterima" untuk Fill-in-the-Blank, dan Correct Answer harus dikosongkan untuk tipe itu; salah paham = kunci hilang. (4) **CSV tidak diterima** — guru yang sudah punya CSV harus pindah ke template `.xlsx` dulu. (5) **Import Google Forms berubah menjadi Poll** — nilai hilang kecuali dikonversi manual per soal.

---

## 4. Kahoot! — import spreadsheet

Sumber utama resmi: Kahoot! Help & Resource Center — "How to import questions from a spreadsheet to your kahoot" (https://support.kahoot.com/hc/en-us/articles/115002812547-How-to-import-questions-from-a-spreadsheet-to-your-kahoot, diperbarui 15 Okt 2026). Dibaca penuh.

- **Format file — TERBUKTI (Kahoot Help):** Wajib memakai template resmi; simpan sebagai `.xlsx` (MS Excel 2007 atau lebih baru).
- **Template yang disediakan — TERBUKTI (Kahoot Help):** YA — tautan "Download the official Kahoot! template" di dalam artikel + di dialog Import spreadsheet (Add → Import → Import spreadsheet). Template dibuka di Excel / LibreOffice / Google Sheets / Numbers.
- **Panduan pengisian — TERBUKTI (Kahoot Help):** 6 langkah + FAQ, ringkas dan eksplisit: isi kolom yang ditentukan; batas karakter; nilai waktu yang diizinkan; cara menandai jawaban benar; simpan `.xlsx`.
- **Pemetaan kolom → field soal — TERBUKTI (Kahoot Help):** Nama kolom persis TIDAK tercantum dalam teks artikel yang terbaca (artikel menyebut "designated columns" tanpa daftar nama kolom di teks; daftar kolom hanya terlihat di gambar template yang tidak terbaca teksnya). Yang TERBUKTI dari teks:
  - Soal (maks **120 karakter**), jawaban (maks **75 karakter**), tiap soal **minimal 2 jawaban**.
  - Time limit per soal hanya boleh **5, 10, 20, 30, 60, atau 120** detik.
  - Jawaban benar dinyatakan dengan **nomor jawaban dipisah koma** (mis. `1` atau `2,3`) — sama seperti Quizizz: nomor, bukan huruf.
  - **Hanya tipe soal Quiz standar** yang bisa diimpor; tipe lain (True/False, Puzzle, dsb.) tidak didukung lewat spreadsheet.
  - **TIDAK DIKETAHUI** dari teks artikel: nama kolom persis, jumlah kolom jawaban maks (apakah 4), dan apakah ada kolom gambar/penjelasan. Panduan pihak ketiga menyebut batas lebih ketat di versi lama (soal 95 karakter, jawaban 60) — perbedaan ini TIDAK diselesaikan di sini; yang dipakai adalah angka artikel resmi terbaru (120/75). Jangan pakai angka lama.
- **Validasi + pesan error — TERBUKTI (Kahoot Help):** "Sistem akan otomatis memeriksa error. Bila ditemukan, Anda bisa kembali memperbaiki spreadsheet atau lanjut dan membuang soal yang bermasalah." + FAQ: "sistem import akan memeriksa formatting errors; Anda bisa memperbaikinya di file atau lanjut tanpa mengimpor soal yang bermasalah." Ini adalah model **skip_errors dengan pilihan eksplisit** (perbaiki vs buang yang bermasalah), mirip mode `skip_errors` v2 tetapi dengan keputusan di tangan guru per batch. **TIDAK DIKETAHUI:** apakah pesan error menyebut nomor baris / nama kolom yang salah — artikel tidak merincinya.
- **Preview sebelum simpan — TERBUKTI (Kahoot Help):** Setelah import berhasil, soal "langsung muncul di kahoot creator" — guru dapat melihat/mengeditnya di creator sebelum dipakai. **DISIMPULKAN:** Ini preview pasca-import di editor (soal sudah masuk draf kahoot), bukan dialog preview khusus sebelum simpan; setara dengan v2 (soal masuk sebagai draft lalu diedit), tetapi dengan tahap cek error sebelum masuk.
- **Batas jumlah — TIDAK DIKETAHUI** dari artikel resmi yang dibaca (tidak ada angka maks soal per file).
- **Gambar/media — TERBUKTI (Kahoot Help FAQ):** **TIDAK BISA** — gambar dan karakter khusus tidak dapat diimpor lewat spreadsheet. Simbol matematika dapat disalin-tempel ke creator setelahnya (langsung atau sebagai kode LaTeX). Mengunduh kahoot yang sudah ada sebagai spreadsheet TIDAK didukung; sebagai gantinya ada fitur gabung kahoot + laporan spreadsheet pasca-permainan (berisi soal + opsi).
- **Gampang vs bingung (DISIMPULKAN):**
  - Gampang: template resmi + aturan sangat eksplisit (batas karakter, daftar nilai waktu yang diizinkan, contoh penanda jawaban benar); cek error otomatis + opsi "buang yang bermasalah" membuat satu soal rusak tidak menggagalkan seluruh file; alur import hanya 6 langkah.
  - Bingung: (1) **Batas karakter sangat ketat** (120/75) — soal cerita / stimulus panjang tidak muat; guru harus memangkas soal agar lolos, bukan sekadar memindahkan soal. (2) **Hanya tipe Quiz** — guru yang punya bank campuran (B/S, urutan, isian) harus membuat sisanya manual. (3) **Jawaban benar = nomor** — sama seperti Quizizz, bertentangan dengan konvensi huruf; artikel memberi contoh (`1`, `2,3`) sehingga risikonya lebih kecil daripada Quizizz, tetapi tetap rawan tertukar bila guru mengira kolom itu huruf. (4) **Tanpa media** — soal bergambar harus ditambahkan satu per satu di creator. (5) **Nama kolom tidak tertulis di teks artikel** — guru yang hanya membaca teks (tanpa membuka template) tidak tahu nama kolomnya.

---

## 5. Exam.net — apakah ada import soal?

Sumber utama resmi: Exam.net Support — "Create an exam" (https://support.exam.net/s/article/create-an-exam), "The exam format 'Use question types'" (https://support.exam.net/s/article/the-exam-format-use-question-types?language=en_US), "The exam format 'Create document'" (https://support.exam.net/s/article/the-exam-format-create-document). Dibaca penuh. Temuan ini MEMPERBARUI laporan sebelumnya (yang berlabel TIDAK DIKETAHUI untuk import Exam.net).

### 5.1 Jawaban singkat

- **TERBUKTI (Exam.net Support):** Exam.net **TIDAK memiliki import file soal massal** (CSV/Excel/Word/XML) seperti Moodle/Kahoot/Quizizz. Yang ada adalah tiga hal berbeda: (1) **salin soal dari ujian yang sudah ada** ("Import questions" = copy antar-ujian di dalam platform), (2) **tempel teks** dari Word ke editor dokumen, (3) **unggah PDF** sebagai naskah ujian + konverter PDF. Ketiganya bukan "import bank soal" dalam arti v2.

### 5.2 Jalur 1 — "Import questions" = salin dari ujian lama (bukan dari file)

- **Format file — TERBUKTI:** Tidak ada file. Sumbernya adalah ujian lain di dalam akun/platform yang sama.
- **Template — TERBUKTI:** Tidak ada template file; yang ada adalah daftar ujian sumber yang dapat dicari.
- **Panduan — TERBUKTI (Exam.net Support):** 3 langkah: cari ujian sumber → centang soal yang ingin disalin (atau "All") → klik Insert. Soal hasil salinan dapat diedit bebas tanpa memengaruhi ujian lama.
- **Pemetaan kolom — TIDAK BERLAKU:** Tidak ada kolom; soal disalin utuh (teks + tipe + aturan marking) dari ujian sumber.
- **Validasi + error — TIDAK DIKETAHUI:** Dokumentasi tidak menyebut validasi/error untuk jalur salin ini.
- **Preview — TERBUKTI:** Ada preview ujian di editor (ikon kaca pembesar) untuk mencoba marking — preview ini berlaku untuk seluruh ujian, bukan khusus dialog import.
- **Batas — TIDAK DIKETAHUI:** Tidak ada angka batas soal yang disalin.
- **Media — TERBUKTI:** Soal dapat berisi teks, gambar, rumus, tabel — tetapi itu karena soal dibuat/diisi di editor, bukan karena diimpor dari file. Soal hasil salinan membawa medianya karena medianya sudah ada di ujian sumber.
- **Gampang vs bingung (DISIMPULKAN):** Gampang untuk "pakai lagi soal semester lalu" (3 klik, tanpa format file). Bingung untuk guru baru/migrasi: tidak ada cara membawa bank soal dari luar (dari Word/Excel/sekolah lain) kecuali mengetik ulang / menempel satu per satu — tidak ada jalan pintas massal.

### 5.3 Jalur 2 — "Create document": tulis/tempel teks (bukan parsing soal)

- **Format — TERBUKTI (Exam.net Support):** Di format "Upload or create a document", guru memilih "Write or paste in exam questions" dan mengisi editor teks — dapat menulis langsung atau menempel teks dari tempat lain, mis. dokumen Word. Setelah jadi, guru dapat preview lalu lanjut ke konfigurasi ujian. Di bawah Resources, dapat mengunggah file PDF tambahan.
- **Template — TERBUKTI:** Tidak ada template file soal; editor teks adalah kanvas kosong.
- **Panduan — TERBUKTI:** Singkat (satu halaman): pilih format → tulis/tempel → preview → konfigurasi. FAQ: ujian TIDAK dapat dibagi ke beberapa bagian; komentar pada jawaban siswa dapat ditambahkan lewat Inline Feedback saat marking.
- **Pemetaan — TIDAK BERLAKU:** Tidak ada parsing kolom/kunci — teks yang ditempel adalah naskah ujian (soal + instruksi) untuk dijawab siswa di area tulis terpisah (pola "naskah + area tulis" dari laporan sebelumnya), bukan soal terstruktur dengan kunci.
- **Validasi + error — TIDAK DIKETAHUI:** Tidak ada validasi soal terstruktur karena tidak ada parsing.
- **Preview — TERBUKTI:** Ada preview sebelum konfigurasi.
- **Batas — TIDAK DIKETAHUI:** Tidak ada angka batas.
- **Media — TERBUKTI:** PDF tambahan lewat Resources; editor dokumen menerima tempelan teks (gambar/tempelan media tidak dirinci di halaman ini — TIDAK DIKETAHUI apakah gambar hasil tempel ikut terbawa).
- **Gampang vs bingung (DISIMPULKAN):** Gampang untuk ujian uraian/naskah (cara kerja = mengetik dokumen, tanpa belajar sintaks). Bingung bila guru mengira "tempel dari Word" = "import soal terstruktur": kunci, bobot, dan tipe soal TIDAK terbentuk dari tempelan — guru yang butuh auto-marking harus memakai jalur 5.4, bukan jalur ini.

### 5.4 Jalur 3 — "Use question types": soal dibuat di editor + salin + AI (bukan import file)

- **Format — TERBUKTI (Exam.net Support):** Soal dibuat dari tipe bawaan: auto-marked (Multiple-choice, Simple answer, Fill the gaps, Match answers, Grid) + tanpa auto-marking (Free-text, Answer with attachment) + blok informasi. Ada generator soal AI + "Import questions" (salin dari ujian lama, lihat 5.2). Ada juga konverter PDF → ujian tipe assignment (disebut di "Create an exam" sebagai tip).
- **Template — TERBUKTI:** Tidak ada template file.
- **Panduan — TERBUKTI:** Lengkap per tipe (cara membuat PG, Fill the gaps dengan gap + jawaban alternatif + aturan grading seperti case-sensitive / ignore spaces / strict koma-titik, marking rules). Preview marking tersedia lewat kaca pembesar (coba jawaban benar/salah untuk melihat auto-marking).
- **Pemetaan — TIDAK BERLAKU** (tidak ada file).
- **Validasi + error — TIDAK DIKETAHUI** untuk import (tidak ada import file).
- **Preview — TERBUKTI:** Preview ujian + uji marking tersedia di editor.
- **Batas — TIDAK DIKETAHUI.**
- **Media — TERBUKTI:** Soal dapat berisi teks, gambar, rumus, tabel (diisi di editor).
- **Gampang vs bingung (DISIMPULKAN):** Gampang untuk kualitas per soal (aturan marking eksplisit, preview marking). Bingung untuk skala: membuat 100 soal berarti 100 kali interaksi editor — tidak ada jalan bulk. Untuk SMK, Exam.net adalah contoh "import paling lemah, editor per soal paling kuat".

---

## 6. ProProfs Quiz Maker — import Excel (+ AI)

Sumber utama resmi: ProProfs Quiz Maker Help — "How to Bulk-Import Questions From Excel" (https://quiz.proprofs.com/how-to-bulk-import-questions-from-an-excel-document-to-your-quiz, diperbarui 8 Jan 2026). Dibaca penuh. Temuan ini memperdalam laporan sebelumnya (yang hanya tahu "ada import Excel").

- **Format file — TERBUKTI (ProProfs Help):** File Excel (`.xls`/`.xlsx` — artikel menyebut "Excel file"/"Excel sheet"; ekstensi persis tidak dirinci di teks). Selain itu ada jalur AI (lihat bawah) yang menerima `.DOCX`, `.PDF`, `.TXT` maks 20 MB.
- **Template yang disediakan — TERBUKTI (ProProfs Help):** YA — file Excel contoh ("sample Excel file") dapat diunduh dari dalam alur import (Add → Import Questions → Excel → Next) dan dari artikel. Instruksi eksplisit: unduh file contoh, ganti soal/opsi/detail dummy dengan data asli.
- **Panduan pengisian — TERBUKTI (ProProfs Help):** Ringkas: artikel menyatakan di Excel dapat diisi soal + jawaban + tipe soal + **nilai/poin per soal** + **topik/kategori per soal**, lalu 2 langkah (pilih Excel → Upload file). **TIDAK DIKETAHUI:** nama kolom persis + contoh isi sel per tipe — gambar contoh file Excel di artikel hanya terbaca sebagai keterangan gambar ("Sample questions in the Excel Sheet"); isi selnya tidak terbaca lewat alat baca teks, sehingga TIDAK ditebak. Ini celah bukti yang harus dicatat: kolom persis ProProfs TIDAK dapat dikutip dari riset ini.
- **Pemetaan kolom → field soal — TIDAK DIKETAHUI (nama kolom persis):** Yang TERBUKTI hanya kelompok field yang didukung: teks soal, jawaban, tipe soal, poin per soal, topik/kategori per soal. **DISIMPULKAN:** Cakupan field ini lebih dekat ke kebutuhan v2 daripada Kahoot/Quizizz (ada poin + kategori dalam satu file — dua hal yang hilang di CSV v2).
- **Validasi + pesan error — TIDAK DIKETAHUI:** Artikel tidak merinci pesan error / apakah error bernomor baris. Yang TERBUKTI: setelah unggah, soal yang ditambahkan "muncul sebagai soal biasa" (dapat diedit di editor).
- **Preview sebelum simpan — DISIMPULKAN (dari artikel):** Artikel menunjukkan tangkapan layar "setelah selesai mengunggah" + soal muncul sebagai soal biasa — tetapi apakah ada dialog preview/konfirmasi sebelum soal masuk kuis, TIDAK DIKETAHUI dari teks yang dibaca. Jangan klaim ada preview.
- **Batas jumlah — TERBUKTI (ProProfs Help):** Paket Forever Free: import Excel mendukung **maks 8 soal**; untuk tak terbatas perlu upgrade. Ini satu-satunya pembanding dengan batas import yang dinyatakan eksplisit dan terikat paket. **TIDAK DIKETAHUI:** batas untuk paket berbayar (artikel hanya menyebut "unlimited questions" sebagai manfaat upgrade — angka pastinya tidak dirinci).
- **Gambar/media — TIDAK DIKETAHUI untuk jalur Excel:** Artikel import Excel tidak menyebut kolom gambar/media. ProProfs secara umum mendukung unggah dokumen/video/audio sebagai konten soal (dari laporan sebelumnya, panduan lama), tetapi apakah itu bisa lewat file Excel import — TIDAK DIKETAHUI. Jangan klaim bisa/tidak.
- **Jalur AI (pembeda ProProfs) — TERBUKTI (ProProfs Help):** Selain Excel, ada "bulk-upload via ProProfs AI": lampirkan dokumen di editor AI; bila dokumen sudah berisi soal → AI mengidentifikasi + mengimpor; bila dokumen berisi materi (catatan, manual) → AI menggenerasi soal baru. Format `.DOCX`, `.PDF`, `.TXT`, maks 20 MB. Soal hasil AI masuk sebagai soal biasa (dapat ditinjau/diedit/susun ulang). **DISIMPULKAN:** Ini "impor tanpa format" — guru tidak perlu belajar kolom; risikonya adalah kunci/teks hasil inferensi AI yang wajib ditinjau (sama seperti rute AI Wayground).
- **Gampang vs bingung (DISIMPULKAN):**
  - Gampang: file contoh unduhan + instruksi "ganti dummy dengan data asli" adalah pola template terbaik di antara pembanding (guru tidak mengetik header dari nol); authoring offline (siapkan Excel tanpa login/internet) + multi-author (siapa pun bisa menyiapkan file, tanpa akun ProProfs); ada poin + kategori dalam satu file; jalur AI untuk guru yang tidak mau format apa pun.
  - Bingung: (1) **Nama kolom tidak tertulis di teks artikel** — guru yang hanya membaca teks (mis. lewat screen reader / teks saja) tidak tahu nama kolomnya; harus membuka file contoh. (2) **Batas 8 soal di Free** — guru yang mencoba 50 soal di Free akan terhenti tanpa penjelasan di muka bila tidak membaca catatan; batas ini tersembunyi di tengah artikel, bukan di dialog import (dari bukti yang dibaca). (3) **Dua jalur (Excel vs AI) dengan jaminan berbeda** — Excel menegaskan kunci, AI menginferensi kunci; guru yang salah pilih jalur bisa mendapat kunci tebakan. (4) **Validasi tidak didokumentasikan** — tidak jelas apa yang terjadi bila satu baris salah format (ditolak semua? dilewati? masuk dengan kunci kosong?).

---

## 7. Google Forms — apakah bisa import soal?

Sumber: panduan pihak ketiga (Guiding Tech — "How to Import Questions Into Google Forms Easily", https://www.guidingtech.com/import-questions-into-google-forms/) + temuan laporan sebelumnya. **Catatan bukti:** tidak ditemukan halaman Help Center resmi Google yang merinci import soal dalam riset ini; rincian di bawah berasal dari Guiding Tech + laporan sebelumnya (yang juga memakai sumber yang sama).

- **Format file — TERBUKTI (Guiding Tech + laporan sebelumnya):** TIDAK ada import file soal (tidak ada CSV/Word/Excel bawaan). Satu-satunya import bawaan adalah **import soal dari form lain** (antar-Form).
- **Template — TIDAK BERLAKU:** Tidak ada template file soal karena tidak ada import file.
- **Panduan — TERBUKTI (Guiding Tech):** 4 langkah: buka form baru → menu Import → pilih form sumber → centang soal yang diinginkan (atau centang atas untuk semua) → Import Questions. Tombol Change Form untuk ganti form sumber; tab Previously Selected mengingat form sebelumnya. Hanya satu form sumber per giliran (tidak bisa pilih banyak form sekaligus).
- **Pemetaan — TERBUKTI (Guiding Tech):** Soal disalin utuh dari form sumber, **termasuk validasinya**. Tidak ada pemetaan kolom.
- **Validasi + error — TIDAK BERLAKU / TIDAK DIKETAHUI:** Tidak ada validasi file; error import file tidak ada karena tidak ada file.
- **Preview — TERBUKTI (Guiding Tech):** Daftar soal form sumber ditampilkan untuk dipilih (centang per soal) sebelum import — ini adalah preview-pilih, bukan preview isi hasil parsing.
- **Batas — TIDAK DIKETAHUI:** Tidak ada angka batas soal yang diimpor per giliran di sumber yang dibaca.
- **Gambar/media — TERBUKTI (DISIMPULKAN dari sifat salin):** Soal disalin utuh dari form sumber, sehingga gambar yang sudah ada di soal sumber ikut tersalin (DISIMPULKAN — tidak dinyatakan eksplisit di Guiding Tech, tetapi mengikuti sifat salin utuh + validasi ikut tersalin). Yang TIDAK ADA adalah jalur "unggah file berisi soal bergambar".
- **Jalur pihak ketiga — TERBUKTI (laporan sebelumnya + Guiding Tech):** Import dari dokumen lain memerlukan add-on pihak ketiga (mis. migrator antar-platform). Bukan fitur bawaan.
- **Gampang vs bingung (DISIMPULKAN):**
  - Gampang: tanpa format file sama sekali — guru cukup memilih soal dari form lama; validasi ikut tersalin (tidak perlu setting ulang); cocok untuk "pakai lagi soal tahun lalu".
  - Bingung: (1) **Tidak ada bulk dari luar** — bank soal 500 soal di Word/Excel tidak bisa masuk kecuali diketik ulang atau lewat add-on pihak ketiga; untuk SMK ini adalah kegagalan fungsi, bukan sekadar ketidaknyamanan. (2) **Satu form sumber per giliran** — menggabungkan soal dari 5 form lama berarti 5 kali alur import. (3) **Tidak ada pemetaan kisi-kisi/kompetensi/bobot** — Forms tidak punya konsep itu; bobot hanya poin per soal di mode kuis. (4) Guru yang mencari "import CSV ke Forms" akan menemukan add-on pihak ketiga yang kualitas/keamanannya TIDAK terverifikasi di sini — risiko privasi soal.

---

## 8. Aplikasi CBT / kuis lain yang terdokumentasi publik (tambahan)

### 8.1 Respondus (alat impor soal ke LMS: Canvas, Blackboard, Moodle, dsb.)

Sumber: panduan institusi (UConn Knowledge Base, https://kb.uconn.edu/space/TL/26024247323; SIUE KB, https://kb.siue.edu/page.php?id=62185; Lone Star VTAC, http://vtac.lonestar.edu/help/a195; Groningen/Nestor; Terry/UGA, https://goterry.atlassian.net/wiki/spaces/BC/pages/1030717441/Respondus+Exam+Format+Guide) + panduan Respondus (Docsbay). Respondus bukan platform ujian mandiri, melainkan jembatan Word → LMS — dimasukkan karena formatnya adalah pola "Word guru → soal terstruktur" yang paling dekat dengan naskah soal guru SMK.

- **Format file — TERBUKTI (panduan institusi):** File "Standard Format": teks biasa (`.txt`), Rich Text (`.rtf`), atau Word (`.doc`/`.docx`). File `.txt` disebut "paling bersih" karena Word kadang mengembalikan auto-numbering dan menghapus tanda asterisk.
- **Template — TERBUKTI (UConn):** Ada file contoh Word Respondus yang dapat diunduh ("Respondus Word Sample File") + panduan format. Guru mengisi file contoh, bukan mengetik format dari nol.
- **Panduan — TERBUKTI:** Sangat eksplisit (aturan wajib + opsional + contoh per tipe). Aturan wajib: tiap soal diawali nomor + `.` atau `)` + minimal satu spasi; tiap jawaban diawali huruf (a–t) + `.` atau `)`; tanpa header/footer halaman; tanpa bullet (huruf diketik manual, bukan auto-numbering Word).
- **Pemetaan — TERBUKTI (panduan institusi), diringkas:**
  - Kunci: tanda asterisk `*` tepat sebelum huruf jawaban, tanpa spasi (`*b. …`) — atau daftar jawaban di akhir file (`Answers: 1. A 2. C 3. B`).
  - Esai: baris pertama `Type: E`. Isian: `Type: F` + semua jawaban benar yang diterima ditulis di field jawaban (dipisah koma). Menjodohkan: `Type: MT` + pasangan `kiri=kanan` (spasi di sekitar `=` diizinkan).
  - Opsional: `Title: …` di atas soal; feedback; nilai poin.
  - Benar-Salah: jawaban `True` harus di atas `False` — bila `False` duluan, soal diimpor sebagai PG (jebakan urutan yang didokumentasikan).
- **Validasi + error — TERBUKTI (Groningen + UConn):** Ada tombol **Preview** sebelum Finish yang menampilkan warnings/notes. Contoh warning yang didokumentasikan: judul duplikat (boleh diabaikan — terjadi karena Respondus memakai 20 karakter pertama soal sebagai judul bila tidak ada `Title:`) dan "soal tanpa jawaban benar" (JANGAN diabaikan — kunci hilang atau format salah). **DISIMPULKAN:** Model warning berlapis (abaikan vs jangan abaikan) lebih informatif daripada sekadar "gagal/sukses".
- **Preview — TERBUKTI:** Ada Preview sebelum Finish (lihat di atas).
- **Batas — TIDAK DIKETAHUI** (batas soal per file tidak disebut di panduan yang dibaca).
- **Gambar/media — TERBUKTI (Docsbay + panduan):** Gambar tertanam di Word (`.doc`/`.docx`) ikut terimpor otomatis. Untuk `.txt`/`.rtf`/CSV, gambar diimpor lewat metode terpisah (menunjuk folder gambar). Format kaya (bold, italic, underline, superscript, subscript) dipertahankan untuk `.rtf`/Word; `.txt` kehilangan format.
- **Gampang vs bingung (DISIMPULKAN):**
  - Gampang: formatnya adalah cara guru SUDAH menulis soal di Word (nomor + huruf + asterisk) — beban belajar paling rendah di antara semua pembanding; file contoh + preview + warning berlapis; gambar ikut dari Word tanpa langkah tambahan.
  - Bingung: (1) **Auto-numbering Word adalah musuh** — Word dapat menghapus `*` atau mengubah penomoran; guru harus mematikan auto-numbering atau menyimpan sebagai `.txt` (yang justru menghilangkan gambar). (2) **Urutan True/False bermakna** — tertukar = tipe soal berubah diam-diam. (3) **Judul otomatis 20 karakter** — judul duplikat membanjiri warning sehingga warning penting mudah terlewat. (4) Respondus adalah aplikasi desktop berbayar/lisensi kampus (dari konteks panduan institusi) — bukan fitur web gratis; untuk SMK, ini adalah ketergantungan lisensi + instalasi.

### 8.2 Canvas LMS — QTI (.zip)

Sumber: panduan pihak ketiga (ASEE 2024, "Best Practices for Using Generative AI to Create Quiz Content for the Canvas LMS", https://peer.asee.org/best-practices-for-using-generative-ai-to-create-quiz-content-for-theCanvas-lms.pdf — URL sesuai hasil pencarian; bila tautan tidak terbuka, gunakan pencarian judulnya) + repositori konverter (text2qti, moodletocanvasquizconversion) + panduan konverter K-State.

- **Format file — TERBUKTI (ASEE + panduan konverter):** Paket **QTI `.zip`** (berisi XML QTI). Alur: Settings → Import Course Content → Content Type "QTI .zip file" → unggah ZIP (jangan diekstrak) → pilih bank soal tujuan → Import.
- **Template — TERBUKTI (ASEE + konverter):** Tidak ada template `.zip` resmi yang dibagikan sebagai file unduhan di sumber yang dibaca; yang ada adalah (a) **konverter CSV → QTI** (Kansas State, canconvert.kstate.edu/qti) dengan tata letak CSV yang ditentukan, dan (b) format teks text2qti (`*a)` untuk benar, `[*]` untuk multi-benar, `= 2` untuk isian, front matter YAML untuk judul/poin/acakan).
- **Panduan — TERBUKTI (ASEE):** Tata letak CSV konverter: **tanpa baris header** (header akan menyebabkan error); Kolom A = tipe soal (MC, MR, TF), Kolom B = kosong, Kolom C = nilai poin, Kolom D = soal, Kolom E = jawaban benar, Kolom F–J = pilihan jawaban. Simpan sebagai CSV (bukan CSV UTF-8). Lalu konversi → ZIP → import.
- **Pemetaan — TERBUKTI (ASEE + text2qti):** Kunci dinyatakan lewat posisi/kolom (kolom E di CSV konverter; `*` di text2qti). QTI sendiri mendukung bobot, feedback, bank soal, dan tipe soal LMS.
- **Validasi + error — TERBUKTI (ASEE):** Header di baris atas = error (aturan eksplisit). **TIDAK DIKETAHUI:** daftar pesan error lain / apakah error bernomor baris.
- **Preview — TIDAK DIKETAHUI** apakah ada preview isi sebelum import selesai; ASEE menyarankan verifikasi soal di Canvas setelah import ("navigate to quiz section to verify").
- **Batas — TIDAK DIKETAHUI** dari sumber yang dibaca.
- **Gambar/media — TIDAK DIKETAHUI** dari sumber yang dibaca untuk jalur CSV→QTI (konverter CSV yang didokumentasikan ASEE tidak menyebut kolom gambar). QTI sebagai standar mendukung media, tetapi jalur CSV yang didokumentasikan tidak menunjukkannya — jangan klaim.
- **Gampang vs bingung (DISIMPULKAN):**
  - Gampang: QTI adalah standar interoperabilitas (soal dapat pindah antar-LMS); konverter CSV memberi jalan masuk spreadsheet tanpa menulis XML.
  - Bingung: (1) **CSV tanpa header + kolom B kosong yang bermakna** — layout positional yang rapuh: satu kolom bergeser = seluruh pemetaan salah; kolom kosong yang "harus ada" adalah jebakan klasik. (2) **Dua tahap + satu larangan** (konversi → ZIP → import; jangan ekstrak ZIP) — guru yang mengekstrak ZIP karena kebiasaan akan gagal. (3) **Tidak ada template resmi satu klik** — guru bergantung pada konverter pihak ketiga (K-State) yang ketersediaannya di luar kendali Canvas.

### 8.3 Socrative — import Excel (template + batas eksplisit)

Sumber: panduan pihak ketiga (skill templating yang mengutip Help Center Socrative, https://help.socrative.com/en/articles/2155335-import-a-quiz, terverifikasi 13 Agu 2026 menurut kutipannya) + panduan UAB (https://www.uab.edu/icac/images/How-to_Guides/Socrative_How-to_Guide.pdf) + panduan template (Scribd "Socrative Quiz Template"). **Catatan bukti:** halaman Help Center Socrative sendiri tidak dibuka penuh dalam riset ini; rincian di bawah berasal dari sumber yang mengutipnya — label TERBUKTI (via kutipan Help Center) dipakai untuk butir yang dikutip eksplisit, sisanya DISIMPULKAN.

- **Format file — TERBUKTI (kutipan Help Center + UAB):** Template Excel (`.xls`/`.xlsx`) + jalur kedua: import kuis dari guru lain lewat kode SOC (`SOC-XXXXX`).
- **Template — TERBUKTI:** YA — template Excel resmi dengan sheet contoh ("Example Sheet") sebagai referensi + instruksi 5–6 langkah di dalam template.
- **Panduan — TERBUKTI (UAB + template):** 6 langkah: nama kuis → tipe soal (dropdown) → soal → jawaban A–E → pilih jawaban benar (opsional, boleh kosong) → penjelasan (opsional). Hanya karakter alfanumerik di template (per instruksi template).
- **Pemetaan — TERBUKTI (kutipan Help Center via skill + UAB + template):** Kolom: tipe soal (teks eksak, mis. `Multiple choice` / `Open-ended`) → soal → Jawaban A–E (maks 5) → kolom kunci (satu atau lebih huruf jawaban; siswa harus memilih SEMUA yang ditandai agar benar) → penjelasan. Untuk soal terbuka, jawaban tidak diisi.
- **Validasi + error — TIDAK DIKETAHUI** (pesan error per baris tidak dirinci di sumber yang dibaca).
- **Preview — TIDAK DIKETAHUI** (tidak ada bukti dialog preview; kuis hasil import langsung tersedia di "My Quizzes" dan dapat diedit).
- **Batas — TERBUKTI (kutipan Help Center):** **Maks 100 soal per kuis hasil import** (template memakai baris 7–200 untuk soal; bila lebih, pecah lalu gabungkan). Ini batas import paling eksplisit kedua setelah ProProfs.
- **Gambar/media — TIDAK DIKETAHUI** untuk jalur template (template yang dikutip tidak menyebut kolom gambar; editor Socrative mendukung gambar per soal, tetapi apakah gambar bisa lewat template — TIDAK DIKETAHUI).
- **Gampang vs bingung (DISIMPULKAN):**
  - Gampang: dropdown tipe soal di template (tidak perlu hafal kode tipe seperti `pg`/`pg_kompleks` di v2); kunci huruf (konvensi umum); kunci boleh kosong (soal tanpa kunci tetap masuk — berguna untuk polling/diskusi); sheet contoh di dalam file yang sama.
  - Bingung: (1) **Kunci kosong = sah** — guru yang lupa mengisi kunci tidak mendapat peringatan (dari bukti yang dibaca), sehingga soal tanpa kunci bisa lolos ke kelas. (2) **Tipe soal harus teks eksak** (`Multiple choice`) — salah ketik/salah kapital dapat gagal atau salah tipe. (3) **Batas 100 soal** memaksa pecah-gabung untuk bank besar — langkah tambahan yang tidak ada di v2 (500).

### 8.4 Extraordinary CBT v4.7.0 — import (dari kajian, bukan uji live)

Sumber: `~/workspace/extraordinary-cbt/PELAJARI-EXTRAORDINARY-CBT.md` + ANALISIS.md Bab 13 (seperti laporan sebelumnya).

- **TERBUKTI (kajian):** Import bank soal; **Google Docs / Word (.docx) termasuk rumus**; import bank soal + upload CSV peserta. Itu seluruh rincian import yang ada di kajian.
- **TIDAK DIKETAHUI dari kajian:** nama kolom/format file import bank soal, apakah ada template unduhan, panduan pengisian, pemetaan kolom, validasi + pesan error, preview sebelum simpan, batas jumlah, penanganan gambar. Kajian tidak merinci mekanisme import bank soal sejauh yang dikutip di laporan sebelumnya. **Jangan** mengisi celah ini dengan asumsi dari v2 atau dari aplikasi lain.
- **DISIMPULKAN:** Satu-satunya keunggulan import Exo yang dapat diklaim dari bukti adalah dukungan `.docx` termasuk rumus + Google Docs — dua hal yang TIDAK dimiliki v2 (kode) saat ini. Selain itu, tidak ada dasar untuk membandingkan kualitas mekanisme import Exo vs v2.

---

## 9. Tabel perbandingan mekanisme import (fokus import saja)

Legenda: ✅ = didukung (TERBUKTI) · ⚠️ = didukung sebagian / dengan jebakan · ❌ = tidak ada (TERBUKTI tidak ada) · ? = TIDAK DIKETAHUI · — = tidak berlaku.

| Dimensi | Moodle (GIFT/Aiken/XML) | Quizizz/Wayground | Kahoot! | Exam.net | ProProfs | Google Forms | Respondus | Canvas (QTI) | Socrative | CBT SMK v2 (kode) |
|---|---|---|---|---|---|---|---|---|---|---|
| Format file import | `.txt` (GIFT/Aiken), `.xml` (XML), `.zip` (GIFT+media, plugin) | `.xlsx` | `.xlsx` | Tidak ada file (salin antar-ujian / tempel teks / PDF) | Excel; AI: `.docx`/`.pdf`/`.txt` | Tidak ada file (antar-Form) | `.txt`/`.rtf`/`.doc`/`.docx` | QTI `.zip` (via konverter CSV→QTI) | Template Excel | CSV teks via JSON |
| Template unduhan resmi | ❌ (contoh di dokumentasi; XML via ekspor contoh) | ✅ (Google Sheet) | ✅ (`.xlsx`) | — | ✅ (Excel contoh) | — | ✅ (Word contoh) | ❌ (konverter pihak ketiga) | ✅ (Excel + sheet contoh) | ❌ (hanya header + 1 contoh di placeholder) |
| Panduan pengisian | ✅ sangat lengkap (GIFT/XML); singkat (Aiken) | ⚠️ ringkas (11 kolom) | ✅ ringkas + eksplisit (batas + nilai waktu) | ⚠️ singkat, per jalur | ⚠️ ringkas; nama kolom tidak di teks | ✅ singkat (4 langkah) | ✅ sangat eksplisit + contoh per tipe | ⚠️ tata letak CSV konverter | ✅ 6 langkah di template | ⚠️ 1 kalimat + rujuk docs |
| Pemetaan kolom | Sintaks teks (GIFT/Aiken) / tag XML | 11 kolom (kunci = nomor) | Kolom (kunci = nomor; nama kolom tidak di teks) | — (salin utuh) | Field: soal/jawaban/tipe/poin/kategori (nama kolom ?) | — (salin utuh + validasi) | Sintaks teks (`*` kunci, `Type:`) | Positional CSV (A=tipe, C=poin, D=soal, E=kunci, F–J=opsi) | Kolom (kunci = huruf, boleh kosong) | Header by-name (kunci per tipe, 3 pemisah + `::`) |
| Validasi + pesan error | GIFT/Aiken: ketat, pesan ?; XML: parser tidak laporkan error (validasi eksternal wajib) | ? (gagal diam-diam bila kunci huruf) | ✅ cek otomatis; perbaiki / buang yang bermasalah | — | ? | — | ✅ Preview + warning berlapis | Header = error; lainnya ? | ? | ✅ error per baris bernomor + 2 mode |
| Preview sebelum simpan | ⚠️ tahap parse + Continue (isi ?) | ✅ review lalu Publish | ⚠️ soal masuk creator lalu edit | ✅ preview ujian/editor | ? | ✅ pilih soal sebelum import | ✅ Preview + warning | ? (verifikasi pasca-import) | ? | ❌ (hanya ringkasan + error sesudah kirim) |
| Batas jumlah | ? | ? | ? | — | ✅ 8 (Free); tak terbatas (berbayar, angka ?) | ? | ? | ? | ✅ 100 per kuis | ✅ 500 baris |
| Gambar/media | GIFT: via ZIP `@@PLUGINFILE@@` (plugin); XML: base64/URL; Aiken: ❌ | ⚠️ 1 gambar/soal via URL, teks polos saja | ❌ via spreadsheet | Soal salinan membawa media; tempel: ? | ? (Excel); AI: file dilampirkan | Salinan membawa media (DISIMPULKAN) | ✅ gambar Word ikut; `.txt` via folder | ? (jalur CSV) | ? | ❌ (`media_url` selalu kosong) |
| Tipe soal yang bisa diimpor | GIFT/XML: banyak; Aiken: PG saja | MC, Checkbox, Fill-in-Blank, Open-Ended, Poll, Draw | Quiz standar saja | — | ? (Excel); AI: inferensi | — (soal Forms apa pun) | PG, B/S, Esai, Isian, Multi-jawaban, Menjodohkan | MC, MR, TF (jalur CSV) | PG + terbuka | 16 tipe (kode) — paling banyak |
| Kunci ditegaskan guru vs inferensi | ✅ ditegaskan (sintaks/kolom) | ✅ ditegaskan (spreadsheet); ⚠️ inferensi (AI) | ✅ ditegaskan | — | ✅ ditegaskan (Excel); ⚠️ inferensi (AI) | ✅ disalin | ✅ ditegaskan (`*`) | ✅ ditegaskan | ✅ ditegaskan (boleh kosong) | ✅ ditegaskan |
| Kategori/kompetensi ikut terimpor | ✅ kategori (XML dummy kategori) | ❌ | ❌ | — | ✅ topik/kategori (Excel) | — | ❌ (judul/poin saja) | ⚠️ via bank tujuan | ❌ | ❌ (manual pasca-import) |

**DISIMPULKAN dari tabel — posisi v2:** v2 unggul di tiga hal yang TIDAK dimiliki pembanding mana pun sekaligus: (1) error per baris bernomor + dua mode gagal yang eksplisit, (2) 16 tipe soal dalam satu file, (3) batas 500 baris yang besar. v2 tertinggal di empat hal yang dimiliki hampir semua pembanding spreadsheet: (a) template unduhan resmi, (b) preview isi sebelum simpan, (c) kolom media, (d) kolom kategori/kompetensi. Dan v2 satu-satunya yang memakai tiga pemisah kunci dalam satu format — semua pembanding lain memakai maksimal satu pemisah kunci per format (koma) atau penanda tunggal (`*`, `=`).

---

## 10. Rekomendasi desain untuk pengganti import v2 (bukan perubahan kode — bahan keputusan)

Rekomendasi di bawah ini adalah DISIMPULKAN dari bukti perbandingan di atas. Tiap rekomendasi diberi alasan "dipelajari dari siapa" agar koordinator dapat menelusur balik.

1. **Sediakan template unduhan resmi (`.csv` + `.xlsx`) berisi header + minimal 1 baris contoh PER TIPE, bukan 1 baris PG.** Dipelajari dari Kahoot/Quizizz/ProProfs/Respondus/Socrative (semua punya template unduhan) dan dari kegagalan v2 saat ini (guru menyalin header dari kalimat UI). Template `.xlsx` penting karena guru SMK umumnya bekerja di Excel; konversi `.xlsx` → CSV saat ini dibebankan ke guru. Template harus menyertakan kolom yang saat ini hilang (lihat no. 3–4) agar guru tidak menemukan kolom baru setelah import.
2. **Tambahkan preview isi sebelum simpan (tabel soal ter-parse: teks, opsi, kunci terpetakan, bobot, peringatan).** Dipelajari dari Quizizz (review → Publish), Kahoot (cek error → perbaiki/buang), Respondus (Preview + warning berlapis). Preview menutup celah terbesar v2: kunci yang terpetakan salah (mis. salah pemisah) baru ketahuan setelah soal tersimpan. Mode `all_or_nothing` yang ada saat ini adalah validasi, bukan preview — jangan dianggap pengganti preview.
3. **Tambahkan kolom media (minimal: `media_url` / `gambar_url`) + dokumentasikan aturannya.** Dipelajari dari Quizizz (Image Link URL, 1 gambar/soal) dan Moodle XML (base64/URL) — keduanya membuktikan "URL gambar" adalah kompromi paling sederhana (tanpa upload ZIP). Aturan yang harus eksplisit: 1 gambar per soal atau lebih? URL eksternal saja atau upload juga? Apa yang terjadi bila URL mati (warning vs error)? Saat ini v2 memaksa `media_url` kosong — soal bergambar selalu dua tahap.
4. **Tambahkan kolom kategori / kisi-kisi / kompetensi (opsional) agar import selesai dalam satu langkah.** Dipelajari dari Moodle XML (dummy kategori dalam file) dan ProProfs (topik/kategori per soal). Tanpa kolom ini, import 500 soal menghasilkan 500 soal "tanpa alamat kurikulum" yang harus dipetakan satu per satu — mengalahkan tujuan massal. Kolom harus opsional (agar file lama tetap jalan) dan nilai yang tidak ditemukan harus menjadi error/warning bernomor baris, bukan diam-diam dibuang.
5. **Seragamkan pemisah kunci — satu pemisah untuk semua tipe, atau pisahkan kunci ke kolom sendiri per tipe.** Dipelajari dari kegagalan diam-diam Quizizz (kunci nomor vs huruf) dan dari kerapuhan positional Canvas (kolom bergeser = semua salah). Desain v2 saat ini (`,`, `;`, `\|`, `::` bergantung tipe) adalah yang paling rawan salah di antara semua pembanding. Opsi yang paling aman: kunci multi-nilai selalu koma + pasangan selalu `::` — dan itu harus tertulis di template + panduan, bukan hanya di docs. Bila format diubah, sediakan migrasi/kompatibilitas untuk CSV lama (v2 saat ini sudah kompatibel-mundur dengan CSV lama CBT-Go 4 kolom — pertahankan sifat itu).
6. **Pertahankan yang sudah unggul di v2 — jangan dibuang saat mengganti:** error per baris bernomor, dua mode (`all_or_nothing` / `skip_errors`), validasi ketat yang menolak (bukan menebak), audit log import, status draft hasil import, tolak import saat bank terkunci. Ini adalah praktik terbaik yang TIDAK dimiliki sebagian pembanding (Moodle XML tidak melaporkan error; Quizizz gagal diam-diam untuk kunci huruf; Socrative mengizinkan kunci kosong tanpa peringatan).
7. **Adopsi "warning berlapis" ala Respondus (error vs peringatan yang boleh diabaikan).** Contoh: kunci kosong untuk soal polling/diskusi = warning; kunci tidak cocok opsi = error; URL gambar mati = warning; tipe tidak dikenal = error. Saat ini v2 hanya punya dua tingkat (error / sukses) — warning membuat guru dapat memutuskan sendiri tanpa mengorbankan validasi ketat.
8. **Pertimbangkan jalur kedua "Word → soal" ala Respondus untuk guru yang sudah punya naskah Word — tetapi JANGAN meniru tanpa kontrak format ketat.** Dipelajari dari Respondus (format `*` + `Type:` + file contoh) dan dari rencana `.docx` v2 yang belum terwujud. Bila jalur Word ditambahkan, kontraknya harus: file contoh unduhan + preview + warning berlapis + dokumen tidak sesuai kontrak ditolak (prinsip yang sudah ada di spesifikasi v2). Jangan menambahkan jalur AI-inferensi kunci (ala ProProfs AI / Wayground AI) tanpa label jelas "kunci hasil inferensi — wajib ditinjau", karena kunci tebakan adalah risiko nilai yang salah.
9. **Dokumentasikan batas + perilaku batas di dalam UI import, bukan hanya di docs.** Dipelajari dari ProProfs (batas 8 soal Free tertulis di artikel) dan Socrative (batas 100 tertulis di template). Untuk v2: tulis "Maks 500 baris" + "Baris ke-501 ke atas ditolak" + "Bila file > 500, pecah menjadi beberapa file" langsung di panel import.
10. **Jangan meniru dari pembanding:** (a) Moodle XML — jangan mengandalkan validator eksternal; importer v2 harus melaporkan letak error sendiri. (b) Quizizz — jangan menerima nilai kunci yang tidak valid secara diam-diam; tolak dengan error bernomor baris. (c) Kahoot — jangan membatasi panjang soal/jawaban seketat 120/75 karakter untuk ujian sekolah (soal cerita SMK butuh lebih panjang). (d) Google Forms/Exam.net — jangan menganggap "salin antar-ujian" sebagai pengganti import file; guru SMK butuh jalur file dari Word/Excel.

---

## 11. Daftar sumber

| # | Sumber | URL | Dipakai untuk |
|---|---|---|---|
| 1 | MoodleDocs — GIFT format | https://docs.moodle.org/34/en/GIFT_format | GIFT: sintaks, contoh, UTF-8, escape, feedback, bobot |
| 2 | MoodleDocs — Aiken format | https://docs.moodle.org/36/en/Aiken_format | Aiken: aturan, kapital, UTF-8, opsi import, tanpa ekspor Aiken |
| 3 | MoodleDocs — Moodle XML format | https://docs.moodle.org/39/en/Moodle_XML_format | XML: struktur, tag, fraction, kategori, CDATA, base64, validator |
| 4 | Clarkson BookStack — Quiz: Prepare a Word Quiz for Import into Moodle (Aiken) | https://bookstack.clarkson.edu/books/moodle/page/quiz-prepare-a-word-quiz-for-import-into-the-moodle-question-bank | Aiken: alur import, review errors + Continue, gambar manual pasca-import |
| 5 | Ulethbridge — Importing Aiken Formatted Questions | https://sites.ulethbridge.ca/moodle-answers/2024/07/12/importing-aiken-formatted-questions-2/ | Aiken: satu baris per soal, ANSWER kapital, `.txt` tanpa gambar |
| 6 | Bethel University — How to Import Questions from a Text File | https://betheluniversity.atlassian.net/wiki/spaces/MOODLE/pages/52073285/How+to+Import+Questions+from+a+Text+File | Aiken: contoh, smartquotes, pemisah soal |
| 7 | BaL Digital (Portsmouth) — Import Quiz Questions to Moodle Using the Aiken Format | https://baldigital.port.ac.uk/import-quiz-questions-to-moodle-using-the-aiken-format/ | Aiken: alur kategori + import |
| 8 | arXiv 2010.15577 — Import test questions into Moodle LMS | https://arxiv.org/pdf/2010.15577 | Perbandingan Aiken/GIFT/XML, maks 10 opsi Aiken, GIFT+media ZIP |
| 9 | GIFT Format (COTC PDF) | https://bpb-us-e1.wpmucdn.com/u.cotc.edu/dist/3/152/files/2020/03/GIFT-Format.pdf | GIFT: alur import, contoh |
| 10 | VidQuiz — Import Questions Into Quizizz and Wayground From Excel | https://vidquiz.com/blog/how-to-import-questions-into-quizizz | Quizizz: 11 kolom, kunci nomor, `.xlsx` saja, teks polos, 1 gambar via URL, review, jalur Forms/AI |
| 11 | Monsha — How to Import Questions From a Spreadsheet Into Quizizz | https://monsha.ai/blog/how-to-import-questions-from-a-spreadsheet-into-quizizz | Quizizz: template resmi, alur import |
| 12 | Kahoot! Help — How to import questions from a spreadsheet to your kahoot | https://support.kahoot.com/hc/en-us/articles/115002812547-How-to-import-questions-from-a-spreadsheet-to-your-kahoot | Kahoot: template, 120/75 karakter, waktu 5–120 dtk, kunci nomor, cek error, tanpa gambar, hanya Quiz |
| 13 | Exam.net Support — Create an exam | https://support.exam.net/s/article/create-an-exam | Exam.net: 3 format ujian, konverter PDF, tulis/tempel |
| 14 | Exam.net Support — The exam format "Use question types" | https://support.exam.net/s/article/the-exam-format-use-question-types?language=en_US | Exam.net: tipe auto-marked, salin soal antar-ujian, preview marking |
| 15 | Exam.net Support — The exam format "Create document" | https://support.exam.net/s/article/the-exam-format-create-document | Exam.net: tulis/tempel dari Word, PDF Resources, preview, tanpa bagi-bagian |
| 16 | ProProfs Help — How to Bulk-Import Questions From Excel | https://quiz.proprofs.com/how-to-bulk-import-questions-from-an-excel-document-to-your-quiz | ProProfs: Excel contoh, poin + kategori, batas 8 (Free), jalur AI (`.docx`/`.pdf`/`.txt`, 20 MB) |
| 17 | Guiding Tech — How to Import Questions Into Google Forms Easily | https://www.guidingtech.com/import-questions-into-google-forms/ | Google Forms: import antar-Form, centang soal, validasi ikut, satu form per giliran |
| 18 | UConn KB — Formatting Word Documents for Respondus | https://kb.uconn.edu/space/TL/26024247323 | Respondus: aturan wajib, `*` kunci, `Title:`, daftar jawaban akhir, True di atas False |
| 19 | SIUE KB — Blackboard (Faculty) - Respondus: Question Format | https://kb.siue.edu/page.php?id=62185 | Respondus: `.doc`/`.docx`/`.rtf`/`.txt`, `.txt` paling bersih, auto-numbering Word |
| 20 | Lone Star VTAC — Format for Importing Multiple Choice Questions into Respondus | http://vtac.lonestar.edu/help/a195 | Respondus: format standar, `*`, judul/ feedback / poin opsional |
| 21 | Terry/UGA — Respondus Exam Format Guide | https://goterry.atlassian.net/wiki/spaces/BC/pages/1030717441/Respondus+Exam+Format+Guide | Respondus: `Type: E/F/MT`, pasangan `=`, isian multi-jawaban |
| 22 | Groningen/Nestor — Respondus (Confluence) | https://nestorsupport.atlassian.net/wiki/spaces/NH/pages/259588127/Respondus | Respondus: Preview, warning judul duplikat vs tanpa jawaban benar |
| 23 | Respondus User Guide (Docsbay) | https://docsbay.net/respondus-user-guide | Respondus: gambar Word ikut, `.txt`/`.rtf` via folder gambar, judul 20 karakter |
| 24 | UH Instruction — How to Import Quiz Questions in Word with Respondus | https://www.instruction.uh.edu/knowledgebase/how-to-import-questions-with-respondus/ | Respondus: alur Preview → Finish → Publish ke Canvas |
| 25 | ASEE 2024 — Best Practices for Using Generative AI to Create Quiz Content for Canvas LMS | https://peer.asee.org/best-practices-for-using-generative-ai-to-create-quiz-content-for-the-canvas-lms.pdf | Canvas QTI: CSV tanpa header, kolom A/B/C/D/E/F–J, konverter K-State, import ZIP |
| 26 | text2qti (PyPI) | https://pypi.org/project/text-to-qti/ | Canvas QTI: format teks `*`, `[*]`, `=`, front matter YAML |
| 27 | Socrative Help Center — Import a Quiz (via kutipan skill templating) | https://help.socrative.com/en/articles/2155335-import-a-quiz | Socrative: template Excel, maks 100 soal, kunci huruf multi-benar, penjelasan |
| 28 | UAB — Technology in the Classroom: Socrative | https://www.uab.edu/icac/images/How-to_Guides/Socrative_How-to_Guide.pdf | Socrative: 6 langkah template, jawaban A–E, kunci opsional |
| 29 | Laporan pembanding sebelumnya (bahan awal) | `~/workspace/cbt-smk-v2/docs/LAPORAN-TIM-PEMBANDING-LUAR.md` | Titik awal dimensi import 7 aplikasi + kolom Exo/v2 spesifikasi |
| 30 | Kode import v2 (baseline faktual) | `~/workspace/cbt-smk-v2/backend/internal/banksoal/banksoal.go` + `~/workspace/cbt-smk-v2/frontend/src/pages/admin/banksoal/BankSoalDetailPage.tsx` | Baseline v2: CSV, 500 baris, mode, error per baris, tanpa template/preview/media |

---

## 12. Penutup (untuk koordinator)

- **Yang sudah terjawab tuntas dari laporan sebelumnya:** import Quizizz (11 kolom, kunci nomor, `.xlsx`), Kahoot (template resmi, batas karakter, cek error, tanpa gambar), Exam.net (TIDAK ada import file — hanya salin antar-ujian + tempel teks + PDF), ProProfs (Excel contoh + poin/kategori + batas 8 Free + jalur AI).
- **Yang masih TIDAK DIKETAHUI dan tidak boleh dikarang:** pesan error persis per baris di Quizizz/ProProfs/Socrative; nama kolom persis ProProfs dan Kahoot (teks artikel tidak mencantumkannya); batas soal per file di Moodle/Quizizz/Kahoot/Respondus/Canvas; mekanisme import bank soal Exo selain "`.docx` termasuk rumus + Google Docs".
- **Satu koreksi faktual terhadap laporan sebelumnya:** import `.docx` v2 yang disebut di laporan sebelumnya adalah rencana (ANALISIS.md), bukan fitur — kode yang ada hanya mengimplementasikan CSV. Bila koordinator memakai laporan ini sebagai dasar desain pengganti import, baseline yang benar adalah Bab 1 (kode), bukan Bab 10 laporan sebelumnya (spesifikasi).
- **Tidak ada kode yang diubah.** Laporan ini hanya analisis.
