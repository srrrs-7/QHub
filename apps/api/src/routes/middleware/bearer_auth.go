package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"utils/logger"
)

// contextKey is an unexported type to prevent context key collisions
type contextKey int

const (
	// bearerTokenKey is the context key for the bearer token
	bearerTokenKey contextKey = iota
)

// TokenValidator is a function type for validating bearer tokens
type TokenValidator func(token string) (bool, error)

// BearerAuth returns a middleware that validates Bearer tokens
func BearerAuth(validate TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				unauthorized(w, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				unauthorized(w, "invalid authorization header format")
				return
			}

			token := parts[1]
			if token == "" {
				unauthorized(w, "empty bearer token")
				return
			}

			valid, err := validate(token)
			if err != nil {
				logger.Error("token validation error", "error", err)
				unauthorized(w, "token validation failed")
				return
			}

			if !valid {
				unauthorized(w, "invalid token")
				return
			}

			// Store token in context for downstream handlers
			ctx := context.WithValue(r.Context(), bearerTokenKey, token)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

func unauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(errorResponse{Error: message})
}

// GetBearerToken retrieves the bearer token from context
func GetBearerToken(ctx context.Context) string {
	token, _ := ctx.Value(bearerTokenKey).(string)
	return token
}
