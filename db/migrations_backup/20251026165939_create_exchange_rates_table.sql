-- migrate:up
CREATE TABLE IF NOT EXISTS exchange_rates (
    id SERIAL PRIMARY KEY,
    base_currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    target_currency VARCHAR(3) NOT NULL,
    rate DECIMAL(20, 10) NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(base_currency, target_currency)
);

CREATE INDEX idx_exchange_rates_currencies ON exchange_rates(base_currency, target_currency);
CREATE INDEX idx_exchange_rates_updated_at ON exchange_rates(updated_at);

-- migrate:down
DROP INDEX IF EXISTS idx_exchange_rates_updated_at;
DROP INDEX IF EXISTS idx_exchange_rates_currencies;
DROP TABLE IF EXISTS exchange_rates;
