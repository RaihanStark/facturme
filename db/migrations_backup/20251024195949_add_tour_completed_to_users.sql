-- migrate:up
ALTER TABLE users ADD COLUMN tour_completed BOOLEAN DEFAULT FALSE;

-- migrate:down
ALTER TABLE users DROP COLUMN IF EXISTS tour_completed;

