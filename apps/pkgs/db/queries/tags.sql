-- name: CreateTag :one
INSERT INTO tags (organization_id, name, color) VALUES ($1, $2, $3) RETURNING *;

-- name: GetTag :one
SELECT * FROM tags WHERE id = $1;

-- name: ListTagsByOrg :many
SELECT * FROM tags WHERE organization_id = $1 ORDER BY name;

-- name: DeleteTag :exec
DELETE FROM tags WHERE id = $1;

-- name: AddPromptTag :exec
INSERT INTO prompt_tags (prompt_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: RemovePromptTag :exec
DELETE FROM prompt_tags WHERE prompt_id = $1 AND tag_id = $2;

-- name: ListTagsByPrompt :many
SELECT t.* FROM tags t JOIN prompt_tags pt ON t.id = pt.tag_id WHERE pt.prompt_id = $1 ORDER BY t.name;

-- name: ListPromptsByTag :many
SELECT p.* FROM prompts p JOIN prompt_tags pt ON p.id = pt.prompt_id WHERE pt.tag_id = $1;
