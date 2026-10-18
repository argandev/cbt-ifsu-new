# R3 — Riset Konvensi Skripsi / Tugas Akhir Indonesia untuk Diagram UML

Proyek: CBT SMK v2 — TIM R3 (Riset Konvensi Skripsi/Tugas Akhir Indonesia untuk Diagram UML)
Tanggal riset: 17 Oktober 2026
Metode: `browser.search` (kueri Bahasa Indonesia), 17 kueri. File `docs/uml/cbt-smk-v2.mdj` TIDAK diubah.

## Cara membaca dokumen ini

- **OTORITATIF** = pedoman resmi perguruan tinggi (SK Rektor / buku pedoman resmi kampus).
  Berlaku hanya untuk kampus penerbitnya — TIDAK ada satu pedoman nasional yang
  mengatur notasi UML untuk semua skripsi di Indonesia.
- **LITERATUR / BAHAN AJAR** = buku, artikel jurnal, modul kuliah, tutorial UML yang
  banyak dikutip skripsi. Berpengaruh, tetapi BUKAN peraturan kampus.
- **CONTOH SATU SKRIPSI / SATU JURNAL** = praktik dalam satu karya. Menunjukkan
  "begini orang melakukannya", bukan "begini aturannya".
- **TIDAK DITEMUKAN SUMBER OTORITATIF — BUKAN ATURAN** = dalam pencarian ini tidak
  ditemukan pedoman resmi kampus yang menetapkan konvensi tersebut. Konvensi itu
  TIDAK BOLEH dijadikan aturan wajib untuk diagram CBT SMK v2.
- Semua URL di bawah ini disalin verbatim dari hasil pencarian web. Beberapa pedoman
  kampus hanya terjangkau lewat salinan cermin (123dok / Scribd / SlideShare /
  Academia.edu) — hal itu dicatat eksplisit, dan salinan cermin TIDAK diperlakukan
  setara dengan dokumen resmi di domain kampus.

---

## A. Sumber otoritatif (pedoman resmi kampus)

### A1. Universitas Bina Sarana Informatika (UBSI) — Fakultas Teknik dan Informatika

- **Dokumen:** *Pedoman Penyusunan Laporan Skripsi, Tugas Akhir dan Pelaksanaan
  Ujian Komprehensif, Program Sarjana (S1) dan Program Diploma Tiga (D3),
  Fakultas Teknik dan Informatika, Universitas Bina Sarana Informatika* —
  Keputusan Rektor UBSI Nomor 077/2.02/UBSI/III/2021.
- **URL (domain resmi kampus):**
  https://elibrary.bsi.ac.id/homepage/download/dokumen/20230216103710-pedoman-penyusunan-laporan-skripsi-tugas-akhir-dan-pelaksanaan-ujian-komprehensip-fakultas-teknik-dan-informatika.pdf
- **Status:** OTORITATIF (untuk UBSI / FTI UBSI).
- **Isi yang terverifikasi dari hasil pencarian:**
  - Bab II (Landasan Teori): "Peralatan pendukung (tools system) menggunakan
    Unified Modelling Language (UML), menjelaskan teori UML dan diagram-diagram
    yang digunakan yaitu Activity, Use Case, Sequence, Class dan Deployment,
    tapi jika dalam pembuatan perangkat lunak tidak menggunakan class maka teori
    dan pembahasan hanya meliputi empat diagram UML yaitu Activity, Use Case,
    Sequence, Deployment, dalam teori UML ini harus memiliki konsistensi
    terhadap pembahasan."
  - Lampiran 33 (Daftar Simbol): "Simbol UML — sesuaikan dengan symbol UML yang
    dipakai dalam penulisan skripsi, sesuaikan dengan referensi yang digunakan."
    (Artinya: UBSI TIDAK menetapkan satu set simbol UML baku sendiri; mahasiswa
    wajib konsisten dengan referensi yang dipakainya.)
- **Konvensi yang didukung:**
  1. **Konsistensi teori ↔ pembahasan UML adalah WAJIB** (teori diagram yang
     ditulis di Bab II harus sama dengan diagram yang dibahas di Bab III/IV).
  2. **Set diagram bersifat kondisional:** 5 diagram (Activity, Use Case,
     Sequence, Class, Deployment) bila program memakai class; 4 diagram
     (tanpa Class) bila tidak memakai class.
  3. **Tidak ada notasi UML baku versi UBSI** — rujuk ke referensi yang dipakai
     dan konsisten dengan referensi itu.

### A2. Universitas Ciputra — Program Sarjana

- **Dokumen:** *Pedoman Penulisan Skripsi / Tugas Akhir Mahasiswa Program Sarjana,
  Universitas Ciputra*.
- **URL (salinan Academia.edu — BUKAN domain resmi kampus, perlu verifikasi ke
  dokumen resmi UC bila dipakai sebagai acuan formal):**
  https://www.academia.edu/8263617/PEDOMAN_PENULISAN_SKRIPSI_TUGAS_AKHIR_MAHASISWA_PROGRAM_SARJANA_UNIVERSITAS_CIPUTRA
- **Status:** OTORITATIF untuk Universitas Ciputra (isinya pedoman resmi; akses
  lewat salinan cermin, bukan situs resmi UC).
- **Isi yang terverifikasi:**
  - "Penulisan Gambar dan Tabel mengikuti ketentuan: (a) gambar dan tabel harus
    diberi judul; (b) ditempatkan di antara bagian teks yang paling banyak
    membahasnya; (c) gambar dan tabel harus terletak dalam satu halaman, tidak
    terpotong; (d) judul gambar diletakkan di bawah gambar dan ditulis dengan
    Times New Roman 10 point dan spasi 1 rata kiri; (e) judul tabel diletakkan di
    atas tabel; (f) gambar dan tabel harus diberi nomor dengan ketentuan bahwa
    angka pertama merujuk pada bab tempat gambar/tabel tersebut berada dan angka
    kedua menunjukkan urutan gambar/tabel dalam bab tersebut; (g) gambar/tabel
    terlalu besar boleh landscape / kertas A3 dilipat / diperkecil ke A4."
  - Pedoman ini juga menegaskan detail format "diserahkan kepada masing-masing
    Program Studi" — jadi format di tingkat prodi bisa berbeda.
- **Konvensi yang didukung:**
  1. **Caption diagram = "Gambar `<bab>.<urut>`", judul di BAWAH gambar.**
  2. **Gambar + judulnya harus dalam satu halaman, tidak terpotong.**
  3. **Detail tipografi caption (TNR 10, spasi 1, rata kiri) adalah aturan UC,**
     bukan aturan universal — kampus lain menetapkan detail berbeda.

### A3. STMIK Nusa Mandiri Jakarta — Pedoman Laporan Skripsi Periode I 2021

- **Dokumen:** *Pedoman Lap Skripsi STMIK Nusa Mandiri Jakarta Periode I 2021*.
- **URL (salinan SlideShare / Scribd — BUKAN domain resmi kampus):**
  https://www.slideshare.net/slideshow/pedoman-lap-skripsi-stmik-nusa-mandiri-jakarta-periode-i-2021pdf/258892158
  https://id.scribd.com/document/505044488/Pedoman-Lap-Skripsi-STMIK-Nusa-Mandiri-Jakarta-Periode-I-2021
- **Status:** OTORITATIF untuk STMIK Nusa Mandiri (isi pedoman resmi; akses lewat
  salinan cermin).
- **Isi yang terverifikasi:**
  - "Desain — Software Architecture: 1. Jika program yang dibuat adalah
    pemrograman berorientasi objek (OOP), pada tahapan ini menggambarkan class
    diagram, sequence diagram, component diagram, dan deployment diagram
    (terkait dengan program yang dibuat dan yang hanya berhubungan dengan proses
    bisnis sistem saja). 2. Jika program yang dibuat adalah pemrograman
    terstruktur, pada tahapan ini menggambarkan deployment diagram dan component
    diagram."
  - Versi lain dari kutipan yang sama menambahkan activity diagram ke kedua
    kelompok di atas.
  - "Desain Database: jika menggunakan ERD maka harus menggambarkan ERD, LRS,
    dan juga spesifikasi filenya."
  - "Proses Bisnis: gunakan Activity Diagram, gambar diletakkan setelah narasi
    proses bisnis." / "Analisa Kebutuhan Sistem: disertai penggambaran use case
    diagram dan activity diagram."
- **Konvensi yang didukung:**
  1. **Set diagram tergantung paradigma (OOP vs terstruktur), bukan satu set
     universal.**
  2. **Setiap diagram UML harus terkait dengan proses bisnis sistem** (bukan
     diagram hiasan).
  3. **Activity diagram diletakkan setelah narasi proses bisnis** (aturan urutan
     penyajian, bukan aturan notasi).

### A4. STMIK Pelita Nusantara Medan — Pedoman Penulisan Usulan Proposal dan Skripsi S1

- **Dokumen:** *Pedoman Penulisan Usulan Proposal dan Skripsi Jenjang Strata Satu
  (S1), STMIK Pelita Nusantara Medan*.
- **URL (salinan 123dok — BUKAN domain resmi kampus):**
  https://123dok.com/document/qv7leldq-pedoman-penulisan-usulan-proposal-skripsi-jenjang-strata-nusantara.html
- **Status:** OTORITATIF untuk STMIK Pelita Nusantara Medan (isi pedoman resmi;
  akses lewat salinan cermin).
- **Isi yang terverifikasi:**
  - Bab perancangan: "Use Case Diagram + Class Diagram + Activity/State Diagram
    + Sequence Diagram (bilamana menggunakan UML) atau Flowchart."
  - Struktur Bab 3 (Analisa dan Perancangan): 3.2.1 Use Case + Class +
    Activity/State + Sequence + Flowchart; 3.2.3 Desain Database (jika
    menggunakan database).
- **Konvensi yang didukung:** set diagram perancangan = Use Case + Class +
  Activity/State + Sequence (bila memakai UML). Sekali lagi kondisional dan
  khas kampus tersebut.

### A5. STMIK Nuri — Buku Pedoman Skripsi

- **Dokumen:** *Buku Pedoman Skripsi STMIK Nuri* (Program Studi Teknik Informatika).
- **URL (salinan 123dok — BUKAN domain resmi kampus):**
  https://123dok.com/document/zlren36z-buku-pedoman-skripsi-stmik-nuri-dosen.html
- **Status:** OTORITATIF untuk STMIK Nuri (isi pedoman resmi; akses lewat salinan
  cermin).
- **Isi yang terverifikasi:**
  - "Desain aplikasi: a. Jika program menggunakan pemrograman terstruktur dan
    desain menggunakan UML, diagram yang digunakan adalah Activity diagram,
    Deployment diagram, Component, jika tidak menggunakan UML dibolehkan
    menggunakan DFD dan Flowchart. b. Jika program yang digunakan OOP serta
    tidak menggunakan class maka hanya menggunakan diagram UML (Activity,
    Use Case, Sequence, Deployment), jika program yang digunakan menggunakan
    class maka 5 diagram UML (Activity, Use Case, Sequence, Class, Deployment)."
  - "Desain Database menggunakan Entity Relationship Diagram (ERD)."
- **Konvensi yang didukung:** sama dengan pola UBSI — set diagram kondisional
  (4 vs 5 diagram) menurut pemakaian class; ERD untuk desain database.

### A6. Universitas Indonesia — Pedoman Teknis Penulisan Tugas Akhir

- **Dokumen:** *Pedoman Teknis Penulisan Tugas Akhir Mahasiswa Universitas
  Indonesia*.
- **URL (salinan 123dok / DocPlayer / Academia.edu — BUKAN domain resmi UI):**
  https://123dok.com/document/zk693epy-universitas-indonesia-pedoman-teknis-penulisan-mahasiswa-universitas-indonesia.html
  http://docplayer.info/32587806-Universitas-indonesia-pedoman-teknis-penulisan-tugas-akhir-mahasiswa-universitas-indonesia.html
- **Status:** OTORITATIF untuk UI (isi pedoman resmi; akses lewat salinan cermin).
- **Isi yang terverifikasi:** mengatur kertas HVS 80 gram A4, margin
  (kiri 4 cm, kanan 3 cm, atas 3 cm, bawah 3 cm), huruf Times New Roman 12,
  spasi 1,5, penomoran halaman, sampul. **TIDAK ditemukan ketentuan tentang
  notasi UML** dalam cuplikan yang terjangkau.
- **Kesimpulan:** pedoman UI adalah bukti negatif yang penting — pedoman kampus
  besar pun tidak selalu mengatur notasi UML; yang diatur adalah format naskah.

### A7. Pedoman gambar generik (institusi tidak teridentifikasi dari cuplikan)

- **Dokumen:** *Pedoman Penulisan Skripsi* (cuplikan 123dok, institusi penerbit
  tidak terbaca dari hasil pencarian).
- **URL:** https://123dok.com/document/q59e9erz-pedoman-penulisan-skripsi.html
- **Status:** BUKAN OTORITATIF yang dapat diklaim untuk kampus tertentu —
  institusinya tidak teridentifikasi. Hanya bukti bahwa pola ini lazim.
- **Isi:** "Nomor urut gambar ditulis dengan 2 (dua) angka Arab, angka pertama
  menunjukkan nomor bab, dan angka kedua menunjukkan nomor urut gambar dalam
  bab tersebut, misalnya: 1.1, 2.3, 4.1, dst. Judul gambar ditulis di bagian
  bawah gambar dengan huruf pertama huruf kapital. Jarak antar baris dari judul
  gambar adalah satu spasi." + "Gambar yang dikutip dari sumber pustaka diberi
  keterangan di bagian bawah gambar (nama akhir penulis + tahun)."
- **Konvensi yang didukung (sebagai pola umum, bukan aturan kampus tertentu):**
  caption "Gambar `<bab>.<urut>`", judul di bawah gambar, huruf pertama kapital.

---

## B. Literatur / bahan ajar yang banyak dikutip skripsi (BUKAN pedoman kampus)

### B1. Dharwiyanti & Wahono (2003) — "Pengenalan UML"

- **Dokumen:** Sri Dharwiyanti & Romi Satria Wahono, *Pengenalan Unified Modeling
  Language (UML)*, 2003 (artikel/tutorial; diakses lewat salinan 123dok).
- **URL (salinan):** https://123dok.com/document/y93jj5vy-sri-dharwiyanti-romi-satria-wahono.html
- **Status:** LITERATUR / TUTORIAL — sangat banyak dikutip skripsi Indonesia,
  tetapi BUKAN pedoman resmi kampus mana pun.
- **Isi yang terverifikasi:**
  - Sequence diagram: dimensi vertikal = waktu, dimensi horizontal = objek-objek
    terkait; message = garis berpanah dari satu objek ke objek lain; pada fase
    desain berikutnya message dipetakan menjadi operasi/metode class; activation
    bar menunjukkan lamanya eksekusi, diawali diterimanya message; "untuk
    objek-objek yang memiliki sifat khusus, standar UML mendefinisikan icon
    khusus untuk objek boundary, controller dan persistent entity."
  - **Collaboration diagram (BUKAN sequence):** "Setiap message memiliki
    sequence number, di mana message dari level tertinggi memiliki nomor 1.
    Messages dari level yang sama memiliki prefiks yang sama."
  - Class: tiga area pokok — nama dan stereotype, atribut, metoda; atribut dan
    metoda dapat bersifat Private (tidak dapat dipanggil dari luar class),
    Protected (hanya class ybs + anak yang mewarisinya), Public (dapat
    dipanggil siapa saja).
  - Statechart: "Transisi antar state umumnya memiliki kondisi guard yang
    merupakan syarat terjadinya transisi yang bersangkutan, dituliskan dalam
    kurung siku. Action yang dilakukan sebagai akibat dari event tertentu
    dituliskan dengan diawali garis miring."
  - Format detail transisi (dikutip di materi turunan): `event [guard condition] / action`.
- **Konvensi yang didukung (sebagai literatur, bukan aturan kampus):**
  1. Urutan waktu sequence = atas → bawah; objek berjajar kiri → kanan.
  2. Penomoran `1`, `1.1`, `2` (prefiks level sama) berlaku untuk
     **Collaboration / Communication diagram**, bukan untuk sequence diagram.
  3. Guard ditulis dalam kurung siku `[kondisi]` — terbukti untuk **transisi
     statechart**; format `event [guard]/action` juga untuk transisi state.
  4. Class = 3 kompartemen (nama+stereotip, atribut, metoda); visibilitas
     Private / Protected / Public.

### B2. Sholiq (2006 / 2010) — Pemodelan Sistem Berorientasi Objek, dikutip banyak skripsi

- **Dokumen:** Buku Sholiq, *Pemodelan Sistem Berorientasi Objek dengan UML*
  (2006; edisi/kutipan 2010), diakses lewat kutipan di berbagai skripsi.
- **URL contoh kutipan (skripsi yang mengutipnya):**
  https://123dok.com/article/sequence-diagram-perancangan-proses-analisis-dan-perancangan-sistem.zpnplnvy
- **Status:** LITERATUR (buku teks) — dikutip sebagai referensi teori, bukan
  pedoman kampus.
- **Isi yang terverifikasi (lewat kutipan skripsi):** "Obyek-obyek tersebut
  kemudian diurutkan dari kiri ke kanan, aktor yang melakukan interaksi biasanya
  ditaruh di paling kiri dari diagram (Sholiq, 2006)." / "Sequence diagram
  mampu menceritakan situasi yang rumit dan dinamis mencakup object yang
  dihasilkan oleh class diagram (Sholiq, 2006)."
- **Konvensi yang didukung (sebagai literatur):** aktor biasanya paling kiri
  pada sequence diagram. Kata kuncinya "biasanya" — bukan kewajiban mutlak.

### B3. Nugroho (2010) — UML, dikutip banyak skripsi

- **Dokumen:** Nugroho, buku UML (dikutip sebagai "Nugroho (2010:10)" dan
  "(2010:117)" di makalah-makalah UML Indonesia).
- **URL contoh kutipan:**
  https://id.scribd.com/doc/225983677/MAKALAH-UML
- **Status:** LITERATUR (buku teks).
- **Isi yang terverifikasi (lewat kutipan):** bangunan dasar UML = things,
  relationships, diagrams; view dibagi structural classification, dynamic
  behaviour, model management.
- **Konvensi yang didukung:** kerangka klasifikasi diagram UML (struktural /
  perilaku / manajemen model). Tidak ada aturan penomoran pesan atau urutan
  lifeline.

### B4. Henderi (2008) — Langkah-langkah penggunaan UML, dikutip di materi kuliah

- **Dokumen:** Henderi (2008:6), dikutip di slide/materi "Perancangan UML dan
  Use Case Diagram" dan modul-modul UML Indonesia.
- **URL contoh kutipan:**
  https://id.scribd.com/document/827294473/8-UML
- **Status:** LITERATUR / BAHAN AJAR.
- **Isi yang terverifikasi (10 langkah):** (1) daftar business process; (2) petakan
  use case per business process; (3) deployment diagram kasar; (4) requirement
  non-fungsional/security; (5) activity diagram dari use case; (6) sequence
  dan/atau collaboration per alur pekerjaan — "jika sebuah use case memiliki
  kemungkinan alur normal dan error, buat lagi satu diagram untuk
  masing-masing alur"; (7) rancangan UI; (8) class diagram; (9) component
  diagram; (10) perhalus deployment diagram.
- **Konvensi yang didukung (sebagai bahan ajar):** satu diagram sequence per
  alur (normal dan error dipisah); class diagram dibangun sesudah model-model
  sebelumnya. Bukan aturan pedoman kampus.

### B5. Modul Politeknik Telkom — "Pengenalan UML dan Analisis Use Case"

- **Dokumen:** Modul *Analisis dan Desain Sistem Informasi*, Politeknik Telkom —
  bab "Pengenalan UML dan Analisis Use Case".
- **URL (salinan 123dok):**
  https://text-id.123dok.com/document/wq2opmrpz-pengenalan-uml-dan-analisi-use-case-analisis-dan-desain-sistem-informasi-1.html
- **Status:** BAHAN AJAR kampus (modul kuliah), bukan pedoman penulisan skripsi.
- **Isi yang terverifikasi:**
  - Use case: "biasanya dinyatakan dengan menggunakan kata kerja di awal frase
    nama use case."
  - Aktor: "biasanya dinyatakan menggunakan kata benda di awal frase nama aktor."
  - Include `<<include>>` / `<<uses>>`, extend `<<extend>>`, generalisasi.
- **Konvensi yang didukung (sebagai bahan ajar):** nama use case diawali kata
  kerja; nama aktor diawali kata benda. Kata kuncinya "biasanya".

### B6. Materi dosen UPI YAI — "Unified Modelling Language" (PDF di domain kampus)

- **Dokumen:** Materi kuliah UML, dosen Universitas Persada Indonesia YAI.
- **URL (domain kampus):**
  https://dosen.upi-yai.ac.id/v5/dokumen/materi/060060/52_20241210032909_UML.pdf
- **Status:** BAHAN AJAR (materi kuliah di domain kampus), bukan pedoman skripsi.
- **Isi yang terverifikasi:** statechart — guard = syarat transisi, "dituliskan
  dalam kurung siku"; action diawali garis miring; titik awal/akhir lingkaran
  penuh/setengah.
- **Konvensi yang didukung (sebagai bahan ajar):** guard `[…]` untuk transisi
  statechart — konsisten dengan B1.

### B7. OMG — Unified Modeling Language, Version 2.5 (standar internasional)

- **Dokumen:** Object Management Group (OMG), *OMG Unified Modeling Language
  (OMG UML), Version 2.5*, 2015.
- **URL:** https://www.omg.org/spec/UML/2.5
  (URL ini muncul di daftar pustaka artikel "Pemodelan Sistem Informasi Dengan
  UML", Academia.edu: https://www.academia.edu/127867864/PEMODELAN_SISTEM_INFORMASI_DENGAN_UML)
- **Status:** STANDAR INTERNASIONAL UML — otoritatif untuk notasi UML secara
  umum, tetapi BUKAN pedoman skripsi Indonesia dan TIDAK mengatur format
  naskah skripsi (caption, bab, dll).
- **Catatan:** OMG UML 2.5 adalah acuan yang sah bila tim perlu membenarkan
  notasi (mis. combined fragment `alt`/`opt` dengan guard) — tetapi klaim
  "wajib untuk skripsi" tetap tidak dapat disandarkan pada OMG, karena OMG
  tidak mengatur skripsi.

---

## C. Contoh satu skripsi / satu jurnal (BUKAN aturan)

### C1. Skripsi UNPAS — Rancang Bangun Aplikasi Otomatisasi E-Sales

- **Dokumen:** Skripsi Teknik Informatika, Universitas Pasundan (repositori
  resmi UNPAS).
- **URL:** https://repository.unpas.ac.id/83672/1/Mhs_213040108_Teknik%20Informatika%20%281%29.pdf
  (Catatan: satu percobaan fetch langsung ke URL ini gagal timeout di sisi
  tool; isinya di bawah dikutip dari cuplikan hasil pencarian, bukan dari
  PDF yang berhasil diunduh penuh.)
- **Status:** CONTOH SATU SKRIPSI (repositori resmi, tetapi satu karya).
- **Isi yang terverifikasi (dari cuplikan pencarian):** "Dalam pengembangan
  perangkat lunak berbasis UML, objek pada Sequence Diagram diklasifikasikan
  menjadi tiga stereotipe, yaitu boundary, control, dan entity. Pendekatan ini
  pertama kali diperkenalkan oleh Jacobson dan digunakan secara luas untuk
  mendukung arsitektur Model–View–Controller (MVC). Menurut Kurniawan dkk.
  [KUR20], objek boundary berfungsi sebagai antarmuka yang menghubungkan aktor
  dengan sistem. … objek entity merepresentasikan informasi yang bersifat
  permanen … objek control berperan sebagai pengelola alur logika dari skenario
  use case."
- **Nilai bukti:** membuktikan klasifikasi BCE dipakai di skripsi Indonesia dan
  disandarkan ke Jacobson + Kurniawan dkk. TIDAK membuktikan adanya aturan
  "urutan lifeline harus Actor → Boundary → Control → Entity".

### C2. Skripsi UIN Ar-Raniry — Perancangan SI Dokumentasi Tugas Akhir + Crowdfunding

- **Dokumen:** Skripsi, UIN Ar-Raniry Banda Aceh (repositori resmi).
- **URL:** https://repository.ar-raniry.ac.id/id/eprint/39680/contents
- **Status:** CONTOH SATU SKRIPSI.
- **Isi:** metode RAD; "Kerangka aplikasi dibuat menggunakan UML. UML yang
  digunakan meliputi activity diagram, use case diagram, class diagram dan
  sequence diagram. Sedangkan untuk rancangan basis data menggunakan ERD."
  Caption contoh: "Gambar III.1 Rapid Application Development".
- **Nilai bukti:** contoh set 4 diagram + ERD terpisah; contoh caption
  "Gambar III.1" (angka Romawi bab) — VARIASI dari pola "Gambar 3.1",
  membuktikan format caption TIDAK seragam antar kampus/karya.

### C3. Skripsi Universitas BSI — Perancangan SI Pendaftaran Pelayanan BLK Kota Bogor

- **Dokumen:** Tugas Akhir, Universitas BSI (repositori resmi BSI).
- **URL:** https://repository.bsi.ac.id/repo/files/485249/download/12220494_M-Nur-Nugroho.pdf
- **Status:** CONTOH SATU TUGAS AKHIR.
- **Isi:** "Untuk memberikan penjelasan lebih lanjut tentang rancangan pembuatan
  program menggunakan diagram UML yang terdiri dari Use Case Diagram, Activity
  Diagram, Class Diagram dan Sequence Diagram. Untuk perancangan database
  menggunakan ERD dan LRS."
- **Nilai bukti:** contoh set 4 diagram + ERD/LRS.

### C4. Skripsi Universitas Bangka Selatan — Analisis & Perancangan SI Kepegawaian BKD

- **Dokumen:** Skripsi (salinan 123dok).
- **URL:** https://123dok.com/document/qo5x5w7k-analisis-perancangan-informasi-kepegawaian-kepegawaian-kabupaten-selatan-skripsi.html
- **Status:** CONTOH SATU SKRIPSI.
- **Isi:** tabel simbol sequence diagram — Actor, Boundary ("pembatas sistem
  dengan dunia luar"), Control ("mengkoordinasikan perilaku sistem"),
  Entity ("informasi yang harus disimpan"), Object Message, Message to Self,
  Return Message, Activation; multiplicity "1 tepat Satu, 0..* Nol atau lebih,
  1..* Satu atau lebih, 0..1 Nol atau satu".
- **Nilai bukti:** contoh bahwa simbol BCE + multiplicity dipakai di skripsi.
  Bukan aturan pedoman.

### C5. Jurnal — Yusup & Anwar (2024), IKRA-ITH Informatika (UPI YAI)

- **Dokumen:** Yusup, M. E., & Anwar, N. (2024). "Studi Kasus Pemodelan Sistem
  Penjualan Stiker: Implementasi UML dengan Python dan PlantUML."
  *IKRA-ITH Informatika: Jurnal Komputer dan Informatika*, 8(3), 174–184.
- **URL:** http://journals.upi-yai.ac.id/index.php/ikraith-informatika/article/view/4508
- **Status:** CONTOH SATU ARTIKEL JURNAL (terakreditasi jurnal kampus, bukan
  pedoman).
- **Isi:** "Diagram tersebut dibuat dengan memanfaatkan library plantuml …
  Use Case Diagram menunjukkan fungsionalitas sistem dari sudut pandang
  pengguna. Class Diagram memodelkan struktur data dan relasi antar entitas.
  Sequence Diagram mengilustrasikan interaksi antar objek, sedangkan Activity
  Diagram memperjelas alur aktivitas dalam sistem."
- **Nilai bukti:** contoh set 4 diagram sebagai praktik lazim di jurnal Indonesia.

### C6. Jurnal — Setiawan, Agustin & Ningsih (2022), Jurnal Algoritma (ITG)

- **Dokumen:** "Perancangan Sistem Informasi Kerja Praktik Berbasis Web",
  *Jurnal Algoritma*, 19(1), 42–53.
- **URL:** https://jurnal.itg.ac.id/index.php/algoritma/article/view/996
- **Status:** CONTOH SATU ARTIKEL JURNAL.
- **Isi:** metodologi Rational Unified Process (inception, elaboration,
  construction) dengan "use case diagram, activity diagram, sequence diagram,
  dan class diagram".
- **Nilai bukti:** contoh set 4 diagram dalam kerangka RUP.

### C7. Jurnal — artikel-artikel perancangan UML lain (contoh pola umum)

- Pasassa dkk., "Sistem Informasi pada Kantor Desa Pettalandung Berbasis
  Website", *JITET* Univ. Lampung, DOI 10.23960/jitet.v13i2.6141 —
  https://journal.eng.unila.ac.id/index.php/jitet/article/view/6141
  (UML: use case, activity, sequence, class; waterfall; draw.io).
- Dewi & Voutama, "Implementasi UML dalam Perancangan SI Keuangan Masjid
  Al-Amanah Berbasis Web", *JATI* —
  https://ejournal.itn.ac.id/jati/article/view/11026
  (use case, class, sequence, activity).
- Septianto & Voutama, "Perancangan SI Katalog Makanan Warteg Berbasis Web
  Menggunakan UML", *JITET* —
  https://journal.eng.unila.ac.id/index.php/jitet/article/view/7592
  (use case, activity, sequence, class).
- **Status:** CONTOH JURNAL. Nilai bukti: pola set 4 diagram (Use Case,
  Activity, Sequence, Class) adalah pola yang sangat lazim di jurnal Indonesia
  — tetapi kelaziman ≠ aturan pedoman.

### C8. Ropianto — "Pemahaman Penggunaan UML" (STT Ibnu Sina Batam)

- **Dokumen:** Muhammad Ropianto, "Pemahaman Penggunaan Unified Modelling
  Language", Program Studi Teknik Informatika, STT Ibnu Sina Batam.
- **URL:** https://123dok.com/document/q5pp087y-pemahaman-penggunaan-unified-modelling-language.html
- **Status:** ARTIKEL JURNAL / TUTORIAL.
- **Isi:** UML diturunkan dari notasi Booch, OMT Rumbaugh, OOSE Jacobson;
  tiga kategori: struktur, behaviour, interaction.
- **Nilai bukti:** konteks sejarah UML (Three Amigos) — bukan aturan notasi.

---

## D. Daftar konvensi — putusan per konvensi

| # | Konvensi yang diuji | Putusan | Dasar |
|---|---|---|---|
| D1 | Teori UML di Bab II harus konsisten dengan diagram yang dibahas di Bab III/IV | **DIDUKUNG SUMBER OTORITATIF (UBSI)** | A1 |
| D2 | Set diagram kondisional: 5 diagram bila pakai class, 4 bila tidak | **DIDUKUNG SUMBER OTORITATIF (UBSI, STMIK Nuri)** | A1, A5 |
| D3 | Set diagram tergantung paradigma OOP vs terstruktur | **DIDUKUNG SUMBER OTORITATIF (STMIK Nusa Mandiri, STMIK Pelita Nusantara)** | A3, A4 |
| D4 | Setiap diagram UML harus terkait proses bisnis sistem | **DIDUKUNG SUMBER OTORITATIF (STMIK Nusa Mandiri)** | A3 |
| D5 | Caption diagram: "Gambar `<bab>.<urut>`", judul DI BAWAH gambar, gambar+judul satu halaman tidak terpotong | **DIDUKUNG SUMBER OTORITATIF (Univ. Ciputra)** untuk pola umumnya; detail tipografi (TNR 10, rata kiri) hanya aturan UC | A2, A7 |
| D6 | Caption memakai angka Romawi bab ("Gambar III.1") | **VARIASI CONTOH, BUKAN ATURAN** — ditemukan di satu skripsi (C2); kampus lain memakai "Gambar 3.1". Ikuti pedoman kampus target | C2 |
| D7 | Aktor diletakkan paling kiri pada sequence diagram | **DIDUKUNG LITERATUR (Sholiq 2006, kata "biasanya") — BUKAN pedoman kampus** | B2 |
| D8 | Urutan waktu sequence: pesan dari atas ke bawah; objek berjajar kiri→kanan | **DIDUKUNG LITERATUR (Dharwiyanti & Wahono 2003)** — konsisten dengan UML | B1 |
| D9 | Urutan lifeline WAJIB Actor → Boundary → Control → Entity | **TIDAK DITEMUKAN SUMBER OTORITATIF — BUKAN ATURAN.** Klasifikasi BCE terbukti dipakai (C1, C4, B1), tetapi tidak ada sumber yang menetapkan urutan kiri→kanan itu sebagai kewajiban | C1, C4, B1 |
| D10 | Klasifikasi objek sequence menjadi Boundary / Control / Entity (asal Jacobson; dipakai untuk MVC) | **DIDUKUNG CONTOH SKRIPSI + LITERATUR** — praktik lazim, bukan aturan pedoman | C1, B1 |
| D11 | Penomoran pesan sequence `1`, `1.1`, `2` (nested) | **TIDAK DITEMUKAN SUMBER OTORITATIF — BUKAN ATURAN.** Aturan "level tertinggi = 1, level sama = prefiks sama" dalam sumber (Dharwiyanti & Wahono) berlaku untuk **Collaboration/Communication diagram**, BUKAN sequence. Menempelkannya ke sequence sebagai "aturan skripsi" adalah salah kutip | B1 |
| D12 | Guard ditulis `[kondisi]` | **DIDUKUNG LITERATUR untuk transisi statechart** (`event [guard]/action`); untuk sequence (combined fragment `alt`/`opt`) hanya didukung standar OMG UML 2.5, **TIDAK ditemukan pedoman kampus Indonesia yang mengaturnya — BUKAN ATURAN SKRIPSI** | B1, B6, B7 |
| D13 | Penulisan guard ganda `[[kondisi]]` | **TIDAK DITEMUKAN SUMBER APA PUN — BUKAN ATURAN.** Tidak ada sumber yang menulis guard dengan kurung siku ganda; ini artefak rendering tool, bukan notasi | — |
| D14 | Nama use case diawali kata kerja; nama aktor diawali kata benda | **DIDUKUNG BAHAN AJAR (Politeknik Telkom, kata "biasanya") — BUKAN pedoman kampus** | B5 |
| D15 | Nama kelas PascalCase / kata benda tunggal | **TIDAK DITEMUKAN SUMBER OTORITATIF (Indonesia) — BUKAN ATURAN.** Sumber Inggris (vendor tool) menyebut nama kelas biasanya huruf kapital + kata benda tunggal, tetapi itu bukan pedoman skripsi Indonesia | — |
| D16 | Class = 3 kompartemen (nama+stereotip, atribut, metoda); visibilitas Private/Protected/Public | **DIDUKUNG LITERATUR (Dharwiyanti & Wahono)** — konsisten dengan UML; bukan aturan pedoman kampus | B1 |
| D17 | Simbol visibilitas `+` Public, `-` Private, `#` Protected, `~` Package | **DIDUKUNG LITERATUR / PANDUAN TOOL** (Dharwiyanti untuk +/−/#; `~` dari panduan vendor). Bukan aturan pedoman kampus | B1 |
| D18 | Multiplicity `1`, `0..*`, `1..*`, `0..1` | **DIDUKUNG CONTOH SATU SKRIPSI** — praktik lazim UML, bukan aturan pedoman | C4 |
| D19 | Satu sequence diagram per alur (alur normal dan error dipisah) | **DIDUKUNG BAHAN AJAR (Henderi 2008)** — anjuran metodologis, bukan aturan pedoman | B4 |
| D20 | Urutan pengerjaan: business process → use case → activity → sequence → class | **DIDUKUNG BAHAN AJAR (Henderi 2008)** — urutan kerja yang dianjurkan, bukan aturan pedoman | B4 |
| D21 | Message pada sequence dipetakan menjadi operasi/metode class | **DIDUKUNG LITERATUR (Dharwiyanti & Wahono)** — prinsip traceability desain→kode | B1 |
| D22 | Satu set diagram universal untuk semua skripsi (mis. "wajib 7 diagram") | **TIDAK DITEMUKAN — BUKAN ATURAN.** Semua pedoman yang ditemukan menetapkan set secara kondisional dan berbeda-beda antar kampus | A1–A5 |
| D23 | Pedoman kampus selalu mengatur notasi UML | **SALAH.** Contoh: pedoman UI (A6) hanya mengatur format naskah, tidak mengatur notasi UML | A6 |

---

## E. Kesimpulan — laporan singkat

### Yang DIDUKUNG sumber otoritatif (pedoman resmi kampus)

1. **Konsistensi teori ↔ pembahasan UML** — wajib (UBSI).
2. **Set diagram kondisional, bukan universal** — 5 vs 4 diagram menurut
   pemakaian class (UBSI, STMIK Nuri); OOP vs terstruktur (STMIK Nusa Mandiri,
   STMIK Pelita Nusantara).
3. **Diagram harus terkait proses bisnis** (STMIK Nusa Mandiri).
4. **Caption "Gambar `<bab>.<urut>`", judul di bawah gambar, gambar + judul
   satu halaman tidak terpotong** (Universitas Ciputra; pola yang sama lazim di
   pedoman lain). Detail tipografi caption mengikuti kampus masing-masing.
5. **Tidak ada satu pun pedoman yang ditemukan menetapkan notasi UML baku
   versinya sendiri** — UBSI justru memerintahkan mahasiswa menyesuaikan simbol
   dengan referensi yang dipakai dan konsisten dengan referensi itu.

### Yang didukung literatur / bahan ajar / contoh — BOLEH dipakai sebagai praktik baik, TIDAK BOLEH diklaim sebagai "aturan skripsi"

- Aktor biasanya paling kiri; waktu sequence atas → bawah (Sholiq 2006;
  Dharwiyanti & Wahono 2003).
- Klasifikasi BCE (Jacobson) untuk objek sequence; boundary = antarmuka,
  control = logika alur, entity = data permanen (skripsi UNPAS; Dharwiyanti).
- Class 3 kompartemen; visibilitas `+`/`-`/`#`; multiplicity `1`, `0..*`,
  `1..*`, `0..1`.
- Nama use case diawali kata kerja, nama aktor diawali kata benda ("biasanya").
- Satu sequence per alur; alur normal dan error dipisah (Henderi 2008).
- Guard `[kondisi]` untuk transisi statechart; format `event [guard]/action`.
- Penomoran `1` / prefiks sama untuk **Collaboration/Communication diagram**.

### Yang TIDAK didukung — TIDAK BOLEH dijadikan aturan

- **Urutan lifeline WAJIB Actor → Boundary → Control → Entity** — tidak ada
  sumber otoritatif. (TIDAK DITEMUKAN SUMBER OTORITATIF — BUKAN ATURAN.)
- **Penomoran pesan sequence `1`, `1.1`, `2`** — aturan penomoran itu milik
  Collaboration diagram; tidak ada sumber yang menetapkannya untuk sequence.
  (TIDAK DITEMUKAN SUMBER OTORITATIF — BUKAN ATURAN; salah kutip bila
  diklaim untuk sequence.)
- **Guard `[[…]]` kurung ganda** — tidak ada sumber apa pun; artefak tool.
- **Nama kelas PascalCase / kata benda tunggal sebagai "aturan skripsi"** —
  tidak ditemukan di sumber Indonesia. (TIDAK DITEMUKAN SUMBER OTORITATIF —
  BUKAN ATURAN.)
- **Satu set diagram universal untuk semua skripsi** — semua pedoman yang
  ditemukan bersifat kondisional dan berbeda antar kampus.
- **Guard `[…]` pada sequence sebagai "aturan pedoman kampus"** — hanya
  didukung standar OMG UML 2.5 secara umum, bukan pedoman kampus Indonesia.

### Implikasi untuk diagram CBT SMK v2

- Jangan mengklaim satu konvensi UML sebagai "wajib skripsi" kecuali konvensi
  itu ada di pedoman kampus target (yang belum ditentukan dalam riset ini).
- Keputusan yang aman dan dapat dipertahankan di sidang: konsisten dengan SATU
  referensi yang dikutip (sesuai perintah UBSI: "sesuaikan dengan referensi
  yang digunakan"), caption "Gambar `<bab>.<urut>`" di bawah gambar, teori Bab
  II konsisten dengan diagram Bab III/IV, dan setiap diagram terkait proses
  bisnis sistem.
- Bila penguji mempersoalkan penomoran pesan sequence atau urutan lifeline BCE,
  jawaban yang jujur: itu pilihan penyajian mengikuti standar UML / literatur
  yang dikutip, bukan ketentuan pedoman kampus — karena memang tidak ada
  pedoman kampus yang mengaturnya (lihat D9, D11).

### Batasan riset ini (jujur)

- Pencarian dilakukan lewat cuplikan hasil pencarian web; sebagian pedoman
  hanya terjangkau lewat salinan cermin (123dok / Scribd / SlideShare /
  Academia.edu), bukan PDF resmi di domain kampus — statusnya dicatat di tiap
  sumber dan salinan cermin tidak diperlakukan setara dokumen resmi.
- Satu percobaan fetch langsung ke PDF skripsi UNPAS gagal timeout di sisi tool
  dan TIDAK diulangi; isinya dikutip dari cuplikan hasil pencarian.
- Tidak ditemukan pedoman resmi kampus yang mengatur notasi UML sampai level
  urutan lifeline, penomoran pesan sequence, atau penulisan guard pada
  sequence — ketiadaan ini adalah hasil riset, bukan kelalaian: kueri yang
  dipakai mencakup "penomoran pesan sequence diagram skripsi", "konvensi
  diagram UML tugas akhir", "pedoman UML universitas", dan "guard condition
  sequence diagram".
- File `docs/uml/cbt-smk-v2.mdj` tidak diubah.
