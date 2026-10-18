// Tipe bersama halaman UKK (Tim Fase 2). Bentuk JSON sesuai
// backend/internal/ukk/ukk.go.
import { formatTanggal } from "../../../lib/format";
export interface Ukk {
  id: string;
  nama: string;
  mapel_id: string;
  kelas_id: string;
  jurusan_id: string;
  tanggal: string | null;
  penguji_internal: string | null;
  penguji_eksternal: string | null;
  status: "draft" | "berjalan" | "selesai";
  created_at: string;
  mapel_nama?: string;
  kelas_nama?: string;
  jurusan_kode?: string;
  jurusan_nama?: string;
  jumlah_peserta?: number;
}

export interface UkkNilai {
  id: string;
  ukk_id: string;
  siswa_id: string;
  nilai_teori: number | null;
  nilai_proses: number | null;
  nilai_produk: number | null;
  nilai_sikap: number | null;
  nilai_praktik: number | null;
  nilai_akhir: number | null;
  lulus: boolean | null;
  predikat: string | null;
  catatan: string | null;
  dinilai_oleh: string | null;
  created_at: string;
  updated_at: string;
  siswa_nisn?: string;
  siswa_nama?: string;
}

export interface UkkBerkas {
  id: string;
  ukk_id: string;
  siswa_id: string;
  nama_file: string;
  ukuran: number;
  diupload_oleh: string | null;
  created_at: string;
  siswa_nisn?: string;
  siswa_nama?: string;
}

export interface UkkSertifikat {
  id: string;
  ukk_id: string;
  siswa_id: string;
  nomor_sertifikat: string;
  diterbitkan_at: string;
  siswa_nisn?: string;
  siswa_nama?: string;
  ukk_nama?: string;
  ukk_tanggal?: string | null;
  penguji_internal?: string | null;
  penguji_eksternal?: string | null;
  jurusan_kode?: string;
  jurusan_nama?: string;
  nilai_teori?: number | null;
  nilai_praktik?: number | null;
  nilai_akhir?: number | null;
  predikat?: string | null;
}

export interface MasterOption {
  id: string;
  kode?: string;
  nama: string;
  jurusan_id?: string | null;
  nisn?: string;
  kelas_id?: string;
}

export function fmtNilai(n: number | null | undefined): string {
  return n === null || n === undefined ? "-" : n.toFixed(2);
}

export function fmtTanggal(t: string | null | undefined): string {
  // Nama lama dipertahankan agar call-site UKK tidak berubah; isi
  // didelegasikan ke helper pusat (bulan singkat: "16 Okt 2026").
  return formatTanggal(t);
}

export function fmtUkuran(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
}
