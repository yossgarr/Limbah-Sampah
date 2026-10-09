-- Sisakan tepat 3 user (superadmin, observer, user), masing-masing dengan id unik
-- dan login email+password (Demo12345!), serta 3 kategori limbah:
-- S1_PLASTIK, S2_KERTAS, S3_ORGANIK.

UPDATE users SET
    email = 'superadmin@example.com',
    password_hash = '$argon2id$v=19$m=65536,t=1,p=8$HcYGM7VFSuY8EhA3i7iVJw$3kBWJzSsW6iRb1vVOhP8pAcbVqdUt742XkqkoaNp94s',
    berlaku_hingga = '2030-12-31 23:59:59',
    failed_login_count = 0,
    locked_until = NULL
WHERE id = '11111111-1111-1111-1111-111111111111';

UPDATE users SET
    email = 'observer@example.com',
    password_hash = '$argon2id$v=19$m=65536,t=1,p=8$HcYGM7VFSuY8EhA3i7iVJw$3kBWJzSsW6iRb1vVOhP8pAcbVqdUt742XkqkoaNp94s',
    berlaku_hingga = '2030-12-31 23:59:59',
    failed_login_count = 0,
    locked_until = NULL
WHERE id = '22222222-2222-2222-2222-222222222222';

UPDATE users SET
    auth_source = 'external',
    email = 'user@example.com',
    password_hash = '$argon2id$v=19$m=65536,t=1,p=8$HcYGM7VFSuY8EhA3i7iVJw$3kBWJzSsW6iRb1vVOhP8pAcbVqdUt742XkqkoaNp94s',
    keycloak_sub = NULL,
    keycloak_iss = NULL,
    berlaku_hingga = '2030-12-31 23:59:59',
    failed_login_count = 0,
    locked_until = NULL
WHERE id = '33333333-3333-3333-3333-333333333333';

-- Buang laporan di luar 3 kategori / milik user lain, lalu user & faktor di luar yang dipakai
DELETE FROM waste_reports WHERE category_code NOT IN ('S1_PLASTIK', 'S2_KERTAS', 'S3_ORGANIK');
DELETE FROM waste_reports WHERE user_id NOT IN (
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    '33333333-3333-3333-3333-333333333333'
);
DELETE FROM users WHERE id NOT IN (
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    '33333333-3333-3333-3333-333333333333'
);
DELETE FROM factors WHERE category_code NOT IN ('S1_PLASTIK', 'S2_KERTAS', 'S3_ORGANIK');
