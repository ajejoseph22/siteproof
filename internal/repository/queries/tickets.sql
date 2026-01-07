-- name: GetTicketByID :one
SELECT * FROM tickets WHERE id = $1;

-- name: GetTicketByNumber :one
SELECT * FROM tickets WHERE company_id = $1 AND ticket_number = $2;

-- name: GetTicketsByProject :many
SELECT * FROM tickets
WHERE project_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetTicketsByCompany :many
SELECT * FROM tickets
WHERE company_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetTicketsByStatus :many
SELECT * FROM tickets
WHERE company_id = $1 AND status = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetTicketsByCreator :many
SELECT * FROM tickets
WHERE created_by_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetNextTicketNumber :one
SELECT COALESCE(
    MAX(CAST(SUBSTRING(ticket_number FROM '[0-9]+$') AS INTEGER)),
    0
) + 1 AS next_number
FROM tickets
WHERE company_id = $1;

-- name: CreateTicket :one
INSERT INTO tickets (
    company_id, project_id, created_by_id, ticket_number, ticket_type,
    title, description_raw, description_cleaned, description_final,
    work_date, status
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: UpdateTicket :one
UPDATE tickets
SET title = COALESCE(sqlc.narg('title'), title),
    description_final = COALESCE(sqlc.narg('description_final'), description_final),
    work_date = COALESCE(sqlc.narg('work_date'), work_date),
    updated_at = now()
WHERE id = $1 AND status = 'draft'
RETURNING *;

-- name: UpdateTicketDescription :one
UPDATE tickets
SET description_raw = COALESCE(sqlc.narg('description_raw'), description_raw),
    description_cleaned = COALESCE(sqlc.narg('description_cleaned'), description_cleaned),
    description_final = COALESCE(sqlc.narg('description_final'), description_final),
    updated_at = now()
WHERE id = $1 AND status = 'draft'
RETURNING *;

-- name: UpdateTicketStatus :one
UPDATE tickets
SET status = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ApproveTicket :one
UPDATE tickets
SET status = 'approved',
    gc_signed_at = now(),
    gc_signed_by_id = $2,
    updated_at = now()
WHERE id = $1 AND status = 'pending_review'
RETURNING *;

-- name: RejectTicket :one
UPDATE tickets
SET status = 'rejected',
    rejected_at = now(),
    rejected_by_id = $2,
    rejection_reason = $3,
    updated_at = now()
WHERE id = $1 AND status = 'pending_review'
RETURNING *;

-- name: LockTicket :one
UPDATE tickets
SET status = 'locked',
    locked_at = now(),
    pdf_generated_at = now(),
    pdf_path = $2,
    pdf_hash = $3,
    content_hash = $4,
    updated_at = now()
WHERE id = $1 AND status = 'approved'
RETURNING *;

-- name: CountTicketsByProject :one
SELECT COUNT(*) FROM tickets WHERE project_id = $1;

-- name: CountTicketsByCompany :one
SELECT COUNT(*) FROM tickets WHERE company_id = $1;

-- name: ListTicketsByCursor :many
-- Cursor-based pagination for tickets by company.
-- cursor_time and cursor_id are the created_at and id of the last item from previous page.
-- For first page, pass NULL for both cursor params (or a far-future timestamp and empty UUID).
SELECT * FROM tickets
WHERE company_id = $1
  AND (sqlc.narg('cursor_time')::timestamptz IS NULL
       OR (created_at, id) < (sqlc.narg('cursor_time')::timestamptz, sqlc.narg('cursor_id')::uuid))
ORDER BY created_at DESC, id DESC
LIMIT $2;

-- name: ListTicketsByStatusCursor :many
-- Cursor-based pagination for tickets by company and status.
SELECT * FROM tickets
WHERE company_id = $1
  AND status = $2
  AND (sqlc.narg('cursor_time')::timestamptz IS NULL
       OR (created_at, id) < (sqlc.narg('cursor_time')::timestamptz, sqlc.narg('cursor_id')::uuid))
ORDER BY created_at DESC, id DESC
LIMIT $3;

-- name: ListTicketsByProjectCursor :many
-- Cursor-based pagination for tickets by project.
SELECT * FROM tickets
WHERE project_id = $1
  AND (sqlc.narg('cursor_time')::timestamptz IS NULL
       OR (created_at, id) < (sqlc.narg('cursor_time')::timestamptz, sqlc.narg('cursor_id')::uuid))
ORDER BY created_at DESC, id DESC
LIMIT $2;

