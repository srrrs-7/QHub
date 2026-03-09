-- name: CreateExecutionLog :one
INSERT INTO execution_logs (organization_id, prompt_id, version_number, request_body, response_body, model, provider, input_tokens, output_tokens, total_tokens, latency_ms, estimated_cost, status, error_message, environment, metadata, executed_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
RETURNING *;

-- name: GetExecutionLog :one
SELECT * FROM execution_logs WHERE id = $1;

-- name: ListExecutionLogsByPrompt :many
SELECT * FROM execution_logs
WHERE prompt_id = $1
ORDER BY executed_at DESC
LIMIT $2 OFFSET $3;

-- name: ListExecutionLogsByOrg :many
SELECT * FROM execution_logs
WHERE organization_id = $1
ORDER BY executed_at DESC
LIMIT $2 OFFSET $3;

-- name: CountExecutionLogsByPrompt :one
SELECT COUNT(*) FROM execution_logs WHERE prompt_id = $1;

-- name: CountExecutionLogsByOrg :one
SELECT COUNT(*) FROM execution_logs WHERE organization_id = $1;

-- name: GetPromptVersionAnalytics :one
SELECT
    prompt_id,
    version_number,
    COUNT(*) AS total_executions,
    AVG(total_tokens)::INTEGER AS avg_tokens,
    AVG(latency_ms)::INTEGER AS avg_latency_ms,
    SUM(estimated_cost)::NUMERIC(10,6) AS total_cost,
    AVG(estimated_cost)::NUMERIC(10,6) AS avg_cost,
    COUNT(*) FILTER (WHERE status = 'success') AS success_count,
    COUNT(*) FILTER (WHERE status = 'error') AS error_count
FROM execution_logs
WHERE prompt_id = $1 AND version_number = $2
GROUP BY prompt_id, version_number;

-- name: GetPromptAnalytics :many
SELECT
    version_number,
    COUNT(*) AS total_executions,
    AVG(total_tokens)::INTEGER AS avg_tokens,
    AVG(latency_ms)::INTEGER AS avg_latency_ms,
    SUM(estimated_cost)::NUMERIC(10,6) AS total_cost,
    COUNT(*) FILTER (WHERE status = 'success') AS success_count,
    COUNT(*) FILTER (WHERE status = 'error') AS error_count
FROM execution_logs
WHERE prompt_id = $1
GROUP BY version_number
ORDER BY version_number DESC;

-- name: GetDailyTrend :many
SELECT
    DATE(executed_at) AS day,
    COUNT(*) AS total_executions,
    AVG(total_tokens)::INTEGER AS avg_tokens,
    AVG(latency_ms)::INTEGER AS avg_latency_ms,
    SUM(estimated_cost)::NUMERIC(10,6) AS total_cost
FROM execution_logs
WHERE prompt_id = $1 AND executed_at >= $2 AND executed_at < $3
GROUP BY DATE(executed_at)
ORDER BY day;

-- name: GetProjectAnalytics :many
SELECT
    p.id AS prompt_id,
    p.name AS prompt_name,
    COUNT(el.id) AS total_executions,
    AVG(el.total_tokens)::INTEGER AS avg_tokens,
    AVG(el.latency_ms)::INTEGER AS avg_latency_ms,
    SUM(el.estimated_cost)::NUMERIC(10,6) AS total_cost
FROM prompts p
LEFT JOIN execution_logs el ON el.prompt_id = p.id
WHERE p.project_id = $1
GROUP BY p.id, p.name
ORDER BY total_executions DESC;

-- name: GetVersionMetrics :many
SELECT
    latency_ms,
    total_tokens,
    COALESCE(e.overall_score, 0)::NUMERIC(5,2) AS overall_score
FROM execution_logs el
LEFT JOIN evaluations e ON e.execution_log_id = el.id
WHERE el.prompt_id = $1 AND el.version_number = $2
ORDER BY el.executed_at;

-- name: GetOrgMonthlyMetrics :one
SELECT
    COUNT(*)::BIGINT AS execution_count,
    COALESCE(AVG(el.latency_ms), 0)::INTEGER AS avg_latency_ms,
    COALESCE(SUM(el.total_tokens), 0)::BIGINT AS total_tokens,
    COALESCE(AVG(e.overall_score::NUMERIC), 0)::NUMERIC(5,2) AS avg_score,
    COUNT(DISTINCT el.prompt_id)::BIGINT AS active_prompts
FROM execution_logs el
LEFT JOIN evaluations e ON e.execution_log_id = el.id
WHERE el.organization_id = $1
    AND el.executed_at >= $2
    AND el.executed_at < $3;
