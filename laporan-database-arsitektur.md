# Laporan Riset Database & Arsitektur — CBT SMK v2

**Proyek:** CBT SMK v2 — web ujian online SMK
**Backend:** Go (single binary)
**Database:** PostgreSQL
**Target skala:** ratusan siswa serentak, mode semi-daring di server lokal sekolah
**Sifat laporan:** riset & desain konseptual. Tidak berisi kode aplikasi, tidak berisi perintah SQL DDL / migrasi dalam bentuk kode.

---

## Ringkasan Eksekutif

CBT SMK v2 adalah evolusi dari CBT-Go lama (database `cbt_go`, PostgreSQL, migrasi + seed otomatis). Arsitektur yang direkomendasikan tetap memakai pola yang sudah terbukti: satu binary Go + satu database PostgreSQL + penyimpanan file lokal, dengan tambahan yang penting untuk skala dan operasi sekolah:

1. **Pemisahan tegas antara data master, bank soal, pelaksanaan ujian, dan hasil.** Soal yang sudah dipakai ujian tidak boleh dihapus atau diubah secara merusak — hanya boleh dinonaktifkan / diarsipkan / dibuat versi baru.
2. **Jawaban peserta disimpan per soal per peserta per sesi, dengan penyimpanan bertahap (autosave).** Nilai akhir dihitung dari jawaban yang tersimpan, bukan dari memori browser.
3. **Indeks difokuskan pada tiga jalur panas:** login siswa, pengambilan soal ujian, dan rekap nilai.
4. **Backup adalah bagian dari desain, bukan tambahan.** Backup terjadwal, terenkripsi, dan diuji restore secara berkala.
5. **Mode semi-daring berarti server lokal sekolah adalah sumber kebenaran (source of truth) saat ujian berlangsung.** Tidak ada ketergantungan pada internet untuk alur ujian inti.

---

## Asumsi & Batasan Desain

- Satu sekolah, satu instalasi, satu database. Multi-sekolah tidak menjadi target v2.
- Ujian berlangsung di jaringan lokal sekolah (LAN / Wi-Fi lokal). Internet hanya dipakai untuk hal non-kritis (misal unduh pembaruan, kirim backup ke luar lokasi jika ada).
- Ratusan siswa serentak berarti beban puncak terkonsentrasi pada: login massal dalam 5–10 menit, pengambilan soal awal, penyimpanan jawaban berkala, dan submit massal di akhir waktu.
- Perangkat peserta beragam (PC lab, laptop, HP). Sistem tidak boleh mengasumsikan satu perangkat per siswa, tapi harus mencatat perangkat yang dipakai.
- Soal memiliki banyak tipe (mengacu pada warisan CBT-Go: 16 tipe, termasuk pilihan ganda, pilihan ganda kompleks, benar-salah, isian singkat, menjodohkan, dan esai/uraian). Desain soal harus generik agar tipe baru tidak memaksa perubahan skema besar.
- Laporan ini tidak menentukan sintaks SQL, nama kolom fisik final, atau framework migrasi tertentu. Itu diputuskan pada tahap implementasi berdasarkan laporan ini.

---

## 1. Daftar Tabel Inti

Konvensi umum yang berlaku untuk hampir semua tabel:

- Setiap baris memiliki pengenal unik internal.
- Setiap baris mencatat kapan dibuat dan kapan terakhir diubah.
- Baris yang penting secara audit (ujian, jawaban, nilai, log) tidak dihapus fisik; penghapusan dilakukan dengan penandaan arsip / nonaktif (soft delete) agar jejak tetap ada.
- Kolom status memakai daftar nilai terkontrol (misal: aktif / nonaktif / arsip / draf / terbit / berjalan / selesai / diblokir), bukan teks bebas.

### 1.1. Pengguna, Peran, dan Kelembagaan

#### 1.1.1. Peran (Roles)

**Deskripsi:** Daftar peran yang dikenal sistem. Mengacu pada warisan CBT-Go: super admin, guru / evaluator, proktor / operator, dan peserta (siswa). Peran menentukan menu dan hak akses.

**Kolom-kolom kunci (dalam kata-kata):**
- Nama peran (misal: Super Admin, Guru, Proktor, Siswa, Penguji UKK).
- Kode peran yang stabil dipakai aplikasi (tidak berubah walau nama tampilan diubah).
- Deskripsi peran.
- Status aktif / nonaktif.

#### 1.1.2. Pengguna (Users)

**Deskripsi:** Akun login semua pihak: admin, guru, proktor, pengawas, siswa, penguji. Satu orang = satu akun. Data identitas sekolah (NIS, NIP, kelas) disimpan di tabel profil terkait, bukan ditumpuk semua di tabel ini.

**Kolom-kolom kunci:**
- Nama lengkap.
- Nama pengguna untuk login (unik, tidak boleh ganda).
- Alamat email (unik bila diisi; dipakai untuk pemulihan / pemberitahuan).
- Kata sandi dalam bentuk hash + penanda algoritma hash yang dipakai (kata sandi asli tidak pernah disimpan).
- Status akun: aktif, nonaktif, terkunci, perlu ganti kata sandi.
- Waktu login terakhir yang berhasil.
- Waktu perubahan kata sandi terakhir.
- Penanda apakah akun demo / akun contoh (agar mudah dibersihkan).
- Penanda arsip (soft delete).

#### 1.1.3. Relasi Pengguna–Peran (User Roles)

**Deskripsi:** Tabel penghubung karena satu pengguna bisa memegang lebih dari satu peran (misal guru yang juga menjadi proktor). Ini adalah relasi many-to-many antara pengguna dan peran.

**Kolom-kolom kunci:**
- Rujukan ke pengguna.
- Rujukan ke peran.
- Kapan peran itu diberikan dan oleh siapa.
- Status penugasan peran (aktif / dicabut).

#### 1.1.4. Izin per Peran (Role Permissions) — opsional tapi dianjurkan

**Deskripsi:** Daftar izin halus per aksi (misal: boleh membuat soal, boleh mengoreksi esai, boleh memulai / menjeda jadwal, boleh melihat rekap). Warisan CBT-Go / Extraordinary CBT memiliki 100+ izin per aksi; v2 tidak harus meniru jumlahnya, tapi mekanismenya perlu ada agar peran tidak dikeraskan di kode.

**Kolom-kolom kunci:**
- Rujukan ke peran.
- Nama izin (misal: buat soal, ubah jadwal, koreksi esai, ekspor nilai).
- Penanda diizinkan / tidak.
- Catatan / kelompok izin.

#### 1.1.5. Jurusan

**Deskripsi:** Data master jurusan / program keahlian di SMK (misal: RPL, TKJ, dll).

**Kolom-kolom kunci:**
- Kode jurusan (unik, singkat, stabil).
- Nama jurusan.
- Deskripsi singkat.
- Status aktif / nonaktif.

#### 1.1.6. Kelas

**Deskripsi:** Data master rombongan belajar. Satu kelas milik satu jurusan dan satu tingkat / angkatan.

**Kolom-kolom kunci:**
- Rujukan ke jurusan.
- Nama kelas (misal: XI RPL 1).
- Tingkat (X / XI / XII).
- Tahun ajaran / angkatan.
- Wali kelas (rujukan ke pengguna guru, boleh kosong).
- Kapasitas / jumlah siswa (boleh dihitung dari data siswa, tapi disimpan sebagai info).
- Status aktif / nonaktif.

#### 1.1.7. Profil Siswa (Student Profiles)

**Deskripsi:** Data akademik siswa, terpisah dari akun login. Satu pengguna dengan peran siswa memiliki tepat satu profil siswa.

**Kolom-kolom kunci:**
- Rujukan ke pengguna (akun login siswa).
- Nomor Induk Siswa / NISN (unik).
- Rujukan ke kelas saat ini.
- Rujukan ke jurusan (boleh diturunkan dari kelas, tapi disimpan untuk memudahkan rekap lintas kelas).
- Jenis kelamin, tanggal lahir (hanya bila diperlukan sekolah; bila tidak, jangan dikumpulkan).
- Nomor peserta ujian / nomor kartu (unik per ujian bila dipakai; lihat bagian jadwal).
- Status siswa: aktif, pindah, lulus, nonaktif.
- Riwayat kelas bila siswa naik kelas (disimpan sebagai catatan, bukan menimpa kelas lama pada data ujian yang sudah selesai — lihat aturan di Bagian 2).

#### 1.1.8. Profil Guru / Staf (Teacher Profiles) — opsional

**Deskripsi:** Data kepegawaian ringkas untuk guru / proktor / pengawas.

**Kolom-kolom kunci:**
- Rujukan ke pengguna.
- Nomor induk pegawai / NIP (bila ada).
- Mata pelajaran yang diampu (lihat juga tabel penugasan guru–mapel).
- Status aktif / nonaktif.

#### 1.1.9. Mata Pelajaran (Mapel)

**Deskripsi:** Data master mata pelajaran / mata diklat.

**Kolom-kolom kunci:**
- Kode mapel (unik).
- Nama mapel.
- Kelompok mapel (misal: produktif, normatif, adaptif — sesuai kebutuhan sekolah).
- Deskripsi singkat.
- Status aktif / nonaktif.

#### 1.1.10. Penugasan Guru–Mapel (Teacher–Subject Assignments)

**Deskripsi:** Tabel penghubung many-to-many: satu guru bisa mengampu banyak mapel, satu mapel bisa diampu banyak guru.

**Kolom-kolom kunci:**
- Rujukan ke guru (pengguna).
- Rujukan ke mapel.
- Tahun ajaran berlaku.
- Status penugasan.

---

### 1.2. Bank Soal

#### 1.2.1. Bank Soal (Question Banks)

**Deskripsi:** Wadah pengelompokan soal. Satu bank soal biasanya untuk satu mapel + satu tingkat / satu tujuan (misal: Bank Soal Pemrograman Web XI, Bank Soal USBK Matematika). Soal tidak berdiri sendiri tanpa bank.

**Kolom-kolom kunci:**
- Nama bank soal.
- Rujukan ke mapel.
- Rujukan ke jurusan / tingkat sasaran (boleh kosong bila umum).
- Pemilik / pembuat (rujukan ke pengguna guru).
- Deskripsi / tujuan bank soal.
- Status: draf, aktif, arsip.
- Penanda terkunci (bank yang terkunci tidak boleh diubah tanpa membuka kunci oleh peran yang berwenang — warisan fitur lock di CBT-Go).

#### 1.2.2. Soal (Questions)

**Deskripsi:** Inti bank soal. Satu baris = satu soal. Desain harus menampung banyak tipe soal tanpa membuat tabel baru per tipe. Perbedaan antar tipe ditampung lewat kolom tipe + kolom isi jawaban yang fleksibel dan terdokumentasi.

**Kolom-kolom kunci:**
- Rujukan ke bank soal.
- Tipe soal (daftar terkontrol: pilihan ganda, pilihan ganda kompleks, benar-salah, isian singkat, menjodohkan, uraian / esai, dan tipe lain yang didukung).
- Teks / isi soal (mendukung format teks kaya / HTML aman + rujukan ke media bila ada).
- Bobot / skor maksimum soal (angka; dipakai dalam perhitungan nilai).
- Tingkat kesulitan (misal: mudah, sedang, sulit) — dipakai untuk analisis dan perakitan paket.
- Kompetensi / indikator yang diukur (teks atau rujukan ke daftar kompetensi bila ada; lihat Bagian 1.2.5).
- Kunci jawaban / pedoman penskoran:
  - Untuk soal objektif: kunci jawaban terstruktur (misal: opsi benar, pasangan benar, atau jawaban yang diterima).
  - Untuk esai: rubrik / panduan koreksi dalam bentuk teks.
- Pembahasan / umpan balik (ditampilkan setelah ujian bila diizinkan).
- Media pendukung: rujukan ke file gambar / audio / video bila soal memakai media (disimpan di tabel file, bukan file biner di database).
- Status soal: draf, aktif, nonaktif, arsip.
- Versi soal (angka versi; setiap perubahan bermakna pada soal yang sudah dipakai ujian sebaiknya membuat versi baru, bukan menimpa — lihat Bagian 2).
- Pembuat dan penyunting terakhir.
- Penanda terkunci (soal terkunci tidak bisa diubah).
- Jumlah pemakaian (boleh dihitung, tapi disimpan sebagai cache untuk memudahkan laporan).

**Catatan desain penting:** Jangan menyimpan jawaban peserta di tabel soal. Jangan menyimpan skor peserta di tabel soal. Tabel soal hanya menyimpan soal + kunci + pedoman.

#### 1.2.3. Opsi Jawaban (Answer Options)

**Deskripsi:** Pilihan jawaban untuk soal tipe pilihan (pilihan ganda, pilihan ganda kompleks, benar-salah). Soal isian / esai tidak memakai tabel ini (jawabannya bebas).

**Kolom-kolom kunci:**
- Rujukan ke soal induk.
- Label opsi (A, B, C, D, dst — atau urutan numerik).
- Teks / isi opsi (mendukung media bila perlu).
- Penanda apakah opsi ini benar (untuk pilihan ganda: tepat satu benar; untuk pilihan ganda kompleks / benar-salah: bisa banyak).
- Urutan tampil opsi.
- Penanda apakah opsi diacak saat ujian (bila pengacakan per peserta diaktifkan, pengacakan dicatat di sisi pelaksanaan, bukan dengan mengubah tabel ini).
- Status aktif / nonaktif.

#### 1.2.4. Pasangan Menjodohkan (Matching Pairs) — bila diperlukan sebagai tabel sendiri

**Deskripsi:** Untuk soal menjodohkan, pasangan kiri–kanan disimpan sebagai baris-baris pasangan, bukan sebagai opsi biasa, agar koreksi otomatis bisa dilakukan per pasangan.

**Kolom-kolom kunci:**
- Rujukan ke soal induk.
- Pernyataan sisi kiri.
- Pasangan benar sisi kanan.
- Urutan tampil.

**Alternatif:** Bila tim implementasi ingin menyederhanakan, pasangan menjodohkan dapat disimpan sebagai struktur terstruktur di dalam kolom kunci jawaban soal, dengan dokumentasi format yang ketat. Keputusan ini harus dicatat dan tidak diubah di tengah jalan.

#### 1.2.5. Kompetensi / Indikator (Competencies) — opsional tahap awal, dianjurkan untuk nilai per kompetensi

**Deskripsi:** Daftar kompetensi / indikator pembelajaran yang diukur soal. Ini adalah fondasi fitur "nilai per kompetensi" yang belum ada di CBT-Go dan sering diminta.

**Kolom-kolom kunci:**
- Kode kompetensi.
- Nama kompetensi / indikator.
- Rujukan ke mapel.
- Deskripsi.
- Status aktif / nonaktif.

**Relasi:** Satu soal dapat mengukur satu atau lebih kompetensi (many-to-many via tabel penghubung Soal–Kompetensi). Bila v2 awal hanya mendukung satu kompetensi per soal, cukup simpan satu rujukan kompetensi di tabel soal dan tunda tabel penghubung sampai dibutuhkan.

#### 1.2.6. File / Media Soal (Question Media)

**Deskripsi:** Metadata file pendukung soal: gambar, audio, video, dokumen pendukung. File fisik disimpan di penyimpanan file (disk lokal server), bukan sebagai biner di database.

**Kolom-kolom kunci:**
- Nama file asli + nama file tersimpan.
- Jenis file / tipe media (gambar, audio, video, dokumen).
- Ukuran file.
- Rujukan ke soal / bank soal yang memakai file.
- Pengunggah dan waktu unggah.
- Status (aktif / dihapus).

---

### 1.3. Paket Ujian, Jadwal, dan Sesi

#### 1.3.1. Paket Ujian (Exam Packages)

**Deskripsi:** Definisi ujian yang dapat dipakai ulang: kumpulan soal + aturan (durasi, jumlah soal, pengacakan, nilai kelulusan, dsb). Contoh jenis: STS, SAS, USBK, UKK, Tryout (warisan CBT SMK / CBT-Go).

**Kolom-kolom kunci:**
- Nama paket ujian.
- Rujukan ke mapel.
- Jenis ujian (STS / SAS / USBK / UKK / Tryout / lainnya).
- Deskripsi / petunjuk ujian yang ditampilkan ke peserta.
- Durasi ujian dalam menit.
- Jumlah soal yang ditampilkan (bila paket mengambil subset acak dari bank; bila paket memakai susunan tetap, jumlah ini diturunkan dari susunan).
- Aturan pengacakan: acak urutan soal (ya / tidak), acak urutan opsi (ya / tidak).
- Nilai minimum kelulusan (KKM / passing grade), bila dipakai.
- Aturan anti-curang: batas pindah tab / keluar jendela sebelum peringatan, batas sebelum blokir otomatis, apakah fullscreen diwajibkan, apakah perangkat dicatat.
- Aturan tampilan hasil: apakah peserta boleh melihat pembahasan, kapan nilai ditampilkan.
- Status paket: draf, aktif, nonaktif, arsip.
- Pembuat paket.

#### 1.3.2. Susunan Soal Paket (Package–Question Items)

**Deskripsi:** Tabel penghubung many-to-many antara paket ujian dan soal: soal mana saja yang masuk paket, dalam urutan apa, dengan bobot apa.

**Kolom-kolom kunci:**
- Rujukan ke paket ujian.
- Rujukan ke soal.
- Urutan soal dalam paket.
- Bobot soal dalam paket ini (boleh berbeda dari bobot default di tabel soal; bobot di sini yang dipakai untuk penilaian paket tersebut).
- Status item (aktif / dikeluarkan dari paket).

**Penting:** Jika soal dihapus dari paket setelah ujian berjalan, baris tidak dihapus fisik — ditandai nonaktif agar hasil ujian lama tetap dapat dihitung ulang.

#### 1.3.3. Jadwal Ujian (Exam Schedules)

**Deskripsi:** Pelaksanaan konkret dari sebuah paket ujian: kapan, untuk kelas mana, dengan token apa. Satu paket bisa dijadwalkan berkali-kali (misal: gelombang 1, 2, 3; atau kelas berbeda di jam berbeda).

**Kolom-kolom kunci:**
- Rujukan ke paket ujian.
- Nama / label jadwal (misal: STS Pemrograman Web — XI RPL 1 — Gelombang 1).
- Waktu mulai dan waktu selesai yang direncanakan.
- Durasi efektif (boleh sama dengan durasi paket, atau dioverride per jadwal).
- Token ujian (kode yang dimasukkan peserta untuk masuk; unik per jadwal pada masa berlaku).
- Status jadwal: draf, dijadwalkan, berjalan, dijeda (pause), selesai, dibatalkan.
- Alasan / catatan jeda (bila dijeda).
- Kelas / rombongan yang diundang (via tabel peserta jadwal, lihat 1.3.4).
- Ruang / lab pelaksanaan (teks atau rujukan ke tabel ruangan bila ada).
- Pengawas yang ditugaskan (rujukan ke pengguna).
- Pengaturan override per jadwal (misal: durasi diperpanjang, token diubah) — dicatat agar audit jelas.
- Pembuat jadwal.

#### 1.3.4. Peserta Jadwal (Schedule Participants)

**Deskripsi:** Daftar siswa yang berhak mengikuti sebuah jadwal. Ini adalah penghubung many-to-many antara jadwal dan siswa.

**Kolom-kolom kunci:**
- Rujukan ke jadwal ujian.
- Rujukan ke siswa (profil siswa).
- Nomor peserta / nomor kartu (bila dipakai).
- Status keikutsertaan: diundang, hadir, tidak hadir, sedang mengerjakan, selesai, diblokir, dikeluarkan.
- Waktu mulai mengerjakan (aktual, bukan jadwal).
- Waktu selesai mengerjakan (aktual).
- Sisa waktu / status timer pada saat terakhir tersimpan.
- Alasan blokir / catatan proktor bila ada.

#### 1.3.5. Sesi Ujian Peserta (Exam Sessions / Attempts)

**Deskripsi:** Satu percobaan mengerjakan oleh satu peserta pada satu jadwal. Umumnya satu peserta = satu sesi per jadwal, tetapi desain harus mendukung percobaan ulang yang sah (misal: sesi dibatalkan karena gangguan dan diulang atas izin proktor) dengan mencatat sesi mana yang sah / dipakai untuk nilai.

**Kolom-kolom kunci:**
- Rujukan ke peserta jadwal (atau langsung ke jadwal + siswa).
- Nomor percobaan (ke-1, ke-2, dst).
- Waktu mulai sesi dan waktu selesai sesi.
- Status sesi: belum mulai, berjalan, dijeda, selesai, dibatalkan, diblokir.
- Perangkat yang dipakai (rujukan ke tabel perangkat / catatan perangkat).
- Alamat jaringan lokal / alamat IP saat sesi (untuk audit).
- Susunan soal yang diterima peserta (urutan soal setelah pengacakan, dan urutan opsi bila diacak) — disimpan sebagai daftar terstruktur per sesi agar hasil dapat direkonstruksi walau paket diubah belakangan.
- Penanda apakah sesi ini yang dipakai untuk nilai akhir (bila ada percobaan ulang).
- Catatan proktor.

**Mengapa susunan soal per sesi disimpan:** Tanpa ini, bila paket diubah setelah ujian, rekonstruksi "soal apa yang dilihat peserta" menjadi mustahil. Ini penting untuk sengketa nilai dan audit.

#### 1.3.6. Token Ujian

**Deskripsi:** Token dapat disimpan sebagai kolom di tabel jadwal (paling sederhana) atau sebagai tabel sendiri bila token perlu masa berlaku sendiri, token per kelas, atau riwayat pergantian token. Untuk v2, rekomendasi: token sebagai kolom jadwal + tabel riwayat perubahan token (kapan token diubah, oleh siapa, token lama tidak disimpan dalam bentuk yang bisa dibaca ulang bila tidak perlu). Bila sekolah butuh token berbeda per kelas dalam satu jadwal, buat tabel token per jadwal–kelas.

**Kolom-kolom kunci (bila dijadikan tabel):**
- Rujukan ke jadwal.
- Rujukan ke kelas (bila token per kelas).
- Nilai token (disimpan dalam bentuk hash bila memungkinkan; token asli hanya ditampilkan sekali saat dibuat).
- Masa berlaku token (mulai–selesai).
- Status token: aktif, kedaluwarsa, dicabut.

---

### 1.4. Jawaban, Nilai, dan Koreksi

#### 1.4.1. Jawaban Peserta (Participant Answers)

**Deskripsi:** Satu baris = jawaban satu peserta untuk satu soal dalam satu sesi. Ini tabel tersibuk kedua setelah log. Setiap penyimpanan (autosave) memperbarui baris yang sama, bukan menambah baris baru setiap ketikan — kecuali bila riwayat jawaban diperlukan (lihat catatan di bawah).

**Kolom-kolom kunci:**
- Rujukan ke sesi ujian.
- Rujukan ke peserta / siswa (boleh diturunkan dari sesi, tapi disimpan untuk mempercepat rekap).
- Rujukan ke soal.
- Rujukan ke paket / jadwal (untuk mempercepat rekap tanpa join berantai).
- Isi jawaban peserta dalam bentuk terstruktur sesuai tipe soal:
  - Pilihan ganda: opsi yang dipilih.
  - Pilihan ganda kompleks: daftar opsi yang dipilih.
  - Benar-salah: pilihan benar / salah per pernyataan.
  - Isian singkat: teks jawaban.
  - Menjodohkan: pasangan yang dipilih.
  - Esai: teks jawaban.
- Penanda ragu-ragu (flag "tandai ragu-ragu" per soal — warisan CBT-Go).
- Penanda sudah dijawab / masih kosong.
- Skor otomatis (hasil penskoran mesin; kosong untuk esai sebelum dikoreksi).
- Skor manual / koreksi (diisi guru untuk esai / isian yang perlu penilaian manusia).
- Skor akhir per soal (gabungan otomatis + manual, sesuai aturan paket).
- Waktu jawaban pertama kali disimpan dan waktu terakhir diubah.
- Status jawaban: tersimpan, diubah, dikoreksi, final.

**Catatan riwayat jawaban:** Bila sekolah ingin audit "jawaban awal vs jawaban akhir" (misal: untuk mendeteksi perubahan mencurigakan), tambahkan tabel riwayat jawaban terpisah yang hanya menambah baris saat jawaban berubah secara signifikan. Jangan membebani tabel jawaban utama dengan riwayat per ketikan.

#### 1.4.2. Hasil / Nilai Ujian (Exam Results)

**Deskripsi:** Rekap per peserta per jadwal / per sesi: nilai akhir, jumlah benar / salah / kosong, status lulus. Tabel ini adalah hasil agregasi dari tabel jawaban, bukan pengganti tabel jawaban. Nilai dihitung ulang dari jawaban bila perlu, dan hasil perhitungan disimpan di sini agar rekap cepat.

**Kolom-kolom kunci:**
- Rujukan ke sesi ujian (dan / atau ke peserta jadwal).
- Rujukan ke siswa, jadwal, paket.
- Skor total yang diperoleh.
- Skor maksimum yang mungkin.
- Nilai akhir dalam skala yang dipakai sekolah (misal 0–100).
- Jumlah soal: total, dijawab, benar, salah, kosong, ragu-ragu.
- Skor komponen objektif (hasil mesin).
- Skor komponen esai / manual (hasil koreksi manusia).
- Status koreksi: belum dikoreksi, sebagian, selesai.
- Status kelulusan: lulus / tidak lulus / belum ditentukan.
- Status hasil: draf, final, diterbitkan (nilai yang sudah diterbitkan tidak boleh diubah tanpa membuka kunci / mencatat alasan).
- Waktu penghitungan terakhir.
- Penghitung terakhir (sistem / pengguna).

#### 1.4.3. Nilai per Kompetensi (Competency Scores) — opsional, dianjurkan

**Deskripsi:** Nilai per kompetensi per peserta per jadwal. Memungkinkan laporan "siswa lemah di kompetensi X" yang sangat berguna bagi guru.

**Kolom-kolom kunci:**
- Rujukan ke hasil ujian / peserta / jadwal.
- Rujukan ke kompetensi.
- Skor yang diperoleh untuk kompetensi tersebut.
- Skor maksimum untuk kompetensi tersebut.
- Nilai dalam skala yang sama dengan nilai akhir.

#### 1.4.4. Koreksi Manual / Penilaian Esai (Manual Grading)

**Deskripsi:** Jejak koreksi manusia untuk esai / isian. Memisahkan koreksi dari jawaban memudahkan audit: siapa mengoreksi, kapan, berapa skor, dan apa komentarnya.

**Kolom-kolom kunci:**
- Rujukan ke jawaban peserta.
- Rujukan ke pengoreksi (pengguna guru).
- Skor yang diberikan.
- Komentar / umpan balik koreksi.
- Waktu koreksi.
- Status koreksi: draf, final.

#### 1.4.5. Analisis Butir Soal (Item Analysis) — hasil perhitungan, bukan input

**Deskripsi:** Hasil perhitungan berkala per soal per jadwal / per paket: tingkat kesulitan empiris, daya beda, dan fungsi pengecoh (distractor). Disimpan sebagai hasil perhitungan agar laporan cepat dan dapat dibandingkan antar periode.

**Kolom-kolom kunci:**
- Rujukan ke soal dan ke jadwal / paket tempat perhitungan dilakukan.
- Jumlah peserta yang menjawab, jumlah benar.
- Tingkat kesulitan empiris (proporsi benar).
- Daya beda (selisih performa kelompok atas vs bawah — definisi persis dicatat di dokumentasi).
- Distribusi pilihan jawaban (berapa peserta memilih tiap opsi) — untuk menilai pengecoh.
- Waktu perhitungan.
- Catatan / rekomendasi (misal: soal terlalu mudah, pengecoh tidak berfungsi).

---

### 1.5. Keamanan, Perangkat, dan Audit

#### 1.5.1. Perangkat Peserta (Participant Devices)

**Deskripsi:** Catatan perangkat yang dipakai peserta saat ujian (warisan fitur "catat perangkat" CBT-Go). Dipakai untuk mendeteksi pergantian perangkat dan untuk investigasi pelanggaran.

**Kolom-kolom kunci:**
- Rujukan ke pengguna / siswa.
- Rujukan ke sesi ujian.
- Sidik perangkat (fingerprint): gabungan info perangkat yang tidak sensitif (jenis perangkat, sistem operasi, peramban — bukan data pribadi).
- Alamat jaringan lokal / IP.
- Waktu pertama terlihat dan waktu terakhir terlihat.
- Status: diizinkan, mencurigakan, diblokir.

#### 1.5.2. Peristiwa Anti-Curang (Proctoring Events)

**Deskripsi:** Catatan peristiwa yang memicu peringatan anti-curang: pindah tab, keluar fullscreen, dll. Satu baris = satu peristiwa.

**Kolom-kolom kunci:**
- Rujukan ke sesi ujian / peserta.
- Jenis peristiwa (pindah tab, keluar fullscreen, buka jendela baru, dll).
- Waktu peristiwa.
- Tindakan yang diambil sistem: peringatan, penguncian sementara, blokir otomatis.
- Jumlah pelanggaran kumulatif pada sesi tersebut.
- Catatan proktor / keputusan (peringatan diabaikan, blokir dicabut, dll).

#### 1.5.3. Log Aktivitas (Activity Logs)

**Deskripsi:** Jejak audit semua aksi penting: login, pembuatan soal, perubahan jadwal, koreksi nilai, ekspor data, perubahan token, dll. Log tidak boleh bisa diubah oleh pengguna biasa.

**Kolom-kolom kunci:**
- Waktu peristiwa (dengan zona waktu yang jelas).
- Rujukan ke pengguna pelaku (bila ada).
- Jenis aksi (login, buat soal, ubah soal, mulai ujian, submit, koreksi, ekspor, dll).
- Objek yang terkena (jenis objek + pengenal objek, misal: soal nomor sekian, jadwal sekian).
- Hasil aksi (berhasil / gagal + alasan singkat bila gagal).
- Alamat jaringan / perangkat pelaku.
- Detail tambahan terstruktur (misal: nilai lama vs nilai baru untuk perubahan penting).
- Tingkat kepentingan (info, peringatan, kritis).

**Prinsip:** Log hanya ditambah (append-only). Tidak ada fitur ubah / hapus log dari aplikasi. Penghapusan log hanya via kebijakan retensi yang terdokumentasi (lihat Bagian 6).

#### 1.5.4. Pengaturan Sistem (Settings)

**Deskripsi:** Konfigurasi aplikasi yang dapat diubah admin tanpa mengubah kode: nama sekolah, tahun ajaran aktif, batas upload, pesan pengumuman, dsb.

**Kolom-kolom kunci:**
- Kunci pengaturan (nama unik, stabil).
- Nilai pengaturan.
- Kelompok pengaturan.
- Deskripsi / batas nilai yang diizinkan.
- Kapan terakhir diubah dan oleh siapa.

#### 1.5.5. Ruangan / Event (Rooms / Events) — opsional

**Deskripsi:** Bila sekolah ingin mengelola ruangan / lab dan event ujian (fitur yang ada di Extraordinary CBT), tambahkan tabel ruangan dan event. Untuk v2 awal, ruangan cukup sebagai teks di jadwal; jadikan tabel bila ruangan perlu dijadwalkan (menghindari bentrok ruangan).

**Kolom-kolom kunci (bila dijadikan tabel):**
- Nama ruangan / lab.
- Kapasitas.
- Lokasi / gedung.
- Status (aktif / perawatan / nonaktif).

---

## 2. Relasi Antar Tabel & Aturan Hapus / Ubah

### 2.1. Peta Relasi (dalam kata-kata)

**Pengguna & peran:**
- Satu pengguna dapat memiliki banyak peran (many-to-many) melalui tabel Relasi Pengguna–Peran.
- Satu peran dapat dimiliki banyak pengguna.
- Satu peran dapat memiliki banyak izin (one-to-many dari peran ke izin).

**Kelembagaan:**
- Satu jurusan memiliki banyak kelas (one-to-many).
- Satu kelas memiliki banyak siswa (one-to-many dari kelas ke profil siswa).
- Satu guru dapat mengampu banyak mapel, satu mapel dapat diampu banyak guru (many-to-many) melalui Penugasan Guru–Mapel.
- Satu siswa memiliki tepat satu profil siswa dan satu akun pengguna (one-to-one dari pengguna ke profil siswa).

**Bank soal:**
- Satu bank soal milik satu mapel (many-to-one dari bank ke mapel).
- Satu bank soal memiliki banyak soal (one-to-many).
- Satu soal memiliki banyak opsi jawaban (one-to-many), kecuali soal esai / isian yang tidak memiliki opsi.
- Satu soal dapat mengukur banyak kompetensi (many-to-many) bila fitur nilai per kompetensi dipakai.
- Satu soal dapat memiliki banyak file media (one-to-many).

**Paket & jadwal:**
- Satu paket ujian milik satu mapel (many-to-one).
- Satu paket ujian berisi banyak soal (many-to-many) melalui Susunan Soal Paket.
- Satu soal dapat masuk banyak paket (many-to-many).
- Satu paket dapat dijadwalkan banyak kali (one-to-many dari paket ke jadwal).
- Satu jadwal mengundang banyak siswa (many-to-many) melalui Peserta Jadwal.
- Satu siswa dapat mengikuti banyak jadwal (many-to-many).
- Satu peserta jadwal dapat memiliki satu atau lebih sesi (one-to-many), dengan satu sesi yang ditandai sah untuk nilai.

**Jawaban & nilai:**
- Satu sesi memiliki banyak jawaban (one-to-many: satu jawaban per soal per sesi).
- Satu jawaban milik satu soal dan satu peserta (many-to-one ke soal, many-to-one ke sesi).
- Satu sesi menghasilkan satu hasil / nilai (one-to-one dari sesi ke hasil).
- Satu jawaban esai dapat memiliki satu atau lebih catatan koreksi manual (one-to-many), dengan koreksi terakhir yang final dipakai.
- Satu hasil dapat memiliki banyak nilai per kompetensi (one-to-many).

**Audit & keamanan:**
- Satu sesi dapat memiliki banyak peristiwa anti-curang (one-to-many).
- Satu pengguna dapat memiliki banyak catatan perangkat (one-to-many).
- Satu pengguna dapat menghasilkan banyak log aktivitas (one-to-many).
- Satu jadwal dapat memiliki banyak peristiwa, banyak peserta, banyak sesi (one-to-many ke masing-masing).

### 2.2. Aturan Hapus / Ubah (Delete / Update Rules)

Prinsip utama: **data yang sudah dipakai dalam ujian adalah arsip, bukan sampah.** Menghapusnya merusak nilai, audit, dan kemampuan rekonstruksi.

| Situasi | Aturan yang dianjurkan |
|---|---|
| Menghapus jurusan / kelas / mapel yang masih dipakai | **Ditolak.** Sistem menolak penghapusan bila masih ada kelas / siswa / bank soal / paket yang merujuk. Data master yang tidak lagi dipakai ditandai nonaktif / arsip, bukan dihapus. |
| Menghapus pengguna (guru / siswa) | **Tidak dihapus fisik.** Akun ditandai nonaktif / arsip. Riwayat ujian, jawaban, dan log yang merujuk pengguna tersebut tetap utuh. Nama pengguna tidak dipakai ulang untuk orang lain. |
| Menghapus soal yang belum pernah dipakai ujian | **Diizinkan** oleh peran berwenang, dengan konfirmasi. Soal masuk arsip (soft delete) agar dapat dipulihkan. Opsi jawaban soal tersebut ikut diarsipkan. |
| Menghapus soal yang sudah dipakai dalam paket / jadwal / ujian | **Ditolak / dibatasi.** Soal tidak boleh dihapus. Yang diizinkan: (a) mengeluarkan soal dari paket yang belum berjalan, (b) menandai soal nonaktif agar tidak dipakai paket baru, (c) membuat versi baru soal untuk perbaikan. Hasil ujian lama tetap merujuk soal versi lama. |
| Mengubah isi soal / kunci jawaban setelah soal dipakai ujian | **Ditolak bila ujian sudah berjalan / selesai.** Perubahan isi soal yang sudah dipakai harus membuat versi baru, bukan menimpa. Perubahan kunci jawaban setelah ujian selesai hanya boleh dilakukan oleh peran super admin dengan alasan tertulis yang dicatat di log, dan nilai yang terdampak harus dihitung ulang secara eksplisit (bukan diam-diam). |
| Menghapus paket ujian | **Ditolak bila paket sudah memiliki jadwal / ujian.** Paket yang belum dijadwalkan boleh diarsipkan. Paket yang sudah dijadwalkan hanya boleh dinonaktifkan untuk jadwal baru. |
| Menghapus jadwal ujian | **Ditolak bila jadwal sudah memiliki peserta / sesi / jawaban.** Jadwal yang belum dimulai boleh dibatalkan (status dibatalkan, bukan dihapus). Jadwal yang sudah berjalan hanya boleh dihentikan / diselesaikan dengan alasan. |
| Menghapus jawaban peserta | **Tidak pernah dihapus dari aplikasi.** Jawaban adalah bukti ujian. Koreksi dilakukan dengan memperbarui skor / menambah catatan koreksi, bukan menghapus jawaban. |
| Menghapus nilai / hasil ujian | **Tidak pernah dihapus dari aplikasi.** Nilai yang salah diperbaiki lewat koreksi tercatat + hitung ulang, dengan alasan dan pelaku tercatat di log. |
| Mengubah nilai yang sudah diterbitkan | **Hanya dengan membuka kunci + alasan.** Sistem mencatat nilai lama, nilai baru, siapa, kapan, dan mengapa. |
| Menghapus log aktivitas | **Tidak ada fitur hapus log di aplikasi.** Penghapusan hanya melalui kebijakan retensi (lihat Bagian 6) yang dijalankan admin sistem, bukan dari menu aplikasi. |
| Menghapus file media soal | **Ditolak bila file masih dipakai soal aktif.** File yang tidak dipakai lagi diarsipkan; file fisik dihapus hanya setelah masa tenggang dan setelah dipastikan tidak ada soal yang merujuk. |
| Mengubah token ujian | **Diizinkan** oleh proktor / admin, dengan riwayat perubahan dicatat. Token lama langsung tidak berlaku. Peserta yang sudah masuk dengan token lama tidak dikeluarkan paksa kecuali diputuskan proktor. |

### 2.3. Aturan Konsistensi Tambahan

- **Satu jawaban per soal per sesi:** sistem menolak jawaban ganda untuk pasangan sesi + soal yang sama. Penyimpanan ulang memperbarui baris yang sama.
- **Satu sesi sah per peserta per jadwal:** bila ada percobaan ulang, hanya satu sesi yang ditandai sah untuk nilai. Sesi lain ditandai dibatalkan / tidak dipakai, tetapi tetap tersimpan untuk audit.
- **Susunan soal per sesi tidak berubah setelah sesi dimulai:** walau paket diubah di tengah ujian, peserta yang sudah mulai tetap mengerjakan susunan yang ia terima di awal.
- **Nilai = turunan dari jawaban:** nilai tidak boleh diubah langsung tanpa mengubah / mencatat dasarnya (jawaban / koreksi). Bila nilai diubah manual, alasannya wajib dicatat.
- **Kelas siswa pada hasil ujian dibekukan:** bila siswa pindah kelas setelah ujian, hasil ujian lama tetap mencatat kelas saat ujian berlangsung (disalin ke hasil / peserta jadwal), bukan mengikuti kelas terbaru.

---

## 3. Indeks yang Dibutuhkan untuk Skala

Indeks adalah penanda di database yang mempercepat pencarian pada kolom tertentu, dengan biaya: penulisan menjadi sedikit lebih lambat dan penyimpanan bertambah. Untuk CBT SMK v2, indeks diprioritaskan pada jalur panas (login, ambil soal, simpan jawaban, rekap), bukan pada semua kolom.

**Prinsip umum:** Indeks dibuat untuk kolom yang dipakai sebagai kunci pencarian, penyaringan, pengurutan, atau penggabungan antar tabel dalam kueri yang sering dijalankan. Kolom yang jarang dicari tidak perlu indeks.

### 3.1. Indeks untuk Login & Akun

| Kebutuhan | Kolom yang perlu indeks | Alasan |
|---|---|---|
| Login siswa / guru | Nama pengguna pada tabel pengguna | Setiap login mencari akun berdasarkan nama pengguna. Tanpa indeks, database memindai seluruh tabel pengguna setiap login — lambat saat ratusan siswa login bersamaan. |
| Pencarian akun & pencegahan ganda | Alamat email pada tabel pengguna (bila dipakai) | Mempercepat pengecekan email ganda dan pemulihan akun. |
| Penentuan peran setelah login | Rujukan pengguna pada tabel Relasi Pengguna–Peran | Setelah login, sistem langsung mencari peran pengguna tersebut. Indeks ini membuat penentuan menu / izin instan. |
| Pengecekan akun demo / arsip | Penanda status + penanda arsip pada tabel pengguna (indeks gabungan dengan nama pengguna bila perlu) | Mempercepat penyaringan "hanya akun aktif" saat login massal. |
| Profil siswa dari akun | Rujukan pengguna pada tabel Profil Siswa | Setelah login siswa, sistem perlu profil siswa (kelas, NIS) untuk menentukan jadwal yang berhak diikuti. |

### 3.2. Indeks untuk Pengambilan Soal Ujian

| Kebutuhan | Kolom yang perlu indeks | Alasan |
|---|---|---|
| Mengambil daftar soal dalam paket | Rujukan paket pada tabel Susunan Soal Paket + urutan soal | Saat sesi dimulai, sistem mengambil semua soal paket dalam urutan yang benar. Indeks ini membuat pengambilan paket berukuran puluhan soal menjadi cepat. |
| Mengambil opsi jawaban per soal | Rujukan soal pada tabel Opsi Jawaban + urutan tampil | Setiap soal pilihan perlu opsi-opsinya. Tanpa indeks, setiap soal memicu pemindaian seluruh tabel opsi. |
| Mengambil soal aktif per bank | Rujukan bank + status soal pada tabel Soal | Guru membuka bank soal dan menyaring hanya soal aktif. Indeks gabungan mempercepat penyaringan ini. |
| Mengambil soal per mapel / kompetensi | Rujukan mapel / kompetensi pada tabel Soal atau Bank Soal | Mempercepat pencarian soal untuk merakit paket baru dan untuk laporan per kompetensi. |
| Mengambil susunan soal sesi | Rujukan sesi pada tabel Jawaban / Sesi | Saat peserta melanjutkan ujian (refresh / reconnect), sistem harus cepat menemukan susunan soal sesi tersebut. |
| Mengambil jadwal yang berhak diikuti siswa | Rujukan siswa pada tabel Peserta Jadwal + rujukan jadwal | Dashboard siswa menampilkan jadwal ujian yang berhak ia ikuti. Indeks ini menghindari pemindaian seluruh peserta. |
| Validasi token saat masuk ujian | Token pada tabel Jadwal (atau tabel Token) + masa berlaku | Peserta memasukkan token; sistem harus menemukan jadwal yang cocok dengan cepat dan memeriksa masa berlaku. |
| Mengambil sesi yang sedang berjalan | Rujukan peserta + status sesi pada tabel Sesi | Saat peserta kembali ke halaman ujian, sistem mencari sesi berjalan miliknya. |

### 3.3. Indeks untuk Penyimpanan Jawaban (Jalur Tersibuk)

| Kebutuhan | Kolom yang perlu indeks | Alasan |
|---|---|---|
| Autosave jawaban | Pasangan sesi + soal pada tabel Jawaban Peserta (indeks unik gabungan) | Setiap autosave mencari baris jawaban untuk pasangan sesi + soal. Indeks unik ini membuat autosave menjadi pencarian langsung, sekaligus menegakkan aturan "satu jawaban per soal per sesi". |
| Mengambil semua jawaban satu sesi | Rujukan sesi pada tabel Jawaban Peserta | Saat peserta membuka halaman ujian atau saat submit, sistem mengambil seluruh jawabannya. Indeks ini membuat pengambilan satu sesi cepat walau tabel jawaban sudah berisi ratusan ribu baris. |
| Menghitung nilai satu sesi | Rujukan sesi + status jawaban pada tabel Jawaban | Perhitungan nilai memindai jawaban satu sesi. Indeks membatasi pemindaian hanya pada sesi tersebut. |
| Koreksi esai per jadwal | Rujukan jadwal + tipe soal / status koreksi pada tabel Jawaban (via sesi) | Guru membuka daftar esai yang belum dikoreksi untuk satu jadwal. Indeks mempercepat penyaringan ini. |

### 3.4. Indeks untuk Rekap Nilai & Laporan

| Kebutuhan | Kolom yang perlu indeks | Alasan |
|---|---|---|
| Rekap nilai per jadwal | Rujukan jadwal pada tabel Hasil / Nilai | Rekap utama guru / proktor adalah "semua nilai untuk jadwal X". Indeks ini membuat rekap satu jadwal cepat. |
| Rekap nilai per kelas / per siswa | Rujukan kelas + rujukan siswa pada tabel Hasil (indeks gabungan) | Laporan per kelas dan riwayat nilai satu siswa adalah kueri yang sering dijalankan. Indeks gabungan mempercepat keduanya. |
| Rekap nilai per paket / mapel | Rujukan paket / mapel pada tabel Hasil | Mempercepat laporan lintas jadwal untuk satu paket / mapel. |
| Nilai per kompetensi | Rujukan hasil + rujukan kompetensi pada tabel Nilai per Kompetensi | Laporan "nilai per kompetensi" mengambil baris per kompetensi untuk satu hasil. |
| Log aktivitas per pengguna / per objek | Rujukan pengguna + waktu pada tabel Log; jenis objek + pengenal objek pada tabel Log | Audit sering mencari "apa yang dilakukan pengguna X" atau "apa yang terjadi pada objek Y". Indeks ini membuat audit cepat tanpa memindai seluruh log. |
| Peristiwa anti-curang per sesi | Rujukan sesi + waktu pada tabel Peristiwa Anti-Curang | Monitor proktor menampilkan peristiwa terbaru per sesi. Indeks membuat pemantauan real-time ringan. |

### 3.5. Indeks yang Sebaiknya Tidak Berlebihan

- Jangan membuat indeks pada setiap kolom "demi keamanan". Setiap indeks memperlambat penulisan (insert / update) — dan tabel jawaban adalah tabel yang paling sering ditulis.
- Kolom teks panjang (isi soal, teks jawaban esai) tidak diindeks penuh; pencarian teks soal bila diperlukan memakai pencarian teks khusus database, bukan indeks biasa.
- Kolom status yang nilainya hanya 2–3 macam (misal aktif / nonaktif) tidak efektif bila diindeks sendirian; gabungkan dengan kolom lain (misal: bank + status) agar indeks benar-benar selektif.
- Setelah sistem berjalan, indeks perlu ditinjau berkala berdasarkan pola kueri nyata (kueri mana yang lambat), bukan ditambah membabi buta.

### 3.6. Catatan Operasional Indeks

- Indeks dibuat bersamaan dengan migrasi skema, bukan setelah masalah muncul.
- Pembuatan indeks pada tabel yang sudah besar (misal jawaban setelah ribuan ujian) harus dilakukan di luar jam ujian agar tidak mengunci tabel.
- Untuk tabel jawaban yang akan membesar cepat, pertimbangkan partisi di masa depan (misal per tahun ajaran / per semester) — bukan untuk v2 awal, tetapi desain harus tidak menghalangi partisi (kolom waktu / tahun ajaran tersedia pada baris jawaban atau sesi).

---

## 4. Strategi Migrasi & Seed Data Awal

### 4.1. Prinsip Migrasi

- **Migrasi berversi dan berurutan.** Setiap perubahan skema adalah satu langkah migrasi bernomor (misal: langkah 001, 002, dst) yang dapat dijalankan maju dan, bila memungkinkan, dibatalkan. Urutan migrasi tidak boleh diubah setelah diterapkan di lingkungan manapun.
- **Migrasi dijalankan otomatis saat aplikasi mulai** (mengikuti pola CBT-Go: migrasi + seed otomatis), tetapi dengan pengaman:
  - Aplikasi mencatat versi migrasi terakhir yang sudah diterapkan di database.
  - Migrasi yang sudah diterapkan tidak dijalankan ulang.
  - Kegagalan migrasi menghentikan startup dengan pesan yang jelas (bukan menjalankan aplikasi dengan skema setengah jadi).
  - Sebelum migrasi yang mengubah data besar, sistem membuat backup otomatis (lihat Bagian 5).
- **Pisahkan migrasi skema dari migrasi data.**
  - Migrasi skema: menambah / mengubah struktur tabel dan indeks.
  - Migrasi data: memindahkan / membersihkan data lama agar sesuai struktur baru. Migrasi data yang besar dijalankan di luar jam ujian dan dapat dilanjutkan bila terputus.
- **Tidak ada perubahan skema diam-diam.** Setiap migrasi mencatat: nomor, deskripsi, tanggal, siapa yang membuat, dan apakah reversibel.
- **Lingkungan bertahap:** migrasi diuji di lingkungan pengembangan → lingkungan staging / salinan data sekolah → baru produksi. Migrasi tidak diuji langsung di produksi.
- **Jendela perubahan:** migrasi yang berpotensi mengunci tabel (tambah kolom dengan nilai default pada tabel besar, tambah indeks, ubah tipe kolom) dijadwalkan di luar jam ujian dan diumumkan ke proktor.

### 4.2. Urutan Migrasi yang Dianjurkan

Urutan ini mengikuti ketergantungan antar tabel (tabel induk dulu, baru tabel yang merujuk):

1. **Tahap fondasi:** peran, izin per peran, pengguna, relasi pengguna–peran, pengaturan sistem.
2. **Tahap master kelembagaan:** jurusan, kelas, mapel, profil siswa, profil guru, penugasan guru–mapel, kompetensi.
3. **Tahap bank soal:** bank soal, soal, opsi jawaban, pasangan menjodohkan (bila terpisah), soal–kompetensi, file media.
4. **Tahap paket & jadwal:** paket ujian, susunan soal paket, jadwal ujian, peserta jadwal, token (bila terpisah), sesi ujian.
5. **Tahap jawaban & nilai:** jawaban peserta, hasil / nilai, nilai per kompetensi, koreksi manual, analisis butir soal.
6. **Tahap audit & keamanan:** perangkat peserta, peristiwa anti-curang, log aktivitas.
7. **Tahap indeks & penguat:** indeks jalur panas (lihat Bagian 3), batasan keunikan (misal: satu jawaban per sesi + soal, nama pengguna unik).

### 4.3. Seed Data Awal (Data Contoh / Demo)

Seed adalah data awal yang dimasukkan otomatis pada instalasi baru agar sistem langsung dapat dipakai dan didemokan. Seed harus jelas dibedakan dari data nyata (penanda akun demo / data contoh).

#### 4.3.1. Akun Demo

| Akun | Peran | Kegunaan |
|---|---|---|
| Akun super admin demo | Super Admin | Mengelola seluruh sistem, membuat data master, melihat semua laporan. Kata sandi awal wajib diganti saat pertama login. |
| Akun guru demo (1–2 akun) | Guru | Membuat bank soal, merakit paket, mengoreksi esai, melihat rekap. |
| Akun proktor demo | Proktor | Membuat jadwal, mengelola token, memantau peserta, menangani blokir. |
| Akun siswa demo per kelas contoh (misal 3–5 siswa) | Siswa | Mengikuti ujian contoh, menguji alur peserta end-to-end. |
| Akun pengawas / penguji UKK demo (opsional) | Pengawas / Penguji | Menguji alur pengawasan / UKK bila diaktifkan. |

**Aturan akun demo:**
- Akun demo memakai nama pengguna yang jelas demo (misal mengandung kata "demo") agar tidak tertukar dengan akun nyata.
- Kata sandi awal akun demo kuat dan terdokumentasi di panduan instalasi, bukan kata sandi lemah yang mudah ditebak.
- Sistem memaksa ganti kata sandi super admin pada login pertama.
- Tersedia perintah / menu "bersihkan data demo" yang menghapus akun demo + data ujian contoh, tanpa menyentuh data nyata. Pembersihan ini hanya diizinkan sebelum data nyata masuk, dan meminta konfirmasi ganda.

#### 4.3.2. Data Master Contoh

- **Jurusan contoh:** 2 jurusan (misal: RPL dan TKJ — sesuaikan dengan sekolah).
- **Kelas contoh:** 2–3 kelas (misal: XI RPL 1, XI TKJ 1).
- **Mapel contoh:** 2–3 mapel (misal: Pemrograman Web, Jaringan Komputer, Matematika).
- **Bank soal contoh:** 1 bank soal per mapel contoh, masing-masing berisi 10–20 soal campuran tipe (pilihan ganda, benar-salah, isian singkat, pilihan ganda kompleks, esai).
- **Paket ujian contoh:** 1 paket "STS Pemrograman Web — Contoh" dengan durasi singkat (misal 30 menit) dan token contoh.
- **Jadwal ujian contoh:** 1 jadwal untuk kelas contoh, status draf (tidak langsung berjalan), agar admin dapat melihat alur tanpa mengganggu ujian nyata.
- **Pengaturan awal:** nama sekolah (placeholder yang wajib diubah), tahun ajaran aktif, batas upload file, pesan pengumuman contoh.

#### 4.3.3. Seed untuk Lingkungan Pengembangan vs Produksi

- **Pengembangan:** seed lengkap (semua akun demo + data master contoh + soal contoh + jadwal contoh) agar pengembang dapat menguji tanpa membuat data manual.
- **Produksi sekolah:** seed minimal — hanya peran + izin + akun super admin + pengaturan dasar. Data master sekolah diimpor dari data nyata sekolah (lihat 4.4), bukan dari data contoh. Data contoh hanya dimasukkan bila sekolah meminta mode demo.
- **Seed harus idempotent:** menjalankan seed dua kali tidak boleh membuat akun ganda / data ganda. Seed memeriksa keberadaan data sebelum menambah.

### 4.4. Impor Data Master Nyata

- Sekolah umumnya sudah memiliki data siswa / kelas / mapel dalam bentuk spreadsheet atau sistem lain. Sediakan impor berkas CSV untuk: jurusan, kelas, mapel, siswa, guru.
- Impor harus memiliki: pratinjau sebelum disimpan, laporan baris yang gagal + alasannya, dan pembatalan bila sebagian besar baris gagal (tidak menyimpan setengah-setengah).
- Impor siswa harus memvalidasi: NIS unik, kelas tujuan ada, nama pengguna tidak bentrok dengan akun yang sudah ada.
- Impor soal dari CSV (warisan CBT-Go) dipertahankan untuk bank soal, dengan format kolom yang terdokumentasi dan contoh berkas.

### 4.5. Migrasi dari CBT-Go Lama (bila v2 menggantikan instalasi lama)

- Jangan menimpa database lama. Buat database baru untuk v2, lalu migrasikan data yang masih relevan (pengguna, data master, bank soal) dengan skrip migrasi terpisah.
- Data ujian lama (jawaban, nilai) dimigrasikan hanya bila formatnya kompatibel; bila tidak, arsipkan database lama sebagai baca-saja (read-only) untuk keperluan audit dan rekap historis, dan mulai v2 dengan data ujian kosong.
- Uji migrasi pada salinan database lama terlebih dahulu, bandingkan jumlah baris sebelum–sesudah per tabel, dan verifikasi contoh nilai lama tetap sama setelah migrasi.

---

## 5. Backup & Restore

### 5.1. Apa yang Di-backup

Backup harus mencakup **semua** yang diperlukan untuk membangun ulang sistem dari nol di server lain:

1. **Database PostgreSQL (isi penuh):** semua tabel — pengguna, master, bank soal, paket, jadwal, sesi, jawaban, nilai, koreksi, log, pengaturan. Ini adalah prioritas utama.
2. **File yang diunggah:** gambar / media soal, file lampiran, dan file lain di penyimpanan file lokal. Database hanya menyimpan metadata file; file fisiknya harus ikut di-backup, atau soal akan kehilangan medianya setelah restore.
3. **Konfigurasi aplikasi:** berkas konfigurasi server (koneksi database, port, pengaturan server lokal), bukan kata sandi dalam bentuk teks biasa — kata sandi / kunci disimpan terpisah secara aman (lihat Bagian 6).
4. **Catatan versi:** versi aplikasi + versi migrasi database pada saat backup diambil. Tanpa ini, restore ke versi aplikasi yang berbeda bisa gagal.

**Yang tidak perlu di-backup sebagai data:** hasil build aplikasi (binary Go dapat dibangun ulang), cache sementara, log aplikasi mentah yang sudah masuk tabel log (bila diduplikasi), dan berkas sementara.

### 5.2. Cara Backup PostgreSQL (pg_dump)

- **Perkakas utama:** `pg_dump` (perkakas bawaan PostgreSQL) untuk membuat salinan database dalam bentuk berkas backup.
- **Bentuk backup yang dianjurkan untuk v2:**
  - **Backup logis penuh (via pg_dump):** salinan seluruh database dalam satu berkas. Mudah dipindahkan, mudah di-restore ke server lain, cocok untuk skala sekolah dan untuk backup harian.
  - **Backup berkas file (file storage):** salinan direktori file yang diunggah, diambil pada waktu yang sama dengan backup database (atau sedekat mungkin), agar database dan file konsisten.
- **Prinsip backup:**
  - Backup diambil tanpa menghentikan layanan ujian bila memungkinkan (backup tidak boleh mengganggu ujian yang sedang berjalan).
  - Backup tidak mengambil data dari database yang sedang dalam migrasi.
  - Setiap berkas backup diberi nama yang memuat tanggal–waktu pengambilan dan versi aplikasi / migrasi, agar mudah menemukan backup yang tepat.
  - Setelah backup dibuat, sistem memverifikasi berkas backup: ukuran tidak nol, dapat dibaca, dan checksum (sidik berkas) dicatat untuk mendeteksi kerusakan / perubahan.
  - Berkas backup langsung dienkripsi (lihat Bagian 6) sebelum disimpan / dipindahkan.

### 5.3. Jadwal Backup

| Jenis backup | Frekuensi yang dianjurkan | Keterangan |
|---|---|---|
| Backup database penuh | **Setiap hari**, di luar jam ujian (misal dini hari) | Backup utama harian. Di musim ujian, tambahkan backup kedua di tengah hari setelah sesi ujian selesai. |
| Backup database sebelum perubahan berisiko | **Setiap kali** sebelum migrasi, impor massal, atau pembersihan data | Backup sekali pakai (pre-change) yang diberi label perubahan yang akan dilakukan. |
| Backup file yang diunggah | **Setiap hari**, bersamaan dengan backup database | File soal jarang berubah, tetapi harus ikut terjadwal agar tidak tertinggal. |
| Salinan luar lokasi (off-site) | **Setiap minggu** (atau setiap hari bila memungkinkan) | Salinan backup disimpan di lokasi berbeda dari server ujian (misal: hard disk eksternal terenkripsi yang disimpan di ruang berbeda, atau penyimpanan cloud sekolah bila ada). Melindungi dari kerusakan fisik server / kebakaran / pencurian. |
| Arsip jangka panjang | **Setiap akhir semester / tahun ajaran** | Satu backup yang disimpan lebih lama sebagai arsip tahun ajaran, terpisah dari rotasi harian. |

**Rotasi backup (agar disk tidak penuh):**
- Backup harian disimpan 14–30 hari terakhir.
- Backup mingguan disimpan 8–12 minggu terakhir.
- Backup arsip semester disimpan hingga kebijakan retensi sekolah mengizinkan penghapusannya (lihat Bagian 6).
- Penghapusan backup lama dilakukan otomatis sesuai rotasi, dengan log penghapusan.

### 5.4. Cara Restore

Restore adalah proses membangun ulang sistem dari berkas backup. Langkah-langkahnya dalam urutan yang benar:

1. **Hentikan layanan aplikasi** agar tidak ada yang menulis ke database selama restore.
2. **Siapkan database tujuan:** database kosong baru (atau database yang akan ditimpa, setelah dipastikan backup terbarunya sudah aman).
3. **Kembalikan database** dari berkas backup menggunakan perkakas restore bawaan PostgreSQL (`pg_restore` untuk format arsip, atau pemulihan via `psql` untuk format teks — sesuai format backup yang dipakai). Pastikan versi PostgreSQL tujuan kompatibel dengan backup.
4. **Kembalikan file yang diunggah** ke direktori penyimpanan file yang sama (atau direktori baru bila konfigurasi diubah), dan periksa izin akses berkas.
5. **Kembalikan konfigurasi** (koneksi database, port, direktori file) dan sesuaikan bila server tujuan berbeda (misal alamat / port berbeda).
6. **Periksa versi migrasi** yang tercatat di database hasil restore; jalankan migrasi yang tertinggal bila versi aplikasi tujuan lebih baru — atau, bila versi aplikasi tujuan lebih lama dari backup, jangan memaksakan restore (kembalikan ke versi aplikasi yang sama dengan backup).
7. **Jalankan pemeriksaan kesehatan:** aplikasi dapat start, login admin berhasil, satu bank soal dapat dibuka, satu jadwal contoh dapat dibaca, dan satu hasil ujian contoh cocok dengan sebelum backup.
8. **Baru buka layanan untuk pengguna** setelah pemeriksaan kesehatan lulus.

### 5.5. Tes Restore (Wajib, Bukan Opsional)

Backup yang belum pernah diuji restore dianggap **tidak ada**.

- **Tes restore terjadwal:** minimal **satu kali per bulan**, dan **satu kali sebelum musim ujian besar** (STS / SAS / USBK).
- **Cara tes:** restore backup terbaru ke server / komputer terpisah (bukan menimpa produksi), lalu jalankan pemeriksaan kesehatan (lihat 5.4 langkah 7) + verifikasi jumlah baris pada tabel-tabel kunci (pengguna, soal, jadwal, jawaban, nilai) cocok dengan sumber.
- **Catat hasil tes:** tanggal tes, backup mana yang diuji, siapa yang menguji, hasil pemeriksaan, dan masalah yang ditemukan. Bila tes gagal, backup dianggap tidak dapat dipercaya dan jadwal backup diperbaiki sebelum ujian berikutnya.
- **Latihan restore untuk operator:** minimal satu operator selain admin utama pernah melakukan restore, agar tidak ada ketergantungan pada satu orang saat darurat.

### 5.6. Skenario Darurat

- **Database rusak / server mati di tengah ujian:** hentikan ujian, pulihkan dari backup terbaru + file yang diunggah, verifikasi sesi yang sedang berjalan (sesi yang belum tersimpan di backup terakhir mungkin perlu diulang atas keputusan proktor — peserta harus diberi tahu secara transparan).
- **Salah hapus data (misal bank soal terhapus):** jangan panik dan jangan menimpa backup terbaru dengan keadaan yang salah. Restore ke database terpisah, ambil data yang hilang, lalu gabungkan kembali dengan pengawasan admin.
- **Server hilang / rusak fisik:** bangun server pengganti, install PostgreSQL + aplikasi versi yang sama, restore backup luar lokasi (off-site) terbaru, lalu verifikasi.

---

## 6. Keamanan Data

### 6.1. Enkripsi Koneksi (Data dalam Perjalanan)

- **Koneksi aplikasi → database:** aktifkan koneksi terenkripsi (TLS / SSL) antara aplikasi Go dan PostgreSQL, walau keduanya di server yang sama. Ini mencegah penyadapan kredensial dan data di jaringan lokal.
- **Koneksi browser → aplikasi:** seluruh akses aplikasi memakai HTTPS (atau HTTP hanya di jaringan lokal yang benar-benar tertutup — dan tetap dianjurkan HTTPS walau lokal). Sertifikat dapat berupa sertifikat internal sekolah / sertifikat lokal yang dipercaya perangkat sekolah.
- **Kata sandi database** tidak ditulis di kode dan tidak ditulis di berkas konfigurasi dalam bentuk teks biasa yang dibagikan sembarangan. Kata sandi disimpan di mekanisme rahasia yang disediakan lingkungan server (environment / vault / berkas konfigurasi dengan izin akses ketat, hanya dapat dibaca pemilik layanan).
- **Kredensial berbeda per lingkungan:** kata sandi database pengembangan, staging, dan produksi berbeda. Kredensial produksi tidak dipakai di pengembangan.

### 6.2. Enkripsi Data Tersimpan (Data at Rest)

- **Kata sandi pengguna:** disimpan hanya dalam bentuk hash dengan algoritma hash kata sandi modern + garam (salt) per pengguna. Hash tidak dapat dibalik menjadi kata sandi asli. Algoritma hash lama (misal MD5 / SHA-1 tanpa garam) tidak boleh dipakai untuk kata sandi.
- **Token ujian:** bila token disimpan di database, simpan dalam bentuk hash bila memungkinkan (seperti kata sandi), agar kebocoran database tidak langsung membocorkan token yang masih berlaku. Token asli hanya ditampilkan sekali saat dibuat.
- **Data sensitif lain** (misal: jawaban esai yang memuat data pribadi, catatan koreksi): tidak perlu enkripsi per kolom pada v2 awal, tetapi aksesnya dibatasi lewat izin peran + log audit. Bila sekolah menangani data yang sangat sensitif, enkripsi disk / volume server (full-disk encryption) adalah lapisan pertahanan dasar.
- **Backup terenkripsi:** semua berkas backup (database + file) dienkripsi sebelum disimpan / dipindahkan / diunggah ke luar lokasi. Kunci enkripsi backup disimpan terpisah dari berkas backup (jangan simpan kunci di folder yang sama dengan backup). Tanpa kunci, backup yang dicuri tidak dapat dibaca.

### 6.3. Kontrol Akses & Prinsip Hak Minimum

- Setiap akun hanya memiliki izin yang diperlukan perannya (lihat Bagian 1.1.4). Guru tidak dapat mengubah pengaturan sistem; siswa tidak dapat melihat bank soal / kunci jawaban; proktor tidak dapat mengubah nilai.
- Akses langsung ke database (misal via alat database) hanya untuk admin sistem, tidak untuk pengguna aplikasi sehari-hari. Akses langsung dicatat.
- Akun bersama (satu akun dipakai banyak orang) dilarang. Setiap guru / proktor / siswa memiliki akun sendiri agar log audit bermakna.
- Akun yang tidak aktif lama (misal: siswa lulus) dinonaktifkan, bukan dibiarkan aktif.
- Upaya login gagal dibatasi (misal: jeda / kunci sementara setelah beberapa kali gagal) untuk menghambat tebakan kata sandi. Upaya login gagal dicatat di log.

### 6.4. Retensi Log & Data

| Jenis data | Retensi yang dianjurkan | Keterangan |
|---|---|---|
| Log aktivitas (audit) | **Minimal 1 tahun ajaran**, idealnya 2 tahun ajaran | Log adalah bukti audit. Jangan hapus log sebelum masa retensi habis. Penghapusan log setelah masa retensi dilakukan oleh admin sistem dengan pencatatan. |
| Peristiwa anti-curang | **Selama hasil ujian terkait masih disimpan** | Peristiwa ini adalah bukti pendukung nilai / keputusan blokir. Dihapus bersamaan dengan arsip hasil ujian, bukan lebih awal. |
| Jawaban & nilai ujian | **Sesuai kebijakan sekolah / dinas** (minimal hingga akhir tahun ajaran + masa sanggah nilai) | Jangan hapus jawaban / nilai selama masa sanggah masih berjalan. Arsipkan, bukan hapus. |
| Data siswa yang sudah lulus / pindah | **Diarsipkan, bukan dihapus langsung** | Akun dinonaktifkan; data ujian historis tetap tersimpan untuk rekap / legal. Penghapusan permanen hanya sesuai kebijakan privasi sekolah dan setelah masa retensi. |
| Backup harian / mingguan | Sesuai rotasi di Bagian 5.3 (14–30 hari / 8–12 minggu) | Backup arsip semester disimpan lebih lama. |
| Log aplikasi teknis (bukan audit) | **30–90 hari** | Log teknis (error, performa) boleh dirotasi lebih cepat daripada log audit. |

**Prinsip retensi:** retensi ditetapkan tertulis oleh sekolah (bukan diputuskan sepihak oleh aplikasi), diumumkan ke pengguna, dan dijalankan otomatis. Setiap penghapusan data karena retensi dicatat (apa yang dihapus, kapan, berdasarkan kebijakan apa).

### 6.5. Keamanan Operasional Lainnya

- **Pembaruan rutin:** PostgreSQL, sistem operasi, dan dependensi Go diperbarui secara berkala. Pembaruan diuji di staging sebelum produksi.
- **Cadangan konfigurasi firewall:** server lokal sekolah hanya membuka port yang diperlukan (port aplikasi + port database hanya untuk localhost / aplikasi, bukan untuk seluruh jaringan bila tidak perlu).
- **Pencatatan perubahan:** setiap perubahan konfigurasi, migrasi, dan restore dicatat (siapa, kapan, apa) agar insiden dapat ditelusuri.
- **Rencana respons insiden:** bila terjadi kebocoran / akses tidak sah, langkahnya adalah: isolasi, simpan bukti (jangan hapus log), ganti kredensial yang terdampak, beri tahu pimpinan sekolah, pulihkan dari backup yang bersih, dan catat pelajaran.

---

## 7. Arsitektur Keseluruhan

### 7.1. Diagram Teks Arsitektur

```
                    ┌─────────────────────────────────────────┐
                    │        JARINGAN LOKAL SEKOLAH            │
                    │   (LAN / Wi-Fi lokal, tanpa internet     │
                    │    wajib untuk ujian inti)               │
                    └─────────────────────────────────────────┘
                                     │
        ┌────────────────────────────┼────────────────────────────┐
        │                            │                            │
   Perangkat siswa            Perangkat guru / proktor       Perangkat admin
   (PC lab / laptop / HP)     (laptop / PC)                  (laptop / PC)
   Browser web                Browser web                    Browser web
        │                            │                            │
        └────────────────────────────┼────────────────────────────┘
                                     │  HTTPS (terenkripsi)
                                     ▼
                    ┌─────────────────────────────────────────┐
                    │         SERVER LOKAL SEKOLAH             │
                    │                                          │
                    │  ┌─────────────────────────────────────┐ │
                    │  │  APLIKASI GO (single binary)        │ │
                    │  │                                     │ │
                    │  │  - HTTP API / halaman web           │ │
                    │  │  - WebSocket (monitor real-time)    │ │
                    │  │  - Logika ujian: timer, token,      │ │
                    │  │    anti-curang, penskoran otomatis  │ │
                    │  │  - Manajemen file (upload media)    │ │
                    │  │  - Scheduler: backup, pembersihan   │ │
                    │  │    sesi, penghitungan nilai         │ │
                    │  └──────┬──────────────┬───────────────┘ │
                    │         │              │                 │
                    │         ▼              ▼                 │
                    │  ┌──────────────┐ ┌──────────────────┐   │
                    │  │  PostgreSQL  │ │  File Storage    │   │
                    │  │  (database   │ │  (disk lokal)    │   │
                    │  │   utama)     │ │  - Media soal    │   │
                    │  │              │ │  - Lampiran      │   │
                    │  │  Semua tabel │ │  - Ekspor CSV    │   │
                    │  │  Bagian 1    │ │  - Backup lokal  │   │
                    │  └──────────────┘ └──────────────────┘   │
                    │                                          │
                    │  ┌─────────────────────────────────────┐ │
                    │  │  Backup & Pemeliharaan              │ │
                    │  │  - pg_dump terjadwal                │ │
                    │  │  - Enkripsi backup                  │ │
                    │  │  - Salinan luar lokasi              │ │
                    │  └─────────────────────────────────────┘ │
                    └─────────────────────────────────────────┘
                                     │
                                     ▼
                    ┌─────────────────────────────────────────┐
                    │   LUAR LOKASI (opsional, non-kritis)    │
                    │   - Salinan backup terenkripsi          │
                    │   - Unduh pembaruan aplikasi            │
                    └─────────────────────────────────────────┘
```

**Keterangan komponen:**

| Komponen | Peran | Catatan |
|---|---|---|
| Browser peserta / guru / proktor | Antarmuka pengguna | Tidak menyimpan data ujian sebagai sumber kebenaran. Browser hanya menampilkan dan mengirim jawaban ke server. |
| Aplikasi Go (single binary) | Otak sistem | Satu binary berisi HTTP API, halaman web, logika ujian, WebSocket, dan scheduler. Mudah di-deploy (satu berkas) — sesuai warisan CBT-Go. |
| PostgreSQL | Sumber kebenaran data | Semua data terstruktur (pengguna, soal, jadwal, jawaban, nilai, log). Satu-satunya tempat yang menentukan nilai akhir. |
| File Storage (disk lokal) | Penyimpanan berkas | Media soal, lampiran, ekspor CSV, backup lokal. Metadata file tercatat di database; isi file di disk. |
| WebSocket | Saluran real-time | Dipakai untuk: monitor peserta real-time (proktor melihat siapa sedang mengerjakan), notifikasi (peringatan anti-curang, jeda jadwal, submit berhasil), dan pembaruan status tanpa refresh. WebSocket tidak dipakai untuk mengirim soal / jawaban utama — soal / jawaban tetap lewat HTTP API agar dapat dicoba ulang (retry) dengan aman. |
| Backup & Pemeliharaan | Operasi terjadwal | Backup database + file, rotasi backup, pembersihan sesi kedaluwarsa, dan penghitungan ulang nilai bila diperlukan. Berjalan sebagai tugas terjadwal di server. |

### 7.2. Aliran Data Saat Ujian (End-to-End)

Urutan berikut adalah aliran data lengkap dari persiapan hingga nilai diterbitkan. Setiap langkah menyebutkan siapa yang terlibat, data apa yang bergerak, dan di mana data disimpan.

#### Tahap A — Persiapan (sebelum hari ujian)

1. **Guru membuat / memilih bank soal** → Aplikasi menyimpan soal + opsi + kunci ke database (tabel Bank Soal, Soal, Opsi).
2. **Guru merakit paket ujian** → Aplikasi menyimpan susunan soal paket (tabel Susunan Soal Paket) + aturan paket (durasi, pengacakan, batas pelanggaran).
3. **Proktor membuat jadwal** → Aplikasi menyimpan jadwal + token + daftar peserta (tabel Jadwal, Peserta Jadwal). Token dibagikan ke siswa melalui saluran resmi sekolah.
4. **Siswa login** → Browser mengirim nama pengguna + kata sandi via HTTPS → Aplikasi memverifikasi hash kata sandi di database → Aplikasi membuat sesi login → Browser menyimpan token sesi.
5. **Siswa melihat dashboard** → Aplikasi mengambil daftar jadwal yang berhak diikuti siswa (via tabel Peserta Jadwal + Jadwal) dan menampilkannya.

#### Tahap B — Memulai Ujian

6. **Siswa memasukkan token ujian** → Aplikasi memeriksa token + masa berlaku + status jadwal + apakah siswa termasuk peserta jadwal.
7. **Aplikasi membuat sesi ujian** → Aplikasi menyimpan baris Sesi Ujian baru (waktu mulai, perangkat, alamat jaringan) + menyimpan **susunan soal per sesi** (urutan soal + urutan opsi setelah pengacakan, bila diaktifkan). Susunan ini dibekukan untuk sesi tersebut.
8. **Aplikasi mengirim soal ke browser** → Browser menerima soal satu per satu atau per halaman (sesuai desain tampilan), **tanpa kunci jawaban**. Kunci jawaban tidak pernah dikirim ke browser.
9. **Proktor memantau via WebSocket** → Dashboard proktor menerima pembaruan real-time: siapa sudah mulai, siapa belum, siapa terkena peringatan.

#### Tahap C — Mengerjakan Ujian (jalur tersibuk)

10. **Siswa menjawab soal** → Browser menyimpan jawaban sementara di memori browser (untuk ketahanan terhadap refresh singkat).
11. **Autosave berkala** → Browser mengirim jawaban ke server via HTTP API (misal setiap perubahan jawaban + setiap interval waktu). Server memperbarui baris Jawaban Peserta untuk pasangan sesi + soal (update, bukan insert baru). Server mengembalikan konfirmasi "tersimpan" + stempel waktu server.
12. **Timer** → Timer resmi adalah **timer server** (waktu mulai sesi + durasi). Timer di browser hanya tampilan. Bila waktu habis, server menutup sesi walau browser tidak mengirim submit.
13. **Tandai ragu-ragu** → Browser mengirim penanda ragu-ragu per soal → Server memperbarui kolom penanda pada baris jawaban.
14. **Peristiwa anti-curang** → Browser mendeteksi pindah tab / keluar fullscreen → Browser mengirim peristiwa ke server → Server menyimpan baris Peristiwa Anti-Curang + menghitung jumlah pelanggaran sesi → Bila melewati batas: server mengirim peringatan (via WebSocket / tampilan) atau memblokir sesi otomatis sesuai aturan paket. Keputusan proktor (cabut blokir / lanjutkan) juga dicatat.
15. **Gangguan koneksi** → Bila jaringan terputus singkat, browser menyimpan jawaban yang belum terkirim dan mengirim ulang saat koneksi kembali. Server menerima kiriman ulang secara idempotent (kiriman yang sama tidak membuat jawaban ganda). Bila perangkat berganti, server mencatat perangkat baru di tabel Perangkat Peserta.

#### Tahap D — Submit & Penskoran

16. **Siswa menekan Submit (atau waktu habis)** → Browser mengirim semua jawaban terakhir + konfirmasi submit → Server mengunci sesi (status selesai), mencatat waktu selesai aktual, dan menolak perubahan jawaban setelah submit.
17. **Penskoran otomatis** → Server menghitung skor per soal untuk soal objektif (pilihan ganda, benar-salah, isian, menjodohkan) dengan membandingkan jawaban peserta dengan kunci jawaban yang tersimpan di database. Skor otomatis disimpan pada baris Jawaban Peserta.
18. **Agregasi nilai** → Server menjumlahkan skor per soal menjadi Hasil / Nilai Ujian (skor total, nilai 0–100, jumlah benar / salah / kosong). Hasil disimpan dengan status "menunggu koreksi" bila masih ada esai yang belum dikoreksi.
19. **Koreksi manual esai** → Guru membuka daftar jawaban esai yang belum dikoreksi untuk jadwal tersebut → Guru memberi skor + komentar → Aplikasi menyimpan baris Koreksi Manual + memperbarui skor manual pada baris jawaban → Setelah semua esai terkoreksi, server menghitung ulang Hasil / Nilai (skor akhir = objektif + esai) dan menandai status koreksi selesai.
20. **Penerbitan nilai** → Proktor / admin menerbitkan nilai (status hasil menjadi diterbitkan) → Siswa dapat melihat nilai + pembahasan (bila diizinkan paket). Nilai yang sudah diterbitkan tidak dapat diubah tanpa membuka kunci + alasan (lihat Bagian 2.2).

#### Tahap E — Rekap, Ekspor, dan Audit

21. **Rekap nilai** → Guru / proktor membuka rekap per jadwal / per kelas → Aplikasi membaca tabel Hasil (bukan menghitung ulang dari jawaban setiap kali) + tabel Nilai per Kompetensi bila ada → Menampilkan tabel rekap + statistik (rata-rata, kelulusan, distribusi).
22. **Ekspor** → Aplikasi menghasilkan berkas CSV rekap dari tabel Hasil dan menyimpannya sementara di File Storage untuk diunduh. Ekspor dicatat di Log Aktivitas (siapa mengekspor apa, kapan).
23. **Analisis butir soal** → Secara berkala (atau atas permintaan), sistem menghitung Analisis Butir Soal dari tabel Jawaban + Hasil dan menyimpan hasilnya (tingkat kesulitan, daya beda, pengecoh) untuk laporan guru.
24. **Audit** → Semua aksi penting selama Tahap A–E sudah tercatat di Log Aktivitas. Bila ada sengketa nilai, admin dapat menelusuri: sesi apa, susunan soal apa, jawaban apa, koreksi siapa, dan kapan nilai berubah.

### 7.3. Keputusan Arsitektur Kunci & Alasannya

| Keputusan | Alasan |
|---|---|
| Tetap satu binary Go + satu PostgreSQL | Sederhana di-deploy di server sekolah, mudah di-backup, dan sudah terbukti di CBT-Go. Tidak perlu arsitektur microservices untuk skala ratusan siswa. |
| PostgreSQL sebagai satu-satunya sumber kebenaran | Menghindari dua sumber kebenaran (browser vs server). Nilai, jawaban, dan status sesi hanya sah bila tercatat di database. |
| Kunci jawaban tidak pernah dikirim ke browser | Mencegah kecurangan via inspeksi browser / alat pengembang. Browser hanya menerima soal, bukan kunci. |
| Timer resmi di server, bukan di browser | Mencegah manipulasi timer dengan mengubah jam perangkat / memodifikasi browser. |
| Soal / jawaban utama via HTTP API, WebSocket hanya untuk monitor / notifikasi | HTTP lebih mudah dicoba ulang (retry) dan lebih dapat diandalkan untuk data penting. WebSocket dipakai untuk hal real-time yang boleh hilang (status monitor) tanpa merusak ujian. |
| File media di disk, metadata di database | Database tetap ramping; file besar tidak membebani database. Backup harus mencakup keduanya. |
| Server lokal sebagai sumber kebenaran saat ujian | Mode semi-daring: ujian tidak bergantung pada internet. Internet hanya untuk backup luar lokasi / pembaruan, bukan untuk alur ujian inti. |
| Soft delete + versi soal untuk data ujian | Melindungi audit dan nilai historis. Penghapusan fisik hanya untuk data yang benar-benar belum pernah dipakai. |

### 7.4. Risiko & Mitigasi Arsitektur

| Risiko | Dampak | Mitigasi |
|---|---|---|
| Lonjakan login bersamaan (ratusan siswa login dalam 5 menit) | Login lambat / timeout | Indeks login (Bagian 3.1), sesi login ringan, jadwal login bertahap per gelombang bila perlu, uji beban sebelum ujian besar. |
| Autosave bersamaan membebani database | Jawaban gagal tersimpan | Indeks unik sesi + soal, autosave dengan interval yang wajar (bukan per ketikan), retry idempotent di browser, dan pemantauan beban saat ujian. |
| Server mati di tengah ujian | Sesi terputus, jawaban terakhir belum tersimpan | Autosave berkala + timer server + backup harian + prosedur resume sesi (siswa login ulang dan melanjutkan sesi yang sama dengan sisa waktu yang tersisa, bila diizinkan proktor). |
| Token bocor ke siswa yang tidak berhak | Siswa tidak berhak masuk ujian | Token per jadwal + masa berlaku + daftar peserta yang berhak + pergantian token yang mudah oleh proktor + log percobaan masuk yang gagal. |
| Soal / kunci bocor | Ujian tidak sah | Kunci tidak dikirim ke browser, akses bank soal dibatasi peran, log akses soal, dan rotasi soal antar gelombang bila memungkinkan. |
| Disk penuh karena backup / log | Aplikasi berhenti | Rotasi backup + rotasi log + pemantauan ruang disk + peringatan dini sebelum disk penuh. |
| Satu operator memegang semua pengetahuan | Gagal pulih saat operator absen | Dokumentasi operasi + latihan restore + minimal dua orang yang dapat melakukan backup / restore. |

---

## Daftar Periksa Implementasi (Handoff ke Tim Pengembang)

- [ ] Skema dibuat mengikuti daftar tabel Bagian 1 (nama kolom fisik diputuskan di implementasi, bukan di laporan ini).
- [ ] Relasi + aturan hapus / ubah Bagian 2 ditegakkan di lapisan aplikasi (dan, bila memungkinkan, di lapisan database).
- [ ] Indeks Bagian 3 dibuat bersama migrasi awal.
- [ ] Migrasi berversi + seed idempotent (Bagian 4) teruji di pengembangan → staging → produksi.
- [ ] Backup harian + rotasi + tes restore bulanan (Bagian 5) dijadwalkan dan didokumentasikan.
- [ ] Enkripsi koneksi + hash kata sandi + backup terenkripsi + retensi log (Bagian 6) diterapkan.
- [ ] Aliran ujian Bagian 7.2 diuji end-to-end: login → token → soal → autosave → submit → koreksi esai → nilai → rekap.
- [ ] Uji beban: login massal + autosave bersamaan + submit massal, dengan ratusan sesi simulasi.

---

*Laporan ini adalah dokumen desain konseptual. Tidak berisi kode aplikasi atau perintah SQL. Keputusan implementasi (nama kolom fisik, sintaks migrasi, framework, library) mengikuti laporan ini dan dicatat terpisah.*
