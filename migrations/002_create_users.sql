CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'technician',
    branch_id INT REFERENCES branches(id),
    is_active BOOLEAN DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    created_by INT REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_branch ON users(branch_id);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- Default super_admin (password: changeme123 - CHANGE IMMEDIATELY)
INSERT INTO users (name, email, password_hash, role, branch_id, created_at, updated_at) VALUES
  ('Super Admin', 'admin@phoenixlab.local',
   '$2a$12$q6MxBh1a7Pwwn4wt.mymmuwCNzYl04KZWKt2XMzvQ7XODacw2cPbq', 'super_admin', 1, NOW(), NOW())
ON CONFLICT (email) DO NOTHING;
