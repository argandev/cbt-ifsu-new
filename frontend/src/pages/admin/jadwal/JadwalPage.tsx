// JadwalPage — daftar jadwal (halaman utama) + form buat/ubah jadwal di
// halaman terpisah (Tim B, penataan — audit §4.2).
// SEBELUMNYA form + daftar berdempetan dalam satu grid dua kolom;
// SEKARANG daftar adalah tabel penuh, form tampil sendiri.
// Token TIDAK lagi di-generate dari kartu daftar: tombol "Token" membuka
// Dialog Token (generate → tampil sekali + Salin + info kedaluwarsa).
// Logika TIDAK berubah: endpoint, payload, token tampil-sekali, pause/resume,
// status token, dan hapus via ConfirmDialog sama seperti sebelumnya.
// Pengawas TIDAK melihat token: halaman ini untuk admin/proktor/guru.
import { useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Copy, KeyRound, Pause, Play, Pencil, Trash2, Plus, Monitor as MonitorIcon } from "lucide-react";
import { apiClient } from "../../../lib/api";
import { useAuthStore } from "../../../store/auth";
import { formatJam, formatTanggal, formatTanggalJam } from "../../../lib/format";
import { statusTagColor, statusTagLabel } from "../../../components/common/statusTag";
import JadwalFormPage, { type JadwalFormValues } from "./JadwalFormPage";
import type { Jadwal, Paket } from "../../../lib/exam";

import { Alert, Button, Empty, Input, Modal, Table, Tag, Breadcrumb, Space, Typography } from "antd";
interface Kelas {
  id: string;
  nama: string;
  tingkat: number;
  jurusan_id: string;
  tahun_ajaran: string;
}

const formKosong: JadwalFormValues = {
  paket_id: "",
  kelas_id: "",
  tanggal: "",
  jam_mulai: "",
  jam_selesai: "",
  status: "draft",
};

type FilterStatus = "semua" | "draft" | "aktif" | "pause" | "selesai";
const FILTER_STATUS: Array<{ value: FilterStatus; label: string }> = [
  { value: "semua", label: "Semua" },
  { value: "draft", label: "Draft" },
  { value: "aktif", label: "Aktif" },
  { value: "pause", label: "Jeda" },
  { value: "selesai", label: "Selesai" },
];

export default function JadwalPage() {
  const { hasPermission } = useAuthStore();
  const qc = useQueryClient();
  const [tampilan, setTampilan] = useState<"daftar" | "form">("daftar");
  const [form, setForm] = useState<JadwalFormValues>(formKosong);
  const [editId, setEditId] = useState<string | null>(null);
  const [tokenSekali, setTokenSekali] = useState<Record<string, { token: string; kedaluwarsa: string }>>({});
  const [tokenStatus, setTokenStatus] = useState<Record<string, any>>({});
  const [pesan, setPesan] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [disalin, setDisalin] = useState<string | null>(null);
  const [dialogTokenId, setDialogTokenId] = useState<string | null>(null);
  const [cari, setCari] = useState("");
  const [filterStatus, setFilterStatus] = useState<FilterStatus>("semua");

  const jadwalQuery = useQuery({
    queryKey: ["jadwal"],
    queryFn: () => apiClient.get<{ items: Jadwal[] }>("/api/v1/jadwal"),
  });
  const paketQuery = useQuery({
    queryKey: ["paket"],
    queryFn: () => apiClient.get<{ items: Paket[] }>("/api/v1/paket"),
    // Gate dengan izin ENDPOINT-nya (paket.read), bukan jadwal.create:
    // proktor punya jadwal.create tapi tidak punya paket.read — gate lama
    // membuatnya tetap menembak GET /paket → 403. Dropdown paket proktor
    // tetap kosong (D2: butuh keputusan backend), tapi tidak ada lagi 403.
    enabled: hasPermission("paket.read"),
  });
  // Daftar kelas untuk dropdown — endpoint yang sama dipakai SiswaPage.
  const kelasQuery = useQuery({
    queryKey: ["master", "kelas", "all"],
    queryFn: () => apiClient.get<{ data: Kelas[] } | Kelas[]>("/api/v1/master/kelas?per_page=100"),
    // Gate dengan izin endpoint-nya (master.kelas.read).
    enabled: hasPermission("master.kelas.read"),
  });
  const kelasList: Kelas[] = Array.isArray(kelasQuery.data)
    ? kelasQuery.data
    : (kelasQuery.data?.data ?? []);

  const semuaJadwal = jadwalQuery.data?.items ?? [];
  const tampil = useMemo(() => {
    const q = cari.trim().toLowerCase();
    return semuaJadwal.filter((j) => {
      if (filterStatus !== "semua" && j.status !== filterStatus) return false;
      if (q && !(j.paket_nama ?? j.paket_id).toLowerCase().includes(q)) return false;
      return true;
    });
  }, [semuaJadwal, cari, filterStatus]);

  const refresh = () => qc.invalidateQueries({ queryKey: ["jadwal"] });

  const bukaFormBaru = () => {
    setEditId(null);
    setForm(formKosong);
    setError(null);
    setPesan(null);
    setTampilan("form");
  };

  const bukaFormUbah = (j: Jadwal) => {
    setEditId(j.id);
    setForm({
      paket_id: j.paket_id, kelas_id: j.kelas_id, tanggal: j.tanggal,
      jam_mulai: j.jam_mulai, jam_selesai: j.jam_selesai, status: j.status,
    });
    setError(null);
    setPesan(null);
    setTampilan("form");
  };

  const batalForm = () => {
    setEditId(null);
    setForm(formKosong);
    setError(null);
    setTampilan("daftar");
  };

  const simpan = async () => {
    setError(null);
    setPesan(null);
    // Pengganti `required` Dropdown lama (hidden input native): antd Select
    // tidak merender input required, jadi blokir di handler dengan pesan sama.
    if (!form.paket_id) {
      setError("Paket wajib dipilih.");
      return;
    }
    try {
      if (editId) {
        await apiClient.put(`/api/v1/jadwal/${editId}`, form);
        setPesan("Jadwal diubah.");
      } else {
        await apiClient.post("/api/v1/jadwal", form);
        setPesan("Jadwal dibuat.");
      }
      setForm(formKosong);
      setEditId(null);
      setTampilan("daftar");
      refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Gagal menyimpan jadwal.");
    }
  };

  const hapus = async (id: string) => {
    try {
      await apiClient.del(`/api/v1/jadwal/${id}`);
      refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Gagal menghapus jadwal.");
    }
  };

  const aksi = async (id: string, nama: "pause" | "resume") => {
    setError(null);
    try {
      await apiClient.post(`/api/v1/jadwal/${id}/${nama}`, {});
      setPesan(`Jadwal ${nama === "pause" ? "dijeda (timer dibekukan)" : "dilanjutkan"}.`);
      refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : `Gagal ${nama} jadwal.`);
    }
  };

  const generateToken = async (id: string) => {
    setError(null);
    try {
      const res = await apiClient.post<{ token: string; token_expires_at: string }>(
        `/api/v1/jadwal/${id}/token`,
        {},
      );
      // Token plaintext hanya ada di respons ini — simpan di state lokal
      // halaman, JANGAN pernah disimpan ke localStorage / log.
      setTokenSekali((prev) => ({ ...prev, [id]: { token: res.token, kedaluwarsa: res.token_expires_at } }));
      setPesan("Token baru dibuat — token lama (bila ada) sudah hangus. Salin sekarang.");
      refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Gagal membuat token.");
    }
  };

  const cekTokenStatus = async (id: string) => {
    try {
      const res = await apiClient.get<any>(`/api/v1/jadwal/${id}/token-status`);
      setTokenStatus((prev) => ({ ...prev, [id]: res }));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Gagal membaca status token.");
    }
  };

  const salin = async (id: string, token: string) => {
    try {
      await navigator.clipboard.writeText(token);
      setDisalin(id);
      setTimeout(() => setDisalin((d) => (d === id ? null : d)), 2000);
    } catch {
      setError("Gagal menyalin — salin manual dari layar.");
    }
  };

  const jadwalDialog = dialogTokenId ? semuaJadwal.find((j) => j.id === dialogTokenId) : undefined;
  const tokenStatusDialog = dialogTokenId ? tokenStatus[dialogTokenId] : undefined;
  const tokenSekaliDialog = dialogTokenId ? tokenSekali[dialogTokenId] : undefined;

  if (tampilan === "form") {
    return (
      <div>
        <JadwalFormPage
          mode={editId ? "ubah" : "buat"}
          form={form}
          setForm={setForm}
          paketList={paketQuery.data?.items ?? []}
          kelasList={kelasList}
          kelasError={kelasQuery.isError}
          error={error}
          onSubmit={simpan}
          onBatal={batalForm}
        />
      </div>
    );
  }

  function konfirmasiHapusJadwal(j: { id: string }) {
    Modal.confirm({
      title: "Hapus jadwal?",
      content: "Jadwal ini akan dihapus permanen.",
      okText: "Ya, Hapus",
      cancelText: "Batal",
      okButtonProps: { danger: true },
      onOk: () => void hapus(j.id),
    });
  }

  return (
    <div className="space-y-4">
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: "Jadwal" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>Jadwal Ujian</Typography.Title>
          <Typography.Text type="secondary">Kelola jadwal ujian: buat jadwal, hasilkan token, jeda/lanjutkan, dan pantau lewat Monitor.</Typography.Text>
        </div>
        <Space>
          {/* D2-koreksi (UI-ONLY): tombol Buat Jadwal tampil hanya bila
              punya paket.read DAN jadwal.create. Proktor punya
              jadwal.create tapi TIDAK punya paket.read — form buat jadwal
              butuh dropdown Paket yang endpoint-nya memang ditolak untuk
              proktor, jadi tombolnya disembunyikan untuk proktor (bukan
              endpointnya yang diubah). Pause/Resume/Monitor tidak diganggu. */}
          {hasPermission("paket.read") && hasPermission("jadwal.create") && (
            <Button onClick={bukaFormBaru} type="primary">
              <Plus size={16} /> Buat Jadwal
            </Button>
          )}
        </Space>
      </div>
      {pesan && <Alert type="success" showIcon message={<>{pesan}</>} />}
      {error && <Alert type="error" showIcon message={<>{error}</>} />}

      {/* Filter: cari + status (satu bar filter di atas tabel). */}
      <div className="flex flex-col gap-2">
        <Input.Search value={cari} onChange={(e) => setCari(e.target.value)} placeholder="Cari nama paket…" aria-label="Cari jadwal" size="large" allowClear />
        <div className="flex flex-wrap gap-2" role="group" aria-label="Filter status jadwal">
          {FILTER_STATUS.map((f) => (
            <Button key={f.value} onClick={() => setFilterStatus(f.value)} aria-pressed={filterStatus === f.value} type="primary">
              {f.label}
            </Button>
          ))}
        </div>
      </div>
      <p className="text-sm text-muted">
        Menampilkan {tampil.length} dari {semuaJadwal.length} jadwal.
      </p>

      {jadwalQuery.isLoading ? (
        <p className="text-muted">Memuat jadwal…</p>
      ) : tampil.length === 0 ? (
        <Empty
          description={
            <>
              <div>{semuaJadwal.length === 0 ? "Belum ada jadwal" : "Tidak ada jadwal yang cocok"}</div>
              <div>
                {semuaJadwal.length === 0
                  ? "Buat jadwal pertama untuk paket ujian yang sudah ada."
                  : "Ubah kata kunci atau filter status untuk melihat jadwal lain."}
              </div>
            </>
          }
        >
          {semuaJadwal.length === 0 && hasPermission("paket.read") && hasPermission("jadwal.create") ? (
            <Button type="primary" onClick={bukaFormBaru}>
              Buat Jadwal
            </Button>
          ) : null}
        </Empty>
      ) : (
          <Table columns={[{ key: "col0", title: "Paket", dataIndex: "paket_nama", sorter: (a: any, b: any) => { const _va = a.paket_nama, _vb = b.paket_nama; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, j: any) => (<>{j.paket_nama ?? j.paket_id}</>) }, { key: "col1", title: "Kelas", dataIndex: "kelas_nama", sorter: (a: any, b: any) => { const _va = a.kelas_nama, _vb = b.kelas_nama; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, j: any) => (<>{j.kelas_nama ?? j.kelas_id}</>) }, { key: "col2", title: "Tanggal", render: (_: unknown, j: any) => (<>{formatTanggal(j.tanggal)}</>) }, { key: "col3", title: "Jam", render: (_: unknown, j: any) => (<>{formatJam(j.jam_mulai)}–{formatJam(j.jam_selesai)}</>) }, { key: "col4", title: "Status", render: (_: unknown, j: any) => (<><Tag color={statusTagColor(j.status)}>{statusTagLabel(j.status)}</Tag></>) }, { key: "col5", title: "Token", render: (_: unknown, j: any) => (<>
                      {tokenStatus[j.id]
                        ? tokenStatus[j.id].ada_token
                          ? tokenStatus[j.id].kedaluwarsa
                            ? "Kedaluwarsa"
                            : `Aktif, kedaluwarsa dalam ${tokenStatus[j.id].kedaluwarsa_dalam_detik ?? "?"} detik`
                          : "Belum dibuat"
                        : "Belum dicek"}
                    </>) }, { key: "col6", title: "Aksi", render: (_: unknown, j: any) => (<>
                      <div className="flex flex-wrap gap-1">
                        <Link
                          to={`/monitor/${j.id}`}
                          className="flex min-h-[44px] items-center gap-1 rounded-lg border border-anbk px-2 text-sm font-semibold text-anbk"
                        >
                          <MonitorIcon size={15} /> Monitor
                        </Link>
                        {/* proxy: tidak ada permission bernama untuk token di backend;
                            jadwal.pause dipakai sebagai proxy gate Token. */}
                        {hasPermission("jadwal.pause") && (
                          <Button onClick={() => setDialogTokenId(j.id)} type="default">
                            <KeyRound size={15} /> Token
                          </Button>
                        )}
                        {hasPermission("jadwal.update") && (
                          <Button aria-label={`Ubah jadwal ${j.paket_nama}`} onClick={() => bukaFormUbah(j)} type="default">
                            <Pencil size={15} /> Ubah
                          </Button>
                        )}
                        {j.status === "aktif" && hasPermission("jadwal.pause") && (
                          <Button onClick={() => aksi(j.id, "pause")} type="default">
                            <Pause size={15} /> Pause
                          </Button>
                        )}
                        {j.status === "pause" && hasPermission("jadwal.resume") && (
                          <Button onClick={() => aksi(j.id, "resume")} type="default">
                            <Play size={15} /> Resume
                          </Button>
                        )}
                        {hasPermission("jadwal.delete") && (
                          <Button aria-label={`Hapus jadwal ${j.paket_nama}`} onClick={() => konfirmasiHapusJadwal(j)} type="default" danger>
                            <Trash2 size={16} />
                          </Button>
                        )}
                      </div>
                    </>) }]} dataSource={tampil} rowKey="id" pagination={false} locale={{ emptyText: "Belum ada data" }} scroll={{ x: "max-content" }} />
      )}

      {/* Dialog Token — token hanya tampil sekali di sini, tidak di daftar. */}
      <Modal open={dialogTokenId !== null} onCancel={() => setDialogTokenId(null)} title={`Token Ujian — ${jadwalDialog?.paket_nama ?? "Jadwal"}`} footer={null} width={520} destroyOnClose>
        <div className="flex flex-col gap-3">
          {tokenSekaliDialog ? (
            <Alert
              type="info"
              showIcon
              message="Token baru (hanya tampil sekali — token lama hangus):"
              description={
                <>
                  <p className="mt-1 font-mono text-2xl font-bold tracking-widest text-anbk">
                    {tokenSekaliDialog.token}
                  </p>
                  <p className="text-xs text-anbk">
                    Kedaluwarsa: {formatTanggalJam(tokenSekaliDialog.kedaluwarsa)}
                  </p>
                  <Button onClick={() => dialogTokenId && salin(dialogTokenId, tokenSekaliDialog.token)} type="primary">
                    <Copy size={15} /> {disalin === dialogTokenId ? "Tersalin!" : "Salin Token"}
                  </Button>
                </>
              }
            />
          ) : (
            <p className="text-sm text-muted">
              Status token:{" "}
              {tokenStatusDialog
                ? tokenStatusDialog.ada_token
                  ? tokenStatusDialog.kedaluwarsa
                    ? "kedaluwarsa"
                    : `aktif, kedaluwarsa dalam ${tokenStatusDialog.kedaluwarsa_dalam_detik ?? "?"} detik`
                  : "belum dibuat"
                : "belum dicek"}
              {" "}(nilai token tidak pernah ditampilkan di sini)
            </p>
          )}
          <div className="flex flex-wrap gap-2">
            {/* proxy: tidak ada permission bernama untuk token di backend;
                jadwal.pause dipakai sebagai proxy gate Token. */}
            {hasPermission("jadwal.pause") && (
              <Button onClick={() => dialogTokenId && generateToken(dialogTokenId)} type="primary">
                <KeyRound size={15} /> Generate Token
              </Button>
            )}
            {/* proxy: tidak ada permission bernama untuk token di backend;
                jadwal.pause dipakai sebagai proxy gate Token. */}
            {hasPermission("jadwal.pause") && (
              <Button onClick={() => dialogTokenId && cekTokenStatus(dialogTokenId)} type="default">
                Status Token
              </Button>
            )}
          </div>
          <p className="text-xs text-muted">
            Generate token baru membuat token lama (bila ada) hangus.
          </p>
        </div>
      </Modal>

    </div>
  );
}
