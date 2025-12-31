-- Migration: 000013_refresh_tokens
-- Description: Create refresh token storage for JWT authentication
-- Author: SiteProof
-- Date: 2024-12-31

CREATE TABLE refresh_tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Token (stored as hash for security)
    token_hash      TEXT NOT NULL,
    
    -- Expiration
    expires_at      TIMESTAMPTZ NOT NULL,
    
    -- Device info for session management
    device_name     TEXT,
    
    -- Usage tracking
    last_used_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    -- Revocation
    revoked_at      TIMESTAMPTZ,
    
    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    -- Constraints
    CONSTRAINT refresh_tokens_hash_format CHECK (length(token_hash) = 64),  -- SHA-256 hex
    CONSTRAINT refresh_tokens_expires_future CHECK (expires_at > created_at)
);

-- Indexes
CREATE INDEX idx_refresh_user ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_token ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_active ON refresh_tokens(user_id, expires_at) 
    WHERE revoked_at IS NULL;
CREATE INDEX idx_refresh_expired ON refresh_tokens(expires_at) 
    WHERE revoked_at IS NULL;  -- For cleanup job

-- Comments
COMMENT ON TABLE refresh_tokens IS 'Refresh tokens for JWT authentication with rotation';
COMMENT ON COLUMN refresh_tokens.token_hash IS 'SHA-256 hash of refresh token (never store raw token)';
COMMENT ON COLUMN refresh_tokens.device_name IS 'User-friendly device name for session management UI';
COMMENT ON COLUMN refresh_tokens.last_used_at IS 'Updated each time token is used to refresh access token';
COMMENT ON COLUMN refresh_tokens.revoked_at IS 'Set when user logs out or token is invalidated';

