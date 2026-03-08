package logs

import (
	"api/src/domain/executionlog"
	"encoding/json"
)

type logResponse struct {
	ID            string          `json:"id"`
	OrgID         string          `json:"org_id"`
	PromptID      string          `json:"prompt_id"`
	VersionNumber int             `json:"version_number"`
	RequestBody   json.RawMessage `json:"request_body"`
	ResponseBody  json.RawMessage `json:"response_body"`
	Model         string          `json:"model"`
	Provider      string          `json:"provider"`
	InputTokens   int             `json:"input_tokens"`
	OutputTokens  int             `json:"output_tokens"`
	TotalTokens   int             `json:"total_tokens"`
	LatencyMs     int             `json:"latency_ms"`
	EstimatedCost string          `json:"estimated_cost"`
	Status        string          `json:"status"`
	ErrorMessage  string          `json:"error_message"`
	Environment   string          `json:"environment"`
	Metadata      json.RawMessage `json:"metadata"`
	ExecutedAt    string          `json:"executed_at"`
	CreatedAt     string          `json:"created_at"`
}

func toLogResponse(l executionlog.ExecutionLog) logResponse {
	return logResponse{
		ID:            l.ID.String(),
		OrgID:         l.OrgID.String(),
		PromptID:      l.PromptID.String(),
		VersionNumber: l.VersionNumber,
		RequestBody:   l.RequestBody,
		ResponseBody:  l.ResponseBody,
		Model:         l.Model,
		Provider:      l.Provider,
		InputTokens:   l.InputTokens,
		OutputTokens:  l.OutputTokens,
		TotalTokens:   l.TotalTokens,
		LatencyMs:     l.LatencyMs,
		EstimatedCost: l.EstimatedCost,
		Status:        l.Status,
		ErrorMessage:  l.ErrorMessage,
		Environment:   l.Environment,
		Metadata:      l.Metadata,
		ExecutedAt:    l.ExecutedAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt:     l.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

type listLogsResponse struct {
	Data  []logResponse `json:"data"`
	Total int64         `json:"total"`
}
