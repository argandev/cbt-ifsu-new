// useExamStore — state ujian siswa (Zustand, Tim 1).
// Nilai TIDAK PERNAH dihitung di sini; store hanya menyimpan jawaban
// sementara sebelum autosave ke server (Tim Ujian yang mengelola autosave).
import { create } from "zustand";

export type JawabanValue = unknown;

interface ExamState {
  attemptId: string | null;
  /** jawaban[soalId] = jawaban siswa. */
  answers: Record<string, JawabanValue>;
  /** ragu[soalId] = true bila ditandai ragu-ragu. */
  ragu: Record<string, boolean>;
  start: (attemptId: string) => void;
  setAnswer: (soalId: string, jawaban: JawabanValue) => void;
  toggleRagu: (soalId: string) => void;
  reset: () => void;
}

export const useExamStore = create<ExamState>((set) => ({
  attemptId: null,
  answers: {},
  ragu: {},

  start: (attemptId) => set({ attemptId, answers: {}, ragu: {} }),
  setAnswer: (soalId, jawaban) =>
    set((s) => ({ answers: { ...s.answers, [soalId]: jawaban } })),
  toggleRagu: (soalId) =>
    set((s) => ({ ragu: { ...s.ragu, [soalId]: !s.ragu[soalId] } })),
  reset: () => set({ attemptId: null, answers: {}, ragu: {} }),
}));
