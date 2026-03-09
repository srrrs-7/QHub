package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
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

// devBypassUserID is the fixed synthetic user ID injected when DEV_BYPASS_RBAC=true.
// Using a well-known UUID makes it easy to identify in logs and DB records.
var devBypassUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

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
//  1. If DEV_BYPASS_RBAC=true, all checks are skipped (development only).
//  2. If ApiKeyAuth set an org_id in context, the request is treated as having
//     org-level access and the role check is bypassed.
//  3. Otherwise, the user_id is read from the X-User-ID header (temporary until
//     JWT/Cognito integration provides identity via BearerAuth).
//
// The org_id is extracted from the chi URL parameter "org_id".
//
// On success, the member's role and user ID are stored in the request context
// for downstream handlers (accessible via GetMemberRole and GetUserID), and the
// WHO fields in the per-request RequestLog are populated for the HTTP Logger.
func RequireRole(q db.Querier, minRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Eagerly grab the RequestLog pointer once; every branch below writes
			// as many WHO fields as it can resolve, so even failed requests
			// carry maximum identity context in the http.request log entry.
			rl := logger.RequestLogFrom(r.Context())

			// DEV_BYPASS_RBAC=true skips all role checks (development only).
			// Set this when running the API locally without JWT/Cognito.
			if os.Getenv("DEV_BYPASS_RBAC") == "true" {
				rl.AuthMethod = "bypass"
				rl.UserID = devBypassUserID.String()
				// Capture org_id from URL if present (best-effort; may be empty for non-org routes).
				if orgIDStr := chi.URLParam(r, "org_id"); orgIDStr != "" {
					rl.OrgID = orgIDStr
				}
				ctx := context.WithValue(r.Context(), memberRoleKey, RoleOwner)
				ctx = context.WithValue(ctx, userIDKey, devBypassUserID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// 1. Check if API key auth already established org-level access
			if _, ok := GetApiKeyOrgID(r.Context()); ok {
				next.ServeHTTP(w, r)
				return
			}

			// 2. Parse org_id from URL — write to RequestLog immediately so it
			//    appears in the log even when subsequent steps fail.
			orgIDStr := chi.URLParam(r, "org_id")
			orgID, err := uuid.Parse(orgIDStr)
			if err != nil {
				writeRBACError(w, http.StatusBadRequest, "invalid organization ID")
				return
			}
			rl.OrgID = orgIDStr

			// 3. Extract user identity from X-User-ID header — write to RequestLog
			//    immediately so it appears in the log even when membership lookup fails.
			//    TODO: Replace with JWT claim extraction when Cognito is integrated
			userIDStr := r.Header.Get("X-User-ID")
			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				writeRBACError(w, http.StatusUnauthorized, "invalid or missing user identity")
				return
			}
			rl.UserID = userIDStr

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
				logger.Error("rbac.membership_lookup_failed",
					slog.Group("who",
						slog.String("user_id", userIDStr),
						slog.String("org_id", orgIDStr),
					),
					slog.Group("where",
						slog.String("layer", "middleware"),
						slog.String("component", "RequireRole"),
					),
					slog.Group("why",
						slog.String("outcome", "error"),
						slog.String("error", err.Error()),
					),
					slog.Group("how",
						slog.String("request_id", rl.RequestID),
					),
				)
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
