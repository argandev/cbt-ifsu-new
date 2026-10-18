// NotFoundPage — halaman 404 (Tim Tambahan 3).
// Dipakai tim integrasi untuk menggantikan komponen NotFound inline di App.tsx.
// Migrasi Tim Layout: markup rakitan -> antd Result (status 404). Teks tidak berubah.
import { Button, Result } from "antd";
import { Link } from "react-router-dom";

export default function NotFoundPage() {
  return (
    <div className="mx-auto max-w-md px-4 py-16">
      <Result
        status="404"
        title="Halaman tidak ditemukan"
        subTitle="Alamat yang Anda tuju tidak ada atau sudah dipindahkan. Periksa kembali alamatnya atau kembali ke beranda."
        extra={
          <Link to="/"><Button type="primary">
            Kembali ke Beranda
          </Button></Link>
        }
      />
    </div>
  );
}
