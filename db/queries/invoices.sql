-- name: CreateInvoice :one
INSERT INTO invoices (user_id, client_id, invoice_number, issue_date, due_date, status, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, user_id, client_id, invoice_number, issue_date, due_date, status, notes, created_at, updated_at;

-- name: GetInvoiceByID :one
SELECT id, user_id, client_id, invoice_number, issue_date, due_date, status, notes, created_at, updated_at
FROM invoices
WHERE id = $1 AND user_id = $2;

-- name: GetInvoicesByUserID :many
SELECT id, user_id, client_id, invoice_number, issue_date, due_date, status, notes, created_at, updated_at
FROM invoices
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateInvoice :one
UPDATE invoices
SET client_id = $3, invoice_number = $4, issue_date = $5, due_date = $6, status = $7, notes = $8, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, client_id, invoice_number, issue_date, due_date, status, notes, created_at, updated_at;

-- name: UpdateInvoiceStatus :one
UPDATE invoices
SET status = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, client_id, invoice_number, issue_date, due_date, status, notes, created_at, updated_at;

-- name: DeleteInvoice :exec
DELETE FROM invoices
WHERE id = $1 AND user_id = $2;

-- name: AddTimeEntryToInvoice :exec
INSERT INTO invoice_time_entries (invoice_id, time_entry_id)
VALUES ($1, $2);

-- name: GetInvoiceTimeEntries :many
SELECT te.id, te.user_id, te.client_id, te.date, te.hours, te.description, te.hourly_rate, te.created_at, te.updated_at
FROM time_entries te
INNER JOIN invoice_time_entries ite ON te.id = ite.time_entry_id
WHERE ite.invoice_id = $1;

-- name: GetAvailableTimeEntriesForClient :many
SELECT te.id, te.user_id, te.client_id, te.date, te.hours, te.description, te.hourly_rate, te.created_at, te.updated_at
FROM time_entries te
WHERE te.client_id = $1
  AND te.user_id = $2
  AND NOT EXISTS (
    SELECT 1 FROM invoice_time_entries ite WHERE ite.time_entry_id = te.id
  )
ORDER BY te.date DESC;
