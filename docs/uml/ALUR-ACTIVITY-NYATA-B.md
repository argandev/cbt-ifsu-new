# ALUR ACTIVITY NYATA — Kelompok B (Bank Soal + Paket + Jadwal + Token + Kegiatan)

> **Status:** Dokumen telusur. Disusun HANYA dari membaca kode aplikasi (bukan dari diagram Activity lama, bukan dari ingatan).
> **Tanggal telusur:** 18 Okt 2026.
> **Cakupan:** UC03, UC04, UC05, UC06, UC09, UC15.
> **Aturan penulisan:** Bahasa desain. Perilaku yang ditulis adalah perilaku yang terlihat di kode: langkah tampilan, pemeriksaan sistem, cabang, dan hasil akhir. Tidak ada nama teknis implementasi.
> **Konvensi umum yang berlaku untuk semua use case di dokumen ini:**
> - Setiap aksi ubah data menuntut pengguna sudah masuk dan memiliki izin untuk aksi itu; tanpa izin, aksi ditolak.
> - Setiap aksi ubah data yang berhasil dicatat dalam jejak audit (siapa, aksi apa, kapan).
> - Penghapusan dan penguncian selalu meminta konfirmasi dulu dan bersifat permanen (tidak ada keranjang sampah / pulihkan).
> - Pesan penolakan menjelaskan alasannya dalam bahasa pengguna.

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

## UC09 — Validasi Token (generate token, salin, cek status / kedaluwarsa, token kedua menghanguskan pertama)

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

## Catatan lintas use case (hasil telusur)

1. **Kunci bank soal adalah satu-satunya pengaman isi soal.** Kunci tidak menghentikan ubah metadata bank (kode/nama/mapel/kelas tetap dapat diubah saat bank terkunci); yang diblokir adalah mutasi soal + impor + hapus bank.
2. **Pemisahan tugas kunci bank** (pembuat tidak boleh mengunci bank sendiri) hanya berlaku untuk aksi kunci/buka kunci — bukan untuk mutasi soal (guru pemilik tetap boleh mengubah soalnya sendiri).
3. **Impor selalu menulis ulang dari berkas** pada konfirmasi; pratinjau tidak pernah dipakai sebagai dasar simpan.
4. **Soal hasil impor selalu Draf** — tidak ada impor langsung Aktif.
5. **Ubah status jadwal lewat formulir tidak membekukan timer** — hanya aksi Pause/Resume khusus yang membekukan/memulihkan timer. Ini perbedaan perilaku yang penting bagi proktor.
6. **Token tidak pernah dapat dilihat lagi** sesudah generate; satu-satunya pemulihan adalah generate token baru (yang menghanguskan token lama).
7. **Kartu peserta tidak memuat token** dalam kondisi apa pun; token hanya dari proktor saat ujian.
