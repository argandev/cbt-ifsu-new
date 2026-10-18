# Analisis Butir Soal & Nilai per Kompetensi — Rumus

Dokumen ini adalah acuan RUMUS untuk fitur Analisis Butir Soal CBT SMK v2.
Definisi di sini adalah **FINAL** dan mengikat. Bila bertentangan dengan
`ANALISIS.md` Bab 8.4 (mis. ambang kesukaran / label daya beda), definisi
di sini yang dipakai — Bab 8.4 ditulis sebelum keputusan final.

Cakupan perhitungan: **per soal dari attempt SELESAI pada satu jadwal**.
Peserta yang belum submit TIDAK dihitung. Implementasi fungsi murni ada di
`backend/internal/analisisx/statistik.go` (tanpa DB); visualisasi ada di
`frontend/src/components/analisis/`.

---

## 1. Tingkat kesukaran (P)

```
P = B / N
B = jumlah peserta yang menjawab BENAR soal ini
N = jumlah peserta yang menjawab soal ini (attempt selesai)
```

| P | Label |
|---|---|
| P < 0,3 | Sukar |
| 0,3 – 0,7 | Sedang |
| P > 0,7 | Mudah |

Batas inklusif: tepat 0,3 dan tepat 0,7 = **Sedang**.

**Apa itu "benar":**

- Soal auto (PG, PG Kompleks, Benar-Salah, Menjodohkan, Isian Singkat):
  benar bila **poin == bobot** (poin penuh). PG Kompleks / Menjodohkan
  yang proporsional: jawaban setengah benar (poin < bobot) = **tidak benar**
  untuk keperluan P dan D.
- Soal uraian / manual: benar bila **poin >= 0,6 × bobot**.
- Esai yang belum dikoreksi = "belum dinilai", BUKAN 0 — tidak masuk
  hitungan benar maupun penjawab sampai terkoreksi.

**Kasus batas:** N = 0 → P = 0, label "tidak ada data".

---

## 2. Daya beda (D)

```
D = (BA − BB) / n
BA = jumlah jawaban benar di kelompok ATAS
BB = jumlah jawaban benar di kelompok BAWAH
n  = jumlah peserta PER kelompok (atas dan bawah sama besar)
```

### 2.1 Pembentukan kelompok

Peserta diurutkan menurun berdasarkan **nilai total attempt** pada jadwal
itu (bukan nilai soal ini).

- **N ≥ 10 — metode 27%:** tiap kelompok berisi pembulatan terdekat dari
  27% × N peserta (minimal 1). Contoh: N = 100 → 27 + 27; N = 10 → 3 + 3.
- **N < 10 — median split + flag "sampel kecil":** kelompok atas =
  separuh atas, kelompok bawah = separuh bawah. Bila N ganjil, peserta
  median tengah TIDAK masuk kelompok mana pun. Contoh: N = 9 → 4 + 4
  (1 terbuang); N = 8 → 4 + 4.
- **N ≤ 1:** tidak ada kelompok; D tidak dapat dihitung.

Peserta dengan nilai sama persis di batas kelompok dipisah berdasarkan
urutan stabil (urutan awal dipertahankan) — keterbatasan yang dicatat di
`BagiKelompok`. Seri (ties) yang masif di batas sebaiknya ditelaah manual.

### 2.2 Interpretasi

| D | Interpretasi |
|---|---|
| D ≥ 0,40 | Sangat Baik |
| 0,30 – 0,39 | Baik |
| 0,20 – 0,29 | Cukup |
| 0,00 – 0,19 | Buruk |
| D < 0,00 | Sangat Buruk |

**D negatif** = kelompok bawah lebih banyak benar daripada kelompok atas →
kunci jawaban dicurigai, prioritas tinjau **tertinggi**. Catatan ini
ditampilkan bersama hasil, bukan disembunyikan.

---

## 3. Pengecoh (distractor)

Untuk soal berbasis opsi (PG, PG Kompleks, Benar-Salah, Menjodohkan per
pasangan), per opsi:

```
persen opsi = pemilih opsi / total penjawab × 100
```

| Kondisi | Status |
|---|---|
| Dipilih ≥ 5% peserta | Berfungsi |
| Dipilih < 5% peserta | **Tidak berfungsi** (opsi mati, perlu revisi/ganti) |

Aturan ini menilai **pengecoh**, bukan kunci: untuk kunci jawaban, kolom
"berfungsi" tetap dihitung tetapi tidak dibaca sebagai penilaian kunci
(kunci yang dipilih sedikit adalah soal sukar / kunci dicurigai, bukan
"pengecoh tidak berfungsi"). Soal uraian / manual tidak memiliki opsi →
pengecoh **N/A**.

Catatan praktis: dengan N kecil, ambang 5% hampir tidak pernah tercapai
oleh satu pemilih pun — 1 pemilih dari 10 peserta = 10%. Pengecoh
"tidak berfungsi" baru bermakna pada N ≥ 20 (1 dari 20 = 5%).

---

## 4. Nilai per kompetensi

Per siswa, per jadwal, per kompetensi:

```
Nilai kompetensi = Σ poin soal kompetensi itu / Σ bobot soal kompetensi itu × 100
```

- Hanya soal **yang dijawab** yang masuk pembilang maupun penyebut.
  Soal kosong tidak menekan nilai kompetensi.
- Bila siswa tidak menjawab satu pun soal kompetensi itu → nilai
  **null** + flag **"tidak ada soal"** (ditampilkan "—" di heatmap).
- Bila ada jawaban tetapi total bobot 0 → nilai null + flag "bobot nol"
  (data tidak konsisten, perlu diperbaiki).
- Satu soal → satu kompetensi utama (multi-kompetensi ditunda, sesuai
  ANALISIS.md).
- Esai belum dikoreksi = "belum dinilai", bukan 0 — nilai kompetensi
  yang mengandung esai belum terkoreksi bersifat sementara.

---

## 5. Contoh hitung manual langkah-demi-langkah

### 5.1 Data: 10 peserta, 1 soal PG (kunci: B)

Nilai total attempt (untuk pengelompokan) dan jawaban soal ini:

| # | Peserta | Nilai total | Jawaban | Benar? |
|---|---------|-------------|---------|--------|
| 1 | Andi | 92 | B | Ya |
| 2 | Budi | 88 | B | Ya |
| 3 | Citra | 85 | B | Ya |
| 4 | Dedi | 78 | A | Tidak |
| 5 | Eka | 70 | B | Ya |
| 6 | Fajar | 65 | C | Tidak |
| 7 | Gita | 58 | A | Tidak |
| 8 | Hadi | 50 | D | Tidak |
| 9 | Intan | 44 | A | Tidak |
| 10 | Joko | 38 | B | Ya |

### 5.2 Langkah 1 — Kesukaran

Benar (B) = Andi, Budi, Citra, Eka, Joko = **5**.
Penjawab (N) = **10**.

```
P = 5 / 10 = 0,50  →  0,3 ≤ 0,50 ≤ 0,7  →  Sedang
```

### 5.3 Langkah 2 — Kelompok atas/bawah

N = 10 → N ≥ 10, pakai **metode 27%** (bukan median split):

```
n per kelompok = round(27% × 10) = round(2,7) = 3
```

Peserta sudah terurut menurun berdasarkan nilai total:

- Kelompok **atas** (3 teratas): Andi (92), Budi (88), Citra (85).
- Kelompok **bawah** (3 terbawah): Hadi (50), Intan (44), Joko (38).
- 4 peserta tengah (Dedi, Eka, Fajar, Gita) tidak dipakai untuk D.

### 5.4 Langkah 3 — Daya beda

Benar di kelompok atas (BA): Andi, Budi, Citra = **3**.
Benar di kelompok bawah (BB): hanya Joko = **1** (Hadi D, Intan A).

```
D = (3 − 1) / 3 = 2 / 3 ≈ 0,67  →  D ≥ 0,40  →  Sangat Baik
```

Catatan: tidak ada flag "sampel kecil" (N = 10), D positif (kunci tidak
dicurigai). Soal ini membedakan peserta pandai dan kurang pandai dengan
sangat baik.

### 5.5 Langkah 4 — Pengecoh

Total penjawab = 10.

| Opsi | Pemilih | Persen | Status |
|------|---------|--------|--------|
| A (pengecoh) | Dedi, Gita, Intan = 3 | 3/10 × 100 = 30% | Berfungsi (≥ 5%) |
| **B (kunci)** | Andi, Budi, Citra, Eka, Joko = 5 | 50% | Kunci — tidak dinilai sebagai pengecoh |
| C (pengecoh) | Fajar = 1 | 10% | Berfungsi (≥ 5%) |
| D (pengecoh) | Hadi = 1 | 10% | Berfungsi (≥ 5%) |

Semua pengecoh berfungsi pada contoh ini. Perhatikan: dengan N = 10,
satu pemilih saja sudah 10% — ambang 5% belum bisa membedakan.

### 5.6 Varian: pengecoh tidak berfungsi (N = 100)

Bila pada jadwal lain soal yang sama dijawab 100 peserta dan opsi D
hanya dipilih 4 peserta:

```
persen D = 4 / 100 × 100 = 4%  →  < 5%  →  Tidak berfungsi
```

Opsi D ditandai "Tidak berfungsi" di tabel pengecoh dan perlu direvisi
atau diganti. Bandingkan: 5 dari 100 = tepat 5% → masih **berfungsi**;
4,9% (mis. 49 dari 1000) → tidak berfungsi; 5,1% (51 dari 1000) →
berfungsi.

### 5.7 Varian: sampel kecil (N = 8, median split)

Bila hanya 8 peserta (nilai total 88, 81, 75, 70, 62, 55, 48, 40):

```
Median split: atas = 4 teratas (88, 81, 75, 70), bawah = 4 terbawah (62, 55, 48, 40)
```

D dihitung dengan rumus yang sama, tetapi hasil WAJIB ditampilkan
dengan flag **"sampel kecil (n < 10): daya beda tidak dapat
dipercaya"**. Guru tidak boleh membuang soal hanya berdasarkan D dari
sampel sekecil ini.

---

## 6. Contoh nilai per kompetensi

Siswa menjawab 3 soal kompetensi "Pemrograman Dasar" (bobot masing-masing
10) dan 2 soal kompetensi "Jaringan" (bobot masing-masing 10):

| Kompetensi | Soal | Poin | Bobot | Dijawab? |
|---|---|---|---|---|
| Pemrograman Dasar | S1 | 8 | 10 | Ya |
| Pemrograman Dasar | S2 | 6 | 10 | Ya |
| Pemrograman Dasar | S3 | 0 | 10 | Tidak (kosong) |
| Jaringan | S4 | — | 10 | Tidak |
| Jaringan | S5 | — | 10 | Tidak |

```
Pemrograman Dasar = (8 + 6) / (10 + 10) × 100 = 14 / 20 × 100 = 70
  (S3 yang kosong TIDAK masuk: bukan (8+6+0)/30 × 100 = 46,7)

Jaringan = null + flag "tidak ada soal"  →  tampil "—" di heatmap
```

---

## 7. Tabel interpretasi ringkas

| Metrik | Rumus | Ambang |
|---|---|---|
| Kesukaran P | B / N | < 0,3 Sukar; 0,3–0,7 Sedang; > 0,7 Mudah |
| Daya beda D | (BA − BB) / n | ≥ 0,40 Sangat Baik; 0,30–0,39 Baik; 0,20–0,29 Cukup; 0,00–0,19 Buruk; < 0 Sangat Buruk |
| Pengecoh | pemilih / total × 100 | ≥ 5% berfungsi; < 5% tidak berfungsi |
| Nilai kompetensi | Σ poin / Σ bobot × 100 | null + "tidak ada soal" bila tak ada jawaban |

---

## 8. Peringatan N kecil

1. **N < 10:** D tidak dapat dipercaya — selalu tampilkan flag
   "sampel kecil". Jangan jadikan D satu-satunya dasar revisi/buang soal.
2. **N < 20:** ambang pengecoh 5% hampir tidak bermakna (1 pemilih =
   > 5%). Pengecoh "berfungsi" pada N kecil belum tentu benar-benar
   berfungsi.
3. **N = 0 (tidak ada attempt selesai):** tidak ada statistik yang dapat
   dihitung — tampilkan "tidak ada data", bukan angka 0 yang menyesatkan.
4. **N ≤ 1:** kelompok atas/bawah tidak dapat dibentuk — D tidak dapat
   dihitung.
5. **Soal dengan P = 0 atau P = 1** (tidak ada / semua peserta benar)
   wajib ditinjau terlepas dari D — soal seperti itu tidak memberi
   informasi apa pun tentang perbedaan kemampuan peserta.
6. **Selalu tampilkan N (pembagi)** di samping setiap statistik, sesuai
   aturan ANALISIS.md: "jangan campur N (pembagi harus eksplisit)".

---

## 9. Catatan integrasi (untuk Tim Analisis utama)

- **Backend:** `analisisx` adalah paket fungsi murni — handler Tim
  Analisis memanggil `analisisx.Kesukaran`, `analisisx.BagiKelompok` +
  `analisisx.DayaBeda`, `analisisx.Pengecoh`, `analisisx.NilaiKompetensi`
  setelah mengambil agregat dari DB (attempt selesai, jawaban, nilai).
  Handler tetap milik Tim Analisis (`backend/internal/analisis/`) dan
  TIDAK diubah oleh paket ini. (Satu penyimpangan penamaan yang
  dipaksakan bahasa Go: tipe hasil pengecoh bernama `PengecohItem`,
  karena tipe dan fungsi tidak boleh sama-sama bernama `Pengecoh`.)
- **Frontend:** halaman Analisis utama mengimpor komponen dari
  `frontend/src/components/analisis/`:
  `DistribusiChart` (`{nilai: number[]}`),
  `DayaBedaBadge` (`{nilai: number}`),
  `PengecohTable` (`{pengecoh: {opsi, pemilih, persen, berfungsi}[]}`),
  `KompetensiHeatmap` (`{siswa: {nama, nilai: Record<string, number|null>}[]}`).
  Semua komponen hanya menampilkan data dari API — tidak menghitung
  nilai (kontrak: nilai SELALU dihitung backend).
- **Tidak ada endpoint baru** dari paket ini; endpoint Analisis tetap
  milik Tim Analisis utama.
