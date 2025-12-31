-- Migration: 000002_companies (rollback)

DROP TRIGGER IF EXISTS companies_updated_at ON companies;
DROP TABLE IF EXISTS companies CASCADE;

