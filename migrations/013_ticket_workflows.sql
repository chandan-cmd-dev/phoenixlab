CREATE TABLE IF NOT EXISTS ticket_workflows (
    id              SERIAL PRIMARY KEY,
    ticket_id       INT NOT NULL REFERENCES tickets(id),
    workflow_type   TEXT NOT NULL,
    current_step    TEXT NOT NULL,
    step_data       JSONB DEFAULT '{}',
    started_at      TIMESTAMPTZ DEFAULT NOW(),
    completed_at    TIMESTAMPTZ,
    started_by      INT REFERENCES users(id),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workflows_ticket ON ticket_workflows(ticket_id);
CREATE INDEX IF NOT EXISTS idx_workflows_type ON ticket_workflows(workflow_type);
