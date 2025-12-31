-- Migration: 000014_ticket_lock_enforcement (rollback)

DROP TRIGGER IF EXISTS ticket_status_transition ON tickets;
DROP TRIGGER IF EXISTS ticket_lock_enforcement ON tickets;
DROP FUNCTION IF EXISTS validate_ticket_status_transition();
DROP FUNCTION IF EXISTS prevent_locked_ticket_changes();

