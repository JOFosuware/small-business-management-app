CREATE TABLE IF NOT EXISTS witness (
    id SERIAL PRIMARY KEY,
    customer_id VARCHAR,
    first_name VARCHAR,
    last_name VARCHAR,
    phone INTEGER,
    terms VARCHAR,
    witness_image BYTEA,
    user_id INTEGER,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
)