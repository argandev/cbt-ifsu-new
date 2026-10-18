// Tipe data Bank Soal (Tim 3). Bentuk jawaban/kunci per tipe:
// lihat docs/TIPE-SOAL.md (jawaban = tampil ke peserta, kunci = guru saja).
export interface BankSoal {
  id: string;
  kode: string;
  nama: string;
  mapel_id: string;
  kelas_tingkat: number;
  guru_id: string | null;
  is_locked: boolean;
  jumlah_soal: number;
  created_at: string;
  updated_at: string;
}

export interface Soal {
  id: string;
  bank_id: string;
  tipe: string;
  kisi_kisi_id: string | null;
  kompetensi_id: string | null;
  stimulus: string;
  pertanyaan: string;
  media_url: string;
  jawaban: Record<string, unknown>;
  kunci?: Record<string, unknown>;
  /** Pembahasan guru (kolom soal.pembahasan, migrasi 0011). Tidak pernah dikirim ke siswa sebelum submit. */
  pembahasan?: string;
  bobot: number;
  level: "mudah" | "sedang" | "sulit";
  status: "draft" | "aktif" | "nonaktif" | "perlu_revisi" | "diarsipkan";
  created_at: string;
  updated_at: string;
}

export interface PageData<T> {
  items: T[];
  total: number;
  limit: number;
  offset: number;
}

export interface SoalFormValue {
  tipe: string;
  stimulus: string;
  pertanyaan: string;
  media_url: string;
  jawaban: Record<string, unknown>;
  kunci: Record<string, unknown>;
  /** Pembahasan guru — tampil di Tryout / bila hasil ditampilkan (kolom soal.pembahasan). */
  pembahasan?: string;
  bobot: number;
  level: "mudah" | "sedang" | "sulit";
  status: "draft" | "aktif" | "nonaktif" | "perlu_revisi" | "diarsipkan";
}

export function soalBaru(): SoalFormValue {
  return {
    tipe: "pg",
    stimulus: "",
    pertanyaan: "",
    media_url: "",
    jawaban: { opsi: [{ label: "A", teks: "" }, { label: "B", teks: "" }, { label: "C", teks: "" }, { label: "D", teks: "" }] },
    kunci: { opsi: "A" },
    pembahasan: "",
    bobot: 1,
    level: "sedang",
    status: "aktif",
  };
}

export function soalToForm(s: Soal): SoalFormValue {
  return {
    tipe: s.tipe,
    stimulus: s.stimulus,
    pertanyaan: s.pertanyaan,
    media_url: s.media_url,
    jawaban: s.jawaban ?? {},
    kunci: s.kunci ?? {},
    pembahasan: s.pembahasan ?? "",
    bobot: s.bobot,
    level: s.level,
    status: s.status,
  };
}
