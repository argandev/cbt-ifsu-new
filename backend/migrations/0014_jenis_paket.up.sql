-- 0014: STS -> PTS, SAS -> PAS (koreksi istilah)
UPDATE paket_ujian SET jenis = 'PTS' WHERE jenis = 'STS';
UPDATE paket_ujian SET jenis = 'PAS' WHERE jenis = 'SAS';
ALTER TABLE paket_ujian DROP CONSTRAINT IF EXISTS paket_ujian_jenis_check;
ALTER TABLE paket_ujian ADD CONSTRAINT paket_ujian_jenis_check CHECK (jenis IN ('PTS', 'PAS', 'USBK', 'UKK', 'Tryout'));
