import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Download, FileText, Lock, RefreshCw } from "lucide-react";
import apiClient from "../../../lib/api";
import { formatJam, formatTanggal } from "../../../lib/format";
import { useAuthStore } from "../../../store/auth";

import { Alert, Button, InputNumber, Select, Table, Tag, Breadcrumb, Typography, Form, Card, Statistic } from "antd";
interface NilaiRow {
  id: string;
  attempt_id: string;
  jadwal_id: string;
  siswa_id: string;
  nilai: number;
  status: "menunggu_koreksi" | "final";
  predikat: string | null;
  nisn: string;
  nama: string;
  kelas: string;
  jumlah_dijawab: number;
}

interface RekapData {
  items: NilaiRow[];
  statistik: {
    jumlah_peserta: number;
    rata_rata: number | null;
    tertinggi: number | null;
    terendah: number | null;
    jumlah_lulus: number;
    jumlah_menunggu_koreksi: number;
  };
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

export default function RekapPage() {
  const queryClient = useQueryClient();
  // Tombol aksi disembunyikan bila role tidak punya izinnya (menu Rekap
  // tampil untuk rekap.read, tapi Export/Finalisasi/Remedial punya izin
  // sendiri — menampilkan tombol yang pasti 403 adalah bug).
  const { hasPermission } = useAuthStore();
  const [jadwalId, setJadwalId] = useState("");
  const [jadwalInput, setJadwalInput] = useState("");
  const [remedialNilai, setRemedialNilai] = useState<Record<string, string>>({});
  const [pesan, setPesan] = useState("");

  // Daftar jadwal untuk dropdown — endpoint yang sama dipakai JadwalPage.
  const jadwalQuery = useQuery({
    queryKey: ["jadwal"],
    queryFn: () => apiClient.get<{ items: JadwalItem[] }>("/api/v1/jadwal"),
  });
  const jadwalList = jadwalQuery.data?.items ?? [];

  const rekapQuery = useQuery({
    queryKey: ["rekap", jadwalId],
    queryFn: () => apiClient.get<RekapData>(`/api/v1/rekap?jadwal_id=${encodeURIComponent(jadwalId)}`),
    enabled: jadwalId !== "",
  });

  const finalisasiMutation = useMutation({
    mutationFn: () => apiClient.post<{ jumlah_difinalisasi: number }>(`/api/v1/rekap/${jadwalId}/finalisasi`),
    onSuccess: (data) => {
      setPesan(`${data.jumlah_difinalisasi} nilai difinalisasi.`);
      void queryClient.invalidateQueries({ queryKey: ["rekap", jadwalId] });
    },
    onError: (err: Error) => setPesan(err.message),
  });

  const remedialMutation = useMutation({
    mutationFn: ({ siswaId, nilaiBaru }: { siswaId: string; nilaiBaru: number }) =>
      apiClient.post(`/api/v1/rekap/${jadwalId}/remedial`, { siswa_id: siswaId, nilai_baru: nilaiBaru }),
    onSuccess: () => {
      setPesan("Nilai remedial disimpan. Nilai akhir memakai yang tertinggi.");
      void queryClient.invalidateQueries({ queryKey: ["rekap", jadwalId] });
    },
    onError: (err: Error) => setPesan(err.message),
  });

  useEffect(() => {
    setPesan("");
  }, [jadwalId]);

  const rows = useMemo(() => rekapQuery.data?.items ?? [], [rekapQuery.data]);
  const stats = rekapQuery.data?.statistik;

  const handleCari = () => {
    if (!jadwalInput.trim()) return;
    setJadwalId(jadwalInput.trim());
  };

  const handleRemedial = (siswaId: string) => {
    const raw = remedialNilai[siswaId];
    const nilaiBaru = raw === undefined ? NaN : Number(raw);
    if (Number.isNaN(nilaiBaru) || nilaiBaru < 0 || nilaiBaru > 100) {
      setPesan("Nilai remedial harus angka 0-100.");
      return;
    }
    remedialMutation.mutate({ siswaId, nilaiBaru });
  };

  return (
    <div className="space-y-4">
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: "Rekap" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>Rekap Nilai</Typography.Title>
          <Typography.Text type="secondary">Nilai per jadwal beserta statistik. Nilai final terkunci; remedial tidak menimpa nilai awal.</Typography.Text>
        </div>
      </div>
      {pesan !== "" && (
        <Alert type="info" showIcon message={<>
          {pesan}
        </>} />
      )}

      <Form layout="vertical" onFinish={handleCari} className="flex flex-col gap-2 sm:flex-row sm:items-end">
        <Form.Item label="Jadwal" className="mb-0 w-full sm:max-w-sm">
          <Select showSearch options={[
              { value: "", label: "— Pilih jadwal —" },
              ...jadwalList.map((j) => ({ value: j.id, label: labelJadwal(j) })),
            ]} value={jadwalInput} onChange={(v) => setJadwalInput(v)} placeholder="— Pilih jadwal —" id="jadwal-id" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" />
        </Form.Item>
        <Button htmlType="submit" type="primary">
          Tampilkan
        </Button>
      </Form>
      {jadwalQuery.isError && (
        <Alert type="error" showIcon message="Gagal memuat daftar jadwal." />
      )}

      {jadwalId !== "" && stats && (
        <div className="grid grid-cols-2 gap-3 lg:grid-cols-5">
          {[
            { label: "Peserta", value: String(stats.jumlah_peserta) },
            { label: "Rata-rata", value: fmtNilai(stats.rata_rata) },
            { label: "Tertinggi", value: fmtNilai(stats.tertinggi) },
            { label: "Terendah", value: fmtNilai(stats.terendah) },
            { label: "Lulus (>=70)", value: String(stats.jumlah_lulus) },
          ].map((s) => (
            <Card key={s.label}>
              <Statistic title={s.label} value={s.value} />
            </Card>
          ))}
          {stats.jumlah_menunggu_koreksi > 0 && (
            <Typography.Text type="warning" className="col-span-2 text-sm lg:col-span-5">
              {stats.jumlah_menunggu_koreksi} nilai masih menunggu koreksi.
            </Typography.Text>
          )}
        </div>
      )}

      {jadwalId !== "" && (
        <div className="flex flex-wrap gap-2">
          {hasPermission("rekap.export") && (
            <Button onClick={() => apiClient.download(`/api/v1/rekap/${jadwalId}/export.csv`, `rekap-${jadwalId}.csv`)} type="default">
              <Download size={16} /> Export CSV
            </Button>
          )}
          {hasPermission("rekap.finalisasi") && (
            <Button onClick={() => finalisasiMutation.mutate()} loading={finalisasiMutation.isPending} type="primary">
              <Lock size={16} /> Finalisasi Nilai
            </Button>
          )}
          <Button onClick={() => void rekapQuery.refetch()} type="default">
            <RefreshCw size={16} /> Muat Ulang
          </Button>
          <Link to={`/adm/rekap/${jadwalId}/berita-acara`}><Button icon={<FileText size={16} />}>
            Berita Acara
          </Button></Link>
        </div>
      )}

      {rekapQuery.isPending && jadwalId !== "" && <p className="text-sm text-muted">Memuat rekap...</p>}
      {rekapQuery.isError && (
        <Alert type="error" showIcon message={`Gagal memuat rekap: ${(rekapQuery.error as Error).message}`} />
      )}

      {jadwalId !== "" && rekapQuery.isSuccess && (
          <Table columns={[{ key: "col0", title: "NISN", dataIndex: "nisn", sorter: (a: any, b: any) => { const _va = a.nisn, _vb = b.nisn; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, r: any) => (<>{r.nisn}</>) }, { key: "col1", title: "Nama", dataIndex: "nama", sorter: (a: any, b: any) => { const _va = a.nama, _vb = b.nama; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, r: any) => (<>{r.nama}</>) }, { key: "col2", title: "Kelas", render: (_: unknown, r: any) => (<>{r.kelas || "-"}</>) }, { key: "col3", title: "Dijawab", dataIndex: "jumlah_dijawab", sorter: (a: any, b: any) => { const _va = a.jumlah_dijawab, _vb = b.jumlah_dijawab; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, r: any) => (<>{r.jumlah_dijawab}</>) }, { key: "col4", title: "Nilai", render: (_: unknown, r: any) => (<>{r.nilai.toFixed(2)}</>) }, { key: "col5", title: "Predikat", dataIndex: "predikat", sorter: (a: any, b: any) => { const _va = a.predikat, _vb = b.predikat; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, r: any) => (<>{r.predikat ?? "-"}</>) }, { key: "col6", title: "Status", render: (_: unknown, r: any) => (<>
                    <Tag color={r.status === "final" ? "green" : "gold"}>
                      {r.status === "final" ? "Final" : "Menunggu koreksi"}
                    </Tag>
                  </>) }, ...(hasPermission("rekap.remedial") ? [{ key: "col7", title: "Remedial (nilai baru)", render: (_: unknown, r: any) => (<>
                    <div className="flex items-center gap-2">
                      <label className="sr-only" htmlFor={`remedial-${r.siswa_id}`}>
                        Nilai remedial {r.nama}
                      </label>
                      <InputNumber id={`remedial-${r.siswa_id}`} min={0} max={100} step={0.01} value={remedialNilai[r.siswa_id] === undefined || remedialNilai[r.siswa_id] === "" ? null : Number(remedialNilai[r.siswa_id])} onChange={(v) =>
                          setRemedialNilai((prev) => ({ ...prev, [r.siswa_id]: v === null ? "" : String(v) }))
                        } placeholder="0-100" size="large" style={{ width: "100%" }} />
                      <Button onClick={() => handleRemedial(r.siswa_id)} loading={remedialMutation.isPending} type="primary">
                        Simpan
                      </Button>
                    </div>
                  </>) }] : [])]} dataSource={rows} rowKey="id" pagination={false} locale={{ emptyText: "Belum ada nilai untuk jadwal ini." }} scroll={{ x: "max-content" }} />
      )}
    </div>
  );
}
