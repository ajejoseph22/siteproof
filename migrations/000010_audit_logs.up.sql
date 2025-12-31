-- Migration: 000010_audit_logs
-- Description: Create immutable, hash-chained audit log table
-- Author: SiteProof
-- Date: 2024-12-31

CREATE TABLE audit_logs (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    
    -- What changed
    entity_type     TEXT NOT NULL,
    entity_id       UUID NOT NULL,
    action          TEXT NOT NULL,
    
    -- Who and when
    actor_id        UUID REFERENCES users(id) ON DELETE SET NULL,
    actor_ip        INET,
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    -- Change details
    changes         JSONB,
    
    -- Hash chain for tamper evidence
    prev_hash       TEXT,
    entry_hash      TEXT NOT NULL,
    
    -- Constraints
    CONSTRAINT audit_logs_entity_type_not_empty CHECK (length(trim(entity_type)) > 0),
    CONSTRAINT audit_logs_action_not_empty CHECK (length(trim(action)) > 0),
    CONSTRAINT audit_logs_entry_hash_format CHECK (length(entry_hash) = 64)
);

-- Indexes for common queries
CREATE INDEX idx_audit_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_actor ON audit_logs(actor_id) WHERE actor_id IS NOT NULL;
CREATE INDEX idx_audit_time ON audit_logs(occurred_at DESC);
CREATE INDEX idx_audit_action ON audit_logs(entity_type, action);

-- =============================================================================
-- HASH CHAIN TRIGGER
-- =============================================================================
-- Computes entry_hash as SHA-256 of (prev_hash + entity_type + entity_id + action + occurred_at + changes)
-- This creates a tamper-evident chain: if any entry is modified, all subsequent hashes become invalid

CREATE OR REPLACE FUNCTION compute_audit_hash()
RETURNS TRIGGER AS $$
BEGIN
    -- Get hash of previous entry
    SELECT entry_hash INTO NEW.prev_hash
    FROM audit_logs
    ORDER BY id DESC
    LIMIT 1;
    
    -- If no previous entry, prev_hash stays NULL
    -- Compute this entry's hash
    NEW.entry_hash := encode(
        sha256(
            (
                COALESCE(NEW.prev_hash, '') ||
                NEW.entity_type ||
                NEW.entity_id::text ||
                NEW.action ||
                NEW.occurred_at::text ||
                COALESCE(NEW.changes::text, '')
            )::bytea
        ),
        'hex'
    );
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER audit_hash_chain
    BEFORE INSERT ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION compute_audit_hash();

-- =============================================================================
-- IMMUTABILITY TRIGGER
-- =============================================================================
-- Prevents any UPDATE or DELETE on audit_logs

CREATE OR REPLACE FUNCTION prevent_audit_modification()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Audit logs cannot be modified or deleted';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER audit_immutable
    BEFORE UPDATE OR DELETE ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION prevent_audit_modification();

-- Comments
COMMENT ON TABLE audit_logs IS 'Immutable, hash-chained audit trail for all significant actions';
COMMENT ON COLUMN audit_logs.entity_type IS 'Type of entity: ticket, photo, signature, user, etc.';
COMMENT ON COLUMN audit_logs.action IS 'Action performed: created, updated, signed, locked, deleted, etc.';
COMMENT ON COLUMN audit_logs.changes IS 'JSON object with {field: {old: value, new: value}} for each changed field';
COMMENT ON COLUMN audit_logs.prev_hash IS 'Hash of previous audit log entry (NULL for first entry)';
COMMENT ON COLUMN audit_logs.entry_hash IS 'SHA-256 hash of this entry for chain verification';

