-- name: GetIdempotencyKey :one
SELECT * FROM idempotency_keys
WHERE key_hash = $1 AND expires_at > now();

-- name: CreateIdempotencyKey :one
INSERT INTO idempotency_keys (
    key_hash, user_id, endpoint, method, response_status, response_body, expires_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: DeleteExpiredIdempotencyKeys :exec
DELETE FROM idempotency_keys WHERE expires_at < now();

