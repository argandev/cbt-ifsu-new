// KompetensiHeatmap — heatmap siswa x kompetensi.
// Tiap sel menampilkan nilai per kompetensi (0–100) dengan warna sesuai
// nilai; sel null (siswa tidak menjawab satu pun soal kompetensi itu)
// berwarna abu-abu dengan tanda "—" + flag "tidak ada soal" dari backend.
// Komponen ini hanya MENAMPILKAN data; nilai dihitung backend.
// Warna sel dari preset antd Tag (standar audit §2.2): >= 90 hijau,
// 75–89 hijau, 60–74 kuning, 40–59 oranye, < 40 merah, null abu-abu.
import { Empty, Table, Tag } from "antd";

export type KompetensiSiswaRow = {
  nama: string;
  /** Nilai per kompetensi; null = tidak ada soal yang dijawab. */
  nilai: Record<string, number | null>;
};

type KompetensiHeatmapProps = {
  siswa: KompetensiSiswaRow[];
};

function warnaTag(nilai: number | null): "green" | "gold" | "orange" | "red" | "default" {
  if (nilai === null) {
    return "default";
  }
  if (nilai >= 75) return "green";
  if (nilai >= 60) return "gold";
  if (nilai >= 40) return "orange";
  return "red";
}

export function KompetensiHeatmap({ siswa }: KompetensiHeatmapProps) {
  if (siswa.length === 0) {
    return <Empty description="Tidak ada data nilai per kompetensi." />;
  }

  // Kolom = gabungan kunci kompetensi, urutan kemunculan pertama.
  const kompetensi: string[] = [];
  for (const row of siswa) {
    for (const k of Object.keys(row.nilai)) {
      if (!kompetensi.includes(k)) {
        kompetensi.push(k);
      }
    }
  }

  return (
    <div>
      <Table
        columns={[
          { title: "Siswa", dataIndex: "nama", key: "nama", render: (_: unknown, row: any) => (<>{row.nama}</>) },
          ...kompetensi.map((k) => ({
            title: k,
            key: k,
            align: "center" as const,
            render: (_: unknown, row: any) => {
              const v: number | null = row.nilai[k] ?? null;
              return (
                <Tag
                  color={warnaTag(v)}
                  title={
                    v === null
                      ? `${row.nama} — ${k}: tidak ada soal`
                      : `${row.nama} — ${k}: ${v.toFixed(1)}`
                  }
                >
                  {v === null ? "—" : v.toFixed(0)}
                </Tag>
              );
            },
          })),
        ]}
        dataSource={siswa.map((row, i) => ({ ...row, key: `${row.nama}-${i}` }))}
        rowKey="key"
        pagination={false}
        locale={{ emptyText: "Tidak ada data nilai per kompetensi." }}
        scroll={{ x: "max-content" }}
      />
      <p className="mt-2 text-xs text-muted">
        Skala warna: 90+ hijau tua, 75–89 hijau, 60–74 kuning, 40–59
        oranye, &lt; 40 merah, abu-abu (—) = tidak ada soal yang dijawab.
      </p>
    </div>
  );
}
