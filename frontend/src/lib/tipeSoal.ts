// Metadata 16 tipe soal — CERMINAN backend (backend/internal/banksoal/types.go)
// + docs/TIPE-SOAL.md. Bila menambah/mengubah tipe, ubah KETIGA tempat.
//
// Label JUJUR: "5 tipe standar AKM + 11 varian/ekstensi aplikasi".
// JANGAN PERNAH menulis "16 tipe standar Kemendikbud" di UI/dokumen.
export interface TipeInfo {
  kode: string;
  label: string;
  kelompok: "inti" | "varian";
  penilaian: "auto" | "manual" | "non-nilai";
  dinilai: boolean;
  deskripsi: string;
}

export const TIPE_SOAL: Record<string, TipeInfo> = {
  // ---- 5 INTI AKM ----
  pg: { kode: "pg", label: "Pilihan Ganda", kelompok: "inti", penilaian: "auto", dinilai: true, deskripsi: "Pilih 1 dari beberapa opsi (A–E). Nilai otomatis 1/0." },
  pg_kompleks: { kode: "pg_kompleks", label: "PG Kompleks", kelompok: "inti", penilaian: "auto", dinilai: true, deskripsi: "Pilih semua opsi yang benar. Nilai otomatis PROPORSIONAL = benar_dipilih/total_kunci × bobot, tanpa penalti." },
  menjodohkan: { kode: "menjodohkan", label: "Menjodohkan", kelompok: "inti", penilaian: "auto", dinilai: true, deskripsi: "Pasangkan premis (kiri) dengan respons (kanan). Nilai otomatis PROPORSIONAL per pasangan benar." },
  isian: { kode: "isian", label: "Isian / Jawaban Singkat", kelompok: "inti", penilaian: "auto", dinilai: true, deskripsi: "Ketik jawaban singkat. Nilai otomatis exact-match setelah normalisasi (trim + lowercase + rapatkan spasi)." },
  uraian: { kode: "uraian", label: "Uraian / Esai", kelompok: "inti", penilaian: "manual", dinilai: true, deskripsi: "Jawaban bebas terstruktur. Nilai MANUAL 0–100 × bobot, dikoreksi guru, disembunyikan sampai final." },
  // ---- 11 VARIAN / EKSTENSI APLIKASI ----
  benar_salah: { kode: "benar_salah", label: "Benar-Salah", kelompok: "varian", penilaian: "auto", dinilai: true, deskripsi: "Tentukan Benar/Salah per pernyataan. Nilai otomatis PROPORSIONAL per pernyataan benar." },
  pg_multi: { kode: "pg_multi", label: "PG Multi-Jawaban", kelompok: "varian", penilaian: "auto", dinilai: true, deskripsi: "Seperti PG Kompleks tapi opsi bisa >1 dan label beda. Nilai otomatis PROPORSIONAL, tanpa penalti." },
  urutkan: { kode: "urutkan", label: "Mengurutkan", kelompok: "varian", penilaian: "auto", dinilai: true, deskripsi: "Urutkan langkah/peristiwa. Nilai otomatis PROPORSIONAL per posisi benar." },
  isi_ganda: { kode: "isi_ganda", label: "Isian Multi-Kolom", kelompok: "varian", penilaian: "auto", dinilai: true, deskripsi: "Beberapa kolom isian dalam satu soal. Nilai otomatis PROPORSIONAL per kolom benar." },
  teks_rumpang: { kode: "teks_rumpang", label: "Teks Rumpang", kelompok: "varian", penilaian: "auto", dinilai: true, deskripsi: "Isian dalam paragraf (tanda ___). Nilai otomatis PROPORSIONAL per rumpang benar." },
  peta_konsep: { kode: "peta_konsep", label: "Peta Konsep", kelompok: "varian", penilaian: "auto", dinilai: true, deskripsi: "Pasangkan simpul-simpul konsep. Nilai otomatis PROPORSIONAL per pasangan benar." },
  studi_kasus: { kode: "studi_kasus", label: "Studi Kasus + PG", kelompok: "varian", penilaian: "auto", dinilai: true, deskripsi: "Stimulus panjang (kasus) + soal PG. Nilai otomatis 1/0. Stimulus wajib diisi." },
  pernyataan_sikap: { kode: "pernyataan_sikap", label: "Skala Sikap", kelompok: "varian", penilaian: "non-nilai", dinilai: false, deskripsi: "Skala 1–4 per pernyataan. NON-NILAI — hanya survei/refleksi, tidak masuk nilai akhir." },
  upload_file: { kode: "upload_file", label: "Upload Berkas", kelompok: "varian", penilaian: "manual", dinilai: true, deskripsi: "Jawaban berupa berkas yang diunggah. Nilai MANUAL oleh guru/penguji." },
  praktik_singkat: { kode: "praktik_singkat", label: "Praktik Singkat", kelompok: "varian", penilaian: "manual", dinilai: true, deskripsi: "Jawaban teks + berkas opsional. Nilai MANUAL dengan rubrik sederhana." },
  pilihan_gambar: { kode: "pilihan_gambar", label: "Pilihan Gambar", kelompok: "varian", penilaian: "auto", dinilai: true, deskripsi: "Opsi berupa gambar. Nilai otomatis 1/0." },
};

export const TIPE_LIST: TipeInfo[] = Object.values(TIPE_SOAL);

export function badgeKelompok(kelompok: "inti" | "varian"): string {
  return kelompok === "inti" ? "Inti AKM" : "Varian aplikasi";
}
