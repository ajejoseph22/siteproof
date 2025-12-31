-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL;

-- name: GetUsersByCompany :many
SELECT * FROM users
WHERE company_id = $1 AND deleted_at IS NULL
ORDER BY name
LIMIT $2 OFFSET $3;

-- name: CreateUser :one
INSERT INTO users (company_id, email, password_hash, name, phone, role, active)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET name = COALESCE(sqlc.narg('name'), name),
    phone = COALESCE(sqlc.narg('phone'), phone),
    role = COALESCE(sqlc.narg('role'), role),
    active = COALESCE(sqlc.narg('active'), active),
    updated_at = now()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteUser :exec
UPDATE users
SET deleted_at = now(), active = false, updated_at = now()
WHERE id = $1;

-- name: CountUsersByCompany :one
SELECT COUNT(*) FROM users
WHERE company_id = $1 AND deleted_at IS NULL;

-- name: GetUsersByRole :many
SELECT * FROM users
WHERE company_id = $1 AND role = $2 AND active = true AND deleted_at IS NULL
ORDER BY name;

