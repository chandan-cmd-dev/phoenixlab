CREATE TABLE IF NOT EXISTS branches (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    code TEXT UNIQUE NOT NULL,
    address TEXT,
    phone TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO branches (name, code, address, created_at) VALUES 
    ('Head Office', 'HQ01', 'Newark, DE', NOW())
ON CONFLICT (code) DO NOTHING;
