package logs

import (
	"api/src/routes/requtil"
	"encoding/json"
	"net/http"
	"time"
)

type postLogRequest struct {
	OrgID         string          `json:"org_id" validate:"required,uuid"`
	PromptID      string          `json:"prompt_id" validate:"required,uuid"`
	VersionNumber int             `json:"version_number" validate:"required,min=1"`
	RequestBody   json.RawMessage `json:"request_body" validate:"required"`
	ResponseBody  json.RawMessage `json:"response_body"`
	Model         string          `json:"model" validate:"required,max=100"`
	Provider      string          `json:"provider" validate:"required,max=100"`
	InputTokens   int             `json:"input_tokens" validate:"min=0"`
	OutputTokens  int             `json:"output_tokens" validate:"min=0"`
	TotalTokens   int             `json:"total_tokens" validate:"min=0"`
	LatencyMs     int             `json:"latency_ms" validate:"min=0"`
	EstimatedCost string          `json:"estimated_cost" validate:"required"`
	Status        string          `json:"status" validate:"required,oneof=success error"`
	ErrorMessage  string          `json:"error_message" validate:"omitempty,max=2000"`
	Environment   string          `json:"environment" validate:"required,oneof=development staging production"`
	Metadata      json.RawMessage `json:"metadata"`
	ExecutedAt    time.Time       `json:"executed_at" validate:"required"`
}

func decodePostLogRequest(r *http.Request) (postLogRequest, error) {
	return requtil.Decode(r, func(req *postLogRequest) {
		req.ErrorMessage = requtil.Sanitize.Sanitize(req.ErrorMessage)
	})
}

type postBatchLogRequest struct {
	Logs []postLogRequest `json:"logs" validate:"required,min=1,max=100,dive"`
}

func decodePostBatchLogRequest(r *http.Request) (postBatchLogRequest, error) {
	return requtil.Decode[postBatchLogRequest](r, nil)
}
