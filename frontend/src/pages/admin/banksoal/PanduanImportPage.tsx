// PanduanImportPage — panduan ringkas import soal untuk guru (Tim Panduan + Template).
// Versi lengkap ada di docs/PANDUAN-IMPORT-SOAL.md. Halaman ini hanya
// merangkum: langkah-langkah, tabel tipe soal, error paling umum, contoh.
// Tanpa CDN, tanpa logo/nama instansi resmi. Label jujur: 5 tipe standar
// AKM + 11 varian/ekstensi aplikasi.
import { Link } from "react-router-dom";
import { HomeOutlined } from "@ant-design/icons";

import { Alert, Button, Card, Steps, Table, Breadcrumb, Space, Typography } from "antd";
const LANGKAH = [
  {
    judul: "1. Unduh template",
    isi: "Di halaman Import Soal, klik Unduh Template dan pilih Excel (.xlsx) atau Word (.docx). Isi file template itu — jangan membuat file baru dari nol. Satu file = satu bank soal (maksimal 500 soal, maksimal 5 MB).",
  },
  {
    judul: "2. Isi template",
    isi: "Excel: isi HANYA sheet “Soal” (satu baris = satu soal). Word: satu blok SOAL = satu soal. Kolom tipe / label TIPE: diisi KODE (mis. pg), bukan nama panjang. Cara menulis kunci BERBEDA per tipe — lihat tabel di bawah.",
  },
  {
    judul: "3. Unggah file",
    isi: "Pilih file yang sudah diisi di dialog Import Soal, pastikan bank soal tujuannya benar, lalu klik Pratinjau (bukan langsung Simpan).",
  },
  {
    judul: "4. Baca hasil pratinjau",
    isi: "Periksa ringkasan (valid / error / warning / duplikat) dan tabel pratinjau. Yang terpenting: kunci ditampilkan sebagai label + teks opsi yang dirujuknya (mis. B — Semarang). Pastikan hurufnya menunjuk opsi yang benar-benar Anda maksud.",
  },
  {
    judul: "5. Pilih mode + duplikat",
    isi: "Mode: tolak semua bila ada error (default, paling aman) atau lewati baris yang error. Duplikat: lewati (default) atau simpan ganda bila memang disengaja.",
  },
  {
    judul: "6. Simpan + cek di bank soal",
    isi: "Klik Simpan, lalu buka “Lihat bank soal”. Soal masuk berstatus DRAF — wajib ditelaah (kunci, opsi, stimulus, gambar, rubrik) sebelum dipakai di paket ujian.",
  },
];

const TIPE = [
  { kode: "pg", kunci: "Satu label: B", contoh: "pg → kunci B", kelompok: "Inti AKM" },
  { kode: "pg_kompleks", kunci: "Label dipisah koma: A,C", contoh: "pg_kompleks → kunci A,C", kelompok: "Inti AKM" },
  { kode: "menjodohkan", kunci: "Pasangan kiri::kanan, antar-pasangan dipisah ;", contoh: "Bandung::Jawa Barat;Semarang::Jawa Tengah", kelompok: "Inti AKM" },
  { kode: "isian", kunci: "Jawaban, alternatif dipisah |", contoh: "Jakarta|DKI Jakarta", kelompok: "Inti AKM" },
  { kode: "uraian", kunci: "Rubrik (boleh kosong + warning)", contoh: "Rubrik: pengertian (50), contoh (50).", kelompok: "Inti AKM" },
  { kode: "benar_salah", kunci: "B/S dipisah koma, sebanyak pernyataan", contoh: "3 pernyataan → B,S,B", kelompok: "Varian" },
  { kode: "pg_multi", kunci: "Label dipisah koma: A,C", contoh: "Sama seperti pg_kompleks", kelompok: "Varian" },
  { kode: "urutkan", kunci: "Label dalam urutan benar, dipisah koma", contoh: "4 butir → B,C,A,D", kelompok: "Varian" },
  { kode: "isi_ganda", kunci: "Jawaban per kolom dipisah |", contoh: "50|30", kelompok: "Varian" },
  { kode: "teks_rumpang", kunci: "Jawaban per rumpang dipisah |", contoh: "___ ada 2 → Jakarta|Rupiah", kelompok: "Varian" },
  { kode: "peta_konsep", kunci: "Sama seperti menjodohkan", contoh: "Gula::Manis;Jeruk::Asam", kelompok: "Varian" },
  { kode: "studi_kasus", kunci: "Satu label: B (+ stimulus wajib)", contoh: "studi_kasus → kunci B", kelompok: "Varian" },
  { kode: "pernyataan_sikap", kunci: "Diabaikan (skala 1–4, tidak dinilai)", contoh: "Kunci boleh kosong", kelompok: "Varian" },
  { kode: "upload_file", kunci: "Rubrik (boleh kosong + warning)", contoh: "Rubrik: kelengkapan (40), format (30)…", kelompok: "Varian" },
  { kode: "praktik_singkat", kunci: "Rubrik (boleh kosong + warning)", contoh: "Rubrik: urutan benar (50)…", kelompok: "Varian" },
  { kode: "pilihan_gambar", kunci: "Satu label: A (opsi = tautan gambar)", contoh: "opsi_a: https://…/utp.png", kelompok: "Varian" },
];

const ERROR_UMUM = [
  { kode: "E2", sebab: "Kode tipe tidak dikenal", perbaiki: "Tulis kode persis dari tabel di atas (mis. pg). Jangan tulis “Pilihan Ganda”." },
  { kode: "E4", sebab: "Kunci kosong padahal tipe dinilai otomatis", perbaiki: "Isi kunci sesuai format tipe. Kunci boleh kosong hanya untuk uraian / upload_file / praktik_singkat (warning rubrik)." },
  { kode: "E13", sebab: "Jumlah nilai B/S tidak sama dengan jumlah pernyataan", perbaiki: "3 pernyataan → 3 nilai, mis. B,S,B." },
  { kode: "E17", sebab: "Pasangan kunci menjodohkan salah format", perbaiki: "Pakai kiri::kanan dan antar-pasangan ; — mis. Bandung::Jawa Barat;Semarang::Jawa Tengah." },
  { kode: "E25", sebab: "Jumlah jawaban tidak sama dengan jumlah ___", perbaiki: "Samakan jumlah jawaban dengan jumlah ___ di pertanyaan." },
  { kode: "E26", sebab: "Stimulus kosong untuk studi_kasus", perbaiki: "Isi stimulus / bacaannya." },
  { kode: "E28 / E29", sebab: "Tautan gambar bukan URL http(s) yang valid", perbaiki: "Isi tautan lengkap diawali https://. Gambar tempelan di file tidak ikut terimpor." },
  { kode: "E30 / E31", sebab: "Kode kisi-kisi / kompetensi tidak ditemukan", perbaiki: "Salin kode persis dari halaman Kisi-kisi / Kompetensi, atau kosongkan." },
  { kode: "W1", sebab: "Rubrik belum diisi (warning, tetap tersimpan)", perbaiki: "Lengkapi rubrik di editor sebelum soal dipakai." },
  { kode: "W4 / W5", sebab: "Duplikat dalam file / dengan bank (warning)", perbaiki: "Pilih lewati (default) agar soal tidak masuk dua kali, atau simpan ganda bila disengaja." },
];

export default function PanduanImportPage() {
  return (
    <div className="space-y-6">
      <Breadcrumb items={[{ title: (<Link to="/" aria-label="Beranda"><HomeOutlined /></Link>) }, { title: (<Link to="/adm/bank-soal">Bank Soal</Link>) }, { title: "Panduan Import Soal" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>Panduan Import Soal</Typography.Title>
          <Typography.Text type="secondary">Cara memasukkan banyak soal sekaligus lewat file Excel (.xlsx) atau Word (.docx). Panduan lengkap untuk guru ada di dokumen PANDUAN-IMPORT-SOAL.md — halaman ini adalah ringkasannya.</Typography.Text>
        </div>
        <Space>
          <Button href="/adm/bank-soal">Kembali ke Bank Soal</Button>
        </Space>
      </div>

<Card title="Langkah-langkah">
        <Steps
          direction="vertical"
          current={LANGKAH.length}
          items={LANGKAH.map((l) => ({ title: l.judul, description: l.isi }))}
        />
        <Alert
          type="warning"
          showIcon
          message="Soal hasil import berstatus DRAF — wajib ditelaah sebelum dipakai. Kunci ditegaskan guru, tidak ditebak mesin."
          className="mt-3"
        />
      </Card>

      <Card title="Tabel tipe soal — cara menulis kunci">
        <p className="mt-1 text-sm text-muted">
          5 tipe standar AKM + 11 varian/ekstensi aplikasi. Pemisah kunci tidak sama antar-tipe — baca baris tipe Anda dengan teliti.
        </p>
          <Table columns={[{ key: "col0", title: "Kode tipe", dataIndex: "kode", sorter: (a: any, b: any) => { const _va = a.kode, _vb = b.kode; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, t: any) => (<>{t.kode}</>) }, { key: "col1", title: "Kelompok", dataIndex: "kelompok", sorter: (a: any, b: any) => { const _va = a.kelompok, _vb = b.kelompok; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, t: any) => (<>{t.kelompok}</>) }, { key: "col2", title: "Cara menulis kunci", dataIndex: "kunci", sorter: (a: any, b: any) => { const _va = a.kunci, _vb = b.kunci; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, t: any) => (<>{t.kunci}</>) }, { key: "col3", title: "Contoh", dataIndex: "contoh", sorter: (a: any, b: any) => { const _va = a.contoh, _vb = b.contoh; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, t: any) => (<>{t.contoh}</>) }]} dataSource={TIPE} rowKey="kode" pagination={false} locale={{ emptyText: "Belum ada data" }} scroll={{ x: "max-content" }} />
        <p className="mt-3 text-sm text-muted">
          Opsi: untuk tipe yang memakai opsi, isi minimal 2 opsi. Menjodohkan / peta konsep: satu opsi = satu pasangan{" "}
          <code className="font-mono text-xs">kiri::kanan</code>. Pilihan gambar: satu opsi = satu tautan gambar. Isian / isi_ganda /
          teks_rumpang / uraian / upload_file / praktik_singkat: kolom opsi tidak dipakai.
        </p>
      </Card>

      <Card title="Contoh — benar vs salah">
        <div className="mt-3 grid gap-3 md:grid-cols-2">
          <Alert type="success" showIcon message="Benar" description={<>
            <pre className="mt-2 overflow-x-auto whitespace-pre-wrap font-mono text-xs leading-relaxed text-ink">{`SOAL 1
TIPE: pg
PERTANYAAN: Ibu kota provinsi Jawa Barat adalah …
A. Bandung
B. Semarang
C. Medan
D. Surabaya
KUNCI: A`}</pre>
            <p className="mt-2 text-sm text-ink">Excel: kolom tipe = <code className="font-mono text-xs">pg</code>, kolom kunci = <code className="font-mono text-xs">A</code>.</p>
          </>} />
          <Alert type="error" showIcon message="Salah" description={<>
            <ul className="mt-2 list-disc space-y-1 pl-5 text-sm text-ink">
              <li>TIPE: Pilihan Ganda — harus kode <code className="font-mono text-xs">pg</code>.</li>
              <li>Kunci menjodohkan <code className="font-mono text-xs">Bandung-Jawa Barat</code> — harus <code className="font-mono text-xs">Bandung::Jawa Barat</code>.</li>
              <li>Kunci PG kompleks <code className="font-mono text-xs">A dan C</code> — harus <code className="font-mono text-xs">A,C</code>.</li>
              <li>Kunci isian <code className="font-mono text-xs">Jakarta atau DKI Jakarta</code> — harus <code className="font-mono text-xs">Jakarta|DKI Jakarta</code>.</li>
              <li>Gambar ditempel di file — harus tautan di kolom <code className="font-mono text-xs">media_url</code> / label <code className="font-mono text-xs">MEDIA:</code>.</li>
            </ul>
          </>} />
        </div>
        <p className="mt-3 text-sm text-muted">
          Rumus matematika ditulis sebagai teks LaTeX di pertanyaan, diapit tanda dolar — mis.{" "}
          <code className="font-mono text-xs">Hitung nilai dari $2x + 3$ untuk $x = 4$.</code> Di halaman ujian rumus dirender
          sebagai rumus matematika. Jangan menulis rumus Excel / formula di sel, dan jangan mengandalkan equation editor Word.
        </p>
      </Card>

      <Card title="Error paling umum + cara memperbaiki">
        <p className="mt-1 text-sm text-muted">
          Semua pesan memakai nomor baris Excel (baris 1 = header, data mulai baris 2) atau nomor soal Word (SOAL N). Katalog
          lengkap E1–E32 + warning W1–W7 ada di panduan lengkap.
        </p>
          <Table columns={[{ key: "col0", title: "Kode", dataIndex: "kode", sorter: (a: any, b: any) => { const _va = a.kode, _vb = b.kode; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, e: any) => (<>{e.kode}</>) }, { key: "col1", title: "Sebab", dataIndex: "sebab", sorter: (a: any, b: any) => { const _va = a.sebab, _vb = b.sebab; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, e: any) => (<>{e.sebab}</>) }, { key: "col2", title: "Cara memperbaiki", dataIndex: "perbaiki", sorter: (a: any, b: any) => { const _va = a.perbaiki, _vb = b.perbaiki; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, e: any) => (<>{e.perbaiki}</>) }]} dataSource={ERROR_UMUM} rowKey="kode" pagination={false} locale={{ emptyText: "Belum ada data" }} scroll={{ x: "max-content" }} />
      </Card>

      <Card title="Mode + duplikat — kapan pakai yang mana">
        <div className="mt-3 grid gap-3 md:grid-cols-2">
          <div>
            <Typography.Text strong className="block">Mode</Typography.Text>
            <ul className="mt-2 list-disc space-y-1 pl-5 text-sm text-ink">
              <li><strong>Tolak semua bila ada error</strong> (default): ada satu baris error → tidak ada yang tersimpan. Pakai ini dalam hampir semua kasus.</li>
              <li><strong>Lewati baris yang error</strong>: baris valid tersimpan, baris error dilaporkan. Pakai bila file besar dan sebagian soal harus masuk sekarang.</li>
            </ul>
          </div>
          <div>
            <Typography.Text strong className="block">Duplikat</Typography.Text>
            <ul className="mt-2 list-disc space-y-1 pl-5 text-sm text-ink">
              <li><strong>Lewati</strong> (default): soal yang pertanyaannya sama (setelah dirapikan) tidak disimpan dua kali. Pakai bila mengunggah ulang file yang sama.</li>
              <li><strong>Simpan ganda</strong>: semua baris valid disimpan termasuk yang duplikat. Pakai hanya bila memang disengaja.</li>
            </ul>
          </div>
        </div>
        <p className="mt-3 text-sm text-muted">
          Duplikat adalah warning, bukan error — duplikat tidak pernah menggagalkan file. CSV tidak didukung di UI baru; pindahkan
          isi CSV lama ke template Excel. Google Docs tidak didukung langsung — ekspor ke Word dulu, lalu sesuaikan ke format di atas.
        </p>
      </Card>
    </div>
  );
}
