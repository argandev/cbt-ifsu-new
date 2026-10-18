# AUDIT OMG + INDUSTRI — Diagram UML CBT SMK v2

- **File:** `docs/uml/cbt-smk-v2.mdj`
- **Tim:** DIVISI AUDIT OMG + INDUSTRI (20 orang — dikerjakan satu penulis, tanpa spawn)
- **Acuan aturan:** `docs/uml/RISET-STANDAR-UML.md` (dibaca sampai habis sebelum audit;
  §1 OMG 75 butir, §2 kualitas R2-01–R2-28, §3 konvensi SK-01–SK-12,
  §4 BUKAN ATURAN NA-01–NA-25, §5 tiga koreksi premis, §6 hierarki bukti)
- **Script perbaikan:** `docs/uml/fix_audit_industri.py` (idempotent; laporan
  `docs/uml/fix_audit_industri_report.json`; backup
  `docs/uml/cbt-smk-v2.mdj.pre-audit-industri.bak`)
- **SHA-256 sebelum:** `d310d2c4…` (2.194.431 byte)
- **SHA-256 sesudah:** `cc749b66b24f82cc8ea0a9e22a038ae5adf23a28f41baf3ed90303a18a2c684a`
  (2.198.375 byte)
- **Validator** `docs/uml/validate-mdj-real.py`: **LOLOS, 0 pelanggaran**
  (dijalankan sesudah perbaikan; run ke-2 script perbaikan: 0 perubahan —
  idempotent terbukti)

## 1. Ringkasan angka

| Tingkat (sesuai §6 hierarki bukti) | Temuan (kelompok) | Diperbaiki | Ditunda karena semantik / di luar cakupan |
|---|---|---|---|
| WAJIB OMG (§1) | 1 (A1) | 0 | 1 (A1) |
| Kualitas jurnal (§2) + konsistensi ERD↔Class (NA-10) | 7 (K1–K7) | 2 (K4, K7 — 26 edit) | 5 (K1, K2, K3, K5, K6) |
| Konvensi skripsi (§3) terhadap `.mdj` | 0 | — | — |
| Butir terbuka untuk NASKAH (bukan `.mdj`) | 2 (SK-02, SK-10) | — | 2 (keputusan user) |

Total edit yang diterapkan: **26** (19 kardinalitas ERD + 7 documentation
diagram). Total temuan yang sengaja TIDAK diubah: **6 kelompok**
(A1, K1, K2, K3, K5, K6) + 2 observasi yang didokumentasikan (K8, K9) +
2 butir naskah (SK-02, SK-10).

Prinsip yang dipakai: bila memperbaiki berarti mengubah SEMANTIK BISNIS
(relasi, atribut, alur, pesan, klasifikasi, nama) — JANGAN UBAH; catat
sebagai temuan + rekomendasi. Yang diperbaiki hanya (a) tampilan
kardinalitas ERD agar sesuai ground truth database yang sudah terverifikasi,
dan (b) teks dokumentasi/narasi diagram (non-semantik).

## 2. Daftar temuan per diagram

Format: diagram | aturan (ID) | bukti (nama elemen/ID) | tingkat | status.

| Diagram | Aturan | Bukti | Tingkat | Status |
|---|---|---|---|---|
| Use Case | UC-09 `extensionLocation [1..*]` | 2× UMLExtend tanpa ExtensionPoint: `UC18 Usul & Proses Unblock → UC07 Monitor Ujian Real-time` (condition `attempt diblokir`, id `VW3VFH11…`) dan `UC04 Import Soal → UC03 Kelola Bank Soal` (condition `import file dipilih`, id `K5XYTJYZ…`); tidak ada satu pun elemen `UMLExtensionPoint` di model (hanya compartment view kosong) | WAJIB OMG | DITUNDA (A1, §4) |
| Use Case | R2-01 persilangan | 244 persilangan segmen (44 edge, 187 segmen; metode §8) | Kualitas | DITUNDA (K1, §4) |
| Class | R2-01 persilangan | 259 persilangan segmen (46 edge, 196 segmen) — sama dengan laporan fase line-neatness | Kualitas | DITUNDA (K1, §4) |
| Class | R2-21 / SK-L10 traceability | Hanya 11 operasi pada 6 dari 34 class; 28 class tanpa operasi; operasi menyimpan signature sebagai string nama (tanpa parameter terstruktur) sehingga cek argumen SEQ-02/SEQ-10 tidak dimungkinkan | Kualitas | DITUNDA (K2, §4) |
| Class | R2-17 penamaan lintas diagram | `hitung()` (Class NilaiAkhir) vs `hitungNilai` (Activity 2a/2b, Sequence 3b) vs `HitungPoin per soal` (Sequence 3b); `sisaWaktu(now)` vs label Sequence `sisa waktu server` | Kualitas | DITUNDA (K5, §4) |
| Class | Konsistensi ERD↔Class (NA-10) | 8 relasi ERD tanpa padanan asosiasi Class (lihat K6, §4) | Kualitas | DITUNDA (K6, §4) |
| ERD | Kardinalitas tidak presisi | 19 relasi: 16 FK nullable tertulis `1` (seharusnya `0..1`), 2 FK UNIQUE tertulis `0..*` (seharusnya `0..1`), 1 normalisasi `1..1` → `1` | Kualitas | DIPERBAIKI (K7, §3) |
| ERD | Konsistensi ERD↔Class (NA-10) | 8 relasi ERD tanpa padanan asosiasi Class (lihat K6, §4) | Kualitas | DITUNDA (K6, §4) |
| ERD | R2-01 persilangan | 406 persilangan segmen (54 edge, 233 segmen) — sama dengan laporan fase line-neatness | Kualitas | DITUNDA (K1, §4) |
| Activity 2a | R2-01 persilangan | 7 persilangan segmen (25 edge, 73 segmen) — sama dengan laporan fase line-neatness | Kualitas | DITUNDA (K1, §4) |
| Activity 2a, 2b, Sequence 3a, 3b, Use Case, Class, ERD | R2-18 narasi diagram | 7/7 diagram tanpa `documentation` di `.mdj` | Kualitas | DIPERBAIKI (K4, §3) |
| Activity 2b | R2-01 persilangan | 0 persilangan (11 edge, 16 segmen) | Kualitas | Bersih |
| Sequence 3a | R2-27 stereotip lifeline | 4/4 lifeline tanpa stereotip peran boundary/control/entity | Kualitas | DITUNDA (K3, §4) |
| Sequence 3a | SEQ-09 kelengkapan (bukan pelanggaran formal) | 16/16 pesan tanpa signature (label naratif) | Kualitas | DITUNDA (K2, §4) |
| Sequence 3b | R2-27 stereotip lifeline | 6/6 lifeline tanpa stereotip peran boundary/control/entity | Kualitas | DITUNDA (K3, §4) |
| Sequence 3b | SEQ-09 kelengkapan (bukan pelanggaran formal) | 16/16 pesan tanpa signature (label naratif) | Kualitas | DITUNDA (K2, §4) |

## 3. Yang diperbaiki

### K7 — Kardinalitas ERD (19 relasi)

Sebelum: 51× `0..* : 1` + 3× `0..* : 1..1` (54 relasi).
Sesudah: 36× `0..* : 1` + 16× `0..* : 0..1` + 2× `0..1 : 1` (54 relasi).
Konvensi baca (ditetapkan dari orientasi field StarUML:
`end1` = induk, `end2` = anak pemegang FK): gabungan =
`<anak-per-induk> : <induk-per-anak>`; `anak-per-induk` = `0..1` bila
kolom FK UNIQUE selain itu `0..*`; `induk-per-anak` = `1` bila kolom FK
NOT NULL selain itu `0..1`. Yang diubah: `ERDRelationship.cardinality`
+ `end2.cardinality`. Label view relasi ERD hanya menampilkan NAMA
relasi (StarUML me-render kardinalitas dari field model), jadi tidak ada
teks view yang perlu diubah.

Ground truth: flag `nullable`/`unique` 54 kolom FK di ERD dicocokkan
dengan migrasi `backend/migrations/0001–0012` — **0 perbedaan**
(50 definisi `CREATE TABLE` + 4 definisi `ALTER TABLE` di 0006/0009,
semua nullable non-UNIQUE). Contoh perubahan:

| Relasi | Sebelum → Sesudah | Dasar |
|---|---|---|
| `berita_acara.jadwal_id -> jadwal_ujian` | `0..* : 1..1` → `0..1 : 1` | `jadwal_id UNIQUE` (migrasi 0005); cocok dengan Class `(JadwalUjian:1, BeritaAcara:0..1)` |
| `nilai_akhir.attempt_id -> attempt` | `0..* : 1..1` → `0..1 : 1` | `attempt_id UNIQUE` (migrasi 0005); cocok dengan Class `(Attempt:1, NilaiAkhir:0..1)` |
| `nilai_akhir.jadwal_id -> jadwal_ujian` | `0..* : 1..1` → `0..* : 1` | Normalisasi penulisan (`1..1` ≡ `1`); `jadwal_id` NOT NULL, tidak UNIQUE |
| 16 relasi FK nullable (mis. `audit_log.actor_id`, `siswa.user_id`, `guru.user_id`, `bank_soal.guru_id`, `soal.kompetensi_id`, `soal.kisi_kisi_id`, `jadwal_ujian.event_id`, `jadwal_ujian.ruangan_id`, `mapel.jurusan_id`, `kisi_kisi.kompetensi_id`, + 6 kolom audit `dibuat_oleh`/`dinilai_oleh`/`diupload_oleh`/`diproses_oleh`/`actor_id`) | `0..* : 1` → `0..* : 0..1` | Kolom FK nullable (SET NULL); cocok dengan multiplisitas Class hasil koreksi F2 audit OMG sebelumnya |

### K4 — Narasi + legenda diagram (R2-18)

`documentation` ditambahkan pada 7/7 diagram (sebelumnya kosong):
Use Case 608 karakter, Class 499, Activity 2a 511, Activity 2b 323,
Sequence 3a 547, Sequence 3b 602, ERD 578. Isi: narasi alur + legenda
simbol + catatan yang relevan (mis. konvensi baca kardinalitas ERD,
status ExtensionPoint, status signature pesan). Ini teks dokumentasi
non-semantik; tidak mengubah elemen model apa pun.

## 4. Yang SENGAJA TIDAK diubah (temuan + rekomendasi untuk user)

### A1 — Extend tanpa `extensionLocation` (UC-09, WAJIB OMG)

Bukti: 2 relasi `UMLExtend` tidak punya `extensionLocation`; tidak ada
elemen `UMLExtensionPoint` di model. StarUML tidak menyimpan field
`extensionLocation` pada serialisasi `UMLExtend` di file ini, dan nama
extension point adalah keputusan pemodelan (penambahan elemen = semantik).
**Rekomendasi:** di StarUML, tambahkan ExtensionPoint pada extended
UseCase, lalu kaitkan: UC07 Monitor Ujian Real-time ← ExtensionPoint
`attempt diblokir` (untuk Extend UC18, condition `attempt diblokir`);
UC03 Kelola Bank Soal ← ExtensionPoint `import file dipilih` (untuk
Extend UC04, condition `import file dipilih`). Nama yang diusulkan
disalin dari condition yang sudah ada — bukan karangan baru.

### K1 — Persilangan garis (R2-01)

Angka: Use Case 244, Class 259, ERD 406, Activity 2a 7, Activity 2b 0,
Sequence 3a 0, Sequence 3b 0 (metode §8). Semua persilangan yang
tersisa adalah tegak lurus (rute ortogonal; 0 segmen diagonal di semua
diagram). **Tidak diperbaiki** karena: (a) fase line-neatness sebelumnya
secara sadar menukar persilangan dengan sudut 90°; (b) menurunkan ke 0
butuh redesain tata letak (pekerjaan besar, berisiko menggeser ratusan
view); (c) R2-24 melarang mengklaim "layout sudah rapih" hanya dari satu
aspek — laporan ini TIDAK mengklaim layout rapih. **Rekomendasi:** terima
sebagai utang kualitas yang diketahui, atau tugaskan redesain tata letak
khusus (bukan audit) bila penguji mempersoalkan. Catatan transfer P1
berlaku: eksperimen Purchase (1997) memakai graf abstrak, bukan diagram
UML — penerapannya ke UML adalah ekstrapolasi yang masuk akal, bukan
bukti langsung pada UML.

### K2 — Traceability pesan → operasi lemah (R2-21, SK-L10)

Bukti: 32/32 pesan Sequence tanpa `signature`; 11 operasi pada 6/34
class; 28 class tanpa operasi; operasi menyimpan signature sebagai
string nama (mis. `verifyPassword(p): Boolean`) tanpa parameter
terstruktur, sehingga korespondensi argumen SEQ-02/SEQ-10 tidak dapat
diperiksa. Ini BUKAN pelanggaran formal OMG (field `signature` Message
bersifat opsional; constraint SEQ-09 hanya berlaku bila signature ada).
**Tidak diperbaiki** karena menambahkan operasi/signature berarti
mengarang semantik API. **Rekomendasi:** pemodel menambahkan operasi
pada class yang pesannya merujuk (mis. `Attempt`, `Jawaban`, `Soal`,
`PaketUjian`, `JadwalUjian`, `NilaiAkhir`) sesuai endpoint di
`backend/internal/*`, lalu mengaitkan `signature` tiap pesan; atau
dokumentasikan di naskah bahwa label pesan Sequence adalah naratif
(endpoint + keterangan), bukan rujukan Operation. Kelengkapan pasangan
synchCall/reply (NA-02 — BUKAN pelanggaran): semua synchCall tanpa reply
dapat dijelaskan — pesan aktor↔UI (`isi …`, `tampilkan …`, `redirect …`),
self-call internal (`verify Argon2id…`, `cek deadline lewat…`), dan
`INSERT sessions` yang hasilnya tercakup dalam reply
`200 + session cookie + CSRF token`.

### K3 — Lifeline tanpa stereotip peran (R2-27)

Bukti: 10/10 lifeline tanpa stereotip; tidak ada stereotip apa pun di
model (hanya label `«include»`/`«extend»` di view Use Case).
**Tidak diperbaiki** karena klasifikasi peran adalah penilaian semantik
(mis. apakah `Monitor WebSocket` boundary atau control dapat
diperdebatkan). **Rekomendasi** pemetaan yang diusulkan (bila user
setuju): boundary = `Siswa / Pengguna`, `LoginPage / GantiPasswordPage`,
`UjianPage`, `Monitor WebSocket`; control = `Auth API`, `Exam API`,
`Grading`; entity = `PostgreSQL: users, sessions`,
`PostgreSQL: jawaban, attempt, nilai_akhir`. Cantumkan definisi
stereotip di legenda naskah (R2-28).

### K5 — Penamaan tidak identik lintas diagram (R2-17)

Bukti: `hitung()` vs `hitungNilai` vs `HitungPoin per soal`;
`sisaWaktu(now)` vs `sisa waktu server`. **Tidak diperbaiki** karena
penggantian nama adalah perubahan model. **Rekomendasi:** tetapkan satu
nama kanonis per operasi di naskah (mis. `hitungNilai`) dan pakai nama
yang sama di Activity, Sequence, dan Class pada revisi berikutnya;
pengecualian yang masih dapat diterima: `autoSubmitJikaLewat` sudah
identik di ketiga diagram.

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

**Tidak diperbaiki** karena menambahkan asosiasi adalah penambahan
semantik. **Rekomendasi:** tambahkan minimal 2 asosiasi domain
(Mapel↔Jurusan, Siswa↔UkkBerkas) dan putuskan 6 sisanya: tambahkan
sebagai asosiasi audit, atau dokumentasikan di naskah bahwa FK audit
sengaja tidak dimodelkan sebagai asosiasi domain. Yang sudah dijelaskan
(bukan temuan): `Role↔Permission` di Class (`*..*` via tabel join
`role_permissions` di ERD) dan `Role↔User` di Class (asosiasi logis
`users.role -> roles.code`, bukan FK — sesuai namanya).

### K8 (observasi, bukan temuan) — Class lebih ketat dari DB pada Guru/Siswa

Class menyatakan `Guru` 0..1 dan `Siswa` 0..1 per `User`, sedangkan
migrasi tidak menetapkan UNIQUE pada `guru.user_id` / `siswa.user_id`
(DB membolehkan 0..* secara prinsip). Ini dibaca sebagai aturan bisnis
(satu user = paling banyak satu profil guru/siswa) yang tidak ditegakkan
DB — bukan kesalahan. **Rekomendasi untuk user:** putuskan — tambah
UNIQUE via migrasi baru (perubahan skema, di luar cakupan audit ini),
atau pertahankan sebagai aturan bisnis yang didokumentasikan.

### K9 (observasi, bukan temuan) — Konvensi penamaan atribut vs kolom

Atribut Class memakai `snake_case`, kolom ERD memakai `camelCase`;
setelah normalisasi, himpunan atribut = himpunan kolom pada semua
35 entitas (0 perbedaan). **Tidak diubah** (penggantian nama massal
229 atribut / 231 kolom hanya churn). **Rekomendasi:** cantumkan
konvensi ini di legenda naskah (sudah ditulis di documentation
diagram Class dan ERD).

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

## 5. Yang diperiksa dan BERSIH

### Use Case (UC-01–UC-12)
19 UseCase bernama (UC-01 ✓); 6 Actor bernama, asosiasinya hanya ke
UseCase dan biner — 39/39 Actor↔UseCase (UC-02 ✓); tidak ada asosiasi
antar dua UseCase (UC-03 ✓); tidak ada siklus include (UC-04 ✓ —
3 include: UC08→UC09, UC05→UC03, UC11→UC10); arah include
including→addition dan extend extension→extendedCase benar, dengan
`condition` yang sudah diperbaiki audit OMG sebelumnya (UC-05, UC-07,
UC-08 ✓ — notasi putus-putus + label `«include»`/`«extend»` ada di
view); UC-06/UC-08 notasi ✓; UC-09 kecuali `extensionLocation` (A1);
UC-10 multiplisitas tidak ditampilkan pada asosiasi aktor (diperbolehkan
— bukan kesalahan, dan tidak ada klaim pembacaan yang salah); UC-11
notasi aktor stick-man ✓; UC-12 elips + nama ✓, tidak ada subject
boundary yang dimodelkan (tidak wajib bila subject tidak dimodelkan).

### Class (CL-01–CL-12)
46 asosiasi, semuanya ≥ 2 memberEnd (CL-01 ✓), tidak ada agregasi yang
dimodelkan (CL-07 vakum — tidak ada belah ketupat yang salah), semua
navigable dua arah (CL-05/CL-06: panah terbuka kedua ujung, konsisten),
tidak ada pasangan class ganda (duplikat asosiasi: 0), semua ujung
punya multiplisitas dengan format tanpa kurung siku (CL-03 ✓; CL-04
tidak berlaku karena tidak ada ujung tanpa angka), 46/46 asosiasi
bernama (CL-12 ✓), 229 atribut semua bertipe dan bervisibilitas
`public` (CL-08/CL-09 ✓), 11 operasi berformat benar (CL-10 ✓ —
signature sebagai string nama adalah gaya pemodelan file ini, lihat K2).

### Sequence (SEQ-01–SEQ-26) + CombinedFragment (CF-01–CF-13)
32 pesan: sort hanya `synchCall` (19), `reply` (12), `asynchCall` (1:
`broadcastAttempt`, sudah dikoreksi audit OMG sebelumnya) (SEQ-01 ✓);
tidak ada `createMessage`/`deleteMessage` (SEQ-07/SEQ-08/SEQ-23 vakum);
tidak ada `signature` pada pesan apa pun (SEQ-09/SEQ-10 tidak
terlanggar — lihat K2); tidak ada CombinedFragment (CF-01–CF-13 vakum —
percabangan dimodelkan sebagai alur linear naratif, bukan kesalahan);
tidak ada ExecutionSpecification/OccurrenceSpecification sebagai elemen
model (lihat §7); semua pesan horizontal atau menurun — 30 horizontal,
2 self-message berbentuk persegi kecil menurun (SEQ-11 ✓); lifeline
10/10 beridentifikasi tidak kosong (SEQ-24 ✓); tidak ada selector
lifeline (SEQ-26 vakum). SEQ-14 (kepala panah per sort) tidak dapat
diverifikasi dari JSON — StarUML me-render-nya dari `messageSort`
secara otomatis; tidak diklaim sebagai temuan.

### Activity (AC-01–AC-12)
2a: Initial `Mulai` in=0/out=1, semua outgoing ControlFlow (AC-05 ✓);
3 Decision masing-masing in=1/out=2 dengan guard di KEDUA cabang
(`tidak valid`/`valid`, `ke-1`/`ke-2`, `siswa submit`/`deadline lewat`)
(AC-03 ✓ — termasuk anjuran determinisme); 2b: Decision in=1/out=2
dengan guard `belum semua`/`semua terkoreksi` (AC-03 ✓); tidak ada
guard `else` ganda; tidak ada Merge/Fork/Join (AC-04/AC-07/AC-08 vakum);
tidak ada simbol gabungan (AC-09 vakum); FinalNode out=0 (AC-06 ✓);
tidak ada guard berkurung ganda (AC-02 ✓ — guard tersimpan tanpa
kurung, StarUML menambah `[ ]` saat render); tidak ada Decision dengan
2 incoming (AC-11 `decisionInputFlow` vakum).

### ERD (diperiksa sebagai konsistensi internal + konvensi proyek, NA-10)
35 entitas ↔ 34 class + 1 tabel join `role_permissions` (pemetaan
1:1, §2 K9: 0 perbedaan atribut↔kolom); 44 pasangan relasi cocok
dengan asosiasi Class; 54 relasi FK semuanya bernama
(`tabel_anak.kolom -> tabel_induk`); kardinalitas kini presisi (K7);
`identifying: false` pada semua relasi (konsisten — tidak ada FK yang
menjadi bagian PK anak, kecuali PK komposit tabel join yang dimodelkan
sebagai entitas sendiri).

### Kualitas jurnal (§2) — yang bersih
R2-02 belokan: semua segmen axis-aligned (0 diagonal) di semua diagram;
belokan adalah konsekuensi rute ortogonal (Use Case 143, Class 150,
ERD 179, Activity 2a 48, Activity 2b 5, Sequence 2+2) — dicatat, bukan
temuan. R2-03 simetri: penilaian visual, tidak diklaim. R2-05 satu gaya
notasi konsisten (tidak ada varian setara yang dicampur). R2-07 tidak
ada dekorasi bertele-tele. R2-09 perilaku dinamis memakai sequence
(2 sequence untuk 2 alur utama). R2-10/R2-16 satu sequence = satu alur;
tidak ada diagram raksasa semua-alur. R2-11/R2-12 stereotip yang ada
(`«include»`, `«extend»`) bermakna dan konsisten (lihat K3 untuk yang
belum ada). R2-13 satu simbol = satu makna (putus-putus hanya
include/extend/reply; solid hanya asosiasi/alur/panggilan). R2-14 dapat
dibedakan hitam-putih (pembeda berupa teks + gaya garis, bukan warna —
file tidak menyimpan override warna). R2-15 arah panah sesuai arah
pesan/alur. R2-19 ragam jenis simbol minimal per diagram (Use Case 5,
Class 2, Activity ≤ 4, Sequence 3, ERD 2 jenis). R2-20 label lengkap
Bahasa Indonesia + legenda kini di documentation diagram (R2-18).
R2-21 detail kunci: multiplisitas lengkap di Class, label pesan lengkap
di Sequence — kecuali operasi (K2). R2-22/R2-23/R2-04/R2-26: tidak ada
klaim kualitas yang dilarang di dokumen ini.

## 6. Hal §4 BUKAN ATURAN yang dicek dan TIDAK dijadikan temuan

NA-01 penomoran pesan Sequence (tidak dinilai — yang mengikat urutan
vertikal ✓); NA-02 reply wajib per synchCall (tidak ditegakkan — lihat
K2); NA-03 guard wajib per cabang Decision (tidak ditegakkan — semua
cabang kebetulan ber-guard); NA-04 urutan lifeline BCE wajib (tidak
ditegakkan — lihat K3); NA-05 batas angka lifeline (tidak ada angka
yang ditegakkan); NA-06 "label harus nama operasi" sebagai aturan
jurnal (tidak ditegakkan — yang berlaku SEQ-09 OMG, lihat K2);
NA-07 "diagram harus muat satu halaman" (tidak ditegakkan; yang berlaku
SK-06/SK-08 untuk naskah); NA-08 larangan mutlak diagonal (tidak
ditegakkan — faktanya 0 segmen diagonal); NA-09 guard `[[…]]` (tidak
ada yang tersisa; bukan notasi benar); NA-10 aturan notasi ERD dari
OMG (tidak ditegakkan — ERD hanya diaudit konsistensi internal);
NA-11/NA-12 angka "ukuran optimal"/"maksimal N elemen" (tidak ada angka
yang dikarang di dokumen ini); NA-13 "maksimal 6 simbol" dari Moody
(tidak dikutip sebagai angka mutlak); NA-14 klaim kualitas dari grid
ortogonal (tidak diklaim); NA-15 klaim kecepatan baca dari detail
(tidak diklaim); NA-16/NA-17 kepala lifeline create / tanda X
destruction (tidak ada pesan create/delete — vakum); NA-18 PascalCase
sebagai aturan skripsi (tidak ditegakkan — faktanya nama class
PascalCase); NA-19 satu set universal 7 diagram (tidak ditegakkan);
NA-20 "pedoman kampus selalu mengatur notasi" (salah — lihat §4);
NA-21 caption Romawi, NA-22 tipografi caption universal (naskah-level,
ikuti kampus target); NA-23 kesan "terasa jelas" (tidak dipakai sebagai
bukti); NA-24 guard sequence sebagai aturan pedoman kampus (tidak
diklaim — guard adalah aturan OMG CF-11/AC-02); NA-25 intuisi sebagai
bukti (tidak dipakai). Tiga koreksi premis §5 diterapkan: penomoran
pesan, reply wajib, guard wajib — tidak satu pun dijadikan temuan.

## 7. Koreksi terhadap artefak audit sebelumnya (bukan temuan baru)

Pemeriksaan awal dengan nama field yang salah sempat menghasilkan
tiga "temuan" palsu; ketiganya ditarik setelah field StarUML yang benar
diperiksa:

1. **Kepemilikan Include/Extend (UC-05/UC-09):** StarUML menyimpan
   SEMUA relasi (Association, Include, Extend, ControlFlow) dengan
   `_parent` = package (`Aktor & Use Case` / `Domain`), bukan UseCase
   pemilik. Ini konvensi serialisasi tool, bukan pelanggaran
   kepemilikan semantik — arah `source`/`target` sudah benar
   (including→addition, extension→extendedCase). TIDAK diubah.
2. **Siklus include (UC-04) dan Decision tanpa incoming/outgoing
   (AC-03):** artefak dari pemakaian field `includingCase` /
   `incoming`/`outgoing` yang tidak ada di serialisasi StarUML.
   Dengan field benar (`source`/`target` untuk relasi; edge dihitung
   dari `source`/`target` flow): tidak ada siklus include; semua
   Decision in=1/out=2. TIDAK ada pelanggaran.
3. **OccurrenceSpecification / ExecutionSpecification (SEQ-18–SEQ-22):**
   StarUML tidak membuat elemen `OccurrenceSpecification` untuk ujung
   pesan (ujung pesan merujuk langsung ke Lifeline) dan tidak ada
   `ExecutionSpecification` di model ini. Ini konvensi representasi
   tool yang diterima validator — dicatat sebagai batasan, bukan
   temuan.

## 8. Metode & batasan

- Audit dilakukan dengan skrip pemeriksa programatis atas JSON `.mdj`
  (indeks `$ref`, pasangan ERD↔Class, pasangan ERD↔migrasi, geometri
  view) + inspeksi manual field mentah untuk setiap temuan sebelum
  dicatat. Skrip audit awal yang memakai nama field salah dibuang;
  semua angka di dokumen ini berasal dari pemeriksaan ulang dengan
  field yang benar.
- **Persilangan** dihitung sebagai perpotongan interior antarsegmen
  dari edge yang berbeda (tidak termasuk titik ujung bersama dan
  segmen kolinear). Angka Class (259), ERD (406), Activity 2a (7)
  cocok dengan laporan fase line-neatness sebelumnya — metode
  terkalibrasi silang. Persilangan Use Case (244) baru diukur pada
  audit ini.
- **Tidak dapat diverifikasi dari JSON:** kepala panah Sequence
  (SEQ-14, StarUML me-render dari `messageSort`); perilaku render
  StarUML atas field baru (documentation, kardinalitas) — file tidak
  pernah dibuka-disimpan lewat StarUML sesuai aturan; kualitas visual
  akhir (simetri R2-03, keterbacaan label) — perlu spot-check 1× di
  StarUML asli oleh user (seperti yang sudah dilakukan pada ronde
  sebelumnya untuk aspek lain).
- **Tidak diubah:** kode aplikasi, exe, zip, halaman ujian, kartu,
  dan file `.mdj` lain (baseline, `.bak`) — hanya
  `docs/uml/cbt-smk-v2.mdj` yang diubah via `fix_audit_industri.py`.

## 9. Keputusan yang dibutuhkan dari user

1. A1: setujui nama ExtensionPoint (`attempt diblokir` di UC07,
   `import file dipilih` di UC03) — lalu divisi standardisasi dapat
   menerapkannya di StarUML / skrip lanjutan.
2. K2: putuskan — tambah operasi + signature (pekerjaan pemodelan),
   atau dokumentasikan label naratif di naskah.
3. K3: setujui pemetaan stereotip lifeline (§4 K3) atau tolak.
4. K5: tetapkan nama kanonis operasi (`hitungNilai` vs `hitung`).
5. K6: putuskan 8 relasi ERD — tambah asosiasi Class (minimal
   Mapel↔Jurusan, Siswa↔UkkBerkas) atau dokumentasikan FK audit
   sebagai sengaja tidak dimodelkan.
6. K8: putuskan — tambah UNIQUE `guru.user_id` / `siswa.user_id` via
   migrasi baru, atau pertahankan sebagai aturan bisnis.
7. SK-02/SK-10: pastikan naskah memuat Deployment (bila diwajibkan
   pedoman kampus target) serta LRS + spesifikasi file untuk ERD.
