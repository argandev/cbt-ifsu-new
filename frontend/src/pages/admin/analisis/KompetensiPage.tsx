import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { Download, RefreshCw } from "lucide-react";
import apiClient from "../../../lib/api";
import { useAuthStore } from "../../../store/auth";
import { formatJam, formatTanggal } from "../../../lib/format";

import { Alert, Button, Select, Table, Breadcrumb, Typography, Form, Card, Statistic } from "antd";
interface KompetensiNilai {
  kompetensi_id: string | null;
  kompetensi_kode: string;
  kompetensi: string;
  nilai: number | null;
  jumlah_soal: number;
  dijawab: number;
  belum_dikoreksi: number;
  tidak_ada_soal: boolean;
  flag: string;
}

interface KompetensiSiswa {
  siswa_id: string;
  nisn: string;
  nama: string;
  kelas: string;
  nilai_per_kompetensi: KompetensiNilai[];
  rata_rata: number | null;
}

interface KompetensiKelas {
  kompetensi_id: string | null;
  kompetensi_kode: string;
  kompetensi: string;
  rata_rata: number | null;
  jumlah_siswa: number;
}

interface KompetensiData {
  jadwal_id: string;
  jumlah_siswa: number;
  rata_rata_kelas: number | null;
  rata_rata_per_kompetensi: KompetensiKelas[];
  siswa: KompetensiSiswa[];
}

function fmtNilai(n: number | null): string {
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

export default function KompetensiPage() {
  const { hasPermission } = useAuthStore();
  const [jadwalId, setJadwalId] = useState("");
  const [jadwalInput, setJadwalInput] = useState("");
  const [pesan, setPesan] = useState("");

  // Daftar jadwal untuk dropdown — endpoint yang sama dipakai JadwalPage.
  const jadwalQuery = useQuery({
    queryKey: ["jadwal"],
    queryFn: () => apiClient.get<{ items: JadwalItem[] }>("/api/v1/jadwal"),
  });
  const jadwalList = jadwalQuery.data?.items ?? [];

  const kompetensiQuery = useQuery({
    queryKey: ["kompetensi", jadwalId],
    queryFn: () => apiClient.get<KompetensiData>(`/api/v1/kompetensi/${encodeURIComponent(jadwalId)}`),
    enabled: jadwalId !== "",
  });

  useEffect(() => {
    setPesan("");
  }, [jadwalId]);

  const data = kompetensiQuery.data;
  const kolomKompetensi = useMemo(
    () => data?.rata_rata_per_kompetensi ?? [],
    [data],
  );
  const rows = useMemo(() => data?.siswa ?? [], [data]);

  const handleCari = () => {
    if (!jadwalInput.trim()) return;
    setJadwalId(jadwalInput.trim());
  };

  return (
    <div className="space-y-4">
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: "Nilai Kompetensi" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>Nilai per Kompetensi</Typography.Title>
          <Typography.Text type="secondary">Nilai tiap siswa per kompetensi (Σ poin / Σ bobot × 100). Siswa yang tidak menjawab satu pun soal suatu kompetensi mendapat nilai kosong (bukan 0) dengan keterangan &quot;tidak ada soal&quot;. Soal manual yang belum dikoreksi tidak dihitung (belum dinilai, bukan 0).</Typography.Text>
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
        <div className="grid grid-cols-2 gap-3 lg:grid-cols-3">
          <Card>
            <Statistic title="Siswa" value={data.jumlah_siswa} />
          </Card>
          <Card>
            <Statistic title="Rata-rata kelas" value={fmtNilai(data.rata_rata_kelas)} />
          </Card>
          <Card>
            <Statistic title="Jumlah kompetensi" value={kolomKompetensi.length} />
          </Card>
        </div>
      )}

      {jadwalId !== "" && (
        <div className="flex flex-wrap gap-2">
          {hasPermission("rekap.export") && (
            <Button onClick={() =>
                apiClient.download(
                  `/api/v1/kompetensi/${jadwalId}/export.csv`,
                  `nilai-kompetensi-${jadwalId}.csv`,
                )
              } type="default">
              <Download size={16} /> Export CSV
            </Button>
          )}
          <Button onClick={() => void kompetensiQuery.refetch()} type="default">
            <RefreshCw size={16} /> Muat Ulang
          </Button>
        </div>
      )}

      {kompetensiQuery.isPending && jadwalId !== "" && (
        <p className="text-sm text-muted">Memuat nilai per kompetensi...</p>
      )}
      {kompetensiQuery.isError && (
        <Alert type="error" showIcon message={<>Gagal memuat nilai per kompetensi: {(kompetensiQuery.error as Error).message}</>} />
      )}

      {jadwalId !== "" && kompetensiQuery.isSuccess && (
<Table
            columns={[
              {
                key: "nisn",
                title: "NISN",
                render: (_: unknown, s: any) => <span className="font-mono text-xs">{s.nisn || "-"}</span>,
              },
              {
                key: "nama",
                title: "Nama",
                render: (_: unknown, s: any) => <span className="font-medium">{s.nama || "-"}</span>,
              },
              {
                key: "kelas",
                title: "Kelas",
                render: (_: unknown, s: any) => <>{s.kelas || "-"}</>,
              },
              ...kolomKompetensi.map((k, idx) => ({
                key: k.kompetensi_id ?? `tanpa-${idx}`,
                title:
                  k.kompetensi_kode && k.kompetensi_kode !== "\u2013" ? k.kompetensi_kode : "Tanpa Kompetensi",
                render: (_: unknown, s: any) => {
                  const nilai = s.nilai_per_kompetensi[idx];
                  return (
                    <>
                      <p className="font-semibold">{fmtNilai(nilai?.nilai ?? null)}</p>
                      {nilai?.flag ? <p className="text-xs text-warning">{nilai.flag}</p> : null}
                    </>
                  );
                },
              })),
              {
                key: "rata",
                title: "Rata-rata",
                render: (_: unknown, s: any) => <span className="font-semibold">{fmtNilai(s.rata_rata)}</span>,
              },
            ]}
            dataSource={rows}
            rowKey="siswa_id"
            pagination={false}
            locale={{ emptyText: "Belum ada siswa yang menyelesaikan jadwal ini." }}
            scroll={{ x: "max-content" }}
            summary={() =>
              rows.length > 0 ? (
                <Table.Summary.Row className="bg-surface font-medium">
                  <Table.Summary.Cell index={0} colSpan={3}>
                    Rata-rata kelas (N pembagi eksplisit)
                  </Table.Summary.Cell>
                  {kolomKompetensi.map((k, idx) => (
                    <Table.Summary.Cell index={1 + idx} key={k.kompetensi_id ?? `tanpa-${idx}`}>
                      <p>{fmtNilai(k.rata_rata)}</p>
                      <p className="text-xs font-normal text-muted">N = {k.jumlah_siswa}</p>
                    </Table.Summary.Cell>
                  ))}
                  <Table.Summary.Cell index={1 + kolomKompetensi.length}>
                    {fmtNilai(data?.rata_rata_kelas ?? null)}
                  </Table.Summary.Cell>
                </Table.Summary.Row>
              ) : null
            }
          />
      )}
    </div>
  );
}
