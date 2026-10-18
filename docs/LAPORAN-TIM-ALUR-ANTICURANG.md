# LAPORAN TIM ALUR UJIAN & ANTI-CURANG — CBT SMK v2 vs Extraordinary CBT

**Tim:** Alur Ujian & Anti-Curang (10 orang) — analisis pembanding CBT SMK v2
**Tanggal:** 16 Oktober 2026
**Status:** HANYA ANALISIS. Tidak ada kode yang diubah. Binary Extraordinary CBT TIDAK dijalankan, TIDAK ada pembobolan lisensi/crack/decompile, TIDAK ada kredensial yang dipakai, TIDAK ada fitur yang dikarang.

## 0. Cara membaca label bukti (dipakai di seluruh laporan)

| Label | Arti |
|---|---|
| **TERBUKTI** | Tertulis eksplisit di sumber yang disebut (kajian / ANALISIS.md / dokumentasi publik yang dikutip). |
| **DISIMPULKAN** | Tidak tertulis eksplisit, tetapi mengikuti secara logis dari bukti yang disebut + alasan ditulis. Bisa salah bila premisnya salah. |
| **TIDAK DIKETAHUI** | Tidak ada bukti di bahan yang dibaca. Tidak boleh dipakai sebagai dasar keputusan. |
| **USULAN / BELUM DIPUTUSKAN** | Tertulis di ANALISIS.md sebagai usulan/asumsi/pertanyaan terbuka (Bab 23/24), bukan keputusan final. |

**Sumber utama:**
- S1 = `~/workspace/extraordinary-cbt/PELAJARI-EXTRAORDINARY-CBT.md` (kajian Exo v4.7.0 — instalasi terhenti di gerbang lisensi; fitur Exo diketahui dari daftar menu/permission + skema dump 38 tabel, BUKAN dari uji live).
- S2 = `~/workspace/cbt-smk-v2/ANALISIS.md` (spesifikasi/usulan CBT SMK v2, Bab 4, 7, 10.3, 11, 12, 13, 15, 23, 24).
- S3 = pencarian web dokumentasi publik untuk pembanding umum (Moodle / Exam.net / Quizizz) — lihat §5. Hasil pencarian untuk Exam.net dan Quizizz lemah (didominasi sumber pihak ketiga), sehingga klaim tentang keduanya diberi label bukti rendah / TIDAK DIKETAHUI bila ragu.

**Batasan penting yang memengaruhi semua kesimpulan:**
- Karena Exo tidak pernah jalan live dalam kajian (S1), TIDAK ADA satu pun perilaku runtime Exo (mis. "timer Exo berjangkar server", "server Exo menolak submit telat") yang bisa dilabeli TERBUKTI. Yang TERBUKTI hanyalah *keberadaan* nama fitur/tabel/setting di menu/skema.
- CBT SMK v2 pada tahap ANALISIS.md adalah spesifikasi/usulan, bukan kode yang sudah teruji. Semua mekanisme v2 di laporan ini adalah **USULAN** kecuali dinyatakan lain. Jangan membaca laporan ini sebagai "v2 sudah punya fitur X".

---

## 1. State machine sesi ujian

### 1.1 Dua level yang jangan dicampur

Kesalahan paling umum dalam membandingkan sistem CBT adalah mencampur dua mesin status yang berbeda:

| Level | Yang berubah status | Contoh status |
|---|---|---|
| **Jadwal** (kolektif) | Satu jadwal/sesi untuk banyak siswa | DRAF → BERJALAN → DIJEDA → SELESAI |
| **Sesi peserta** (per siswa per jadwal) | Satu siswa dalam satu jadwal | BELUM MULAI → BERJALAN → DIBLOKIR / SELESAI |

Pause adalah aksi level **jadwal** (berdampak ke semua siswa dalam jadwal itu). Blokir adalah aksi level **sesi peserta** (satu siswa). Force finish bisa di kedua level (per siswa atau per jadwal). Mencampur keduanya membuat aturan seperti "pause membekukan timer" menjadi ambigu — timer siapa yang dibekukan, dan apa yang terjadi pada siswa yang belum mulai?

### 1.2 State machine JADWAL

#### Extraordinary CBT (dari S1)

```
DRAF ──▶ BERJALAN ──▶ DIJEDA ──▶ BERJALAN ──▶ SELESAI
              │                      │
              └────▶ SELESAI ◀────────┘
```

| Transisi | Label | Bukti |
|---|---|---|
| Ada jadwal yang bisa di-pause/resume | **TERBUKTI** (S1) | S1 §4: "Tambah/edit/hapus, **pause & resume** jadwal, 8 sesi" |
| Pause → resume kembali ke BERJALAN | **DISIMPULKAN** | Alasan: "resume" tanpa makna lain selain kembali berjalan; S1 tidak menjelaskan detail semantik resume. |
| Pause membekukan timer siswa | **TIDAK DIKETAHUI** | S1 tidak menyebut timer, `sisa_waktu`, atau perilaku timer saat pause. Jangan mengasumsikan Exo membekukan timer. |
| Jadwal SELESAI → DINILAI → TERBIT | **TIDAK DIKETAHUI** | S1 menyebut hasil/ledger/koreksi, tetapi tidak memberi urutan status jadwal pasca-selesai. |
| Jadwal bisa di-force-finish / reset | **TERBUKTI** (S1, keberadaan fitur) | S1 §4: "Reset ujian, force finish, reset login peserta". Detail siapa boleh, dan efek ke sesi peserta yang sedang berjalan: TIDAK DIKETAHUI. |

#### CBT SMK v2 (dari S2)

```
DRAF ──▶ SIAP ──▶ DIJADWALKAN ──▶ BERJALAN ──▶ DIJEDA ──▶ BERJALAN ──▶ SELESAI ──▶ DINILAI ──▶ TERBIT
                                  │                        │
                                  └──────▶ SELESAI ◀───────┘
```

| Transisi | Label | Bukti |
|---|---|---|
| `DRAF → SIAP → DIJADWALKAN → BERJALAN → DIJEDA → BERJALAN → SELESAI → DINILAI → TERBIT` | **TERBUKTI** (S2, sebagai spesifikasi) | S2 Bab 4.1 menulis rantai status jadwal persis seperti ini. |
| Pause membekukan timer semua siswa | **USULAN / BELUM DIPUTUSKAN** | S2 Bab 4.3: "Pause jadwal — Membekukan timer semua siswa (asumsi; …)". S2 Bab 23 butir 14 menegaskan ini asumsi v2, bukan perilaku CBT-Go (di CBT-Go pause HANYA gerbang masuk, timer TIDAK dibekukan). S2 Bab 24 Q12 menanyakan persetujuan user. **Jadi: jangan mengklaim v2 "sudah" membekukan timer.** |
| Mekanisme pembekuan: simpan `sisa_waktu` saat pause, lanjutkan dari nilai itu saat resume | **USULAN** | S2 Bab 13.4: "pause harus membekukan waktu (simpan `sisa_waktu` saat pause, lanjutkan dari nilai itu saat resume)". |
| Force finish per siswa / per jadwal oleh proktor/admin | **TERBUKTI** (S2, sebagai spesifikasi) | S2 Bab 4.3 + Bab 11.2 + Bab 15.2 (endpoint `force-finish`). |
| Restore ditolak bila ada ujian berjalan | **TERBUKTI** (S2, sebagai spesifikasi) | S2 Bab 15.5: restore → 409 `UJIAN_BERJALAN` + `konfirmasi:true` wajib. |

**Perbandingan jadwal:** rantai status v2 lebih eksplisit (ada SIAP, DIJADWALKAN, DINILAI, TERBIT) daripada yang terdokumentasi untuk Exo. Itu bukan bukti Exo "tidak punya" tahap-tahap itu — hanya bukti bahwa kajian Exo tidak mendokumentasikannya. Jangan mengklaim keunggulan dari ketiadaan bukti.

### 1.3 State machine SESI PESERTA (per siswa per jadwal)

```
BELUM MULAI ──▶ BERJALAN ──┬──▶ SELESAI ──▶ DINILAI ──▶ TERBIT (nilai final)
                            │
                            ├──▶ DIBLOKIR ──▶ BERJALAN (unblock: lanjut)   [v2: USULAN, 1 opsi dari 3]
                            │                  │
                            │                  └──▶ BELUM MULAI (unblock: reset) [v2: USULAN]
                            │                  │
                            │                  └──▶ SELESAI (unblock: tetap diblokir + akhiri) [v2: USULAN]
                            │
                            └──▶ SELESAI (force finish / auto-submit saat deadline / submit siswa)
```

Jalur masuk ke BERJALAN (keduanya mensyaratkan login + token valid — detail berbeda, lihat §1.4):

| Langkah masuk | Exo | CBT SMK v2 |
|---|---|---|
| Login siswa | **TERBUKTI** (S1: peserta, `LOGIN_ONCE_ATIME`) | **TERBUKTI** (S2: login NIS + password, Bab 4.2) |
| Token | **TERBUKTI** (S1: generate, revoke, expiry 900 detik default) | **TERBUKTI** (S2, spesifikasi): token 6 karakter per jadwal/sesi, valid bila jadwal BERJALAN + belum revoke + belum lewat expiry; TIDAK ditampilkan di dashboard siswa sebelum sesi |
| Token expiry default | **TERBUKTI** (S1: 900 detik default) | **USULAN / BELUM DIPUTUSKAN** — S2 Bab 15.2 menyebut `token_expiry_detik` default 900; S2 Bab 24 Q15 menanyakan persetujuan. Catatan: di CBT-Go (pendahulu v2) expiry nyata adalah akhir hari jadwal, bukan 900 detik (S2 Bab 23 butir 14) — jadi angka 900 di v2 adalah usulan, bukan perilaku yang diwarisi. |
| Token kedua untuk jadwal yang sama | **TIDAK DIKETAHUI** (Exo) | **BELUM DIPUTUSKAN** (v2) — S2 Bab 24 Q14: berlaku keduanya / pertama hangus / ditolak? Belum diputuskan. Ini penting: jawaban yang salah di sini membuat token pertama bisa dipakai joki setelah token kedua dirilis. |
| Toleransi telat | **TIDAK DIKETAHUI** (Exo) | **USULAN** — S2 Bab 12.2 (pola TKA/AN: maks 15 menit) + Bab 24 Q16. Sisa waktu telat = jam selesai − waktu masuk, tanpa tambahan (usulan). |

#### Transisi keluar dari BERJALAN (sesi peserta)

| Transisi | Exo | CBT SMK v2 |
|---|---|---|
| Submit siswa (2 langkah: review → konfirmasi final) | **TIDAK DIKETAHUI** — S1 tidak menjelaskan alur submit Exo (satu klik / dua langkah / konfirmasi). | **TERBUKTI** (S2, spesifikasi) — Bab 4.2 + Bab 9.3: layar ringkasan review → modal konfirmasi final; submit tanpa `konfirmasi:true` → 422 (Bab 15.3). |
| Auto-submit saat deadline | **TIDAK DIKETAHUI** — S1 tidak menyebut auto-submit / deadline enforcement / job penyapu. | **USULAN** — S2 Bab 10.3: "deadline di server (otoritatif, auto-submit server + job penyapu tiap 1 menit)". Ini usulan perbaikan atas CBT-Go ("tidak ada enforcement deadline server"). **Bukan perilaku teruji.** |
| Force finish oleh proktor/admin | **TERBUKTI** (S1, keberadaan) | **TERBUKTI** (S2, spesifikasi) |
| Auto-block karena pelanggaran | **TERBUKTI** (S1, keberadaan + ambang) — `AUTO_BLOCK=Y`, `MAX_COUNTER_OUT=1`, `out_ujian_counter`, alasan blokir, `antiblock` | **TERBUKTI** (S2, spesifikasi) — 1x peringatan + counter 1; 2x auto-block (Bab 7.1 T1). Debounce 5 detik agar tidak double-count (Bab 7.1, 15.5). **Tetapi** Bab 24 Q19 menanyakan persetujuan ambang ini — jadi ambang `MAX_COUNTER_OUT=1` adalah usulan default yang masih butuh persetujuan, bukan keputusan final. |
| Unblock | **TERBUKTI** (S1, keberadaan `antiblock`) — siapa mengeksekusi, apakah jawaban dipertahankan, apakah timer lanjut: **TIDAK DIKETAHUI** | **USULAN / BELUM DIPUTUSKAN** — S2 Bab 4.3: hanya Proktor/Admin mengeksekusi, Pengawas hanya mengusulkan; opsi lanjut / reset / tetap diblokir + akhiri. S2 Bab 24 Q13: harus satu aturan, belum diputuskan. |
| Putus jaringan / tutup tab lalu kembali | **TIDAK DIKETAHUI** (Exo) | **USULAN** — S2 Bab 9.3: "tutup tab lalu buka lagi saat ujian berjalan → kembali ke soal terakhir dengan sisa waktu server"; Bab 10.3: antrean offline (localStorage/IndexedDB + kirim ulang + indikator tersimpan/menunggu/offline), heartbeat 15–30 detik, `out_counter` hanya bertambah saat terhubung + aktif. Semua usulan, bukan perilaku teruji. |
| Pindah tab tepat saat auto-submit | **TIDAK DIKETAHUI** (Exo) | **USULAN / target** — S2 Bab 9.3: "tidak ada nilai ganda/hilang/menggantung". Diwujudkan via submit idempotent (`Idempotency-Key`) + upsert jawaban per `(siswa_ujian_id, soal_id)` (Bab 15.1). Target desain, bukan bukti. |

#### Transisi SELESAI → DINILAI → TERBIT (sesi peserta)

| Transisi | Exo | CBT SMK v2 |
|---|---|---|
| Nilai otomatis (objektif) + koreksi manual (esai) | **TERBUKTI** (S1: koreksi otomatis + manual uraian + `ai_opinion` di koreksi) | **TERBUKTI** (S2, spesifikasi) — nilai akhir = Σ poin / Σ bobot × 100; esai dikoreksi guru dengan rubrik skor 0 s.d. bobot (Bab 4.2, 6.3). |
| Nilai esai disembunyikan sampai final | **TIDAK DIKETAHUI** (Exo) | **USULAN** — setting `tampil_hasil_ke_siswa`, default sembunyikan sampai final (Bab 15.2, 15.6; Bab 24 Q18 belum diputuskan). |
| Finalisasi ditolak bila masih ada jawaban manual belum dikoreksi | **TIDAK DIKETAHUI** (Exo) | **TERBUKTI** (S2, spesifikasi) — Bab 15.5: finalisasi ditolak + daftar `belum_dikoreksi`. |
| Nilai final = ledger terkunci; perubahan setelah publish hanya revisi tercatat | **TIDAK DIKETAHUI** (Exo — S1 menyebut "ledger nilai" tetapi tidak menjelaskan aturan kunci/revisi) | **TERBUKTI** (S2, spesifikasi) — Bab 8.1: nilai final dikunci/publish menjadi ledger. |

### 1.4 Catatan state machine: apa yang TIDAK boleh disimpulkan

1. **Tidak ada bukti bahwa Exo memakai timer berjangkar server.** S1 tidak menyebut timer sama sekali di luar kata "Timer" pada baris fitur ujian peserta (S1 §6, tanpa detail). Klaim "Exo menegakkan deadline di server" adalah **TIDAK DIKETAHUI**, bukan TERBUKTI.
2. **Tidak ada bukti bahwa Exo menolak save/submit setelah deadline.** Sama seperti di atas — tidak ada di S1.
3. **Tidak ada bukti bahwa Exo tidak mengirim kunci ke browser.** S1 tidak menjelaskan isi endpoint soal Exo. Ini **TIDAK DIKETAHUI**, bukan kekurangan Exo dan bukan keunggulan Exo.
4. **Status "DIBLOKIR" di Exo vs v2 tidak bisa dibandingkan kedalamannya.** Exo TERBUKTI punya counter + alasan blokir + `antiblock`; v2 punya spesifikasi lebih rinci (debounce, unblock 3 opsi, SoD pengawas-mengusulkan/proktor-mengeksekusi). Spesifikasi yang lebih rinci bukan bukti implementasi yang lebih baik — v2 belum dibangun.

---

## 2. Klasifikasi mekanisme anti-curang

Klasifikasi tiga lapis (sesuai tugas):

- **(a) DITEGAKKAN SERVER** — server adalah otoritas; klien yang dimodifikasi / dimatikan tidak bisa mengubah hasil. Ini satu-satunya kategori yang benar-benar "mencegah".
- **(b) DETEKSI SAJA** — sistem mencatat / memberi peringatan / memblokir *setelah* peristiwa terdeteksi; peristiwa itu sendiri tidak dicegah, dan deteksi bisa diakali atau menghasilkan false positive.
- **(c) TIDAK BISA DICEGAH WEB** — tidak ada mekanisme web app (tanpa pengawasan manusia / perangkat khusus) yang bisa mencegahnya.

### 2.1 Klasifikasi per mekanisme (berlaku umum, lalu per aplikasi di §2.2)

| # | Mekanisme | Kelas | Alasan |
|---|---|---|---|
| M1 | Kunci jawaban tidak dikirim ke browser (endpoint soal tanpa kunci/poin/pembahasan) | **(a) DITEGAKKAN SERVER** | Jika server tidak pernah mengirim kunci, tidak ada yang bisa diintip di DevTools/Network, bagaimanapun klien dimodifikasi. Satu-satunya cara bocor adalah kebocoran di luar jalur ini (mis. guru membagikan kunci). |
| M2 | Timer berjangkar server (sisa waktu dihitung dari `mulai_ujian` / `sisa_waktu` tersimpan di server) | **(a) DITEGAKKAN SERVER** | Jam browser bisa diubah; jam server tidak bisa diubah siswa. Syarat: server menghitung deadline, bukan sekadar menampilkan hitung mundur dari nilai yang dikirim klien. |
| M3 | Server tolak save/submit setelah deadline | **(a) DITEGAKKAN SERVER** | Penolakan terjadi di server; request telat ditolak apa pun yang dilakukan klien. Syarat: cek deadline di handler save/submit, bukan hanya di UI. |
| M4 | Nilai dihitung server; nilai kiriman browser diabaikan | **(a) DITEGAKKAN SERVER** | Nilai tidak bisa dipalsukan dari browser bila server menghitung ulang dari jawaban tersimpan. |
| M5 | Permission ditegakkan server (403 + catat; anti-IDOR; 404 untuk resource yang tidak berhak) | **(a) DITEGAKKAN SERVER** | Menyembunyikan menu di frontend bukan penegakan. Penegakan = server menolak. |
| M6 | Token: acak, terikat jadwal + jendela waktu, expiry, revoke, rate limit | **(a) DITEGAKKAN SERVER** *(dengan syarat)* | Token 6 karakter dari 32 simbol ≈ 1 miliar kombinasi — S2 Bab 7.1 T5 menegaskan ini cukup kuat **HANYA** jika ada rate limit + masa berlaku singkat + revoke. Tanpa ketiganya, token kuat di atas kertas tetapi bisa ditebak/dibagikan. |
| M7 | Lock bank soal (bank dipakai jadwal aktif tidak bisa diubah; snapshot soal per paket/jadwal) | **(a) DITEGAKKAN SERVER** | Server menolak edit + hasil merujuk snapshot, bukan soal yang sedang diedit. Mencegah "soal berubah di tengah ujian" dan "kunci diubah setelah ujian". |
| M8 | Audit log append-only (login, submit, pelanggaran, block/unblock, token, perubahan nilai, export, akses ditolak) | **(a) DITEGAKKAN SERVER** *(sebagai jejak)* | Audit log tidak mencegah pelanggaran, tetapi membuat pelanggaran tidak bisa dihapus diam-diam — bila append-only ditegakkan server dan tidak ada hapus audit log (S2 Bab 7.4, SoD-6). Ini penegakan *akuntabilitas*, bukan penegakan *pencegahan*. |
| M9 | Deteksi pindah tab / minimize / blur (`visibilitychange` + `blur`) + counter + auto-block | **(b) DETEKSI SAJA** | Deteksi terjadi setelah siswa sudah pindah tab; tidak mencegah siswa melihat tab lain. Bisa diakali (perangkat kedua, browser kedua, DevTools, script). Bisa false positive (cek Wi-Fi, notifikasi OS, putus jaringan — S2 Bab 10.3). Debounce 5 detik mengurangi double-count, bukan menghilangkan false positive. |
| M10 | Fullscreen enforcement (auto-block bila keluar fullscreen) | **(b) DETEKSI SAJA** | S2 Bab 7.2: "Fullscreen di web hanya deteksi (Esc selalu bisa keluar)". Tidak ada API web yang bisa mengunci siswa di fullscreen. Exo mengklaim "fullscreen enforcement (auto-block kalau tidak fullscreen)" (S1) — itu tetap deteksi + konsekuensi, bukan pencegahan keluar fullscreen. |
| M11 | Device lock / satu sesi aktif / `LOGIN_ONCE_ATIME` | **(b) DETEKSI SAJA** *(di web)* | S2 Bab 7.2: "Device lock web hanya deteksi/penghambat (cookie bisa dihapus); kunci kuat butuh kiosk / Safe Exam Browser / MDM". Di v2, `X-Device-Id` adalah UUID di localStorage (Bab 15.4) — localStorage bisa dihapus, cookie bisa dihapus, browser lain = device baru. Menolak perangkat kedua adalah penegakan *di dalam satu browser yang jujur*, bukan kunci perangkat keras. |
| M12 | Nonaktifkan copy-paste / klik kanan / seleksi teks | **(b) DETEKSI SAJA** *(tepatnya: penghambat lemah)* | S2 Bab 7.1 T2: "penghambat, bukan jaminan". DevTools, print, screenshot, atau menonaktifkan JS melewatinya. |
| M13 | Watermark + larangan screenshot | **(b) DETEKSI SAJA** *(tepatnya: deterrence + jejak)* | Watermark tidak mencegah screenshot; ia membuat screenshot yang bocor bisa ditelusur (bila watermark memuat identitas + timestamp) dan memberi efek jera. |
| M14 | Pengacakan isi paket + urutan soal + urutan opsi per siswa | **(b) DETEKSI SAJA** *(tepatnya: penghambat kolusi)* | Mengurangi nilai berbagi jawaban ("nomor 3 jawabannya B" tidak berlaku antar siswa), tetapi tidak mencegah berbagi jawaban berbasis isi soal ("soal tentang fotosintesis jawabannya X") atau joki. S2 Bab 7.1 T6 menempatkannya sebagai pencegahan berbagi jawaban — klasifikasikan secara jujur sebagai penghambat, bukan penegakan. |
| M15 | Rate limit login/token (5 gagal per akun per 15 menit → 429 + kunci 15 menit) | **(a) DITEGAKKAN SERVER** | Penolakan terjadi di server. Syarat: limit per akun + per IP + per token, dan tidak bisa di-reset dengan ganti IP saja (bila hanya per IP, penyerang rotasi IP). |
| M16 | CSRF + hapus hanya via POST + cookie `Secure` + HTTPS | **(a) DITEGAKKAN SERVER** | Mencegah serangan dari situs lain / link berbahaya, bukan mencegah siswa curang. Termasuk di sini karena sering disalahpahami sebagai "anti-curang" — ini keamanan platform. |
| M17 | Monitor real-time proktor (status + out_counter + tombol unblock/reset/force-finish) | **(b) DETEKSI SAJA** | Monitor memberi visibilitas, bukan pencegahan. Proktor yang melihat pelanggaran masih butuh bertindak; monitor tanpa aksi = pencatatan (pelajaran dari CBT Laravel, S2 Bab 13.2). |
| M18 | HP kedua / perangkat kedua di luar browser ujian | **(c) TIDAK BISA DICEGAH WEB** | Web app tidak melihat perangkat lain di ruangan. Hanya pengawasan manusia (pengawas ruang, razia HP — S2 Bab 12.2: peserta kumpulkan HP sebelum ujian) + kiosk/MDM yang bisa mengurangi. |
| M19 | Foto / screenshot layar (kamera HP mengarah ke layar) | **(c) TIDAK BISA DICEGAH WEB** | S2 Bab 7.1 T3, Bab 7.2: tidak bisa dicegah penuh di web. Watermark (M13) hanya jejak/jera. |
| M20 | Menoleh / berbisik / isyarat antar siswa | **(c) TIDAK BISA DICEGAH WEB** | Hanya pengawasan manusia + tata letak ruangan + pengacakan (M14, penghambat). |
| M21 | Penyamaran / joki (orang lain yang mengerjakan) | **(c) TIDAK BISA DICEGAH WEB** | S2 Bab 7.1 T7: pengawas cek foto kartu login vs wajah + daftar hadir; software tidak bisa cegah penuh. Verifikasi biometrik / proctoring video bisa mengurangi, tetapi bukan bagian dari v2 maupun Exo yang terdokumentasi. |
| M22 | Kolusi pengawas / admin (token dibocorkan, nilai diubah, audit dihapus) | **(c) TIDAK BISA DICEGAH WEB** *(oleh teknologi saja)* | Hanya pemisahan tugas (SoD), audit log, dan pengawasan organisasi yang mengurangi (S2 Bab 7.1 T10, Bab 11.4). Teknologi tidak bisa mencegah admin yang sah menyalahgunakan aksesnya — ia hanya bisa membuat penyalahgunaan tercatat. |
| M23 | Listrik / jaringan padam (server / lab) | **(c) TIDAK BISA DICEGAH WEB** | Software tidak bisa mencegah padam. Yang bisa: UPS, pause resmi, antrean offline, backup, rencana susulan (S2 Bab 10.3, 10.7). |
| M24 | Hafalan antar gelombang / soal bocor antar sesi | **(c) TIDAK BISA DICEGAH WEB** *(sepenuhnya)* | Pengacakan + rasio bank 3× + tidak memakai soal yang sama untuk paket bersamaan (S2 Bab 7.1 T12, Bab 5.6) mengurangi, tetapi siswa gelombang 1 tetap bisa menghafal dan menceritakan soal ke gelombang 2. |
| M25 | AI di perangkat lain (chatbot di HP kedua, earpiece) | **(c) TIDAK BISA DICEGAH WEB** | Sama seperti M18 — di luar jangkauan web app. |
| M26 | DevTools / manipulasi klien (ubah JS, kirim request manual) | **(c) TIDAK BISA DICEGAH WEB** *(sebagai perilaku)* — tetapi dampaknya **bisa** dibuat nol oleh (a) | S2 Bab 7.2: web TIDAK bisa cegah DevTools. Yang benar: DevTools tidak boleh *berdampak* — karena kunci tidak dikirim (M1), nilai dihitung server (M4), deadline ditolak server (M3). "Mencegah DevTools" adalah target yang salah; "membuat DevTools tidak berguna" adalah target yang benar. |

### 2.2 Status tiap mekanisme: Exo vs CBT SMK v2

Legenda status: ✅ ada (TERBUKTI di sumber) · 📋 spesifikasi/usulan (TERBUKTI tertulis di ANALISIS.md sebagai usulan) · ❓ TIDAK DIKETAHUI · ⚠️ ada tetapi nominal / tidak ditegakkan (di pendahulu) · — tidak ada.

| Mekanisme | Kelas | Exo (S1) | CBT SMK v2 (S2) |
|---|---|---|---|
| M1 Kunci tidak dikirim ke browser | (a) | ❓ TIDAK DIKETAHUI — S1 tidak menjelaskan isi endpoint soal | 📋 TERBUKTI sebagai spesifikasi (Bab 15.3: endpoint soal TIDAK BOLEH berisi kunci/poin/pembahasan) |
| M2 Timer berjangkar server | (a) | ❓ TIDAK DIKETAHUI — S1 hanya menyebut "Timer" tanpa detail | 📋 TERBUKTI sebagai spesifikasi + usulan perbaikan (Bab 7.1 T8; Bab 10.3 deadline server + job penyapu; Bab 23: CBT-Go TIDAK punya ini) |
| M3 Tolak save/submit setelah deadline | (a) | ❓ TIDAK DIKETAHUI | 📋 TERBUKTI sebagai spesifikasi/usulan (Bab 7.1 T8, Bab 10.3) |
| M4 Nilai dihitung server | (a) | ❓ TIDAK DIKETAHUI — S1 menyebut koreksi otomatis, tetapi tidak menyebut "nilai kiriman browser diabaikan" | 📋 TERBUKTI sebagai spesifikasi (Bab 15.3: "server hitung nilai sendiri; nilai kiriman browser diabaikan") |
| M5 Permission server-side | (a) | ✅ TERBUKTI keberadaan (100+ permission per aksi, `role_menus`) — penegakan server vs hanya menu: ❓ | 📋 TERBUKTI sebagai spesifikasi (Bab 11.5: deny-by-default, 403 + catat, anti-IDOR, 404) |
| M6 Token kuat + expiry + revoke + rate limit | (a) | ✅ TERBUKTI (generate, revoke, expiry 900 detik default) — rate limit token: ❓ | 📋 TERBUKTI sebagai spesifikasi (Bab 7.1 T5, Bab 15.2, Bab 24 Q14/Q15 belum diputuskan) |
| M7 Lock bank + snapshot | (a) | ✅ TERBUKTI (kunci bank soal; backup/restore bank) — snapshot per paket/jadwal: ❓ | 📋 TERBUKTI sebagai spesifikasi (Bab 5.4: lock 2 lapis + snapshot per paket/jadwal) |
| M8 Audit log append-only | (a) | ❓ TIDAK DIKETAHUI — S1 tidak menyebut audit log (ada log keluar ujian, bukan audit log umum) | 📋 TERBUKTI sebagai spesifikasi (Bab 7.4: format, append-only, retensi) |
| M9 Deteksi pindah tab + counter + auto-block | (b) | ✅ TERBUKTI (`out_ujian_counter`, `AUTO_BLOCK=Y`, `MAX_COUNTER_OUT=1`, alasan blokir) — debounce / dedup: ❓ | 📋 TERBUKTI sebagai spesifikasi (1x peringatan, 2x block, debounce 5 detik; Bab 24 Q19 belum diputuskan) |
| M10 Fullscreen enforcement | (b) | ✅ TERBUKTI (klaim S1: auto-block kalau tidak fullscreen) — mekanisme deteksi: ❓ | 📋 TERBUKTI sebagai spesifikasi (Bab 15.6: fullscreen *deteksi*; Bab 13.2: penegakan keras ditunda, bertahap) |
| M11 Device lock / login-once | (b) | ✅ TERBUKTI (tabel `devices`, 1 perangkat/peserta, `LOGIN_ONCE_ATIME=Y`) — apakah perangkat kedua *ditolak* vs hanya dicatat: ❓ (S1 memakai kata "verifikasi", bukan "tolak") | 📋 TERBUKTI sebagai spesifikasi (Bab 13.2 P2: "perangkat dicatat DAN ditegakkan"; Bab 15.4 `X-Device-Id`) — dengan catatan jujur Bab 7.2: device lock web hanya deteksi/penghambat |
| M12 Nonaktifkan copy-paste | (b) | ❓ TIDAK DIKETAHUI | 📋 TERBUKTI sebagai spesifikasi (Bab 7.1 T2: penghambat, bukan jaminan) |
| M13 Watermark + larangan screenshot | (b) | ❓ TIDAK DIKETAHUI | 📋 TERBUKTI sebagai spesifikasi (Bab 7.1 T3: watermark + larangan + berita acara) |
| M14 Pengacakan isi + urutan + opsi | (b) | ❓ TIDAK DIKETAHUI — S1 tidak menyebut pengacakan | 📋 TERBUKTI sebagai spesifikasi (Bab 5.6: 2 lapis pengacakan + seed per siswa+jadwal untuk audit) |
| M15 Rate limit | (a) | ❓ TIDAK DIKETAHUI | 📋 TERBUKTI sebagai spesifikasi (Bab 7.1 T13, Bab 15.1: login 5 gagal/15 menit → 429 + kunci 15 menit; simpan jawaban 60 req/menit) |
| M16 CSRF + POST-only delete + Secure cookie | (a) | ❓ TIDAK DIKETAHUI (binary tertutup) | 📋 TERBUKTI sebagai spesifikasi P0 (Bab 7.5, 15.1, 15.4) |
| M17 Monitor real-time | (b) | ✅ TERBUKTI (monitor peserta real-time, WebSocket aktif — S1 §2) | 📋 TERBUKTI sebagai spesifikasi (Bab 13.2 P2, Bab 15.2; WebSocket vs polling masih diputuskan — Bab 23 butir 15, Bab 24 Q46) |
| M18–M25 (HP kedua, foto, menoleh, joki, kolusi, padam, hafalan, AI perangkat lain) | (c) | ❓ TIDAK DIKETAHUI apakah Exo mengklaim mencegahnya — S1 tidak mengklaim itu | 📋 TERBUKTI sebagai batasan jujur (Bab 7.2 L1–L10: TIDAK bisa dicegah penuh; mitigasi = pengawas + tata tertib + pause/susulan + audit) |

**Kesimpulan §2.2:** Di atas kertas spesifikasi, v2 lebih lengkap dan lebih jujur klasifikasinya daripada yang terdokumentasi untuk Exo. Tetapi perbandingan ini tidak simetris: Exo dinilai dari kajian menu/skema (yang tidak mendokumentasikan mekanisme server-side), v2 dinilai dari spesifikasi yang memang ditulis untuk mendokumentasikan mekanisme server-side. **Jangan menyimpulkan "Exo tidak punya M1–M4"** — yang benar adalah "kajian Exo tidak membuktikan M1–M4". Sebaliknya, jangan menyimpulkan "v2 sudah punya M1–M4" — yang benar adalah "v2 menspesifikasikan M1–M4, belum dibangun".

---

## 3. Matriks "mekanisme × aplikasi × kekuatan bukti"

Kekuatan bukti: **Kuat** = TERBUKTI dari sumber primer yang relevan (kode teruji / spesifikasi eksplisit / dokumentasi resmi) · **Sedang** = TERBUKTI keberadaan fitur tetapi mekanisme tidak didokumentasikan · **Lemah** = DISIMPULKAN / sumber sekunder · **Tidak ada** = TIDAK DIKETAHUI / tidak ada bukti.

| Mekanisme | Exo | CBT SMK v2 | Moodle | Exam.net | Quizizz |
|---|---|---|---|---|---|
| Kunci tidak dikirim ke browser | Tidak ada (❓) | Kuat 📋 (spesifikasi eksplisit) | Kuat — DISIMPULKAN dari arsitektur quiz server-side + review options (kunci ditahan sampai kuis ditutup) — bukti: pencarian S3, sumber sekunder | Tidak ada (❓) | Lemah — bukti pihak ketiga menunjukkan jawaban pernah bisa diakses via klien (bookmarklet cheat beredar); TIDAK DIKETAHUI apakah itu masih berlaku saat ini |
| Timer berjangkar server | Tidak ada (❓) | Kuat 📋 (spesifikasi + job penyapu) | Kuat — timer kuis Moodle ditegakkan server (attempt ditutup server saat waktu habis) — DISIMPULKAN dari dokumentasi umum Moodle | Tidak ada (❓) | Lemah — timer ada, penegakan server TIDAK DIKETAHUI |
| Tolak save/submit setelah deadline | Tidak ada (❓) | Kuat 📋 (spesifikasi) | Kuat — DISIMPULKAN dari model attempt Moodle | Tidak ada (❓) | Tidak ada (❓) |
| Nilai dihitung server | Tidak ada (❓) | Kuat 📋 (spesifikasi) | Kuat — DISIMPULKAN dari model grading Moodle | Tidak ada (❓) | Tidak ada (❓) |
| Deteksi pindah tab + counter + auto-block | Sedang (keberadaan + ambang TERBUKTI; debounce ❓) | Kuat 📋 (spesifikasi + debounce 5 detik) | Sedang — deteksi tab bukan fitur inti Moodle; yang ada adalah SEB + proctoring plugin (pihak ketiga) | Sedang — Exam.net disebut memakai SEB (S3: blog Exam.net tentang SEB) — detail deteksi tab TIDAK DIKETAHUI | Lemah — deteksi tab-switch diklaim ada (produk "Focus Mode" / deteksi), tetapi bukti yang ditemukan dalam pencarian ini adalah alat bypass pihak ketiga, bukan dokumentasi resmi — TIDAK DIKETAHUI secara pasti |
| Fullscreen enforcement | Sedang (klaim auto-block TERBUKTI; mekanisme ❓) | Kuat 📋 (spesifikasi: deteksi, penegakan keras ditunda) | Sedang — via SEB (lockdown browser), bukan via web biasa | Sedang — via SEB / lockdown (klaim umum) | Tidak ada (❓) |
| Device lock / login-once | Sedang (tabel + setting TERBUKTI; penolakan perangkat kedua ❓) | Kuat 📋 (spesifikasi: dicatat DAN ditegakkan) | Sedang — SEB config key + plugin proctoring; bukan fitur inti kuis | Tidak ada (❓) | Tidak ada (❓) |
| Token + expiry + revoke | Sedang (generate/revoke/expiry TERBUKTI) | Kuat 📋 (spesifikasi; expiry 900 detik belum diputuskan) | Lemah — Moodle memakai password kuis / SEB key, bukan token per sesi seperti CBT — TIDAK SEBANDING langsung | Tidak ada (❓) | Lemah — kode game (game code) untuk masuk, bukan token ujian dengan expiry/revoke — TIDAK SEBANDING langsung |
| Lock bank soal + snapshot | Sedang (lock TERBUKTI; snapshot ❓) | Kuat 📋 (spesifikasi 2 lapis + snapshot) | Sedang — Moodle punya versi soal / question versioning — DISIMPULKAN, detail TIDAK DIKETAHUI | Tidak ada (❓) | Tidak ada (❓) |
| Audit log append-only | Tidak ada (❓) | Kuat 📋 (spesifikasi) | Kuat — Moodle punya log standar (event log) — DISIMPULKAN dari dokumentasi umum | Tidak ada (❓) | Lemah — laporan hasil ada; audit log keamanan TIDAK DIKETAHUI |
| Pengacakan soal/opsi | Tidak ada (❓) | Kuat 📋 (spesifikasi 2 lapis + seed) | Kuat — shuffle soal/opsi adalah fitur kuis Moodle — DISIMPULKAN dari dokumentasi umum | Tidak ada (❓) | Sedang — pengacakan urutan soal/opsi diklaim di sumber sekunder (S3) |
| Proctoring video / AI (webcam, deteksi wajah/HP) | Tidak ada (❓) | Tidak ada (tidak direncanakan di v2) | Sedang — BUKAN inti; via plugin/integrasi (proctoring plugin, SMOWL) — bukti S3 | Lemah — Exam.net diklaim punya monitoring AI (klaim pemasaran umum); dokumentasi resmi tidak diverifikasi dalam tugas ini — TIDAK DIKETAHUI | Tidak ada (❓) |
| Lockdown browser (SEB / kiosk) | Tidak ada (❓) | Tidak ada (kiosk hanya bila sekolah siap — S2 Bab 7.2) | Kuat — SEB terintegrasi default di Moodle Quiz (S3: moodle.com) | Sedang — disebut memakai SEB (S3: blog Exam.net) | Tidak ada (❓) |
| HP kedua / foto / menoleh / joki | Tidak ada (❓) | Kuat 📋 (dinyatakan TIDAK BISA — batasan jujur) | Kuat (prinsip umum — tidak ada web app yang bisa; Moodle mengandalkan SEB + proctor + pengawas) | Kuat (prinsip umum) | Kuat (prinsip umum) |

**Catatan matriks:** Kolom Moodle/Exam.net/Quizizz diberi label DISIMPULKAN dari dokumentasi publik / sumber sekunder, BUKAN dari uji live dalam tugas ini. Bila keputusan desain bergantung pada satu sel matriks pembanding umum, verifikasi ke dokumentasi resmi masing-masing produk sebelum mengunci keputusan. Khusus Quizizz: Quizizz/Wayground adalah alat kuis formatif/gamifikasi, bukan sistem ujian sumatif tertutup — membandingkannya head-to-head dengan CBT sumatif menyesatkan bila konteksnya tidak disebut.

---

## 4. Kesalahpahaman umum (klaim anti-curang yang terdengar kuat tapi sebenarnya hanya deteksi)

| # | Klaim yang menyesatkan | Kenyataannya | Kelas benar |
|---|---|---|---|
| K1 | "Kami mencegah siswa pindah tab." | Tidak. Yang ada: deteksi *setelah* pindah + konsekuensi (peringatan/block). Siswa sudah sempat melihat tab lain. | (b) DETEKSI SAJA |
| K2 | "Fullscreen enforcement mengunci siswa di ujian." | Tidak. Di web, Esc selalu bisa keluar fullscreen (S2 Bab 7.2). Yang ditegakkan adalah *konsekuensi* keluar fullscreen, bukan fullscreen itu sendiri. Kunci sungguhan butuh SEB/kiosk/MDM. | (b) DETEKSI SAJA |
| K3 | "Device lock memastikan 1 perangkat per siswa." | Tidak — di web, cookie/localStorage bisa dihapus dan browser lain = "perangkat baru". Device lock web adalah penghambat + deteksi, bukan kunci perangkat keras (S2 Bab 7.2). | (b) DETEKSI SAJA |
| K4 | "Kami menonaktifkan copy-paste, jadi soal tidak bisa dicuri." | Tidak. DevTools, print, screenshot, foto HP, atau menonaktifkan JS melewatinya (S2 Bab 7.1 T2: penghambat, bukan jaminan). | (b) DETEKSI SAJA |
| K5 | "Auto-block berarti ujian kami aman." | Tidak. Auto-block adalah *reaksi* terhadap pelanggaran yang sudah terjadi, dan bisa false positive (cek Wi-Fi, notifikasi OS, putus jaringan — S2 Bab 10.3). Tanpa aturan unblock yang adil + audit, auto-block justru menciptakan sengketa. | (b) DETEKSI SAJA |
| K6 | "Token 6 karakter = 1 miliar kombinasi, jadi token kami aman." | Hanya bila ada rate limit + expiry singkat + revoke (S2 Bab 7.1 T5). Token kuat tanpa ketiganya bisa ditebak (brute force tanpa rate limit) atau dibagikan (tanpa expiry/revoke). | (a) hanya bila 3 syarat terpenuhi |
| K7 | "Kami mencatat semua pelanggaran di log, jadi kami menegakkan anti-curang." | Tidak. Pencatatan tanpa konsekuensi adalah teater (pelajaran dari CBT Laravel — S2 Bab 13.2: "Log saja tidak menegakkan; block tanpa log tidak bisa diaudit"). Penegakan = konsekuensi + jejak. | (b) DETEKSI SAJA |
| K8 | "Perangkat dicatat, jadi perangkat dikunci." | Tidak. "Dicatat ≠ ditegakkan" (S2 Bab 13.4). CBT-Go mencatat perangkat (cookie + IP) tetapi perangkat kedua tidak ditolak — S2 menyebut ini janji palsu bila diklaim sebagai kunci. | (b) DETEKSI SAJA |
| K9 | "Pause membekukan timer." | Belum tentu — tergantung implementasi. Di CBT-Go, pause HANYA gerbang masuk; timer TIDAK dibekukan (S2 Bab 23 butir 14). Jangan mengklaim "pause membekukan timer" tanpa memverifikasi implementasi `sisa_waktu`. | Tergantung implementasi |
| K10 | "Kami memakai AI proctoring, jadi kecurangan terdeteksi otomatis." | AI memberi *sinyal* (flag), bukan vonis. Sinyal AI butuh review manusia (prinsip umum proctoring — S3: "AI-generated alert does not automatically indicate misconduct"). AI juga punya false positive (cahaya, kamera, gerakan) dan tidak bisa membuktikan niat. | (b) DETEKSI SAJA |
| K11 | "Soal kami aman karena siswa tidak bisa membuka DevTools." | Tidak ada web app yang bisa mencegah DevTools (S2 Bab 7.2). Target yang benar: DevTools tidak berguna — karena kunci tidak dikirim (M1), nilai dihitung server (M4), deadline ditolak server (M3). | (c) sebagai perilaku; dampaknya (a) |
| K12 | "Kami mencegah screenshot." | Tidak — di web, tidak bisa dicegah penuh (S2 Bab 7.1 T3). Watermark hanya jejak + jera. Satu-satunya "pencegahan" screenshot adalah SEB/kiosk di perangkat yang dikelola — dan itu pun tidak menghentikan foto HP ke layar. | (c) TIDAK BISA DICEGAH WEB |
| K13 | "Monitor real-time berarti pengawas mengawasi." | Monitor memberi visibilitas; pengawasan adalah tindakan manusia. Monitor tanpa aksi proktor (unblock/reset/force-finish/insiden) = dashboard yang ditonton (S2 Bab 13.2: proktor buta di CBT Laravel karena tanpa monitor; kebalikannya — monitor tanpa aksi — sama tidak bergunanya). | (b) DETEKSI SAJA |
| K14 | "Nilai kami aman karena dihitung di frontend lalu dikirim." | Ini justru lubang paling kritis. Nilai HARUS dihitung server; nilai kiriman browser harus diabaikan (S2 Bab 15.3). Nilai dari klien bisa dipalsukan dengan satu request manual. | (a) bila dihitung server; ❌ bila dihitung klien |
| K15 | "Sistem kami mencegah joki / penyamaran." | Tidak ada software ujian web yang bisa mencegah joki sendirian (S2 Bab 7.1 T7). Yang bisa: cek foto kartu vs wajah oleh pengawas, daftar hadir, dan (di luar cakupan v2/Exo) verifikasi biometrik/proctoring video. | (c) TIDAK BISA DICEGAH WEB |

---

## 5. Pembanding umum — ringkasan (dengan label bukti)

### Moodle (Quiz)
- **TERBUKTI (S3, moodle.com):** Moodle Quiz terintegrasi SEB (Safe Exam Browser) secara default — SEB membatasi akses situs/fungsi sistem selama ujian; siswa wajib menginstal SEB sebelum mengerjakan.
- **TERBUKTI (S3, moodle.com):** Integrasi proctoring bersertifikat (mis. SMOWL) tersedia — monitoring berkelanjutan + deteksi tindakan tidak sah + AI + deteksi VM.
- **DISIMPULKAN (dokumentasi umum Moodle):** Timer kuis, grading, dan review options ditegakkan server (kunci ditahan sampai kuis ditutup); soal/opsi bisa diacak; event log tersedia. Alasan: arsitektur quiz server-side Moodle yang terdokumentasi luas.
- **TIDAK DIKETAHUI (dalam tugas ini):** Detail penegakan deadline per attempt, perilaku pause attempt, dan aturan auto-block bawaan Moodle — tidak diverifikasi ke dokumentasi resmi Moodle dalam tugas ini, jadi tidak diklaim.
- **Pelajaran untuk v2:** Moodle menunjukkan pola yang benar — web biasa untuk kuis sehari-hari, SEB/kiosk untuk ujian berisiko tinggi, proctoring sebagai lapisan terpisah. v2 yang meniru pola ini (web + deteksi + SEB/kiosk opsional) lebih jujur daripada v2 yang mengklaim web-nya sendiri "mengunci" siswa.

### Exam.net
- **TERBUKTI (S3, blog Exam.net):** Exam.net menyebut pemakaian SEB bersama Exam.net.
- **DISIMPULKAN (klaim umum produk):** Exam.net dipasarkan sebagai platform ujian terkunci dengan monitoring. Detail mekanisme (apakah timer server-side, apakah ada AI proctoring bawaan, bagaimana auto-block bekerja) **TIDAK DIKETAHUI** dari pencarian dalam tugas ini — hasil pencarian didominasi sumber pihak ketiga, bukan dokumentasi resmi Exam.net.
- **Pelajaran untuk v2:** Jangan mengutip klaim pemasaran Exam.net sebagai bukti mekanisme. Bila v2 ingin membandingkan diri dengan Exam.net, verifikasi ke dokumentasi resmi Exam.net dulu.

### Quizizz / Wayground
- **DISIMPULKAN (sumber sekunder, S3):** Quizizz punya pengacakan urutan soal/opsi, limit waktu, dan laporan; diklaim ada deteksi pindah tab ("Focus Mode" / deteksi tab-switch).
- **TIDAK DIKETAHUI:** Apakah deteksi tab-switch Quizizz ditegakkan (auto-block) atau hanya dicatat; apakah timer ditegakkan server; apakah kunci dikirim ke klien. Bukti yang ditemukan dalam pencarian ini justru adalah alat bypass pihak ketiga dan bookmarklet cheat yang beredar — itu bukti bahwa deteksi klien-side bisa diakali, bukan bukti dokumentasi resmi.
- **Pelajaran untuk v2:** Quizizz adalah pembanding yang salah konteks bila dipakai untuk ujian sumatif tertutup — ia dirancang untuk kuis formatif/gamifikasi. Mengambil "deteksi tab" Quizizz sebagai standar anti-curang ujian sumatif adalah kesalahan kategori.

---

## 6. Rekomendasi untuk implementasi v2 (kriteria terima, bukan kode)

Rekomendasi ini diterjemahkan dari S2 agar bisa diuji — setiap butir harus punya tes terima sebelum diklaim "ada":

1. **M1–M4 adalah P0 dan harus diuji dengan klien yang dimodifikasi.** Tes terima: (a) response endpoint soal tidak mengandung kunci/poin/pembahasan — periksa body mentah; (b) ubah jam browser → sisa waktu tidak berubah; (c) kirim save/submit setelah deadline via request manual → server menolak (bukan hanya UI yang menyembunyikan tombol); (d) kirim nilai palsu dari browser → nilai server tidak berubah.
2. **Putuskan Bab 24 Q12–Q15 + Q19 sebelum menulis kode ujian.** Pause membekukan timer atau tidak; unblock = lanjut / reset / akhiri (SATU aturan); token kedua = pertama hangus atau tidak; expiry token; ambang auto-block. Lima keputusan ini mengubah state machine — menulis kode sebelum memutuskannya berarti menulis ulang state machine.
3. **Debounce + dedup pelanggaran + aturan jaringan.** Tes terima: satu pindah tab tidak terhitung 2x; putus jaringan murni tidak menambah `out_counter`; 1 pelanggaran jaringan yang dikonfirmasi proktor tidak menghukum siswa (S2 Bab 10.3).
4. **Jangan klaim device lock / fullscreen sebagai "pencegahan".** Dokumentasikan sebagai deteksi + konsekuensi. Bila sekolah butuh pencegahan sungguhan, jalurnya SEB/kiosk/MDM — bukan menambah JavaScript.
5. **Audit log harus append-only sejak hari pertama.** Tes terima: tidak ada endpoint / role yang bisa menghapus atau mengubah baris audit log; setiap block/unblock, token generate/revoke, perubahan nilai, dan akses ditolak tercatat dengan aktor + alasan + sebelum/sesudah.
6. **SoD-5 (pengawas mengusulkan, proktor mengeksekusi unblock) ditegakkan server**, bukan hanya di UI. Tes terima: pengawas memanggil endpoint unblock langsung → 403 + tercatat.
7. **Jangan bandingkan v2 dengan Exo memakai klaim yang tidak terbukti.** Perbandingan yang jujur: "Exo TERBUKTI punya X di menu/skema; v2 MENSPESIFIKASIKAN X; X di v2 belum teruji." Klaim "v2 lebih aman dari Exo" baru sah setelah tes terima §6 butir 1 lulus untuk v2.
8. **Siapkan prosedur hari-H untuk kategori (c).** Karena HP kedua / foto / menoleh / joki / padam tidak bisa dicegah web, sekolah butuh: razia HP + tata tertib (S2 Bab 12.2), pengawas cek kartu vs wajah, pause resmi + susulan, dan berita acara pelanggaran (ringan/sedang/berat + bukti + saksi). Tanpa ini, anti-curang v2 hanya setengah sistem.

---

## 7. Pertanyaan terbuka yang memblokir finalisasi state machine

| # | Pertanyaan | Sumber |
|---|---|---|
| 1 | Pause membekukan timer? | S2 Bab 24 Q12 |
| 2 | Unblock: lanjut sisa waktu / reset / tetap diblokir + akhiri? (harus satu aturan) | S2 Bab 24 Q13 |
| 3 | Token kedua: pertama hangus / berlaku keduanya / ditolak? | S2 Bab 24 Q14 |
| 4 | Token expiry default 900 detik? Token dirilis berapa menit sebelum mulai? | S2 Bab 24 Q15 |
| 5 | Toleransi telat berapa menit? (usulan 15 menit, pola TKA/AN) | S2 Bab 24 Q16 |
| 6 | Ambang auto-block: `MAX_COUNTER_OUT = 1` (1x peringatan, 2x block)? Device lock sejak hari pertama? Fullscreen di v2.0 atau ditunda? | S2 Bab 24 Q19, Q21 |
| 7 | Nilai esai disembunyikan sampai final? (`tampil_hasil_ke_siswa`, default sembunyikan) | S2 Bab 24 Q18 |
| 8 | Auth: session cookie server-side (revokasi instan — peserta diblokir berhenti di request berikutnya) vs JWT + refresh (revokasi sulit)? Keputusan ini menentukan apakah "diblokir" benar-benar menghentikan siswa di request berikutnya. | S2 Bab 23 butir 10, Bab 24 Q42 |
| 9 | WebSocket sejak v2.0 atau polling 10 detik dulu + WebSocket menyusul? | S2 Bab 23 butir 15, Bab 24 Q46 |
| 10 | Berapa pelanggaran (pindah tab, keluar fullscreen) sebelum auto-block, dan apakah pengawas dapat mengubah ambang per jadwal? | S2 Bab 24 Q58 |

**Pertanyaan 8 adalah yang paling kritis untuk anti-curang:** bila auth memakai JWT access token tanpa denylist, siswa yang baru diblokir masih bisa memakai token-nya sampai kedaluwarsa — "diblokir" menjadi label di database, bukan penghentian nyata. Session cookie server-side (atau JWT + denylist `jti` / access token sangat pendek) menutup lubang ini. Keputusan ini harus diambil sebelum implementasi block/unblock, bukan sesudah.

---

## 8. Daftar sumber

- S1: `~/workspace/extraordinary-cbt/PELAJARI-EXTRAORDINARY-CBT.md` (kajian Exo v4.7.0).
- S2: `~/workspace/cbt-smk-v2/ANALISIS.md` (Bab 4, 5, 7, 8, 9, 10, 11, 12, 13, 15, 23, 24).
- S3: Pencarian web 16 Okt 2026 — Moodle & proctoring (moodle.com, "Moodle and proctoring"), blog Exam.net tentang SEB, dan sumber sekunder Quizizz. URL lengkap ada di hasil pencarian; tidak ada halaman yang dibuka lewat browser live dalam tugas ini.
