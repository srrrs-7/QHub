-- name: CreateOrganization :one
INSERT INTO organizations (name, slug, plan)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetOrganization :one
SELECT * FROM organizations WHERE id = $1;

-- name: GetOrganizationBySlug :one
SELECT * FROM organizations WHERE slug = $1;

-- name: ListOrganizationsByUser :many
SELECT o.* FROM organizations o
JOIN organization_members om ON o.id = om.organization_id
WHERE om.user_id = $1
ORDER BY o.created_at DESC;

-- name: UpdateOrganization :one
UPDATE organizations
SET name = COALESCE(sqlc.narg('name'), name),
    slug = COALESCE(sqlc.narg('slug'), slug),
    plan = COALESCE(sqlc.narg('plan'), plan),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteOrganization :exec
DELETE FROM organizations WHERE id = $1;

-- name: ListAllOrganizations :many
SELECT * FROM organizations ORDER BY created_at;
