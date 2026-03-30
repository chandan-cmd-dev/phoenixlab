CREATE TABLE IF NOT EXISTS tickets (
    id                  SERIAL PRIMARY KEY,
    serial_number       TEXT NOT NULL UNIQUE,
    ir_number           TEXT,
    upc                 TEXT,
    model               TEXT,
    brand               TEXT,
    branch_id           INT REFERENCES branches(id),
    warranty_status     TEXT NOT NULL DEFAULT 'out_of_warranty',
    issue_description   TEXT,
    issue_category      TEXT,
    assigned_to         INT REFERENCES users(id),
    priority            TEXT DEFAULT 'normal',
    status              TEXT DEFAULT 'open',
    parts_cost          NUMERIC(10,2) DEFAULT 0,
    labour_cost         NUMERIC(10,2) DEFAULT 0,
    total_cost          NUMERIC(10,2) GENERATED ALWAYS AS (parts_cost + labour_cost) STORED,
    courier_name        TEXT,
    courier_tracking    TEXT,
    tracking_link       TEXT,
    return_part         BOOLEAN DEFAULT FALSE,
    return_tracking     TEXT,
    customer_name       TEXT,
    customer_email      TEXT,
    customer_phone      TEXT,
    odoo_ticket_id      TEXT,
    odoo_synced_at      TIMESTAMPTZ,
    received_at         TIMESTAMPTZ DEFAULT NOW(),
    resolved_at         TIMESTAMPTZ,
    due_date            DATE,
    created_by          INT REFERENCES users(id),
    version             INT DEFAULT 1,
    notes               TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tickets_branch    ON tickets(branch_id);
CREATE INDEX IF NOT EXISTS idx_tickets_status    ON tickets(status);
CREATE INDEX IF NOT EXISTS idx_tickets_assigned  ON tickets(assigned_to);
CREATE INDEX IF NOT EXISTS idx_tickets_warranty  ON tickets(warranty_status);
CREATE INDEX IF NOT EXISTS idx_tickets_sn        ON tickets(serial_number);
CREATE INDEX IF NOT EXISTS idx_tickets_created   ON tickets(created_at DESC);
