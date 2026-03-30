CREATE TABLE IF NOT EXISTS branches (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    code TEXT UNIQUE NOT NULL,
    address TEXT,
    phone TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO branches (name, code) VALUES 
    ('Head Office', 'HQ01'),
    ('Kuala Lumpur', 'KL01'),
    ('Penang', 'PG02'),
    ('Johor Bahru', 'JB03')
ON CONFLICT (code) DO NOTHING;
