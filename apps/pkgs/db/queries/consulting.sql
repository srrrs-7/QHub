-- name: CreateConsultingSession :one
INSERT INTO consulting_sessions (organization_id, title, industry_config_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetConsultingSession :one
SELECT * FROM consulting_sessions WHERE id = $1;

-- name: ListConsultingSessionsByOrg :many
SELECT * FROM consulting_sessions
WHERE organization_id = $1
ORDER BY updated_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateConsultingSessionStatus :one
UPDATE consulting_sessions SET status = $2, updated_at = NOW() WHERE id = $1 RETURNING *;

-- name: CreateConsultingMessage :one
INSERT INTO consulting_messages (session_id, role, content, citations, actions_taken)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListConsultingMessages :many
SELECT * FROM consulting_messages WHERE session_id = $1 ORDER BY created_at ASC;

-- name: GetIndustryConfig :one
SELECT * FROM industry_configs WHERE id = $1;

-- name: GetIndustryConfigBySlug :one
SELECT * FROM industry_configs WHERE slug = $1;

-- name: ListIndustryConfigs :many
SELECT * FROM industry_configs ORDER BY name;

-- name: CreateIndustryConfig :one
INSERT INTO industry_configs (slug, name, description, knowledge_base, compliance_rules)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateIndustryConfig :one
UPDATE industry_configs SET
    name = COALESCE(NULLIF(@name::VARCHAR, ''), name),
    description = COALESCE(NULLIF(@description::TEXT, ''), description),
    knowledge_base = COALESCE(@knowledge_base, knowledge_base),
    compliance_rules = COALESCE(@compliance_rules, compliance_rules),
    updated_at = NOW()
WHERE id = @id
RETURNING *;
