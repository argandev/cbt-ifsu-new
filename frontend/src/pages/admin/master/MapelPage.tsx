// MapelPage — Master Data: Mata Pelajaran.
// Milik Tim 2. Lihat catatan kontrak apiClient di JurusanPage.tsx.
// Aturan: mapel Produktif wajib terikat satu jurusan; mapel Umum
// jurusannya kosong (ditegakkan backend + CHECK database).
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "../../../lib/api";
import { useAuthStore } from "../../../store/auth";

import { Alert, Button, Input, Modal, Pagination, Select, Table, Breadcrumb, Space, Typography, Form } from "antd";
type Mapel = { id: string; kode: string; nama: string; jenis: string; jurusan_id: string | null; kelas_tingkat: number | null };
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





export default function MapelPage() {
  const { hasPermission } = useAuthStore();
  const qc = useQueryClient();
  const [q, setQ] = useState("");
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [filterJenis, setFilterJenis] = useState("");
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Mapel | null>(null);
  const [kode, setKode] = useState("");
  const [nama, setNama] = useState("");
  const [jenis, setJenis] = useState("umum");
  const [jurusanId, setJurusanId] = useState("");
  const [kelasTingkat, setKelasTingkat] = useState("");
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
  const jurusanNama = (id: string | null) => (id ? (jurusanList.find((j) => j.id === id)?.nama ?? "-") : "Semua Jurusan");

  const listQuery = useQuery({
    queryKey: ["master", "mapel", search, page, filterJenis],
    queryFn: () =>
      apiClient.get<ListRes<Mapel> | Mapel[]>(
        `/api/v1/master/mapel?page=${page}&per_page=20&q=${encodeURIComponent(search)}` +
          (filterJenis ? `&jenis=${filterJenis}` : ""),
      ),
  });
  const { items, meta } = unwrapList(listQuery.data ?? []);

  const invalidate = () => qc.invalidateQueries({ queryKey: ["master", "mapel"] });
  const saveMutation = useMutation({
    mutationFn: async () => {
      const body = {
        kode: kode.trim(),
        nama: nama.trim(),
        jenis,
        jurusan_id: jenis === "produktif" ? jurusanId || null : null,
        kelas_tingkat: kelasTingkat ? Number(kelasTingkat) : null,
      };
      if (editing) return apiClient.put(`/api/v1/master/mapel/${editing.id}`, body);
      return apiClient.post("/api/v1/master/mapel", body);
    },
    onSuccess: () => {
      setModalOpen(false);
      setFormError("");
      invalidate();
    },
    onError: (err) => setFormError(apiErrorMessage(err)),
  });
  const deleteMutation = useMutation({
    mutationFn: (id: string) => apiClient.del(`/api/v1/master/mapel/${id}`),
    onSuccess: () => {
      invalidate();
    },
    onError: (err) => setPageError(apiErrorMessage(err)),
  });

  function openAdd() {
    setEditing(null);
    setKode("");
    setNama("");
    setJenis("umum");
    setJurusanId("");
    setKelasTingkat("");
    setFormError("");
    setModalOpen(true);
  }
  function openEdit(item: Mapel) {
    setEditing(item);
    setKode(item.kode);
    setNama(item.nama);
    setJenis(item.jenis);
    setJurusanId(item.jurusan_id ?? "");
    setKelasTingkat(item.kelas_tingkat != null ? String(item.kelas_tingkat) : "");
    setFormError("");
    setModalOpen(true);
  }
  function submit() {
    if (!kode.trim() || !nama.trim()) return setFormError("Kode dan nama mapel wajib diisi.");
    if (jenis === "produktif" && !jurusanId) return setFormError("Mapel produktif wajib terikat satu jurusan.");
    if (kelasTingkat && !["10", "11", "12"].includes(kelasTingkat)) return setFormError("Kelas tingkat harus 10, 11, atau 12.");
    saveMutation.mutate();
  }

  function konfirmasiHapus(item: Mapel) {
    Modal.confirm({
      title: "Hapus Mapel?",
      content: (
        <p className="text-sm text-muted">
        Mapel <strong>{item.nama}</strong> ({item.kode}) akan dihapus permanen. Mapel yang masih
        dipakai kompetensi/kisi-kisi/bank soal tidak bisa dihapus.
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
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: "Mata Pelajaran" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>Mata Pelajaran</Typography.Title>
          <Typography.Text type="secondary">Kelola mata pelajaran umum dan produktif.</Typography.Text>
        </div>
        <Space>
          {hasPermission("master.mapel.create") && (
            <Button onClick={openAdd} type="primary">
              Tambah Mapel
            </Button>
          )}
        </Space>
      </div>
      <div className="flex flex-col gap-2 sm:flex-row">
        <Input.Search placeholder="Cari kode / nama mapel…" value={q} onChange={(e) => setQ(e.target.value)} aria-label="Cari mapel" size="large" allowClear />
        <Select options={[
            { value: "", label: "Semua Jenis" },
            { value: "umum", label: "Umum" },
            { value: "produktif", label: "Produktif" },
          ]} value={filterJenis} onChange={(v) => {
            setFilterJenis(v);
            setPage(1);
          }} placeholder="Semua Jenis" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" />
      </div>
      {pageError && (
        <Alert type="error" showIcon message={pageError} />
      )}
      {listQuery.isError && (
        <Alert type="error" showIcon message={<>Gagal memuat mapel: {apiErrorMessage(listQuery.error)}</>} />
      )}
        <Table columns={[{ key: "col0", title: "Kode", dataIndex: "kode", sorter: (a: any, b: any) => { const _va = a.kode, _vb = b.kode; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, item: any) => (<>{item.kode}</>) }, { key: "col1", title: "Nama", dataIndex: "nama", sorter: (a: any, b: any) => { const _va = a.nama, _vb = b.nama; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, item: any) => (<>{item.nama}</>) }, { key: "col2", title: "Jenis", render: (_: unknown, item: any) => (<>{item.jenis === "produktif" ? "Produktif" : "Umum"}</>) }, { key: "col3", title: "Jurusan", render: (_: unknown, item: any) => (<>{jurusanNama(item.jurusan_id)}</>) }, { key: "col4", title: "Tingkat", dataIndex: "kelas_tingkat", sorter: (a: any, b: any) => { const _va = a.kelas_tingkat, _vb = b.kelas_tingkat; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, item: any) => (<>{item.kelas_tingkat ?? "Semua"}</>) }, { key: "col5", title: "Aksi", render: (_: unknown, item: any) => (<>
                  <div className="flex justify-end gap-2">
                    {hasPermission("master.mapel.update") && (
                      <Button onClick={() => openEdit(item)} type="default">
                        Ubah
                      </Button>
                    )}
                    {hasPermission("master.mapel.delete") && (
                      <Button onClick={() => konfirmasiHapus(item)} type="default" danger>
                        Hapus
                      </Button>
                    )}
                  </div>
                </>) }]} dataSource={items} rowKey="id" pagination={false} locale={{ emptyText: "Belum ada mapel." }} scroll={{ x: "max-content" }} />
      {meta && meta.total_pages > 1 && (<Pagination current={meta.page} total={meta.total} pageSize={meta.per_page} onChange={(p) => setPage(p)} showTotal={(total) => `Halaman ${meta.page} dari ${meta.total_pages} · Total ${total}`} />)}
      <Modal open={modalOpen} onCancel={() => setModalOpen(false)} title={editing ? "Ubah Mapel" : "Tambah Mapel"} footer={null} destroyOnClose>
          <Form layout="vertical" onFinish={submit} className="w-full max-w-lg">
            {formError && (
              <Alert type="error" showIcon message={formError} />
            )}
            <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
              <Form.Item label="Kode *"><Input value={kode} onChange={(e) => setKode(e.target.value)} placeholder="mis. PROG" size="large" /></Form.Item>
              <Form.Item label="Nama *"><Input value={nama} onChange={(e) => setNama(e.target.value)} placeholder="mis. Pemrograman Web" size="large" /></Form.Item>
              <div className="block">
                <Form.Item label="Jenis *"><Select options={[
                    { value: "umum", label: "Umum (semua jurusan)" },
                    { value: "produktif", label: "Produktif (satu jurusan)" },
                  ]} value={jenis} onChange={(v) => setJenis(v)} id="mapel-jenis" placeholder="Pilih…" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" /></Form.Item>
              </div>
              <div className="block">
                <Form.Item label={`Jurusan ${jenis === "produktif" ? "*" : "(kosong untuk Umum)"}`}><Select showSearch options={[
                    { value: "", label: "— Pilih Jurusan —" },
                    ...jurusanList.map((j) => ({ value: j.id, label: `${j.nama} (${j.kode})` })),
                  ]} value={jurusanId} onChange={(v) => setJurusanId(v)} placeholder="— Pilih Jurusan —" disabled={jenis !== "produktif"} id="mapel-jurusan" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" /></Form.Item>
              </div>
              <div className="block">
                <Form.Item label="Kelas Tingkat (opsional)"><Select options={[
                    { value: "", label: "Semua Tingkat" },
                    { value: "10", label: "10" },
                    { value: "11", label: "11" },
                    { value: "12", label: "12" },
                  ]} value={kelasTingkat} onChange={(v) => setKelasTingkat(v)} placeholder="Semua Tingkat" id="mapel-kelas-tingkat" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" /></Form.Item>
              </div>
            </div>
            <div className="mt-4 flex flex-col gap-2 sm:flex-row sm:justify-end">
              <Button onClick={() => setModalOpen(false)} type="default">
                Batal
              </Button>
              {hasPermission(editing ? "master.mapel.update" : "master.mapel.create") && (
                <Button htmlType="submit" loading={saveMutation.isPending} type="primary">
                  {saveMutation.isPending ? "Menyimpan…" : "Simpan"}
                </Button>
              )}
            </div>
          </Form></Modal>
    </div>
  );
}
