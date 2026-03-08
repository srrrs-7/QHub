package industries

import (
	domain "api/src/domain/consulting"
	"database/sql"
	"encoding/json"
	"time"
	"utils/db/db"
)

type industryConfigResponse struct {
	ID              string          `json:"id"`
	Slug            string          `json:"slug"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	KnowledgeBase   json.RawMessage `json:"knowledge_base"`
	ComplianceRules json.RawMessage `json:"compliance_rules"`
	CreatedAt       string          `json:"created_at"`
	UpdatedAt       string          `json:"updated_at"`
}

func toIndustryConfigResponse(c domain.IndustryConfig) industryConfigResponse {
	return industryConfigResponse{
		ID:              c.ID.String(),
		Slug:            c.Slug,
		Name:            c.Name,
		Description:     c.Description,
		KnowledgeBase:   c.KnowledgeBase,
		ComplianceRules: c.ComplianceRules,
		CreatedAt:       c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       c.UpdatedAt.Format(time.RFC3339),
	}
}

type benchmarkResponse struct {
	ID                string `json:"id"`
	IndustryConfigID  string `json:"industry_config_id"`
	Period            string `json:"period"`
	AvgQualityScore   string `json:"avg_quality_score"`
	AvgLatencyMs      int    `json:"avg_latency_ms"`
	AvgCostPerRequest string `json:"avg_cost_per_request"`
	TotalExecutions   int64  `json:"total_executions"`
	P50Quality        string `json:"p50_quality"`
	P90Quality        string `json:"p90_quality"`
	OptInCount        int    `json:"opt_in_count"`
	CreatedAt         string `json:"created_at"`
}

func toBenchmarkResponse(b db.PlatformBenchmark) benchmarkResponse {
	return benchmarkResponse{
		ID:                b.ID.String(),
		IndustryConfigID:  b.IndustryConfigID.String(),
		Period:            b.Period,
		AvgQualityScore:   nullString(b.AvgQualityScore),
		AvgLatencyMs:      nullInt32(b.AvgLatencyMs),
		AvgCostPerRequest: nullString(b.AvgCostPerRequest),
		TotalExecutions:   b.TotalExecutions,
		P50Quality:        nullString(b.P50Quality),
		P90Quality:        nullString(b.P90Quality),
		OptInCount:        int(b.OptInCount),
		CreatedAt:         b.CreatedAt.Format(time.RFC3339),
	}
}

type complianceCheckResponse struct {
	Compliant  bool              `json:"compliant"`
	Violations []complianceIssue `json:"violations"`
}

type complianceIssue struct {
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

func nullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func nullInt32(ni sql.NullInt32) int {
	if ni.Valid {
		return int(ni.Int32)
	}
	return 0
}
