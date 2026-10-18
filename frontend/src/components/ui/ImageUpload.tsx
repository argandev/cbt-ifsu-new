// ImageUpload — upload gambar asli: drag-drop + klik pilih file + pratinjau
// + indikator progress + pesan error jelas + tombol hapus/ganti (Tim D).
//
// CARA PAKAI (dipakai SettingsPage untuk logo; Tim E memakai di SoalEditor
// untuk media soal & gambar opsi `pilihan_gambar` — TIDAK perlu ubah
// komponen ini, cukup pasang seperti contoh):
//
// ```tsx
// import ImageUpload from "../../components/ui/ImageUpload";
//
// // 1) Paling sederhana — upload langsung ke endpoint bawaan,
// //    hasilnya (URL internal) masuk ke state lewat onChange:
// <ImageUpload
//   label="Media soal"
//   value={soal.media_url}
//   onChange={(url) => setSoal({ ...soal, media_url: url })}
// />
//
// // 2) Upload kustom (mis. perlu field tambahan / endpoint lain):
// //    onUpload WAJIB mengembalikan Promise<string> berisi URL hasil upload;
// //    komponen lalu memanggil onChange(url) dengan URL tersebut.
// <ImageUpload
//   label="Gambar opsi"
//   value={opsi.gambar}
//   onChange={(url) => setOpsi({ ...opsi, gambar: url })}
//   onUpload={async (file) => uploadKeEndpointLain(file)} // => Promise<string>
// />
// ```
//
// KONTRAK PROPS (jangan ditebak — ini API resminya):
// - `value: string | null | undefined` — URL gambar saat ini ("" / null =
//   belum ada gambar). Bisa URL internal hasil upload
//   ("/api/v1/upload/image/<uuid>.png") maupun URL lama yang sudah tersimpan
//   (soal lama tetap tampil — komponen hanya menampilkan, bukan memvalidasi
//   ulang URL lama).
// - `onChange(url: string)` — dipanggil dengan URL baru SETELAH upload
//   berhasil, atau "" saat pengguna menekan Hapus. Pemanggil cukup simpan
//   string ini ke field-nya (logo_url / media soal / gambar opsi).
// - `onUpload?(file: File) => Promise<string>` — OPSIONAL. Bila diisi,
//   komponen memakai fungsi ini untuk mengunggah (bukan endpoint bawaan).
//   Fungsi HARUS resolve ke URL gambar (string). Bila reject, pesan error
//   dari Error-nya ditampilkan. Bila TIDAK diisi, komponen mengunggah
//   langsung ke `POST /api/v1/upload/image` (multipart field "file",
//   session cookie + header X-CSRF-Token dari cookie `cbt_csrf` — pola yang
//   sama dengan upload berkas UKK) dan progress dihitung dari XHR.
// - `disabled?: boolean` — nonaktifkan seluruh interaksi (dipakai saat form
//   induk sedang menyimpan).
// - `label?: string` — label di atas area upload (default "Gambar").
// - `maxSizeMB?: number` — batas ukuran di client (default 5, sama dengan
//   server; server tetap menegakkan batasnya sendiri — validasi client
//   hanya untuk pesan error lebih dini).
//
// VALIDASI (client, sebelum upload — server memvalidasi ulang semuanya):
// hanya PNG / JPEG / WebP / GIF, maks `maxSizeMB` MB. File tidak valid
// ditolak dengan pesan error di bawah area upload, TIDAK diunggah.
// File lama yang sudah tersimpan sebagai URL eksternal TIDAK pernah
// diunggah ulang — komponen hanya menampilkan & menggantinya bila pengguna
// memilih file baru. TIDAK ADA dependensi baru (React + antd saja).
//
// CHROME REDESIGN ANT DESIGN (Tim Form): zona drag-drop + tombol memakai
// antd Upload / Upload.Dragger (validasi di beforeUpload, upload nyata di
// customRequest — endpoint, field "file", CSRF, dan progress XHR TETAP
// sama), progress memakai antd Progress, error memakai antd Alert,
// pratinjau memakai antd Image, tombol memakai antd Button.
import { useCallback, useEffect, useState } from "react";
import { Alert, Button, Image, Progress, Upload } from "antd";
import type { UploadProps } from "antd";
import { DeleteOutlined, InboxOutlined, UploadOutlined } from "@ant-design/icons";

export interface ImageUploadProps {
  value?: string | null;
  onChange: (url: string) => void;
  onUpload?: (file: File) => Promise<string>;
  disabled?: boolean;
  label?: string;
  /** Batas ukuran file di client (MB). Default 5 — sama dengan server. */
  maxSizeMB?: number;
}

const TIPE_DIIZINKAN = ["image/png", "image/jpeg", "image/webp", "image/gif"];
const EKSTENSI_DIIZINKAN = [".png", ".jpg", ".jpeg", ".webp", ".gif"];
const PESAN_TIPE =
  "File bukan gambar yang diizinkan. Gunakan PNG, JPEG, WebP, atau GIF.";
const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "";

function getCsrfToken(): string {
  const match = document.cookie.match(/(?:^|;\s*)cbt_csrf=([^;]*)/);
  return match ? decodeURIComponent(match[1]) : "";
}

/** Pesan error API kontrak {"error": {"message"}} → string yang bisa dibaca. */
function pesanErrorApi(json: unknown, fallback: string): string {
  if (typeof json === "object" && json !== null) {
    const err = (json as { error?: { message?: unknown } }).error;
    if (err && typeof err.message === "string" && err.message !== "") return err.message;
  }
  return fallback;
}

/**
 * Upload bawaan ke POST /api/v1/upload/image via XMLHttpRequest agar
 * progress unggah bisa dilaporkan (fetch tidak punya progress upload).
 * Mengembalikan URL internal hasil upload. Progress via onProgress (0–100).
 */
function uploadBawaan(file: File, onProgress: (persen: number) => void): Promise<string> {
  return new Promise((resolve, reject) => {
    const form = new FormData();
    form.append("file", file);
    const xhr = new XMLHttpRequest();
    xhr.open("POST", `${BASE_URL}/api/v1/upload/image`, true);
    xhr.withCredentials = true;
    xhr.setRequestHeader("X-CSRF-Token", getCsrfToken());
    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable) onProgress(Math.round((e.loaded / e.total) * 100));
    };
    xhr.onload = () => {
      let json: unknown = null;
      try {
        json = xhr.responseText ? JSON.parse(xhr.responseText) : null;
      } catch {
        // Respons bukan JSON → jatuh ke pesan fallback di bawah.
      }
      if (xhr.status >= 200 && xhr.status < 300) {
        const url = (json as { data?: { url?: unknown } } | null)?.data?.url;
        if (typeof url === "string" && url !== "") {
          onProgress(100);
          resolve(url);
          return;
        }
        reject(new Error("Server tidak mengembalikan URL gambar."));
        return;
      }
      reject(new Error(pesanErrorApi(json, `Upload gagal (${xhr.status}).`)));
    };
    xhr.onerror = () => reject(new Error("Upload gagal: tidak dapat menghubungi server."));
    xhr.onabort = () => reject(new Error("Upload dibatalkan."));
    xhr.send(form);
  });
}

export default function ImageUpload({
  value,
  onChange,
  onUpload,
  disabled = false,
  label = "Gambar",
  maxSizeMB = 5,
}: ImageUploadProps) {
  const [error, setError] = useState("");
  const [mengunggah, setMengunggah] = useState(false);
  const [progress, setProgress] = useState(0);
  // Pratinjau lokal file yang baru dipilih (object URL) — diganti URL server
  // begitu upload berhasil, dibersihkan saat komponen unmount / berganti.
  const [pratinjauLokal, setPratinjauLokal] = useState<string | null>(null);

  useEffect(() => {
    return () => {
      if (pratinjauLokal) URL.revokeObjectURL(pratinjauLokal);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [pratinjauLokal]);

  const validasiFile = useCallback(
    (file: File): string | null => {
      const namaLower = file.name.toLowerCase();
      const ekstensiOk = EKSTENSI_DIIZINKAN.some((ext) => namaLower.endsWith(ext));
      if (!TIPE_DIIZINKAN.includes(file.type) && !ekstensiOk) return PESAN_TIPE;
      if (file.size <= 0) return "File kosong.";
      if (file.size > maxSizeMB * 1024 * 1024)
        return `Ukuran gambar melebihi ${maxSizeMB} MB.`;
      return null;
    },
    [maxSizeMB],
  );

  // Props Upload antd yang dipakai BERSAMA oleh Dragger (belum ada gambar)
  // dan tombol Ganti (sudah ada gambar): validasi di beforeUpload, upload
  // nyata di customRequest. showUploadList={false} — pratinjau dirender
  // sendiri di bawah agar URL lama / object URL lokal tetap tampil sama.
  const uploadProps: UploadProps = {
    accept: "image/png,image/jpeg,image/webp,image/gif",
    disabled: disabled || mengunggah,
    multiple: false,
    maxCount: 1,
    showUploadList: false,
    beforeUpload: (file) => {
      if (disabled) return Upload.LIST_IGNORE;
      const pesanTidakValid = validasiFile(file);
      if (pesanTidakValid) {
        setError(pesanTidakValid);
        return Upload.LIST_IGNORE;
      }
      setError("");
      return true;
    },
    customRequest: async (options) => {
      const { file, onProgress, onSuccess, onError } = options;
      setMengunggah(true);
      setProgress(0);
      const lokal = URL.createObjectURL(file as File);
      setPratinjauLokal(lokal);
      try {
        const url = onUpload
          ? await onUpload(file as File)
          : await uploadBawaan(file as File, (p) => {
              setProgress(p);
              onProgress?.({ percent: p });
            });
        if (typeof url !== "string" || url === "") {
          throw new Error("Upload tidak mengembalikan URL gambar.");
        }
        setPratinjauLokal(null);
        URL.revokeObjectURL(lokal);
        onSuccess?.({});
        onChange(url);
      } catch (err) {
        URL.revokeObjectURL(lokal);
        setPratinjauLokal(null);
        const pesan = err instanceof Error ? err.message : "Upload gagal.";
        setError(pesan);
        onError?.(err instanceof Error ? err : new Error(pesan));
      } finally {
        setMengunggah(false);
        setProgress(0);
      }
    },
  };

  const hapus = () => {
    if (disabled || mengunggah) return;
    setError("");
    setPratinjauLokal(null);
    onChange("");
  };

  const srcPratinjau = pratinjauLokal ?? (value && value !== "" ? value : null);

  return (
    <div>
      <p className="mb-1 block text-sm font-semibold text-ink">{label}</p>
      {srcPratinjau ? (
        <div className="overflow-hidden rounded-lg border border-border bg-surface">
          <Image
            src={srcPratinjau}
            alt={`Pratinjau ${label.toLowerCase()}`}
            preview={false}
            style={{ maxHeight: 224, width: "100%", objectFit: "contain" }}
          />
          {mengunggah ? (
            <div className="px-3 py-2">
              <Progress percent={progress} status="active" aria-label="Progress upload" />
              <p className="mt-1 text-xs text-muted">Mengunggah… {progress}%</p>
            </div>
          ) : null}
          <div className="flex flex-wrap gap-2 border-t border-border bg-white px-3 py-2">
            <Upload {...uploadProps}>
              <Button type="default" disabled={disabled || mengunggah} icon={<UploadOutlined />}>
                Ganti gambar
              </Button>
            </Upload>
            <Button
              type="default"
              danger
              disabled={disabled || mengunggah}
              onClick={hapus}
              icon={<DeleteOutlined />}
            >
              Hapus gambar
            </Button>
          </div>
        </div>
      ) : (
        <Upload.Dragger {...uploadProps}>
          {mengunggah ? (
            <>
              <p className="ant-upload-drag-icon">
                <InboxOutlined />
              </p>
              <p className="ant-upload-text">Mengunggah… {progress}%</p>
              <div className="mx-auto max-w-xs">
                <Progress percent={progress} status="active" aria-label="Progress upload" />
              </div>
            </>
          ) : (
            <>
              <p className="ant-upload-drag-icon">
                <InboxOutlined />
              </p>
              <p className="ant-upload-text">Klik untuk pilih gambar atau seret file ke sini</p>
              <p className="ant-upload-hint">PNG, JPEG, WebP, atau GIF — maks. {maxSizeMB} MB.</p>
            </>
          )}
        </Upload.Dragger>
      )}
      {error !== "" && (
        <Alert className="mt-1" type="error" showIcon message={error} role="alert" />
      )}
    </div>
  );
}
