-- Migrasi 0011 (QA Fase 2): kolom pembahasan pada tabel soal.
--
-- Selama ini soal tidak punya kolom pembahasan (exam.go mencatat
-- "Tim 3 tidak punya kolom pembahasan di soal"): halaman hasil hanya
-- bisa menampilkan kunci benar yang dirangkum (kunciTampil), bukan
-- penjelasan guru. Kolom ini diisi guru lewat SoalEditor dan
-- ditampilkan di HasilPage bila paket Tryout ATAU tampil_hasil
-- efektif. Endpoint attempt siswa TIDAK BOLEH mengembalikan kolom
-- ini sebelum submit (exam.go tidak memilih kolom pembahasan untuk
-- soal ujian) — pembahasan hanya muncul di endpoint hasil setelah
-- submit.
--
-- Backend soal guru (banksoal.go) + hasilAttempt (exam.go) perlu
-- membaca/menulis kolom ini — di luar cakupan migrasi ini; lihat
-- laporan QA Fase 2.

ALTER TABLE soal
    ADD COLUMN IF NOT EXISTS pembahasan TEXT NOT NULL DEFAULT '';
