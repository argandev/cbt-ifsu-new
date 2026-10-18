// apiClient — satu-satunya tempat fetch ke backend (kontrak bersama, Tim 1).
//
// Auth memakai session cookie + CSRF (BUKAN JWT): cookie dikirim via
// credentials "include". Token CSRF dibaca dari cookie `cbt_csrf` (ditulis
// backend saat login) dan dikirim sebagai header X-CSRF-Token untuk mutasi.
//
// Path: pemanggil boleh memakai path pendek ("/settings") atau lengkap
// ("/api/v1/settings") — keduanya dinormalisasi ke "/api/v1/...".
// Base URL: "" (same origin, produksi satu binary) atau VITE_API_BASE_URL
// saat dev terpisah.
//
// Retry 429 (BUG-03, Fase 3): bila server menjawab 429 dengan kode
// `rate_limited`, request diulang otomatis maks 3x dengan backoff
// eksponensial 500ms -> 1s -> 2s, atau mengikuti header `Retry-After`
// bila ada (dibatasi maks 10 detik agar UI tidak menggantung). Retry
// hanya untuk GET/HEAD + POST: 429 ditolak middleware SEBELUM handler,
// jadi request yang di-retry belum pernah dieksekusi — aman diulang
// (jawaban = upsert, submit = dicek submission existing oleh server,
// login = tanpa efek samping). PUT/DELETE tidak di-retry otomatis.
const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "";

function getCsrfToken(): string {
  const match = document.cookie.match(/(?:^|;\s*)cbt_csrf=([^;]*)/);
  return match ? decodeURIComponent(match[1]) : "";
}

function normalizePath(path: string): string {
  if (path.startsWith("/api/v1")) return path;
  if (path.startsWith("/")) return `/api/v1${path}`;
  return `/api/v1/${path}`;
}

/** Error API sesuai format kontrak {"error": {"code", "message"}}. */
export class ApiError extends Error {
  code: string;
  status: number;
  constructor(code: string, message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.status = status;
  }
}

/** Maksimum pengulangan untuk 429 `rate_limited` (tidak termasuk percobaan pertama). */
export const MAX_RETRY_429 = 3;
/** Backoff eksponensial antar-retry 429: retry ke-1, ke-2, ke-3 (ms). */
export const BACKOFF_429_MS = [500, 1000, 2000];

/**
 * Parse nilai header `Retry-After` menjadi milidetik.
 * Format yang didukung: detik ("120") atau tanggal HTTP ("Wed, 21 Oct 2015 07:28:00 GMT").
 * Mengembalikan null bila header kosong / tidak dapat diparse / sudah lewat.
 */
export function parseRetryAfter(value: string | null): number | null {
  if (value === null || value.trim() === "") return null;
  const trimmed = value.trim();
  const detik = Number(trimmed);
  if (Number.isFinite(detik) && detik >= 0) return detik * 1000;
  const tanggal = Date.parse(trimmed);
  if (Number.isFinite(tanggal)) {
    const sisa = tanggal - Date.now();
    return sisa > 0 ? sisa : null;
  }
  return null;
}

/**
 * Jeda sebelum retry 429 ke-`attempt` (0-based: 0 = retry pertama).
 * `Retry-After` menang bila ada, dibatasi maks 10 detik.
 */
export function retryDelayMs(attempt: number, retryAfter: string | null): number {
  const dariHeader = parseRetryAfter(retryAfter);
  if (dariHeader !== null) return Math.min(dariHeader, 10_000);
  return BACKOFF_429_MS[attempt] ?? BACKOFF_429_MS[BACKOFF_429_MS.length - 1];
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const method = init?.method ?? "GET";
  const isMutation = method !== "GET" && method !== "HEAD";
  // Hanya GET/HEAD + POST yang di-retry otomatis pada 429 (lihat kepala file).
  const bolehRetry429 = method === "GET" || method === "HEAD" || method === "POST";
  // FormData (upload file, mis. ImportSiswaPage) WAJIB diteruskan apa adanya:
  // jangan set Content-Type JSON agar browser menulis boundary multipart
  // sendiri. Header X-CSRF-Token tetap dikirim untuk mutasi.
  const isFormData =
    typeof FormData !== "undefined" && init?.body instanceof FormData;

  for (let attempt = 0; ; attempt++) {
    const res = await fetch(`${BASE_URL}${normalizePath(path)}`, {
      ...init,
      credentials: "include",
      headers: {
        ...(isFormData ? {} : { "Content-Type": "application/json" }),
        ...(isMutation ? { "X-CSRF-Token": getCsrfToken() } : {}),
        ...(init?.headers ?? {}),
      },
    });
    const text = await res.text();
    let json: unknown = null;
    try {
      json = text ? JSON.parse(text) : null;
    } catch {
      // Respons bukan JSON (mis. CSV) ditangani pemanggil via download.
    }
    if (res.ok) {
      return (json as { data: T }).data;
    }
    const err = (json as { error?: { code?: string; message?: string } } | null)?.error;
    const code = err?.code ?? "request_failed";
    if (
      res.status === 429 &&
      code === "rate_limited" &&
      bolehRetry429 &&
      attempt < MAX_RETRY_429
    ) {
      await sleep(retryDelayMs(attempt, res.headers.get("Retry-After")));
      continue;
    }
    throw new ApiError(
      code,
      err?.message ?? `Permintaan gagal (${res.status})`,
      res.status,
    );
  }
}

export const apiClient = {
  get: <T>(path: string) => request<T>(path, { method: "GET" }),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: "POST",
      body:
        body instanceof FormData
          ? body
          : body === undefined
            ? undefined
            : JSON.stringify(body),
    }),
  put: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: "PUT",
      body:
        body instanceof FormData
          ? body
          : body === undefined
            ? undefined
            : JSON.stringify(body),
    }),
  delete: <T>(path: string) => request<T>(path, { method: "DELETE" }),
  /** Alias lama untuk DELETE. */
  del: <T>(path: string) => request<T>(path, { method: "DELETE" }),

  /** Unduh file (mis. export CSV) lewat browser. */
  download: (path: string, filename: string) => {
    const a = document.createElement("a");
    a.href = `${BASE_URL}${normalizePath(path)}`;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    a.remove();
  },
};

export default apiClient;
