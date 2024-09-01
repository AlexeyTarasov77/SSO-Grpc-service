BEGIN;
DROP TYPE IF EXISTS role;
CREATE TYPE role AS ENUM ('user', 'moderator', 'admin');

CREATE TABLE IF NOT EXISTS users (
    id serial PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    password CHAR(60) NOT NULL,
    role role DEFAULT 'user',
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE OR REPLACE FUNCTION update_modified_column()   
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;   
END;
$$ language 'plpgsql';

CREATE TRIGGER update_customer_modtime BEFORE UPDATE ON users
FOR EACH ROW EXECUTE PROCEDURE update_modified_column();

CREATE INDEX IF NOT EXISTS users_email_idx ON users (email);

CREATE TABLE IF NOT EXISTS apps (
    id serial PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    secret TEXT NOT NULL UNIQUE
);
COMMIT;