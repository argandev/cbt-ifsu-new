# RANCANGAN DIAGRAM ACTIVITY — CBT SMK v2

> **Status:** Rancangan Fase 2. Dokumen ini HANYA rancangan — berkas diagram
> (`.mdj`) TIDAK ditulis/diubah di sini (tim Sequence sedang menulisnya paralel).
> **Sumber alur:** `ALUR-ACTIVITY-NYATA.md` (gabungan UC01–UC19, telusur dari kode,
> 18 Okt 2026). Setiap butir di bawah merujuk ke UC sumbernya; bila rancangan ini
> berbeda dari telusur, perbedaan itu ditandai eksplisit di §5.
> **Aturan mengikat untuk penulis diagram:**
> (1) Nama diagram DAN dokumentasi diagram WAJIB menyebut kode UC yang dicakup
> (mis. nama `D-ACT-08 — Ikuti Ujian (UC08)`, dokumentasi diawali `UC08 — …`).
> (2) Level DESAIN (PIM), Bahasa Indonesia, tanpa nama berkas / alamat layanan /
> perintah basis data / pustaka / nama tabel-kolom / nama fungsi.
> (3) Alur nyambung: tanpa elemen melayang — setiap aksi terhubung dari simpul
> awal sampai ke simpul akhir; setiap cabang decision bertemu merge-nya (atau
> berakhir di simpul akhir yang sah).
> (4) Guard ditulis dalam bahasa desain di dalam kurung siku, mis. `[token cocok]`.

---

## 1. Daftar diagram yang dirancang

**Jumlah diagram: 19 — satu diagram per UC, TANPA penggabungan.**

| Kode diagram | Judul diagram (nama + kode UC) | UC | Penanda | Alasan penandaan |
|---|---|---|---|---|
| D-ACT-01 | Masuk & Wajib Ganti Kata Sandi (UC01) | UC01 | **UTAMA** | Login — gerbang masuk seluruh aplikasi |
| D-ACT-02 | Kelola Master Data (UC02) | UC02 | LAMPIRAN | Penyiapan data induk, bukan siklus ujian |
| D-ACT-03 | Kelola Bank Soal (UC03) | UC03 | LAMPIRAN | Penyiapan soal, bukan siklus ujian |
| D-ACT-04 | Impor Soal (UC04) | UC04 | **UTAMA** | Impor — disebut eksplisit sebagai alur inti |
| D-ACT-05 | Susun Paket Ujian (UC05) | UC05 | LAMPIRAN | Penyiapan paket, bukan siklus ujian |
| D-ACT-06 | Kelola Jadwal Ujian (UC06) | UC06 | LAMPIRAN | Penyiapan jadwal; pause/resume-nya dipakai UC07 lewat pemanggil |
| D-ACT-07 | Pantau Ujian Langsung (UC07) | UC07 | **UTAMA** | Pantau — disebut eksplisit sebagai alur inti |
| D-ACT-08 | Ikuti Ujian (UC08) | UC08 | **UTAMA** | Ujian — alur inti pelaksanaan |
| D-ACT-09 | Validasi Token (UC09) | UC09 | **UTAMA** | Gerbang ujian — tanpa token yang sah, UC08 tidak dapat dimulai |
| D-ACT-10 | Koreksi Esai / Manual (UC10) | UC10 | **UTAMA** | Penilaian/koreksi — disebut eksplisit sebagai alur inti |
| D-ACT-11 | Rekap Nilai & Berita Acara (UC11) | UC11 | **UTAMA** | Rekap — disebut eksplisit sebagai alur inti; memuat aksi pengubah status (finalisasi, remedial) |
| D-ACT-12 | Analisis Butir Soal (UC12) | UC12 | LAMPIRAN | Laporan analitis baca-saja pasca-penilaian, bukan alur penilaian inti |
| D-ACT-13 | Nilai per Kompetensi (UC13) | UC13 | LAMPIRAN | Laporan baca-saja pasca-penilaian, bukan alur penilaian inti |
| D-ACT-14 | Kelola UKK (UC14) | UC14 | LAMPIRAN | Alur UKK terpisah (bukan pengerjaan soal ujian) |
| D-ACT-15 | Kelola Kegiatan, Ruangan & Kartu Peserta (UC15) | UC15 | LAMPIRAN | Penyiapan kegiatan/ruangan/kartu, bukan siklus ujian |
| D-ACT-16 | Cadangkan & Pulihkan (UC16) | UC16 | LAMPIRAN | Administrasi basis data, bukan siklus ujian |
| D-ACT-17 | Kelola Pengaturan, Audit & Notifikasi (UC17) | UC17 | LAMPIRAN | Administrasi pengaturan/audit, bukan siklus ujian |
| D-ACT-18 | Usul & Proses Buka Blokir (UC18) | UC18 | **UTAMA** | Lanjutan ujian yang diblokir — tanpa buka blokir, ujian yang terblokir tidak pernah selesai |
| D-ACT-19 | Lihat Hasil & Pembahasan (UC19) | UC19 | LAMPIRAN | Tampilan baca-saja hasil sesudah penilaian, bukan alur penilaian |

**Prinsip penandaan:** UTAMA = alur yang menggerakkan siklus ujian
(masuk → gerbang token → ujian → pantau → buka blokir → koreksi → rekap) +
impor soal. LAMPIRAN = penyiapan (master, bank, paket, jadwal, kegiatan, UKK,
pengaturan, backup) + laporan/tampilan baca-saja (analisis, kompetensi, hasil).

### 1.1 Mengapa TIDAK ada penggabungan (0 diagram gabungan)

Pasangan yang diperiksa dan DITOLAK penggabungannya, karena alurnya tidak
benar-benar identik:

| Pasangan yang dipertimbangkan | Alasan ditolak |
|---|---|
| UC12 Analisis Butir + UC13 Nilai per Kompetensi (sama-sama laporan on-demand) | Hitungan berbeda (kesukaran/daya beda/pengecoh vs nilai per kompetensi + rata-rata kelas); penyembunyian proktor hanya di UC12; batasan siswa-melihat-sendiri hanya di UC13; pengelompokan berbeda (per soal vs per kompetensi + "Tanpa Kompetensi") |
| UC02 Kelola Master Data + UC15 Kelola Kegiatan/Ruangan (sama-sama CRUD) | UC15 juga memuat penugasan per jadwal + kartu peserta 1-kartu (bukan CRUD); guard kapasitas ruangan dan penomoran peserta tidak ada padanannya di UC02 |
| UC06 Pause/Resume + UC07 Pantau (jeda tampil di keduanya) | UC06 pemilik definisi pause/resume; UC07 hanya memanggilnya dari pemantauan. Diselesaikan dengan simpul pemanggil, bukan penggabungan |
| UC09 Validasi Token + UC08 Ikuti Ujian (validasi token tampil di keduanya) | UC09 juga memuat generate/salin/cek status token (pengelolaan token); UC08 hanya memanggil validasinya saat mulai. Diselesaikan dengan simpul pemanggil |
| UC18 Buka Blokir + UC07 Pantau (buka blokir tampil di keduanya) | UC07 hanya memanggil usulan/eksekusi dari pemantauan; definisi lengkap (hitung sisa, tutup usulan, jejak) milik UC18. Diselesaikan dengan simpul pemanggil |
| UC10 Koreksi + UC11 Rekap (sama-sama soal nilai) | Aktor dan objek berbeda (koreksi = skor per jawaban; rekap = finalisasi/remedial/berita acara per jadwal); koreksi mengubah poin jawaban, rekap mengubah status nilai |

**Konsekuensi:** tiap UC mendapat tepat satu diagram. Alur bersama yang muncul di
lebih dari satu UC (pause/resume, validasi token, usul/buka blokir) digambar
SEKALI pada diagram UC pemiliknya, dan diagram UC pemanggil memakai simpul
pemanggil perilaku (*call behavior action*) yang merujuk ke diagram pemilik —
tidak digambar ulang, tidak diduplikasi.

---

## 2. Rancangan per diagram

Konvensi penulisan di bawah: **Aksi** = simpul aksi; **Keputusan** = decision +
merge pasangannya; **Garpu/Gabung** = fork + join pasangannya (bila ada);
**Guard** ditulis `[…]` dalam bahasa desain. Setiap diagram dibuka satu simpul
awal dan ditutup simpul akhir; hasil akhir yang berbeda (tersimpan / ditolak /
selesai) memakai simpul akhir sendiri-sendiri yang berlabel jelas.

### D-ACT-01 — Masuk & Wajib Ganti Kata Sandi (UC01) — UTAMA

- **Dokumentasi diagram:** `UC01 — Masuk & Wajib Ganti Kata Sandi. Aktor: Pengguna (siswa/staf), Sistem.`
- **Partisi:** Pengguna | Sistem.
- **Simpul awal:** `Mulai masuk`. **Simpul akhir:** `Masuk berhasil`; `Masuk ditolak`; `Keluar selesai`; `Kata sandi diganti` (mengalir kembali ke `Mulai masuk`, bukan akhir — pengguna wajib masuk ulang).
- **Aksi (urutan utama):** Buka aplikasi → Periksa sesi → [bila sesi berlaku] Tampilkan beranda sesuai peran → Isi email/nama pengguna + kata sandi → Kirim formulir → Periksa kelengkapan isian → Cari akun (email ATAU nama pengguna, tanpa beda huruf, spasi diabaikan) → Periksa kunci login sementara → Verifikasi kata sandi → Periksa status aktif → Periksa penanda wajib ganti → Buat sesi baru → Catat jejak audit masuk → Arahkan ke beranda sesuai peran.
- **Cabang ganti kata sandi (dari keputusan wajib ganti, dan dari dalam sesi):** Tampilkan halaman ganti kata sandi (keterangan sesuai konteks) → Isi nama login + kata sandi lama + baru + konfirmasi → Validasi di tampilan (baru ≥ 8 karakter, konfirmasi sama, baru ≠ lama) → Tentukan konteks (sesi aktif / tanpa sesi) → Verifikasi kata sandi lama → Perbarui kata sandi tersimpan → Matikan penanda wajib ganti → Hapus SELURUH sesi pengguna → Catat jejak audit ganti kata sandi → Kembali ke halaman masuk.
- **Cabang keluar:** Keluar → Hapus sesi saat itu → Bersihkan penanda sesi → Catat jejak audit keluar → Kosongkan status login tersimpan.
- **Keputusan + guard:** `[sesi berlaku]` / `[sesi kedaluwarsa]`; `[isian kosong]`; `[nama login terkunci]` (5 gagal / 15 menit); `[akun tidak ditemukan]`; `[kata sandi salah]`; `[akun tidak aktif]`; `[wajib ganti kata sandi]`; `[token pengaman tidak cocok]` (pada perubahan data); `[sudah login]` (buka halaman masuk saat login); `[kata sandi lama salah]`; `[kata sandi baru tidak sah]`.
- **Garpu/Gabung:** tidak ada (alur sekuensial; penguncian dihitung per nama login, bukan cabang paralel).
- **Catatan:** Penolakan akun tidak aktif dan wajib ganti TIDAK menambah hitungan penguncian; hanya "akun tidak ditemukan" dan "kata sandi salah" yang menambah. Pesan penolakan kredensial generik (tidak membedakan akun tidak ditemukan vs kata sandi salah).

### D-ACT-02 — Kelola Master Data (UC02) — LAMPIRAN

- **Dokumentasi diagram:** `UC02 — Kelola Master Data (jurusan, kelas, mapel, guru, siswa, termasuk impor siswa). Aktor: Staf pengelola, Sistem.`
- **Partisi:** Staf Pengelola | Sistem.
- **Simpul awal:** `Mulai kelola master data`. **Simpul akhir:** `Data tersimpan`; `Data terhapus`; `Aksi ditolak`; `Impor selesai` (dengan ringkasan).
- **Aksi (pola umum, berlaku untuk kelima entitas):** Buka halaman daftar entitas → Periksa izin baca → Tampilkan daftar (cari + saring + penomoran) → Pilih aksi [tambah / ubah / hapus] → [tambah/ubah] Isi formulir → Validasi isian → Periksa keunikan → Simpan → Catat jejak audit → [hapus] Tampilkan dialog konfirmasi → Konfirmasi → Periksa pemakaian data (guard) → Hapus permanen → Catat jejak audit.
- **Keputusan + guard (per entitas, sebagai cabang dari keputusan "entitas mana"):** `[entitas = jurusan]` (kode + nama wajib; guard hapus: dipakai kelas / mapel produktif / siswa); `[entitas = kelas]` (jurusan wajib, tingkat 10/11/12, kombinasi nama+tingkat+tahun ajaran unik per jurusan; guard hapus: dipakai siswa / jadwal); `[entitas = mapel]` (jenis umum/produktif; produktif wajib jurusan, umum wajib tanpa jurusan; guard hapus: dipakai kompetensi / kisi-kisi / bank soal / paket / jadwal); `[entitas = guru]` (nama wajib, NIP unik bila diisi, akun login opsional; guard hapus: pemilik bank soal); `[entitas = siswa]` (NISN + nama + kelas + jurusan wajib; tambah manual membuat akun login otomatis bila tidak diisi akun — nama pengguna = NISN, kata sandi awal = NISN, penanda wajib ganti aktif; ubah mempertahankan/mengosongkan akun sesuai isian; guard hapus: punya pengerjaan / nilai akhir / nilai remedial → tolak, arahkan ke nonaktifkan).
- **Cabang impor siswa (dari daftar siswa):** Buka halaman impor → Unduh templat (opsional) → Pilih berkas CSV → Periksa berkas (format, ukuran ≤ 5 MB, tidak kosong) → Normalisasi kepala kolom (alias diterima) → Periksa kolom wajib (nama, NISN, kelas) → Proses tiap baris (wilayah perluasan: satu aksi per baris — periksa nama/NISN/kelas, duplikat dalam berkas, NISN sudah ada → hitung "sudah ada", kelas tidak ditemukan → gagal, lolos → buat akun + simpan siswa aktif) → Terapkan transaksi → Catat jejak audit impor → Tampilkan ringkasan (berhasil / gagal / sudah ada + tabel kesalahan per baris).
- **Keputusan impor + guard:** `[berkas sah]`; `[kolom wajib ada]`; `[baris ≤ 2.000]`; `[NISN sudah terdaftar]` (→ sudah ada, bukan gagal); `[kelas ditemukan]`.
- **Garpu/Gabung:** tidak ada. Pemrosesan per baris dimodelkan sebagai wilayah perluasan (expansion region), bukan fork.
- **Catatan:** Kompetensi/kisi-kisi TIDAK digambar sebagai kelola (tidak ada halaman kelolanya); hanya muncul sebagai guard hapus mapel. Impor siswa TIDAK memakai pratinjau — langsung simpan (berbeda dari UC04).

### D-ACT-03 — Kelola Bank Soal (UC03) — LAMPIRAN

- **Dokumentasi diagram:** `UC03 — Kelola Bank Soal. Aktor: Admin, Guru, Proktor (daftar saja), Sistem.`
- **Partisi:** Pengguna | Sistem.
- **Simpul awal:** `Mulai kelola bank soal`. **Simpul akhir:** `Bank/soal tersimpan`; `Bank/soal terhapus`; `Status kunci berubah`; `Aksi ditolak`.
- **Aksi (cabang bank):** Lihat daftar bank → Buat bank (isi kode, nama, mapel, tingkat) → Validasi → Simpan (guru pembuat = pemilik) → Ubah bank (validasi sama; kunci tidak ikut berubah) → Hapus bank (konfirmasi → periksa tidak terkunci + tidak dipakai paket → hapus bank + seluruh soalnya).
- **Aksi (cabang soal):** Lihat isi bank + saring soal → Tambah soal (editor: pilih tipe → isi stimulus/pertanyaan/media/jawaban/kunci/pembahasan/bobot/level/status) → Pratinjau (tampilan peserta tanpa kunci; pratinjau guru dengan kunci) → Validasi struktur per tipe → Simpan → Ubah soal (ganti tipe: bila jawaban/kunci kosong → ganti langsung; bila terisi → minta konfirmasi karena jawaban/kunci dikosongkan) → Hapus soal (konfirmasi → periksa bank tidak terkunci → hapus).
- **Aksi (cabang kunci):** Kunci bank (konfirmasi) / Buka kunci bank → Periksa pemisahan tugas (pembuat bukan pengunci) → Ubah status kunci.
- **Keputusan + guard:** `[izin kelola bank]`; `[proktor]` (→ hanya metadata, tanpa isi soal); `[bank terkunci]` (→ tambah/ubah/hapus/impor soal ditolak, hapus bank ditolak; ubah metadata bank tetap boleh); `[kode bank dipakai]`; `[mapel ditemukan]`; `[tingkat 10/11/12]`; `[bank milik guru lain]`; `[soal dipakai paket]`; `[pembuat mengunci bank sendiri]` (→ ditolak); `[tipe dikenal, 16 tipe]`; `[pertanyaan terisi]`; `[bobot > 0]`; `[struktur jawaban/kunci sah per tipe]`; `[ganti tipe + jawaban/kunci terisi]` (→ konfirmasi).
- **Garpu/Gabung:** tidak ada.
- **Catatan:** Kunci bank adalah satu-satunya pengaman isi soal. Urutan tampil soal = urutan pembuatan (tidak ada urut-ulang).

### D-ACT-04 — Impor Soal (UC04) — UTAMA

- **Dokumentasi diagram:** `UC04 — Impor Soal (Excel + Word). Aktor: Admin, Guru, Sistem.`
- **Partisi:** Pengguna | Sistem.
- **Simpul awal:** `Mulai impor soal`. **Simpul akhir:** `Soal tersimpan sebagai Draf`; `Tidak ada yang tersimpan`; `Impor ditolak`.
- **Aksi (tahap pratinjau):** Pilih format (Excel/Word) → Unduh templat → Isi berkas (satu berkas = satu bank) → Pilih berkas → Periksa di tampilan (ekstensi sesuai format, tidak kosong, ≤ 5 MB) → Kirim untuk pratinjau (dry-run, TIDAK menyimpan) → Periksa berkas (format sah, lembar "Soal" ada untuk Excel, kolom wajib ada, ≤ 500 baris) → Validasi tiap baris (wilayah perluasan) → Kembalikan ringkasan + tabel pratinjau + daftar temuan → Baca hasil pratinjau (total, valid, error, warning, duplikat, total bobot, ringkasan per tipe; kunci dipetakan ke label + teks opsi).
- **Aksi (tahap simpan):** Pilih mode simpan [tolak-semua-bila-error / lewati-baris-error] + perlakuan duplikat [lewati / simpan ganda] → Konfirmasi simpan → Kirim ulang berkas + validasi ulang dari awal → Simpan baris sesuai mode → Perbarui daftar bank → Tampilkan hasil (tersimpan / gagal / duplikat dilewati + warning + error) → Telaah soal Draf di editor (di luar diagram ini — lihat D-ACT-03).
- **Keputusan + guard:** `[berkas sah di tampilan]`; `[bank terkunci]` (→ pratinjau + simpan ditolak); `[error level berkas]`; `[baris error]`; `[mode = tolak-semua + ada error]` (→ tidak ada yang tersimpan); `[mode = lewati]` (→ baris valid tersimpan); `[baris duplikat + lewati]` / `[baris duplikat + simpan ganda]`; `[ada baris valid]`.
- **Garpu/Gabung:** tidak ada (pratinjau lalu simpan adalah dua fase sekuensial, bukan paralel; simpan selalu memvalidasi ulang).
- **Catatan:** Kolom kisi-kisi yang terisi selalu gagal (E30) — kolom itu dikosongkan dalam praktik. Gambar tempelan tidak terimpor; hanya tautan. Teks alternatif gambar tidak tersimpan.

### D-ACT-05 — Susun Paket Ujian (UC05) — LAMPIRAN

- **Dokumentasi diagram:** `UC05 — Susun Paket Ujian. Aktor: Admin, Guru, Sistem.`
- **Partisi:** Pengguna | Sistem.
- **Simpul awal:** `Mulai susun paket`. **Simpul akhir:** `Paket tersimpan`; `Paket terhapus`; `Susunan diperbarui`; `Aksi ditolak`.
- **Aksi (cabang paket):** Buat paket (isi kode, nama, mapel, jenis STS/SAS/USBK/UKK/Tryout, durasi, saklar acak soal / acak opsi / tampil hasil) → Validasi → Simpan (paket kosong) → Ubah paket (validasi sama; cabang ganti mapel: bila susunan berisi soal mapel lain → tolak "kosongkan susunan dulu") → Hapus paket (konfirmasi → periksa tidak dipakai jadwal → hapus paket + susunannya).
- **Aksi (cabang susun):** Buka detail paket → Pilih bank se-mapel → Tampilkan soal bank → Tambah soal manual (periksa soal se-mapel → tambah di akhir, nomor urut = terbesar + 1; soal sudah ada → tanpa perubahan) → Hapus soal dari paket (soal tetap ada di bank; nomor urut lain tidak dirapatkan) → Generate acak (isi jumlah → ambil acak soal Aktif yang belum ada di paket dari bank se-mapel → tambah berurutan di akhir).
- **Keputusan + guard:** `[kode paket dipakai]`; `[jenis sah, 5 jenis]`; `[durasi > 0]`; `[mapel ditemukan]`; `[paket dipakai jadwal]` (→ hapus ditolak); `[ganti mapel + susunan berisi soal mapel lain]`; `[soal se-mapel dengan paket]`; `[soal sudah di paket]`; `[jumlah acak > 0]`; `[soal tersedia < diminta]` (→ tambah yang tersedia + peringatan).
- **Garpu/Gabung:** tidak ada.
- **Catatan:** Durasi paket hanya ditampilkan; batas waktu peserta mengikuti jam selesai jadwal (lihat D-ACT-08). Tambah manual tidak menyaring status soal; hanya generate acak yang membatasi ke soal Aktif.

### D-ACT-06 — Kelola Jadwal Ujian (UC06) — LAMPIRAN

- **Dokumentasi diagram:** `UC06 — Kelola Jadwal Ujian. Aktor: Admin, Proktor, Guru, Pengawas (lihat), Sistem.`
- **Partisi:** Staf | Sistem.
- **Simpul awal:** `Mulai kelola jadwal`. **Simpul akhir:** `Jadwal tersimpan`; `Jadwal terhapus`; `Jadwal dijeda`; `Jadwal dilanjutkan`; `Aksi ditolak`.
- **Aksi (cabang jadwal):** Buat jadwal (isi paket, kelas, tanggal, jam mulai, jam selesai, status bawaan Draf) → Validasi → Simpan → Ubah jadwal (validasi sama; bila status Selesai → tolak seluruhnya; ubah status lewat formulir HANYA mengubah label) → Hapus jadwal (konfirmasi → periksa tidak ada pengerjaan peserta → hapus) → Lihat daftar jadwal (cari nama paket, saring status, tampilkan status token per jadwal).
- **Aksi (cabang jeda — definisi pemilik, dipanggil D-ACT-07):** Pause jadwal → Periksa jadwal Aktif → Ubah status menjadi Jeda → Bekukan sisa waktu semua pengerjaan berjalan (sisa = tenggat − sekarang, minimal 0; pengerjaan diblokir TIDAK ikut dibekukan) → Siarkan peristiwa jeda ke pemantau.
- **Aksi (cabang lanjut — definisi pemilik, dipanggil D-ACT-07):** Resume jadwal → Periksa jadwal Jeda → Ubah status menjadi Aktif → Beri tenggat baru tiap pengerjaan berjalan (sekarang + sisa beku) + bersihkan penanda beku → Siarkan peristiwa lanjut ke pemantau.
- **Keputusan + guard:** `[isian lengkap + tanggal/jam sah]`; `[jam selesai > jam mulai]`; `[status sah: Draf/Aktif/Jeda/Selesai]`; `[paket + kelas ditemukan]`; `[jadwal Selesai]` (→ ubah ditolak); `[ada pengerjaan peserta]` (→ hapus ditolak); `[jadwal Aktif]` (syarat pause); `[jadwal Jeda]` (syarat resume).
- **Garpu/Gabung:** tidak ada. Pembekuan/pemulihan banyak pengerjaan dimodelkan sebagai wilayah perluasan (satu aksi per pengerjaan), bukan fork.
- **Catatan:** Peserta pertama yang memulai mengubah jadwal Draf menjadi Aktif otomatis (lihat D-ACT-08). Status Selesai hanya ditetapkan manual lewat formulir. Mengubah status menjadi Jeda lewat formulir TIDAK sama dengan Pause (tidak membekukan timer).

### D-ACT-07 — Pantau Ujian Langsung (UC07) — UTAMA

- **Dokumentasi diagram:** `UC07 — Pantau Ujian Langsung. Aktor: Proktor, Admin, Guru, Pengawas (memantau), Sistem. Siswa adalah sumber peristiwa, bukan pelaku.`
- **Partisi:** Pemantau | Sistem.
- **Simpul awal:** `Mulai memantau`. **Simpul akhir:** `Pemantauan ditutup`; `Sambungan ditolak`.
- **Aksi:** Buka halaman pemantauan → Pilih jadwal (bila belum memilih → tampilkan daftar jadwal) → Tampilkan daftar peserta + ringkasan (total, sedang ujian, selesai, diblokir, jawaban terkumpul) → Sambung saluran langsung (tiket sekali pakai) → Terima potret daftar peserta → [garpu] Cabang siaran: terima peristiwa siswa (jawaban, ragu, pelanggaran, mulai/selesai, koreksi, jeda/lanjut, usulan, buka blokir) → Muat ulang potret | Cabang berkala: tarik daftar usulan tiap 10 detik → Tampilkan usulan menunggu sebagai teks → [gabung] Saring daftar menurut status → Tutup halaman → Tutup saluran.
- **Simpul pemanggil (call behavior, bukan alur ulang):** `Jeda Jadwal` → memanggil D-ACT-06 (cabang jeda); `Lanjutkan Jadwal` → memanggil D-ACT-06 (cabang lanjut); `Usulkan Buka Blokir` → memanggil D-ACT-18 (cabang usulkan); `Buka Blokir` → memanggil D-ACT-18 (cabang eksekusi).
- **Keputusan + guard:** `[jadwal dipilih]`; `[tiket sah]` (ada / belum dipakai / belum kedaluwarsa) — bila tidak sah → sambungan ditolak; `[saluran terputus]` → sambung ulang otomatis dengan tiket baru + mode penarikan berkala sementara; `[tombol Jeda aktif]` (jadwal berjalan); `[tombol Lanjut aktif]` (jadwal jeda).
- **Garpu/Gabung:** ADA — satu fork sesudah tersambung (cabang siaran langsung + cabang tarik berkala daftar usulan), satu join sebelum tutup halaman.
- **Catatan:** Tidak ada aksi "paksa selesai" dari pemantauan. Siswa tidak menerima siaran apa pun.

### D-ACT-08 — Ikuti Ujian (UC08) — UTAMA

- **Dokumentasi diagram:** `UC08 — Ikuti Ujian. Aktor: Siswa, Sistem.`
- **Partisi:** Siswa | Sistem.
- **Simpul awal:** `Mulai ujian`. **Simpul akhir:** `Ujian selesai (hasil tampil)`; `Ujian selesai (hasil disembunyikan)`; `Mulai ditolak`; `Ujian diblokir (menunggu buka blokir)`.
- **Aksi (mulai):** Buka beranda / halaman ujian → Tampilkan jadwal hari ini untuk kelas siswa → Pilih jadwal → Masukkan token (kecuali Tryout) → Tekan Mulai → [bila pengerjaan berjalan sudah ada → kembalikan pengerjaan lama, lanjut tanpa token baru] → Panggil validasi token (D-ACT-09) → Periksa toleransi telat + paket berisi soal → Buat pengerjaan baru (batas waktu = jam selesai jadwal) → Kocok soal / opsi sesuai saklar paket (hasil kocokan disimpan per pengerjaan) → [bila peserta pertama → ubah jadwal Draf menjadi Aktif] → Muat soal pengerjaan (tanpa kunci, tanpa pembahasan).
- **Aksi (mengerjakan — cabang utama garpu):** Kerjakan soal satu per satu (navigasi nomor soal: aktif / dijawab / ragu / belum) → Ubah jawaban → Simpan otomatis (tunggu 500 milidetik tanpa perubahan → kirim nilai terakhir → simpan + hitung poin per soal) → Tandai ragu-ragu (kirim segera; ragu boleh pada soal belum dijawab) → Tarik status berkala tiap 30 detik (sisa waktu, dijawab, ragu, status) → Tekan Selesai → Konfirmasi (jumlah dijawab/belum + peringatan belum dijawab = 0 + tidak dapat diubah) → Finalisasi (hitung nilai, status selesai, catat waktu kumpul) → Tampilkan kartu hasil pengumpulan.
- **Aksi (anti-curang — cabang kedua garpu):** Awasi pindah tab / halaman disembunyikan / jendela kehilangan fokus / keluar layar penuh → Laporkan peristiwa (deduplikasi 5 detik di tampilan dan di sistem) → Catat pelanggaran + jejak audit → [hitungan < batas] Tampilkan spanduk peringatan, ujian berlanjut | [hitungan mencapai batas] Blokir otomatis (status diblokir + bekukan sisa waktu) → Tampilkan layar diblokir.
- **Aksi (timer — menyatu dengan cabang utama):** Timer hitung mundur (merah bila sisa ≤ 5 menit) → Saat timer nol → Kumpulkan otomatis satu kali (dijaga tidak ganda). Finalisasi malas: setiap permintaan yang menyentuh pengerjaan lebih dulu memeriksa batas waktu lewat → finalisasi dulu bila lewat.
- **Keputusan + guard:** `[kelas siswa ditemukan]`; `[pengerjaan berjalan ada]` (→ lanjut); `[pengerjaan diblokir]` (→ ditolak, hubungi proktor); `[pengerjaan selesai]` (→ ditolak); `[jadwal belum mulai / sudah berakhir]`; `[telat melewati toleransi]` (hanya pengerjaan baru); `[paket tanpa soal]`; `[jadwal jeda]` (→ penyimpanan ditolak, timer beku); `[jadwal selesai]`; `[waktu habis]`; `[pelanggaran < batas]` / `[pelanggaran mencapai batas]`; `[pengumpulan ganda]` (→ kembalikan hasil tersimpan, idempotent); `[kumpul pada pengerjaan diblokir]` (→ ditolak).
- **Garpu/Gabung:** ADA — satu fork sesudah soal dimuat (cabang mengerjakan + cabang awasi anti-curang), satu join pada pengumpulan / pemblokiran / penutupan.
- **Catatan:** Urutan pemeriksaan saat mulai mengikuti D-ACT-09 (hipotesis urutan Bagian B — lihat §5 butir 1). Salin/tempel dicegah di tampilan saja, bukan pelanggaran. Tidak ada kirim ulang otomatis jawaban gagal; tidak ada proses latar finalisasi.

### D-ACT-09 — Validasi Token (UC09) — UTAMA

- **Dokumentasi diagram:** `UC09 — Validasi Token. Aktor: Admin, Proktor, Siswa (memasukkan token), Sistem.`
- **Partisi:** Admin/Proktor | Sistem | Siswa.
- **Simpul awal:** `Mulai kelola/validasi token`. **Simpul akhir:** `Token diterbitkan`; `Status ditampilkan`; `Token sah (lanjut ujian)`; `Token ditolak`.
- **Aksi (pengelolaan — partisi Admin/Proktor):** Buka dialog token → Tampilkan status token (tanpa nilai) → Generate token → Buat nilai acak 6 karakter (tanpa karakter mudah tertukar) → Simpan hanya sidik hash + waktu kedaluwarsa (waktu generate + masa berlaku) → Hanguskan token lama (tertimpa) → Catat jejak audit pembuatan (tanpa nilai token) → Tampilkan token baru SEKALI + waktu kedaluwarsa → Salin token (tombol "Tersalin!" 2 detik; bila gagal → pesan salin manual).
- **Aksi (cek status):** Tekan Status Token / lihat kolom token daftar jadwal → Kembalikan ada-tidak, waktu kedaluwarsa, sisa detik (minimal 0), sudah kedaluwarsa ya/tidak → Tampilkan "Belum dibuat" / "Aktif, kedaluwarsa dalam N detik" / "Kedaluwarsa" (nilai token TIDAK PERNAH dikembalikan).
- **Aksi (validasi saat mulai — dipanggil D-ACT-08, urutan hipotesis Bagian B, lihat §5 butir 1):** Terima token → Periksa jadwal ada → Periksa peserta dari kelas jadwal → Periksa jadwal belum Selesai + tidak Jeda → Periksa pengerjaan lama (berjalan → kembalikan; diblokir/selesai → tolak SEBELUM token) → Periksa waktu (belum mulai / sudah berakhir → tolak) → Periksa toleransi telat (pengerjaan baru) → Periksa token (kosong / jadwal belum punya token → tidak valid; sidik hash tidak cocok → tidak valid; melewati kedaluwarsa → kedaluwarsa) → Lolos (lanjut buat pengerjaan di D-ACT-08).
- **Keputusan + guard:** `[jadwal Selesai]` (→ generate ditolak); `[token lama ada]` (→ hangus); `[salin berhasil]`; `[token kosong / belum ada]`; `[sidik hash cocok]`; `[token kedaluwarsa]`; `[paket Tryout]` (→ tanpa token); `[pengerjaan berjalan ada]` (→ lanjut tanpa token).
- **Garpu/Gabung:** tidak ada.
- **Catatan:** Tidak ada aksi cabut token tersendiri; tidak ada token per peserta; tidak ada riwayat token; tidak ada perpanjangan (perpanjangan = generate ulang). Token tidak tercantum di kartu peserta.

### D-ACT-10 — Koreksi Esai / Manual (UC10) — UTAMA

- **Dokumentasi diagram:** `UC10 — Koreksi Esai / Manual. Aktor: Guru, Admin, Sistem.`
- **Partisi:** Guru/Admin | Sistem.
- **Simpul awal:** `Mulai koreksi`. **Simpul akhir:** `Jawaban terkoreksi`; `Koreksi ditolak`.
- **Aksi:** Buka daftar koreksi satu jadwal → Tampilkan jawaban menunggu koreksi (sudah diisi, belum dikoreksi, bertipe manual: uraian/esai, unggah berkas, praktik singkat; memuat identitas siswa, status pengerjaan, pertanyaan, bobot, jawaban, tanda ragu) → Pilih satu jawaban → Beri skor 0–100 → Periksa skor sah → Hitung poin (skor/100 × bobot) → Tandai jawaban sudah dikoreksi → Catat jejak audit → [bila pengerjaan sudah selesai] Hitung ulang nilai akhir → Selaraskan nilai ke rekap (best-effort; gagal selaras tidak menggagalkan koreksi; status final di rekap tidak diturunkan) → Siarkan pembaruan ke pemantau.
- **Keputusan + guard:** `[peran Guru/Admin]`; `[jadwal ditemukan]`; `[skor 0–100]`; `[jawaban belum dikoreksi + terisi + manual]` (yang lain tidak tampil di daftar); `[pengerjaan sudah selesai]` (→ hitung ulang seketika) / `[pengerjaan masih berjalan]` (→ nilai akhir menyusul saat pengumpulan).
- **Garpu/Gabung:** tidak ada.
- **Catatan:** Yang dikoreksi adalah jawaban per soal, bukan nilai keseluruhan siswa. Soal manual belum dijawab tidak menunggu koreksi. Layanan koreksi tersedia di sistem; tidak ada halaman tampilan khusus koreksi di aplikasi — diagram menggambarkan alur layanan ini dengan simpul Catatan yang menyatakan fakta tersebut (bukan dengan menghilangkan alur).

### D-ACT-11 — Rekap Nilai & Berita Acara (UC11) — UTAMA

- **Dokumentasi diagram:** `UC11 — Rekap Nilai & Berita Acara. Aktor: Staf berizin, Siswa (nilai sendiri), Sistem.`
- **Partisi:** Staf | Sistem | Siswa.
- **Simpul awal:** `Mulai rekap`. **Simpul akhir:** `Rekap ditampilkan`; `Nilai difinalisasi`; `Remedial tersimpan`; `CSV terunduh`; `Berita acara tampil`.
- **Aksi (rekap):** Pilih jadwal → Tampilkan daftar nilai per siswa (NISN, nama, kelas, dijawab, nilai, predikat, status Final/Menunggu koreksi) → Tampilkan statistik (peserta, rata-rata 2 desimal, tertinggi, terendah, lulus ≥ 70, menunggu koreksi) → [cabang siswa] Tampilkan hanya nilai sendiri bila final DAN saklar tampil hasil aktif (bila tidak → daftar kosong).
- **Aksi (finalisasi):** Tekan Finalisasi → Ubah SEMUA nilai jadwal menjadi final dalam satu transaksi + isi predikat tiap baris → Catat jejak audit (jumlah yang difinalisasi) → Tampilkan jumlah yang difinalisasi.
- **Aksi (remedial):** Isi nilai remedial 0–100 pada satu baris siswa → Periksa nilai 0–100 + nilai akhir siswa ada → Catat baris remedial → Perbarui nilai akhir = tertinggi (nilai awal, nilai remedial) + predikat → Catat jejak audit (awal, remedial, akhir).
- **Aksi (ekspor):** Tekan Export CSV → Tulis berkas (NISN, nama, kelas, nilai 2 desimal, predikat, status; nama berkas berbasis jadwal + tanggal) → Catat jejak audit ekspor (jumlah baris) → Unduh berkas.
- **Aksi (berita acara):** Buka berita acara satu jadwal → Tampilkan dokumen (mapel, kelas, tanggal, jam, peserta, hadir = selesai ATAU diblokir, tidak hadir = pengerjaan − hadir, rata-rata, lulus, catatan, kolom tanda tangan Proktor/Pengawas/Kepala Sekolah, stempel waktu) → Cetak lewat tombol Print.
- **Keputusan + guard:** `[jadwal dipilih]`; `[izin lihat rekap / ekspor / finalisasi / remedial]` (per aksi); `[pemanggil siswa]` (→ hanya nilai sendiri + final + tampil hasil aktif); `[nilai remedial 0–100]`; `[nilai akhir siswa ada]`; `[jadwal tanpa mapel]` (→ info minimal); `[kelas jadwal kosong]` (→ gabungkan nama kelas peserta).
- **Garpu/Gabung:** tidak ada (finalisasi, remedial, ekspor, berita acara adalah cabang-cabang alternatif dari halaman rekap, bukan paralel).
- **Catatan:** "Dijawab" hanya menghitung jawaban yang benar-benar diisi; baris jawaban kosong dari tanda ragu saja TIDAK terhitung dijawab. Catatan berita acara selalu kosong dari sistem (tidak ada isian catatan). Remedial adalah entri nilai manual, BUKAN ujian ulang.

### D-ACT-12 — Analisis Butir Soal (UC12) — LAMPIRAN

- **Dokumentasi diagram:** `UC12 — Analisis Butir Soal. Aktor: Staf berizin analisis, Proktor (tanpa teks pertanyaan), Sistem.`
- **Partisi:** Staf | Sistem.
- **Simpul awal:** `Mulai analisis butir`. **Simpul akhir:** `Analisis ditampilkan`; `CSV terunduh`; `Akses ditolak`.
- **Aksi:** Pilih jadwal → Hitung on-demand (tanpa penyimpanan; populasi = pengerjaan selesai ATAU bernilai akhir final; berjalan/diblokir TIDAK dihitung) → Tampilkan ringkasan (peserta, soal, metode pengelompokan, penanda sampel kecil, catatan) → Hitung per soal (wilayah perluasan: kesukaran P = benar/penjawab + label; daya beda D = (benar atas − benar bawah)/n + interpretasi; pengecoh per opsi) → Tampilkan tabel per soal (urutan, pertanyaan, tipe, bobot, kompetensi, penjawab, benar, tidak dijawab, belum dikoreksi, kesukaran + label, daya beda + interpretasi, pengecoh) → Ekspor CSV (penyembunyian proktor berlaku juga di ekspor) / Muat ulang.
- **Keputusan + guard:** `[izin analisis]`; `[jadwal ditemukan]`; `[pemanggil Proktor]` (→ teks pertanyaan dikosongkan, statistik tetap tampil); `[peserta ≥ 10]` (→ n = 27%) / `[peserta < 10]` (→ median split 50% + penanda sampel kecil) / `[peserta < 2]` (→ daya beda tidak dapat dihitung); `[soal tanpa penjawab ternilai]` (→ kesukaran kosong, label "Belum ada jawaban"); `[soal berbasis opsi]` (→ pengecoh berlaku; bila tidak → tidak berlaku); `[jawaban benar]` (otomatis → poin penuh; manual → poin ≥ 0,6 × bobot; tidak dijawab / belum dikoreksi → bukan benar).
- **Garpu/Gabung:** tidak ada. Hitungan per soal dimodelkan sebagai wilayah perluasan.
- **Catatan:** Jawaban soal di luar paket diabaikan. Jawaban manual belum dikoreksi dikeluarkan dari kesukaran/daya beda, tetapi dihitung pada kolom "belum dikoreksi".

### D-ACT-13 — Nilai per Kompetensi (UC13) — LAMPIRAN

- **Dokumentasi diagram:** `UC13 — Nilai per Kompetensi. Aktor: Staf berizin rekap, Siswa (nilai sendiri), Sistem.`
- **Partisi:** Staf | Sistem.
- **Simpul awal:** `Mulai nilai per kompetensi`. **Simpul akhir:** `Nilai ditampilkan`; `CSV terunduh`; `Akses ditolak`.
- **Aksi:** Pilih jadwal → Hitung on-demand (populasi sama dengan UC12: pengerjaan selesai ATAU bernilai akhir final) → Kelompokkan soal menurut kompetensi (tanpa kompetensi → "Tanpa Kompetensi", urutan terakhir; kelompok urut kode kompetensi) → Hitung nilai per siswa per kompetensi (wilayah perluasan: poin ternilai / penyebut × 100, 2 desimal) → Tampilkan per siswa (NISN, nama, kelas, nilai per kompetensi + soal/dijawab/belum dikoreksi/penanda, rata-rata siswa) → Tampilkan rata-rata kelas per kompetensi + rata-rata kelas keseluruhan → Ekspor CSV / Muat ulang.
- **Keputusan + guard:** `[izin rekap]`; `[jadwal ditemukan]`; `[pemanggil Siswa]` (→ hanya pengerjaannya sendiri); `[siswa tidak menjawab satu pun soal kompetensi]` (→ nilai kosong + "tidak ada soal", BUKAN 0); `[semua dijawab belum dikoreksi]` (→ nilai kosong + "belum dikoreksi"); `[sebagian belum dikoreksi]` (→ nilai dari yang ternilai + penanda).
- **Garpu/Gabung:** tidak ada.
- **Catatan:** Soal tidak dijawab → poin 0, bobot TETAP masuk penyebut. Rata-rata siswa dihitung atas semua soal (bukan rata-rata nilai per kompetensi). Rata-rata kelas hanya menghitung siswa bernilai (kosong tidak dihitung). Siswa yang mengekspor hanya mendapat barisnya sendiri.

### D-ACT-14 — Kelola UKK (UC14) — LAMPIRAN

- **Dokumentasi diagram:** `UC14 — Kelola UKK. Aktor: Staf berizin UKK, Siswa (berkas + sertifikat sendiri), Sistem.`
- **Partisi:** Staf | Sistem | Siswa.
- **Simpul awal:** `Mulai kelola UKK`. **Simpul akhir:** `UKK tersimpan`; `Nilai tersimpan`; `Berkas tercatat`; `Sertifikat diterbitkan`; `UKK terhapus`; `Aksi ditolak`.
- **Aksi (daftar + buat):** Buka daftar UKK (saring status) → Buat UKK (nama wajib, mapel, kelas, jurusan, tanggal, penguji internal/eksternal, status awal draf) → Validasi (rujukan ada, status draf/berjalan/selesai, tanggal berformat benar) → Simpan.
- **Aksi (detail):** Buka detail UKK (identitas, penguji, peserta dinilai + lulus, tabel nilai, daftar berkas) → Ubah status (kirim ulang seluruh data + status baru).
- **Aksi (input nilai):** Buka Input Nilai → Pilih siswa kelas UKK (bila sudah punya nilai → form terisi, mode ubah) → Isi komponen 0–100, boleh kosong (teori, proses, produk, sikap, catatan) → Lihat pratinjau hitungan (tampilan saja) → Simpan → Sistem hitung ulang resmi (praktik = proses×0,3 + produk×0,5 + sikap×0,2, kosong bila ada komponen rubrik belum diisi; akhir = teori×0,4 + praktik×0,6, kosong bila belum lengkap; lulus = akhir ≥ 70; predikat dari akhir) → Tampilkan nilai turunan server.
- **Aksi (berkas praktik):** Unggah berkas (pilih siswa pemilik + berkas) → Validasi (siswa ada, berkas ada, ≤ 10 MB, ekstensi pdf/doc/docx/jpg/jpeg/png/zip; siswa hanya untuk diri sendiri) → Bersihkan nama + awalan unik → Simpan per UKK per siswa → Catat → Daftar berkas (staf: semua; siswa: milik sendiri) → Unduh berkas (izin diperiksa; unduhan dicatat di jejak audit).
- **Aksi (sertifikat):** Tekan Sertifikat pada baris siswa lulus → Periksa (siswa ada; sertifikat sudah ada → kembalikan apa adanya, idempotent; tidak lulus / belum dinilai → tolak) → Terbitkan nomor `UKK/<tahun>/<kode-jurusan>/<urut 4 digit>` (tahun dari tanggal UKK atau tahun berjalan; urut per jurusan + tahun; bentrok → coba lagi) → Catat jejak audit → Tampilkan sertifikat (nomor, nama, NISN, kompetensi keahlian, LULUS, predikat, nilai teori/praktik/akhir, penguji, tanggal terbit + tanggal UKK, tombol Print; ditegaskan sebagai data/draf sistem sekolah, bukan penerbitan resmi).
- **Aksi (hapus):** Minta hapus UKK (konfirmasi) → Periksa tidak ada nilai/berkas/sertifikat → Hapus.
- **Keputusan + guard:** `[izin UKK sesuai aksi]`; `[UKK ditemukan]`; `[nama terisi + rujukan ada + status sah]`; `[komponen nilai 0–100]`; `[ubah nilai saat UKK selesai]` (→ hanya Admin, dicatat sebagai override); `[siswa mengubah nilai]` (→ ditolak); `[berkas ≤ 10 MB + ekstensi diizinkan]`; `[siswa mengunggah/melihat/mengunduh milik orang lain]` (→ ditolak / tidak ditemukan); `[siswa lulus]`; `[sertifikat sudah ada]` (→ kembalikan apa adanya); `[UKK punya turunan]` (→ hapus ditolak).
- **Garpu/Gabung:** tidak ada.
- **Catatan:** Tidak ada ujian UKK daring di alur ini — UKK dicatat sebagai kegiatan + nilai. Nilai teori diisi manual; tidak ada penarikan otomatis dari pengerjaan ujian.

### D-ACT-15 — Kelola Kegiatan, Ruangan & Kartu Peserta (UC15) — LAMPIRAN

- **Dokumentasi diagram:** `UC15 — Kelola Kegiatan, Ruangan & Kartu Peserta. Aktor: Admin, Guru, Proktor, Siswa (kartu sendiri), Sistem.`
- **Partisi:** Staf | Sistem | Siswa.
- **Simpul awal:** `Mulai kelola kegiatan/ruangan/kartu`. **Simpul akhir:** `Kegiatan/ruangan tersimpan`; `Kegiatan/ruangan terhapus`; `Penugasan tersimpan`; `PDF terunduh`; `Aksi ditolak`.
- **Aksi (kegiatan):** Buka halaman Kegiatan → Daftar (cari kode/nama, saring jenis) → Tambah/Ubah (kode, nama, jenis STS/SAS/USBK/UKK/Tryout/Lainnya bawaan Lainnya, tanggal mulai/selesai opsional, deskripsi opsional) → Validasi (kode + nama wajib, jenis sah, tanggal sah, selesai ≥ mulai, kode belum dipakai kegiatan lain) → Simpan → Hapus (konfirmasi → periksa tidak dipakai jadwal → hapus).
- **Aksi (ruangan):** Buka halaman Ruangan → Daftar (cari kode/nama/lokasi) → Tambah/Ubah (kode, nama, kapasitas kursi, lokasi/gedung opsional, keterangan opsional) → Validasi (kode + nama wajib, kapasitas bilangan bulat > 0, kode belum dipakai ruangan lain) → Simpan → Hapus (konfirmasi → periksa tidak dipakai jadwal → hapus). Cabang ubah kapasitas: kapasitas baru < siswa aktif kelas jadwal pemakai → tolak (menyebut jadwalnya).
- **Aksi (penugasan per jadwal):** Buka seksi Penugasan di halaman Kartu Peserta → Pilih jadwal → Baca penugasan tersimpan (isi awal pilihan) → Pilih kegiatan dan/atau ruangan ("Tanpa …" = lepas) → Simpan Penugasan → Periksa (jadwal ada; kegiatan/ruangan dipilih ada; kapasitas ruangan ≥ siswa AKTIF kelas jadwal) → Simpan / lepas penugasan.
- **Aksi (kartu staf):** Buka halaman Kartu Peserta → Daftar siswa (saring kegiatan/kelas, cari nama/NISN, penomoran 20/halaman) → [satu siswa] Unduh PDF → [gabungan] Atur saringan → Unduh Semua (PDF, satu halaman per siswa) → Periksa (saringan ada; ada kartu untuk saringan; jumlah siswa ≤ 500).
- **Aksi (kartu siswa):** Buka "Kartu Saya" → Tampilkan pratinjau (identitas dari sesi + tabel jadwal HARI INI, pratinjau saja) → Unduh Kartu (PDF lengkap: identitas sekali + tabel SEMUA jadwal kelasnya).
- **Keputusan + guard:** `[kode kegiatan/ruangan dipakai]`; `[tanggal selesai ≥ tanggal mulai]`; `[kapasitas bilangan bulat > 0]`; `[kegiatan/ruangan dipakai jadwal]` (→ hapus ditolak, lepas dulu); `[kapasitas baru ≥ siswa kelas jadwal pemakai]`; `[penugasan: kegiatan/ruangan ditemukan]`; `[kapasitas ruangan ≥ siswa aktif kelas jadwal]`; `[kartu gabungan > 500 siswa]` (→ tolak, persempit saringan); `[kartu gabungan tanpa hasil]` (→ tidak ditemukan); `[siswa meminta kartu siswa lain / siswa tidak aktif]` (→ tidak ditemukan).
- **Garpu/Gabung:** tidak ada.
- **Catatan:** Model 1-KARTU: satu kartu per siswa, identitas sekali + tabel semua jadwal kelasnya (termasuk Draf, urut tanggal + jam); nomor peserta unik PER JADWAL (`<kode kegiatan atau US>/<tahun tanggal jadwal>/<urut 3 digit>`, urut nama A–Z di antara siswa aktif kelas itu). Kartu TIDAK berisi token. Tidak ada pemeriksaan bentrok ruangan, kustomisasi desain kartu, foto, atau kode QR.

### D-ACT-16 — Cadangkan & Pulihkan (UC16) — LAMPIRAN

- **Dokumentasi diagram:** `UC16 — Cadangkan & Pulihkan. Aktor: Staf berizin backup, Sistem.`
- **Partisi:** Staf | Sistem.
- **Simpul awal:** `Mulai cadangkan/pulihkan`. **Simpul akhir:** `Backup tercatat`; `Berkas terunduh`; `Pulih berhasil`; `Pulih dibatalkan/ditolak`; `Backup terhapus`.
- **Aksi (daftar):** Buka halaman Backup & Restore → Tampilkan daftar backup (terbaru dulu, penomoran bawaan 50, maks 200: nama berkas, jenis Manual/Pre-restore, ukuran, waktu, pembuat, keterangan) → Pindah halaman.
- **Aksi (buat manual):** Isi keterangan opsional (maks 500 karakter) → Tekan Buat Backup → Buat salinan penuh (nama unik berbasis waktu + akhiran acak, tidak menimpa) → Catat metadata (nama, ukuran, jenis manual, pembuat, keterangan, waktu) → Catat jejak audit → Tampilkan nama berkas + ukuran.
- **Aksi (unduh):** Tekan Unduh → Periksa izin (izin buat ATAU pulihkan) + metadata + nama berkas aman (tolak pemisah jalur) → Alirkan berkas sebagai unduhan → Catat jejak audit unduhan.
- **Aksi (pulihkan):** Tekan Restore → Dialog konfirmasi (peringatan keras: menimpa SEMUA data; backup pengaman dibuat otomatis; perubahan sesudah backup hilang; ditolak bila ada jadwal aktif/jeda atau peserta mengerjakan) → Konfirmasi → Periksa izin pulihkan + metadata + berkas + perkakas pulih tersedia → Gerbang ujian: periksa jadwal aktif/jeda ATAU pengerjaan berjalan (bila ada → tolak, catat penolakan di jejak audit, pesan menyebut jumlah penghalang) → Buat backup pengaman pre-restore + catat metadata + jejak auditnya (bila gagal → batalkan restore) → Pulihkan isi berkas ke basis data (bersihkan + buat ulang semua isi) → Tulis kembali metadata backup pengaman + metadata backup target + jejak audit pembuatan pengaman + jejak audit restore → Tampilkan berkas sumber + berkas pengaman.
- **Aksi (hapus):** Tekan Hapus → Dialog konfirmasi (berkas + metadata dihapus permanen; pre-restore < 7 hari tidak boleh dihapus) → Periksa izin pulihkan → [pre-restore < 7 hari → tolak + catat penolakan] → Hapus berkas + metadata → Catat jejak audit penghapusan.
- **Keputusan + guard:** `[izin buat backup]` (daftar + buat; daftar juga boleh dengan izin pulihkan); `[izin pulihkan]` (restore + hapus); `[keterangan ≤ 500 karakter]`; `[ID backup sah / ditemukan]`; `[metadata ada, berkas hilang]` (→ unduh/restore ditolak dengan pesan yang membedakannya); `[perkakas salin/pulih tersedia]`; `[ada jadwal aktif/jeda atau pengerjaan berjalan]` (→ restore ditolak); `[backup pengaman berhasil dibuat]` (→ bila tidak, restore dibatalkan); `[pulih gagal / batas waktu 10 menit terlampaui]` (→ pesan menunjuk berkas pengaman); `[pre-restore berusia < 7 hari]` (→ hapus ditolak).
- **Garpu/Gabung:** tidak ada. Serialisasi operasi (pulih tidak bersamaan dengan salin/pulih lain; hapus tidak balapan dengan pulih) dicatat sebagai batasan pada simpul aksi, bukan fork.
- **Catatan:** Hanya dua jenis backup: manual dan pre-restore. Nama berkas selalu dibuat sistem; pengguna tidak pernah menentukan nama berkas. Kata sandi basis data tidak pernah ditulis ke log/respons. Tidak ada backup otomatis harian di dalam aplikasi (hanya lewat skrip cron di luar aplikasi). Salinan gagal → berkas parsial dihapus.

### D-ACT-17 — Kelola Pengaturan, Audit & Notifikasi (UC17) — LAMPIRAN

- **Dokumentasi diagram:** `UC17 — Kelola Pengaturan, Audit & Notifikasi. Aktor: Admin / staf berizin, Sistem.`
- **Partisi:** Staf | Sistem | Pengguna (penerima notifikasi).
- **Simpul awal:** `Mulai kelola pengaturan/audit/notifikasi`. **Simpul akhir:** `Pengaturan tersimpan`; `Notifikasi ditandai dibaca`; `Pengecualian SoD tercatat`; `Aksi ditolak`.
- **Aksi (pengaturan):** Buka halaman Pengaturan → Periksa izin baca → Tampilkan nilai saat ini (kunci belum disimpan → nilai bawaan) → Ubah nilai (saklar tampil hasil; angka masa berlaku token / toleransi telat / maks pelanggaran / durasi bawaan; nama sekolah; logo sebagai berkas unggahan; KKTP per mapel sebagai teks JSON; mode remedial baca-saja) → Validasi KKTP di tampilan (harus objek JSON) → Tekan Simpan → Periksa izin ubah → Terima perubahan (terbungkus maupun datar; bila tidak ada pengaturan → tolak) → Validasi tiap kunci dalam satu transaksi (kunci tidak dikenal → tolak; nilai tidak sesuai tipe/rentang → tolak; satu kunci tidak valid → seluruh penyimpanan batal) → Simpan semua sekaligus → Catat jejak audit (kunci yang diubah) → Tampilkan "berhasil disimpan" + muat ulang nilai.
- **Aksi (jejak audit — layanan belakang, tanpa halaman daftar):** Setiap operasi penting menulis satu baris jejak audit (pelaku, aksi, entitas, pengenal, rincian, IP, waktu) → Pembaca berizin menyaring (aksi, entitas, pelaku; urut terbaru dulu; penomoran bawaan 50, maks 200) → Baca hasil. Jejak audit tidak pernah diubah/dihapus (hanya tambah + baca).
- **Aksi (notifikasi):** Sistem membuat notifikasi pada kejadian yang membutuhkan (pengecualian SoD → ke semua admin aktif kecuali pelaku) → Bel notifikasi polling daftar milik sendiri tiap 60 detik → Tampilkan lencana belum dibaca (maks tampil 99) + menu turun (judul, isi, waktu, tombol tandai dibaca) → Tekan tandai dibaca → Tandai notifikasi itu (hanya bila milik sendiri) → Catat jejak audit tandai-dibaca → Muat ulang daftar. Bila layanan notifikasi galat → bel menampilkan 0 dan daftar kosong.
- **Aksi (pengecualian SoD darurat — layanan belakang, tanpa formulir tampilan):** Admin berizin mengisi aksi yang dikecualikan + entitas + pengenal + alasan tertulis → Periksa (login; izin khusus SoD; aksi + entitas wajib; alasan ≥ 20 karakter setelah dipangkas) → Simpan pengecualian → Catat jejak audit berpenanda khusus SoD → Kirim notifikasi ke semua admin aktif kecuali pelaku.
- **Keputusan + guard:** `[izin baca/ubah pengaturan]`; `[KKTP objek JSON]`; `[ada pengaturan dikirim]`; `[kunci dikenal]`; `[nilai sesuai tipe/rentang kunci]` (masa berlaku token 60–86400; toleransi telat 0–120; maks pelanggaran 1–10; durasi bawaan 5–600; mode remedial hanya "tertinggi"); `[izin baca audit]`; `[notifikasi milik sendiri]` (→ bila bukan → "tidak ditemukan"); `[izin pengecualian SoD]`; `[alasan ≥ 20 karakter]`.
- **Garpu/Gabung:** tidak ada.
- **Catatan:** KKTP per mapel adalah objek deskripsi/rubrik/interval per mapel/TP — BUKAN satu angka KKM untuk semua mapel. Mode remedial selalu "tertinggi". Pengecualian SoD TIDAK PERNAH dipakai untuk mengubah nilai tanpa jejak atau menghapus jejak audit. Jejak audit dan formulir SoD tidak memiliki halaman tampilan — diagram menggambarkan alur layanan dengan simpul Catatan yang menyatakan fakta tersebut.

### D-ACT-18 — Usul & Proses Buka Blokir (UC18) — UTAMA

- **Dokumentasi diagram:** `UC18 — Usul & Proses Buka Blokir. Aktor: Siswa yang diblokir, Pengawas (mengusulkan), Proktor/Admin (mengeksekusi), Sistem.`
- **Partisi:** Pengawas | Proktor/Admin | Sistem | Siswa.
- **Simpul awal:** `Mulai usulan / buka blokir`. **Simpul akhir:** `Usulan menunggu tercatat`; `Peserta dibuka blokir (lanjut sisa waktu)`; `Buka blokir ditolak`; `Usulan ditolak`.
- **Konteks (simpul Catatan di awal diagram, bukan aksi):** Pengerjaan diblokir otomatis saat pelanggaran mencapai batas maks (bawaan 2; pertama = peringatan). Saat diblokir, sisa waktu dibekukan; jawaban + tanda ragu TIDAK dihapus.
- **Aksi (usulkan):** Ajukan usulan untuk satu pengerjaan diblokir + alasan → Periksa peran (pengawas/proktor/admin) → Periksa isian (pengenal pengerjaan + alasan wajib) → Periksa pengerjaan (ditemukan; milik jadwal yang dimaksud; berstatus diblokir) → Periksa usulan menunggu (satu pengerjaan hanya boleh punya satu usulan menunggu) → Simpan usulan berstatus menunggu (pencatat + alasan) → Catat jejak audit usulan → Siarkan perubahan status ke pemantau.
- **Aksi (eksekusi):** Proktor/admin melihat baris peserta diblokir di pemantau → Isi alasan buka blokir → Tekan Buka Blokir → Periksa peran (hanya proktor/admin; pengawas ditolak) → Periksa isian + pengerjaan (sama seperti usulan) → Hitung sisa waktu lanjutan (sisa beku bila ada; bila tidak → jam selesai jadwal − sekarang) → Periksa sisa > 0 → Dalam satu transaksi: status kembali berjalan + tenggat baru (sekarang + sisa) + kosongkan sisa beku (jawaban TIDAK direset) → Tutup usulan (usulan menunggu terbaru → disetujui + pemroses + waktu proses; bila tidak ada usulan → buat baris usulan disetujui sebagai jejak) → Catat jejak audit buka blokir (alasan + sisa detik) → Siarkan perubahan status → Tampilkan "Peserta di-unblock — ujian dilanjutkan dari sisa waktu" + kosongkan kolom alasan.
- **Aksi (siswa lanjut):** Siswa mengetahui status lewat pemeriksaan status berkala (30 detik) / permintaan berikutnya (tidak ada pemberitahuan dorong) → Lanjutkan dari soal terakhir (jawaban + ragu utuh) dengan tenggat baru.
- **Keputusan + guard:** `[peran pengusul sah]`; `[alasan terisi]`; `[pengerjaan ditemukan + milik jadwal ini]`; `[pengerjaan diblokir]`; `[sudah ada usulan menunggu]` (→ usulan ditolak); `[peran eksekutor = proktor/admin]`; `[sisa waktu > 0]` (→ bila ≤ 0: "Ujian sudah berakhir… Unblock tidak bisa dilakukan.").
- **Garpu/Gabung:** tidak ada (usulkan → eksekusi → lanjut adalah fase sekuensial; eksekusi dapat terjadi tanpa usulan).
- **Catatan:** Tidak ada status usulan "ditolak" — usulan hanya menunggu atau disetujui. Pembuatan usulan tidak memiliki tombol/formulir di tampilan (hanya layanan belakang); diagram menggambarkan alur layanan dengan simpul Catatan yang menyatakan fakta tersebut. Pemisahan tugas: pengawas mengajukan, proktor/admin mengeksekusi.

### D-ACT-19 — Lihat Hasil & Pembahasan (UC19) — LAMPIRAN

- **Dokumentasi diagram:** `UC19 — Lihat Hasil & Pembahasan. Aktor: Siswa, Sistem.`
- **Partisi:** Siswa | Sistem.
- **Simpul awal:** `Mulai lihat hasil`. **Simpul akhir:** `Hasil lengkap ditampilkan`; `Hasil terbatas ditampilkan`; `Hasil belum tersedia`.
- **Aksi:** Buka halaman hasil satu pengerjaan (dari kartu hasil pengumpulan / beranda) → Periksa pengerjaan milik siswa sendiri + sudah selesai (bila belum selesai → hasil belum tersedia; bila milik orang lain → tidak ditemukan) → Tampilkan kartu Nilai Akhir (nilai bila boleh tampil; bila tidak → "Nilai disembunyikan sampai koreksi final"; bila masih ada menunggu koreksi → "Masih ada soal uraian yang menunggu koreksi guru"; Tryout → "Tryout — pembahasan langsung tampil") → Tampilkan per soal (wilayah perluasan: nomor, pertanyaan, jawaban siswa, penanda ragu bila ada, status soal; bila boleh tampil → tambah poin, ringkasan kunci benar, pembahasan guru; bila tidak → hanya pertanyaan + jawaban siswa) → Soal tanpa jawaban tetap ditampilkan (status salah bila boleh tampil; disembunyikan bila tidak).
- **Keputusan + guard:** `[pengerjaan milik sendiri]`; `[pengerjaan selesai]`; `[hasil boleh tampil]` (saklar tampil hasil paket AKTIF DAN pengaturan tampil hasil AKTIF) / `[paket Tryout]` (→ selalu tampil lengkap); `[masih ada soal menunggu koreksi]`.
- **Status per soal (label pada keluaran, bukan cabang):** Benar (poin = bobot penuh); Sebagian benar (0 < poin < bobot); Salah (poin 0, termasuk tidak dijawab bila boleh tampil); Menunggu koreksi (poin kosong); Disembunyikan (tidak boleh tampil).
- **Garpu/Gabung:** tidak ada. Tampilan per soal dimodelkan sebagai wilayah perluasan.
- **Catatan:** Kunci yang ditampilkan adalah RINGKASAN kunci benar per tipe, bukan kunci mentah. Kunci + pembahasan tidak pernah dikirim selama ujian. Jawaban teks tampil literal (tidak ditafsirkan sebagai markup). Tidak ada halaman hasil untuk staf.

---

## 3. Pemetaan dari Diagram Activity lama ke diagram baru

Diagram lama **diganti** (tidak diedit). Penulis diagram membuat diagram baru
sesuai §2 dan memberi nama sesuai kolom "Kode diagram".

| Diagram lama | Nasib | Diganti / dipecah menjadi |
|---|---|---|
| **2a Pelaksanaan Ujian** | DIGANTI, dipecah menjadi 5 diagram | **D-ACT-08 Ikuti Ujian (UC08)** — inti pelaksanaan (mulai, kerjakan, autosave, ragu, timer, kumpul, anti-curang). **D-ACT-09 Validasi Token (UC09)** — gerbang mulai yang dulu tercampur di 2a. **D-ACT-07 Pantau Ujian Langsung (UC07)** — sisi pemantau yang dulu tercampur di 2a. **D-ACT-18 Usul & Proses Buka Blokir (UC18)** — kelanjutan ujian yang diblokir. **D-ACT-19 Lihat Hasil & Pembahasan (UC19)** — hasil sesudah pengumpulan. Penilaian koreksi manual dan rekap TIDAK termasuk 2a; keduanya sudah tercakup diagram baru D-ACT-10 dan D-ACT-11. |
| **2b Impor Soal** | DIGANTI, satu-ke-satu | **D-ACT-04 Impor Soal (UC04)** — cakupan diperbarui dari telusur baru: Excel + Word, pratinjau dry-run, validasi ulang saat simpan, mode tolak-semua / lewati-error, perlakuan duplikat, katalog temuan, soal hasil impor selalu Draf. |
| (tidak ada) | BARU | 14 diagram sisanya (D-ACT-01, 02, 03, 05, 06, 10, 11, 12, 13, 14, 15, 16, 17) tidak memiliki padanan diagram lama; semuanya digambar baru dari telusur. |

**Aturan pemetaan:** bila satu alur lama tersebar ke beberapa diagram baru,
diagram baru TIDAK saling menduplikasi alur tersebut — diagram pemanggil memakai
simpul pemanggil perilaku ke diagram pemilik (lihat §1.1 dan §4 butir 5).

---

## 4. Aturan penggambaran (mengikat penulis diagram)

1. **Judul + dokumentasi:** nama diagram = `D-ACT-NN — <Judul> (UCxx)`; dokumentasi
   diagram diawali kode UC dan nama UC persis dari diagram Use Case (lihat §2).
2. **Bahasa:** Bahasa Indonesia penuh. Tanpa nama berkas, alamat layanan,
   perintah basis data, pustaka, nama tabel/kolom, nama fungsi, atau istilah
   implementasi lain. Istilah yang boleh dipakai: aktor, aksi, kondisi, guard,
   hasil akhir, aturan bisnis — sesuai telusur.
3. **Keterhubungan:** setiap simpul aksi/keputusan/garpu memiliki tepat satu
   alur masuk dan alur keluar yang jelas (kecuali simpul awal/akhir dan simpul
   pemanggil yang kembali ke pemanggilnya); tidak ada elemen melayang; tidak ada
   alur yang berhenti di tengah tanpa simpul akhir.
4. **Decision/merge:** setiap decision memiliki merge pasangannya; setiap cabang
   decision berlabel guard `[…]` dalam bahasa desain; cabang yang tidak ber-guard
   berarti selalu diambil bila cabang lain tidak terpenuhi — tuliskan guard
   `[selebihnya]` agar eksplisit, jangan biarkan kosong.
5. **Fork/join:** fork dan join selalu berpasangan dan seimbang (jumlah cabang
   keluar fork = jumlah cabang masuk join). Fork HANYA dipakai di D-ACT-07 dan
   D-ACT-08 (lihat §2); pengulangan per baris / per soal / per pengerjaan /
   per siswa memakai wilayah perluasan (expansion region), bukan fork.
6. **Simpul pemanggil (call behavior):** dipakai untuk pause/resume (D-ACT-07 →
   D-ACT-06), validasi token (D-ACT-08 → D-ACT-09), usul/eksekusi buka blokir
   (D-ACT-07 → D-ACT-18). Simpul pemanggil diberi nama alur yang dipanggil +
   kode diagram tujuan, dan kembali ke diagram pemanggil sesudah selesai.
7. **Partisi (swimlane):** pakai partisi sesuai §2 per diagram. Aktor yang hanya
   menjadi sumber peristiwa (siswa pada D-ACT-07) tidak mendapat partisi pelaku;
   bila perlu, partisinya dinamai sumber peristiwa, bukan pelaku.
8. **Simpul akhir ganda:** diizinkan — satu simpul akhir per hasil akhir yang
   berbeda (tersimpan / ditolak / selesai / diblokir), masing-masing berlabel.
   Simpul akhir yang sama tidak dipakai untuk dua hasil yang maknanya berbeda.
9. **Simpul Catatan (Note):** dipakai untuk (a) konteks yang bukan aksi
   (mis. konteks pemblokiran di D-ACT-18), (b) fakta "tidak ada tampilan"
   (koreksi UC10, jejak audit + SoD UC17, usulan UC18), (c) batasan yang bukan
   cabang (serialisasi operasi UC16). Catatan TIDAK menggantikan aksi.
10. **Tampilan vs sistem:** bila tampilan dan sistem melakukan pemeriksaan yang
    sama (mis. validasi KKTP, alasan unblock kosong), gambarkan keduanya sebagai
    dua aksi berurutan (tampilan dulu, lalu sistem) — jangan digabung menjadi
    satu aksi, karena penegakan utama ada di sistem.
11. **Loop:** setiap wilayah perluasan diberi label jumlahnya ("untuk tiap
    baris", "untuk tiap soal", "untuk tiap pengerjaan berjalan") dan kondisi
    keluarnya jelas (habis diproses / ditemukan pelanggaran).
12. **Konsistensi lintas diagram:** nama aksi yang sama pada dua diagram
    (mis. "Bekukan sisa waktu") harus memiliki arti yang sama; bila artinya
    berbeda, beri nama berbeda. Guard yang sama ditulis dengan kata yang sama
    di semua diagram.

---

## 5. Hal yang perlu konfirmasi sebelum menggambar `.mdj`

1. **Urutan pemeriksaan saat mulai ujian (UC08 vs UC09)** — pertentangan
   antar-bagian telusur (lihat `ALUR-ACTIVITY-NYATA.md` § Pertentangan, butir 1).
   Rancangan ini memakai urutan Bagian B (pengerjaan diblokir/selesai ditolak
   SEBELUM token) sebagai hipotesis kerja. Penulis diagram WAJIB mengonfirmasi
   ke induk sebelum menggambar D-ACT-08/D-ACT-09; bila diputuskan urutan Bagian C,
   ubah rancangan D-ACT-09 cabang validasi (urutan decision-nya), bukan
   diagram lain.
2. **Halaman tampil-baca Kisi-kisi** — verifikasi satu kali apakah halaman
   tampil-baca Kisi-kisi ada (lihat § Pertentangan, butir 3). Bila tidak ada,
   hapus penyebutan halaman Kisi-kisi dari dokumentasi D-ACT-04.
3. **Posisi pemeriksaan kunci login sementara (D-ACT-01)** — telusur tidak
   menyatakan eksplisit di titik mana penguncian diperiksa; rancangan menaruhnya
   sesudah kelengkapan isian, sebelum verifikasi kata sandi. Konfirmasi satu
   kali; bila berbeda, geser decision-nya tanpa mengubah aksi lain.
4. **Alur layanan tanpa tampilan (UC10 koreksi, UC17 audit + SoD, UC18 usulan)** —
   rancangan menggambarkannya sebagai alur layanan + simpul Catatan. Bila induk
   memutuskan alur tanpa tampilan TIDAK digambar (hanya dicatat), hapus aksi
   layanannya dari D-ACT-10/D-ACT-17/D-ACT-18 dan sisakan Catatan; keputusan ini
   tidak mengubah diagram lain.
5. **Penandaan UTAMA/LAMPIRAN** mengikuti definisi §1. Bila induk ingin badan
   skripsi lebih ramping (mis. UC09 dan UC18 dipindah ke lampiran), cukup ubah
   penanda pada §1 dan §2 — struktur diagram tidak berubah.

---

## Lampiran — Keterlacakan

| Diagram | UC sumber | Bagian telusur di `ALUR-ACTIVITY-NYATA.md` |
|---|---|---|
| D-ACT-01 | UC01 | § UC01 |
| D-ACT-02 | UC02 | § UC02 |
| D-ACT-03 | UC03 | § UC03 |
| D-ACT-04 | UC04 | § UC04 |
| D-ACT-05 | UC05 | § UC05 |
| D-ACT-06 | UC06 | § UC06 |
| D-ACT-07 | UC07 | § UC07 (+ memanggil UC06, UC18) |
| D-ACT-08 | UC08 | § UC08 (+ memanggil UC09) |
| D-ACT-09 | UC09 | § UC09 |
| D-ACT-10 | UC10 | § UC10 |
| D-ACT-11 | UC11 | § UC11 |
| D-ACT-12 | UC12 | § UC12 |
| D-ACT-13 | UC13 | § UC13 |
| D-ACT-14 | UC14 | § UC14 |
| D-ACT-15 | UC15 | § UC15 |
| D-ACT-16 | UC16 | § UC16 |
| D-ACT-17 | UC17 | § UC17 |
| D-ACT-18 | UC18 | § UC18 |
| D-ACT-19 | UC19 | § UC19 |

**Verifikasi rancangan:** 19 diagram = 19 UC (tidak ada UC tanpa diagram, tidak
ada diagram tanpa UC); 8 UTAMA + 11 LAMPIRAN = 19; 0 penggabungan; 2 diagram
dengan fork/join (D-ACT-07, D-ACT-08); 3 relasi pemanggil (D-ACT-07 → D-ACT-06,
D-ACT-07 → D-ACT-18, D-ACT-08 → D-ACT-09); 2 diagram lama diganti (2a → 5
diagram baru, 2b → 1 diagram baru).
