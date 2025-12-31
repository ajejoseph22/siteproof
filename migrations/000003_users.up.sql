-- Migration: 000003_users
-- Description: Create users table with role-based access control
-- Author: SiteProof
-- Date: 2024-12-31

-- Define valid user roles as a check constraint (not enum for flexibility)
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id      UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    
    -- Authentication
    email           TEXT UNIQUE NOT NULL,
    password_hash   TEXT,  -- NULL if using OAuth only
    
    -- Profile
    name            TEXT NOT NULL,
    phone           TEXT,
    
    -- Authorization
    role            TEXT NOT NULL,
    active          BOOLEAN NOT NULL DEFAULT true,
    
    -- Soft delete support
    deleted_at      TIMESTAMPTZ,
    
    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    -- Constraints
    CONSTRAINT users_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    CONSTRAINT users_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT users_role_valid CHECK (role IN (
        'foreman',      -- Creates tickets, captures photos
        'gc_super',     -- Signs tickets on behalf of GC
        'pm',           -- Project manager, reviews tickets
        'accounting',   -- Views approved tickets for billing
        'admin'         -- Company admin, manages users
    )),
    CONSTRAINT users_phone_format CHECK (
        phone IS NULL OR phone ~ '^\+?[1-9]\d{1,14}$'
    )
);

-- Indexes for common queries
CREATE INDEX idx_users_company ON users(company_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role ON users(company_id, role) WHERE active = true AND deleted_at IS NULL;

-- Auto-update updated_at
CREATE TRIGGER users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE users IS 'Application users with role-based access control';
COMMENT ON COLUMN users.password_hash IS 'bcrypt hash of password. NULL if user authenticates via OAuth only.';
COMMENT ON COLUMN users.role IS 'User role: foreman, gc_super, pm, accounting, admin';
COMMENT ON COLUMN users.deleted_at IS 'Soft delete timestamp. Non-NULL means user is deleted.';

