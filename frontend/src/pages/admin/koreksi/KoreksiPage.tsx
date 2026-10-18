// KoreksiPage — halaman Koreksi Esai / UC10 (pekerja D6).
// Daftar jawaban esai yang menunggu dikoreksi per jadwal + form koreksi
// (skor 0–100 per jawaban, poin = skor/100 × bobot, nilai akhir attempt
// dihitung ulang backend bila attempt sudah selesai).
//
// Izin (backend sumber kebenaran, UI ngikutin izin — JANGAN ubah backend
// untuk ngepasin UI): halaman ini di-gate pada koreksi.update (hanya
// admin/guru). Proktor TIDAK melihat menu Koreksi (gate menu di
// AdminLayout juga koreksi.update) dan bila membuka URL langsung
// dialihkan ke /403 — handler backend listKoreksi memang menolak proktor.
// Filter: Mata Pelajaran (bank soal, via paket jadwal) + Jadwal + tipe
// soal + cari nama/NISN/pertanyaan. Tanpa ubah payload API apa pun.
import { useEffect, useMemo, useState } from "react";
import { Link, Navigate } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import apiClient from "../../../lib/api";
import { formatJam, formatTanggal } from "../../../lib/format";
import { TIPE_SOAL } from "../../../lib/tipeSoal";
import type { Paket } from "../../../lib/exam";
import { useAuthStore } from "../../../store/auth";
import {
  Alert,
  Breadcrumb,
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  Modal,
  Select,
  Statistic,
  Table,
  Tag,
  Typography,
} from "antd";

interface JadwalItem {
  id: string;
  paket_id: string;
  paket_nama?: string;
  paket_kode?: string;
  kelas_nama?: string;
  tanggal: string;
  jam_mulai: string;
  jam_selesai: string;
  status: string;
}

interface KoreksiItem {
  jawaban_id: string;
  attempt_id: string;
  soal_id: string;
  jawaban: unknown;
  ragu: boolean;
  poin: number | null;
  dikoreksi: boolean;
  updated_at: string;
  siswa_id: string;
  attempt_status: string;
  nama_siswa: string;
  nisn: string;
  tipe: string;
  pertanyaan: string;
  bobot: number;
}

interface KoreksiData {
  jadwal_id: string;
  items: KoreksiItem[];
}

interface SiswaItem {
  id: string;
  user_id: string | null;
  nisn: string;
  nama: string;
}

// Catatan envelope: endpoint list master memakai writeList backend —
// apiClient.get mengembalikan json.data yang untuk endpoint ini adalah
// ARRAY siswa langsung (meta ikut terbuang apiClient, seperti di
// SiswaPage). Karena itu loop halaman berhenti saat halaman < 100 item.

interface KoreksiHasil {
  jawaban_id: string;
  attempt_id: string;
  poin: number;
  menunggu_koreksi: boolean;
  nilai_akhir_baru?: number;
  final?: boolean;
}

function labelJadwal(j: JadwalItem): string {
  const paket = j.paket_nama ?? j.paket_kode ?? "Paket";
  const kelas = j.kelas_nama ? ` · ${j.kelas_nama}` : "";
  return `${paket}${kelas} · ${formatTanggal(j.tanggal)} · ${formatJam(j.jam_mulai)}–${formatJam(j.jam_selesai)} WIB`;
}

/** Jawaban siswa disimpan sebagai JSON apa pun: string / array / objek. */
function teksJawaban(jawaban: unknown): string {
  if (jawaban === null || jawaban === undefined) return "—";
  if (typeof jawaban === "string") return jawaban === "" ? "—" : jawaban;
  if (typeof jawaban === "number" || typeof jawaban === "boolean") return String(jawaban);
  if (Array.isArray(jawaban)) {
    if (jawaban.length === 0) return "—";
    return jawaban
      .map((it) => {
        if (typeof it === "string") return it;
        if (it && typeof it === "object") {
          const o = it as Record<string, unknown>;
          const teks = o.teks ?? o.jawaban ?? o.label ?? o.nama_file ?? o.file_url;
          return typeof teks === "string" ? teks : JSON.stringify(it);
        }
        return String(it);
      })
      .join("; ");
  }
  if (typeof jawaban === "object") {
    const o = jawaban as Record<string, unknown>;
    const teks = o.teks ?? o.jawaban ?? o.nama_file ?? o.file_url;
    if (typeof teks === "string" && teks !== "") return teks;
    return JSON.stringify(jawaban);
  }
  return String(jawaban);
}

function labelTipe(tipe: string): string {
  return TIPE_SOAL[tipe]?.label ?? tipe;
}

export default function KoreksiPage() {
  const queryClient = useQueryClient();
  const { hasPermission, user, loading } = useAuthStore();
  // Izin halaman: koreksi.update (admin/guru). Guard-nya di bawah,
  // SETELAH semua hooks (Rules of Hooks: tidak boleh return sebelum hooks).

  const [mapelId, setMapelId] = useState("");
  const [jadwalInput, setJadwalInput] = useState("");
  const [jadwalId, setJadwalId] = useState("");
  const [tipeFilter, setTipeFilter] = useState("");
  const [cari, setCari] = useState("");
  const [pesan, setPesan] = useState("");
  const [itemAktif, setItemAktif] = useState<KoreksiItem | null>(null);
  const [skor, setSkor] = useState<number | null>(null);

  const jadwalQuery = useQuery({
    queryKey: ["jadwal"],
    queryFn: () => apiClient.get<{ items: JadwalItem[] }>("/api/v1/jadwal"),
  });
  const paketQuery = useQuery({
    queryKey: ["paket"],
    queryFn: () => apiClient.get<{ items: Paket[] }>("/api/v1/paket"),
  });
  const jadwalList = jadwalQuery.data?.items ?? [];
  const paketList = paketQuery.data?.items ?? [];

  // Mapel (bank soal) unik dari paket-paket yang dipakai jadwal.
  const mapelOptions = useMemo(() => {
    const seen = new Map<string, string>();
    for (const p of paketList) {
      if (p.mapel_id && !seen.has(p.mapel_id)) {
        seen.set(p.mapel_id, p.mapel_nama ?? p.mapel_id);
      }
    }
    return [...seen.entries()]
      .map(([value, label]) => ({ value, label }))
      .sort((a, b) => a.label.localeCompare(b.label, "id"));
  }, [paketList]);

  const paketById = useMemo(() => new Map(paketList.map((p) => [p.id, p])), [paketList]);

  // Jadwal yang tampil dibatasi bank soal (mapel) yang dipilih.
  const jadwalTersaring = useMemo(() => {
    if (mapelId === "") return jadwalList;
    return jadwalList.filter((j) => paketById.get(j.paket_id)?.mapel_id === mapelId);
  }, [jadwalList, mapelId, paketById]);

  const koreksiQuery = useQuery({
    queryKey: ["koreksi", jadwalId],
    queryFn: () => apiClient.get<KoreksiData>(`/api/v1/koreksi/${encodeURIComponent(jadwalId)}`),
    enabled: jadwalId !== "",
  });

  // WORKAROUND (frontend saja, tanpa ubah backend): listKoreksi backend
  // meng-join `siswa s ON s.id = a.siswa_id`, padahal attempt.siswa_id
  // menyimpan users.id (lihat middleware auth + monitor yang memakai
  // `s.user_id = a.siswa_id`). Akibatnya nama_siswa/nisn dari endpoint
  // koreksi selalu kosong pada data nyata (tes E2E lama tidak menangkap
  // ini karena memakai UUID yang sama untuk user & siswa). Sampai backend
  // memperbaiki join-nya, halaman ini memetakan identitas lewat daftar
  // siswa (user_id -> nama/nisn). Bila backend sudah diperbaiki, nilai
  // dari endpoint dipakai duluan (lihat namaTampil/nisnTampil di bawah).
  // Daftar siswa dipaginasi maks 100/halaman — ambil semua halaman.
  const siswaQuery = useQuery({
    queryKey: ["siswa", "koreksi-nama"],
    queryFn: async () => {
      const semua: SiswaItem[] = [];
      let page = 1;
      for (;;) {
        const items = await apiClient.get<SiswaItem[]>(`/api/v1/master/siswa?per_page=100&page=${page}`);
        semua.push(...items);
        if (items.length < 100) break;
        page += 1;
      }
      return semua;
    },
  });
  const siswaByUserId = useMemo(
    () => new Map((siswaQuery.data ?? []).filter((s) => s.user_id).map((s) => [s.user_id as string, s])),
    [siswaQuery.data],
  );
  // Identitas tampil: utamakan dari endpoint koreksi, fallback ke peta siswa.
  const namaTampil = (it: KoreksiItem): string =>
    it.nama_siswa !== "" ? it.nama_siswa : (siswaByUserId.get(it.siswa_id)?.nama ?? "—");
  const nisnTampil = (it: KoreksiItem): string =>
    it.nisn !== "" ? it.nisn : (siswaByUserId.get(it.siswa_id)?.nisn ?? "—");

  const koreksiMutation = useMutation({
    mutationFn: ({ jawabanId, skorNilai }: { jawabanId: string; skorNilai: number }) =>
      apiClient.post<KoreksiHasil>(`/api/v1/koreksi/${encodeURIComponent(jawabanId)}`, { skor: skorNilai }),
    onSuccess: (data) => {
      const nilaiBaru =
        data.nilai_akhir_baru !== undefined ? ` Nilai akhir pengerjaan: ${data.nilai_akhir_baru.toFixed(2)}.` : "";
      const statusFinal = data.final === true ? " Semua soal manual sudah terkoreksi (final)." : "";
      setPesan(`Koreksi tersimpan (poin ${data.poin.toFixed(2)}).${nilaiBaru}${statusFinal}`);
      setItemAktif(null);
      setSkor(null);
      void queryClient.invalidateQueries({ queryKey: ["koreksi", jadwalId] });
    },
    onError: (err: Error) => setPesan(`Gagal menyimpan koreksi: ${err.message}`),
  });

  useEffect(() => {
    setPesan("");
  }, [jadwalId]);

  const handlePilihMapel = (v: string) => {
    setMapelId(v);
    setJadwalInput("");
    setJadwalId("");
    setTipeFilter("");
  };

  const handleTampilkan = () => {
    if (jadwalInput.trim() === "") return;
    setJadwalId(jadwalInput.trim());
  };

  const items = koreksiQuery.data?.items ?? [];
  const tipeOptions = useMemo(() => {
    const seen = new Set(items.map((it) => it.tipe));
    return [...seen].map((t) => ({ value: t, label: labelTipe(t) })).sort((a, b) => a.label.localeCompare(b.label, "id"));
  }, [items]);

  const rows = useMemo(() => {
    const kata = cari.trim().toLowerCase();
    return items.filter((it) => {
      if (tipeFilter !== "" && it.tipe !== tipeFilter) return false;
      if (kata === "") return true;
      return (
        namaTampil(it).toLowerCase().includes(kata) ||
        nisnTampil(it).toLowerCase().includes(kata) ||
        it.pertanyaan.toLowerCase().includes(kata)
      );
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [items, tipeFilter, cari, siswaByUserId]);

  const poinPreview =
    itemAktif !== null && skor !== null ? (skor / 100) * itemAktif.bobot : null;

  // Guard halaman pada koreksi.update: proktor (dan role lain tanpa izin
  // ini) dialihkan ke /403. Menu di AdminLayout memakai gate yang sama.
  // Semua yang lolos guard di sini boleh melihat daftar + mengoreksi.
  if (!loading && user && !hasPermission("koreksi.update")) {
    return <Navigate to="/403" replace />;
  }

  return (
    <div className="space-y-4">
      <Breadcrumb items={[{ title: <Link to="/adm">Dashboard</Link> }, { title: "Koreksi Esai" }]} />
      <div className="border-b border-border pb-4">
        <Typography.Title level={3} style={{ margin: 0 }}>
          Koreksi Esai
        </Typography.Title>
        <Typography.Text type="secondary">
          Daftar jawaban esai yang menunggu dikoreksi per jadwal. Skor 0–100 per jawaban; poin = skor/100 × bobot.
          Nilai akhir dihitung ulang otomatis bila pengerjaan sudah selesai.
        </Typography.Text>
      </div>

      {pesan !== "" && <Alert type="info" showIcon message={pesan} />}

      <Form layout="vertical" onFinish={handleTampilkan} className="flex flex-col gap-2 sm:flex-row sm:items-end">
        <Form.Item label="Mata Pelajaran (Bank Soal)" className="mb-0 w-full sm:max-w-xs">
          <Select
            showSearch
            options={[{ value: "", label: "Semua mata pelajaran" }, ...mapelOptions]}
            value={mapelId}
            onChange={handlePilihMapel}
            placeholder="Semua mata pelajaran"
            filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())}
            optionFilterProp="label"
            notFoundContent="Tidak ada data"
            style={{ width: "100%" }}
            size="large"
          />
        </Form.Item>
        <Form.Item label="Jadwal" className="mb-0 w-full sm:max-w-sm">
          <Select
            showSearch
            options={[
              { value: "", label: "— Pilih jadwal —" },
              ...jadwalTersaring.map((j) => ({ value: j.id, label: labelJadwal(j) })),
            ]}
            value={jadwalInput}
            onChange={(v) => setJadwalInput(v)}
            placeholder="— Pilih jadwal —"
            filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())}
            optionFilterProp="label"
            notFoundContent="Tidak ada data"
            style={{ width: "100%" }}
            size="large"
          />
        </Form.Item>
        <Button htmlType="submit" type="primary">
          Tampilkan
        </Button>
      </Form>
      {(jadwalQuery.isError || paketQuery.isError) && (
        <Alert type="error" showIcon message="Gagal memuat daftar jadwal / paket." />
      )}

      {jadwalId !== "" && (
        <div className="flex flex-col gap-2 sm:flex-row sm:items-end">
          <Form.Item label="Tipe Soal" className="mb-0 w-full sm:max-w-xs">
            <Select
              options={[{ value: "", label: "Semua tipe" }, ...tipeOptions]}
              value={tipeFilter}
              onChange={(v) => setTipeFilter(v)}
              placeholder="Semua tipe"
              style={{ width: "100%" }}
              size="large"
            />
          </Form.Item>
          <Form.Item label="Cari (nama / NISN / pertanyaan)" className="mb-0 w-full sm:max-w-sm">
            <Input
              value={cari}
              onChange={(e) => setCari(e.target.value)}
              placeholder="Cari nama, NISN, atau pertanyaan…"
              allowClear
              size="large"
            />
          </Form.Item>
          <Button onClick={() => void koreksiQuery.refetch()}>Muat Ulang</Button>
        </div>
      )}

      {jadwalId !== "" && koreksiQuery.isSuccess && (
        <Card>
          <Statistic title="Menunggu koreksi" value={items.length} suffix="jawaban" />
        </Card>
      )}

      {jadwalId !== "" && koreksiQuery.isPending && <p className="text-sm text-muted">Memuat daftar koreksi...</p>}
      {jadwalId !== "" && koreksiQuery.isError && (
        <Alert type="error" showIcon message={`Gagal memuat daftar koreksi: ${(koreksiQuery.error as Error).message}`} />
      )}

      {jadwalId !== "" && koreksiQuery.isSuccess && (
        <Table
          columns={[
            {
              key: "nisn",
              title: "NISN",
              sorter: (a, b) => nisnTampil(a).localeCompare(nisnTampil(b), "id"),
              render: (_: unknown, r: KoreksiItem) => <>{nisnTampil(r)}</>,
            },
            {
              key: "nama",
              title: "Nama Siswa",
              sorter: (a, b) => namaTampil(a).localeCompare(namaTampil(b), "id"),
              render: (_: unknown, r: KoreksiItem) => <>{namaTampil(r)}</>,
            },
            {
              key: "tipe",
              title: "Tipe",
              dataIndex: "tipe",
              render: (_: unknown, r: KoreksiItem) => <Tag>{labelTipe(r.tipe)}</Tag>,
            },
            {
              key: "pertanyaan",
              title: "Pertanyaan",
              dataIndex: "pertanyaan",
              render: (_: unknown, r: KoreksiItem) => (
                <span style={{ display: "block", maxWidth: 320, whiteSpace: "pre-wrap" }}>{r.pertanyaan}</span>
              ),
            },
            {
              key: "bobot",
              title: "Bobot",
              dataIndex: "bobot",
              sorter: (a, b) => a.bobot - b.bobot,
              render: (_: unknown, r: KoreksiItem) => <>{r.bobot}</>,
            },
            {
              key: "jawaban",
              title: "Jawaban Siswa",
              render: (_: unknown, r: KoreksiItem) => (
                <span style={{ display: "block", maxWidth: 360, whiteSpace: "pre-wrap" }}>{teksJawaban(r.jawaban)}</span>
              ),
            },
            {
              key: "ragu",
              title: "Ragu",
              render: (_: unknown, r: KoreksiItem) =>
                r.ragu ? <Tag color="gold">Ragu</Tag> : <Tag>—</Tag>,
            },
            {
              key: "status",
              title: "Status Pengerjaan",
              dataIndex: "attempt_status",
              render: (_: unknown, r: KoreksiItem) => <Tag color={r.attempt_status === "selesai" ? "green" : "blue"}>{r.attempt_status}</Tag>,
            },
            // Kolom Aksi selalu tampil di sini: yang tanpa
            // koreksi.update sudah dialihkan ke /403 oleh guard di atas.
            {
              key: "aksi",
              title: "Aksi",
              render: (_: unknown, r: KoreksiItem) => (
                <Button
                  type="primary"
                  onClick={() => {
                    setItemAktif(r);
                    setSkor(null);
                  }}
                >
                  Koreksi
                </Button>
              ),
            },
          ]}
          dataSource={rows}
          rowKey="jawaban_id"
          pagination={false}
          locale={{ emptyText: "Tidak ada jawaban esai yang menunggu koreksi untuk jadwal ini." }}
          scroll={{ x: "max-content" }}
        />
      )}

      <Modal
        title="Koreksi Jawaban Esai"
        open={itemAktif !== null}
        onCancel={() => {
          setItemAktif(null);
          setSkor(null);
        }}
        onOk={() => {
          if (itemAktif === null || skor === null) return;
          koreksiMutation.mutate({ jawabanId: itemAktif.jawaban_id, skorNilai: skor });
        }}
        okText="Simpan Koreksi"
        cancelText="Batal"
        confirmLoading={koreksiMutation.isPending}
        okButtonProps={{ disabled: skor === null }}
        width={720}
        destroyOnClose
      >
        {itemAktif !== null && (
          <div className="space-y-3">
            <p>
              <strong>{namaTampil(itemAktif)}</strong> ({nisnTampil(itemAktif)}) — {labelTipe(itemAktif.tipe)} · Bobot{" "}
              {itemAktif.bobot}
            </p>
            <p style={{ whiteSpace: "pre-wrap" }}>
              <strong>Pertanyaan:</strong> {itemAktif.pertanyaan}
            </p>
            <p style={{ whiteSpace: "pre-wrap" }}>
              <strong>Jawaban siswa:</strong> {teksJawaban(itemAktif.jawaban)}
            </p>
            <Form layout="vertical">
              <Form.Item
                label="Skor (0–100)"
                required
                help={
                  poinPreview !== null
                    ? `Poin yang akan disimpan: ${poinPreview.toFixed(2)} (skor ${skor} / 100 × bobot ${itemAktif.bobot}).`
                    : "Isi skor 0–100. Poin = skor/100 × bobot."
                }
              >
                <InputNumber
                  min={0}
                  max={100}
                  step={0.01}
                  value={skor}
                  onChange={(v) => setSkor(typeof v === "number" ? v : null)}
                  placeholder="0–100"
                  style={{ width: "100%" }}
                  size="large"
                  autoFocus
                />
              </Form.Item>
            </Form>
            <Alert
              type="warning"
              showIcon
              message="Koreksi tidak dapat dibatalkan dari halaman ini. Periksa skor sebelum menyimpan."
            />
          </div>
        )}
      </Modal>
    </div>
  );
}
