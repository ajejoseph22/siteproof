-- Migration: 000003_users (rollback)

DROP TRIGGER IF EXISTS users_updated_at ON users;
DROP TABLE IF EXISTS users CASCADE;

