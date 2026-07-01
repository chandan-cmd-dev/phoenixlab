-- Make the ticket self/user foreign keys DEFERRABLE INITIALLY DEFERRED.
--
-- Unassigned tickets are inserted with assigned_to = 0 (Beego stores 0 for an
-- int column and cannot emit SQL NULL for a non-pointer field), then nulled in
-- the same transaction by insertTicket(). With an immediately-checked FK the
-- INSERT of 0 fails at statement time before the null-out runs. Deferring the
-- check to COMMIT — when the column is already NULL — lets that work while
-- keeping referential integrity for real (non-zero) references.

-- Clean up any 0 sentinels left by earlier code before re-adding the constraints.
UPDATE tickets SET assigned_to = NULL WHERE assigned_to = 0;
UPDATE tickets SET parent_ticket_id = NULL WHERE parent_ticket_id = 0;
UPDATE tickets SET linked_ticket_id = NULL WHERE linked_ticket_id = 0;

-- Postgres cannot ALTER a constraint to be deferrable, so drop then re-add each.
ALTER TABLE tickets DROP CONSTRAINT IF EXISTS tickets_assigned_to_fkey;
ALTER TABLE tickets ADD CONSTRAINT tickets_assigned_to_fkey
    FOREIGN KEY (assigned_to) REFERENCES users(id)
    DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE tickets DROP CONSTRAINT IF EXISTS tickets_parent_ticket_id_fkey;
ALTER TABLE tickets ADD CONSTRAINT tickets_parent_ticket_id_fkey
    FOREIGN KEY (parent_ticket_id) REFERENCES tickets(id)
    DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE tickets DROP CONSTRAINT IF EXISTS tickets_linked_ticket_id_fkey;
ALTER TABLE tickets ADD CONSTRAINT tickets_linked_ticket_id_fkey
    FOREIGN KEY (linked_ticket_id) REFERENCES tickets(id)
    DEFERRABLE INITIALLY DEFERRED;
