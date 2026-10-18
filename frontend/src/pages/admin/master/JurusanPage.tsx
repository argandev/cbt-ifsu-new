// JurusanPage — Master Data: Jurusan.
// Milik Tim 2. Kontrak apiClient (Tim 1, frontend/src/lib/api.ts):
//   apiClient.get<T>(url) / post<T>(url, body) / put<T>(url, body) / del<T>(url)
//   mengembalikan JSON terurai; untuk list berupa envelope
//   { data: [...], meta: { page, per_page, total, total_pages } }.
// Halaman ini menoleransi apiClient yang mengembalikan envelope maupun
// array mentah. CSRF (header X-CSRF-Token) ditangani di dalam apiClient.
// Komponen UI Tim 1 belum ada di repo saat halaman ini ditulis, jadi
// halaman memakai elemen HTML + Tailwind (mudah diganti komponen UI
// nanti): tombol primary token anbk, font Plus Jakarta Sans (global),
// NOL emoji di UI.
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "../../../lib/api";
import { useAuthStore } from "../../../store/auth";

import { Alert, Button, Input, Modal, Pagination, Table, Breadcrumb, Space, Typography, Form } from "antd";
type Jurusan = { id: string; kode: string; nama: string };
type Meta = { page: number; per_page: number; total: number; total_pages: number };
type ListRes = { data: Jurusan[]; meta?: Meta };

function unwrapList(res: ListRes | Jurusan[]): { items: Jurusan[]; meta?: Meta } {
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






export default function JurusanPage() {
  const { hasPermission } = useAuthStore();
  const qc = useQueryClient();
  const [q, setQ] = useState("");
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Jurusan | null>(null);
  const [kode, setKode] = useState("");
  const [nama, setNama] = useState("");
  const [formError, setFormError] = useState("");
  const [pageError, setPageError] = useState("");

  // Debounce pencarian 300ms.
  useEffect(() => {
    const t = setTimeout(() => {
      setSearch(q);
      setPage(1);
    }, 300);
    return () => clearTimeout(t);
  }, [q]);

  const listQuery = useQuery({
    queryKey: ["master", "jurusan", search, page],
    queryFn: () =>
      apiClient.get<ListRes | Jurusan[]>(
        `/api/v1/master/jurusan?page=${page}&per_page=20&q=${encodeURIComponent(search)}`,
      ),
  });
  const { items, meta } = unwrapList(listQuery.data ?? []);

  const invalidate = () => qc.invalidateQueries({ queryKey: ["master", "jurusan"] });

  const saveMutation = useMutation({
    mutationFn: async () => {
      const body = { kode: kode.trim(), nama: nama.trim() };
      if (editing) return apiClient.put(`/api/v1/master/jurusan/${editing.id}`, body);
      return apiClient.post("/api/v1/master/jurusan", body);
    },
    onSuccess: () => {
      setModalOpen(false);
      setFormError("");
      invalidate();
    },
    onError: (err) => setFormError(apiErrorMessage(err)),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => apiClient.del(`/api/v1/master/jurusan/${id}`),
    onSuccess: () => {
      invalidate();
    },
    onError: (err) => setPageError(apiErrorMessage(err)),
  });

  function openAdd() {
    setEditing(null);
    setKode("");
    setNama("");
    setFormError("");
    setModalOpen(true);
  }
  function openEdit(item: Jurusan) {
    setEditing(item);
    setKode(item.kode);
    setNama(item.nama);
    setFormError("");
    setModalOpen(true);
  }
  function submit() {
    if (!kode.trim() || !nama.trim()) {
      setFormError("Kode dan nama jurusan wajib diisi.");
      return;
    }
    saveMutation.mutate();
  }

  function konfirmasiHapus(item: Jurusan) {
    Modal.confirm({
      title: "Hapus Jurusan?",
      content: (
        <p className="text-sm text-muted">
        Jurusan <strong>{item.nama}</strong> ({item.kode}) akan dihapus permanen. Jurusan yang
        masih dipakai kelas/mapel/siswa tidak bisa dihapus.
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
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: "Jurusan" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>Jurusan</Typography.Title>
          <Typography.Text type="secondary">Kelola jurusan / kompetensi keahlian.</Typography.Text>
        </div>
        <Space>
          {hasPermission("master.jurusan.create") && (
            <Button onClick={openAdd} type="primary">
              Tambah Jurusan
            </Button>
          )}
        </Space>
      </div>

      <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
        <Input.Search placeholder="Cari kode / nama jurusan…" value={q} onChange={(e) => setQ(e.target.value)} aria-label="Cari jurusan" size="large" allowClear />
      </div>

      {pageError && (
        <Alert type="error" showIcon message={pageError} />
      )}
      {listQuery.isError && (
        <Alert type="error" showIcon message={<>Gagal memuat jurusan: {apiErrorMessage(listQuery.error)}</>} />
      )}

        <Table columns={[{ key: "col0", title: "Kode", dataIndex: "kode", sorter: (a: any, b: any) => { const _va = a.kode, _vb = b.kode; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, item: any) => (<>{item.kode}</>) }, { key: "col1", title: "Nama", dataIndex: "nama", sorter: (a: any, b: any) => { const _va = a.nama, _vb = b.nama; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, item: any) => (<>{item.nama}</>) }, { key: "col2", title: "Aksi", render: (_: unknown, item: any) => (<>
                  <div className="flex justify-end gap-2">
                    {hasPermission("master.jurusan.update") && (
                      <Button onClick={() => openEdit(item)} type="default">
                        Ubah
                      </Button>
                    )}
                    {hasPermission("master.jurusan.delete") && (
                      <Button onClick={() => konfirmasiHapus(item)} type="default" danger>
                        Hapus
                      </Button>
                    )}
                  </div>
                </>) }]} dataSource={items} rowKey="id" pagination={false} locale={{ emptyText: "Belum ada jurusan." }} scroll={{ x: "max-content" }} />

      {meta && meta.total_pages > 1 && (<Pagination current={meta.page} total={meta.total} pageSize={meta.per_page} onChange={(p) => setPage(p)} showTotal={(total) => `Halaman ${meta.page} dari ${meta.total_pages} · Total ${total}`} />)}

      <Modal open={modalOpen} onCancel={() => setModalOpen(false)} title={editing ? "Ubah Jurusan" : "Tambah Jurusan"} footer={null} destroyOnClose>
          <Form layout="vertical" onFinish={submit} className="w-full max-w-lg">
            {formError && (
              <Alert type="error" showIcon message={formError} />
            )}
            <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
              <Form.Item label="Kode *"><Input value={kode} onChange={(e) => setKode(e.target.value)} placeholder="mis. RPL" size="large" /></Form.Item>
              <Form.Item label="Nama *"><Input value={nama} onChange={(e) => setNama(e.target.value)} placeholder="mis. Rekayasa Perangkat Lunak" size="large" /></Form.Item>
            </div>
            <div className="mt-4 flex flex-col gap-2 sm:flex-row sm:justify-end">
              <Button onClick={() => setModalOpen(false)} type="default">
                Batal
              </Button>
              {hasPermission(editing ? "master.jurusan.update" : "master.jurusan.create") && (
                <Button htmlType="submit" loading={saveMutation.isPending} type="primary">
                  {saveMutation.isPending ? "Menyimpan…" : "Simpan"}
                </Button>
              )}
            </div>
          </Form></Modal>

    </div>
  );
}
