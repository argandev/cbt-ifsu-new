// ImportSiswaPage — Import Siswa dari CSV.
// Milik Tim 2. Lihat catatan kontrak apiClient di JurusanPage.tsx.
// Upload memakai FormData lewat apiClient.post (apiClient Tim 1 WAJIB
// meneruskan FormData apa adanya, tanpa header Content-Type JSON,
// agar boundary multipart tetap benar + header X-CSRF-Token tetap dikirim).
// Format CSV: nama,nisn,nis,kelas_kode  ATAU  nama,nisn,kelas_id.
// kelas_kode = ID UUID kelas ATAU nama kelas (mis. "XI RPL 1").
// Tiap siswa yang berhasil ikut dibuatkan akun login (username = NISN,
// password awal = NISN, wajib ganti saat login pertama).
import { useState } from "react";
import { Link } from "react-router-dom";
import { useMutation } from "@tanstack/react-query";
import { UploadOutlined } from "@ant-design/icons";
import { apiClient } from "../../../lib/api";
import { useAuthStore } from "../../../store/auth";

import { Alert, Button, Table, Breadcrumb, Space, Typography, Statistic, Card, Upload } from "antd";
type ImportError = { baris: number; pesan: string };
type ImportResult = { sukses: number; gagal: number; errors: ImportError[]; catatan?: string };

function apiErrorMessage(err: unknown): string {
  if (typeof err === "object" && err !== null) {
    const e = err as { message?: string; error?: { message?: string } };
    if (e.error?.message) return e.error.message;
    if (e.message) return e.message;
  }
  return "Terjadi kesalahan. Coba lagi.";
}



const TEMPLATE_CSV = "nama,nisn,nis,kelas_kode\nAndi Pratama,1234567890,001,XI RPL 1\nBella Sari,2234567890,002,XI RPL 1\n";

export default function ImportSiswaPage() {
  const { hasPermission } = useAuthStore();
  const [file, setFile] = useState<File | null>(null);
  const [result, setResult] = useState<ImportResult | null>(null);
  const [error, setError] = useState("");

  const uploadMutation = useMutation({
    mutationFn: async () => {
      if (!file) throw new Error("Pilih file CSV dulu.");
      const form = new FormData();
      form.append("file", file);
      const res = await apiClient.post<{ data: ImportResult }>("/api/v1/master/siswa/import", form);
      const data = (res as { data?: ImportResult }).data ?? (res as unknown as ImportResult);
      return data;
    },
    onSuccess: (data) => {
      setResult(data);
      setError("");
    },
    onError: (err) => setError(apiErrorMessage(err)),
  });

  function downloadTemplate() {
    const blob = new Blob([TEMPLATE_CSV], { type: "text/csv;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "template-import-siswa.csv";
    a.click();
    URL.revokeObjectURL(url);
  }

  return (
    <div className="space-y-4">
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: (<Link to="/adm/master/siswa">Siswa</Link>) }, { title: "Import CSV" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>Import Siswa (CSV)</Typography.Title>
          <Typography.Text type="secondary">Import siswa massal dari file CSV. Baris yang gagal dilewati dan dilaporkan.</Typography.Text>
        </div>
        <Space>
          <Button href="/adm/master/siswa">
            Kembali ke Daftar Siswa
          </Button>
        </Space>
      </div>
      <p className="text-sm text-muted">
        Unggah CSV dengan header <code className="rounded-lg bg-surface px-1">nama,nisn,nis,kelas_kode</code> atau{" "}
        <code className="rounded-lg bg-surface px-1">nama,nisn,kelas_id</code>. Kolom <code className="rounded-lg bg-surface px-1">nis</code>{" "}
        opsional. <code className="rounded-lg bg-surface px-1">kelas_kode</code> diisi ID kelas atau nama kelas (mis.{" "}
        <code className="rounded-lg bg-surface px-1">XI RPL 1</code>). Maksimal 2.000 baris per file, 5 MB. Baris yang
        gagal dilewati dan dilaporkan; baris yang valid tetap disimpan.
      </p>

      <Card>
        <div className="flex flex-col gap-3">
          <div className="block">
            <Typography.Text strong className="mb-1 block">File CSV *</Typography.Text>
            <Upload
              accept=".csv,text/csv"
              beforeUpload={(f) => {
                setFile(f);
                setResult(null);
                setError("");
                return false;
              }}
              maxCount={1}
              showUploadList={false}
            >
              <Button icon={<UploadOutlined />}>Pilih File CSV</Button>
            </Upload>
          </div>
          {file && <p className="text-sm text-muted">File dipilih: {file.name}</p>}
          <div className="flex flex-col gap-2 sm:flex-row">
            <Button onClick={downloadTemplate} type="default">
              Unduh Template CSV
            </Button>
            {hasPermission("master.siswa.import") && (
              <Button disabled={!file} loading={uploadMutation.isPending} onClick={() => uploadMutation.mutate()} type="primary">
                {uploadMutation.isPending ? "Mengunggah…" : "Upload & Import"}
              </Button>
            )}
          </div>
        </div>
      </Card>

      {error && (
        <Alert type="error" showIcon message={error} />
      )}

      {result && (
        <div className="space-y-3">
          <div className="flex flex-col gap-2 sm:flex-row">
            <div className="flex-1">
              <Card>
                <Statistic title="Berhasil" value={result.sukses} />
              </Card>
            </div>
            <div className="flex-1">
              <Card>
                <Statistic title="Gagal" value={result.gagal} />
              </Card>
            </div>
          </div>
          {result.catatan && (
            <Alert type="warning" showIcon message={result.catatan} />
          )}
          {result.errors.length > 0 && (
              <Table columns={[{ key: "col0", title: "Baris", dataIndex: "baris", sorter: (a: any, b: any) => { const _va = a.baris, _vb = b.baris; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, e: any) => (<>{e.baris}</>) }, { key: "col1", title: "Pesan Kesalahan", dataIndex: "pesan", sorter: (a: any, b: any) => { const _va = a.pesan, _vb = b.pesan; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, e: any) => (<>{e.pesan}</>) }]} dataSource={result.errors} rowKey={(_record: any, _index?: number) => _record?.id ?? _index} pagination={false} locale={{ emptyText: "Belum ada data" }} scroll={{ x: "max-content" }} />
          )}
          {result.errors.length === 0 && result.sukses > 0 && (
            <p className="text-sm text-muted">Semua baris berhasil diimport.</p>
          )}
        </div>
      )}
    </div>
  );
}
