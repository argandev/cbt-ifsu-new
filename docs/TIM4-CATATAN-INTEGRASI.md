# Catatan Integrasi — Tim 4 (Mesin Ujian + Monitor)

Dokumen ini mencatat asumsi antar-tim, konflik yang ditemukan, dan
keputusan yang diambil Tim 4. Dibaca koordinator + Tim 1/2/3/5/6.

## 1. Konflik yang WAJIB diputuskan koordinator

### 1.1 Nama tabel: `jadwal_ujian` / `attempt` (Tim 4) vs `jadwal` / `attempts` (Tim 5)

- Spesifikasi tugas Tim 4 menetapkan tabel `jadwal_ujian` dan `attempt`
  (singular). Migrasi `0004_exam.up.sql` mengikuti spesifikasi itu.
- Migrasi `0005_rekap` (Tim 5, sudah ada) mengasumsikan tabel bernama
  `jadwal` dan `attempts` (FK `nilai_akhir -> attempts(id)`, kueri
  `FROM attempts`, `FROM jadwal ... j.mapel_id`).
- **Dampak:** bila 0004 + 0005 dijalankan apa adanya, 0005 GAGAL
  (tabel rujukan tidak ada).
- **Rekomendasi Tim 4:** ubah 0005 (1 berkas: ganti `jadwal` ->
  `jadwal_ujian`, `attempts` -> `attempt`; perhatikan juga
  `j.mapel_id` di kueri berita-acara Tim 5 — `jadwal_ujian` tidak punya
  `mapel_id` langsung; mapel didapat via `JOIN paket_ujian`).
  Alternatif: Tim 4 mengganti nama di 0004 (menyimpang dari spesifikasi
  tugas Tim 4). Keputusan di tangan koordinator — jangan ubah diam-diam.

### 1.2 Echo v5: handler Tim 5 memakai signature Echo v4

- `go.mod` memakai Echo v5 (`HandlerFunc func(c *Context) error`,
  `*echo.Context`). Berkas Tim 5 (`audit.go`, `rekap.go`,
  `settings.go`) menulis handler `func(c echo.Context) error`
  (signature Echo v4) — **tidak compile** di Echo v5
  (`go vet` gagal di `audit.go:120`).
- **Yang dilakukan Tim 4:** paket `exam` dan `monitor` memakai
  `*echo.Context` (benar untuk v5) dan TIDAK mengimpor paket `audit`
  (membaca `auth_user_id`/`auth_role`/`auth_siswa_id` langsung dari
  context + INSERT audit_log sendiri). Bila Tim 5 memperbaiki signature,
  Tim 4 boleh beralih ke `audit.Catat`/`audit.CurrentUser`.

### 1.3 Seed (Tim 6) vs matriks ANALISIS Bab 11

| Hal | Seed Tim 6 | Matriks ANALISIS Bab 11 + tugas Tim 4 | Yang ditegakkan Tim 4 |
|---|---|---|---|
| Proktor kelola paket | proktor dapat `paket.*` | Proktor ❌ (tidak boleh lihat isi soal) | Matriks: proktor DITOLAK di endpoint paket |
| Pengawas usulan-unblock | pengawas TIDAK dapat `jadwal.usulan-unblock` | Pengawas ⚠️ hanya mengusulkan | Matriks: pengawas BOLEH mengusulkan |

### 1.4 `settings.value`: jsonb (Tim settings) vs TEXT (asumsi seed)

Tim settings memakai `value jsonb`; seed mengasumsikan `value TEXT`.
Kode Tim 4 membaca setting secara toleran (coba parse JSON dulu, lalu
text) sehingga jalan di kedua bentuk. Koordinator tetap perlu
menyeragamkan skema `settings` di migrasi 0001 (Tim 1).

## 2. Kontrak yang diasumsikan Tim 4 (tim lain wajib penuhi)

### 2.1 Tabel `soal` (Tim 3, bank soal)

Model soal Tim 3 yang dipakai (sumber: `frontend/src/lib/banksoal.ts`
+ `lib/tipeSoal.ts` — backend Tim 3 belum menulis skema saat Tim 4
bekerja, jadi kontrak ini diambil dari frontend Tim 3):

- `soal(id, bank_id, tipe, stimulus, pertanyaan, media_url,
  jawaban jsonb, kunci jsonb, bobot, level, status)`.
- `jawaban` = data TAMPIL ke peserta (BUKAN jawaban siswa):
  `{opsi:[{label,teks}]}`, `{pernyataan:[...]}`,
  `{pasangan:[{kiri,kanan}]}`, `{kolom:[...]}`, `{item:[...]}`,
  `{skala:[...]}`.
- `kunci` = guru saja: `{"opsi":"B"}` / `{"opsi":["A","C"]}` /
  `{"nilai":[bool,...]}` / `{"pasangan":[{kiri,kanan}]}` /
  `{"jawaban":[...]}` / `{"urutan":[indeks,...]}` / `{"rubrik":"..."}`.
- `bank_soal(id, mapel_id, ...)` — mapel soal didapat via
  `bank_soal`, BUKAN kolom `soal.mapel_id` (tidak ada).
- `soal.status` ∈ draft/aktif/nonaktif/perlu_revisi/diarsipkan;
  generate-acak hanya mengambil `status='aktif'`.
- Kode tipe Tim 3: `pg, pg_kompleks, menjodohkan, isian, uraian,
  benar_salah, pg_multi, urutkan, isi_ganda, teks_rumpang, peta_konsep,
  studi_kasus, pernyataan_sikap, upload_file, praktik_singkat,
  pilihan_gambar`. Grading Tim 4 (`grading.go`) menangani SEMUA kode
  ini + alias lama (`isian_singkat`, `mengurutkan`, `skala_sikap`, ...).
- **Tidak ada kolom `pembahasan`** di model Soal Tim 3. Halaman hasil
  Tim 4 menampilkan "pembahasan" sebagai kunci benar yang dirangkum
  (`kunciTampil`) untuk Tryout / tampil_hasil.
- `components/soal/SoalRenderer.tsx` (Tim 3) adalah pratinjau READ-ONLY
  untuk guru ("JANGAN dipakai di halaman siswa"). Tim 4 membuat
  renderer interaktif sendiri di `components/exam/SoalRenderer.tsx`.
- Endpoint bank soal Tim 3 yang dipakai PaketPage:
  `GET /api/v1/bank-soal`, `GET /api/v1/bank-soal/:id/soal`.
- Stub `register.go` Tim 1 di `internal/exam/` dan
  `internal/monitor/` (yang tertulis "Tim Ujian/Monitor mengganti
  berkas ini") SUDAH dihapus Tim 4 karena implementasi penuh ada di
  `exam.go` / `monitor.go` dengan signature Register yang sama.

### 2.2 Tabel `siswa`, `kelas`, `mapel` (Tim 2)

- `siswa(id, nisn, nama, kelas_id)` — dipakai monitor + koreksi +
  penentuan kelas siswa di `ujian/aktif`. Fallback: `siswa.user_id`.
- `kelas(id, nama)`, `mapel(id, nama)`.
- `attempt.siswa_id` menyimpan nilai `auth_siswa_id` dari context auth
  Tim 1. Bila `siswa.id != users.id`, Tim 1 WAJIB mengisi context
  `auth_siswa_id` dengan `siswa.id` yang benar (bukan `users.id`).

### 2.3 Auth context (Tim 1)

Middleware Tim 1 mengisi: `auth_user_id`, `auth_role`,
`auth_siswa_id`, `auth_permissions`. Role yang dipakai Tim 4:
`admin`, `proktor`, `guru`, `pengawas`, `siswa`, `penguji_ukk`.
CSRF ditegakkan middleware Tim 1 (semua POST/PUT/DELETE Tim 4
mengandalkan itu — handler tidak mengecek CSRF sendiri).

## 3. Keputusan desain Tim 4 (di luar teks tugas, dicatat eksplisit)

1. **Jadwal `draft` -> `aktif`** saat peserta pertama memulai ujian
   yang valid (membaca "jadwal aktif/draft→aktif" di tugas).
2. **`POST /ujian/mulai` idempotent:** bila sudah ada attempt berjalan,
   kembalikan attempt itu (bukan 409) agar refresh/reconnect tidak
   mengunci siswa. Attempt `selesai` -> 409; `diblokir` -> 403.
3. **Token untuk Tryout:** dilewati sepenuhnya (tugas: "Tryout boleh
   tanpa token").
4. **Nilai tampil ke siswa** = `paket.tampil_hasil` AND
   `settings.tampil_hasil_ke_siswa` (keduanya harus true).
   `nilai_akhir` tetap disimpan di DB untuk guru/proktor walau
   disembunyikan dari siswa.
5. **Deduplikasi pelanggaran:** event APAPUN dalam 5 detik setelah
   pelanggaran terakhir diabaikan (bukan hanya jenis yang sama) —
   karena `blur` + `visibilitychange` berapi hampir bersamaan.
6. **Attempt diblokir membekukan sisa waktu** (`sisa_waktu_beku`)
   sehingga unblock bisa lanjut dari sisa yang sama persis.
7. **Soal tanpa baris jawaban** dihitung kosong (poin 0) dalam
   `nilai_akhir` agar penyebut Σ bobot lengkap.
8. **Endpoint tambahan** (di luar daftar tugas, untuk kebutuhan UI):
   `GET /api/v1/jadwal/:id/usulan-unblock` (daftar usulan untuk
   MonitorPage) dan `GET /api/v1/ujian/attempt/:id/hasil` (rekap
   per-soal HasilPage: nilai + pembahasan bila Tryout/tampil_hasil).
   Semua endpoint tugas tetap ada.
9. **`DELETE /api/v1/paket/:id/soal`** menerima `soal_id` dari body
   JSON atau query `?soal_id=` (DELETE ber-body tidak selalu didukung
   client).
10. **Pengawas tidak boleh melihat token** — termasuk endpoint
    `token-status` (hanya admin/proktor).
11. **`jumlah_soal` paket** disinkronkan dengan jumlah susunan aktual
    setiap tambah/hapus/generate-acak soal.

## 4. Yang BELUM dikerjakan Tim 4 (di luar tugas / butuh tim lain)

- Routing shell (`src/main.tsx`, `src/App.tsx`) milik Tim 1 — belum
  ada. Halaman Tim 4 siap dipasang di route:
  `/ujian/:attemptId` (UjianPage), `/hasil/:attemptId` (HasilPage),
  `/monitor/:jadwalId` (MonitorPage), `/admin/paket` (PaketPage),
  `/admin/jadwal` (JadwalPage). Halaman membaca ID dari
  `window.location` sehingga jalan di bawah router apa pun.
- Komponen UI Tim 1 (`components/ui`) TIDAK dipakai — halaman Tim 4
  memakai elemen HTML + Tailwind langsung agar tidak bergantung pada
  berkas yang belum ada.
- Daftar soal untuk PaketPage memakai `GET /api/v1/bank-soal` +
  `/api/v1/bank-soal/:id/soal` (Tim 3) bila tersedia; bila belum,
  input `soal_id` manual tetap jalan.
- Uji end-to-end dengan DB PostgreSQL nyata belum dilakukan (lihat
  laporan Tim 4) — perlu migrasi 0001–0003 dari Tim 1/2/3 dulu.
