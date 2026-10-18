// SoalRenderer (INTERAKTIF, untuk halaman SISWA) — Tim 4.
//
// BERBEDA dari components/soal/SoalRenderer.tsx (Tim 3) yang read-only
// untuk pratinjau guru dan TIDAK BOLEH dipakai di halaman siswa.
// Komponen ini menerima input siswa dan memanggil onChange; kunci
// jawaban tidak pernah ada di sini (tidak dikirim backend ke browser).
//
// Bentuk `tampilan` (kolom soal.jawaban Tim 3):
//   pg/pg_kompleks/pg_multi/pilihan_gambar/studi_kasus: {opsi:[{label,teks}]}
//   benar_salah: {pernyataan:[...]} ; pernyataan_sikap: {pernyataan, skala}
//   menjodohkan/peta_konsep: {pasangan:[{kiri,kanan}]}
//   isi_ganda: {kolom:[...]} ; urutkan: {item:[...]} ; lainnya: {}
//
// TAMPILAN: semua teks soal (opsi / pernyataan / pasangan / butir /
// segmen rumpang) di-render lewat RichContent (HTML tersanitasi DOMPurify
// + rumus KaTeX; teks polos lama tampil persis seperti sebelumnya).
// Nilai `value` / format yang dikirim ke backend TIDAK BERUBAH.
// <option> tidak bisa menampung HTML -> teksnya di-strip ke teks polos,
// tapi `value` option tetap string mentah (kunci pemetaan jawaban).
//   pg/dll: string (label) | pg_kompleks/pg_multi: string[]
//   benar_salah: boolean[] | menjodohkan: {kiri: kanan}
//   isian: string | isi_ganda/teks_rumpang: string[]
//   urutkan: number[] (indeks ke item asli) | uraian/dll: string
//   pernyataan_sikap: number[]
//
// REDESIGN antd (hanya presentasi): Radio / Checkbox / Input /
// Input.TextArea / Button / Select antd. Menjodohkan / Peta Konsep memakai
// antd Select (§2.10 AUDIT-KONSISTENSI — pengecualian lama DICABUT):
// onChange memetakan undefined → "" agar format value backend IDENTIK
// ({ kiri: "" } untuk jawaban kosong). Nilai onChange, format value,
// dan perilaku jawab TIDAK diubah.
import { useMemo } from "react";
import { Button, Checkbox, Input, Radio, Select } from "antd";
import RichContent, { stripHtmlToText } from "../soal/RichContent";

export type NilaiJawaban = unknown;

interface Props {
  tipe: string;
  tampilan: Record<string, unknown> | null;
  pertanyaan: string;
  value: NilaiJawaban;
  onChange: (value: NilaiJawaban) => void;
  disabled?: boolean;
}

function normalisasiTipe(t: string): string {
  const x = t.toLowerCase().replace(/-/g, "_").replace(/ /g, "_");
  switch (x) {
    case "isian_singkat":
    case "jawaban_singkat":
      return "isian";
    case "mengurutkan":
      return "urutkan";
    case "skala_sikap":
    case "skala":
      return "pernyataan_sikap";
    case "esai":
    case "essay":
    case "isian_panjang":
    case "isi_panjang":
    case "esai_singkat":
      return "uraian";
    default:
      return x;
  }
}

function asArray<T>(v: unknown): T[] {
  return Array.isArray(v) ? (v as T[]) : [];
}

interface Opsi {
  label: string;
  teks: string;
}

const inputStyle: React.CSSProperties = { fontSize: 16 };

export default function SoalRenderer({ tipe, tampilan, pertanyaan, value, onChange, disabled }: Props) {
  const t = normalisasiTipe(tipe);
  const data = tampilan ?? {};

  const opsi = useMemo(() => asArray<Opsi>(data.opsi), [data.opsi]);
  const pernyataan = useMemo(() => asArray<string>(data.pernyataan), [data.pernyataan]);
  const pasangan = useMemo(
    () => asArray<{ kiri: string; kanan: string }>(data.pasangan),
    [data.pasangan],
  );
  const kolom = useMemo(() => asArray<string>(data.kolom), [data.kolom]);
  const item = useMemo(() => asArray<string>(data.item), [data.item]);
  const skala = useMemo(() => {
    const s = asArray<number>(data.skala);
    return s.length > 0 ? s : [1, 2, 3, 4];
  }, [data.skala]);
  const kananOptions = useMemo(() => pasangan.map((p) => p.kanan), [pasangan]);

  // --- PG tunggal / pilihan gambar / studi kasus: radio, value = label ---
  if (t === "pg" || t === "pilihan_gambar" || t === "studi_kasus") {
    const dipilih = typeof value === "string" ? value : "";
    return (
      <div className="space-y-2" role="radiogroup" aria-label="Pilihan jawaban">
        {opsi.map((o) => (
          <div
            key={o.label}
            className={`flex min-h-[44px] min-w-0 items-start gap-3 break-words rounded-lg border px-3 py-2 text-base ${
              dipilih === o.label ? "border-primary bg-primary-light" : "border-border"
            }`}
          >
            <Radio
              checked={dipilih === o.label}
              disabled={disabled}
              onChange={() => onChange(o.label)}
              className="mt-[3px] shrink-0"
            >
              <span className="font-semibold">{o.label}.</span>{" "}
              {t === "pilihan_gambar" ? (
                <img src={o.teks} alt={`Opsi ${o.label}`} className="max-h-24 rounded object-contain" />
              ) : (
                <span className="min-w-0 flex-1 break-words">
                  <RichContent html={o.teks} />
                </span>
              )}
            </Radio>
          </div>
        ))}
      </div>
    );
  }

  // --- PG Kompleks / PG Multi: checkbox, value = string[] ---
  if (t === "pg_kompleks" || t === "pg_multi") {
    const dipilih = asArray<string>(value);
    const toggle = (label: string) => {
      onChange(dipilih.includes(label) ? dipilih.filter((x) => x !== label) : [...dipilih, label]);
    };
    return (
      <div className="space-y-2">
        <p className="text-sm text-muted">Pilih semua jawaban yang benar.</p>
        {opsi.map((o) => (
          <div
            key={o.label}
            className={`flex min-h-[44px] min-w-0 items-start gap-3 break-words rounded-lg border px-3 py-2 text-base ${
              dipilih.includes(o.label) ? "border-primary bg-primary-light" : "border-border"
            }`}
          >
            <Checkbox
              checked={dipilih.includes(o.label)}
              disabled={disabled}
              onChange={() => toggle(o.label)}
              className="mt-[3px] shrink-0"
            >
              <span className="font-semibold">{o.label}.</span>{" "}
              <span className="min-w-0 flex-1 break-words">
                <RichContent html={o.teks} />
              </span>
            </Checkbox>
          </div>
        ))}
      </div>
    );
  }

  // --- Benar-Salah: per pernyataan, value = boolean[] ---
  if (t === "benar_salah") {
    const nilai = asArray<boolean>(value);
    const setNilai = (i: number, v: boolean) => {
      const next = [...nilai];
      next[i] = v;
      onChange(next);
    };
    return (
      <div className="space-y-2">
        {pernyataan.map((p, i) => (
          <div key={i} className="min-w-0 rounded-lg border border-border px-3 py-2">
            <div className="break-words text-base">
              <RichContent html={p} />
            </div>
            <div className="mt-2 flex gap-2" role="radiogroup" aria-label={`Pernyataan ${i + 1}`}>
              {([true, false] as const).map((v) => (
                <div
                  key={String(v)}
                  className={`flex min-h-[44px] flex-1 items-center justify-center gap-2 rounded-lg border px-3 text-base ${
                    nilai[i] === v ? "border-primary bg-primary-light font-semibold" : "border-border"
                  }`}
                >
                  <Radio
                    checked={nilai[i] === v}
                    disabled={disabled}
                    onChange={() => setNilai(i, v)}
                  >
                    {v ? "Benar" : "Salah"}
                  </Radio>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    );
  }

  // --- Menjodohkan / Peta Konsep: Select kanan per kiri, value = {kiri: kanan} ---
  // Migrasi antd Select (§2.10) — format value TIDAK berubah:
  // kosong di Select = undefined, tapi onChange memetakan undefined → ""
  // agar backend tetap menerima { kiri: "" } (bukan undefined / key hilang).
  // allowClear = padanan opsi kosong "— Pilih pasangan —" (bisa batal pilih).
  // showSearch TIDAK diaktifkan: perilaku keyboard sedekat mungkin dengan
  // <select> native (panah + Enter) pada jalur penilaian ini.
  // Satu baris = satu grid [nomor 28px | kiri 1fr | panah 24px | Select 1fr]
  // (P-06 AUDIT-HALAMAN-UJIAN): panah SELALU terlihat (termasuk HP),
  // kotak kiri min-height 44px sejajar Select 44px, semua sel center.
  if (t === "menjodohkan" || t === "peta_konsep") {
    const peta = (value && typeof value === "object" && !Array.isArray(value)
      ? (value as Record<string, string>)
      : {}) as Record<string, string>;
    return (
      <div className="space-y-2">
        {pasangan.map((p, i) => (
          <div
            key={i}
            className="grid items-center gap-2"
            style={{ gridTemplateColumns: "28px minmax(0, 1fr) 24px minmax(0, 1fr)" }}
          >
            <span className="text-center text-sm font-semibold text-muted">{i + 1}</span>
            <span className="flex min-h-[44px] min-w-0 items-center break-words rounded-lg bg-surface px-3 py-2 text-base">
              <RichContent html={p.kiri} />
            </span>
            <span className="text-center text-muted" aria-hidden>
              →
            </span>
            <Select
              aria-label={`Pasangan untuk ${stripHtmlToText(p.kiri)}`}
              value={peta[p.kiri] || undefined}
              placeholder="— Pilih pasangan —"
              allowClear
              disabled={disabled}
              onChange={(v) => onChange({ ...peta, [p.kiri]: v ?? "" })}
              options={kananOptions.map((k) => ({ value: k, label: stripHtmlToText(k) }))}
              size="large"
              className="w-full min-w-0"
              style={{ fontSize: 16 }}
            />
          </div>
        ))}
      </div>
    );
  }

  // --- Isian singkat: text, value = string ---
  if (t === "isian") {
    return (
      <Input
        value={typeof value === "string" ? value : ""}
        disabled={disabled}
        onChange={(e) => onChange(e.target.value)}
        placeholder="Ketik jawaban singkat…"
        aria-label="Jawaban isian singkat"
        size="large"
        style={inputStyle}
      />
    );
  }

  // --- Isi ganda: satu input per kolom, value = string[] ---
  if (t === "isi_ganda") {
    const nilai = asArray<string>(value);
    return (
      <div className="space-y-2">
        {kolom.map((k, i) => (
          <div key={i} className="flex flex-col gap-2 sm:flex-row sm:items-center">
            <span className="min-w-0 break-words text-base font-medium sm:w-32 sm:shrink-0">
              <RichContent html={k} />
            </span>
            <Input
              value={nilai[i] ?? ""}
              disabled={disabled}
              onChange={(e) => {
                const next = [...nilai];
                next[i] = e.target.value;
                onChange(next);
              }}
              aria-label={`Isian ${stripHtmlToText(k)}`}
              size="large"
              style={inputStyle}
            />
          </div>
        ))}
      </div>
    );
  }

  // --- Teks rumpang: input inline di antara segmen "___", value = string[] ---
  if (t === "teks_rumpang") {
    const segmen = pertanyaan.split("___");
    const nilai = asArray<string>(value);
    return (
      <div className="text-base leading-loose">
        {segmen.map((seg, i) => (
          <span key={i}>
            <RichContent html={seg} />
            {i < segmen.length - 1 && (
              <Input
                value={nilai[i] ?? ""}
                disabled={disabled}
                onChange={(e) => {
                  const next = [...nilai];
                  next[i] = e.target.value;
                  onChange(next);
                }}
                aria-label={`Rumpang ${i + 1}`}
                size="large"
                // Teks Rumpang: lebar adaptif (min-width + max-width 100%),
                // bukan lebar tetap 160px (P-06).
                style={{ minWidth: 128, width: "12em", maxWidth: "100%", display: "inline-block", margin: "0 4px", ...inputStyle }}
              />
            )}
          </span>
        ))}
      </div>
    );
  }

  // --- Urutkan: tombol naik/turun, value = number[] (indeks item asli) ---
  if (t === "urutkan") {
    const urutan: number[] = asArray<number>(value).length === item.length
      ? (asArray<number>(value) as number[])
      : item.map((_, i) => i);
    const geser = (pos: number, arah: -1 | 1) => {
      const next = [...urutan];
      const tujuan = pos + arah;
      if (tujuan < 0 || tujuan >= next.length) return;
      [next[pos], next[tujuan]] = [next[tujuan], next[pos]];
      onChange(next);
    };
    return (
      <div className="space-y-2">
        <p className="text-sm text-muted">Urutkan dengan tombol naik / turun.</p>
        {urutan.map((idx, pos) => (
          <div key={idx} className="flex items-center gap-2 rounded-lg border border-border px-3 py-2">
            <span className="w-6 shrink-0 font-semibold text-muted">{pos + 1}.</span>
            <span className="min-w-0 flex-1 break-words text-base">
              <RichContent html={item[idx]} />
            </span>
            <Button
              disabled={disabled || pos === 0}
              onClick={() => geser(pos, -1)}
              aria-label={`Naikkan ${stripHtmlToText(item[idx])}`}
              style={{ minWidth: 44, minHeight: 44 }}
            >
              ↑
            </Button>
            <Button
              disabled={disabled || pos === urutan.length - 1}
              onClick={() => geser(pos, 1)}
              aria-label={`Turunkan ${stripHtmlToText(item[idx])}`}
              style={{ minWidth: 44, minHeight: 44 }}
            >
              ↓
            </Button>
          </div>
        ))}
      </div>
    );
  }

  // --- Skala sikap: radio skala per pernyataan, value = number[] (tidak dinilai) ---
  if (t === "pernyataan_sikap") {
    const nilai = asArray<number>(value);
    return (
      <div className="space-y-2">
        <p className="text-sm text-muted">Survei — tidak memengaruhi nilai.</p>
        {pernyataan.map((p, i) => (
          <div key={i} className="min-w-0 rounded-lg border border-border px-3 py-2">
            <div className="break-words text-base">
              <RichContent html={p} />
            </div>
            <div className="mt-2 flex flex-wrap gap-2" role="radiogroup" aria-label={`Sikap ${i + 1}`}>
              {skala.map((s) => (
                <div
                  key={s}
                  className={`flex min-h-[44px] min-w-[44px] items-center justify-center rounded-lg border px-3 ${
                    nilai[i] === s ? "border-primary bg-primary-light font-semibold" : "border-border"
                  }`}
                >
                  <Radio
                    checked={nilai[i] === s}
                    disabled={disabled}
                    onChange={() => {
                      const next = [...nilai];
                      next[i] = s;
                      onChange(next);
                    }}
                  >
                    {s}
                  </Radio>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    );
  }

  // --- Uraian / upload / praktik: textarea, value = string (manual) ---
  return (
    <Input.TextArea
      value={typeof value === "string" ? value : ""}
      disabled={disabled}
      onChange={(e) => onChange(e.target.value)}
      placeholder={
        t === "upload_file"
          ? "Tulis keterangan berkas jawaban… (unggah berkas menyusul sesuai arahan pengawas)"
          : "Ketik jawaban Anda…"
      }
      rows={6}
      aria-label="Jawaban uraian"
      style={inputStyle}
    />
  );
}
