// App — routing (Tim 1).
// Peta rute lengkap ada di docs/KONTRAK.md §9. Halaman tim lain diimpor
// dari path persis kontrak; halaman yang belum ada diganti STUB oleh tim
// terkait tanpa mengubah path / nama export.
import { useEffect } from "react";
import { Navigate, Route, Routes, useLocation } from "react-router-dom";
import { ConfigProvider } from "antd";
import idID from "antd/locale/id_ID";
import { useAuthStore } from "./store/auth";
import Layout from "./components/Layout";
import AdminLayout from "./components/AdminLayout";
import LoginPage from "./pages/LoginPage";
import GantiPasswordPage from "./pages/GantiPasswordPage";
import SiswaHomePage from "./pages/siswa/SiswaHomePage";
import DashboardPage from "./pages/admin/DashboardPage";
import JurusanPage from "./pages/admin/master/JurusanPage";
import KelasPage from "./pages/admin/master/KelasPage";
import MapelPage from "./pages/admin/master/MapelPage";
import GuruPage from "./pages/admin/master/GuruPage";
import SiswaPage from "./pages/admin/master/SiswaPage";
import ImportSiswaPage from "./pages/admin/master/ImportSiswaPage";
import BankSoalListPage from "./pages/admin/banksoal/BankSoalListPage";
import BankSoalFormPage from "./pages/admin/banksoal/BankSoalFormPage";
import BankSoalDetailPage from "./pages/admin/banksoal/BankSoalDetailPage";
import PanduanImportPage from "./pages/admin/banksoal/PanduanImportPage";
import PaketPage from "./pages/admin/paket/PaketPage";
import JadwalPage from "./pages/admin/jadwal/JadwalPage";
import MonitorPage from "./pages/admin/monitor/MonitorPage";
import RekapPage from "./pages/admin/rekap/RekapPage";
import KoreksiPage from "./pages/admin/koreksi/KoreksiPage";
import BeritaAcaraPage from "./pages/admin/rekap/BeritaAcaraPage";
import AnalisisButirPage from "./pages/admin/analisis/AnalisisButirPage";
import KompetensiPage from "./pages/admin/analisis/KompetensiPage";
import SettingsPage from "./pages/admin/settings/SettingsPage";
import UkkListPage from "./pages/admin/ukk/UkkListPage";
import UkkDetailPage from "./pages/admin/ukk/UkkDetailPage";
import UkkNilaiPage from "./pages/admin/ukk/UkkNilaiPage";
import UkkSertifikatPage from "./pages/admin/ukk/UkkSertifikatPage";
import EventPage from "./pages/admin/event/EventPage";
import RuanganPage from "./pages/admin/event/RuanganPage";
import KartuPesertaPage from "./pages/admin/kartu/KartuPesertaPage";
import PenggunaPage from "./pages/admin/pengguna/PenggunaPage";
import BackupPage from "./pages/admin/backup/BackupPage";
import UjianPage from "./pages/siswa/UjianPage";
import HasilPage from "./pages/siswa/HasilPage";
import KartuSayaPage from "./pages/siswa/KartuSayaPage";
import NotFoundPage from "./pages/NotFoundPage";
import ForbiddenPage from "./pages/ForbiddenPage";

const STAFF_ROLES = ["admin", "proktor", "guru", "pengawas", "penguji_ukk"];

function RequireAuth({ children, staffOnly = false }: { children: React.ReactNode; staffOnly?: boolean }) {
  const { user, loading } = useAuthStore();
  const location = useLocation();
  if (loading) {
    return <p className="flex min-h-screen items-center justify-center bg-surface p-6 text-muted">Memuat...</p>;
  }
  if (!user) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }
  if (staffOnly && !STAFF_ROLES.includes(user.role)) {
    return <Navigate to="/403" replace />;
  }
  return <>{children}</>;
}

export default function App() {
  const fetchMe = useAuthStore((s) => s.fetchMe);
  useEffect(() => {
    void fetchMe();
  }, [fetchMe]);

  // ConfigProvider antd (Tim Setup + Tema): HANYA bungkus + theme.
  // Route, logika auth, dan semua element di bawah ini TIDAK diubah.
  // theme: colorPrimary #1e4e9e = biru primer yang sudah dipakai
  // (token --color-anbk / --color-primary di index.css) — dipertahankan.
  // locale id_ID: teks bawaan komponen antd (mis. "Pilih", pagination,
  // Upload, DatePicker) berbahasa Indonesia. Tanpa import CSS apa pun:
  // antd v6 memakai CSS-in-JS (gaya ikut bundle lokal, tanpa CDN).
  return (
    <ConfigProvider
      locale={idID}
      theme={{
        token: {
          // Token antd — satu-satunya tempat token didefinisikan
          // (AUDIT-KONSISTENSI.md §2.1). Nilai di sini adalah sumber
          // kebenaran; token Tailwind di index.css @theme hanya jembatan
          // migrasi dengan nilai yang SAMA — jangan buat palet baru.
          colorPrimary: "#1e4e9e",
          // Hover primer SATU nilai: #173d7a. Nilai ganda lama #163d7d
          // DIHAPUS (jangan dipakai lagi di theme maupun di @theme).
          colorPrimaryHover: "#173d7a",
          colorSuccess: "#15803d",
          colorWarning: "#b45309",
          colorError: "#dc2626",
          colorInfo: "#1e4e9e",
          colorText: "#0f172a",
          colorTextSecondary: "#475569",
          colorBorder: "#cbd5e1",
          colorBgLayout: "#eef2f7",
          borderRadius: 8,
          fontFamily: '"Plus Jakarta Sans", sans-serif',
          fontSize: 14,
          controlHeightLG: 44,
        },
      }}
    >
    <Routes>
      <Route path="/login" element={<Layout><LoginPage /></Layout>} />
      <Route path="/ganti-password" element={<Layout><GantiPasswordPage /></Layout>} />
      <Route path="/403" element={<Layout><ForbiddenPage /></Layout>} />

      <Route
        path="/"
        element={
          <RequireAuth>
            <Layout>
              <SiswaHomePage />
            </Layout>
          </RequireAuth>
        }
      />
      <Route
        path="/ujian/:attemptId"
        element={
          <RequireAuth>
            <Layout>
              <UjianPage />
            </Layout>
          </RequireAuth>
        }
      />
      <Route
        path="/hasil/:attemptId"
        element={
          <RequireAuth>
            <Layout>
              <HasilPage />
            </Layout>
          </RequireAuth>
        }
      />
      <Route
        path="/kartu"
        element={
          <RequireAuth>
            <Layout>
              <KartuSayaPage />
            </Layout>
          </RequireAuth>
        }
      />
      <Route
        path="/kartu/:jadwalId"
        element={
          <RequireAuth>
            <Layout>
              <Navigate to="/kartu" replace />
            </Layout>
          </RequireAuth>
        }
      />

      <Route
        path="/adm/*"
        element={
          <RequireAuth staffOnly>
            <AdminLayout />
          </RequireAuth>
        }
      >
        <Route index element={<DashboardPage />} />
        <Route path="master/jurusan" element={<JurusanPage />} />
        <Route path="master/kelas" element={<KelasPage />} />
        <Route path="master/mapel" element={<MapelPage />} />
        <Route path="master/guru" element={<GuruPage />} />
        <Route path="master/siswa" element={<SiswaPage />} />
        <Route path="master/siswa/import" element={<ImportSiswaPage />} />
        <Route path="bank-soal" element={<BankSoalListPage />} />
        <Route path="bank-soal/baru" element={<BankSoalFormPage />} />
        <Route path="bank-soal/:id/ubah" element={<BankSoalFormPage />} />
        <Route path="bank-soal/panduan-import" element={<PanduanImportPage />} />
        <Route path="bank-soal/:id" element={<BankSoalDetailPage />} />
        <Route path="paket" element={<PaketPage />} />
        <Route path="jadwal" element={<JadwalPage />} />
        <Route path="monitor" element={<MonitorPage />} />
        {/* Monitor per jadwal: JadwalPage menaut ke /monitor/:jadwalId dan
            MonitorPage membaca jadwalId dari pathname (/monitor/<uuid>).
            Tanpa route ini, tautan tersebut jatuh ke NotFound. */}
        <Route path="monitor/:jadwalId" element={<MonitorPage />} />
        <Route path="rekap" element={<RekapPage />} />
        <Route path="koreksi" element={<KoreksiPage />} />
        <Route path="rekap/:jadwalId/berita-acara" element={<BeritaAcaraPage />} />
        <Route path="analisis" element={<AnalisisButirPage />} />
        <Route path="kompetensi" element={<KompetensiPage />} />
        <Route path="event" element={<EventPage />} />
        <Route path="ruangan" element={<RuanganPage />} />
        <Route path="kartu" element={<KartuPesertaPage />} />
        <Route path="settings" element={<SettingsPage />} />
        <Route path="ukk" element={<UkkListPage />} />
        <Route path="ukk/:id" element={<UkkDetailPage />} />
        <Route path="ukk/:id/nilai" element={<UkkNilaiPage />} />
        <Route path="ukk/:id/sertifikat/:siswaId" element={<UkkSertifikatPage />} />
        <Route path="backup" element={<BackupPage />} />
        <Route path="pengguna" element={<PenggunaPage />} />
        <Route path="*" element={<NotFoundPage />} />
      </Route>

      <Route
        path="/pengawas/monitor"
        element={
          <RequireAuth staffOnly>
            <AdminLayout>
              <MonitorPage />
            </AdminLayout>
          </RequireAuth>
        }
      />
      {/* Tautan JadwalPage: href `/monitor/:jadwalId` (tanpa prefix /adm).
          Tanpa route ini, klik Monitor dari Jadwal jatuh ke NotFound. */}
      <Route
        path="/monitor/:jadwalId"
        element={
          <RequireAuth staffOnly>
            <AdminLayout>
              <MonitorPage />
            </AdminLayout>
          </RequireAuth>
        }
      />
      <Route
        path="/pengawas/monitor/:jadwalId"
        element={
          <RequireAuth staffOnly>
            <AdminLayout>
              <MonitorPage />
            </AdminLayout>
          </RequireAuth>
        }
      />

      <Route path="*" element={<Layout><NotFoundPage /></Layout>} />
    </Routes>
    </ConfigProvider>
  );
}
