-- Migration: 000004_projects
-- Description: Create projects table for construction job sites
-- Author: SiteProof
-- Date: 2024-12-31

CREATE TABLE projects (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id          UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    
    -- Project details
    name                TEXT NOT NULL,
    address             TEXT,
    
    -- General contractor info
    gc_company_name     TEXT,
    gc_contact_email    TEXT,
    
    -- Status
    status              TEXT NOT NULL DEFAULT 'active',
    
    -- Display settings
    timezone            TEXT NOT NULL DEFAULT 'America/New_York',
    settings            JSONB NOT NULL DEFAULT '{}',
    
    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    -- Constraints
    CONSTRAINT projects_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT projects_status_valid CHECK (status IN ('active', 'completed', 'archived')),
    CONSTRAINT projects_gc_email_format CHECK (
        gc_contact_email IS NULL OR 
        gc_contact_email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'
    )
);

-- Indexes
CREATE INDEX idx_projects_company ON projects(company_id);
CREATE INDEX idx_projects_status ON projects(company_id, status) WHERE status = 'active';

-- Auto-update updated_at
CREATE TRIGGER projects_updated_at
    BEFORE UPDATE ON projects
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE projects IS 'Construction projects/job sites. Created by Admin/PM before field work.';
COMMENT ON COLUMN projects.gc_contact_email IS 'GC superintendent email for ticket approval notifications';
COMMENT ON COLUMN projects.timezone IS 'IANA timezone for PDF timestamp display (e.g., America/New_York)';
COMMENT ON COLUMN projects.settings IS 'Project-specific config (notification rules, custom fields, etc.)';

