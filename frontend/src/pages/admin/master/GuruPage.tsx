// GuruPage — Master Data: Guru.
// Milik Tim 2. Lihat catatan kontrak apiClient di JurusanPage.tsx.
// Mapel yang diampu disimpan sebagai array ID (kolom mapel_ids JSONB).
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "../../../lib/api";
import { useAuthStore } from "../../../store/auth";

import { Alert, Button, Checkbox, Input, Modal, Pagination, Table, Breadcrumb, Space, Typography, Form } from "antd";
type Guru = { id: string; user_id: string | null; nip: string | null; nama: string; mapel_ids: string[] };
type Mapel = { id: string; kode: string; nama: string };
type Meta = { page: number; per_page: number; total: number; total_pages: number };
type ListRes<T> = { data: T[]; meta?: Meta };

function unwrapList<T>(res: ListRes<T> | T[]): { items: T[]; meta?: Meta } {
  if (Array.isArray(res)) return { items: res };
  return { items: res.data ?? [], meta: res.meta };
}
function apiErrorMessage(err: unknown): string {
  if (typeof err === "object" && err !== null) {
    const e = err as { message?: string; error?: { message?: string } };
    if (e.error?.message) return e.error.message;
    if (e.message) return e.message;
  }
  return "Terjadi kesalahan. Coba lagi.";
}





export default function GuruPage() {
  const { hasPermission } = useAuthStore();
  const qc = useQueryClient();
  const [q, setQ] = useState("");
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Guru | null>(null);
  const [nip, setNip] = useState("");
  const [nama, setNama] = useState("");
  const [userId, setUserId] = useState("");
  const [mapelIds, setMapelIds] = useState<string[]>([]);
  const [formError, setFormError] = useState("");
  const [pageError, setPageError] = useState("");

  useEffect(() => {
    const t = setTimeout(() => {
      setSearch(q);
      setPage(1);
    }, 300);
    return () => clearTimeout(t);
  }, [q]);

  const mapelQuery = useQuery({
    queryKey: ["master", "mapel", "all"],
    queryFn: () => apiClient.get<ListRes<Mapel> | Mapel[]>("/api/v1/master/mapel?per_page=100"),
  });
  const mapelList = unwrapList(mapelQuery.data ?? []).items;

  const listQuery = useQuery({
    queryKey: ["master", "guru", search, page],
    queryFn: () =>
      apiClient.get<ListRes<Guru> | Guru[]>(
        `/api/v1/master/guru?page=${page}&per_page=20&q=${encodeURIComponent(search)}`,
      ),
  });
  const { items, meta } = unwrapList(listQuery.data ?? []);

  const invalidate = () => qc.invalidateQueries({ queryKey: ["master", "guru"] });
  const saveMutation = useMutation({
    mutationFn: async () => {
      const body = { user_id: userId.trim() || null, nip: nip.trim() || null, nama: nama.trim(), mapel_ids: mapelIds };
      if (editing) return apiClient.put(`/api/v1/master/guru/${editing.id}`, body);
      return apiClient.post("/api/v1/master/guru", body);
    },
    onSuccess: () => {
      setModalOpen(false);
      setFormError("");
      invalidate();
    },
    onError: (err) => setFormError(apiErrorMessage(err)),
  });
  const deleteMutation = useMutation({
    mutationFn: (id: string) => apiClient.del(`/api/v1/master/guru/${id}`),
    onSuccess: () => {
      invalidate();
    },
    onError: (err) => setPageError(apiErrorMessage(err)),
  });

  function openAdd() {
    setEditing(null);
    setNip("");
    setNama("");
    setUserId("");
    setMapelIds([]);
    setFormError("");
    setModalOpen(true);
  }
  function openEdit(item: Guru) {
    setEditing(item);
    setNip(item.nip ?? "");
    setNama(item.nama);
    setUserId(item.user_id ?? "");
    setMapelIds(Array.isArray(item.mapel_ids) ? item.mapel_ids : []);
    setFormError("");
    setModalOpen(true);
  }
  function toggleMapel(id: string) {
    setMapelIds((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]));
  }
  function submit() {
    if (!nama.trim()) return setFormError("Nama guru wajib diisi.");
    saveMutation.mutate();
  }
  const mapelNama = (ids: string[]) =>
    ids.map((id) => mapelList.find((m) => m.id === id)?.kode ?? id).join(", ") || "-";

  function konfirmasiHapus(item: Guru) {
    Modal.confirm({
      title: "Hapus Guru?",
      content: (
        <p className="text-sm text-muted">
        Guru <strong>{item.nama}</strong> akan dihapus permanen. Guru yang masih pemilik bank soal tidak
        bisa dihapus.
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
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: "Guru" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>Guru</Typography.Title>
          <Typography.Text type="secondary">Kelola data guru dan mapel yang diampu.</Typography.Text>
        </div>
        <Space>
          {hasPermission("master.guru.create") && (
            <Button onClick={openAdd} type="primary">
              Tambah Guru
            </Button>
          )}
        </Space>
      </div>
      <Input.Search placeholder="Cari nama / NIP guru…" value={q} onChange={(e) => setQ(e.target.value)} aria-label="Cari guru" size="large" allowClear />
      {pageError && (
        <Alert type="error" showIcon message={pageError} />
      )}
      {listQuery.isError && (
        <Alert type="error" showIcon message={<>Gagal memuat guru: {apiErrorMessage(listQuery.error)}</>} />
      )}
        <Table columns={[{ key: "col0", title: "Nama", dataIndex: "nama", sorter: (a: any, b: any) => { const _va = a.nama, _vb = b.nama; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, item: any) => (<>{item.nama}</>) }, { key: "col1", title: "NIP", dataIndex: "nip", sorter: (a: any, b: any) => { const _va = a.nip, _vb = b.nip; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, item: any) => (<>{item.nip ?? "-"}</>) }, { key: "col2", title: "Mapel Diampu", render: (_: unknown, item: any) => (<>{mapelNama(item.mapel_ids ?? [])}</>) }, { key: "col3", title: "Aksi", render: (_: unknown, item: any) => (<>
                  <div className="flex justify-end gap-2">
                    {hasPermission("master.guru.update") && (
                      <Button onClick={() => openEdit(item)} type="default">
                        Ubah
                      </Button>
                    )}
                    {hasPermission("master.guru.delete") && (
                      <Button onClick={() => konfirmasiHapus(item)} type="default" danger>
                        Hapus
                      </Button>
                    )}
                  </div>
                </>) }]} dataSource={items} rowKey="id" pagination={false} locale={{ emptyText: "Belum ada guru." }} scroll={{ x: "max-content" }} />
      {meta && meta.total_pages > 1 && (<Pagination current={meta.page} total={meta.total} pageSize={meta.per_page} onChange={(p) => setPage(p)} showTotal={(total) => `Halaman ${meta.page} dari ${meta.total_pages} · Total ${total}`} />)}
      <Modal open={modalOpen} onCancel={() => setModalOpen(false)} title={editing ? "Ubah Guru" : "Tambah Guru"} footer={null} destroyOnClose>
          <Form layout="vertical" onFinish={submit} className="w-full max-w-lg">
            {formError && (
              <Alert type="error" showIcon message={formError} />
            )}
            <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
              <Form.Item label="Nama *"><Input value={nama} onChange={(e) => setNama(e.target.value)} placeholder="Nama lengkap guru" size="large" /></Form.Item>
              <Form.Item label="NIP (opsional)"><Input value={nip} onChange={(e) => setNip(e.target.value)} placeholder="Nomor Induk Pegawai" size="large" /></Form.Item>
              <Form.Item label="Akun login (opsional)"><Input value={userId} onChange={(e) => setUserId(e.target.value)} placeholder="Kosongkan bila guru belum punya akun" size="large" />
                <span className="mt-1 block text-xs text-muted">
                  Kosongkan bila guru belum punya akun login. Bila sudah punya, isi ID akun (UUID)
                  pengguna tersebut. (Tidak ada daftar akun di sistem ini, jadi ID diisi manual.)
                </span></Form.Item>
              <Form.Item label="Mapel yang diampu" className="md:col-span-2">
                <div className="max-h-44 space-y-1 overflow-y-auto rounded-lg border border-border p-2">
                  {mapelList.length === 0 && <p className="text-sm text-muted">Belum ada mapel.</p>}
                  {mapelList.map((m) => (
                    <label key={m.id} className="flex items-center gap-2 text-sm text-ink">
                      <Checkbox checked={mapelIds.includes(m.id)} onChange={() => toggleMapel(m.id)} />
                      {m.nama} ({m.kode})
                    </label>
                  ))}
                </div>
              </Form.Item>
            </div>
            <div className="mt-4 flex flex-col gap-2 sm:flex-row sm:justify-end">
              <Button onClick={() => setModalOpen(false)} type="default">
                Batal
              </Button>
              {hasPermission(editing ? "master.guru.update" : "master.guru.create") && (
                <Button htmlType="submit" loading={saveMutation.isPending} type="primary">
                  {saveMutation.isPending ? "Menyimpan…" : "Simpan"}
                </Button>
              )}
            </div>
          </Form></Modal>
    </div>
  );
}
