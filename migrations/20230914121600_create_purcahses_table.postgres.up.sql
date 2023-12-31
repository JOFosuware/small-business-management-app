CREATE TABLE IF NOT EXISTS purchases (
    id SERIAL PRIMARY KEY,
    serial VARCHAR,
    quantity INTEGER,
    amount real,
    user_id INTEGER,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
)