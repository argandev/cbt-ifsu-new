# PANDUAN IMPORT SOAL — untuk Guru

**Dokumen:** `docs/PANDUAN-IMPORT-SOAL.md`
**Tanggal:** 16 Oktober 2026
**Untuk siapa:** Guru / penyusun soal yang mengisi bank soal lewat file. Tidak perlu pengetahuan teknis.
**Acuan teknis:** `docs/ANALISIS-IMPORT-SOAL.md` (spesifikasi lengkap). Dokumen ini adalah versi guru dari spesifikasi tersebut, ditambah format Word yang disetujui koordinator.
**Status:** Panduan + spesifikasi isi template. Batas file 5 MB. Mode `all_or_nothing` + duplikat `simpan_ganda` diizinkan. Kunci boleh kosong HANYA untuk tipe rubrik (uraian / upload_file / praktik_singkat — menjadi warning), dan diabaikan untuk pernyataan_sikap.

> **Label jujur yang berlaku di mana-mana:** aplikasi ini mendukung **"5 tipe standar AKM + 11 varian/ekstensi aplikasi"** — total 16 tipe soal yang bisa diimpor. Jangan menulis "16 tipe standar Kemendikbud".

---

## 1. Gambaran singkat: import soal itu apa?

Import soal adalah cara memasukkan banyak soal sekaligus ke satu bank soal dengan cara mengisi file template (Excel atau Word), lalu mengunggahnya. Satu file = satu bank soal. Bank soal menentukan mata pelajaran dan tingkat kelas — jadi Anda **tidak perlu** menulis kode bank di dalam file.

Alurnya hanya 6 langkah:

```
1. Unduh template (Excel .xlsx atau Word .docx)
2. Isi template (satu baris = satu soal di Excel; satu blok = satu soal di Word)
3. Unggah file di halaman Import Soal
4. Baca hasil PRATINJAU (preview) — periksa kunci terpetakan ke opsi yang benar
5. Pilih mode + perlakuan duplikat
6. Klik Simpan → cek soal di bank soal (soal masuk berstatus DRAF, wajib ditelaah)
```

**Tiga hal yang paling sering membuat import gagal** (baca dulu sebelum mengisi):

1. **Salah menulis kunci.** Cara menulis kunci BERBEDA untuk tiap tipe soal (lihat Bagian 5). Ini sumber salah nomor satu.
2. **Salah menulis kode tipe.** Di kolom `tipe` (Excel) / label `TIPE:` (Word) harus ditulis **kode**, misalnya `pg` — bukan nama panjang "Pilihan Ganda".
3. **Langsung klik Simpan tanpa memeriksa pratinjau.** Pratinjau menampilkan kunci yang terpetakan ke teks opsinya. Periksa pemetaan itu, bukan hanya tulisan "tidak ada error".

**Yang TIDAK bisa diimpor:** file CSV (jalur CSV lama dihapus dari UI — gunakan template Excel/Word), file `.xls` (format Excel lama), gambar yang ditempel langsung di file (harus berupa tautan URL, lihat Bagian 3.4 dan 4.4), dan soal dari Google Docs langsung (ekspor dulu ke Word `.docx`, lalu sesuaikan ke format Bagian 4).

---

## 2. Langkah 1 — Unduh template

1. Buka detail **Bank Soal** tujuan → klik **Import Soal**.
2. Klik **Unduh Template**, pilih **Excel (.xlsx)** atau **Word (.docx)**.
3. Simpan file template tersebut. Isi file template itu — jangan membuat file baru dari nol, karena nama sheet / nama kolom / nama label di template sudah diatur agar bisa dibaca aplikasi.

| File | Nama file template | Isi |
|---|---|---|
| Excel | `template-import-soal-cbt-smk-v2.xlsx` | 4 sheet: `Petunjuk`, `Soal`, `Contoh`, `Referensi Tipe` |
| Word | `template-import-soal-cbt-smk-v2.docx` | Petunjuk + 2–3 contoh soal format kita |

**Aturan yang berlaku untuk kedua format:**

- Batas: **maksimal 500 soal per file** dan **maksimal 5 MB** ukuran file. Bila soal Anda lebih dari 500, pecah menjadi beberapa file.
- Jangan mengganti nama sheet (Excel), nama kolom (Excel), atau nama label (Word).
- Jangan menulis rumus / formula di sel mana pun. Tulis teks biasa.
- Soal hasil import berstatus **DRAF** — wajib ditelaah di editor sebelum dipakai di paket ujian.

---

## 3. Langkah 2a — Mengisi template Excel (.xlsx)

### 3.1 Sheet apa saja yang ada

| Sheet | Gunanya | Dibaca aplikasi? |
|---|---|---|
| `Petunjuk` | Panduan pengisian singkat di dalam file | Tidak — hanya untuk dibaca manusia |
| `Soal` | **Di sinilah Anda mengisi soal.** Satu baris = satu soal | Ya — satu-satunya sheet yang dibaca |
| `Contoh` | 16 baris contoh, satu per tipe soal. Boleh disalin ke sheet `Soal` lalu diganti isinya | Tidak |
| `Referensi Tipe` | Tabel 16 tipe: kode, kelompok, format kunci, kolom opsi yang dipakai | Tidak |

Urutan baris di sheet `Soal` = urutan soal masuk ke bank. Baris yang **semua** selnya kosong dilewati (bukan error). Baris yang sebagian terisi tetapi kolom `tipe`-nya kosong akan error.

### 3.2 Daftar kolom sheet `Soal`

Header ditulis persis seperti di bawah (huruf kecil, garis bawah). Besar-kecil huruf tidak masalah, tetapi **jangan mengganti nama kolom**. Kolom tambahan yang tidak dikenal akan diabaikan + muncul satu warning yang menyebut nama kolomnya.

| # | Kolom | Wajib? | Bila kosong | Artinya |
|---|---|---|---|---|
| 1 | `tipe` | **Wajib** | Error | Kode tipe, persis dari Bagian 5, mis. `pg`. Jangan tulis nama panjang. |
| 2 | `pertanyaan` | **Wajib** | Error | Teks soal. Untuk `teks_rumpang`, tulis rumpang sebagai `___` (tiga garis bawah). Rumus matematika ditulis sebagai teks LaTeX `$...$` (lihat Bagian 3.5). |
| 3 | `stimulus` | Opsional, **wajib untuk `studi_kasus`** | Kosong | Teks bacaan/kasus. Bila beberapa soal memakai stimulus yang sama, salin teksnya ke tiap baris. |
| 4–8 | `opsi_a` … `opsi_e` | Tergantung tipe (Bagian 5) | Kosong / diabaikan | Opsi / pernyataan / pasangan / butir, sesuai tipe. Maksimal 5 (A–E). |
| 9 | `kunci` | **Wajib**, kecuali pengecualian di Bagian 3.3 | Error / warning | Kunci / rubrik. Formatnya BERBEDA per tipe — lihat Bagian 5. |
| 10 | `bobot` | Opsional | `1` | Angka bulat lebih dari 0. Nilai lain (mis. `0`, `1,5`, `dua`) = error. |
| 11 | `level` | Opsional | `sedang` | Salah satu: `mudah`, `sedang`, `sulit`. Nilai lain = error. |
| 12 | `pembahasan` | Opsional | Kosong | Penjelasan jawaban. Ditampilkan ke siswa sesuai aturan tayang pembahasan. |
| 13 | `media_url` | Opsional | Kosong | **Satu** tautan gambar untuk soal ini, mis. `https://.../gambar.png`. Harus tautan `http://` atau `https://` yang valid. |
| 14 | `media_alt` | Opsional, warning bila `media_url` terisi tetapi kolom ini kosong | Kosong | Teks alternatif gambar (untuk pembaca layar / aksesibilitas). |
| 15 | `kisi_kisi` | Opsional | Soal masuk sebagai "belum terpetakan" | **Kode** kisi-kisi persis seperti tampil di halaman Kisi-kisi bank ini. Bila kode tidak ditemukan → **error** (tidak dibuang diam-diam). |
| 16 | `kompetensi` | Opsional | Soal masuk sebagai "belum terpetakan" | **Kode** kompetensi (CP/TP) persis seperti tampil di master Kompetensi. Bila kode tidak ditemukan → **error**. |

### 3.3 Kapan kolom `kunci` boleh kosong?

| Tipe | `kunci` kosong? |
|---|---|
| 12 tipe yang dinilai otomatis (`pg`, `pg_kompleks`, `menjodohkan`, `isian`, `benar_salah`, `pg_multi`, `urutkan`, `isi_ganda`, `teks_rumpang`, `peta_konsep`, `studi_kasus`, `pilihan_gambar`) | **ERROR** — baris tidak bisa disimpan. |
| `uraian`, `upload_file`, `praktik_singkat` | **Boleh kosong** — tetapi muncul **warning** "rubrik belum diisi". Rubrik bisa dilengkapi di editor sesudah import. Sebaiknya tetap isi rubriknya di file. |
| `pernyataan_sikap` | **Diabaikan total** — tipe ini memakai skala 1–4 dan **tidak dinilai**. Bila terisi, muncul info bahwa kunci diabaikan. |

### 3.4 Gambar di Excel

- **Gambar yang ditempel langsung di file Excel TIDAK ikut terimpor.** Satu-satunya cara memberi gambar pada soal adalah kolom `media_url` (tautan gambar) + `media_alt` (teks alternatif).
- Tautan gambar harus bisa dibuka. Server **tidak** memeriksa apakah tautan bisa dibuka — periksa sendiri dengan mengklik tautan di pratinjau.
- Untuk `pilihan_gambar`, tiap opsi (`opsi_a` …) berisi **tautan gambar**, bukan teks.

### 3.5 Rumus matematika di Excel

Tulis rumus sebagai **teks LaTeX** di kolom `pertanyaan` (dan `stimulus` bila perlu), diapit tanda dolar: `$...$`. Contoh: `Hitung nilai dari $2x + 3$ untuk $x = 4$.`

- Di halaman ujian, rumus dirender sebagai rumus matematika (KaTeX).
- Jangan menulis rumus Excel / formula (mis. `=SUM(A1:A2)`) di sel mana pun.
- Contoh benar: `Tentukan hasil dari $\frac{1}{2} + \frac{1}{4}$.`
- Contoh salah: menulis `=1/2+1/4` di sel (akan dibaca sebagai teks/formula, bukan rumus soal).

---

## 4. Langkah 2b — Mengisi template Word (.docx)

Format Word dirancang sendiri untuk aplikasi ini — **bukan** format aplikasi lain. Jangan memakai format `S:/A:/JAWAB:` atau kode `TS/KD/KJ/ABS` dari tempat lain; format-format itu **tidak dikenali** di sini.

### 4.1 Struktur satu soal

Setiap soal adalah satu blok yang dimulai dengan baris `SOAL <nomor>`, diikuti label-label berikut (tulis label persis seperti ini, termasuk titik dua). Label yang tidak dipakai boleh dihilangkan, kecuali yang wajib. Pisahkan antar-soal dengan **satu baris kosong**.

```
SOAL 1
TIPE: pg
PERTANYAAN: Ibu kota provinsi Jawa Barat adalah …
A. Bandung
B. Semarang
C. Medan
D. Surabaya
KUNCI: A
BOBOT: 1
LEVEL: sedang
PEMBAHASAN: Bandung adalah ibu kota provinsi Jawa Barat.
KISI-KISI: (kosongkan bila belum dipetakan)
KOMPETENSI: (kosongkan bila belum dipetakan)
MEDIA: (kosongkan bila tidak ada gambar; bila ada, isi tautan https://...)
```

### 4.2 Arti tiap label

| Label | Wajib? | Artinya |
|---|---|---|
| `SOAL <nomor>` | Ya, pembuka tiap blok | Nomor urut soal (1, 2, 3, …). Dipakai untuk menyebut soal di pesan error ("Soal 3: …"). |
| `TIPE:` | Ya | Kode tipe, persis dari Bagian 5 (mis. `pg`). Bukan nama panjang. |
| `STIMULUS:` | Wajib untuk `studi_kasus`; opsional untuk tipe lain | Teks bacaan/kasus. Bila beberapa soal memakai stimulus yang sama, salin teksnya ke tiap blok. |
| `PERTANYAAN:` | Ya | Teks soal. Rumpang ditulis `___`. Rumus ditulis sebagai teks LaTeX `$...$` (lihat Bagian 4.3). Bila pertanyaan panjang (lebih dari satu paragraf), tulis terus di baris-baris berikutnya sampai label berikutnya. |
| `A.`–`E.` | Tergantung tipe (Bagian 5) | Opsi / pernyataan / pasangan / butir. Tulis `A.` lalu isi, `B.` lalu isi, dst. Untuk `menjodohkan` / `peta_konsep`, satu opsi = satu pasangan `kiri::kanan`. Untuk `pilihan_gambar`, satu opsi = satu tautan gambar. |
| `KUNCI:` | Sama seperti Excel (Bagian 3.3) | Kunci / rubrik. Format per tipe di Bagian 5. Rubrik untuk `uraian` / `upload_file` / `praktik_singkat` boleh kosong (warning). Untuk `pernyataan_sikap`, diabaikan. |
| `BOBOT:` | Opsional (default 1) | Angka bulat lebih dari 0. |
| `LEVEL:` | Opsional (default `sedang`) | `mudah` / `sedang` / `sulit`. |
| `PEMBAHASAN:` | Opsional | Penjelasan jawaban. |
| `KISI-KISI:` | Opsional | Kode kisi-kisi persis seperti tampil di halaman Kisi-kisi. Kode tak ditemukan → error. Kosong = belum terpetakan. |
| `KOMPETENSI:` | Opsional | Kode kompetensi (CP/TP). Kode tak ditemukan → error. Kosong = belum terpetakan. |
| `MEDIA:` | Opsional | Satu tautan gambar `http(s)://...` untuk soal ini. |
| `ALT MEDIA:` | Opsional, warning bila `MEDIA:` terisi tetapi label ini tidak ada | Teks alternatif gambar (aksesibilitas). |

### 4.3 Rumus matematika di Word

Sama seperti Excel: tulis rumus sebagai teks LaTeX `$...$` di `PERTANYAAN:` / `STIMULUS:`. Contoh benar: `PERTANYAAN: Tentukan hasil dari $\frac{1}{2} + \frac{1}{4}$.` Jangan memakai equation editor Word sebagai pengganti teks LaTeX, karena isinya tidak terbaca sebagai teks soal oleh importer.

### 4.4 Gambar di Word

Gambar yang ditempel langsung di dokumen Word **tidak ikut terimpor** (termasuk gambar melayang / floating image). Gunakan label `MEDIA:` berisi tautan gambar + `ALT MEDIA:` berisi teks alternatif. Periksa tautan dengan mengkliknya di pratinjau.

### 4.5 Contoh blok Word: menjodohkan

```
SOAL 2
TIPE: menjodohkan
PERTANYAAN: Pasangkan kota dengan provinsinya.
A. Bandung::Jawa Barat
B. Semarang::Jawa Tengah
C. Surabaya::Jawa Timur
D. Medan::Sumatera Utara
KUNCI: Bandung::Jawa Barat;Semarang::Jawa Tengah;Surabaya::Jawa Timur;Medan::Sumatera Utara
BOBOT: 2
LEVEL: sedang
```

### 4.6 Contoh blok Word: uraian dengan rubrik

```
SOAL 3
TIPE: uraian
PERTANYAAN: Jelaskan perbedaan antara RAM dan ROM beserta contoh penggunaannya dalam kehidupan sehari-hari.
KUNCI: Rubrik: pengertian RAM (30), pengertian ROM (30), perbedaan (20), contoh (20).
BOBOT: 1
LEVEL: sulit
```

---

## 5. Langkah 2c — Format kunci per tipe soal (Excel + Word sama)

Baca baris tipe Anda dengan teliti. Pemisah kunci **tidak sama** antar-tipe — inilah sumber salah yang paling sering terjadi.

| # | Tipe (kode) | Kelompok | `opsi_a`…`opsi_e` berisi | `kunci` berisi — contoh | Aturan validasi |
|---|---|---|---|---|---|
| 1 | `pg` | Inti AKM | Teks opsi (min 2) | Satu label: `B` | Label harus cocok dengan label opsi yang terisi |
| 2 | `pg_kompleks` | Inti AKM | Teks opsi (min 2) | Label dipisah koma: `A,C` | Min 1 label; semua label harus cocok dengan opsi |
| 3 | `menjodohkan` | Inti AKM | Pasangan `kiri::kanan` (min 2) | Pasangan benar, antar-pasangan dipisah `;`: `Bandung::Jawa Barat;Semarang::Jawa Tengah` | Tiap opsi & tiap pasangan kunci harus berformat `kiri::kanan` dengan kedua sisi terisi |
| 4 | `isian` | Inti AKM | Diabaikan | Jawaban; alternatif dipisah `\|`: `Jakarta\|DKI Jakarta` | Min 1 alternatif |
| 5 | `uraian` | Inti AKM | Diabaikan | Rubrik (teks bebas); boleh kosong + warning | — |
| 6 | `benar_salah` | Varian | Pernyataan (min 2) | Nilai `B`/`S` dipisah koma, sebanyak pernyataan: `B,S,B`. Juga diterima: `benar`/`salah`, `true`/`false`, `1`/`0` | Jumlah nilai harus = jumlah pernyataan |
| 7 | `pg_multi` | Varian | Teks opsi (min 2) | Sama seperti `pg_kompleks`: `A,C` | Sama seperti `pg_kompleks` |
| 8 | `urutkan` | Varian | Butir yang diurutkan (min 2) | Urutan benar sebagai label dipisah koma, sebanyak butir: `C,A,B,D` | Jumlah label = jumlah butir; tidak boleh ada label duplikat |
| 9 | `isi_ganda` | Varian | Diabaikan | Jawaban per kolom dipisah `\|`: `10\|cm` | Min 1 jawaban. Nama kolom dibuat otomatis (`Kolom 1`, `Kolom 2`, …) |
| 10 | `teks_rumpang` | Varian | Diabaikan | Jawaban per rumpang dipisah `\|`: `Jakarta\|Rupiah` | Jumlah jawaban harus = jumlah `___` di pertanyaan |
| 11 | `peta_konsep` | Varian | Pasangan `kiri::kanan` (min 2) | Sama seperti `menjodohkan` | Sama seperti `menjodohkan` |
| 12 | `studi_kasus` | Varian | Teks opsi (min 2) | Satu label: `B` | Sama seperti `pg` + `stimulus` wajib terisi |
| 13 | `pernyataan_sikap` | Varian | Pernyataan sikap (min 2) | **Diabaikan** (boleh kosong) | Skala 1–4 otomatis. Tipe ini **tidak dinilai** |
| 14 | `upload_file` | Varian | Diabaikan | Rubrik (teks bebas); boleh kosong + warning | — |
| 15 | `praktik_singkat` | Varian | Diabaikan | Rubrik (teks bebas); boleh kosong + warning | — |
| 16 | `pilihan_gambar` | Varian | **Tautan gambar** per opsi (min 2) | Satu label: `A` | Tiap opsi yang terisi harus tautan `http(s)` yang valid |

### Contoh benar vs salah (paling sering terjadi)

| Kasus | ❌ Salah | ✅ Benar | Kenapa |
|---|---|---|---|
| Kode tipe | `tipe: Pilihan Ganda` | `tipe: pg` | Parser hanya menerima kode |
| Kunci PG kompleks | `kunci: A dan C` | `kunci: A,C` | Pemisah yang dikenal hanya koma |
| Kunci menjodohkan | `kunci: Bandung-Jawa Barat; Semarang-Jawa Tengah` | `kunci: Bandung::Jawa Barat;Semarang::Jawa Tengah` | Pasangan memakai `::`, antar-pasangan memakai `;` |
| Kunci isian | `kunci: Jakarta atau DKI Jakarta` | `kunci: Jakarta\|DKI Jakarta` | Alternatif dipisah `\|` (garis tegak) |
| Kunci benar-salah | `kunci: B,S` untuk 3 pernyataan | `kunci: B,S,B` | Jumlah nilai harus sama dengan jumlah pernyataan |
| Kunci teks rumpang | Pertanyaan punya 2 `___`, `kunci: Jakarta` | `kunci: Jakarta\|Rupiah` | Jumlah jawaban harus sama dengan jumlah `___` |
| Opsi pilihan gambar | `opsi_a: gambar kabel UTP` | `opsi_a: https://.../utp.png` | Opsi pilihan gambar harus tautan gambar |
| Kunci kosong | `kunci: (kosong)` untuk tipe `pg` | Isi kuncinya, mis. `B` | Kunci kosong untuk 12 tipe dinilai otomatis = error |
| Stimulus studi kasus | `stimulus: (kosong)` untuk `studi_kasus` | Isi stimulusnya | `stimulus` wajib untuk `studi_kasus` |

**Cara penilaian soal hasil import (tidak berubah):** PG / studi_kasus / pilihan_gambar: benar = bobot penuh, salah = 0. PG kompleks / PG multi: proporsional, tanpa pengurangan. Benar-salah: proporsional per pernyataan. Menjodohkan / peta konsep: proporsional per pasangan. Isian: cocok persis setelah normalisasi (spasi berlebih dirapikan, huruf besar/kecil tidak dibedakan). Uraian / upload_file / praktik_singkat: dinilai manual oleh guru. Urutkan: proporsional per posisi. Isi ganda / teks rumpang: proporsional per kolom / per rumpang. Pernyataan sikap: tidak dinilai.

---

## 6. Langkah 3 — Upload file

1. Di dialog Import Soal, pilih file template yang sudah diisi (Excel `.xlsx` atau Word `.docx`).
2. Pastikan bank soal yang dituju sudah benar (nama bank tampil di dialog). File tidak membawa info bank — bank berasal dari pilihan Anda di aplikasi.
3. Klik **Pratinjau** (bukan langsung Simpan). File yang diunggah akan dibaca dan divalidasi **tanpa menyimpan apa pun** ke bank soal.

**File ditolak sebelum dibaca?** Lihat katalog error level file di Bagian 9.1 (mis. bukan `.xlsx`/`.docx`, file lebih dari 5 MB, file rusak).

---

## 7. Langkah 4 — Membaca hasil pratinjau (preview)

Pratinjau adalah tahap **wajib**. Tidak ada soal yang masuk bank sebelum Anda melihat pratinjau dan menekan Simpan.

### 7.1 Ringkasan

Pratinjau menampilkan ringkasan:

- **Total baris/soal** yang dibaca dari file.
- **Jumlah valid** (siap disimpan), **jumlah error** (tidak bisa disimpan), **jumlah warning** (bisa disimpan, tetapi perlu diperhatikan), **jumlah duplikat**.
- **Ringkasan per tipe** (mis. `pg: 12`, `isian: 3`, …).
- **Total bobot** soal yang valid.

### 7.2 Tabel pratinjau soal

Tabel menampilkan **semua** soal yang terbaca dari file (maksimal 500), satu baris per soal:

| Kolom pratinjau | Yang harus Anda periksa |
|---|---|
| Nomor baris Excel / nomor soal Word (`SOAL N`) | Cocokkan dengan file Anda saat memperbaiki error. Baris 1 di Excel = header; data mulai baris 2. |
| Tipe (label + badge Inti AKM / Varian aplikasi) | Apakah tipenya sesuai maksud Anda? |
| Pertanyaan | Apakah teksnya lengkap dan tidak terpotong? |
| Opsi + **kunci yang terpetakan** | **Ini yang terpenting.** Kunci ditampilkan sebagai label + **teks opsi yang dirujuknya** (bukan hanya huruf). Contoh: `B — Semarang`. Pastikan hurufnya menunjuk ke opsi yang benar-benar Anda maksud. Kunci yang salah pemisah sering lolos sebagai "tidak ada error" di mata sekilas — pemetaan inilah yang membongkarnya. |
| Bobot, level | Apakah sesuai? |
| Kisi-kisi / kompetensi terpetakan | Apakah kodenya ketemu? |
| `media_url` | Ditampilkan sebagai **tautan yang bisa diklik**. Klik untuk memastikan gambar bisa dibuka. |
| Warning per baris | Baca dan putuskan: perbaiki di file, atau biarkan dan lengkapi di editor nanti. |

Baris yang error ditampilkan bersama pesan error-nya (lihat Bagian 9).

### 7.3 Daftar error dan warning

Di bawah/besamping tabel ada daftar error bernomor (baris / soal + pesan) dan daftar warning. Perbaiki error di **file Anda**, lalu unggah ulang — jangan mencoba "mengakali" dengan mengedit angka di pratinjau, karena pratinjau tidak bisa diedit.

---

## 8. Langkah 5 — Memilih mode + perlakuan duplikat

Dua pilihan ini muncul di pratinjau, sebelum tombol Simpan.

### 8.1 Mode: apa yang terjadi bila ada baris yang error

| Mode | Artinya | Kapan pakai |
|---|---|---|
| `all_or_nothing` — **Tolak semua bila ada error** (default) | Bila ada **satu saja** baris yang error, **tidak ada satu pun** soal yang disimpan. Anda memperbaiki file, lalu mengunggah ulang sampai bersih. | **Pakai ini dalam hampir semua kasus.** Paling aman: Anda tahu persis file Anda bersih sebelum ada yang masuk bank. |
| `skip_errors` — **Lewati baris yang error** | Baris yang valid disimpan; baris yang error dilewati dan dilaporkan. | Pakai bila file besar dan Anda rela sebagian soal masuk dulu sementara baris yang error diperbaiki terpisah. Hati-hati: mudah lupa memperbaiki baris yang terlewat. |

Catatan: bila mode `all_or_nothing` dan masih ada error, tombol **Simpan tidak aktif** sampai error diperbaiki.

### 8.2 Duplikat: apa yang terjadi bila ada soal yang sama

Dua soal disebut **duplikat** bila teks pertanyaannya sama setelah dirapikan (spasi berlebih dirapikan, huruf besar/kecil tidak dibedakan). Tidak ada pencocokan "mirip-mirip" — hanya teks yang sama persis setelah dirapikan.

| Pilihan | Artinya | Kapan pakai |
|---|---|---|
| `lewati` — **Lewati duplikat** (default) | Baris duplikat **tidak disimpan**: duplikat dalam file (kemunculan kedua dst.) dan duplikat dengan soal yang sudah ada di bank. Jumlah yang dilewati dilaporkan terpisah (`dilewati_duplikat`). | **Pakai ini bila Anda mengunggah ulang file yang sama** (mis. setelah memperbaiki error) atau tidak sengaja menulis soal yang sama dua kali. |
| `simpan_ganda` — **Simpan ganda** | Semua baris valid disimpan, **termasuk** yang duplikat. | Pakai hanya bila Anda **sengaja** ingin soal yang sama masuk dua kali (mis. soal yang sama dipakai dengan bobot/level berbeda, atau bank memang butuh soal kembar). |

Duplikat adalah **warning, bukan error** — duplikat tidak pernah menggagalkan file dalam mode `all_or_nothing`. Pilihan duplikat berlaku untuk kedua mode.

### 8.3 Contoh keputusan

- **"Saya mengunggah file 50 soal, 2 baris error, sisanya sudah benar."** → Pakai `all_or_nothing` + `lewati`. Perbaiki 2 baris, unggah ulang, semua tersimpan bersih.
- **"File saya 400 soal, 3 baris error di ujung, saya butuh 397 soal masuk sekarang."** → Pakai `skip_errors` + `lewati`. Simpan, lalu perbaiki 3 baris itu di file terpisah / input manual.
- **"Saya mengunggah ulang file yang sama setelah memperbaiki error."** → Pakai `lewati` untuk duplikat, agar soal yang sudah masuk tidak masuk dua kali.

---

## 9. Langkah 6 — Konfirmasi simpan + cek soal di bank soal

1. Setelah pilihan mode + duplikat sesuai, klik **Simpan**. Sistem mengunggah ulang file yang sama dan memvalidasi sekali lagi (agar hasil simpan selalu mencerminkan file yang benar-benar disimpan), lalu menulis soal sesuai pilihan Anda.
2. Hasil yang ditampilkan: **disimpan (N)**, **gagal (N + daftar error)**, **dilewati_duplikat (N)**, **dilewati karena error (N, bila `skip_errors`)**, + tautan **"Lihat bank soal"**.
3. Klik **Lihat bank soal** dan periksa soal yang baru masuk:
   - Soal berstatus **DRAF** — wajib ditelaah sebelum dipakai di paket ujian.
   - Buka beberapa soal, periksa kunci, opsi, stimulus, gambar, dan rubriknya.
   - Lengkapi yang masih kurang (rubrik yang masih kosong, teks alternatif gambar, kisi-kisi/kompetensi yang belum terpetakan).
4. Bila hasil simpan tidak sesuai harapan, jangan mengunggah ulang file yang sama berulang-ulang tanpa memeriksa duplikat — gunakan pilihan `lewati` atau hapus soal hasil import yang salah di bank soal, lalu ulangi dengan file yang sudah diperbaiki.

---

## 10. Katalog error + warning (bahasa guru)

Semua pesan memakai nomor baris Excel (baris 1 = header, data mulai baris 2) atau nomor soal Word (`SOAL N`). Pesan level file memakai nomor 0.

### 10.1 Error level file (seluruh file ditolak)

| Gejala | Pesan yang tampil | Cara memperbaiki |
|---|---|---|
| File bukan `.xlsx` / `.docx` | `Hanya file .xlsx yang diterima…` / pesan setara untuk Word | Unduh template resmi dan isi template itu. Jangan memakai `.xls`, `.csv`, `.txt`, atau `.pdf`. |
| File lebih dari 5 MB | `File melebihi batas 5 MB.` | Pecah file menjadi beberapa file; hapus gambar tertanam (gambar tertanam tidak ikut terimpor). |
| File rusak / tidak bisa dibaca | `File tidak bisa dibaca: …` | Buka file di Excel/Word, simpan ulang, lalu unggah lagi. |
| Sheet `Soal` tidak ditemukan (Excel) | `Sheet "Soal" tidak ditemukan. Gunakan template resmi — jangan mengganti nama sheet.` | Jangan mengganti nama sheet. Kembali ke template resmi. |
| File hanya berisi header / tanpa soal | `File harus punya header + minimal 1 baris soal.` | Isi minimal 1 soal. |
| Kolom wajib hilang di header (Excel) | `Kolom wajib hilang di header: <nama_kolom>.` | Kembalikan kolom `tipe`, `pertanyaan`, `kunci` ke header. |
| Lebih dari 500 soal | `File melebihi batas 500 baris soal.` | Pecah menjadi beberapa file. |
| Bank terkunci | `Bank soal terkunci — import ditolak.` | Minta admin membuka kunci bank, atau import ke bank lain. |

### 10.2 Error per baris/soal (baris tersebut tidak disimpan)

"Baris" = nomor baris Excel / nomor soal Word. Pesan yang tampil persis seperti kolom "Pesan yang tampil".

| Kode | Artinya (bahasa guru) | Pesan yang tampil | Cara memperbaiki |
|---|---|---|---|
| E1 | Kolom `tipe` / label `TIPE:` kosong | `tipe wajib diisi` | Isi kode tipenya (mis. `pg`). |
| E2 | Kode tipe tidak dikenal | `tipe "<nilai>" tidak dikenal` | Tulis kode persis dari Bagian 5. Jangan tulis nama panjang ("Pilihan Ganda"). |
| E3 | Pertanyaan kosong | `pertanyaan wajib diisi` | Isi teks pertanyaannya. |
| E4 | Kunci kosong padahal tipe ini dinilai otomatis | `kunci wajib diisi` | Isi kunci sesuai format tipe (Bagian 5). Pengecualian: lihat Bagian 3.3. |
| E5 | Bobot bukan bilangan bulat > 0 | `bobot "<nilai>" harus bilangan bulat > 0` | Isi bobot dengan angka bulat, mis. `1`, `2`. Bukan `0`, `1,5`, atau huruf. |
| E6 | Level bukan mudah/sedang/sulit | `level "<nilai>" harus mudah / sedang / sulit` | Isi salah satu dari tiga kata itu (atau kosongkan = `sedang`). |
| E7 | PG / studi_kasus / pilihan_gambar: opsi terisi kurang dari 2 | `opsi_a..opsi_e: minimal 2 opsi terisi` | Isi minimal 2 opsi. |
| E8 | PG / studi_kasus / pilihan_gambar: kunci tidak cocok dengan opsi yang ada | `kunci "<nilai>" tidak cocok dengan label opsi yang terisi` | Kunci harus huruf opsi yang terisi (mis. bila hanya A–C terisi, kunci `D` salah). |
| E9 | PG kompleks / PG multi: opsi kurang dari 2 | `opsi_a..opsi_e: minimal 2 opsi terisi` | Isi minimal 2 opsi. |
| E10 | PG kompleks / PG multi: ada huruf kunci yang tidak cocok opsi | `kunci "<label>" tidak cocok dengan label opsi yang terisi` | Semua huruf kunci harus merujuk opsi yang terisi. |
| E11 | PG kompleks / PG multi: kunci kosong sesudah dipisah | `kunci wajib minimal satu label opsi` | Isi minimal satu huruf kunci, mis. `A,C`. |
| E12 | Benar-salah: pernyataan kurang dari 2 | `opsi_a..opsi_e: isi pernyataan (minimal 2)` | Isi minimal 2 pernyataan. |
| E13 | Benar-salah: jumlah nilai kunci tidak sama dengan jumlah pernyataan | `kunci harus <N> nilai B/S, dipisah koma` | Isi nilai B/S sebanyak pernyataan, mis. 3 pernyataan → `B,S,B`. |
| E14 | Benar-salah: nilai kunci bukan B/S (atau bentuk yang diterima) | `kunci "<nilai>" harus B atau S` | Pakai `B`/`S` (atau `benar`/`salah`, `true`/`false`, `1`/`0`). |
| E15 | Menjodohkan / peta konsep: pasangan tampil kurang dari 2 | `opsi_a..opsi_e: isi pasangan format kiri::kanan (minimal 2)` | Isi minimal 2 pasangan `kiri::kanan`. |
| E16 | Menjodohkan / peta konsep: opsi tidak berformat pasangan | `opsi <LABEL> harus format kiri::kanan` | Tulis tiap opsi sebagai `kiri::kanan`, kedua sisi terisi. |
| E17 | Menjodohkan / peta konsep: pasangan kunci tidak berformat pasangan | `kunci harus pasangan format kiri::kanan dipisah ;` | Tulis kunci sebagai `kiri::kanan;kiri::kanan`. |
| E18 | Isian: tidak ada alternatif jawaban | `kunci isian wajib diisi (alternatif dipisah \|)` | Isi minimal satu jawaban; alternatif dipisah `\|`. |
| E19 | Urutkan: butir kurang dari 2 | `opsi_a..opsi_e: isi butir yang diurutkan (minimal 2)` | Isi minimal 2 butir. |
| E20 | Urutkan: jumlah label kunci tidak sama dengan jumlah butir | `kunci harus <N> label opsi (urutan benar), dipisah koma` | Tulis semua label dalam urutan benar, mis. 4 butir → `B,C,A,D`. |
| E21 | Urutkan: label kunci tidak cocok opsi | `kunci "<label>" tidak cocok dengan label opsi yang terisi` | Semua label kunci harus merujuk opsi yang terisi. |
| E22 | Urutkan: label kunci duplikat | `kunci berisi label "<label>" duplikat` | Tiap label dipakai tepat satu kali (urutan, bukan pilihan). |
| E23 | Isi ganda: tidak ada jawaban per kolom | `kunci wajib jawaban per kolom, dipisah \|` | Isi jawaban per kolom dipisah `\|`, mis. `50\|30`. |
| E24 | Teks rumpang: tidak ada jawaban per rumpang | `kunci wajib jawaban per rumpang, dipisah \|` | Isi jawaban per rumpang dipisah `\|`. |
| E25 | Teks rumpang: jumlah jawaban tidak sama dengan jumlah `___` | `jumlah rumpang ___ harus sama dengan jumlah jawaban kunci (pisah dengan \|)` | Samakan jumlah jawaban dengan jumlah `___` di pertanyaan. |
| E26 | Studi kasus: stimulus kosong | `stimulus wajib diisi untuk tipe studi_kasus` | Isi stimulus/bacaannya. |
| E27 | Pernyataan sikap: pernyataan kurang dari 2 | `opsi_a..opsi_e: isi pernyataan sikap (minimal 2)` | Isi minimal 2 pernyataan sikap. |
| E28 | `media_url` / `MEDIA:` bukan tautan `http(s)` yang valid | `media_url "<nilai>" bukan URL http(s) yang valid` | Isi tautan lengkap diawali `https://` (atau `http://`). |
| E29 | Pilihan gambar: opsi bukan tautan gambar | `opsi <LABEL> harus URL gambar http(s)` | Isi tiap opsi dengan tautan gambar `https://...`. |
| E30 | Kode kisi-kisi tidak ditemukan | `kisi_kisi "<nilai>" tidak ditemukan di bank ini` | Salin kode persis dari halaman Kisi-kisi, atau kosongkan. |
| E31 | Kode kompetensi tidak ditemukan | `kompetensi "<nilai>" tidak ditemukan` | Salin kode persis dari master Kompetensi, atau kosongkan. |
| E32 | Gagal menyimpan ke database | `Gagal menyimpan ke database.` | Coba lagi. Bila berulang, hubungi admin (sertakan nomor baris + pesan). |

**Yang SENGAJA tidak divalidasi (bukan bug):** panjang pertanyaan/opsi tidak dibatasi (soal cerita panjang boleh); keterjangkauan tautan gambar tidak diperiksa server (periksa sendiri dengan mengkliknya); pasangan kunci menjodohkan tidak diwajibkan sama dengan pasangan tampil (pengecoh di sisi kanan didukung).

### 10.3 Warning (baris tetap disimpan, tetapi perhatikan)

| Kode | Artinya | Pesan yang tampil | Yang harus dilakukan |
|---|---|---|---|
| W1 | Rubrik belum diisi (`uraian` / `upload_file` / `praktik_singkat`) | `kunci kosong — rubrik belum diisi. Rubrik bisa dilengkapi di editor sesudah import.` | Lengkapi rubrik di editor sebelum soal dipakai. |
| W2 | Gambar tanpa teks alternatif | `media_url diisi tetapi media_alt kosong — gambar tanpa teks alternatif.` | Isi `media_alt` / `ALT MEDIA:` di editor. |
| W3 | Kolom tidak dikenal di sheet `Soal` (Excel) | `Kolom "<nama>" tidak dikenal — diabaikan.` | Kolom itu diabaikan — pindahkan isinya ke kolom yang benar bila penting. |
| W4 | Duplikat dalam file | `pertanyaan duplikat dengan baris <N>.` | Putuskan di pilihan duplikat (Bagian 8.2): lewati atau simpan ganda. |
| W5 | Duplikat dengan soal yang sudah ada di bank | `pertanyaan sudah ada di bank ini.` | Sama seperti W4. |
| W6 | Kunci terisi untuk pernyataan sikap (info) | `kunci diabaikan untuk pernyataan_sikap (skala 1–4, tidak dinilai).` | Tidak perlu tindakan — kunci memang diabaikan untuk tipe ini. |
| W7 | Tautan gambar tidak diperiksa (info) | `URL gambar tidak diperiksa keterjangkauannya — klik URL di preview untuk memastikan gambar bisa dibuka.` | Klik tautan di pratinjau. |

---

## 11. FAQ

**T: File saya ditolak. Apa yang pertama kali harus saya cek?**
J: Cek tiga hal: (1) format file — harus `.xlsx` (Excel) atau `.docx` (Word) dari template resmi; (2) ukuran — maksimal 5 MB; (3) di Excel, nama sheet `Soal` tidak diubah dan kolom wajib `tipe`, `pertanyaan`, `kunci` ada di header. Pesan error level file (Bagian 10.1) menyebut penyebabnya.

**T: Kunci saya kosong karena soalnya uraian. Apakah boleh?**
J: Boleh — untuk `uraian`, `upload_file`, dan `praktik_singkat` saja. Kolom `kunci` diisi **rubrik**, dan bila kosong akan muncul warning W1 (rubrik belum diisi). Lengkapi rubrik di editor sebelum soal dipakai. Untuk 12 tipe yang dinilai otomatis, kunci kosong = error (E4). Untuk `pernyataan_sikap`, kunci diabaikan karena tipe ini tidak dinilai.

**T: Saya mengunggah file yang sama dua kali. Apakah soal menjadi ganda?**
J: Tidak, bila Anda memilih duplikat = `lewati` (default). Baris yang pertanyaannya sudah ada di bank akan dilewati dan dilaporkan sebagai `dilewati_duplikat`. Bila Anda memilih `simpan_ganda`, soal akan masuk dua kali — pakai hanya bila memang disengaja.

**T: Bagaimana cara memberi gambar pada soal?**
J: Tulis **tautan** gambar di kolom `media_url` (Excel) / label `MEDIA:` (Word), plus teks alternatif di `media_alt` / `ALT MEDIA:`. Gambar yang ditempel langsung di file Excel/Word **tidak** ikut terimpor. Untuk `pilihan_gambar`, tiap opsi berisi tautan gambar. Selalu klik tautan di pratinjau untuk memastikan gambar bisa dibuka.

**T: Bagaimana cara menulis rumus matematika?**
J: Tulis sebagai teks LaTeX di kolom/label pertanyaan (dan stimulus bila perlu), diapit tanda dolar: `$...$`. Contoh: `Hitung nilai dari $2x + 3$ untuk $x = 4$.` Di halaman ujian, rumus dirender sebagai rumus matematika (KaTeX). Jangan menulis rumus Excel / formula di sel, dan jangan mengandalkan equation editor Word — isinya tidak terbaca sebagai teks soal oleh importer.

**T: Soal saya lebih dari 500. Bagaimana?**
J: Pecah menjadi beberapa file (masing-masing ≤ 500 soal dan ≤ 5 MB), lalu import satu per satu. Gunakan pilihan duplikat `lewati` bila ada soal yang sama di antar-file.

**T: Saya salah ketik kode kisi-kisi / kompetensi. Apa yang terjadi?**
J: Baris tersebut error (E30 / E31) dan tidak disimpan diam-diam tanpa pemetaan. Salin kode persis dari halaman Kisi-kisi / master Kompetensi, atau kosongkan kolomnya (soal masuk sebagai "belum terpetakan", bisa dipetakan di editor).

**T: Apakah saya bisa mengimpor file CSV lama?**
J: Tidak — jalur CSV dihapus dari UI. Pindahkan isi CSV ke template Excel (kolomnya sama, ditambah `media_url`, `media_alt`, `kisi_kisi`, `kompetensi`), lalu import sebagai `.xlsx`.

**T: Saya punya soal di Google Docs. Bagaimana?**
J: Ekspor dokumen Google Docs ke Word (`.docx`), lalu susun ulang isinya mengikuti format Bagian 4 (label `SOAL`, `TIPE:`, `PERTANYAAN:`, dst.). Tidak ada import langsung dari Google Docs.

**T: Soal sudah tersimpan, tetapi kuncinya ternyata salah. Bagaimana?**
J: Perbaiki di editor bank soal (soal hasil import berstatus DRAF, bisa diubah). Jangan mengunggah ulang file yang sama tanpa pilihan duplikat `lewati` — nanti soal masuk dua kali.

**T: Bank soal saya terkunci, import ditolak. Bagaimana?**
J: Bank yang terkunci tidak bisa ditambah isinya. Minta admin membuka kunci bank, atau import ke bank lain yang terbuka.

**T: Mengapa tombol Simpan tidak aktif?**
J: Karena mode `all_or_nothing` masih mendeteksi error. Perbaiki error di file, unggah ulang, dan pastikan pratinjau bersih — atau ganti mode ke `skip_errors` bila Anda memang ingin menyimpan baris yang valid saja.

**T: Apa bedanya warning dan error?**
J: Error = baris tidak bisa disimpan (dan dalam mode `all_or_nothing` menggagalkan seluruh file). Warning = baris tetap disimpan, tetapi ada hal yang perlu Anda ketahui / lengkapi (rubrik kosong, gambar tanpa teks alternatif, duplikat). Info = catatan saja.

---

## 12. Spesifikasi isi template (untuk Tim Excel / Tim Word)

Bagian ini merumuskan **teks final** isi template. Tim pembuat template (Excel/Word) memakai bagian ini sebagai naskah yang harus dimuat persis — bukan membuat teks sendiri. Tim ini (Panduan + Template) **tidak mengedit generator template** (file backend); yang dirumuskan di sini hanyalah isi/tekstualnya.

### 12.1 Sheet `Petunjuk` Excel — teks final

Tulis teks berikut di sheet `Petunjuk` (ditulis dengan kalimat pendek, tanpa istilah teknis yang tidak dijelaskan). Satu sel per baris panduan, atau satu blok teks per paragraf — pilih tata letak yang mudah dibaca guru; isi kalimatnya harus sama dengan teks di bawah.

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
    biasa. Rumus matematika dalam soal ditulis sebagai teks LaTeX
    diapit tanda dolar, mis. "$2x + 3$".
14. Batas: maks 500 baris soal per file, maks 5 MB.
    Bila soal Anda lebih dari 500, pecah menjadi beberapa file.
15. Alur: unggah file → PERIKSA PRATINJAU (pastikan kunci terpetakan
    ke opsi yang benar — jangan hanya melihat "tidak ada error") →
    pilih mode (tolak semua bila ada error / lewati baris yang error) →
    pilih perlakuan duplikat (lewati / simpan ganda) → Simpan.
    Soal hasil import berstatus DRAF — wajib ditelaah sebelum dipakai.
16. Bila ada error, perbaiki di file Excel Anda lalu unggah ulang.
    Daftar error menyebut NOMOR BARIS Excel (baris 1 = header).
```

### 12.2 Sheet `Contoh` Excel — 16 baris contoh (isi final)

Sheet `Contoh` memakai header yang **sama persis** dengan sheet `Soal` (16 kolom Bagian 3.2). Setiap baris di bawah adalah satu baris contoh. Kolom yang tidak disebut memakai default (`bobot` = 1, `level` = `sedang`, sisanya kosong). Contoh memakai materi umum yang netral (bukan soal ujian sungguhan) agar bisa disalin apa adanya sebagai kerangka. **Template harus memuat contoh persis ini** (bukan contoh lain yang dibuat pengembang template) agar dokumentasi dan file yang diunduh guru selalu sama.

| # | `tipe` | `pertanyaan` | `stimulus` | `opsi_a` … `opsi_e` | `kunci` | Catatan untuk guru |
|---|---|---|---|---|---|---|
| 1 | `pg` | Ibu kota provinsi Jawa Barat adalah … | (kosong) | A: Bandung; B: Semarang; C: Medan; D: Surabaya | `A` | — |
| 2 | `pg_kompleks` | Manakah yang termasuk perangkat keras komputer? (Pilih semua yang benar.) | (kosong) | A: Monitor; B: Sistem operasi; C: Keyboard; D: Peramban | `A,C` | Kunci dipisah koma |
| 3 | `menjodohkan` | Pasangkan kota dengan provinsinya. | (kosong) | A: `Bandung::Jawa Barat`; B: `Semarang::Jawa Tengah`; C: `Surabaya::Jawa Timur`; D: `Medan::Sumatera Utara` | `Bandung::Jawa Barat;Semarang::Jawa Tengah;Surabaya::Jawa Timur;Medan::Sumatera Utara` | Pasangan `::`, antar-pasangan `;` |
| 4 | `isian` | Ibu kota negara Indonesia adalah …. | (kosong) | (diabaikan) | `Jakarta\|DKI Jakarta` | Alternatif dipisah `\|` |
| 5 | `uraian` | Jelaskan perbedaan antara RAM dan ROM beserta contoh penggunaannya dalam kehidupan sehari-hari. | (kosong) | (diabaikan) | `Rubrik: pengertian RAM (30), pengertian ROM (30), perbedaan (20), contoh (20).` | Kunci = rubrik |
| 6 | `benar_salah` | Tentukan benar atau salah untuk setiap pernyataan berikut. | (kosong) | A: `Air mendidih pada suhu 100°C pada tekanan 1 atm.`; B: `Bulan menghasilkan cahayanya sendiri.`; C: `Fotosintesis menghasilkan oksigen.` | `B,S,B` | Jumlah nilai = jumlah pernyataan |
| 7 | `pg_multi` | Manakah yang termasuk bilangan prima? (Pilih semua yang benar.) | (kosong) | A: 2; B: 4; C: 5; D: 9 | `A,C` | Sama seperti pg_kompleks |
| 8 | `urutkan` | Urutkan langkah mencuci tangan yang benar. | (kosong) | A: `Bilas tangan dengan air bersih.`; B: `Basahi tangan dengan air mengalir.`; C: `Gosok kedua telapak tangan dengan sabun.`; D: `Keringkan tangan.` | `B,C,A,D` | Label dalam urutan benar, dipisah koma |
| 9 | `isi_ganda` | Sebuah persegi panjang memiliki panjang 10 dan lebar 5. Isi kolom berikut. | (kosong) | (diabaikan) | `50\|30` | Jawaban per kolom dipisah `\|`; kolom jadi "Kolom 1" (luas), "Kolom 2" (keliling) — nama kolom belum bisa diatur |
| 10 | `teks_rumpang` | Ibu kota Indonesia adalah ___ dan mata uangnya adalah ___. | (kosong) | (diabaikan) | `Jakarta\|Rupiah` | Jumlah jawaban = jumlah `___` (2) |
| 11 | `peta_konsep` | Pasangkan konsep dengan pasangannya. | (kosong) | A: `Gula::Manis`; B: `Jeruk::Asam`; C: `Garam::Asin` | `Gula::Manis;Jeruk::Asam;Garam::Asin` | Sama seperti menjodohkan |
| 12 | `studi_kasus` | Berdasarkan kasus di atas, tindakan pertama yang paling tepat adalah …. | `Sebuah bengkel menerima komputer yang tidak mau menyala sama sekali: tidak ada lampu indikator, tidak ada suara kipas.` | A: Langsung mengganti motherboard; B: Memeriksa kabel daya dan stopkontak; C: Menginstal ulang sistem operasi; D: Mengganti RAM | `B` | `stimulus` wajib |
| 13 | `pernyataan_sikap` | Beri tanggapanmu untuk setiap pernyataan berikut. | (kosong) | A: `Saya senang belajar dengan diskusi kelompok.`; B: `Saya lebih paham bila belajar sambil praktik langsung.` | (kosong — diabaikan) | Pratinjau menampilkan "tidak dinilai" |
| 14 | `upload_file` | Buatlah laporan praktikum sesuai format yang diberikan, lalu unggah berkasnya. | (kosong) | (diabaikan) | `Rubrik: kelengkapan isi (40), format laporan (30), ketepatan data (30).` | Kunci = rubrik; boleh kosong + warning |
| 15 | `praktik_singkat` | Praktikkan cara membuat kabel UTP straight dan jelaskan hasilnya. | (kosong) | (diabaikan) | `Rubrik: urutan warna benar (50), konektor terpasang rapi (30), penjelasan hasil (20).` | Kunci = rubrik; boleh kosong + warning |
| 16 | `pilihan_gambar` | Manakah gambar yang menunjukkan kabel UTP? | (kosong) | A: `https://contoh.sekolah.id/gambar/utp.png`; B: `https://contoh.sekolah.id/gambar/hdmi.png`; C: `https://contoh.sekolah.id/gambar/vga.png`; D: `https://contoh.sekolah.id/gambar/power.png` | `A` | Opsi = tautan gambar; ganti dengan tautan gambar sekolah masing-masing |

Catatan: contoh memakai domain `contoh.sekolah.id` sebagai placeholder yang jelas-jelas bukan tautan sungguhan — guru wajib menggantinya dengan tautan gambar sekolah masing-masing.

### 12.3 Sheet `Referensi Tipe` Excel — isi final

Sheet ini berisi tabel referensi 16 tipe. Kolom: `kode` | `label` | `kelompok` | `format kunci` | `pemisah kunci` | `kolom opsi yang dipakai`. Label memakai nama yang tampil di UI bank soal (bukan klaim standar). Isi baris mengikuti tabel Bagian 5 dokumen ini:

| kode | label | kelompok | format kunci | pemisah kunci | kolom opsi yang dipakai |
|---|---|---|---|---|---|
| `pg` | Pilihan Ganda | Inti AKM | Satu label | — | `opsi_a`–`opsi_e` (teks, min 2) |
| `pg_kompleks` | Pilihan Ganda Kompleks | Inti AKM | Label-label benar | `,` (koma) | `opsi_a`–`opsi_e` (teks, min 2) |
| `menjodohkan` | Menjodohkan | Inti AKM | Pasangan benar | `::` (pasangan), `;` (antar-pasangan) | `opsi_a`–`opsi_e` (pasangan, min 2) |
| `isian` | Isian Singkat | Inti AKM | Jawaban / alternatif | `\|` (garis tegak) | Tidak dipakai |
| `uraian` | Uraian | Inti AKM | Rubrik (boleh kosong + warning) | — | Tidak dipakai |
| `benar_salah` | Benar–Salah | Varian aplikasi | Nilai B/S per pernyataan | `,` (koma) | `opsi_a`–`opsi_e` (pernyataan, min 2) |
| `pg_multi` | Pilihan Ganda Multi-Jawaban | Varian aplikasi | Label-label benar | `,` (koma) | `opsi_a`–`opsi_e` (teks, min 2) |
| `urutkan` | Mengurutkan | Varian aplikasi | Label dalam urutan benar | `,` (koma) | `opsi_a`–`opsi_e` (butir, min 2) |
| `isi_ganda` | Isian Ganda | Varian aplikasi | Jawaban per kolom | `\|` (garis tegak) | Tidak dipakai |
| `teks_rumpang` | Teks Rumpang | Varian aplikasi | Jawaban per rumpang | `\|` (garis tegak) | Tidak dipakai |
| `peta_konsep` | Peta Konsep | Varian aplikasi | Pasangan benar | `::` (pasangan), `;` (antar-pasangan) | `opsi_a`–`opsi_e` (pasangan, min 2) |
| `studi_kasus` | Studi Kasus | Varian aplikasi | Satu label | — | `opsi_a`–`opsi_e` (teks, min 2); `stimulus` wajib |
| `pernyataan_sikap` | Pernyataan Sikap | Varian aplikasi | Diabaikan (skala 1–4, tidak dinilai) | — | `opsi_a`–`opsi_e` (pernyataan, min 2) |
| `upload_file` | Unggah Berkas | Varian aplikasi | Rubrik (boleh kosong + warning) | — | Tidak dipakai |
| `praktik_singkat` | Praktik Singkat | Varian aplikasi | Rubrik (boleh kosong + warning) | — | Tidak dipakai |
| `pilihan_gambar` | Pilihan Gambar | Varian aplikasi | Satu label | — | `opsi_a`–`opsi_e` (tautan gambar, min 2) |

Di bawah tabel, tambahkan catatan satu baris: `Label "Inti AKM" / "Varian aplikasi" mengikuti label jujur: 5 tipe standar AKM + 11 varian/ekstensi aplikasi.`

### 12.4 Dokumen `template.docx` — isi final

Dokumen template Word terdiri dari: (a) blok petunjuk singkat, (b) 3 contoh soal format kita. Urutan label dalam tiap blok contoh mengikuti Bagian 4.1. **Jangan** memakai format lain (bukan `S:/A:/JAWAB:`, bukan `TS/KD/KJ/ABS`).

**Blok petunjuk** (di awal dokumen, sebelum contoh):

```
CARA MENGISI TEMPLATE IMPORT SOAL (WORD) — CBT SMK v2

1. Satu blok SOAL = satu soal. Blok dimulai dengan baris "SOAL <nomor>".
   Pisahkan antar-soal dengan satu baris kosong.
2. Tulis label persis seperti contoh (termasuk titik dua):
   SOAL, TIPE, STIMULUS, PERTANYAAN, A. s.d. E., KUNCI, BOBOT, LEVEL,
   PEMBAHASAN, KISI-KISI, KOMPETENSI, MEDIA, ALT MEDIA.
   Label yang tidak dipakai boleh dihilangkan, kecuali yang wajib.
3. TIPE wajib diisi dengan KODE, bukan nama panjang.
   Contoh benar: pg. Contoh salah: Pilihan Ganda.
4. Cara menulis KUNCI BERBEDA per tipe:
   - pg / studi_kasus / pilihan_gambar: SATU huruf, mis. B
   - pg_kompleks / pg_multi: huruf dipisah KOMA, mis. A,C
   - benar_salah: B/S dipisah KOMA sebanyak pernyataan, mis. B,S,B
   - menjodohkan / peta_konsep: pasangan kiri::kanan, antar-pasangan
     dipisah TITIK KOMA, mis. Bandung::Jawa Barat;Semarang::Jawa Tengah
   - isian: jawaban, alternatif dipisah GARIS TEGAK |, mis. Jakarta|DKI Jakarta
   - isi_ganda / teks_rumpang: jawaban dipisah GARIS TEGAK |, mis. 50|30
   - uraian / upload_file / praktik_singkat: tulis RUBRIK di KUNCI.
     Boleh kosong (nanti dilengkapi di editor), tetapi muncul peringatan.
   - pernyataan_sikap: KUNCI DIABAIKAN (skala 1-4, tidak dinilai).
5. Opsi ditulis A. s.d. E. (maksimal 5). Untuk menjodohkan / peta_konsep,
   satu opsi = satu pasangan kiri::kanan. Untuk pilihan_gambar, satu opsi
   = satu tautan gambar (https://...). Untuk isian / isi_ganda /
   teks_rumpang / uraian / upload_file / praktik_singkat, opsi TIDAK dipakai.
6. STIMULUS wajib untuk studi_kasus. Bila beberapa soal memakai stimulus
   yang sama, salin teksnya ke tiap blok soal tersebut.
7. BOBOT: angka bulat lebih dari 0 (kosong = 1). LEVEL: mudah / sedang /
   sulit (kosong = sedang). PEMBAHASAN: opsional.
8. KISI-KISI dan KOMPETENSI: isi dengan KODE persis seperti tampil di
   halaman Kisi-kisi / Kompetensi. Kode tidak ditemukan = ERROR.
   Kosongkan bila belum dipetakan — soal tetap masuk sebagai draft.
9. MEDIA: SATU tautan gambar (https://...) untuk soal ini. Gambar yang
   ditempel langsung di dokumen Word TIDAK ikut terimpor — harus berupa
   tautan. ALT MEDIA: teks alternatif gambar (bila MEDIA diisi tetapi
   ALT MEDIA kosong, muncul peringatan).
10. Teks rumpang ditulis dengan tiga garis bawah: ___
    Contoh: "Ibu kota Indonesia adalah ___." Jawabannya di KUNCI.
11. Rumus matematika ditulis sebagai teks LaTeX diapit tanda dolar,
    mis. "$2x + 3$". Jangan menulis rumus Word / equation editor sebagai
    pengganti teks LaTeX.
12. Jangan menulis format lain (bukan S:/A:/JAWAB:, bukan TS/KD/KJ/ABS).
    Format-format itu tidak dikenali di sini.
13. Batas: maks 500 soal per file, maks 5 MB. Alur: unggah file →
    PERIKSA PRATINJAU (pastikan kunci terpetakan ke opsi yang benar) →
    pilih mode (tolak semua bila ada error / lewati baris yang error) →
    pilih perlakuan duplikat (lewati / simpan ganda) → Simpan.
    Soal hasil import berstatus DRAF — wajib ditelaah sebelum dipakai.
```

**Contoh 1 — Pilihan Ganda (pg):**

```
SOAL 1
TIPE: pg
PERTANYAAN: Ibu kota provinsi Jawa Barat adalah …
A. Bandung
B. Semarang
C. Medan
D. Surabaya
KUNCI: A
BOBOT: 1
LEVEL: sedang
PEMBAHASAN: Bandung adalah ibu kota provinsi Jawa Barat.
KISI-KISI:
KOMPETENSI:
MEDIA:
ALT MEDIA:
```

**Contoh 2 — Menjodohkan (menjodohkan):**

```
SOAL 2
TIPE: menjodohkan
PERTANYAAN: Pasangkan kota dengan provinsinya.
A. Bandung::Jawa Barat
B. Semarang::Jawa Tengah
C. Surabaya::Jawa Timur
D. Medan::Sumatera Utara
KUNCI: Bandung::Jawa Barat;Semarang::Jawa Tengah;Surabaya::Jawa Timur;Medan::Sumatera Utara
BOBOT: 2
LEVEL: sedang
PEMBAHASAN:
KISI-KISI:
KOMPETENSI:
MEDIA:
ALT MEDIA:
```

**Contoh 3 — Uraian (uraian, kunci = rubrik):**

```
SOAL 3
TIPE: uraian
PERTANYAAN: Jelaskan perbedaan antara RAM dan ROM beserta contoh penggunaannya dalam kehidupan sehari-hari.
KUNCI: Rubrik: pengertian RAM (30), pengertian ROM (30), perbedaan (20), contoh (20).
BOBOT: 1
LEVEL: sulit
PEMBAHASAN:
KISI-KISI:
KOMPETENSI:
MEDIA:
ALT MEDIA:
```

Catatan untuk Tim Word: contoh memakai placeholder tautan `https://contoh.sekolah.id/...` bila contoh pilihan gambar ditambahkan; guru wajib menggantinya. Blok contoh harus bisa dihapus guru dan diganti soal sungguhan tanpa merusak struktur label.

---

## 13. Batasan yang perlu diketahui guru (jujur)

- **CSV tidak didukung** di UI baru. Pindahkan isi CSV lama ke template Excel.
- **Google Docs tidak didukung** langsung. Ekspor ke Word dulu, lalu sesuaikan ke format Bagian 4.
- **Gambar tertanam** di file Excel/Word tidak ikut terimpor. Gunakan tautan gambar.
- **Nama kolom kustom** untuk `isi_ganda` belum didukung (kolom jadi "Kolom 1", "Kolom 2", …). Tulis nama kolom di teks pertanyaan untuk sementara.
- **Stimulus bersama** belum ada: stimulus yang sama harus disalin ke tiap soal.
- **Level kognitif (C1–C6)** dan **estimasi waktu per soal** belum tersedia sebagai kolom import (belum ada kolomnya di database).
- **Tidak ada jalur AI** untuk import kunci: kunci selalu ditulis guru, tidak ditebak mesin.
- Soal hasil import selalu **DRAF** dan wajib ditelaah — import adalah asisten pengetikan, bukan jalan pintas validasi.
