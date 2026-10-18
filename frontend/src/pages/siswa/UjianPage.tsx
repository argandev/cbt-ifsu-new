// UjianPage — HALAMAN UTAMA ujian siswa (Tim 4).
//
// Alur: tanpa attemptId -> daftar jadwal aktif + token + tombol Mulai
// (minta fullscreen, POST /ujian/mulai). Dengan attemptId -> muat soal,
// timer dari server (polling /status tiap 30 dtk + hitung mundur lokal),
// autosave debounce 500ms, anti-curang (pindah tab / blur / keluar
// fullscreen -> POST pelanggaran; 1x peringatan, 2x diblokir),
// auto-submit saat timer habis, submit + dialog konfirmasi.
//
// Timer berjangkar server: frontend hanya menampilkan; server menolak
// save/submit setelah deadline. Kunci jawaban tidak pernah ada di sini.
//
// REDESIGN antd (hanya presentasi): Card / Button / Modal / Alert / Tag /
// Checkbox / Input antd. Layout ujian ala ANBK DIPERTAHANKAN: bilah info
// peserta + timer monospace sticky, grid kotak nomor soal, bar bawah
// Sebelumnya / Ragu-ragu / Berikutnya / Selesai. Semua handler, endpoint,
// payload, timer, anti-curang, dan teks TIDAK diubah.
import { useCallback, useEffect, useRef, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Alert, Badge, Button, Card, Checkbox, Divider, Drawer, Empty, Input, Modal, Progress, Statistic, Tag, theme, Typography } from "antd";
import { CheckCircle2, ChevronLeft, ChevronRight, Flag, LayoutGrid, Loader2 } from "lucide-react";
import { apiClient } from "../../lib/api";
import { formatJam, formatTanggal } from "../../lib/format";
import { useAuthStore } from "../../store/auth";
import Timer from "../../components/exam/Timer";
import NavigasiSoal from "../../components/exam/NavigasiSoal";
import SoalRenderer from "../../components/exam/SoalRenderer";
import RichContent from "../../components/soal/RichContent";
import BannerPeringatan from "../../components/exam/BannerPeringatan";

interface JadwalAktif {
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
  attempt_id?: string | null;
  attempt_status?: string | null;
}

interface SoalUjian {
  soal_id: string;
  nomor: number;
  tipe: string;
  pertanyaan: string;
  stimulus?: string;
  media_url?: string;
  tampilan: Record<string, unknown> | null;
  bobot: number;
  jawaban: unknown;
  ragu: boolean;
}

interface AttemptDetail {
  attempt_id: string;
  jadwal_id: string;
  status: string;
  deadline: string;
  sisa_waktu_detik: number;
  jadwal_status: string;
  paket_jenis: string;
  server_now: string;
  soal: SoalUjian[];
  nilai_akhir?: number | null;
  menunggu_koreksi?: boolean;
}

function attemptIdDariUrl(): string | null {
  const m = window.location.pathname.match(/\/ujian\/([0-9a-f-]{36})/i);
  return m ? m[1] : null;
}

function dijawab(soal: SoalUjian, jawaban: unknown): boolean {
  const j = jawaban ?? soal.jawaban;
  if (j === null || j === undefined || j === "") return false;
  if (typeof j === "string") return j.trim() !== "";
  if (Array.isArray(j)) {
    if (j.length === 0) return false;
    // boolean[] (benar-salah): dijawab bila ada yang terisi (bukan undefined).
    if (j.every((x) => x === undefined || x === null)) return false;
    return j.some((x) => x !== undefined && x !== null && x !== "");
  }
  if (typeof j === "object") {
    const vals = Object.values(j as Record<string, unknown>);
    return vals.some((v) => v !== "" && v !== undefined && v !== null);
  }
  return true;
}

export default function UjianPage() {
  const { token } = theme.useToken();
  const [attemptId, setAttemptId] = useState<string | null>(() => attemptIdDariUrl());
  const [indexAktif, setIndexAktif] = useState(0);
  const [jawabanMap, setJawabanMap] = useState<Record<string, unknown>>({});
  const [raguMap, setRaguMap] = useState<Record<string, boolean>>({});
  const [statusSimpan, setStatusSimpan] = useState<"idle" | "menyimpan" | "tersimpan" | "gagal">("idle");
  const [peringatan, setPeringatan] = useState<{ jumlah: number; tampil: boolean }>({ jumlah: 0, tampil: false });
  const [diblokir, setDiblokir] = useState(false);
  const [showKonfirmasi, setShowKonfirmasi] = useState(false);
  const [showNav, setShowNav] = useState(false);
  const [hasilSubmit, setHasilSubmit] = useState<{ nilai: number | null; menunggu: boolean; pesan?: string } | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [tokenPerJadwal, setTokenPerJadwal] = useState<Record<string, string>>({});
  const [memulai, setMemulai] = useState(false);

  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const terakhirLaporRef = useRef(0);
  const fullscreenAktifRef = useRef(false);
  const sudahSubmitRef = useRef(false);
  const attemptIdRef = useRef(attemptId);
  attemptIdRef.current = attemptId;

  // --- Daftar jadwal aktif (bila belum ada attemptId) ---
  const jadwalQuery = useQuery({
    queryKey: ["ujian-aktif"],
    queryFn: () => apiClient.get<{ items: JadwalAktif[] }>("/api/v1/ujian/aktif"),
    enabled: !attemptId,
    refetchInterval: 30000,
  });

  // --- Detail attempt ---
  const attemptQuery = useQuery({
    queryKey: ["attempt", attemptId],
    queryFn: () => apiClient.get<AttemptDetail>(`/api/v1/ujian/attempt/${attemptId}`),
    enabled: !!attemptId,
    refetchInterval: false,
  });

  // --- Status polling tiap 30 dtk (timer server + deteksi blokir/selesai) ---
  const statusQuery = useQuery({
    queryKey: ["attempt-status", attemptId],
    queryFn: () =>
      apiClient.get<{
        status: string;
        sisa_waktu_detik: number;
        jadwal_status: string;
        dijawab: number;
        ragu: number;
        total_soal: number;
      }>(`/api/v1/ujian/attempt/${attemptId}/status`),
    enabled: !!attemptId && !hasilSubmit,
    refetchInterval: 30000,
  });

  const attempt = attemptQuery.data;
  const soalList = attempt?.soal ?? [];
  const soalAktif = soalList[indexAktif];

  // --- Bilah info peserta (hanya tampil; tidak mengubah alur ujian) ---
  // Nama/NISN/kelas/mapel TIDAK dikirim endpoint detail attempt, jadi
  // dibaca dari kartu peserta yang sudah ada (GET /api/v1/kartu/...).
  // Bila kartu belum tersedia (404), bilah memakai nama akun login dan
  // tanda "—" untuk field yang tidak ada — ujian tetap jalan normal.
  const authUser = useAuthStore((s) => s.user);
  const kartuQuery = useQuery({
    queryKey: ["kartu-ujian", attempt?.jadwal_id ?? "", authUser?.id ?? ""],
    queryFn: () =>
      apiClient.get<{ nama_siswa: string; nisn: string; kelas: string; mapel: string }>(
        `/api/v1/kartu/${attempt!.jadwal_id}/${authUser!.id}`,
      ),
    enabled: !!attemptId && !!attempt?.jadwal_id && !!authUser?.id,
    retry: false,
  });
  const kartu = kartuQuery.data;

  // Inisialisasi jawaban + ragu dari server.
  useEffect(() => {
    if (!attempt) return;
    const jm: Record<string, unknown> = {};
    const rm: Record<string, boolean> = {};
    for (const s of attempt.soal) {
      if (s.jawaban !== null && s.jawaban !== undefined) jm[s.soal_id] = s.jawaban;
      rm[s.soal_id] = s.ragu;
    }
    setJawabanMap(jm);
    setRaguMap(rm);
    if (attempt.status === "diblokir") setDiblokir(true);
    if (attempt.status === "selesai") {
      setHasilSubmit({
        nilai: attempt.nilai_akhir ?? null,
        menunggu: attempt.menunggu_koreksi ?? false,
      });
    }
  }, [attempt?.attempt_id]); // eslint-disable-line react-hooks/exhaustive-deps

  // Status polling: blokir / selesai dari server (mis. dijeda proktor).
  useEffect(() => {
    const st = statusQuery.data;
    if (!st) return;
    if (st.status === "diblokir") setDiblokir(true);
    if (st.status === "selesai" && !hasilSubmit) {
      setHasilSubmit({ nilai: null, menunggu: false, pesan: "Waktu ujian habis — jawaban tersimpan otomatis." });
    }
  }, [statusQuery.data]); // eslint-disable-line react-hooks/exhaustive-deps

  // --- Autosave (debounce 500ms) ---
  const simpan = useCallback(
    (soalId: string, nilai: unknown) => {
      const id = attemptIdRef.current;
      if (!id) return;
      setStatusSimpan("menyimpan");
      apiClient
        .post(`/api/v1/ujian/attempt/${id}/jawaban`, { soal_id: soalId, jawaban: nilai })
        .then(() => setStatusSimpan("tersimpan"))
        .catch(() => setStatusSimpan("gagal"));
    },
    [],
  );

  const ubahJawaban = (soalId: string, nilai: unknown) => {
    setJawabanMap((prev) => ({ ...prev, [soalId]: nilai }));
    setStatusSimpan("menyimpan");
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => simpan(soalId, nilai), 500);
  };

  const ubahRagu = (soalId: string, ragu: boolean) => {
    setRaguMap((prev) => ({ ...prev, [soalId]: ragu }));
    const id = attemptIdRef.current;
    if (id) {
      apiClient.post(`/api/v1/ujian/attempt/${id}/ragu`, { soal_id: soalId, ragu }).catch(() => undefined);
    }
  };

  // --- Anti-curang: pindah tab / blur / keluar fullscreen ---
  const laporPelanggaran = useCallback((jenis: string) => {
    const id = attemptIdRef.current;
    if (!id || sudahSubmitRef.current) return;
    const now = Date.now();
    if (now - terakhirLaporRef.current < 5000) return; // dedupe client 5 dtk
    terakhirLaporRef.current = now;
    apiClient
      .post<{ peringatan?: boolean; diblokir?: boolean; jumlah_pelanggaran?: number }>(
        `/api/v1/ujian/attempt/${id}/pelanggaran`,
        { jenis },
      )
      .then((res) => {
        if (res.diblokir) {
          setDiblokir(true);
        } else if (res.peringatan) {
          setPeringatan({ jumlah: res.jumlah_pelanggaran ?? 1, tampil: true });
        }
      })
      .catch(() => undefined);
  }, []);

  useEffect(() => {
    if (!attemptId || hasilSubmit) return;
    const onVisibility = () => {
      if (document.hidden) laporPelanggaran("pindah_tab");
    };
    const onBlur = () => laporPelanggaran("blur");
    const onFullscreen = () => {
      if (!document.fullscreenElement && fullscreenAktifRef.current) {
        fullscreenAktifRef.current = false;
        laporPelanggaran("fullscreen_exit");
      }
    };
    document.addEventListener("visibilitychange", onVisibility);
    window.addEventListener("blur", onBlur);
    document.addEventListener("fullscreenchange", onFullscreen);
    return () => {
      document.removeEventListener("visibilitychange", onVisibility);
      window.removeEventListener("blur", onBlur);
      document.removeEventListener("fullscreenchange", onFullscreen);
    };
  }, [attemptId, hasilSubmit, laporPelanggaran]);

  // --- Submit ---
  const doSubmit = useCallback(() => {
    const id = attemptIdRef.current;
    if (!id || sudahSubmitRef.current) return;
    sudahSubmitRef.current = true;
    setShowKonfirmasi(false);
    apiClient
      .post<{ nilai_akhir?: number | null; menunggu_koreksi?: boolean; pesan?: string }>(
        `/api/v1/ujian/attempt/${id}/submit`,
        {},
      )
      .then((res) => {
        setHasilSubmit({
          nilai: res.nilai_akhir ?? null,
          menunggu: res.menunggu_koreksi ?? false,
          pesan: res.pesan,
        });
      })
      .catch((e: Error) => {
        sudahSubmitRef.current = false;
        setError(e.message);
      });
  }, []);

  const onTimerHabis = useCallback(() => {
    doSubmit();
  }, [doSubmit]);

  // --- Mulai ujian (dari daftar jadwal) ---
  const mulai = async (jadwalId: string, paketJenis: string) => {
    setError(null);
    setMemulai(true);
    try {
      // Minta fullscreen saat mulai (keluar fullscreen = pelanggaran).
      try {
        await document.documentElement.requestFullscreen();
        fullscreenAktifRef.current = true;
      } catch {
        // Fullscreen ditolak browser -> ujian tetap jalan; tidak penalti.
      }
      const res = await apiClient.post<{ attempt_id: string }>(`/api/v1/ujian/mulai`, {
        jadwal_id: jadwalId,
        token: paketJenis === "Tryout" ? undefined : tokenPerJadwal[jadwalId] ?? "",
      });
      setAttemptId(res.attempt_id);
      window.history.pushState(null, "", `/ujian/${res.attempt_id}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Gagal memulai ujian.");
    } finally {
      setMemulai(false);
    }
  };

  // --- Render: hasil submit / selesai ---
  if (hasilSubmit) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-10">
        <Card className="overflow-hidden" styles={{ body: { padding: 0 } }}>
          <div className="p-6 text-center">
            <Typography.Title level={3} style={{ margin: 0 }}>
              Ujian Selesai
            </Typography.Title>
            <p className="mt-2 text-base text-muted">
              Jawaban Anda sudah tersimpan.
              {hasilSubmit.pesan ? ` ${hasilSubmit.pesan}` : ""}
            </p>
            {hasilSubmit.nilai !== null && hasilSubmit.nilai !== undefined ? (
              <Statistic
                value={hasilSubmit.nilai}
                precision={2}
                valueStyle={{ fontSize: 36, fontWeight: 700, color: token.colorPrimary }}
                style={{ textAlign: "center" }}
                className="mt-4"
              />
            ) : (
              <p className="mt-4 text-base text-muted">
                Nilai disembunyikan sampai koreksi final.
                {hasilSubmit.menunggu ? " (Masih ada soal uraian yang menunggu koreksi guru.)" : ""}
              </p>
            )}
            <div className="mt-6 flex justify-center gap-3">
              <Button type="primary" size="large" href={`/hasil/${attemptId}`}>
                Lihat Hasil
              </Button>
              <Button size="large" href="/">
                Kembali ke Beranda
              </Button>
            </div>
          </div>
        </Card>
      </div>
    );
  }

  // --- Render: daftar jadwal (belum mulai) ---
  if (!attemptId) {
    const items = jadwalQuery.data?.items ?? [];
    return (
      <div className="mx-auto max-w-3xl px-4 py-6">
        <Typography.Title level={3} style={{ margin: "16px 0 0" }}>
          Ujian Hari Ini
        </Typography.Title>
        {error && <Alert type="error" message={error} showIcon className="mt-3" />}
        {jadwalQuery.isLoading && <p className="mt-4 text-muted">Memuat jadwal…</p>}
        {!jadwalQuery.isLoading && items.length === 0 && (
          <Empty className="mt-4" description="Tidak ada jadwal ujian untuk kelas Anda hari ini." />
        )}
        <div className="mt-4 space-y-4">
          {items.map((j) => (
            <Card
              key={j.jadwal_id}
              title={j.paket_nama}
              extra={<Tag color="blue">{j.paket_jenis}</Tag>}
              className="overflow-hidden"
            >
              <p className="text-sm text-muted">
                {formatTanggal(j.tanggal)} · {formatJam(j.jam_mulai)}–{formatJam(j.jam_selesai)} WIB · {j.kelas_nama} · {j.durasi_menit} menit
              </p>
              {j.attempt_status === "berjalan" && j.attempt_id ? (
                <Button
                  type="primary"
                  size="large"
                  className="mt-3 w-full sm:w-auto"
                  onClick={() => setAttemptId(j.attempt_id ?? null)}
                >
                  Lanjutkan Ujian
                </Button>
              ) : j.attempt_status === "diblokir" ? (
                <Alert
                  type="error"
                  message="Attempt Anda diblokir. Hubungi proktor untuk unblock."
                  showIcon
                  className="mt-3"
                />
              ) : j.attempt_status === "selesai" ? (
                <p className="mt-3 text-sm text-muted">Anda sudah menyelesaikan ujian ini.</p>
              ) : j.status === "pause" ? (
                <Alert type="warning" message="Ujian sedang dijeda proktor." showIcon className="mt-3" />
              ) : (
                <div className="mt-3 flex flex-col gap-2 sm:flex-row">
                  {j.paket_jenis !== "Tryout" && (
                    <Input
                      value={tokenPerJadwal[j.jadwal_id] ?? ""}
                      onChange={(e) =>
                        setTokenPerJadwal((prev) => ({ ...prev, [j.jadwal_id]: e.target.value.toUpperCase() }))
                      }
                      placeholder="Masukkan token ujian (6 karakter)"
                      aria-label={`Token untuk ${j.paket_nama}`}
                      maxLength={6}
                      size="large"
                      style={{ fontFamily: "monospace", textTransform: "uppercase" }}
                      className="flex-1"
                    />
                  )}
                  <Button
                    type="primary"
                    size="large"
                    loading={memulai}
                    disabled={memulai}
                    onClick={() => mulai(j.jadwal_id, j.paket_jenis)}
                  >
                    {memulai ? "Memulai…" : "Mulai Ujian"}
                  </Button>
                </div>
              )}
            </Card>
          ))}
        </div>
      </div>
    );
  }

  // --- Render: loading / error attempt ---
  if (attemptQuery.isLoading) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-10">
        <p className="text-muted">Memuat soal ujian…</p>
      </div>
    );
  }
  if (attemptQuery.isError || !attempt) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-10">
        <Alert type="error" message="Gagal memuat ujian. Coba muat ulang halaman." showIcon />
      </div>
    );
  }

  if (diblokir) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-10">
        <BannerPeringatan mode="diblokir" jumlahPelanggaran={2} />
      </div>
    );
  }

  const dijawabCount = soalList.filter((s) => dijawab(s, jawabanMap[s.soal_id])).length;
  const belumCount = soalList.length - dijawabCount;
  const belumNomor = soalList.filter((s) => !dijawab(s, jawabanMap[s.soal_id])).map((s) => s.nomor);
  const persenDijawab = soalList.length > 0 ? Math.round((dijawabCount / soalList.length) * 100) : 0;
  const sisaServer = statusQuery.data?.sisa_waktu_detik ?? attempt.sisa_waktu_detik;
  const jadwalDijeda = (statusQuery.data?.jadwal_status ?? attempt.jadwal_status) === "pause";

  return (
    // Root: lebar mengikuti shell Layout (1100px) — TANPA max-w-7xl sendiri
    // dan TANPA px-4 (padding sudah dari Layout Content). Ini bagian dari
    // perbaikan P-01: bar bawah tidak lagi memakai trik -mx-4.
    <div className="mx-auto py-4">
      {/* Slot banner peringatan: tinggi TETAP 40px (= tinggi banner kompak
          satu baris). Muncul/hilangnya banner TIDAK menggeser soal (P-08).
          Saat tidak ada peringatan, slot ini hanya ruang kosong 40px. */}
      <div className="mb-4 flex h-10 items-center">
        {peringatan.tampil && (
          <BannerPeringatan
            mode="peringatan"
            jumlahPelanggaran={peringatan.jumlah}
            onTutup={() => setPeringatan((p) => ({ ...p, tampil: false }))}
          />
        )}
      </div>
      {error && <Alert type="error" message={error} showIcon className="mb-4" />}
      {jadwalDijeda && (
        <Alert
          type="warning"
          message="Ujian dijeda proktor — timer dibekukan. Tunggu hingga dilanjutkan."
          showIcon
          className="mb-4"
        />
      )}

      {/* Header ujian SATU-SATUNYA (tanpa bilah brand biru ganda — brand
          sudah di header global Layout). Sticky top-0. PERSIS 2 baris ×
          48px di lebar apa pun (P-09): baris tidak pernah wrap — identitas
          dipotong elipsis bila sempit — sehingga offset sticky navigasi
          (top-28 = 96 + 16 celah) selalu tepat. */}
      <div className="sticky top-0 z-20 mb-4 border-b border-border" style={{ backgroundColor: token.colorBgContainer }}>
        {/* Baris 1: identitas (satu baris, elipsis) + status simpan + timer. */}
        <div className="flex h-12 items-center justify-between gap-2 px-4">
          <p className="min-w-0 flex-1 truncate text-sm text-ink">
            <span className="text-muted">Nama: </span>
            <span className="font-semibold">{kartu?.nama_siswa ?? authUser?.nama ?? "—"}</span>
            <span className="text-muted"> · NISN: </span>
            <span className="font-semibold">{kartu?.nisn ?? "—"}</span>
            <span className="text-muted"> · Kelas: </span>
            <span className="font-semibold">{kartu?.kelas ?? "—"}</span>
            <span className="text-muted"> · Mapel: </span>
            <span className="font-semibold">{kartu?.mapel ?? "—"}</span>
          </p>
          {/* Status simpan: ikon + warna, hanya saat aktif (idle = kosong,
              seperti sebelumnya). Lebar menyusut dari identitas (truncate),
              bukan menggeser timer. */}
          {statusSimpan !== "idle" && (
            <span aria-live="polite" className="flex shrink-0 items-center gap-1 text-sm">
              {statusSimpan === "menyimpan" && (
                <>
                  <Loader2 size={16} className="shrink-0 animate-spin text-muted" />
                  <span className="text-muted">Menyimpan…</span>
                </>
              )}
              {statusSimpan === "tersimpan" && (
                <>
                  <CheckCircle2 size={16} className="shrink-0 text-success" />
                  <span className="text-success">Tersimpan</span>
                </>
              )}
              {statusSimpan === "gagal" && (
                <Alert
                  type="error"
                  showIcon
                  message="Gagal — coba lagi"
                  className="shrink-0"
                  style={{ padding: "4px 12px" }}
                />
              )}
            </span>
          )}
          <div className="shrink-0">
            <Timer sisaDetik={sisaServer} onExpire={onTimerHabis} />
          </div>
        </div>
        {/* Baris 2: progres + tombol Daftar Soal. */}
        <div className="flex h-12 items-center justify-between gap-2 px-4">
          <div className="flex min-w-0 flex-1 items-center gap-2 text-sm">
            <span className="shrink-0 whitespace-nowrap font-semibold text-ink">
              Soal {soalList.length > 0 ? `${indexAktif + 1} dari ${soalList.length}` : "—"}
            </span>
            <Progress
              percent={persenDijawab}
              showInfo={false}
              size="small"
              className="min-w-16 flex-1 sm:max-w-32"
              aria-label={`Progres dijawab ${persenDijawab} persen`}
            />
            <span className="hidden shrink-0 text-muted sm:inline">
              Dijawab {dijawabCount} / {soalList.length}
            </span>
          </div>
          {/* Pembungkus <div> biasa: `lg:hidden` TIDAK boleh ditempel
              langsung di antd Button — `.ant-btn{display:inline-flex}`
              (unlayered) mengalahkan utilitas Tailwind ber-layer.
              Pembungkus div-lah yang disembunyikan di lg+. */}
          <div className="shrink-0 lg:hidden">
            <Badge count={belumCount} offset={[-4, 4]} size="small">
              <Button
                onClick={() => setShowNav(true)}
                aria-expanded={showNav}
                icon={<LayoutGrid size={18} />}
              >
                Daftar Soal
              </Button>
            </Badge>
          </div>
        </div>
      </div>

      {/* Grid soal : navigasi. items-start = anak TIDAK stretch (P-03);
          kolom navigasi tidak lagi memanjang setinggi kolom soal. */}
      <div className="flex flex-col gap-4 lg:flex-row lg:items-start">
        {/* Kolom soal — lebar baca dibatasi 768px (P-04). */}
        <section
          className="min-w-0 flex-1"
          onCopy={(e) => e.preventDefault()}
          onPaste={(e) => e.preventDefault()}
          onCut={(e) => e.preventDefault()}
        >
          {soalAktif ? (
            <Card
              title={`Soal ${soalAktif.nomor}`}
              extra={<Tag color="blue">Bobot {soalAktif.bobot}</Tag>}
              className="mx-auto w-full max-w-3xl overflow-hidden"
              styles={{ body: { padding: 0 } }}
            >
              <div className="p-4 sm:p-6">
                {soalAktif.stimulus && (
                  <div>
                    <p className="mb-2 text-xs font-semibold text-muted">Stimulus</p>
                    <div className="break-words rounded-lg border border-border bg-surface px-4 py-3 text-base leading-relaxed text-ink">
                      <RichContent html={soalAktif.stimulus} />
                    </div>
                    <Divider style={{ margin: "16px 0" }} />
                  </div>
                )}
                <p className="mb-2 text-xs font-semibold text-muted">Pertanyaan</p>
                <div className="break-words text-base leading-relaxed text-ink">
                  <RichContent html={soalAktif.pertanyaan} />
                </div>
                {soalAktif.media_url && (
                  <figure className="mt-4 text-center">
                    <img
                      src={soalAktif.media_url}
                      alt="Media soal"
                      className="mx-auto max-h-72 max-w-full rounded-lg object-contain"
                    />
                    <figcaption className="mt-2 text-xs text-muted">Media soal</figcaption>
                  </figure>
                )}
                <div className="mt-4">
                  <SoalRenderer
                    tipe={soalAktif.tipe}
                    tampilan={soalAktif.tampilan}
                    pertanyaan={soalAktif.pertanyaan}
                    value={jawabanMap[soalAktif.soal_id]}
                    onChange={(v) => ubahJawaban(soalAktif.soal_id, v)}
                  />
                </div>
              </div>
            </Card>
          ) : (
            <p className="text-muted">Tidak ada soal dalam paket ini.</p>
          )}
        </section>

        {/* Navigasi soal DESKTOP (≥1024px): sidebar sticky.
            Pembungkusnya <div> BIASA (bukan Card antd) — karena posisi
            sticky Tailwind kalah cascade dari `.ant-card{position:relative}`
            milik antd bila ditempel langsung di Card (itulah akar D-01:
            `lg:sticky` tidak pernah berlaku, tapi `lg:top-36` tetap
            menggeser kartu 144px ke bawah hingga menimpa bar).
            Wrapper ini: align-self flex-start + max-height viewport −
            header − bar + scroll internal (P-03). */}
        <div className="hidden w-full shrink-0 lg:sticky lg:top-28 lg:block lg:max-h-[calc(100vh-200px)] lg:w-80 lg:self-start lg:overflow-y-auto">
          <Card title="Navigasi Soal">
            <Progress percent={persenDijawab} showInfo={false} size="small" aria-label={`Progres dijawab ${persenDijawab} persen`} />
            <p className="mt-2 text-sm text-muted">
              Dijawab: {dijawabCount} / {soalList.length} · Belum: {belumCount}
            </p>
            <div className="mt-4">
              <NavigasiSoal
                items={soalList.map((s) => ({
                  soal_id: s.soal_id,
                  nomor: s.nomor,
                  dijawab: dijawab(s, jawabanMap[s.soal_id]),
                  ragu: raguMap[s.soal_id] ?? false,
                }))}
                aktif={indexAktif}
                onPilih={(i) => {
                  setIndexAktif(i);
                  setShowNav(false);
                }}
              />
            </div>
          </Card>
        </div>
      </div>

      {/* Drawer navigasi (<1024px): overlay, TIDAK mendorong konten (P-03). */}
      <Drawer
        title="Navigasi Soal"
        placement="right"
        width={320}
        open={showNav}
        onClose={() => setShowNav(false)}
      >
        <Progress percent={persenDijawab} showInfo={false} size="small" aria-label={`Progres dijawab ${persenDijawab} persen`} />
        <p className="mt-2 text-sm text-muted">
          Dijawab: {dijawabCount} / {soalList.length} · Belum: {belumCount}
        </p>
        <div className="mt-4">
          <NavigasiSoal
            items={soalList.map((s) => ({
              soal_id: s.soal_id,
              nomor: s.nomor,
              dijawab: dijawab(s, jawabanMap[s.soal_id]),
              ragu: raguMap[s.soal_id] ?? false,
            }))}
            aktif={indexAktif}
            onPilih={(i) => {
              setIndexAktif(i);
              setShowNav(false);
            }}
          />
        </div>
      </Drawer>

      {/* Spacer: ruang kosong SETINGGI bar + 16px di antara konten terakhir
          dan bar bawah, agar bar sticky tidak pernah menutupi konten (D-02).
          Tinggi bar: HP 3 baris ≈ 173px → spacer 192px; sm+ ≤ 2 baris →
          spacer 144px. */}
      <div aria-hidden className="h-48 sm:h-36" />

      {/* Bar bawah ujian (sticky): SATU elemen, SATU latar, SATU border-t,
          DI LUAR grid dua kolom (P-01). Tanpa -mx-4, tanpa max-w-7xl.
          Satu-satunya primary di tampilan ini: "Selesai Ujian" (P-02).
          Di <sm, grup kanan ditumpuk (ragu-ragu di atas Selesai) karena
          label tetap "Tandai ragu-ragu" + tombol Selesai tidak muat
          sebaris 296px tanpa memotong kata — bar tetap SATU kotak. */}
      {soalAktif && (
        <div className="sticky bottom-0 z-20 border-t border-border px-4 py-3" style={{ backgroundColor: token.colorBgContainer }}>
          <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-center sm:justify-between">
            <div className="flex items-center gap-2">
              <Button
                size="large"
                disabled={indexAktif === 0}
                onClick={() => setIndexAktif((i) => Math.max(0, i - 1))}
                icon={<ChevronLeft size={18} />}
                className="flex-1 sm:flex-none"
              >
                Sebelumnya
              </Button>
              <Button
                size="large"
                disabled={indexAktif >= soalList.length - 1}
                onClick={() => setIndexAktif((i) => Math.min(soalList.length - 1, i + 1))}
                className="flex-1 sm:flex-none"
              >
                Berikutnya <ChevronRight size={18} />
              </Button>
            </div>
            <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
              <Checkbox
                checked={raguMap[soalAktif.soal_id] ?? false}
                onChange={(e) => ubahRagu(soalAktif.soal_id, e.target.checked)}
                className="flex min-h-[44px] min-w-0 items-center gap-2 whitespace-nowrap rounded-lg border border-warning bg-warning-bg px-3 text-base font-medium text-ink"
              >
                <Flag size={16} className="text-warning" />
                Tandai ragu-ragu
              </Checkbox>
              <Button
                type="primary"
                size="large"
                onClick={() => setShowKonfirmasi(true)}
                className="w-full shrink-0 sm:w-auto"
              >
                Selesai Ujian
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* Dialog konfirmasi submit (Modal antd) */}
      <Modal
        open={showKonfirmasi}
        title="Kumpulkan Jawaban?"
        okText="Ya, Kumpulkan"
        cancelText="Batal"
        onOk={doSubmit}
        onCancel={() => setShowKonfirmasi(false)}
      >
        <p className="text-base text-muted">
          Dijawab: {dijawabCount} dari {soalList.length} soal.
          {belumCount > 0 ? ` Masih ada ${belumCount} soal belum dijawab (dihitung 0).` : " Semua soal sudah dijawab."}
        </p>
        {belumCount > 0 && (
          <div className="mt-2 flex flex-wrap items-center gap-2">
            <span className="text-sm text-muted">Soal belum dijawab:</span>
            {belumNomor.map((n) => (
              <Tag key={n} color="gold">
                Soal {n}
              </Tag>
            ))}
          </div>
        )}
        <p className="mt-2 text-sm text-muted">
          Setelah dikumpulkan, jawaban tidak bisa diubah.
        </p>
      </Modal>
    </div>
  );
}
