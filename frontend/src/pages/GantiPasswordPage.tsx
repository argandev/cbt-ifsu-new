// GantiPasswordPage (Tim 1).
// Dipakai saat login pertama (403 PASSWORD_CHANGE_REQUIRED) maupun dari
// dalam sesi login. Setelah berhasil, SEMUA session dihapus server —
// pengguna wajib login ulang dengan kata sandi baru.
// Design system: kartu dengan pita header biru tua. Logika sama.
// Migrasi Tim Layout: form native -> antd Form + Form.Item dengan rules
// (memetakan `required` native yang sudah ada; cek manual panjang /
// kecocokan / perbedaan kata sandi tetap di handler); label di-Indonesiakan.
// Nama field API (login, password_lama, password_baru) TIDAK berubah.
import { useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { KeyRound } from "lucide-react";
import apiClient, { ApiError } from "../lib/api";
import { Alert, Button, Card, Form, Input, Typography, theme } from "antd";

export default function GantiPasswordPage() {
  const location = useLocation();
  const navigate = useNavigate();
  const loginAwal = (location.state as { login?: string } | null)?.login ?? "";
  const [login, setLogin] = useState(loginAwal);
  const [passwordLama, setPasswordLama] = useState("");
  const [passwordBaru, setPasswordBaru] = useState("");
  const [konfirmasi, setKonfirmasi] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const { token } = theme.useToken();

  const handleSubmit = async () => {
    setError("");
    if (passwordBaru.length < 8) {
      setError("Kata sandi baru minimal 8 karakter.");
      return;
    }
    if (passwordBaru !== konfirmasi) {
      setError("Konfirmasi kata sandi baru tidak sama.");
      return;
    }
    if (passwordBaru === passwordLama) {
      setError("Kata sandi baru harus berbeda dari kata sandi lama.");
      return;
    }
    setBusy(true);
    try {
      await apiClient.post("/api/v1/auth/ganti-password", {
        login: login || undefined,
        password_lama: passwordLama,
        password_baru: passwordBaru,
      });
      navigate("/login", { state: { pesan: "Kata sandi berhasil diganti. Silakan login dengan kata sandi baru." } });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Ganti kata sandi gagal.");
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="mx-auto flex min-h-[70vh] max-w-md items-center px-4">
      <Card className="w-full overflow-hidden" styles={{ body: { padding: 0 } }}>
        <div
          className="flex flex-col items-center gap-1 px-6 py-5 text-center"
          style={{ backgroundColor: token.colorPrimary, color: token.colorWhite }}
        >
          <span className="flex min-h-[48px] min-w-[48px] items-center justify-center rounded-lg bg-white/15" aria-hidden="true">
            <KeyRound size={28} />
          </span>
          <p className="text-xl font-bold tracking-tight">CBT SMK</p>
          <p className="text-sm text-white/85">Ujian Berbasis Komputer — SMK</p>
        </div>
        <div className="p-5">
          <Typography.Title level={4} style={{ margin: 0 }}>Ganti Kata Sandi</Typography.Title>
          <Typography.Text type="secondary">
            {loginAwal
              ? "Kata sandi awal harus diganti sebelum login."
              : "Masukkan kata sandi lama dan kata sandi baru (minimal 8 karakter)."}
          </Typography.Text>
          <Form
            layout="vertical"
            initialValues={{ login: loginAwal }}
            onFinish={handleSubmit}
            className="mt-4"
          >
            <Form.Item label="Email / Nama pengguna (kosongkan bila sudah login)" name="login">
              <Input
                id="login"
                name="login"
                autoComplete="username"
                value={login}
                onChange={(e) => setLogin(e.target.value)}
                size="large"
              />
            </Form.Item>
            <Form.Item
              label="Kata Sandi Lama"
              name="password_lama"
              rules={[{ required: true, message: "Kata sandi lama wajib diisi." }]}
            >
              <Input.Password
                id="password_lama"
                name="password_lama"
                autoComplete="current-password"
                value={passwordLama}
                onChange={(e) => setPasswordLama(e.target.value)}
                size="large"
              />
            </Form.Item>
            <Form.Item
              label="Kata Sandi Baru"
              name="password_baru"
              rules={[{ required: true, message: "Kata sandi baru wajib diisi." }]}
            >
              <Input.Password
                id="password_baru"
                name="password_baru"
                autoComplete="new-password"
                value={passwordBaru}
                onChange={(e) => setPasswordBaru(e.target.value)}
                size="large"
              />
            </Form.Item>
            <Form.Item
              label="Konfirmasi Kata Sandi Baru"
              name="konfirmasi"
              rules={[{ required: true, message: "Konfirmasi kata sandi baru wajib diisi." }]}
            >
              <Input.Password
                id="konfirmasi"
                name="konfirmasi"
                autoComplete="new-password"
                value={konfirmasi}
                onChange={(e) => setKonfirmasi(e.target.value)}
                size="large"
              />
            </Form.Item>
            {error ? (
              <Alert type="error" message={error} showIcon className="mb-4" />
            ) : null}
            <Button htmlType="submit" loading={busy} type="primary" size="large" icon={<KeyRound size={18} />} block>
              {busy ? "Memproses..." : "Ganti Kata Sandi"}
            </Button>
            <Link to="/login" className="mt-3 block text-center text-sm font-medium text-primary underline">
              Kembali ke login
            </Link>
          </Form>
        </div>
      </Card>
    </div>
  );
}
