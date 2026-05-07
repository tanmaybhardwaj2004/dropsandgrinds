-- Create sales calendar table for tracking known sale events
CREATE TABLE IF NOT EXISTS sales_calendar (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    platform VARCHAR(50) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    is_recurring BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for finding active sales
CREATE INDEX IF NOT EXISTS idx_sales_calendar_dates ON sales_calendar(start_date, end_date);

-- Index for platform filtering
CREATE INDEX IF NOT EXISTS idx_sales_calendar_platform ON sales_calendar(platform);

-- Index for recurring sales
CREATE INDEX IF NOT EXISTS idx_sales_calendar_recurring ON sales_calendar(is_recurring);

-- Insert known Steam sale patterns for 2026
INSERT INTO sales_calendar (name, platform, start_date, end_date, is_recurring)
SELECT name, platform, start_date::date, end_date::date, is_recurring
FROM (
    VALUES
    ('Steam Spring Sale', 'Steam', '2026-03-13', '2026-03-26', TRUE),
    ('Steam Summer Sale', 'Steam', '2026-06-25', '2026-07-09', TRUE),
    ('Steam Halloween Sale', 'Steam', '2026-10-28', '2026-11-01', TRUE),
    ('Steam Autumn Sale', 'Steam', '2026-11-24', '2026-12-01', TRUE),
    ('Steam Winter Sale', 'Steam', '2026-12-19', '2027-01-02', TRUE),
    ('Epic Mega Sale', 'Epic Games', '2026-05-16', '2026-06-13', TRUE),
    ('Epic Holiday Sale', 'Epic Games', '2026-12-15', '2027-01-05', TRUE),
    ('GOG Summer Sale', 'GOG', '2026-06-01', '2026-06-15', TRUE),
    ('GOG Winter Sale', 'GOG', '2026-12-01', '2026-12-15', TRUE)
) AS seed(name, platform, start_date, end_date, is_recurring)
WHERE NOT EXISTS (
    SELECT 1
    FROM sales_calendar sc
    WHERE sc.name = seed.name
      AND sc.platform = seed.platform
      AND sc.start_date = seed.start_date::date
);

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_sales_calendar_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_sales_calendar_updated_at ON sales_calendar;

CREATE TRIGGER trigger_update_sales_calendar_updated_at
    BEFORE UPDATE ON sales_calendar
    FOR EACH ROW
    EXECUTE FUNCTION update_sales_calendar_updated_at();
