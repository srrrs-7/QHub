-- name: CreatePrompt :one
INSERT INTO prompts (project_id, name, slug, prompt_type, description)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetPrompt :one
SELECT * FROM prompts WHERE id = $1;

-- name: GetPromptByProjectAndSlug :one
SELECT * FROM prompts
WHERE project_id = $1 AND slug = $2;

-- name: ListPromptsByProject :many
SELECT * FROM prompts
WHERE project_id = $1
ORDER BY created_at DESC;

-- name: UpdatePrompt :one
UPDATE prompts
SET name = COALESCE(sqlc.narg('name'), name),
    slug = COALESCE(sqlc.narg('slug'), slug),
    description = COALESCE(sqlc.narg('description'), description),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdatePromptLatestVersion :one
UPDATE prompts
SET latest_version = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdatePromptProductionVersion :one
UPDATE prompts
SET production_version = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
