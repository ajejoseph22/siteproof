-- name: GetCompanyByID :one
SELECT * FROM companies WHERE id = $1;

-- name: GetCompanyBySlug :one
SELECT * FROM companies WHERE slug = $1;

-- name: CreateCompany :one
INSERT INTO companies (name, slug, settings)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateCompany :one
UPDATE companies
SET name = COALESCE(sqlc.narg('name'), name),
    settings = COALESCE(sqlc.narg('settings'), settings),
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListCompanies :many
SELECT * FROM companies
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: CountCompanies :one
SELECT COUNT(*) FROM companies;

