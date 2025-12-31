-- Migration: 000015_idempotency_keys
-- Description: Create idempotency key storage for safe request retry
-- Author: SiteProof
-- Date: 2024-12-31

-- =============================================================================
-- IDEMPOTENCY KEYS
-- =============================================================================
-- Stores idempotency keys for POST requests to handle network retries safely.
-- Client sends Idempotency-Key header; server returns cached response on replay.

CREATE TABLE idempotency_keys (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Key identification
    key_hash        TEXT NOT NULL,          -- SHA-256(key + endpoint)
    user_id         UUID REFERENCES users(id) ON DELETE CASCADE,
    
    -- Request context
    endpoint        TEXT NOT NULL,
    method          TEXT NOT NULL DEFAULT 'POST',
    
    -- Cached response
    response_status INTEGER NOT NULL,
    response_body   JSONB NOT NULL,
    
    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at      TIMESTAMPTZ NOT NULL,
    
    -- Constraints
    CONSTRAINT idempotency_keys_hash_unique UNIQUE (key_hash),
    CONSTRAINT idempotency_keys_hash_format CHECK (length(key_hash) = 64),
    CONSTRAINT idempotency_keys_expires_future CHECK (expires_at > created_at)
);

-- Indexes
CREATE INDEX idx_idempotency_hash ON idempotency_keys(key_hash);
CREATE INDEX idx_idempotency_user ON idempotency_keys(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_idempotency_expired ON idempotency_keys(expires_at);  -- For cleanup job

-- Comments
COMMENT ON TABLE idempotency_keys IS 'Cached responses for idempotent request handling';
COMMENT ON COLUMN idempotency_keys.key_hash IS 'SHA-256 hash of (idempotency_key + endpoint)';
COMMENT ON COLUMN idempotency_keys.response_status IS 'HTTP status code of original response';
COMMENT ON COLUMN idempotency_keys.response_body IS 'JSON body of original response';
COMMENT ON COLUMN idempotency_keys.expires_at IS 'Keys expire after 24 hours';

