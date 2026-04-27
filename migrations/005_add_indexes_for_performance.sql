-- Add indexes for performance optimization

-- Index on prices.game_id for faster game price lookups
CREATE INDEX IF NOT EXISTS idx_prices_game_id ON prices(game_id);

-- Index on prices.fetched_at for time-based price history queries
CREATE INDEX IF NOT EXISTS idx_prices_fetched_at ON prices(fetched_at DESC);

-- Index on prices.price_inr for price range queries
CREATE INDEX IF NOT EXISTS idx_prices_price_inr ON prices(price_inr);

-- Index on games.platform for platform filtering
CREATE INDEX IF NOT EXISTS idx_games_platform ON games(platform);

-- Index on deals.expires_at for active deal queries
CREATE INDEX IF NOT EXISTS idx_deals_expires_at ON deals(expires_at);

-- Index on deals.game_id for deal lookups
CREATE INDEX IF NOT EXISTS idx_deals_game_id ON deals(game_id);

-- Index on clicks.user_id and clicks.game_id for analytics queries
CREATE INDEX IF NOT EXISTS idx_clicks_user_game ON clicks(user_id, game_id);

-- Index on review_scores.game_id and review_scores.source for review lookups
CREATE INDEX IF NOT EXISTS idx_review_scores_game_source ON review_scores(game_id, source);

-- Index on wishlist.user_id for user wishlist queries
CREATE INDEX IF NOT EXISTS idx_wishlist_user_id ON wishlist(user_id);
