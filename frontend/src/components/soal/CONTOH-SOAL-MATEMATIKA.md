# Contoh Soal Matematika — Bukti Rich Editor + Rumus (Tim E)

Dokumen ini adalah **bukti fungsi** editor soal rich + rumus, BUKAN soal
produksi. Tidak ada seed / data produksi yang diubah. Untuk mencoba:
Bank Soal → buka bank → Tambah Soal → salin isi contoh di bawah ke
field yang ditunjuk → klik **Pratinjau soal** di dalam editor (atau
Pratinjau di halaman detail) → rumus harus tampil cantik (KaTeX),
bukan teks `$...$`.

## Cara menulis rumus

- Di editor, klik tombol **Σ** di toolbar → ketik LaTeX **tanpa** tanda
  `$` (contoh: `\frac{1}{2}`) → pilih Inline / Blok → **Sisipkan**.
- Pratinjau rumus langsung muncul di panel Σ sebelum disisipkan.
- Rumus tersimpan sebagai teks LaTeX (`$...$` inline, `$$...$$` blok),
  BUKAN gambar — tetap bisa dicari & dinilai.
- Bila rumus tidak valid, yang tampil adalah teks LaTeX mentahnya
  (halaman tidak kosong / tidak error merah).

## Contoh 1 — Pecahan (PG, tipe `pg`)

- **Pertanyaan** (ketik di editor, sisipkan rumus inline lewat tombol Σ):
  `Hitunglah hasil dari $\\frac{3}{4} + \\frac{1}{2}$!`
- **Opsi**: A = `$\\frac{5}{4}$` · B = `$\\frac{4}{6}$` · C = `$1$` · D = `$\\frac{3}{8}$`
- **Kunci**: A. **Pembahasan**: `Samakan penyebut: $\\frac{3}{4} + \\frac{2}{4} = \\frac{5}{4}$.`
- Harapan render: pecahan tampil bertingkat (pembilang di atas penyebut),
  bukan `3/4`.

## Contoh 2 — Pangkat & Akar (PG Kompleks, tipe `pg_kompleks`)

- **Pertanyaan**: `Manakah pernyataan berikut yang BENAR?`
- **Stimulus** (opsional): `Gunakan sifat $a^{m} \\cdot a^{n} = a^{m+n}$ dan $\\sqrt{a^2} = |a|$.`
- **Opsi**: A = `$2^{3} \\cdot 2^{2} = 2^{5}$` · B = `$\\sqrt{16} = 4$` ·
  C = `$3^{2} = 6$` · D = `$\\sqrt{9} + \\sqrt{16} = 7$`
- **Kunci**: A, B, D. Nilai proporsional otomatis (3 kunci) — tidak berubah
  oleh editor.
- Harapan render: pangkat tampil sebagai superscript, akar tampil dengan
  tanda √.

## Contoh 3 — Integral (Uraian, tipe `uraian`)

- **Pertanyaan** (rumus BLOK lewat tombol Σ → mode "Blok"):
  `Hitunglah integral berikut: $$\\int_{0}^{1} x^2\\,dx$$`
- **Rubrik** (kunci, teks polos — tetap seperti sebelumnya):
  `Benar bila jawaban 1/3 beserta langkah: anti-turunan x^3/3, substitusi 0 dan 1.`
- Harapan render: integral tampil besar di tengah baris sendiri (mode
  display), dengan batas bawah 0 dan batas atas 1.

## Contoh 4 — Campuran format (tabel + list + kode tidak berlaku di sini)

- **Stimulus** (uji tabel + list): buat tabel 2×2 lewat tombol tabel,
  isi dengan `x` / `$x^2$` / `1` / `1`, lalu daftar bullet berisi
  `$x = 1$` dan `$x = 2$`.
- Harapan render: tabel bergaris, list ber-bullet, rumus di dalam sel
  tabel tetap ter-render.

## Verifikasi yang sudah dilakukan (Tim E)

1. `npm run build` (tsc + vite) lolos — lihat laporan Tim E.
2. Bundle dicek: tidak ada URL CDN (tanpa `fonts.googleapis.com`,
   `unpkg.com`, `jsdelivr.net`); font KaTeX di-bundle lokal oleh Vite.
3. Soal lama (teks polos) tetap tampil normal: editor menyimpan teks
   polos apa adanya bila tidak ada format/rumus blok/tabel/gambar/video,
   dan `RichContent` menampilkan teks polos persis seperti sebelumnya.
4. Kunci jawaban tidak berubah bentuknya: PG tetap label (`"A"`),
   PG Kompleks tetap array label, Menjodohkan tetap pasangan
   `{kiri, kanan}` dengan sinkronisasi kunci yang sama seperti fix
   sebelumnya, Isian/Rumpang tetap kunci teks polos.

## Catatan integrasi (di luar cakupan Tim E — perlu Tim halaman siswa)

`components/soal/RichContent.tsx` adalah renderer bersama yang SUDAH
siap dipakai halaman siswa: `UjianPage` + `components/exam/SoalRenderer`
+ `HasilPage` saat ini menampilkan `pertanyaan`/`stimulus`/opsi sebagai
teks polos. Sampai integrasi itu dilakukan:

- Soal yang HANYA teks polos / rumus inline `$...$` tampil normal di
  halaman siswa (rumus inline tampil sebagai teks `$...$` — masih
  terbaca; setelah integrasi tampil KaTeX).
- Soal yang memakai format rich (tebal/list/tabel/gambar/video/rumus
  blok) HARUS dipratinjau guru dulu dan baru dipakai di ujian setelah
  `RichContent` dipakai di halaman siswa — jika tidak, siswa melihat
  source HTML.

Integrasi yang dibutuhkan (satu baris per tempat): ganti
`{soalAktif.pertanyaan}` / `{o.teks}` / `{p}` di halaman siswa menjadi
`<RichContent html={...} />` dengan import dari
`../../components/soal/RichContent`. File halaman siswa TIDAK diubah
oleh Tim E (di luar daftar file Tim E).
