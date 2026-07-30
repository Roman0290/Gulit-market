DROP INDEX IF EXISTS idx_users_status;
ALTER TABLE users DROP COLUMN status;
DROP TYPE IF EXISTS user_status;
