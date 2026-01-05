-- =============================================================================
-- SiteProof Seed Data Script
-- =============================================================================
-- Usage: make db-seed ENV=dev|test
-- WARNING: This will clear existing data! Do NOT run in production.
-- =============================================================================

-- Clear existing data (in reverse dependency order)
TRUNCATE
    audit_logs,
    transcription_versions,
    photos,
    ticket_line_items,
    signatures,
    signing_links,
    tickets,
    projects,
    refresh_tokens,
    email_events,
    idempotency_keys,
    users,
    companies
CASCADE;

-- =============================================================================
-- COMPANIES
-- =============================================================================
INSERT INTO companies (id, name, slug, settings) VALUES
    ('11111111-1111-1111-1111-111111111111',
     'Acme Electrical',
     'acme-electrical',
     '{"notification_emails": ["admin@acme.com"]}'),
    ('11111111-1111-1111-1111-111111111112',
     'Metro Plumbing Co',
     'metro-plumbing',
     '{"notification_emails": ["office@metroplumbing.com"]}');

-- =============================================================================
-- USERS
-- =============================================================================
-- Password for all users: "password123"
-- bcrypt hash: $2a$12$6spVyjemkF4.ctI11qtRluejMhAcO9mn0BQ0r.TfXYnpSRNtXfGB.

-- Acme Electrical users
INSERT INTO users (id, company_id, email, password_hash, name, role, phone) VALUES
    -- Admin
    ('22222222-2222-2222-2222-222222222221',
     '11111111-1111-1111-1111-111111111111',
     'admin@acme.com',
     '$2a$12$6spVyjemkF4.ctI11qtRluejMhAcO9mn0BQ0r.TfXYnpSRNtXfGB.',
     'Alice Admin',
     'admin',
     '+15551000001'),
    -- Foreman 1
    ('22222222-2222-2222-2222-222222222222',
     '11111111-1111-1111-1111-111111111111',
     'john@acme.com',
     '$2a$12$6spVyjemkF4.ctI11qtRluejMhAcO9mn0BQ0r.TfXYnpSRNtXfGB.',
     'John Foreman',
     'foreman',
     '+15551000002'),
    -- Foreman 2
    ('22222222-2222-2222-2222-222222222223',
     '11111111-1111-1111-1111-111111111111',
     'jane@acme.com',
     '$2a$12$6spVyjemkF4.ctI11qtRluejMhAcO9mn0BQ0r.TfXYnpSRNtXfGB.',
     'Jane Foreman',
     'foreman',
     '+15551000003');

-- Metro Plumbing users
INSERT INTO users (id, company_id, email, password_hash, name, role, phone) VALUES
    ('22222222-2222-2222-2222-222222222231',
     '11111111-1111-1111-1111-111111111112',
     'admin@metroplumbing.com',
     '$2a$12$6spVyjemkF4.ctI11qtRluejMhAcO9mn0BQ0r.TfXYnpSRNtXfGB.',
     'Mike Metro',
     'admin',
     '+15552000001');

-- =============================================================================
-- PROJECTS
-- =============================================================================
-- Acme Electrical projects
INSERT INTO projects (id, company_id, name, address, gc_company_name, gc_contact_name, gc_contact_email, gc_contact_phone, timezone, status) VALUES
    ('44444444-4444-4444-4444-444444444441',
     '11111111-1111-1111-1111-111111111111',
     'Downtown Office Tower',
     '123 Main St, New York, NY 10001',
     'BuildCo General Contractors',
     'Bob Builder',
     'bob@buildco.com',
     '+15553001001',
     'America/New_York',
     'active'),
    ('44444444-4444-4444-4444-444444444442',
     '11111111-1111-1111-1111-111111111111',
     'Harbor View Apartments',
     '500 Harbor Dr, San Francisco, CA 94105',
     'Pacific Construction',
     'Sarah Super',
     'sarah@pacificconstruction.com',
     '+15553001002',
     'America/Los_Angeles',
     'active'),
    ('44444444-4444-4444-4444-444444444443',
     '11111111-1111-1111-1111-111111111111',
     'Retail Center Phase 2',
     '789 Shopping Blvd, Chicago, IL 60601',
     'Midwest Builders Inc',
     'Tom Superintendent',
     'tom@midwestbuilders.com',
     '+15553001003',
     'America/Chicago',
     'active');

-- Metro Plumbing projects
INSERT INTO projects (id, company_id, name, address, gc_company_name, gc_contact_name, gc_contact_email, gc_contact_phone, timezone, status) VALUES
    ('44444444-4444-4444-4444-444444444451',
     '11111111-1111-1111-1111-111111111112',
     'City Hospital Renovation',
     '200 Health Ave, Boston, MA 02101',
     'Healthcare Builders LLC',
     'Dr. Build',
     'contact@healthcarebuilders.com',
     '+15554001001',
     'America/New_York',
     'active');

-- =============================================================================
-- TICKETS (sample T&M records)
-- =============================================================================
INSERT INTO tickets (id, company_id, project_id, created_by, ticket_number, status, work_date, description) VALUES
    -- Draft ticket
    ('55555555-5555-5555-5555-555555555551',
     '11111111-1111-1111-1111-111111111111',
     '44444444-4444-4444-4444-444444444441',
     '22222222-2222-2222-2222-222222222222',
     'TM-001',
     'draft',
     CURRENT_DATE,
     'Emergency panel replacement in basement electrical room. GC requested additional circuits for new HVAC units.'),
    -- Pending review ticket
    ('55555555-5555-5555-5555-555555555552',
     '11111111-1111-1111-1111-111111111111',
     '44444444-4444-4444-4444-444444444441',
     '22222222-2222-2222-2222-222222222222',
     'TM-002',
     'pending_review',
     CURRENT_DATE - INTERVAL '1 day',
     'Ran conduit for relocated server room. 150ft of 2" EMT conduit with associated fittings.'),
    -- Approved ticket
    ('55555555-5555-5555-5555-555555555553',
     '11111111-1111-1111-1111-111111111111',
     '44444444-4444-4444-4444-444444444442',
     '22222222-2222-2222-2222-222222222223',
     'TM-003',
     'approved',
     CURRENT_DATE - INTERVAL '3 days',
     'Added 6 additional receptacles in unit 302 per owner change request.');

-- =============================================================================
-- LINE ITEMS
-- =============================================================================
INSERT INTO ticket_line_items (id, ticket_id, description, quantity, unit, unit_price, line_total) VALUES
    -- TM-001 line items
    ('66666666-6666-6666-6666-666666666661',
     '55555555-5555-5555-5555-555555555551',
     'Journeyman Electrician',
     8.0, 'hours', 85.00, 680.00),
    ('66666666-6666-6666-6666-666666666662',
     '55555555-5555-5555-5555-555555555551',
     '200A Panel',
     1.0, 'each', 450.00, 450.00),
    ('66666666-6666-6666-6666-666666666663',
     '55555555-5555-5555-5555-555555555551',
     '20A Circuit Breaker',
     6.0, 'each', 25.00, 150.00),
    -- TM-002 line items
    ('66666666-6666-6666-6666-666666666671',
     '55555555-5555-5555-5555-555555555552',
     'Journeyman Electrician',
     12.0, 'hours', 85.00, 1020.00),
    ('66666666-6666-6666-6666-666666666672',
     '55555555-5555-5555-5555-555555555552',
     'Apprentice Electrician',
     12.0, 'hours', 55.00, 660.00),
    ('66666666-6666-6666-6666-666666666673',
     '55555555-5555-5555-5555-555555555552',
     '2" EMT Conduit',
     150.0, 'feet', 8.50, 1275.00),
    -- TM-003 line items
    ('66666666-6666-6666-6666-666666666681',
     '55555555-5555-5555-5555-555555555553',
     'Journeyman Electrician',
     4.0, 'hours', 85.00, 340.00),
    ('66666666-6666-6666-6666-666666666682',
     '55555555-5555-5555-5555-555555555553',
     'Duplex Receptacle',
     6.0, 'each', 12.00, 72.00);

-- =============================================================================
-- OUTPUT SUMMARY
-- =============================================================================
SELECT
    'Seeding complete!' AS status,
    (SELECT COUNT(*) FROM companies) AS companies,
    (SELECT COUNT(*) FROM users) AS users,
    (SELECT COUNT(*) FROM projects) AS projects,
    (SELECT COUNT(*) FROM tickets) AS tickets,
    (SELECT COUNT(*) FROM ticket_line_items) AS line_items;