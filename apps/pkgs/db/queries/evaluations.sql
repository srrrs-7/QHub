-- name: CreateEvaluation :one
INSERT INTO evaluations (execution_log_id, overall_score, accuracy_score, relevance_score, fluency_score, safety_score, feedback, evaluator_type, evaluator_id, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetEvaluation :one
SELECT * FROM evaluations WHERE id = $1;

-- name: GetEvaluationByLog :one
SELECT * FROM evaluations WHERE execution_log_id = $1;

-- name: ListEvaluationsByLog :many
SELECT * FROM evaluations WHERE execution_log_id = $1 ORDER BY created_at DESC;

-- name: GetAvgScoresByPromptVersion :one
SELECT
    AVG(e.overall_score)::NUMERIC(3,2) AS avg_overall,
    AVG(e.accuracy_score)::NUMERIC(3,2) AS avg_accuracy,
    AVG(e.relevance_score)::NUMERIC(3,2) AS avg_relevance,
    AVG(e.fluency_score)::NUMERIC(3,2) AS avg_fluency,
    AVG(e.safety_score)::NUMERIC(3,2) AS avg_safety,
    COUNT(e.id) AS total_evaluations
FROM evaluations e
JOIN execution_logs el ON e.execution_log_id = el.id
WHERE el.prompt_id = $1 AND el.version_number = $2;
