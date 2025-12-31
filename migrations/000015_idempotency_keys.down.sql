-- Migration: 000015_idempotency_keys (rollback)

DROP TABLE IF EXISTS idempotency_keys CASCADE;

