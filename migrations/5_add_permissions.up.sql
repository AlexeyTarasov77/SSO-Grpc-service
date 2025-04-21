CREATE TABLE IF NOT EXISTS permissions ( 
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    code text UNIQUE NOT NULL
);
CREATE TABLE IF NOT EXISTS users_permissions (
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    permission_id bigint NOT NULL REFERENCES permissions ON DELETE CASCADE, 
    granted_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, permission_id)
);
