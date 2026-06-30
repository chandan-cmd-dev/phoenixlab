ALTER TABLE tickets ADD COLUMN IF NOT EXISTS custom_fields JSONB;

CREATE TABLE IF NOT EXISTS google_tokens (
    id            SERIAL PRIMARY KEY,
    access_token  TEXT NOT NULL,
    refresh_token TEXT,
    token_type    VARCHAR(50),
    expiry        TIMESTAMPTZ,
    scope         TEXT,
    account_email VARCHAR(200),
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sheet_connections (
    id                    SERIAL PRIMARY KEY,
    spreadsheet_id        VARCHAR(200) NOT NULL,
    spreadsheet_title     VARCHAR(500),
    tab_name              VARCHAR(200),
    brand                 VARCHAR(100),
    branch_id             INT REFERENCES branches(id),
    status                VARCHAR(20) DEFAULT 'draft',
    sync_direction        VARCHAR(20) DEFAULT 'two_way',
    conflict_policy       VARCHAR(20) DEFAULT 'review',
    header_row            INT DEFAULT 0,
    identity_key          VARCHAR(300) DEFAULT 'SerialNumber,IssueDescription',
    auto_sync_enabled     BOOLEAN DEFAULT FALSE,
    sync_interval_minutes INT DEFAULT 15,
    last_auto_run_at      TIMESTAMPTZ,
    last_auto_status      VARCHAR(20),
    last_auto_message     TEXT,
    last_synced_at        TIMESTAMPTZ,
    last_pushed_at        TIMESTAMPTZ,
    created_by            INT REFERENCES users(id),
    created_at            TIMESTAMPTZ DEFAULT NOW(),
    updated_at            TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_sheet_conn_status ON sheet_connections(status);
CREATE INDEX IF NOT EXISTS idx_sheet_conn_branch ON sheet_connections(branch_id);

CREATE TABLE IF NOT EXISTS sheet_column_mappings (
    id            SERIAL PRIMARY KEY,
    connection_id INT NOT NULL REFERENCES sheet_connections(id) ON DELETE CASCADE,
    column_index  INT NOT NULL,
    header        VARCHAR(300),
    target_field  VARCHAR(200) NOT NULL,
    transform     VARCHAR(20) DEFAULT 'text',
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_sheet_map_conn ON sheet_column_mappings(connection_id);

CREATE TABLE IF NOT EXISTS sheet_row_links (
    id                SERIAL PRIMARY KEY,
    connection_id     INT NOT NULL REFERENCES sheet_connections(id) ON DELETE CASCADE,
    sheet_row_uid     VARCHAR(500) NOT NULL,
    ticket_id         INT NOT NULL REFERENCES tickets(id),
    content_hash      VARCHAR(100),
    baseline_snapshot JSONB,
    stamped_uid       VARCHAR(100),
    last_pushed_at    TIMESTAMPTZ,
    last_pulled_at    TIMESTAMPTZ,
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_sheet_link_conn ON sheet_row_links(connection_id);
CREATE INDEX IF NOT EXISTS idx_sheet_link_uid ON sheet_row_links(connection_id, sheet_row_uid);
CREATE INDEX IF NOT EXISTS idx_sheet_link_ticket ON sheet_row_links(ticket_id);

CREATE TABLE IF NOT EXISTS sheet_conflicts (
    id             SERIAL PRIMARY KEY,
    connection_id  INT NOT NULL REFERENCES sheet_connections(id) ON DELETE CASCADE,
    link_id        INT REFERENCES sheet_row_links(id) ON DELETE CASCADE,
    ticket_id      INT REFERENCES tickets(id),
    field_name     VARCHAR(200) NOT NULL,
    baseline_value TEXT,
    sheet_value    TEXT,
    db_value       TEXT,
    status         VARCHAR(20) DEFAULT 'open',
    resolution     VARCHAR(20),
    resolved_by    INT REFERENCES users(id),
    resolved_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    updated_at     TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_sheet_conflict_conn ON sheet_conflicts(connection_id, status);

CREATE TABLE IF NOT EXISTS sheet_adoptions (
    id               SERIAL PRIMARY KEY,
    connection_id    INT NOT NULL REFERENCES sheet_connections(id) ON DELETE CASCADE,
    sheet_row_uid    VARCHAR(500) NOT NULL,
    natural_key      VARCHAR(500),
    row_data_json    JSONB,
    candidate_ids    VARCHAR(500),
    status           VARCHAR(20) DEFAULT 'open',
    resolution       VARCHAR(20),
    result_ticket_id INT REFERENCES tickets(id),
    resolved_by      INT REFERENCES users(id),
    resolved_at      TIMESTAMPTZ,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_sheet_adopt_conn ON sheet_adoptions(connection_id, status);
