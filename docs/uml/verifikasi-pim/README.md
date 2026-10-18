# Verifikasi Visual PIM — CBT SMK v2 (Worker 4)

Bukti visual untuk file `docs/uml/cbt-smk-v2.mdj` final (SHA-256
`c5bd08be78b5e8b4d739231352b9bf6b5cc834c9627070682f2db7a9f737a016`).

## Cara verifikasi (jujur)

1. **StarUML asli 7.1.0** dibuka (Xvfb `:99`, file salinan di `/tmp` —
   file asli tidak pernah dibuka-simpan dari GUI). Status bar menunjukkan
   **centang hijau ✅ = 0 error validasi StarUML**. File dikenali 7 diagram.
2. **Validator independen** `docs/uml/validate-mdj-real.py` → **0 pelanggaran,
   LOLOS SEMUA** (kontrol negatif pada salinan mutasi → GAGAL dengan benar).
3. **Pindai PSM** (auth.go, exam.go, Argon2id, `/api/v1/`, SQL, password_hash,
   csrf_token, JSONB, PostgreSQL, VARCHAR, TIMESTAMPTZ, UUID, dsb.) → **0 temuan
   asli** (hit `403`/`JWT` hanya di dalam angka float/ID acak, bukan konten).
4. **Render visual**: ketujuh diagram di-render lewat pipeline ekspor StarUML
   sendiri (`staruml image -f svg`, 1 diagram per perintah), lalu SVG → PNG 2×
   via cairosvg dengan latar putih. PNG di folder ini adalah hasil render itu.

## Batasan jujur (baca sebelum mengklaim "screenshot live canvas")

- **Kanvas live di GUI StarUML TIDAK melukis diagram** di lingkungan Xvfb ini
  (panel kanvas hanya menampilkan logo placeholder; diukur: ~160/48000 piksel
  berwarna). Ini keterbatasan lingkungan yang sama dengan temuan lama
  "CLI `staruml image` memberi background garbage". Karena itu **tidak ada
  screenshot live canvas yang bermuatan diagram** — bukti visual adalah render
  ekspor StarUML di atas, bukan tangkapan kanvas GUI.
- Klik tetikus X11 tidak sampai ke area konten Chromium StarUML di sandbox ini
  (menu native bisa diklik; konten tidak). Navigasi diagram di GUI dilakukan
  lewat konsol DevTools (`app.diagrams.openDiagram(...)`), yang terbukti jalan.
- Semua PNG/SVG mengandung watermark **UNREGISTERED** (trial StarUML) —
  watermark berasal dari StarUML, bukan ditambahkan/dihilangkan oleh asisten.
  File `.mdj` bersih dari watermark.
- PNG dibuat dari SVG ekspor StarUML via cairosvg (metode yang sama dengan
  ronde sebelumnya), bukan dari ekspor PNG GUI/CLI langsung (yang memberi
  background kotor di Xvfb).

## Hasil per diagram (semua LAYAK)

| # | Diagram | Hasil |
|---|---------|-------|
| 01 | Use Case | LAYAK. 6 aktor + 19 elips UC01–UC19 (teks muat), «include»/«extend» tampil, 2 Note BATASAN utuh. Catatan kosmetik: elips UC17 menyentuh tepi kanan kanvas (teks tetap utuh). |
| 02 | Class Domain | LAYAK. Semua kelas + atribut tipe generik (String/Integer/Boolean/Date/Time/Decimal), operasi Indonesia, nama asosiasi + multiplisitas tampil. |
| 03 | Model Data Logis (ERD) | LAYAK. 35 entitas nama PIM, tipe generik, flag PK/FK/N/U, crow's foot di sisi FK. Klaster `Pengguna` TIDAK disentuh (keputusan A/B menunggu user). |
| 04 | Activity Pelaksanaan Ujian (2a) | LAYAK. Aksi `Mulai Ujian` 4 baris muat di kotaknya; guard `[tidak valid]`/`[valid]`/`[ke-1]`/`[ke-2]`/dll. tampil, tidak terpotong/tumpang tindih. |
| 05 | Activity Koreksi Esai (2b) | LAYAK. Alur linear bersih; guard `[belum semua]`/`[semua terkoreksi]` tampil penuh. |
| 06 | Sequence Login (3a) | LAYAK setelah perbaikan Worker 4: celah bekas lifeline ditutup (pitch seragam 165), label self-message `verifikasiKataSandi()` tidak lagi terpotong tepi (distance 158→10). Note penjelasan self-message utuh. |
| 07 | Sequence Simpan Jawaban (3b) | LAYAK setelah perbaikan Worker 4: celah bekas lifeline ditutup (pitch seragam 165). Self-message `periksaBatasWaktu()` + Note utuh. |

## Perubahan Worker 4 (dicatat juga di PEMETAAN-PIM.md, Addendum)

- Penutupan celah Sequence: 3a Basis Data 760→595; 3b Penghitung Nilai 760→595,
  Basis Data 925→760, Layanan Pemantau 1090→925 + 14 `points` pesan disesuaikan
  (diff total 18 daun; skrip idempotent; backup pra-perubahan:
  `cbt-smk-v2.mdj.pre-sequence-opsi-b.bak` di direktori kerja Worker 4).
- Label self-message 3a `verifikasiKataSandi()`: `distance` 158→10 (mengikuti
  preseden perbaikan 3b) agar tidak terpotong tepi kanan setelah kanvas menyempit.
- Verifikasi ulang penuh sesudah perubahan: validator 0 pelanggaran, pindai PSM 0,
  re-ekspor + inspeksi visual 3a/3b.
