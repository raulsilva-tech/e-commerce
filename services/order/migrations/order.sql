CREATE TABLE IF NOT EXISTS orders(
    id SERIAL PRIMARY KEY,
    product_id integer NOT NULL,
    quantity integer NOT NULL,
    total DOUBLE PRECISION
);