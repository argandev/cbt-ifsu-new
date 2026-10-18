// Package analisis — Analisis Butir Soal + Nilai per Kompetensi (Tim Fase 2).
//
// Dihitung ON-DEMAND dari tabel yang sudah ada (jadwal_ujian, attempt,
// jawaban, soal, kompetensi, siswa) — TANPA tabel cache.
//
// Definisi FINAL (spesifikasi tugas Fase 2, mengikat; beda ambang dengan
// ANALISIS.md Bab 8.4 yang lebih longgar — yang dipakai adalah definisi
// di bawah ini):
//
//   - Populasi: attempt yang SUDAH selesai pada satu jadwal
//     (attempt.status = 'selesai', atau punya baris nilai_akhir final).
//     Attempt 'berjalan' / 'diblokir' TIDAK dihitung.
//   - Tingkat kesukaran P = jumlah benar / jumlah penjawab.
//     Label: P < 0.3 Sukar; 0.3–0.7 Sedang; P > 0.7 Mudah.
//   - Daya beda D = (benar kelompok atas − benar kelompok bawah) / n,
//     n = ukuran tiap kelompok (kelompok sama besar). Kelompok atas/bawah
//     = peserta ber-nilai tertinggi/terendah pada jadwal itu.
//     N >= 10: n = round(27% × N) (metode 27%).
//     N < 10: n = floor(N/2) (median split 50%) + catatan "sampel kecil".
//     Interpretasi: D >= 0.40 Sangat Baik; 0.30–0.39 Baik; 0.20–0.29
//     Cukup; 0.00–0.19 Buruk; D < 0.00 Sangat Buruk (soal bermasalah).
//   - Pengecoh: per opsi jawaban, jumlah pemilih + persentase dari
//     penjawab soal itu. Pengecoh "tidak berfungsi" bila dipilih < 5%.
//     Hanya untuk soal berbasis opsi (PG / PG Kompleks / dsb.); untuk
//     soal lain pengecoh = N/A.
//   - Nilai per kompetensi (per siswa, per jadwal):
//     nilai = Σ poin / Σ bobot × 100 atas soal kompetensi itu, dengan:
//   - soal tidak dijawab = poin 0, bobot TETAP masuk penyebut
//     (sama seperti nilai akhir di grading.HitungNilaiAkhir);
//   - soal manual yang SUDAH dijawab tapi BELUM dikoreksi = "belum
//     dinilai" (ANALISIS Bab 8.2): poin tidak masuk pembilang,
//     bobot tidak masuk penyebut;
//   - siswa tidak menjawab satu pun soal kompetensi itu ->
//     nilai = null + flag "tidak ada soal" (BUKAN 0).
//
// "Benar" untuk satu jawaban:
//   - Soal otomatis: poin == bobot (poin penuh). Berlaku juga untuk
//     soal proporsional (PG Kompleks, Menjodohkan, dsb.): hanya poin
//     penuh yang dihitung benar.
//   - Soal manual (uraian / upload_file / praktik_singkat): poin >= 0.6 ×
//     bobot (ambang lulus per soal).
//
// Urutan nilai untuk pemeringkatan kelompok: nilai_akhir.nilai bila ada,
// bila tidak ada dihitung dari jawaban.poin (Σ poin / Σ bobot paket).
//
// Endpoint (Register):
//
//	GET /api/v1/analisis/:jadwalId            (permission analisis.read)
//	GET /api/v1/analisis/:jadwalId/export.csv (permission analisis.export)
//	GET /api/v1/kompetensi/:jadwalId          (permission rekap.read)
//	GET /api/v1/kompetensi/:jadwalId/export.csv (permission rekap.export)
//
// Nilai per kompetensi memakai permission rekap.* (sesuai spesifikasi
// tugas) karena ia adalah laporan rekap, bukan analisis butir.
//
// Struktur layer (hasil refactor arsitektur):
//   - analisis.go:         dokumentasi paket
//   - handler.go:          Echo handler + Register + helper auth
//   - handler_analisis.go: handler analisis butir soal + CSV
//   - handler_kompetensi.go: handler nilai per kompetensi + CSV
//   - service.go:          fungsi murni hitung + logika bisnis + audit
//   - repository.go:       pemuatan data dari DB
//   - platform:            envelope JSON (menggantikan helper lokal)
//   - grading:            dipakai fungsi murni (Manual, NormalisasiTipe)
//
// Perilaku (route, status, bentuk JSON, pesan error) IDENTIK dengan
// sebelum refactor.
package analisis
