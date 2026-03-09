package admin

import (
	"api/src/services/batchservice"
)

type aggregationResponse struct {
	OrgsProcessed       int    `json:"orgs_processed"`
	PromptsProcessed    int    `json:"prompts_processed"`
	ExecutionsProcessed int    `json:"executions_processed"`
	Period              string `json:"period"`
	DurationMs          int64  `json:"duration_ms"`
}

func toAggregationResponse(r *batchservice.AggregationResult) aggregationResponse {
	return aggregationResponse{
		OrgsProcessed:       r.OrgsProcessed,
		PromptsProcessed:    r.PromptsProcessed,
		ExecutionsProcessed: r.ExecutionsProcessed,
		Period:              r.Period,
		DurationMs:          r.Duration.Milliseconds(),
	}
}
