-- migrate:up
ALTER TABLE clients ADD COLUMN hourly_rate DECIMAL(10, 2) DEFAULT 0.00;

-- migrate:down
ALTER TABLE clients DROP COLUMN hourly_rate;
