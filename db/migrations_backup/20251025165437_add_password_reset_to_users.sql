-- migrate:up
ALTER TABLE users ADD COLUMN password_reset_token VARCHAR(255);
ALTER TABLE users ADD COLUMN password_reset_token_expires TIMESTAMP;

CREATE INDEX idx_users_password_reset_token ON users(password_reset_token);

-- migrate:down
DROP INDEX IF EXISTS idx_users_password_reset_token;
ALTER TABLE users DROP COLUMN IF EXISTS password_reset_token_expires;
ALTER TABLE users DROP COLUMN IF EXISTS password_reset_token;

