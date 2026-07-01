-- Normalize sheet_adoptions.result_ticket_id: store 0 as SQL NULL on every write.
--
-- Same Beego limitation as tickets: ResultTicketId is a non-pointer int, so an
-- adoption with no resulting ticket is persisted as 0, which violates the
-- result_ticket_id -> tickets(id) foreign key. A BEFORE INSERT OR UPDATE trigger
-- converts 0 -> NULL for all write paths (see migration 017 for the tickets one).

CREATE OR REPLACE FUNCTION sheet_adoptions_nullify_zero_fks() RETURNS trigger AS $$
BEGIN
    NEW.result_ticket_id := NULLIF(NEW.result_ticket_id, 0);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_sheet_adoptions_nullify_zero_fks ON sheet_adoptions;
CREATE TRIGGER trg_sheet_adoptions_nullify_zero_fks
    BEFORE INSERT OR UPDATE ON sheet_adoptions
    FOR EACH ROW EXECUTE PROCEDURE sheet_adoptions_nullify_zero_fks();
