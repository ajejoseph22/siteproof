-- Migration: 000005_tickets (rollback)

DROP TRIGGER IF EXISTS tickets_updated_at ON tickets;
DROP TABLE IF EXISTS tickets CASCADE;

