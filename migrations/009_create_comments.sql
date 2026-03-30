CREATE TABLE IF NOT EXISTS comments (
    id          SERIAL PRIMARY KEY,
    ticket_id   INT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    user_id     INT NOT NULL REFERENCES users(id),
    body        TEXT NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_comments_ticket ON comments(ticket_id);
CREATE INDEX IF NOT EXISTS idx_comments_created ON comments(created_at DESC);
