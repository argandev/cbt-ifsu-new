-- rollback 0014
UPDATE paket_ujian SET jenis = 'STS' WHERE jenis = 'PTS';
UPDATE paket_ujian SET jenis = 'SAS' WHERE jenis = 'PAS';
ALTER TABLE paket_ujian DROP CONSTRAINT IF EXISTS paket_ujian_jenis_check;
ALTER TABLE paket_ujian ADD CONSTRAINT paket_ujian_jenis_check CHECK (jenis IN ('STS', 'SAS', 'USBK', 'UKK', 'Tryout'));
