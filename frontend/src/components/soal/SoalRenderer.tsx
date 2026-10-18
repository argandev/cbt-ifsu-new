import { AlertTriangle, CheckCircle2 } from "lucide-react";
import { Alert, Tag } from "antd";
import type { Soal } from "../../lib/banksoal";
import { TIPE_SOAL, badgeKelompok } from "../../lib/tipeSoal";
import RichContent, { embedUrlFromWatchUrl, stripHtmlToText } from "./RichContent";

interface Props {
  soal: Soal;
  /** Tampilkan kunci — hanya untuk halaman GURU. Jangan pernah true di halaman siswa. */
  showKunci?: boolean;
}

interface Opsi {
  label: string;
  teks: string;
}

interface Pasangan {
  kiri: string;
  kanan: string;
}

function asOpsi(v: unknown): Opsi[] {
  if (!Array.isArray(v)) return [];
  return v as Opsi[];
}

function asStringArray(v: unknown): string[] {
  if (!Array.isArray(v)) return [];
  return v as string[];
}

/**
 * Media soal: URL gambar -> <img>; URL YouTube/Vimeo -> <iframe> embed.
 * (Sebelumnya semua media_url selalu di-render <img>, sehingga link
 * video tampil sebagai gambar rusak.)
 */
function MediaSoal({ url }: { url: string }) {
  const embed = embedUrlFromWatchUrl(url);
  if (embed) {
    return (
      <iframe
        src={embed}
        title="Video soal"
        allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
        allowFullScreen
        loading="lazy"
        className="rich-video-frame mt-3"
      />
    );
  }
  return (
    <img src={url} alt="Media soal" loading="lazy" className="mt-3 max-h-64 rounded-lg object-contain" />
  );
}

/** SoalRenderer — pratinjau soal untuk GURU (read-only, semua tipe).
 *  Dengan showKunci, kunci ditampilkan karena ini halaman guru.
 *  JANGAN dipakai di halaman siswa (kunci tidak boleh ke browser siswa).
 *  Semua teks soal di-render lewat RichContent: HTML disanitasi
 *  (DOMPurify) + rumus LaTeX di-render KaTeX; teks polos lama tampil
 *  persis seperti sebelumnya.
 *
 *  REDESIGN antd (hanya presentasi): Tag untuk badge, Alert untuk
 *  kotak kunci / rubrik / pembahasan. Logika & isi TIDAK diubah.
 */
export default function SoalRenderer({ soal, showKunci }: Props) {
  const info = TIPE_SOAL[soal.tipe];
  const jawaban = soal.jawaban ?? {};
  const kunci = soal.kunci ?? {};
  const kunciOpsi = kunci.opsi;
  const kunciMulti: string[] = Array.isArray(kunciOpsi) ? (kunciOpsi as string[]) : [];

  const isKunciOpsi = (label: string) =>
    showKunci && (kunciOpsi === label || kunciMulti.includes(label));

  const renderJawaban = () => {
    switch (soal.tipe) {
      case "pg":
      case "studi_kasus":
      case "pg_kompleks":
      case "pg_multi": {
        const multi = soal.tipe === "pg_kompleks" || soal.tipe === "pg_multi";
        return (
          <ul className="space-y-2">
            {asOpsi(jawaban.opsi).map((o) => (
              <li
                key={o.label}
                className={`flex min-h-[44px] items-center gap-2 rounded-lg border px-3 py-2 text-[16px] ${
                  isKunciOpsi(o.label) ? "border-2 border-success bg-success-bg" : "border-border bg-white"
                }`}
              >
                <span className="font-semibold">{o.label}.</span>
                <span className="min-w-0 flex-1"><RichContent html={o.teks} /></span>
                {isKunciOpsi(o.label) && (
                  <span className="ml-auto inline-flex shrink-0 items-center gap-1 text-xs font-semibold text-success">
                    <CheckCircle2 size={14} /> Kunci
                  </span>
                )}
                {multi && showKunci && <span className="sr-only">opsi kunci</span>}
              </li>
            ))}
          </ul>
        );
      }
      case "pilihan_gambar":
        return (
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
            {asOpsi(jawaban.opsi).map((o) => (
              <figure
                key={o.label}
                className={`rounded-lg border p-2 ${
                  isKunciOpsi(o.label) ? "border-2 border-success bg-success-bg" : "border-border bg-white"
                }`}
              >
                {/* opsi.teks pada tipe ini adalah URL gambar (upload via ImageUpload) — bukan teks. */}
                <img src={o.teks} alt={`Opsi ${o.label}`} loading="lazy" className="h-24 w-full rounded object-cover" />
                <figcaption className="mt-1 text-center text-sm font-semibold">
                  {o.label}{isKunciOpsi(o.label) ? " — Kunci" : ""}
                </figcaption>
              </figure>
            ))}
          </div>
        );
      case "benar_salah": {
        const pernyataan = asStringArray(jawaban.pernyataan);
        const nilai = Array.isArray(kunci.nilai) ? (kunci.nilai as boolean[]) : [];
        return (
          <ul className="space-y-2">
            {pernyataan.map((p, i) => (
              <li key={i} className="flex min-h-[44px] items-center justify-between gap-2 rounded-lg border border-border bg-white px-3 py-2 text-[16px]">
                <span className="min-w-0 flex-1"><RichContent html={p} /></span>
                <span className="shrink-0">
                  Benar / Salah
                  {showKunci && <strong className="ml-2 text-success">Kunci: {nilai[i] ? "Benar" : "Salah"}</strong>}
                </span>
              </li>
            ))}
          </ul>
        );
      }
      case "pernyataan_sikap": {
        const pernyataan = asStringArray(jawaban.pernyataan);
        const skala = Array.isArray(jawaban.skala) ? (jawaban.skala as number[]) : [1, 2, 3, 4];
        return (
          <ul className="space-y-2">
            {pernyataan.map((p, i) => (
              <li key={i} className="rounded-lg border border-border bg-white px-3 py-2 text-[16px]">
                <div><RichContent html={p} /></div>
                <div className="mt-1 flex gap-2">
                  {skala.map((s) => (
                    <span key={s} className="rounded-full border border-border px-2 py-0.5 text-xs">{s}</span>
                  ))}
                </div>
              </li>
            ))}
          </ul>
        );
      }
      case "menjodohkan":
      case "peta_konsep": {
        const pasangan = Array.isArray(jawaban.pasangan) ? (jawaban.pasangan as Pasangan[]) : [];
        return (
          <ul className="space-y-2">
            {pasangan.map((p, i) => (
              <li key={i} className="flex min-h-[44px] items-center gap-2 rounded-lg border border-border bg-white px-3 py-2 text-[16px]">
                <span className="min-w-0 flex-1"><RichContent html={p.kiri} /></span>
                <span className="shrink-0 text-muted">→</span>
                <span className="min-w-0 flex-1"><RichContent html={p.kanan} /></span>
              </li>
            ))}
          </ul>
        );
      }
      case "isian":
        return <p className="text-sm text-muted">Jawaban diketik siswa (isian singkat).</p>;
      case "uraian":
        return <p className="text-sm text-muted">Jawaban bebas diketik siswa, dikoreksi manual oleh guru.</p>;
      case "urutkan": {
        const items = asStringArray(jawaban.item);
        const urutan = Array.isArray(kunci.urutan) ? (kunci.urutan as number[]) : [];
        return (
          <ol className="space-y-2">
            {items.map((item, i) => (
              <li key={i} className="flex items-baseline gap-2 rounded-lg border border-border bg-white px-3 py-2 text-[16px]">
                <span className="shrink-0">{i + 1}.</span>
                <span className="min-w-0 flex-1"><RichContent html={item} /></span>
                {showKunci && urutan.length === items.length && (
                  <span className="ml-2 shrink-0 text-xs text-success">Posisi benar: {urutan.indexOf(i) + 1}</span>
                )}
              </li>
            ))}
          </ol>
        );
      }
      case "isi_ganda": {
        const kolom = asStringArray(jawaban.kolom);
        return (
          <div className="space-y-2">
            {kolom.map((k, i) => (
              <div key={i} className="flex items-center gap-2 text-sm">
                <span className="w-32 shrink-0 font-medium"><RichContent html={k} /></span>
                <span className="flex-1 rounded-lg border border-dashed border-border px-3 py-2 text-muted">Isian siswa</span>
              </div>
            ))}
          </div>
        );
      }
      case "teks_rumpang":
        return <p className="text-sm text-muted">Rumpang (___) diisi siswa langsung di paragraf pertanyaan.</p>;
      case "upload_file":
        return <p className="text-sm text-muted">Siswa mengunggah berkas jawaban. Nilai manual oleh guru/penguji.</p>;
      case "praktik_singkat":
        return <p className="text-sm text-muted">Siswa menjawab dengan teks + berkas opsional. Nilai manual dengan rubrik.</p>;
      default:
        return <p className="text-sm text-muted">Pratinjau belum tersedia untuk tipe ini.</p>;
    }
  };

  const kunciText = () => {
    if (!showKunci) return null;
    switch (soal.tipe) {
      case "pg":
      case "studi_kasus":
      case "pilihan_gambar":
        return `Kunci: opsi ${String(kunci.opsi ?? "-")}`;
      case "pg_kompleks":
      case "pg_multi":
        return `Kunci: opsi ${kunciMulti.join(", ") || "-"}`;
      case "isian":
      case "isi_ganda":
      case "teks_rumpang":
        return `Kunci: ${asStringArray(kunci.jawaban).join(" | ") || "-"}`;
      case "urutkan": {
        const items = asStringArray(jawaban.item);
        const urutan = Array.isArray(kunci.urutan) ? (kunci.urutan as number[]) : [];
        return `Urutan benar: ${urutan.map((idx) => stripHtmlToText(items[idx] ?? "")).join(" → ") || "-"}`;
      }
      case "pernyataan_sikap":
        return "Tidak dinilai (survei/refleksi).";
      default:
        return null;
    }
  };

  /** Kunci pasangan (Menjodohkan/Peta Konsep) + rubrik: di-render rich agar rumus ikut tampil. */
  const kunciRich = () => {
    if (!showKunci) return null;
    if (soal.tipe === "menjodohkan" || soal.tipe === "peta_konsep") {
      const pas = Array.isArray(kunci.pasangan) ? (kunci.pasangan as Pasangan[]) : [];
      if (pas.length === 0) return <Alert type="success" showIcon className="mt-3" message="Kunci: -" />;
      return (
        <Alert
          type="success"
          showIcon
          className="mt-3"
          message="Kunci pasangan:"
          description={
          <ul className="mt-1 space-y-1">
            {pas.map((p, i) => (
              <li key={i} className="flex items-center gap-2">
                <span className="min-w-0 flex-1"><RichContent html={p.kiri} /></span>
                <span aria-hidden>→</span>
                <span className="min-w-0 flex-1"><RichContent html={p.kanan} /></span>
              </li>
            ))}
          </ul>
          }
        />
      );
    }
    if (soal.tipe === "uraian" || soal.tipe === "upload_file" || soal.tipe === "praktik_singkat") {
      return (
        <Alert
          type="success"
          showIcon
          className="mt-3"
          message="Rubrik:"
          description={<RichContent html={String(kunci.rubrik ?? "-")} />}
        />
      );
    }
    return null;
  };

  return (
    <article className="rounded-lg border border-border bg-white p-4 sm:p-5">
      <div className="flex flex-wrap items-center gap-2">
        {info && (
          <Tag color={info.kelompok === "inti" ? "blue" : "purple"}>[{badgeKelompok(info.kelompok)}]</Tag>
        )}
        <Tag>{info?.label ?? soal.tipe}</Tag>
        <Tag>Bobot {soal.bobot}</Tag>
        <Tag>{soal.level}</Tag>
        {info && !info.dinilai && (
          <Tag color="gold" icon={<AlertTriangle size={12} />}>
            Tidak dinilai
          </Tag>
        )}
      </div>

      {soal.stimulus && (
        <div className="mt-3 rounded-lg border border-border bg-surface px-4 py-3 text-[16px] leading-relaxed text-ink">
          <RichContent html={soal.stimulus} />
        </div>
      )}
      <div className="mt-3 text-[16px] leading-relaxed text-ink">
        <RichContent html={soal.pertanyaan} />
      </div>
      {soal.media_url && <MediaSoal url={soal.media_url} />}
      <div className="mt-3">{renderJawaban()}</div>
      {showKunci && kunciRich()}
      {showKunci && kunciText() && (
        <Alert type="success" showIcon className="mt-3" message={kunciText()} />
      )}
      {soal.pembahasan && (
        <Alert
          type="info"
          showIcon
          className="mt-3"
          message="Pembahasan"
          description={<RichContent html={soal.pembahasan} />}
        />
      )}
    </article>
  );
}
