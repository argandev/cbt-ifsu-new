// RichContent — render konten soal (stimulus / pertanyaan / pembahasan /
// opsi) yang tersimpan sebagai HTML ATAU teks polos (Tim E).
//
// KONTRAK PENTING:
// - Input boleh teks polos lama (tanpa tag HTML) ATAU HTML dari RichEditor.
//   Teks polos tetap tampil persis seperti sebelumnya (baris baru -> <br>),
//   sehingga soal lama TIDAK berubah tampilannya.
// - Rumus disimpan sebagai LaTeX di dalam teks: `$...$` (inline) atau
//   `$$...$$` (blok). Di-render KaTeX di sini. Bila KaTeX gagal parse,
//   LaTeX mentah ditampilkan (bukan halaman kosong / error merah).
// - HTML TIDAK PERNAH di-render mentah: selalu lewat DOMPurify dulu.
//   Tag/attr berbahaya (script, iframe, on*, form, ...) dibuang.
//   Video TIDAK disimpan sebagai <iframe>: editor menyimpan
//   `<div data-video-embed="...embed-url...">`, komponen ini yang mengubahnya
//   menjadi <iframe> — dan hanya untuk URL embed YouTube/Vimeo yang
//   tervalidasi (daftar putih ketat). URL lain -> tautan biasa.
// - TIDAK ada dependensi CDN: KaTeX + font-nya di-bundle lokal oleh Vite.
//
// Dipakai: components/soal/SoalRenderer (pratinjau guru) + panel pratinjau
// di RichEditor. Tim halaman siswa dapat memakai komponen ini juga untuk
// menampilkan pertanyaan/opsi rich (lihat CONTOH-SOAL-MATEMATIKA.md).
import { useMemo } from "react";
import DOMPurify from "dompurify";
import katex from "katex";
import "katex/dist/katex.min.css";

const TAG_DIIZINKAN = [
  "p", "br", "strong", "b", "em", "i", "u", "s", "strike", "sub", "sup",
  "ul", "ol", "li", "h1", "h2", "h3", "h4", "blockquote", "pre", "code",
  "table", "thead", "tbody", "tfoot", "tr", "th", "td",
  "img", "a", "span", "div", "hr",
];
const ATTR_DIIZINKAN = [
  "href", "target", "rel", "src", "alt", "title", "colspan", "rowspan",
  "start", "data-video-embed", "data-video-title",
];

// URL embed yang BOLEH dijadikan <iframe>: hanya host + pola ini.
const EMBED_YOUTUBE = /^https:\/\/www\.youtube-nocookie\.com\/embed\/[A-Za-z0-9_-]{11}$/;
const EMBED_VIMEO = /^https:\/\/player\.vimeo\.com\/video\/\d+$/;

export function isEmbedVideoUrl(url: string): boolean {
  return EMBED_YOUTUBE.test(url) || EMBED_VIMEO.test(url);
}

/**
 * Ubah URL tontonan YouTube/Vimeo (watch / youtu.be / shorts / embed /
 * player) menjadi URL embed yang aman dipakai <iframe>. Mengembalikan
 * null bila bukan URL video yang dikenal.
 */
export function embedUrlFromWatchUrl(url: string): string | null {
  const bersih = url.trim();
  if (bersih === "") return null;
  if (isEmbedVideoUrl(bersih)) return bersih;
  let m = bersih.match(
    /(?:youtube\.com\/(?:watch\?[^#]*v=|embed\/|shorts\/|v\/)|youtu\.be\/)([A-Za-z0-9_-]{11})/,
  );
  if (m) return `https://www.youtube-nocookie.com/embed/${m[1]}`;
  m = bersih.match(/(?:vimeo\.com\/(?:video\/)?|player\.vimeo\.com\/video\/)(\d+)/);
  if (m) return `https://player.vimeo.com/video/${m[1]}`;
  return null;
}

/** Apakah string ini mengandung tag HTML (konten rich) atau teks polos? */
export function isRichHtml(value: string): boolean {
  return /<\/?(p|div|ul|ol|li|table|strong|em|u|h[1-4]|blockquote|pre|code|img|span|br|a)\b[^>]*>/i.test(value);
}

export function escapeHtml(text: string): string {
  return text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

/** Ambil teks polos dari HTML (untuk label Dropdown, kunci teks, dll). */
export function stripHtmlToText(value: string): string {
  if (value === "") return "";
  if (!isRichHtml(value)) return value;
  const tpl = document.createElement("template");
  tpl.innerHTML = value;
  return tpl.content.textContent ?? "";
}

const POLA_MATEMATIKA = /\$\$([\s\S]+?)\$\$|\$([^$\n]+?)\$/g;

/** Render satu segmen LaTeX; gagal parse -> kembalikan teks mentah (fallback). */
function renderLatex(latex: string, displayMode: boolean): string {
  try {
    return katex.renderToString(latex, {
      displayMode,
      throwOnError: true,
      trust: false,
      strict: false,
    });
  } catch {
    return escapeHtml(displayMode ? `$$${latex}$$` : `$${latex}$`);
  }
}

/**
 * Sanitasi + render rumus + ubah penanda video menjadi iframe.
 * Mengembalikan HTML AMAN yang boleh dipakai di dangerouslySetInnerHTML.
 */
export function renderRichHtml(input: string): string {
  if (input === "") return "";
  // Teks polos: escape dulu, baris baru -> <br>, lalu jalurnya sama dengan HTML.
  const sumber = isRichHtml(input)
    ? input
    : escapeHtml(input).replace(/\r\n/g, "\n").replace(/\n/g, "<br>");
  const bersih = DOMPurify.sanitize(sumber, {
    ALLOWED_TAGS: TAG_DIIZINKAN,
    ALLOWED_ATTR: ATTR_DIIZINKAN,
    FORBID_TAGS: ["script", "style", "form", "input", "button", "iframe", "object", "embed", "link", "meta"],
    FORBID_ATTR: ["onerror", "onclick", "onload", "style"],
    ALLOW_DATA_ATTR: false,
  });

  const tpl = document.createElement("template");
  tpl.innerHTML = bersih;

  // 1) Div penanda video -> <iframe> (hanya URL embed tervalidasi).
  const penandaVideo = [...tpl.content.querySelectorAll("div[data-video-embed]")];
  for (const el of penandaVideo) {
    const src = el.getAttribute("data-video-embed") ?? "";
    const title = el.getAttribute("data-video-title") ?? "Video";
    if (isEmbedVideoUrl(src)) {
      const frame = document.createElement("iframe");
      frame.setAttribute("src", src);
      frame.setAttribute("title", title);
      frame.setAttribute("allow", "accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share");
      frame.setAttribute("allowFullscreen", "");
      frame.setAttribute("loading", "lazy");
      frame.className = "rich-video-frame";
      el.replaceWith(frame);
    } else {
      // URL tidak tervalidasi -> jangan render iframe; tampilkan judulnya saja.
      el.replaceWith(document.createTextNode(`Video: ${title}`));
    }
  }

  // 2) Rumus $...$ / $$...$$ di dalam text node -> HTML KaTeX.
  const walker = document.createTreeWalker(tpl.content, NodeFilter.SHOW_TEXT);
  const nodeTeks: Text[] = [];
  while (walker.nextNode()) nodeTeks.push(walker.currentNode as Text);
  for (const node of nodeTeks) {
    const teks = node.nodeValue ?? "";
    if (!teks.includes("$")) continue;
    POLA_MATEMATIKA.lastIndex = 0;
    let cocok: RegExpExecArray | null;
    let terakhir = 0;
    const fragmen = document.createDocumentFragment();
    let adaRumus = false;
    while ((cocok = POLA_MATEMATIKA.exec(teks)) !== null) {
      adaRumus = true;
      if (cocok.index > terakhir) {
        fragmen.appendChild(document.createTextNode(teks.slice(terakhir, cocok.index)));
      }
      const blok = cocok[1] !== undefined;
      const latex = (blok ? cocok[1] : cocok[2]).trim();
      const span = document.createElement("span");
      // renderLatex mengembalikan HTML KaTeX (dari library, bukan input user)
      // atau teks mentah yang sudah di-escape bila parse gagal.
      span.innerHTML = renderLatex(latex, blok);
      fragmen.appendChild(span);
      terakhir = cocok.index + cocok[0].length;
    }
    if (!adaRumus) continue;
    if (terakhir < teks.length) {
      fragmen.appendChild(document.createTextNode(teks.slice(terakhir)));
    }
    node.replaceWith(fragmen);
  }

  // 3) Gambar: lazy-load + batas ukuran lewat CSS (.rich-content img).
  for (const img of [...tpl.content.querySelectorAll("img")]) {
    img.setAttribute("loading", "lazy");
    if (!img.getAttribute("alt")) img.setAttribute("alt", "Gambar soal");
  }

  return tpl.innerHTML;
}

interface RichContentProps {
  /** HTML dari RichEditor ATAU teks polos lama. Kosong -> tidak render apa-apa. */
  html?: string | null;
  className?: string;
}

/** Render konten soal yang sudah disanitasi + rumus KaTeX. */
export default function RichContent({ html, className }: RichContentProps) {
  const aman = useMemo(() => renderRichHtml(html ?? ""), [html]);
  if (!html) return null;
  return (
    <div
      className={`rich-content ${className ?? ""}`}
      dangerouslySetInnerHTML={{ __html: aman }}
    />
  );
}
