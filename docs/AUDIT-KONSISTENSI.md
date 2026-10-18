# AUDIT KONSISTENSI — CBT SMK v2 (Frontend)

> **Tim:** Audit Konsistensi (10 orang, tim pertama). **Tanggal:** 17 Okt 2026.
> **Ruang lingkup:** HANYA lapisan presentasi frontend React 19 + TypeScript + Vite +
> Ant Design 6.6.6 (`frontend/src/**`). **Tidak ada kode aplikasi yang diubah** dalam audit ini.
> **Metode bukti:** grep + baca file langsung (`file:baris` + kutipan). **Tidak ada screenshot**
> dalam audit ini — klaim screenshot tidak dibuat; bukti `file:baris` + kutipan kode
> dianggap cukup sesuai penugasan. Semua angka dihitung dari grep pada 17 Okt 2026
> (77 file `*.tsx` di `frontend/src`).
>
> **Status awal yang ditemukan:** redesign antd SUDAH berjalan jauh (56 file mengimpor
> `from "antd"`, 206 `<Button>` antd, 29 `<Table>` antd, 27 `<Modal>` antd), TETAPI
> masih ada **dua sistem desain yang hidup berdampingan**: sistem antd + sistem
> custom lama (Tailwind rakitan + 19 file komponen custom + class `.ds-*` +
> token Tailwind `@theme`). Inilah akar ketidakkonsistenan yang dirasakan user.
> Audit ini menetapkan SATU sistem (antd + token ConfigProvider) dan cara mematikan
> sistem lama sampai `grep` membuktikan 0 sisa.

---

## 0. Ringkasan eksekutif

| # | Pelanggaran teratas | Bukti | Dampak |
|---|---|---|---|
| 1 | **19 file komponen custom** masih ada, menduplikasi antd | §1.6 | Dua sistem tombol/kartu/modal/badge hidup bersama |
| 2 | **16 `<button>` mentah + 1 `<select>` native + 4 radio/checkbox native + 3 `<table>` mentah + 2 dialog `fixed` rakitan + 1 modal duplikat** | §1.1, §1.5 | Perilaku/fokus/aksesibilitas tidak seragam |
| 3 | **~700 kemunculan class warna non-token** (`text-slate-*` 386, `bg-slate-*` 51, `border-slate-*` 117, keluarga red/green/amber/blue/orange/emerald ~150) | §1.2 | Warna abu-abu/warna status tidak terpeta ke token |
| 4 | **7 sistem warna status paralel** (Tag antd + `__badgeColor` ×4 file + StatusBadge/PETA + Badge custom + span inline + DayaBedaBadge + Heatmap) | §1.2 | Status yang sama beda warna antar halaman |
| 5 | **Tipografi tidak satu skala**: judul halaman `text-xl` / `text-2xl` / `text-lg` / `ds-page-title` bercampur; `text-[15px]/[16px]/[17px]` arbitrer | §1.3 | Hierarki visual tidak konsisten |
| 6 | **Radius 6 varian, shadow 7 varian** (`rounded-lg` 109, `rounded-xl` 49, `rounded-full` 21, `rounded` 11, `rounded-md` 8, `rounded-2xl` 8; `shadow-sm` 63 … `shadow-2xl` 1) | §1.5 | Kartu/modal/badge tidak seragam |
| 7 | **0 antd `<Form>` / `Form.Item`, 0 `Typography`, 0 `Input.Search`, 0 `Modal.confirm`**; 25 `<form>` native, 36 `ds-label` | §1.4, §1.6 | Form/label/pencarian/konfirmasi masih rakitan |
| 8 | **51 `disabled={…isPending}` vs hanya 9 `loading`** pada antd Button | §1.1 | State loading tidak konsisten (spinner vs tombol mati) |
| 9 | **71 `style={{…}}` inline** (mayoritas `width: "100%"` berulang di Select) + 6 hex mentah `#ffffff` di style | §1.2 | Override tersebar, sulit dirawat |
| 10 | **Label login masih Inggris** (12 titik) + `size="small"` Button di 4 halaman + ikon 2 set (lucide 43 file vs antd icons 8 file) | §1.1, §1.7 | Inkonsistensi bahasa/ukuran/ikon |

**Jumlah pelanggaran terhitung (struktural, di luar kemunculan class warna): 211 instance**
= 19 file komponen custom + 26 elemen HTML mentah (button 16 + select 1 + radio/checkbox 4 +
table 3 + input hidden/file 2) + 3 dialog bermasalah (2 `fixed` + 1 duplikat Modal) +
6 hex `#ffffff` + 71 inline `style` + 45 class `.ds-*` legacy + 25 `<form>` native +
12 label Inggris login + 4 duplikat `__badgeColor`. **Ditambah ~700 kemunculan class
warna non-token** dan **7 sistem warna status paralel** (§1.2). Detail per kategori di §1.

---

## 1. TEMUAN PELANGGARAN KONSISTENSI

### 1.1 Tombol

**Keadaan yang sudah benar (dipertahankan):** 206 `<Button>` antd di seluruh halaman;
pola `type="primary"` untuk aksi utama, `type="default"` untuk sekunder, `danger`
untuk hapus — konsisten di halaman master (contoh `pages/admin/master/KelasPage.tsx:160,215`).

**Pelanggaran:**

| # | Pelanggaran | Bukti `file:baris` |
|---|---|---|
| T1 | 16 `<button>` mentah (bukan antd Button) | `components/ui/Dialog.tsx:48`, `components/ui/Button.tsx:34` (di dalam komponen custom itu sendiri), `components/ui/Dropdown.tsx:209`, `components/soal/RichEditor.tsx:282,332,367,388,395,420,433,449` (8 tombol toolbar/panel rumus/video/gambar), `components/common/SearchInput.tsx:51`, `components/common/Pagination.tsx:41,50,61` (3 tombol), `components/exam/NavigasiSoal.tsx:32` (grid nomor soal) |
| T2 | 4 `<Button>` custom `ui/Button` masih dipakai (bukan antd) | `pages/GantiPasswordPage.tsx:114` (`type="submit"` — BUKAN `type` antd; ini `type` HTML), `pages/NotFoundPage.tsx:18`, `pages/ForbiddenPage.tsx:22`, `pages/ErrorBoundary.tsx:53,58` |
| T3 | `size="small"` pada Button antd (melanggar 1 pola ukuran) — 10 titik | `pages/admin/backup/BackupPage.tsx:201,204,207,215,218`, `pages/admin/event/EventPage.tsx:179,182`, `pages/admin/event/RuanganPage.tsx:150,153`, `pages/admin/kartu/KartuPesertaPage.tsx:106,331`, `components/exam/BannerPeringatan.tsx:32` |
| T4 | `size="large"` hanya di sebagian halaman siswa + login; mayoritas Button admin tanpa `size` (middle) — tidak ada aturan kapan large vs middle | `pages/siswa/UjianPage.tsx:351,354,393,421,428,635,643,658`, `pages/siswa/SiswaHomePage.tsx:149`, `pages/LoginPage.tsx:112` vs seluruh Button admin tanpa `size` |
| T5 | State loading tidak konsisten: **51** Button memakai `disabled={…isPending}` (tombol mati tanpa spinner) vs **hanya 9** memakai `loading` | Contoh `disabled`: `pages/admin/master/KelasPage.tsx:215`, `pages/admin/backup/BackupPage.tsx:207,239,260`, `pages/admin/event/EventPage.tsx:230`, `pages/admin/event/RuanganPage.tsx:200`; contoh `loading` yang benar: `pages/admin/banksoal/BankSoalDetailPage.tsx:277,284,291`, `pages/LoginPage.tsx:112` |
| T6 | Tombol tautan rakitan (class `inline-flex … rounded-lg border …`) menduplikasi Button | `pages/admin/banksoal/ImportSoalPanel.tsx:696` (`<Link>` "Lihat Panduan"), `pages/admin/banksoal/ImportSoalPanel.tsx:993` (`<Link>` "Kembali ke bank soal"), `pages/admin/DashboardPage.tsx:46-58` (13 `<Link>` menu cepat bergaya tombol-tautan) |
| T7 | Tombol ikon tanpa pola tunggal: ada `icon={<…/>}` antd, ada ikon sebagai children lucide, ada `type="text"` + ikon | `pages/siswa/UjianPage.tsx:635` (`icon={<ChevronLeft/>}`) vs `pages/siswa/SiswaHomePage.tsx:102` (`icon={<IdCard/>}`) vs `pages/admin/banksoal/BankSoalDetailPage.tsx:243` (`<Button><X/></Button>`) |
| T8 | Varian `danger` tidak seragam: hapus baris kadang `type="default" danger`, kadang `type="primary" danger`, kadang `type="primary" danger size="large"` di modal | `pages/admin/master/KelasPage.tsx:160` (default danger) vs `pages/admin/master/KelasPage.tsx:215` (primary danger) vs `pages/admin/backup/BackupPage.tsx:239,260` (primary danger large) vs `pages/admin/event/EventPage.tsx:182` (primary danger small) |
| T9 | 2 dialog `fixed inset-0 … bg-black/40` rakitan (bukan antd Modal) di BankSoalDetail | `pages/admin/banksoal/BankSoalDetailPage.tsx:238` (editor soal), `:260` (pratinjau soal) — keduanya dengan tombol Tutup antd di dalamnya, tapi overlay + panel + Esc + fokus tidak memakai antd |
| T10 | Modal konfirmasi hapus duplikat identik (2× Modal `open={hapusPaketId !== null}` yang sama) | `pages/admin/paket/PaketPage.tsx:256` dan `:324` |
| T11 | 0 `Modal.confirm` — semua konfirmasi hapus memakai `<Modal footer={null}>` + tombol manual (±15 modal) | `grep "Modal.confirm" = 0`; contoh `pages/admin/master/KelasPage.tsx:206`, `pages/admin/banksoal/BankSoalDetailPage.tsx:273,280,287`, `pages/admin/jadwal/JadwalPage.tsx:357`, `pages/admin/backup/BackupPage.tsx:225,249` |
| T12 | `type="submit"` pada Button custom GantiPassword (`pages/GantiPasswordPage.tsx:114`) — di antd, `type` berarti varian; submit adalah `htmlType="submit"`. Ini jebakan migrasi. | `pages/GantiPasswordPage.tsx:114` |

### 1.2 Warna

**Satu-satunya sumber kebenaran yang benar:** `ConfigProvider` di `App.tsx:83`
(`colorPrimary: "#1e4e9e"`, `borderRadius: 8`, `fontFamily: "Plus Jakarta Sans"`)
+ token Tailwind `@theme` di `frontend/src/index.css:89-132`. Semua hex di bawah ini
HARUS dipetakan ke token (lihat standar §2.2).

**Pelanggaran:**

| # | Pelanggaran | Bukti `file:baris` |
|---|---|---|
| W1 | Hex mentah `#ffffff` di inline `style` (6 titik) — seharusnya `token.colorWhite` / class antd / token | `components/layout/Topbar.tsx:66,90`, `components/Layout.tsx:54,75`, `components/AdminLayout.tsx:287,304` (semua `style={{ color: "#ffffff" }}` pada tombol/menu pengguna di header biru) |
| W2 | **~700 kemunculan class warna Tailwind non-token**: `text-slate-*` 386, `bg-slate-*` 51, `border-slate-*` 117, plus keluarga `red-*`, `green-*`, `amber-*`, `blue-*`, `orange-*`, `emerald-*` (~150) | Contoh terbanyak: `text-slate-500` 116, `text-slate-600` 100, `text-slate-700` 96, `text-slate-900` 49, `border-slate-200` 72, `border-slate-300` 33, `text-red-700` 29, `bg-red-50` 29, `border-red-200` 27, `bg-slate-100` 26, `text-green-800` 8, `text-blue-100` 8, `bg-green-100` 6, `bg-blue-50` 5, `bg-amber-100` 5. Daftar lengkap via `grep -rhoE "(bg\|text\|border\|ring)-…"` |
| W3 | **7 sistem warna status paralel** (status yang sama bisa beda warna antar halaman) | (a) antd `Tag color="blue/green/gold/red/default/purple"` — 30 `<Tag>`, warna `blue` dipakai untuk SEMUA status di `pages/siswa/SiswaHomePage.tsx:141,142` (status + attempt_status sama-sama biru); (b) `__badgeColor` diduplikasi identik di **4 file**: `pages/admin/master/SiswaPage.tsx:39`, `pages/admin/ukk/UkkListPage.tsx:18`, `pages/admin/ukk/UkkDetailPage.tsx:43`, `pages/admin/analisis/AnalisisButirPage.tsx:84`; (c) `StatusBadge` + PETA 30 status → `Badge` custom: `components/common/StatusBadge.tsx:13-44`; (d) `Badge` custom 9 varian: `components/ui/Badge.tsx:12-22`; (e) span inline: `pages/admin/banksoal/BankSoalListPage.tsx:95` (`bg-warning-bg` vs `bg-green-100` — "Terbuka" memakai green-100, bukan token success-bg); (f) `DayaBedaBadge`: `components/analisis/DayaBedaBadge.tsx:17-28` (`emerald-100/green-100/amber-100/orange-100/red-100`); (g) `KompetensiHeatmap.warnaCell`: `components/analisis/KompetensiHeatmap.tsx:21-29` (`emerald-500/green-400/amber-400/orange-400/red-400` — palet BERBEDA dari DayaBedaBadge untuk makna "baik/buruk" yang sejenis); (h) span inline PengecohTable: `components/analisis/PengecohTable.tsx:59,63` (`bg-green-100`/`bg-red-100`) |
| W4 | Warna primer ganda di komentar/kode lama: `#2563EB`, `#1d4ed8` disebut di `index.css:98` (komentar larangan — benar sebagai larangan, tapi bukti ada 2 nilai primer liar yang pernah dipakai); `--color-primary-dark #163d7d` vs `--color-anbk-dark #173d7a` — **dua nilai hover berbeda** untuk satu makna "hover primer" | `frontend/src/index.css:90` (`#163d7d`) vs `:100` (`#173d7a`); pemakaian: `bg-primary-dark` 2× vs `bg-anbk-dark` 6× |
| W5 | `--color-accent: #2f6fd0` di `index.css:102` — warna ke-3 di luar palet (tidak dipakai di class manapun yang terhitung; token yatim) | `frontend/src/index.css:102` |
| W6 | Warna-warna hardcode di `index.css` untuk `.ds-alert-*` dan `.rich-content` yang bukan token: `#bbf7d0/#166534`, `#fde68a/#92400e`, `#fecaca/#991b1b`, `#bfdbfe/#1e40af`, `#1e293b`, `#94a3b8`, `#0f172a/#e2e8f0` (pre code), `#f8fafc` | `frontend/src/index.css:175,187,199,201,204,206,209,211,214,216,295,296,351,361,385` — seluruhnya harus dipetakan ke token saat `.ds-*` dihapus (§2.2); `.rich-*` (editor) dipertahankan tapi warnanya dipetakan ke token |
| W7 | `print.css` memakai `#fff/#000` — **DIIZINKAN** (print murni, bukan UI) | `frontend/src/print.css:29,30,58,61,75` |
| W8 | `Timer` memakai `Tag color="error"/"blue"` + `style fontSize 22` — warna sudah via antd, tapi ukuran inline | `components/exam/Timer.tsx:57` |
| W9 | `DistribusiChart` memakai `var(--color-*)` SVG — sudah token, **contoh benar**, dipertahankan | `components/analisis/DistribusiChart.tsx` (`fill="var(--color-anbk)"` dst.) |

**Yang sudah benar:** tidak ada CDN (grep `cdn|unpkg|jsdelivr` = 0 di luar `<Link>` router);
tidak ada logo/nama instansi resmi; disclaimer "Bukan aplikasi ujian nasional resmi."
ada di `components/Layout.tsx` (Footer), `components/AdminLayout.tsx` (Footer),
`pages/LoginPage.tsx` (di bawah kartu).

### 1.3 Tipografi

**Pelanggaran:**

| # | Pelanggaran | Bukti `file:baris` |
|---|---|---|
| Y1 | Judul halaman 5+ pola: `ds-page-title` (24px + aksen kiri biru, hanya via PageHeader), `text-2xl font-bold text-anbk` + `border-b-2 border-anbk`, `text-2xl font-bold text-slate-900`, `text-xl font-bold text-slate-900`, `text-xl font-bold text-ink`, `text-lg font-bold` | `components/common/PageHeader.tsx:17` (ds-page-title); `pages/siswa/UjianPage.tsx:373` (`text-xl`), `:337` (`text-xl`); `pages/siswa/HasilPage.tsx` (`text-xl`); `pages/admin/banksoal/BankSoalDetailPage.tsx:241,263` (`text-lg font-bold text-anbk` — judul modal rakitan); `pages/siswa/KartuSayaPage.tsx` (`text-xl font-bold uppercase`); `pages/NotFoundPage.tsx` (`text-xl`), `pages/ForbiddenPage.tsx` (`text-xl`) |
| Y2 | Ukuran arbitrer `text-[15px]` / `text-[16px]` / `text-[17px]` tersebar (body soal 16px, form 15px, header ujian 17px) tanpa skala | `components/exam/SoalRenderer.tsx` (`text-[16px]` ×5), `components/soal/SoalRenderer.tsx` (`text-[16px]` ×4), `pages/siswa/UjianPage.tsx:338,343,345,502,571` (`text-[16px]/[17px]`), `pages/siswa/KartuSayaPage.tsx:218` (`text-[15px]`), `pages/admin/paket/PaketDetailPage.tsx` (`text-[15px]`), `components/common/ConfirmDialog.tsx:44` (`text-[15px]`) |
| Y3 | 0 `Typography` antd (0 `Typography.Title/Text/Paragraph`) — judul/teks masih `<h1>/<p>` + class Tailwind | `grep Typography = 0` |
| Y4 | Angka raksasa 404/403 `text-6xl font-bold` rakitan (harusnya antd `Result`) | `pages/NotFoundPage.tsx` (`text-6xl`), `pages/ForbiddenPage.tsx` (`text-6xl`) |
| Y5 | Bobot campur: `font-bold` 69, `font-semibold` 101, `font-medium` 124, `font-normal` 2 — tanpa aturan kapan memakai yang mana (label kadang `font-medium`, kadang `font-semibold`, kadang `ds-label`) | Hitungan grep; contoh label: `pages/admin/paket/PaketFormPage.tsx:74` (`text-sm font-medium`) vs `pages/admin/ukk/UkkListPage.tsx:175` (`ds-label`) vs `pages/admin/banksoal/BankSoalFormPage.tsx` (`ds-label` via form) |
| Y6 | `tracking-wide/widest` dipakai tidak konsisten pada brand/kartu | `pages/siswa/UjianPage.tsx:333,370,502` (`tracking-wide`), `pages/siswa/KartuSayaPage.tsx` (`tracking-wide/widest`) |
| Y7 | Font soal: `text-[16px]` di SoalRenderer vs body antd default 14px — perbedaan ini **sengaja** (keterbacaan soal) dan dipertahankan sebagai pengecualian tercatat (§2.3) | `components/exam/SoalRenderer.tsx`, `components/soal/SoalRenderer.tsx` |

### 1.4 Spacing

| # | Pelanggaran | Bukti |
|---|---|---|
| S1 | Gap grid bervariasi tanpa skala: `gap-1` 51, `gap-2` 147, `gap-3` 59, `gap-4` 38, `gap-6` 7, `gap-8` 1, `gap-10` 2 | Hitungan grep; contoh: halaman admin `gap-6` (`pages/admin/DashboardPage.tsx:30`), form `gap-4`/`space-y-4`, grid soal `gap-1.5`/`gap-2` |
| S2 | Padding bervariasi: `p-1` 56, `p-2` 153, `p-3` 76, `p-4` 85, `p-5` 10, `p-6` 30, `p-8` 3, `p-10` 3, `p-36` 1; `space-y-1..6` semua dipakai | Hitungan grep; contoh ekstrem: `p-36` (1×) dan `p-10` (3×) — bukan kelipatan skala 8 yang konsisten |
| S3 | Padding kartu tidak satu: antd Card default (24px) vs Card custom `p-5` (20px) vs `p-4`/`p-6`/`p-8` manual di halaman | `components/ui/Card.tsx:14` (`p-5`), `pages/ForbiddenPage.tsx` (`p-8` pada Card custom), `pages/admin/banksoal/BankSoalDetailPage.tsx:240,262` (`p-4 sm:p-6` pada dialog rakitan), `pages/admin/banksoal/ImportSoalPanel.tsx:607,649,708,763` (`p-3 sm:p-4`) |
| S4 | Padding konten halaman: admin `px-4 py-6 sm:px-6` vs siswa `px-4 py-6` — hampir sama tapi ditulis terpisah di 2 layout | `components/AdminLayout.tsx:311`, `components/Layout.tsx:84` |
| S5 | `.ds-form` (gap 16px) vs `space-y-*` vs `flex flex-col gap-*` — 3 pola jarak form | `frontend/src/index.css` (`.ds-form`), `pages/LoginPage.tsx:77` (`ds-form`), form lain memakai `space-y-3/4` atau `gap-3/4` |
| S6 | `ds-actions` (baris tombol aksi) hanya dipakai 1× (`ConfirmDialog`) — pola aksi modal tidak seragam; modal lain memakai `flex flex-col gap-2 sm:flex-row sm:justify-end` manual | `components/common/ConfirmDialog.tsx` vs `pages/admin/banksoal/BankSoalDetailPage.tsx:275,282,289` (pola manual, 3× identik) |

### 1.5 Radius, shadow, header/kartu/empty/modal/badge/filter/breadcrumb

| # | Pelanggaran | Bukti `file:baris` |
|---|---|---|
| R1 | Radius 6 varian: `rounded-lg` 109, `rounded-xl` 49, `rounded-full` 21, `rounded` 11, `rounded-md` 8, `rounded-2xl` 8, `rounded-t` 6, `rounded-sm` 3 | Hitungan grep. `rounded-2xl` hanya di dialog rakitan (`components/ui/Dialog.tsx:41`, `pages/admin/banksoal/BankSoalDetailPage.tsx:240,262`); `rounded-xl` mendominasi kartu/table wrapper manual; antd token = 8 (`rounded-lg`) |
| R2 | Shadow 7 varian: `shadow-sm` 63, `shadow-xl` 7, `shadow-none` 3, `shadow-md` 1, `shadow-lg` 1, `shadow-2xl` 1, `shadow` 1; shadow arbitrer `shadow-[0_-2px_10px_rgba(0,0,0,0.08)]` di bar bawah ujian | `pages/siswa/UjianPage.tsx:631` (shadow arbitrer + `bg-white/95 backdrop-blur`); `components/ui/Dialog.tsx:41` (`shadow-2xl`); `pages/admin/banksoal/BankSoalDetailPage.tsx:240,262` (`shadow-xl`) |
| R3 | Header halaman: PageHeader custom (27 pemakaian, 26 file) + Breadcrumbs custom (27 pemakaian) — bukan antd `PageHeader`/`Typography` + `Breadcrumb` | `components/common/PageHeader.tsx` (23 baris, `ds-page-title` + `border-b`); `components/layout/Breadcrumbs.tsx` (52 baris, wrapper antd Breadcrumb). Daftar pemakai PageHeader: Rekap, BeritaAcara, Settings, Kelas, Mapel, Guru, ImportSiswa, Siswa, Jurusan, PanduanImport, Monitor, Paket, PaketForm, PaketDetail, Jadwal, JadwalForm, Backup, UkkList, UkkDetail, UkkNilai, Kompetensi, AnalisisButir, Event, Ruangan, Kartu, Dashboard (26 file). Halaman TANPA PageHeader/Breadcrumbs: seluruh halaman siswa (Hasil, Ujian, SiswaHome, KartuSaya), Login, GantiPassword, NotFound, Forbidden, ErrorBoundary, BankSoalList, BankSoalForm, BankSoalDetail, ImportSoalPanel — header mereka rakitan sendiri |
| R4 | Kartu: 3 sistem — antd `Card` (46 `<Card>`), Card custom `ui/Card` (2 pemakai), `div rounded-xl border … bg-white shadow-sm` rakitan (puluhan) | Card custom: `pages/GantiPasswordPage.tsx:12`, `pages/ForbiddenPage.tsx:7`; rakitan: `pages/admin/banksoal/ImportSoalPanel.tsx:607,649,708,763`, `pages/siswa/UjianPage.tsx:369` (`rounded-xl bg-anbk … shadow-sm`) |
| R5 | Empty state: antd `<Empty>` 5× vs `EmptyState` custom 2× vs `locale={{ emptyText }}` di Table (29×) vs `<p>` "Tidak ada …" manual | antd Empty: `pages/siswa/SiswaHomePage.tsx:120` + 4 lain; EmptyState: `pages/admin/paket/PaketPage.tsx`, `pages/admin/jadwal/JadwalPage.tsx`; manual: `pages/siswa/UjianPage.tsx:593`, `components/analisis/DistribusiChart.tsx:17`, `components/analisis/KompetensiHeatmap.tsx:31` |
| R6 | Modal: antd `<Modal>` 27× (mayoritas `footer={null}` + tombol manual, width tak seragam: default ~19×, `width={520}` 5×, `width={720}` 3×) + 2 dialog `fixed` rakitan + 1 duplikat | Width: `pages/admin/jadwal/JadwalPage.tsx:313` (520), `pages/admin/backup/BackupPage.tsx:225,249` (520), `pages/admin/ukk/UkkListPage.tsx:173` (720), `pages/admin/event/EventPage.tsx:190,221`, `pages/admin/event/RuanganPage.tsx:161,191` (720/520) |
| R7 | Filter: tidak ada `Input.Search` (0×); pencarian memakai antd `Input type="search"` + `Input` biasa + `SearchInput` custom (0 pemakai) | `grep "Input.Search" = 0`; contoh pencarian: `pages/admin/master/KelasPage.tsx:135` (`Input … size="large"`), `pages/admin/paket/PaketPage.tsx:285` (`Input type="search"`), `pages/admin/jadwal/JadwalPage.tsx:244` |
| R8 | Pagination: antd `<Pagination>` 5× (hanya halaman master) vs `Pagination` custom 0 pemakai vs Table `pagination={false}` di ~24 tabel + Pagination manual di Backup (`size="small"` prev/next) | antd Pagination: `pages/admin/master/KelasPage.tsx:166`, `MapelPage.tsx:178`, `GuruPage.tsx:158`, `SiswaPage.tsx:217`, `JurusanPage.tsx:162`; manual: `pages/admin/backup/BackupPage.tsx:215,218` |
| R9 | `ds-header-band` (pita header biru kartu login) dipakai 2× dengan class custom, bukan antd Card `title`/cover | `pages/LoginPage.tsx` (`.ds-header-band` + `bg-white/15`), `pages/GantiPasswordPage.tsx` (sama) |
| R10 | `ds-table-wrap` hanya 1× — mayoritas tabel dibungkus `div overflow-x-auto rounded-xl border … bg-white` manual (duplikasi wrapper di ~20 halaman) | `grep "ds-table-wrap" = 1`; contoh wrapper manual: `pages/admin/master/KelasPage.tsx:154`, `pages/admin/rekap/RekapPage.tsx:202`, `pages/admin/monitor/MonitorPage.tsx:279` |

### 1.6 Komponen custom yang menduplikasi antd (daftar lengkap + importer + rencana)

| File custom | Baris | Importer (file yang mengimpor) | Rencana (lihat §2.8) |
|---|---|---|---|
| `components/ui/Button.tsx` | 41 | 7 file: `pages/GantiPasswordPage.tsx`, `pages/NotFoundPage.tsx`, `pages/ForbiddenPage.tsx`, `pages/ErrorBoundary.tsx`, `components/common/EmptyState.tsx`, `components/common/ConfirmDialog.tsx`, `components/common/PrintLayout.tsx` | Migrasi ke antd `Button` → **HAPUS** |
| `components/ui/Card.tsx` (+ `StatCard` di dalamnya) | 60 | 2 file: `pages/GantiPasswordPage.tsx`, `pages/ForbiddenPage.tsx` | Migrasi ke antd `Card` → **HAPUS** |
| `components/ui/Input.tsx` | 34 | 1 file: `pages/GantiPasswordPage.tsx` | Migrasi ke antd `Input`/`Input.Password` → **HAPUS** |
| `components/ui/Select.tsx` | 65 | 0 importer langsung (wrapper atas `Dropdown`); `Dropdown` diimpor oleh `Select` itu sendiri | Migrasi ke antd `Select` → **HAPUS** bersama Dropdown |
| `components/ui/Table.tsx` | 46 | **0 importer** | **HAPUS langsung** (bukti grep 0 + tsc lolos) |
| `components/ui/Dialog.tsx` | 61 | 1 file: `components/common/ConfirmDialog.tsx` | Migrasi ke antd `Modal`/`Modal.confirm` → **HAPUS** |
| `components/ui/Badge.tsx` | 36 | 1 file: `components/common/StatusBadge.tsx` | Migrasi ke antd `Tag` → **HAPUS** |
| `components/ui/Dropdown.tsx` | 353 | 1 file: `components/ui/Select.tsx` | Migrasi ke antd `Select showSearch` → **HAPUS** |
| `components/common/PageHeader.tsx` | 23 | 26 file (27 pemakaian `<PageHeader>`, §1.5 R3) | Migrasi ke antd `Breadcrumb` + `Typography.Title` → **HAPUS** |
| `components/common/StatusBadge.tsx` | 54 | 2 pemakaian: `pages/admin/monitor/MonitorPage.tsx`, `pages/admin/jadwal/JadwalPage.tsx` | Migrasi ke antd `Tag` + peta status §2.2 → **HAPUS** |
| `components/common/EmptyState.tsx` | 33 | 2 pemakaian: `pages/admin/paket/PaketPage.tsx`, `pages/admin/jadwal/JadwalPage.tsx` | Migrasi ke antd `Empty` → **HAPUS** |
| `components/common/ConfirmDialog.tsx` | 64 | **0 pemakai** | **HAPUS langsung** (ganti pola konfirmasi ke `Modal.confirm` di pemanggil yang butuh) |
| `components/common/Pagination.tsx` | 68 | **0 pemakai** | **HAPUS langsung** |
| `components/common/SearchInput.tsx` | 62 | **0 pemakai** | **HAPUS langsung** (pencarian → antd `Input.Search`) |
| `components/common/StatCard.tsx` | 54 | **0 pemakai** | Migrasi ke antd `Card` + `Statistic` → **HAPUS langsung** |
| `components/common/PrintLayout.tsx` | 65 | **0 pemakai** | **HAPUS langsung** (halaman cetak memakai pola sendiri; bila dibutuhkan lagi, bangun ulang dengan antd) |
| `components/layout/Sidebar.tsx` | 102 | **0 pemakai** (AdminLayout memakai `AdminSidebar` sendiri) | **HAPUS langsung** (duplikat `AdminSidebar` di `AdminLayout.tsx`) |
| `components/layout/Topbar.tsx` | 99 | **0 pemakai** (AdminLayout/Layout punya header sendiri) | **HAPUS langsung** (duplikat header) |
| `components/layout/Breadcrumbs.tsx` | 52 | 26 file (27 pemakaian) | Migrasi ke antd `Breadcrumb` langsung → **HAPUS** |

**Total: 19 file komponen custom.** 8 di antaranya sudah **0 importer/pemakai**
(`ui/Table`, `common/ConfirmDialog`, `common/Pagination`, `common/SearchInput`,
`common/StatCard`, `common/PrintLayout`, `layout/Sidebar`, `layout/Topbar`) dan dapat dihapus begitu
tim eksekusi berjalan, dengan bukti `grep 0 importer + tsc lolos`.

**Bukan komponen custom duplikat (TETAP, bukan pengecualian — memang bukan komponen UI):**
`components/soal/RichEditor.tsx` (area editor TipTap), `components/soal/RichContent.tsx`
(render sanitasi + KaTeX), `lib/format.ts` (fungsi `formatTanggal`).
**Komponen domain yang TETAP (bukan duplikat antd, berisi logika/status domain):**
`components/exam/Timer.tsx` (Tag antd di dalam), `components/exam/BannerPeringatan.tsx`
(Alert antd), `components/exam/NavigasiSoal.tsx` (harus diubah ke antd Button, §2.9),
`components/exam/SoalRenderer.tsx` (interaktif siswa; `<select>` harus ke antd Select, §2.10),
`components/soal/SoalEditor.tsx` (editor guru; radio/checkbox kunci + tombol → antd),
`components/soal/SoalRenderer.tsx` (read-only guru), `components/analisis/*` (harus ke
antd Table/Tag, §2.2), `components/common/NotifikasiBell.tsx` (sudah antd Badge/Popover/Button),
`components/ui/ImageUpload.tsx` (sudah antd Upload/Button/Progress/Alert/Image),
`components/Layout.tsx` + `components/AdminLayout.tsx` (kerangka antd Layout — **tetap**,
bukan duplikat `layout/Sidebar`/`Topbar`).

### 1.7 Responsif (titik yang dicurigai meluber — untuk Tim Responsif)

Semua titik di bawah adalah **dugaan berbasis kode** (belum diuji di browser dalam audit ini).
Tim Responsif WAJIB memverifikasi di 360 / 768 / 1280 px dan memperbaiki yang terbukti meluber.

| # | Titik | Bukti `file:baris` | Dugaan |
|---|---|---|---|
| P1 | Bar bawah ujian (sticky): 2 Button + Checkbox ragu-ragu + Button Selesai dalam `flex-wrap justify-between` | `pages/siswa/UjianPage.tsx:631-662` | Di 360px, Checkbox (min 44px + teks 16px) + 3 tombol bisa bertumpuk tidak rapi / menekan area soal; `shadow-[0_-2px…]` arbitrer |
| P2 | Bilah info peserta ujian: `<dl>` flex-wrap nama/NISN/kelas/mapel/soal + Timer Tag 22px | `pages/siswa/UjianPage.tsx:499-532` | Di 360px, 5 pasangan dt/dd + timer bisa meluber vertikal dan memakan layar; `text-[17px]` brand + `-mx-4` sticky |
| P3 | Grid nomor soal: `grid-cols-4 gap-1.5`, tombol min 44px | `components/exam/NavigasiSoal.tsx:24` | 4×44 + gap = ~190px — muat di 360, tapi di 320px (komentar menyebut 320px) mepet; legenda `flex-wrap` perlu cek |
| P4 | Import preview: Table `scroll={{ x: 1000 }}` dalam Card | `pages/admin/banksoal/ImportSoalPanel.tsx:833-843` | Di 360px scroll horizontal 1000px di dalam kartu — pastikan scroll berfungsi, header tidak terpotong, tombol Simpan tidak terdorong |
| P5 | Dialog `fixed` editor/pratinjau soal: `p-4`, `max-w-3xl`, `rounded-2xl`, tanpa `max-h` + scroll yang jelas pada panel dalam | `pages/admin/banksoal/BankSoalDetailPage.tsx:238,260` | Di 360px / tinggi kecil, panel bisa melebihi viewport; setelah migrasi ke antd Modal, verifikasi ulang |
| P6 | Modal `width={720}` (UKK, Event, Ruangan) di 360px | `pages/admin/ukk/UkkListPage.tsx:173`, `pages/admin/event/EventPage.tsx:190`, `pages/admin/event/RuanganPage.tsx:161` | antd Modal default membatasi, TAPI wajib verifikasi 360px (form di dalam: Input + Select berdampingan?) |
| P7 | Tabel antd `scroll={{ x: "max-content" }}` + wrapper `div overflow-x-auto` ganda | ±20 halaman, contoh `pages/admin/master/KelasPage.tsx:154,164` | Scroll ganda (wrapper + Table) — di HP bisa terjadi scroll bersarang; standarkan ke Table `scroll` saja (§2.7) |
| P8 | Kartu peserta cetak: `grid-cols-2` tanda tangan + `w-40` garis + `min-w-[320px]` tabel | `pages/siswa/KartuSayaPage.tsx:217-218, 246-256` | Di 360px, 2 kolom tanda tangan + garis 160px bisa mepet; ini dokumen cetak — verifikasi tampilan layar 360 + hasil print |
| P9 | `BankSoalListPage`: tampilan ganda kartu (HP) + tabel (`hidden lg:block`) | `pages/admin/banksoal/BankSoalListPage.tsx:86` | Satu-satunya halaman dengan pola ganda — tidak konsisten dengan halaman list lain (Table scroll). Diputuskan: pola ganda DIHAPUS, samakan ke Table scroll (§2.7) |
| P10 | Sidebar: `Sider width={256}` + Drawer HP `width={288}` — sudah benar; JANGAN dirusak pembungkus `flex-1 overflow-y-auto` | `components/AdminLayout.tsx:234-262` | Verifikasi: 19 menu di layar pendek (1366×768 ke bawah) tetap bisa di-scroll; Drawer menutup setelah navigasi |
| P11 | `NotifikasiBell` Popover `width: 320, maxWidth: calc(100vw - 2rem)` | `components/common/NotifikasiBell.tsx:72` | Sudah responsif — contoh benar, dipertahankan |
| P12 | Form berkolom: `grid-cols-2` tanpa `sm:` di beberapa form (cek `grid-cols-2` 34×) — kolom ganda di 360px bisa sempit | Hitungan `grid-cols-2` = 34 pemakaian; contoh `components/soal/SoalRenderer.tsx:108` (`grid-cols-2 sm:grid-cols-4` — benar) vs yang tanpa breakpoint | Tim Responsif: setiap `grid-cols-2` tanpa breakpoint harus jadi `grid-cols-1 sm:grid-cols-2` |
| P13 | `UjianPage` layout utama: `flex` soal + navigasi (`lg:w-72`, `lg:sticky lg:top-36`) + `showNav` toggle | `pages/siswa/UjianPage.tsx:604-612` | Verifikasi: navigasi tersembunyi di HP sampai toggle ditekan; konten soal tidak terdorong |

### 1.8 Label bahasa (Inggris yang masih tampil ke user)

| # | Lokasi | Bukti `file:baris` | Seharusnya |
|---|---|---|---|
| L1 | Login: label `Email / Username`, placeholder `Email / Username`, label `Password`, placeholder `Password` | `pages/LoginPage.tsx:80,90,95,106` | `Email / Nama pengguna`, `Kata sandi` (§2.11) |
| L2 | GantiPassword: judul `Ganti Password`, label `Email / Username (kosongkan bila sudah login)`, `Password Lama`, `Password Baru`, `Konfirmasi Password Baru`, tombol `Ganti Password`, pesan error `Password baru minimal 8 karakter.` / `Konfirmasi password baru tidak sama.` / `Password baru harus berbeda dari password lama.` / `Password awal harus diganti…` / `Password berhasil diganti…` | `pages/GantiPasswordPage.tsx:66,69,76,83,88,92,97,101,115` + pesan di `:29,32,35,47` | Indonesiakan semua (`Ganti Kata Sandi`, `Kata Sandi Lama/Baru`, `Konfirmasi Kata Sandi Baru`, dst.) — **nama field API (`password_lama`, `password_baru`, `login`) TIDAK berubah** |
| L3 | KartuSaya: label kartu `Username:`, `Password awal:` | `pages/siswa/KartuSayaPage.tsx:232,235` | `Nama pengguna:`, `Kata sandi awal:` |
| L4 | UjianPage: tombol `Kembali ke Dashboard` (siswa tidak punya dashboard) | `pages/siswa/UjianPage.tsx:355` | `Kembali ke Beranda` |
| L5 | **Dipertahankan (bukan pelanggaran):** label menu admin `Dashboard`, breadcrumb `Dashboard`, judul `Settings`, `Backup`, `Event` — nama fitur/route yang dipakai di 27 breadcrumb; tidak diubah dalam audit ini | `components/AdminLayout.tsx` (NAV_GROUPS), `pages/*/…` (breadcrumb `Dashboard`) | Tetap |

### 1.9 Lain-lain

| # | Pelanggaran | Bukti |
|---|---|---|
| X1 | 0 `Modal.confirm`, 0 `Input.Search`, 0 `Form`/`Form.Item`, 0 `Typography` (§1.5) | grep |
| X2 | Ikon 2 set: lucide-react di **43 file** vs `@ant-design/icons` di **8 file** (AdminLayout, Sidebar, Topbar, Layout, Breadcrumbs, NotifikasiBell, ImageUpload, +1) | `grep -rln` |
| X3 | `type="submit"` vs `htmlType` — jebakan di GantiPassword (§1.1 T12) | `pages/GantiPasswordPage.tsx:114` |
| X4 | `Modal` konfirmasi memakai `maskClosable={!(…)}` + `keyboard={!(…)}` manual berulang — pola ini benar tapi ditulis manual di ±10 tempat; `Modal.confirm` menyeragamkannya | Contoh `pages/admin/banksoal/BankSoalDetailPage.tsx:273,280,287` |
| X5 | `Table` antd: `pagination={false}` di ~24 tabel + `<Pagination>` terpisah hanya di 5 halaman master — pola pagination ganda | §1.5 R8 |
| X6 | `Card` Dashboard memakai `border-t-4 border-t-anbk` (aksen atas) — tidak ada di kartu lain | `pages/admin/DashboardPage.tsx:37-40,43` |
| X7 | `ds-label` 36× + label `text-sm font-medium` manual — 2 pola label | §1.3 Y5 |

---

## 2. STANDAR YANG DISEPAKATI

> **Aturan tertinggi:** HANYA lapisan presentasi. JANGAN mengubah route, endpoint API,
> logika bisnis, grading, auth, token, timer, anti-curang, validasi, nama field, payload.
> Bila ragu, jangan ubah — tanya koordinator.
> **Satu sumber kebenaran:** komponen UI dari **Ant Design** + token tema di
> `ConfigProvider` (`App.tsx`). DILARANG warna primer baru / gaya tombol baru di luar sistem.
> Hex mentah di luar token DILARANG — semua hex dipetakan ke token (§2.2).

### 2.1 ConfigProvider (satu-satunya tempat token didefinisikan)

`frontend/src/App.tsx` — dipertahankan + dilengkapi (Tim Token). Nilai yang SUDAH ada
TIDAK berubah maknanya; yang baru hanya melengkapi token antd agar tidak ada hex di komponen:

```tsx
<ConfigProvider
  locale={idID}
  theme={{
    token: {
      colorPrimary: "#1e4e9e",   // TETAP — biru primer
      colorSuccess: "#15803d",   // BARU — hijau sukses (sebelumnya hanya token Tailwind)
      colorWarning: "#b45309",   // BARU — kuning tua ragu-ragu/menunggu
      colorError: "#dc2626",     // BARU — merah bahaya
      colorInfo: "#1e4e9e",      // BARU — info = primer
      colorText: "#0f172a",      // BARU — teks utama (ink)
      colorTextSecondary: "#475569", // BARU — teks sekunder (muted)
      colorBorder: "#cbd5e1",    // BARU — garis pemisah
      colorBgLayout: "#eef2f7",  // BARU — latar halaman (surface)
      borderRadius: 8,           // TETAP
      fontFamily: '"Plus Jakarta Sans", sans-serif', // TETAP
      fontSize: 14,              // BARU — body default antd
      controlHeightLG: 44,       // BARU — target sentuh min 44px (large = 44)
    },
  }}
>
```

Aturan token:
- Komponen antd MEMBACA token lewat `type`/`color`/`status` (mis. `type="primary"`,
  `Tag color="green"`, `Alert type="error"`), BUKAN lewat hex di `style`/`className`.
- Bila sebuah warna status tidak bisa diwakili `type`/`color` antd (mis. latar tombol
  grid soal "sudah dijawab"), ambil dari `theme.useToken()` (`token.colorSuccess`,
  `token.colorWarning`, …) — **bukan** menulis `"#15803d"` di komponen.
- Token Tailwind `@theme` di `index.css` dipertahankan SEMENTARA sebagai jembatan
  selama migrasi, lalu **dihapus** setelah `grep -rhoE "bg-anbk|text-anbk|…"` = 0.
  Yang pertama-tama dihapus nilainya yang ganda: `--color-primary-dark #163d7d`
  (hapus; hover primer SATU nilai: **`#173d7a`**), `--color-accent #2f6fd0` (hapus; yatim).
- DILARANG menambah token baru / warna primer baru tanpa persetujuan koordinator.

### 2.2 Palet (1 palet — nilai TETAP, hanya dipetakan ke token)

| Peran | Nilai | Token antd / pemakaian |
|---|---|---|
| Primer | `#1e4e9e` | `colorPrimary`; `Button type="primary"`, `Tag color="blue"` (info), header |
| Primer hover | `#173d7a` | antd menghitung hover sendiri dari `colorPrimary`; di Tailwind lama `bg-anbk-dark` → hapus setelah migrasi |
| Primer active / pekat | `#122f5e` | `token.colorPrimary` + antd active; `bg-primary-darker` → hapus |
| Primer muda (latar lembut) | `#e8effa` | `bg-primary-light` → hapus; ganti antd `Tag`/`Alert`/`Card` |
| Primer soft | `#dbe6f7` | `bg-primary-soft` → hapus |
| Sukses | `#15803d` / latar `#dcfce7` | `colorSuccess`; `Tag color="green"`, `Alert type="success"` |
| Peringatan | `#b45309` / latar `#fef3c7` | `colorWarning`; `Tag color="gold"`, `Alert type="warning"` |
| Error / bahaya | `#dc2626` / latar `#fee2e2` | `colorError`; `Tag color="red"`, `Alert type="error"`, `Button danger` |
| Info | `#1e4e9e` / latar `#dbeafe` | `colorInfo`; `Tag color="blue"`, `Alert type="info"` |
| Nonaktif | `#64748b` / latar `#f1f5f9` | `Tag color="default"` |
| Teks utama | `#0f172a` | `colorText` |
| Teks sekunder | `#475569` | `colorTextSecondary` |
| Garis | `#cbd5e1` | `colorBorder` |
| Latar halaman | `#eef2f7` | `colorBgLayout` |
| Putih | `#ffffff` | `token.colorWhite` / antd default — **dilarang tulis `#ffffff` di `style`** (ganti 6 titik §1.2 W1) |
| Hitam (print) | `#000` | Hanya `print.css` — diizinkan |

**Peta status → Tag (SATU peta, menggantikan 7 sistem §1.2 W3):**

| Status (teks tampil) | `Tag color` | Dipakai di |
|---|---|---|
| Aktif, Lulus, Dijawab, Terbuka, Terkirim, Disetujui, Terbaca | `green` | Semua halaman |
| Menunggu, Jeda, Ragu-ragu, Terkunci | `gold` | Semua halaman |
| Diblokir, Tidak Lulus, Ditolak | `red` | Semua halaman |
| Berjalan, Selesai, Dikerjakan | `blue` | Semua halaman |
| Draft, Tidak Aktif, Belum Dijawab | `default` | Semua halaman |

- `__badgeColor` di 4 file (§1.2 W3) DIHAPUS — pemanggil memakai peta di atas langsung.
- `DayaBedaBadge` → antd `Tag`: Sangat Baik `green`, Baik `green` (dengan teks nilai),
  Cukup `gold`, Buruk `orange`, Sangat Buruk `red`. Warna preset antd
  (`green/gold/orange/red/blue/purple/default`) BOLEH untuk Tag — itu bagian dari
  sistem antd, bukan hex baru.
- `KompetensiHeatmap` → antd `Table` + sel `Tag`: ≥90 `green`, 75–89 `green` (varian
  teks), 60–74 `gold`, 40–59 `orange`, <40 `red`, null `default` ("—"). Skala 5 tingkat
  lama (`emerald-500/green-400/amber-400/orange-400/red-400`) DIHAPUS — maknanya sama,
  warnanya kini dari preset antd.
- `PengecohTable` → antd `Table` + `Tag` (`Berfungsi` green / `Tidak berfungsi` red).
- Status SELALU warna + teks, bukan warna saja (aksesibilitas — dipertahankan).
- `SiswaHomePage.tsx:141,142` — `Tag color="blue"` untuk status + attempt_status
  diganti mengikuti peta (status ujian → peta; attempt_status diblokir → `red`).

### 2.3 Tipografi (1 skala)

| Elemen | Pola | Nilai |
|---|---|---|
| Judul halaman (H1) | antd `Typography.Title level={3}` (atau `Title` di dalam pola header §2.6) | 24px, 700 |
| Judul kartu | antd `Card title` / `Modal title` (bawaan antd) | 16px, 600 |
| Judul seksi (H2) | antd `Typography.Title level={4}` | 18px, 700 |
| Body | antd default | 14px, 400, `colorText` |
| Teks soal / opsi jawaban | **Pengecualian tercatat:** 16px | Keterbacaan soal — dipertahankan dari `SoalRenderer` (16px). Bukan pola umum |
| Label form | antd `Form.Item label` | 14px, 600 |
| Caption / teks bantu | antd `Typography.Text type="secondary"` | 12px |
| Timer ujian | antd `Tag` + `fontFamily: token.fontFamilyCode` (monospace token) | 22px, 700 (dipertahankan; angka timer harus monospace tabular) |
| Angka besar hasil | antd `Statistic` / `Typography.Title` | Sesuai konteks; `text-4xl` di `UjianPage.tsx:343` diganti `Statistic`/`Title` |
| Halaman error (404/403/kesalahan) | antd `Result` (`status="404"/"403"/"error"`) | Menggantikan `text-6xl` rakitan (§1.3 Y4) |

DILARANG: `text-[15px]/[16px]/[17px]` arbitrer di luar teks soal; `text-xl/text-2xl`
manual untuk judul halaman; `tracking-wide/widest` kecuali brand `CBT SMK` di header
(yang sudah ada, dipertahankan).

### 2.4 Spacing (1 sistem — basis 8px)

| Konteks | Pola |
|---|---|
| Jarak antar blok halaman | `gap-6` (24px) — `flex flex-col gap-6` / antd `Space size={24}` |
| Jarak antar field form | `gap-4` (16px) — via antd `Form` (vertical) |
| Jarak grid kartu / tabel | `gap-4` (16px) |
| Padding konten halaman | Admin: `px-6 py-6` (konten `max-w-1280`); Siswa: `px-4 py-6` (konten `max-w-1100`) — ditulis SATU kali di `AdminLayout`/`Layout`, bukan per halaman |
| Padding Card | antd default (24px) — DILARANG override `p-4/p-5/p-6/p-8` pada Card |
| Padding Modal | antd default |
| Baris tombol aksi (modal/form) | antd `Space` / `Flex justify="flex-end" gap={8}` — menggantikan `ds-actions` + pola manual `flex flex-col gap-2 sm:flex-row sm:justify-end` |
| Target sentuh | min 44px — via `controlHeightLG: 44` + `size="large"` untuk aksi utama; grid nomor soal min 44px (§2.9) |

### 2.5 Radius & shadow (1 pola)

- **Radius: 8px untuk SEMUA** (token `borderRadius: 8`). `rounded-lg` = 8 adalah
  satu-satunya radius komponen. `rounded-xl` (12), `rounded-2xl` (16), `rounded-md` (6),
  `rounded` (4), `rounded-sm` (2) DILARANG untuk komponen UI.
- Pengecualian: `rounded-full` HANYA untuk `Avatar` / lencana ikon lingkaran
  (diwakili antd `Avatar` / `Badge`); `rounded-sm` HANYA di dalam sel heatmap/legenda
  kecil bila masih dipakai (diutamakan `Tag`).
- **Shadow: antd default saja.** `shadow-xl/2xl/lg/md`, shadow arbitrer
  `shadow-[0_-2px_10px_rgba(0,0,0,0.08)]` (§1.5 R2) DIHAPUS — bar bawah ujian memakai
  `border-t` saja, tanpa shadow. `shadow-sm` pada wrapper manual hilang bersama wrappernya.

### 2.6 Pola header / kartu / modal / empty-state / tag / filter / pagination

| Pola | Standar |
|---|---|
| Header halaman | antd `Breadcrumb` (items: Beranda + …; terakhir tanpa link) + antd `Typography.Title level={3}` + deskripsi (`Typography.Text type="secondary"`) + aksi kanan (antd `Space`). Menggantikan `PageHeader` + `Breadcrumbs` custom. Breadcrumb admin selalu mulai dari `Dashboard` (`/adm`). Halaman siswa tanpa breadcrumb admin (judul saja + deskripsi) — konsisten dengan yang sudah berjalan |
| Kartu | antd `Card` (`title`/`extra`/`actions`). DILARANG `border-t-4` aksen (hapus di Dashboard), DILARANG wrapper `div rounded-xl border … bg-white shadow-sm` |
| Statistik | antd `Card` + `Statistic` (`title`, `value`, `prefix`/`suffix`, `loading`). Menggantikan `StatCard` custom + `StatCard` di `ui/Card` |
| Modal | antd `Modal`: `open`, `onCancel`, `title`, `destroyOnClose`. Width: konfirmasi 520, form 720, default 520. Saat loading: `maskClosable={false} keyboard={false}`. Footer: `footer={null}` + tombol antd di dalam, ATAU `onOk/onCancel` + `okText/cancelText` untuk modal sederhana (pola UjianPage: `okText="Ya, Kumpulkan"`) |
| Konfirmasi hapus / berbahaya | antd `Modal.confirm`: `title`, `content`, `okText="Ya, Hapus"` (atau `Ya, Lanjutkan`), `cancelText="Batal"`, `okButtonProps={{ danger: true, loading }}`. Menggantikan ±15 Modal `footer={null}` manual + `ConfirmDialog` (0 pemakai) |
| Empty state | antd `Empty` (`description` Bahasa Indonesia). `Table` memakai `locale={{ emptyText: "…" }}` Bahasa Indonesia. DILARANG teks Inggris `No data` |
| Tag / badge | antd `Tag` + peta §2.2 |
| Filter | Bar filter: antd `Space` + `Input.Search` (`allowClear`, `size="large"`, placeholder `Cari…`) + `Select showSearch` (`placeholder="Pilih…"`, `notFoundContent="Tidak ada data"`, `filterOption` seperti yang sudah ada, `style={{ width: "100%" }}` diganti `block`/`className="w-full"`). Perilaku filter (termasuk debounce bila ada) TIDAK berubah — hanya komponennya |
| Pagination | antd `Pagination` (`showTotal`: `Halaman X dari Y · Total N`, teks ikut `locale={idID}`). Tabel memakai `pagination={false}` + `<Pagination>` terpisah (pola 5 halaman master dipertahankan) ATAU `pagination` bawaan Table bila halaman itu memang memakainya — SATU pola per halaman, jangan campur |
| Alert / pesan | antd `Alert` (`type`, `showIcon`, `message`, `description`). Menggantikan `ds-alert` |
| Deskripsi label-nilai (detail) | antd `Descriptions` (`bordered`, `column={1}` di HP) untuk halaman detail |
| Langkah / wizard | antd `Steps` (sudah dipakai ImportSoalPanel) |
| Upload | antd `Upload` / `Upload.Dragger` (sudah dipakai ImageUpload + ImportSoalPanel) |

### 2.7 Tabel (1 pola)

- antd `Table`: `columns`, `dataSource`, `rowKey`, `pagination={false}` +
  `<Pagination>` terpisah (atau `pagination` bawaan — satu pola per halaman),
  `locale={{ emptyText: "…" }}`, **`scroll={{ x: "max-content" }}`**.
- DILARANG wrapper `div overflow-x-auto rounded-xl border … bg-white` di sekeliling
  Table — scroll ditangani `Table scroll` saja (menghapus scroll ganda §1.7 P7).
  Pengecualian: `ImportSoalPanel` preview `scroll={{ x: 1000 }}` dipertahankan
  (kolom preview banyak; verifikasi Tim Responsif §1.7 P4).
- DILARANG `<table>` mentah untuk data — kecuali **dokumen cetak**
  (kartu peserta, sertifikat, berita acara) yang tata letaknya dokumen, bukan UI
  (`KartuSayaPage.tsx:218` dipertahankan sebagai dokumen cetak; tombol/aksi di
  sekitarnya tetap antd). `KompetensiHeatmap` + `PengecohTable` (data) → antd `Table`.
- DILARANG tampilan ganda kartu-HP + tabel-desktop (`BankSoalListPage.tsx:86`
  pola ganda DIHAPUS; samakan ke Table scroll seperti halaman list lain).

### 2.8 Aturan "antd untuk semuanya" + daftar migrasi + aturan hapus file

**Aturan:** setiap komponen UI HARUS dari antd. Komponen custom yang menduplikasi
komponen antd DILARANG — daftar migrasi wajib (§1.6):

| Dari (custom) | Ke (antd) |
|---|---|
| `PageHeader` | antd `Breadcrumb` + `Typography.Title` + deskripsi + aksi (§2.6) |
| `StatusBadge` | antd `Tag` + peta §2.2 |
| `EmptyState` | antd `Empty` |
| `ui/Button` | antd `Button` |
| `ui/Input` / `ui/Select` | antd `Input` / `Select` |
| `ui/Card` (+ `StatCard` di dalamnya) | antd `Card` (+ `Statistic`) |
| `ui/Table` | antd `Table` |
| `ui/Dialog` | antd `Modal` |
| `common/ConfirmDialog` | antd `Modal.confirm` |
| `common/Pagination` | antd `Pagination` |
| `common/SearchInput` | antd `Input.Search` |
| `common/StatCard` | antd `Card` + `Statistic` |
| `layout/Sidebar` / `Topbar` | antd `Layout` / `Menu` (sudah ada di `AdminLayout`/`Layout` — file duplikat dihapus) |
| `Dropdown` custom (+ wrapper `ui/Select`) | antd `Select` (`showSearch`, `allowClear`) |
| `ui/Badge` | antd `Tag` |
| `common/Breadcrumbs` (wrapper) | antd `Breadcrumb` langsung |
| `common/PrintLayout` | antd `Button` + `Card` (bila dibutuhkan lagi; saat ini 0 pemakai → hapus) |

**Aturan hapus file 0-importer:** setelah migrasi, file komponen custom yang
`grep -rln "<nama-komponen>" + import` = **0 importer** WAJIB DIHAPUS, dengan bukti:
(1) `grep` 0 importer, (2) `tsc` lolos (atau `vite build` lolos), (3) halaman pemakai
lama sudah diverifikasi tampil. 7 file sudah 0-importer hari ini (§1.6) dan dihapus
di Fase 1. DILARANG menghapus file yang masih punya importer.
**Yang TETAP bukan antd (bukan komponen UI, jadi bukan pengecualian):** area editor
TipTap `RichEditor` (konten editor), render KaTeX (rumus), helper `formatTanggal`
(fungsi, bukan komponen). Semua yang lain — termasuk halaman
GantiPassword/Forbidden/NotFound/ErrorBoundary — WAJIB jadi antd.
**Ikon:** satu set `@ant-design/icons` untuk semua kode baru/diubah; dilarang mencampur
lucide + antd icons dalam SATU file (lucide di 43 file dimigrasi per-file saat file itu
disentuh Tim Komponen/Layout; jangan migrasi ikon massal tanpa verifikasi tampilan).

### 2.9 Grid nomor soal ujian (antd Button)

- `components/exam/NavigasiSoal.tsx:32` (`<button>` mentah) → antd `Button`.
- Layout ANBK TETAP: bilah peserta + timer sticky, grid nomor soal, bar bawah
  (Sebelumnya / Ragu-ragu / Berikutnya / Selesai). Hanya komponen kotaknya yang berubah.
- Pemetaan status → Button (warna via `type`/token, BUKAN hex mentah):

| Status kotak | antd Button |
|---|---|
| Belum dijawab (tidak aktif) | `type="default"` |
| Sudah dijawab (tidak aktif) | `style={{ backgroundColor: token.colorSuccess, borderColor: token.colorSuccess, color: token.colorWhite }}` (token dari `theme.useToken()`) |
| Ragu-ragu (tidak aktif) | `style={{ backgroundColor: token.colorWarningBg ?? "#fef3c7", borderColor: token.colorWarning, color: token.colorWarning }}` — pakai token antd (`colorWarningBg` bila tersedia di v6; bila tidak, `token.colorWarning` + class latar antd). DILARANG menulis hex di komponen — ambil dari token |
| Aktif (soal yang dibuka) | `type="primary"` + `aria-current="true"` (primer mengalahkan warna status; status aktif tetap terbaca dari legenda + `aria-label`) |
| Ragu-ragu + aktif | `type="primary"` + ikon `Flag` (bendera ragu tetap tampil) |

- Ukuran: `minWidth 44, minHeight 44` (target sentuh), `grid-cols-4 gap-1.5 sm:grid-cols-5 sm:gap-2`
  (dipertahankan), legenda di bawah grid dipertahankan (Dijawab / Ragu-ragu / Belum).
- `aria-label="Soal N (ragu-ragu) (dijawab)"` dipertahankan.

### 2.10 Migrasi `<select>` Menjodohkan → antd `Select` + uji grading (WAJIB)

**Pengecualian `<select>` native jawaban Menjodohkan DICABUT.** Standar baru:

- `components/exam/SoalRenderer.tsx:206-239` (`<select>` native + komentar "PENGECUALIAN
  DISETUJUI USER") → antd `Select`.
- Pemetaan nilai (TIDAK BOLEH berubah — jalur penilaian):
  - `value`: `peta[p.kiri] ?? ""` → antd `Select value={peta[p.kiri] || undefined}`
    dengan `placeholder="— Pilih pasangan —"`. Kosong = `undefined` di Select,
    tapi `onChange` HARUS memetakan `undefined` → `""` agar format value yang dikirim
    ke backend TETAP `{ kiri: "" }` untuk jawaban kosong (bukan `undefined`, bukan key hilang).
  - `options`: `kananOptions.map(k => ({ value: k, label: stripHtmlToText(k) }))`
    (label teks polos — antd `Select` tidak menampung HTML di option, sama seperti
    `<option>` sebelumnya).
  - `onChange`: `onChange({ ...peta, [p.kiri]: v ?? "" })` — identik dengan sebelumnya.
  - `disabled`, `aria-label` dipertahankan. `showSearch` BOLEH (membantu pasangan banyak);
    bila `showSearch` mengubah perilaku keyboard, matikan dan laporkan ke koordinator.
  - Komentar pengecualian di `SoalRenderer.tsx:28,77,206` DIHAPUS dan diganti komentar
    standar baru ("Migrasi antd Select — format value TIDAK berubah").
- **Uji regresi grading Menjodohkan end-to-end (WAJIB, oleh Tim QA — bukan tim eksekusi):**
  1. Buat/ambil soal Menjodohkan dengan N pasangan (N ≥ 3) yang sudah punya nilai acuan
     (nilai SEBELUM migrasi, dari backend yang sama).
  2. Kasus: (a) semua pasangan benar, (b) semua kosong (tidak memilih apa pun),
     (c) sebagian benar + sebagian kosong, (d) semua salah (pasangan tertukar),
     (e) ubah jawaban lalu ubah lagi sebelum submit.
  3. Jawab → submit → nilai HARUS **IDENTIK** dengan sebelum migrasi untuk semua kasus,
     termasuk jawaban kosong/parsial.
  4. **Bila uji grading gagal:** Tim QA melapor ke koordinator — JANGAN diam-diam
     mengembalikan native `<select>`. Koordinator memutuskan (perbaiki migrasi / tunda rilis).
- Aturan yang sama berlaku untuk `peta_konsep` (cabang yang sama di file itu).

### 2.11 Label login (Indonesia)

| Lokasi | Sebelum | Sesudah |
|---|---|---|
| `pages/LoginPage.tsx:80,90` | `Email / Username` (label + placeholder) | `Email / Nama pengguna` |
| `pages/LoginPage.tsx:95,106` | `Password` (label + placeholder) | `Kata sandi` |
| `pages/LoginPage.tsx` (teks bantu) | `Pakai email atau username + password. Siswa login dengan username (NIS/NISN).` | `Pakai email atau nama pengguna + kata sandi. Siswa login dengan nama pengguna (NIS/NISN).` |
| `pages/GantiPasswordPage.tsx` | `Ganti Password`, `Email / Username (kosongkan bila sudah login)`, `Password Lama`, `Password Baru`, `Konfirmasi Password Baru`, tombol `Ganti Password`, pesan `Password baru minimal 8 karakter.` / `Konfirmasi password baru tidak sama.` / `Password baru harus berbeda dari password lama.` / `Password awal harus diganti sebelum login.` / `Password berhasil diganti. Silakan login dengan password baru.` | `Ganti Kata Sandi`, `Email / Nama pengguna (kosongkan bila sudah login)`, `Kata Sandi Lama`, `Kata Sandi Baru`, `Konfirmasi Kata Sandi Baru`, tombol `Ganti Kata Sandi`, pesan `Kata sandi baru minimal 8 karakter.` / `Konfirmasi kata sandi baru tidak sama.` / `Kata sandi baru harus berbeda dari kata sandi lama.` / `Kata sandi awal harus diganti sebelum login.` / `Kata sandi berhasil diganti. Silakan login dengan kata sandi baru.` |
| `pages/siswa/KartuSayaPage.tsx:232,235` | `Username:`, `Password awal:` | `Nama pengguna:`, `Kata sandi awal:` |
| `pages/siswa/UjianPage.tsx:355` | `Kembali ke Dashboard` | `Kembali ke Beranda` |

**Nama field API / state / `autoComplete` TIDAK berubah** (`password`, `password_lama`,
`password_baru`, `login`, `autoComplete="current-password"` dst.) — hanya teks tampil.

### 2.12 Tombol (1 pola — ringkasan operasional)

- Komponen: antd `Button` — SELALU. DILARANG `<button>` mentah, `ui/Button`, tautan
  bergaya tombol (`<Link className="inline-flex … rounded-lg …">` → antd `Button href=`).
- Varian: `type="primary"` = aksi utama (MAKSIMAL 1 per tampilan/modal);
  `type="default"` = aksi sekunder; `type="text"` = tombol ikon / aksi ringan di atas
  latar berwarna; `danger` = hapus/berbahaya (di modal konfirmasi: `type="primary" danger`;
  di baris tabel: `type="default" danger` — SATU aturan ini menggantikan campur aduk §1.1 T8).
  `type="link"` / `dashed` DILARANG kecuali disetujui koordinator.
- Ukuran: `size="large"` untuk CTA utama + tombol submit form; default (middle) untuk
  semua tombol lain; **`size="small"` DILARANG untuk Button** (hapus 10 titik §1.1 T3).
  Pengecualian satu-satunya: toolbar `RichEditor` (kepadatan editor) boleh `size="small"`
  + `type="text"` — dicatat sebagai pengecualian eksplisit, bukan pola umum.
- State: mutasi → `loading={mutation.isPending}` (spinner), BUKAN `disabled` saja.
  `disabled` hanya untuk yang benar-benar tidak tersedia (bank terkunci, `indexAktif === 0`,
  form belum valid karena aturan yang sudah ada). Teks loading Bahasa Indonesia:
  `Menyimpan…`, `Menghapus…`, `Memproses…` (dipertahankan dari yang sudah ada).
- Tombol ikon: `icon={<…/>}` (antd icons untuk kode baru) + `aria-label`. Ikon di dalam
  children (lucide) dimigrasi ke `icon` saat file disentuh.

### 2.13 Form (pola antd)

- Form memakai antd `Form` + `Form.Item` (`label`, `validateStatus`/`help` untuk error).
  Menggantikan 25 `<form>` native + 36 `ds-label` + label `text-sm font-medium` manual.
- Validasi: pertahankan perilaku yang sudah ada (pesan + kondisi yang sama). DILARANG
  menambah aturan validasi baru yang memblokir submit (mis. `rules` baru) tanpa
  persetujuan koordinator — migrasi ini presentasi, bukan pengetatan validasi.
- Input/Select di form: `size="large"` (sudah konsisten di ~100 pemakaian — dipertahankan).
- `ds-form`, `ds-label`, `ds-help`, `ds-alert`, `ds-actions`, `ds-page-title`,
  `ds-section-title`, `ds-header-band`, `ds-table-wrap` DIHAPUS dari `index.css`
  setelah seluruh pemakainya dimigrasi (bukti grep 0).

### 2.14 Halaman khusus (pola yang dipertahankan + yang diubah)

- **Login:** antd `Card` + pita header biru dipertahankan bentuknya (brand `CBT SMK` +
  ikon), tapi dibangun dengan antd `Card` + `Typography` (bukan `ds-header-band` +
  `<p>` manual). Label Indonesia (§2.11).
- **GantiPassword / Forbidden / NotFound / ErrorBoundary:** antd `Card`/`Result`/
  `Button`/`Alert` (§1.6, §2.3). `ErrorBoundary` → antd `Result status="error"` +
  `Button` Muat Ulang + Kembali ke Beranda.
- **Halaman siswa (Ujian/Hasil/SiswaHome/KartuSaya):** layout ANBK dipertahankan
  (bilah peserta + timer, grid nomor, bar bawah). `HasilPage` Tag → peta §2.2.
- **Dashboard:** antd `Card` + `Statistic` (sudah benar) MINUS `border-t-4` (hapus);
  menu cepat → antd `Button type="link"`? TIDAK — menu cepat adalah navigasi, bukan
  tombol: pakai antd `Space` + `Typography.Link` (atau `Button type="link"` bila memang
  perlu tampilan tombol). Diputuskan: `Typography.Link` (link navigasi, bukan tombol).
- **SoalEditor (guru):** radio/checkbox kunci (`<input type="radio|checkbox">` di
  `SoalEditor.tsx:193,210,431`) → antd `Radio`/`Checkbox` (sisi guru, bukan jalur
  penilaian siswa; risiko rendah). Tombol → antd `Button` (sudah). Modal ganti tipe →
  antd `Modal` (sudah, `:904`).
- **RichEditor:** toolbar `<button>` (8×) → antd `Button type="text" size="small"`;
  panel rumus/video/gambar: tombol → antd `Button`, input → antd `Input`, radio mode
  rumus → antd `Radio.Group`. **Area edit TipTap + render KaTeX TETAP** (bukan komponen UI).
- **SoalRenderer (guru, read-only):** kartu opsi rakitan (`rounded-lg border …`) →
  antd `Card`/`List` + `Tag`; `Tag color="purple"` dipertahankan sebagai warna preset
  antd untuk kelompok soal (bukan hex baru).
- **Analisis (AnalisisButir/Kompetensi):** `__badgeColor` dihapus (§2.2); Heatmap +
  PengecohTable → antd `Table` + `Tag`; `DistribusiChart` SVG dipertahankan (sudah token).
- **Import (Excel/Word):** alur + endpoint + preview TIDAK berubah (jangan dirusak, §3);
  hanya `Link` rakitan → antd `Button` (§1.1 T6) dan wrapper kartu → antd `Card`.
- **Backup / UKK / Event / Ruangan:** konfirmasi → `Modal.confirm` (§2.6); Button
  `size="small"` dihapus (§2.12); `size="large"` pada tombol hapus di modal
  (`BackupPage.tsx:239,260`, `EventPage.tsx:230`, `RuanganPage.tsx:200`) → default
  (middle) agar seragam dengan modal konfirmasi lain.

---

## 3. DAFTAR "JANGAN DIRUSAK"

> Semua item di bawah ini adalah perilaku/kontrak yang SUDAH terverifikasi. Tim eksekusi
> DILARANG mengubahnya dengan alasan presentasi. Bila sebuah perubahan presentasi
> tampak menyentuh item ini, hentikan dan tanya koordinator.

1. **Endpoint + kartu angka Dashboard** `GET /api/v1/dashboard/ringkasan` (bank_soal,
   jadwal_ujian, peserta, rekap) + 4 kartu `Card` + `Statistic` + `loading`
   (`pages/admin/DashboardPage.tsx:37-40`). Hanya `border-t-4` yang dihapus; angka,
   endpoint, dan `loading` TETAP.
2. **Import Excel/Word + preview** (`ImportSoalPanel.tsx`): endpoint
   `import/template.xlsx`, `import/template.docx`, `import/preview`, `import/xlsx`;
   mode `all_or_nothing`/`skip_errors`, duplikat lewati/simpan_ganda, validasi
   magic bytes / 5 MB / 500 baris, katalog E1–E32 + W1–W7. Hanya komponen presentasi
   yang boleh berubah.
3. **Fix pasangan Menjodohkan** (kunci `[]` default, kunci disaring dari pasangan kosong,
   edit/hapus pasangan menyinkronkan kunci) di `SoalEditor` — JANGAN sentuh logika kunci.
   Migrasi Select §2.10 hanya menyentuh `SoalRenderer` siswa, bukan `SoalEditor`.
4. **Expiry token (backend)** — `token_expiry_detik` (default 900) di Settings adalah
   nilai backend; form Settings hanya menampilkan/mengirimnya. Jangan ubah nilai,
   validasi, atau nama field.
5. **D3** — Rekap menghitung jawaban HANYA bila `jawaban IS NOT NULL`
   (`COUNT … AND j.jawaban IS NOT NULL` di `rekap.go`). Jangan ubah query/rumus.
6. **D8** — middleware rate limiter melewati (tanpa hitung) semua path non-`/api/`
   dan non-`/ws/`. Jangan ubah middleware.
7. **Sidebar scroll sendiri** — `AdminLayout.tsx:242` (`flex-1 overflow-y-auto`) +
   `Sider sticky` + Drawer HP. Jangan hapus pembungkus scroll; 19 menu × 44px
   melebihi tinggi layar laptop.
8. **Helper format tanggal Indonesia** `lib/format.ts` (`formatTanggal` →
   `16 Okt 2026`, `formatTanggalJam` → `16 Okt 2026, 22:31 WIB`, `formatJam`) —
   fungsi, bukan komponen; JANGAN ubah format/zonasi/nama bulan. Semua tanggal tampil
   WAJIB lewat helper ini (jangan format manual dengan `toLocaleDateString` baru).
9. **RichEditor TipTap + KaTeX** — area editor, rumus `$…$`, embed video, sanitasi
   DOMPurify, render `RichContent`. Hanya chrome toolbar/panel yang boleh ke antd.
10. **Pemisahan daftar vs form Jadwal/Paket** — `JadwalPage`/`PaketPage` (daftar) vs
    `JadwalFormPage`/`PaketFormPage` (form) vs `PaketDetailPage` (detail). Jangan
    gabungkan kembali; Token via Dialog di `JadwalPage` dipertahankan.
11. **ConfigProvider antd (`#1e4e9e`)** di `App.tsx` — primer, radius, font, locale
    `id_ID` dipertahankan; Tim Token hanya MELENGKAPI token (§2.1), bukan mengganti.
12. **Kontrak lain yang sudah ada:** route `/adm/*`, `/ujian/:attemptId`, `/hasil/:attemptId`;
    auth + `PASSWORD_CHANGE_REQUIRED` + wajib ganti password; timer + auto-submit +
    autosave; anti-curang (pindah tab 1× peringatan, 2× auto-block) + unblock lanjut
    sisa waktu; pause bekukan timer; token kedua = token pertama hangus; toleransi
    telat 15 menit; tryout tanpa token + pembahasan langsung; nilai esai disembunyikan
    sampai final; grading (PG Kompleks proporsional, Menjodohkan proporsional per
    pasangan, nilai = Σ poin / Σ bobot × 100); KKTP per mapel/TP; remedial nilai
    tertinggi; UKK teori 40% + praktik 60%, rubrik Proses 30% / Produk 50% / Sikap 20%,
    lulus ≥70; teks `Ragu-ragu`, `Sebelumnya`, `Berikutnya`, `Selesai` (jangan diubah);
    nama aplikasi `CBT SMK` + disclaimer bukan aplikasi ujian nasional resmi.
13. **Pertanyaan terbuka yang BUKAN wewenang tim eksekusi:** toleransi telat (tetap
    15 menit vs boleh masuk selama jadwal berjalan) — menunggu keputusan Ujang.

---

## 4. INSTRUKSI KERJA 4 TIM EKSEKUSI

> **Prinsip anti-tabrakan:** setiap file hanya disentuh SATU tim dalam SATU fase.
> Urutan fase: **Fase 0 Token → Fase 1 Komponen → Fase 2 Layout → Fase 3 Responsif**.
> Fase berikutnya mulai setelah fase sebelumnya `tsc` + `vite build` lolos dan
> `grep` bukti dicatat di bawah. Tim QA (di luar 4 tim ini) melakukan uji regresi
> grading Menjodohkan (§2.10) + regresi visual per fase.

### Fase 0 — Tim Token (fondasi; tanpa menyentuh halaman)

| # | Tugas | File (HANYA ini) |
|---|---|---|
| 0.1 | Lengkapi token ConfigProvider (§2.1): tambah `colorSuccess`, `colorWarning`, `colorError`, `colorInfo`, `colorText`, `colorTextSecondary`, `colorBorder`, `colorBgLayout`, `fontSize: 14`, `controlHeightLG: 44`. JANGAN ubah `colorPrimary/borderRadius/fontFamily/locale` | `frontend/src/App.tsx` (blok `theme` saja, baris ±79-90) |
| 0.2 | Hapus token yatim/ganda dari `@theme`: `--color-primary-dark #163d7d` (hover primer = `#173d7a` satu nilai), `--color-accent #2f6fd0`. Tambah komentar "token Tailwind = jembatan migrasi, dihapus di akhir" | `frontend/src/index.css` (`@theme`, baris ±89-132) |
| 0.3 | Tulis peta token → komponen (tabel §2.1–§2.2) ke panduan agar 3 tim lain memakai nama yang sama | `frontend/src/ANTD-PANDUAN.md` (seksi token saja) |
| 0.4 | Bukti selesai: `grep -rn "#163d7d\|#2f6fd0" frontend/src` = 0 (di luar komentar larangan yang ikut dihapus) | — |

### Fase 1 — Tim Komponen (komponen + pemakaian komponen di halaman)

**File komponen (milik penuh Tim Komponen):** seluruh `frontend/src/components/**`
KECUALI `components/AdminLayout.tsx` dan `components/Layout.tsx` (milik Tim Layout, Fase 2).

| # | Tugas | File |
|---|---|---|
| 1.1 | Hapus 8 file 0-importer (§1.6) dengan bukti grep + tsc | `components/ui/Table.tsx`, `components/common/ConfirmDialog.tsx`, `components/common/Pagination.tsx`, `components/common/SearchInput.tsx`, `components/common/StatCard.tsx`, `components/common/PrintLayout.tsx`, `components/layout/Sidebar.tsx`, `components/layout/Topbar.tsx` |
| 1.2 | Migrasi + hapus komponen custom yang masih punya pemakai (§1.6): `ui/Button`, `ui/Card`, `ui/Input`, `ui/Select`+`ui/Dropdown`, `ui/Dialog`, `ui/Badge`, `common/PageHeader`, `common/StatusBadge`, `common/EmptyState`, `common/StatCard`, `common/Breadcrumbs` — ubah pemakainya dulu, lalu hapus file-nya | File komponen di atas + seluruh pemakainya (26 file PageHeader/Breadcrumbs, 4 file ui/Button, dst. — lihat tabel §1.6) |
| 1.3 | `NavigasiSoal` → antd `Button` (§2.9) | `components/exam/NavigasiSoal.tsx` |
| 1.4 | `SoalRenderer` (siswa) `<select>` → antd `Select` (§2.10); hapus komentar pengecualian; serahkan ke Tim QA untuk uji grading | `components/exam/SoalRenderer.tsx` |
| 1.5 | `SoalEditor`: radio/checkbox kunci → antd `Radio`/`Checkbox`; `SoalRenderer` (guru): kartu opsi rakitan → antd `Card`/`Tag` | `components/soal/SoalEditor.tsx`, `components/soal/SoalRenderer.tsx` |
| 1.6 | `RichEditor`: toolbar/panel `<button>`/`<input>`/radio → antd `Button`/`Input`/`Radio.Group`; area TipTap + KaTeX TETAP | `components/soal/RichEditor.tsx` |
| 1.7 | Analisis → antd `Table`/`Tag`: `KompetensiHeatmap`, `PengecohTable`, `DayaBedaBadge` (§2.2); `DistribusiChart` TETAP (sudah token) | `components/analisis/*.tsx` |
| 1.8 | Hapus `__badgeColor` di 4 file + ganti `Tag` ke peta §2.2; perbaiki `SiswaHomePage` Tag biru ganda; hapus span badge inline `BankSoalListPage` | `pages/admin/master/SiswaPage.tsx:39`, `pages/admin/ukk/UkkListPage.tsx:18`, `pages/admin/ukk/UkkDetailPage.tsx:43`, `pages/admin/analisis/AnalisisButirPage.tsx:84`, `pages/siswa/SiswaHomePage.tsx:141,142`, `pages/admin/banksoal/BankSoalListPage.tsx:95` |
| 1.9 | `ImportSoalPanel`: 2 `<Link>` rakitan → antd `Button href=`; wrapper kartu → antd `Card` | `pages/admin/banksoal/ImportSoalPanel.tsx:696,993,607,649,708,763` |
| 1.10 | Konfirmasi hapus → `Modal.confirm` di seluruh halaman yang memakai Modal `footer={null}` manual; hapus Modal duplikat `PaketPage`; 2 dialog `fixed` BankSoalDetail → antd `Modal` | `pages/admin/paket/PaketPage.tsx:256,324`, `pages/admin/banksoal/BankSoalDetailPage.tsx:238,260,273,280,287`, + seluruh Modal konfirmasi manual (§1.1 T11) |
| 1.11 | Pencarian → antd `Input.Search` di seluruh halaman filter; `Pagination` manual Backup → antd `Pagination` | `pages/admin/master/*.tsx`, `pages/admin/paket/PaketPage.tsx:285`, `pages/admin/jadwal/JadwalPage.tsx:244`, `pages/admin/backup/BackupPage.tsx:215,218` |
| 1.12 | `Timer`: `fontSize 22` inline → token (`fontFamilyCode`); `BannerPeringatan`: Button `size="small"` → default | `components/exam/Timer.tsx:57`, `components/exam/BannerPeringatan.tsx:32` |
| 1.13 | Bukti selesai: `grep -rn "components/ui/Button\|components/ui/Card\|components/ui/Input\|components/ui/Select\|components/ui/Table\|components/ui/Dialog\|components/ui/Badge\|components/ui/Dropdown\|components/common/PageHeader\|components/common/StatusBadge\|components/common/EmptyState\|components/common/ConfirmDialog\|components/common/Pagination\|components/common/SearchInput\|components/common/StatCard\|components/common/PrintLayout\|components/layout/Sidebar\|components/layout/Topbar\|components/layout/Breadcrumbs" frontend/src --include="*.tsx" | grep` = 0 (kecuali file komponen itu sendiri sebelum dihapus) + `tsc` lolos |

### Fase 2 — Tim Layout (kerangka + halaman auth/error + sapuan tipografi/spacing)

| # | Tugas | File (HANYA ini — jangan sentuh file Fase 1 yang masih dikerjakan) |
|---|---|---|
| 2.1 | `AdminLayout` + `Layout`: ganti 6 hex `#ffffff` → token/antd (`color` putih via `theme.useToken()` / hapus `style`); samakan padding konten; pertahankan `flex-1 overflow-y-auto` sidebar + Drawer | `frontend/src/components/AdminLayout.tsx`, `frontend/src/components/Layout.tsx` |
| 2.2 | Halaman auth/error → antd penuh (§2.11, §2.14): Login (label Indonesia + `Typography`), GantiPassword (antd `Card`/`Input`/`Button`/`Alert`, label Indonesia), NotFound/Forbidden/ErrorBoundary (antd `Result` + `Button`) | `frontend/src/pages/LoginPage.tsx`, `pages/GantiPasswordPage.tsx`, `pages/NotFoundPage.tsx`, `pages/ForbiddenPage.tsx`, `pages/ErrorBoundary.tsx` |
| 2.3 | Label Indonesia lanjutan: `KartuSayaPage` (Username/Password awal), `UjianPage` (Kembali ke Beranda) | `pages/siswa/KartuSayaPage.tsx:232,235`, `pages/siswa/UjianPage.tsx:355` |
| 2.4 | Sapuan tipografi/spacing/radius/shadow (§2.3–§2.5) di halaman yang TIDAK disentuh Fase 1: judul halaman → `Typography.Title level={3}`; `text-[15/16/17px]` → skala (§2.3, kecuali teks soal 16px); `rounded-xl/2xl/md` → `rounded-lg` (8); shadow custom → hapus; `p-*/space-y-*` liar → skala §2.4; `ds-*` terakhir → antd; `Dashboard` `border-t-4` → hapus; menu cepat → `Typography.Link` | Seluruh `pages/**` + `components/**` yang tersisa (koordinasi daftar file dengan Tim Komponen agar tidak tabrakan — file yang sudah diubah Fase 1 dicatat di log fase) |
| 2.5 | Form → antd `Form` + `Form.Item` (§2.13) di halaman form yang belum dimigrasi Fase 1 (JadwalForm, PaketForm, master modal, UKK, Event, Ruangan, Settings, BankSoalForm) — pertahankan validasi yang ada | `pages/admin/jadwal/JadwalFormPage.tsx`, `pages/admin/paket/PaketFormPage.tsx`, `pages/admin/master/*.tsx` (modal), `pages/admin/ukk/*.tsx`, `pages/admin/event/*.tsx`, `pages/admin/settings/SettingsPage.tsx`, `pages/admin/banksoal/BankSoalFormPage.tsx` |
| 2.6 | Hapus `.ds-*` dari `index.css` setelah grep 0 pemakai; hapus token Tailwind `@theme` setelah grep 0 `bg-anbk|text-anbk|…` (kecuali `.rich-*` yang dipertahankan + dipetakan ke token) | `frontend/src/index.css` |
| 2.7 | Bukti selesai: `grep -rn "ds-\|#ffffff\|rounded-xl\|rounded-2xl\|shadow-xl\|shadow-2xl\|<button\|<select\|<table" frontend/src --include="*.tsx"` = 0 (kecuali dokumen cetak KartuSaya + komentar yang menjelaskan) + `tsc` lolos |

### Fase 3 — Tim Responsif (verifikasi breakpoint + perbaikan overflow)

| # | Tugas | File / titik |
|---|---|---|
| 3.1 | Uji 360 / 768 / 1280 px untuk SEMUA halaman (daftar §1.7 P1–P13). Perbaiki yang terbukti meluber; JANGAN ubah yang hanya dicurigai tapi terbukti baik | Daftar §1.7 |
| 3.2 | Tabel: hapus wrapper `div overflow-x-auto` ganda (sisakan `Table scroll`); `BankSoalListPage` pola ganda kartu+tabel → Table scroll saja (§2.7) | ±20 halaman wrapper tabel + `pages/admin/banksoal/BankSoalListPage.tsx:86` |
| 3.3 | Grid: setiap `grid-cols-2` tanpa breakpoint → `grid-cols-1 sm:grid-cols-2`; verifikasi `grid-cols-4` NavigasiSoal di 320–360px | Hitungan `grid-cols-2` 34× + `components/exam/NavigasiSoal.tsx:24` |
| 3.4 | Modal: verifikasi `width={520/720}` di 360px (form di dalam tidak meluber); bar bawah ujian (§1.7 P1) + bilah peserta (§1.7 P2) di 360px | `pages/siswa/UjianPage.tsx:499-532,631-662`, modal 520/720 (§1.5 R6) |
| 3.5 | Import preview di 360px (§1.7 P4) + kartu cetak 360px + print (§1.7 P8) | `pages/admin/banksoal/ImportSoalPanel.tsx:833-843`, `pages/siswa/KartuSayaPage.tsx:217-218,246-256` |
| 3.6 | Bukti selesai: daftar halaman × breakpoint dengan status (OK / diperbaiki + file:baris) + `tsc` + `vite build` lolos. TIDAK ada rebuild `.exe`/zip dalam tugas ini | — |

### Verifikasi akhir (koordinator + Tim QA, setelah Fase 3)

1. `grep` bukti: 0 komponen custom duplikat (§1.6), 0 `<button>/<select>/<table>` mentah
   (kecuali dokumen cetak), 0 hex mentah di luar token/print, 0 `ds-*`, 0 `size="small"`
   pada Button (kecuali toolbar RichEditor), 0 `Modal.confirm` yang terlewat (semua
   konfirmasi hapus memakai `Modal.confirm`).
2. `tsc` lolos + `vite build` lolos + `go test` backend tidak tersentuh (tidak diubah).
3. Uji regresi grading Menjodohkan (§2.10) IDENTIK sebelum/sesudah — syarat rilis.
4. Spot-check manual 1× (login → form → modal → ujian → import file) — menutup celah
   "tidak ada uji klik browser sungguhan" dari redesign sebelumnya.
5. TIDAK ada rebuild `.exe` / zip dalam rangkaian ini (di luar wewenang tim eksekusi
   ini; koordinator yang memutuskan rebuild).

---

## 5. Lampiran — angka hitung (grep, 17 Okt 2026)

| Metrik | Nilai |
|---|---|
| File `*.tsx` di `frontend/src` | 77 |
| File mengimpor `from "antd"` | 56 |
| `<Button>` antd | 206 (primary 69, default 70, text 4, `danger` 30, `size="large"` 9 Button, `size="small"` 10 Button, `disabled` 51, `loading` 9, `icon` 5) |
| `<button>` mentah | 16 |
| `<select>` native (elemen) | 1 (`SoalRenderer` siswa) + 3 baris komentar menyebut `<select>` |
| `<input type="radio|checkbox">` native | 4 (SoalEditor 3, RichEditor 1) |
| `<table>` mentah | 3 (KartuSaya, KompetensiHeatmap, PengecohTable) |
| antd `<Table>` / `<Modal>` / `<Tag>` / `<Empty>` / `<Pagination>` / `<Breadcrumb>` / `<Card>`+`<Statistic>` / `<Select>` / `<Steps>` / `<Descriptions>` | 29 / 27 / 30 / 5 / 5 / 28 / 46 / 46 / 2 / 2 |
| `Modal.confirm` / `Input.Search` / antd `<Form>` / `Form.Item` / `Typography` | 0 / 0 / 0 / 0 / 0 |
| `<form>` native | 25 |
| `ds-label` / total class `.ds-*` | 36 / 45 |
| `style={{…}}` inline | 71 |
| `text-slate-*` / `bg-slate-*` / `border-slate-*` | 386 / 51 / 117 |
| `rounded-lg` / `rounded-xl` / `rounded-full` / `rounded` / `rounded-md` / `rounded-2xl` / `rounded-t` / `rounded-sm` | 109 / 49 / 21 / 11 / 8 / 8 / 6 / 3 |
| `shadow-sm` / `shadow-xl` / `shadow-none` / `shadow-md` / `shadow-lg` / `shadow-2xl` / `shadow` | 63 / 7 / 3 / 1 / 1 / 1 / 1 |
| `gap-1/2/3/4/6/8/10` | 51 / 147 / 59 / 38 / 7 / 1 / 2 |
| `font-bold` / `font-semibold` / `font-medium` / `font-normal` | 69 / 101 / 124 / 2 |
| File impor `lucide-react` / `@ant-design/icons` | 43 / 8 |
| File komponen custom duplikat antd | 19 (§1.6) |
| Pemakaian `<PageHeader>` / `<Breadcrumbs>` | 27 / 27 (26 file) |
| Pemakaian `<StatusBadge>` / `<EmptyState>` / `<StatCard>` / `<ConfirmDialog>` / `<Pagination>` custom / `<SearchInput>` / `<PrintLayout>` | 2 / 2 / 0 / 0 / 0 / 0 / 0 |

---

*Dokumen ini adalah standar yang disepakati hasil audit. Tim eksekusi bekerja dari dokumen
ini; penyimpangan dari standar harus disetujui koordinator dan dicatat di sini.*

## 6. Catatan model kartu 1-KARTU (17 Okt 2026, Tim QA + Dokumen)

Model kartu berubah SETELAH audit di atas: **1 kartu per siswa untuk semua
ujian** (identitas sekali + tabel semua jadwal kelasnya; PDF dibuat backend
gofpdf; `window.print()` DIHAPUS dari alur kartu; `/kartu/:jadwalId`
redirect ke `/kartu`). Akibatnya temuan audit yang merujuk tata letak cetak
lama `KartuSayaPage` (mis. §1.7 P8 — grid tanda tangan / garis / tabel cetak,
dan nomor-nomor baris `KartuSayaPage.tsx` di §1–§3) **tidak lagi berlaku**
untuk halaman kartu yang ditulis ulang — halaman kartu kini hanya pratinjau
ringkas + tombol unduh PDF backend. Standar desain di dokumen ini (1 palet,
antd, tipografi, spacing) tetap berlaku untuk halaman kartu yang baru.
Perubahan model kartu dicatat di `ANALISIS.md` §8.3 dan `docs/KONTRAK.md` §9.
