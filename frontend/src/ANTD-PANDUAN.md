# Panduan Ant Design — CBT SMK v2 (Tim Setup + Tema)

> Fondasi redesign. Tim Layout / Admin / Siswa / Form memakai panduan ini.
> Aturan tertinggi: **hanya lapisan presentasi**. Jangan ubah route,
> endpoint API, logika bisnis, grading, auth, token, timer, anti-curang,
> validasi, nama field, atau payload. Bila ragu, jangan ubah — tanya.

## 0. Yang sudah terpasang (jangan diulang / jangan diubah sembarangan)

- `antd` **6.6.6** + `@ant-design/icons` **6.3.4** di `frontend/package.json`
  (install lokal via npm; React 19 — peer antd `>=18.0.0`, cocok).
  **Tanpa CDN apa pun.** antd v6 memakai CSS-in-JS: gayanya dihasilkan
  saat runtime dari bundle lokal. **Jangan** `import "antd/dist/antd.css"`,
  jangan `import "antd/dist/reset.css"`, jangan `<link>` font/CDN.
  Font tetap `@fontsource/plus-jakarta-sans` (lokal, sudah ada).
- `ConfigProvider` di `frontend/src/App.tsx` — membungkus `<Routes>`:
  ```tsx
  import { ConfigProvider } from "antd";
  import idID from "antd/locale/id_ID";

  <ConfigProvider
    locale={idID}
    theme={{
      token: {
        // Satu-satunya tempat token didefinisikan
        // (AUDIT-KONSISTENSI.md §2.1). Jangan buat palet baru.
        colorPrimary: "#1e4e9e",   // biru primer — TETAP
        colorPrimaryHover: "#173d7a", // hover primer SATU nilai
          // (nilai ganda lama #163d7d DIHAPUS)
        colorSuccess: "#15803d",
        colorWarning: "#b45309",
        colorError: "#dc2626",
        colorInfo: "#1e4e9e",      // info = primer
        colorText: "#0f172a",
        colorTextSecondary: "#475569",
        colorBorder: "#cbd5e1",
        colorBgLayout: "#eef2f7",
        borderRadius: 8,
        fontFamily: '"Plus Jakarta Sans", sans-serif',
        fontSize: 14,
        controlHeightLG: 44,       // target sentuh min 44px
      },
    }}
  >
    <Routes>…</Routes>
  </ConfigProvider>
  ```
  `locale={idID}` membuat teks bawaan antd berbahasa Indonesia
  ("Pilih", "Tidak ada data", tombol pagination, teks Upload, dll).
- **Token Tailwind `@theme` di `index.css` = jembatan migrasi** (bukan
  sumber kebenaran kedua). Setiap nilainya SAMA PERSIS dengan token antd
  di atas. Setelah `grep bg-anbk|text-anbk|…` = 0, blok `@theme` dihapus.
  Yang sudah dibersihkan: `--color-primary-dark` dipetakan ke `#173d7a`
  (satu nilai hover; nama variabel dipertahankan karena class
  `bg-primary-dark`/`text-primary-dark` masih dipakai komponen lain),
  `--color-accent #2f6fd0` dihapus (token yatim, 0 pemakai, di luar palet).
- **Pakai token, bukan hex.** Komponen antd membaca token lewat
  `type`/`color`/`status` (`type="primary"`, `Tag color="green"`,
  `Alert type="error"`). Bila warna tidak bisa diwakili `type`/`color`
  (mis. latar tombol grid soal "sudah dijawab"), ambil dari
  `theme.useToken()` (`token.colorSuccess`, …) — DILARANG menulis hex
  (`"#15803d"`, `"#ffffff"`, …) di komponen. Peta status → `Tag` ada di
  `docs/AUDIT-KONSISTENSI.md` §2.2 dan `docs/DESIGN-SYSTEM.md`.
- Halaman contoh: `frontend/src/pages/LoginPage.tsx` sudah memakai antd
  (`Card`, `Input`, `Input.Password`, `Button`, `Alert`). Logika loginnya
  (state, `handleSubmit`, navigasi, `ApiError` / `PASSWORD_CHANGE_REQUIRED`)
  **tidak diubah** — hanya komponen presentasinya yang diganti.
  Jadikan Login acuan gaya penulisan.
- `frontend/src/index.css`: **tidak ada** yang dihapus. Audit membuktikan
  tidak ada bentrok (lihat komentar "Ant Design — koeksistensi" di akhir
  file): file itu tidak mendefinisikan utility Tailwind, tidak punya
  selector `.ant-*`, dan class `.ds-*` / `.rich-*` masih dipakai halaman
  yang belum di-redesign. Token `--color-anbk` / `--color-primary`
  (`#1e4e9e`) dipertahankan.

## 1. Komponen antd dipakai untuk apa

| Kebutuhan lama | Komponen antd | Catatan |
|---|---|---|
| `ui/Button` | `Button` | `type="primary" \| "default" \| "dashed" \| "text" \| "link"`, `danger`, `loading`, `block`, `size="large"` (target sentuh min 44px — pakai `large` untuk aksi utama) |
| `ui/Card` / `StatCard` | `Card` | Judul via `title`, aksi via `extra`; statistik ringkas via `Statistic` di dalam `Card` |
| `ui/Table` / tabel HTML | `Table` | `columns`, `dataSource`, `rowKey`, `pagination`, `loading`, `locale={{ emptyText: ... }}`; header biru mengikuti `colorPrimary` via theme (jangan hardcode warna) |
| `ui/Input`, `ui/Select`, textarea | `Input`, `Input.Password`, `Input.TextArea`, `Select` | Lihat §2 untuk `Select` pengganti Dropdown |
| `ui/Dialog`, `ConfirmDialog` | `Modal` | `open`, `onOk`, `onCancel`, `okText="Ya"`, `cancelText="Batal"`, `okButtonProps={{ danger: true }}` untuk hapus |
| `ui/Badge`, `StatusBadge` | `Tag` | `color="green" \| "gold" \| "red" \| "blue" \| "default"`; status SELALU warna + teks, bukan warna saja |
| `PageHeader` + breadcrumb manual | `Breadcrumb` + judul halaman | `Breadcrumb` dengan `items={[{ title: "Beranda" }, ...]}`; judul pakai class `.ds-page-title` sampai halaman itu di-redesign penuh |
| Upload file (import, gambar) | `Upload` | `beforeUpload` untuk validasi tipe/ukuran (pesan Indonesia), `maxCount`, `accept`; JANGAN ubah endpoint / nama field upload yang sudah ada |
| Langkah / wizard | `Steps` | `items={[{ title: ... }]}`, `current` |
| Kotak pesan / error form | `Alert` | `type="success" \| "info" \| "warning" \| "error"`, `showIcon`, `message` |
| Data kosong | `Empty` | `description="Belum ada data"` (jangan biarkan teks Inggris "No data") |
| Label + nilai (detail) | `Descriptions` | `items={[{ label, children }]}`, `bordered`, `column={1}` di HP |
| Angka ringkas dashboard | `Statistic` | `title`, `value`, `suffix`; di dalam `Card` |
| Pagination bawah tabel | `Pagination` | Bila tabel tidak pakai `Table` bawaan; teks ikut `locale={idID}` |
| Layout admin (header + sidebar) | `Layout` (`Layout.Header`, `Layout.Sider`, `Layout.Content`) + `Menu` | Tim Layout yang mengerjakan; `Menu` dengan `items`, `selectedKeys` dari route aktif |
| Ikon | `@ant-design/icons` ATAU lucide yang sudah ada | Boleh salah satu; jangan campur dua gaya ikon dalam satu halaman baru. Contoh: `import { PlusOutlined } from "@ant-design/icons"` |

Aturan umum:

- **Tailwind TETAP dipakai** untuk layout/spacing (`flex`, `gap`, `px/py`,
  `max-w-*`, `grid`, `overflow-x-auto`) — jangan menulis ulang layout
  dengan inline style.
- **Komponen UI pakai antd**, bukan class Tailwind rakitan
  (`rounded border bg-...` untuk tombol/input/tabel/modal).
- Jangan hardcode hex `#1e4e9e` di `className` / `style` baru — pakai
  `type="primary"` / token theme. Token Tailwind lama (`bg-anbk`,
  `text-primary`, dll) hanya dipakai halaman yang belum di-redesign.
- Teks antarmuka **Bahasa Indonesia**. Jangan biarkan teks Inggris bawaan
  antd muncul (cek `Empty`, `Table`, `Upload`, `Modal`, `Pagination`).
- Status selalu warna + teks. Target sentuh min 44px dipertahankan.
  Teks "Ragu-ragu", "Sebelumnya", "Berikutnya", "Selesai" jangan diubah.

## 2. `Select` showSearch sebagai pengganti Dropdown custom

Dropdown custom (`components/ui/Dropdown.tsx`, wrapper `ui/Select.tsx`)
diganti `Select` antd. Pola pengganti — **satu-satunya pola yang dipakai**:

```tsx
import { Select } from "antd";

// options: { value: string; label: string }[] — value = UUID / id dari API,
// label = nama yang dibaca manusia. JANGAN tampilkan UUID mentah.
<Select
  showSearch
  allowClear
  placeholder="Pilih mapel"
  options={mapelOptions}
  value={mapelId || undefined}
  onChange={(v) => setMapelId(v ?? "")}
  filterOption={(input, opt) =>
    (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())
  }
  optionFilterProp="label"
  style={{ width: "100%" }}
  size="large"
  notFoundContent="Tidak ada data"
/>
```

- `showSearch` + `filterOption` = cari ketik (pengganti searchable Dropdown).
- `allowClear` = tombol hapus pilihan (pengganti opsi kosong manual).
- `value` harus `undefined` saat kosong, bukan `""` — kalau tidak,
  antd menampilkan pilihan kosong yang aneh.
- `placeholder` Bahasa Indonesia, selalu diisi.
- `notFoundContent="Tidak ada data"` — jangan biarkan "No data".
- Untuk Select di dalam `Form`, pakai `Form.Item` (§3), bukan `value` +
  `onChange` manual (kecuali filter yang memang di luar Form).
- Jangan pakai `mode="multiple"` / `mode="tags"` kecuali UI lama memang
  multi-pilih — jangan menambah perilaku baru.

## 3. `Form` + rules dengan pesan error Bahasa Indonesia

Pola standar untuk form admin (ganti `ui/Input` + validasi manual):

```tsx
import { Button, Form, Input, Select } from "antd";

<Form
  layout="vertical"
  onFinish={(values) => simpan(values)} // values = nama field yang SAMA
  requiredMark={false}
>
  <Form.Item
    label="Nama jurusan"
    name="nama"
    rules={[
      { required: true, message: "Nama jurusan wajib diisi." },
      { max: 100, message: "Nama jurusan maksimal 100 karakter." },
    ]}
  >
    <Input placeholder="Contoh: Teknik Informatika" size="large" />
  </Form.Item>

  <Form.Item
    label="Mapel"
    name="mapel_id"
    rules={[{ required: true, message: "Mapel wajib dipilih." }]}
  >
    <Select showSearch placeholder="Pilih mapel" options={mapelOptions} size="large" />
  </Form.Item>

  <Form.Item>
    <Button type="primary" htmlType="submit" loading={menyimpan} size="large">
      Simpan
    </Button>
  </Form.Item>
</Form>
```

- `name` Form.Item **harus sama** dengan nama field payload API yang
  sudah ada — jangan ganti nama field.
- Setiap `rule` wajib punya `message` Bahasa Indonesia, diakhiri titik.
- Validasi di Form hanya validasi **presentasi** (wajib diisi, panjang,
  format). Validasi bisnis / server tetap milik backend + pesan error
  server ditampilkan apa adanya via `Alert`.
- `layout="vertical"` (label di atas input) — konsisten dengan form lama.
- Untuk form edit: isi awal via `initialValues` / `form.setFieldsValue`,
  jangan ubah bentuk data yang dikirim ke API.
- `Input.Password` untuk password; jangan pernah log / tampilkan password.
- LoginPage sengaja TIDAK memakai `Form` antd (mempertahankan
  `onSubmit` + state lama apa adanya sebagai contoh migrasi minimal).
  Form baru memakai pola `Form` di atas.

## 4. LARANGAN (mengikat semua tim)

1. **Jangan ubah route** (`App.tsx`), endpoint API, nama field, payload,
   grading, auth, token, timer, anti-curang, atau logika bisnis apa pun.
   Hanya presentasi.
2. **Jangan sentuh `<select>` native di
   `components/exam/SoalRenderer.tsx:215`.** Itu kontrol JAWABAN siswa
   soal Menjodohkan / Peta Konsep — pengecualian yang **disetujui user**,
   sengaja tidak dimigrasi ke Dropdown/Select karena jalur penilaian dan
   aksesibilitas keyboard. Tim Siswa dilarang "memperbaikinya".
3. **Jangan hapus komponen custom** (`ui/Button`, `ui/Input`, `ui/Card`,
   `ui/Table`, `ui/Dialog`, `ui/Dropdown`, `ui/Select`, `common/*`) yang
   masih dipakai halaman lain. Hapus hanya bila halaman pemakainya sudah
   diganti DAN `grep` membuktikan tidak ada pemakai lagi.
4. **Tanpa logo / nama instansi resmi pemerintah, tanpa klaim aplikasi
   ANBK resmi.** Nama tetap "CBT SMK". Footer disclaimer
   **"Bukan aplikasi ujian nasional resmi."** tetap ada di Login dan
   Layout — jangan hapus / ubah redaksinya. Kata "ANBK" hanya boleh di
   nama class CSS / komentar kode yang sudah ada, bukan di teks yang
   dibaca user.
5. **Tanpa CDN / font eksternal.** Semua aset dari bundle lokal (LAN
   sekolah offline). Jangan tambah `<link>`, `@import url(...)`, atau
   dependensi yang memuat aset dari internet saat runtime.
6. **Jangan ubah teks ujian siswa** ("Ragu-ragu", "Sebelumnya",
   "Berikutnya", "Selesai", label status) dan jangan ubah perilaku
   komponen ujian (timer, navigasi soal, anti-curang).
7. **Jangan ganti RichEditor (TipTap) / RichContent / KaTeX** — tetap
   dipakai untuk stimulus, pertanyaan, pembahasan, dan rumus.
8. **Jangan ubah backend Go.** Bila butuh endpoint baru, itu di luar
   redesign — laporkan, jangan akali dari frontend.

## 5. Checklist sebelum tim berikutnya bilang "selesai"

- [ ] `npx tsc --noEmit -p frontend/tsconfig.json` lolos.
- [ ] `npm --prefix frontend run build` lolos.
- [ ] Tidak ada teks Inggris bawaan antd yang tertinggal (cek Empty,
      Table, Upload, Modal, Select kosong, Pagination).
- [ ] Tidak ada `console.log`, `TODO`, atau komponen custom yang
      dihapus diam-diam.
- [ ] `grep -rn "from \"antd\""` menunjukkan hanya file yang memang
      di-redesign; halaman lain tidak tersentuh.
- [ ] Disclaimer "Bukan aplikasi ujian nasional resmi." masih ada.
- [ ] Halaman yang diubah diuji manual minimal 1x di browser (form
      tersimpan, tabel tampil, modal buka/tutup).
