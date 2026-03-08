package logs

import (
	"api/src/domain/executionlog"
	"api/src/routes/response"
	"net/http"

	"github.com/google/uuid"
)

func (h *LogHandler) Post() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decodePostLogRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		log := toLogEntity(req)

		created, err := h.logRepo.Create(r.Context(), log)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		response.Created(w, toLogResponse(created))
	}
}

func toLogEntity(req postLogRequest) executionlog.ExecutionLog {
	return executionlog.ExecutionLog{
		OrgID:         uuid.MustParse(req.OrgID),
		PromptID:      uuid.MustParse(req.PromptID),
		VersionNumber: req.VersionNumber,
		RequestBody:   req.RequestBody,
		ResponseBody:  req.ResponseBody,
		Model:         req.Model,
		Provider:      req.Provider,
		InputTokens:   req.InputTokens,
		OutputTokens:  req.OutputTokens,
		TotalTokens:   req.TotalTokens,
		LatencyMs:     req.LatencyMs,
		EstimatedCost: req.EstimatedCost,
		Status:        req.Status,
		ErrorMessage:  req.ErrorMessage,
		Environment:   req.Environment,
		Metadata:      req.Metadata,
		ExecutedAt:    req.ExecutedAt,
	}
}
