-- Migration: 000016_seed_data
-- Description: Optional seed data for development/testing
-- Author: SiteProof
-- Date: 2024-12-31
-- Note: This migration is intentionally empty in production. 
--       Use separate seed scripts for dev/test environments.

-- =============================================================================
-- DEVELOPMENT SEED DATA
-- =============================================================================
-- Uncomment and modify for development environments only.
-- Do NOT run in production.

/*
-- Sample company
INSERT INTO companies (id, name, slug, settings) VALUES
    ('11111111-1111-1111-1111-111111111111', 'Acme Electrical', 'acme-electrical', 
     '{"notification_emails": ["admin@acme.com"]}');

-- Sample admin user (password: "password123" - bcrypt hash)
INSERT INTO users (id, company_id, email, password_hash, name, role) VALUES
    ('22222222-2222-2222-2222-222222222222', 
     '11111111-1111-1111-1111-111111111111',
     'admin@acme.com',
     '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.Z/sMVz5fX.rvLu',  -- "password123"
     'Admin User',
     'admin');

-- Sample foreman
INSERT INTO users (id, company_id, email, password_hash, name, role, phone) VALUES
    ('33333333-3333-3333-3333-333333333333', 
     '11111111-1111-1111-1111-111111111111',
     'foreman@acme.com',
     '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.Z/sMVz5fX.rvLu',
     'John Foreman',
     'foreman',
     '+15551234567');

-- Sample project
INSERT INTO projects (id, company_id, name, address, gc_company_name, gc_contact_email, timezone) VALUES
    ('44444444-4444-4444-4444-444444444444',
     '11111111-1111-1111-1111-111111111111',
     'Downtown Office Tower',
     '123 Main St, New York, NY 10001',
     'BuildCo General Contractors',
     'super@buildco.com',
     'America/New_York');
*/

-- Placeholder to make migration valid
SELECT 1;

