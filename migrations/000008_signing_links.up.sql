-- Migration: 000008_signing_links
-- Description: Create signing links for secure GC signature requests
-- Author: SiteProof
-- Date: 2024-12-31

CREATE TABLE signing_links (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id               UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    
    -- Token for URL: /sign/{token}
    token                   TEXT UNIQUE NOT NULL,
    
    -- Who should sign
    signer_type             TEXT NOT NULL,
    signer_email            TEXT NOT NULL,
    signer_phone            TEXT,
    
    -- Delivery tracking
    sent_via                TEXT[] NOT NULL DEFAULT '{}',
    sent_at                 TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    -- Expiration and usage
    expires_at              TIMESTAMPTZ NOT NULL,
    used_at                 TIMESTAMPTZ,
    revoked_at              TIMESTAMPTZ,
    
    -- SMS verification for same-device fallback
    verification_code       TEXT,
    verification_attempts   INTEGER NOT NULL DEFAULT 0,
    
    -- Timestamps
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    -- Constraints
    CONSTRAINT signing_links_token_length CHECK (length(token) >= 32),
    CONSTRAINT signing_links_signer_type_valid CHECK (signer_type IN ('gc_super')),
    CONSTRAINT signing_links_email_format CHECK (
        signer_email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'
    ),
    CONSTRAINT signing_links_phone_format CHECK (
        signer_phone IS NULL OR signer_phone ~ '^\+?[1-9]\d{1,14}$'
    ),
    CONSTRAINT signing_links_expires_future CHECK (expires_at > created_at),
    CONSTRAINT signing_links_verification_attempts_max CHECK (verification_attempts <= 3),
    CONSTRAINT signing_links_verification_code_format CHECK (
        verification_code IS NULL OR verification_code ~ '^\d{6}$'
    )
);

-- Indexes
CREATE INDEX idx_signing_links_token ON signing_links(token);
CREATE INDEX idx_signing_links_ticket ON signing_links(ticket_id);
CREATE INDEX idx_signing_links_active ON signing_links(ticket_id, expires_at) 
    WHERE used_at IS NULL AND revoked_at IS NULL;

-- Comments
COMMENT ON TABLE signing_links IS 'Secure one-time signing links sent to GC for ticket approval';
COMMENT ON COLUMN signing_links.token IS 'Cryptographically random token (32+ chars) for URL';
COMMENT ON COLUMN signing_links.sent_via IS 'Array of delivery methods used: email, sms';
COMMENT ON COLUMN signing_links.verification_code IS '6-digit code for same-device SMS verification';
COMMENT ON COLUMN signing_links.verification_attempts IS 'Failed verification attempts (max 3)';

