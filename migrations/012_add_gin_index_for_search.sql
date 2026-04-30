-- Phase 10: Add GIN index for full-text search optimization
-- This enables efficient ILIKE pattern matching on game titles

-- Add GIN index on LOWER(title) for case-insensitive search
CREATE INDEX IF NOT EXISTS idx_games_title_gin ON games USING GIN (to_tsvector('english', title));

-- Add trigram index for partial/fuzzy matching (requires pg_trgm extension)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_games_title_trgm ON games USING GIN (title gin_trgm_ops);

-- Add index for combined platform + title search
CREATE INDEX IF NOT EXISTS idx_games_platform_title_gin ON games (platform, to_tsvector('english', title));
