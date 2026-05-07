-- Add steam_id column to users table
ALTER TABLE users ADD COLUMN steam_id VARCHAR(20) UNIQUE;
ALTER TABLE users ADD COLUMN consent_analytics BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN consent_alerts BOOLEAN DEFAULT FALSE;
