-- Create clicks table for analytics tracking
CREATE TABLE IF NOT EXISTS clicks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    game_id BIGINT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    platform VARCHAR(50) NOT NULL,
    clicked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for user's click history
CREATE INDEX IF NOT EXISTS idx_clicks_user_id ON clicks(user_id);

-- Index for game_id lookups
CREATE INDEX IF NOT EXISTS idx_clicks_game_id ON clicks(game_id);

-- Index for clicked_at for time-based queries
CREATE INDEX IF NOT EXISTS idx_clicks_clicked_at ON clicks(clicked_at);
