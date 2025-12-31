-- Migration: 000010_audit_logs (rollback)
-- WARNING: This will delete all audit history!

DROP TRIGGER IF EXISTS audit_immutable ON audit_logs;
DROP TRIGGER IF EXISTS audit_hash_chain ON audit_logs;
DROP FUNCTION IF EXISTS prevent_audit_modification();
DROP FUNCTION IF EXISTS compute_audit_hash();
DROP TABLE IF EXISTS audit_logs CASCADE;

