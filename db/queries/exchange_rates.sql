-- name: GetExchangeRate :one
SELECT id, base_currency, target_currency, rate, updated_at
FROM exchange_rates
WHERE base_currency = $1 AND target_currency = $2;

-- name: UpsertExchangeRate :exec
INSERT INTO exchange_rates (base_currency, target_currency, rate, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (base_currency, target_currency)
DO UPDATE SET rate = $3, updated_at = NOW();

-- name: GetAllExchangeRates :many
SELECT id, base_currency, target_currency, rate, updated_at
FROM exchange_rates
ORDER BY base_currency, target_currency;

-- name: DeleteOldExchangeRates :exec
DELETE FROM exchange_rates
WHERE updated_at < $1;
