# Pemetaan PSM → PIM — CBT SMK v2 (UML)

Dokumen ini memetakan **setiap** perubahan nama dari level **implementasi (PSM —
spesifik Go/PostgreSQL, hasil reverse-engineering kode)** ke level **desain
(PIM — platform-independent)** pada diagram UML di `cbt-smk-v2.mdj`, agar
dapat ditelusuri dosen pembimbing. File ini ditulis bertahap per diagram;
**Worker 1 (Sequence)** sudah selesai, **Worker 2 (Class + ERD)**
sudah selesai, **Worker 3 (Use Case + Activity)** sudah selesai.

- **File:** `docs/uml/cbt-smk-v2.mdj`
- **SHA-256 sebelum PIM Sequence:** `fdb0607feb46d92c4825e5323e9251b8557f758ef73056a656e06a49652d3727`
- **SHA-256 sesudah PIM Sequence:** `1b7367ca02b4716639ce977ff0ec5a00c2bf4e44b186731ea60e4fb4b0b5eb4a`
- **Script perbaikan:** `docs/uml/fix_pim_sequence.py` (idempotent — rerun
  menghasilkan SHA identik; semua pemetaan memakai ID elemen, bukan
  pencocokan teks)
- **SHA-256 sebelum PIM Class+ERD:** `1b7367ca02b4716639ce977ff0ec5a00c2bf4e44b186731ea60e4fb4b0b5eb4a`
  (= SHA sesudah PIM Sequence; diverifikasi sebelum menulis file)
- **SHA-256 sesudah PIM Class+ERD:** `6e4ec9fb50b2277bb586c8a447a85b4caf09ea644e3c0422481c35fafe9cd859`
- **Script perbaikan:** `docs/uml/fix_pim_class_erd.py` (idempotent —
  rerun menghasilkan SHA identik; pemetaan entitas/relasi/asosiasi/operasi
  memakai ID elemen, atribut/kolom memakai (induk, nama-lama) + asersi
  ketat; backup pra-perubahan:
  `docs/uml/cbt-smk-v2.mdj.pre-pim-class-erd.bak`)
- **Validator:** `docs/uml/validate-mdj-real.py` — 0 pelanggaran; ekspektasi
  diperbarui (Lifeline 12→10, NameCompartment 93→91, Label nama 128→126,
  NoteView 2/2→4/4) tanpa melemahkan cek (3 kontrol negatif diuji: NoteView
  tanpa teks, nama pesan ganda, label kompartemen ≠ nama model — ketiganya
  tetap GAGAL dengan benar).
- **SHA-256 sebelum PIM Use Case+Activity:**
  `6e4ec9fb50b2277bb586c8a447a85b4caf09ea644e3c0422481c35fafe9cd859`
  (= SHA sesudah PIM Class+ERD; diverifikasi sebelum menulis file)
- **SHA-256 sesudah PIM Use Case+Activity:**
  `82b6c940f28297c30642eab72d1e1c6523245bb48584509551def7241de9c24c`
- **Script perbaikan:** `docs/uml/fix_pim_usecase_activity.py` (idempotent —
  rerun menghasilkan SHA identik; semua pemetaan memakai ID elemen + asersi
  ketat nilai lama; backup pra-perubahan:
  `docs/uml/cbt-smk-v2.mdj.pre-pim-usecase-activity.bak`)
- **Validator (sesudah PIM Use Case+Activity):**
  `docs/uml/validate-mdj-real.py` — 0 pelanggaran, LOLOS SEMUA; ekspektasi
  TIDAK berubah (tidak ada hitungan yang berubah); 3 kontrol negatif
  diuji ulang pada salinan di /tmp (NoteView tanpa teks, nama pesan ganda,
  label kompartemen ≠ nama model — ketiganya tetap GAGAL dengan benar).

## Prinsip pemetaan (berlaku untuk semua diagram)

1. **Perilaku/desain sistem TETAP SAMA** — hanya bahasa teknis implementasi
   diganti bahasa desain (Bahasa Indonesia). Tidak ada langkah bisnis yang
   ditambah/dihapus.
2. **Dihapus dari nama/dokumentasi:** endpoint HTTP (`POST /api/v1/...`,
   `GET ...`), query SQL (`SELECT/INSERT/UPDATE ...`), nama kolom/tabel DB
   (`password_hash`, `csrf_token`, `must_change_password`, `expires_at`,
   `password_lama`, `JSONB`, `sisa_waktu`, `users`, `sessions`, …), kode
   status implementasi (`403 ...`, `200 ...`), nama file/library (`auth.go`,
   `exam.go`, `grading.go`, `Argon2id`), dan istilah platform
   (`PostgreSQL`, `frontend`, `cookie`, `CSRF`, `token`, `attempt`, `WS`).
3. **Operasi desain** ditulis gaya camelCase + tanda kurung, mis.
   `masuk(kredensial)`, `cariPengguna()`, `hitungNilai()`.
4. **Lifeline = partisipan** (aktor/objek/komponen) sesuai OMG — fungsi
   TIDAK BOLEH menjadi lifeline.
5. **Geometri dipertahankan** kecuali yang dinyatakan eksplisit: posisi
   lifeline, points pesan horizontal, dan urutan pesan TIDAK berubah.

---

## Sequence 3a — Diagram Sequence Login + Wajib Ganti Password (SELESAI)

### 3a. Lifeline

| Nama lama (PSM) | Nama baru (PIM) | Keterangan |
|---|---|---|
| `Siswa / Pengguna` | `Pengguna` | Aktor; posisi tetap (x=100) |
| `LoginPage / GantiPasswordPage (frontend)` | `Halaman Login` | Posisi tetap (x=265) |
| `Auth API (auth.go)` | `Pengontrol Autentikasi` | Posisi tetap (x=430) |
| `Argon2id (library)` | **DIHAPUS** | Lifeline fungsi — melanggar OMG. Pesannya (`verifikasiKataSandi()`) kembali menjadi **self-message** pada `Pengontrol Autentikasi` di posisi semantiknya (setelah `kembalikanDataPengguna()`, sebelum `ditolakWajibGantiKataSandi()`); geometri loop dipulihkan dari backup `cbt-smk-v2.mdj.pre-sequence-opsi-b.bak` (keadaan sebelum opsi-lifeline): points `500:415;530:415;530:435;500:435` |
| `PostgreSQL: users, sessions` | `Basis Data` | Posisi tetap (x=760). Celah di x=595 (bekas lifeline yang dihapus) dibiarkan kosong — bukan cacat, melainkan jejak penghapusan |

### 3a. Pesan (16 pesan, urutan & arah & jenis TIDAK berubah)

| # | Nama lama (PSM) | Nama baru (PIM) |
|---|---|---|
| 1 | `isi username/email + password` | `isiKredensial()` |
| 2 | `POST /api/v1/auth/login` | `masuk(kredensial)` |
| 3 | `SELECT users WHERE username/email` | `cariPengguna()` |
| 4 | `return user + password_hash + must_change_password` | `kembalikanDataPengguna()` |
| 5 | `verify Argon2id(password, password_hash)` | `verifikasiKataSandi()` (self-message) |
| 6 | `403 {error.code: PASSWORD_CHANGE_REQUIRED}` | `ditolakWajibGantiKataSandi()` |
| 7 | `tampilkan form Ganti Password` | `tampilkanFormulirGantiKataSandi()` |
| 8 | `isi password_lama + password_baru` | `isiKataSandiLamaDanBaru()` |
| 9 | `POST /api/v1/auth/change-password` | `ubahKataSandi()` |
| 10 | `UPDATE users SET password_hash, must_change_password=…`[^1] | `perbaruiKataSandi()` |
| 11 | `OK` | `kembalikanKonfirmasi()` |
| 12 | `200 password berhasil diubah` | `kataSandiBerhasilDiubah()` |
| 13 | `POST /api/v1/auth/login (ulang)` | `masukUlang(kredensial)` |
| 14 | `INSERT sessions (csrf_token, expires_at)` | `buatSesi()` |
| 15 | `200 + session cookie + CSRF token` | `kembalikanSesi()` |
| 16 | `redirect ke dashboard / ujian` | `arahkanKeDasbor()` |

[^1]: Nilai penanda akhir pada nama lama disamarkan di sini karena merupakan
    detail implementasi; pemetaan di script memakai ID pesan
    `GXJ9BR7VJ3KW3A24P12GSJ`.

### 3a. Dokumentasi diagram

- **Lama:** "Diagram Sequence Login + Wajib Ganti Password: alur aktor Siswa,
  LoginPage / GantiPasswordPage (frontend), Auth API (auth.go), Argon2id
  (library), dan PostgreSQL (users, sessions) untuk login, verifikasi
  Argon2id, penanganan 403 PASSWORD_CHANGE_REQUIRED, ganti password, login
  ulang, pembuatan session, dan redirect. Urutan dibaca atas ke bawah.
  Legenda: garis solid panah terisi = panggilan sinkron; garis putus-putus =
  balasan (reply); pesan tanpa signature adalah label naratif
  (endpoint/operasi + keterangan), bukan rujukan Operation (lihat
  AUDIT-INDUSTRI.md K2)."
- **Baru:** "Diagram Sequence Login + Wajib Ganti Password: alur Pengguna,
  Halaman Login, Pengontrol Autentikasi, dan Basis Data untuk masuk,
  verifikasi kata sandi, penanganan penolakan wajib ganti kata sandi, ganti
  kata sandi, masuk ulang, pembuatan sesi, dan pengalihan ke dasbor. Urutan
  dibaca atas ke bawah. Legenda: garis solid panah terisi = panggilan
  sinkron; garis putus-putus = balasan (reply); pesan tanpa signature adalah
  label naratif (operasi + keterangan), bukan rujukan Operation (lihat
  AUDIT-INDUSTRI.md K2)."

### 3a. Catatan (Note) — ditulis ulang dalam bahasa desain

Note model + NoteView dibuat kembali (ID dipakai ulang dari backup
pra-opsi-b; opsi (b) sebelumnya menghapus Note ini total). Geometri view
sama dengan backup (left=100, top=1400, w=635, h=100).

- **Lama (backup pra-opsi-b):** "Catatan: 'verify
  Argon2id(password, password_hash)' secara semantik terjadi setelah 'return
  user + password_hash + must_change_password' dan sebelum '403 {error.code:
  PASSWORD_CHANGE_REQUIRED}'. StarUML selalu me-render self-message di bawah
  diagram."
- **Baru:** "Catatan: 'verifikasiKataSandi()' secara semantik terjadi setelah
  'kembalikanDataPengguna()' dan sebelum 'ditolakWajibGantiKataSandi()'.
  StarUML selalu me-render self-message di bawah diagram."

---

## Sequence 3b — Diagram Sequence Simpan Jawaban + Auto-Submit (SELESAI)

### 3b. Lifeline

| Nama lama (PSM) | Nama baru (PIM) | Keterangan |
|---|---|---|
| `Siswa` | `Pengguna` | Aktor; posisi tetap (x=100) |
| `UjianPage (frontend)` | `Halaman Ujian` | Posisi tetap (x=265) |
| `Exam API (exam.go)` | `Pengontrol Ujian` | Posisi tetap (x=430) |
| `autoSubmitJikaLewat (exam.go)` | **DIHAPUS** | Lifeline fungsi — melanggar OMG. Pesannya (`periksaBatasWaktu()`) kembali menjadi **self-message** pada `Pengontrol Ujian` di posisi semantiknya (setelah `kembalikanSisaWaktu()`, sebelum `hitungNilai()`); geometri loop dipulihkan dari backup `cbt-smk-v2.mdj.pre-sequence-opsi-b.bak`: points `500:815;530:815;530:835;500:835` |
| `Grading (grading.go)` | `Penghitung Nilai` | Posisi tetap (x=760) |
| `PostgreSQL: jawaban, attempt, nilai_akhir` | `Basis Data` | Posisi tetap (x=925). Celah di x=595 (bekas lifeline yang dihapus) dibiarkan kosong |
| `Monitor WebSocket` | `Layanan Pemantau` | Posisi tetap (x=1090) |

### 3b. Pesan (16 pesan, urutan & arah & jenis TIDAK berubah)

| # | Nama lama (PSM) | Nama baru (PIM) |
|---|---|---|
| 1 | `pilih / ubah jawaban soal` | `pilihJawaban()` |
| 2 | `POST /api/v1/attempt/:id/jawaban (autosave)` | `kirimJawaban()` |
| 3 | `INSERT/UPDATE jawaban (jawaban JSONB, ragu, poin)` | `simpanJawaban()` |
| 4 | `OK - jawaban tersimpan (autosave)` | `kembalikanKonfirmasiJawaban()` |
| 5 | `200 jawaban tersimpan` | `jawabanTersimpan()` |
| 6 | `GET status attempt (timer berjangkar server: deadline, sisa_waktu)` | `mintaStatusUjian()` |
| 7 | `SELECT attempt.deadline, jadwal.status` | `cariStatusUjian()` |
| 8 | `deadline + status` | `kembalikanStatusUjian()` |
| 9 | `sisa waktu server` | `kembalikanSisaWaktu()` |
| 10 | `cek deadline lewat (autoSubmitJikaLewat)` | `periksaBatasWaktu()` (self-message) |
| 11 | `hitungNilai: HitungPoin per soal (otomatis)` | `hitungNilai()` |
| 12 | `poin + nilai sementara` | `kembalikanNilaiSementara()` |
| 13 | `UPDATE attempt status=selesai, submit_at; syncNilaiAkhir` | `tandaiUjianSelesai()` |
| 14 | `OK - attempt selesai tersimpan (auto-submit)` | `kembalikanKonfirmasiSelesai()` |
| 15 | `broadcastAttempt (WS: status peserta berubah)` | `siarkanStatusPeserta()` |
| 16 | `200 hasil + nilai (final bila tanpa esai)` | `kembalikanHasilUjian()` |

Catatan: pesan #4 dan #14 sengaja dibedakan
(`kembalikanKonfirmasiJawaban()` vs `kembalikanKonfirmasiSelesai()`) karena
validator melarang nama ganda dalam satu interaksi.

### 3b. Dokumentasi diagram

- **Lama:** "Diagram Sequence Simpan Jawaban + Auto-Submit: alur aktor
  Siswa, UjianPage (frontend), Exam API (exam.go), autoSubmitJikaLewat
  (exam.go), Grading (grading.go), PostgreSQL (jawaban, attempt,
  nilai_akhir), dan Monitor WebSocket untuk autosave jawaban, polling status
  attempt berjangkar server, cek deadline, hitung nilai, auto-submit, dan
  siaran status via WebSocket. Urutan dibaca atas ke bawah. Legenda: garis
  solid panah terisi = panggilan sinkron; garis putus-putus = balasan;
  garis solid panah terbuka = panggilan asinkron (broadcastAttempt, tanpa
  balasan); pesan tanpa signature adalah label naratif (lihat
  AUDIT-INDUSTRI.md K2)."
- **Baru:** "Diagram Sequence Simpan Jawaban + Auto-Submit: alur Pengguna,
  Halaman Ujian, Pengontrol Ujian, Penghitung Nilai, Basis Data, dan Layanan
  Pemantau untuk simpan jawaban otomatis, permintaan status ujian berjangkar
  server, periksa batas waktu, hitung nilai, kirim otomatis, dan siaran
  status. Urutan dibaca atas ke bawah. Legenda: garis solid panah terisi =
  panggilan sinkron; garis putus-putus = balasan; garis solid panah terbuka
  = panggilan asinkron (siarkanStatusPeserta, tanpa balasan); pesan tanpa
  signature adalah label naratif (lihat AUDIT-INDUSTRI.md K2)."

### 3b. Catatan (Note) — ditulis ulang dalam bahasa desain

Note model + NoteView dibuat kembali (ID dipakai ulang dari backup
pra-opsi-b). Geometri view sama dengan backup (left=100, top=1480, w=965,
h=80).

- **Lama (backup pra-opsi-b):** "Catatan: 'cek deadline lewat
  (autoSubmitJikaLewat)' secara semantik terjadi setelah 'sisa waktu server'
  dan sebelum 'hitungNilai: HitungPoin per soal (otomatis)'. StarUML selalu
  me-render self-message di bawah diagram."
- **Baru:** "Catatan: 'periksaBatasWaktu()' secara semantik terjadi setelah
  'kembalikanSisaWaktu()' dan sebelum 'hitungNilai()'. StarUML selalu
  me-render self-message di bawah diagram."

---

## Yang TIDAK berubah pada Sequence (dijamin script + diverifikasi diff)

- Jumlah pesan tetap 32 (16 + 16); urutan, arah (source/target), dan
  `messageSort` (synchCall/reply/asynchCall) tetap.
- Points pesan horizontal tetap; hanya 2 pesan self-message yang points-nya
  dikembalikan ke loop backup.
- Posisi lifeline yang tersisa tetap (3a: 100/265/430/760; 3b:
  100/265/430/760/925/1090).
- `showSequenceNumber: false` dipertahankan di kedua diagram.
- Label pesan dihitung ulang dengan rumus lama (`width = floor(7.5 ×
  panjang teks)`, `left = titik-tengah-edge − width/2`); 4 label yang lebih
  panjang dari sebelumnya dicek ulang clearance-nya (tidak tumpang tindih
  label lain, tidak menyentuh kotak kepala lifeline).
- Tinggi kompartemen nama lifeline TIDAK diubah (StarUML me-reflow kotak
  nama saat render — temuan R9c); hanya teks label nama yang diselaraskan
  dengan nama model (diikat validator).
- Diagram lain (Use Case, Class, ERD, Activity) TIDAK disentuh — diff
  membuktikan hanya 142 objek berubah dan semuanya berada dalam rumpun
  Sequence (2 interaksi + 2 diagram + 10 lifeline + 32 pesan + view/label
  terkait + 2 Note baru + induk Project/Model).

## Batasan jujur

- Verifikasi visual di **StarUML asli belum dilakukan** pada hasil PIM ini
  (StarUML tidak tersedia di lingkungan worker ini) — perlu spot-check
  1× di StarUML sungguhan (buka file, pastikan 2 self-message + 2 Note
  tampil benar dan tidak ada kotak kosong).
- Label yang menyeberangi garis lifeline diterima sebagai pola standar
  (temuan sebelumnya: 28/30 label lebih lebar dari jarak antar-lifeline);
  self-message tetap di-render StarUML di bawah diagram (perilaku StarUML),
  sehingga Note penjelasan tetap diperlukan.

---

## Use Case — Diagram Use Case + 2 Note BATASAN (SELESAI)

- **Nama use case:** 7/19 diganti (yang mengandung istilah non-desain).
  12 nama lain sudah bahasa desain, tetap.
- **Dokumentasi:** 19/19 use case ditulis ulang + dokumentasi 6 aktor +
  dokumentasi 3 Include/Extend + condition 2 Extend + dokumentasi diagram
  Use Case + 2 Note BATASAN (model `documentation` **dan** `text` — model
  UMLNote menyimpan teks di dua field; keduanya disinkronkan).
- **Jumlah TETAP:** 6 aktor, 19 use case, 39 asosiasi aktor–use case,
  3 Include, 2 Extend, 2 Note.
- **Geometri:** posisi (left/top) SEMUA view TETAP; points garis TETAP.
  Pelebaran hanya 2 elips yang namanya memanjang (left/top tetap, tidak
  menggeser view lain): UC01 260→280, UC17 260→340. Label nama asosiasi /
  include / extend (EdgeLabelView, `visible: false`) ikut disinkronkan
  (21 label). Label nama use case di view (NameCompartment) disinkronkan
  dengan nama model (diikat validator).

### Use Case. Nama (hanya yang BERUBAH)

| Nama lama (PSM) | Nama baru (PIM) |
|---|---|
| `UC01 Login & Ganti Password Wajib` | `UC01 Masuk & Wajib Ganti Kata Sandi` |
| `UC04 Import Soal` | `UC04 Impor Soal` |
| `UC07 Monitor Ujian Real-time` | `UC07 Pantau Ujian Langsung` |
| `UC15 Kelola Event, Ruangan & Kartu Peserta` | `UC15 Kelola Kegiatan, Ruangan & Kartu Peserta` |
| `UC16 Backup & Restore` | `UC16 Cadangkan & Pulihkan` |
| `UC17 Kelola Settings, Audit & Notifikasi` | `UC17 Kelola Pengaturan, Audit & Notifikasi` |
| `UC18 Usul & Proses Unblock` | `UC18 Usul & Proses Buka Blokir` |

### Use Case. Dokumentasi aktor

| Aktor | Lama | Baru |
|---|---|---|
| Admin, Proktor, Guru, Pengawas, Siswa, Penguji UKK | `Aktor CBT SMK v2. Role users.role = '<nilai>'.` | `Aktor CBT SMK v2. Peran: <NamaPeran>.` |

(`users.role` = rujukan kolom DB; `<nilai>` = `admin` / `proktor` /
`guru` / `pengawas` / `siswa` / `penguji_ukk`.)

### Use Case. Dokumentasi use case (19/19)

| UC | Lama | Baru |
|---|---|---|
| UC01 | `Login session cookie + CSRF; bila users.must_change_password=true -> 403 PASSWORD_CHANGE_REQUIRED, wajib ganti password sebelum lanjut. Endpoint auth.go.` | `Masuk dengan sesi; bila pengguna wajib ganti kata sandi, sistem menolak masuk dan menuntut ganti kata sandi sebelum dapat melanjutkan.` |
| UC02 | `CRUD Jurusan, Kelas, Mapel, Guru, Siswa (+ Import Siswa CSV). Hapus permanen (hard delete), ditolak 409 bila masih dipakai.` | `Kelola (tambah, ubah, lihat, hapus) Jurusan, Kelas, Mapel, Guru, Siswa (+ Impor Siswa dari berkas). Hapus permanen, ditolak bila data masih dipakai.` |
| UC03 | `CRUD Bank Soal + Soal (16 tipe: 5 inti AKM + 11 varian). Bank terkunci (is_locked) tidak bisa diubah. Proktor DILARANG lihat isi soal.` | `Kelola (tambah, ubah, lihat, hapus) Bank Soal + Soal (16 tipe: 5 inti AKM + 11 varian). Bank terkunci tidak bisa diubah. Proktor DILARANG lihat isi soal.` |
| UC04 | `Import Excel (.xlsx) & Word (.docx) via preview dry-run, template, validasi E1-E32/W1-W7, mode all_or_nothing/skip_errors.` | `Impor Excel & Word melalui pratinjau tanpa menyimpan, templat, validasi E1-E32/W1-W7, mode semua-atau-tidak-sama-sekali / lewati-yang-salah.` |
| UC05 | `CRUD Paket Ujian (STS/SAS/USBK/UKK/Tryout) + tambah/hapus soal ke paket, pembangkitan acak, sinkron jumlah soal.` | `Kelola (tambah, ubah, lihat, hapus) Paket Ujian (STS/SAS/USBK/UKK/Tryout) + tambah/hapus soal ke paket, pembangkitan acak, sinkron jumlah soal.` |
| UC06 | `CRUD Jadwal Ujian, pause/resume (pause membekukan timer), generate token (hash SHA, expiry 900 detik, token baru menghanguskan lama). Pengawas DILARANG lihat token.` | `Kelola (tambah, ubah, lihat, hapus) Jadwal Ujian, jeda/lanjut (jeda membekukan pewaktu), buat token (token baru menghanguskan yang lama, berlaku 900 detik). Pengawas DILARANG lihat token.` |
| UC07 | `Monitor peserta per jadwal via WebSocket, status attempt, pelanggaran, unblock.` | `Pantau peserta per jadwal secara langsung, status pengerjaan ujian, pelanggaran, buka blokir.` |
| UC08 | `Siswa: lihat ujian aktif, input token, kerjakan soal, autosave jawaban, tandai ragu-ragu, submit / auto-submit saat deadline.` | `Siswa: lihat ujian aktif, masukkan token, kerjakan soal, simpan jawaban otomatis, tandai ragu-ragu, kirim / kirim otomatis saat batas waktu.` |
| UC09 | `Validasi token jadwal: hash cocok, belum kedaluwarsa (900 dtk), dalam toleransi telat 15 menit.` | `Validasi token jadwal: token cocok, belum kedaluwarsa (900 detik), dalam toleransi telat 15 menit.` |
| UC10 | `Guru mengoreksi jawaban uraian/manual (listKoreksi, koreksiJawaban), poin manual, status dikoreksi.` | `Guru mengoreksi jawaban uraian/manual (daftar koreksi, koreksi jawaban), poin manual, status dikoreksi.` |
| UC11 | `Rekap nilai akhir per jadwal, predikat, remedial (nilai tertinggi), berita acara (isi JSONB).` | `Rekap nilai akhir per jadwal, predikat, remedial (nilai tertinggi), berita acara (isi terstruktur).` |
| UC12 | `Analisis daya beda + pengecoh per soal, export CSV. Permission analisis.read / analisis.export.` | `Analisis daya beda + pengecoh per soal, ekspor berkas. Izin: baca analisis / ekspor analisis.` |
| UC13 | `Nilai per kompetensi (soal -> kompetensi_id) per siswa/jadwal.` | `Nilai per kompetensi (soal terhubung ke kompetensi) per siswa/jadwal.` |
| UC14 | `CRUD UKK Ujian, nilai rubrik (Proses 30% / Produk 50% / Sikap 20%, teori 40% + praktik 60%, lulus >=70), berkas praktik, sertifikat (hanya lulus, nomor UKK/tahun/kode-jurusan/urut).` | `Kelola (tambah, ubah, lihat, hapus) UKK Ujian, nilai rubrik (Proses 30% / Produk 50% / Sikap 20%, teori 40% + praktik 60%, lulus >=70), berkas praktik, sertifikat (hanya lulus, nomor UKK/tahun/kode-jurusan/urut).` |
| UC15 | `CRUD Event & Ruangan (jadwal.event_id, jadwal.ruangan_id), kartu peserta per jadwal; siswa melihat Kartu Saya.` | `Kelola (tambah, ubah, lihat, hapus) Kegiatan & Ruangan (jadwal terhubung ke kegiatan dan ruangan), kartu peserta per jadwal; siswa melihat Kartu Saya.` |
| UC16 | `Buat backup (metadata backup_meta, file .dump di storage/backup, nama unik), restore dengan pre-restore otomatis, hapus backup (pre-restore <7 hari dilindungi).` | `Buat cadangan (data cadangan, berkas cadangan di penyimpanan cadangan, nama unik), pulihkan dengan pra-pemulihan otomatis, hapus cadangan (pra-pemulihan <7 hari dilindungi).` |
| UC17 | `Settings key/value JSONB, audit_log append-only, notifikasi dalam aplikasi, SoD override (sod.override, alasan tertulis).` | `Pengaturan pasangan kunci-nilai, jejak audit hanya-tambah, notifikasi dalam aplikasi, pengecualian pemisahan tugas (alasan tertulis).` |
| UC18 | `Pengawas mengusulkan unblock (usulan_unblock, alasan); Proktor/Admin menyetujui/menolak; unblock = lanjut sisa waktu.` | `Pengawas mengusulkan buka blokir (usulan buka blokir, alasan); Proktor/Admin menyetujui/menolak; buka blokir = lanjut sisa waktu.` |
| UC19 | `Siswa melihat hasil bila tampil_hasil efektif (paket.tampil_hasil AND settings.tampil_hasil_ke_siswa); pembahasan hanya setelah submit / Tryout.` | `Siswa melihat hasil bila tampil hasil aktif (pengaturan paket DAN pengaturan sistem); pembahasan hanya setelah kirim / Tryout.` |

Catatan: nilai penanda pada dokumentasi lama UC01 ditulis apa adanya dari
file (`must_change_password=true`); pemetaan di script memakai ID use case
`PGJQUVQY8OOR3PPEMVTQ7U`.

### Use Case. Include / Extend (dokumentasi + condition)

| Relasi | Field | Lama | Baru |
|---|---|---|---|
| Extend Unblock→Monitor (`VW3VFH11QHGKA491UCKFQI`) | dokumentasi | `Unblock adalah alur perpanjangan dari monitor saat attempt diblokir.` | `Buka blokir adalah alur perpanjangan dari pemantauan saat pengerjaan ujian diblokir.` |
| Extend Unblock→Monitor | condition | `attempt diblokir` | `pengerjaan ujian diblokir` |
| Extend Import→Bank Soal (`K5XYTJYZDP926I4YUGGNAK`) | dokumentasi | `Import adalah cara alternatif mengisi bank soal.` | `Impor adalah cara alternatif mengisi bank soal.` |
| Extend Import→Bank Soal | condition | `import file dipilih` | `berkas impor dipilih` |
| Include Rekap→Koreksi (`AP3V228P5UWK9YR2EWQHJZ`) | dokumentasi | `Rekap final menunggu koreksi esai selesai (status menunggu_koreksi -> final).` | `Rekap final menunggu koreksi esai selesai (status menunggu koreksi → final).` |

Dokumentasi 2 Include lain (`Ikuti Ujian selalu memvalidasi token…`,
`Susun paket mengambil soal dari bank soal.`) sudah bahasa desain, tetap.

### Use Case. 2 Note BATASAN

| Note | Lama | Baru |
|---|---|---|
| Pengawas–token | `BATASAN: Pengawas TIDAK boleh melihat token ujian (token hanya Proktor/Admin; endpoint tokenStatus & generate token memerlukan permission jadwal).` | `BATASAN: Pengawas TIDAK boleh melihat token ujian (token hanya Proktor/Admin; pembuatan token dan pengecekan status token memerlukan izin jadwal).` |
| Proktor–soal | `BATASAN: Proktor TIDAK boleh melihat isi soal (forbidSoalUntukProktor di banksoal.go; endpoint soal menolak role proktor).` | `BATASAN: Proktor TIDAK boleh melihat isi soal (sistem menolak akses soal bagi peran proktor).` |

Isi aturan TETAP SAMA — hanya bahasanya yang desain. NoteView
(`UMLNoteView.text`) disinkronkan dengan model.

### Use Case. Dokumentasi diagram

- **Lama:** "…yang mencakup master data, bank soal, paket, jadwal,
  pelaksanaan ujian, koreksi, rekap, UKK, event, backup, dan unblock. …"
- **Baru:** "…yang mencakup data master, bank soal, paket, jadwal,
  pelaksanaan ujian, koreksi, rekap, UKK, kegiatan, cadangan, dan buka
  blokir. …" (sisa dokumentasi diagram tidak berubah).

---

## Activity — Diagram Activity 2a + 2b (SELESAI)

- **Nama aksi/decision:** 2a: 14/19 berubah (5 sudah bahasa desain:
  `Tolak: token salah / kedaluwarsa`, `Kerjakan Soal (…)`,
  `Pelanggaran ke-1: …`, `Mulai`, `Selesai`); 2b: 5/11 berubah (6 tetap:
  `Pilih Jawaban Esai / Uraian siswa`, `Nilai dengan Rubrik: …`,
  `Semua jawaban manual terkoreksi?`, `Rekap Nilai per Jadwal: …`,
  `Mulai`, `Selesai`).
- **Guard:** 2a: 7/25 berubah; 2b: 0/11 berubah (`belum semua`,
  `semua terkoreksi` sudah bahasa desain).
- **Dokumentasi:** 2 diagram Activity ditulis ulang.
- **Jumlah TETAP:** 2a: 19 node + 25 flow; 2b: 11 node + 11 flow
  (36 flow total di kedua diagram). Tidak ada aksi,
  guard, decision, atau alur yang ditambah/dihapus.
- **Geometri:** posisi (left/top) + points SEMUA view TETAP; persilangan
  garis TIDAK berubah. Lebar label guard diskalakan proporsional
  (`lebar_baru = lebar_lama × panjang_baru / panjang_lama`) mengikuti
  kalibrasi tersimpan. Penambahan tinggi hanya 1 aksi yang teks barunya
  +1 baris (per ~30 karakter/baris): `Mulai Ujian` 76→92px (view +
  kompartemen + label). Aksi lain yang memendek TIDAK dikecilkan
  (teks lebih pendek tidak pernah meluber).

### Activity 2a. Nama aksi / decision

| Nama lama (PSM) | Nama baru (PIM) |
|---|---|
| `Login (session + CSRF)` | `Masuk` |
| `Input Token Ujian` | `Masukkan Token Ujian` |
| `Token valid? (hash, expiry 900 dtk, toleransi 15 mnt)` | `Token valid? (cocok, belum kedaluwarsa 900 detik, toleransi telat 15 menit)` |
| `Mulai Ujian: buat Attempt, deadline = server, urutan soal diacak per-attempt` | `Mulai Ujian: buat pengerjaan ujian, batas waktu dari sistem, urutan soal diacak per pengerjaan` |
| `Autosave: simpanJawaban -> UPDATE jawaban (jawaban JSONB, ragu, poin)` | `Simpan jawaban otomatis (jawaban, ragu-ragu, poin)` |
| `Tandai Ragu-ragu (baris jawaban NULL dibuat bila belum ada)` | `Tandai Ragu-ragu (jawaban kosong dibuat bila belum ada)` |
| `Pindah tab / blur?` | `Pindah tab / jendela tidak aktif?` |
| `Pelanggaran ke-2: auto-block Attempt (status diblokir, sisa waktu dibekukan)` | `Pelanggaran ke-2: blokir otomatis pengerjaan ujian (status diblokir, sisa waktu dibekukan)` |
| `Usul Unblock (Pengawas) -> disetujui Proktor -> lanjut sisa waktu` | `Usul Buka Blokir (Pengawas) -> disetujui Proktor -> lanjut sisa waktu` |
| `Jadwal Pause: bekukan timer (sisa_waktu_beku); Resume: deadline = now + sisa` | `Jeda Jadwal: bekukan pewaktu (sisa waktu beku); Lanjut: batas waktu = saat ini + sisa` |
| `Waktu habis / deadline lewat?` | `Waktu habis / batas waktu lewat?` |
| `Submit Ujian (siswa menekan Selesai)` | `Kirim Ujian (siswa menekan Selesai)` |
| `Auto-Submit server (autoSubmitJikaLewat)` | `Kirim Otomatis (bila lewat batas waktu)` |
| `Finalisasi: hitungNilai (otomatis) + syncNilaiAkhir (menunggu_koreksi bila ada esai)` | `Finalisasi: hitung nilai (otomatis) + selaraskan nilai akhir (menunggu koreksi bila ada esai)` |

### Activity 2a. Guard

| Guard lama | Guard baru |
|---|---|
| `usul unblock` | `usul buka blokir` |
| `jadwal pause` | `jadwal dijeda` |
| `resume` | `lanjut` |
| `cek waktu / submit` | `cek waktu / kirim` |
| `siswa submit` | `siswa kirim` |
| `deadline lewat` | `batas waktu lewat` |
| `pindah tab / blur` | `pindah tab / jendela tidak aktif` |

Guard lain 2a (`tidak valid`, `valid`, `jawab`, `tandai ragu`, `ke-1`,
`ke-2`, `disetujui`) sudah bahasa desain, tetap.

### Activity 2a. Dokumentasi diagram

- **Lama:** "…alur siswa dari login, input token, validasi token,
  mengerjakan soal (autosave, ragu-ragu), penanganan pindah tab / blur
  (peringatan ke-1, auto-block ke-2, usul unblock), pause/resume jadwal,
  sampai submit / auto-submit dan finalisasi nilai. … teks [guard] di
  dekat ekor edge menyatakan syarat token melewati edge itu."
- **Baru:** "…alur siswa dari masuk, masukkan token, validasi token,
  mengerjakan soal (simpan jawaban otomatis, ragu-ragu), penanganan
  pindah tab / jendela tidak aktif (peringatan ke-1, blokir otomatis
  ke-2, usul buka blokir), jeda/lanjut jadwal, sampai kirim / kirim
  otomatis dan finalisasi nilai. … teks [guard] di dekat ekor edge
  menyatakan syarat untuk melewati edge itu."

### Activity 2b. Nama aksi

| Nama lama (PSM) | Nama baru (PIM) |
|---|---|
| `Guru: buka Daftar Koreksi (listKoreksi, jawaban dikoreksi=false, tipe manual)` | `Guru: buka Daftar Koreksi (jawaban belum dikoreksi, tipe manual)` |
| `Simpan Koreksi (koreksiJawaban -> UPDATE jawaban.poin, dikoreksi=true)` | `Simpan Koreksi (perbarui poin jawaban, sudah dikoreksi)` |
| `Hitung Nilai Akhir: SUM poin / SUM bobot x 100 (grading.HitungNilaiAkhir)` | `Hitung Nilai Akhir: jumlah poin / jumlah bobot × 100` |
| `Status NilaiAkhir -> final (bila tidak ada esai, final otomatis saat submit)` | `Status Nilai Akhir → final (bila tidak ada esai, final otomatis saat kirim)` |
| `Buat Berita Acara (berita_acara.isi JSONB, dibuat_oleh)` | `Buat Berita Acara (isi berita acara, dibuat oleh)` |

### Activity 2b. Dokumentasi diagram

- **Lama:** "…lalu hitung nilai akhir (SUM poin / SUM bobot x 100),
  finalisasi status nilai, …"
- **Baru:** "…lalu hitung nilai akhir (jumlah poin / jumlah bobot × 100),
  finalisasi status nilai, …" (sisa dokumentasi tidak berubah).

### Yang TIDAK berubah pada Use Case + Activity (dijamin script + diverifikasi diff)

- Jumlah elemen TETAP (6 aktor, 19 use case, 39 asosiasi UC, 3 Include,
  2 Extend, 2 Note BATASAN, 2a 19 node + 25 flow, 2b 11 node + 11 flow,
  total objek file 3401, 0 tambah / 0 hapus).
- Perilaku/desain TETAP: tidak ada langkah bisnis, use case, aktor,
  asosiasi, aksi, guard, decision, condition, atau alur yang
  ditambah/dihapus — hanya bahasanya.
- Posisi (left/top) + points garis SEMUA view TETAP; hanya lebar 2 elips
  (UC01, UC17), tinggi 1 aksi (Mulai Ujian), dan lebar label guard/edge
  yang disesuaikan proporsional.
- Sequence, Class, ERD TIDAK disentuh — diff membuktikan 120 objek yang
  berubah field skalarnya semuanya berada dalam rumpun Use Case +
  Activity (sisanya hanya kontainer induk yang berubah karena isi
  turunannya berubah).
- `token` dipertahankan sebagai istilah domain (bukan istilah
  implementasi): token ujian adalah konsep bisnis (siswa memasukkan
  token, Pengawas dilarang melihat token). Yang dihapus hanya mekanisme
  implementasi di sekitarnya (`endpoint tokenStatus`, `hash`, `expiry`,
  `cookie`, `CSRF`, `WebSocket`, nama file, nama kolom, kode status).

### Batasan jujur

- Verifikasi visual di **StarUML asli belum dilakukan** pada hasil PIM ini
  (StarUML tidak tersedia di lingkungan worker ini) — perlu spot-check
  1× di StarUML sungguhan (buka file, pastikan elips UC01/UC17 tidak
  meluber, aksi `Mulai Ujian` tidak terpotong, dan label guard 2a tidak
  menutupi garis).
- Label guard yang memanjang (`pindah tab / jendela tidak aktif`,
  17→30 karakter) lebarnya diskalakan proporsional dari kalibrasi
  tersimpan; clearance terhadap garis lain mengikuti aturan kalibrasi
  sebelumnya (posisi label ditentukan `(edgePosition, distance, alpha)`
  + titik edge TERGAMBAR — `left/top` tersimpan DIABAIKAN saat render).

## Class — Diagram Class Domain (MINIMAL, SELESAI)

Class dipertahankan strukturnya (dosen pembimbing menilai Class sudah LUMAYAN — TIDAK dirombak total). Nama class TIDAK berubah (sudah PascalCase bahasa desain). Yang berubah hanya bahasa/tipe atribut, operasi, asosiasi, dan dokumentasi diagram.

- **Atribut:** 123 nama diganti (snake_case kolom-DB → camelCase Bahasa Indonesia), 115 tipe digenerikkan (dari 229 atribut).
- **Operasi:** 9 dari 11 diganti nama/tipe kembaliannya; `wajibGantiPassword()` dan `kunci()` sudah bahasa desain, tetap.
- **Asosiasi:** 46/46 diganti nama (`tabel.kolom` → `Class.atributDesain`) + label view-nya. Multiplisitas, ujung (source/target), dan rute garis TIDAK berubah.
- **Geometri:** posisi (left/top) SEMUA view TETAP. Pelebaran kotak hanya 3 ClassView yang teks barunya meluber (left/top tetap, tidak menggeser view lain): Jawaban 170→190, Pelanggaran 160→190, Kompetensi 160→180.
- **Jumlah TETAP:** 34 class, 229 atribut, 11 operasi, 46 asosiasi.

### Class. Pemetaan tipe atribut

| Tipe lama (PSM) | Tipe baru (PIM) | Keterangan |
|---|---|---|
| `UUID` | `String` | Pengenal generik (bukan tipe DB) |
| `JSONB` | `String` | Isi terstruktur sebagai teks |
| `DateTime` | `Date` | Waktu generik |
| `Numeric`, `Numeric(5,2)`, `Numeric(8,2)` | `Decimal` | Angka desimal generik |
| `String`, `Integer`, `Boolean`, `Date`, `Time` | tetap | Sudah generik |
| Token enum `penguji_ukk` / `perlu_revisi` / `menunggu_koreksi` / `pre-restore` | `pengujiUkk` / `perluRevisi` / `menungguKoreksi` / `praPemulihan` | Nilai enum dibersihkan |

### Class. Atribut (hanya yang BERUBAH; atribut lain tetap)

#### Class `User`

| Nama lama | Nama baru |
|---|---|
| `username` | `namaPengguna` |
| `password_hash` | `kataSandi` |
| `role` | `peran` |
| `must_change_password` | `wajibGantiKataSandi` |
| `is_active` | `aktif` |
| `created_at` | `dibuatPada` |
| `updated_at` | `diperbaruiPada` |

#### Class `Role`

| Nama lama | Nama baru |
|---|---|
| `code` | `kode` |
| `name` | `nama` |
| `description` | `deskripsi` |
| `is_active` | `aktif` |

#### Class `Permission`

| Nama lama | Nama baru |
|---|---|
| `code` | `kode` |
| `description` | `deskripsi` |
| `group_name` | `namaKelompok` |

#### Class `Siswa`

| Nama lama | Nama baru |
|---|---|
| `user_id` | `penggunaId` |
| `kelas_id` | `kelasId` |
| `jurusan_id` | `jurusanId` |
| `tahun_masuk` | `tahunMasuk` |
| `is_active` | `aktif` |

#### Class `Guru`

| Nama lama | Nama baru |
|---|---|
| `user_id` | `penggunaId` |
| `mapel_ids` | `daftarMapel` |

#### Class `Kelas`

| Nama lama | Nama baru |
|---|---|
| `jurusan_id` | `jurusanId` |
| `tahun_ajaran` | `tahunAjaran` |

#### Class `Mapel`

| Nama lama | Nama baru |
|---|---|
| `jurusan_id` | `jurusanId` |
| `kelas_tingkat` | `tingkatKelas` |

#### Class `Kompetensi`

| Nama lama | Nama baru |
|---|---|
| `mapel_id` | `mapelId` |

#### Class `KisiKisi`

| Nama lama | Nama baru |
|---|---|
| `mapel_id` | `mapelId` |
| `kelas_tingkat` | `tingkatKelas` |
| `kompetensi_id` | `kompetensiId` |
| `jumlah_soal` | `jumlahSoal` |

#### Class `BankSoal`

| Nama lama | Nama baru |
|---|---|
| `mapel_id` | `mapelId` |
| `kelas_tingkat` | `tingkatKelas` |
| `guru_id` | `guruId` |
| `is_locked` | `terkunci` |

#### Class `Soal`

| Nama lama | Nama baru |
|---|---|
| `bank_id` | `bankSoalId` |
| `kisi_kisi_id` | `kisiKisiId` |
| `kompetensi_id` | `kompetensiId` |
| `media_url` | `tautanMedia` |
| `level` | `tingkat` |

#### Class `PaketUjian`

| Nama lama | Nama baru |
|---|---|
| `mapel_id` | `mapelId` |
| `durasi_menit` | `durasiMenit` |
| `jumlah_soal` | `jumlahSoal` |
| `acak_soal` | `acakSoal` |
| `acak_jawaban` | `acakJawaban` |
| `tampil_hasil` | `tampilHasil` |

#### Class `PaketSoal`

| Nama lama | Nama baru |
|---|---|
| `paket_id` | `paketUjianId` |
| `soal_id` | `soalId` |

#### Class `JadwalUjian`

| Nama lama | Nama baru |
|---|---|
| `paket_id` | `paketUjianId` |
| `kelas_id` | `kelasId` |
| `jam_mulai` | `jamMulai` |
| `jam_selesai` | `jamSelesai` |
| `token_hash` | `token` |
| `token_expires_at` | `tokenBerlakuHingga` |
| `event_id` | `kegiatanId` |
| `ruangan_id` | `ruanganId` |

#### Class `Attempt`

| Nama lama | Nama baru |
|---|---|
| `jadwal_id` | `jadwalUjianId` |
| `siswa_id` | `siswaId` |
| `mulai_at` | `mulaiPada` |
| `deadline` | `batasWaktu` |
| `sisa_waktu_beku` | `sisaWaktuBeku` |
| `urutan_soal` | `urutanSoal` |
| `urutan_opsi` | `urutanOpsi` |
| `submit_at` | `dikirimPada` |
| `nilai_akhir` | `nilaiAkhir` |

#### Class `Jawaban`

| Nama lama | Nama baru |
|---|---|
| `attempt_id` | `pengerjaanUjianId` |
| `soal_id` | `soalId` |

#### Class `Pelanggaran`

| Nama lama | Nama baru |
|---|---|
| `attempt_id` | `pengerjaanUjianId` |

#### Class `UsulanUnblock`

| Nama lama | Nama baru |
|---|---|
| `attempt_id` | `pengerjaanUjianId` |
| `pengawas_id` | `pengawasId` |
| `diproses_oleh` | `diprosesOleh` |
| `diproses_at` | `diprosesPada` |

#### Class `NilaiAkhir`

| Nama lama | Nama baru |
|---|---|
| `attempt_id` | `pengerjaanUjianId` |
| `jadwal_id` | `jadwalUjianId` |
| `siswa_id` | `siswaId` |

#### Class `BeritaAcara`

| Nama lama | Nama baru |
|---|---|
| `jadwal_id` | `jadwalUjianId` |
| `dibuat_oleh` | `dibuatOleh` |

#### Class `Remedial`

| Nama lama | Nama baru |
|---|---|
| `jadwal_id` | `jadwalUjianId` |
| `siswa_id` | `siswaId` |

#### Class `UkkUjian`

| Nama lama | Nama baru |
|---|---|
| `mapel_id` | `mataPelajaranId` |
| `kelas_id` | `kelasId` |
| `jurusan_id` | `jurusanId` |
| `penguji_internal` | `pengujiInternal` |
| `penguji_eksternal` | `pengujiEksternal` |

#### Class `UkkNilai`

| Nama lama | Nama baru |
|---|---|
| `ukk_id` | `ujianUkkId` |
| `siswa_id` | `siswaId` |
| `nilai_teori` | `nilaiTeori` |
| `nilai_proses` | `nilaiProses` |
| `nilai_produk` | `nilaiProduk` |
| `nilai_sikap` | `nilaiSikap` |
| `nilai_praktik` | `nilaiPraktik` |
| `nilai_akhir` | `nilaiAkhir` |
| `dinilai_oleh` | `dinilaiOleh` |

#### Class `UkkBerkas`

| Nama lama | Nama baru |
|---|---|
| `ukk_id` | `ujianUkkId` |
| `siswa_id` | `siswaId` |
| `nama_file` | `namaBerkas` |
| `path` | `lokasiBerkas` |
| `diupload_oleh` | `diunggahOleh` |

#### Class `UkkSertifikat`

| Nama lama | Nama baru |
|---|---|
| `ukk_id` | `ujianUkkId` |
| `siswa_id` | `siswaId` |
| `nomor_sertifikat` | `nomorSertifikat` |
| `diterbitkan_at` | `diterbitkanPada` |

#### Class `Event`

| Nama lama | Nama baru |
|---|---|
| `tanggal_mulai` | `tanggalMulai` |
| `tanggal_selesai` | `tanggalSelesai` |

#### Class `BackupMeta`

| Nama lama | Nama baru |
|---|---|
| `nama_file` | `namaBerkas` |
| `ukuran_bytes` | `ukuranBita` |
| `dibuat_oleh` | `dibuatOleh` |

#### Class `SodOverride`

| Nama lama | Nama baru |
|---|---|
| `actor_id` | `pelakuId` |
| `entitas_id` | `entitasId` |

#### Class `Notifikasi`

| Nama lama | Nama baru |
|---|---|
| `penerima_user_id` | `penerimaId` |

#### Class `AuditLog`

| Nama lama | Nama baru |
|---|---|
| `actor_id` | `pelakuId` |
| `entitas_id` | `entitasId` |
| `ip` | `alamatIp` |

#### Class `Settings`

| Nama lama | Nama baru |
|---|---|
| `key` | `kunci` |
| `value` | `nilai` |
| `group_name` | `namaKelompok` |
| `description` | `deskripsi` |

#### Class `Session`

| Nama lama | Nama baru |
|---|---|
| `user_id` | `penggunaId` |
| `role` | `peran` |
| `csrf_token` | `tokenSesi` |
| `ip` | `alamatIp` |
| `user_agent` | `agenPengguna` |
| `expires_at` | `berlakuHingga` |
| `last_seen_at` | `terakhirTerlihatPada` |

Atribut yang tidak tercantum di atas TIDAK berubah namanya; tipenya digenerikkan bila semula teknis (lihat hitungan tipe). Class `Jurusan` dan `Ruangan` tidak punya rename atribut (hanya tipe `id` UUID → String).

### Class. Operasi

| Nama lama | Nama baru |
|---|---|
| `verifyPassword(p): Boolean` | `verifikasiKataSandi(kataSandi): Boolean` |
| `wajibGantiPassword(): Boolean` | `wajibGantiPassword(): Boolean` (tetap) |
| `kunci(): Boolean` | `kunci(): Boolean` (tetap) |
| `generateToken(): String` | `hasilkanToken(): String` |
| `pause(): void` | `jeda(): void` |
| `resume(): void` | `lanjutkan(): void` |
| `sisaWaktu(now): Integer` | `sisaWaktu(sekarang): Integer` |
| `autoSubmitJikaLewat(): Boolean` | `kirimOtomatisJikaLewat(): Boolean` |
| `hitung(): Numeric` | `hitung(): Decimal` |
| `hitungPraktik(): Numeric` | `hitungPraktik(): Decimal` |
| `hitungAkhir(): Numeric` | `hitungAkhir(): Decimal` |

### Class. Asosiasi (46)

| Nama lama | Nama baru |
|---|---|
| `attempt.jadwal_id` | `Attempt.jadwalUjianId` |
| `attempt.siswa_id -> users.id` | `Attempt.siswaId -> User.id` |
| `audit_log.actor_id` | `AuditLog.pelakuId` |
| `backup_meta.dibuat_oleh` | `BackupMeta.dibuatOleh` |
| `bank_soal.mapel_id` | `BankSoal.mataPelajaranId` |
| `berita_acara.jadwal_id UNIQUE` | `BeritaAcara.jadwalUjianId UNIK` |
| `guru.user_id` | `Guru.penggunaId` |
| `jadwal_ujian.event_id` | `JadwalUjian.kegiatanId` |
| `jadwal_ujian.kelas_id` | `JadwalUjian.kelasId` |
| `jadwal_ujian.paket_id` | `JadwalUjian.paketUjianId` |
| `jadwal_ujian.ruangan_id` | `JadwalUjian.ruanganId` |
| `jawaban.attempt_id` | `Jawaban.pengerjaanUjianId` |
| `jawaban.soal_id` | `Jawaban.soalId` |
| `kelas.jurusan_id` | `Kelas.jurusanId` |
| `kisi_kisi.kompetensi_id` | `KisiKisi.kompetensiId` |
| `kisi_kisi.mapel_id` | `KisiKisi.mataPelajaranId` |
| `kompetensi.mapel_id` | `Kompetensi.mataPelajaranId` |
| `nilai_akhir.attempt_id UNIQUE` | `NilaiAkhir.pengerjaanUjianId UNIK` |
| `nilai_akhir.jadwal_id` | `NilaiAkhir.jadwalUjianId` |
| `nilai_akhir.siswa_id` | `NilaiAkhir.siswaId` |
| `notifikasi.penerima_user_id` | `Notifikasi.penerimaId` |
| `paket_soal.paket_id` | `PaketSoal.paketUjianId` |
| `paket_soal.soal_id` | `PaketSoal.soalId` |
| `paket_ujian.mapel_id` | `PaketUjian.mataPelajaranId` |
| `pelanggaran.attempt_id` | `Pelanggaran.pengerjaanUjianId` |
| `remedial.jadwal_id` | `Remedial.jadwalUjianId` |
| `remedial.siswa_id` | `Remedial.siswaId` |
| `role_permissions (role_permissions.role_id, permission_id)` | `PeranIzin (PeranIzin.peranId, izinId)` |
| `sessions.user_id` | `Session.penggunaId` |
| `siswa.jurusan_id` | `Siswa.jurusanId` |
| `siswa.kelas_id` | `Siswa.kelasId` |
| `siswa.user_id` | `Siswa.penggunaId` |
| `soal.bank_id` | `Soal.bankSoalId` |
| `soal.kisi_kisi_id` | `Soal.kisiKisiId` |
| `soal.kompetensi_id` | `Soal.kompetensiId` |
| `sod_overrides.actor_id` | `SodOverride.pelakuId` |
| `ukk_berkas.ukk_id` | `UkkBerkas.ujianUkkId` |
| `ukk_nilai.siswa_id` | `UkkNilai.siswaId` |
| `ukk_nilai.ukk_id` | `UkkNilai.ujianUkkId` |
| `ukk_sertifikat.siswa_id` | `UkkSertifikat.siswaId` |
| `ukk_sertifikat.ukk_id` | `UkkSertifikat.ujianUkkId` |
| `ukk_ujian.jurusan_id` | `UkkUjian.jurusanId` |
| `ukk_ujian.kelas_id` | `UkkUjian.kelasId` |
| `ukk_ujian.mapel_id` | `UkkUjian.mataPelajaranId` |
| `users.role -> roles.code (logis, bukan FK)` | `User.peran -> Role.kode (logis, bukan kunci tamu)` |
| `usulan_unblock.attempt_id` | `UsulanUnblock.pengerjaanUjianId` |

### Class. Dokumentasi diagram

- **Lama:** "Diagram Class Domain CBT SMK v2: 34 class domain dan 46 asosiasi yang memodelkan pengguna, master data, bank soal, paket, jadwal, attempt, jawaban, nilai, UKK, event, dan audit. Atribut memakai snake_case sesuai kolom database; himpunan atribut sama dengan kolom ERD (ERD memakai camelCase). Legenda: garis solid = asosiasi; angka di ujung asosiasi = multiplisitas, dibaca terhadap ujung tersebut (banyaknya instance di sisi itu per satu instance di sisi lawan); semua atribut bervisibilitas public."
- **Baru:** "Diagram Class Domain CBT SMK v2: 34 class domain dan 46 asosiasi yang memodelkan pengguna, master data, bank soal, paket, jadwal, pengerjaan ujian, jawaban, nilai, UKK, kegiatan, dan audit. Atribut memakai camelCase Bahasa Indonesia; himpunan atribut sama dengan atribut Model Data Logis. Tipe memakai tipe generik (String, Integer, Boolean, Date, Time, Decimal). Legenda: garis solid = asosiasi; angka di ujung asosiasi = multiplisitas, dibaca terhadap ujung tersebut (banyaknya instance di sisi itu per satu instance di sisi lawan); semua atribut bervisibilitas public."

### Yang TIDAK berubah pada Class (dijamin script + diverifikasi diff)

- Nama class, jumlah class/atribut/operasi/asosiasi, multiplisitas, ujung asosiasi, rute garis (points), dan posisi view (left/top) tetap; hanya 3 lebar kotak berubah (lihat di atas).
- Tidak ada elemen model yang ditambah/dihapus (jumlah objek JSON sama sebelum/sesudah).

## ERD / Model Data Logis — (TULIS ULANG PENUH, SELESAI)

ERD ditulis ulang penuh menjadi **Model Data Logis**: nama diagram, nama model, nama entitas, nama atribut, tipe, nama relasi, dan `referenceTo` semuanya bahasa desain. Kardinalitas + relasi dipertahankan (sudah benar per ground truth DB).

- **Entitas:** 35/35 diganti nama.
- **Atribut:** 125 nama diganti (snake_case → camelCase Bahasa Indonesia), 231/231 tipe digenerikkan (semua atribut semula bertipe DB). `referenceTo` (nama entitas induk sebagai string) ikut diganti ke nama baru.
- **Relasi:** 54/54 diganti nama (`EntitasAnak.atribut -> EntitasInduk`) + label view-nya. Kardinalitas, ujung relasi, dan rute garis TIDAK berubah.
- **Geometri:** posisi (left/top) SEMUA view TETAP. Tata letak klaster `Pengguna` (ex-`users`) TIDAK disentuh — keputusan A/B masih menunggu jawaban user. Pelebaran kotak hanya 9 EntityView yang teks barunya meluber (left/top tetap, tidak menggeser view lain): Jawaban 180→210, Pelanggaran 180→210, UsulanBukaBlokir 200→210, NilaiAkhir 180→210, Kompetensi 160→200, BankSoal 160→200, PaketUjian 180→200, UjianUkk 190→200, PengecualianPemisahanTugas 160→200.
- **Jumlah TETAP:** 35 entitas, 231 atribut, 54 relasi.

### ERD. Nama diagram + model + dokumentasi

| Elemen | Nama lama | Nama baru |
|---|---|---|
| Diagram | `Diagram ERD Skema Database CBT SMK v2 (PostgreSQL, migrasi 0001-0012)` | `Diagram Model Data Logis CBT SMK v2` |
| Model | `ERD Data Model - CBT SMK v2` | `Model Data Logis - CBT SMK v2` |

- **Dokumentasi diagram lama:** "Diagram ERD Skema Database CBT SMK v2 (PostgreSQL, migrasi 0001-0012): 35 entitas (34 tabel domain + tabel join role_permissions) dan 54 relasi foreign key. Kolom memakai camelCase; himpunan kolom sama dengan atribut Class Diagram (yang memakai snake_case). Legenda: kotak = entitas/tabel; garis = relasi foreign key dengan nama 'tabel_anak.kolom -> tabel_induk'; kardinalitas 'A : B' dibaca 'satu induk punya A baris anak; satu baris anak merujuk B induk' (B = 0..1 bila kolom FK nullable, A = 0..1 bila kolom FK UNIQUE). ERD bukan diagram UML (lihat AUDIT-INDUSTRI.md, NA-10)."
- **Dokumentasi diagram baru:** "Diagram Model Data Logis CBT SMK v2: 35 entitas dan 54 relasi. Atribut memakai camelCase Bahasa Indonesia; himpunan atribut sama dengan atribut Diagram Class. Tipe memakai tipe generik (String, Integer, Boolean, Date, Time, Decimal). Legenda: kotak = entitas; garis = relasi dengan nama 'EntitasAnak.atribut -> EntitasInduk'; kardinalitas 'A : B' dibaca 'satu induk punya A baris anak; satu baris anak merujuk B induk' (B = 0..1 bila atribut rujukan boleh kosong, A = 0..1 bila atribut rujukan UNIK). Model ini bukan diagram UML (lihat AUDIT-INDUSTRI.md, NA-10)."

### ERD. Entitas (35)

| Nama lama | Nama baru |
|---|---|
| `users` | `Pengguna` |
| `roles` | `Peran` |
| `permissions` | `Izin` |
| `role_permissions` | `PeranIzin` |
| `sessions` | `Sesi` |
| `audit_log` | `JejakAudit` |
| `settings` | `Pengaturan` |
| `jurusan` | `Jurusan` |
| `kelas` | `Kelas` |
| `mapel` | `MataPelajaran` |
| `guru` | `Guru` |
| `siswa` | `Siswa` |
| `kompetensi` | `Kompetensi` |
| `kisi_kisi` | `KisiKisi` |
| `bank_soal` | `BankSoal` |
| `soal` | `Soal` |
| `paket_ujian` | `PaketUjian` |
| `paket_soal` | `PaketSoal` |
| `jadwal_ujian` | `JadwalUjian` |
| `attempt` | `PengerjaanUjian` |
| `jawaban` | `Jawaban` |
| `pelanggaran` | `Pelanggaran` |
| `usulan_unblock` | `UsulanBukaBlokir` |
| `nilai_akhir` | `NilaiAkhir` |
| `berita_acara` | `BeritaAcara` |
| `remedial` | `Remedial` |
| `ukk_ujian` | `UjianUkk` |
| `ukk_nilai` | `NilaiUkk` |
| `ukk_berkas` | `BerkasUkk` |
| `ukk_sertifikat` | `SertifikatUkk` |
| `event` | `Kegiatan` |
| `ruangan` | `Ruangan` |
| `backup_meta` | `MetaCadangan` |
| `sod_overrides` | `PengecualianPemisahanTugas` |
| `notifikasi` | `Notifikasi` |

### ERD. Pemetaan tipe atribut

| Tipe lama (PSM) | Tipe baru (PIM) |
|---|---|
| `BIGINT` | `Integer` |
| `BIGSERIAL` | `Integer` |
| `BOOLEAN` | `Boolean` |
| `DATE` | `Date` |
| `INT` | `Integer` |
| `JSONB` | `String` |
| `NUMERIC` | `Decimal` |
| `NUMERIC(5,2)` | `Decimal` |
| `NUMERIC(8,2)` | `Decimal` |
| `TEXT` | `String` |
| `TIME` | `Time` |
| `TIMESTAMP` | `Date` |
| `TIMESTAMPTZ` | `Date` |
| `UUID` | `String` |
| `VARCHAR` | `String` |

### ERD. Atribut (hanya yang BERUBAH namanya)

#### Entitas `role_permissions` → `PeranIzin`

| Nama lama | Nama baru |
|---|---|
| `role_id` | `peranId` |
| `permission_id` | `izinId` |

#### Entitas `users` → `Pengguna`

| Nama lama | Nama baru |
|---|---|
| `username` | `namaPengguna` |
| `password_hash` | `kataSandi` |
| `role` | `peran` |
| `must_change_password` | `wajibGantiKataSandi` |
| `is_active` | `aktif` |
| `created_at` | `dibuatPada` |
| `updated_at` | `diperbaruiPada` |

#### Entitas `roles` → `Peran`

| Nama lama | Nama baru |
|---|---|
| `code` | `kode` |
| `name` | `nama` |
| `description` | `deskripsi` |
| `is_active` | `aktif` |

#### Entitas `permissions` → `Izin`

| Nama lama | Nama baru |
|---|---|
| `code` | `kode` |
| `description` | `deskripsi` |
| `group_name` | `namaKelompok` |

#### Entitas `siswa` → `Siswa`

| Nama lama | Nama baru |
|---|---|
| `user_id` | `penggunaId` |
| `kelas_id` | `kelasId` |
| `jurusan_id` | `jurusanId` |
| `tahun_masuk` | `tahunMasuk` |
| `is_active` | `aktif` |

#### Entitas `guru` → `Guru`

| Nama lama | Nama baru |
|---|---|
| `user_id` | `penggunaId` |
| `mapel_ids` | `daftarMapel` |

#### Entitas `kelas` → `Kelas`

| Nama lama | Nama baru |
|---|---|
| `jurusan_id` | `jurusanId` |
| `tahun_ajaran` | `tahunAjaran` |

#### Entitas `mapel` → `MataPelajaran`

| Nama lama | Nama baru |
|---|---|
| `jurusan_id` | `jurusanId` |
| `kelas_tingkat` | `tingkatKelas` |

#### Entitas `kompetensi` → `Kompetensi`

| Nama lama | Nama baru |
|---|---|
| `mapel_id` | `mapelId` |

#### Entitas `kisi_kisi` → `KisiKisi`

| Nama lama | Nama baru |
|---|---|
| `mapel_id` | `mapelId` |
| `kelas_tingkat` | `tingkatKelas` |
| `kompetensi_id` | `kompetensiId` |
| `jumlah_soal` | `jumlahSoal` |

#### Entitas `bank_soal` → `BankSoal`

| Nama lama | Nama baru |
|---|---|
| `mapel_id` | `mapelId` |
| `kelas_tingkat` | `tingkatKelas` |
| `guru_id` | `guruId` |
| `is_locked` | `terkunci` |

#### Entitas `soal` → `Soal`

| Nama lama | Nama baru |
|---|---|
| `bank_id` | `bankSoalId` |
| `kisi_kisi_id` | `kisiKisiId` |
| `kompetensi_id` | `kompetensiId` |
| `media_url` | `tautanMedia` |
| `level` | `tingkat` |

#### Entitas `paket_ujian` → `PaketUjian`

| Nama lama | Nama baru |
|---|---|
| `mapel_id` | `mapelId` |
| `durasi_menit` | `durasiMenit` |
| `jumlah_soal` | `jumlahSoal` |
| `acak_soal` | `acakSoal` |
| `acak_jawaban` | `acakJawaban` |
| `tampil_hasil` | `tampilHasil` |

#### Entitas `paket_soal` → `PaketSoal`

| Nama lama | Nama baru |
|---|---|
| `paket_id` | `paketUjianId` |
| `soal_id` | `soalId` |

#### Entitas `jadwal_ujian` → `JadwalUjian`

| Nama lama | Nama baru |
|---|---|
| `paket_id` | `paketUjianId` |
| `kelas_id` | `kelasId` |
| `jam_mulai` | `jamMulai` |
| `jam_selesai` | `jamSelesai` |
| `token_hash` | `token` |
| `token_expires_at` | `tokenBerlakuHingga` |
| `event_id` | `kegiatanId` |
| `ruangan_id` | `ruanganId` |

#### Entitas `attempt` → `PengerjaanUjian`

| Nama lama | Nama baru |
|---|---|
| `jadwal_id` | `jadwalUjianId` |
| `siswa_id` | `siswaId` |
| `mulai_at` | `mulaiPada` |
| `deadline` | `batasWaktu` |
| `sisa_waktu_beku` | `sisaWaktuBeku` |
| `urutan_soal` | `urutanSoal` |
| `urutan_opsi` | `urutanOpsi` |
| `submit_at` | `dikirimPada` |
| `nilai_akhir` | `nilaiAkhir` |

#### Entitas `jawaban` → `Jawaban`

| Nama lama | Nama baru |
|---|---|
| `attempt_id` | `pengerjaanUjianId` |
| `soal_id` | `soalId` |

#### Entitas `pelanggaran` → `Pelanggaran`

| Nama lama | Nama baru |
|---|---|
| `attempt_id` | `pengerjaanUjianId` |

#### Entitas `usulan_unblock` → `UsulanBukaBlokir`

| Nama lama | Nama baru |
|---|---|
| `attempt_id` | `pengerjaanUjianId` |
| `pengawas_id` | `pengawasId` |
| `diproses_oleh` | `diprosesOleh` |
| `diproses_at` | `diprosesPada` |

#### Entitas `nilai_akhir` → `NilaiAkhir`

| Nama lama | Nama baru |
|---|---|
| `attempt_id` | `pengerjaanUjianId` |
| `jadwal_id` | `jadwalUjianId` |
| `siswa_id` | `siswaId` |

#### Entitas `berita_acara` → `BeritaAcara`

| Nama lama | Nama baru |
|---|---|
| `jadwal_id` | `jadwalUjianId` |
| `dibuat_oleh` | `dibuatOleh` |

#### Entitas `remedial` → `Remedial`

| Nama lama | Nama baru |
|---|---|
| `jadwal_id` | `jadwalUjianId` |
| `siswa_id` | `siswaId` |

#### Entitas `ukk_ujian` → `UjianUkk`

| Nama lama | Nama baru |
|---|---|
| `mapel_id` | `mataPelajaranId` |
| `kelas_id` | `kelasId` |
| `jurusan_id` | `jurusanId` |
| `penguji_internal` | `pengujiInternal` |
| `penguji_eksternal` | `pengujiEksternal` |

#### Entitas `ukk_nilai` → `NilaiUkk`

| Nama lama | Nama baru |
|---|---|
| `ukk_id` | `ujianUkkId` |
| `siswa_id` | `siswaId` |
| `nilai_teori` | `nilaiTeori` |
| `nilai_proses` | `nilaiProses` |
| `nilai_produk` | `nilaiProduk` |
| `nilai_sikap` | `nilaiSikap` |
| `nilai_praktik` | `nilaiPraktik` |
| `nilai_akhir` | `nilaiAkhir` |
| `dinilai_oleh` | `dinilaiOleh` |

#### Entitas `ukk_berkas` → `BerkasUkk`

| Nama lama | Nama baru |
|---|---|
| `ukk_id` | `ujianUkkId` |
| `siswa_id` | `siswaId` |
| `nama_file` | `namaBerkas` |
| `path` | `lokasiBerkas` |
| `diupload_oleh` | `diunggahOleh` |

#### Entitas `ukk_sertifikat` → `SertifikatUkk`

| Nama lama | Nama baru |
|---|---|
| `ukk_id` | `ujianUkkId` |
| `siswa_id` | `siswaId` |
| `nomor_sertifikat` | `nomorSertifikat` |
| `diterbitkan_at` | `diterbitkanPada` |

#### Entitas `event` → `Kegiatan`

| Nama lama | Nama baru |
|---|---|
| `tanggal_mulai` | `tanggalMulai` |
| `tanggal_selesai` | `tanggalSelesai` |

#### Entitas `backup_meta` → `MetaCadangan`

| Nama lama | Nama baru |
|---|---|
| `nama_file` | `namaBerkas` |
| `ukuran_bytes` | `ukuranBita` |
| `dibuat_oleh` | `dibuatOleh` |

#### Entitas `sod_overrides` → `PengecualianPemisahanTugas`

| Nama lama | Nama baru |
|---|---|
| `actor_id` | `pelakuId` |
| `entitas_id` | `entitasId` |

#### Entitas `notifikasi` → `Notifikasi`

| Nama lama | Nama baru |
|---|---|
| `penerima_user_id` | `penerimaId` |

#### Entitas `audit_log` → `JejakAudit`

| Nama lama | Nama baru |
|---|---|
| `actor_id` | `pelakuId` |
| `entitas_id` | `entitasId` |
| `ip` | `alamatIp` |

#### Entitas `settings` → `Pengaturan`

| Nama lama | Nama baru |
|---|---|
| `key` | `kunci` |
| `value` | `nilai` |
| `group_name` | `namaKelompok` |
| `description` | `deskripsi` |

#### Entitas `sessions` → `Sesi`

| Nama lama | Nama baru |
|---|---|
| `user_id` | `penggunaId` |
| `role` | `peran` |
| `csrf_token` | `tokenSesi` |
| `ip` | `alamatIp` |
| `user_agent` | `agenPengguna` |
| `expires_at` | `berlakuHingga` |
| `last_seen_at` | `terakhirTerlihatPada` |

Atribut yang tidak tercantum di atas TIDAK berubah namanya; tipenya digenerikkan (231/231). Entitas `jurusan` dan `ruangan` tidak punya rename atribut (hanya tipe digenerikkan).

### ERD. Relasi (54)

| Nama lama | Nama baru |
|---|---|
| `sessions.user_id -> users` | `Sesi.penggunaId -> Pengguna` |
| `audit_log.actor_id -> users` | `JejakAudit.pelakuId -> Pengguna` |
| `guru.user_id -> users` | `Guru.penggunaId -> Pengguna` |
| `siswa.user_id -> users` | `Siswa.penggunaId -> Pengguna` |
| `bank_soal.guru_id -> users` | `BankSoal.guruId -> Pengguna` |
| `attempt.siswa_id -> users` | `PengerjaanUjian.siswaId -> Pengguna` |
| `usulan_unblock.pengawas_id -> users` | `UsulanBukaBlokir.pengawasId -> Pengguna` |
| `usulan_unblock.diproses_oleh -> users` | `UsulanBukaBlokir.diprosesOleh -> Pengguna` |
| `berita_acara.dibuat_oleh -> users` | `BeritaAcara.dibuatOleh -> Pengguna` |
| `ukk_nilai.dinilai_oleh -> users` | `NilaiUkk.dinilaiOleh -> Pengguna` |
| `ukk_berkas.diupload_oleh -> users` | `BerkasUkk.diunggahOleh -> Pengguna` |
| `backup_meta.dibuat_oleh -> users` | `MetaCadangan.dibuatOleh -> Pengguna` |
| `sod_overrides.actor_id -> users` | `PengecualianPemisahanTugas.pelakuId -> Pengguna` |
| `notifikasi.penerima_user_id -> users` | `Notifikasi.penerimaId -> Pengguna` |
| `role_permissions.role_id -> roles` | `PeranIzin.peranId -> Peran` |
| `role_permissions.permission_id -> permissions` | `PeranIzin.izinId -> Izin` |
| `kelas.jurusan_id -> jurusan` | `Kelas.jurusanId -> Jurusan` |
| `mapel.jurusan_id -> jurusan` | `MataPelajaran.jurusanId -> Jurusan` |
| `siswa.jurusan_id -> jurusan` | `Siswa.jurusanId -> Jurusan` |
| `ukk_ujian.jurusan_id -> jurusan` | `UjianUkk.jurusanId -> Jurusan` |
| `siswa.kelas_id -> kelas` | `Siswa.kelasId -> Kelas` |
| `jadwal_ujian.kelas_id -> kelas` | `JadwalUjian.kelasId -> Kelas` |
| `ukk_ujian.kelas_id -> kelas` | `UjianUkk.kelasId -> Kelas` |
| `kompetensi.mapel_id -> mapel` | `Kompetensi.mataPelajaranId -> MataPelajaran` |
| `kisi_kisi.mapel_id -> mapel` | `KisiKisi.mataPelajaranId -> MataPelajaran` |
| `bank_soal.mapel_id -> mapel` | `BankSoal.mataPelajaranId -> MataPelajaran` |
| `paket_ujian.mapel_id -> mapel` | `PaketUjian.mataPelajaranId -> MataPelajaran` |
| `ukk_ujian.mapel_id -> mapel` | `UjianUkk.mataPelajaranId -> MataPelajaran` |
| `nilai_akhir.siswa_id -> siswa` | `NilaiAkhir.siswaId -> Siswa` |
| `remedial.siswa_id -> siswa` | `Remedial.siswaId -> Siswa` |
| `ukk_nilai.siswa_id -> siswa` | `NilaiUkk.siswaId -> Siswa` |
| `ukk_berkas.siswa_id -> siswa` | `BerkasUkk.siswaId -> Siswa` |
| `ukk_sertifikat.siswa_id -> siswa` | `SertifikatUkk.siswaId -> Siswa` |
| `kisi_kisi.kompetensi_id -> kompetensi` | `KisiKisi.kompetensiId -> Kompetensi` |
| `soal.kompetensi_id -> kompetensi` | `Soal.kompetensiId -> Kompetensi` |
| `soal.kisi_kisi_id -> kisi_kisi` | `Soal.kisiKisiId -> KisiKisi` |
| `soal.bank_id -> bank_soal` | `Soal.bankSoalId -> BankSoal` |
| `paket_soal.soal_id -> soal` | `PaketSoal.soalId -> Soal` |
| `jawaban.soal_id -> soal` | `Jawaban.soalId -> Soal` |
| `paket_soal.paket_id -> paket_ujian` | `PaketSoal.paketUjianId -> PaketUjian` |
| `jadwal_ujian.paket_id -> paket_ujian` | `JadwalUjian.paketUjianId -> PaketUjian` |
| `attempt.jadwal_id -> jadwal_ujian` | `PengerjaanUjian.jadwalUjianId -> JadwalUjian` |
| `nilai_akhir.jadwal_id -> jadwal_ujian` | `NilaiAkhir.jadwalUjianId -> JadwalUjian` |
| `berita_acara.jadwal_id -> jadwal_ujian` | `BeritaAcara.jadwalUjianId -> JadwalUjian` |
| `remedial.jadwal_id -> jadwal_ujian` | `Remedial.jadwalUjianId -> JadwalUjian` |
| `jawaban.attempt_id -> attempt` | `Jawaban.pengerjaanUjianId -> PengerjaanUjian` |
| `pelanggaran.attempt_id -> attempt` | `Pelanggaran.pengerjaanUjianId -> PengerjaanUjian` |
| `usulan_unblock.attempt_id -> attempt` | `UsulanBukaBlokir.pengerjaanUjianId -> PengerjaanUjian` |
| `nilai_akhir.attempt_id -> attempt` | `NilaiAkhir.pengerjaanUjianId -> PengerjaanUjian` |
| `ukk_nilai.ukk_id -> ukk_ujian` | `NilaiUkk.ujianUkkId -> UjianUkk` |
| `ukk_berkas.ukk_id -> ukk_ujian` | `BerkasUkk.ujianUkkId -> UjianUkk` |
| `ukk_sertifikat.ukk_id -> ukk_ujian` | `SertifikatUkk.ujianUkkId -> UjianUkk` |
| `jadwal_ujian.event_id -> event` | `JadwalUjian.kegiatanId -> Kegiatan` |
| `jadwal_ujian.ruangan_id -> ruangan` | `JadwalUjian.ruanganId -> Ruangan` |

### Yang TIDAK berubah pada ERD (dijamin script + diverifikasi diff)

- Jumlah entitas/atribut/relasi, kardinalitas (`cardinality` relasi + ends), ujung relasi (`reference`), `identifying`, flag atribut (`primaryKey`, `foreignKey`, `nullable`, `unique`), rute garis (points), dan posisi view (left/top) tetap; hanya 9 lebar kotak berubah (lihat di atas).
- Tidak ada elemen model yang ditambah/dihapus (jumlah objek JSON sama sebelum/sesudah).
- Tata letak klaster `Pengguna` TIDAK disentuh (keputusan A/B masih menunggu user).

## Addendum Worker 4 — Penutupan Celah Sequence + Perbaikan Label (18 Okt 2026)

Perubahan ini dilakukan Worker 4 (validasi + verifikasi visual + ekspor) setelah
inspeksi visual ekspor StarUML membuktikan celah bekas lifeline fungsi di kedua
diagram Sequence terlihat jelas seperti lifeline hilang (jarak antar lifeline
2× lipat di tengah diagram). File final SHA-256:
`c5bd08be78b5e8b4d739231352b9bf6b5cc834c9627070682f2db7a9f737a016`
(sebelum perubahan: `82b6c940f28297c30642eab72d1e1c6523245bb48584509551def7241de9c24c`).

- **3a (Diagram Sequence Login + Wajib Ganti Password):** lifeline Basis Data
  `left` 760→595; 5 `points` pesan yang menyentuh Basis Data digeser −165.
- **3b (Diagram Sequence Simpan Jawaban + Auto-Submit):** lifeline Penghitung
  Nilai 760→595, Basis Data 925→760, Layanan Pemantau 1090→925; 9 `points`
  pesan yang menyentuh ketiganya digeser −165.
- **Total diff:** 18 daun JSON (4 `left` lifeline + 14 `points` pesan). Tidak ada
  perubahan semantik: nama pesan, `messageSort`, source/target, nilai Y, kotak
  kepala lifeline, label pesan, dan diagram lain tidak disentuh. Label
  (EdgeLabelView) tidak disentuh — posisinya dihitung dari
  `(edgePosition, distance, alpha)` + titik edge ter-render.
- **Perbaikan label lanjutan 3a:** setelah celah ditutup, label self-message
  `verifikasiKataSandi()` terpotong tepi kanan kanvas (kanvas menyempit,
  label punya `distance` 158). Diperbaiki `distance` 158→10 mengikuti preseden
  perbaikan label 3b; label kini tampil utuh. 3b tidak diubah (labelnya sudah
  tampil utuh).
- **Pitch lifeline** kini seragam 165 di kedua diagram (3a: 170/335/500/665;
  3b: 170/335/500/665/830/995). Skrip idempotent (jalan kedua = 0 perubahan).
- **Verifikasi ulang penuh sesudah perubahan:** validator 0 pelanggaran
  (LOLOS SEMUA), pindai PSM 0 temuan, re-ekspor SVG+PNG 2× dari file final
  dan inspeksi visual 3a/3b (celah tertutup, semua pesan berakhir di lifeline
  yang benar, self-message + Note utuh).
- Klaster `Pengguna` ERD tetap TIDAK disentuh (keputusan A/B menunggu user).
