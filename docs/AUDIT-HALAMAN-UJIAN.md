# AUDIT HALAMAN UJIAN SISWA — CBT SMK v2

**Tim:** Analisis UX (5 orang) — tim pertama; hasil ini acuan Tim Coding.
**Tanggal:** 17 Okt 2026.
**Ruang lingkup:** halaman ujian (`/ujian/:id`) + `HasilPage` (konsistensi).
**Aturan tim ini:** HANYA analisis + dokumen. Tidak mengubah kode apa pun.
Tidak menyentuh file kartu, file `.mdj`, backend, atau halaman lain.
Tanpa CDN. Tanpa logo/nama instansi resmi. Hanya lapisan presentasi —
tidak ada usulan ubah route, endpoint API, logika grading, timer,
anti-curang, autosave, payload jawaban, atau validasi.

## 0. Metode & bukti

File yang dibaca sampai paham struktur layout-nya:

| File | Baris |
|---|---|
| `frontend/src/pages/siswa/UjianPage.tsx` | 709 |
| `frontend/src/components/exam/SoalRenderer.tsx` | 412 |
| `frontend/src/components/exam/NavigasiSoal.tsx` | 95 |
| `frontend/src/components/exam/Timer.tsx` | 77 |
| `frontend/src/components/exam/BannerPeringatan.tsx` | 39 |
| `frontend/src/pages/siswa/HasilPage.tsx` | 212 |
| `frontend/src/components/Layout.tsx` (kerangka induk, dibaca sebagai konteks) | 92 |
| `frontend/src/App.tsx` (token `controlHeightLG`, dibaca sebagai konteks) | — |

Bukti visual (bukan ditebak):

1. **Screenshot user** (halaman `/ujian/:id`, dikirim bersama tugas):
   bar bawah pecah dua kotak, panel Navigasi tidak sejajar, kontrol
   Menjodohkan sempit, bilah info berdesakan, hierarki soal belum enak dibaca.
2. **Screenshot repo** `screenshots/qa-final/ujian-pengerjaan-{360,768,1280}.png`
   — diverifikasi ulang oleh tim ini dengan crop + perbesar 2×:
   - **768px: bar bawah SATU strip utuh** (kiri Sebelumnya/Berikutnya,
     kanan ragu-ragu + Selesai, satu kotak putih).
   - **1280px: bar bawah PECAH DUA KOTAK** — kotak kiri berisi
     Sebelumnya/Berikutnya, kotak kanan berisi Tandai ragu-ragu +
     Selesai Ujian, kotak kanan bergeser turun ±10–17px, ada celah
     latar abu-abu di antaranya, dan kotak kanan menimpa area putih
     bawah kartu Navigasi. Ini bukti visual D-01.
   - **360px: bar bawah SATU kotak dua baris**, tapi bar menutupi
     baris pasangan Menjodohkan terakhir ("Ibukota Jepang" terpotong
     bar) — bukti D-02. Bilah info 4 baris, "Soal: 1 dari 1" terpotong
     tengah frasa — bukti D-05. Area kosong abu-abu besar di atas
     header ujian — bukti D-03.
3. **1440px: TIDAK ADA screenshot.** Semua klaim 1440 di dokumen ini
   adalah prediksi dari kode dan WAJIB dibuktikan Tim Coding dengan
   screenshot baru (§7).

Dokumen sebelumnya yang dipakai sebagai pembanding (bukan diulang):
`docs/AUDIT-KONSISTENSI.md` §1.7 P1/P2/P3/P13 (bar bawah, bilah info,
grid nomor, layout ujian) dan `docs/AUDIT-UI-UX.md` §3.5.
Dua klaim lama di `AUDIT-UI-UX.md` **tidak sesuai kode saat ini** dan
dikoreksi di sini: (a) "grid soal sticky di desktop + drawer di HP" —
kode saat ini TIDAK punya Drawer; panel HP adalah blok yang di-toggle
(D-07). (b) "Halaman Ujian sudah bagus, hanya poles kecil" — dibantah
oleh screenshot user + screenshot 1280 (D-01).

---

## 1. Daftar cacat (dengan bukti `file:baris`)

Skala: **KRITIS** = merusak alur ujian / terlihat pertama kali;
**SEDANG** = mengganggu nyata; **RINGAN** = poles.

### D-01 — KRITIS. Bar bawah PECAH dua kotak di desktop (≥1024px)

**Gejala (terbukti visual):** di 1280px, bar bawah tampil sebagai dua
kotak terpisah — kiri = Sebelumnya/Berikutnya, kanan = Tandai
ragu-ragu + Selesai Ujian — kotak kanan bergeser turun, ada celah di
tengah, dan kotak kanan menimpa area bawah kartu Navigasi.
Di 768px bar masih satu strip. Jadi cacat ini **aktif tepat saat
breakpoint `lg` aktif** (grid dua kolom + navigasi sticky).

**Bukti kode:**

- Bar bawah adalah SATU elemen di sumber:
  `UjianPage.tsx:647` (`sticky bottom-0 … border-t … bg-white`)
  + `:648` (wadah dalam `flex … sm:justify-between`).
- Tetapi bar itu **duduk SETELAH grid dua kolom** `UjianPage.tsx:565`
  (`flex flex-col gap-4 lg:flex-row`) dan **tidak berbagi satu grid
  penjajaran** dengan kolom soal + kolom navigasi di atasnya.
- Lebar bar dikendalikan TIGA batas bertumpuk yang tidak sama:
  `Layout.tsx:84` (Content `max-w-[1100px]`) →
  `UjianPage.tsx:485` (root `max-w-7xl` = 1280px, **mati** karena induk
  membatasi 1100px — menyesatkan) →
  `UjianPage.tsx:648` (wadah dalam bar `max-w-7xl` lagi).
- Perataan bar mengandalkan trik margin negatif di DUA tempat:
  `UjianPage.tsx:514` (header sticky `-mx-4`) dan `:647` (bar `-mx-4`).
- Kartu Navigasi `UjianPage.tsx:619` (`lg:sticky lg:top-36 lg:w-80`)
  ikut `stretch` (default flex) sehingga kotak putihnya memanjang
  sampai dasar baris — area putih kosong di bawahnya menjadi latar
  tempat kotak kanan bar "menempel" secara visual.

**Diskrepansi yang wajib dijelaskan Tim Coding (§7):** sumber
menunjukkan satu elemen bar, render menunjukkan dua kotak. Tim Coding
WAJIB mereproduksi dari build FRESH dan melaporkan mekanisme persisnya
(interaksi sticky nav + sticky bar, atau build lama yang ter-embed).
Apa pun mekanismenya, perbaikan struktural P-01 berlaku — bukan poles
margin.

### D-02 — SEDANG. Bar bawah sticky menutupi konten terakhir

**Gejala (terbukti visual):** di screenshot 360px, bar bawah menutupi
baris pasangan Menjodohkan terakhir. Saat scroll, konten berjalan
DI BAWAH bar yang menempel.

**Bukti kode:** wadah halaman hanya `py-4` (`UjianPage.tsx:485`) —
tidak ada padding bawah / scroll-padding sebesar tinggi bar.
Bar `sticky bottom-0` (`UjianPage.tsx:647`) tingginya ±110px di HP
(dua baris) dan ±68px di desktop. Konten terakhir halaman tidak punya
ruang bebas di bawahnya.

### D-03 — SEDANG. Slot banner = ruang kosong mati ±128px

**Gejala (terbukti visual):** di screenshot 1280px ada celah abu-abu
besar antara header global dan header ujian saat TIDAK ada peringatan.

**Bukti kode:** slot dicadangkan `UjianPage.tsx:490`
(`min-h-[140px] sm:min-h-[88px]`), ditambah `Content py-6`
(`Layout.tsx:84`) + halaman `py-4` (`UjianPage.tsx:485`).
Niat slot permanen (anti-geser) benar, tetapi harganya ruang mati
permanen. Banner-nya sendiri (`BannerPeringatan.tsx:25-38`) adalah
Alert lengkap (message + description 2 baris + tombol action) sehingga
slot harus setinggi itu.

### D-04 — SEDANG. Header ganda: brand diulang dua kali

**Gejala:** dua bilah biru bertumpuk di semua screenshot — header
global + header ujian, keduanya menulis "CBT SMK".

**Bukti kode:** `Layout.tsx:48-49` (Header global `h-16`, brand
"CBT SMK v2") + `UjianPage.tsx:515-521` (bilah biru dalam halaman:
"CBT SMK / Ujian Berbasis Komputer"). Bilah kedua memakan tinggi
sticky atas tanpa menambah informasi (nama/mapel sudah ada di bilah
info di bawahnya).

### D-05 — SEDANG. Bilah info berdesakan; status tersimpan tanpa visual state

**Gejala (terbukti visual):** di 360px bilah info menjadi 4 baris dan
"Soal: 1 dari 1" terpotong tengah frasa ("1 dari" / "1").

**Bukti kode:**

- `UjianPage.tsx:522-523` — satu `<dl>` `flex-wrap` menampung 5
  pasangan dt/dd (Nama/NISN/Kelas/Mapel/Soal) + status + tombol "Soal".
  Tidak ada aturan "Soal X dari Y tidak boleh terpotong" — pasangan
  `:540-543` bisa wrap di tengah frasa.
- `UjianPage.tsx:547-553` — status simpan hanya `<span>` teks polos
  (`Menyimpan…` / `Tersimpan` / `Gagal menyimpan — coba lagi`),
  tanpa ikon dan tanpa warna status (gagal tidak merah, tersimpan
  tidak hijau). Saat `idle` tidak tampil apa-apa sehingga siswa tidak
  tahu jawaban terakhir sudah tersimpan atau belum. Ini juga
  menutup temuan lama `AUDIT-UI-UX.md` §3.5 butir (3) yang masih terbuka.
- `UjianPage.tsx:553-560` — tombol `lg:hidden` berlabel "Soal"
  (ambigu — maksudnya "Daftar Soal" / navigasi), tanpa badge jumlah
  soal belum dijawab.

### D-06 — SEDANG. Offset sticky navigasi rapuh (`lg:top-36`)

**Bukti kode:** `UjianPage.tsx:619` (`lg:sticky lg:top-36` = 144px).
Angka 144px mengasumsikan tinggi header tetap, padahal tinggi header
BERVARIASI karena isinya `flex-wrap` (`UjianPage.tsx:515` bilah biru,
`:522` bilah info). Di lebar menengah bilah info wrap 2 baris dan
offset 144px salah: kartu navigasi menimpa header atau melayang
dengan celah. Selain itu kartu navigasi **tanpa `max-height` +
scroll internal** — dengan soal banyak (mis. 40 soal), bagian bawah
kartu (legenda, hitungan) tidak terjangkau karena kartu sticky
terjepit viewport.

### D-07 — SEDANG. Panel navigasi bukan drawer di HP; ruang kosong janggal di desktop

**Bukti kode:**

- `UjianPage.tsx:618-621` — panel navigasi adalah `Card` yang
  di-toggle `showNav ? "block" : "hidden"` dan di HP muncul
  **di bawah kartu soal** (urutan DOM), mendorong bar bawah ke bawah.
  Ini bukan Drawer. Klaim "drawer di HP" di `AUDIT-UI-UX.md:123`
  tidak sesuai kode.
- Flex row `UjianPage.tsx:565` memakai `align-items: stretch`
  (default) sehingga kartu navigasi memanjang setinggi kolom soal.
  Isinya hanya grid + legenda + hitungan (`UjianPage.tsx:637`),
  sehingga di desktop tampil **kotak putih kosong besar** di bawah
  "Dijawab: 0 / 1 · Belum: 1" — terlihat jelas di screenshot 1280px.
  Inilah "ruang kosong janggal" + "tidak sejajar" yang dilaporkan user:
  panel terlihat meluber ke bawah melewati garis dasar kartu soal.
- Grid nomor `NavigasiSoal.tsx:29` (`grid-cols-4 gap-1.5`) memakai
  gap 6px — di luar basis 8px (§P-07). Tombolnya sendiri sudah benar
  44px (`NavigasiSoal.tsx:55`). Legenda `text-xs` (`:64`) rapat tapi
  masih terbaca; yang kurang adalah indikator progres, bukan hanya
  teks hitungan.

### D-08 — SEDANG. Kontrol Menjodohkan: panah hilang di HP, baris tanpa struktur

**Bukti kode** (`SoalRenderer.tsx`, blok Menjodohkan/Peta Konsep `:216-235`):

- Panah `→` memakai `hidden … sm:inline` (`:221-223`) — **di HP
  panah HILANG total**, sehingga kotak kiri dan dropdown tampil sebagai
  dua kotak bertumpuk tanpa penanda pasangan.
- Baris (`:217`, `flex flex-col gap-2 sm:flex-row`) tidak punya
  wadah baris (tanpa border/latar/nomor) — deretan pasangan melayang
  di ruang kosong, sulit dilacak saat pasangan banyak.
- Kotak kiri (`:218`) **tanpa `min-height`** — tingginya mengikuti isi
  teks (±40px untuk 1 baris, lebih untuk teks 2 baris), sedangkan
  Select `size="large"` = 44px (via `controlHeightLG: 44`,
  `App.tsx:102`). Akibatnya pasangan kiri-kanan tidak sejajar rapi
  saat teks kiri membungkus 2 baris.
- Pembagian lebar 50:50 (`:218` kiri `flex-1`, `:233` Select
  `sm:flex-1`) tanpa kolom nomor. Teks placeholder Select
  ("— Pilih pasangan —") tampil pucat di samping kotak kiri yang
  ber-latar abu-abu — kontras visual tidak konsisten.
- Yang SUDAH BENAR dan dipertahankan: Select tetap antd `Select`
  (`:224`), `showSearch` TIDAK aktif (`:208` — keputusan lama jalur
  penilaian, dipertahankan), `allowClear` sebagai padanan opsi kosong,
  pemetaan `undefined → ""` agar payload backend identik.

### D-09 — SEDANG. Hierarki stimulus → pertanyaan → opsi belum enak dibaca

**Bukti kode** (`UjianPage.tsx`):

- Stimulus `:582` vs pertanyaan `:586` — keduanya `text-[16px]
  leading-relaxed`, bedanya hanya kotak stimulus punya border + latar
  `bg-slate-50`. **Tanpa label** ("Bacaan"/"Stimulus"/"Pertanyaan")
  dan tanpa pemisah — stimulus panjang menyatu dengan pertanyaan.
  Jarak stimulus→pertanyaan hanya `mt-4` (`:586`).
- Media soal `:593` (`max-h-72`) tanpa caption / keterangan.
- Tidak ada indikator progres visual (hanya teks hitungan
  `UjianPage.tsx:637` dan di Modal `:700`). Pola ujian yang baik
  menampilkan progres "dijawab/total" sebagai bar/angka yang selalu
  terlihat.

### D-10 — RINGAN–SEDANG. Spacing off-grid + tipografi arbitrer

- `NavigasiSoal.tsx:29` `gap-1.5` = 6px (bukan kelipatan 8).
- `UjianPage.tsx:596` `mt-5` = 20px di antara pertanyaan dan kontrol
  jawaban, sementara jarak lain memakai `mt-4` (16px).
- `text-[16px]` arbitrer di `UjianPage.tsx:582,586` dan
  `SoalRenderer.tsx:218` (dan 4 tempat lain) — nilai benar (16px
  untuk soal) tapi ditulis sebagai nilai arbitrer, bukan skala.
- `Timer.tsx:60` `fontSize: 22` — di luar skala tipografi (20/24).
- Yang SUDAH BENAR: Timer memakai `tabular-nums` + `minWidth: 8ch`
  (`Timer.tsx:70`) sehingga angka tidak bergeser; dipertahankan.

### D-11 — SEDANG. Modal konfirmasi hanya menampilkan hitungan, bukan nomor soal

**Bukti kode:** Modal `UjianPage.tsx:688+` menampilkan
"Dijawab: X dari Y soal" (`:700`) + peringatan jawaban tidak bisa
diubah. Siswa yang masih punya soal kosong **tidak diberi tahu soal
nomor berapa** yang kosong — ia harus menutup modal dan mencari
sendiri lewat grid. Pola ujian yang baik (Moodle) mencantumkan
daftar nomor soal belum dijawab di dialog sebelum kumpul.

### D-12 — RINGAN. Input Teks Rumpang lebar tetap 160px

**Bukti kode:** `SoalRenderer.tsx:305`
(`width: 160, maxWidth: "100%"`). Lebar tetap tidak adaptif: di layar
sempit input memaksa baris membungkus canggung; di layar lebar input
terlihat kerdil untuk jawaban panjang. (`maxWidth: "100%"` sudah
mencegah meluber — yang kurang adalah adaptivitas.)

### D-13 — SEDANG. HasilPage tidak konsisten dengan alur ujian

**Bukti kode** (`HasilPage.tsx`):

- `:118` `max-w-3xl` (768px) vs alur ujian yang shell-nya 1100px
  (`Layout.tsx:84`) — lompatan lebar saat pindah Ujian → Hasil.
  (Untuk keterbacaan, kolom baca 768px itu benar — yang salah adalah
  tidak adanya shell luar yang sama; lihat P-11.)
- `:44-56` `tampilJawaban` meratakan jawaban Menjodohkan
  (`{kiri: kanan}`) menjadi SATU string "kiri → kanan; kiri → kanan; …"
  — struktur per pasangan yang ada di halaman ujian HILANG di hasil.
- `:184` "Poin: X / Y" sebagai `<p>` polos, bukan Tag seperti status
  (`:154-161` memakai Tag + ikon — bagus, dipertahankan).
- `:154` judul Card memakai `<span>` mentah, bukan tipografi sistem.
- Yang SUDAH BENAR: `Statistic` nilai 36px sama dengan layar
  selesai ujian (`UjianPage.tsx` hasil-submit), Tag status + ikon
  konsisten, jawaban ketikan siswa tampil literal (tidak
  diinterpretasi HTML).

### D-14 — LULUS (dicatat agar tidak "diperbaiki" sia-sia)

- Satu primary per tampilan: hanya "Selesai Ujian" `type="primary"`
  (`UjianPage.tsx:677-686`); Sebelumnya/Berikutnya default.
  Dipertahankan (P-02).
- Baris PG / PG Kompleks / Benar-Salah sudah `min-h-[44px]`
  (`SoalRenderer.tsx:106,143,183`) dengan area klik kartu penuh.
  Dipertahankan.
- Timer `tabular-nums` + `minWidth: 8ch` (`Timer.tsx:70`).
  Dipertahankan.
- Modal konfirmasi sudah ada (mengurangi risiko klik Selesai tidak
  sengaja). Dipertahankan + dilengkapi daftar nomor (D-11).

---

## 2. Perbandingan pola UI ujian yang baik (POLANYA saja)

Tanpa menjiplak branding/aset siapa pun. Pola dirujuk dari perilaku
yang sudah didokumentasikan di `docs/AUDIT-UI-UX.md` + pola umum
aplikasi ujian (ANBK, Moodle, Exam.net).

| Pola ujian yang baik | Keadaan CBT SMK v2 saat ini | Kesenjangan |
|---|---|---|
| **Satu bilah aksi bawah** yang selalu menempel, satu latar, semua aksi navigasi + selesai dalam satu baris/kotak | Bar pecah dua kotak di desktop (D-01); menutupi konten (D-02) | Struktur bar harus di luar grid dua kolom (P-01) |
| **Satu header ujian**: identitas + timer + progres, tanpa ulang brand | Dua bilah biru brand bertumpuk (D-04); info 5 field + status + tombol dalam satu `<dl>` wrap (D-05) | Satu header, info ≤2 baris (P-09) |
| **Navigasi soal**: panel samping sticky di desktop (tinggi dibatasi + scroll sendiri), drawer/overlay di HP; grid nomor + legenda + progres | Panel stretch penuh tinggi baris + ruang kosong (D-07); HP = blok di bawah soal, bukan drawer; tanpa max-height/scroll (D-06) | Sticky + `align-self: flex-start` + max-height + Drawer HP (P-03) |
| **Satu soal satu layar**, hierarki Bacaan → Pertanyaan → Jawaban, progres selalu terlihat | Tanpa label stimulus/pertanyaan, tanpa pemisah, tanpa progress bar (D-09) | Label + pemisah + progres (P-05) |
| **Kontrol jawaban sejajar**: tiap baris jawaban satu grid rapi, target sentuh ≥44px, pasangan kiri-kanan tersambung visual (nomor + panah + kontrol) | Menjodohkan: panah hilang di HP, tanpa nomor baris, kiri tanpa min-height (D-08) | Grid baris Menjodohkan (P-06) |
| **Banner peringatan tidak menggeser soal**: slot/cadangan tinggi tetap ATAU overlay; banner kompak satu baris | Slot permanen 140/88px = ruang mati (D-03) | Slot = tinggi banner kompak (P-08) |
| **Konfirmasi kumpul** mencantumkan soal yang belum dijawab (nomornya) | Modal hanya hitungan (D-11) | Daftar nomor belum dijawab di modal (P-02) |
| **Hasil mencerminkan ujian**: struktur jawaban yang sama dengan saat mengerjakan | Menjodohkan diratakan jadi satu string (D-13) | Render per pasangan di Hasil (P-11) |

---

## 3. Matriks responsif

| Lebar | Yang teramati (screenshot) / diprediksi dari kode | Wajib setelah perbaikan |
|---|---|---|
| **360px** (HP kecil) | Bar 1 kotak 2 baris, menutupi konten (D-02). Info 4 baris, "Soal: 1 dari 1" wrap tengah frasa (D-05). Panah Menjodohkan hilang (D-08). Slot banner 140px (D-03). Chrome atas+bawah memakan ±40% tinggi viewport (header global 64px + bilah ujian ~160px saat wrap + bar ~110px). | Bar 1 kotak 2 baris, konten punya padding bawah ≥ tinggi bar + 16px. Info ≤2 baris. Panah selalu terlihat. Slot ≤56px. |
| **768px** (tablet) | Bar SATU strip (terbukti). Layout 1 kolom (breakpoint `lg` belum aktif). Panel nav tersembunyi sampai tombol "Soal" ditekan, lalu muncul sebagai blok penuh di bawah soal. | Bar tetap 1 strip. Nav = Drawer, bukan blok pendorong konten. Info ≤2 baris. |
| **1280px** (desktop) | Bar PECAH 2 kotak (terbukti). Nav kartu stretch + ruang putih kosong (D-07). Info 1 baris (OK). | Bar 1 kotak 1 baris. Nav `align-self: flex-start`, tinggi dibatasi + scroll. |
| **1440px** (desktop lebar) | TIDAK ADA screenshot. Diprediksi dari kode: sama seperti 1280 + ruang mati samping membesar karena shell 1100px di tengah; `max-w-7xl` (1280px) di halaman mati karena induk 1100px (D-01) — bila shell berubah, lebar halaman melompat. | Bar 1 kotak, lebar = lebar shell (1100px). Screenshot baru wajib (P-12). |

Juga: **tinggi viewport kecil** (laptop 1366×768, browser tidak
maksimal) — kartu navigasi tanpa max-height (D-06) + header sticky
tinggi membuat grid nomor bawah tidak terjangkau. Wajib diuji.

---

## 4. PRINSIP DESAIN MENGIKAT untuk Tim Coding

"Mengikat" = Tim Coding tidak boleh menyimpang tanpa persetujuan
koordinator + bukti screenshot. Semua prinsip HANYA lapisan presentasi.

### P-01 — Satu bar bawah tunggal, selalu menempel bawah viewport

- Bar bawah adalah **SATU elemen, SATU latar putih, SATU `border-t`**,
  `sticky bottom-0`, dan **BERADA DI LUAR grid dua kolom**
  (bukan anak grid, bukan sejajar kolom manapun).
- **DILARANG** kotak ngambang kedua, `fixed` kedua, atau elemen
  aksi melayang lain di halaman ujian.
- Lebar bar = lebar shell konten (1100px, mengikuti `Layout`),
  bukan `max-w-7xl` sendiri. **Hapus** `max-w-7xl` di halaman
  ujian dan **hapus** trik `-mx-4` (header sticky + bar).
- Isi: grup kiri (Sebelumnya / Berikutnya) + grup kanan
  (Tandai ragu-ragu + Selesai Ujian). Di ≥768px SATU baris
  (`justify-between`); di <768px dua baris **dalam kotak yang sama**.
- Halaman diberi padding bawah / scroll-padding **≥ tinggi bar +
  16px** agar konten terakhir tidak pernah tertutup bar (D-02).
- Tanpa shadow arbitrer (sesuai `DESIGN-SYSTEM.md`: bar memakai
  `border-t` saja).

### P-02 — Satu tombol aksi primer per tampilan: hanya "Selesai Ujian"

- "Selesai Ujian" satu-satunya `type="primary"`. Sebelumnya /
  Berikutnya `type="default"`. Ragu-ragu tetap Checkbox berbingkai
  (bukan Button primary).
- Checkbox ragu-ragu di HP tidak boleh mendesak tombol Selesai
  (beri lebar minimum yang adil; Selesai tidak boleh terpotong).
- Modal konfirmasi **mencantumkan nomor soal yang belum dijawab**
  (dari state jawaban yang sudah ada — tanpa endpoint baru), selain
  hitungan yang sudah ada. Tombol modal: "Ya, Kumpulkan" primary,
  "Batal" default. Satu primary per modal.

### P-03 — Grid soal : navigasi proporsional, sejajar, sticky yang benar

- Desktop (≥1024px): kolom soal `1fr` + navigasi **320px tetap**
  (`w-80`), `align-self: flex-start` (DILARANG stretch — inilah
  penyebab ruang putih kosong D-07).
- Navigasi sticky dengan offset = tinggi header AKTUAL (bukan angka
  magis `top-36`). Caranya: buat tinggi header ujian TETAP
  (bilah info ≤2 baris, P-09) lalu samakan offset dengan tinggi itu,
  atau satukan header + navigasi dalam satu kerangka sticky.
- Navigasi punya `max-height = viewport − header − bar` +
  `overflow-y-auto` internal — grid nomor + legenda + hitungan harus
  selalu terjangkau walau soal 40+.
- Di <1024px navigasi = **Drawer antd** (bukan blok di bawah soal):
  dibuka tombol "Daftar Soal" (label baru, menggantikan "Soal")
  dengan badge jumlah soal belum dijawab; menutup drawer saat
  memilih nomor; tidak mendorong konten.
- Grid nomor: gap basis 8px (`gap-2`), tombol 44×44px, legenda
  12px. Header kartu navigasi menampilkan progres
  "Dijawab A dari B" + progress bar (bukan hanya teks hitungan).

### P-04 — Area soal max-width terbaca

- Kolom soal: teks dibatasi **max-width 768px** (reading measure);
  Card soal padding 16px di HP / 24px di ≥sm (sudah benar —
  dipertahankan). Kolom soal tidak boleh melebar mengikuti seluruh
  1100px.

### P-05 — Hierarki stimulus → pertanyaan → opsi enak dibaca

- Label kecil (12px, 600, abu-abu) di atas kotak stimulus:
  "Bacaan" / "Stimulus"; label "Pertanyaan" di atas pertanyaan
  (atau pertahankan judul Card "Soal N" + tambah label ini).
- Pemisah 16px antara stimulus dan pertanyaan (bukan hanya `mt-4`
  tanpa penanda visual). Stimulus tetap berlatar netral; pertanyaan
  tanpa latar — perbedaan dipertahankan + diperkuat label.
- Media soal diberi caption singkat di bawah gambar.
- Kontrol jawaban mulai 16px di bawah pertanyaan (samakan —
  saat ini 20px di satu tempat, 16px di tempat lain).

### P-06 — Kontrol jawaban full-width, sejajar, mudah diklik (min 44px)

- Semua kontrol jawaban **tinggi sentuh min 44px** (sudah dijamin
  `controlHeightLG: 44` + `size="large"` — JANGAN diturunkan).
  Baris PG / Benar-Salah yang sudah `min-h-[44px]` dipertahankan.
- **Menjodohkan / Peta Konsep** — satu baris = satu grid:
  `[nomor 28px | kiri 1fr | panah 24px | Select 1fr]`,
  tinggi baris min 44px, semua sel `align-items: center`.
  Nomor baris 1..n ditampilkan. Kotak kiri `min-height: 44px` +
  konten terpusat vertikal.
- **Panah SELALU terlihat**, termasuk HP (hapus `hidden sm:inline`).
  Di HP sempit, panah tetap inline kecil di antara kiri dan Select —
  jangan hilangkan penanda pasangan.
- Select **TETAP antd `Select`** (JANGAN kembalikan ke native),
  tanpa `showSearch` (keputusan jalur penilaian dipertahankan),
  `allowClear` dipertahankan, placeholder sama, pemetaan
  `undefined → ""` tidak berubah.
- Teks Rumpang: input inline memakai lebar adaptif
  (`min-width` + `max-width: 100%`), bukan lebar tetap 160px.

### P-07 — Jarak konsisten basis 8px

- Semua gap / padding / margin kelipatan 8px:
  `gap-1.5` (6px) → `gap-2`; `mt-5` → `mt-4`; dst.
- Radius 8 untuk semua. Satu pola tombol (P-02).
- Header halaman = Breadcrumb + Title (aturan konsistensi yang
  berlaku); halaman ujian memakai header ujian khusus P-09
  (pengecualian yang sudah ada — dipertahankan, bukan ditambah).

### P-08 — Banner peringatan: slot permanen KOMPAK, tidak menggeser soal

- Slot dicadangkan dengan tinggi TETAP = tinggi banner kompak.
- Banner dibuat **satu baris kompak**: ikon + teks singkat
  ("Peringatan 1× — pelanggaran berikutnya memblokir ujian.") +
  tombol "Mengerti". Description panjang dipadatkan (makna sama).
- Muncul / hilangnya banner **TIDAK menggeser soal** (slot tetap).
- Saat tidak ada peringatan, slot tampil sebagai ruang kosong
  **setinggi banner kompak saja (≤56px)**, bukan 140/88px (D-03).
- Alert "diblokir" (layar penuh) tidak memakai slot — tampil
  sebagai layar sendiri seperti sekarang.

### P-09 — Bilah info peserta rapi + timer lebar tetap + status tersimpan jelas

- **Satu header ujian, tanpa ulang brand**: hapus bilah brand biru
  dalam halaman (header global sudah membawa brand), ATAU
  sembunyikan header global selama ujian — pilih SATU, jangan dua
  (D-04). Bilah info + timer menjadi satu-satunya header sticky.
- Bilah info **maksimal 2 baris di lebar apa pun**: baris 1 =
  identitas (Nama / NISN / Kelas / Mapel) dengan elipsis bila
  sempit; baris 2 (atau sisi kanan baris 1 di desktop) =
  "Soal X dari Y" + progres + status simpan + tombol Daftar Soal.
- "Soal X dari Y" **SELALU satu kesatuan `nowrap`** — DILARANG
  wrap tengah frasa (D-05).
- Timer dipertahankan: `tabular-nums` + lebar cadangan `8ch`
  (sudah benar), angka 20–24px, merah bila <5 menit.
- Status tersimpan punya **visual state**: ikon + warna
  (menyimpan = abu-abu, tersimpan = hijau + cek, gagal = merah +
  teks "coba lagi"), posisi TETAP di bilah info dengan lebar
  cadangan agar teks yang berubah tidak menggeser tombol lain.
- Tombol navigasi HP berlabel **"Daftar Soal"** + badge jumlah
  belum dijawab (menggantikan label "Soal" yang ambigu).

### P-10 — Tipografi

- Soal (stimulus, pertanyaan, opsi): **16px / line-height 1.6**.
- Judul halaman 24/700; judul kartu 16/600; label 14/600;
  caption 12; angka Timer 20–24px `tabular-nums`.
- Hapus nilai arbitrer `text-[Npx]` baru — pakai skala yang ada.
  (Nilai 16px untuk soal sudah benar — yang diperbaiki adalah cara
  menulisnya, bukan nilainya.)

### P-11 — Konsistensi dengan HasilPage

- HasilPage memakai **shell yang sama**: luar 1100px (Layout),
  kolom baca 768px terpusat untuk daftar soal — agar transisi
  Ujian → Hasil tidak melompat lebar.
- Card "Soal N" + Tag status + ikon status sama dengan ujian.
- Jawaban Menjodohkan di Hasil dirender **per pasangan**
  (satu baris "kiri → kanan" per pasangan, cerminan baris ujian),
  BUKAN satu string yang diratakan.
- "Poin: X / Y" memakai Tag (bukan `<p>` polos).
- Nilai `Statistic` 36px dipertahankan (sama dengan layar selesai).
- Spacing basis 8px, tipografi P-10.

### P-12 — Responsif dibuktikan dengan screenshot

- Wajib screenshot **360 / 768 / 1280 / 1440** setelah perbaikan.
- Kriteria lolos per lebar: bar bawah SATU kotak; bilah info ≤2
  baris; navigasi tidak meluber / tidak mendorong konten (Drawer
  di HP); tidak ada konten tertutup bar; tidak ada scroll
  horizontal; tidak ada ruang putih kosong janggal di kartu navigasi.

---

## 5. Yang TIDAK boleh diubah Tim Coding (non-goal)

- Route, endpoint API, payload jawaban, format `value` per tipe soal
  (termasuk `{kiri: ""}` Menjodohkan), logika grading, timer
  (jangkar server + `onExpire`), anti-curang (peringatan 1× →
  blokir 2×), autosave (debounce 500ms), validasi.
- Teks tombol yang sudah ditetapkan: "Sebelumnya", "Berikutnya",
  "Tandai ragu-ragu", "Selesai Ujian" (boleh tambah label
  "Daftar Soal" untuk tombol navigasi + nomor di modal — itu
  penambah presentasi, bukan pengganti teks yang ditetapkan).
- Menjodohkan tetap antd `Select` (P-06). Tanpa CDN. Tanpa
  logo/nama instansi resmi. Tanpa menyentuh kartu, `.mdj`,
  backend, atau halaman lain.

---

## 6. Checklist verifikasi Tim Coding + QA

1. **Build FRESH** lewat `make build-windows` (frontend → sync-dist
   → go build). DILARANG `go build` langsung — exe meng-embed
   `backend/dist`, dan build tanpa sync-dist menghasilkan exe
   ber-frontend LAMA (pelajaran yang sudah tercatat). Bila bar
   ternyata sudah satu kotak di build fresh, lampirkan screenshot
   baru sebagai bukti dan tutup D-01 sebagai "build basi" — TETAP
   terapkan P-01 (hapus `-mx-4`, `max-w-7xl` mati, stretch) agar
   tidak regresi.
2. Jelaskan mekanisme persis D-01 (satu elemen di sumber vs dua
   kotak di render) dalam laporan — jangan hanya "sudah diperbaiki".
3. Screenshot baru 360 / 768 / 1280 / 1440 (halaman ujian +
   modal konfirmasi + banner peringatan + HasilPage), dinilai
   dengan kriteria P-12.
4. Regresi grading Menjodohkan (payload identik, termasuk
   `{kiri: ""}`) — seperti regresi 6/6 sebelumnya.
5. Uji navigasi: 40 soal, Drawer HP buka/tutup, kartu nav scroll
   internal sampai legenda terbaca, tombol Sebelumnya/Berikutnya
   di soal pertama/terakhir.
6. Uji banner: muncul/hilang tanpa geser soal (ukur posisi soal
   sebelum/sesudah), slot ≤56px saat tidak ada peringatan.
7. Spot-check manual 1× di browser sungguhan (klik, bukan hanya
   screenshot): login → mulai ujian → jawab Menjodohkan → ragu-ragu
   → Selesai → Hasil.

---

## 7. Catatan kejujuran tim analisis

- Tim ini tidak punya browser live; penilaian visual bersandar pada
  screenshot user (per deskripsi tugas) + screenshot repo yang
  di-crop dan diperbesar. Klaim 1440px adalah prediksi kode.
- Diskrepansi "satu elemen bar di sumber vs dua kotak di render"
  dicatat apa adanya (§1 D-01, §6 butir 1–2) — bukan ditebak
  penyebab pastinya. Tim Coding wajib membuktikannya.
- Prinsip di §4 sengaja ditulis agar benar di KEDUA kemungkinan
  penyebab (struktur bar yang rapuh ATAU build lama yang ter-embed):
  satu bar, di luar grid, tanpa trik margin negatif, dibuktikan
  screenshot baru.
