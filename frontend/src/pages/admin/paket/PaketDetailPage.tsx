// PaketDetailPage — halaman detail paket + susunan soal (Tim B, penataan).
// Dipanggil dari PaketPage saat mode "detail". Panel "Soal dalam Paket"
// (dropdown bank + daftar soal + "+ Tambah" per soal + Generate Acak)
// dipindah utuh dari bawah halaman daftar lama ke sini — "menyusun paket"
// dan "mengisi soal" kini dua tugas yang tidak saling mengganggu.
// Semua endpoint, payload, dan perilaku tambah/hapus/generate TIDAK berubah.
import { Dices, Pencil, Trash2, ArrowLeft } from "lucide-react";
import { Link } from "react-router-dom";
import type { BankSoal, Soal } from "../../../lib/banksoal";
import { TIPE_SOAL } from "../../../lib/tipeSoal";

import { Alert, Button, InputNumber, Select, Breadcrumb, Space, Typography, Card } from "antd";
export interface PaketDetailData {
  id: string;
  kode: string;
  nama: string;
  mapel_id: string;
  mapel_nama?: string;
  jenis: string;
  durasi_menit: number;
  jumlah_soal: number;
  jumlah_soal_tersusun?: number;
  acak_soal: boolean;
  acak_jawaban: boolean;
  tampil_hasil: boolean;
}

export interface PaketSoalItem {
  soal_id: string;
  urutan: number;
  tipe: string;
  pertanyaan: string;
  bobot: number;
}

interface PaketDetailPageProps {
  paket: PaketDetailData;
  soalPaket: PaketSoalItem[];
  bankMapelSama: BankSoal[];
  bankDipilih: string;
  setBankDipilih: (id: string) => void;
  soalBank: Soal[];
  jumlahAcak: number;
  setJumlahAcak: (n: number) => void;
  pesan: string | null;
  error: string | null;
  onTambahSoal: (soalId: string) => void;
  onHapusSoal: (soalId: string) => void;
  onGenerateAcak: () => void;
  onUbah: () => void;
  onKembali: () => void;
}

export default function PaketDetailPage({
  paket,
  soalPaket,
  bankMapelSama,
  bankDipilih,
  setBankDipilih,
  soalBank,
  jumlahAcak,
  setJumlahAcak,
  pesan,
  error,
  onTambahSoal,
  onHapusSoal,
  onGenerateAcak,
  onUbah,
  onKembali,
}: PaketDetailPageProps) {
  return (
    <div className="space-y-4">
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: "Paket Ujian" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>{paket.nama}</Typography.Title>
          <Typography.Text type="secondary">{`${paket.kode} · ${paket.jenis} · ${paket.mapel_nama ?? ""} · ${paket.durasi_menit} menit · ${paket.jumlah_soal_tersusun ?? paket.jumlah_soal} soal tersusun`}</Typography.Text>
        </div>
        <Space wrap>
          <>
            <Button onClick={onKembali} type="default">
              <ArrowLeft size={16} /> Kembali ke Daftar
            </Button>
            <Button onClick={onUbah} type="primary">
              <Pencil size={16} /> Ubah Paket
            </Button>
          </>
        </Space>
      </div>
      {pesan && <Alert type="success" showIcon message={<>{pesan}</>} />}
      {error && <Alert type="error" showIcon message={<>{error}</>} />}

      {/* Seksi 1 — Info paket */}
      <Card title="Info Paket">
        <dl className="mt-3 grid grid-cols-1 gap-x-8 gap-y-2 text-sm sm:grid-cols-2">
          <div className="flex justify-between gap-4"><dt className="text-muted">Kode</dt><dd className="font-medium">{paket.kode}</dd></div>
          <div className="flex justify-between gap-4"><dt className="text-muted">Jenis</dt><dd className="font-medium">{paket.jenis}</dd></div>
          <div className="flex justify-between gap-4"><dt className="text-muted">Mapel</dt><dd className="font-medium">{paket.mapel_nama ?? "-"}</dd></div>
          <div className="flex justify-between gap-4"><dt className="text-muted">Durasi</dt><dd className="font-medium">{paket.durasi_menit} menit</dd></div>
          <div className="flex justify-between gap-4"><dt className="text-muted">Acak soal</dt><dd className="font-medium">{paket.acak_soal ? "Ya" : "Tidak"}</dd></div>
          <div className="flex justify-between gap-4"><dt className="text-muted">Acak jawaban</dt><dd className="font-medium">{paket.acak_jawaban ? "Ya" : "Tidak"}</dd></div>
          <div className="flex justify-between gap-4"><dt className="text-muted">Tampilkan hasil ke siswa</dt><dd className="font-medium">{paket.tampil_hasil ? "Ya" : "Tidak"}</dd></div>
        </dl>
      </Card>

      {/* Seksi 2 — Soal dalam Paket */}
      <Card title={`Soal dalam Paket: ${paket.nama}`}>
        <div className="mt-3 flex flex-col gap-2 sm:flex-row">
          <div className="flex flex-wrap gap-2">
            <InputNumber min={1} value={jumlahAcak} onChange={(v) => setJumlahAcak(typeof v === "number" ? v : 0)} aria-label="Jumlah soal acak" size="large" style={{ width: "100%" }} />
            <Button onClick={onGenerateAcak} type="primary">
              <Dices size={16} /> Generate Acak
            </Button>
          </div>
        </div>

        {bankMapelSama.length > 0 && (
          <div className="mt-3 flex flex-col gap-2 sm:flex-row">
            <div className="flex-1">
              <Select showSearch options={[
                  { value: "", label: "— Pilih bank soal mapel ini —" },
                  ...bankMapelSama.map((b) => ({ value: b.id, label: `${b.nama} (${b.jumlah_soal} soal)` })),
                ]} value={bankDipilih} onChange={(v) => setBankDipilih(v)} placeholder="— Pilih bank soal mapel ini —" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" />
            </div>
          </div>
        )}
        {bankDipilih && (
          <ul className="mt-2 max-h-64 space-y-1 overflow-y-auto rounded-lg border border-border p-2">
            {soalBank.map((s) => (
              <li key={s.id} className="flex items-center justify-between gap-2 text-sm">
                <span className="truncate">
                  [{TIPE_SOAL[s.tipe]?.label ?? s.tipe}] {s.pertanyaan}
                </span>
                <Button onClick={() => onTambahSoal(s.id)} type="default">
                  + Tambah
                </Button>
              </li>
            ))}
            {soalBank.length === 0 && (
              <li className="text-muted">Bank ini belum berisi soal.</li>
            )}
          </ul>
        )}

        <ol className="mt-3 space-y-2">
          {soalPaket.map((s) => (
            <li key={s.soal_id} className="flex items-center justify-between gap-2 rounded-lg border border-border px-3 py-2">
              <span className="text-sm">
                <strong>{s.urutan}.</strong> [{TIPE_SOAL[s.tipe]?.label ?? s.tipe}] {s.pertanyaan}
                <span className="ml-2 text-xs text-muted">Bobot {s.bobot}</span>
              </span>
              <Button aria-label={`Hapus soal ${s.urutan}`} onClick={() => onHapusSoal(s.soal_id)} type="default" danger>
                <Trash2 size={16} />
              </Button>
            </li>
          ))}
          {soalPaket.length === 0 && (
            <li className="text-muted">Paket belum berisi soal.</li>
          )}
        </ol>
      </Card>
    </div>
  );
}
