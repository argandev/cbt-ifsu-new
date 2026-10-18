import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { Printer } from "lucide-react";
import apiClient from "../../../lib/api";
import { formatJam, formatTanggal, formatTanggalJam } from "../../../lib/format";

import { Alert, Button, Descriptions, Select, Breadcrumb, Space, Typography, Form } from "antd";
interface JadwalItem {
  id: string;
  paket_nama?: string;
  paket_kode?: string;
  kelas_nama?: string;
  tanggal: string;
  jam_mulai: string;
  jam_selesai: string;
  status: string;
}

function labelJadwal(j: JadwalItem): string {
  const paket = j.paket_nama ?? j.paket_kode ?? "Paket";
  const kelas = j.kelas_nama ? ` · ${j.kelas_nama}` : "";
  return `${paket}${kelas} · ${formatTanggal(j.tanggal)} · ${formatJam(j.jam_mulai)}–${formatJam(j.jam_selesai)} WIB`;
}

interface BeritaAcaraData {
  jadwal_id: string;
  mapel: string;
  kelas: string;
  tanggal: string | null;
  jam_mulai: string | null;
  jam_selesai: string | null;
  jumlah_peserta: number;
  hadir: number;
  tidak_hadir: number;
  rata_rata: number | null;
  lulus: number;
  catatan: string;
  generated_at: string;
}

export default function BeritaAcaraPage() {
  // ID jadwal dari route /adm/rekap/:jadwalId/berita-acara (kontrak §9).
  // Form input manual hanya tampil bila halaman dibuka tanpa param.
  const { jadwalId: paramJadwalId } = useParams();
  const [jadwalId, setJadwalId] = useState(paramJadwalId ?? "");
  const [jadwalInput, setJadwalInput] = useState("");

  useEffect(() => {
    if (paramJadwalId) setJadwalId(paramJadwalId);
  }, [paramJadwalId]);

  // Daftar jadwal untuk dropdown (hanya bila dibuka tanpa param route)
  // — endpoint yang sama dipakai JadwalPage.
  const jadwalQuery = useQuery({
    queryKey: ["jadwal"],
    queryFn: () => apiClient.get<{ items: JadwalItem[] }>("/api/v1/jadwal"),
    enabled: !paramJadwalId,
  });
  const jadwalList = jadwalQuery.data?.items ?? [];

  const query = useQuery({
    queryKey: ["berita-acara", jadwalId],
    queryFn: () => apiClient.get<BeritaAcaraData>(`/api/v1/rekap/${jadwalId}/berita-acara`),
    enabled: jadwalId !== "",
  });

  const handleCari = () => {
    if (!jadwalInput.trim()) return;
    setJadwalId(jadwalInput.trim());
  };

  const data = query.data;

  return (
    <div className="space-y-4">
      <div className="print:hidden">
        <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: (<Link to="/adm/rekap">Rekap</Link>) }, { title: "Berita Acara" }]} />
        <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="min-w-0">
            <Typography.Title level={3} style={{ margin: 0 }}>Berita Acara Ujian</Typography.Title>
            <Typography.Text type="secondary">Data diambil dari API; halaman ini siap cetak (Print menghasilkan dokumen print-friendly).</Typography.Text>
          </div>
          <Space>
            <Link to="/adm/rekap"><Button>Kembali ke Rekap</Button></Link>
          </Space>
        </div>
      </div>

      {paramJadwalId ? (
        <div className="flex flex-wrap items-center gap-2 print:hidden">
          {data && (
            <p className="text-sm text-muted">
              Jadwal:{" "}
              {[data.mapel, data.kelas, ...(data.tanggal ? [formatTanggal(data.tanggal)] : []), [data.jam_mulai, data.jam_selesai].filter(Boolean).map((j) => formatJam(j)).join(" - ")]
                .filter(Boolean)
                .join(" · ") || "-"}
            </p>
          )}
          {data && (
            <Button onClick={() => window.print()} type="default">
              <Printer size={16} /> Print
            </Button>
          )}
        </div>
      ) : (
        <Form layout="vertical" onFinish={handleCari} className="flex flex-col gap-2 print:hidden sm:flex-row sm:items-end">
          <Form.Item label="Jadwal" className="mb-0 w-full sm:max-w-sm">
            <Select showSearch options={[
                { value: "", label: "— Pilih jadwal —" },
                ...jadwalList.map((j) => ({ value: j.id, label: labelJadwal(j) })),
              ]} value={jadwalInput} onChange={(v) => setJadwalInput(v)} placeholder="— Pilih jadwal —" id="jadwal-id-ba" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" />
          </Form.Item>
          <Button htmlType="submit" type="primary">
            Tampilkan
          </Button>
          {data && (
            <Button onClick={() => window.print()} type="default">
              <Printer size={16} /> Print
            </Button>
          )}
        </Form>
      )}

      {query.isPending && jadwalId !== "" && <p className="text-sm text-muted print:hidden">Memuat...</p>}
      {query.isError && (
        <Alert type="error" showIcon message={<>Gagal memuat berita acara: {(query.error as Error).message}</>} className="print:hidden" />
      )}

      {data && (
        <article className="mx-auto max-w-3xl rounded-lg border border-border bg-white p-6 print:border-0 print:shadow-none sm:p-10">
          <header className="border-b-2 border-anbk pb-4 text-center">
            <h2 className="text-xl font-bold uppercase tracking-wide text-anbk">
              Berita Acara Ujian
            </h2>
            <p className="mt-1 text-sm text-muted">Computer Based Test (CBT) SMK</p>
          </header>

          <Descriptions
            column={{ xs: 1, sm: 2 }}
            bordered
            size="small"
            className="mt-6"
            items={[
              { key: "mapel", label: "Mata Pelajaran", children: data.mapel || "-" },
              { key: "kelas", label: "Kelas", children: data.kelas || "-" },
              { key: "tanggal", label: "Tanggal", children: formatTanggal(data.tanggal) },
              { key: "jam", label: "Jam", children: `${formatJam(data.jam_mulai)} - ${formatJam(data.jam_selesai)}` },
              { key: "jumlah", label: "Jumlah Peserta", children: data.jumlah_peserta },
              { key: "hadir", label: "Hadir", children: data.hadir },
              { key: "tidak-hadir", label: "Tidak Hadir", children: data.tidak_hadir },
              { key: "rata", label: "Rata-rata Nilai", children: data.rata_rata === null ? "-" : data.rata_rata.toFixed(2) },
              { key: "lulus", label: "Jumlah Lulus (>=70)", children: data.lulus },
            ]}
          />

          <section className="mt-6">
            <h3 className="text-sm font-semibold uppercase text-anbk">Catatan / Kejadian</h3>
            <p className="mt-2 min-h-20 whitespace-pre-wrap rounded-lg border border-border p-3 text-sm">
              {data.catatan || "Tidak ada catatan."}
            </p>
          </section>

          <section className="mt-10 grid grid-cols-2 gap-6 text-center text-sm sm:grid-cols-3">
            {["Proktor", "Pengawas", "Kepala Sekolah"].map((role) => (
              <div key={role}>
                <p className="text-muted">{role}</p>
                <div className="mx-auto mt-10 border-b border-border pb-1" />
                <p className="mt-2 text-muted">(Nama terang &amp; tanda tangan)</p>
              </div>
            ))}
          </section>

          <p className="mt-8 text-center text-xs text-muted">
            Dicetak dari CBT SMK v2 pada {formatTanggalJam(data.generated_at)}
          </p>
        </article>
      )}
    </div>
  );
}
