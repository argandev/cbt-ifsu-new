import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { Lock, LockOpen, Plus, Search } from "lucide-react";
import apiClient from "../../../lib/api";
import { useAuthStore } from "../../../store/auth";
import type { BankSoal, PageData } from "../../../lib/banksoal";

import { Alert, Button, Input, Table, Form, Breadcrumb, Tag, Typography } from "antd";
export default function BankSoalListPage() {
  const { hasPermission } = useAuthStore();
  const [searchInput, setSearchInput] = useState("");
  const [search, setSearch] = useState("");

  const banksQuery = useQuery({
    queryKey: ["bank-soal", search],
    queryFn: () =>
      apiClient.get<PageData<BankSoal>>(`/api/v1/bank-soal?q=${encodeURIComponent(search)}&limit=100`),
  });

  const banks = banksQuery.data?.items ?? [];

  return (
    <div className="space-y-6">
      <Breadcrumb items={[{ title: (<Link to="/adm">Dashboard</Link>) }, { title: "Bank Soal" }]} />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>Bank Soal</Typography.Title>
          <Typography.Text type="secondary">
            5 tipe standar AKM + 11 varian/ekstensi aplikasi. Bank terkunci tidak bisa diubah isinya.
          </Typography.Text>
        </div>
        {hasPermission("banksoal.create") && (
          <Button href="/adm/bank-soal/baru" type="primary" icon={<Plus size={16} />}>
            Buat Bank Soal
          </Button>
        )}
      </div>

      <Form layout="vertical" onFinish={() => setSearch(searchInput)} className="flex flex-col gap-2 sm:flex-row sm:items-end">
        <Form.Item label="Cari bank soal" className="mb-0 flex-1">
          <Input id="cari-bank" value={searchInput} onChange={(e) => setSearchInput(e.target.value)} placeholder="Cari kode / nama bank…" size="large" />
        </Form.Item>
        <Button htmlType="submit" type="primary">
          <Search size={16} /> Cari
        </Button>
      </Form>

      {banksQuery.isError && (
        <Alert type="error" showIcon message={<>
          {(banksQuery.error as Error).message}
        </>} />
      )}
      {banksQuery.isLoading && <p className="text-sm text-muted">Memuat bank soal…</p>}

      {banks.length === 0 && !banksQuery.isLoading && (
        <p className="rounded-lg border border-dashed border-border px-4 py-8 text-center text-sm text-muted">
          Belum ada bank soal. Buat bank soal baru untuk mulai menyusun soal.
        </p>
      )}

        <Table columns={[{ key: "col0", title: "Kode", dataIndex: "kode", sorter: (a: any, b: any) => { const _va = a.kode, _vb = b.kode; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, b: any) => (<>{b.kode}</>) }, { key: "col1", title: "Nama Bank", dataIndex: "nama", sorter: (a: any, b: any) => { const _va = a.nama, _vb = b.nama; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, b: any) => (<>{b.nama}</>) }, { key: "col2", title: "Kelas", dataIndex: "kelas_tingkat", sorter: (a: any, b: any) => { const _va = a.kelas_tingkat, _vb = b.kelas_tingkat; if (typeof _va === "number" && typeof _vb === "number") return _va - _vb; return String(_va ?? "").localeCompare(String(_vb ?? ""), "id"); }, render: (_: unknown, b: any) => (<>{b.kelas_tingkat}</>) }, { key: "col3", title: "Jumlah Soal", render: (_: unknown, b: any) => (<>{b.jumlah_soal} soal</>) }, { key: "col4", title: "Status", render: (_: unknown, b: any) => (<>
                  {b.is_locked ? (
                    <Tag color="gold"><Lock size={12} /> Terkunci</Tag>
                  ) : (
                    <Tag color="green"><LockOpen size={12} /> Terbuka</Tag>
                  )}
                </>) }, { key: "col5", title: "Aksi", render: (_: unknown, b: any) => (<>
                  {/* proxy: tidak ada permission bernama untuk "lihat isi soal";
                      banksoal.create dipakai sebagai proxy gate karena proktor
                      (banksoal.read) dilarang lihat isi soal oleh
                      forbidSoalUntukProktor di backend. */}
                  {hasPermission("banksoal.create") ? (
                    <Link to={`/adm/bank-soal/${b.id}`} className="font-medium text-anbk hover:underline">Kelola Soal</Link>
                  ) : (
                    <span className="text-muted">—</span>
                  )}
                </>) }]} dataSource={banks} rowKey="id" pagination={false} locale={{ emptyText: "Belum ada data" }} scroll={{ x: "max-content" }} />
    </div>
  );
}
