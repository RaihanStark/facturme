-- migrate:up
-- Drop the old global unique constraint on invoice_number
ALTER TABLE invoices DROP CONSTRAINT IF EXISTS invoices_invoice_number_key;

-- Add a composite unique constraint: invoice_number must be unique per user
ALTER TABLE invoices ADD CONSTRAINT invoices_user_invoice_number_unique UNIQUE (user_id, invoice_number);

-- migrate:down
-- Drop the composite unique constraint
ALTER TABLE invoices DROP CONSTRAINT IF EXISTS invoices_user_invoice_number_unique;

-- Restore the old global unique constraint
ALTER TABLE invoices ADD CONSTRAINT invoices_invoice_number_key UNIQUE (invoice_number);

