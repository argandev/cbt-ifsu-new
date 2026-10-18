// BackupPage — daftar backup + buat / unduh / restore / hapus (Tim Fase 2).
// Restore memakai Dialog konfirmasi dengan peringatan keras: restore
// menimpa SEMUA data saat ini (backup pengaman dibuat otomatis).
// NOL emoji; ikon lucide-react; primary token anbk via Button.
import { useState } from "react";
import { Link } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Download,
  Plus,
  RefreshCw,
  Trash2,
} from "lucide-react";
import apiClient, { ApiError } from "../../../lib/api";
import { formatTanggalJam } from "../../../lib/format";

import { Alert, Button, Card, Input, Modal, Table, Tag, Breadcrumb, Space, Typography } from "antd";
interface BackupMeta {
  id: string;
  nama_file: string;
  ukuran_bytes: number;
  jenis: "manual" | "pre-restore";
  dibuat_oleh: string | null;
  pembuat_nama?: string | null;
  keterangan: string | null;
  created_at: string;
}

interface RestoreResult {
  dipulihkan_dari: string;
  pre_restore_backup: { id: string; nama_file: string; ukuran_bytes: number; created_at: string };
}

function formatBytes(n: number): string {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  return `${(n / (1024 * 1024)).toFixed(2)} MB`;
}

export default function BackupPage() {
  const queryClient = useQueryClient();
  const [pesan, setPesan] = useState("");
  const [pesanError, setPesanError] = useState("");
  const [keterangan, setKeterangan] = useState("");
  const [offset, setOffset] = useState(0);
  const limit = 50;

  const query = useQuery({
    queryKey: ["backup", offset],
    queryFn: () =>
      apiClient.get<BackupMeta[]>(`/api/v1/backup?limit=${limit}&offset=${offset}`),
  });

  const invalidate = () =>
    queryClient.invalidateQueries({ queryKey: ["backup"] });

  const tampilPesan = (msg: string) => {
    setPesan(msg);
    setPesanError("");
  };
  const tampilError = (err: unknown) => {
    setPesanError(err instanceof Error ? err.message : "Terjadi kesalahan.");
    setPesan("");
  };

  const createMutation = useMutation({
    mutationFn: () =>
      apiClient.post<{ id: string; nama_file: string; ukuran_bytes: number }>(
        "/api/v1/backup",
        { keterangan: keterangan.trim() || undefined },
      ),
    onSuccess: (data) => {
      tampilPesan(`Backup dibuat: ${data.nama_file} (${formatBytes(data.ukuran_bytes)}).`);
      setKeterangan("");
      setOffset(0);
      void invalidate();
    },
    onError: tampilError,
  });

  const restoreMutation = useMutation({
    mutationFn: (id: string) =>
      apiClient.post<RestoreResult>(`/api/v1/backup/${id}/restore`, {}),
    onSuccess: (data) => {
      tampilPesan(
        `Restore selesai dari ${data.dipulihkan_dari}. Backup pengaman: ${data.pre_restore_backup.nama_file}.`,
      );
      void invalidate();
    },
    onError: (err) => {
      tampilError(err);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) =>
      apiClient.delete<{ deleted: boolean; id: string; nama_file: string }>(
        `/api/v1/backup/${id}`,
      ),
    onSuccess: (data) => {
      tampilPesan(`Backup dihapus: ${data.nama_file}.`);
      void invalidate();
    },
    onError: (err) => {
      tampilError(err);
    },
  });

  const items = query.data ?? [];
  const mutationBusy =
    restoreMutation.isPending || createMutation.isPending || deleteMutation.isPending;

  function konfirmasiRestore(b: BackupMeta) {
    Modal.confirm({
      title: "Konfirmasi restore",
      content: (
        <div className="flex flex-col gap-3">
          <Alert
            type="error"
            showIcon
            message="Restore menimpa SEMUA data saat ini — backup pengaman dibuat otomatis sebelum restore dimulai."
          />
          <p className="text-sm text-ink">
            Database akan dikembalikan ke isi file{" "}
            <span className="break-all font-mono text-xs">{b.nama_file}</span>.
            Semua perubahan setelah backup itu dibuat akan hilang. Restore ditolak
            bila ada jadwal aktif/pause atau peserta sedang mengerjakan ujian.
          </p>
        </div>
      ),
      okText: "Ya, restore sekarang",
      cancelText: "Batal",
      okButtonProps: { danger: true },
      onOk: () => restoreMutation.mutate(b.id),
    });
  }

  function konfirmasiHapus(b: BackupMeta) {
    Modal.confirm({
      title: "Konfirmasi hapus backup",
      content: (
        <div className="flex flex-col gap-3">
          <Alert
            type="warning"
            showIcon
            message="File backup dan metadatanya akan dihapus permanen. Pre-restore backup berusia kurang dari 7 hari tidak boleh dihapus."
          />
          <p className="break-all font-mono text-xs text-ink">{b.nama_file}</p>
        </div>
      ),
      okText: "Ya, hapus",
      cancelText: "Batal",
      okButtonProps: { danger: true },
      onOk: () => deleteMutation.mutate(b.id),
    });
  }

  return (
    <div className="flex flex-col gap-4">
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: "Backup & Restore" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>Backup & Restore</Typography.Title>
          <Typography.Text type="secondary">Backup manual database. Restore menimpa SEMUA data saat ini — backup pengaman dibuat otomatis sebelum setiap restore.</Typography.Text>
        </div>
        <Space>
          <Button onClick={() => {
              setPesan("");
              setPesanError("");
              createMutation.mutate();
            }} loading={createMutation.isPending} type="primary">
            <Plus size={18} />
            {createMutation.isPending ? "Membuat backup..." : "Buat Backup"}
          </Button>
        </Space>
      </div>

      {pesan && (
        <Alert type="success" showIcon message={<>
          {pesan}
        </>} />
      )}
      {pesanError && (
        <Alert type="error" showIcon message={<>
          {pesanError}
        </>} />
      )}

      <Card title="Buat backup manual">
        
        <div className="flex flex-col gap-2">
          <label className="text-sm font-semibold text-ink" htmlFor="backup-keterangan">
            Keterangan (opsional)
          </label>
          <Input id="backup-keterangan" value={keterangan} onChange={(e) => setKeterangan(e.target.value)} placeholder="Mis. Sebelum pembaruan jadwal ujian" maxLength={500} size="large" />
          <p className="text-xs text-muted">
            Backup otomatis harian tidak dibuat di aplikasi — jadwal cron memakai
            skrip deploy/backup.sh. Backup di sini hanya manual.
          </p>
        </div>
      </Card>

      <Card title="Daftar backup">
        
        <div className="p-0 sm:p-0">
          {query.isLoading ? (
            <p className="p-5 text-sm text-muted">Memuat daftar backup...</p>
          ) : query.isError ? (
            <Alert type="error" showIcon message={<>Gagal memuat daftar backup:{" "}
              {query.error instanceof ApiError
                ? query.error.message
                : "Terjadi kesalahan."}</>} className="m-5" />
          ) : items.length === 0 ? (
            <p className="p-5 text-sm text-muted">
              Belum ada backup. Klik &quot;Buat Backup&quot; untuk membuat yang pertama.
            </p>
          ) : (
              <Table columns={[{ key: "col0", title: "Nama file", dataIndex: "nama_file", sorter: (a: any, b: any) => { const _va = a.nama_file, _vb = b.nama_file; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, b: any) => (<>
                        {b.nama_file}
                      </>) }, { key: "col1", title: "Jenis", render: (_: unknown, b: any) => (<>
                        <Tag color="default">{b.jenis === "pre-restore" ? "Pre-restore" : "Manual"}</Tag>
                      </>) }, { key: "col2", title: "Ukuran", render: (_: unknown, b: any) => (<>{formatBytes(b.ukuran_bytes)}</>) }, { key: "col3", title: "Dibuat", render: (_: unknown, b: any) => (<>
                        {formatTanggalJam(b.created_at)}
                      </>) }, { key: "col4", title: "Pembuat", dataIndex: "pembuat_nama", sorter: (a: any, b: any) => { const _va = a.pembuat_nama, _vb = b.pembuat_nama; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, b: any) => (<>{b.pembuat_nama ?? "-"}</>) }, { key: "col5", title: "Keterangan", dataIndex: "keterangan", sorter: (a: any, b: any) => { const _va = a.keterangan, _vb = b.keterangan; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, b: any) => (<>
                        {b.keterangan ?? "-"}
                      </>) }, { key: "col6", title: "Aksi", render: (_: unknown, b: any) => (<>
                        <div className="flex flex-wrap gap-2">
                          <Button onClick={() =>
                              apiClient.download(
                                `/api/v1/backup/${b.id}/download`,
                                b.nama_file,
                              )
                            } type="default">
                            <Download size={16} /> Unduh
                          </Button>
                          <Button disabled={mutationBusy} onClick={() => konfirmasiRestore(b)} type="default">
                            <RefreshCw size={16} /> Restore
                          </Button>
                          <Button loading={deleteMutation.isPending} onClick={() => konfirmasiHapus(b)} type="default" danger>
                            <Trash2 size={16} /> Hapus
                          </Button>
                        </div>
                      </>) }]} dataSource={items} rowKey="id" pagination={false} locale={{ emptyText: "Belum ada data" }} scroll={{ x: "max-content" }} />
          )}
          <div className="flex items-center gap-2 p-4">
            <Button disabled={offset === 0} onClick={() => setOffset((o) => Math.max(0, o - limit))} type="default">
              Sebelumnya
            </Button>
            <Button disabled={items.length < limit} onClick={() => setOffset((o) => o + limit)} type="default">
              Berikutnya
            </Button>
          </div>
        </div>
      </Card>


    </div>
  );
}
