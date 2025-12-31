-- Migration: 000014_ticket_lock_enforcement
-- Description: Trigger to prevent modification of locked tickets
-- Author: SiteProof
-- Date: 2024-12-31

-- =============================================================================
-- LOCKED TICKET PROTECTION
-- =============================================================================
-- Once a ticket is locked (status = 'locked'), it cannot be modified.
-- This is critical for immutability and legal compliance.

CREATE OR REPLACE FUNCTION prevent_locked_ticket_changes()
RETURNS TRIGGER AS $$
BEGIN
    -- Only enforce on already-locked tickets
    IF OLD.status = 'locked' THEN
        -- Check if this is a substantive change (not just metadata)
        -- We allow very limited updates even on locked tickets (e.g., tracking email delivery)
        IF NEW.status != OLD.status OR
           NEW.content_hash IS DISTINCT FROM OLD.content_hash OR
           NEW.pdf_hash IS DISTINCT FROM OLD.pdf_hash OR
           NEW.title != OLD.title OR
           NEW.description_final IS DISTINCT FROM OLD.description_final OR
           NEW.description_raw IS DISTINCT FROM OLD.description_raw OR
           NEW.description_cleaned IS DISTINCT FROM OLD.description_cleaned OR
           NEW.work_date != OLD.work_date OR
           NEW.locked_at IS DISTINCT FROM OLD.locked_at THEN
            RAISE EXCEPTION 'Locked tickets cannot be modified. Ticket ID: %', OLD.id
                USING HINT = 'This ticket has been locked for legal compliance. Create a new ticket or correction if needed.';
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER ticket_lock_enforcement
    BEFORE UPDATE ON tickets
    FOR EACH ROW
    EXECUTE FUNCTION prevent_locked_ticket_changes();

-- =============================================================================
-- STATUS TRANSITION VALIDATION
-- =============================================================================
-- Enforce valid status transitions to prevent invalid state changes

CREATE OR REPLACE FUNCTION validate_ticket_status_transition()
RETURNS TRIGGER AS $$
DECLARE
    valid_transition BOOLEAN;
BEGIN
    -- Skip if status unchanged
    IF OLD.status = NEW.status THEN
        RETURN NEW;
    END IF;
    
    -- Define valid transitions
    valid_transition := CASE
        -- From draft
        WHEN OLD.status = 'draft' AND NEW.status = 'pending_review' THEN true
        
        -- From pending_review
        WHEN OLD.status = 'pending_review' AND NEW.status = 'approved' THEN true
        WHEN OLD.status = 'pending_review' AND NEW.status = 'rejected' THEN true
        WHEN OLD.status = 'pending_review' AND NEW.status = 'draft' THEN true  -- Cancel request
        
        -- From approved
        WHEN OLD.status = 'approved' AND NEW.status = 'locked' THEN true
        
        -- From rejected
        WHEN OLD.status = 'rejected' AND NEW.status = 'draft' THEN true  -- Revise and resubmit
        
        -- All other transitions are invalid
        ELSE false
    END;
    
    IF NOT valid_transition THEN
        RAISE EXCEPTION 'Invalid ticket status transition from % to %', OLD.status, NEW.status
            USING HINT = 'Check the allowed status workflow: draft → pending_review → approved/rejected → locked';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER ticket_status_transition
    BEFORE UPDATE ON tickets
    FOR EACH ROW
    EXECUTE FUNCTION validate_ticket_status_transition();

-- Comments
COMMENT ON FUNCTION prevent_locked_ticket_changes() IS 
    'Prevents modification of locked tickets for immutability compliance';
COMMENT ON FUNCTION validate_ticket_status_transition() IS 
    'Enforces valid ticket status transitions';

