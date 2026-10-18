-- Migrasi 0006 (Integrasi Fase 1): FK soal -> kisi_kisi / kompetensi.
--
-- Kolom soal.kisi_kisi_id / soal.kompetensi_id di 0003 sengaja dibuat
-- TANPA FK karena tabel kisi_kisi / kompetensi (0002) belum ada saat
-- itu. Sekarang kedua tabel sudah ada, FK ditambahkan di sini.
-- ON DELETE SET NULL: menghapus kisi-kisi/kompetensi tidak menghapus
-- soal — soal kehilangan rujukan itu saja.
-- Data yang tidak valid (merujuk ID yang tidak ada) dibersihkan dulu
-- (di-NULL-kan) agar ADD CONSTRAINT tidak gagal.

UPDATE soal SET kisi_kisi_id = NULL
WHERE kisi_kisi_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM kisi_kisi WHERE kisi_kisi.id = soal.kisi_kisi_id);

UPDATE soal SET kompetensi_id = NULL
WHERE kompetensi_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM kompetensi WHERE kompetensi.id = soal.kompetensi_id);

ALTER TABLE soal
    ADD CONSTRAINT fk_soal_kisi_kisi
        FOREIGN KEY (kisi_kisi_id) REFERENCES kisi_kisi (id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_soal_kompetensi
        FOREIGN KEY (kompetensi_id) REFERENCES kompetensi (id) ON DELETE SET NULL;
