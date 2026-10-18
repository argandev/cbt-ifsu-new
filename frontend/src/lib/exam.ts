// Tipe respons API Tim 4 (paket, jadwal, ujian, monitor, koreksi).
// Terpisah dari lib/api.ts (Tim 1) agar tidak mengubah berkas Tim 1.
// apiClient tetap dari lib/api.ts.
export interface Paket {
  id: string;
  kode: string;
  nama: string;
  mapel_id: string;
  mapel_nama?: string;
  jenis: "PTS" | "PAS" | "USBK" | "UKK" | "Tryout";
  durasi_menit: number;
  jumlah_soal: number;
  jumlah_soal_tersusun?: number;
  acak_soal: boolean;
  acak_jawaban: boolean;
  tampil_hasil: boolean;
  created_at?: string;
}

export interface Jadwal {
  id: string;
  paket_id: string;
  paket_kode?: string;
  paket_nama?: string;
  paket_jenis?: string;
  kelas_id: string;
  kelas_nama?: string;
  tanggal: string;
  jam_mulai: string;
  jam_selesai: string;
  status: "draft" | "aktif" | "pause" | "selesai";
  token_ada?: boolean;
  token_expires_at?: string | null;
  durasi_menit?: number;
  tampil_hasil?: boolean;
  created_at?: string;
}

export interface SoalUjian {
  soal_id: string;
  nomor: number;
  tipe: string;
  pertanyaan: string;
  stimulus?: string;
  media_url?: string;
  tampilan: Record<string, unknown> | null;
  bobot: number;
  jawaban: unknown;
  ragu: boolean;
}

export interface AttemptDetail {
  attempt_id: string;
  jadwal_id: string;
  status: string;
  deadline: string;
  mulai_at?: string;
  sisa_waktu_detik: number;
  jadwal_status: string;
  paket_jenis: string;
  server_now: string;
  soal: SoalUjian[];
  nilai_akhir?: number | null;
  menunggu_koreksi?: boolean;
}

export interface AttemptStatus {
  attempt_id: string;
  jadwal_id: string;
  status: string;
  deadline: string;
  jadwal_status: string;
  sisa_waktu_detik: number;
  dijawab: number;
  ragu: number;
  total_soal: number;
  belum_dijawab: number;
  server_now: string;
  nilai_akhir?: number | null;
  menunggu_koreksi?: boolean;
}

export interface PesertaMonitor {
  attempt_id: string;
  siswa_id: string;
  nama_siswa: string;
  nisn: string;
  kelas_nama: string;
  status: string;
  mulai_at?: string;
  deadline?: string;
  sisa_waktu_detik?: number | null;
  nilai_akhir?: number | null;
  dijawab: number;
  ragu: number;
  pelanggaran: number;
  total_soal: number;
}

export interface MonitorSnapshot {
  jadwal_id: string;
  generated_at: string;
  peserta: PesertaMonitor[];
}

export interface UsulanUnblock {
  id: string;
  attempt_id: string;
  pengawas_id: string;
  alasan: string;
  status: "menunggu" | "disetujui" | "ditolak";
  diproses_oleh?: string | null;
  diproses_at?: string | null;
  created_at?: string;
  nama_siswa: string;
}
