ALTER TABLE tickets ADD COLUMN IF NOT EXISTS parent_ticket_id INT REFERENCES tickets(id);
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS linked_ticket_id INT REFERENCES tickets(id);

CREATE INDEX IF NOT EXISTS idx_tickets_parent ON tickets(parent_ticket_id);
CREATE INDEX IF NOT EXISTS idx_tickets_linked ON tickets(linked_ticket_id);
