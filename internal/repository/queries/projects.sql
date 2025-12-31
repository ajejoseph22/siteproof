-- name: GetProjectByID :one
SELECT * FROM projects WHERE id = $1;

-- name: GetProjectsByCompany :many
SELECT * FROM projects
WHERE company_id = $1
ORDER BY 
    CASE status WHEN 'active' THEN 0 WHEN 'completed' THEN 1 ELSE 2 END,
    name
LIMIT $2 OFFSET $3;

-- name: GetActiveProjectsByCompany :many
SELECT * FROM projects
WHERE company_id = $1 AND status = 'active'
ORDER BY name;

-- name: CreateProject :one
INSERT INTO projects (company_id, name, address, gc_company_name, gc_contact_email, status, timezone, settings)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateProject :one
UPDATE projects
SET name = COALESCE(sqlc.narg('name'), name),
    address = COALESCE(sqlc.narg('address'), address),
    gc_company_name = COALESCE(sqlc.narg('gc_company_name'), gc_company_name),
    gc_contact_email = COALESCE(sqlc.narg('gc_contact_email'), gc_contact_email),
    status = COALESCE(sqlc.narg('status'), status),
    timezone = COALESCE(sqlc.narg('timezone'), timezone),
    settings = COALESCE(sqlc.narg('settings'), settings),
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CountProjectsByCompany :one
SELECT COUNT(*) FROM projects WHERE company_id = $1;

