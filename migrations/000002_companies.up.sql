-- Migration: 000002_companies
-- Description: Create companies table (multi-tenant root entity)
-- Author: SiteProof
-- Date: 2024-12-31

CREATE TABLE companies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    slug            TEXT UNIQUE NOT NULL,
    settings        JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    -- Constraints
    CONSTRAINT companies_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT companies_slug_format CHECK (slug ~ '^[a-z0-9][a-z0-9-]*[a-z0-9]$')
);

-- Indexes
CREATE INDEX idx_companies_slug ON companies(slug);

-- Auto-update updated_at
CREATE TRIGGER companies_updated_at
    BEFORE UPDATE ON companies
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE companies IS 'Root entity for multi-tenancy. All data is scoped to a company.';
COMMENT ON COLUMN companies.slug IS 'URL-friendly identifier, lowercase alphanumeric with hyphens';
COMMENT ON COLUMN companies.settings IS 'Company-specific configuration (notification preferences, branding, etc.)';

