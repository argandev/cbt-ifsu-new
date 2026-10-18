// RuanganPage — CRUD Ruangan (ruang fisik ujian: nama, kapasitas,
// lokasi/gedung, keterangan). Milik Tim Fase 2 (Event + Ruangan +
// Kartu Peserta). Izin ditegakkan backend (event.read / create / update
// / delete). TanStack Query + apiClient + komponen ui Tim 1. NOL emoji.
import { useState } from "react";
import { Link } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Pencil, Plus, Trash2 } from "lucide-react";
import { apiClient } from "../../../lib/api";

import { Alert, Button, Input, InputNumber, Modal, Table, Breadcrumb, Space, Typography, Form } from "antd";
type RuanganItem = {
  id: string;
  kode: string;
  nama: string;
  kapasitas: number;
  lokasi: string | null;
  keterangan: string | null;
  created_at: string;
};

type ListRes = { items: RuanganItem[]; total: number; limit: number; offset: number };

const formKosong = { kode: "", nama: "", kapasitas: "", lokasi: "", keterangan: "" };

function apiErrorMessage(err: unknown): string {
  if (typeof err === "object" && err !== null) {
    const e = err as { message?: string; error?: { message?: string } };
    if (e.error?.message) return e.error.message;
    if (e.message) return e.message;
  }
  return "Terjadi kesalahan. Coba lagi.";
}

export default function RuanganPage() {
  const qc = useQueryClient();
  const [q, setQ] = useState("");
  const [form, setForm] = useState(formKosong);
  const [editing, setEditing] = useState<RuanganItem | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [formError, setFormError] = useState("");
  const [pageError, setPageError] = useState("");

  const listQuery = useQuery({
    queryKey: ["ruangan", q],
    queryFn: () =>
      apiClient.get<ListRes>(`/api/v1/ruangan?limit=100${q ? `&q=${encodeURIComponent(q)}` : ""}`),
  });
  const items = listQuery.data?.items ?? [];

  const invalidate = () => qc.invalidateQueries({ queryKey: ["ruangan"] });

  const saveMutation = useMutation({
    mutationFn: async () => {
      const body = {
        kode: form.kode.trim(),
        nama: form.nama.trim(),
        kapasitas: Number(form.kapasitas),
        lokasi: form.lokasi.trim() || null,
        keterangan: form.keterangan.trim() || null,
      };
      if (editing) return apiClient.put(`/api/v1/ruangan/${editing.id}`, body);
      return apiClient.post("/api/v1/ruangan", body);
    },
    onSuccess: () => {
      setDialogOpen(false);
      setFormError("");
      invalidate();
    },
    onError: (err) => setFormError(apiErrorMessage(err)),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => apiClient.del(`/api/v1/ruangan/${id}`),
    onSuccess: () => {
      setPageError("");
      invalidate();
    },
    onError: (err) => setPageError(apiErrorMessage(err)),
  });

  function openAdd() {
    setEditing(null);
    setForm(formKosong);
    setFormError("");
    setDialogOpen(true);
  }

  function openEdit(item: RuanganItem) {
    setEditing(item);
    setForm({
      kode: item.kode,
      nama: item.nama,
      kapasitas: String(item.kapasitas),
      lokasi: item.lokasi ?? "",
      keterangan: item.keterangan ?? "",
    });
    setFormError("");
    setDialogOpen(true);
  }

  function submit() {
    if (!form.kode.trim() || !form.nama.trim()) {
      setFormError("Kode dan nama ruangan wajib diisi.");
      return;
    }
    const kap = Number(form.kapasitas);
    if (!Number.isInteger(kap) || kap <= 0) {
      setFormError("Kapasitas harus bilangan bulat lebih dari 0.");
      return;
    }
    saveMutation.mutate();
  }

  function konfirmasiHapus(item: RuanganItem) {
    Modal.confirm({
      title: "Hapus Ruangan?",
      content: (
        <p className="text-sm text-muted">
          Ruangan <strong>{item.nama}</strong> ({item.kode}) akan dihapus permanen. Ruangan yang
          masih dipakai jadwal ujian tidak bisa dihapus.
        </p>
      ),
      okText: "Ya, Hapus",
      cancelText: "Batal",
      okButtonProps: { danger: true },
      onOk: () => deleteMutation.mutate(item.id),
    });
  }

  return (
    <div className="space-y-4">
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: "Ruangan" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>Ruangan</Typography.Title>
          <Typography.Text type="secondary">Ruang fisik ujian. Jadwal dapat ditugaskan ke satu ruangan; kapasitas ruangan harus cukup untuk siswa kelas jadwal.</Typography.Text>
        </div>
        <Space>
          <Button onClick={openAdd} type="primary">
            <Plus size={16} /> Tambah Ruangan
          </Button>
        </Space>
      </div>

      <div className="max-w-md">
        <div><Typography.Text strong className="mb-1 block">Cari ruangan</Typography.Text><Input.Search placeholder="Cari kode / nama / lokasi…" value={q} onChange={(e) => setQ(e.target.value)} aria-label="Cari ruangan" size="large" allowClear /></div>
      </div>

      {pageError && (
        <Alert type="error" showIcon message={pageError} />
      )}
      {listQuery.isError && (
        <Alert type="error" showIcon message={`Gagal memuat ruangan: ${apiErrorMessage(listQuery.error)}`} />
      )}

        <Table columns={[{ key: "col0", title: "Kode", dataIndex: "kode", sorter: (a: any, b: any) => { const _va = a.kode, _vb = b.kode; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, item: any) => (<>{item.kode}</>) }, { key: "col1", title: "Nama", dataIndex: "nama", sorter: (a: any, b: any) => { const _va = a.nama, _vb = b.nama; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, item: any) => (<>{item.nama}</>) }, { key: "col2", title: "Kapasitas", render: (_: unknown, item: any) => (<>{item.kapasitas} kursi</>) }, { key: "col3", title: "Lokasi / Gedung", dataIndex: "lokasi", sorter: (a: any, b: any) => { const _va = a.lokasi, _vb = b.lokasi; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, item: any) => (<>{item.lokasi ?? "-"}</>) }, { key: "col4", title: "Keterangan", dataIndex: "keterangan", sorter: (a: any, b: any) => { const _va = a.keterangan, _vb = b.keterangan; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, item: any) => (<>{item.keterangan ?? "-"}</>) }, { key: "col5", title: "Aksi", render: (_: unknown, item: any) => (<>
                  <div className="flex justify-end gap-2">
                    <Button onClick={() => openEdit(item)} type="default">
                      <Pencil size={14} /> Ubah
                    </Button>
                    <Button onClick={() => konfirmasiHapus(item)} type="default" danger>
                      <Trash2 size={14} /> Hapus
                    </Button>
                  </div>
                </>) }]} dataSource={items} rowKey="id" pagination={false} locale={{ emptyText: "Belum ada ruangan." }} scroll={{ x: "max-content" }} />
      <p className="text-sm text-muted">Total: {listQuery.data?.total ?? 0} ruangan.</p>

      <Modal open={dialogOpen} onCancel={() => setDialogOpen(false)} title={editing ? "Ubah Ruangan" : "Tambah Ruangan"} footer={null} width={720} destroyOnClose>
        <Form layout="vertical" onFinish={submit}>
          {formError && (
            <Alert type="error" showIcon message={formError} />
          )}
          <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
            <Form.Item label="Kode *"><Input placeholder="mis. R-01" value={form.kode} onChange={(e) => setForm({ ...form, kode: e.target.value })} size="large" /></Form.Item>
            <Form.Item label="Nama *"><Input placeholder="mis. Lab Komputer 1" value={form.nama} onChange={(e) => setForm({ ...form, nama: e.target.value })} size="large" /></Form.Item>
          </div>
          <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
            <Form.Item label="Kapasitas (kursi) *"><InputNumber min={1} step={1} placeholder="mis. 36" value={form.kapasitas === "" ? null : Number(form.kapasitas)} onChange={(v) => setForm({ ...form, kapasitas: v === null ? "" : String(v) })} size="large" style={{ width: "100%" }} /></Form.Item>
            <Form.Item label="Lokasi / Gedung"><Input placeholder="mis. Gedung A Lantai 2" value={form.lokasi} onChange={(e) => setForm({ ...form, lokasi: e.target.value })} size="large" /></Form.Item>
          </div>
          <Form.Item label="Keterangan"><Input.TextArea value={form.keterangan} onChange={(e) => setForm({ ...form, keterangan: e.target.value })} placeholder="Keterangan ruangan (opsional)" size="large"></Input.TextArea></Form.Item>
          <div className="flex flex-col gap-2 sm:flex-row sm:justify-end">
            <Button onClick={() => setDialogOpen(false)} type="default">
              Batal
            </Button>
            <Button htmlType="submit" loading={saveMutation.isPending} type="primary">
              {saveMutation.isPending ? "Menyimpan…" : "Simpan"}
            </Button>
          </div>
        </Form>
      </Modal>

    </div>
  );
}
