// BannerPeringatan — banner pelanggaran anti-curang (peringatan ke-1)
// dan layar diblokir (pelanggaran ke-2).
// REDESIGN antd (hanya presentasi): Alert antd. Teks & perilaku TIDAK diubah.
import { Alert, Button } from "antd";

interface Props {
  mode: "peringatan" | "diblokir";
  jumlahPelanggaran: number;
  onTutup?: () => void;
}

export default function BannerPeringatan({ mode, jumlahPelanggaran, onTutup }: Props) {
  if (mode === "diblokir") {
    return (
      <Alert
        type="error"
        showIcon
        message="Ujian Diblokir"
        description={`Anda diblokir karena pelanggaran berulang (${jumlahPelanggaran}x meninggalkan halaman ujian). Jawaban Anda tersimpan. Silakan hubungi proktor / pengawas untuk meminta unblock — ujian akan dilanjutkan dari sisa waktu Anda.`}
        className="p-6 text-center"
      />
    );
  }
  // Mode peringatan: SATU BARIS kompak (tanpa description) agar tingginya
  // selalu sama dengan slot yang dicadangkan (40px) — muncul/hilangnya
  // banner tidak menggeser soal. Teks dipotong elipsis bila sempit.
  return (
    <Alert
      type="warning"
      showIcon
      message={
        <span className="block truncate">
          Peringatan ({jumlahPelanggaran}×) — pelanggaran berikutnya memblokir ujian.
        </span>
      }
      action={
        onTutup ? (
          <Button size="small" onClick={onTutup}>
            Mengerti
          </Button>
        ) : undefined
      }
    />
  );
}
