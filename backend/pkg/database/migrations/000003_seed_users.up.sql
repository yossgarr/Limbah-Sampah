-- 1. Superadmin (Auth External)
INSERT INTO users (
    id, 
    auth_source, 
    email, 
    password_hash, 
    role, 
    is_active, 
    berlaku_hingga
) VALUES (
    '11111111-1111-1111-1111-111111111111',
    'external',
    'admin@example.com',
    '$argon2id$v=19$m=65536,t=3,p=2$c2FsdHNhbHQ$Demo12345HashValuePlaceholder', -- Hash Argon2id untuk 'Demo12345!'
    'superadmin',
    true,
    '2030-12-31 23:59:59'
) ON CONFLICT (email) DO NOTHING;

-- 2. Observer / Pengamat Eksternal (Auth External)
INSERT INTO users (
    id, 
    auth_source, 
    email, 
    password_hash, 
    role, 
    is_active, 
    berlaku_hingga
) VALUES (
    '22222222-2222-2222-2222-222222222222',
    'external',
    'observer@example.com',
    '$argon2id$v=19$m=65536,t=3,p=2$c2FsdHNhbHQ$Demo12345HashValuePlaceholder', -- Hash Argon2id untuk 'Demo12345!'
    'observer',
    true,
    '2026-12-31 23:59:59'
) ON CONFLICT (email) DO NOTHING;

-- 3. Pegawai Internal (Auth SSO Keycloak - Subdomain example.com)
INSERT INTO users (
    id, 
    auth_source, 
    email,
    keycloak_sub, 
    keycloak_iss, 
    role, 
    is_active,
    satuan_kerja, 
    unit_kerja, 
    status_pegawai, 
    golongan_group, 
    age_band
) VALUES (
    '33333333-3333-3333-3333-333333333333',
    'sso',
    'pegawai@example.com',
    'kc-sub-pegawai-001',
    'https://sso.example.com/realms/ecolimbah',
    'user',
    true,
    'Deputi Bidang Lingkungan Hidup',
    'Direktorat Pengelolaan Sampah',
    'PNS',
    'III',
    '31-40'
) ON CONFLICT (keycloak_sub) DO NOTHING;