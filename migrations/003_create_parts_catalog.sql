CREATE TABLE IF NOT EXISTS parts_catalog (
    id SERIAL PRIMARY KEY,
    part_number TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL,
    unit_cost NUMERIC(10,2) DEFAULT 0,
    supplier TEXT,
    lead_time_days INT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO parts_catalog (part_number, description, unit_cost, supplier, lead_time_days, created_at) VALUES
    ('LCD-001', '14 inch LCD Panel', 250.00, 'Dell Supplier', 7, NOW()),
    ('BATT-002', '6-cell Laptop Battery', 180.00, 'Battery Pro', 5, NOW()),
    ('KB-003', 'US Keyboard Layout', 45.00, 'Key Components', 3, NOW()),
    ('SSD-004', '256GB SSD SATA', 120.00, 'Storage Solutions', 4, NOW()),
    ('RAM-005', '8GB DDR4 RAM', 65.00, 'Memory Masters', 2, NOW())
ON CONFLICT (part_number) DO NOTHING;
