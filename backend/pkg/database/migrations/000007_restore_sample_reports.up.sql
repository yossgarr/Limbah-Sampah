-- Pastikan 3 data contoh (satu per kategori) milik user@example.com ada.
INSERT INTO waste_reports (
    id, user_id, category_code, activity_date, input_amount,
    factor_value, factor_unit, factor_source, factor_version, result_total,
    user_satuan_kerja, user_unit_kerja, user_status_pegawai, user_golongan_group, user_age_band
) VALUES
    ('44444444-4444-4444-4444-444444444441', '33333333-3333-3333-3333-333333333333', 'S1_PLASTIK', '2026-10-01', 10.5, 1.5, 'Poin/Kg', 'KLHK (2025)', 'EcoLimbah-v1-2026', 15.75, 'Deputi Bidang Lingkungan Hidup', 'Direktorat Pengelolaan Sampah', 'PNS', 'III', '31-40'),
    ('44444444-4444-4444-4444-444444444442', '33333333-3333-3333-3333-333333333333', 'S2_KERTAS',  '2026-10-03', 25,   0.8, 'Poin/Kg', 'KLHK (2025)', 'EcoLimbah-v1-2026', 20,    'Deputi Bidang Lingkungan Hidup', 'Direktorat Pengelolaan Sampah', 'PNS', 'III', '31-40'),
    ('44444444-4444-4444-4444-444444444443', '33333333-3333-3333-3333-333333333333', 'S3_ORGANIK', '2026-10-05', 50,   0.5, 'Poin/Kg', 'KLHK (2025)', 'EcoLimbah-v1-2026', 25,    'Deputi Bidang Lingkungan Hidup', 'Direktorat Pengelolaan Sampah', 'PNS', 'III', '31-40')
ON CONFLICT (id) DO UPDATE SET is_deleted = false;
