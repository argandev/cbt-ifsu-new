# AUDIT UI/UX + INFORMATION ARCHITECTURE BARU — CBT SMK v2

> **Tim:** Tim A — Riset + Audit UI/UX
> **Tanggal:** 16 Oktober 2026
> **Ruang lingkup:** `frontend/src/pages/` (35 file) + `frontend/src/components/` (31 file). Kode aplikasi TIDAK diubah; dokumen ini murni audit + usulan.
> **Keluhan pemicu:** UI admin membingungkan, penempatan berantakan. Contoh terverifikasi: `JadwalPage.tsx:142` — form "Buat Jadwal" + "Daftar Jadwal" dalam `grid lg:grid-cols-2` satu halaman sehingga sumpek.
> **Prinsip yang tidak boleh dilanggar:** nama aplikasi tetap **"CBT SMK"**; TANPA logo/nama instansi resmi, TANPA klaim sebagai aplikasi ujian nasional resmi; TANPA menjiplak aset/logo/nama produk referensi. Referensi di bawah dipakai hanya sebagai **pola UI**, bukan untuk ditiru tampilannya.

---

## 0. Ringkasan eksekutif — apa yang sebenarnya salah

Bukan satu halaman yang rusak; **tiga inkonsistensi sistemik** yang membuat seluruh admin terasa berantakan:

1. **Tiga pola "daftar vs form" dipakai bergantian tanpa aturan** — (a) form + daftar berdampingan satu layar (Jadwal, Paket), (b) form dalam modal (Master, Event, Ruangan, UKK-List), (c) form di halaman terpisah (Bank Soal). Pengguna tidak bisa memprediksi "kalau saya klik Tambah, apa yang terjadi: modal? halaman baru? form muncul di samping?".
2. **Komponen bersama sudah dibuat tapi hampir tidak dipakai.** `PageHeader`, `Breadcrumbs`, `Pagination`, `StatusBadge`, `EmptyState`, `Topbar`, `Sidebar` (layout) ada di repo, tetapi **0 halaman** memakai `PageHeader`/`Breadcrumbs`/`Pagination`/`StatusBadge`/`EmptyState` (verifikasi grep, §5). Akibatnya setiap halaman merakit header sendiri → 15+ varian gaya `<h1>` (§5.1).
3. **Dua "dunia" input yang berbeda:** halaman Tim 2 (Master) memakai modal buatan tangan + kelas Tailwind inline yang diulang per file, halaman Tim 3/4 memakai komponen `ui/*` + token `anbk`. Hasilnya: tombol, input, dan dialog terlihat mirip tapi tidak identik, dan perilakunya beda (mis. modal Master tidak pakai komponen `Dialog`, tidak ada tombol tutup X, tidak ada handler Esc — bandingkan `components/ui/Dialog.tsx` yang punya ketiganya).

**Kabar baik — yang sudah benar dan WAJIB dipertahankan:**
- Halaman siswa (Ujian) sudah bagus: sticky header + timer, bar bawah sticky Sebelumnya/Berikutnya/Ragu-ragu/Selesai, grid nomor soal dengan legenda warna (`UjianPage.tsx:209`, `components/exam/NavigasiSoal.tsx`). Jangan dirombak; hanya poles kecil (§3.5).
- `BankSoalListPage` + `BankSoalFormPage` + `BankSoalDetailPage` adalah pola daftar → form-terpisah → detail yang PALING BERSIH di repo ini — jadikan pola acuan untuk Jadwal & Paket (§4.2).
- Target sentuh min 44px sudah konsisten di sebagian besar tombol; `skip link` ("Lewati ke konten") ada di `AdminLayout` & `Layout`; `role="status/alert"` dipakai untuk pesan.

---

## 1. Metode audit

1. Membaca seluruh 35 file `pages/` + 31 file `components/` (struktur JSX, penempatan form/daftar, header, aksi).
2. `grep` terukur: 35 kemunculan `<select` (di luar `Select.tsx` sendiri), 15+ varian `<h1>`, 0 pemakaian `PageHeader`/`Breadcrumbs`/`Pagination`/`StatusBadge`/`EmptyState` di halaman, 3 implementasi modal yang berbeda.
3. Riset pola referensi via pencarian web (§2) — diambil polanya saja.

---

## 2. Riset pola referensi (pola UI, bukan jiplakan)

| Aspek | Pola terbaik yang ditemukan | Sumber |
|---|---|---|
| Daftar vs form (CRUD) | Daftar adalah halaman utama; **tambah/ubah = halaman terpisah atau modal**, tidak pernah berdampingan dengan daftar dalam satu grid. Daftar punya aksi per baris: edit, duplikat, pratinjau, hapus. | Moodle Question Bank: daftar soal + ikon edit/duplikat/pratinjau/hapus per soal; form edit di halaman sendiri (tutorial Spoken-Tutorial / Moodle, https://script.spoken-tutorial.org/index.php/Moodle-Learning-Management-System/C2/Question-bank-in-Moodle/English-timed) |
| Editor soal | Satu **kartu per soal** (judul, tipe, poin, toggle wajib, duplikat, hapus, pindah, drag). Media (gambar/audio/video) adalah lampiran per soal & per opsi, dengan pratinjau + teks alt, bukan kolom URL mentah. Toolbar tambah (soal / judul / gambar / video / seksi) menempel di samping kanvas di desktop, jadi bar aksi bawah di HP. Setelan (jadwal, batas percobaan, mode anti-curang, rilis hasil) TIDAK selalu terlihat — dibuka lewat tab/drawer Setelan. Pratinjau memakai draf yang sama & tidak menyimpan sendiri. | Pola editor formulir/kuesioner (Google Forms): kartu soal + floating toolbar + panel setelan terpisah + pratinjau dari draf (deskripsi pola: https://www.makeuseof.com/tag/best-guide-google-forms/ ; spesifikasi redesign: https://github.com/tuantran-05/edutrack/blob/HEAD/docs/superpowers/specs/2026-06-13-advanced-assignment-google-forms-redesign-design.md) |
| Editor soal (kuis) | Alur: beri nama kuis → tambah soal satu per satu → tiap soal: tulis soal, tambah media, tulis opsi, **tandai jawaban benar**, atur waktu per soal → simpan → terbitkan. Opsi jawaban adalah kartu-kartu, bukan baris teks polos. | Pola editor kuis (Quizizz/Kahoot): editor daftar soal + panel setelan kanan + opsi sebagai kartu + tandai benar (ringkasan: https://englishteaching101.com/kahoot-create-beginner-guide/ ; https://www.scribd.com/presentation/511859135/LED110-QUIZIZZ) |
| Media soal | Pustaka media terpusat: semua gambar yang pernah diunggah tersimpan & bisa dipakai ulang; tambah gambar lewat ikon di editor → popup pilih dari pustaka / unggah baru. | Pola pustaka media kuis (ProProfs Quiz Maker FAQ, User Media Library: https://quiz.proprofs.com/export/download_pdf?topic_id=1962526) |
| Rumus matematika | Editor punya tombol khusus "Σ" yang membuka **editor rumus** (pustaka simbol + pilih rumus → sisipkan). Rumus bisa disisipkan di teks soal MAUPUN di opsi jawaban. Subscript/superscript tersedia di editor yang sama. | Pola editor matematika kuis (ProProfs, "Maths Editor": https://quiz.proprofs.com/video-tutorial-how-to-add-subscript/superscript-to-questions) |
| Rumus — teknologi | Dua peran berbeda: **MathLive** = editor interaktif (ketik rumus dengan keyboard virtual, output LaTeX); **KaTeX** = perender cepat (LaTeX → tampilan). Pola yang benar: simpan LaTeX di data soal, render dengan KaTeX di semua tempat (editor, pratinjau guru, halaman siswa, hasil). Jangan simpan rumus sebagai gambar. | MathLive contoh: https://codesandbox.io/examples/package/mathlive ; KaTeX: https://github.com/iffanmajid/Katex |
| Dropdown / pilihan | Aturan praktis yang dipakai industri: **2–5 opsi → radio/segmented, langsung terlihat** (tanpa dropdown); **>5 opsi → combobox yang bisa dicari** (ketik untuk menyaring); pilihan ganda >5 → daftar checkbox + kolom cari, bukan multi-select dropdown. Native `<select>` tidak bisa dicari — untuk daftar puluhan/ratusan item (mapel, kelas, jadwal, siswa) wajib combobox. | Panduan keputusan komponen (Volt UX): https://github.com/laravolt/volt/blob/HEAD/.claude/skills/volt-ux/SKILL.md ; diskusi filterable select: https://github.com/whatwg/html/issues/11288 ; panduan kontrol Microsoft: https://learn.microsoft.com/nb-no/windows/apps/develop/ui/controls/combo-box |
| Manajemen ujian (guru) | Tabel ujian terbit + **Edit lewat modal** (judul, token, durasi, jadwal, toggle aktif) + hapus. Soal dibuat di editor khusus (Markdown/WYSIWYG + lampiran file per soal). Dashboard live: status tiap siswa, soal aktif, sisa waktu; kontrol tengah ujian: tambah waktu per siswa, paksa submit. | Pola platform ujian (online-exam, exam_box): https://github.com/hambaallahdev01/online-exam ; https://github.com/cofrian/exam_box |

**Yang TIDAK diambil dari referensi:** nama, logo, warna brand, bunyi teks, dan klaim resmi apa pun. Semua referensi di atas hanya membenarkan pola struktur.

---

## 3. Temuan audit per halaman

Legenda pola daftar-vs-form: **[D+F]** = daftar + form berdampingan satu layar (masalah utama) · **[MODAL]** = daftar + form di modal · **[HAL]** = daftar & form halaman terpisah (pola target) · **[FILTER]** = halaman laporan: filter + tabel (bukan CRUD, wajar satu layar) · **[DETAIL]** = halaman detail multi-seksi.

### 3.1 Admin — Dashboard & kerangka

| Halaman | File:baris bukti | Temuan |
|---|---|---|
| Dashboard | `pages/admin/DashboardPage.tsx:22-43` | StatCard menampilkan `value="-"` untuk keempat kartu (Bank Soal, Jadwal, Peserta, Rekap) — dashboard terlihat "kosong/rusak" padahal datanya ada di endpoint tiap modul (komentar file sendiri mengakui "tanpa data palsu"). Menu Cepat hanya 6 dari 19 modul, tanpa ikon, dan duplikat sidebar. **Tidak ada jalan pintas ke tugas paling sering:** buat jadwal, monitor ujian berjalan, koreksi menunggu. |
| Kerangka admin | `components/AdminLayout.tsx` (NAV, ±19 item) | Sidebar SATU daftar datar 19 item tanpa pengelompokan (Jurusan, Kelas, Mapel, Guru, Siswa, Bank Soal, Paket, Jadwal, Monitor, Rekap, Analisis, Kompetensi, Event, Ruangan, Kartu, UKK, Backup, Settings). 4 item memakai ikon `GraduationCap` yang sama (Jurusan/Kelas/Guru/Siswa) dan 3 item memakai `FileText` yang sama (Rekap/Analisis/Kompetensi) — ikon tidak membantu membedakan. Tidak ada breadcrumb di halaman mana pun (§5). |
| Kerangka siswa | `components/Layout.tsx` | Header hanya brand + nama + Keluar. Tidak ada navigasi siswa (Beranda / Kartu Saya) — siswa hanya bisa kembali via tombol browser atau link di dalam kartu. |

### 3.2 Admin — Master Data (pola MODAL buatan tangan, inkonsisten dengan `Dialog`)

| Halaman | Pola | File:baris bukti | Temuan |
|---|---|---|---|
| Jurusan | MODAL | `master/JurusanPage.tsx:239` | Modal dirakit manual (`fixed inset-0 …`), bukan komponen `Dialog`. Tidak ada tombol X, tidak ada handler Esc (bandingkan `Dialog.tsx` yang punya keduanya). Kelas Tailwind `inputCls/btnPrimary/btnGhost/btnDanger` didefinisikan ulang di tiap file Master — 5 definisi duplikat. |
| Kelas | MODAL | `master/KelasPage.tsx:236-288` | Sama seperti Jurusan + filter jurusan di atas tabel memakai `<select>` native (`:139`). |
| Mapel | MODAL | `master/MapelPage.tsx:246-306` | Sama + 4 `<select>` native (`:150,266,273,284`). Field "jurusan" di-disable bila jenis ≠ produktif tanpa penjelasan terlihat (hanya `disabled`). |
| Guru | MODAL | `master/GuruPage.tsx:217-288` | Sama. Field akun bersifat opsional tapi tidak ada penjelasan konsekuensi di form (akun staf hanya bisa dibuat via SQL — keterbatasan ini tidak tertulis di UI). |
| Siswa | MODAL | `master/SiswaPage.tsx:313` | Sama + 4 `<select>` native. Ada 2 filter (kelas, jurusan) di atas tabel tapi tidak ada indikator "filter aktif" / tombol reset. Tombol "Import" adalah link terpisah — pengguna harus tahu Import ada di halaman lain. |
| Import Siswa | HAL (wizard 1 langkah) | `master/ImportSiswaPage.tsx:79` | Polanya sudah benar (pilih file → hasil sukses/gagal per baris). Kekurangan: input file polos tanpa drag-and-drop, tanpa pratinjau isi CSV sebelum import, tanpa validasi ukuran file di client (server menolak >5 MB tanpa peringatan dini). Hasil error ditampilkan sebagai tabel baris — bagus, pertahankan. |

**Keputusan untuk Tim B:** seluruh Master Data memakai **satu** komponen `Dialog` (`components/ui/Dialog.tsx`), bukan modal rakitan. Lihat §4.3.

### 3.3 Admin — Bank Soal (pola HAL — JADIKAN ACUAN)

| Halaman | Pola | File:baris bukti | Temuan |
|---|---|---|---|
| Daftar Bank Soal | HAL (daftar) | `banksoal/BankSoalListPage.tsx:27-127` | Pola paling bersih: header + tombol "Buat Bank Soal" kanan atas → cari → kartu (HP) / tabel (desktop). Kekurangan kecil: header dirakit manual (bukan `PageHeader`), tidak ada breadcrumb "Dashboard / Bank Soal". |
| Form Bank Soal | HAL (form) | `banksoal/BankSoalFormPage.tsx:84-133` | Halaman form terpisah, ada link "Kembali ke daftar" (`:84`). Kekurangan: tidak ada breadcrumb, tidak ada indikator "mode buat vs ubah" selain judul. |
| Detail Bank Soal | DETAIL + MODAL | `banksoal/BankSoalDetailPage.tsx:239-391` | Isi soal di daftar + editor soal di modal (`:325`, rakitan manual, bukan `Dialog`) + pratinjau di modal (`:348`) + panel import CSV inline yang bisa dibuka (`showImport`, `:239`). **Tiga lapisan dalam satu halaman** (daftar soal + import + editor + pratinjau) — masih bisa diikuti karena modal menutupi layar, tapi editor & pratinjau WAJIB pindah ke `Dialog`. Import CSV inline (textarea paste) membingungkan: ada DUA cara import (paste teks di detail + upload file di tempat lain?) — satukan jadi satu: upload file, dengan contoh format. |
| SoalEditor (komponen) | — | `components/soal/SoalEditor.tsx:1-668` | **Ini editor soal yang sebenarnya.** Kekuatan: tipe soal dikelompokkan optgroup "5 tipe standar AKM / 11 varian" (`:575-589`), badge kelompok + deskripsi, peringatan "tidak dinilai" (`:598`), kunci ditandai langsung pada opsi (radio/checkbox di kartu opsi, `:122-166`), hitung rumpang otomatis (`:552`). **Kekurangan (P0/P1):** (1) stimulus/pertanyaan/pembahasan masih `<textarea>` polos — TANPA bold/italic/list, TANPA rumus (§4.6); (2) media hanya **input URL** (`:620-624`, placeholder `https://…`) — tidak ada upload file, tidak ada pratinjau, tidak ada alt text; (3) opsi `pilihan_gambar` juga minta **URL gambar** per opsi (`:134`); (4) ganti tipe soal me-reset jawaban+kunci tanpa konfirmasi (`handleTipe`, `:92-95`) — guru bisa kehilangan soal yang sudah diketik; (5) tidak ada tombol Pratinjau di dalam editor (pratinjau hanya ada di halaman detail). |

### 3.4 Admin — Paket & Jadwal (pola D+F — SUMBER KELUHAN UTAMA)

| Halaman | Pola | File:baris bukti | Temuan |
|---|---|---|---|
| Paket Ujian | **D+F** | `paket/PaketPage.tsx:190-343` | **Tiga hal dalam satu layar:** form paket (kiri) + daftar paket (kanan) + "Soal dalam Paket" (bawah, `:343`). Konsekuensi nyata: (1) klik "Ubah" pada paket → form kiri berubah mode tapi daftar tetap di tempatnya — pengguna tidak sadar sedang mengedit karena tidak ada penanda visual kuat selain judul "Ubah Paket"; (2) klik nama paket di daftar → memuat panel soal di bawah — tetapi panel ini juga berisi dropdown bank + daftar soal + tombol tambah, sehingga halaman punya **dua daftar bersarang**; (3) di HP, grid 2 kolom runtuh jadi 1 kolom → urutan menjadi: form panjang → daftar → panel soal (pengguna harus scroll melewati form untuk mencapai daftar). Form memakai 3 `<select>` native (`:207,229,367`). |
| Jadwal Ujian | **D+F** | `jadwal/JadwalPage.tsx:142-382` | **Kasus yang dilaporkan user, terverifikasi.** Form kiri + daftar kanan dalam `grid lg:grid-cols-2` (`:142`). Masalah tambahan yang tidak ada di Paket: (1) tiap item daftar punya **5 tombol aksi** (Generate Token, Status Token, Edit, Pause/Resume, Hapus — lihat `:260-382`) sehingga kartu daftar sangat padat; (2) token yang baru di-generate tampil di dalam kartu daftar (satu-satunya tempat) — mudah terlewat, tidak ada panel khusus "Token terakhir"; (3) tidak ada filter sama sekali (semua jadwal campur: draft/aktif/pause/selesai); (4) tidak ada tautan cepat ke Monitor per jadwal (Monitor adalah halaman terpisah yang harus dibuka manual dengan jadwal_id). Form memakai 3 `<select>` native (`:148,164,215`). |

**Keputusan untuk Tim B:** Jadwal & Paket WAJIB dipecah ke pola HAL (daftar → halaman form terpisah), mengikuti `BankSoalFormPage`. Alasan: formnya panjang (6–9 field + panel soal), bukan 2–3 field. Lihat §4.2 untuk spesifikasi persisnya. Modal DITOLAK untuk kedua halaman ini (terlalu besar untuk modal; panel "Soal dalam Paket" tidak muat di `Dialog wide`).

### 3.5 Admin — Operasional ujian (Monitor, Rekap, Analisis, Kompetensi)

| Halaman | Pola | File:baris bukti | Temuan |
|---|---|---|---|
| Monitor | FILTER + tabel | `monitor/MonitorPage.tsx:170-333` | Strukturnya sudah baik: header + badge status WS + 4 kartu statistik + aksi Pause/Resume + filter pill (Semua/Sedang/Selesai/Diblokir) + tabel. Kekurangan: (1) bila dibuka tanpa `:jadwalId`, halaman hanya menampilkan pesan "Buka monitor dari halaman jadwal" (`:161`) — **jalan buntu** (tidak ada pemilih jadwal di halaman ini); (2) aksi per peserta (unblock) tersembunyi di kolom Aksi tanpa label jelas; (3) usulan unblock pengawas tampil sebagai daftar teks (`:223`) tanpa tombol setuju/tolak di tempat yang sama. |
| Rekap Nilai | FILTER + tabel | `rekap/RekapPage.tsx:118-297` | Header + filter jadwal + 5 kartu statistik + bar aksi (Export CSV, Finalisasi, Muat Ulang, Berita Acara) + tabel. Kekurangan: (1) aksi koreksi manual esai & remedial berada di dalam baris tabel (tombol kecil, mudah terlewat); (2) tidak ada ringkasan "berapa nilai menunggu koreksi" yang menonjol sebelum tabel — hanya satu baris teks amber (`:170`); (3) filter jadwal memakai `<select>` native + tombol "Tampilkan" — pola ini diulang IDENTIK di 4 halaman (§5.3). |
| Berita Acara | FILTER + dokumen | `rekap/BeritaAcaraPage.tsx:76-220` | Pola print sudah benar (`print:hidden`, tombol Print). Kekurangan: tanpa `paramJadwalId`, pengguna harus pilih jadwal dulu — wajar, tapi tidak ada link "kembali ke Rekap". |
| Analisis Butir | FILTER + tabel | `analisis/AnalisisButirPage.tsx:115-281` | Kekurangan: tabel `min-w-[960px]` + 4 kartu statistik + grafik pengecoh — informasinya padat tapi tanpa urutan baca yang jelas (mana yang dilihat dulu: ringkasan atau tabel?). Tidak ada penanda soal "bermasalah" (daya beda negatif / pengecoh tidak berfungsi) yang bisa diklik untuk lompat ke soal. |
| Nilai Kompetensi | FILTER + tabel | `analisis/KompetensiPage.tsx:102-244` | Sama seperti Analisis; heatmap kompetensi (`KompetensiHeatmap.tsx`) ada tapi penempatannya di bawah tabel — heatmap seharusnya DI ATAS tabel sebagai ringkasan. |
| Kartu Peserta (admin) | FILTER + aksi + pratinjau | `kartu/KartuPesertaPage.tsx:272-418` | Tiga fungsi tercampur: pilih jadwal → tugaskan Event & Ruangan → daftar kartu + cetak satu/semua. Penempatan "Tugaskan Event & Ruangan" di tengah halaman membuat alur tidak linear. Pratinjau kartu (`KartuView`, `:100-160`) bagus. Kekurangan: tidak ada pratinjau SEBELUM cetak massal ("Cetak semua" langsung print). |

### 3.6 Admin — Event, Ruangan, UKK, Backup, Settings

| Halaman | Pola | File:baris bukti | Temuan |
|---|---|---|---|
| Event | MODAL (`Dialog`) | `event/EventPage.tsx:248-327` | **Ini pola modal yang BENAR** (pakai `Dialog`, ada tombol X + Esc). Jadikan acuan modal bersama Master. Kekurangan: form di modal memakai 2 `<textarea>` + input biasa tanpa pengelompokan; tidak ada info "event dipakai jadwal mana". |
| Ruangan | MODAL (`Dialog`) | `event/RuanganPage.tsx:207-283` | Sama seperti Event — pola benar. Kekurangan: tidak ada indikator "ruangan dipakai event/jadwal mana" di daftar. |
| UKK — Daftar | MODAL (`Dialog`) | `ukk/UkkListPage.tsx:214-288` | Pola benar (Dialog wide). Kekurangan: form 8 field di modal terasa penuh; tanggal + penguji + status dalam satu kolom vertikal tanpa grup. |
| UKK — Detail | DETAIL (3 kartu) | `ukk/UkkDetailPage.tsx:149-358` | Kartu "Penguji & Status" + "Nilai Peserta" + "Berkas Praktik" — pemisahan sudah baik. Kekurangan: (1) upload berkas (pilih siswa + pilih file + Upload) berada di DALAM kartu "Berkas Praktik" bersama tabel — wajar, tapi tidak ada daftar berkas per siswa; (2) tombol "Input Nilai" ada di kartu 1, tetapi tabel nilai di kartu 2 — pengguna mengira tombol itu mengubah tabel di tempat; (3) tabel nilai punya 11 kolom (`:210-230`) — di HP tidak terbaca meski `overflow-x-auto`. |
| UKK — Input Nilai | FILTER (pilih siswa) + FORM + tabel | `ukk/UkkNilaiPage.tsx:181-360` | Alur: pilih siswa → form 4 nilai (Teori/Proses/Produk/Sikap) + catatan → tabel riwayat. Kekurangan: (1) 4 input nilai dalam grid 4 kolom menyamarkan bobot (Teori 40%, Praktik 60% = Proses 30 + Produk 50 + Sikap 20) — tidak ada kalkulasi pratinjau "Nilai akhir = …" sebelum simpan; (2) tidak ada validasi rentang terlihat (0–100) di label; (3) tidak ada tombol "Siswa berikutnya" — setelah simpan, guru harus buka dropdown lagi. |
| UKK — Sertifikat | Dokumen | `ukk/UkkSertifikatPage.tsx:61-128` | Halaman dokumen, sudah benar. Kekurangan: tidak ada pratinjau sebelum cetak dari daftar nilai (tombol Sertifikat langsung ke halaman ini). |
| Backup | FILTER + tabel | `backup/BackupPage.tsx:148-372` | Kartu "Buat backup manual" + "Daftar backup" + 3 Dialog (buat/restore/hapus). Pola benar. Kekurangan: tidak ada info kapan backup otomatis terakhir / jadwal backup berikutnya; tombol Restore dan Hapus berdekatan di satu baris tanpa pemisah visual yang kuat (risiko klik salah pada aksi destruktif). |
| Settings | HAL (form panjang) | `settings/SettingsPage.tsx:100-264` | **Tiga masalah P0:** (1) **Logo masih input URL** (`:206-214`, `type="url"`, placeholder `https://...`) — tidak ada upload file, tidak ada pratinjau logo, tidak ada validasi; padahal UKK Detail SUDAH punya upload file (`UkkDetailPage.tsx:304`) — pola upload sudah ada di repo, tinggal dipakai; (2) **KKTP per mapel masih textarea JSON mentah** (`:229-243`, 8 baris, font mono) — guru/admin harus menulis JSON; salah koma = error; tidak ada pemilih mapel, tidak ada field deskripsi/interval per mapel; (3) Mode remedial ditampilkan sebagai teks read-only + `input type="hidden"` (`:254-259`) — field yang tidak bisa diubah tidak boleh tampil seperti setting. Kekurangan lain: tidak ada pengelompokan (semua setting dalam satu kolom `max-w-2xl`); tidak ada tab/seksi "Ujian / Tampilan / Nilai"; tidak ada tombol "Kembalikan default". |

### 3.7 Siswa

| Halaman | File:baris bukti | Temuan |
|---|---|---|
| Beranda siswa | `siswa/SiswaHomePage.tsx:125-215` | Kartu per jadwal (nama paket, waktu, status) + input token per kartu + tombol Masuk/Lanjut/Lihat Hasil. Kekurangan: (1) token diminta di dalam kartu jadwal — bila ada 3 jadwal, ada 3 input token; tidak ada SATU tempat "Masukkan token" yang jelas; (2) tidak ada status "ujian besok / minggu ini" — hanya jadwal hari ini (sesuai API `ujian/aktif`); (3) tidak ada link ke Kartu Saya di header — hanya di dalam kartu (`:82-87`). |
| Ujian | `siswa/UjianPage.tsx:183-718` | **Halaman terbaik di repo — jangan dirombak.** Sticky header biru (identitas + Timer, `:209-215`), bilah info peserta (`:217-242`), grid soal sticky di desktop + drawer di HP (`:260-…`), bar bawah sticky (Sebelumnya / Ragu-ragu / Berikutnya / Selesai, `:608+`). Kekurangan kecil: (1) tombol "Selesai Ujian" berada di bar bawah bersama navigasi — risiko klik tidak sengaja (sudah ada dialog konfirmasi, pertahankan); (2) saat diblokir, halaman menampilkan `BannerPeringatan` tapi tidak ada tombol "Minta unblock ke proktor" (pengawas punya usulan unblock di Monitor, tapi siswa tidak bisa memicunya dari sini); (3) tidak ada indikator "jawaban tersimpan" per soal (autosave 500ms terjadi diam-diam). |
| Hasil | `siswa/HasilPage.tsx:1-158` | Menampilkan nilai + per-soal benar/salah + pembahasan (bila diizinkan). Kekurangan: tidak ada ringkasan atas (nilai, jumlah benar/salah/dilewat) sebelum daftar soal; tidak ada tombol "Kembali ke Beranda" yang menonjol. |
| Kartu Saya | `siswa/KartuSayaPage.tsx:1-254` | Tanpa param: daftar jadwal → pilih → kartu + tombol Print. Kekurangan: pemilih jadwal memakai input teks + daftar (`jadwalInput`, `:70-…`), bukan dropdown/daftar kartu — tidak konsisten dengan pemilih jadwal di halaman admin. |
| Login / Ganti Password | `pages/LoginPage.tsx:61-104`, `pages/GantiPasswordPage.tsx` | Login: Card tengah, `ds-form` — sudah rapi. Kekurangan: tidak ada pesan bantuan "Lupa password? Hubungi proktor/admin" (siswa tidak punya jalur reset); Ganti Password tidak memberi tahu syarat password (panjang/karakter) sebelum submit. |
| Forbidden / NotFound / ErrorBoundary | `pages/ForbiddenPage.tsx`, `pages/NotFoundPage.tsx`, `pages/ErrorBoundary.tsx` | Ada, tapi NotFound tidak memberi tombol "Kembali ke Dashboard/Beranda" sesuai role. |

---

## 4. Information Architecture BARU (usulan konkret — Tim B mengeksekusi ini)

### 4.1 Prinsip penempatan (berlaku untuk SEMUA halaman — tanpa kecuali)

1. **Satu halaman, satu tugas utama.** Daftar dan form TIDAK PERNAH berdampingan dalam satu grid. Pilih salah satu dari dua pola di §4.2/§4.3 berdasarkan jumlah field (aturan pemilih ada di sana).
2. **Page header SERAGAM di semua halaman:** `Breadcrumbs` (di atas) + `PageHeader` (judul + deskripsi 1 baris + aksi kanan). Komponen sudah ada — tinggal dipakai (§5.2). Header jangan lagi dirakit manual.
3. **Tombol aksi utama ("Tambah / Buat …") SELALU di kanan atas header** (di dalam `PageHeader aksi`), BUKAN di bawah form, BUKAN di tengah daftar. Aksi destruktif (Hapus, Restore, Finalisasi) selalu memakai `ConfirmDialog` + dipisah visual dari aksi biasa (jangan berdampingan tanpa jarak).
4. **Aksi per baris di tabel/daftar:** urutan tetap: **Lihat/Detail → Ubah → Duplikat (bila ada) → Hapus**. Ikon + label teks (jangan ikon saja) pada aksi yang jarang dipakai; ikon saja hanya boleh untuk Ubah/Hapus yang sudah dikenal, dengan `aria-label` + `title`.
5. **Filter selalu di ATAS tabel dalam satu bar filter** (cari + dropdown + tombol Tampilkan/Reset), dan hasil filter selalu menampilkan jumlah hasil ("Menampilkan X dari Y"). Filter yang mengubah data (mis. pilih jadwal di Rekap/Analisis/Kompetensi/Berita Acara) memakai komponen yang SAMA (§5.3).
6. **Pesan sukses/error SELALU di bawah header, di atas konten** (`role="status"` / `role="alert"`), dengan gaya yang sama. Jangan menaruh pesan di dalam kartu daftar.
7. **Setiap halaman daftar yang kosong memakai `EmptyState`** (komponen sudah ada, belum dipakai) + tombol aksi yang sesuai ("Belum ada X — Buat X").
8. **Mobile:** aksi utama full-width di bawah judul (sudah didukung `PageHeader`); tabel `overflow-x-auto`; form 1 kolom; modal menjadi bottom-sheet (sudah didukung `Dialog`).

### 4.2 Pola DAFTAR vs FORM — aturan pemilih + spesifikasi per halaman

**Aturan pemilih (mengikat):**

| Jumlah field form | Pola | Contoh di repo |
|---|---|---|
| 1–4 field sederhana | **MODAL** dengan `Dialog` | Master Data, Event, Ruangan, UKK-Baru |
| ≥5 field, atau form berisi sub-daftar (mis. soal dalam paket), atau form berisi editor panjang | **HALAMAN TERPISAH** (`/…/baru`, `/…/:id/ubah`) dengan breadcrumb + tombol Kembali | Bank Soal (acuan), Jadwal (baru), Paket (baru) |
| Bukan CRUD (laporan) | **FILTER + tabel**, tanpa form tambah di halaman yang sama | Rekap, Analisis, Kompetensi, Berita Acara, Monitor, Backup |

**Spesifikasi Jadwal (memecah `JadwalPage.tsx:142`):**
- `JadwalPage` (daftar) → header (`PageHeader` + tombol "Buat Jadwal" kanan atas) + filter (status: Semua/Draft/Aktif/Pause/Selesai + cari nama paket) + daftar sebagai **tabel** (kolom: Paket, Kelas, Tanggal, Jam, Status, Token, Aksi) menggantikan kartu `<ul>`. Token TIDAK lagi di-generate dari kartu daftar: kolom Token hanya menampilkan status ("Belum dibuat / Aktif s.d. …") + tombol "Token" yang membuka **Dialog Token** (generate → tampil sekali + tombol Salin + hitung mundur kedaluwarsa). Baris aksi: Monitor (link `/adm/monitor/:id`, ikon Monitor) → Ubah → Hapus. Pause/Resume PINDAH ke Monitor (tempat aksi itu berada) — bukan di daftar Jadwal.
- `JadwalFormPage` (BARU, route `/adm/jadwal/baru` & `/adm/jadwal/:id/ubah`) → breadcrumb `Dashboard / Jadwal / Buat Jadwal`; form 6 field (Paket, Kelas, Tanggal, Jam mulai, Jam selesai, Status) dalam 1 kolom `max-w-2xl`; tombol Simpan + Batal (Batal = kembali ke daftar, tanpa menyimpan). Setelah simpan → kembali ke daftar + pesan sukses.
- Route lama `/adm/jadwal` tetap daftar; tidak ada lagi `grid lg:grid-cols-2` form+daftar.

**Spesifikasi Paket (memecah `PaketPage.tsx:190`):**
- `PaketPage` (daftar) → header + tombol "Buat Paket" + tabel paket (kolom: Kode, Nama, Mapel, Jenis, Durasi, Jumlah soal, Aksi). Klik baris → halaman detail paket.
- `PaketFormPage` (BARU, `/adm/paket/baru`, `/adm/paket/:id/ubah`) → form paket (Kode, Nama, Mapel, Jenis, Durasi, 3 checkbox) — TANPA panel soal.
- `PaketDetailPage` (BARU, `/adm/paket/:id`) → breadcrumb `Dashboard / Paket Ujian / {nama}`; seksi 1 "Info paket" (ringkasan + tombol Ubah); seksi 2 "Soal dalam Paket" (dropdown bank → daftar soal → "+ Tambah" per soal → tombol "Generate Acak") — panel soal yang sekarang di `PaketPage.tsx:343` PINDAH ke sini utuh. Ini memisahkan "menyusun paket" dari "mengisi soal" menjadi dua tugas yang tidak saling mengganggu.

**Spesifikasi Master Data (menyeragamkan ke `Dialog`):**
- Daftar tetap; Tambah/Ubah membuka `Dialog` (bukan modal rakitan). `Dialog` dipakai dengan `title` "Tambah X / Ubah X", tombol Simpan (primer, kanan bawah) + Batal (sekunder). Hapus memakai `ConfirmDialog` (bukan modal rakitan kedua seperti di `KelasPage.tsx:288`).
- Form Master WAJIB memakai komponen `ui/Input`, `ui/Select` (→ combobox, §4.5), `ui/Button` — hapus definisi `inputCls/btnPrimary/btnGhost/btnDanger` yang diulang per file.
- Tombol "Import" di `SiswaPage` menjadi tombol sekunder di header ("Import CSV") yang menuju `/adm/master/siswa/import`; halaman Import mendapat breadcrumb `Dashboard / Siswa / Import CSV` + pratinjau CSV (§4.4).

**Spesifikasi Bank Soal (dipertahankan + poles):**
- Struktur daftar → form → detail TETAP. Yang berubah: editor & pratinjau soal di `BankSoalDetailPage.tsx:325,348` pindah ke `Dialog` (editor = `Dialog wide` + tombol Pratinjau di dalamnya; pratinjau = `Dialog` kedua atau panel dalam editor — pilih panel dalam editor agar tidak ada modal ganda). Import CSV disatukan ke upload file (§4.4). Tambah `PageHeader` + breadcrumb di ketiga halaman.

### 4.3 Pengelompokan menu sidebar BARU (menggantikan 19 item datar)

Struktur (grup + item; label & route TIDAK berubah, hanya dikelompokkan; permission filter tetap berlaku per item):

```
Dashboard                         /adm
— UJIAN —
Bank Soal                         /adm/bank-soal
Paket Ujian                       /adm/paket
Jadwal                            /adm/jadwal
Monitor                           /adm/monitor
— HASIL & ANALISIS —
Rekap                             /adm/rekap
Analisis Butir                    /adm/analisis
Nilai Kompetensi                  /adm/kompetensi
— PESERTA & PELAKSANAAN —
Siswa                             /adm/master/siswa
Kartu Peserta                     /adm/kartu
Event                             /adm/event
Ruangan                           /adm/ruangan
UKK                               /adm/ukk
— DATA INDUK —
Jurusan                           /adm/master/jurusan
Kelas                             /adm/master/kelas
Mapel                             /adm/master/mapel
Guru                              /adm/master/guru
— SISTEM —
Backup                            /adm/backup
Settings                          /adm/settings
```

Aturan: judul grup kecil, uppercase, `text-xs`, tidak bisa diklik; grup yang semua itemnya tersembunyi karena permission ikut tersembunyi. Ikon: tiap item ikon BERBEDA (jangan 4× `GraduationCap`): Jurusan=`Building`/`Landmark`, Kelas=`Users`, Mapel=`BookOpen`, Guru=`Briefcase`/`UserCog`, Siswa=`GraduationCap`; Rekap=`FileText`, Analisis=`BarChart3` (sudah dipakai di halaman), Kompetensi=`Target` (sudah dipakai di halaman). Implementasi: ubah `NAV` di `AdminLayout.tsx` menjadi daftar grup; JANGAN memakai komponen `layout/Sidebar.tsx` yang belum terpakai — hapuskan duplikasi dengan memakai SATU sumber menu (lihat §5.2).

**Header siswa:** tambah nav `Beranda (/)` + `Kartu Saya (/kartu)` di `Layout.tsx` (tampil di desktop; di HP masuk menu yang sama dengan tombol Keluar).

### 4.4 Pola-pola baru yang WAJIB dipakai (spesifikasi untuk Tim B/C/D)

1. **Upload file (menggantikan input URL):** berlaku untuk Logo (Settings), Media soal (SoalEditor), gambar opsi `pilihan_gambar`. Pola: area upload (klik / drag-and-drop) → pratinjau gambar (thumbnail) atau nama file + ukuran → tombol Ganti/Hapus → validasi di client (tipe + maks ukuran, tampilkan sebelum upload). Logo: pratinjau logo saat ini + "Upload logo baru". Media soal: simpan sebagai file terunggah (URL hasil upload), BUKAN URL eksternal yang diketik; field URL lama dipertahankan sebagai fallback "atau tempel URL" di bawah upload (agar soal lama yang memakai URL tetap jalan). Backend: Tim D menyiapkan endpoint upload (mengikuti pola upload berkas UKK yang sudah ada: `UkkDetailPage.tsx:285-330`, `type="file"`, maks 10 MB) — Tim B tidak boleh mengarang endpoint baru tanpa kontrak Tim D.
2. **Import CSV (satu pola):** file picker + pratinjau 5 baris pertama + hitung baris + tombol Import + hasil sukses/gagal per baris (pola `ImportSiswaPage` dipertahankan). Bank Soal: hapus textarea paste (`BankSoalDetailPage.tsx:239-…`); ganti upload file + pratinjau + contoh format yang bisa diunduh.
3. **KKTP (menggantikan textarea JSON):** Settings → seksi "KKTP per Mapel": pemilih Mapel (combobox) → daftar TP untuk mapel itu, tiap TP punya field Deskripsi + Interval (mis. `70–100`) + tombol Tambah TP / Hapus TP → Simpan. TIDAK ADA lagi textarea JSON. Data yang disimpan SAMA (objek `kktp_per_mapel`) — hanya UI yang berubah.
4. **Pemilih jadwal SERAGAM** di Rekap / Analisis / Kompetensi / Berita Acara / Kartu (admin): satu komponen `PemilihJadwal` (combobox + tombol Tampilkan + tombol Reset) — menggantikan 4 salinan `<select>` + tombol "Tampilkan" yang identik (§5.3).
5. **Token ujian:** Dialog Token di Jadwal (§4.2) + di Beranda Siswa SATU input token terpusat tidak perlu diubah (token per kartu jadwal dipertahankan karena tiap jadwal tokennya beda — tapi tambahkan label jadwal yang jelas di atas tiap input).
6. **Pratinjau nilai UKK:** di `UkkNilaiPage`, tampilkan hitung mundur langsung "Nilai akhir (pratinjau) = Teori×40% + Praktik×60%" saat keempat nilai diisi, sebelum tombol Simpan. Tambah tombol "Simpan & Siswa Berikutnya".

### 4.5 Dropdown → combobox + radio (menggantikan 35 `<select>` native)

**Aturan (mengikat, dari riset §2):**
- **≤5 opsi, pilihan tunggal, perlu dibandingkan → radio/segmented, bukan dropdown.** Berlaku untuk: Status jadwal (draft/aktif/pause/selesai), Jenis paket (PTS/PAS/USBK/UKK/Tryout), Level soal (mudah/sedang/sulit), Status soal (draft/aktif/nonaktif/perlu_revisi/diarsipkan), Filter status Monitor (sudah pill — pertahankan). Jangan ubah jadi dropdown lagi.
- **>5 opsi ATAU daftar dinamis dari server (mapel, kelas, jurusan, siswa, jadwal, bank soal, paket) → Combobox** (ketik untuk cari + keyboard navigasi + `aria-expanded/aria-activedescendant` + fallback ke `<select>` native di layar sentuh kecil bila perlu). Satu komponen `ui/Combobox` dipakai di SEMUA tempat — jangan buat combobox kedua.
- **Tipe soal (16 opsi, 2 kelompok) → tetap `optgroup`, tapi dalam Combobox berkelompok** (grup "5 tipe standar AKM" / "11 varian") + deskripsi tipe tampil di bawah pemilih (pola yang sudah ada di `SoalEditor.tsx:575-589` dipertahankan).
- **Filter jadwal di halaman laporan → `PemilihJadwal` (§4.4.4) yang memakai Combobox.**
- **Daftar yang TIDAK diubah:** `type="date"`, `type="time"`, `type="number"` tetap native (sesuai riset: kontrol native mempertahankan perilaku platform).

**Daftar `<select>` yang harus diganti (35 kemunculan, hasil grep):** `JadwalPage` 3, `PaketPage` 3, `BankSoalDetailPage` 3 (filter tipe/level + …), `BankSoalFormPage` 2, `KelasPage` 3, `MapelPage` 4, `SiswaPage` 4, `RekapPage` 1, `BeritaAcaraPage` 1, `AnalisisButirPage` 1, `KompetensiPage` 1, `KartuSayaPage` 1, `SoalEditor` 5, `exam/SoalRenderer` 1. Perhatian: `components/ui/Select.tsx` sendiri berisi 2 `<select>` (definisi komponen) — itu yang diganti implementasinya menjadi Combobox, bukan dihapus.

### 4.6 Editor soal — rich text + rumus + media (spesifikasi untuk Tim C)

**Struktur editor BARU (satu kolom, urutan tetap):**
1. Tipe soal (Combobox berkelompok + badge + deskripsi + peringatan tidak-dinilai) — dipertahankan.
2. **Stimulus** — rich editor (§4.6.1) menggantikan textarea.
3. **Pertanyaan** — rich editor + tombol rumus (§4.6.2) menggantikan textarea.
4. **Media** — upload (§4.4.1) + pratinjau + alt text, menggantikan input URL.
5. **Opsi / jawaban dinamis per tipe** — dipertahankan, dengan perubahan: tiap opsi PG mendapat tombol rumus + tombol gambar (upload, untuk `pilihan_gambar` WAJIB upload, bukan URL); ganti tipe soal → **dialog konfirmasi** "Ganti tipe akan mengosongkan jawaban & kunci. Lanjutkan?" (menggantikan reset diam-diam di `SoalEditor.tsx:92-95`).
6. **Pembahasan** — rich editor (tanpa rumus — cukup format dasar).
7. **Baris meta:** Bobot + Level (radio) + Status (radio/dropdown kecil) + tombol **Pratinjau** (panel pratinjau di dalam Dialog editor, memakai `soal/SoalRenderer`) + Simpan / Batal.

**4.6.1 Rich editor — spesifikasi minimal (jangan over-engineering):** toolbar: Tebal, Miring, Garis bawah, Daftar bullet, Daftar nomor, Hapus format + tombol Rumus (Σ). Disimpan sebagai **HTML yang disanitasi** ATAU **Markdown** — Tim C memilih SATU format dan menuliskannya di kontrak; format harus sama untuk stimulus, pertanyaan, pembahasan, dan opsi. Render di siswa & hasil memakai renderer yang sama (jangan `dangerouslySetInnerHTML` mentah — sanitasi wajib; konten lama teks polos harus tetap tampil benar). TIDAK perlu: warna teks, font picker, tabel, embed video di tahap ini.

**4.6.2 Rumus matematika — spesifikasi:** tombol Σ di toolbar membuka popover editor rumus (**MathLive** — ketik rumus dengan keyboard matematika, pratinjau langsung) → Sisipkan → rumus tersimpan sebagai **LaTeX** (`$…$` inline / `$$…$$` blok) di dalam konten → **KaTeX** me-render di editor (pratinjau), pratinjau guru (`soal/SoalRenderer`), halaman siswa (`exam/SoalRenderer`), dan halaman hasil. Fallback: bila KaTeX gagal parse, tampilkan LaTeX mentah (jangan halaman kosong). Cakupan: stimulus, pertanyaan, opsi PG, pembahasan. TIDAK perlu: editor rumus tulisan tangan, grafik.

### 4.7 Breadcrumb + Page Header — spesifikasi persis (Tim B)

Setiap halaman admin & siswa memakai, dalam urutan ini:
```
<Breadcrumbs items={[...]} />        // "Dashboard / … / Halaman ini"
<PageHeader judul="…" deskripsi="…" aksi={<…>} />
```
- Breadcrumb: Dashboard = `/adm` (admin) atau Beranda = `/` (siswa). Item terakhir = halaman aktif (tanpa link). Contoh Jadwal: `Dashboard / Jadwal`; Form Jadwal: `Dashboard / Jadwal / Buat Jadwal`; Detail Bank: `Dashboard / Bank Soal / {nama bank}`; Detail UKK: `Dashboard / UKK / {nama UKK}`; Input Nilai UKK: `Dashboard / UKK / {nama} / Input Nilai`.
- `PageHeader`: judul = nama halaman (bukan nama entitas — nama entitas masuk deskripsi); deskripsi = 1 kalimat apa yang bisa dilakukan di halaman ini; `aksi` = tombol utama (+ tombol sekunder bila ada, maks 2). H1 hanya ada SATU per halaman, yaitu di dalam `PageHeader` (hapus `<h1>` rakitan di tiap halaman).
- `components/layout/Breadcrumbs.tsx` perlu satu perbaikan kecil: link Beranda memakai `/` — untuk admin harusnya `/adm`. Tim B: tambahkan prop `berandaTo` (default `/`) atau tentukan dari prefix route; jangan hardcode dua komponen.

### 4.8 Penempatan tombol aksi — matriks

| Konteks | Penempatan | Contoh |
|---|---|---|
| Aksi utama halaman (Tambah/Buat) | Kanan atas header (`PageHeader aksi`) | "Buat Jadwal", "Buat Paket", "Buat Bank Soal" |
| Simpan/Batal form halaman terpisah | Bawah form, Simpan primer kiri + Batal sekunder; di form panjang, ulangi di sticky bar bawah bila perlu | JadwalForm, PaketForm, BankSoalForm |
| Simpan/Batal modal | Footer modal: Batal (kiri) / Simpan (kanan), atau sesuai `Dialog` | Master, Event, UKK-Baru |
| Aksi per baris tabel | Kolom Aksi paling kanan, urutan Lihat → Ubah → Duplikat → Hapus | Semua daftar |
| Aksi destruktif | Selalu `ConfirmDialog`; jangan di sebelah tombol primer tanpa pemisah; Restore/Hapus Backup dipisah baris/grup | Backup, Finalisasi Rekap |
| Aksi dalam kartu daftar (Jadwal lama) | DILARANG — pindah ke tabel / Dialog Token / Monitor | Jadwal |
| Aksi ujian siswa | Bar bawah sticky (Sebelumnya/Ragu/Berikutnya/Selesai) — dipertahankan | UjianPage |
| Print / Export | Bar aksi di bawah header, di samping filter, `print:hidden` | Rekap, Berita Acara, Kartu |

---

## 5. Temuan lintas halaman (sistemik)

### 5.1 Header `<h1>` — 15+ varian, 0 `PageHeader`

Hasil grep `<h1` di seluruh `pages/` + `components/` (35 file): varian yang ditemukan antara lain `border-b-2 border-anbk pb-2 text-xl font-bold text-anbk` (9×), `text-2xl font-bold text-white` (4×, di atas banner `bg-anbk`), `border-b-2 … text-2xl …` (4×), `text-xl font-bold text-white` (2×), `text-xl font-bold text-slate-900` (3×), `ds-page-title` (1× — hanya di dalam `PageHeader` yang tidak dipakai halaman mana pun). Artinya: **tidak ada dua halaman yang headernya persis sama.** Perbaikan: §4.7 — satu komponen, satu varian.

### 5.2 Komponen bersama yang sudah ada tapi menganggur / ganda

| Komponen | Status | Tindakan |
|---|---|---|
| `common/PageHeader.tsx` | Ada, **0 halaman memakai** | WAJIB dipakai semua halaman (§4.7) |
| `layout/Breadcrumbs.tsx` | Ada, **0 halaman memakai** | WAJIB dipakai semua halaman; tambah prop `berandaTo` |
| `common/Pagination.tsx` | Ada, **0 halaman memakai** | Pakai di semua daftar bertabel (Siswa, Kelas, dst. yang punya `meta.total_pages`) |
| `common/StatusBadge.tsx` | Ada, **0 halaman memakai** | Pakai untuk semua badge status (jadwal, UKK, attempt) — hapus span badge rakitan per halaman |
| `common/EmptyState.tsx` | Ada, **0 halaman memakai** | Pakai untuk semua daftar kosong |
| `common/SearchInput.tsx` | Dipakai 2 halaman (BankSoal List/Detail, dan itu pun hanya impor — cek pemakaian nyata) | Jadikan SATU komponen cari di semua daftar |
| `layout/Topbar.tsx`, `layout/Sidebar.tsx` | Ada, **0 halaman memakai** (AdminLayout merakit sendiri) | KEPUTUSAN: jangan adopsi keduanya apa adanya — `Sidebar.tsx` menduplikasi menu `AdminLayout`. Satu sumber menu (§4.3). Hapus file ganda atau jadikan `Sidebar.tsx` sebagai SATU-SATUNYA menu dan `AdminLayout` memakainya — pilih satu, jangan dua. |
| `ui/Dialog.tsx` vs modal rakitan | `Dialog` dipakai Event/Ruangan/UKK/Backup; Master + BankSoalDetail merakit modal sendiri | SATUKAN ke `Dialog` (§4.2) |
| `soal/SoalRenderer.tsx` vs `exam/SoalRenderer.tsx` | Dua komponen bernama sama, beda tugas (guru read-only vs siswa interaktif) — sengaja dipisah, komentar di file menjelaskan | PERTAHANKAN pemisahan; jangan digabung. Tapi beri nama file yang membedakan di import (alias `SoalRendererGuru` / `SoalRendererSiswa`) agar tidak tertukar |
| Token warna | `index.css:89-118`: `--color-primary` == `--color-anbk` == `#1e4e9e`, plus `primary-dark` vs `anbk-dark` yang NILAINYA BEDA (`#163d7d` vs `#173d7a`) | SATUKAN: satu nama token (`primary`), hapus alias `anbk` secara bertahap; jangan menambah token baru |

### 5.3 Pola filter "pilih jadwal + Tampilkan" disalin 4×

`RekapPage.tsx:124-143`, `AnalisisButirPage.tsx:124-141`, `KompetensiPage.tsx:112-130`, `BeritaAcaraPage.tsx:103-119` — struktur IDENTIK (label sr-only + `<select>` + tombol "Tampilkan"). Ganti ke satu komponen `PemilihJadwal` (§4.4.4). Ini juga menghilangkan 4 dari 35 `<select>` native.

### 5.4 Inkonsistensi lain yang harus dibereskan Tim B

- **Wadah halaman:** sebagian halaman memakai `<div className="space-y-6 p-4 sm:p-6">`, sebagian `<main className="mx-auto max-w-6xl px-4 py-6">`, sebagian `max-w-2xl/3xl/7xl`. SATUKAN: konten admin = `max-w-6xl` (form = `max-w-2xl` di dalamnya); konten siswa = `max-w-[1100px]` (sudah di `Layout.tsx` — pertahankan).
- **Tabel:** sebagian memakai komponen `ui/Table`, sebagian `<table>` mentah dengan header `bg-anbk text-white`, sebagian `divide-y`. SATUKAN ke `ui/Table` + header navy + `StatusBadge`.
- **Pesan sukses/error:** sebagian `<p role="status">`, sebagian `pesan` state lokal tanpa role, sebagian di dalam kartu. SATUKAN ke posisi §4.1.6.
- **Tanggal/jam:** Jadwal menampilkan `tanggal` mentah (`YYYY-MM-DD`) di daftar (`JadwalPage.tsx:254`), sementara Kartu memakai `formatTanggal` Indonesia. SATUKAN: satu util format tanggal Indonesia (`d MMMM yyyy`, `HH:mm`) dipakai di semua halaman.
- **Loading/empty:** sebagian halaman menampilkan "Memuat…" teks polos, sebagian tidak ada skeleton, sebagian daftar kosong tanpa `EmptyState`. Minimal: `EmptyState` untuk kosong + teks "Memuat…" yang konsisten.
- **Aksesibilitas (cepat):** semua `<select>`/input yang diganti Combobox WAJIB punya label terasosiasi (yang sekarang sudah `sr-only` di filter — pertahankan pola itu); tombol ikon-only WAJIB `aria-label`; modal WAJIB fokus trap minimal (fokus ke judul saat dibuka, kembali ke pemicu saat ditutup) — `Dialog` saat ini belum punya focus trap, tambahkan di Tim B.

---

## 6. Prioritas perbaikan untuk Tim B–E

### P0 — Wajib dulu (menghilangkan keluhan utama; tanpa ini UI tetap terasa berantakan)

| # | Pekerjaan | Pemilik | Bukti / acuan |
|---|---|---|---|
| P0-1 | Pecah Jadwal: daftar (tabel + Dialog Token + filter status) + `JadwalFormPage` terpisah; hapus `grid lg:grid-cols-2` form+daftar | Tim B | §4.2, `JadwalPage.tsx:142` |
| P0-2 | Pecah Paket: daftar + `PaketFormPage` + `PaketDetailPage` (panel soal pindah ke detail) | Tim B | §4.2, `PaketPage.tsx:190,343` |
| P0-3 | `PageHeader` + `Breadcrumbs` di SEMUA halaman; hapus `<h1>` rakitan; satu `<h1>` per halaman | Tim B | §4.7, §5.1 |
| P0-4 | Sidebar dikelompokkan (§4.3) + ikon dibedakan + nav siswa (Beranda/Kartu Saya) | Tim B | §4.3 |
| P0-5 | Semua modal Master + BankSoalDetail pindah ke `Dialog`; hapus modal rakitan; Hapus → `ConfirmDialog` | Tim B | §4.2, `JurusanPage.tsx:239`, `BankSoalDetailPage.tsx:325` |
| P0-6 | Satu komponen `Combobox` gantikan `ui/Select` + seluruh `<select>` native (35 buah); aturan radio ≤5 opsi | Tim B | §4.5 |
| P0-7 | `PemilihJadwal` seragam gantikan 4 salinan filter jadwal | Tim B | §5.3 |
| P0-8 | Settings: logo → upload + pratinjau (hapus input URL); KKTP → editor per-mapel/per-TP (hapus textarea JSON); hapus field remedial read-only dari form | Tim B (+ Tim D untuk endpoint upload) | §3.6, `SettingsPage.tsx:206-259` |

### P1 — Penting (menyelesaikan janji "soal dukung rumus & media")

| # | Pekerjaan | Pemilik | Acuan |
|---|---|---|---|
| P1-1 | SoalEditor: media → upload + pratinjau + alt text (ganti input URL); opsi `pilihan_gambar` → upload | Tim C (+ Tim D endpoint) | §4.4.1, `SoalEditor.tsx:620-624,134` |
| P1-2 | SoalEditor: rich editor (tebal/miring/list/hapus format) untuk stimulus, pertanyaan, pembahasan, opsi | Tim C | §4.6.1 |
| P1-3 | SoalEditor + kedua SoalRenderer + Hasil: rumus LaTeX (MathLive di editor, KaTeX di render) | Tim C | §4.6.2 |
| P1-4 | SoalEditor: konfirmasi sebelum ganti tipe mereset jawaban; tombol Pratinjau di dalam editor | Tim C | §3.3, `SoalEditor.tsx:92-95` |
| P1-5 | Dashboard: angka nyata (dari endpoint modul) + jalan pintas tugas (Buat Jadwal, Monitor berjalan, Koreksi menunggu) | Tim B | §3.1 |
| P1-6 | Monitor: pemilih jadwal di halaman (tanpa jadwal → bukan jalan buntu); tombol setuju/tolak usulan unblock di tempat; aksi Unblock berlabel jelas | Tim B | §3.5 |
| P1-7 | UKK Nilai: pratinjau nilai akhir (40/60) sebelum simpan + tombol "Simpan & Siswa Berikutnya" + label rentang 0–100 | Tim B | §3.6 |
| P1-8 | Import CSV Bank Soal: upload file + pratinjau + template (hapus textarea paste) | Tim B/C | §4.4.2 |
| P1-9 | Satukan token warna (hapus alias `anbk`), wadah halaman, tabel (`ui/Table`), `StatusBadge`, `EmptyState`, `Pagination`, format tanggal Indonesia | Tim B | §5.2, §5.4 |

### P2 — Pelengkap (poles, bisa sesudah P0–P1)

| # | Pekerjaan | Pemilik | Acuan |
|---|---|---|---|
| P2-1 | Siswa: tombol "Minta unblock" di halaman diblokir; indikator "jawaban tersimpan" per soal; ringkasan atas di Hasil; tombol Kembali di NotFound sesuai role; info syarat password di Ganti Password; teks "Lupa password? Hubungi proktor" di Login | Tim B | §3.7 |
| P2-2 | Kartu Peserta (admin): alur linear pilih jadwal → tugaskan → pratinjau → cetak; pratinjau sebelum cetak massal; indikator ruangan dipakai di mana | Tim B | §3.5 |
| P2-3 | Backup: info backup otomatis terakhir/berikutnya; pemisah visual lebih kuat untuk Restore/Hapus | Tim B | §3.6 |
| P2-4 | Analisis/Kompetensi: heatmap di atas tabel; soal bermasalah bisa diklik → lompat ke soal; urutan baca ringkasan → detail | Tim B | §3.5 |
| P2-5 | Import Siswa: drag-and-drop + pratinjau CSV + peringatan ukuran file di client | Tim B | §3.2 |
| P2-6 | Aksesibilitas: focus trap di `Dialog`, `aria-label` semua tombol ikon-only, uji keyboard penuh (Tab/Esc/Enter) di Combobox + modal + editor | Tim B/E (QA) |
| P2-7 | UKK Detail: pemisah visual tombol "Input Nilai" dari tabel; tabel 11 kolom → kolom prioritas + detail expandable di HP | Tim B | §3.6 |

---

## 7. Yang TIDAK boleh diubah (batas untuk Tim B–E)

1. Logika bisnis, perhitungan nilai, bobot, KKTP, remedial, token, anti-curang, permission — TIDAK berubah. Dokumen ini hanya memindahkan/mengelompokkan UI.
2. Branding: nama "CBT SMK", satu biru primer `#1e4e9e`, footer "Bukan aplikasi ujian nasional resmi." — dipertahankan.
3. Halaman Ujian siswa: struktur sticky header + bar bawah + grid soal dipertahankan; hanya poles P2-1.
4. Pemisahan `soal/SoalRenderer` (guru) vs `exam/SoalRenderer` (siswa) dipertahankan — kunci tidak boleh sampai ke browser siswa.
5. Format data yang disimpan (termasuk `kktp_per_mapel`, struktur soal, LaTeX rumus) — perubahan format data HARUS lewat kontrak Tim D + migrasi, tidak boleh diam-diam diubah Tim B/C.
6. Jangan menambah dependensi baru tanpa persetujuan (Combobox, rich editor, MathLive/KaTeX adalah pengecualian yang SUDAH disetujui lewat dokumen ini — Tim C memilih implementasi paling ringan: utamakan tanpa dependensi berat bila pola yang sama bisa dicapai dengan komponen sendiri + KaTeX saja).

---

## 8. Lampiran — inventaris pola per halaman (cepat)

| Halaman | Pola saat ini | Pola target |
|---|---|---|
| Dashboard | Header banner + StatCard `-` + Menu Cepat | §4.1 + angka nyata (P1-5) |
| Jurusan / Kelas / Mapel / Guru / Siswa | Daftar + MODAL rakitan | Daftar + `Dialog` + Combobox |
| Import Siswa | HAL wizard | HAL wizard + pratinjau (P2-5) |
| Bank Soal (daftar) | HAL daftar | HAL daftar + PageHeader/Breadcrumb |
| Bank Soal (form) | HAL form | HAL form + Breadcrumb |
| Bank Soal (detail) | DETAIL + modal rakitan + import paste | DETAIL + `Dialog` + import file (P1-8) |
| Paket | **D+F** | Daftar + Form terpisah + Detail (P0-2) |
| Jadwal | **D+F** | Daftar + Form terpisah + Dialog Token (P0-1) |
| Monitor | FILTER + tabel | FILTER + tabel + pemilih jadwal (P1-6) |
| Rekap | FILTER + tabel | FILTER + tabel + PemilihJadwal |
| Berita Acara | FILTER + dokumen | FILTER + dokumen + PemilihJadwal |
| Analisis Butir | FILTER + tabel | FILTER + ringkasan di atas + tabel (P2-4) |
| Kompetensi | FILTER + tabel | FILTER + heatmap di atas + tabel (P2-4) |
| Event / Ruangan | Daftar + `Dialog` | Tetap (acuan modal) |
| Kartu Peserta | FILTER + aksi + pratinjau | Alur linear (P2-2) |
| UKK Daftar | Daftar + `Dialog` | Tetap + grup field |
| UKK Detail | DETAIL 3 kartu | Tetap + poles (P2-7) |
| UKK Nilai | Pilih siswa + FORM + tabel | Tetap + pratinjau nilai (P1-7) |
| UKK Sertifikat | Dokumen | Tetap |
| Backup | FILTER + tabel + Dialog | Tetap + poles (P2-3) |
| Settings | HAL form panjang | HAL form + seksi + upload logo + editor KKTP (P0-8) |
| Beranda Siswa | Kartu jadwal + token | Tetap + nav header |
| Ujian | Sticky header + grid + bar bawah | Tetap (jangan dirombak) |
| Hasil | Daftar per soal | Tetap + ringkasan atas (P2-1) |
| Kartu Saya | Pilih jadwal + kartu | Tetap + pemilih konsisten |
| Login / Ganti Password | Card tengah | Tetap + teks bantuan (P2-1) |

---

*Dokumen ini adalah acuan Tim B, C, D, E. Bila ada konflik antara dokumen ini dan kebiasaan lama di kode, dokumen ini yang menang — kecuali menyangkut logika bisnis (§7), yang tidak boleh diubah.*
