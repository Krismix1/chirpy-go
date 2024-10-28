-- +goose Up
BEGIN;
ALTER TABLE users ADD COLUMN hashed_PASSWORD TEXT NOT NULL DEFAULT 'unset';
ALTER TABLE users ALTER COLUMN hashed_password DROP DEFAULT;
COMMIT;

-- +goose Down
ALTER TABLE users DROP COLUMN hashed_password;
