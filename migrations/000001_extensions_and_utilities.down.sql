-- Migration: 000001_extensions_and_utilities (rollback)
-- WARNING: Dropping extensions may affect other databases if shared

DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;

-- Note: We don't drop pgcrypto as it may be used by other applications
-- DROP EXTENSION IF EXISTS pgcrypto;

