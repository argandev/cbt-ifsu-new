// PengecohTable — tabel distribusi pilihan jawaban (pengecoh) per opsi.
// Baris pengecoh yang tidak berfungsi (dipilih < 5% peserta) ditandai
// dengan latar merah muda + badge "Tidak berfungsi" + keterangan di
// bawah tabel. Soal uraian/manual tidak punya opsi: pemanggil
// menampilkan "N/A" dan tidak memakai komponen ini (atau props kosong).
// Komponen ini hanya MENAMPILKAN data; persen dihitung backend.
import { Empty, Table, Tag } from "antd";

export type PengecohRow = {
  /** Label opsi, mis. "A", "B", "C", "D". */
  opsi: string;
  /** Jumlah peserta yang memilih opsi ini. */
  pemilih: number;
  /** Persen pemilih terhadap total penjawab (0–100). */
  persen: number;
  /** True bila persen >= 5 (pengecoh berfungsi). */
  berfungsi: boolean;
};

type PengecohTableProps = {
  pengecoh: PengecohRow[];
};

export function PengecohTable({ pengecoh }: PengecohTableProps) {
  if (pengecoh.length === 0) {
    return (
      <Empty description="Tidak ada data pengecoh. Soal uraian/manual tidak memiliki pengecoh (N/A)." />
    );
  }

  const tidakBerfungsi = pengecoh.filter((p) => !p.berfungsi);

  return (
    <div>
      <Table
        columns={[
          { title: "Opsi", dataIndex: "opsi", key: "opsi", render: (_: unknown, p: any) => (<>{p.opsi}</>) },
          { title: "Pemilih", dataIndex: "pemilih", key: "pemilih", align: "right" as const, render: (_: unknown, p: any) => (<>{p.pemilih}</>) },
          {
            title: "Persen",
            dataIndex: "persen",
            key: "persen",
            align: "right" as const,
            render: (_: unknown, p: any) => (<>{p.persen.toFixed(1)}%</>),
          },
          {
            title: "Status",
            key: "status",
            render: (_: unknown, p: any) => (
              <Tag color={p.berfungsi ? "green" : "red"}>
                {p.berfungsi ? "Berfungsi" : "Tidak berfungsi"}
              </Tag>
            ),
          },
        ]}
        dataSource={pengecoh.map((p, i) => ({ ...p, key: `${p.opsi}-${i}` }))}
        rowKey="key"
        pagination={false}
        locale={{ emptyText: "Tidak ada data pengecoh." }}
        scroll={{ x: "max-content" }}
        rowClassName={(p: any) => (p.berfungsi ? "" : "bg-danger-bg")}
      />
      {tidakBerfungsi.length > 0 && (
        <p className="mt-2 text-xs text-danger">
          {tidakBerfungsi.length} pengecoh tidak berfungsi (dipilih &lt; 5%
          peserta): opsi {tidakBerfungsi.map((p) => p.opsi).join(", ")}.
          Pertimbangkan merevisi atau mengganti opsi tersebut.
        </p>
      )}
    </div>
  );
}
