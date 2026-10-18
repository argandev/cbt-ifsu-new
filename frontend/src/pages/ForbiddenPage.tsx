// ForbiddenPage — halaman 403 (Tim Tambahan 3).
// Ditampilkan saat pengguna login tetapi tidak punya akses ke halaman tersebut.
// Migrasi Tim Layout: kartu rakitan -> antd Result (status 403). Teks tidak berubah.
import { Button, Result } from "antd";
import { Link } from "react-router-dom";

export default function ForbiddenPage() {
  return (
    <div className="mx-auto max-w-md px-4 py-16">
      <Result
        status="403"
        title="Anda tidak punya akses"
        subTitle="Akun Anda tidak memiliki izin untuk membuka halaman ini. Bila Anda merasa seharusnya bisa mengaksesnya, hubungi admin atau proktor."
        extra={
          <Link to="/"><Button type="primary">
            Kembali ke Beranda
          </Button></Link>
        }
      />
    </div>
  );
}
