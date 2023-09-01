CREATE TABLE purchased_oncredit (
    id SERIAL PRIMARY KEY,
    customer_id VARCHAR,
    serial VARCHAR,
    price INTEGER,
    quantity INTEGER,
    deposit INTEGER,
    balance INTEGER,
    item_image BYTEA,
    user_id INTEGER,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
)