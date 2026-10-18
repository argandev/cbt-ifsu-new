// EventPage — CRUD Event (pengelompokan ujian: PTS, PAS, USBK, UKK,
// Tryout, Lainnya). Milik Tim Fase 2 (Event + Ruangan + Kartu Peserta).
// Izin ditegakkan backend (event.read / create / update / delete).
// TanStack Query + apiClient + komponen ui Tim 1. NOL emoji di UI.
import { useState } from "react";
import { Link } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Pencil, Plus, Trash2 } from "lucide-react";
import dayjs from "dayjs";
import { apiClient } from "../../../lib/api";
import { formatTanggal } from "../../../lib/format";

import { Alert, Button, DatePicker, Input, Modal, Select, Table, Tag, Breadcrumb, Space, Typography, Form } from "antd";
type EventItem = {
  id: string;
  kode: string;
  nama: string;
  jenis: string;
  tanggal_mulai: string | null;
  tanggal_selesai: string | null;
  deskripsi: string | null;
  created_at: string;
};

type ListRes = { items: EventItem[]; total: number; limit: number; offset: number };

const JENIS_OPTIONS = [
  { value: "PTS", label: "PTS" },
  { value: "PAS", label: "PAS" },
  { value: "USBK", label: "USBK" },
  { value: "UKK", label: "UKK" },
  { value: "Tryout", label: "Tryout" },
  { value: "Lainnya", label: "Lainnya" },
];

const formKosong = {
  kode: "",
  nama: "",
  jenis: "Lainnya",
  tanggal_mulai: "",
  tanggal_selesai: "",
  deskripsi: "",
};

function apiErrorMessage(err: unknown): string {
  if (typeof err === "object" && err !== null) {
    const e = err as { message?: string; error?: { message?: string } };
    if (e.error?.message) return e.error.message;
    if (e.message) return e.message;
  }
  return "Terjadi kesalahan. Coba lagi.";
}

export default function EventPage() {
  const qc = useQueryClient();
  const [q, setQ] = useState("");
  const [jenisFilter, setJenisFilter] = useState("");
  const [form, setForm] = useState(formKosong);
  const [editing, setEditing] = useState<EventItem | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [formError, setFormError] = useState("");
  const [pageError, setPageError] = useState("");

  const listQuery = useQuery({
    queryKey: ["event", q, jenisFilter],
    queryFn: () =>
      apiClient.get<ListRes>(
        `/api/v1/event?limit=100${jenisFilter ? `&jenis=${encodeURIComponent(jenisFilter)}` : ""}${q ? `&q=${encodeURIComponent(q)}` : ""}`,
      ),
  });
  const items = listQuery.data?.items ?? [];

  const invalidate = () => qc.invalidateQueries({ queryKey: ["event"] });

  const saveMutation = useMutation({
    mutationFn: async () => {
      const body = {
        kode: form.kode.trim(),
        nama: form.nama.trim(),
        jenis: form.jenis,
        tanggal_mulai: form.tanggal_mulai || null,
        tanggal_selesai: form.tanggal_selesai || null,
        deskripsi: form.deskripsi.trim() || null,
      };
      if (editing) return apiClient.put(`/api/v1/event/${editing.id}`, body);
      return apiClient.post("/api/v1/event", body);
    },
    onSuccess: () => {
      setDialogOpen(false);
      setFormError("");
      invalidate();
    },
    onError: (err) => setFormError(apiErrorMessage(err)),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => apiClient.del(`/api/v1/event/${id}`),
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

  function openEdit(item: EventItem) {
    setEditing(item);
    setForm({
      kode: item.kode,
      nama: item.nama,
      jenis: item.jenis,
      tanggal_mulai: (item.tanggal_mulai ?? "").slice(0, 10),
      tanggal_selesai: (item.tanggal_selesai ?? "").slice(0, 10),
      deskripsi: item.deskripsi ?? "",
    });
    setFormError("");
    setDialogOpen(true);
  }

  function submit() {
    if (!form.kode.trim() || !form.nama.trim()) {
      setFormError("Kode dan nama event wajib diisi.");
      return;
    }
    if (form.tanggal_mulai && form.tanggal_selesai && form.tanggal_selesai < form.tanggal_mulai) {
      setFormError("Tanggal selesai tidak boleh sebelum tanggal mulai.");
      return;
    }
    saveMutation.mutate();
  }

  function konfirmasiHapus(item: EventItem) {
    Modal.confirm({
      title: "Hapus Event?",
      content: (
        <p className="text-sm text-muted">
          Event <strong>{item.nama}</strong> ({item.kode}) akan dihapus permanen. Event yang masih
          dipakai jadwal ujian tidak bisa dihapus.
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
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: "Event Ujian" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>Event Ujian</Typography.Title>
          <Typography.Text type="secondary">Pengelompokan ujian (mis. PTS Ganjil 2026/2027). Jadwal ujian dapat merujuk ke satu event.</Typography.Text>
        </div>
        <Space>
          <Button onClick={openAdd} type="primary">
            <Plus size={16} /> Tambah Event
          </Button>
        </Space>
      </div>

      <div className="flex flex-col gap-2 sm:flex-row sm:items-end">
        <div className="flex-1">
          <div><Typography.Text strong className="mb-1 block">Cari event</Typography.Text><Input.Search placeholder="Cari kode / nama event…" value={q} onChange={(e) => setQ(e.target.value)} aria-label="Cari event" size="large" allowClear /></div>
        </div>
        <div className="w-full sm:w-48">
          <div><Typography.Text strong className="mb-1 block">Jenis</Typography.Text><Select showSearch options={[{ value: "", label: "Semua jenis" }, ...JENIS_OPTIONS]} value={jenisFilter} onChange={(v) => setJenisFilter((v ?? ""))} placeholder="Pilih…" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" /></div>
        </div>
      </div>

      {pageError && (
        <Alert type="error" showIcon message={pageError} />
      )}
      {listQuery.isError && (
        <Alert type="error" showIcon message={`Gagal memuat event: ${apiErrorMessage(listQuery.error)}`} />
      )}

        <Table columns={[{ key: "col0", title: "Kode", dataIndex: "kode", sorter: (a: any, b: any) => { const _va = a.kode, _vb = b.kode; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, item: any) => (<>{item.kode}</>) }, { key: "col1", title: "Nama", dataIndex: "nama", sorter: (a: any, b: any) => { const _va = a.nama, _vb = b.nama; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, item: any) => (<>{item.nama}</>) }, { key: "col2", title: "Jenis", render: (_: unknown, item: any) => (<>
                  <Tag color="blue">{item.jenis}</Tag>
                </>) }, { key: "col3", title: "Tanggal Mulai", render: (_: unknown, item: any) => (<>{formatTanggal(item.tanggal_mulai)}</>) }, { key: "col4", title: "Tanggal Selesai", render: (_: unknown, item: any) => (<>{formatTanggal(item.tanggal_selesai)}</>) }, { key: "col5", title: "Deskripsi", dataIndex: "deskripsi", sorter: (a: any, b: any) => { const _va = a.deskripsi, _vb = b.deskripsi; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, item: any) => (<>{item.deskripsi ?? "-"}</>) }, { key: "col6", title: "Aksi", render: (_: unknown, item: any) => (<>
                  <div className="flex justify-end gap-2">
                    <Button onClick={() => openEdit(item)} type="default">
                      <Pencil size={14} /> Ubah
                    </Button>
                    <Button onClick={() => konfirmasiHapus(item)} type="default" danger>
                      <Trash2 size={14} /> Hapus
                    </Button>
                  </div>
                </>) }]} dataSource={items} rowKey="id" pagination={false} locale={{ emptyText: "Belum ada event." }} scroll={{ x: "max-content" }} />
      <p className="text-sm text-muted">Total: {listQuery.data?.total ?? 0} event.</p>

      <Modal open={dialogOpen} onCancel={() => setDialogOpen(false)} title={editing ? "Ubah Event" : "Tambah Event"} footer={null} width={720} destroyOnClose>
        <Form layout="vertical" onFinish={submit}>
          {formError && (
            <Alert type="error" showIcon message={formError} />
          )}
          <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
            <Form.Item label="Kode *"><Input placeholder="mis. PTS-GANJIL-26" value={form.kode} onChange={(e) => setForm({ ...form, kode: e.target.value })} size="large" /></Form.Item>
            <Form.Item label="Nama *"><Input placeholder="mis. PTS Ganjil 2026/2027" value={form.nama} onChange={(e) => setForm({ ...form, nama: e.target.value })} size="large" /></Form.Item>
          </div>
          <Form.Item label="Jenis"><Select showSearch options={JENIS_OPTIONS} value={form.jenis} onChange={(v) => setForm({ ...form, jenis: (v ?? "") })} placeholder="Pilih…" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" /></Form.Item>
          <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
            <Form.Item label="Tanggal mulai"><DatePicker format="YYYY-MM-DD" value={form.tanggal_mulai ? dayjs(form.tanggal_mulai, "YYYY-MM-DD") : null} onChange={(d) => setForm({ ...form, tanggal_mulai: d ? d.format("YYYY-MM-DD") : "" })} size="large" style={{ width: "100%" }} /></Form.Item>
            <Form.Item label="Tanggal selesai"><DatePicker format="YYYY-MM-DD" value={form.tanggal_selesai ? dayjs(form.tanggal_selesai, "YYYY-MM-DD") : null} onChange={(d) => setForm({ ...form, tanggal_selesai: d ? d.format("YYYY-MM-DD") : "" })} size="large" style={{ width: "100%" }} /></Form.Item>
          </div>
          <Form.Item label="Deskripsi"><Input.TextArea value={form.deskripsi} onChange={(e) => setForm({ ...form, deskripsi: e.target.value })} placeholder="Keterangan event (opsional)" size="large"></Input.TextArea></Form.Item>
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
