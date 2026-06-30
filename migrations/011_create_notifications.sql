CREATE TABLE IF NOT EXISTS notifications (
    id           SERIAL PRIMARY KEY,
    recipient_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    actor_id     INT NOT NULL REFERENCES users(id),
    actor_name   VARCHAR(100),
    ticket_id    INT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    action       VARCHAR(50) NOT NULL,
    message      TEXT NOT NULL,
    is_read      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    read_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_notifications_recipient ON notifications(recipient_id, is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_created ON notifications(created_at DESC);
