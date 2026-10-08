-- 1. Buat Set Faktor Versi 1
INSERT INTO factor_sets (id, version_name, valid_from, is_active)
VALUES (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'EcoLimbah-v1-2026',
    '2026-01-01',
    true
) ON CONFLICT (version_name) DO NOTHING;

-- 2. Isi Data Faktor Referensi Limbah
INSERT INTO factors (set_id, category_code, name, value, unit, source_ref)
VALUES 
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'S1_PLASTIK', 'Daur Ulang Plastik (PET/HDPE)', 1.500000, 'Poin/Kg', 'KLHK (2025)'),
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'S2_KERTAS', 'Daur Ulang Kertas/Kardus', 0.800000, 'Poin/Kg', 'KLHK (2025)'),
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'S3_ORGANIK', 'Komposting Limbah Organik', 0.500000, 'Poin/Kg', 'KLHK (2025)')
ON CONFLICT DO NOTHING;