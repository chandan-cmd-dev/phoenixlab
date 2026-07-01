-- Normalize ticket FK columns: store 0 as SQL NULL on every write.
--
-- Beego persists 0 (the int zero value) for an unassigned/unlinked ticket, which
-- violates the assigned_to -> users(id) and parent/linked_ticket_id -> tickets(id)
-- foreign keys. Full-row ORM updates (o.Update(t)) rewrite these columns on any
-- save, so the violation reappears wherever a NULL-FK ticket is loaded and saved.
-- A BEFORE INSERT OR UPDATE trigger converts 0 -> NULL for all write paths (ORM
-- inserts/updates, raw SQL) with no per-call-site handling.

CREATE OR REPLACE FUNCTION tickets_nullify_zero_fks() RETURNS trigger AS $$
BEGIN
    NEW.assigned_to := NULLIF(NEW.assigned_to, 0);
    NEW.parent_ticket_id := NULLIF(NEW.parent_ticket_id, 0);
    NEW.linked_ticket_id := NULLIF(NEW.linked_ticket_id, 0);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_tickets_nullify_zero_fks ON tickets;
CREATE TRIGGER trg_tickets_nullify_zero_fks
    BEFORE INSERT OR UPDATE ON tickets
    FOR EACH ROW EXECUTE PROCEDURE tickets_nullify_zero_fks();
