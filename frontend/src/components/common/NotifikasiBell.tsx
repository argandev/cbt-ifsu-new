// NotifikasiBell — bel notifikasi di header admin (Fase 3).
// GET /api/v1/notifikasi (polling 60 detik), badge jumlah belum dibaca,
// dropdown daftar + tandai dibaca (POST /api/v1/notifikasi/:id/baca).
// Backend notifikasi BELUM ada saat komponen ini dibuat (Tim Backend
// Hardening): bila endpoint menjawab 404 / error apa pun, komponen
// menampilkan "0" dan daftar kosong — TIDAK PERNAH crash.
//
// CHROME REDESIGN ANT DESIGN (Tim Form): badge jumlah → antd Badge,
// dropdown manual (klik-di-luar + Escape rakitan sendiri) → antd Popover
// (menangani buka/tutup + klik-di-luar), tombol → antd Button.
// Query polling, mutasi tandai-baca, dan normalisasi respons TETAP sama.
import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Badge, Button, Popover } from "antd";
import { BellOutlined, CheckOutlined } from "@ant-design/icons";
import apiClient from "../../lib/api";
import { formatTanggalJam } from "../../lib/format";

interface Notifikasi {
  id: string;
  judul?: string;
  pesan?: string;
  isi?: string;
  dibaca?: boolean;
  sudah_dibaca?: boolean;
  created_at?: string;
}

type NotifikasiResponse =
  | { items?: Notifikasi[]; total_belum_dibaca?: number }
  | Notifikasi[];

function normalisasi(data: NotifikasiResponse | undefined): { items: Notifikasi[]; belumDibaca: number } {
  if (!data) return { items: [], belumDibaca: 0 };
  const items = Array.isArray(data) ? data : (data.items ?? []);
  const dariServer = !Array.isArray(data) ? data.total_belum_dibaca : undefined;
  const belumDibaca =
    dariServer !== undefined ? dariServer : items.filter((n) => !(n.dibaca ?? n.sudah_dibaca ?? false)).length;
  return { items, belumDibaca };
}

export default function NotifikasiBell() {
  const queryClient = useQueryClient();
  const [terbuka, setTerbuka] = useState(false);

  const query = useQuery({
    queryKey: ["notifikasi"],
    queryFn: () => apiClient.get<NotifikasiResponse>("/api/v1/notifikasi"),
    refetchInterval: 60_000,
    retry: false,
    // 404 (endpoint belum ada) bukan error fatal: tampilkan 0 dengan aman.
    throwOnError: false,
  });

  const bacaMutation = useMutation({
    mutationFn: (id: string) => apiClient.post(`/api/v1/notifikasi/${id}/baca`, {}),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["notifikasi"] }),
    // Gagal tandai-baca (mis. endpoint belum ada) diabaikan diam-diam.
    onError: () => undefined,
  });

  const { items, belumDibaca } = normalisasi(query.data);

  return (
    <Popover
      open={terbuka}
      onOpenChange={setTerbuka}
      trigger="click"
      placement="bottomRight"
      title="Notifikasi"
      content={
        <div style={{ width: 320, maxWidth: "calc(100vw - 2rem)" }}>
          <p className="mb-2 text-xs text-muted">
            {belumDibaca > 0 ? `${belumDibaca} belum dibaca` : "Tidak ada notifikasi baru"}
          </p>
          <ul className="max-h-80 overflow-y-auto">
            {query.isError ? (
              <li className="px-1 py-6 text-center text-sm text-muted">
                Notifikasi belum tersedia.
              </li>
            ) : items.length === 0 ? (
              <li className="px-1 py-6 text-center text-sm text-muted">Belum ada notifikasi.</li>
            ) : (
              items.map((n) => {
                const sudahDibaca = n.dibaca ?? n.sudah_dibaca ?? false;
                return (
                  <li key={n.id} className="flex items-start justify-between gap-2 border-b border-border px-1 py-3 last:border-0">
                    <div className="min-w-0">
                      <p className="text-sm font-medium text-ink">{n.judul ?? "Notifikasi"}</p>
                      {n.pesan ?? n.isi ? (
                        <p className="mt-0.5 text-sm text-muted">{n.pesan ?? n.isi}</p>
                      ) : null}
                      {n.created_at ? (
                        <p className="mt-0.5 text-xs text-muted">
                          {formatTanggalJam(n.created_at)}
                        </p>
                      ) : null}
                    </div>
                    {!sudahDibaca ? (
                      <Button
                        type="text"
                        aria-label="Tandai dibaca"
                        title="Tandai dibaca"
                        loading={bacaMutation.isPending}
                        onClick={() => bacaMutation.mutate(n.id)}
                        icon={<CheckOutlined />}
                        className="flex min-h-[44px] min-w-[44px] shrink-0 items-center justify-center"
                      />
                    ) : null}
                  </li>
                );
              })
            )}
          </ul>
        </div>
      }
    >
      <Badge count={belumDibaca} overflowCount={99} showZero={false}>
        <Button
          type="text"
          aria-label={belumDibaca > 0 ? `Notifikasi: ${belumDibaca} belum dibaca` : "Notifikasi"}
          aria-expanded={terbuka}
          icon={<BellOutlined style={{ fontSize: 20 }} />}
          className="flex min-h-[44px] min-w-[44px] items-center justify-center !text-white"
        />
      </Badge>
    </Popover>
  );
}
