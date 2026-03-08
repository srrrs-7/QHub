-- name: CreatePlatformBenchmark :one
INSERT INTO platform_benchmarks (industry_config_id, period, avg_quality_score, avg_latency_ms, avg_cost_per_request, total_executions, p50_quality, p90_quality, opt_in_count)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (industry_config_id, period) DO UPDATE SET
    avg_quality_score = EXCLUDED.avg_quality_score,
    avg_latency_ms = EXCLUDED.avg_latency_ms,
    avg_cost_per_request = EXCLUDED.avg_cost_per_request,
    total_executions = EXCLUDED.total_executions,
    p50_quality = EXCLUDED.p50_quality,
    p90_quality = EXCLUDED.p90_quality,
    opt_in_count = EXCLUDED.opt_in_count
RETURNING *;

-- name: GetPlatformBenchmark :one
SELECT * FROM platform_benchmarks
WHERE industry_config_id = $1 AND period = $2;

-- name: ListPlatformBenchmarks :many
SELECT * FROM platform_benchmarks
WHERE industry_config_id = $1
ORDER BY period DESC
LIMIT $2;
