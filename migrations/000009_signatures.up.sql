-- Migration: 000009_signatures
-- Description: Create signatures table for foreman and GC signatures
-- Author: SiteProof
-- Date: 2024-12-31

CREATE TABLE signatures (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id           UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    signing_link_id     UUID REFERENCES signing_links(id) ON DELETE SET NULL,
    signer_id           UUID REFERENCES users(id) ON DELETE RESTRICT,
    
    -- Signer details
    signer_type         TEXT NOT NULL,
    signer_name         TEXT NOT NULL,
    signer_email        TEXT,
    
    -- Signature file
    signature_image     TEXT NOT NULL,
    signature_hash      TEXT NOT NULL,
    
    -- Method for audit/dispute resolution
    signing_method      TEXT NOT NULL,
    
    -- Device fingerprinting
    ip_address          INET,
    user_agent          TEXT,
    device_id           TEXT,
    
    -- Timestamps
    signed_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    -- Constraints
    CONSTRAINT signatures_one_per_type UNIQUE (ticket_id, signer_type),
    CONSTRAINT signatures_signer_type_valid CHECK (signer_type IN ('foreman', 'gc_super')),
    CONSTRAINT signatures_name_not_empty CHECK (length(trim(signer_name)) > 0),
    CONSTRAINT signatures_image_not_empty CHECK (length(trim(signature_image)) > 0),
    CONSTRAINT signatures_hash_format CHECK (length(signature_hash) = 64),  -- SHA-256 hex
    CONSTRAINT signatures_method_valid CHECK (signing_method IN (
        'link_own_device',      -- GC opened link on their device (strongest)
        'link_same_device',     -- GC used link on foreman's device + SMS code
        'same_device_verified'  -- Same device with SMS verification (fallback)
    )),
    CONSTRAINT signatures_email_format CHECK (
        signer_email IS NULL OR 
        signer_email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'
    )
);

-- Indexes
CREATE INDEX idx_signatures_ticket ON signatures(ticket_id);
CREATE INDEX idx_signatures_signer ON signatures(signer_id) WHERE signer_id IS NOT NULL;
CREATE INDEX idx_signatures_link ON signatures(signing_link_id) WHERE signing_link_id IS NOT NULL;

-- Comments
COMMENT ON TABLE signatures IS 'Digital signatures on tickets from foreman and GC superintendent';
COMMENT ON COLUMN signatures.signing_link_id IS 'Reference to signing link used (for GC signatures)';
COMMENT ON COLUMN signatures.signer_id IS 'User ID if signer has an account; NULL for external signers';
COMMENT ON COLUMN signatures.signature_image IS 'Path to signature PNG file';
COMMENT ON COLUMN signatures.signature_hash IS 'SHA-256 hash of signature image for integrity';
COMMENT ON COLUMN signatures.signing_method IS 'How signature was captured: link_own_device (strongest), link_same_device, same_device_verified';
COMMENT ON COLUMN signatures.device_id IS 'Client-generated device fingerprint for fraud detection';

