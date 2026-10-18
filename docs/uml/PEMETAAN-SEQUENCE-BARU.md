# PEMETAAN SEQUENCE BARU — CBT SMK v2 (Fase 2)

> **Status:** Diagram Sequence dibangun ulang dari nol, 18 Okt 2026.
> **Dasar satu-satunya:** `ALUR-SEQUENCE-NYATA.md` (hasil telusur Fase 1
> dari kode). Diagram Sequence lama TIDAK dipakai sebagai dasar desain.
> **Skrip pembangun:** `rebuild_sequence_fase2.py` (idempotent, hanya
> mengubah `.mdj` lewat skrip; tanpa buka-simpan StarUML).
> **Backup:** `cbt-smk-v2.mdj.pre-sequence-redo.bak`
> (SHA-256 = SHA sebelum perubahan).
> **SHA `.mdj` sebelum:** `c5bd08be78b5e8b4d739231352b9bf6b5cc834c9627070682f2db7a9f737a016`
> **SHA `.mdj` sesudah:** `9b7bbc06260522c2c8e565a1defd33dc47532befd2255ab476a2c8ad31bb27ff`
> (rerun skrip menghasilkan SHA identik — idempotent terverifikasi).

## 1. Mengapa diagram lama dihapus

Dua diagram lama (Login + Wajib Ganti Password; Simpan Jawaban +
Auto-Submit) ditolak user karena alurnya tidak nyambung: self-message
ter-render di bawah diagram, ditambah Note penjelas posisi, dan banyak
alur berbeda dikompres ke 1 diagram. Perbaikan kali ini bukan menambal
diagram lama, melainkan menghapus isinya dan membangun 9 diagram baru
dari alur nyata Fase 1, dengan aturan: semua pesan panah normal
berurutan atas → bawah, tanpa self-message, tanpa Note penjelas
posisi.

Selain itu, diagram lama mengandung kesalahan isi alur (Fase 1 §6)
yang ikut diperbaiki di diagram baru — lihat §5.

## 2. Pemetaan diagram lama → diagram baru

| Diagram lama (dihapus) | Menjadi diagram baru | Keterangan |
|---|---|---|
| Login + Wajib Ganti Password | **Masuk** + **Ganti Kata Sandi** | Dipisah: masuk sukses adalah satu cerita; penolakan wajib ganti adalah cerita tersendiri yang berakhir di halaman masuk (bukan masuk ulang di diagram yang sama). |
| Simpan Jawaban + Auto-Submit | **Mulai Ujian**, **Kerjakan Ujian**, **Selesai, Koreksi dan Lihat Hasil** | Dipisah tiga: mulai (daftar + token + muat soal) mendahului autosave — di diagram lama alur ini hilang; finalisasi/nilai pindah ke diagram selesai. |
| (tidak ada — alur C tidak digambar di diagram lama) | **Pantau Ujian Langsung**, **Anti-Curang**, **Jeda Jadwal**, **Buka Blokir** | Baru dari Fase 1 Alur C. Di diagram lama, pemantauan hanya muncul sebagai satu panah siaran; pelanggaran, jeda, dan buka blokir tidak ada sama sekali. |

## 3. Daftar diagram Sequence baru

Total: **9 diagram, 59 lifeline, 209 pesan.** Semua pesan berurutan
atas → bawah, tanpa self-message, tanpa Note. Penomoran pesan mati
(`showSequenceNumber: false`) di semua diagram.

### D1 — Diagram Sequence Masuk (6 lifeline, 36 pesan)

- **Lifeline:** Pengguna, Halaman Masuk, Layanan Autentikasi,
  Layanan Keamanan, Layanan Sesi, Penyimpanan Data.
- **Alur:** Buka aplikasi → periksa sesi tersimpan (tidak ada) →
  tampilkan formulir → isi kredensial → masuk → periksa kunci masuk
  → cari pengguna → verifikasi kata sandi → buat sesi → arahkan ke
  beranda sesuai peran → rantai penjaga perubahan terlindungi
  (periksa sesi + periksa penanda keamanan) → keluar (hapus sesi).
- **Sumber Fase 1:** §A.1, §A.3, §A.4.
- **Alasan pemisahan:** satu siklus hidup sesi (buka → masuk →
  perubahan terlindungi → keluar) yang berurutan dan nyambung.

### D2 — Diagram Sequence Ganti Kata Sandi (7 lifeline, 20 pesan)

- **Lifeline:** Pengguna, Halaman Masuk, Halaman Ganti Kata Sandi,
  Layanan Autentikasi, Layanan Keamanan, Layanan Sesi,
  Penyimpanan Data.
- **Alur:** Percobaan masuk ditolak wajib ganti kata sandi (tanpa
  sesi) → buka formulir ganti → isi kata sandi lama + baru → ubah →
  verifikasi kata sandi lama → perbarui kata sandi → hapus SEMUA sesi
  pengguna → konfirmasi → arahkan ke halaman masuk (wajib masuk
  ulang; masuk ulang mengikuti D1).
- **Sumber Fase 1:** §A.2.
- **Alasan pemisahan:** cabang wajib-ganti adalah cerita sendiri
  dengan prasyarat berbeda (tanpa sesi, berakhir dengan penghapusan
  semua sesi). Mencampurnya ke D1 akan memutus alur masuk sukses.

### D3 — Diagram Sequence Mulai Ujian (7 lifeline, 25 pesan)

- **Lifeline:** Siswa, Halaman Ujian, Layanan Ujian,
  Layanan Batas Waktu, Layanan Keamanan, Penyimpanan Data,
  Layanan Pemantau.
- **Alur:** Buka daftar ujian hari ini → isi token → mulai ujian →
  periksa jadwal (waktu + toleransi telat) → periksa token → buat
  pengerjaan ujian (batas waktu = jam selesai jadwal) → siarkan ujian
  dimulai → minta soal → periksa batas waktu → muat soal tanpa kunci
  → tampilkan soal.
- **Sumber Fase 1:** §B.1, §B.2, §B.3.
- **Alasan pemisahan:** mulai ujian mendahului autosave dan punya
  urutan pemeriksaan sendiri (14 pemeriksaan di kode, diringkas
  menjadi pemeriksaan jadwal + token + pembuatan). Di diagram lama
  alur ini hilang — ujian dimulai langsung dari "pilih jawaban".

### D4 — Diagram Sequence Kerjakan Ujian (7 lifeline, 25 pesan)

- **Lifeline:** Siswa, Halaman Ujian, Layanan Ujian,
  Layanan Penilaian, Layanan Batas Waktu, Penyimpanan Data,
  Layanan Pemantau.
- **Alur:** Pilih jawaban → simpan jawaban (periksa batas waktu →
  hitung poin saat itu juga → tulis → siarkan → tampilkan tersimpan)
  → tandai ragu-ragu (simpan tanda → siarkan) → minta status ujian
  berkala (periksa batas waktu → hitung status → perbarui pewaktu).
- **Sumber Fase 1:** §B.4, §B.5, §B.6, §B.7, §B.9.
- **Alasan pemisahan:** satu sesi mengerjakan yang berurutan
  (jawab → ragu → pantau status). Finalisasi/kirim TIDAK di sini —
  itu D5 — agar diagram ini tidak melompat dari autosave ke nilai.

### D5 — Diagram Sequence Selesai, Koreksi dan Lihat Hasil (8 lifeline, 29 pesan)

- **Lifeline:** Siswa, Halaman Ujian, Halaman Hasil, Layanan Ujian,
  Layanan Penilaian, Penyimpanan Data, Guru, Layanan Pemantau.
- **Alur:** Pilih selesai → konfirmasi pengumpulan → kirim ujian
  (hitung nilai akhir → tandai selesai → siarkan → tampilkan kartu
  selesai) → guru koreksi jawaban manual (minta daftar → beri skor →
  simpan poin → hitung ulang nilai → perbarui nilai) → siswa buka
  hasil (minta hasil → tampilkan jawaban + poin + kunci + pembahasan
  sesuai aturan tampil).
- **Sumber Fase 1:** §B.8, §B.10.
- **Alasan pemisahan:** kirim → koreksi → lihat hasil adalah satu
  cerita pasca-ujian yang berurutan (nilai akhir lahir di kirim,
  berubah di koreksi, dibaca di hasil). Menggabungnya ke D4 akan
  mengompres tiga fase berbeda ke satu diagram.

### D6 — Diagram Sequence Pantau Ujian Langsung (5 lifeline, 18 pesan)

- **Lifeline:** Pemantau, Halaman Pantau, Layanan Pantau,
  Layanan Ujian, Penyimpanan Data.
- **Alur:** Buka halaman pantau → minta tiket pantauan sekali pakai
  → sambungkan pantauan langsung → cuplikan awal peserta → detak
  jantung berkala → perubahan ujian diteruskan sebagai pembaruan
  ringan → halaman meminta ulang cuplikan terbaru lewat jalur biasa
  → tampilkan terbaru. Langkah 17–18 adalah VARIAN fallback (minta
  ulang berkala 5 detik saat sambungan langsung mati), bukan lanjutan
  langkah 16 pada operasi normal.
- **Sumber Fase 1:** §C.1, §C.2.
- **Alasan pemisahan:** pemantauan adalah cerita staf yang
  independen dari cerita siswa (D3–D5): aktor, pemicu, dan arah
  pesannya berbeda. Di diagram lama ia hanya satu panah siaran.

### D7 — Diagram Sequence Anti-Curang (6 lifeline, 22 pesan)

- **Lifeline:** Siswa, Halaman Ujian, Layanan Ujian,
  Layanan Batas Waktu, Penyimpanan Data, Layanan Pemantau.
- **Alur:** Alihkan fokus → laporkan pelanggaran → periksa
  pelanggaran terakhir → catat → siarkan → tampilkan peringatan
  (pelanggaran ke-1) → alihkan fokus lagi → laporkan → catat →
  bekukan sisa waktu → tandai diblokir → siarkan → tampilkan layar
  diblokir (pelanggaran ke-2).
- **Sumber Fase 1:** §C.3.
- **Alasan pemisahan:** dua siklus pelanggaran yang berulang dan
  berurutan (peringatan → blokir) dengan akibat berbeda. Mencampurnya
  ke diagram pantau atau kerjakan akan mengaburkan ambang
  peringatan/blokir.

### D8 — Diagram Sequence Jeda Jadwal (6 lifeline, 18 pesan)

- **Lifeline:** Proktor, Halaman Pantau, Layanan Ujian,
  Layanan Batas Waktu, Penyimpanan Data, Layanan Pemantau.
- **Alur:** Pilih jeda → tandai jadwal dijeda → bekukan sisa waktu
  semua pengerjaan berjalan → siarkan jeda → tampilkan status dijeda
  → pilih lanjut → tandai jadwal aktif → pulihkan batas waktu
  (saat ini + sisa beku) → siarkan lanjutan → tampilkan status aktif.
- **Sumber Fase 1:** §C.4.
- **Alasan pemisahan:** jeda/lanjut adalah satu aksi proktor yang
  berpasangan dan berurutan (bekukan → pulihkan). Alurnya tidak
  berhubungan langsung dengan pelanggaran (D7) maupun buka blokir
  (D9) — ketiganya hanya berbagi aktor proktor, bukan cerita.

### D9 — Diagram Sequence Buka Blokir (7 lifeline, 16 pesan)

- **Lifeline:** Pengawas, Proktor, Halaman Pantau, Layanan Ujian,
  Layanan Batas Waktu, Penyimpanan Data, Layanan Pemantau.
- **Alur:** Minta daftar usulan menunggu berkala → tampilkan usulan
  → proktor pilih buka blokir → ambil pengerjaan diblokir → hitung
  sisa waktu lanjutan (dari sisa beku) → pulihkan pengerjaan menjadi
  berjalan dengan batas waktu baru → tandai usulan disetujui → siarkan
  status → tampilkan status berjalan.
- **Sumber Fase 1:** §C.5.
- **Alasan pemisahan:** buka blokir adalah cerita pemulihan dengan
  aktor ganda (pengawas mengusulkan lewat daftar, proktor memutuskan)
  dan aturan sisa waktu sendiri. Dipisah dari D8 karena jeda
  menyangkut jadwal, sedangkan buka blokir menyangkut satu pengerjaan.

## 4. Alasan tiap komponen desain (lifeline)

Lifeline adalah **komponen/subsistem desain yang kohesif**, bukan
fungsi telanjang. Ini yang membedakan dari kesalahan lama (satu fungsi
dijadikan lifeline):

| Komponen | Alasan / tanggung jawab | Dipakai di |
|---|---|---|
| Pengguna / Siswa / Guru / Proktor / Pengawas / Pemantau | Aktor manusia per peran. "Pemantau" adalah peran (proktor/admin/guru/pengawas yang memantau), bukan orang baru. | Semua |
| Halaman Masuk, Halaman Ganti Kata Sandi, Halaman Ujian, Halaman Hasil, Halaman Pantau | Tampilan per tugas. Satu halaman = satu tanggung jawab tampil; halaman yang berbeda tidak digabung agar panah tidak melompat antar tugas. | D1–D9 |
| Layanan Autentikasi | Masuk, ganti kata sandi, keluar, kunci masuk, dan rantai penjaga perubahan terlindungi. | D1, D2 |
| Layanan Keamanan | Verifikasi kata sandi, pemeriksaan kunci masuk, pemeriksaan penanda keamanan, pemeriksaan token. Satu tempat untuk semua pemeriksaan kepercayaan — inilah yang menggantikan self-message "verifikasi" lama menjadi pesan normal antar komponen. | D1, D2, D3 |
| Layanan Sesi | Membuat, memeriksa, dan menghapus sesi (termasuk hapus semua sesi saat ganti kata sandi). | D1, D2 |
| Layanan Ujian | Daftar jadwal, mulai ujian, simpan jawaban, ragu, status, kirim, koreksi, hasil, jeda, buka blokir, pelanggaran. Pengatur alur ujian. | D3–D9 |
| Layanan Penilaian | Hitung poin per soal (saat simpan) dan hitung nilai akhir (saat finalisasi/koreksi). Dipisah dari Layanan Ujian karena aturan nilainya berdiri sendiri dan dipakai di dua momen berbeda. | D4, D5 |
| Layanan Batas Waktu | Memeriksa batas waktu (finalisasi malas), membekukan sisa waktu (blokir/jeda), memulihkan batas waktu (lanjut/buka blokir), menghitung sisa lanjutan. Satu tempat untuk semua logika waktu server — inilah yang menggantikan self-message "periksa batas waktu" lama menjadi pesan normal antar komponen. | D3, D4, D7, D8, D9 |
| Layanan Pemantau | Tiket pantauan, cuplikan peserta, detak jantung, dan penyaluran siaran ke halaman pantau. | D3, D4, D5, D6, D7, D8, D9 |
| Penyimpanan Data | Semua baca/tulis data. Satu lifeline untuk penyimpanan agar diagram tidak melebar ke banyak tabel. | Semua |

Pesan yang sama boleh muncul dua kali dalam satu diagram bila operasi
yang sama sungguh dipanggil dua kali (mis. `periksaBatasWaktu()` saat
simpan dan saat minta status di D4; `hitungNilaiAkhir()` saat kirim
dan saat koreksi di D5; `alihkanFokus()` + `laporkanPelanggaran()`
untuk pelanggaran ke-1 dan ke-2 di D7). Itu bukan duplikasi gambar —
itu pengulangan alur yang nyata.

## 5. Koreksi isi alur terhadap diagram lama (dari Fase 1 §6)

1. Ganti kata sandi: diagram lama menempatkan pembuatan sesi hanya
   pada masuk ulang; yang benar — sesi dibuat pada SETIAP masuk
   sukses (D1), sedangkan ganti kata sandi menghapus SEMUA sesi
   lalu mewajibkan masuk ulang (D2).
2. Masuk: diagram lama tidak punya kunci masuk, akun nonaktif,
   kredensial salah, pemeriksaan penanda keamanan, umur sesi, dan
   navigasi berdasarkan peran — semuanya kini ada di D1 (sebagai
   pesan atau sebagai cabang/batasan di dokumentasi diagram).
3. Ujian: diagram lama memulai dari "pilih jawaban"; D3 menambahkan
   daftar + token + mulai + muat soal yang hilang itu.
4. Penilaian: diagram lama menempatkan hitung poin per soal saat
   kirim otomatis; yang benar — hitung poin per soal terjadi saat
   SIMPAN jawaban (D4), sedangkan finalisasi hanya agregasi
   (D5).
5. Batas waktu: diagram lama menggambarkannya satu kali sesudah
   polling status; yang benar — pemeriksaan batas waktu berjalan di
   awal tiap handler (muat soal di D3; simpan dan status di D4) dan
   dilewati bila jadwal dijeda.
6. Kirim otomatis: diagram lama menggambarkannya sebagai aksi
   proaktif server; yang benar — tidak ada penjadwal latar;
   finalisasi terjadi malas pada permintaan berikutnya (D4, cabang
   dokumentasi).
7. Siaran: diagram lama menggambarkannya sebagai pengiriman langsung
   ke monitor; yang benar — siaran adalah pembaruan ringan, halaman
   pantau menanggapinya dengan meminta ulang cuplikan lewat jalur
   biasa (D6).
8. Batas waktu ujian = jam selesai jadwal; durasi paket tidak dipakai
   menghitungnya (D3, batasan dokumentasi).
9. Diagram lama tidak punya ragu-ragu, pelanggaran, blokir, jeda,
   dan buka blokir — kini D4 (ragu) dan D7–D9.
10. Diagram lama tidak punya koreksi esai, nilai akhir, dan aturan
    tampil hasil — kini D5.

## 6. Konvensi gambar

- Pitch lifeline seragam 165 px (pusat ke pusat); lifeline pertama
  di x = 100, lebar 140. Panah horizontal murni (`y` sama di kedua
  ujung); tanpa garis diagonal, tanpa elemen melayang.
- Pesan pertama di y = 135, pitch vertikal 60 px; garis lifeline
  berakhir 40 px di bawah pesan terakhir.
- Label pesan: lebar = `floor(7,5 × panjang teks)`, diletakkan
  terpusat pada panah, dan TIDAK PERNAH lebih lebar dari bentang
  lifeline yang dilewati panah. Label yang panjang dipendekkan dan
  detailnya dipindah ke dokumentasi diagram (mis. `periksaJadwal()`
  mencakup jendela waktu + toleransi telat; `kembalikanPengguna()`
  membawa pengguna + izin).
- Jenis pesan: garis solid panah terisi = panggilan sinkron; garis
  putus-putus = balasan (reply); garis solid panah terbuka =
  panggilan asinkron (siaran, cuplikan, detak jantung — tanpa
  balasan).
- Tanpa Note di semua diagram Sequence baru (0 NoteView Sequence;
  2 NoteView yang tersisa adalah milik Use Case, tidak disentuh).

## 7. Batasan jujur — yang sengaja TIDAK digambar sebagai pesan alur

Sesuai Fase 1 §5 (hal yang tidak ditemukan di kode) dan aturan
tugas Fase 2, berikut ini TIDAK digambar sebagai pesan alur.
Masing-masing dicatat di dokumentasi diagram yang relevan:

1. **Kirim ulang autosave** — tidak ada. Jawaban yang gagal
   tersimpan tidak dikirim ulang otomatis (D4).
2. **Siaran ke siswa** — tidak ada. Siswa mengetahui blokir/jeda/
   selesai lewat minta status berkala atau permintaan berikutnya
   (D6, D8, D9).
3. **Penjadwal latar untuk kirim otomatis** — tidak ada. Finalisasi
   terjadi malas pada permintaan berikutnya (D4).
4. **Tombol pembuatan usulan buka blokir di halaman** — tidak ada.
   Yang digambar hanya daftar usulan menunggu + buka blokir langsung
   proktor (D9). Pembuatan usulan hanya tersedia lewat jalur layanan.
5. **Aksi minta buka blokir dari siswa** — tidak ada. Siswa yang
   diblokir hanya melihat pesan hubungi proktor/pengawas (D7).
6. **Jalur khusus mengambil penanda keamanan** — tidak dipakai
   halaman; penanda dibaca dari sesi (D1, batasan).
7. **Pengalihan balik ke halaman asal sesudah masuk** — tidak ada;
   pengalihan selalu berdasarkan peran (D1).
8. **Pesan berhasil ganti kata sandi di halaman masuk** — tidak
   ditampilkan (D2).
9. **Token ulang ke siswa** — tidak ada; token tampil sekali saat
   dibuat staf, token baru menghanguskan yang lama (D3).
10. **Pemeriksaan penanda keamanan pada ganti kata sandi mode
    kredensial** — tidak ada (disengaja, karena belum ada sesi) (D2).
11. **Remedial sebagai ujian ulang** — bukan; remedial adalah entri
    nilai manual, nilai tertinggi yang menang (D5, tidak digambar).
12. **Pencegahan salin/tempel sebagai pelanggaran** — hanya di sisi
    tampilan, bukan pelanggaran tercatat (D7).

Cabang yang hanya ditulis di dokumentasi diagram (bukan digambar
sebagai alur terpisah): kredensial salah, akun nonaktif, kunci masuk
terkunci, pengerjaan sudah berjalan/diblokir/selesai, jadwal
dijeda/selesai/belum mulai/sudah berakhir, lewat toleransi telat,
token salah/kedaluwarsa, paket tanpa soal, simpan ditolak
(diblokir/selesai/dijeda/luar paket), kirim idempotent, hasil sebelum
selesai, jeda/lanjut dari status yang salah, usulan ganda menunggu,
alasan kosong, sisa waktu lanjutan habis, peran pengawas tidak bisa
buka blokir langsung. Menggambar setiap cabang sebagai alur
terpisah akan mengembalikan masalah lama (diagram tidak nyambung);
alur utama tiap diagram adalah jalur sukses, dan cabang-cabangnya
tercatat eksplisit di dokumentasi diagram.

## 8. Verifikasi yang dilakukan

- Skrip membangun ulang dari nol dan lolos semua guard internal:
  semua `$ref` mengarah ke ID yang ada; 0 self-message di seluruh
  file; `showSequenceNumber = false` di 9 diagram; 0 NoteView di
  Sequence; label muat dalam bentang panah; tidak ada istilah
  implementasi di nama/dokumentasi Sequence baru.
- Perbandingan independen terhadap backup: 248 ID subtree Sequence
  lama terhapus tepat (tidak kurang, tidak lebih); 1535 ID subtree
  Sequence baru adalah tepat selisihnya; 5 diagram lain (Use Case,
  Class, ERD, Activity 2a/2b) tidak berubah — jumlah diagram dan isi
  elemennya identik dengan backup.
- Rerun skrip menghasilkan SHA identik (idempotent).

## 9. Yang belum / tidak diselesaikan di Fase 2

- **Belum ada uji buka di StarUML sungguhan.** Geometri mengikuti
  pola file lama yang valid (rumus label, struktur view, dan
  `edgePosition/distance/alpha` yang sama), dan lolos verifikasi
  struktural independen — tetapi render visual di aplikasi StarUML
  belum diklik/dilihat. Disarankan spot-check 1× saat Fase 3
  mengekspor (buka file → lihat 9 diagram → periksa label tidak
  tumpang tindih).
- **Fase 3 (di luar tugas ini)** menangani ekspor PNG/SVG + salin
  exe/zip/`your_files/` — tidak disentuh di sini.
- **Dokumen Fase 1 tidak diubah.** Diagram ERD, Use Case, Class,
  dan Activity tidak disentuh. Perilaku aplikasi/kode tidak diubah.

---

## 10. Fase 2b — Diagram Sequence untuk SEMUA UC + penanda UC & Utama/Lampiran (18 Okt 2026)

**Skrip pembangun:** `build_sequence_fase2b.py` (idempotent, hanya
mengubah `.mdj` lewat skrip; tanpa buka-simpan StarUML; memakai ulang
fungsi bangun + konvensi geometri Fase 2).
**Backup:** `cbt-smk-v2.mdj.pre-sequence-2b.bak`
(SHA-256 = SHA sebelum perubahan).
**SHA `.mdj` sebelum:** `9b7bbc06260522c2c8e565a1defd33dc47532befd2255ab476a2c8ad31bb27ff`
**SHA `.mdj` sesudah:** `118caee05be5f425c5c8f6f496f0133235ea55951e74f4a47fc460b7609a59dc`
(rerun skrip menghasilkan SHA identik — idempotent terverifikasi).

### 10.1 Aturan penanda UC + Utama/Lampiran

- Tiap diagram judulnya menyebut kode UC (mis. "Sequence — UC04
  Impor Soal") DAN field dokumentasinya menyebut kode UC yang
  dicakup, "Penempatan: UTAMA" atau "Penempatan: LAMPIRAN", dan
  alasan singkat.
- **Penanda diagram mengikuti UC PRIMER** (pemilik cerita diagram).
- Bila satu diagram mencakup beberapa UC, SEMUA UC dicatat di
  dokumentasi; penanda diagram tetap mengikuti UC primer.
- **Fragmen** (UC09 di D3; token/pause UC06 di D3/D8) dicatat
  sebagai rujukan, tidak mengubah penanda diagram induknya.
- Klasifikasi dipakai (dari Fase 1b §9, tanpa perubahan):
  **UTAMA** = UC01, UC04, UC07, UC08, UC09, UC10, UC11, UC18, UC19;
  **LAMPIRAN** = UC02, UC03, UC05, UC06, UC12, UC13, UC14, UC15,
  UC16, UC17.

### 10.2 Keputusan D5: TIDAK dipecah

D5 tetap satu diagram yang mencakup UC08 (kirim) + UC10 (koreksi) +
UC19 (lihat hasil). Alasan: ketiganya satu cerita pasca-ujian yang
berurutan — nilai akhir lahir di kirim, berubah di koreksi, dibaca
di hasil. Memecah akan memutus rantai sebab-akibat nilai yang sama
dan mengulang pesan kirim/koreksi di tiga diagram. Dokumentasi D5
menyebut ketiga UC secara eksplisit (lihat header dokumentasi D5).

### 10.3 Daftar LENGKAP diagram Sequence (D1–D21)

Total: **21 diagram, 130 lifeline, 601 pesan.** Semua pesan berurutan
atas → bawah, tanpa self-message, tanpa Note, penomoran pesan mati
(`showSequenceNumber: false`) di semua diagram. D1–D9: hanya nama +
dokumentasi yang berubah (judul + blok UC/Penempatan/Alasan);
pesan/lifeline/geometri D1–D9 TIDAK diubah.

| Diagram | UC yang dicakup | Penempatan | Lifeline | Pesan | Alasan penempatan / pemisahan |
|---|---|---|---|---|---|
| D1 — Sequence — UC01 Masuk | UC01 | UTAMA | 6 | 36 | UC01 gerbang semua alur. |
| D2 — Sequence — UC01 Ganti Kata Sandi (Wajib) | UC01 | UTAMA | 7 | 20 | Bagian UC01 (cabang wajib ganti). |
| D3 — Sequence — UC08 Mulai Ujian (fragmen UC09 Validasi Token) | UC08 (primer) + UC09 (fragmen) | UTAMA | 7 | 25 | UC09 tak terpisahkan dari mulai ujian; fragmen, bukan diagram sendiri. |
| D4 — Sequence — UC08 Kerjakan Ujian | UC08 | UTAMA | 7 | 25 | UC08 inti ujian. |
| D5 — Sequence — UC08/UC10/UC19 Selesai, Koreksi dan Lihat Hasil | UC08 (kirim) + UC10 + UC19 | UTAMA | 8 | 29 | Ketiga UC semuanya UTAMA; satu cerita pasca-ujian (lihat §10.2, tidak dipecah). |
| D6 — Sequence — UC07 Pantau Ujian Langsung | UC07 | UTAMA | 5 | 18 | Pantau disebut eksplisit untuk badan skripsi. |
| D7 — Sequence — UC08 Anti-Curang | UC08 (primer) + UC07 (sisi terima siaran) | UTAMA | 6 | 22 | Anti-curang terjadi selama siswa mengikuti ujian (UC08); siaran ke pemantau hanya efek samping. |
| D8 — Sequence — UC06 Jeda Jadwal | UC06 (primer) + UC08/UC07 (dampak) | LAMPIRAN | 6 | 18 | Jeda/lanjut adalah fungsi kelola jadwal (UC06); dirujuk dari alur UTAMA karena jeda terjadi saat ujian berjalan, tetapi rujukan tidak mengubah penanda. |
| D9 — Sequence — UC18 Buka Blokir | UC18 (primer) + UC07 (daftar usulan) | UTAMA | 7 | 16 | UC18 bagian loop pantau + anti-curang; tanpa buka blokir, pengerjaan diblokir tidak bisa pulih. |
| D10 — Sequence — UC02 Kelola Master Data | UC02 | LAMPIRAN | 6 | 38 | Prasyarat administratif (data referensi), bukan alur ujian. Termasuk impor siswa (tanpa pratinjau, tanpa mode semua-atau-tidak — berbeda dari impor soal). |
| D11 — Sequence — UC03 Kelola Bank Soal | UC03 | LAMPIRAN | 6 | 34 | Persiapan konten soal; kunci bank + pemisahan tugas adalah aturan keamanan, bukan alur ujian. |
| D12 — Sequence — UC04 Impor Soal | UC04 | UTAMA | 6 | 26 | Jalur utama pengisian bank soal (templat → pratinjau → simpan), disebut eksplisit sebagai impor. |
| D13 — Sequence — UC05 Susun Paket Ujian | UC05 | LAMPIRAN | 5 | 36 | Persiapan paket sebelum ujian; hasilkan acak + susun soal adalah setup. |
| D14 — Sequence — UC06 Kelola Jadwal Ujian (CRUD + Buat Token) | UC06 (CRUD + buat token) | LAMPIRAN | 5 | 30 | CRUD jadwal adalah setup. Validasi token dirujuk ke D3 (fragmen UC09); jeda/lanjut dirujuk ke D8. |
| D15 — Sequence — UC11 Rekap Nilai & Berita Acara | UC11 | UTAMA | 5 | 36 | Rekap penutup siklus ujian (finalisasi + remedial), disebut eksplisit. |
| D16 — Sequence — UC12 Analisis Butir Soal | UC12 | LAMPIRAN | 6 | 18 | Analitik pasca-ujian (daya beda/pengecoh); bukan penentu kelangsungan ujian. |
| D17 — Sequence — UC13 Nilai per Kompetensi | UC13 | LAMPIRAN | 6 | 18 | Analitik pasca-ujian per kompetensi; turunan dari data yang sama dengan UC12. |
| D18 — Sequence — UC14 Kelola UKK | UC14 | LAMPIRAN | 6 | 40 | Jalur ujian terpisah (UKK + sertifikat) di luar ujian CBT reguler; besar tapi bukan inti CBT. |
| D19 — Sequence — UC15 Kelola Kegiatan, Ruangan & Kartu Peserta | UC15 | LAMPIRAN | 6 | 36 | Dukungan administratif + kartu; alur ujian bisa dipahami tanpa detail kartu PDF. |
| D20 — Sequence — UC16 Cadangkan & Pulihkan | UC16 | LAMPIRAN | 6 | 38 | Operasional (cadangan/pulihkan), bukan alur fungsional ujian. |
| D21 — Sequence — UC17 Kelola Pengaturan, Audit & Notifikasi | UC17 | LAMPIRAN | 6 | 42 | Konfigurasi + jejak audit + pemberitahuan; tidak ada alur pengguna inti (jejak audit/pengecualian bahkan tanpa halaman). |

### 10.4 Konvensi yang dipakai diagram baru (sama seperti Fase 2)

- Pitch lifeline seragam 165 px; pesan pertama y = 135, pitch
  vertikal 60 px; label muat dalam bentang panah (dicek assert per
  pesan di skrip).
- Operasi internal → komponen desain sah (Layanan Audit, Layanan
  Berkas, Layanan Perhitungan, Layanan Basis Data, Penyimpanan
  Berkas), bukan fungsi telanjang; tanpa self-message; tanpa Note.
- Level PIM Bahasa Indonesia (tanpa nama file/endpoint/SQL/library/
  kode status implementasi); penomoran pesan tetap mati.
- Hal yang TIDAK ADA di kode (Fase 1 §5 + §10) TIDAK digambar
  sebagai pesan alur — dicatat di dokumentasi diagram: penolakan
  usulan UC18 (tidak ada setuju/tolak usulan — hanya unblock yang
  menandai usulan disetujui sebagai efek samping), kirim ulang
  autosave, siaran ke siswa, penjadwal latar untuk kirim otomatis,
  tombol pembuatan usulan buka blokir di halaman, pengiriman
  pemberitahuan untuk peristiwa selain pengecualian pemisahan
  tugas, cadangan otomatis di dalam aplikasi, unggah berkas
  cadangan dari luar, ubah urutan soal dalam paket, ubah rubrik
  UKK, sertifikat UKK sebagai PDF, impor CSV di panel impor soal.

### 10.5 Verifikasi yang dilakukan (Fase 2b)

- Guard internal skrip lolos: semua `$ref` mengarah ke ID yang ada;
  0 self-message di 21 diagram; `showSequenceNumber = false` di 21
  diagram; 0 NoteView di Sequence; label muat dalam bentang panah;
  tidak ada istilah implementasi di nama/dokumentasi Sequence baru +
  header D1–D9.
- Perbandingan independen: elemen di luar Sequence (3151 objek)
  tidak berubah; subtree D1–D9 tidak berubah kecuali
  nama/dokumentasi (dicek per-ID); 5 diagram lain (Use Case, Class,
  ERD, Activity 2a/2b) tidak berubah.
- `validate-mdj-real.py`: **LOLOS SEMUA, 0 pelanggaran**.
- Rerun skrip menghasilkan SHA identik (idempotent).

### 10.6 Yang belum / tidak diselesaikan di Fase 2b

- **Belum ada uji buka di StarUML sungguhan** (sama seperti Fase 2
  §9). Geometri mengikuti pola file lama yang valid dan lolos
  verifikasi struktural independen — tetapi render visual di
  aplikasi StarUML belum diklik/dilihat. Disarankan spot-check 1×
  saat Fase 3b mengekspor (buka file → lihat 21 diagram → periksa
  label tidak tumpang tindih).
- **Fase 3b (di luar tugas ini)** menangani ekspor PNG/SVG + salin
  exe/zip/`your_files/` — tidak disentuh di sini.
- **Dokumen telusur Fase 1/1b tidak diubah.** Diagram ERD, Use Case,
  Class, dan Activity tidak disentuh. Perilaku aplikasi/kode tidak
  diubah.

