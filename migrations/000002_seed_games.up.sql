INSERT INTO games (id, slug, name, genre, platform) VALUES
    ('00000000-0000-0000-0000-000000000001', 'mobile-legends', 'Mobile Legends: Bang Bang', 'MOBA', 'Mobile'),
    ('00000000-0000-0000-0000-000000000002', 'dota-2', 'Dota 2', 'MOBA', 'PC'),
    ('00000000-0000-0000-0000-000000000003', 'league-of-legends', 'League of Legends', 'MOBA', 'PC'),
    ('00000000-0000-0000-0000-000000000004', 'valorant', 'Valorant', 'FPS', 'PC'),
    ('00000000-0000-0000-0000-000000000005', 'counter-strike-2', 'Counter-Strike 2', 'FPS', 'PC'),
    ('00000000-0000-0000-0000-000000000006', 'ea-sports-fc', 'EA Sports FC 24', 'SPORTS', 'Multiplatform'),
    ('00000000-0000-0000-0000-000000000007', 'efootball', 'eFootball 2024', 'SPORTS', 'Multiplatform'),
    ('00000000-0000-0000-0000-000000000008', 'pubg-mobile', 'PUBG Mobile', 'BATTLE_ROYALE', 'Mobile'),
    ('00000000-0000-0000-0000-000000000009', 'free-fire', 'Free Fire', 'BATTLE_ROYALE', 'Mobile'),
    ('00000000-0000-0000-0000-000000000010', 'apex-legends', 'Apex Legends', 'BATTLE_ROYALE', 'PC/Console')
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    genre = EXCLUDED.genre,
    platform = EXCLUDED.platform;
