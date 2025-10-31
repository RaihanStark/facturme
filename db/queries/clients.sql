-- name: CreateClient :one
INSERT INTO clients (user_id, name, email, phone, company, address, hourly_rate, currency)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, user_id, name, email, phone, company, address, hourly_rate, currency, created_at, updated_at;

-- name: GetClientByID :one
SELECT id, user_id, name, email, phone, company, address, hourly_rate, currency, created_at, updated_at
FROM clients
WHERE id = $1 AND user_id = $2;

-- name: GetClientsByUserID :many
SELECT id, user_id, name, email, phone, company, address, hourly_rate, currency, created_at, updated_at
FROM clients
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateClient :one
UPDATE clients
SET name = $3, email = $4, phone = $5, company = $6, address = $7, hourly_rate = $8, currency = $9, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, name, email, phone, company, address, hourly_rate, currency, created_at, updated_at;

-- name: DeleteClient :exec
DELETE FROM clients
WHERE id = $1 AND user_id = $2;

-- name: DeleteDemoClients :exec
DELETE FROM clients
WHERE user_id = $1 AND (name LIKE '%🎭%' OR name LIKE '%(Demo)%');
