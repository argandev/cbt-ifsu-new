import { useEffect, useState } from "react";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ArrowLeft, Save } from "lucide-react";
import apiClient from "../../../lib/api";
import type { MasterOption, Ukk, UkkNilai } from "./types";
import { fmtNilai } from "./types";

import { Alert, Button, Card, Input, InputNumber, Select, Table, Tag, Breadcrumb, Space, Typography, Form } from "antd";
interface NilaiFormState {
  nilai_teori: string;
  nilai_proses: string;
  nilai_produk: string;
  nilai_sikap: string;
  catatan: string;
}

const EMPTY_FORM: NilaiFormState = {
  nilai_teori: "",
  nilai_proses: "",
  nilai_produk: "",
  nilai_sikap: "",
  catatan: "",
};

function toNum(s: string): number | null {
  if (s.trim() === "") return null;
  const v = Number(s);
  return Number.isFinite(v) ? v : null;
}

/** Pratinjau hitung di form (tampilan saja). Nilai RESMI selalu dihitung
 *  ulang backend saat simpan dan yang dipakai adalah hasil dari server. */
function pratinjau(form: NilaiFormState) {
  const teori = toNum(form.nilai_teori);
  const proses = toNum(form.nilai_proses);
  const produk = toNum(form.nilai_produk);
  const sikap = toNum(form.nilai_sikap);
  const praktik =
    proses !== null && produk !== null && sikap !== null
      ? Math.round((proses * 0.3 + produk * 0.5 + sikap * 0.2) * 100) / 100
      : null;
  const akhir =
    teori !== null && praktik !== null
      ? Math.round((teori * 0.4 + praktik * 0.6) * 100) / 100
      : null;
  const lulus = akhir !== null ? akhir >= 70 : null;
  const predikat =
    akhir === null ? null : akhir < 70 ? "Belum Tuntas" : akhir < 80 ? "Cukup" : akhir <= 90 ? "Baik" : "Sangat Baik";
  return { praktik, akhir, lulus, predikat };
}

export default function UkkNilaiPage() {
  const { id = "" } = useParams();
  const [searchParams] = useSearchParams();
  const queryClient = useQueryClient();
  const [pesan, setPesan] = useState("");
  const [siswaId, setSiswaId] = useState(searchParams.get("siswa") ?? "");
  const [form, setForm] = useState<NilaiFormState>(EMPTY_FORM);
  const [editingId, setEditingId] = useState<string | null>(null);

  const ukkQuery = useQuery({
    queryKey: ["ukk", id],
    queryFn: () => apiClient.get<Ukk>(`/ukk/${id}`),
    enabled: id !== "",
  });
  const nilaiQuery = useQuery({
    queryKey: ["ukk-nilai", id],
    queryFn: () => apiClient.get<UkkNilai[]>(`/ukk/${id}/nilai`),
    enabled: id !== "",
  });
  // D5: daftar siswa untuk dropdown dari endpoint khusus UKK
  // GET /ukk/:id/peserta (izin ukk.read) — bukan /master/siswa (izin
  // master.siswa.read) — agar penguji_ukk bisa tambah nilai siswa baru
  // tanpa 403, tanpa izin master tambahan.
  const siswaQuery = useQuery({
    queryKey: ["ukk-peserta", id],
    queryFn: () => apiClient.get<MasterOption[]>(`/ukk/${id}/peserta`),
    enabled: id !== "",
  });

  const nilaiRows = nilaiQuery.data ?? [];
  // Saat ganti siswa, isi form dari nilai yang sudah ada (bila ada).
  useEffect(() => {
    if (!siswaId) {
      setForm(EMPTY_FORM);
      setEditingId(null);
      return;
    }
    const found = nilaiRows.find((n) => n.siswa_id === siswaId);
    if (found) {
      setEditingId(found.id);
      setForm({
        nilai_teori: found.nilai_teori !== null && found.nilai_teori !== undefined ? String(found.nilai_teori) : "",
        nilai_proses: found.nilai_proses !== null && found.nilai_proses !== undefined ? String(found.nilai_proses) : "",
        nilai_produk: found.nilai_produk !== null && found.nilai_produk !== undefined ? String(found.nilai_produk) : "",
        nilai_sikap: found.nilai_sikap !== null && found.nilai_sikap !== undefined ? String(found.nilai_sikap) : "",
        catatan: found.catatan ?? "",
      });
    } else {
      setEditingId(null);
      setForm(EMPTY_FORM);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [siswaId]);

  const saveMutation = useMutation({
    mutationFn: (payload: { siswa_id: string; nilaiId: string | null; body: Record<string, unknown> }) =>
      payload.nilaiId
        ? apiClient.put<UkkNilai>(`/ukk/${id}/nilai/${payload.nilaiId}`, payload.body)
        : apiClient.post<UkkNilai>(`/ukk/${id}/nilai`, { siswa_id: payload.siswa_id, ...payload.body }),
    onSuccess: (saved) => {
      setPesan(
        `Nilai tersimpan (dihitung server): praktik ${fmtNilai(saved.nilai_praktik)}, ` +
          `akhir ${fmtNilai(saved.nilai_akhir)}, ${saved.predikat ?? "-"}${saved.lulus ? ", Lulus" : ""}.`,
      );
      void queryClient.invalidateQueries({ queryKey: ["ukk-nilai", id] });
    },
    onError: (err: Error) => setPesan(err.message),
  });

  const preview = pratinjau(form);
  const ukk = ukkQuery.data;

  const setField = (key: keyof NilaiFormState, value: string) =>
    setForm((prev) => ({ ...prev, [key]: value }));

  const handleSubmit = () => {
    setPesan("");
    if (!siswaId) {
      setPesan("Pilih siswa terlebih dahulu.");
      return;
    }
    const teori = toNum(form.nilai_teori);
    const proses = toNum(form.nilai_proses);
    const produk = toNum(form.nilai_produk);
    const sikap = toNum(form.nilai_sikap);
    for (const [label, v] of [
      ["Teori", teori],
      ["Proses", proses],
      ["Produk", produk],
      ["Sikap", sikap],
    ] as Array<[string, number | null]>) {
      if (v !== null && (v < 0 || v > 100)) {
        setPesan(`Nilai ${label} harus 0-100.`);
        return;
      }
    }
    saveMutation.mutate({
      siswa_id: siswaId,
      nilaiId: editingId,
      body: {
        nilai_teori: teori,
        nilai_proses: proses,
        nilai_produk: produk,
        nilai_sikap: sikap,
        catatan: form.catatan.trim() || null,
      },
    });
  };

  if (ukkQuery.isLoading) return <p className="p-1 text-muted">Memuat...</p>;
  if (ukkQuery.isError || !ukk) {
    return <Alert type="error" showIcon message="UKK tidak ditemukan." />;
  }

  return (
    <div className="flex flex-col gap-4">
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: (<Link to="/adm/ukk">UKK</Link>) }, { title: "Input Nilai" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>Input Nilai UKK</Typography.Title>
          <Typography.Text type="secondary">{`${ukk.nama} — Teori 40% + Praktik 60%. Praktik = Proses 30% + Produk 50% + Sikap 20%. Lulus ≥ 70. Nilai akhir, lulus, dan predikat dihitung server saat simpan.`}</Typography.Text>
        </div>
        <Space>
          <Link to={`/adm/ukk/${id}`}>
            <Button icon={<ArrowLeft size={16} />}>
              Kembali ke Detail UKK
            </Button>
          </Link>
        </Space>
      </div>
      {ukk.status === "selesai" ? (
        <Alert type="warning" showIcon message={<>
          UKK sudah selesai: nilai hanya dapat diubah oleh admin (dicatat di audit log).
        </>} />
      ) : null}

      {pesan ? <Alert type="info" showIcon message={pesan} /> : null}

      <Card title="Form Nilai per Siswa">
        
        <div>
          <Form layout="vertical" onFinish={handleSubmit}>
            <Form.Item label="Siswa"><Select showSearch allowClear options={(siswaQuery.data ?? []).map((s) => ({
                value: s.id,
                label: s.nisn ? `${s.nama} (${s.nisn})` : s.nama,
              }))} value={siswaId} onChange={(v) => setSiswaId((v ?? ""))} placeholder="Pilih siswa" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" /></Form.Item>
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
              <Form.Item label="Nilai Teori (0-100)"><InputNumber min={0} max={100} step={0.01} value={form.nilai_teori === "" ? null : Number(form.nilai_teori)} onChange={(v) => setField("nilai_teori", v === null ? "" : String(v))} placeholder="mis. 85" size="large" style={{ width: "100%" }} /></Form.Item>
              <Form.Item label="Rubrik Proses (0-100)"><InputNumber min={0} max={100} step={0.01} value={form.nilai_proses === "" ? null : Number(form.nilai_proses)} onChange={(v) => setField("nilai_proses", v === null ? "" : String(v))} placeholder="mis. 80" size="large" style={{ width: "100%" }} /></Form.Item>
              <Form.Item label="Rubrik Produk (0-100)"><InputNumber min={0} max={100} step={0.01} value={form.nilai_produk === "" ? null : Number(form.nilai_produk)} onChange={(v) => setField("nilai_produk", v === null ? "" : String(v))} placeholder="mis. 90" size="large" style={{ width: "100%" }} /></Form.Item>
              <Form.Item label="Rubrik Sikap (0-100)"><InputNumber min={0} max={100} step={0.01} value={form.nilai_sikap === "" ? null : Number(form.nilai_sikap)} onChange={(v) => setField("nilai_sikap", v === null ? "" : String(v))} placeholder="mis. 88" size="large" style={{ width: "100%" }} /></Form.Item>
            </div>
            <Form.Item label="Catatan penguji"><Input value={form.catatan} onChange={(e) => setField("catatan", e.target.value)} placeholder="Catatan hasil praktik (opsional)" size="large" /></Form.Item>

            <Card title="Pratinjau (dihitung ulang server saat simpan)">
              <div className="mt-2 grid gap-2 text-sm text-ink sm:grid-cols-4">
                <span>
                  Praktik: <strong>{preview.praktik === null ? "-" : preview.praktik.toFixed(2)}</strong>
                </span>
                <span>
                  Akhir: <strong>{preview.akhir === null ? "-" : preview.akhir.toFixed(2)}</strong>
                </span>
                <span>
                  Lulus:{" "}
                  <strong>
                    {preview.lulus === null ? "-" : preview.lulus ? "Ya" : "Tidak"}
                  </strong>
                </span>
                <span>
                  Predikat: <strong>{preview.predikat ?? "-"}</strong>
                </span>
              </div>
              {preview.akhir === null ? (
                <p className="mt-2 text-xs text-muted">
                  Nilai akhir muncul setelah teori + ketiga rubrik terisi.
                </p>
              ) : null}
            </Card>

            <Button htmlType="submit" loading={saveMutation.isPending} type="primary">
              <Save size={16} /> {saveMutation.isPending ? "Menyimpan..." : "Simpan Nilai"}
            </Button>
          </Form>
        </div>
      </Card>

      <Card title="Nilai Tersimpan">
        
        <div>
            <Table columns={[{ key: "col0", title: "Siswa", dataIndex: "siswa_nama", sorter: (a: any, b: any) => { const _va = a.siswa_nama, _vb = b.siswa_nama; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, n: any) => (<>{n.siswa_nama ?? "-"}</>) }, { key: "col1", title: "Teori", render: (_: unknown, n: any) => (<>{fmtNilai(n.nilai_teori)}</>) }, { key: "col2", title: "Proses", render: (_: unknown, n: any) => (<>{fmtNilai(n.nilai_proses)}</>) }, { key: "col3", title: "Produk", render: (_: unknown, n: any) => (<>{fmtNilai(n.nilai_produk)}</>) }, { key: "col4", title: "Sikap", render: (_: unknown, n: any) => (<>{fmtNilai(n.nilai_sikap)}</>) }, { key: "col5", title: "Praktik", render: (_: unknown, n: any) => (<>{fmtNilai(n.nilai_praktik)}</>) }, { key: "col6", title: "Akhir", render: (_: unknown, n: any) => (<>{fmtNilai(n.nilai_akhir)}</>) }, { key: "col7", title: "Lulus", render: (_: unknown, n: any) => (<>
                      {n.lulus === null || n.lulus === undefined ? (
                        "-"
                      ) : n.lulus ? (
                        <Tag color="green">Lulus</Tag>
                      ) : (
                        <Tag color="red">Tidak Lulus</Tag>
                      )}
                    </>) }, { key: "col8", title: "Predikat", dataIndex: "predikat", sorter: (a: any, b: any) => { const _va = a.predikat, _vb = b.predikat; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, n: any) => (<>{n.predikat ?? "-"}</>) }, { key: "col9", title: "Aksi", render: (_: unknown, n: any) => (<>
                      <Button onClick={() => setSiswaId(n.siswa_id)} type="default">
                        Edit
                      </Button>
                    </>) }]} dataSource={nilaiRows} rowKey="id" pagination={false} locale={{ emptyText: "Belum ada nilai." }} scroll={{ x: "max-content" }} />
        </div>
      </Card>
    </div>
  );
}
