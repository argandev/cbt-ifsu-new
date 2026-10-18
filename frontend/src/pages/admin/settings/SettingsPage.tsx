import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Save } from "lucide-react";
import apiClient from "../../../lib/api";
import { useAuthStore } from "../../../store/auth";
import ImageUpload from "../../../components/ui/ImageUpload";

import { Alert, Button, Checkbox, Input, InputNumber, Breadcrumb, Typography, Form, Card } from "antd";
interface SettingsData {
  tampil_hasil_ke_siswa: boolean;
  token_expiry_detik: number;
  toleransi_telat_menit: number;
  max_pelanggaran: number;
  durasi_default: number;
  nama_sekolah: string;
  logo_url: string;
  kktp_per_mapel: Record<string, unknown>;
  remedial_mode: string;
}

const NUMBER_KEYS: Array<keyof SettingsData> = [
  "token_expiry_detik",
  "toleransi_telat_menit",
  "max_pelanggaran",
  "durasi_default",
];

export default function SettingsPage() {
  const { hasPermission } = useAuthStore();
  const queryClient = useQueryClient();
  const [pesan, setPesan] = useState("");
  const [form, setForm] = useState<Partial<SettingsData>>({});
  const [kktpText, setKktpText] = useState("{}");
  const [kktpError, setKktpError] = useState("");

  const query = useQuery({
    queryKey: ["settings"],
    queryFn: () => apiClient.get<SettingsData>("/api/v1/settings"),
  });

  useEffect(() => {
    if (query.data) {
      setForm(query.data);
      setKktpText(JSON.stringify(query.data.kktp_per_mapel ?? {}, null, 2));
    }
  }, [query.data]);

  const mutation = useMutation({
    mutationFn: (payload: Record<string, unknown>) =>
      apiClient.put<{ updated: Record<string, unknown> }>("/api/v1/settings", { settings: payload }),
    onSuccess: () => {
      setPesan("Settings berhasil disimpan.");
      void queryClient.invalidateQueries({ queryKey: ["settings"] });
    },
    onError: (err: Error) => setPesan(err.message),
  });

  const setField = <K extends keyof SettingsData>(key: K, value: SettingsData[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }));
  };

  const handleSubmit = () => {
    setPesan("");
    setKktpError("");
    let kktp: Record<string, unknown>;
    try {
      const parsed: unknown = JSON.parse(kktpText);
      if (parsed === null || typeof parsed !== "object" || Array.isArray(parsed)) {
        throw new Error("KKTP harus berupa objek JSON.");
      }
      kktp = parsed as Record<string, unknown>;
    } catch (err) {
      setKktpError(err instanceof Error ? err.message : "JSON KKTP tidak valid.");
      return;
    }
    mutation.mutate({
      tampil_hasil_ke_siswa: form.tampil_hasil_ke_siswa ?? false,
      token_expiry_detik: Number(form.token_expiry_detik ?? 900),
      toleransi_telat_menit: Number(form.toleransi_telat_menit ?? 15),
      max_pelanggaran: Number(form.max_pelanggaran ?? 2),
      durasi_default: Number(form.durasi_default ?? 90),
      nama_sekolah: form.nama_sekolah ?? "",
      logo_url: form.logo_url ?? "",
      kktp_per_mapel: kktp,
      remedial_mode: "tertinggi",
    });
  };

  if (query.isPending) return <p className="text-sm text-muted">Memuat settings...</p>;
  if (query.isError)
    return (
      <Alert type="error" showIcon message={`Gagal memuat settings: ${(query.error as Error).message}`} />
    );


  return (
    <div className="space-y-6">
      {/* Breadcrumb: tampilkanBeranda=false agar link Beranda tidak mengarah
          ke "/" (halaman siswa) — Dashboard admin ada di "/adm". */}
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: "Settings" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>Settings</Typography.Title>
          <Typography.Text type="secondary">Pengaturan sekolah untuk CBT. Perubahan tercatat di audit log.</Typography.Text>
        </div>
      </div>

      {pesan !== "" && (
        <Alert type="info" showIcon message={<>
          {pesan}
        </>} />
      )}

      <Form layout="vertical" onFinish={handleSubmit} className="max-w-2xl">
        <Card>
        <div className="flex items-center justify-between gap-4">
          <div>
            <Typography.Text strong>Tampilkan hasil ke siswa</Typography.Text>
            <p className="text-xs text-muted">
              Jika aktif, siswa dapat melihat nilainya sendiri (hanya nilai final).
            </p>
          </div>
          <Checkbox id="tampil-hasil" checked={form.tampil_hasil_ke_siswa ?? false} onChange={(e) => setField("tampil_hasil_ke_siswa", e.target.checked)} />
        </div>
        </Card>

        <Card>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Form.Item label="Token expiry (detik)">
            <InputNumber id="token-expiry" min={60} max={86400} value={form.token_expiry_detik ?? 900} onChange={(v) => setField("token_expiry_detik", typeof v === "number" ? v : 0)} size="large" style={{ width: "100%" }} />
          </Form.Item>
          <Form.Item label="Toleransi telat (menit)">
            <InputNumber id="toleransi" min={0} max={120} value={form.toleransi_telat_menit ?? 15} onChange={(v) => setField("toleransi_telat_menit", typeof v === "number" ? v : 0)} size="large" style={{ width: "100%" }} />
          </Form.Item>
          <Form.Item label="Max pelanggaran">
            <InputNumber id="max-pelanggaran" min={1} max={10} value={form.max_pelanggaran ?? 2} onChange={(v) => setField("max_pelanggaran", typeof v === "number" ? v : 0)} size="large" style={{ width: "100%" }} />
            <p className="mt-1 text-xs text-muted">Pelanggaran ke-1 = peringatan, ke-2 = auto-block (per ujian).</p>
          </Form.Item>
          <Form.Item label="Durasi default (menit)">
            <InputNumber id="durasi" min={5} max={600} value={form.durasi_default ?? 90} onChange={(v) => setField("durasi_default", typeof v === "number" ? v : 0)} size="large" style={{ width: "100%" }} />
          </Form.Item>
        </div>
        </Card>

        <Card>
          <Form.Item label="Nama sekolah">
          <Input id="nama-sekolah" type="text" value={form.nama_sekolah ?? ""} onChange={(e) => setField("nama_sekolah", e.target.value)} size="large" />
          </Form.Item>
        </Card>

        <Card>
          {/* Logo diunggah sebagai file (bukan URL ketik) — tersimpan di
              server sebagai /api/v1/upload/image/<uuid>.<ext>. Logo lama
              yang masih berupa URL eksternal tetap ditampilkan sampai
              diganti/dihapus. */}
          <ImageUpload
            label="Logo sekolah"
            value={form.logo_url ?? ""}
            onChange={(url) => setField("logo_url", url)}
            disabled={mutation.isPending}
          />
        </Card>

        <Card>
          <Form.Item label="KKTP per mapel (JSON)">
          <p className="mb-2 text-xs text-muted">
            KKTP di Kurikulum Merdeka adalah deskripsi / rubrik / interval per Tujuan Pembelajaran (TP) untuk
            tiap mapel — bukan satu angka KKM=75 untuk semua mapel. Contoh:{" "}
            <code className="rounded-lg bg-surface px-1">
              {`{"MAT": {"deskripsi": "Tuntas bila ...", "interval": "70-100"}}`}
            </code>
          </p>
          <Input.TextArea id="kktp" value={kktpText} onChange={(e) => setKktpText(e.target.value)} rows={8} spellCheck={false} size="large"></Input.TextArea>
          {kktpError !== "" && (
            <Alert type="error" showIcon message={kktpError} style={{ marginTop: 4 }} />
          )}
          </Form.Item>
        </Card>

        <Card>
          <Typography.Text strong>Mode remedial</Typography.Text>
          <p className="mt-1 text-sm text-muted">
            <span className="font-semibold">Tertinggi</span> (read-only) — nilai akhir remedial memakai nilai
            tertinggi antara nilai awal dan nilai remedial. Nilai awal tidak pernah ditimpa.
          </p>
          <input type="hidden" value={form.remedial_mode ?? "tertinggi"} readOnly />
        </Card>

        {hasPermission("settings.update") && (
          <Button htmlType="submit" loading={mutation.isPending} type="primary">
            <Save size={16} /> Simpan Settings
          </Button>
        )}
        <span className="hidden">{NUMBER_KEYS.length}</span>
      </Form>
    </div>
  );
}
