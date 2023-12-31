CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR,
    last_name VARCHAR,
    user_name VARCHAR,
    password VARCHAR,
    user_image BYTEA,
    access_level VARCHAR,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
)