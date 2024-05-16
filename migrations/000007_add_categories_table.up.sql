CREATE TABLE IF NOT EXISTS categories (
    id bigserial PRIMARY KEY,
    name text NOT NULL,
    description text
);