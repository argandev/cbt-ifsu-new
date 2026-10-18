import { useState } from "react";
import { Eye, EyeOff, Plus, Trash2 } from "lucide-react";
import { Alert, Button, Checkbox, Input, InputNumber, Modal, Radio, Select, Tag } from "antd";
import type { Soal, SoalFormValue } from "../../lib/banksoal";
import { TIPE_LIST, TIPE_SOAL, badgeKelompok } from "../../lib/tipeSoal";
import ImageUpload from "../ui/ImageUpload";
import RichEditor from "./RichEditor";
import SoalRenderer from "./SoalRenderer";
import { stripHtmlToText } from "./RichContent";

const LABELS = ["A", "B", "C", "D", "E", "F", "G", "H"];

interface Opsi {
  label: string;
  teks: string;
}

interface Pasangan {
  kiri: string;
  kanan: string;
}

interface Props {
  value: SoalFormValue;
  onChange: (value: SoalFormValue) => void;
  disabled?: boolean;
}

function defaultForTipe(tipe: string): Pick<SoalFormValue, "jawaban" | "kunci"> {
  const opsi4 = { opsi: ["A", "B", "C", "D"].map((label) => ({ label, teks: "" })) };
  switch (tipe) {
    case "pg":
    case "studi_kasus":
    case "pilihan_gambar":
      return { jawaban: opsi4, kunci: { opsi: "A" } };
    case "pg_kompleks":
    case "pg_multi":
      return { jawaban: opsi4, kunci: { opsi: ["A"] } };
    case "benar_salah":
      return { jawaban: { pernyataan: ["", ""] }, kunci: { nilai: [true, true] } };
    case "menjodohkan":
    case "peta_konsep":
      return {
        jawaban: { pasangan: [{ kiri: "", kanan: "" }, { kiri: "", kanan: "" }] },
        // Kunci mulai KOSONG: pasangan yang dicentang "Benar" ditambahkan lewat checkbox.
        // (Sebelumnya kunci ikut diisi 2 pasangan kosong → backend selalu menolak
        // dengan "pasangan[0].kiri dan .kanan wajib diisi".)
        kunci: { pasangan: [] },
      };
    case "isian":
      return { jawaban: {}, kunci: { jawaban: [""] } };
    case "uraian":
    case "upload_file":
    case "praktik_singkat":
      return { jawaban: {}, kunci: { rubrik: "" } };
    case "urutkan":
      return { jawaban: { item: ["", "", ""] }, kunci: { urutan: [0, 1, 2] } };
    case "isi_ganda":
      return { jawaban: { kolom: ["Kolom 1", "Kolom 2"] }, kunci: { jawaban: ["", ""] } };
    case "teks_rumpang":
      return { jawaban: {}, kunci: { jawaban: [""] } };
    case "pernyataan_sikap":
      return { jawaban: { pernyataan: ["", ""], skala: [1, 2, 3, 4] }, kunci: { skala: [1, 2, 3, 4] } };
    default:
      return { jawaban: {}, kunci: {} };
  }
}

const labelCls = "block text-sm font-semibold text-ink";

/** Apakah ada string tak-kosong di dalam struktur jawaban/kunci (rekursif)? */
function adaIsi(isi: unknown): boolean {
  if (typeof isi === "string") return stripHtmlToText(isi).trim() !== "";
  if (Array.isArray(isi)) return isi.some(adaIsi);
  if (isi !== null && typeof isi === "object") return Object.values(isi as Record<string, unknown>).some(adaIsi);
  return false;
}

/**
 * Media soal untuk ImageUpload: hanya URL gambar yang diteruskan.
 * URL video (YouTube/Vimeo) BUKAN gambar — ImageUpload akan mencoba
 * menampilkannya sebagai <img> rusak, jadi diteruskan "" + keterangan.
 */
function urlGambarUntukUpload(url: string): string {
  if (url === "") return "";
  if (/\/api\/v1\/upload\/image\//.test(url)) return url;
  if (/\.(png|jpe?g|webp|gif)(\?|#|$)/i.test(url)) return url;
  return "";
}
function adalahUrlVideo(url: string): boolean {
  return /youtube\.com|youtu\.be|vimeo\.com/.test(url);
}

/** SoalEditor — form dinamis per tipe soal (Tim 3, halaman guru).
 *  Teks soal diedit dengan RichEditor (TipTap): format + tabel + gambar +
 *  video + rumus LaTeX. Penyimpanan TIDAK berubah: teks polos tetap
 *  tersimpan sebagai teks polos (grading tidak berubah); konten berformat
 *  tersimpan sebagai HTML. Kunci jawaban tetap struktur yang sama.
 *
 *  CHROME REDESIGN ANT DESIGN (Tim Form): Dropdown → antd Select,
 *  input/textarea native → antd Input / Input.TextArea / InputNumber,
 *  tombol → antd Button, ConfirmDialog → antd Modal, peringatan →
 *  antd Alert, badge kelompok → antd Tag. Pola state TETAP controlled
 *  value/onChange milik parent (tombol Simpan ada di BankSoalDetailPage,
 *  file tim lain) — karena itu TIDAK dibungkus antd Form onFinish: Form
 *  di sini akan menduplikasi state dan berisiko desinkron. Validasi
 *  tetap milik backend/server seperti sebelumnya.
 *  Kontrol kunci (radio/checkbox) sengaja tetap native — jalur grading.
 */
export default function SoalEditor({ value, onChange, disabled }: Props) {
  const info = TIPE_SOAL[value.tipe];
  const patch = (part: Partial<SoalFormValue>) => onChange({ ...value, ...part });
  const patchJawaban = (jawaban: Record<string, unknown>) => patch({ jawaban });
  const patchKunci = (kunci: Record<string, unknown>) => patch({ kunci });

  const opsi = (Array.isArray(value.jawaban.opsi) ? (value.jawaban.opsi as Opsi[]) : []) as Opsi[];
  const setOpsi = (next: Opsi[]) => patchJawaban({ ...value.jawaban, opsi: next });

  // --- Ganti tipe: konfirmasi dulu bila jawaban/kunci sudah berisi sesuatu,
  // karena ganti tipe mengosongkan jawaban & kunci (sebelumnya reset diam-diam). ---
  const [tipeTertunda, setTipeTertunda] = useState<string | null>(null);
  const [tampilPratinjau, setTampilPratinjau] = useState(false);

  const terapkanTipe = (tipe: string) => {
    const d = defaultForTipe(tipe);
    onChange({ ...value, tipe, jawaban: d.jawaban, kunci: d.kunci });
  };

  const mintaGantiTipe = (tipe: string) => {
    if (tipe === value.tipe) return;
    const jawabanBerisi = adaIsi(value.jawaban);
    const kunciBerubah = JSON.stringify(value.kunci) !== JSON.stringify(defaultForTipe(value.tipe).kunci);
    if (jawabanBerisi || kunciBerubah) {
      setTipeTertunda(tipe);
      return;
    }
    terapkanTipe(tipe);
  };

  const soalPratinjau: Soal = {
    id: "pratinjau",
    bank_id: "",
    tipe: value.tipe,
    kisi_kisi_id: null,
    kompetensi_id: null,
    stimulus: value.stimulus,
    pertanyaan: value.pertanyaan,
    media_url: value.media_url,
    jawaban: value.jawaban,
    kunci: value.kunci,
    pembahasan: value.pembahasan ?? "",
    bobot: value.bobot,
    level: value.level,
    status: value.status,
    created_at: "",
    updated_at: "",
  };

  const renderDinamis = () => {
    switch (value.tipe) {
      case "pg":
      case "studi_kasus":
      case "pilihan_gambar":
      case "pg_kompleks":
      case "pg_multi": {
        const multi = value.tipe === "pg_kompleks" || value.tipe === "pg_multi";
        const kunciOpsi: string[] = multi
          ? (Array.isArray(value.kunci.opsi) ? (value.kunci.opsi as string[]) : [])
          : [];
        const kunciSingle = !multi ? (value.kunci.opsi as string | undefined) ?? "" : "";
        return (
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className={labelCls}>Opsi jawaban</span>
              <Button
                type="default"
                disabled={disabled || opsi.length >= LABELS.length}
                onClick={() => setOpsi([...opsi, { label: LABELS[opsi.length], teks: "" }])}
                icon={<Plus size={14} />}
              >
                Tambah opsi
              </Button>
            </div>
            {opsi.map((o, i) => (
              <div key={o.label} className={`rounded-lg border px-3 py-3 ${(multi ? kunciOpsi.includes(o.label) : kunciSingle === o.label) ? "border-anbk bg-primary-light" : "border-border bg-white"}`}>
                <div className="flex min-h-[44px] items-center gap-2">
                  <span className={`flex h-9 w-9 shrink-0 items-center justify-center rounded-lg text-sm font-bold ${(multi ? kunciOpsi.includes(o.label) : kunciSingle === o.label) ? "bg-anbk text-white" : "bg-surface text-ink"}`}>{o.label}</span>
                  <span className="flex-1 text-sm font-semibold text-ink">
                    {value.tipe === "pilihan_gambar" ? `Gambar opsi ${o.label}` : `Teks opsi ${o.label}`}
                  </span>
                  {multi ? (
                    <Checkbox
                      className="flex min-h-[44px] shrink-0 items-center gap-1.5 text-sm font-medium text-ink"
                      disabled={disabled}
                      checked={kunciOpsi.includes(o.label)}
                      onChange={(e) => {
                        const next = e.target.checked
                          ? [...kunciOpsi, o.label]
                          : kunciOpsi.filter((k) => k !== o.label);
                        patchKunci({ ...value.kunci, opsi: next });
                      }}
                    >
                      Kunci
                    </Checkbox>
                  ) : (
                    <Radio
                      className="flex min-h-[44px] shrink-0 items-center gap-1.5 text-sm font-medium text-ink"
                      name="kunci-opsi"
                      disabled={disabled}
                      checked={kunciSingle === o.label}
                      onChange={() => patchKunci({ ...value.kunci, opsi: o.label })}
                    >
                      Kunci
                    </Radio>
                  )}
                  <Button
                    type="text"
                    danger
                    disabled={disabled || opsi.length <= 2}
                    onClick={() => setOpsi(opsi.filter((_, j) => j !== i).map((item, j) => ({ ...item, label: LABELS[j] })))}
                    title="Hapus opsi"
                    aria-label={`Hapus opsi ${o.label}`}
                    icon={<Trash2 size={16} />}
                  />
                </div>
                {value.tipe === "pilihan_gambar" ? (
                  <div className="mt-2 space-y-2">
                    <ImageUpload
                      label={`Gambar opsi ${o.label}`}
                      value={urlGambarUntukUpload(o.teks)}
                      disabled={disabled}
                      onChange={(url) => {
                        const next = [...opsi];
                        next[i] = { ...o, teks: url };
                        setOpsi(next);
                      }}
                    />
                    <div>
                      <label className="block text-xs font-medium text-muted" htmlFor={`opsi-url-${o.label}`}>
                        atau tempel URL gambar (soal lama)
                      </label>
                      <Input
                        id={`opsi-url-${o.label}`}
                        className="mt-1"
                        size="large"
                        disabled={disabled}
                        value={o.teks}
                        onChange={(e) => {
                          const next = [...opsi];
                          next[i] = { ...o, teks: e.target.value };
                          setOpsi(next);
                        }}
                        placeholder="https://… (gambar opsi)"
                      />
                    </div>
                  </div>
                ) : (
                  <div className="mt-2">
                    <RichEditor
                      compact
                      label={`Teks opsi ${o.label}`}
                      placeholder={`Tulis opsi ${o.label}… (boleh rumus, contoh: $\\frac{1}{2}$)`}
                      value={o.teks}
                      disabled={disabled}
                      onChange={(teks) => {
                        const next = [...opsi];
                        next[i] = { ...o, teks };
                        setOpsi(next);
                      }}
                    />
                  </div>
                )}
              </div>
            ))}
            <p className="text-xs text-muted">
              {multi
                ? "Centang semua opsi yang benar. Nilai proporsional: benar dipilih / total kunci × bobot, tanpa penalti."
                : "Pilih radio pada opsi yang benar (kunci). Nilai otomatis 1/0."}
            </p>
          </div>
        );
      }
      case "benar_salah":
      case "pernyataan_sikap": {
        const pernyataan = (Array.isArray(value.jawaban.pernyataan)
          ? (value.jawaban.pernyataan as string[])
          : []) as string[];
        const nilai = (Array.isArray(value.kunci.nilai)
          ? (value.kunci.nilai as boolean[])
          : []) as boolean[];
        const setPernyataan = (next: string[]) =>
          patchJawaban({ ...value.jawaban, pernyataan: next });
        return (
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className={labelCls}>Pernyataan</span>
              <Button
                type="default"
                disabled={disabled}
                onClick={() => {
                  setPernyataan([...pernyataan, ""]);
                  if (value.tipe === "benar_salah") {
                    patchKunci({ ...value.kunci, nilai: [...nilai, true] });
                  }
                }}
                icon={<Plus size={14} />}
              >
                Tambah pernyataan
              </Button>
            </div>
            {pernyataan.map((p, i) => (
              <div key={i} className="rounded-lg border border-border bg-white p-3">
                <div className="flex items-center gap-2">
                  <span className="w-6 shrink-0 text-center text-sm font-semibold">{i + 1}.</span>
                  <span className="flex-1 text-sm font-semibold text-ink">Pernyataan {i + 1}</span>
                  <Button
                    type="text"
                    danger
                    disabled={disabled || pernyataan.length <= 1}
                    onClick={() => {
                      setPernyataan(pernyataan.filter((_, j) => j !== i));
                      if (value.tipe === "benar_salah") {
                        patchKunci({ ...value.kunci, nilai: nilai.filter((_, j) => j !== i) });
                      }
                    }}
                    title="Hapus pernyataan"
                    aria-label={`Hapus pernyataan ${i + 1}`}
                    icon={<Trash2 size={16} />}
                  />
                </div>
                <div className="mt-2">
                  <RichEditor
                    compact
                    label={`Pernyataan ${i + 1}`}
                    placeholder={`Tulis pernyataan ${i + 1}…`}
                    value={p}
                    disabled={disabled}
                    onChange={(teks) => {
                      const next = [...pernyataan];
                      next[i] = teks;
                      setPernyataan(next);
                    }}
                  />
                </div>
                {value.tipe === "benar_salah" ? (
                  <div className="mt-2 max-w-xs">
                    <label className="block text-xs font-medium text-muted" htmlFor={`bs-${i}`}>
                      Kunci pernyataan ini
                    </label>
                    <Select
                      id={`bs-${i}`}
                      placeholder="Pilih kunci"
                      options={[
                        { value: "benar", label: "Benar" },
                        { value: "salah", label: "Salah" },
                      ]}
                      value={nilai[i] ? "benar" : "salah"}
                      disabled={disabled}
                      onChange={(v) => {
                        const next = [...nilai];
                        next[i] = v === "benar";
                        patchKunci({ ...value.kunci, nilai: next });
                      }}
                      style={{ width: "100%" }}
                      size="large"
                      notFoundContent="Tidak ada data"
                    />
                  </div>
                ) : (
                  <span className="mt-2 block text-xs text-muted">Skala 1–4</span>
                )}
              </div>
            ))}
            {value.tipe === "benar_salah" && (
              <p className="text-xs text-muted">Nilai proporsional per pernyataan benar.</p>
            )}
          </div>
        );
      }
      case "menjodohkan":
      case "peta_konsep": {
        const pasangan = (Array.isArray(value.jawaban.pasangan)
          ? (value.jawaban.pasangan as Pasangan[])
          : []) as Pasangan[];
        // Kunci disaring dari pasangan kosong (sisa data lama) agar tidak pernah
        // terkirim ke backend dan memicu "pasangan[i].kiri dan .kanan wajib diisi".
        const kunciPasangan = (Array.isArray(value.kunci.pasangan)
          ? (value.kunci.pasangan as Pasangan[])
          : []).filter(
            (k) => k.kiri.trim() !== "" && k.kanan.trim() !== "",
          ) as Pasangan[];
        const setPasangan = (next: Pasangan[]) =>
          patchJawaban({ ...value.jawaban, pasangan: next });
        // Sinkronisasi kunci saat pasangan diubah/dihapus: entri kunci yang nilainya
        // sama dengan pasangan lama ikut diperbarui/dihapus, tidak dibiarkan basi.
        const samakanKunci = (lama: Pasangan, baru: Pasangan | null) => {
          if (!kunciPasangan.some((k) => k.kiri === lama.kiri && k.kanan === lama.kanan)) return;
          patchKunci({
            ...value.kunci,
            pasangan: kunciPasangan
              .map((k) => (k.kiri === lama.kiri && k.kanan === lama.kanan ? baru : k))
              .filter((k): k is Pasangan => k !== null),
          });
        };
        return (
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className={labelCls}>Pasangan (kiri → kanan)</span>
              <Button
                type="default"
                disabled={disabled}
                onClick={() => setPasangan([...pasangan, { kiri: "", kanan: "" }])}
                icon={<Plus size={14} />}
              >
                Tambah pasangan
              </Button>
            </div>
            {pasangan.map((p, i) => {
              const isKunci = kunciPasangan.some((k) => k.kiri === p.kiri && k.kanan === p.kanan);
              return (
                <div key={i} className="rounded-lg border border-border bg-white p-3">
                  <div className="flex items-center gap-2">
                    <span className="flex-1 text-sm font-semibold text-ink">Pasangan {i + 1}</span>
                    <Checkbox
                      className="flex min-h-[44px] shrink-0 items-center gap-1.5 text-sm font-medium text-ink"
                      disabled={disabled}
                      checked={isKunci}
                      onChange={(e) => {
                        const next = e.target.checked
                          ? [...kunciPasangan, { ...p }]
                          : kunciPasangan.filter((k) => !(k.kiri === p.kiri && k.kanan === p.kanan));
                        patchKunci({ ...value.kunci, pasangan: next });
                      }}
                    >
                      Benar
                    </Checkbox>
                    <Button
                      type="text"
                      danger
                      disabled={disabled || pasangan.length <= 1}
                      onClick={() => {
                        setPasangan(pasangan.filter((_, j) => j !== i));
                        samakanKunci(p, null);
                      }}
                      title="Hapus pasangan"
                      aria-label={`Hapus pasangan ${i + 1}`}
                      icon={<Trash2 size={16} />}
                    />
                  </div>
                  <div className="mt-2 grid grid-cols-1 gap-2 md:grid-cols-2">
                    <RichEditor
                      compact
                      label={`Kiri pasangan ${i + 1} (premis / simpul)`}
                      placeholder="Kiri (premis / simpul)"
                      value={p.kiri}
                      disabled={disabled}
                      onChange={(teks) => {
                        const next = [...pasangan];
                        next[i] = { ...p, kiri: teks };
                        setPasangan(next);
                        samakanKunci(p, next[i]);
                      }}
                    />
                    <RichEditor
                      compact
                      label={`Kanan pasangan ${i + 1} (respons / simpul)`}
                      placeholder="Kanan (respons / simpul)"
                      value={p.kanan}
                      disabled={disabled}
                      onChange={(teks) => {
                        const next = [...pasangan];
                        next[i] = { ...p, kanan: teks };
                        setPasangan(next);
                        samakanKunci(p, next[i]);
                      }}
                    />
                  </div>
                </div>
              );
            })}
            <p className="text-xs text-muted">
              Centang pasangan yang benar (kunci). Nilai proporsional per pasangan benar.
            </p>
          </div>
        );
      }
      case "isian":
        return (
          <div>
            <label className={labelCls} htmlFor="kunci-isian">Kunci jawaban (satu per baris = jawaban alternatif)</label>
            <Input.TextArea
              id="kunci-isian"
              disabled={disabled}
              rows={3}
              size="large"
              value={(Array.isArray(value.kunci.jawaban) ? (value.kunci.jawaban as string[]) : []).join("\n")}
              onChange={(e) =>
                patchKunci({ ...value.kunci, jawaban: e.target.value.split("\n").map((s) => s.trim()).filter(Boolean) })
              }
              placeholder={"Jakarta\nDKI Jakarta"}
            />
            <p className="mt-1 text-xs text-muted">Nilai exact-match setelah normalisasi (trim + lowercase).</p>
          </div>
        );
      case "uraian":
      case "upload_file":
      case "praktik_singkat":
        return (
          <div>
            <label className={labelCls} htmlFor="rubrik">Rubrik penilaian (kunci)</label>
            <Input.TextArea
              id="rubrik"
              disabled={disabled}
              rows={4}
              size="large"
              value={(value.kunci.rubrik as string | undefined) ?? ""}
              onChange={(e) => patchKunci({ ...value.kunci, rubrik: e.target.value })}
              placeholder="Tulis kriteria penskoran untuk pengoreksi…"
            />
            <p className="mt-1 text-xs text-muted">Nilai MANUAL oleh guru/penguji (0–100 × bobot).</p>
          </div>
        );
      case "urutkan": {
        const items = (Array.isArray(value.jawaban.item) ? (value.jawaban.item as string[]) : []) as string[];
        const urutan = (Array.isArray(value.kunci.urutan) ? (value.kunci.urutan as number[]) : []) as number[];
        return (
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className={labelCls}>Butir yang diurutkan</span>
              <Button
                type="default"
                disabled={disabled}
                onClick={() => {
                  patchJawaban({ ...value.jawaban, item: [...items, ""] });
                  patchKunci({ ...value.kunci, urutan: [...urutan, items.length] });
                }}
                icon={<Plus size={14} />}
              >
                Tambah butir
              </Button>
            </div>
            {items.map((item, i) => (
              <div key={i} className="rounded-lg border border-border bg-white p-3">
                <div className="flex items-center gap-2">
                  <span className="w-6 shrink-0 text-center text-sm font-semibold">{i + 1}.</span>
                  <span className="flex-1 text-sm font-semibold text-ink">Butir {i + 1}</span>
                  <Button
                    type="text"
                    danger
                    disabled={disabled || items.length <= 2}
                    onClick={() => {
                      patchJawaban({ ...value.jawaban, item: items.filter((_, j) => j !== i) });
                      patchKunci({ ...value.kunci, urutan: urutan.filter((v) => v !== i).map((v) => (v > i ? v - 1 : v)) });
                    }}
                    title="Hapus butir"
                    aria-label={`Hapus butir ${i + 1}`}
                    icon={<Trash2 size={16} />}
                  />
                </div>
                <div className="mt-2">
                  <RichEditor
                    compact
                    label={`Butir ${i + 1} yang diurutkan`}
                    placeholder={`Butir ${i + 1}`}
                    value={item}
                    disabled={disabled}
                    onChange={(teks) => {
                      const next = [...items];
                      next[i] = teks;
                      patchJawaban({ ...value.jawaban, item: next });
                    }}
                  />
                </div>
              </div>
            ))}
            <div>
              <span className={labelCls}>Urutan benar (pilih butir untuk tiap posisi)</span>
              <div className="mt-1 flex flex-col gap-2 sm:flex-row sm:flex-wrap">
                {urutan.map((idx, pos) => (
                  <div key={pos} className="min-w-56 flex-1">
                    <label className="block text-xs font-medium text-muted" htmlFor={`urutan-${pos}`}>
                      Posisi {pos + 1}
                    </label>
                    <Select
                      id={`urutan-${pos}`}
                      placeholder="Pilih butir"
                      options={items.map((item, j) => ({
                        value: String(j),
                        label: `Posisi ${pos + 1}: ${stripHtmlToText(item) || `Butir ${j + 1}`}`,
                      }))}
                      value={String(idx)}
                      disabled={disabled}
                      onChange={(v) => {
                        const next = [...urutan];
                        next[pos] = Number(v);
                        patchKunci({ ...value.kunci, urutan: next });
                      }}
                      style={{ width: "100%" }}
                      size="large"
                      notFoundContent="Tidak ada data"
                    />
                  </div>
                ))}
              </div>
              <p className="mt-1 text-xs text-muted">Nilai proporsional per posisi benar.</p>
            </div>
          </div>
        );
      }
      case "isi_ganda": {
        const kolom = (Array.isArray(value.jawaban.kolom) ? (value.jawaban.kolom as string[]) : []) as string[];
        const jawaban = (Array.isArray(value.kunci.jawaban) ? (value.kunci.jawaban as string[]) : []) as string[];
        return (
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className={labelCls}>Kolom isian</span>
              <Button
                type="default"
                disabled={disabled}
                onClick={() => {
                  patchJawaban({ ...value.jawaban, kolom: [...kolom, `Kolom ${kolom.length + 1}`] });
                  patchKunci({ ...value.kunci, jawaban: [...jawaban, ""] });
                }}
                icon={<Plus size={14} />}
              >
                Tambah kolom
              </Button>
            </div>
            {kolom.map((k, i) => (
              <div key={i} className="flex flex-col gap-2 rounded-lg border border-border bg-white p-3 sm:flex-row sm:items-center">
                <Input
                  disabled={disabled}
                  size="large"
                  value={k}
                  onChange={(e) => {
                    const next = [...kolom];
                    next[i] = e.target.value;
                    patchJawaban({ ...value.jawaban, kolom: next });
                  }}
                  placeholder={`Label kolom ${i + 1}`}
                  aria-label={`Label kolom ${i + 1}`}
                />
                <Input
                  disabled={disabled}
                  size="large"
                  value={jawaban[i] ?? ""}
                  onChange={(e) => {
                    const next = [...jawaban];
                    next[i] = e.target.value;
                    patchKunci({ ...value.kunci, jawaban: next });
                  }}
                  placeholder={`Kunci kolom ${i + 1}`}
                  aria-label={`Kunci kolom ${i + 1}`}
                />
                <Button
                  type="text"
                  danger
                  disabled={disabled || kolom.length <= 1}
                  onClick={() => {
                    patchJawaban({ ...value.jawaban, kolom: kolom.filter((_, j) => j !== i) });
                    patchKunci({ ...value.kunci, jawaban: jawaban.filter((_, j) => j !== i) });
                  }}
                  title="Hapus kolom"
                  aria-label={`Hapus kolom ${i + 1}`}
                  icon={<Trash2 size={16} />}
                />
              </div>
            ))}
            <p className="text-xs text-muted">
              Label kolom &amp; kunci tetap teks polos (kunci dipakai grading exact-match). Nilai proporsional per kolom benar.
            </p>
          </div>
        );
      }
      case "teks_rumpang": {
        const rumpang = (value.pertanyaan.match(/___/g) ?? []).length;
        return (
          <div>
            <label className={labelCls} htmlFor="kunci-rumpang">
              Kunci per rumpang (satu per baris, urut sesuai ___ di pertanyaan)
            </label>
            <Input.TextArea
              id="kunci-rumpang"
              disabled={disabled}
              rows={3}
              size="large"
              value={(Array.isArray(value.kunci.jawaban) ? (value.kunci.jawaban as string[]) : []).join("\n")}
              onChange={(e) =>
                patchKunci({ ...value.kunci, jawaban: e.target.value.split("\n").map((s) => s.trim()).filter(Boolean) })
              }
            />
            <p className="mt-1 text-xs text-muted">
              Ditemukan {rumpang} rumpang (___) di pertanyaan. Jumlah kunci harus sama.
            </p>
          </div>
        );
      }
      default:
        return null;
    }
  };

  return (
    <div className="space-y-4">
      <div>
        <label className={labelCls} htmlFor="tipe-soal">Tipe soal</label>
        <Select
          id="tipe-soal"
          placeholder="Pilih tipe soal"
          showSearch
          options={TIPE_LIST.map((t) => ({
            value: t.kode,
            label: `${t.label} · ${badgeKelompok(t.kelompok)}`,
          }))}
          value={value.tipe}
          disabled={disabled}
          onChange={mintaGantiTipe}
          filterOption={(input, opt) =>
            String(opt?.label ?? "").toLowerCase().includes(input.toLowerCase())
          }
          optionFilterProp="label"
          style={{ width: "100%" }}
          size="large"
          notFoundContent="Tidak ada data"
        />
        {info && (
          <div className="mt-2 flex flex-wrap items-center gap-2">
            <Tag color={info.kelompok === "inti" ? "blue" : "purple"}>
              [{badgeKelompok(info.kelompok)}]
            </Tag>
            <span className="text-xs text-muted">{info.deskripsi}</span>
          </div>
        )}
        {info && !info.dinilai && (
          <Alert
            className="mt-2"
            type="warning"
            showIcon
            message={
              <span>Soal ini <strong>tidak dinilai</strong> — hanya survei/refleksi, tidak masuk nilai akhir.</span>
            }
          />
        )}
      </div>

      <div>
        <span className={labelCls} id="label-stimulus">Stimulus / bacaan / kasus (opsional{value.tipe === "studi_kasus" ? ", wajib untuk Studi Kasus" : ""})</span>
        <div className="mt-1" aria-labelledby="label-stimulus">
          <RichEditor
            label="Stimulus"
            placeholder="Teks bacaan, kasus, data, atau tabel pendukung soal…"
            value={value.stimulus}
            disabled={disabled}
            onChange={(stimulus) => patch({ stimulus })}
          />
        </div>
      </div>

      <div>
        <span className={labelCls} id="label-pertanyaan">Pertanyaan</span>
        <div className="mt-1" aria-labelledby="label-pertanyaan">
          <RichEditor
            label="Pertanyaan"
            placeholder="Tulis pertanyaan soal… (untuk Teks Rumpang, tandai isian dengan ___)"
            value={value.pertanyaan}
            disabled={disabled}
            onChange={(pertanyaan) => patch({ pertanyaan })}
          />
        </div>
      </div>

      <div className="space-y-2">
        <span className={labelCls} id="label-media">Media (gambar / video, opsional)</span>
        <ImageUpload
          label="Media soal (gambar)"
          value={urlGambarUntukUpload(value.media_url)}
          disabled={disabled}
          onChange={(url) => patch({ media_url: url })}
        />
        {adalahUrlVideo(value.media_url) && (
          <p className="text-xs text-primary">
            Media saat ini adalah video — akan ditampilkan sebagai video embed di pratinjau. Ganti gambar di atas, atau ganti URL video di bawah.
          </p>
        )}
        <div>
          <label className="block text-xs font-medium text-muted" htmlFor="media-url">
            atau tempel URL (gambar / video YouTube / Vimeo — soal lama)
          </label>
          <Input
            id="media-url"
            className="mt-1"
            size="large"
            disabled={disabled}
            value={value.media_url}
            onChange={(e) => patch({ media_url: e.target.value })}
            placeholder="https://…"
          />
        </div>
        <p className="text-xs text-muted">
          URL gambar tampil sebagai gambar; URL YouTube/Vimeo tampil sebagai video embed. File video TIDAK diunggah (tidak ada endpoint-nya).
        </p>
      </div>

      <div>
        <span className={labelCls} id="label-pembahasan">Pembahasan (tampil di Tryout / bila hasil ditampilkan)</span>
        <div className="mt-1" aria-labelledby="label-pembahasan">
          <RichEditor
            label="Pembahasan"
            placeholder="Jelaskan mengapa jawaban benar itu benar (dan mengapa pengecoh salah)…"
            value={value.pembahasan ?? ""}
            disabled={disabled}
            showVideo={false}
            showImage={false}
            showTable={false}
            showHeading={false}
            onChange={(pembahasan) => patch({ pembahasan })}
          />
        </div>
        <p className="mt-1 text-xs text-muted">
          Tidak pernah dikirim ke siswa sebelum submit — hanya tampil di halaman hasil setelah ujian selesai.
        </p>
      </div>

      {renderDinamis()}

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <div>
          <label className={labelCls} htmlFor="bobot">Bobot</label>
          <InputNumber
            id="bobot"
            min={1}
            disabled={disabled}
            value={value.bobot}
            onChange={(v) => patch({ bobot: Math.max(1, typeof v === "number" ? v : 1) })}
            style={{ width: "100%" }}
            size="large"
          />
        </div>
        <div>
          <label className={labelCls} htmlFor="level">Level</label>
          <Select
            id="level"
            placeholder="Pilih level"
            options={[
              { value: "mudah", label: "Mudah" },
              { value: "sedang", label: "Sedang" },
              { value: "sulit", label: "Sulit" },
            ]}
            value={value.level}
            disabled={disabled}
            onChange={(v) => patch({ level: v as SoalFormValue["level"] })}
            style={{ width: "100%" }}
            size="large"
            notFoundContent="Tidak ada data"
          />
        </div>
        <div>
          <label className={labelCls} htmlFor="status">Status</label>
          <Select
            id="status"
            placeholder="Pilih status"
            options={[
              { value: "draft", label: "Draft" },
              { value: "aktif", label: "Aktif" },
              { value: "nonaktif", label: "Nonaktif" },
              { value: "perlu_revisi", label: "Perlu Revisi" },
              { value: "diarsipkan", label: "Diarsipkan" },
            ]}
            value={value.status}
            disabled={disabled}
            onChange={(v) => patch({ status: v as SoalFormValue["status"] })}
            style={{ width: "100%" }}
            size="large"
            notFoundContent="Tidak ada data"
          />
        </div>
      </div>

      <div>
        <Button
          type="default"
          disabled={disabled}
          onClick={() => setTampilPratinjau(!tampilPratinjau)}
          aria-expanded={tampilPratinjau}
          icon={tampilPratinjau ? <EyeOff size={16} /> : <Eye size={16} />}
        >
          {tampilPratinjau ? "Sembunyikan pratinjau" : "Pratinjau soal"}
        </Button>
        {tampilPratinjau && (
          <div className="mt-3">
            <SoalRenderer soal={soalPratinjau} showKunci />
          </div>
        )}
      </div>

      <Modal
        open={tipeTertunda !== null}
        title="Ganti tipe soal?"
        okText="Ya, ganti tipe"
        cancelText="Batal"
        onOk={() => {
          if (tipeTertunda !== null) terapkanTipe(tipeTertunda);
          setTipeTertunda(null);
        }}
        onCancel={() => setTipeTertunda(null)}
      >
        <span>
          Ganti tipe akan <strong>mengosongkan jawaban &amp; kunci</strong> yang sudah diisi.
          Stimulus, pertanyaan, dan pembahasan tetap tersimpan. Lanjutkan?
        </span>
      </Modal>
    </div>
  );
}
