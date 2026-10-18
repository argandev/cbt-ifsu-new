# QA Live-App — CBT SMK v2 (18 Okt 2026)

Verifikasi klik + isi + simpan end-to-end di **aplikasi yang benar-benar jalan**
(backend Go + frontend hasil `npm run build`, disinkronkan ke `backend/dist`
seperti `make build`), **bukan** mock / fetch HTML mentah / build saja.

- Server: binary Go di port `:9991` (bukan 9988/9989), `DATABASE_URL` →
  database QA terpisah **`cbt_smk_v2_qa`** (data uji saja, sudah di-drop
  setelah selesai — tidak ada data nyata yang diubah/dihapus).
- Browser: Chromium headless sungguhan via Playwright-core (CDP), zona waktu
  `Asia/Jakarta`, klik + isi + submit lewat event DOM nyata, request jaringan
  nyata ke backend live. Tiap nilai yang disimpan diverifikasi ganda:
  **reload halaman + cek via API/DB backend**.
- Screenshots: `screenshots/qa-live/` (55 file, desktop 1280 + HP 375).
- Akun uji: admin `qaadmin`, siswa NISN `9900000001` (Siswa QA, kelas
  XI RPL 1), mapel Pemrograman Web (PW), bank soal 2 soal (1 PG + 1 Menjodohkan).

## Hasil: 15 LOLOS, 1 GAGAL

| # | Formulir | Status | Bukti (nilai terverifikasi via API + screenshot) |
|---|----------|--------|--------------------------------------------------|
| 1 | Login admin (UI) | LOLOS | Login via UI → mendarat `/adm` (`00-login-admin.png`) |
| 2 | SettingsPage — 4× InputNumber | LOLOS | token_expiry=1800 (dikosongkan lalu diisi lagi), toleransi=30, max_pelanggaran=3, durasi=121, tampil_hasil=true (`01`–`04`) |
| 3 | RuanganPage — kapasitas | LOLOS | Buat 40 → ubah (kosongkan+isi) 36 (`05`–`08`) |
| 4 | EventPage — 2× DatePicker | LOLOS | tanggal_mulai=2026-10-18, tanggal_selesai=2026-10-25 (`09`–`12`) |
| 5 | SiswaPage — tahun masuk | LOLOS | Buat 2024 → ubah (kosongkan+isi) 2025 (`13`–`16`) |
| 6 | PaketFormPage — durasi | LOLOS | Buat 75 (Tryout, tampil_hasil=true) → ubah 80 (`17`–`20`) |
| 7 | PaketDetailPage — jumlah soal acak | LOLOS | Generate acak jumlah=2 → 2 soal tersusun + pesan UI (`21`–`23`) |
| 8 | JadwalFormPage — DatePicker + 2× TimePicker | LOLOS | Buat tanggal=2026-10-18, jam 00:00–23:59, status=aktif; ubah jam selesai=23:30 (`24`–`29`) |
| 9 | Dialog Token JadwalPage (spot-check) | LOLOS | Dialog dibuka, Generate Token tampil, Salin → "Tersalin!". Token **tidak** dipakai untuk ujian (`30`–`32`) |
| 10 | Modal.confirm BackupPage (spot-check) | LOLOS | Buat Backup (non-destruktif) → dialog Konfirmasi restore tampil (desktop+HP) → **BATAL** diklik; tidak ada restore/hapus dieksekusi (`33`–`35`) |
| 11 | UkkListPage — DatePicker | LOLOS | Buat UKK tanggal=2026-10-20 (`36`–`38`) |
| 12 | UkkDetailPage — Upload berkas | LOLOS | Upload `berkas-qa.pdf` untuk Siswa QA → 1 berkas terdaftar (`39`–`41`) |
| 13 | UkkNilaiPage — 4× InputNumber | LOLOS | teori=80, proses=80, produk=90, sikap=88 → server: praktik=**86.6**, akhir=**83.96**, lulus=true, predikat=**Baik** (`42`–`44`) |
| 14 | **ImportSiswaPage — Upload CSV** | **GAGAL** | **Bug aplikasi, lihat § Bug** (`45`, `45b`) |
| 15 | UjianPage siswa — Menjodohkan (antd Select) + PG + Submit | LOLOS | Menjodohkan via klik Select (Jakarta→Indonesia, Tokyo→Jepang), PG=JavaScript, submit OK → HasilPage **100.00**, rekap backend nilai=**100** (`48`–`52`) |
| 16 | RekapPage — remedial | LOLOS | Remedial 60 (< 100) disimpan; nilai akhir tetap **100** (aturan tertinggi) (`53`–`55`) |

## Bug: ImportSiswaPage — GAGAL di aplikasi jalan (lolos build)

**Gejala (aplikasi jalan):** pilih file CSV uji → klik "Upload & Import" →
Alert merah: **"Body multipart tidak valid (maks. 5 MB)."** Tidak ada siswa
yang terimport lewat UI. Bukti: `screenshots/qa-live/45b-importsiswa-error-body-multipart.png`.

**Akar masalah (terbukti, bukan tebakan):**
`ImportSiswaPage.tsx` mengunggah via `apiClient.post(..., FormData)`, tetapi
`frontend/src/lib/api.ts` (`request()`) memaksa header
`"Content-Type": "application/json"` untuk **semua** body, termasuk FormData —
sehingga browser tidak menulis boundary multipart dan backend menolak body.

Reproduksi di level HTTP (server live yang sama):
- FormData + `Content-Type: application/json` (persis yang dikirim apiClient)
  → `400 {"error":{"code":"validation_error","message":"Body multipart tidak valid (maks. 5 MB)."}}`
- FormData + multipart benar (tanpa Content-Type manual)
  → `200 {"data":{"sukses":1,"gagal":0,...}}` dan siswa terbentuk.

**Perbaikan yang disarankan (TIDAK diterapkan — di luar wewenang
presentasi-only):** di `apiClient.request`, jangan set header `Content-Type`
JSON bila `body instanceof FormData` (biarkan browser menulis
`multipart/form-data; boundary=...`). Satu-satunya pemanggil FormData lewat
apiClient adalah ImportSiswaPage; Upload lain (ImportSoalPanel, UkkDetailPage,
ImageUpload) memakai `fetch` langsung sehingga tidak terkena bug ini —
itulah kenapa upload berkas UKK (§12) LOLOS sementara import siswa GAGAL.

**Daftar eksplisit "hanya lolos build tapi GAGAL di aplikasi jalan":**
`ImportSiswaPage` (Upload CSV) — satu-satunya. Semua formulir lain LOLOS
di aplikasi jalan.

## Catatan jujur / batasan

- **Klik manusia di HP fisik belum terverifikasi.** Semua interaksi memakai
  otomasi Chromium headless (viewport 375 untuk HP). Klik manusia di browser
  sungguhan + HP fisik tetap disarankan sebagai spot-check 1×.
- **Ronde 1 sempat mencatat 4 GAGAL; 3 di antaranya ternyata kesalahan
  skrip/setelan uji, bukan bug aplikasi** (didokumentasikan agar tidak
  terulang): (a) tombol Restore Backup berada di luar scroll horizontal
  tabel + antd merender judul dialog 2× (div tersembunyi + span terlihat)
  sehingga `getByText` gagal strict-mode; (b) ujian pertama ditolak backend
  `403 "Batas toleransi keterlambatan habis"` — **perilaku benar sesuai
  desain** (mulai ujian hanya dalam `jam_mulai + toleransi_telat_menit`);
  jadwal uji awal `jam_mulai 00:00` sudah lewat toleransi 30 menit saat
  ujian dijalankan ~02:0x WIB — diperbaiki dengan mengubah `jam_mulai`
  lewat form Ubah Jadwal yang sama menjadi 02:05, lalu ujian LOLOS;
  (c) Rekap gagal di ronde 1 semata karena ujian belum berjalan.
- Nilai token ujian tidak ditulis di laporan mana pun.
- Cleanup: server dimatikan, database `cbt_smk_v2_qa` di-drop, skrip uji
  (di `/tmp/qa-live`, bukan di repo) dihapus. Satu-satunya perubahan di repo
  akibat ronde ini: rebuild `frontend/dist` + `backend/dist` (setara
  `make build`) dan laporan ini + folder `screenshots/qa-live/`. Tidak ada
  perubahan logika bisnis / grading / auth / token / timer / anti-curang /
  autosave / payload API / endpoint backend.
