// RichEditor — editor soal berbasis TipTap (Tim E, tanpa CDN).
//
// Fitur: tebal / miring / garis bawah / coret, list bullet + nomor,
// heading H2/H3, tabel sederhana, blok kode (+ kode inline), undo/redo,
// hapus format, upload gambar (pakai ImageUpload Tim D — komponen ini
// TIDAK punya upload sendiri), embed video (tempel link YouTube/Vimeo ->
// disisipkan sebagai node videoEmbed; TIDAK ada upload file video —
// tidak ada endpoint-nya), tombol rumus Σ (tulis LaTeX, pratinjau
// langsung via KaTeX, disisipkan sebagai `$...$` / `$$...$$`),
// pratinjau hasil render (sanitasi + KaTeX via RichContent).
//
// FORMAT PENYIMPANAN (kontrak dengan backend — TIDAK berubah):
// - Konten yang hanya teks polos (satu/beberapa paragraf tanpa format,
//   tanpa rumus blok, tanpa tabel/gambar/video) disimpan APA ADANYA
//   sebagai teks polos — soal lama & halaman siswa yang masih menampilkan
//   teks polos TIDAK rusak, dan grading (yang membandingkan string)
//   TIDAK berubah.
// - Konten yang mengandung format / rumus / tabel / gambar / video
//   disimpan sebagai HTML (dari TipTap). Rumus di dalamnya tetap LaTeX
//   ber-delimiter `$...$` / `$$...$$` (bukan gambar).
// - Video disimpan sebagai `<div data-video-embed="URL-embed">`
//   (bukan <iframe>), diubah menjadi <iframe> saat render oleh RichContent.
import { useEffect, useMemo, useRef, useState } from "react";
import { EditorContent, useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Underline from "@tiptap/extension-underline";
import Image from "@tiptap/extension-image";
import { Table } from "@tiptap/extension-table";
import { TableRow } from "@tiptap/extension-table-row";
import { TableHeader } from "@tiptap/extension-table-header";
import { TableCell } from "@tiptap/extension-table-cell";
import Placeholder from "@tiptap/extension-placeholder";
import { Node as TiptapNode } from "@tiptap/core";
import katex from "katex";
import {
  Bold, Italic, Underline as UnderlineIcon, Strikethrough, List, ListOrdered,
  Heading2, Heading3, Table as TableIcon, Undo2, Redo2, Eraser, Code2,
  Braces, ImagePlus, PlaySquare, Eye, EyeOff, X, Plus, Minus, Trash2,
} from "lucide-react";
import { Alert, Button, Input, Radio, Typography } from "antd";
import ImageUpload from "../ui/ImageUpload";
import RichContent, { embedUrlFromWatchUrl, isRichHtml } from "./RichContent";

/** Node blok video: disimpan sebagai div[data-video-embed], di-render jadi iframe oleh RichContent. */
const VideoEmbed = TiptapNode.create({
  name: "videoEmbed",
  group: "block",
  atom: true,
  draggable: true,
  addAttributes() {
    return {
      src: { default: null },
      title: { default: "Video" },
    };
  },
  parseHTML() {
    return [{
      tag: "div[data-video-embed]",
      getAttrs: (el) => ({
        src: (el as HTMLElement).getAttribute("data-video-embed"),
        title: (el as HTMLElement).getAttribute("data-video-title") ?? "Video",
      }),
    }];
  },
  renderHTML({ HTMLAttributes }) {
    return [
      "div",
      {
        "data-video-embed": HTMLAttributes.src,
        "data-video-title": HTMLAttributes.title,
      },
      ["span", {}, `▶ Video: ${String(HTMLAttributes.title ?? "Video")}`],
    ];
  },
});

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

/** Teks polos -> HTML TipTap (paragraf per blok kosong, <br> per baris). */
function plainToHtml(value: string): string {
  if (value === "") return "<p></p>";
  if (isRichHtml(value)) return value;
  return value
    .split(/\n{2,}/)
    .map((paragraf) => `<p>${paragraf.split("\n").map(escapeHtml).join("<br>")}</p>`)
    .join("");
}

/**
 * HTML -> teks polos bila isinya HANYA paragraf teks polos (teks + <br>).
 * Mengembalikan null bila ada format / tabel / gambar / video / list /
 * heading / kode / rumus blok (rumus inline `$...$` adalah teks biasa,
 * jadi tetap polos). Ini yang menjaga soal lama & grading tetap utuh.
 */
function htmlToPlain(html: string): string | null {
  const tpl = document.createElement("template");
  tpl.innerHTML = html.trim();
  const node = [...tpl.content.childNodes].filter(
    (n) => !(n.nodeType === Node.TEXT_NODE && (n.textContent ?? "").trim() === ""),
  );
  if (node.length === 0) return "";
  const paragraf: string[] = [];
  for (const n of node) {
    if (n.nodeType !== Node.ELEMENT_NODE || (n as Element).tagName !== "P") return null;
    let teks = "";
    for (const c of [...n.childNodes]) {
      if (c.nodeType === Node.TEXT_NODE) teks += c.textContent ?? "";
      else if (c.nodeType === Node.ELEMENT_NODE && (c as Element).tagName === "BR") teks += "\n";
      else return null;
    }
    paragraf.push(teks);
  }
  return paragraf.join("\n\n");
}

export interface RichEditorProps {
  value: string;
  onChange: (value: string) => void;
  disabled?: boolean;
  /** Label untuk aksesibilitas + placeholder default. */
  label?: string;
  placeholder?: string;
  /** Mode ringkas (opsi jawaban, pernyataan, pasangan): tanpa heading/tabel/video/gambar. */
  compact?: boolean;
  /** Tampilkan tombol rumus Σ (default true). */
  showMath?: boolean;
  /** Tampilkan tombol video (default !compact). */
  showVideo?: boolean;
  /** Tampilkan tombol gambar (default !compact). */
  showImage?: boolean;
  /** Tampilkan tombol tabel (default !compact). */
  showTable?: boolean;
  /** Tampilkan tombol heading (default !compact). */
  showHeading?: boolean;
  /** Tampilkan tombol kode (default true). */
  showCode?: boolean;
  /** Tampilkan tombol pratinjau (default true). */
  showPreview?: boolean;
}

const CONTOH_RUMUS: { label: string; latex: string }[] = [
  { label: "Pecahan", latex: "\\frac{a}{b}" },
  { label: "Pangkat", latex: "x^{2}" },
  { label: "Akar", latex: "\\sqrt{x + 1}" },
  { label: "Integral", latex: "\\int_{0}^{1} x^2\\,dx" },
  { label: "Sigma", latex: "\\sum_{i=1}^{n} i" },
  { label: "Limit", latex: "\\lim_{x \\to 0} \\frac{\\sin x}{x}" },
];

export default function RichEditor({
  value,
  onChange,
  disabled = false,
  label = "Isi",
  placeholder,
  compact = false,
  showMath = true,
  showVideo,
  showImage,
  showTable,
  showHeading,
  showCode = true,
  showPreview = true,
}: RichEditorProps) {
  const tampilVideo = showVideo ?? !compact;
  const tampilGambar = showImage ?? !compact;
  const tampilTabel = showTable ?? !compact;
  const tampilHeading = showHeading ?? !compact;

  const [panel, setPanel] = useState<"tidak" | "rumus" | "video" | "gambar">("tidak");
  const [lihatPratinjau, setLihatPratinjau] = useState(false);
  const [latex, setLatex] = useState("");
  const [modeRumus, setModeRumus] = useState<"inline" | "blok">("inline");
  const [urlVideo, setUrlVideo] = useState("");
  const [errorVideo, setErrorVideo] = useState("");
  const [htmlPratinjau, setHtmlPratinjau] = useState("");

  // Nilai terakhir yang dikirim ke parent — dipakai untuk mendeteksi
  // perubahan dari luar (ganti soal) vs perubahan dari editor sendiri.
  const terkirimRef = useRef(value);

  const editor = useEditor({
    extensions: [
      StarterKit.configure({ heading: { levels: [2, 3] } }),
      Underline,
      Image.configure({ inline: false, allowBase64: false }),
      Table.configure({ resizable: false }),
      TableRow,
      TableHeader,
      TableCell,
      Placeholder.configure({ placeholder: placeholder ?? `Tulis ${label.toLowerCase()}…` }),
      VideoEmbed,
    ],
    content: plainToHtml(value),
    editable: !disabled,
    editorProps: {
      attributes: {
        class: "rich-editor-content",
        "aria-label": label,
      },
    },
    onUpdate: ({ editor: ed }) => {
      const html = ed.getHTML();
      setHtmlPratinjau(html);
      const berikutnya = htmlToPlain(html) ?? html;
      terkirimRef.current = berikutnya;
      onChange(berikutnya);
    },
  });

  useEffect(() => {
    editor?.setEditable(!disabled);
  }, [disabled, editor]);

  // Sinkronisasi perubahan dari luar (mis. ganti soal di modal editor).
  useEffect(() => {
    if (!editor) return;
    if (terkirimRef.current === value) return;
    const htmlSekarang = editor.getHTML();
    const nilaiSekarang = htmlToPlain(htmlSekarang) ?? htmlSekarang;
    if (nilaiSekarang === value) {
      terkirimRef.current = value;
      return;
    }
    editor.commands.setContent(plainToHtml(value));
    setHtmlPratinjau(editor.getHTML());
    terkirimRef.current = value;
  }, [value, editor]);

  const pratinjauRumus = useMemo(() => {
    if (latex.trim() === "") return null;
    try {
      return katex.renderToString(latex, {
        displayMode: modeRumus === "blok",
        throwOnError: true,
        trust: false,
        strict: false,
      });
    } catch (err) {
      return `ERROR:${err instanceof Error ? err.message : "Rumus tidak valid."}`;
    }
  }, [latex, modeRumus]);

  const sisipRumus = () => {
    if (!editor || latex.trim() === "") return;
    const teks = modeRumus === "blok" ? `$$${latex.trim()}$$` : `$${latex.trim()}$`;
    editor.chain().focus().insertContent({ type: "text", text: teks }).run();
    setPanel("tidak");
    setLatex("");
  };

  const sisipVideo = () => {
    if (!editor) return;
    const embed = embedUrlFromWatchUrl(urlVideo);
    if (!embed) {
      setErrorVideo("Link tidak dikenali. Tempel link YouTube (watch / youtu.be / shorts) atau Vimeo.");
      return;
    }
    const judul = embed.includes("youtube") ? "Video YouTube" : "Video Vimeo";
    editor
      .chain()
      .focus()
      .insertContent({ type: "videoEmbed", attrs: { src: embed, title: judul } })
      .run();
    setPanel("tidak");
    setUrlVideo("");
    setErrorVideo("");
  };

  const btnToolbar = (
    aktif: boolean,
    judul: string,
    onClick: () => void,
    anak: React.ReactNode,
    nonaktif?: boolean,
  ) => (
    <Button
      type={aktif ? "primary" : "text"}
      size="small"
      title={judul}
      aria-label={judul}
      aria-pressed={aktif}
      disabled={disabled || nonaktif}
      onClick={onClick}
      className="flex min-h-[36px] min-w-[36px] items-center justify-center px-1.5"
    >
      {anak}
    </Button>
  );

  const dalamTabel = editor?.isActive("table") ?? false;

  return (
    <div className="overflow-hidden rounded-lg border border-border bg-white focus-within:border-anbk focus-within:ring-1 focus-within:ring-anbk">
      {/* Toolbar */}
      <div role="toolbar" aria-label={`Format ${label}`} className="flex flex-wrap items-center gap-1 border-b border-border bg-surface px-2 py-1.5">
        {btnToolbar(editor?.isActive("bold") ?? false, "Tebal (Ctrl+B)", () => editor?.chain().focus().toggleBold().run(), <Bold size={16} />)}
        {btnToolbar(editor?.isActive("italic") ?? false, "Miring (Ctrl+I)", () => editor?.chain().focus().toggleItalic().run(), <Italic size={16} />)}
        {btnToolbar(editor?.isActive("underline") ?? false, "Garis bawah (Ctrl+U)", () => editor?.chain().focus().toggleUnderline().run(), <UnderlineIcon size={16} />)}
        {btnToolbar(editor?.isActive("strike") ?? false, "Coret", () => editor?.chain().focus().toggleStrike().run(), <Strikethrough size={16} />)}
        <span className="mx-1 h-6 w-px bg-surface" aria-hidden />
        {btnToolbar(editor?.isActive("bulletList") ?? false, "Daftar bullet", () => editor?.chain().focus().toggleBulletList().run(), <List size={16} />)}
        {btnToolbar(editor?.isActive("orderedList") ?? false, "Daftar nomor", () => editor?.chain().focus().toggleOrderedList().run(), <ListOrdered size={16} />)}
        {tampilHeading && (
          <>
            {btnToolbar(editor?.isActive("heading", { level: 2 }) ?? false, "Judul 2", () => editor?.chain().focus().toggleHeading({ level: 2 }).run(), <Heading2 size={16} />)}
            {btnToolbar(editor?.isActive("heading", { level: 3 }) ?? false, "Judul 3", () => editor?.chain().focus().toggleHeading({ level: 3 }).run(), <Heading3 size={16} />)}
          </>
        )}
        {showCode && (
          <>
            {btnToolbar(editor?.isActive("codeBlock") ?? false, "Blok kode", () => editor?.chain().focus().toggleCodeBlock().run(), <Braces size={16} />)}
            {btnToolbar(editor?.isActive("code") ?? false, "Kode inline", () => editor?.chain().focus().toggleCode().run(), <Code2 size={16} />)}
          </>
        )}
        <span className="mx-1 h-6 w-px bg-surface" aria-hidden />
        {tampilTabel && btnToolbar(false, "Sisipkan tabel 3×3", () => editor?.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run(), <TableIcon size={16} />)}
        {tampilGambar && btnToolbar(panel === "gambar", "Sisipkan gambar (upload)", () => setPanel(panel === "gambar" ? "tidak" : "gambar"), <ImagePlus size={16} />)}
        {tampilVideo && btnToolbar(panel === "video", "Sisipkan video (YouTube / Vimeo)", () => setPanel(panel === "video" ? "tidak" : "video"), <PlaySquare size={16} />)}
        {showMath && btnToolbar(panel === "rumus", "Sisipkan rumus matematika (Σ)", () => setPanel(panel === "rumus" ? "tidak" : "rumus"), <span className="font-serif text-base italic">Σ</span>)}
        <span className="mx-1 h-6 w-px bg-surface" aria-hidden />
        {btnToolbar(false, "Hapus format", () => editor?.chain().focus().unsetAllMarks().clearNodes().run(), <Eraser size={16} />)}
        {btnToolbar(false, "Urungkan (Ctrl+Z)", () => editor?.chain().focus().undo().run(), <Undo2 size={16} />, !(editor?.can().undo() ?? false))}
        {btnToolbar(false, "Ulangi (Ctrl+Y)", () => editor?.chain().focus().redo().run(), <Redo2 size={16} />, !(editor?.can().redo() ?? false))}
        {showPreview && (
          <Button
            type={lihatPratinjau ? "primary" : "text"}
            size="small"
            title="Pratinjau hasil render (sanitasi + KaTeX)"
            aria-pressed={lihatPratinjau}
            disabled={disabled}
            onClick={() => {
              setHtmlPratinjau(editor?.getHTML() ?? "");
              setLihatPratinjau(!lihatPratinjau);
            }}
            className="ml-auto inline-flex min-h-[36px] items-center gap-1 px-2"
            icon={lihatPratinjau ? <EyeOff size={15} /> : <Eye size={15} />}
          >
            Pratinjau
          </Button>
        )}
      </div>

      {/* Kontrol tabel (hanya saat kursor di dalam tabel) */}
      {dalamTabel && tampilTabel && (
        <div className="flex flex-wrap items-center gap-1 border-b border-border bg-primary-light px-2 py-1.5">
          <span className="text-xs font-semibold text-muted">Tabel:</span>
          {btnToolbar(false, "Tambah baris di bawah", () => editor?.chain().focus().addRowAfter().run(), <><Plus size={14} /> Baris</>)}
          {btnToolbar(false, "Tambah kolom di kanan", () => editor?.chain().focus().addColumnAfter().run(), <><Plus size={14} /> Kolom</>)}
          {btnToolbar(false, "Hapus baris", () => editor?.chain().focus().deleteRow().run(), <><Minus size={14} /> Baris</>)}
          {btnToolbar(false, "Hapus kolom", () => editor?.chain().focus().deleteColumn().run(), <><Minus size={14} /> Kolom</>)}
          {btnToolbar(false, "Hapus tabel", () => editor?.chain().focus().deleteTable().run(), <><Trash2 size={14} /> Tabel</>)}
        </div>
      )}

      {/* Panel rumus */}
      {panel === "rumus" && (
        <div className="border-b border-border bg-white px-3 py-3">
          <div className="flex items-center justify-between">
            <p className="text-sm font-semibold text-ink">Sisipkan rumus matematika (LaTeX)</p>
            <Button type="text" size="small" aria-label="Tutup panel rumus" icon={<X size={16} />} onClick={() => setPanel("tidak")} />
          </div>
          <div className="mt-2 flex flex-col gap-2 sm:flex-row">
            <Input
              value={latex}
              disabled={disabled}
              onChange={(e) => setLatex(e.target.value)}
              placeholder="Contoh: \\frac{a}{b} atau x^{2} + \\sqrt{y}"
              aria-label="Rumus LaTeX (tanpa tanda $)"
              className="min-h-[44px] flex-1 font-mono"
              size="large"
            />
            <div className="flex gap-2" role="radiogroup" aria-label="Jenis rumus">
              <Radio.Group
                value={modeRumus}
                onChange={(e) => setModeRumus(e.target.value)}
                optionType="button"
                buttonStyle="solid"
                options={[
                  { label: "Inline $…$", value: "inline" },
                  { label: "Blok $$…$$", value: "blok" },
                ]}
              />
            </div>
            <Button type="primary" disabled={disabled || latex.trim() === ""} onClick={sisipRumus}>
              Sisipkan
            </Button>
          </div>
          <div className="mt-2 flex flex-wrap gap-1.5">
            <span className="self-center text-xs text-muted">Contoh cepat:</span>
            {CONTOH_RUMUS.map((c) => (
              <Button key={c.label} size="small" disabled={disabled} onClick={() => setLatex(c.latex)}>
                {c.label}
              </Button>
            ))}
          </div>
          <div className="mt-2 min-h-[44px] rounded-lg bg-surface px-3 py-2" aria-live="polite">
            {latex.trim() === "" ? (
              <span className="text-sm text-muted">Pratinjau rumus muncul di sini saat Anda mengetik.</span>
            ) : pratinjauRumus?.startsWith("ERROR:") ? (
              <Typography.Text type="danger">Rumus belum valid: {pratinjauRumus.slice("ERROR:".length)}</Typography.Text>
            ) : (
              <span dangerouslySetInnerHTML={{ __html: pratinjauRumus ?? "" }} />
            )}
          </div>
          <p className="mt-1 text-xs text-muted">
            Tulis LaTeX tanpa tanda $. Rumus disimpan sebagai teks LaTeX (bukan gambar) — tetap bisa dinilai & dicari.
          </p>
        </div>
      )}

      {/* Panel video */}
      {panel === "video" && (
        <div className="border-b border-border bg-white px-3 py-3">
          <div className="flex items-center justify-between">
            <p className="text-sm font-semibold text-ink">Sisipkan video (YouTube / Vimeo)</p>
            <Button type="text" size="small" aria-label="Tutup panel video" icon={<X size={16} />} onClick={() => setPanel("tidak")} />
          </div>
          <div className="mt-2 flex flex-col gap-2 sm:flex-row">
            <Input
              value={urlVideo}
              disabled={disabled}
              onChange={(e) => { setUrlVideo(e.target.value); setErrorVideo(""); }}
              placeholder="Tempel link YouTube / Vimeo… (contoh: https://www.youtube.com/watch?v=…)"
              aria-label="Link video YouTube atau Vimeo"
              className="min-h-[44px] flex-1"
              size="large"
            />
            <Button type="primary" disabled={disabled || urlVideo.trim() === ""} onClick={sisipVideo}>
              Sisipkan video
            </Button>
          </div>
          {errorVideo !== "" && <Alert type="error" showIcon message={errorVideo} className="mt-1" />}
          <p className="mt-1 text-xs text-muted">
            Video TIDAK diunggah ke server — hanya link embed yang disimpan. Video tampil di pratinjau & halaman hasil.
          </p>
        </div>
      )}

      {/* Panel gambar — memakai ImageUpload Tim D (upload + pratinjau + validasi). */}
      {panel === "gambar" && (
        <div className="border-b border-border bg-white px-3 py-3">
          <div className="flex items-center justify-between">
            <p className="text-sm font-semibold text-ink">Sisipkan gambar</p>
            <Button type="text" size="small" aria-label="Tutup panel gambar" icon={<X size={16} />} onClick={() => setPanel("tidak")} />
          </div>
          <div className="mt-2">
            <ImageUpload
              label="Gambar untuk disisipkan ke teks"
              value=""
              disabled={disabled}
              onChange={(url) => {
                if (url !== "") {
                  editor?.chain().focus().setImage({ src: url }).run();
                  setPanel("tidak");
                }
              }}
            />
          </div>
        </div>
      )}

      {/* Area edit */}
      <EditorContent editor={editor} />

      {/* Pratinjau hasil render */}
      {lihatPratinjau && (
        <div className="border-t border-border bg-surface px-3 py-3">
          <p className="mb-1 text-xs font-semibold tracking-wide text-muted uppercase">Pratinjau (seperti yang dilihat guru/siswa)</p>
          <div className="rounded-lg border border-border bg-white px-3 py-2">
            <RichContent html={htmlPratinjau} />
          </div>
        </div>
      )}
    </div>
  );
}
