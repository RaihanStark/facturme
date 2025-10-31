-- name: CreateTimeEntry :one
INSERT INTO time_entries (user_id, client_id, date, hours, description, hourly_rate)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, client_id, date, hours, description, hourly_rate, created_at, updated_at;

-- name: GetTimeEntryByID :one
SELECT id, user_id, client_id, date, hours, description, hourly_rate, created_at, updated_at
FROM time_entries
WHERE id = $1 AND user_id = $2;

-- name: GetTimeEntriesByUserID :many
SELECT id, user_id, client_id, date, hours, description, hourly_rate, created_at, updated_at
FROM time_entries
WHERE user_id = $1
ORDER BY date DESC, created_at DESC;

-- name: GetTimeEntriesByClientID :many
SELECT id, user_id, client_id, date, hours, description, hourly_rate, created_at, updated_at
FROM time_entries
WHERE client_id = $1 AND user_id = $2
ORDER BY date DESC, created_at DESC;

-- name: UpdateTimeEntry :one
UPDATE time_entries
SET client_id = $3, date = $4, hours = $5, description = $6, hourly_rate = $7, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, client_id, date, hours, description, hourly_rate, created_at, updated_at;

-- name: DeleteTimeEntry :exec
DELETE FROM time_entries
WHERE id = $1 AND user_id = $2;

-- name: GetTimeEntriesByDateRange :many
SELECT date, CAST(SUM(CAST(hours AS DECIMAL)) AS TEXT) as total_hours
FROM time_entries
WHERE user_id = $1 AND date >= $2 AND date <= $3
GROUP BY date
ORDER BY date ASC;

-- name: GetDetailedTimeEntriesByDateRange :many
SELECT id, user_id, client_id, date, hours, description, hourly_rate, created_at, updated_at
FROM time_entries
WHERE user_id = $1 AND date >= $2 AND date <= $3
ORDER BY date DESC, created_at DESC;
