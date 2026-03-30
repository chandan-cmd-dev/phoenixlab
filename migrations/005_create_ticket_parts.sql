CREATE TABLE IF NOT EXISTS ticket_parts (
    id              SERIAL PRIMARY KEY,
    ticket_id       INT REFERENCES tickets(id) ON DELETE CASCADE,
    part_number     TEXT NOT NULL,
    description     TEXT,
    quantity        INT DEFAULT 1,
    unit_cost       NUMERIC(10,2) DEFAULT 0,
    status          TEXT DEFAULT 'pending',
    ordered_at      TIMESTAMPTZ,
    received_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ticket_parts_ticket ON ticket_parts(ticket_id);
CREATE INDEX IF NOT EXISTS idx_ticket_parts_status ON ticket_parts(status);
