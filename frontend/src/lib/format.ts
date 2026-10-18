// lib/format.ts — helper PUSAT format tanggal/jam tampilan (hanya tampilan,
// TIDAK mengubah nilai yang dikirim ke backend, TIDAK mengubah input
// type="date"/"time" yang tetap memakai format mentah YYYY-MM-DD / HH:MM).
//
// Aturan:
// - Semua tanggal tampil dalam format Indonesia, zona Asia/Jakarta.
// - Nama bulan SINGKAT: Jan, Feb, Mar, Apr, Mei, Jun, Jul, Agu, Sep, Okt, Nov, Des.
// - Input kosong / tidak valid -> "—" (bukan string mentah, bukan "Invalid Date").
//
// Perilaku per bentuk input:
// - "2026-10-16" (tanggal saja) -> dipakai apa adanya, tanpa geser zona waktu.
// - "2026-10-16T22:31:00" (naif, tanpa offset) -> dianggap jam dinding WIB
//   (backend menyimpan jam dinding WIB), dipakai apa adanya.
// - "2026-10-16T15:31:00Z" / "... +07:00" (ber-offset) -> dikonversi ke WIB
//   via Intl.DateTimeFormat('id-ID', { timeZone: 'Asia/Jakarta', ... }).
// - "22:31:00" / "22:31" (jam saja) -> "22:31".

const KOSONG = "—";

const BULAN_PENDEK = [
  "Jan", "Feb", "Mar", "Apr", "Mei", "Jun",
  "Jul", "Agu", "Sep", "Okt", "Nov", "Des",
];

type BagianTanggal = { tahun: number; bulan: number; tanggal: number };
type BagianJam = { jam: number; menit: number };

const RE_TANGGAL = /^(\d{4})-(\d{2})-(\d{2})$/;
const RE_TANGGAL_JAM_NAIF = /^(\d{4})-(\d{2})-(\d{2})[T ](\d{2}):(\d{2})(?::(\d{2}))?$/;
const RE_JAM = /^(\d{2}):(\d{2})(?::(\d{2}))?$/;

function validTanggal(t: BagianTanggal): boolean {
  return (
    Number.isInteger(t.bulan) && t.bulan >= 1 && t.bulan <= 12 &&
    Number.isInteger(t.tanggal) && t.tanggal >= 1 && t.tanggal <= 31 &&
    Number.isInteger(t.tahun) && t.tahun >= 0
  );
}

function validJam(j: BagianJam): boolean {
  return j.jam >= 0 && j.jam <= 23 && j.menit >= 0 && j.menit <= 59;
}

function fmtTanggalDari(t: BagianTanggal): string {
  if (!validTanggal(t)) return KOSONG;
  return `${t.tanggal} ${BULAN_PENDEK[t.bulan - 1]} ${t.tahun}`;
}

function fmtJamDari(j: BagianJam): string {
  if (!validJam(j)) return KOSONG;
  return `${String(j.jam).padStart(2, "0")}:${String(j.menit).padStart(2, "0")}`;
}

// Ambil bagian tanggal (+ jam bila ada) dari string ISO apa pun.
// Mengembalikan null bila input kosong / tidak bisa diparse.
function uraikan(
  iso: string | null | undefined,
): { tanggal: BagianTanggal; jam: BagianJam | null } | null {
  if (iso === null || iso === undefined) return null;
  const s = String(iso).trim();
  if (s === "") return null;

  const mTanggal = s.match(RE_TANGGAL);
  if (mTanggal) {
    const tanggal = { tahun: Number(mTanggal[1]), bulan: Number(mTanggal[2]), tanggal: Number(mTanggal[3]) };
    return validTanggal(tanggal) ? { tanggal, jam: null } : null;
  }

  const mNaif = s.match(RE_TANGGAL_JAM_NAIF);
  if (mNaif) {
    const tanggal = { tahun: Number(mNaif[1]), bulan: Number(mNaif[2]), tanggal: Number(mNaif[3]) };
    const jam = { jam: Number(mNaif[4]), menit: Number(mNaif[5]) };
    return validTanggal(tanggal) && validJam(jam) ? { tanggal, jam } : null;
  }

  // Ber-offset / bentuk lain: serahkan ke Date + konversi WIB via Intl.
  const d = new Date(s);
  if (Number.isNaN(d.getTime())) return null;
  const parts = new Intl.DateTimeFormat("id-ID", {
    timeZone: "Asia/Jakarta",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    hourCycle: "h23",
  }).formatToParts(d);
  const ambil = (type: Intl.DateTimeFormatPartTypes): number => {
    const p = parts.find((x) => x.type === type);
    return p ? Number(p.value) : NaN;
  };
  const tanggal = { tahun: ambil("year"), bulan: ambil("month"), tanggal: ambil("day") };
  const jam = { jam: ambil("hour"), menit: ambil("minute") };
  // Intl id-ID kadang mengembalikan jam "24" untuk tengah malam.
  if (jam.jam === 24) jam.jam = 0;
  return validTanggal(tanggal) && validJam(jam) ? { tanggal, jam } : null;
}

/** "2026-10-16" / "2026-10-16T22:31:00Z" -> "16 Okt 2026". Kosong/invalid -> "—". */
export function formatTanggal(iso: string | null | undefined): string {
  const u = uraikan(iso);
  return u ? fmtTanggalDari(u.tanggal) : KOSONG;
}

/**
 * "2026-10-16T22:31:00Z" -> "16 Okt 2026, 22:31 WIB".
 * Input tanggal saja (tanpa jam) -> hanya tanggal, TANPA jam yang dikarang
 * ("2026-10-16" -> "16 Okt 2026"). Kosong/invalid -> "—".
 */
export function formatTanggalJam(iso: string | null | undefined): string {
  const u = uraikan(iso);
  if (!u) return KOSONG;
  const tgl = fmtTanggalDari(u.tanggal);
  if (!u.jam) return tgl;
  return `${tgl}, ${fmtJamDari(u.jam)} WIB`;
}

/**
 * "22:31:00" / "22:31" / "2026-10-16T22:31:00Z" -> "22:31".
 * Input tanggal saja (tanpa jam) -> "—". Kosong/invalid -> "—".
 */
export function formatJam(iso: string | null | undefined): string {
  if (iso === null || iso === undefined) return KOSONG;
  const s = String(iso).trim();
  if (s === "") return KOSONG;
  const mJam = s.match(RE_JAM);
  if (mJam) {
    const jam = { jam: Number(mJam[1]), menit: Number(mJam[2]) };
    return validJam(jam) ? fmtJamDari(jam) : KOSONG;
  }
  const u = uraikan(s);
  return u?.jam ? fmtJamDari(u.jam) : KOSONG;
}
