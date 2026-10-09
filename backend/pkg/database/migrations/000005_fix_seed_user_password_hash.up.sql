-- Ganti hash placeholder di 000003 dengan hash Argon2id valid untuk 'Demo12345!'
UPDATE users
SET password_hash = '$argon2id$v=19$m=65536,t=1,p=8$HcYGM7VFSuY8EhA3i7iVJw$3kBWJzSsW6iRb1vVOhP8pAcbVqdUt742XkqkoaNp94s'
WHERE email IN ('admin@example.com', 'observer@example.com');
