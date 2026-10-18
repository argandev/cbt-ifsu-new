-- Rollback migrasi 0002 (Tim 2: Master Data).
-- Urutan: anak dulu (kisi_kisi, kompetensi, siswa, guru), lalu mapel,
-- kelas, jurusan. Dipakai hanya sebelum data produksi masuk; JANGAN
-- rollback di produksi yang sudah berisi data ujian.
DROP TABLE IF EXISTS kisi_kisi;
DROP TABLE IF EXISTS kompetensi;
DROP TABLE IF EXISTS siswa;
DROP TABLE IF EXISTS guru;
DROP TABLE IF EXISTS mapel;
DROP TABLE IF EXISTS kelas;
DROP TABLE IF EXISTS jurusan;
