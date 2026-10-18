// DashboardPage — ringkasan /adm.
// Kartu menampilkan angka nyata dari GET /api/v1/dashboard/ringkasan.
// Bila endpoint gagal, kartu menampilkan "-" (tanpa angka palsu).
import { useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { BookOpen, CalendarDays, FileText, Users } from "lucide-react";
import { useAuthStore } from "../../store/auth";
import apiClient from "../../lib/api";

import { Card, Statistic, Typography } from "antd";

interface Ringkasan {
  bank_soal: number;
  jadwal_ujian: number;
  peserta: number;
  rekap: number;
}

const MENU_CEPAT: Array<{ to: string; label: string; perm?: string }> = [
  { to: "/adm/bank-soal", label: "Bank Soal", perm: "banksoal.read" },
  { to: "/adm/paket", label: "Paket Ujian", perm: "paket.read" },
  { to: "/adm/jadwal", label: "Jadwal", perm: "jadwal.read" },
  { to: "/adm/monitor", label: "Monitor", perm: "ujian.monitor" },
  { to: "/adm/rekap", label: "Rekap", perm: "rekap.read" },
  { to: "/adm/analisis", label: "Analisis Butir", perm: "analisis.read" },
  // Nilai Kompetensi memanggil GET /kompetensi/:id yang di-gate backend
  // dengan rekap.read (backend adalah sumber kebenaran).
  { to: "/adm/kompetensi", label: "Nilai Kompetensi", perm: "rekap.read" },
  { to: "/adm/kartu", label: "Kartu Peserta", perm: "event.read" },
  { to: "/adm/ukk", label: "UKK", perm: "ukk.read" },
  { to: "/adm/event", label: "Event", perm: "event.read" },
  { to: "/adm/master/siswa", label: "Siswa", perm: "master.siswa.read" },
  { to: "/adm/backup", label: "Backup", perm: "backup.create" },
  { to: "/adm/settings", label: "Settings", perm: "settings.read" },
];

export default function DashboardPage() {
  const { user, hasPermission } = useAuthStore();
  const navigate = useNavigate();
  const ringkasanQuery = useQuery({
    queryKey: ["dashboard-ringksan"],
    queryFn: () => apiClient.get<Ringkasan>("/api/v1/dashboard/ringkasan"),
  });
  const d = ringkasanQuery.data;
  const loading = ringkasanQuery.isLoading;

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>Dashboard</Typography.Title>
          <Typography.Text type="secondary">{`Selamat datang, ${user?.nama ?? "pengguna"} (${user?.role ?? "-"}).`}</Typography.Text>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <Card><Statistic title="Bank Soal" value={d?.bank_soal ?? "-"} loading={loading} prefix={<BookOpen size={24} />} /></Card>
        <Card><Statistic title="Jadwal Ujian" value={d?.jadwal_ujian ?? "-"} loading={loading} prefix={<CalendarDays size={24} />} /></Card>
        <Card><Statistic title="Peserta" value={d?.peserta ?? "-"} loading={loading} prefix={<Users size={24} />} /></Card>
        <Card><Statistic title="Rekap" value={d?.rekap ?? "-"} loading={loading} prefix={<FileText size={24} />} /></Card>
      </div>

      <Card title="Menu Cepat">
        <div className="flex flex-wrap gap-3">
          {MENU_CEPAT.filter((m) => !m.perm || hasPermission(m.perm)).map((m) => (
            <Typography.Link key={m.to} onClick={() => navigate(m.to)} className="inline-flex min-h-[44px] items-center font-semibold">
              {m.label}
            </Typography.Link>
          ))}
        </div>
      </Card>
    </div>
  );
}
