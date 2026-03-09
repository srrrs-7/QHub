package middleware

import (
	"net/http"
	"os"
	"strings"
)

const (
	corsAllowMethods  = "GET, POST, PUT, DELETE, OPTIONS"
	corsAllowHeaders  = "Authorization, Content-Type, X-API-Key"
	corsExposeHeaders = "X-Request-ID"
	corsMaxAge        = "300"
)

// CORS returns a middleware that handles Cross-Origin Resource Sharing.
// It accepts a corsOrigins string (comma-separated allowed origins).
// If corsOrigins is empty or whitespace-only, it falls back to the CORS_ORIGINS
// env var. If that is also empty, it defaults to "*" (allow all origins).
//
// When origins are set to "*", credentials are not supported (per the CORS spec).
// When specific origins are configured, Access-Control-Allow-Credentials is set to "true".
func CORS(corsOrigins string) func(http.Handler) http.Handler {
	origins := parseOrigins(corsOrigins)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// No Origin header means this is not a CORS request
			if origin == "" {
				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusNoContent)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			// Determine if the origin is allowed
			allowAll := len(origins) == 0
			allowed := allowAll

			if !allowAll {
				for _, o := range origins {
					if o == origin {
						allowed = true
						break
					}
				}
			}

			if !allowed {
				// Origin not allowed; do not set CORS headers
				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusNoContent)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			// Set CORS response headers
			if allowAll {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Vary", "Origin")
			}

			w.Header().Set("Access-Control-Expose-Headers", corsExposeHeaders)

			// Handle preflight
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", corsAllowMethods)
				w.Header().Set("Access-Control-Allow-Headers", corsAllowHeaders)
				w.Header().Set("Access-Control-Max-Age", corsMaxAge)
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// NewCORSFromEnv creates a CORS middleware reading allowed origins from the
// CORS_ORIGINS environment variable.
func NewCORSFromEnv() func(http.Handler) http.Handler {
	return CORS(os.Getenv("CORS_ORIGINS"))
}

// parseOrigins splits a comma-separated origins string and trims whitespace.
// Returns nil if the input is empty/whitespace-only (meaning allow all).
func parseOrigins(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}

	parts := strings.Split(trimmed, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		o := strings.TrimSpace(p)
		if o != "" {
			origins = append(origins, o)
		}
	}

	if len(origins) == 0 {
		return nil
	}
	return origins
}
