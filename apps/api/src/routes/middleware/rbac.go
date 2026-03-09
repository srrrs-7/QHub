package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"utils/db/db"
	"utils/logger"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Role constants define the available roles in the organization membership system.
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
	RoleViewer = "viewer"
)

const (
	// memberRoleKey is the context key for the authenticated member's role.
	memberRoleKey contextKey = 20
	// userIDKey is the context key for the authenticated user's ID.
	userIDKey contextKey = 21
)

// roleLevels maps role names to numeric permission levels.
// Higher values indicate more permissions.
var roleLevels = map[string]int{
	RoleOwner:  4,
	RoleAdmin:  3,
	RoleMember: 2,
	RoleViewer: 1,
}

// RoleLevel returns the numeric permission level for a given role.
// Returns 0 for unknown roles.
func RoleLevel(role string) int {
	return roleLevels[role]
}

// RequireRole returns middleware that checks if the authenticated user has at least
// the specified minimum role in the organization.
//
// The middleware resolves identity in this order:
//  1. If ApiKeyAuth set an org_id in context, the request is treated as having
//     org-level access and the role check is bypassed.
//  2. Otherwise, the user_id is read from the X-User-ID header (temporary until
//     JWT/Cognito integration provides identity via BearerAuth).
//
// The org_id is extracted from the chi URL parameter "org_id".
//
// On success, the member's role and user ID are stored in the request context
// for downstream handlers (accessible via GetMemberRole and GetUserID).
func RequireRole(q db.Querier, minRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// DEV_BYPASS_RBAC=true skips all role checks (development only).
			// Set this when running the API locally without JWT/Cognito.
			if os.Getenv("DEV_BYPASS_RBAC") == "true" {
				ctx := context.WithValue(r.Context(), memberRoleKey, RoleOwner)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// 1. Check if API key auth already established org-level access
			if _, ok := GetApiKeyOrgID(r.Context()); ok {
				next.ServeHTTP(w, r)
				return
			}

			// 2. Parse org_id from URL
			orgIDStr := chi.URLParam(r, "org_id")
			orgID, err := uuid.Parse(orgIDStr)
			if err != nil {
				writeRBACError(w, http.StatusBadRequest, "invalid organization ID")
				return
			}

			// 3. Extract user identity from X-User-ID header
			//    TODO: Replace with JWT claim extraction when Cognito is integrated
			userIDStr := r.Header.Get("X-User-ID")
			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				writeRBACError(w, http.StatusUnauthorized, "invalid or missing user identity")
				return
			}

			// 4. Look up membership
			member, err := q.GetOrganizationMember(r.Context(), db.GetOrganizationMemberParams{
				OrganizationID: orgID,
				UserID:         userID,
			})
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					writeRBACError(w, http.StatusForbidden, "not a member of this organization")
					return
				}
				logger.Error("RBAC membership lookup failed", "error", err)
				writeRBACError(w, http.StatusInternalServerError, "internal server error")
				return
			}

			// 5. Check role level
			if RoleLevel(member.Role) < RoleLevel(minRole) {
				writeRBACError(w, http.StatusForbidden, "insufficient permissions: requires "+minRole+" role")
				return
			}

			// 6. Store role and user ID in context for downstream use
			ctx := context.WithValue(r.Context(), memberRoleKey, member.Role)
			ctx = context.WithValue(ctx, userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetMemberRole retrieves the authenticated member's role from the request context.
// Returns an empty string if no role is set.
func GetMemberRole(ctx context.Context) string {
	role, _ := ctx.Value(memberRoleKey).(string)
	return role
}

// GetUserID retrieves the authenticated user's ID from the request context.
// Returns the zero UUID and false if no user ID is set.
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(userIDKey)
	if v == nil {
		return uuid.UUID{}, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

// writeRBACError writes a JSON error response for RBAC failures.
func writeRBACError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	//nolint:errcheck
	json.NewEncoder(w).Encode(errorResponse{Error: message})
}
