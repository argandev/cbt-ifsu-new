// ImportSoalPanel — panel import bank soal Excel (.xlsx) / Word (.docx).
//
// Alur (spesifikasi docs/ANALISIS-IMPORT-SOAL.md Bab 2.8 + brief Tim Import):
//   1. Unduh template  -> GET /api/v1/bank-soal/import/template.xlsx
//                        -> GET /api/v1/bank-soal/import/template.docx
//   2. Unggah file     -> POST /api/v1/bank-soal/import/preview
//      (multipart: file + bank_id; dry-run, TIDAK menulis ke database)
//   3. Konfirmasi      -> POST /api/v1/bank-soal/import/xlsx
//      (multipart: file yang sama + bank_id + mode + duplikat)
//
// ASUMSI ENDPOINT (backend mungkin belum selesai saat panel ini ditulis):
// - Endpoint template memakai suffix .xlsx / .docx sesuai brief Tim Import
//   (spesifikasi Bab 2.8 menulis satu endpoint /template — brief yang berlaku).
// - Endpoint konfirmasi TUNGGAL /import/xlsx menerima file Excel maupun Word
//   (sesuai brief); bila backend memisahnya ke /import/docx, ubah konstanta
//   CONFIRM_ENDPOINT di bawah.
// - Respons JSON mengikuti amplop backend { data: {...} } atau objek langsung;
//   parser di bawah menerima keduanya + nama field hitungan alternatif.
// - Route panduan "/adm/bank-soal/panduan-import" milik Tim Panduan (App.tsx)
//   — SUDAH LIVE (dikonfirmasi Tim Panduan), tombol "Lihat Panduan" menaut ke sana.
//
// Upload memakai fetch + FormData langsung (BUKAN apiClient.post): apiClient
// memaksa header Content-Type application/json + JSON.stringify body, yang
// merusak boundary multipart. CSRF diambil dari cookie `cbt_csrf` dengan cara
// yang sama seperti apiClient. Format CSV TIDAK lagi didukung di panel ini.
//
// Hanya presentasi + pemanggilan endpoint: tidak mengubah logika grading/auth.
//
// CHROME REDESIGN ANT DESIGN (Tim Form): pemilih format → antd Radio,
// zona unggah → antd Upload.Dragger (beforeUpload memanggil pilihFile +
// return false — file TIDAK auto-upload, pratinjau tetap via fetch manual),
// mode/duplikat → antd Select, pratinjau → antd Table + Statistic,
// langkah → antd Steps, pesan error/warning → antd Alert, tombol → antd
// Button. Endpoint, payload, urutan alur, dan parsing respons TETAP sama.
import { useMemo, useRef, useState } from "react";
import {
  Alert,
  Button,
  Radio,
  Select,
  Statistic,
  Steps,
  Table,
  Tag,
  Typography,
  Upload,
} from "antd";
import type { TableColumnsType } from "antd";
import {
  BookOpen,
  CheckCircle2,
  Download,
  Eye,
  FileSpreadsheet,
  FileText,
  Inbox,
  RotateCcw,
  X,
} from "lucide-react";
import { TIPE_SOAL } from "../../../lib/tipeSoal";

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "";
const MAKS_UKURAN = 5 * 1024 * 1024; // 5 MB (keputusan user).
const TEMPLATE_XLSX_URL = "/api/v1/bank-soal/import/template.xlsx";
const TEMPLATE_DOCX_URL = "/api/v1/bank-soal/import/template.docx";
const PREVIEW_URL = "/api/v1/bank-soal/import/preview";
const CONFIRM_ENDPOINT = "/api/v1/bank-soal/import/xlsx"; // Lihat ASUMSI di kepala file.
const PANDUAN_URL = "/adm/bank-soal/panduan-import";

type FormatFile = "xlsx" | "docx";
type ModeImport = "all_or_nothing" | "skip_errors";
type ModeDuplikat = "lewati" | "simpan_ganda";

interface BarisPesan {
  baris: number;
  pesan: string;
}

interface DuplikatDetail {
  baris: number;
  baris_duplikat?: number;
  soal_id?: string;
  pesan: string;
}

interface PreviewRow {
  baris: number;
  tipe: string;
  tipe_label?: string;
  kelompok?: string;
  pertanyaan: string;
  stimulus?: string;
  opsi?: string[] | Record<string, string> | string;
  // PERHATIAN: backend mengirim kunci_terpetakan dalam BENTUK BERAGAM
  // (bukan selalu string): pg -> {label, teks}; pg_kompleks/pg_multi ->
  // [{label, teks}]; benar_salah -> [{pernyataan, nilai}]; menjodohkan/
  // peta_konsep -> {pasangan: [...]}; isi_ganda -> {kolom, jawaban};
  // urutkan -> {urutan_label, urutan_teks}; uraian/rubrik -> string.
  // JANGAN render langsung — React crash "Objects are not valid as a
  // React child" (temuan QA 16 Okt 2026). Selalu lewat teksKunci().
  kunci_terpetakan?: unknown;
  kunci?: unknown;
  bobot?: number;
  level?: string;
  pembahasan?: string;
  media_url?: string;
  media_alt?: string;
  kisi_kisi?: string;
  kompetensi?: string;
  // Backend mengirim warnings per-baris sebagai OBJEK {baris, kode, pesan}
  // (bukan string). Jangan render langsung — lewat teksPesan() di bawah
  // (temuan QA 16 Okt 2026: render objek -> crash error boundary).
  warnings?: unknown[];
}

interface PreviewResult {
  total_baris: number;
  valid: number;
  jumlah_error: number;
  jumlah_warning: number;
  jumlah_duplikat: number;
  ringkasan_tipe?: Record<string, number>;
  total_bobot_valid?: number;
  errors: BarisPesan[];
  warnings: BarisPesan[];
  duplikat_detail: DuplikatDetail[];
  preview: PreviewRow[];
}

interface ConfirmResult {
  sukses: number;
  gagal: number;
  dilewati_duplikat: number;
  dilewati_error: number;
  warnings: BarisPesan[];
  errors: BarisPesan[];
  disimpan: boolean;
}

interface ImportSoalPanelProps {
  bankId: string;
  terkunci?: boolean;
  onImported: () => void;
  onTutup: () => void;
}

function getCsrfToken(): string {
  const match = document.cookie.match(/(?:^|;\s*)cbt_csrf=([^;]*)/);
  return match ? decodeURIComponent(match[1]) : "";
}

/** Ambil isi amplop { data } bila ada, atau objek langsung. */
function unwrap<T>(json: unknown): T {
  if (typeof json === "object" && json !== null && "data" in json) {
    return (json as { data: T }).data;
  }
  return json as T;
}

function pesanErrorApi(json: unknown, status: number): string {
  if (typeof json === "object" && json !== null && "error" in json) {
    const err = (json as { error?: { message?: string } }).error;
    if (err?.message) return err.message;
  }
  return `Permintaan gagal (${status}).`;
}

function angka(...nilai: unknown[]): number {
  for (const n of nilai) {
    if (typeof n === "number" && Number.isFinite(n)) return n;
  }
  return 0;
}

function daftar<T>(nilai: unknown): T[] {
  return Array.isArray(nilai) ? (nilai as T[]) : [];
}

function parsePreview(json: unknown): PreviewResult {
  const raw = unwrap<Record<string, unknown>>(json);
  const errors = daftar<BarisPesan>(raw.errors);
  const warnings = daftar<BarisPesan>(raw.warnings);
  const duplikatDetail = daftar<DuplikatDetail>(raw.duplikat_detail ?? raw.duplikat);
  return {
    total_baris: angka(raw.total_baris, raw.total, raw.preview ? daftar(raw.preview).length : 0),
    valid: angka(raw.valid, raw.valid_count, raw.jumlah_valid),
    jumlah_error: angka(raw.error, raw.jumlah_error, raw.error_count, errors.length),
    jumlah_warning: angka(raw.warning, raw.jumlah_warning, raw.warning_count, warnings.length),
    jumlah_duplikat: angka(raw.duplikat, raw.jumlah_duplikat, raw.duplikat_count, duplikatDetail.length),
    ringkasan_tipe: (raw.ringkasan_tipe as Record<string, number> | undefined) ?? undefined,
    total_bobot_valid: typeof raw.total_bobot_valid === "number" ? raw.total_bobot_valid : undefined,
    errors,
    warnings,
    duplikat_detail: duplikatDetail,
    preview: daftar<PreviewRow>(raw.preview ?? raw.rows ?? raw.soal),
  };
}

function parseConfirm(json: unknown): ConfirmResult {
  const raw = unwrap<Record<string, unknown>>(json);
  return {
    sukses: angka(raw.sukses, raw.tersimpan, raw.disimpan_count, 0),
    gagal: angka(raw.gagal, 0),
    dilewati_duplikat: angka(raw.dilewati_duplikat, 0),
    dilewati_error: angka(raw.dilewati_error, raw.dilewati_karena_error, 0),
    warnings: daftar<BarisPesan>(raw.warnings),
    errors: daftar<BarisPesan>(raw.errors),
    disimpan: typeof raw.disimpan === "boolean" ? raw.disimpan : angka(raw.sukses, 0) > 0,
  };
}

async function postMultipart<T>(path: string, form: FormData, parse: (json: unknown) => T): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    method: "POST",
    body: form,
    credentials: "include",
    headers: { "X-CSRF-Token": getCsrfToken() },
  });
  const text = await res.text();
  let json: unknown = null;
  try {
    json = text ? JSON.parse(text) : null;
  } catch {
    // Respons bukan JSON (mis. halaman error) -> pesan generik di bawah.
  }
  if (!res.ok) throw new Error(pesanErrorApi(json, res.status));
  return parse(json);
}

function namaFileDariHeader(header: string | null, fallback: string): string {
  if (!header) return fallback;
  const match = header.match(/filename\*?=(?:UTF-8''|")?([^";\n]+)/i);
  if (!match) return fallback;
  try {
    return decodeURIComponent(match[1].replace(/"$/, ""));
  } catch {
    return match[1].replace(/"$/, "");
  }
}

async function unduhTemplate(path: string, fallback: string): Promise<void> {
  const res = await fetch(`${BASE_URL}${path}`, { method: "GET", credentials: "include" });
  if (!res.ok) {
    let json: unknown = null;
    try {
      json = await res.clone().json();
    } catch {
      // Bukan JSON — pakai pesan generik.
    }
    throw new Error(pesanErrorApi(json, res.status));
  }
  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = namaFileDariHeader(res.headers.get("Content-Disposition"), fallback);
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}

function formatUkuran(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
}

function validasiFile(file: File, format: FormatFile): string | null {
  const ext = file.name.split(".").pop()?.toLowerCase() ?? "";
  if (ext !== format) {
    return `File harus berekstensi .${format} untuk format ${format === "xlsx" ? "Excel" : "Word"}. File yang dipilih: ${file.name}.`;
  }
  if (file.size === 0) return "File kosong — pilih file yang berisi soal.";
  if (file.size > MAKS_UKURAN) return "File melebihi batas 5 MB.";
  return null;
}

function teksOpsi(opsi: PreviewRow["opsi"] | unknown): string {
  if (!opsi) return "—";
  if (typeof opsi === "string") return opsi;
  if (Array.isArray(opsi)) {
    return (
      opsi
        .map((item) => {
          if (item !== null && typeof item === "object") {
            const o = item as Record<string, unknown>;
            if (typeof o.label === "string" && typeof o.teks === "string") return `${o.label}. ${o.teks}`;
            if (typeof o.kiri === "string" || typeof o.kanan === "string")
              return `${String(o.kiri ?? "")}::${String(o.kanan ?? "")}`;
            return JSON.stringify(item);
          }
          return String(item);
        })
        .join(" | ") || "—"
    );
  }
  if (typeof opsi === "object") {
    const o = opsi as Record<string, unknown>;
    // Bentuk backend: {opsi: [{label, teks}]}, {pasangan: [...]},
    // {pernyataan: [...]}, {item: [...]}, {kolom: [...]}.
    if (Array.isArray(o.opsi)) return teksOpsi(o.opsi);
    if (Array.isArray(o.pasangan)) return teksOpsi(o.pasangan);
    if (Array.isArray(o.pernyataan)) return o.pernyataan.map((v) => String(v)).join(" | ");
    if (Array.isArray(o.item)) return o.item.map((v) => String(v)).join(" | ");
    if (Array.isArray(o.kolom)) return o.kolom.map((v) => String(v)).join(" | ");
    return (
      Object.entries(o)
        .map(([k, v]) => `${k}. ${typeof v === "object" ? JSON.stringify(v) : String(v)}`)
        .join(" | ") || "—"
    );
  }
  return String(opsi);
}

/**
 * Format kunci_terpetakan backend (bentuk beragam, lihat PreviewRow)
 * menjadi SATU STRING untuk ditampilkan. Fallback JSON agar tidak
 * pernah melempar objek mentah ke React (crash error boundary).
 */
function teksKunci(nilai: unknown): string {
  if (nilai === null || nilai === undefined) return "—";
  if (typeof nilai === "string") return nilai === "" ? "—" : nilai;
  if (typeof nilai === "number" || typeof nilai === "boolean") return String(nilai);
  if (Array.isArray(nilai)) {
    if (nilai.length === 0) return "—";
    return (
      nilai
        .map((item) => {
          if (item !== null && typeof item === "object") {
            const o = item as Record<string, unknown>;
            if (typeof o.label === "string" && typeof o.teks === "string") return `${o.label} — ${o.teks}`;
            if (typeof o.pernyataan === "string") return `${o.pernyataan}: ${String(o.nilai ?? "")}`;
            if (typeof o.kiri === "string" || typeof o.kanan === "string")
              return `${String(o.kiri ?? "")}::${String(o.kanan ?? "")}`;
            return JSON.stringify(item);
          }
          return String(item);
        })
        .join("; ") || "—"
    );
  }
  if (typeof nilai === "object") {
    const o = nilai as Record<string, unknown>;
    if (typeof o.label === "string" && typeof o.teks === "string") return `${o.label} — ${o.teks}`;
    if (Array.isArray(o.pasangan)) return teksKunci(o.pasangan);
    if (Array.isArray(o.kolom) && Array.isArray(o.jawaban))
      return o.kolom.map((k, i) => `${String(k)} = ${String((o.jawaban as unknown[])[i] ?? "")}`).join("; ");
    if (Array.isArray(o.urutan_label) && Array.isArray(o.urutan_teks))
      return (o.urutan_label as unknown[]).map((l, i) => `${String(l)} (${String((o.urutan_teks as unknown[])[i] ?? "")})`).join(" → ");
    if (Array.isArray(o.jawaban)) return (o.jawaban as unknown[]).map((j) => String(j)).join(" | ");
    if (typeof o.rubrik === "string") return o.rubrik === "" ? "—" : o.rubrik;
    try {
      return JSON.stringify(nilai);
    } catch {
      return "—";
    }
  }
  return "—";
}

function labelKelompok(kelompok: string | undefined): string {
  const k = (kelompok ?? "").toLowerCase();
  if (k.includes("inti")) return "Inti AKM";
  if (k.includes("varian")) return "Varian aplikasi";
  return kelompok ?? "—";
}

function labelTipe(row: PreviewRow): string {
  return row.tipe_label ?? TIPE_SOAL[row.tipe]?.label ?? row.tipe;
}

/** Ubah satu entri warning/error (string ATAU objek {pesan}) jadi string. */
function teksPesan(nilai: unknown): string {
  if (typeof nilai === "string") return nilai;
  if (nilai !== null && typeof nilai === "object") {
    const o = nilai as Record<string, unknown>;
    if (typeof o.pesan === "string") return o.pesan;
    try {
      return JSON.stringify(nilai);
    } catch {
      return "—";
    }
  }
  return String(nilai ?? "—");
}

export default function ImportSoalPanel({ bankId, terkunci, onImported, onTutup }: ImportSoalPanelProps) {
  const [format, setFormat] = useState<FormatFile>("xlsx");
  const [file, setFile] = useState<File | null>(null);
  const [fileError, setFileError] = useState("");
  const [preview, setPreview] = useState<PreviewResult | null>(null);
  const [mode, setMode] = useState<ModeImport>("all_or_nothing");
  const [duplikat, setDuplikat] = useState<ModeDuplikat>("lewati");
  const [hasil, setHasil] = useState<ConfirmResult | null>(null);
  const [pesan, setPesan] = useState("");
  const [mengunduhTemplate, setMengunduhTemplate] = useState<FormatFile | null>(null);
  const [mempratinjau, setMempratinjau] = useState(false);
  const [menyimpan, setMenyimpan] = useState(false);
  // QA Ant Design: anchor ke seksi unggah agar tombol "Perbaiki file &
  // unggah ulang" bisa mengembalikan fokus pengguna ke Dragger.
  const unggahRef = useRef<HTMLDivElement>(null);

  const barisError = useMemo(() => new Set((preview?.errors ?? []).map((e) => e.baris)), [preview]);
  const errorPerBaris = useMemo(() => {
    const map = new Map<number, string[]>();
    for (const e of preview?.errors ?? []) {
      map.set(e.baris, [...(map.get(e.baris) ?? []), e.pesan]);
    }
    return map;
  }, [preview]);
  const warningPerBaris = useMemo(() => {
    const map = new Map<number, string[]>();
    for (const w of preview?.warnings ?? []) {
      map.set(w.baris, [...(map.get(w.baris) ?? []), w.pesan]);
    }
    for (const row of preview?.preview ?? []) {
      for (const w of row.warnings ?? []) {
        map.set(row.baris, [...(map.get(row.baris) ?? []), teksPesan(w)]);
      }
    }
    return map;
  }, [preview]);

  const konfirmasiDiblokir =
    !!preview && mode === "all_or_nothing" && preview.jumlah_error > 0;

  // Langkah wizard (indikator presentasi saja — alur tetap sama):
  // 0 Template → 1 Unggah file → 2 Pratinjau → 3 Simpan.
  const langkahAktif = hasil ? 3 : preview ? 2 : file ? 1 : 0;

  const kolomPratinjau: TableColumnsType<PreviewRow> = [
    { title: "Baris", dataIndex: "baris", key: "baris", width: 72, fixed: "left" },
    {
      title: "Tipe",
      key: "tipe",
      width: 160,
      render: (_, row) => (
        <span>
          <span className="block text-ink">{labelTipe(row)}</span>
          <span className="block text-xs text-muted">{labelKelompok(row.kelompok)}</span>
          {row.tipe === "pernyataan_sikap" && <Tag className="mt-1">Tidak dinilai</Tag>}
        </span>
      ),
    },
    {
      title: "Pertanyaan",
      key: "pertanyaan",
      width: 280,
      render: (_, row) => (
        <span>
          <span className="line-clamp-3 block text-ink" title={row.pertanyaan}>
            {row.pertanyaan}
          </span>
          {row.media_url && (
            <a
              href={row.media_url}
              target="_blank"
              rel="noreferrer"
              className="mt-1 block break-all text-xs text-anbk underline"
            >
              Gambar: {row.media_url}
            </a>
          )}
          {row.kisi_kisi && <span className="block text-xs text-muted">Kisi-kisi: {row.kisi_kisi}</span>}
          {row.kompetensi && (
            <span className="block text-xs text-muted">Kompetensi: {row.kompetensi}</span>
          )}
        </span>
      ),
    },
    {
      title: "Opsi",
      key: "opsi",
      width: 220,
      render: (_, row) => (
        <span className="line-clamp-3 block text-ink" title={teksOpsi(row.opsi)}>
          {teksOpsi(row.opsi)}
        </span>
      ),
    },
    {
      title: "Kunci terpetakan",
      key: "kunci",
      width: 180,
      render: (_, row) => (
        <span
          className="line-clamp-3 block text-ink"
          title={teksKunci(row.kunci_terpetakan ?? row.kunci)}
        >
          {teksKunci(row.kunci_terpetakan ?? row.kunci)}
        </span>
      ),
    },
    { title: "Bobot", dataIndex: "bobot", key: "bobot", width: 72, render: (v) => v ?? "—" },
    { title: "Level", dataIndex: "level", key: "level", width: 88, render: (v) => v ?? "—" },
    {
      title: "Catatan",
      key: "catatan",
      width: 220,
      render: (_, row) => {
        const errBaris = errorPerBaris.get(row.baris) ?? [];
        const warnBaris = warningPerBaris.get(row.baris) ?? [];
        if (errBaris.length === 0 && warnBaris.length === 0) {
          return <span className="text-xs text-muted">—</span>;
        }
        return (
          <span>
            {errBaris.map((m, i) => (
              <p key={`e-${i}`} className="text-xs text-danger">{m}</p>
            ))}
            {warnBaris.map((m, i) => (
              <p key={`w-${i}`} className="text-xs text-warning">{m}</p>
            ))}
          </span>
        );
      },
    },
  ];

  const resetHasil = () => {
    setPreview(null);
    setHasil(null);
    setPesan("");
  };

  const pilihFormat = (f: FormatFile) => {
    setFormat(f);
    setFile(null);
    setFileError("");
    resetHasil();
  };

  const pilihFile = (f: File | undefined | null) => {
    if (!f) return;
    const err = validasiFile(f, format);
    if (err) {
      setFile(null);
      setFileError(err);
      resetHasil();
      return;
    }
    setFile(f);
    setFileError("");
    resetHasil();
  };

  const handlePreview = async () => {
    if (!file || !bankId) return;
    setMempratinjau(true);
    setPesan("");
    setHasil(null);
    try {
      const form = new FormData();
      form.append("file", file);
      form.append("bank_id", bankId);
      const data = await postMultipart(PREVIEW_URL, form, parsePreview);
      setPreview(data);
    } catch (err) {
      setPesan(err instanceof Error ? err.message : "Pratinjau gagal.");
    } finally {
      setMempratinjau(false);
    }
  };

  const handleKonfirmasi = async () => {
    if (!file || !bankId || !preview) return;
    setMenyimpan(true);
    setPesan("");
    try {
      const form = new FormData();
      form.append("file", file);
      form.append("bank_id", bankId);
      form.append("mode", mode);
      form.append("duplikat", duplikat);
      const data = await postMultipart(CONFIRM_ENDPOINT, form, parseConfirm);
      setHasil(data);
      if (data.disimpan) onImported();
    } catch (err) {
      setPesan(err instanceof Error ? err.message : "Konfirmasi simpan gagal.");
    } finally {
      setMenyimpan(false);
    }
  };

  const handleUnduhTemplate = async (f: FormatFile) => {
    setMengunduhTemplate(f);
    setPesan("");
    try {
      if (f === "xlsx") {
        await unduhTemplate(TEMPLATE_XLSX_URL, "template-import-soal-cbt-smk-v2.xlsx");
      } else {
        await unduhTemplate(TEMPLATE_DOCX_URL, "template-import-soal-cbt-smk-v2.docx");
      }
    } catch (err) {
      setPesan(err instanceof Error ? err.message : "Unduh template gagal.");
    } finally {
      setMengunduhTemplate(null);
    }
  };

  if (!bankId) {
    return <p className="text-sm text-danger">Bank soal tidak ditemukan.</p>;
  }

  return (
    <div className="space-y-4 rounded-lg border border-border bg-surface p-3 sm:p-4">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <Typography.Title level={4} style={{ margin: 0 }}>Import Soal — Excel / Word</Typography.Title>
          <p className="mt-1 text-sm text-muted">
            Unduh template, isi, unggah, periksa pratinjau, lalu konfirmasi simpan. Soal hasil import berstatus
            Draft dan wajib ditelaah. Maks 500 baris soal per file, maks 5 MB. Format CSV tidak lagi didukung —
            gunakan Excel (.xlsx) atau Word (.docx).
          </p>
        </div>
        <Button
          type="text"
          onClick={onTutup}
          aria-label="Tutup panel import"
          icon={<X size={20} />}
        />
      </div>

      <Steps
        current={langkahAktif}
        items={[
          { title: "Template" },
          { title: "Unggah file" },
          { title: "Pratinjau" },
          { title: "Simpan" },
        ]}
      />

      {terkunci && (
        <Alert
          type="warning"
          showIcon
          message="Bank soal terkunci — import ditolak sampai bank dibuka lagi."
          role="alert"
        />
      )}

      {pesan && (
        <Alert type="error" showIcon message={pesan} role="alert" />
      )}

      {/* 1. Pemilih format + template + panduan */}
      <div className="space-y-3 rounded-lg border border-border bg-white p-3">
        <p className="text-sm font-medium text-ink">1. Pilih format &amp; unduh template</p>
        <Radio.Group
          value={format}
          onChange={(e) => pilihFormat(e.target.value as FormatFile)}
          className="flex flex-col gap-2 sm:flex-row"
        >
          <Radio.Button value="xlsx" className="flex-1">
            <span className="inline-flex items-center gap-2">
              <FileSpreadsheet size={18} className="shrink-0 text-success" aria-hidden />
              <span>
                <span className="block text-sm font-semibold text-ink">Excel (.xlsx)</span>
                <span className="block text-xs text-muted">Satu baris = satu soal, sheet &ldquo;Soal&rdquo;.</span>
              </span>
            </span>
          </Radio.Button>
          <Radio.Button value="docx" className="flex-1">
            <span className="inline-flex items-center gap-2">
              <FileText size={18} className="shrink-0 text-primary" aria-hidden />
              <span>
                <span className="block text-sm font-semibold text-ink">Word (.docx)</span>
                <span className="block text-xs text-muted">Satu soal per blok, label eksplisit di template.</span>
              </span>
            </span>
          </Radio.Button>
        </Radio.Group>
        <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap">
          <Button
            type="default"
            disabled={mengunduhTemplate !== null}
            loading={mengunduhTemplate === "xlsx"}
            onClick={() => void handleUnduhTemplate("xlsx")}
            icon={<Download size={16} aria-hidden />}
          >
            Download Template Excel
          </Button>
          <Button
            type="default"
            disabled={mengunduhTemplate !== null}
            loading={mengunduhTemplate === "docx"}
            onClick={() => void handleUnduhTemplate("docx")}
            icon={<Download size={16} aria-hidden />}
          >
            Download Template Word
          </Button>
          <Button href={PANDUAN_URL} icon={<BookOpen size={16} aria-hidden />}>
            Lihat Panduan
          </Button>
        </div>
        <p className="text-xs text-muted">
          Isi hanya sesuai template. Jangan mengganti nama sheet (Excel) / label soal (Word). Kunci ditulis sesuai
          tipe — lihat panduan bila ragu.
        </p>
      </div>

      {/* 2. Upload file */}
      <div ref={unggahRef} className="space-y-3 rounded-lg border border-border bg-white p-3">
        <p className="text-sm font-medium text-ink">2. Unggah file {format === "xlsx" ? "Excel (.xlsx)" : "Word (.docx)"}</p>
        <Upload.Dragger
          accept={format === "xlsx" ? ".xlsx" : ".docx"}
          multiple={false}
          maxCount={1}
          showUploadList={false}
          // File TIDAK diunggah otomatis: beforeUpload hanya memilih file
          // (validasi) lalu return false; pratinjau tetap via fetch manual.
          beforeUpload={(f) => {
            pilihFile(f);
            return false;
          }}
        >
          <p className="ant-upload-drag-icon">
            <Inbox />
          </p>
          <p className="ant-upload-text">Seret file ke sini atau klik untuk pilih file</p>
          <p className="ant-upload-hint">Hanya .{format} · maks 5 MB · maks 500 soal</p>
        </Upload.Dragger>
        {file && (
          <p className="flex flex-wrap items-center gap-2 text-sm text-ink">
            <CheckCircle2 size={16} className="shrink-0 text-success" aria-hidden />
            File dipilih: <strong className="break-all">{file.name}</strong> · {formatUkuran(file.size)}
            <Button
              type="link"
              danger
              onClick={() => {
                setFile(null);
                setFileError("");
                resetHasil();
              }}
            >
              Ganti
            </Button>
          </p>
        )}
        {fileError && (
          <Alert type="error" showIcon message={fileError} role="alert" />
        )}
        <div>
          <Button
            type="primary"
            disabled={!file || !!fileError || mempratinjau || terkunci}
            loading={mempratinjau}
            onClick={() => void handlePreview()}
            icon={<Eye size={16} aria-hidden />}
          >
            {mempratinjau ? "Memuat pratinjau…" : "Preview"}
          </Button>
        </div>
      </div>

      {/* 3. Hasil pratinjau */}
      {preview && !hasil && (
        <div className="space-y-3 rounded-lg border border-border bg-white p-3">
          <Typography.Text strong className="block">3. Pratinjau hasil parsing</Typography.Text>
          <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-6">
            <Statistic title="Total baris" value={preview.total_baris} />
            <Statistic title="Valid" value={preview.valid} />
            <Statistic title="Error" value={preview.jumlah_error} />
            <Statistic title="Warning" value={preview.jumlah_warning} />
            <Statistic title="Duplikat" value={preview.jumlah_duplikat} />
            <Statistic title="Total bobot valid" value={preview.total_bobot_valid ?? "—"} />
          </div>
          {preview.ringkasan_tipe && Object.keys(preview.ringkasan_tipe).length > 0 && (
            <p className="text-sm text-ink">
              Ringkasan tipe:{" "}
              {Object.entries(preview.ringkasan_tipe)
                .map(([kode, n]) => `${TIPE_SOAL[kode]?.label ?? kode}: ${n}`)
                .join(" · ")}
            </p>
          )}

          {preview.errors.length > 0 && (
            <Alert
              type="error"
              showIcon
              message={`Daftar error (${preview.errors.length})`}
              description={
                <ul className="list-disc space-y-1 pl-5">
                  {preview.errors.map((er, i) => (
                    <li key={`${er.baris}-${i}`}>
                      Baris {er.baris}: {er.pesan}
                    </li>
                  ))}
                </ul>
              }
            />
          )}
          {preview.warnings.length > 0 && (
            <Alert
              type="warning"
              showIcon
              message={`Daftar warning (${preview.warnings.length})`}
              description={
                <ul className="list-disc space-y-1 pl-5">
                  {preview.warnings.map((w, i) => (
                    <li key={`${w.baris}-${i}`}>
                      Baris {w.baris}: {w.pesan}
                    </li>
                  ))}
                </ul>
              }
            />
          )}
          {preview.duplikat_detail.length > 0 && (
            <Alert
              type="warning"
              showIcon
              message={`Duplikat terdeteksi (${preview.duplikat_detail.length})`}
              description={
                <ul className="list-disc space-y-1 pl-5">
                  {preview.duplikat_detail.map((d, i) => (
                    <li key={`${d.baris}-${i}`}>
                      Baris {d.baris}: {d.pesan}
                      {d.baris_duplikat ? ` (duplikat baris ${d.baris_duplikat})` : ""}
                    </li>
                  ))}
                </ul>
              }
            />
          )}

          {preview.preview.length > 0 && (
            <Table<PreviewRow>
              columns={kolomPratinjau}
              dataSource={preview.preview}
              rowKey="baris"
              rowClassName={(row) => (barisError.has(row.baris) ? "bg-danger-bg" : "")}
              pagination={{
                pageSize: 10,
                showSizeChanger: false,
                showTotal: (total) => `Total ${total} soal`,
              }}
              scroll={{ x: 1000 }}
              locale={{ emptyText: "Belum ada soal pada pratinjau." }}
            />
          )}

          {/* 4. Pilihan mode + duplikat + konfirmasi */}
          <div className="space-y-3 border-t border-border pt-3">
            <Typography.Text strong className="block">4. Konfirmasi simpan</Typography.Text>
            <div className="flex flex-col gap-3 sm:flex-row">
              <div className="flex-1">
                <label className="block text-sm font-semibold text-ink" htmlFor="mode-import">
                  Mode bila ada error
                </label>
                <Select
                  id="mode-import"
                  value={mode}
                  onChange={(v) => setMode(v as ModeImport)}
                  options={[
                    { value: "all_or_nothing", label: "Tolak semua bila ada error (default)" },
                    { value: "skip_errors", label: "Lewati baris yang error" },
                  ]}
                  style={{ width: "100%" }}
                  size="large"
                  notFoundContent="Tidak ada data"
                />
              </div>
              <div className="flex-1">
                <label className="block text-sm font-semibold text-ink" htmlFor="perlakuan-duplikat">
                  Perlakuan duplikat
                </label>
                <Select
                  id="perlakuan-duplikat"
                  value={duplikat}
                  onChange={(v) => setDuplikat(v as ModeDuplikat)}
                  options={[
                    { value: "lewati", label: "Lewati duplikat (default)" },
                    { value: "simpan_ganda", label: "Simpan ganda" },
                  ]}
                  style={{ width: "100%" }}
                  size="large"
                  notFoundContent="Tidak ada data"
                />
              </div>
            </div>
            {konfirmasiDiblokir && (
              <Alert
                type="warning"
                showIcon
                message="Mode “Tolak semua” tidak bisa menyimpan selama masih ada error. Perbaiki file lalu unggah ulang, atau pilih mode “Lewati baris yang error”."
                role="alert"
              />
            )}
            <div className="flex flex-col gap-2 sm:flex-row">
              <Button
                type="primary"
                disabled={!file || menyimpan || konfirmasiDiblokir || terkunci}
                loading={menyimpan}
                onClick={() => void handleKonfirmasi()}
                icon={<CheckCircle2 size={16} aria-hidden />}
              >
                {menyimpan ? "Menyimpan…" : "Konfirmasi Simpan"}
              </Button>
              <Button
                type="default"
                onClick={() => {
                  // QA Ant Design: sebelumnya hanya preview yang dibersihkan,
                  // file LAMA tetap terpilih — klik "Preview" lagi akan
                  // mempratinjau file lama yang belum diperbaiki. Sekarang
                  // file ikut dibuang agar pengguna wajib memilih file yang
                  // sudah diperbaiki lewat Dragger (yang kembali terlihat
                  // setelah preview hilang). Dragger TIDAK diklik otomatis:
                  // klik programatis pada input tersembunyi antd Upload
                  // rapuh antar-browser + mengejutkan; pengguna klik
                  // Dragger sekali — seksi unggah di-scroll ke tampilan.
                  setPreview(null);
                  setPesan("");
                  setFile(null);
                  setFileError("");
                  window.setTimeout(() => {
                    unggahRef.current?.scrollIntoView({ behavior: "smooth", block: "center" });
                  }, 50);
                }}
                icon={<RotateCcw size={16} aria-hidden />}
              >
                Perbaiki file &amp; unggah ulang
              </Button>
              <Button type="text" onClick={onTutup}>
                Batal
              </Button>
            </div>
            <p className="text-xs text-muted">
              Konfirmasi mengunggah ulang file yang sama — server memvalidasi ulang sebelum menyimpan, agar hasil
              simpan selalu mencerminkan file yang benar-benar disimpan.
            </p>
          </div>
        </div>
      )}

      {/* 5. Hasil konfirmasi */}
      {hasil && (
        <div className="space-y-3 rounded-lg border border-border bg-white p-3">
          <Typography.Text strong className="block">Hasil import</Typography.Text>
          <p className="text-sm text-ink">
            Disimpan: <strong>{hasil.sukses}</strong> · Gagal: <strong>{hasil.gagal}</strong> · Dilewati karena
            duplikat: <strong>{hasil.dilewati_duplikat}</strong>
            {hasil.dilewati_error > 0 && (
              <>
                {" "}· Dilewati karena error: <strong>{hasil.dilewati_error}</strong>
              </>
            )}{" "}
            · Status: {hasil.disimpan ? "tersimpan" : "tidak ada yang tersimpan"}
          </p>
          {hasil.errors.length > 0 && (
            <Alert
              type="error"
              showIcon
              message={`Error import (${hasil.errors.length})`}
              description={
                <ul className="list-disc space-y-1 pl-5">
                  {hasil.errors.map((er, i) => (
                    <li key={`h-e-${i}`}>
                      Baris {er.baris}: {er.pesan}
                    </li>
                  ))}
                </ul>
              }
            />
          )}
          {hasil.warnings.length > 0 && (
            <Alert
              type="warning"
              showIcon
              message={`Warning import (${hasil.warnings.length})`}
              description={
                <ul className="list-disc space-y-1 pl-5">
                  {hasil.warnings.map((w, i) => (
                    <li key={`h-w-${i}`}>
                      Baris {w.baris}: {w.pesan}
                    </li>
                  ))}
                </ul>
              }
            />
          )}
          {hasil.sukses === 0 && hasil.errors.length === 0 && (
            <p className="text-sm text-muted">Tidak ada soal yang tersimpan.</p>
          )}
          <div className="flex flex-col gap-2 sm:flex-row">
            <Button href="/adm/bank-soal" type="primary">
              Kembali ke bank soal
            </Button>
            <Button
              type="default"
              onClick={() => {
                setHasil(null);
                setPreview(null);
                setFile(null);
                setFileError("");
              }}
            >
              Import file lain
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
