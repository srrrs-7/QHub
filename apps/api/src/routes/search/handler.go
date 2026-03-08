package search

import (
	"api/src/services/embeddingservice"
	db "utils/db/db"
)

// SearchHandler handles semantic search endpoints.
type SearchHandler struct {
	embSvc *embeddingservice.EmbeddingService
	q      db.Querier
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(embSvc *embeddingservice.EmbeddingService, q db.Querier) *SearchHandler {
	return &SearchHandler{embSvc: embSvc, q: q}
}
