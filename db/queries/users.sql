-- name: CreateUser :one
INSERT INTO users (email, password_hash, name, verification_token, verification_token_expires)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, email, name, email_verified, onboarding_completed, tour_completed, currency, date_format, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, name, email_verified, onboarding_completed, tour_completed, currency, date_format, created_at, updated_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, name, email_verified, onboarding_completed, tour_completed, currency, date_format, created_at, updated_at
FROM users
WHERE id = $1;

-- name: CompleteOnboarding :one
UPDATE users
SET onboarding_completed = TRUE,
    currency = $2,
    date_format = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, email, name, email_verified, onboarding_completed, tour_completed, currency, date_format, created_at, updated_at;

-- name: CompleteTour :one
UPDATE users
SET tour_completed = TRUE,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, email, name, email_verified, onboarding_completed, tour_completed, currency, date_format, created_at, updated_at;

-- name: GetUserByVerificationToken :one
SELECT id, email, name, email_verified, verification_token_expires, onboarding_completed, tour_completed, currency, date_format, created_at, updated_at
FROM users
WHERE verification_token = $1;

-- name: VerifyUserEmail :one
UPDATE users
SET email_verified = TRUE,
    verification_token = NULL,
    verification_token_expires = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, email, name, email_verified, onboarding_completed, tour_completed, currency, date_format, created_at, updated_at;

-- name: UpdateVerificationToken :one
UPDATE users
SET verification_token = $2,
    verification_token_expires = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND email_verified = FALSE
RETURNING id, email, name, email_verified, onboarding_completed, tour_completed, currency, date_format, created_at, updated_at;

-- name: UpdatePasswordResetToken :one
UPDATE users
SET password_reset_token = $2,
    password_reset_token_expires = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE email = $1
RETURNING id, email, name;

-- name: GetUserByPasswordResetToken :one
SELECT id, email, name, password_reset_token_expires
FROM users
WHERE password_reset_token = $1;

-- name: ResetPassword :one
UPDATE users
SET password_hash = $2,
    password_reset_token = NULL,
    password_reset_token_expires = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, email, name, email_verified, onboarding_completed, tour_completed, currency, date_format, created_at, updated_at;

-- name: ChangePassword :one
UPDATE users
SET password_hash = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, email, name, email_verified, onboarding_completed, tour_completed, currency, date_format, created_at, updated_at;

-- name: UpdateUserCurrency :one
UPDATE users
SET currency = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, email, name, email_verified, onboarding_completed, tour_completed, currency, date_format, created_at, updated_at;
