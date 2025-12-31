-- name: GetSignatureByID :one
SELECT * FROM signatures WHERE id = $1;

-- name: GetSignaturesByTicket :many
SELECT * FROM signatures
WHERE ticket_id = $1
ORDER BY signed_at;

-- name: GetSignatureByTicketAndType :one
SELECT * FROM signatures
WHERE ticket_id = $1 AND signer_type = $2;

-- name: CreateSignature :one
INSERT INTO signatures (
    ticket_id, signing_link_id, signer_id, signer_type, signer_name,
    signer_email, signature_image, signature_hash, signing_method,
    ip_address, user_agent, device_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: HasForemanSignature :one
SELECT EXISTS(
    SELECT 1 FROM signatures
    WHERE ticket_id = $1 AND signer_type = 'foreman'
) AS has_signature;

-- name: HasGCSignature :one
SELECT EXISTS(
    SELECT 1 FROM signatures
    WHERE ticket_id = $1 AND signer_type = 'gc_super'
) AS has_signature;

