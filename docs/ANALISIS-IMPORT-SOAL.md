# ANALISIS IMPORT SOAL — Sintesis Pembanding + Spesifikasi Import Excel CBT SMK v2

**Dokumen:** `docs/ANALISIS-IMPORT-SOAL.md`
**Tanggal:** 16 Oktober 2026
**Penyusun:** Tim Sintesis → Spesifikasi Import Kita (kerja mandiri — tanpa subagent)
**Status:** ANALISIS + SPESIFIKASI SAJA. Tidak ada kode yang diubah.
**Tujuan:** Satu acuan tunggal untuk tugas berikutnya (membangun import Excel). Spesifikasi di Bab 2 ditulis agar bisa langsung diimplementasikan.

## Cara membaca dokumen ini (aturan bukti)

Setiap klaim memakai salah satu label:

- **TERBUKTI** — ada di sumber yang dikutip (nama dokumen + bab/kode/situs).
- **DISIMPULKAN** — inferensi logis dari bukti yang dikutip; alasannya ditulis.
- **TIDAK DIKETAHUI** — tidak ditemukan buktinya; jangan dianggap ada.
- **KEPUTUSAN SPESIFIKASI** — bukan klaim tentang dunia, melainkan pilihan desain yang ditetapkan dokumen ini untuk v2. Bila bertentangan dengan perilaku lama, perbedaannya ditulis eksplisit.

**Dokumen ini boleh MENYEBUT nama aplikasi lain untuk perbandingan faktual. Tidak ada aset, logo, branding, atau teks berhak cipta pihak lain yang disalin ke sini.** Contoh sintaks yang ditampilkan adalah parafrase/ringkasan format yang didokumentasikan publik.

## Sumber

| # | Sumber | Dipakai untuk |
|---|---|---|
| 1 | `docs/LAPORAN-TIM-EXO-IMPORT.md` (Tim Exo Import, 16 Okt 2026) | Semua klaim tentang Extraordinary CBT (Exo) di dokumen ini. **Menggantikan** ringkasan Exo di laporan pembanding lama (§8.4) yang hanya bersumber dari kajian lama dan kini terbukti tidak lengkap. |
| 2 | `docs/LAPORAN-TIM-PEMBANDING-IMPORT.md` (Tim Pembanding Import, 16 Okt 2026) | Semua klaim tentang Moodle, Quizizz/Wayground, Kahoot, Exam.net, ProProfs, Google Forms, Respondus, Canvas QTI, Socrative, dan baseline import v2. |
| 3 | `ANALISIS.md` (khususnya Bab 5 Bank Soal, Bab 6 Tipe Soal, Bab 5.5 Import) | Konteks v2: 5 tipe inti AKM + 11 varian, atribut soal, prinsip import. |
| 4 | `docs/TIPE-SOAL.md` + `backend/internal/banksoal/types.go` + `backend/internal/banksoal/banksoal.go` (dibaca untuk verifikasi, tidak diubah) | Kebenaran kode: 16 kode tipe, format kunci per tipe, perilaku parser CSV lama, kolom tabel `soal`. Dipakai untuk meng-grounding spesifikasi agar tidak mengarang nama tipe/kolom. |
| 5 | `ANALISIS.md` Bab 5.5 (spesifikasi CSV usulan) | Dicatat sebagai **usulan yang tidak pernah diimplementasikan** — lihat §2.10. Bukan baseline. |

**Koreksi penting terhadap laporan sebelumnya (TERBUKTI dari kode):** ANALISIS.md Bab 5.5 menyebut v2 punya "Import Word (.docx)". Di kode yang ada, **TIDAK ditemukan endpoint import .docx** — yang diimplementasikan hanya CSV. Baseline yang benar adalah CSV lama (§2.10); klaim `.docx` diperlakukan sebagai rencana yang belum terwujud. Hal yang sama berlaku untuk daftar kolom CSV di ANALISIS.md Bab 5.5 (`kode_bank, nomor_urut, …`) — daftar itu **tidak pernah diimplementasikan**; kolom CSV yang benar-benar jalan adalah daftar di §2.10.

**Label jujur yang berlaku di seluruh dokumen ini (TERBUKTI, ANALISIS.md Bab 6 + `docs/TIPE-SOAL.md`):** CBT SMK v2 mendukung **"5 tipe standar AKM + 11 varian/ekstensi aplikasi"**. Jangan pernah menulis "16 tipe standar Kemendikbud" atau "16 tipe import standar" di dokumentasi/marketing v2.

---

## Bab 1 — Tabel perbandingan mekanisme import semua aplikasi

Legenda: ✅ = didukung (TERBUKTI) · ⚠️ = didukung sebagian / dengan jebakan · ❌ = tidak ada (TERBUKTI tidak ada) · ? = TIDAK DIKETAHUI · — = tidak berlaku.

### 1.1 Format file yang diterima

| Aplikasi | Format import soal |
|---|---|
| **Extraordinary CBT (Exo)** | Word `.docx` Format 1 (gaya baris `S:` / `A:)` / `JAWAB:`) ✅ · Word `.docx` Format 2 (tabel 2 kolom per soal, marker `TS/KD/KJ/ABS`) ✅ · HTML Zip Format 2 ✅ · Excel `.xlsx` 8 kolom (hanya PG + Esai dicontohkan) ✅ · Word gaya `##...##` + `***` ✅ sebagai file template bawaan, pemakaian di dialog import TIDAK DIKETAHUI · Google Docs: hanya klaim marketing + konfigurasi env, mekanisme TIDAK DIKETAHUI · CSV untuk soal: TIDAK ADA BUKTI · PDF/TXT: TIDAK ADA BUKTI |
| **Moodle** | GIFT `.txt` ✅ · Aiken `.txt` ✅ · Moodle XML `.xml` ✅ · GIFT + media sebagai arsip `.zip` (plugin pihak ketiga, bukan format inti) ⚠️ · Blackboard/WebCT/dll. ada tetapi TIDAK DIDALAMI (detail TIDAK DIKETAHUI) |
| **Quizizz / Wayground** | `.xlsx` ✅ (template resmi berupa Google Sheet). CSV TIDAK dibaca importer — guru harus memindahkan kolom ke template lalu simpan `.xlsx` |
| **Kahoot!** | `.xlsx` ✅ (wajib memakai template resmi; Excel 2007+) |
| **Exam.net** | ❌ Tidak ada import file soal massal. Yang ada: salin soal dari ujian lain di platform ("Import questions" = copy antar-ujian), tempel teks dari Word ke editor dokumen, unggah PDF sebagai naskah |
| **ProProfs** | Excel (`.xls`/`.xlsx`) ✅ · Jalur AI: `.docx` / `.pdf` / `.txt` maks 20 MB (kunci diinferensi AI, bukan ditegaskan guru) |
| **Google Forms** | ❌ Tidak ada import file soal. Yang ada: import soal dari Form lain (antar-Form). Import dari dokumen lain hanya via add-on pihak ketiga (bukan bawaan) |
| **Respondus** | `.txt` / `.rtf` / `.doc` / `.docx` ✅ (file `.txt` disebut "paling bersih"; Word kadang merusak auto-numbering / menghapus asterisk) |
| **Canvas LMS (QTI)** | Paket QTI `.zip` ✅ (berisi XML QTI). Jalan masuk spreadsheet: konverter CSV → QTI pihak ketiga (K-State), lalu ZIP diimpor tanpa diekstrak |
| **Socrative** | Template Excel (`.xls`/`.xlsx`) ✅ · Jalur kedua: import kuis dari guru lain lewat kode SOC (`SOC-XXXXX`) |
| **CBT SMK v2 saat ini (kode)** | CSV teks yang dikirim sebagai JSON (`csv_text` / `csv` + `bank_id` + `mode`) ✅. Bukan upload file biner, bukan `.xlsx`, bukan `.docx` |

### 1.2 Template unduhan

| Aplikasi | Template unduhan |
|---|---|
| **Exo** | ✅ Template bawaan di paket unduhan (`storage/exo-templates/`: Format 1, Format 2, `##/***` PG + Esai, Excel, template kompetensi). Dari dialog Import bisa unduh contoh ("Download format tipe file 2" — TERBUKTI pihak ketiga). Apakah template bisa diunduh dari situs publik tanpa login: TIDAK DIKETAHUI |
| **Moodle** | ❌ Tidak ada file template unduhan resmi di halaman GIFT/Aiken/XML. Yang ada: contoh di dokumentasi (GIFT), dua contoh soal (Aiken), dan "template by example" lewat ekspor soal contoh (XML) |
| **Quizizz** | ✅ Template resmi (Google Sheet) + tombol unduh template di alur import |
| **Kahoot!** | ✅ Tautan "Download the official Kahoot! template" di artikel + di dialog import |
| **Exam.net** | — (tidak ada file; editor adalah kanvas kosong / daftar ujian sumber) |
| **ProProfs** | ✅ File Excel contoh dapat diunduh dari alur import dan artikel. Instruksi: "ganti soal/opsi/detail dummy dengan data asli" |
| **Google Forms** | — (tidak ada file) |
| **Respondus** | ✅ File contoh Word dapat diunduh + panduan format |
| **Canvas QTI** | ❌ Tidak ada template `.zip` resmi satu klik; bergantung konverter pihak ketiga |
| **Socrative** | ✅ Template Excel resmi dengan sheet contoh ("Example Sheet") + instruksi 5–6 langkah di dalam template |
| **CBT SMK v2 saat ini** | ❌ TIDAK ada tombol unduh template. Hanya satu baris teks nama kolom + placeholder textarea berisi header + 1 baris contoh PG. Guru harus mengetik/menyalin header sendiri |

### 1.3 Panduan pengisian

| Aplikasi | Panduan |
|---|---|
| **Exo** | ⚠️ Tersebar: instruksi terbaik tertanam di dalam template Format 2 (blok "Perhatian" + penjelasan semantik kunci per tipe) + komentar developer di file kompetensi + video resmi (isi TIDAK DIVERIFIKASI) + tutorial audio in-app (isi tidak ditranskripsikan). Tidak ada satu halaman manual import resmi yang lengkap dan terpusat |
| **Moodle** | ✅ Sangat lengkap (GIFT: tabel simbol + contoh per tipe + opsi lanjutan; XML: lengkap per tipe). Aiken: singkat (satu halaman) |
| **Quizizz** | ⚠️ Ringkas (satu tabel 11 kolom + contoh baris di template) |
| **Kahoot!** | ✅ Ringkas + eksplisit: 6 langkah + FAQ, batas karakter, nilai waktu yang diizinkan, cara menandai jawaban benar. Nama kolom persis TIDAK tercantum dalam teks artikel (hanya terlihat di gambar template) |
| **Exam.net** | ⚠️ Singkat per jalur (3 langkah salin; tulis/tempel → preview → konfigurasi) |
| **ProProfs** | ⚠️ Ringkas: Excel dapat diisi soal + jawaban + tipe + poin per soal + topik/kategori per soal. Nama kolom persis + contoh isi sel per tipe TIDAK DIKETAHUI (gambar contoh tidak terbaca teksnya) |
| **Google Forms** | ✅ Singkat (4 langkah: pilih form sumber → centang soal → Import) |
| **Respondus** | ✅ Sangat eksplisit: aturan wajib + opsional + contoh per tipe (nomor + `.`/`)`, huruf jawaban, `*` kunci, `Type:`, `Title:`) |
| **Canvas QTI** | ⚠️ Tata letak CSV konverter: tanpa baris header (header = error); Kolom A = tipe, B = kosong, C = poin, D = soal, E = kunci, F–J = opsi |
| **Socrative** | ✅ 6 langkah di dalam template: nama kuis → tipe soal (dropdown) → soal → jawaban A–E → kunci (boleh kosong) → penjelasan |
| **CBT SMK v2 saat ini** | ⚠️ Satu kalimat ("Format kunci per tipe: lihat docs/TIPE-SOAL.md") + placeholder 1 baris PG. Tidak ada panduan per tipe di UI import; tidak ada contoh baris untuk 15 tipe lain |

### 1.4 Pemetaan kolom / sintaks → field soal

| Aplikasi | Pemetaan |
|---|---|
| **Exo** | Format 2: marker `TS` (kode tipe) / `KD` (kode kompetensi) / `KJ` (kunci, format bergantung tipe) / `ABS` (flag, arti DISIMPULKAN = penilaian absolut, belum TERBUKTI) + nomor soal = pertanyaan + baris opsi. Format 1: `S:<n>)` soal, `A:)`…`E:)` opsi, `JAWAB:` kunci. Excel: 8 kolom (`SOAL`, `Tipe Soal` angka, `Kunci Jawaban`, `Opsi 1`–`Opsi 5`). **Tidak ada** field pembahasan, bobot per soal, level, atau stimulus per soal di template import maupun tabel `soals` Exo (TERBUKTI dari skema) |
| **Moodle** | GIFT: sintaks teks (`::Judul::`, `{ }`, `=` benar, `~` salah, `#` feedback, `%n%` bobot, `->` menjodohkan, `//` komentar). Aiken: baris soal + baris `A. …` + baris `ANSWER: D` (semua kapital, soal satu baris). XML: tag (`<answer fraction="100|0">`, `<subquestion>`, dummy `<question type="category">`, CDATA untuk HTML) |
| **Quizizz** | 11 kolom: Question Text (wajib), Question Type (kosong = Multiple Choice), Option 1–5, **Correct Answer = NOMOR opsi 1–5** (Checkbox: beberapa nomor dipisah koma), Time in seconds (kosong = 30), Image Link, Answer explanation. Semantik kolom berubah per tipe (Option = pilihan untuk PG, = jawaban diterima untuk Fill-in-the-Blank; Correct Answer dikosongkan untuk Fill-in-the-Blank) |
| **Kahoot!** | Kolom template (nama kolom persis TIDAK DIKETAHUI dari teks artikel). Jawaban benar = **nomor jawaban dipisah koma** (`1`, `2,3`). Hanya tipe Quiz standar |
| **Exam.net** | — (salin utuh: teks + tipe + aturan marking; tidak ada pemetaan kolom) |
| **ProProfs** | Nama kolom persis TIDAK DIKETAHUI. Kelompok field yang TERBUKTI didukung: teks soal, jawaban, tipe soal, poin per soal, topik/kategori per soal |
| **Google Forms** | — (salin utuh dari form sumber, termasuk validasi) |
| **Respondus** | Kunci: asterisk `*` tepat sebelum huruf jawaban (`*b. …`) atau daftar jawaban di akhir file. Esai: `Type: E`. Isian: `Type: F`. Menjodohkan: `Type: MT` + pasangan `kiri=kanan`. Opsional: `Title:`, feedback, poin. Jebakan: Benar-Salah — `True` harus di atas `False`, bila terbalik soal diimpor sebagai PG |
| **Canvas QTI** | Positional CSV konverter: A = tipe (MC/MR/TF), B = kosong (wajib ada), C = poin, D = soal, E = kunci, F–J = opsi. text2qti: `*a)` benar, `[*]` multi-benar, `= 2` isian, front matter YAML |
| **Socrative** | Kolom: tipe soal (teks eksak, mis. `Multiple choice` / `Open-ended`, via dropdown di template) → soal → Jawaban A–E (maks 5) → kunci (satu atau lebih **huruf**; siswa harus memilih SEMUA yang ditandai) → penjelasan. Kunci boleh kosong |
| **CBT SMK v2 saat ini** | Header by-name (case-insensitive, trim): wajib `tipe`, `pertanyaan`, `kunci`; opsional `opsi_a`…`opsi_e`, `bobot` (default 1, bulat > 0), `level` (default `sedang`), `stimulus`, `pembahasan`. Satu baris = satu soal. **Tidak ada** kolom kisi-kisi/kompetensi, kategori, nama soal, feedback per opsi, atau media. **Tiga pemisah kunci + satu pemisah pasangan** dalam satu format: `,` (multi-label, benar-salah, urutan), `;` (antar-pasangan menjodohkan), `\|` (alternatif isian / jawaban per kolom / per rumpang), `::` (pasangan kiri-kanan) |

### 1.5 Validasi + pesan error

| Aplikasi | Validasi + error |
|---|---|
| **Exo** | Aturan tertulis di template Format 2: dilarang tabel selain format Exo; dilarang floating image; `TS` wajib; PG: `KJ` harus merujuk opsi yang ada. Riwayat bug: pengecoh JDX pernah hilang saat import ZIP; gambar `.docx` pernah gagal masuk filemedia; urutan import pernah salah. **TIDAK DIKETAHUI:** teks pesan error yang sesungguhnya; perilaku saat validasi gagal (tolak semua vs lewati baris); validasi jumlah opsi minimum / duplikasi label / `KD` tak terdaftar; `KJ` tak cocok opsi ditolak atau disimpan |
| **Moodle** | GIFT/Aiken: ketat, tetapi pesan error per baris TIDAK DIKETAHUI dari sumber yang dibaca. Aiken: huruf jawaban + `ANSWER` harus kapital (bila tidak, import gagal); smart quotes / karakter non-ASCII dapat menyebabkan error; opsi `Stop on error [Yes/No]`. XML: parser **mengasumsikan file well-formed dan tidak mendeteksi/melaporkan error** — guru harus memvalidasi sendiri lewat validator eksternal |
| **Quizizz** | Pesan error per baris TIDAK DIKETAHUI. **Jebakan TERBUKTI:** spreadsheet berisi huruf A/B/C/D di kolom Correct Answer **terunggah tanpa error** tetapi menghasilkan kunci yang salah pada semua soal ("fails quietly") — kegagalan validasi diam-diam |
| **Kahoot!** | ✅ "Sistem akan otomatis memeriksa error. Bila ditemukan, Anda bisa kembali memperbaiki spreadsheet atau lanjut dan membuang soal yang bermasalah." Model skip_errors dengan keputusan di tangan guru per batch. Apakah pesan error menyebut nomor baris / nama kolom: TIDAK DIKETAHUI |
| **Exam.net** | — (tidak ada validasi file karena tidak ada file) |
| **ProProfs** | Pesan error / error bernomor baris TIDAK DIKETAHUI dari artikel yang dibaca |
| **Google Forms** | — (tidak ada validasi file) |
| **Respondus** | ✅ Tombol **Preview** sebelum Finish menampilkan **warnings/notes berlapis**: judul duplikat (boleh diabaikan — dari judul otomatis 20 karakter pertama) vs "soal tanpa jawaban benar" (JANGAN diabaikan) |
| **Canvas QTI** | Header di baris atas = error (aturan eksplisit). Pesan error lain / error bernomor baris TIDAK DIKETAHUI |
| **Socrative** | Pesan error per baris TIDAK DIKETAHUI. **Jebakan:** kunci kosong = sah (soal tanpa kunci tetap masuk, tanpa peringatan dari bukti yang dibaca) |
| **CBT SMK v2 saat ini** | ✅ Validasi per baris: tipe dikenal, pertanyaan terisi, kunci terisi, bobot bulat > 0, level valid, aturan per tipe (minimal 2 opsi, kunci cocok label opsi, pasangan `kiri::kanan`, jumlah rumpang = jumlah kunci), plus `ValidateSoal`. Setiap error = `{baris, pesan}` dengan nomor baris file (header = baris 1, data mulai baris 2). Error level file: CSV tak terbaca, tanpa header + 1 baris, kolom wajib hilang, > 500 baris. Header tidak sensitif kapital. Baris kosong dilewati. Dua mode eksplisit: `all_or_nothing` (default; ada error apa pun → `sukses: 0`, `disimpan: false`, tidak ada yang tersimpan) dan `skip_errors` (baris valid disimpan, baris gagal dilaporkan). Mode tidak valid → 400. Bank terkunci → 409, import ditolak |

### 1.6 Preview sebelum simpan

| Aplikasi | Preview |
|---|---|
| **Exo** | ❌ Tidak ada bukti pratinjau / ringkasan / konfirmasi pra-simpan. Yang TERBUKTI: tombol **"Preview Bank Soal" sesudah import** (verifikasi pasca-import) + pratinjau soal menjodohkan di admin. DISIMPULKAN: alur import tampaknya langsung (pilih format → Upload → tersimpan) |
| **Moodle** | ⚠️ Tahap parse + "review for errors … Continue" (DISIMPULKAN dari panduan institusi). Apakah preview menampilkan isi soal per soal atau hanya hitungan: TIDAK DIKETAHUI. Jangan samakan dengan preview isi penuh |
| **Quizizz** | ✅ Setelah unggah, soal yang terbaca ditampilkan untuk ditinjau/diedit, baru Publish (preview isi) |
| **Kahoot!** | ⚠️ Setelah import berhasil, soal "langsung muncul di kahoot creator" — guru dapat melihat/mengedit di creator. DISIMPULKAN: preview pasca-import di editor (soal sudah masuk draf), bukan dialog preview khusus sebelum simpan; tetapi ada tahap cek error sebelum masuk |
| **Exam.net** | ✅ Preview ujian di editor (ikon kaca pembesar) + uji marking — berlaku untuk seluruh ujian, bukan khusus dialog import |
| **ProProfs** | ? Artikel menunjukkan soal muncul sebagai soal biasa setelah unggah; ada/tidak dialog preview/konfirmasi sebelum masuk: TIDAK DIKETAHUI |
| **Google Forms** | ✅ Daftar soal form sumber ditampilkan untuk dipilih (centang per soal) sebelum import — preview-pilih, bukan preview isi hasil parsing |
| **Respondus** | ✅ Preview sebelum Finish + warning berlapis (lihat §1.5) |
| **Canvas QTI** | ? ASEE menyarankan verifikasi soal di Canvas setelah import (pasca-import) |
| **Socrative** | ? Tidak ada bukti dialog preview; kuis hasil import langsung tersedia di "My Quizzes" dan dapat diedit |
| **CBT SMK v2 saat ini** | ❌ TIDAK ada preview per soal sebelum simpan. Alur: kirim → server validasi → langsung simpan (`skip_errors`) atau tolak semua (`all_or_nothing`). Yang ditampilkan sesudahnya hanya ringkasan `Sukses / Gagal / Disimpan` + daftar error per baris. Mode `all_or_nothing` adalah dry-run validasi, bukan pratinjau isi |

### 1.7 Batas jumlah / ukuran

| Aplikasi | Batas |
|---|---|
| **Exo** | TIDAK DIKETAHUI. Tidak ada angka batas maksimum di sumber yang diperiksa. Jangan keliru: template Excel 50 baris slot dan template Word Format 2 berisi 10 soal contoh adalah ukuran template contoh, BUKAN bukti batas engine |
| **Moodle** | TIDAK DIKETAHUI (tidak ada batas soal per file di halaman yang dibaca; Aiken: maks 10 opsi A–J per soal — itu batas opsi, bukan batas soal) |
| **Quizizz** | TIDAK DIKETAHUI (batas produk lain yang disebut — sesi live Basic 30 peserta — bukan batas import) |
| **Kahoot!** | TIDAK DIKETAHUI (batas soal per file tidak ada di artikel). Batas lain yang TERBUKTI: soal maks 120 karakter, jawaban maks 75 karakter, minimal 2 jawaban per soal, time limit hanya 5/10/20/30/60/120 detik |
| **Exam.net** | — |
| **ProProfs** | ✅ Paket Forever Free: import Excel **maks 8 soal**; berbayar "unlimited" (angka pastinya TIDAK DIKETAHUI). Jalur AI: file maks 20 MB |
| **Google Forms** | TIDAK DIKETAHUI (tidak ada angka batas soal per giliran) |
| **Respondus** | TIDAK DIKETAHUI (batas soal per file tidak disebut) |
| **Canvas QTI** | TIDAK DIKETAHUI |
| **Socrative** | ✅ **Maks 100 soal per kuis hasil import** (template memakai baris 7–200; bila lebih, pecah lalu gabungkan) |
| **CBT SMK v2 saat ini** | ✅ Maks **500 baris** data per berkas (dihitung sebelum validasi per baris; bila > 500, seluruh file ditolak dengan satu error `baris: 0`). Tidak ada batas ukuran byte eksplisit di parser |

### 1.8 Gambar / media

| Aplikasi | Media dalam import |
|---|---|
| **Exo** | ✅ Gambar inline di `.docx` ikut terimpor (stimulus sebelum pertanyaan; gambar per opsi; gambar di sel tabel Format 2). Floating image DILARANG. HTML Zip: file HTML + folder gambar di-zip, diunggah manual lewat file manager lalu dipilih di dialog import. Audio/video: TIDAK ADA BUKTI bisa lewat file import — koroborasi sekunder: audio/video diedit manual pasca-import di menu Daftar Soal. Rumus Word (OMML): TIDAK DIKETAHUI cara parser menanganinya; satu-satunya bukti adalah judul video resmi (isi tidak diverifikasi) |
| **Moodle** | GIFT inti: ❌ tidak membawa file gambar; jalur lanjutan via tag `<img>` + `@@PLUGINFILE@@/...` + arsip ZIP (plugin pihak ketiga, rawan salah path). XML: ✅ `<image>` berisi URL / `<image_base64>` (file membengkak, tidak ramah tulis tangan — realistisnya jalur ekspor). Aiken: ❌ `.txt` tidak bisa membawa gambar; gambar ditambahkan manual per soal sesudah import |
| **Quizizz** | ⚠️ Satu gambar per soal lewat kolom Image Link berupa URL (gambar harus sudah ter-host). Hanya teks polos yang diimpor — bold/italic/warna di-strip. Rumus harus dibangun ulang manual. Gambar tertanam di file `.xlsx` tidak ikut |
| **Kahoot!** | ❌ Gambar dan karakter khusus tidak dapat diimpor lewat spreadsheet. Simbol matematika disalin-tempel ke creator sesudahnya (langsung atau sebagai kode LaTeX) |
| **Exam.net** | Soal hasil salinan membawa medianya (karena sudah ada di ujian sumber). Soal dari tempelan teks: gambar ikut atau tidak TIDAK DIKETAHUI. Soal dapat berisi teks, gambar, rumus, tabel (diisi di editor) |
| **ProProfs** | Jalur Excel: TIDAK DIKETAHUI apakah ada kolom gambar/media. Jalur AI: file dilampirkan ke editor AI |
| **Google Forms** | Salinan membawa media yang sudah ada di soal sumber (DISIMPULKAN dari sifat salin utuh). Tidak ada jalur "unggah file berisi soal bergambar" |
| **Respondus** | ✅ Gambar tertanam di Word (`.doc`/`.docx`) ikut terimpor otomatis. Untuk `.txt`/`.rtf`/CSV, gambar lewat metode terpisah (menunjuk folder gambar). Format kaya (bold/italic/underline/superscript/subscript) dipertahankan untuk `.rtf`/Word; `.txt` kehilangan format |
| **Canvas QTI** | Jalur CSV→QTI yang didokumentasikan: TIDAK DIKETAHUI (tidak menyebut kolom gambar). QTI sebagai standar mendukung media, tetapi jangan klaim untuk jalur CSV |
| **Socrative** | TIDAK DIKETAHUI untuk jalur template (template yang dikutip tidak menyebut kolom gambar; editor mendukung gambar per soal) |
| **CBT SMK v2 saat ini** | ❌ TIDAK ADA. `INSERT` import selalu mengisi `media_url` dengan string kosong (`""`); tidak ada kolom untuk media, tidak ada upload file pendamping, tidak ada kolom URL gambar. Soal bergambar tidak bisa diimpor |

### 1.9 Penanganan duplikat

| Aplikasi | Duplikat |
|---|---|
| **Exo** | TIDAK DIKETAHUI apakah aplikasi memeriksa duplikat saat import. TERBUKTI dari skema: tabel `soals` hanya punya `PRIMARY KEY (id)`, tanpa constraint UNIQUE pada teks pertanyaan. DISIMPULKAN: import ulang file yang sama kemungkinan besar membuat soal ganda (inferensi dari skema, bukan perilaku teramati) |
| **Moodle** | TIDAK DIKETAHUI dari sumber yang dibaca (tidak dibahas di laporan pembanding) |
| **Quizizz** | TIDAK DIKETAHUI |
| **Kahoot!** | TIDAK DIKETAHUI |
| **Exam.net** | — (salin antar-ujian; tidak ada konsep duplikat file) |
| **ProProfs** | TIDAK DIKETAHUI |
| **Google Forms** | — (centang per soal memberi kontrol manual atas apa yang masuk) |
| **Respondus** | ⚠️ Warning "judul duplikat" di Preview (boleh diabaikan) — berasal dari judul otomatis 20 karakter pertama soal bila tidak ada `Title:` |
| **Canvas QTI** | TIDAK DIKETAHUI |
| **Socrative** | TIDAK DIKETAHUI |
| **CBT SMK v2 saat ini** | ❌ Tidak ada pemeriksaan duplikat di parser import (TERBUKTI kode: `parseCSVImport` tidak membandingkan pertanyaan dengan soal yang sudah ada maupun antar-baris). Import ulang file yang sama membuat soal ganda |

### 1.10 Ringkasan posisi v2 (DISIMPULKAN dari tabel)

v2 unggul di tiga hal yang TIDAK dimiliki pembanding mana pun sekaligus: (1) error per baris bernomor + dua mode gagal yang eksplisit, (2) 16 tipe soal dalam satu file, (3) batas 500 baris yang besar. v2 tertinggal di empat hal yang dimiliki hampir semua pembanding spreadsheet: (a) template unduhan resmi, (b) preview isi sebelum simpan, (c) kolom media, (d) kolom kategori/kompetensi. Dan v2 satu-satunya yang memakai tiga pemisah kunci dalam satu format — semua pembanding lain memakai maksimal satu pemisah kunci per format (koma) atau penanda tunggal (`*`, `=`).

---

## Bab 2 — SPESIFIKASI import soal CBT SMK v2 yang "resmi" dan anti-bingung

Bab ini adalah **KEPUTUSAN SPESIFIKASI** (bukan laporan riset), kecuali bila suatu butir diberi label TERBUKTI/DISIMPULKAN. Tugas implementasi berikutnya wajib mengikuti Bab ini; bila ada konflik antara Bab ini dan ANALISIS.md Bab 5.5, Bab ini yang berlaku untuk import (lihat §2.10).

### 2.1 Prinsip desain (yang mengikat semua keputusan di bawah)

1. **Satu file, satu alur, satu makna per kolom.** Guru mengisi satu workbook `.xlsx`, mengunggah, meninjau hasil parsing, lalu mengonfirmasi. Tidak ada dialek kedua yang harus dihafal.
2. **Kunci ditegaskan guru, bukan ditebak mesin.** Tidak ada inferensi AI, tidak ada parsing "pintar" yang menebak maksud guru. Dokumen yang tidak mengikuti kontrak ditolak dengan laporan kesalahan — bukan "ditebak" mesin (prinsip yang sudah ada di ANALISIS.md Bab 5.5 untuk Word, diterapkan juga di sini).
3. **Validasi ketat + warning berlapis.** Error = baris tidak bisa disimpan. Warning = baris bisa disimpan, tetapi guru harus tahu (ala Respondus). Tidak ada kegagalan diam-diam (anti-Quizizz): nilai kunci yang tidak valid selalu ditolak dengan error bernomor baris, tidak pernah disimpan salah.
4. **Preview adalah bagian dari alur, bukan fitur tambahan.** Tidak ada soal yang masuk bank sebelum guru melihat hasil parsing-nya.
5. **Kompatibilitas mundur untuk file CSV lama.** File CSV lama (format parser saat ini) tetap terbaca dengan nilai default yang sama seperti sekarang. Yang berubah adalah jalur utama (Excel) dan UI; bukan semantik kunci CSV lama.
6. **Import adalah asisten pengetikan, bukan jalan pintas validasi.** Semua soal hasil import masuk berstatus `draft`, wajib telaah. (TERBUKTI sebagai prinsip di ANALISIS.md Bab 5.5 + TERBUKTI di kode: INSERT import memakai status `'draft'`.)
7. **Satu bank, satu konteks.** Import selalu ke satu bank yang sudah dipilih (bank menentukan mapel + tingkat). Tidak ada kolom `kode_bank` di file — bank berasal dari konteks UI + parameter `bank_id`, bukan dari isi file.

### 2.2 Format file dan struktur workbook

| Keputusan | Isi |
|---|---|
| Format file | **`.xlsx` saja.** File `.xls`, `.csv`, `.docx`, `.txt` ditolak di jalur baru dengan error level file (lihat §2.6). Alasan: `.xlsx` adalah format yang bisa ditulis Excel / LibreOffice / Google Sheets — semua alat yang dipakai guru SMK; `.xls` adalah format biner legacy yang menambah beban parser tanpa manfaat. (KEPUTUSAN SPESIFIKASI) |
| Nama file template | `template-import-soal-cbt-smk-v2.xlsx` |
| Endpoint unduh template | `GET /api/v1/bank-soal/import/template` → mengunduh workbook di bawah. Permission: `banksoal.import`. (KEPUTUSAN SPESIFIKASI — endpoint baru) |
| Sheet 1 — `Petunjuk` | Panduan pengisian lengkap (§2.9), ditulis di dalam file. Sheet ini tidak dibaca parser; hanya untuk manusia. |
| Sheet 2 — `Soal` | **Satu-satunya sheet yang dibaca parser.** Berisi baris header (§2.3) + baris data. Satu baris = satu soal. Urutan baris dalam file = urutan soal yang ditambahkan ke bank. (KEPUTUSAN SPESIFIKASI) |
| Sheet 3 — `Contoh` | Header yang sama dengan sheet `Soal` + **16 baris contoh, satu per tipe** (§2.5). Guru boleh menyalin baris contoh ke sheet `Soal` lalu mengganti isinya. Sheet ini tidak dibaca parser. |
| Sheet 4 — `Referensi Tipe` | Tabel referensi 16 tipe: kode, label UI, kelompok (`[Inti AKM]` / `[Varian aplikasi]`), format kunci, pemisah yang dipakai, kolom opsi yang dipakai. Tidak dibaca parser. |
| Batas file | Maks **500 baris data** per berkas (dipertahankan dari CSV lama — TERBUKTI kode) + maks **5 MB** ukuran file (KEPUTUSAN SPESIFIKASI baru; sebelumnya tidak ada batas byte — TERBUKTI kode tidak punya batas byte. Alasan 5 MB: file teks soal 500 baris tidak akan mendekati angka ini; batas ini hanya menghentikan file yang salah unggah / berulang-ulang gambar tertanam). |
| Encoding / sel | Nilai sel dibaca sebagai teks. Sel kosong = string kosong. Sel numerik (mis. `bobot` ditulis sebagai angka Excel, bukan teks) wajib diterima dan dibaca sebagai teksnya (`1` → `"1"`). Sel formula TIDAK dievaluasi secara khusus — nilai yang dibaca adalah nilai hasil kalkulasi terakhir yang tersimpan di file (perilaku standar library pembaca `.xlsx`). (KEPUTUSAN SPESIFIKASI) |
| Baris kosong | Baris yang semua selnya kosong dilewati, bukan error. (Dipertahankan dari CSV lama — TERBUKTI kode.) Baris yang hanya sebagian terisi tetapi `tipe`-nya kosong → error `tipe wajib diisi` (bukan dilewati). |

### 2.3 Daftar kolom (sheet `Soal` dan `Contoh`)

Header ditulis **persis** seperti di bawah (huruf kecil, underscore). Pencocokan header: case-insensitive + trim (dipertahankan dari CSV lama — TERBUKTI kode). Header yang tidak dikenal → diabaikan + satu warning level file yang menyebut nama kolomnya (lihat §2.6) — jangan gagalkan file karena satu kolom tambahan (pelajaran dari kerapuhan positional Canvas QTI: kolom bergeser = semua salah; pencocokan by-name menghindari itu).

| # | Kolom | Wajib? | Default bila kosong | Arti |
|---|---|---|---|---|
| 1 | `tipe` | **Wajib diisi per baris** (header wajib ada) | — (error bila kosong) | Kode tipe persis dari `docs/TIPE-SOAL.md`: `pg`, `pg_kompleks`, `menjodohkan`, `isian`, `uraian`, `benar_salah`, `pg_multi`, `urutkan`, `isi_ganda`, `teks_rumpang`, `peta_konsep`, `studi_kasus`, `pernyataan_sikap`, `upload_file`, `praktik_singkat`, `pilihan_gambar`. Kode salah / tidak dikenal → error. **Jangan** tulis label UI ("Pilihan Ganda") di kolom ini — parser hanya menerima kode. (TERBUKTI kode: `ValidTipe`) |
| 2 | `pertanyaan` | **Wajib** | — (error bila kosong) | Teks soal. Untuk `teks_rumpang`, rumpang ditulis `___` (tiga underscore) di dalam teks ini. Untuk `isi_ganda`, pertanyaan boleh menyebut kolom secara umum; nama kolom di-generate otomatis (lihat §2.4). |
| 3 | `stimulus` | Opsional, **wajib untuk `studi_kasus`** | `""` | Teks stimulus/bacaan/kasus. Dipakai bersama bila beberapa soal memakai stimulus yang sama: guru menyalin teks stimulus ke tiap baris (v1 tidak punya ID stimulus bersama — lihat §2.11). |
| 4–8 | `opsi_a`, `opsi_b`, `opsi_c`, `opsi_d`, `opsi_e` | Tergantung tipe (§2.4) | `""` (diabaikan bila tipe tidak memakai opsi) | Opsi / pernyataan / pasangan / butir, sesuai tipe. Maks 5 (A–E). Opsi yang kosong di tengah (mis. `opsi_b` kosong tetapi `opsi_c` terisi) tetap dibaca berurutan sesuai label yang terisi — label dipakai apa adanya (A–E), bukan dirapatkan. |
| 9 | `kunci` | Header wajib ada; **isi wajib** kecuali dua pengecualian di bawah | — | Kunci / rubrik, format per tipe di §2.4. Pengecualian isi kosong: (a) `pernyataan_sikap` — `kunci` diabaikan total (skala 1–4 tetap); bila terisi, diabaikan + info di preview. (b) `uraian`, `upload_file`, `praktik_singkat` — `kunci` berisi rubrik; boleh kosong, tetapi menghasilkan **warning** "rubrik belum diisi" (rubrik bisa dilengkapi di editor sesudah import). **Perubahan dari CSV lama (KEPUTUSAN SPESIFIKASI, dicatat eksplisit):** parser lama menolak SEMUA baris dengan `kunci` kosong (`"kunci wajib diisi"`). Aturan baru melonggarkan khusus untuk 4 tipe di atas, karena pada tipe-tipe itu `kunci` bukan kunci jawaban melainkan rubrik / tidak berlaku. Untuk 12 tipe lain yang dinilai otomatis, `kunci` kosong tetap ERROR. |
| 10 | `bobot` | Opsional | `1` | Bilangan bulat > 0. Bukan bilangan bulat / ≤ 0 → error. (TERBUKTI kode) |
| 11 | `level` | Opsional | `sedang` | Salah satu: `mudah`, `sedang`, `sulit` (case-insensitive). Nilai lain → error. (TERBUKTI kode: `NormalLevel` + `ValidLevel`) |
| 12 | `pembahasan` | Opsional | `""` | Pembahasan soal, ditampilkan ke siswa sesuai aturan tayang pembahasan yang berlaku (tryout / sesudah nilai final). |
| 13 | `media_url` | Opsional | `""` | **Kolom baru.** Satu URL gambar per soal (model Quizizz — TERBUKTI Quizizz memakai 1 gambar/soal via URL). Harus URL `http://` atau `https://` yang valid; bila tidak → error. Keterjangkauan URL **tidak diperiksa** server saat import (server sekolah/LAN belum tentu bisa menjangkau internet; pemeriksaan dilakukan guru dengan mengklik URL di preview — lihat §2.8). File gambar tertanam di `.xlsx` **tidak** ikut terimpor (sama seperti Quizizz — TERBUKTI gambar tertanam tidak ikut). **Perubahan dari CSV lama:** `INSERT` import lama selalu menulis `media_url = ""` (TERBUKTI kode); mulai jalur Excel ini, nilai kolom dipakai. |
| 14 | `media_alt` | Opsional, **warning bila `media_url` terisi tetapi kolom ini kosong** | `""` | Teks alternatif gambar (aksesibilitas). Diwajibkan secara moral oleh ANALISIS.md Bab 5.2 ("gambar wajib teks alternatif") tetapi dijadikan warning — bukan error — agar import tidak gagal hanya karena alt tertinggal; guru melengkapi di editor. |
| 15 | `kisi_kisi` | Opsional | NULL (soal "belum terpetakan") | **Kolom baru.** Kode kisi-kisi sebagaimana tampil di halaman Kisi-kisi bank/mapel tersebut. Bila terisi tetapi tidak ditemukan → **error** (bukan diam-diam dibuang — pelajaran dari ProProfs/Moodle: nilai tak dikenal harus ditolak, bukan diabaikan). Bila kosong, soal masuk sebagai draft belum terpetakan; pemakaian di paket resmi tetap mengikuti aturan kisi-kisi yang berlaku. Pencocokan kode: eksak setelah trim; case-insensitive. **Catatan implementasi:** bila master kisi-kisi belum punya kolom `kode` yang tampil di UI, tugas implementasi harus menambahkannya dulu — lihat Bab 4 Q2. (KEPUTUSAN SPESIFIKASI) |
| 16 | `kompetensi` | Opsional | NULL | **Kolom baru.** Kode kompetensi (CP/TP) sebagaimana tampil di master Kompetensi. Aturan sama dengan `kisi_kisi`: tidak ditemukan → error; kosong → NULL. (KEPUTUSAN SPESIFIKASI) |

**Kolom yang SENGAJA TIDAK ADA di v1 jalur Excel (lihat §2.11):** `level_kognitif` (C1–C6) dan `estimasi_detik` — keduanya ada di usulan ANALISIS.md Bab 5.5 tetapi **tidak ada kolomnya di tabel `soal` saat ini** (TERBUKTI dari daftar kolom `SELECT` di `banksoal.go`: `id, bank_id, tipe, kisi_kisi_id, kompetensi_id, stimulus, pertanyaan, media_url, jawaban, kunci, pembahasan, bobot, level, status, …` — tidak ada `level_kognitif` / `estimasi_detik`). Menambahkan kolom ke template tanpa kolom DB adalah janji kosong; keduanya ditunda ke migrasi berikutnya. `nama_soal`/`judul` juga tidak ada — judul soal di-generate sistem dari pertanyaan (pola Respondus memakai 20 karakter pertama sebagai judul bila tidak ada `Title:` adalah inspirasi, bukan kewajiban v1).

### 2.4 Semantik kunci per tipe (wajib sama dengan parser CSV lama)

Pemisah **tidak berubah** dari CSV lama (kompatibilitas mundur). Yang baru adalah: pemisah kini tertulis di template + panduan + sheet Referensi, bukan hanya di `docs/TIPE-SOAL.md`.

| Tipe | `opsi_a`…`opsi_e` berisi | `kunci` berisi | Aturan validasi (pesan di §2.6) |
|---|---|---|---|
| `pg` | Teks opsi (min 2) | Satu label, mis. `B` | Label harus cocok dengan label opsi yang terisi |
| `pg_kompleks` | Teks opsi (min 2) | Label dipisah koma, mis. `A,C` | Minimal 1 label; semua label harus cocok dengan opsi yang terisi |
| `menjodohkan` | Pasangan `kiri::kanan` (min 2) | Pasangan benar `kiri::kanan;kiri::kanan` (antar-pasangan dipisah `;`) | Tiap opsi harus berformat `kiri::kanan` dengan kedua sisi terisi; tiap pasangan kunci harus berformat sama. **Tidak** diwajibkan kunci = opsi (pengecoh di sisi kanan didukung: pasangan kunci boleh memuat pasangan yang tidak ada di opsi tampil — perilaku dipertahankan dari parser lama yang tidak memvalidasi kecocokan kunci-vs-opsi untuk tipe ini) |
| `isian` | Diabaikan | Jawaban; alternatif dipisah `\|`, mis. `Jakarta\|DKI Jakarta` | Minimal 1 alternatif |
| `uraian` | Diabaikan | Rubrik (teks bebas); boleh kosong + warning | — |
| `benar_salah` | Pernyataan (min 2) | Nilai `B`/`S` dipisah koma, sebanyak pernyataan, mis. `B,S,B`. Juga diterima: `benar`/`salah`, `true`/`false`, `1`/`0` (case-insensitive) | Jumlah nilai harus = jumlah pernyataan; tiap nilai harus salah satu bentuk di atas |
| `pg_multi` | Teks opsi (min 2) | Sama seperti `pg_kompleks`: label dipisah koma, mis. `A,C` | Sama seperti `pg_kompleks` |
| `urutkan` | Butir yang diurutkan (min 2) | Urutan benar sebagai label dipisah koma, sebanyak butir, mis. `C,A,B,D` | Jumlah label = jumlah butir; tiap label harus cocok dengan opsi yang terisi; tidak boleh ada label duplikat |
| `isi_ganda` | Diabaikan | Jawaban per kolom dipisah `\|`, mis. `10\|cm` | Minimal 1 jawaban. **Nama kolom di-generate otomatis** menjadi `Kolom 1`, `Kolom 2`, … sesuai jumlah jawaban (dipertahankan dari parser lama — TERBUKTI kode). Guru belum bisa memberi nama kolom sendiri lewat import v1 (§2.11) |
| `teks_rumpang` | Diabaikan | Jawaban per rumpang dipisah `\|`, mis. `Jakarta\|1945` | Jumlah jawaban kunci harus = jumlah `___` di `pertanyaan` |
| `peta_konsep` | Pasangan `kiri::kanan` (min 2) | Sama seperti `menjodohkan` | Sama seperti `menjodohkan` |
| `studi_kasus` | Teks opsi (min 2) | Satu label, mis. `B` | Sama seperti `pg` + `stimulus` wajib terisi |
| `pernyataan_sikap` | Pernyataan sikap (min 2) | **Diabaikan** (boleh kosong) | Skala 1–4 otomatis. Tipe ini **tidak dinilai** (survei/refleksi) — UI preview wajib menampilkan label "tidak dinilai" |
| `upload_file` | Diabaikan | Rubrik (teks bebas); boleh kosong + warning | — |
| `praktik_singkat` | Diabaikan | Rubrik (teks bebas); boleh kosong + warning | — |
| `pilihan_gambar` | **URL gambar** per opsi (min 2), mis. `https://…/a.png` | Satu label, mis. `B` | Tiap opsi yang terisi harus URL `http(s)` yang valid; label kunci harus cocok dengan label opsi yang terisi |

**Aturan penilaian yang berlaku untuk soal hasil import** (TERBUKTI `docs/TIPE-SOAL.md` + `types.go`, tidak berubah): PG/studi_kasus/pilihan_gambar 1/0 × bobot; pg_kompleks/pg_multi proporsional tanpa penalti; benar_salah proporsional per pernyataan; menjodohkan/peta_konsep proporsional per pasangan; isian exact-match setelah normalisasi (trim + lowercase + rapatkan spasi); uraian/upload_file/praktik_singkat manual; urutkan proporsional per posisi; isi_ganda/teks_rumpang proporsional per kolom/rumpang; pernyataan_sikap non-nilai.

### 2.5 Contoh baris per tipe (isi sheet `Contoh` — 16 baris)

Setiap contoh di bawah adalah satu baris di sheet `Contoh` dengan header §2.3. Kolom yang tidak disebut memakai default (`bobot` = 1, `level` = `sedang`, sisanya kosong), kecuali dicatat. Contoh memakai materi umum yang netral (bukan soal ujian sungguhan) agar bisa disalin apa adanya sebagai kerangka.

| # | `tipe` | `pertanyaan` | `stimulus` | `opsi_a` … `opsi_e` | `kunci` | Catatan |
|---|---|---|---|---|---|---|
| 1 | `pg` | Ibu kota provinsi Jawa Barat adalah … | — | A: Bandung; B: Semarang; C: Medan; D: Surabaya | `A` | — |
| 2 | `pg_kompleks` | Manakah yang termasuk perangkat keras komputer? (Pilih semua yang benar.) | — | A: Monitor; B: Sistem operasi; C: Keyboard; D: Peramban | `A,C` | Kunci dipisah koma |
| 3 | `menjodohkan` | Pasangkan kota dengan provinsinya. | — | A: `Bandung::Jawa Barat`; B: `Semarang::Jawa Tengah`; C: `Surabaya::Jawa Timur`; D: `Medan::Sumatera Utara` | `Bandung::Jawa Barat;Semarang::Jawa Tengah;Surabaya::Jawa Timur;Medan::Sumatera Utara` | Pasangan `::`, antar-pasangan `;` |
| 4 | `isian` | Ibu kota negara Indonesia adalah …. | — | (diabaikan) | `Jakarta\|DKI Jakarta` | Alternatif dipisah `\|` |
| 5 | `uraian` | Jelaskan perbedaan antara RAM dan ROM beserta contoh penggunaannya dalam kehidupan sehari-hari. | — | (diabaikan) | `Rubrik: pengertian RAM (30), pengertian ROM (30), perbedaan (20), contoh (20).` | Kunci = rubrik |
| 6 | `benar_salah` | Tentukan benar atau salah untuk setiap pernyataan berikut. | — | A: `Air mendidih pada suhu 100°C pada tekanan 1 atm.`; B: `Bulan menghasilkan cahayanya sendiri.`; C: `Fotosintesis menghasilkan oksigen.` | `B,S,B` | Jumlah nilai = jumlah pernyataan |
| 7 | `pg_multi` | Manakah yang termasuk bilangan prima? (Pilih semua yang benar.) | — | A: 2; B: 4; C: 5; D: 9 | `A,C` | Sama seperti pg_kompleks |
| 8 | `urutkan` | Urutkan langkah mencuci tangan yang benar. | — | A: `Bilas tangan dengan air bersih.`; B: `Basahi tangan dengan air mengalir.`; C: `Gosok kedua telapak tangan dengan sabun.`; D: `Keringkan tangan.` | `B,C,A,D` | Label dalam urutan benar, dipisah koma |
| 9 | `isi_ganda` | Sebuah persegi panjang memiliki panjang 10 dan lebar 5. Isi kolom berikut. | — | (diabaikan) | `50\|30` | Jawaban per kolom dipisah `\|`; kolom jadi "Kolom 1" (luas), "Kolom 2" (keliling) — nama kolom belum bisa diatur (§2.11) |
| 10 | `teks_rumpang` | Ibu kota Indonesia adalah ___ dan mata uangnya adalah ___. | — | (diabaikan) | `Jakarta\|Rupiah` | Jumlah jawaban = jumlah `___` (2) |
| 11 | `peta_konsep` | Pasangkan konsep dengan pasangannya. | — | A: `Gula::Manis`; B: `Jeruk::Asam`; C: `Garam::Asin` | `Gula::Manis;Jeruk::Asam;Garam::Asin` | Sama seperti menjodohkan |
| 12 | `studi_kasus` | Berdasarkan kasus di atas, tindakan pertama yang paling tepat adalah …. | `Sebuah bengkel menerima komputer yang tidak mau menyala sama sekali: tidak ada lampu indikator, tidak ada suara kipas.` | A: Langsung mengganti motherboard; B: Memeriksa kabel daya dan stopkontak; C: Menginstal ulang sistem operasi; D: Mengganti RAM | `B` | `stimulus` wajib |
| 13 | `pernyataan_sikap` | Beri tanggapanmu untuk setiap pernyataan berikut. | — | A: `Saya senang belajar dengan diskusi kelompok.`; B: `Saya lebih paham bila belajar sambil praktik langsung.` | (kosong — diabaikan) | Preview menampilkan "tidak dinilai" |
| 14 | `upload_file` | Buatlah laporan praktikum sesuai format yang diberikan, lalu unggah berkasnya. | — | (diabaikan) | `Rubrik: kelengkapan isi (40), format laporan (30), ketepatan data (30).` | Kunci = rubrik; boleh kosong + warning |
| 15 | `praktik_singkat` | Praktikkan cara membuat kabel UTP straight dan jelaskan hasilnya. | — | (diabaikan) | `Rubrik: urutan warna benar (50), konektor terpasang rapi (30), penjelasan hasil (20).` | Kunci = rubrik; boleh kosong + warning |
| 16 | `pilihan_gambar` | Manakah gambar yang menunjukkan kabel UTP? | — | A: `https://contoh.sekolah.id/gambar/utp.png`; B: `https://contoh.sekolah.id/gambar/hdmi.png`; C: `https://contoh.sekolah.id/gambar/vga.png`; D: `https://contoh.sekolah.id/gambar/power.png` | `A` | Opsi = URL gambar; ganti dengan URL gambar sekolah masing-masing |

**Catatan untuk implementasi template:** contoh di atas memakai domain `contoh.sekolah.id` sebagai placeholder yang jelas-jelas bukan URL sungguhan — guru wajib menggantinya. Template harus memuat contoh persis ini (bukan contoh lain yang dibuat pengembang template) agar dokumentasi dan file yang diunduh guru selalu sama.

### 2.6 Validasi + katalog pesan error / warning

**Konvensi penomoran (KEPUTUSAN SPESIFIKASI, dipertahankan dari CSV lama):** header = baris 1; baris data pertama = baris 2. Semua pesan error/warning memakai format `{baris, pesan}` dengan `baris` = nomor baris Excel (bukan nomor urut soal). Error level file memakai `baris: 0`, kecuali "kolom wajib hilang" yang memakai `baris: 1` (dipertahankan dari CSV lama — TERBUKTI kode).

**Tingkat keparahan:** ERROR = baris tidak disimpan (dan dalam mode `all_or_nothing` menggagalkan seluruh file). WARNING = baris tetap disimpan; guru melihatnya di preview dan di hasil akhir. INFO = catatan di preview saja (tidak menghalangi apa pun).

#### 2.6.1 Error level file

| Kondisi | Pesan (`baris: 0` kecuali dicatat) |
|---|---|
| File bukan `.xlsx` | `Hanya file .xlsx yang diterima. Unduh template dari tombol "Unduh Template" bila belum punya.` |
| File > 5 MB | `File melebihi batas 5 MB.` |
| File tidak bisa dibaca / rusak | `File tidak bisa dibaca: <detail dari library>.` |
| Sheet `Soal` tidak ditemukan | `Sheet "Soal" tidak ditemukan. Gunakan template resmi — jangan mengganti nama sheet.` |
| Sheet `Soal` hanya berisi header / tanpa baris data | `File harus punya header + minimal 1 baris soal.` |
| Header wajib hilang (`tipe` / `pertanyaan` / `kunci`) | `Kolom wajib hilang di header: <nama_kolom>.` (`baris: 1`) |
| Baris data > 500 | `File melebihi batas 500 baris soal.` |
| Bank tidak ditemukan / pengguna tidak berhak | Error HTTP yang sudah berlaku (403/404 + audit) — bukan error per baris |
| Bank terkunci | `409 bank_locked` — `Bank soal terkunci — import ditolak.` (dipertahankan — TERBUKTI kode) |

#### 2.6.2 Error per baris (pesan lama dipakai verbatim bila sudah ada di kode)

| # | Kondisi | Pesan |
|---|---|---|
| E1 | `tipe` kosong | `tipe wajib diisi` (KEPUTUSAN SPESIFIKASI baru — parser lama tidak pernah mencapai titik ini karena baris dengan tipe kosong gagal di `ValidTipe` dengan pesan E2; pesan baru ini lebih jelas) |
| E2 | `tipe` tidak dikenal | `tipe "<nilai>" tidak dikenal` (dipertahankan — TERBUKTI kode `ValidateSoal`/`parseCSVImport`) |
| E3 | `pertanyaan` kosong | `pertanyaan wajib diisi` (dipertahankan) |
| E4 | `kunci` kosong untuk 12 tipe yang dinilai otomatis | `kunci wajib diisi` (dipertahankan; cakupan dipersempit — lihat §2.3 kolom 9) |
| E5 | `bobot` bukan bilangan bulat / ≤ 0 | `bobot "<nilai>" harus bilangan bulat > 0` (dipertahankan) |
| E6 | `level` bukan mudah/sedang/sulit | `level "<nilai>" harus mudah / sedang / sulit` (dipertahankan) |
| E7 | PG / studi_kasus / pilihan_gambar: opsi terisi < 2 | `opsi_a..opsi_e: minimal 2 opsi terisi` (dipertahankan) |
| E8 | PG / studi_kasus / pilihan_gambar: kunci tidak cocok label opsi | `kunci "<nilai>" tidak cocok dengan label opsi yang terisi` (dipertahankan) |
| E9 | PG kompleks / PG multi: opsi terisi < 2 | `opsi_a..opsi_e: minimal 2 opsi terisi` (dipertahankan) |
| E10 | PG kompleks / PG multi: ada label kunci tidak cocok opsi | `kunci "<label>" tidak cocok dengan label opsi yang terisi` (dipertahankan) |
| E11 | PG kompleks / PG multi: kunci kosong sesudah dipisah | `kunci wajib minimal satu label opsi` (dipertahankan) |
| E12 | Benar-salah: pernyataan < 2 | `opsi_a..opsi_e: isi pernyataan (minimal 2)` (dipertahankan) |
| E13 | Benar-salah: jumlah nilai kunci ≠ jumlah pernyataan | `kunci harus <N> nilai B/S, dipisah koma` (dipertahankan) |
| E14 | Benar-salah: nilai kunci bukan B/S/benar/salah/true/false/1/0 | `kunci "<nilai>" harus B atau S` (dipertahankan) |
| E15 | Menjodohkan / peta konsep: pasangan tampil < 2 | `opsi_a..opsi_e: isi pasangan format kiri::kanan (minimal 2)` (dipertahankan) |
| E16 | Menjodohkan / peta konsep: opsi tidak berformat `kiri::kanan` / salah satu sisi kosong | `opsi <LABEL> harus format kiri::kanan` (dipertahankan) |
| E17 | Menjodohkan / peta konsep: pasangan kunci tidak berformat `kiri::kanan` / sisi kosong | `kunci harus pasangan format kiri::kanan dipisah ;` (dipertahankan) |
| E18 | Isian: tidak ada alternatif jawaban | `kunci isian wajib diisi (alternatif dipisah \|)` (dipertahankan) |
| E19 | Urutkan: butir < 2 | `opsi_a..opsi_e: isi butir yang diurutkan (minimal 2)` (dipertahankan) |
| E20 | Urutkan: jumlah label kunci ≠ jumlah butir | `kunci harus <N> label opsi (urutan benar), dipisah koma` (dipertahankan) |
| E21 | Urutkan: label kunci tidak cocok opsi | `kunci "<label>" tidak cocok dengan label opsi yang terisi` (dipertahankan) |
| E22 | Urutkan: label kunci duplikat | `kunci berisi label "<label>" duplikat` (dipertahankan) |
| E23 | Isi ganda: tidak ada jawaban per kolom | `kunci wajib jawaban per kolom, dipisah \|` (dipertahankan) |
| E24 | Teks rumpang: tidak ada jawaban per rumpang | `kunci wajib jawaban per rumpang, dipisah \|` (dipertahankan) |
| E25 | Teks rumpang: jumlah jawaban ≠ jumlah `___` | `jumlah rumpang ___ harus sama dengan jumlah jawaban kunci (pisah dengan \|)` (dipertahankan) |
| E26 | Studi kasus: `stimulus` kosong | `stimulus wajib diisi untuk tipe studi_kasus` (dipertahankan) |
| E27 | Pernyataan sikap: pernyataan < 2 | `opsi_a..opsi_e: isi pernyataan sikap (minimal 2)` (dipertahankan) |
| E28 | `media_url` terisi tetapi bukan URL `http(s)` valid | `media_url "<nilai>" bukan URL http(s) yang valid` (KEPUTUSAN SPESIFIKASI baru) |
| E29 | `pilihan_gambar`: opsi yang terisi bukan URL `http(s)` valid | `opsi <LABEL> harus URL gambar http(s)` (KEPUTUSAN SPESIFIKASI baru) |
| E30 | `kisi_kisi` terisi tetapi tidak ditemukan | `kisi_kisi "<nilai>" tidak ditemukan di bank ini` (KEPUTUSAN SPESIFIKASI baru) |
| E31 | `kompetensi` terisi tetapi tidak ditemukan | `kompetensi "<nilai>" tidak ditemukan` (KEPUTUSAN SPESIFIKASI baru) |
| E32 | Gagal simpan ke database (per baris, di tahap konfirmasi) | `Gagal menyimpan ke database.` (dipertahankan — TERBUKTI kode) |

**Aturan validasi yang SENGAJA TIDAK ditambahkan (KEPUTUSAN SPESIFIKASI):** tidak ada validasi panjang maksimum pertanyaan/opsi (pelajaran dari Kahoot — batas 120/75 karakter terlalu ketat untuk soal cerita SMK; jangan ditiru). Tidak ada pemeriksaan keterjangkauan `media_url` (server LAN belum tentu menjangkau internet). Tidak ada pemeriksaan bahwa kunci menjodohkan cocok dengan opsi tampil (pengecoh di sisi kanan didukung — dipertahankan dari parser lama).

#### 2.6.3 Warning per baris + warning level file (tidak menghalangi simpan)

| # | Kondisi | Pesan |
|---|---|---|
| W1 | `kunci` kosong untuk `uraian` / `upload_file` / `praktik_singkat` | `kunci kosong — rubrik belum diisi. Rubrik bisa dilengkapi di editor sesudah import.` |
| W2 | `media_url` terisi tetapi `media_alt` kosong | `media_url diisi tetapi media_alt kosong — gambar tanpa teks alternatif.` |
| W3 | Header tidak dikenal di sheet `Soal` (level file, satu warning per kolom) | `Kolom "<nama>" tidak dikenal — diabaikan.` |
| W4 | Duplikat dalam file (lihat §2.7) | `pertanyaan duplikat dengan baris <N>.` |
| W5 | Duplikat dengan soal yang sudah ada di bank (lihat §2.7) | `pertanyaan sudah ada di bank ini.` |
| W6 | `kunci` terisi untuk `pernyataan_sikap` | INFO (bukan warning): `kunci diabaikan untuk pernyataan_sikap (skala 1–4, tidak dinilai).` |
| W7 | `media_url` terisi | INFO di preview: `URL gambar tidak diperiksa keterjangkauannya — klik URL di preview untuk memastikan gambar bisa dibuka.` |

### 2.7 Duplikat (baru — tidak ada di import lama)

**Definisi (KEPUTUSAN SPESIFIKASI):** dua soal disebut duplikat bila teks `pertanyaan`-nya sama setelah **normalisasi**: trim + lowercase + rapatkan spasi berlebih menjadi satu spasi. Tidak ada stemming, tidak ada pembandingan opsi/kunci. Normalisasi ini deterministik dan dapat dijelaskan ke guru dalam satu kalimat; normalisasi yang lebih "pintar" (fuzzy match) ditolak karena tidak dapat dijelaskan dan rawan salah tebak.

| Cakupan | Perilaku |
|---|---|
| Duplikat **dalam file** (dua baris atau lebih dengan pertanyaan ternormalisasi sama) | Setiap baris kedua dan seterusnya mendapat warning W4 yang menyebut nomor baris baris pertama. Baris tetap valid dan tetap disimpan kecuali guru memilih `duplikat=lewati` (di bawah). |
| Duplikat **dengan bank** (pertanyaan ternormalisasi sama dengan soal yang sudah ada di bank tujuan, semua status kecuali yang sudah dihapus) | Baris mendapat warning W5 di preview. Baris tetap valid. |
| Pilihan guru di tahap konfirmasi | Parameter `duplikat` (default `lewati`): `lewati` = baris yang ter-flag duplikat (W4 kedua-dst. dan W5) TIDAK disimpan, sisanya disimpan; `simpan_ganda` = semua baris valid disimpan apa pun flag duplikatnya. Pilihan ini berlaku untuk kedua mode gagal (`all_or_nothing` / `skip_errors`): duplikat adalah warning, bukan error, sehingga tidak pernah menggagalkan file dalam mode `all_or_nothing`. |
| Laporan hasil | Respons konfirmasi memuat `dilewati_duplikat` (jumlah baris yang dilewati karena duplikat) terpisah dari `gagal` (error). |

### 2.8 Alur lengkap + endpoint

**Alur (UI):**

```
1. Guru membuka detail Bank Soal → klik "Import Soal" → klik "Unduh Template"
   (GET /api/v1/bank-soal/import/template → template-import-soal-cbt-smk-v2.xlsx)
2. Guru mengisi sheet "Soal" (boleh menyalin baris dari sheet "Contoh"),
   menyimpan sebagai .xlsx.
3. Guru mengunggah file di dialog import → klik "Pratinjau"
   → POST /api/v1/bank-soal/import/preview (multipart: file + bank_id)
   → server parsing + validasi TANPA menulis apa pun ke database.
4. UI menampilkan:
   a. Ringkasan: total baris, jumlah valid, jumlah error, jumlah warning,
      jumlah duplikat, ringkasan per tipe (mis. pg: 12, isian: 3, …),
      total bobot soal valid.
   b. Tabel preview SEMUA baris ter-parse (maks 500): nomor baris Excel,
      tipe (label UI + badge [Inti AKM]/[Varian aplikasi]), pertanyaan,
      opsi/kunci yang terpetakan (kunci ditampilkan sebagai label +
      teks opsi yang dirujuknya — bukan hanya huruf — agar guru bisa
      memverifikasi pemetaan), bobot, level, kisi-kisi/kompetensi
      terpetakan, media_url sebagai tautan yang bisa diklik, warning.
      Baris error ditampilkan dengan pesan error-nya (W1–W6 / E1–E32).
   c. Daftar error bernomor baris + daftar warning.
   d. Pilihan mode: all_or_nothing (default) / skip_errors.
   e. Pilihan duplikat: lewati (default) / simpan_ganda.
   f. Tombol "Simpan" (nonaktif bila mode all_or_nothing dan ada error),
      tombol "Perbaiki file & unggah ulang", tombol "Batal".
5. Guru klik "Simpan"
   → POST /api/v1/bank-soal/import/xlsx (multipart: file + bank_id
     + mode + duplikat; FILE YANG SAMA diunggah ulang)
   → server parsing + validasi ulang (agar tidak ada temp storage yang
     bisa basi / diubah di antara preview dan konfirmasi), lalu menulis
     sesuai mode + pilihan duplikat.
6. UI menampilkan hasil: disimpan (N), gagal (N + daftar error),
   dilewati_duplikat (N), dilewati karena error (N, bila skip_errors),
   + tautan "Lihat bank soal" (daftar soal yang baru masuk, status draft).
   Audit log mencatat import (sukses/gagal/mode/duplikat) — dipertahankan.
```

**Mengapa konfirmasi mengunggah ulang file yang sama (KEPUTUSAN SPESIFIKASI):** alternatifnya adalah server menyimpan hasil parsing sementara (preview_id + TTL). Itu menambah state sementara yang bisa basi, perlu dibersihkan, dan perlu diamankan. Mengunggah ulang membuat konfirmasi stateless dan deterministik: hasil konfirmasi selalu mencerminkan file yang benar-benar disimpan. Biaya: satu parsing tambahan (murah untuk ≤ 500 baris).

**Endpoint (KEPUTUSAN SPESIFIKASI — baru, kecuali dicatat):**

| Endpoint | Metode | Body | Respons |
|---|---|---|---|
| `/api/v1/bank-soal/import/template` | GET | — | File `.xlsx` template (4 sheet). Permission `banksoal.import`. |
| `/api/v1/bank-soal/import/preview` | POST | multipart: `file` (.xlsx), `bank_id` (UUID) | `{bank_id, total_baris, valid, error: <n>, warning: <n>, duplikat: <n>, ringkasan_tipe: {<kode>: <n>}, total_bobot_valid, errors: [{baris, pesan}], warnings: [{baris, pesan}], duplikat_detail: [{baris, baris_duplikat \| soal_id, pesan}], preview: [{baris, tipe, tipe_label, kelompok, pertanyaan, stimulus, opsi, kunci_terpetakan, bobot, level, pembahasan, media_url, media_alt, kisi_kisi, kompetensi, warnings: [...]}]}`. **Tidak menulis ke database.** Bank terkunci → 409 (sama seperti import). |
| `/api/v1/bank-soal/import/xlsx` | POST | multipart: `file`, `bank_id`, `mode` (`all_or_nothing` default / `skip_errors`), `duplikat` (`lewati` default / `simpan_ganda`) | `{sukses, gagal, dilewati_duplikat, warnings: [{baris, pesan}], errors: [{baris, pesan}], mode, duplikat, disimpan: bool}`. `disimpan = sukses > 0`. Mode `all_or_nothing` + ada error → `sukses: 0`, `disimpan: false`, tidak ada yang tersimpan (dipertahankan dari endpoint lama). Audit log dicatat (dipertahankan). |
| `/api/v1/bank-soal/import` (lama, JSON `csv_text`) | POST | JSON lama | **Dipertahankan, ditandai deprecated di docs.** Perilaku tidak berubah (termasuk: `media_url` tetap `""`, `kisi_kisi_id`/`kompetensi_id` tetap NULL, `kunci` kosong tetap error untuk semua tipe). UI baru tidak memakainya. Alasan dipertahankan: kompatibilitas mundur untuk klien/skrip yang sudah memakainya + tes yang sudah ada. |

**Otorisasi (dipertahankan — TERBUKTI kode + `docs/TIPE-SOAL.md`):** permission `banksoal.import`; proktor dilarang mutasi soal; guru dibatasi scope banknya (`checkGuruScope`); bank terkunci → 409. Soal hasil import berstatus `draft`.

### 2.9 Panduan pengisian lengkap (isi sheet `Petunjuk` + halaman bantuan import)

Teks di bawah adalah isi yang harus ada di sheet `Petunjuk` (ditulis dengan kalimat pendek, tanpa istilah teknis yang tidak dijelaskan). Tugas implementasi template wajib memuat teks ini — bukan teks lain yang dibuat sendiri.

```
CARA MENGISI TEMPLATE IMPORT SOAL — CBT SMK v2

1. Isi HANYA sheet "Soal". Sheet "Contoh" dan "Referensi Tipe" hanya contoh.
   Jangan mengganti nama sheet dan jangan menghapus/mengganti nama kolom.
2. Satu baris = satu soal. Urutan baris = urutan soal masuk ke bank.
   Baris yang semua selnya kosong akan dilewati (tidak dihitung error).
3. Kolom "tipe" wajib diisi dengan KODE, bukan nama panjang.
   Contoh benar: pg. Contoh salah: Pilihan Ganda.
   Daftar kode ada di sheet "Referensi Tipe" dan sheet "Contoh".
4. Kolom "pertanyaan" wajib diisi. Tulis satu pertanyaan yang jelas.
5. Kolom "kunci" — cara menulisnya BERBEDA per tipe (ini sumber salah
   yang paling sering terjadi, jadi baca baris tipe Anda di bawah):
   - pg / studi_kasus / pilihan_gambar: SATU huruf, mis. B
   - pg_kompleks / pg_multi: huruf dipisah KOMA, mis. A,C
   - benar_salah: B/S dipisah KOMA sebanyak pernyataan, mis. B,S,B
   - menjodohkan / peta_konsep: pasangan kiri::kanan, antar-pasangan
     dipisah TITIK KOMA, mis. Bandung::Jawa Barat;Semarang::Jawa Tengah
   - isian: jawaban, alternatif dipisah GARIS TEGAK |, mis. Jakarta|DKI Jakarta
   - isi_ganda / teks_rumpang: jawaban dipisah GARIS TEGAK |, mis. 50|30
   - uraian / upload_file / praktik_singkat: tulis RUBRIK di kolom kunci.
     Boleh kosong (nanti dilengkapi di editor), tetapi akan muncul
     peringatan "rubrik belum diisi".
   - pernyataan_sikap: kolom kunci DIABAIKAN (skala 1-4, tidak dinilai).
6. Kolom "opsi_a" s.d. "opsi_e":
   - Untuk pg / pg_kompleks / pg_multi / benar_salah / urutkan /
     pernyataan_sikap / studi_kasus: isi dengan teks opsi/pernyataan/butir.
   - Untuk menjodohkan / peta_konsep: isi dengan pasangan kiri::kanan.
   - Untuk isian / isi_ganda / teks_rumpang / uraian / upload_file /
     praktik_singkat: kolom opsi TIDAK dipakai, kosongkan saja.
   - Untuk pilihan_gambar: isi dengan URL gambar (https://...), bukan
     teks. Satu opsi = satu gambar.
   - Minimal 2 opsi terisi untuk tipe yang memakai opsi.
7. Kolom "bobot": angka bulat lebih dari 0. Kosong = 1.
   Kolom "level": mudah / sedang / sulit. Kosong = sedang.
8. Kolom "stimulus": teks bacaan/kasus. WAJIB diisi untuk studi_kasus.
   Bila beberapa soal memakai stimulus yang sama, salin teksnya ke
   tiap baris soal tersebut.
9. Kolom "pembahasan": penjelasan jawaban. Opsional.
10. Kolom "media_url": SATU tautan gambar (https://...) untuk soal ini.
    Gambar yang ditempel langsung di file Excel TIDAK ikut terimpor —
    harus berupa tautan. Kolom "media_alt": teks alternatif gambar
    (untuk pembaca layar). Bila media_url diisi tetapi media_alt kosong,
    muncul peringatan (tidak menggagalkan import).
11. Kolom "kisi_kisi" dan "kompetensi": isi dengan KODE yang tampil di
    halaman Kisi-kisi / Kompetensi bank ini. Bila kode tidak ditemukan,
    baris tersebut ERROR (tidak disimpan diam-diam tanpa kisi-kisi).
    Kosongkan bila belum dipetakan — soal tetap masuk sebagai draft.
12. Teks rumpang ditulis dengan tiga garis bawah: ___
    Contoh: "Ibu kota Indonesia adalah ___." Jawabannya di kolom kunci.
13. Jangan menulis rumus Excel / formula di sel mana pun. Tulis teks
    biasa. Rumus matematika dalam soal ditulis sebagai teks, mis. "2x + 3".
14. Batas: maks 500 baris soal per file, maks 5 MB.
    Bila soal Anda lebih dari 500, pecah menjadi beberapa file.
15. Alur: unggah file → PERIKSA PRATINJAU (pastikan kunci terpetakan
    ke opsi yang benar — jangan hanya melihat "tidak ada error") →
    pilih mode (tolak semua bila ada error / lewati baris yang error) →
    pilih perlakuan duplikat (lewati / simpan ganda) → Simpan.
    Soal hasil import berstatus DRAFT — wajib ditelaah sebelum dipakai.
16. Bila ada error, perbaiki di file Excel Anda lalu unggah ulang.
    Daftar error menyebut NOMOR BARIS Excel (baris 1 = header).
```

### 2.10 Apa yang harus diganti dari import CSV lama

Tabel migrasi untuk tugas implementasi. "Tetap" = perilaku dipertahankan verbatim; "Ganti" = perilaku lama dihapus/diganti; "Baru" = belum ada di implementasi lama.

| # | Bagian | Lama (TERBUKTI kode) | Baru (KEPUTUSAN SPESIFIKASI) | Status |
|---|---|---|---|---|
| 1 | Cara kirim file | JSON `csv_text` (teks CSV ditempel di textarea UI) | Upload file `.xlsx` multipart di UI baru; endpoint JSON lama dipertahankan tapi deprecated dan tidak dipakai UI | Ganti (UI) / Tetap (endpoint lama) |
| 2 | Template | Tidak ada unduhan template; hanya header + 1 contoh PG di placeholder | `GET …/import/template` → workbook 4 sheet (§2.2) dengan 16 contoh baris | Baru |
| 3 | Panduan | Satu kalimat + rujuk `docs/TIPE-SOAL.md` | Sheet `Petunjuk` di dalam template (§2.9) + panduan di UI import | Baru |
| 4 | Preview | Tidak ada | `POST …/import/preview` (dry-run) + tabel preview isi + konfirmasi (§2.8) | Baru |
| 5 | Kolom | `tipe, pertanyaan, opsi_a..e, kunci, bobot, level, stimulus, pembahasan` | Tambah `media_url`, `media_alt`, `kisi_kisi`, `kompetensi`. Header by-name dipertahankan; header tak dikenal → warning (baru), bukan diabaikan diam-diam | Baru (kolom) / Tetap (mekanisme header) |
| 6 | `media_url` | Selalu `""` di INSERT import | Diisi dari kolom `media_url` | Ganti |
| 7 | `kisi_kisi_id` / `kompetensi_id` | Tidak di-INSERT import (NULL) | Diisi dari kolom `kisi_kisi` / `kompetensi` lewat lookup kode; tidak ditemukan → error | Ganti |
| 8 | Aturan `kunci` kosong | Error untuk semua tipe | Error untuk 12 tipe dinilai otomatis; warning untuk uraian/upload_file/praktik_singkat; diabaikan untuk pernyataan_sikap (§2.3 kolom 9) | Ganti (dilonggarkan, dicatat) |
| 9 | Pemisah kunci | `,` / `;` / `\|` / `::` per tipe, hanya terdokumentasi di `docs/TIPE-SOAL.md` | Pemisah TIDAK berubah; kini tertulis di template + panduan + sheet Referensi | Tetap (semantik) / Baru (dokumentasi di tempat) |
| 10 | Validasi per baris + pesan error | Ada, `{baris, pesan}` (E2–E27) | Dipertahankan verbatim; ditambah E1, E28–E32 + warning W1–W7 | Tetap + Baru |
| 11 | Mode gagal | `all_or_nothing` (default) / `skip_errors` | Dipertahankan; ditambah pilihan `duplikat` (lewati/simpan_ganda) | Tetap + Baru |
| 12 | Duplikat | Tidak diperiksa | Deteksi + warning + pilihan lewati/simpan_ganda (§2.7) | Baru |
| 13 | Batas | 500 baris; tanpa batas byte | 500 baris (tetap) + 5 MB (baru) | Tetap + Baru |
| 14 | Status hasil import | `draft` | `draft` (tetap) | Tetap |
| 15 | Otorisasi + audit + lock | `banksoal.import`, scope guru, bank terkunci → 409, audit log | Tetap untuk semua endpoint baru | Tetap |
| 16 | Spesifikasi ANALISIS.md Bab 5.5 (kolom `kode_bank, nomor_urut, …`, import Word `.docx`, `level_kognitif`, `estimasi_detik`) | Tidak pernah diimplementasikan | **Tidak diadopsi mentah-mentah.** `kode_bank`/`nomor_urut` tidak dipakai (bank dari konteks, urutan dari baris). Word `.docx` tetap rencana tahap lanjutan (§2.11). `level_kognitif`/`estimasi_detik` ditunda (tidak ada kolom DB — §2.3) | Dikoreksi / Ditunda |

### 2.11 Yang SENGAJA DITUNDA (di luar tugas import Excel ini)

| Ditunda | Alasan (jujur) |
|---|---|
| `level_kognitif` (C1–C6) + `estimasi_detik` sebagai kolom import | Tidak ada kolomnya di tabel `soal` saat ini (TERBUKTI daftar kolom SELECT). Menambahkannya butuh migrasi DB + keputusan (apakah kognitif per soal wajib? default C2?) — di luar tugas import Excel. Dicatat agar tidak diklaim "sudah didukung" hanya karena ada di ANALISIS.md Bab 5.5. |
| Upload file gambar pendamping (ZIP / multi-file) | Model URL-satu-gambar (Quizizz) sudah menutup kebutuhan mayoritas soal bergambar dengan kompleksitas jauh lebih kecil. Upload file pendamping (model Exo HTML Zip / Moodle GIFT+ZIP) menambah alur file-manager + storage + pemetaan path yang rawan salah; ditunda sampai ada kebutuhan nyata dari guru. |
| Nama kolom kustom untuk `isi_ganda` | Parser lama men-generate `Kolom 1…N` (TERBUKTI kode). Memberi nama kolom butuh kolom tambahan / struktur jawaban baru — ditunda; guru menulis nama kolom di teks pertanyaan untuk sementara. |
| Import Word (`.docx`) | Rencana yang belum terwujud (lihat koreksi di kepala Bab 2). Bila dibangun nanti: kontrak format ketat + file contoh + preview + warning berlapis + dokumen tak sesuai kontrak ditolak (pola Respondus + prinsip ANALISIS.md Bab 5.5). Bukan bagian tugas ini. |
| Import Google Docs | **Tidak didukung di v2** (ANALISIS.md Bab 5.5: butuh API Google + izin dokumen; guru ekspor ke `.docx` dulu — fitur tahap lanjutan). Jangan pernah diklaim sebagai fitur v2 — pelajaran dari Exo: Google Docs di sana pun hanya klaim marketing + konfigurasi env, mekanismenya TIDAK DIKETAHUI. |
| Jalur AI (generate / inferensi kunci dari dokumen) | Ditolak untuk import kunci: kunci hasil inferensi (ala ProProfs AI / Wayground AI) adalah risiko nilai yang salah. Bila suatu hari ditambahkan, kuncinya wajib berlabel "hasil inferensi — wajib ditinjau" dan tidak boleh langsung masuk bank tanpa konfirmasi per soal. |
| ID stimulus bersama / bank stimulus | Stimulus ditulis ulang per baris di v1 (disalin). Stimulus bersama terpusat (perpustakaan media/stimulus ANALISIS.md Bab 5.5) adalah pekerjaan model data terpisah. |
| Format teks kaya di sel (bold/italic/warna, rumus KaTeX) | Sel Excel dibaca sebagai teks polos (sama seperti Quizizz — TERBUKTI hanya teks polos yang diimpor). Format kaya + rumus terstruktur tetap lewat editor soal (RichEditor TipTap + KaTeX, sesuai audit UI/UX) — bukan lewat import. |

---

## Bab 3 — 3 pelajaran terbaik yang WAJIB diadopsi + yang TIDAK boleh ditiru

### 3.1 Tiga pelajaran yang WAJIB diadopsi

**Pelajaran 1 — Template unduhan resmi + contoh per tipe + panduan DI DALAM file (wajib).**
Dipelajari dari: Kahoot (template resmi + aturan eksplisit), Quizizz (template Google Sheet + 11 kolom), ProProfs (file contoh + instruksi "ganti dummy dengan data asli"), Respondus (file contoh Word + panduan format), Socrative (template + sheet contoh + 6 langkah di dalam template), dan Exo Format 2 (instruksi terbaik justru tertanam di dalam file template itu sendiri — blok "Perhatian" + penjelasan semantik kunci per tipe).
Penerapan di v2: §2.2 (workbook 4 sheet), §2.5 (16 contoh baris), §2.9 (teks Petunjuk). Kegagalan yang diperbaiki: v2 lama memaksa guru menyalin header dari kalimat UI — satu huruf salah ketik = kolom hilang.
Uji penerimaan: guru yang belum pernah memakai import v2 harus bisa mengisi template tanpa membuka dokumentasi lain.

**Pelajaran 2 — Preview isi + cek error SEBELUM simpan, dengan keputusan di tangan guru (wajib).**
Dipelajari dari: Quizizz (review → Publish), Kahoot (cek otomatis → perbaiki spreadsheet ATAU lanjut dan buang soal yang bermasalah), Respondus (Preview + warning berlapis sebelum Finish), dan dari kegagalan Exo (tidak ada bukti preview pra-simpan — verifikasi hanya pasca-import lewat "Preview Bank Soal").
Penerapan di v2: §2.8 (preview dry-run → tabel isi ter-parse → pilihan mode + duplikat → konfirmasi). Celah yang ditutup: kunci yang terpetakan salah (mis. salah pemisah) baru ketahuan setelah soal tersimpan. Mode `all_or_nothing` lama adalah validasi, bukan preview — jangan dianggap pengganti preview.
Uji penerimaan: guru harus bisa melihat, untuk setiap baris, kunci yang terpetakan sebagai label + teks opsi yang dirujuknya — bukan hanya huruf.

**Pelajaran 3 — Warning berlapis + penolakan tegas atas nilai kunci yang tidak valid; jangan pernah gagal diam-diam (wajib).**
Dipelajari dari: Respondus (warning "judul duplikat" boleh diabaikan vs "soal tanpa jawaban benar" JANGAN diabaikan) sebagai model warning berlapis; dan dari kegagalan Quizizz (huruf di kolom Correct Answer terunggah tanpa error tetapi menghasilkan kunci salah pada semua soal) sebagai contoh yang TIDAK boleh terjadi di v2.
Penerapan di v2: §2.6 (ERROR vs WARNING vs INFO; pesan lama dipertahankan verbatim; nilai kunci tak valid selalu error bernomor baris). Yang dipertahankan dari v2 lama: error per baris bernomor, dua mode gagal, validasi ketat yang menolak (bukan menebak), audit log, status draft, tolak import saat bank terkunci — praktik terbaik yang TIDAK dimiliki sebagian pembanding (Moodle XML tidak melaporkan error; Socrative mengizinkan kunci kosong tanpa peringatan).
Uji penerimaan: setiap nilai kunci yang tidak valid pada file uji harus menghasilkan error bernomor baris yang tepat — tidak ada satu pun kasus "tersimpan tetapi kuncinya salah".

**Prinsip pengiring (bukan pelajaran ke-4, tetapi mengikat ketiganya):** kunci ditegaskan guru, bukan diinferensi mesin. Rute AI (ProProfs AI / Wayground AI) membuktikan "impor tanpa format" itu mungkin — tetapi kunci hasil inferensi adalah tebakan. v2 menolak jalur itu untuk import kunci (§2.11).

### 3.2 Yang TIDAK boleh ditiru

| # | Jangan ditiru | Dari siapa | Alasan (bukti) | Konsekuensi bila ditiru di v2 |
|---|---|---|---|---|
| 1 | **Klaim "16 tipe import"** | Exo | TERBUKTI: hanya 9 tipe Format 2 + PG/PGX Format 1 + PG/Esai Excel yang punya contoh penulisan import. 13 kode tercantum di komentar file kompetensi; 3 kode dari klaim "16" tidak ditemukan di sumber mana pun. Angka 16 kemungkinan mencampur tipe soal dengan slot kompetensi `code01–code16`. | Dokumen v2 berbohong tentang cakupan import; guru mencoba mengimpor tipe yang parsernya tidak ada. v2 menulis jujur: 16 tipe didukung import Excel (semuanya punya contoh di §2.5) dengan label "5 inti AKM + 11 varian" — klaim ini TERBUKTI dari kode + spesifikasi, bukan dari Exo. |
| 2 | **Klaim "import Google Docs"** | Exo | TERBUKTI: Google Docs di Exo hanya klaim marketing + key env (`GOOGLE_DOC_ACTIVE`, `GOOGLE_CREDENTIAL` — nama key saja); prosedur impornya TIDAK DIKETAHUI. | Fitur yang diklaim tetapi tidak ada; guru menunggu fitur yang tidak pernah datang. v2: Google Docs tidak didukung (§2.11), guru ekspor ke `.docx` dulu. |
| 3 | **Klaim "import CSV soal" untuk Exo** | Kajian lama Exo (dikoreksi) | TERBUKTI: tidak ada template soal `.csv` di paket Exo; template CSV/XLSX yang ada hanya untuk peserta/matpel/user/grup/sesi/ruangan, bukan soal. | Perbandingan yang salah + guru mencari format yang tidak ada. |
| 4 | **Klaim "import Word termasuk rumus" tanpa bukti mekanisme** | Exo / kajian lama | TERBUKTI: satu-satunya bukti rumus adalah judul video di README (isi tidak diverifikasi); template tidak berisi satu pun contoh rumus; cara parser menangani OMML TIDAK DIKETAHUI. | Janji rumus yang tidak bisa ditepati. v2: rumus di sel Excel dibaca sebagai teks polos (§2.11); rumus terstruktur tetap lewat editor. |
| 5 | **Gagal diam-diam (silent failure)** | Quizizz (huruf di kolom kunci nomor terunggah tanpa error) | TERBUKTI: kunci huruf menghasilkan kunci salah pada semua soal tanpa error apa pun. | Nilai siswa salah massal tanpa ada yang tahu. v2: nilai kunci tak valid = error bernomor baris (§2.6). |
| 6 | **Mengandalkan validator eksternal** | Moodle XML (parser tidak melaporkan error; guru validasi sendiri lewat browser) | TERBUKTI MoodleDocs: parser mengasumsikan well-formed dan tidak mendeteksi/melaporkan error. | Guru tidak tahu letak kesalahan; seluruh file gagal tanpa pesan jelas. v2: importer melaporkan letak error sendiri. |
| 7 | **Batas panjang soal/jawaban yang sangat ketat** | Kahoot (soal 120 karakter, jawaban 75 karakter) | TERBUKTI dari artikel resmi Kahoot. | Soal cerita / stimulus SMK tidak muat; guru memangkas soal agar lolos. v2: tidak ada batas panjang (§2.6). |
| 8 | **Kolom positional tanpa header** | Canvas QTI via konverter CSV (tanpa header; kolom B kosong yang bermakna; header = error) | TERBUKTI dari ASEE. | Satu kolom bergeser = seluruh pemetaan salah. v2: header by-name, header tak dikenal → warning (§2.3). |
| 9 | **Kunci kosong = sah tanpa peringatan** | Socrative (kunci boleh kosong, soal tanpa kunci tetap masuk) | TERBUKTI dari panduan Socrative yang dibaca. | Soal tanpa kunci lolos ke kelas; guru yang lupa mengisi kunci tidak mendapat peringatan. v2: kunci kosong = error untuk 12 tipe dinilai otomatis; warning untuk rubrik; diabaikan hanya untuk pernyataan_sikap yang memang tidak dinilai (§2.3). |
| 10 | **Kunci hasil inferensi AI tanpa label jelas** | ProProfs AI / Wayground AI | TERBUKTI: kedua jalur AI menginferensi kunci dari dokumen/materi. | Kunci tebakan masuk bank dan dinilai seolah-olah kunci guru. v2: tidak ada jalur AI untuk import kunci (§2.11). |
| 11 | **Format multi-dialek tanpa contoh penulisan yang terverifikasi** | Exo (gaya `##...##` + `***` ada sebagai file template, pemakaian di dialog import TIDAK DIKETAHUI; arti `ABS` DISIMPULKAN, belum TERBUKTI) | TERBUKTI dari laporan Exo. | Guru memakai dialek yang parser tidak menerima / makna marker yang ambigu. v2: satu format, satu makna per kolom, contoh per tipe (§2.2–§2.5). Jangan mengadopsi `ABS` atau `##/***` ke v2 tanpa verifikasi di aplikasi asli. |
| 12 | **"Salin antar-ujian" dianggap pengganti import file** | Google Forms / Exam.net | TERBUKTI: keduanya tidak punya import file soal; salin antar-Form/antar-ujian adalah satu-satunya jalan bawaan. | Guru SMK yang punya bank soal di Word/Excel tidak punya jalan masuk massal. v2: jalur file `.xlsx` adalah jalur utama (§2.2), salin antar-bank adalah fitur terpisah (bukan pengganti). |
| 13 | **Menyalin aset/logo/branding aplikasi lain** | — | Bukan praktik yang ditemukan di sumber; dicatat sebagai batasan tugas ini. | Pelanggaran hak cipta / merek + kebingungan pengguna. Dokumen ini hanya menyebut nama aplikasi lain untuk perbandingan faktual. |

---

## Bab 4 — Daftar pertanyaan terbuka yang masih perlu verifikasi

Pertanyaan di bawah ini **belum terjawab** dari tiga bahan baca + verifikasi kode. Jangan mengarang jawabannya; tugas implementasi harus menjawabnya (atau memutuskan) sebelum/saat membangun, dan mencatat jawabannya di CHANGELOG / dokumen implementasi.

| # | Pertanyaan | Mengapa penting | Petunjuk yang sudah ada |
|---|---|---|---|
| Q1 | Apakah library pembaca `.xlsx` yang akan dipakai (di Go) membaca sel numerik / tanggal / formula sesuai §2.2 (nilai hasil kalkulasi terakhir)? Dan bagaimana perilaku file yang dibuat Google Sheets vs LibreOffice vs Excel untuk sel kosong / sel berisi spasi? | Spesifikasi §2.2 membuat janji perilaku parsing; janji itu harus diuji di tiga produsen file, bukan diasumsikan. | Belum ada riset library `.xlsx` Go di tiga bahan baca. Uji dengan file contoh dari ketiga aplikasi sebelum rilis. |
| Q2 | Apakah master Kisi-kisi dan Kompetensi punya kolom `kode` yang tampil di UI (yang dirujuk kolom `kisi_kisi` / `kompetensi` di §2.3)? Bila tidak, bagaimana guru mengetahui kode yang harus ditulis — dan apakah tugas implementasi harus menambahkan kolom `kode` + menampilkannya di halaman master dulu? | Tanpa kode yang terlihat guru, kolom baru ini tidak bisa diisi dengan benar; error E30/E31 akan membanjiri guru yang menebak kode. | TERBUKTI kode: `soal` punya `kisi_kisi_id` + `kompetensi_id` (UUID). TIDAK DIKETAHUI: bentuk tampilan/identifikasi kisi-kisi & kompetensi di UI master. |
| Q3 | Apakah `media_url` boleh menunjuk ke file di storage sekolah sendiri (path relatif / URL internal LAN), atau hanya URL `http(s)` eksternal? Bila path internal diizinkan, bagaimana aturan path-nya (agar tidak menjadi celah path traversal / SSRF saat dirender)? | §2.3 membatasi ke URL `http(s)` absolut. Bila sekolah ingin memakai gambar dari server lokalnya sendiri, spesifikasi ini terlalu sempit — tetapi melonggarkannya tanpa aturan adalah risiko keamanan. | ANALISIS.md Bab 5.5 menyebut perpustakaan media terpusat per sekolah (bukan menempel di satu soal) — belum terwujud sebagai fitur. Keputusan Q3 menentukan apakah `media_url` cukup untuk tahap ini. |
| Q4 | Apakah pratinjau (§2.8) harus menampilkan SEMUA baris ter-parse (maks 500) atau cukup 100 pertama + ringkasan? Spesifikasi ini memilih "semua" — apakah itu masih nyaman di layar guru (performa render + scroll)? | Mempengaruhi implementasi UI preview (paginasi / virtualisasi tabel). | Tidak ada data UX dari tiga bahan baca; uji dengan file 500 baris sungguhan. |
| Q5 | Apakah mode `all_or_nothing` + pilihan `duplikat=simpan_ganda` perlu dibatasi (mis. guru tidak boleh menyimpan ganda dalam mode tolak-semua)? Spesifikasi ini mengizinkan kombinasinya (duplikat = warning, bukan error). Apakah koordinator setuju? | Menentukan apakah ada kombinasi mode yang dilarang di UI. | Keputusan spesifikasi (§2.7); perlu persetujuan koordinator, bukan verifikasi teknis. |
| Q6 | Apakah batas 5 MB (§2.2) sudah tepat, atau perlu angka lain (mis. 2 MB / 10 MB)? | Batas ini adalah keputusan baru tanpa preseden di kode lama (tidak ada batas byte). | Tidak ada bukti pembanding yang memberi angka batas ukuran file import soal (ProProfs AI: 20 MB untuk dokumen AI — konteks berbeda). |
| Q7 | Bagaimana perilaku import bila `kisi_kisi` merujuk kisi-kisi milik mapel/tingkat lain (bukan mapel bank tujuan)? Error E30 hanya memeriksa "ditemukan" — apakah perlu pemeriksaan "milik konteks yang sama"? | Mencegah soal terpetakan ke kisi-kisi mapel lain (salah alamat kurikulum yang lolos validasi). | TIDAK DIKETAHUI: model relasi kisi-kisi ↔ bank di kode saat ini belum diverifikasi dalam tugas ini. |
| Q8 | Apakah kolom `kisi_kisi` / `kompetensi` perlu mendukung BANYAK nilai (mis. satu soal → satu kompetensi utama + maks 2 pendukung, seperti ANALISIS.md Bab 5.3)? Spesifikasi ini memakai satu nilai per kolom. | Menentukan apakah skema kolom perlu diubah (satu vs banyak) sebelum implementasi. | ANALISIS.md Bab 5.3: satu kompetensi utama + maks 2 pendukung (pendukung hanya untuk analisis). Kode `soal` hanya punya satu `kompetensi_id` (TERBUKTI). Pendukung berarti kolom/tabel tambahan — di luar tugas ini kecuali diputuskan. |
| Q9 | Untuk `pilihan_gambar`, apakah URL opsi harus diakhiri ekstensi gambar (`.png`/`.jpg`/…) atau cukup URL `http(s)` valid? Spesifikasi ini memilih "cukup URL valid" (E29). Apakah perlu pemeriksaan tipe konten? | Mempengaruhi ketatnya validasi E29. | Tidak ada preseden pembanding yang memeriksa ekstensi URL gambar (Quizizz hanya memakai Image Link URL). |
| Q10 | Apakah endpoint lama `POST /api/v1/bank-soal/import` (JSON `csv_text`) benar-benar harus dipertahankan, atau boleh dihapus setelah UI baru jalan (dengan migrasi klien yang diketahui)? Spesifikasi ini memilih "dipertahankan + deprecated". | Menentukan beban pemeliharaan (dua parser) vs risiko memutus klien lama. | TIDAK DIKETAHUI: apakah ada klien/skrip di luar UI yang memakai endpoint lama. Perlu inventarisasi pemakai sebelum memutuskan hapus. |
| Q11 | (Warisan dari laporan Exo — bukan penghambat tugas import Excel, tetapi jangan diklaim sudah terjawab) Apa arti pasti `ABS` di Format 2 Exo? Bagaimana mekanisme import Google Docs Exo? Bagaimana rumus Word ditangani parser Exo? Apa 3 tipe "16" Exo yang belum teridentifikasi? | Relevan hanya bila v2 suatu hari mengadopsi format Exo / klaim kompatibilitas Exo. Sampai terjawab, v2 tidak mengadopsi `ABS`, Google Docs, atau klaim "16 tipe import" dari Exo. | Lihat `docs/LAPORAN-TIM-EXO-IMPORT.md` Bab 12 (10 pertanyaan terbuka Exo). |
| Q12 | (Warisan dari laporan pembanding — bukan penghambat tugas ini) Apa pesan error persis per baris di Quizizz / ProProfs / Socrative? Apa nama kolom persis ProProfs dan Kahoot? Apa batas soal per file di Moodle / Quizizz / Kahoot / Respondus / Canvas? | Relevan bila perbandingan Bab 1 diperdalam lagi. Tidak menghambat pembangunan import Excel v2. | Lihat `docs/LAPORAN-TIM-PEMBANDING-IMPORT.md` Bab 12. |

---

## Penutup (untuk koordinator)

- **Yang sudah selesai di dokumen ini:** (1) tabel perbandingan 11 aplikasi × 10 dimensi (§Bab 1), dengan koreksi Exo yang menggantikan ringkasan lama; (2) spesifikasi import Excel yang konkret sampai level nama kolom, contoh baris per tipe (16 tipe), katalog pesan error/warning per baris, alur lengkap preview → konfirmasi, panduan pengisian, dan tabel migrasi dari CSV lama (Bab 2); (3) 3 pelajaran wajib adopsi + 13 hal yang tidak boleh ditiru (Bab 3); (4) 12 pertanyaan terbuka (Bab 4).
- **Yang sengaja TIDAK diklaim:** tidak ada klaim "16 tipe import standar", tidak ada klaim "import Google Docs", tidak ada klaim "rumus ikut terimpor", tidak ada kolom template tanpa kolom DB (`level_kognitif`, `estimasi_detik` ditunda dengan alasan eksplisit), tidak ada perilaku parser `.xlsx` yang belum diuji (Q1).
- **Tidak ada kode yang diubah.** Dokumen ini hanya analisis + spesifikasi.
