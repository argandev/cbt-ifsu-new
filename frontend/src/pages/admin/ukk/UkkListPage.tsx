import { useState } from "react";
import { Link } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Eye, Plus, Trash2 } from "lucide-react";
import dayjs from "dayjs";
import apiClient from "../../../lib/api";
import type { MasterOption, Ukk } from "./types";
import { fmtTanggal } from "./types";

import { Alert, Button, Card, DatePicker, Input, Modal, Select, Table, Tag, Breadcrumb, Space, Typography, Form } from "antd";
import { statusTagColor } from "../../../components/common/statusTag";

export default function UkkListPage() {
  const queryClient = useQueryClient();
  const [statusFilter, setStatusFilter] = useState("");
  const [dialogOpen, setDialogOpen] = useState(false);
  const [pesan, setPesan] = useState("");
  const [form, setForm] = useState({
    nama: "",
    mapel_id: "",
    kelas_id: "",
    jurusan_id: "",
    tanggal: "",
    penguji_internal: "",
    penguji_eksternal: "",
    status: "draft",
  });

  const listQuery = useQuery({
    queryKey: ["ukk", statusFilter],
    queryFn: () =>
      apiClient.get<Ukk[]>(
        `/ukk?limit=100${statusFilter ? `&status=${encodeURIComponent(statusFilter)}` : ""}`,
      ),
  });

  const jurusanQuery = useQuery({
    queryKey: ["ukk-master-jurusan"],
    queryFn: () => apiClient.get<MasterOption[]>("/master/jurusan?per_page=100"),
    enabled: dialogOpen,
  });
  const kelasQuery = useQuery({
    queryKey: ["ukk-master-kelas", form.jurusan_id],
    queryFn: () =>
      apiClient.get<MasterOption[]>(
        `/master/kelas?per_page=100${form.jurusan_id ? `&jurusan_id=${form.jurusan_id}` : ""}`,
      ),
    enabled: dialogOpen,
  });
  const mapelQuery = useQuery({
    queryKey: ["ukk-master-mapel"],
    queryFn: () => apiClient.get<MasterOption[]>("/master/mapel?per_page=100"),
    enabled: dialogOpen,
  });

  const createMutation = useMutation({
    mutationFn: (payload: Record<string, unknown>) => apiClient.post<Ukk>("/ukk", payload),
    onSuccess: () => {
      setPesan("UKK berhasil dibuat.");
      setDialogOpen(false);
      setForm({
        nama: "",
        mapel_id: "",
        kelas_id: "",
        jurusan_id: "",
        tanggal: "",
        penguji_internal: "",
        penguji_eksternal: "",
        status: "draft",
      });
      void queryClient.invalidateQueries({ queryKey: ["ukk"] });
    },
    onError: (err: Error) => setPesan(err.message),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => apiClient.delete(`/ukk/${id}`),
    onSuccess: () => {
      setPesan("UKK berhasil dihapus.");
      void queryClient.invalidateQueries({ queryKey: ["ukk"] });
    },
    onError: (err: Error) => setPesan(err.message),
  });

  const setField = (key: keyof typeof form, value: string) =>
    setForm((prev) => ({ ...prev, [key]: value }));

  const handleSubmit = () => {
    setPesan("");
    // Pengganti `required` Select lama (hidden input native): antd Select tidak
    // merender input required, jadi blokir di handler dengan pesan yang sama.
    if (!form.jurusan_id || !form.kelas_id || !form.mapel_id) {
      setPesan("Jurusan, kelas, dan mapel wajib dipilih.");
      return;
    }
    createMutation.mutate({
      nama: form.nama.trim(),
      mapel_id: form.mapel_id,
      kelas_id: form.kelas_id,
      jurusan_id: form.jurusan_id,
      tanggal: form.tanggal || null,
      penguji_internal: form.penguji_internal.trim() || null,
      penguji_eksternal: form.penguji_eksternal.trim() || null,
      status: form.status,
    });
  };

  const rows = listQuery.data ?? [];

  function konfirmasiHapusUkk(u: Ukk) {
    Modal.confirm({
      title: "Hapus UKK?",
      content: `UKK "${u.nama}" akan dihapus. Hanya bisa dihapus bila belum ada nilai/berkas/sertifikat.`,
      okText: "Ya, Hapus",
      cancelText: "Batal",
      okButtonProps: { danger: true },
      onOk: () => deleteMutation.mutate(u.id),
    });
  }

  return (
    <div className="flex flex-col gap-4">
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: "UKK" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>UKK (Uji Kompetensi Keahlian)</Typography.Title>
          <Typography.Text type="secondary">Bobot nilai akhir: teori 40% + praktik 60%. Rubrik praktik: Proses 30% / Produk 50% / Sikap 20%. Lulus ≥ 70. Sertifikat hanya bagi yang lulus.</Typography.Text>
        </div>
        <Space>
          <Button onClick={() => setDialogOpen(true)} type="primary">
            <Plus size={16} /> Buat UKK
          </Button>
        </Space>
      </div>

      {pesan ? <Alert type="info" showIcon message={pesan} /> : null}
      {listQuery.isError ? (
        <Alert type="error" showIcon message={`Gagal memuat daftar UKK: ${(listQuery.error as Error).message}`} />
      ) : null}

      <Card title="Daftar UKK">
        
        <div>
          <div className="mb-4 max-w-xs">
            <div><Typography.Text strong className="mb-1 block">Filter status</Typography.Text><Select showSearch options={[
                { value: "", label: "Semua status" },
                { value: "draft", label: "Draft" },
                { value: "berjalan", label: "Berjalan" },
                { value: "selesai", label: "Selesai" },
              ]} value={statusFilter} onChange={(v) => setStatusFilter((v ?? ""))} placeholder="Pilih…" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" /></div>
          </div>
            <Table columns={[{ key: "col0", title: "Nama UKK", dataIndex: "nama", sorter: (a: any, b: any) => { const _va = a.nama, _vb = b.nama; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, u: any) => (<>{u.nama}</>) }, { key: "col1", title: "Mapel", dataIndex: "mapel_nama", sorter: (a: any, b: any) => { const _va = a.mapel_nama, _vb = b.mapel_nama; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, u: any) => (<>{u.mapel_nama ?? "-"}</>) }, { key: "col2", title: "Kelas", dataIndex: "kelas_nama", sorter: (a: any, b: any) => { const _va = a.kelas_nama, _vb = b.kelas_nama; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, u: any) => (<>{u.kelas_nama ?? "-"}</>) }, { key: "col3", title: "Jurusan", dataIndex: "jurusan_nama", sorter: (a: any, b: any) => { const _va = a.jurusan_nama, _vb = b.jurusan_nama; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, u: any) => (<>{u.jurusan_nama ?? u.jurusan_kode ?? "-"}</>) }, { key: "col4", title: "Tanggal", render: (_: unknown, u: any) => (<>{fmtTanggal(u.tanggal)}</>) }, { key: "col5", title: "Status", render: (_: unknown, u: any) => (<>
                      <Tag color={statusTagColor(u.status)}>{u.status}</Tag>
                    </>) }, { key: "col6", title: "Peserta (dinilai)", dataIndex: "jumlah_peserta", sorter: (a: any, b: any) => { const _va = a.jumlah_peserta, _vb = b.jumlah_peserta; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, u: any) => (<>{u.jumlah_peserta ?? 0}</>) }, { key: "col7", title: "Aksi", render: (_: unknown, u: any) => (<>
                      <div className="flex items-center gap-2">
                        <Link
                          to={`/adm/ukk/${u.id}`}
                          className="inline-flex min-h-[44px] items-center gap-1 rounded-lg px-2 text-sm font-medium text-primary hover:bg-surface"
                        >
                          <Eye size={16} /> Detail
                        </Link>
                        <Button aria-label={`Hapus UKK ${u.nama}`} onClick={() => konfirmasiHapusUkk(u)} type="default" danger>
                          <Trash2 size={16} />
                        </Button>
                      </div>
                    </>) }]} dataSource={rows} rowKey="id" pagination={false} locale={{ emptyText: "Belum ada UKK." }} scroll={{ x: "max-content" }} />
          {listQuery.isLoading ? <p className="mt-3 text-sm text-muted">Memuat...</p> : null}
        </div>
      </Card>

      <Modal open={dialogOpen} onCancel={() => setDialogOpen(false)} title="Buat UKK Baru" footer={null} width={720} destroyOnClose>
        <Form layout="vertical" onFinish={handleSubmit}>
          <Form.Item label="Nama UKK"><Input value={form.nama} onChange={(e) => setField("nama", e.target.value)} placeholder="mis. UKK Rekayasa Perangkat Lunak 2026" required size="large" /></Form.Item>
          <Form.Item label="Jurusan"><Select showSearch allowClear options={(jurusanQuery.data ?? []).map((j) => ({
              value: j.id,
              label: j.kode ? `${j.kode} — ${j.nama}` : j.nama,
            }))} value={form.jurusan_id} onChange={(v) => {
              setField("jurusan_id", (v ?? ""));
              setField("kelas_id", "");
            }} placeholder="Pilih jurusan" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" /></Form.Item>
          <Form.Item label="Kelas"><Select showSearch allowClear options={(kelasQuery.data ?? []).map((k) => ({ value: k.id, label: k.nama }))} value={form.kelas_id} onChange={(v) => setField("kelas_id", (v ?? ""))} placeholder="Pilih kelas" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" /></Form.Item>
          <Form.Item label="Mapel"><Select showSearch allowClear options={(mapelQuery.data ?? []).map((m) => ({
              value: m.id,
              label: m.kode ? `${m.kode} — ${m.nama}` : m.nama,
            }))} value={form.mapel_id} onChange={(v) => setField("mapel_id", (v ?? ""))} placeholder="Pilih mapel" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" /></Form.Item>
          <Form.Item label="Tanggal"><DatePicker format="YYYY-MM-DD" value={form.tanggal ? dayjs(form.tanggal, "YYYY-MM-DD") : null} onChange={(d) => setField("tanggal", d ? d.format("YYYY-MM-DD") : "")} size="large" style={{ width: "100%" }} /></Form.Item>
          <Form.Item label="Penguji internal"><Input value={form.penguji_internal} onChange={(e) => setField("penguji_internal", e.target.value)} placeholder="Nama guru penguji internal" size="large" /></Form.Item>
          <Form.Item label="Penguji eksternal (DUDIKA)"><Input value={form.penguji_eksternal} onChange={(e) => setField("penguji_eksternal", e.target.value)} placeholder="Nama penguji eksternal / asal DUDIKA" size="large" /></Form.Item>
          <Form.Item label="Status awal"><Select showSearch options={[
              { value: "draft", label: "Draft" },
              { value: "berjalan", label: "Berjalan" },
              { value: "selesai", label: "Selesai" },
            ]} value={form.status} onChange={(v) => setField("status", (v ?? ""))} placeholder="Pilih…" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" /></Form.Item>
          <Button htmlType="submit" loading={createMutation.isPending} type="primary">
            {createMutation.isPending ? "Menyimpan..." : "Simpan UKK"}
          </Button>
        </Form>
      </Modal>

    </div>
  );
}
