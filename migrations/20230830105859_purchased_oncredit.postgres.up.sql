CREATE TABLE IF NOT EXISTS purchased_oncredit (
    id SERIAL PRIMARY KEY,
    customer_id VARCHAR,
    serial VARCHAR,
    price real,
    quantity INTEGER,
    deposit real,
    balance real,
    user_id INTEGER,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
)