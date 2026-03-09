package middleware

import (
	"net/http"
	"strings"
	"utils/db/db"
)

// BearerOrApiKeyAuth returns a middleware that tries Bearer auth first,
// then falls back to API Key auth. If both fail, it returns 401 Unauthorized.
func BearerOrApiKeyAuth(validate TokenValidator, q db.Querier) func(http.Handler) http.Handler {
	bearerMw := BearerAuth(validate)
	apiKeyMw := ApiKeyAuth(q)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try Bearer auth if Authorization header is present
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
				bearerMw(next).ServeHTTP(w, r)
				return
			}

			// Try API Key auth if X-API-Key header is present
			apiKey := r.Header.Get("X-API-Key")
			if apiKey != "" {
				apiKeyMw(next).ServeHTTP(w, r)
				return
			}

			// Neither auth method provided
			unauthorized(w, "missing authentication: provide Authorization or X-API-Key header")
		})
	}
}
