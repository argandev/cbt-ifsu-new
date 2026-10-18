// Timer — hitung mundur dari sisa detik server.
// Props: sisaDetik awal (dari server), onExpire (dipanggil sekali saat 0).
// Menampilkan jam:menit:detik, monospace 20-24px, merah bila < 5 menit.
// REDESIGN antd (hanya presentasi): Tag antd untuk tampilan angka.
// Logika hitung mundur, sinkron server, dan onExpire TIDAK diubah.
import { useEffect, useRef, useState } from "react";
import { Tag, theme } from "antd";
import { Clock } from "lucide-react";

interface Props {
  sisaDetik: number;
  onExpire: () => void;
}

function format(detik: number): string {
  const d = Math.max(0, detik);
  const jam = Math.floor(d / 3600);
  const menit = Math.floor((d % 3600) / 60);
  const sisa = d % 60;
  const pad = (n: number) => String(n).padStart(2, "0");
  return jam > 0 ? `${pad(jam)}:${pad(menit)}:${pad(sisa)}` : `${pad(menit)}:${pad(sisa)}`;
}

export default function Timer({ sisaDetik, onExpire }: Props) {
  const { token } = theme.useToken();
  const [sisa, setSisa] = useState(sisaDetik);
  const sudahExpire = useRef(false);

  // Sinkron bila server mengirim sisa baru (polling /status).
  useEffect(() => {
    setSisa(sisaDetik);
    if (sisaDetik > 0) sudahExpire.current = false;
  }, [sisaDetik]);

  useEffect(() => {
    const id = setInterval(() => {
      setSisa((prev) => {
        const next = prev - 1;
        if (next <= 0) {
          clearInterval(id);
          if (!sudahExpire.current) {
            sudahExpire.current = true;
            onExpire();
          }
          return 0;
        }
        return next;
      });
    }, 1000);
    return () => clearInterval(id);
  }, [onExpire]);

  const bahaya = sisa <= 300; // < 5 menit
  return (
    <div role="timer" aria-label={`Sisa waktu ${format(sisa)}`}>
      <Tag
        color={bahaya ? "error" : "blue"}
        style={{
          fontFamily: token.fontFamilyCode,
          fontSize: 24,
          fontWeight: 700,
          padding: "4px 12px",
          margin: 0,
        }}
      >
        {/* Lebar angka dicadangkan (min-width 8ch + tabular-nums) supaya
            Tag tidak bergeser saat angka berubah / format MM:SS ↔ H:MM:SS. */}
        <span className="inline-flex items-center gap-2 tabular-nums">
          <Clock size={20} aria-hidden />
          <span style={{ display: "inline-block", minWidth: "8ch", textAlign: "center" }}>
            {format(sisa)}
          </span>
        </span>
      </Tag>
    </div>
  );
}
