// SiswaPage — Master Data: Siswa.
// Milik Tim 2. Lihat catatan kontrak apiClient di JurusanPage.tsx.
// Tambah siswa manual ikut membuat akun login (username = NISN,
// password awal = NISN, wajib ganti saat login pertama) bila user_id
// dikosongkan — sama seperti import CSV.
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "../../../lib/api";
import { useAuthStore } from "../../../store/auth";

import { Alert, Button, Checkbox, Input, InputNumber, Modal, Pagination, Select, Table, Tag, Breadcrumb, Space, Typography, Form } from "antd";
import { statusTagColor } from "../../../components/common/statusTag";
type Siswa = {
  id: string; nisn: string; nis: string | null; nama: string;
  kelas_id: string; jurusan_id: string; tahun_masuk: number | null; is_active: boolean;
};
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






export default function SiswaPage() {
  const { hasPermission } = useAuthStore();
  const qc = useQueryClient();
  const [q, setQ] = useState("");
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [filterKelas, setFilterKelas] = useState("");
  const [filterJurusan, setFilterJurusan] = useState("");
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Siswa | null>(null);
  const [nisn, setNisn] = useState("");
  const [nis, setNis] = useState("");
  const [nama, setNama] = useState("");
  const [kelasId, setKelasId] = useState("");
  const [jurusanId, setJurusanId] = useState("");
  const [tahunMasuk, setTahunMasuk] = useState("");
  const [isActive, setIsActive] = useState(true);
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
  const kelasQuery = useQuery({
    queryKey: ["master", "kelas", "all"],
    queryFn: () => apiClient.get<ListRes<Kelas> | Kelas[]>("/api/v1/master/kelas?per_page=100"),
  });
  const kelasList = unwrapList(kelasQuery.data ?? []).items;
  const kelasNama = (id: string) => kelasList.find((k) => k.id === id)?.nama ?? "-";
  const jurusanNama = (id: string) => jurusanList.find((j) => j.id === id)?.nama ?? "-";

  const listQuery = useQuery({
    queryKey: ["master", "siswa", search, page, filterKelas, filterJurusan],
    queryFn: () =>
      apiClient.get<ListRes<Siswa> | Siswa[]>(
        `/api/v1/master/siswa?page=${page}&per_page=20&q=${encodeURIComponent(search)}` +
          (filterKelas ? `&kelas_id=${filterKelas}` : "") +
          (filterJurusan ? `&jurusan_id=${filterJurusan}` : ""),
      ),
  });
  const { items, meta } = unwrapList(listQuery.data ?? []);

  const invalidate = () => qc.invalidateQueries({ queryKey: ["master", "siswa"] });
  const saveMutation = useMutation({
    mutationFn: async () => {
      const body = {
        nisn: nisn.trim(),
        nis: nis.trim() || null,
        nama: nama.trim(),
        kelas_id: kelasId,
        jurusan_id: jurusanId,
        tahun_masuk: tahunMasuk ? Number(tahunMasuk) : null,
        is_active: isActive,
      };
      if (editing) return apiClient.put(`/api/v1/master/siswa/${editing.id}`, body);
      return apiClient.post("/api/v1/master/siswa", body);
    },
    onSuccess: () => {
      setModalOpen(false);
      setFormError("");
      invalidate();
    },
    onError: (err) => setFormError(apiErrorMessage(err)),
  });
  const deleteMutation = useMutation({
    mutationFn: (id: string) => apiClient.del(`/api/v1/master/siswa/${id}`),
    onSuccess: () => {
      invalidate();
    },
    onError: (err) => setPageError(apiErrorMessage(err)),
  });

  function onKelasChange(id: string) {
    setKelasId(id);
    const kelas = kelasList.find((k) => k.id === id);
    if (kelas) setJurusanId(kelas.jurusan_id);
  }
  function openAdd() {
    setEditing(null);
    setNisn("");
    setNis("");
    setNama("");
    setKelasId("");
    setJurusanId("");
    setTahunMasuk("");
    setIsActive(true);
    setFormError("");
    setModalOpen(true);
  }
  function openEdit(item: Siswa) {
    setEditing(item);
    setNisn(item.nisn);
    setNis(item.nis ?? "");
    setNama(item.nama);
    setKelasId(item.kelas_id);
    setJurusanId(item.jurusan_id);
    setTahunMasuk(item.tahun_masuk != null ? String(item.tahun_masuk) : "");
    setIsActive(item.is_active);
    setFormError("");
    setModalOpen(true);
  }
  function submit() {
    if (!nisn.trim() || !nama.trim()) return setFormError("NISN dan nama siswa wajib diisi.");
    if (!kelasId || !jurusanId) return setFormError("Kelas dan jurusan wajib dipilih.");
    saveMutation.mutate();
  }

  function konfirmasiHapus(item: Siswa) {
    Modal.confirm({
      title: "Hapus Siswa?",
      content: (
        <p className="text-sm text-muted">
        Siswa <strong>{item.nama}</strong> (NISN {item.nisn}) akan dihapus permanen. Akun
        loginnya tidak ikut dihapus. Siswa yang sudah punya riwayat ujian sebaiknya dinonaktifkan saja.
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
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: "Siswa" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>Siswa</Typography.Title>
          <Typography.Text type="secondary">Kelola data siswa. Akun login dibuat otomatis saat tambah manual atau import CSV.</Typography.Text>
        </div>
        <Space wrap>
          <>
            {hasPermission("master.siswa.import") && (
              <Button href="/adm/master/siswa/import">
                Import CSV
              </Button>
            )}
            {hasPermission("master.siswa.create") && (
              <Button onClick={openAdd} type="primary">
                Tambah Siswa
              </Button>
            )}
          </>
        </Space>
      </div>
      <div className="flex flex-col gap-2 sm:flex-row">
        <Input.Search placeholder="Cari nama / NISN / NIS…" value={q} onChange={(e) => setQ(e.target.value)} aria-label="Cari siswa" size="large" allowClear />
        <Select showSearch options={[
            { value: "", label: "Semua Jurusan" },
            ...jurusanList.map((j) => ({ value: j.id, label: `${j.nama} (${j.kode})` })),
          ]} value={filterJurusan} onChange={(v) => {
            setFilterJurusan(v);
            setPage(1);
          }} placeholder="Semua Jurusan" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" />
        <Select showSearch options={[
            { value: "", label: "Semua Kelas" },
            ...kelasList.map((k) => ({ value: k.id, label: k.nama })),
          ]} value={filterKelas} onChange={(v) => {
            setFilterKelas(v);
            setPage(1);
          }} placeholder="Semua Kelas" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" />
      </div>
      {pageError && (
        <Alert type="error" showIcon message={pageError} />
      )}
      {listQuery.isError && (
        <Alert type="error" showIcon message={`Gagal memuat siswa: ${apiErrorMessage(listQuery.error)}`} />
      )}
        <Table columns={[{ key: "col0", title: "NISN", dataIndex: "nisn", sorter: (a: any, b: any) => { const _va = a.nisn, _vb = b.nisn; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, item: any) => (<>{item.nisn}</>) }, { key: "col1", title: "NIS", dataIndex: "nis", sorter: (a: any, b: any) => { const _va = a.nis, _vb = b.nis; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, item: any) => (<>{item.nis ?? "-"}</>) }, { key: "col2", title: "Nama", dataIndex: "nama", sorter: (a: any, b: any) => { const _va = a.nama, _vb = b.nama; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, item: any) => (<>{item.nama}</>) }, { key: "col3", title: "Kelas", render: (_: unknown, item: any) => (<>{kelasNama(item.kelas_id)}</>) }, { key: "col4", title: "Jurusan", render: (_: unknown, item: any) => (<>{jurusanNama(item.jurusan_id)}</>) }, { key: "col5", title: "Status", render: (_: unknown, item: any) => (<><Tag color={statusTagColor(item.is_active ? "aktif" : "tidak aktif")}>{item.is_active ? "Aktif" : "Nonaktif"}</Tag></>) }, { key: "col6", title: "Aksi", render: (_: unknown, item: any) => (<>
                  <div className="flex justify-end gap-2">
                    {hasPermission("master.siswa.update") && (
                      <Button onClick={() => openEdit(item)} type="default">
                        Ubah
                      </Button>
                    )}
                    {hasPermission("master.siswa.delete") && (
                      <Button onClick={() => konfirmasiHapus(item)} type="default" danger>
                        Hapus
                      </Button>
                    )}
                  </div>
                </>) }]} dataSource={items} rowKey="id" pagination={false} locale={{ emptyText: "Belum ada siswa." }} scroll={{ x: "max-content" }} />
      {meta && meta.total_pages > 1 && (<Pagination current={meta.page} total={meta.total} pageSize={meta.per_page} onChange={(p) => setPage(p)} showTotal={(total) => `Halaman ${meta.page} dari ${meta.total_pages} · Total ${total}`} />)}
      <Modal open={modalOpen} onCancel={() => setModalOpen(false)} title={editing ? "Ubah Siswa" : "Tambah Siswa"} footer={null} destroyOnClose>
          <Form layout="vertical" onFinish={submit} className="w-full max-w-lg">
            {formError && (
              <Alert type="error" showIcon message={formError} style={{ marginBottom: 12 }} />
            )}
            {!editing && (
              <Alert type="info" showIcon message={<>
                Akun login dibuat otomatis: username = NISN, password awal = NISN. Siswa wajib ganti password saat
                login pertama.
              </>} />
            )}
            <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
              <Form.Item label="NISN *"><Input value={nisn} onChange={(e) => setNisn(e.target.value)} placeholder="10 digit NISN" size="large" /></Form.Item>
              <Form.Item label="NIS (opsional)"><Input value={nis} onChange={(e) => setNis(e.target.value)} placeholder="Nomor Induk Siswa" size="large" /></Form.Item>
              <Form.Item label="Nama *"><Input value={nama} onChange={(e) => setNama(e.target.value)} placeholder="Nama lengkap siswa" size="large" /></Form.Item>
              <div className="block">
                <Form.Item label="Kelas *"><Select showSearch options={[
                    { value: "", label: "— Pilih Kelas —" },
                    ...kelasList.map((k) => ({ value: k.id, label: k.nama })),
                  ]} value={kelasId} onChange={(v) => onKelasChange(v)} placeholder="— Pilih Kelas —" id="siswa-kelas" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" /></Form.Item>
              </div>
              <div className="block">
                <Form.Item label="Jurusan *"><Select showSearch options={[
                    { value: "", label: "— Pilih Jurusan —" },
                    ...jurusanList.map((j) => ({ value: j.id, label: `${j.nama} (${j.kode})` })),
                  ]} value={jurusanId} onChange={(v) => setJurusanId(v)} placeholder="— Pilih Jurusan —" id="siswa-jurusan" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" /></Form.Item>
              </div>
              <Form.Item label="Tahun Masuk (opsional)"><InputNumber value={tahunMasuk === "" ? null : Number(tahunMasuk)} onChange={(v) => setTahunMasuk(v === null ? "" : String(v))} placeholder="mis. 2026" size="large" style={{ width: "100%" }} /></Form.Item>
              <Form.Item style={{ marginBottom: 0 }}><Checkbox checked={isActive} onChange={(e) => setIsActive(e.target.checked)}>Siswa aktif</Checkbox></Form.Item>
            </div>
            <div className="mt-4 flex flex-col gap-2 sm:flex-row sm:justify-end">
              <Button onClick={() => setModalOpen(false)} type="default">
                Batal
              </Button>
              {hasPermission(editing ? "master.siswa.update" : "master.siswa.create") && (
                <Button htmlType="submit" loading={saveMutation.isPending} type="primary">
                  {saveMutation.isPending ? "Menyimpan…" : "Simpan"}
                </Button>
              )}
            </div>
          </Form></Modal>
    </div>
  );
}
