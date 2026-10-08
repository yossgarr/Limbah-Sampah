-- Mengaktifkan ekstensi UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enum Types
DO $$ BEGIN
    CREATE TYPE user_role AS ENUM ('user', 'observer', 'superadmin');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE auth_source_type AS ENUM ('sso', 'external');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- 1. Users Table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    auth_source auth_source_type NOT NULL,
    keycloak_sub VARCHAR(255) UNIQUE,
    keycloak_iss VARCHAR(255),
    email VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255),
    berlaku_hingga TIMESTAMP,
    failed_login_count INT DEFAULT 0,
    locked_until TIMESTAMP,
    role user_role NOT NULL DEFAULT 'user',
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_login TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    satuan_kerja VARCHAR(255),
    unit_kerja VARCHAR(255),
    status_pegawai VARCHAR(50),
    golongan_group VARCHAR(10),
    age_band VARCHAR(20)
);

-- 2. Factor Sets Table
CREATE TABLE IF NOT EXISTS factor_sets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    version_name VARCHAR(100) NOT NULL UNIQUE,
    valid_from DATE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 3. Factors Table
CREATE TABLE IF NOT EXISTS factors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    set_id UUID NOT NULL REFERENCES factor_sets(id) ON DELETE CASCADE,
    category_code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    value NUMERIC(15, 6) NOT NULL,
    unit VARCHAR(50) NOT NULL,
    source_ref VARCHAR(255) NOT NULL
);

-- 4. Waste Reports Table
CREATE TABLE IF NOT EXISTS waste_reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    category_code VARCHAR(50) NOT NULL,
    activity_date DATE NOT NULL,
    input_amount NUMERIC(15, 6) NOT NULL,
    factor_value NUMERIC(15, 6) NOT NULL,
    factor_unit VARCHAR(50) NOT NULL,
    factor_source VARCHAR(255) NOT NULL,
    factor_version VARCHAR(100) NOT NULL,
    result_total NUMERIC(15, 6) NOT NULL,
    user_satuan_kerja VARCHAR(255),
    user_unit_kerja VARCHAR(255),
    user_status_pegawai VARCHAR(50),
    user_golongan_group VARCHAR(10),
    user_age_band VARCHAR(20),
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_reports_activity_date ON waste_reports(activity_date);
CREATE INDEX IF NOT EXISTS idx_reports_satuan_kerja ON waste_reports(user_satuan_kerja);
CREATE INDEX IF NOT EXISTS idx_users_auth_source ON users(auth_source);