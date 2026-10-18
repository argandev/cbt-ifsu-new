# AUDIT UI — CBT SMK v2 (Tim Tambahan 3 — Frontend)

Tanggal audit: 16 Okt 2026. Cakupan: `frontend/src/App.tsx` vs halaman yang ada,
file baru Tim Tambahan 3, import/typecheck, responsif.
Aturan: temuan di file tim lain hanya DILAPORKAN, tidak diperbaiki.

## 0. File baru Tim Tambahan 3 (satu-satunya file yang ditulis tim ini)

Komponen bersama (`frontend/src/components/common/`):

| File | Isi |
|---|---|
| `StatCard.tsx` | Kartu statistik: `label`, `nilai`, `ikon?`, `subteks?`, `tone?` (default/primary/success/warning/danger/info) |
| `PageHeader.tsx` | Judul + deskripsi + aksi kanan; aksi full-width di HP |
| `EmptyState.tsx` | Ikon + judul + deskripsi + aksi (`aksi?` ReactNode atau `labelAksi` + `onAksi`) |
| `ConfirmDialog.tsx` | Dialog konfirmasi (pengganti `window.confirm`): judul, pesan, batal/konfirmasi, `loading`, `berbahaya`; Esc/overlay tidak menutup saat loading |
| `Pagination.tsx` | Sebelumnya/berikutnya + nomor halaman (dengan …) + info "Menampilkan X–Y dari Z"; halaman basis 1 |
| `SearchInput.tsx` | Input search + ikon + tombol hapus + debounce 300ms (`debounceMs` bisa diubah) |
| `StatusBadge.tsx` | Badge status warna konsisten: draft/aktif/pause/selesai/berjalan/diblokir/lulus/tidak-lulus + menunggu/disetujui/ditolak/terkunci/terbuka; status tak dikenal dirender sebagai badge default berisi teks mentahnya |
| `PrintLayout.tsx` | Kop sekolah + judul + subjudul + tanggal + isi + blok tanda tangan + tombol Cetak (`window.print()`) |

Komponen layout BARU yang mandiri (`frontend/src/components/layout/`):

| File | Isi |
|---|---|
| `Sidebar.tsx` | Menu `/adm/*` mandiri + `DEFAULT_SIDEBAR_ITEMS` (mencakup SEMUA route admin, termasuk event/ruangan/kartu/backup yang belum ada di menu `AdminLayout` lama). Filter permission via `useAuthStore`. TIDAK mengubah `AdminLayout.tsx` |
| `Topbar.tsx` | Tombol menu (HP) + judul + nama/role + tombol keluar. TIDAK mengubah `Layout.tsx`/`AdminLayout.tsx` |
| `Breadcrumbs.tsx` | `items: { label, to? }[]`; item terakhir = halaman aktif (`aria-current="page"`); scroll horizontal di HP |

Halaman (`frontend/src/pages/`):

| File | Isi |
|---|---|
| `NotFoundPage.tsx` | Halaman 404 + tombol kembali ke beranda |
| `ForbiddenPage.tsx` | Halaman 403 + tombol kembali ke beranda |
| `ErrorBoundary.tsx` | Error boundary class component: pesan ramah + tombol Muat Ulang + kembali ke beranda; menerima `fallback?` opsional |

Lainnya:

| File | Isi |
|---|---|
| `frontend/src/print.css` | Stylesheet khusus print: sembunyikan navigasi/tombol saat print, tabel full-width garis hitam, hemat tinta, `@page A4`. BARU — tidak mengubah `index.css` |

**TIDAK dibuat: `frontend/src/pages/siswa/JadwalSayaPage.tsx`.**
Alasan: `SiswaHomePage.tsx` SUDAH menampilkan daftar jadwal aktif siswa
(`GET /api/v1/ujian/aktif` via `apiClient` + TanStack Query). Membuat halaman
duplikat hanya akan memecah dua tampilan jadwal siswa. Lihat temuan §C.1:
bentuk data yang dipakai `SiswaHomePage` tidak cocok dengan respons backend
saat ini — itu yang perlu diperbaiki tim pemilik halaman, bukan halaman baru.

Semua file baru: default export, TypeScript ketat, Tailwind 4, lucide-react,
NOL emoji (scan emoji seluruh `frontend/src/**/*.tsx`: bersih), responsif
(tombol full-width HP, tabel di pemanggil dibungkus `overflow-x-auto`).

Verifikasi: `npx tsc --noEmit` dari `frontend/` → **EXIT 0, nol error**
(mencakup file baru + seluruh file lama).

## A. Route di App.tsx vs halaman yang ada

Semua route di `App.tsx` memiliki file halaman yang ada. Tabel:

| Route | Halaman | Status |
|---|---|---|
| `/login` | `pages/LoginPage.tsx` | Cocok |
| `/ganti-password` | `pages/GantiPasswordPage.tsx` | Cocok |
| `/` | `pages/siswa/SiswaHomePage.tsx` | Cocok (lihat temuan §C.1) |
| `/ujian/:attemptId` | `pages/siswa/UjianPage.tsx` | Cocok |
| `/hasil/:attemptId` | `pages/siswa/HasilPage.tsx` | Cocok |
| `/adm` (index) | `pages/admin/DashboardPage.tsx` | Cocok |
| `/adm/master/jurusan` | `pages/admin/master/JurusanPage.tsx` | Cocok |
| `/adm/master/kelas` | `pages/admin/master/KelasPage.tsx` | Cocok |
| `/adm/master/mapel` | `pages/admin/master/MapelPage.tsx` | Cocok |
| `/adm/master/guru` | `pages/admin/master/GuruPage.tsx` | Cocok |
| `/adm/master/siswa` | `pages/admin/master/SiswaPage.tsx` | Cocok |
| `/adm/master/siswa/import` | `pages/admin/master/ImportSiswaPage.tsx` | Cocok |
| `/adm/bank-soal` | `pages/admin/banksoal/BankSoalListPage.tsx` | Cocok |
| `/adm/bank-soal/baru` | `pages/admin/banksoal/BankSoalFormPage.tsx` | Cocok |
| `/adm/bank-soal/:id` | `pages/admin/banksoal/BankSoalDetailPage.tsx` | Cocok |
| `/adm/paket` | `pages/admin/paket/PaketPage.tsx` | Cocok |
| `/adm/jadwal` | `pages/admin/jadwal/JadwalPage.tsx` | Cocok |
| `/adm/monitor` | `pages/admin/monitor/MonitorPage.tsx` | Cocok |
| `/adm/rekap` | `pages/admin/rekap/RekapPage.tsx` | Cocok |
| `/adm/rekap/berita-acara` | `pages/admin/rekap/BeritaAcaraPage.tsx` | Cocok file, **TIDAK COCOK kontrak** (lihat §C.3) |
| `/adm/analisis` | `pages/admin/analisis/AnalisisButirPage.tsx` | Cocok file; tidak ada di KONTRAK §9 (tambahan Fase 2) |
| `/adm/kompetensi` | `pages/admin/analisis/KompetensiPage.tsx` | Cocok file; tidak ada di KONTRAK §9 (tambahan Fase 2) |
| `/adm/event` | `pages/admin/event/EventPage.tsx` | Cocok file; tidak ada di KONTRAK §9 (tambahan Fase 2) |
| `/adm/ruangan` | `pages/admin/event/RuanganPage.tsx` | Cocok file; tidak ada di KONTRAK §9 (tambahan Fase 2) |
| `/adm/kartu` | `pages/admin/kartu/KartuPesertaPage.tsx` | Cocok file; tidak ada di KONTRAK §9 (tambahan Fase 2) |
| `/adm/settings` | `pages/admin/settings/SettingsPage.tsx` | Cocok |
| `/adm/ukk` | `pages/admin/ukk/UkkListPage.tsx` | Cocok file; tidak ada di KONTRAK §9 (tambahan Fase 2) |
| `/adm/ukk/:id` | `pages/admin/ukk/UkkDetailPage.tsx` | Cocok file; tidak ada di KONTRAK §9 (tambahan Fase 2) |
| `/adm/ukk/:id/nilai` | `pages/admin/ukk/UkkNilaiPage.tsx` | Cocok file; tidak ada di KONTRAK §9 (tambahan Fase 2) |
| `/adm/ukk/:id/sertifikat/:siswaId` | `pages/admin/ukk/UkkSertifikatPage.tsx` | Cocok file; tidak ada di KONTRAK §9 (tambahan Fase 2) |
| `/adm/backup` | `pages/admin/backup/BackupPage.tsx` | Cocok file; tidak ada di KONTRAK §9 (tambahan Fase 2) |
| `/adm/*` | Komponen `NotFound` inline di `App.tsx` | Ada — sebaiknya diganti `NotFoundPage` (§B) |
| `/pengawas/monitor` | `MonitorPage` dipakai ulang (dalam `AdminLayout`) | Cocok |
| `*` | Komponen `NotFound` inline di `App.tsx` | Ada — sebaiknya diganti `NotFoundPage` (§B) |

Catatan tambahan:

- Menu di `AdminLayout.tsx` (`AdminSidebar`) **tidak menampilkan** route yang
  sudah ada: `/adm/event`, `/adm/ruangan`, `/adm/kartu`, `/adm/backup`.
  Route-nya bisa dibuka via URL langsung, tapi tidak ada jalan dari menu.
  `Sidebar.tsx` baru sudah mencakup semuanya; tim integrasi dapat memakai
  daftar itu saat menyelaraskan menu.
- KONTRAK §9 perlu diperbarui agar mencakup route Fase 2 (analisis,
  kompetensi, event, ruangan, kartu, ukk, backup) — dokumen tertinggal
  dari implementasi.

## B. Halaman baru + route yang perlu ditambahkan tim integrasi

Tim integrasi HANYA menyentuh `App.tsx` / `main.tsx` (file terlarang bagi
Tim Tambahan 3). Perubahan yang diminta:

```tsx
// App.tsx — ganti komponen NotFound inline dengan halaman baru:
import NotFoundPage from "./pages/NotFoundPage";
import ForbiddenPage from "./pages/ForbiddenPage";
// Hapus fungsi NotFound() inline, lalu:
<Route path="/403" element={<Layout><ForbiddenPage /></Layout>} />
// ... di dalam <Route path="/adm/*">:
<Route path="*" element={<NotFoundPage />} />
// ... rute catch-all paling bawah:
<Route path="*" element={<Layout><NotFoundPage /></Layout>} />
```

```tsx
// main.tsx — aktifkan stylesheet print + error boundary:
import "./print.css";
import ErrorBoundary from "./pages/ErrorBoundary";
// Bungkus <App />: <ErrorBoundary><App /></ErrorBoundary>
```

Opsional (tidak wajib, keputusan tim integrasi):

- `RequireAuth staffOnly` saat ini me-redirect diam-diam ke `/` bila role
  tidak diizinkan. Bila ingin pengguna melihat penjelasan, arahkan ke
  `/403` (`<Navigate to="/403" replace />`) alih-alih `/`.
- Refactor `AdminLayout` memakai `Sidebar.tsx` / `Topbar.tsx` baru, dan
  ganti 6 pemakaian `window.confirm` (§C.4) dengan `ConfirmDialog`.
- `JadwalSayaPage` TIDAK perlu ditambahkan (lihat §0).

## C. Import rusak / gagal typecheck (LAPORKAN SAJA)

**C.0 — Typecheck: `npx tsc --noEmit` → EXIT 0, nol error.** Tidak ada
import rusak dan tidak ada halaman yang gagal typecheck. Semua import
relatif di `frontend/src` mengarah ke file yang ada. Temuan di bawah ini
adalah ketidaksesuaian perilaku runtime / kontrak, BUKAN error typecheck —
sengaja tidak diperbaiki karena file-file tersebut milik tim lain.

**C.1 — `SiswaHomePage.tsx`: field respons tidak cocok dengan backend
(DAMPAK TINGGI, perilaku runtime).**
Backend `GET /api/v1/ujian/aktif` (`backend/internal/exam/exam.go`,
`ujianAktif`) mengembalikan `items[]` berisi: `jadwal_id`, `paket_id`,
`tanggal`, `jam_mulai`, `jam_selesai`, `status`, `paket_kode`,
`paket_nama`, `paket_jenis`, `durasi_menit`, `kelas_nama`, `attempt_id`,
`attempt_status`. Halaman memakai `item.id`, `item.nama_ujian`,
`item.mapel`, `item.mulai` — semuanya TIDAK ADA di respons. Akibatnya:

- Judul kartu jatuh ke fallback `Ujian ${item.jadwal_id}` (tidak fatal,
  tapi bukan nama paket).
- Link `Masuk Ujian` menjadi `/ujian/${item.id}` → **`/ujian/undefined`**
  (rusak).
- Tidak ada input token: alur mulai yang benar adalah
  `POST /api/v1/ujian/mulai { jadwal_id, token }` lalu navigasi ke
  `/ujian/:attemptId` dari `attempt_id` respons. Alur token siswa saat
  ini tidak ada di UI mana pun.
- `attempt_id` / `attempt_status` (`berjalan`/`diblokir`/`selesai`)
  diabaikan: siswa yang sudah punya attempt berjalan/diblokir/selesai
  tidak mendapat aksi Lanjut / Lihat Hasil yang benar.
- Backend hanya mengembalikan jadwal HARI INI (`tanggal = today`,
  status `draft`/`aktif`/`pause`) — bila itu disengaja, label halaman
  sebaiknya "Ujian Hari Ini", bukan daftar jadwal umum.

**C.2 — `AdminLayout.tsx`: permission `analisis.read` tidak ada di kontrak.**
Menu Analisis Butir digerakkan `perm: "analisis.read"`, tetapi daftar
kanonis KONTRAK §5 (59 permission) tidak memuat permission `analisis.*`
sama sekali. Akibat: `hasPermission("analisis.read")` selalu `false`
untuk non-admin → menu hilang untuk semua role non-admin, termasuk yang
seharusnya boleh melihat analisis. `Sidebar.tsx` baru memakai
`rekap.read` untuk Analisis/Kompetensi sebagai pengganti sementara —
keputusan permission yang benar harus dikonfirmasi pemilik RBAC (tambah
permission analisis ke kontrak + seed + `rbac.go`, atau tetapkan memakai
`rekap.read`).

**C.3 — Route berita acara tidak cocok kontrak.**
KONTRAK §9 menulis `/adm/rekap/:jadwalId/berita-acara`, tetapi `App.tsx`
memakai `/adm/rekap/berita-acara` (tanpa param) dan `BeritaAcaraPage`
meminta pengguna mengetik ID jadwal manual di input. Selain itu tidak ada
link dari `RekapPage` ke Berita Acara. Salah satu harus dipilih:
selaraskan route ke kontrak (param + link dari Rekap) atau ubah kontrak.

**C.4 — `window.confirm` masih dipakai di 6 tempat** (mestinya
`ConfirmDialog` baru): `BankSoalDetailPage.tsx` (3x: kunci bank, hapus
bank, hapus soal), `PaketPage.tsx` (1x: hapus paket), `JadwalPage.tsx`
(1x: hapus jadwal), `UkkListPage.tsx` (1x: hapus UKK).

**C.5 — `print.css` belum diimpor.** File dibuat, tetapi sampai tim
integrasi menambahkan `import "./print.css"` di `main.tsx` (§B), aturan
print kustom belum aktif. Halaman yang memakai kelas `print:hidden`
(`UkkSertifikatPage`, `KartuPesertaPage`, `BeritaAcaraPage`) saat ini
hanya mengandalkan varian `print:` bawaan Tailwind — yang tetap jalan —
tetapi aturan tambahan (tabel garis hitam, sembunyikan header/footer/nav)
belum berlaku.

**C.6 — Tidak ada route 403 yang terpasang.** `RequireAuth staffOnly`
me-redirect ke `/` tanpa penjelasan (lihat §B opsional).

## D. Cek responsif (LAPORKAN SAJA, tidak diperbaiki)

Metode: grep pola layout + inspeksi manual file berisiko. Tidak ditemukan
lebar piksel tetap (`w-[...px]`) di luar tabel. Ringkasan per area:

- **Tabel: AMAN kecuali satu.** Semua halaman bertabel membungkus `<table>`
  dalam `overflow-x-auto` (`min-w-[720px]`/`[960px]` di dalam wrapper —
  sesuai desain). Pengecualian: `KartuPesertaPage.tsx` — tabel info 2
  kolom tanpa wrapper `overflow-x-auto` (risiko RENDAH; nilai panjang
  hanya membuat kolom menyempit, bukan meluber halaman). `ui/Table.tsx`
  sengaja tidak membungkus sendiri (komentar di file meminta pemanggil
  membungkus); semua pemanggil terverifikasi membungkus, kecuali kasus
  kartu di atas.
- **Grid statistik 2 kolom di HP — PERHATIKAN, risiko RENDAH.**
  `DashboardPage` (`grid-cols-2 lg:grid-cols-4`), `RekapPage`
  (`grid-cols-2 lg:grid-cols-5`), `MonitorPage` (`grid-cols-2
  sm:grid-cols-4`), `AnalisisButirPage` (`grid-cols-2 lg:grid-cols-4`),
  `KompetensiPage` (`grid-cols-2 lg:grid-cols-3`), `UkkSertifikatPage`
  (`grid-cols-2` blok tanda tangan), `KartuPesertaPage` (`grid-cols-2`
  tanda tangan). Di layar 320–360px kartu menjadi sempit; tidak meluber,
  tetapi label/angka panjang bisa terpotong visual. Bila ingin aman,
  turunkan ke 1 kolom di bawah `sm` pada grid statistik.
- **`NavigasiSoal` — risiko RENDAH.** `grid-cols-5` tetap (5 tombol 44px
  + gap ≈ 240px): muat di 360px, mepet di 320px.
- **`UjianPage` — risiko RENDAH.** Header sticky memakai `flex-wrap`
  (aman); dialog submit memakai `flex gap-2` tanpa wrap — dua tombol
  berdampingan bisa terjepit di layar sangat sempit. Tombol
  Sebelumnya/Berikutnya memakai `justify-between` (aman).
- **`JadwalPage` / `PaketPage` — risiko RENDAH–SEDANG di 320px.**
  Baris aksi kartu memakai `flex gap-1`/`gap-2` tanpa `flex-wrap` —
  beberapa tombol kecil berdampingan bisa meluber bila label panjang.
  Pertimbangkan `flex-wrap` pada baris aksi kartu.
- **`SoalEditor` / `SoalRenderer` (exam) — AMAN.** Baris opsi memakai
  `flex-wrap` atau `flex-col` di HP; radiogroup Benar-Salah memakai
  `flex-1` (aman).
- **Halaman master (Jurusan/Kelas/Mapel/Guru/Siswa) — AMAN.** Header
  `flex-col sm:flex-row`, tabel `overflow-x-auto`, dialog bottom-sheet
  di HP (`items-end` + `sm:items-center`).
- **Layout global — AMAN.** `AdminLayout` drawer HP + sidebar desktop;
  `Layout` header + konten `max-w-[1100px]`.
- **Komponen baru Tim Tambahan 3 — AMAN.** Semua memakai pola
  `flex-col` di HP → `sm:flex-row`, tombol `w-full sm:w-auto`,
  `Breadcrumbs` dengan `overflow-x-auto`.
