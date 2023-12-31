CREATE TABLE IF NOT EXISTS payments (
    id SERIAL PRIMARY KEY,
    customer_id VARCHAR,
    month VARCHAR,
    amount real,
    payment_date TIMESTAMP,
    user_id INTEGER,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
)