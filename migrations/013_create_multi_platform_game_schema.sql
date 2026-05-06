-- Enhanced multi-platform game aggregation schema
-- Supports Steam, Epic Games, Xbox, PlayStation, Nintendo, GreenManGaming, Fanatical, Humble Bundle, and Indian stores

-- Enhanced games table with more comprehensive metadata
ALTER TABLE games ADD COLUMN IF NOT EXISTS external_id VARCHAR(255);
ALTER TABLE games ADD COLUMN IF NOT EXISTS store_url TEXT;
ALTER TABLE games ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE games ADD COLUMN IF NOT EXISTS release_date DATE;
ALTER TABLE games ADD COLUMN IF NOT EXISTS developer VARCHAR(255);
ALTER TABLE games ADD COLUMN IF NOT EXISTS publisher VARCHAR(255);
ALTER TABLE games ADD COLUMN IF NOT EXISTS genres TEXT[];
ALTER TABLE games ADD COLUMN IF NOT EXISTS platforms TEXT[];
ALTER TABLE games ADD COLUMN IF NOT EXISTS screenshots TEXT[];
ALTER TABLE games ADD COLUMN IF NOT EXISTS trailers TEXT[];
ALTER TABLE games ADD COLUMN IF NOT EXISTS system_requirements JSONB;
ALTER TABLE games ADD COLUMN IF NOT EXISTS editions JSONB;
ALTER TABLE games ADD COLUMN IF NOT EXISTS dlc_info JSONB;
ALTER TABLE games ADD COLUMN IF NOT EXISTS rating VARCHAR(10);
ALTER TABLE games ADD COLUMN IF NOT EXISTS user_rating DECIMAL(3,1);
ALTER TABLE games ADD COLUMN IF NOT EXISTS is_dlc BOOLEAN DEFAULT FALSE;
ALTER TABLE games ADD COLUMN IF NOT EXISTS parent_game_id BIGINT REFERENCES games(id);
ALTER TABLE games ADD COLUMN IF NOT EXISTS region VARCHAR(50) DEFAULT 'IN';
ALTER TABLE games ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT TRUE;
ALTER TABLE games ADD COLUMN IF NOT EXISTS last_price_update TIMESTAMP;

-- Create stores table for multi-platform support
CREATE TABLE IF NOT EXISTS stores (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE,
    website_url TEXT NOT NULL,
    api_endpoint TEXT,
    api_key_required BOOLEAN DEFAULT FALSE,
    supports_india BOOLEAN DEFAULT TRUE,
    region VARCHAR(50) DEFAULT 'IN',
    currency VARCHAR(3) DEFAULT 'INR',
    conversion_rate_to_inr DECIMAL(10,4) DEFAULT 1.0,
    logo_url TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Enhanced prices table with store and region support
ALTER TABLE prices ADD COLUMN IF NOT EXISTS store_id BIGINT REFERENCES stores(id);
ALTER TABLE prices ADD COLUMN IF NOT EXISTS external_id VARCHAR(255);
ALTER TABLE prices ADD COLUMN IF NOT EXISTS region VARCHAR(50) DEFAULT 'IN';
ALTER TABLE prices ADD COLUMN IF NOT EXISTS currency VARCHAR(3) DEFAULT 'INR';
ALTER TABLE prices ADD COLUMN IF NOT EXISTS original_price DECIMAL(10,2);
ALTER TABLE prices ADD COLUMN IF NOT EXISTS discount_amount DECIMAL(10,2);
ALTER TABLE prices ADD COLUMN IF NOT EXISTS is_available BOOLEAN DEFAULT TRUE;
ALTER TABLE prices ADD COLUMN IF NOT EXISTS stock_status VARCHAR(50);
ALTER TABLE prices ADD COLUMN IF NOT EXISTS deal_type VARCHAR(50) DEFAULT 'regular'; -- regular, bundle, preorder, etc.

-- Create price history table for tracking
CREATE TABLE IF NOT EXISTS price_history (
    id BIGSERIAL PRIMARY KEY,
    game_id BIGINT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    store_id BIGINT NOT NULL REFERENCES stores(id),
    external_id VARCHAR(255),
    price_inr DECIMAL(10,2) NOT NULL,
    original_price DECIMAL(10,2),
    discount_percent INTEGER,
    region VARCHAR(50) DEFAULT 'IN',
    currency VARCHAR(3) DEFAULT 'INR',
    is_available BOOLEAN DEFAULT TRUE,
    recorded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create game metadata table for additional information
CREATE TABLE IF NOT EXISTS game_metadata (
    id BIGSERIAL PRIMARY KEY,
    game_id BIGINT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    key VARCHAR(100) NOT NULL,
    value TEXTB,
    data_type VARCHAR(50) DEFAULT 'text', -- text, json, array, etc.
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(game_id, key)
);

-- Create store integrations table for API configurations
CREATE TABLE IF NOT EXISTS store_integrations (
    id BIGSERIAL PRIMARY KEY,
    store_id BIGINT NOT NULL REFERENCES stores(id),
    integration_type VARCHAR(50) NOT NULL, -- api, webhook, scraper
    config JSONB NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    last_sync TIMESTAMP,
    sync_frequency INTEGER DEFAULT 3600, -- seconds
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create regional pricing table
CREATE TABLE IF NOT EXISTS regional_pricing (
    id BIGSERIAL PRIMARY KEY,
    game_id BIGINT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    store_id BIGINT NOT NULL REFERENCES stores(id),
    region VARCHAR(50) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    original_price DECIMAL(10,2),
    discount_percent INTEGER,
    is_available BOOLEAN DEFAULT TRUE,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(game_id, store_id, region)
);

-- Create deal alerts table for wishlist notifications
CREATE TABLE IF NOT EXISTS deal_alerts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    game_id BIGINT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    target_price DECIMAL(10,2) NOT NULL,
    store_id BIGINT REFERENCES stores(id),
    region VARCHAR(50) DEFAULT 'IN',
    currency VARCHAR(3) DEFAULT 'INR',
    is_active BOOLEAN DEFAULT TRUE,
    notification_sent BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    triggered_at TIMESTAMP,
    UNIQUE(user_id, game_id, store_id)
);

-- Create trending deals table
CREATE TABLE IF NOT EXISTS trending_deals (
    id BIGSERIAL PRIMARY KEY,
    game_id BIGINT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    store_id BIGINT NOT NULL REFERENCES stores(id),
    trend_score DECIMAL(5,2) NOT NULL,
    view_count INTEGER DEFAULT 0,
    click_count INTEGER DEFAULT 0,
    conversion_rate DECIMAL(5,2) DEFAULT 0.0,
    trend_period VARCHAR(20) DEFAULT '24h', -- 1h, 6h, 24h, 7d
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create Indian payment offers table
CREATE TABLE IF NOT EXISTS indian_payment_offers (
    id BIGSERIAL PRIMARY KEY,
    store_id BIGINT NOT NULL REFERENCES stores(id),
    offer_type VARCHAR(50) NOT NULL, -- upi_discount, card_cashback, wallet_bonus
    provider VARCHAR(100), -- phonepe, gpay, paytm, etc.
    description TEXT,
    discount_percent INTEGER,
    max_discount_amount DECIMAL(10,2),
    min_order_amount DECIMAL(10,2),
    valid_from TIMESTAMP,
    valid_until TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Insert default stores
INSERT INTO stores (name, slug, website_url, api_endpoint, supports_india, region, currency) VALUES
('Steam', 'steam', 'https://store.steampowered.com', 'https://store.steampowered.com/api', TRUE, 'IN', 'INR'),
('Epic Games', 'epic', 'https://store.epicgames.com', 'https://store.epicgames.com/graphql', TRUE, 'IN', 'INR'),
('GOG', 'gog', 'https://www.gog.com', 'https://api.gog.com', TRUE, 'IN', 'INR'),
('Xbox', 'xbox', 'https://www.xbox.com', 'https://displaycatalog.mp.microsoft.com', TRUE, 'IN', 'INR'),
('PlayStation', 'playstation', 'https://store.playstation.com', 'https://store.playstation.com', TRUE, 'IN', 'INR'),
('Nintendo', 'nintendo', 'https://www.nintendo.com', 'https://api.nintendo.com', TRUE, 'IN', 'INR'),
('GreenManGaming', 'greenmangaming', 'https://www.greenmangaming.com', 'https://www.greenmangaming.com/api', TRUE, 'IN', 'INR'),
('Fanatical', 'fanatical', 'https://www.fanatical.com', 'https://www.fanatical.com/api', TRUE, 'IN', 'INR'),
('Humble Bundle', 'humble', 'https://www.humblebundle.com', 'https://www.humblebundle.com/api', TRUE, 'IN', 'INR'),
('Instant Gaming', 'instantgaming', 'https://www.instant-gaming.com', 'https://www.instant-gaming.com/api', TRUE, 'IN', 'INR'),
('Gamivo', 'gamivo', 'https://www.gamivo.com', 'https://www.gamivo.com/api', TRUE, 'IN', 'INR'),
('Eneba', 'eneba', 'https://www.eneba.com', 'https://www.eneba.com/api', TRUE, 'IN', 'INR')
ON CONFLICT (slug) DO NOTHING;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_games_external_id ON games(external_id);
CREATE INDEX IF NOT EXISTS idx_games_platforms ON games USING GIN(platforms);
CREATE INDEX IF NOT EXISTS idx_games_genres ON games USING GIN(genres);
CREATE INDEX IF NOT EXISTS idx_games_release_date ON games(release_date DESC);
CREATE INDEX IF NOT EXISTS idx_prices_store_id ON prices(store_id);
CREATE INDEX IF NOT EXISTS idx_prices_game_store ON prices(game_id, store_id);
CREATE INDEX IF NOT EXISTS idx_price_history_game_store ON price_history(game_id, store_id, recorded_at DESC);
CREATE INDEX IF NOT EXISTS idx_regional_pricing_game_region ON regional_pricing(game_id, region);
CREATE INDEX IF NOT EXISTS idx_deal_alerts_user_game ON deal_alerts(user_id, game_id);
CREATE INDEX IF NOT EXISTS idx_trending_deals_score ON trending_deals(trend_score DESC, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_trending_deals_active ON trending_deals(is_active, updated_at DESC);
