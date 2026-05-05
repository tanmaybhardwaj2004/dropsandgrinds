-- Add indexes for performance optimization

-- Index on prices.game_id for faster game price lookups
CREATE INDEX IF NOT EXISTS idx_prices_game_id ON prices(game_id);

-- Index on prices.fetched_at for time-based price history queries
CREATE INDEX IF NOT EXISTS idx_prices_fetched_at ON prices(fetched_at DESC);

-- Index on prices.price_inr for price range queries
CREATE INDEX IF NOT EXISTS idx_prices_price_inr ON prices(price_inr);

-- Index on games.platform for platform filtering
CREATE INDEX IF NOT EXISTS idx_games_platform ON games(platform);

-- Index on deals.game_id for deal lookups
CREATE INDEX IF NOT EXISTS idx_deals_game_id ON deals(game_id);
