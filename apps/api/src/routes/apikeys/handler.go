package apikeys

import (
	"utils/db/db"
)

// ApiKeyHandler handles API key management endpoints.
type ApiKeyHandler struct {
	q db.Querier
}

// NewApiKeyHandler creates a new ApiKeyHandler with the given querier.
func NewApiKeyHandler(q db.Querier) *ApiKeyHandler {
	return &ApiKeyHandler{q: q}
}
