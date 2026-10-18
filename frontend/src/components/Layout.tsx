// Layout — kerangka halaman siswa (Tim Layout, redesign Ant Design).
// HANYA lapisan presentasi: logika logout, brand, dan redaksi footer
// TIDAK berubah dari versi sebelumnya.
// Design system: header BIRU TUA (token colorPrimary #1e4e9e) dengan
// brand putih, konten max 1100px (lebar ujian) + footer.
// Nama aplikasi tetap "CBT SMK" — bukan aplikasi ujian nasional resmi.
import { ReactNode } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Avatar, Button, Dropdown, Layout as AntdLayout, theme } from "antd";
import type { MenuProps } from "antd";
import {
  DownOutlined,
  LogoutOutlined,
  ReadOutlined,
  UserOutlined,
} from "@ant-design/icons";
import { useAuthStore } from "../store/auth";

const { Header, Content, Footer } = AntdLayout;

export default function Layout({ children }: { children: ReactNode }) {
  const { user, logout } = useAuthStore();
  const navigate = useNavigate();
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
        className="sr-only focus:not-sr-only focus:absolute focus:left-2 focus:top-2 focus:rounded focus:bg-white focus:px-3 focus:py-2 focus:text-primary"
      >
        Lewati ke konten
      </a>
      <Header style={{ backgroundColor: token.colorPrimary, padding: 0 }}>
        <div className="mx-auto flex h-16 max-w-[1100px] items-center justify-between gap-3 px-4">
          <Link to="/" className="flex items-center gap-2 text-lg font-bold text-white">
            <Avatar
              shape="square"
              icon={<ReadOutlined />}
              style={{ backgroundColor: "rgba(255, 255, 255, 0.15)", color: token.colorWhite }}
            />
            CBT SMK
            <span className="rounded-lg bg-white/20 px-1.5 py-0.5 text-xs font-bold">v2</span>
          </Link>
          <div className="flex items-center gap-2">
            {user ? (
              <span className="hidden text-sm text-white/85 sm:inline">
                {user.nama} ({user.role})
              </span>
            ) : null}
            {user ? (
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
            ) : null}
          </div>
        </div>
      </Header>
      <Content id="konten" className="mx-auto w-full max-w-[1100px] flex-1 px-4 py-6">
        {children}
      </Content>
      <Footer className="border-t border-border bg-white py-4 text-center text-xs text-muted">
        CBT SMK v2 — Ujian online SMK. Bukan aplikasi ujian nasional resmi.
      </Footer>
    </AntdLayout>
  );
}
