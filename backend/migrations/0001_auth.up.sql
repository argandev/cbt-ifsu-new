-- Migrasi 0001 (Tim 1: Scaffolding + Auth + RBAC)
-- Tabel fondasi: roles, permissions, role_permissions, users, sessions,
-- audit_log, settings.
--
-- KONVENSI yang dipakai tim lain (JANGAN ubah nama kolom tanpa koordinasi):
-- - roles(id BIGSERIAL PK, code UNIQUE, name, description, is_active)
--   dipakai seed Tim 6 (backend/db/seed/seed.sql).
-- - permissions(id BIGSERIAL PK, code UNIQUE, description, group_name)
--   dipakai seed Tim 6.
-- - role_permissions(role_id BIGINT -> roles(id), permission_id BIGINT ->
--   permissions(id), PK(role_id, permission_id)) dipakai seed Tim 6.
-- - users(id UUID) dirujuk migrasi 0003 (bank_soal.guru_id) dan Tim 2.
-- - audit_log: audit.Catat INSERT (actor_id, aksi, entitas, entitas_id,
--   detail, ip); GET audit membaca (id, actor_id, aksi, entitas, entitas_id,
--   detail, ip, created_at). entitas_id TEXT (boleh UUID / kode teks).
-- - settings(key PK, value JSONB, updated_at): settings.go membaca
--   SELECT key, value dan PUT INSERT (key, value, updated_at).
--   Kolom group_name/description ditambahkan nullable agar INSERT seed
--   Tim 6 (yang menyebut kedua kolom itu) tidak gagal karena kolom
--   tidak ada. CATATAN UNTUK TIM 6: nilai string di seed WAJIB literal
--   JSONB valid — '"tertinggi"'::jsonb, bukan 'tertinggi' (lihat KONTRAK).
-- - Seed DATA (59 permission + matriks + settings default) BUKAN di
--   migrasi ini, melainkan di backend/db/seed/seed.sql (Tim 6).
-- - TIDAK ADA akun demo di seed maupun migrasi.

CREATE TABLE IF NOT EXISTS roles (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS permissions (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    description TEXT,
    group_name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id BIGINT NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions (id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX IF NOT EXISTS idx_role_permissions_permission ON role_permissions (permission_id);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nama TEXT NOT NULL,
    email TEXT UNIQUE,
    username TEXT UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('admin', 'proktor', 'guru', 'pengawas', 'siswa', 'penguji_ukk')),
    must_change_password BOOLEAN NOT NULL DEFAULT TRUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (email IS NOT NULL OR username IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users (lower(email));
CREATE INDEX IF NOT EXISTS idx_users_username ON users (lower(username));
CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);

CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    csrf_token TEXT NOT NULL,
    ip TEXT,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions (expires_at);

-- Audit log: append-only. TIDAK ADA endpoint ubah/hapus lewat API.
CREATE TABLE IF NOT EXISTS audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id UUID REFERENCES users (id) ON DELETE SET NULL,
    aksi TEXT NOT NULL,
    entitas TEXT NOT NULL,
    entitas_id TEXT,
    detail JSONB NOT NULL DEFAULT '{}'::jsonb,
    ip TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_audit_actor ON audit_log (actor_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_entitas ON audit_log (entitas, entitas_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_aksi ON audit_log (aksi, created_at DESC);

CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value JSONB NOT NULL,
    group_name TEXT,
    description TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
