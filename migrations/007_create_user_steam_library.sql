-- Create user_steam_library table to store owned games per user
CREATE TABLE IF NOT EXISTS user_steam_library (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    steam_app_id BIGINT NOT NULL,
    game_id BIGINT REFERENCES games(id) ON DELETE SET NULL,
    imported_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, steam_app_id)
);

-- Index for fast lookup of owned games by user
CREATE INDEX IF NOT EXISTS idx_user_steam_library_user_id ON user_steam_library(user_id);

-- Index for checking if a specific game is owned by a user
CREATE INDEX IF NOT EXISTS idx_user_steam_library_game_id ON user_steam_library(game_id);

-- Index for DLC flagging (find base games owned by user)
CREATE INDEX IF NOT EXISTS idx_user_steam_library_steam_app_id ON user_steam_library(steam_app_id);
