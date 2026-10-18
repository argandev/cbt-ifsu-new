# VALIDASI BPMN — CBT SMK v2

**Tanggal validasi:** 17 Oktober 2026
**Validator:** Tim Validasi BPMN (subagent, bekerja sendiri — tanpa subagent turunan)
**Ruang lingkup:** 14 file `.bpmn` di `docs/bpmn/as-is/` (7 file) dan `docs/bpmn/to-be/` (7 file).
**Prinsip:** Tidak ada kode aplikasi yang diubah. Tidak ada isi diagram yang diubah. Tidak ada file `.bpmn`
yang gagal parse, sehingga tidak ada perbaikan apa pun yang dilakukan.

## Ringkasan hasil

| Hasil | Jumlah |
|---|---|
| **LOLOS** | **14** |
| GAGAL | **0** |

**Validasi XSD BPMN 2.0.2 resmi OMG: TIDAK DILAKUKAN** — seluruh URL XSD resmi yang dicoba
mengembalikan HTTP 404 (dan satu mirror GitHub timeout). Lihat § "Upaya unduh XSD" di bawah.
Pernyataan "LOLOS" di bawah berarti lolos cek 1, 2, 3, dan 5 (cek pengganti XSD), BUKAN lolos
validasi skema XSD resmi.

## Metode validasi

Skrip Python (`xml.etree.ElementTree`, Python stdlib) memeriksa setiap file terhadap 5 kelompok cek:

1. **XML well-formed** — `ET.parse()` harus sukses tanpa `ParseError`.
2. **Namespace BPMN 2.0.2** — root `definitions` memakai namespace
   `http://www.omg.org/spec/BPMN/20100524/MODEL`; `targetNamespace`, `exporter`, dan
   `exporterVersion` wajib terisi.
3. **Cek struktural:**
   - setiap `sequenceFlow` punya `sourceRef`/`targetRef` yang merujuk ID yang ada;
   - setiap flow node punya incoming + outgoing (kecuali start tanpa incoming, end tanpa outgoing);
   - tepat 1 start event, ≥ 1 end event;
   - setiap flow node punya elemen DI (`BPMNShape`), setiap `sequenceFlow` punya `BPMNEdge`,
     lane/participant juga punya `BPMNShape`, dan tidak ada DI yang merujuk ID tak dikenal;
   - tidak ada tumpang tindih bounding box antar `BPMNShape` flow node;
   - setiap `BPMNEdge` punya ≥ 2 `di:waypoint`.
4. **XSD BPMN 2.0.2 resmi OMG** — dicoba diunduh; GAGAL (lihat § "Upaya unduh XSD").
   Sebagai pengganti, cek 1–3 di atas + cek 5 dilakukan secara menyeluruh.
5. **Konsistensi label (cek ringan, format saja — bukan penilaian isi):**
   - gateway berlabel pertanyaan (diakhiri `?`);
   - task berlabel kata kerja + objek, Bahasa Indonesia;
   - elemen `documentation` ada per file (1 per file, di level `process`, berisi narasi alur + asumsi);
   - tambahan yang diperiksa: label start/end event, nama lane, dan outgoing flow gateway
     (harus punya nama/kondisi kecuali flow `default`).

Kriteria LOLOS/GAGAL: file dinyatakan **LOLOS** bila seluruh cek keras (well-formed, namespace,
referensi flow, konektivitas, jumlah start/end, kelengkapan DI, tanpa tumpang tindih) terpenuhi.
Cek label yang gagal akan dicatat sebagai temuan, bukan otomatis GAGAL — dalam validasi ini
tidak ada cek label yang gagal.

## Hasil per file

### As-Is (`docs/bpmn/as-is/`)

| File | Hasil | Task | Gateway | Lane | Start | End | SeqFlow | BPMNShape | BPMNEdge | Doc |
|---|---|---|---|---|---|---|---|---|---|---|
| 01-pembuatan-bank-soal.bpmn | ✅ LOLOS | 11 | 2 | 3 | 1 | 1 | 16 | 19 | 16 | 1 |
| 02-penyusunan-paket-jadwal.bpmn | ✅ LOLOS | 12 | 2 | 3 | 1 | 1 | 17 | 20 | 17 | 1 |
| 03-pelaksanaan-ujian-siswa.bpmn | ✅ LOLOS | 13 | 3 | 3 | 1 | 2 | 20 | 23 | 20 | 1 |
| 04-penilaian-koreksi-esai.bpmn | ✅ LOLOS | 10 | 2 | 2 | 1 | 1 | 15 | 17 | 15 | 1 |
| 05-rekap-berita-acara.bpmn | ✅ LOLOS | 9 | 2 | 3 | 1 | 1 | 14 | 17 | 14 | 1 |
| 06-ukk.bpmn | ✅ LOLOS | 13 | 1 | 4 | 1 | 1 | 16 | 21 | 16 | 1 |
| 07-import-soal.bpmn | ✅ LOLOS | 10 | 1 | 2 | 1 | 1 | 13 | 16 | 13 | 1 |

Bukti per file (as-is): XML well-formed ✅; namespace BPMN 2.0.2 ✅;
`targetNamespace = http://cbt-smk-v2/bpmn/as-is` ✅;
`exporter = "CBT SMK v2 - Tim BPMN As-Is"`, `exporterVersion = "1.0 (2026-10-17)"` ✅;
semua `sequenceFlow` merujuk ID yang ada ✅; konektivitas incoming/outgoing lengkap ✅;
tepat 1 start, ≥ 1 end ✅; semua flow node + lane + participant punya `BPMNShape`, semua
`sequenceFlow` punya `BPMNEdge` (masing-masing ≥ 2 waypoint), tanpa DI menggantung ✅;
tanpa tumpang tindih bounding box ✅; semua gateway berlabel pertanyaan (diakhiri `?`) ✅;
semua task berlabel kata kerja + objek Bahasa Indonesia ✅; `documentation` di level
`process` ada dan terisi ✅; semua outgoing flow gateway punya nama/kondisi ✅.
Semua `process` as-is memakai `isExecutable="false"` (model deskriptif, bukan executable).
Semua file as-is memakai prefix `bpmn:`, elemen `collaboration` + `participant` (1 pool),
dan task generik `<bpmn:task>`.

### To-Be (`docs/bpmn/to-be/`)

| File | Hasil | Task | Gateway | Lane | Start | End | SeqFlow | BPMNShape | BPMNEdge | Doc |
|---|---|---|---|---|---|---|---|---|---|---|
| 01-pembuatan-bank-soal.bpmn | ✅ LOLOS | 11 (8 userTask + 3 serviceTask) | 4 | 2 | 1 | 1 | 20 | 19 | 20 | 1 |
| 02-penyusunan-paket-jadwal.bpmn | ✅ LOLOS | 13 (8 userTask + 5 serviceTask) | 3 | 3 | 1 | 1 | 20 | 21 | 20 | 1 |
| 03-pelaksanaan-ujian-siswa.bpmn | ✅ LOLOS | 18 (7 userTask + 11 serviceTask) | 8 | 4 | 1 | 4 | 36 | 35 | 36 | 1 |
| 04-penilaian-koreksi-esai.bpmn | ✅ LOLOS | 8 (2 userTask + 6 serviceTask) | 3 | 2 | 1 | 2 | 15 | 16 | 15 | 1 |
| 05-rekap-berita-acara.bpmn | ✅ LOLOS | 7 (2 userTask + 5 serviceTask) | 2 | 3 | 1 | 1 | 12 | 14 | 12 | 1 |
| 06-ukk.bpmn | ✅ LOLOS | 12 (7 userTask + 5 serviceTask) | 2 | 3 | 1 | 2 | 17 | 20 | 17 | 1 |
| 07-import-soal.bpmn | ✅ LOLOS | 10 (7 userTask + 3 serviceTask) | 3 | 2 | 1 | 2 | 17 | 18 | 17 | 1 |

Bukti per file (to-be): XML well-formed ✅; namespace BPMN 2.0.2 ✅;
`targetNamespace = http://cbt-smk-v2/bpmn/to-be` ✅;
`exporter = "CBT SMK v2 Tim To-Be BPMN"`, `exporterVersion = "1.0"` ✅;
semua `sequenceFlow` merujuk ID yang ada ✅; konektivitas incoming/outgoing lengkap ✅;
tepat 1 start, ≥ 1 end ✅; semua flow node + lane punya `BPMNShape`, semua `sequenceFlow`
punya `BPMNEdge` (masing-masing ≥ 2 waypoint), tanpa DI menggantung ✅; tanpa tumpang tindih
bounding box ✅; semua gateway berlabel pertanyaan (diakhiri `?`) ✅; semua task berlabel
kata kerja + objek Bahasa Indonesia ✅; `documentation` di level `process` ada dan terisi ✅;
semua outgoing flow gateway punya nama/kondisi ✅. Semua `process` to-be memakai
`isExecutable="false"`. File to-be memakai default namespace (tanpa prefix), task bertipe
`userTask`/`serviceTask`, dan TANPA elemen `collaboration`/`participant` (tidak ada pool).

Catatan hitung: kolom "Task" pada as-is menghitung `<bpmn:task>` generik; pada to-be menghitung
`userTask + serviceTask`. Kolom BPMNShape mencakup shape untuk flow node + lane (+ participant
pada as-is); selisih Shape vs (flow node + lane) adalah shape participant pada as-is.

## Upaya unduh XSD

Validasi terhadap XSD BPMN 2.0.2 resmi OMG **tidak dilakukan**, karena file XSD tidak berhasil
diperoleh. URL yang dicoba (via `curl`, 17 Okt 2026):

| URL yang dicoba | Hasil |
|---|---|
| `https://www.omg.org/spec/BPMN/20100524/BPMN20.xsd` | HTTP 404 — `{"status":404,"error":"Not Found"}` |
| `https://www.omg.org/spec/BPMN/2.0.2/BPMN20.xsd` | HTTP 404 |
| `https://www.omg.org/spec/BPMN/20100524/MODEL/BPMN20.xsd` | HTTP 404 |
| `https://www.omg.org/spec/BPMN/20100524/MODEL.xsd` | HTTP 404 |
| `https://www.omg.org/spec/BPMN/20100524/DI.xsd` | HTTP 404 |
| `https://www.omg.org/spec/BPMN/20100524/DC.xsd` | HTTP 404 |
| `https://www.omg.org/spec/BPMN/2.0.2/MODEL.xsd` | HTTP 404 |
| `https://www.omg.org/spec/BPMN/20100524/BPMN20.zip` | HTTP 404 |
| `https://www.omg.org/spec/BPMN/2.0.2/BPMN20.zip` | HTTP 404 |
| `https://raw.githubusercontent.com/omg-spec/BPMN/refs/heads/master/BPMN20.xsd` | Timeout (25 dtk, 0 byte) — repo/mirror tidak terkonfirmasi |

Penyebab yang teramati: situs `omg.org` telah direstrukturisasi sehingga path XSD lama tidak
lagi dilayani (semua mengembalikan JSON 404 generik). **Jangan mengklaim ke-14 file "valid
terhadap XSD OMG"** — klaim yang benar adalah: ke-14 file lolos seluruh cek struktural,
namespace, DI, dan label di atas. Bila koordinator menghendaki validasi XSD penuh, unduh
`BPMN20.xsd` (+ `DI.xsd`/`DC.xsd` dependensinya) dari salinan spesifikasi resmi OMG
(dokumen `formal/2013-12-03` atau paket spesifikasi BPMN 2.0.2) lalu jalankan
`lxml.etree.XMLSchema` terhadap ke-14 file — `lxml 5.2.1` tersedia di lingkungan ini.

## Temuan

**Tidak ada temuan yang menyebabkan GAGAL.** Daftar di bawah adalah observasi non-blocking
untuk koordinator (tidak memerlukan perbaikan diagram):

1. **Perbedaan gaya penulisan as-is vs to-be (konsistensi, bukan kesalahan).**
   As-is memakai prefix `bpmn:` + `collaboration`/`participant` + task generik `<bpmn:task>`;
   to-be memakai default namespace (tanpa prefix) + task bertipe `userTask`/`serviceTask`
   tanpa `collaboration`/`participant`. Keduanya sah menurut BPMN 2.0.2. Bila koordinator
   menginginkan keseragaman (mis. satu gaya untuk dibuka di tool yang sama), penyeragaman
   dapat dilakukan sebagai tugas terpisah — bukan perbaikan validasi.
2. **Satu label mengandung istilah Inggris (tetap LOLOS).**
   To-be `05-rekap-berita-acara.bpmn`: task *"Cetak Berita Acara Print-Friendly"*.
   Istilah "Print-Friendly" adalah istilah domain produk (fitur cetak di aplikasi), dan
   sisanya tetap Bahasa Indonesia + kata kerja + objek. Boleh dibiarkan atau diubah menjadi
   "Cetak Berita Acara Ramah Cetak" bila diinginkan konsistensi bahasa penuh.
3. **Tidak ada end event bernama "Selesai" generik yang ambigu pada to-be** — end event
   to-be sudah diberi nama hasil spesifik (mis. "Selesai: Bank Terkunci",
   "Selesai: Attempt Diblokir"). As-is memakai "Selesai" generik (wajar untuk proses manual).
   Bukan temuan negatif; dicatat sebagai bukti kelengkapan label.

## Kesimpulan

- **14 dari 14 file LOLOS** seluruh cek well-formed XML, namespace BPMN 2.0.2, referensi
  `sequenceFlow`, konektivitas, jumlah start/end, kelengkapan DI (Shape/Edge/waypoint),
  bebas tumpang tindih DI, dan konsistensi label.
- **Validasi XSD resmi OMG belum dilakukan** karena XSD tidak dapat diunduh (10 URL dicoba:
  9× HTTP 404 dari omg.org, 1× timeout mirror). Status ini dicatat jujur dan bukan
  diklaim sebagai validasi XSD.
- **Tidak ada file yang diperbaiki** selama validasi (tidak ada yang gagal parse; tidak ada
  perubahan isi diagram maupun kode aplikasi).
