# UKK — Rubrik, Sertifikat & Upload (Paket Pendukung `ukkx`)

Spesifikasi + implementasi referensi untuk penilaian UKK CBT SMK v2.
Paket Go: `backend/internal/ukkx/` (fungsi murni, tanpa DB, tanpa HTTP,
tanpa dependensi di luar standard library). Paket ini adalah **cross-check**
independen terhadap handler UKK utama (paket `ukk` milik Tim UKK utama):
bila kedua implementasi memberi hasil berbeda untuk input yang sama,
salah satunya bug — jangan rilis sebelum selisihnya dijelaskan.

Aturan FINAL (satu format rubrik + satu format sertifikat, semua jurusan).
Sumber: tugas Tim UKK Support + `ANALISIS.md` Bab UKK / §12.3–12.4 +
`docs/KONTRAK.md` §8.6. Perbedaan label predikat terhadap ANALISIS
dijelaskan di §5.

---

## 1. Rubrik & nilai akhir

| Komponen | Bobot |
|---|---|
| Nilai praktik: Proses | 30% |
| Nilai praktik: Produk | 50% |
| Nilai praktik: Sikap | 20% |
| **Nilai akhir: Teori** | **40%** |
| **Nilai akhir: Praktik** | **60%** |

Rumus:

```
praktik = proses × 0.30 + produk × 0.50 + sikap × 0.20
akhir   = teori  × 0.40 + praktik × 0.60
```

- Semua komponen dalam skala **0–100**. Nilai di luar itu ditolak
  (`ValidasiNilai`: menolak `< 0`, `> 100`, `NaN`, `±Inf`).
- Semua hasil dibulatkan ke **2 desimal, round half-up** pada representasi
  desimalnya. Implementasi memakai aritmetika rasional eksak
  (`math/big.Rat`) atas representasi desimal terpendek, BUKAN
  `math.Round(v*100)/100` mentah — karena `3.35 × 0.3` tersimpan sebagai
  `1.0049999999999999` dalam biner dan akan salah dibulatkan ke `1.00`
  oleh cara naif. Dengan cara eksak: `1.005 → 1.01`.
- Nilai dihitung **backend**; frontend TIDAK PERNAH menghitung nilai
  (KONTRAK §0).

### Kelulusan & predikat

| Nilai akhir | Status | Predikat |
|---|---|---|
| `< 70` | Belum lulus | Belum Tuntas |
| `70 – 79.99…` | Lulus | Cukup |
| `80 – 90.99…` | Lulus | Baik |
| `91 – 100` | Lulus | Sangat Baik |

- Ambang lulus: **nilai akhir ≥ 70** (tepat 70 lulus, 69.99 tidak lulus).
- Batas predikat dievaluasi pada nilai yang sudah dibulatkan 2 desimal
  dari `HitungAkhir`. `Predikat(NaN)` = "Belum Tuntas" (input tidak valid
  tidak boleh menghasilkan predikat lulus).

---

## 2. Contoh hitung lengkap

> **KOREKSI terhadap contoh di brief tugas.** Brief menulis:
> "teori 80, proses 75, produk 85, sikap 90 → praktik = 83.5 → akhir = 82.1".
> Angka itu **tidak mengikuti rumus di atas**. Hitung yang benar:
>
> ```
> praktik = 75×0.30 + 85×0.50 + 90×0.20
>         = 22.50 + 42.50 + 18.00
>         = 83.00          (bukan 83.5)
> akhir   = 80×0.40 + 83.00×0.60
>         = 32.00 + 49.80
>         = 81.80          (bukan 82.1)
> ```
>
> Hasil akhir: **81.80 → lulus, predikat Baik**. Implementasi `ukkx`
> memakai rumus (83.0 / 81.8), dan unit test `TestHitungAkhir_RantaiPenuh`
> mengunci nilai yang benar ini. Bila brief dimaksudkan memakai bobot /
> input lain, perbaiki brief-nya — jangan "memperbaiki" rumus agar cocok
> dengan 83.5/82.1.

Contoh tambahan (kasus batas):

| Teori | Proses | Produk | Sikap | Praktik | Akhir | Lulus | Predikat |
|---|---|---|---|---|---|---|---|
| 70 | 70 | 70 | 70 | 70.00 | 70.00 | Ya | Cukup |
| 69.95 | 70 | 70 | 70 | 70.00 | 69.98 | Tidak | Belum Tuntas |
| 100 | 100 | 100 | 100 | 100.00 | 100.00 | Ya | Sangat Baik |
| 80 | 90 | 80 | 95 | 86.00 | 83.60 | Ya | Baik |
| 0 | 0 | 0 | 0 | 0.00 | 0.00 | Tidak | Belum Tuntas |

---

## 3. Sertifikat

### Format nomor

```
UKK/<tahun>/<kode-jurusan>/<nomor-urut 4 digit>
```

Contoh: `UKK/2026/RPL/0001`, `UKK/2026/TKJ/0042`, `UKK/2026/MM/9999`.

Aturan:

- Tahun: `2000–2100`, persis 4 digit.
- Kode jurusan: alfanumerik (`[A-Za-z0-9]`), **2–10 karakter**,
  dinormalisasi ke huruf besar saat dibentuk (`rpl` → `RPL`).
- Nomor urut: `1–9999`, ditulis zero-pad 4 digit (`1` → `0001`).
- `NomorSertifikat` mengembalikan `""` untuk input tidak valid
  (signature-nya mengembalikan string saja, sesuai kontrak tugas);
  pemanggil wajib memeriksa `""` sebelum menyimpan.
- `ParseNomorSertifikat` memvalidasi penuh (termasuk urut `0000`
  ditolak, urut `1` tanpa pad ditolak, prefix harus `UKK`) dan
  mengembalikan komponen + error.

### Kapan sertifikat boleh terbit

```
BolehTerbit(nilai, lulus) = lulus && nilai >= 70
```

**Sertifikat HANYA bagi yang lulus.** Kedua syarat harus terpenuhi:
nilai ≥ 70 saja tanpa flag lulus (mis. nilai belum difinalisasi)
TIDAK cukup. Finalisasi nilai UKK tetap kewenangan terpisah dari
penguji (SoD KONTRAK §5.3: penguji UKK ≠ finalisator nilai UKK).

Catatan kewenangan penerbitan: sesuai ANALISIS Bab UKK, CBT menyediakan
**data + draf sertifikat**, bukan menerbitkan sertifikat hukum sendiri —
sertifikat resmi diterbitkan lembaga berwenang (BNSP/LSP/DUDIKA/sekolah).
Nomor di sini adalah nomor draf/administratif CBT; tim integrasi wajib
menjamin keunikan nomor per `(tahun, kode-jurusan)` di database
(UNIQUE constraint + alokasi urut transaksional) — `ukkx` tidak
memiliki akses DB sehingga tidak bisa menjamin keunikan itu sendiri.

---

## 4. Upload berkas UKK (bukti praktik / portofolio)

### Matriks validasi upload

| Aturan | Nilai | Hasil |
|---|---|---|
| Ekstensi diizinkan | `pdf, doc, docx, jpg, jpeg, png, zip` (case-insensitive) | ✅ Diterima |
| Ekstensi ditolak | `exe, sh, bat, html, js, php, xlsm, xlsb`, tanpa ekstensi, `file.` | ❌ Ditolak |
| Ekstensi ganda | `laporan.pdf.exe` → dinilai dari ekstensi TERAKHIR (`.exe`) | ❌ Ditolak |
| Ekstensi ganda aman | `laporan.final.pdf` → ekstensi terakhir `.pdf` | ✅ Diterima |
| Ukuran | `0` s.d. tepat `10 MB` (10.485.760 byte) | ✅ Diterima |
| Ukuran | `10 MB + 1 byte`, negatif | ❌ Ditolak |
| Nama file | mengandung `/`, `\`, `..`, karakter kontrol (termasuk NUL) | ❌ Ditolak |
| Nama file | kosong | ❌ Ditolak |
| Nama file | `../../etc/passwd` | ❌ Ditolak (traversal + tanpa ekstensi diizinkan) |
| Nama file | ada spasi / strip / titik ganda (`laporan akhir (1).pdf` di lapisan path menjadi `laporan_akhir__1_.pdf`) | ✅ Diterima (lihat §4.1) |

### 4.1 Path penyimpanan

```
storage/ukk/<ukkID>/<siswaID>/<nama-bersih>
```

`PathAman(ukkID, siswaID, namaFile)` membersihkan nama: hanya
`[a-zA-Z0-9._-]` yang dipertahankan, sisanya menjadi `_`; titik-titik
di awal nama diganti `_`; ID UKK dan ID siswa ikut dibersihkan dengan
aturan yang sama agar ID kiriman client tidak bisa menyisipkan
traversal. Hasilnya SELALU berada di bawah `storage/ukk/` (tidak pernah
mengandung segmen `..` sebagai komponen path).

Contoh:

| Input | Path |
|---|---|
| `("ukk-1", "siswa-1", "laporan.pdf")` | `storage/ukk/ukk-1/siswa-1/laporan.pdf` |
| `("ukk-1", "siswa-1", "../../etc/passwd")` | `storage/ukk/ukk-1/siswa-1/___.._etc_passwd` |
| `("ukk-1", "siswa-1", "laporan akhir (1).pdf")` | `storage/ukk/ukk-1/siswa-1/laporan_akhir__1_.pdf` |

### 4.2 Lapisan yang TIDAK ditangani `ukkx` (wajib di handler)

`ValidasiBerkas` hanya memeriksa **nama + ukuran**. Tim integrasi TETAP
wajib menambahkan, sesuai ANALISIS Bab upload:

1. **Sniffing magic bytes** — jangan percaya ekstensi / `Content-Type` client.
2. **ZIP**: batasi jumlah entri + ukuran total hasil ekstrak (anti zip bomb),
   tolak path traversal `../` di dalam arsip, tolak symlink.
3. **Penyimpanan di luar webroot** + serve hanya via handler terautentikasi;
   berkas tidak pernah dieksekusi.
4. **Kuota per peserta** per ujian praktik + rate limit upload per jam.
5. **Batas ukuran di reverse proxy / `MaxBytesReader`** sebelum parsing,
   dan streaming ke disk (bukan memuat seluruh file ke RAM).

---

## 5. Catatan perbedaan dokumen

- **Label predikat.** ANALISIS §12.3 memakai label POS UKK:
  "<70 Belum Kompeten; 70–79 Cukup Kompeten; 80–90 Kompeten;
  91–100 Sangat Kompeten". Tugas FINAL Tim UKK Support memakai label:
  "Belum Tuntas / Cukup / Baik / Sangat Baik". Paket `ukkx` memakai
  label tugas final (yang mengikat tim ini). Batas angkanya identik
  (70 / 80 / 91), hanya labelnya berbeda — selaraskan label di UI/rekap
  agar tidak tampil dua versi.
- **Ambang per unit vs total.** ANALISIS juga menyebut kelulusan per
  unit kompetensi + uji ulang per unit. Paket `ukkx` menghitung nilai
  per skema/peserta (total); logika per-unit kompetensi adalah tanggung
  jawab handler UKK utama (paket `ukk`), bukan paket pendukung ini.

---

## 6. Referensi API paket `ukkx`

```go
import "cbt-smk-v2/backend/internal/ukkx"

// Rubrik (rubrik.go)
ukkx.HitungPraktik(proses, produk, sikap float64) float64 // proses*.3+produk*.5+sikap*.2
ukkx.HitungAkhir(teori, praktik float64) float64          // teori*.4+praktik*.6
ukkx.Lulus(nilai float64) bool                            // nilai >= 70
ukkx.Predikat(nilai float64) string                       // Belum Tuntas/Cukup/Baik/Sangat Baik
ukkx.ValidasiNilai(v float64) error                       // 0-100, tolak NaN/Inf

// Sertifikat (sertifikat.go)
ukkx.NomorSertifikat(tahun int, kodeJurusan string, urut int) string
ukkx.ParseNomorSertifikat(s string) (tahun, kode, urut, err)
ukkx.BolehTerbit(nilai float64, lulus bool) bool           // lulus && nilai>=70

// Upload (upload.go)
ukkx.ValidasiBerkas(namaFile string, ukuran int64) error
ukkx.PathAman(ukkID, siswaID, namaFile string) string
ukkx.UkuranMaksBerkasUKK // = 10 * 1024 * 1024
```

Verifikasi:

```
cd backend
go test ./internal/ukkx/ -v   # 14 test, semua lolos
go vet ./internal/ukkx/       # bersih
```
