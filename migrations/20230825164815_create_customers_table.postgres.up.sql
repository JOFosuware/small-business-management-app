CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    customer_id VARCHAR,
    cust_image BYTEA,
    id_type VARCHAR,
    card_image BYTEA,
    first_name VARCHAR,
    last_name VARCHAR,
    house_address VARCHAR,
    phone INTEGER,
    location VARCHAR,
    landmark VARCHAR,
    agreement VARCHAR,
    contract_status VARCHAR,
    months INTEGER,
    user_id INTEGER,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
)