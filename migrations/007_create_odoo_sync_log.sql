CREATE TABLE IF NOT EXISTS odoo_sync_log (
    id              SERIAL PRIMARY KEY,
    ticket_id       INT REFERENCES tickets(id),
    direction       TEXT NOT NULL,
    odoo_ticket_id  TEXT,
    status          TEXT NOT NULL,
    payload         JSONB,
    error_message   TEXT,
    synced_at       TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_odoo_sync_ticket ON odoo_sync_log(ticket_id);
CREATE INDEX IF NOT EXISTS idx_odoo_sync_status ON odoo_sync_log(status);
