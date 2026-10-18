# ANALISIS — CBT SMK v2

**Proyek:** CBT SMK v2 — web ujian online untuk SMK
**Stack target:** Go (backend) + React JS (frontend) + PostgreSQL
**Tanggal laporan:** 16 Oktober 2026
**Status:** FASE ANALISIS + RISET SAJA. **Belum ada kode aplikasi yang ditulis.** Development dimulai SETELAH user menyetujui laporan ini.
**Penyusun:** Koordinator, dari 20 laporan worker (8 Analisis, 6 Research, 6 Development-mode-standby).

---

## Daftar Isi

1. [Ringkasan Eksekutif](#1-ringkasan-eksekutif)
2. [Konteks & Referensi yang Sudah Ada](#2-konteks--referensi-yang-sudah-ada)
3. [Kebutuhan per Role (Stakeholder)](#3-kebutuhan-per-role-stakeholder)
4. [Alur Ujian End-to-End](#4-alur-ujian-end-to-end)
5. [Bank Soal, Kurikulum & Kisi-kisi](#5-bank-soal-kurikulum--kisi-kisi)
6. [16 Tipe Soal — Koreksi Penting + Daftar + Penilaian](#6-16-tipe-soal--koreksi-penting--daftar--penilaian)
7. [Anti-Curang & Keamanan](#7-anti-curang--keamanan)
8. [Laporan, Nilai & Analisis Butir Soal](#8-laporan-nilai--analisis-butir-soal)
9. [UX/UI — Standar "Lebih Cantik"](#9-uxui--standar-lebih-cantik)
10. [Infrastruktur & Operasional](#10-infrastruktur--operasional)
11. [Roles & Permission — Matriks](#11-roles--permission--matriks)
12. [Regulasi yang Berlaku](#12-regulasi-yang-berlaku)
13. [Benchmark: Exo vs CBT-Go vs CBT Laravel](#13-benchmark-exo-vs-cbt-go-vs-cbt-laravel)
14. [Arsitektur yang Diusulkan (Go + React)](#14-arsitektur-yang-diusulkan-go--react)
15. [Desain API (Ringkas)](#15-desain-api-ringkas)
16. [ERD (Ringkas)](#16-erd-ringkas)
17. [Rencana UI (Halaman & Komponen)](#17-rencana-ui-halaman--komponen)
18. [Rencana Sprint & Milestone](#18-rencana-sprint--milestone)
19. [Rencana Testing (QA)](#19-rencana-testing-qa)
20. [Rencana Build/Deploy (DevOps)](#20-rencana-builddeploy-devops)
21. [Risiko + Mitigasi](#21-risiko--mitigasi)
22. [Yang Ditunda ke Pasca-v2](#22-yang-ditunda-ke-pasca-v2)
23. [Kontradiksi Antar Laporan (Belum Diputuskan)](#23-kontradiksi-antar-laporan-belum-diputuskan)
24. [Daftar Pertanyaan untuk User (Wajib Dijawab Sebelum Development)](#24-daftar-pertanyaan-untuk-user-wajib-dijawab-sebelum-development)
25. [Sumber](#25-sumber)

---

## 1. Ringkasan Eksekutif

CBT SMK v2 adalah rebuild ujian online SMK dengan arsitektur **Go (backend) + React (frontend) + PostgreSQL**, menggantikan template HTML polos CBT-Go dengan UI modern yang responsif, sambil mengadopsi pelajaran dari Extraordinary CBT v4.7.0 (analisis butir soal, nilai per kompetensi, device lock, event+ruangan+kartu peserta) dan modul UKK dari CBT Laravel.

**Yang sudah ada (titik berangkat):**
- CBT-Go: Go single binary + PostgreSQL, port 9988, admin `/adm`, bank soal 16 tipe + lock + import CSV, jadwal + pause/resume + token, ujian peserta (timer, auto-submit, ragu-ragu), koreksi manual esai, anti-curang dasar, rekap + export CSV, monitor, settings. **Terbukti jalan & ter-screenshot.**
- CBT Laravel: modul UKK teori+praktik, berita acara, 6 role.
- Extraordinary CBT v4.7.0: dikaji dari paket resmi (38 tabel), tetapi **terhenti di gerbang lisensi** — sekolah perlu lisensi resmi dari ekosistemnya agar bisa jalan. v2 dibangun dari nol, tanpa menyalin mekanisme lisensi Exo.

**Temuan paling penting dari riset:**

1. **KOREKSI KRITIS — "16 tipe soal standar Kemendikbud" TIDAK DITEMUKAN.** Riset tidak menemukan dokumen resmi Kemendikbud/Pusmendik/BSNP yang menetapkan "16 tipe soal". Dokumen resmi paling otoritatif, *Framework AKM* (Pusat Asesmen Pendidikan, BSKAP, Kemendikbudristek, edisi revisi 2023, Lampiran 1 hlm. 147–149), menetapkan **hanya 5 bentuk soal**: Pilihan Ganda, Pilihan Ganda Kompleks, Menjodohkan, Isian/Jawaban Singkat, Uraian/Esai. Dari 16 nama di CBT-Go: 5 cocok penuh, sisanya adalah duplikat, submodel, varian penyajian, salah klasifikasi, label survei, tipe generik non-Kemendikbud, atau paket stimulus. **Tidak ada tipe resmi yang hilang.** Rekomendasi: pakai 5 tipe resmi sebagai inti + varian/ekstensi dengan label jujur. Lihat Bab 6.
2. **"USBK" bukan ujian nasional tersendiri.** Tidak ditemukan POS nasional bernama "POS USBK" yang masih berlaku. USBK diperlakukan sebagai **Ujian Sekolah bermoda CBT yang diatur POS sekolah**. ANBK/TKA adalah sistem nasional terpisah — CBT v2 adalah sistem **sekolah** (US + UKK), bukan pengganti ANBK/TKA. Lihat Bab 12.
3. **KKTP menggantikan KKM** di Kurikulum Merdeka (deskripsi/rubrik/interval per Tujuan Pembelajaran, bukan angka mutlak tunggal). CBT tidak boleh meng-hardcode KKM=75.
4. **UKK punya aturan tegas:** skala 0–100, lulus ≥ 70, predikat <70 / 70–79 / 80–90 / 91–100, teori + praktik + sikap, penguji internal + eksternal DUDIKA, sertifikat hanya bagi yang lulus.
5. **Batasan jujur anti-curang web:** web app TIDAK bisa cegah penuh HP kedua, screenshot/foto, menoleh/berbisik, penyamaran, DevTools, listrik/jaringan padam, kolusi. Yang bisa ditegakkan: kunci jawaban tidak dikirim ke browser, timer berjangkar server, server tolak save/submit setelah deadline, audit log, device lock sebagai deteksi/penghambat (bukan kunci kuat).
6. **Utang keamanan CBT-Go yang wajib dilunasi sebelum produksi:** CSRF belum ada, hapus via GET, akun demo dengan password bawaan masih aktif, tanpa rate limit, cookie tanpa `Secure`, sesi in-memory (restart = logout massal), autosave tanpa retry/antrean offline, `MAX_COUNTER_OUT` di settings tidak berpengaruh (logika hardcode), satu pindah tab bisa terhitung 2x (tanpa deduplikasi), auto-block mengubah status akun (bukan per-ujian), tanpa acak soal/jawaban, tanpa fullscreen, tanpa audit log, tanpa proteksi brute force, pool DB `MaxOpenConns(10)` terlalu kecil untuk 300 siswa serentak, hanya 4 indeks.

**Catatan integrasi (16 Okt 2026, setelah sintesis pertama):** empat laporan detail — Researcher Go Backend, Researcher Benchmark, Researcher React Frontend, Researcher Database & Arsitektur — diterima lengkap setelah draf pertama ANALISIS.md dan telah diintegrasikan ke Bab 13, 14.4, 14.5, 16, 23, dan 24. Integrasi ini menambahkan koreksi fakta penting (pause CBT-Go tidak membekukan timer; token expiry CBT-Go = akhir hari jadwal, bukan 900 detik; CBT-Go hanya 8 tipe yang auto-grade dengan 3 jalur render; CBT SMK anti-curang hanya pencatatan) dan dua usulan arsitektur yang bertentangan dengan laporan sebelumnya (Echo v5 vs stdlib; session cookie vs JWT; Argon2id vs bcrypt; `pgx`+`sqlc` vs `database/sql`+`lib/pq`) — semuanya dicatat sebagai kontradiksi yang belum diputuskan (Bab 23) + pertanyaan baru (Bab 24 bagian I), bukan dipilih diam-diam.

**Keputusan arsitektur utama (usulan, perlu persetujuan user):**
- Monorepo: `backend/` (Go) + `frontend/` (React + TypeScript) + `docs/` + `deploy/` — satu commit/PR bisa ubah API + UI, satu pipeline, satu artefak.
- Satu binary Go: frontend hasil build di-embed via `go:embed` (Opsi B1, disarankan) — deploy tetap satu file, tidak ada risiko aset dan backend tidak sinkron. Jangan pisah frontend ke hosting/CDN terpisah (memecah origin, masalah CORS/cookie, ujian mati saat internet putus).
- REST JSON `/api/v1/...`, nilai selalu dihitung backend (frontend TIDAK PERNAH menghitung nilai).
- Skenario deploy utama: **server lokal sekolah (LAN, semi-daring)**. VPS sebagai pelengkap opsional (tryout dari rumah). **cPanel shared hosting DICORET** untuk aplikasi CBT (tidak ada proses latar persisten, port kustom, PostgreSQL, systemd).
- Estimasi: **±73–79 hari kerja ≈ 15–16 minggu kalender** (1 developer fokus + AI, belum termasuk UAT di sekolah, jeda persetujuan, pengadaan server).

**Yang BELUM diputuskan (jangan dianggap sudah final):** lihat Bab 23 (kontradiksi) dan Bab 24 (daftar pertanyaan). Yang paling mendesak: (a) label jujur untuk 16 tipe soal; (b) React vs pertahankan template (perintah user: React — tetapi 3 laporan mencatat template lebih sederhana); (c) JWT vs session cookie; (d) aturan pause/unblock/token kedua; (e) apakah PROKTOR dan PENGAWAS role terpisah (6 vs 7 role).

---

## 2. Konteks & Referensi yang Sudah Ada

| Sistem | Stack | Status | Yang dipakai sebagai referensi v2 |
|---|---|---|---|
| **CBT-Go (v1)** | Go single binary ~11 MB, `net/http` + `html/template`, PostgreSQL (`cbt_go`), port 9988, `/adm` + `/` | Jalan, terverifikasi + ter-screenshot | Arsitektur single binary + PostgreSQL; alur ujian; anti-curang dasar; struktur 16 tipe (dengan koreksi Bab 6) |
| **CBT Laravel (SMK)** | Laravel 13, PHP 8.3, SQLite lokal | Jalan lokal; deploy VPS terblokir SSH | Modul UKK teori+praktik, berita acara, 6 role, import CSV |
| **Extraordinary CBT v4.7.0** | Binary Go + PostgreSQL, port 9988, `/adm` | Instalasi terhenti di gerbang lisensi | Desain fitur: 100+ permission, nilai per kompetensi, analisis butir soal, device lock, event+ruangan+kartu peserta, AI opsional — **bukan** mekanisme lisensinya |

**Target v2:** SMK semua jurusan, Kurikulum Merdeka, jenis ujian STS/SAS/USBK/UKK/Tryout, mode semi-daring (server lokal sekolah), role admin/proktor/guru/pengawas/siswa/penguji UKK. UI **lebih cantik** dari CBT-Go (template HTML polos) — modern, responsif, enak dipakai siswa.

**Catatan tentang data demo:** Akun demo CBT-Go (EVIL `admin@shellrean.id`/`admin123`, EVAL `guru@sekolah.id`/`guru123`, OPRA `proktor@sekolah.id`/`proktor123`, peserta `2024001`/`2024002`/`siswa123`) dan token demo `ABC123` hanya untuk dokumentasi/pengembangan. Di produksi: wajib ganti password saat login pertama, akun demo dinonaktifkan, token demo tidak dipakai.

---

## 3. Kebutuhan per Role (Stakeholder)

### 3.1 Siapa pemakainya

| Role | Siapa | Tugas inti |
|---|---|---|
| **Admin (EVIL)** | Super admin / penanggung jawab CBT sekolah | Kelola user, master data, settings, backup, audit log, finalisasi nilai |
| **Proktor (OPRA)** | Operator/proktor hari-H | Kelola jadwal, generate/rilis token, monitor, pause/resume, force finish, unblock, reset login |
| **Guru (EVAL)** | Guru mapel | Buat bank soal, kisi-kisi, paket, koreksi esai, lihat rekap + analisis butir mapelnya |
| **Pengawas** | Pengawas ruang (ROLE BARU di v2) | Awasi ruangan/sesinya, catat insiden, usulkan unblock — read-only + catat, tidak ubah soal/jadwal/nilai, **tidak boleh melihat token** |
| **Siswa (Peserta)** | Siswa SMK | Login → token → kerjakan ujian → lihat nilainya sendiri |
| **Penguji UKK** | Asesor UKK internal/eksternal (ROLE BARU) | Nilai praktik UKK via rubrik, hanya skema + kompetensi + peserta yang ditugaskan, akun eksternal masa berlaku terbatas |

**Stakeholder tidak langsung:** kepala sekolah, wali kelas, orang tua. Asumsi yang perlu konfirmasi: orang tua tidak diberi akun; wali kelas akses baca perlu konfirmasi; penguji eksternal UKK perlu konfirmasi.

### 3.2 Masalah ujian kertas/manual yang diselesaikan

Biaya/waktu cetak & distribusi; risiko bocor soal; pengawasan tidak seragam; koreksi lambat; rekap manual rawan salah; waktu tidak seragam antar ruangan; soal tidak pernah dievaluasi kualitasnya; arsip hilang; UKK praktik terpisah dari teori; internet terbatas (maka mode semi-daring).

### 3.3 Kebutuhan fungsional per role (ringkas)

- **Admin:** semua (lihat matriks Bab 11) + backup/restore dari aplikasi + event+ruangan+kartu peserta + import Word + analisis butir soal + kelola KKTP.
- **Proktor:** jadwal + token + monitor + tindakan hari-H + daftar hadir + berita acara. Tidak boleh lihat isi soal, tidak boleh koreksi.
- **Guru:** bank soal + kisi-kisi + paket (mapelnya) + koreksi esai + rekap + analisis butir. Tidak boleh ubah bank yang sedang dipakai jadwal aktif.
- **Pengawas:** monitor ruangan/sesinya + catat insiden + daftar hadir + berita acara. Tidak boleh melihat token, soal, jadwal lain, atau nilai.
- **Siswa:** dashboard jadwalnya + token + ujian + hasil sendiri. Tidak boleh lihat jadwal/nilai siswa lain walau menebak URL/ID.
- **Penguji UKK:** modul UKK (skema + unit kompetensi + peserta yang ditugaskan) + rubrik + bukti + rekap UKK.

### 3.4 Kebutuhan non-fungsional (target usulan, bukan fakta terukur)

Login/dashboard <2 detik di LAN; buka soal <1 detik; autosave <1 detik; monitor update 5–10 detik (atau WebSocket); dukung 1 lab penuh (30–40) sekaligus; target desain seluruh siswa satu sesi (300 siswa serentak). Target ini perlu divalidasi uji beban (Bab 19).

---

## 4. Alur Ujian End-to-End

```
Bank Soal → Kisi-kisi → Paket → Jadwal + Token → Kerjakan → Koreksi → Nilai → Laporan
```

### 4.1 Status

- **Bank soal:** `DRAF → SIAP → TERKUNCI → DIARSIPKAN`.
- **Jadwal:** `DRAF → SIAP → DIJADWALKAN → BERJALAN → DIJEDA → BERJALAN → SELESAI → DINILAI → TERBIT`.
- **Sesi peserta:** belum mulai → sedang ujian → selesai / diblokir / belum mulai (dengan alasan).

### 4.2 Alur lengkap

1. **Buat soal:** guru buat bank soal (satu mapel + satu tingkat), isi soal lengkap (tipe, stimulus, kunci, bobot, kesulitan, indikator, kompetensi, level kognitif), telaah oleh orang kedua, kunci bank.
2. **Paket:** dari bank, pilih soal manual / acak berstrata / campuran → paket (durasi, jumlah soal, aturan acak, KKTP). Satu bank → banyak paket (reguler, remedial, susulan, acak per siswa). **Paket adalah entitas baru di v2** — di CBT-Go jadwal langsung menunjuk ke bank soal.
3. **Jadwal + token:** proktor buat jadwal (tanggal, jam, sesi, rombel, ruangan, paket), generate token 6 karakter per jadwal/sesi. Token **TIDAK** ditampilkan di dashboard siswa sebelum sesi (diumumkan di ruangan). Token valid jika jadwal/sesi BERJALAN + belum revoke + belum lewat expiry.
4. **Kerjakan:** siswa login (NIS + password) → masukkan token → kerjakan (satu soal per layar, autosave per soal, timer dari server, tanda ragu-ragu, navigasi soal) → submit 2 langkah (layar ringkasan review → modal konfirmasi final). Siswa telat: sisa waktu = jam selesai − waktu masuk (tidak ada tambahan); batas toleransi ditetapkan sekolah (asumsi, default boleh masuk selama BERJALAN).
5. **Koreksi:** esai dikoreksi guru dengan rubrik (skor 0 s.d. bobot); nilai akhir = Σ poin semua soal / Σ bobot semua soal × 100 (proporsional bobot, bukan 50:50 objektif:esai). Nilai final dikunci/publish menjadi ledger.
6. **Nilai & laporan:** rekap per kelas/paket, nilai per kompetensi, analisis butir soal, berita acara, kartu peserta, daftar hadir, export CSV/PDF.

### 4.3 Aturan khusus

| Situasi | Aturan (asumsi — perlu persetujuan, lihat Bab 24) |
|---|---|
| Pause jadwal | Membekukan timer semua siswa (asumsi; beda dengan putus per-siswa yang timer tetap jalan) |
| Force finish | Per siswa / per jadwal; oleh proktor/admin |
| Unblock | Hanya Proktor/Admin yang mengeksekusi; Pengawas hanya mengusulkan. Opsi: lanjut (sisa waktu jalan terus) / reset (hapus jawaban, ulang dari awal) / tetap diblokir + akhiri |
| Isian Singkat | Default tidak case-sensitive + trim spasi; opsi `case_sensitive` / `absolute_eval` per soal |
| Pengurangan poin salah | Default tidak ada |
| Remedial | Nilai remedial dicatat terpisah; nilai asli tidak ditimpa |
| Tryout | Boleh tanpa token (asumsi); pembahasan boleh langsung tampil; nilai tidak resmi; tanpa label LULUS/TIDAK LULUS kaku |
| UKK | Teori (CBT) + praktik (rubrik, Penguji UKK); bobot default teori 40% + praktik 60% (asumsi, dapat diubah); kelulusan per unit kompetensi, bukan hanya total |
| Salah kunci ketahuan saat ujian berjalan | JANGAN ubah kunci di bank terkunci. Catat di berita acara, lanjutkan ujian, perbaiki lewat koreksi pasca-ujian seragam untuk semua peserta (persetujuan Admin + jejak audit) |
| Soal bocor | Kunci bank, arsipkan, buat bank pengganti dari duplikat, buat jadwal pengganti |

### 4.4 Perbedaan per jenis ujian

| Jenis | Ciri utama |
|---|---|
| **STS** | Sumatif tengah semester; campuran soal wajib + acak berstrata; urutan soal + opsi diacak |
| **SAS** | Sumatif akhir semester; kisi-kisi ketat, manual/campuran; kontrol kisi-kisi diutamakan |
| **USBK** | Ujian Sekolah bermoda CBT (POS sekolah); skala besar; kesetaraan antar sesi penting; pengacakan isi per siswa dibatasi agar kesulitan total setara |
| **UKK** | Teori + praktik + sikap; per unit kompetensi; asesor internal + eksternal; sertifikat hanya bagi yang lulus (≥70) |
| **Tryout** | Latihan/diagnosis; acak berstrata penuh; pembahasan boleh langsung tampil; nilai tidak resmi |

---

## 5. Bank Soal, Kurikulum & Kisi-kisi

### 5.1 Hierarki data

```
Tahun Ajaran (mis. 2026/2027, Semester Ganjil/Genap)
  └─ Jurusan / Konsentrasi Keahlian (mis. PPLG, TJKT, DKV)
       └─ Kelas / Rombel (mis. X PPLG 1, XI TJKT 2, XII DKV 1)
            │   Peserta terdaftar di SATU rombel aktif per tahun ajaran
            └─ Mapel
                 ├─ Mapel Umum (berlaku semua jurusan)
                 └─ Mapel Produktif / Kejuruan (terikat satu jurusan)
                      └─ Bank Soal (satu mapel bisa punya banyak bank)
                           └─ Soal (satu bank punya banyak soal)
                                └─ Opsi / Pasangan / Urutan / Rubrik (tergantung tipe)
                                     └─ dipakai oleh → Paket Ujian → Jadwal Ujian → Jawaban Peserta → Hasil
```

**Yang baru di v2 vs CBT-Go:** entitas Tahun Ajaran/Semester, Tingkat, Rombel, Kisi-kisi, Paket Ujian, Kompetensi (tabel referensi, bukan teks bebas), Stimulus bersama, Media soal, Level kesulitan/kognitif, Status soal, Versi/snapshot soal, UKK (skema, unit kompetensi, tugas praktik, rubrik, bukti).

**Aturan kunci:**
- Satu bank soal hanya untuk **satu mapel + satu tingkat**. Bank gabungan lintas mapel dibuat di level Paket, bukan Bank.
- Mapel Produktif wajib punya `jurusan_id`; Mapel Umum wajib `jurusan_id` kosong. Ini mencegah soal produktif tertukar jurusan.
- Satu siswa hanya aktif di satu rombel per semester. Riwayat pindah rombel/naik kelas disimpan (jangan menimpa), agar nilai tahun lalu tetap merujuk ke rombel lama.
- Kode mapel unik per sekolah. Daftar jurusan diambil dari Dapodik sekolah, bukan ditebak.
- **Tingkat vs Rombel dipisah:** Tingkat = X/XI/XII (Fase E untuk X; Fase F untuk XI–XII). Rombel = kelas fisik/administratif (`XI PPLG 1`).

### 5.2 Atribut soal yang dibutuhkan

| Atribut | Wajib? | Keterangan |
|---|---|---|
| ID soal, bank soal, nomor urut di bank | Wajib | Nomor urut = urutan penyusunan, bukan urutan tampil ujian |
| Tipe soal | Wajib | Salah satu dari 16 (Bab 6) |
| Stimulus | Opsional (wajib untuk tipe stimulus) | Teks/bacaan/kasus/data; bisa dipakai bersama oleh beberapa soal |
| Teks soal | Wajib | Satu pertanyaan yang jelas; hindari soal ganda dalam satu butir |
| Media | Opsional | Gambar/audio/video/tabel/grafik/rumus; rumus disimpan sebagai teks terstruktur, bukan hanya gambar; gambar wajib teks alternatif |
| Opsi jawaban | Wajib untuk objektif | PG default 4 opsi (A–D), dapat diubah per bank; PG 3–6 opsi |
| Kunci jawaban | Wajib untuk auto-grade | Per tipe (Bab 6) |
| Bobot / skor maksimal | Wajib, default 1 | Esai/studi kasus biasanya lebih besar dari PG |
| Poin pengurang | Opsional, default 0 | Tidak ada penalti jawaban salah kecuali sekolah meminta eksplisit + diumumkan ke siswa |
| Level kesulitan | Wajib | Mudah / Sedang / Sulit (diisi penulis, lalu dikalibrasi dari data) |
| Indikator / kisi-kisi | Wajib | Merujuk ke satu baris kisi-kisi; soal tanpa indikator tidak boleh masuk paket resmi |
| Kompetensi (CP/TP) | Wajib | Merujuk ke tabel Kompetensi (ID, bukan teks bebas); satu kompetensi utama + maks. 2 pendukung (pendukung hanya untuk analisis, tidak dihitung ganda) |
| Level kognitif | Wajib | Bloom revisi C1–C6 (Mengingat s.d. Mencipta); tampilan boleh disederhanakan L1/L2/L3 |
| Estimasi waktu | Opsional | Detik/menit perkiraan; bantu hitung durasi paket realistis |
| Status soal | Wajib | Draft / Aktif / Nonaktif / Perlu Revisi / Diarsipkan |
| Pembuat & tanggal, riwayat revisi/versi | Wajib (sistem) | Jika soal diubah setelah pernah dipakai ujian, versi lama tetap tersimpan untuk audit hasil lama |
| Catatan validator | Opsional | Hasil telaah (layak / revisi / tolak) |
| Statistik soal | Otomatis | Tingkat kesukaran empiris, daya beda, analisis pengecoh — dihitung dari hasil, bukan diisi manual |

### 5.3 Kisi-kisi

Satu kisi-kisi = satu rencana ujian untuk satu mapel + satu tingkat + satu jenis ujian + satu semester. Kolom: No. | Elemen/Materi | CP (rujukan) | TP/Indikator Soal | Level kognitif | Tipe soal | Jumlah soal | Level kesulitan | Bobot per soal | Soal terpilih.

**Istilah Kurikulum Merdeka (disederhanakan untuk CBT):** CP (Capaian Pembelajaran, akhir Fase E/F) → Elemen → TP (Tujuan Pembelajaran) → ATP (Alur TP) → Indikator soal → Kisi-kisi. **Profil Pelajar Pancasila tidak dinilai lewat soal CBT objektif** (dinilai lewat observasi/projek). CP resmi diambil dari dokumen Kurikulum Merdeka (Kemendikbudristek/BSKAP) + CP Konsentrasi Keahlian, lalu disesuaikan sekolah menjadi TP/ATP — CBT tidak boleh mengarang CP. Jika sekolah belum punya TP/ATP, pembuatan kisi-kisi harus didahului penyusunan TP/ATP.

**Pemetaan soal → kompetensi:** satu soal → satu kompetensi utama (dipakai hitung nilai per kompetensi) + satu baris kisi-kisi (wajib). Soal tanpa kompetensi/indikator boleh ada di bank Draft/Latihan, tetapi sistem harus menolak memasukkannya ke paket ujian resmi + memberi tahu guru alasannya.

### 5.4 Lock bank soal (2 lapis)

| Lapis | Pemicu | Siapa bisa membuka | Efek |
|---|---|---|---|
| **Lock Otomatis (System Lock)** | Bank dipakai jadwal berstatus Draft/Aktif/Paused (belum Selesai) ATAU ada peserta sedang/sudah mengerjakan | Tidak ada — terbuka otomatis saat semua jadwal terkait Selesai + hasil final | Mencegah perubahan apa pun pada soal/kunci/bobot/opsi |
| **Lock Manual (Admin Lock)** | Admin atau penanggung jawab bank mengunci manual | Hanya Admin, atau guru pemilik bank + alasan tertulis (audit) | Mencegah tambah/edit/hapus soal; masih boleh lihat dan duplikat bank |

Lock berlaku pada **seluruh bank**, bukan per soal. Jika satu soal perlu diperbaiki saat bank terkunci: duplikat bank → perbaiki di salinan → buat jadwal baru dari salinan (susulan/remedial), bukan membuka bank yang sedang dipakai.

**Versioning/snapshot:** saat paket dibuat / jadwal dimulai, sistem menyimpan salinan beku (snapshot) soal + opsi + kunci + bobot yang dipakai jadwal itu. Hasil, koreksi, analisis selalu merujuk ke snapshot, bukan ke soal yang sedang diedit. Snapshot disimpan per paket/jadwal (bukan per siswa), karena semua siswa dalam satu jadwal mengerjakan soal yang sama (kecuali urutan tampil).

**Validasi & telaah soal (quality gate):** soal baru berstatus Draft sampai lolos telaah substansi (sesuai indikator, kunci benar, pengecoh masuk akal), bahasa (jelas, tidak ambigu, tidak membocorkan kunci), konstruksi (satu soal satu pertanyaan, opsi homogen, stimulus relevan), dan kunci (dicek orang kedua, bukan penulis). Minimal 1 penelaah selain penulis untuk bank ujian resmi.

### 5.5 Import soal

**Prinsip:** import adalah asisten pengetikan, bukan jalan pintas validasi. Semua soal hasil import masuk berstatus Draft, wajib telaah. Setiap import: validasi dulu (dry-run) → laporan kesalahan per baris/soal → hanya simpan jika pengguna setuju → tidak pernah simpan sebagian (semua-berhasil atau semua-gagal, atau "lewati baris yang salah" eksplisit). Default: tambah baru saja (tidak mengubah soal yang sudah ada kecuali mode "perbarui" eksplisit). Template import disediakan dari dalam aplikasi.

**Import CSV (v2):** satu baris = satu soal. Kolom: `kode_bank, nomor_urut, tipe_soal, stimulus, pertanyaan, media, opsi, pasangan, kunci_isian, rubrik, bobot, kesulitan, level_kognitif, kode_indikator, kode_kompetensi, estimasi_detik, catatan`. Aturan: UTF-8, delimiter koma, quoting standar; maks. 500 baris per berkas; validasi per baris (tipe valid, jumlah opsi sesuai tipe, kunci sesuai, bobot > 0, kode indikator/kompetensi dikenali); laporan: berhasil + gagal (nomor baris + alasan) + pratinjau 3 soal pertama. Format CSV v2 tetap kompatibel-mundur dengan CSV lama CBT-Go (4 kolom) — berkas lama dibaca dengan nilai default (kesulitan = Sedang, kognitif = C2, kompetensi/indikator kosong → Draft "belum terpetakan").

**Import Word (.docx):** satu dokumen = satu bank soal. Kontrak format: judul dokumen (kode bank + mapel + tingkat); setiap soal dimulai penanda bernomor `Soal 1. [PG] [Sedang] [C3] [Kompetensi: ...] [Indikator: ...]`; stimulus dengan penanda `Stimulus:`; opsi PG sebagai daftar berhuruf `A. ...`; kunci ditulis eksplisit di akhir soal (`Kunci: B`) — **bukan** lewat warna/tebal (hilang/ambigu saat parsing); tabel Word dibaca sebagai stimulus tabel / pasangan menjodohkan (jika ditandai `Menjodohkan:`); gambar diimpor sebagai media; rumus Equation Editor diimpor sebagai rumus terstruktur bila memungkinkan, jika tidak sebagai gambar + peringatan; esai/rubrik ditulis sebagai `Rubrik:` + deskripsi + skor. **Batasan jujur:** Word tidak bisa diandalkan untuk urutan acak opsi, menjodohkan kompleks, mengurutkan banyak butir, studi kasus multi-sub-soal — untuk itu CSV atau input manual lebih andal. Dokumen yang tidak mengikuti kontrak format **ditolak dengan laporan kesalahan**, bukan "ditebak" mesin. Setelah import Word, guru wajib pratinjau 100% soal sebelum bank bisa berstatus Siap. **Import Google Docs tidak didukung di v2** (butuh API Google + izin dokumen; guru ekspor ke .docx dulu) — fitur tahap lanjutan.

**Perpustakaan media:** media disimpan terpusat per sekolah (bukan menempel di satu soal) agar bisa dipakai banyak soal tanpa duplikasi. Batas ukuran (asumsi, dapat diubah di settings): gambar 2 MB, audio 5 MB, video 20 MB.

### 5.6 Pemilihan soal ke paket

| Mode | Cara kerja | Cocok untuk |
|---|---|---|
| **Manual murni** | Guru pilih satu per satu mengacu kisi-kisi | USBK, SAS, UKK-Teori (kisi-kisi ketat) |
| **Acak berstrata (stratified random)** | Sistem mengacak di dalam strata kisi-kisi (per baris: ambil N soal acak dari soal yang memenuhi indikator + tipe + kesulitan sama) | Tryout, latihan, kuis, STS dengan bank besar, remedial |
| **Campuran** | Sebagian soal dikunci manual (wajib, mis. studi kasus/esai), sisanya acak berstrata | SAS/USBK dengan soal wajib + variasi |
| **Acak murni (tanpa strata)** | Ambil N soal acak dari seluruh bank tanpa kisi-kisi | **Tidak disediakan untuk ujian resmi** — hanya mode latihan dengan label jelas |

**Syarat acak berstrata:** per baris kisi-kisi, bank harus punya minimal **3× jumlah soal yang diminta** (asumsi); atribut soal wajib lengkap; soal yang sudah dipakai paket aktif tidak boleh dipakai paket lain yang jadwalnya bersamaan; soal bermasalah (Perlu Revisi/Nonaktif/flag daya beda negatif) tidak ikut undian.

**Pengacakan per siswa (2 lapis):**
1. **Isi paket per siswa** — soal mana yang didapat tiap siswa (dari bank, dalam strata kisi-kisi). Setiap siswa mendapat paket yang memenuhi kisi-kisi yang sama (jumlah per indikator, sebaran kesulitan, total bobot sama), tetapi butirnya berbeda. Seed acak disimpan per siswa+jadwal agar dapat direproduksi untuk audit.
2. **Urutan tampil + urutan opsi** — urutan soal di layar + urutan opsi (A–D) diacak per siswa. Soal satu stimulus tetap berurutan dan berdekatan. Opsi "Semua jawaban benar/salah" / "Tidak ada jawaban yang benar" tetap di posisi terakhir.

Pengacakan tidak mengubah isi soal, kunci, atau bobot. Paket per siswa **dibekukan** saat siswa mulai mengerjakan (snapshot). Untuk USBK/SAS, pengacakan isi per siswa dibatasi agar kesulitan total setara.

### 5.7 Kebutuhan khusus soal UKK

UKK bukan ujian mapel biasa: penilaian **kompeten / belum kompeten per unit kompetensi**, bukan hanya nilai 0–100; dua komponen wajib **teori + praktik**; penilai praktik adalah **asesor** (guru internal dan/atau asesor eksternal industri/LSP); bukti (evidence) praktik harus disimpan.

| Entitas UKK | Isi |
|---|---|
| Skema / Paket UKK | Skema yang diujikan, jurusan, tahun ajaran, daftar unit kompetensi, ambang per unit, penguji, jadwal |
| Unit Kompetensi | Kode unit, judul, elemen/kriteria unjuk kerja (KUK), standar acuan (SKKNI/LSP/industri) |
| Soal Teori UKK | Soal dari bank soal yang dipetakan ke **unit kompetensi** (bukan hanya CP mapel); penekanan PG, PG Kompleks, Studi Kasus, Esai Singkat |
| Tugas Praktik | Deskripsi, skenario/brief, batas waktu, alat/bahan, hasil yang dikumpulkan, kriteria penilaian |
| Rubrik / Lembar Observasi | Aspek Proses (30%) + Produk (50%) + Sikap (20%) — bobot default, dapat diubah per skema; skor per aspek + ambang kompeten + catatan asesor |
| Bukti (Evidence) | Berkas diunggah siswa/asesor: kode, dokumen, foto produk, video demo, laporan; terikat ke tugas + siswa + asesor; tanpa bukti, nilai praktik tidak sah |
| Hasil UKK | Per siswa, per unit: kompeten / belum kompeten; nilai teori; nilai praktik; catatan asesor; rekomendasi (lulus / remedial / uji ulang) |

**Aturan UKK:** ambang kelulusan teori berlaku per unit (bukan hanya total) — siswa yang totalnya lulus tetapi satu unitnya gagal dinyatakan belum kompeten pada unit itu dan menjalani uji ulang unit tersebut. Siswa yang belum kompeten pada satu unit hanya mengulang unit itu, bukan seluruh UKK. Bank UKK dikunci lebih ketat (manual sejak disahkan, tidak boleh dibuka tanpa persetujuan Admin + catatan). Penguji eksternal: akun terbatas (hanya siswa + tugas yang ditugaskan, hanya isi rubrik). **CBT menyediakan data dan draf sertifikat, bukan menerbitkan sertifikat hukum sendiri** — sertifikat resmi diterbitkan lembaga berwenang (BNSP/LSP/DUDIKA/sekolah). Berita acara UKK per skema (daftar siswa, unit, hasil, asesor, tanda tangan). Mode semi-daring praktik: rubrik bisa diunduh/dicetak sebagai cadangan kertas + diisi offline lalu disinkronkan/dientri susulan (dengan catatan "dientri dari kertas, tanggal entri X"); unggah bukti bertahap; tidak ada streaming video real-time.

Alur UKK: tetapkan skema + unit + asesor + jadwal → susun bank teori + tugas praktik + rubrik, kunci bank → siswa kerjakan teori via CBT → siswa kerjakan praktik di lab/bengkel, asesor isi rubrik + verifikasi bukti → sistem hitung hasil per unit (teori + praktik → Kompeten/Belum Kompeten) → uji ulang per unit bila perlu → rekap + berita acara + data sertifikat.

---

## 6. 16 Tipe Soal — Koreksi Penting + Daftar + Penilaian

### 6.1 KOREKSI: tidak ada "16 tipe soal standar Kemendikbud"

**Temuan riset (paling penting, baca dulu): TIDAK DITEMUKAN dokumen resmi Kemendikbud/Pusmendik/BSNP yang menetapkan "16 tipe soal".**

Sumber patokan yang ditemukan:

| Dokumen | Isi terkait bentuk soal |
|---|---|
| **Framework AKM** (Pusat Asesmen Pendidikan, BSKAP, Kemendikbudristek, edisi revisi 2023, Lampiran 1 hlm. 147–149) | Menetapkan **hanya 5 bentuk soal**: (1) Pilihan Ganda; (2) Pilihan Ganda Kompleks; (3) Menjodohkan; (4) Isian/Jawaban Singkat; (5) Uraian/Esai. Framework **tidak mengatur** pengacakan, bobot per soal, konversi ke 0–100, atau rumus daya beda/pengecoh (yang terakhir dari literatur evaluasi, mis. pengecoh berfungsi bila dipilih ≥5% — bukan standar Kemendikbud) |
| **TKA** (Perkaban No. 47/2025 — sumber sekunder, PDF tidak berhasil dibuka langsung) | Hanya 2 bentuk (PG + PGK) |
| **UN 2019** | Semua PG kecuali Matematika SMA/SMK (PG + isian) |
| **USBN 2019** | PG + uraian (90/10) |
| **POS Ujian Sekolah** | Bervariasi antar daerah (bukan standar nasional tunggal) |

**Situs Extraordinary CBT hanya mengklaim 6 label** (PG tunggal & kompleks, benar-salah, menjodohkan, isian singkat, esai). Nama 16 tipe Exo tidak terverifikasi (situs Exo hanya 6 label, dump tidak memuat pemetaan nama `tipe_soal`). **`code01–code16` di dump Exo adalah kode kelompok kompetensi / bagian penilaian, BUKAN daftar nama 16 tipe soal** — kajian sebelumnya yang menyebut "16 kode soal" adalah salah tafsir. Koreksi ini dicatat eksplisit.

**Ketidakpastian yang dinyatakan jujur oleh researcher:** nama 16 tipe Exo tidak terverifikasi; PDF Perkaban TKA belum dibuka langsung; skema PGK 2-1-0 konflik sumber (Framework diprioritaskan); POS dinas sumber sekunder.

**Rekomendasi yang diusulkan (perlu keputusan user, Bab 24):** pakai **5 tipe resmi AKM sebagai inti** + Benar-Salah sebagai submodel PGK; stimulus jadi atribut soal (bukan tipe); jika tetap ingin 16 pilihan di UI, label jujur **"5 tipe standar AKM + 11 varian/ekstensi aplikasi"**, jangan klaim "16 tipe standar Kemendikbud" di dokumentasi/marketing v2. **Tidak ada tipe resmi yang hilang** — 5 bentuk AKM sudah mencakup semuanya.

### 6.2 Pemetaan 16 nama CBT-Go terhadap 5 bentuk resmi AKM

| # | Nama di CBT-Go | Status terhadap standar AKM |
|---|---|---|
| 1 | Pilihan Ganda (PG) | ✅ Cocok penuh (bentuk resmi 1) |
| 2 | PG Kompleks | ✅ Cocok penuh (bentuk resmi 2) |
| 3 | Benar–Salah | Submodel PGK (bentuk resmi 2) |
| 4 | Menjodohkan | ✅ Cocok penuh (bentuk resmi 3) |
| 5 | Isian Singkat | ✅ Cocok penuh (bentuk resmi 4) |
| 6 | Esai / Uraian | ✅ Cocok penuh (bentuk resmi 5) |
| 7 | Mengurutkan | Tipe generik non-Kemendikbud (tipe Moodle) — ekstensi aplikasi |
| 8 | Setuju–Tidak Setuju | Label survei — dalam ujian harus punya kunci (bukan survei); jika survei, ditandai "tidak dinilai" |
| 9 | PG Stimulus Gambar | Varian penyajian — stimulus adalah atribut, bukan tipe |
| 10 | PG Berbasis Tabel | Varian penyajian — stimulus adalah atribut, bukan tipe |
| 11 | PG Berbasis Grafik | Varian penyajian — stimulus adalah atribut, bukan tipe |
| 12 | Isian Panjang | Salah klasifikasi — ini uraian (bentuk resmi 5) |
| 13 | Menjodohkan Gambar | Varian penyajian — stimulus adalah atribut, bukan tipe |
| 14 | PG Kompleks Multi-Benar | Duplikat PGK (bentuk resmi 2) — bedakan lewat format: tipe 2 berbasis pernyataan, tipe 14 berbasis opsi multi-pilih |
| 15 | Esai Singkat | Submodel uraian (bentuk resmi 5) — bedakan dari Esai lewat panjang jawaban yang diharapkan dan bobot |
| 16 | Studi Kasus | Paket stimulus — 1 stimulus + N butir (masing-masing punya tipe sendiri) |

### 6.3 Definisi + penilaian per bentuk resmi AKM

| Bentuk resmi | Definisi | Penilaian |
|---|---|---|
| **(1) Pilihan Ganda** | Pilih 1 dari beberapa opsi. Jumlah opsi per jenjang: kelas 1–3: 3 opsi; 4–9: 4 opsi; 10–12: 5 opsi (Framework) | Skor 1 bila benar, 0 bila salah; otomatis |
| **(2) Pilihan Ganda Kompleks** | 2 model: MCMA (checkbox, pilih semua yang benar) + PGK kategori (Benar/Salah/Ya-Tidak per pernyataan) | Framework: 1/0 all-or-nothing. Varian pelatihan 2-1-0 untuk >5 pernyataan + >2 kategori (konflik sumber — Framework diprioritaskan). Di v2, aturan skor disimpan **per soal** (bukan global) agar bank lama tidak berubah diam-diam; default yang diusulkan: **proporsional** (poin = bobot × pilihan benar yang dipilih / jumlah kunci, dikurangi jika ada pilihan salah, minimal 0) karena lebih adil dan lebih membedakan kemampuan |
| **(3) Menjodohkan** | Pasangkan premis (kiri) dengan respon (kanan); jumlah respon harus lebih banyak dari premis | 1/0 all-or-nothing **per butir** (bukan per pasangan) menurut Framework. Di v2, untuk menjodohkan/mengurutkan, default yang diusulkan: per pasangan/posisi benar diberi poin (skor = bobot × pasangan benar / total pasangan) — perlu diputuskan apakah mengikuti Framework ketat atau skema proporsional (Bab 24) |
| **(4) Isian/Jawaban Singkat** | Satu bentuk: isian = kalimat rumpang; jawaban singkat = kalimat tanya | 1/0 otomatis. Framework tidak mengatur case sensitivity / toleransi spasi / kunci alternatif. Di v2: default tidak peka huruf besar/kecil, spasi berlebih diabaikan, satuan dipisah ("10 cm" vs "10"), daftar jawaban alternatif didukung; opsi `case_sensitive` / `absolute_eval` per soal, default MATI |
| **(5) Uraian/Esai** | Jawaban bebas terstruktur | Manual oleh guru, wajib pedoman penskoran (rubrik). Jawaban kompleks 2-1-0, sederhana 1-0 (Framework). Di v2: skor 0 s.d. bobot, rubrik disarankan, komentar opsional, opini AI opsional tahap lanjut (saran, bukan nilai resmi) |

### 6.4 Penilaian varian/ekstensi (16 nama CBT-Go)

| Tipe | Cara jawab | Cara nilai |
|---|---|---|
| PG (1, 9, 10, 11) | Pilih 1 opsi | Benar = bobot penuh; salah/kosong = 0; tanpa penalti |
| PG Kompleks (2, 14) | Pilih beberapa | Proporsional (default usulan) atau semua-benarnya-saja — disimpan per soal |
| Benar-Salah (3) | Benar/Salah per pernyataan | Per pernyataan; skor = bobot × pernyataan benar / total pernyataan |
| Menjodohkan (4, 13) | Pasangkan kiri–kanan | Per pasangan benar; urutan kanan diacak saat tampil |
| Isian Singkat (5) | Ketik singkat | Otomatis dengan toleransi (6.3) |
| Esai/Uraian (6), Isian Panjang (12), Esai Singkat (15) | Ketik jawaban | Manual dengan rubrik; bedakan lewat panjang jawaban yang diharapkan + bobot |
| Mengurutkan (7) | Urutkan langkah/peristiwa | Default: skor per posisi benar; alternatif semua-benarnya-saja (disimpan per soal) |
| Setuju–Tidak Setuju (8) | Setuju/Tidak Setuju per pernyataan | Jika dinilai: seperti Benar-Salah. Jika survei: skor 0, tidak masuk nilai, ditandai eksplisit |
| Studi Kasus (16) | Rangkaian sub-soal dari satu kasus | Campuran per sub-soal; nilai studi kasus = jumlah nilai sub-soal |

**Aturan umum:** soal kosong (tidak dijawab) = skor 0, dicatat sebagai "kosong" (dibedakan dari "salah" di analisis butir soal). Tanda ragu-ragu tidak memengaruhi nilai.

---

## 7. Anti-Curang & Keamanan

### 7.1 Matriks ancaman (ringkas — 14 ancaman T1–T14)

| # | Ancaman | Pencegahan utama |
|---|---|---|
| T1 | Pindah tab / minimize / Alt-Tab | Deteksi `visibilitychange` + `blur` dengan **debounce 5 detik** (agar tidak double-count); 1x peringatan + counter 1; 2x auto-block |
| T2 | Copy-paste | Nonaktifkan copy-paste di halaman ujian (penghambat, bukan jaminan) |
| T3 | Screenshot / foto layar | Tidak bisa dicegah penuh di web; watermark + larangan + berita acara |
| T4 | Login ganda / perangkat lain | Device lock (deteksi) + satu sesi aktif; sesi kedua ditolak/ditendang (aturan tertulis) |
| T5 | Token bocor / ditebak | Token 6 karakter dari 32 simbol ≈ 1 miliar kombinasi — cukup kuat **HANYA** jika ada rate limit + masa berlaku singkat + revoke |
| T6 | Berbagi jawaban | Pengacakan isi + urutan per siswa (Bab 5.6) |
| T7 | Penyamaran (joki) | Pengawas cek foto kartu login vs wajah; daftar hadir; tidak bisa dicegah penuh oleh software |
| T8 | Manipulasi waktu/jawaban | Timer berjangkar server; server tolak save/submit setelah deadline; nilai dihitung server (nilai kiriman browser diabaikan — kritis) |
| T9 | Intip kunci via DevTools | **Jangan kirim kunci jawaban ke browser** (endpoint soal ujian tidak boleh berisi kunci, poin benar/salah, atau pembahasan) |
| T10 | Kolusi pengawas/admin | Pemisahan tugas (Bab 11), audit log, tidak ada hapus audit log |
| T11 | Gangguan teknis palsu | Log insiden, verifikasi proktor, aturan unblock tertulis |
| T12 | Bocor antar gelombang | Soal yang sudah dipakai paket aktif tidak dipakai paket lain yang jadwalnya bersamaan; rasio bank 3× |
| T13 | Brute force token / login | Rate limit: login maks 5 gagal per akun per 15 menit → ke-6 → 429 + kunci 15 menit; token: ambang → rate limit/jeda/blokir sementara |
| T14 | Ubah nilai / akses data orang lain | Penegakan permission di server (403 + catat), anti-IDOR, 404 untuk resource yang tidak berhak (agar tidak membocorkan keberadaan) |

### 7.2 Batasan jujur (L1–L10) — jangan diklaim lebih

Web app **TIDAK bisa cegah penuh**: HP kedua; screenshot/foto; menoleh/berbisik; penyamaran; DevTools; copy-paste non-standar; listrik/jaringan padam; kolusi; hafalan antar gelombang; AI di perangkat lain. **Kunci:** jangan kirim kunci jawaban ke browser; timer berjangkar server; server tolak save/submit setelah deadline; token kuat hanya dengan rate limit + masa berlaku singkat. Device lock web hanya deteksi/penghambat (cookie bisa dihapus); kunci kuat butuh kiosk / Safe Exam Browser / MDM. Fullscreen di web hanya deteksi (Esc selalu bisa keluar). Kiosk hanya jika sekolah siap.

### 7.3 Peran pengawas dalam anti-curang

Monitor, teguran, force finish, pause/resume (proktor), unblock wajib alasan (proktor mengeksekusi, pengawas mengusulkan), reset login, revoke token, ruangan, berita acara.

### 7.4 Audit log (wajib — belum ada di CBT-Go)

Dicatat: login/logout, masuk ujian, submit, pelanggaran, block/unblock, token (generate/rilis/revoke), perubahan nilai, export, percobaan akses ditolak. Format: waktu (Asia/Jakarta + UTC) + aktor (ID + nama + role aktif) + aksi + target (ID + label) + hasil + IP/perangkat + alasan (wajib untuk aksi sensitif) + sebelum/sesudah. Append-only, hanya admin boleh lihat, retensi usulan 1 semester lalu arsip. Tidak menyimpan password/token plaintext/kunci jawaban di log.

### 7.5 Prioritas keamanan

1. **Kritis sebelum produksi:** CSRF + hapus hapus via GET + wajib ganti password pertama + hapus/nonaktifkan akun demo + rate limit + cookie `Secure` + HTTPS.
2. Anti-curang inti (Bab 7.1).
3. Pengawasan (monitor, insiden, berita acara).
4. Penguatan (indeks, pool DB, sesi tahan restart, antrean offline).
5. Kiosk — hanya jika sekolah siap.

### 7.6 Keamanan data (UU PDP — lihat Bab 12)

Hashing password (bcrypt), RBAC, audit log, backup terenkripsi, pengaturan retensi + penghapusan/anonimisasi oleh admin sekolah, prosedur laporan insiden. Log tidak boleh memuat password/token/jawaban siswa.

---

## 8. Laporan, Nilai & Analisis Butir Soal

### 8.1 Rekap nilai

**Per kelas/paket:** NISN/NIS, nama, status hadir, nilai objektif / manual / akhir, KKTP/KKM, status LULUS/TIDAK LULUS/BELUM DINILAI/TIDAK HADIR, nilai per kompetensi, pelanggaran, waktu. **Agregat:** rata-rata + N pembagi eksplisit, median, tertinggi/terendah/rentang, simpangan baku, jumlah + % lulus, histogram.

**Aturan:** jangan campur N (pembagi harus eksplisit); nilai sementara berlabel SEMENTARA; nilai final dikunci/publish menjadi ledger (perubahan setelah publish hanya revisi tercatat). UKK: teori + praktik masing-masing ≥ ambang (asumsi, bukan hanya rata-rata gabungan). Tryout: tanpa label LULUS/TIDAK LULUS kaku.

### 8.2 Nilai akhir & nilai per kompetensi

- **Nilai akhir** = Σ poin semua soal / Σ bobot semua soal × 100 (proporsional bobot, bukan 50:50 objektif:esai) + tampilkan kontribusi komponen.
- **Nilai per kompetensi (k)** = Σ poin didapat pada soal kompetensi k / Σ bobot soal kompetensi k × 100. Satu soal → satu kompetensi utama (asumsi; multi-kompetensi ditunda). Soal esai yang belum dikoreksi = "belum dinilai" (bukan 0). Peringatan jika kompetensi hanya diwakili <3 soal, dan jika >20% bobot paket tanpa kompetensi. Ketuntasan per kompetensi ≥ KKTP/KKM.
- **Nilai remedial** dicatat terpisah; nilai asli tidak ditimpa; riwayat remedial (nilai awal, nilai remedial, tanggal, TP yang diremedi) tersimpan; aturan gabungan nilai remedial dapat dikonfigurasi sekolah.

### 8.3 Dokumen resmi

| Dokumen | Aturan |
|---|---|
| **Berita Acara Ujian (BAU)** | Per jadwal + sesi + ruangan, wajib PDF; isi: hadir/tidak hadir, kejadian (pelanggaran, kendala teknis, soal bermasalah), tanda tangan proktor + pengawas + kepala sekolah/wakil; BAPU (Berita Acara Pelanggaran) untuk pelanggaran |
| **Kartu Peserta** | Wajib PDF; **satu kartu per siswa untuk semua ujian (model 1-KARTU, berlaku sejak 17 Okt 2026 — menggantikan "satu kartu per siswa per jenis ujian")**: identitas siswa sekali + tabel SEMUA jadwal kelasnya (per baris: nomor peserta per jadwal + mapel + paket + tanggal + jam + ruangan); PDF dibuat di **backend** (Go, gofpdf — `GET /api/v1/kartu/siswa/:siswaId/pdf` + `GET /api/v1/kartu/pdf` gabungan staf); **tanpa `window.print()`** di alur kartu; **token TIDAK dicetak di kartu** (token = kredensial, kartu sering difoto/ditinggalkan); cetak ulang tidak mengubah nomor kartu; QR terpisah acak |
| **Daftar Hadir** | Per sesi per ruangan, wajib PDF |
| **Sertifikat UKK** | Data + draf dari CBT; penerbitan resmi oleh lembaga berwenang |

### 8.4 Analisis butir soal

Per jadwal (asumsi: hanya peserta yang sudah submit; esai belum dikoreksi = "belum dinilai" bukan 0):

| Metrik | Rumus / aturan | Interpretasi |
|---|---|---|
| **Tingkat kesukaran (P)** | P = B/N (B = jumlah benar, N = jumlah peserta) | >0,80 terlalu mudah; 0,30–0,70 ideal; <0,20 terlalu sulit; P=0/1 wajib tinjau |
| **Daya beda (D)** | D = (BA−BB)/n, kelompok atas/bawah 27% | ≥0,40 baik; 0,30–0,39 cukup; 0,20–0,29 revisi; <0,20 buruk; negatif = kunci dicurigai (prioritas tertinggi) |
| **Pengecoh** | Distribusi pilihan per opsi | Berfungsi bila dipilih >5% + lebih banyak kelompok bawah; 0% = opsi mati; lebih banyak kelompok atas = menjebak/ambigu; kunci lebih dipilih bawah = kunci dicurigai |

**Syarat N:** N<10 → D tidak dapat dipercaya; N≥27 → metode 27% penuh. Di bawah itu, sistem tetap menghitung tetapi menampilkan peringatan eksplisit.

**Status soal pasca-analisis:** Pertahankan / Revisi / Buang-Arsipkan (jangan hapus permanen) / Perlu pengajaran ulang. Soal yang di-flag masuk status "Perlu Revisi" dan tidak boleh dipakai lagi sampai diperbaiki + ditelaah ulang. **Cakupan kisi-kisi pasca-ujian:** sistem menunjukkan apakah realisasi sesuai kisi-kisi (mis. kisi-kisi menuntut 3 soal C4, realisasi 1 soal C4 → pelanggaran kisi-kisi yang harus dijelaskan).

**Koreksi esai:** skor 0 s.d. bobot, rubrik disarankan, komentar opsional, opini AI opsional tahap lanjut (saran, bukan nilai resmi), publish setelah semua esai terkoreksi.

### 8.5 Export

- **CSV** untuk data mentah/olah lanjut (UTF-8, header Indonesia, nama file berisi identitas + tanggal); karakter khusus tidak merusak kolom.
- **PDF** untuk dokumen resmi (kop, nomor dokumen, tanggal cetak, tanda tangan, nomor halaman, label DRAF/FINAL/SEMENTARA).

### 8.6 Matriks privasi laporan

EVIL semua; EVAL hanya mapelnya; OPRA hanya status pelaksanaan + agregat tanpa nilai per siswa (kecuali tugas khusus); siswa hanya nilainya sendiri; tidak ada peringkat nama terbuka; export dicatat di log.

---

## 9. UX/UI — Standar "Lebih Cantik"

### 9.1 Masalah baseline CBT-Go (yang harus diperbaiki)

Semua soal sekaligus dalam satu daftar panjang (melelahkan, rawan salah klik); tidak ada navigasi soal; timer hanya teks kecil 20px; styling inline generik (system-ui, radius 6–10px); 13 link horizontal admin meluber di HP; tidak ada ringkasan sebelum submit; banner anti-curang mudah terlewat setelah scroll.

### 9.2 Design system (usulan — perlu persetujuan final)

Dua usulan palet masuk dari laporan (Analis UX/UI vs Frontend Dev) — **perlu diputuskan satu** (Bab 24):

| Elemen | Usulan A (Analis UX/UI) | Usulan B (Frontend Dev) |
|---|---|---|
| Primary | **#1D4ED8** (biru royal), gelap #1E3A8A | **#2563EB** (biru indigo), hover #1D4ED8 |
| Aksen | **#F59E0B** (amber) | — |
| Latar / card | #F1F5F9 / #FFFFFF | #F9FAFB / #FFFFFF |
| Teks | #0F172A / sekunder #475569 | #111827 / #374151 / #6B7280 |
| Garis | #E2E8F0 | #E5E7EB |
| Sukses / peringatan / bahaya / info | #16A34A / #D97706 / #DC2626 / #0284C7 | #16A34A/#F0FDF4 / #D97706/#FFFBEB / #DC2626/#FEF2F2 |

**Yang disepakati kedua laporan:** status selalu warna + bentuk/ikon + teks (bukan warna saja); kontras memenuhi standar; satu keluarga sans-serif (**Inter** atau **Plus Jakarta Sans** — perlu diputuskan); judul halaman 22–24px tebal; isi soal **16–17px** tinggi baris 1.5–1.7; timer 20–24px monospace/tabular; target sentuh minimal 44px; radius tombol/input/badge 8px, kartu/modal 12px, modal besar/bottom-sheet 16px; bayangan lembut; garis 1px; animasi halus ≤200ms; **ikon: satu set ikon garis tipis (react-icons / Lucide), bukan emoji** untuk ikon fungsional; emoji hanya untuk pesan kosong/ilustrasi ringan. **Dark mode TIDAK diprioritaskan** (backlog).

**Spacing:** satuan dasar 4px; padding kartu 16px HP / 20–24px desktop; jarak seksi 24px HP / 32px desktop; tinggi baris tabel minimal 48px; jarak antar opsi ujian minimal 12px; lebar isi maks 1280px admin / 1100px ujian.

### 9.3 Halaman ujian siswa (halaman paling kritis)

- **Satu soal per layar.** Header sticky ramping (nama ujian + timer + progress bar waktu). Timer besar, merah saat <5 menit + berdenyut, peringatan modal saat sisa 1 menit.
- **Panel navigasi soal:** desktop kolom kanan sticky; HP bottom sheet via tombol "Daftar Soal".
- **Warna status grid:** abu = belum; hijau + centang = sudah; kuning + "?" = ragu-ragu; biru/border tebal = sedang dibuka; merah hanya di ringkasan akhir untuk belum dijawab.
- Tombol Sebelumnya/Berikutnya; checkbox ragu-ragu; autosave + indikator "Tersimpan ✓".
- **Submit 2 langkah:** layar ringkasan review → modal konfirmasi final, peringatan merah jika ada soal kosong ("Soal yang kosong dihitung salah"). Tidak ada submit satu klik. Layar sukses setelah submit.
- HP: header 2 baris, timer besar, progress, soal + opsi card selebar penuh, tombol navigasi sticky bawah, tombol Selesaikan **TIDAK** sticky (agar tidak tertekan tidak sengaja); bar bawah sticky [☰ Daftar Soal] [Selesaikan] → bottom-sheet navigator.
- Perilaku: autosave per soal; timer selalu terlihat; navigasi tanpa reload; pindah tab → pelanggaran +1 → ke-2 auto-block → halaman `/blocked`; fokus keyboard; setelah submit, tombol back browser tidak boleh kembalikan ke soal; tutup tab lalu buka lagi saat ujian berjalan → kembali ke soal terakhir dengan sisa waktu server.
- Pindah tab tepat saat auto-submit → tidak ada nilai ganda/hilang/menggantung.

### 9.4 Responsif (mobile-first)

Breakpoint usulan: HP ≤640px, tablet 641–1024px, desktop ≥1025px. Navigasi admin: drawer/bottom nav HP, sidebar lipat tablet, sidebar tetap desktop. Tabel → card list HP. Header: hamburger di HP, search full-width di bawah header pada HP. Modal `mx-4` di HP. Form full-width HP. Dashboard stat cards 2 kolom HP → 4 desktop. Sticky buy bar tidak relevan (itu e-commerce) — untuk CBT: tombol navigasi ujian sticky bawah di HP.

**Uji responsif:** HP kecil 360px, HP umum 375–414px, tablet 768px portrait + landscape, laptop/desktop 1280px, desktop besar 1920px. Tanpa scroll horizontal; tombol penting terlihat/mudah dijangkau; timer selalu terlihat; soal + opsi panjang tidak meluber; gambar soal mengecil proporsional; tabel bisa dibaca di HP (scroll horizontal di dalam tabel saja atau kartu); form bisa diisi penuh di HP; modal tidak meluber; rotasi HP tidak menghilangkan jawaban / mereset timer; zoom 125–150% desktop tidak merusak layout ujian. Emulasi + minimal 1 HP Android asli + 1 tablet/laptop asli.

### 9.5 Aksesibilitas

Kontras (teks sekunder tidak lebih muda dari #475569); font soal minimal 16px; kontrol ukuran font soal (A−/A+); jangan matikan zoom browser; navigasi keyboard penuh (Tab, Spasi/Enter, panah); outline fokus biru ≥2px; skip link; label form terhubung (placeholder bukan pengganti label); satu `<h1>` per halaman; aria-live untuk perubahan penting.

### 9.6 Inspirasi visual

Fondasi "kelas modern bersih" + ujian "lembar ujian kertas ditingkatkan" + sentuhan "kuis hidup" pada opsi + monitor "ruang kontrol" + pesan "nada hangat".

**Prioritas UI:** (1) halaman ujian v2, (2) navigasi responsif, (3) design system, (4) monitor, (5) dashboard per role, (6) aksesibilitas dasar.

---

## 10. Infrastruktur & Operasional

### 10.1 Arsitektur runtime

1 binary Go + 1 database PostgreSQL di 1 host, di belakang 1 reverse proxy (Caddy/Nginx :80/:443 → Go 127.0.0.1:9988; PostgreSQL 127.0.0.1:5432 tidak diekspos; SSH dibatasi). **Koreksi brief:** kode `cbt-go` saat ini adalah Go + `html/template` server-rendered (embed, CSS inline, tanpa React/build step) — bukan Go + React. Untuk v2, perintah user adalah React — frontend build di-embed via `go:embed` (Opsi B1) atau dilayani dari folder `static/` (Opsi B2). **JANGAN** pisah frontend ke hosting/CDN terpisah.

### 10.2 Skala (300 siswa serentak — asumsi: 40 soal, 60 menit, teks)

Rata-rata <10 req/detik; puncak desain 50–100 req/detik; beban ringan untuk Go; bottleneck pool DB + kontensi tabel jawaban. Hardware: minimum (≤100 siswa) 2 vCPU/4 GB/50 GB SSD; direkomendasikan (300 siswa) **4 vCPU/8 GB/100 GB SSD**, LAN Gigabit + AP cukup. Data: 300 siswa × 40 soal = 12.000 baris jawaban ≈ 6 MB per jadwal; 100 jadwal/tahun <1 GB/tahun untuk data murni (yang membesarkan disk adalah media + backup).

**Pool DB:** naikkan `MaxOpenConns` 10 → **25–50**, `MaxIdleConns` 10–25, `ConnMaxLifetime` ~5 menit, `max_connections` ≥100, `shared_buffers` ~25% RAM; 500–1000 siswa → PgBouncer.

**Indeks yang perlu ditambah (v2):** `jadwals(tanggal, status)`, `tokens(token)` + `tokens(jadwal_id)`, `hasil_ujians(jadwal_id)`, `devices(peserta_id)`, `jawaban_pesertas(soal_id)` — plus indeks jalur tersibuk (Bab 16).

**Cache:** TIDAK perlu Redis (satu service lagi yang bisa mati); cukup HTTP caching aset statis + in-memory cache Go untuk soal bank terkunci + PostgreSQL buffers; halaman ujian/jawaban/token/hasil selalu `no-store`.

**Uji beban wajib sebelum hari-H:** simulasi 300 login + muat + autosave + submit dalam 10 menit; latensi p95 autosave >1 detik → perbaiki dulu.

### 10.3 Jaringan buruk / offline

Perilaku saat ini (CBT-Go): autosave tanpa retry/catch/antrean (jawaban terakhir hilang diam-diam jika gagal), tanpa localStorage/IndexedDB, timer JS dari `Remaining` saat load, tidak ada enforcement deadline server, resume bekerja untuk jawaban tersimpan di server, sesi in-memory (restart = logout massal), putus jaringan murni tidak menambah `out_counter` tetapi siswa yang buka tab lain untuk cek Wi-Fi AKAN terhitung pelanggaran (risiko false positive).

**Usulan v2:** antrean offline di browser (tulis ke localStorage/IndexedDB dulu + kirim ulang otomatis + indikator 🟢 tersimpan / 🟡 menunggu / 🔴 offline); heartbeat 15–30 detik; deadline di server (otoritatif, auto-submit server + job penyapu tiap 1 menit); sesi tahan restart (pindah ke tabel sesi di PostgreSQL); pause resmi untuk putus massal; `out_counter` hanya bertambah saat terhubung + aktif; 1 pelanggaran gratis yang dikonfirmasi proktor untuk kasus jaringan.

### 10.4 Backup & restore

**Yang di-backup:** DB `pg_dump -Fc` + uploads + konfigurasi + binary + versi + log (opsional). **Jadwal:** sebelum ujian (setiap jadwal), segera setelah ujian selesai, harian via cron (rotasi 14 hari), mingguan off-server. Prinsip 3-2-1 disederhanakan: Salinan 1 di server yang sama (folder backup khusus, terpisah dari folder data DB); Salinan 2 di luar server (wajib: hard disk eksternal / komputer lain di sekolah / penyimpanan cloud sekolah). Penamaan: tanggal + waktu + versi aplikasi + nama DB (mis. `cbt_go_2026-10-20_0600_v2.0.0.dump`). Retensi: harian 7 hari terakhir, mingguan 4 minggu terakhir, per periode ujian (STS/SAS/USBK/UKK) simpan sampai nilai final/rapor selesai + masa sanggah berakhir lalu arsipkan/hapus sesuai retensi.

**RPO/RTO:** RPO <5 menit selama ujian (dengan WAL archiving jika mampu), 24 jam di luar ujian; RTO <30 menit di mesin yang sama, <4 jam jika mesin diganti. **Uji restore minimal 1x per semester + 1x sebelum ujian besar (USBK).** Backup yang belum pernah diuji restore dianggap **belum terbukti**.

**Restore:** hanya jika DB rusak/hilang/tercampur data salah dan tidak bisa diperbaiki cara lain; restore menimpa DB saat ini — data setelah backup akan hilang → atas persetujuan penanggung jawab ujian. Langkah: hentikan aplikasi → backup dulu kondisi saat ini (beri nama "backup sebelum restore") → siapkan DB tujuan → pulihkan dump → verifikasi isi (bandingkan jumlah data) → jalankan aplikasi + uji. **Belum ada fitur backup/restore dari dalam aplikasi** (tercatat di README CBT-Go) — sampai fitur itu ada, backup dari luar aplikasi oleh operator/admin server.

### 10.5 Keamanan server

Akun demo wajib diganti/dimatikan; CSRF wajib; cookie `Secure`; DSN `sslmode=disable` dapat diterima HANYA jika DB di localhost yang sama; sesi in-memory (perbaiki ke tabel sesi); belum ada rate limit/audit log; upload perlu validasi. Firewall default deny, izinkan hanya 80/443 + SSH; SSH kunci saja, nonaktifkan root langsung, fail2ban; HTTPS wajib di luar LAN; PostgreSQL bind 127.0.0.1, user khusus hak minimal; header HSTS, `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`.

### 10.6 Tiga skenario deploy

| Skenario | Keputusan |
|---|---|
| **A. Server lokal sekolah (LAN/semi-daring)** | **SKENARIO UTAMA, DISARANKAN.** Satu komputer/server di sekolah, siswa lewat LAN/WiFi sekolah, tidak bergantung internet, alamat akses lokal mis. `http://<IP-server>:9988` peserta dan `http://<IP-server>:9988/adm` admin, IP server statis. Kelebihan: ujian jalan walau internet putus (keunggulan terbesar untuk USBK/STS/SAS), latensi <5 ms, data di sekolah, biaya rendah (Mini PC Rp 3–7 jt / PC bekas Rp 0–2 jt + switch/AP Rp 0,5–2 jt + UPS Rp 0,5–1,5 jt; rutin listrik Rp 30–80 rb/bln), kontrol penuh. Kekurangan: akses terbatas di sekolah, tergantung listrik + perangkat sekolah (perlu UPS + penanggung jawab), tidak ada HTTPS otomatis, backup/update tanggung jawab sekolah, satu titik kegagalan (mitigasi: backup rutin + komputer cadangan/rencana pinjam lab lain) |
| **B. VPS (systemd + reverse proxy + HTTPS)** | **UNTUK AKSES LUAR / CADANGAN / TRYOUT JARAK JAUH.** VPS + domain/subdomain (mis. `cbt.sekolah.sch.id`), PostgreSQL 16 di VPS, service systemd, reverse proxy Nginx/Caddy HTTPS:443 → `localhost:9988`, aplikasi TIDAK diakses langsung dari internet lewat 9988, Let's Encrypt + perpanjangan otomatis, firewall (hanya 443 + 80 redirect terbuka, PostgreSQL hanya localhost). Biaya: VPS 4 vCPU/8 GB/100 GB ~Rp 150–400 rb/bln + domain Rp 100–200 rb/thn. Kekurangan: bergantung internet (**tidak disarankan sebagai satu-satunya tempat untuk ujian resmi**), biaya sewa, data di luar sekolah, permukaan serangan lebih luas, sesi hilang saat restart |
| **C. cPanel shared hosting** | **TIDAK MUNGKIN SECARA LANGSUNG, JANGAN DIPAKSAKAN.** Bukan kekurangan konfigurasi kecil — ketidakcocokan arsitektur: shared hosting tidak mengizinkan proses server berjalan terus di port 9988, banyak paket hanya MySQL/MariaDB bukan PostgreSQL, tidak ada port kustom + service systemd + reverse proxy, akses shell/root/firewall terbatas. Yang BISA: C1 jangan jalankan CBT di cPanel — jalankan di server lokal (A) atau VPS (B), cPanel hanya untuk landing/pengumuman; C2 cPanel di VPS/dedicated — yang dipakai sebenarnya VPS-nya, ikuti B; C3 porting total ke PHP + MySQL — proyek besar (menulis ulang ke stack lain seperti CBT Laravel), jangan dipilih hanya demi memakai cPanel |

### 10.7 Operasional hari-H

Peran: penanggung jawab CBT, operator/proktor 1 per 1–2 ruangan + 1 cadangan, teknisi/admin IT, guru mapel. Checklist H-7/H-1/Hari-H/selama/setelah ujian (detail di Bab 19 — checklist "Siap Ujian" QA). Prinsip: **tidak ada update/restart/perubahan konfigurasi setelah titik H-1/sesi dimulai** sampai ujian selesai, kecuali darurat yang disetujui penanggung jawab. **Tidak ada yang me-restart server / meng-update aplikasi / me-restore backup saat ujian berjalan.**

---

## 11. Roles & Permission — Matriks

### 11.1 Definisi role

6 role: **Admin (EVIL), Proktor (OPRA), Guru (EVAL), Pengawas (BARU), Siswa (Peserta), Penguji UKK (BARU)**. Migrasi CBT-Go → v2: EVIL→Admin, EVAL→Guru, OPRA→Proktor, peserta→Siswa; Pengawas & Penguji UKK baru. Perlu tabel permission + penugasan + kelas/rombel + audit log + CSRF.

**Pertanyaan terbuka:** DB Dev mempertanyakan apakah PROKTOR dan PENGAWAS harus jadi role terpisah (jadi 7 role)? Laporan lain memakai 6 role. Lihat Bab 24.

### 11.2 Matriks permission (ringkas — ✅ boleh, ⚠️ terbatas/bersyarat, ❌ tidak boleh)

| Area | Admin | Proktor | Guru | Pengawas | Siswa | Penguji UKK |
|---|---|---|---|---|---|---|
| Kelola user & role | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Master data (jurusan/kelas/mapel) | ✅ | ⚠️ lihat | ⚠️ lihat | ❌ | ❌ | ❌ |
| Bank soal (buat/edit) | ✅ | ❌ (tidak boleh lihat isi soal) | ✅ mapelnya | ❌ | ❌ | ❌ |
| Kunci bank soal | ✅ | ❌ | ⚠️ banknya | ❌ | ❌ | ❌ |
| Paket ujian | ✅ | ❌ | ✅ mapelnya | ❌ | ❌ | ❌ |
| Jadwal & token | ✅ | ✅ jadwal yang ditugaskan | ⚠️ lihat/buat jika diizinkan | ❌ | ❌ | ❌ |
| Pelaksanaan hari-H (pause/resume/force finish/unblock/reset login) | ✅ | ✅ | ❌ | ⚠️ usulkan unblock, catat insiden | ❌ | ❌ |
| Monitor real-time | ✅ | ✅ | ⚠️ lihat saja | ✅ ruangan/sesinya | ❌ | ❌ |
| Koreksi esai & nilai | ✅ | ❌ | ✅ mapelnya | ❌ | ❌ | ❌ (UKK saja) |
| Laporan & ekspor | ✅ | ⚠️ status + agregat tanpa nilai per siswa | ✅ mapelnya | ⚠️ daftar hadir + berita acara | ✅ nilainya sendiri | ⚠️ rekap UKK yang ditugaskan |
| UKK (skema, rubrik, nilai praktik) | ✅ | ❌ | ⚠️ teori | ❌ | ❌ | ✅ yang ditugaskan |
| Settings & sistem | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Audit log | ✅ lihat | ❌ | ❌ | ❌ | ❌ | ❌ |

### 11.3 Aturan cakupan (scope)

- Guru hanya mapel + kelas yang ditugaskan (penegakan di server, bukan hanya sembunyikan menu).
- Proktor hanya jadwal yang ditugaskan, tidak boleh lihat isi soal.
- Siswa hanya jadwal kelasnya + jendela waktu + token valid + belum diblokir/belum selesai; tidak boleh lihat jadwal/nilai siswa lain walau menebak URL/ID.
- Penguji UKK hanya skema + kompetensi + peserta yang ditugaskan, akun eksternal masa berlaku terbatas.
- Pengawas hanya ruangan/sesinya, read-only + catat insiden, **tidak boleh melihat token**.
- Jalur login terpisah staff `/adm` vs siswa `/`. Token bukan password (acak, tidak berurutan, terikat jadwal + jendela waktu).

### 11.4 Pemisahan Tugas (SoD) — sistem harus MENOLAK, bukan hanya menyembunyikan

| # | Aturan |
|---|---|
| SoD-1 | Pembuat/edit bank soal ≠ pengunci |
| SoD-2 | Pembuat soal ≠ pemegang token |
| SoD-3 | Pengoreksi esai ≠ pengubah nilai final |
| SoD-4 | Penguji UKK ≠ finalisator nilai UKK |
| SoD-5 | Pengawas mengajukan, proktor mengeksekusi |
| SoD-6 | Pengelola user ≠ pengelola audit log |
| SoD-7 | Pengaju draft jadwal ≠ finalisator |

**Pengecualian darurat:** diizinkan hanya dengan alasan tertulis + audit + flag `SOD_OVERRIDE` + notifikasi kepala sekolah/admin kedua (asumsi — perlu keputusan sekolah; tidak pernah untuk ubah nilai tanpa jejak / hapus audit log).

### 11.5 Prinsip keamanan permission

Deny-by-default; penegakan di server (403 + catat); anti-IDOR; satu sesi = satu role aktif (asumsi); 404 untuk resource yang tidak berhak (agar tidak membocorkan keberadaan); route guard di frontend DAN backend (frontend hanya sembunyikan menu, backend menolak aksi terlarang).

---

## 12. Regulasi yang Berlaku

### 12.1 USBK vs ANBK vs UKK (dan TKA)

| Aspek | USBK (Ujian Sekolah Berbasis Komputer) | ANBK (Asesmen Nasional Berbasis Komputer) | UKK (Uji Kompetensi Keahlian SMK) |
|---|---|---|---|
| Sifat / tujuan | Penilaian sumatif oleh satuan pendidikan; dasar kenaikan kelas & kelulusan. Moda CBT dari Ujian Sekolah / Penilaian Sumatif Akhir Jenjang | Pemotretan & pemetaan mutu sistem pendidikan, bukan ujian kelulusan. Instrumen: AKM (literasi & numerasi), Survei Karakter, Survei Lingkungan Belajar | Mengukur pencapaian kompetensi keahlian SMK setara kualifikasi jenjang 2/3 KKNI; syarat sertifikasi kompetensi |
| Peserta | Seluruh peserta didik kelas yang diujikan di sekolah itu, sesuai POS sekolah | Sebagian peserta didik kelas V, VIII, XI dipilih acak oleh Pemerintah sebagai sampel (per penjelasan resmi GTK). Kepmen 56/2026 menggabungkan TKA & AN dan menyebut peserta TKA juga merupakan peserta AN — peserta TKA adalah murid kelas akhir (kelas 6 SD, 9 SMP, 12 SMA/SMK, 13 SMK 4 tahun). Perbedaan definisi ini perlu dikonfirmasi ke dinas/pedoman tahun berjalan | Peserta didik SMK tahun terakhir yang telah menuntaskan materi yang diujikan dan memiliki rapor lengkap sampai semester I tahun terakhir (per POS UKK contoh) |
| Penyelenggara | Satuan pendidikan (kepala sekolah, panitia US, guru/pengawas yang ditetapkan kepala sekolah) | Kemendikdasmen melalui Pusmendik/Badan Kebijakan Pendidikan + pemda + satuan pendidikan pelaksana. Sekolah **tidak** menyelenggarakan ANBK sendiri di luar sistem nasional | Satuan pendidikan terakreditasi bersama mitra DUDIKA / LSP / asosiasi profesi; Dinas Pendidikan Provinsi memverifikasi & menetapkan SMK yang layak jadi Tempat Uji Kompetensi (TUK) |
| Nilai individu? | **Ya.** Nilai per mata pelajaran, dipakai untuk rapor / nilai sekolah / kelulusan | **Tidak.** Hasil AN tidak memuat skor/nilai individual; tidak menentukan kelulusan; agregat tingkat satuan pendidikan | **Ya.** Nilai 0–100 (UKK Mandiri) + predikat kompetensi; menentukan lulus/tidak UKK |
| Konsekuensi kelulusan | Berpengaruh | Tidak berpengaruh pada kelulusan individu | Lulus UKK (≥70) syarat sertifikat; pengumuman kelulusan UKK mendahului pengumuman kelulusan |

**TKA (Tes Kemampuan Akademik) — dibedakan dari AN:** TKA **memiliki hasil individu** berupa Sertifikat Hasil TKA (SHTKA) & Daftar Kolektif Hasil TKA (DKHTKA), dengan kategori capaian. Hasil TKA bermanfaat untuk perencanaan berbasis mutu sekaligus validator nilai rapor untuk seleksi jalur prestasi PT. Jadi: **AN = tanpa nilai individu; TKA = ada nilai individu; UKK/US = ada nilai individu.** CBT SMK v2 adalah sistem **sekolah** (US + UKK), bukan pengganti ANBK/TKA nasional.

### 12.2 Aturan POS yang relevan untuk CBT

**Pola nasional TKA/AN (Kepmen 56/2026) — acuan paling resmi untuk tata tertib CBT terstandar:**
- **Token:** proktor melakukan rilis token dan menyerahkan kode token kepada pengawas/peserta. Peserta login dengan kartu login (username/password) lalu memasukkan token. Mekanisme ini juga muncul di POS ANBK: proktor merilis token via CBTSync beberapa menit sebelum mulai; peserta memasukkan token di komputer masing-masing.
- **Jadwal:** proktor/teknisi/pengawas wajib hadir **45 menit** sebelum tes; proktor & pengawas masuk ruangan **30 menit** sebelum pelaksanaan; peserta memasuki ruangan **15 menit** sebelum tes dimulai.
- **Durasi:** **tidak ditemukan** durasi baku per mata uji TKA/AN dalam potongan pedoman yang dibuka. Durasi mengikuti jadwal/sesi & instrumen masing-masing — jangan mengunci satu durasi nasional di CBT; buat durasi per paket/jadwal yang dapat dikonfigurasi.
- **Pengawas:** peran terpisah — pengawas ruang (cek foto kartu login vs wajah, edarkan daftar hadir, beri peringatan/sanksi), proktor (jalankan aplikasi, pastikan login sesuai yang hadir, beri solusi kendala teknis, unggah hasil moda semi-daring per sesi), teknisi (sarana/prasarana), penyelia pengawas (memantau via konferensi video, mencatat kejadian di luar prosedur). Pengawas dilarang: merokok, mengobrol, membaca, memberi isyarat/petunjuk/bantuan jawaban, merekam/memfoto/menyebarkan soal, membawa perangkat komunikasi/kamera kecuali untuk tugas.
- **Tata tertib peserta:** bawa kartu peserta; duduk sesuai sesi & tempat; kumpulkan tas, buku, HP, perangkat komunikasi/optik, kamera, kalkulator; isi daftar hadir; tidak boleh diwakilkan; hanya boleh keluar ruangan seizin pengawas; dilarang membuat kegaduhan, bekerja sama/menyontek, memakai alat bantu/minta bantuan, memakai joki, merekam/memfoto/menyebarkan soal. Toleransi keterlambatan **maksimal 15 menit** setelah tes dimulai (dapat mengikuti tes setelah diizinkan pengawas ruang).
- **Berita acara:** proktor/pengawas membuat & mengunggah berita acara + daftar hadir; pakta integritas ditandatangani proktor & pengawas. Pelanggaran diklasifikasikan **ringan / sedang / berat** dengan penanganan berjenjang. Laporan pelanggaran memuat: identitas pelapor, pelaku, bentuk, tempat, waktu, bukti, saksi.
- **Kerahasiaan soal & akun:** proktor & pengawas wajib menjaga kerahasiaan akun dan soal; dilarang merekam/memfoto/menyebarkan soal.

**Pola Ujian Sekolah / USBK (POS tingkat satuan pendidikan — bervariasi antar sekolah, tidak ada satu POS nasional USBK yang ditemukan):**
- US tulis dapat bermoda CBT; US praktik didampingi penguji.
- Pengawas ditetapkan kepala sekolah; pengawas adalah guru yang mata pelajarannya **tidak sedang diujikan**; jumlah pengawas per ruang **bervariasi** (ada contoh 2, ada contoh 1) — CBT harus mendukung konfigurasi jumlah pengawas per ruang, bukan hardcode.
- Pengawas masuk ruang lebih awal (contoh 10–20 menit), memeriksa kesiapan ruang, meminta peserta menunjukkan kartu peserta & menempati tempat duduk sesuai nomor, memastikan peserta tidak membawa tas/buku/catatan/kalkulator kecuali yang diizinkan, membacakan tata tertib, meminta tanda tangan daftar hadir, memastikan semua peserta login.
- Setelah tanda mulai: pengawas memberitahukan password/token/passkey ujian, mempersilakan mulai, mengingatkan baca petunjuk & kerjakan soal yang mudah dulu.
- Selama ujian pengawas menjaga ketertiban, memberi peringatan/sanksi kepada yang curang, melarang orang lain masuk ruang; dilarang memberi isyarat/petunjuk/bantuan jawaban.
- Peringatan "waktu tinggal 5 menit" sebelum selesai; setelah selesai: peserta berhenti, submit, pengawas memastikan tidak ada foto/screenshot soal.
- Peserta yang terlambat hanya boleh ikut setelah izin panitia/penguji **tanpa perpanjangan waktu**.
- Gangguan koneksi: peserta lapor ke proktor/pengawas; bila waktu habis dan tetap belum terhubung, peserta mengikuti ujian susulan sesuai jadwal.

**Pola UKK:** dilaksanakan di TUK yang terverifikasi; peserta mengikuti semua kompetensi yang diujikan; peserta berhalangan dengan bukti sah dapat ikut UKK susulan; peserta tidak lulus dapat ikut UKK tahun berikutnya sesuai POS. Nilai diserahkan penguji ke Panitia UKK dan **dijaga kerahasiaannya**.

### 12.3 Ketentuan nilai: KKTP (Kurikulum Merdeka) — bukan KKM

Kerangka: **Permendikbudristek 21/2022** (Standar Penilaian Pendidikan) — penilaian berkeadilan, objektif, edukatif; prosedur: perumusan tujuan → pemilihan/pengembangan instrumen → pelaksanaan → pengolahan → pelaporan; bentuk **formatif** (memantau & memperbaiki pembelajaran) dan **sumatif** (dasar penentuan kenaikan kelas & kelulusan); sumatif membandingkan pencapaian peserta didik dengan **kriteria ketercapaian tujuan pembelajaran**; kenaikan kelas mempertimbangkan laporan kemajuan belajar; kelulusan mempertimbangkan laporan kemajuan belajar pada kelas akhir jenjang.

**KKM (K13) vs KKTP (Kurikulum Merdeka):**
- **KKM (K13):** angka ambang batas per mata pelajaran / KD, ditetapkan sekolah melalui musyawarah (intake, kompleksitas, daya dukung); contoh praktik lama: tuntas bila ≥60 (sebagian referensi). Angka ini **bukan** angka nasional tunggal.
- **KKTP (Kurikulum Merdeka):** **menggantikan KKM**. KKTP adalah deskripsi kemampuan yang perlu ditunjukkan peserta didik sebagai bukti mencapai Tujuan Pembelajaran (TP). Dikembangkan guru saat merencanakan asesmen, dapat dinyatakan sebagai **deskripsi, rubrik, atau skala/interval nilai** — bukan angka mutlak tunggal. Panduan yang dirujuk melarang (tidak menyarankan) penggunaan angka mutlak seperti 75/80 sebagai kriteria; yang disarankan deskripsi, bila perlu interval (mis. 70–85, 85–100).
- Contoh interval KKTP yang banyak dipakai sekolah (bukan angka resmi nasional): 0–40% belum mencapai (remedial seluruh bagian); 41–65% belum mencapai ketuntasan (remedial bagian yang diperlukan); 66–85% sudah mencapai ketuntasan (tidak perlu remedial); 86–100% sudah mencapai ketuntasan (perlu pengayaan/tantangan lebih).
- **Implikasi:** CBT v2 **tidak boleh** meng-hardcode KKM=75 atau ambang tunggal. Setiap paket/mapel/TP harus punya KKTP yang dapat dikonfigurasi (deskripsi/rubrik/interval) + status tuntas/belum tuntas per TP.

**Remedi/pengayaan:** peserta yang belum mencapai kriteria mendapat **remedial** (perbaikan pembelajaran pada TP yang belum tercapai); yang sudah melampaui mendapat **pengayaan**. Untuk UKK: penguji dapat memberi kesempatan **pengulangan/perbaikan** bagi komponen yang belum mencapai standar **sampai batas tanggal ujian terakhir**; penentuan skor mempertimbangkan jumlah pengulangan, pemenuhan standar waktu, dan aspek sikap. **Tidak ditemukan** rumus nasional baku untuk nilai remedial (menggantikan / tidak menggantikan nilai awal, jumlah maksimal remedial) — ditetapkan sekolah dalam POS/kebijakan penilaian. CBT harus mencatat riwayat remedial dan membiarkan sekolah menentukan aturan penggabungannya.

**Skala nilai UKK (tegas, dari POS UKK):** UKK Mandiri **0–100**. Predikat: **<70 Belum Kompeten; 70–79 Cukup Kompeten; 80–90 Kompeten; 91–100 Sangat Kompeten.** Skema LSP/LSK/PTUK yang sertifikasinya tidak memunculkan skor: penguji **wajib mengonversi** capaian ke 0–100. **Lulus UKK: nilai ≥70.**

### 12.4 Ketentuan UKK SMK (ringkas)

- UKK menguji **pengetahuan, keterampilan, dan sikap dalam satu event penilaian**, didasarkan pada unjuk kerja/kinerja/produk.
- Perangkat UKK Mandiri: (1) **Instrumen Soal UKK (SPK):** tes penugasan menghasilkan produk/jasa; instrumen pemerintah menguji **keterampilan & sikap**, dilengkapi instrumen **pengetahuan** (PG, uraian, jawaban singkat, dan/atau wawancara); (2) **Lembar Pedoman Penilaian (PPsp):** lembar penilaian + rubrik (pengetahuan, keterampilan, sikap; komponen, sub-komponen, rekapitulasi); (3) **Instrumen Verifikasi Penyelenggara (InV):** kelayakan TUK (peralatan, tempat/ruang, persyaratan penguji).
- Penguji **mengupayakan teknik penilaian portofolio** sebelum teknik lainnya. Penguji menilai dengan lembar penilaian yang disediakan; dapat menambah indikator/komponen lebih tinggi dari yang ditetapkan.
- **Penguji UKK Mandiri:** gabungan internal + eksternal. Internal: guru mapel muatan produktif yang relevan, memenuhi persyaratan InV. Eksternal: SDM dari DUDIKA atau asosiasi profesi, berlatar belakang pendidikan dan/atau memiliki **sertifikat kompetensi + pengalaman kerja relevan**. Syarat DUDIKA mitra: telah bekerja sama dengan SMK **minimal 1 tahun** dan berkontribusi pada pengembangan sekolah.
- **6 skema UKK:** (1) sistem sertifikasi mitra DUDIKA/asosiasi profesi; (2) LSP-P1; (3) LSP-P2; (4) LSP-P3; (5) PTUK (BNSP); (6) **UKK Mandiri** (SMK terakreditasi, instrumen pemerintah sebagai standar minimal, melibatkan mitra DUDIKA, berorientasi SKL). SMK dapat memilih satu atau beberapa skema.
- **Sertifikat:** hanya bagi yang dinyatakan kompeten (lulus UKK). Bentuk: (a) berlogo Garuda Pancasila, diterbitkan **BNSP**; (b) berlogo mitra DUDIKA / asosiasi profesi; (c) berlogo **Tut Wuri Handayani**. Isi minimal: **identitas peserta, nama kompetensi keahlian, daftar kompetensi/unit yang telah diujikan dan dinyatakan kompeten**. Sertifikat UKK Mandiri diterbitkan DUDIKA yang terlibat, ditandatangani penguji/penguji eksternal atau perwakilan DUDIKA; satuan pendidikan menerbitkan sertifikat uji kompetensi yang ditandatangani perwakilan satuan pendidikan **bersama** perwakilan DUDIKA. Setiap sertifikat harus memenuhi **kaidah mampu telusur** (traceability). Hasil UKK dilaporkan satuan pendidikan beserta daftar nilainya ke Dinas Pendidikan Provinsi, Direktorat SMK, dan/atau tim yang ditunjuk, dan/atau e-Rapor.
- **Status Permendikbud 34/2018:** satu sumber sekunder menyebutnya dicabut oleh Permendikdasmen 21/2025, tetapi **belum diverifikasi** ke JDIH dalam riset ini. Acuan yang dipakai: Pedoman Penyelenggaraan UKK + POS UKK sekolah.

### 12.5 Data & privasi (UU No. 27/2022 tentang Pelindungan Data Pribadi)

- **Data pribadi** = data tentang orang perseorangan yang teridentifikasi / dapat diidentifikasi sendiri atau dikombinasi dengan informasi lain, baik langsung maupun tidak langsung, melalui sistem elektronik maupun nonelektronik.
- **Pengendali Data Pribadi** = pihak yang menentukan tujuan dan melakukan kendali pemrosesan. **Sekolah sebagai pengendali data** — bukan sekadar pengguna aplikasi — wajib memastikan tata kelola data patuh, aman, dapat dipertanggungjawabkan.
- **Data anak termasuk data pribadi spesifik** yang butuh perlindungan lebih ketat. **Pemrosesan data pribadi anak diselenggarakan secara khusus dan wajib mendapat persetujuan orang tua/wali anak** (Pasal 25 UU PDP, per transkrip UU).
- Pengendali wajib menyampaikan informasi kepada subjek data: legalitas pemrosesan, tujuan, jenis & relevansi data, jangka waktu retensi dokumen, rincian informasi yang dikumpulkan, jangka waktu pemrosesan, dan hak subjek data; bila ada perubahan, wajib memberitahukan sebelum perubahan terjadi (Pasal 21, per transkrip).
- Persetujuan pemrosesan melalui **persetujuan tertulis atau terekam** (dapat elektronik/nonelektronik); pengendali wajib dapat menunjukkan bukti persetujuan (Pasal 22 & 24, per transkrip).
- Prinsip: melindungi keamanan data dari pengaksesan/pengungkapan/pengubahan/perusakan/penghilangan yang tidak sah; memberitahukan tujuan & aktivitas pemrosesan serta kegagalan pelindungan; **data dimusnahkan/dihapus setelah masa retensi berakhir atau atas permintaan subjek data**, kecuali ditentukan lain oleh peraturan; pemrosesan dilakukan bertanggung jawab & dapat dibuktikan.
- **Dilarang** secara melawan hukum memperoleh/mengumpulkan, mengungkapkan, atau menggunakan data pribadi yang bukan miliknya (Pasal 65). Sanksi pidana & administratif diatur UU PDP — rincian besaran sanksi (disebut di sumber sekunder: pidana penjara & denda miliaran rupiah, denda administratif hingga 2% pendapatan tahunan) **belum diverifikasi** ke teks resmi UU dalam riset ini, jangan dikutip sebagai angka pasti tanpa cek JDIH/BPK. Kewajiban pelaporan kebocoran dalam **3×24 jam** — hanya ditemukan di sumber sekunder, belum diverifikasi ke teks resmi.

**Siapa boleh akses nilai (dari regulasi ujian):**
- **Nilai UKK:** penguji menyerahkan nilai kepada **Panitia UKK** dan menjaga kerahasiaannya. Nilai dilaporkan ke Dinas Provinsi / Direktorat SMK / tim yang ditunjuk, dan/atau e-Rapor. Akses terbatas: panitia, penguji, pimpinan sekolah, dinas — **bukan publik, bukan semua guru**.
- **Nilai AN:** tidak ada nilai individu; hasil agregat untuk perbaikan pembelajaran.
- **Nilai TKA:** disampaikan melalui DKHTKA via dinas/kantor Kemenag ke satuan pendidikan; peserta mendapat SHTKA; salinan SHTKA ditatausahakan satuan pendidikan.
- **Data peserta TKA/AN:** bersumber dari **Dapodik/EMIS**, diverifikasi-validasi berbasis **NISN**; pendaftaran memakai Surat Pernyataan Keikutsertaan yang ditandatangani orang tua/wali dan disimpan di satuan pendidikan.
- **Tidak ditemukan:** ketentuan retensi spesifik (berapa tahun menyimpan data/nilai ujian di CBT sekolah); ketentuan siapa persis yang boleh melihat nilai US harian di dalam sekolah (di luar panitia/penguji/guru mapel/wali kelas) — diatur kebijakan sekolah & UU PDP. **Jangan mengarang masa retensi.** CBT harus menyediakan pengaturan retensi & penghapusan/anonimisasi oleh admin sekolah, plus log akses.

### 12.6 Implikasi regulasi ke fitur (ringkas — pemetaan aturan → fitur)

Token per sesi/jadwal (generate, rilis, kedaluwarsa, reset login, log rilis); jadwal + gelombang + sesi + ruang + pemetaan tempat duduk; timer server-side per peserta + auto-submit + aturan keterlambatan & susulan yang dapat dikonfigurasi + log keterlambatan; monitor real-time + deteksi pindah tab + tombol catat pelanggaran (ringan/sedang/berat + bukti + saksi) + auto-block sesuai aturan sekolah + BAPU; generator BAU per sesi/ruang + daftar hadir digital + cetak + unggah/simpan pakta integritas; kontrol akses bank soal berbasis peran + soal tidak dapat diunduh/disalin peserta + watermark/nonaktifkan copy + log akses soal + hashing password + tidak ada endpoint publik yang mengekspos soal/kunci; pemisahan jenis penilaian (formatif/sumatif/US/UKK/tryout) + rekap sumatif + status kenaikan/kelulusan diputuskan rapat, bukan otomatis; master KKTP per mapel/TP (deskripsi, rubrik, interval) + tuntas/belum tuntas **per TP** + laporan ketercapaian per TP; alur remedial (tandai TP belum tuntas → jadwalkan → catat nilai remedial + riwayat → aturan gabungan dapat dikonfigurasi); modul UKK terpisah (soal teori, lembar observasi praktik, unggah bukti/portofolio, input nilai per komponen oleh penguji, rekap otomatis → nilai akhir 0–100); master penguji (internal/eksternal, asal DUDIKA, sertifikat, pengalaman, tanggal mulai kerja sama; validasi DUDIKA <1 tahun diberi peringatan); field skema UKK + checklist verifikasi TUK; kalkulasi nilai UKK otomatis dengan ambang & predikat yang dapat dikonfigurasi (default 70 & rentang Bab 12.3) + konversi capaian → skor 0–100 untuk skema LSP/PTUK; generator data sertifikat (bukan penerbitan BNSP/LSP itu sendiri) + nomor sertifikat + penandatangan + jejak audit; ekspor rekap nilai UKK (CSV/Excel) untuk dinas & e-Rapor + hak akses nilai dibatasi peran; persetujuan orang tua/wali terdokumentasi + halaman privasi (tujuan, jenis data, retensi) + log persetujuan + mekanisme permintaan penghapusan data; keamanan (hashing, RBAC, audit log, backup terenkripsi, retensi & penghapusan/anonimisasi terjadwal, prosedur laporan insiden); impor/sinkron data peserta dari Dapodik (NISN sebagai kunci) + validasi NISN duplikat/tidak valid + jejak perubahan data peserta; fitur operasional (reset login, pindah perangkat/sesi cadangan, penandaan "terkendala teknis", penjadwalan susulan, log insiden per sesi).

**Yang sengaja TIDAK disimpulkan (tidak ditemukan sumbernya):** durasi baku nasional per mata pelajaran USBK/US-CBT; POS USBK nasional yang masih berlaku; masa retensi spesifik data/nilai ujian; rumus nasional nilai remedial; besaran sanksi UU PDP & tenggat 3×24 jam (sumber sekunder, belum diverifikasi); status terbaru Permendikbud 34/2018 (belum diverifikasi ke JDIH); ketentuan TKA/AN tahun berjalan (jadwal, durasi, daftar mata uji — cek edaran terbaru sebelum mengunci jadwal).

**Rekomendasi tindak lanjut (bukan klaim):** sebelum mengunci spesifikasi, minta/cek ke sekolah: (a) POS US & POS UKK sekolah tahun berjalan, (b) edaran TKA/AN terbaru dari dinas, (c) teks resmi Permendikbudristek 21/2022 & UU 27/2022 dari JDIH untuk pasal-per-pasal, (d) kebijakan retensi data & persetujuan orang tua yang dipakai sekolah.

---

## 13. Benchmark: Exo vs CBT-Go vs CBT Laravel

**Catatan kualitas bukti (penting):**
- **Extraordinary:** fitur disimpulkan dari daftar menu/permission (100+) dan skema dump 38 tabel, **bukan** dari uji live — instalasi kajian terhenti di gerbang validasi lisensi (signature offline tidak cocok, cek online timeout, percobaan dihentikan karena risiko IP-block). Jadi status ✅ untuk Exo artinya "terdapat di menu/skema menurut kajian", bukan "terverifikasi jalan".
- **CBT-Go:** fitur diverifikasi dari kode sumber (handler, store, template).
- **CBT SMK (Laravel):** README-nya adalah README bawaan Laravel (tidak menjelaskan aplikasi), jadi fitur ditetapkan dari routes, controller, migrasi, dan view — bukan dari README.

Legenda: ✅ ada · ⚠️ sebagian / nominal / belum ditegakkan · ❌ tidak ada.

| # | Fitur | Extraordinary CBT v4.7.0 | CBT-Go (Go + PostgreSQL) | CBT SMK (Laravel) |
|---|---|---|---|---|
| 1 | Bank soal | ✅ CRUD + duplikat soal + kunci/lock bank + backup & restore bank + lampiran | ✅ CRUD bank + lock/unlock + import CSV soal. ❌ Tanpa duplikat, tanpa edit soal (hanya tambah + hapus), tanpa lampiran, tanpa backup | ✅ CRUD soal penuh (tambah + edit + hapus) + filter. Soal menempel ke mapel+guru (tanpa entitas bank). ❌ Tanpa lock, tanpa duplikat, tanpa import soal, tanpa backup |
| 2 | Tipe soal | ✅ 16 kode (`code01`–`code16`): PG, PG kompleks, benar-salah, menjodohkan, isian singkat, esai, mengurutkan, setuju-tidak, dll (klaim situs) + stimulus teks/gambar/audio, layout soal, `case_sensitive`, `absolute_eval` | ⚠️ 16 label ada di `model.go`, tapi UI ujian hanya punya 3 jalur render: radio (tipe 1,3,8,9,10,11), checkbox (2,14), textarea untuk sisanya — menjodohkan, mengurutkan, isian panjang, studi kasus dll. semuanya jadi textarea bebas. Tipe stimulus 9–11 tampil sebagai radio biasa, tanpa gambar/tabel/grafik. **Hanya 8 tipe yang auto-grade** | ✅ 5 tipe, semuanya terimplementasi sungguhan: PG, PG Kompleks (cocok himpunan persis), Menjodohkan (pasangan terstruktur + skor proporsional), Isian Singkat (case-insensitive), Uraian (rubrik di kunci) + kolom stimulus + level mudah/sedang/sulit. Lebih sedikit tipe, tapi tiap tipe nyata |
| 3 | Paket ujian | ⚠️ Tanpa entitas "paket" terpisah; jadwal langsung menempel ke bank soal. Jenis STS/SAS/USBK tidak disebut di kajian | ⚠️ Tanpa entitas paket; bank soal + jadwal saja. Tanpa jenis ujian, tanpa acak soal/jawaban | ✅ Model paket paling kaya: jenis STS/SAS/USBK/UKK Teori/Tryout, durasi, jumlah soal, token, `acak_soal`, `acak_jawaban`, status draft/aktif/nonaktif, kelola soal manual + ambil acak |
| 4 | Jadwal | ✅ Tambah/edit/hapus, **pause & resume**, 8 sesi, reset ujian, force finish, reset login peserta | ✅ Buat + pause/resume/finish/hapus, nomor sesi, tanggal/jam/durasi. ⚠️ **Pause hanya gerbang masuk** (ujian yang di-pause menampilkan halaman "di-pause"), **timer tidak dibekukan** — sisa waktu dihitung dari `mulai_ujian`, jadi pause tidak menambah waktu. ❌ Tanpa reset per peserta / force finish / reset login | ⚠️ "Sesi ujian" (paket + kelas + tanggal + jam mulai/selesai + status dijadwalkan/berlangsung/selesai). ❌ Tanpa pause/resume, tanpa force finish, tanpa reset |
| 5 | Token | ✅ Generate, revoke, expiry (default 900 detik) | ✅ Generate 6 karakter + revoke + expiry **dicek** saat validasi (`tok.Status`, `JadwalID`, `ExpiredAt`). ⚠️ **Expiry = akhir hari jadwal (23:59:59), bukan 900 detik; setting `TOKEN_EXPIRED=900` tersimpan di DB tapi tidak dirujuk kode yang dibaca** | ⚠️ Satu token per paket (6 karakter, bisa regenerate), divalidasi case-insensitive. ❌ Tanpa expiry, tanpa revoke, tanpa token per sesi/jadwal |
| 6 | Ujian peserta | ✅ Timer, tanda ragu-ragu, `LOGIN_ONCE_ATIME`, fullscreen enforcement | ✅ Semua soal satu halaman, timer countdown + auto-submit, autosave per soal (fetch ke `/exam/save`), checkbox ragu-ragu, halaman blocked. ❌ Tanpa acak soal/jawaban, tanpa navigasi satu-per-satu / panel tinjau soal | ✅ Navigasi satu soal per halaman, urutan soal diacak stabil per siswa (seed `siswa_id`), opsi jawaban diacak per siswa, timer + auto-submit saat habis, autosave per soal, halaman hasil. ❌ Tanpa ragu-ragu, tanpa panel tinjau |
| 7 | Koreksi | ✅ Otomatis (objektif) + manual uraian + **opini AI** di koreksi | ✅ Otomatis untuk 8 tipe + manual untuk jawaban tertunda; nilai akhir = auto + manual. ⚠️ Menjodohkan/mengurutkan tidak auto-grade (masuk antrean manual, tanpa skor proporsional) | ✅ Otomatis via `Soal::koreksi` (PG, PG kompleks, isian, menjodohkan proporsional) + koreksi manual uraian dengan hitung ulang nilai akhir. Hanya uraian yang masuk antrean koreksi |
| 8 | Anti-curang | ✅ Device lock (tabel `devices`, 1 perangkat/peserta), `AUTO_BLOCK`, `MAX_COUNTER_OUT=1`, `out_ujian_counter`, alasan blokir, `antiblock`, fullscreen enforcement, log keluar ujian | ✅ `visibilitychange` + `blur` → `out_counter` +1; >1 = auto-block (status peserta 0 + status ujian 3) + unblock oleh proktor; perangkat dicatat (cookie `cbt_device` + IP). ⚠️ Perangkat hanya **dicatat, tidak ditegakkan** — perangkat kedua tidak ditolak. ❌ Tanpa fullscreen enforcement, tanpa login-once | ⚠️ Hanya **pencatatan**: `visibilitychange`/`blur` → `log_aktivitas` (`pindah_tab`, `jendela_tidak_fokus`). ❌ Tanpa counter, tanpa blokir, tanpa rekam perangkat, tanpa fullscreen |
| 9 | Laporan | ✅ Ledger nilai per peserta, hasil ujian, berita acara (dengan kop sekolah), export Excel, kartu peserta | ⚠️ Rekap per jadwal + export CSV. ❌ Tanpa berita acara, tanpa rekap per kelas/paket, tanpa kartu peserta, tanpa Excel | ✅ Rekap per paket/kelas + rata-rata + ambang lulus 70 + export CSV + berita acara print-friendly (dilengkapi log aktivitas). ❌ Tanpa kartu peserta, tanpa Excel |
| 10 | Analisis butir soal | ✅ Paling canggih: `banksoal_summaries` (rata-rata, SD, durasi min/max, nilai min/max) + `soal_summaries` (**daya beda**, **analisis pengecoh**, total benar/salah/kosong) | ❌ Tidak ada di kode. Yang ada hanya **DRAF spesifikasi** (`ANALISIS-BUTIR-SPESIFIKASI.md`): rumus kesukaran P, daya beda D metode 27% atas-bawah, pengecoh PG, query SQL, peringatan N kecil — belum dibangun | ❌ Tidak ada |
| 11 | Nilai per kompetensi | ✅ Tiap soal terikat `competence_item` (bagian, bobot, poin salah); hasil per kode kompetensi (`code01_...` s.d. 16) | ❌ Tidak ada di penilaian. Kolom `soals.competence_code` **sudah ada** dan diisi saat tambah soal, tapi tidak dipakai menghitung hasil. Draf spesifikasinya ada (rumus skor per kompetensi), kodenya belum | ❌ Tidak ada (tidak ada kolom kompetensi sama sekali) |
| 12 | UKK | ❌ Tidak ada modul khusus | ❌ Tidak ada | ✅ **Satu-satunya yang punya**: tabel + CRUD (siswa, judul, nilai teori, nilai praktik, penguji, status dijadwalkan/lulus/tidak_lulus/belum_dinilai, keterangan) + role `penguji_ukk`. ⚠️ Tanpa upload file praktik/sertifikat, tanpa rubrik praktik detail, terpisah dari alur paket |
| 13 | Event / ruangan / kartu peserta | ✅ Event ujian (gelombang), sesi peserta, ruangan + `room_members`, kartu peserta, copy/upload sesi peserta, upload anggota ruangan | ❌ Tidak ada | ❌ Tidak ada (hanya sesi per kelas) |
| 14 | Backup / restore | ✅ Backup & restore bank soal + backup & restore master system | ❌ Tidak ada | ❌ Tidak ada |
| 15 | Import | ✅ Google Docs / Word (.docx) termasuk rumus + import bank soal + upload CSV peserta | ⚠️ CSV peserta + CSV soal (format: tipe, pertanyaan, bobot, opsi `"teks\|1;..."`). ❌ Tanpa Word/Google Docs, tanpa .docx | ⚠️ Import CSV **siswa** saja (NISN, nama, kelas). ❌ Tanpa import soal, tanpa Word/Docs |
| 16 | AI | ✅ Opsional: OpenAI (gpt-4o-mini) + Gemini untuk generate soal & opini koreksi (butuh API key sendiri) | ❌ Tidak ada | ❌ Tidak ada |

**Konteks arsitektur & peran (di luar tabel):**

| Aspek | Extraordinary | CBT-Go | CBT SMK |
|---|---|---|---|
| Stack | Go binary tunggal ~48 MB + PostgreSQL 16, port 9988, `/adm` | Go 1.22 single binary + PostgreSQL 16, port 9988, `/adm`, template embed, ~1.905 baris Go | Laravel 13 + SQLite (file `database/database.sqlite` ada di repo), Blade, Vite |
| Role | EVIL / EVAL / OPRA + **100+ permission per aksi** (`role_menus`, mis. `0010H` = Tambah Soal) | EVIL / EVAL / OPRA, cek role kasar per handler | 6 role: admin, proktor, guru, pengawas, siswa, penguji_ukk; `RoleMiddleware` cek role per grup route |
| Monitor | ✅ Real-time (WebSocket aktif) + progress peserta | ⚠️ Polling refresh 10 detik (bukan WebSocket), status + out_counter saja | ❌ Tanpa halaman monitor (hanya hitungan di dashboard + log di berita acara) |
| Keamanan form | Tidak terverifikasi (binary tertutup) | ❌ **Tanpa CSRF** di semua form; hapus jurusan/matpel via **GET `?del=`** (terdokumentasi di `SECURITY-FIX-GET-DELETE.md` sebagai DRAF belum diterapkan); password demo di seed; tanpa wajib ganti password | ✅ CSRF bawaan Laravel (`X-CSRF-TOKEN` terlihat di view ujian); session regenerate saat login |
| Lisensi / keterbukaan | ❌ Binary tertutup + lisensi per server via ecosystem; password default bawaan paket (wajib diganti) | ✅ Full source, gratis | ✅ Full source, gratis |

### 13.1 Kelebihan & kekurangan masing-masing

**Extraordinary CBT v4.7.0**
- Kelebihan: set fitur paling lengkap dan paling matang secara konseptual (analisis butir, kompetensi, backup/restore, event/ruangan/kartu, import .docx, AI); skema 38 tabel yang koheren; kontrol operasional proktor paling lengkap (pause/resume, force finish, reset ujian, reset login, token expiry, device lock, auto-block); arsitektur deployment sederhana (single binary + PostgreSQL) yang terbukti bisa ditiru.
- Kekurangan: binary tertutup + gerbang lisensi (tidak bisa diperbaiki/dikustomisasi sendiri; instalasi kajian terhenti di validasi lisensi, risiko IP-block; sekolah bergantung server lisensi vendor untuk aplikasi LAN); **tidak ada modul UKK** — kekurangan fatal untuk SMK; granularitas permission 100+ per aksi adalah beban administrasi yang tidak proporsional untuk satu sekolah; password default bawaan paket yang dipublikasikan di panduan adalah jebakan keamanan; fitur belum terverifikasi jalan secara live (hanya dari menu/skema); AI bergantung API key + biaya eksternal, opini AI butuh pengawasan guru.

**CBT-Go**
- Kelebihan: arsitektur deployment terbaik (single binary, template + migrasi ter-embed, tanpa build step frontend, PostgreSQL); kode kecil dan mudah dibaca (~1.905 baris Go), mudah diaudit; sudah mengadopsi hal operasional terpenting dari Exo (pause/resume, token expiry + revoke, auto-block, ragu-ragu, lock bank, monitor, koreksi manual gabung ke nilai akhir); fondasi masa depan sudah ditanam (kolom `competence_code`, draf spesifikasi analisis butir yang matang, draf perbaikan keamanan); skema bersih dengan UUID, constraint, dan index yang masuk akal.
- Kekurangan: **keamanan belum layak deploy** (tanpa CSRF, hapus via GET, akun demo password lemah di seed, tanpa wajib ganti password, cookie sesi belum `Secure` — semuanya DRAF belum diterapkan); CRUD tidak lengkap (tanpa edit untuk master data, user, dan soal — hanya tambah + hapus); 16 tipe soal sebagian nominal (menjodohkan/mengurutkan/studi kasus dll. hanya textarea; tipe stimulus tanpa media; tanpa acak soal/jawaban); tanpa UKK; analisis butir + nilai per kompetensi masih dokumen, bukan kode; **pause tidak membekukan timer**; perangkat hanya dicatat bukan dikunci; monitor polling bukan real-time; tanpa reset per peserta/force finish; tanpa berita acara; tanpa backup/restore; tanpa event/ruangan/kartu; satu halaman berisi semua soal (berisiko berat untuk bank soal besar dan memudahkan intip jawaban antar soal).

**CBT SMK (Laravel)**
- Kelebihan: CRUD paling lengkap (edit tersedia untuk jurusan/kelas/mapel, soal, paket, sesi, UKK — hanya akun guru/siswa yang tanpa edit); model paket paling kaya (jenis ujian, acak soal + acak jawaban, kelola soal manual/acak) dan 6 role termasuk pengawas + penguji UKK; satu-satunya dengan modul UKK (teori + praktik + penguji + status kelulusan); penilaian paling teliti untuk tipe yang ada (menjodohkan proporsional, PG kompleks himpunan persis, isian case-insensitive, rubrik uraian); laporan administratif paling siap pakai (rekap + CSV + berita acara dengan log aktivitas); keamanan bawaan Laravel (CSRF, session regenerate, hashing password); pengacakan paling adil (urutan soal + opsi diacak stabil per siswa via seed `siswa_id`).
- Kekurangan: **anti-curang hanya teater** (pindah tab hanya dicatat di log, tanpa konsekuensi); token tanpa expiry/revoke dan hanya satu per paket (token bocor berlaku selamanya); tanpa pause/resume; hanya 5 tipe soal; tanpa analisis butir; tanpa nilai per kompetensi; tanpa backup/restore; tanpa event/ruangan/kartu peserta; tanpa halaman monitor real-time (proktor buta selama ujian); tanpa import soal; UKK tanpa upload file praktik/sertifikat; deployment lebih berat (PHP + Composer + Node/Vite) dan riwayatnya pernah gagal dipublikasikan via tunnel; README repo masih bawaan Laravel.

### 13.2 Fitur yang WAJIB diadopsi ke v2 (prioritas)

| Prioritas | Fitur (sumber) | Alasan |
|---|---|---|
| P0 — sebelum deploy apa pun | **Keamanan dasar: CSRF di semua form + hapus hanya via POST + wajib ganti password saat login pertama + hapus/nonaktifkan seed demo sebelum produksi** (pelajaran dari CBT-Go; pola dari SMK/Laravel) | CBT-Go mendokumentasikan sendiri lubang-lubangnya sebagai DRAF belum diterapkan. Satu link GET `?del=` yang diklik admin bisa menghapus jurusan/matpel |
| P0 | **Full CRUD termasuk edit** untuk master, user, soal, jadwal (pola SMK) | CBT-Go yang hanya tambah+hapus tidak bisa dipakai harian |
| P1 | **Pause/resume jadwal + force finish + reset ujian per peserta + reset login** (Exo via CBT-Go) — dengan perbaikan: **pause harus membekukan timer** | Gangguan listrik/jaringan di lab adalah kejadian rutin. Tanpa ini, satu gangguan = ujian gagal massal. Cacat timer CBT-Go wajib diperbaiki, bukan ditiru |
| P1 | **Auto-block berbasis counter + unblock oleh proktor + log aktivitas sebagai jejak audit** (Exo + CBT-Go + SMK) | Penegakan dari CBT-Go, jejak audit dari SMK, alasan blokir dari Exo. Log saja (SMK) tidak menegakkan; block tanpa log tidak bisa diaudit |
| P1 | **Token per jadwal dengan expiry + revoke** (Exo + CBT-Go) | Token abadi per paket (SMK) berarti token bocor di grup chat berlaku selamanya |
| P1 | **Tanda ragu-ragu** (Exo + CBT-Go) | Biaya kecil, nilai besar bagi siswa |
| P1 | **Kunci/lock bank soal** (Exo + CBT-Go) + **duplikat soal/bank** (Exo) | Lock mencegah soal diubah saat ujian berjalan. Duplikat mempercepat guru membangun varian bank |
| P1 | **Model paket SMK: jenis STS/SAS/USBK/UKK/Tryout + acak soal + acak jawaban + kelola soal manual/acak** | Keunggulan SMK yang hilang saat membangun ulang di CBT-Go. v2 harus punya paket sebagai entitas |
| P1 | **Modul UKK: teori + praktik + penguji + status kelulusan** (SMK) + **upload file praktik + sertifikat** (menutup celah SMK) | Tanpa UKK, v2 bukan CBT SMK. File praktik/sertifikat yang membuat modul ini lengkap secara administratif |
| P1 | **Berita acara + rekap + export CSV/Excel** (SMK + Exo) | Kebutuhan administrasi sekolah yang tidak bisa ditawar |
| P2 | **Analisis butir soal: kesukaran (P), daya beda (D, metode 27% atas-bawah), analisis pengecoh** (Exo; draf CBT-Go siap pakai) — lengkap dengan **peringatan N kecil** (N < 10 = belum dapat dipercaya) | Satu-satunya fitur yang membuat guru bisa membedakan "siswa bodoh" dari "soal jelek". Tanpa peringatan N kecil, angka D menyesatkan untuk kelas kecil |
| P2 | **Nilai per kompetensi** (Exo; kolom `competence_code` sudah ada di CBT-Go) | Nilai total tidak memberi tahu guru kompetensi mana yang gagal. Soal tanpa kode kompetensi tampil terpisah sebagai "Tanpa Kompetensi" |
| P2 | **Monitor peserta + aksi proktor dari monitor** (Exo + CBT-Go) | Proktor buta selama ujian di SMK. Minimal: status per peserta + out_counter + tombol unblock/reset/force-finish. Polling 10 detik (CBT-Go) cukup untuk v2; WebSocket bisa menyusul |
| P2 | **Import CSV peserta + CSV soal** (CBT-Go + SMK) | Tanpa import, pengisian bank soal 100+ soal dan peserta satu angkatan tidak praktis |
| P2 | **Perangkat dicatat DAN ditegakkan sebagai device lock** (Exo; CBT-Go baru mencatat) | "Dicatat tapi tidak ditolak" adalah janji palsu. Aturannya eksplisit: 1 perangkat per peserta per ujian, perangkat kedua ditolak + jalur unblock proktor |
| P3 | **Event + ruangan + kartu peserta** (Exo) | Diperlukan untuk USBK skala besar. Bisa fase kedua bila waktu mepet (sesi per kelas SMK cukup untuk ujian harian) |
| P3 | **Backup/restore dari dalam aplikasi** (Exo) | Pengaman operasional: sekolah tanpa staf IT butuh tombol backup, bukan perintah dump manual. Fase kedua, bukan penghambat peluncuran |

### 13.3 Fitur yang TIDAK perlu diadopsi

| Fitur (sumber) | Alasan menolak / menunda |
|---|---|
| Binary tertutup + lisensi per server + ketergantungan server lisensi vendor (Exo) | Bertentangan dengan tujuan v2: kode sendiri yang bisa diperbaiki sendiri. Aplikasi ujian LAN sekolah tidak boleh mati karena server lisensi pihak ketiga tidak terjangkau |
| 100+ permission per aksi (Exo) | Over-engineering untuk satu sekolah. 4–6 role dengan cek per menu sudah cukup dan jauh lebih mudah diaudit |
| AI generate soal / opini koreksi sebagai fitur inti (Exo) | Butuh API key + biaya berjalan + risiko akurasi. Nilai AI baru terasa setelah bank soal, koreksi manual, dan analisis butir sudah rapi. Bila kelak ditambah, opsional dan tidak pernah menggantikan koreksi guru tanpa review |
| Klaim "16 tipe soal" tanpa render & grading per tipe (celah CBT-Go) | Jangan mengadopsi angka pemasarannya. v2 hanya boleh mengklaim tipe soal yang benar-benar punya input terstruktur + aturan penilaian. Lebih baik 8 tipe yang nyata daripada 16 label dengan textarea |
| Hapus via GET / tanpa CSRF (celah CBT-Go) | Secara eksplisit dilarang di v2 — kerentanan yang memungkinkan penghapusan data lewat satu klik link |
| Token abadi satu-per-paket (SMK) | Digantikan token per jadwal + expiry + revoke (P1) |
| Log anti-curang tanpa konsekuensi (SMK) | Memberi ilusi pengawasan. v2 memakai counter + auto-block + unblock proktor (P1) |
| Fullscreen enforcement dengan auto-block keras sejak hari pertama (Exo) | Terlalu berisiko false positive di perangkat sekolah. Adopsi bertahap: peringatan + counter + keputusan proktor; penegakan keras hanya setelah uji coba |
| Import Word/Google Docs (.docx) dengan rumus sebagai syarat v1 (Exo) | Upaya besar, frekuensi pakai rendah. CSV dulu; .docx fase berikutnya bila benar-benar diminta guru. (Berkaitan dengan kontradiksi Bab 23 — perlu keputusan user) |
| Reliabilitas tes (Cronbach's alpha) di v1 | Draf analisis CBT-Go sendiri menyatakan ini bukan kebutuhan guru sehari-hari. Daya beda + pengecoh + kesukaran sudah cukup |
| Pelengkap Exo: tutorial audio in-app, shortcuts, personal access token | Nice-to-have produk komersial, bukan kebutuhan ujian sekolah |
| Password default / akun demo yang ikut ter-deploy (Exo + CBT-Go + SMK) | Seed demo hanya untuk lokal dan wajib nonaktif sebelum produksi |
| Dua basis kode yang dibiarkan hidup berdampingan (situasi saat ini: SMK vs Go) | Dua kodebase berarti dua tempat bug dan kebingungan "yang mana yang dipakai". v2 adalah SATU basis kode; yang satu lagi dibekukan/dipensiunkan setelah migrasi fitur (UKK, paket, berita acara) selesai |

### 13.4 Pelajaran dari tiap sistem — apa yang harus dihindari di v2

**Dari Extraordinary:** jangan membangun inti yang tidak bisa diperbaiki sendiri; jangan menggantungkan aplikasi LAN pada server eksternal; jangan mengirim password default yang diketahui publik; jangan membiarkan granularitas izin meledak; jangan mengutip fitur dari daftar menu sebagai "terbukti" — v2 mengklaim fitur hanya setelah uji end-to-end dengan bukti.

**Dari CBT-Go:** keamanan bukan "tahap berikutnya" (CSRF, hapus via POST, ganti password paksa harus ada di commit awal, bukan dokumen DRAF); jangan memberi label yang tidak ditepati implementasi ("16 tipe" yang sebagian besar berakhir sebagai textarea); **pause harus membekukan waktu** (simpan `sisa_waktu` saat pause, lanjutkan dari nilai itu saat resume); "dicatat" ≠ "ditegakkan"; jangan biarkan spesifikasi menjadi kuburan fitur (draf analisis butir + draf keamanan yang bagus tapi belum menjadi kode — di v2, spesifikasi yang disetujui harus langsung masuk daftar kerja dengan tes terima); seed demo tidak boleh mencapai produksi.

**Dari CBT SMK (Laravel):** pencatatan tanpa penegakan adalah teater; token tanpa kedaluwarsa adalah lubang; jangan membangun tanpa halaman monitor; jangan meninggalkan modul setengah jadi (UKK tanpa upload file praktik/sertifikat berhenti satu langkah sebelum berguna administratif — setiap modul v2 dinilai dari "apakah berkas administrasi sekolah sudah lengkap", bukan "apakah CRUD-nya jalan"); jangan membiarkan README berbohong (README repo SMK masih bawaan Laravel); jangan mengulang pembangunan dari nol sebelum memindahkan keunggulan lama (CBT-Go kehilangan paket, UKK, berita acara, dan pengacakan SMK karena dibangun ulang tanpa daftar "yang tidak boleh hilang" — v2 dimulai dari daftar migrasi fitur Bab 13.2, bukan dari kanvas kosong).

**Ringkasan benchmark (satu paragraf):** Extraordinary memberi **cetak biru fitur** (analisis butir, kompetensi, event/ruangan, backup, pause/auto-block) tetapi tidak boleh ditiru model tertutup/berlisensinya; CBT-Go memberi **arsitektur deployment** (single binary + PostgreSQL) dan sudah mengadopsi sebagian cetak biru itu, tetapi masih punya utang keamanan dan banyak fitur nominal; CBT SMK memberi **kelengkapan operasional sekolah** (paket, UKK, berita acara, pengacakan, CRUD) tetapi anti-curangnya hanya pencatatan dan tokennya tanpa kedaluwarsa. v2 yang benar = **satu basis kode** yang mengambil arsitektur CBT-Go + keamanan P0 + paket/UKK/laporan SMK + analisis/kompetensi/event dari Exo — dengan setiap klaim fitur dibuktikan uji end-to-end, bukan dari daftar menu.

## 14. Arsitektur yang Diusulkan (Go + React)

### 14.1 Keputusan arsitektur

| Keputusan | Usulan | Alasan / catatan |
|---|---|---|
| Repo | **Monorepo** (1 repo: `backend/` + `frontend/` + `docs/` + `deploy/`) | Satu commit/PR bisa ubah API + UI, satu pipeline, satu artefak. 2 repo ditolak (fragmentasi) |
| Runtime | Browser React SPA → JSON/HTTPS → Go API (`net/http` standar, `database/sql` + PostgreSQL, SQL eksplisit tanpa ORM) → PostgreSQL | — |
| Artefak | **Satu binary Go** (frontend hasil build di-embed via `go:embed`), tanpa server Node di produksi | Vite dev server terpisah dengan proxy saat development |
| Komunikasi | REST JSON | Frontend TIDAK PERNAH menghitung nilai — nilai selalu dihitung backend |
| Mode semi-daring | Satu server di sekolah, satu database sumber kebenaran tunggal | **Tidak ada sinkronisasi offline antar-server di v2** (risiko jawaban ganda/timer tidak konsisten); gantinya autosave per soal + toleransi putus sesaat + pause/resume |
| Monitor real-time | WebSocket (pengganti polling 10 detik) | — |
| File | `storage/` (bukan DB) | Backup mencakup DB + `storage/` |
| Yang TIDAK diusulkan | Microservices, multi-database/multi-server, template HTML server-side seperti CBT-Go v1 untuk UI baru, gerbang lisensi ala Exo | — |

### 14.2 Struktur folder monorepo

```
cbt-smk-v2/
├── backend/
│   ├── cmd/server/                  # entry point
│   ├── internal/
│   │   ├── config/ auth/ rbac/ model/ store/ service/ handler/
│   │   ├── exam/ scoring/ ukk/ report/ analysis/ upload/ ws/
│   ├── migrations/                   # bernomor urut, sekali jalan, tidak pernah diedit setelah rilis
│   └── seed/                        # data awal minimal: role + permission, BUKAN akun demo
├── frontend/
│   └── src/
│       ├── main/ api/ routes/ pages/ components/
│       ├── features/{exam, monitor, correction, reports}/
│       ├── hooks/ store/ styles/    # styles = token design system
├── docs/ + adr/                     # ADR = Architecture Decision Records
├── db/
├── deploy/
└── scripts/
```

**Aturan batas folder:** handler tipis (validasi → service → respons); service tidak tahu HTTP; store satu-satunya tempat SQL; `frontend/src/api/` satu-satunya tempat fetch; state ujian di backend; timer browser hanya tampilan.

### 14.3 Konvensi kode

- Nama variabel/fungsi/tabel/endpoint/commit type Bahasa Inggris; komentar kode + dokumen + README + pesan commit + teks UI + pesan error **Bahasa Indonesia**; log teknis satu bahasa konsisten.
- **Go:** `gofmt` + `go vet`; package huruf kecil satu kata; file `snake_case.go` satu tanggung jawab (bukan `handler.go` 700+ baris seperti CBT-Go v1); error dibungkus konteks; tabel/kolom `snake_case` jamak; JSON `snake_case` 1:1 dengan kolom; SQL parameter binding (cegah injection); transaksi untuk operasi multi-tabel (submit = simpan jawaban + hitung nilai + tutup sesi dalam satu transaksi); hasil ujian tidak pernah dihapus fisik.
- **React/TypeScript wajib** (bukan JS polos); komponen `PascalCase.tsx`; satu komponen satu file; tidak ada duplikasi salinan jawaban; validasi client untuk UX + backend untuk keamanan; aksesibilitas + responsif; tidak ada emoji sebagai ikon; tidak ada `console.log` sisa / data rahasia di log browser.
- **API:** REST kata benda jamak; `/api/v1/...` sejak awal; respons konsisten; waktu disimpan UTC, ditampilkan Asia/Jakarta; timer dari cap waktu server; migrasi bernomor urut sekali jalan, tidak pernah diedit setelah rilis; seed produksi hanya role + permission + settings default tanpa akun demo.
- **Git:** `main` selalu bisa rilis + branch fitur `fitur/nama-singkat`; commit `tipe: deskripsi singkat Bahasa Indonesia` (feat/fix/docs/refactor/test/chore); PR berisi cara uji manual + screenshot bila ubah UI + kasus uji bila sentuh penskoran; keputusan besar sebagai ADR.

### 14.4 Riset Go backend (detail — dari laporan Researcher Go Backend)

**Catatan versi (terverifikasi via pencarian web, Okt 2026):** Echo mayor terbaru v5 (sejak 18 Jan 2026; v4 hanya update keamanan/bug sampai 31 Des 2026), Gin v1.12.0, Fiber v2.52.15 dalam satu benchmark / lini v3.x sudah rilis (mis. v3.4.0), `validator` v10.30.5, `coder/websocket` v1.8.15 (Juni 2026), Go 1.27.1 pada rig benchmark. Angka benchmark di bawah adalah hasil benchmark pihak ketiga, bukan jaminan di server sekolah — bottleneck nyata CBT hampir selalu database, bukan framework.

#### Framework: stdlib `net/http` vs Echo vs Gin vs Fiber

| Aspek | stdlib `net/http` | Echo | Gin | Fiber |
|---|---|---|---|---|
| Basis | Standar library Go | `net/http` | `net/http` | `fasthttp` (bukan `net/http`) |
| Routing | `ServeMux` Go 1.22+: pola `METHOD /path/{param}`, wildcard | Radix tree, group route, middleware per-route/group | `httprouter`-based, sangat cepat, group route | Radix, gaya Express.js |
| Binding / validasi | Tulis sendiri (`encoding/json`, parsing manual) | Bawaan: bind JSON/XML/form + integrasi validator | Bawaan: binding + validasi via tag | Bind bawaan + validator via `StructValidator` |
| Error handling | Manual per handler | Terbaik: handler mengembalikan `error`, ditangani terpusat | Paling membingungkan: pola `Abort`, `c.Error`, tidak mengembalikan error | Handler mengembalikan `error`, cukup baik |
| Middleware | Fungsi pembungkus `http.Handler` | Kaya, bawaan: logger, recover, CORS, rate limit, JWT, CSRF | Paling kaya ekosistemnya | Kaya, gaya Express |
| Kompatibilitas ekosistem | 100% | 100% | 100% | **Tidak** — middleware/library `net/http` (termasuk banyak library WebSocket, `pprof`, OpenTelemetry) tidak langsung bisa dipakai / perlu adapter |
| Ukuran binary | Paling kecil (nol dependensi framework) | Kecil–sedang | Paling besar di kelompok `net/http` — satu studi kasus: melepas Gin menurunkan binary ~45 MB → ~27 MB (~40%) | Kecil–sedang, tapi `fasthttp` sendiri tidak ringan |

**Performa (benchmark localhost, tanpa DB/TLS — layanan identik, rute 2 parameter, JSON kecil):** Fiber ~205.958 req/detik (0,034 ms); Echo ~169.826 (0,042 ms); Gin ~169.715 (0,042 ms); `net/http` ServeMux ~169.708 (0,042 ms). **Interpretasi:** keempat stack berbasis `net/http` dalam selisih ~4%. Fiber ~21% lebih cepat di workload localhost ini, tetapi selisih itu langsung tidak relevan begitu satu request melakukan satu query database (0,5–5 ms). Pada benchmark routing murni (203 rute gaya GitHub API), Gin, BunRouter, dan Echo membentuk tier teratas dengan nol alokasi. **Kesimpulan performa:** untuk CBT, perbedaan performa antar-framework **bukan** penentu. Penentunya adalah query database, pool koneksi, dan desain WebSocket/monitor.

**REKOMENDASI Researcher Go Backend: Echo v5 (di atas `net/http`).** Alasan: (1) produktivitas untuk API sebesar CBT — binding, route group (`/api/v1/admin`, `/api/v1/peserta`, `/api/v1/pengawas` dengan middleware berbeda), error handling terpusat; (2) kompatibel penuh dengan ekosistem `net/http` (syarat untuk `coder/websocket`, `pprof`, middleware standar); (3) performa cukup; (4) mudah di-test (handler mengembalikan error, context berupa interface mudah di-mock); (5) binary tetap kecil. **Kapan stdlib tetap lebih tepat:** jika tim memutuskan meminimalkan dependensi sampai nol dan menerima boilerplate helper sendiri (routing + binding + validasi + error envelope ditulis sekali di paket internal dan dipakai konsisten). **Fiber tidak direkomendasikan:** bukan `net/http`; jebakan goroutine (context Fiber dipakai ulang/pooled — nilai dari context tidak boleh dipakai di goroutine lain tanpa disalin, sangat mudah salah pada pola hub WebSocket/broadcast monitor); testing lebih sulit; migrasi v2→v3 membawa breaking change. **Aturan lapisan (apa pun framework-nya):** framework hanya boleh muncul di lapisan transport (handler + routing + middleware). Lapisan bisnis/domain hanya bergantung pada `context.Context` standar dan tipe domain — tidak mengimpor framework.

#### Auth: Session Cookie vs JWT — DUA USULAN YANG BERTENTANGAN (perlu keputusan, Bab 24)

| Aspek | Session Cookie (server-side) | JWT stateless (Bearer) |
|---|---|---|
| Penyimpanan state | Server (tabel sesi di PostgreSQL); client hanya pegang ID sesi acak (opaque) | Client pegang seluruh token; server tidak menyimpan state |
| Biaya per request | Satu lookup sesi (DB) | Verifikasi tanda tangan (CPU, murah) |
| Revokasi / logout paksa | **Instan:** hapus baris sesi → langsung tidak berlaku | **Sulit:** harus tunggu kedaluwarsa, atau denylist / rotasi kunci |
| Logout di semua perangkat | Mudah | Perlu denylist per `jti` atau rotasi kunci |
| Ukuran di wire | Kecil (~20 byte ID) | Besar (300–1.500 byte) |
| Skala lintas banyak service | Butuh session store bersama | Gratis — service mana pun dengan kunci publik bisa verifikasi |
| Klien non-browser | Canggung | Pas secara natural |
| Risiko XSS | Aman jika `HttpOnly` | Berbahaya jika disimpan di `localStorage` |
| Risiko CSRF | Ada — wajib proteksi CSRF | Rendah jika dikirim via header `Authorization` |
| Audit sesi aktif | Mudah (lihat daftar sesi di DB) | Tidak ada |

**Usulan A (Backend Dev, Bab 15.4):** JWT access token (15 menit, di memori React) + refresh token (cookie `HttpOnly`, rotasi) + CSRF token.

**Usulan B (Researcher Go Backend):** **Session Cookie server-side sebagai auth utama + token ujian terpisah.** Alasan utama — **revokasi instan adalah kebutuhan fungsional CBT, bukan sekadar preferensi keamanan.** Pengawas harus bisa memblokir peserta yang curang **saat itu juga**, mengakhiri ujian peserta, mengeluarkan paksa akun yang dipakai ganda, menonaktifkan sesi saat jadwal dijeda/dibatalkan. Dengan JWT stateless, token yang sudah diterbitkan tetap berlaku sampai kedaluwarsa — peserta yang baru diblokir masih bisa mengirim jawaban selama tokennya hidup. Dengan sesi server-side, penghapusan baris sesi = akses berhenti pada request berikutnya. Alasan pendukung: CBT v2 adalah monolit backend tunggal (kelemahan utama sesi — butuh store bersama — tidak berlaku); klien adalah browser (cookie bekerja natural); tidak ada data sensitif di client; daftar sesi aktif bisa diaudit. **JWT tidak dipakai sebagai auth utama** (hanya jika di masa depan ada kebutuhan service terpisah / klien mobile/CLI — polanya harus hibrida: access token 10–15 menit + refresh token opaque server-side dengan rotasi + denylist; jangan pernah simpan JWT di `localStorage`).

**Desain auth Usulan B (bila dipilih):**
- Login: password di-hash dengan **Argon2id** (`golang.org/x/crypto/argon2`) — rekomendasi utama OWASP Password Storage Cheat Sheet saat ini (parameter minimum OWASP: memori 19 MiB, iterasi 2, paralelisme 1; konfigurasi lebih kuat mis. 64 MiB, iterasi 3, paralelisme 2 dapat dipakai jika server cukup RAM, dengan catatan login serentak ratusan siswa tidak boleh membuat server kehabisan memori — uji beban login sebelum produksi). bcrypt masih dapat diterima sebagai fallback/legacy (cost minimal 10, direkomendasikan 12+), tetapi bukan pilihan untuk sistem baru karena limit input 72 byte dan tidak memory-hard. Jangan pernah memakai MD5/SHA-1/SHA-256 tanpa salt. Perbandingan hash memakai perbandingan waktu-konstan. **(Kontradiksi: Backend Dev mengusulkan bcrypt — perlu keputusan, Bab 24.)**
- Rate limiting + lockout sementara pada endpoint login (batasi percobaan per IP + per akun). Log percobaan login gagal tanpa mencatat password.
- Setelah login berhasil: buat sesi server-side baru, ID acak dari CSPRNG (minimal 32 byte), simpan **hash** ID sesi di database (bukan ID mentah — agar kebocoran DB tidak langsung memberi sesi aktif), kaitkan dengan `user_id`, `role`, `user_agent`, `ip`, `created_at`, `last_seen_at`, `expires_at`. Regenerasi ID sesi setiap login dan setiap kenaikan hak akses (anti session fixation).
- Cookie sesi: `HttpOnly`, `SameSite=Lax` (atau `Strict` jika alur tidak butuh navigasi lintas-situs), `Path=/`, dan `Secure`. **Catatan server lokal:** atribut `Secure` mengharuskan HTTPS — rekomendasi: jalankan backend di belakang HTTPS (reverse proxy dengan sertifikat lokal / CA internal sekolah) meskipun di LAN. Jika benar-benar terpaksa HTTP murni di LAN, `Secure` tidak bisa dipakai — dokumentasikan sebagai penurunan keamanan yang diketahui, jangan pernah membuka server seperti itu ke internet.
- Masa berlaku: idle timeout (mis. 30 menit tanpa aktivitas untuk peserta ujian; lebih pendek untuk admin) + absolute timeout (mis. 8–12 jam) — sesi berakhir mana pun yang tercapai lebih dulu. Sliding renewal pada aktivitas, tetapi absolute timeout tidak diperpanjang. Logout: hapus baris sesi + hapus cookie. Logout semua perangkat: hapus semua sesi milik user.
- **Token ujian (berbeda dari token login — jangan dicampur):** token opaque satu-kali-per-peserta-per-jadwal. Terikat ke satu `user_id` + satu `jadwal_id` + satu attempt aktif; memiliki masa berlaku (jendela jadwal + toleransi); satu attempt aktif per peserta per jadwal (attempt kedua ditolak atau attempt pertama diakhiri sesuai kebijakan); setiap pemakaian dicatat; token dinonaktifkan saat ujian selesai / diblokir / jadwal berakhir. Alur: peserta login (sesi) → memasukkan token ujian → server membuat attempt → semua request menjawab soal wajib membawa sesi + attempt yang valid. Pemisahan ini membuat pencabutan akses ujian (blokir) tidak perlu mencabut login, dan sebaliknya.
- Sesi pengawas/admin: RBAC di middleware + di lapisan service (peran saja tidak cukup — periksa juga kepemilikan/scope); proteksi CSRF wajib untuk semua request yang mengubah state (token CSRF per-sesi via header + `SameSite` sebagai pertahanan lapis kedua); header keamanan (`Content-Security-Policy`, `X-Content-Type-Options: nosniff`, `Referrer-Policy`, `X-Frame-Options`/`frame-ancestors`, `Strict-Transport-Security` jika HTTPS); timeout idle lebih ketat untuk admin/pengawas; log audit untuk aksi sensitif.
- WebSocket dan auth: autentikasi dilakukan **saat handshake** dengan memverifikasi cookie sesi yang sama (bukan token terpisah di URL, karena token di URL bisa bocor ke log/history). Setelah upgrade, periksa peran + scope ruangan/jadwal sebelum mengizinkan subscribe ke room monitor.

#### WebSocket untuk monitor real-time: `coder/websocket` + pola Hub

**Rekomendasi: `github.com/coder/websocket`** (penerus `nhooyr.io/websocket`; rilis terverifikasi v1.8.15, Juni 2026). Alasan: terawat aktif, nol dependensi transitif, lisensi permisif (ISC); native `context.Context` pada semua operasi baca/tulis; aman untuk tulis konkuren secara internal; berbasis `net/http` standar (langsung cocok dengan Echo); lulus uji kesesuaian protokol WebSocket (Autobahn), mendukung kompresi `permessage-deflate`. **Yang ditolak: `gorilla/websocket`** — repository diarsipkan sejak 2022 dan tidak lagi terawat; API mendahului `context`; tulis konkuren membutuhkan mutex eksternal. **SSE dipertimbangkan dan ditolak** untuk kasus ini (lebih sederhana, cukup jika arus murni server → client; tetapi monitor CBT butuh juga arah client → server yang andal — heartbeat, ack, sinyal fokus — dan latensi rendah dua arah; SSE bisa jadi pelengkap, bukan pengganti).

**Arsitektur Hub:** satu goroutine Hub per server yang menjadi satu-satunya pemilik daftar koneksi; semua goroutine lain berkomunikasi dengan Hub lewat channel. Per koneksi: satu goroutine pembaca + satu goroutine penulis (atau satu antrean kirim per koneksi); tulis ke koneksi hanya dari satu goroutine penulis. **Room:** pengawas subscribe ke room per jadwal (`exam:{jadwal_id}`) atau per ruangan (`room:{ruang_id}`); peserta hanya terhubung ke room ujian-nya sendiri dan hanya menerima pesan yang ditujukan kepadanya; pengawas tidak boleh menerima data ruangan yang bukan tugasnya — otorisasi room diperiksa saat subscribe. **Heartbeat/ping-pong:** server mengirim ping berkala; koneksi tanpa pong dalam batas waktu dianggap mati. **Backpressure:** antrean kirim per koneksi berbatas; jika client lambat dan antrean penuh, putuskan koneksi lambat itu (slow consumer) daripada memblokir Hub. **Reconnect:** client React reconnect dengan backoff eksponensial + jitter; saat reconnect, client meminta snapshot terbaru (pesan selama terputus tidak dijamin terkirim ulang; jangan implementasikan antrean replay tak terbatas). **Graceful shutdown:** Hub menutup semua koneksi dengan kode penutupan normal + waktu drain singkat. **Skala:** satu Hub dalam satu proses cukup untuk ratusan koneksi (goroutine Go murah). Jika suatu hari multi-instance, Hub harus pindah ke broker (Redis Pub/Sub atau Postgres `LISTEN/NOTIFY`) — desain Hub dengan antarmuka publish/subscribe agar penggantian broker hanya menyentuh satu tempat.

**Pola pesan (envelope JSON konsisten):** `type`, `room`/`jadwal_id`, `payload`, `timestamp` (waktu server, UTC), `request_id`, versi protokol (`v: 1`).
- Server → pengawas: `presence.online`/`presence.offline`, `progress.answered` (nomor soal + status ragu-ragu, **bukan** isi jawaban), `progress.current_question` (throttle), `violation.tab_switch`/`violation.focus_lost`/`violation.fullscreen_exit`, `violation.device_changed`, `exam.blocked`, `exam.submitted`/`exam.finished`, `monitor.snapshot` (snapshot penuh saat subscribe/reconnect), `exam.paused`/`exam.resumed`, `error`.
- Server → peserta: `exam.block` (alasan + instruksi), `exam.pause`/`exam.resume`, `exam.time_sync`, `exam.force_submit`.
- Client → server (minimal): `heartbeat`, `ack` (konfirmasi penerimaan pesan penting, mis. `exam.block`), `client.violation` (laporan client — **tidak boleh dipercaya begitu saja**; server mencatatnya sebagai sinyal, memvalidasi, dan keputusan blokir diambil server; client yang dimodifikasi bisa saja tidak melaporkan; server juga mendeteksi anomali dari sisi server: tidak ada heartbeat, tidak ada aktivitas, pola jawaban tidak wajar).
- **Aturan data:** pelanggaran dan event penting **tetap ditulis ke PostgreSQL** (tabel audit/event), bukan hanya disiarkan via WebSocket — WebSocket adalah kanal notifikasi, database adalah sumber kebenaran. Jangan kirim jawaban peserta, kunci jawaban, atau data pribadi yang tidak perlu lewat WebSocket. Batasi ukuran pesan masuk (mis. maks beberapa KB) dan rate pesan per koneksi.
- **Fallback jika WebSocket gagal:** client React fallback ke polling REST berkala (5–10 detik) untuk panel monitor dan pengiriman event via REST untuk peserta. Uji fallback ini secara eksplisit — jangan anggap WebSocket selalu tersedia di jaringan sekolah.

#### Upload file: hasil kerja praktik UKK & import soal

**Batas ukuran (dapat dikonfigurasi via env, bukan hardcode):** import soal (CSV/JSON/Excel) 5–10 MB per file, 1 file per aksi import; hasil kerja praktik UKK (dokumen/gambar/arsip) 10–25 MB per file, total 100 MB per peserta per ujian praktik (dapat dikonfigurasi per paket soal); foto/dokumen pendukung 5–10 MB (termasuk dalam total di atas); kompres di client sebelum upload jika memungkinkan. Batasi di **tiga lapis**: konfigurasi server (`MaxBytesReader`/limit body framework), validasi di handler sebelum parsing, batas di reverse proxy. Batasi juga jumlah file per request dan jumlah soal per file import (mis. maks 2.000 soal per file). Timeout upload terpisah dan lebih longgar dari timeout API biasa, tetapi tetap berbatas. Upload harus **streaming ke disk** (atau buffer berbatas), bukan memuat seluruh file ke RAM.

**Validasi berlapis (jangan percaya client):** (1) ukuran — tolak sebelum membaca isi jika `Content-Length` melebihi batas; (2) ekstensi whitelist per jenis upload (import soal: `.csv`, `.json`, `.xlsx`; UKK: `.pdf`, `.docx`, `.xlsx`, `.pptx`, `.jpg`, `.jpeg`, `.png`, `.zip`; tolak `.exe`, `.bat`, `.sh`, `.html`, `.js`, `.php`, file macro Excel `.xlsm`, `.xlsb`); (3) tipe MIME via sniffing isi (magic bytes) — jangan percaya ekstensi dan header `Content-Type` client; (4) validasi struktur isi (import soal: parse + validasi skema; UKK dokumen: verifikasi file bisa dibuka/di-parse; arsip ZIP: batasi jumlah entri, ukuran total hasil ekstrak — anti zip bomb, tolak path traversal `../`, tolak symlink); (5) pemindaian antivirus (mis. ClamAV) bila tersedia; bila tidak, dokumentasikan sebagai risiko yang diterima dan pastikan lapisan lain tetap jalan (penyimpanan di luar webroot, serve via handler terautentikasi, tidak pernah dieksekusi); (6) nama file — abaikan nama asli client untuk penyimpanan; simpan dengan nama acak (UUID) + ekstensi yang sudah divalidasi; nama asli hanya metadata; (7) kuota & rate limit (jumlah upload per peserta per jam, total penyimpanan per sekolah).

**Penyimpanan (server lokal sekolah): filesystem lokal + metadata di PostgreSQL.** File disimpan di direktori khusus di luar webroot (mis. `/data/uploads/...`), struktur `ujian_id / peserta_id / uuid-file`; file **tidak pernah** bisa diakses lewat URL langsung. Metadata di PostgreSQL: pemilik, ujian/soal terkait, nama asli, nama tersimpan, ukuran, MIME tervalidasi, hash SHA-256 isi file, waktu upload, status (pending/valid/quarantine/rejected). Hash SHA-256 untuk deteksi duplikat, verifikasi integritas saat unduh, audit. Pengambilan file hanya lewat endpoint terautentikasi + otorisasi, header `Content-Disposition: attachment` + `Content-Type` dari metadata tervalidasi + `X-Content-Type-Options: nosniff`; jangan pernah sajikan file upload sebagai `inline` HTML/JS. File tidak pernah dieksekusi/di-`include`/di-unzip ke lokasi yang bisa dieksekusi. **Cadangan:** direktori upload masuk rencana backup bersama database (backup metadata tanpa file, atau sebaliknya, sama-sama bermasalah). **Object storage (S3/MinIO)** hanya jika deployment menjadi multi-server; untuk satu server sekolah, object storage menambah kompleksitas tanpa manfaat sepadan; abstraksikan akses file di belakang satu antarmuka penyimpanan.

**Import soal — pola khusus:** pratinjau sebelum commit (dry-run: parse, validasi semua baris, tampilkan ringkasan jumlah valid + baris bermasalah + alasannya + contoh baris gagal) — baru commit jika operator menyetujui; jangan pernah commit separuh file lalu melaporkan error. Transaksional: import yang disetujui berjalan dalam satu transaksi database; gagal di tengah → rollback. Tandai bank soal yang diimpor dengan sumber + waktu + pengimpor (audit). Sanitasi isi: teks soal/jawaban dibersihkan dari HTML/JS berbahaya (soal ditampilkan ke ratusan siswa — XSS lewat isi soal adalah risiko nyata); bila soal mendukung format kaya, izinkan hanya subset tag yang aman dan terdaftar. Jangan jalankan rumus/macro dari file spreadsheet — parse sebagai data saja.

#### Validasi input & error handling

**Validasi berlapis:** (1) handler/transport — parsing body, batas ukuran body, tipe JSON benar, field wajib ada; decoder menolak field tak dikenal (agar typo client ketahuan) dan menolak body terlalu besar; (2) validasi struct — aturan deklaratif per field (wajib, panjang, rentang, format, enum, antar-field); library `go-playground/validator` v10 (standar de facto ekosistem Go); (3) validasi domain/bisnis — aturan yang tidak bisa dinyatakan sebagai tag struct (jadwal masih dibuka, peserta terdaftar di jadwal itu, token berlaku, soal termasuk dalam paket, nilai dalam rentang, pengawas berwenang atas ruangan ini) — di lapisan service/domain, **bukan** hanya di handler; (4) database — constraint sebagai jaring pengaman terakhir (`NOT NULL`, `UNIQUE`, `CHECK`, foreign key, tipe kolom tepat).

Prinsip: jangan percaya client (validasi frontend hanya untuk UX, bukan keamanan); gagal cepat, gagal jelas (kembalikan semua error field sekaligus); semua query berparameter (jangan pernah merangkai SQL dari input user; dengan `pgx` + `sqlc`, parameterisasi adalah default — jangan melanggarnya dengan `fmt.Sprintf` ke dalam query); validasi output juga (encode output agar tidak menjadi XSS); idempotensi untuk aksi kritis (submit jawaban dan submit ujian harus idempotent — kunci idempotensi per attempt — agar retry jaringan / klik ganda tidak menciptakan dua nilai / dua attempt).

**Error handling:** error domain bertipe, bukan string bebas (`ErrNotFound`, `ErrValidation`, `ErrForbidden`, `ErrExamBlocked`, `ErrTokenInvalid`, `ErrConflict`, `ErrTooManyAttempts`); wrap dengan konteks, jangan telan error; satu pemeta error terpusat (error domain → status HTTP + kode error stabil + pesan aman); envelope respons konsisten (sukses: `data` + `meta` untuk paging; error: `code` stabil — mis. `VALIDATION_ERROR`, `EXAM_BLOCKED`, `TOKEN_EXPIRED` — + `message` ramah user Bahasa Indonesia tanpa detail internal + `details` daftar error per field); client React hanya boleh bergantung pada `code`, bukan teks pesan; jangan bocorkan detail internal ke client (query SQL, nama tabel/kolom, stack trace, path file server, detail driver DB) — detail penuh hanya masuk log server (dengan `request_id`); bedakan error yang diharapkan vs bug (middleware `recover` menangkap panic, mencatat stack trace ke log, mengembalikan 500 generik); `request_id` di semua tempat; jangan mengembalikan error database mentah (terjemahkan ke error domain bermakna); timeout & context di semua panggilan keluar (query DB dan panggilan eksternal memakai `context.Context` dengan timeout dari request; jika client disconnect, pekerjaan yang tidak perlu dibatalkan).

#### Logging & konfigurasi; migrasi DB; akses DB

- **Konfigurasi: environment variables** — satu sumber kebenaran; semua yang berubah antar-lingkungan (DSN database, port, secret sesi, batas upload, URL frontend untuk CORS, level log) berasal dari env; jangan hardcode nilai lingkungan di kode. Struktur konfigurasi tervalidasi — saat startup, baca env ke satu struct, validasi, **gagal cepat (fail-fast)** jika tidak valid. Nilai default yang aman untuk dev, tetapi **tanpa default untuk secret** — secret kosong harus membuat startup gagal. File `.env` hanya untuk development lokal, tidak pernah di-commit; produksi memakai env yang disuntikkan sistem (systemd unit, Docker env, secret manager); sediakan `.env.example` (nama variabel + contoh nilai palsu + penjelasan, tanpa nilai asli). Jangan log konfigurasi (secret tidak boleh muncul di log, pesan error, atau endpoint debug). Pisahkan konfigurasi per lingkungan (`APP_ENV`); perilaku berbahaya (mode debug verbose, endpoint debug) nonaktif secara default di produksi.
- **Logging: `log/slog`** (stdlib, sejak Go 1.21) dengan handler JSON untuk produksi dan handler teks untuk development. Alasan: nol dependensi, cukup cepat, API standar, terintegrasi dengan `context.Context`; jika logging menjadi bottleneck terukur, ganti handler `slog` ke backend berperforma tinggi (mis. `zerolog`/`zap`) tanpa mengubah call site. `logrus` tidak direkomendasikan untuk proyek baru (jauh lebih lambat, banyak alokasi). Aturan: terstruktur (JSON) di produksi (field konsisten: `time`, `level`, `msg`, `request_id`, `user_id`, `exam_id`/`jadwal_id`, `method`, `path`, `status`, `duration_ms`, `error`); level disiplin (DEBUG/INFO/WARN/ERROR); satu middleware request logging; **jangan pernah log data sensitif** (password, token sesi/JWT, isi jawaban peserta, token ujian, PII yang tidak perlu); log audit terpisah dari log teknis (aksi sensitif ditulis sebagai event audit ke database — tabel audit, append-only, aktor + waktu + sebelum/sesudah ringkas — bukan hanya ke file log); rotasi & retensi diserahkan ke sistem (logrotate/journald); jangan log di hot path yang tidak perlu (log peristiwa penting — mulai/selesai/blokir/pelanggaran — bukan setiap keystroke).
- **Migrasi database: `golang-migrate`** (paket `github.com/golang-migrate/migrate`) — file SQL `up`/`down` berversi, dijalankan berurutan. Alasan untuk CBT v2: tim kecil, satu database PostgreSQL, server sekolah yang mungkin offline — SQL murni yang bisa dibaca siapa pun adalah keunggulan terbesar; tanpa ketergantungan akun/layanan eksternal. Atlas ditolak untuk v2 (diff otomatis, drift detection tidak sebanding dengan kompleksitasnya untuk satu aplikasi CBT; beberapa alur memerlukan akun/login). Manual (SQL dijalankan tangan) **tidak direkomendasikan** (tidak ada pelacakan versi, drift tidak terdeteksi, rollback tidak terdefinisi). Aturan migrasi: satu tool per repository; pin versi tool; nama file `{versi}_{deskripsi}.up.sql` + `.down.sql` berpasangan; satu concern per migrasi; migrasi immutable (yang sudah di-apply tidak boleh diedit; perbaikan = migrasi baru); DDL di PostgreSQL bersifat transaksional (gagal → rollback); waspadai flag `dirty` pada `golang-migrate`; jangan menaruh logika bisnis di migrasi; jangan menghapus data produksi di dalam migrasi; perubahan besar (indeks pada tabel besar, perubahan tipe kolom) memakai pola aman (tambah kolom baru → backfill bertahap → tukar; `CREATE INDEX CONCURRENTLY` — tidak bisa di dalam blok transaksi, pisahkan migrasinya); **jangan auto-migrate tanpa pengaman di produksi** (di development auto-migrate saat startup boleh; di produksi jalankan migrasi sebagai langkah terpisah sebelum deploy, atau perintah `migrate` eksplisit; untuk satu server sekolah, auto-migrate saat startup masih dapat diterima **jika** dilindungi advisory lock — tetapi perintah migrasi eksplisit tetap lebih mudah diaudit); pisahkan **seed** dari migrasi (migrasi = skema; seed = data awal, dijalankan terpisah dan idempotent); uji migrasi di CI terhadap PostgreSQL asli (naik + turun untuk setiap migrasi, lalu pastikan aplikasi jalan setelah migrasi). Alternatif yang masih dapat diterima: `goose` (SQL + opsi migrasi Go untuk backfill) — tetapi tetap satu tool saja.
- **Akses database: `pgx` v5 (`pgxpool`) + `sqlc`.** `pgx` v5 adalah driver PostgreSQL Go murni (tanpa CGO) — lebih cepat dari `database/sql` generik, mendukung fitur spesifik Postgres (protokol biner, `COPY` untuk import massal, `LISTEN/NOTIFY`), pool concurrency-safe. `sqlc` mengompilasi query SQL di file `.sql` menjadi fungsi Go bertipe — kesalahan SQL ketahuan saat codegen (compile time), tanpa overhead runtime ORM. **GORM ditolak** untuk v2 sebagai lapisan utama: ORM berbasis refleksi menyembunyikan SQL sebenarnya (risiko N+1, query tidak efisien); CBT memiliki query kritis yang butuh SQL presisi (penguncian baris, transaksi submit, penilaian, `FOR UPDATE SKIP LOCKED` untuk antrean). Satu `pgxpool.Pool` dibuat sekali di `main` dan di-inject ke repository; jangan pernah memegang koneksi melewati batas satu request. Ukuran pool disesuaikan dengan database (mis. `MaxConns` 20–30 untuk server sekolah; ratusan siswa tidak berarti ratusan koneksi DB karena request HTTP jauh lebih singkat dari ujian itu sendiri). Gunakan `database/sql` + `pgx/v5/stdlib` hanya jika ada library yang menuntut `*sql.DB`. **(Kontradiksi: laporan lain menyebut `database/sql` + `lib/pq` — perlu keputusan, Bab 24.)**

#### Struktur paket Go yang disarankan (`cmd/` / `internal/` / `pkg/`)

Pola `cmd`/`internal`/`pkg` dari `golang-standards/project-layout` adalah **konvensi komunitas, bukan standar resmi Go** — tetapi dipahami luas dan ditegakkan sebagian oleh compiler (paket di bawah `internal` tidak bisa diimpor dari luar modul). Gunakan secukupnya.
- **`cmd/` — titik masuk aplikasi (tipis):** satu subdirektori per binary. Untuk CBT v2, minimal satu: `cmd/server/` (API + WebSocket + penyajian API untuk React). Opsional: `cmd/migrate/`, `cmd/seed/`. Isi `cmd/...` minimal: hanya memuat konfigurasi, merakit dependency (wiring), menjalankan server. **Tidak ada logika bisnis di sini.**
- **`internal/` — kode privat aplikasi (sebagian besar kode):** organisasi **per fitur/domain, bukan per jenis teknis**. Hindari paket raksasa `controllers`/`services`/`models` di tingkat atas yang mencampur semua fitur.
  - `internal/config/` — struct konfigurasi, pembacaan + validasi env. Satu-satunya tempat yang membaca env.
  - `internal/app/` atau `internal/server/` — perakitan aplikasi: pool DB, logger, router Echo, middleware global, pendaftaran rute, graceful shutdown (SIGINT/SIGTERM, hentikan HTTP + WebSocket dengan timeout drain). Lapisan wiring, bukan bisnis.
  - `internal/transport/http/` — lapisan transport HTTP: definisi rute, handler Echo, middleware (auth, RBAC, CSRF, rate limit, request_id, recover, logging, CORS), DTO request/response, pemeta error HTTP, binding/validasi tingkat transport. **Satu-satunya tempat yang mengimpor Echo.** Handler tipis: parsing + validasi transport → panggil service → format respons.
  - `internal/transport/ws/` — lapisan transport WebSocket: handshake + otorisasi upgrade, Hub, room, pola pesan, heartbeat, fallback snapshot. Tidak mengandung aturan bisnis ujian — ia memanggil service untuk keputusan.
  - `internal/domain/` — inti bisnis murni: entitas dan aturan domain (peserta, soal, jadwal, attempt, penilaian, pelanggaran, blokir) sebagai tipe Go biasa. **Tidak mengimpor framework, driver DB, atau library HTTP.** Aturan seperti "attempt hanya satu yang aktif", "blokir setelah N pelanggaran", "nilai dihitung dari jawaban + koreksi manual" tinggal di sini dan dapat di-test tanpa database/HTTP.
  - `internal/service/` (atau `internal/usecase/`) — orkestrasi kasus-pakai: mengoordinasikan domain + repository + transaksi (mis. "mulai ujian", "submit jawaban", "blokir peserta", "impor soal"). Service bergantung pada **interface** repository, bukan implementasi konkret — bisa di-test dengan repository palsu.
  - `internal/repository/` — akses data: implementasi interface repository dengan `pgx` + kode hasil generate `sqlc` (query SQL di file `.sql` terpisah), plus transaksi. Satu-satunya tempat yang tahu SQL. Jangan letakkan SQL di handler/service.
  - `internal/auth/` — autentikasi & sesi: pembuatan/verifikasi ID sesi, hashing password (Argon2id), token ujian, CSRF, RBAC helper. Terpisah dari transport agar bisa dipakai handler HTTP dan handshake WebSocket.
  - `internal/upload/` — upload & penyimpanan file: validasi ukuran/tipe/magic bytes, penamaan acak, tulis streaming ke disk, hash, karantina, pengambilan file terautentikasi. Abstraksi penyimpanan di sini (antarmuka) agar suatu hari bisa diganti object storage tanpa menyentuh service.
  - `internal/migration/` (atau `migrations/` di root) — file migrasi SQL `golang-migrate` + skrip/perintah menjalankannya. File SQL di-embed ke binary atau dijalankan dari direktori yang jelas — putuskan satu cara dan konsisten.
  - `internal/observability/` — logging terstruktur (`slog`), metrik dasar, tracing. Logger dibuat sekali dan di-inject; bukan variabel global.
  - `internal/validator/` — konfigurasi validator bersama (instance `validator` + pesan error field yang konsisten).
  - Aturan dependensi: `cmd` boleh mengimpor `internal` dan `pkg`; `transport` boleh mengimpor `service` dan `domain`, **tidak** sebaliknya; `service` boleh mengimpor `domain` dan interface `repository`, **tidak** boleh mengimpor `transport` atau Echo; `domain` tidak mengimpor apa pun dari `transport`/`service`/`repository`/framework/driver DB; `repository` boleh mengimpor `domain` dan driver DB, **tidak** boleh mengimpor `transport`/`service`; interface didefinisikan di tempat pemakainya (service mendefinisikan interface repository yang ia butuhkan), bukan di paket implementasi.
- **`pkg/` — kode yang boleh dipakai pihak luar (sedikit saja):** hanya untuk kode yang benar-benar generik, stabil, dan masuk akal untuk diimpor proyek lain. Untuk CBT v2, sebagian besar proyek tidak membutuhkan `pkg/` sama sekali. Aturan praktis: **jika ragu, taruh di `internal/`, bukan `pkg/`.**
- **Direktori pendukung:** `api/` (kontrak API: spesifikasi OpenAPI — jika memakai spec-first, OpenAPI adalah sumber kebenaran yang dipakai untuk generate client TypeScript React, mencegah drift antara frontend dan backend); `web/` — tidak dipakai untuk React (frontend terpisah; backend hanya menyajikan API + WebSocket, bukan HTML); `configs/` (contoh konfigurasi — `.env.example`, contoh unit systemd/compose — bukan secret asli); `scripts/` (build, migrasi, backup, deploy); `docs/` (ADR untuk pilihan-pilihan di laporan ini); `testdata/` (fixture untuk test: file CSV soal contoh, file upload contoh valid/tidak valid).
- **Pengujian yang menyertai struktur ini:** test unit untuk `domain` (tanpa DB/HTTP, cepat); test service dengan repository palsu; test handler dengan server Echo test (status, envelope error, validasi, otorisasi); test integrasi terhadap PostgreSQL asli (bukan mock) untuk repository + migrasi; test WebSocket (Hub dengan client palsu — broadcast per-room, otorisasi room, slow consumer); test beban sebelum produksi (login serentak, submit serentak, 200+ koneksi WebSocket monitor — di server sekolah yang sebenarnya, bukan di laptop dev).

**Ringkasan keputusan Researcher Go Backend:** Framework HTTP **Echo v5**; auth utama **session cookie server-side**; token ujian **opaque, server-side, per-attempt**; password **Argon2id**; WebSocket **`coder/websocket` + pola Hub**; upload **filesystem lokal + metadata Postgres**; validasi **`validator` v10 + validasi domain + constraint DB**; error **error domain bertipe + pemeta terpusat + envelope konsisten**; logging **`log/slog` JSON**; konfigurasi **env + validasi fail-fast**; migrasi DB **`golang-migrate` (SQL up/down)**; akses DB **`pgx` v5 + `sqlc`**; struktur **`cmd/` tipis + `internal/` per fitur + `pkg/` minimal**.

**Hal yang perlu dikonfirmasi sebelum implementasi (Go Backend):** (1) apakah server sekolah akan memakai **HTTPS** (agar cookie `Secure` bisa aktif)? Jika tidak, perlu keputusan eksplisit untuk menerima penurunan keamanan tersebut; (2) apakah session store cukup di **PostgreSQL** (satu server), atau perlu Redis? Rekomendasi: PostgreSQL saja untuk v2; (3) batas upload UKK final per jenis praktik (video demo membutuhkan batas jauh lebih besar daripada dokumen) — angka di atas adalah titik awal, bukan keputusan final; (4) kebijakan anti-curang final (berapa pelanggaran sebelum auto-block, apakah pengawas dapat mengubah ambang per jadwal) — harus ditetapkan sebelum Hub WebSocket diimplementasikan.

### 14.5 Riset React frontend (detail — dari laporan Researcher React Frontend)

**Konteks:** Frontend baru untuk CBT SMK v2. Backend tetap Go, frontend disajikan sebagai build statis oleh server Go, tanpa SSR. Target utama: siswa SMK di HP / tablet / laptop. Halaman ujian harus fokus, minim distraksi, dan tahan terhadap kondisi semi-daring (server lokal, jaringan tidak selalu stabil). **Verifikasi versi:** melalui pencarian web, Oktober 2026. Versi yang disebut adalah titik acuan saat riset; tim implementasi wajib mengunci versi di `package.json` / lockfile pada saat scaffolding, bukan mengandalkan `latest` tanpa pin.

#### Tooling: Vite vs CRA vs Next.js

| Aspek | Vite | Create React App (CRA) | Next.js |
|---|---|---|---|
| Jenis alat | Build tool + dev server untuk SPA | Build tool lama berbasis Webpack | Framework full-stack React (routing + SSR/SSG + RSC + API routes) |
| Model rendering utama | Client-Side Rendering (CSR), SPA | CSR, SPA | SSR / SSG / React Server Components |
| Dev server / HMR | Sangat cepat, start < 1 detik, HMR sub-50 ms pada banyak kasus | Lambat, cold boot puluhan detik | Cepat, tapi lebih berat dari Vite untuk SPA murni |
| Hasil build | Folder statis `dist/` — langsung bisa disajikan server statis mana pun | Folder `build/` statis, tetapi toolchain sudah usang | Butuh server Node.js yang jalan terus untuk SSR / fungsi framework; mode export statis ada tapi membatasi fitur |
| Routing | Bawa sendiri (React Router) | Bawa sendiri | File-based routing bawaan |
| SEO | Lemah untuk konten publik (tidak relevan untuk CBT yang di balik login) | Lemah | Kuat untuk halaman publik |
| Status pemeliharaan | Aktif. Acuan riset: Vite 8, berbasis Rolldown | **Deprecated / tidak lagi dipelihara.** Dokumentasi React resmi tidak lagi merekomendasikannya | Aktif. Acuan riset: Next.js 16 (App Router) |
| Kompatibilitas React modern | Mendukung React 19. Acuan riset: React 19.2.x | Bermasalah / breaking untuk React 19 | Mendukung React 19 |
| Kecocokan dengan Go | **Sangat cocok.** Go hanya menyajikan file statis hasil build | Cocok secara teknis, tetapi toolchain mati — jangan dipakai untuk proyek baru | **Tidak cocok.** Menambah server Node.js yang harus jalan berdampingan dengan Go, tanpa manfaat SSR karena CBT berada di balik login |

**Rekomendasi: Vite 8 + React 19 + TypeScript + React Router (acuan riset: v7) + Vitest + Testing Library** untuk pengujian komponen kritis (timer, grid navigasi, logika simpan jawaban). Alasan utama: satu-satunya yang pas dengan arsitektur Go + build statis (Vite selesai bekerja begitu menghasilkan `dist/`; setelah itu Go tidak perlu tahu apa-apa tentang React); CRA sudah mati; kecepatan iterasi (dev server nyaris instan, HMR cepat); SEO tidak relevan (seluruh alur CBT di balik login); bundle runtime lebih kecil untuk aplikasi di balik login. Alternatif yang layak untuk routing: TanStack Router bila tim menginginkan routing yang sangat type-safe, tetapi React Router lebih umum, lebih mudah dipelajari, dan cukup untuk rute CBT yang sedikit. **Yang TIDAK direkomendasikan:** Next.js untuk v2 ini; CRA dalam kondisi apa pun. **Catatan arsitektur:** jangan memasukkan logika bisnis / API ke dalam frontend framework. Semua logika ujian (deadline, nilai, blokir, validasi) tetap di backend Go. Frontend hanya menampilkan dan mengirim.

#### State management: pisahkan server state vs client state

Sebelum memilih library, pisahkan dua jenis state:
- **Server state** — data yang sumber kebenarannya ada di server (daftar soal, jadwal ujian, jawaban yang sudah tersimpan di server, status sesi, hasil). Untuk ini **jangan** memakai Context / Zustand / Redux sebagai cache utama. Gunakan **TanStack Query (React Query)** atau pola fetch + cache yang setara (caching, deduplikasi request, refetch, retry, status loading/error bawaan).
- **Client state** — data yang hidup di browser selama ujian (jawaban yang sedang diketik/dipilih, status per soal, soal aktif, sisa waktu tampil, status simpan, antrean jawaban yang belum terkirim, status pelanggaran). Untuk inilah perbandingan Context vs Zustand vs Redux berlaku. Mencampur keduanya dalam satu store global adalah sumber bug paling umum (data basi, jawaban tertimpa, status simpan tidak sinkron).

| Aspek | React Context (bawaan) | Zustand | Redux Toolkit (RTK) |
|---|---|---|---|
| Ukuran bundle | 0 KB (bawaan React) | Sangat kecil, ~2–3 KB | Paling besar, ~11 KB dengan RTK |
| Boilerplate | Rendah | Sangat rendah | Sedang (slice, action, reducer, store setup) |
| Selektif subscribe | **Tidak.** Setiap perubahan nilai Context me-render ulang semua komponen konsumennya | **Ya.** Komponen hanya me-render ulang bila potongan state yang dipilihnya berubah | Ya, via selector |
| Perilaku pada update frekuensi tinggi | Buruk. Timer yang tick tiap detik + jawaban yang berubah sering akan me-render ulang seluruh subtree | Baik. Timer, jawaban, dan grid navigasi bisa subscribe ke potongan yang berbeda | Baik, tetapi dengan boilerplate lebih banyak |
| Middleware / persist / DevTools | Harus ditulis manual | Tersedia (persist, devtools, immer) | Ekosistem middleware paling lengkap, Redux DevTools dengan time-travel |
| Kurva belajar | Paling rendah | Rendah | Paling tinggi |
| Cocok untuk | State yang jarang berubah dan dibaca luas: tema, bahasa, user login | State UI + state ujian yang berubah sering dan dibaca banyak komponen | Aplikasi enterprise sangat besar, tim besar yang butuh konvensi ketat |

**Rekomendasi: Zustand untuk state ujian + TanStack Query untuk server state + Context hanya untuk yang jarang berubah + `useState` lokal untuk yang lokal.** Pembagian: jawaban per soal, status per soal, soal aktif, flag ragu-ragu, antrean simpan, status simpan → **Zustand** (satu store ujian, dengan selector per komponen); sisa waktu tampil (turunan dari deadline server) → **Zustand** atau state lokal di komponen timer, dihitung dari deadline server + offset; soal, jadwal, sesi, hasil dari server → **TanStack Query**; user login, tema, preferensi tampilan → **Context**; input yang hanya dipakai satu komponen → **`useState`/`useReducer` lokal**. **Mengapa bukan Context untuk state ujian:** Context tidak punya selective subscription — setiap tick timer (1 detik) berpotensi me-render ulang seluruh halaman ujian, termasuk daftar opsi jawaban dan grid navigasi; di HP spek rendah ini terasa sebagai jank/lag. **Mengapa bukan Redux Toolkit:** overkill untuk CBT ini — state ujian strukturnya sederhana; biaya yang dibayar adalah boilerplate, ukuran bundle lebih besar, kurva belajar lebih tinggi. Zustand memberi 90% manfaat dengan sebagian kecil biaya. **Persistensi state ujian:** store Zustand sebaiknya dipersist ke `localStorage` (atau IndexedDB untuk kapasitas lebih besar) sebagai **cadangan pemulihan**, bukan sebagai sumber kebenaran; setelah jawaban terkonfirmasi tersimpan di server, salinan lokal dianggap basi; saat memuat ulang sesi, server selalu menang bila terjadi konflik.

#### Komponen ujian — prinsip: server adalah sumber kebenaran; frontend hanya menampilkan, mengantrekan, dan memulihkan

**Timer countdown — sinkron server, anti-manipulasi jam lokal.** Masalah: jam lokal perangkat siswa bisa diubah, zona waktu bisa berbeda, `setInterval` di browser tidak presisi (di-throttle saat tab tidak aktif, melambat saat perangkat tidur). Pola yang direkomendasikan: (1) server mengirim deadline absolut, bukan durasi — saat sesi dimulai/dimuat, server mengirim `deadline` (waktu berakhir sesi, format waktu absolut, mis. UTC) bersama `serverNow`; (2) frontend menghitung offset sekali, lalu sisa waktu = `deadline - (jamLokalSekarang + offset)` — mengubah jam lokal tidak menambah waktu; (3) re-sinkronisasi berkala dan pada momen penting (tiap 30–60 detik, atau tiap heartbeat/simpan jawaban, dan saat tab kembali aktif); (4) tampilan tick lokal, keputusan di server — saat timer lokal mencapai nol, frontend memanggil auto-submit; server memvalidasi bahwa deadline memang sudah lewat sebelum menerima/menutup sesi; bila server menyatakan sesi sudah berakhir lebih awal (dijeda/diblokir proktor), frontend harus menutup ujian seketika meski timer lokal belum nol; (5) tangani jeda dan tidur perangkat — saat tab kembali aktif, frontend langsung re-sinkronisasi ke server dan menampilkan sisa waktu yang benar, bukan melanjutkan dari angka lama; jangan pernah "mengejar ketertinggalan" dengan menambah waktu; (6) tampilan: countdown besar dan jelas di header ujian, status warna bertahap (normal → peringatan di mis. 10 menit terakhir → kritis di mis. 1 menit terakhir) + penanda waktu server/status sinkronisasi yang halus; jangan tampilkan jam dinding perangkat sebagai acuan; (7) aksesibilitas: timer harus bisa dibaca screen reader tanpa mengumumkannya tiap detik — umumkan perubahan pada ambang penting saja (10 menit, 5 menit, 1 menit, 30 detik).

**Navigasi soal — grid nomor + warna status.** Status per soal (minimal 6 status, dengan warna + bentuk yang dibedakan, bukan hanya warna): soal aktif/sedang dibuka (outline tebal + latar menonjol); sudah dijawab (warna solid, mis. hijau/warna primer); ragu-ragu (warna peringatan + ikon tanda tanya/bendera); dijawab + ragu-ragu (kombinasi — jangan sampai status ini hilang); belum dijawab/belum dikunjungi (netral, latar terang/outline tipis); belum tersimpan (jawaban baru dipilih, belum terkonfirmasi server — indikator kecil titik/spinner mini di atas status dasarnya). Aturan penting: **jangan hanya mengandalkan warna** (sekitar 1 dari 12 laki-laki memiliki buta warna merah-hijau — setiap status juga dibedakan oleh bentuk/ikon/label + legenda status); grid harus tidak me-render ulang tiap detik (subscribe hanya ke status jawaban, bukan ke timer — inilah alasan Zustand dengan selector); responsif (HP: panel bawah/drawer yang bisa dibuka-tutup; tablet/desktop: sidebar; tombol Sebelumnya/Berikutnya tetap tersedia di area soal); skala (untuk 100+ soal, grid harus mendukung scroll internal dan/atau pengelompokan, pertimbangkan virtualisasi); aksesibilitas (setiap tombol nomor soal harus memiliki label yang dapat dibaca screen reader, mis. "Soal 12, sudah dijawab, ragu-ragu").

**Tanda ragu-ragu.** Per soal, bukan global; checkbox/tombol toggle "Tandai ragu-ragu" di area soal — menandainya tidak menghapus jawaban; menghilangkan jawaban tidak otomatis menghilangkan flag ragu-ragu. Persistensi: flag ragu-ragu adalah bagian dari state ujian dan harus disimpan ke server bersama jawaban (atau sebagai field pada jawaban per soal), bukan hanya di memori lokal (jika hanya lokal, flag hilang saat refresh). Filter/ringkasan: di panel navigasi, sediakan filter atau hitungan "X soal ragu-ragu"; saat konfirmasi submit, tampilkan ringkasan (jumlah belum dijawab + jumlah ragu-ragu) dengan opsi "Kembali meninjau" vs "Tetap kumpulkan". Visual: flag ragu-ragu harus terlihat di tiga tempat sekaligus (area soal, grid navigasi, ringkasan akhir).

**Auto-submit saat waktu habis.** Alur: (1) pemicu — timer lokal mencapai nol **atau** server menyatakan sesi berakhir (mana yang lebih dulu); (2) kunci UI seketika (nonaktifkan input jawaban, tombol navigasi, tombol submit manual; tampilkan "Waktu habis — mengumpulkan jawaban…"); (3) flush antrean — kirim semua jawaban yang masih mengantre dalam satu upaya terakhir, dengan timeout yang wajar; bila jaringan putus total, jawaban yang masih di perangkat tetap tersimpan di penyimpanan lokal dan ditandai belum terkirim (jangan menghapusnya); (4) satu kali submit, idempotent (submit dikirim dua kali diperlakukan sebagai satu submit yang sama — via kunci idempotency / status sesi; frontend mencegah submit ganda di UI); (5) konfirmasi dari server adalah penutup resmi — hanya setelah server mengembalikan konfirmasi "sesi ditutup", frontend menampilkan halaman hasil/"ujian selesai"; bila server tidak merespons (jaringan putus), frontend menampilkan "Menunggu konfirmasi server — jawaban Anda tersimpan di perangkat, jangan tutup tab ini" dan terus mencoba ulang di latar belakang; (6) jangan andalkan `beforeunload` untuk auto-submit (prompt "tinggalkan halaman?" tidak dapat diandalkan untuk mengirim data — browser dapat membatalkan request saat halaman ditutup); auto-submit harus terjadi saat timer habis selagi halaman masih terbuka; (7) kasus pause/resume oleh proktor — bila jadwal dijeda, server mengirim status jeda; frontend menghentikan countdown (menampilkan "Ujian dijeda oleh proktor") dan tidak menghitung waktu jeda sebagai waktu ujian; saat dilanjutkan, frontend re-sinkronisasi deadline baru dari server.

**Simpan jawaban — debounce + kirim ke server + perilaku saat internet putus.** Pola yang direkomendasikan: simpan lokal seketika + debounce kirim + antrean + retry + indikator status yang jujur. Tahap: pilih/ketik jawaban → simpan **seketika** ke state lokal + penyimpanan lokal perangkat (localStorage/IndexedDB) — jawaban selamat dari refresh/crash meski jaringan mati; debounce pengiriman (jangan kirim request HTTP tiap ketikan untuk isian/esai — tunggu jeda hening singkat mis. 500–1000 ms setelah ketikan terakhir lalu kirim; untuk pilihan ganda kirim langsung); kirim ke server per soal (atau batch kecil) — server mengembalikan konfirmasi per soal (versi/timestamp); frontend menandai soal "tersimpan" hanya setelah konfirmasi ini; optimistic UI dengan batas (tampilkan jawaban terpilih seketika, tetapi **jangan** menampilkan status "tersimpan" sebelum konfirmasi server — bedakan jelas "terpilih" vs "tersimpan di server"); retry dengan backoff (gagal → antrean, coba ulang dengan jeda membesar mis. 1 dtk, 2 dtk, 5 dtk, bukan spam retry tiap detik); antrean offline (outbox) — saat offline, jawaban menumpuk di antrean lokal yang persisten (IndexedDB lebih baik dari localStorage untuk volume besar / esai panjang); tampilkan indikator "Offline — X jawaban menunggu dikirim"; kembali online — saat event `online`/heartbeat berikutnya berhasil, kirim antrean dalam urutan yang benar (jawaban terbaru per soal menang bila ada duplikat); tampilkan "Mengirim jawaban tertunda…" hingga antrean kosong; konflik — server menang (bila jawaban di server lebih baru, frontend memakai versi server dan memberi tahu siswa; jangan pernah menimpa jawaban server yang lebih baru dengan salinan lokal yang lebih lama tanpa konfirmasi); batas keamanan (batasi ukuran antrean/ukuran jawaban esai di frontend dengan pesan yang jelas bila terlampaui; jangan biarkan penyimpanan lokal penuh secara diam-diam). **Indikator status simpan (wajib ada, di header ujian):** "Menyimpan…" (spinner); "Semua jawaban tersimpan" (centang); "X jawaban belum terkirim" (peringatan) + tombol "Coba kirim ulang"; "Offline" (jelas, bukan sekadar ikon). **Yang harus dihindari:** menyimpan hanya di memori (hilang saat refresh); mengirim tiap ketikan tanpa debounce (banjir request, beban server lokal, baterai boros); menampilkan "tersimpan" tanpa konfirmasi server (kebohongan UI yang baru ketahuan saat nilai hilang); menghapus antrean lokal sebelum konfirmasi server. **Soal esai panjang:** simpan draf lokal pada tiap perubahan (debounce pendek) dan kirim ke server dengan debounce lebih panjang; pertimbangkan penyimpanan draf esai di IndexedDB, bukan localStorage (batas localStorage umumnya ~5 MB mudah terlampaui oleh teks panjang + banyak soal).

#### Desain anti-cheat sisi frontend

**Prinsip utama: tidak ada anti-cheat frontend yang tidak bisa diakali.** Semua di bawah ini adalah *deteksi + pencatatan + penghambat*, bukan *pencegahan mutlak*. Penegakan yang sesungguhnya harus terjadi di server (deadline server, validasi jawaban server, hitung pelanggaran server, blokir server). Frontend yang mengklaim "100% anti-curang" adalah klaim yang salah.

| Mekanisme | Cara kerja | Tujuan |
|---|---|---|
| Deteksi pindah tab / sembunyikan tab | Event `visibilitychange` + `blur`/`focus` pada window; setiap kejadian dicatat dan dikirim ke server sebagai pelanggaran/peringatan | Mendeteksi siswa membuka tab/aplikasi lain |
| Deteksi keluar fullscreen | Event `fullscreenchange`; bila siswa keluar dari fullscreen selama ujian, catat sebagai pelanggaran | Mendeteksi siswa meninggalkan mode ujian |
| Fullscreen saat ujian | Minta fullscreen saat ujian dimulai (harus dari interaksi pengguna, mis. klik "Mulai ujian"); tampilkan tombol "Kembali ke fullscreen" bila keluar | Mengurangi distraksi dan mempersulit membuka jendela lain berdampingan |
| Blokir copy-paste | Cegah event `copy`, `cut`, `paste` di area soal, cegah `contextmenu` (klik kanan); blokir pintasan umum (Ctrl+C/Ctrl+V/Ctrl+X/Ctrl+U/Ctrl+Shift+I/F12) sejauh dimungkinkan browser | Mempersulit menyalin soal keluar dan menempel jawaban dari luar |
| Blokir seleksi teks | Nonaktifkan seleksi teks di area soal via CSS (`user-select: none`) | Mempersulit copy soal dengan drag |
| Peringatan tinggalkan halaman | Prompt konfirmasi browser saat siswa mencoba menutup/me-refresh/navigasi pergi selama ujian | Mencegah kehilangan jawaban karena navigasi tak sengaja |
| Heartbeat + pelanggaran ke server | Kirim heartbeat berkala + event pelanggaran ke server; server yang menghitung jumlah pelanggaran dan memutuskan peringatan vs auto-block | Memastikan pelanggaran tercatat meski frontend dimanipulasi/ditutup |
| Watermark sesi (opsional) | Watermark halus (nama/NIS/ID sesi siswa) di latar halaman ujian | Mempersulit foto layar yang dibagikan tanpa jejak |

**Kebijakan pelanggaran yang direkomendasikan:** jangan menghukum pada pelanggaran pertama yang bisa jadi tidak sengaja (notifikasi OS, klik tak sengaja, Alt+Tab yang salah). Pelanggaran ke-1: peringatan modal yang jelas + wajib klik untuk kembali. Pelanggaran ke-2: peringatan keras + dicatat, dengan pesan bahwa pelanggaran berikutnya akan mengakhiri ujian. Pelanggaran ke-N (ambang yang dikonfigurasi per ujian): auto-block/auto-submit oleh **server**, bukan hanya frontend. Ambang harus dapat dikonfigurasi proktor per jadwal (ujian latihan vs ujian resmi memiliki toleransi berbeda). **Debounce pelanggaran:** event `blur` dan `visibilitychange` sering menyala berpasangan untuk satu tindakan yang sama — gabungkan event yang terjadi dalam jendela sangat singkat (mis. < 1 detik) menjadi satu pelanggaran agar hitungan tidak mengganda. Semua pelanggaran harus memiliki jejak audit di server (siapa, kapan waktu server, jenis pelanggaran, status sesi saat itu) — penting untuk banding/peninjauan proktor.

**Batasan — wajib disampaikan kepada sekolah/proktor:** DevTools tidak bisa diblokir total (pintasan dan menu konteks bisa dicegah, tetapi siswa yang paham DevTools/ekstensi dapat menonaktifkan atau mengakalinya); perangkat kedua tidak terdeteksi (frontend tidak dapat melihat HP kedua, tablet kedua, atau catatan kertas — hanya bisa diatasi dengan pengawasan fisik proktor); menonaktifkan JavaScript (bila JavaScript dimatikan, seluruh deteksi frontend mati — server harus tetap menegakkan deadline dan menolak pengiriman yang tidak valid); manipulasi kode di browser (kode JavaScript yang dikirim ke browser dapat dibaca dan diubah — hitungan pelanggaran di frontend **tidak boleh** menjadi satu-satunya dasar keputusan; server harus memiliki sinyal independen: heartbeat hilang, pola pengiriman jawaban tidak wajar, deadline); false positive (notifikasi sistem, panggilan masuk, baterai rendah, atau klik tak sengaja dapat memicu `blur`/`visibilitychange` — kebijakan harus memberi toleransi dan jalur banding, bukan auto-block pada kejadian pertama); fullscreen tidak bisa dipaksa (browser mengharuskan fullscreen dimulai dari gestur pengguna dan pengguna selalu bisa keluar mis. tombol Esc — frontend hanya bisa meminta dan mendeteksi, tidak mengunci); privasi (jangan membaca isi clipboard, jangan merekam layar/webcam tanpa persetujuan yang jelas dan dasar kebijakan sekolah; bila fitur webcam ditambahkan di masa depan, ia adalah proyek tersendiri — izin, penyimpanan, privasi — bukan sekadar "tambahkan satu event listener"); aksesibilitas (blokir pintasan keyboard secara membabi buta dapat merusak penggunaan screen reader/navigasi keyboard — uji anti-cheat bersama pengguna keyboard dan screen reader). **Kesimpulan:** frontend anti-cheat adalah *lapisan pertama yang jujur tentang keterbatasannya*, bukan benteng. Nilai sebenarnya ada pada: (a) deteksi + pencatatan yang konsisten, (b) kebijakan pelanggaran yang adil dan terkonfigurasi, (c) penegakan server-side, (d) pengawasan fisik proktor di ruangan. Untuk ujian berisiko tinggi, pertimbangkan lockdown browser khusus sebagai proyek terpisah — browser web biasa tidak pernah bisa memberikan jaminan yang sama.

#### UI: Tailwind CSS vs komponen siap pakai (shadcn/ui, MUI, Ant Design)

**Klarifikasi kategori (sering dicampuradukkan):** Tailwind CSS adalah *mesin styling* (utility-first CSS), bukan library komponen — ia menghasilkan CSS; Anda tetap menulis komponen sendiri. MUI / Ant Design adalah *library komponen lengkap* (npm package): komponen jadi beserta sistem desainnya. shadcn/ui adalah *koleksi komponen copy-paste*: CLI menyalin file komponen ke repo Anda (berbasis Radix UI + Tailwind) — Anda memiliki kodenya dan memodifikasinya langsung; bukan dependency npm untuk komponennya sendiri. Ketiganya tidak saling eksklusif pada level yang sama: Tailwind adalah fondasi, shadcn/ui duduk di atasnya, sedangkan MUI / Ant Design adalah alternatif satu paket yang biasanya **tidak** dipadukan dengan Tailwind.

| Aspek | Tailwind CSS saja | Tailwind + shadcn/ui | MUI (Material UI) | Ant Design |
|---|---|---|---|---|
| Pendekatan | Utility class di JSX; semua komponen ditulis sendiri | Utility + komponen copy-paste yang dimiliki sendiri, berbasis primitif aksesibel Radix | Komponen jadi bergaya Material Design (Google) | Komponen jadi bergaya enterprise / back-office |
| Tampilan bawaan | Tidak ada — sepenuhnya tergantung desain Anda | Netral, minimal, modern; mudah dibentuk menjadi identitas CBT sendiri | Sangat khas Material (ripple, elevation, bentuk khas Google) | Sangat khas Ant Design, terasa seperti dashboard admin enterprise |
| Aksesibilitas | Tergantung Anda | Kuat (Radix menangani ARIA, fokus, keyboard, focus trap) | Kuat (mengikuti WAI-ARIA) | Baik, sebagian komponen butuh penyesuaian |
| Bundle | Kecil (utility yang terpakai saja, di-purge) | Minimal — hanya komponen yang disalin yang ikut bundle | Besar (~80 KB+ untuk inti, di luar data grid/date picker lanjutan) | Besar (~60 KB+, CSS-in-JS) |
| Kustomisasi | Total | Sangat fleksibel — ubah file komponen langsung | Kuat via theme object, tetapi melawan default Material butuh usaha | Baik via token tema, tetapi visualnya tetap terasa Ant Design |
| Cakupan komponen | 0 (Anda tulis semua) | 40+ komponen dasar (button, dialog, dropdown, dll.); tidak ada data grid enterprise | 50–80+ termasuk komponen lanjutan + ekosistem MUI X (Data Grid, Date Picker, Charts) | 60–70+ termasuk tabel data kaya, form kompleks, dll. |
| Dark mode | Via CSS variable / class | Bawaan via Tailwind + CSS variable | Bawaan via theme provider | Tersedia, butuh setup tema |
| Kecocokan dengan Tailwind | — (itu sendiri) | Dirancang untuk Tailwind | Tidak dipadukan dengan baik (sistem styling sendiri) | Tidak dipadukan dengan baik |
| Pembaruan | Dependency normal | Manual — karena Anda memiliki kodenya, tidak ada update otomatis dari upstream | Dependency normal | Dependency normal |

**Rekomendasi: Tailwind CSS 4 + shadcn/ui (seperlunya) + komponen ujian custom.** Alasan: (1) CBT ini butuh identitas visual sendiri, bukan tampilan "template admin" — halaman ujian harus terasa tenang, fokus, dan khas sekolah; (2) bundle ringan = penting untuk HP spek rendah; (3) kepemilikan kode (komponen yang dipakai hidup di repo Anda — bila ada bug/kebutuhan khusus, ubah file komponen langsung); (4) aksesibilitas tanpa membangun dari nol (primitif Radix sudah menangani hal-hal yang mudah salah bila ditulis sendiri — focus trap di dialog, navigasi keyboard di dropdown, atribut ARIA — penting untuk dialog konfirmasi submit, drawer navigasi soal, modal peringatan); (5) Tailwind 4 menyederhanakan setup (integrasi Vite via plugin Vite, tema didefinisikan via CSS `@theme`, tanpa file konfigurasi JavaScript yang berat). **Yang TIDAK direkomendasikan untuk v2:** MUI dan Ant Design sebagai fondasi utama (keduanya sangat baik untuk dashboard enterprise/back-office yang butuh banyak komponen data berat dengan cepat, tetapi untuk CBT yang mengutamakan halaman ujian yang fokus, ringan, dan beridentitas sendiri, keduanya memberi lebih banyak beban daripada manfaat). Bila di masa depan halaman admin membutuhkan tabel data yang sangat kompleks, pertimbangkan library tabel headless (mis. TanStack Table) di atas Tailwind, bukan mengganti seluruh sistem desain. **Penggunaan shadcn/ui yang direkomendasikan — selektif, bukan semua komponen:** ambil komponen yang generik dan sulit dibuat aksesibel dengan benar (Button, Dialog/Alert Dialog, Dropdown Menu, Tooltip, Toast/Sonner, Sheet/Drawer, Progress, Badge, Tabs). **Jangan** memaksakan komponen generik untuk komponen inti ujian — kartu soal, opsi jawaban, grid navigasi soal, timer, dan panel ragu-ragu harus **komponen custom** yang dirancang khusus untuk ujian (ukuran ketuk besar, hierarki visual jelas, status yang mudah dibaca).

**Prinsip desain halaman ujian (fokus & minim distraksi):** satu soal per layar pada mode ujian, dengan navigasi yang jelas (jangan menampilkan semua soal dalam satu scroll panjang); hierarki visual (soal paling menonjol → opsi jawaban besar, mudah diketuk, jarak antar opsi cukup → aksi sekunder → navigasi grid tersembunyi di drawer pada HP); ruang kosong yang cukup, kontras tinggi (teks soal harus mudah dibaca pada layar kecil di bawah cahaya ruang kelas; hindari teks abu-abu muda untuk konten utama); target ketuk minimal 44×44 px untuk opsi jawaban dan tombol navigasi (opsi jawaban harus bisa diketuk di seluruh area kartunya, bukan hanya pada radio button kecil); responsif yang benar (HP 1 kolom, grid soal di drawer, header ringkas; tablet 1 kolom + sidebar yang bisa dilipat; laptop soal di tengah + sidebar grid tetap terlihat; uji pada lebar 375 px HP dan 1280 px laptop sebagai acuan); ikon vektor (mis. Lucide) — bukan emoji; hindari animasi berlebihan, banner/ilustrasi besar di halaman ujian, warna latar mencolok, notifikasi yang mengganggu selama ujian (keindahan halaman ujian datang dari kejelasan, bukan dekorasi); mode gelap opsional, bukan prioritas v2 (jika ditambahkan, pastikan kontras status soal tetap terbaca di kedua mode).

#### Build & deploy statis: hasil build disajikan Go, routing SPA (fallback)

**Alur build yang direkomendasikan:** (1) frontend dibangun sekali menjadi file statis (perintah build Vite menghasilkan folder `dist/` berisi `index.html` + aset ber-hash — JS/CSS/media di dalam `assets/`); (2) backend Go menyajikan `dist/` tersebut — dua cara yang sama-sama valid: **embed ke dalam binary Go** (via `go:embed` — menghasilkan satu binary yang membawa frontend-nya; kelebihan: deploy satu file, tidak ada risiko file frontend tertinggal/tidak sinkron; kekurangan: setiap perubahan frontend mengharuskan rebuild binary Go) atau **sajikan dari folder di disk** (binary Go membaca folder `dist/` dari filesystem saat startup; kelebihan: frontend bisa diperbarui tanpa rebuild Go; kekurangan: harus memastikan folder `dist/` ikut ter-deploy dan versinya cocok dengan backend). **Rekomendasi untuk CBT SMK v2: embed ke binary** untuk kesederhanaan deploy di server sekolah (satu binary + database); bila alur rilis frontend menjadi sangat sering di masa depan, tinjau ulang ke mode folder. (3) Pipeline build yang direkomendasikan (urutan): install dependensi frontend → build frontend → salin `dist/` ke lokasi yang di-embed Go (bila memakai embed) → build binary Go → uji binary (buka halaman login, halaman ujian, refresh di rute SPA). (4) Jangan commit `dist/` ke repo (`dist/` adalah artefak build yang dihasilkan pipeline, bukan source; yang di-commit adalah source frontend + lockfile).

**Routing SPA (fallback) — aturan yang harus ditegakkan di Go:** request ke `/api/...` → **jangan pernah** fallback ke `index.html` (kembalikan JSON 404 bila endpoint tidak ada — kritis: bila API 404 dibalas `index.html`, frontend akan gagal parse JSON dengan error yang membingungkan); request GET ke file aset nyata (mis. `/assets/index-XXXX.js`, `/favicon.ico`) → sajikan file tersebut (bila tidak ada, kembalikan 404, bukan `index.html`); request GET ke rute SPA tanpa ekstensi file (mis. `/ujian`, `/dashboard`, `/hasil`) → sajikan `index.html` agar React Router mengambil alih; request non-GET ke rute SPA (mis. POST ke `/ujian`) → jangan fallback ke `index.html` (kembalikan 404/405 yang benar); request ke path yang mengandung ekstensi file tetapi tidak ditemukan → kembalikan 404, bukan `index.html` (heuristik "ada titik di path = kemungkinan file"). **Urutan pendaftaran route di Go:** daftarkan semua route API **terlebih dahulu**, baru kemudian handler statis + fallback SPA sebagai catch-all terakhir (bila urutannya terbalik, fallback SPA akan menelan request API).

**Cache headers — penting untuk update yang tidak "nyangkut":** aset ber-hash (`/assets/*` dengan hash di nama file): `Cache-Control: public, max-age=31536000, immutable` (aset ini tidak pernah berubah isinya — bila isinya berubah, namanya berubah; cache agresif aman dan mempercepat load berulang); `index.html`: `Cache-Control: no-cache` (atau `no-store` untuk keamanan ekstra — `index.html` adalah pintu masuk yang menunjuk ke aset ber-hash terbaru; bila `index.html` di-cache agresif, siswa bisa terus memuat versi lama meski sudah deploy versi baru — sumber klasik bug "sudah diperbaiki tapi masih error"); jangan men-cache respons API ujian dengan header statis yang sama (respons API harus memiliki kebijakan cache-nya sendiri, umumnya `no-store` untuk data ujian yang berubah).

**Konfigurasi & lingkungan:** base path — bila frontend disajikan dari root (`/`), tidak ada konfigurasi khusus; bila disajikan dari sub-path (mis. `/app/`), Vite dan React Router harus dikonfigurasi dengan base yang sama, dan Go harus melayani dengan prefix yang konsisten; untuk v2, rekomendasikan **root (`/`)** agar tidak menambah kompleksitas. Variabel lingkungan build: gunakan prefix Vite (`VITE_*`) untuk nilai yang memang boleh terekspos ke browser (mis. nama aplikasi); **jangan** menaruh secret, token, atau kredensial apa pun dalam variabel `VITE_*` — semua yang diawali `VITE_*` ikut terbungkus ke dalam bundle JS dan dapat dibaca siapa pun. Satu origin, tanpa CORS untuk API utama — karena Go menyajikan frontend dan API dari origin yang sama, frontend cukup memanggil path relatif (`/api/...`); ini menghilangkan seluruh kategori masalah CORS dan menyederhanakan deployment di server lokal sekolah. Dev vs produksi: selama development, jalankan Vite dev server terpisah (dengan proxy ke API Go) untuk HMR cepat; untuk produksi, selalu uji **build produksi yang disajikan Go** — jangan menganggap perilaku dev server sama dengan perilaku build produksi.

#### Komunikasi dengan backend: fetch / axios, pola API client, penanganan error / token

**fetch vs axios — rekomendasi: `fetch` bawaan + wrapper tipis sendiri, dipadukan dengan TanStack Query.** Perbandingan: ketergantungan (fetch: tidak ada; axios: satu dependency tambahan); parsing JSON (fetch: manual; axios: otomatis); error pada status HTTP error (fetch: tidak throw otomatis pada 4xx/5xx — harus diperiksa manual; axios: throw otomatis pada status di luar 2xx); interceptor (fetch: tidak ada bawaan — harus ditulis sebagai wrapper; axios: ada bawaan); timeout (fetch: via `AbortController`/`AbortSignal.timeout`; axios: opsi timeout bawaan); pembatalan (keduanya mendukung `AbortController`); upload progress (fetch: terbatas; axios: ada dukungan progress). Alasan: CBT ini tidak membutuhkan fitur axios yang paling membedakannya (interceptor dapat digantikan oleh satu wrapper `apiClient` yang melakukan hal yang sama dengan puluhan baris kode yang mudah dipahami); satu dependency lebih sedikit = satu permukaan serangan/pembaruan lebih sedikit; TanStack Query sudah menangani sebagian besar kebutuhan yang dulu membuat orang memilih axios (retry, caching, deduplikasi, pembatalan via signal). Pengecualian: bila tim sudah memiliki keahlian axios yang dalam dan sudah ada kode axios yang terbukti, memakai axios bukan kesalahan — tetapi untuk proyek baru v2, `fetch` + wrapper adalah pilihan yang lebih sederhana dan cukup. Library HTTP yang lebih baru (mis. `ky`, `tanstack-fetch`) adalah pembungkus `fetch` yang valid, tetapi untuk v2 mereka menambah lapisan yang belum diperlukan.

**Pola API client yang direkomendasikan:** satu titik masuk, bukan `fetch` berserakan di komponen — satu modul `apiClient` yang menjadi satu-satunya tempat komponen/hook memanggil HTTP (komponen tidak pernah memanggil `fetch` mentah ke URL bebas); satu modul `api` per domain di atas `apiClient` (mis. auth — login/logout/me; exam — mulai sesi, ambil soal, simpan jawaban, submit, heartbeat, pelanggaran; admin — jadwal, soal, rekap) — setiap fungsi mengembalikan data yang sudah diketik (TypeScript), bukan `Response` mentah; kontrak respons yang konsisten dari backend (setiap respons API Go mengikuti satu bentuk dengan kode status HTTP yang benar — frontend tidak boleh menebak-nebak bentuk error dari tiap endpoint); timeout yang eksplisit pada setiap request (tanpa timeout, request yang menggantung dapat membuat UI terkunci pada status "menyimpan…" selamanya — nilai timeout dibedakan: request biasa mis. 10 detik, simpan jawaban mis. 15 detik, submit akhir lebih panjang karena mem-flush antrean); pembatalan dengan `AbortSignal` (saat komponen unmount atau soal berganti, batalkan request yang sudah tidak relevan agar tidak terjadi race — TanStack Query sudah meneruskan signal ini ke query function); pemisahan TanStack Query vs panggilan langsung (data yang dibaca — soal, jadwal, hasil — lewat TanStack Query; aksi yang mengubah data — simpan jawaban, submit, lapor pelanggaran — lewat mutation/panggilan langsung `apiClient`, **kecuali** simpan jawaban ujian yang memakai antrean offline sendiri, yang tidak boleh diserahkan sepenuhnya ke cache query karena kebutuhan retry/antrean/persistensi lokalnya khusus).

**Penanganan token & autentikasi — rekomendasi utama: cookie `HttpOnly` + `SameSite` yang diatur server Go.** Perbandingan penyimpanan token: cookie HttpOnly (server-set) — token tidak dapat dibaca JavaScript → kebal terhadap pencurian token via XSS; browser mengirim otomatis; cocok dengan satu origin (tanpa CORS); membutuhkan perlindungan CSRF; **direkomendasikan untuk v2**. Token di `localStorage`/`sessionStorage` — mudah dibaca JS, mudah di-debug, tetapi rentan dicuri via XSS; **hindari** untuk token akses, kecuali ada alasan arsitektur yang memaksa (dan bila terpaksa, harus dibarengi sanitasi XSS yang ketat). Token di memori JS saja — paling aman dari pencurian persisten, tetapi hilang saat refresh (siswa harus login ulang tiap refresh — tidak dapat diterima selama ujian); tidak cocok sebagai satu-satunya penyimpanan untuk CBT. Aturan penanganan autentikasi: satu tempat menyisipkan kredensial (`apiClient` menyisipkan kredensial secara terpusat — cookie dikirim otomatis oleh browser bila `credentials: include` diatur, atau header Authorization bila memakai token yang dipegang JS — pilih satu, jangan campur); penanganan 401 terpusat, dengan single-flight refresh (bila memakai refresh token: bila banyak request bersamaan mendapat 401, hanya **satu** refresh yang boleh berjalan; request lain menunggu hasil refresh tersebut, lalu diulang sekali; tanpa single-flight, lima request yang gagal bersamaan akan memicu lima refresh yang saling membatalkan); jangan retry tanpa batas pada 401 (satu retry setelah refresh; bila masih 401, anggap sesi berakhir → arahkan ke halaman login dengan pesan yang jelas "Sesi berakhir, silakan login kembali" dan **selamatkan jawaban yang belum terkirim ke penyimpanan lokal terlebih dahulu** sebelum mengarahkan pergi); CSRF bila memakai cookie (token CSRF / pola double-submit cookie atau validasi header `Origin`/`Referer` di server Go, ditambah `SameSite=Lax` atau `Strict` pada cookie — tanggung jawab backend, tetapi frontend harus meneruskan token CSRF/header yang disyaratkan pada setiap request mutasi); jangan menyimpan token di URL, log, atau error report (token tidak boleh muncul di query string — akan tersimpan di log server/riwayat browser — atau di pesan error yang dikirim ke layanan logging); logout yang bersih (memanggil endpoint logout server yang menghapus sesi/cookie di server **dan** membersihkan state lokal — store ujian, antrean yang sudah terkonfirmasi, cache query; jangan hanya menghapus di sisi klien); peran (role) ditegakkan di server, bukan di frontend (frontend boleh menyembunyikan menu admin dari siswa untuk kerapian UI, tetapi **tidak boleh** menjadi satu-satunya penjaga — setiap endpoint admin harus memvalidasi peran di server; menyembunyikan tombol bukan keamanan).

**Penanganan error — taksonomi yang direkomendasikan** (frontend harus membedakan jenis error karena respons UI-nya berbeda): jaringan/offline (`TypeError: Failed to fetch`, timeout, `navigator.onLine === false`) → indikator offline + antrean jawaban (jangan tampilkan "terjadi kesalahan" yang menakutkan; tampilkan "Offline — jawaban aman di perangkat"); 401 Unauthorized (sesi berakhir/token tidak valid) → selamatkan draf lokal → arahkan ke login dengan pesan sesi berakhir; 403 Forbidden (siswa mencoba akses endpoint admin/ujian milik orang lain) → pesan "Anda tidak memiliki akses" + kembali ke dashboard (jangan tampilkan detail internal); 404 Not Found (sesi/soal tidak ditemukan) → pesan "Sesi ujian tidak ditemukan" + kembali ke dashboard; 409 Conflict (sesi sudah disubmit/sudah diblokir/jawaban versi lama) → tampilkan status sesi yang benar dari server (mis. "Ujian ini sudah dikumpulkan") — jangan biarkan siswa terus mengerjakan sesi yang sudah ditutup; 422/400 Validation (jawaban tidak valid) → pesan validasi yang spesifik per field, bukan "error" generik; 429 Too Many Requests → tunggu + retry dengan backoff (jangan spam; tampilkan "Terlalu banyak permintaan, mencoba lagi…"); 5xx Server error → pesan "Server bermasalah, jawaban Anda aman di perangkat, mencoba lagi…" + retry antrean (jangan menghapus antrean lokal). Aturan presentasi error: pesan untuk siswa harus dalam Bahasa Indonesia yang sederhana dan memberi tindakan berikutnya ("Coba lagi", "Periksa koneksi", "Hubungi proktor"), bukan pesan teknis mentah ("500 Internal Server Error", stack trace); jangan tampilkan detail internal (pesan error database, path file, stack trace) kepada siswa (detail itu masuk log, bukan ke layar siswa); satu tempat untuk menampilkan error global (sistem toast/banner yang konsisten, bukan `alert()` browser yang memblokir dan jelek); **error pada halaman ujian tidak boleh menghapus jawaban** — aturan keras: tidak ada jalur error apa pun yang boleh membersihkan jawaban lokal sebelum terkonfirmasi tersimpan di server.

**Ringkasan rekomendasi Researcher React Frontend:** Tooling **Vite 8 + React 19 + TypeScript + React Router**; state ujian **Zustand** (dengan selector); server state **TanStack Query**; state jarang berubah (auth, tema) **Context**; state lokal komponen **`useState`/`useReducer`**; timer — deadline absolut dari server + offset + re-sinkronisasi, keputusan waktu habis divalidasi server; simpan jawaban — simpan lokal seketika + debounce kirim + antrean offline + retry + indikator status jujur, server menang saat konflik; anti-cheat frontend — `visibilitychange` + `blur` + `fullscreenchange` + blokir copy-paste/klik kanan + heartbeat + pelanggaran dihitung server, **tidak pernah** diklaim 100%; UI — **Tailwind CSS 4 + shadcn/ui selektif + komponen ujian custom**, MUI/Ant Design ditolak untuk v2; build/deploy — Vite `dist/` disajikan Go (embed direkomendasikan), fallback SPA hanya untuk GET rute tanpa ekstensi, **tidak pernah** untuk `/api/*`, `index.html` no-cache, aset ber-hash immutable; HTTP client — **`fetch` + wrapper `apiClient` sendiri + TanStack Query**, axios tidak diperlukan untuk proyek baru ini; auth — **cookie HttpOnly + CSRF protection** (server Go), hindari token di `localStorage`; prinsip lintas semua area — **server adalah sumber kebenaran** (frontend menampilkan, mengantrekan, memulihkan — tidak pernah memutuskan sendiri soal deadline, nilai, blokir, atau peran).

**Risiko & hal yang harus diputuskan di luar frontend:** CSRF + keamanan cookie harus dirancang di backend Go bersamaan dengan frontend (bukan ditunda); kebijakan pelanggaran (berapa pelanggaran = peringatan vs auto-block) harus ditetapkan sekolah/proktor per jenis ujian, bukan di-hardcode oleh frontend; jalur banding untuk siswa yang merasa terblokir secara tidak adil harus ada sebelum anti-cheat diaktifkan pada ujian resmi; uji di perangkat nyata siswa (HP Android spek rendah, jaringan lokal sekolah yang lambat) harus menjadi bagian dari definisi selesai — bukan hanya uji di laptop pengembang; rencana offline total (bila server lokal mati di tengah ujian, apa yang terjadi pada jawaban di perangkat dan bagaimana pemulihannya) harus menjadi prosedur tertulis proktor, bukan hanya fitur frontend.

### 14.6 Modul & urutan

M1 Auth + RBAC → M2 Master Data → M3 Bank Soal → M4 Paket & Jadwal → M5 Ujian Peserta → M6 Koreksi & Nilai → M7 Laporan & Analisis → M8 UKK. MX lintas modul (backup/restore, monitor real-time, upload, notifikasi, keamanan) bertahap.

**Kurikulum Merdeka:** jurusan → konsentrasi keahlian → mapel → elemen/CP; setiap soal menunjuk ke satu kompetensi.

---

## 15. Desain API (Ringkas)

### 15.1 Konvensi

- Base path **`/api/v1`**. Frontend React memanggil `/api/v1/...` (route HTML v1 tetap jalan selama transisi, lalu dinonaktifkan). Breaking change → `/api/v2` + header `Deprecation`/`Sunset`.
- Content-Type `application/json; charset=utf-8` (kecuali download/upload multipart). ID UUID string. Waktu ISO 8601 UTC `Z`, tampilan WIB di frontend. Boolean true/false. Field tidak dikenal diabaikan (kecuali dilarang untuk role → 403). `null` = tidak ada/belum diisi; field tidak dikirim = tidak diubah (PUT/PATCH).
- Respons sukses: `{success, message Bahasa Indonesia, data, meta, request_id}`. Error: `{success:false, message, error:{code, details[]}, request_id}`.
- Pagination `?page=1&per_page=20&sort=&order=asc&q=&status=` (default page 1, per_page 20, maks 100, >100 dipotong bukan error). List `data` selalu array (bukan null).
- HTTP: 200/201/200+body untuk DELETE (dipilih konsisten); 400 `BAD_REQUEST`; 401 `UNAUTHENTICATED`; 403 `FORBIDDEN`; 404 `NOT_FOUND` (juga dipakai saat resource ada tapi pemanggil tidak berhak — agar tidak membocorkan keberadaan); 409 `CONFLICT`; 422 `VALIDATION_ERROR`; 429 `RATE_LIMITED`; 500 `INTERNAL_ERROR`. Pesan error tidak membocorkan stack trace/SQL/nama tabel.
- **Idempotensi:** simpan jawaban upsert per `(siswa_ujian_id, soal_id)`; submit idempotent dengan `Idempotency-Key` (submit kedua dengan key sama → hasil sama; tanpa key pada ujian sudah selesai → 409); optimistic locking pada edit bank soal/soal/jadwal via `updated_at` (berubah → 409).
- **Rate limit:** login maks 5 gagal per akun per 15 menit → ke-6 → 429 + kunci 15 menit; simpan jawaban maks 60 req/menit per peserta; semua mutasi dari browser wajib header CSRF; upload maks 10 MB per file.

### 15.2 Daftar endpoint (per modul)

| Modul | Endpoint (ringkas) |
|---|---|
| **Auth** | login staf, login peserta, refresh, logout, ganti password, lupa password (prosedur sekolah), profil saya |
| **Master Data** | jurusan, kelas/rombel, mapel, users, peserta, group, kompetensi, agama |
| **Bank Soal** | CRUD bank + lock/unlock + duplicate + import + export + media + analisis |
| **Paket Ujian** | CRUD paket + publish/unpublish |
| **Jadwal** | CRUD jadwal + pause/resume/finish/clone + peserta + reset-ujian + force-finish + reset-login |
| **Token** | list / generate / revoke / validate |
| **Ujian Siswa** | jadwal saya / mulai / soal / status / jawaban / pelanggaran / submit / hasil |
| **Koreksi** | list / detail / beri nilai / finalisasi / ringkasan |
| **Rekap/Laporan** | rekap / detail / kompetensi / analisis-butir / export / berita-acara / ledger |
| **Monitor** | REST snapshot + force-finish/unblock/reset-login + WS + ws-ticket |
| **UKK** | paket / peserta / nilai-praktik / upload / sertifikat / rekap |
| **Event & Ruangan** | events / ruangan / anggota / kartu-peserta |
| **Settings** | list / ubah per kode — termasuk `nama_sekolah`, `alamat_sekolah`, `kota`, `kop_sekolah`, `max_counter_out` (default 1), `login_once_atime` (Y/N), `auto_block` (Y/N), `token_expiry_detik` (default 900), `tampil_hasil_ke_siswa` (Y/N) |
| **Backup** | list / buat / download / restore / hapus |
| **Media** | upload / hapus |

**Role per endpoint:** EVIL semua; EVAL bank soal/paket/jadwal miliknya + koreksi + rekap + analisis; OPRA peserta/jadwal/token/monitor/event/ruangan/reset; PENGAWAS monitor (baca + lapor pelanggaran) + daftar hadir + berita acara, tidak ubah soal/jadwal/nilai; SISWA hanya data `peserta_id`-nya sendiri (akses data peserta lain → 403/404); UKK hanya modul UKK.

### 15.3 Kontrak 5 endpoint terpenting (ringkas, bukan kode)

1. **Login staf:** email + password → JWT access + refresh cookie + CSRF token. Login pertama dengan password default → 403 `PASSWORD_CHANGE_REQUIRED` + diarahkan ganti password.
2. **Ambil soal ujian:** **TIDAK BOLEH** berisi kunci jawaban, poin benar/salah, atau pembahasan. Bentuk `jawaban_saya` tergantung tipe: string untuk PG/benar-salah/isian/esai; array string untuk PG kompleks/menjodohkan/mengurutkan; objek untuk pasangan.
3. **Simpan jawaban:** upsert per `(siswa_ujian_id, soal_id)`; server hitung nilai sendiri; nilai kiriman browser diabaikan.
4. **Submit ujian:** `konfirmasi:true` wajib (tanpa itu → 422). Nilai yang dikembalikan adalah nilai otomatis sementara; esai menunggu koreksi. Bila `tampil_hasil_ke_siswa=N` → nilai disembunyikan.
5. **Rekap:** termasuk nilai per kompetensi, `out_counter`, diblokir, alasan blokir.

### 15.4 Auth

**JWT access token** (header `Authorization: Bearer`, 15 menit, disimpan di memori React, BUKAN localStorage) + **refresh token** (cookie `HttpOnly; Secure; SameSite=Lax; Path=/api/v1/auth` bernama `cbt_refresh`, 8 jam staf / 4 jam siswa, rotasi — refresh yang sudah dipakai tidak bisa dipakai lagi) + **CSRF token** (header `X-CSRF-Token` pada semua mutasi, cookie `cbt_csrf` bisa dibaca JS). Logout hapus refresh cookie + cabut sesi. Tanpa access token → 401; CSRF tidak cocok → 403 `CSRF_MISMATCH`. Password tetap bcrypt.

**Header tambahan ujian:** `X-Device-Id` (UUID per browser/perangkat, localStorage, untuk device lock bila `login_once_atime=Y`), `Idempotency-Key` opsional, `X-Request-Id` opsional.

**WebSocket auth via ticket satu-kali:** `POST /api/v1/monitor/ws-ticket` (Bearer) → ticket 30 detik → `WS /api/v1/ws/monitor?jadwal_id=...&ticket=...`; ticket dipakai/kedaluwarsa → 4401.

### 15.5 Validasi (ringkas)

Validasi global + per modul, termasuk: validasi per tipe soal 1–16; token validate → 401 `TOKEN_INVALID` / 409 `TOKEN_EXPIRED`; pelanggaran debounce 5 detik agar tidak double-count; koreksi poin manual 0 s.d. bobot; finalisasi ditolak bila masih ada jawaban manual belum dikoreksi + daftar `belum_dikoreksi`; restore ditolak bila ada ujian berjalan → 409 `UJIAN_BERJALAN` + `konfirmasi:true` wajib.

### 15.6 Pemetaan v1 → v2

**Dipertahankan:** alur ujian, token, pause/resume, koreksi manual, rekap CSV. **Baru:** edit master data, CSRF, fullscreen (deteksi), WebSocket, backup/restore, event+ruangan+kartu, analisis butir, nilai per kompetensi, UKK. **Ditunda:** import Word/Google Docs (Word di v2, Google Docs pasca-v2), AI generate soal.

**Open questions API (Bab 24):** setuju pindah session cookie → JWT + refresh? Nilai esai sebelum finalisasi — siswa boleh lihat nilai otomatis sementara langsung setelah submit, atau disembunyikan total sampai koreksi final? (desain memakai setting `tampil_hasil_ke_siswa`, default sembunyikan sampai final). Export XLSX perlu atau cukup CSV tahap 1? Upload UKK — siswa boleh upload berkas praktik sendiri, atau hanya penguji? WebSocket ticket 30 detik, atau access token lewat query param WS?

---

## 16. ERD (Ringkas)

### 16.1 Daftar tabel (40 tabel, 6 kelompok — DB Dev standby) + detail Researcher Database & Arsitektur

| Kelompok | Tabel |
|---|---|
| **A. Akun/role/izin (4)** | pengguna, peran, pengguna_peran, izin_peran |
| **B. Master sekolah (8)** | tahun_ajaran, jurusan, kelas, mata_pelajaran, agama, siswa, guru, kompetensi |
| **C. Bank soal (5)** | bank_soal, soal, opsi_jawaban, media_soal, soal_kompetensi |
| **D. Paket/jadwal/event/ruangan/token (8)** | paket_ujian, paket_bank_soal, jadwal_ujian, jadwal_peserta, event_ujian, ruangan, ruangan_anggota, token_ujian |
| **E. Pelaksanaan/jawaban/nilai/UKK (9)** | sesi_ujian_peserta, jawaban_peserta, hasil_ujian, koreksi_manual, nilai_kompetensi, analisis_soal, analisis_bank_soal, ukk_penilaian, ukk_dokumen |
| **F. Perangkat/log/sistem (5)** | perangkat_peserta, log_aktivitas, log_pelanggaran, pengaturan, riwayat_cadangan |
| **Terpisah** | kartu_peserta (dihitung terpisah) |

**Detail dari Researcher Database & Arsitektur** (laporan lengkap 987 baris di `~/workspace/cbt-smk-v2/laporan-database-arsitektur.md`):

**Konvensi umum hampir semua tabel:** setiap baris memiliki pengenal unik internal; setiap baris mencatat kapan dibuat dan kapan terakhir diubah; baris yang penting secara audit (ujian, jawaban, nilai, log) tidak dihapus fisik — penghapusan dilakukan dengan penandaan arsip/nonaktif (soft delete) agar jejak tetap ada; kolom status memakai daftar nilai terkontrol (aktif/nonaktif/arsip/draf/terbit/berjalan/selesai/diblokir), bukan teks bebas.

**1.1 Pengguna, peran, kelembagaan.** Peran: nama peran (Super Admin, Guru, Proktor, Siswa, Penguji UKK), kode peran yang stabil dipakai aplikasi (tidak berubah walau nama tampilan diubah), deskripsi, status aktif/nonaktif. Pengguna (Users): satu orang = satu akun; nama lengkap; nama pengguna untuk login (unik); email (unik bila diisi); kata sandi dalam bentuk hash + penanda algoritma hash (kata sandi asli tidak pernah disimpan); status akun (aktif, nonaktif, terkunci, perlu ganti kata sandi); waktu login terakhir yang berhasil; waktu perubahan kata sandi terakhir; penanda akun demo/contoh (agar mudah dibersihkan); penanda arsip (soft delete). Data identitas sekolah (NIS, NIP, kelas) disimpan di tabel profil terkait, bukan ditumpuk semua di tabel ini. Relasi Pengguna–Peran: many-to-many (satu pengguna bisa memegang lebih dari satu peran, mis. guru yang juga menjadi proktor) — kolom: rujukan pengguna, rujukan peran, kapan peran diberikan dan oleh siapa, status penugasan (aktif/dicabut). Izin per Peran (opsional tapi dianjurkan): daftar izin halus per aksi (boleh membuat soal, mengoreksi esai, memulai/menjeda jadwal, melihat rekap) — warisan CBT-Go/Exo memiliki 100+ izin per aksi; v2 tidak harus meniru jumlahnya, tapi mekanismenya perlu ada agar peran tidak dikeraskan di kode. Jurusan: kode jurusan (unik, singkat, stabil), nama, deskripsi, status. Kelas: rujukan jurusan, nama kelas (mis. XI RPL 1), tingkat (X/XI/XII), tahun ajaran/angkatan, wali kelas (rujukan guru, boleh kosong), kapasitas, status. Profil Siswa: rujukan pengguna (akun login siswa) — satu pengguna dengan peran siswa memiliki tepat satu profil siswa; NIS/NISN (unik); rujukan kelas saat ini; rujukan jurusan (boleh diturunkan dari kelas, tapi disimpan untuk memudahkan rekap lintas kelas); jenis kelamin, tanggal lahir (hanya bila diperlukan sekolah; bila tidak, jangan dikumpulkan); nomor peserta ujian/nomor kartu (unik per ujian bila dipakai); status siswa (aktif, pindah, lulus, nonaktif); riwayat kelas bila siswa naik kelas (disimpan sebagai catatan, bukan menimpa kelas lama pada data ujian yang sudah selesai). Profil Guru/Staf (opsional): rujukan pengguna, NIP (bila ada), mapel yang diampu, status. Mapel: kode mapel (unik), nama, kelompok mapel (produktif, normatif, adaptif — sesuai kebutuhan sekolah), deskripsi, status. Penugasan Guru–Mapel: many-to-many (satu guru bisa mengampu banyak mapel, satu mapel bisa diampu banyak guru) — rujukan guru, rujukan mapel, tahun ajaran berlaku, status penugasan.

**1.2 Bank soal.** Bank Soal: nama bank, rujukan mapel, rujukan jurusan/tingkat sasaran (boleh kosong bila umum), pemilik/pembuat (rujukan guru), deskripsi/tujuan, status (draf, aktif, arsip), penanda terkunci (bank yang terkunci tidak boleh diubah tanpa membuka kunci oleh peran yang berwenang). Soal: rujukan bank soal; tipe soal (daftar terkontrol); teks/isi soal (mendukung format teks kaya/HTML aman + rujukan ke media); bobot/skor maksimum; tingkat kesulitan (mudah, sedang, sulit); kompetensi/indikator yang diukur (teks atau rujukan ke daftar kompetensi); kunci jawaban/pedoman penskoran (objektif: kunci terstruktur; esai: rubrik/panduan koreksi teks); pembahasan/umpan balik (ditampilkan setelah ujian bila diizinkan); media pendukung (rujukan ke file); status soal (draf, aktif, nonaktif, arsip); versi soal (setiap perubahan bermakna pada soal yang sudah dipakai ujian sebaiknya membuat versi baru, bukan menimpa); pembuat dan penyunting terakhir; penanda terkunci; jumlah pemakaian (boleh dihitung, tapi disimpan sebagai cache). **Catatan desain penting:** jangan menyimpan jawaban peserta di tabel soal; jangan menyimpan skor peserta di tabel soal — tabel soal hanya menyimpan soal + kunci + pedoman. Opsi Jawaban: rujukan soal induk; label opsi (A, B, C, D); teks/isi opsi (mendukung media bila perlu); penanda apakah opsi ini benar (PG: tepat satu benar; PG kompleks/benar-salah: bisa banyak); urutan tampil opsi; penanda apakah opsi diacak saat ujian (bila pengacakan per peserta diaktifkan, pengacakan dicatat di sisi pelaksanaan, bukan dengan mengubah tabel ini); status aktif/nonaktif. Pasangan Menjodohkan (bila diperlukan sebagai tabel sendiri): rujukan soal induk; pernyataan sisi kiri; pasangan benar sisi kanan; urutan tampil. Alternatif: pasangan menjodohkan dapat disimpan sebagai struktur terstruktur di dalam kolom kunci jawaban soal, dengan dokumentasi format yang ketat — keputusan ini harus dicatat dan tidak diubah di tengah jalan. Kompetensi/Indikator (opsional tahap awal, dianjurkan untuk nilai per kompetensi): kode kompetensi, nama kompetensi/indikator, rujukan mapel, deskripsi, status; relasi many-to-many via tabel penghubung Soal–Kompetensi (bila v2 awal hanya mendukung satu kompetensi per soal, cukup simpan satu rujukan kompetensi di tabel soal dan tunda tabel penghubung sampai dibutuhkan). File/Media Soal: metadata file pendukung (gambar, audio, video, dokumen) — nama file asli + nama file tersimpan, jenis file/tipe media, ukuran file, rujukan ke soal/bank soal yang memakai file, pengunggah dan waktu unggah, status (aktif/dihapus). File fisik disimpan di penyimpanan file (disk lokal server), bukan sebagai biner di database.

**1.3 Paket ujian, jadwal, sesi.** Paket Ujian: nama paket; rujukan mapel; jenis ujian (STS/SAS/USBK/UKK/Tryout/lainnya); deskripsi/petunjuk ujian yang ditampilkan ke peserta; durasi ujian dalam menit; jumlah soal yang ditampilkan (bila paket mengambil subset acak dari bank; bila paket memakai susunan tetap, jumlah ini diturunkan dari susunan); aturan pengacakan (acak urutan soal ya/tidak, acak urutan opsi ya/tidak); nilai minimum kelulusan (KKM/passing grade) bila dipakai; aturan anti-curang (batas pindah tab/keluar jendela sebelum peringatan, batas sebelum blokir otomatis, apakah fullscreen diwajibkan, apakah perangkat dicatat); aturan tampilan hasil (apakah peserta boleh melihat pembahasan, kapan nilai ditampilkan); status paket (draf, aktif, nonaktif, arsip); pembuat paket. Susunan Soal Paket: penghubung many-to-many antara paket ujian dan soal — rujukan paket, rujukan soal, urutan soal dalam paket, bobot soal dalam paket ini (boleh berbeda dari bobot default di tabel soal; bobot di sini yang dipakai untuk penilaian paket tersebut), status item (aktif/dikeluarkan dari paket). **Penting:** jika soal dihapus dari paket setelah ujian berjalan, baris tidak dihapus fisik — ditandai nonaktif agar hasil ujian lama tetap dapat dihitung ulang. Jadwal Ujian: rujukan paket ujian; nama/label jadwal (mis. STS Pemrograman Web — XI RPL 1 — Gelombang 1); waktu mulai dan waktu selesai yang direncanakan; durasi efektif (boleh sama dengan durasi paket, atau dioverride per jadwal); token ujian (kode yang dimasukkan peserta untuk masuk; unik per jadwal pada masa berlaku); status jadwal (draf, dijadwalkan, berjalan, dijeda/pause, selesai, dibatalkan); alasan/catatan jeda; kelas/rombongan yang diundang (via tabel peserta jadwal); ruang/lab pelaksanaan (teks atau rujukan ke tabel ruangan bila ada); pengawas yang ditugaskan (rujukan pengguna); pengaturan override per jadwal (mis. durasi diperpanjang, token diubah) — dicatat agar audit jelas; pembuat jadwal. Peserta Jadwal: penghubung many-to-many antara jadwal dan siswa — rujukan jadwal, rujukan siswa, nomor peserta/nomor kartu (bila dipakai), status keikutsertaan (diundang, hadir, tidak hadir, sedang mengerjakan, selesai, diblokir, dikeluarkan), waktu mulai mengerjakan (aktual), waktu selesai mengerjakan (aktual), sisa waktu/status timer pada saat terakhir tersimpan, alasan blokir/catatan proktor bila ada. Sesi Ujian Peserta (Attempts): satu percobaan mengerjakan oleh satu peserta pada satu jadwal — umumnya satu peserta = satu sesi per jadwal, tetapi desain harus mendukung percobaan ulang yang sah (sesi dibatalkan karena gangguan dan diulang atas izin proktor) dengan mencatat sesi mana yang sah/dipakai untuk nilai. Kolom: rujukan peserta jadwal (atau langsung ke jadwal + siswa); nomor percobaan (ke-1, ke-2, dst); waktu mulai dan waktu selesai sesi; status sesi (belum mulai, berjalan, dijeda, selesai, dibatalkan, diblokir); perangkat yang dipakai (rujukan tabel perangkat/catatan perangkat); alamat jaringan lokal/IP saat sesi (audit); **susunan soal yang diterima peserta (urutan soal setelah pengacakan, dan urutan opsi bila diacak) — disimpan sebagai daftar terstruktur per sesi agar hasil dapat direkonstruksi walau paket diubah belakangan**; penanda apakah sesi ini yang dipakai untuk nilai akhir (bila ada percobaan ulang); catatan proktor. **Mengapa susunan soal per sesi disimpan:** tanpa ini, bila paket diubah setelah ujian, rekonstruksi "soal apa yang dilihat peserta" menjadi mustahil — penting untuk sengketa nilai dan audit. Token Ujian: dapat disimpan sebagai kolom di tabel jadwal (paling sederhana) atau sebagai tabel sendiri bila token perlu masa berlaku sendiri, token per kelas, atau riwayat pergantian token. Untuk v2, rekomendasi: token sebagai kolom jadwal + tabel riwayat perubahan token (kapan token diubah, oleh siapa; token lama tidak disimpan dalam bentuk yang bisa dibaca ulang bila tidak perlu). Bila sekolah butuh token berbeda per kelas dalam satu jadwal, buat tabel token per jadwal–kelas. Kolom bila dijadikan tabel: rujukan jadwal, rujukan kelas (bila token per kelas), nilai token (**disimpan dalam bentuk hash bila memungkinkan; token asli hanya ditampilkan sekali saat dibuat**), masa berlaku token (mulai–selesai), status token (aktif, kedaluwarsa, dicabut).

**1.4 Jawaban, nilai, koreksi.** Jawaban Peserta: satu baris = jawaban satu peserta untuk satu soal dalam satu sesi. Ini tabel tersibuk kedua setelah log. Setiap penyimpanan (autosave) memperbarui baris yang sama, bukan menambah baris baru setiap ketikan — kecuali bila riwayat jawaban diperlukan. Kolom: rujukan sesi ujian; rujukan peserta/siswa (boleh diturunkan dari sesi, tapi disimpan untuk mempercepat rekap); rujukan soal; rujukan paket/jadwal (untuk mempercepat rekap tanpa join berantai); isi jawaban peserta dalam bentuk terstruktur sesuai tipe soal (PG: opsi yang dipilih; PG kompleks: daftar opsi yang dipilih; benar-salah: pilihan benar/salah per pernyataan; isian singkat: teks jawaban; menjodohkan: pasangan yang dipilih; esai: teks jawaban); penanda ragu-ragu per soal; penanda sudah dijawab/masih kosong; skor otomatis (hasil penskoran mesin; kosong untuk esai sebelum dikoreksi); skor manual/koreksi (diisi guru untuk esai/isian yang perlu penilaian manusia); skor akhir per soal (gabungan otomatis + manual, sesuai aturan paket); waktu jawaban pertama kali disimpan dan waktu terakhir diubah; status jawaban (tersimpan, diubah, dikoreksi, final). **Catatan riwayat jawaban:** bila sekolah ingin audit "jawaban awal vs jawaban akhir" (misal: mendeteksi perubahan mencurigakan), tambahkan tabel riwayat jawaban terpisah yang hanya menambah baris saat jawaban berubah secara signifikan — jangan membebani tabel jawaban utama dengan riwayat per ketikan. Hasil/Nilai Ujian: rekap per peserta per jadwal/per sesi — rujukan sesi ujian (dan/atau peserta jadwal); rujukan siswa, jadwal, paket; skor total yang diperoleh; skor maksimum yang mungkin; nilai akhir dalam skala yang dipakai sekolah (mis. 0–100); jumlah soal (total, dijawab, benar, salah, kosong, ragu-ragu); skor komponen objektif (hasil mesin); skor komponen esai/manual (hasil koreksi manusia); status koreksi (belum dikoreksi, sebagian, selesai); status kelulusan (lulus/tidak lulus/belum ditentukan); status hasil (draf, final, diterbitkan — nilai yang sudah diterbitkan tidak boleh diubah tanpa membuka kunci/mencatat alasan); waktu penghitungan terakhir; penghitung terakhir (sistem/pengguna). Tabel ini adalah hasil agregasi dari tabel jawaban, bukan pengganti tabel jawaban — nilai dihitung ulang dari jawaban bila perlu, dan hasil perhitungan disimpan di sini agar rekap cepat. Nilai per Kompetensi (opsional, dianjurkan): rujukan hasil ujian/peserta/jadwal; rujukan kompetensi; skor yang diperoleh untuk kompetensi tersebut; skor maksimum untuk kompetensi tersebut; nilai dalam skala yang sama dengan nilai akhir. Koreksi Manual/Penilaian Esai: jejak koreksi manusia — rujukan jawaban peserta; rujukan pengoreksi (guru); skor yang diberikan; komentar/umpan balik koreksi; waktu koreksi; status koreksi (draf, final). Memisahkan koreksi dari jawaban memudahkan audit. Analisis Butir Soal (hasil perhitungan, bukan input): hasil perhitungan berkala per soal per jadwal/per paket — rujukan soal dan jadwal/paket tempat perhitungan dilakukan; jumlah peserta yang menjawab, jumlah benar; tingkat kesulitan empiris (proporsi benar); daya beda (selisih performa kelompok atas vs bawah — definisi persis dicatat di dokumentasi); distribusi pilihan jawaban (berapa peserta memilih tiap opsi) — untuk menilai pengecoh; waktu perhitungan; catatan/rekomendasi (mis. soal terlalu mudah, pengecoh tidak berfungsi). Disimpan sebagai hasil perhitungan agar laporan cepat dan dapat dibandingkan antar periode.

**1.5 Keamanan, perangkat, audit.** Perangkat Peserta: catatan perangkat yang dipakai peserta saat ujian — rujukan pengguna/siswa; rujukan sesi ujian; sidik perangkat (fingerprint: gabungan info perangkat yang tidak sensitif — jenis perangkat, sistem operasi, peramban — bukan data pribadi); alamat jaringan lokal/IP; waktu pertama terlihat dan waktu terakhir terlihat; status (diizinkan, mencurigakan, diblokir). Peristiwa Anti-Curang (Proctoring Events): satu baris = satu peristiwa — rujukan sesi ujian/peserta; jenis peristiwa (pindah tab, keluar fullscreen, buka jendela baru, dll); waktu peristiwa; tindakan yang diambil sistem (peringatan, penguncian sementara, blokir otomatis); jumlah pelanggaran kumulatif pada sesi tersebut; catatan proktor/keputusan (peringatan diabaikan, blokir dicabut, dll). Log Aktivitas: jejak audit semua aksi penting (login, pembuatan soal, perubahan jadwal, koreksi nilai, ekspor data, perubahan token, dll) — waktu peristiwa (dengan zona waktu yang jelas); rujukan pengguna pelaku (bila ada); jenis aksi (login, buat soal, ubah soal, mulai ujian, submit, koreksi, ekspor, dll); objek yang terkena (jenis objek + pengenal objek); hasil aksi (berhasil/gagal + alasan singkat bila gagal); alamat jaringan/perangkat pelaku; detail tambahan terstruktur (mis. nilai lama vs nilai baru untuk perubahan penting); tingkat kepentingan (info, peringatan, kritis). **Prinsip:** log hanya ditambah (append-only) — tidak ada fitur ubah/hapus log dari aplikasi; penghapusan log hanya via kebijakan retensi yang terdokumentasi. Pengaturan Sistem (Settings): kunci pengaturan (nama unik, stabil), nilai pengaturan, kelompok pengaturan, deskripsi/batas nilai yang diizinkan, kapan terakhir diubah dan oleh siapa. Ruangan/Event (opsional): bila sekolah ingin mengelola ruangan/lab dan event ujian (fitur Exo), tambahkan tabel ruangan dan event. Untuk v2 awal, ruangan cukup sebagai teks di jadwal; jadikan tabel bila ruangan perlu dijadwalkan (menghindari bentrok ruangan). Kolom bila dijadikan tabel: nama ruangan/lab, kapasitas, lokasi/gedung, status (aktif/perawatan/nonaktif).

### 16.2 Aturan integritas + relasi

**Peta relasi (dalam kata-kata):** Pengguna & peran — satu pengguna dapat memiliki banyak peran (many-to-many) melalui Relasi Pengguna–Peran; satu peran dapat dimiliki banyak pengguna; satu peran dapat memiliki banyak izin (one-to-many). Kelembagaan — satu jurusan memiliki banyak kelas (one-to-many); satu kelas memiliki banyak siswa (one-to-many); satu guru dapat mengampu banyak mapel, satu mapel dapat diampu banyak guru (many-to-many) melalui Penugasan Guru–Mapel; satu siswa memiliki tepat satu profil siswa dan satu akun pengguna (one-to-one). Bank soal — satu bank soal milik satu mapel (many-to-one); satu bank soal memiliki banyak soal (one-to-many); satu soal memiliki banyak opsi jawaban (one-to-many), kecuali soal esai/isian yang tidak memiliki opsi; satu soal dapat mengukur banyak kompetensi (many-to-many) bila fitur nilai per kompetensi dipakai; satu soal dapat memiliki banyak file media (one-to-many). Paket & jadwal — satu paket ujian milik satu mapel (many-to-one); satu paket berisi banyak soal (many-to-many) melalui Susunan Soal Paket; satu soal dapat masuk banyak paket (many-to-many); satu paket dapat dijadwalkan banyak kali (one-to-many); satu jadwal mengundang banyak siswa (many-to-many) melalui Peserta Jadwal; satu siswa dapat mengikuti banyak jadwal (many-to-many); satu peserta jadwal dapat memiliki satu atau lebih sesi (one-to-many), dengan satu sesi yang ditandai sah untuk nilai. Jawaban & nilai — satu sesi memiliki banyak jawaban (one-to-many: satu jawaban per soal per sesi); satu jawaban milik satu soal dan satu peserta (many-to-one ke soal, many-to-one ke sesi); satu sesi menghasilkan satu hasil/nilai (one-to-one); satu jawaban esai dapat memiliki satu atau lebih catatan koreksi manual (one-to-many), dengan koreksi terakhir yang final dipakai; satu hasil dapat memiliki banyak nilai per kompetensi (one-to-many). Audit & keamanan — satu sesi dapat memiliki banyak peristiwa anti-curang (one-to-many); satu pengguna dapat memiliki banyak catatan perangkat (one-to-many); satu pengguna dapat menghasilkan banyak log aktivitas (one-to-many); satu jadwal dapat memiliki banyak peristiwa, banyak peserta, banyak sesi (one-to-many ke masing-masing).

**Dilarang hapus fisik** (hanya kunci + arsip): bank soal/soal/opsi/jadwal/sesi/jawaban/hasil/koreksi/token yang sedang dipakai, log, kartu yang sudah diterbitkan. **Yang boleh dihapus:** draft yang belum pernah dipakai jadwal manapun.
- **Hapus staf:** dilarang hapus fisik bila pernah membuat bank soal/paket, mengoreksi, menguji UKK, atau menjadi proktor jadwal berjalan/selesai — nonaktifkan saja.
- **Hapus kelas:** dilarang bila punya siswa aktif/jadwal/hasil — arsipkan, siswa dipindahkan dulu.
- **Hapus siswa:** dilarang bila punya sesi/jawaban/hasil/kartu/UKK — hanya status nonaktif/lulus/pindah.

**Aturan hapus/ubah (detail Researcher DB):** menghapus jurusan/kelas/mapel yang masih dipakai → **ditolak** (data master yang tidak lagi dipakai ditandai nonaktif/arsip, bukan dihapus). Menghapus pengguna (guru/siswa) → tidak dihapus fisik (akun ditandai nonaktif/arsip; riwayat ujian, jawaban, dan log tetap utuh; nama pengguna tidak dipakai ulang untuk orang lain). Menghapus soal yang belum pernah dipakai ujian → diizinkan oleh peran berwenang, dengan konfirmasi; soal masuk arsip (soft delete) agar dapat dipulihkan; opsi jawaban soal tersebut ikut diarsipkan. Menghapus soal yang sudah dipakai dalam paket/jadwal/ujian → **ditolak/dibatasi** (tidak boleh dihapus; yang diizinkan: mengeluarkan soal dari paket yang belum berjalan; menandai soal nonaktif agar tidak dipakai paket baru; membuat versi baru soal untuk perbaikan; hasil ujian lama tetap merujuk soal versi lama). Mengubah isi soal/kunci jawaban setelah soal dipakai ujian → **ditolak bila ujian sudah berjalan/selesai** (perubahan isi soal yang sudah dipakai harus membuat versi baru, bukan menimpa; perubahan kunci jawaban setelah ujian selesai hanya boleh dilakukan oleh peran super admin dengan alasan tertulis yang dicatat di log, dan nilai yang terdampak harus dihitung ulang secara eksplisit — bukan diam-diam). Menghapus paket ujian → ditolak bila paket sudah memiliki jadwal/ujian (paket yang belum dijadwalkan boleh diarsipkan; paket yang sudah dijadwalkan hanya boleh dinonaktifkan untuk jadwal baru). Menghapus jadwal ujian → ditolak bila jadwal sudah memiliki peserta/sesi/jawaban (jadwal yang belum dimulai boleh dibatalkan — status dibatalkan, bukan dihapus; jadwal yang sudah berjalan hanya boleh dihentikan/diselesaikan dengan alasan). Menghapus jawaban peserta → **tidak pernah dihapus dari aplikasi** (jawaban adalah bukti ujian; koreksi dilakukan dengan memperbarui skor/menambah catatan koreksi, bukan menghapus jawaban). Menghapus nilai/hasil ujian → **tidak pernah dihapus dari aplikasi** (nilai yang salah diperbaiki lewat koreksi tercatat + hitung ulang, dengan alasan dan pelaku tercatat di log). Mengubah nilai yang sudah diterbitkan → hanya dengan membuka kunci + alasan (sistem mencatat nilai lama, nilai baru, siapa, kapan, mengapa). Menghapus log aktivitas → tidak ada fitur hapus log di aplikasi (penghapusan hanya melalui kebijakan retensi yang dijalankan admin sistem, bukan dari menu aplikasi). Menghapus file media soal → ditolak bila file masih dipakai soal aktif (file yang tidak dipakai lagi diarsipkan; file fisik dihapus hanya setelah masa tenggang dan setelah dipastikan tidak ada soal yang merujuk). Mengubah token ujian → diizinkan oleh proktor/admin, dengan riwayat perubahan dicatat (token lama langsung tidak berlaku; peserta yang sudah masuk dengan token lama tidak dikeluarkan paksa kecuali diputuskan proktor).

**Konsistensi:** jurusan siswa = jurusan kelasnya; satu siswa satu sesi berjalan per jadwal; satu perangkat satu siswa per sesi berjalan; jumlah soal paket ≤ jumlah soal aktif bank; nilai akhir = otomatis + manual.

**Aturan konsistensi tambahan (Researcher DB):** satu jawaban per soal per sesi (sistem menolak jawaban ganda untuk pasangan sesi + soal yang sama; penyimpanan ulang memperbarui baris yang sama). Satu sesi sah per peserta per jadwal (bila ada percobaan ulang, hanya satu sesi yang ditandai sah untuk nilai; sesi lain ditandai dibatalkan/tidak dipakai, tetapi tetap tersimpan untuk audit). Susunan soal per sesi tidak berubah setelah sesi dimulai (walau paket diubah di tengah ujian, peserta yang sudah mulai tetap mengerjakan susunan yang ia terima di awal). Nilai = turunan dari jawaban (nilai tidak boleh diubah langsung tanpa mengubah/mencatat dasarnya — jawaban/koreksi; bila nilai diubah manual, alasannya wajib dicatat). Kelas siswa pada hasil ujian dibekukan (bila siswa pindah kelas setelah ujian, hasil ujian lama tetap mencatat kelas saat ujian berlangsung — disalin ke hasil/peserta jadwal — bukan mengikuti kelas terbaru).

### 16.3 Kode (penamaan)

- Bank soal: `BS-{kode mapel}-{urut 3 digit}` (mis. `BS-PROG-001`).
- Paket: `PKT-{STS/SAS/USBK/UKK/TRY/SUS}-{tahun singkat}{GJL/GNP}-{urut 3 digit}` (mis. `PKT-STS-26GJL-001`).
- Jadwal: kode paket + `-J{urut 2 digit}`.
- Kartu peserta: `KRT-{tahun 4 digit}-{jenis}-{urut 5 digit}` (mis. `KRT-2026-USBK-00001`); cetak ulang tidak mengubah nomor; QR terpisah acak.
- Token: 6 karakter huruf + angka tanpa O/0/I/1 (token demo `ABC123` hanya untuk dokumentasi).
- Event: `EVT-{tahun}-{urut 3 digit}`. Ruangan: `R-{nama}`. Kompetensi: `KOMP-{mapel}-{urut 2 digit}`.

### 16.4 Seed

6 akun demo (SUPER_ADMIN `admin@sekolah.id`/`admin123`, ADMIN_SEKOLAH/operator `operator@sekolah.id`/`operator123`, GURU `guru@sekolah.id`/`guru123`, PROKTOR `proktor@sekolah.id`/`proktor123`, PENGUJI_UKK `penguji@sekolah.id`/`penguji123`, SISWA NISN `2024001` Andi / `2024002` Bella / `siswa123` — semua wajib ganti saat login pertama; tombol "hapus data demo" hanya hapus data bertanda demo yang belum dipakai produksi). Master contoh: tahun ajaran 2026/2027 Ganjil aktif; jurusan RPL/TKJ/AKL; kelas X RPL 1/XI RPL 1/XII RPL 1/XI TKJ 1; mapel PROG/MTK/BIND/BING/INF; agama 6; kompetensi PROG 4; guru 3; siswa 10 termasuk 1 nonaktif. Bank demo `BS-PROG-001` (5 soal, sama seperti CBT-Go), `BS-MTK-001` (3 soal, terkunci), paket `PKT-STS-26GJL-001`, 1 jadwal demo, 1 event + 1 ruangan `R-LAB1`, kartu contoh Andi/Bella, 1 hasil contoh bertanda "DATA DEMO".

**Aturan seed (Researcher DB):** seed adalah data awal yang dimasukkan otomatis pada instalasi baru agar sistem langsung dapat dipakai dan didemokan — harus jelas dibedakan dari data nyata (penanda akun demo/data contoh). Akun demo memakai nama pengguna yang jelas demo (mis. mengandung kata "demo") agar tidak tertukar dengan akun nyata; kata sandi awal akun demo kuat dan terdokumentasi di panduan instalasi, bukan kata sandi lemah yang mudah ditebak; sistem memaksa ganti kata sandi super admin pada login pertama; tersedia perintah/menu "bersihkan data demo" yang menghapus akun demo + data ujian contoh, tanpa menyentuh data nyata — pembersihan ini hanya diizinkan sebelum data nyata masuk, dan meminta konfirmasi ganda. Data master contoh: 2 jurusan, 2–3 kelas, 2–3 mapel, 1 bank soal per mapel contoh (10–20 soal campuran tipe), 1 paket contoh ("STS Pemrograman Web — Contoh", durasi singkat mis. 30 menit, token contoh), 1 jadwal contoh untuk kelas contoh (status draf — tidak langsung berjalan), pengaturan awal (nama sekolah placeholder yang wajib diubah, tahun ajaran aktif, batas upload file, pesan pengumuman contoh). **Pengembangan:** seed lengkap (semua akun demo + data master contoh + soal contoh + jadwal contoh). **Produksi sekolah:** seed minimal — hanya peran + izin + akun super admin + pengaturan dasar; data master sekolah diimpor dari data nyata sekolah, bukan dari data contoh; data contoh hanya dimasukkan bila sekolah meminta mode demo. **Seed harus idempotent:** menjalankan seed dua kali tidak boleh membuat akun ganda/data ganda (seed memeriksa keberadaan data sebelum menambah).

**Pertanyaan terbuka:** apakah seed demo boleh masuk database produksi? (Jawaban yang diusulkan: tidak — seed produksi hanya role + permission + settings default. Lihat Bab 24.)

### 16.5 Skala & indeks

Tabel terbesar `jawaban_peserta` (1.000 peserta × 50 soal = 50.000 baris per ujian besar, ~500.000 baris per semester dengan 10 ujian), lalu `log_aktivitas`, `log_pelanggaran`. **Strategi indeks:** kolom penghubung + gabungan (jadwal+siswa), (sesi+soal), (jadwal+waktu), (peran+kode izin); unik untuk kode/token/NISN/email; jangan indeks berlebihan di jawaban/log karena memperlambat tulis. **Indeks login & akun:** email/username, NISN, token. **Indeks pengambilan soal:** bank+status, paket+urutan. **Indeks jawaban (jalur tersibuk):** (sesi+soal) unik, (soal), (jadwal). **Indeks rekap:** (jadwal+siswa), (jadwal+kompetensi). **Arsip:** jawaban >2 tahun ajaran → tabel arsip, rekap tetap aktif selamanya; log 1 tahun ajaran terakhir aktif, sisanya arsip read-only admin; analisis hitung berkala, bukan live.

**Indeks detail (Researcher DB) — prinsip:** indeks dibuat untuk kolom yang dipakai sebagai kunci pencarian, penyaringan, pengurutan, atau penggabungan antar tabel dalam kueri yang sering dijalankan; kolom yang jarang dicari tidak perlu indeks. **Login & akun:** indeks pada nama pengguna di tabel pengguna (setiap login mencari akun berdasarkan nama pengguna — tanpa indeks, database memindai seluruh tabel pengguna setiap login, lambat saat ratusan siswa login bersamaan); indeks pada alamat email (mempercepat pengecekan email ganda dan pemulihan akun); indeks pada rujukan pengguna di tabel Relasi Pengguna–Peran; indeks gabungan penanda status + penanda arsip di tabel pengguna (mempercepat penyaringan "hanya akun aktif" saat login massal); indeks pada rujukan pengguna di tabel Profil Siswa (setelah login siswa, sistem perlu profil siswa — kelas, NIS — untuk menentukan jadwal yang berhak diikuti). **Pengambilan soal ujian:** indeks pada rujukan paket di tabel Susunan Soal Paket + urutan soal; indeks pada rujukan soal di tabel Opsi Jawaban + urutan tampil; indeks gabungan rujukan bank + status soal di tabel Soal; indeks pada rujukan mapel/kompetensi di tabel Soal atau Bank Soal; indeks pada rujukan sesi di tabel Jawaban/Sesi (saat peserta melanjutkan ujian — refresh/reconnect — sistem harus cepat menemukan susunan soal sesi tersebut); indeks pada rujukan siswa di tabel Peserta Jadwal + rujukan jadwal (dashboard siswa menampilkan jadwal ujian yang berhak ia ikuti); indeks pada token di tabel Jadwal (atau tabel Token) + masa berlaku; indeks pada rujukan peserta + status sesi di tabel Sesi. **Penyimpanan jawaban (jalur tersibuk):** indeks unik gabungan pasangan sesi + soal di tabel Jawaban Peserta (setiap autosave mencari baris jawaban untuk pasangan sesi + soal — indeks unik ini membuat autosave menjadi pencarian langsung, sekaligus menegakkan aturan "satu jawaban per soal per sesi"); indeks pada rujukan sesi di tabel Jawaban Peserta (mengambil semua jawaban satu sesi; saat peserta membuka halaman ujian atau saat submit); indeks pada rujukan sesi + status jawaban di tabel Jawaban (perhitungan nilai memindai jawaban satu sesi); indeks pada rujukan jadwal + tipe soal/status koreksi di tabel Jawaban via sesi (guru membuka daftar esai yang belum dikoreksi untuk satu jadwal). **Rekap nilai & laporan:** indeks pada rujukan jadwal di tabel Hasil/Nilai; indeks gabungan rujukan kelas + rujukan siswa di tabel Hasil; indeks pada rujukan paket/mapel di tabel Hasil; indeks pada rujukan hasil + rujukan kompetensi di tabel Nilai per Kompetensi; indeks pada rujukan pengguna + waktu di tabel Log dan jenis objek + pengenal objek di tabel Log (audit sering mencari "apa yang dilakukan pengguna X" atau "apa yang terjadi pada objek Y"); indeks pada rujukan sesi + waktu di tabel Peristiwa Anti-Curang (monitor proktor menampilkan peristiwa terbaru per sesi). **Indeks yang sebaiknya tidak berlebihan:** jangan membuat indeks pada setiap kolom "demi keamanan" — setiap indeks memperlambat penulisan (insert/update), dan tabel jawaban adalah tabel yang paling sering ditulis; kolom teks panjang (isi soal, teks jawaban esai) tidak diindeks penuh (pencarian teks soal bila diperlukan memakai pencarian teks khusus database, bukan indeks biasa); kolom status yang nilainya hanya 2–3 macam (mis. aktif/nonaktif) tidak efektif bila diindeks sendirian — gabungkan dengan kolom lain (mis. bank + status) agar indeks benar-benar selektif; setelah sistem berjalan, indeks perlu ditinjau berkala berdasarkan pola kueri nyata (kueri mana yang lambat), bukan ditambah membabi buta. **Catatan operasional indeks:** indeks dibuat bersamaan dengan migrasi skema, bukan setelah masalah muncul; pembuatan indeks pada tabel yang sudah besar (mis. jawaban setelah ribuan ujian) harus dilakukan di luar jam ujian agar tidak mengunci tabel; untuk tabel jawaban yang akan membesar cepat, pertimbangkan partisi di masa depan (mis. per tahun ajaran/per semester) — bukan untuk v2 awal, tetapi desain harus tidak menghalangi partisi (kolom waktu/tahun ajaran tersedia pada baris jawaban atau sesi).

### 16.6 Migrasi & seed

**Prinsip migrasi:** bernomor urut, sekali jalan, tidak pernah diedit setelah rilis; migrasi baru wajib aditif + idempotent (tambah tabel/kolom/index, aman dijalankan dua kali); jangan pernah `DROP DATABASE`/`DROP TABLE`/migrasi ulang dari nol di produksi; jangan jalankan seeder demo di atas DB produksi setelah berisi data asli (kecuali terbukti idempotent + tidak menimpa/mereset; di produksi seeder idealnya dinonaktifkan/dibatasi instalasi pertama). Saat ini hanya satu file `001_schema.sql` — jika fitur Tahap 2 ditambah, migrasi harus dipecah menjadi file berversi (`002_...sql`, `003_...sql`) — jangan menumpuk semua perubahan ke `001_schema.sql` setelah DB produksi sudah dibuat. Setiap migrasi baru dicatat di catatan rilis.

**Migrasi dari CBT-Go lama (bila v2 menggantikan instalasi lama):** jadwal lama dimigrasikan menjadi paket implisit (satu paket per jadwal lama) agar riwayat tidak hilang; `jadwals.banksoal_id` (langsung ke bank) → `jadwals → paket → (bank + daftar soal terpilih)`.

**Impor data master nyata:** dari Dapodik (NISN sebagai kunci), validasi duplikat/tidak valid, jejak perubahan data peserta.

**Strategi migrasi (Researcher DB):** migrasi berversi dan berurutan (setiap perubahan skema adalah satu langkah migrasi bernomor yang dapat dijalankan maju dan, bila memungkinkan, dibatalkan; urutan migrasi tidak boleh diubah setelah diterapkan di lingkungan manapun). Migrasi dijalankan otomatis saat aplikasi mulai (mengikuti pola CBT-Go: migrasi + seed otomatis), tetapi dengan pengaman: aplikasi mencatat versi migrasi terakhir yang sudah diterapkan di database; migrasi yang sudah diterapkan tidak dijalankan ulang; kegagalan migrasi menghentikan startup dengan pesan yang jelas (bukan menjalankan aplikasi dengan skema setengah jadi); sebelum migrasi yang mengubah data besar, sistem membuat backup otomatis. Pisahkan migrasi skema dari migrasi data (migrasi skema: menambah/mengubah struktur tabel dan indeks; migrasi data: memindahkan/membersihkan data lama agar sesuai struktur baru — migrasi data yang besar dijalankan di luar jam ujian dan dapat dilanjutkan bila terputus). Tidak ada perubahan skema diam-diam (setiap migrasi mencatat nomor, deskripsi, tanggal, siapa yang membuat, apakah reversibel). Lingkungan bertahap: migrasi diuji di pengembangan → staging/salinan data sekolah → baru produksi; migrasi tidak diuji langsung di produksi. Jendela perubahan: migrasi yang berpotensi mengunci tabel (tambah kolom dengan nilai default pada tabel besar, tambah indeks, ubah tipe kolom) dijadwalkan di luar jam ujian dan diumumkan ke proktor. **Urutan migrasi yang dianjurkan** (mengikuti ketergantungan antar tabel — tabel induk dulu, baru tabel yang merujuk): (1) tahap fondasi — peran, izin per peran, pengguna, relasi pengguna–peran, pengaturan sistem; (2) tahap master kelembagaan — jurusan, kelas, mapel, profil siswa, profil guru, penugasan guru–mapel, kompetensi; (3) tahap bank soal — bank soal, soal, opsi jawaban, pasangan menjodohkan (bila terpisah), soal–kompetensi, file media; (4) tahap paket & jadwal — paket ujian, susunan soal paket, jadwal ujian, peserta jadwal, token (bila terpisah), sesi ujian; (5) tahap jawaban & nilai — jawaban peserta, hasil/nilai, nilai per kompetensi, koreksi manual, analisis butir soal; (6) tahap audit & keamanan — perangkat peserta, peristiwa anti-curang, log aktivitas; (7) tahap indeks & penguat — indeks jalur panas (Bab 16.5), batasan keunikan (satu jawaban per sesi + soal, nama pengguna unik). **Impor data master nyata:** sekolah umumnya sudah memiliki data siswa/kelas/mapel dalam bentuk spreadsheet atau sistem lain — sediakan impor berkas CSV untuk jurusan, kelas, mapel, siswa, guru. Impor harus memiliki pratinjau sebelum disimpan, laporan baris yang gagal + alasannya, dan pembatalan bila sebagian besar baris gagal (tidak menyimpan setengah-setengah). Impor siswa harus memvalidasi: NIS unik, kelas tujuan ada, nama pengguna tidak bentrok dengan akun yang sudah ada. Impor soal dari CSV (warisan CBT-Go) dipertahankan untuk bank soal, dengan format kolom yang terdokumentasi dan contoh berkas. **Migrasi dari CBT-Go lama:** jangan menimpa database lama — buat database baru untuk v2, lalu migrasikan data yang masih relevan (pengguna, data master, bank soal) dengan skrip migrasi terpisah. Data ujian lama (jawaban, nilai) dimigrasikan hanya bila formatnya kompatibel; bila tidak, arsipkan database lama sebagai baca-saja (read-only) untuk keperluan audit dan rekap historis, dan mulai v2 dengan data ujian kosong. Uji migrasi pada salinan database lama terlebih dahulu, bandingkan jumlah baris sebelum–sesudah per tabel, dan verifikasi contoh nilai lama tetap sama setelah migrasi.

### 16.7 Backup & restore (detail Researcher DB)

**Apa yang di-backup:** (1) database PostgreSQL (isi penuh) — semua tabel (pengguna, master, bank soal, paket, jadwal, sesi, jawaban, nilai, koreksi, log, pengaturan) — prioritas utama; (2) file yang diunggah — gambar/media soal, file lampiran, dan file lain di penyimpanan file lokal (database hanya menyimpan metadata file; file fisiknya harus ikut di-backup, atau soal akan kehilangan medianya setelah restore); (3) konfigurasi aplikasi — berkas konfigurasi server (koneksi database, port, pengaturan server lokal) — bukan kata sandi dalam bentuk teks biasa (kata sandi/kunci disimpan terpisah secara aman); (4) catatan versi — versi aplikasi + versi migrasi database pada saat backup diambil (tanpa ini, restore ke versi aplikasi yang berbeda bisa gagal). **Yang tidak perlu di-backup sebagai data:** hasil build aplikasi (binary Go dapat dibangun ulang), cache sementara, log aplikasi mentah yang sudah masuk tabel log (bila diduplikasi), berkas sementara.

**Cara backup PostgreSQL (`pg_dump`):** perkakas utama `pg_dump` (bawaan PostgreSQL). Bentuk backup yang dianjurkan untuk v2: backup logis penuh (via `pg_dump`) — salinan seluruh database dalam satu berkas (mudah dipindahkan, mudah di-restore ke server lain, cocok untuk skala sekolah dan backup harian) + backup berkas file (salinan direktori file yang diunggah, diambil pada waktu yang sama dengan backup database — atau sedekat mungkin — agar database dan file konsisten). Prinsip: backup diambil tanpa menghentikan layanan ujian bila memungkinkan (backup tidak boleh mengganggu ujian yang sedang berjalan); backup tidak mengambil data dari database yang sedang dalam migrasi; setiap berkas backup diberi nama yang memuat tanggal–waktu pengambilan dan versi aplikasi/migrasi; setelah backup dibuat, sistem memverifikasi berkas backup (ukuran tidak nol, dapat dibaca, checksum/sidik berkas dicatat untuk mendeteksi kerusakan/perubahan); berkas backup langsung dienkripsi sebelum disimpan/dipindahkan.

**Jadwal backup:** backup database penuh — **setiap hari**, di luar jam ujian (mis. dini hari) — backup utama harian; di musim ujian, tambahkan backup kedua di tengah hari setelah sesi ujian selesai. Backup database sebelum perubahan berisiko — **setiap kali** sebelum migrasi, impor massal, atau pembersihan data (backup sekali pakai/pre-change yang diberi label perubahan yang akan dilakukan). Backup file yang diunggah — **setiap hari**, bersamaan dengan backup database. Salinan luar lokasi (off-site) — **setiap minggu** (atau setiap hari bila memungkinkan) — salinan backup disimpan di lokasi berbeda dari server ujian (mis. hard disk eksternal terenkripsi yang disimpan di ruang berbeda, atau penyimpanan cloud sekolah bila ada) — melindungi dari kerusakan fisik server/kebakaran/pencurian. Arsip jangka panjang — **setiap akhir semester/tahun ajaran** (satu backup yang disimpan lebih lama sebagai arsip tahun ajaran, terpisah dari rotasi harian). **Rotasi backup (agar disk tidak penuh):** backup harian disimpan 14–30 hari terakhir; backup mingguan disimpan 8–12 minggu terakhir; backup arsip semester disimpan hingga kebijakan retensi sekolah mengizinkan penghapusannya; penghapusan backup lama dilakukan otomatis sesuai rotasi, dengan log penghapusan.

**Cara restore (urutan yang benar):** (1) hentikan layanan aplikasi agar tidak ada yang menulis ke database selama restore; (2) siapkan database tujuan — database kosong baru (atau database yang akan ditimpa, setelah dipastikan backup terbarunya sudah aman); (3) kembalikan database dari berkas backup menggunakan perkakas restore bawaan PostgreSQL (`pg_restore` untuk format arsip, atau pemulihan via `psql` untuk format teks — sesuai format backup yang dipakai) — pastikan versi PostgreSQL tujuan kompatibel dengan backup; (4) kembalikan file yang diunggah ke direktori penyimpanan file yang sama (atau direktori baru bila konfigurasi diubah), dan periksa izin akses berkas; (5) kembalikan konfigurasi (koneksi database, port, direktori file) dan sesuaikan bila server tujuan berbeda; (6) periksa versi migrasi yang tercatat di database hasil restore — jalankan migrasi yang tertinggal bila versi aplikasi tujuan lebih baru, atau bila versi aplikasi tujuan lebih lama dari backup, jangan memaksakan restore (kembalikan ke versi aplikasi yang sama dengan backup); (7) jalankan pemeriksaan kesehatan — aplikasi dapat start, login admin berhasil, satu bank soal dapat dibuka, satu jadwal contoh dapat dibaca, dan satu hasil ujian contoh cocok dengan sebelum backup; (8) baru buka layanan untuk pengguna setelah pemeriksaan kesehatan lulus.

**Tes restore (wajib, bukan opsional):** backup yang belum pernah diuji restore dianggap **tidak ada**. Tes restore terjadwal: minimal **satu kali per bulan**, dan **satu kali sebelum musim ujian besar** (STS/SAS/USBK). Cara tes: restore backup terbaru ke server/komputer terpisah (bukan menimpa produksi), lalu jalankan pemeriksaan kesehatan + verifikasi jumlah baris pada tabel-tabel kunci (pengguna, soal, jadwal, jawaban, nilai) cocok dengan sumber. Catat hasil tes (tanggal tes, backup mana yang diuji, siapa yang menguji, hasil pemeriksaan, masalah yang ditemukan); bila tes gagal, backup dianggap tidak dapat dipercaya dan jadwal backup diperbaiki sebelum ujian berikutnya. Latihan restore untuk operator: minimal satu operator selain admin utama pernah melakukan restore, agar tidak ada ketergantungan pada satu orang saat darurat.

**Skenario darurat:** database rusak/server mati di tengah ujian — hentikan ujian, pulihkan dari backup terbaru + file yang diunggah, verifikasi sesi yang sedang berjalan (sesi yang belum tersimpan di backup terakhir mungkin perlu diulang atas keputusan proktor — peserta harus diberi tahu secara transparan). Salah hapus data (mis. bank soal terhapus) — jangan panik dan jangan menimpa backup terbaru dengan keadaan yang salah; restore ke database terpisah, ambil data yang hilang, lalu gabungkan kembali dengan pengawasan admin. Server hilang/rusak fisik — bangun server pengganti, install PostgreSQL + aplikasi versi yang sama, restore backup luar lokasi (off-site) terbaru, lalu verifikasi.

### 16.8 Keamanan data (detail Researcher DB)

**Enkripsi koneksi (data dalam perjalanan):** koneksi aplikasi → database — aktifkan koneksi terenkripsi (TLS/SSL) antara aplikasi Go dan PostgreSQL, walau keduanya di server yang sama (mencegah penyadapan kredensial dan data di jaringan lokal). Koneksi browser → aplikasi — seluruh akses aplikasi memakai HTTPS (atau HTTP hanya di jaringan lokal yang benar-benar tertutup — dan tetap dianjurkan HTTPS walau lokal); sertifikat dapat berupa sertifikat internal sekolah/sertifikat lokal yang dipercaya perangkat sekolah. Kata sandi database tidak ditulis di kode dan tidak ditulis di berkas konfigurasi dalam bentuk teks biasa yang dibagikan sembarangan — kata sandi disimpan di mekanisme rahasia yang disediakan lingkungan server (environment/vault/berkas konfigurasi dengan izin akses ketat, hanya dapat dibaca pemilik layanan). Kredensial berbeda per lingkungan (pengembangan, staging, produksi berbeda; kredensial produksi tidak dipakai di pengembangan).

**Enkripsi data tersimpan (data at rest):** kata sandi pengguna — disimpan hanya dalam bentuk hash dengan algoritma hash kata sandi modern + garam (salt) per pengguna (hash tidak dapat dibalik menjadi kata sandi asli; algoritma hash lama — MD5/SHA-1 tanpa garam — tidak boleh dipakai untuk kata sandi). Token ujian — bila token disimpan di database, simpan dalam bentuk hash bila memungkinkan (seperti kata sandi), agar kebocoran database tidak langsung membocorkan token yang masih berlaku; token asli hanya ditampilkan sekali saat dibuat. Data sensitif lain (mis. jawaban esai yang memuat data pribadi, catatan koreksi) — tidak perlu enkripsi per kolom pada v2 awal, tetapi aksesnya dibatasi lewat izin peran + log audit; bila sekolah menangani data yang sangat sensitif, enkripsi disk/volume server (full-disk encryption) adalah lapisan pertahanan dasar. Backup terenkripsi — semua berkas backup (database + file) dienkripsi sebelum disimpan/dipindahkan/diunggah ke luar lokasi; kunci enkripsi backup disimpan terpisah dari berkas backup (jangan simpan kunci di folder yang sama dengan backup); tanpa kunci, backup yang dicuri tidak dapat dibaca.

**Kontrol akses & prinsip hak minimum:** setiap akun hanya memiliki izin yang diperlukan perannya (guru tidak dapat mengubah pengaturan sistem; siswa tidak dapat melihat bank soal/kunci jawaban; proktor tidak dapat mengubah nilai). Akses langsung ke database (mis. via alat database) hanya untuk admin sistem, tidak untuk pengguna aplikasi sehari-hari; akses langsung dicatat. Akun bersama (satu akun dipakai banyak orang) dilarang — setiap guru/proktor/siswa memiliki akun sendiri agar log audit bermakna. Akun yang tidak aktif lama (mis. siswa lulus) dinonaktifkan, bukan dibiarkan aktif. Upaya login gagal dibatasi (jeda/kunci sementara setelah beberapa kali gagal) untuk menghambat tebakan kata sandi; upaya login gagal dicatat di log.

**Retensi log & data:** log aktivitas (audit) — **minimal 1 tahun ajaran**, idealnya 2 tahun ajaran (log adalah bukti audit; jangan hapus log sebelum masa retensi habis; penghapusan log setelah masa retensi dilakukan oleh admin sistem dengan pencatatan). Peristiwa anti-curang — selama hasil ujian terkait masih disimpan (bukti pendukung nilai/keputusan blokir; dihapus bersamaan dengan arsip hasil ujian, bukan lebih awal). Jawaban & nilai ujian — sesuai kebijakan sekolah/dinas (minimal hingga akhir tahun ajaran + masa sanggah nilai; jangan hapus jawaban/nilai selama masa sanggah masih berjalan; arsipkan, bukan hapus). Data siswa yang sudah lulus/pindah — diarsipkan, bukan dihapus langsung (akun dinonaktifkan; data ujian historis tetap tersimpan untuk rekap/legal; penghapusan permanen hanya sesuai kebijakan privasi sekolah dan setelah masa retensi). Backup harian/mingguan — sesuai rotasi (14–30 hari / 8–12 minggu); backup arsip semester disimpan lebih lama. Log aplikasi teknis (bukan audit) — 30–90 hari (log teknis — error, performa — boleh dirotasi lebih cepat daripada log audit). **Prinsip retensi:** retensi ditetapkan tertulis oleh sekolah (bukan diputuskan sepihak oleh aplikasi), diumumkan ke pengguna, dan dijalankan otomatis; setiap penghapusan data karena retensi dicatat (apa yang dihapus, kapan, berdasarkan kebijakan apa).

**Keamanan operasional lainnya:** pembaruan rutin (PostgreSQL, sistem operasi, dependensi Go diperbarui secara berkala; pembaruan diuji di staging sebelum produksi). Cadangan konfigurasi firewall (server lokal sekolah hanya membuka port yang diperlukan — port aplikasi + port database hanya untuk localhost/aplikasi, bukan untuk seluruh jaringan bila tidak perlu). Pencatatan perubahan (setiap perubahan konfigurasi, migrasi, dan restore dicatat — siapa, kapan, apa — agar insiden dapat ditelusuri). Rencana respons insiden (bila terjadi kebocoran/akses tidak sah, langkahnya: isolasi, simpan bukti — jangan hapus log — ganti kredensial yang terdampak, beri tahu pimpinan sekolah, pulihkan dari backup yang bersih, catat pelajaran).

### 16.9 Arsitektur keseluruhan (Researcher DB)

```
Browser siswa (PC lab/laptop/HP) ──┐
Browser guru/proktor (laptop/PC) ──┼── HTTPS ──▶ SERVER LOKAL SEKOLAH
Browser admin (laptop/PC) ─────────┘                      │
                              ┌─────────────────────────────┼─────────────────────────────┐
                              │ APLIKASI GO (single binary) │                             │
                              │ - HTTP API / halaman web    │                             │
                              │ - WebSocket (monitor)       │                             │
                              │ - Logika ujian: timer, token, anti-curang, penskoran     │
                              │ - Manajemen file (upload)   │                             │
                              │ - Scheduler: backup, pembersihan sesi, penghitungan nilai │
                              └──────┬──────────────┬───────┘                             │
                                     ▼              ▼                                     │
                              PostgreSQL        File Storage (disk lokal)                 │
                              (sumber kebenaran) - Media soal, lampiran, ekspor, backup   │
                              └───────────────────────────────────────────────────────────┘
                                                   │
                                                   ▼
                              LUAR LOKASI (opsional, non-kritis): salinan backup
                              terenkripsi + unduh pembaruan aplikasi
```

**Keterangan komponen:** browser peserta/guru/proktor — antarmuka pengguna (tidak menyimpan data ujian sebagai sumber kebenaran; hanya menampilkan dan mengirim jawaban ke server). Aplikasi Go (single binary) — otak sistem (satu binary berisi HTTP API, halaman web, logika ujian, WebSocket, scheduler; mudah di-deploy — sesuai warisan CBT-Go). PostgreSQL — sumber kebenaran data (semua data terstruktur; satu-satunya tempat yang menentukan nilai akhir). File Storage (disk lokal) — penyimpanan berkas (media soal, lampiran, ekspor CSV, backup lokal; metadata file tercatat di database, isi file di disk). WebSocket — saluran real-time (monitor peserta real-time, notifikasi peringatan anti-curang, jeda jadwal, submit berhasil, pembaruan status tanpa refresh; WebSocket **tidak** dipakai untuk mengirim soal/jawaban utama — soal/jawaban tetap lewat HTTP API agar dapat dicoba ulang/retry dengan aman). Backup & Pemeliharaan — operasi terjadwal (backup database + file, rotasi backup, pembersihan sesi kedaluwarsa, penghitungan ulang nilai bila diperlukan; berjalan sebagai tugas terjadwal di server). Jaringan lokal sekolah (LAN/Wi-Fi lokal, tanpa internet wajib untuk ujian inti). Luar lokasi (opsional, non-kritis): salinan backup terenkripsi + unduh pembaruan aplikasi.

**Aliran data saat ujian (end-to-end):**
- **Tahap A — Persiapan (sebelum hari ujian):** (1) guru membuat/memilih bank soal → soal + opsi + kunci tersimpan di database; (2) guru merakit paket ujian → susunan soal paket + aturan paket (durasi, pengacakan, batas pelanggaran) tersimpan; (3) proktor membuat jadwal → jadwal + token + daftar peserta tersimpan (token dibagikan ke siswa melalui saluran resmi sekolah); (4) siswa login → browser mengirim nama pengguna + kata sandi via HTTPS → aplikasi memverifikasi hash kata sandi di database → aplikasi membuat sesi login → browser menyimpan token sesi; (5) siswa melihat dashboard → aplikasi mengambil daftar jadwal yang berhak diikuti siswa (via Peserta Jadwal + Jadwal).
- **Tahap B — Memulai ujian:** (6) siswa memasukkan token ujian → aplikasi memeriksa token + masa berlaku + status jadwal + apakah siswa termasuk peserta jadwal; (7) aplikasi membuat sesi ujian → baris Sesi Ujian baru (waktu mulai, perangkat, alamat jaringan) + menyimpan **susunan soal per sesi** (urutan soal + urutan opsi setelah pengacakan, bila diaktifkan) — susunan ini dibekukan untuk sesi tersebut; (8) aplikasi mengirim soal ke browser → browser menerima soal satu per satu atau per halaman, **tanpa kunci jawaban** (kunci jawaban tidak pernah dikirim ke browser); (9) proktor memantau via WebSocket → dashboard proktor menerima pembaruan real-time (siapa sudah mulai, siapa belum, siapa terkena peringatan).
- **Tahap C — Mengerjakan ujian (jalur tersibuk):** (10) siswa menjawab soal → browser menyimpan jawaban sementara di memori browser (untuk ketahanan terhadap refresh singkat); (11) autosave berkala → browser mengirim jawaban ke server via HTTP API (setiap perubahan jawaban + setiap interval waktu) — server memperbarui baris Jawaban Peserta untuk pasangan sesi + soal (update, bukan insert baru) — server mengembalikan konfirmasi "tersimpan" + stempel waktu server; (12) timer — timer resmi adalah **timer server** (waktu mulai sesi + durasi); timer di browser hanya tampilan; bila waktu habis, server menutup sesi walau browser tidak mengirim submit; (13) tandai ragu-ragu → browser mengirim penanda ragu-ragu per soal → server memperbarui kolom penanda pada baris jawaban; (14) peristiwa anti-curang → browser mendeteksi pindah tab/keluar fullscreen → browser mengirim peristiwa ke server → server menyimpan baris Peristiwa Anti-Curang + menghitung jumlah pelanggaran sesi → bila melewati batas: server mengirim peringatan (via WebSocket/tampilan) atau memblokir sesi otomatis sesuai aturan paket; keputusan proktor (cabut blokir/lanjutkan) juga dicatat; (15) gangguan koneksi → bila jaringan terputus singkat, browser menyimpan jawaban yang belum terkirim dan mengirim ulang saat koneksi kembali; server menerima kiriman ulang secara idempotent (kiriman yang sama tidak membuat jawaban ganda); bila perangkat berganti, server mencatat perangkat baru di tabel Perangkat Peserta.
- **Tahap D — Submit & penskoran:** (16) siswa menekan Submit (atau waktu habis) → browser mengirim semua jawaban terakhir + konfirmasi submit → server mengunci sesi (status selesai), mencatat waktu selesai aktual, menolak perubahan jawaban setelah submit; (17) penskoran otomatis → server menghitung skor per soal untuk soal objektif dengan membandingkan jawaban peserta dengan kunci jawaban yang tersimpan di database; skor otomatis disimpan pada baris Jawaban Peserta; (18) agregasi nilai → server menjumlahkan skor per soal menjadi Hasil/Nilai Ujian (skor total, nilai 0–100, jumlah benar/salah/kosong) — hasil disimpan dengan status "menunggu koreksi" bila masih ada esai yang belum dikoreksi; (19) koreksi manual esai → guru membuka daftar jawaban esai yang belum dikoreksi untuk jadwal tersebut → guru memberi skor + komentar → aplikasi menyimpan baris Koreksi Manual + memperbarui skor manual pada baris jawaban → setelah semua esai terkoreksi, server menghitung ulang Hasil/Nilai (skor akhir = objektif + esai) dan menandai status koreksi selesai; (20) penerbitan nilai → proktor/admin menerbitkan nilai (status hasil menjadi diterbitkan) → siswa dapat melihat nilai + pembahasan (bila diizinkan paket); nilai yang sudah diterbitkan tidak dapat diubah tanpa membuka kunci + alasan.
- **Tahap E — Rekap, ekspor, audit:** (21) rekap nilai → guru/proktor membuka rekap per jadwal/per kelas → aplikasi membaca tabel Hasil (bukan menghitung ulang dari jawaban setiap kali) + tabel Nilai per Kompetensi bila ada → menampilkan tabel rekap + statistik (rata-rata, kelulusan, distribusi); (22) ekspor → aplikasi menghasilkan berkas CSV rekap dari tabel Hasil dan menyimpannya sementara di File Storage untuk diunduh — ekspor dicatat di Log Aktivitas (siapa mengekspor apa, kapan); (23) analisis butir soal → secara berkala (atau atas permintaan), sistem menghitung Analisis Butir Soal dari tabel Jawaban + Hasil dan menyimpan hasilnya (tingkat kesulitan, daya beda, pengecoh) untuk laporan guru; (24) audit → semua aksi penting selama Tahap A–E sudah tercatat di Log Aktivitas — bila ada sengketa nilai, admin dapat menelusuri: sesi apa, susunan soal apa, jawaban apa, koreksi siapa, dan kapan nilai berubah.

**Keputusan arsitektur kunci & alasannya:** tetap satu binary Go + satu PostgreSQL (sederhana di-deploy di server sekolah, mudah di-backup, sudah terbukti di CBT-Go; tidak perlu microservices untuk skala ratusan siswa); PostgreSQL sebagai satu-satunya sumber kebenaran (menghindari dua sumber kebenaran browser vs server; nilai, jawaban, dan status sesi hanya sah bila tercatat di database); kunci jawaban tidak pernah dikirim ke browser (mencegah kecurangan via inspeksi browser/alat pengembang; browser hanya menerima soal, bukan kunci); timer resmi di server, bukan di browser (mencegah manipulasi timer dengan mengubah jam perangkat/memodifikasi browser); soal/jawaban utama via HTTP API, WebSocket hanya untuk monitor/notifikasi (HTTP lebih mudah dicoba ulang/retry dan lebih dapat diandalkan untuk data penting; WebSocket dipakai untuk hal real-time yang boleh hilang — status monitor — tanpa merusak ujian); file media di disk, metadata di database (database tetap ramping; file besar tidak membebani database; backup harus mencakup keduanya); server lokal sebagai sumber kebenaran saat ujian (mode semi-daring: ujian tidak bergantung pada internet; internet hanya untuk backup luar lokasi/pembaruan, bukan untuk alur ujian inti); soft delete + versi soal untuk data ujian (melindungi audit dan nilai historis; penghapusan fisik hanya untuk data yang benar-benar belum pernah dipakai).

**Risiko & mitigasi arsitektur:** lonjakan login bersamaan (ratusan siswa login dalam 5 menit) → login lambat/timeout — mitigasi: indeks login, sesi login ringan, jadwal login bertahap per gelombang bila perlu, uji beban sebelum ujian besar. Autosave bersamaan membebani database → jawaban gagal tersimpan — mitigasi: indeks unik sesi + soal, autosave dengan interval yang wajar (bukan per ketikan), retry idempotent di browser, pemantauan beban saat ujian. Server mati di tengah ujian → sesi terputus, jawaban terakhir belum tersimpan — mitigasi: autosave berkala + timer server + backup harian + prosedur resume sesi (siswa login ulang dan melanjutkan sesi yang sama dengan sisa waktu yang tersisa, bila diizinkan proktor). Token bocor ke siswa yang tidak berhak → siswa tidak berhak masuk ujian — mitigasi: token per jadwal + masa berlaku + daftar peserta yang berhak + pergantian token yang mudah oleh proktor + log percobaan masuk yang gagal. Soal/kunci bocor → ujian tidak sah — mitigasi: kunci tidak dikirim ke browser, akses bank soal dibatasi peran, log akses soal, rotasi soal antar gelombang bila memungkinkan. Disk penuh karena backup/log → aplikasi berhenti — mitigasi: rotasi backup + rotasi log + pemantauan ruang disk + peringatan dini sebelum disk penuh. Satu operator memegang semua pengetahuan → gagal pulih saat operator absen — mitigasi: dokumentasi operasi + latihan restore + minimal dua orang yang dapat melakukan backup/restore.

**Daftar periksa implementasi (handoff ke tim pengembang):** skema dibuat mengikuti daftar tabel Bab 16.1 (nama kolom fisik diputuskan di implementasi, bukan di laporan ini); relasi + aturan hapus/ubah Bab 16.2 ditegakkan di lapisan aplikasi (dan, bila memungkinkan, di lapisan database); indeks Bab 16.5 dibuat bersama migrasi awal; migrasi berversi + seed idempotent (Bab 16.6) teruji di pengembangan → staging → produksi; backup harian + rotasi + tes restore bulanan (Bab 16.7) dijadwalkan dan didokumentasikan; enkripsi koneksi + hash kata sandi + backup terenkripsi + retensi log (Bab 16.8) diterapkan; aliran ujian Bab 16.9 diuji end-to-end (login → token → soal → autosave → submit → koreksi esai → nilai → rekap); uji beban: login massal + autosave bersamaan + submit massal, dengan ratusan sesi simulasi.

## 17. Rencana UI (Halaman & Komponen)

### 17.1 Route

Admin `/adm/*`, peserta `/*` (tanpa prefix). **Perlu persetujuan:** route v2 tetap pola `/adm/*` + `/*` seperti v1 (diusulkan: ya).

| Area | Halaman |
|---|---|
| Publik | `/login` (peserta NIS + password), `/adm/login` (staf email + password), `/blocked`, 404 BARU |
| Peserta | `/dashboard`, `/token` (token 6 karakter), `/exam/:jadwalId` (halaman paling kritis), `/hasil-saya`, `/profil` BARU |
| Admin (EVIL) | Semua halaman |
| EVAL | Bank soal + koreksi utama, jadwal lihat/buat jika diizinkan, monitor lihat saja, token lihat/generate untuk jadwalnya, master lihat saja; **tidak boleh** users/settings |
| OPRA | Dashboard versi proktor + monitor utama + unblock + jadwal pause/resume/finish + token + hasil lihat saja + peserta lihat saja; **tidak boleh** bank soal/koreksi/users/settings/jurusan/matpel |

**Catatan:** route guard di frontend DAN backend (frontend hanya sembunyikan menu, backend menolak aksi terlarang). EVAL/OPRA membuka route yang dilarang → "Akses Ditolak" (bukan blank). Admin buka `/exam` → Dashboard Admin + pesan. Peserta buka `/adm/*` → Login Admin atau Dashboard Peserta. Sesi kedaluwarsa → Toast + login. Link jadwal dibagikan → Login → kembali ke link → Token. Peserta diblokir lalu di-unblock → refresh Dashboard → status normal. Siswa tidak bisa buka `/exam` tanpa token valid + jadwal aktif → dialihkan ke `/token`.

### 17.2 Komponen dipakai ulang

- **Layout:** AppHeader, Sidebar, BottomNav HP, PageHeader, EmptyState, ErrorState, LoadingSkeleton.
- **Data:** Card, StatCard, DataTable, SearchInput, FilterBar, Pagination, Badge/StatusChip, Avatar.
- **Form:** TextInput, Select/Dropdown, Textarea, Checkbox, RadioGroup, FileUpload, Modal/Dialog, ConfirmDialog, Toast/Snackbar, Tabs, Tooltip.
- **Khusus ujian:** ExamTimer, QuestionNavigator, QuestionCard, AnswerOption, MatchingRow, EssayAnswer, DoubtToggle, AntiCheatWarning, SubmitConfirmModal.
- **Khusus admin:** TokenCard, MonitorRow, CorrectionForm, ScoreChart BARU, ImportPreview.

### 17.3 Wireframe teks 3 halaman terpenting

**(1) Halaman Ujian Siswa** — lihat Bab 9.3 (satu soal per layar, header sticky + timer + progress, QuestionNavigator, SubmitConfirmModal, perilaku HP/desktop, anti-cheat).

**(2) Dashboard Admin** — desktop: sidebar kiri + StatCard (peserta/bank soal/jadwal aktif/diblokir, bisa diklik ke halaman terkait) + Card Ujian Berlangsung + Card Perlu Perhatian + Card Grafik Mini + Card Aktivitas Terbaru. HP: drawer + StatCard stacking + BottomNav/FAB + Jadwal Baru. Badge angka sidebar sinkron dengan data.

**(3) Bank Soal** — desktop: DataTable (kode/mapel/soal/status/aksi) + filter + pagination. Detail soal: Tabs (Daftar Soal/Pratinjau/Pengaturan) + Card per soal + form tambah soal yang berubah sesuai tipe + ImportPreview (baris valid/tidak valid terlihat dulu) + ConfirmDialog untuk hapus. HP: Card per bank soal. Bank terkunci: ikon gembok + Edit nonaktif + penjelasan cara buka kunci.

### 17.4 Flow per role (ringkas)

- **Peserta:** Login → Dashboard → Token → Ujian → SubmitConfirmModal → Hasil. Diblokir → `/blocked` → proktor unblock → Dashboard.
- **EVIL:** Login Admin → Dashboard → Bank Soal/Jadwal/Token/Monitor/Koreksi/Hasil/Master/Users/Settings → Logout.
- **EVAL:** Bank Soal/Koreksi/Hasil/Jadwal/Monitor/Token; menu Users/Settings tidak tampil.
- **OPRA:** Dashboard versi proktor → Monitor → Jadwal pause/resume/finish → Token → Hasil/Peserta lihat saja; Bank Soal/Koreksi/Users/Settings tidak tampil.

**Yang perlu persetujuan (Bab 24):** palet warna final (biru indigo diusulkan) dan font final (Plus Jakarta Sans diusulkan); apakah Analisis Butir Soal ikut v2 atau ditunda; dark mode backlog atau dibatalkan; route v2 tetap pola `/adm/*` + `/*` (diusulkan: ya). Setelah persetujuan: scaffolding React + prioritas (1) Halaman Ujian Siswa, (2) Dashboard Admin, (3) Bank Soal, lalu lainnya.

**Catatan kejujuran:** chat 2 kolom simultan desktop belum direncanakan — Chat masih dua halaman terpisah (daftar + thread) tanpa props silang; mengubahnya butuh ubah route/props/controller. Sebagai gantinya: satu panel per tampilan + tombol kembali + layout responsif. (Catatan ini berasal dari pengalaman ShopeeKu; untuk CBT v2, chat buyer-seller tidak ada — yang ada adalah monitor + notifikasi. Prinsip yang sama berlaku: jangan klaim layout 2 kolom simultan bila belum dirancang.)

---

## 18. Rencana Sprint & Milestone

### 18.1 Sprint (Lead Dev — 1 developer fokus + AI)

| Sprint | Isi | Durasi |
|---|---|---|
| **Sprint 0** | Fondasi (repo, monorepo, CI, migrasi dasar, design system, auth kerangka) | 3–4 hari |
| **Sprint 1** | Auth + RBAC | 7–10 hari |
| **Sprint 2** | Master Data | 6–7 hari |
| **Sprint 3** | Bank Soal | 9–10 hari |
| **Sprint 4** | Paket & Jadwal | 6–7 hari |
| **Sprint 5** | Ujian Peserta | 9–10 hari |
| **Sprint 6** | Koreksi & Nilai | 6–7 hari |
| **Sprint 7** | Laporan & Analisis | 6–7 hari |
| **Sprint 8** | UKK | 9–10 hari |
| **Sprint 9** | Hardening & Rilis | 6–7 hari |
| **Total** | | **±73–79 hari kerja ≈ 15–16 minggu kalender** (belum termasuk UAT di sekolah, jeda persetujuan, pengadaan server) |

### 18.2 Milestone

- **A (akhir Sprint 1):** matriks 6 role disetujui.
- **B (akhir Sprint 4):** satu paket end-to-end bisa didemo.
- **C (akhir Sprint 5):** satu kelas bisa ujian simulasi penuh + tryout kecil dengan siswa nyata sebelum lanjut.
- **D (akhir Sprint 7):** nilai + laporan + analisis divalidasi guru vs hitungan manual.
- **E (akhir Sprint 9):** rilis v2.0 (UAT lulus, backup teruji, tanpa akun demo di produksi).

### 18.3 Definition of Done (umum)

Fitur selesai bila: kode + migrasi + seed (bila perlu) + uji unit/integrasi lulus + uji manual alur terkait lulus + dokumentasi diperbarui + tidak ada bug Kritis terbuka + review/ADR bila keputusan besar + backup/restore teruji (bila menyentuh data).

### 18.4 Keputusan yang perlu persetujuan user sebelum Sprint 0

(1) monorepo + single binary Go (frontend di-embed) vs 2 repo; (2) 6 role + matriks permission (khususnya batasan pengawas + penguji UKK); (3) aturan anti-curang default (jumlah pelanggaran sebelum auto-block, device lock sejak hari pertama?, fullscreen di v2.0 atau ditunda); (4) ruang lingkup UKK v2 (satu format rubrik + satu format sertifikat untuk semua jurusan, variasi ditunda v2.1); (5) penundaan AI + import .docx ke pasca-v2 (catatan: laporan ini mengusulkan import Word .docx di v2 — lihat kontradiksi Bab 23; perlu diputuskan).

---

## 19. Rencana Testing (QA)

### 19.1 Prinsip

Bug saat ujian produksi = bencana. Pengujian harus membuktikan alur ujian ujung-ke-ujung, bukan hanya "bisa login".

**Catatan ketidaksesuaian:** judul penugasan QA menyebut "Go + React", tetapi README CBT-Go menyatakan Go + `html/template` server-rendered tanpa build step/JS framework — rencana QA disusun mengikuti README (server-rendered); jika v2 benar-benar React, Bagian e2e perlu revisi (tambah pengujian komponen React + build Vite). Laporan ini mengikuti perintah user: v2 adalah Go + React.

### 19.2 Strategi 3 level

| Level | Tool | Cakupan |
|---|---|---|
| **(1) Unit (Go)** | `go test` + `httptest` + `testify` (assert/require), table-driven untuk 16 tipe soal | Cakupan penilaian per tipe, nilai akhir (auto + manual), token (format 6 karakter, expiry, revoke), anti-curang (`out_counter`, `MAX_COUNTER_OUT=1`), validasi input, helper auth (verify bcrypt saja). Tidak menguji query SQL asli / tampilan HTML / DB produksi |
| **(2) Integrasi API/Handler + DB** | `go test` + PostgreSQL khusus tes (terpisah, di-reset per run) + `httptest` lewat HTTP sungguhan | Login → session, CRUD master, bank soal + lock/unlock, jadwal + pause/resume/finish, token generate/revoke, mulai ujian → autosave → auto-submit → koreksi manual → nilai akhir, monitor, export CSV, settings. DB tes terpisah; seed dummy (bukan NIS/nama asli); password demo dilarang di produksi — pengujian harus verifikasi password demo sudah diganti sebelum hari-H |
| **(3) E2E browser** | **Playwright** (utama: Chromium + Firefox + emulasi HP/tablet, auto-wait, screenshot/video saat gagal, simulasi pindah tab/blur/minimize) | Alur klik + timer + tampilan. Alternatif ditolak: Selenium (setup driver berat, auto-wait lemah), Cypress (dukungan Firefox terbatas, kurang cocok multi-tab; boleh untuk alur admin saja, skenario anti-curang multi-tab tetap Playwright) |

Pembagian: logika nilai → unit; penyimpanan/transaksi → integrasi; alur klik + timer + tampilan → e2e. E2E tidak menguji 16 tipe satu per satu (cukup 5 representatif seperti `BS-PROG-001`: PG, benar-salah, isian singkat, PG kompleks, esai).

### 19.3 Skenario alur lengkap (A1–D6)

- **Tahap A (persiapan admin/guru):** login EVAL, buat jurusan + matpel, tambah peserta manual + CSV campuran valid/duplikat/kosong, buat bank soal 5 soal, uji kunci tiap tipe, lock/unlock, import CSV benar/salah.
- **Tahap B (jadwal + token):** buat jadwal 60 menit sesi 1, generate token 6 karakter, generate token kedua (perilaku terdokumentasi konsisten), revoke → ditolak, pause → resume, force finish.
- **Tahap C (pelaksanaan peserta):** login NIS + password, token benar → 5 soal + timer dari durasi (mis. 59:58), jawab soal 1 lalu refresh → jawaban tetap tersimpan + timer tidak reset, ragu-ragu bolak-balik, kosongkan 1 soal + isi esai + submit, peserta kedua bersamaan independen, monitor proktor "sedang ujian" + refresh.
- **Tahap D (koreksi + nilai + rekap):** nilai otomatis benar + esai "menunggu koreksi", guru koreksi esai (mis. 80) → nilai akhir = auto + manual sesuai bobot (sama dengan unit test), ubah nilai koreksi → nilai akhir berubah + riwayat tercatat, rekap per jadwal lengkap, export CSV kolom lengkap + karakter khusus tidak merusak kolom, peserta hanya lihat nilainya sendiri.

**Kriteria lulus:** seluruh A1–D6 lulus; satu kegagalan Tahap C/D = tidak siap ujian.

### 19.4 Skenario anti-curang (AC1–AC28, ringkas)

- **AC1–AC8 pindah tab/keluar ujian:** 1x peringatan + counter 1; 2x auto-block + terkunci + monitor "diblokir"; minimize = pindah tab; Alt-Tab sama (jika tidak terdeteksi dicatat celah); refresh F5 bukan pelanggaran; diblokir tetap diblokir sampai proktor unblock/reset; proktor unblock → aturan mulai dari awal vs lanjut sisa waktu harus satu aturan tertulis; pindah tab tepat saat auto-submit → tidak ada nilai ganda/hilang/menggantung.
- **AC9–AC12 login ganda & perangkat:** 2 browser/tab bersamaan — sesi lama ditendang / sesi baru ditolak / diblokir, tidak boleh keduanya menjawab tanpa aturan; login perangkat kedua setelah diblokir → tetap diblokir; 2 peserta bergantian di 1 perangkat — jawaban tidak tertukar; perangkat dicatat + tampil di monitor/log.
- **AC13–AC18 token:** salah 1 karakter → ditolak + tercatat; kedaluwarsa → ditolak + proktor generate baru; revoked → ditolak; token jadwal A untuk jadwal B → ditolak; token kosong/spasi/huruf kecil-besar — konsisten + didokumentasikan; brute force token — setelah ambang → rate limit/jeda/blokir sementara, jika belum ada dicatat celah wajib diperbaiki sebelum produksi.
- **AC19–AC23 waktu habis & gangguan jaringan:** waktu habis saat mengetik → auto-submit + jawaban autosave tersimpan; waktu habis saat offline → saat kembali: tidak ada jawaban autosave hilang + status jelas; putus 30 detik → lanjut + autosave gagal dicoba ulang / diberitahu; tombol Back/tutup tab → jawaban tersimpan tidak hilang; copy-paste/klik kanan/DevTools — didokumentasikan, jangan diklaim "anti-copy" bila belum diuji/diimplementasikan.
- **AC24–AC28 keamanan terkait:** peserta akses `/adm` langsung → ditolak/diarahkan login; guru EVAL coba aksi EVIL → ditolak (permission ditegakkan di server bukan hanya sembunyikan menu); peserta ubah ID peserta/jadwal di URL → hanya datanya sendiri; submit dengan field tambahan/nilai dimanipulasi → server hitung nilai sendiri, nilai kiriman browser diabaikan (kritis); form tanpa CSRF — kondisi saat ini CSRF belum ada, dicatat utang keamanan, request lintas situs yang mengubah data tidak boleh bisa dipicu dari situs lain, jika lolos CSRF wajib ditambah sebelum produksi.

### 19.5 Uji beban 300 siswa serentak

Desain: baseline (1 pengguna) → ramp-up bertahap 50→100→200→300 selama 5–10 menit (jangan langsung 300 pada run pertama) → puncak 300 aktif bersamaan 30–60 menit (login → token → buka ujian → jawab + autosave tiap soal → submit) → spike start (300 menekan "Mulai Ujian" dalam 60 detik yang sama — kasus terburuk yang paling sering menjatuhkan CBT) → ramp-down + recovery. Di **staging** (server + DB setara produksi), bukan produksi; tanpa izin dilarang tes ke server orang lain/produksi.

**Metrik + ambang:** halaman ujian p95 <1 detik (ideal <500 ms); autosave p95 <500 ms; login + token p95 <1 detik; submit p95 <2 detik; error rate (5xx + timeout) <0,1% (maks toleransi <0,5% dengan penyebab jelas; error saat submit = Kritis); throughput dicatat baseline (tidak boleh anjlok >30% dari baseline saat 300 pengguna); koneksi DB tanpa antrean panjang / "too many connections"; CPU <80% berkelanjutan; memori tidak naik terus (indikasi bocor) dan tidak OOM; kehilangan autosave **0** (satu jawaban hilang = Kritis); timer tidak meleset >5 detik dari server + auto-submit tepat waktu; monitor tetap terbuka + refresh <5 detik tanpa timeout. Tool: **k6** (utama: skenario sebagai skrip, thresholds di dalam skenario, hemat resource, mudah CI); alternatif JMeter (GUI) atau Locust (Python); wajib ramp-up bertahap + ambang + laporan p95 + error rate (bukan hanya rata-rata).

**Kriteria lulus beban:** semua ambang terpenuhi pada 300 pengguna DAN spike start, tanpa kehilangan autosave, tanpa error submit; jika gagal, perbaiki kapasitas/konfigurasi (pool DB, timeout, indeks, resource) — bukan menurunkan ambang diam-diam.

### 19.6 Responsif + browser

Perangkat: HP kecil 360px, HP umum 375–414px, tablet 768px portrait + landscape, laptop/desktop 1280px, desktop besar 1920px. Browser: Chrome desktop + Android (wajib, versi terbaru), Firefox desktop (wajib, versi terbaru), Edge desktop (opsional, alur utama), Safari iPhone/iPad (opsional, alur utama). Per browser: seluruh skenario alur + anti-curang (khususnya deteksi pindah tab — `visibilitychange`/`blur` bisa berbeda antar browser) + konsistensi timer. Satu browser gagal = bug, bukan "siswa disuruh ganti browser" — kecuali sekolah sudah menetapkan satu browser resmi dan diumumkan sebelum hari-H. Matriks lulus: Chrome HP + Tablet + Desktop alur lengkap, Firefox HP + Tablet alur utama + Desktop alur lengkap; minimal Chrome HP + Chrome desktop + Firefox desktop lulus penuh, sisanya alur utama bila waktu terbatas — harus tercatat, bukan diasumsikan.

### 19.7 Checklist "Siap Ujian" (pre-flight)

- **H-7:** skenario alur + anti-curang + beban lulus di staging + bukti tersimpan; nol bug Kritis; bug Mayor punya workaround teruji; password demo sudah diganti (verifikasi login password lama ditolak); data master produksi benar dicek 2 orang; bank soal sudah di-lock (verifikasi edit ditolak); jadwal + durasi + sesi benar; token sudah di-generate + cara distribusi disepakati (dibagikan di ruang ujian, bukan disebar di grup chat sehari sebelumnya); backup produksi sudah dibuat + sudah diuji restore; server produksi dicek (disk, memori, CPU, PostgreSQL jalan, binary versi terbaru, port 9988 bisa diakses dari jaringan sekolah); rencana B/fallback disepakati tertulis (jika server mati: jeda? pindah ke kertas? ulangi sesi?).
- **H-1:** gladi bersih 5–10 siswa/panitia di ruangan + perangkat yang sama hari-H (login → token → ujian → submit → nilai, lulus penuh); perangkat dicek (browser terbaru, jam benar, baterai/charger, koneksi); daftar hadir + NIS + password siap + dicetak/dicadangkan (bukan hanya di satu HP panitia); token cadangan disiapkan; proktor dilatih (buka blokir, pause/resume, force finish, baca monitor, siapa dihubungi bila server bermasalah); jam server vs perangkat dicocokkan (selisih >1 menit diperbaiki); backup terbaru dibuat lagi setelah gladi.
- **H-Hour (60–30 menit sebelum ujian):** server hidup (login admin + peserta terbuka, waktu respons normal); jadwal hari ini berstatus benar; token hari ini valid (uji dengan 1 akun dummy, bukan akun siswa); monitor menampilkan daftar peserta yang diharapkan (jumlah cocok dengan daftar hadir); kertas cadangan (daftar token, daftar NIS, nomor kontak teknisi) ada di meja proktor; perangkat cadangan + jalur internet cadangan bila ada siap; pengumuman ke siswa (aturan anti-curang dibacakan — pindah tab 1x = peringatan, 2x = diblokir, dilarang login ganda, waktu habis = auto-submit).
- **Selama ujian:** proktor pantau monitor + catat setiap blokir + alasannya; setiap laporan "jawaban hilang / timer aneh / tidak bisa submit" dicatat (nama peserta, jam, soal ke-berapa, screenshot bila ada, jangan hanya "katanya error"); tidak ada perubahan bank soal/jadwal/token di tengah ujian tanpa persetujuan proktor utama + dicatat.
- **Setelah ujian:** semua peserta berstatus final (selesai/diblokir/belum mulai dengan alasan, tidak ada yang "menggantung"); nilai otomatis dicek kewajarannya (tidak ada nilai 0 massal / 100 massal mencurigakan tanpa penjelasan); koreksi manual diselesaikan sesuai tenggat + nilai akhir diverifikasi ulang; rekap + export CSV dibuat + dicocokkan jumlah peserta dengan daftar hadir; backup pasca-ujian dibuat + disimpan; laporan insiden ditulis (apa error, berapa siswa terdampak, bagaimana penyelesaian) untuk perbaikan ujian berikutnya.

### 19.8 Severity & triage

- **Kritis — ujian tidak boleh jalan:** peserta tidak bisa login/token saat jadwal aktif; jawaban autosave hilang setelah refresh/submit; auto-submit tidak menyimpan jawaban / nilai 0 massal; nilai akhir salah hitung; peserta A bisa lihat/ubah jawaban/nilai peserta B; peserta bisa ubah nilainya sendiri lewat manipulasi request; token bisa ditebak/dipakai lintas jadwal/tidak bisa di-revoke; server crash/error 500 massal saat 300 siswa mulai bersamaan; DB corrupt/backup tidak bisa di-restore; blokir anti-curang salah sasaran massal.
- **Mayor — mengganggu serius, ada workaround teruji:** monitor tidak refresh; export CSV berantakan untuk nama berkarakter khusus; ragu-ragu tidak bekerja di satu browser; pause tidak menghentikan timer / aturannya ambigu; upload CSV menolak 1 baris valid tanpa keterangan; peringatan pindah tab tidak muncul tetapi counter tetap bertambah; halaman koreksi lambat saat bank soal besar; token kedaluwarsa terlalu cepat.
- **Minor — kosmetik:** typo label/pesan error; tabel rekap kurang rapi di tablet tetapi terbaca; timer berkedip/tampil "59:60" sesaat; ikon/warna tidak konsisten; pesan sukses hilang terlalu cepat; urutan soal di daftar tidak terurut (isi benar); screenshot login kurang rapi di lebar jarang dipakai.
- **Triage:** Kritis → perbaiki dulu + uji ulang skenario terkait penuh; Mayor → perbaiki bila sempat, bila tidak workaround wajib sudah diuji + proktor dilatih sebelum hari-H; Minor → kumpulkan, perbaiki berkelompok setelah ujian. Setiap laporan bug wajib berisi langkah reproduksi + hasil terjadi + hasil diharapkan + akun/peran + browser + perangkat + jam + screenshot bila ada; laporan "ujian error" tanpa langkah reproduksi dikembalikan.

### 19.9 Keputusan yang dibutuhkan dari user (QA)

(1) Konfirmasi stack v2 — Go + `html/template` (sesuai README) atau Go + React (sesuai perintah user)? (menentukan revisi rencana testing). (2) Konfirmasi aturan ambigu — perilaku token kedua, perilaku pause terhadap timer peserta yang sedang ujian, perilaku unblock (mulai dari awal vs lanjut sisa waktu) — ketiganya harus satu aturan tertulis sebelum pengujian. (3) Konfirmasi utang keamanan — CSRF, fullscreen enforcement, device lock saat ini belum ada, apakah ketiganya wajib selesai sebelum produksi atau dicatat sebagai keterbatasan yang diterima? (menentukan apakah skenario terkait menjadi Kritis atau risiko yang diterima).

---

## 20. Rencana Build/Deploy (DevOps)

### 20.1 Pipeline

1. **Backend Go:** ambil kode pada versi yang ditandai (commit/tag, mis. `v2.0.0`, catat commit + tag di catatan rilis) → periksa dependensi (`go.mod`/`go.sum`, `github.com/lib/pq` + `golang.org/x/crypto`) → periksa kualitas (format + vet/test Go; build GAGAL jika ada test gagal) → build binary statis Linux dari `cmd/server` → verifikasi isi binary (mengandung `templates/*.html` + `migrations/*.sql` terbaru — cara cek paling aman: jalankan di staging + buka login admin + peserta) → beri nama artefak + checksum (mis. `cbt-go-v2.0.0-linux-amd64`) → smoke test staging (login admin, login peserta, buka ujian, simpan satu jawaban, cek monitor) → hasil akhir: **satu file binary** (±11 MB) + catatan versi, tanpa runtime Node, tanpa folder vendor, tanpa file template terpisah (sudah di dalam binary). *Catatan: untuk v2 React, verifikasi binary juga mencakup aset frontend yang di-embed.*
2. **Frontend React (HANYA jika v2 benar-benar pindah ke React):** instal dependensi dari `package-lock.json` (versi terkunci) → build aset statis (index.html + JS/CSS/asset minify + hash) → pilih cara penyajian: **Opsi B1 disarankan (embed):** aset build React di-embed ke binary Go via `go:embed` (deploy tetap satu file); Opsi B2: binary Go layani folder aset statis dari disk (`dist/`/`public/` di samping binary — deploy = binary + folder aset sebagai satu paket, versi berpasangan, jangan campur binary lama + aset baru); Opsi B3: aset disajikan web server lain/reverse proxy, Go hanya API — **tidak disarankan** untuk server lokal sekolah. Verifikasi build: halaman utama/login/ujian, tidak ada aset 404, tidak blank.
3. **Aturan rilis:** satu nomor versi + catatan perubahan (apa berubah, migrasi DB apa ikut, perlu restart?, perubahan env?); tidak ada deploy langsung dari laptop developer ke produksi tanpa artefak berversi; artefak disimpan di satu lokasi jelas (bukan hanya di satu laptop).

### 20.2 Cara jalan di server

File wajib: binary rilis; PostgreSQL 16 (skema dibuat otomatis saat start pertama, migrasi `001_schema.sql` idempotent `CREATE TABLE IF NOT EXISTS`); konfigurasi env (lihat bawah); folder backup; folder log; folder aset `dist/` hanya jika Opsi B2. **File `.env` ala Exo TIDAK dipakai** (CBT-Go tidak memakai `.env` + tidak memakai lisensi ecosystem; konfigurasi lewat env; jangan salin mekanisme lisensi Exo).

**Env yang dibaca dari kode:** `DB_PASS` (wajib, tanpa default, jika kosong koneksi gagal, jangan password demo/lemah), `DB_HOST` (default `localhost`, tetapkan eksplisit), `DB_PORT` (default `5432`), `DB_NAME` (default `cbt_go`), `DB_USER` (default `cbt_user`, buat user khusus aplikasi bukan `postgres`), `PORT` (default `9988`, tetapkan eksplisit).

**Yang TIDAK ada sebagai env:** tidak ada `SESSION_SECRET` — sesi cookie `cbt_session` token acak 32 byte di memori, 12 jam, `HttpOnly`, `SameSite=Lax` → restart aplikasi = semua sesi hangus (jadwalkan di luar jam ujian), tidak bisa 2 instance bersamaan dengan sesi saling dikenali (jalankan **satu instance saja**); `MAX_COUNTER_OUT` bukan env (di tabel `settings`, default 1, diatur dari menu Settings); tidak ada kunci lisensi (`SERVER_SECRET_LICENSE_KEY` tidak ada, tidak ada cek lisensi online); tidak ada API key AI (fitur AI belum ada; jika ditambah, API key lewat env tersendiri, tidak ditulis di laporan/chat).

**Perilaku saat start:** buka koneksi PostgreSQL (gagal → berhenti dengan pesan `db:...`, perbaiki lalu jalankan ulang) → jalankan migrasi + seeder — **PERINGATAN SEEDER:** seeder membuat akun demo bawaan; di produksi wajib ganti password/dinonaktifkan segera setelah login pertama; jangan biarkan akun demo aktif dengan password bawaan di server yang bisa diakses siswa/publik → mendengarkan di `:<PORT>` + cetak alamat.

**Prinsip jalan:** sebagai **service sistem** (bukan terminal terbuka; tetap hidup setelah logout + restart otomatis jika crash; VPS: systemd; server lokal: service setara + autostart saat boot); user sistem khusus (bukan `root`) sebisa mungkin dengan izin hanya ke binary + file env + folder log + folder backup; file env izin baca terbatas (berisi `DB_PASS`); keamanan transport: aplikasi layani HTTP polos (tidak ada TLS di dalam binary); HTTPS harus dari reverse proxy di depannya; di LAN HTTP masih dapat diterima selama jaringan terkontrol tetapi password tetap terkirim di jaringan lokal — pastikan WiFi/LAN tidak terbuka untuk umum.

### 20.3 Strategi update tanpa kehilangan data

Data ujian (jawaban, hasil, jadwal, peserta, soal) tidak pernah dihapus oleh update — update hanya ganti kode (binary), bukan data. **Aturan keras:** jangan pernah `DROP DATABASE`/`DROP TABLE`/migrasi ulang dari nol di produksi; migrasi yang ada idempotent, migrasi baru wajib aditif + idempotent; jangan pernah jalankan seeder demo di atas DB produksi setelah berisi data asli (kecuali terbukti idempotent + tidak menimpa/mereset; di produksi seeder idealnya dinonaktifkan/dibatasi instalasi pertama); jangan update saat ujian berjalan (jendela perawatan malam/di luar jam ujian + pengumuman ke proktor); setiap update didahului backup yang sudah diverifikasi.

**Langkah update:** bekukan perubahan data (tidak ada ujian berjalan, tidak ada koreksi/import berlangsung, jadwal aktif diselesaikan/dijeda dulu) → buat backup penuh (dump DB + catat versi binary yang sedang jalan + salin file env ke tempat aman, verifikasi ukuran file wajar — tidak 0 byte, uji restore di staging secara berkala) → uji rilis baru di staging dengan **salinan** DB produksi (bukan DB produksi) → ganti binary di produksi (hentikan service, ganti binary lama → baru, simpan binary lama sebagai cadangan berversi untuk rollback, jalankan kembali) → verifikasi pasca-update (service hidup, port merespons, login tampil, **jumlah data tidak berubah** — jumlah peserta/jadwal/soal/hasil sama seperti sebelum update, bandingkan dengan catatan, satu alur ujian percobaan berhasil) → umumkan selesai (beri tahu proktor/admin apa yang berubah).

**Rollback:** hentikan service, kembalikan binary lama (cadangan), jalankan kembali; **jangan restore backup DB kecuali** terbukti ada kerusakan data akibat migrasi baru (restore akan menghapus data yang masuk setelah backup); jika migrasi baru sudah mengubah skema, rollback skema harus dengan migrasi balik yang sudah disiapkan — bukan restore membabi buta; catat kejadian (versi gagal, gejala, waktu).

### 20.4 Monitoring & log

Pantau: proses aplikasi (service/binary hidup, restart berulang → cek log + DB + restart + rollback jika perlu); port & HTTP (9988 merespons, `/` + `/adm` 200); DB (PostgreSQL hidup, koneksi berhasil, query tidak macet, disk DB penuh); disk (ruang kosong partisi data DB + backup + log; >80% → hapus log/backup lama sesuai retensi, pindahkan backup, jangan tunggu 100%); CPU/RAM (beban saat ujian, jenuh → kurangi beban lain); jaringan LAN (konektivitas siswa ke server, WiFi/switch); error aplikasi (log error: gagal koneksi DB, gagal migrasi, gagal simpan jawaban, panic — berulang / saat submit = prioritas tertinggi); ujian fungsional (peserta sedang/selesai/diblokir, token aktif, jadwal aktif, lonjakan diblokir + `out_counter` naik bersamaan + token tidak berfungsi → cek monitor + tanya proktor, jangan reset massal tanpa verifikasi); waktu server (jam sinkron NTP, melenceng → jadwal/durasi/token bisa salah, sinkronkan sebelum ujian); HTTPS/sertifikat VPS (valid, belum kedaluwarsa).

**Log:** aplikasi (output standar termasuk `log.Fatal` saat DB/migrasi gagal; systemd → jurnal service; server lokal → file/jurnal tidak hilang saat terminal ditutup); service/sistem; DB (log PostgreSQL); reverse proxy VPS (akses + error HTTP). Aturan log: jangan menyimpan password/token/jawaban siswa di log; log harus memuat waktu + level + konteks. Retensi log: simpan minimal selama periode ujian + masa koreksi/sanggah, lalu arsipkan/hapus sesuai kebijakan sekolah.

**Belum tersedia:** belum ada endpoint health check khusus (`/healthz`) — pemeriksaan "server hidup" dengan membuka `/` atau `/adm` + memastikan 200; jika ditambah sebaiknya memeriksa koneksi DB juga. Monitor polling 10 detik (bukan WebSocket real-time) — untuk skala satu sekolah cukup, proktor harus me-refresh/mantau berkala. Belum ada alerting otomatis (minimal proktor memeriksa monitor sebelum + berkala selama ujian; jika memungkinkan tambah pemeriksaan berkala + notifikasi ke admin).

### 20.5 Checklist deploy hari-H (DevOps)

- **H-7 s.d. H-2:** versi ditetapkan (tag + checksum, sudah diuji di staging; tidak ada update mendadak H-1/hari-H kecuali darurat); backup terbaru di dua lokasi; semua akun demo diganti password/dinonaktifkan; akun admin/proktor/guru dibuat + sudah diuji login, password dibagikan lewat jalur aman (bukan ditempel di ruang ujian); master data lengkap (jurusan, matpel, peserta nama + nomor ujian benar, diverifikasi dengan daftar resmi); bank soal diimport/dibuat + diperiksa (jumlah soal, kunci, bobot) + **sudah dikunci (lock)**; jadwal benar (tanggal, jam mulai, durasi, sesi, bank soal benar, berstatus aktif); token sudah di-generate + diketahui cara membagikannya ke proktor (bukan dibagikan jauh sebelum waktunya); pengaturan anti-curang diperiksa (`MAX_COUNTER_OUT` sesuai kebijakan sekolah, default 1, proktor paham artinya + prosedur buka blokir yang sah); jam server sinkron NTP; server sudah diuji beban satu kelas penuh bersamaan; rencana cadangan siap (UPS, ruang cadangan, jadwal susulan); proktor sudah dilatih.
- **H-1 gladi/verifikasi akhir:** backup H-1 di dua lokasi; server di-restart sekali + terbukti hidup sendiri (service autostart) + `/` + `/adm` merespons 200; disk masih longgar, PostgreSQL hidup, log tidak menunjukkan error berulang; gladi bersih akun peserta percobaan (login → token → kerjakan soal → simpan jawaban → submit → nilai muncul di rekap; setelah gladi data percobaan dibersihkan/dipisahkan dari data asli); token hari-H siap (token gladi sudah dicabut/dibedakan); daftar hadir/kartu peserta jika dipakai sudah dicetak/siap; perangkat siswa diperiksa; kontak darurat teknis dibagikan ke proktor.
- **Hari-H pagi (30–60 menit sebelum sesi pertama):** server hidup, service jalan, `/` + `/adm` merespons; PostgreSQL hidup, disk longgar, tidak ada error baru di log; backup pagi sudah dibuat; jadwal hari ini berstatus aktif, bank soal terkunci, token hari ini aktif + belum kedaluwarsa; uji satu perangkat siswa end-to-end (login → token → buka soal → simpan satu jawaban → keluar tanpa submit / atau submit percobaan lalu bersihkan); monitor admin dibuka di perangkat proktor + terlihat peserta yang akan ujian; proktor memegang token, daftar peserta, prosedur blokir/buka blokir, kontak teknis; **tidak ada update/restart/perubahan konfigurasi setelah titik ini** sampai ujian selesai, kecuali darurat yang disetujui penanggung jawab.
- **Selama ujian:** proktor pantau monitor berkala; setiap peserta terblokir ditangani satu per satu dengan verifikasi (bukan reset massal tanpa alasan); setiap gangguan dicatat waktunya + ruang + jumlah peserta terdampak (penting untuk keputusan susulan/remedial); **tidak ada yang me-restart server / meng-update aplikasi / me-restore backup saat ujian berjalan**.
- **Setelah ujian selesai:** semua peserta berstatus selesai/terblokir sudah tercatat, jadwal diselesaikan/ditutup sesuai prosedur; backup setelah ujian langsung dibuat (dua lokasi) **sebelum** koreksi/perubahan nilai besar; rekap/hasil diexport + diperiksa (jumlah peserta yang mengerjakan = jumlah hasil, tidak ada yang hilang); koreksi manual esai/uraian dilakukan, lalu nilai akhir diverifikasi sebelum diumumkan; log/catatan gangguan selama ujian disimpan bersama backup periode ujian.

### 20.6 Keputusan yang perlu user setujui sebelum development lanjut (DevOps)

(1) Frontend — tetap template server-rendered yang sudah ada, atau benar-benar pindah ke React? Jika pindah ke React, tentukan Opsi B1 (embed) vs B2 (folder aset) — rekomendasi **B1 (embed)** agar tetap single binary. (2) Skenario deploy utama — server lokal sekolah (LAN) sebagai default, VPS hanya tryout/cadangan — dan **cPanel shared hosting tidak dipakai** untuk aplikasi CBT. (3) Tahap 2A keamanan (CSRF + wajib ganti password + nonaktifkan seed demo di produksi) diselesaikan **sebelum** deploy ke server yang bisa diakses siswa/publik. (4) Penanggung jawab operasional — siapa yang memegang backup, siapa yang boleh restart/update server, siapa yang memutuskan restore — agar hari-H tidak ada yang bertindak tanpa otorisasi.

---

## 21. Risiko + Mitigasi

| # | Risiko | Dampak | Mitigasi |
|---|---|---|---|
| R1 | Timer/submit tidak akurat (timer client-side, tidak ada enforcement deadline server) | Nilai tidak adil, sengketa | Timer berjangkar server; server tolak save/submit setelah deadline; auto-submit server + job penyapu tiap 1 menit; uji AC19–AC23 |
| R2 | Kecurangan (HP kedua, foto, kolusi, joki) | Nilai tidak valid | Batasan jujur (Bab 7.2) + pengacakan + device lock deteksi + pengawas + audit log + BAPU; jangan klaim cegah penuh |
| R3 | Lonjakan serentak (300 siswa mulai bersamaan) | Server lambat/crash, jawaban hilang | Pool DB 25–50 + indeks + uji beban + spike test (Bab 19.5); jangan turunkan ambang diam-diam |
| R4 | Kehilangan data (jawaban hilang, DB corrupt, backup tidak bisa di-restore) | Ujian harus diulang; sengketa | Autosave + antrean offline + transaksi submit + backup 2 lokasi + uji restore berkala + RPO/RTO (Bab 10.4) |
| R5 | Penskoran salah (kunci salah, aturan PGK ambigu, nilai manual tidak tercatat) | Nilai salah massal | Telaah kunci orang kedua; aturan skor per soal; nilai dihitung server; riwayat koreksi; validasi guru vs hitungan manual (Milestone D) |
| R6 | Keamanan (CSRF, IDOR, brute force, akun demo aktif, token bocor) | Akses tidak sah, nilai diubah | Tahap 2A keamanan sebelum produksi (Bab 7.5); rate limit; audit log; penegakan permission di server |
| R7 | Jaringan tidak stabil (putus, lambat, Wi-Fi lab bermasalah) | Jawaban hilang, false positive pelanggaran | Antrean offline + heartbeat + pause resmi + `out_counter` hanya saat terhubung + 1 pelanggaran gratis terkonfirmasi proktor |
| R8 | Lingkungan deploy tidak stabil (pernah terjadi: VPS di-rebuild, paket hilang, port berubah) | Deploy gagal, jadwal mundur | Skenario utama server lokal; artefak berversi + checksum; checklist H-7/H-1; jangan update mendadak H-1/hari-H |
| R9 | UKK meluas tanpa batas (variasi rubrik/sertifikat per jurusan) | Scope creep, jadwal molor | Satu format rubrik + satu format sertifikat untuk semua jurusan di v2; variasi ditunda v2.1 (perlu persetujuan, Bab 18.4) |
| R10 | Cakupan Kurikulum Merdeka berubah (CP/TP/KKTP, POS dinas) | Spesifikasi kedaluwarsa | KKTP dapat dikonfigurasi (bukan hardcode); cek POS US & POS UKK sekolah tahun berjalan + edaran TKA/AN terbaru sebelum mengunci spesifikasi |

---

## 22. Yang Ditunda ke Pasca-v2

- AI generate soal / opini koreksi AI (opini AI hanya saran, bukan nilai resmi).
- Import Google Docs (ekspor ke .docx dulu).
- Aplikasi mobile native.
- Dark mode (backlog atau dibatalkan — perlu keputusan).
- Chat 2 kolom simultan desktop (tidak relevan untuk CBT; monitor + notifikasi yang dipakai).
- Sinkronisasi offline antar-server (tidak ada di v2 — risiko jawaban ganda/timer tidak konsisten).
- Microservices, multi-database/multi-server.
- Gerbang lisensi ala Exo (tidak diadopsi).

---

## 23. Kontradiksi Antar Laporan (Belum Diputuskan)

**Jangan dianggap sudah diputuskan — semua butuh jawaban user (Bab 24):**

1. **Frontend: React vs template.** Lead Dev + Frontend Dev + Backend Dev mengasumsikan React + TypeScript + Vite + JWT. Analis Infrastruktur + QA Dev + DevOps Dev mencatat kode CBT-Go saat ini adalah Go + `html/template` server-rendered tanpa React/build step, dan Infra merekomendasikan **pertahankan template untuk v2 awal** (paling sederhana, nol build step). Perintah user untuk v2 adalah **"FRONTEND React JS"** — tetapi keputusan final (React vs pertahankan template, JWT vs session cookie) belum dikonfirmasi user.
2. **16 tipe soal:** perintah user "sesuai standard Kemendikbud 16 tipe soal" vs temuan riset "dokumen resmi hanya 5 bentuk, klaim 16 tidak ditemukan". Jangan klaim "16 tipe standar Kemendikbud" di dokumentasi/marketing v2 sampai user memutuskan label yang jujur.
3. **Peran:** DB Dev mempertanyakan apakah PROKTOR dan PENGAWAS harus jadi role terpisah (jadi 7 role)? Matriks lain memakai 6 role (admin, proktor, guru, pengawas, siswa, penguji UKK).
4. **Import Word:** Analis Bank Soal mengusulkan import Word (.docx) di v2; Lead Dev mengusulkan penundaan AI + import .docx ke pasca-v2; Researcher Benchmark mengusulkan CSV dulu, .docx fase berikutnya bila benar-benar diminta guru. Perlu diputuskan: Word di v2 atau ditunda?
5. **Palet warna & font:** Analis UX/UI mengusulkan primary #1D4ED8 + Inter; Frontend Dev mengusulkan primary #2563EB + Plus Jakarta Sans. Perlu diputuskan satu.
6. **Menjodohkan — all-or-nothing vs proporsional:** Framework AKM: 1/0 all-or-nothing per butir. Analis Bank Soal mengusulkan skor per pasangan. Perlu diputuskan mana yang dipakai (atau per-soal?).
7. **Skema PGK 2-1-0:** konflik sumber (Framework diprioritaskan — 1/0). Perlu diputuskan aturan default PGK.
8. **Beberapa asumsi yang saling berbeda antar laporan** (mis. siapa yang mengunci bank soal — guru vs admin; siapa yang generate token — proktor vs guru; apakah pengawas boleh melihat token — mayoritas: tidak; apakah Tryout boleh tanpa token; bobot UKK; perilaku pause terhadap timer; perilaku unblock) — semua dikonsolidasikan di laporan ini + daftar pertanyaan Bab 24, bukan dipilih diam-diam.
9. **Framework Go: Echo vs stdlib.** Researcher Go Backend merekomendasikan **Echo v5** (binding, route group, error handling terpusat, kompatibel penuh `net/http`); laporan lain (Lead Dev, Backend Dev) mengasumsikan stdlib `net/http` + `database/sql` (dependensi minimal, pola CBT-Go). Performa bukan penentu (selisih ~4% pada workload tanpa DB). Perlu diputuskan: Echo v5 atau stdlib + helper internal.
10. **Auth: JWT vs session cookie.** Backend Dev mengusulkan JWT access token (15 menit, di memori React) + refresh token (cookie `HttpOnly`, rotasi) + CSRF. Researcher Go Backend mengusulkan **session cookie server-side** sebagai auth utama + token ujian opaque terpisah — dengan alasan revokasi instan adalah kebutuhan fungsional CBT (peserta yang baru diblokir harus berhenti pada request berikutnya, bukan setelah token kedaluwarsa). Researcher React Frontend mengusulkan cookie `HttpOnly` + CSRF protection (sejalan dengan session cookie). Perlu diputuskan satu pola auth — keputusan ini memengaruhi Bab 15.4, struktur `internal/auth/`, dan handling 401/refresh di frontend.
11. **Hash password: bcrypt vs Argon2id.** Backend Dev mengusulkan bcrypt. Researcher Go Backend mengusulkan **Argon2id** (rekomendasi utama OWASP saat ini; bcrypt hanya fallback/legacy — limit input 72 byte, tidak memory-hard). Perlu diputuskan satu algoritma (dan parameternya — dengan catatan uji beban login serentak agar Argon2id tidak menghabiskan memori server).
12. **Akses DB: `database/sql` + `lib/pq` vs `pgx` v5 + `sqlc`.** Laporan awal menyebut `database/sql` + PostgreSQL (pola CBT-Go). Researcher Go Backend mengusulkan **`pgx` v5 (`pgxpool`) + `sqlc`** (SQL eksplisit bertipe, `COPY` untuk import massal, `LISTEN/NOTIFY`; **GORM ditolak** sebagai lapisan utama). Perlu diputuskan satu pola akses DB.
13. **Migrasi DB: file SQL ad-hoc vs `golang-migrate`.** Laporan awal menyebut migrasi bernomor urut (`001_schema.sql`, lalu `002_...`, `003_...`). Researcher Go Backend mengusulkan **`golang-migrate`** (file SQL `up`/`down` berversi) — konsisten dengan migrasi bernomor, tetapi perlu diputuskan tool-nya (alternatif yang masih dapat diterima: `goose`). Manual (SQL dijalankan tangan) tidak direkomendasikan.
14. **KOREKSI FAKTA terhadap klaim di laporan awal — dari Benchmark (verifikasi kode CBT-Go):**
    - **Pause CBT-Go TIDAK membekukan timer.** Klaim awal "pause jadwal membekukan timer semua siswa (asumsi)" adalah asumsi untuk v2, bukan perilaku CBT-Go. Di CBT-Go, pause hanya gerbang masuk (ujian yang di-pause menampilkan halaman "di-pause"), sisa waktu dihitung dari `mulai_ujian` — jadi pause tidak menambah waktu. **v2 harus membekukan timer** (simpan `sisa_waktu` saat pause, lanjutkan dari nilai itu saat resume) — cacat ini wajib diperbaiki, bukan ditiru.
    - **Token expiry CBT-Go = akhir hari jadwal (23:59:59), bukan 900 detik.** Setting `TOKEN_EXPIRED=900` tersimpan di DB tapi tidak dirujuk kode yang dibaca. Untuk v2, expiry token harus diputuskan eksplisit (Bab 24).
    - **CBT-Go hanya 8 tipe yang auto-grade; UI ujian hanya 3 jalur render** (radio, checkbox, textarea). Menjodohkan/mengurutkan/studi kasus dll. hanya textarea bebas; tipe stimulus 9–11 tampil sebagai radio biasa tanpa gambar/tabel/grafik. Klaim "16 tipe" di CBT-Go sebagian nominal — v2 hanya boleh mengklaim tipe yang benar-benar punya input terstruktur + aturan penilaian.
    - **CBT-Go tidak punya edit** untuk master data, user, dan soal (hanya tambah + hapus) — v2 wajib full CRUD termasuk edit (P0, Bab 13.2).
    - **CBT SMK (Laravel): anti-curang hanya pencatatan** (pindah tab hanya dicatat di log, tanpa konsekuensi); token tanpa expiry/revoke, satu per paket; tanpa pause/resume; tanpa halaman monitor. Ini menguatkan prioritas P1 Bab 13.2.
15. **Monitor: polling vs WebSocket.** CBT-Go memakai polling refresh 10 detik; Lead Dev mengusulkan WebSocket (pengganti polling); Researcher Go Backend mengusulkan `coder/websocket` + pola Hub dengan fallback polling REST bila handshake gagal; Researcher Benchmark menilai polling 10 detik cukup untuk v2, WebSocket bisa menyusul. Perlu diputuskan: WebSocket sejak v2.0 atau polling dulu + WebSocket menyusul.
16. **Logging: `log` standar vs `log/slog`.** Researcher Go Backend mengusulkan `log/slog` JSON (stdlib Go 1.21+); `logrus` tidak direkomendasikan untuk proyek baru. Perlu diputuskan bersama tooling observabilitas.
17. **Struktur paket Go: `internal/{config,auth,rbac,model,store,service,handler,...}` (Lead Dev) vs `internal/{config,app,transport,domain,service,repository,auth,upload,migration,observability,validator}` (Researcher Go Backend — organisasi per fitur/domain, framework hanya di transport).** Keduanya monorepo + lapisan terpisah; yang kedua lebih ketat (domain murni, tanpa import framework/driver DB). Perlu diputuskan satu struktur sebelum Sprint 0.
18. **State management React: Zustand + TanStack Query (Researcher React Frontend) vs deskripsi umum "state ujian di backend + store khusus" (laporan sebelumnya).** Researcher React Frontend memberi rekomendasi spesifik: Zustand (state ujian, dengan selector) + TanStack Query (server state) + Context (yang jarang berubah) + `useState` lokal; Redux Toolkit ditolak (overkill); Context ditolak untuk state ujian (tidak ada selective subscription — setiap tick timer me-render ulang seluruh halaman). Perlu diputuskan: setuju rekomendasi ini?
19. **UI: Tailwind + shadcn/ui selektif + komponen ujian custom (Researcher React Frontend) vs "Tailwind v4 build lokal, NOL CDN" (laporan sebelumnya).** Konsisten secara umum, tetapi perlu diputuskan: setuju shadcn/ui selektif (komponen generik yang sulit dibuat aksesibel) + komponen inti ujian custom? MUI / Ant Design ditolak untuk v2.
20. **HTTP client: `fetch` + wrapper `apiClient` sendiri + TanStack Query (Researcher React Frontend) vs tidak ditentukan sebelumnya.** Axios tidak diperlukan untuk proyek baru ini. Perlu diputuskan: setuju `fetch` + wrapper?
21. **Upload: batas dan pola.** Researcher Go Backend mengusulkan batas import soal 5–10 MB, UKK 10–25 MB per file / 100 MB per peserta per ujian praktik, foto/dokumen 5–10 MB, maks 2.000 soal per file import, streaming ke disk, validasi magic bytes, nama acak UUID, hash SHA-256, serve via handler terautentikasi. Analis Bank Soal mengusulkan batas media gambar 2 MB, audio 5 MB, video 20 MB. Perlu diputuskan satu tabel batas upload (dan apakah batas media soal = batas upload UKK atau dibedakan).
22. **Backup/restore: dari dalam aplikasi vs dari luar.** Researcher Benchmark menilai backup/restore dari dalam aplikasi sebagai P3 (fase kedua). Researcher DB memberi desain backup lengkap (pg_dump harian + off-site mingguan + tes restore bulanan + sebelum musim ujian besar). DevOps memberi prosedur backup dari luar aplikasi. Perlu diputuskan: backup/restore dari dalam aplikasi ikut v2.0 (P3) atau hanya prosedur luar aplikasi + tombol backup menyusul?

## 24. Daftar Pertanyaan untuk User (Wajib Dijawab Sebelum Development)

### A. Keputusan arsitektur & stack (paling mendesak)

1. **Frontend:** setuju pindah ke **React + TypeScript + Vite** (sesuai perintah awal), atau pertahankan **template Go server-rendered** (lebih sederhana, nol build step)? Jika React, setuju **Opsi B1 (aset di-embed ke binary Go)** agar tetap single binary?
2. **Auth:** setuju pindah session cookie → **JWT access token + refresh cookie + CSRF** (lebih cocok SPA + mobile), atau pertahankan session cookie (lebih sederhana)?
3. **Monorepo + single binary Go** (frontend di-embed) vs 2 repo — setuju monorepo?
4. **Skenario deploy utama:** setuju **server lokal sekolah (LAN)** sebagai default, VPS hanya tryout/cadangan, dan **cPanel shared hosting tidak dipakai** untuk aplikasi CBT?

### B. Tipe soal & label jujur

5. **Label 16 tipe:** setuju label jujur **"5 tipe standar AKM + 11 varian/ekstensi aplikasi"** (jangan klaim "16 tipe standar Kemendikbud"), atau ingin tetap memakai label "16 tipe standar Kemendikbud" dengan catatan? (Riset: klaim 16 tidak ditemukan di dokumen resmi.)
6. **Menjodohkan:** penilaian **all-or-nothing per butir** (Framework AKM) atau **proporsional per pasangan**?
7. **PG Kompleks:** aturan default **proporsional** atau **semua-benarnya-saja (1/0)**? (Aturan disimpan per soal — tetapi default global perlu diputuskan.)
8. **Bobot per soal, konversi ke 0–100, pengacakan:** Framework tidak mengaturnya — setuju aturan v2 (bobot per soal, nilai akhir = Σ poin / Σ bobot × 100)?

### C. Roles & permission

9. **Jumlah role:** 6 role (admin, proktor, guru, pengawas, siswa, penguji UKK) atau 7 role (proktor dan pengawas dipisah lebih jauh)? Setuju matriks Bab 11 (khususnya: pengawas **tidak boleh melihat token**; proktor **tidak boleh lihat isi soal**)?
10. **Pemisahan Tugas (SoD):** setuju 7 aturan SoD Bab 11.4 + pengecualian darurat (alasan tertulis + audit + flag `SOD_OVERRIDE` + notifikasi kepala sekolah/admin kedua)?
11. **Satu sesi = satu role aktif** — setuju?

### D. Aturan ujian (harus satu aturan tertulis sebelum pengujian)

12. **Pause:** apakah pause jadwal membekukan timer semua siswa? (Usulan: ya.)
13. **Unblock:** setelah di-unblock, siswa **lanjut sisa waktu** atau **reset (hapus jawaban, ulang dari awal)**? (Harus satu aturan — tidak boleh dua perilaku berbeda.)
14. **Token kedua:** bila token di-generate kedua kalinya untuk jadwal yang sama, apa yang terjadi dengan token pertama? (Berlaku keduanya / pertama hangus / ditolak?)
15. **Token expiry:** setuju `token_expiry_detik` default 900 (15 menit)? Token dirilis berapa menit sebelum mulai?
16. **Toleransi keterlambatan:** berapa menit? (Contoh TKA/AN: 15 menit.) Siswa telat: sisa waktu = jam selesai − waktu masuk (tanpa tambahan) — setuju?
17. **Tryout:** boleh tanpa token? (Usulan: ya.) Pembahasan boleh langsung tampil? (Usulan: ya.)
18. **Nilai esai sebelum finalisasi:** siswa boleh lihat nilai otomatis sementara langsung setelah submit, atau disembunyikan total sampai koreksi final? (Desain memakai setting `tampil_hasil_ke_siswa`, default sembunyikan sampai final — setuju?)

### E. Anti-curang & keamanan

19. **Jumlah pelanggaran sebelum auto-block:** setuju `MAX_COUNTER_OUT = 1` (1x peringatan, 2x auto-block)? Device lock ditegakkan sejak hari pertama? Fullscreen di v2.0 atau ditunda?
20. **CSRF + wajib ganti password + nonaktifkan seed demo** — setuju diselesaikan **sebelum** deploy ke server yang bisa diakses siswa/publik? (Rekomendasi: ya — ini Kritis.)
21. **Device lock / fullscreen / CSRF** — apakah ketiganya wajib selesai sebelum produksi, atau dicatat sebagai keterbatasan yang diterima? (Menentukan apakah skenario QA terkait menjadi Kritis atau risiko yang diterima.)
22. **Pengecualian darurat SoD** — setuju mekanisme Bab 11.4?

### F. Bank soal, kurikulum, UKK

23. **Kisi-kisi:** setuju model kisi-kisi Bab 5.3 (satu baris = satu kelompok soal sejenis; soal wajib merujuk ke satu baris kisi-kisi + satu kompetensi utama)?
24. **KKTP:** setuju KKTP dapat dikonfigurasi per mapel/TP (deskripsi/rubrik/interval), bukan hardcode KKM=75? Interval default mana yang dipakai sekolah?
25. **Remedial:** aturan gabungan nilai remedial — nilai remedial menggantikan / tidak menggantikan / rata-rata dengan nilai awal? (Tidak ada rumus nasional — kebijakan sekolah.)
26. **UKK v2:** setuju satu format rubrik + satu format sertifikat untuk semua jurusan (variasi ditunda v2.1)? Bobot UKK default teori 40% + praktik 60% — setuju? Bobot rubrik praktik Proses 30% / Produk 50% / Sikap 20% — setuju? Ambang kelulusan per unit (bukan hanya total) — setuju?
27. **Import Word (.docx):** ikut v2 atau ditunda pasca-v2? (Kontradiksi Bab 23 butir 4.) Import Google Docs — setuju ditunda?
28. **Seed demo di produksi:** setuju seed produksi hanya role + permission + settings default (tanpa akun demo)? Siapa yang memegang backup, siapa yang boleh restart/update server, siapa yang memutuskan restore?
29. **Daftar jurusan / mapel / TP/ATP sekolah:** apakah sekolah sudah punya TP/ATP per mapel? (Jika belum, pembuatan kisi-kisi harus didahului penyusunan TP/ATP — CBT tidak boleh mengarang CP.) Minta salinan POS US & POS UKK sekolah tahun berjalan + edaran TKA/AN terbaru + kebijakan retensi data & persetujuan orang tua.

### G. UI & fitur

30. **Palet warna final:** #1D4ED8 (biru royal) atau #2563EB (biru indigo)? **Font final:** Inter atau Plus Jakarta Sans?
31. **Analisis Butir Soal:** ikut v2 atau ditunda? (Rekomendasi: ikut v2 — ini keunggulan utama vs CBT-Go.)
32. **Dark mode:** backlog atau dibatalkan?
33. **Route:** tetap pola `/adm/*` + `/*` seperti v1? (Di usulkan: ya.)
34. **Export:** cukup CSV tahap 1, atau perlu XLSX juga?
35. **Upload UKK:** siswa boleh upload berkas praktik sendiri, atau hanya penguji? (Desain mengizinkan keduanya, siswa hanya dirinya sendiri — setuju?)
36. **WebSocket:** ticket 30 detik, atau access token lewat query param WS (lebih sederhana tapi terekspos di log)?

### I. Keputusan stack backend & frontend (tambahan dari laporan detail)

41. **Framework Go:** setuju **Echo v5** (binding, route group, error handling terpusat, kompatibel penuh `net/http`) atau **stdlib `net/http` + helper internal** (nol dependensi, pola CBT-Go)? (Performa bukan penentu — selisih ~4% pada workload tanpa DB.)
42. **Auth:** setuju **session cookie server-side** (revokasi instan — peserta yang baru diblokir berhenti pada request berikutnya; daftar sesi aktif bisa diaudit) atau **JWT access token + refresh cookie** (lebih cocok SPA + mobile, tetapi revokasi sulit)? Jika JWT, bagaimana revokasi instan untuk peserta yang diblokir (denylist per `jti` / rotasi kunci / access token sangat pendek)? Jika session cookie, setuju token ujian opaque terpisah (per-peserta-per-jadwal, satu attempt aktif)?
43. **Hash password:** **Argon2id** (rekomendasi utama OWASP; parameter: memori 19 MiB minimum / 64 MiB bila server cukup RAM, iterasi 2–3, paralelisme 1–2 — dengan uji beban login serentak) atau **bcrypt** (cost 12+)? (Bcrypt limit input 72 byte, tidak memory-hard.)
44. **Akses DB:** **`pgx` v5 (`pgxpool`) + `sqlc`** (SQL eksplisit bertipe, `COPY` untuk import massal, `LISTEN/NOTIFY`; GORM ditolak sebagai lapisan utama) atau **`database/sql` + `lib/pq`** (pola CBT-Go)? Ukuran pool: setuju `MaxConns` 20–30 untuk server sekolah?
45. **Migrasi DB:** setuju **`golang-migrate`** (file SQL `up`/`down` berversi) atau `goose`? Setuju aturan: satu tool per repository, migrasi immutable, jangan auto-migrate tanpa pengaman di produksi (perintah migrasi eksplisit / advisory lock)?
46. **WebSocket:** setuju **`coder/websocket` + pola Hub** (room per jadwal/ruangan, otorisasi room saat subscribe, heartbeat, backpressure/slow consumer, reconnect + snapshot, fallback polling REST bila handshake gagal) sejak v2.0, atau polling 10 detik dulu + WebSocket menyusul? (Gorilla/websocket ditolak — repository diarsipkan sejak 2022.) Setuju fallback polling REST bila WebSocket diblokir proxy sekolah?
47. **Struktur paket Go:** setuju struktur **`cmd/` tipis + `internal/` per fitur/domain** (transport / domain / service / repository / auth / upload / migration / observability / validator; framework hanya di transport; domain murni tanpa import framework/driver DB) atau struktur `internal/{config,auth,rbac,model,store,service,handler,...}` (Lead Dev)?
48. **Logging:** setuju **`log/slog` JSON** (stdlib Go 1.21+, handler JSON produksi + teks development; log audit terpisah ke database, append-only) atau `log` standar? (Logrus tidak direkomendasikan.)
49. **Stack frontend:** setuju **Vite 8 + React 19 + TypeScript + React Router v7 + Vitest + Testing Library**? (Next.js dan CRA ditolak — Next.js menuntut server Node.js yang hidup tanpa manfaat SSR karena CBT di balik login; CRA deprecated.)
50. **State management:** setuju **Zustand (state ujian, dengan selector) + TanStack Query (server state) + Context (yang jarang berubah) + `useState` lokal**? (Redux Toolkit ditolak — overkill; Context ditolak untuk state ujian — tidak ada selective subscription.)
51. **UI:** setuju **Tailwind CSS 4 + shadcn/ui selektif** (komponen generik yang sulit dibuat aksesibel: Button, Dialog, Dropdown, Tooltip, Toast, Sheet/Drawer, Progress, Badge, Tabs) **+ komponen inti ujian custom** (kartu soal, opsi jawaban, grid navigasi, timer, panel ragu-ragu)? MUI / Ant Design ditolak untuk v2 — setuju?
52. **HTTP client:** setuju **`fetch` + wrapper `apiClient` sendiri + TanStack Query** (satu modul `apiClient`, satu modul `api` per domain, timeout eksplisit per jenis request, pembatalan via `AbortSignal`)? Axios tidak diperlukan — setuju?
53. **Penyajian frontend:** setuju **embed ke binary Go** (satu binary + database; setiap perubahan frontend mengharuskan rebuild binary Go) atau sajikan dari folder `dist/` di disk (frontend bisa diperbarui tanpa rebuild Go, tetapi harus memastikan versi `dist/` cocok dengan backend)? Setuju aturan fallback SPA (API dulu, catch-all terakhir; `/api/*` tidak pernah fallback ke `index.html`; `index.html` no-cache, aset ber-hash immutable) dan root (`/`) sebagai base path?
54. **Upload:** setuju tabel batas upload (import soal 5–10 MB, UKK 10–25 MB per file / 100 MB per peserta per ujian praktik, foto/dokumen 5–10 MB, maks 2.000 soal per file import) atau batas media soal (gambar 2 MB, audio 5 MB, video 20 MB)? Setuju validasi berlapis (ukuran + ekstensi whitelist + sniffing magic bytes + validasi struktur + antivirus bila ada + nama acak UUID + hash SHA-256 + serve via handler terautentikasi, tidak pernah `inline` HTML/JS, tidak pernah dieksekusi)? Setuju penyimpanan filesystem lokal + metadata Postgres (object storage hanya bila multi-server)?
55. **HTTPS di server lokal:** apakah server sekolah akan memakai **HTTPS** (agar cookie `Secure` bisa aktif)? Jika tidak, setuju mendokumentasikan penurunan keamanan tersebut secara eksplisit (dan jangan pernah membuka server HTTP murni ke internet)? (Menentukan apakah cookie `Secure` bisa aktif dan apakah `SameSite=Lax` + CSRF cukup.)
56. **Session store:** cukup di **PostgreSQL** (satu server) atau perlu Redis? (Rekomendasi: PostgreSQL saja untuk v2.)
57. **Batas upload UKK final per jenis praktik:** video demo praktik membutuhkan batas jauh lebih besar daripada dokumen — angka di atas adalah titik awal, bukan keputusan final. Berapa batas final per jenis praktik?
58. **Kebijakan anti-curang final (sebelum Hub WebSocket diimplementasikan):** berapa pelanggaran (pindah tab, keluar fullscreen) sebelum auto-block, dan apakah pengawas dapat mengubah ambang per jadwal? (Menentukan implementasi Hub + aturan pelanggaran.)
59. **Backup/restore dari dalam aplikasi:** ikut v2.0 atau fase kedua (hanya prosedur luar aplikasi dulu)? Setuju jadwal backup (harian di luar jam ujian + sebelum perubahan berisiko + file harian + off-site mingguan + arsip akhir semester), rotasi (harian 14–30 hari, mingguan 8–12 minggu), dan tes restore (minimal 1x per bulan + 1x sebelum musim ujian besar, ke server terpisah, dengan pemeriksaan kesehatan + verifikasi jumlah baris)?
60. **Retensi data:** setuju retensi (log audit minimal 1 tahun ajaran, idealnya 2 tahun; peristiwa anti-curang selama hasil ujian terkait masih disimpan; jawaban & nilai hingga akhir tahun ajaran + masa sanggah; log teknis 30–90 hari; backup sesuai rotasi)? Siapa yang menjalankan penghapusan karena retensi, dan bagaimana pencatatannya?

### H. Operasional

37. **Penanggung jawab operasional:** siapa (nama + peran) yang memegang backup, siapa yang boleh restart/update server, siapa yang memutuskan restore?
38. **Retensi data:** berapa lama menyimpan data/nilai ujian? (Tidak ada ketentuan nasional — kebijakan sekolah + UU PDP.) Bagaimana persetujuan orang tua/wali didokumentasikan?
39. **Skala:** berapa siswa serentak maksimal yang harus didukung hari-H? (Asumsi desain: 300.) Berapa lab/ruangan? Perangkat siswa (HP/tablet/laptop lab) — campuran apa?
40. **Jadwal development:** setuju estimasi ±73–79 hari kerja (≈15–16 minggu) + milestone A–E (Bab 18)? Kapan UAT di sekolah dijadwalkan?

---

## 25. Sumber

### Dokumen resmi & regulasi

1. Framework AKM (Pusat Asesmen Pendidikan, BSKAP, Kemendikbudristek, edisi revisi 2023, Lampiran 1 hlm. 147–149) — https://repositori.kemendikdasmen.go.id/33375/1/FRAMEWORK%20ASESMEN%20KOMPETENSI%20MINIMUM.pdf
2. Kepmen Dikdasmen No. 56/2026, Pedoman Penyelenggaraan TKA & AN — https://pusmendik.kemendikdasmen.go.id/pdf/file-201
3. GTK Dikdasmen, "Menyingkap Asesmen Nasional" — https://gtk.dikdasmen.go.id/read-news/menyingkap-asesmen-nasional
4. Siaran Pers Kemendikdasmen, "Diseminasi Tes Kemampuan Akademik" — https://internal-portal.kemdikbud.go.id/siaran-pers/13887-diseminasi-tes-kemampuan-akademik-membangun-integritas-untuk
5. POS UKK SMKS Amir Hamzah 1 BM TP 2024/2025 — https://id.scribd.com/document/875219322/628151601-POS-UKK-SMK
6. POS USBN 2018/2019 (BSNP, historis) — https://repositori.kemendikdasmen.go.id/8856/1/(0048) POS USBN 2018-2019 versi 29 nov final.pdf
7. Ringkasan Permendikbudristek 21/2022 — https://www.ainamulyana.com/2022/05/permendikbud-ristek-nomor-21-tahun-2022.html
8. Ringkasan KKTP / Panduan Pembelajaran & Asesmen — https://minorrahman.sch.id/blog/analisis-capaian-pembelajaran-tujuan-pembelajaran-alur-tujuan-pembelajaran-dan-kriteria-ketercapaian-tujuan/
9. Contoh KKTP sekolah — https://www.min1kotabaru.sch.id/download/file/KKTP_Pengganti_KKM_pd_Kumer.pdf
10. Penjelasan KKM K13 — https://www.quipper.com/id/blog/info-guru/kriteria-ketuntasan-minimal/
11. Teks UU 27/2022 (DPR) — https://berkas.dpr.go.id/jdih/document/uu/1814.pdf
12. Transkrip UU 27/2022 (Wikisource) — https://id.wikisource.org/wiki/Halaman:UU_Nomor_27_Tahun_2022_Tentang_Perlindungan_Data_Pribadi.pdf/1
13. Transkrip Pasal 20–25 UU PDP (sekunder) — https://github.com/zakyrmh/robotik-pnp/blob/HEAD/docs/07-compliance-security/regulations/uu-pdp-27-2022.md
14. Sekolah sebagai pengendali data (UU PDP) — https://www.slbbiromaru.sch.id/urgensi-perlindungan-data-pribadi-di-era-digitalisasi-pendidikan-tanggung-jawab-pengelola-data-di-sekolah
15. Perlindungan data anak (UU PDP) — https://fh.untar.ac.id/2025/09/11/perlindungan-data-pribadi-implementasi-uu-no-27-tahun-2022-dan-tantangan-penegakannya/
16. UKK & Permendikbud 34/2018 (konteks) — https://www.smkntanjungsari.sch.id/blog/2021/02/uji-kompetensi-keahlian-tahun-pelajaran-2020-2021/
17. 6 skema UKK — https://akademikombas.co.id/ukk-smk-kabupaten-temanggung/
18. Contoh POS ANBK sekolah (token/berita acara) — https://id.scribd.com/document/539940033/POS-ANBK-SD-1-TANJUNGREJO

### Referensi internal (workspace)

- `~/workspace/cbt-go/` — README + kode CBT-Go v1 (arsitektur, 16 tipe, anti-curang, README).
- `~/workspace/cbt-go/TAHAP-2-RENCANA.md` dan `~/workspace/cbt-go/ANALISIS-BUTIR-SPESIFIKASI.md` — spesifikasi analisis butir soal.
- `~/workspace/extraordinary-cbt/PELAJARI-EXTRAORDINARY-CBT.md` — kajian Extraordinary CBT v4.7.0.
- `~/workspace/cbt-smk/` — CBT Laravel (referensi UKK, berita acara).
- `~/workspace/cbt-smk-v2/laporan-database-arsitektur.md` — laporan lengkap Researcher Database & Arsitektur (987 baris, detail tabel/indeks/migrasi/backup/keamanan).

---

**Penutup:** Laporan ini adalah hasil sintesis 20 laporan worker. Semua asumsi ditandai sebagai asumsi. Semua yang tidak ditemukan sumbernya dinyatakan tidak ditemukan — bukan diasumsikan. **Tidak ada kode aplikasi yang ditulis dalam fase ini.** Development dimulai setelah user menjawab Daftar Pertanyaan (Bab 24) dan menyetujui laporan ini.
