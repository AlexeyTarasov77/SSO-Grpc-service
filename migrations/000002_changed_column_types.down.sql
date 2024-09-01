BEGIN;
ALTER TABLE users
ALTER COLUMN email TYPE varchar(100),
ALTER COLUMN password TYPE char(60) USING password::text,
ALTER COLUMN username TYPE varchar(255),
COMMIT;