package search

import (
	"api/src/domain/apperror"
	"api/src/routes/response"
	"encoding/json"
	"net/http"
	"time"

	db "utils/db/db"

	"github.com/google/uuid"
)

// SemanticSearch handles POST /search/semantic.
// It takes a query text, generates an embedding, and finds similar prompt versions.
func (h *SearchHandler) SemanticSearch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !h.embSvc.Available() {
			response.HandleError(w, &unavailableError{})
			return
		}

		var req semanticSearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.HandleError(w, err)
			return
		}

		if req.Query == "" {
			response.HandleError(w, &validationError{msg: "query is required"})
			return
		}

		orgID, err := uuid.Parse(req.OrgID)
		if err != nil {
			response.HandleError(w, &validationError{msg: "invalid org_id"})
			return
		}

		limit := req.Limit
		if limit <= 0 || limit > 50 {
			limit = 10
		}

		// Generate embedding for the search query
		queryEmbedding, err := h.embSvc.GenerateEmbedding(r.Context(), req.Query)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		// Search by cosine similarity
		results, err := h.q.SearchPromptVersionsByEmbedding(r.Context(), db.SearchPromptVersionsByEmbeddingParams{
			Column1:        queryEmbedding,
			OrganizationID: orgID,
			Limit:          int32(limit),
		})
		if err != nil {
			response.HandleError(w, err)
			return
		}

		resp := make([]searchResultResponse, 0, len(results))
		for _, row := range results {
			if row.Similarity < req.MinScore {
				continue
			}
			resp = append(resp, toSearchResult(row))
		}

		response.OK(w, searchResponse{
			Query:   req.Query,
			Results: resp,
			Total:   len(resp),
		})
	}
}

type semanticSearchRequest struct {
	Query    string  `json:"query"`
	OrgID    string  `json:"org_id"`
	Limit    int     `json:"limit"`
	MinScore float64 `json:"min_score"`
}

type searchResponse struct {
	Query   string               `json:"query"`
	Results []searchResultResponse `json:"results"`
	Total   int                  `json:"total"`
}

type searchResultResponse struct {
	ID                string          `json:"id"`
	PromptID          string          `json:"prompt_id"`
	PromptName        string          `json:"prompt_name"`
	PromptSlug        string          `json:"prompt_slug"`
	VersionNumber     int             `json:"version_number"`
	Status            string          `json:"status"`
	Content           json.RawMessage `json:"content"`
	ChangeDescription string          `json:"change_description"`
	Similarity        float64         `json:"similarity"`
	CreatedAt         string          `json:"created_at"`
}

func toSearchResult(row db.SearchPromptVersionsByEmbeddingRow) searchResultResponse {
	changeDesc := ""
	if row.ChangeDescription.Valid {
		changeDesc = row.ChangeDescription.String
	}
	return searchResultResponse{
		ID:                row.ID.String(),
		PromptID:          row.PromptID.String(),
		PromptName:        row.PromptName,
		PromptSlug:        row.PromptSlug,
		VersionNumber:     int(row.VersionNumber),
		Status:            row.Status,
		Content:           row.Content,
		ChangeDescription: changeDesc,
		Similarity:        row.Similarity,
		CreatedAt:         row.CreatedAt.Format(time.RFC3339),
	}
}

// Custom error types for search-specific errors

type unavailableError struct{}

func (e *unavailableError) Error() string     { return "embedding service not available" }
func (e *unavailableError) ErrorName() string  { return "ServiceUnavailableError" }
func (e *unavailableError) DomainName() string { return "Search" }
func (e *unavailableError) Unwrap() error     { return nil }

type validationError struct{ msg string }

func (e *validationError) Error() string     { return e.msg }
func (e *validationError) ErrorName() string  { return "ValidationError" }
func (e *validationError) DomainName() string { return "Search" }
func (e *validationError) Unwrap() error     { return nil }

// Compile-time check: both error types satisfy apperror.AppError.
var (
	_ apperror.AppError = (*unavailableError)(nil)
	_ apperror.AppError = (*validationError)(nil)
)

// EmbeddingStatus returns a handler that shows embedding service status.
func (h *SearchHandler) EmbeddingStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := "disabled"
		if h.embSvc.Available() {
			if _, err := h.embSvc.GenerateEmbedding(r.Context(), "health check"); err == nil {
				status = "healthy"
			} else {
				status = "unhealthy"
			}
		}
		response.OK(w, map[string]string{"embedding_service": status})
	}
}


