package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const tenantOrgIDKey contextKey = 30

// TenantContext extracts the organization ID from the request and stores it in context.
// It checks (in order): API key org_id context, URL param "org_id".
func TenantContext() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Check API key context
			if orgID, ok := GetApiKeyOrgID(r.Context()); ok {
				ctx := context.WithValue(r.Context(), tenantOrgIDKey, orgID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// 2. Check URL param org_id
			if orgIDStr := chi.URLParam(r, "org_id"); orgIDStr != "" {
				if orgID, err := uuid.Parse(orgIDStr); err == nil {
					ctx := context.WithValue(r.Context(), tenantOrgIDKey, orgID)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			// No tenant context - proceed without it (not all routes are org-scoped)
			next.ServeHTTP(w, r)
		})
	}
}

// GetTenantOrgID retrieves the organization ID from the request context.
func GetTenantOrgID(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(tenantOrgIDKey)
	if v == nil {
		return uuid.UUID{}, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}
