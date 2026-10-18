# DIAGNOSIS KEJELASAN SEQUENCE 3a + 3b — CBT SMK v2

- **File:** `docs/uml/cbt-smk-v2.mdj`
- **SHA-256 awal (dicek):** `c068a179be3df130e566dd26ea826c27a4368e09463d2cf5937eae0dee1025df` — COCOK, diagnosis dilanjutkan.
- **SHA-256 akhir (setelah diagnosis):** `c068a179…` (sama) — file `.mdj` proyek **TIDAK diubah**. Hanya salinan `/tmp/seq-diagnosis-copy.mdj` yang dibuka di StarUML; StarUML dimatikan tanpa save.
- **Tanggal:** 18 Okt 2026.
- **Metode:** (1) analisis JSON `.mdj` langsung; (2) analisis SVG ekspor StarUML `your_files/uml-svg/06-sequence-3a.svg` + `07-sequence-3b.svg` (teks `<text>` + posisi); (3) verifikasi **live canvas StarUML asli 7.1.0** (Xvfb, salinan, read-only) + screenshot + pengukuran piksel.
- **Kalibrasi yang dipakai (terbukti ronde sebelumnya, tidak diulang):** posisi label render ditentukan `(edgePosition, distance, alpha)` + titik edge tergambar; bentang panah = celah antar kotak kepala lifeline, `points` X pesan diabaikan (Y = `points[0].y`); self-message selalu render terakhir; lebar kotak kepala render = `max(width, fit-teks)`; guard dirender `"[" + text + "]"`.

## Ringkasan verdict

| # | Butir | Verdict | Inti |
|---|-------|---------|------|
| 1 | Penomoran pesan | **TERBUKTI, dengan nuansa kunci** | Klaim `1,2,3,0,4,6,0,8,7,10,0,12,9,5,13,11` **benar untuk EKSPOR** (SVG/PNG). Tapi **live canvas menampilkan 1..16 sekuensial**. Angka TIDAK tersimpan di file; dihitung StarUML. |
| 2 | Self-message | **TERBUKTI** | 3a `verify Argon2id…` (Y=415) dan 3b `cek deadline lewat…` (Y=815) render **paling bawah** (live + ekspor), bernomor 16, loop-nya **melayang** ~30px dari garis lifeline, labelnya menyeberangi garis lifeline lain. |
| 3 | Label panjang nyeberang garis lifeline | **TERBUKTI** | 15/16 label 3a dan 13/16 label 3b (teks terukur) menyeberangi ≥1 garis putus-putus lifeline. Penyebab: label terpusat (centered) + lebar teks (s.d. ~440px) > pitch lifeline (165px). TIDAK ada label menimpa label lain / kotak kepala / panah tetangga. |
| 4 | Arah panah | **TIDAK TERBUKTI ada yang salah** | Semua 32 pesan: source/target sesuai semantik, `messageSort` benar, render sesuai sort (solid+isi / putus-putus+terbuka / solid+terbuka). |
| 5 | Jarak vertikal | **TIDAK TERBUKTI ada penumpukan / celah raksasa** | Gap min 60px, maks 120px (3a, pemisah fase). Tidak ada pesan bertumpuk. "Celah kosong" di tengah adalah slot self-message yang render-nya pindah ke bawah (lihat butir 2). |
| 6 | Celah panah–lifeline | **TIDAK TERBUKTI ada celah bermasalah** | Semua ujung panah tepat di garis putus-putus (celah 0). Kotak kepala seragam 140px, pitch 165px, celah antar kotak 25px, `wordWrap=true`, tidak ada pelebaran. Satu-satunya anomali celah adalah loop self-message yang melayang (butir 2). |

**Kesimpulan utama:** yang membuat Sequence "gak jelas" BUKAN isi modelnya (modelnya benar: arah, sort, urutan Y semua benar), melainkan **tiga hal render**: (1) penomoran ekspor yang acak + `0 :` — inilah yang user lihat di PNG/SVG; (2) self-message yang render-nya terlempar ke bawah + loop melayang + label menyeberang; (3) label-label panjang yang menyeberangi garis lifeline (faktor pendukung, standar pada diagram padat). Rencana perbaikan di bawah menargetkan ketiganya.

---

## 1. Penomoran pesan — TERBUKTI (untuk ekspor); live canvas SEKUENSIAL

### 1.1 Angka TIDAK ada di file

- Field `UMLMessage` hanya: `_id, _parent, _type, messageSort, name, source, target`. **Tidak ada field nomor** di pesan maupun di `UMLSeqMessageView`.
- Objek `UMLSequenceDiagram` hanya punya kunci `_id, _parent, _type, documentation, name, ownedViews` — **tidak ada field penomoran tersimpan**.
- Panel Properties StarUML untuk diagram menunjukkan: `showSequenceNumber = true` (tercentang), `sequenceNumbering = auto`, `showSignature = true`, `showActivation = true`. Itu **default StarUML**, bukan nilai tersimpan di `.mdj`.
- Nama pesan di model TIDAK mengandung angka (mis. `return user + password_hash + must_change_password`, bukan `0 : return…`). Prefix `N : ` ditambahkan StarUML saat render.

### 1.2 Ekspor (SVG) — klaim awal TERBUKTI PERSIS

Urutan baca **sesuai posisi render atas-ke-bawah** (self-message dihitung di posisi render-nya, yaitu paling bawah):

- **3a ekspor:** `1, 2, 3, 0, 4, 6, 0, 8, 7, 10, 0, 12, 9, 5, 13, 11` — **identik dengan klaim awal**, termasuk tiga `0 :` dan self-message `11` di akhir.
- **3b ekspor:** `1, 2, 3, 0, 4, 5, 6, 0, 8, 10, 0, 9, 12, 11, 7, 13` (self-message `13` di akhir; tiga `0 :`).

Pesan yang diekspor sebagai `0 : …` (tiga di tiap diagram):

| Diagram | Pesan `0 : …` (Y) | `messageSort` |
|---|---|---|
| 3a | `return user + password_hash + must_change_password` (Y=335) | reply |
| 3a | `isi password_lama + password_baru` (Y=635) | synchCall |
| 3a | `200 password berhasil diubah` (Y=915) | reply |
| 3b | `OK - jawaban tersimpan (autosave)` (Y=375) | reply |
| 3b | `deadline + status` (Y=695) | reply |
| 3b | `poin + nilai sementara` (Y=975) | reply |

Tidak ada pola `messageSort` yang menjelaskan `0` (5 dari 6 adalah reply, tapi reply lain bernomor: 3a `403`=4, `OK`=10, `200 + session…`=5; 3b `200 jawaban…`=4, `sisa waktu…`=8, `OK - attempt…`=12, `200 hasil…`=7). Algoritma penomoran exporter tidak dapat diturunkan dari isi file — yang terbukti adalah **ekspor ≠ live canvas**. (Jangan berasumsi exporter memakai urutan Y / urutan model / urutan view — ketiganya sama dengan urutan Y, dan ketiganya TIDAK menghasilkan angka ekspor.)

### 1.3 Live canvas StarUML 7.1.0 — SEKUENSIAL, bukan acak

Dibuka salinan di StarUML asli, screenshot + dibaca langsung:

- **3a live:** `1` isi… (Y=135), `2` POST login (195), `3` SELECT (255), `4` return (335), `5` 403 (495), `6` tampilkan (575), `7` isi password_lama (635), `8` POST change-password (695), `9` UPDATE (775), `10` OK (855), `11` 200 password berhasil (915), `12` POST login ulang (1035), `13` INSERT sessions (1095), `14` 200 + session cookie (1155), `15` redirect (1215), **`16` verify Argon2id (render paling bawah)**.
- **3b live:** `1` pilih (135), `2` POST jawaban (215), `3` INSERT/UPDATE (295), `4` OK autosave (375), `5` 200 jawaban (435), `6` GET status (535), `7` SELECT deadline (635), `8` deadline + status (695), `9` sisa waktu (755), `10` hitungNilai (895), `11` poin (975), `12` UPDATE (1055), `13` OK auto-submit (1135), `14` broadcastAttempt (1215), `15` 200 hasil (1295), **`16` cek deadline lewat (render paling bawah)**.

Jadi: **live canvas menomori 1..15 sesuai urutan Y + self-message = 16 di bawah**. Yang rusak hanya **jalur ekspor** (SVG/PNG) — dan semua artefak kirim (`uml-svg/`, `uml-png/` yang di-render dari SVG via cairosvg) memakai jalur ekspor itu. Inilah yang user lihat sebagai "garis-garisnya gak jelas".

### 1.4 Tabel lengkap pesan 3a (urut Y)

`source`/`target` = nama lifeline. `Eks` = nomor di SVG ekspor (urutan Y; self di slot Y-nya). `Live` = nomor di live canvas.

| # | id | name | source → target | Y | messageSort | Eks | Live |
|---|---|---|---|---|---|---|---|
| 1 | `YQNHDHOOX2F1T0OSBEBTEN` | `isi username/email + password` | Siswa / Pengguna → LoginPage / GantiPasswordPage (frontend) | 135 | synchCall | 1 | 1 |
| 2 | `R255F0YT8R3G4D5S0KDSKD` | `POST /api/v1/auth/login` | LoginPage / GantiPasswordPage (frontend) → Auth API (auth.go) | 195 | synchCall | 2 | 2 |
| 3 | `CWYGPRXTUIRDU4SLLB5CEH` | `SELECT users WHERE username/email` | Auth API (auth.go) → PostgreSQL: users, sessions | 255 | synchCall | 3 | 3 |
| 4 | `791UCXP7G8J2YMD8YB6QGQ` | `return user + password_hash + must_change_password` | PostgreSQL: users, sessions → Auth API (auth.go) | 335 | reply | 0 | 4 |
| 5 | `VJ3PTSKMGLJDE7066LX9XD` | `verify Argon2id(password, password_hash)` | Auth API (auth.go) → Auth API (auth.go) **[SELF]** | 415 | synchCall | 11 | 16 |
| 6 | `ESA1WPD19JBB1LT21XKLRU` | `403 {error.code: PASSWORD_CHANGE_REQUIRED}` | Auth API (auth.go) → LoginPage / GantiPasswordPage (frontend) | 495 | reply | 4 | 5 |
| 7 | `10CDUL8MY71JU572UQUQGA` | `tampilkan form Ganti Password` | LoginPage / GantiPasswordPage (frontend) → Siswa / Pengguna | 575 | synchCall | 6 | 6 |
| 8 | `CIC5U9RO4BW0AC06K8V732` | `isi password_lama + password_baru` | Siswa / Pengguna → LoginPage / GantiPasswordPage (frontend) | 635 | synchCall | 0 | 7 |
| 9 | `ISWTLMSHHV1GLH3NHN7UTI` | `POST /api/v1/auth/change-password` | LoginPage / GantiPasswordPage (frontend) → Auth API (auth.go) | 695 | synchCall | 8 | 8 |
| 10 | `GXJ9BR7VJ3KW3A24P12GSJ` | `UPDATE users SET password_hash, must_change_password=FALSE` (nilai flag FALSE terlihat di live canvas) | Auth API (auth.go) → PostgreSQL: users, sessions | 775 | synchCall | 7 | 9 |
| 11 | `XGMXN3C6NPOJQ6U67A972A` | `OK` | PostgreSQL: users, sessions → Auth API (auth.go) | 855 | reply | 10 | 10 |
| 12 | `OMG3AREPLY000000000001` | `200 password berhasil diubah` | Auth API (auth.go) → LoginPage / GantiPasswordPage (frontend) | 915 | reply | 0 | 11 |
| 13 | `FWKXGK8OL28ZUHAZ9NXDY1` | `POST /api/v1/auth/login (ulang)` | LoginPage / GantiPasswordPage (frontend) → Auth API (auth.go) | 1035 | synchCall | 12 | 12 |
| 14 | `3K28HTO8GAX2DQM9BB18AJ` | `INSERT sessions (csrf_token, expires_at)` | Auth API (auth.go) → PostgreSQL: users, sessions | 1095 | synchCall | 9 | 13 |
| 15 | `LQ7JXF9FGHNPUTLZBU0SHV` | `200 + session cookie + CSRF token` | Auth API (auth.go) → LoginPage / GantiPasswordPage (frontend) | 1155 | reply | 5 | 14 |
| 16 | `73H8P8V1J4NAZGXITEQR9G` | `redirect ke dashboard / ujian` | LoginPage / GantiPasswordPage (frontend) → Siswa / Pengguna | 1215 | synchCall | 13 | 15 |

Catatan: urutan `ownedElements` model 3a = persis urutan Y di atas (model order == Y order == urutan view). Jadi penomoran ekspor yang acak TIDAK berasal dari urutan apa pun di file.

### 1.5 Tabel lengkap pesan 3b (urut Y)

| # | id | name | source → target | Y | messageSort | Eks | Live |
|---|---|---|---|---|---|---|---|
| 1 | `F27TRO9JN9XZTNU3BOY03B` | `pilih / ubah jawaban soal` | Siswa → UjianPage (frontend) | 135 | synchCall | 1 | 1 |
| 2 | `O6W055NXW2H23RAJ1B6DG9` | `POST /api/v1/attempt/:id/jawaban (autosave)` | UjianPage (frontend) → Exam API (exam.go) | 215 | synchCall | 2 | 2 |
| 3 | `NCJPMHZJA578H516CH8MV0` | `INSERT/UPDATE jawaban (jawaban JSONB, ragu, poin)` | Exam API (exam.go) → PostgreSQL: jawaban, attempt, nilai_akhir | 295 | synchCall | 3 | 3 |
| 4 | `NN7Y13HPAQEAVXSAA4M357` | `OK - jawaban tersimpan (autosave)` | PostgreSQL: jawaban, attempt, nilai_akhir → Exam API (exam.go) | 375 | reply | 0 | 4 |
| 5 | `GG8QUYNS8L4IUJMDHE50UB` | `200 jawaban tersimpan` | Exam API (exam.go) → UjianPage (frontend) | 435 | reply | 4 | 5 |
| 6 | `VHQ7JKIIZU2J5FW9RGLO66` | `GET status attempt (timer berjangkar server: deadline, sisa_waktu)` | UjianPage (frontend) → Exam API (exam.go) | 535 | synchCall | 5 | 6 |
| 7 | `5ATD4TPBNQZENRVX5QL9JO` | `SELECT attempt.deadline, jadwal.status` | Exam API (exam.go) → PostgreSQL: jawaban, attempt, nilai_akhir | 635 | synchCall | 6 | 7 |
| 8 | `HEBOOVV5V2J56IOSWXUDUI` | `deadline + status` | PostgreSQL: jawaban, attempt, nilai_akhir → Exam API (exam.go) | 695 | reply | 0 | 8 |
| 9 | `YAHGEKOYA1E9D84DXJZIKS` | `sisa waktu server` | Exam API (exam.go) → UjianPage (frontend) | 755 | reply | 8 | 9 |
| 10 | `UN7WCVBZWOTRUHY9NVOKV0` | `cek deadline lewat (autoSubmitJikaLewat)` | Exam API (exam.go) → Exam API (exam.go) **[SELF]** | 815 | synchCall | 13 | 16 |
| 11 | `S13XJD0PYY3LC3SLMMJUA6` | `hitungNilai: HitungPoin per soal (otomatis)` | Exam API (exam.go) → Grading (grading.go) | 895 | synchCall | 10 | 10 |
| 12 | `79ANOJ13H94FSVYSDK1WB6` | `poin + nilai sementara` | Grading (grading.go) → Exam API (exam.go) | 975 | reply | 0 | 11 |
| 13 | `656E4L0YXV6X7O1MM47CUO` | `UPDATE attempt status=selesai, submit_at; syncNilaiAkhir` | Exam API (exam.go) → PostgreSQL: jawaban, attempt, nilai_akhir | 1055 | synchCall | 9 | 12 |
| 14 | `QWMU4LYKY0GJJS006FUYIY` | `OK - attempt selesai tersimpan (auto-submit)` | PostgreSQL: jawaban, attempt, nilai_akhir → Exam API (exam.go) | 1135 | reply | 12 | 13 |
| 15 | `4RCWWVZ8OW9Y2YGD4HVWZD` | `broadcastAttempt (WS: status peserta berubah)` | Exam API (exam.go) → Monitor WebSocket | 1215 | asynchCall | 11 | 14 |
| 16 | `L2KM2C81GME4FTVU3MCM5O` | `200 hasil + nilai (final bila tanpa esai)` | Exam API (exam.go) → UjianPage (frontend) | 1295 | reply | 7 | 15 |

---

## 2. Self-message — TERBUKTI

### 2.1 Identifikasi (source == target, terbukti dari JSON)

| Diagram | Pesan | id model / id view | source == target | Y (slot semantik) | Posisi render |
|---|---|---|---|---|---|
| 3a | `verify Argon2id(password, password_hash)` | model `VJ3PTSKMGLJDE7066LX9XD` / view `ETPM0NI326GSBWAP3WBE88` | Auth API (auth.go) == Auth API (auth.go) | 415 (ke-5 dari 16) | **Paling bawah**, setelah `redirect` (Y=1215) |
| 3b | `cek deadline lewat (autoSubmitJikaLewat)` | model `UN7WCVBZWOTRUHY9NVOKV0` / view `MXGI09OCO1G45WO9CDLMVS` | Exam API (exam.go) == Exam API (exam.go) | 815 (ke-10 dari 16) | **Paling bawah**, setelah `200 hasil…` (Y=1295) |

Keduanya `messageSort = synchCall`. Keduanya TERBUKTI render terakhir di **kedua jalur** (live canvas + ekspor SVG): label `16 : …` di bawah seluruh diagram; di ekspor 3a bernomor 11, 3b bernomor 13 (nomor ekspor acak, lihat butir 1).

### 2.2 Tiga cacat render self-message (terukur di live canvas 3a, berlaku sama 3b)

1. **Slot kosong di tengah.** Di posisi semantiknya (3a: antara `return` Y=335 dan `403` Y=495; 3b: antara `sisa waktu server` Y=755 dan `hitungNilai` Y=895) tidak ada panah apa pun — terbaca sebagai pesan hilang / celah misterius.
2. **Loop melayang, tidak menyentuh lifeline.** Loop tersimpan `points = 500:415;530:415;530:435;500:435` (keluar 30px ke kanan lifeline, turun 20px, kembali). Di live canvas, loop ter-render sebagai bentuk ⊐ kecil yang **melayang ~30px di kanan garis putus-putus lifeline** (pengukuran piksel screenshot 3a: garis Auth di x=701, loop di x≈735–765; ujung panah menunjuk ke kiri ke ruang kosong, tidak menyentuh garis). Penyebabnya perilaku render StarUML, bukan `points` (titik X diabaikan untuk pesan biasa; untuk self-message geometri dihitung ulang).
3. **Label jauh dari loop + menyeberangi garis lifeline lain.** Label 3a ter-render di span teks terukur [623, 898] (pusat 757, lebar teks ~275px untuk 45 karakter, ~6,26 px/karakter pada Arial 13): mulai **~93px di kanan ujung loop (530)** dan **menyeberangi garis PostgreSQL (665)** — garis putus-putus PostgreSQL memotong teks label (terlihat di screenshot: garis melewati teks `verify Argon2id…`). Label 3b ter-render di span yang sama [616, 898] dan **menyeberangi DUA garis: Grading (665) dan PostgreSQL (830)**.

### 2.3 Opsi perbaikan + rekomendasi

**(a) Terima posisi render terakhir + nomori sesuai posisi render + tambah catatan.**
- Realita: StarUML SUDAH melakukan ini di live canvas (nomor 16, posisi bawah). Tidak ada yang perlu diubah pada pesan itu sendiri.
- **Konflik penomoran:** bila penomoran mengikuti posisi render (16), nomor TIDAK lagi menyatakan urutan semantik (semantiknya langkah ke-5 di 3a / ke-10 di 3b). Konflik ini diselesaikan dengan **memindahkan informasi urutan semantik keluar dari nomor**: nomor = urutan render (apa yang StarUML hitung), urutan semantik dinyatakan di **Note + dokumentasi diagram** dengan merujuk NAMA pesan (bukan nomor — apalagi bila penomoran dimatikan, lihat Rencana P0). Contoh teks Note 3a: *"Catatan: 'verify Argon2id…' secara semantik terjadi setelah 'return user + …' dan sebelum '403 …'. StarUML selalu me-render self-message di bawah diagram."*
- Kelemahan (a): diagram tetap menyimpan elemen yang membingungkan (slot kosong di tengah + loop melayang + label menyeberang). Pembaca harus membaca catatan untuk tidak salah paham.

**(b) Ubah jadi pesan ke lifeline terpisah (MENGUBAH model).**
- 3a: tambah lifeline `Argon2id (library)`; `verify…` menjadi pesan biasa Auth API → Argon2id, di Y=415 (posisi semantik benar), panah normal, label normal di atas panah.
- 3b: tambah lifeline `autoSubmitJikaLewat (exam.go)`; `cek deadline lewat…` menjadi pesan biasa Exam API → lifeline baru, di Y=815.
- Ini BUKAN pemalsuan model: kode benar-benar memanggil fungsi/pustaka yang berbeda (`verify Argon2id` = verifikasi kriptografi; `cek deadline lewat` = fungsi auto-submit internal). Mendekomposisi self-call menjadi panggilan ke lifeline helper adalah praktik UML yang sah dan umum.
- Kelemahan (b): mengubah model (tambah 1 lifeline per diagram + geser lifeline di kanannya + sesuaikan `points` + label pesan-pesan yang menyentuh lifeline tergeser — daftar lengkap di Rencana P1). Perlu persetujuan user.

**REKOMENDASI: opsi (b) sebagai utama; opsi (a) sebagai fallback bila user menolak perubahan model.**
Alasan: (b) adalah SATU-SATUNYA opsi yang membuat **urutan visual = urutan semantik = urutan baca**, tanpa catatan kaki, tanpa elemen melayang, tanpa label menyeberang ekstra. (a) mempertahankan elemen yang terbukti membingungkan dan hanya menambal dengan catatan. Untuk skripsi — di mana penguji membaca diagram tanpa didampingi penjelasan — (b) jauh lebih aman. Jika user menolak (b), jalankan (a) lengkap dengan Note + pembaruan dokumentasi (spesifikasi di Rencana P1-fallback).

Varian (a2) — ganti self-message dengan Note di lifeline pada Y semantiknya ("verifikasi Argon2id di sini") — TIDAK direkomendasikan: menghilangkan pesan dari interaction (mengubah hitungan pesan 16→15) tanpa memberi panah apa pun; lebih buruk daripada (a) dan (b).

---

## 3. Label panjang menyeberangi garis lifeline — TERBUKTI

### 3.1 Metode ukur

- Pusat label = pusat `EdgeLabelView` tersimpan (tervalidasi: pusat tersimpan vs pusat teks SVG ekspor selisih ≤2px untuk SEMUA pesan non-self).
- Lebar teks terukur dari screenshot live: label 45 karakter → ~282px ⇒ **~6,26 px/karakter** (Arial 13). Span teks = pusat ± (panjang teks × 6,26 / 2), teks = `"N : " + nama` dengan N = nomor live.
- Garis putus-putus: 3a di x = 170, 335, 500, 665. 3b di x = 170, 335, 500, 665, 830, 995.
- "Menyeberang" = garis lifeline berada ketat di dalam span teks. Kasus "menyentuh" (garis dalam ±3px tepi span) dicatat sebagai grazing.

### 3.2 Per label (3a) — 15/16 menyeberang

| Live# | Pesan (disingkat) | Span teks | Garis yang diseberangi |
|---|---|---|---|
| 1 | isi username/email + password | [149, 356] | Siswa@170, LoginPage@335 |
| 2 | POST /api/v1/auth/login | [334, 502] | LoginPage@335 (grazing, tepi kiri), Auth@500 |
| 3 | SELECT users WHERE username/email | [467, 698] | Auth@500, PostgreSQL@665 |
| 4 | return user + … | [414, 751] | Auth@500, PostgreSQL@665 |
| 16 | verify Argon2id… (self) | [616, 898] | PostgreSQL@665 |
| 5 | 403 {error.code: …} | [274, 561] | LoginPage@335, Auth@500 |
| 6 | tampilkan form Ganti Password | [149, 356] | Siswa@170, LoginPage@335 |
| 7 | isi password_lama + password_baru | [137, 368] | Siswa@170, LoginPage@335 |
| 8 | POST /api/v1/auth/change-password | [302, 533] | LoginPage@335, Auth@500 |
| 9 | UPDATE users SET … | [389, 776] | Auth@500, PostgreSQL@665 |
| 10 | OK | [561, 604] | — (TIDAK menyeberang) |
| 11 | 200 password berhasil diubah | [315, 521] | LoginPage@335, Auth@500 |
| 12 | POST /api/v1/auth/login (ulang) | [305, 531] | LoginPage@335, Auth@500 |
| 13 | INSERT sessions (…) | [442, 724] | Auth@500, PostgreSQL@665 |
| 14 | 200 + session cookie + CSRF token | [299, 536] | LoginPage@335, Auth@500 |
| 15 | redirect ke dashboard / ujian | [146, 359] | Siswa@170, LoginPage@335 |

### 3.3 Per label (3b) — 13/16 menyeberang jelas + 1 grazing

| Live# | Pesan (disingkat) | Span teks | Garis yang diseberangi |
|---|---|---|---|
| 1 | pilih / ubah jawaban soal | [162, 343] | Siswa@170, UjianPage@335 |
| 2 | POST …/jawaban (autosave) | [271, 565] | UjianPage@335, Exam@500 |
| 3 | INSERT/UPDATE jawaban (…) | [500, 831] | Exam@500 (grazing), Grading@665, PostgreSQL@830 |
| 4 | OK - jawaban tersimpan (autosave) | [550, 781] | Grading@665 |
| 5 | 200 jawaban tersimpan | [339, 496] | — (TIDAK) |
| 6 | GET status attempt (…) | [199, 636] | UjianPage@335, Exam@500 |
| 7 | SELECT attempt.deadline, jadwal.status | [534, 797] | Grading@665 |
| 8 | deadline + status | [600, 731] | Grading@665 |
| 9 | sisa waktu server | [352, 483] | — (TIDAK) |
| 16 | cek deadline lewat… (self) | [616, 898] | Grading@665, PostgreSQL@830 |
| 10 | hitungNilai: HitungPoin… | [433, 733] | Exam@500, Grading@665 |
| 11 | poin + nilai sementara | [498, 667] | Exam@500 (grazing), Grading@665 (grazing) |
| 12 | UPDATE attempt status=selesai… | [474, 856] | Exam@500, Grading@665, PostgreSQL@830 |
| 13 | OK - attempt selesai tersimpan… | [512, 818] | Grading@665 |
| 14 | broadcastAttempt (WS: …) | [591, 904] | Grading@665, PostgreSQL@830 |
| 15 | 200 hasil + nilai (…) | [274, 561] | UjianPage@335, Exam@500 |

### 3.4 Yang TIDAK terjadi (sama pentingnya)

- **Tidak ada label menimpa label lain.** Interval vertikal label (tinggi 13px, offset ±16px dari panahnya) tidak pernah tumpang tindih antar pesan — jarak antar panah min 60px.
- **Tidak ada label menimpa kotak kepala lifeline.** Kotak kepala berakhir di y=78–110 (tergantung lifeline); label pertama mulai y=119.
- **Tidak ada label menimpa panah tetangga.** Label pesan L→R duduk di atas panahnya, label R→L di bawah panahnya; panah tetangga berjarak ≥60px.
- **Tidak ada label menimpa loop self-message.** Label self berjarak ~93px dari ujung loop.

### 3.5 Akar masalah

Label StarUML dipusatkan pada titik tengah panah (`edgePosition=1`, `alpha=π/2`, `distance=10` untuk SEMUA pesan non-self). Panah antar lifeline tetangga hanya selebar pitch = 165px, sedangkan 28 dari 30 label non-self lebih lebar dari itu (teks 130–440px). **Label yang lebih lebar dari panahnya PASTI menyeberangi garis lifeline** — ini konsekuensi matematis dari desain terpusat + pitch 165px, bukan kesalahan satu-dua label. Label 2-hop (3b: Exam→PostgreSQL, span 330px) tetap menyeberang karena labelnya 260–380px.

---

## 4. Arah panah — TIDAK TERBUKTI ada yang salah

Verifikasi per pesan (32/32): `source`/`target` di JSON cocok dengan semantik narasi (mis. `GET status attempt` = UjianPage → Exam API, bukan sebaliknya; `SELECT…` = Exam API → PostgreSQL; semua reply mengarah kembali ke pemanggilnya; `broadcastAttempt` = Exam API → Monitor WebSocket; `redirect` = LoginPage → Siswa). `messageSort` benar untuk semuanya: 26× `synchCall`, 5× `reply` (3a) + 5× `reply` (3b), 1× `asynchCall` (broadcast 3b), 2× `synchCall` self. Render di live canvas sesuai sort: synchCall = garis solid + kepala panah terisi (termasuk `tampilkan form`, `pilih / ubah jawaban`, `redirect`); reply = garis putus-putus + kepala terbuka (terlihat pada `return`, `403`, `OK`, `200…`, `deadline + status`, `poin…`); asynchCall = garis solid + kepala terbuka (`broadcastAttempt`, terlihat di live 3b). **Tidak ada panah terbalik, tidak ada sort yang salah.** Tidak ada perbaikan pada butir ini.

---

## 5. Jarak vertikal — TIDAK TERBUKTI ada penumpukan / celah raksasa

Gap Y antar pesan berurutan (px):

- **3a:** 60, 60, 80, 80, 80, 80, 60, 60, 80, 80, 60, **120**, 60, 60, 60. Min 60, maks 120.
- **3b:** 80, 80, 80, 60, 100, 100, 60, 60, 60, 80, 80, 80, 80, 80, 80. Min 60, maks 100.

Tidak ada pesan bertumpuk (gap min 60px ≫ tinggi label 13px + offset). Gap 120px di 3a (antara `200 password berhasil diubah` Y=915 dan `POST /api/v1/auth/login (ulang)` Y=1035) dan gap 100px di 3b (antara `200 jawaban tersimpan` Y=435 dan `GET status attempt` Y=535; antara `GET…` Y=535 dan `SELECT…` Y=635) adalah **pemisah fase** yang wajar, bukan cacat. Satu-satunya "celah kosong" yang menyesatkan adalah slot self-message (butir 2.2 poin 1): total jarak `return`→`403` = 160px dengan slot Y=415 kosong di live canvas. Itu gejala butir 2, bukan masalah jarak.

---

## 6. Celah panah–lifeline — TIDAK TERBUKTI ada celah bermasalah

- **Ujung panah tepat di garis lifeline (celah 0)** untuk 32/32 pesan: setiap `points` dimulai/berakhir tepat pada x garis putus-putus (mis. 3a `170:135;335:135`, `665:335;500:335`; 3b `500:295;830:295`, `500:1215;995:1215`). Terverifikasi programatis: `pts-endpoints-on-dashed = [True, True]` untuk semua pesan.
- **Kotak kepala:** semua lifeline tersimpan `width = 140`, `top = 20`; pitch seragam **165px** (3a: 100→265→430→595; 3b: 100→265→430→595→760→925); **celah antar kotak = 25px** merata. `wordWrap = true` pada semua lifeline; nama terpanjang (`LoginPage / GantiPasswordPage (frontend)`, 38 karakter; `PostgreSQL: jawaban, attempt, nilai_akhir`, 41 karakter) membungkus jadi 2–3 baris (tinggi kompartemen 58–90px) — **tidak ada pelebaran kotak** (lebar render = 140 = lebar tersimpan). Terkonfirmasi di screenshot live: semua kotak kepala sama lebar.
- **Tidak ada label menimpa kotak kepala** (butir 3.4).
- Satu-satunya anomali "celah" adalah loop self-message yang melayang ~30px dari garis lifelinenya (butir 2.2 poin 2) — itu gejala butir 2, bukan butir 6.

---

## RENCANA PERBAIKAN untuk worker fix

Prinsip: **perbaiki penyebab render, jangan ubah semantik model** — kecuali opsi (b) self-message yang sudah disetujui user. Kerjakan pada salinan, verifikasi di StarUML asli, baru timpa `.mdj` proyek (dengan backup). Setelah SEMUA langkah, SHA baru dicatat dan file dibuka ulang di StarUML asli untuk verifikasi final.

### P0 — Penomoran: matikan penomoran otomatis di kedua diagram (WAJIB)

Masalah: penomoran ekspor acak + `0 :` (butir 1) adalah penyebab kebingungan nomor satu, dan tidak ada field nomor yang bisa "diperbaiki" per pesan.

1. Pada objek `UMLSequenceDiagram` 3a (`_id 4YF2OCLYJIEHEMSG6PD01C`) dan 3b (`_id GIWQO8EXC2IM60QE75P21N`), tambahkan field **`"showSequenceNumber": false`**.
   - Perhatian: field ini saat ini TIDAK ada di JSON (default StarUML `true`). Worker WAJIB membuktikan StarUML menghormatinya: buka salinan hasil fix di StarUML 7.1.0 → panel Properties diagram harus menunjukkan `showSequenceNumber` tidak tercentang → screenshot live 3a+3b harus menampilkan label TANPA prefix `N : ` → ekspor ulang SVG → `grep` SVG untuk pola ` : ` di awal teks label harus 0 hasil.
   - Bila StarUML ternyata TIDAK menghormati field tersebut (label tetap bernomor di live canvas): JANGAN mengarang field lain. Fallback = biarkan `showSequenceNumber=true` + tambahkan satu kalimat ke dokumentasi kedua diagram (lihat poin 3) + instruksikan user memakai screenshot live canvas (xwd), bukan ekspor, untuk skripsi. Laporkan fallback ini secara eksplisit.
2. **JANGAN** mengubah `sequenceNumbering` (tetap `auto`), `showSignature`, `showActivation`.
3. Tambahkan satu kalimat ke `documentation` kedua diagram (setelah kalimat "Urutan dibaca atas ke bawah."): *"Penomoran tidak ditampilkan karena penomoran otomatis StarUML tidak konsisten antara kanvas dan ekspor; urutan dibaca dari atas ke bawah."* — dokumentasi yang ada sudah menyatakan urutan dibaca atas-ke-bawah + legenda garis, jadi ini konsisten, bukan perubahan semantik.
4. **JANGAN** mengubah: nama pesan, `messageSort`, `source`/`target`, nilai Y, `points` (kecuali P1), `documentation` selain kalimat di atas.

### P1 — Self-message: opsi (b) lifeline terpisah (UTAMA, bila user setuju)

#### P1-3a — lifeline baru `Argon2id (library)`

1. **Model:** tambah `UMLLifeline` baru bernama `Argon2id (library)` ke `ownedElements` interaction 3a (`WZPCL7FO9OJB7J386YQRIF`), diletakkan bersebelahan dengan lifeline lain. Beri `_id` baru yang unik.
2. **View lifeline baru** (`UMLSeqLifelineView` di `ownedViews` diagram 3a): `left = 595`, `width = 140`, `top = 20`, `height = 1255`, `wordWrap = true`, + subview `UMLNameCompartmentView` (height ~74, 2 baris teks) + `UMLLinePartView` (left relatif 70) — cermin lifeline yang ada, dengan `_id` baru.
3. **Geser PostgreSQL:** view lifeline `PostgreSQL: users, sessions` (`FAEQTRJNL6YA7XDK6045IA`): `left` 595 → **760** (dashed 665 → 830). Model lifeline tidak berubah.
4. **Pesan `verify…`** (model `VJ3PTSKMGLJDE7066LX9XD`, view `ETPM0NI326GSBWAP3WBE88`): `target` model = lifeline baru (source tetap Auth API). View: `points` = **`500:415;665:415`** (ganti loop `500:415;530:415;530:435;500:435`). Label (`EdgeLabelView` `QR2GOQW60VO72WFC0AT77W`): `edgePosition = 1`, `distance = 10`, `alpha = 1.5707963267948966` (tetap), `left` = 582,5 − 300/2 = **432,5**, `top` = 415 − 16 = **399**, `width = 300` (tetap), teks tetap. (Label di atas panah, pola sama seperti pesan L→R lain.)
5. **Pesan-pesan yang menyentuh PostgreSQL** (target/source berubah karena PostgreSQL geser; Y, sort, nama, label `top`/`width` TETAP; label `left` = titik tengah panah baru − width/2):
   | Pesan | `points` lama → baru | label `left` baru (= titik tengah − width/2) |
   |---|---|---|
   | `SELECT users WHERE username/email` (view `GJJFEOM2ZQOU91H3UAHMXI`, label width 247) | `500:255;665:255` → `500:255;830:255` | 665 − 123,5 = **541,5** (top tetap 239) |
   | `return user + …` (view `OFDZTOXS9OBGCNJY8PC86Q`, label width 375) | `665:335;500:335` → `830:335;500:335` | 665 − 187,5 = **477,5** (top tetap 339) |
   | `UPDATE users SET …` (view `TY3OD9G1E3OCCS3GP159UP`, label width 435) | `500:775;665:775` → `500:775;830:775` | 665 − 217,5 = **447,5** (top tetap 759) |
   | `OK` (view `L0UMA6CQG05PWOBLDC938J`, label width 15) | `665:855;500:855` → `830:855;500:855` | 665 − 7,5 = **657,5** (top tetap 859) |
   | `INSERT sessions (…)` (view `91ICMC3VKUFO1YTAU5TZ9R`, label width 300) | `500:1095;665:1095` → `500:1095;830:1095` | 665 − 150 = **515** (top tetap 1079) |
   Pesan lain 3a TIDAK disentuh (points + label tetap).
6. **Dokumentasi diagram 3a:** tambahkan `Argon2id (library)` ke daftar lifeline dalam `documentation` (setelah "Auth API (auth.go)").
7. **Verifikasi:** buka salinan di StarUML → `verify…` harus tampil sebagai panah normal di Y=415 antara `return` dan `403`, label di atas panah, TIDAK menyeberangi garis lain (span teks ≈ [432, 732] vs garis 500/665 — akan menyentuh/melewati tipis di kedua ujung karena teks ~282px > pitch 165px; ini pola standar butir 3, diterima) → tidak ada lagi slot kosong / loop melayang / label di bawah diagram.

#### P1-3b — lifeline baru `autoSubmitJikaLewat (exam.go)`

1. **Model:** tambah `UMLLifeline` baru bernama `autoSubmitJikaLewat (exam.go)` ke `ownedElements` interaction 3b (`ZN6CLMFXY4TGM2OWCAQIZE`), `_id` baru.
2. **View lifeline baru:** `left = 595`, `width = 140`, `top = 20`, `height = 1335`, `wordWrap = true` + name compartment (~74) + line part (relatif 70) — cermin lifeline lain.
3. **Geser tiga lifeline:** Grading (`N6FL9REZYEHN6L1JH6IGEJ`) `left` 595 → **760**; PostgreSQL (`JAUQFMS4MKRBSAK3T123T7`) 760 → **925**; Monitor WebSocket (`NPFFM72JP1I2GZRRNCUO96`) 925 → **1090**. (Dashed baru: Grading 830, PostgreSQL 995, Monitor 1160.)
4. **Pesan `cek deadline lewat…`** (model `UN7WCVBZWOTRUHY9NVOKV0`, view `MXGI09OCO1G45WO9CDLMVS`): `target` = lifeline baru. View: `points` = **`500:815;665:815`**. Label: `left` = 582,5 − 150 = **432,5**, `top` = 815 − 16 = **799**, `width = 300` tetap, EP/distance/alpha tetap.
5. **Pesan-pesan yang menyentuh Grading / PostgreSQL / Monitor** (Y, sort, nama, label `top`/`width` TETAP; label `left` = titik tengah baru − width/2):
   | Pesan | `points` lama → baru | label `left` baru |
   |---|---|---|
   | `INSERT/UPDATE jawaban (…)` (label w 367) | `500:295;830:295` → `500:295;995:295` | 747,5 − 183,5 = **564** (top 279) |
   | `OK - jawaban tersimpan (autosave)` (label w 247) | `830:375;500:375` → `995:375;500:375` | 747,5 − 123,5 = **624** (top 379) |
   | `SELECT attempt.deadline, jadwal.status` (label w 285) | `500:635;830:635` → `500:635;995:635` | 747,5 − 142,5 = **605** (top 619) |
   | `deadline + status` (label w 127) | `830:695;500:695` → `995:695;500:695` | 747,5 − 63,5 = **684** (top 699) |
   | `hitungNilai: HitungPoin…` (label w 322) | `500:895;665:895` → `500:895;830:895` | 665 − 161 = **504** (top 879) |
   | `poin + nilai sementara` (label w 165) | `665:975;500:975` → `830:975;500:975` | 665 − 82,5 = **582,5** (top 979) |
   | `UPDATE attempt status=selesai…` (label w 420) | `500:1055;830:1055` → `500:1055;995:1055` | 747,5 − 210 = **537,5** (top 1039) |
   | `OK - attempt selesai tersimpan…` (label w 330) | `830:1135;500:1135` → `995:1135;500:1135` | 747,5 − 165 = **582,5** (top 1139) |
   | `broadcastAttempt (WS: …)` (label w 337) | `500:1215;995:1215` → `500:1215;1160:1215` | 830 − 168,5 = **661,5** (top 1199) |
   Pesan lain 3b TIDAK disentuh.
6. **Dokumentasi diagram 3b:** tambahkan `autoSubmitJikaLewat (exam.go)` ke daftar lifeline dalam `documentation`.
7. **Verifikasi:** sama seperti 3a — `cek deadline lewat…` tampil sebagai panah normal di Y=815 antara `sisa waktu server` dan `hitungNilai`.

#### P1-fallback — opsi (a), bila user menolak (b)

1. Pertahankan kedua self-message persis seperti sekarang (points, label, Y — JANGAN diubah).
2. Tambah **Note** di tiap diagram (coba `UMLNoteView` pada salinan; verifikasi tampil di StarUML sebelum kirim):
   - 3a: *"Catatan: 'verify Argon2id(password, password_hash)' secara semantik terjadi setelah 'return user + password_hash + must_change_password' dan sebelum '403 {error.code: PASSWORD_CHANGE_REQUIRED}'. StarUML selalu me-render self-message di bawah diagram."*
   - 3b: *"Catatan: 'cek deadline lewat (autoSubmitJikaLewat)' secara semantik terjadi setelah 'sisa waktu server' dan sebelum 'hitungNilai: HitungPoin per soal (otomatis)'. StarUML selalu me-render self-message di bawah diagram."*
   - Note merujuk NAMA pesan, bukan nomor (karena P0 mematikan penomoran).
3. Tambahkan kalimat yang sama ke `documentation` masing-masing diagram (sebagai cadangan bila Note gagal tampil).

### P2 — Label menyeberangi garis lifeline: TERIMA (default) + syarat

1. **Default: TERIMA.** Setelah P0 + P1, penyeberangan yang tersisa adalah pola standar label terpusat (butir 3.5). Jangan melebar pitch (165→~400px akan membuat diagram ~2× lebih lebar dan lebih sulit dibaca) dan jangan memperpendek nama pesan (mengorbankan presisi endpoint/field) tanpa persetujuan user.
2. **Syarat verifikasi:** setelah P0+P1, worker membuka salinan di StarUML + screenshot live 3a/3b atas-bawah + ekspor ulang SVG + jalankan ulang analisis penyeberangan; bila user masih menilai tidak jelas, tuas berikutnya (perlu keputusan baru, BUKAN default) adalah: (i) `edgePosition` label → 0/2 (uji di salinan: apakah label berpindah ke ujung panah dan apakah itu mengurangi penyeberangan tanpa merusak keterbacaan — JANGAN asumsikan, buktikan di StarUML), atau (ii) perpendek 4 label terpanjang dengan persetujuan user (3a `UPDATE…` 62 kar, `return…` 54 kar; 3b `GET status…` 70 kar, `UPDATE attempt…` 61 kar).
3. **JANGAN** mengubah `edgePosition`/`distance`/`alpha` label mana pun tanpa bukti screenshot StarUML bahwa hasilnya lebih jelas.

### Yang TIDAK boleh diubah + alasannya

- **Nama pesan, `messageSort`, `source`/`target` pesan non-self, nilai Y, gap antar pesan** — semuanya terbukti benar (butir 4–5); mengubahnya merusak semantik yang sudah tepat.
- **`points` pesan non-self** — ujung panah sudah tepat di garis lifeline (butir 6); menggesernya justru menimbulkan celah.
- **Lebar/posisi kotak kepala lifeline, `wordWrap`** — sudah seragam dan benar (butir 6).
- **Diagram lain** (Use Case, Class, ERD, Activity) — di luar cakupan tugas ini.
- **File lain** (ANALISIS.md, dll) — kecuali: worker WAJIB memeriksa apakah dokumen lain menyebut "self-message" / hitungan pesan Sequence; bila ya, selaraskan seperlunya dan catat perubahannya (jangan diam-diam mengubah dokumen akademik).
- **Jangan menambah field yang belum dibuktikan** (mis. mode `sequenceNumbering` lain / field nomor per pesan) — bila P0 gagal dibuktikan, pakai fallback yang tertulis di P0.

### Checklist verifikasi worker fix (wajib sebelum kirim)

1. `sha256sum` sebelum/sesudah dicatat; backup `.mdj` lama disimpan.
2. Validator 0 pelanggaran + file dibuka di StarUML asli 7.1.0 tanpa error (status bar ✅).
3. Screenshot live 3a + 3b (atas + bawah): tanpa prefix `N : ` (bila P0 berhasil); `verify…`/`cek deadline…` sebagai panah normal di posisi semantiknya (bila P1-(b)); tidak ada slot kosong / loop melayang / label di bawah diagram.
4. Ekspor ulang SVG: `grep` tidak menemukan pola `0 : ` / `N : ` pada label pesan (bila P0 berhasil).
5. Analisis penyeberangan dijalankan ulang: label self tidak lagi menyeberangi garis lifeline lain (bila P1-(b)); penyeberangan tersisa hanya pola terpusat standar.
6. File `.mdj` proyek tidak pernah dibuka-tulis oleh StarUML (edit hanya via skrip pada salinan, lalu verifikasi) — atau bila dibuka di StarUML, ditutup tanpa save.

---

## Lampiran: bukti mentah yang dipakai

- `.mdj` SHA `c068a179…` (sebelum & sesudah diagnosis, tidak berubah).
- SVG ekspor: `your_files/uml-svg/06-sequence-3a.svg`, `07-sequence-3b.svg` — teks label + nomor ekspor diekstrak dari `<text>` (watermark `UNREGISTERED` dikecualikan; koordinat = atribut + `transform`).
- Screenshot live canvas StarUML 7.1.0 (salinan `/tmp/seq-diagnosis-copy.mdj`): 3a atas (nomor live 1–9), 3a bawah (nomor live 7–16, loop self melayang + label menyeberangi garis PostgreSQL), 3b atas (nomor live 1–9), 3b bawah (nomor live 9–16, panah `broadcastAttempt` solid + kepala terbuka).
- Pengukuran piksel: garis putus-putus 3a di screenshot x=371/539/701/866 (offset kanvas +201); teks label `16 : verify…` di [824, 1093] (lebar 269px / 45 karakter ≈ 6,26 px/karakter); loop self di x≈735–765, melayang ~30px dari garis Auth (701).
- Panel Properties diagram (live): `showSequenceNumber=true`, `sequenceNumbering=auto`, `showSignature=true`, `showActivation=true`.
