CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS psychologist_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    phone VARCHAR(32) NOT NULL,
    role VARCHAR(32) NOT NULL DEFAULT 'psychologist',
    token UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    is_used BOOLEAN NOT NULL DEFAULT FALSE,
    access_until DATE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE IF EXISTS psychologist_invitations
    DROP CONSTRAINT IF EXISTS psychologist_invitations_email_key;

CREATE INDEX IF NOT EXISTS idx_psychologist_invitations_token ON psychologist_invitations(token);
CREATE INDEX IF NOT EXISTS idx_psychologist_invitations_expires_at ON psychologist_invitations(expires_at);
CREATE INDEX IF NOT EXISTS idx_psychologist_invitations_email ON psychologist_invitations(email);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    phone VARCHAR(32) NOT NULL,
    role VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    access_until DATE NOT NULL,
    photo_url TEXT,
    about TEXT NOT NULL DEFAULT '',
    refresh_token VARCHAR(512) UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_access_until ON users(access_until);
