-- name: CreateAuditLog :one
INSERT INTO audit_logs (entity_type, entity_id, action, actor_id, actor_ip, changes)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetAuditLogsByEntity :many
SELECT * FROM audit_logs
WHERE entity_type = $1 AND entity_id = $2
ORDER BY occurred_at DESC
LIMIT $3 OFFSET $4;

-- name: GetAuditLogsByActor :many
SELECT * FROM audit_logs
WHERE actor_id = $1
ORDER BY occurred_at DESC
LIMIT $2 OFFSET $3;

-- name: GetRecentAuditLogs :many
SELECT * FROM audit_logs
ORDER BY occurred_at DESC
LIMIT $1 OFFSET $2;

-- name: VerifyAuditChain :one
-- Verify the hash chain is intact by checking each entry's hash
SELECT COUNT(*) AS broken_entries
FROM (
    SELECT id, entry_hash, prev_hash,
        LAG(entry_hash) OVER (ORDER BY id) AS expected_prev_hash
    FROM audit_logs
) t
WHERE prev_hash IS DISTINCT FROM expected_prev_hash AND id > 1;

-- name: VerifyAuditChainForEntity :one
-- Verify the hash chain is intact for a specific entity (e.g., a ticket)
-- This checks that all audit entries for this entity have valid prev_hash references
-- Note: The global chain is what provides tamper-evidence, but this helps verify
-- that the entity's entries are part of a valid chain
SELECT COUNT(*) AS broken_entries
FROM (
    SELECT sub.id, sub.entry_hash, sub.prev_hash,
        LAG(sub.entry_hash) OVER (ORDER BY sub.id) AS expected_prev_hash
    FROM audit_logs sub
    WHERE sub.id IN (
        -- Get all audit log IDs that could affect this entity's chain
        -- We need to check from the first entry for this entity to the last
        SELECT a.id FROM audit_logs a
        WHERE a.id >= (SELECT MIN(b.id) FROM audit_logs b WHERE b.entity_type = $1 AND b.entity_id = $2)
          AND a.id <= (SELECT MAX(c.id) FROM audit_logs c WHERE c.entity_type = $1 AND c.entity_id = $2)
    )
) t
WHERE t.prev_hash IS DISTINCT FROM t.expected_prev_hash
  AND t.id > (SELECT MIN(m.id) FROM audit_logs m);

-- name: GetAuditLogsByEntityWithHashes :many
-- Get audit logs for an entity including hash information for verification display
SELECT id, entity_type, entity_id, action, actor_id, actor_ip, occurred_at, changes, prev_hash, entry_hash
FROM audit_logs
WHERE entity_type = $1 AND entity_id = $2
ORDER BY occurred_at ASC, id ASC;

-- name: CountAuditLogsByEntity :one
-- Count audit logs for an entity
SELECT COUNT(*) FROM audit_logs
WHERE entity_type = $1 AND entity_id = $2;

