CREATE TABLE IF NOT EXISTS games (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    platform VARCHAR(64) NOT NULL,
    cover_url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS prices (
    id BIGSERIAL PRIMARY KEY,
    game_id BIGINT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    price_inr INTEGER NOT NULL CHECK (price_inr >= 0),
    fetched_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS deals (
    id BIGSERIAL PRIMARY KEY,
    game_id BIGINT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    discount_percent INTEGER NOT NULL CHECK (discount_percent >= 0 AND discount_percent <= 100),
    original_inr INTEGER NOT NULL CHECK (original_inr >= 0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    cached_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_games_platform ON games(platform);
CREATE INDEX IF NOT EXISTS idx_prices_game_id_fetched_at ON prices(game_id, fetched_at DESC);
CREATE INDEX IF NOT EXISTS idx_deals_game_id_active ON deals(game_id, is_active);

INSERT INTO games (title, slug, platform, cover_url)
VALUES
    ('Cyberpunk 2077', 'cyberpunk-2077', 'Steam', 'https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1091500/header.jpg'),
    ('Elden Ring', 'elden-ring', 'Steam', 'https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1245620/header.jpg'),
    ('Alan Wake Remastered', 'alan-wake-remastered', 'Epic Games', 'https://cdn2.unrealengine.com/egs-alanwakeRemastered-remedyentertainment-s2-1200x1600-b6f4e150f584.jpg'),
    ('The Witcher 3', 'the-witcher-3', 'GOG', 'https://images.gog-statics.com/1445585698466185bb212ae17d45e5df5a36371c10787a740703e2c340d12e8b_glx_logo_284x400.png'),
    ('Red Dead Redemption 2', 'red-dead-redemption-2', 'Steam', 'https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1174180/header.jpg'),
    ('Hades', 'hades', 'Epic Games', 'https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1145360/header.jpg'),
    ('Stardew Valley', 'stardew-valley', 'Steam', 'https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/413150/header.jpg'),
    ('Control', 'control', 'GOG', 'https://images.gog-statics.com/97adbbdcdab1889814c8d5c4142f2edab2838bed820d885b546bed1d5a711422_glx_logo_284x400.png')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO prices (game_id, price_inr, fetched_at)
SELECT g.id, d.price_inr, CURRENT_TIMESTAMP
FROM games g
JOIN (
    VALUES
        ('cyberpunk-2077', 1499),
        ('elden-ring', 2399),
        ('alan-wake-remastered', 450),
        ('the-witcher-3', 299),
        ('red-dead-redemption-2', 999),
        ('hades', 549),
        ('stardew-valley', 384),
        ('control', 899)
) AS d(slug, price_inr) ON d.slug = g.slug
WHERE NOT EXISTS (
    SELECT 1 FROM prices p WHERE p.game_id = g.id
);

INSERT INTO deals (game_id, discount_percent, original_inr, is_active, cached_at)
SELECT g.id, d.discount_percent, d.original_inr, TRUE, CURRENT_TIMESTAMP
FROM games g
JOIN (
    VALUES
        ('cyberpunk-2077', 50, 2999),
        ('elden-ring', 40, 3999),
        ('alan-wake-remastered', 70, 1500),
        ('the-witcher-3', 70, 999),
        ('red-dead-redemption-2', 69, 3199),
        ('hades', 50, 1099),
        ('stardew-valley', 20, 479),
        ('control', 70, 2999)
) AS d(slug, discount_percent, original_inr) ON d.slug = g.slug
WHERE NOT EXISTS (
    SELECT 1 FROM deals dl WHERE dl.game_id = g.id AND dl.is_active = TRUE
);