-- Migration: 000005_tickets
-- Description: Create tickets table for T&M and variation records
-- Author: SiteProof
-- Date: 2024-12-31

CREATE TABLE tickets (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id          UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE RESTRICT,
    created_by_id       UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    
    -- Identification
    ticket_number       TEXT NOT NULL,
    ticket_type         TEXT NOT NULL,
    
    -- Content
    title               TEXT NOT NULL,
    description_raw     TEXT,
    description_cleaned TEXT,
    description_final   TEXT,
    
    -- Work details
    work_date           DATE NOT NULL,
    
    -- Status workflow
    status              TEXT NOT NULL DEFAULT 'draft',
    
    -- GC signature tracking
    gc_signed_at        TIMESTAMPTZ,
    gc_signed_by_id     UUID REFERENCES users(id) ON DELETE RESTRICT,
    
    -- Rejection tracking
    rejected_at         TIMESTAMPTZ,
    rejected_by_id      UUID REFERENCES users(id) ON DELETE RESTRICT,
    rejection_reason    TEXT,
    
    -- PDF record (set when locked)
    pdf_generated_at    TIMESTAMPTZ,
    pdf_path            TEXT,
    pdf_hash            TEXT,
    
    -- Immutability
    locked_at           TIMESTAMPTZ,
    content_hash        TEXT,
    
    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    -- Constraints
    CONSTRAINT tickets_number_unique_per_company UNIQUE (company_id, ticket_number),
    CONSTRAINT tickets_title_not_empty CHECK (length(trim(title)) > 0),
    CONSTRAINT tickets_type_valid CHECK (ticket_type IN ('time_materials', 'variation')),
    CONSTRAINT tickets_status_valid CHECK (status IN (
        'draft',
        'pending_review',
        'approved',
        'rejected',
        'locked'
    )),
    CONSTRAINT tickets_rejection_reason_required CHECK (
        (status != 'rejected') OR (rejection_reason IS NOT NULL AND length(trim(rejection_reason)) > 0)
    ),
    CONSTRAINT tickets_pdf_hash_when_locked CHECK (
        (status != 'locked') OR (pdf_hash IS NOT NULL AND content_hash IS NOT NULL)
    )
);

-- Indexes for common queries
CREATE INDEX idx_tickets_company ON tickets(company_id);
CREATE INDEX idx_tickets_project ON tickets(project_id);
CREATE INDEX idx_tickets_created_by ON tickets(created_by_id);
CREATE INDEX idx_tickets_status ON tickets(company_id, status);
CREATE INDEX idx_tickets_work_date ON tickets(work_date DESC);
CREATE INDEX idx_tickets_pending ON tickets(company_id, status, created_at DESC) 
    WHERE status IN ('draft', 'pending_review');

-- Auto-update updated_at
CREATE TRIGGER tickets_updated_at
    BEFORE UPDATE ON tickets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE tickets IS 'T&M and variation tickets capturing extra work outside contract scope';
COMMENT ON COLUMN tickets.ticket_number IS 'Human-readable identifier (e.g., TM-2024-0042). Unique per company.';
COMMENT ON COLUMN tickets.description_raw IS 'Original voice transcription, preserved for audit';
COMMENT ON COLUMN tickets.description_cleaned IS 'LLM-cleaned version of transcription';
COMMENT ON COLUMN tickets.description_final IS 'User-edited final version shown on PDF';
COMMENT ON COLUMN tickets.content_hash IS 'SHA-256 hash of all ticket content at lock time';
COMMENT ON COLUMN tickets.pdf_hash IS 'SHA-256 hash of generated PDF file';

