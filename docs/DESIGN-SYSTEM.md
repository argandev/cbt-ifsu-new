# Design System — CBT SMK v2 (Token + Rujukan Ringkas)

> **Satu sumber kebenaran:** token antd di `ConfigProvider`
> (`frontend/src/App.tsx`). Dokumen ini hanya ringkasan untuk tim lain —
> standar lengkap dan mengikat ada di `docs/AUDIT-KONSISTENSI.md` §2
> (dan §3 "Jangan Dirusak"). Bila dokumen ini dan audit bertentangan,
> **audit yang menang**.
>
> **Aturan tertinggi:** hanya lapisan presentasi. Jangan ubah route, API,
> logika, grading, auth, token, timer, anti-curang, validasi, nama field,
> payload. Hex mentah di luar token DILARANG. Tanpa CDN. Tanpa logo/nama
> instansi resmi. Tanpa klaim aplikasi ujian nasional resmi.

## 1. Palet (1 palet — nilai TETAP)

| Peran | Nilai | Token antd / pemakaian |
|---|---|---|
| Primer | `#1e4e9e` | `colorPrimary`; `Button type="primary"`, `Tag color="blue"` (info), header |
| Primer hover | `#173d7a` | `colorPrimaryHover` — SATU nilai (nilai ganda lama `#163d7d` dihapus) |
| Primer active / pekat | `#122f5e` | antd active; jangan hardcode |
| Primer muda (latar lembut) | `#e8effa` | Latar lembut (hover nav, ikon stat) |
| Primer soft | `#dbe6f7` | Latar lembut lebih pekat |
| Sukses | `#15803d` / latar `#dcfce7` | `colorSuccess`; `Tag color="green"`, `Alert type="success"` |
| Peringatan | `#b45309` / latar `#fef3c7` | `colorWarning`; `Tag color="gold"`, `Alert type="warning"` |
| Error / bahaya | `#dc2626` / latar `#fee2e2` | `colorError`; `Tag color="red"`, `Alert type="error"`, `Button danger` |
| Info | `#1e4e9e` / latar `#dbeafe` | `colorInfo` (= primer); `Tag color="blue"`, `Alert type="info"` |
| Nonaktif | `#64748b` / latar `#f1f5f9` | `Tag color="default"` |
| Teks utama | `#0f172a` | `colorText` |
| Teks sekunder | `#475569` | `colorTextSecondary` |
| Garis | `#cbd5e1` | `colorBorder` |
| Latar halaman | `#eef2f7` | `colorBgLayout` |
| Putih | `#ffffff` | `token.colorWhite` / antd default — dilarang tulis `#ffffff` di `style` |
| Hitam (print) | `#000` | Hanya `print.css` — diizinkan |

**Peta status → Tag (SATU peta, detail di audit §2.2):** Aktif / Lulus /
Dijawab / Terbuka / Terkirim / Disetujui / Terbaca → `green`; Menunggu /
Jeda / Ragu-ragu / Terkunci → `gold`; Diblokir / Tidak Lulus / Ditolak →
`red`; Berjalan / Selesai / Dikerjakan → `blue`; Draft / Tidak Aktif /
Belum Dijawab → `default`. Status SELALU warna + teks, bukan warna saja.

**Cara pakai warna:** lewat `type`/`color`/`status` antd, atau
`theme.useToken()` bila tidak terwakili (mis. latar kotak grid soal).
DILARANG menulis hex di komponen. Token Tailwind `@theme` di `index.css`
hanya jembatan migrasi (nilai sama persis) — dihapus setelah grep
`bg-anbk|text-anbk|…` = 0.

## 2. Tipografi (1 skala — detail audit §2.3)

| Elemen | Pola | Nilai |
|---|---|---|
| Judul halaman (H1) | antd `Typography.Title level={3}` | 24px, 700 |
| Judul kartu | antd `Card title` / `Modal title` (bawaan) | 16px, 600 |
| Judul seksi (H2) | antd `Typography.Title level={4}` | 18px, 700 |
| Body | antd default | 14px, 400, `colorText` |
| Teks soal / opsi jawaban | Pengecualian tercatat | 16px (keterbacaan soal; bukan pola umum) |
| Label form | antd `Form.Item label` | 14px, 600 |
| Caption / teks bantu | antd `Typography.Text type="secondary"` | 12px |
| Timer ujian | antd `Tag` + `fontFamily: token.fontFamilyCode` | 22px, 700, monospace |
| Angka besar hasil | antd `Statistic` / `Typography.Title` | Sesuai konteks |
| Halaman error (404/403) | antd `Result` (`status="404"/"403"/"error"`) | Menggantikan `text-6xl` rakitan |

Font: **Plus Jakarta Sans** (lokal via `@fontsource`, tanpa CDN).
DILARANG: `text-[15px]/[16px]/[17px]` arbitrer di luar teks soal,
`text-xl/text-2xl` manual untuk judul halaman, `tracking-wide/widest`
kecuali brand `CBT SMK` di header.

## 3. Tombol (1 pola — detail audit §2.12)

- Komponen: **antd `Button` — SELALU**. DILARANG `<button>` mentah,
  `ui/Button`, tautan bergaya tombol (`<Link>` → antd `Button href=`).
- Varian: `type="primary"` = aksi utama (MAKSIMAL 1 per tampilan/modal);
  `type="default"` = sekunder; `type="text"` = tombol ikon / aksi ringan;
  `danger` = hapus/berbahaya (di modal konfirmasi: `type="primary" danger`;
  di baris tabel: `type="default" danger`). `type="link"` / `dashed`
  DILARANG kecuali disetujui koordinator.
- Ukuran: `size="large"` untuk CTA utama + tombol submit form; default
  (middle) untuk semua tombol lain; **`size="small"` DILARANG** untuk
  Button (pengecualian satu-satunya: toolbar `RichEditor`).
- State: mutasi → `loading={mutation.isPending}` (spinner), BUKAN `disabled`
  saja. `disabled` hanya untuk yang benar-benar tidak tersedia.
- Tombol ikon: `icon={<…/>}` (antd icons untuk kode baru) + `aria-label`.
- Teks loading Bahasa Indonesia: `Menyimpan…`, `Menghapus…`, `Memproses…`.

## 4. Spacing, radius, shadow (detail audit §2.4–§2.5)

| Konteks | Pola |
|---|---|
| Jarak antar blok halaman | `gap-6` (24px) / antd `Space size={24}` |
| Jarak antar field form | `gap-4` (16px) via antd `Form` (vertical) |
| Jarak grid kartu / tabel | `gap-4` (16px) |
| Padding konten halaman | Admin `px-6 py-6` (konten `max-w-1280`); siswa `px-4 py-6` (konten `max-w-1100`) — ditulis SATU kali di layout, bukan per halaman |
| Padding Card / Modal | antd default (24px) — DILARANG override `p-4/p-5/p-6/p-8` |
| Baris tombol aksi (modal/form) | antd `Space` / `Flex justify="flex-end" gap={8}` |
| Target sentuh | min 44px — via `controlHeightLG: 44` + `size="large"` untuk aksi utama; grid nomor soal min 44px |

- **Radius: 8px untuk SEMUA** (token `borderRadius: 8`; `rounded-lg` =
  satu-satunya radius komponen). `rounded-xl/2xl/md`, `rounded` (4),
  `rounded-sm` DILARANG untuk komponen UI. Pengecualian: `rounded-full`
  HANYA untuk `Avatar` / lencana ikon lingkaran.
- **Shadow: antd default saja.** `shadow-xl/2xl/lg/md` dan shadow arbitrer
  DIHAPUS — bar bawah ujian memakai `border-t` saja, tanpa shadow.

## 5. Pola komponen (ringkas — detail audit §2.6–§2.8)

Header halaman: antd `Breadcrumb` (mulai dari `Dashboard` `/adm` untuk
admin) + `Typography.Title level={3}` + deskripsi + aksi kanan.
Kartu: antd `Card` (DILARANG `border-t-4` aksen dan wrapper `div
rounded-xl border … bg-white shadow-sm`). Statistik: antd `Card` +
`Statistic`. Modal: antd `Modal` (konfirmasi 520, form 720, default 520;
saat loading `maskClosable={false} keyboard={false}`). Konfirmasi hapus:
antd `Modal.confirm` (`okText="Ya, Hapus"`, `cancelText="Batal"`).
Empty: antd `Empty` (deskripsi Bahasa Indonesia). Filter: antd
`Input.Search` (`allowClear`, `size="large"`, placeholder `Cari…`) +
`Select showSearch` (placeholder `Pilih…`, `notFoundContent="Tidak ada
data"`). Tabel: antd `Table` + `scroll={{ x: "max-content" }}` — DILARANG
wrapper `div overflow-x-auto` ganda dan `<table>` mentah untuk data
(kecuali dokumen cetak). Pesan: antd `Alert`. Label-nilai detail: antd
`Descriptions`. Ikon: satu set `@ant-design/icons` untuk kode baru/diubah.

*Ditulis Tim Design Token + Sistem, 17 Okt 2026 (Fase 0). Standar mengikat:
`docs/AUDIT-KONSISTENSI.md` §2–§3.*
