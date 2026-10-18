# ALUR ACTIVITY NYATA — CBT SMK v2 (Gabungan UC01–UC19)

> **Status:** Dokumen telusur gabungan. Digabung dari tiga dokumen bagian:
> Bagian A (UC01, UC02, UC17, UC18), Bagian B (UC03, UC04, UC05, UC06, UC09, UC15),
> Bagian C (UC07, UC08, UC10, UC11, UC12, UC13, UC14, UC16, UC19).
> **Tanggal telusur:** 18 Okt 2026.
> **Cara baca:** Setiap UC berdiri sendiri dan berurutan UC01–UC19. Isi substansi
> bagian-bagian dipertahankan apa adanya; penggabung hanya merapikan struktur,
> menggabungkan pendahuluan dan penutup, menghilangkan duplikasi antar-bagian pada
> framing, dan menambahkan catatan yang ditandai eksplisit `[Catatan penggabung]`.
> Bila ada pertentangan antar-bagian, dicatat di § Pertentangan antar-bagian —
> tidak dipilih diam-diam salah satu.
> **Aturan penulisan:** Bahasa desain (PIM). TIDAK ada nama berkas, alamat layanan,
> perintah basis data, pustaka, nama tabel/kolom, atau nama fungsi — yang ditulis
> adalah perilaku yang terlihat dari kode: aktor, langkah, kondisi, hasil akhir,
> dan aturan bisnis.
> **Batasan:** Dokumen ini read-only terhadap model diagram; tidak mengubah kode
> aplikasi dan tidak menulis berkas diagram (`.mdj`).

## Konvensi global (berlaku untuk semua UC)

- Satu jadwal = satu paket soal untuk satu kelas pada satu tanggal dan rentang jam.
  Peserta hanya dapat mengerjakan jadwal kelasnya sendiri.
- Satu pengerjaan (attempt) per siswa per jadwal. Status pengerjaan: berjalan,
  diblokir, atau selesai.
- Nilai akhir = jumlah poin yang diperoleh dibagi jumlah bobot soal yang dinilai,
  dikali 100, dijepit ke rentang 0–100. Soal yang tidak dinilai (mis. survei sikap)
  tidak masuk penyebut.
- Predikat nilai: di bawah 70 = Belum Tuntas; 70–79 = Cukup; 80–90 = Baik;
  91–100 = Sangat Baik. Lulus bila nilai akhir ≥ 70.
- Setiap aksi ubah data menuntut pengguna sudah masuk dan memiliki izin untuk aksi
  itu; tanpa izin, aksi ditolak.
- Setiap aksi penting dicatat dalam jejak audit (siapa, aksi apa, kapan, untuk
  entitas mana). Isi jawaban siswa tidak ditulis ke jejak audit. Kegagalan menulis
  jejak audit TIDAK menggagalkan operasi utama, kecuali dinyatakan lain pada UC
  yang bersangkutan.
- Penghapusan dan penguncian selalu meminta konfirmasi dulu dan bersifat permanen
  (hapus keras, tanpa tempat sampah / pulihkan), kecuali dinyatakan lain.
- Pesan penolakan menjelaskan alasannya dalam bahasa pengguna.
- Semua waktu memakai zona waktu Asia/Jakarta.

## Daftar 19 UC

| Kode | Nama UC (persis dari diagram Use Case) | Sumber bagian |
|---|---|---|
| UC01 | Masuk & Wajib Ganti Kata Sandi | A |
| UC02 | Kelola Master Data | A |
| UC03 | Kelola Bank Soal | B |
| UC04 | Impor Soal | B |
| UC05 | Susun Paket Ujian | B |
| UC06 | Kelola Jadwal Ujian | B |
| UC07 | Pantau Ujian Langsung | C |
| UC08 | Ikuti Ujian | C |
| UC09 | Validasi Token | B |
| UC10 | Koreksi Esai / Manual | C |
| UC11 | Rekap Nilai & Berita Acara | C |
| UC12 | Analisis Butir Soal | C |
| UC13 | Nilai per Kompetensi | C |
| UC14 | Kelola UKK | C |
| UC15 | Kelola Kegiatan, Ruangan & Kartu Peserta | B |
| UC16 | Cadangkan & Pulihkan | C |
| UC17 | Kelola Pengaturan, Audit & Notifikasi | A |
| UC18 | Usul & Proses Buka Blokir | A |
| UC19 | Lihat Hasil & Pembahasan | C |

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

## UC03 — Kelola Bank Soal (tambah / ubah / hapus / kunci soal, semua tipe soal)

### Aktor

| Aktor | Hak |
|---|---|
| Admin | Semua aksi: bank + soal + kunci/buka kunci |
| Guru | Kelola bank miliknya sendiri dan bank tanpa pemilik; kelola soal di bank tersebut; kunci/buka kunci bank orang lain (lihat aturan pemisahan tugas) |
| Proktor | Hanya melihat daftar bank (metadata: kode, nama, kelas, jumlah soal, status kunci). TIDAK boleh melihat isi soal sama sekali |
| Pengawas, siswa, penguji UKK | Tidak punya akses ke bank soal |

### Konsep dasar

- Satu bank soal = satu mata pelajaran + satu tingkat kelas (10, 11, atau 12).
- Satu soal terdiri atas: tipe, stimulus (bacaan/kasus), pertanyaan, media (gambar/video), jawaban tampil (yang dilihat peserta), kunci (hanya guru), pembahasan (penjelasan guru), bobot, level, status.
- Status soal: Draf, Aktif, Nonaktif, Perlu Revisi, Diarsipkan.
- Kunci jawaban tidak pernah dikirim ke peserta; hanya guru/admin yang melihatnya (di pratinjau guru).
- Pembahasan hanya tampil ke peserta pada Tryout atau bila hasil ditampilkan sesudah ujian.

### Alur utama

#### 03a. Melihat daftar bank soal
1. Pengguna membuka halaman Bank Soal.
2. Sistem menampilkan daftar bank: kode, nama, kelas, jumlah soal, status kunci (Terkunci / Terbuka).
3. Pengguna dapat mencari bank berdasarkan kode / nama.
4. Hasil akhir: daftar bank tampil (atau pesan kosong bila belum ada bank).

#### 03b. Membuat bank soal
1. Pengguna menekan "Buat Bank Soal".
2. Pengguna mengisi: kode bank, nama bank, mata pelajaran (pilih dari daftar), tingkat kelas (10/11/12).
3. Pengguna menekan Simpan.
4. Sistem memeriksa: kode dan nama wajib diisi; mata pelajaran wajib dipilih dan harus ada; tingkat kelas harus 10, 11, atau 12; kode belum dipakai bank lain.
5. Bila lolos, bank baru tersimpan. Guru yang membuat otomatis menjadi pemilik bank; admin dapat menetapkan pemilik.
6. Hasil akhir: pengguna diarahkan ke halaman detail bank baru (siap diisi soal).

#### 03c. Mengubah bank soal
1. Dari halaman detail bank, pengguna menekan "Ubah Bank".
2. Formulir terisi data lama (kode, nama, mata pelajaran, tingkat kelas).
3. Pengguna mengubah lalu menyimpan; pemeriksaan sama seperti saat membuat.
4. Hasil akhir: data bank diperbarui. Catatan: mengubah bank TIDAK mengubah/membuka kunci bank; kunci adalah aksi tersendiri (03j).

#### 03d. Menghapus bank soal
1. Dari halaman detail bank, pengguna menekan "Hapus Bank" dan mengonfirmasi.
2. Sistem memeriksa: bank tidak terkunci; tidak ada satu pun soal bank ini yang tersusun di paket ujian mana pun.
3. Bila lolos, bank beserta SELURUH soalnya dihapus permanen.
4. Hasil akhir: bank hilang dari daftar; pengguna kembali ke daftar bank.

#### 03e. Melihat isi bank + menyaring soal
1. Pengguna membuka detail sebuah bank.
2. Sistem menampilkan identitas bank + daftar soalnya (nomor urut, tipe, level, bobot, status, cuplikan pertanyaan).
3. Pengguna dapat menyaring berdasarkan tipe soal, level, dan kata kunci isi pertanyaan.
4. Hasil akhir: daftar soal tampil sesuai saringan.

#### 03f. Menambah soal
1. Pengguna menekan "Tambah Soal" (tombol nonaktif bila bank terkunci).
2. Editor soal terbuka dengan nilai awal: tipe Pilihan Ganda, 4 opsi kosong (A–D), kunci A, bobot 1, level Sedang, status Aktif.
3. Pengguna memilih tipe soal; editor menampilkan isian yang sesuai tipe (lihat "Isian editor per tipe" di bawah).
4. Pengguna mengisi stimulus, pertanyaan, media, jawaban tampil, kunci, pembahasan, bobot, level, status.
5. Pengguna dapat menekan pratinjau untuk melihat soal persis seperti yang akan dilihat peserta (tanpa kunci) — pratinjau guru terpisah menampilkan kunci.
6. Pengguna menekan "Simpan Soal".
7. Sistem memeriksa (lihat "Pemeriksaan soal" di bawah).
8. Hasil akhir: soal baru tersimpan dan muncul di daftar; atau penolakan dengan pesan kesalahan.

#### 03g. Mengubah soal
1. Pengguna menekan "Ubah" pada sebuah soal (nonaktif bila bank terkunci).
2. Editor terbuka terisi data soal lama.
3. Pengguna mengubah lalu menyimpan; pemeriksaan sama seperti saat menambah.
4. Hasil akhir: soal diperbarui.
5. Cabang khusus — ganti tipe soal di tengah penyuntingan:
   - Bila jawaban/kunci masih kosong (belum diisi / masih nilai awal), tipe langsung berganti dan isian jawaban + kunci diisi ulang dengan nilai awal tipe baru.
   - Bila jawaban/kunci sudah berisi sesuatu, sistem meminta konfirmasi dulu, karena ganti tipe mengosongkan jawaban dan kunci. Bila pengguna membatalkan, tipe lama dipertahankan beserta isinya.

#### 03h. Menghapus soal
1. Pengguna menekan "Hapus" pada sebuah soal (nonaktif bila bank terkunci) dan mengonfirmasi.
2. Sistem memeriksa bank tidak terkunci.
3. Hasil akhir: soal dihapus permanen dari bank.

#### 03i. Pratinjau soal (guru)
1. Pengguna menekan "Pratinjau" pada sebuah soal.
2. Sistem menampilkan soal lengkap BESERTA kunci (tampilan guru).
3. Hasil akhir: pratinjau tampil; tidak ada perubahan data.

#### 03j. Mengunci / membuka kunci bank
1. Pengguna menekan "Kunci Bank" (bila terbuka) atau "Buka Kunci" (bila terkunci).
2. Mengunci meminta konfirmasi ("Soal tidak bisa diubah sampai bank dibuka lagi"); membuka kunci langsung berjalan.
3. Sistem memeriksa aturan pemisahan tugas: pembuat bank TIDAK boleh mengunci/membuka bank buatannya sendiri — yang mengunci harus guru lain atau admin.
4. Bila bank sudah berada pada keadaan yang diminta, sistem mengembalikan keadaan bank apa adanya (tidak error).
5. Hasil akhir: status kunci bank berubah.
6. Selama bank terkunci: tambah/ubah/hapus soal ditolak, impor soal ditolak, hapus bank ditolak. Tombol-tombol terkait di tampilan juga dinonaktifkan (penegakan utama tetap di sistem, bukan hanya di tombol).

### Isian editor per tipe (apa yang diisi pengguna)

| Tipe | Isian jawaban tampil | Isian kunci |
|---|---|---|
| Pilihan Ganda, Studi Kasus, Pilihan Gambar | Daftar opsi (label + teks; untuk Pilihan Gambar teks opsi berupa tautan gambar) | Satu label opsi |
| PG Kompleks, PG Multi-Jawaban | Daftar opsi | Beberapa label opsi (dicentang) |
| Benar-Salah | Daftar pernyataan | Nilai Benar/Salah per pernyataan |
| Menjodohkan, Peta Konsep | Daftar pasangan kiri → kanan | Pasangan yang ditandai "Benar" (mulai dari daftar kosong; pasangan kosong tidak pernah terkirim) |
| Isian / Jawaban Singkat | (tidak ada isian jawaban tampil) | Daftar jawaban alternatif, satu per baris |
| Uraian / Esai, Upload Berkas, Praktik Singkat | (tidak ada isian jawaban tampil) | Rubrik penilaian (boleh kosong untuk draf) |
| Mengurutkan | Daftar butir | Urutan benar, dipilih per posisi |
| Isian Multi-Kolom | Daftar label kolom | Satu jawaban per kolom |
| Teks Rumpang | (pertanyaan berisi tanda `___`) | Satu jawaban per rumpang |
| Skala Sikap | Daftar pernyataan + skala 1–4 | Skala 1–4 (tidak dinilai) |

Perilaku editor lainnya:
- Opsi jawaban dapat ditambah (sampai 8 opsi, label A–H) dan dihapus (minimal 2 opsi); label diurutkan ulang otomatis sesudah hapus.
- Pernyataan, pasangan, butir, dan kolom dapat ditambah/dihapus; pada Menjodohkan, kunci ikut diperbarui/dihapus mengikuti perubahan pasangan (tidak dibiarkan basi).
- Stimulus, pertanyaan, dan pembahasan ditulis dengan editor teks kaya (format, tabel, gambar, video, rumus).
- Media soal dapat berupa gambar yang diunggah atau tautan video; tautan video ditampilkan sebagai video tertanam di pratinjau, bukan sebagai gambar rusak.

### Pemeriksaan soal (berlaku untuk tambah, ubah, dan impor)

Pemeriksaan umum:
- Tipe harus salah satu dari 16 tipe yang dikenal (lihat bawah).
- Pertanyaan wajib diisi (tidak boleh kosong / hanya spasi).
- Bobot harus bilangan bulat > 0.
- Level harus Mudah / Sedang / Sulit (kosong dianggap Sedang).
- Status harus salah satu dari Draf / Aktif / Nonaktif / Perlu Revisi / Diarsipkan (kosong: Aktif saat tambah, dipertahankan saat ubah).
- Stimulus wajib diisi untuk tipe Studi Kasus.
- Untuk Teks Rumpang: jumlah tanda `___` di pertanyaan harus sama dengan jumlah jawaban kunci.

Pemeriksaan struktur jawaban/kunci per tipe:
- Pilihan Ganda / Studi Kasus / Pilihan Gambar: daftar opsi wajib ada, label tidak boleh duplikat; kunci harus satu label yang ada di daftar opsi; untuk Pilihan Gambar setiap opsi wajib berisi tautan gambar.
- PG Kompleks / PG Multi-Jawaban: kunci wajib daftar label tidak kosong, dan setiap label harus ada di daftar opsi.
- Benar-Salah: daftar pernyataan wajib tidak kosong; jumlah nilai kunci harus sama dengan jumlah pernyataan.
- Skala Sikap: daftar pernyataan wajib tidak kosong; skala kunci wajib daftar angka tidak kosong.
- Menjodohkan / Peta Konsep: setiap pasangan (jawaban tampil maupun kunci) wajib kiri dan kanan terisi; setiap pasangan kunci harus ada di pasangan jawaban tampil.
- Isian: kunci wajib daftar teks tidak kosong; tidak boleh berisi teks kosong.
- Uraian / Upload Berkas / Praktik Singkat: rubrik boleh kosong (draf) — disarankan dilengkapi sebelum bank dikunci.
- Mengurutkan: daftar butir minimal 2; kunci wajib daftar indeks sepanjang daftar butir, tanpa duplikat, dan setiap indeks dalam rentang.
- Isian Multi-Kolom: jumlah jawaban kunci harus sama dengan jumlah label kolom.
- Teks Rumpang: kunci wajib daftar jawaban per rumpang tidak kosong.

### 16 tipe soal (5 inti + 11 varian)

| # | Tipe | Kelompok | Penilaian | Masuk nilai |
|---|---|---|---|---|
| 1 | Pilihan Ganda | Inti AKM | Otomatis (benar semua / 0) | Ya |
| 2 | PG Kompleks | Inti AKM | Otomatis proporsional, tanpa penalti | Ya |
| 3 | Menjodohkan | Inti AKM | Otomatis proporsional per pasangan | Ya |
| 4 | Isian / Jawaban Singkat | Inti AKM | Otomatis (cocok persis setelah normalisasi: huruf kecil, spasi dirapatkan) | Ya |
| 5 | Uraian / Esai | Inti AKM | Manual oleh guru | Ya |
| 6 | Benar-Salah | Varian aplikasi | Otomatis proporsional per pernyataan | Ya |
| 7 | PG Multi-Jawaban | Varian aplikasi | Otomatis proporsional, tanpa penalti | Ya |
| 8 | Mengurutkan | Varian aplikasi | Otomatis proporsional per posisi | Ya |
| 9 | Isian Multi-Kolom | Varian aplikasi | Otomatis proporsional per kolom | Ya |
| 10 | Teks Rumpang | Varian aplikasi | Otomatis proporsional per rumpang | Ya |
| 11 | Peta Konsep | Varian aplikasi | Otomatis proporsional per pasangan | Ya |
| 12 | Studi Kasus + PG | Varian aplikasi | Otomatis (benar semua / 0) | Ya |
| 13 | Skala Sikap | Varian aplikasi | Non-nilai (survei/refleksi) | Tidak |
| 14 | Upload Berkas | Varian aplikasi | Manual oleh guru/penguji | Ya |
| 15 | Praktik Singkat | Varian aplikasi | Manual dengan rubrik | Ya |
| 16 | Pilihan Gambar | Varian aplikasi | Otomatis (benar semua / 0) | Ya |

Label jujur yang dipakai tampilan: "5 tipe standar AKM + 11 varian/ekstensi aplikasi" (bukan "16 tipe standar").

### Cabang / hasil akhir UC03 (ringkas)

| Kondisi | Hasil |
|---|---|
| Kode bank sudah dipakai | Ditolak: kode sudah dipakai |
| Mata pelajaran tidak dipilih / tidak ditemukan | Ditolak: mata pelajaran wajib / tidak ditemukan |
| Tingkat kelas bukan 10/11/12 | Ditolak |
| Guru menyentuh bank milik guru lain | Ditolak: bank milik guru lain |
| Proktor membuka isi soal | Ditolak: proktor tidak boleh melihat isi soal |
| Bank terkunci + tambah/ubah/hapus/impor soal | Ditolak: bank terkunci |
| Bank terkunci + hapus bank | Ditolak: buka kunci dulu |
| Bank dipakai paket + hapus bank | Ditolak: bank dipakai paket ujian |
| Pembuat bank mengunci/membuka bank sendiri | Ditolak: pemisahan tugas — minta guru lain atau admin |
| Soal gagal pemeriksaan struktur | Ditolak dengan pesan per aturan yang dilanggar |
| Semua pemeriksaan lolos | Data tersimpan + tercatat di jejak audit |

### Hal yang tidak ditemukan di kode untuk UC03

- Tidak ada salin/duplikat bank soal, tidak ada hapus banyak soal sekaligus, tidak ada urut-ulang soal di dalam bank (urutan tampil = urutan pembuatan), tidak ada riwayat revisi / pulihkan soal yang dihapus, tidak ada penelaahan/validasi berjenjang (soal langsung tersimpan sesuai isian editor).

---

## UC04 — Impor Soal (Excel + Word: template, preview, validasi, simpan)

> [Catatan penggabung] Penyebutan "halaman Kisi-kisi / Kompetensi" pada bagian format Word di bawah vs pernyataan UC02 bahwa tidak ada halaman kelola kompetensi/kisi-kisi — lihat § Pertentangan antar-bagian, butir 3.

### Aktor

| Aktor | Hak |
|---|---|
| Admin | Impor ke bank mana pun |
| Guru | Impor ke bank miliknya sendiri dan bank tanpa pemilik |
| Proktor | Ditolak (tidak boleh melihat isi soal) |
| Pengawas, siswa, penguji UKK | Tidak punya akses |

### Alur utama

1. **Unduh template.** Pengguna di panel impor memilih format Excel atau Word, lalu menekan "Unduh Template". Sistem membuatkan berkas template dan pengguna mengunduhnya. Satu berkas = satu bank soal; maksimal 500 soal, maksimal 5 MB.
2. **Isi template.** Pengguna mengisi berkas template (lihat "Format berkas" di bawah). Jangan membuat berkas baru dari nol; jangan mengganti nama kolom / nama sheet.
3. **Pilih berkas.** Pengguna memilih format (Excel/Word), lalu memilih berkas. Tampilan memeriksa sebelum dikirim: ekstensi harus sesuai format yang dipilih, berkas tidak boleh kosong, ukuran tidak boleh melebihi 5 MB. Bila gagal, berkas ditolak dengan pesan dan tidak dikirim.
4. **Pratinjau (dry-run).** Pengguna menekan "Pratinjau". Berkas dikirim untuk diperiksa; TIDAK ada soal yang tersimpan pada tahap ini. Sistem mengembalikan ringkasan + tabel pratinjau + daftar temuan.
5. **Baca hasil pratinjau.** Tampilan menunjukkan: total baris, jumlah valid, jumlah error, jumlah warning, jumlah duplikat, total bobot baris valid, ringkasan per tipe, daftar error, daftar warning, detail duplikat, dan tabel pratinjau per baris (dengan kunci yang sudah dipetakan ke label + teks opsi yang dirujuk, mis. "B — Semarang", agar pengguna memastikan huruf kunci menunjuk opsi yang dimaksud). Baris yang error ditandai; baris dengan warning tetap ditampilkan.
6. **Pilih mode + perlakuan duplikat.** Pengguna memilih mode simpan dan perlakuan duplikat (lihat bawah).
7. **Konfirmasi simpan.** Pengguna menekan "Simpan". Berkas yang sama dikirim ulang dan diperiksa ulang dari awal (hasil pratinjau tidak dipakai mentah-mentah sebagai dasar simpan — simpan selalu memvalidasi ulang berkas).
8. **Hasil.** Tampilan menunjukkan: jumlah tersimpan, jumlah gagal, jumlah duplikat dilewati, daftar warning, daftar error, dan status tersimpan / tidak ada yang tersimpan. Bila ada yang tersimpan, daftar bank diperbarui.
9. **Telaah.** Soal hasil impor masuk berstatus DRAF. Soal wajib ditelaah (kunci, opsi, stimulus, gambar, rubrik) di editor sebelum dipakai di paket ujian.

### Format berkas — Excel

- Berkas berisi 4 lembar: "Soal" (satu-satunya yang diisi), "Contoh", "Referensi Tipe", "Petunjuk".
- Lembar "Soal" memakai 16 kolom kanonik, dalam urutan ini: tipe, pertanyaan, stimulus, opsi_a, opsi_b, opsi_c, opsi_d, opsi_e, kunci, bobot, level, pembahasan, media_url, media_alt, kisi_kisi, kompetensi.
- Satu baris = satu soal. Baris kosong dilewati (tidak dihitung).
- Kolom wajib di kepala: tipe, pertanyaan, kunci. Bila salah satu hilang → berkas ditolak sebagai error level berkas.
- Kolom kepala yang tidak dikenal → warning level berkas; kolom itu diabaikan.
- Nilai awal bila kosong: bobot = 1, level = Sedang.
- Cara menulis kunci per tipe (sama seperti Word):
  - Pilihan Ganda / Studi Kasus / Pilihan Gambar: satu label (mis. `B`).
  - PG Kompleks / PG Multi-Jawaban: label dipisah koma (mis. `A,C`).
  - Benar-Salah: `B`/`S` dipisah koma, sebanyak pernyataan (mis. `B,S,B`).
  - Menjodohkan / Peta Konsep: pasangan `kiri::kanan`, antar-pasangan dipisah titik koma (mis. `Bandung::Jawa Barat;Semarang::Jawa Tengah`).
  - Isian / Isian Multi-Kolom / Teks Rumpang: jawaban dipisah garis tegak `|` (mis. `Jakarta|DKI Jakarta`).
  - Mengurutkan: label dalam urutan benar, dipisah koma (mis. `B,C,A,D`).
  - Uraian / Upload Berkas / Praktik Singkat: tulis rubrik (boleh kosong → warning).
  - Skala Sikap: kolom kunci diabaikan.
- Kolom opsi_a–opsi_e: dipakai tipe pilihan; untuk tipe isian/uraian dan sejenisnya dikosongkan.
- Kolom media_url: SATU tautan gambar (`https://…`). Gambar yang ditempel langsung di berkas Excel TIDAK ikut terimpor — harus berupa tautan. Kolom media_alt: teks alternatif gambar; teks alternatif ini TIDAK tersimpan sebagai data soal (hanya dipakai untuk peringatan bila kosong saat media_url terisi).
- Kolom kompetensi: isi kode kompetensi yang tampil di halaman Kompetensi bank ini. Kolom kisi_kisi: saat ini SELALU gagal (lihat E30) — praktisnya kolom ini harus dikosongkan.

### Format berkas — Word

- Satu blok soal diawali baris `SOAL 1`, `SOAL 2`, dan seterusnya. Teks sebelum `SOAL 1` pertama dianggap petunjuk dan diabaikan.
- Bila dokumen tidak memakai label `SOAL`, pemisah antar-soal adalah garis pemisah (`---` dan sejenisnya).
- Nomor `SOAL` yang duplikat: blok kedua dan seterusnya diabaikan.
- Di dalam tiap blok, label ditulis persis diakhiri titik dua: `TIPE:`, `STIMULUS:`, `PERTANYAAN:`, opsi `A.`–`E.`, `KUNCI:`, `BOBOT:`, `LEVEL:`, `PEMBAHASAN:`, `KISI-KISI:`, `KOMPETENSI:`, `MEDIA:`, `ALT MEDIA:`.
- Label bernilai tunggal (TIPE, KUNCI, BOBOT, LEVEL, MEDIA, ALT MEDIA, KISI-KISI, KOMPETENSI) yang ditulis dua kali dalam satu blok → error.
- Opsi yang ditulis dua kali → error. Baris opsi multi-paragraf ditempel ke opsi terakhir.
- Baris sebelum label apa pun di dalam blok dianggap pertanyaan.
- Cara menulis KUNCI sama seperti Excel (per tipe). Bobot kosong = 1; level kosong = Sedang.
- `MEDIA:` bila terisi harus tautan `http(s)` yang valid. Gambar yang ditempel langsung di dokumen TIDAK ikut terimpor — harus berupa tautan. `ALT MEDIA:` = teks alternatif gambar tersebut.
- `KISI-KISI:` / `KOMPETENSI:` diisi kode yang tampil di halaman Kisi-kisi / Kompetensi bank ini; kosongkan bila belum dipetakan (KISI-KISI yang terisi selalu gagal — lihat E30).

### Pemeriksaan impor — error (menggagalkan baris)

| Kode | Arti |
|---|---|
| E1 | Tipe wajib diisi |
| E2 | Tipe tidak dikenal (harus kode, mis. `pg`, bukan nama panjang "Pilihan Ganda") |
| E3 | Pertanyaan wajib diisi |
| E4 | Kunci wajib diisi (untuk tipe yang dinilai otomatis; tidak berlaku untuk tipe rubrik dan Skala Sikap) |
| E5 | Bobot harus bilangan bulat > 0 |
| E6 | Level harus mudah / sedang / sulit |
| E7–E24 | Kesalahan bentuk jawaban/kunci per tipe (mis. E13: jumlah nilai B/S tidak sama dengan jumlah pernyataan; E17: pasangan kunci Menjodohkan salah format; label kunci tidak ada di daftar opsi; indeks urutan duplikat / di luar rentang; jumlah jawaban kolom tidak sama dengan jumlah kolom) |
| E25 | Jumlah jawaban tidak sama dengan jumlah tanda `___` (Teks Rumpang) |
| E26 | Stimulus kosong untuk Studi Kasus |
| E28 | `media_url` / `MEDIA:` bukan tautan `http(s)` yang valid |
| E29 | Opsi Pilihan Gambar bukan tautan gambar `http(s)` |
| E30 | Kode kisi_kisi tidak ditemukan — SELALU terjadi bila kolom kisi_kisi terisi (pemetaan kode kisi-kisi belum tersedia), sehingga kolom kisi_kisi praktis harus dikosongkan |
| E31 | Kode kompetensi tidak ditemukan di mata pelajaran bank ini (pencocokan kode tidak peka huruf besar/kecil) |
| E32 | Gagal menyimpan ke penyimpanan |

Error level berkas (bukan per baris): berkas bukan format Excel/Word yang valid (tanda pengenal berkas tidak cocok), berkas kosong, berkas melebihi 5 MB, lembar "Soal" tidak ditemukan (Excel), kolom wajib di kepala hilang, berkas melebihi 500 baris soal, berkas tidak punya kepala + minimal 1 baris soal.

### Pemeriksaan impor — warning (tidak menggagalkan)

| Kode | Arti |
|---|---|
| W1 | Kunci kosong untuk tipe rubrik — rubrik belum diisi, bisa dilengkapi di editor sesudah impor |
| W2 | `media_url` terisi tetapi `media_alt` kosong — gambar tanpa teks alternatif |
| W3 | Kolom kepala tidak dikenal — diabaikan |
| W4 | Pertanyaan duplikat dengan baris lain di berkas yang sama (disebutkan nomor baris pertamanya) |
| W5 | Pertanyaan duplikat dengan soal yang sudah ada di bank |
| W6 | Kunci diabaikan untuk Skala Sikap (skala 1–4, tidak dinilai) |
| W7 | Tautan gambar tidak diperiksa keterjangkauannya — klik tautan di pratinjau untuk memastikan gambar bisa dibuka |

Aturan duplikat:
- Duplikat dikenali dari teks pertanyaan yang dinormalisasi (spasi dipangkas/dirapatkan, huruf dikecilkan).
- Duplikat adalah warning, BUKAN error: tidak pernah menggagalkan berkas dalam mode apa pun.
- Penanganan duplikat dipilih pengguna saat konfirmasi: "Lewati duplikat" (bawaan — baris duplikat tidak disimpan, dihitung sebagai dilewati) atau "Simpan ganda" (duplikat tetap disimpan).
- Pengecualian pemeriksaan kunci Menjodohkan/Peta Konsep saat impor: kunci boleh memuat pasangan yang tidak ada di opsi tampil (pengecoh sisi kanan) — berbeda dari pemeriksaan editor soal manual yang menuntut pasangan kunci ada di pasangan tampil.

### Mode simpan

| Mode | Perilaku |
|---|---|
| Tolak semua bila ada error (bawaan, paling aman) | Bila ada error apa pun → TIDAK ADA soal yang tersimpan. Tampilan memblokir tombol konfirmasi selama masih ada error; pengguna memperbaiki berkas lalu mengunggah ulang, atau pindah ke mode lewati. |
| Lewati baris yang error | Baris valid tersimpan; baris error dilewati dan dilaporkan. |

Cabang hasil konfirmasi:
- Mode tolak-semua + ada error → tersimpan 0, gagal = jumlah error, status tidak tersimpan.
- Mode lewati + ada error → baris valid tersimpan, baris error dilaporkan sebagai gagal.
- Tidak ada baris valid → tidak ada yang tersimpan.
- Bank terkunci → pratinjau dan konfirmasi ditolak (bank terkunci — impor ditolak).

### Cabang / hasil akhir UC04 (ringkas)

| Kondisi | Hasil |
|---|---|
| Ekstensi berkas tidak sesuai format yang dipilih / berkas kosong / > 5 MB | Ditolak di tampilan, berkas tidak dikirim |
| Berkas bukan Excel/Word valid | Ditolak dengan pesan |
| Bank terkunci | Pratinjau + simpan ditolak |
| Lembar "Soal" hilang / kolom wajib hilang / > 500 baris | Error level berkas, tidak ada yang tersimpan |
| Baris gagal pemeriksaan | Error per baris; nasibnya mengikuti mode simpan |
| Baris duplikat | Warning; disimpan atau dilewati sesuai pilihan duplikat |
| Semua baris valid | Semua tersimpan sebagai soal Draf |

### Hal yang tidak ditemukan di kode untuk UC04

- Tidak ada pilihan impor CSV di panel (hanya Excel dan Word); tidak ada simpan sebagian baris pilihan (hanya mode tolak-semua / lewati-error); tidak ada sunting baris di dalam pratinjau (perbaiki berkas → unggah ulang); tidak ada impor gambar tempelan (hanya tautan); tidak ada pemetaan kode kisi-kisi (kolom kisi_kisi yang terisi selalu gagal); teks alternatif gambar tidak tersimpan; tidak ada impor gambar inline dokumen Word.

---

## UC05 — Susun Paket Ujian (paket + susun soal manual / acak)

### Aktor

| Aktor | Hak |
|---|---|
| Admin, guru | Buat / ubah / hapus paket; susun soal (manual + acak) |
| Proktor, pengawas, siswa, penguji UKK | Tidak punya akses (proktor tidak boleh melihat isi soal) |

### Konsep dasar

- Paket ujian = wadah soal untuk satu mata pelajaran: identitas + susunan soal.
- Jenis paket: STS, SAS, USBK, UKK, Tryout.
- Durasi paket hanya ditampilkan; batas waktu ujian peserta mengikuti jam selesai jadwal, bukan durasi paket.
- Jumlah soal paket selalu diselaraskan dengan jumlah susunan aktual setiap kali susunan berubah.
- Soal dalam paket berurutan; soal baru selalu ditambahkan di akhir dengan nomor urut berikutnya.
- Nilai tampil ke peserta hanya efektif bila saklar paket AKTIF dan saklar global tampil-hasil juga AKTIF (Tryout selalu menampilkan hasil lengkap — aturan tampil hasil diuraikan di alur ujian).

### Alur utama

#### 05a. Membuat paket
1. Pengguna menekan "Buat Paket".
2. Pengguna mengisi: kode, nama paket, mata pelajaran (pilih dari daftar), jenis (STS/SAS/USBK/UKK/Tryout), durasi menit, serta tiga saklar: acak urutan soal per siswa (bawaan Ya), acak urutan opsi jawaban per siswa (bawaan Ya), tampilkan hasil ke siswa (bawaan Tidak).
3. Pengguna menyimpan.
4. Sistem memeriksa: kode, nama, dan mata pelajaran wajib diisi; jenis harus salah satu dari 5 jenis; durasi harus > 0; jumlah soal tidak boleh negatif; mata pelajaran harus ada; kode belum dipakai paket lain.
5. Hasil akhir: paket baru tersimpan (masih kosong, belum berisi soal) dan muncul di daftar.

#### 05b. Mengubah paket
1. Pengguna menekan "Ubah" pada sebuah paket; formulir terisi data lama.
2. Pengguna mengubah lalu menyimpan; pemeriksaan sama seperti saat membuat.
3. Cabang khusus — ganti mata pelajaran: bila paket masih berisi soal dari mata pelajaran lain, perubahan ditolak ("kosongkan susunan soal dulu sebelum ganti mapel"). Bila susunan sudah kosong (atau semua soal sudah se-mapel dengan yang baru), perubahan diterima.
4. Hasil akhir: paket diperbarui.

#### 05c. Menghapus paket
1. Pengguna menekan "Hapus" pada sebuah paket dan mengonfirmasi ("Paket ini akan dihapus permanen beserta susunan soalnya").
2. Sistem memeriksa: paket tidak dipakai jadwal ujian mana pun.
3. Bila dipakai jadwal → ditolak ("hapus jadwal dulu").
4. Hasil akhir: paket beserta susunan soalnya dihapus permanen.

#### 05d. Menyusun soal manual
1. Pengguna membuka detail paket.
2. Sistem menampilkan info paket + susunan soal saat ini (nomor urut, tipe, pertanyaan, bobot) + kontrol susun.
3. Pengguna memilih bank soal dari daftar bank yang se-mata-pelajaran dengan paket.
4. Sistem menampilkan daftar soal bank tersebut.
5. Pengguna menekan "+ Tambah" pada sebuah soal.
6. Sistem memeriksa: paket ada; soal ada; soal berasal dari mata pelajaran yang sama dengan paket.
7. Bila lolos, soal ditambahkan di akhir susunan (nomor urut = nomor terbesar + 1). Menambahkan soal yang sudah ada di paket tidak menduplikasi (tidak ada perubahan).
8. Hasil akhir: susunan bertambah; jumlah soal paket terselaraskan.

#### 05e. Menghapus soal dari paket
1. Pengguna menekan "Hapus" pada sebuah soal dalam susunan.
2. Sistem memeriksa soal memang ada di paket.
3. Hasil akhir: soal terlepas dari paket (soal tetap ada di bank soalnya). Nomor urut soal lain TIDAK dirapatkan kembali; soal yang ditambahkan kemudian memakai nomor terbesar + 1.

#### 05f. Generate soal acak
1. Pengguna mengisi jumlah yang diminta (bawaan 10) lalu menekan "Generate Acak".
2. Sistem memeriksa jumlah > 0 dan paket ada.
3. Sistem mengambil soal secara acak dari bank-bank yang se-mata-pelajaran dengan paket, dengan syarat: soal berstatus Aktif dan belum ada di paket.
4. Soal terpilih ditambahkan berurutan di akhir susunan.
5. Cabang: bila soal yang tersedia kurang dari yang diminta → sistem menambahkan yang tersedia saja dan mengembalikan peringatan ("soal tersedia di bank mapel ini kurang dari yang diminta").
6. Hasil akhir: susunan bertambah sejumlah yang berhasil diambil + pesan (termasuk peringatan bila ada).

### Cabang / hasil akhir UC05 (ringkas)

| Kondisi | Hasil |
|---|---|
| Kode paket sudah dipakai | Ditolak: kode sudah dipakai |
| Jenis bukan salah satu dari 5 jenis / durasi ≤ 0 | Ditolak |
| Mata pelajaran tidak ditemukan | Ditolak |
| Paket dipakai jadwal + hapus paket | Ditolak: hapus jadwal dulu |
| Ganti mapel saat susunan berisi soal mapel lain | Ditolak: kosongkan susunan dulu |
| Soal dari mapel berbeda + tambah manual | Ditolak: soal bukan dari mapel paket ini |
| Soal sudah ada di paket + tambah lagi | Tidak menduplikasi (tanpa perubahan) |
| Generate acak, soal tersedia kurang dari diminta | Tambah yang tersedia + peringatan |
| Generate acak, jumlah ≤ 0 | Ditolak: jumlah harus > 0 |

### Hal yang tidak ditemukan di kode untuk UC05

- Tidak ada urut-ulang soal dalam paket (naik/turun/seret), tidak ada salin/duplikat paket, tidak ada pratinjau paket sebagai peserta, tidak ada tambah soal banyak sekaligus selain generate acak, tidak ada batas jumlah soal maksimum paket. Tambah manual tidak menyaring status soal (soal berstatus apa pun dapat ditambahkan manual; hanya generate acak yang membatasi ke soal Aktif).

---

## UC06 — Kelola Jadwal Ujian (jadwal + pause / resume jadwal)

### Aktor

| Aktor | Hak |
|---|---|
| Admin, proktor, guru | Buat / ubah / hapus / lihat jadwal |
| Pengawas | Lihat jadwal (tanpa ubah, tanpa token) |
| Admin, proktor | Pause / resume jadwal; generate token (lihat UC09) |
| Siswa | Tidak mengelola jadwal (hanya mengikuti jadwal kelasnya) |

### Konsep dasar

- Jadwal ujian = paket + kelas + tanggal + jam mulai + jam selesai + status.
- Status jadwal: Draf, Aktif, Jeda (pause), Selesai.
- Jadwal yang sudah Selesai tidak boleh diubah sama sekali.
- Jadwal yang sudah punya pengerjaan peserta tidak boleh dihapus (hasil ujian tidak pernah dihapus).
- Peserta pertama yang memulai ujian mengaktifkan jadwal Draf menjadi Aktif secara otomatis (bukan hanya lewat formulir).
- Tidak ada perubahan status otomatis ke Selesai — status Selesai hanya ditetapkan manual lewat formulir.

### Alur utama

#### 06a. Membuat jadwal
1. Pengguna menekan "Buat Jadwal".
2. Pengguna mengisi: paket (pilih dari daftar), kelas (pilih dari daftar), tanggal, jam mulai, jam selesai, status (bawaan Draf).
3. Pengguna menyimpan.
4. Sistem memeriksa: semua isian wajib; tanggal/jam valid; jam selesai harus sesudah jam mulai; status harus salah satu dari Draf/Aktif/Jeda/Selesai; paket dan kelas yang dipilih harus ada.
5. Hasil akhir: jadwal baru tersimpan.

#### 06b. Mengubah jadwal
1. Pengguna menekan "Ubah" pada sebuah jadwal; formulir terisi data lama.
2. Pengguna mengubah lalu menyimpan; pemeriksaan format sama seperti saat membuat.
3. Cabang: bila jadwal sudah berstatus Selesai → ditolak seluruhnya ("jadwal yang sudah selesai tidak boleh diubah").
4. Hasil akhir: jadwal diperbarui.
5. Perhatian: mengubah status lewat formulir ubah HANYA mengubah label status — tidak membekukan / memulihkan timer peserta. Pembekuan timer hanya terjadi lewat aksi Pause / Resume khusus (06e/06f). Mengubah status menjadi Jeda lewat formulir tidak sama dengan Pause.

#### 06c. Menghapus jadwal
1. Pengguna menekan "Hapus" pada sebuah jadwal dan mengonfirmasi.
2. Sistem memeriksa: tidak ada pengerjaan peserta apa pun untuk jadwal ini.
3. Bila sudah ada pengerjaan peserta → ditolak (hasil ujian tidak pernah dihapus).
4. Hasil akhir: jadwal dihapus permanen.

#### 06d. Melihat daftar jadwal
1. Pengguna membuka halaman Jadwal.
2. Sistem menampilkan daftar: paket, kelas, tanggal, jam, status, status token.
3. Pengguna dapat mencari berdasarkan nama paket dan menyaring berdasarkan status (Semua / Draf / Aktif / Jeda / Selesai).
4. Status token per jadwal tampil sebagai: belum dicek / belum dibuat / aktif (sisa detik) / kedaluwarsa (lihat UC09).
5. Hasil akhir: daftar tampil sesuai saringan.

#### 06e. Pause jadwal (jeda — timer dibekukan)
1. Proktor/admin menekan "Pause" pada jadwal Aktif (tombol hanya tampil untuk jadwal Aktif; juga tersedia dari halaman pemantauan).
2. Sistem memeriksa jadwal berstatus Aktif.
3. Sistem dalam satu kesatuan aksi: status jadwal menjadi Jeda + sisa waktu SEMUA pengerjaan yang sedang berjalan dibekukan (sisa = tenggat − sekarang, minimal 0). Sisa waktu pengerjaan yang sudah diblokir (karena pelanggaran) TIDAK ikut dibekukan di sini — sisa waktunya sudah dibekukan tersendiri saat pemblokiran.
4. Pemantau menerima kabar "jadwal dijeda".
5. Efek selama jeda: peserta tidak dapat memulai; jawaban tidak dapat disimpan; auto-submit ditunda; penghitung waktu peserta menampilkan nilai beku; peserta melihat pemberitahuan "Ujian dijeda proktor — timer dibekukan."
6. Hasil akhir: jadwal Jeda, timer peserta beku.

#### 06f. Resume jadwal (lanjutkan)
1. Proktor/admin menekan "Resume" pada jadwal Jeda (tombol hanya tampil untuk jadwal Jeda).
2. Sistem memeriksa jadwal berstatus Jeda.
3. Sistem dalam satu kesatuan aksi: status jadwal kembali Aktif + setiap pengerjaan berjalan yang punya sisa beku mendapat tenggat baru (sekarang + sisa beku) dan sisa bekunya dibersihkan.
4. Pemantau menerima kabar "jadwal dilanjutkan".
5. Hasil akhir: jadwal Aktif; peserta melanjutkan dari sisa waktu yang dibekukan; jawaban dan tanda ragu tersimpan utuh.

### Cabang / hasil akhir UC06 (ringkas)

| Kondisi | Hasil |
|---|---|
| Jam selesai tidak sesudah jam mulai / tanggal-jam tidak valid | Ditolak |
| Paket / kelas tidak ditemukan (saat buat) | Ditolak |
| Status bukan Draf/Aktif/Jeda/Selesai | Ditolak |
| Jadwal Selesai + ubah | Ditolak seluruhnya |
| Jadwal punya pengerjaan peserta + hapus | Ditolak (hasil tidak pernah dihapus) |
| Pause jadwal bukan Aktif | Ditolak |
| Resume jadwal bukan Jeda | Ditolak |
| Peserta pertama memulai jadwal Draf | Jadwal otomatis menjadi Aktif |
| Jadwal Selesai + peserta mulai / generate token | Ditolak |

### Hal yang tidak ditemukan di kode untuk UC06

- Tidak ada penjadwalan berulang (harian/mingguan), tidak ada deteksi bentrok (dua jadwal kelas/ruangan yang tumpang tindih tidak ditolak), tidak ada pengingat otomatis menjelang ujian, tidak ada pengalihan status otomatis ke Selesai saat jam selesai lewat (status Selesai hanya manual), tidak ada ubah massal jadwal.

---

## UC07 — Pantau Ujian Langsung

> [Catatan penggabung] Jeda/lanjutkan jadwal di UC ini adalah sudut pandang pemantau; definisi pemiliknya ada di UC06. Usulan/buka blokir adalah sudut pandang pemantau; definisi pemiliknya ada di UC18. Rancangan diagram menyelesaikan tumpang-tindih ini dengan simpul pemanggil (lihat RANCANGAN-ACTIVITY.md), bukan dengan menduplikasi alur.

**Aktor:** Proktor, Admin (memantau + menjeda/melanjutkan + membuka blokir). Pengawas dan Guru (memantau; Pengawas hanya boleh mengusulkan buka blokir, tidak membuka langsung). Siswa adalah sumber peristiwa yang dipantau, bukan pelaku alur ini.

### Alur utama

1. Staf membuka halaman pemantauan dan memilih jadwal yang akan dipantau.
2. Sistem menampilkan daftar peserta jadwal itu beserta, per peserta: nama dan identitas, status pengerjaan (berjalan / selesai / diblokir), jumlah soal dijawab dari total soal, jumlah tanda ragu, jumlah pelanggaran, dan sisa waktu.
3. Sistem menampilkan ringkasan: total peserta, jumlah sedang ujian, jumlah selesai, jumlah diblokir, serta total jawaban terkumpul dari total soal.
4. Sistem membuka saluran langsung (real-time) ke pemantau dengan tiket sekali pakai yang hanya berlaku singkat. Begitu tersambung, sistem langsung mengirim potret (snapshot) daftar peserta saat itu.
5. Selama ujian berjalan, setiap perubahan dari sisi siswa (jawaban tersimpan, tanda ragu berubah, pelanggaran tercatat, pengerjaan dimulai/diselesaikan, koreksi disimpan, jadwal dijeda/dilanjutkan, usulan buka blokir dibuat, buka blokir berhasil) disiarkan ke semua pemantau jadwal itu; pemantau memuat ulang potret daftar peserta.
6. Pemantau dapat menyaring daftar peserta menurut status: semua, sedang ujian, selesai, diblokir.
7. Pemantau menutup halaman → saluran langsung ditutup.

### Cabang / kondisi

| Kondisi | Perilaku |
|---|---|
| Tiket saluran langsung tidak ada / sudah dipakai / kedaluwarsa | Sambungan ditolak; pemantau tidak menerima siaran langsung |
| Saluran langsung terputus | Sistem mencoba sambung ulang otomatis dengan tiket baru setelah jeda singkat; selama saluran mati, daftar peserta diperbarui dengan penarikan berkala tiap beberapa detik dan label tampilan berubah menjadi mode penarikan berkala |
| Pemantau membuka halaman tanpa memilih jadwal | Sistem menampilkan daftar jadwal untuk dipilih, bukan daftar peserta |
| Jadwal dijeda (lihat langkah pause di bawah) | Semua timer peserta dibekukan; penyimpanan jawaban ditolak; peserta baru tidak dapat mulai; finalisasi otomatis ditunda |
| Jadwal dilanjutkan (resume) | Timer peserta berjalan dilanjutkan dari sisa waktu yang dibekukan; peserta dapat menyimpan jawaban lagi |

### Sub-alur 1 — Jeda jadwal (pause)

1. Proktor/Admin menekan tombol Jeda Jadwal (hanya aktif bila status jadwal sedang berjalan).
2. Sistem menolak bila status jadwal bukan berjalan.
3. Sistem mengubah status jadwal menjadi jeda dan membekukan sisa waktu setiap pengerjaan yang sedang berjalan (sisa waktu = batas waktu dikurangi waktu sekarang, tidak pernah negatif).
4. Sistem menyiarkan peristiwa jeda ke semua pemantau; pemantau memuat ulang daftar.
5. **Hasil akhir:** jadwal berstatus jeda; timer semua peserta berjalan membeku; jawaban tidak dapat disimpan; pengerjaan baru tidak dapat dimulai; finalisasi otomatis ditunda selama jeda.
6. **Catatan penting:** pengerjaan yang sedang diblokir saat jeda TIDAK ikut dibekukan/dilanjutkan oleh jeda/lanjut ini — sisa waktunya dibekukan tersendiri saat pemblokiran dan dipakai saat buka blokir.

### Sub-alur 2 — Lanjutkan jadwal (resume)

1. Proktor/Admin menekan tombol Lanjutkan Jadwal (hanya aktif bila status jadwal sedang jeda).
2. Sistem menolak bila status jadwal bukan jeda.
3. Sistem mengubah status jadwal menjadi berjalan dan menggeser batas waktu setiap pengerjaan yang sedang berjalan menjadi waktu sekarang ditambah sisa waktu yang dibekukan; penanda beku dihapus.
4. Sistem menyiarkan peristiwa lanjut ke semua pemantau; pemantau memuat ulang daftar.
5. **Hasil akhir:** jadwal berjalan kembali; timer peserta berjalan dilanjutkan dari sisa waktu yang sama seperti sebelum jeda.

### Sub-alur 3 — Usulan buka blokir (oleh Pengawas)

1. Pengawas mengajukan usulan buka blokir untuk satu pengerjaan yang diblokir, disertai alasan wajib.
2. Sistem menolak bila: pengerjaan bukan milik jadwal itu; pengerjaan tidak berstatus diblokir; sudah ada usulan menunggu untuk pengerjaan yang sama; alasan kosong.
3. Sistem menyimpan usulan berstatus menunggu dan menyiarkan peristiwa usulan ke pemantau.
4. Halaman pemantauan menampilkan usulan yang masih menunggu sebagai teks "nama siswa — alasan", diperbarui berkala tiap beberapa detik.
5. **Hasil akhir:** usulan menunggu tercatat; pengerjaan tetap diblokir sampai Proktor/Admin membuka blokir. Tidak ada tombol setuju/tolak usulan di tampilan — usulan hanya ditampilkan, dan penyelesaiannya terjadi lewat aksi buka blokir.

### Sub-alur 4 — Buka blokir peserta (unblock)

1. Pada baris peserta yang diblokir, Proktor/Admin mengisi alasan buka blokir (wajib) dan menekan tombol Buka Blokir.
2. Sistem menolak bila: pelaku bukan Proktor/Admin (Pengawas hanya boleh mengusulkan); pengerjaan bukan milik jadwal itu; pengerjaan tidak berstatus diblokir; alasan kosong.
3. Sistem menentukan sisa waktu lanjutan: bila pengerjaan menyimpan sisa waktu beku, pakai nilai itu; bila tidak, pakai selisih batas jam jadwal dengan waktu sekarang.
4. Bila sisa waktu lanjutan ≤ 0 → sistem menolak: ujian sudah berakhir, buka blokir tidak dapat dilakukan.
5. Sistem mengubah pengerjaan kembali menjadi berjalan dengan batas waktu baru = waktu sekarang + sisa waktu lanjutan, dan menghapus penanda beku. **Jawaban dan tanda ragu yang tersimpan TIDAK dihapus.**
6. Sistem menyelesaikan usulan menunggu terbaru untuk pengerjaan itu sebagai disetujui (dengan pencatat dan waktu pemrosesan); bila tidak ada usulan — Proktor membuka langsung — sistem membuat catatan usulan berstatus disetujui sebagai jejak.
7. Sistem menyiarkan pembaruan ke pemantau.
8. **Hasil akhir:** peserta dapat melanjutkan ujian dari soal terakhir dengan jawaban dan tanda ragu utuh, memakai sisa waktu yang sama seperti saat diblokir.

### Validasi / aturan bisnis yang terlihat di kode

- Pemantauan hanya untuk peran pemantau (Admin, Proktor, Guru, Pengawas); siswa tidak dapat membuka pemantauan.
- Tombol Jeda hanya aktif saat jadwal berjalan; tombol Lanjut hanya aktif saat jadwal jeda.
- Alasan buka blokir wajib diisi; alasan usulan wajib diisi.
- Siswa TIDAK menerima siaran apa pun: siswa mengetahui jeda/blokir/selesai hanya lewat penarikan status berkala atau lewat permintaan berikutnya (simpan jawaban, tandai ragu, muat detail).

### Yang TIDAK ditemukan di kode untuk UC07

- **Tidak ada aksi "paksa selesai" (memaksa satu/all peserta selesai dari pemantauan).** Yang ada di kode hanya: pantau, jeda/lanjutkan jadwal, usulkan buka blokir, buka blokir. Bila diagram menuntut "paksa selesai", itu belum diimplementasikan.

---

## UC08 — Ikuti Ujian

> [Catatan penggabung] Validasi token diuraikan penuh di UC09; UC08 memanggilnya. Untuk urutan pemeriksaan saat mulai ujian, lihat § Pertentangan antar-bagian, butir 1.

**Aktor:** Siswa (pelaku utama). Proktor/Admin menyiapkan jadwal dan token di luar alur ini.

### Alur utama

1. Siswa yang sudah masuk membuka beranda atau halaman ujian. Sistem menampilkan daftar jadwal ujian hari ini untuk kelas siswa, masing-masing dengan nama paket, jenis paket, tanggal, jam, status jadwal, dan status pengerjaan siswa bila ada.
2. Siswa memilih jadwal dan memasukkan token (kecuali paket jenis Tryout yang tanpa token), lalu menekan Mulai Ujian.
3. Sistem memvalidasi permintaan mulai (urutan lengkap pada tabel cabang di bawah). Bila lolos, sistem membuat pengerjaan baru berstatus berjalan dengan batas waktu = jam selesai jadwal, dan mengurutkan soal paket (dikocok bila paket diatur acak soal; tampilan jawaban dikocok bila paket diatur acak jawaban, dan hasil kocokan disimpan per pengerjaan agar stabil).
4. Bila ini peserta pertama yang memulai, sistem mengubah status jadwal dari draf menjadi aktif.
5. Siswa diarahkan ke halaman ujian. Sistem memuat soal-soal pengerjaan: nomor, tipe, pertanyaan, stimulus, media, tampilan jawaban, bobot, jawaban tersimpan, dan tanda ragu tersimpan. **Kunci jawaban dan pembahasan tidak pernah dikirim ke tampilan selama ujian.**
6. Siswa mengerjakan soal satu per satu, berpindah lewat navigasi nomor soal yang menunjukkan status tiap soal (aktif / dijawab / ragu / belum).
7. Setiap perubahan jawaban disimpan otomatis (lihat sub-alur autosave). Siswa dapat menandai soal sebagai ragu-ragu.
8. Timer hitung mundur tampil selama ujian, disinkronkan berkala dari server. Saat timer mencapai nol, sistem mengirim pengumpulan otomatis (sama dengan pengumpulan manual).
9. Siswa menekan Selesai Ujian → dialog konfirmasi menampilkan jumlah dijawab/belum dan peringatan bahwa soal belum dijawab dihitung 0 dan jawaban tidak dapat diubah setelah dikumpulkan → siswa mengonfirmasi → sistem memfinalisasi pengerjaan (hitung nilai, ubah status menjadi selesai, catat waktu pengumpulan).
10. Sistem menampilkan kartu hasil pengumpulan: nilai bila boleh tampil, atau pesan bahwa nilai disembunyikan sampai koreksi final, ditambah keterangan bila masih ada soal menunggu koreksi, serta tombol Lihat Hasil dan Kembali ke Beranda.

### Cabang / kondisi — mulai ujian

| Kondisi | Perilaku |
|---|---|
| Kelas siswa tidak ditemukan | Daftar ujian kosong + pesan agar menghubungi admin/proktor |
| Jadwal belum mulai / sudah berakhir | Mulai ditolak |
| Mulai lewat toleransi keterlambatan (pengaturan, bawaan 15 menit setelah jam mulai) — hanya untuk pengerjaan baru | Mulai ditolak |
| Token kosong / tidak cocok dengan token jadwal | Mulai ditolak (token tidak valid) |
| Token kedaluwarsa (pengaturan, bawaan 15 menit sejak token dibuat) | Mulai ditolak (token kedaluwarsa) |
| Paket tanpa soal | Mulai ditolak |
| Pengerjaan berjalan sudah ada | Sistem mengembalikan pengerjaan lama apa adanya (lanjut tanpa token baru) |
| Pengerjaan diblokir sudah ada | Mulai ditolak; pesan agar menghubungi proktor untuk buka blokir |
| Pengerjaan selesai sudah ada | Mulai ditolak; pesan bahwa ujian sudah diselesaikan |
| Jadwal berstatus jeda | Mulai ditolak |
| Jadwal berstatus selesai | Mulai ditolak |
| Pengerjaan milik siswa lain | Diperlakukan sebagai tidak ditemukan (identitas pemilik disamarkan) |
| Jenis paket Tryout | Token tidak diperlukan |

### Cabang / kondisi — selama ujian

| Kondisi | Perilaku |
|---|---|
| Jadwal dijeda saat ujian berjalan | Tampilan menunjukkan pemberitahuan jeda; timer membeku; penyimpanan jawaban ditolak; finalisasi otomatis ditunda |
| Pengerjaan diblokir (pelanggaran) | Tampilan layar diblokir; jawaban tersimpan; pesan agar menghubungi proktor/pengawas; ujian dilanjutkan dari sisa waktu setelah buka blokir |
| Batas waktu lewat saat menyimpan jawaban / menandai ragu / memeriksa status / memuat detail | Sistem memfinalisasi dulu pengerjaan itu, lalu permintaan ditolak atau status terbaru dikembalikan |
| Pengerjaan sudah selesai saat menyimpan jawaban | Penyimpanan ditolak |
| Soal bukan anggota paket jadwal ini | Penyimpanan ditolak |
| Pengumpulan pada pengerjaan yang sudah selesai | Idempotent: sistem mengembalikan hasil tersimpan tanpa menghitung ulang |
| Pengumpulan pada pengerjaan diblokir | Ditolak |
| Siswa menutup peramban tepat saat batas waktu lewat dan tidak mengirim permintaan lagi | Pengerjaan tetap berstatus berjalan di sistem sampai ada permintaan berikutnya yang menyentuhnya — tidak ada proses latar yang menyapu pengerjaan kedaluwarsa sendiri |

### Sub-alur 1 — Autosave jawaban

1. Siswa mengubah jawaban → tampilan langsung memperbarui status simpan menjadi "menyimpan".
2. Sistem menunggu 500 milidetik tanpa perubahan (debounce); hanya nilai terakhir yang dikirim.
3. Sistem menyimpan jawaban: memeriksa kepemilikan, status pengerjaan, batas waktu, status jeda, dan keanggotaan soal dalam paket.
4. Pada saat yang sama, sistem menghitung poin soal itu (lihat UC10 untuk aturan per tipe): soal manual mendapat poin kosong dan status belum dikoreksi; soal tidak dinilai mendapat poin 0 dan status sudah dinilai; soal lain mendapat poin dari kunci; tipe tak dikenal diperlakukan manual agar diperiksa guru, bukan diberi nilai palsu.
5. Sistem menyimpan/memperbarui jawaban dengan mempertahankan tanda ragu lama (penyimpanan jawaban tidak menyentuh tanda ragu).
6. Tampilan: sukses → status "tersimpan"; gagal → status "gagal — coba lagi". **Tidak ada antrean kirim ulang otomatis** — jawaban yang gagal tersimpan tidak dikirim ulang sendiri; siswa harus mengubah jawaban lagi untuk memicu simpan baru (nilai terakhir tetap ada di tampilan lokal).
7. **Hasil akhir:** jawaban tersimpan beserta poinnya (kecuali soal manual yang menunggu koreksi); monitor pemantau menerima pembaruan ringan.

### Sub-alur 2 — Tandai ragu-ragu

1. Siswa mencentang "Tandai ragu-ragu" pada soal → tampilan langsung diperbarui dan permintaan dikirim segera (tanpa debounce; kegagalan diabaikan diam-diam di tampilan).
2. Sistem menyimpan tanda ragu; **ragu dapat ditandai pada soal yang belum dijawab** (baris jawaban dibuat dengan jawaban kosong).
3. **Hasil akhir:** kotak nomor soal menampilkan status ragu (latar/border peringatan + penanda). Ragu saja tanpa jawaban dihitung sebagai BELUM dijawab, baik di tampilan maupun di penghitung server.

### Sub-alur 3 — Timer dan pengumpulan otomatis

1. Timer tampilan menghitung mundur tiap detik dari nilai yang disinkronkan dari server; tampil merah bila sisa ≤ 5 menit.
2. Tampilan menarik status pengerjaan dari server tiap 30 detik (sisa waktu, jumlah dijawab, jumlah ragu, status jadwal, status pengerjaan, nilai bila selesai dan boleh tampil).
3. Saat timer mencapai nol, tampilan memicu pengumpulan normal satu kali (dijaga agar tidak terkirim ganda).
4. Di sisi server, setiap permintaan terkait pengerjaan (muat detail, simpan jawaban, tandai ragu, status) lebih dulu memeriksa: bila pengerjaan masih berjalan, jadwal tidak dijeda, dan waktu sekarang sudah melewati batas waktu → pengerjaan difinalisasi secara malas (lazy) pada permintaan itu.
5. **Hasil akhir:** pengerjaan selesai dengan nilai terhitung; jawaban yang sudah tersimpan tidak hilang; pengumpulan ganda tidak membuat duplikat.

### Sub-alur 4 — Anti-curang pindah tab (bagian dari UC08; pemantauan ada di UC07)

1. Selama ujian berjalan dan belum dikumpulkan, tampilan mengawasi: halaman disembunyikan / pindah tab, jendela kehilangan fokus (blur), dan keluar dari mode layar penuh (bila layar penuh sempat aktif).
2. Setiap peristiwa dilaporkan ke server dengan deduplikasi tampilan 5 detik (laporan dalam 5 detik sejak laporan sebelumnya dibuang).
3. Server menerapkan deduplikasi 5 detik pula (pelanggaran terakhir < 5 detik lalu → diabaikan sebagai duplikat, tanpa catatan baru).
4. Selain itu, server mencatat satu pelanggaran, menambah hitungan, dan mencatat jejak audit.
5. Batas pelanggaran dari pengaturan (bawaan 2; bila pengaturan < 1 → 2):
   - Hitungan di bawah batas → peringatan: tampilan menampilkan spanduk peringatan ("Peringatan: … Pelanggaran berikutnya akan memblokir ujian Anda."), ujian berlanjut.
   - Hitungan mencapai batas → blokir otomatis: status pengerjaan menjadi diblokir dan sisa waktu dibekukan (sisa = batas waktu dikurangi waktu sekarang, tidak negatif) agar buka blokir dapat melanjutkan dari sisa yang sama.
6. Pelanggaran pada pengerjaan selesai → ditolak. Pelanggaran pada pengerjaan yang sudah diblokir → dijawab "sudah diblokir" tanpa catatan baru.
7. **Hasil akhir jalur peringatan:** ujian berlanjut dengan satu peringatan tercatat. **Hasil akhir jalur blokir:** ujian terkunci; siswa melihat layar diblokir dan harus meminta buka blokir; jawaban tersimpan utuh.
8. **Catatan:** peristiwa salin/tempel/potong pada kolom soal dicegah di tampilan saja, bukan dicatat sebagai pelanggaran.

### Validasi / aturan bisnis yang terlihat di kode

- Batas waktu pengerjaan = jam selesai jadwal, BUKAN waktu mulai ditambah durasi paket. Durasi paket hanya ditampilkan, tidak dipakai menghitung batas waktu.
- Token dibuat per jadwal, hanya berisi huruf/angka tanpa karakter yang mudah tertukar; hanya nilai hash-nya yang disimpan; nilai token asli hanya ditampilkan sekali saat dibuat dan tidak dapat dilihat lagi; token kedua langsung menghanguskan token pertama; jadwal selesai tidak dapat dibuatkan token.
- Saat memulai dari halaman ujian, sistem lebih dulu meminta mode layar penuh; bila peramban menolak, ujian tetap berjalan tanpa penalti. Titik masuk dari beranda tidak meminta layar penuh.
- Penilaian objektif terjadi per soal pada saat jawaban disimpan, bukan hanya saat pengumpulan. Nilai akhir saat pengumpulan adalah agregasi poin tersimpan ditambah soal tanpa jawaban (poin 0).
- Input token di halaman ujian diubah ke huruf besar dan dibatasi 6 karakter; titik masuk beranda tidak mengubah ke huruf besar (perbedaan kecil antar dua titik masuk; token yang dibuat hanya memakai huruf/angka huruf besar).

---

## UC09 — Validasi Token (generate token, salin, cek status / kedaluwarsa, token kedua menghanguskan pertama)

> [Catatan penggabung] Untuk urutan pemeriksaan token saat mulai ujian (UC09 vs UC08), lihat § Pertentangan antar-bagian, butir 1.

### Aktor

| Aktor | Hak |
|---|---|
| Admin, proktor | Generate token, salin token, cek status token |
| Guru, pengawas | Melihat daftar jadwal, tetapi TIDAK dapat generate / cek status token (ditolak) |
| Siswa | Memasukkan token saat memulai ujian (bukan mengelola token) |

### Konsep dasar

- Token = kode 6 karakter (huruf besar + angka; karakter yang mudah tertukar — 0, 1, I, O — tidak dipakai) yang membuka satu jadwal ujian.
- Satu jadwal hanya punya SATU token aktif pada satu waktu.
- Nilai token asli hanya ditampilkan SEKALI, pada saat generate. Sesudah itu sistem hanya menyimpan sidik hash-nya — nilai asli tidak dapat dilihat lagi oleh siapa pun, termasuk lewat cek status.
- Masa berlaku token dihitung sejak WAKTU GENERATE (bawaan 15 menit), bukan sejak jam mulai jadwal. Token yang di-generate jauh sebelum ujian dapat kedaluwarsa sebelum dipakai — generate ulang menjelang ujian.
- Token TIDAK tercantum di kartu peserta; token hanya diberikan proktor saat ujian.
- Jenis Tryout tidak memakai token sama sekali.

### Alur utama

#### 09a. Generate token
1. Admin/proktor membuka Dialog Token pada sebuah jadwal.
2. Dialog menampilkan status token saat ini (belum dibuat / aktif + sisa detik / kedaluwarsa) — tanpa nilai token.
3. Admin/proktor menekan "Generate Token".
4. Sistem memeriksa jadwal ada dan belum berstatus Selesai.
5. Sistem membuat token baru: nilai acak 6 karakter; yang disimpan hanya sidik hash + waktu kedaluwarsa (waktu generate + masa berlaku). Bila sudah ada token lama, token lama langsung hangus (tertimpa — hanya satu hash yang disimpan).
6. Jejak audit mencatat pembuatan token beserta waktu kedaluwarsa — TIDAK PERNAH mencatat nilai token asli.
7. Hasil akhir: dialog menampilkan token baru SEKALI (huruf besar, ditambah waktu kedaluwarsa) dengan pesan "token lama (bila ada) sudah hangus — salin sekarang". Token hanya disimpan di keadaan halaman (tidak disimpan permanen di perangkat / log).

#### 09b. Salin token
1. Admin/proktor menekan "Salin Token" pada token yang baru tampil.
2. Sistem menyalin nilai token ke papan klip; tombol berubah menjadi "Tersalin!" selama 2 detik.
3. Cabang: bila salin otomatis gagal → pesan "gagal menyalin — salin manual dari layar".
4. Hasil akhir: token tersalin (atau disalin manual oleh pengguna).

#### 09c. Cek status / kedaluwarsa token
1. Admin/proktor menekan "Status Token" (di dialog) atau melihat kolom Token di daftar jadwal.
2. Sistem mengembalikan: ada/tidak ada token; waktu kedaluwarsa (bila ada); sisa detik hingga kedaluwarsa (dibatasi minimal 0); sudah kedaluwarsa ya/tidak.
3. Nilai token TIDAK PERNAH dikembalikan dalam pemeriksaan status.
4. Hasil akhir: status tampil sebagai "Belum dibuat" / "Aktif, kedaluwarsa dalam N detik" / "Kedaluwarsa".

#### 09d. Validasi token saat peserta memulai ujian
Urutan pemeriksaan saat peserta memulai (relevan dengan token):

1. Jadwal harus ada; peserta harus dari kelas jadwal itu; jadwal belum Selesai; jadwal tidak sedang Jeda.
2. Bila peserta sudah punya pengerjaan berjalan → pengerjaan lama dikembalikan apa adanya (lanjut tanpa token baru).
3. Bila pengerjaan peserta diblokir / sudah selesai → ditolak (sebelum pemeriksaan token).
4. Bila sekarang belum jam mulai / sudah lewat jam selesai → ditolak.
5. Bila peserta terlambat melewati toleransi keterlambatan (bawaan 15 menit dari jam mulai, hanya untuk pengerjaan baru) → ditolak.
6. Pemeriksaan token (dilewati bila paket berjenis Tryout):
   - Token kosong ATAU jadwal belum punya token → ditolak: token tidak valid.
   - Sidik hash token masukan tidak cocok dengan yang tersimpan → ditolak: token tidak valid.
   - Sekarang melewati waktu kedaluwarsa token → ditolak: token kedaluwarsa.
7. Bila lolos semua (dan paket berisi soal) → pengerjaan dibuat; peserta masuk ujian.
8. Catatan masukan: token bersifat peka huruf besar. Satu titik masuk ujian mengubah masukan menjadi huruf besar otomatis; titik masuk lain tidak — token huruf kecil dari titik masuk kedua ditolak sebagai tidak valid.

### Cabang / hasil akhir UC09 (ringkas)

| Kondisi | Hasil |
|---|---|
| Jadwal Selesai + generate token | Ditolak: jadwal sudah selesai |
| Generate token baru saat token lama masih aktif | Token lama langsung hangus; hanya token baru yang berlaku |
| Status token dicek | Nilai token tidak pernah ditampilkan; hanya ada/tidak, sisa detik, kedaluwarsa |
| Token kosong / tidak cocok | Ditolak: token tidak valid |
| Token kedaluwarsa | Ditolak: token kedaluwarsa |
| Paket Tryout | Tanpa token |
| Pengerjaan berjalan sudah ada | Lanjut tanpa token |
| Pengerjaan diblokir / selesai | Ditolak sebelum pemeriksaan token |

### Hal yang tidak ditemukan di kode untuk UC09

- Tidak ada aksi cabut token tersendiri (satu-satunya cara menghanguskan token adalah generate token baru); tidak ada token per peserta (token berlaku untuk seluruh jadwal); tidak ada riwayat token (hanya token aktif terakhir yang tersimpan); tidak ada perpanjangan masa berlaku (perpanjangan = generate ulang); tidak ada pengiriman token otomatis ke peserta (token disampaikan proktor secara manual).

---

## UC10 — Koreksi Esai / Manual

**Aktor:** Guru, Admin (mengkoreksi). Siswa tidak mengoreksi; siswa hanya melihat status "menunggu koreksi" bila ada soal manual yang belum dikoreksi.

### Alur utama

1. Guru/Admin membuka daftar koreksi untuk satu jadwal.
2. Sistem menampilkan daftar jawaban yang masih menunggu koreksi: jawaban yang sudah diisi, belum dikoreksi, dan bertipe manual (uraian/esai dan variannya, unggah berkas, praktik singkat). Setiap baris memuat identitas siswa, status pengerjaan, pertanyaan, bobot, jawaban siswa, dan tanda ragu.
3. Guru/Admin memberi skor 0–100 untuk satu jawaban.
4. Sistem menghitung poin = skor/100 × bobot, menandai jawaban sudah dikoreksi, dan mencatat jejak audit.
5. Bila pengerjaan sudah selesai: sistem menghitung ulang nilai akhir pengerjaan, memperbarui nilai tersimpan, dan menyelaraskan nilai akhir ke rekap. Bila pengerjaan masih berjalan: nilai akhir dihitung saat pengumpulan.
6. Sistem menyiarkan pembaruan ke pemantau.
7. **Hasil akhir:** jawaban terkoreksi dengan poin final; nilai akhir pengerjaan diperbarui (bila sudah selesai); penanda "menunggu koreksi" berkurang/hilang bila semua soal manual sudah dikoreksi.

### Cabang / kondisi

| Kondisi | Perilaku |
|---|---|
| Peran bukan Guru/Admin | Akses ditolak |
| Jadwal tidak ditemukan | Daftar gagal dimuat |
| Skor di luar 0–100 | Ditolak |
| Jawaban tidak ditemukan | Ditolak |
| Jawaban sudah dikoreksi | Tidak muncul di daftar menunggu koreksi (daftar hanya menampilkan yang belum dikoreksi) |
| Jawaban kosong (belum dijawab) | Tidak muncul di daftar koreksi |
| Soal bukan tipe manual | Tidak muncul di daftar koreksi; dinilai otomatis saat jawaban disimpan |
| Pengerjaan masih berjalan saat dikoreksi | Nilai akhir tidak dihitung ulang saat itu; dihitung saat pengumpulan |
| Pengerjaan sudah selesai saat dikoreksi | Nilai akhir dihitung ulang seketika dan diselaraskan ke rekap |

### Validasi / aturan bisnis yang terlihat di kode

- Yang dikoreksi adalah jawaban per soal (satu baris koreksi = satu jawaban satu soal), bukan nilai keseluruhan siswa secara langsung.
- Poin koreksi selalu diturunkan dari skor: skor 0–100 dikalikan bobot soal dibagi 100.
- Soal manual yang dijawab tetapi belum dikoreksi membuat status "menunggu koreksi" pada nilai; soal manual yang belum dijawab tidak menunggu koreksi (tidak ada jawaban untuk dikoreksi).
- Nilai akhir baru hanya dihitung ulang seketika bila pengerjaan sudah selesai; untuk pengerjaan berjalan, koreksi tersimpan dan nilai akhir menyusul saat pengumpulan.
- Sinkronisasi nilai ke rekap bersifat best-effort: gagal menyelaraskan tidak menggagalkan koreksi. Status nilai yang sudah final di rekap tidak pernah diturunkan oleh sinkronisasi ini.

### Yang TIDAK ditemukan di kode untuk UC10

- **Tidak ada halaman tampilan khusus koreksi di aplikasi.** Layanan koreksi tersedia di sisi server (daftar menunggu koreksi + simpan skor), tetapi tidak ada halaman yang memanggilnya di kode tampilan yang dibaca. Koreksi saat ini hanya dapat dipakai lewat layanan tersebut, belum lewat antarmuka aplikasi.

---

## UC11 — Rekap Nilai & Berita Acara

**Aktor:** Staf yang berizin (Admin / Proktor / Guru sesuai izin rekap). Siswa hanya dapat melihat nilainya sendiri, dan hanya bila nilainya final dan pengaturan tampil hasil ke siswa aktif.

### Alur utama — Rekap nilai

1. Staf memilih jadwal pada halaman rekap.
2. Sistem menampilkan daftar nilai jadwal itu, per siswa: NISN, nama, kelas, jumlah dijawab, nilai, predikat, dan status (Final / Menunggu koreksi).
3. Sistem menampilkan statistik jadwal: jumlah peserta, rata-rata, tertinggi, terendah, jumlah lulus (≥ 70), dan jumlah nilai yang masih menunggu koreksi.
4. Staf dapat: mengekspor rekap sebagai berkas CSV, memfinalisasi nilai, mencatat remedial per siswa, membuka berita acara, atau memuat ulang data.

### Sub-alur 1 — Finalisasi nilai

1. Staf menekan Finalisasi Nilai.
2. Sistem mengubah SEMUA nilai pada jadwal itu menjadi status final dalam satu transaksi, sekaligus mengisi predikat tiap baris dari nilainya.
3. Sistem mencatat jejak audit berisi jumlah nilai yang difinalisasi.
4. **Hasil akhir:** semua nilai jadwal berstatus final dengan predikat terisi; jumlah nilai yang difinalisasi ditampilkan. Siswa baru dapat melihat nilainya sendiri setelah finalisasi ini DAN pengaturan tampil hasil aktif (lihat cabang di bawah).

### Sub-alur 2 — Remedial

1. Staf mengisi nilai remedial 0–100 pada baris seorang siswa dan menekan Simpan.
2. Sistem memvalidasi nilai 0–100 dan memastikan ada nilai akhir siswa itu pada jadwal tersebut; bila tidak ada → ditolak.
3. Sistem mencatat baris remedial (jadwal, siswa, nilai remedial), lalu memperbarui nilai akhir menjadi nilai tertinggi antara nilai awal dan nilai remedial, serta memperbarui predikat dari nilai tertinggi itu.
4. Sistem mencatat jejak audit berisi nilai awal, nilai remedial, dan nilai akhir.
5. **Hasil akhir:** nilai akhir siswa = nilai tertinggi; nilai awal tidak pernah ditimpa/dihilangkan (tersimpan lewat baris remedial + jejak audit). Mode remedial di pengaturan bersifat baca-saja "tertinggi".
6. **Batasan penting:** remedial di kode adalah entri nilai manual — BUKAN ujian ulang. Tidak ada pengerjaan baru, tidak ada soal, tidak ada token remedial.

### Sub-alur 3 — Ekspor rekap

1. Staf menekan Export CSV.
2. Sistem menulis berkas CSV berisi NISN, nama, kelas, nilai (2 desimal), predikat, dan status, dengan nama berkas berbasis jadwal dan tanggal.
3. Sistem mencatat jejak audit ekspor berisi jumlah baris.
4. **Hasil akhir:** berkas CSV terunduh.

### Sub-alur 4 — Berita acara

1. Staf membuka berita acara untuk satu jadwal (dari halaman rekap, atau memilih jadwal bila halaman dibuka tanpa jadwal).
2. Sistem menampilkan dokumen berita acara: mata pelajaran, kelas, tanggal, jam, jumlah peserta, hadir, tidak hadir, rata-rata nilai, jumlah lulus (≥ 70), catatan/kejadian, kolom tanda tangan Proktor / Pengawas / Kepala Sekolah, dan stempel waktu pembuatan dokumen.
3. Staf dapat mencetak lewat tombol Print (dokumen ramah cetak).
4. **Hasil akhir:** berita acara tampil dan siap cetak. Catatan selalu kosong dari sistem ("Tidak ada catatan." bila kosong) — tidak ada kolom isian catatan di tampilan.

### Cabang / kondisi

| Kondisi | Perilaku |
|---|---|
| Jadwal belum dipilih | Daftar rekap tidak dimuat; statistik tidak tampil |
| Tanpa izin melihat rekap | Akses ditolak |
| Tanpa izin ekspor / finalisasi / remedial | Aksi tersebut ditolak |
| Siswa membuka rekap | Hanya nilainya sendiri yang tampil, dan hanya bila nilainya final DAN pengaturan tampil hasil ke siswa aktif; bila pengaturan nonaktif → daftar kosong |
| Nilai remedial di luar 0–100 | Ditolak (validasi di tampilan dan di server) |
| Nilai akhir siswa pada jadwal itu tidak ditemukan | Remedial ditolak |
| Jadwal tanpa mapel | Berita acara memakai info minimal (jadwal tanpa mapel) |
| Kelas jadwal kosong | Berita acara menggabungkan nama-nama kelas peserta |

### Validasi / aturan bisnis yang terlihat di kode

- "Hadir" di berita acara = pengerjaan berstatus selesai ATAU diblokir. "Tidak hadir" = jumlah pengerjaan dikurangi hadir.
- "Dijawab" di rekap hanya menghitung jawaban yang benar-benar diisi; baris jawaban kosong yang dibuat saat siswa hanya menandai ragu TIDAK terhitung dijawab.
- Statistik: rata-rata dibulatkan 2 desimal; lulus = nilai ≥ 70; menunggu koreksi = jumlah baris berstatus menunggu koreksi.
- Finalisasi tidak menyaring apa pun: semua nilai jadwal itu menjadi final sekaligus.
- Tanggal/jam berita acara dikirim sebagai teks agar tidak gagal dirender di tampilan.

---

## UC12 — Analisis Butir Soal

**Aktor:** Staf yang berizin analisis. Proktor melihat statistik TANPA teks pertanyaan (pemisahan tugas: proktor tidak boleh melihat isi soal, hanya metadata/statistik). Guru/Admin melihat pertanyaan.

### Alur utama

1. Staf memilih jadwal pada halaman analisis butir.
2. Sistem menghitung analisis secara langsung (on-demand) dari data yang sudah ada — tanpa penyimpanan hasil hitung — memakai populasi: pengerjaan yang sudah selesai pada jadwal itu (berstatus selesai, atau memiliki nilai akhir final). Pengerjaan berjalan/diblokir TIDAK dihitung.
3. Sistem menampilkan ringkasan: jumlah peserta, jumlah soal, metode pengelompokan, penanda sampel kecil, dan catatan bila ada.
4. Sistem menampilkan per soal: urutan, pertanyaan (dikosongkan untuk Proktor), tipe, bobot, kompetensi, jumlah penjawab, jumlah benar, jumlah tidak dijawab, jumlah belum dikoreksi, tingkat kesukaran + label, daya beda + interpretasi, dan analisis pengecoh per opsi.
5. Staf dapat mengekspor analisis sebagai berkas CSV (ekspor Proktor juga tanpa teks pertanyaan) atau memuat ulang hitungan.

### Aturan hitung (mengikat, terlihat di kode)

- **Tingkat kesukaran P** = jumlah benar / jumlah penjawab. Label: P < 0,30 = Sukar; 0,30–0,70 = Sedang; P > 0,70 = Mudah. Soal tanpa penjawab yang ternilai: kesukaran kosong, label "Belum ada jawaban".
- **"Benar" untuk satu jawaban:** soal otomatis → poin penuh (poin = bobot) — untuk soal proporsional (pilihan ganda kompleks, menjodohkan, dsb.) hanya poin penuh yang dihitung benar. Soal manual → poin ≥ 0,6 × bobot. Jawaban tidak dijawab / belum dikoreksi → bukan benar.
- **Daya beda D** = (benar kelompok atas − benar kelompok bawah) / n, dengan n = ukuran tiap kelompok (kelompok sama besar). Kelompok atas/bawah = peserta bernilai tertinggi/terendah pada jadwal itu.
  - Peserta ≥ 10: n = pembulatan 27% × jumlah peserta (metode 27%).
  - Peserta < 10: n = setengah jumlah peserta (pembagian median 50%) + catatan "sampel kecil".
  - Peserta < 2: daya beda tidak dapat dihitung.
  - Interpretasi (dari nilai mentah sebelum pembulatan): D ≥ 0,40 = Sangat Baik; 0,30–0,39 = Baik; 0,20–0,29 = Cukup; 0,00–0,19 = Buruk; D < 0,00 = Sangat Buruk (soal bermasalah).
- **Pengecoh:** hanya untuk soal berbasis opsi (pilihan ganda dan pilihan ganda kompleks); soal lain = tidak berlaku. Per opsi: jumlah pemilih + persentase dari penjawab soal itu. Pengecoh "tidak berfungsi" bila dipilih < 5%.
- **Perlakuan soal manual belum dikoreksi:** tidak masuk penjawab maupun benar (dikeluarkan dari hitungan kesukaran/daya beda), tetapi dihitung pada kolom "belum dikoreksi".
- **Perlakuan tidak dijawab:** masuk hitungan "jumlah tidak dijawab" = jumlah peserta − jumlah yang menjawab; tidak masuk penjawab.
- **Pemeringkatan kelompok:** memakai nilai akhir bila ada; bila tidak ada, dihitung dari poin jawaban (jumlah poin / jumlah bobot paket × 100). Urutan seri diputus deterministik.
- **Penyembunyian untuk Proktor** berlaku juga pada ekspor CSV, sehingga Proktor tidak dapat mengelak dari penyembunyian lewat ekspor.

### Cabang / kondisi

| Kondisi | Perilaku |
|---|---|
| Belum masuk / tanpa izin analisis | Akses ditolak |
| Jadwal tidak ditemukan / ID tidak valid | Gagal memuat |
| Pemanggil adalah Proktor | Teks pertanyaan dikosongkan; statistik tetap tampil |
| Peserta < 10 | Metode median split 50% + penanda sampel kecil + catatan bahwa daya beda tidak dapat dipercaya sepenuhnya |
| Peserta < 2 | Daya beda tidak dapat dihitung + catatan |
| Soal tanpa penjawab ternilai | Kesukaran kosong; daya beda tetap dihitung dari kelompok (0 bila tidak ada benar di kedua kelompok) |
| Soal bukan berbasis opsi | Pengecoh = tidak berlaku |
| Jawaban untuk soal di luar paket | Diabaikan |

### Hasil akhir

- Tabel analisis per soal lengkap dengan kesukaran, daya beda, dan pengecoh, atau pesan gagal bila jadwal tidak valid / tidak berizin.
- Berkas CSV terekspor bila diminta (dengan penyembunyian Proktor bila berlaku).

---

## UC13 — Nilai per Kompetensi

**Aktor:** Staf yang berizin rekap (nilai per kompetensi memakai izin rekap, karena ia adalah laporan rekap, bukan analisis butir). Siswa hanya melihat nilainya sendiri.

### Alur utama

1. Staf memilih jadwal pada halaman nilai per kompetensi.
2. Sistem menghitung nilai per kompetensi secara langsung (on-demand) dari data yang sudah ada, memakai populasi yang sama dengan UC12: pengerjaan yang sudah selesai pada jadwal itu (berstatus selesai, atau memiliki nilai akhir final).
3. Sistem mengelompokkan soal menurut kompetensinya; soal tanpa kompetensi masuk kelompok "Tanpa Kompetensi" yang ditampilkan terakhir. Kelompok diurutkan menurut kode kompetensi.
4. Sistem menampilkan per siswa: NISN, nama, kelas, nilai per kompetensi (beserta jumlah soal, jumlah dijawab, jumlah belum dikoreksi, penanda), dan rata-rata siswa. Di bawah tabel, sistem menampilkan rata-rata kelas per kompetensi (rata-rata nilai siswa yang memiliki nilai pada kompetensi itu + jumlah siswa yang dihitung) dan rata-rata kelas keseluruhan.
5. Staf dapat mengekspor nilai per kompetensi sebagai berkas CSV atau memuat ulang hitungan.

### Aturan hitung (mengikat, terlihat di kode)

- **Nilai satu siswa untuk satu kompetensi** = jumlah poin soal ternilai / penyebut × 100, dibulatkan 2 desimal, dengan ketentuan:
  - Soal tidak dijawab → poin 0, bobot TETAP masuk penyebut (sama seperti nilai akhir).
  - Soal manual yang sudah dijawab tetapi belum dikoreksi → "belum dinilai": poin tidak masuk pembilang, bobot tidak masuk penyebut.
  - Siswa tidak menjawab satu pun soal kompetensi itu → nilai kosong (null) + penanda "tidak ada soal" (BUKAN 0).
  - Semua yang dijawab belum dikoreksi → nilai kosong + penanda "belum dikoreksi".
  - Sebagian belum dikoreksi → nilai dihitung dari yang sudah ternilai + penanda "sebagian belum dikoreksi".
- **Rata-rata siswa** dihitung dengan aturan yang sama atas semua soal (bukan rata-rata dari nilai per kompetensi).
- **Rata-rata kelas per kompetensi** = rata-rata nilai siswa yang memiliki nilai pada kompetensi itu (siswa bernilai kosong tidak dihitung); rata-rata kelas keseluruhan = rata-rata rata-rata siswa yang memiliki nilai.

### Cabang / kondisi

| Kondisi | Perilaku |
|---|---|
| Belum masuk / tanpa izin rekap | Akses ditolak |
| Jadwal tidak ditemukan / ID tidak valid | Gagal memuat |
| Pemanggil adalah Siswa | Hanya pengerjaannya sendiri yang dihitung/ditampilkan (pencegahan melihat milik orang lain) |
| Soal tanpa kompetensi | Masuk kelompok "Tanpa Kompetensi", urutan terakhir |
| Siswa tidak menjawab satu pun soal suatu kompetensi | Nilai kosong + penanda "tidak ada soal" |
| Semua jawaban siswa pada suatu kompetensi belum dikoreksi | Nilai kosong + penanda "belum dikoreksi" |

### Hasil akhir

- Tabel nilai per kompetensi per siswa beserta rata-rata kelas, atau pesan gagal bila jadwal tidak valid / tidak berizin.
- Berkas CSV terekspor bila diminta (siswa yang mengekspor hanya mendapat barisnya sendiri).

---

## UC14 — Kelola UKK

**Aktor:** Staf yang berizin UKK (Admin / Proktor / penguji). Siswa: hanya boleh mengunggah berkas praktik untuk dirinya sendiri dan melihat/mengunduh berkas serta sertifikat miliknya sendiri. Tidak ada ujian UKK daring di dalam alur ini — UKK dicatat sebagai kegiatan + nilai, bukan sebagai pengerjaan soal di sistem ujian.

### Alur utama — Daftar dan buat UKK

1. Staf membuka daftar UKK. Sistem menampilkan daftar dengan penyaring status, masing-masing memuat nama, mata pelajaran, kelas, jurusan, tanggal, penguji internal/eksternal, status, dan jumlah peserta dinilai.
2. Staf membuat UKK baru lewat dialog: nama (wajib), mata pelajaran, kelas, jurusan (ketiganya wajib dipilih), tanggal (format tanggal kalender), penguji internal, penguji eksternal, dan status awal (bawaan draf).
3. Sistem memvalidasi: nama tidak kosong; rujukan mata pelajaran/kelas/jurusan valid dan ada; status hanya draf/berjalan/selesai; tanggal berformat benar bila diisi.
4. **Hasil akhir:** UKK baru tercatat berstatus draf (atau status yang dipilih) dan muncul di daftar.

### Alur utama — Detail UKK

1. Staf membuka detail satu UKK. Sistem menampilkan: identitas UKK, penguji internal/eksternal, jumlah peserta dinilai dan jumlah lulus, tabel nilai peserta, dan daftar berkas praktik.
2. Staf dapat mengubah status UKK (draf / berjalan / selesai) lewat form status — perubahan status dilakukan dengan mengirim ulang seluruh data UKK beserta status baru.
3. Staf dapat masuk ke halaman Input Nilai atau menerbitkan sertifikat untuk siswa yang lulus.

### Sub-alur 1 — Input nilai UKK

1. Staf membuka halaman Input Nilai, memilih siswa (dari kelas UKK itu; bila siswa sudah punya nilai, form terisi dari nilai yang ada dan mode menjadi ubah).
2. Staf mengisi komponen nilai, masing-masing 0–100 dan boleh kosong: nilai teori, nilai proses, nilai produk, nilai sikap, plus catatan.
3. Tampilan menunjukkan pratinjau hitungan (tampilan saja): nilai praktik, nilai akhir, lulus, predikat. **Nilai resmi selalu dihitung ulang oleh sistem saat simpan; hasil dari server-lah yang dipakai.**
4. Sistem menyimpan nilai (buat baru bila belum ada untuk pasangan UKK–siswa; perbarui bila sudah ada — satu baris nilai per pasangan UKK–siswa) dan menghitung nilai turunan:
   - Nilai praktik = proses × 0,3 + produk × 0,5 + sikap × 0,2. Kosong bila ada satu komponen rubrik yang belum diisi.
   - Nilai akhir = teori × 0,4 + praktik × 0,6. Kosong bila teori/praktik belum lengkap.
   - Lulus = nilai akhir ≥ 70 (hanya bila nilai akhir ada). Predikat dari nilai akhir.
5. Sistem menampilkan nilai turunan dari server sebagai konfirmasi tersimpan.
6. **Hasil akhir:** nilai siswa tersimpan dengan nilai praktik/akhir/lulus/predikat terhitung; pencatat nilai terekam.

### Sub-alur 2 — Berkas praktik

1. Staf (atau siswa untuk dirinya sendiri) mengunggah berkas praktik untuk satu siswa: memilih siswa pemilik berkas dan berkasnya.
2. Sistem memvalidasi: siswa ada; berkas wajib ada; ukuran ≤ 10 MB; ekstensi hanya pdf/doc/docx/jpg/jpeg/png/zip. Siswa hanya boleh mengunggah untuk dirinya sendiri (bila siswa tidak mengisi pemilik, pemilik diisi dirinya sendiri; bila mengisi orang lain → ditolak).
3. Sistem membersihkan nama berkas, memberi awalan unik berbasis waktu agar tidak bertabrakan dan tidak dapat ditebak, menyimpan berkas di penyimpanan per UKK per siswa, dan mencatatnya.
4. **Hasil akhir:** berkas tercatat dan tampil di daftar berkas UKK.
5. Daftar berkas: staf melihat semua berkas UKK; siswa hanya melihat berkas miliknya sendiri.
6. Unduh berkas: staf mengunduh berkas mana pun; siswa hanya boleh mengunduh miliknya sendiri (upaya mengunduh milik orang lain dijawab "tidak ditemukan"). Setiap unduhan dicatat di jejak audit. Berkas dilayani lewat pemeriksaan izin — bukan lewat penyajian berkas statis.

### Sub-alur 3 — Sertifikat

1. Dari detail UKK, staf menekan tombol Sertifikat pada baris siswa yang lulus (tombol hanya tampil bila siswa lulus).
2. Sistem memeriksa: siswa ada; bila sertifikat sudah ada → kembalikan yang sudah ada apa adanya (idempotent); bila siswa tidak lulus (atau belum ada nilai) → ditolak.
3. Sistem menerbitkan nomor sertifikat format `UKK/<tahun>/<kode-jurusan>/<nomor-urut 4 digit>`; tahun dari tanggal UKK bila ada, bila tidak tahun berjalan; nomor urut per jurusan + tahun; bila terjadi bentrok nomor karena dua penerbitan bersamaan, sistem mencoba lagi hingga beberapa kali.
4. Sistem mencatat jejak audit penerbitan.
5. **Hasil akhir:** sertifikat tercatat dan tampilan sertifikat menunjukkan nomor, nama siswa, NISN, kompetensi keahlian, status LULUS, predikat, nilai teori/praktik/akhir, penguji internal/eksternal, tanggal terbit dan tanggal UKK, serta tombol Print. Dokumen ditegaskan sebagai data/draf dari sistem sekolah — bukan penerbitan resmi lembaga berwenang.
6. Melihat sertifikat: staf melihat sertifikat siapa pun; siswa hanya sertifikat miliknya sendiri (milik orang lain dijawab "tidak ditemukan").

### Sub-alur 4 — Hapus UKK

1. Staf meminta hapus satu UKK dengan dialog konfirmasi.
2. Sistem menolak bila UKK sudah memiliki nilai/berkas/sertifikat.
3. **Hasil akhir:** UKK terhapus bila belum memiliki turunan apa pun; bila tidak, UKK tetap ada dengan pesan penolakan.

### Cabang / kondisi

| Kondisi | Perilaku |
|---|---|
| Tanpa izin UKK yang sesuai (lihat / ubah / buat) | Aksi ditolak |
| UKK tidak ditemukan / ID tidak valid | Gagal memuat |
| Nama UKK kosong / rujukan tidak ada / status di luar draf-berjalan-selesai / tanggal tidak berformat benar | Buat/ubah ditolak |
| Nilai UKK diubah saat UKK berstatus selesai | Ditolak, kecuali oleh Admin — dan perubahan Admin dicatat di jejak audit sebagai override |
| Komponen nilai di luar 0–100 | Ditolak |
| Siswa mengubah nilai UKK | Ditolak (nilai hanya oleh staf berizin) |
| Berkas > 10 MB / ekstensi tidak diizinkan / tanpa berkas | Unggah ditolak |
| Siswa mengunggah/melihat/mengunduh untuk orang lain | Ditolak / dijawab tidak ditemukan |
| Sertifikat untuk siswa tidak lulus / belum dinilai | Ditolak |
| Sertifikat sudah ada | Dikembalikan apa adanya, tanpa duplikat |
| Hapus UKK yang sudah punya nilai/berkas/sertifikat | Ditolak |

### Validasi / aturan bisnis yang terlihat di kode

- Satu format rubrik untuk semua jurusan: Proses 30% / Produk 50% / Sikap 20%; bobot akhir Teori 40% / Praktik 60%. Nilai selalu dihitung sistem, bukan dari tampilan.
- Nilai teori BOLEH diisi manual; penarikan otomatis nilai teori dari pengerjaan ujian teori TIDAK dipaksakan — tidak ada penyalinan otomatis nilai ujian ke nilai UKK di kode.
- Nilai 0–100 untuk semua komponen; nilai kosong berarti belum diisi (bukan 0).
- Sertifikat HANYA bagi yang lulus (nilai akhir ≥ 70). Sertifikat di sistem adalah data + draf, bukan penerbitan hukum.
- Hapus UKK bersifat menjaga turunan: UKK yang sudah punya nilai/berkas/sertifikat tidak dapat dihapus.

---

## UC15 — Kelola Kegiatan, Ruangan & Kartu Peserta (event, ruangan, penugasan ruangan per jadwal, kartu peserta 1-kartu)

### Aktor

| Aktor | Hak |
|---|---|
| Admin | Kelola kegiatan + ruangan + penugasan + kartu (semua) |
| Guru | Kartu peserta (lihat + unduh). TIDAK dapat kelola kegiatan/ruangan (ditolak) |
| Proktor | Kartu peserta (lihat + unduh); penugasan kegiatan/ruangan ke jadwal |
| Pengawas, penguji UKK | Tidak punya akses |
| Siswa | Hanya kartu miliknya sendiri (lihat pratinjau + unduh); kartu orang lain tidak ditemukan |

### Konsep dasar

- Kegiatan (event) = pengelompokan ujian (mis. STS Ganjil, SAS Genap, USBK, UKK). Satu jadwal dapat merujuk ke satu kegiatan (opsional).
- Ruangan = ruang fisik ujian. Satu jadwal ditugaskan ke satu ruangan.
- Kartu peserta menganut model 1-KARTU: SATU kartu per siswa berlaku untuk SEMUA ujian — identitas siswa ditulis sekali + tabel SEMUA jadwal kelasnya. Nomor peserta unik PER JADWAL (siswa yang sama di jadwal berbeda boleh bernomor berbeda).
- Kartu TIDAK berisi token ujian.

### Alur utama

#### 15a. Kelola kegiatan (event)
1. Pengguna membuka halaman Kegiatan.
2. Sistem menampilkan daftar kegiatan: kode, nama, jenis, tanggal mulai, tanggal selesai, deskripsi; dapat dicari (kode/nama) dan disaring per jenis.
3. **Tambah:** pengguna menekan "Tambah Event", mengisi kode, nama, jenis (STS/SAS/USBK/UKK/Tryout/Lainnya — bawaan Lainnya), tanggal mulai/selesai (opsional), deskripsi (opsional), lalu menyimpan.
4. **Ubah:** pengguna menekan "Ubah", mengubah, menyimpan.
5. **Hapus:** pengguna menekan "Hapus" dan mengonfirmasi.
6. Pemeriksaan: kode dan nama wajib diisi; jenis harus salah satu dari 6 jenis (kosong dianggap Lainnya); tanggal harus format tanggal yang benar; tanggal selesai tidak boleh sebelum tanggal mulai; kode belum dipakai kegiatan lain (saat ubah: belum dipakai kegiatan lain selain dirinya).
7. Cabang hapus: bila kegiatan masih dipakai jadwal ujian mana pun → ditolak ("kegiatan masih dipakai N jadwal — lepas kegiatan dari jadwal dulu").
8. Hasil akhir: kegiatan tersimpan / diperbarui / terhapus.

#### 15b. Kelola ruangan
1. Pengguna membuka halaman Ruangan.
2. Sistem menampilkan daftar ruangan: kode, nama, kapasitas kursi, lokasi/gedung, keterangan; dapat dicari (kode/nama/lokasi).
3. **Tambah:** pengguna menekan "Tambah Ruangan", mengisi kode, nama, kapasitas kursi, lokasi/gedung (opsional), keterangan (opsional), lalu menyimpan.
4. **Ubah:** pengguna menekan "Ubah", mengubah, menyimpan.
5. **Hapus:** pengguna menekan "Hapus" dan mengonfirmasi.
6. Pemeriksaan: kode dan nama wajib diisi; kapasitas harus bilangan bulat > 0 (tampilan juga memeriksa bilangan bulat); kode belum dipakai ruangan lain (saat ubah: belum dipakai ruangan lain selain dirinya).
7. Cabang ubah — kapasitas baru lebih kecil: bila kapasitas baru kurang dari jumlah siswa aktif kelas dari jadwal yang sudah memakai ruangan ini → ditolak, dengan pesan yang menyebut jadwalnya ("turunkan kapasitas setelah jadwal dipindah").
8. Cabang hapus: bila ruangan masih dipakai jadwal ujian mana pun → ditolak ("ruangan masih dipakai N jadwal — lepas ruangan dari jadwal dulu").
9. Hasil akhir: ruangan tersimpan / diperbarui / terhapus.

#### 15c. Penugasan kegiatan + ruangan per jadwal
1. Staf membuka seksi "Penugasan Event & Ruangan per Jadwal" di halaman Kartu Peserta.
2. Staf memilih jadwal dari daftar (label: tanggal + jam + paket + kelas + status).
3. Sistem membaca penugasan tersimpan jadwal itu dan mengisi awal pilihan kegiatan + ruangan.
4. Staf memilih kegiatan dan/atau ruangan (masing-masing dapat dipilih "Tanpa …" untuk melepas penugasan), lalu menekan "Simpan Penugasan".
5. Sistem memeriksa: jadwal ada; kegiatan yang dipilih harus ada (bila diisi); ruangan yang dipilih harus ada (bila diisi); kapasitas ruangan ≥ jumlah siswa AKTIF kelas jadwal — bila kurang → ditolak dengan pesan yang menyebut nama ruangan, kapasitasnya, kelas, dan jumlah siswanya.
6. Hasil akhir: penugasan tersimpan (kegiatan dan ruangan baru, atau dilepas bila dikosongkan). Penugasan ini dipakai untuk nomor peserta dan ruangan yang tercantum di kartu.

#### 15d. Kartu peserta — unduh staf (satu siswa / gabungan)
1. Staf membuka halaman Kartu Peserta.
2. Sistem menampilkan daftar siswa (nama, NISN, kelas, jurusan) dengan saringan: kegiatan (untuk PDF gabungan), kelas, pencarian nama/NISN, dan penomoran halaman (20 siswa per halaman).
3. **Satu siswa:** staf menekan "Unduh PDF" pada baris siswa → berkas kartu satu siswa diunduh (satu kartu, identitas sekali + tabel semua jadwal kelasnya).
4. **Gabungan:** staf mengatur saringan kegiatan/kelas lalu menekan "Unduh Semua (PDF)" → berkas gabungan diunduh (satu halaman per siswa).
5. Pemeriksaan gabungan: saringan kelas/kegiatan yang diisi harus ada; bila tidak ada kartu untuk saringan itu → tidak ditemukan; bila jumlah siswa melebihi batas 500 → ditolak ("gunakan saringan kelas / kegiatan").
6. Hasil akhir: berkas PDF terunduh.

#### 15e. Kartu peserta — siswa melihat kartunya sendiri
1. Siswa membuka "Kartu Saya".
2. Sistem menampilkan pratinjau: identitas dari sesi masuk (nama, nama pengguna) + tabel jadwal ujian HARI INI (pratinjau saja — bukan isi kartu lengkap).
3. Siswa menekan "Unduh Kartu (PDF)" → kartu lengkap miliknya diunduh (identitas sekali + tabel SEMUA jadwal kelasnya).
4. Cabang: siswa meminta kartu siswa lain → tidak ditemukan (disamarkan, bukan "dilarang"). Siswa tidak aktif → tidak ditemukan.
5. Hasil akhir: kartu milik sendiri terunduh.

### Isi kartu peserta (1 kartu)

- Kop: nama sekolah (bila diatur), judul kartu, nama kegiatan gabungan (atau nama baku bila tidak ada kegiatan).
- Identitas (sekali): nama siswa, NISN, kelas, jurusan, nama pengguna. Nomor peserta TIDAK ada di blok identitas.
- Tabel jadwal ujian (semua jadwal kelas siswa — tanpa saringan status, termasuk yang masih Draf — urut tanggal + jam): nomor urut, nomor peserta per jadwal, mata pelajaran, paket, tanggal, jam, ruangan (nama saja).
- Nomor peserta per jadwal: `<kode kegiatan atau US>/<tahun tanggal jadwal>/<nomor urut 3 digit>` — contoh: `STS-GANJIL-26/2026/001`; tanpa kegiatan: `US/2026/001`. Nomor urut dihitung dari urutan nama siswa A–Z di antara siswa aktif kelas itu.
- Kotak Login Ujian: nama pengguna + kata sandi awal (= NISN, wajib diganti saat login pertama).
- Tanda tangan: dua kolom sejajar — Peserta Ujian (nama terang = nama siswa) dan Panitia.
- Bila kelas belum punya jadwal: tabel menampilkan "Belum ada jadwal ujian untuk kelas ini."
- Catatan cetak: tabel panjang berpindah halaman dengan kepala tabel diulang; identitas tidak diulang. Setiap halaman berkas gabungan = satu siswa. Kaki halaman berisi catatan + nomor halaman.

### Cabang / hasil akhir UC15 (ringkas)

| Kondisi | Hasil |
|---|---|
| Kode kegiatan / ruangan sudah dipakai | Ditolak: kode sudah dipakai |
| Tanggal selesai kegiatan sebelum tanggal mulai | Ditolak (tampilan + sistem) |
| Kapasitas ruangan ≤ 0 / bukan bilangan bulat | Ditolak |
| Kegiatan / ruangan dipakai jadwal + hapus | Ditolak: lepas dari jadwal dulu |
| Kapasitas baru ruangan < siswa kelas jadwal pemakai | Ditolak, menyebut jadwalnya |
| Penugasan: kegiatan / ruangan tidak ditemukan | Ditolak |
| Penugasan: kapasitas ruangan < siswa aktif kelas jadwal | Ditolak, menyebut angka-angkanya |
| Penugasan dikosongkan | Penugasan dilepas |
| Kartu: siswa lain (oleh siswa) / siswa tidak aktif | Tidak ditemukan |
| Kartu gabungan: > 500 siswa | Ditolak: persempit saringan |
| Kartu gabungan: tidak ada kartu untuk saringan | Tidak ditemukan |

### Hal yang tidak ditemukan di kode untuk UC15

- Tidak ada pemeriksaan bentrok (ruangan yang sama dapat ditugaskan ke jadwal yang tumpang tindih; kegiatan tidak diperiksa kesesuaian jenisnya dengan jenis paket); tidak ada penugasan ruangan per siswa (ruangan selalu per jadwal); tidak ada kustomisasi desain kartu, foto, kode QR / barcode; tidak ada cetak lewat peramban (unduhan hanya PDF dari sistem); tidak ada kartu untuk siswa tidak aktif; tidak ada pengingat / pemberitahuan kartu.

---

## UC16 — Cadangkan & Pulihkan (Backup & Restore)

**Aktor:** Staf yang berizin backup. Membuat dan melihat daftar backup memerlukan izin buat backup; memulihkan dan menghapus memerlukan izin pulihkan backup.

### Alur utama — Daftar backup

1. Staf membuka halaman Backup & Restore. Sistem menampilkan daftar backup, terbaru dulu, dengan pagination (bawaan 50 per halaman, maks 200): nama berkas, jenis (Manual / Pre-restore), ukuran, waktu dibuat, pembuat, dan keterangan.
2. Staf dapat berpindah halaman daftar (Sebelumnya / Berikutnya).

### Sub-alur 1 — Buat backup manual

1. Staf mengisi keterangan opsional (maks 500 karakter) dan menekan Buat Backup.
2. Sistem membuat salinan penuh basis data ke berkas dengan nama unik berbasis waktu + akhiran acak (tidak pernah menimpa berkas lama; nama berkas unik), lalu mencatat metadata backup (nama berkas, ukuran, jenis manual, pembuat, keterangan, waktu).
3. Sistem mencatat jejak audit pembuatan backup.
4. **Hasil akhir:** backup baru tercatat di daftar; pesan menampilkan nama berkas dan ukurannya. Bila perkakas salinan basis data tidak tersedia di server → gagal dengan pesan jelas.

### Sub-alur 2 — Unduh backup

1. Staf menekan Unduh pada satu backup.
2. Sistem memeriksa izin (izin buat ATAU izin pulihkan), memeriksa metadata, dan memastikan nama berkas aman (menolak nama yang mengandung pemisah jalur / traversal).
3. Sistem mengalirkan berkas sebagai unduhan lampiran dan mencatat jejak audit unduhan.
4. **Hasil akhir:** berkas terunduh. Bila metadata ada tetapi berkas hilang di server → ditolak dengan pesan "metadata ada, file hilang". Direktori penyimpanan backup TIDAK disajikan sebagai berkas statis — unduhan hanya lewat pemeriksaan izin ini.

### Sub-alur 3 — Pulihkan (restore)

1. Staf menekan Restore pada satu backup → dialog konfirmasi dengan peringatan keras: restore menimpa SEMUA data saat ini; backup pengaman dibuat otomatis sebelum restore dimulai; semua perubahan setelah backup itu dibuat akan hilang; restore ditolak bila ada jadwal aktif/jeda atau peserta sedang mengerjakan ujian. Staf mengonfirmasi.
2. Sistem memeriksa izin pulihkan, memeriksa metadata dan keberadaan berkas, serta memastikan perkakas pulih tersedia.
3. **Gerbang ujian:** sistem memeriksa apakah ada jadwal berstatus aktif/jeda ATAU pengerjaan berstatus berjalan. Bila ada → restore DITOLAK, penolakan dicatat di jejak audit, dan pesan menyebut jumlah penghalang.
4. **Backup pengaman otomatis:** sistem membuat backup jenis pre-restore dan mencatat metadata + jejak auditnya SEBELUM restore dimulai. Bila backup pengaman gagal dibuat → restore dibatalkan demi keamanan data.
5. Sistem memulihkan isi berkas dump ke basis data yang sama (membersihkan dan membuat ulang semua isi dari dump).
6. Karena pemulihan menghapus dan membuat ulang semua isi — termasuk metadata backup pengaman dan jejak audit yang ditulis sebelum restore — sistem menulis kembali (mengembalikan) metadata backup pengaman + metadata backup target + jejak audit pembuatan pengaman dan jejak audit restore ke basis data hasil restore, sehingga jejak "restore dari berkas X, pengaman berkas Y" tetap ada.
7. **Hasil akhir jalur berhasil:** data kembali ke isi backup target; pesan menampilkan berkas sumber restore dan berkas backup pengaman.
8. **Hasil akhir jalur gagal:** pesan menyebut nama berkas pengaman agar admin dapat memulihkan manual; backup pengaman tetap tercatat (restore belum berhasil menimpa data secara utuh). Bila melewati batas waktu (10 menit) → pesan menyatakan data saat ini mungkin sebagian terpulihkan dan menunjuk backup pengaman.
9. Batas waktu salin/pulih: 10 menit per operasi.

### Sub-alur 4 — Hapus backup

1. Staf menekan Hapus pada satu backup → dialog konfirmasi (peringatan: berkas dan metadata dihapus permanen; backup pre-restore berusia < 7 hari tidak boleh dihapus).
2. Sistem memeriksa izin pulihkan.
3. Bila backup berjenis pre-restore dan berusia < 7 hari → hapus DITOLAK, penolakan dicatat di jejak audit beserta alasannya.
4. Selain itu: sistem menghapus berkas, menghapus metadata, dan mencatat jejak audit penghapusan.
5. **Hasil akhir:** backup hilang dari daftar; atau tetap ada dengan pesan penolakan bila pre-restore muda.

### Cabang / kondisi

| Kondisi | Perilaku |
|---|---|
| Tanpa izin buat backup | Membuat/melihat daftar ditolak (melihat daftar: perlu izin buat ATAU pulihkan) |
| Tanpa izin pulihkan | Restore dan hapus ditolak |
| Keterangan > 500 karakter | Buat backup ditolak |
| ID backup tidak valid / tidak ditemukan | Ditolak |
| Metadata ada, berkas hilang | Unduh/restore ditolak dengan pesan yang membedakannya |
| Perkakas salin/pulih tidak tersedia di server | Operasi gagal dengan pesan jelas |
| Ada jadwal aktif/jeda atau pengerjaan berjalan | Restore ditolak |
| Backup pengaman gagal dibuat | Restore dibatalkan |
| Pulih gagal / batas waktu terlampaui | Pesan menunjuk berkas pengaman untuk pemulihan manual |
| Backup pre-restore berusia < 7 hari | Hapus ditolak + penolakan dicatat |
| Operasi salin/pulih/hapus bersamaan | Operasi diserialkan: pulih tidak boleh berjalan bersamaan dengan salin atau pulih lain; penghapusan berkas tidak boleh balapan dengan pulih yang memakai berkas itu |

### Validasi / aturan bisnis yang terlihat di kode

- Hanya ada dua jenis backup: manual dan pre-restore. Tidak ada jenis lain.
- Nama berkas selalu dibuat sistem (berbasis waktu + akhiran acak); pengguna TIDAK pernah menentukan nama berkas — satu-satunya masukan pengguna adalah keterangan bebas dan ID backup yang dipilih. Nama berkas yang dipakai operasi selalu diambil dari metadata di basis data berdasarkan ID, bukan dari parameter/body.
- Kata sandi basis data TIDAK PERNAH ditulis ke log/respons — hanya diteruskan sebagai variabel lingkungan ke proses anak salin/pulih.
- **Tidak ada backup otomatis harian di dalam aplikasi.** Backup di halaman ini hanya manual. Backup otomatis harian dijadwalkan lewat skrip cron di luar aplikasi (disebut di tampilan).
- Salinan gagal → berkas parsial dihapus, tidak disisakan. Gagal mencatat metadata setelah salin berhasil → berkas dihapus.
- Penghapusan backup pre-restore muda (< 7 hari) ditolak agar pengaman restore terakhir tetap tersedia.

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

> [Catatan penggabung] Daftar usulan dimuat pemantau tiap 10 detik menurut bagian A; bagian C menyebut "beberapa detik" — lihat § Pertentangan antar-bagian, butir 4.

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

## UC19 — Lihat Hasil & Pembahasan

**Aktor:** Siswa (melihat hasil pengerjaannya sendiri). Guru/Admin tidak memiliki halaman hasil siswa ini; mereka melihat nilai lewat rekap/analisis/koreksi.

### Alur utama

1. Siswa membuka halaman hasil untuk satu pengerjaannya (dari kartu hasil pengumpulan atau dari beranda — tautan hasil tampil pada jadwal yang sudah selesai).
2. Sistem memeriksa: pengerjaan milik siswa sendiri dan sudah selesai; bila belum selesai → hasil belum tersedia.
3. Sistem menampilkan kartu Nilai Akhir: nilai bila boleh tampil, atau pesan "Nilai disembunyikan sampai koreksi final." Bila masih ada soal menunggu koreksi → pesan tambahan "Masih ada soal uraian yang menunggu koreksi guru." Untuk Tryout → keterangan "Tryout — pembahasan langsung tampil."
4. Sistem menampilkan per soal, diurutkan sesuai urutan pengerjaan: nomor soal, pertanyaan, jawaban siswa, penanda ragu-ragu bila ada, dan status soal.
5. Bila hasil boleh tampil (lihat aturan tampil di bawah), per soal juga tampil: poin (poin / bobot), ringkasan kunci benar, dan pembahasan guru. Bila tidak boleh tampil: hanya pertanyaan + jawaban siswa (tanpa benar/salah/poin/kunci/pembahasan).
6. Soal paket yang tidak dijawab sama sekali tetap ditampilkan: status salah bila hasil boleh tampil, atau disembunyikan bila tidak boleh tampil.

### Aturan tampil hasil (mengikat, terlihat di kode)

- Hasil boleh tampil bila: saklar tampil hasil pada paket AKTIF **dan** pengaturan tampil hasil ke siswa AKTIF. Bawaan pengaturan tampil hasil ke siswa adalah nonaktif — jadi bawaan sistem menyembunyikan nilai kecuali kedua saklar diaktifkan.
- **Pengecualian Tryout:** paket jenis Tryout SELALU menampilkan hasil lengkap (kunci + pembahasan), tanpa peduli kedua saklar di atas.
- Nilai akhir hanya diisi bila hasil boleh tampil dan nilainya tersimpan; bila tidak, nilai kosong.
- Kunci yang ditampilkan adalah RINGKASAN kunci benar per tipe (mis. "Jawaban benar: B — …", "Pasangan benar: …", "Jawaban yang diterima: …", "Urutan benar: …", "Dinilai manual oleh guru (rubrik).", "Tidak dinilai (survei)."), bukan kunci mentah. Kunci dan pembahasan tidak pernah dikirim selama ujian — hanya setelah selesai dan bila boleh tampil.

### Status per soal (mengikat, terlihat di kode)

| Status | Arti |
|---|---|
| Benar | Poin mencapai bobot penuh |
| Sebagian benar | Poin di antara 0 dan bobot penuh |
| Salah | Poin 0 (termasuk soal yang tidak dijawab, bila hasil boleh tampil) |
| Menunggu koreksi | Poin kosong (soal manual belum dikoreksi) |
| Disembunyikan | Hasil tidak boleh tampil — tanpa benar/salah/poin |

### Penyajian jawaban siswa

- Jawaban teks/angka tampil literal apa adanya (tidak ditafsirkan sebagai markup).
- Jawaban pilihan ganda tampil sebagai label yang dipilih; pilihan ganda kompleks sebagai daftar label.
- Jawaban menjodohkan dirender per pasangan, satu baris "kiri → kanan" per pasangan; pasangan yang belum dipilih ditandai "belum dipilih".
- Jawaban benar-salah / setuju-tidak-setuju tampil sebagai daftar nilai per pernyataan.
- Jawaban kosong tampil sebagai "Tidak dijawab".

### Cabang / kondisi

| Kondisi | Perilaku |
|---|---|
| Pengerjaan bukan milik siswa | Diperlakukan sebagai tidak ditemukan |
| Pengerjaan belum selesai | Hasil belum tersedia (pesan: hasil belum tersedia — ujian mungkin belum selesai) |
| ID pengerjaan tidak ada di alamat halaman | Pesan ID tidak ditemukan |
| Hasil tidak boleh tampil | Nilai kosong + pesan disembunyikan; per soal hanya pertanyaan + jawaban siswa |
| Hasil boleh tampil / Tryout | Nilai + poin + ringkasan kunci + pembahasan tampil |
| Masih ada soal manual belum dikoreksi | Penanda menunggu koreksi di kartu nilai dan status menunggu koreksi pada soal tersebut |

### Hasil akhir

- Halaman hasil lengkap per soal dengan nilai (bila boleh tampil) dan pembahasan (bila boleh tampil / Tryout), atau halaman hasil terbatas (pertanyaan + jawaban siswa saja) bila hasil disembunyikan.

---

---

## Pertentangan antar-bagian (catatan penggabung — BELUM diputuskan, jangan pilih diam-diam)

### 1. Urutan pemeriksaan saat siswa mulai ujian: UC09 (Bagian B) vs UC08 (Bagian C) — PERTENTANGAN NYATA

- **Bagian B, UC09 § 09d** menyajikan urutan pemeriksaan secara eksplisit dan
  berlabel urutan: pengerjaan berjalan dikembalikan → pengerjaan diblokir /
  selesai ditolak **sebelum** pemeriksaan token → toleransi telat → pemeriksaan token.
- **Bagian C, UC08, tabel cabang "mulai ujian"** menyajikan urutan yang berbeda:
  pemeriksaan token (kosong / tidak cocok / kedaluwarsa) tercantum **sebelum**
  pengerjaan berjalan / diblokir / selesai, dan status jadwal jeda/selesai tercantum
  sesudahnya.
- **Hasil akhir tiap kondisi sama** di kedua bagian (ditolak dengan pesan yang sama);
  yang berbeda hanya **urutan** pemeriksaannya — berpengaruh pada pesan mana yang
  dilihat siswa bila dua kondisi gagal sekaligus (mis. pengerjaan diblokir DAN
  token salah).
- **Sikap penggabung:** rancangan diagram memakai urutan Bagian B sebagai hipotesis
  kerja (B melabelinya eksplisit "urutan pemeriksaan", dan menolak siswa yang
  diblokir sebelum meminta token lebih masuk akal), TETAPI ini bukan keputusan —
  penulis diagram wajib mengonfirmasi ke induk sebelum menggambar UC08/UC09.

### 2. Rujukan aturan poin otomatis: UC08 (Bagian C) merujuk ke UC10, seharusnya ke UC03 — RUJUKAN TIDAK TEPAT

- **Bagian C, UC08, sub-alur Autosave langkah 4** menulis "(lihat UC10 untuk
  aturan per tipe)".
- **Faktanya:** aturan poin otomatis per tipe soal (proporsional, benar-semua/0,
  normalisasi isian, dst.) tercantum pada **tabel 16 tipe soal di UC03 (Bagian B)**;
  UC10 hanya mengatur koreksi manual (skor 0–100 → poin = skor/100 × bobot).
- **Sikap penggabung:** rujukan yang benar adalah UC03 untuk penilaian otomatis,
  UC10 untuk koreksi manual. Teks UC08 di atas dipertahankan apa adanya; catatan
  ini yang mengoreksi rujukannya.

### 3. Halaman "Kisi-kisi": UC04 (Bagian B) vs UC02 (Bagian A) — BUKAN pertentangan substansi, perlu verifikasi

- **Bagian B, UC04, format Word** menulis `KISI-KISI:` / `KOMPETENSI:` diisi kode
  "yang tampil di halaman Kisi-kisi / Kompetensi bank ini".
- **Bagian A, UC02** menyatakan tidak ada **halaman kelola** kompetensi/kisi-kisi
  di tampilan; halaman Kompetensi yang ada hanya menampilkan nilai per kompetensi
  (baca saja).
- **Penilaian penggabung:** kedua pernyataan dapat sama-sama benar bila yang ada
  adalah halaman tampil-baca Kisi-kisi (bukan halaman kelola). Bagian B tidak
  pernah menyebut halaman kelola. Perlu verifikasi satu kali apakah halaman
  tampil-baca Kisi-kisi memang ada; bila tidak ada, penyebutan di UC04 harus
  dikoreksi menjadi "kode kisi-kisi (bila diketahui)" dan kolom itu tetap
  dikosongkan dalam praktik (E30).

### 4. Interval penyegaran daftar usulan buka blokir: UC18 (Bagian A) vs UC07 (Bagian C) — PERBEDAAN PRESISI

- **Bagian A, UC18:** halaman pemantau memuat daftar usulan "setiap 10 detik".
- **Bagian C, UC07, sub-alur 3:** daftar usulan "diperbarui berkala tiap
  beberapa detik".
- **Penilaian penggabung:** bukan pertentangan — C hanya kurang presisi. Angka
  yang mengikat adalah **10 detik** (A). Rancangan diagram memakai 10 detik.

### 5. Hal yang TIDAK bertentangan (dicatat agar tidak dianggap pertentangan)

- Token peka huruf besar + satu titik masuk mengubah ke huruf besar otomatis,
  titik masuk lain tidak (B UC09 dan C UC08 — sama).
- Pengerjaan diblokir tidak ikut dibekukan/dilanjutkan oleh jeda/lanjut jadwal;
  sisa waktunya dibekukan tersendiri saat pemblokiran (B UC06 dan C UC07 — sama).
- Impor siswa (UC02) TANPA pratinjau vs impor soal (UC04) DENGAN pratinjau —
  dua fitur berbeda, bukan pertentangan.
- Tidak ada tombol usulan buka blokir di tampilan; usulan hanya tersedia di
  layanan belakang (A UC18 dan C UC07 § "Yang TIDAK ditemukan" — sama).
- Ragu saja tanpa jawaban dihitung BELUM dijawab, baik di tampilan maupun di
  penghitung server (C UC08 dan C UC11 — sama).

## Temuan lintas UC (gabungan)

1. **Kunci bank soal adalah satu-satunya pengaman isi soal.** Kunci tidak
   menghentikan ubah metadata bank (kode/nama/mapel/kelas tetap dapat diubah saat
   bank terkunci); yang diblokir adalah mutasi soal + impor + hapus bank.
2. **Pemisahan tugas kunci bank** (pembuat tidak boleh mengunci bank sendiri) hanya
   berlaku untuk aksi kunci/buka kunci — bukan untuk mutasi soal (guru pemilik
   tetap boleh mengubah soalnya sendiri).
3. **Impor selalu menulis ulang dari berkas** pada konfirmasi; pratinjau tidak
   pernah dipakai sebagai dasar simpan.
4. **Soal hasil impor selalu Draf** — tidak ada impor langsung Aktif.
5. **Ubah status jadwal lewat formulir tidak membekukan timer** — hanya aksi
   Pause/Resume khusus yang membekukan/memulihkan timer. Ini perbedaan perilaku
   yang penting bagi proktor.
6. **Token tidak pernah dapat dilihat lagi** sesudah generate; satu-satunya
   pemulihan adalah generate token baru (yang menghanguskan token lama).
7. **Kartu peserta tidak memuat token** dalam kondisi apa pun; token hanya dari
   proktor saat ujian.
8. **Tiga perilaku muncul di dua UC dari dua sudut pandang** (bukan duplikasi yang
   harus dihapus): jeda/lanjutkan jadwal (UC06 pemilik definisi, UC07 sudut
   pandang pemantau); validasi token (UC09 pemilik definisi, UC08 pemanggil);
   usul/proses buka blokir (UC18 pemilik definisi, UC07 sudut pandang pemantau).
   Rancangan diagram menyelesaikan ini dengan simpul pemanggil (call behavior),
   bukan dengan menggambar alur yang sama dua kali.
9. **Mode remedial selalu "tertinggi"** (UC17 pengaturan + UC11 remedial): nilai
   akhir = nilai tertinggi antara nilai awal dan nilai remedial; nilai awal tidak
   pernah ditimpa.
10. **Batas waktu pengerjaan = jam selesai jadwal**, bukan waktu mulai ditambah
    durasi paket; durasi paket hanya ditampilkan (UC05, UC08).
11. **Finalisasi malas (lazy):** tidak ada proses latar yang menyapu pengerjaan
    kedaluwarsa; finalisasi terjadi pada permintaan berikutnya yang menyentuh
    pengerjaan itu (UC08).
12. **Siswa tidak menerima siaran apa pun** — siswa mengetahui jeda/blokir/selesai
    hanya lewat penarikan status berkala atau permintaan berikutnya (UC07, UC08,
    UC18).

## Hal yang TIDAK ditemukan di kode (gabungan Bagian A + Bagian C; daftar per-UC Bagian B tetap di dalam masing-masing UC di atas)

1. **Tidak ada halaman daftar jejak audit di tampilan.** Layanan baca jejak audit
   ada, tetapi tidak ada pemanggilan dari kode tampilan. (A)
2. **Tidak ada formulir pengecualian SoD di tampilan.** Layanan pencatatan
   pengecualian ada, tetapi tidak ada pemanggilan dari kode tampilan. (A)
3. **Tidak ada tombol/formulir membuat usulan buka blokir di tampilan.**
   Layanan pembuatan usulan ada, tetapi kode tampilan hanya MEMBACA daftar usulan
   (di halaman pemantau). (A + C)
4. **Tidak ada halaman kelola kompetensi / kisi-kisi di tampilan.** Operasi
   kelolanya ada di layanan belakang; halaman Kompetensi yang ada hanya
   menampilkan nilai per kompetensi (baca saja). (A; lihat juga Pertentangan 3)
5. **Tidak ada status usulan "ditolak".** Usulan buka blokir hanya mengenal
   menunggu dan disetujui; tidak ada penolakan usulan. (A)
6. **Tidak ada pemulihan hapus (hapus lunak / tempat sampah).** Semua hapus
   master data adalah hapus keras permanen; perlindungan satu-satunya adalah guard
   pemakaian dan dialog konfirmasi. (A)
7. **Tidak ada penguncian login berbasis basis data.** Penguncian login (5 gagal /
   15 menit) disimpan di memori proses — tidak bertahan bila proses dimulai ulang. (A)
8. **Tidak ada pemberitahuan dorong / siaran ke siswa** untuk buka blokir maupun
   notifikasi umum. Siswa mengetahui status lewat pemeriksaan berkala; notifikasi
   hanya dibaca lewat bel / daftar tarik. (A + C)
9. **Tidak ada pembuatan akun staf dari tampilan master data.** Akun login dibuat
   otomatis hanya untuk siswa. Guru hanya dapat dikaitkan ke pengenal akun yang
   sudah ada. (A)
10. **Tidak ada pratinjau (dry-run) pada impor siswa CSV.** Import siswa langsung
    menyimpan baris valid dalam satu transaksi. (A)
11. **Tidak ada aksi "paksa selesai"** dari pemantauan. Yang ada: pantau,
    jeda/lanjutkan jadwal, usulkan buka blokir, buka blokir. (C)
12. **Tidak ada halaman tampilan khusus koreksi.** Layanan daftar menunggu koreksi
    + simpan skor ada di sisi server, tetapi tidak ada halaman yang memanggilnya
    di kode tampilan yang dibaca. (C)
13. **Tidak ada proses latar untuk finalisasi otomatis.** Finalisasi saat batas
    waktu lewat terjadi secara malas pada permintaan berikutnya. (C)
14. **Tidak ada kirim ulang otomatis autosave.** Jawaban yang gagal tersimpan tidak
    dikirim ulang sendiri; siswa harus mengubah jawaban lagi. (C)
15. **Tidak ada isian catatan pada berita acara.** Catatan selalu kosong dari
    sistem. (C)
16. **Remedial bukan ujian ulang.** Remedial adalah entri nilai manual (nilai
    tertinggi menang). Tidak ada pengerjaan/token/soal remedial. (C)
17. **Tidak ada penarikan otomatis nilai teori UKK dari ujian.** Nilai teori UKK
    diisi manual. (C)
18. **Tidak ada backup otomatis harian di dalam aplikasi.** Backup otomatis hanya
    lewat skrip cron di luar aplikasi; halaman backup hanya manual. (C)
19. **Tidak ada halaman hasil untuk staf.** Halaman hasil hanya untuk siswa
    pemilik pengerjaan; staf melihat nilai lewat rekap/analisis/koreksi. (C)

## Sumber telusur

- Bagian A dan Bagian B tidak mencantumkan daftar berkas kode yang dibaca.
- Daftar berkas di bawah berasal dari Bagian C (cakupan UC07, UC08, UC10, UC11,
  UC12, UC13, UC14, UC16, UC19) dan dipertahankan apa adanya.

## File kode yang dibaca (telusur Kelompok C)

### Sisi server

| # | Berkas | Isi yang dipakai |
|---|--------|------------------|
| 1 | `backend/internal/exam/exam.go` | Mulai ujian + validasi token/toleransi, muat soal, simpan jawaban, tandai ragu, pelanggaran, submit, status, hasil per soal + ringkasan kunci, koreksi (daftar + simpan skor), pause/resume jadwal, usulan unblock, unblock, finalisasi malas |
| 2 | `backend/internal/grading/grading.go` | Aturan poin per tipe soal, aturan nilai akhir, penanda dinilai/manual |
| 3 | `backend/internal/monitor/monitor.go` | Tiket sekali pakai, hub siaran per jadwal, potret peserta, heartbeat |
| 4 | `backend/internal/rekap/rekap.go` | Rekap nilai + statistik, ekspor CSV, berita acara, finalisasi, remedial |
| 5 | `backend/internal/analisis/analisis.go` | Analisis butir (kesukaran, daya beda, pengecoh) + nilai per kompetensi + ekspor CSV + penyembunyian pertanyaan untuk Proktor |
| 6 | `backend/internal/ukk/ukk.go` | CRUD UKK, nilai UKK + rumus praktik/akhir, berkas praktik, sertifikat |
| 7 | `backend/internal/backup/backup.go` | Backup manual, daftar, unduh, restore + backup pengaman + gerbang ujian, hapus + perlindungan pre-restore muda |

### Sisi tampilan

| # | Berkas | Isi yang dipakai |
|---|--------|------------------|
| 8 | `frontend/src/pages/siswa/UjianPage.tsx` | Daftar jadwal + mulai, soal, autosave, ragu, anti-curang, timer, submit, kartu hasil |
| 9 | `frontend/src/pages/siswa/SiswaHomePage.tsx` | Titik masuk ujian #1: daftar jadwal hari ini + input token + mulai |
| 10 | `frontend/src/pages/siswa/HasilPage.tsx` | Hasil per soal + pembahasan + status per soal |
| 11 | `frontend/src/pages/admin/monitor/MonitorPage.tsx` | Pemantauan real-time + fallback, pause/resume, unblock, daftar usulan |
| 12 | `frontend/src/pages/admin/rekap/RekapPage.tsx` | Rekap + statistik + finalisasi + remedial + ekspor |
| 13 | `frontend/src/pages/admin/rekap/BeritaAcaraPage.tsx` | Berita acara + cetak |
| 14 | `frontend/src/pages/admin/analisis/AnalisisButirPage.tsx` | Analisis butir + ekspor |
| 15 | `frontend/src/pages/admin/analisis/KompetensiPage.tsx` | Nilai per kompetensi + ekspor |
| 16 | `frontend/src/pages/admin/ukk/UkkListPage.tsx` | Daftar + buat + hapus UKK |
| 17 | `frontend/src/pages/admin/ukk/UkkDetailPage.tsx` | Detail UKK + ubah status + nilai peserta + berkas + terbit sertifikat |
| 18 | `frontend/src/pages/admin/ukk/UkkNilaiPage.tsx` | Input nilai UKK + pratinjau hitungan |
| 19 | `frontend/src/pages/admin/ukk/UkkSertifikatPage.tsx` | Tampilan sertifikat + cetak |
| 20 | `frontend/src/pages/admin/backup/BackupPage.tsx` | Daftar + buat + unduh + restore + hapus backup |
| 21 | `frontend/src/App.tsx` | Rute halaman Kelompok C (hanya dibaca bagian rute) |

### Referensi silang (bukan sumber alur)

| # | Berkas | Pemakaian |
|---|--------|-----------|
| 22 | `docs/uml/ALUR-SEQUENCE-NYATA.md` | Pembanding alur login/ujian/pemantauan yang sudah ditelusuri sebelumnya; alur Kelompok C ditulis ulang dari kode, bukan disalin dari dokumen ini |

> Catatan: sesuai tugas, berkas diagram `.mdj` TIDAK ditulis/dimodifikasi. Satu-satunya berkas yang ditulis worker ini adalah dokumen telusur ini.
