-- Rollback migrasi 0005 (Tim 5). Urutan: remedial -> berita_acara -> nilai_akhir.
DROP TABLE IF EXISTS remedial;
DROP TABLE IF EXISTS berita_acara;
DROP TABLE IF EXISTS nilai_akhir;
