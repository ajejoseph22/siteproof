-- Migration: 000001_extensions_and_utilities
-- Description: Enable required PostgreSQL extensions and create utility functions
-- Author: SiteProof
-- Date: 2024-12-31

-- =============================================================================
-- EXTENSIONS
-- =============================================================================

-- pgcrypto: Required for gen_random_uuid() and cryptographic functions
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- =============================================================================
-- UTILITY FUNCTIONS
-- =============================================================================

-- Automatically update updated_at timestamp on row modification
-- Usage: Apply trigger to any table with an updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_updated_at_column() IS 
    'Trigger function to automatically set updated_at to current timestamp on UPDATE';

