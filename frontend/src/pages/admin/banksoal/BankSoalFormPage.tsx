import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link, useNavigate, useParams } from "react-router-dom";
import { ArrowLeft } from "lucide-react";
import apiClient from "../../../lib/api";
import { useAuthStore } from "../../../store/auth";
import type { BankSoal } from "../../../lib/banksoal";

import { Alert, Button, Input, Select, Form, Card, Breadcrumb, Typography } from "antd";
type Mapel = { id: string; kode: string; nama: string };



export default function BankSoalFormPage() {
  const { hasPermission } = useAuthStore();
  const { id } = useParams();
  // Route "bank-soal/baru" tidak punya param :id (id === undefined) -> mode buat baru.
  // Route "bank-soal/:id/ubah" membawa id bank soal -> mode edit.
  const isBaru = id === "baru" || id === undefined;
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [pesan, setPesan] = useState("");

  const [kode, setKode] = useState("");
  const [nama, setNama] = useState("");
  const [mapelId, setMapelId] = useState("");
  const [kelasTingkat, setKelasTingkat] = useState(10);

  const detailQuery = useQuery({
    queryKey: ["bank-soal", id],
    queryFn: () => apiClient.get<BankSoal>(`/api/v1/bank-soal/${id}`),
    enabled: !isBaru && !!id,
  });

  // Daftar mapel untuk dropdown — endpoint yang sama dipakai GuruPage.
  const mapelQuery = useQuery({
    queryKey: ["master", "mapel", "all"],
    queryFn: () => apiClient.get<{ data: Mapel[] } | Mapel[]>("/api/v1/master/mapel?per_page=100"),
  });
  const mapelList: Mapel[] = Array.isArray(mapelQuery.data)
    ? mapelQuery.data
    : (mapelQuery.data?.data ?? []);

  useEffect(() => {
    const b = detailQuery.data;
    if (b) {
      setKode(b.kode);
      setNama(b.nama);
      setMapelId(b.mapel_id);
      setKelasTingkat(b.kelas_tingkat);
    }
  }, [detailQuery.data]);

  const simpanMutation = useMutation({
    mutationFn: () => {
      const body = { kode, nama, mapel_id: mapelId, kelas_tingkat: kelasTingkat };
      return isBaru
        ? apiClient.post<BankSoal>("/api/v1/bank-soal", body)
        : apiClient.put<BankSoal>(`/api/v1/bank-soal/${id}`, body);
    },
    onSuccess: (data) => {
      void queryClient.invalidateQueries({ queryKey: ["bank-soal"] });
      navigate(`/adm/bank-soal/${data.id}`);
    },
    onError: (err: Error) => setPesan(err.message),
  });

  return (
    <div className="space-y-6">
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: (<Link to="/adm/bank-soal">Bank Soal</Link>) }, { title: isBaru ? "Buat Bank Soal" : "Ubah Bank Soal" }]} />
      <div>
        <Link to="/adm/bank-soal" className="inline-flex items-center gap-1 text-sm text-muted hover:text-anbk">
          <ArrowLeft size={16} /> Kembali ke daftar bank soal
        </Link>
        <Typography.Title level={3} style={{ margin: "8px 0 0" }}>{isBaru ? "Buat Bank Soal" : "Ubah Bank Soal"}</Typography.Title>
        <Typography.Text type="secondary">Satu bank soal = satu mata pelajaran + satu tingkat kelas.</Typography.Text>
      </div>

      {detailQuery.isLoading && <p className="text-sm text-muted">Memuat…</p>}
      {pesan && <Alert type="error" showIcon message={<>{pesan}</>} />}

      <Card className="max-w-xl">
      <Form
        layout="vertical"
        onFinish={() => {
          setPesan("");
          // Cerminan `required` native pada Kode/Nama (sebelumnya divalidasi browser).
          if (!kode.trim() || !nama.trim()) {
            setPesan("Kode bank dan nama bank wajib diisi.");
            return;
          }
          // Pengganti `required` Dropdown lama (hidden input native): antd Select
          // tidak merender input required, jadi blokir di handler dengan pesan sama.
          if (!mapelId) {
            setPesan("Mata pelajaran wajib dipilih.");
            return;
          }
          simpanMutation.mutate();
        }}
      >
        <Form.Item label="Kode bank">
          <Input id="kode" value={kode} onChange={(e) => setKode(e.target.value)} required placeholder="mis. MTK-XI-01" size="large" />
        </Form.Item>
        <Form.Item label="Nama bank">
          <Input id="nama" value={nama} onChange={(e) => setNama(e.target.value)} required placeholder="mis. Bank Soal Matematika XI" size="large" />
        </Form.Item>
        <Form.Item label="Mata pelajaran">
          <Select showSearch options={[
              { value: "", label: "— Pilih mata pelajaran —" },
              ...mapelList.map((m) => ({ value: m.id, label: `${m.nama} (${m.kode})` })),
            ]} value={mapelId} onChange={(v) => setMapelId(v)} placeholder="— Pilih mata pelajaran —" id="mapel" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" />
          {mapelQuery.isError && (
            <Alert type="error" showIcon message="Gagal memuat daftar mapel." className="mt-1" />
          )}
        </Form.Item>
        <Form.Item label="Tingkat kelas">
          <Select options={[
              { value: "10", label: "Kelas 10" },
              { value: "11", label: "Kelas 11" },
              { value: "12", label: "Kelas 12" },
            ]} value={String(kelasTingkat)} onChange={(v) => setKelasTingkat(Number(v))} id="kelas" placeholder="Pilih…" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" />
        </Form.Item>
        <div className="flex flex-wrap gap-2">
          {hasPermission(isBaru ? "banksoal.create" : "banksoal.update") && (
            <Button htmlType="submit" loading={simpanMutation.isPending} type="primary">
              {simpanMutation.isPending ? "Menyimpan…" : "Simpan"}
            </Button>
          )}
          <Button href="/adm/bank-soal">Batal</Button>
        </div>
      </Form>
      </Card>
    </div>
  );
}
