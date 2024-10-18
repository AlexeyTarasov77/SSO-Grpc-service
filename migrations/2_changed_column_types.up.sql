BEGIN;
ALTER TABLE users
ALTER COLUMN email TYPE citext,
ALTER COLUMN password TYPE bytea USING password::bytea,
ALTER COLUMN username TYPE text;
COMMIT;