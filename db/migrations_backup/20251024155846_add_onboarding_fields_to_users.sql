-- migrate:up
ALTER TABLE users ADD COLUMN onboarding_completed BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN plan VARCHAR(50) DEFAULT 'trial';
ALTER TABLE users ADD COLUMN plan_end_date TIMESTAMP DEFAULT (CURRENT_TIMESTAMP + INTERVAL '7 days');
ALTER TABLE users ADD COLUMN currency VARCHAR(10) DEFAULT 'USD';
ALTER TABLE users ADD COLUMN date_format VARCHAR(20) DEFAULT 'MM/DD/YYYY';

CREATE INDEX idx_users_plan ON users(plan);
CREATE INDEX idx_users_plan_end_date ON users(plan_end_date);

-- migrate:down
DROP INDEX IF EXISTS idx_users_plan_end_date;
DROP INDEX IF EXISTS idx_users_plan;

ALTER TABLE users DROP COLUMN IF EXISTS date_format;
ALTER TABLE users DROP COLUMN IF EXISTS currency;
ALTER TABLE users DROP COLUMN IF EXISTS plan_end_date;
ALTER TABLE users DROP COLUMN IF EXISTS plan;
ALTER TABLE users DROP COLUMN IF EXISTS onboarding_completed;

