# R2 — Riset Jurnal Ilmiah: Kualitas Diagram UML / Sequence Diagram / Keterbacaan & Layout

**Proyek:** CBT SMK v2 — standardisasi diagram UML (StarUML `.mdj`)
**Tim:** R2 — Riset Jurnal Ilmiah Kualitas Diagram UML
**Tanggal:** 17 Oktober 2026
**Status:** 10 paper TERVERIFIKASI (judul + penulis + tahun + venue + DOI/URL cocok dengan hasil pencarian web).
**Batasan kerja:** File `docs/uml/cbt-smk-v2.mdj` TIDAK diubah. Tidak ada paper yang dikarang.

## Cara verifikasi & batasan bukti

- Setiap paper dicari lewat pencarian web; metadata (judul, penulis, tahun, venue, halaman, DOI) dicocokkan dari **minimal satu sumber independen** (halaman penerbit / repositori universitas / daftar referensi paper lain yang mengutipnya).
- Sebagian besar naskah lengkap (full text) berbayar / tidak dapat diakses gratis; **temuan kunci diambil dari abstrak resmi dan ringkasan penulis pengutip**, BUKAN dari naskah lengkap. Bila suatu detail hanya berasal dari ringkasan sekunder, itu ditandai eksplisit.
- Aturan turunan (R2-xx) HANYA memuat apa yang benar-benar dinyatakan paper tersebut. Klaim umum yang tidak didukung paper spesifik TIDAK dijadikan aturan — lihat bagian "Klaim yang SENGAJA TIDAK dijadikan aturan".
- Semua DOI di bawah ditulis persis seperti yang muncul di hasil pencarian. Bila DOI tidak berhasil diverifikasi, ditulis "DOI tidak terverifikasi".

---

## Paper terverifikasi

### P1 — Purchase (1997): estetika layout graf — persilangan adalah yang terpenting

| Field | Isi |
|---|---|
| Judul | Which aesthetic has the greatest effect on human understanding? |
| Penulis | Helen C. Purchase |
| Tahun | 1997 |
| Venue | Graph Drawing (Simposium Graph Drawing), Springer, Lecture Notes in Computer Science (bab buku prosiding) |
| DOI/URL | DOI tidak terverifikasi dari hasil pencarian. URL halaman penerbit: https://www.springerprofessional.de/which-aesthetic-has-the-greatest-effect-on-human-understanding/13633544 |
| Jenis bukti | Eksperimen daring (online), 5 estetika diuji, pemahaman diukur lewat waktu + jumlah kesalahan |

**Temuan kunci (dari abstrak resmi):**
- Mengurangi jumlah **persilangan garis (edge crossings)** adalah estetika **yang paling penting** — "by far the most important aesthetic".
- Meminimalkan **belokan (bends)** dan memaksimalkan **simetri** berpengaruh lebih kecil.
- Memaksimalkan **sudut minimum** antar garis yang keluar dari satu node dan menempelkan garis/node ke **grid ortogonal TIDAK signifikan** secara statistik.

**Catatan transfer (penting):** eksperimen ini memakai **graf abstrak**, bukan diagram UML. Menerapkannya ke UML adalah ekstrapolasi yang masuk akal, tetapi bukan bukti langsung pada UML. Jangan mengutip P1 sebagai "terbukti pada diagram UML" tanpa catatan ini.

**Aturan turunan:**
- **R2-01 (prioritas tertinggi):** Minimalkan persilangan garis. Cara cek: hitung persilangan pada tiap diagram; target 0. Bila persilangan tidak dapat dihindari, utamakan mengorbankan aspek lain (belokan/simetri) daripada menambah persilangan.
- **R2-02:** Minimalkan belokan garis. Cara cek: rute garis memakai segmen horizontal/vertikal dengan belokan sesedikit mungkin (rute L/Z sederhana); hindari garis berliku-liku.
- **R2-03 (aturan lunak):** Jaga keseimbangan/simetri tata letak (mis. penyebaran node merata, tidak menumpuk di satu sisi). Prioritas di bawah R2-01 dan R2-02. Cara cek: penilaian visual — bukan metrik keras.
- **R2-04 (larangan klaim palsu):** JANGAN mengklaim "sudah menempel grid ortogonal sehingga diagram bagus" — efek grid ortogonal dan sudut minimum terbukti TIDAK signifikan pada P1. Grid boleh dipakai sebagai alat bantu perataan, bukan sebagai bukti kualitas.

### P2 — Purchase, Colpoys, McGill, Carrington & Britton (2001): sintaks diagram kelas UML

| Field | Isi |
|---|---|
| Judul | UML class diagram syntax: an empirical study of comprehension |
| Penulis | Helen C. Purchase, Linda Colpoys, Matthew McGill, David Carrington, Carol Britton |
| Tahun | 2001 |
| Venue | Proceedings of the 2001 Asia-Pacific Symposium on Information Visualisation (APVis 2001), hlm. 113–120 |
| DOI/URL | https://dl.acm.org/citation.cfm?id=564054 (ID kutipan ACM 564054; DOI dalam bentuk 10.5555/… tidak terverifikasi sebagai DOI resmi) |

**Temuan kunci (dari abstrak):**
- Notasi sintaksis UML yang dipakai di buku, paper, dokumentasi, dan CASE tool **sering berbeda** padahal UML adalah standar.
- Eksperimen: subjek diminta menilai apakah spesifikasi yang diberikan cocok dengan diagram eksperimental, untuk 5 notasi yang masing-masing punya 2 varian setara secara semantik.
- Hasil: **notasi yang berkinerja terbaik dapat bergantung pada tugas** yang dikerjakan, dan prediksi intuitif penulis (berdasarkan kompleksitas notasi) **hanya terkonfirmasi sebagian**.

**Aturan turunan:**
- **R2-05:** Pakai SATU gaya notasi yang konsisten di semua diagram; jangan campur varian notasi setara (mis. dua cara menulis asosiasi/agregasi yang berbeda) dalam satu paket diagram. Cara cek: audit konsistensi simbol di semua diagram.
- **R2-06:** Jangan mengandalkan intuisi "notasi ini terasa lebih jelas" sebagai bukti. Cara cek: nilai diagram dengan tugas konkret (apakah pembaca dapat menjawab pertanyaan pemahaman dari diagram?), bukan dengan kesan sekilas.

### P3 — Purchase, Colpoys, McGill & Carrington (2002): sintaks diagram kolaborasi UML

| Field | Isi |
|---|---|
| Judul | UML collaboration diagram syntax: an empirical study of comprehension |
| Penulis | Helen C. Purchase, Linda Colpoys, Matthew McGill, David Carrington |
| Tahun | 2002 |
| Venue | Proceedings, First International Workshop on Visualizing Software for Understanding and Analysis (VISSOFT 2002), IEEE, hlm. 13–22 |
| DOI | 10.1109/VISSOF.2002.1019790 |
| URL verifikasi | http://eprints.gla.ac.uk/35825/ (repositori Universitas Glasgow; status: refereed = Yes) |

**Temuan kunci (dari abstrak):**
- Sama seperti P2, untuk diagram kolaborasi (kini disebut communication diagram): 5 notasi, masing-masing 2 varian setara semantik.
- Intuisi informal penulis **tervalidasi untuk tugas mengonfirmasi** bahwa spesifikasi cocok dengan diagram, tetapi **TIDAK terkonfirmasi untuk tugas mengidentifikasi kesalahan** dalam diagram.
- **Preferensi subjek berpihak pada varian notasi yang lebih ringkas (concise).**

**Aturan turunan:**
- **R2-07:** Bila dua varian notasi setara secara semantik, pilih yang lebih ringkas. Cara cek: tidak ada dekorasi/notasi bertele-tele yang tidak menambah makna.
- **R2-08:** Lakukan DUA jenis pengujian diagram secara terpisah: (a) "apakah diagram ini cocok dengan spesifikasi?" dan (b) "temukan kesalahan di diagram ini". Diagram yang lulus (a) belum terbukti mudah dipakai untuk (b). Cara cek: checklist review mencakup kedua jenis tugas.

### P4 — Otero & Dolado (2004): pemahaman pemodelan dinamis UML

| Field | Isi |
|---|---|
| Judul | Evaluation of the comprehension of the dynamic modeling in UML |
| Penulis | Mari Carmen Otero, José Javier Dolado |
| Tahun | 2004 |
| Venue | Information and Software Technology, vol. 46, no. 1, hlm. 35–53 |
| DOI | 10.1016/S0950-5849(03)00108-3 |

**Temuan kunci (dari ringkasan hasil eksperimen):**
- Eksperimen faktorial 3×3 dengan pengukuran berulang; 18 mahasiswa Informatika; 3 jenis notasi perilaku dinamis (sequence diagram, collaboration diagram, state diagram) × 3 domain aplikasi; metrik: total waktu + total skor.
- Kesimpulan utama: pemahaman pemodelan dinamis **bergantung pada jenis diagram dan pada kompleksitas dokumen**.
- Desain **lebih mudah dipahami bila perilaku dinamis dimodelkan dengan sequence diagram**; bila dimodelkan dengan collaboration diagram, pemahaman **menjadi kurang baik seiring domain/dokumen menjadi lebih kompleks**.

**Aturan turunan:**
- **R2-09:** Untuk perilaku dinamis (alur antar objek seiring waktu), utamakan **sequence diagram** daripada communication/collaboration diagram. Cara cek: setiap alur dinamis utama (mis. alur ujian, koreksi esai, token) dimodelkan sebagai sequence diagram.
- **R2-10:** Karena pemahaman menurun seiring kompleksitas, sequence diagram yang kompleks harus dipecah per skenario/alur (satu sequence = satu alur utama + alur alternatifnya yang terkait), bukan satu sequence raksasa untuk semua skenario. Cara cek: tiap sequence diagram punya satu judul alur yang jelas; tidak ada sequence yang mencampur banyak skenario tak terkait.

### P5 — Staron, Kuzniarz & Wohlin (2006): stereotip meningkatkan pemahaman model UML

| Field | Isi |
|---|---|
| Judul | Empirical assessment of using stereotypes to improve comprehension of UML models: A set of experiments |
| Penulis | Miroslaw Staron, Ludwik Kuzniarz, Claes Wohlin |
| Tahun | 2006 |
| Venue | Journal of Systems and Software, vol. 79, no. 5, hlm. 727–742 |
| DOI | 10.1016/j.jss.2005.09.014 |

**Temuan kunci (dari abstrak/ringkasan):**
- Serangkaian eksperimen terkontrol di akademia dan industri tentang peran stereotip terhadap pemahaman model UML.
- Hasil: stereotip **memainkan peran signifikan** dalam pemahaman model, dan peningkatan pemahaman dicapai **baik oleh mahasiswa maupun profesional industri**.

**Aturan turunan:**
- **R2-11:** Pakai stereotip secara konsisten untuk menambah informasi klasifikasi pada elemen model (mis. `«include»`, `«extend»` pada use case; `«boundary»`, `«control»`, `«entity»` pada lifeline/kelas). Cara cek: setiap stereotip yang muncul di diagram terdefinisi maknanya dan dipakai dengan makna yang sama di semua diagram.
- **R2-12:** Jangan memakai stereotip sebagai dekorasi tanpa makna (stereotip yang tidak menambah klasifikasi/informasi dilarang). Cara cek: hapus stereotip yang bila dihapus tidak mengubah pemahaman diagram.

### P6 — Moody (2009): "Physics of Notations" — teori 9 prinsip notasi visual

| Field | Isi |
|---|---|
| Judul | The "Physics" of Notations: Toward a Scientific Basis for Constructing Visual Notations in Software Engineering |
| Penulis | Daniel L. Moody |
| Tahun | 2009 |
| Venue | IEEE Transactions on Software Engineering, vol. 35, no. 6, hlm. 756–779 |
| DOI | 10.1109/TSE.2009.67 |

**Sifat bukti (penting):** P6 adalah **sintesis teori** berbasis literatur psikologi kognitif, persepsi visual, semiotika, dan komunikasi — BUKAN satu eksperimen tunggal atas diagram CBT. Kesembilan prinsipnya dirumuskan dari teori + bukti empiris yang dirujuk paper tersebut. Daftar 9 prinsip di bawah diverifikasi lewat ringkasan sekunder (arXiv / MDPI / CEUR) karena naskah lengkap tidak dapat diakses gratis.

**Sembilan prinsip (terverifikasi lewat ringkasan sekunder):**
1. *Semiotic clarity* — korespondensi 1:1 antara konstruk semantik dan simbol grafis.
2. *Perceptual discriminability* — simbol yang berbeda harus mudah dibedakan.
3. *Semantic transparency* — tampilan simbol menyarankan maknanya.
4. *Complexity management* — mekanisme eksplisit untuk menangani kompleksitas.
5. *Cognitive integration* — mekanisme eksplisit untuk mengintegrasikan informasi dari diagram yang berbeda.
6. *Visual expressiveness* — memakai seluruh rentang dan kapasitas variabel visual.
7. *Dual coding* — teks melengkapi grafik.
8. *Graphic economy* — jumlah simbol grafis yang berbeda harus mudah dikelola secara kognitif.
9. *Cognitive fit* — dialek visual berbeda untuk tugas dan audiens yang berbeda.

**Aturan turunan:**
- **R2-13 (semiotic clarity):** Satu simbol = satu makna. Dilarang: *symbol overload* (satu simbol untuk dua makna), *symbol redundancy* (dua simbol untuk satu makna), *symbol excess* (simbol tanpa makna semantik), *symbol deficit* (konsep penting tanpa simbol). Cara cek: audit kamus simbol — tiap jenis garis/simbol di diagram dipetakan ke tepat satu makna; contoh: garis putus-putus hanya untuk dependensi/`include`/`extend`, garis solid hanya untuk asosiasi.
- **R2-14 (perceptual discriminability):** Simbol berbeda harus dapat dibedakan lewat bentuk + label teks, BUKAN hanya lewat warna. Cara cek: diagram tetap dapat dibedakan maknanya bila dicetak hitam-putih; `«include»` vs `«extend»` dibedakan oleh teks stereotip + arah panah, bukan hanya warna.
- **R2-15 (semantic transparency):** Arah panah mengikuti arah pesan/alur yang sebenarnya; aktor manusia digambar sebagai aktor, bukan kotak. Cara cek: tidak ada panah yang arahnya berlawanan dengan arah pesan/alur yang dinarasikan.
- **R2-16 (complexity management):** Pecah diagram besar menjadi beberapa diagram terfokus (satu diagram = satu alur / satu sudut pandang). Cara cek: tiap diagram punya judul alur tunggal; tidak ada diagram "semua alur dalam satu gambar".
- **R2-17 (cognitive integration):** Sediakan mekanisme integrasi antar diagram: penamaan yang IDENTIK antar diagram + referensi silang eksplisit (mis. UC15 ↔ Sequence 3a ↔ Class terkait). Cara cek: nama use case / nama operasi sama persis di diagram Use Case, Sequence, dan Class; tidak ada nama yang hanya muncul di satu diagram tanpa penjelasan.
- **R2-18 (dual coding):** Setiap diagram disertai teks pelengkap: judul + narasi/keterangan singkat. Jangan mengandalkan grafik saja. Cara cek: tiap diagram punya judul dan minimal satu paragraf narasi / keterangan.
- **R2-19 (graphic economy):** Jaga jumlah JENIS simbol grafis yang berbeda seminimal mungkin. Ringkasan sekunder atas Moody menyebut angka panduan "enam" jenis simbol — angka ini DITANDAI sebagai berasal dari ringkasan sekunder, bukan dari verifikasi langsung naskah Moody. Cara cek: inventarisasi jenis simbol per diagram; bila ragam jenis simbol terlalu banyak, sederhanakan atau pecah diagram. Jangan mengutip "maksimal 6 simbol" sebagai angka mutlak dari Moody tanpa membaca naskah aslinya.
- **R2-20 (cognitive fit):** Sesuaikan kelengkapan dialek visual dengan audiens: diagram untuk penguji skripsi memakai label lengkap + legenda eksplisit, bukan singkatan internal. Cara cek: diagram paket skripsi memuat legenda dan label lengkap; tidak ada singkatan yang tidak dijelaskan.

### P7 — Nugroho (2009): tingkat detail (Level of Detail) dalam model UML dan dampaknya pada pemahaman

| Field | Isi |
|---|---|
| Judul | Level of detail in UML models and its impact on model comprehension: A controlled experiment |
| Penulis | Ariadi Nugroho |
| Tahun | 2009 |
| Venue | Information and Software Technology, vol. 51, no. 12, hlm. 1670–1685 |
| DOI | 10.1016/j.infsof.2009.04.007 |

**Temuan kunci:**
- Eksperimen terkontrol: **pengaruh Level of Detail (LoD) terhadap pemahaman model adalah signifikan** — model dengan LoD lebih tinggi meningkatkan kebenaran (correctness) dan efisiensi pemahaman.
- Replikasi kemudian (23 mahasiswa S2 Ilmu Komputer, Universitas Bina Nusantara, mereplikasi eksperimen awal dengan 53 mahasiswa S2 TU Eindhoven): LoD lebih tinggi konsisten meningkatkan **kebenaran** pemahaman; namun efek pada **efisiensi lebih lemah dan tidak signifikan** secara statistik pada replikasi (kemungkinan karena sampel lebih kecil dan perbedaan latar belakang subjek). Replikasi juga menemukan **kesenjangan persepsi vs kinerja**: subjek yang menerima model LoD rendah menilai modelnya tampak jelas/mudah, padahal pemahaman sebenarnya terganggu oleh informasi yang terbatas.

**Aturan turunan:**
- **R2-21:** Jangan menghilangkan detail kunci demi "kesederhanaan": cantumkan atribut/operasi kunci, label pesan lengkap (nama operasi + parameter penting bila relevan), dan multiplisitas pada relasi. Cara cek: tiap pesan sequence berlabel nama operasi yang bermakna; tiap relasi kelas yang penting berlabel multiplisitas.
- **R2-22:** Jangan menilai kualitas diagram dari kesan "terasa jelas/mudah" saja — persepsi kejelasan TIDAK mencerminkan pemahaman sebenarnya pada model ber-detail rendah. Cara cek: uji diagram dengan pertanyaan pemahaman konkret (dapatkah pembaca menjawab pertanyaan isi diagram dengan benar?), bukan dengan kesan.
- **R2-23 (batas klaim):** Jangan klaim "detail lebih tinggi selalu membuat diagram lebih cepat dibaca" — efek efisiensi tidak terkonfirmasi kuat pada replikasi. Klaim yang didukung: detail yang cukup meningkatkan KEBENARAN pemahaman.

### P8 — Störrle (2011): kualitas layout berpengaruh bila heuristik diterapkan bersamaan

| Field | Isi |
|---|---|
| Judul | On the impact of layout quality to understanding UML diagrams |
| Penulis | Harald Störrle |
| Tahun | 2011 |
| Venue | 2011 IEEE Symposium on Visual Languages and Human-Centric Computing (VL/HCC), hlm. 135–142 |
| DOI | 10.1109/VLHCC.2011.6070390 |

**Temuan kunci (dari abstrak):**
- Bukti eksperimental sebelumnya tentang efek layout lemah dan tidak konklusif. Paper ini menguji dua penjelasan: (1) kualitas visual lebih menonjol pada fase awal siklus hidup, dan (2) dalam praktik, layout yang baik memakai BANYAK heuristik secara bersamaan, sedangkan penelitian sebelumnya hanya menguji heuristik satu per satu (in isolation).
- Hasil: serangkaian eksperimen terkontrol memakai **compound layouts** (layout gabungan) pada model analisis kebutuhan menemukan **dampak nyata kualitas layout** terhadap pemahaman, dengan signifikansi sangat tinggi, diukur lewat berbagai aspek beban kognitif.

**Aturan turunan:**
- **R2-24:** Audit layout sebagai PAKET, bukan satu kriteria saja: persilangan + belokan + kesejajaran + jarak antar elemen + keterbacaan label dinilai bersamaan. Cara cek: checklist audit layout mencakup seluruh aspek sekaligus; dilarang menyatakan "layout sudah rapih" hanya karena satu aspek (mis. hanya persilangan nol) membaik sementara aspek lain memburuk.

### P9 — Störrle (2016): ukuran diagram dan pemahaman

| Field | Isi |
|---|---|
| Judul | On the impact of size to the understanding of UML diagrams |
| Penulis | Harald Störrle |
| Tahun | 2016 (terbit di jurnal; versi konferensi MODELS 2014 berjudul "On the Impact of Layout Quality to Understanding UML Diagrams: Size Matters", DOI 10.1007/978-3-319-11653-2_32) |
| Venue | Software & Systems Modeling, vol. 17, no. 1, hlm. 115–134 |
| DOI | 10.1007/s10270-016-0529-x |

**Temuan kunci (dari abstrak):**
- Tiga metrik ukuran diagram alternatif yang masuk akal ternyata sangat berkorelasi; dipilih yang paling sederhana.
- Ditemukan **korelasi negatif yang kuat antara ukuran diagram dan kinerja pemodel**; hasil sangat signifikan secara statistik dan memiliki validitas yang sangat besar.
- Besaran efek bervariasi, tetapi umumnya menunjukkan bahwa **dampak ukuran setara dengan atau melebihi faktor-faktor lain** dalam pemahaman diagram. Rekomendasi ukuran optimal diturunkan dari hasil ini dan diimplementasikan sebagai plug-in tool pemodelan.

**Aturan turunan:**
- **R2-25:** Batasi ukuran tiap diagram; pecah diagram besar menjadi beberapa diagram kecil yang terfokus. Cara cek: diagram yang "terasa besar" (banyak elemen/relasi) wajib dipecah; tidak ada diagram tunggal yang memuat seluruh sistem.
- **R2-26 (larangan mengarang angka):** Angka spesifik "ukuran optimal" (mis. jumlah elemen maksimum) **TIDAK dikutip** di dokumen ini karena tidak berhasil diverifikasi dari abstrak yang dapat diakses. Tim lain **DILARANG** mengarang angka (mis. "maksimal 20 elemen menurut Störrle") atas nama paper ini. Bila angka diperlukan untuk skripsi, baca naskah lengkap P9 dan kutip angka beserta definisi metrik ukurannya.

### P10 — Cruz-Lemus, Genero, Caivano, Abrahão, Insfrán & Carsí (2011): stereotip pada sequence diagram

| Field | Isi |
|---|---|
| Judul | Assessing the influence of stereotypes on the comprehension of UML sequence diagrams: A family of experiments |
| Penulis | José A. Cruz-Lemus, Marcela Genero, Danilo Caivano, Silvia Abrahão, Emilio Insfrán, José A. Carsí |
| Tahun | 2011 |
| Venue | Information and Software Technology, vol. 53, no. 12, hlm. 1391–1403 |
| DOI | 10.1016/j.infsof.2011.07.002 |

**Temuan kunci (dari abstrak):**
- Family of experiments: satu eksperimen + dua replikasi dengan masing-masing 78, 29, dan 36 mahasiswa S1 Ilmu Komputer. Pemahaman sequence diagram dengan dan tanpa stereotip dianalisis dari tiga perspektif Teori Kognitif Pembelajaran Multimedia (CTML): pemahaman semantik, retensi, dan transfer; plus meta-analisis untuk mengintegrasikan sampel.
- Hasil analisis statistik + meta-analisis: penggunaan stereotip yang diusulkan **membantu meningkatkan pemahaman diagram, terutama bila subjek tidak familiar dengan domain**.
- Kesimpulan: stereotip tersebut tampak membantu pemahaman sequence diagram yang lebih baik, **terutama pada domain yang belum dikenal baik**; penelitian lanjutan tetap diperlukan.

**Nuansa penting:** eksperimen tunggal pendahulunya (Genero et al., MODELS 2008, hlm. 280–294, DOI 10.1007/978-3-540-87875-9_20) hanya menunjukkan "kecenderungan lemah yang mendukung stereotip, namun tidak signifikan secara statistik" dan penulisnya sendiri menyatakan replikasi lebih lanjut diperlukan. Bukti kuat berasal dari family of experiments + meta-analisis 2011, bukan dari eksperimen tunggal 2008.

**Aturan turunan:**
- **R2-27:** Pada sequence diagram, beri stereotip pada pesan/lifeline kunci (mis. `«create»`, `«destroy»`, lifeline `«boundary»` / `«control»` / `«entity»`). Alasan kontekstual: pembaca skripsi (penguji) belum tentu familiar dengan domain CBT, dan justru pada kondisi itulah stereotip terbukti paling membantu. Cara cek: pesan-pesan kunci berstereotip; lifeline berstereotip peran.
- **R2-28:** Stereotip yang dipakai harus yang maknanya sudah didefinisikan (bukan stereotip karangan bebas tanpa definisi). Cara cek: daftar stereotip + definisinya tercantum di legenda/dokumen pendamping.

---

## Rekap aturan turunan (28 aturan, semua terikat bukti)

| ID | Aturan (ringkas) | Sumber |
|---|---|---|
| R2-01 | Minimalkan persilangan garis (prioritas tertinggi) | P1 |
| R2-02 | Minimalkan belokan garis | P1 |
| R2-03 | Jaga simetri/keseimbangan tata letak (lunak) | P1 |
| R2-04 | Jangan klaim grid ortogonal = bukti kualitas | P1 |
| R2-05 | Satu gaya notasi konsisten di semua diagram | P2 |
| R2-06 | Nilai diagram dengan tugas konkret, bukan intuisi | P2 |
| R2-07 | Pilih varian notasi yang lebih ringkas bila setara | P3 |
| R2-08 | Uji diagram dengan 2 tugas: konfirmasi kecocokan + temukan kesalahan | P3 |
| R2-09 | Perilaku dinamis → sequence diagram | P4 |
| R2-10 | Sequence kompleks → pecah per skenario | P4 |
| R2-11 | Pakai stereotip konsisten untuk klasifikasi | P5 |
| R2-12 | Larang stereotip dekoratif tanpa makna | P5 |
| R2-13 | Satu simbol = satu makna (1:1) | P6 |
| R2-14 | Bedakan simbol lewat bentuk + label, bukan hanya warna | P6 |
| R2-15 | Arah panah = arah pesan/alur sebenarnya | P6 |
| R2-16 | Pecah diagram besar menjadi diagram terfokus | P6 |
| R2-17 | Penamaan identik + referensi silang antar diagram | P6 |
| R2-18 | Tiap diagram: judul + narasi/keterangan | P6 |
| R2-19 | Jaga ragam jenis simbol seminimal mungkin (angka "6" hanya dari ringkasan sekunder) | P6 |
| R2-20 | Sesuaikan kelengkapan label dengan audiens (skripsi = label lengkap + legenda) | P6 |
| R2-21 | Jangan hilangkan detail kunci (label pesan, multiplisitas, atribut/operasi kunci) | P7 |
| R2-22 | Jangan nilai kualitas dari kesan "terasa jelas" | P7 |
| R2-23 | Jangan klaim detail tinggi selalu lebih cepat dibaca | P7 |
| R2-24 | Audit layout sebagai paket, bukan satu kriteria | P8 |
| R2-25 | Batasi ukuran tiap diagram; pecah yang besar | P9 |
| R2-26 | Larangan mengarang angka "ukuran optimal" atas nama Störrle | P9 |
| R2-27 | Stereotip pada pesan/lifeline kunci sequence diagram | P10 |
| R2-28 | Stereotip harus terdefinisi, bukan karangan bebas | P10 |

---

## Klaim yang SENGAJA TIDAK dijadikan aturan (tidak didukung paper yang diverifikasi)

Daftar ini ditulis eksplisit agar tim lain tidak mengutipnya sebagai "temuan jurnal":

1. **"Batasi jumlah lifeline" dengan angka tertentu** — tidak ada paper terverifikasi di atas yang menyatakan batas angka jumlah lifeline. Yang didukung adalah prinsip umum: kurangi kompleksitas / pecah diagram (P4, P9), tanpa angka.
2. **"Label pesan HARUS memakai nama operasi"** — tidak ada paper di atas yang menyatakan kalimat persis ini. R2-21 hanya menuntut label pesan yang lengkap dan bermakna (berbasis temuan LoD, P7), bukan format penamaan tertentu.
3. **"Diagram harus muat satu halaman"** — tidak ada paper di atas yang menyatakan kalimat persis ini. Yang didukung: diagram yang lebih kecil lebih mudah dipahami (P9) dan diagram besar harus dipecah (P6, P9) — tanpa patokan "satu halaman".
4. **"Garis diagonal/curved dilarang"** — P1 hanya menyatakan efek grid ortogonal tidak signifikan dan bends berpengaruh lebih kecil; larangan mutlak garis diagonal berasal dari buku panduan praktisi (Ambler, lihat bawah), BUKAN dari paper empiris yang diverifikasi. R2-02 memakai rumusan yang didukung P1: minimalkan belokan.
5. **"Maksimal N elemen per diagram" (angka apa pun)** — tidak ada angka yang dikutip (lihat R2-26).

---

## Kandidat yang ditelusuri tetapi TIDAK dihitung sebagai paper terverifikasi

| Kandidat | Alasan tidak dihitung |
|---|---|
| Scott W. Ambler, *The Elements of UML 2.0 Style* (Cambridge University Press, 2005) — pedoman: hindari garis bersilangan, gambarkan persilangan sebagai "lompatan", hindari garis diagonal/lengkung, simbol berukuran konsisten, tempelkan garis ke tengah simbol, ratakan label horizontal | **Buku panduan praktisi, bukan paper jurnal peer-reviewed.** Isinya sejalan dengan P1 (persilangan, belokan), tetapi tidak boleh dikutip sebagai bukti empiris jurnal. Boleh dipakai sebagai pendukung gaya, dengan label "buku panduan, bukan jurnal". |
| "Definition of the Criteria for Layout of the UML Use Case Diagrams" (ditemukan di Academia.edu) | Venue / status peer-review **tidak berhasil diverifikasi** dari hasil pencarian. Tidak dipakai sebagai bukti. |
| Glezer et al. (2005), "Quality and comprehension of UML interaction diagrams — an experimental comparison", *Information and Software Technology* 47(10):675–692 | Hanya ditemukan **sitasinya** di daftar referensi paper lain; temuan kuncinya **tidak berhasil diambil** dari hasil pencarian. Tidak dipakai sebagai bukti. |
| Gemino & Parker (2009), "Use case diagrams in support of use case modeling", *Journal of Database Management* 20(1):1–24 | Hanya ditemukan sitasinya; temuan kuncinya tidak berhasil diambil. Tidak dipakai sebagai bukti. |
| Purchase et al. (2003), "UML Class Diagrams: An Empirical Study of Comprehension", dalam *Software Visualization*, Springer, DOI 10.1007/978-1-4615-0457-3_6 | Bab buku ini memuat eksperimen notasi + eksperimen layout atas diagram kelas UML — relevan, tetapi **temuan spesifik eksperimen layout-nya tidak berhasil diambil** dari hasil pencarian (hanya abstrak umum). Agar tidak mengarang, bab ini tidak dihitung sebagai paper ke-11; tim dapat menambahkannya bila naskahnya dibaca langsung. |

---

## Metode pencarian

Kueri pencarian web (Bahasa Inggris) yang dipakai, antara lain: "sequence diagram quality empirical study", "UML diagram layout quality metrics", "UML diagram comprehension empirical study", "use case diagram quality guidelines", "Purchase UML class diagram syntax empirical study comprehension DOI", "Moody Physics of Notations IEEE Transactions Software Engineering 2009 DOI", "Storrle impact layout quality understanding UML diagrams", "Nugroho Chaudron level of detail UML diagram comprehension empirical", "Otero Dolado Evaluation comprehension dynamic modeling UML DOI", "Cruz-Lemus Genero stereotypes comprehension UML sequence diagrams family experiments", "Moody physics notations nine principles list". Hasil pencarian yang tidak relevan (blog vendor, artikel SEO, situs bimbingan belajar) dibuang dan tidak dikutip.

## Rekomendasi untuk tim standardisasi (R1/R3/R4/R5)

1. Pakai tabel rekap R2-01 s.d. R2-28 sebagai **checklist audit** — tiap aturan punya cara cek yang dapat diobservasi pada file `.mdj`/ekspor PNG.
2. Untuk kutipan skripsi: kutip P1–P10 dari metadata di atas; bila penguji meminta angka spesifik (mis. ukuran optimal Störrle, angka "6" Moody), baca naskah lengkapnya terlebih dahulu — dokumen ini sengaja tidak mengarang angka tersebut.
3. Jangan menambahkan aturan baru "atas nama jurnal" tanpa menambahkan paper terverifikasi baru ke dokumen ini dengan format yang sama (judul + penulis + tahun + venue + DOI/URL + temuan + aturan).

— Akhir dokumen R2 —
