// AdminLayout — kerangka /adm/* (Tim Layout, redesign Ant Design).
// HANYA lapisan presentasi: route, permission, logika logout, dan semua
// label/to menu TIDAK berubah dari versi sebelumnya.
// Design system: header atas BIRU TUA (token colorPrimary #1e4e9e),
// sidebar putih + menu aktif biru tua teks putih.
// Sidebar tetap di desktop, Drawer antd (hamburger) di HP.
// Menu difilter permission (hanya sembunyikan — penegakan tetap di backend).
// Sidebar bisa di-scroll sendiri: brand shrink-0, nav dibungkus
// flex-1 overflow-y-auto — JANGAN hapus pembungkus scroll ini (perbaikan
// overflow lama: 19 menu × 44px melebihi tinggi layar laptop).
// Footer disclaimer "Bukan aplikasi ujian nasional resmi." dipertahankan.
import { useState } from "react";
import type { ReactNode } from "react";
import { Link, Outlet, useLocation, useNavigate } from "react-router-dom";
import {
  Avatar,
  Button,
  Drawer,
  Dropdown,
  Layout as AntdLayout,
  Menu,
  theme,
} from "antd";
import type { MenuProps } from "antd";
import {
  AimOutlined,
  AppstoreOutlined,
  BankOutlined,
  BarChartOutlined,
  BookOutlined,
  CalendarOutlined,
  CloudUploadOutlined,
  DashboardOutlined,
  DatabaseOutlined,
  DownOutlined,
  EditOutlined,
  HomeOutlined,
  IdcardOutlined,
  LogoutOutlined,
  MenuOutlined,
  MonitorOutlined,
  ReadOutlined,
  ScheduleOutlined,
  SettingOutlined,
  SolutionOutlined,
  TeamOutlined,
  TrophyOutlined,
  UserOutlined,
} from "@ant-design/icons";
import { useAuthStore } from "../store/auth";
import NotifikasiBell from "./common/NotifikasiBell";

const { Header, Sider, Content, Footer } = AntdLayout;

interface NavItem {
  to: string;
  label: string;
  icon: ReactNode;
  perm?: string;
  end?: boolean;
}

interface NavGroup {
  /** Judul grup (null = tanpa judul, untuk Dashboard). */
  judul: string | null;
  items: NavItem[];
}

// Sidebar dikelompokkan menjadi 5 grup + Dashboard (audit §4.3) — label,
// route, permission, dan jumlah item (20 menu, termasuk Pengguna) TIDAK
// berubah kecuali penambahan Pengguna, hanya dikelompokkan. SATU set ikon
// konsisten dari @ant-design/icons.
// Grup yang semua itemnya tersembunyi permission ikut tersembunyi.
const NAV_GROUPS: NavGroup[] = [
  {
    judul: null,
    items: [
      { to: "/adm", label: "Dashboard", icon: <DashboardOutlined />, end: true },
    ],
  },
  {
    judul: "Ujian",
    items: [
      { to: "/adm/bank-soal", label: "Bank Soal", icon: <DatabaseOutlined />, perm: "banksoal.read" },
      { to: "/adm/paket", label: "Paket Ujian", icon: <AppstoreOutlined />, perm: "paket.read" },
      { to: "/adm/jadwal", label: "Jadwal", icon: <CalendarOutlined />, perm: "jadwal.read" },
      { to: "/adm/monitor", label: "Monitor", icon: <MonitorOutlined />, perm: "ujian.monitor" },
    ],
  },
  {
    judul: "Hasil & Analisis",
    items: [
      { to: "/adm/rekap", label: "Rekap", icon: <BookOutlined />, perm: "rekap.read" },
      // D6: Koreksi Esai (UC10). Gate pada koreksi.update (hanya
      // admin/guru) — BUKAN koreksi.read: handler backend listKoreksi
      // menolak proktor, jadi menu JANGAN tampil untuk proktor
      // (UI ngikutin izin backend, bukan sebaliknya). Halaman
      // KoreksiPage sendiri juga meng-guard koreksi.update.
      { to: "/adm/koreksi", label: "Koreksi Esai", icon: <EditOutlined />, perm: "koreksi.update" },
      { to: "/adm/analisis", label: "Analisis Butir", icon: <BarChartOutlined />, perm: "analisis.read" },
      // Nilai Kompetensi memanggil GET /kompetensi/:id yang di-gate backend
      // dengan rekap.read (backend adalah sumber kebenaran) — perm
      // analisis.read sebelumnya salah.
      { to: "/adm/kompetensi", label: "Nilai Kompetensi", icon: <AimOutlined />, perm: "rekap.read" },
    ],
  },
  {
    judul: "Peserta & Pelaksanaan",
    items: [
      { to: "/adm/master/siswa", label: "Siswa", icon: <UserOutlined />, perm: "master.siswa.read" },
      // Kartu Peserta memanggil /api/v1/event + /ruangan + master (izin
      // event.read), bukan rekap.read — perm lama membuat menu tampil untuk
      // penguji_ukk (dan guru untuk event/ruangan) yang kemudian 403.
      { to: "/adm/kartu", label: "Kartu Peserta", icon: <IdcardOutlined />, perm: "event.read" },
      { to: "/adm/event", label: "Event", icon: <ScheduleOutlined />, perm: "event.read" },
      { to: "/adm/ruangan", label: "Ruangan", icon: <HomeOutlined />, perm: "event.read" },
      { to: "/adm/ukk", label: "UKK", icon: <TrophyOutlined />, perm: "ukk.read" },
    ],
  },
  {
    judul: "Data Induk",
    items: [
      { to: "/adm/master/jurusan", label: "Jurusan", icon: <BankOutlined />, perm: "master.jurusan.read" },
      { to: "/adm/master/kelas", label: "Kelas", icon: <TeamOutlined />, perm: "master.kelas.read" },
      { to: "/adm/master/mapel", label: "Mapel", icon: <ReadOutlined />, perm: "master.mapel.read" },
      { to: "/adm/master/guru", label: "Guru", icon: <SolutionOutlined />, perm: "master.guru.read" },
    ],
  },
  {
    judul: "Sistem",
    items: [
      { to: "/adm/backup", label: "Backup", icon: <CloudUploadOutlined />, perm: "backup.create" },
      { to: "/adm/settings", label: "Settings", icon: <SettingOutlined />, perm: "settings.read" },
      { to: "/adm/pengguna", label: "Pengguna", icon: <TeamOutlined />, perm: "pengguna.read" },
    ],
  },
];

/** Kunci Menu yang terpilih dari pathname: pencocokan prefix terpanjang
 *  (halaman turunan seperti /adm/bank-soal/baru tetap menyorot Bank Soal).
 *  Dashboard (end) hanya cocok persis. Monitor juga dipakai di luar /adm
 *  (route /monitor/:id dan /pengawas/monitor memakai AdminLayout). */
function kunciTerpilih(pathname: string): string {
  if (
    pathname === "/monitor" ||
    pathname.startsWith("/monitor/") ||
    pathname.startsWith("/pengawas/monitor")
  ) {
    return "/adm/monitor";
  }
  let terbaik = "";
  for (const grup of NAV_GROUPS) {
    for (const item of grup.items) {
      const cocok = item.end
        ? pathname === item.to
        : pathname === item.to || pathname.startsWith(`${item.to}/`);
      if (cocok && item.to.length > terbaik.length) terbaik = item.to;
    }
  }
  return terbaik;
}

export function AdminSidebar({ onNavigate }: { onNavigate?: () => void }) {
  const { hasPermission } = useAuthStore();
  const location = useLocation();
  const navigate = useNavigate();
  const terpilih = kunciTerpilih(location.pathname);

  const items: MenuProps["items"] = [];
  for (const grup of NAV_GROUPS) {
    const terlihat = grup.items.filter((item) => !item.perm || hasPermission(item.perm));
    if (terlihat.length === 0) continue;
    const anak = terlihat.map((item) => ({
      key: item.to,
      icon: item.icon,
      label: item.label,
    }));
    if (grup.judul === null) {
      items.push(...anak);
    } else {
      items.push({
        key: `grup-${grup.judul}`,
        type: "group",
        label: grup.judul,
        children: anak,
      });
    }
  }

  return (
    <Menu
      mode="inline"
      selectedKeys={terpilih ? [terpilih] : []}
      items={items}
      onClick={({ key }) => {
        navigate(key);
        onNavigate?.();
      }}
      style={{ borderInlineEnd: "none" }}
    />
  );
}

function Brand() {
  const { token } = theme.useToken();
  return (
    <Link to="/adm" className="flex items-center gap-2 text-lg font-bold text-ink">
      <Avatar shape="square" icon={<ReadOutlined />} style={{ backgroundColor: token.colorPrimary }} />
      <span>
        CBT SMK <span style={{ color: token.colorPrimary }}>v2</span>
        <span className="block text-xs font-medium leading-tight text-muted">Panel Sekolah</span>
      </span>
    </Link>
  );
}

export default function AdminLayout({ children }: { children?: ReactNode }) {
  const { user, logout } = useAuthStore();
  const navigate = useNavigate();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const { token } = theme.useToken();

  const handleLogout = async () => {
    await logout();
    navigate("/login");
  };

  const menuPengguna: MenuProps["items"] = [
    {
      key: "keluar",
      label: "Keluar",
      icon: <LogoutOutlined />,
      danger: true,
    },
  ];

  return (
    <AntdLayout style={{ minHeight: "100vh" }}>
      <a
        href="#konten"
        className="sr-only focus:not-sr-only focus:absolute focus:left-2 focus:top-2 focus:z-50 focus:rounded focus:bg-white focus:px-3 focus:py-2 focus:text-primary"
      >
        Lewati ke konten
      </a>
      {/* Sidebar desktop — nav bisa di-scroll sendiri: 20 menu × 44px
          melebihi tinggi layar laptop, sebelumnya kepotong tanpa scroll. */}
      <Sider
        width={256}
        theme="light"
        className="hidden lg:block"
        style={{ height: "100vh", position: "sticky", top: 0, overflow: "hidden" }}
      >
        <div className="flex h-full flex-col">
          <div className="flex h-16 shrink-0 items-center border-b border-border px-4">
            <Brand />
          </div>
          <div className="flex-1 overflow-y-auto">
            <AdminSidebar />
          </div>
        </div>
      </Sider>

      {/* Drawer HP (antd Drawer) */}
      <Drawer
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        placement="left"
        width={288}
        title={
          <span onClick={() => setDrawerOpen(false)}>
            <Brand />
          </span>
        }
        styles={{ body: { padding: 0 } }}
      >
        <div className="overflow-y-auto">
          <AdminSidebar onNavigate={() => setDrawerOpen(false)} />
        </div>
      </Drawer>

      <AntdLayout>
        <Header
          style={{
            backgroundColor: token.colorPrimary,
            padding: "0 16px",
            position: "sticky",
            top: 0,
            zIndex: 40,
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            gap: 12,
          }}
        >
          <div className="flex min-w-0 items-center gap-2">
            <Button
              type="text"
              aria-label="Buka menu"
              icon={<MenuOutlined />}
              onClick={() => setDrawerOpen(true)}
              className="lg:hidden"
              style={{ color: token.colorWhite }}
            />
            <span className="truncate text-sm font-medium text-white/90">
              {user ? `${user.nama} (${user.role})` : ""}
            </span>
          </div>
          <div className="flex shrink-0 items-center gap-1">
            <NotifikasiBell />
            <Dropdown
              menu={{
                items: menuPengguna,
                onClick: ({ key }) => {
                  if (key === "keluar") void handleLogout();
                },
              }}
              trigger={["click"]}
            >
              <Button type="text" aria-label="Menu pengguna" style={{ color: token.colorWhite }}>
                <Avatar size="small" icon={<UserOutlined />} />
                <DownOutlined />
              </Button>
            </Dropdown>
          </div>
        </Header>
        <Content id="konten" className="mx-auto w-full max-w-[1280px] flex-1 px-4 py-4 sm:px-6 sm:py-6">
          {children ?? <Outlet />}
        </Content>
        <Footer className="border-t border-border bg-white py-4 text-center text-xs text-muted">
          CBT SMK v2 — Ujian online SMK. Bukan aplikasi ujian nasional resmi.
        </Footer>
      </AntdLayout>
    </AntdLayout>
  );
}
