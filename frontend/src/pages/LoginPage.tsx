// LoginPage (Tim 1; komponen UI antd oleh Tim Setup + Tema).
// Bila server menjawab 403 PASSWORD_CHANGE_REQUIRED, arahkan ke /ganti-password.
// Design system: kartu dengan pita header biru tua (nama aplikasi tetap
// "CBT SMK" — bukan aplikasi ujian nasional resmi). Logika login sama:
// state, handleSubmit, navigasi, dan penanganan ApiError TIDAK diubah —
// yang diganti hanya komponen presentasi (Card/Input/Button/Alert custom
// -> antd). Halaman ini contoh acuan pemakaian antd untuk tim lain.
// Migrasi Tim Layout: form native -> antd Form + Form.Item dengan rules
// (memetakan `required` native yang sudah ada); label di-Indonesiakan.
import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Alert, Button, Card, Form, Input, Typography, theme } from "antd";
import { GraduationCap, LogIn } from "lucide-react";
import { useAuthStore } from "../store/auth";
import { ApiError } from "../lib/api";

export default function LoginPage() {
  const [login, setLogin] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const { login: doLogin, user } = useAuthStore();
  const navigate = useNavigate();
  const { token } = theme.useToken();

  const handleSubmit = async () => {
    setError("");
    setBusy(true);
    try {
      await doLogin(login, password);
      const role = useAuthStore.getState().user?.role;
      if (role === "siswa") {
        navigate("/");
      } else {
        navigate("/adm");
      }
    } catch (err) {
      if (err instanceof ApiError && err.code === "PASSWORD_CHANGE_REQUIRED") {
        navigate("/ganti-password", { state: { login } });
        return;
      }
      setError(err instanceof Error ? err.message : "Login gagal.");
    } finally {
      setBusy(false);
    }
  };

  // Sudah login -> jangan tampilkan form lagi.
  if (user) {
    return (
      <div className="mx-auto max-w-md py-10 text-center">
        <Typography.Text>
          Sudah login sebagai {user.nama}.{" "}
          <Link to={user.role === "siswa" ? "/" : "/adm"} className="font-medium text-primary underline">
            Lanjut ke beranda
          </Link>
        </Typography.Text>
      </div>
    );
  }

  return (
    <div className="mx-auto flex min-h-[70vh] max-w-md flex-col items-center justify-center gap-3 px-4">
      {/* styles.body padding 0 agar pita header biru menempel ke tepi kartu
          (seperti sebelumnya); isi kartu diberi padding sendiri di bawah. */}
      <Card className="w-full overflow-hidden" styles={{ body: { padding: 0 } }}>
        <div
          className="flex flex-col items-center gap-1 px-6 py-5 text-center"
          style={{ backgroundColor: token.colorPrimary, color: token.colorWhite }}
        >
          <span className="flex min-h-[48px] min-w-[48px] items-center justify-center rounded-lg bg-white/15" aria-hidden="true">
            <GraduationCap size={28} />
          </span>
          <p className="text-xl font-bold tracking-tight">CBT SMK</p>
          <p className="text-sm text-white/85">Ujian Berbasis Komputer — SMK</p>
        </div>
        <div className="px-6 py-5">
          <Typography.Title level={4} style={{ margin: 0 }}>Masuk CBT SMK v2</Typography.Title>
          <Typography.Text type="secondary">
            Pakai email atau nama pengguna + kata sandi. Siswa login dengan nama pengguna (NIS/NISN).
          </Typography.Text>
          <Form
            layout="vertical"
            initialValues={{ login: "", password: "" }}
            onFinish={handleSubmit}
            className="mt-4"
          >
            <Form.Item
              label="Email / Nama pengguna"
              name="login"
              rules={[{ required: true, message: "Email / Nama pengguna wajib diisi." }]}
            >
              <Input
                id="login"
                name="login"
                autoComplete="username"
                value={login}
                onChange={(e) => setLogin(e.target.value)}
                size="large"
                placeholder="Email / Nama pengguna"
              />
            </Form.Item>
            <Form.Item
              label="Kata sandi"
              name="password"
              rules={[{ required: true, message: "Kata sandi wajib diisi." }]}
            >
              <Input.Password
                id="password"
                name="password"
                type="password"
                autoComplete="current-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                size="large"
                placeholder="Kata sandi"
              />
            </Form.Item>
            {error ? (
              <Alert type="error" message={error} showIcon className="mb-4" />
            ) : null}
            <Button type="primary" htmlType="submit" loading={busy} block size="large" icon={<LogIn size={18} />}>
              {busy ? "Memproses..." : "Masuk"}
            </Button>
          </Form>
        </div>
      </Card>
      <p className="px-4 text-center text-xs leading-relaxed text-muted">
        Aplikasi ujian berbasis komputer milik sekolah. Bukan aplikasi ujian nasional resmi.
      </p>
    </div>
  );
}
