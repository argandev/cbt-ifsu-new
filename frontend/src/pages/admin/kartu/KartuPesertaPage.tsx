// KartuPesertaPage — daftar + unduh Kartu Peserta (model 1-KARTU).
//
// Model baru: 1 kartu per siswa berlaku untuk SEMUA ujian (bukan 1 kartu
// per jadwal). Kartu diunduh sebagai PDF dari backend:
//   GET /api/v1/kartu/siswa/:siswaId/pdf   (PDF satu siswa)
//   GET /api/v1/kartu/pdf?kelas_id=&event_id= (PDF gabungan, staf)
// File ini TIDAK memakai cetak browser — semua unduhan lewat tautan
// ke PDF backend (tombol antd dengan href).
//
// Daftar siswa dibaca dari GET /api/v1/master/siswa (filter kelas + cari),
// difilter event/kelas untuk PDF gabungan. Nomor peserta tetap unik PER
// JADWAL: <kode-event atau US>/<tahun>/<urut 3 digit> — tercantum per
// baris jadwal di dalam kartu PDF. Kartu TIDAK berisi token ujian.
//
// Penugasan event & ruangan per jadwal TETAP ADA di seksi tersendiri
// di bawah (PUT /api/v1/jadwal/:id/event-ruangan); penugasan tersimpan
// dibaca lewat GET /api/v1/kartu/:jadwalId (JSON lama, hanya untuk
// prefill — bukan untuk unduhan). TanStack Query + apiClient + antd.
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Download } from "lucide-react";
import { apiClient } from "../../../lib/api";
import { formatJam, formatTanggal } from "../../../lib/format";

import { Alert, Button, Card, Input, Pagination, Select, Table, Breadcrumb, Space, Typography } from "antd";

type Siswa = {
  id: string;
  nisn: string;
  nis: string | null;
  nama: string;
  kelas_id: string;
  jurusan_id: string;
  tahun_masuk: number | null;
  is_active: boolean;
};
type Kelas = { id: string; jurusan_id: string; tingkat: number; nama: string; tahun_ajaran: string };
type Jurusan = { id: string; kode: string; nama: string };
type Meta = { page: number; per_page: number; total: number; total_pages: number };
type ListRes<T> = { data: T[]; meta?: Meta };

type JadwalItem = {
  id: string;
  paket_nama?: string;
  paket_kode?: string;
  kelas_nama?: string;
  tanggal: string;
  jam_mulai: string;
  jam_selesai: string;
  status: string;
};

type EventItem = { id: string; kode: string; nama: string };
type RuanganItem = { id: string; kode: string; nama: string; kapasitas: number; lokasi: string | null };

/** Bentuk ringkas GET /api/v1/kartu/:jadwalId — hanya penugasan jadwal. */
type PenugasanJadwalRes = {
  jadwal: {
    event: { id: string | null; kode: string | null; nama: string | null } | null;
    ruangan: { id: string | null; nama: string | null; lokasi: string | null } | null;
  };
};

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? "";
const PER_PAGE = 20;

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

function pdfSiswaUrl(siswaId: string): string {
  return `${API_BASE}/api/v1/kartu/siswa/${siswaId}/pdf`;
}

function pdfGabunganUrl(kelasId: string, eventId: string): string {
  const params = new URLSearchParams();
  if (kelasId) params.set("kelas_id", kelasId);
  if (eventId) params.set("event_id", eventId);
  const query = params.toString();
  return `${API_BASE}/api/v1/kartu/pdf${query ? `?${query}` : ""}`;
}

export default function KartuPesertaPage() {
  const qc = useQueryClient();

  // --- Filter daftar siswa + PDF gabungan ---
  const [filterKelasId, setFilterKelasId] = useState("");
  const [filterEventId, setFilterEventId] = useState("");
  const [q, setQ] = useState("");
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);

  // --- Penugasan event & ruangan per jadwal (seksi tersendiri) ---
  const [jadwalId, setJadwalId] = useState("");
  const [assignEventId, setAssignEventId] = useState("");
  const [assignRuanganId, setAssignRuanganId] = useState("");
  const [assignMsg, setAssignMsg] = useState("");
  const [assignError, setAssignError] = useState("");

  useEffect(() => {
    const t = setTimeout(() => {
      setSearch(q);
      setPage(1);
    }, 300);
    return () => clearTimeout(t);
  }, [q]);

  const kelasQuery = useQuery({
    queryKey: ["master", "kelas", "all"],
    queryFn: () => apiClient.get<ListRes<Kelas> | Kelas[]>("/api/v1/master/kelas?per_page=100"),
  });
  const kelasList = unwrapList(kelasQuery.data ?? []).items;
  const kelasNama = (id: string) => kelasList.find((k) => k.id === id)?.nama ?? "-";

  const jurusanQuery = useQuery({
    queryKey: ["master", "jurusan", "all"],
    queryFn: () => apiClient.get<ListRes<Jurusan> | Jurusan[]>("/api/v1/master/jurusan?per_page=100"),
  });
  const jurusanList = unwrapList(jurusanQuery.data ?? []).items;
  const jurusanNama = (id: string) => jurusanList.find((j) => j.id === id)?.nama ?? "-";

  const eventQuery = useQuery({
    queryKey: ["event"],
    queryFn: () => apiClient.get<{ items: EventItem[] }>("/api/v1/event?limit=100"),
  });
  const ruanganQuery = useQuery({
    queryKey: ["ruangan"],
    queryFn: () => apiClient.get<{ items: RuanganItem[] }>("/api/v1/ruangan?limit=100"),
  });

  const siswaQuery = useQuery({
    queryKey: ["master", "siswa", "kartu", search, page, filterKelasId],
    queryFn: () =>
      apiClient.get<ListRes<Siswa> | Siswa[]>(
        `/api/v1/master/siswa?page=${page}&per_page=${PER_PAGE}&q=${encodeURIComponent(search)}` +
          (filterKelasId ? `&kelas_id=${filterKelasId}` : ""),
      ),
  });
  const { items: siswaItems, meta } = unwrapList(siswaQuery.data ?? []);

  // --- Penugasan: daftar jadwal + prefill dari penugasan tersimpan ---
  const jadwalQuery = useQuery({
    queryKey: ["jadwal"],
    queryFn: () => apiClient.get<{ items: JadwalItem[] }>("/api/v1/jadwal"),
  });

  const penugasanQuery = useQuery({
    queryKey: ["kartu-penugasan", jadwalId],
    queryFn: () => apiClient.get<PenugasanJadwalRes>(`/api/v1/kartu/${jadwalId}`),
    enabled: jadwalId !== "",
  });

  // Sinkronkan pilihan event/ruangan dengan penugasan jadwal yang tersimpan.
  useEffect(() => {
    setAssignEventId(penugasanQuery.data?.jadwal.event?.id ?? "");
    setAssignRuanganId(penugasanQuery.data?.jadwal.ruangan?.id ?? "");
  }, [penugasanQuery.data?.jadwal.event?.id, penugasanQuery.data?.jadwal.ruangan?.id]);

  const assignMutation = useMutation({
    mutationFn: () =>
      apiClient.put(`/api/v1/jadwal/${jadwalId}/event-ruangan`, {
        event_id: assignEventId || null,
        ruangan_id: assignRuanganId || null,
      }),
    onSuccess: () => {
      setAssignMsg("Event dan ruangan berhasil ditugaskan ke jadwal.");
      setAssignError("");
      qc.invalidateQueries({ queryKey: ["kartu-penugasan", jadwalId] });
    },
    onError: (err) => {
      setAssignError(apiErrorMessage(err));
      setAssignMsg("");
    },
  });

  const jadwalOptions = (jadwalQuery.data?.items ?? []).map((j) => ({
    value: j.id,
    label: `${formatTanggal(j.tanggal)} ${formatJam(j.jam_mulai)}–${formatJam(j.jam_selesai)} · ${j.paket_nama ?? j.paket_kode ?? "Paket"} · ${j.kelas_nama ?? ""} (${j.status})`,
  }));
  const eventOptions = [
    { value: "", label: "Tanpa event / Semua event" },
    ...(eventQuery.data?.items ?? []).map((ev) => ({ value: ev.id, label: `${ev.kode} — ${ev.nama}` })),
  ];
  const kelasOptions = [
    { value: "", label: "Semua kelas" },
    ...kelasList.map((k) => ({ value: k.id, label: k.nama })),
  ];
  const ruanganOptions = [
    { value: "", label: "Tanpa ruangan" },
    ...(ruanganQuery.data?.items ?? []).map((r) => ({
      value: r.id,
      label: `${r.nama} (${r.kapasitas} kursi${r.lokasi ? ` · ${r.lokasi}` : ""})`,
    })),
  ];
  const assignEventOptions = [
    { value: "", label: "Tanpa event" },
    ...(eventQuery.data?.items ?? []).map((ev) => ({ value: ev.id, label: `${ev.kode} — ${ev.nama}` })),
  ];

  return (
    <div className="space-y-4">
      <Breadcrumb items={[{ title: <Link to="/adm">Dashboard</Link> }, { title: "Kartu Peserta" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>
            Kartu Peserta
          </Typography.Title>
          <Typography.Text type="secondary">
            Satu kartu per siswa untuk semua ujian. Setiap baris jadwal di kartu mencantumkan nomor peserta,
            mata pelajaran, tanggal, jam, dan ruangan jadwal tersebut. Kartu tidak berisi token ujian — token
            diberikan proktor saat ujian.
          </Typography.Text>
        </div>
        <Space>
          <Button
            type="primary"
            href={pdfGabunganUrl(filterKelasId, filterEventId)}
            target="_blank"
            rel="noreferrer"
            icon={<Download size={16} />}
          >
            Unduh Semua (PDF)
          </Button>
        </Space>
      </div>

      <Card title="Daftar Siswa & Unduh Kartu">
        <div className="flex flex-col gap-3">
          <div className="grid grid-cols-1 gap-3 md:grid-cols-3">
            <div>
              <Typography.Text strong className="mb-1 block">
                Event
              </Typography.Text>
              <Select
                showSearch
                options={eventOptions}
                value={filterEventId}
                onChange={(v) => setFilterEventId(v ?? "")}
                placeholder="Pilih event…"
                filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())}
                optionFilterProp="label"
                notFoundContent="Tidak ada data"
                style={{ width: "100%" }}
                size="large"
              />
            </div>
            <div>
              <Typography.Text strong className="mb-1 block">
                Kelas
              </Typography.Text>
              <Select
                showSearch
                options={kelasOptions}
                value={filterKelasId}
                onChange={(v) => {
                  setFilterKelasId(v ?? "");
                  setPage(1);
                }}
                placeholder="Pilih kelas…"
                filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())}
                optionFilterProp="label"
                notFoundContent="Tidak ada data"
                style={{ width: "100%" }}
                size="large"
              />
            </div>
            <div>
              <Typography.Text strong className="mb-1 block">
                Cari siswa
              </Typography.Text>
              <Input
                value={q}
                onChange={(e) => setQ(e.target.value)}
                placeholder="Cari nama / NISN…"
                size="large"
                allowClear
              />
            </div>
          </div>
          <p className="text-sm text-muted">
            Filter event dipakai untuk PDF gabungan (“Unduh Semua”). Filter kelas + pencarian dipakai untuk
            daftar siswa di bawah dan untuk PDF gabungan.
          </p>

          {siswaQuery.isError && (
            <Alert type="error" showIcon message={`Gagal memuat siswa: ${apiErrorMessage(siswaQuery.error)}`} />
          )}

          <Table
            columns={[
              {
                key: "nama",
                title: "Nama",
                dataIndex: "nama",
                sorter: (a: Siswa, b: Siswa) => a.nama.localeCompare(b.nama, "id"),
              },
              { key: "nisn", title: "NISN", dataIndex: "nisn" },
              {
                key: "kelas",
                title: "Kelas",
                render: (_: unknown, s: Siswa) => kelasNama(s.kelas_id),
              },
              {
                key: "jurusan",
                title: "Jurusan",
                render: (_: unknown, s: Siswa) => jurusanNama(s.jurusan_id),
              },
              {
                key: "aksi",
                title: "Aksi",
                render: (_: unknown, s: Siswa) => (
                  <Button
                    href={pdfSiswaUrl(s.id)}
                    target="_blank"
                    rel="noreferrer"
                    icon={<Download size={14} />}
                  >
                    Unduh PDF
                  </Button>
                ),
              },
            ]}
            dataSource={siswaItems}
            rowKey="id"
            pagination={false}
            loading={siswaQuery.isPending}
            locale={{ emptyText: "Tidak ada siswa yang cocok dengan filter." }}
            scroll={{ x: "max-content" }}
          />
          {meta && meta.total > 0 && (
            <Pagination
              current={meta.page}
              pageSize={meta.per_page}
              total={meta.total}
              onChange={(p) => setPage(p)}
              showTotal={(total) => `Total ${total} siswa`}
            />
          )}
        </div>
      </Card>

      <Card title="Penugasan Event & Ruangan per Jadwal">
        <div className="flex flex-col gap-3">
          <p className="text-sm text-muted">
            Event dan ruangan ditugaskan per jadwal ujian (dipakai untuk nomor peserta dan ruangan yang
            tercantum di kartu). Pilih jadwal, lalu pilih event dan ruangan, lalu simpan.
          </p>
          {assignMsg && <Alert type="success" showIcon message={assignMsg} />}
          {assignError && <Alert type="error" showIcon message={assignError} />}
          {jadwalQuery.isError && (
            <Alert type="error" showIcon message={`Gagal memuat jadwal: ${apiErrorMessage(jadwalQuery.error)}`} />
          )}
          <div>
            <Typography.Text strong className="mb-1 block">
              Pilih jadwal
            </Typography.Text>
            <Select
              showSearch
              allowClear
              options={[{ value: "", label: "— Pilih jadwal —" }, ...jadwalOptions]}
              value={jadwalId}
              onChange={(v) => {
                setJadwalId(v ?? "");
                setAssignMsg("");
                setAssignError("");
              }}
              placeholder=""
              filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())}
              optionFilterProp="label"
              notFoundContent="Tidak ada data"
              style={{ width: "100%" }}
              size="large"
            />
          </div>
          {jadwalId && (
            <>
              <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
                <div>
                  <Typography.Text strong className="mb-1 block">
                    Event
                  </Typography.Text>
                  <Select
                    showSearch
                    options={assignEventOptions}
                    value={assignEventId}
                    onChange={(v) => setAssignEventId(v ?? "")}
                    placeholder="Pilih…"
                    filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())}
                    optionFilterProp="label"
                    notFoundContent="Tidak ada data"
                    style={{ width: "100%" }}
                    size="large"
                  />
                </div>
                <div>
                  <Typography.Text strong className="mb-1 block">
                    Ruangan
                  </Typography.Text>
                  <Select
                    showSearch
                    options={ruanganOptions}
                    value={assignRuanganId}
                    onChange={(v) => setAssignRuanganId(v ?? "")}
                    placeholder="Pilih…"
                    filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())}
                    optionFilterProp="label"
                    notFoundContent="Tidak ada data"
                    style={{ width: "100%" }}
                    size="large"
                  />
                </div>
              </div>
              <p className="text-sm text-muted">
                Kapasitas ruangan harus cukup untuk seluruh siswa kelas jadwal; bila kurang, server menolak dengan
                pesan yang jelas.
              </p>
              <div>
                <Button onClick={() => assignMutation.mutate()} loading={assignMutation.isPending} type="primary">
                  {assignMutation.isPending ? "Menyimpan…" : "Simpan Penugasan"}
                </Button>
              </div>
            </>
          )}
        </div>
      </Card>
    </div>
  );
}
