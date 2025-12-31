-- name: GetPhotoByID :one
SELECT * FROM photos WHERE id = $1;

-- name: GetPhotosByTicket :many
SELECT * FROM photos
WHERE ticket_id = $1
ORDER BY created_at;

-- name: GetPhotoByHash :one
SELECT * FROM photos WHERE file_hash = $1;

-- name: CreatePhoto :one
INSERT INTO photos (
    ticket_id, uploaded_by_id, file_path, file_size, file_hash,
    mime_type, caption, taken_at, latitude, longitude
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: UpdatePhotoCaption :one
UPDATE photos
SET caption = $2
WHERE id = $1
RETURNING *;

-- name: DeletePhoto :exec
DELETE FROM photos WHERE id = $1;

-- name: CountPhotosByTicket :one
SELECT COUNT(*) FROM photos WHERE ticket_id = $1;

