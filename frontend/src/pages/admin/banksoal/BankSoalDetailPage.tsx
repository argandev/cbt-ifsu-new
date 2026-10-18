import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link, useNavigate, useParams } from "react-router-dom";
import { ArrowLeft, Eye, FileUp, Lock, LockOpen, Pencil, Plus, Trash2 } from "lucide-react";
import apiClient from "../../../lib/api";
import { useAuthStore } from "../../../store/auth";
import type { BankSoal, PageData, Soal, SoalFormValue } from "../../../lib/banksoal";
import { soalBaru, soalToForm } from "../../../lib/banksoal";
import { TIPE_LIST } from "../../../lib/tipeSoal";
import SoalEditor from "../../../components/soal/SoalEditor";
import SoalRenderer from "../../../components/soal/SoalRenderer";
import ImportSoalPanel from "./ImportSoalPanel";


import { Alert, Button, Input, Modal, Select, Form, Card, Breadcrumb, Tag, Typography } from "antd";

export default function BankSoalDetailPage() {
  const { hasPermission } = useAuthStore();
  const { id: bankId } = useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [pesan, setPesan] = useState("");
  const [filterTipe, setFilterTipe] = useState("");
  const [filterLevel, setFilterLevel] = useState("");
  const [searchInput, setSearchInput] = useState("");
  const [search, setSearch] = useState("");

  // Modal soal: null = tutup; {mode, soal?} = buka.
  const [editor, setEditor] = useState<{ mode: "tambah" | "ubah"; soal?: Soal } | null>(null);
  const [editorValue, setEditorValue] = useState<SoalFormValue>(soalBaru());
  const [preview, setPreview] = useState<Soal | null>(null);

  // Dialog konfirmasi (pengganti window.confirm).

  // Panel import soal (Excel/Word) — komponen terpisah di ImportSoalPanel.tsx.
  const [showImport, setShowImport] = useState(false);

  const bankQuery = useQuery({
    queryKey: ["bank-soal", bankId],
    queryFn: () => apiClient.get<BankSoal>(`/api/v1/bank-soal/${bankId}`),
    enabled: !!bankId,
  });

  const soalQuery = useQuery({
    queryKey: ["soal", bankId, filterTipe, filterLevel, search],
    queryFn: () => {
      const params = new URLSearchParams({ limit: "100", offset: "0" });
      if (filterTipe) params.set("tipe", filterTipe);
      if (filterLevel) params.set("level", filterLevel);
      if (search) params.set("q", search);
      return apiClient.get<PageData<Soal>>(`/api/v1/bank-soal/${bankId}/soal?${params.toString()}`);
    },
    // proxy: tidak ada permission bernama untuk "lihat isi soal";
    // banksoal.create dipakai sebagai proxy gate karena proktor
    // (banksoal.read) dilarang lihat isi soal oleh forbidSoalUntukProktor
    // di backend — cegah 403 proktor via URL langsung.
    enabled: !!bankId && hasPermission("banksoal.create"),
  });

  const refresh = () => {
    void queryClient.invalidateQueries({ queryKey: ["bank-soal", bankId] });
    void queryClient.invalidateQueries({ queryKey: ["soal", bankId] });
  };

  const lockMutation = useMutation({
    mutationFn: (locked: boolean) =>
      apiClient.post<BankSoal>(`/api/v1/bank-soal/${bankId}/${locked ? "lock" : "unlock"}`),
    onSuccess: refresh,
    onError: (err: Error) => setPesan(err.message),
  });

  const hapusBankMutation = useMutation({
    mutationFn: () => apiClient.del(`/api/v1/bank-soal/${bankId}`),
    onSuccess: () => navigate("/adm/bank-soal"),
    onError: (err: Error) => setPesan(err.message),
  });

  const simpanSoalMutation = useMutation({
    mutationFn: (value: SoalFormValue) => {
      const body = { ...value };
      return editor?.mode === "ubah" && editor.soal
        ? apiClient.put<Soal>(`/api/v1/bank-soal/${bankId}/soal/${editor.soal.id}`, body)
        : apiClient.post<Soal>(`/api/v1/bank-soal/${bankId}/soal`, body);
    },
    onSuccess: () => {
      setEditor(null);
      refresh();
    },
    onError: (err: Error) => setPesan(err.message),
  });

  const hapusSoalMutation = useMutation({
    mutationFn: (soalId: string) => apiClient.del(`/api/v1/bank-soal/${bankId}/soal/${soalId}`),
    onSuccess: refresh,
    onError: (err: Error) => setPesan(err.message),
  });

  const bank = bankQuery.data;
  const soalList = soalQuery.data?.items ?? [];

  const bukaTambah = () => {
    setEditorValue(soalBaru());
    setEditor({ mode: "tambah" });
  };
  const bukaUbah = (s: Soal) => {
    setEditorValue(soalToForm(s));
    setEditor({ mode: "ubah", soal: s });
  };

  function konfirmasiKunci() {
    Modal.confirm({
      title: "Kunci bank soal?",
      content: "Soal tidak bisa diubah sampai bank dibuka lagi.",
      okText: "Ya, Kunci",
      cancelText: "Batal",
      onOk: () => lockMutation.mutate(true),
    });
  }

  function konfirmasiHapusBank() {
    Modal.confirm({
      title: "Hapus bank soal?",
      content: `Bank soal "${bank?.nama ?? ""}" beserta seluruh soalnya akan dihapus permanen.`,
      okText: "Ya, Hapus",
      cancelText: "Batal",
      okButtonProps: { danger: true },
      onOk: () => hapusBankMutation.mutate(),
    });
  }

  function konfirmasiHapusSoal(soalId: string) {
    Modal.confirm({
      title: "Hapus soal?",
      content: "Soal ini akan dihapus permanen dari bank soal.",
      okText: "Ya, Hapus",
      cancelText: "Batal",
      okButtonProps: { danger: true },
      onOk: () => hapusSoalMutation.mutate(soalId),
    });
  }

  return (
    <div className="space-y-6">
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: (<Link to="/adm/bank-soal">Bank Soal</Link>) }, { title: "Detail" }]} />
      <div>
        <Link to="/adm/bank-soal" className="inline-flex items-center gap-1 text-sm text-muted hover:text-anbk">
          <ArrowLeft size={16} /> Kembali ke daftar bank soal
        </Link>
        {bankQuery.isLoading && <p className="mt-2 text-sm text-muted">Memuat…</p>}
        {bank && (
          <div className="mt-2 flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
            <div>
              <Typography.Title level={3} style={{ margin: 0 }}>{bank.nama}</Typography.Title>
              <Typography.Text type="secondary">
                Kode: {bank.kode} · Kelas {bank.kelas_tingkat} · {bank.jumlah_soal} soal
              </Typography.Text>
              {bank.is_locked ? (
                <Tag color="gold" className="mt-2"><Lock size={12} /> Terkunci — soal tidak bisa diubah/dihapus</Tag>
              ) : (
                <Tag color="green" className="mt-2"><LockOpen size={12} /> Terbuka</Tag>
              )}
            </div>
            <div className="flex flex-wrap gap-2">
              {hasPermission("banksoal.update") && (
                <Link to={`/adm/bank-soal/${bank.id}/ubah`}><Button>
                  <Pencil size={14} /> Ubah Bank
                </Button></Link>
              )}
              {hasPermission("banksoal.lock") && (
                <Button loading={lockMutation.isPending} onClick={() => {
                    if (bank.is_locked) {
                      lockMutation.mutate(false);
                    } else {
                      konfirmasiKunci();
                    }
                  }} type="default">
                  {bank.is_locked ? <><LockOpen size={14} /> Buka Kunci</> : <><Lock size={14} /> Kunci Bank</>}
                </Button>
              )}
              {hasPermission("banksoal.delete") && (
                <Button loading={hapusBankMutation.isPending} onClick={() => konfirmasiHapusBank()} type="default" danger>
                  <Trash2 size={14} /> Hapus Bank
                </Button>
              )}
            </div>
          </div>
        )}
      </div>

      {pesan && <Alert type="error" showIcon message={<>{pesan}</>} />}

      {/* Filter + aksi soal */}
      <Card>
        <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-end">
          <div className="flex-1 min-w-40">
            <Typography.Text strong className="mb-1 block">Filter tipe</Typography.Text>
            <Select showSearch options={[
                { value: "", label: "Semua tipe" },
                ...TIPE_LIST.map((t) => ({ value: t.kode, label: t.label })),
              ]} value={filterTipe} onChange={(v) => setFilterTipe(v)} placeholder="Semua tipe" id="filter-tipe" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" />
          </div>
          <div className="w-full sm:w-40">
            <Typography.Text strong className="mb-1 block">Filter level</Typography.Text>
            <Select options={[
                { value: "", label: "Semua level" },
                { value: "mudah", label: "Mudah" },
                { value: "sedang", label: "Sedang" },
                { value: "sulit", label: "Sulit" },
              ]} value={filterLevel} onChange={(v) => setFilterLevel(v)} placeholder="Semua level" id="filter-level" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" />
          </div>
          <Form layout="vertical" onFinish={() => setSearch(searchInput)} className="flex flex-1 items-end gap-2 min-w-52">
            <Form.Item label="Cari soal" className="mb-0 flex-1">
              <Input id="cari-soal" value={searchInput} onChange={(e) => setSearchInput(e.target.value)} placeholder="Cari isi pertanyaan…" size="large" />
            </Form.Item>
            <Button htmlType="submit" type="default">Cari</Button>
          </Form>
          <div className="flex flex-wrap gap-2">
            {hasPermission("banksoal.create") && (
              <Button disabled={bank?.is_locked} onClick={bukaTambah} type="primary">
                <Plus size={16} /> Tambah Soal
              </Button>
            )}
            {hasPermission("banksoal.import") && (
              <Button disabled={bank?.is_locked} onClick={() => setShowImport((v) => !v)} type="default">
                <FileUp size={16} /> Import Soal
              </Button>
            )}
          </div>
        </div>

        {showImport && (
          <ImportSoalPanel
            bankId={bankId ?? ""}
            terkunci={!!bank?.is_locked}
            onImported={refresh}
            onTutup={() => setShowImport(false)}
          />
        )}
      </Card>

      {/* Daftar soal */}
      {soalQuery.isLoading && <p className="text-sm text-muted">Memuat soal…</p>}
      {soalList.length === 0 && !soalQuery.isLoading && (
        <p className="rounded-lg border border-dashed border-border px-4 py-8 text-center text-sm text-muted">
          Belum ada soal di bank ini (atau tidak cocok dengan filter).
        </p>
      )}
      <div className="space-y-3">
        {soalList.map((s, i) => (
          <Card key={s.id}>
          <div className="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
            <div className="min-w-0">
              <p className="text-xs font-medium text-muted">Soal {i + 1} · {TIPE_LIST.find((t) => t.kode === s.tipe)?.label ?? s.tipe} · {s.level} · Bobot {s.bobot} · {s.status}</p>
              <p className="mt-1 line-clamp-2 text-sm text-ink">{s.pertanyaan}</p>
            </div>
            <div className="flex shrink-0 flex-wrap gap-2">
              {/* proxy: tidak ada permission bernama untuk "lihat isi soal";
                  banksoal.create dipakai sebagai proxy gate karena proktor
                  (banksoal.read) dilarang lihat isi soal oleh
                  forbidSoalUntukProktor di backend. */}
              {hasPermission("banksoal.create") && (
                <Button onClick={() => setPreview(s)} type="default">
                  <Eye size={14} /> Pratinjau
                </Button>
              )}
              {hasPermission("banksoal.update") && (
                <Button disabled={bank?.is_locked} onClick={() => bukaUbah(s)} type="default">
                  <Pencil size={14} /> Ubah
                </Button>
              )}
              {hasPermission("banksoal.delete") && (
                <Button disabled={bank?.is_locked} loading={hapusSoalMutation.isPending} onClick={() => konfirmasiHapusSoal(s.id)} type="default" danger>
                  <Trash2 size={14} /> Hapus
                </Button>
              )}
            </div>
          </div>
        </Card>
        ))}
      </div>

      {/* Modal editor soal */}
      <Modal
        open={!!editor}
        onCancel={() => setEditor(null)}
        title={editor?.mode === "tambah" ? "Tambah Soal" : "Ubah Soal"}
        footer={null}
        width={800}
        destroyOnClose
      >
        <SoalEditor value={editorValue} onChange={setEditorValue} />
        {pesan && <Alert type="error" showIcon message={<>{pesan}</>} />}
        <div className="mt-4 flex flex-wrap gap-2">
          {hasPermission(editor?.mode === "tambah" ? "banksoal.create" : "banksoal.update") && (
            <Button loading={simpanSoalMutation.isPending} onClick={() => simpanSoalMutation.mutate(editorValue)} type="primary">
              {simpanSoalMutation.isPending ? "Menyimpan…" : "Simpan Soal"}
            </Button>
          )}
          <Button onClick={() => setEditor(null)} type="default">Batal</Button>
        </div>
      </Modal>

      {/* Modal pratinjau guru (dengan kunci) */}
      <Modal
        open={!!preview}
        onCancel={() => setPreview(null)}
        title="Pratinjau Soal (Guru — kunci ditampilkan)"
        footer={null}
        width={800}
        destroyOnClose
      >
        {preview && <SoalRenderer soal={preview} showKunci />}
      </Modal>

    </div>
  );
}
