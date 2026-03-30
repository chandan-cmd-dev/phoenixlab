-- Add fields to match Excel tracking sheet columns
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS diagnostic_code TEXT;
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS needed_part TEXT;
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS problem_description TEXT;
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS machine_purchase_price NUMERIC(10,2) DEFAULT 0;
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS part_number TEXT;
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS po_number TEXT;
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS case_number TEXT;
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS work_order_number TEXT;
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS part_arrived_fixed BOOLEAN DEFAULT FALSE;
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS defective_part_shipped TEXT;
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS case_finished BOOLEAN DEFAULT FALSE;

-- Remove unique constraint on serial_number (same SN can have multiple repair tickets)
ALTER TABLE tickets DROP CONSTRAINT IF EXISTS tickets_serial_number_key;
