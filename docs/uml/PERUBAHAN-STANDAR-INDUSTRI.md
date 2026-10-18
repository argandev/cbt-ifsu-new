# PERUBAHAN — Standardisasi OMG + Industri Diagram UML CBT SMK v2

- **File yang didokumentasikan:** `docs/uml/cbt-smk-v2.mdj`
- **SHA-256 final:** `c068a179be3df130e566dd26ea826c27a4368e09463d2cf5937eae0dee1025df` (2.199.061 byte)
- **Tanggal penyelesaian fase ini:** 17 Oktober 2026
- **Cakupan dokumen ini:** fase STANDARDISASI OMG + INDUSTRI saja (riset → audit →
  Sequence → Diagram Lain → Perbaikan Putaran 1 (T1+T2) → Perbaikan Putaran 2
  (verifikasi, 2 usulan DITOLAK) → ekspor + dokumentasi). Fase-fase sebelumnya
  (audit OMG F1–F5, standardisasi skripsi, line-neatness G1–G5, R9a–R9d, R10)
  dirujuk hanya sebagai konteks / rantai SHA.
- **Prinsip yang mengikat seluruh fase:** bila memperbaiki berarti mengubah
  SEMANTIK BISNIS (relasi, atribut, alur, pesan, klasifikasi, nama) — JANGAN
  UBAH; catat sebagai temuan + rekomendasi. Yang diubah hanya (a) field model
  yang nilainya sudah benar di ground truth tetapi salah ter-render, (b) teks
  dokumentasi/narasi diagram (non-semantik), dan (c) geometri view (posisi,
  points, label) tanpa mengubah model.

## Rantai SHA file (jejak audit)

| Tahap | SHA-256 (8 awal) | Ukuran | Bukti |
|---|---|---|---|
| Baseline pra-audit-OMG | `10192994…` | — | `cbt-smk-v2.mdj.pre-omg-audit.bak` |
| Sesudah audit OMG F1–F5 + skripsi + line-neatness (masuk fase industri) | `d310d2c4…` | 2.194.431 | `cbt-smk-v2.mdj.pre-audit-industri.bak` |
| Sesudah Audit Industri (K7 + K4) | `cc749b66…` | 2.198.375 | `cbt-smk-v2.mdj.pre-sequence-industri.bak` |
| Sesudah Divisi Sequence (pitch 180→165 + label) | `158d9dd1…` | — | `cbt-smk-v2.mdj.pre-t1t2.bak`; Divisi Diagram Lain: 0 perubahan, SHA tetap |
| **FINAL sesudah Putaran 1 (T1+T2)** | **`c068a179…`** | **2.199.061** | `docs/uml/cbt-smk-v2.mdj` + `your_files/cbt-smk-v2.mdj` (tersinkron) |

Skrip perbaikan tiap divisi bersifat idempotent (run ke-2: 0 perubahan) dan
dijaga diff guard + verifikasi geometri sebelum simpan (gagal → file TIDAK
ditulis). File TIDAK pernah dibuka-disimpan lewat StarUML pada fase ini;
semua perubahan adalah edit JSON `.mdj` yang terverifikasi.

## 1. Ringkasan riset

Satu-satunya acuan aturan adalah dokumen sintesis
`docs/uml/RISET-STANDAR-UML.md` (17 Okt 2026), yang disusun dari TIGA dokumen
riset — tidak ada aturan yang dikarang, tidak ada sumber baru yang
ditambahkan. `cbt-smk-v2.mdj` TIDAK diubah selama penyusunan dokumen aturan.

### 1.1 R1 — Spesifikasi OMG UML 2.5.1 (tingkat bukti 1, mengikat)

- **Dokumen:** `docs/uml/riset/R1-omg-uml251.md`.
- **Sumber:** OMG Unified Modeling Language, Version 2.5.1, OMG Document Number
  **formal/2017-12-05**, Desember 2017, https://www.omg.org/spec/UML/2.5.1/PDF.
  Dibaca langsung dari teks spesifikasi resmi (PDF) pada 17 Okt 2026.
- **Isi:** **75 aturan WAJIB** — Sequence SEQ-01–SEQ-26, CombinedFragment
  CF-01–CF-13, Use Case UC-01–UC-12, Class CL-01–CL-12, Activity AC-01–AC-12.
  Setiap aturan mengutip nomor sub-klausul spesifikasi + status verifikasi
  (`TERVERIFIKASI` / `TIDAK TERVERIFIKASI` / `KOREKSI PREMIS`).
- **Yang SENGAJA tidak dipakai R1:** uml-diagrams.org (sekunder, tidak
  dikutip); dokumentasi StarUML (perilaku tool bukan spesifikasi); isi gambar
  contoh spesifikasi yang tidak terekspos sebagai teks (penyebab SEQ-27/SEQ-28
  TIDAK TERVERIFIKASI — lihat §4).
- **3 koreksi premis (membatalkan asumsi lama — lihat §5):** penomoran pesan
  Sequence BUKAN bagian Sequence Diagram; reply TIDAK WAJIB per synchCall;
  guard TIDAK WAJIB per cabang Decision. Setiap temuan audit lama yang
  bersandar pada ketiganya BATAL dan harus ditarik.

### 1.2 R2 — Jurnal peer-reviewed (tingkat bukti 2, rekomendasi kualitas)

- **Dokumen:** `docs/uml/riset/R2-jurnal.md`.
- **Isi:** **10 paper TERVERIFIKASI** (judul + penulis + tahun + venue +
  DOI/URL dicocokkan dengan minimal satu sumber independen) → **28 aturan
  kualitas R2-01–R2-28** (termasuk 3 larangan klaim: R2-04, R2-23, R2-26).
- **Batasan bukti (berlaku untuk SEMUA aturan §2):** sebagian besar naskah
  lengkap berbayar / tidak dapat diakses gratis; temuan kunci diambil dari
  abstrak resmi dan ringkasan penulis pengutip, BUKAN dari naskah lengkap,
  kecuali dinyatakan lain. Aturan §2 adalah rekomendasi kualitas berbasis
  bukti, BUKAN pelanggaran formal OMG.

| Paper | Temuan kunci yang dipakai |
|---|---|
| P1 Purchase (1997), *Graph Drawing*, Springer LNCS — "Which aesthetic has the greatest effect on human understanding?" (DOI tidak terverifikasi) | Persilangan garis = estetika terpenting ("by far the most important"); belokan + simetri berpengaruh lebih kecil; grid ortogonal + sudut minimum TIDAK signifikan → R2-01–R2-04. **Catatan transfer wajib:** eksperimen memakai graf abstrak, bukan diagram UML — penerapan ke UML adalah ekstrapolasi yang masuk akal, bukan bukti langsung pada UML. |
| P2 Purchase et al. (2001), APVis 2001, hlm. 113–120 (ACM ID 564054) — sintaks class diagram | Satu gaya notasi konsisten (R2-05); intuisi "terasa jelas" bukan bukti (R2-06). |
| P3 Purchase et al. (2002), VISSOFT 2002, IEEE, hlm. 13–22, DOI 10.1109/VISSOF.2002.1019790 — collaboration diagram | Pilih varian notasi yang lebih ringkas (R2-07); uji diagram dengan dua jenis tugas terpisah: konfirmasi vs temukan-kesalahan (R2-08). |
| P4 Otero & Dolado (2004), *Information and Software Technology* 46(1):35–53, DOI 10.1016/S0950-5849(03)00108-3 | Perilaku dinamis → utamakan sequence diagram (R2-09); satu sequence = satu alur (R2-10). |
| P5 Staron, Kuzniarz & Wohlin (2006), *Journal of Systems and Software* 79(5):727–742, DOI 10.1016/j.jss.2005.09.014 | Stereotip konsisten menambah klasifikasi (R2-11); stereotip tanpa makna dilarang (R2-12). |
| P6 Moody (2009), *IEEE TSE* 35(6):756–779, DOI 10.1109/TSE.2009.67 — "Physics of Notations" | 9 prinsip notasi visual (R2-13–R2-20). **Sifat bukti wajib dicantumkan:** sintesis teori berbasis literatur, BUKAN satu eksperimen atas diagram CBT. Angka panduan "enam" jenis simbol (R2-19) berasal dari ringkasan sekunder — JANGAN dikutip sebagai angka mutlak Moody (NA-13). |
| P7 Nugroho (2009), *Information and Software Technology* 51(12):1670–1685, DOI 10.1016/j.infsof.2009.04.007 | Jangan hilangkan detail kunci (R2-21); kesan "terasa jelas" ≠ pemahaman (R2-22); detail cukup meningkatkan KEBENARAN pemahaman — klaim kecepatan baca DITARIK karena replikasi tidak signifikan (R2-23, NA-15). |
| P8 Störrle (2011), VL/HCC 2011, hlm. 135–142, DOI 10.1109/VLHCC.2011.6070390 | Audit layout sebagai PAKET (persilangan + belokan + kesejajaran + jarak + keterbacaan label), bukan satu kriteria (R2-24). |
| P9 Störrle (2016), *Software & Systems Modeling* 17(1):115–134, DOI 10.1007/s10270-016-0529-x | Batasi ukuran tiap diagram; pecah diagram besar (R2-25). **LARANGAN MENGARANG ANGKA:** angka "ukuran optimal" TIDAK dikutip karena tidak terverifikasi dari abstrak (R2-26, NA-11, NA-12). |
| P10 Cruz-Lemus et al. (2011), *Information and Software Technology* 53(12):1391–1403, DOI 10.1016/j.infsof.2011.07.002 | Stereotip pada sequence membantu pemahaman, terutama bila pembaca tidak familiar dengan domain (R2-27, R2-28). Nuansa wajib: eksperimen tunggal pendahulu (Genero et al. 2008) hanya "kecenderungan lemah, tidak signifikan" — bukti kuat berasal dari family of experiments + meta-analisis 2011. |
| Kandidat yang ditelusuri tetapi TIDAK dihitung sebagai paper terverifikasi | Ambler (2005) — buku panduan praktisi, bukan jurnal; "Definition of the Criteria for Layout of the UML Use Case Diagrams" — venue/peer-review tidak terverifikasi; Glezer et al. (2005) — hanya sitasi; Gemino & Parker (2009) — hanya sitasi; Purchase et al. (2003) bab buku — temuan layout tidak berhasil diambil. |

### 1.3 R3 — Konvensi skripsi Indonesia (tingkat bukti 3–4)

- **Dokumen:** `docs/uml/riset/R3-skripsi.md` (17 kueri pencarian web, 17 Okt 2026).
- **Temuan utama (batas penting):** tidak ada satu pun pedoman kampus yang
  ditemukan menetapkan notasi UML baku versinya sendiri sampai level urutan
  lifeline, penomoran pesan sequence, atau penulisan guard pada sequence.
  Yang mengatur notasi UML adalah standar OMG (§1), bukan pedoman kampus.
  Pedoman kampus mengatur: konsistensi teori↔pembahasan, set diagram,
  keterkaitan dengan proses bisnis, dan format naskah (caption, penempatan
  gambar). Jangan mengklaim satu konvensi UML sebagai "wajib skripsi"
  kecuali konvensi itu ada di pedoman kampus target.
- **Status sumber:** **OTORITATIF** = pedoman resmi perguruan tinggi (SK
  Rektor / buku pedoman resmi kampus), berlaku hanya untuk kampus
  penerbitnya — TIDAK ada satu pedoman nasional yang mengatur notasi UML
  untuk semua skripsi di Indonesia. Beberapa pedoman hanya terjangkau lewat
  salinan cermin (123dok / Scribd / SlideShare / Academia.edu) — dicatat
  eksplisit per sumber, dan salinan cermin TIDAK diperlakukan setara dengan
  dokumen resmi di domain kampus. Hanya A1 (UBSI) yang diakses lewat domain
  resmi kampus.
- **Isi yang disintesis:** **12 aturan otoritatif SK-01–SK-12** (UBSI,
  STMIK Nuri, STMIK Nusa Mandiri Jakarta, STMIK Pelita Nusantara Medan,
  Universitas Ciputra, Universitas Indonesia) + **11 praktik literatur /
  bahan ajar / contoh SK-L01–SK-L11** (Sholiq 2006; Dharwiyanti & Wahono
  2003; skripsi UNPAS; modul Politeknik Telkom; Henderi 2008; dll) yang
  TIDAK BOLEH diklaim sebagai "aturan skripsi".

### 1.4 Dokumen aturan + hierarki bukti

- **Dokumen aturan:** `docs/uml/RISET-STANDAR-UML.md` — §1 WAJIB OMG (75),
  §2 kualitas jurnal (R2-01–R2-28), §3 konvensi skripsi (SK-01–SK-12 +
  SK-L01–SK-L11), §4 BUKAN ATURAN (NA-01–NA-25), §5 tiga koreksi premis,
  §6 hierarki bukti + 15 konflik antar-sumber yang diselesaikan (K-01–K-15).
- **Hierarki bukti (§6, mengikat semua tim):**

| Tingkat | Sumber | Kekuatan |
|---|---|---|
| 1 (tertinggi) | Spek formal OMG UML 2.5.1, formal/2017-12-05 | Mengikat notasi & semantik UML. Pelanggaran = pelanggaran formal. |
| 2 | Jurnal peer-reviewed (10 paper R2) | Bukti kualitas / keterbacaan. Rekomendasi, bukan pelanggaran formal OMG. Tidak boleh menciptakan "aturan OMG" baru. |
| 3 | Pedoman kampus resmi (SK Rektor / buku pedoman resmi) | Mengikat naskah skripsi di kampus penerbitnya saja (format, set diagram, konsistensi). Tidak mengatur notasi UML. |
| 4 (terendah) | Literatur / bahan ajar / contoh | Praktik baik / kebiasaan. Boleh dipakai bila dikutip jujur sebagai literatur/contoh. TIDAK BOLEH diklaim sebagai "aturan skripsi" atau "aturan OMG". Di bawah ini: buku panduan praktisi (mis. Ambler 2005) dan salinan cermin pedoman (bukan dokumen resmi). |

- **Kaidah pemakaian:** temuan "pelanggaran formal" hanya boleh bersandar
  pada §1; temuan §2 ditulis "rekomendasi kualitas (R2-xx)"; temuan §3
  ditulis "ketidaksesuaian pedoman kampus (SK-xx, kampus penerbit)". Bila
  §2/§3 tampak bertentangan dengan §1, §1 menang. Bila pedoman kampus
  tidak mengatur sesuatu (mis. notasi UML), kekosongan itu adalah hasil
  riset, bukan celah yang boleh diisi dengan mengarang aturan.

## 2. Perubahan per divisi

### 2.0 Konteks — Audit OMG sebelumnya (F1–F5, bukan fase industri)

Sebelum fase industri, audit OMG memeriksa file terhadap spesifikasi dan
menghasilkan **0 perubahan** (`fix_omg_audit_report.json`: `applied: []`,
21 butir `skipped` — semuanya sudah benar). Yang diperiksa dan dinyatakan
sudah benar: F1 condition Extend (`attempt diblokir`, `import file
dipilih`); F2 multiplisitas Class (13 ujung sudah `0..1`, 1 ujung `*`, 2
ujung `0..1`); F3 guard (`[pindah tab / blur]`, `[cek waktu / submit]`);
F4 `broadcastAttempt` sudah `asynchCall`; F5 reply `200 password berhasil
diubah` sudah ada. Audit ini penting karena menetapkan baseline
`d310d2c4…` yang masuk fase industri — dan karena F2 menjadi dasar
koreksi K7 (§2.1).

### 2.1 Divisi Audit OMG + Industri (20 orang) — K7 + K4

- **Dokumen:** `docs/uml/AUDIT-INDUSTRI.md`; **skrip:**
  `docs/uml/fix_audit_industri.py`; **laporan:**
  `docs/uml/fix_audit_industri_report.json`; **backup:**
  `cbt-smk-v2.mdj.pre-audit-industri.bak`.
- **SHA:** `d310d2c4…` → `cc749b66…` (2.194.431 → 2.198.375 byte).
- **Validator** `docs/uml/validate-mdj-real.py`: **LOLOS, 0 pelanggaran**
  (dijalankan sesudah perbaikan; run ke-2 script: 0 perubahan — idempotent
  terbukti).
- **Total edit: 26** (19 kardinalitas ERD + 7 documentation diagram).
  Total temuan yang sengaja TIDAK diubah: 6 kelompok (A1, K1, K2, K3, K5,
  K6 — lihat §4) + 2 observasi (K8, K9) + 2 butir naskah (SK-02, SK-10).

#### K7 — Kardinalitas ERD (19 relasi)

Sebelum: 51× `0..* : 1` + 3× `0..* : 1..1` (54 relasi).
Sesudah: 36× `0..* : 1` + 16× `0..* : 0..1` + 2× `0..1 : 1` (54 relasi).
Konvensi baca (ditetapkan dari orientasi field StarUML: `end1` = induk,
`end2` = anak pemegang FK): gabungan = `<anak-per-induk> :
<induk-per-anak>`; `anak-per-induk` = `0..1` bila kolom FK UNIQUE, selain
itu `0..*`; `induk-per-anak` = `1` bila kolom FK NOT NULL, selain itu
`0..1`. Yang diubah: `ERDRelationship.cardinality` + `end2.cardinality`.
Label view relasi ERD hanya menampilkan NAMA relasi (StarUML me-render
kardinalitas dari field model), jadi tidak ada teks view yang perlu diubah.

Ground truth: flag `nullable`/`unique` 54 kolom FK di ERD dicocokkan dengan
migrasi `backend/migrations/0001–0012` — **0 perbedaan** (50 definisi
`CREATE TABLE` + 4 definisi `ALTER TABLE` di 0006/0009).

| Relasi | Sebelum → Sesudah | Dasar |
|---|---|---|
| `berita_acara.jadwal_id -> jadwal_ujian` | `0..* : 1..1` → `0..1 : 1` | `jadwal_id UNIQUE` (migrasi 0005); cocok dengan Class `(JadwalUjian:1, BeritaAcara:0..1)` |
| `nilai_akhir.attempt_id -> attempt` | `0..* : 1..1` → `0..1 : 1` | `attempt_id UNIQUE` (migrasi 0005); cocok dengan Class `(Attempt:1, NilaiAkhir:0..1)` |
| `nilai_akhir.jadwal_id -> jadwal_ujian` | `0..* : 1..1` → `0..* : 1` | Normalisasi penulisan (`1..1` ≡ `1`); `jadwal_id` NOT NULL, tidak UNIQUE |
| 16 relasi FK nullable (mis. `audit_log.actor_id`, `siswa.user_id`, `guru.user_id`, `bank_soal.guru_id`, `soal.kompetensi_id`, `soal.kisi_kisi_id`, `jadwal_ujian.event_id`, `jadwal_ujian.ruangan_id`, `mapel.jurusan_id`, `kisi_kisi.kompetensi_id`, + 6 kolom audit `dibuat_oleh`/`dinilai_oleh`/`diupload_oleh`/`diproses_oleh`/`actor_id`) | `0..* : 1` → `0..* : 0..1` | Kolom FK nullable (SET NULL); cocok dengan multiplisitas Class hasil koreksi F2 audit OMG sebelumnya |

#### K4 — Narasi + legenda diagram (R2-18)

`documentation` ditambahkan pada 7/7 diagram (sebelumnya kosong): Use Case
608 karakter, Class 499, Activity 2a 511, Activity 2b 323, Sequence 3a 547,
Sequence 3b 602, ERD 578. Isi: narasi alur + legenda simbol + catatan yang
relevan (mis. konvensi baca kardinalitas ERD, status ExtensionPoint, status
signature pesan). Ini teks dokumentasi non-semantik; tidak mengubah elemen
model apa pun.

#### Yang diperiksa dan BERSIH (Audit Industri)

- **Use Case (UC-01–UC-12):** 19 UseCase bernama; 6 Actor bernama,
  asosiasinya hanya ke UseCase dan biner — 39/39 Actor↔UseCase; tidak ada
  asosiasi antar dua UseCase; tidak ada siklus include (3 include: UC08→UC09,
  UC05→UC03, UC11→UC10); arah include/extend + `condition` benar; UC-09
  kecuali `extensionLocation` (A1, §4).
- **Class (CL-01–CL-12):** 46 asosiasi, semuanya ≥ 2 memberEnd, tidak ada
  agregasi yang dimodelkan, semua navigable dua arah, 0 duplikat asosiasi,
  semua ujung punya multiplisitas tanpa kurung siku, 46/46 asosiasi
  bernama, 229 atribut semua bertipe + `public`, 11 operasi berformat benar.
- **Sequence (SEQ-01–SEQ-26) + CombinedFragment (CF-01–CF-13):** 32 pesan:
  sort hanya `synchCall` (19), `reply` (12), `asynchCall` (1:
  `broadcastAttempt`); tidak ada `createMessage`/`deleteMessage`; tidak ada
  `signature` pada pesan apa pun (SEQ-09/SEQ-10 tidak terlanggar — lihat
  K2); tidak ada CombinedFragment (percabangan = alur linear naratif, sah);
  tidak ada ExecutionSpecification/OccurrenceSpecification sebagai elemen
  model (konvensi representasi StarUML — lihat §5.4); semua pesan horizontal
  atau menurun (30 horizontal, 2 self-message persegi kecil menurun);
  lifeline 10/10 beridentifikasi tidak kosong; tidak ada selector lifeline.
  SEQ-14 (kepala panah per sort) tidak dapat diverifikasi dari JSON —
  StarUML me-render-nya dari `messageSort` secara otomatis; tidak diklaim
  sebagai temuan.
- **Activity (AC-01–AC-12):** 2a: Initial `Mulai` in=0/out=1, semua outgoing
  ControlFlow; 3 Decision masing-masing in=1/out=2 dengan guard di KEDUA
  cabang; 2b: Decision in=1/out=2 dengan guard `belum semua`/`semua
  terkoreksi`; tidak ada guard `else` ganda; tidak ada Merge/Fork/Join;
  tidak ada simbol gabungan; FinalNode out=0; tidak ada guard berkurung
  ganda (guard tersimpan tanpa kurung, StarUML menambah `[ ]` saat render);
  tidak ada Decision dengan 2 incoming.
- **ERD (konsistensi internal + konvensi proyek, NA-10):** 35 entitas ↔ 34
  class + 1 tabel join `role_permissions`; 44 pasangan relasi cocok dengan
  asosiasi Class; 54 relasi FK semuanya bernama
  (`tabel_anak.kolom -> tabel_induk`); kardinalitas kini presisi (K7);
  `identifying: false` pada semua relasi (konsisten — tidak ada FK yang
  menjadi bagian PK anak, kecuali PK komposit tabel join yang dimodelkan
  sebagai entitas sendiri).
- **Kualitas jurnal (§2) — yang bersih:** R2-02 belokan (semua segmen
  axis-aligned, 0 diagonal); R2-05 satu gaya notasi; R2-07 tanpa dekorasi
  bertele-tele; R2-09 perilaku dinamis memakai sequence; R2-10/R2-16 satu
  sequence = satu alur; R2-11/R2-12 stereotip yang ada bermakna dan
  konsisten; R2-13 satu simbol = satu makna; R2-14 dapat dibedakan
  hitam-putih; R2-15 arah panah sesuai arah pesan/alur; R2-19 ragam jenis
  simbol minimal per diagram (Use Case 5, Class 2, Activity ≤ 4, Sequence 3,
  ERD 2 jenis); R2-20 label lengkap Bahasa Indonesia + legenda di
  documentation diagram; R2-21 detail kunci (multiplisitas lengkap di
  Class, label pesan lengkap di Sequence — kecuali operasi, K2).

### 2.2 Divisi Sequence (30 orang) — rapikan Sequence 3a + 3b

- **Skrip:** `docs/uml/fix_sequence_industri.py`; **laporan:**
  `docs/uml/fix_sequence_industri_report.json`; **backup:**
  `cbt-smk-v2.mdj.pre-sequence-industri.bak` (SHA awal `cc749b66…`).
- **SHA:** `cc749b66…` → `158d9dd1…`.
- **HANYA geometri view; semantik 0** (snapshot semantik objek bukan-View
  identik; diff guard: hanya `UMLSeqLifelineView.left`,
  `UMLSeqMessageView {points, left}`, `EdgeLabelView {left, top}` yang boleh
  berubah).

#### Perubahan

1. **PITCH lifeline 180 → 165** (celah kotak tersimpan 40 → 25). Lifeline
   pertama tetap `left=100`; berikutnya +165. Alasan: celah panah = pitch −
   lebar-render-kotak, jadi celah HANYA bisa mengecil lewat pitch. Lebar
   tersimpan 140 DIPERTAHANKAN seragam — (a) di bawah `max(width,
   fit-teks)`, mengecilkan width tersimpan TIDAK PERNAH mengecilkan celah;
   (b) lebar-fit-teks tiap lifeline tidak terukur dari `.mdj`, jadi width
   per-lifeline berisiko membuat celah TIDAK seragam; (c) width seragam 140
   menjaga celah seragam. Batas aman: kalibrasi R10 memberi fit-teks ≤ 146
   (file TES) → celah ter-render ≥ 165−146 = 19 pada kasus terburuk;
   perkiraan dari teks lifeline file ini (baris terpanjang ~18 karakter)
   memberi celah ter-render ~25. Lebar total diagram: 3a 680 → 635 (−45),
   3b 1040 → 965 (−75).
2. **Points pesan + left view pesan** dihitung ulang ke pusat lifeline BARU
   (konvensi R9d, center-to-center; points X diabaikan StarUML tapi
   disimpan konsisten). Self-message: loop 30×20 menempel di garis
   lifeline-nya (pusat baru). Y pesan TIDAK berubah.
3. **Semua EdgeLabelView** (ber-teks maupun kosong) dihitung ulang
   `left/top`-nya dari `(edgePosition, distance, alpha)` + points BARU
   memakai aturan kalibrasi ronde 8 untuk Sequence (R = titik tengah
   segmen tengah; C tanpa offset +1; `left/top` = round-half-up). Ini
   sekaligus menormalkan 31 nilai `left/top` lama yang memakai pembulatan
   floor (−1px; `left/top` DIABAIKAN saat render) dan 2 label kosong
   self-message yang posisinya basi. `distance`/`alpha` TIDAK diubah.
   Total label yang dihitung ulang: **96**.

#### Verifikasi geometri (gagal = file TIDAK ditulis)

Clearance label teks vs kotak kepala ≥ 8; label-label tanpa tumpang tindih
dan gap ≥ 8; label vs panah pesan LAIN ≥ 8; label self vs segmen vertikal
loop-nya ≥ 8; label vs panahnya sendiri = 3,5 (non-self, by-design) / ≥ 8
(self). 0 segmen diagonal, 0 persilangan antar-pesan, Y pesan tak berubah,
lebar lifeline tak berubah, clearance label pesan pertama ≥ 8.
Hasil ukur: 3a clearance label–kotak min 8,5; gap label–label min 27,0. 3b
clearance label–kotak min 24,5; gap label–label min 47,0.

#### Yang SENGAJA TIDAK diubah (dengan alasan)

- Urutan / isi / pengirim / penerima / jenis pesan: semantik bisnis.
- Self-message TIDAK dipindah ke akhir: di 3a ia pesan ke-5 dari 16
  (`verify Argon2id…`), di 3b pesan ke-10 dari 16 (`cek deadline lewat…`)
  — secara semantik BUKAN pesan terakhir. Memindahkannya = mengubah urutan
  alur. Rendering StarUML yang menaruhnya terakhir adalah BATASAN TOOL yang
  dilaporkan, bukan cacat `.mdj` (lihat §5.1).
- Activation bar: TIDAK ADA (0 `UMLActivationView`) — tidak ada yang
  disejajarkan; menambahkannya berarti membuat elemen model baru (keputusan
  pemodelan, di luar "rapikan").
- CombinedFragment / guard: TIDAK ADA di 3a/3b (percabangan dimodelkan
  sebagai alur linear naratif — sah per audit, CF vakum). Menambah fragment
  alt/opt = mengubah struktur interaksi.
- `distance`/`alpha` label: TIDAK diubah. Jarak label ke panahnya sendiri
  3,5px (offset 10 − setengah tinggi label 6,5) adalah by-design dari ronde
  sebelumnya: label TIDAK menimpa panahnya.
- Label yang melewati garis putus-putus lifeline: praktik standar sequence
  diagram (diputuskan G5; render native StarUML sendiri menaruh label
  melewati garis lifeline karena label lebih lebar dari celah
  antar-lifeline). Yang dihitung sebagai cacat dan diverifikasi: label vs
  KOTAK kepala (padat) dan label vs panah.
- `head`/`tail` `$ref`, model pesan/lifeline, diagram lain: tidak disentuh.

### 2.3 Divisi Diagram Lain (20 orang) — verifikasi Use Case, Class, ERD, Activity

- **Skrip:** `docs/uml/fix_diagram_lain_industri.py` (HANYA memverifikasi dan
  menulis laporan — file `.mdj` TIDAK disentuh: tidak ada backup baru,
  tidak ada penulisan `.mdj`); **laporan:**
  `docs/uml/fix_diagram_lain_industri_report.json`.
- **SHA:** `158d9dd1…` → `158d9dd1…` (**0 perubahan**, `mdj_unchanged: true`).
- **Hasil:** setelah audit ulang programatis terhadap
  RISET-STANDAR-UML.md + AUDIT-INDUSTRI.md, TIDAK ADA perbaikan yang
  memenuhi SEMUA syarat sekaligus (didukung sumber riset + tidak mengubah
  semantik + tidak merusak kerapian garis/label ronde sebelumnya).

#### Yang diverifikasi

- **UC:** UC-01–UC-12 (19 UseCase bernama, 6 Actor bernama, 39 asosiasi
  Actor↔UseCase, tanpa siklus include, arah include/extend, condition
  Extend; A1 `extensionLocation` tetap DITUNDA — butuh keputusan user soal
  nama ExtensionPoint).
- **Class:** CL-01–CL-12 (46 asosiasi bernama, multiplisitas lengkap,
  navigable dua arah, tanpa agregasi; 229 atribut bertipe + public; 11
  operasi; K2/K3/K5/K6 tetap DITUNDA sesuai audit).
- **ERD:** 35 entitas, 231 kolom, 54 relasi; K7 diverifikasi ulang
  (kardinalitas vs nullable/unique FK: **0 selisih**); label nama == nama
  model 54/54; 53/54 label tersimpan == prediksi aturan rata-rata, 1/54
  (`jawaban.attempt_id`) == prediksi aturan titik-tengah-bawah (perbaikan
  R10) — AMBIGUITAS ATURAN RENDER dicatat, BUKAN diperbaiki massal (lihat
  §5.4).
- **Activity 2a/2b:** AC-01–AC-12 (Initial in=0, Final out=0, Decision
  in=1/out=2, guard tanpa kurung ganda, tanpa Merge/Fork/Join); R2-18
  documentation ada di 7/7 diagram.
- **Include/Extend:** 3 include, 2 extend dengan condition
  (`attempt diblokir`, `import file dipilih`); 0 extension point; 0 siklus
  include.

### 2.4 Perbaikan Putaran 1 (T1+T2) — ERD end1 + titik masuk garis

- **Skrip:** `docs/uml/fix_t1t2_industri.py`; **laporan:**
  `docs/uml/fix_t1t2_industri_report.json`; **backup:**
  `cbt-smk-v2.mdj.pre-t1t2.bak` (SHA awal `158d9dd1…`).
- **SHA:** `158d9dd1…` → **`c068a179…` (FINAL)**.
- **Hanya 17 edit** (16× satu field + 1× points satu view); diff guard +
  verifikasi akhir (T1 16× end1, T2 points + sibling, `$ref`).

#### T1 — 16× `end1.cardinality = '0..1'` pada relasi ERD sisi induk `0..1`

Sebelumnya `end1` TIDAK punya field `cardinality`, sehingga StarUML memakai
default konstruktor `'1'` dan me-render ujung induk `'||'` (seperti `'1'`),
padahal model sudah benar (`'0..1'` sejak K7).
Bukti mekanisme (StarUML 7.1.0, `app.asar`): `ERDRelationshipEnd`:
constructor `this.cardinality = '1'`; `ERDRelationshipView.update()`:
`switch (this.model.end1.cardinality)` — `'0..1'` →
`ES_CROWFOOT_ZERO_ONE` (o|), `'1'` → `ES_CROWFOOT_ONE` (||).
Bukti visual A/B di StarUML asli (screenshot di `~/workspace/staruml/t1t2/`):
salinan KONTROL me-render ujung induk `audit_log.actor_id` sebagai `'||'`;
salinan EKSPERIMEN (hanya field ini yang ditambah) me-render `'o|'` —
relasi lain di tangkapan yang sama tetap `'||'`. Jadi StarUML MENGHORMATI
`end1.cardinality`.
Yang diubah HANYA `end1.cardinality`; `relationship.cardinality` dan
`end2.cardinality` sudah benar sejak audit K7 dan tidak disentuh. 38 relasi
lain (sisi induk `'1'`) sengaja TIDAK diberi field eksplisit (nilai default
`'1'` sudah benar) agar diff minimal.

Ke-16 relasi: `audit_log.actor_id`, `guru.user_id`, `siswa.user_id`,
`bank_soal.guru_id`, `usulan_unblock.diproses_oleh`,
`berita_acara.dibuat_oleh`, `ukk_nilai.dinilai_oleh`,
`ukk_berkas.diupload_oleh`, `backup_meta.dibuat_oleh`,
`sod_overrides.actor_id` (semuanya → `users`); `mapel.jurusan_id → jurusan`;
`kisi_kisi.kompetensi_id → kompetensi`; `soal.kompetensi_id → kompetensi`;
`soal.kisi_kisi_id → kisi_kisi`; `jadwal_ujian.event_id → event`;
`jadwal_ujian.ruangan_id → ruangan`.

#### T2 — Geser titik masuk garis `berita_acara.dibuat_oleh -> users`

Titik masuk di tepi atas entitas `berita_acara` dari x=1267 ke x=1317
(+50px). Sebelumnya dua simbol crow's foot (`jadwal_id` x=1277 +
`dibuat_oleh` x=1267) hanya 10px apart sehingga lingkaran `'o'`-nya
bertumpuk dan sulit dibaca pada 100%. Sesudahnya jarak 40px: `'o|'`
(`jadwal_id`) dan `'o<'` (`dibuat_oleh`) terbaca terpisah (screenshot A/B
di `~/workspace/staruml/t1t2/`). Efek samping geometri: persilangan garis
ini berkurang 1 (vertikal baru x=1317 tidak lagi memotong horizontal
`jadwal_id` y=1312; horizontal y=1268 memotong 22 vertikal yang sama
seperti sebelumnya — 0 persilangan baru). Hanya field `points` SATU view
yang diubah; label, entitas, dan garis lain tidak disentuh. View tetangga
(`berita_acara.jadwal_id`) diverifikasi tidak berubah.

### 2.5 Perbaikan Putaran 2 — yang DITOLAK + alasannya

Putaran 2 = verifikasi pasca-T1T2 oleh Divisi Diagram Lain (skrip §2.3,
`changes: 0`). Dua usulan perbaikan yang sempat dipertimbangkan DITOLAK
dengan bukti terukur — keduanya didokumentasikan di
`fix_diagram_lain_industri_report.json` sebagai `action: "TIDAK DIGESER"`:

| # | Usulan yang ditolak | Alasan (bukti terukur) |
|---|---|---|
| 1 | Geser label guard Activity 2a agar gap `pindah tab / blur` vs `siswa submit` mencapai 8px (saat ini 5,0px) | Slot vertikal antara bawah label `siswa submit` (773,5) dan segmen edge `siswa submit` (y=802) hanya **28,5px < 13 (tinggi label) + 8 + 8 = 29px**. Penempatan ronde 8 (5px ke label, 9,5px ke edge, 9,46px ke kotak) adalah **optimum slot tersebut** — menggeser label ke mana pun memperburuk salah satu sisi. DITOLAK: tidak ada penempatan yang memenuhi 8px kedua sisi. |
| 2 | Geser label `«include»` Use Case yang tampak menimpa elips UC04 | Tumpang tindih hanya pada **kotak pembatas (bounding box)**, bukan pada elips. Bukti: elips UC04 kotak (620,280)–(880,340), pusat (750,310), rx=130, ry=30; titik sudut label (858,282): `((858−750)/130)² + ((282−310)/30)² = 1,56 > 1` → titik di **LUAR elips**: tidak ada tumpang tindih visual. DITOLAK: menggeser label hanya memperbaiki kotak pembatas yang tidak terlihat. |

Selain kedua penolakan di atas, Putaran 2 juga mencatat satu ambiguitas
terbuka yang TIDAK diperbaiki massal (lihat §5.4): aturan render label
"rata-rata dua titik tengah" vs "titik tengah bawah" memprediksi posisi
berbeda untuk edge ber-n GENAP — perlu SATU eksperimen decisive di StarUML
asli sebelum mengubah 100+ label.

## 3. Aturan yang diterapkan

### 3.1 Aturan OMG (§1) yang ditegakkan

- **K7 (ERD, via konsistensi internal NA-10 + multiplisitas Class):**
  kardinalitas presisi sesuai nullable/UNIQUE FK; normalisasi `1..1` → `1`
  (CL-03: `1` ≡ `1..1`).
- **T1 (ERD, render):** `end1.cardinality` eksplisit agar ujung induk
  ter-render `o|` bila `0..1` (mekanisme StarUML, §5.3).
- **UC-01–UC-12, CL-01–CL-12, SEQ-01–SEQ-26, CF-01–CF-13, AC-01–AC-12:**
  diverifikasi bersih (lihat §2.1 "Yang diperiksa dan BERSIH"); tidak ada
  pelanggaran formal OMG yang tersisa kecuali A1 (§4).

### 3.2 Aturan kualitas jurnal (§2) yang diterapkan

- **R2-18 (dual coding):** documentation + legenda pada 7/7 diagram (K4).
- **R2-21 (detail kunci):** multiplisitas lengkap di Class; label pesan
  lengkap di Sequence; kardinalitas presisi di ERD (K7).
- **R2-01/R2-02 (layout):** persilangan dipertahankan ortogonal (0 segmen
  diagonal di semua diagram); T2 mengurangi 1 persilangan ERD tanpa
  menambah persilangan baru; pitch Sequence dikecilkan agar celah panah
  lebih rapat dan seragam.
- **R2-17 (integrasi antar diagram):** penamaan identik ditegakkan lewat
  dokumentasi (K5 ditunda — lihat §4, karena penggantian nama = perubahan
  model).
- **R2-13/R2-14/R2-15:** satu simbol = satu makna; pembeda berupa teks +
  gaya garis (bukan warna saja); arah panah sesuai arah pesan/alur —
  diverifikasi bersih.
- **R2-24 (layout sebagai paket):** laporan ini TIDAK mengklaim "layout
  sudah rapih" hanya dari satu aspek — persilangan yang tersisa (K1)
  dilaporkan apa adanya (lihat §4).

### 3.3 Konvensi skripsi (§3) yang diterapkan

- **R2-18 + SK-11:** documentation diagram mencantumkan referensi notasi
  yang dipakai (OMG UML 2.5.1) — naskah wajib konsisten dengan SATU
  referensi yang dikutip.
- **SK-05/SK-06/SK-08, SK-10, SK-02:** aturan naskah-level — tidak dapat
  diterapkan ke `.mdj`; dirangkum sebagai kewajiban naskah di §6.

## 4. Aturan yang SENGAJA tidak ditegakkan (§4 BUKAN ATURAN)

Bagian ini sama pentingnya dengan §3. Setiap butir di bawah DILARANG
dijadikan temuan / pelanggaran. Bila sebuah temuan audit bersandar pada
salah satu butir ini, temuan itu harus ditarik. (Ringkasan dari
RISET-STANDAR-UML.md §4, NA-01–NA-25.)

| ID | Klaim yang DILARANG ditegakkan | Alasan + sumber yang membatalkannya |
|---|---|---|
| NA-01 | Penomoran pesan Sequence `1`, `1.1`, `2` sebagai persyaratan OMG / aturan skripsi | Penomoran berurutan hanya untuk **Communication Diagram** (SEQ-17: §17.4.4.1, §17.9.1.3, §17.9). Sintaks label Message Sequence tidak memuat nomor urut. Aturan penomoran dalam Dharwiyanti & Wahono (2003) berlaku untuk Collaboration/Communication, bukan sequence — menempelkannya ke sequence adalah **salah kutip**. |
| NA-02 | "Setiap `synchCall` WAJIB punya reply" sebagai pelanggaran formal | Tidak ada constraint di §17.12.18.7 yang mewajibkannya; §17.2.3.4 hanya berkata "Typically". Tanpa reply = tidak lengkap, bukan pelanggaran. |
| NA-03 | "Setiap cabang Decision WAJIB punya guard" sebagai pelanggaran formal | Guard per edge opsional; edge tanpa guard = selalu true (AC-03, AC-01). Yang mengikat: satu token paling banyak satu cabang + anjuran determinisme. |
| NA-04 | Urutan lifeline WAJIB Actor → Boundary → Control → Entity | **Tidak ditemukan sumber otoritatif.** Klasifikasi BCE terbukti dipakai, tetapi tidak ada sumber yang menetapkan urutan kiri→kanan itu sebagai kewajiban. OMG juga tidak menetapkan urutan lifeline tersebut. |
| NA-05 | Batas angka jumlah lifeline ("maksimal N lifeline") | Tidak ada paper terverifikasi di R2 yang menyatakan batas angka jumlah lifeline. Yang didukung hanya prinsip umum: kurangi kompleksitas / pecah diagram, tanpa angka. |
| NA-06 | "Label pesan HARUS memakai nama operasi" sebagai format penamaan yang diwajibkan jurnal | Tidak ada paper di R2 yang menyatakan kalimat persis ini. R2-21 hanya menuntut label pesan yang lengkap dan bermakna, bukan format penamaan tertentu. (OMG SEQ-09 mewajibkan nama Message = nama NamedElement yang dirujuk signature — itu aturan OMG §1, bukan aturan jurnal.) |
| NA-07 | "Diagram harus muat satu halaman" | Tidak ada paper di R2 yang menyatakan kalimat persis ini. Pedoman Ciputra justru membolehkan gambar terlalu besar memakai landscape / A3 dilipat / diperkecil ke A4 (SK-08); yang harus satu halaman adalah gambar + judulnya agar tidak terpotong (SK-06). |
| NA-08 | Larangan mutlak garis diagonal / garis lengkung | P1 hanya menyatakan efek grid ortogonal tidak signifikan dan bends berpengaruh lebih kecil; larangan mutlak berasal dari buku panduan praktisi (Ambler, *The Elements of UML 2.0 Style*, 2005) — bukan paper jurnal peer-reviewed — dan tidak boleh dikutip sebagai bukti empiris jurnal. Yang didukung P1: minimalkan belokan (R2-02). |
| NA-09 | Guard `[[…]]` kurung ganda sebagai notasi yang benar / wajib | **Tidak ditemukan sumber apa pun.** Tidak ada sumber yang menulis guard dengan kurung siku ganda; ini artefak rendering tool, bukan notasi. Guard yang benar: satu pasang kurung siku (CF-11, AC-02). |
| NA-10 | Aturan notasi ERD (crow's foot, kunci primer / kunci tamu) dari OMG | **ERD bukan diagram UML 2.5.1.** Pencarian teks penuh spesifikasi tidak menemukan "Entity-Relationship" / ERD; taksonomi diagram spesifikasi tidak memuatnya (CL-13 TIDAK TERVERIFIKASI). ERD hanya diaudit untuk (a) konsistensi internal terhadap Class Diagram dan (b) konvensi proyek sendiri — BUKAN terhadap OMG. |
| NA-11 | Klaim angka spesifik ("ukuran optimal", "maksimal 20 elemen") atas nama paper yang tidak terverifikasi dari naskah lengkap — khususnya atas nama Störrle (P9) | Angka "ukuran optimal" Störrle TIDAK dikutip di R2 karena tidak berhasil diverifikasi dari abstrak yang dapat diakses (R2-26). Mengarang angka atas nama paper ini DILARANG. |
| NA-12 | "Maksimal N elemen per diagram" (angka apa pun, atas nama siapa pun) | Tidak ada angka yang dikutip di R2. Yang didukung: batasi ukuran dan pecah diagram besar (R2-25), tanpa angka. |
| NA-13 | "Maksimal 6 jenis simbol" sebagai angka mutlak dari Moody (P6) | Angka panduan "enam" pada R2-19 berasal dari **ringkasan sekunder** atas Moody, bukan verifikasi langsung naskah Moody. Yang didukung: jaga ragam jenis simbol seminimal mungkin. |
| NA-14 | "Sudah menempel grid ortogonal sehingga diagram bagus" sebagai bukti kualitas | Efek grid ortogonal dan sudut minimum terbukti TIDAK signifikan secara statistik pada P1 (R2-04). Grid hanya alat bantu perataan. |
| NA-15 | "Detail lebih tinggi selalu membuat diagram lebih cepat dibaca" | Efek efisiensi tidak terkonfirmasi kuat pada replikasi Nugroho (R2-23). Yang didukung: detail cukup meningkatkan KEBENARAN pemahaman. |
| NA-16 | "Kepala lifeline hasil create mulai tepat di titik pesan create" sebagai temuan formal | TIDAK TERVERIFIKASI di R1 (SEQ-27): tidak ada kalimat eksplisit di teks spesifikasi yang diakses. Yang terverifikasi hanya arti `createMessage` (SEQ-07) dan gaya garisnya (SEQ-14). |
| NA-17 | "Ujung lifeline yang dihancurkan ditandai tanda X" sebagai temuan formal | TIDAK TERVERIFIKASI di R1 (SEQ-28): tidak ditemukan di teks spesifikasi yang diakses (kemungkinan hanya pada gambar contoh). Yang terverifikasi: `deleteMessage` berakhir di DestructionOccurrenceSpecification (SEQ-08) dan tidak ada event sesudahnya (SEQ-23). |
| NA-18 | Nama kelas PascalCase / kata benda tunggal sebagai "aturan skripsi" | TIDAK DITEMUKAN SUMBER OTORITATIF (Indonesia) di R3 (D15). Sumber Inggris (vendor tool) menyebut nama kelas biasanya huruf kapital + kata benda tunggal, tetapi itu bukan pedoman skripsi Indonesia. |
| NA-19 | Satu set diagram universal untuk semua skripsi (mis. "wajib 7 diagram") | TIDAK DITEMUKAN di R3 (D22). Semua pedoman yang ditemukan menetapkan set secara kondisional dan berbeda-beda antar kampus (SK-02, SK-03). |
| NA-20 | "Pedoman kampus selalu mengatur notasi UML" | SALAH (R3 D23). Contoh: pedoman UI (A6) hanya mengatur format naskah, tidak mengatur notasi UML. |
| NA-21 | Caption memakai angka Romawi bab ("Gambar III.1") sebagai aturan | VARIASI CONTOH, BUKAN ATURAN (R3 D6): ditemukan di satu skripsi (C2, UIN Ar-Raniry); kampus lain memakai "Gambar 3.1". Ikuti pedoman kampus target (SK-05). |
| NA-22 | Detail tipografi caption (mis. Times New Roman 10, spasi 1, rata kiri) sebagai aturan universal | Detail tersebut adalah aturan Universitas Ciputra saja, bukan aturan universal — kampus lain menetapkan detail berbeda (R3 A2; SK-12). |
| NA-23 | Menilai kualitas diagram dari kesan "terasa jelas/mudah" saja | Persepsi kejelasan TIDAK mencerminkan pemahaman sebenarnya pada model ber-detail rendah (R2-22, P7). Uji dengan pertanyaan pemahaman konkret. |
| NA-24 | Guard `[…]` pada sequence sebagai "aturan pedoman kampus Indonesia" | Hanya didukung standar OMG UML 2.5 secara umum (CF-11, AC-02) dan literatur untuk transisi statechart (SK-L09); TIDAK ditemukan pedoman kampus Indonesia yang mengaturnya (R3 D12). Jangan klaim sebagai aturan pedoman kampus. |
| NA-25 | "Notasi ini terasa lebih jelas" (intuisi) sebagai bukti kualitas notasi | Prediksi intuitif penulis hanya terkonfirmasi sebagian pada P2 (R2-06). Nilai dengan tugas konkret, bukan kesan sekilas. |

### Tiga koreksi premis (§5 RISET-STANDAR-UML.md)

1. **Penomoran pesan `1`, `1.1`, `2` BUKAN bagian Sequence Diagram**
   (SEQ-17; NA-01). Yang mengikat: urutan vertikal (atas → bawah) dan
   korespondensi pesan.
2. **Reply TIDAK WAJIB untuk setiap `synchCall`** (SEQ-05, SEQ-21; NA-02).
   SynchCall tanpa reply = model tidak lengkap / perlu ditanya, BUKAN
   pelanggaran formal.
3. **Guard TIDAK WAJIB di tiap cabang Decision** (AC-03; NA-03). Yang
   mengikat: satu-token-satu-cabang + anjuran determinisme; `else` paling
   banyak satu.

### Yang SENGAJA TIDAK dimasukkan ke dokumen aturan (tidak terverifikasi)

| Butir | Alasan |
|---|---|
| SEQ-27 — Kepala lifeline hasil create mulai tepat di titik pesan create | Tidak ada kalimat eksplisit di teks spesifikasi yang diakses. Lihat NA-16. |
| SEQ-28 — Ujung lifeline yang dihancurkan ditandai tanda X | Tidak ditemukan di teks spesifikasi yang diakses; kemungkinan hanya pada gambar contoh. Lihat NA-17. Bila perlu kepastian, periksa Figure 17.14 / 17.17 spesifikasi secara visual. |
| CL-13 — Aturan notasi ERD dari OMG | ERD bukan diagram UML 2.5.1; tidak ada di spesifikasi. Lihat NA-10. |

## 5. Temuan yang ditunda karena semantik (butuh keputusan user)

Enam kelompok temuan SENGAJA TIDAK diubah pada fase ini. Alasan bersama:
memperbaikinya berarti mengubah SEMANTIK BISNIS (relasi, atribut, alur,
pesan, klasifikasi, nama) — dan itu adalah keputusan pemodelan, bukan
perbaikan audit. Setiap butir dilengkapi rekomendasi konkret untuk user.

### A1 — Extend tanpa `extensionLocation` (UC-09, WAJIB OMG)

- **Bukti:** 2 relasi `UMLExtend` tidak punya `extensionLocation`; tidak ada
  elemen `UMLExtensionPoint` di model. Relasi: `UC18 Usul & Proses Unblock
  → UC07 Monitor Ujian Real-time` (condition `attempt diblokir`, id
  `VW3VFH11…`) dan `UC04 Import Soal → UC03 Kelola Bank Soal` (condition
  `import file dipilih`, id `K5XYTJYZ…`); tidak ada satu pun elemen
  `UMLExtensionPoint` di model (hanya compartment view kosong).
- **Alasan ditunda:** StarUML tidak menyimpan field `extensionLocation` pada
  serialisasi `UMLExtend` di file ini, dan nama extension point adalah
  keputusan pemodelan (penambahan elemen = semantik).
- **Rekomendasi:** di StarUML, tambahkan ExtensionPoint pada extended
  UseCase, lalu kaitkan: UC07 Monitor Ujian Real-time ← ExtensionPoint
  `attempt diblokir` (untuk Extend UC18, condition `attempt diblokir`);
  UC03 Kelola Bank Soal ← ExtensionPoint `import file dipilih` (untuk
  Extend UC04, condition `import file dipilih`). Nama yang diusulkan
  disalin dari condition yang sudah ada — bukan karangan baru.

### K1 — Persilangan garis (R2-01)

- **Angka:** Use Case 244, Class 259, ERD 406, Activity 2a 7, Activity 2b
  0, Sequence 3a 0, Sequence 3b 0 (metode: perpotongan interior
  antarsegmen dari edge yang berbeda; tidak termasuk titik ujung bersama
  dan segmen kolinear; angka Class/ERD/Activity 2a cocok dengan laporan
  fase line-neatness — metode terkalibrasi silang; Use Case 244 baru
  diukur pada audit ini). Semua persilangan yang tersisa adalah tegak
  lurus (rute ortogonal; 0 segmen diagonal di semua diagram).
- **Alasan ditunda:** (a) fase line-neatness sebelumnya secara sadar
  menukar persilangan dengan sudut 90°; (b) menurunkan ke 0 butuh redesain
  tata letak (pekerjaan besar, berisiko menggeser ratusan view);
  (c) R2-24 melarang mengklaim "layout sudah rapih" hanya dari satu aspek
  — laporan ini TIDAK mengklaim layout rapih.
- **Rekomendasi:** terima sebagai utang kualitas yang diketahui, atau
  tugaskan redesain tata letak khusus (bukan audit) bila penguji
  mempersoalkan. Catatan transfer P1 berlaku: eksperimen Purchase (1997)
  memakai graf abstrak, bukan diagram UML — penerapannya ke UML adalah
  ekstrapolasi yang masuk akal, bukan bukti langsung pada UML.

### K2 — Traceability pesan → operasi lemah (R2-21, SK-L10)

- **Bukti:** 32/32 pesan Sequence tanpa `signature`; 11 operasi pada 6/34
  class; 28 class tanpa operasi; operasi menyimpan signature sebagai string
  nama (mis. `verifyPassword(p): Boolean`) tanpa parameter terstruktur,
  sehingga korespondensi argumen SEQ-02/SEQ-10 tidak dapat diperiksa. Ini
  BUKAN pelanggaran formal OMG (field `signature` Message bersifat
  opsional; constraint SEQ-09 hanya berlaku bila signature ada).
- **Alasan ditunda:** menambahkan operasi/signature berarti mengarang
  semantik API.
- **Rekomendasi:** pemodel menambahkan operasi pada class yang pesannya
  merujuk (mis. `Attempt`, `Jawaban`, `Soal`, `PaketUjian`, `JadwalUjian`,
  `NilaiAkhir`) sesuai endpoint di `backend/internal/*`, lalu mengaitkan
  `signature` tiap pesan; atau dokumentasikan di naskah bahwa label pesan
  Sequence adalah naratif (endpoint + keterangan), bukan rujukan Operation.
  Kelengkapan pasangan synchCall/reply (NA-02 — BUKAN pelanggaran): semua
  synchCall tanpa reply dapat dijelaskan — pesan aktor↔UI (`isi …`,
  `tampilkan …`, `redirect …`), self-call internal (`verify Argon2id…`,
  `cek deadline lewat…`), dan `INSERT sessions` yang hasilnya tercakup
  dalam reply `200 + session cookie + CSRF token`.

### K3 — Lifeline tanpa stereotip peran (R2-27)

- **Bukti:** 10/10 lifeline tanpa stereotip; tidak ada stereotip apa pun di
  model (hanya label `«include»`/`«extend»` di view Use Case).
- **Alasan ditunda:** klasifikasi peran adalah penilaian semantik (mis.
  apakah `Monitor WebSocket` boundary atau control dapat diperdebatkan).
- **Rekomendasi** pemetaan yang diusulkan (bila user setuju): boundary =
  `Siswa / Pengguna`, `LoginPage / GantiPasswordPage`, `UjianPage`,
  `Monitor WebSocket`; control = `Auth API`, `Exam API`, `Grading`; entity
  = `PostgreSQL: users, sessions`, `PostgreSQL: jawaban, attempt,
  nilai_akhir`. Cantumkan definisi stereotip di legenda naskah (R2-28).

### K5 — Penamaan tidak identik lintas diagram (R2-17)

- **Bukti:** `hitung()` (Class NilaiAkhir) vs `hitungNilai` (Activity
  2a/2b, Sequence 3b) vs `HitungPoin per soal` (Sequence 3b);
  `sisaWaktu(now)` vs label Sequence `sisa waktu server`.
- **Alasan ditunda:** penggantian nama adalah perubahan model.
- **Rekomendasi:** tetapkan satu nama kanonis per operasi di naskah (mis.
  `hitungNilai`) dan pakai nama yang sama di Activity, Sequence, dan Class
  pada revisi berikutnya; pengecualian yang masih dapat diterima:
  `autoSubmitJikaLewat` sudah identik di ketiga diagram.

### K6 — 8 relasi ERD tanpa padanan asosiasi Class (NA-10)

Perbandingan pasangan (setelah normalisasi nama): 44 pasangan cocok,
8 relasi ERD tanpa padanan Class, 2 asosiasi Class tanpa relasi ERD
langsung yang keduanya sudah dijelaskan (lihat bawah). Kedelapan relasi
ERD tersebut (FK terverifikasi di migrasi):

| Relasi ERD | Sifat |
|---|---|
| `bank_soal.guru_id -> users` | FK ke `users.id` (migrasi 0003, SET NULL) — bukan ke tabel `guru`; Class tidak punya asosiasi BankSoal→User maupun BankSoal→Guru |
| `mapel.jurusan_id -> jurusan` | FK domain (nullable) — Class tidak punya asosiasi Mapel↔Jurusan |
| `ukk_berkas.siswa_id -> siswa` | FK domain — Class punya Siswa↔UkkNilai dan Siswa↔UkkSertifikat, tetapi tidak Siswa↔UkkBerkas (penghilangan tidak konsisten) |
| `berita_acara.dibuat_oleh -> users` | FK audit |
| `ukk_berkas.diupload_oleh -> users` | FK audit |
| `ukk_nilai.dinilai_oleh -> users` | FK audit |
| `usulan_unblock.pengawas_id -> users` | FK audit/proses (NOT NULL) |
| `usulan_unblock.diproses_oleh -> users` | FK audit (nullable) |

- **Alasan ditunda:** menambahkan asosiasi adalah penambahan semantik.
- **Rekomendasi:** tambahkan minimal 2 asosiasi domain (Mapel↔Jurusan,
  Siswa↔UkkBerkas) dan putuskan 6 sisanya: tambahkan sebagai asosiasi
  audit, atau dokumentasikan di naskah bahwa FK audit sengaja tidak
  dimodelkan sebagai asosiasi domain. Yang sudah dijelaskan (bukan temuan):
  `Role↔Permission` di Class (`*..*` via tabel join `role_permissions` di
  ERD) dan `Role↔User` di Class (asosiasi logis `users.role ->
  roles.code`, bukan FK — sesuai namanya).

### Observasi (bukan temuan) — K8, K9

- **K8 — Class lebih ketat dari DB pada Guru/Siswa:** Class menyatakan
  `Guru` 0..1 dan `Siswa` 0..1 per `User`, sedangkan migrasi tidak
  menetapkan UNIQUE pada `guru.user_id` / `siswa.user_id` (DB membolehkan
  0..* secara prinsip). Ini dibaca sebagai aturan bisnis (satu user =
  paling banyak satu profil guru/siswa) yang tidak ditegakkan DB — bukan
  kesalahan. **Rekomendasi untuk user:** putuskan — tambah UNIQUE via
  migrasi baru (perubahan skema, di luar cakupan audit ini), atau
  pertahankan sebagai aturan bisnis yang didokumentasikan.
- **K9 — Konvensi penamaan atribut vs kolom:** atribut Class memakai
  `snake_case`, kolom ERD memakai `camelCase`; setelah normalisasi,
  himpunan atribut = himpunan kolom pada semua 35 entitas (0 perbedaan).
  **Tidak diubah** (penggantian nama massal 229 atribut / 231 kolom hanya
  churn). **Rekomendasi:** cantumkan konvensi ini di legenda naskah
  (sudah ditulis di documentation diagram Class dan ERD).

### Butir terbuka untuk NASKAH (konvensi skripsi, bukan `.mdj`)

- **SK-02 (set diagram):** `.mdj` berisi Use Case, Activity (2), Sequence
  (2), Class, ERD. Diagram Deployment TIDAK ada di `.mdj`. Bila pedoman
  kampus target mewajibkan Deployment (set 5 diagram bila memakai class),
  naskah harus memuatnya dari sumber lain (mis. dokumen arsitektur) atau
  menjustifikasi ketiadaannya. Tidak dapat diputuskan dari `.mdj` saja.
- **SK-10 (ERD → LRS + spesifikasi file):** tidak ditemukan dokumen LRS
  atau spesifikasi file terpisah di workspace; ERD + migrasi 0001–0012
  dapat berfungsi sebagai spesifikasi, tetapi user perlu memastikan
  naskah memuat ketiganya (ERD, LRS, spesifikasi file) sesuai pedoman
  kampus target.
- **Naskah-level (tidak dapat diverifikasi dari `.mdj`):** SK-01
  konsistensi teori Bab II ↔ pembahasan; SK-04 tiap diagram terkait
  proses bisnis; SK-05 caption `Gambar <bab>.<urut>` di BAWAH gambar;
  SK-06 gambar + judul satu halaman tidak terpotong; SK-07 penempatan
  gambar di dekat teks pembahasnya; SK-08 diagram besar boleh landscape /
  A3 dilipat / diperkecil A4; SK-09 activity setelah narasi proses
  bisnis; SK-11 konsisten dengan SATU referensi notasi yang dikutip
  (dokumen ini memakai OMG UML 2.5.1, formal/2017-12-05); SK-12 format
  naskah mengikuti pedoman kampus target. Periksa terhadap naskah,
  bukan terhadap `.mdj`.

### Keputusan yang dibutuhkan dari user (rekap §5 + §4 AUDIT-INDUSTRI.md)

1. A1: setujui nama ExtensionPoint (`attempt diblokir` di UC07,
   `import file dipilih` di UC03) — lalu dapat diterapkan di StarUML /
   skrip lanjutan.
2. K2: putuskan — tambah operasi + signature (pekerjaan pemodelan),
   atau dokumentasikan label naratif di naskah.
3. K3: setujui pemetaan stereotip lifeline (§5 K3) atau tolak.
4. K5: tetapkan nama kanonis operasi (`hitungNilai` vs `hitung`).
5. K6: putuskan 8 relasi ERD — tambah asosiasi Class (minimal
   Mapel↔Jurusan, Siswa↔UkkBerkas) atau dokumentasikan FK audit
   sebagai sengaja tidak dimodelkan.
6. K8: putuskan — tambah UNIQUE `guru.user_id` / `siswa.user_id` via
   migrasi baru, atau pertahankan sebagai aturan bisnis.
7. SK-02/SK-10: pastikan naskah memuat Deployment (bila diwajibkan
   pedoman kampus target) serta LRS + spesifikasi file untuk ERD.
8. Klaster tepi bawah `users` (§6.3): pilih Opsi A (terima apa adanya)
   atau Opsi B (redesain tata letak ERD).

## 6. Batasan yang terbukti (perilaku tool + batas render)

### 6.1 Sequence — dua batasan StarUML (kalibrasi empiris, file TES)

1. **Self-message SELALU render terakhir.** Self-message selalu ter-render
   SETELAH semua pesan non-self, apa pun `points`/`top`/urutan ownedViews.
   TIDAK ADA field `.mdj` yang memindahkannya tanpa merusak semantik.
   Buktinya terlihat pada ekspor Sequence 3a: `11: verify
   Argon2id(password, password_hash)` (pesan ke-5 dari 16 secara semantik)
   ter-render di bawah semua pesan non-self. Ini batasan tool, bukan cacat
   `.mdj`. Untuk skripsi, urutan yang mengikat adalah urutan vertikal
   pesan di live canvas / `.mdj` (lihat §7), bukan posisi render
   self-message pada ekspor.
2. **Bentang panah non-self = CELAH antar kotak kepala lifeline
   TER-RENDER.** Points X pesan DIABAIKAN (Y = `points[0].y`, di-clamp ≤
   bawah-lifeline+20); kotak ter-render = (`left` tersimpan,
   `max(width tersimpan, lebar-fit-teks)`) — lebar 140→40 tetap render
   84–146px (fit teks). Konsekuensi: celah panah HANYA bisa dikecilkan
   lewat pitch lifeline (lihat §2.2), dan lebar tersimpan dipertahankan
   seragam 140 agar celah seragam.

Tambahan batasan Sequence yang dihormati: `head`/`tail` `$ref` bersifat
load-bearing — JANGAN PERNAH disentuh.

### 6.2 ERD — `end1` dan render kardinalitas

StarUML me-render kardinalitas ERD dari field model, BUKAN dari teks label
view (label view relasi ERD hanya menampilkan NAMA relasi). Mekanisme
(StarUML 7.1.0, `app.asar`): `ERDRelationshipEnd` constructor
`this.cardinality = '1'`; `ERDRelationshipView.update()` memakai
`switch (this.model.end1.cardinality)` — `'0..1'` → `ES_CROWFOOT_ZERO_ONE`
(o|), `'1'` → `ES_CROWFOOT_ONE` (||). Sebelum T1, 16 relasi bersisi induk
`0..1` tidak punya field `end1.cardinality`, sehingga ter-render `||`
padahal model sudah benar — diperbaiki T1 dengan menambahkan field
eksplisit (§2.4). 38 relasi lain sengaja tidak diberi field eksplisit
(default `'1'` sudah benar).

### 6.3 Klaster tepi bawah `users` (ERD) — tidak dapat direnggangkan tanpa merusak label

**Fakta:** 12 garis vertikal keluar dari tepi bawah entitas `users`
dengan jarak 10px (x=131..270). Label relasi `audit_log.actor_id -> users`
selebar 202px yang ditempatkan di celah antara bawah `users` (y=280) dan
atas `audit_log` (y=360) PASTI dilintasi ~6 garis lain. Bukti (pencarian
sistematis pada perbaikan R9b, bukan dugaan): kapsul 60px di sekitar
garis audit hanya beririsan dengan ruang bebas di dalam celah itu, dan
setiap rect 202px di sana memotong vertikal-vertikal koridor. Label
ditempatkan menempel garisnya (jarak pusat-ke-garis 5px); garis-garis
koridor lewat DI BELAKANG rect label. Clearance kotak ≥ 8px dan 0
tumpang tindih label terpenuhi. Aturan ketat "clearance ≥ 8px dari SEMUA
garis" TIDAK dapat dipenuhi bersamaan dengan "pusat label ≤ 60px dari
garisnya" untuk label ini (dan untuk `jawaban.attempt_id`, yang celah
vertikal terlebarnya 226px < 217+8+8 = 233px yang dibutuhkan; saku terbaik
memberi gap garis min 5,5px > 3px standar G3).

**Dua opsi tindak lanjut — BUTUH PERSETUJUAN USER** (belum ada yang
dipilih pada fase ini):

- **Opsi A — Terima apa adanya.** Garis koridor lewat di belakang rect
  label; keterbacaan label audit/jawaban terhadap garis yang melintas di
  belakangnya sudah diperiksa visual. Tidak ada pekerjaan lanjutan pada
  `.mdj`. Konsekuensi: bila penguji mempersoalkan garis yang melintas di
  belakang label, jawabannya adalah keterbatasan tata letak yang
  didokumentasikan di sini + screenshot pembuktian.
- **Opsi B — Redesain tata letak ERD.** Renggangkan fan-out koridor /
  pindahkan entitas agar koridor tidak lagi padat (mis. sebar titik keluar
  garis dari tepi bawah `users`, atau geser `audit_log` dan entitas di
  bawahnya). Konsekuensi: pekerjaan tata letak besar — berisiko menggeser
  ratusan view + label dan merusak kerapian yang sudah dicapai ronde
  sebelumnya; wajib verifikasi visual ulang penuh di StarUML asli. Ini
  pekerjaan redesain (bukan audit/perbaikan), dan TIDAK dimulai tanpa
  persetujuan eksplisit user.

### 6.4 Ambiguitas aturan render label (terbuka — JANGAN perbaiki massal)

Aturan "rata-rata dua titik tengah" (dipakai G1–G4, ronde 8, divisi
Sequence) vs "titik tengah bawah `points[(n-1)//2]`" (R10, 2 eksperimen
file TES) memprediksi posisi berbeda untuk edge ber-n GENAP. Bukti
screenshot StarUML asli (verifikasi-r10): label ERD
`backup_meta.dibuat_oleh -> users` ter-render di posisi tersimpan (aturan
titik-tengah-bawah memprediksi y=−511, tak terlihat) DAN label
`jawaban.attempt_id -> attempt` ter-render di posisi tersimpan (aturan
rata-rata memprediksi x=−81, tak terlihat). Kedua aturan murni terbantahkan
untuk sedikitnya satu label; hipotesis yang konsisten dengan kedua
screenshot adalah render ≈ posisi tersimpan (`left`/`top`). Bukti belum
memutuskan — perlu SATU eksperimen decisive di StarUML asli (ubah
`distance` SATU label pada SALINAN file, muat ulang, screenshot) sebelum
mengubah 100+ label. Sampai ada bukti itu, JANGAN me-re-solve label
massal: risikonya merusak file yang saat ini ter-render benar.

### 6.5 Koreksi terhadap artefak audit sebelumnya (bukan temuan baru)

Pemeriksaan awal audit dengan nama field yang salah sempat menghasilkan
tiga "temuan" palsu; ketiganya ditarik setelah field StarUML yang benar
diperiksa:

1. **Kepemilikan Include/Extend (UC-05/UC-09):** StarUML menyimpan SEMUA
   relasi (Association, Include, Extend, ControlFlow) dengan `_parent` =
   package (`Aktor & Use Case` / `Domain`), bukan UseCase pemilik. Ini
   konvensi serialisasi tool, bukan pelanggaran kepemilikan semantik —
   arah `source`/`target` sudah benar (including→addition,
   extension→extendedCase). TIDAK diubah.
2. **Siklus include (UC-04) dan Decision tanpa incoming/outgoing
   (AC-03):** artefak dari pemakaian field `includingCase` /
   `incoming`/`outgoing` yang tidak ada di serialisasi StarUML. Dengan
   field benar (`source`/`target` untuk relasi; edge dihitung dari
   `source`/`target` flow): tidak ada siklus include; semua Decision
   in=1/out=2. TIDAK ada pelanggaran.
3. **OccurrenceSpecification / ExecutionSpecification (SEQ-18–SEQ-22):**
   StarUML tidak membuat elemen `OccurrenceSpecification` untuk ujung
   pesan (ujung pesan merujuk langsung ke Lifeline) dan tidak ada
   `ExecutionSpecification` di model ini. Ini konvensi representasi tool
   yang diterima validator — dicatat sebagai batasan, bukan temuan.

### 6.6 Yang tidak dapat diverifikasi dari JSON `.mdj`

- Kepala panah Sequence (SEQ-14 — StarUML me-render dari `messageSort`);
  perilaku render StarUML atas field baru (documentation, kardinalitas) —
  file tidak pernah dibuka-disimpan lewat StarUML sesuai aturan; kualitas
  visual akhir (simetri R2-03, keterbacaan label) — perlu spot-check 1× di
  StarUML asli oleh user.
- **Tidak diubah pada fase ini:** kode aplikasi, exe, zip, halaman ujian,
  kartu, dan file `.mdj` lain (baseline, `.bak`) — hanya
  `docs/uml/cbt-smk-v2.mdj` yang diubah via skrip perbaikan.

## 7. Cara pakai file untuk skripsi

1. **Buka `.mdj` di StarUML.** File `cbt-smk-v2.mdj` (SHA `c068a179…`)
   dibuka di StarUML (7.x). Jangan simpan ulang dari StarUML kecuali
   memang ingin mengubah model — StarUML dapat menulis ulang format /
   field (mis. menimpa `visible` EdgeLabelView tiap render; Note
   dirender dari `text` milik VIEW bukan model). Bila StarUML bertanya
   "Save changes?" setelah hanya melihat-lihat, jawab **No / Don't Save**.
   Validasi bawaan StarUML berjalan otomatis saat file dibuka: ✅ centang
   hijau di status bar = 0 error; ⚠️ segitiga kuning = ada error (klik
   untuk membuka panel `VALIDATION RESULTS`).
2. **Ekspor ulang dari StarUML BERLISENSI agar tanpa watermark.** Ekspor
   PNG/SVG yang disertakan di `your_files/uml-png/` + `your_files/uml-svg/`
   dibuat dari StarUML trial, sehingga memuat watermark `UNREGISTERED`
   (pola berulang di seluruh gambar — lihat §8). Watermark tersebut
   TIDAK dihilangkan dengan cara apa pun (tidak ada edit gambar untuk
   menghapus watermark). Untuk naskah skripsi, buka `.mdj` di StarUML
   berlisensi dan ekspor ulang diagramnya (File > Export Diagram, atau
   CLI `staruml image <file> -f svg/png -s @Diagram`) — file `.mdj`
   sendiri BERSIH dari watermark.
3. **Penomoran pesan pakai live canvas / `.mdj`.** Penomoran pesan pada
   PNG ekspor dihitung ulang oleh exporter (berbeda dari live canvas) —
   dan penomoran `1`, `1.1` pada Sequence Diagram adalah anotasi tool /
   konvensi proyek, BUKAN persyaratan OMG (NA-01, SEQ-17). Untuk skripsi,
   pakai penomoran dari live canvas / `.mdj` (urutan vertikal atas →
   bawah yang mengikat), dan bila mengutip nomor pesan, sebutkan bahwa
   nomor tersebut adalah anotasi tool. Perhatikan juga: self-message
   ter-render terakhir pada ekspor StarUML (§6.1) — urutan semantiknya
   tetap urutan di live canvas / `.mdj`.
4. **Caption + penempatan gambar (SK-05–SK-08).** Caption diagram:
   `Gambar <bab>.<urut>`, judul DI BAWAH gambar (angka pertama = bab,
   angka kedua = urutan gambar dalam bab). Gambar + judulnya harus dalam
   satu halaman, tidak terpotong. Tempatkan gambar di antara bagian teks
   yang paling banyak membahasnya. Diagram yang terlalu besar boleh
   landscape / kertas A3 dilipat / diperkecil ke A4 — bukan pelanggaran
   "tidak muat satu halaman" (NA-07). Detail tipografi caption mengikuti
   pedoman kampus target (SK-12), bukan satu format universal (NA-22).
5. **ERD → LRS + spesifikasi file (SK-10).** Naskah yang memakai ERD
   harus memuat ERD + LRS + spesifikasi file. ERD di `.mdj` + migrasi
   `backend/migrations/0001–0012` dapat berfungsi sebagai spesifikasi,
   tetapi ketiganya harus dipastikan ada di naskah sesuai pedoman kampus
   target. Konvensi yang dipakai: atribut Class `snake_case`, kolom ERD
   `camelCase` (himpunan atribut = himpunan kolom, 0 perbedaan — K9);
   konvensi baca kardinalitas ERD: `<anak-per-induk> : <induk-per-anak>`
   (lihat §2.1 K7).
6. **Deployment (SK-02).** Diagram Deployment TIDAK ada di `.mdj`. Bila
   pedoman kampus target mewajibkan Deployment, muat dari sumber lain
   (mis. dokumen arsitektur) atau justifikasi ketiadaannya di naskah.
7. **Satu referensi notasi (SK-11 + SK-01).** Kutip SATU referensi notasi
   — OMG UML 2.5.1, formal/2017-12-05 — dan pastikan semua simbol
   diagram konsisten dengan referensi itu. Teori UML di Bab II harus
   konsisten dengan diagram yang dibahas di Bab III/IV (SK-01); tiap
   diagram harus terkait proses bisnis sistem (SK-04); activity diagram
   diletakkan setelah narasi proses bisnis (SK-09).
8. **Butir yang menunggu keputusan user (§5).** Bila A1/K2/K3/K5/K6/K8
   atau Opsi klaster `users` (§6.3) diputuskan kemudian, terapkan
   keputusannya sebagai revisi `.mdj` baru (dengan backup + laporan
   seperti fase ini) dan perbarui dokumen ini — jangan mengedit ekspor
   PNG/SVG yang sudah ada untuk meniru hasil revisi.

## 8. Ekspor PNG + SVG + status XMI

### 8.1 Metode ekspor (17 Okt 2026, Divisi Ekspor + Dokumentasi)

1. Salinan file final (`SHA c068a179…`) diekspor dengan **CLI resmi
   StarUML** (`staruml image <file> -f svg -s @Diagram -o
   "…/<%=filenamify(element.name)%>.svg"`) pada Xvfb display `:99` —
   ekspor native StarUML, bukan render ulang pihak ketiga. CLI `image`
   format PNG TIDAK dipakai (di sandbox ini output PNG-nya berisi
   background garbage framebuffer — temuan ronde sebelumnya).
2. Hanya atribut `width`/`height` kanvas SVG diubah ke ukuran kanvas yang
   sudah terverifikasi pada ronde sebelumnya (5 kanvas diperlebar dari
   ukuran native agar tidak ada konten yang terpotong; Class dan ERD
   ukurannya sama dengan native). Isi SVG (elemen, teks, watermark) TIDAK
   diubah.
3. PNG dirender dari SVG dengan **cairosvg 2.9.1** (venv terisolasi),
   skala 2× (lebar/tinggi output = 2× kanvas SVG), `background_color`
   putih — sama seperti ronde sebelumnya (PNG mode RGB).
4. Verifikasi: 7/7 SVG + PNG terbaca; tidak ada piksel gelap diagram yang
   menyentuh tepi kanvas (0 piksel di keempat sisi pada semua PNG — tidak
   ada konten terpotong); spot-check visual thumbnail ERD (35 entitas) dan
   Sequence 3a (4 lifeline, 16 pesan) ter-render benar, termasuk
   self-message yang ter-render terakhir (§6.1). File `.mdj` TIDAK diubah
   pada fase ekspor (hanya dibaca).

### 8.2 Daftar file ekspor

Semua file di `~/workspace/your_files/uml-png/` + `~/workspace/your_files/uml-svg/`.
Ukuran kanvas SVG = kolom "Kanvas"; PNG = 2× kanvas, mode RGB, latar putih.

| # | File | Kanvas SVG (px) | PNG (px) | Ukuran SVG | Ukuran PNG |
|---|---|---|---|---|---|
| 1 | `01-use-case` — Use Case | 1520 × 1290 | 3040 × 2580 | 649.202 byte | 755.155 byte |
| 2 | `02-class` — Class Domain | 1771 × 2240 | 3542 × 4480 | 1.456.133 byte | 1.545.918 byte |
| 3 | `03-erd` — ERD | 2047 × 2165 | 4094 × 4330 | 1.887.015 byte | 1.723.592 byte |
| 4 | `04-activity-2a` — Activity Pelaksanaan Ujian | 1550 × 1033 | 3100 × 2066 | 486.805 byte | 466.880 byte |
| 5 | `05-activity-2b` — Activity Koreksi Esai ke Rekap | 465 × 1253 | 930 × 2506 | 181.346 byte | 352.621 byte |
| 6 | `06-sequence-3a` — Sequence Login + Wajib Ganti Password | 975 × 1335 | 1950 × 2670 | 326.372 byte | 395.949 byte |
| 7 | `07-sequence-3b` — Sequence Simpan Jawaban + Auto-Submit | 1160 × 1390 | 2320 × 2780 | 465.793 byte | 438.920 byte |

File `.mdj` yang dikirim: `~/workspace/your_files/cbt-smk-v2.mdj`
(tersinkron dengan `docs/uml/cbt-smk-v2.mdj`) — **SHA-256
`c068a179be3df130e566dd26ea826c27a4368e09463d2cf5937eae0dee1025df`**.

### 8.3 DISCLOSURE — Watermark UNREGISTERED

Ekspor PNG/SVG di atas dibuat dari **StarUML 7.1.0 trial** (tanpa kunci
lisensi, tanpa crack, tanpa modifikasi file lisensi), sehingga memuat
**watermark `UNREGISTERED`** — teks abu-abu terang berulang di seluruh
gambar (jumlah kemunculan string `UNREGISTERED` per SVG: 01: 2.268,
02: 4.862, 03: 5.430, 04: 1.740, 05: 630, 06: 1.188, 07: 1.725).
Watermark tersebut **TIDAK dihilangkan dengan cara apa pun** — tidak ada
edit gambar untuk menghapus watermark, tidak ada modifikasi `app.asar`,
tidak ada kunci lisensi palsu. File `.mdj` sendiri BERSIH dari watermark.
Untuk naskah skripsi, ekspor ulang dari StarUML berlisensi (§7 butir 2).

### 8.4 Status XMI — TIDAK TERSEDIA (tidak ada file `.xmi` yang dikarang)

Ekspor XMI di StarUML disediakan oleh **ekstensi** `staruml.xmi`
(bukan fitur bawaan). Hasil pengecekan fase ini:

- Folder ekstensi pengguna `~/.config/StarUML/extensions/user/` **KOSONG**
  — ekstensi XMI tidak terpasang. CLI StarUML (`staruml --help`) hanya
  menyediakan perintah `ejs`, `image`, `html`, `pdf`, `exec` — **tidak
  ada jalur CLI untuk XMI**.
- Registry ekstensi resmi **dapat dijangkau** pada pengecekan ini
  (HTTP 200, `https://staruml.io/api/extensions/registry.json`) dan
  mencantumkan `staruml.xmi` v0.9.6 (2025-04-16, untuk StarUML ≥ 6.0.0).
  Koreksi atas laporan ronde sebelumnya: klaim "registry tidak terjangkau"
  TIDAK lagi berlaku pada pengecekan ini — yang tidak tersedia adalah
  ekstensi yang terpasang, bukan registry-nya. Instalasi ekstensi lewat
  Extension Manager GUI tidak dapat dilakukan di sandbox ini (tidak ada
  window manager / otomasi GUI yang tersedia), dan ekstensi tidak
  diunduh/diinstal manual.
- **Tidak ada file `.xmi` yang dibuat atau dikarang.** Bila user
  membutuhkan XMI: buka `.mdj` di StarUML reguler → Extension Manager →
  instal ekstensi `xmi` → File > Export > XMI. Alternatif tanpa ekstensi:
  file `.mdj` sendiri adalah JSON StarUML yang dapat dibaca programatis
  (struktur `$ref`/`_id`).

— Akhir dokumen —
