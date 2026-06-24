CREATE TABLE IF NOT EXISTS customizations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id INTEGER,
    agent_name TEXT NOT NULL,
    customer_code TEXT NOT NULL,
    customer_business_name TEXT NOT NULL,
    customer_store_name TEXT NOT NULL,
    customer_address TEXT NOT NULL,
    customer_cap TEXT NOT NULL,
    customer_city TEXT NOT NULL,
    customer_email TEXT NOT NULL,
    customer_phone TEXT NOT NULL,
    annual_sell_in_estimate TEXT NOT NULL,
    
    -- Customization 1
    cust1_subject TEXT NOT NULL,
    cust1_type TEXT NOT NULL,
    cust1_width_cm REAL NOT NULL,
    cust1_height_cm REAL NOT NULL,
    cust1_material TEXT NOT NULL,
    
    -- Customization 2
    cust2_subject TEXT,
    cust2_type TEXT,
    cust2_width_cm REAL,
    cust2_height_cm REAL,
    cust2_material TEXT,
    
    -- Customization 3
    cust3_subject TEXT,
    cust3_type TEXT,
    cust3_width_cm REAL,
    cust3_height_cm REAL,
    cust3_material TEXT,
    
    start_date TEXT,
    end_date TEXT,
    
    printing_cost_responsibility TEXT NOT NULL,
    assembly_cost_responsibility TEXT NOT NULL,
    
    shipping_address TEXT NOT NULL,
    shipping_civic TEXT NOT NULL,
    shipping_city TEXT NOT NULL,
    shipping_province TEXT NOT NULL,
    shipping_cap TEXT NOT NULL,
    
    photo_url TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY(agent_id) REFERENCES agents(id)
);
