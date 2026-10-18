# LAPORAN TIM PEMBANDING LUAR — Mekanisme Aplikasi Ujian / Kuesioner Lain vs Extraordinary CBT vs CBT SMK v2

**Tanggal:** 16 Oktober 2026
**Penyusun:** Tim Pembanding Luar (12 orang, kerja mandiri — tanpa subagent)
**Status:** HANYA ANALISIS. Tidak ada kode yang diubah.

## 0. Cara membaca laporan ini (aturan bukti)

Setiap klaim memakai salah satu label:

- **TERBUKTI** — ada di dokumentasi / halaman resmi / dokumen publik yang disebut sumbernya. Nama sumber + URL dicantumkan.
- **DISIMPULKAN** — penalaran yang masuk akal dari bukti yang dikutip, tetapi tidak dinyatakan eksplisit oleh sumber. Ditandai sebagai simpulan, bukan fakta vendor.
- **TIDAK DIKETAHUI** — tidak ditemukan bukti publik yang cukup dalam riset ini. Tidak ditebak.

Aturan tambahan yang dipatuhi:

- Untuk ANBK: hanya dari dokumen publik (POS / juknis / penjelasan Pusmendik-BSKAP dan turunan yang mengutipnya). Klaim yang tidak ada di dokumen publik ditulis TIDAK DIKETAHUI, bukan dikarang.
- Dokumen ini boleh MENYEBUT nama aplikasi lain untuk perbandingan faktual. Tidak ada aset, logo, branding, atau teks berhak cipta pihak lain yang disalin ke sini.
- Kolom Extraordinary CBT (Exo) dan CBT SMK v2 diisi HANYA dari dua bahan baca yang ditugaskan:
  - `~/workspace/cbt-smk-v2/ANALISIS.md` (1670 baris, 16 Okt 2026)
  - `~/workspace/extraordinary-cbt/PELAJARI-EXTRAORDINARY-CBT.md` (kajian Exo v4.7.0)
- Penting untuk kolom Exo: sesuai ANALISIS.md Bab 13, fitur Exo disimpulkan dari daftar menu/permission (100+) dan skema dump 38 tabel, **bukan dari uji live** — instalasi kajian terhenti di gerbang validasi lisensi (signature offline tidak cocok, cek online timeout, percobaan dihentikan karena risiko IP-block). Jadi status "ada di Exo" di laporan ini artinya "terdapat di menu/skema menurut kajian", bukan "terverifikasi jalan".
- Penting untuk kolom v2: ANALISIS.md adalah dokumen **rencana/spesifikasi** (status saat ditulis: "FASE ANALISIS + RISET SAJA. Belum ada kode aplikasi yang ditulis"). Beberapa butir masih berlabel asumsi / belum diputuskan (Bab 23–24 ANALISIS.md). Laporan ini tidak mengubah status itu menjadi "sudah jadi".

### Keterbatasan riset ini

- Riset memakai pencarian web publik + halaman hasil pencarian. Beberapa halaman resmi tidak dapat dibuka penuh lewat alat baca teks (mis. halaman harga dinamis), sehingga harga untuk beberapa vendor ditulis dari agregator/pembanding pihak ketiga dan diberi label yang sesuai — bukan diklaim sebagai harga resmi mutlak.
- Harga SaaS berubah sewaktu-waktu dan sering berbeda per negara/mata uang. Semua angka harga di laporan ini harus dicek ulang ke halaman harga resmi vendor sebelum dipakai untuk keputusan pembelian.
- Fitur yang hanya muncul di ulasan pengguna / agregator (bukan di dokumentasi resmi) diberi label DISIMPULKAN atau TIDAK DIKETAHUI, bukan TERBUKTI.

---

## 1. Ringkasan eksekutif

| Aplikasi | Posisi inti (DISIMPULKAN dari bukti di bawah) | Cocok untuk ujian sumatif sekolah? |
|---|---|---|
| Moodle Quiz | LMS open-source; kuis adalah satu aktivitas di dalam course. Paling lengkap untuk bank soal + statistik butir di antara pembanding gratis. | Ya — dengan catatan butuh server + admin teknis. |
| Google Forms (+ mode kuis) | Formulir/kuesioner gratis; mode kuis menambah kunci + poin. Bukan sistem ujian. | Tidak untuk ujian sumatif resmi. Cocok untuk kuis/latihan/survei. |
| Exam.net | Platform ujian digital cloud; fokus keamanan (3 level, termasuk Safe Exam Browser) + monitoring real-time + auto-mark. | Ya — untuk ujian yang butuh pengawasan kuat, dengan biaya langganan. |
| Quizizz (kini Wayground) | Kuis gamifikasi live / tugas mandiri; 20+ tipe soal (klaim halaman harga), laporan standar, anti-cheating di paket berbayar. | Sebagian — kuat untuk formatif/latihan; bukan pengganti sistem ujian sumatif + UKK + berita acara. |
| Kahoot! | Game kuis live; poin berbasis kecepatan; peserta terbatas per paket. | Tidak untuk ujian sumatif. Cocok untuk review/energi kelas. |
| ProProfs Quiz Maker | Pembuat kuis SaaS; bank soal + pooling + acak + laporan; kontrol keamanan dasar. | Sebagian — untuk tes/kuis; bukan sistem ujian sekolah lengkap (tidak ada UKK, berita acara, ruangan). |
| ANBK | Sistem asesmen nasional (Pusmendik/BSKAP): AKM + Survei Karakter + Survei Lingkungan Belajar; hasil agregat sekolah, bukan nilai individu / bukan penentu kelulusan. | TIDAK DAPAT dipakai sebagai pengganti ujian sekolah. Ini sistem nasional terpisah. |
| Extraordinary CBT v4.7.0 | Binary Go + PostgreSQL; set fitur operasional proktor paling lengkap di antara pembanding (pause/resume, device lock, auto-block, event+ruangan+kartu, analisis butir, nilai per kompetensi, AI opsional). | Ya untuk ujian sekolah — dengan catatan: binary tertutup + gerbang lisensi, TIDAK ada modul UKK. |
| CBT SMK v2 (rencana) | Rebuild Go + React + PostgreSQL untuk SMK: US (PTS/PAS/USBK) + UKK + tryout, server lokal sekolah (LAN/semi-daring), 6 role, paket ujian + kisi-kisi + KKTP + analisis butir + nilai per kompetensi. | Dirancang untuk itu — status: spesifikasi, bukan klaim sudah jadi. |

Kesimpulan utama untuk v2 (DISIMPULKAN):

1. **Tidak ada satu pun pembanding luar yang menggabungkan semua yang dibutuhkan SMK**: ujian sumatif terjadwal + UKK teori/praktik + berita acara + kartu peserta + analisis butir + nilai per kompetensi + operasi LAN/semi-daring. Moodle paling dekat dari sisi bank soal/statistik, Exo paling dekat dari sisi operasional proktor, ANBK paling dekat dari sisi pola token/proktor — tetapi ketiganya masing-masing punya celah yang justru menjadi alasan v2 ada.
2. **Mekanisme yang layak diadopsi v2** (sudah ada di ANALISIS.md, dikonfirmasi pembanding): bank soal + pooling/acak berstrata (Moodle, ProProfs), statistik butir + analisis pengecoh (Moodle), kunci tidak dikirim ke browser (prinsip umum; juga dinyatakan eksplisit untuk satu platform lain), token yang dirilis proktor menjelang mulai (ANBK, Exo), monitor real-time + reset login + force finish (Exam.net, Exo, ANBK), nilai per kompetensi (Exo), pause yang membekukan timer (asumsi v2 — perlu dicatat: pause CBT-Go TIDAK membekukan timer).
3. **Mekanisme yang TIDAK boleh ditiru mentah-mentah**: poin berbasis kecepatan (Kahoot/Quizizz) untuk ujian sumatif — tidak adil untuk soal berpikir; timer per soal yang kaku untuk esai; hasil ANBK dipakai sebagai nilai individu (salah secara konsep); mengandalkan browser biasa tanpa pengawas untuk ujian berisiko tinggi.

---

## 2. Moodle Quiz

**Sumber utama:** dokumentasi dan blog resmi Moodle (moodle.com, docs.moodle.org), panduan institusi, dan agregator yang mengutip pengaturan kuis.

### 2.1 Alur ujian

- **TERBUKTI:** Kuis adalah satu aktivitas di dalam course. Guru membuat kuis (nama, deskripsi), lalu menambah soal dari bank soal; soal disimpan terpisah dari kuis sehingga bisa dipakai ulang di kuis lain. Sumber: blog resmi Moodle "Test your learners' knowledge using the quiz feature!" (https://moodle.com/news/test-learners-knowledge-using-quiz-feature/) dan panduan institusi (mis. York University, University of Illinois).
- **TERBUKTI:** Urutan soal bisa diacak ("shuffled randomly") dan jumlah soal per halaman dapat diatur. Sumber: panduan instruktur (https://osgoode.yorku.ca/wp-content/uploads/2014/09/Creating-Quizzes-in-your-course.pdf) dan panduan institusi lain.
- **TERBUKTI:** Siswa dapat melihat pratinjau (preview) sebelum kuis dirilis; ada opsi navigasi sekuensial (tidak bisa kembali ke soal sebelumnya) dan opsi jumlah attempt. Sumber: panduan institusi yang sama.
- **DISIMPULKAN:** Alur ini adalah alur LMS generik (course → aktivitas → attempt → nilai), bukan alur ujian sekolah terjadwal (jadwal + sesi + ruangan + token + berita acara). Tidak ada entitas jadwal/ruangan/berita acara di dalam aktivitas Quiz itu sendiri.

### 2.2 Token / timer

- **TERBUKTI:** Ada jendela buka/tutup (open/close date) dan batas waktu (time limit, dalam menit); setelah waktu habis, attempt dapat di-submit otomatis. Sumber: panduan institusi (Swarthmore KB, https://swatkb.atlassian.net/wiki/spaces/MOODLE/pages/20906321/Creating+a+timed+quiz+or+exam+in+Moodle).
- **TERBUKTI:** Ada password kuis (quiz password) dan pembatasan alamat jaringan (network address restriction) sebagai "extra restrictions on attempts". Sumber: ringkasan pengaturan kuis institusi + artikel sekunder yang mengutip pengaturan Moodle.
- **TIDAK DIKETAHUI:** Apakah ada token 6-karakter per sesi seperti Exo/ANBK/v2. Yang TERBUKTI adalah password kuis + batasan IP, bukan token sesi yang dirilis proktor. Jangan samakan keduanya.

### 2.3 Anti-curang

- **TERBUKTI:** Pengaturan bawaan yang membatasi kecurangan: soal acak dari bank soal yang lebih besar, opsi jawaban diacak, batas waktu + jendela buka/tutup, navigasi sekuensial, attempt terbatas + jeda antar attempt, password kuis / batasan alamat jaringan, dan kontrol kapan siswa boleh melihat jawaban benar (review options). Sumber: artikel sekunder yang merangkum pengaturan Moodle (https://acemyhomework.com/blog/can-moodle-detect-cheating).
- **TERBUKTI:** Moodle mencatat log aktivitas (waktu, pengguna, IP), detail attempt (mulai/selesai, durasi, skor), dan riwayat respons per soal (kapan dilihat, kapan jawaban disimpan/diubah). Sumber: artikel yang sama.
- **DISIMPULKAN:** Moodle mengandalkan kombinasi pengaturan + log + pengawasan manusia, bukan kunci perangkat keras. Tidak ada klaim resmi di sumber yang dibaca bahwa Moodle bawaan memblokir tab kedua / mengunci perangkat. Safe Exam Browser adalah integrasi terpisah (bukan bawaan Moodle) — TIDAK DIKETAHUI dari riset ini apakah konfigurasi SEB tersedia di semua instalasi Moodle; jangan diklaim sebagai fitur bawaan.

### 2.4 Penilaian

- **TERBUKTI:** Sebagian tipe soal dinilai otomatis; esai harus dinilai manual. Kuis bernilai 100 poin secara default tetapi dapat diubah; metode agregasi nilai antar attempt dapat dipilih (mis. nilai tertinggi). Sumber: panduan instruktur York University.
- **TERBUKTI:** Ada umpan balik keseluruhan (overall feedback) berdasarkan skor akhir dan umpan balik per soal. Sumber: panduan yang sama.
- **TIDAK DIKETAHUI:** Apakah Moodle bawaan menghitung nilai per kompetensi (CP/TP) seperti Exo/v2. Yang TERBUKTI adalah gradebook + agregasi nilai kuis; pemetaan kompetensi adalah fitur terpisah (competency framework) di Moodle dan tidak diteliti mendalam di sini.

### 2.5 Analisis soal

- **TERBUKTI:** Laporan Statistics kuis menampilkan informasi kuis (rata-rata, jumlah attempt selesai, distribusi skor, simpangan baku, skewness, error ratio), analisis struktur kuis per soal (discrimination index), dan statistik per soal termasuk frekuensi pemilihan tiap jawaban / pengecoh. Sumber: blog resmi Moodle "Quiz Analytics Uncovered" (https://moodle.com/us/news/quiz-analytics-uncovered/).
- **DISIMPULKAN:** Ini adalah padanan terdekat analisis butir soal Exo/v2 di antara pembanding luar. Perbedaannya: di Moodle, analisis ini berada di dalam laporan kuis per course; di Exo/v2, analisis butir terikat ke bank soal + jadwal dan dipakai untuk status "Perlu Revisi".

### 2.6 Import soal

- **TERBUKTI:** Moodle mendukung import soal dari bank soal (question bank → import) dalam format GIFT dan Moodle XML. Sumber: panduan import GIFT (https://bpb-us-e1.wpmucdn.com/u.cotc.edu/dist/3/152/files/2020/03/GIFT-Format.pdf) dan artikel impor soal (https://arxiv.org/pdf/2010.15577).
- **TIDAK DIKETAHUI:** Apakah import Word (.docx) langsung didukung bawaan. Yang TERBUKTI adalah GIFT/XML; konversi Word → GIFT biasanya lewat alat bantu pihak ketiga. Jangan klaim import .docx bawaan Moodle.

### 2.7 Harga / model bisnis

- **TERBUKTI:** Moodle adalah LMS open-source; tidak ada biaya lisensi — yang dibutuhkan adalah hosting + dukungan sendiri, atau bantuan dari penyedia layanan bersertifikat. Sumber: halaman resmi Moodle (https://moodle.com/us/about/open-source/).
- **TIDAK DIKETAHUI:** Harga pasti MoodleCloud / hosting terkelola saat ini (halaman harga dinamis, tidak terverifikasi dalam riset ini). Jangan pakai angka lama (mis. batas MoodleCloud lama) sebagai harga saat ini.

### 2.8 Celah untuk konteks SMK (DISIMPULKAN)

Moodle tidak menyediakan, sebagai satu alur bawaan: jadwal + sesi + ruangan + token per sesi, berita acara, kartu peserta, modul UKK teori+praktik + rubrik asesor, dan nilai per kompetensi per soal yang terikat kisi-kisi Kurikulum Merdeka. Semua itu harus dibangun/dikonfigurasi di atas Moodle atau di luar Moodle.

---

## 3. Google Forms (+ mode kuis)

**Sumber utama:** halaman Google Workspace / Google for Education, panduan pihak ketiga yang menjelaskan mode kuis, dan dokumentasi add-on timer (pihak ketiga).

### 3.1 Alur ujian

- **TERBUKTI:** Google Forms adalah pembuat formulir. Alurnya: buat form → tambah pertanyaan → bagikan lewat link/email/embed → kumpulkan respons di tab Responses / Google Sheets. Sumber: panduan umum Google Forms (mis. https://gadgets360.com/how-to/features/google-form-create-link-free-surveys-quiz-questionnaire-2299138).
- **TERBUKTI:** Mode kuis diaktifkan lewat Settings → "Make this a quiz". Setelah itu pembuat dapat memilih kapan nilai dirilis (langsung setelah submit / nanti setelah review manual) dan apa yang dilihat responden (soal yang salah, jawaban benar, nilai poin). Sumber: panduan yang sama + panduan Formfacade (https://chatfacade.com/calculate/create-quiz-google-forms-calculate-score.html).
- **DISIMPULKAN:** Tidak ada konsep jadwal + sesi + ruangan + proktor + token. "Ujian" di Google Forms adalah satu form yang dibuka/ditutup secara manual.

### 3.2 Token / timer

- **TERBUKTI:** Google Forms bawaan TIDAK memiliki timer per peserta / timer hitung mundur bawaan. Yang ada adalah pembatasan kustom lewat add-on pihak ketiga (mis. Form Presenter+Timer, Boloforms) yang menambah timer + kontrol buka/tutup + jeda. Sumber: halaman Marketplace add-on (https://workspace.google.com/marketplace/app/form_presenter+timer/93519377337 dan https://workspace.google.com/marketplace/app/timer_+_proctor_for_google_forms_quiz_te/1083655933830).
- **TERBUKTI:** Tidak ada token ujian bawaan. Kontrol akses bawaan adalah: batasi ke pengguna dalam organisasi / kumpulkan email, dan tutup form secara manual. Timer add-on pihak ketiga adalah layanan berbayar (Form Presenter menyebut 20 tes pertama gratis sebagai trial). Sumber: halaman Marketplace yang sama.
- **DISIMPULKAN:** Mengandalkan Google Forms + add-on timer untuk ujian sumatif berarti menggantungkan integritas ujian pada pihak ketiga, bukan pada sistem itu sendiri.

### 3.3 Anti-curang

- **TERBUKTI:** Fitur bawaan yang relevan: acak urutan pertanyaan (shuffle question order), batasi 1 respons, kumpulkan email, batasi ke domain. Sumber: panduan umum Google Forms + halaman add-on Quiz Helper (https://workspace.google.com/marketplace/app/quiz_helper_by_alice_keeler/581441704194).
- **TIDAK DIKETAHUI:** Apakah ada deteksi pindah tab / auto-block / device lock bawaan. Tidak ditemukan bukti bawaan; add-on pihak ketiga mengklaim fitur proctoring (mis. foto live dari webcam) tetapi itu bukan fitur Google Forms bawaan dan tidak diverifikasi live di sini.
- **DISIMPULKAN:** Google Forms tidak cocok untuk ujian berisiko tinggi karena tidak ada kunci jawaban yang dilindungi secara arsitektural seperti sistem ujian (kunci form dapat dilihat pembuat form; tidak ada pemisahan peran proktor/pengawas/guru).

### 3.4 Penilaian

- **TERBUKTI:** Setiap pertanyaan dapat diberi kunci jawaban + poin lewat "Answer key"; ada umpan balik jawaban benar/salah. Nilai total = jumlah poin soal yang benar. Sumber: panduan Formfacade + Gadgets360.
- **TERBUKTI:** Keterbatasan bawaan (menurut Formfacade): tidak dapat memberi poin berbeda per pilihan jawaban dalam satu soal pilihan ganda; tidak dapat menyebar poin antar jawaban checkbox; tidak dapat menghitung skor per seksi; tidak dapat memberi poin desimal (mis. 0,5). Sumber: https://chatfacade.com/calculate/create-quiz-google-forms-calculate-score.html.
- **TIDAK DIKETAHUI:** Apakah ada koreksi manual esai terstruktur + rubrik bawaan. Yang TERBUKTI adalah review manual nilai (release grade later after manual review) — bukan rubrik.

### 3.5 Analisis soal

- **TERBUKTI:** Tab Responses menampilkan ringkasan per pertanyaan (distribusi jawaban) dan spreadsheet respons mentah. Sumber: panduan umum Google Forms.
- **TIDAK DIKETAHUI:** Daya beda (discrimination), analisis pengecoh formal, tingkat kesukaran P. Tidak ditemukan bukti fitur tersebut di Google Forms bawaan. Jangan klaim ada.

### 3.6 Import soal

- **TERBUKTI:** Google Forms bawaan dapat mengimpor pertanyaan dari form lain (menu Import → pilih form → pilih pertanyaan). Sumber: https://www.guidingtech.com/import-questions-into-google-forms/.
- **TERBUKTI:** Tidak ada import CSV/Word bawaan untuk soal. Import dari dokumen lain memerlukan add-on pihak ketiga. Sumber: panduan yang sama (yang menekankan import antar-form sebagai cara bawaan).
- **DISIMPULKAN:** Untuk bank soal besar (ratusan soal), Google Forms tidak efisien karena tidak ada import massal bawaan.

### 3.7 Harga / model bisnis

- **TERBUKTI:** Google Forms adalah bagian dari Google Workspace; ada opsi personal tanpa biaya dan paket bisnis berbayar. Untuk pendidikan: Google Workspace for Education Fundamentals tanpa biaya untuk institusi yang memenuhi syarat; edisi Standard/Plus berbayar. Sumber: https://ipv4.google.com/forms/about/?hl=en_uk dan https://edu.google.com/intl/ALL_au/workspace-for-education/editions/education-fundamentals/ serta FAQ Workspace for Education (https://knowledge.workspace.google.com/admin/getting-started/editions/google-workspace-for-education-faq).
- **TIDAK DIKETAHUI:** Harga pasti paket bisnis saat ini untuk Indonesia (halaman harga dinamis per wilayah). Jangan pakai angka lama.

### 3.8 Celah untuk konteks SMK (DISIMPULKAN)

Google Forms tidak punya: jadwal/sesi/ruangan, token, timer bawaan, anti-curang ujian, koreksi esai rubrik, analisis butir, nilai per kompetensi, UKK, berita acara, kartu peserta, audit log. Cocok untuk kuis cepat / survei / latihan, bukan ujian sumatif.

---

## 4. Exam.net

**Sumber utama:** blog resmi Exam.net (SEB), panduan trainee pihak ketiga, dan agregator (Capterra / OnlineExamMaker).

### 4.1 Alur ujian

- **TERBUKTI:** Guru membuat ujian dan membagikannya lewat exam key (kunci ujian). Siswa memasukkan exam key, mengisi identitas (nama, kelas, ID siswa), lalu klik Begin Exam. Sumber: panduan trainee (https://jobsplus.gov.mt/media/j0cd3szr/tcqa-401-online-assessment-guidelines-for-trainees-rev-02-21012022.pdf).
- **TERBUKTI:** Ada dua pola tampilan: (a) ujian pilihan ganda penuh — siswa menandai jawaban; (b) ujian campuran — naskah soal di atas (tidak dapat diedit) dan area tulis jawaban (word pad) di bawah. Sumber: panduan trainee yang sama.
- **DISIMPULKAN:** Pola "naskah + area tulis" ini mirip ujian kertas yang didigitalkan, berbeda dari pola v2 (satu soal per layar + autosave per soal + navigasi soal).

### 4.2 Token / timer

- **TERBUKTI:** Akses memakai exam key (bukan login akun siswa permanen dalam alur yang didokumentasikan panduan trainee). Sumber: panduan trainee yang sama.
- **TERBUKTI:** Ada batas waktu (time limits) sebagai fitur platform. Sumber: daftar fitur Capterra untuk Exam.net (https://www.capterra.in/compare/1015788/1063218/examnet/vs/doexam) yang mencantumkan Time Limits, Scheduling, Monitoring.
- **TIDAK DIKETAHUI:** Panjang/format exam key, masa berlaku (expiry), dan apakah ada token per sesi yang dirilis proktor seperti ANBK/Exo/v2. Tidak ditemukan bukti eksplisit dalam sumber yang dibaca.

### 4.3 Anti-curang

- **TERBUKTI:** Exam.net menawarkan tiga level keamanan; level tertinggi untuk pengguna Windows memakai Safe Exam Browser (SEB), yang mengunci layar komputer dan membatasi akses ke selain konten ujian; untuk iOS direkomendasikan aplikasi native Exam.net. Sumber: blog resmi Exam.net (https://exam.net/blog/ensuring-academic-integrity-with-safe-exam-browser-seb-and-exam-net).
- **TERBUKTI:** Exam.net mengklaim proctoring jarak jauh + monitoring real-time + alert cerdas + auto-mark. Sumber: ringkasan OnlineExamMaker (https://onlineexammaker.com/kb/10-best-anti-cheating-online-test-makers-for-teachers/) yang menyebut Exam.net: paperless e-assessment, efficient online proctoring, Auto-Mark, intelligent real-time monitoring & alerts.
- **TERBUKTI:** Dalam panduan trainee: jika siswa menekan Esc (keluar dari asesmen), siswa diminta memberi alasan sebelum diizinkan kembali. Sumber: panduan trainee JobsPlus.
- **DISIMPULKAN:** Exam.net adalah pembanding terkuat dari sisi keamanan browser (SEB). Kelemahannya untuk SMK: SEB memerlukan instalasi + konfigurasi di tiap komputer; tidak semua lab sekolah siap. v2 sengaja tidak mengandalkan SEB sebagai satu-satunya pertahanan (ANALISIS.md Bab 7: fullscreen web hanya deteksi, Esc selalu bisa keluar).

### 4.4 Penilaian

- **TERBUKTI:** Ada Auto-Mark (penilaian otomatis) untuk soal objektif; soal uraian dinilai lewat area tulis. Sumber: OnlineExamMaker + panduan trainee.
- **TIDAK DIKETAHUI:** Rumus penilaian per tipe (proporsional vs all-or-nothing), bobot per soal, nilai per kompetensi. Tidak ditemukan bukti eksplisit dalam sumber yang dibaca.

### 4.5 Analisis soal

- **TERBUKTI:** Exam.net mengklaim grading + analyzing results. Sumber: OnlineExamMaker.
- **TIDAK DIKETAHUI:** Apakah ada daya beda + analisis pengecoh formal seperti Moodle/Exo/v2. Tidak ditemukan bukti eksplisit. Jangan klaim ada.

### 4.6 Import soal

- **TERBUKTI:** Exam.net mengklaim question library + customizable questions. Sumber: Capterra.
- **TIDAK DIKETAHUI:** Format import yang didukung (CSV / Word / Google Docs), dan apakah ada import massal. Tidak ditemukan bukti eksplisit dalam sumber yang dibaca.

### 4.7 Harga / model bisnis

- **TERBUKTI (dari agregator, bukan halaman harga resmi yang terverifikasi):** OnlineExamMaker menyebut Exam.net: trial gratis 75 hari, harga custom quote. Sumber: https://onlineexammaker.com/kb/10-best-anti-cheating-online-test-makers-for-teachers/.
- **TIDAK DIKETAHUI:** Harga per siswa / per sekolah saat ini. Halaman harga resmi tidak terverifikasi dalam riset ini. Jangan pakai angka "from 1 USD" dari agregator perbandingan sebagai harga resmi Exam.net (angka itu berasal dari halaman perbandingan pihak ketiga dan tidak dapat dipastikan).

### 4.8 Celah untuk konteks SMK (DISIMPULKAN)

Exam.net adalah SaaS cloud (bergantung internet + langganan). Untuk SMK yang butuh LAN/semi-daring + UKK + berita acara + kartu peserta + nilai per kompetensi, Exam.net tidak terbukti menyediakan semuanya dalam satu paket. Keunggulan yang layak ditiru: pola exam key + monitoring real-time + SEB sebagai opsi keamanan tertinggi (bukan wajib).

---

## 5. Quizizz (kini Wayground)

**Sumber utama:** halaman harga resmi Wayground (Quizizz), agregator (Capterra, SaaSworthy), dan panduan pengguna.

### 5.1 Alur ujian

- **TERBUKTI:** Ada dua mode pengiriman: live (langsung) dan student-paced (tugas / pekerjaan rumah). Guru dapat membuat kuis, lalu memilih "Assign homework" (dengan deadline) atau "Start a live quiz" (Classic / Instructor-paced). Sumber: tutorial pengguna (https://www.scribd.com/presentation/511859135/LED110-QUIZIZZ) dan halaman harga Wayground (https://wayground.com/ca/plans) yang mencantumkan "Live and student-paced delivery".
- **TERBUKTI:** Siswa bergabung dengan game code (dalam panduan lama: 5 digit) lewat join link. Sumber: panduan workshop (https://quizizzworkshop.weebly.com/uploads/5/1/5/7/51575227/why_quizizz.pdf).
- **DISIMPULKAN:** Alur ini adalah alur kuis kelas, bukan alur ujian terjadwal (tidak ada sesi + ruangan + proktor + berita acara).

### 5.2 Token / timer

- **TERBUKTI:** Akses memakai game code + nama/nickname, bukan token sesi proktor. Sumber: panduan workshop yang sama.
- **TERBUKTI:** Ada timer per soal (question timer) yang dapat diaktifkan/dinonaktifkan. Dalam panduan lama: jika timer ON, siswa mendapat 600 poin untuk benar + bonus waktu hingga 400; jika OFF, 1000 poin untuk benar. Sumber: panduan workshop yang sama.
- **DISIMPULKAN:** Timer di Quizizz adalah timer per soal + mekanisme poin kecepatan, bukan timer ujian keseluruhan (durasi paket) seperti Exo/v2/ANBK.

### 5.3 Anti-curang

- **TERBUKTI:** Halaman harga Wayground mencantumkan "Anti-cheating alerts" dan "Anti-cheating prevention" sebagai fitur (di paket berbayar / School & District). Sumber: https://wayground.com/ca/plans.
- **TERBUKTI:** Ada opsi acak urutan soal per siswa (jumble questions) dan acak jawaban per soal (jumble answers). Sumber: panduan workshop yang sama.
- **TIDAK DIKETAHUI:** Mekanisme detail anti-cheating (apa yang dideteksi, apakah auto-block, apakah device lock). Halaman harga hanya mencantumkan nama fitur, bukan mekanismenya. Jangan mengarang detail.

### 5.4 Penilaian

- **TERBUKTI:** Penilaian otomatis (autograding) + gamifikasi (poin, leaderboard, meme). Halaman harga mencantumkan "Gamification features, Autograding". Sumber: https://wayground.com/ca/plans.
- **TERBUKTI:** Ada AI grading dengan rubrik kustom / rubrik negara bagian (state rubrics) di paket School & District. Sumber: halaman harga yang sama ("Full AI grading with custom and state rubrics").
- **TIDAK DIKETAHUI:** Rumus penilaian per tipe (proporsional vs all-or-nothing), bobot per soal, nilai per kompetensi. Tidak ditemukan bukti eksplisit.

### 5.5 Analisis soal

- **TERBUKTI:** Ada laporan standar (standards-aligned reports), grafik pertumbuhan longitudinal (longitudinal growth graphs), dan laporan per soal (per-question data) menurut panduan pihak ketiga. Sumber: halaman harga Wayground + artikel EduGenius (https://www.edugenius.app/blog/microsoft-copilot-vs-quizizz-which-is-better-for-teachers).
- **TIDAK DIKETAHUI:** Daya beda + analisis pengecoh formal. Tidak ditemukan bukti eksplisit.

### 5.6 Import soal

- **TERBUKTI:** Ada fitur "Generate and import content, add questions" dan "Teleport from millions of standards-aligned resources". Sumber: halaman harga Wayground.
- **TIDAK DIKETAHUI:** Format import yang didukung (CSV / Word / spreadsheet) dan batasannya. Tidak ditemukan bukti eksplisit dalam sumber yang dibaca. (Alat pihak ketiga mengklaim konversi PDF → Quizizz, tetapi itu bukan fitur resmi Quizizz.)

### 5.7 Harga / model bisnis

- **TERBUKTI:** Model freemium: ada paket Basic gratis (terbatas: tipe soal terbatas, penyimpanan 20 aktivitas) dan paket School & District dengan harga custom (request a quote, pembayaran via PO / cek / kartu / ACH). Sumber: https://wayground.com/ca/plans.
- **TERBUKTI (dari agregator, perlu cek ulang):** Capterra mencantumkan paket Premier $99 per user per bulan, Standard $59, Intro $0. Sumber: https://www.capterra.com/compare/207437-234436/Quizizz-vs-Remind. Angka ini adalah data agregator, bukan klaim harga resmi mutlak — harga resmi per wilayah dapat berbeda.
- **DISIMPULKAN:** Untuk satu sekolah dengan banyak guru, biaya SaaS per user per bulan dapat menjadi beban berulang yang signifikan dibanding sistem self-hosted sekali bangun seperti v2.

### 5.8 Celah untuk konteks SMK (DISIMPULKAN)

Quizizz tidak terbukti menyediakan: jadwal + sesi + ruangan + token proktor, UKK, berita acara, kartu peserta, audit log, nilai per kompetensi. Kuat untuk formatif, lemah untuk sumatif resmi.

---

## 6. Kahoot!

**Sumber utama:** halaman harga resmi Kahoot!, panduan pihak ketiga, dan artikel pembanding.

### 6.1 Alur ujian

- **TERBUKTI:** Alur inti adalah game live: host menampilkan soal di layar bersama, peserta menjawab di perangkat masing-masing, lalu leaderboard ditampilkan. Ada juga mode assigned (tugas mandiri) untuk kahoot yang dikerjakan sendiri. Sumber: halaman harga Kahoot! (https://kahoot.com/register/pricing-personal/) dan artikel EduGenius (https://www.edugenius.app/blog/kahoot-vs-edugenius-which-is-better-for-teachers).
- **TERBUKTI:** Ada mode unplugged (satu layar bersama, host membacakan soal, host memilih jawaban konsensus) yang melacak rasio akurasi keseluruhan, bukan skor individu. Sumber: halaman harga Kahoot! yang sama.
- **DISIMPULKAN:** Kahoot! adalah platform game-based learning, bukan sistem ujian. Tidak ada jadwal/sesi/ruangan/proktor/berita acara.

### 6.2 Token / timer

- **TERBUKTI:** Akses memakai PIN game (game PIN) + nickname, bukan token sesi proktor. (PIN game adalah mekanisme umum Kahoot! yang dikenal luas; dalam riset ini tidak ditemukan halaman resmi yang merinci format PIN, sehingga formatnya TIDAK DIKETAHUI dari sumber yang dikutip.)
- **TERBUKTI:** Timer adalah timer per soal (hingga 4 menit per soal menurut wiki komunitas; halaman harga resmi tidak merinci durasi maksimum dalam cuplikan yang dibaca). Poin berbasis kecepatan: jawaban benar yang lebih cepat mendapat poin lebih banyak. Sumber: wiki komunitas Kahoot! (https://kahoot.fandom.com/wiki/Quiz) + artikel EduGenius.
- **DISIMPULKAN:** Timer per soal + poin kecepatan cocok untuk game, tidak cocok untuk ujian sumatif (menghukum siswa yang berpikir lama tetapi benar).

### 6.3 Anti-curang

- **TIDAK DIKETAHUI:** Tidak ditemukan bukti fitur anti-curang ujian (deteksi tab, auto-block, device lock, SEB) di Kahoot! dalam sumber yang dibaca. Yang ada adalah pengawasan langsung oleh host di ruangan (karena game live).
- **DISIMPULKAN:** Kahoot! mengandalkan kehadiran host + layar bersama, bukan kontrol teknis anti-curang.

### 6.4 Penilaian

- **TERBUKTI:** Penilaian berbasis poin kecepatan: jawaban benar mendapat poin (dalam mode klasik hingga 1000 poin per soal, dengan bonus kecepatan); jawaban salah 0. Ada juga mode akurasi (tanpa kecepatan) menurut dokumentasi komunitas. Sumber: wiki komunitas + artikel pembanding (https://triviamaker.com/kahoot-vs-quizziz/).
- **TERBUKTI:** Tipe soal di paket personal gratis: Quiz (pilihan ganda) + True/False. Tipe lain (typed answer, puzzle/urutan, poll, word cloud, brainstorm) tersedia di paket berbayar / pembanding. Sumber: halaman harga Kahoot! + artikel pembanding.
- **TIDAK DIKETAHUI:** Apakah ada esai dengan rubrik, bobot per soal, nilai per kompetensi. Tidak ditemukan bukti.

### 6.5 Analisis soal

- **TERBUKTI:** Ada laporan pasca-game: kinerja per soal + per siswa. Sumber: artikel EduGenius + panduan komunitas.
- **TIDAK DIKETAHUI:** Daya beda + analisis pengecoh formal. Tidak ditemukan bukti.

### 6.6 Import soal

- **TERBUKTI:** Ada question bank + template dari komunitas (halaman harga menyebut "half a billion questions created by our community"). Sumber: halaman harga Kahoot!.
- **TIDAK DIKETAHUI:** Format import massal (CSV / Word) yang didukung. Tidak ditemukan bukti eksplisit.

### 6.7 Harga / model bisnis

- **TERBUKTI (halaman harga resmi, personal):** Paket personal: Go gratis (10 peserta), + Bronze $3/bln ($36 ditagih tahunan), + Silver $7/bln ($84 tahunan), + Gold $12/bln ($144 tahunan), One $19/bln ($228 tahunan). Batas peserta: 10 / 50 / 100 / 200 / 800. Sumber: https://kahoot.com/register/pricing-personal/.
- **DISIMPULKAN:** Harga sekolah / bisnis berbeda dari harga personal dan harus dicek di halaman harga sekolah/bisnis Kahoot! sebelum dipakai untuk anggaran.

### 6.8 Celah untuk konteks SMK (DISIMPULKAN)

Kahoot! tidak cocok untuk ujian sumatif: poin kecepatan, batas peserta per paket, tidak ada UKK / berita acara / kartu peserta / audit log / nilai per kompetensi. Nilai yang dapat ditiru v2: hanya pola "energi kelas + umpan balik instan" untuk tryout/latihan, bukan untuk ujian resmi.

---

## 7. ProProfs Quiz Maker

**Sumber utama:** blog resmi ProProfs, agregator (TrustRadius, Capterra, Slashdot, SourceForge), dan panduan pengguna.

### 7.1 Alur ujian

- **TERBUKTI:** Alur: buat kuis dari nol atau dari template / pustaka soal → atur pengaturan → bagikan → kumpulkan hasil. Ada kuis scored, survei, dan personality quiz. Sumber: deskripsi agregator (https://www.saasworthy.com/product/proprofs-quiz-maker/reviews) + TrustRadius.
- **DISIMPULKAN:** Tidak ada alur jadwal + sesi + ruangan + proktor + berita acara seperti sistem ujian sekolah.

### 7.2 Token / timer

- **TERBUKTI:** Ada batas waktu kuis ("Max. Time to Complete Quiz", maksimum 240 menit menurut blog resmi ProProfs). Sumber: https://www.proprofs.com/c/c/news-and-updates/five-settings-to-consider-when-creating-practice-quizzes/?p=628.
- **TERBUKTI:** Ada pengaturan retake (unlimited atau hingga 5 attempt) dan password-protection. Sumber: blog resmi ProProfs yang sama + daftar fitur agregator (Slashdot).
- **TIDAK DIKETAHUI:** Apakah ada token per sesi yang dirilis proktor. Yang TERBUKTI adalah password-protection, bukan token sesi.

### 7.3 Anti-curang

- **TERBUKTI:** Ada pengaturan keamanan & pelacakan (Security & Tracking), pengacakan soal (shuffling), dan pooling soal (tampilkan subset dari bank besar per attempt). Sumber: blog resmi ProProfs yang sama.
- **TERBUKTI (dari agregator, bukan dokumentasi resmi yang terverifikasi):** Slashdot mengutip deskripsi vendor bahwa versi gratis "deterring cheating by disabling print screen and copy-paste functions during assessments". Sumber: https://slashdot.org/software/comparison/Exam-Testing-vs-ProProfs-Quiz-Maker/. Klaim ini diberi label agregator karena tidak ditemukan di halaman dokumentasi resmi yang dibaca langsung.
- **TIDAK DIKETAHUI:** Apakah ada deteksi pindah tab / auto-block / device lock / proctoring video. Tidak ditemukan bukti eksplisit. Satu ulasan pengguna menyebut harapan "advanced proctoring functionalities" ditambahkan — DISIMPULKAN bahwa proctoring lanjutan belum menjadi fitur inti yang terdokumentasi.

### 7.4 Penilaian

- **TERBUKTI:** Ada automated grading, weighted questions (soal berbobot), dan opsi penilaian fleksibel. Sumber: daftar fitur Slashdot (https://slashdot.org/software/p/ProProfs-Quiz-Maker/) yang mencantumkan Automatic Grading, Weighted Questions, Question Randomizing, Time Limits.
- **TERBUKTI:** Ada sertifikat yang dapat disesuaikan (customizable quiz certificates) dan hasil yang dipersonalisasi. Sumber: deskripsi agregator.
- **TIDAK DIKETAHUI:** Rumus penilaian per tipe (proporsional vs all-or-nothing), nilai per kompetensi, rubrik esai. Tidak ditemukan bukti eksplisit.

### 7.5 Analisis soal

- **TERBUKTI:** Ada laporan + analitik: dashboard real-time, laporan skor detail, analitik per soal (question-wise analytics), pelacakan progres peserta. Sumber: panduan pengguna pihak ketiga + agregator.
- **TIDAK DIKETAHUI:** Daya beda + analisis pengecoh formal. Tidak ditemukan bukti eksplisit.

### 7.6 Import soal

- **TERBUKTI:** Ada import soal dari kuis yang sudah ada dan dari file Excel. Sumber: panduan pengguna (https://www.scribd.com/document/407719765/265020227-Web-2-0-Quiz-Assessment-Tools-A-Quick-Guide-pdf) yang menjelaskan "Import Questions" dari kuis / file Excel; juga deskripsi Slashdot yang menyebut "author offline using Excel" + upload bank soal dari Excel.
- **DISIMPULKAN:** Pola import Excel ini mirip import CSV v2 (tapi format berbeda). ProProfs tidak terbukti mendukung import Word (.docx) langsung.

### 7.7 Harga / model bisnis

- **TERBUKTI (dari agregator, perlu cek ulang):** TrustRadius mencantumkan: Free $0 selamanya, Basic $9/bln (tahunan), Team $39/bln, Business $69/bln, Enterprise $199/bln. Sumber: https://www.trustradius.com/compare-products/google-surveys-vs-proprofs-quiz-maker. Angka ini adalah data agregator, bukan klaim harga resmi mutlak.
- **TIDAK DIKETAHUI:** Harga resmi saat ini di halaman ProProfs (tidak terverifikasi langsung dalam riset ini).

### 7.8 Celah untuk konteks SMK (DISIMPULKAN)

ProProfs tidak terbukti menyediakan: jadwal/sesi/ruangan, token proktor, UKK, berita acara, kartu peserta, audit log, nilai per kompetensi. Mekanisme yang layak ditiru v2: pooling soal (subset acak dari bank besar) + shuffling + soal berbobot + import Excel/CSV + sertifikat (sebagai inspirasi untuk data sertifikat UKK, bukan penerbitan sertifikat resmi).

---

## 8. ANBK (Asesmen Nasional Berbasis Komputer) — dokumen publik saja

**Batasan:** Bagian ini HANYA memakai dokumen publik Pusmendik/BSKAP dan turunan yang mengutip POS/juknis ANBK. Tidak ada klaim fitur dari sumber non-publik.

**Sumber utama:** POS ANBK 2023 (salinan Scribd yang mengutip POS), panduan teknis ANBK semi-daring, juknis ANBK 2024 (ringkasan Sonora), dan penjelasan bentuk soal AN (Pusmendik).

### 8.1 Alur ujian (TERBUKTI dari POS/juknis yang dikutip)

1. Proktor hadir 30 menit sebelum pelaksanaan; memeriksa kesiapan ruang + peralatan; menjalankan aplikasi ANBK di komputer proktor/server lokal.
2. Proktor membagikan kartu login kepada peserta pada awal sesi.
3. Proktor merilis token agar peserta bisa memasuki laman AN (rilis token via CBTSync beberapa menit sebelum peserta memasukkan token; untuk semi-daring, token dirilis ke server pusat).
4. Peserta login dengan username + password sesuai kartu login, lalu memasukkan token yang sudah dirilis proktor.
5. Setelah login, keikutsertaan peserta tidak dapat digantikan orang lain.
6. Peserta melakukan latihan menjawab soal sebelum mengerjakan AN, lalu mulai mengerjakan setelah tanda waktu mulai.
7. Peserta yang ingin meninggalkan ruangan selama AN harus izin pengawas, tanpa penambahan waktu.
8. Setelah waktu selesai, pengawas mempersilakan peserta berhenti; proktor melakukan logout peserta yang lupa logout, mengunggah jawaban ke server pusat setiap sesi berakhir, melakukan force selesai bila ada peserta yang masih berstatus mengerjakan, lalu kepala satuan pendidikan + proktor + pengawas menandatangani berita acara + daftar hadir.
9. Proktor melakukan backup file (.source) setiap sesi (menu Backup Hapus) dan mengunggah hasil cetak form laporan yang ditandatangani ke laman ANBK.

Sumber: POS ANBK 2023 (https://id.scribd.com/document/637578195/POS-ANBK-2023-ok), panduan teknis semi-daring (https://id.scribd.com/document/670945607/Moda-Pelaksanaan-ANBK-Semi-Daring), ringkasan juknis ANBK 2024 (https://www.sonora.id/read/424155112/link-download-juknis-anbk-2024-pdf-dan-tahapan-pelaksanaannya?page=all).

### 8.2 Token / timer

- **TERBUKTI:** Token dirilis proktor (bukan dicetak di kartu login; kartu login berisi username/password, token diumumkan/dirilis terpisah). Token sesi dibuka pada jam yang ditentukan (contoh gladi bersih: token Sesi 1 dibuka 07.20, Sesi 2 10.30, Sesi 3 14.10 waktu setempat). Sumber: prosedur gladi bersih ANBK 2023 (https://smpn8solo.sch.id/prosedur-gladi-bersih-anbk-2023-jenjang-smp-smplb-paket-b-sederajat/page/2/?et_blog) + POS ANBK 2023.
- **TERBUKTI:** Jika peserta keluar tanpa logout lalu relogin: proktor mencatat token aktif di menu status tes, menekan "Simpan Semua", mereset login peserta (Reset Login Peserta), lalu menginformasikan token ke peserta. Sumber: panduan teknis semi-daring.
- **TERBUKTI:** Meninggalkan ruangan tidak menambah waktu. Sumber: POS ANBK 2023.
- **TIDAK DIKETAHUI:** Format/panjang token ANBK, masa berlaku token dalam detik, dan apakah ada rate limit / revoke token. Tidak ditemukan bukti eksplisit dalam dokumen publik yang dibaca. Jangan mengarang angka.
- **TIDAK DIKETAHUI:** Apakah timer ANBK berjangkar server atau browser. Dokumen publik yang dibaca tidak merinci mekanisme timer. Jangan klaim.

### 8.3 Anti-curang

- **TERBUKTI:** Tata tertib peserta melarang: menanyakan jawaban kepada siapa pun; bekerja sama; memberi/menerima bantuan dalam menjawab; memperlihatkan pekerjaan sendiri / melihat pekerjaan peserta lain; membawa catatan dan/atau menggunakan perangkat komunikasi. Sumber: POS ANBK 2023.
- **TERBUKTI:** Pengawas wajib: menjaga ketertiban, memberi peringatan + sanksi kepada peserta yang curang, melarang orang tidak berwenang masuk ruang AN, tidak membawa/menggunakan alat komunikasi/kamera, tidak memberi isyarat/petunjuk/bantuan terkait jawaban. Sumber: tata tertib ANBK (https://id.scribd.com/document/604517055/8-Tata-Tertib-Peserta-Pengawas-Proktor-dan-Teknisi-ANBK-Tahun-2022).
- **TERBUKTI:** Peserta mengerjakan lewat ExamBrowser (aplikasi klien) + ProktorBrowser / CBTSync di sisi proktor (moda daring: ProktorBrowser; moda semi-daring: VirtualBox/VHD + Exambrowser Admin). Sumber: ringkasan juknis ANBK 2024 + panduan teknis semi-daring.
- **TIDAK DIKETAHUI:** Apakah ExamBrowser mendeteksi pindah tab / memblokir screenshot / mengunci perangkat. Dokumen publik yang dibaca tidak merinci mekanisme teknis ExamBrowser. Jangan klaim fitur teknis yang tidak ada di dokumen publik.

### 8.4 Penilaian

- **TERBUKTI:** ANBK tidak memuat skor/nilai individual; tidak menentukan kelulusan; hasilnya agregat tingkat satuan pendidikan. (Pernyataan ini juga tercantum di ANALISIS.md Bab 12 dan konsisten dengan penjelasan resmi AN.)
- **TERBUKTI:** Bentuk soal AN: Pilihan Ganda, Pilihan Ganda Kompleks, Menjodohkan, Isian Singkat, Uraian. Sumber: penjelasan bentuk soal AN (https://kuliahdimana.id/index.php/news/read/1202/Serba-Serbi-ANBK-2022) yang merujuk Pusmendik + materi bimtek (https://id.scribd.com/document/646466117/tEKNIS-Pelaksanaan-Asesmen-Nasional).
- **TERBUKTI:** AKM dilaksanakan secara adaptif (soal menyesuaikan kemampuan siswa). Sumber: materi bimtek yang sama.
- **TIDAK DIKETAHUI:** Rumus penilaian per bentuk soal (mis. skor PGK), bobot, konversi ke skala. Dokumen publik yang dibaca tidak merinci rumus. Jangan klaim rumus.
- **DISIMPULKAN:** Karena ANBK tidak menghasilkan nilai individu, mekanisme penilaiannya TIDAK dapat dipakai sebagai model penilaian ujian sekolah v2. Yang dapat dipakai adalah pola operasionalnya (token, proktor, berita acara), bukan penilaiannya.

### 8.5 Analisis soal

- **TIDAK DIKETAHUI:** Tidak ditemukan bukti di dokumen publik yang dibaca bahwa ANBK menyediakan laporan daya beda / analisis pengecoh untuk sekolah. Hasil AN dilaporkan sebagai agregat mutu sekolah (Rapor Pendidikan), bukan sebagai analisis butir untuk guru. Jangan klaim ANBK menyediakan analisis butir seperti Exo/v2.

### 8.6 Import soal

- **TERBUKTI (DISIMPULKAN dari sifat sistem):** Soal AN disiapkan pusat (BSKAP/Pusmendik) sebagai dokumen negara yang bersifat rahasia; sekolah TIDAK mengimpor/membuat soal AN. Sumber: penjelasan instrumen AN (https://www.sinau-thewe.com/2021/08/bentuk-soal-komponen-dan-instrumen.html) yang menyebut instrumen disiapkan BSKAP dan bersifat rahasia.
- **DISIMPULKAN:** Konsep "import soal" tidak berlaku untuk ANBK. Ini adalah perbedaan mendasar dengan v2 (guru sekolah membuat + mengimpor soal sendiri).

### 8.7 Harga / model bisnis

- **TERBUKTI (DISIMPULKAN dari sifat program):** ANBK adalah program pemerintah (bukan produk SaaS berbayar per sekolah). Biaya yang ditanggung sekolah adalah sarana (komputer, jaringan, proktor) — bukan lisensi per siswa. Tidak ditemukan dokumen publik yang menyebut harga lisensi ANBK, karena memang bukan produk komersial.
- **TIDAK DIKETAHUI:** Anggaran pelaksanaan ANBK per sekolah (bervariasi per daerah/sarana). Jangan mengarang angka.

### 8.8 Yang BOLEH dan TIDAK BOLEH ditiru v2 dari ANBK

- **BOLEH ditiru (pola operasional, TERBUKTI di POS):** rilis token oleh proktor menjelang mulai; kartu login terpisah dari token; reset login peserta; force selesai; berita acara + daftar hadir ditandatangani; backup per sesi; pengawas memberi sanksi + mencatat insiden.
- **TIDAK BOLEH ditiru / tidak berlaku untuk v2:** hasil tanpa nilai individu (v2 HARUS memberi nilai individu); soal rahasia pusat (v2 soal dibuat guru sekolah); peserta sampel acak (v2 seluruh siswa kelas yang diujikan); tidak ada UKK (v2 wajib UKK).

---

## 9. Extraordinary CBT v4.7.0 — kolom pembanding (dari bahan baca)

Semua butir di bawah ini berasal dari `PELAJARI-EXTRAORDINARY-CBT.md` + ANALISIS.md Bab 13. Label TERBUKTI di sini artinya "terdapat di kajian", bukan "terverifikasi jalan live" (lihat catatan di Bab 0).

| Dimensi | Extraordinary CBT v4.7.0 |
|---|---|
| **(1) Alur ujian** | Bank soal → jadwal (8 sesi) → token → ujian → koreksi → hasil/ledger. Ada event (gelombang) + ruangan (room_members) + sesi peserta. Pause & resume jadwal; reset ujian; force finish; reset login peserta. |
| **(2) Token / timer** | Token: generate, revoke, expiry default 900 detik. Timer ujian + tanda ragu-ragu. `LOGIN_ONCE_ATIME=Y`. Fullscreen enforcement. |
| **(3) Anti-curang** | Tabel `devices` (verifikasi perangkat, device lock — 1 perangkat per peserta). `AUTO_BLOCK=Y`, `MAX_COUNTER_OUT=1` → keluar ujian 1x = auto-block. Log keluar ujian (`out_ujian_counter`), alasan blokir, `antiblock`. |
| **(4) Penilaian** | Koreksi otomatis (objektif) + koreksi manual uraian. AI opinion di koreksi (`ai_opinion` di `check_answers` & `jawaban_pesertas`) — opsional, butuh API key OpenAI (gpt-4o-mini) / Gemini. Nilai per kompetensi: tiap soal terikat `competence_item` → hasil per kode kompetensi (`code01_correct/wrong/point/void` … s.d. 16). Koreksi: Exo mengklaim 16 kode soal (PG, PG kompleks, benar-salah, menjodohkan, isian singkat, esai, mengurutkan, setuju-tidak, dll) — ANALISIS.md mengoreksi bahwa `code01–code16` di dump adalah kode kelompok kompetensi / bagian penilaian, BUKAN daftar nama 16 tipe soal yang terverifikasi. |
| **(5) Analisis soal** | `banksoal_summaries` (rata-rata, standar deviasi, durasi min/max, nilai min/max) + `soal_summaries` (daya beda / discrimination, analisis pengecoh / distractor, total benar/salah/kosong). |
| **(6) Import soal** | Import bank soal; Google Docs / Word (.docx) termasuk rumus; import bank soal + upload CSV peserta. |
| **(7) Harga / model bisnis** | Gratis non-komersial + lisensi ecosystem (per kajian). Lisensi: generate dari akun di ecosystem.ekstraordinary.com → isi `SERVER_SECRET_LICENSE_KEY` di `.env`; satu lisensi untuk satu server. Binary tertutup (Go ~48 MB) + PostgreSQL 16, port 9988, `/adm`. Keterbatasan: instalasi kajian terhenti di gerbang lisensi; tidak dapat dikustomisasi sendiri; TIDAK ada modul UKK. |
| **Role / permission** | 3 role: EVIL (super admin), EVAL (guru), OPRA (proktor). Permission berbasis role-menu (`role_menus`), 100+ permission per aksi (mis. `0010H` = Tambah Soal). |
| **Laporan / dokumen** | Ledger nilai per peserta, hasil ujian, berita acara (dengan kop sekolah), kartu peserta, export Excel untuk rapor. Backup & restore bank soal + backup & restore master system. |

---

## 10. CBT SMK v2 — kolom pembanding (dari ANALISIS.md)

Semua butir di bawah ini berasal dari `ANALISIS.md`. Status: spesifikasi/rencana. Butir yang di ANALISIS.md berlabel asumsi / belum diputuskan tetap ditulis sebagai asumsi.

| Dimensi | CBT SMK v2 (spesifikasi ANALISIS.md) |
|---|---|
| **(1) Alur ujian** | Bank Soal → Kisi-kisi → Paket → Jadwal + Token → Kerjakan → Koreksi → Nilai → Laporan. Status bank: DRAF → SIAP → TERKUNCI → DIARSIPKAN. Status jadwal: DRAF → SIAP → DIJADWALKAN → BERJALAN → DIJEDA → BERJALAN → SELESAI → DINILAI → TERBIT. Siswa: login (NIS + password) → token → kerjakan (satu soal per layar, autosave per soal, timer server, ragu-ragu, navigasi soal) → submit 2 langkah (ringkasan review → modal konfirmasi). Jenis: PTS / PAS / USBK / UKK / Tryout. |
| **(2) Token / timer** | Token 6 karakter per jadwal/sesi; TIDAK ditampilkan di dashboard siswa sebelum sesi (diumumkan di ruangan); valid jika jadwal/sesi BERJALAN + belum revoke + belum lewat expiry. Timer berjangkar server; server tolak save/submit setelah deadline. Pause (asumsi): membekukan timer semua siswa — berbeda dari CBT-Go yang pause-nya TIDAK membekukan timer. Token kedua = token pertama hangus (keputusan final di MEMORY: token expiry 900 detik; toleransi telat 15 menit — ANALISIS.md Bab 24 masih mencatat aturan pause/unblock/token kedua sebagai belum diputuskan; bila ada perbedaan, ikuti keputusan terbaru yang dicatat koordinator). Tryout boleh tanpa token (asumsi). |
| **(3) Anti-curang** | Deteksi `visibilitychange` + `blur` dengan debounce 5 detik; 1x peringatan + counter 1; 2x auto-block (per-ujian). Nonaktifkan copy-paste (penghambat). Device lock sebagai deteksi/penghambat (bukan kunci kuat). Kunci jawaban TIDAK dikirim ke browser. Nilai dihitung server. Rate limit login (5 gagal per akun per 15 menit → 429 + kunci 15 menit). Token kuat hanya dengan rate limit + masa berlaku singkat + revoke. Pengacakan isi + urutan per siswa (2 lapis: isi paket per siswa + urutan tampil/opsi). Audit log append-only. Batasan jujur: web TIDAK bisa cegah penuh HP kedua, screenshot/foto, menoleh/berbisik, penyamaran, DevTools, listrik/jaringan padam, kolusi. |
| **(4) Penilaian** | Nilai akhir = Σ poin / Σ bobot × 100 (proporsional bobot). Nilai per kompetensi: Σ poin soal kompetensi k / Σ bobot soal kompetensi k × 100; satu soal → satu kompetensi utama. PGK default usulan proporsional (disimpan per soal); Menjodohkan default per pasangan (perlu diputuskan apakah mengikuti Framework AKM ketat atau proporsional). Isian Singkat default tidak case-sensitive + trim spasi. Esai manual dengan rubrik (skor 0 s.d. bobot); opini AI opsional tahap lanjut (saran, bukan nilai resmi). Soal kosong = 0, dicatat "kosong" (beda dari "salah"). 5 tipe standar AKM + 11 varian/ekstensi dengan label jujur (BUKAN klaim "16 tipe standar Kemendikbud"). UKK: teori 40% + praktik 60% (asumsi, dapat diubah); lulus ≥ 70; per unit kompetensi. |
| **(5) Analisis soal** | Tingkat kesukaran P = B/N; daya beda D = (BA−BB)/n (kelompok atas/bawah 27%); pengecoh (berfungsi bila dipilih >5% + lebih banyak kelompok bawah; 0% = opsi mati). Syarat N: N<10 → D tidak dapat dipercaya; N≥27 → metode 27% penuh. Status pasca-analisis: Pertahankan / Revisi / Buang-Arsipkan / Perlu pengajaran ulang. Rekap: rata-rata + N eksplisit, median, min/max/rentang, simpangan baku, % lulus, histogram. |
| **(6) Import soal** | Import CSV (1 baris = 1 soal; maks. 500 baris/berkas; validasi dry-run; kompatibel-mundur dengan CSV lama CBT-Go 4 kolom). Import Word (.docx): 1 dokumen = 1 bank; kontrak format ketat (`Soal 1. [PG] …`, `Stimulus:`, `Kunci: B`, `Rubrik:`); dokumen tidak sesuai kontrak ditolak, bukan ditebak; guru wajib pratinjau 100% sebelum bank Siap. Import Google Docs TIDAK didukung di v2 (guru ekspor ke .docx dulu). |
| **(7) Harga / model bisnis** | TIDAK ADA harga lisensi — v2 adalah sistem milik sekolah, self-hosted. Biaya adalah sarana: server lokal sekolah (LAN/semi-daring) sebagai skenario utama; VPS opsional untuk tryout dari rumah; cPanel DICORET. Stack: Go + React + PostgreSQL, satu binary (frontend di-embed via `go:embed` — usulan). Backup/restore dari dalam aplikasi + event/ruangan/kartu peserta termasuk v2. |
| **Role / permission** | 6 role: Admin (EVIL), Proktor (OPRA), Guru (EVAL), Pengawas (baru — tidak boleh lihat token, tidak ubah soal/jadwal/nilai), Siswa, Penguji UKK (baru). Pemisahan tugas (SoD): pembuat soal ≠ pemegang token; penguji UKK ≠ finalisator nilai UKK. |
| **Laporan / dokumen** | Rekap per kelas/paket, nilai per kompetensi, analisis butir, berita acara (per jadwal+sesi+ruangan, PDF), kartu peserta (PDF; token TIDAK dicetak di kartu), daftar hadir (PDF), export CSV/PDF. Sertifikat UKK: data + draf dari CBT; penerbitan resmi oleh lembaga berwenang. |

---

## 11. Tabel perbandingan mekanisme (7 dimensi × 9 sistem)

Legenda: ✅ = ada (TERBUKTI dari sumber yang dikutip) · ◐ = sebagian / terbatas / hanya di paket berbayar · ❌ = tidak ada bukti / tidak ada · ? = TIDAK DIKETAHUI (tidak ditemukan bukti cukup) · (A) = asumsi/spesifikasi v2, bukan bukti jalan · (K) = menurut kajian menu/skema, bukan uji live.

### 11.1 Alur ujian

| Sistem | Bank soal | Kisi-kisi / paket | Jadwal + sesi + ruangan | Berita acara | Kartu peserta | UKK |
|---|---|---|---|---|---|---|
| Moodle Quiz | ✅ bank soal + reuse | ◐ (soal dipilih ke kuis; tidak ada entitas paket/kisi-kisi Kurikulum Merdeka) | ❌ (jendela buka/tutup saja) | ❌ | ❌ | ❌ |
| Google Forms | ❌ (form, bukan bank) | ❌ | ❌ | ❌ | ❌ | ❌ |
| Exam.net | ✅ question library | ? | ✅ scheduling (klaim agregator) | ? | ? | ? |
| Quizizz | ✅ pustaka + teleport | ❌ | ❌ (live / tugas saja) | ❌ | ❌ | ❌ |
| Kahoot! | ✅ question bank komunitas | ❌ | ❌ (live / assigned saja) | ❌ | ❌ | ❌ |
| ProProfs | ✅ question bank + template | ◐ (pooling) | ❌ | ❌ | ◐ (sertifikat kuis, bukan kartu peserta ujian) | ❌ |
| ANBK | ❌ (soal pusat, rahasia) | ❌ (bukan kisi-kisi sekolah) | ✅ sesi + kelompok tes + ruangan (POS) | ✅ ditandatangani | ✅ kartu login | ❌ (ANBK bukan UKK) |
| Exo (K) | ✅ CRUD + duplikat + lock + backup bank | ◐ (tanpa entitas paket terpisah; jadwal → bank) | ✅ jadwal + 8 sesi + event + ruangan | ✅ + kop sekolah | ✅ | ❌ |
| v2 (A) | ✅ CRUD + lock 2 lapis + snapshot + versi | ✅ kisi-kisi + paket (manual / acak berstrata / campuran) | ✅ jadwal + sesi + ruangan + event | ✅ PDF per jadwal+sesi+ruangan | ✅ PDF (tanpa token) | ✅ teori + praktik + rubrik + bukti |

### 11.2 Token / timer

| Sistem | Token / kunci akses | Timer | Pause membekukan timer | Reset login / force finish |
|---|---|---|---|---|
| Moodle Quiz | ◐ password kuis + batas IP (bukan token sesi) | ✅ batas waktu + jendela buka/tutup + auto-submit | ? | ? |
| Google Forms | ❌ (link + batasi domain/email) | ❌ bawaan (add-on pihak ketiga) | ❌ (add-on: jeda — mekanisme TIDAK DIKETAHUI) | ❌ |
| Exam.net | ✅ exam key | ✅ time limits (klaim agregator) | ? | ? |
| Quizizz | ✅ game code (5 digit, panduan lama) | ◐ timer per soal + poin kecepatan | ❌ | ❌ |
| Kahoot! | ✅ game PIN | ◐ timer per soal + poin kecepatan | ❌ | ❌ |
| ProProfs | ◐ password-protection | ✅ maks. 240 menit (blog resmi) | ? | ? |
| ANBK | ✅ token dirilis proktor menjelang mulai | ✅ durasi sesi (mekanisme timer TIDAK DIKETAHUI) | ❌ (tinggalkan ruangan = tanpa tambah waktu) | ✅ reset login + force selesai (POS) |
| Exo (K) | ✅ generate + revoke + expiry 900 detik (default) | ✅ timer + ragu-ragu | ✅ pause & resume (apakah membekukan timer TIDAK DIKETAHUI dari kajian — jangan samakan dengan v2) | ✅ reset ujian + force finish + reset login |
| v2 (A) | ✅ token 6 karakter per jadwal/sesi + revoke + expiry 900 detik; token kedua hanguskan pertama | ✅ timer berjangkar server + auto-submit server | ✅ (asumsi — beda dari CBT-Go) | ✅ force finish + unblock + reset login |

### 11.3 Anti-curang

| Sistem | Deteksi pindah tab | Auto-block | Device lock | Kunci tidak ke browser | SEB / kiosk | Audit log |
|---|---|---|---|---|---|---|
| Moodle Quiz | ? | ❌ | ❌ | ? | ◐ (SEB integrasi terpisah, bukan bawaan) | ✅ log aktivitas + riwayat respons |
| Google Forms | ❌ bawaan | ❌ | ❌ | ❌ | ❌ | ❌ (hanya respons) |
| Exam.net | ? (Esc → minta alasan) | ? | ? | ? | ✅ SEB (level tertinggi, Windows) + app iOS | ? |
| Quizizz | ? (alert di paket berbayar — mekanisme ?) | ? | ? | ? | ❌ | ? |
| Kahoot! | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| ProProfs | ? | ? | ? | ? | ❌ (klaim agregator: nonaktifkan print screen + copy-paste) | ? |
| ANBK | ? (mekanisme ExamBrowser TIDAK DIKETAHUI) | ? | ? | ? | ✅ ExamBrowser (klien) — mekanisme detail TIDAK DIKETAHUI | ◐ (berita acara + daftar hadir; log teknis ?) |
| Exo (K) | ✅ counter keluar ujian | ✅ 1x keluar = auto-block (MAX_COUNTER_OUT=1) | ✅ tabel devices (1 perangkat/peserta) | ? | ✅ fullscreen enforcement | ◐ (log keluar ujian + alasan blokir) |
| v2 (A) | ✅ visibilitychange + blur, debounce 5 detik | ✅ 1x peringatan, 2x auto-block per-ujian | ◐ deteksi/penghambat (bukan kunci kuat) | ✅ (prinsip arsitektur) | ◐ deteksi fullscreen (Esc selalu bisa keluar) | ✅ append-only |

### 11.4 Penilaian

| Sistem | Auto-grade objektif | Esai manual + rubrik | Bobot per soal | Nilai per kompetensi | AI opini / AI grading |
|---|---|---|---|---|---|
| Moodle Quiz | ✅ | ✅ manual (rubrik = fitur terpisah, ?) | ✅ (poin per soal) | ? | ? |
| Google Forms | ✅ (kunci + poin) | ◐ (review manual; rubrik ❌) | ◐ (poin per soal; tanpa poin per pilihan / desimal) | ❌ | ❌ |
| Exam.net | ✅ Auto-Mark | ✅ area tulis (rubrik ?) | ? | ? | ? |
| Quizizz | ✅ autograding | ? | ? | ? (laporan standar, bukan nilai per kompetensi) | ✅ AI grading + rubrik (paket School/District) |
| Kahoot! | ✅ (poin kecepatan) | ❌ | ❌ (poin kecepatan, bukan bobot) | ❌ | ❌ |
| ProProfs | ✅ automated grading | ? | ✅ weighted questions | ? | ? |
| ANBK | ✅ / ❌ (tidak ada nilai individu) | — (tidak ada nilai individu) | ? | ❌ (agregat sekolah) | ❌ |
| Exo (K) | ✅ | ✅ manual | ✅ (bobot + poin salah per competence_item) | ✅ per kode kompetensi | ✅ AI opinion (opsional, API key sendiri) |
| v2 (A) | ✅ | ✅ rubrik (skor 0 s.d. bobot) | ✅ proporsional Σ poin/Σ bobot × 100 | ✅ per kompetensi (satu soal → satu kompetensi utama) | ◐ opini AI tahap lanjut (saran, bukan nilai resmi) |

### 11.5 Analisis soal

| Sistem | Tingkat kesukaran (P) | Daya beda (D) | Analisis pengecoh | Peringatan N kecil |
|---|---|---|---|---|
| Moodle Quiz | ✅ (statistik kuis) | ✅ discrimination index | ✅ frekuensi per jawaban | ? |
| Google Forms | ◐ (distribusi jawaban di Responses) | ❌ | ❌ | ❌ |
| Exam.net | ? | ? | ? | ? |
| Quizizz | ◐ (laporan per soal) | ? | ? | ? |
| Kahoot! | ◐ (laporan per soal) | ? | ? | ? |
| ProProfs | ◐ (analitik per soal) | ? | ? | ? |
| ANBK | ❌ (tidak ada bukti di dokumen publik) | ❌ | ❌ | — |
| Exo (K) | ✅ (banksoal_summaries) | ✅ (soal_summaries) | ✅ (soal_summaries) | ? |
| v2 (A) | ✅ P = B/N | ✅ D = (BA−BB)/n, 27% | ✅ >5% + kelompok bawah | ✅ N<10 tidak dipercaya; N≥27 metode 27% penuh |

### 11.6 Import soal

| Sistem | CSV | Word (.docx) | Google Docs | Antar-form / pustaka | Validasi dry-run |
|---|---|---|---|---|---|
| Moodle Quiz | ❌ (format utama GIFT/XML) | ❌ bawaan (via konversi ke GIFT) | ❌ | ✅ antar-course (bank soal) | ? |
| Google Forms | ❌ bawaan | ❌ bawaan (add-on) | ❌ bawaan | ✅ antar-form | ❌ |
| Exam.net | ? | ? | ? | ✅ question library | ? |
| Quizizz | ? | ? | ? | ✅ teleport + generate/import | ? |
| Kahoot! | ? | ? | ? | ✅ question bank komunitas | ? |
| ProProfs | ◐ (Excel) | ? | ? | ✅ dari kuis yang ada | ? |
| ANBK | — (soal pusat) | — | — | — | — |
| Exo (K) | ✅ CSV peserta; import bank soal | ✅ .docx termasuk rumus | ✅ Google Docs | ✅ import bank soal | ? |
| v2 (A) | ✅ 500 baris/berkas + dry-run + kompatibel-mundur | ✅ kontrak format ketat + pratinjau 100% | ❌ (ekspor ke .docx dulu) | ✅ duplikat bank | ✅ |

### 11.7 Harga / model bisnis

| Sistem | Model | Biaya yang TERBUKTI / tercatat | Catatan |
|---|---|---|---|
| Moodle Quiz | Open-source, self-hosted | Tanpa biaya lisensi (moodle.com); hosting + dukungan ditanggung sendiri / penyedia layanan | MoodleCloud / hosting terkelola: harga TIDAK DIKETAHUI dari riset ini |
| Google Forms | SaaS (bagian Workspace) | Personal tanpa biaya; Education Fundamentals tanpa biaya untuk institusi yang memenuhi syarat (Google for Education) | Timer/proctoring butuh add-on berbayar pihak ketiga |
| Exam.net | SaaS cloud | Trial 75 hari + custom quote (agregator OnlineExamMaker) | Harga resmi saat ini TIDAK DIKETAHUI |
| Quizizz (Wayground) | Freemium SaaS | Basic gratis (terbatas); School/District custom quote / PO (wayground.com) | Agregator: Premier $99, Standard $59 per user/bln — perlu cek ulang |
| Kahoot! | Freemium SaaS | Personal: Go gratis (10 peserta) s.d. One $19/bln (800 peserta) — halaman harga resmi personal | Harga sekolah/bisnis berbeda; cek halaman sekolah/bisnis |
| ProProfs | Freemium SaaS | Agregator: Free $0 s.d. Enterprise $199/bln | Harga resmi saat ini TIDAK DIKETAHUI |
| ANBK | Program pemerintah | Bukan produk berbayar; biaya = sarana (komputer, jaringan, proktor) | Tidak ada harga lisensi |
| Exo | Binary tertutup + lisensi ecosystem | Gratis non-komersial + lisensi ecosystem (per kajian); 1 lisensi = 1 server | Instalasi kajian terhenti di gerbang lisensi |
| v2 | Milik sekolah, self-hosted | Tanpa lisensi; biaya = server lokal (LAN) + listrik + perawatan | VPS opsional untuk tryout; cPanel dicoret |

---

## 12. Rekomendasi untuk CBT SMK v2 (DISIMPULKAN — bukan klaim vendor)

1. **Ambil dari Moodle:** bank soal terpusat + reuse antar kuis + statistik kuis (discrimination index + distribusi pengecoh) sebagai referensi implementasi analisis butir v2. Jangan meniru Moodle sebagai LMS — v2 adalah sistem ujian, bukan LMS.
2. **Ambil dari ANBK/Exo:** pola token dirilis proktor menjelang mulai + kartu login terpisah dari token + reset login + force selesai + berita acara ditandatangani. Ini adalah pola yang sudah teruji di POS nasional dan di Exo.
3. **Ambil dari Exam.net:** opsi keamanan berlapis — browser biasa untuk ujian risiko rendah (dengan pengawas), SEB/kiosk sebagai opsi untuk ujian risiko tinggi bila lab siap. Jangan jadikan SEB wajib di v2 (banyak lab SMK belum siap).
4. **Ambil dari ProProfs/Quizizz:** pooling soal (subset acak dari bank besar) + shuffling + soal berbobot. Jangan ambil poin kecepatan (Kahoot!/Quizizz) untuk ujian sumatif.
5. **Jangan ambil dari Google Forms/Kahoot!:** arsitektur form/game sebagai ujian sumatif. Keduanya tidak punya jadwal/sesi/ruangan, audit log, UKK, berita acara, atau analisis butir formal.
6. **Pertahankan keunggulan v2 yang tidak dimiliki pembanding mana pun secara sekaligus:** UKK teori+praktik + rubrik asesor + bukti + nilai per unit; nilai per kompetensi; KKTP per mapel/TP (bukan KKM hardcode); operasi LAN/semi-daring; audit log append-only; snapshot soal per paket/jadwal.
7. **Keputusan yang masih perlu koordinator:** (a) apakah pause membekukan timer — ANALISIS.md mengusulkan ya, tetapi Bab 24 mencatat aturan pause/unblock/token kedua belum diputuskan; (b) skema skor PGK/Menjodohkan — Framework AKM ketat vs proporsional; (c) apakah 1 pelanggaran gratis untuk kasus jaringan diberlakukan.

---

## 13. Daftar sumber

### Sumber bahan baca (tugas)

- `~/workspace/cbt-smk-v2/ANALISIS.md`
- `~/workspace/extraordinary-cbt/PELAJARI-EXTRAORDINARY-CBT.md`

### Sumber publik — Moodle

- https://moodle.com/news/test-learners-knowledge-using-quiz-feature/
- https://moodle.com/us/news/quiz-analytics-uncovered/
- https://moodle.com/us/about/open-source/
- https://osgoode.yorku.ca/wp-content/uploads/2014/09/Creating-Quizzes-in-your-course.pdf
- https://swatkb.atlassian.net/wiki/spaces/MOODLE/pages/20906321/Creating+a+timed+quiz+or+exam+in+Moodle
- https://acemyhomework.com/blog/can-moodle-detect-cheating
- https://bpb-us-e1.wpmucdn.com/u.cotc.edu/dist/3/152/files/2020/03/GIFT-Format.pdf
- https://arxiv.org/pdf/2010.15577

### Sumber publik — Google Forms

- https://ipv4.google.com/forms/about/?hl=en_uk
- https://edu.google.com/intl/ALL_au/workspace-for-education/editions/education-fundamentals/
- https://knowledge.workspace.google.com/admin/getting-started/editions/google-workspace-for-education-faq
- https://gadgets360.com/how-to/features/google-form-create-link-free-surveys-quiz-questionnaire-2299138
- https://chatfacade.com/calculate/create-quiz-google-forms-calculate-score.html
- https://www.guidingtech.com/import-questions-into-google-forms/
- https://workspace.google.com/marketplace/app/form_presenter+timer/93519377337
- https://workspace.google.com/marketplace/app/timer_+_proctor_for_google_forms_quiz_te/1083655933830
- https://workspace.google.com/marketplace/app/quiz_helper_by_alice_keeler/581441704194

### Sumber publik — Exam.net

- https://exam.net/blog/ensuring-academic-integrity-with-safe-exam-browser-seb-and-exam-net
- https://jobsplus.gov.mt/media/j0cd3szr/tcqa-401-online-assessment-guidelines-for-trainees-rev-02-21012022.pdf
- https://onlineexammaker.com/kb/10-best-anti-cheating-online-test-makers-for-teachers/
- https://www.capterra.in/compare/1015788/1063218/examnet/vs/doexam

### Sumber publik — Quizizz / Wayground

- https://wayground.com/ca/plans
- https://www.capterra.com/compare/207437-234436/Quizizz-vs-Remind
- https://quizizzworkshop.weebly.com/uploads/5/1/5/7/51575227/why_quizizz.pdf
- https://www.scribd.com/presentation/511859135/LED110-QUIZIZZ
- https://www.edugenius.app/blog/microsoft-copilot-vs-quizizz-which-is-better-for-teachers

### Sumber publik — Kahoot!

- https://kahoot.com/register/pricing-personal/
- https://www.edugenius.app/blog/kahoot-vs-edugenius-which-is-better-for-teachers
- https://triviamaker.com/kahoot-vs-quizziz/
- https://kahoot.fandom.com/wiki/Quiz

### Sumber publik — ProProfs

- https://www.proprofs.com/c/c/news-and-updates/five-settings-to-consider-when-creating-practice-quizzes/?p=628
- https://www.saasworthy.com/product/proprofs-quiz-maker/reviews
- https://slashdot.org/software/p/ProProfs-Quiz-Maker/
- https://slashdot.org/software/comparison/Exam-Testing-vs-ProProfs-Quiz-Maker/
- https://www.trustradius.com/compare-products/google-surveys-vs-proprofs-quiz-maker
- https://www.scribd.com/document/407719765/265020227-Web-2-0-Quiz-Assessment-Tools-A-Quick-Guide-pdf

### Sumber publik — ANBK / Pusmendik-BSKAP (dokumen publik saja)

- https://id.scribd.com/document/637578195/POS-ANBK-2023-ok (salinan POS ANBK 2023)
- https://id.scribd.com/document/670945607/Moda-Pelaksanaan-ANBK-Semi-Daring (panduan teknis semi-daring)
- https://id.scribd.com/document/604517055/8-Tata-Tertib-Peserta-Pengawas-Proktor-dan-Teknisi-ANBK-Tahun-2022
- https://id.scribd.com/document/646466117/tEKNIS-Pelaksanaan-Asesmen-Nasional (materi bimtek bentuk soal AN)
- https://kuliahdimana.id/index.php/news/read/1202/Serba-Serbi-ANBK-2022 (bentuk soal AN, merujuk Pusmendik)
- https://www.sinau-thewe.com/2021/08/bentuk-soal-komponen-dan-instrumen.html (instrumen AN disiapkan BSKAP, bersifat rahasia)
- https://smpn8solo.sch.id/prosedur-gladi-bersih-anbk-2023-jenjang-smp-smplb-paket-b-sederajat/page/2/?et_blog (prosedur gladi bersih, rilis token)
- https://www.sonora.id/read/424155112/link-download-juknis-anbk-2024-pdf-dan-tahapan-pelaksanaannya?page=all (ringkasan juknis ANBK 2024)

---

## 14. Penutup

Laporan ini adalah analisis pembanding, bukan spesifikasi baru dan bukan perubahan kode. Semua klaim tentang aplikasi luar diberi label bukti agar koordinator dapat membedakan fakta vendor, simpulan tim, dan hal yang belum diketahui. Bila koordinator memerlukan verifikasi lebih dalam untuk satu vendor (mis. harga resmi Exam.net / ProProfs saat ini, atau mekanisme teknis ExamBrowser ANBK), verifikasi tersebut harus dilakukan ke halaman resmi vendor / dokumen POS terbaru — bukan dari agregator — sebelum dipakai untuk keputusan pembelian atau desain.
