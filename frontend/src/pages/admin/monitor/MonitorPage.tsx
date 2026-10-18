// MonitorPage — monitor ujian real-time (Tim 4).
// WebSocket /ws/monitor?ticket= (ticket 30 detik) + fallback polling
// REST tiap 5 detik bila WS gagal. Filter status, tombol pause/resume
// jadwal, unblock peserta (wajib alasan), usul buka blokir oleh
// pengawas (Modal alasan -> POST usulan-unblock), statistik.
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { Pause, Play, Send, Unlock } from "lucide-react";
import { apiClient } from "../../../lib/api";
import { useAuthStore } from "../../../store/auth";
import { formatJam, formatTanggal } from "../../../lib/format";
import type { MonitorSnapshot, PesertaMonitor, UsulanUnblock } from "../../../lib/exam";
import { statusTagColor, statusTagLabel } from "../../../components/common/statusTag";

import { Alert, Button, Form, Input, Modal, Table, Tag, Breadcrumb, Space, Typography, Card, Statistic } from "antd";
type FilterStatus = "semua" | "berjalan" | "selesai" | "diblokir";

function jadwalIdDariUrl(): string {
  return window.location.pathname.match(/\/monitor\/([0-9a-f-]{36})/i)?.[1] ?? "";
}

function formatSisa(detik?: number | null): string {
  if (detik === null || detik === undefined) return "—";
  const m = Math.floor(detik / 60);
  const s = detik % 60;
  return `${m}:${String(s).padStart(2, "0")}`;
}

export default function MonitorPage() {
  const { hasPermission } = useAuthStore();
  const jadwalId = jadwalIdDariUrl();
  const [peserta, setPeserta] = useState<PesertaMonitor[]>([]);
  const [wsAktif, setWsAktif] = useState(false);
  const [filter, setFilter] = useState<FilterStatus>("semua");
  const [alasanPerAttempt, setAlasanPerAttempt] = useState<Record<string, string>>({});
  const [pesan, setPesan] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  // Usul buka blokir (pengawas): target peserta + isi Modal.
  const [usulanTarget, setUsulanTarget] = useState<PesertaMonitor | null>(null);
  const [usulanAlasan, setUsulanAlasan] = useState("");
  const [usulanMengirim, setUsulanMengirim] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);

  // HANYA pengawas yang melihat tombol "Usul Buka Blokir": pengawas
  // adalah satu-satunya role yang punya jadwal.usulan-unblock TANPA
  // jadwal.unblock (proktor/admin unblock langsung, jadi tidak usul).
  const bisaUsulUnblock =
    hasPermission("jadwal.usulan-unblock") && !hasPermission("jadwal.unblock");

  const jadwalQuery = useQuery({
    queryKey: ["jadwal", jadwalId],
    queryFn: () => apiClient.get<any>(`/api/v1/jadwal/${jadwalId}`),
    enabled: !!jadwalId,
  });
  const usulanQuery = useQuery({
    queryKey: ["usulan-unblock", jadwalId],
    queryFn: () => apiClient.get<{ items: UsulanUnblock[] }>(`/api/v1/jadwal/${jadwalId}/usulan-unblock`),
    enabled: !!jadwalId && hasPermission("jadwal.usulan-unblock"),
    refetchInterval: 10000,
  });

  // Attempt yang sudah punya usulan menunggu -> tombol jadi
  // "Usulan Terkirim" (backend juga menolak duplikat dengan 409).
  const usulanMenungguAttempt = useMemo(
    () =>
      new Set(
        (usulanQuery.data?.items ?? [])
          .filter((u) => u.status === "menunggu")
          .map((u) => u.attempt_id),
      ),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [usulanQuery.data],
  );
  // (sebelumnya halaman ini jalan buntu tanpa pemilih jadwal).
  const daftarJadwalQuery = useQuery({
    queryKey: ["jadwal"],
    queryFn: () =>
      apiClient.get<{ items: Array<{ id: string; paket_nama?: string; kelas_nama?: string; tanggal: string; jam_mulai: string; jam_selesai: string; status: string }> }>("/api/v1/jadwal"),
    enabled: !jadwalId,
  });

  // Snapshot REST (dipakai awal + fallback polling bila WS mati).
  const muatSnapshot = useCallback(async () => {
    if (!jadwalId) return;
    try {
      const snap = await apiClient.get<MonitorSnapshot>(`/api/v1/monitor/${jadwalId}`);
      setPeserta(snap.peserta);
    } catch {
      // Abaikan: WS mungkin masih jalan.
    }
  }, [jadwalId]);

  // WS: minta ticket -> sambung -> terima snapshot/update.
  useEffect(() => {
    if (!jadwalId) return;
    let batal = false;
    let reconnectTimer: ReturnType<typeof setTimeout>;
    const sambung = async () => {
      try {
        const t = await apiClient.post<{ ticket: string }>("/api/v1/monitor/ticket", { jadwal_id: jadwalId });
        if (batal) return;
        const proto = window.location.protocol === "https:" ? "wss" : "ws";
        const ws = new WebSocket(`${proto}://${window.location.host}/ws/monitor?ticket=${t.ticket}`);
        wsRef.current = ws;
        ws.onopen = () => setWsAktif(true);
        ws.onmessage = (ev) => {
          try {
            const msg = JSON.parse(ev.data as string);
            if (msg.type === "snapshot" && msg.data?.peserta) {
              setPeserta(msg.data.peserta as PesertaMonitor[]);
            } else if (msg.type === "update") {
              // Update ringan -> muat ulang snapshot REST (murah, jarang).
              muatSnapshot();
            } else if (msg.type === "jadwal_pause" || msg.type === "jadwal_resume") {
              jadwalQuery.refetch();
              muatSnapshot();
            } else if (msg.type === "usulan_unblock") {
              usulanQuery.refetch();
            }
          } catch {
            // Pesan non-JSON (ping) diabaikan.
          }
        };
        ws.onclose = () => {
          setWsAktif(false);
          if (!batal) reconnectTimer = setTimeout(sambung, 5000);
        };
        ws.onerror = () => ws.close();
      } catch {
        setWsAktif(false);
        if (!batal) reconnectTimer = setTimeout(sambung, 5000);
      }
    };
    muatSnapshot();
    sambung();
    return () => {
      batal = true;
      clearTimeout(reconnectTimer);
      wsRef.current?.close();
    };
  }, [jadwalId, muatSnapshot]); // eslint-disable-line react-hooks/exhaustive-deps

  // Fallback polling tiap 5 dtk bila WS tidak aktif.
  useEffect(() => {
    if (wsAktif) return;
    const id = setInterval(muatSnapshot, 5000);
    return () => clearInterval(id);
  }, [wsAktif, muatSnapshot]);

  const tampil = useMemo(
    () => (filter === "semua" ? peserta : peserta.filter((p) => p.status === filter)),
    [peserta, filter],
  );
  const stat = useMemo(() => {
    const total = peserta.length;
    const berjalan = peserta.filter((p) => p.status === "berjalan").length;
    const selesai = peserta.filter((p) => p.status === "selesai").length;
    const diblokir = peserta.filter((p) => p.status === "diblokir").length;
    const totalJawaban = peserta.reduce((a, p) => a + p.dijawab, 0);
    const totalSoal = peserta.reduce((a, p) => a + p.total_soal, 0);
    return { total, berjalan, selesai, diblokir, totalJawaban, totalSoal };
  }, [peserta]);

  const aksiJadwal = async (aksi: "pause" | "resume") => {
    setError(null);
    try {
      await apiClient.post(`/api/v1/jadwal/${jadwalId}/${aksi}`, {});
      setPesan(`Jadwal ${aksi === "pause" ? "dijeda (timer dibekukan)" : "dilanjutkan"}.`);
      jadwalQuery.refetch();
      muatSnapshot();
    } catch (e) {
      setError(e instanceof Error ? e.message : `Gagal ${aksi} jadwal.`);
    }
  };

  const unblock = async (attemptId: string) => {
    const alasan = (alasanPerAttempt[attemptId] ?? "").trim();
    if (!alasan) {
      setError("Alasan unblock wajib diisi.");
      return;
    }
    setError(null);
    try {
      await apiClient.post(`/api/v1/jadwal/${jadwalId}/unblock`, { attempt_id: attemptId, alasan });
      setPesan("Peserta di-unblock — ujian dilanjutkan dari sisa waktu.");
      setAlasanPerAttempt((prev) => ({ ...prev, [attemptId]: "" }));
      muatSnapshot();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Gagal unblock.");
    }
  };

  const kirimUsulanUnblock = async () => {
    if (!usulanTarget) return;
    const alasan = usulanAlasan.trim();
    if (!alasan) {
      setError("Alasan usul buka blokir wajib diisi.");
      return;
    }
    setError(null);
    setUsulanMengirim(true);
    try {
      await apiClient.post(`/api/v1/jadwal/${jadwalId}/usulan-unblock`, {
        attempt_id: usulanTarget.attempt_id,
        alasan,
      });
      setPesan("Usulan buka blokir terkirim — menunggu persetujuan proktor.");
      setUsulanTarget(null);
      setUsulanAlasan("");
      await usulanQuery.refetch();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Gagal mengirim usulan buka blokir.");
    } finally {
      setUsulanMengirim(false);
    }
  };

  const statusJadwal = jadwalQuery.data?.status as string | undefined;

  if (!jadwalId) {
    return (
      <div className="space-y-4">
        <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: "Monitor" }]} />
        <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="min-w-0">
            <Typography.Title level={3} style={{ margin: 0 }}>Monitor Ujian</Typography.Title>
            <Typography.Text type="secondary">Pantau peserta ujian real-time, jeda/lanjutkan jadwal, dan unblock peserta yang diblokir.</Typography.Text>
          </div>
        </div>
        <p className="text-muted">Pilih jadwal yang akan dipantau:</p>
        {daftarJadwalQuery.isLoading && <p className="text-muted">Memuat jadwal…</p>}
        <ul className="space-y-2">
          {(daftarJadwalQuery.data?.items ?? []).map((j) => (
            <li key={j.id}>
              <Link
                to={`/monitor/${j.id}`}
                className="block rounded-lg border border-border bg-white px-3 py-2 text-sm font-medium text-anbk hover:bg-primary-light"
              >
                {j.paket_nama ?? "Paket"} · {j.kelas_nama ?? ""} · {formatTanggal(j.tanggal)} {formatJam(j.jam_mulai)}–{formatJam(j.jam_selesai)} WIB · {j.status}
              </Link>
            </li>
          ))}
        </ul>
        {(daftarJadwalQuery.data?.items ?? []).length === 0 && !daftarJadwalQuery.isLoading && (
          <p className="text-muted">Belum ada jadwal.</p>
        )}
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: "Monitor" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>Monitor Ujian</Typography.Title>
          <Typography.Text type="secondary">Pantau peserta ujian real-time, jeda/lanjutkan jadwal, dan unblock peserta yang diblokir.</Typography.Text>
        </div>
        <Space>
          <Tag color={wsAktif ? "green" : "gold"}>{wsAktif ? "Real-time (WebSocket)" : "Fallback polling 5 dtk"}</Tag>
        </Space>
      </div>
      <p className="mt-1 text-sm text-muted">
        {jadwalQuery.data?.paket_nama ?? ""} · {jadwalQuery.data?.kelas_nama ?? ""} · Status jadwal:{" "}
        <strong>{statusJadwal ?? "…"}</strong>
      </p>
      {pesan && <Alert type="success" showIcon message={<>{pesan}</>} />}
      {error && <Alert type="error" showIcon message={<>{error}</>} />}

      {/* Statistik */}
      <div className="mt-4 grid grid-cols-2 gap-3 sm:grid-cols-4">
        {[
          { label: "Peserta", nilai: stat.total },
          { label: "Sedang Ujian", nilai: stat.berjalan },
          { label: "Selesai", nilai: stat.selesai },
          { label: "Diblokir", nilai: stat.diblokir },
        ].map((s) => (
          <Card key={s.label}>
            <Statistic title={s.label} value={s.nilai} />
          </Card>
        ))}
      </div>
      <p className="mt-2 text-sm text-muted">
        Jawaban terkumpul: {stat.totalJawaban} / {stat.totalSoal || "—"}
      </p>

      {/* Aksi jadwal */}
      <div className="mt-4 flex flex-wrap gap-2">
        {hasPermission("jadwal.pause") && (
          <Button disabled={statusJadwal !== "aktif"} onClick={() => aksiJadwal("pause")} type="default">
            <Pause size={16} /> Pause Jadwal
          </Button>
        )}
        {hasPermission("jadwal.resume") && (
          <Button disabled={statusJadwal !== "pause"} onClick={() => aksiJadwal("resume")} type="default">
            <Play size={16} /> Resume Jadwal
          </Button>
        )}
      </div>

      {/* Usulan unblock dari pengawas */}
      {hasPermission("jadwal.usulan-unblock") && (usulanQuery.data?.items ?? []).filter((u) => u.status === "menunggu").length > 0 && (
        <Card title="Usulan Unblock (menunggu)" className="mt-4">
          <ul className="mt-2 space-y-1 text-sm text-warning">
            {(usulanQuery.data?.items ?? [])
              .filter((u) => u.status === "menunggu")
              .map((u) => (
                <li key={u.id}>
                  {u.nama_siswa} — “{u.alasan}”
                </li>
              ))}
          </ul>
        </Card>
      )}

      {/* Filter */}
      <div className="mt-4 flex flex-wrap gap-2" role="group" aria-label="Filter status">
        {(["semua", "berjalan", "selesai", "diblokir"] as const).map((f) => (
          <Button key={f} onClick={() => setFilter(f)} aria-pressed={filter === f} type="primary">
            {f === "semua" ? "Semua" : f === "berjalan" ? "Sedang Ujian" : f === "selesai" ? "Selesai" : "Diblokir"}
          </Button>
        ))}
      </div>

      {/* Daftar peserta */}
        <Table columns={[{ key: "col0", title: "Siswa", render: (_: unknown, p: any) => (<>
                  <div className="font-semibold">{p.nama_siswa || p.siswa_id}</div>
                  <div className="text-xs text-muted">{p.nisn}</div>
                  {p.kelas_nama && <div className="text-xs text-muted">{p.kelas_nama}</div>}
                </>) }, { key: "col1", title: "Status", render: (_: unknown, p: any) => (<>
                  <Tag color={statusTagColor(p.status)}>{statusTagLabel(p.status)}</Tag>
                </>) }, { key: "col2", title: "Dijawab", render: (_: unknown, p: any) => (<>
                  {p.dijawab} / {p.total_soal}
                </>) }, { key: "col3", title: "Ragu", dataIndex: "ragu", sorter: (a: any, b: any) => { const _va = a.ragu, _vb = b.ragu; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, p: any) => (<>{p.ragu}</>) }, { key: "col4", title: "Pelanggaran", dataIndex: "pelanggaran", sorter: (a: any, b: any) => { const _va = a.pelanggaran, _vb = b.pelanggaran; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, p: any) => (<>{p.pelanggaran}</>) }, { key: "col5", title: "Sisa Waktu", render: (_: unknown, p: any) => (<>{formatSisa(p.sisa_waktu_detik)}</>) }, { key: "col6", title: "Aksi", render: (_: unknown, p: any) => (<>
                  {p.status === "diblokir" && hasPermission("jadwal.unblock") ? (
                    <div className="flex flex-col gap-1 sm:flex-row">
                      <Input type="text" value={alasanPerAttempt[p.attempt_id] ?? ""} onChange={(e) =>
                          setAlasanPerAttempt((prev) => ({ ...prev, [p.attempt_id]: e.target.value }))
                        } placeholder="Alasan unblock (wajib)" aria-label={`Alasan unblock untuk ${p.nama_siswa}`} size="large" />
                      <Button onClick={() => unblock(p.attempt_id)} type="primary">
                        <Unlock size={14} /> Unblock
                      </Button>
                    </div>
                  ) : p.status === "diblokir" && bisaUsulUnblock ? (
                    usulanMenungguAttempt.has(p.attempt_id) ? (
                      <Button disabled type="default">
                        Usulan Terkirim
                      </Button>
                    ) : (
                      <Button onClick={() => { setUsulanTarget(p); setUsulanAlasan(""); }} type="default">
                        <Send size={14} /> Usul Buka Blokir
                      </Button>
                    )
                  ) : (
                    <span className="text-muted">—</span>
                  )}
                </>) }]} dataSource={tampil} rowKey="attempt_id" pagination={false} locale={{ emptyText: "Belum ada peserta." }} scroll={{ x: "max-content" }} />

      {/* Modal usul buka blokir (pengawas) */}
      <Modal
        open={usulanTarget !== null}
        onCancel={() => { setUsulanTarget(null); setUsulanAlasan(""); }}
        title={`Usul Buka Blokir — ${usulanTarget?.nama_siswa ?? ""}`}
        footer={null}
        width={520}
        destroyOnClose
      >
        <Form layout="vertical" onFinish={kirimUsulanUnblock}>
          <Form.Item label="Alasan *" name="alasan" rules={[{ required: true, message: "Alasan usul buka blokir wajib diisi." }]}>
            <Input.TextArea
              value={usulanAlasan}
              onChange={(e) => setUsulanAlasan(e.target.value)}
              placeholder="mis. Siswa tidak sengaja pindah tab saat jaringan putus"
              rows={4}
              size="large"
            />
          </Form.Item>
          <div className="flex flex-col gap-2 sm:flex-row sm:justify-end">
            <Button onClick={() => { setUsulanTarget(null); setUsulanAlasan(""); }} type="default">
              Batal
            </Button>
            <Button htmlType="submit" loading={usulanMengirim} type="primary">
              {usulanMengirim ? "Mengirim…" : "Kirim Usulan"}
            </Button>
          </div>
        </Form>
      </Modal>
    </div>
  );
}
