CREATE TABLE IF NOT EXISTS audit_log (
    id              BIGSERIAL PRIMARY KEY,
    entity_type     TEXT NOT NULL,
    entity_id       INT NOT NULL,
    action          TEXT NOT NULL,
    field_name      TEXT,
    old_value       TEXT,
    new_value       TEXT,
    changed_by      INT REFERENCES users(id),
    changed_by_name TEXT,
    ip_address      TEXT,
    user_agent      TEXT,
    changed_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_entity   ON audit_log(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_user     ON audit_log(changed_by);
CREATE INDEX IF NOT EXISTS idx_audit_time     ON audit_log(changed_at DESC);
