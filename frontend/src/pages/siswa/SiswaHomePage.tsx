// SiswaHomePage — beranda siswa (Tim 1, diperbaiki Fase 3 — temuan C.1).
// Menampilkan jadwal ujian HARI INI siswa (GET /api/v1/ujian/aktif).
// Field dipakai PERSIS seperti respons backend: jadwal_id, paket_id,
// tanggal, jam_mulai, jam_selesai, status, paket_kode, paket_nama,
// paket_jenis, durasi_menit, kelas_nama, attempt_id, attempt_status.
// JANGAN memakai item.id / nama_ujian / mapel / mulai — field itu tidak
// ada di respons (versi lama membuat link /ujian/undefined).
// Alur masuk: tanpa attempt -> input token (kecuali Tryout) lalu
// POST /api/v1/ujian/mulai { jadwal_id, token } -> navigasi ke
// /ujian/:attemptId dari respons. Attempt berjalan -> "Lanjut";
// attempt selesai -> "Lihat Hasil" (/hasil/:attemptId);
// attempt diblokir -> pesan hubungi proktor (tanpa tombol masuk).
//
// REDESIGN antd (hanya presentasi): Card / Button / Tag / Input / Alert /
// Empty antd. Logika, endpoint, payload, dan teks TIDAK diubah.
import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link, useNavigate } from "react-router-dom";
import { Alert, Button, Card, Empty, Input, Tag, Typography } from "antd";
import { CalendarDays, IdCard, Play } from "lucide-react";
import apiClient from "../../lib/api";
import { formatJam, formatTanggal } from "../../lib/format";

/** Satu item GET /api/v1/ujian/aktif (lihat backend/internal/exam/exam.go ujianAktif). */
interface UjianAktif {
  jadwal_id: string;
  paket_id: string;
  tanggal: string;
  jam_mulai: string;
  jam_selesai: string;
  status: string;
  paket_kode: string;
  paket_nama: string;
  paket_jenis: string;
  durasi_menit: number;
  kelas_nama: string;
  attempt_id: string | null;
  attempt_status: string | null;
  /** Bila backend mengirimnya: hasil tampil ke siswa (paket.tampil_hasil AND settings). */
  tampil_hasil?: boolean;
}

interface MulaiResponse {
  attempt_id: string;
  jadwal_id: string;
  status: string;
}

export default function SiswaHomePage() {
  const navigate = useNavigate();
  const [tokenPerJadwal, setTokenPerJadwal] = useState<Record<string, string>>({});
  const [errorPerJadwal, setErrorPerJadwal] = useState<Record<string, string>>({});
  const [memulaiJadwal, setMemulaiJadwal] = useState<string | null>(null);

  const query = useQuery({
    queryKey: ["ujian-aktif"],
    queryFn: () => apiClient.get<{ items?: UjianAktif[] } | UjianAktif[]>("/api/v1/ujian/aktif"),
    retry: false,
  });

  const items: UjianAktif[] = Array.isArray(query.data) ? query.data : (query.data?.items ?? []);

  const handleMasuk = async (item: UjianAktif) => {
    // Attempt berjalan yang sudah ada: langsung lanjut, tanpa POST mulai lagi.
    if (item.attempt_id && item.attempt_status === "berjalan") {
      navigate(`/ujian/${item.attempt_id}`);
      return;
    }
    setErrorPerJadwal((prev) => ({ ...prev, [item.jadwal_id]: "" }));
    setMemulaiJadwal(item.jadwal_id);
    try {
      const res = await apiClient.post<MulaiResponse>("/api/v1/ujian/mulai", {
        jadwal_id: item.jadwal_id,
        token: item.paket_jenis === "Tryout" ? "" : (tokenPerJadwal[item.jadwal_id] ?? ""),
      });
      if (!res.attempt_id) throw new Error("Server tidak mengembalikan attempt_id.");
      navigate(`/ujian/${res.attempt_id}`);
    } catch (err) {
      setErrorPerJadwal((prev) => ({
        ...prev,
        [item.jadwal_id]: err instanceof Error ? err.message : "Gagal memulai ujian.",
      }));
    } finally {
      setMemulaiJadwal(null);
    }
  };

  return (
    <div className="flex flex-col gap-6">
      <div className="rounded-lg bg-anbk px-5 py-4 text-white">
        <p className="text-lg font-bold tracking-wide">CBT SMK</p>
        <p className="text-sm text-white">Ujian Berbasis Komputer</p>
      </div>
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <Typography.Title level={3} style={{ margin: 0 }}>Ujian Hari Ini</Typography.Title>
          <p className="mt-1 text-sm text-muted">
            Jadwal ujian hari ini untuk kelas Anda. Token ujian diberikan proktor di ruangan.
          </p>
        </div>
        <Link to="/kartu">
          <Button icon={<IdCard size={16} />}>Kartu Saya</Button>
        </Link>
      </div>
      {query.isPending ? (
        <p className="text-muted">Memuat jadwal ujian...</p>
      ) : query.isError ? (
        <Card>
          <Alert
            type="error"
            message={`Jadwal ujian belum bisa dimuat (${(query.error as Error).message}). Coba muat ulang nanti.`}
            showIcon
          />
          <Button className="mt-3" onClick={() => void query.refetch()}>
            Muat Ulang
          </Button>
        </Card>
      ) : items.length === 0 ? (
        <Card>
          <Empty description="Tidak ada ujian hari ini." />
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          {items.map((item) => {
            const attemptBerjalan = item.attempt_status === "berjalan" && !!item.attempt_id;
            const attemptSelesai = item.attempt_status === "selesai" && !!item.attempt_id;
            const attemptDiblokir = item.attempt_status === "diblokir";
            const butuhToken = item.paket_jenis !== "Tryout" && !attemptBerjalan && !attemptSelesai;
            const error = errorPerJadwal[item.jadwal_id] ?? "";
            return (
              <Card key={item.jadwal_id} title={item.paket_nama} extra={<Tag color="blue">{item.paket_jenis}</Tag>}>
                <div className="flex flex-col gap-3">
                  <div className="flex items-center gap-2 text-sm text-muted">
                    <CalendarDays size={16} />
                    {item.paket_kode} · {item.paket_jenis} · {item.kelas_nama}
                  </div>
                  <p className="text-sm text-muted">
                    {formatTanggal(item.tanggal)} · {formatJam(item.jam_mulai)}–{formatJam(item.jam_selesai)} WIB · {item.durasi_menit} menit
                  </p>
                  <div className="flex flex-wrap gap-2">
                    <Tag color="blue">{item.status}</Tag>
                    {item.attempt_status ? <Tag color="blue">Attempt: {item.attempt_status}</Tag> : null}
                  </div>
                  {attemptDiblokir ? (
                    <Alert type="error" message="Attempt Anda diblokir. Hubungi proktor untuk unblock." showIcon />
                  ) : null}
                  {attemptSelesai ? (
                    <Link to={`/hasil/${item.attempt_id}`}>
                      <Button type="primary" block size="large">
                        Lihat Hasil
                      </Button>
                    </Link>
                  ) : (
                    <>
                      {butuhToken ? (
                        <div>
                          <label className="mb-1 block text-sm font-semibold text-ink" htmlFor={`token-${item.jadwal_id}`}>
                            Token ujian
                          </label>
                          <Input
                            id={`token-${item.jadwal_id}`}
                            autoComplete="off"
                            value={tokenPerJadwal[item.jadwal_id] ?? ""}
                            onChange={(e) =>
                              setTokenPerJadwal((prev) => ({ ...prev, [item.jadwal_id]: e.target.value }))
                            }
                            placeholder="Masukkan token dari proktor"
                            size="large"
                            className="mt-1"
                          />
                        </div>
                      ) : item.paket_jenis === "Tryout" && !attemptBerjalan ? (
                        <p className="text-sm text-muted">Tryout — tanpa token.</p>
                      ) : null}
                      {error ? <Alert type="error" message={error} showIcon /> : null}
                      <Button
                        type="primary"
                        block
                        size="large"
                        loading={memulaiJadwal === item.jadwal_id}
                        disabled={memulaiJadwal === item.jadwal_id}
                        onClick={() => void handleMasuk(item)}
                        icon={<Play size={16} />}
                      >
                        {memulaiJadwal === item.jadwal_id
                          ? "Memulai..."
                          : attemptBerjalan
                            ? "Lanjut Ujian"
                            : "Masuk Ujian"}
                      </Button>
                    </>
                  )}
                </div>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}
