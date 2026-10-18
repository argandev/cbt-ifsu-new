import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { Download, RefreshCw } from "lucide-react";
import apiClient from "../../../lib/api";
import { formatJam, formatTanggal } from "../../../lib/format";

import { Alert, Button, Select, Table, Tag, Breadcrumb, Typography, Form, Card, Statistic } from "antd";
import { statusTagColor } from "../../../components/common/statusTag";
interface Pengecoh {
  opsi: string;
  teks?: string;
  pemilih: number;
  persen: number;
  berfungsi: boolean;
}

interface AnalisisItem {
  soal_id: string;
  urutan: number;
  pertanyaan: string;
  tipe: string;
  bobot: number;
  kompetensi_kode: string;
  kompetensi: string;
  jumlah_penjab: number;
  jumlah_benar: number;
  jumlah_tidak_dijawab: number;
  belum_dikoreksi: number;
  kesukaran: number | null;
  label_kesukaran: string;
  daya_beda: number | null;
  interpretasi_daya_beda: string;
  pengecoh: Pengecoh[];
  pengecoh_status: string;
}

interface AnalisisData {
  jadwal_id: string;
  jumlah_peserta: number;
  jumlah_soal: number;
  metode_kelompok: string;
  sampel_kecil: boolean;
  catatan: string;
  items: AnalisisItem[];
}

function fmt2(n: number | null): string {
  return n === null || n === undefined ? "-" : n.toFixed(2);
}

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


export default function AnalisisButirPage() {
  const [jadwalId, setJadwalId] = useState("");
  const [jadwalInput, setJadwalInput] = useState("");
  const [pesan, setPesan] = useState("");

  // Daftar jadwal untuk dropdown — endpoint yang sama dipakai JadwalPage.
  const jadwalQuery = useQuery({
    queryKey: ["jadwal"],
    queryFn: () => apiClient.get<{ items: JadwalItem[] }>("/api/v1/jadwal"),
  });
  const jadwalList = jadwalQuery.data?.items ?? [];

  const analisisQuery = useQuery({
    queryKey: ["analisis", jadwalId],
    queryFn: () => apiClient.get<AnalisisData>(`/api/v1/analisis/${encodeURIComponent(jadwalId)}`),
    enabled: jadwalId !== "",
  });

  useEffect(() => {
    setPesan("");
  }, [jadwalId]);

  const items = useMemo(() => analisisQuery.data?.items ?? [], [analisisQuery.data]);
  const data = analisisQuery.data;

  const handleCari = () => {
    if (!jadwalInput.trim()) return;
    setJadwalId(jadwalInput.trim());
  };

  return (
    <div className="space-y-4">
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: "Analisis Butir" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>Analisis Butir Soal</Typography.Title>
          <Typography.Text type="secondary">Tingkat kesukaran, daya beda, dan analisis pengecoh per soal — dihitung dari attempt yang sudah selesai pada satu jadwal. Soal manual yang belum dikoreksi tidak dihitung sebagai benar/salah.</Typography.Text>
        </div>
      </div>

      <Form layout="vertical" onFinish={handleCari} className="flex flex-col gap-2 sm:flex-row sm:items-end">
        <Form.Item label="Jadwal" className="mb-0 w-full sm:max-w-sm">
          <Select showSearch options={[
              { value: "", label: "— Pilih jadwal —" },
              ...jadwalList.map((j) => ({ value: j.id, label: labelJadwal(j) })),
            ]} value={jadwalInput} onChange={(v) => setJadwalInput(v)} placeholder="— Pilih jadwal —" id="jadwal-id" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" />
        </Form.Item>
        <Button htmlType="submit" type="primary">Tampilkan</Button>
      </Form>
      {jadwalQuery.isError && (
        <Alert type="error" showIcon message="Gagal memuat daftar jadwal." />
      )}

      {pesan !== "" && (
        <Alert type="info" showIcon message={<>
          {pesan}
        </>} />
      )}

      {jadwalId !== "" && data && (
        <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
          {[
            { label: "Peserta selesai", value: String(data.jumlah_peserta) },
            { label: "Jumlah soal", value: String(data.jumlah_soal) },
            { label: "Metode kelompok", value: data.metode_kelompok },
            {
              label: "Catatan",
              value: data.sampel_kecil ? "Sampel kecil" : "Sampel cukup",
            },
          ].map((s) => (
            <Card key={s.label}>
              <Statistic title={s.label} value={s.value} />
            </Card>
          ))}
          {data.catatan !== "" && (
            <p className="col-span-2 text-sm text-warning lg:col-span-4">{data.catatan}</p>
          )}
        </div>
      )}

      {jadwalId !== "" && (
        <div className="flex flex-wrap gap-2">
          <Button onClick={() =>
              apiClient.download(
                `/api/v1/analisis/${jadwalId}/export.csv`,
                `analisis-butir-${jadwalId}.csv`,
              )
            } type="default">
            <Download size={16} /> Export CSV
          </Button>
          <Button onClick={() => void analisisQuery.refetch()} type="default">
            <RefreshCw size={16} /> Muat Ulang
          </Button>
        </div>
      )}

      {analisisQuery.isPending && jadwalId !== "" && (
        <p className="text-sm text-muted">Memuat analisis...</p>
      )}
      {analisisQuery.isError && (
        <Alert type="error" showIcon message={<>Gagal memuat analisis: {(analisisQuery.error as Error).message}</>} />
      )}

      {jadwalId !== "" && analisisQuery.isSuccess && (
<Table
            columns={[
              {
                key: "no",
                title: "No",
                render: (_: unknown, _item: any, index?: number) => <>{(index ?? 0) + 1}</>,
              },
              {
                key: "soal",
                title: "Soal",
                render: (_: unknown, item: any) => (
                  <>
                    <p className="font-medium text-ink">{item.pertanyaan}</p>
                    <p className="mt-1 text-xs text-muted">
                      {item.kompetensi_kode !== "" && item.kompetensi_kode !== "-"
                        ? `Kompetensi: ${item.kompetensi_kode}`
                        : "Tanpa kompetensi"}
                      {item.belum_dikoreksi > 0 && ` · ${item.belum_dikoreksi} belum dikoreksi`}
                      {item.jumlah_tidak_dijawab > 0 && ` · ${item.jumlah_tidak_dijawab} tidak dijawab`}
                    </p>
                  </>
                ),
              },
              {
                key: "tipe",
                title: "Tipe",
                render: (_: unknown, item: any) => <span className="font-mono text-xs">{item.tipe}</span>,
              },
              {
                key: "penjab",
                title: "Penjab / Benar",
                render: (_: unknown, item: any) => (
                  <>
                    {item.jumlah_penjab} / {item.jumlah_benar}
                  </>
                ),
              },
              {
                key: "kesukaran",
                title: "Kesukaran",
                render: (_: unknown, item: any) => (
                  <>
                    <p className="font-semibold">{fmt2(item.kesukaran)}</p>
                    <Tag color={statusTagColor(item.label_kesukaran)}>{item.label_kesukaran}</Tag>
                  </>
                ),
              },
              {
                key: "daya-beda",
                title: "Daya Beda",
                render: (_: unknown, item: any) => (
                  <>
                    <p className="font-semibold">{fmt2(item.daya_beda)}</p>
                    <Tag color={statusTagColor(item.interpretasi_daya_beda)}>
                      {item.interpretasi_daya_beda}
                    </Tag>
                  </>
                ),
              },
              {
                key: "pengecoh",
                title: "Pengecoh per Opsi",
                render: (_: unknown, item: any) => (
                  <>
                    {item.pengecoh_status === "n/a" ? (
                      <span className="text-muted">N/A</span>
                    ) : (
                      <ul className="space-y-1">
                        {item.pengecoh.map((p: any) => (
                          <li key={p.opsi} className="flex flex-wrap items-center gap-2">
                            <span className="font-semibold">{p.opsi}</span>
                            <span className="text-muted">
                              {p.pemilih} ({p.persen.toFixed(2)}%)
                            </span>
                            {!p.berfungsi && <Tag color="red">tidak berfungsi</Tag>}
                          </li>
                        ))}
                      </ul>
                    )}
                  </>
                ),
              },
            ]}
            dataSource={items}
            rowKey="soal_id"
            pagination={false}
            locale={{ emptyText: "Belum ada soal untuk jadwal ini." }}
            scroll={{ x: "max-content" }}
          />
      )}
    </div>
  );
}
