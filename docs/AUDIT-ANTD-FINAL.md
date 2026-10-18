# AUDIT Ant Design — CBT SMK v2 (FINAL)

> **Tim Audit (8 orang) — HANYA audit, TIDAK mengubah file source.**
> Tanggal audit: 18 Okt 2026. Ruang lingkup: seluruh `frontend/src`
> (58 file `.tsx`, file `.ts`, `index.css`, `print.css`).
> Metode: `grep` + baca kode di sekitar tiap temuan.
> Referensi desain: `src/App.tsx` (`ConfigProvider`, token),
> `src/ANTD-PANDUAN.md`, `src/index.css` (jembatan `@theme`).

## Ringkasan jumlah temuan per kategori

| # | Kategori | Jumlah | Prioritas dominan |
|---|----------|--------|-------------------|
| 1 | `<Input type="date">` / `type="time">` / `datetime-local` | **6** (date 4, time 2, datetime-local 0) | **P0** |
| 2 | `<select>` native | **0** (hanya 1 komentar) | — |
| 3 | `<textarea>` native | **0** | — |
| 4 | `<input type="checkbox">` / `type="radio">` native | **0** | — |
| 5 | `<input type="file">` native | **2** (1 raw `<input>`, 1 `<Input type="file">`) | **P0** |
| 6 | `<Input type="number">` | **13** | **P0** |
| 7 | `<button>` native | **0** (183× antd `<Button>`) | — |
| 8 | `<form>` native | **0** (semua antd `<Form>`) | — |
| 9 | `window.confirm` / `window.alert` / `window.prompt` | **0** (15× antd `Modal.confirm`) | — |
| 10 | Hex warna mentah di luar token | **0 di `.tsx`/`.ts`** (selain token `App.tsx` + komentar); CSS: jembatan `@theme` (by design) + gaya konten/print (P2) | P2 |
| 11 | Class Tailwind warna non-token (`*-slate-500`, `bg-red-50`, …) | **472** kemunculan di **44 file** + 47 class `bg-white`/`text-white` | **P1** |
| 12a | Alert rakitan (`<div>`/`<p role="alert">`, bukan antd `Alert`) | **34** | **P1** |
| 12b | Teks error tanpa `role` + kotak info/status rakitan lain | **10** blok teks + **~10** kotak | P1/P2 |
| 12c | Definisi komponen custom duplikat antd (Dropdown/Select/Modal/Alert/…) | **0** | — |
| 13 | Pola yang sudah benar: `DatePicker` **0** pakai, `TimePicker` **0** pakai, `Upload` 2 file, `InputNumber` 1 pakai, `Input.TextArea` 7 pakai, `dayjs` **bukan** dependensi langsung | — (lihat §13) | — |

**Total temuan yang perlu aksi: 21 kontrol input P0 + 34 Alert rakitan P1 + 472 class warna P1 + ~20 teks/kotak P1/P2 + beberapa P2.**
**Kabar baik:** kategori 2, 3, 4, 7, 8, 9 sudah **NOL** — migrasi sebelumnya berhasil untuk
`<select>`, `<textarea>`, checkbox/radio native, `<button>`, `<form>`, dan dialog konfirmasi.
Sisa pekerjaan terkonsentrasi di: **input tanggal/jam/angka/file (21)** + **Alert rakitan (34)** +
**class warna Tailwind (472)**.

### Definisi prioritas

- **P0** = kontrol input yang dipakai user langsung (tanggal/jam/angka/file). Wajib migrasi.
- **P1** = tampilan (Alert rakitan, class warna, kotak info). Migrasi setelah P0.
- **P2** = kode mati / komentar basi / observasi, tidak memengaruhi user. Boleh ditunda / dihapus sekalian.

### Catatan penting tentang bentuk temuan

Hampir semua temuan kategori 1, 5, 6 **bukan** elemen HTML mentah, melainkan
komponen antd `<Input>` dengan prop `type="date" / "time" / "number" / "file"` —
antd meneruskan `type` itu ke `<input>` browser sehingga yang tampil tetap
**picker bawaan browser** (bukan `DatePicker`/`TimePicker`/`InputNumber`/`Upload` antd).
Satu-satunya pengecualian elemen mentah adalah `ImportSiswaPage.tsx:89` (`<input type="file">`)
dan `SettingsPage.tsx:189` (`<input type="hidden">`, P2).

---

## 1. Input tanggal / jam — 6 temuan (P0)

Pengganti: antd `DatePicker` (tanggal) / `TimePicker` (jam).
**Kontrak nilai yang harus dipertahankan:** state & payload API tetap string mentah
`YYYY-MM-DD` untuk tanggal dan `HH:mm` untuk jam (lihat §13 — `format.ts` eksplisit
menyatakan input `type="date"/"time"` memakai format mentah). Konversi hanya di batas
komponen: `DatePicker value={tanggal ? dayjs(tanggal, "YYYY-MM-DD") : null}`
+ `onChange={(d) => set(..., d ? d.format("YYYY-MM-DD") : "")}`; `TimePicker` analog
dengan format `"HH:mm"`. `dayjs` belum dependensi langsung — lihat §13.

| File | Baris | Kode yang ditemukan | Pengganti antd | Prioritas |
|------|-------|---------------------|----------------|-----------|
| `pages/admin/jadwal/JadwalFormPage.tsx` | 70 | `<Form.Item label="Tanggal"><Input type="date" value={form.tanggal} onChange={(e) => setForm({ ...form, tanggal: e.target.value })} required size="large" /></Form.Item>` | `DatePicker format="YYYY-MM-DD" size="large"` | P0 |
| `pages/admin/jadwal/JadwalFormPage.tsx` | 72 | `<Form.Item label="Jam mulai"><Input type="time" value={form.jam_mulai} onChange={(e) => setForm({ ...form, jam_mulai: e.target.value })} required size="large" /></Form.Item>` | `TimePicker format="HH:mm" size="large"` | P0 |
| `pages/admin/jadwal/JadwalFormPage.tsx` | 73 | `<Form.Item label="Jam selesai"><Input type="time" value={form.jam_selesai} onChange={(e) => setForm({ ...form, jam_selesai: e.target.value })} required size="large" /></Form.Item>` | `TimePicker format="HH:mm" size="large"` | P0 |
| `pages/admin/ukk/UkkListPage.tsx` | 185 | `<Form.Item label="Tanggal"><Input type="date" value={form.tanggal} onChange={(e) => setField("tanggal", e.target.value)} size="large" /></Form.Item>` | `DatePicker format="YYYY-MM-DD" size="large"` | P0 |
| `pages/admin/event/EventPage.tsx` | 215 | `<Form.Item label="Tanggal mulai"><Input type="date" value={form.tanggal_mulai} onChange={(e) => setForm({ ...form, tanggal_mulai: e.target.value })} size="large" /></Form.Item>` | `DatePicker format="YYYY-MM-DD" size="large"` | P0 |
| `pages/admin/event/EventPage.tsx` | 216 | `<Form.Item label="Tanggal selesai"><Input type="date" value={form.tanggal_selesai} onChange={(e) => setForm({ ...form, tanggal_selesai: e.target.value })} size="large" /></Form.Item>` | `DatePicker format="YYYY-MM-DD" size="large"` | P0 |

Tidak ada `type="datetime-local"` di mana pun (grep = 0).

## 2. `<select>` native — 0 temuan ✅

Tidak ada `<select>` native. Satu-satunya kemunculan kata `<select>` adalah komentar
di `components/exam/SoalRenderer.tsx:209` (“`// <select> native (panah + Enter) pada jalur
penilaian ini.`”) — **bukan kode**, jangan diubah (file ujian: grading/timer/anti-curang
bukan wewenang Tim Tampilan; lihat peringatan di §14).
`Select` antd sudah dipakai luas (±40+ JSX, `showSearch`, `allowClear`, `filterOption`).

## 3. `<textarea>` native — 0 temuan ✅

Tidak ada `<textarea>`. `Input.TextArea` antd sudah dipakai 7× (lihat §13).

## 4. Checkbox / radio native — 0 temuan ✅

Tidak ada `<input type="checkbox">` / `type="radio">`.
`Checkbox` antd 10× (`UjianPage`, `SettingsPage`, `GuruPage`, `SiswaPage`,
`PaketFormPage` ×3, `SoalEditor` ×2, `exam/SoalRenderer`).
`Radio.Group` antd 2× (`ImportSoalPanel` + `RichEditor`).
`Switch` antd: **0 pemakaian** — bukan pelanggaran; bila Tim Form butuh toggle,
pakai `Switch`, jangan `<input type="checkbox">`.

## 5. Input file — 2 temuan (P0)

| File | Baris | Kode yang ditemukan | Pengganti antd | Prioritas |
|------|-------|---------------------|----------------|-----------|
| `pages/admin/master/ImportSiswaPage.tsx` | 89–96 | `<input type="file" accept=".csv,text/csv" className="w-full text-sm" onChange={(e) => { setFile(e.target.files?.[0] ?? null); setResult(null); setError(""); }} />` (di dalam `<label>`, baris 87–100) — **satu-satunya `<input>` mentah di seluruh `src`** | `Upload` dengan `beforeUpload={(f) => { setFile(f); return false; }}` + `maxCount={1}` + `showUploadList` — tiru pola `ImportSoalPanel` (§13). Pertahankan `accept=".csv,text/csv"`, batas 2.000 baris / 5 MB, dan perilaku “file dipilih → tombol Upload & Import aktif”. | P0 |
| `pages/admin/ukk/UkkDetailPage.tsx` | 223 | `<Input id="berkas-file" type="file" accept=".pdf,.doc,.docx,.jpg,.jpeg,.png,.zip" onChange={(e) => setUploadFile(e.target.files?.[0] ?? null)} size="large" />` | `Upload` (`beforeUpload` return false, `maxCount={1}`) — pertahankan `accept`, maks 10 MB, dan `Form`/`onFinish={handleUpload}` di sekitarnya. Perhatian: di baris 226 ada `<Upload size={16} />` — itu **ikon lucide-react**, bukan komponen antd; setelah migrasi akan ada dua `Upload` (ikon + antd) — beri alias impor ikonnya agar tidak bentrok nama. | P0 |

## 6. Input angka — 13 temuan (P0)

Pengganti: antd `InputNumber`. **Kontrak nilai yang harus dipertahankan:**
`InputNumber onChange` mengembalikan `number | null` (bukan event) — sesuaikan handler
(`onChange={(v) => set(..., v ?? 0)}` atau `typeof v === "number" ? v : …`).
Perhatikan kasus khusus: `RekapPage` dan `SiswaPage` saat ini menyimpan nilai sebagai
**string** (`e.target.value`); `UkkNilaiPage` juga string; `SettingsPage`/`PaketFormPage`/
`PaketDetailPage` sudah `Number(...)`. Tim Form harus memastikan payload API tidak berubah
(cek tipe yang diterima backend: angka vs string angka — jangan menebak, cek pemakaian
`Number(...)` yang sudah ada di tiap file).
Pola acuan yang sudah benar: `SoalEditor.tsx:829` (`<InputNumber id="bobot" min={1} … />`).

| File | Baris | Kode yang ditemukan | Pengganti antd | Prioritas |
|------|-------|---------------------|----------------|-----------|
| `pages/admin/settings/SettingsPage.tsx` | 131 | `<Input id="token-expiry" type="number" min={60} max={86400} value={form.token_expiry_detik ?? 900} onChange={(e) => setField("token_expiry_detik", Number(e.target.value))} size="large" />` | `InputNumber min={60} max={86400} size="large"` | P0 |
| `pages/admin/settings/SettingsPage.tsx` | 134 | `<Input id="toleransi" type="number" min={0} max={120} value={form.toleransi_telat_menit ?? 15} onChange={(e) => setField("toleransi_telat_menit", Number(e.target.value))} size="large" />` | `InputNumber min={0} max={120} size="large"` | P0 |
| `pages/admin/settings/SettingsPage.tsx` | 137 | `<Input id="max-pelanggaran" type="number" min={1} max={10} value={form.max_pelanggaran ?? 2} onChange={(e) => setField("max_pelanggaran", Number(e.target.value))} size="large" />` | `InputNumber min={1} max={10} size="large"` | P0 |
| `pages/admin/settings/SettingsPage.tsx` | 141 | `<Input id="durasi" type="number" min={5} max={600} value={form.durasi_default ?? 90} onChange={(e) => setField("durasi_default", Number(e.target.value))} size="large" />` | `InputNumber min={5} max={600} size="large"` | P0 |
| `pages/admin/rekap/RekapPage.tsx` | 208 | `<Input id={\`remedial-${r.siswa_id}\`} type="number" min={0} max={100} step="0.01" value={remedialNilai[r.siswa_id] ?? ""} onChange={(e) => setRemedialNilai((prev) => ({ ...prev, [r.siswa_id]: e.target.value }))} placeholder="0-100" size="large" />` | `InputNumber min={0} max={100} step={0.01}` — **hati-hati**: state saat ini string per-siswa (`remedialNilai`), dan `handleRemedial` memakainya; konversi `number ↔ string` harus dijaga agar payload remedial tidak berubah. | P0 |
| `pages/admin/master/SiswaPage.tsx` | 258 | `<Form.Item label="Tahun Masuk (opsional)"><Input type="number" value={tahunMasuk} onChange={(e) => setTahunMasuk(e.target.value)} placeholder="mis. 2026" size="large" /></Form.Item>` | `InputNumber` — **hati-hati**: state `tahunMasuk` saat ini string; jaga konversi agar payload tidak berubah (string kosong vs `null` saat dikosongkan). | P0 |
| `pages/admin/paket/PaketFormPage.tsx` | 74 | `<Form.Item label="Durasi (menit)"><Input type="number" min={1} value={form.durasi_menit} onChange={(e) => setForm({ ...form, durasi_menit: Number(e.target.value) })} required size="large" /></Form.Item>` | `InputNumber min={1} size="large"` | P0 |
| `pages/admin/paket/PaketDetailPage.tsx` | 110 | `<Input type="number" min={1} value={jumlahAcak} onChange={(e) => setJumlahAcak(Number(e.target.value))} aria-label="Jumlah soal acak" size="large" />` | `InputNumber min={1} size="large" aria-label` dipertahankan | P0 |
| `pages/admin/ukk/UkkNilaiPage.tsx` | 197 | `<Form.Item label="Nilai Teori (0-100)"><Input type="number" min={0} max={100} step="0.01" value={form.nilai_teori} onChange={(e) => setField("nilai_teori", e.target.value)} placeholder="mis. 85" size="large" /></Form.Item>` | `InputNumber min={0} max={100} step={0.01}` — **hati-hati**: state string; jaga konversi. | P0 |
| `pages/admin/ukk/UkkNilaiPage.tsx` | 198 | `<Form.Item label="Rubrik Proses (0-100)"><Input type="number" … value={form.nilai_proses} onChange={(e) => setField("nilai_proses", e.target.value)} … /></Form.Item>` | `InputNumber min={0} max={100} step={0.01}` — sama, jaga konversi string. | P0 |
| `pages/admin/ukk/UkkNilaiPage.tsx` | 199 | `<Form.Item label="Rubrik Produk (0-100)"><Input type="number" … value={form.nilai_produk} onChange={(e) => setField("nilai_produk", e.target.value)} … /></Form.Item>` | `InputNumber min={0} max={100} step={0.01}` — sama. | P0 |
| `pages/admin/ukk/UkkNilaiPage.tsx` | 200 | `<Form.Item label="Rubrik Sikap (0-100)"><Input type="number" … value={form.nilai_sikap} onChange={(e) => setField("nilai_sikap", e.target.value)} … /></Form.Item>` | `InputNumber min={0} max={100} step={0.01}` — sama. | P0 |
| `pages/admin/event/RuanganPage.tsx` | 185 | `<Form.Item label="Kapasitas (kursi) *"><Input type="number" min={1} step={1} placeholder="mis. 36" value={form.kapasitas} onChange={(e) => setForm({ ...form, kapasitas: e.target.value })} size="large" /></Form.Item>` | `InputNumber min={1} step={1}` — **hati-hati**: state string (`e.target.value`); jaga konversi. | P0 |

Tidak ditemukan `type="number"` di file lain (total 13 = 4 Settings + 1 Rekap + 1 Siswa + 1 PaketForm + 1 PaketDetail + 4 UkkNilai + 1 Ruangan).
`<Input type="text">` (2×: `SettingsPage:145` nama sekolah, `LoginPage`-adjacent) dan
`<Input.Password>` (`LoginPage:108`, `GantiPasswordPage:98,112,126`) adalah pemakaian antd
yang benar — **bukan temuan**.

## 7. `<button>` native — 0 temuan ✅

Tidak ada `<button>`. antd `<Button>` dipakai 183×. `print.css` menyembunyikan `button`
saat print — itu gaya print, bukan komponen.

## 8. `<form>` native — 0 temuan ✅

Tidak ada `<form>`. Semua formulir memakai antd `<Form>` (`layout="vertical"`,
`onFinish`). `<label>` HTML mentah masih dipakai 18× (lihat §12c, observasi P2 —
`<label>` adalah elemen semantik, bukan komponen UI duplikat).

## 9. `window.confirm` / `alert` / `prompt` — 0 temuan ✅

Tidak ada `window.confirm` / `window.alert` / `window.prompt` (satu-satunya kemunculan
kata “confirm” adalah komentar `BankSoalDetailPage.tsx:31` dan antd `Modal.confirm`).
Konfirmasi memakai antd `Modal.confirm` 15×: `KelasPage:119`, `MapelPage:130`,
`GuruPage:120`, `SiswaPage:157`, `JurusanPage:118`, `BankSoalDetailPage:105,115,126`,
`PaketPage:257`, `JadwalPage:221`, `BackupPage:116,141`, `UkkListPage:110`,
`EventPage:138`, `RuanganPage:116`. Pola sudah benar — Tim Form tidak perlu
mengganti apa pun di sini.
`window.` lain yang sah (bukan dialog): `window.print()` (`BeritaAcaraPage`,
`UkkSertifikatPage`), `window.location` (parsing ID dari URL), `window.history.pushState`
(`UjianPage`), `window.addEventListener("blur")` (anti-curang `UjianPage` — JANGAN disentuh),
`window.setTimeout` (`ImportSoalPanel`).

## 10. Hex warna mentah di luar token

**Di `.tsx` / `.ts`: NOL hex di luar token.** Satu-satunya hex di file TSX adalah
12 kemunculan token di `App.tsx:87–98` (`colorPrimary #1e4e9e`, `colorPrimaryHover #173d7a`,
`colorSuccess #15803d`, `colorWarning #b45309`, `colorError #dc2626`, `colorInfo #1e4e9e`,
`colorText #0f172a`, `colorTextSecondary #475569`, `colorBorder #cbd5e1`,
`colorBgLayout #eef2f7` + 2 di komentar) — **ini sumber kebenaran, BUKAN temuan**.
Komentar di `Layout.tsx:4` / `AdminLayout.tsx:4` menyebut `#1e4e9e` sebagai dokumentasi — bukan kode.

| Lokasi | Temuan | Penilaian |
|--------|--------|-----------|
| `src/index.css` `@theme` (baris 53–84) | `--color-primary #1e4e9e`, `--color-primary-darker #122f5e`, `--color-primary-light #e8effa`, `--color-anbk #1e4e9e`, `--color-primary-soft #dbe6f7`, `--color-success #15803d`, `--color-success-bg #dcfce7`, `--color-warning #b45309`, `--color-warning-bg #fef3c7`, `--color-danger #dc2626`, `--color-danger-bg #fee2e2`, `--color-inactive #64748b`, `--color-surface #eef2f7`, `--color-ink #0f172a`, `--color-muted #475569`, `--color-border #cbd5e1` | **By design (jembatan migrasi)** — nilainya sama persis dengan token antd. Bukan pelanggaran; hapus blok `@theme` hanya bila pemakai class jembatan sudah 0 (saat ini masih 83 pemakaian, lihat §11). **P2** (backlog migrasi). |
| `src/index.css` gaya konten (baris 156, 166, 193, 212, 222, 246) | `.rich-content pre { background: #0f172a; color: #e2e8f0 }`, `code { background: #f1f5f9 }`, `th { background: #f1f5f9 }`, `.rich-video-frame { background: #000 }`, `div[data-video-embed] { background: #f8fafc }`, placeholder `{ color: #94a3b8 }` | Gaya konten editor (TipTap), bukan komponen UI. Bisa dipetakan ke `var(--color-…)` bila sempat, tapi tidak wajib. **P2**. |
| `src/print.css` (baris 29–30, 58, 61, 75) | `background: #fff`, `color: #000`, `border: 1px solid #000` (semua `!important`, di dalam `@media print`) | **Sengaja** — hemat tinta, print hitam-putih. **Jangan diubah. P2 / keep.** |
| `src/index.css:59` (komentar), `src/App.tsx:88` (komentar), `src/ANTD-PANDUAN.md:29,56` (komentar/dok) | Hex basi di komentar: `#2563EB`, `#1d4ed8`, `#163d7d`, `#2f6fd0` | Komentar “jangan pakai lagi” — tidak dieksekusi. Bersihkan sekalian bila menyentuh file. **P2 trivial.** |
| Class arbitrer `bg-[#…]` / `text-[#…]` di `className` | **0** (grep `\[#` = 0) | ✅ |

Tidak ada `rgb()/rgba()/hsl()` warna di `.tsx` kecuali `Layout.tsx:54`
`rgba(255, 255, 255, 0.15)` (latar hover tombol menu di header — memakai
`token.colorWhite`, dapat dibiarkan; P2).
Inline `style={{…}}` warna: hanya via `theme.useToken()` (`LoginPage:69`,
`GantiPasswordPage:62`, `NavigasiSoal:67`, `Layout:48,54,75`,
`AdminLayout:193,195,287,304`) — **pola benar** sesuai `ANTD-PANDUAN.md`
(“ambil dari `theme.useToken()`, dilarang menulis hex di komponen”).

## 11. Class Tailwind warna non-token — 472 kemunculan di 44 file (P1)

Ini kategori terbesar. Semua class di bawah memakai palet Tailwind (`slate/red/green/
amber/blue/…-NNN`), bukan token antd. **Bukan komponen input** — murni tampilan —
maka prioritasnya **P1** (bukan P0). Strategi migrasi yang disarankan Tim Tampilan:
(1) teks status → antd `Tag`/`Typography.Text type=`/peta `statusTag.ts` yang sudah ada;
(2) pesan error/sukses/info → antd `Alert` (lihat §12);
(3) latar/border status → token via `theme.useToken()` atau class jembatan `@theme`
yang sudah ada (`bg-success-bg`, `text-muted`, …) — **jangan** menambah token baru
tanpa persetujuan koordinator (aturan di `index.css`).
Pengecualian yang **bukan** temuan: `bg-white`/`text-white` (47×, warna netral kertas/teks
di atas header biru — dapat dibiarkan, atau ganti `token.colorWhite` bila menyentuh file)
dan 83 class jembatan token (`border-border`, `text-anbk`, `bg-anbk`, `text-primary`,
`text-muted`, `bg-primary-light`, `bg-success-bg`, `bg-warning-bg`, `bg-surface`,
`text-ink`, `ring-anbk`, …) — jembatan **by design**, bukan pelanggaran, tapi masuk
backlog hapus-`@theme` (P2).

### 11a. Distribusi per class (total 472)

| Class | Jumlah | Class | Jumlah | Class | Jumlah |
|---|---|---|---|---|---|
| `text-slate-500` | 100 | `border-slate-300` | 19 | `bg-green-50` | 4 |
| `text-slate-600` | 78 | `text-slate-900` | 14 | `bg-blue-50` | 4 |
| `text-slate-700` | 36 | `text-slate-400` | 11 | `text-red-800` | 3 |
| `border-slate-200` | 30 | `bg-slate-100` | 11 | `text-blue-100` | 3 |
| `bg-red-50` | 27 | `bg-slate-50` | 9 | `bg-slate-300` | 3 |
| `text-red-700` | 26 | `text-slate-800` | 6 | `bg-green-100` | 3 |
| `border-red-200` | 24 | `text-green-700` | 6 | `bg-amber-50` | 3 |
| `text-red-600` | 21 | `text-green-800` | 4 | `text-blue-700` | 2 |
| | | `text-green-600` | 4 | `border-green-500` | 2 |
| | | `text-amber-800` | 4 | `border-green-200` | 2 |
| | | `text-amber-700` | 4 | `bg-amber-100` | 2 |
| | | | | `text-amber-900/600/500`, `border-slate-400/100`, `border-amber-300/200` | 1 masing-masing |

### 11b. Jumlah per file (44 file, total 472)

| File | Jumlah | File | Jumlah |
|------|--------|------|--------|
| `pages/siswa/UjianPage.tsx` | 38 | `pages/admin/master/SiswaPage.tsx` | 10 |
| `pages/admin/banksoal/ImportSoalPanel.tsx` | 37 | `pages/admin/master/MapelPage.tsx` | 10 |
| `components/soal/SoalEditor.tsx` | 33 | `pages/admin/master/KelasPage.tsx` | 10 |
| `components/soal/SoalRenderer.tsx` | 28 | `pages/admin/master/JurusanPage.tsx` | 10 |
| `components/soal/RichEditor.tsx` | 26 | `pages/admin/banksoal/BankSoalDetailPage.tsx` | 9 |
| `pages/admin/master/ImportSiswaPage.tsx` | 21 | `pages/admin/settings/SettingsPage.tsx` | 8 |
| `pages/admin/rekap/BeritaAcaraPage.tsx` | 19 | `pages/admin/analisis/AnalisisButirPage.tsx` | 8 |
| `pages/admin/banksoal/PanduanImportPage.tsx` | 16 | `pages/siswa/SiswaHomePage.tsx` | 7 |
| `pages/admin/monitor/MonitorPage.tsx` | 15 | `pages/admin/ukk/UkkDetailPage.tsx` | 7 |
| `pages/admin/master/GuruPage.tsx` | 14 | `pages/admin/rekap/RekapPage.tsx` | 7 |
| `components/exam/SoalRenderer.tsx` | 14 | `components/common/NotifikasiBell.tsx` | 7 |
| `pages/admin/ukk/UkkSertifikatPage.tsx` | 12 | `pages/admin/ukk/UkkNilaiPage.tsx` | 6 |
| `pages/admin/paket/PaketDetailPage.tsx` | 12 | `pages/admin/analisis/KompetensiPage.tsx` | 6 |
| `pages/admin/event/RuanganPage.tsx` | 11 | `pages/admin/ukk/UkkListPage.tsx` | 5 |
| `pages/admin/event/EventPage.tsx` | 11 | `pages/admin/jadwal/JadwalPage.tsx` | 5 |
| `pages/admin/backup/BackupPage.tsx` | 11 | `pages/admin/banksoal/BankSoalListPage.tsx` | 5 |
| `pages/siswa/HasilPage.tsx` | 10 | `components/ui/ImageUpload.tsx` | 5 |
| `pages/siswa/KartuSayaPage.tsx` | 3 | `pages/admin/kartu/KartuPesertaPage.tsx` | 3 |
| `pages/admin/banksoal/BankSoalFormPage.tsx` | 3 | `pages/admin/paket/PaketPage.tsx` | 2 |
| `pages/admin/paket/PaketFormPage.tsx` | 1 | `components/analisis/PengecohTable.tsx` | 2 |
| `pages/admin/jadwal/JadwalFormPage.tsx` | 1 | `components/analisis/DistribusiChart.tsx` | 2 |
| `components/exam/NavigasiSoal.tsx` | 1 | `components/analisis/KompetensiHeatmap.tsx` | 1 |

Daftar baris lengkap dapat direproduksi dengan:
`grep -rnE '(bg|text|border|ring|divide|from|via|to|fill|stroke)-(slate|gray|red|orange|amber|yellow|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|pink|rose|white|black)-[0-9]+' frontend/src --include='*.tsx'`.
Catatan: `DistribusiChart.tsx` memakai `fill="var(--color-anbk)"` / `var(--color-ink)` /
`var(--color-inactive)` / `stroke="var(--color-border)"` (baris 60, 70, 80, 92) — **pola benar**
(token via `var()`), 2 hitungan di atas berasal dari class lain di file itu.
`KompetensiHeatmap.tsx` memakai `Tag color={warnaTag(v)}` antd — **pola benar**.

## 12. Komponen custom duplikat antd

### 12a. Definisi komponen custom duplikat — 0 ✅

Tidak ada definisi `Dropdown` / `Select` / `Modal` / `Alert` / `ConfirmDialog` /
`StatusBadge` / `PageHeader` / `EmptyState` / `StatCard` / `Pagination` /
`SearchInput` rakitan sendiri (grep definisi fungsi/konstanta = 0).
`components/ui/` hanya berisi **satu** file: `ImageUpload.tsx` — itu wrapper sah
di atas antd `Upload` (+ `Upload.Dragger`), dipakai `SettingsPage` (logo sekolah),
bukan duplikat. **Pertahankan, jangan hapus.**
`<dl>/<dt>/<dd>` dipakai di 2 file (`BeritaAcaraPage:141–180`,
`PaketDetailPage:95–102`) sebagai daftar detail semantik — bukan komponen UI
duplikat; bila Tim Tampilan ingin menyeragamkan, gantinya adalah antd
`Descriptions` (P2, opsional).
Tidak ada `<table>` native (0) — semua tabel memakai antd `Table`.
`lucide-react` dipakai di 33 file untuk **ikon** (bersamaan dengan
`@ant-design/icons`) — ini duplikasi *library ikon*, bukan duplikasi komponen;
dicatat sebagai observasi P2: ke depan, ikon baru sebaiknya dari
`@ant-design/icons` agar satu sumber. **Jangan** mengganti 33 file ikon dalam
tugas migrasi ini (di luar lingkup + risiko visual).

### 12b. Alert rakitan — 34 temuan (P1)

Semua berbentuk `<div className="…border-…-200 bg-…-50…">` atau
`<p className="…text-red-600">` dengan `role="alert"` — ganti dengan antd
`<Alert type="error|warning|success|info" showIcon message={…} />` (pola yang
sudah dipakai 60+× di codebase, termasuk 3× antd `Alert` yang sudah memakai
`role="alert"`: `ImportSoalPanel:645,743`, `ImageUpload:296`).
Keempat file master (Kelas/Mapel/Guru/Siswa/Jurusan) memakai **pola identik**
3× (`role="alert"` di daftar, di modal, di dalam form) — Tim Tampilan dapat
memperbaikinya dengan satu pola yang sama di semua file.

| File | Baris | Bentuk rakitan | Pengganti antd |
|------|-------|----------------|----------------|
| `pages/admin/rekap/RekapPage.tsx` | 144 | `<p role="alert" className="text-sm text-red-600">Gagal memuat daftar jadwal.</p>` | `<Alert type="error" showIcon … />` |
| `pages/admin/rekap/RekapPage.tsx` | 187 | `<p role="alert" className="text-sm text-red-600">…</p>` (error load detail) | `<Alert type="error" showIcon … />` |
| `pages/admin/rekap/BeritaAcaraPage.tsx` | 127 | `<p role="alert" className="text-sm text-red-600 print:hidden">…</p>` | `<Alert type="error" showIcon … />` (pertahankan `print:hidden` via `className`) |
| `pages/admin/settings/SettingsPage.tsx` | 91 | `<p role="alert" className="text-sm text-red-600">…</p>` | `<Alert type="error" showIcon … />` |
| `pages/admin/settings/SettingsPage.tsx` | 176 | `<p role="alert" className="mt-1 text-xs text-red-600">{kktpError}</p>` | `<Alert type="error" showIcon … />` (atau `Form.Item validateStatus="error" help={…}` bila dipindah ke validasi form) |
| `pages/admin/master/KelasPage.tsx` | 159, 164, 182 | 3× `<div className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700" role="alert">` | `<Alert type="error" showIcon … />` |
| `pages/admin/master/MapelPage.tsx` | 171, 176, 194 | 3× `<div … border-red-200 bg-red-50 … text-red-700" role="alert">` | `<Alert type="error" showIcon … />` |
| `pages/admin/master/GuruPage.tsx` | 151, 156, 174 | 3× `<div … border-red-200 bg-red-50 … text-red-700" role="alert">` | `<Alert type="error" showIcon … />` |
| `pages/admin/master/ImportSiswaPage.tsx` | 113 | `<div … border-red-200 bg-red-50 … text-red-700" role="alert">{error}</div>` | `<Alert type="error" showIcon … />` |
| `pages/admin/master/ImportSiswaPage.tsx` | 131 | `<div … border-amber-200 bg-amber-50 … text-amber-800" role="alert">{result.catatan}</div>` | `<Alert type="warning" showIcon … />` |
| `pages/admin/master/SiswaPage.tsx` | 209, 214, 232 | 3× `<div … border-red-200 bg-red-50 … text-red-700" role="alert">` | `<Alert type="error" showIcon … />` |
| `pages/admin/master/JurusanPage.tsx` | 153, 158, 179 | 3× `<div … border-red-200 bg-red-50 … text-red-700" role="alert">` | `<Alert type="error" showIcon … />` |
| `pages/admin/backup/BackupPage.tsx` | 211 | `<p role="alert" className="p-5 text-sm text-red-700">Gagal memuat daftar backup…</p>` | `<Alert type="error" showIcon … />` |
| `pages/admin/analisis/KompetensiPage.tsx` | 121, 164 | 2× `<p role="alert" className="text-sm text-red-600">…</p>` | `<Alert type="error" showIcon … />` |
| `pages/admin/analisis/AnalisisButirPage.tsx` | 120, 170 | 2× `<p role="alert" className="text-sm text-red-600">…</p>` | `<Alert type="error" showIcon … />` |
| `pages/admin/event/EventPage.tsx` | 178, 183, 205 | 3× `<div … border-red-200 bg-red-50 … text-red-700" role="alert">` | `<Alert type="error" showIcon … />` |
| `pages/admin/event/RuanganPage.tsx` | 151, 156, 176 | 3× `<div … border-red-200 bg-red-50 … text-red-700" role="alert">` | `<Alert type="error" showIcon … />` |
| `components/soal/RichEditor.tsx` | 439 | `<p role="alert" className="mt-1 text-xs text-red-600">{errorVideo}</p>` | `<Alert type="error" showIcon … />` (atau `Typography.Text type="danger"`) |

### 12c. Teks error tanpa `role` + kotak info/status rakitan lain (P1/P2)

| File | Baris | Temuan | Saran |
|------|-------|--------|-------|
| `pages/siswa/UjianPage.tsx` | 542–543 | `<CircleX … className="…text-red-600" />` + `<span className="text-red-600">Gagal — coba lagi</span>` (tanpa `role="alert"`) | antd `Alert type="error"`; **jangan ubah logika anti-curang/timer di file ini** — hanya presentasi pesan error (lihat §14). |
| `pages/siswa/HasilPage.tsx` | 63 | `<CircleX … className="…text-red-600" aria-label="Salah" />` (ikon status jawaban) | Ikon status jawaban di halaman hasil — biarkan bila Tim Ujian menyatakan bagian dari rendering soal; bila murni dekorasi, ganti `color` via token. P2. |
| `pages/admin/banksoal/BankSoalFormPage.tsx` | 111 | `<p className="mt-1 text-xs text-red-600">Gagal memuat daftar mapel.</p>` | `Alert` / `Form.Item help`. P1. |
| `pages/admin/paket/PaketFormPage.tsx` | 72 | `<span className="mt-1 block text-xs text-red-600">Gagal memuat daftar mapel.</span>` | Sama. P1. |
| `pages/admin/jadwal/JadwalFormPage.tsx` | 68 | `<span className="mt-1 block text-xs text-red-600">Gagal memuat daftar kelas.</span>` | Sama. P1. |
| `pages/admin/ukk/UkkListPage.tsx` | 137 | `<p className="text-sm text-red-600">Gagal memuat daftar UKK…</p>` | `<Alert type="error" … />`. P1. |
| `pages/admin/ukk/UkkDetailPage.tsx` | 132 | `<p className="text-sm text-red-600">UKK tidak ditemukan…</p>` | `<Alert type="error" … />`. P1. |
| `pages/admin/ukk/UkkNilaiPage.tsx` | 163 | `<p className="text-sm text-red-600">UKK tidak ditemukan.</p>` | `<Alert type="error" … />`. P1. |
| `pages/admin/ukk/UkkSertifikatPage.tsx` | 24 | `<p className="text-sm text-red-600">…</p>` | `<Alert type="error" … />`. P1. |
| `components/soal/RichEditor.tsx` | 407 | `<span className="text-sm text-red-600">Rumus belum valid…</span>` | `Typography.Text type="danger"` / `Alert`. P1. |
| `pages/admin/master/ImportSiswaPage.tsx` | 121, 125 | Kotak hasil import: 2× `<div className="flex-1 rounded-lg border border-green-200 bg-green-50 …">` / `border-red-200 bg-red-50` (Berhasil/Gagal + `Statistic`) | antd `Alert` / `Card` + `Statistic` (Statistic sudah antd). P1. |
| `pages/admin/banksoal/PanduanImportPage.tsx` | 111, 123 | 2× kotak info `<div className="rounded-lg border border-green-200 bg-green-50 p-3">` / `border-red-200 bg-red-50` (tanpa `role`) | antd `Alert type="success|error"`. P1. |
| `pages/admin/backup/BackupPage.tsx` | 120, 145 | 2× `<p className="flex items-start gap-2 rounded-lg bg-red-50 …">` / `bg-amber-50` di dalam `Modal.confirm` (`content`) | antd `Alert` di dalam `content` Modal.confirm. P1. |
| `pages/admin/jadwal/JadwalPage.tsx` | 328 | `<div role="status" className="rounded-lg bg-blue-50 px-3 py-2">` (dialog token) | antd `Alert type="info"`. P1 (sudah punya `role`, tinggal ganti komponen). |
| `pages/admin/rekap/RekapPage.tsx` | 197–199 | Pill status `<span className="rounded-full bg-success-bg … text-green-800">` / `bg-amber-100 … text-amber-800` | antd `Tag` (pakai peta `statusTag.ts` yang sudah ada). P1. |
| `components/soal/SoalEditor.tsx` | 185 | Highlight opsi terpilih `<div … className={… ? "border-anbk bg-blue-50" : "border-slate-200 bg-white"}>` | **P1 tampilan, P0 kehati-hatian**: ini indikator kunci jawaban di editor — ganti class-nya saja, jangan ubah state `kunciOpsi`/`kunciSingle`. |
| `components/soal/RichEditor.tsx` | 352 | Toolbar `<div className="… border-b border-slate-200 bg-blue-50 …">` | Ganti `bg-blue-50` → token/`bg-primary-light`. P1. |
| `pages/admin/banksoal/ImportSoalPanel.tsx` | 834 | `rowClassName={(row) => (barisError.has(row.baris) ? "bg-red-50" : "")}` | `rowClassName` antd Table — ganti class ke token / pertahankan karena ini penanda baris error (P1, opsional). |
| `components/analisis/PengecohTable.tsx` | 61 | `rowClassName={(p) => (p.berfungsi ? "" : "bg-red-50")}` | Sama — penanda pengecoh tidak berfungsi. P1, opsional. |
| `pages/siswa/UjianPage.tsx` | 744 | Bar/kotak `bg-amber-50` (area bawah ujian) | Ganti class ke token; **jangan ubah perilaku tombol/bar bawah** (lihat §14). P1 tampilan. |

### 12d. `<label>` HTML — 18×, observasi P2 (bukan pelanggaran)

`<label htmlFor=…>` dipakai di: `SoalEditor` (12×: label field editor),
`ImportSiswaPage` (1×), `ImportSoalPanel` (2×), `BackupPage` (1×),
`SettingsPage`-adjacent, `GuruPage` (1×), `SiswaHomePage` (1×), `RekapPage` (1× label `sr-only`).
`<label>` adalah elemen semantik HTML, bukan komponen UI — di dalam antd `Form`,
ganti ke `Form.Item label` bila menyentuh file; di luar `Form`, biarkan.
**Bukan temuan wajib.**

### 12e. `<input type="hidden">` — 1×, P2

`SettingsPage.tsx:189`: `<input type="hidden" value={form.remedial_mode ?? "tertinggi"} readOnly />`
— field read-only “Mode remedial = tertinggi”, tidak tampil ke user. Tidak perlu
komponen antd; boleh dihapus sekalian bila nilai ini memang tidak pernah dikirim
(cek: `Form` di file ini memakai `onFinish` — pastikan tidak ada yang membaca
input hidden ini sebelum menghapus).

## 13. Pola yang sudah benar (acuan migrasi — jangan diubah, tiru polanya)

### 13a. Komponen antd yang SUDAH dipakai

| Komponen | Status | Lokasi acuan |
|----------|--------|--------------|
| `DatePicker` | **0 pemakaian** — belum ada satu pun | — (Tim Form membuat pola pertama di `JadwalFormPage`) |
| `TimePicker` | **0 pemakaian** — belum ada satu pun | — |
| `Upload` / `Upload.Dragger` | Dipakai 2 file | `ImportSoalPanel.tsx:707` (`Upload.Dragger`, `beforeUpload` return `false`, `maxCount={1}`, `showUploadList={false}`, file dipilih → pratinjau via fetch manual) — **pola acuan untuk migrasi file**; `ImageUpload.tsx:256,273` (`Upload` + `Upload.Dragger` + `uploadProps`, dipakai `SettingsPage` untuk logo sekolah) |
| `InputNumber` | Dipakai 1× | `SoalEditor.tsx:829` (`<InputNumber id="bobot" min={1} … onChange={(v) => patch({ bobot: … })} style={{ width: "100%" }} />`) — **pola acuan untuk migrasi angka** |
| `Input.TextArea` | Dipakai 7× di 5 file | `SettingsPage.tsx:174` (KKTP, `rows={8}`), `EventPage.tsx:218`, `RuanganPage.tsx:188`, `SoalEditor.tsx:491,511,682`, `exam/SoalRenderer.tsx:408` (jawaban uraian siswa) |
| `Input.Password` | Dipakai 4× | `LoginPage.tsx:108`, `GantiPasswordPage.tsx:98,112,126` |
| `Select` (antd) | Dipakai luas (±40+ JSX) | Semua halaman form; pola: `showSearch`, `allowClear`, `filterOption`, `optionFilterProp="label"`, `notFoundContent="Tidak ada data"`, `size="large"` |
| `Checkbox` / `Radio.Group` | 10× / 2× | Lihat §4 |
| `Modal` + `Modal.confirm` | Dipakai luas | `Modal.confirm` 15× (§9); `<Modal>` JSX di `Kelas/Mapel/Guru/Siswa/JurusanPage`, `BankSoalDetailPage`, `JadwalPage` (token, `width={520}`), `UkkListPage`/`EventPage`/`RuanganPage` (`width={720}`), `SoalEditor:896`, `UjianPage:763` |
| `Alert` (antd) | Dipakai 60+× | Hampir semua halaman; pola: `<Alert type="error|success|warning|info" showIcon message={…} />` |
| `Form` / `Form.Item` | Semua formulir | — |
| `Table`, `Card`, `Button`, `Tag`, `Typography`, `Space`, `Statistic`, `Steps`, `Popconfirm`(tidak dipakai — konfirmasi via `Modal.confirm`) | Dipakai luas | — |
| `ConfigProvider` | 1×, di `App.tsx:84–105` | `locale={idID}`, token: `colorPrimary #1e4e9e`, `colorPrimaryHover #173d7a`, `colorSuccess #15803d`, `colorWarning #b45309`, `colorError #dc2626`, `colorInfo #1e4e9e`, `colorText #0f172a`, `colorTextSecondary #475569`, `colorBorder #cbd5e1`, `colorBgLayout #eef2f7`, `borderRadius 8`, `fontFamily "Plus Jakarta Sans"`, `fontSize 14`, `controlHeightLG 44` — **satu-satunya tempat token didefinisikan; jangan buat palet baru** |

### 13b. `dayjs` — BELUM dependensi langsung (penting untuk Tim Form)

- `frontend/package.json` **tidak** mencantumkan `dayjs` (dependensi: antd, icons,
  fontsource, react-query, tiptap, dompurify, katex, lucide-react, react, router, zustand).
- `dayjs@1.11.23` **ada** di `node_modules` sebagai dependensi transitif antd
  (`antd/package.json: "dayjs": "^1.11.11"`), dan **0** impor `from "dayjs"` di `src`.
- **Rekomendasi:** Tim Form menambahkan `dayjs` sebagai dependensi langsung
  (`npm install dayjs` — tanpa CDN, sesuai aturan proyek) sebelum memakai
  `DatePicker`/`TimePicker`, karena mengimpor paket transitif tanpa mendeklarasikannya
  rapuh (bisa hilang bila antd mengganti dependensinya). Alternatif tanpa `dayjs`:
  tetap memakai string + `format.ts` saja — tapi `DatePicker`/`TimePicker` antd
  **wajib** menerima/mengembalikan objek `Dayjs`, jadi tanpa `dayjs` migrasi
  tanggal/jam tidak bisa dilakukan dengan benar.
- Pola konversi yang wajib dipakai (agar payload API tidak berubah):
  `value={form.tanggal ? dayjs(form.tanggal, "YYYY-MM-DD") : null}` +
  `onChange={(d) => set(..., d ? d.format("YYYY-MM-DD") : "")}` untuk tanggal;
  `value={form.jam_mulai ? dayjs(form.jam_mulai, "HH:mm") : null}` +
  `onChange={(t) => set(..., t ? t.format("HH:mm") : "")}` untuk jam.
  Perlu `import dayjs from "dayjs"` + (bila perlu) `import "dayjs/locale/id"` —
  `ConfigProvider locale={idID}` sudah membuat teks bawaan antd berbahasa Indonesia;
  jangan mengeset locale global dayjs kecuali diperlukan.

### 13c. `src/lib/format.ts` — format tanggal pusat (sudah benar, jangan diubah)

Tiga fungsi ekspor, dipakai di **14 file**:

| Fungsi | Input → Output | Dipakai di |
|--------|---------------|------------|
| `formatTanggal(iso)` | `"2026-10-16"` / ISO → `"16 Okt 2026"`; kosong/invalid → `"—"` | `SiswaHomePage`, `UjianPage`, `KartuSayaPage`, `RekapPage`, `BeritaAcaraPage`, `MonitorPage`, `JadwalPage`, `KompetensiPage`, `AnalisisButirPage`, `EventPage`, `KartuPesertaPage`, `ukk/types.ts` |
| `formatTanggalJam(iso)` | ISO → `"16 Okt 2026, 22:31 WIB"`; tanggal saja → hanya tanggal (tanpa jam karangan) | `BeritaAcaraPage`, `JadwalPage`, `BackupPage`, `NotifikasiBell` |
| `formatJam(iso)` | `"22:31:00"` / `"22:31"` / ISO → `"22:31"`; tanggal saja → `"—"` | `SiswaHomePage`, `UjianPage`, `KartuSayaPage`, `RekapPage`, `BeritaAcaraPage`, `MonitorPage`, `JadwalPage`, `KompetensiPage`, `AnalisisButirPage`, `KartuPesertaPage` |

Aturan yang ditegakkan `format.ts` (wajib dipertahankan Tim Form):
- Semua tanggal **tampil** format Indonesia, zona `Asia/Jakarta`, bulan singkat
  (`Jan, Feb, Mar, Apr, Mei, Jun, Jul, Agu, Sep, Okt, Nov, Des`).
- `"2026-10-16"` (tanggal saja) dipakai apa adanya — **tanpa geser zona waktu**.
- `"2026-10-16T22:31:00"` naif dianggap jam dinding WIB; ber-offset (`Z`/`+07:00`)
  dikonversi ke WIB via `Intl.DateTimeFormat('id-ID', { timeZone: 'Asia/Jakarta' })`.
- Kosong/invalid → `"—"` (bukan string mentah / `"Invalid Date"`).
- **Hanya tampilan**: `format.ts` TIDAK mengubah nilai yang dikirim ke backend.
  Input `type="date"/"time"` tetap memakai format mentah `YYYY-MM-DD` / `HH:MM`
  — setelah migrasi ke `DatePicker`/`TimePicker`, konversi `Dayjs ↔ string mentah`
  (format `"YYYY-MM-DD"` / `"HH:mm"`) harus menghasilkan string mentah yang sama
  persis, agar `formatTanggal`/`formatJam` di halaman tampil tetap bekerja.
- Jangan memformat tanggal tampil dengan `dayjs` di komponen — selalu lewat
  `formatTanggal`/`formatJam`/`formatTanggalJam` (agar tidak ada dua sumber format).

### 13d. Pola benar lain yang dicatat

- `statusTag.ts` (`components/common/statusTag.ts`): satu peta status → warna `Tag`
  (`statusTagColor` / `statusTagLabel`) — Tim Tampilan memakai ini untuk pill status,
  bukan class `bg-…-…` rakitan.
- Warna via `theme.useToken()` (`token.colorPrimary`, `token.colorSuccess`,
  `token.colorWhite`) — pola benar bila warna tidak bisa diwakili `type`/`color` antd.
- `DistribusiChart.tsx`: `fill="var(--color-…)"` — token via CSS `var()`, pola benar
  untuk SVG.
- `SoalEditor.tsx:101` (komentar): mencatat migrasi input/textarea native → antd
  sudah dilakukan di file itu — konsisten dengan hasil grep (0 native di `SoalEditor`).
- `index.css`: **0** selector `.ant-*`, **0** definisi utility Tailwind,
  selector global hanya `body`, `::selection`, `:focus-visible` — tidak ada bentrok
  antd vs CSS kustom. Class `.ds-*` sudah dihapus (0 pemakai).

## 14. File yang SUDAH bersih (tidak perlu disentuh)

### Tier A — bersih dari SEMUA kategori (1–12), tidak perlu disentuh sama sekali

14 file (tidak ada temuan kategori 1–12; yang ada hanya token/jembatan yang sah):

| File | Keterangan |
|------|------------|
| `src/App.tsx` | Token `ConfigProvider` — sumber kebenaran. Jangan diubah. |
| `src/main.tsx` | Entry point, tanpa temuan. |
| `src/components/AdminLayout.tsx` | Header/sidebar antd, warna via token. |
| `src/components/Layout.tsx` | Header antd, warna via token. |
| `src/components/analisis/DayaBedaBadge.tsx` | antd `Tag`, tanpa temuan. |
| `src/components/exam/BannerPeringatan.tsx` | antd `Alert` (2×), tanpa temuan. |
| `src/components/exam/Timer.tsx` | antd, tanpa temuan. |
| `src/components/soal/RichContent.tsx` | Sanitasi + render konten, tanpa temuan UI. |
| `src/pages/ErrorBoundary.tsx` | antd, tanpa temuan. |
| `src/pages/ForbiddenPage.tsx` | antd, tanpa temuan. |
| `src/pages/GantiPasswordPage.tsx` | antd (`Input.Password`, `Alert`, `Button`), tanpa temuan. |
| `src/pages/LoginPage.tsx` | antd, halaman contoh di `ANTD-PANDUAN.md` — tanpa temuan. |
| `src/pages/NotFoundPage.tsx` | antd, tanpa temuan. |
| `src/pages/admin/DashboardPage.tsx` | antd, tanpa temuan. |

File non-UI (`src/lib/*.ts`, `src/store/*.ts`, `src/pages/admin/ukk/types.ts`,
`src/vite-env.d.ts`, `src/components/common/statusTag.ts`) juga bersih dari temuan UI —
tidak perlu disentuh (pengecualian: `statusTag.ts` adalah **acuan** yang dipakai,
bukan yang diperbaiki).

### Tier B — bersih untuk kategori keras (1–9, 12a/12c), hanya sisa class warna P1 (§11)

23 file — **tidak ada** input native / `<button>` / `<form>` / `window.confirm` /
komponen duplikat; yang tersisa hanya class Tailwind warna (§11) dan/atau kotak
tampilan P1 yang dicatat di §12c. Tim Tampilan boleh menyentuh file ini HANYA
untuk §11/§12c, bukan untuk input:

| File | Sisa (hanya §11 / §12c) |
|------|--------------------------|
| `components/analisis/DistribusiChart.tsx` | §11: 2 class (SVG sudah token `var()` — pola benar) |
| `components/analisis/KompetensiHeatmap.tsx` | §11: 1 class (`Tag color=` — pola benar) |
| `components/analisis/PengecohTable.tsx` | §11: 2 + §12c `rowClassName` |
| `components/common/NotifikasiBell.tsx` | §11: 7 |
| `components/exam/NavigasiSoal.tsx` | §11: 1 (warna via `token.colorSuccess` — pola benar) |
| `components/exam/SoalRenderer.tsx` | §11: 14 (⚠️ **file ujian**: hanya ganti class tampilan; JANGAN sentuh `Select` Menjodohkan / jalur grading / `Input.TextArea` uraian) |
| `components/soal/SoalEditor.tsx` | §11: 33 + §12c highlight opsi (jangan ubah state kunci) |
| `components/soal/SoalRenderer.tsx` | §11: 28 (halaman tampil kunci jawaban — ganti class saja) |
| `components/ui/ImageUpload.tsx` | §11: 5 (wrapper `Upload` sah — dipertahankan) |
| `pages/admin/banksoal/BankSoalDetailPage.tsx` | §11: 9 |
| `pages/admin/banksoal/BankSoalFormPage.tsx` | §11: 3 + §12c teks error |
| `pages/admin/banksoal/BankSoalListPage.tsx` | §11: 5 |
| `pages/admin/banksoal/ImportSoalPanel.tsx` | §11: 37 + §12c `rowClassName` (pola `Upload.Dragger` sudah benar — jangan diubah) |
| `pages/admin/banksoal/PanduanImportPage.tsx` | §11: 16 + §12c 2 kotak info |
| `pages/admin/jadwal/JadwalPage.tsx` | §11: 5 + §12c kotak token `role="status"` |
| `pages/admin/monitor/MonitorPage.tsx` | §11: 15 |
| `pages/admin/paket/PaketPage.tsx` | §11: 2 |
| `pages/admin/kartu/KartuPesertaPage.tsx` | §11: 3 |
| `pages/siswa/KartuSayaPage.tsx` | §11: 3 |
| `pages/siswa/SiswaHomePage.tsx` | §11: 7 |
| `pages/siswa/UjianPage.tsx` | §11: 38 + §12c teks error & kotak amber (⚠️ **file ujian**: hanya presentasi; JANGAN sentuh timer, anti-curang, autosave, grading, `window.history.pushState`, fullscreen) |
| `pages/siswa/HasilPage.tsx` | §11: 10 + §12c ikon status (P2) |
| `pages/admin/ukk/UkkSertifikatPage.tsx` | §11: 12 + §12c teks error |

### File yang MEMILIKI temuan keras (harus dikerjakan — 21 file)

`JadwalFormPage`, `UkkListPage`, `EventPage`, `ImportSiswaPage`, `UkkDetailPage`,
`RekapPage`, `SettingsPage`, `SiswaPage`, `PaketFormPage`, `PaketDetailPage`,
`UkkNilaiPage`, `RuanganPage`, `KelasPage`, `MapelPage`, `GuruPage`, `JurusanPage`,
`BackupPage`, `KompetensiPage`, `AnalisisButirPage`, `BeritaAcaraPage`, `RichEditor`.
(Lihat §1, §5, §6, §12b/§12c untuk baris persisnya.)

## 15. Batasan & kejujuran audit

- Audit ini **statis** (grep + baca kode) — tidak menjalankan dev server / build /
  klik browser. Validasi runtime (apakah `DatePicker` ter-render benar, apakah
  payload API tetap sama setelah ganti `InputNumber`) adalah tugas Tim QA.
- Jumlah class Tailwind (§11) dihitung dari pola regex warna Tailwind standar;
  class seperti `bg-white`/`text-white` (47×) dan class jembatan token (83×)
  dihitung terpisah dan **tidak** dimasukkan ke angka 472.
- Untuk kategori 6, audit tidak dapat memastikan dari frontend saja apakah backend
  menerima angka atau string angka untuk tiap field — Tim Form wajib memeriksa
  pemakaian `Number(...)` yang sudah ada / kontrak API sebelum mengubah tipe state
  (`RekapPage`, `SiswaPage`, `UkkNilaiPage`, `RuanganPage` menyimpan string saat ini).
- Untuk `ImportSiswaPage`, audit mencatat label “File CSV *” + `accept=".csv"` —
  bila backend sudah menerima `.xlsx`, itu di luar temuan audit ini (tidak ada
  bukti di kode frontend yang dibaca).
- File `docs/uml/`, `~/workspace/your_files/`, exe, dan zip **tidak disentuh**
  (sesuai aturan keras). Laporan ini adalah satu-satunya file yang ditulis.

## 16. Rekomendasi urutan kerja untuk koordinator

1. **Tim Form — P0 dulu (21 kontrol):** `JadwalFormPage` (3) → `EventPage` (2) →
   `UkkListPage` (1) → `SettingsPage` (4 angka) → `UkkNilaiPage` (4 angka) →
   `RekapPage`/`SiswaPage`/`PaketFormPage`/`PaketDetailPage`/`RuanganPage` (1 masing-masing)
   → `ImportSiswaPage` + `UkkDetailPage` (file). Prasyarat: tambahkan `dayjs`
   sebagai dependensi langsung. Uji: nilai string mentah ke API tidak berubah.
2. **Tim Tampilan — P1:** 34 Alert rakitan (§12b) → teks error/kotak (§12c) →
   472 class warna (§11), dimulai dari file dengan jumlah terbesar
   (`UjianPage`, `ImportSoalPanel`, `SoalEditor`, `SoalRenderer`).
3. **Tim Ujian — presentasi saja:** `UjianPage` (§11: 38 + §12c) dan
   `exam/SoalRenderer` (§11: 14) — ganti class/Alert saja; **dilarang** menyentuh
   grading, timer, anti-curang, autosave, `Select` Menjodohkan, fullscreen,
   `window.history.pushState`.
4. **P2 (bila sempat):** hapus hex basi di komentar, tokenisasi gaya `.rich-content`
   di `index.css`, selaraskan ikon ke `@ant-design/icons` untuk kode baru,
   hapus `<input type="hidden">` di `SettingsPage` bila terbukti tidak terpakai,
   hapus blok `@theme` bila pemakai class jembatan sudah 0.
