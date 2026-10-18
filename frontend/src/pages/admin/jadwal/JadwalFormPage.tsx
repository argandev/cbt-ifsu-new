// JadwalFormPage — halaman form buat/ubah jadwal (Tim B, penataan).
// Tampilan form SAJA: dipanggil dari JadwalPage saat mode "form", bukan lewat
// route baru (App.tsx bukan wewenang Tim B — route /adm/jadwal tetap daftar).
// Semua field, Dropdown, validasi required, dan payload TIDAK berubah —
// hanya dipindah dari kolom kiri grid lama ke halaman form sendiri.
import type { Paket } from "../../../lib/exam";
import { Link } from "react-router-dom";
import { useAuthStore } from "../../../store/auth";
import dayjs from "dayjs";
import customParseFormat from "dayjs/plugin/customParseFormat";

// WAJIB untuk TimePicker: dayjs inti MENGABAIKAN argumen format kedua —
// dayjs("10:15", "HH:mm") tanpa plugin ini = Invalid Date sehingga
// TimePicker selalu tampil kosong. Plugin membuat parse sesuai format
// ("HH:mm" maupun "HH:MM:SS" dari backend) + round-trip format benar.
// QA 18 Okt 2026: tanpa plugin, edit jadwal menampilkan jam kosong.
dayjs.extend(customParseFormat);

import { Alert, Button, DatePicker, Select, TimePicker, Breadcrumb, Typography, Form, Card } from "antd";
export interface JadwalFormValues {
  paket_id: string;
  kelas_id: string;
  tanggal: string;
  jam_mulai: string;
  jam_selesai: string;
  status: string;
}

interface Kelas {
  id: string;
  nama: string;
}

interface JadwalFormPageProps {
  mode: "buat" | "ubah";
  form: JadwalFormValues;
  setForm: (form: JadwalFormValues) => void;
  paketList: Paket[];
  kelasList: Kelas[];
  kelasError: boolean;
  error: string | null;
  onSubmit: () => void;
  onBatal: () => void;
}

export default function JadwalFormPage({
  mode,
  form,
  setForm,
  paketList,
  kelasList,
  kelasError,
  error,
  onSubmit,
  onBatal,
}: JadwalFormPageProps) {
  const { hasPermission } = useAuthStore();
  return (
    <div className="space-y-4">
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: "Jadwal" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>{mode === "ubah" ? "Ubah Jadwal" : "Buat Jadwal"}</Typography.Title>
          <Typography.Text type="secondary">Atur paket, kelas, tanggal, jam, dan status jadwal ujian.</Typography.Text>
        </div>
      </div>
      {error && <Alert type="error" showIcon message={<>{error}</>} />}
      <Card className="max-w-2xl">
        <Form layout="vertical" onFinish={onSubmit}>
            <Form.Item label="Paket"><Select showSearch options={[
                { value: "", label: "— Pilih paket —" },
                ...paketList.map((p) => ({ value: p.id, label: `${p.nama} (${p.jenis})` })),
              ]} value={form.paket_id} onChange={(v) => setForm({ ...form, paket_id: v })} placeholder="— Pilih paket —" id="jadwal-paket" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" /></Form.Item>
            <Form.Item label="Kelas"><Select showSearch options={[
                { value: "", label: "— Pilih kelas —" },
                ...kelasList.map((k) => ({ value: k.id, label: k.nama })),
              ]} value={form.kelas_id} onChange={(v) => setForm({ ...form, kelas_id: v })} placeholder="— Pilih kelas —" id="jadwal-kelas" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" /></Form.Item>
            {kelasError && (
              <Typography.Text type="danger" className="mt-1 block text-xs">Gagal memuat daftar kelas.</Typography.Text>
            )}
          <Form.Item label="Tanggal" rules={[{ required: true, message: "Tanggal wajib diisi." }]}><DatePicker format="YYYY-MM-DD" value={form.tanggal ? dayjs(form.tanggal, "YYYY-MM-DD") : null} onChange={(d) => setForm({ ...form, tanggal: d ? d.format("YYYY-MM-DD") : "" })} size="large" style={{ width: "100%" }} /></Form.Item>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <Form.Item label="Jam mulai" rules={[{ required: true, message: "Jam mulai wajib diisi." }]}><TimePicker format="HH:mm" value={form.jam_mulai ? dayjs(form.jam_mulai, "HH:mm") : null} onChange={(t) => setForm({ ...form, jam_mulai: t ? t.format("HH:mm") : "" })} size="large" style={{ width: "100%" }} /></Form.Item>
            <Form.Item label="Jam selesai" rules={[{ required: true, message: "Jam selesai wajib diisi." }]}><TimePicker format="HH:mm" value={form.jam_selesai ? dayjs(form.jam_selesai, "HH:mm") : null} onChange={(t) => setForm({ ...form, jam_selesai: t ? t.format("HH:mm") : "" })} size="large" style={{ width: "100%" }} /></Form.Item>
          </div>
            <Form.Item label="Status"><Select options={["draft", "aktif", "pause", "selesai"].map((s) => ({ value: s, label: s }))} value={form.status} onChange={(v) => setForm({ ...form, status: v })} id="jadwal-status" placeholder="Pilih…" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" /></Form.Item>
          <div className="flex flex-wrap gap-2 pt-1">
            {hasPermission(mode === "ubah" ? "jadwal.update" : "jadwal.create") && (
              <Button htmlType="submit" type="primary">
                {mode === "ubah" ? "Simpan Perubahan" : "Buat Jadwal"}
              </Button>
            )}
            <Button onClick={onBatal} type="default">
              Batal
            </Button>
          </div>
        </Form>
      </Card>
    </div>
  );
}
