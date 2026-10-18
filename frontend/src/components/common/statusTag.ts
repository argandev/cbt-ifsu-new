// statusTag — SATU peta status -> warna antd Tag (standar audit §2.2).
// Menggantikan: 4x `__badgeColor` duplikat, StatusBadge + PETA, Badge custom,
// dan span badge inline. Satu-satunya tempat pemetaan status -> warna.
//
// Peta (§2.2): Aktif/Lulus/Dijawab/Terbuka/Terkirim/Disetujui/Terbaca -> green;
// Menunggu/Jeda/Ragu-ragu/Terkunci -> gold; Diblokir/Tidak Lulus/Ditolak -> red;
// Berjalan/Selesai/Dikerjakan -> blue; Draft/Tidak Aktif/Belum Dijawab -> default.
// Ditambah: kesukaran (Sedang green / Mudah blue / Sukar red) dan daya beda
// (Sangat Baik green / Baik green / Cukup gold / Buruk orange / Sangat Buruk red)
// agar AnalisisButir memakai peta yang sama. Alias varian lama
// (success/answered/warning/doubt/danger/info/inactive/outline) dipertahankan
// sebagai jembatan agar pemanggil lama yang memakai nama varian tetap benar.

const PETA_WARNA: Record<string, "green" | "gold" | "red" | "blue" | "orange" | "default"> = {
  // green — positif / sukses
  aktif: "green",
  lulus: "green",
  dijawab: "green",
  "sudah-dijawab": "green",
  terbuka: "green",
  terkirim: "green",
  disetujui: "green",
  terbaca: "green",
  sedang: "green",
  "sangat-baik": "green",
  baik: "green",
  success: "green",
  answered: "green",
  // gold — menunggu / ragu-ragu / terkunci
  menunggu: "gold",
  jeda: "gold",
  pause: "gold",
  "ragu-ragu": "gold",
  ragu: "gold",
  terkunci: "gold",
  cukup: "gold",
  warning: "gold",
  doubt: "gold",
  // red — bahaya / gagal
  diblokir: "red",
  "tidak-lulus": "red",
  ditolak: "red",
  sukar: "red",
  "sangat-buruk": "red",
  danger: "red",
  // blue — berjalan / selesai / dikerjakan
  berjalan: "blue",
  selesai: "blue",
  dikerjakan: "blue",
  mudah: "blue",
  info: "blue",
  // orange — buruk (daya beda)
  buruk: "orange",
  // default — draft / tidak aktif / belum dijawab
  draft: "default",
  "tidak-aktif": "default",
  "belum-dijawab": "default",
  belum: "default",
  default: "default",
  inactive: "default",
  outline: "default",
};

const PETA_LABEL: Record<string, string> = {
  draft: "Draft",
  aktif: "Aktif",
  pause: "Jeda",
  selesai: "Selesai",
  berjalan: "Berjalan",
  diblokir: "Diblokir",
  lulus: "Lulus",
  "tidak-lulus": "Tidak Lulus",
  menunggu: "Menunggu",
  disetujui: "Disetujui",
  ditolak: "Ditolak",
  terkunci: "Terkunci",
  terbuka: "Terbuka",
  dijawab: "Dijawab",
  "sudah-dijawab": "Dijawab",
  "ragu-ragu": "Ragu-ragu",
  ragu: "Ragu-ragu",
  "belum-dijawab": "Belum Dijawab",
  belum: "Belum Dijawab",
  "tidak-aktif": "Tidak Aktif",
  dikerjakan: "Dikerjakan",
  terkirim: "Terkirim",
  terbaca: "Terbaca",
};

/** Normalisasi kunci status: kecil + pemisah "_" / spasi -> "-". */
function normalisasi(status: string): string {
  return status.trim().toLowerCase().replace(/[_\s]+/g, "-");
}

/** Warna antd Tag untuk sebuah status (default bila tak dikenal). */
export function statusTagColor(status: string): "green" | "gold" | "red" | "blue" | "orange" | "default" {
  return PETA_WARNA[normalisasi(status)] ?? "default";
}

/** Label tampil untuk sebuah status (teks mentah bila tak dikenal). */
export function statusTagLabel(status: string): string {
  return PETA_LABEL[normalisasi(status)] ?? status;
}
