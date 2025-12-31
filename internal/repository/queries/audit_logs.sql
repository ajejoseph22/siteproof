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

