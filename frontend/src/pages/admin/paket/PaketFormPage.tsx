// PaketFormPage — halaman form buat/ubah paket (Tim B, penataan).
// Tampilan form SAJA: dipanggil dari PaketPage saat mode "form", bukan lewat
// route baru (App.tsx bukan wewenang Tim B — route /adm/paket tetap daftar).
// Panel "Soal dalam Paket" TIDAK ada di sini — sudah pindah ke
// PaketDetailPage. Semua field, Dropdown, checkbox, dan payload TIDAK berubah.

import { Alert, Button, Checkbox, Input, InputNumber, Select, Breadcrumb, Typography, Form, Card } from "antd";
import { Link } from "react-router-dom";
export interface PaketFormValues {
  kode: string;
  nama: string;
  mapel_id: string;
  jenis: string;
  durasi_menit: number;
  jumlah_soal: number;
  acak_soal: boolean;
  acak_jawaban: boolean;
  tampil_hasil: boolean;
}

interface Mapel {
  id: string;
  kode: string;
  nama: string;
}

const JENIS = ["STS", "SAS", "USBK", "UKK", "Tryout"];

interface PaketFormPageProps {
  mode: "buat" | "ubah";
  form: PaketFormValues;
  setForm: (form: PaketFormValues) => void;
  mapelList: Mapel[];
  mapelError: boolean;
  error: string | null;
  onSubmit: () => void;
  onBatal: () => void;
}

export default function PaketFormPage({
  mode,
  form,
  setForm,
  mapelList,
  mapelError,
  error,
  onSubmit,
  onBatal,
}: PaketFormPageProps) {
  return (
    <div className="space-y-4">
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: "Paket Ujian" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>{mode === "ubah" ? "Ubah Paket" : "Buat Paket"}</Typography.Title>
          <Typography.Text type="secondary">Atur identitas paket, mapel, durasi, dan pengacakan. Soal diisi dari halaman detail paket.</Typography.Text>
        </div>
      </div>
      {error && <Alert type="error" showIcon message={<>{error}</>} />}
      <Card className="max-w-2xl">
        <Form layout="vertical" onFinish={onSubmit}>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <Form.Item label="Kode"><Input value={form.kode} onChange={(e) => setForm({ ...form, kode: e.target.value })} required size="large" /></Form.Item>
              <Form.Item label="Jenis"><Select options={JENIS.map((j) => ({ value: j, label: j }))} value={form.jenis} onChange={(v) => setForm({ ...form, jenis: v })} id="paket-jenis" placeholder="Pilih…" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" /></Form.Item>
          </div>
          <Form.Item label="Nama paket"><Input value={form.nama} onChange={(e) => setForm({ ...form, nama: e.target.value })} required size="large" /></Form.Item>
            <Form.Item label="Mata pelajaran"><Select showSearch options={[
                { value: "", label: "— Pilih mata pelajaran —" },
                ...mapelList.map((m) => ({ value: m.id, label: `${m.nama} (${m.kode})` })),
              ]} value={form.mapel_id} onChange={(v) => setForm({ ...form, mapel_id: v })} placeholder="— Pilih mata pelajaran —" id="paket-mapel" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" /></Form.Item>
            {mapelError && (
              <Typography.Text type="danger" className="mt-1 block text-xs">Gagal memuat daftar mapel.</Typography.Text>
            )}
          <Form.Item label="Durasi (menit)" rules={[{ required: true, message: "Durasi wajib diisi." }]}><InputNumber min={1} value={form.durasi_menit} onChange={(v) => setForm({ ...form, durasi_menit: typeof v === "number" ? v : 0 })} size="large" style={{ width: "100%" }} /></Form.Item>
          <div className="space-y-2 text-sm">
            <Form.Item style={{ marginBottom: 0 }}><Checkbox checked={form.acak_soal} onChange={(e) => setForm({ ...form, acak_soal: e.target.checked })}>Acak urutan soal per siswa</Checkbox></Form.Item>
            <Form.Item style={{ marginBottom: 0 }}><Checkbox checked={form.acak_jawaban} onChange={(e) => setForm({ ...form, acak_jawaban: e.target.checked })}>Acak urutan opsi jawaban per siswa</Checkbox></Form.Item>
            <Form.Item style={{ marginBottom: 0 }}><Checkbox checked={form.tampil_hasil} onChange={(e) => setForm({ ...form, tampil_hasil: e.target.checked })}>Tampilkan nilai ke siswa (perlu settings tampil_hasil_ke_siswa juga)</Checkbox></Form.Item>
          </div>
          <div className="flex flex-wrap gap-2 pt-1">
            <Button htmlType="submit" type="primary">
              {mode === "ubah" ? "Simpan Perubahan" : "Buat Paket"}
            </Button>
            <Button onClick={onBatal} type="default">
              Batal
            </Button>
          </div>
        </Form>
      </Card>
    </div>
  );
}
