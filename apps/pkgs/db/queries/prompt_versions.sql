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
SET status = @status::varchar,
    published_at = CASE WHEN @status::varchar = 'production' THEN NOW() ELSE published_at END
WHERE id = @id
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

-- name: UpdatePromptVersionEmbedding :exec
UPDATE prompt_versions
SET embedding = $2
WHERE id = $1;

-- name: SearchPromptVersionsByEmbedding :many
SELECT
    pv.id,
    pv.prompt_id,
    pv.version_number,
    pv.status,
    pv.content,
    pv.variables,
    pv.change_description,
    pv.author_id,
    pv.published_at,
    pv.created_at,
    p.name AS prompt_name,
    p.slug AS prompt_slug,
    cosine_similarity(pv.embedding, $1::real[]) AS similarity
FROM prompt_versions pv
JOIN prompts p ON pv.prompt_id = p.id
JOIN projects pr ON p.project_id = pr.id
WHERE pv.embedding IS NOT NULL
  AND pr.organization_id = $2
ORDER BY cosine_similarity(pv.embedding, $1::real[]) DESC
LIMIT $3;
