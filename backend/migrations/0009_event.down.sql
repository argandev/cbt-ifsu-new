-- Rollback migrasi 0009 (Tim Fase 2 — Event + Ruangan + Kartu Peserta).
-- Urutan: lepas kolom jadwal dulu, lalu hapus tabel ruangan + event.
-- Dipakai hanya sebelum data produksi masuk; JANGAN rollback di
-- produksi yang sudah berisi data ujian.
ALTER TABLE jadwal_ujian
    DROP COLUMN IF EXISTS ruangan_id,
    DROP COLUMN IF EXISTS event_id;

DROP TABLE IF EXISTS ruangan;
DROP TABLE IF EXISTS event;
