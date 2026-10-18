-- Rollback migrasi 0006: lepas FK soal -> kisi_kisi / kompetensi.
ALTER TABLE soal
    DROP CONSTRAINT IF EXISTS fk_soal_kisi_kisi,
    DROP CONSTRAINT IF EXISTS fk_soal_kompetensi;
