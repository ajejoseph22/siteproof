-- name: GetEmailEventByID :one
SELECT * FROM email_events WHERE id = $1;

-- name: GetEmailEventByMessageID :one
SELECT * FROM email_events WHERE message_id = $1;

-- name: GetEmailEventsByTicket :many
SELECT * FROM email_events
WHERE ticket_id = $1
ORDER BY sent_at DESC;

-- name: CreateEmailEvent :one
INSERT INTO email_events (ticket_id, message_id, to_email, to_name, subject, template, status)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateEmailEventDelivered :exec
UPDATE email_events
SET status = 'delivered', delivered_at = now()
WHERE message_id = $1;

-- name: UpdateEmailEventOpened :exec
UPDATE email_events
SET status = 'opened', opened_at = now()
WHERE message_id = $1 AND status != 'bounced';

-- name: UpdateEmailEventBounced :exec
UPDATE email_events
SET status = 'bounced', bounced_at = now(), error_code = $2, error_message = $3
WHERE message_id = $1;

-- name: UpdateEmailEventSpamComplaint :exec
UPDATE email_events
SET status = 'spam_complaint'
WHERE message_id = $1;

