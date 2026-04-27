-- Create user_savings table to track purchase savings
CREATE TABLE IF NOT EXISTS user_savings (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    game_id BIGINT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    game_title VARCHAR(255) NOT NULL,
    paid_price_inr INT NOT NULL,
    original_price_inr INT NOT NULL,
    saved_amount_inr INT GENERATED ALWAYS AS (original_price_inr - paid_price_inr) STORED,
    purchased_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for user's savings history
CREATE INDEX idx_user_savings_user_id ON user_savings(user_id);

-- Index for game_id lookups
CREATE INDEX idx_user_savings_game_id ON user_savings(game_id);

-- Index for purchased_at for monthly breakdown queries
CREATE INDEX idx_user_savings_purchased_at ON user_savings(purchased_at);
