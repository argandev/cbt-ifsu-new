// DayaBedaBadge — badge warna untuk interpretasi daya beda (D).
// Ambang (definisi final): >= 0,40 Sangat Baik; 0,30–0,39 Baik;
// 0,20–0,29 Cukup; 0,00–0,19 Buruk; < 0,00 Sangat Buruk.
// Status SELALU warna + teks (bukan warna saja) — aksesibilitas.
// Komponen ini hanya MENAMPILKAN label; D dihitung backend.
// Warna dari preset antd Tag (standar audit §2.2).
import { Tag } from "antd";

type DayaBedaBadgeProps = {
  /** Nilai daya beda D (biasanya -1 s.d. 1). */
  nilai: number;
};

function interpretasi(nilai: number): {
  label: string;
  color: "green" | "gold" | "orange" | "red";
} {
  if (nilai >= 0.4) {
    return { label: "Sangat Baik", color: "green" };
  }
  if (nilai >= 0.3) {
    return { label: "Baik", color: "green" };
  }
  if (nilai >= 0.2) {
    return { label: "Cukup", color: "gold" };
  }
  if (nilai >= 0) {
    return { label: "Buruk", color: "orange" };
  }
  return { label: "Sangat Buruk", color: "red" };
}

export function DayaBedaBadge({ nilai }: DayaBedaBadgeProps) {
  const { label, color } = interpretasi(nilai);
  return (
    <Tag color={color} title={`Daya beda ${nilai.toFixed(2)}: ${label}`}>
      <span>{nilai.toFixed(2)}</span> <span aria-label={`Interpretasi: ${label}`}>{label}</span>
    </Tag>
  );
}
