-- Performance optimization indexes for Phase 9
-- These indexes optimize the most common query patterns

-- Index for games title search (ILIKE pattern matching)
CREATE INDEX IF NOT EXISTS idx_games_title_lower ON games(LOWER(title));

-- Composite index for games filtering by platform and title
CREATE INDEX IF NOT EXISTS idx_games_platform_title ON games(platform, LOWER(title));

-- Index for deals filtering by is_active and expires_at
CREATE INDEX IF NOT EXISTS idx_deals_active_expires ON deals(is_active, expires_at);

-- Index for wishlist user filtering with game_id
CREATE INDEX IF NOT EXISTS idx_wishlist_user_game_target ON wishlists(user_id, target_price_inr);

-- Index for savings monthly breakdown queries
CREATE INDEX IF NOT EXISTS idx_user_savings_user_purchased ON user_savings(user_id, purchased_at DESC);

-- Index for review scores by game for aggregation
CREATE INDEX IF NOT EXISTS idx_review_scores_game_score ON review_scores(game_id, score);

-- Index for user library queries
CREATE INDEX IF NOT EXISTS idx_user_steam_library_user_game_unique ON user_steam_library(user_id, game_id);
