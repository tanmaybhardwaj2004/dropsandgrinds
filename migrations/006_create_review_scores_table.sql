-- Create review_scores table for multi-source review aggregation
CREATE TABLE IF NOT EXISTS review_scores (
    id BIGSERIAL PRIMARY KEY,
    game_id BIGINT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    source VARCHAR(64) NOT NULL, -- metacritic, opencritic, steam, ign, gamespot
    score INTEGER NOT NULL CHECK (score >= 0 AND score <= 100),
    url TEXT,
    fetched_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(game_id, source)
);

-- Index for efficient review score lookups
CREATE INDEX IF NOT EXISTS idx_review_scores_game_source ON review_scores(game_id, source);
