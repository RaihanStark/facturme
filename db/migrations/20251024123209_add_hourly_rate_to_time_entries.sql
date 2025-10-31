-- migrate:up
ALTER TABLE time_entries ADD COLUMN hourly_rate DECIMAL(10, 2) DEFAULT 0.00;

-- migrate:down
ALTER TABLE time_entries DROP COLUMN hourly_rate;

