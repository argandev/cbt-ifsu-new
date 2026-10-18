// PaketPage — daftar paket (halaman utama) + form buat/ubah + detail paket
// (Tim B, penataan — audit §4.2).
// SEBELUMNYA form + daftar + panel soal dalam 1 halaman (grid 2 kolom +
// panel soal di bawah); SEKARANG: daftar adalah tabel penuh, form tampil
// sendiri, panel soal pindah ke halaman detail paket.
// Logika TIDAK berubah: endpoint, payload, pilih soal via dropdown bank +
// tombol "+ Tambah", generate acak, dan hapus via ConfirmDialog sama.
// Soal ditambahkan lewat dropdown bank soal + tombol "+ Tambah" per soal
// (tidak ada lagi input ID manual).
import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Eye, Pencil, Trash2, Plus } from "lucide-react";
import { apiClient } from "../../../lib/api";
import PaketFormPage, { type PaketFormValues } from "./PaketFormPage";
import PaketDetailPage, { type PaketDetailData, type PaketSoalItem } from "./PaketDetailPage";
import type { BankSoal, PageData, Soal } from "../../../lib/banksoal";

import { Alert, Button, Empty, Input, Modal, Table, Breadcrumb, Space, Typography } from "antd";
interface Mapel {
  id: string;
  kode: string;
  nama: string;
}

const formKosong: PaketFormValues = {
  kode: "",
  nama: "",
  mapel_id: "",
  jenis: "STS",
  durasi_menit: 90,
  jumlah_soal: 0,
  acak_soal: true,
  acak_jawaban: true,
  tampil_hasil: false,
};

export default function PaketPage() {
  const qc = useQueryClient();
  const [tampilan, setTampilan] = useState<"daftar" | "form" | "detail">("daftar");
  const [form, setForm] = useState<PaketFormValues>(formKosong);
  const [editId, setEditId] = useState<string | null>(null);
  const [paketDipilih, setPaketDipilih] = useState<string | null>(null);
  const [bankDipilih, setBankDipilih] = useState("");
  const [jumlahAcak, setJumlahAcak] = useState(10);
  const [pesan, setPesan] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [cari, setCari] = useState("");

  const paketQuery = useQuery({
    queryKey: ["paket"],
    queryFn: () => apiClient.get<{ items: PaketDetailData[] }>("/api/v1/paket"),
  });
  // Daftar mapel untuk dropdown — endpoint yang sama dipakai GuruPage.
  const mapelQuery = useQuery({
    queryKey: ["master", "mapel", "all"],
    queryFn: () => apiClient.get<{ data: Mapel[] } | Mapel[]>("/api/v1/master/mapel?per_page=100"),
  });
  const mapelList: Mapel[] = Array.isArray(mapelQuery.data)
    ? mapelQuery.data
    : (mapelQuery.data?.data ?? []);
  const soalPaketQuery = useQuery({
    queryKey: ["paket-soal", paketDipilih],
    queryFn: () =>
      apiClient.get<{ items: PaketSoalItem[] }>(`/api/v1/paket/${paketDipilih}/soal`),
    enabled: !!paketDipilih,
  });
  const bankQuery = useQuery({
    queryKey: ["bank-soal"],
    queryFn: () => apiClient.get<PageData<BankSoal>>("/api/v1/bank-soal?limit=100"),
    enabled: !!paketDipilih,
    retry: false,
  });
  const soalBankQuery = useQuery({
    queryKey: ["bank-soal-items", bankDipilih],
    queryFn: () =>
      apiClient.get<PageData<Soal>>(`/api/v1/bank-soal/${bankDipilih}/soal?limit=100`),
    enabled: !!bankDipilih,
    retry: false,
  });

  const semuaPaket = paketQuery.data?.items ?? [];
  const tampil = useMemo(() => {
    const q = cari.trim().toLowerCase();
    if (!q) return semuaPaket;
    return semuaPaket.filter(
      (p) =>
        p.nama.toLowerCase().includes(q) ||
        p.kode.toLowerCase().includes(q) ||
        (p.mapel_nama ?? "").toLowerCase().includes(q),
    );
  }, [semuaPaket, cari]);

  const paketAktif = useMemo(
    () => semuaPaket.find((p) => p.id === paketDipilih),
    [semuaPaket, paketDipilih],
  );

  useEffect(() => {
    if (paketAktif) setBankDipilih("");
  }, [paketAktif?.id]); // eslint-disable-line react-hooks/exhaustive-deps

  const refresh = () => {
    qc.invalidateQueries({ queryKey: ["paket"] });
    qc.invalidateQueries({ queryKey: ["paket-soal", paketDipilih] });
  };

  const bukaFormBaru = () => {
    setEditId(null);
    setForm(formKosong);
    setError(null);
    setPesan(null);
    setTampilan("form");
  };

  const bukaFormUbah = (p: PaketDetailData) => {
    setEditId(p.id);
    setForm({
      kode: p.kode, nama: p.nama, mapel_id: p.mapel_id, jenis: p.jenis,
      durasi_menit: p.durasi_menit, jumlah_soal: p.jumlah_soal,
      acak_soal: p.acak_soal, acak_jawaban: p.acak_jawaban,
      tampil_hasil: p.tampil_hasil,
    });
    setError(null);
    setPesan(null);
    setTampilan("form");
  };

  const bukaDetail = (id: string) => {
    setPaketDipilih(id);
    setError(null);
    setPesan(null);
    setTampilan("detail");
  };

  const kembaliKeDaftar = () => {
    setPaketDipilih(null);
    setEditId(null);
    setForm(formKosong);
    setTampilan("daftar");
  };

  const simpanPaket = async () => {
    setError(null);
    setPesan(null);
    try {
      if (editId) {
        await apiClient.put(`/api/v1/paket/${editId}`, form);
        setPesan("Paket diubah.");
      } else {
        await apiClient.post("/api/v1/paket", form);
        setPesan("Paket dibuat.");
      }
      setForm(formKosong);
      setEditId(null);
      setTampilan("daftar");
      refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Gagal menyimpan paket.");
    }
  };

  const hapusPaket = async (id: string) => {
    try {
      await apiClient.del(`/api/v1/paket/${id}`);
      if (paketDipilih === id) setPaketDipilih(null);
      refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Gagal menghapus paket.");
    }
  };

  const tambahSoal = async (soalId: string) => {
    if (!paketDipilih || !soalId) return;
    try {
      await apiClient.post(`/api/v1/paket/${paketDipilih}/soal`, { soal_id: soalId });
      qc.invalidateQueries({ queryKey: ["paket-soal", paketDipilih] });
      qc.invalidateQueries({ queryKey: ["paket"] });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Gagal menambah soal.");
    }
  };

  const hapusSoal = async (soalId: string) => {
    if (!paketDipilih) return;
    try {
      await apiClient.del(`/api/v1/paket/${paketDipilih}/soal?soal_id=${encodeURIComponent(soalId)}`);
      qc.invalidateQueries({ queryKey: ["paket-soal", paketDipilih] });
      qc.invalidateQueries({ queryKey: ["paket"] });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Gagal menghapus soal.");
    }
  };

  const generateAcak = async () => {
    if (!paketDipilih) return;
    try {
      const res = await apiClient.post<{ ditambahkan: number; peringatan?: string }>(
        `/api/v1/paket/${paketDipilih}/generate-acak`,
        { jumlah: jumlahAcak },
      );
      setPesan(`Generate acak: ${res.ditambahkan} soal ditambahkan.${res.peringatan ? " " + res.peringatan : ""}`);
      qc.invalidateQueries({ queryKey: ["paket-soal", paketDipilih] });
      qc.invalidateQueries({ queryKey: ["paket"] });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Gagal generate acak.");
    }
  };

  const bankMapelSama = useMemo(() => {
    if (!paketAktif) return [];
    return (bankQuery.data?.items ?? []).filter((b) => b.mapel_id === paketAktif.mapel_id);
  }, [bankQuery.data, paketAktif]);

  if (tampilan === "form") {
    return (
      <div>
        <PaketFormPage
          mode={editId ? "ubah" : "buat"}
          form={form}
          setForm={setForm}
          mapelList={mapelList}
          mapelError={mapelQuery.isError}
          error={error}
          onSubmit={simpanPaket}
          onBatal={kembaliKeDaftar}
        />
      </div>
    );
  }

  if (tampilan === "detail" && paketAktif) {
    return (
      <div>
        <PaketDetailPage
          paket={paketAktif}
          soalPaket={soalPaketQuery.data?.items ?? []}
          bankMapelSama={bankMapelSama}
          bankDipilih={bankDipilih}
          setBankDipilih={setBankDipilih}
          soalBank={soalBankQuery.data?.items ?? []}
          jumlahAcak={jumlahAcak}
          setJumlahAcak={setJumlahAcak}
          pesan={pesan}
          error={error}
          onTambahSoal={tambahSoal}
          onHapusSoal={hapusSoal}
          onGenerateAcak={generateAcak}
          onUbah={() => bukaFormUbah(paketAktif)}
          onKembali={kembaliKeDaftar}
        />
      </div>
    );
  }

  function konfirmasiHapusPaket(p: { id: string }) {
    Modal.confirm({
      title: "Hapus paket?",
      content: "Paket ini akan dihapus permanen beserta susunan soalnya.",
      okText: "Ya, Hapus",
      cancelText: "Batal",
      okButtonProps: { danger: true },
      onOk: () => void hapusPaket(p.id),
    });
  }

  return (
    <div className="space-y-4">
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: "Paket Ujian" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>Paket Ujian</Typography.Title>
          <Typography.Text type="secondary">Kelola paket ujian: buat paket, susun soal dari bank soal, dan hasilkan soal acak.</Typography.Text>
        </div>
        <Space>
          <Button onClick={bukaFormBaru} type="primary">
            <Plus size={16} /> Buat Paket
          </Button>
        </Space>
      </div>
      {pesan && <Alert type="success" showIcon message={<>{pesan}</>} />}
      {error && <Alert type="error" showIcon message={<>{error}</>} />}

      <Input.Search value={cari} onChange={(e) => setCari(e.target.value)} placeholder="Cari kode / nama paket / mapel…" aria-label="Cari paket" size="large" allowClear />
      <p className="text-sm text-muted">
        Menampilkan {tampil.length} dari {semuaPaket.length} paket.
      </p>

      {paketQuery.isLoading ? (
        <p className="text-muted">Memuat paket…</p>
      ) : tampil.length === 0 ? (
        <Empty
          description={
            <>
              <div>{semuaPaket.length === 0 ? "Belum ada paket" : "Tidak ada paket yang cocok"}</div>
              <div>
                {semuaPaket.length === 0
                  ? "Buat paket pertama, lalu susun soalnya dari bank soal."
                  : "Ubah kata kunci pencarian untuk melihat paket lain."}
              </div>
            </>
          }
        >
          {semuaPaket.length === 0 ? (
            <Button type="primary" onClick={bukaFormBaru}>
              Buat Paket
            </Button>
          ) : null}
        </Empty>
      ) : (
          <Table columns={[{ key: "col0", title: "Kode", dataIndex: "kode", sorter: (a: any, b: any) => { const _va = a.kode, _vb = b.kode; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, p: any) => (<>{p.kode}</>) }, { key: "col1", title: "Nama", render: (_: unknown, p: any) => (<>
                    <Button onClick={() => bukaDetail(p.id)} type="default">
                      {p.nama}
                    </Button>
                  </>) }, { key: "col2", title: "Mapel", dataIndex: "mapel_nama", sorter: (a: any, b: any) => { const _va = a.mapel_nama, _vb = b.mapel_nama; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, p: any) => (<>{p.mapel_nama ?? "-"}</>) }, { key: "col3", title: "Jenis", dataIndex: "jenis", sorter: (a: any, b: any) => { const _va = a.jenis, _vb = b.jenis; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, p: any) => (<>{p.jenis}</>) }, { key: "col4", title: "Durasi", render: (_: unknown, p: any) => (<>{p.durasi_menit} mnt</>) }, { key: "col5", title: "Jumlah Soal", render: (_: unknown, p: any) => (<>{p.jumlah_soal_tersusun ?? p.jumlah_soal} soal</>) }, { key: "col6", title: "Aksi", render: (_: unknown, p: any) => (<>
                    <div className="flex flex-wrap gap-1">
                      <Button onClick={() => bukaDetail(p.id)} type="default">
                        <Eye size={15} /> Lihat
                      </Button>
                      <Button aria-label={`Ubah ${p.nama}`} onClick={() => bukaFormUbah(p)} type="default">
                        <Pencil size={15} /> Ubah
                      </Button>
                      <Button aria-label={`Hapus ${p.nama}`} onClick={() => konfirmasiHapusPaket(p)} type="default" danger>
                        <Trash2 size={16} />
                      </Button>
                    </div>
                  </>) }]} dataSource={tampil} rowKey="id" pagination={false} locale={{ emptyText: "Belum ada data" }} scroll={{ x: "max-content" }} />
      )}
    </div>
  );
}
