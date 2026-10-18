// KelasPage — Master Data: Kelas / Rombel.
// Milik Tim 2. Lihat catatan kontrak apiClient di JurusanPage.tsx.
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "../../../lib/api";
import { useAuthStore } from "../../../store/auth";

import { Alert, Button, Input, Modal, Pagination, Select, Table, Breadcrumb, Space, Typography, Form } from "antd";
type Kelas = { id: string; jurusan_id: string; tingkat: number; nama: string; tahun_ajaran: string };
type Jurusan = { id: string; kode: string; nama: string };
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





export default function KelasPage() {
  const { hasPermission } = useAuthStore();
  const qc = useQueryClient();
  const [q, setQ] = useState("");
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [filterJurusan, setFilterJurusan] = useState("");
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Kelas | null>(null);
  const [jurusanId, setJurusanId] = useState("");
  const [tingkat, setTingkat] = useState("11");
  const [nama, setNama] = useState("");
  const [tahunAjaran, setTahunAjaran] = useState("2026/2027");
  const [formError, setFormError] = useState("");
  const [pageError, setPageError] = useState("");

  useEffect(() => {
    const t = setTimeout(() => {
      setSearch(q);
      setPage(1);
    }, 300);
    return () => clearTimeout(t);
  }, [q]);

  const jurusanQuery = useQuery({
    queryKey: ["master", "jurusan", "all"],
    queryFn: () => apiClient.get<ListRes<Jurusan> | Jurusan[]>("/api/v1/master/jurusan?per_page=100"),
  });
  const jurusanList = unwrapList(jurusanQuery.data ?? []).items;
  const jurusanNama = (id: string) => jurusanList.find((j) => j.id === id)?.nama ?? "-";

  const listQuery = useQuery({
    queryKey: ["master", "kelas", search, page, filterJurusan],
    queryFn: () =>
      apiClient.get<ListRes<Kelas> | Kelas[]>(
        `/api/v1/master/kelas?page=${page}&per_page=20&q=${encodeURIComponent(search)}` +
          (filterJurusan ? `&jurusan_id=${filterJurusan}` : ""),
      ),
  });
  const { items, meta } = unwrapList(listQuery.data ?? []);

  const invalidate = () => qc.invalidateQueries({ queryKey: ["master", "kelas"] });
  const saveMutation = useMutation({
    mutationFn: async () => {
      const body = { jurusan_id: jurusanId, tingkat: Number(tingkat), nama: nama.trim(), tahun_ajaran: tahunAjaran.trim() };
      if (editing) return apiClient.put(`/api/v1/master/kelas/${editing.id}`, body);
      return apiClient.post("/api/v1/master/kelas", body);
    },
    onSuccess: () => {
      setModalOpen(false);
      setFormError("");
      invalidate();
    },
    onError: (err) => setFormError(apiErrorMessage(err)),
  });
  const deleteMutation = useMutation({
    mutationFn: (id: string) => apiClient.del(`/api/v1/master/kelas/${id}`),
    onSuccess: () => {
      invalidate();
    },
    onError: (err) => setPageError(apiErrorMessage(err)),
  });

  function openAdd() {
    setEditing(null);
    setJurusanId(jurusanList[0]?.id ?? "");
    setTingkat("11");
    setNama("");
    setTahunAjaran("2026/2027");
    setFormError("");
    setModalOpen(true);
  }
  function openEdit(item: Kelas) {
    setEditing(item);
    setJurusanId(item.jurusan_id);
    setTingkat(String(item.tingkat));
    setNama(item.nama);
    setTahunAjaran(item.tahun_ajaran);
    setFormError("");
    setModalOpen(true);
  }
  function submit() {
    if (!jurusanId) return setFormError("Jurusan wajib dipilih.");
    if (!["10", "11", "12"].includes(tingkat)) return setFormError("Tingkat harus 10, 11, atau 12.");
    if (!nama.trim() || !tahunAjaran.trim()) return setFormError("Nama kelas dan tahun ajaran wajib diisi.");
    saveMutation.mutate();
  }

  function konfirmasiHapus(item: Kelas) {
    Modal.confirm({
      title: "Hapus Kelas?",
      content: (
        <p className="text-sm text-muted">
        Kelas <strong>{item.nama}</strong> akan dihapus permanen. Kelas yang masih dipakai siswa tidak
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
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: "Kelas / Rombel" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>Kelas / Rombel</Typography.Title>
          <Typography.Text type="secondary">Kelola kelas / rombel per jurusan dan tahun ajaran.</Typography.Text>
        </div>
        <Space>
          {hasPermission("master.kelas.create") && (
            <Button onClick={openAdd} type="primary">
              Tambah Kelas
            </Button>
          )}
        </Space>
      </div>
      <div className="flex flex-col gap-2 sm:flex-row">
        <Input.Search placeholder="Cari nama kelas / tahun ajaran…" value={q} onChange={(e) => setQ(e.target.value)} aria-label="Cari kelas" size="large" allowClear />
        <Select showSearch options={[
            { value: "", label: "Semua Jurusan" },
            ...jurusanList.map((j) => ({ value: j.id, label: `${j.nama} (${j.kode})` })),
          ]} value={filterJurusan} onChange={(v) => {
            setFilterJurusan(v);
            setPage(1);
          }} placeholder="Semua Jurusan" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" />
      </div>
      {pageError && (
        <Alert type="error" showIcon message={pageError} />
      )}
      {listQuery.isError && (
        <Alert type="error" showIcon message={<>Gagal memuat kelas: {apiErrorMessage(listQuery.error)}</>} />
      )}
        <Table columns={[{ key: "col0", title: "Nama Kelas", dataIndex: "nama", sorter: (a: any, b: any) => { const _va = a.nama, _vb = b.nama; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, item: any) => (<>{item.nama}</>) }, { key: "col1", title: "Jurusan", render: (_: unknown, item: any) => (<>{jurusanNama(item.jurusan_id)}</>) }, { key: "col2", title: "Tingkat", dataIndex: "tingkat", sorter: (a: any, b: any) => { const _va = a.tingkat, _vb = b.tingkat; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, item: any) => (<>{item.tingkat}</>) }, { key: "col3", title: "Tahun Ajaran", dataIndex: "tahun_ajaran", sorter: (a: any, b: any) => { const _va = a.tahun_ajaran, _vb = b.tahun_ajaran; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, item: any) => (<>{item.tahun_ajaran}</>) }, { key: "col4", title: "Aksi", render: (_: unknown, item: any) => (<>
                  <div className="flex justify-end gap-2">
                    {hasPermission("master.kelas.update") && (
                      <Button onClick={() => openEdit(item)} type="default">
                        Ubah
                      </Button>
                    )}
                    {hasPermission("master.kelas.delete") && (
                      <Button onClick={() => konfirmasiHapus(item)} type="default" danger>
                        Hapus
                      </Button>
                    )}
                  </div>
                </>) }]} dataSource={items} rowKey="id" pagination={false} locale={{ emptyText: "Belum ada kelas." }} scroll={{ x: "max-content" }} />
      {meta && meta.total_pages > 1 && (<Pagination current={meta.page} total={meta.total} pageSize={meta.per_page} onChange={(p) => setPage(p)} showTotal={(total) => `Halaman ${meta.page} dari ${meta.total_pages} · Total ${total}`} />)}
      <Modal open={modalOpen} onCancel={() => setModalOpen(false)} title={editing ? "Ubah Kelas" : "Tambah Kelas"} footer={null} destroyOnClose>
          <Form layout="vertical" onFinish={submit} className="w-full max-w-lg">
            {formError && (
              <Alert type="error" showIcon message={formError} />
            )}
            <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
              <div className="block">
                <Form.Item label="Jurusan *"><Select showSearch options={[
                    { value: "", label: "— Pilih Jurusan —" },
                    ...jurusanList.map((j) => ({ value: j.id, label: `${j.nama} (${j.kode})` })),
                  ]} value={jurusanId} onChange={(v) => setJurusanId(v)} placeholder="— Pilih Jurusan —" id="kelas-jurusan" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" /></Form.Item>
              </div>
              <div className="block">
                <Form.Item label="Tingkat *"><Select options={[
                    { value: "10", label: "10" },
                    { value: "11", label: "11" },
                    { value: "12", label: "12" },
                  ]} value={tingkat} onChange={(v) => setTingkat(v)} id="kelas-tingkat" placeholder="Pilih…" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" /></Form.Item>
              </div>
              <Form.Item label="Nama Kelas *"><Input value={nama} onChange={(e) => setNama(e.target.value)} placeholder="mis. XI RPL 1" size="large" /></Form.Item>
              <Form.Item label="Tahun Ajaran *"><Input value={tahunAjaran} onChange={(e) => setTahunAjaran(e.target.value)} placeholder="mis. 2026/2027" size="large" /></Form.Item>
            </div>
            <div className="mt-4 flex flex-col gap-2 sm:flex-row sm:justify-end">
              <Button onClick={() => setModalOpen(false)} type="default">
                Batal
              </Button>
              {hasPermission(editing ? "master.kelas.update" : "master.kelas.create") && (
                <Button htmlType="submit" loading={saveMutation.isPending} type="primary">
                  {saveMutation.isPending ? "Menyimpan…" : "Simpan"}
                </Button>
              )}
            </div>
          </Form></Modal>
    </div>
  );
}
