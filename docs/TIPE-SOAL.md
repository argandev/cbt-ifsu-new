# 16 Tipe Soal — CBT SMK v2 (Tim 3: Bank Soal)

> **Label JUJUR — baca dulu.**
> CBT SMK v2 mendukung **5 tipe standar AKM + 11 varian/ekstensi aplikasi**.
> Kami **TIDAK** mengklaim "16 tipe standar Kemendikbud": klaim tersebut
> tidak ditemukan di dokumen resmi Kemendikbud. Dokumen resmi paling
> otoritatif, *Framework AKM* (Pusat Asesmen Pendidikan, BSKAP,
> Kemendikbudristek, edisi revisi 2023, Lampiran 1 hlm. 147–149),
> menetapkan **hanya 5 bentuk soal**: Pilihan Ganda, Pilihan Ganda
> Kompleks, Menjodohkan, Isian/Jawaban Singkat, Uraian/Esai. Label ini
> wajib tampil di UI (badge `[Inti AKM]` / `[Varian aplikasi]`) dan di
> dokumentasi — jangan pernah menulis "16 tipe standar Kemendikbud"
> di dokumentasi/marketing v2.
>
> Sumber kebenaran kode: `TipeSoal` di
> `backend/internal/banksoal/types.go` dan cerminannya di
> `frontend/src/lib/tipeSoal.ts`. Bila menambah/mengubah tipe, ubah
> KETIGA tempat (backend, frontend, dokumen ini).

## Aturan umum penilaian

- Nilai akhir = Σ poin semua soal / Σ bobot semua soal × 100
  (proporsional bobot). Nilai **hanya** dihitung backend — frontend
  TIDAK PERNAH menghitung nilai.
- Soal kosong (tidak dijawab) = skor 0, dicatat sebagai "kosong"
  (dibedakan dari "salah" di analisis butir soal).
- Tanda ragu-ragu tidak memengaruhi nilai.
- Tidak ada penalti jawaban salah (default), kecuali sekolah meminta
  eksplisit + diumumkan ke siswa.
- `kunci` (kunci jawaban) **TIDAK PERNAH** dikirim ke frontend siswa.
  Endpoint siswa (Tim 4) wajib menghilangkannya. Endpoint guru BOLEH
  menampilkan kunci. Proktor TIDAK boleh melihat isi soal sama sekali.

## 5 tipe INTI AKM

| # | Kode | Label UI | Cara jawab | Aturan nilai |
|---|------|----------|------------|--------------|
| 1 | `pg` | Pilihan Ganda `[Inti AKM]` | Pilih 1 dari opsi A–E | Otomatis 1/0 × bobot. Benar = bobot penuh; salah/kosong = 0. |
| 2 | `pg_kompleks` | PG Kompleks `[Inti AKM]` | Pilih semua opsi yang benar (checkbox) | Otomatis **PROPORSIONAL** = benar_dipilih / total_kunci × bobot, **tanpa penalti** jawaban salah. |
| 3 | `menjodohkan` | Menjodohkan `[Inti AKM]` | Pasangkan premis (kiri) → respons (kanan) | Otomatis **PROPORSIONAL** per pasangan benar = pasangan_benar / total_pasangan × bobot. |
| 4 | `isian` | Isian / Jawaban Singkat `[Inti AKM]` | Ketik jawaban singkat | Otomatis exact-match setelah normalisasi: trim + lowercase + rapatkan spasi berlebih. Daftar jawaban alternatif didukung (`kunci.jawaban` = array). |
| 5 | `uraian` | Uraian / Esai `[Inti AKM]` | Ketik jawaban bebas terstruktur | **MANUAL** 0–100 × bobot, dikoreksi guru (rubrik disarankan). Nilai disembunyikan sampai final. |

## 11 tipe VARIAN / EKSTENSI APLIKASI

| # | Kode | Label UI | Cara jawab | Aturan nilai |
|---|------|----------|------------|--------------|
| 6 | `benar_salah` | Benar-Salah `[Varian aplikasi]` | Benar/Salah per pernyataan | Otomatis **PROPORSIONAL** per pernyataan benar = pernyataan_benar / total_pernyataan × bobot. |
| 7 | `pg_multi` | PG Multi-Jawaban `[Varian aplikasi]` | Pilih beberapa opsi (seperti PG Kompleks, label/format beda) | Otomatis **PROPORSIONAL** = benar_dipilih / total_kunci × bobot, tanpa penalti. |
| 8 | `urutkan` | Mengurutkan `[Varian aplikasi]` | Urutkan langkah/peristiwa | Otomatis **PROPORSIONAL** per posisi benar = posisi_benar / total_item × bobot. |
| 9 | `isi_ganda` | Isian Multi-Kolom `[Varian aplikasi]` | Isi beberapa kolom | Otomatis **PROPORSIONAL** per kolom benar (normalisasi sama seperti isian). |
| 10 | `teks_rumpang` | Teks Rumpang `[Varian aplikasi]` | Isi rumpang (`___`) dalam paragraf | Otomatis **PROPORSIONAL** per rumpang benar. Jumlah `___` harus sama dengan jumlah jawaban kunci. |
| 11 | `peta_konsep` | Peta Konsep `[Varian aplikasi]` | Pasangkan simpul-simpul konsep | Otomatis **PROPORSIONAL** per pasangan benar (seperti menjodohkan). |
| 12 | `studi_kasus` | Studi Kasus + PG `[Varian aplikasi]` | Baca stimulus panjang (kasus), jawab PG | Otomatis 1/0 × bobot. **Stimulus wajib diisi.** |
| 13 | `pernyataan_sikap` | Skala Sikap `[Varian aplikasi]` | Skala 1–4 per pernyataan | **NON-NILAI — hanya survei/refleksi, tidak masuk nilai akhir.** UI wajib menulis jelas "tidak dinilai". |
| 14 | `upload_file` | Upload Berkas `[Varian aplikasi]` | Unggah berkas jawaban | **MANUAL** oleh guru/penguji (rubrik). |
| 15 | `praktik_singkat` | Praktik Singkat `[Varian aplikasi]` | Teks + berkas opsional | **MANUAL** dengan rubrik sederhana. |
| 16 | `pilihan_gambar` | Pilihan Gambar `[Varian aplikasi]` | Pilih 1 opsi berupa gambar | Otomatis 1/0 × bobot. |

## Skema `jawaban` / `kunci` (JSONB) per tipe

`jawaban` = apa yang ditampilkan ke peserta (tanpa kunci).
`kunci` = kunci + rubrik (hanya guru/admin; tidak pernah ke siswa).

| Tipe | `jawaban` | `kunci` | Jawaban siswa (untuk penilaian) |
|------|-----------|---------|----------------------------------|
| `pg`, `studi_kasus`, `pilihan_gambar` | `{opsi: [{label: "A", teks: "..."}]}` — untuk `pilihan_gambar`, `teks` = URL gambar | `{opsi: "B"}` | `{opsi: "B"}` |
| `pg_kompleks`, `pg_multi` | `{opsi: [...]}` | `{opsi: ["A", "C"]}` | `{opsi: ["A", "C"]}` |
| `benar_salah` | `{pernyataan: ["..."]}` | `{nilai: [true, false]}` | `{nilai: [true, false]}` |
| `menjodohkan`, `peta_konsep` | `{pasangan: [{kiri, kanan}]}` | `{pasangan: [{kiri, kanan}]}` (pasangan benar) | `{pasangan: [{kiri, kanan}]}` |
| `isian` | `{}` | `{jawaban: ["Jakarta", "DKI Jakarta"]}` (alternatif) | `{jawaban: "jakarta"}` |
| `uraian` | `{}` (boleh `{panduan: "..."}`) | `{rubrik: "..."}` | `{jawaban: "teks esai..."}` |
| `urutkan` | `{item: ["Langkah 1", ...]}` (urutan acak tampil) | `{urutan: [2, 0, 1]}` (indeks urutan benar) | `{urutan: [2, 0, 1]}` |
| `isi_ganda` | `{kolom: ["Kolom 1", "Kolom 2"]}` | `{jawaban: ["...", "..."]}` | `{jawaban: ["...", "..."]}` |
| `teks_rumpang` | `{}` (rumpang ada di `pertanyaan` sebagai `___`) | `{jawaban: ["...", "..."]}` | `{jawaban: ["...", "..."]}` |
| `pernyataan_sikap` | `{pernyataan: [...], skala: [1,2,3,4]}` | `{skala: [1,2,3,4]}` | `{nilai: [3, 4]}` (tidak dinilai) |
| `upload_file` | `{format: ["pdf","docx"], maks_mb: 10}` | `{rubrik: "..."}` | `{file_url: "..."}` |
| `praktik_singkat` | `{}` | `{rubrik: "...", aspek: [...]}` | `{jawaban: "...", file_url: "..."}` |

## Import CSV

`POST /api/v1/bank-soal/import` — body JSON
`{bank_id, csv_text, mode}`. `mode`: `all_or_nothing` (default —
bila ada 1 baris salah, tidak ada yang disimpan) atau `skip_errors`
(baris valid disimpan, baris salah dilaporkan). Hasil:
`{sukses, gagal, errors: [{baris, pesan}], mode, disimpan}`.
Maks **500 baris** per berkas. Soal hasil import masuk berstatus
`draft` (wajib telaah).

Kolom CSV:
`tipe,pertanyaan,opsi_a,opsi_b,opsi_c,opsi_d,opsi_e,kunci,bobot,level,stimulus`

Nilai yang mengandung koma (mis. kunci `B,S` untuk benar-salah) wajib
dikutip sesuai standar CSV (`"B,S"`).

Arti `kunci` per tipe di CSV:

| Tipe | `opsi_a..opsi_e` | `kunci` |
|------|------------------|---------|
| `pg`, `studi_kasus`, `pilihan_gambar` | teks opsi / URL gambar | satu label, mis. `B` |
| `pg_kompleks`, `pg_multi` | teks opsi | label dipisah koma, mis. `A,C` |
| `benar_salah` | pernyataan | `B,S` (B=benar, S=salah), dipisah koma, sebanyak pernyataan |
| `menjodohkan`, `peta_konsep` | pasangan `kiri::kanan` | pasangan benar `kiri::kanan;kiri::kanan` (dipisah `;`) |
| `isian` | (diabaikan) | jawaban, alternatif dipisah `\|` |
| `uraian`, `upload_file`, `praktik_singkat` | (diabaikan) | rubrik (teks) |
| `urutkan` | butir | urutan benar sebagai label, mis. `C,A,B` |
| `isi_ganda` | (diabaikan) | jawaban per kolom dipisah `\|` |
| `teks_rumpang` | (diabaikan) | jawaban per rumpang dipisah `\|` |
| `pernyataan_sikap` | pernyataan | (diabaikan — skala 1–4) |

## Lock bank soal

- Bank terkunci → soal **tidak bisa** ditambah/diubah/dihapus/diimport
  (mutasi soal = `409 bank_locked`). Bank masih bisa dibaca.
- Hapus bank ditolak bila terkunci (`409`) atau dipakai paket (`409`).
- Lock/unlock: `POST /api/v1/bank-soal/:id/lock` dan `/unlock`
  (permission `banksoal.lock`).
- **SoD-1** (ANALISIS Bab 11.4): pembuat bank TIDAK BOLEH
  mengunci/membuka bank buatannya sendiri — pengunci harus guru lain
  atau admin. Pelanggaran = `403 sod_violation`.

## Permission (seed Tim 6)

| Role | Akses bank soal |
|------|-----------------|
| Admin | Semua (`banksoal.*`) |
| Guru | Semua (`banksoal.*`), mutasi hanya bank miliknya / tanpa pemilik |
| Proktor | Hanya `banksoal.read` untuk **metadata bank** (daftar bank + jumlah soal). **Tidak boleh melihat isi soal** — endpoint soal mengembalikan `403`. |
| Pengawas | Tidak ada akses sama sekali → `403` di semua endpoint bank soal |
| Siswa | Tidak ada akses (endpoint siswa milik Tim 4) → `403` |
| Penguji UKK | Tidak ada akses → `403` |

## Endpoint (Tim 3)

| Method & path | Permission |
|---|---|
| `GET /api/v1/bank-soal` (search `q`, filter `mapel_id`, `kelas_tingkat`, `is_locked`, pagination) | `banksoal.read` |
| `POST /api/v1/bank-soal` | `banksoal.create` |
| `GET /api/v1/bank-soal/:id` | `banksoal.read` |
| `PUT /api/v1/bank-soal/:id` | `banksoal.update` |
| `DELETE /api/v1/bank-soal/:id` | `banksoal.delete` |
| `POST /api/v1/bank-soal/:id/lock` | `banksoal.lock` |
| `POST /api/v1/bank-soal/:id/unlock` | `banksoal.lock` |
| `POST /api/v1/bank-soal/import` | `banksoal.import` |
| `GET /api/v1/bank-soal/:bankId/soal` (filter `tipe`, `level`, `kisi_kisi_id`, `q`, pagination) | `banksoal.read` (guru/admin) |
| `POST /api/v1/bank-soal/:bankId/soal` | `banksoal.create` |
| `GET /api/v1/bank-soal/:bankId/soal/:soalId` | `banksoal.read` (guru/admin) |
| `PUT /api/v1/bank-soal/:bankId/soal/:soalId` | `banksoal.update` |
| `DELETE /api/v1/bank-soal/:bankId/soal/:soalId` | `banksoal.delete` |

Format sukses: `{"data": ...}`. Format error:
`{"error": {"code": "...", "message": "..."}}`.
