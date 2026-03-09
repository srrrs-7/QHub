package apikeys

import (
	"api/src/domain/apperror"
	"api/src/routes/response"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"utils/db/db"
)

const (
	keyPrefix   = "pl_live_"
	keyRawBytes = 32
)

// Post creates a new API key, returning the raw key (shown only once).
func (h *ApiKeyHandler) Post() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decodePostRequest(r)
		if err != nil {
			response.HandleError(w, err)
			return
		}

		orgID, err := uuid.Parse(req.OrganizationID)
		if err != nil {
			response.HandleError(w, apperror.NewValidationError(err, "organization_id"))
			return
		}

		rawKey, keyHash, prefix, err := generateApiKey()
		if err != nil {
			response.HandleError(w, apperror.NewInternalServerError(err, "api_key"))
			return
		}

		apiKey, err := h.q.CreateApiKey(r.Context(), db.CreateApiKeyParams{
			OrganizationID: orgID,
			Name:           req.Name,
			KeyHash:        keyHash,
			KeyPrefix:      prefix,
		})
		if err != nil {
			response.HandleError(w, apperror.NewDatabaseError(err, "api_key"))
			return
		}

		response.Created(w, toApiKeyCreatedResponse(apiKey, rawKey))
	}
}

// generateApiKey generates a random API key and returns the raw key, its SHA-256 hash, and the prefix.
func generateApiKey() (rawKey string, hash string, prefix string, err error) {
	b := make([]byte, keyRawBytes)
	if _, err := rand.Read(b); err != nil {
		return "", "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	hexStr := hex.EncodeToString(b)
	rawKey = keyPrefix + hexStr
	prefix = keyPrefix + hexStr[:8]

	h := sha256.Sum256([]byte(rawKey))
	hash = hex.EncodeToString(h[:])

	return rawKey, hash, prefix, nil
}
