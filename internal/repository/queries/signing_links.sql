-- name: GetSigningLinkByID :one
SELECT * FROM signing_links WHERE id = $1;

-- name: GetSigningLinkByToken :one
SELECT * FROM signing_links WHERE token = $1;

-- name: GetActiveSigningLinkByTicket :one
SELECT * FROM signing_links
WHERE ticket_id = $1
    AND used_at IS NULL
    AND revoked_at IS NULL
    AND expires_at > now()
ORDER BY created_at DESC
LIMIT 1;

-- name: CreateSigningLink :one
INSERT INTO signing_links (
    ticket_id, token, signer_type, signer_email, signer_phone,
    sent_via, expires_at, verification_code
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: MarkSigningLinkUsed :one
UPDATE signing_links
SET used_at = now()
WHERE id = $1
RETURNING *;

-- name: RevokeSigningLink :exec
UPDATE signing_links
SET revoked_at = now()
WHERE id = $1;

-- name: RevokeSigningLinksByTicket :exec
UPDATE signing_links
SET revoked_at = now()
WHERE ticket_id = $1 AND used_at IS NULL AND revoked_at IS NULL;

-- name: IncrementVerificationAttempts :one
UPDATE signing_links
SET verification_attempts = verification_attempts + 1
WHERE id = $1
RETURNING *;

