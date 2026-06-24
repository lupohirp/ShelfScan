CREATE TABLE IF NOT EXISTS stores (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    province TEXT NOT NULL,
    province_name TEXT NOT NULL,
    address TEXT NOT NULL,
    region TEXT NOT NULL,
    region_code TEXT NOT NULL,
    city TEXT NOT NULL,
    agent_name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS visits (
    id TEXT PRIMARY KEY,
    store_id INTEGER NOT NULL,
    status TEXT NOT NULL,
    coverage INTEGER NOT NULL,
    missing_count INTEGER NOT NULL,
    agent TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    finalized_at DATETIME,
    FOREIGN KEY(store_id) REFERENCES stores(id)
);

CREATE TABLE IF NOT EXISTS visit_scans (
    id TEXT PRIMARY KEY,
    visit_id TEXT NOT NULL,
    photo_url TEXT NOT NULL,
    photo_tone TEXT NOT NULL,
    FOREIGN KEY(visit_id) REFERENCES visits(id)
);

CREATE TABLE IF NOT EXISTS visit_products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    visit_id TEXT NOT NULL,
    sku TEXT NOT NULL,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    is_exposed BOOLEAN NOT NULL,
    FOREIGN KEY(visit_id) REFERENCES visits(id)
);
