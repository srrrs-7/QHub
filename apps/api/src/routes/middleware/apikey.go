package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"
	"utils/db/db"
	"utils/logger"

	"github.com/google/uuid"
)

const (
	// apiKeyOrgIDKey is the context key for the organization ID from API key auth.
	apiKeyOrgIDKey contextKey = iota + 10
)

// ApiKeyAuth returns a middleware that validates API keys from the X-API-Key header.
// It hashes the provided key, looks it up in the database, checks expiry and revocation,
// and updates last_used_at. On success, it stores the organization ID in the request context.
func ApiKeyAuth(q db.Querier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawKey := r.Header.Get("X-API-Key")
			if rawKey == "" {
				unauthorized(w, "missing X-API-Key header")
				return
			}

			h := sha256.Sum256([]byte(rawKey))
			keyHash := hex.EncodeToString(h[:])

			apiKey, err := q.GetApiKeyByHash(r.Context(), keyHash)
			if err != nil {
				logger.Error("API key lookup failed", "error", err)
				unauthorized(w, "invalid API key")
				return
			}

			// Check if the key has been revoked
			if apiKey.RevokedAt.Valid {
				unauthorized(w, "API key has been revoked")
				return
			}

			// Check if the key has expired
			if apiKey.ExpiresAt.Valid && apiKey.ExpiresAt.Time.Before(time.Now()) {
				unauthorized(w, "API key has expired")
				return
			}

			// Update last_used_at in the background (best-effort)
			go func() {
				if err := q.UpdateApiKeyLastUsed(context.Background(), apiKey.ID); err != nil {
					logger.Error("failed to update API key last_used_at", "error", err)
				}
			}()

			// Store organization ID in context
			ctx := context.WithValue(r.Context(), apiKeyOrgIDKey, apiKey.OrganizationID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetApiKeyOrgID retrieves the organization ID set by API key auth middleware from context.
func GetApiKeyOrgID(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(apiKeyOrgIDKey)
	if v == nil {
		return uuid.UUID{}, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}
