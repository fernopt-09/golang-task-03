-- every GetRates call adds one row
CREATE TABLE IF NOT EXISTS rates (
    id BIGSERIAL PRIMARY KEY,
    ask NUMERIC(18, 8) NOT NULL,
    bid NUMERIC(18, 8) NOT NULL,
    -- time of the rate
    exchange_timestamp TIMESTAMPTZ NOT NULL,
    calculation_method VARCHAR(32) NOT NULL,
    -- N and M that were used for calculation
    params JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- to get the latest rates quickly
CREATE INDEX IF NOT EXISTS idx_rates_created_at ON rates(created_at DESC);
