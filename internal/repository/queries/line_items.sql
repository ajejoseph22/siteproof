-- name: GetLineItemsByTicket :many
SELECT * FROM ticket_line_items
WHERE ticket_id = $1
ORDER BY sort_order, created_at;

-- name: CreateLineItem :one
INSERT INTO ticket_line_items (
    ticket_id, item_type, description, quantity, unit, unit_rate, total, sort_order
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateLineItem :one
UPDATE ticket_line_items
SET description = COALESCE(sqlc.narg('description'), description),
    quantity = COALESCE(sqlc.narg('quantity'), quantity),
    unit = COALESCE(sqlc.narg('unit'), unit),
    unit_rate = COALESCE(sqlc.narg('unit_rate'), unit_rate),
    total = COALESCE(sqlc.narg('total'), total),
    sort_order = COALESCE(sqlc.narg('sort_order'), sort_order)
WHERE id = $1
RETURNING *;

-- name: DeleteLineItem :exec
DELETE FROM ticket_line_items WHERE id = $1;

-- name: DeleteLineItemsByTicket :exec
DELETE FROM ticket_line_items WHERE ticket_id = $1;

