-- name: CreateApiKey :one
INSERT INTO api_keys (organization_id, name, key_hash, key_prefix)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetApiKeyByHash :one
SELECT * FROM api_keys
WHERE key_hash = $1 AND revoked_at IS NULL;

-- name: ListApiKeysByOrganization :many
SELECT * FROM api_keys
WHERE organization_id = $1
ORDER BY created_at DESC;

-- name: UpdateApiKeyLastUsed :exec
UPDATE api_keys
SET last_used_at = NOW()
WHERE id = $1;

-- name: RevokeApiKey :one
UPDATE api_keys
SET revoked_at = NOW()
WHERE id = $1 AND organization_id = $2
RETURNING *;
