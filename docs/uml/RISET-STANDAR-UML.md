# RISET — STANDAR UML TERPADU untuk Audit & Standardisasi Diagram CBT SMK v2

- **Dokumen:** `docs/uml/RISET-STANDAR-UML.md`
- **Tanggal sintesis:** 17 Oktober 2026
- **Tim:** TIM SINTESIS — PENYUSUN DOKUMEN ATURAN STANDAR UML
- **Sumber SATU-SATUNYA dokumen ini:** tiga dokumen riset yang sudah ada —
  `docs/uml/riset/R1-omg-uml251.md`, `docs/uml/riset/R2-jurnal.md`,
  `docs/uml/riset/R3-skripsi.md`. **Tidak ada aturan yang dikarang, tidak ada
  sumber baru yang ditambahkan.** Setiap aturan di bawah menunjuk ke ID dan
  sumber di ketiga dokumen tersebut.
- **File `docs/uml/cbt-smk-v2.mdj` TIDAK diubah** dalam penyusunan dokumen ini.
- **Cara pakai:** §1 = temuan WAJIB (pelanggaran formal bila dilanggar).
  §2 = temuan KUALITAS (rekomendasi berbasis bukti, bukan pelanggaran formal
  OMG). §3 = konvensi NASKAH skripsi (format & kelengkapan, bukan notasi).
  §4 = daftar hal yang **DILARANG** dijadikan temuan. Bila ragu, lihat §6
  (hierarki bukti).

> ⚠️ **PERHATIAN — 3 KOREKSI PREMIS (lihat §5 untuk detail lengkap).**
> Tiga anggapan yang selama ini dipakai sebagai "aturan" ternyata **SALAH**
> menurut spesifikasi OMG UML 2.5.1 dan **TIDAK BOLEH** lagi dijadikan temuan:
>
> 1. **Penomoran pesan Sequence `1`, `1.1`, `2` — BUKAN bagian Sequence Diagram.**
>    Penomoran hanya untuk Communication Diagram.
> 2. **Reply TIDAK WAJIB untuk setiap `synchCall`.** Spesifikasi hanya berkata
>    "Typically". SynchCall tanpa reply = tidak lengkap, BUKAN pelanggaran formal.
> 3. **Guard TIDAK WAJIB di tiap cabang Decision.** Guard per edge bersifat
>    opsional; edge tanpa guard = selalu true. Yang mengikat: satu token paling
>    banyak satu cabang + anjuran determinisme.
>
> Setiap temuan audit lama yang bersandar pada ketiga premis di atas
> **BATAL** dan harus ditarik.

---

## 1. Aturan WAJIB (spek OMG UML 2.5.1)

Sumber: R1 — OMG Unified Modeling Language, Version 2.5.1,
OMG Document Number **formal/2017-12-05**, Desember 2017,
https://www.omg.org/spec/UML/2.5.1/PDF.
Semua aturan di bawah berstatus TERVERIFIKASI di R1 (75 aturan).
Cara cek default: periksa model/diagram CBT SMK v2 terhadap bunyi aturan;
bila bunyi aturan dilanggar, itu pelanggaran formal.

### 1.1 Sequence Diagram — Message, MessageSort, Occurrence, Execution, Lifeline

| ID | Aturan | Sumber (OMG UML 2.5.1) | Cara cek |
|---|---|---|---|
| SEQ-01 | `MessageSort` hanya punya literal: `synchCall`, `asynchCall`, `asynchSignal`, `createMessage`, `deleteMessage`, `reply`. Tidak ada literal lain (`self`, `found`, `lost` sebagai sort tidak ada — found/lost adalah `MessageKind`). | §17.12.22.3; §17.12.20.3 | Periksa tiap pesan: sort-nya salah satu dari 6 literal tersebut. |
| SEQ-02 | `synchCall` = pesan dari pemanggilan sinkron sebuah Operation. Argumennya = parameter `in` dan `inout` Operation, sesuai urutan `ownedParameter`. | §17.12.22.3; §17.4.3.1 | Cocokkan argumen pesan synchCall dengan parameter `in`/`inout` Operation yang dirujuk, sesuai urutan. |
| SEQ-03 | `asynchCall` = pesan dari pemanggilan asinkron sebuah Operation, yaitu `CallAction` dengan `isSynchronous = false`. Pengirim TIDAK menunggu. | §17.12.22.3 | Periksa pesan asinkron: sumbernya CallAction asinkron; pengirim tidak menunggu balasan. |
| SEQ-04 | `reply` = pesan balasan (return) dari pemanggilan sinkron sebuah Operation. Argumennya = parameter `out`, `inout`, dan `return` Operation, sesuai urutan `ownedParameter`. | §17.4.3.1; §17.12.22.3 | Cocokkan argumen reply dengan parameter `out`/`inout`/`return` Operation, sesuai urutan. |
| SEQ-05 | **Reply TIDAK WAJIB untuk setiap `synchCall`.** Tidak ada constraint di §17.12.18.7 yang mewajibkan pasangan reply untuk tiap synchCall. Spesifikasi hanya berkata "Typically" start/finish ExecutionSpecification = receive + send reply (§17.2.3.4) — "typically", bukan "must". SynchCall tanpa reply = model tidak lengkap / perlu ditanya, BUKAN pelanggaran formal. | §17.12.18.7 (tidak adanya constraint); §17.2.3.4 | Jangan catat sebagai pelanggaran formal. Bila ditemukan, catat sebagai pertanyaan kelengkapan model. Lihat §5. |
| SEQ-06 | Jika signature Message adalah Signal, maka `messageSort` HARUS `asynchSignal`, dan Message = pengiriman + penerimaan asinkron sebuah instance Signal. Argumen = atribut-atribut Signal. | §17.4.3.1; §17.12.18.7 `signature_refer_to` | Periksa pesan ber-signature Signal: sort-nya `asynchSignal`, argumennya = atribut Signal. |
| SEQ-07 | `createMessage` = pesan yang menandai pembuatan (creation) objek lifeline lain. Notasi: garis putus-putus dengan kepala panah terbuka (open arrowhead). | §17.12.22.3; §17.4.4.1 | Periksa pesan create: garis putus-putus + panah terbuka, menuju lifeline yang dibuat. |
| SEQ-08 | `deleteMessage` = pesan yang menandai terminasi (penghancuran) lifeline lain. **Harus** berakhir di sebuah `DestructionOccurrenceSpecification`. | §17.12.22.3; §17.4.4.1 | Periksa pesan delete: ujungnya adalah DestructionOccurrenceSpecification. |
| SEQ-09 | Signature Message harus merujuk ke Operation (maka `messageSort` = `synchCall` / `asynchCall` / `reply`) ATAU ke Signal (maka `messageSort` = `asynchSignal`). Nama Message HARUS sama dengan nama NamedElement yang dirujuk signature. | §17.12.18.7 `signature_refer_to` | Cocokkan tiap pesan: signature-nya Operation atau Signal, sort-nya sesuai, namanya sama dengan elemen yang dirujuk. |
| SEQ-10 | Korespondensi argumen: Message `synchCall`/`asynchCall` ber-signature Operation → argumen = parameter `in` + `inout`; Message `reply` → argumen = parameter `out` + `inout` + `return`. Tipe argumen harus conform ke tipe parameter. | §17.12.18.7 `signature_is_operation_request`, `signature_is_operation_reply` | Cocokkan daftar dan tipe argumen pesan dengan parameter Operation. |
| SEQ-11 | Garis Message: setiap fragmen garis harus horizontal atau menurun bila ditelusuri dari send event ke receive event. Send dan receive event BOLEH berada pada lifeline yang sama (self-message). | §17.4.4.1 | Periksa tiap garis pesan: tidak ada segmen yang naik dari send ke receive. |
| SEQ-12 | Message TIDAK BOLEH melintasi batas CombinedFragment / operand-nya: kedua MessageEnd harus tertutup dalam InteractionFragment yang sama (satu InteractionOperand atau satu Interaction). | §17.12.18.7 `cannot_cross_boundaries` | Periksa tiap pesan: send dan receive berada dalam operand/fragment yang sama. |
| SEQ-13 | Self-message (send + receive di lifeline yang sama): sendEvent HARUS terurut sebelum receiveEvent pada lifeline tersebut. | §17.12.18.7 `sending_receiving_message_event` | Periksa self-message: send di atas receive pada lifeline yang sama. |
| SEQ-14 | Kepala panah: asinkron (`asynchCall`/`asynchSignal`) = panah terbuka; sinkron (`synchCall`) = panah terisi (filled); reply = garis putus-putus dengan panah terbuka ATAU terisi; create = garis putus-putus dengan panah terbuka. | §17.4.4.1 | Periksa kepala panah + gaya garis tiap pesan sesuai sort-nya. |
| SEQ-15 | Lost Message = lingkaran hitam kecil di ujung panah (receive tidak diketahui). Found Message = lingkaran hitam kecil di ujung awal (send tidak diketahui). | §17.4.4.1; §17.12.20.3 | Periksa lost/found message: lingkaran hitam di ujung yang benar. |
| SEQ-16 | Sintaks label Message: request = `<nama-pesan>[(daftar-argumen-input)]`; reply = `[<assignment-target> =] <nama-pesan>[(daftar-argumen-output)] [: <value>]`. Label reply BOLEH dihilangkan bila identitasnya sudah jelas (mis. satu-satunya reply dalam satu ExecutionSpecification). | §17.4.4.1 | Periksa format label pesan; label reply yang hilang hanya dapat diterima bila identitasnya jelas. |
| SEQ-17 | **Penomoran pesan (`1`, `1.1`, dst.) BUKAN bagian notasi Sequence Diagram.** Sintaks label Message di §17.4.4.1 tidak memuat nomor urut. Penomoran berurutan (`sequence-expression`, mis. `1b.1`, `3.1.4`) hanya didefinisikan untuk **Communication Diagram** (§17.9.1.3; §17.4.4.1). Angka pada Sequence Diagram adalah anotasi tool / konvensi proyek, bukan persyaratan OMG; yang mengikat adalah urutan vertikal (atas → bawah) dan korespondensi pesan. | §17.4.4.1; §17.9.1.3; §17.9 | Jangan nilai angka penomoran sebagai benar/salah terhadap OMG. Periksa urutan vertikal pesan. Lihat §5. |
| SEQ-18 | `OccurrenceSpecification` adalah unit semantik dasar Interaction. Secara notasi ia hanyalah titik sintaktis di ujung-ujung Message atau di awal/akhir sebuah ExecutionSpecification. | §17.12.23.1; §17.2.4.3 | Periksa occurrence: berada di ujung Message atau awal/akhir ExecutionSpecification. |
| SEQ-19 | `ExecutionSpecification` = spesifikasi eksekusi satu unit Behavior/Action di dalam Lifeline. Durasinya diwakili DUA OccurrenceSpecification: start dan finish. Start dan finish HARUS pada Lifeline yang sama (`start.covered = finish.covered`). | §17.12.8.1; §17.12.8.6 `same_lifeline` | Periksa tiap ExecutionSpecification: start dan finish pada lifeline yang sama. |
| SEQ-20 | Notasi ExecutionSpecification (activation bar): persegi panjang tipis (abu-abu atau putih) yang menutupi garis lifeline. ExecutionSpecification yang tumpang tindih pada lifeline yang sama digambar sebagai persegi panjang yang tumpang tindih. | §17.2.4.4 | Periksa activation bar: persegi panjang tipis menutupi garis lifeline; yang tumpang tindih digambar tumpang tindih. |
| SEQ-21 | Semantik trace sebuah Execution = `<start, finish>`. "Typically" start = receive OccurrenceSpecification sebuah Message dan finish = send OccurrenceSpecification sebuah reply Message — pola umum, BUKAN keharusan formal (lihat SEQ-05). | §17.2.3.4 | Jangan tuntut pola receive→reply sebagai keharusan; periksa pasangan start/finish. |
| SEQ-22 | `ExecutionOccurrenceSpecification` = momen waktu saat Action/Behavior dimulai atau selesai (titik start/finish dari ExecutionSpecification). | §17.12.7.1 | Periksa titik start/finish ExecutionSpecification. |
| SEQ-23 | Setelah sebuah `DestructionOccurrenceSpecification` pada suatu Lifeline, TIDAK BOLEH ada OccurrenceSpecification lain pada Lifeline yang sama di dalam InteractionOperand yang sama (destruction adalah event terakhir). | §17.12.6.4 `no_occurrence_specifications_below` | Periksa lifeline yang dihancurkan: tidak ada event di bawah destruction dalam operand yang sama. |
| SEQ-24 | `Lifeline` mewakili HANYA SATU entitas yang berinteraksi (meskipun part / structural feature-nya bermultiplisitas > 1). Notasi: kepala persegi panjang + garis vertikal (boleh putus-putus) sebagai garis hidup. Format identifikasi: `[<nama>['[' <selector> ']']] [: <tipe>] [<decomposition>]` atau `self`. Identifikasi lifeline TIDAK BOLEH kosong. | §17.12.17.1; §17.3.4.1 | Periksa tiap lifeline: satu entitas, kepala persegi panjang + garis hidup, identifikasi tidak kosong dan sesuai format. |
| SEQ-25 | Urutan event pada satu Lifeline: terurut linear dari atas ke bawah; jarak absolut antar OccurrenceSpecification TIDAK bermakna semantik (hanya urutan yang bermakna). Pengecualian: di dalam CombinedFragment `par` atau `coregion`, event tidak terurut lokal kecuali dipaksakan GeneralOrdering. | §17.3.3.1 | Periksa urutan event atas→bawah; jangan nilai jarak absolut antar event. |
| SEQ-26 | Selector Lifeline hanya boleh ada bila ConnectableElement yang dirujuk multivalued; nilainya harus `LiteralString` atau `LiteralInteger`. | §17.12.17.5 `selector_specified`, `selector_int_or_string` | Periksa selector: hanya pada elemen multivalued, bernilai LiteralString/LiteralInteger. |

### 1.2 Sequence Diagram — CombinedFragment (alt / opt / loop / break)

| ID | Aturan | Sumber (OMG UML 2.5.1) | Cara cek |
|---|---|---|---|
| CF-01 | Notasi CombinedFragment: persegi panjang bergaris tepi solid; operator ditulis di pentagon (kompartemen kecil) di sudut kiri atas frame. Lebih dari satu operator di pentagon = singkatan untuk CombinedFragment bersarang. | §17.6.4.3; §17.6.4.6 | Periksa frame: tepi solid, operator di pentagon kiri atas. |
| CF-02 | Operand-Operand CombinedFragment dipisahkan garis horizontal putus-putus; bersama-sama membentuk frame CombinedFragment. Dalam satu operand Sequence Diagram, urutan fragment ditentukan posisi vertikal teratas. | §17.6.4.1 | Periksa pemisah operand: garis horizontal putus-putus. |
| CF-03 | `alt` = pilihan perilaku: PALING BANYAK SATU operand dipilih. Operand terpilih HARUS punya guard eksplisit atau implisit yang bernilai true pada titik tersebut. Operand tanpa guard dianggap punya guard implisit `true`. | §17.6.3.7; §17.12.15.3 `alt` | Periksa `alt`: paling banyak satu operand terpilih; operand tanpa guard dibaca guard `true`. |
| CF-04 | Operand `alt` yang di-guard `else` berarti guard = negasi dari disjungsi semua guard lain di CombinedFragment yang sama. | §17.6.3.7 | Periksa operand `else`: maknanya = negasi disjungsi guard lain. |
| CF-05 | `alt`: bila TIDAK ADA operand yang guard-nya true, maka TIDAK ADA operand yang dieksekusi dan sisa enclosing InteractionFragment yang dieksekusi. (`alt` tanpa operand `else` dan semua guard false adalah perilaku yang sah, bukan kesalahan.) | §17.6.3.7 | Jangan catat `alt` tanpa `else` sebagai kesalahan; periksa semantik "tidak ada operand dieksekusi". |
| CF-06 | `opt` = pilihan perilaku dengan SATU operand: operand terjadi ATAU tidak terjadi sama sekali. Semantiknya setara `alt` dengan satu operand berisi + satu operand kosong. | §17.6.3.8; §17.12.15.3 `opt` | Periksa `opt`: tepat satu operand (lihat juga CF-10). |
| CF-07 | `break` = skenario pemutus: operand dijalankan SEBAGAI PENGGANTI sisa enclosing InteractionFragment. Guard true → break dipilih, sisa diabaikan. Guard false → break diabaikan, sisa dipilih. Tanpa guard → pilihan non-deterministik. **Break harus mencakup (cover) SEMUA Lifeline** yang dicakup enclosing InteractionFragment. | §17.6.3.9; §17.12.3.7 `break`; §17.12.15.3 `break` | Periksa `break`: cakupannya semua lifeline enclosing fragment; semantik pengganti sisa. (Catatan: §17.6.3.9 memakai kata "should", constraint OCL §17.12.3.7 memakainya sebagai keharusan — yang dipakai adalah constraint formal; lihat §6.) |
| CF-08 | `loop` = operand diulang sejumlah kali. Guard boleh memuat jumlah iterasi minimum (`minint`) dan maksimum (`maxint`) plus ekspresi Boolean. Loop beriterasi minimal `minint` kali dan maksimal `maxint` kali; setelah iterasi minimum tercapai dan ekspresi Boolean false, loop berhenti. Bila ada InteractionConstraint terpisah dengan spesifikasi, loop hanya berlanjut bila spesifikasi itu true — terlepas dari jumlah minimum. | §17.6.3.17 | Periksa semantik loop: batas iterasi + kondisi berhenti sesuai guard. |
| CF-09 | Sintaks loop: `loop[(<minint>[, <maxint>])]`. `minint` = bilangan asli non-negatif; `maxint` = bilangan asli ≥ `minint`, atau `*` (tak terbatas). Hanya `minint` saja → `minint = maxint`. Hanya `loop` saja → batas bawah 0, batas atas tak terbatas. Constraint: `minint` ≥ 0; `maxint` > 0; bila `maxint` ada maka `minint` harus ada dan `maxint` ≥ `minint`. | §17.6.4.9; §17.12.12.5 `minint_non_negative`, `maxint_positive`, `maxint_greater_equal_minint` | Periksa penulisan loop dan nilai minint/maxint terhadap constraint. |
| CF-10 | Bila `interactionOperator` = `opt`, `loop`, `break`, `assert`, atau `neg`, CombinedFragment HARUS punya tepat SATU operand. (`alt` boleh banyak operand; CombinedFragment umum punya operand `[1..*]`.) | §17.12.3.7 `opt_loop_break_neg`; §17.12.3.6 | Hitung operand pada `opt`/`loop`/`break`/`assert`/`neg`: harus tepat 1. |
| CF-11 | Penulisan guard (InteractionConstraint): `[` (ekspresi-Boolean \| `else`) `]`. Bila InteractionConstraint dihilangkan, diasumsikan `true`. Guard ditampilkan dalam kurung siku menutupi lifeline tempat event pertama akan terjadi, diposisikan di atas event itu, di dalam Interaction / InteractionOperand yang memuatnya. | §17.6.4.2; §17.12.12.1 | Periksa format guard: satu pasang kurung siku, posisi di atas event pertama operand. |
| CF-12 | Guard operand: (a) HARUS ditempatkan tepat sebelum (di atas) OccurrenceSpecification yang akan menjadi OccurrenceSpecification pertama di dalam InteractionOperand itu; (b) hanya boleh merujuk nilai lokal Lifeline tempat guard berada, atau nilai global seluruh Interaction. | §17.12.14.5 `guard_directly_prior`, `guard_contain_references` | Periksa posisi guard (tepat di atas occurrence pertama) dan rujukannya (lokal lifeline itu / global Interaction). |
| CF-13 | `minint` / `maxint` hanya boleh ada bila InteractionConstraint terkait dengan operand sebuah CombinedFragment `loop`. | §17.12.12.5 `minint_maxint` | Periksa minint/maxint: hanya pada operand `loop`. |

### 1.3 Use Case Diagram — Include / Extend / Asosiasi Aktor

| ID | Aturan | Sumber (OMG UML 2.5.1) | Cara cek |
|---|---|---|---|
| UC-01 | Setiap UseCase HARUS punya nama. | §18.2.5.6 `must_have_name` | Periksa tiap UseCase: nama tidak kosong. |
| UC-02 | Setiap Actor HARUS punya nama. Actor hanya boleh ber-Asosiasi ke UseCase, Component, dan Class; asosiasi tersebut HARUS biner (2 ujung). | §18.2.1.4 `must_have_name`, `associations` | Periksa tiap Actor: nama tidak kosong; asosiasinya hanya ke UseCase/Component/Class dan biner. |
| UC-03 | UseCase hanya boleh terlibat dalam Asosiasi biner. Dua UseCase TIDAK BOLEH ber-Asosiasi bila keduanya menspesifikasikan subject yang sama. | §18.2.5.6 `binary_associations`, `no_association_to_use_case` | Periksa asosiasi yang melibatkan UseCase: biner; tidak ada asosiasi antar dua UseCase dengan subject sama. |
| UC-04 | Sebuah UseCase TIDAK BOLEH meng-include UseCase yang (langsung maupun tidak langsung) meng-include dirinya — tidak boleh ada siklus include. | §18.2.5.6 `cannot_include_self` | Telusuri rantai include: tidak ada siklus. |
| UC-05 | `Include` = DirectedRelationship antara dua UseCase: perilaku UseCase yang di-include (`addition`) disisipkan ke perilaku UseCase yang meng-include (`includingCase`). `includingCase` adalah source sekaligus pemilik relasi; `addition` adalah target. Included UseCase harus tersedia agar perilaku including UseCase terdeskripsikan lengkap. Kegunaan utama = reuse bagian perilaku yang sama; akibatnya base UseCase yang meng-include biasanya TIDAK lengkap sendiri — bergantung pada bagian yang di-include. Seluruh perilaku included UseCase dieksekusi pada SATU lokasi di dalam including UseCase sebelum eksekusi including dilanjutkan. | §18.1.3.3; §18.2.4.1–4 | Periksa relasi include: arah, source/target, pemilik relasi; addition tersedia dan dipakai sebagai bagian reuse. |
| UC-06 | Notasi Include: panah putus-putus dengan kepala panah terbuka, dari UseCase dasar (including) menuju UseCase yang di-include, berlabel kata kunci `«include»`. | §18.1.4 | Periksa notasi include: garis putus-putus, panah terbuka, arah including → addition, label `«include»`. |
| UC-07 | `Extend` = relasi dari extending UseCase (`extension`) ke extended UseCase (`extendedCase`) yang menentukan bagaimana dan kapan perilaku extending dapat disisipkan ke perilaku extended, pada satu atau lebih extension point yang didefinisikan di extended UseCase. Extended UseCase didefinisikan secara independen dan bermakna sendiri; extending UseCase biasanya TIDAK bermakna sendiri. Bila condition Extend hilang atau bernilai true saat extension point pertama dicapai, seluruh fragmen perilaku extending yang sesuai ikut dieksekusi; bila false, tidak terjadi. | §18.1.3.2; §18.2.2.1–4 | Periksa relasi extend: extension point milik extendedCase, extended bermakna sendiri, semantik condition. |
| UC-08 | Notasi Extend: panah putus-putus dengan kepala panah terbuka, dari extending UseCase menuju extended UseCase, berlabel `«extend»`. Condition Extend dan referensi ExtensionPoint secara OPSIONAL ditunjukkan dalam note yang ditempel pada panah. | §18.1.4 | Periksa notasi extend: garis putus-putus, panah terbuka, arah extension → extendedCase, label `«extend»`; note condition opsional. |
| UC-09 | Relasi Extend DIMILIKI oleh extending UseCase (`extension`). Daftar `extensionLocation` = `[1..*]` ExtensionPoint yang HARUS milik extended UseCase (constraint `extension_points`). Satu extending UseCase boleh meng-extend lebih dari satu UseCase; extending UseCase sendiri boleh di-extend. | §18.2.2.4–5; §18.1.3.2 | Periksa extend: pemilik relasi = extension; extensionLocation ≥ 1 dan milik extendedCase. |
| UC-10 | Asosiasi Aktor–UseCase dengan multiplisitas: multiplisitas > 1 di ujung Actor = lebih dari satu instance Actor terlibat dalam UseCase (cara pastinya — serentak atau berurutan — TIDAK didefinisikan spesifikasi). Multiplisitas > 1 di ujung UseCase = satu Actor dapat terlibat dalam banyak UseCase tipe itu. | §18.1.3.1 | Periksa multiplisitas asosiasi aktor–use case dibaca sesuai aturan ini; jangan klaim cara "serentak vs berurutan". |
| UC-11 | Notasi Actor: ikon "stick man" + nama di dekatnya; ATAU persegi panjang Classifier dengan kata kunci `«actor»`; ATAU ikon lain yang menyampaikan jenis Actor (mis. Actor non-manusia). | §18.1.4 | Periksa notasi tiap Actor: salah satu dari tiga bentuk tersebut. |
| UC-12 | Notasi UseCase: elips berisi nama UseCase, atau nama di bawah elips; kata kunci stereotip opsional di atas nama. ExtensionPoint dapat dicantumkan di kompartemen UseCase dengan judul `extension points`, format `<nama> [: <penjelasan>]`. Subject (system boundary) = persegi panjang dengan nama di sudut kiri atas; UseCase yang berlaku bagi subject digambar di dalamnya. | §18.1.4 | Periksa notasi UseCase: elips + nama; extension point di kompartemen `extension points` bila dicantumkan; subject = persegi panjang bernama di kiri atas. |

### 1.4 Class Diagram — Asosiasi, Multiplisitas, Visibilitas, Navigability

| ID | Aturan | Sumber (OMG UML 2.5.1) | Cara cek |
|---|---|---|---|
| CL-01 | Sebuah Association punya PALING SEDIKIT DUA memberEnd, masing-masing diwakili Property yang tipenya adalah tipe ujung tersebut. Lebih dari satu ujung boleh punya tipe yang sama. | §11.5.3.1 | Periksa tiap Association: ≥ 2 memberEnd, tiap ujung bertipe sesuai ujungnya. |
| CL-02 | Multiplisitas ujung asosiasi membatasi kardinalitas koleksi instance di ujung tersebut: pilih N−1 ujung sebagai konteks; koleksi link yang merujuk instance-instance konteks mengidentifikasi sekumpulan instance di ujung sisanya (oep); multiplisitas oep membatasi kardinalitas koleksi itu. Praktisnya: **angka yang tertulis di ujung X membatasi berapa banyak instance di sisi X yang dapat terhubung ke satu instance di sisi lawan** (per ujung, dibaca terhadap ujung tersebut). | §11.5.3.1 | Baca tiap angka multiplisitas terhadap ujungnya sendiri; jangan membalik pembacaan. |
| CL-03 | Notasi multiplisitas: `<lower>.. <upper>`; `*` = batas atas tak terbatas; bila lower = upper cukup tulis satu angka (`1` ≡ `1..1`); `*` sendiri ≡ `0..*`. Pada elemen tekstual (atribut) multiplisitas ditulis DALAM kurung siku `[...]`; pada elemen simbol (ujung asosiasi) ditulis TANPA kurung siku, diletakkan dekat simbol elemen. | §7.5.4.1 | Periksa penulisan multiplisitas: atribut pakai `[...]`, ujung asosiasi tanpa `[...]`. |
| CL-04 | Bila TIDAK ADA multiplisitas yang ditampilkan pada diagram, TIDAK BOLEH disimpulkan apa pun tentang multiplisitas di model. (Ujung asosiasi tanpa angka = "tidak diperiksa", BUKAN "salah" dan BUKAN "default 1".) | §11.5.4 (NOTE) | Jangan catat ujung tanpa angka sebagai salah / default 1. |
| CL-05 | Notasi navigability: kepala panah terbuka di ujung asosiasi = ujung itu navigable; tanda `x` kecil di ujung = ujung itu TIDAK navigable. Titik (dot) kecil di ujung = kepemilikan ujung oleh Classifier di ujung lawan. Agregasi, navigability, dan kepemilikan ujung adalah TIGA KONSEP TERPISAH dengan notasinya masing-masing; konvensi lama (ujung non-navigable = dimiliki Association) sudah deprecated. Ujung asosiasi yang dimiliki Class SELALU navigable; ujung yang dimiliki Association boleh navigable atau tidak. | §11.5.4 | Periksa notasi ujung: panah terbuka / `x` / dot dibaca sebagai tiga konsep terpisah; jangan samakan non-navigable dengan dimiliki Association. |
| CL-06 | Semantik navigability: ujung Property asosiasi yang dimiliki oleh Class ujungnya, atau yang merupakan `navigableOwnedEnd` dari Association, menandakan Association navigable dari ujung-ujung lawannya. Navigability berarti instance di ujung lawan dapat mengakses instance peserta secara efisien pada runtime. Bila ujung tidak navigable, akses dari ujung lain mungkin tetap bisa, tetapi mungkin tidak efisien. | §11.5.3.1 | Periksa kepemilikan ujung vs navigability; jangan klaim ujung non-navigable = tidak dapat diakses sama sekali. |
| CL-07 | Agregasi: `shared` = belah ketupat KOSONG di ujung garis asosiasi yang BERLAWANAN dengan ujung yang ditandai `shared`; `composite` = belah ketupat TERISI di ujung yang berlawanan dengan ujung `composite`. Ujung Property asosiasi hanya boleh ditandai agregasi shared/composite bila Association biner DAN ujung lainnya tidak ditandai shared/composite. | §11.5.4; §11.5.3.1 | Periksa belah ketupat: kosong = shared, terisi = composite, di ujung berlawanan dengan ujung yang ditandai; hanya pada asosiasi biner dengan ujung lain tak bertanda. |
| CL-08 | `VisibilityKind`: `public` = terlihat semua elemen yang dapat mengakses isi Namespace pemilik; `private` = hanya terlihat di dalam Namespace pemilik; `protected` = terlihat oleh elemen yang punya relasi generalisasi ke Namespace pemilik; `package` = terlihat semua elemen dalam Package terdekat yang mengurungnya. | §7.8.24.3 | Periksa makna visibilitas yang dipakai sesuai empat definisi ini. |
| CL-09 | Notasi Property (atribut): `[<visibility>] [/] <nama> [: <tipe>] [[<multiplisitas>]] [= <default>] [{<modifier>}...]` dengan `<visibility>` ∈ `+` `-` `#` `~` (public, private, protected, package). `/` = derived. Bila multiplisitas dihilangkan, artinya tepat 1. | §9.5.4 | Periksa penulisan atribut sesuai format; atribut tanpa multiplisitas dibaca tepat 1 (berbeda dengan ujung asosiasi — lihat CL-04). |
| CL-10 | Notasi Operation: `[<visibility>] <nama> ([<daftar-parameter>]) [: [<tipe-return>] [[<multiplisitas>]]] [{<properti>}]`, `<visibility>` ∈ `+` `-` `#` `~`. Bila ada Parameter return, tipe Operation = tipe Parameter return tersebut; bila tidak ada, Operation tidak punya tipe. | §9.6.4; §9.6.3.1 | Periksa penulisan operasi sesuai format; tipe operasi = tipe parameter return bila ada. |
| CL-11 | Asosiasi biner normalnya digambar sebagai garis solid yang menghubungkan dua Classifier (atau satu Classifier ke dirinya sendiri — kedua ujungnya tetap distinct). Asosiasi dengan > 2 ujung HANYA dapat digambar sebagai diamond dengan garis solid ke tiap Classifier ujung. | §11.5.4 | Periksa gambar asosiasi: biner = garis solid; > 2 ujung = diamond + garis solid ke tiap ujung. |
| CL-12 | Nama Association ditampilkan sebagai string dekat simbol Association, tetapi tidak terlalu dekat ke ujung hingga tertukar dengan nama ujung. Garis miring `/` di depan nama (atau menggantikan nama) menandai Association derived. Panah segitiga solid di samping nama Association hanya untuk dokumentasi arah baca dan TIDAK punya interpretasi semantik umum. | §11.5.4 | Periksa nama asosiasi: posisi tidak tertukar dengan nama ujung; `/` = derived; panah segitiga baca = dokumentasi, bukan semantik. |

### 1.5 Activity Diagram — Decision / Merge / Fork / Join / Initial / Final / Guard

| ID | Aturan | Sumber (OMG UML 2.5.1) | Cara cek |
|---|---|---|---|
| AC-01 | Guard ActivityEdge: sebuah guard adalah ValueSpecification yang dievaluasi untuk SETIAP token yang ditawarkan ke edge. Tawaran hanya boleh melewati edge bila guard bernilai true untuk token itu. Edge TANPA guard setara dengan guard yang selalu true. Guard diizinkan pada ActivityEdge mana pun (tidak hanya di DecisionNode). | §15.2.3.3 | Periksa tiap edge: token hanya lewat bila guard true; edge tanpa guard dibaca selalu true. |
| AC-02 | Notasi guard: teks dalam kurung siku `[…]` di dekat ekor (tail) garis edge. | §15.2.4 | Periksa penulisan guard: satu pasang kurung siku di dekat ekor edge. |
| AC-03 | **Decision TIDAK WAJIB punya guard di tiap cabang keluar.** Spesifikasi: "If any of the outgoing edges of a DecisionNode have guards, then these are evaluated…" — guard bersifat opsional per edge; edge tanpa guard diperlakukan sebagai guard true. Yang diwajibkan: (a) DecisionNode punya 1–2 incoming edge dan ≥ 1 outgoing edge; (b) SETIAP token yang ditawarkan pada incoming primer PALING BANYAK melewati SATU outgoing edge (token tidak diduplikasi); (c) token tidak boleh melewati edge yang guard-nya false; (d) bila beberapa outgoing edge tidak ber-guard / ber-guard true, token melewati paling banyak satu di antaranya (bila beberapa target menerima simultan, edge yang dilalui TIDAK ditentukan spesifikasi — non-deterministik). Untuk menghindari non-determinisme, modeler **sebaiknya** ("should") mengatur agar paling banyak satu guard true per token — ini anjuran, bukan constraint. Guard khusus `else` hanya untuk DecisionNode, PALING BANYAK pada SATU outgoing edge, dan bernilai true hanya bila token tidak diterima edge keluar lainnya. | §15.3.3.6 | Jangan catat cabang tanpa guard sebagai pelanggaran. Periksa: jumlah incoming/outgoing, satu-token-satu-cabang, `else` paling banyak satu. Lihat §5. |
| AC-04 | MergeNode: menggabungkan banyak aliran TANPA sinkronisasi. HARUS punya tepat SATU outgoing edge dan boleh banyak incoming edge. Semua token yang ditawarkan pada incoming edge ditawarkan ke outgoing edge; tidak ada sinkronisasi / penggabungan token. Merge TIDAK memilih berdasarkan guard — guard pada edge MASUK merge (bila ada) hanya berlaku sebagai guard edge biasa (AC-01), bukan sebagai kriteria pilihan merge. Notasi: simbol belah ketupat (diamond); Merge yang dipakai harus punya ≥ 2 incoming + 1 outgoing. | §15.3.3.5; §15.3.4.3 | Periksa MergeNode: tepat 1 outgoing, ≥ 2 incoming, tanpa sinkronisasi; guard edge masuk bukan kriteria pilihan. |
| AC-05 | InitialNode: titik awal eksekusi Activity. TIDAK BOLEH punya incoming ActivityEdge. Semua outgoing edge HARUS ControlFlow. Satu token kontrol ditempatkan pada tiap InitialNode saat Activity mulai dieksekusi. Activity BOLEH punya lebih dari satu InitialNode → tiap InitialNode memulai satu aliran kontrol konkuren. Notasi: lingkaran solid. | §15.3.3.1; §15.3.4.1 | Periksa InitialNode: tanpa incoming, outgoing semuanya ControlFlow, notasi lingkaran solid. |
| AC-06 | FinalNode: TIDAK BOLEH punya outgoing ActivityEdge. Dua jenis: `FlowFinalNode` = mengakhiri SATU aliran saja (token yang diterimanya dihancurkan; aliran lain tidak terpengaruh). `ActivityFinalNode` = menghentikan SEMUA aliran dalam Activity; token pertama yang mencapainya mengakhiri eksekusi Activity. Notasi: ActivityFinal = lingkaran solid di dalam lingkaran hollow ("bull's eye"); FlowFinal = lingkaran berisi tanda X. | §15.3.3.2; §15.3.4.1 | Periksa FinalNode: tanpa outgoing; bedakan FlowFinal (X) vs ActivityFinal (bull's eye) dan semantiknya. |
| AC-07 | ForkNode: memecah SATU aliran menjadi banyak aliran konkuren. HARUS punya tepat SATU incoming edge dan boleh banyak outgoing edge. Token ditawarkan ke SEMUA outgoing edge; tiap penerima mendapat SALINAN token. Bila incoming = ControlFlow maka semua outgoing = ControlFlow; bila incoming = ObjectFlow maka semua outgoing = ObjectFlow. Notasi: segmen garis (bar). PERINGATAN spesifikasi: bila guard dipakai pada edge keluar ForkNode, modeler harus memastikan tidak ada JoinNode hilir yang bergantung pada kedatangan token lewat edge ber-guard itu; bila tidak dapat dihindari, sisipkan DecisionNode antara ForkNode dan edge ber-guard. | §15.3.3.3; §15.3.4.2 | Periksa ForkNode: tepat 1 incoming, token disalin ke semua outgoing, jenis aliran incoming = outgoing; guard pada edge keluar Fork diperiksa terhadap peringatan ini. |
| AC-08 | JoinNode: mensinkronkan banyak aliran. HARUS punya tepat SATU outgoing edge dan boleh banyak incoming edge. Tanpa `joinSpec`, default = operator Boolean "and": token harus ditawarkan pada SETIAP incoming edge (semua cabang masuk harus tiba) sebelum token dikeluarkan. `joinSpec` eksplisit dinotasikan `{joinSpec = <value-specification>}` di dekat simbol JoinNode. Bila ada incoming ObjectFlow, outgoing HARUS ObjectFlow; bila tidak, outgoing = ControlFlow. Notasi: segmen garis (bar), sama seperti ForkNode — dibedakan dari jumlah incoming/outgoing. | §15.3.3.4; §15.3.4.2 | Periksa JoinNode: tepat 1 outgoing, sinkronisasi "and" bila tanpa joinSpec, jenis outgoing mengikuti incoming. |
| AC-09 | Simbol gabungan: satu simbol diamond yang sama dapat dipakai untuk Merge + Decision (modelnya = MergeNode dengan semua incoming edge + satu edge ke DecisionNode dengan semua outgoing edge; paling banyak satu incoming boleh dianotasi `decisionInputFlow`). Satu simbol bar yang sama dapat dipakai untuk Join + Fork (modelnya = JoinNode + satu edge ke ForkNode). Audit harus memodelkan simbol gabungan sebagai DUA node, bukan satu. | §15.3.4.3; §15.3.4.2 | Bila ada diamond/bar gabungan, audit sebagai dua node; periksa aturan masing-masing node. |
| AC-10 | Token tidak boleh "beristirahat" di ControlNode — ControlNode hanya mengatur aliran token — dengan pengecualian InitialNode dan ForkNode (yang dapat menahan token bila tawaran terblokir). | §15.2.3.2 (dan §15.3.3.1, §15.3.3.3) | Periksa aliran token: tidak ada token yang berhenti di ControlNode selain Initial/Fork yang tawarannya terblokir. |
| AC-11 | DecisionNode: bila incoming primer = ControlFlow, semua outgoing HARUS ControlFlow; bila incoming primer = ObjectFlow, semua outgoing HARUS ObjectFlow. Bila DecisionNode punya 2 incoming edge, satu HARUS diidentifikasi sebagai `decisionInputFlow`, yang lain = incoming primer. `decisionInput` (bila ada) harus Behavior dengan satu Parameter return dan tanpa Parameter output lain, dan tidak boleh punya efek samping. | §15.3.3.6; §15.7.12 | Periksa DecisionNode: jenis outgoing mengikuti incoming primer; bila 2 incoming, satu adalah `decisionInputFlow`; decisionInput memenuhi syarat. |
| AC-12 | Edge keluar DecisionNode yang ber-guard dievaluasi per token; urutan evaluasi guard TIDAK didefinisikan dan boleh dievaluasi secara konkuren. Token yang ditawarkan pada incoming primer tidak boleh melewati edge yang guard-nya false. | §15.3.3.6 | Periksa evaluasi guard per token; jangan tuntut urutan evaluasi tertentu. |

---

## 2. Aturan KUALITAS (bukti jurnal)

Sumber: R2 — 10 paper terverifikasi (judul + penulis + tahun + venue +
DOI/URL dicocokkan dengan minimal satu sumber independen).
Batasan bukti R2 yang berlaku untuk SEMUA aturan di §2:
sebagian besar naskah lengkap berbayar / tidak dapat diakses gratis;
temuan kunci diambil dari abstrak resmi dan ringkasan penulis pengutip,
BUKAN dari naskah lengkap, kecuali dinyatakan lain. Bila suatu detail hanya
berasal dari ringkasan sekunder, itu ditandai eksplisit.
Aturan §2 adalah **rekomendasi kualitas berbasis bukti**, bukan pelanggaran
formal OMG — jangan menuliskan pelanggaran §2 seolah-olah pelanggaran §1.

### 2.1 Layout graf — Purchase (1997) [P1]

Purchase, Helen C. (1997). "Which aesthetic has the greatest effect on human
understanding?" *Graph Drawing* (Simposium Graph Drawing), Springer, Lecture
Notes in Computer Science (bab buku prosiding). DOI tidak terverifikasi dari
hasil pencarian. URL halaman penerbit:
https://www.springerprofessional.de/which-aesthetic-has-the-greatest-effect-on-human-understanding/13633544

| ID | Aturan | Sumber | Cara cek |
|---|---|---|---|
| R2-01 | Minimalkan persilangan garis (prioritas tertinggi — "by far the most important aesthetic"). Bila persilangan tidak dapat dihindari, utamakan mengorbankan aspek lain (belokan/simetri) daripada menambah persilangan. | P1 (abstrak resmi) | Hitung persilangan pada tiap diagram; target 0. |
| R2-02 | Minimalkan belokan garis. | P1 (abstrak resmi) | Rute garis memakai segmen horizontal/vertikal dengan belokan sesedikit mungkin (rute L/Z sederhana); hindari garis berliku-liku. |
| R2-03 | Jaga keseimbangan/simetri tata letak (mis. penyebaran node merata, tidak menumpuk di satu sisi). Prioritas di bawah R2-01 dan R2-02. **Aturan lunak.** | P1 (abstrak resmi) | Penilaian visual — bukan metrik keras. |
| R2-04 | **JANGAN** mengklaim "sudah menempel grid ortogonal sehingga diagram bagus" — efek grid ortogonal dan sudut minimum terbukti TIDAK signifikan secara statistik pada P1. Grid boleh dipakai sebagai alat bantu perataan, bukan sebagai bukti kualitas. | P1 (abstrak resmi) | Periksa laporan audit: tidak ada klaim kualitas yang bersandar pada grid ortogonal / sudut minimum. |

Catatan transfer P1 (wajib dicantumkan bila P1 dikutip): eksperimen ini memakai
**graf abstrak**, bukan diagram UML. Menerapkannya ke UML adalah ekstrapolasi
yang masuk akal, tetapi bukan bukti langsung pada UML. Jangan mengutip P1
sebagai "terbukti pada diagram UML" tanpa catatan ini.

### 2.2 Konsistensi notasi — Purchase et al. (2001) [P2] & (2002) [P3]

P2: Purchase, Helen C., Colpoys, Linda, McGill, Matthew, Carrington, David,
& Britton, Carol (2001). "UML class diagram syntax: an empirical study of
comprehension." *Proceedings of the 2001 Asia-Pacific Symposium on Information
Visualisation (APVis 2001)*, hlm. 113–120.
https://dl.acm.org/citation.cfm?id=564054 (ID kutipan ACM 564054; DOI dalam
bentuk 10.5555/… tidak terverifikasi sebagai DOI resmi).

P3: Purchase, Helen C., Colpoys, Linda, McGill, Matthew, & Carrington, David
(2002). "UML collaboration diagram syntax: an empirical study of
comprehension." *Proceedings, First International Workshop on Visualizing
Software for Understanding and Analysis (VISSOFT 2002)*, IEEE, hlm. 13–22.
DOI 10.1109/VISSOF.2002.1019790. URL verifikasi:
http://eprints.gla.ac.uk/35825/ (repositori Universitas Glasgow; refereed = Yes).

| ID | Aturan | Sumber | Cara cek |
|---|---|---|---|
| R2-05 | Pakai SATU gaya notasi yang konsisten di semua diagram; jangan campur varian notasi setara (mis. dua cara menulis asosiasi/agregasi yang berbeda) dalam satu paket diagram. | P2 (abstrak) | Audit konsistensi simbol di semua diagram. |
| R2-06 | Jangan mengandalkan intuisi "notasi ini terasa lebih jelas" sebagai bukti. | P2 (abstrak: prediksi intuitif penulis hanya terkonfirmasi sebagian; notasi terbaik dapat bergantung pada tugas) | Nilai diagram dengan tugas konkret (apakah pembaca dapat menjawab pertanyaan pemahaman dari diagram?), bukan dengan kesan sekilas. |
| R2-07 | Bila dua varian notasi setara secara semantik, pilih yang lebih ringkas. | P3 (abstrak: preferensi subjek berpihak pada varian notasi yang lebih ringkas / concise) | Tidak ada dekorasi/notasi bertele-tele yang tidak menambah makna. |
| R2-08 | Lakukan DUA jenis pengujian diagram secara terpisah: (a) "apakah diagram ini cocok dengan spesifikasi?" dan (b) "temukan kesalahan di diagram ini". Diagram yang lulus (a) belum terbukti mudah dipakai untuk (b). | P3 (abstrak: intuisi terkonfirmasi untuk tugas mengonfirmasi, TIDAK terkonfirmasi untuk tugas mengidentifikasi kesalahan) | Checklist review mencakup kedua jenis tugas. |

### 2.3 Perilaku dinamis — Otero & Dolado (2004) [P4]

Otero, Mari Carmen, & Dolado, José Javier (2004). "Evaluation of the
comprehension of the dynamic modeling in UML." *Information and Software
Technology*, vol. 46, no. 1, hlm. 35–53. DOI 10.1016/S0950-5849(03)00108-3.

| ID | Aturan | Sumber | Cara cek |
|---|---|---|---|
| R2-09 | Untuk perilaku dinamis (alur antar objek seiring waktu), utamakan **sequence diagram** daripada communication/collaboration diagram. | P4 (ringkasan hasil eksperimen: desain lebih mudah dipahami bila perilaku dinamis dimodelkan dengan sequence diagram; dengan collaboration diagram pemahaman menjadi kurang baik seiring domain/dokumen menjadi lebih kompleks) | Setiap alur dinamis utama (mis. alur ujian, koreksi esai, token) dimodelkan sebagai sequence diagram. |
| R2-10 | Sequence diagram yang kompleks harus dipecah per skenario/alur (satu sequence = satu alur utama + alur alternatifnya yang terkait), bukan satu sequence raksasa untuk semua skenario. | P4 (pemahaman bergantung pada jenis diagram dan kompleksitas dokumen; pemahaman menurun seiring kompleksitas) | Tiap sequence diagram punya satu judul alur yang jelas; tidak ada sequence yang mencampur banyak skenario tak terkait. |

### 2.4 Stereotip — Staron, Kuzniarz & Wohlin (2006) [P5] & Cruz-Lemus et al. (2011) [P10]

P5: Staron, Miroslaw, Kuzniarz, Ludwik, & Wohlin, Claes (2006). "Empirical
assessment of using stereotypes to improve comprehension of UML models:
A set of experiments." *Journal of Systems and Software*, vol. 79, no. 5,
hlm. 727–742. DOI 10.1016/j.jss.2005.09.014.

P10: Cruz-Lemus, José A., Genero, Marcela, Caivano, Danilo, Abrahão, Silvia,
Insfrán, Emilio, & Carsí, José A. (2011). "Assessing the influence of
stereotypes on the comprehension of UML sequence diagrams: A family of
experiments." *Information and Software Technology*, vol. 53, no. 12,
hlm. 1391–1403. DOI 10.1016/j.infsof.2011.07.002.

| ID | Aturan | Sumber | Cara cek |
|---|---|---|---|
| R2-11 | Pakai stereotip secara konsisten untuk menambah informasi klasifikasi pada elemen model (mis. `«include»`, `«extend»` pada use case; `«boundary»`, `«control»`, `«entity»` pada lifeline/kelas). | P5 (abstrak/ringkasan: stereotip memainkan peran signifikan dalam pemahaman model; peningkatan dicapai baik mahasiswa maupun profesional industri) | Setiap stereotip yang muncul di diagram terdefinisi maknanya dan dipakai dengan makna yang sama di semua diagram. |
| R2-12 | Jangan memakai stereotip sebagai dekorasi tanpa makna (stereotip yang tidak menambah klasifikasi/informasi dilarang). | P5 | Hapus stereotip yang bila dihapus tidak mengubah pemahaman diagram. |
| R2-27 | Pada sequence diagram, beri stereotip pada pesan/lifeline kunci (mis. `«create»`, `«destroy»`, lifeline `«boundary»` / `«control»` / `«entity»`). Alasan kontekstual: pembaca skripsi (penguji) belum tentu familiar dengan domain CBT, dan justru pada kondisi itulah stereotip terbukti paling membantu. | P10 (abstrak: family of experiments — 1 eksperimen + 2 replikasi, 78 + 29 + 36 mahasiswa S1 Ilmu Komputer — + meta-analisis: stereotip membantu pemahaman, terutama bila subjek tidak familiar dengan domain) | Pesan-pesan kunci berstereotip; lifeline berstereotip peran. |
| R2-28 | Stereotip yang dipakai harus yang maknanya sudah didefinisikan (bukan stereotip karangan bebas tanpa definisi). | P10 | Daftar stereotip + definisinya tercantum di legenda/dokumen pendamping. |

Nuansa P10 (wajib dicantumkan bila P10 dikutip): eksperimen tunggal
pendahulunya (Genero et al., MODELS 2008, hlm. 280–294,
DOI 10.1007/978-3-540-87875-9_20) hanya menunjukkan "kecenderungan lemah
yang mendukung stereotip, namun tidak signifikan secara statistik" dan
penulisnya sendiri menyatakan replikasi lebih lanjut diperlukan. Bukti kuat
berasal dari family of experiments + meta-analisis 2011, bukan dari
eksperimen tunggal 2008.

### 2.5 Prinsip notasi visual — Moody (2009) [P6]

Moody, Daniel L. (2009). "The 'Physics' of Notations: Toward a Scientific
Basis for Constructing Visual Notations in Software Engineering." *IEEE
Transactions on Software Engineering*, vol. 35, no. 6, hlm. 756–779.
DOI 10.1109/TSE.2009.67.

Sifat bukti P6 (wajib dicantumkan bila P6 dikutip): P6 adalah **sintesis
teori** berbasis literatur psikologi kognitif, persepsi visual, semiotika, dan
komunikasi — BUKAN satu eksperimen tunggal atas diagram CBT. Kesembilan
prinsipnya dirumuskan dari teori + bukti empiris yang dirujuk paper tersebut.
Daftar 9 prinsip diverifikasi lewat ringkasan sekunder (arXiv / MDPI / CEUR)
karena naskah lengkap tidak dapat diakses gratis.

| ID | Aturan | Sumber | Cara cek |
|---|---|---|---|
| R2-13 | Satu simbol = satu makna (semiotic clarity). Dilarang: *symbol overload* (satu simbol untuk dua makna), *symbol redundancy* (dua simbol untuk satu makna), *symbol excess* (simbol tanpa makna semantik), *symbol deficit* (konsep penting tanpa simbol). | P6, prinsip 1 | Audit kamus simbol — tiap jenis garis/simbol di diagram dipetakan ke tepat satu makna; contoh: garis putus-putus hanya untuk dependensi/`include`/`extend`, garis solid hanya untuk asosiasi. |
| R2-14 | Simbol berbeda harus dapat dibedakan lewat bentuk + label teks, BUKAN hanya lewat warna (perceptual discriminability). | P6, prinsip 2 | Diagram tetap dapat dibedakan maknanya bila dicetak hitam-putih; `«include»` vs `«extend»` dibedakan oleh teks stereotip + arah panah, bukan hanya warna. |
| R2-15 | Arah panah mengikuti arah pesan/alur yang sebenarnya; aktor manusia digambar sebagai aktor, bukan kotak (semantic transparency). | P6, prinsip 3 | Tidak ada panah yang arahnya berlawanan dengan arah pesan/alur yang dinarasikan. |
| R2-16 | Pecah diagram besar menjadi beberapa diagram terfokus — satu diagram = satu alur / satu sudut pandang (complexity management). | P6, prinsip 4 | Tiap diagram punya judul alur tunggal; tidak ada diagram "semua alur dalam satu gambar". |
| R2-17 | Sediakan mekanisme integrasi antar diagram: penamaan yang IDENTIK antar diagram + referensi silang eksplisit (mis. UC15 ↔ Sequence 3a ↔ Class terkait) (cognitive integration). | P6, prinsip 5 | Nama use case / nama operasi sama persis di diagram Use Case, Sequence, dan Class; tidak ada nama yang hanya muncul di satu diagram tanpa penjelasan. |
| R2-18 | Setiap diagram disertai teks pelengkap: judul + narasi/keterangan singkat. Jangan mengandalkan grafik saja (dual coding). | P6, prinsip 7 | Tiap diagram punya judul dan minimal satu paragraf narasi / keterangan. |
| R2-19 | Jaga jumlah JENIS simbol grafis yang berbeda seminimal mungkin (graphic economy). **Angka panduan "enam" jenis simbol DITANDAI sebagai berasal dari ringkasan sekunder, bukan dari verifikasi langsung naskah Moody — JANGAN mengutip "maksimal 6 simbol" sebagai angka mutlak dari Moody tanpa membaca naskah aslinya.** | P6, prinsip 8 (via ringkasan sekunder) | Inventarisasi jenis simbol per diagram; bila ragam jenis simbol terlalu banyak, sederhanakan atau pecah diagram. |
| R2-20 | Sesuaikan kelengkapan dialek visual dengan audiens: diagram untuk penguji skripsi memakai label lengkap + legenda eksplisit, bukan singkatan internal (cognitive fit). | P6, prinsip 9 | Diagram paket skripsi memuat legenda dan label lengkap; tidak ada singkatan yang tidak dijelaskan. |

### 2.6 Tingkat detail — Nugroho (2009) [P7]

Nugroho, Ariadi (2009). "Level of detail in UML models and its impact on
model comprehension: A controlled experiment." *Information and Software
Technology*, vol. 51, no. 12, hlm. 1670–1685. DOI 10.1016/j.infsof.2009.04.007.

| ID | Aturan | Sumber | Cara cek |
|---|---|---|---|
| R2-21 | Jangan menghilangkan detail kunci demi "kesederhanaan": cantumkan atribut/operasi kunci, label pesan lengkap (nama operasi + parameter penting bila relevan), dan multiplisitas pada relasi. | P7 (eksperimen terkontrol: LoD lebih tinggi meningkatkan kebenaran / correctness dan efisiensi pemahaman) | Tiap pesan sequence berlabel nama operasi yang bermakna; tiap relasi kelas yang penting berlabel multiplisitas. |
| R2-22 | Jangan menilai kualitas diagram dari kesan "terasa jelas/mudah" saja — persepsi kejelasan TIDAK mencerminkan pemahaman sebenarnya pada model ber-detail rendah. | P7 (replikasi: subjek yang menerima model LoD rendah menilai modelnya tampak jelas/mudah, padahal pemahaman sebenarnya terganggu oleh informasi yang terbatas) | Uji diagram dengan pertanyaan pemahaman konkret (dapatkah pembaca menjawab pertanyaan isi diagram dengan benar?), bukan dengan kesan. |
| R2-23 | **Jangan klaim "detail lebih tinggi selalu membuat diagram lebih cepat dibaca"** — efek efisiensi tidak terkonfirmasi kuat pada replikasi. Klaim yang didukung: detail yang cukup meningkatkan KEBENARAN pemahaman. | P7 (replikasi: 23 mahasiswa S2 Ilmu Komputer, Universitas Bina Nusantara, mereplikasi eksperimen awal dengan 53 mahasiswa S2 TU Eindhoven — LoD lebih tinggi konsisten meningkatkan kebenaran; efek pada efisiensi lebih lemah dan tidak signifikan secara statistik pada replikasi) | Periksa klaim dalam laporan: hanya kebenaran pemahaman yang diklaim, bukan kecepatan baca. |

### 2.7 Layout gabungan & ukuran — Störrle (2011) [P8] & (2016) [P9]

P8: Störrle, Harald (2011). "On the impact of layout quality to understanding
UML diagrams." *2011 IEEE Symposium on Visual Languages and Human-Centric
Computing (VL/HCC)*, hlm. 135–142. DOI 10.1109/VLHCC.2011.6070390.

P9: Störrle, Harald (2016). "On the impact of size to the understanding of UML
diagrams." *Software & Systems Modeling*, vol. 17, no. 1, hlm. 115–134.
DOI 10.1007/s10270-016-0529-x. (Versi konferensi MODELS 2014 berjudul
"On the Impact of Layout Quality to Understanding UML Diagrams: Size Matters",
DOI 10.1007/978-3-319-11653-2_32.)

| ID | Aturan | Sumber | Cara cek |
|---|---|---|---|
| R2-24 | Audit layout sebagai PAKET, bukan satu kriteria saja: persilangan + belokan + kesejajaran + jarak antar elemen + keterbacaan label dinilai bersamaan. | P8 (abstrak: layout yang baik memakai BANYAK heuristik secara bersamaan, sedangkan penelitian sebelumnya hanya menguji heuristik satu per satu; compound layouts berdampak nyata terhadap pemahaman) | Checklist audit layout mencakup seluruh aspek sekaligus; dilarang menyatakan "layout sudah rapih" hanya karena satu aspek (mis. hanya persilangan nol) membaik sementara aspek lain memburuk. |
| R2-25 | Batasi ukuran tiap diagram; pecah diagram besar menjadi beberapa diagram kecil yang terfokus. | P9 (abstrak: korelasi negatif yang kuat antara ukuran diagram dan kinerja pemodel; dampak ukuran setara dengan atau melebihi faktor-faktor lain) | Diagram yang "terasa besar" (banyak elemen/relasi) wajib dipecah; tidak ada diagram tunggal yang memuat seluruh sistem. |
| R2-26 | **LARANGAN MENGARANG ANGKA — DILARANG** mengarang angka spesifik "ukuran optimal" (mis. "maksimal 20 elemen menurut Störrle") atas nama paper ini. Angka spesifik "ukuran optimal" TIDAK dikutip di R2 karena tidak berhasil diverifikasi dari abstrak yang dapat diakses. Bila angka diperlukan untuk skripsi, baca naskah lengkap P9 dan kutip angka beserta definisi metrik ukurannya. | P9 (batasan bukti R2) | Periksa setiap angka ukuran yang diklaim atas nama Störrle: harus ada kutipan naskah lengkap + definisi metrik; bila tidak ada, hapus klaimnya. |

> ⚠️ **LARANGAN EKSPLISIT (Störrle):** JANGAN mengarang angka "ukuran optimal"
> atas nama Störrle. R2-26 adalah aturan yang melarang klaim angka — bukan
> aturan yang memberi angka. Setiap angka "maksimal N elemen" yang dikaitkan
> dengan Störrle tanpa kutipan naskah lengkap P9 adalah klaim yang dikarang
> dan harus dihapus (lihat juga NA-11, NA-12 di §4).

---

## 3. Aturan KONVENSI SKRIPSI (otoritatif)

Sumber: R3. Status **OTORITATIF** = pedoman resmi perguruan tinggi
(SK Rektor / buku pedoman resmi kampus). **Berlaku hanya untuk kampus
penerbitnya** — TIDAK ada satu pedoman nasional yang mengatur notasi UML
untuk semua skripsi di Indonesia.

⚠️ **Batas penting yang ditemukan R3:** tidak ada satu pun pedoman kampus
yang ditemukan menetapkan notasi UML baku versinya sendiri sampai level
urutan lifeline, penomoran pesan sequence, atau penulisan guard pada
sequence. Yang mengatur notasi UML adalah standar OMG (§1), bukan pedoman
kampus. Pedoman kampus mengatur: konsistensi teori↔pembahasan, set diagram,
keterkaitan dengan proses bisnis, dan format naskah (caption, penempatan
gambar). Jangan mengklaim satu konvensi UML sebagai "wajib skripsi" kecuali
konvensi itu ada di pedoman kampus target.

Catatan akses R3 yang dipertahankan: beberapa pedoman hanya terjangkau lewat
salinan cermin (123dok / Scribd / SlideShare / Academia.edu) — hal itu
dicatat eksplisit per sumber, dan salinan cermin TIDAK diperlakukan setara
dengan dokumen resmi di domain kampus. Hanya A1 (UBSI) yang diakses lewat
domain resmi kampus.

### 3.1 Aturan otoritatif (didukung pedoman resmi kampus)

| ID | Aturan | Sumber otoritatif (institusi + dokumen) | Cara cek |
|---|---|---|---|
| SK-01 | **Konsistensi teori ↔ pembahasan UML adalah WAJIB.** Teori diagram yang ditulis di Bab II harus sama dengan diagram yang dibahas di Bab III/IV. | Universitas Bina Sarana Informatika (UBSI), Fakultas Teknik dan Informatika — *Pedoman Penyusunan Laporan Skripsi, Tugas Akhir dan Pelaksanaan Ujian Komprehensif, Program Sarjana (S1) dan Program Diploma Tiga (D3)*, Keputusan Rektor UBSI Nomor 077/2.02/UBSI/III/2021. URL (domain resmi kampus): https://elibrary.bsi.ac.id/homepage/download/dokumen/20230216103710-pedoman-penyusunan-laporan-skripsi-tugas-akhir-dan-pelaksanaan-ujian-komprehensip-fakultas-teknik-dan-informatika.pdf — Bab II: teori UML "harus memiliki konsistensi terhadap pembahasan". | Bandingkan daftar diagram + teori di Bab II dengan diagram yang dibahas di Bab III/IV: harus sama, tidak ada diagram yang dibahas tanpa teori dan sebaliknya. |
| SK-02 | **Set diagram bersifat kondisional:** 5 diagram (Activity, Use Case, Sequence, Class, Deployment) bila program memakai class; 4 diagram (tanpa Class) bila tidak memakai class. | UBSI (dokumen yang sama, Bab II) + STMIK Nuri — *Buku Pedoman Skripsi* (Program Studi Teknik Informatika), salinan 123dok: https://123dok.com/document/zlren36z-buku-pedoman-skripsi-stmik-nuri-dosen.html | Periksa set diagram naskah terhadap pemakaian class di program: 5 bila pakai class, 4 bila tidak. |
| SK-03 | **Set diagram tergantung paradigma (OOP vs terstruktur), bukan satu set universal.** OOP: class diagram, sequence diagram, component diagram, deployment diagram (terkait program yang dibuat dan hanya yang berhubungan dengan proses bisnis sistem). Terstruktur: deployment diagram dan component diagram. (Versi lain kutipan yang sama menambahkan activity diagram ke kedua kelompok.) | STMIK Nusa Mandiri Jakarta — *Pedoman Laporan Skripsi Periode I 2021*, salinan SlideShare/Scribd (bukan domain resmi kampus): https://www.slideshare.net/slideshow/pedoman-lap-skripsi-stmik-nusa-mandiri-jakarta-periode-i-2021pdf/258892158 ; STMIK Pelita Nusantara Medan — *Pedoman Penulisan Usulan Proposal dan Skripsi S1*, salinan 123dok: https://123dok.com/document/qv7leldq-pedoman-penulisan-usulan-proposal-skripsi-jenjang-strata-nusantara.html (Bab perancangan: Use Case + Class + Activity/State + Sequence bila memakai UML, atau Flowchart) | Periksa set diagram naskah terhadap paradigma program; jangan tuntut satu set universal. |
| SK-04 | **Setiap diagram UML harus terkait dengan proses bisnis sistem** (bukan diagram hiasan). | STMIK Nusa Mandiri Jakarta (dokumen yang sama dengan SK-03) | Periksa tiap diagram: ada proses bisnis sistem yang dimodelkannya; diagram tanpa kaitan proses bisnis ditolak. |
| SK-05 | **Caption diagram: "Gambar `<bab>.<urut>`", judul DI BAWAH gambar.** Angka pertama = bab tempat gambar berada, angka kedua = urutan gambar dalam bab tersebut. | Universitas Ciputra — *Pedoman Penulisan Skripsi / Tugas Akhir Mahasiswa Program Sarjana*, salinan Academia.edu (bukan domain resmi kampus): https://www.academia.edu/8263617/PEDOMAN_PENULISAN_SKRIPSI_TUGAS_AKHIR_MAHASISWA_PROGRAM_SARJANA_UNIVERSITAS_CIPUTRA — ketentuan (a), (d), (f). Pola yang sama lazim di pedoman lain (R3 A7, institusi tidak teridentifikasi — hanya bukti pola lazim, bukan aturan kampus tertentu). | Periksa tiap caption diagram: format "Gambar `<bab>.<urut>`", judul di bawah gambar. |
| SK-06 | **Gambar + judulnya harus dalam satu halaman, tidak terpotong.** Gambar dan tabel harus terletak dalam satu halaman, tidak terpotong. | Universitas Ciputra (dokumen yang sama dengan SK-05), ketentuan (c) | Periksa layout naskah: gambar dan judulnya tidak terbelah antar halaman. (Bukan berarti diagram harus muat satu halaman — lihat NA-07.) |
| SK-07 | **Gambar dan tabel ditempatkan di antara bagian teks yang paling banyak membahasnya.** | Universitas Ciputra (dokumen yang sama dengan SK-05), ketentuan (b) | Periksa penempatan tiap gambar: berada di dekat teks yang paling banyak membahasnya. |
| SK-08 | **Gambar/tabel yang terlalu besar boleh landscape / kertas A3 dilipat / diperkecil ke A4.** (Pengecualian eksplisit terhadap keharusan satu halaman.) | Universitas Ciputra (dokumen yang sama dengan SK-05), ketentuan (g) | Bila diagram besar, periksa apakah dipakai salah satu dari tiga cara ini — bukan dijadikan pelanggaran "tidak muat satu halaman". |
| SK-09 | **Activity diagram diletakkan setelah narasi proses bisnis** (aturan urutan penyajian, bukan aturan notasi). Juga: analisa kebutuhan sistem disertai penggambaran use case diagram dan activity diagram. | STMIK Nusa Mandiri Jakarta (dokumen yang sama dengan SK-03) | Periksa urutan naskah: narasi proses bisnis mendahului activity diagram-nya. |
| SK-10 | **Bila menggunakan ERD, harus menggambarkan ERD, LRS, dan spesifikasi file-nya.** Desain database menggunakan ERD. | STMIK Nusa Mandiri Jakarta (dokumen yang sama dengan SK-03: "Desain Database: jika menggunakan ERD maka harus menggambarkan ERD, LRS, dan juga spesifikasi filenya"); STMIK Nuri (dokumen yang sama dengan SK-02: "Desain Database menggunakan Entity Relationship Diagram (ERD)") | Bila naskah memakai ERD, periksa kelengkapan ERD + LRS + spesifikasi file. (ERD sendiri bukan diagram UML — lihat NA-10.) |
| SK-11 | **Tidak ada notasi UML baku versi kampus — mahasiswa wajib konsisten dengan referensi yang dipakainya.** "Simbol UML — sesuaikan dengan symbol UML yang dipakai dalam penulisan skripsi, sesuaikan dengan referensi yang digunakan." Konsekuensi: keputusan yang dapat dipertahankan di sidang adalah konsisten dengan SATU referensi yang dikutip. | UBSI (dokumen yang sama dengan SK-01), Lampiran 33 (Daftar Simbol) | Periksa naskah: satu referensi notasi dikutip, dan semua simbol diagram konsisten dengan referensi itu. Jangan tuntut "notasi baku kampus X" — tidak ada. |
| SK-12 | **Detail format naskah mengikuti pedoman kampus target masing-masing** (bukan satu format universal). Contoh: Universitas Ciputra menetapkan judul gambar di bawah gambar ditulis Times New Roman 10 point, spasi 1, rata kiri, dan menegaskan detail format "diserahkan kepada masing-masing Program Studi". Universitas Indonesia mengatur kertas HVS 80 gram A4, margin kiri 4 cm / kanan 3 cm / atas 3 cm / bawah 3 cm, huruf Times New Roman 12, spasi 1,5, penomoran halaman, sampul — dan TIDAK ditemukan ketentuan tentang notasi UML dalam cuplikan yang terjangkau. | Universitas Ciputra (dokumen yang sama dengan SK-05); Universitas Indonesia — *Pedoman Teknis Penulisan Tugas Akhir Mahasiswa Universitas Indonesia*, salinan 123dok / DocPlayer / Academia.edu (bukan domain resmi UI): https://123dok.com/document/zk693epy-universitas-indonesia-pedoman-teknis-penulisan-mahasiswa-universitas-indonesia.html | Periksa format naskah terhadap pedoman kampus target yang berlaku; jangan terapkan detail tipografi satu kampus (mis. TNR 10 Ciputra) ke kampus lain. Pedoman UI adalah bukti negatif: pedoman kampus besar pun tidak selalu mengatur notasi UML. |

### 3.2 BUKAN aturan otoritatif — literatur / bahan ajar / contoh (boleh dipakai sebagai praktik baik, TIDAK BOLEH diklaim sebagai "aturan skripsi")

Daftar ini dipisahkan secara tegas dari §3.1. Isinya TIDAK BOLEH dijadikan
temuan "melanggar aturan skripsi". Bila dipakai, kutip sebagai literatur /
bahan ajar / contoh, bukan sebagai pedoman kampus.

| ID | Praktik (bukan aturan) | Sumber (status) | Catatan pakai |
|---|---|---|---|
| SK-L01 | Aktor biasanya paling kiri pada sequence diagram. Kata kuncinya "biasanya" — bukan kewajiban mutlak. | Sholiq, *Pemodelan Sistem Berorientasi Objek dengan UML* (2006; edisi/kutipan 2010) — LITERATUR (buku teks), dikutip lewat kutipan di berbagai skripsi, mis. https://123dok.com/article/sequence-diagram-perancangan-proses-analisis-dan-perancangan-sistem.zpnplnvy | Boleh dipakai sebagai kebiasaan penyajian; jangan jadikan temuan bila aktor tidak paling kiri. |
| SK-L02 | Urutan waktu sequence: atas → bawah; objek berjajar kiri → kanan. | Dharwiyanti & Wahono (2003), *Pengenalan Unified Modeling Language (UML)* — LITERATUR / TUTORIAL, salinan 123dok: https://123dok.com/document/y93jj5vy-sri-dharwiyanti-romi-satria-wahono.html | Konsisten dengan OMG (SEQ-25); kutip sebagai literatur, bukan pedoman kampus. |
| SK-L03 | Klasifikasi objek sequence menjadi Boundary / Control / Entity (asal Jacobson; dipakai untuk mendukung arsitektur Model–View–Controller / MVC). Boundary = antarmuka yang menghubungkan aktor dengan sistem; entity = informasi permanen; control = pengelola alur logika skenario use case. | Skripsi UNPAS (CONTOH SATU SKRIPSI, repositori resmi UNPAS): https://repository.unpas.ac.id/83672/1/Mhs_213040108_Teknik%20Informatika%20%281%29.pdf (satu percobaan fetch langsung gagal timeout di sisi tool; dikutip dari cuplikan hasil pencarian) + Dharwiyanti & Wahono (2003) | Praktik lazim, bukan aturan pedoman. Urutan kiri→kanan BCE TIDAK diatur (lihat NA-04). |
| SK-L04 | Class = 3 kompartemen (nama+stereotip, atribut, metoda); visibilitas Private / Protected / Public. | Dharwiyanti & Wahono (2003) — LITERATUR | Konsisten dengan OMG (CL-09, CL-10); bukan aturan pedoman kampus. |
| SK-L05 | Simbol visibilitas `+` Public, `-` Private, `#` Protected, `~` Package. | Dharwiyanti & Wahono (2003) untuk `+`/`-`/`#`; `~` dari panduan vendor — LITERATUR / PANDUAN TOOL | Bukan aturan pedoman kampus. |
| SK-L06 | Multiplicity `1` (tepat satu), `0..*` (nol atau lebih), `1..*` (satu atau lebih), `0..1` (nol atau satu). | Skripsi Universitas Bangka Selatan (CONTOH SATU SKRIPSI, salinan 123dok) — praktik lazim UML | Bukan aturan pedoman; makna multiplisitas yang mengikat adalah OMG (CL-02, CL-03). |
| SK-L07 | Nama use case diawali kata kerja; nama aktor diawali kata benda. Kata kuncinya "biasanya". | Modul Politeknik Telkom, *Analisis dan Desain Sistem Informasi*, bab "Pengenalan UML dan Analisis Use Case" — BAHAN AJAR kampus (modul kuliah), bukan pedoman penulisan skripsi. Salinan 123dok: https://text-id.123dok.com/document/wq2opmrpz-pengenalan-uml-dan-analisi-use-case-analisis-dan-desain-sistem-informasi-1.html | Anjuran gaya penamaan; bukan aturan pedoman. |
| SK-L08 | Satu sequence diagram per alur (alur normal dan error dipisah); urutan pengerjaan: business process → use case → activity → sequence → class. | Henderi (2008:6) — LITERATUR / BAHAN AJAR, dikutip di slide/materi kuliah, mis. https://id.scribd.com/document/827294473/8-UML | Anjuran metodologis, bukan aturan pedoman. Sejalan dengan R2-10 dan R2-16. |
| SK-L09 | Guard `[kondisi]` untuk transisi statechart; format transisi `event [guard condition] / action`. | Dharwiyanti & Wahono (2003) + materi dosen UPI YAI (BAHAN AJAR, domain kampus: https://dosen.upi-yai.ac.id/v5/dokumen/materi/060060/52_20241210032909_UML.pdf) | Terbukti untuk transisi **statechart**. Untuk sequence, guard diatur OMG (CF-11, AC-02), bukan pedoman kampus (lihat NA-24). |
| SK-L10 | Message pada sequence dipetakan menjadi operasi/metode class (traceability desain → kode). | Dharwiyanti & Wahono (2003) — LITERATUR | Prinsip traceability; korespondensi formalnya diatur OMG (SEQ-09, SEQ-10). |
| SK-L11 | Set 4 diagram (Use Case, Activity, Sequence, Class) sebagai pola yang sangat lazim di jurnal Indonesia. | Contoh jurnal: Yusup & Anwar (2024), *IKRA-ITH Informatika* 8(3):174–184, http://journals.upi-yai.ac.id/index.php/ikraith-informatika/article/view/4508 ; Setiawan, Agustin & Ningsih (2022), *Jurnal Algoritma* 19(1):42–53, https://jurnal.itg.ac.id/index.php/algoritma/article/view/996 ; dan contoh lain di R3 C7 | Kelaziman ≠ aturan pedoman. Set diagram mengikuti pedoman kampus target (§3.1). |

---

## 4. BUKAN ATURAN — JANGAN DITEGAKKAN

Bagian ini sama pentingnya dengan §1–§3. Setiap butir di bawah
**DILARANG** dijadikan temuan / pelanggaran. Bila sebuah temuan audit
bersandar pada salah satu butir ini, temuan itu harus ditarik.
Alasan + sumber pembatalnya dicantumkan di tiap baris.

| ID | Klaim yang DILARANG ditegakkan | Alasan + sumber yang membatalkannya |
|---|---|---|
| NA-01 | Penomoran pesan Sequence Diagram `1`, `1.1`, `2` (nested) sebagai persyaratan OMG / aturan skripsi. | Penomoran berurutan hanya didefinisikan untuk **Communication Diagram** (R1 SEQ-17: §17.4.4.1, §17.9.1.3, §17.9). Sintaks label Message Sequence di §17.4.4.1 tidak memuat nomor urut. Di literatur, aturan "level tertinggi = 1, level sama = prefiks sama" dalam Dharwiyanti & Wahono (2003) berlaku untuk **Collaboration/Communication diagram**, bukan sequence (R3 D11). Menempelkannya ke sequence adalah **salah kutip**. |
| NA-02 | "Setiap `synchCall` WAJIB punya reply" sebagai pelanggaran formal. | Tidak ada constraint di §17.12.18.7 yang mewajibkannya; §17.2.3.4 hanya berkata "Typically" (R1 SEQ-05, SEQ-21). SynchCall tanpa reply = model tidak lengkap / perlu ditanya, BUKAN pelanggaran formal. |
| NA-03 | "Setiap cabang Decision WAJIB punya guard" sebagai pelanggaran formal. | Guard per edge bersifat opsional; edge tanpa guard = selalu true (R1 AC-03, §15.3.3.6; AC-01, §15.2.3.3). Yang mengikat: satu token paling banyak satu cabang + anjuran determinisme ("should", bukan constraint). Merge tanpa guard adalah benar (AC-04). |
| NA-04 | Urutan lifeline WAJIB Actor → Boundary → Control → Entity. | **Tidak ditemukan sumber otoritatif.** Klasifikasi BCE terbukti dipakai (R3 C1, C4, B1), tetapi tidak ada sumber yang menetapkan urutan kiri→kanan itu sebagai kewajiban (R3 D9). OMG juga tidak menetapkan urutan lifeline tersebut (R1 §17.3.3.1 hanya mengatur urutan event atas→bawah). |
| NA-05 | Batas angka jumlah lifeline (mis. "maksimal N lifeline"). | Tidak ada paper terverifikasi di R2 yang menyatakan batas angka jumlah lifeline. Yang didukung hanya prinsip umum: kurangi kompleksitas / pecah diagram (P4, P9), tanpa angka (R2, "Klaim yang SENGAJA TIDAK dijadikan aturan" butir 1). |
| NA-06 | "Label pesan HARUS memakai nama operasi" sebagai format penamaan yang diwajibkan jurnal. | Tidak ada paper di R2 yang menyatakan kalimat persis ini. R2-21 (P7) hanya menuntut label pesan yang lengkap dan bermakna, bukan format penamaan tertentu (R2, butir 2). (Catatan: OMG SEQ-09 mewajibkan nama Message = nama NamedElement yang dirujuk signature — itu aturan OMG §1, bukan aturan jurnal, dan jangan diklaim sebagai "aturan jurnal".) |
| NA-07 | "Diagram harus muat satu halaman." | Tidak ada paper di R2 yang menyatakan kalimat persis ini. Yang didukung: diagram lebih kecil lebih mudah dipahami (P9) dan diagram besar harus dipecah (P6, P9) — tanpa patokan "satu halaman" (R2, butir 3). Pedoman Ciputra justru membolehkan gambar terlalu besar memakai landscape / A3 dilipat / diperkecil ke A4 (SK-08); yang harus satu halaman adalah gambar + judulnya agar tidak terpotong (SK-06). |
| NA-08 | Larangan mutlak garis diagonal / garis lengkung. | P1 (Purchase 1997) hanya menyatakan efek grid ortogonal tidak signifikan dan bends berpengaruh lebih kecil; larangan mutlak garis diagonal berasal dari buku panduan praktisi (Ambler, *The Elements of UML 2.0 Style*, Cambridge University Press, 2005) — **buku panduan praktisi, bukan paper jurnal peer-reviewed** — dan tidak boleh dikutip sebagai bukti empiris jurnal (R2, butir 4 + "Kandidat yang ditelusuri"). Yang didukung P1: minimalkan belokan (R2-02). |
| NA-09 | Guard `[[…]]` kurung ganda sebagai notasi yang benar / yang wajib. | **Tidak ditemukan sumber apa pun** (R3 D13). Tidak ada sumber yang menulis guard dengan kurung siku ganda; ini artefak rendering tool, bukan notasi. Guard yang benar: satu pasang kurung siku (R1 CF-11, AC-02). |
| NA-10 | Aturan notasi ERD (crow's foot, kunci primer / kunci tamu sebagai kolom) dari OMG. | **ERD bukan diagram UML 2.5.1.** Pencarian teks penuh spesifikasi tidak menemukan "Entity-Relationship" / ERD; taksonomi diagram spesifikasi tidak memuatnya (R1 CL-13 — TIDAK TERVERIFIKASI, di luar cakupan spesifikasi). Diagram ERD hanya dapat diaudit untuk (a) konsistensi internal terhadap Class Diagram (entitas ↔ class, relasi ↔ asosiasi, kardinalitas ↔ multiplisitas) dan (b) konvensi proyek sendiri — BUKAN terhadap OMG. |
| NA-11 | Klaim angka spesifik (mis. "ukuran optimal", "maksimal 20 elemen") atas nama paper yang tidak terverifikasi dari naskah lengkap — khususnya atas nama Störrle (P9). | Angka "ukuran optimal" Störrle TIDAK dikutip di R2 karena tidak berhasil diverifikasi dari abstrak yang dapat diakses (R2-26). Mengarang angka atas nama paper ini DILARANG. Bila angka diperlukan, baca naskah lengkap P9 dan kutip angka beserta definisi metrik ukurannya. |
| NA-12 | "Maksimal N elemen per diagram" (angka apa pun, atas nama siapa pun). | Tidak ada angka yang dikutip di R2 (R2, butir 5). Yang didukung: batasi ukuran dan pecah diagram besar (R2-25), tanpa angka. |
| NA-13 | "Maksimal 6 jenis simbol" sebagai angka mutlak dari Moody (P6). | Angka panduan "enam" pada R2-19 berasal dari **ringkasan sekunder** atas Moody, bukan dari verifikasi langsung naskah Moody. Jangan mengutipnya sebagai angka mutlak tanpa membaca naskah asli Moody (R2 P6, R2-19). Yang didukung: jaga ragam jenis simbol seminimal mungkin. |
| NA-14 | "Sudah menempel grid ortogonal sehingga diagram bagus" sebagai bukti kualitas. | Efek grid ortogonal dan sudut minimum terbukti TIDAK signifikan secara statistik pada P1 (R2-04). Grid hanya alat bantu perataan. |
| NA-15 | "Detail lebih tinggi selalu membuat diagram lebih cepat dibaca." | Efek efisiensi tidak terkonfirmasi kuat pada replikasi Nugroho (R2-23). Yang didukung: detail cukup meningkatkan KEBENARAN pemahaman. |
| NA-16 | "Kepala lifeline hasil create mulai tepat di titik pesan create" sebagai temuan formal. | TIDAK TERVERIFIKASI di R1 (SEQ-27): tidak ada kalimat eksplisit di teks spesifikasi yang diakses. Yang terverifikasi hanya arti `createMessage` (SEQ-07) dan gaya garisnya (SEQ-14). Jangan jadikan temuan audit formal. |
| NA-17 | "Ujung lifeline yang dihancurkan ditandai tanda X" sebagai temuan formal. | TIDAK TERVERIFIKASI di R1 (SEQ-28): tidak ditemukan di teks spesifikasi yang diakses (kemungkinan hanya pada gambar contoh). Yang terverifikasi: `deleteMessage` berakhir di DestructionOccurrenceSpecification (SEQ-08) dan tidak ada event sesudahnya (SEQ-23). |
| NA-18 | Nama kelas PascalCase / kata benda tunggal sebagai "aturan skripsi". | TIDAK DITEMUKAN SUMBER OTORITATIF (Indonesia) di R3 (D15). Sumber Inggris (vendor tool) menyebut nama kelas biasanya huruf kapital + kata benda tunggal, tetapi itu bukan pedoman skripsi Indonesia. |
| NA-19 | Satu set diagram universal untuk semua skripsi (mis. "wajib 7 diagram"). | TIDAK DITEMUKAN di R3 (D22). Semua pedoman yang ditemukan menetapkan set secara kondisional dan berbeda-beda antar kampus (SK-02, SK-03). |
| NA-20 | "Pedoman kampus selalu mengatur notasi UML." | SALAH (R3 D23). Contoh: pedoman UI (A6) hanya mengatur format naskah, tidak mengatur notasi UML. |
| NA-21 | Caption memakai angka Romawi bab ("Gambar III.1") sebagai aturan. | VARIASI CONTOH, BUKAN ATURAN (R3 D6): ditemukan di satu skripsi (C2, UIN Ar-Raniry); kampus lain memakai "Gambar 3.1". Ikuti pedoman kampus target (SK-05). |
| NA-22 | Detail tipografi caption (mis. Times New Roman 10, spasi 1, rata kiri) sebagai aturan universal. | Detail tersebut adalah aturan Universitas Ciputra saja, bukan aturan universal — kampus lain menetapkan detail berbeda (R3 A2; SK-12). |
| NA-23 | Menilai kualitas diagram dari kesan "terasa jelas/mudah" saja. | Persepsi kejelasan TIDAK mencerminkan pemahaman sebenarnya pada model ber-detail rendah (R2-22, P7). Uji dengan pertanyaan pemahaman konkret. |
| NA-24 | Guard `[…]` pada sequence sebagai "aturan pedoman kampus Indonesia". | Hanya didukung standar OMG UML 2.5 secara umum (CF-11, AC-02) dan literatur untuk transisi statechart (SK-L09); TIDAK ditemukan pedoman kampus Indonesia yang mengaturnya (R3 D12). Jangan klaim sebagai aturan pedoman kampus. |
| NA-25 | "Notasi ini terasa lebih jelas" (intuisi) sebagai bukti kualitas notasi. | Prediksi intuitif penulis hanya terkonfirmasi sebagian pada P2 (R2-06). Nilai dengan tugas konkret, bukan kesan sekilas. |

---

## 5. Tiga Koreksi Premis R1 (membatalkan asumsi lama)

Ketiga koreksi ini ditampilkan menonjol karena membatalkan asumsi yang
selama ini dipakai sebagai dasar temuan. Setiap temuan audit lama yang
bersandar pada premis-premis ini **BATAL**.

### Koreksi 1 — Penomoran pesan `1`, `1.1`, `2` BUKAN bagian Sequence Diagram

- **Premis lama (salah):** "Pesan pada Sequence Diagram harus / boleh
  dinilai dari penomoran `1`, `1.1`, `2`."
- **Yang benar:** Sintaks label Message Sequence Diagram di OMG UML 2.5.1
  §17.4.4.1 tidak memuat nomor urut apa pun. Penomoran berurutan
  (`sequence-expression`, mis. `1b.1`, `3.1.4`) hanya didefinisikan untuk
  **Communication Diagram** (§17.9.1.3: "The sequencing of Messages is given
  through a sequence numbering scheme"; §17.4.4.1: "On Communication
  Diagrams, the Messages are decorated by … sequence number").
- **Konsekuensi audit:** angka `1`, `1.1` pada Sequence Diagram adalah
  **anotasi tool / konvensi proyek**, bukan persyaratan OMG — tidak boleh
  dinilai benar/salah terhadap spesifikasi; yang mengikat adalah urutan
  vertikal (atas → bawah) dan korespondensi pesan (R1 SEQ-17).
- **Konfirmasi dari R3:** aturan penomoran "level tertinggi = 1, level sama
  = prefiks sama" dalam Dharwiyanti & Wahono (2003) berlaku untuk
  Collaboration/Communication diagram, bukan sequence. Menempelkannya ke
  sequence sebagai "aturan skripsi" adalah salah kutip (R3 D11; NA-01).

### Koreksi 2 — Reply TIDAK WAJIB untuk setiap `synchCall`

- **Premis lama (salah):** "Setiap `synchCall` wajib punya reply; tanpa
  reply = pelanggaran formal."
- **Yang benar:** Tidak ada satu pun constraint di §17.12.18.7 yang
  mewajibkan pasangan reply untuk tiap synchCall (daftar constraint:
  `sending_receiving_message_event`, `arguments`, `cannot_cross_boundaries`,
  `signature_is_signal`, `occurrence_specifications`, `signature_refer_to`,
  `signature_is_operation_request`, `signature_is_operation_reply`).
  Spesifikasi hanya berkata "Typically" start/finish ExecutionSpecification
  = receive + send reply (§17.2.3.4) — "typically", bukan "must".
- **Konsekuensi audit:** synchCall tanpa reply adalah **model tidak lengkap
  / perlu ditanya**, BUKAN pelanggaran formal (R1 SEQ-05, SEQ-21; NA-02).

### Koreksi 3 — Guard TIDAK WAJIB di tiap cabang Decision

- **Premis lama (salah):** "Setiap cabang Decision wajib punya guard;
  cabang tanpa guard = pelanggaran formal."
- **Yang benar:** Guard bersifat opsional per edge. Spesifikasi: "If any of
  the outgoing edges of a DecisionNode have guards, then these are
  evaluated…" (§15.3.3.6). Edge tanpa guard diperlakukan sebagai guard true
  (§15.2.3.3). Merge tanpa guard adalah benar (AC-04).
- **Yang mengikat (bukan guard wajib):** (a) DecisionNode punya 1–2
  incoming edge dan ≥ 1 outgoing edge; (b) setiap token paling banyak
  melewati SATU outgoing edge (token tidak diduplikasi); (c) token tidak
  boleh melewati edge yang guard-nya false; (d) anjuran determinisme —
  modeler **sebaiknya** ("should") mengatur agar paling banyak satu guard
  true per token; ini anjuran, bukan constraint. Guard `else` hanya untuk
  DecisionNode, paling banyak pada SATU outgoing edge.
- **Konsekuensi audit:** cabang tanpa guard BUKAN pelanggaran (R1 AC-03;
  NA-03).

---

## 6. Hierarki Kekuatan Bukti

Urutan berikut mengikat untuk SEMUA tim audit / standardisasi. Bila dua
sumber bertentangan, sumber yang lebih tinggi menang. **Aturan dari tingkat
bawah TIDAK BOLEH mengalahkan aturan tingkat atas bila bertentangan.**

| Tingkat | Sumber | Kekuatan | Contoh pemakaian |
|---|---|---|---|
| 1 (tertinggi) | **Spek formal OMG** — OMG UML 2.5.1, formal/2017-12-05 | Mengikat untuk notasi & semantik UML. Pelanggaran = pelanggaran formal. | §1 (SEQ, CF, UC, CL, AC). |
| 2 | **Jurnal peer-reviewed** — 10 paper terverifikasi R2 | Bukti kualitas / keterbacaan. Rekomendasi, bukan pelanggaran formal OMG. Tidak boleh menciptakan "aturan OMG" baru. | §2 (R2-01 s.d. R2-28). |
| 3 | **Pedoman kampus resmi** — SK Rektor / buku pedoman resmi kampus | Mengikat untuk naskah skripsi di kampus penerbitnya saja (format, set diagram, konsistensi). Tidak mengatur notasi UML. | §3.1 (SK-01 s.d. SK-12). |
| 4 (terendah) | **Literatur / bahan ajar / contoh** — buku teks, tutorial, modul kuliah, satu skripsi / satu jurnal | Praktik baik / kebiasaan. Boleh dipakai bila dikutip jujur sebagai literatur/contoh. TIDAK BOLEH diklaim sebagai "aturan skripsi" atau "aturan OMG". | §3.2 (SK-L01 s.d. SK-L11). Di bawah ini: buku panduan praktisi (mis. Ambler 2005 — bukan bukti empiris jurnal) dan salinan cermin pedoman (bukan dokumen resmi). |

### 6.1 Konflik antar-sumber yang ditemukan + cara penyelesaian

| # | Konflik / ketegangan | Penyelesaian (sesuai hierarki) |
|---|---|---|
| K-01 | Premis umum "penomoran pesan sequence wajib" vs spesifikasi OMG (R1) + literatur (R3). | **OMG menang (tingkat 1).** Penomoran hanya untuk Communication Diagram. Premis batal (§5 Koreksi 1; NA-01). |
| K-02 | Premis umum "reply wajib per synchCall" vs teks spesifikasi ("Typically"). | **OMG menang (tingkat 1).** Tanpa reply = tidak lengkap, bukan pelanggaran (§5 Koreksi 2; NA-02). |
| K-03 | Premis umum "guard wajib per cabang Decision" vs spesifikasi (guard opsional). | **OMG menang (tingkat 1).** Yang mengikat: satu-token-satu-cabang + anjuran determinisme (§5 Koreksi 3; NA-03). |
| K-04 | Buku panduan praktisi Ambler (2005): hindari garis diagonal/lengkung, vs eksperimen P1: efek grid ortogonal tidak signifikan, bends berpengaruh lebih kecil. | **Jurnal peer-reviewed (tingkat 2) mengalahkan buku panduan praktisi (di bawah tingkat 4 sebagai bukti empiris).** Tidak ada larangan mutlak garis diagonal (NA-08); yang berlaku: minimalkan belokan (R2-02). Ambler boleh dipakai sebagai pendukung gaya dengan label "buku panduan, bukan jurnal". |
| K-05 | Nugroho (P7) eksperimen awal: LoD tinggi meningkatkan kebenaran DAN efisiensi, vs replikasi: efek efisiensi lebih lemah dan tidak signifikan. | **Bukti yang lebih baru / replikasi membatasi klaim.** Klaim yang dipertahankan: detail cukup meningkatkan KEBENARAN pemahaman; klaim kecepatan baca ditarik (R2-23; NA-15). |
| K-06 | Genero et al. (2008, eksperimen tunggal): stereotip pada sequence hanya "kecenderungan lemah, tidak signifikan", vs Cruz-Lemus et al. (2011, family of experiments + meta-analisis): stereotip membantu pemahaman. | **Family of experiments + meta-analisis (2011) mengalahkan eksperimen tunggal 2008.** Yang dipakai: R2-27, R2-28; eksperimen 2008 dicantumkan sebagai nuansa, bukan dasar aturan. |
| K-07 | Angka "enam" jenis simbol (Moody, P6) dari ringkasan sekunder vs naskah asli Moody yang tidak terverifikasi langsung. | **Hanya yang terverifikasi yang dikutip.** Angka "6" tidak dikutip sebagai angka mutlak Moody (R2-19; NA-13). Bila angka diperlukan, baca naskah lengkap Moody. |
| K-08 | Angka "ukuran optimal" Störrle (P9) yang beredar vs abstrak P9 yang dapat diakses (tidak memuat angka). | **Klaim angka tanpa verifikasi naskah lengkap ditolak.** Larangan mengarang angka (R2-26; NA-11, NA-12). |
| K-09 | Pedoman Ciputra: "gambar + judul satu halaman, tidak terpotong" vs klaim "diagram harus muat satu halaman". | **Teks pedoman dibaca utuh.** Yang satu halaman adalah gambar + judulnya (SK-06); diagram yang terlalu besar boleh landscape / A3 dilipat / diperkecil ke A4 (SK-08). Klaim "diagram harus muat satu halaman" ditolak (NA-07). |
| K-10 | Format caption: Ciputra / pola umum "Gambar 3.1" (angka Arab) vs satu skripsi "Gambar III.1" (angka Romawi). | **Tidak ada format universal (tingkat 3 berlaku per kampus).** Ikuti pedoman kampus target (SK-05); variasi Romawi adalah contoh satu karya, bukan aturan (NA-21). |
| K-11 | UBSI "sesuaikan simbol dengan referensi yang dipakai" vs klaim "notasi baku versi kampus X". | **Teks pedoman menang.** Tidak ada notasi UML baku versi kampus yang ditemukan; yang mengikat adalah konsistensi dengan SATU referensi yang dikutip (SK-11). |
| K-12 | R1 CF-07 (break): prosa §17.6.3.9 memakai kata "should" (anjuran) untuk cakupan semua lifeline, vs constraint OCL §17.12.3.7 yang memakainya sebagai keharusan. | **Constraint formal mengalahkan prosa "should" dalam satu sumber tingkat 1.** Break harus mencakup SEMUA lifeline enclosing fragment (CF-07). |
| K-13 | R2-21 (tampilkan multiplisitas pada relasi penting) vs R1 CL-04 (ujung tanpa multiplisitas = tidak dapat disimpulkan apa-apa). | **Bukan konflik.** R2-21 adalah rekomendasi kualitas (tampilkan detail kunci); CL-04 adalah aturan semantik (jangan menyimpulkan dari ketiadaan). Keduanya berlaku bersama: tampilkan multiplisitas bila penting, dan jangan nilai ujung tanpa angka sebagai salah. |
| K-14 | R3 D12: guard `[…]` didukung literatur untuk statechart + OMG untuk sequence, vs klaim guard sequence sebagai "aturan pedoman kampus". | **Sumber dipisahkan sesuai tingkatnya.** Guard sequence adalah aturan OMG tingkat 1 (CF-11, AC-02), bukan aturan pedoman kampus tingkat 3 (NA-24). Jangan klaim sebagai aturan pedoman kampus. |
| K-15 | P1 (eksperimen graf abstrak) vs pemakaian pada diagram UML. | **Batasan transfer dicantumkan.** P1 adalah ekstrapolasi yang masuk akal ke UML, bukan bukti langsung pada UML; jangan kutip P1 sebagai "terbukti pada diagram UML" tanpa catatan ini (§2.1). |

### 6.2 Kaidah pemakaian hierarki

1. **Temuan "pelanggaran formal" hanya boleh bersandar pada §1 (OMG).**
   Temuan dari §2 ditulis sebagai "rekomendasi kualitas (R2-xx)", temuan
   dari §3 ditulis sebagai "ketidaksesuaian pedoman kampus (SK-xx, kampus
   penerbit)".
2. **Bila §2 atau §3 tampak bertentangan dengan §1, §1 menang.** Contoh:
   rekomendasi "tampilkan multiplisitas" (§2) tidak mengubah arti
   "tanpa multiplisitas = tidak dapat disimpulkan" (§1, CL-04).
3. **Bila literatur/contoh (§3.2) bertentangan dengan pedoman kampus
   resmi (§3.1), pedoman kampus menang** untuk naskah kampus tersebut.
4. **Bila pedoman kampus tidak mengatur sesuatu (mis. notasi UML),
   kekosongan itu adalah hasil riset, bukan celah yang boleh diisi dengan
   mengarang aturan.** Rujuk ke OMG (§1) untuk notasi, dan catat bahwa
   pedoman kampus tidak mengaturnya.
5. **Setiap aturan baru di masa depan** wajib ditambahkan dengan format
   yang sama (ID + bunyi + sumber + cara cek) dan dengan bukti yang
   memenuhi tingkatnya; aturan tanpa sumber = JANGAN dimasukkan.

---

## Lampiran

### A. Rekap jumlah

| Kelompok | Jumlah |
|---|---|
| §1 Aturan WAJIB OMG — Sequence (SEQ-01 s.d. SEQ-26) | 26 |
| §1 Aturan WAJIB OMG — CombinedFragment (CF-01 s.d. CF-13) | 13 |
| §1 Aturan WAJIB OMG — Use Case (UC-01 s.d. UC-12) | 12 |
| §1 Aturan WAJIB OMG — Class (CL-01 s.d. CL-12) | 12 |
| §1 Aturan WAJIB OMG — Activity (AC-01 s.d. AC-12) | 12 |
| **Subtotal §1** | **75** |
| §2 Aturan KUALITAS jurnal (R2-01 s.d. R2-28, termasuk 3 larangan klaim: R2-04, R2-23, R2-26) | 28 |
| §3.1 Aturan KONVENSI SKRIPSI otoritatif (SK-01 s.d. SK-12) | 12 |
| §3.2 Praktik literatur/bahan ajar/contoh — BUKAN aturan (SK-L01 s.d. SK-L11) | 11 |
| §4 BUKAN ATURAN — JANGAN DITEGAKKAN (NA-01 s.d. NA-25) | 25 |
| §5 Koreksi premis R1 | 3 |
| §6 Konflik antar-sumber yang diselesaikan (K-01 s.d. K-15) | 15 |

### B. Yang SENGAJA TIDAK dimasukkan (tidak terverifikasi di sumber)

| Butir (ID R1) | Alasan |
|---|---|
| SEQ-27 — Kepala lifeline hasil create mulai tepat di titik pesan create | Tidak ada kalimat eksplisit di teks spesifikasi yang diakses (R1). Lihat NA-16. |
| SEQ-28 — Ujung lifeline yang dihancurkan ditandai tanda X | Tidak ditemukan di teks spesifikasi yang diakses; kemungkinan hanya pada gambar contoh (R1). Lihat NA-17. |
| CL-13 — Aturan notasi ERD dari OMG | ERD bukan diagram UML 2.5.1; tidak ada di spesifikasi (R1). Lihat NA-10. |

Batasan R1 yang dipertahankan: verifikasi dilakukan atas teks spesifikasi
yang dapat diakses (ekstraksi teks PDF). Isi gambar contoh (figure) tidak
semuanya terekspos sebagai teks — inilah penyebab SEQ-28 TIDAK
TERVERIFIKASI. Bila tim audit membutuhkan kepastian atas notasi X pada
DestructionOccurrenceSpecification, periksa Figure 17.14 / 17.17 spesifikasi
secara visual.

### C. Daftar sumber

1. R1 — OMG Unified Modeling Language, Version 2.5.1, OMG Document Number
   formal/2017-12-05, Desember 2017: https://www.omg.org/spec/UML/2.5.1/PDF
   (satu-satunya sumber §1). Sumber yang TIDAK dipakai R1: uml-diagrams.org
   (sekunder, tidak dikutip); dokumentasi StarUML (perilaku tool bukan
   spesifikasi).
2. R2 — 10 paper P1–P10 beserta DOI/URL masing-masing (lihat §2 per
   kelompok). Kandidat yang ditelusuri tetapi TIDAK dihitung sebagai paper
   terverifikasi (R2): Ambler (2005) — buku panduan praktisi; "Definition of
   the Criteria for Layout of the UML Use Case Diagrams" (venue/peer-review
   tidak terverifikasi); Glezer et al. (2005) — hanya sitasi ditemukan;
   Gemino & Parker (2009) — hanya sitasi ditemukan; Purchase et al. (2003)
   bab buku — temuan layout tidak berhasil diambil.
3. R3 — Pedoman resmi kampus A1–A6 + pola umum A7; literatur/bahan ajar
   B1–B7; contoh C1–C8 (lihat §3). OMG UML 2.5 (2015),
   https://www.omg.org/spec/UML/2.5, dicatat R3 sebagai standar
   internasional — otoritatif untuk notasi UML secara umum, tetapi BUKAN
   pedoman skripsi Indonesia dan TIDAK mengatur format naskah skripsi.

— Akhir dokumen —
