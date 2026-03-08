-- name: CreatePromptVersion :one
INSERT INTO prompt_versions (prompt_id, version_number, status, content, variables, change_description, author_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetPromptVersion :one
SELECT * FROM prompt_versions
WHERE prompt_id = $1 AND version_number = $2;

-- name: GetPromptVersionByID :one
SELECT * FROM prompt_versions WHERE id = $1;

-- name: ListPromptVersions :many
SELECT * FROM prompt_versions
WHERE prompt_id = $1
ORDER BY version_number DESC;

-- name: GetLatestPromptVersion :one
SELECT * FROM prompt_versions
WHERE prompt_id = $1
ORDER BY version_number DESC
LIMIT 1;

-- name: GetProductionPromptVersion :one
SELECT * FROM prompt_versions
WHERE prompt_id = $1 AND status = 'production'
LIMIT 1;

-- name: UpdatePromptVersionStatus :one
UPDATE prompt_versions
SET status = $2,
    published_at = CASE WHEN $2 = 'production' THEN NOW() ELSE published_at END
WHERE id = $1
RETURNING *;

-- name: UpdatePromptVersionSemanticDiff :exec
UPDATE prompt_versions
SET semantic_diff = $2
WHERE id = $1;

-- name: UpdatePromptVersionLintResult :exec
UPDATE prompt_versions
SET lint_result = $2
WHERE id = $1;

-- name: ArchiveProductionVersion :exec
UPDATE prompt_versions
SET status = 'archived'
WHERE prompt_id = $1 AND status = 'production';
