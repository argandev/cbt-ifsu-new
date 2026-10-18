# E2E HALAMAN UJIAN SISWA — CBT SMK v2

**Tim:** E2E Test (5 orang).
**Tanggal:** 17 Okt 2026.
**Ruang lingkup:** halaman ujian (`/ujian/:attemptId`) + `HasilPage`, hasil perbaikan
Tim Coding atas `docs/AUDIT-HALAMAN-UJIAN.md` (prinsip P-01 s.d. P-12).
**Cara uji:** browser Chromium headless (Playwright-core + chromium cache)
terhadap server lokal — backend Go di `:9988` + Vite dev di `:5173`
(kode `src` terbaru, BUKAN exe). Klik/tombol/Select/dropdown diuji lewat
browser sungguhan, bukan hanya screenshot.
**File yang diubah tim E2E:** TIDAK ADA file sumber yang diubah. Satu-satunya
file baru di repo: dokumen ini + 17 screenshot `e2e-*` di
`screenshots/qa-ujian/`. (`npm run build` yang wajib dijalankan
membangun ulang `frontend/dist` — artefak build, bukan sumber.)
**Data uji:** siswa uji terpisah (bukan data produksi): Siswa Uji lama
(`1234567890`) + 8 siswa E2E (`900000001`–`900000008`, password sama
`1234567890`) + 1 soal PG Kompleks baru (urutan 6) di paket uji
`PKT-TEST`. Setelah uji selesai, SEMUA data uji E2E dihapus dan jam
jadwal dikembalikan ke nilai semula (lihat §7).

## 1. Ringkasan hasil

| # | Item | Hasil |
|---|---|---|
| 1 | Alur ujian penuh (login → mulai → 6 tipe soal → ragu → navigasi → submit → Hasil) | **LOLOS** |
| 2 | Bar bawah 1 kotak tunggal di 360/768/1280/1440, tanpa meluber/tumpuk, konten tidak tertutup, tanpa overflow horizontal | **LOLOS** |
| 3 | Regresi grading Menjodohkan (6 kasus, payload `{kiri: kanan}` identik, kontrol tetap antd `Select`) | **LOLOS 6/6** |
| 4 | Regresi lama: timer, autosave, anti-curang, ragu-ragu, modal nomor belum dijawab, kartu PDF tidak tersentuh | **LOLOS** |
| 5 | HasilPage: Menjodohkan per pasangan, Poin sebagai Tag, gaya konsisten | **LOLOS** |
| 6 | `npx tsc --noEmit` + `npm run build` + `go test ./...` | **LOLOS** |
| — | Bug baru yang ditemukan di file-file scope | **NIHIL (0)** |
| — | Observasi di LUAR scope (grading Benar-Salah data lama) | 1, lihat §6 |

## 2. Alur ujian penuh end-to-end (Siswa Uji, attempt `221ca414-…`)

| Langkah | Hasil |
|---|---|
| Login siswa `1234567890` | LOLOS — redirect ke Beranda, jadwal tampil |
| Mulai ujian dari Beranda ("Masuk Ujian", Tryout, tanpa token) | LOLOS — attempt baru, URL `/ujian/<id>`, "Soal 1 dari 6" |
| Soal 1 PG — pilih A | LOLOS — radio tercentang, status "Tersimpan", DB: jawaban `"A"`, poin 1.00 |
| Berikutnya → Soal 2 Benar-Salah — pilih Benar, Benar | LOLOS — DB: `[true, true]` (poin 0.00 — lihat observasi §6, bukan cacat halaman ujian) |
| Berikutnya → Soal 3 Isian — ketik "Bandung" | LOLOS — DB: `"Bandung"`, poin 1.00 |
| Berikutnya → Soal 4 Menjodohkan — pilih Tokyo/Paris/Jakarta via antd Select | LOLOS — native `<select>` = 0, `.ant-select` = 3, panah `→` = 3 (selalu terlihat); DB: peta lengkap, poin 2.00 |
| Berikutnya → Soal 5 Uraian — ketik jawaban | LOLOS — DB tersimpan, menunggu koreksi |
| Berikutnya → Soal 6 PG Kompleks — centang A + B | LOLOS — DB: `["A","B"]`, poin 2.00 |
| Ragu-ragu: tandai di Soal 6 → nav berlabel ragu = 1; lepas → 0; tandai di Soal 1 | LOLOS — DB: `ragu=true` hanya Soal 1 |
| Navigasi: klik nomor 4 → Soal 4; Sebelumnya → Soal 3; Berikutnya → Soal 4 | LOLOS |
| Modal konfirmasi saat 2 soal belum dijawab | LOLOS — teks "Dijawab: 4 dari 6 soal. Masih ada 2 soal belum dijawab (dihitung 0). Soal belum dijawab: Soal 5, Soal 6"; Batal → tetap di ujian |
| Modal konfirmasi saat semua dijawab + "Ya, Kumpulkan" | LOLOS — "Dijawab: 6 dari 6 soal. Semua soal sudah dijawab." |
| Submit | LOLOS — layar "Ujian Selesai", status attempt `selesai`, nilai akhir 66.67 (= 6/9 bobot) |
| HasilPage (`/hasil/<id>`) | LOLOS — nilai 66.67, lihat §5 |

Catatan artefak skrip (BUKAN bug aplikasi): dua kegagalan klik pertama
disebabkan locator skrip yang salah — `text=A.` (tidak memicu Radio) dan
`button[aria-label="Soal 1"]` persis (label sebenarnya "Soal 1 (dijawab)").
Setelah locator diperbaiki (klik Radio via role; klik nomor via substring),
semua langkah lolos. Kegagalan keyboard `Enter` pada Select yang sudah
berisi nilai (KASUS 5 pertama) juga artefak skrip — antd mengaktifkan
opsi nilai saat ini, sehingga Enter memilih ulang nilai lama; klik mouse
pada opsi (cara pengguna nyata) mengubah nilai dengan benar (KASUS 5 redo).

## 3. Bar bawah 1 kotak tunggal (P-01, P-02, D-01, D-02)

Diukur lewat DOM + dibaca visual dari screenshot. Di semua lebar, elemen
sticky bawah HANYA 1 (yang kedua adalah header sticky atas).

| Lebar | overflow-X | Bar: tinggi × lebar | Isi bar | Konten tertutup? |
|---|---|---|---|---|
| 360 | tidak (`scrollWidth=360`) | 173 × 328 px (3 baris dalam 1 kotak) | Sebelumnya/Berikutnya; ragu-ragu; Selesai | Tidak — spacer + ruang kosong di atas bar terlihat di screenshot |
| 768 | tidak (`scrollWidth=768`) | 69 × 736 px (1 baris) | 4 aksi sebaris, `justify-between` | Tidak |
| 1280 | tidak (`scrollWidth=1280`) | 69 × 1068 px (1 baris, = lebar shell 1100 − padding) | 4 aksi sebaris | Tidak — kartu Navigasi berakhir di legenda (tidak stretch), bottom nav di atas bar |
| 1440 | tidak (`scrollWidth=1440`) | 69 × 1068 px (1 baris, = lebar shell) | 4 aksi sebaris | Tidak |

Screenshot: `e2e-ujian-{360,768,1280,1440}.png` (full page),
`e2e-ujian-bawah-{360,768,1280,1440}.png` (posisi scroll bawah),
`e2e-ujian-modal-{360,768,1280,1440}.png` (modal konfirmasi per lebar),
`e2e-ujian-drawer-{360,768}.png` (Drawer navigasi).
Verifikasi visual: bar 360 = satu kotak (Sebelumnya/Berikutnya baris 1,
ragu-ragu baris 2, Selesai baris 3); bar 1280 = satu kotak satu baris;
tidak ada kotak kedua, tidak ada tumpukan, tidak ada elemen meluber.

Drawer HP/tablet (P-03): tombol "Daftar Soal" + badge belum dijawab (5)
membuka Drawer; klik nomor di Drawer → pindah soal + Drawer menutup
(`.ant-drawer-open` = 0). LOLOS di 360 dan 768.

## 4. Regresi grading Menjodohkan (6 kasus)

Kontrol jawaban: antd `Select` (native `<select>` = 0 di semua kasus).
Payload jawaban yang ditangkap dari request `POST /jawaban` berbentuk
`{kiri: kanan}`. Poin dibaca dari kolom `jawaban.poin` per attempt.
Bobot soal Menjodohkan = 2, 3 pasangan → tiap pasangan benar = 0,67
(proporsional; 2 × n/3, pembulatan backend).

| Kasus | Siswa | Payload akhir (atau keterangan) | Poin | Harapan | Hasil |
|---|---|---|---|---|---|
| 1. Semua benar | E2E-1 | `{Jepang: Tokyo, Prancis: Paris, Indonesia: Jakarta}` | 2.00 | 2.00 | LOLOS |
| 2. Semua kosong | E2E-2 | tidak ada request jawaban Menjodohkan; tidak ada baris jawaban (dihitung 0) | — (0) | 0 | LOLOS |
| 3. Parsial (1/3 benar) | E2E-3 | `{Jepang: Tokyo}` (pasangan lain tidak dikirim — perilaku pra-perbaikan: key hanya untuk pasangan yang dipilih) | 0.67 | 0.67 | LOLOS |
| 4. Semua salah | E2E-4 | `{Jepang: Paris, Prancis: Jakarta, Indonesia: Tokyo}` | 0.00 | 0 | LOLOS |
| 5. Ubah jawaban (salah → benar) | E2E-8 (redo) | salah dulu `{Jepang: Paris}` → setelah ubah `{Jepang: Tokyo}` → akhir peta lengkap benar | 2.00 | 2.00 | LOLOS |
| 6. Clear (jawab benar → clear semua) | E2E-6 | akhir `{Jepang: "", Prancis: "", Indonesia: ""}` — pemetaan `undefined → ""` dipertahankan | 0.00 | 0 | LOLOS |

Catatan kejujuran: baseline "pra-perbaikan" tidak dapat dijalankan ulang
dalam sesi ini (file sudah diperbaiki Tim Coding, tanpa git untuk
checkout versi lama). Yang diverifikasi: (a) format payload sesuai
kontrak `{kiri: kanan}` termasuk `{kiri: ""}` saat clear; (b) poin
konsisten dengan rumus backend proporsional; (c) kontrol tetap antd
`Select` (bukan native). Nilai parsial 0.67 konsisten dengan laporan
regresi sebelumnya (pembulatan backend, bukan 66,67 di skala persen).

## 5. Regresi perbaikan lama

| Item | Bukti | Hasil |
|---|---|---|
| Timer (tabular-nums, tidak lompat) | `fontVariantNumeric = tabular-nums`; lebar timer 169.22 px TETAP setelah 2.5 dtk (teks `02:58:24` → `02:58:22`) | LOLOS |
| Autosave | Setiap jawaban muncul di DB setelah ±500 ms–1 dtk (6/6 soal); indikator "Tersimpan" terlihat | LOLOS |
| Anti-curang 1x = peringatan | Siswa E2E-7 (terpisah): simulasikan pindah tab → banner "Peringatan (1×) — pelanggaran berikutnya memblokir ujian." + tombol Mengerti; DB: 1 pelanggaran | LOLOS |
| Anti-curang 2x = auto-block | Pelanggaran ke-2 (>5 dtk sesudahnya, melewati dedupe client) → layar "Ujian Diblokir"; DB: 2 pelanggaran, attempt `diblokir` | LOLOS |
| Banner tidak menggeser soal (P-08) | Posisi top kartu soal 273 px SEBELUM dan SESUDAH banner (geser 0 px); banner 1 baris kompak | LOLOS |
| Ragu-ragu tandai + lepas | §2: tandai → 1, lepas → 0, tandai Soal 1 → Hasil tampil Tag "Ragu-ragu" | LOLOS |
| Modal mencantumkan nomor belum dijawab (D-11) | §2: "Soal belum dijawab: Soal 5, Soal 6" | LOLOS |
| Kartu PDF tidak tersentuh | Tim E2E TIDAK mengubah file sumber apa pun (mtime 6 file scope = sebelum sesi). Grep `window.print` di 6 file scope = 0. File kartu (`KartuSayaPage`, `KartuPesertaPage`, `kartu_pdf.go`) tidak dibuka untuk diubah | LOLOS |

## 6. HasilPage (P-11)

Screenshot: `e2e-hasil-1280.png`.
- Jawaban Menjodohkan tampil PER PASANGAN (3 baris: `Ibukota Indonesia → Jakarta`,
  `Ibukota Jepang → Tokyo`, `Ibukota Prancis → Paris`) — bukan satu string diratakan. LOLOS.
- "Poin: X / Y" sebagai Tag biru antd (`Poin: 2 / 2`, `Poin: 1 / 1`, `Poin: 0 / 1`). LOLOS.
- Gaya konsisten: Card "Soal N" + Tag status (Benar/Salah/Menunggu koreksi) + Tag
  "Ragu-ragu" + `Statistic` nilai 36 px (66.67), shell kolom baca sama. LOLOS.
- Jawaban siswa tampil literal; pembahasan Tryout tampil.

**Observasi DI LUAR SCOPE (bukan bug file scope, dilaporkan jujur, tidak
diperbaiki):** Soal 2 Benar-Salah pada data uji LAMA dinilai 0 padahal
jawaban `[true, true]` dan kunci `{"p1": true, "p2": true}` serta
pembahasan "Kedua pernyataan benar." — format kunci lama (`p1`/`p2`)
kemungkinan tidak cocok dengan format jawaban array dari halaman ujian.
Ini soal grading backend / data lama, di luar file-file scope E2E ini
(frontend ujian mengirim format yang benar: `boolean[]`). Perlu cek
terpisah oleh tim grading/data — TIDAK memengaruhi putusan E2E halaman ujian.

## 7. Perintah wajib (dikutip + output)

```
$ cd frontend && npx tsc --noEmit
(exit 0, tanpa output — tidak ada error tipe)

$ cd frontend && npm run build
> tsc -b && vite build
✓ built in 18.29s
dist/assets/index-LKV_807T.js  2,454.02 kB │ gzip: 747.60 kB
(!) chunk > 500 kB — warning yang sudah dikenal, bukan blocker.

$ cd backend && go test ./...
ok  cbt-smk-v2/backend/internal/analisisx (cached)
ok  cbt-smk-v2/backend/internal/auth (cached)
ok  cbt-smk-v2/backend/internal/event 0.074s
ok  cbt-smk-v2/backend/internal/exam (cached)
ok  cbt-smk-v2/backend/internal/grading (cached)
ok  cbt-smk-v2/backend/internal/master (cached)
ok  cbt-smk-v2/backend/internal/middleware (cached)
ok  cbt-smk-v2/backend/internal/rbac (cached)
ok  cbt-smk-v2/backend/internal/sod (cached)
ok  cbt-smk-v2/backend/internal/ukk (cached)
ok  cbt-smk-v2/backend/internal/ukkx (cached)
(paket lain: [no test files] — sama seperti sebelumnya)
```

## 8. Bug yang ditemukan

NIHIL. Tidak ada bug baru di file-file scope
(`UjianPage.tsx`, `SoalRenderer.tsx`, `NavigasiSoal.tsx`, `Timer.tsx`,
`BannerPeringatan.tsx`, `HasilPage.tsx`). Satu observasi di luar scope
tercatat di §6 (grading Benar-Salah data lama).

## 9. Yang belum diuji (jujur)

- Tanpa browser live manusia / tanpa klik manual manusia — semua interaksi
  lewat Chromium headless otomatis. Perlu spot-check 1× oleh user
  (login → jawab Menjodohkan → ragu-ragu → Selesai → Hasil) di browser
  sungguhan, termasuk di HP fisik.
- Tanpa uji 40 soal (kartu navigasi scroll internal + max-height) — paket
  uji hanya 6 soal; klaim P-03 untuk 40 soal belum dibuktikan E2E ini.
- Tanpa uji viewport tinggi kecil (laptop 1366×768 tidak maksimal).
- Tanpa uji reload/close tab di tengah ujian (sengaja dihindari agar
  tidak memicu pelanggaran pada attempt uji; perilaku reload = di luar
  scope perbaikan presentasi ini).
- Baseline pra-perbaikan tidak dijalankan ulang (lihat §4).
- `npm run build` hanya membangun `frontend/dist`; `.exe`/zip TIDAK
  di-rebuild sesuai instruksi tugas.
