// KartuSayaPage — kartu peserta milik siswa sendiri (model 1-KARTU).
//
// Model baru: 1 kartu per siswa berlaku untuk SEMUA ujian (bukan 1 kartu
// per jadwal). Kartu diunduh sebagai PDF dari backend:
//   GET /api/v1/kartu/siswa/:siswaId/pdf   (siswaId = ID akun login)
// File ini TIDAK memakai cetak browser — unduhan lewat tautan ke PDF
// backend (tombol antd dengan href).
//
// Pratinjau ringkas: identitas dari sesi login + daftar jadwal hari ini
// dari GET /api/v1/ujian/aktif (endpoint yang sama dipakai beranda siswa).
// Kartu TIDAK berisi token ujian (token hanya dari proktor saat ujian).
import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { Alert, Breadcrumb, Button, Card, Descriptions, Empty, Table, Typography } from "antd";
import { Download } from "lucide-react";
import apiClient from "../../lib/api";
import { formatJam, formatTanggal } from "../../lib/format";
import { useAuthStore } from "../../store/auth";

interface UjianAktif {
  jadwal_id: string;
  paket_nama: string;
  paket_jenis: string;
  tanggal: string;
  jam_mulai: string;
  jam_selesai: string;
  kelas_nama: string;
}

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? "";

export default function KartuSayaPage() {
  const user = useAuthStore((s) => s.user);

  // Daftar jadwal hari ini — hanya untuk pratinjau ringkas.
  const jadwalQuery = useQuery({
    queryKey: ["ujian-aktif"],
    queryFn: () => apiClient.get<{ items?: UjianAktif[] } | UjianAktif[]>("/api/v1/ujian/aktif"),
    retry: false,
  });
  const jadwalList: UjianAktif[] = Array.isArray(jadwalQuery.data)
    ? jadwalQuery.data
    : (jadwalQuery.data?.items ?? []);

  const pdfUrl = user ? `${API_BASE}/api/v1/kartu/siswa/${user.id}/pdf` : "";

  return (
    <div className="flex flex-col gap-6">
      <div className="rounded-lg bg-anbk px-5 py-4 text-white">
        <p className="text-lg font-bold tracking-wide">CBT SMK</p>
        <p className="text-sm text-white">Ujian Berbasis Komputer</p>
      </div>

      <div>
        <Breadcrumb items={[{ title: <Link to="/">Beranda</Link> }, { title: "Kartu Saya" }]} />
        <Typography.Title level={3} style={{ margin: "8px 0 0" }}>
          Kartu Saya
        </Typography.Title>
        <p className="mt-1 text-sm text-muted">
          Satu kartu untuk semua ujian. Bawa kartu ini saat ujian; token ujian diberikan proktor di ruangan,
          bukan tercantum di kartu ini.
        </p>
      </div>

      <Card>
        <div className="flex flex-col gap-3">
          <Alert
            type="info"
            showIcon
            message="Unduh kartu Anda sebagai PDF, lalu cetak atau simpan di perangkat."
          />
          <div>
            <Button
              type="primary"
              size="large"
              href={pdfUrl || undefined}
              target="_blank"
              rel="noreferrer"
              disabled={!user}
              icon={<Download size={16} />}
            >
              Unduh Kartu (PDF)
            </Button>
          </div>
          {!user && (
            <Alert type="error" showIcon message="Sesi login tidak ditemukan. Silakan login ulang." />
          )}
        </div>
      </Card>

      <Card title="Identitas Peserta">
        <Descriptions
          column={1}
          size="small"
          items={[
            { key: "nama", label: "Nama Peserta", children: user?.nama ?? "—" },
            { key: "username", label: "Nama Pengguna", children: user?.username ?? "—" },
          ]}
        />
      </Card>

      <Card title="Jadwal Ujian Hari Ini">
        {jadwalQuery.isPending ? (
          <p className="text-muted">Memuat jadwal...</p>
        ) : jadwalQuery.isError ? (
          <Alert
            type="error"
            showIcon
            message={`Jadwal belum bisa dimuat (${(jadwalQuery.error as Error).message}).`}
          />
        ) : jadwalList.length === 0 ? (
          <Empty description="Tidak ada jadwal ujian hari ini." />
        ) : (
          <Table
            columns={[
              { key: "paket", title: "Paket Ujian", dataIndex: "paket_nama" },
              { key: "jenis", title: "Jenis", dataIndex: "paket_jenis" },
              { key: "kelas", title: "Kelas", dataIndex: "kelas_nama" },
              {
                key: "tanggal",
                title: "Tanggal",
                render: (_: unknown, j: UjianAktif) => formatTanggal(j.tanggal),
              },
              {
                key: "jam",
                title: "Jam",
                render: (_: unknown, j: UjianAktif) =>
                  `${formatJam(j.jam_mulai)}–${formatJam(j.jam_selesai)} WIB`,
              },
            ]}
            dataSource={jadwalList}
            rowKey="jadwal_id"
            pagination={false}
            scroll={{ x: "max-content" }}
          />
        )}
      </Card>
    </div>
  );
}
