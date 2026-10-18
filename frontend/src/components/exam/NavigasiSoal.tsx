// NavigasiSoal — grid nomor soal + status dijawab / ragu / belum.
// Target sentuh min 44px. Kotak = antd Button (§2.9 AUDIT-KONSISTENSI):
// belum = type default (putih), sudah = token success (hijau),
// ragu-ragu = token warning (kuning), aktif = type primary + aria-current
// (primer mengalahkan warna status; ragu-ragu + aktif tetap tampil Flag).
// Warna dari theme.useToken() — BUKAN hex mentah.
import { Button, theme } from "antd";
import { Flag } from "lucide-react";

export interface StatusSoal {
  soal_id: string;
  nomor: number;
  dijawab: boolean;
  ragu: boolean;
}

interface Props {
  items: StatusSoal[];
  aktif: number; // index soal aktif (0-based)
  onPilih: (index: number) => void;
}

export default function NavigasiSoal({ items, aktif, onPilih }: Props) {
  const { token } = theme.useToken();
  return (
    <nav aria-label="Navigasi soal">
      {/* Grid nomor: gap basis 8px (P-07). Button 44x44 dipertahankan. */}
      <div className="grid grid-cols-4 gap-2 sm:grid-cols-5">
        {items.map((it, i) => {
          const isAktif = i === aktif;
          // Aktif selalu primary; status non-aktif dipetakan ke warna token.
          const gayaStatus: React.CSSProperties | undefined = isAktif
            ? undefined
            : it.ragu
              ? {
                  backgroundColor: token.colorWarningBg,
                  borderColor: token.colorWarning,
                  color: token.colorWarning,
                }
              : it.dijawab
                ? {
                    backgroundColor: token.colorSuccess,
                    borderColor: token.colorSuccess,
                    color: token.colorWhite,
                  }
                : undefined;
          return (
            <Button
              key={it.soal_id}
              type={isAktif ? "primary" : "default"}
              onClick={() => onPilih(i)}
              aria-label={`Soal ${it.nomor}${it.ragu ? " (ragu-ragu)" : ""}${it.dijawab ? " (dijawab)" : ""}`}
              aria-current={isAktif ? "true" : undefined}
              style={{ width: "100%", minWidth: 44, minHeight: 44, padding: "0 4px", ...gayaStatus }}
            >
              {it.nomor}
              {it.ragu && <Flag size={12} aria-hidden />}
            </Button>
          );
        })}
      </div>
      {/* Legenda memakai token yang sama dengan kotak, supaya warnanya identik. */}
      <div className="mt-3 flex flex-wrap gap-3 text-xs text-muted">
        <span className="inline-flex items-center gap-1">
          <span
            className="inline-block h-3 w-3 rounded-sm"
            style={{ backgroundColor: token.colorSuccess }}
          />{" "}
          Dijawab
        </span>
        <span className="inline-flex items-center gap-1">
          <span
            className="inline-block h-3 w-3 rounded-sm"
            style={{
              backgroundColor: token.colorWarningBg,
              boxShadow: `inset 0 0 0 1px ${token.colorWarning}`,
            }}
          />{" "}
          Ragu-ragu
        </span>
        <span className="inline-flex items-center gap-1">
          <span
            className="inline-block h-3 w-3 rounded-sm"
            style={{
              backgroundColor: token.colorBgContainer,
              boxShadow: `inset 0 0 0 1px ${token.colorBorder}`,
            }}
          />{" "}
          Belum
        </span>
      </div>
    </nav>
  );
}
