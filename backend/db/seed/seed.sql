-- =====================================================================
-- CBT SMK v2 — Seed PRODUKSI (Tim 6: Dokumen + Deploy + Seed)
-- =====================================================================
-- ISI FILE INI:
--   1. roles         (6 role)
--   2. permissions   (67 permission, PERSIS daftar kontrak Tim 1 + Fase 3 + UC20)
--   3. role_permissions (matriks sesuai kontrak)
--   4. settings      (6 pengaturan default)
--
-- YANG SENGAJA TIDAK ADA DI FILE INI:
--   - TIDAK ADA USER / AKUN APAPUN (tanpa akun demo, tanpa admin bawaan).
--   - Akun pertama dibuat via CLI:  go run ./cmd/server create-admin ...
--     (milik Tim 1), lalu PEMILIK AKUN WAJIB ganti password saat login pertama.
--   - TIDAK ADA data master contoh (jurusan/kelas/mapel/guru/siswa),
--     TIDAK ADA soal contoh, TIDAK ADA jadwal contoh.
--     Semua data sekolah diisi admin lewat aplikasi / import CSV.
--
-- IDEMPOTEN: aman dijalankan berulang kali (ON CONFLICT DO NOTHING).
--   Jalankan SETELAH migrasi Tim 1:  make migrate-up && make seed
--
-- ASUMSI SKEMA (milik Tim 1 — bila nama kolom berbeda, sesuaikan file ini):
--   roles(id BIGSERIAL PK, code TEXT UNIQUE NOT NULL,
--         name TEXT NOT NULL, description TEXT, is_active BOOLEAN NOT NULL DEFAULT TRUE)
--   permissions(id BIGSERIAL PK, code TEXT UNIQUE NOT NULL,
--         description TEXT, group_name TEXT NOT NULL)
--   role_permissions(role_id BIGINT REFERENCES roles(id),
--         permission_id BIGINT REFERENCES permissions(id),
--         PRIMARY KEY (role_id, permission_id))
--   settings(key TEXT PRIMARY KEY, value TEXT NOT NULL,
--         group_name TEXT NOT NULL, description TEXT)
--
-- Catatan: file ini TIDAK menyalin mekanisme lisensi apa pun dan TIDAK
-- meniru data dari proyek lain. Daftar permission ditulis ulang dari kontrak.
-- =====================================================================

BEGIN;

-- ---------------------------------------------------------------------
-- 1. Roles (6 role)
-- ---------------------------------------------------------------------
INSERT INTO roles (code, name, description, is_active) VALUES
  ('admin',       'Admin',       'Pengelola sistem penuh: semua permission, backup/restore, settings, audit.', TRUE),
  ('proktor',     'Proktor',     'Pelaksana ujian: kelola paket & jadwal, monitor, unblock, event, backup.', TRUE),
  ('guru',        'Guru',        'Penyusun soal & paket, pengoreksi esai, rekap mapelnya.', TRUE),
  ('pengawas',    'Pengawas',    'Pengawas ruang: monitor ujian yang sedang berjalan, tanpa akses soal/token.', TRUE),
  ('siswa',       'Siswa',       'Peserta ujian: mengerjakan ujian dan melihat rekap nilainya sendiri.', TRUE),
  ('penguji_ukk', 'Penguji UKK', 'Penguji Uji Kompetensi Keahlian: nilai teori + praktik UKK.', TRUE)
ON CONFLICT (code) DO NOTHING;

-- ---------------------------------------------------------------------
-- 2. Permissions (67 permission — PERSIS daftar kontrak)
--    Catatan Fase 3 (BUG-01): analisis.read/analisis.export sebelumnya
--    HANYA ada di migrasi 0008 — grant 0008 ke proktor+guru diam-diam
--    no-op karena tabel roles masih KOSONG saat migrasi berjalan pada
--    DB fresh (roles diisi seed, BUKAN migrasi). SEED adalah sumber
--    kebenaran grant: daftar tetap di bawah ini + admin CROSS JOIN.
--    Migrasi 0008/0012 tetap menyisipkan baris permission (ON CONFLICT
--    DO NOTHING) agar DB yang sudah di-migrate sebelum seed tetap
--    punya permission-nya, tetapi grant yang berlaku adalah dari seed.
-- ---------------------------------------------------------------------
INSERT INTO permissions (code, description, group_name) VALUES
  -- Master (21)
  ('master.jurusan.read',   'Lihat data jurusan',        'master'),
  ('master.jurusan.create', 'Tambah jurusan',             'master'),
  ('master.jurusan.update', 'Ubah jurusan',               'master'),
  ('master.jurusan.delete', 'Hapus jurusan',              'master'),
  ('master.kelas.read',     'Lihat data kelas',           'master'),
  ('master.kelas.create',   'Tambah kelas',               'master'),
  ('master.kelas.update',   'Ubah kelas',                 'master'),
  ('master.kelas.delete',   'Hapus kelas',                'master'),
  ('master.mapel.read',     'Lihat data mata pelajaran',  'master'),
  ('master.mapel.create',   'Tambah mata pelajaran',      'master'),
  ('master.mapel.update',   'Ubah mata pelajaran',        'master'),
  ('master.mapel.delete',   'Hapus mata pelajaran',       'master'),
  ('master.guru.read',      'Lihat data guru',            'master'),
  ('master.guru.create',    'Tambah guru',                'master'),
  ('master.guru.update',    'Ubah guru',                  'master'),
  ('master.guru.delete',    'Hapus guru',                 'master'),
  ('master.siswa.read',     'Lihat data siswa',           'master'),
  ('master.siswa.create',   'Tambah siswa',               'master'),
  ('master.siswa.update',   'Ubah siswa',                 'master'),
  ('master.siswa.delete',   'Hapus siswa',                'master'),
  ('master.siswa.import',   'Import siswa dari CSV',      'master'),
  -- Bank soal (6)
  ('banksoal.read',   'Lihat bank soal',          'banksoal'),
  ('banksoal.create', 'Tambah soal',              'banksoal'),
  ('banksoal.update', 'Ubah soal',                'banksoal'),
  ('banksoal.delete', 'Hapus soal',               'banksoal'),
  ('banksoal.lock',   'Kunci / buka kunci bank soal', 'banksoal'),
  ('banksoal.import', 'Import soal dari CSV',     'banksoal'),
  -- Paket ujian (4)
  ('paket.read',   'Lihat paket ujian',   'paket'),
  ('paket.create', 'Buat paket ujian',   'paket'),
  ('paket.update', 'Ubah paket ujian',   'paket'),
  ('paket.delete', 'Hapus paket ujian',  'paket'),
  -- Jadwal (8)
  ('jadwal.read',           'Lihat jadwal ujian',              'jadwal'),
  ('jadwal.create',         'Buat jadwal ujian',               'jadwal'),
  ('jadwal.update',         'Ubah jadwal ujian',               'jadwal'),
  ('jadwal.delete',         'Hapus jadwal ujian',              'jadwal'),
  ('jadwal.pause',          'Jeda jadwal (timer dibekukan)',   'jadwal'),
  ('jadwal.resume',         'Lanjutkan jadwal yang dijeda',    'jadwal'),
  ('jadwal.unblock',        'Buka blokir peserta (lanjut sisa waktu)', 'jadwal'),
  ('jadwal.usulan-unblock', 'Mengusulkan unblock peserta (pengawas)',  'jadwal'),
  -- Ujian (2)
  ('ujian.kerjakan', 'Mengerjakan ujian (siswa)',        'ujian'),
  ('ujian.monitor',  'Memantau ujian yang sedang jalan', 'ujian'),
  -- Koreksi (2)
  ('koreksi.read',   'Lihat hasil koreksi',   'koreksi'),
  ('koreksi.update', 'Koreksi esai / isian',  'koreksi'),
  -- Rekap (4)
  ('rekap.read',   'Lihat rekap nilai',   'rekap'),
  ('rekap.export', 'Export rekap ke CSV', 'rekap'),
  ('rekap.finalisasi', 'Finalisasi nilai akhir rekap', 'rekap'),
  ('rekap.remedial',   'Catat nilai remedial', 'rekap'),
  -- Settings (2)
  ('settings.read',   'Lihat pengaturan sistem', 'settings'),
  ('settings.update', 'Ubah pengaturan sistem',  'settings'),
  -- Audit (1)
  ('audit.read', 'Lihat audit log', 'audit'),
  -- UKK (3)
  ('ukk.read',   'Lihat data UKK',   'ukk'),
  ('ukk.create', 'Buat data UKK',    'ukk'),
  ('ukk.update', 'Ubah / nilai UKK', 'ukk'),
  -- Event (4)
  ('event.read',   'Lihat event / ruangan',   'event'),
  ('event.create', 'Buat event / ruangan',    'event'),
  ('event.update', 'Ubah event / ruangan',    'event'),
  ('event.delete', 'Hapus event / ruangan',   'event'),
  -- Backup (2)
  ('backup.create',  'Buat backup',    'backup'),
  ('backup.restore', 'Restore backup', 'backup'),
  -- Analisis butir soal (2, Fase 2 via migrasi 0008; ditegaskan di seed
  -- pada Fase 3 — BUG-01: tanpa baris ini proktor+guru 403)
  ('analisis.read',   'Lihat analisis butir soal',        'analisis'),
  ('analisis.export', 'Export analisis butir soal ke CSV', 'analisis'),
  -- SoD override (1, Fase 3, migrasi 0012) — admin saja
  ('sod.override', 'Mencatat pengecualian darurat SoD (alasan tertulis + audit + notifikasi)', 'sod'),
  -- Manajemen Pengguna staf (5, UC20, migrasi 0013) — admin saja.
  -- Akun siswa TIDAK dikelola lewat permission ini (siswa di halaman Siswa).
  ('pengguna.read',           'Lihat daftar & detail akun staf',        'pengguna'),
  ('pengguna.create',         'Buat akun staf',                         'pengguna'),
  ('pengguna.update',         'Ubah nama/email/role akun staf',         'pengguna'),
  ('pengguna.reset-password', 'Reset password akun staf',               'pengguna'),
  ('pengguna.toggle-aktif',  'Aktif/nonaktifkan akun staf',            'pengguna')
ON CONFLICT (code) DO NOTHING;

-- ---------------------------------------------------------------------
-- 3. Role <-> Permission
--    Pola: INSERT ... SELECT ... WHERE code IN (...) ON CONFLICT DO NOTHING.
--    Admin = SEMUA permission (tanpa daftar eksplisit, agar permission baru
--    otomatis ikut admin; permission baru tetap harus ditambahkan di
--    Bagian 2 file ini + kontrak).
-- ---------------------------------------------------------------------

-- 3a. Admin: semua permission.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p WHERE r.code = 'admin'
ON CONFLICT DO NOTHING;

-- 3b. Proktor: master.*.read, banksoal.read, jadwal.* (semua 8),
--     ujian.monitor, koreksi.read, rekap.* (semua 4), settings.read,
--     audit.read, analisis.* (2), event.*, backup.*  (total 30)
--     Proktor TIDAK mendapat paket.* — proktor tidak boleh melihat isi soal.
--     Proktor melihat analisis TANPA teks pertanyaan (disembunyikan di
--     handler analisis — KONTRAK §5.3, SoD proktor vs isi soal).
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p ON TRUE
WHERE r.code = 'proktor' AND p.code IN (
  'master.jurusan.read', 'master.kelas.read', 'master.mapel.read',
  'master.guru.read', 'master.siswa.read',
  'banksoal.read',
  'jadwal.read', 'jadwal.create', 'jadwal.update', 'jadwal.delete',
  'jadwal.pause', 'jadwal.resume', 'jadwal.unblock', 'jadwal.usulan-unblock',
  'ujian.monitor',
  'koreksi.read',
  'rekap.read', 'rekap.export', 'rekap.finalisasi', 'rekap.remedial',
  'settings.read',
  'audit.read',
  'analisis.read', 'analisis.export',
  'event.read', 'event.create', 'event.update', 'event.delete',
  'backup.create', 'backup.restore'
)
ON CONFLICT DO NOTHING;

-- 3c. Guru: master.*.read, banksoal.*, paket.*, jadwal.{read,create,update,delete},
--     ujian.monitor, koreksi.*, rekap.read, rekap.remedial, settings.read,
--     analisis.* (2)  (total 27)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p ON TRUE
WHERE r.code = 'guru' AND p.code IN (
  'master.jurusan.read', 'master.kelas.read', 'master.mapel.read',
  'master.guru.read', 'master.siswa.read',
  'banksoal.read', 'banksoal.create', 'banksoal.update', 'banksoal.delete',
  'banksoal.lock', 'banksoal.import',
  'paket.read', 'paket.create', 'paket.update', 'paket.delete',
  'jadwal.read', 'jadwal.create', 'jadwal.update', 'jadwal.delete',
  'ujian.monitor',
  'koreksi.read', 'koreksi.update',
  'rekap.read', 'rekap.remedial',
  'settings.read',
  'analisis.read', 'analisis.export'
)
ON CONFLICT DO NOTHING;

-- 3d. Pengawas: ujian.monitor, jadwal.read, jadwal.usulan-unblock,
--     rekap.read  (total 4)
--     Pengawas mengusulkan unblock lewat permission jadwal.usulan-unblock;
--     eksekusi unblock tetap proktor/admin.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p ON TRUE
WHERE r.code = 'pengawas' AND p.code IN (
  'ujian.monitor', 'jadwal.read', 'jadwal.usulan-unblock', 'rekap.read'
)
ON CONFLICT DO NOTHING;

-- 3e. Siswa: ujian.kerjakan, rekap.read  (total 2)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p ON TRUE
WHERE r.code = 'siswa' AND p.code IN (
  'ujian.kerjakan', 'rekap.read'
)
ON CONFLICT DO NOTHING;

-- 3f. Penguji UKK: ukk.*, rekap.read  (total 4)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p ON TRUE
WHERE r.code = 'penguji_ukk' AND p.code IN (
  'ukk.read', 'ukk.create', 'ukk.update', 'rekap.read'
)
ON CONFLICT DO NOTHING;

-- ---------------------------------------------------------------------
-- 4. Settings default (produksi)
--    Nilai di sini adalah DEFAULT. Admin dapat mengubahnya dari menu
--    Settings di aplikasi (permission settings.update). Seed tidak
--    menimpa nilai yang sudah diubah admin (ON CONFLICT DO NOTHING).
-- ---------------------------------------------------------------------
INSERT INTO settings (key, value, group_name, description) VALUES
  ('tampil_hasil_ke_siswa', 'false',
   'ujian', 'Nilai esai disembunyikan sampai koreksi final. Jika true, siswa melihat nilai otomatis sementara setelah submit.'),
  ('token_expiry_detik', '900',
   'ujian', 'Masa berlaku token ujian dalam detik (900 = 15 menit). Token kedua membuat token pertama hangus.'),
  ('toleransi_telat_menit', '15',
   'ujian', 'Toleransi keterlambatan siswa masuk ujian, dalam menit.'),
  ('max_pelanggaran', '2',
   'anti-curang', 'Jumlah pelanggaran sebelum auto-block per-attempt: pelanggaran ke-1 = peringatan, ke-2 = auto-block.'),
  ('remedial_mode', '"tertinggi"'::jsonb,
   'nilai', 'Nilai remedial yang dipakai: tertinggi (bukan rata-rata / terakhir).'),
  ('nama_sekolah', '"SMK (isi di settings)"'::jsonb,
   'sekolah', 'Nama sekolah — WAJIB diisi admin di menu Settings sebelum go-live.')
ON CONFLICT (key) DO NOTHING;

COMMIT;

-- ---------------------------------------------------------------------
-- Verifikasi setelah seed (jalankan manual, jangan di-commit sebagai data):
--   SELECT (SELECT count(*) FROM roles) AS roles,               -- harus 6
--          (SELECT count(*) FROM permissions) AS permissions,   -- harus 67
--          (SELECT count(*) FROM settings) AS settings;         -- harus 6
--   SELECT r.code, count(*) FROM role_permissions rp
--     JOIN roles r ON r.id = rp.role_id GROUP BY r.code ORDER BY r.code;
--     -- admin 67, proktor 30, guru 27, pengawas 4, siswa 2, penguji_ukk 4
--   SELECT count(*) AS jumlah_user FROM users; -- harus 0 (TANPA akun demo)
-- ---------------------------------------------------------------------
