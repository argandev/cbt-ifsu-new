# Pemetaan Class Full Indonesia — CBT SMK v2 (Fase 1 Finalisasi UML)

- **File:** `docs/uml/cbt-smk-v2.mdj`
- **SHA sebelum:** `af99c1ea552cbf6509ea8f04c2d3f386c1768c612748cfdb3523fefbdc20c467`
  (Activity selesai, 19 diagram)
- **SHA sesudah:** `787e89d21846b646994b6f8fe506df5d2ac93486e846a630359a86beebe9d3e1`
- **Backup:** `docs/uml/cbt-smk-v2.mdj.pre-finalisasi.bak`
  (SHA = SHA sebelum; dibuat sekali, tidak ditimpa)
- **Script (idempotent, rerun → SHA identik):** `docs/uml/rename_class_full_indo.py`
- **Validator:** `validate-mdj-real.py` → **0 pelanggaran, LOLOS SEMUA**
- **Prinsip:** nama baru selaras pemetaan PIM yang sudah dipakai ERD /
  Model Data Logis (`PEMETAAN-PIM.md`) agar Class dan ERD menyebut konsep
  yang sama dengan nama yang sama. 18 class sudah Bahasa Indonesia penuh
  dan dipertahankan apa adanya; 16 class diubah.

## 34 class: nama lama → nama baru + alasan

| # | Nama lama | Nama baru | Alasan |
|---|---|---|---|
| 1 | `Attempt` | `PengerjaanUjian` | Padanan ERD `attempt → PengerjaanUjian`; istilah proses ujian. |
| 2 | `AuditLog` | `JejakAudit` | Padanan ERD `audit_log → JejakAudit`; "jejak audit" istilah baku. |
| 3 | `BackupMeta` | `MetaCadangan` | Padanan ERD `backup_meta → MetaCadangan` (metadata cadangan). |
| 4 | `BankSoal` | `BankSoal` (tetap) | Sudah Bahasa Indonesia penuh. |
| 5 | `BeritaAcara` | `BeritaAcara` (tetap) | Sudah Bahasa Indonesia penuh. |
| 6 | `Event` | `Kegiatan` | Padanan ERD `event → Kegiatan`; asosiasi `JadwalUjian.kegiatanId` sudah memakai istilah ini. |
| 7 | `Guru` | `Guru` (tetap) | Sudah Bahasa Indonesia penuh. |
| 8 | `JadwalUjian` | `JadwalUjian` (tetap) | Sudah Bahasa Indonesia penuh. |
| 9 | `Jawaban` | `Jawaban` (tetap) | Sudah Bahasa Indonesia penuh. |
| 10 | `Jurusan` | `Jurusan` (tetap) | Sudah Bahasa Indonesia penuh. |
| 11 | `Kelas` | `Kelas` (tetap) | Sudah Bahasa Indonesia penuh. |
| 12 | `KisiKisi` | `KisiKisi` (tetap) | Sudah Bahasa Indonesia penuh. |
| 13 | `Kompetensi` | `Kompetensi` (tetap) | Sudah Bahasa Indonesia penuh. |
| 14 | `Mapel` | `MataPelajaran` | "Mapel" singkatan; bentuk penuh "Mata Pelajaran" = padanan ERD `mapel → MataPelajaran`; asosiasi sudah `mataPelajaranId`. |
| 15 | `NilaiAkhir` | `NilaiAkhir` (tetap) | Sudah Bahasa Indonesia penuh. |
| 16 | `Notifikasi` | `Notifikasi` (tetap) | Sudah Bahasa Indonesia penuh. |
| 17 | `PaketSoal` | `PaketSoal` (tetap) | Sudah Bahasa Indonesia penuh. |
| 18 | `PaketUjian` | `PaketUjian` (tetap) | Sudah Bahasa Indonesia penuh. |
| 19 | `Pelanggaran` | `Pelanggaran` (tetap) | Sudah Bahasa Indonesia penuh. |
| 20 | `Permission` | `Izin` | Padanan ERD `permissions → Izin`; asosiasi join sudah `PeranIzin`. |
| 21 | `Remedial` | `Remedial` (tetap) | Istilah pendidikan Bahasa Indonesia; dipertahankan, selaras ERD. |
| 22 | `Role` | `Peran` | Padanan ERD `roles → Peran`. |
| 23 | `Ruangan` | `Ruangan` (tetap) | Sudah Bahasa Indonesia penuh. |
| 24 | `Session` | `Sesi` | Padanan ERD `sessions → Sesi`; atribut sudah `tokenSesi`. |
| 25 | `Settings` | `Pengaturan` | Padanan ERD `settings → Pengaturan`. |
| 26 | `Siswa` | `Siswa` (tetap) | Sudah Bahasa Indonesia penuh. |
| 27 | `Soal` | `Soal` (tetap) | Sudah Bahasa Indonesia penuh. |
| 28 | `SodOverride` | `PengecualianPemisahanTugas` | Padanan ERD `sod_overrides → PengecualianPemisahanTugas`; SoD = Separation of Duties = Pemisahan Tugas, override = pengecualian (dengan alasan tertulis). |
| 29 | `UkkBerkas` | `BerkasUkk` | Selaras ERD `ukk_berkas → BerkasUkk` (kata benda dulu, seperti atribut `ujianUkkId`). |
| 30 | `UkkNilai` | `NilaiUkk` | Selaras ERD `ukk_nilai → NilaiUkk`. |
| 31 | `UkkSertifikat` | `SertifikatUkk` | Selaras ERD `ukk_sertifikat → SertifikatUkk`. |
| 32 | `UkkUjian` | `UjianUkk` | Selaras ERD `ukk_ujian → UjianUkk`. |
| 33 | `User` | `Pengguna` | Padanan ERD `users → Pengguna`. |
| 34 | `UsulanUnblock` | `UsulanBukaBlokir` | Padanan ERD `usulan_unblock → UsulanBukaBlokir`; "unblock" → "buka blokir". |

## 16 asosiasi Diagram Class yang ikut diubah namanya

| Nama lama | Nama baru |
|---|---|
| `Attempt.jadwalUjianId` | `PengerjaanUjian.jadwalUjianId` |
| `Attempt.siswaId -> User.id` | `PengerjaanUjian.siswaId -> Pengguna.id` |
| `AuditLog.pelakuId` | `JejakAudit.pelakuId` |
| `BackupMeta.dibuatOleh` | `MetaCadangan.dibuatOleh` |
| `Session.penggunaId` | `Sesi.penggunaId` |
| `SodOverride.pelakuId` | `PengecualianPemisahanTugas.pelakuId` |
| `User.peran -> Role.kode (logis, bukan kunci tamu)` | `Pengguna.peran -> Peran.kode (logis, bukan kunci tamu)` |
| `UkkNilai.siswaId` | `NilaiUkk.siswaId` |
| `UkkSertifikat.siswaId` | `SertifikatUkk.siswaId` |
| `UkkUjian.jurusanId` | `UjianUkk.jurusanId` |
| `UkkUjian.kelasId` | `UjianUkk.kelasId` |
| `UkkUjian.mataPelajaranId` | `UjianUkk.mataPelajaranId` |
| `UkkNilai.ujianUkkId` | `NilaiUkk.ujianUkkId` |
| `UkkBerkas.ujianUkkId` | `BerkasUkk.ujianUkkId` |
| `UkkSertifikat.ujianUkkId` | `SertifikatUkk.ujianUkkId` |
| `UsulanUnblock.pengerjaanUjianId` | `UsulanBukaBlokir.pengerjaanUjianId` |

30 asosiasi lainnya tidak menyebut nama class lama (mis. `PeranIzin (…)`,
`JadwalUjian.kegiatanId`) dan tidak diubah. Nama ujung asosiasi kosong
semua, multiplisitas dan rute garis tidak disentuh.

## Label view + geometri

- Teks label nama class (16 LabelView) dan label nama asosiasi
  (16 EdgeLabelView) di Diagram Class ikut diubah ke nama baru.
- Satu-satunya ClassView yang dilebarkan: `SodOverride` 170 → 240
  (nama baru 26 karakter), posisi tetap (kiri 580, tepi kanan 820,
  tetangga `Notifikasi` di 860 — tanpa tabrakan); compartment + 4 label
  nama ikut 240. Validator: overlap utama >10% = 0 pasangan.
- 15 EdgeLabelView nama asosiasi dilebarkan sesuai rumus label lama
  (`floor(7.5 × panjang teks)`); 1 (`Sesi.penggunaId`, sudah 135) tidak perlu.

## Verifikasi (sebelum → sesudah)

| Ukuran | Sebelum | Sesudah |
|---|---|---|
| UMLClass | 34 | 34 |
| Atribut | 229 | 229 |
| Operasi | 11 | 11 |
| Asosiasi bernama (Diagram Class) | 46 | 46 |
| Total objek JSON | 69.611 | 69.611 |
| Nilai leaf yang berubah (diff vs backup) | — | 81, semuanya field `name` (32) / `text` (32) / `width` (17); tidak ada field lain |
| Validator `validate-mdj-real.py` | 0 pelanggaran | 0 pelanggaran |
| Rerun script | — | "Tidak ada perubahan", SHA identik |

Nama baru unik per parent (cek UML002 di dalam script + validator):
tidak ada duplikat `(parent, nama)`.

## Sisa referensi lama — TIDAK diubah (di luar ranah Fase 1)

1. **Dokumentasi Use Case UC02** (`UMLUseCase` id `24OY1K5OO0INTP75Q6NQIW`):
   "…Jurusan, Kelas, **Mapel**, Guru, Siswa…" — satu-satunya penyebutan
   nama class lama utuh di luar Diagram Class. Ranah diagram Use Case;
   perlu diubah ke `MataPelajaran` pada sentuhan Use Case (bukan Fase 1).
2. **Nama atribut (tidak diubah sesuai perintah Fase 1):**
   `mapelId` (4 atribut Class + 4 kolom ERD), `daftarMapel`
   (1 atribut + 1 kolom ERD). Atribut/operasi dinyatakan sudah Indonesia
   dan tidak diubah pada fase ini.
3. **Nama operasi (tidak diubah sesuai perintah Fase 1):**
   `wajibGantiPassword(): Boolean` pada `Pengguna` masih mengandung kata
   "Password" (catatan: PEMETAAN-PIM.md mencatatnya "(tetap)" pada fase PIM).
4. **Teks desain di Sequence/Activity** (bukan nama class, tidak diubah):
   singkatan `SoD` (mis. "pengecualian SoD", "Tolak pengecualian SoD"),
   kata `Unblock`/`di-unblock` pada pesan antarmuka Activity D-ACT-18
   ("Unblock tidak bisa dilakukan", "Peserta di-unblock…"), dan `UKK`
   sebagai akronim pada pesan-pesan Sequence/Activity.
