# ALUR ACTIVITY NYATA — KELOMPOK A (Auth + Master + Pengaturan + Unblock) — CBT SMK v2

> **Status:** Dokumen telusur Fase 1 Kelompok A. Disusun HANYA dari membaca kode aplikasi
> (halaman depan + layanan belakang), bukan dari diagram Activity lama, bukan dari ingatan.
> **Tanggal telusur:** 18 Okt 2026.
> **Aturan penulisan:** Bahasa desain (PIM). TIDAK ada nama berkas, alamat layanan,
> perintah basis data, pustaka, atau nama kolom di dokumen ini — yang ditulis adalah
> perilaku yang terlihat dari kode: aktor, langkah, kondisi, hasil akhir, dan aturan bisnis.
> **Cakupan:** UC01, UC02, UC17, UC18 (nama persis dari diagram Use Case).
> **Batasan:** Dokumen ini read-only terhadap model diagram; tidak mengubah kode aplikasi.

---

## UC01 — Masuk & Wajib Ganti Kata Sandi

### Aktor

- **Pengguna** — siswa atau staf (admin, proktor, guru, pengawas, penguji UKK).
- **Sistem** — memvalidasi kredensial, mengelola sesi, menegakkan ganti kata sandi wajib,
  mengunci akun sementara setelah gagal berulang, mencatat jejak audit.

### Alur utama

1. Pengguna membuka aplikasi. Bila sesi sebelumnya masih berlaku, sistem langsung
   menampilkan beranda sesuai peran tanpa meminta login lagi; bila tidak, pengguna
   diarahkan ke halaman masuk.
2. Pengguna mengisi **email atau nama pengguna** dan **kata sandi**, lalu mengirim formulir.
   Siswa masuk dengan nama pengguna (NIS/NISN).
3. Sistem memeriksa kelengkapan isian: login dan kata sandi kosong ditolak.
4. Sistem mencari akun berdasarkan email ATAU nama pengguna, tanpa membedakan huruf
   besar/kecil dan dengan spasi diabaikan.
5. Sistem memverifikasi kata sandi terhadap kata sandi tersimpan.
6. Bila akun tidak aktif, masuk ditolak dengan pesan agar menghubungi admin.
7. Bila akun ditandai **wajib ganti kata sandi**, masuk ditolak TANPA membuat sesi,
   dan pengguna diarahkan ke halaman ganti kata sandi dengan nama login terbawa.
8. Bila semua pemeriksaan lolos, sistem membuat sesi baru, mencatat jejak audit
   masuk, dan mengarahkan pengguna: siswa ke beranda siswa, staf ke dasbor staf.
9. Selama sesi berjalan, setiap permintaan memvalidasi sesi: sesi yang melewati
   batas umur (12 jam) atau tidak disentuh lebih dari 2 jam dianggap kedaluwarsa,
   sesi dihapus, dan pengguna harus login ulang.
10. Setiap perubahan data (tambah/ubah/hapus) menyertakan token pengaman anti-pemalsuan
    permintaan antar-situs yang dicocokkan dengan token sesi; token tidak cocok
    menyebabkan perubahan ditolak.
11. Pengguna yang sudah login dan membuka halaman masuk tidak melihat formulir lagi,
    melainkan keterangan "Sudah login sebagai …" dengan tautan lanjut ke beranda.
12. Pengguna keluar: sesi saat itu dihapus, penanda sesi dibersihkan, jejak audit
    keluar dicatat, dan penyimpanan lokal status login dikosongkan walau permintaan
    keluar gagal di jaringan.

### Cabang / kondisi

| # | Kondisi | Hasil akhir |
|---|---------|-------------|
| 1 | Isian login atau kata sandi kosong | Ditolak; pesan validasi, tidak ada sesi |
| 2 | Akun tidak ditemukan | Ditolak (pesan generik "email/nama pengguna atau kata sandi salah"); kegagalan dicatat + jejak audit gagal masuk |
| 3 | Kata sandi salah | Ditolak (pesan generik yang sama); kegagalan dicatat + jejak audit gagal masuk |
| 4 | Akun tidak aktif | Ditolak; pesan "Akun tidak aktif. Hubungi admin."; tanpa sesi |
| 5 | Akun wajib ganti kata sandi | Ditolak dengan penanda khusus; tanpa sesi; arah ke ganti kata sandi |
| 6 | 5 kali gagal masuk untuk satu nama login dalam 15 menit | Percobaan ke-6 dan seterusnya dikunci 15 menit; pesan "Akun terkunci sementara… Coba lagi dalam 15 menit." |
| 7 | Masuk berhasil | Hitungan gagal nama login dihapus; sesi baru dibuat |
| 8 | Sesi melewati umur 12 jam atau idle lebih dari 2 jam | Sesi dihapus; permintaan ditolak sebagai belum terautentikasi |
| 9 | Token pengaman tidak cocok pada perubahan data | Perubahan ditolak |
| 10 | Pengguna tanpa login membuka rute terlindungi | Diarahkan ke halaman masuk |
| 11 | Siswa membuka area khusus staf | Diarahkan ke halaman akses ditolak (403) |
| 12 | Ganti kata sandi dari dalam sesi / tanpa sesi | Lihat sub-alur di bawah |

### Sub-alur: Ganti kata sandi (dipakai saat login pertama maupun dari dalam sesi)

1. Halaman ganti kata sandi tampil dengan keterangan sesuai konteks: bila berasal dari
   penolakan wajib ganti, keterangannya "Kata sandi awal harus diganti sebelum login.";
   bila dari dalam sesi, keterangannya meminta kata sandi lama dan baru.
2. Pengguna mengisi: nama login (boleh dikosongkan bila sudah login), kata sandi lama,
   kata sandi baru, konfirmasi kata sandi baru.
3. Validasi di sisi tampilan (permintaan TIDAK dikirim bila gagal):
   kata sandi baru minimal 8 karakter; konfirmasi harus sama dengan kata sandi baru;
   kata sandi baru harus berbeda dari kata sandi lama.
4. Sistem menentukan konteks: bila ada sesi aktif, pengguna diambil dari sesi dan
   isian nama login diabaikan; bila tanpa sesi, pengguna dicari lewat nama login
   (nama login kosong → ditolak sebagai belum terautentikasi).
5. Validasi sistem: kata sandi baru kurang dari 8 karakter ditolak; kata sandi baru
   sama dengan kata sandi lama ditolak.
6. Sistem memverifikasi kata sandi lama; salah → ditolak (pesan "Kata sandi lama salah.").
7. Bila lolos, kata sandi tersimpan diperbarui, penanda wajib ganti dimatikan,
   **SELURUH sesi pengguna dihapus**, penanda sesi dibersihkan, dan jejak audit
   ganti kata sandi dicatat.
8. Pengguna diarahkan kembali ke halaman masuk dengan pesan "Kata sandi berhasil
   diganti. Silakan login dengan kata sandi baru." — **tidak ada login otomatis**;
   pengguna wajib masuk ulang dengan kata sandi baru.
9. Pemicu penanda wajib ganti di sistem: akun admin pertama yang dibuat saat basis
   data pengguna masih kosong, akun yang kata sandinya direset lewat perintah
   pengelola, dan setiap akun siswa baru (dibuat manual maupun lewat import —
   lihat UC02).

### Validasi / aturan bisnis yang terlihat di kode

- Login menerima email ATAU nama pengguna, tanpa membedakan huruf besar/kecil.
- Penguncian login: maks 5 gagal per nama login per 15 menit; kunci 15 menit;
  penguncian dihitung per nama login, disimpan di memori proses (bukan di basis data).
  Hanya kegagalan "akun tidak ditemukan" dan "kata sandi salah" yang menambah hitungan;
  penolakan akun tidak aktif dan wajib ganti kata sandi TIDAK menambah hitungan.
- Kata sandi baru minimal 8 karakter dan harus berbeda dari kata sandi lama
  (ditegakkan di tampilan DAN di sistem).
- Ganti kata sandi yang berhasil selalu mengakhiri semua sesi pengguna
  (termasuk sesi yang sedang berjalan) — pengguna wajib login ulang.
- Penanda wajib ganti TIDAK membuat sesi: pengguna tidak bisa masuk sebelum ganti.
- Jejak audit dicatat untuk: masuk berhasil, masuk gagal (beserta alasan),
  keluar, dan ganti kata sandi.
- Peran menentukan tujuan sesudah masuk: siswa → beranda siswa; staf → dasbor staf.
- Nilai "kembali ke halaman sebelumnya" yang disimpan saat pengalihan ke login
  tidak pernah dipakai untuk redirect balik; navigasi selalu berdasarkan peran.

---

## UC02 — Kelola Master Data (jurusan, kelas, mapel, guru, siswa — termasuk import siswa CSV)

### Aktor

- **Staf pengelola** — admin (lolos semua pemeriksaan izin) dan peran staf lain
  sejauh diberi izin per entitas; siswa tidak dapat mengelola master data.
- **Sistem** — memeriksa izin per aksi, memvalidasi isian, menegakkan keunikan,
  menegakkan aturan keterkaitan antar entitas, menolak hapus data yang masih dipakai,
  membuat akun login siswa otomatis, mencatat jejak audit.

### Pola umum (berlaku untuk kelima entitas)

1. Staf membuka halaman daftar entitas. Sistem memeriksa izin baca entitas itu;
   tanpa izin → ditolak.
2. Daftar tampil dengan pencarian teks, filter (bila ada), dan penomoran halaman
   (bawaan 20 baris per halaman, maks 100; halaman mulai dari 1).
3. **Tambah:** staf menekan tombol tambah → formulir modal terbuka → staf mengisi →
   sistem memvalidasi → bila lolos, data baru tersimpan dan jejak audit pencatatan
   ditulis; bila gagal, pesan kesalahan ditampilkan dan modal tetap terbuka.
4. **Ubah:** staf menekan ubah pada satu baris → formulir modal terbuka terisi data
   lama → staf mengubah → sistem memvalidasi → bila lolos, data diperbarui dan
   jejak audit perubahan ditulis.
5. **Hapus:** staf menekan hapus → dialog konfirmasi tampil → staf mengonfirmasi →
   sistem memeriksa pemakaian data (guard) → bila masih dipakai, hapus ditolak
   dengan pesan yang menyebut jumlah pemakaian; bila tidak dipakai, data dihapus
   permanen (hapus keras, tanpa tempat sampah / tanpa pemulihan) dan jejak audit
   penghapusan ditulis.
6. Bila pengenal data tidak berbentuk pengenal yang sah, sistem menjawab
   "tidak ditemukan" (bukan kesalahan validasi).

### Alur per entitas

#### Jurusan (bidang: kode, nama)

1. Daftar: cari berdasarkan nama atau kode; urut nama menaik.
2. Tambah/ubah validasi: kode wajib diisi; nama wajib diisi.
3. Simpan: kode yang sudah dipakai → ditolak (pesan "Kode jurusan sudah dipakai.").
4. Hapus — guard, sesuai urutan pemeriksaan: bila masih dipakai kelas → ditolak
   ("…masih dipakai N kelas…"); bila masih dipakai mapel produktif → ditolak;
   bila masih dipakai siswa → ditolak. Bila tidak dipakai, hapus berhasil.

#### Kelas (bidang: jurusan, tingkat, nama, tahun ajaran)

1. Daftar: saring per jurusan; cari berdasarkan nama atau tahun ajaran;
   urut tahun ajaran menurun, lalu tingkat menaik, lalu nama menaik.
2. Tambah/ubah validasi: jurusan wajib dipilih (pengenal sah); tingkat harus
   10, 11, atau 12; nama wajib diisi; tahun ajaran wajib diisi (contoh 2026/2027).
3. Simpan: kombinasi nama + tingkat + tahun ajaran yang sama dalam satu jurusan
   → ditolak (pesan duplikat kelas). Jurusan yang tidak ditemukan → ditolak.
4. Hapus — guard: bila masih dipakai siswa → ditolak; bila masih dipakai jadwal
   ujian → ditolak. Bila tidak dipakai, hapus berhasil.

#### Mapel / mata pelajaran (bidang: kode, nama, jenis, jurusan, kelas tingkat)

1. Daftar: saring per jenis (umum/produktif) dan per jurusan; cari berdasarkan
   nama atau kode; urut nama menaik. Nilai jenis selain "umum"/"produktif"
   pada saringan diabaikan (dianggap tanpa saringan jenis).
2. Tambah/ubah validasi: kode wajib diisi; nama wajib diisi; jenis harus
   "umum" atau "produktif"; mapel produktif WAJIB terikat satu jurusan;
   mapel umum TIDAK BOLEH terikat jurusan (jurusan harus kosong);
   kelas tingkat bila diisi harus 10, 11, atau 12.
3. Simpan: kode yang sudah dipakai → ditolak ("Kode mapel sudah dipakai.").
   Jurusan yang tidak ditemukan → ditolak.
4. Hapus — guard, sesuai urutan pemeriksaan: bila masih dipakai kompetensi →
   ditolak; kisi-kisi → ditolak; bank soal → ditolak; paket ujian → ditolak;
   jadwal ujian (melalui paket) → ditolak. Bila tidak dipakai, hapus berhasil.

#### Guru (bidang: nama, NIP opsional, akun login opsional, daftar mapel yang diampu)

1. Daftar: cari berdasarkan nama atau NIP; urut nama menaik.
2. Tambah/ubah validasi: nama wajib diisi; pengenal akun login bila diisi harus
   berbentuk pengenal yang sah; setiap pengenal mapel dalam daftar harus sah.
   NIP boleh kosong; akun login boleh kosong (guru belum tentu punya akun);
   daftar mapel kosong diterima sebagai daftar kosong.
3. Simpan: NIP yang sudah dipakai guru lain → ditolak ("NIP sudah dipakai guru lain.").
   Akun login yang tidak ditemukan → ditolak.
4. Hapus — guard: bila guru masih menjadi pemilik bank soal (melalui akun
   loginnya) → ditolak dengan pesan agar akunnya dinonaktifkan saja, bukan dihapus.
   Bila tidak, hapus berhasil. (Catatan: tidak ada pemeriksaan pemakaian guru
   pada entitas lain selain bank soal.)

#### Siswa (bidang: NISN, NIS opsional, nama, kelas, jurusan, tahun masuk opsional, status aktif)

1. Daftar: saring per kelas, per jurusan, per status aktif; cari berdasarkan
   nama, NISN, atau NIS; urut nama menaik.
2. Tambah manual validasi: NISN wajib diisi; nama wajib diisi; kelas wajib
   dipilih (pengenal sah); jurusan wajib dipilih (pengenal sah); pengenal akun
   login bila diisi harus sah; tahun masuk bila diisi harus 1900–2100.
   Status aktif bawaan aktif bila tidak diisi.
3. Tambah manual — pembuatan akun: bila pengenal akun login TIDAK diisi, sistem
   membuat akun login otomatis dalam transaksi yang sama dengan data siswa:
   nama pengguna = NISN, kata sandi awal = NISN, peran siswa, status aktif,
   dan penanda wajib ganti kata sandi AKTIF (siswa wajib ganti saat login pertama).
   Bila nama pengguna sudah ada (akun yatim), akun lama dipakai kembali dan
   penanda wajib gantinya diaktifkan + perannya dipastikan siswa.
   Bila pengenal akun login diisi, akun tersebut dipakai apa adanya.
4. Ubah: pengenal akun login yang tidak dikirim → akun lama dipertahankan;
   pengenal akun login kosong → akun dikosongkan. NIS kosong → disimpan kosong.
   Perubahan NISN duplikat → ditolak ("NISN sudah terdaftar.").
   Kelas/jurusan/akun yang tidak ditemukan → ditolak.
5. Hapus — guard: bila siswa sudah punya attempt ujian → ditolak
   ("…sudah punya N attempt ujian. Nonaktifkan saja, jangan hapus.");
   bila sudah punya nilai akhir → ditolak; bila sudah punya nilai remedial →
   ditolak. Bila tidak dipakai, data siswa dihapus; **akun loginnya SENGAJA
   TIDAK dihapus** (riwayat terjaga, nama pengguna tidak dipakai ulang).
6. Menonaktifkan siswa (mengubah status aktif menjadi nonaktif) adalah cara yang
   didukung untuk siswa yang sudah punya riwayat ujian — bukan menghapusnya.
   (Akun login yang nonaktif juga ditolak saat masuk — lihat UC01.)

#### Sub-alur: Import siswa dari CSV

1. Staf membuka halaman import dari daftar siswa (tombol "Import CSV").
2. Halaman menjelaskan format: header `nama,nisn,nis,kelas_kode` atau
   `nama,nisn,kelas_id`; kolom NIS opsional; `kelas_kode` diisi pengenal kelas
   ATAU nama kelas (mis. "XI RPL 1"); maks 2.000 baris per berkas, maks 5 MB;
   baris gagal dilewati dan dilaporkan; baris valid tetap disimpan.
3. Staf dapat mengunduh templat CSV contoh dari halaman.
4. Staf memilih berkas CSV (hanya `.csv`), lalu menekan "Upload & Import".
   Izin yang diperiksa adalah izin khusus import siswa, bukan izin tambah siswa.
5. Sistem memeriksa berkas: bukan multipart yang sah / lebih dari 5 MB →
   ditolak; tidak ada berkas → ditolak; bukan CSV yang sah → ditolak;
   berkas kosong → ditolak.
6. Sistem menormalisasi header: huruf kecil, spasi dipangkas, penanda BOM
   dihapus; nama kolom alias diterima (`nama_lengkap`/`nama_siswa` untuk nama;
   `kelas_id`/`kelas_kode`/`kelas`/`kode_kelas` untuk kelas).
   Bila kolom wajib (nama, NISN, kelas) tidak ditemukan → ditolak dengan pesan
   yang menyebut kolom yang harus ada.
7. Bila jumlah baris data melebihi 2.000 → ditolak.
8. Sistem memproses baris satu per satu dalam satu transaksi, dengan hasil
   per baris sebagai berikut:
   - Nama kosong → gagal ("Nama wajib diisi."), baris dilewati.
   - NISN kosong → gagal ("NISN wajib diisi."), baris dilewati.
   - Kelas kosong → gagal ("Kelas … wajib diisi."), baris dilewati.
   - NISN duplikat di dalam berkas yang sama → gagal ("NISN duplikat di dalam file."),
     baris dilewati.
   - NISN sudah terdaftar di basis data → dihitung **sudah ada** (bukan gagal,
     bukan duplikat, tidak membuat akun baru, tidak masuk daftar kesalahan) —
     import bersifat idempotent: mengulang berkas yang sama aman.
   - Kelas tidak ditemukan (baik dicari lewat pengenal maupun lewat nama kelas
     tanpa membedakan huruf besar/kecil; bila beberapa kelas bernama sama,
     dipakai yang terbaru) → gagal ("Kelas … tidak ditemukan."), baris dilewati.
   - Bila lolos, akun login dibuat otomatis (nama pengguna = NISN, kata sandi
     awal = NISN, penanda wajib ganti aktif) dan baris siswa disimpan aktif.
   - Duplikat NISN yang muncul karena import paralel di antara pemeriksaan dan
     penyimpanan → dihitung sudah ada (jaring pengaman).
9. Bila semua baris selesai, transaksi diterapkan dan jejak audit import ditulis
   (jumlah berhasil, gagal, sudah ada).
10. Halaman menampilkan ringkasan: jumlah berhasil, jumlah gagal,
    (ditambah jumlah sudah ada bila ada) dan tabel kesalahan per baris
    (nomor baris + pesan). Bila tidak ada kesalahan dan ada yang berhasil,
    tampil pesan bahwa import selesai.

### Validasi / aturan bisnis yang terlihat di kode

- Izin diperiksa per entitas dan per aksi (baca / tambah / ubah / hapus;
  import siswa memakai izin sendiri). Peran admin lolos semua pemeriksaan izin.
- Penomoran halaman: bawaan halaman 1, 20 baris per halaman, maks 100;
  nilai tidak sah jatuh ke bawaan.
- Semua hapus adalah hapus keras permanen; perlindungan satu-satunya adalah
  guard pemakaian + dialog konfirmasi di tampilan. Tidak ada pemulihan.
- Aturan keterkaitan mapel (produktif wajib jurusan; umum wajib tanpa jurusan)
  ditegakkan pada validasi.
- Aturan tingkat kelas dan kelas tingkat mapel: hanya 10, 11, 12.
- Keunikan: kode jurusan, kode mapel, NISN siswa, NIP guru, nama pengguna akun,
  kode kompetensi per mapel, kombinasi kelas (jurusan + tingkat + nama + tahun ajaran).
- Akun siswa selalu dibuat bersama data siswa (manual maupun import) dengan
  kata sandi awal = NISN dan wajib ganti saat login pertama.
- Import: idempotent terhadap NISN yang sudah ada; baris gagal tidak
  menggagalkan baris valid; rekomendasi operasional maks 100 siswa per berkas
  (perkiraan ±17 detik, jauh di bawah batas waktu tulis 120 detik).
- Setiap tambah/ubah/hapus/import master data menulis jejak audit; kegagalan
  menulis jejak audit tidak menggagalkan operasi utama.
- Kompetensi dan kisi-kisi (sub-master mapel) memiliki operasi baca/tambah/ubah/hapus
  di layanan belakang dengan izin mapel (baca/tambah/ubah/hapus mapel) dan guard
  hapus sendiri (kompetensi dipakai kisi-kisi/soal → ditolak; kisi-kisi dipakai
  soal → ditolak). Di tampilan, TIDAK ditemukan halaman kelola kompetensi /
  kisi-kisi — halaman Kompetensi yang ada hanya menampilkan nilai per kompetensi
  (baca saja). Lihat § "Hal yang tidak ditemukan".

---

## UC17 — Kelola Pengaturan, Audit & Notifikasi

### Aktor

- **Admin / staf berizin** — mengelola pengaturan sekolah; membaca jejak audit;
  menerima dan menandai notifikasi.
- **Sistem** — menyimpan pengaturan kunci-nilai, memvalidasi nilai per kunci,
  mencatat jejak audit (hanya tambah + baca, tidak pernah diubah/dihapus),
  mengirim notifikasi dalam aplikasi, mencatat pengecualian pemisahan tugas (SoD).

### Alur utama — 1. Kelola pengaturan sekolah

1. Staf membuka halaman Pengaturan. Sistem memeriksa izin baca pengaturan;
   tanpa izin → ditolak.
2. Sistem menampilkan nilai saat ini untuk semua kunci yang dikenal; kunci yang
   belum pernah disimpan tampil dengan nilai bawaan (lihat tabel di bawah).
3. Staf mengubah nilai pada formulir: saklar tampilkan hasil ke siswa; angka
   (masa berlaku token, toleransi telat, maks pelanggaran, durasi bawaan);
   nama sekolah; logo sekolah (diunggah sebagai berkas, bukan diketik sebagai
   alamat); KKTP per mapel sebagai teks JSON; mode remedial tampil baca-saja.
4. Staf menekan Simpan. Tampilan memvalidasi KKTP: teks harus berupa objek JSON
   (bukan larik, bukan kosong) — bila tidak valid, pesan kesalahan ditampilkan
   dan permintaan TIDAK dikirim. Nilai remedial selalu dikirim sebagai
   "tertinggi" (tidak dapat diubah dari tampilan).
5. Sistem memeriksa izin ubah pengaturan; tanpa izin → ditolak.
6. Sistem menerima isi perubahan dalam dua bentuk: terbungkus (`settings: {...}`)
   maupun datar (`{kunci: nilai}`); bila tidak ada pengaturan yang dikirim → ditolak.
7. Sistem memvalidasi tiap kunci dalam satu transaksi: kunci yang tidak dikenal
   → ditolak ("Key tidak dikenal: …"); nilai yang tidak sesuai tipe/rentang
   kunci → ditolak ("Nilai tidak valid untuk key: …"). Satu kunci tidak valid
   menggagalkan seluruh penyimpanan (transaksi dibatalkan).
8. Bila lolos, semua perubahan disimpan sekaligus, jejak audit perubahan
   pengaturan ditulis (berisi kunci-kunci yang diubah), dan tampilan menampilkan
   "Settings berhasil disimpan." lalu memuat ulang nilai.

**Kunci pengaturan + nilai bawaan + aturan validasi:**

| Kunci | Bawaan | Aturan |
|---|---|---|
| Tampilkan hasil ke siswa | tidak aktif | harus benar/salah |
| Masa berlaku token (detik) | 900 (15 menit) | bilangan bulat 60–86400 |
| Toleransi telat (menit) | 15 | bilangan bulat 0–120 |
| Maks pelanggaran | 2 | bilangan bulat 1–10 |
| Durasi bawaan (menit) | 90 | bilangan bulat 5–600 |
| Nama sekolah | kosong | teks apa pun |
| Logo sekolah | kosong | teks (alamat hasil unggah) |
| KKTP per mapel | objek kosong | harus objek; tiap mapel/TP punya deskripsi/rubrik/interval sendiri — BUKAN satu angka KKM untuk semua mapel |
| Mode remedial | "tertinggi" | HANYA "tertinggi" yang diterima; nilai lain ditolak |

### Alur utama — 2. Jejak audit (baca saja)

1. Setiap operasi penting di seluruh aplikasi menulis satu baris jejak audit:
   pelaku, aksi, entitas, pengenal entitas, rincian, alamat IP, waktu.
   Kegagalan menulis jejak audit TIDAK menggagalkan operasi utama.
2. Layanan baca jejak audit: memeriksa izin baca audit; tanpa izin → ditolak.
3. Pembaca dapat menyaring berdasarkan aksi, entitas, dan pelaku; urut waktu
   terbaru dulu; penomoran dengan batas bawaan 50, maks 200, offset bawaan 0.
4. Jejak audit **tidak pernah diubah atau dihapus lewat layanan apa pun**:
   hanya ada pencatatan (tambah) dan pembacaan. Tidak ada layanan ubah/hapus audit.
5. **Di tampilan, TIDAK ditemukan halaman daftar jejak audit** — pembacaan
   jejak audit hanya tersedia di layanan belakang. Satu-satunya penanda di
   tampilan adalah kalimat di halaman Pengaturan: "Perubahan tercatat di audit log."
   Lihat § "Hal yang tidak ditemukan".

### Alur utama — 3. Notifikasi dalam aplikasi

1. Sistem membuat notifikasi untuk penerima tertentu pada kejadian yang
   membutuhkannya — yang terlihat di kode: pengecualian pemisahan tugas (SoD)
   mengirim notifikasi ke semua pengguna peran admin yang aktif kecuali pelaku.
   Gagal membuat notifikasi TIDAK menggagalkan operasi pemicu.
2. Setiap pengguna hanya dapat membaca notifikasinya sendiri (notifikasi milik
   pengguna lain tidak dapat dibaca — permintaan milik orang lain dijawab
   "tidak ditemukan").
3. Bel notifikasi di bilah atas tampil untuk staf: polling daftar notifikasi
   milik sendiri setiap 60 detik; lencana menampilkan jumlah belum dibaca
   (maks tampil 99); menu turun menampilkan judul, isi, waktu, dan tombol
   tandai dibaca per notifikasi.
4. Pengguna menekan tandai dibaca pada satu notifikasi → sistem menandai
   notifikasi itu sebagai dibaca (hanya bila milik pengguna itu; bila bukan
   miliknya → "tidak ditemukan") dan menulis jejak audit tandai-dibaca;
   daftar dimuat ulang. Kegagalan tandai-dibaca diabaikan diam-diam di tampilan.
5. Bila layanan notifikasi menjawab galat apa pun (mis. belum tersedia),
   bel menampilkan 0 dan daftar kosong — tidak pernah membuat tampilan gagal.
6. Layanan baca notifikasi mendukung penomoran (batas bawaan 50, maks 200).

### Alur utama — 4. Pengecualian pemisahan tugas (SoD) darurat

Aturan pemisahan tugas yang berlaku (7 aturan): pembuat/pengubah bank soal
bukan pengunci; pembuat soal bukan pemegang token; pengoreksi esai bukan
pengubah nilai final; penguji UKK bukan finalisator nilai UKK; pengawas
mengajukan, proktor mengeksekusi; pengelola pengguna bukan pengelola jejak audit;
pengaju draf jadwal bukan finalisator.

1. Admin yang berizin mencatat pengecualian darurat dengan mengisi: aksi yang
   dikecualikan, entitas, pengenal entitas, dan alasan tertulis.
2. Sistem memeriksa: pengguna harus login; pengguna harus punya izin khusus
   pengecualian SoD (hanya admin); aksi dan entitas wajib diisi;
   alasan wajib minimal 20 karakter (dihitung per karakter, spasi dipangkas)
   — bila kurang → ditolak.
3. Bila lolos, pengecualian disimpan dan sistem menulis jejak audit dengan
   penanda khusus pengecualian SoD (berisi aksi + alasan + entitas).
   Gagal menulis jejak audit tidak menggagalkan pengecualian yang sudah
   tersimpan, tetapi dicatat di log server.
4. Sistem mengirim notifikasi "Pengecualian SoD" ke SEMUA pengguna peran admin
   yang aktif KECUALI pelaku, berisi nama pelaku, aksi, entitas, pengenal
   entitas, dan alasan.
5. Pengecualian TIDAK PERNAH dipakai untuk mengubah nilai tanpa jejak atau
   menghapus jejak audit — jejak audit hanya-tambah dan tidak ada layanan
   ubah/hapus.
6. **Di tampilan, TIDAK ditemukan formulir pengecualian SoD** — pencatatan
   pengecualian hanya tersedia di layanan belakang. Lihat § "Hal yang tidak ditemukan".

### Validasi / aturan bisnis yang terlihat di kode

- Pengaturan: hanya kunci yang dikenal yang diterima; nilai di luar tipe/rentang
  kunci ditolak; seluruh perubahan dalam satu penyimpanan berhasil atau gagal
  bersama (transaksi).
- Mode remedial selalu "tertinggi": nilai akhir remedial memakai nilai tertinggi
  antara nilai awal dan nilai remedial; nilai awal tidak pernah ditimpa.
- Jejak audit hanya-tambah + baca; tidak ada ubah/hapus audit di layanan mana pun.
- Notifikasi: hanya pemilik yang dapat membaca/menandai; bel polling 60 detik.
- Pengecualian SoD: izin khusus (admin saja); alasan minimal 20 karakter;
  jejak audit berpenanda khusus; notifikasi ke semua admin lain yang aktif.
- Perubahan pengaturan menulis jejak audit berisi kunci-kunci yang diubah.

---

## UC18 — Usul & Proses Buka Blokir

### Aktor

- **Siswa yang diblokir** — pengerjaan ujiannya diblokir otomatis setelah
  pelanggaran berulang; menunggu dibuka blokirnya; melanjutkan dari sisa waktu.
- **Pengawas** — mengusulkan buka blokir (mengajukan), TIDAK dapat mengeksekusi
  buka blokir langsung.
- **Proktor / admin** — mengeksekusi buka blokir; juga dapat mengusulkan.
- **Sistem** — memvalidasi usulan dan buka blokir, menghitung sisa waktu lanjutan,
  menutup usulan yang menunggu, menyiarkan perubahan status ke pemantau,
  mencatat jejak audit.

### Konteks pemicu (dari mana blokir berasal)

Pengerjaan ujian diblokir otomatis oleh sistem saat jumlah pelanggaran siswa
mencapai batas maks pelanggaran dari pengaturan (bawaan 2; pelanggaran pertama
= peringatan, pelanggaran mencapai batas = blokir). Saat diblokir, sisa waktu
siswa dibekukan agar buka blokir dapat melanjutkan dari sisa yang sama.
Jawaban dan tanda ragu yang sudah tersimpan TIDAK dihapus oleh blokir.

### Alur utama — 1. Mengusulkan buka blokir

1. Pengawas (atau proktor/admin) mengajukan usulan untuk satu pengerjaan ujian
   yang diblokir, dengan menyertakan alasan.
2. Sistem memeriksa peran: hanya pengawas / proktor / admin yang boleh mengusulkan.
3. Sistem memeriksa isian: pengenal pengerjaan dan alasan wajib diisi (alasan
   kosong setelah dipangkas → ditolak).
4. Sistem memeriksa pengerjaan: tidak ditemukan → ditolak ("tidak ditemukan");
   pengerjaan bukan dari jadwal yang dimaksud → ditolak ("bukan dari jadwal ini");
   pengerjaan tidak berstatus diblokir → ditolak ("Hanya attempt yang diblokir
   yang bisa diusulkan…").
5. Sistem memeriksa usulan yang menunggu: bila sudah ada usulan berstatus
   menunggu untuk pengerjaan yang sama → ditolak ("Sudah ada usulan unblock
   yang menunggu untuk attempt ini."). Satu pengerjaan hanya boleh punya
   satu usulan menunggu.
6. Bila lolos, usulan disimpan berstatus **menunggu** dengan pencatat
   (pengaju) dan alasan; jejak audit usulan ditulis; perubahan status
   disiarkan ke pemantau jadwal.
7. **Di tampilan, TIDAK ditemukan tombol/formulir untuk membuat usulan** —
   pembuatan usulan hanya tersedia di layanan belakang. Halaman pemantau hanya
   membaca daftar usulan (lihat langkah 8). Lihat § "Hal yang tidak ditemukan".
8. Halaman pemantau memuat daftar usulan jadwal itu setiap 10 detik dan
   menampilkan usulan yang masih menunggu sebagai teks
   "nama siswa — alasan" (tanpa tombol setuju/tolak). Daftar usulan lengkap
   (termasuk yang sudah diproses, pemroses, dan waktu proses) tersedia di
   layanan belakang.

### Alur utama — 2. Memproses buka blokir (eksekusi)

1. Proktor/admin membuka halaman pemantau jadwal dan melihat baris peserta
   berstatus diblokir. Pada baris itu tersedia kolom alasan (wajib diisi) dan
   tombol Buka Blokir.
2. Proktor/admin mengisi alasan buka blokir lalu menekan Buka Blokir.
   Alasan kosong (di tampilan maupun di sistem) → ditolak; tampilan menampilkan
   "Alasan unblock wajib diisi." dan permintaan TIDAK dikirim.
3. Sistem memeriksa peran: hanya proktor / admin yang boleh mengeksekusi.
   Pengawas yang mencoba mengeksekusi langsung → ditolak (pengawas hanya
   boleh mengusulkan — pemisahan tugas pengawas-mengajukan / proktor-mengeksekusi).
4. Sistem memeriksa isian: pengenal pengerjaan dan alasan wajib diisi.
5. Sistem memeriksa pengerjaan: tidak ditemukan → ditolak; bukan dari jadwal
   yang dimaksud → ditolak; tidak berstatus diblokir → ditolak
   ("Hanya attempt yang diblokir yang bisa di-unblock…").
6. Sistem menghitung sisa waktu lanjutan: bila ada sisa waktu beku pada
   pengerjaan → dipakai sisa beku itu; bila tidak ada → sisa = jam selesai
   jadwal dikurangi waktu sekarang. Bila sisa ≤ 0 → ditolak
   ("Ujian sudah berakhir… Unblock tidak bisa dilakukan.").
7. Dalam satu transaksi, sistem: mengembalikan status pengerjaan menjadi
   **berjalan**, menetapkan tenggat baru = waktu sekarang + sisa waktu,
   mengosongkan sisa waktu beku — **jawaban tersimpan TIDAK direset**;
   lalu menutup usulan: bila ada usulan menunggu terbaru → statusnya menjadi
   **disetujui** dengan pencatat pemroses dan waktu proses; bila tidak ada
   usulan (proktor buka blokir langsung) → dibuat baris usulan berstatus
   disetujui sebagai jejak audit.
8. Jejak audit buka blokir ditulis (berisi alasan + sisa waktu detik);
   perubahan status disiarkan ke pemantau.
9. Tampilan menampilkan "Peserta di-unblock — ujian dilanjutkan dari sisa waktu."
   dan mengosongkan kolom alasan pada baris itu.

### Alur utama — 3. Siswa melanjutkan sesudah dibuka blokir

1. Siswa mengetahui status dibuka blokir lewat pemeriksaan status berkala
   (30 detik) atau lewat permintaan berikutnya (memuat soal / menyimpan jawaban);
   **tidak ada pemberitahuan dorong ke siswa**.
2. Pengerjaan kembali berstatus berjalan dengan tenggat baru sesuai sisa waktu.
3. Siswa melanjutkan dari soal terakhir; jawaban dan tanda ragu yang tersimpan
   sebelum diblokir tetap utuh.
4. Bila jadwal sudah berakhir (sisa waktu habis) saat buka blokir diminta,
   buka blokir ditolak — siswa tidak dapat melanjutkan.

### Cabang / kondisi (ringkas)

| Kondisi | Hasil akhir |
|---|---|
| Usulan: peran bukan pengawas/proktor/admin | Ditolak |
| Usulan: pengenal pengerjaan / alasan kosong | Ditolak |
| Usulan: pengerjaan tidak ditemukan / bukan dari jadwal ini | Ditolak ("tidak ditemukan" / "bukan dari jadwal ini") |
| Usulan: pengerjaan tidak diblokir | Ditolak |
| Usulan: sudah ada usulan menunggu untuk pengerjaan yang sama | Ditolak (satu usulan menunggu per pengerjaan) |
| Usulan: lolos semua | Usulan tersimpan berstatus menunggu; jejak audit; siaran ke pemantau |
| Buka blokir: alasan kosong | Ditolak (tampilan tidak mengirim; sistem menolak) |
| Buka blokir: peran pengawas | Ditolak (hanya proktor/admin; pengawas hanya mengusulkan) |
| Buka blokir: pengerjaan tidak diblokir | Ditolak |
| Buka blokir: sisa waktu ≤ 0 (jadwal sudah berakhir) | Ditolak |
| Buka blokir: lolos semua | Status kembali berjalan, tenggat baru = sekarang + sisa, jawaban utuh, usulan menunggu menjadi disetujui (atau dibuat jejak usulan disetujui bila tidak ada usulan) |

### Validasi / aturan bisnis yang terlihat di kode

- Pemisahan tugas: pengawas mengajukan, proktor/admin mengeksekusi.
  Tidak ada status usulan "ditolak" di kode — usulan hanya berstatus
  menunggu atau disetujui (usulan yang tidak diproses tetap menunggu).
- Satu pengerjaan hanya boleh punya satu usulan menunggu.
- Alasan wajib diisi pada usulan maupun pada eksekusi buka blokir.
- Buka blokir = LANJUT sisa waktu (beku bila ada, bila tidak sisa sampai
  jam selesai jadwal), bukan mulai ulang dan bukan reset jawaban.
- Buka blokir yang berhasil selalu menutup/mencatat usulan (menjadi disetujui),
  sehingga setiap eksekusi punya jejak usulan.
- Siswa tidak menerima pemberitahuan dorong; status baru diketahui lewat
  pemeriksaan status berkala atau permintaan berikutnya.

---

## Hal yang tidak ditemukan di kode (jangan dianggap ada)

1. **Tidak ada halaman daftar jejak audit di tampilan.** Layanan baca jejak
   audit ada (dengan saringan aksi/entitas/pelaku), tetapi tidak ada satu pun
   pemanggilan layanan itu dari kode tampilan.
2. **Tidak ada formulir pengecualian SoD di tampilan.** Layanan pencatatan
   pengecualian ada, tetapi tidak ada satu pun pemanggilan layanan itu dari
   kode tampilan.
3. **Tidak ada tombol/formulir membuat usulan buka blokir di tampilan.**
   Layanan pembuatan usulan ada, tetapi kode tampilan hanya MEMBACA daftar
   usulan (di halaman pemantau, polling 10 detik, tampil sebagai teks tanpa
   tombol aksi). Pengawas tidak punya tombol "usulkan buka blokir" di tampilan
   yang dibaca.
4. **Tidak ada halaman kelola kompetensi / kisi-kisi di tampilan.**
   Operasi kelola kompetensi dan kisi-kisi ada di layanan belakang (dengan izin
   mapel), tetapi halaman Kompetensi yang ada di tampilan hanya menampilkan
   nilai per kompetensi (baca saja).
5. **Tidak ada status usulan "ditolak".** Usulan buka blokir di kode hanya
   mengenal menunggu dan disetujui; tidak ada penolakan usulan.
6. **Tidak ada pemulihan hapus (hapus lunak / tempat sampah).** Semua hapus
   master data adalah hapus keras permanen; perlindungan satu-satunya adalah
   guard pemakaian dan dialog konfirmasi.
7. **Tidak ada penguncian login berbasis basis data.** Penguncian login
   (5 gagal / 15 menit) disimpan di memori proses — tidak bertahan bila proses
   dimulai ulang.
8. **Tidak ada pemberitahuan dorong ke siswa** untuk buka blokir (maupun untuk
   notifikasi umum). Siswa mengetahui status lewat pemeriksaan berkala;
   notifikasi hanya dibaca lewat bel / daftar tarik.
9. **Tidak ada pembuatan akun staf dari tampilan master data.**
   Akun login dibuat otomatis hanya untuk siswa. Guru hanya dapat dikaitkan
   ke pengenal akun yang sudah ada (diketik manual); tidak ada daftar pilih
   akun di tampilan guru dan tidak ada layanan daftar akun untuk itu.
10. **Tidak ada verifikasi bahwa impor siswa CSV memakai pratinjau (dry-run).**
    Import siswa langsung menyimpan baris valid dalam satu transaksi; tidak ada
    mode pratinjau sebelum simpan.
