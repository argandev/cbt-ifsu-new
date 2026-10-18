# VERIFIKASI FASE 3 — 9 Diagram Sequence Baru CBT SMK v2

> **Tanggal:** 18 Okt 2026 · **File:** `docs/uml/cbt-smk-v2.mdj`
> **SHA-256 `.mdj`:** `9b7bbc06260522c2c8e565a1defd33dc47532befd2255ab476a2c8ad31bb27ff`
> (sama di `docs/uml/cbt-smk-v2.mdj`, `your_files/cbt-smk-v2.mdj`, dan
> `goals/cbt-smk-v2-go-react/files/cbt-smk-v2.mdj` — ketiganya diverifikasi)
> **Dasar:** `PEMETAAN-SEQUENCE-BARU.md` (Fase 2) + `ALUR-SEQUENCE-NYATA.md` (Fase 1).
> Dokumen Fase 1/Fase 2 TIDAK diubah. 5 diagram lain TIDAK diubah.
> Perilaku/kode aplikasi TIDAK diubah. exe/zip TIDAK disentuh. ERD TIDAK diubah.

## 1. Validator `validate-mdj-real.py` — LOLOS SEMUA (0 pelanggaran)

```
Pelanggaran : 0
LOLOS SEMUA        (exit 0)
```

Backup `cbt-smk-v2.mdj.pre-sequence-redo.bak` (SHA `c5bd08be…`) dengan
validator yang sama: **0 pelanggaran, LOLOS SEMUA** (tidak regresi).

### 1.1 Satu temuan validator — FALSE POSITIVE, validator yang diperbaiki (bukan diagram)

Jalankan pertama validator melaporkan **18 pelanggaran**, semuanya
duplikat `(parent, nama)` untuk `UMLMessage` — mis. `periksaBatasWaktu()`
2× di D4, `hitungNilaiAkhir()` 2× di D5, `alihkanFokus()` +
`laporkanPelanggaran()` 2× di D7. Kedelapan belas pasang itu adalah
**pemanggilan ulang yang sah**: operasi yang sama dipanggil dua kali
pada titik berbeda (simpan vs minta status; kirim vs koreksi;
pelanggaran ke-1 vs ke-2). Setiap kemunculan dibedakan oleh posisi
urutannya dalam Interaction dan `_id`-nya masing-masing, bukan oleh nama.

Yang diperbaiki adalah **validator**, bukan diagram — mengganti nama
kemunculan kedua (mis. `periksaBatasWaktu2()`) akan MEMALSUKAN desain
(seolah ada operasi kedua yang tidak ada di kode):

* `validate-mdj-real.py`: cek duplikat `(parent, nama)` KINI
  mengecualikan tipe `UMLMessage` saja; semua tipe lain (Class,
  Operation, Attribute, Lifeline, dsb.) tetap dicek penuh; cek `_id`
  unik + `$ref` resolvable tetap berlaku untuk pesan.
* Alasan didokumentasikan di docstring validator (FALSE POSITIVE (4)):
  pesan Sequence adalah kemunculan berurutan dalam waktu, dan StarUML
  7.1.0 sendiri membuka file berisi 18 pasang pesan berulang ini dengan
  status validasi 0 error dan mengekspornya tanpa keluhan.
* Backup validator pra-perubahan: `/tmp/validate-mdj-real.py.pre-fase3.bak`
  (SHA `c0693d21…`) — salinan kerja, bukan bagian repo.

### 1.2 Kontrol negatif (pada salinan di `/tmp`, file asli tidak tersentuh)

| Mutasi (1 field) | Harapan | Hasil |
|---|---|---|
| Nama satu `UMLClass` digandakan (`Role` → `User`) | GAGAL | GAGAL, exit 1, 2 pelanggaran (duplikat + label compartment) |
| Satu `$ref` model diganti ID tak ada | GAGAL | GAGAL, exit 1, 22 pelanggaran (ref menggantung + turunan) |
| Nama satu `UMLOperation` digandakan dalam class yang sama | GAGAL | GAGAL, exit 1, 1 pelanggaran duplikat |
| Nama satu `UMLMessage` digandakan (uji pengecualian) | LOLOS | LOLOS, exit 0 — pengecualian bekerja sesuai desain |

SHA `.mdj` asli tetap `9b7bbc…` sesudah semua kontrol.

## 2. Lima diagram lain TIDAK berubah vs backup

Perbandingan independen subtree di luar Sequence (3153 objek non-Sequence
di kedua file): **himpunan ID identik, 0 objek dengan field sendiri yang
berbeda** (perbandingan mengecualikan daftar anak `ownedElements` /
`ownedViews` / `subViews` — anak dibandingkan sebagai objek sendiri).
Subtree Sequence: lama 248 ID terhapus, baru 1535 ID — cocok dengan klaim
Fase 2. ERD identik.

Bukti tambahan dari ekspor: SVG 5 diagram lain yang diekspor ulang di
Fase 3 **byte-identik** (SHA-256 sama) dengan file lama di
`your_files/uml-svg/` (01–05).

## 3. Verifikasi visual di StarUML ASLI 7.1.0

StarUML 7.1.0 resmi diunduh dari `staruml.io` (API unduhan resmi,
`StarUML_7.1.0_amd64.deb`), diekstrak tanpa instalasi dpkg
(`~/workspace/.tools/staruml/`, agar tidak dihapus replay paket sistem)
+ lib `libsecret-1-0` / `libnotify4` / `libxtst6` resmi Ubuntu noble
diekstrak ke `~/workspace/.tools/libs/`.

* **GUI:** file dibuka (salinan di `/tmp`, TIDAK PERNAH disimpan dari
  GUI — dialog "Save changes?" dijawab **Don't Save**; SHA salinan
  tetap `9b7bbc…`). Model Explorer menampilkan 3 model. **Status bar:
  centang hijau ✅ = 0 error validasi StarUML.**
* **Keterbatasan jujur (sama seperti ronde sebelumnya):** kanvas live di
  GUI TIDAK melukis diagram di lingkungan Xvfb ini (hanya logo
  placeholder). Karena itu bukti visual per diagram adalah **render
  ekspor StarUML sendiri** (SVG via `cli:image`, lalu PNG 2× via
  cairosvg latar putih) + pemeriksaan geometri `.mdj`, bukan tangkapan
  kanvas GUI.
* **Jalan keluar CLI yang menggantung:** CLI `staruml image` menggantung
  tanpa timeout di sandbox ini (dialog error JS StarUML 7.1.0 yang
  memblokir event loop — `showErrorDialog` mengirim 1 argumen ke
  `showErrorBox` yang butuh 2). Diatasi dengan: GUI +
  `--remote-debugging-port=9222`, hancurkan jendela dialog "Error" dari
  luar (XDestroyWindow — aplikasi survive), muat file lewat CDP
  (`app.project.load(...)`), picu handler ekspor yang sama dengan CLI
  (`app.commands.execute('cli:image', ...)`). Prosedur + skrip:
  `docs/uml/export-sequence-via-cdp.py`. File `.mdj` dimuat bersih
  lewat CDP (14 diagram terbaca) — membuktikan error startup BUKAN
  dari isi file, melainkan dari argumen/cek startup.

### 3.1 Pemeriksaan geometri `.mdj` (semua 9 diagram)

* Pitch lifeline seragam 165 px, lifeline pertama x = 100, lebar 140.
* Pesan pertama y = 135, pitch vertikal 60 px, y naik monotonik
  (atas → bawah), semua panah horizontal murni.
* Head/tail semua panah menempel tepat di pusat garis lifeline
  (0 penyimpangan). Label = nama model pesan; lebar label ≤ bentang panah.
* **0 self-message, 0 Note, `showSequenceNumber = false`** di semua
  diagram Sequence baru.

### 3.2 Pemeriksaan render SVG (teks, tumpang tindih, terpotong)

* Semua label pesan hadir: **209/209** (36+20+25+25+29+18+22+18+16).
  Semua kata nama lifeline hadir (nama panjang terlipat 2 baris di
  dalam kotak kepala: "Halaman Ganti Kata" + "Sandi",
  "Layanan Batas" + "Waktu" — terbaca penuh).
* **0 tumpang tindih teks, 0 teks terpotong** tepi kanvas
  (setelah transform SVG diterapkan). Tanpa artefak penomoran pesan.
* PNG 2×: mode RGB, latar putih, **0 piksel gelap di keempat tepi**
  semua PNG (tidak ada konten terpotong).

### 3.3 Hasil per diagram — SEMUA LAYAK, 0 cacat

| # | Diagram | Lifeline / Pesan | Hasil |
|---|---|---|---|
| 06 | Masuk (D1) | 6 / 36 | **LAYAK.** Alur buka → masuk → penjaga → keluar nyambung; reply putus-putus, siaran panah terbuka; label terbaca, tanpa elemen melayang. |
| 07 | Ganti Kata Sandi (D2) | 7 / 20 | **LAYAK.** Cabang wajib-ganti berurutan; hapus semua sesi + kembali ke masuk terbaca. |
| 08 | Mulai Ujian (D3) | 7 / 25 | **LAYAK.** Daftar → token → buat pengerjaan → muat soal tanpa kunci nyambung. |
| 09 | Kerjakan Ujian (D4) | 7 / 25 | **LAYAK.** Jawab → ragu → status berkala; `periksaBatasWaktu()` 2× pada posisi berbeda, keduanya terbaca. |
| 10 | Selesai, Koreksi dan Lihat Hasil (D5) | 8 / 29 | **LAYAK.** Kirim → koreksi guru → lihat hasil; `hitungNilaiAkhir()` 2× (kirim + koreksi) terbaca. |
| 11 | Pantau Ujian Langsung (D6) | 5 / 18 | **LAYAK.** Tiket → cuplikan → detak jantung → minta ulang; varian fallback terbaca sebagai lanjutan. |
| 12 | Anti-Curang (D7) | 6 / 22 | **LAYAK** (diperiksa piksel per piksel atas + bawah). Dua siklus pelanggaran berurutan: peringatan ke-1 → blokir ke-2 (bekukan → tandai diblokir → layar diblokir). |
| 13 | Jeda Jadwal (D8) | 6 / 18 | **LAYAK.** Jeda (bekukan) → lanjut (pulihkan) berpasangan. |
| 14 | Buka Blokir (D9) | 7 / 16 | **LAYAK.** Usulan menunggu → proktor buka blokir → pulihkan sisa beku. |

**Cacat yang belum diperbaiki: TIDAK ADA.** Karena 0 cacat visual,
tidak ada perbaikan skrip `.mdj` yang diperlukan di Fase 3.

Catatan kosmetik (bukan cacat): watermark `UNREGISTERED` berulang di
semua PNG/SVG (StarUML trial). File `.mdj` bersih — ekspor ulang dari
StarUML berlisensi untuk naskah skripsi.

## 4. File ekspor baru (06–14) + file basi yang dihapus

Ekspor: SVG native StarUML (`cli:image` via CDP) + PNG 2× dari SVG via
cairosvg 2.9.1, latar putih. File 01–05 TIDAK diubah.

| # | PNG (`your_files/uml-png/`) | SVG (`your_files/uml-svg/`) | Kanvas SVG |
|---|---|---|---|
| 06 | `06-sequence-d1-masuk.png` (1972×4590) | `06-sequence-d1-masuk.svg` | 986×2295 |
| 07 | `07-sequence-d2-ganti-kata-sandi.png` (2302×2670) | `07-sequence-d2-ganti-kata-sandi.svg` | 1151×1335 |
| 08 | `08-sequence-d3-mulai-ujian.png` (2302×3270) | `08-sequence-d3-mulai-ujian.svg` | 1151×1635 |
| 09 | `09-sequence-d4-kerjakan-ujian.png` (2302×3270) | `09-sequence-d4-kerjakan-ujian.svg` | 1151×1635 |
| 10 | `10-sequence-d5-selesai-koreksi-hasil.png` (2632×3750) | `10-sequence-d5-selesai-koreksi-hasil.svg` | 1316×1875 |
| 11 | `11-sequence-d6-pantau-ujian-langsung.png` (1642×2430) | `11-sequence-d6-pantau-ujian-langsung.svg` | 821×1215 |
| 12 | `12-sequence-d7-anti-curang.png` (1972×2910) | `12-sequence-d7-anti-curang.svg` | 986×1455 |
| 13 | `13-sequence-d8-jeda-jadwal.png` (1972×2430) | `13-sequence-d8-jeda-jadwal.svg` | 986×1215 |
| 14 | `14-sequence-d9-buka-blokir.png` (2302×2190) | `14-sequence-d9-buka-blokir.svg` | 1151×1095 |

**Dihapus (basi, diagram lama):**
`uml-png/06-sequence-3a.png`, `uml-png/07-sequence-3b.png`,
`uml-svg/06-sequence-3a.svg`, `uml-svg/07-sequence-3b.svg`.
Tidak ada lagi diagram Sequence kadaluwarsa yang tersisa.

## 5. Salinan `.mdj`

`your_files/cbt-smk-v2.mdj` (symlink ke
`goals/cbt-smk-v2-go-react/files/cbt-smk-v2.mdj`, symlink dipertahankan)
SHA-256 `9b7bbc06260522c2c8e565a1defd33dc47532befd2255ab476a2c8ad31bb27ff`
= `docs/uml/cbt-smk-v2.mdj`. ✅

## 6. Yang TIDAK dilakukan / batasan jujur

* Tidak ada spot-check kanvas live GUI per diagram (kanvas tidak
  melukis di Xvfb) — digantikan render ekspor StarUML + pemeriksaan
  geometri + inspeksi piksel PNG (D1, D5, D7 dilihat langsung;
  9 diagram dianalisis teks/posisi SVG).
* VM sempat ke-reset di tengah tugas (StarUML + tool X11 hilang);
  lingkungan dibangun ulang dari `~/workspace` yang utuh
  (StarUML diekstrak ke `~/workspace/.tools/` agar tahan reset).
  SHA `.mdj` dan validator terverifikasi ulang sesudah reset.
* Watermark trial tetap ada di ekspor (lihat §3.3).

---

# VERIFIKASI FASE 3b — 21 Diagram Sequence (D1–D21) CBT SMK v2

> **Tanggal:** 18 Okt 2026 · **File:** `docs/uml/cbt-smk-v2.mdj`
> **SHA-256 `.mdj`:** `118caee05be5f425c5c8f6f496f0133235ea55951e74f4a47fc460b7609a59dc`
> (Fase 2b: 21 Diagram Sequence D1–D21, 601 pesan; D1–D9 hanya judul +
> dokumentasi yang berubah, D10–D21 baru).
> Catatan Fase 3 di atas TIDAK diubah — bagian ini DITAMBAHKAN (append).
> Dokumen Fase 1/1b/2/2b TIDAK diubah. 5 diagram lain TIDAK diubah.
> Perilaku/kode aplikasi TIDAK diubah. exe/zip TIDAK disentuh.
> ERD TIDAK diubah.

## 3b.1 Validator `validate-mdj-real.py` — LOLOS SEMUA (0 pelanggaran)

```
Pelanggaran : 0
LOLOS SEMUA        (exit 0)
```

### Kontrol negatif (pada salinan di `/tmp`, file asli tidak tersentuh)

| Mutasi (1 field) | Harapan | Hasil |
|---|---|---|
| Nama satu `UMLClass` digandakan (`Role` → `User`) | GAGAL | GAGAL, exit 1, 2 pelanggaran |
| Satu `$ref` `source` satu `UMLMessage` diganti ID tak ada | GAGAL | GAGAL, exit 1, 2 pelanggaran |
| Nama satu `UMLOperation` digandakan dalam class yang sama (`User`) | GAGAL | GAGAL, exit 1, 1 pelanggaran |
| Nama satu `UMLMessage` digandakan (uji pengecualian) | LOLOS | LOLOS, exit 0 — pengecualian bekerja sesuai desain |

SHA `.mdj` asli tetap `118cae…` sesudah semua kontrol.

## 3b.2 Lima diagram lain TIDAK berubah vs backup `pre-sequence-2b.bak`

Perbandingan independen subtree di luar Sequence (3153 objek
non-Sequence di kedua file): **himpunan ID identik, 0 objek dengan field
sendiri yang berbeda** (mengecualikan daftar anak `ownedElements` /
`ownedViews` / `subViews` — anak dibandingkan sebagai objek sendiri).
Subtree Sequence: backup 1535 ID → sekarang 4071 ID (D1–D9 ID sama,
hanya nama/dokumentasi berubah + 12 diagram baru D10–D21).
ERD identik (35 entitas di kedua file).
D1–D9: ID view sama persis, **0 perbedaan geometri/label** vs backup.

## 3b.3 Verifikasi visual di StarUML ASLI 7.1.0

Prosedur sama seperti Fase 3 (GUI + `--remote-debugging-port=9222`,
dialog "Error" dihancurkan via XDestroyWindow, file dimuat lewat CDP
dari salinan di `/tmp`, TIDAK PERNAH disimpan dari GUI — SHA salinan
tetap `118cae…`; ekspor via handler `cli:image` yang sama dengan CLI,
`docs/uml/export-sequence-via-cdp.py`). File dimuat bersih: 26 diagram
terbaca (5 lain + 21 Sequence). Kanvas live GUI tetap tidak melukis
di Xvfb — bukti visual adalah render ekspor StarUML sendiri (SVG) +
PNG 2× + pemeriksaan geometri `.mdj` + inspeksi piksel PNG.

### Geometri `.mdj` (semua 21 diagram)

* Pitch lifeline seragam 165 px, lifeline pertama x = 100, lebar 140.
* Pesan pertama y = 135, pitch vertikal 60 px, y naik monotonik
  (atas → bawah), semua panah horizontal murni.
* Head/tail semua panah menempel tepat di pusat garis lifeline
  (0 penyimpangan). Label = nama model pesan; lebar label ≤ bentang
  panah (0 pelanggaran, 601/601).
* **0 self-message (model + view), 0 Note, `showSequenceNumber = false`**
  di semua diagram Sequence.

### Render SVG (teks, tumpang tindih, terpotong)

* Semua label pesan hadir: **601/601** (D1–D9: 209; D10–D21: 392).
  Semua nama lifeline hadir (teks penuh atau terlipat 2 baris).
* **0 tumpang tindih teks, 0 teks terpotong** tepi kanvas
  (setelah transform SVG `translate(-90,-30)` diterapkan).
  Tanpa artefak penomoran pesan.
* PNG 2×: mode RGB, latar putih, **0 piksel gelap di keempat tepi**
  semua PNG (tidak ada konten terpotong).
* Alur D10–D21 dibaca pesan per pesan: semuanya koheren
  (buka → minta → periksa izin → simpan → audit → konfirmasi →
  tampilkan; tiap diagram berakhir di tampilkan/konfirmasi).
* Inspeksi piksel langsung: D10 (atas) + D21 (atas + bawah) dilihat
  sebagai gambar — kepala lifeline terbaca, panah menempel di garis
  putus-putus lifeline, label terbaca, tanpa elemen melayang,
  tanpa self-message loop, tanpa Note.

### Hasil per diagram — SEMUA LAYAK, 0 cacat

| # | Diagram | Lifeline / Pesan | Hasil |
|---|---|---|---|
| 06 | Masuk (D1) | 6 / 36 | **LAYAK.** |
| 07 | Ganti Kata Sandi (D2) | 7 / 20 | **LAYAK.** |
| 08 | Mulai Ujian (D3) | 7 / 25 | **LAYAK.** |
| 09 | Kerjakan Ujian (D4) | 7 / 25 | **LAYAK.** |
| 10 | Selesai, Koreksi dan Lihat Hasil (D5) | 8 / 29 | **LAYAK.** |
| 11 | Pantau Ujian Langsung (D6) | 5 / 18 | **LAYAK.** |
| 12 | Anti-Curang (D7) | 6 / 22 | **LAYAK.** |
| 13 | Jeda Jadwal (D8) | 6 / 18 | **LAYAK.** |
| 14 | Buka Blokir (D9) | 7 / 16 | **LAYAK.** |
| 15 | UC02 Kelola Master Data (D10) | 6 / 38 | **LAYAK.** Daftar → tambah → hapus (cek pemakaian) → impor siswa; audit tiap simpan/hapus. |
| 16 | UC03 Kelola Bank Soal (D11) | 6 / 34 | **LAYAK.** Daftar → bank baru → kunci bank (cek pemisahan) → soal baru (cek status bank). |
| 17 | UC04 Impor Soal (D12) | 6 / 26 | **LAYAK.** Templat → pratinjau (cek berkas + status bank) → konfirmasi simpan. |
| 18 | UC05 Susun Paket Ujian (D13) | 5 / 36 | **LAYAK.** Paket baru → tambah soal (sinkron jumlah) → hasilkan acak. |
| 19 | UC06 Kelola Jadwal Ujian (D14) | 5 / 30 | **LAYAK.** Jadwal baru → hapus (cek pengerjaan) → buat token (tampil sekali). |
| 20 | UC11 Rekap Nilai & Berita Acara (D15) | 5 / 36 | **LAYAK.** Rekap → ekspor → berita acara → finalisasi → remedial. |
| 21 | UC12 Analisis Butir Soal (D16) | 6 / 18 | **LAYAK.** Analisis → tampilkan → ekspor (hitung ulang + audit). |
| 22 | UC13 Nilai per Kompetensi (D17) | 6 / 18 | **LAYAK.** Nilai per kompetensi → tampilkan → ekspor. |
| 23 | UC14 Kelola UKK (D18) | 6 / 40 | **LAYAK.** UKK baru → nilai → unggah berkas → terbitkan sertifikat (cek lulus). |
| 24 | UC15 Kegiatan, Ruangan & Kartu (D19) | 6 / 36 | **LAYAK.** Kegiatan → ruangan → tugaskan (cek kapasitas) → kartu PDF siswa. |
| 25 | UC16 Cadangkan & Pulihkan (D20) | 6 / 38 | **LAYAK.** Cadangan → pulihkan (cek ujian berjalan + cadangan pengaman) → hapus. |
| 26 | UC17 Pengaturan, Audit & Notifikasi (D21) | 6 / 42 | **LAYAK.** Pengaturan → jejak audit → notifikasi (tandai dibaca) → pengecualian. |

**Cacat yang belum diperbaiki: TIDAK ADA.** Karena 0 cacat visual,
tidak ada perbaikan skrip `.mdj` yang diperlukan di Fase 3b.
Skrip pemeriksa render: `docs/uml/verify-sequence-fase3b.py`.

### Catatan jujur — perbedaan ekspor vs Fase 3 (BUKAN cacat, BUKAN perubahan diagram)

Ekspor ulang di lingkungan ini TIDAK byte-identik dengan ekspor
Fase 3, padahal `.mdj` terbukti 0 perubahan di luar Sequence
(§3b.2) dan D1–D9 geometri/label 0 perbedaan. Penyebabnya adalah
**metrik font renderer StarUML**, bukan isi diagram:

* D1, D2, D5, D6: SVG Sequence **byte-identik** dengan Fase 3.
* D3, D4, D7, D8, D9: satu-satunya perbedaan adalah nama lifeline
  `Layanan Batas Waktu` yang kini ter-render **1 baris penuh**
  (sebelumnya terlipat 2 baris: `Layanan Batas` + `Waktu`) — terbaca
  penuh, tidak terpotong, tidak tumpang tindih. Ini LEBIH BAIK.
* 5 diagram lain (01–05): ekspor ulang di `/tmp` juga berbeda
  ±beberapa piksel pada lebar/tinggi node yang dihitung dari ukuran
  teks (mis. elips Use Case, tinggi kotak Activity, `7.22998046875`
  → `7`). **File 01–05 di `your_files/` TIDAK ditimpa** (sesuai
  aturan) — perbedaan ini hanya artefak render, diagramnya identik
  di `.mdj`.
* Watermark `UNREGISTERED` berulang di semua PNG/SVG (StarUML
  trial). File `.mdj` bersih — ekspor ulang dari StarUML berlisensi
  untuk naskah skripsi.

## 3b.4 File ekspor baru (06–26) + file basi yang dihapus

Ekspor: SVG native StarUML (`cli:image` via CDP) + PNG 2× dari SVG via
cairosvg 2.9.1, latar putih. File 01–05 TIDAK diubah.

| # | PNG (`your_files/uml-png/`) | SVG (`your_files/uml-svg/`) | Kanvas SVG |
|---|---|---|---|
| 06 | `06-sequence-d1-masuk.png` (1972×4590) | `06-sequence-d1-masuk.svg` | 986×2295 |
| 07 | `07-sequence-d2-ganti-kata-sandi.png` (2302×2670) | `07-sequence-d2-ganti-kata-sandi.svg` | 1151×1335 |
| 08 | `08-sequence-d3-mulai-ujian.png` (2302×3270) | `08-sequence-d3-mulai-ujian.svg` | 1151×1635 |
| 09 | `09-sequence-d4-kerjakan-ujian.png` (2302×3270) | `09-sequence-d4-kerjakan-ujian.svg` | 1151×1635 |
| 10 | `10-sequence-d5-selesai-koreksi-hasil.png` (2632×3750) | `10-sequence-d5-selesai-koreksi-hasil.svg` | 1316×1875 |
| 11 | `11-sequence-d6-pantau-ujian-langsung.png` (1642×2430) | `11-sequence-d6-pantau-ujian-langsung.svg` | 821×1215 |
| 12 | `12-sequence-d7-anti-curang.png` (1972×2910) | `12-sequence-d7-anti-curang.svg` | 986×1455 |
| 13 | `13-sequence-d8-jeda-jadwal.png` (1972×2430) | `13-sequence-d8-jeda-jadwal.svg` | 986×1215 |
| 14 | `14-sequence-d9-buka-blokir.png` (2302×2190) | `14-sequence-d9-buka-blokir.svg` | 1151×1095 |
| 15 | `15-sequence-d10-uc02-master-data.png` (1972×4830) | `15-sequence-d10-uc02-master-data.svg` | 986×2415 |
| 16 | `16-sequence-d11-uc03-bank-soal.png` (1972×4350) | `16-sequence-d11-uc03-bank-soal.svg` | 986×2175 |
| 17 | `17-sequence-d12-uc04-impor-soal.png` (1972×3390) | `17-sequence-d12-uc04-impor-soal.svg` | 986×1695 |
| 18 | `18-sequence-d13-uc05-paket-ujian.png` (1642×4590) | `18-sequence-d13-uc05-paket-ujian.svg` | 821×2295 |
| 19 | `19-sequence-d14-uc06-jadwal-ujian.png` (1642×3870) | `19-sequence-d14-uc06-jadwal-ujian.svg` | 821×1935 |
| 20 | `20-sequence-d15-uc11-rekap-berita-acara.png` (1642×4590) | `20-sequence-d15-uc11-rekap-berita-acara.svg` | 821×2295 |
| 21 | `21-sequence-d16-uc12-analisis-butir.png` (1972×2430) | `21-sequence-d16-uc12-analisis-butir.svg` | 986×1215 |
| 22 | `22-sequence-d17-uc13-nilai-kompetensi.png` (1972×2430) | `22-sequence-d17-uc13-nilai-kompetensi.svg` | 986×1215 |
| 23 | `23-sequence-d18-uc14-ukk.png` (1972×5070) | `23-sequence-d18-uc14-ukk.svg` | 986×2535 |
| 24 | `24-sequence-d19-uc15-kegiatan-ruangan-kartu.png` (1972×4590) | `24-sequence-d19-uc15-kegiatan-ruangan-kartu.svg` | 986×2295 |
| 25 | `25-sequence-d20-uc16-cadangkan-pulihkan.png` (1972×4830) | `25-sequence-d20-uc16-cadangkan-pulihkan.svg` | 986×2415 |
| 26 | `26-sequence-d21-uc17-pengaturan-audit-notifikasi.png` (1972×5310) | `26-sequence-d21-uc17-pengaturan-audit-notifikasi.svg` | 986×2655 |

**Dihapus (basi): TIDAK ADA.** Direktori `uml-png/` / `uml-svg/`
sudah bersih dari diagram Sequence kadaluwarsa sejak Fase 3
(`06-sequence-3a`, `07-sequence-3b` sudah dihapus saat itu); tidak
ada lagi file Sequence basi yang tersisa.

## 3b.5 Salinan `.mdj`

`your_files/cbt-smk-v2.mdj` (symlink ke
`goals/cbt-smk-v2-go-react/files/cbt-smk-v2.mdj`, symlink dipertahankan)
SHA-256 `118caee05be5f425c5c8f6f496f0133235ea55951e74f4a47fc460b7609a59dc`
= `docs/uml/cbt-smk-v2.mdj`. ✅

## 3b.6 Yang TIDAK dilakukan / batasan jujur

* Tidak ada spot-check kanvas live GUI per diagram (kanvas tidak
  melukis di Xvfb) — digantikan render ekspor StarUML + pemeriksaan
  geometri + inspeksi piksel PNG (D10 atas + D21 atas/bawah dilihat
  langsung sebagai gambar; 21 diagram dianalisis teks/posisi SVG).
* Klaim "byte-identik" Fase 3 untuk 5 diagram lain TIDAK tereproduksi
  di lingkungan ini (lihat §3b.3, catatan jujur) — bukti
  ketidakberubahan yang dipakai adalah diff `.mdj` (0 perubahan),
  bukan byte ekspor.
* Watermark trial tetap ada di ekspor.

