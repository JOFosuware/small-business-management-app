CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    serial VARCHAR,
    name VARCHAR,
    description text,
    price real,
    units INTEGER,
    user_id INTEGER,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
)