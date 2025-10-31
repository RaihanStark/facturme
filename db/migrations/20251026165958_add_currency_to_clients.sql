-- migrate:up
ALTER TABLE clients ADD COLUMN currency VARCHAR(3) NOT NULL DEFAULT 'USD';

CREATE INDEX idx_clients_currency ON clients(currency);

-- migrate:down
DROP INDEX IF EXISTS idx_clients_currency;
ALTER TABLE clients DROP COLUMN currency;
