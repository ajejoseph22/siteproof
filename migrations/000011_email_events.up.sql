-- Migration: 000011_email_events
-- Description: Create email event tracking table for Postmark webhooks
-- Author: SiteProof
-- Date: 2024-12-31

CREATE TABLE email_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id       UUID REFERENCES tickets(id) ON DELETE SET NULL,
    
    -- Postmark tracking
    message_id      TEXT UNIQUE NOT NULL,
    
    -- Recipients
    to_email        TEXT NOT NULL,
    to_name         TEXT,
    
    -- Content reference
    subject         TEXT NOT NULL,
    template        TEXT,
    
    -- Status tracking
    status          TEXT NOT NULL DEFAULT 'sent',
    
    -- Timestamps
    sent_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    delivered_at    TIMESTAMPTZ,
    opened_at       TIMESTAMPTZ,
    bounced_at      TIMESTAMPTZ,
    
    -- Error details
    error_code      TEXT,
    error_message   TEXT,
    
    -- Constraints
    CONSTRAINT email_events_status_valid CHECK (status IN (
        'sent',
        'delivered',
        'opened',
        'bounced',
        'spam_complaint'
    )),
    CONSTRAINT email_events_email_format CHECK (
        to_email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'
    )
);

-- Indexes
CREATE INDEX idx_email_ticket ON email_events(ticket_id) WHERE ticket_id IS NOT NULL;
CREATE INDEX idx_email_message ON email_events(message_id);
CREATE INDEX idx_email_status ON email_events(status) WHERE status IN ('bounced', 'spam_complaint');
CREATE INDEX idx_email_sent_at ON email_events(sent_at DESC);

-- Comments
COMMENT ON TABLE email_events IS 'Email delivery tracking for Postmark transactional emails';
COMMENT ON COLUMN email_events.message_id IS 'Postmark MessageID for webhook correlation';
COMMENT ON COLUMN email_events.template IS 'Email template used: ticket_approved, ticket_rejected, signing_request, etc.';
COMMENT ON COLUMN email_events.status IS 'Delivery status: sent, delivered, opened, bounced, spam_complaint';

