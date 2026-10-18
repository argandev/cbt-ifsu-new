# ALUR ACTIVITY NYATA — KELOMPOK C (Ujian + Penilaian + Rekap + UKK + Backup + Hasil)

> **Status:** Dokumen telusur Fase 1 Kelompok C. Disusun HANYA dari membaca kode sumber aplikasi (bukan dari diagram Activity lama 2a/2b, bukan dari ingatan).
> **Tanggal telusur:** 18 Okt 2026.
> **Aturan penulisan:** Bahasa desain. Dokumen ini TIDAK menyebut nama berkas kode, alamat layanan, nama tabel/kolom, nama pustaka, atau nama fungsi. Yang ditulis adalah perilaku yang terlihat di kode: siapa berbuat apa, dalam urutan apa, dengan kondisi apa, dan apa hasil akhirnya.
> **Referensi silang:** `docs/uml/ALUR-SEQUENCE-NYATA.md` (telusur Alur A/B/C — login, ujian, pemantauan/anti-curang). Kelompok C mengulang bagian ujian/pemantauan seperlunya agar tiap use case berdiri sendiri, tanpa menyalin diagram lama.
> **Cakupan:** UC07, UC08, UC10, UC11, UC12, UC13, UC14, UC16, UC19.

---

## Konvensi yang berlaku untuk semua use case Kelompok C

- **Satu jadwal = satu paket soal untuk satu kelas pada satu tanggal dan rentang jam.** Peserta hanya dapat mengerjakan jadwal kelasnya sendiri.
- **Satu pengerjaan (attempt) per siswa per jadwal.** Pengerjaan berstatus: berjalan, diblokir, atau selesai.
- **Nilai akhir** = jumlah poin yang diperoleh dibagi jumlah bobot soal yang dinilai, dikali 100, dijepit ke rentang 0–100. Soal yang tidak dinilai (mis. survei sikap) tidak masuk penyebut.
- **Predikat nilai** (berlaku sama untuk rekap dan UKK): di bawah 70 = Belum Tuntas; 70–79 = Cukup; 80–90 = Baik; 91–100 = Sangat Baik. Lulus bila nilai akhir ≥ 70.
- **Setiap aksi penting dicatat dalam jejak audit** (siapa, kapan, apa, untuk entitas mana). Isi jawaban siswa tidak ditulis ke jejak audit.
- **Semua waktu memakai zona waktu Asia/Jakarta.**

---

## UC07 — Pantau Ujian Langsung

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

## Hal yang TIDAK ditemukan di kode (Kelompok C)

1. **UC07 — Tidak ada aksi "paksa selesai".** Pemantauan yang ada: pantau, jeda/lanjutkan jadwal, usulkan buka blokir, buka blokir. Tidak ada layanan/tombol yang memaksa satu/all peserta selesai dari pemantauan.
2. **UC10 — Tidak ada halaman tampilan koreksi.** Layanan daftar menunggu koreksi + simpan skor ada di sisi server, tetapi tidak ada halaman yang memanggilnya di kode tampilan yang dibaca. Koreksi belum dapat dilakukan lewat antarmuka aplikasi.
3. **UC08 — Tidak ada proses latar untuk finalisasi otomatis.** Finalisasi saat batas waktu lewat terjadi secara malas pada permintaan berikutnya yang menyentuh pengerjaan itu. Pengerjaan yang ditinggalkan tepat setelah batas waktu lewat tetap berstatus berjalan sampai ada permintaan berikutnya.
4. **UC08 — Tidak ada siaran ke siswa.** Satu-satunya saluran langsung adalah untuk pemantau. Siswa mengetahui jeda/blokir/selesai lewat penarikan status berkala atau permintaan berikutnya.
5. **UC08 — Tidak ada kirim ulang otomatis autosave.** Jawaban yang gagal tersimpan tidak dikirim ulang sendiri; siswa harus mengubah jawaban lagi.
6. **UC11 — Tidak ada isian catatan pada berita acara.** Catatan selalu kosong dari sistem. Tidak ada alur mengisi/menyimpan catatan kejadian berita acara.
7. **UC11 — Remedial bukan ujian ulang.** Remedial adalah entri nilai manual (nilai tertinggi menang). Tidak ada pengerjaan/token/soal remedial.
8. **UC14 — Tidak ada penarikan otomatis nilai teori UKK dari ujian.** Nilai teori UKK diisi manual; tidak ada penyalinan otomatis dari pengerjaan ujian teori.
9. **UC16 — Tidak ada backup otomatis harian di dalam aplikasi.** Backup otomatis hanya lewat skrip cron di luar aplikasi; halaman backup hanya manual.
10. **UC19 — Tidak ada halaman hasil untuk staf.** Halaman hasil hanya untuk siswa pemilik pengerjaan; staf melihat nilai lewat rekap/analisis/koreksi.
11. **UC07/UC08 — Pengawas tidak punya tombol "usulkan buka blokir" di tampilan.** Layanan usulan ada di sisi server dan daftar usulan ditampilkan di pemantauan, tetapi tidak ada tombol di tampilan yang mengirim usulan. (Proktor/Admin membuka blokir langsung dari pemantauan.)

---

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
