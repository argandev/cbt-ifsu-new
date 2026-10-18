// PenggunaPage — Manajemen Pengguna (akun staf).
// HANYA akun staf (Admin, Proktor, Guru, Pengawas, Penguji UKK) — siswa
// tetap dikelola di halaman Siswa (/adm/master/siswa) dan TIDAK tampil di sini.
//
// Kontrak API (tim backend):
//   GET  /api/v1/pengguna?q=&role=&is_active=
//     -> { data: [{ id, username, nama, email, role, is_active,
//          must_change_password, terakhir_masuk, created_at }] } (staf saja)
//   POST /api/v1/pengguna { username, nama, email, role, password } -> 201;
//     duplikat username -> 409 (pesan server ditampilkan apa adanya)
//   GET  /api/v1/pengguna/:id -> detail
//   PUT  /api/v1/pengguna/:id { nama, email, role } (username TIDAK bisa diubah)
//   POST /api/v1/pengguna/:id/reset-password { password_baru } -> 200
//     (akun wajib ganti password saat login berikutnya)
//   POST /api/v1/pengguna/:id/toggle-aktif { is_active } -> 200;
//     nonaktifkan diri sendiri / admin terakhir ditolak server
//   Error: { error: { code, message } } (Bahasa Indonesia) — message server
//     ditampilkan di Alert / Modal apa adanya.
//
// Aturan kata sandi: minimal 8 karakter — sama seperti GantiPasswordPage
// dan backend (auth.go: minPasswordLength = 8).
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "../../../lib/api";
import { formatTanggalJam } from "../../../lib/format";
import { statusTagColor } from "../../../components/common/statusTag";
import {
  Alert,
  Breadcrumb,
  Button,
  Form,
  Input,
  Modal,
  Select,
  Space,
  Table,
  Tag,
  Typography,
} from "antd";

type RoleStaf = "admin" | "proktor" | "guru" | "pengawas" | "penguji_ukk";

type Pengguna = {
  id: string;
  username: string;
  nama: string;
  email: string | null;
  role: string;
  is_active: boolean;
  must_change_password: boolean;
  terakhir_masuk: string | null;
  created_at: string;
};

type ListRes = { data: Pengguna[] };

const ROLE_LABEL: Record<string, string> = {
  admin: "Admin",
  proktor: "Proktor",
  guru: "Guru",
  pengawas: "Pengawas",
  penguji_ukk: "Penguji UKK",
};

const ROLE_OPTIONS = (["admin", "proktor", "guru", "pengawas", "penguji_ukk"] as RoleStaf[]).map(
  (role) => ({ value: role, label: ROLE_LABEL[role] }),
);

const MIN_PASSWORD = 8;

function labelRole(role: string): string {
  return ROLE_LABEL[role] ?? role;
}

function unwrapList(res: ListRes | Pengguna[]): Pengguna[] {
  if (Array.isArray(res)) return res;
  return res.data ?? [];
}

function apiErrorMessage(err: unknown): string {
  if (typeof err === "object" && err !== null) {
    const e = err as { message?: string; error?: { message?: string } };
    if (e.error?.message) return e.error.message;
    if (e.message) return e.message;
  }
  return "Terjadi kesalahan. Coba lagi.";
}

type FormBuat = {
  username: string;
  nama: string;
  email?: string;
  role: RoleStaf;
  password: string;
};

type FormUbah = {
  nama: string;
  email?: string;
  role: RoleStaf;
};

type FormReset = {
  password_baru: string;
};

export default function PenggunaPage() {
  const qc = useQueryClient();
  const [q, setQ] = useState("");
  const [search, setSearch] = useState("");
  const [filterRole, setFilterRole] = useState("");
  const [filterStatus, setFilterStatus] = useState("");
  const [pageError, setPageError] = useState("");

  const [modalBuatOpen, setModalBuatOpen] = useState(false);
  const [modalUbahOpen, setModalUbahOpen] = useState(false);
  const [modalResetOpen, setModalResetOpen] = useState(false);
  const [terpilih, setTerpilih] = useState<Pengguna | null>(null);
  const [formError, setFormError] = useState("");

  const [formBuat] = Form.useForm<FormBuat>();
  const [formUbah] = Form.useForm<FormUbah>();
  const [formReset] = Form.useForm<FormReset>();

  // Debounce pencarian 300ms (pola halaman admin lain).
  useEffect(() => {
    const t = setTimeout(() => setSearch(q), 300);
    return () => clearTimeout(t);
  }, [q]);

  const listQuery = useQuery({
    queryKey: ["pengguna", search, filterRole, filterStatus],
    queryFn: () => {
      const params = new URLSearchParams();
      if (search) params.set("q", search);
      if (filterRole) params.set("role", filterRole);
      if (filterStatus) params.set("is_active", filterStatus);
      const query = params.toString();
      return apiClient.get<ListRes | Pengguna[]>(
        `/api/v1/pengguna${query ? `?${query}` : ""}`,
      );
    },
  });
  const items = unwrapList(listQuery.data ?? []);

  const invalidate = () => qc.invalidateQueries({ queryKey: ["pengguna"] });

  const buatMutation = useMutation({
    mutationFn: (body: FormBuat) =>
      apiClient.post("/api/v1/pengguna", {
        username: body.username.trim(),
        nama: body.nama.trim(),
        email: body.email?.trim() || null,
        role: body.role,
        password: body.password,
      }),
    onSuccess: () => {
      setModalBuatOpen(false);
      setFormError("");
      formBuat.resetFields();
      invalidate();
    },
    onError: (err) => setFormError(apiErrorMessage(err)),
  });

  const ubahMutation = useMutation({
    mutationFn: (args: { id: string; body: FormUbah }) =>
      apiClient.put(`/api/v1/pengguna/${args.id}`, {
        nama: args.body.nama.trim(),
        email: args.body.email?.trim() || null,
        role: args.body.role,
      }),
    onSuccess: () => {
      setModalUbahOpen(false);
      setFormError("");
      invalidate();
    },
    onError: (err) => setFormError(apiErrorMessage(err)),
  });

  const resetMutation = useMutation({
    mutationFn: (args: { id: string; body: FormReset }) =>
      apiClient.post(`/api/v1/pengguna/${args.id}/reset-password`, {
        password_baru: args.body.password_baru,
      }),
    onSuccess: () => {
      setModalResetOpen(false);
      setFormError("");
      formReset.resetFields();
      invalidate();
    },
    onError: (err) => setFormError(apiErrorMessage(err)),
  });

  const toggleMutation = useMutation({
    mutationFn: (args: { id: string; is_active: boolean }) =>
      apiClient.post(`/api/v1/pengguna/${args.id}/toggle-aktif`, {
        is_active: args.is_active,
      }),
    onSuccess: () => {
      setPageError("");
      invalidate();
    },
    // Pesan server ditampilkan apa adanya (mis. penolakan nonaktifkan
    // diri sendiri / admin terakhir).
    onError: (err) => setPageError(apiErrorMessage(err)),
  });

  function openBuat() {
    setFormError("");
    formBuat.resetFields();
    setModalBuatOpen(true);
  }

  function openUbah(item: Pengguna) {
    setTerpilih(item);
    setFormError("");
    formUbah.setFieldsValue({
      nama: item.nama,
      email: item.email ?? undefined,
      role: item.role as RoleStaf,
    });
    setModalUbahOpen(true);
  }

  function openReset(item: Pengguna) {
    setTerpilih(item);
    setFormError("");
    formReset.resetFields();
    setModalResetOpen(true);
  }

  function konfirmasiToggle(item: Pengguna) {
    const aktifkan = !item.is_active;
    Modal.confirm({
      title: aktifkan ? "Aktifkan Akun?" : "Nonaktifkan Akun?",
      content: aktifkan ? (
        <p>
          Akun <strong>{item.username}</strong> ({item.nama}) akan diaktifkan kembali dan dapat
          login.
        </p>
      ) : (
        <p>
          Nonaktifkan akun <strong>{item.username}</strong> ({item.nama})? Akun tidak dapat login
          sampai diaktifkan kembali.
        </p>
      ),
      okText: aktifkan ? "Ya, Aktifkan" : "Ya, Nonaktifkan",
      cancelText: "Batal",
      okButtonProps: aktifkan ? undefined : { danger: true },
      onOk: () => {
        toggleMutation.mutate({ id: item.id, is_active: aktifkan });
      },
    });
  }

  return (
    <div className="space-y-4">
      <Breadcrumb
        items={[{ title: <Link to="/adm">Dashboard</Link> }, { title: "Pengguna" }]}
      />
      <div className="flex flex-col gap-3 border-b border-border pb-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Typography.Title level={3} style={{ margin: 0 }}>
            Pengguna
          </Typography.Title>
          <Typography.Text type="secondary">
            Kelola akun staf: Admin, Proktor, Guru, Pengawas, dan Penguji UKK. Akun siswa
            dikelola di halaman Siswa.
          </Typography.Text>
        </div>
        <Space wrap>
          <Button onClick={openBuat} type="primary">
            Buat Pengguna
          </Button>
        </Space>
      </div>

      <div className="flex flex-col gap-2 sm:flex-row">
        <Input.Search
          placeholder="Cari username / nama / email…"
          value={q}
          onChange={(e) => setQ(e.target.value)}
          aria-label="Cari pengguna"
          size="large"
          allowClear
        />
        <Select
          showSearch
          options={[
            { value: "", label: "Semua Role" },
            ...ROLE_OPTIONS,
          ]}
          value={filterRole}
          onChange={(v) => setFilterRole(v)}
          placeholder="Semua Role"
          filterOption={(input, opt) =>
            (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())
          }
          optionFilterProp="label"
          notFoundContent="Tidak ada data"
          style={{ width: "100%" }}
          size="large"
          aria-label="Filter role"
        />
        <Select
          options={[
            { value: "", label: "Semua Status" },
            { value: "true", label: "Aktif" },
            { value: "false", label: "Nonaktif" },
          ]}
          value={filterStatus}
          onChange={(v) => setFilterStatus(v)}
          placeholder="Semua Status"
          style={{ width: "100%" }}
          size="large"
          aria-label="Filter status"
        />
      </div>

      {pageError && <Alert type="error" showIcon message={pageError} />}
      {listQuery.isError && (
        <Alert
          type="error"
          showIcon
          message={`Gagal memuat pengguna: ${apiErrorMessage(listQuery.error)}`}
        />
      )}

      <Table
        columns={[
          {
            key: "username",
            title: "Username",
            dataIndex: "username",
            sorter: (a: Pengguna, b: Pengguna) => a.username.localeCompare(b.username, "id"),
          },
          {
            key: "nama",
            title: "Nama",
            dataIndex: "nama",
            sorter: (a: Pengguna, b: Pengguna) => a.nama.localeCompare(b.nama, "id"),
          },
          {
            key: "email",
            title: "Email",
            dataIndex: "email",
            render: (email: string | null) => email ?? "—",
          },
          {
            key: "role",
            title: "Role",
            dataIndex: "role",
            render: (role: string) => labelRole(role),
          },
          {
            key: "status",
            title: "Status",
            render: (_: unknown, item: Pengguna) => (
              <Tag color={statusTagColor(item.is_active ? "aktif" : "tidak aktif")}>
                {item.is_active ? "Aktif" : "Nonaktif"}
              </Tag>
            ),
          },
          {
            key: "terakhir_masuk",
            title: "Login Terakhir",
            dataIndex: "terakhir_masuk",
            render: (value: string | null) => formatTanggalJam(value),
          },
          {
            key: "must_change_password",
            title: "Wajib Ganti Password",
            render: (_: unknown, item: Pengguna) =>
              item.must_change_password ? (
                <Tag color="gold">Wajib Ganti Password</Tag>
              ) : (
                "—"
              ),
          },
          {
            key: "aksi",
            title: "Aksi",
            render: (_: unknown, item: Pengguna) => (
              <Space wrap>
                <Button onClick={() => openUbah(item)}>Ubah</Button>
                <Button onClick={() => openReset(item)}>Reset Password</Button>
                <Button
                  danger={!item.is_active ? false : true}
                  onClick={() => konfirmasiToggle(item)}
                  loading={toggleMutation.isPending}
                >
                  {item.is_active ? "Nonaktifkan" : "Aktifkan"}
                </Button>
              </Space>
            ),
          },
        ]}
        dataSource={items}
        rowKey="id"
        loading={listQuery.isLoading}
        pagination={false}
        locale={{ emptyText: "Belum ada pengguna." }}
        scroll={{ x: "max-content" }}
      />

      {/* Modal Buat Pengguna */}
      <Modal
        open={modalBuatOpen}
        onCancel={() => setModalBuatOpen(false)}
        title="Buat Pengguna"
        footer={null}
        destroyOnClose
        width={720}
      >
        <Form
          form={formBuat}
          layout="vertical"
          onFinish={(values) => buatMutation.mutate(values)}
        >
          {formError && (
            <Alert type="error" showIcon message={formError} style={{ marginBottom: 12 }} />
          )}
          <Form.Item
            label="Username"
            name="username"
            rules={[
              { required: true, message: "Username wajib diisi." },
              {
                pattern: /^[A-Za-z0-9._-]+$/,
                message: "Username hanya boleh berisi huruf, angka, titik, garis bawah, dan strip.",
              },
            ]}
          >
            <Input placeholder="mis. proktor1" size="large" autoComplete="username" />
          </Form.Item>
          <Form.Item
            label="Nama"
            name="nama"
            rules={[{ required: true, message: "Nama wajib diisi." }]}
          >
            <Input placeholder="Nama lengkap" size="large" />
          </Form.Item>
          <Form.Item
            label="Email"
            name="email"
            rules={[{ type: "email", message: "Format email tidak valid." }]}
          >
            <Input placeholder="nama@sekolah.id (opsional)" size="large" autoComplete="email" />
          </Form.Item>
          <Form.Item
            label="Role"
            name="role"
            rules={[{ required: true, message: "Role wajib dipilih." }]}
          >
            <Select
              showSearch
              options={ROLE_OPTIONS}
              placeholder="— Pilih Role —"
              filterOption={(input, opt) =>
                (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())
              }
              optionFilterProp="label"
              notFoundContent="Tidak ada data"
              size="large"
            />
          </Form.Item>
          <Form.Item
            label="Kata Sandi"
            name="password"
            rules={[
              { required: true, message: "Kata sandi wajib diisi." },
              {
                min: MIN_PASSWORD,
                message: `Kata sandi minimal ${MIN_PASSWORD} karakter.`,
              },
            ]}
            extra={`Minimal ${MIN_PASSWORD} karakter.`}
          >
            <Input.Password
              placeholder="Kata sandi awal"
              size="large"
              autoComplete="new-password"
            />
          </Form.Item>
          <div className="mt-4 flex flex-col gap-2 sm:flex-row sm:justify-end">
            <Button onClick={() => setModalBuatOpen(false)}>Batal</Button>
            <Button htmlType="submit" loading={buatMutation.isPending} type="primary">
              {buatMutation.isPending ? "Menyimpan…" : "Simpan"}
            </Button>
          </div>
        </Form>
      </Modal>

      {/* Modal Ubah Pengguna */}
      <Modal
        open={modalUbahOpen}
        onCancel={() => setModalUbahOpen(false)}
        title="Ubah Pengguna"
        footer={null}
        destroyOnClose
        width={720}
      >
        <Form
          form={formUbah}
          layout="vertical"
          onFinish={(values) => {
            if (terpilih) ubahMutation.mutate({ id: terpilih.id, body: values });
          }}
        >
          {formError && (
            <Alert type="error" showIcon message={formError} style={{ marginBottom: 12 }} />
          )}
          <Form.Item label="Username">
            <Input value={terpilih?.username ?? ""} disabled size="large" />
            <Typography.Text type="secondary" style={{ fontSize: 12 }}>
              Username tidak dapat diubah.
            </Typography.Text>
          </Form.Item>
          <Form.Item
            label="Nama"
            name="nama"
            rules={[{ required: true, message: "Nama wajib diisi." }]}
          >
            <Input placeholder="Nama lengkap" size="large" />
          </Form.Item>
          <Form.Item
            label="Email"
            name="email"
            rules={[{ type: "email", message: "Format email tidak valid." }]}
          >
            <Input placeholder="nama@sekolah.id (opsional)" size="large" autoComplete="email" />
          </Form.Item>
          <Form.Item
            label="Role"
            name="role"
            rules={[{ required: true, message: "Role wajib dipilih." }]}
          >
            <Select
              showSearch
              options={ROLE_OPTIONS}
              placeholder="— Pilih Role —"
              filterOption={(input, opt) =>
                (opt?.label ?? "").toLowerCase().includes(input.toLowerCase())
              }
              optionFilterProp="label"
              notFoundContent="Tidak ada data"
              size="large"
            />
          </Form.Item>
          <div className="mt-4 flex flex-col gap-2 sm:flex-row sm:justify-end">
            <Button onClick={() => setModalUbahOpen(false)}>Batal</Button>
            <Button htmlType="submit" loading={ubahMutation.isPending} type="primary">
              {ubahMutation.isPending ? "Menyimpan…" : "Simpan"}
            </Button>
          </div>
        </Form>
      </Modal>

      {/* Modal Reset Password */}
      <Modal
        open={modalResetOpen}
        onCancel={() => setModalResetOpen(false)}
        title="Reset Password"
        footer={null}
        destroyOnClose
        width={520}
      >
        <Form
          form={formReset}
          layout="vertical"
          onFinish={(values) => {
            if (terpilih) resetMutation.mutate({ id: terpilih.id, body: values });
          }}
        >
          {formError && (
            <Alert type="error" showIcon message={formError} style={{ marginBottom: 12 }} />
          )}
          <p>
            Reset kata sandi untuk akun <strong>{terpilih?.username}</strong> ({terpilih?.nama}).
          </p>
          <Form.Item
            label="Kata Sandi Baru"
            name="password_baru"
            rules={[
              { required: true, message: "Kata sandi baru wajib diisi." },
              {
                min: MIN_PASSWORD,
                message: `Kata sandi baru minimal ${MIN_PASSWORD} karakter.`,
              },
            ]}
            extra={`Minimal ${MIN_PASSWORD} karakter.`}
          >
            <Input.Password
              placeholder="Kata sandi baru"
              size="large"
              autoComplete="new-password"
            />
          </Form.Item>
          <Alert
            type="info"
            showIcon
            message="Pengguna wajib ganti password saat login berikutnya."
            style={{ marginBottom: 12 }}
          />
          <div className="mt-4 flex flex-col gap-2 sm:flex-row sm:justify-end">
            <Button onClick={() => setModalResetOpen(false)}>Batal</Button>
            <Button htmlType="submit" loading={resetMutation.isPending} type="primary">
              {resetMutation.isPending ? "Memproses…" : "Reset Password"}
            </Button>
          </div>
        </Form>
      </Modal>
    </div>
  );
}
