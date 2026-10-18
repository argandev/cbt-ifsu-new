// HasilPage — hasil ujian siswa (Tim 4).
// Menampilkan nilai bila tampil_hasil, dan pembahasan (kunci benar yang
// dirangkum) bila Tryout atau tampil_hasil. Bila nilai disembunyikan,
// hanya pertanyaan + jawaban siswa yang tampil (tanpa benar/salah/poin).
// TAMPILAN: pertanyaan / kunci / pembahasan di-render lewat RichContent
// (HTML tersanitasi + rumus KaTeX; teks polos lama tampil sama). Jawaban
// ketikan siswa tetap tampil literal (tidak diinterpretasi sebagai HTML).
//
// REDESIGN antd (hanya presentasi): Card / Tag / Alert antd. Logika,
// endpoint, dan aturan tampil/sembunyi nilai TIDAK diubah.
import { useQuery } from "@tanstack/react-query";
import { Alert, Card, Statistic, Tag, theme, Typography } from "antd";
import { CheckCircle2, CircleX, Clock, EyeOff, MinusCircle } from "lucide-react";
import { apiClient } from "../../lib/api";
import RichContent from "../../components/soal/RichContent";

interface HasilItem {
  soal_id: string;
  nomor: number;
  tipe: string;
  pertanyaan: string;
  jawaban_siswa: unknown;
  poin?: number | null;
  bobot: number;
  ragu: boolean;
  status: string;
  kunci_tampil?: string;
  // Pembahasan guru (soal.pembahasan, migrasi 0011). Backend hanya
  // mengisinya bila Tryout ATAU tampil_hasil efektif; bila tidak,
  // field ini kosong dan tidak dirender.
  pembahasan?: string;
}

interface Hasil {
  attempt_id: string;
  jadwal_id: string;
  paket_jenis: string;
  tampil_hasil: boolean;
  menunggu_koreksi: boolean;
  nilai_akhir?: number | null;
  items: HasilItem[];
}

function tampilJawaban(j: unknown): string {
  if (j === null || j === undefined) return "— Tidak dijawab —";
  if (typeof j === "string") return j === "" ? "— Tidak dijawab —" : j;
  if (Array.isArray(j)) return j.map((x) => String(x)).join(", ");
  if (typeof j === "object") {
    return Object.entries(j as Record<string, unknown>)
      .map(([k, v]) => `${k} → ${String(v)}`)
      .join("; ");
  }
  return String(j);
}

function ikonStatus(status: string) {
  switch (status) {
    case "benar":
      return <CheckCircle2 size={20} className="shrink-0 text-success" aria-label="Benar" />;
    case "sebagian":
      return <MinusCircle size={20} className="shrink-0 text-warning" aria-label="Sebagian benar" />;
    case "salah":
      return <CircleX size={20} className="shrink-0 text-danger" aria-label="Salah" />;
    case "menunggu_koreksi":
      return <Clock size={20} className="shrink-0 text-primary" aria-label="Menunggu koreksi" />;
    default:
      return <EyeOff size={20} className="shrink-0 text-muted" aria-label="Disembunyikan" />;
  }
}

function tagStatus(status: string) {
  switch (status) {
    case "benar":
      return <Tag color="green">Benar</Tag>;
    case "sebagian":
      return <Tag color="gold">Sebagian benar</Tag>;
    case "salah":
      return <Tag color="red">Salah</Tag>;
    case "menunggu_koreksi":
      return <Tag color="blue">Menunggu koreksi</Tag>;
    default:
      return <Tag>Disembunyikan</Tag>;
  }
}

export default function HasilPage() {
  const { token } = theme.useToken();
  const attemptId = window.location.pathname.match(/\/hasil\/([0-9a-f-]{36})/i)?.[1] ?? "";
  const q = useQuery({
    queryKey: ["hasil", attemptId],
    queryFn: () => apiClient.get<Hasil>(`/api/v1/ujian/attempt/${attemptId}/hasil`),
    enabled: !!attemptId,
  });

  if (!attemptId) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-10">
        <p className="text-muted">ID attempt tidak ditemukan di URL.</p>
      </div>
    );
  }
  if (q.isLoading) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-10">
        <p className="text-muted">Memuat hasil…</p>
      </div>
    );
  }
  if (q.isError || !q.data) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-10">
        <Alert type="error" message="Hasil belum tersedia — ujian mungkin belum selesai." showIcon />
      </div>
    );
  }
  const hasil = q.data;
  return (
    <div className="mx-auto max-w-3xl px-4 py-6">
      <Typography.Title level={3} style={{ margin: "16px 0 0" }}>
        Hasil Ujian
      </Typography.Title>
      <Card className="mt-4 overflow-hidden text-center" styles={{ body: { padding: 0 } }}>
        <div className="bg-anbk px-5 py-3 text-center">
          <p className="text-sm font-semibold" style={{ color: token.colorWhite }}>Nilai Akhir</p>
        </div>
        <div className="p-5 text-center">
          {hasil.nilai_akhir !== null && hasil.nilai_akhir !== undefined ? (
            <Statistic
              value={hasil.nilai_akhir}
              precision={2}
              valueStyle={{ fontSize: 36, fontWeight: 700, color: token.colorPrimary }}
              style={{ textAlign: "center" }}
            />
          ) : (
            <p className="text-base text-muted">Nilai disembunyikan sampai koreksi final.</p>
          )}
          {hasil.menunggu_koreksi && (
            <p className="mt-2 text-base text-primary">Masih ada soal uraian yang menunggu koreksi guru.</p>
          )}
          <p className="mt-2 text-sm text-muted">
            {hasil.paket_jenis === "Tryout" ? "Tryout — pembahasan langsung tampil." : ""}
          </p>
        </div>
      </Card>

      <div className="mt-4 space-y-4">
        {hasil.items.map((it) => (
          <Card
            key={it.soal_id}
            title={
              <>
                <Typography.Text strong>Soal {it.nomor}</Typography.Text>
                {it.ragu && <Tag color="gold" className="ml-2">Ragu-ragu</Tag>}
              </>
            }
            extra={tagStatus(it.status)}
          >
            <div className="flex items-start gap-2">
              {ikonStatus(it.status)}
              <div className="min-w-0 flex-1 break-words">
                <div className="mt-2 text-base leading-relaxed">
                  <RichContent html={it.pertanyaan} />
                </div>
                <div className="mt-2 text-base text-ink">
                  <span className="font-medium">Jawaban Anda: </span>
                  {/* Jawaban objek (peta Menjodohkan {kiri: kanan}) dirender
                      PER PASANGAN (satu baris "kiri → kanan" per pasangan,
                      cerminan baris ujian) — bukan satu string diratakan.
                      Kiri/kanan berisi konten soal -> render rich. Nilai
                      kosong ("") = pasangan belum dipilih. String/array
                      (isian/uraian/label) adalah ketikan atau label siswa
                      -> tampil literal apa adanya. */}
                  {typeof it.jawaban_siswa === "object" &&
                  it.jawaban_siswa !== null &&
                  !Array.isArray(it.jawaban_siswa) ? (
                    <span className="mt-2 block space-y-2">
                      {Object.entries(it.jawaban_siswa as Record<string, unknown>).map(([kiri, kanan]) => (
                        <span key={kiri} className="flex flex-wrap items-center gap-2">
                          <span className="min-w-0">
                            <RichContent html={kiri} />
                          </span>
                          <span aria-hidden>→</span>
                          <span className="min-w-0">
                            {kanan === "" || kanan === null || kanan === undefined ? (
                              "— belum dipilih —"
                            ) : (
                              <RichContent html={String(kanan)} />
                            )}
                          </span>
                        </span>
                      ))}
                    </span>
                  ) : (
                    tampilJawaban(it.jawaban_siswa)
                  )}
                </div>
                {hasil.tampil_hasil && it.poin !== null && it.poin !== undefined && (
                  <div className="mt-2">
                    <Tag color="blue">
                      Poin: {it.poin} / {it.bobot}
                    </Tag>
                  </div>
                )}
                {hasil.tampil_hasil && it.kunci_tampil && (
                  <Alert
                    type="success"
                    showIcon
                    className="mt-2"
                    message="Kunci"
                    description={<RichContent html={it.kunci_tampil} />}
                  />
                )}
                {hasil.tampil_hasil && it.pembahasan && (
                  <Alert
                    type="info"
                    showIcon
                    className="mt-2"
                    message="Pembahasan"
                    description={<RichContent html={it.pembahasan} />}
                  />
                )}
              </div>
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
