ALTER TABLE games
ADD COLUMN IF NOT EXISTS steam_app_id BIGINT;

ALTER TABLE games
ADD COLUMN IF NOT EXISTS store_url TEXT NOT NULL DEFAULT '';

ALTER TABLE prices
ADD COLUMN IF NOT EXISTS store VARCHAR(64) NOT NULL DEFAULT 'seed';

ALTER TABLE prices
ADD COLUMN IF NOT EXISTS is_historical_low BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE prices
ADD COLUMN IF NOT EXISTS region VARCHAR(32) NOT NULL DEFAULT 'local';

CREATE INDEX IF NOT EXISTS idx_games_steam_app_id ON games(steam_app_id);
CREATE INDEX IF NOT EXISTS idx_prices_game_store_region ON prices(game_id, store, region, fetched_at DESC);

UPDATE games
SET steam_app_id = CASE slug
    WHEN 'cyberpunk-2077' THEN 1091500
    WHEN 'elden-ring' THEN 1245620
    WHEN 'red-dead-redemption-2' THEN 1174180
    WHEN 'stardew-valley' THEN 413150
    ELSE steam_app_id
END
WHERE steam_app_id IS NULL;

UPDATE games
SET store_url = CASE slug
    WHEN 'cyberpunk-2077' THEN 'https://store.steampowered.com/app/1091500'
    WHEN 'elden-ring' THEN 'https://store.steampowered.com/app/1245620'
    WHEN 'red-dead-redemption-2' THEN 'https://store.steampowered.com/app/1174180'
    WHEN 'stardew-valley' THEN 'https://store.steampowered.com/app/413150'
    WHEN 'alan-wake-remastered' THEN 'https://store.epicgames.com/'
    WHEN 'hades' THEN 'https://store.epicgames.com/'
    WHEN 'the-witcher-3' THEN 'https://www.gog.com/'
    WHEN 'control' THEN 'https://www.gog.com/'
    ELSE store_url
END
WHERE store_url = '';

UPDATE prices p
SET store = LOWER(g.platform),
    is_historical_low = TRUE
FROM games g
WHERE p.game_id = g.id
  AND p.store = 'seed';
