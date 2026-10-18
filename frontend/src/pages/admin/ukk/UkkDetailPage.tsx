import { useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Award, Download, PenLine, Upload as UploadIcon } from "lucide-react";
import apiClient from "../../../lib/api";
import type { MasterOption, Ukk, UkkBerkas, UkkNilai } from "./types";
import { fmtNilai, fmtTanggal, fmtUkuran } from "./types";

import { Alert, Button, Card, Select, Table, Tag, Breadcrumb, Space, Typography, Form, Upload } from "antd";
import { statusTagColor } from "../../../components/common/statusTag";
const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "";

function getCsrfToken(): string {
  const match = document.cookie.match(/(?:^|;\s*)cbt_csrf=([^;]*)/);
  return match ? decodeURIComponent(match[1]) : "";
}

/** Upload multipart berkas praktik (apiClient hanya JSON, jadi fetch langsung + CSRF). */
async function uploadBerkas(ukkId: string, siswaId: string, file: File): Promise<UkkBerkas> {
  const form = new FormData();
  form.append("siswa_id", siswaId);
  form.append("file", file);
  const res = await fetch(`${BASE_URL}/api/v1/ukk/${ukkId}/berkas`, {
    method: "POST",
    credentials: "include",
    headers: { "X-CSRF-Token": getCsrfToken() },
    body: form,
  });
  const json = await res.json().catch(() => null);
  if (!res.ok) {
    throw new Error(json?.error?.message ?? `Upload gagal (${res.status})`);
  }
  return (json as { data: UkkBerkas }).data;
}


export default function UkkDetailPage() {
  const { id = "" } = useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [pesan, setPesan] = useState("");
  const [statusBaru, setStatusBaru] = useState("");
  const [uploadSiswa, setUploadSiswa] = useState("");
  const [uploadFile, setUploadFile] = useState<File | null>(null);

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
  const berkasQuery = useQuery({
    queryKey: ["ukk-berkas", id],
    queryFn: () => apiClient.get<UkkBerkas[]>(`/ukk/${id}/berkas`),
    enabled: id !== "",
  });
  const siswaQuery = useQuery({
    queryKey: ["ukk-siswa", ukkQuery.data?.kelas_id],
    queryFn: () =>
      apiClient.get<MasterOption[]>(
        `/master/siswa?per_page=100&kelas_id=${encodeURIComponent(ukkQuery.data?.kelas_id ?? "")}`,
      ),
    enabled: Boolean(ukkQuery.data?.kelas_id),
  });

  const ukk = ukkQuery.data;

  const statusMutation = useMutation({
    mutationFn: (status: string) =>
      apiClient.put<Ukk>(`/ukk/${id}`, {
        nama: ukk?.nama,
        mapel_id: ukk?.mapel_id,
        kelas_id: ukk?.kelas_id,
        jurusan_id: ukk?.jurusan_id,
        tanggal: ukk?.tanggal ? ukk.tanggal.slice(0, 10) : null,
        penguji_internal: ukk?.penguji_internal,
        penguji_eksternal: ukk?.penguji_eksternal,
        status,
      }),
    onSuccess: () => {
      setPesan("Status UKK diperbarui.");
      void queryClient.invalidateQueries({ queryKey: ["ukk", id] });
      void queryClient.invalidateQueries({ queryKey: ["ukk"] });
    },
    onError: (err: Error) => setPesan(err.message),
  });

  const sertifikatMutation = useMutation({
    mutationFn: (siswaId: string) =>
      apiClient.post(`/ukk/${id}/sertifikat/${siswaId}`),
    onSuccess: (_data, siswaId) => {
      navigate(`/adm/ukk/${id}/sertifikat/${siswaId}`);
    },
    onError: (err: Error) => setPesan(err.message),
  });

  const uploadMutation = useMutation({
    mutationFn: ({ siswaId, file }: { siswaId: string; file: File }) =>
      uploadBerkas(id, siswaId, file),
    onSuccess: () => {
      setPesan("Berkas praktik berhasil diupload.");
      setUploadFile(null);
      void queryClient.invalidateQueries({ queryKey: ["ukk-berkas", id] });
    },
    onError: (err: Error) => setPesan(err.message),
  });

  const nilaiRows = nilaiQuery.data ?? [];
  const berkasRows = berkasQuery.data ?? [];
  const siswaOptions = siswaQuery.data ?? [];
  const jumlahLulus = nilaiRows.filter((n) => n.lulus === true).length;

  const handleUpload = () => {
    setPesan("");
    if (!uploadSiswa) {
      setPesan("Pilih siswa pemilik berkas.");
      return;
    }
    if (!uploadFile) {
      setPesan("Pilih berkas yang akan diupload (maks. 10 MB: pdf, doc, docx, jpg, png, zip).");
      return;
    }
    uploadMutation.mutate({ siswaId: uploadSiswa, file: uploadFile });
  };

  if (ukkQuery.isLoading) return <p className="p-1 text-muted">Memuat...</p>;
  if (ukkQuery.isError || !ukk) {
    return <Alert type="error" showIcon message={`UKK tidak ditemukan: ${(ukkQuery.error as Error)?.message ?? ""}`} />;
  }

  return (
    <div className="flex flex-col gap-4">
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: "UKK" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>Detail UKK</Typography.Title>
          <Typography.Text type="secondary">{`${ukk.nama} · ${ukk.mapel_nama ?? "-"} · ${ukk.kelas_nama ?? "-"} · ${ukk.jurusan_nama ?? ukk.jurusan_kode ?? "-"} · ${fmtTanggal(ukk.tanggal)}`}</Typography.Text>
        </div>
        <Space>
          <Tag color={statusTagColor(ukk.status)}>{ukk.status}</Tag>
        </Space>
      </div>

      {pesan ? <Alert type="info" showIcon message={pesan} /> : null}

      <Card title="Penguji &amp; Status">
        
        <div className="flex flex-col gap-3">
          <p className="text-sm text-ink">
            Penguji internal: <strong>{ukk.penguji_internal || "-"}</strong>
            <br />
            Penguji eksternal (DUDIKA): <strong>{ukk.penguji_eksternal || "-"}</strong>
            <br />
            Peserta dinilai: <strong>{nilaiRows.length}</strong> &middot; Lulus:{" "}
            <strong>{jumlahLulus}</strong>
          </p>
          <Form
            layout="vertical"
            onFinish={() => {
              if (statusBaru) statusMutation.mutate(statusBaru);
            }}
            className="flex flex-wrap items-end gap-3"
          >
            <div className="w-full max-w-xs">
              <Form.Item label="Ubah status" className="mb-0 w-full max-w-xs"><Select showSearch options={[
                  { value: "", label: "Pilih status baru" },
                  { value: "draft", label: "Draft" },
                  { value: "berjalan", label: "Berjalan" },
                  { value: "selesai", label: "Selesai" },
                ]} value={statusBaru} onChange={(v) => setStatusBaru((v ?? ""))} placeholder="Pilih…" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" /></Form.Item>
            </div>
            <Button htmlType="submit" disabled={!statusBaru} loading={statusMutation.isPending} type="primary">
              Simpan Status
            </Button>
            <Button href={`/adm/ukk/${id}/nilai`} type="primary" icon={<PenLine size={16} />}>
              Input Nilai
            </Button>
          </Form>
          <p className="text-xs text-muted">
            Nilai ditolak bila UKK sudah selesai, kecuali diubah oleh admin (dicatat di audit log).
          </p>
        </div>
      </Card>

      <Card title="Nilai Peserta">
        
        <div>
            <Table columns={[{ key: "col0", title: "Siswa", dataIndex: "siswa_nama", sorter: (a: any, b: any) => { const _va = a.siswa_nama, _vb = b.siswa_nama; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, n: any) => (<>{n.siswa_nama ?? "-"}</>) }, { key: "col1", title: "NISN", dataIndex: "siswa_nisn", sorter: (a: any, b: any) => { const _va = a.siswa_nisn, _vb = b.siswa_nisn; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, n: any) => (<>{n.siswa_nisn ?? "-"}</>) }, { key: "col2", title: "Teori (40%)", render: (_: unknown, n: any) => (<>{fmtNilai(n.nilai_teori)}</>) }, { key: "col3", title: "Proses (30%)", render: (_: unknown, n: any) => (<>{fmtNilai(n.nilai_proses)}</>) }, { key: "col4", title: "Produk (50%)", render: (_: unknown, n: any) => (<>{fmtNilai(n.nilai_produk)}</>) }, { key: "col5", title: "Sikap (20%)", render: (_: unknown, n: any) => (<>{fmtNilai(n.nilai_sikap)}</>) }, { key: "col6", title: "Praktik (60%)", render: (_: unknown, n: any) => (<>{fmtNilai(n.nilai_praktik)}</>) }, { key: "col7", title: "Akhir", render: (_: unknown, n: any) => (<>{fmtNilai(n.nilai_akhir)}</>) }, { key: "col8", title: "Lulus", render: (_: unknown, n: any) => (<>
                      {n.lulus === null || n.lulus === undefined ? (
                        "-"
                      ) : n.lulus ? (
                        <Tag color="green">Lulus</Tag>
                      ) : (
                        <Tag color="red">Tidak Lulus</Tag>
                      )}
                    </>) }, { key: "col9", title: "Predikat", dataIndex: "predikat", sorter: (a: any, b: any) => { const _va = a.predikat, _vb = b.predikat; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, n: any) => (<>{n.predikat ?? "-"}</>) }, { key: "col10", title: "Sertifikat", render: (_: unknown, n: any) => (<>
                      {n.lulus ? (
                        <Button onClick={() => sertifikatMutation.mutate(n.siswa_id)} type="default">
                          <Award size={16} /> Sertifikat
                        </Button>
                      ) : (
                        <span className="text-sm text-muted">-</span>
                      )}
                    </>) }]} dataSource={nilaiRows} rowKey="id" pagination={false} locale={{ emptyText: "Belum ada nilai. Gunakan tombol Input Nilai." }} scroll={{ x: "max-content" }} />
        </div>
      </Card>

      <Card title="Berkas Praktik">
        
        <div className="flex flex-col gap-4">
          <Form layout="vertical" onFinish={handleUpload} className="flex flex-wrap items-end gap-3">
            <div className="w-full max-w-xs">
              <Form.Item label="Siswa pemilik berkas" className="mb-0 w-full max-w-xs"><Select showSearch allowClear options={siswaOptions.map((s) => ({
                  value: s.id,
                  label: s.nisn ? `${s.nama} (${s.nisn})` : s.nama,
                }))} value={uploadSiswa} onChange={(v) => setUploadSiswa((v ?? ""))} placeholder="Pilih siswa" filterOption={(input, opt) => (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())} optionFilterProp="label" notFoundContent="Tidak ada data" style={{ width: "100%" }} size="large" /></Form.Item>
            </div>
            <Form.Item label="Berkas (maks. 10 MB: pdf, doc, docx, jpg, png, zip)" className="mb-0">
              <Upload
                accept=".pdf,.doc,.docx,.jpg,.jpeg,.png,.zip"
                beforeUpload={(f) => {
                  setUploadFile(f);
                  return false;
                }}
                maxCount={1}
                showUploadList={false}
              >
                <Button icon={<UploadIcon size={16} />}>Pilih Berkas</Button>
              </Upload>
              {uploadFile && (
                <Typography.Text type="secondary" className="mt-1 block text-sm">
                  Berkas dipilih: {uploadFile.name}
                </Typography.Text>
              )}
            </Form.Item>
            <Button htmlType="submit" loading={uploadMutation.isPending} type="primary">
              <UploadIcon size={16} /> Upload Berkas
            </Button>
          </Form>
            <Table columns={[{ key: "col0", title: "Siswa", dataIndex: "siswa_nama", sorter: (a: any, b: any) => { const _va = a.siswa_nama, _vb = b.siswa_nama; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, b: any) => (<>{b.siswa_nama ?? "-"}</>) }, { key: "col1", title: "Nama File", dataIndex: "nama_file", sorter: (a: any, b: any) => { const _va = a.nama_file, _vb = b.nama_file; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, b: any) => (<>{b.nama_file}</>) }, { key: "col2", title: "Ukuran", render: (_: unknown, b: any) => (<>{fmtUkuran(b.ukuran)}</>) }, { key: "col3", title: "Diunggah", render: (_: unknown, b: any) => (<>{fmtTanggal(b.created_at)}</>) }, { key: "col4", title: "Aksi", render: (_: unknown, b: any) => (<>
                      <Button onClick={() =>
                          apiClient.download(`/ukk/${id}/berkas/${b.id}/download`, b.nama_file)
                        } type="default">
                        <Download size={16} /> Unduh
                      </Button>
                    </>) }]} dataSource={berkasRows} rowKey="id" pagination={false} locale={{ emptyText: "Belum ada berkas praktik." }} scroll={{ x: "max-content" }} />
        </div>
      </Card>
    </div>
  );
}
