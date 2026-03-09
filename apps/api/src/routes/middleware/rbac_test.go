package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"utils/db/db"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

// mockQuerier implements a minimal db.Querier for RBAC tests.
// Only GetOrganizationMember is used by the middleware.
type mockQuerier struct {
	db.Querier
	member    db.OrganizationMember
	memberErr error
}

func (m *mockQuerier) GetOrganizationMember(_ context.Context, arg db.GetOrganizationMemberParams) (db.OrganizationMember, error) {
	return m.member, m.memberErr
}

func TestRequireRole(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()

	memberRecord := db.OrganizationMember{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           "member",
		JoinedAt:       time.Now(),
	}

	ownerRecord := db.OrganizationMember{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           "owner",
		JoinedAt:       time.Now(),
	}

	adminRecord := db.OrganizationMember{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           "admin",
		JoinedAt:       time.Now(),
	}

	viewerRecord := db.OrganizationMember{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           "viewer",
		JoinedAt:       time.Now(),
	}

	type args struct {
		minRole    string
		orgID      string // chi URL param "org_id"
		userIDHdr  string // X-User-ID header
		apiKeyOrg  *uuid.UUID
		member     db.OrganizationMember
		memberErr  error
	}
	type expected struct {
		statusCode int
		nextCalled bool
		errMsg     string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 (Happy Path)
		{
			testName: "owner accessing owner-required route passes",
			args: args{
				minRole:   RoleOwner,
				orgID:     orgID.String(),
				userIDHdr: userID.String(),
				member:    ownerRecord,
			},
			expected: expected{statusCode: http.StatusOK, nextCalled: true},
		},
		{
			testName: "owner accessing viewer-required route passes",
			args: args{
				minRole:   RoleViewer,
				orgID:     orgID.String(),
				userIDHdr: userID.String(),
				member:    ownerRecord,
			},
			expected: expected{statusCode: http.StatusOK, nextCalled: true},
		},
		{
			testName: "admin accessing member-required route passes",
			args: args{
				minRole:   RoleMember,
				orgID:     orgID.String(),
				userIDHdr: userID.String(),
				member:    adminRecord,
			},
			expected: expected{statusCode: http.StatusOK, nextCalled: true},
		},
		{
			testName: "API key auth bypasses role check",
			args: args{
				minRole:   RoleOwner,
				orgID:     orgID.String(),
				apiKeyOrg: &orgID,
			},
			expected: expected{statusCode: http.StatusOK, nextCalled: true},
		},

		// 異常系 (Error Cases)
		{
			testName: "viewer accessing admin-required route returns 403",
			args: args{
				minRole:   RoleAdmin,
				orgID:     orgID.String(),
				userIDHdr: userID.String(),
				member:    viewerRecord,
			},
			expected: expected{statusCode: http.StatusForbidden, nextCalled: false, errMsg: "insufficient permissions: requires admin role"},
		},
		{
			testName: "member accessing owner-required route returns 403",
			args: args{
				minRole:   RoleOwner,
				orgID:     orgID.String(),
				userIDHdr: userID.String(),
				member:    memberRecord,
			},
			expected: expected{statusCode: http.StatusForbidden, nextCalled: false, errMsg: "insufficient permissions: requires owner role"},
		},
		{
			testName: "non-existent member returns 403",
			args: args{
				minRole:   RoleViewer,
				orgID:     orgID.String(),
				userIDHdr: userID.String(),
				memberErr: sql.ErrNoRows,
			},
			expected: expected{statusCode: http.StatusForbidden, nextCalled: false, errMsg: "not a member of this organization"},
		},
		{
			testName: "database error returns 500",
			args: args{
				minRole:   RoleViewer,
				orgID:     orgID.String(),
				userIDHdr: userID.String(),
				memberErr: sql.ErrConnDone,
			},
			expected: expected{statusCode: http.StatusInternalServerError, nextCalled: false, errMsg: "internal server error"},
		},

		// 境界値 (Boundary Values)
		{
			testName: "exact role match member accessing member-required route passes",
			args: args{
				minRole:   RoleMember,
				orgID:     orgID.String(),
				userIDHdr: userID.String(),
				member:    memberRecord,
			},
			expected: expected{statusCode: http.StatusOK, nextCalled: true},
		},
		{
			testName: "exact role match viewer accessing viewer-required route passes",
			args: args{
				minRole:   RoleViewer,
				orgID:     orgID.String(),
				userIDHdr: userID.String(),
				member:    viewerRecord,
			},
			expected: expected{statusCode: http.StatusOK, nextCalled: true},
		},
		{
			testName: "viewer accessing member-required route returns 403",
			args: args{
				minRole:   RoleMember,
				orgID:     orgID.String(),
				userIDHdr: userID.String(),
				member:    viewerRecord,
			},
			expected: expected{statusCode: http.StatusForbidden, nextCalled: false, errMsg: "insufficient permissions: requires member role"},
		},

		// 特殊文字 (Special Chars)
		{
			testName: "invalid UUID in org_id returns 400",
			args: args{
				minRole:   RoleViewer,
				orgID:     "not-a-uuid",
				userIDHdr: userID.String(),
			},
			expected: expected{statusCode: http.StatusBadRequest, nextCalled: false, errMsg: "invalid organization ID"},
		},
		{
			testName: "invalid UUID in X-User-ID returns 401",
			args: args{
				minRole:   RoleViewer,
				orgID:     orgID.String(),
				userIDHdr: "not-a-uuid-🔑",
			},
			expected: expected{statusCode: http.StatusUnauthorized, nextCalled: false, errMsg: "invalid or missing user identity"},
		},

		// 空文字 (Empty/whitespace)
		{
			testName: "empty org_id returns 400",
			args: args{
				minRole:   RoleViewer,
				orgID:     "",
				userIDHdr: userID.String(),
			},
			expected: expected{statusCode: http.StatusBadRequest, nextCalled: false, errMsg: "invalid organization ID"},
		},
		{
			testName: "missing user identity without API key returns 401",
			args: args{
				minRole:   RoleViewer,
				orgID:     orgID.String(),
				userIDHdr: "",
			},
			expected: expected{statusCode: http.StatusUnauthorized, nextCalled: false, errMsg: "invalid or missing user identity"},
		},

		// Null/Nil — API key with nil org
		{
			testName: "no API key and no X-User-ID returns 401",
			args: args{
				minRole:   RoleViewer,
				orgID:     orgID.String(),
				apiKeyOrg: nil,
				userIDHdr: "",
			},
			expected: expected{statusCode: http.StatusUnauthorized, nextCalled: false, errMsg: "invalid or missing user identity"},
		},

		// All role combinations
		{
			testName: "admin accessing admin-required route passes",
			args: args{
				minRole:   RoleAdmin,
				orgID:     orgID.String(),
				userIDHdr: userID.String(),
				member:    adminRecord,
			},
			expected: expected{statusCode: http.StatusOK, nextCalled: true},
		},
		{
			testName: "member accessing admin-required route returns 403",
			args: args{
				minRole:   RoleAdmin,
				orgID:     orgID.String(),
				userIDHdr: userID.String(),
				member:    memberRecord,
			},
			expected: expected{statusCode: http.StatusForbidden, nextCalled: false, errMsg: "insufficient permissions: requires admin role"},
		},
		{
			testName: "unknown role in DB returns 403",
			args: args{
				minRole:   RoleViewer,
				orgID:     orgID.String(),
				userIDHdr: userID.String(),
				member: db.OrganizationMember{
					OrganizationID: orgID,
					UserID:         userID,
					Role:           "unknown_role",
					JoinedAt:       time.Now(),
				},
			},
			expected: expected{statusCode: http.StatusForbidden, nextCalled: false, errMsg: "insufficient permissions: requires viewer role"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			nextCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			mock := &mockQuerier{
				member:    tt.args.member,
				memberErr: tt.args.memberErr,
			}

			mw := RequireRole(mock, tt.args.minRole)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			// Set chi URL param org_id
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("org_id", tt.args.orgID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Set X-User-ID header if provided
			if tt.args.userIDHdr != "" {
				req.Header.Set("X-User-ID", tt.args.userIDHdr)
			}

			// Set API key org in context if provided
			if tt.args.apiKeyOrg != nil {
				ctx := context.WithValue(req.Context(), apiKeyOrgIDKey, *tt.args.apiKeyOrg)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			mw(nextHandler).ServeHTTP(w, req)

			resp := w.Result()

			if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
				t.Errorf("status code mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.expected.nextCalled, nextCalled); diff != "" {
				t.Errorf("nextCalled mismatch (-want +got):\n%s", diff)
			}

			if tt.expected.errMsg != "" {
				var body errorResponse
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode error response: %v", err)
				}
				if diff := cmp.Diff(tt.expected.errMsg, body.Error); diff != "" {
					t.Errorf("error message mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestRoleLevel(t *testing.T) {
	type args struct {
		role string
	}
	type expected struct {
		level int
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 (Happy Path)
		{testName: "owner level is 4", args: args{role: RoleOwner}, expected: expected{level: 4}},
		{testName: "admin level is 3", args: args{role: RoleAdmin}, expected: expected{level: 3}},
		{testName: "member level is 2", args: args{role: RoleMember}, expected: expected{level: 2}},
		{testName: "viewer level is 1", args: args{role: RoleViewer}, expected: expected{level: 1}},

		// 異常系 (Error Cases)
		{testName: "unknown role returns 0", args: args{role: "superadmin"}, expected: expected{level: 0}},

		// 空文字 (Empty/whitespace)
		{testName: "empty string returns 0", args: args{role: ""}, expected: expected{level: 0}},
		{testName: "whitespace returns 0", args: args{role: "  "}, expected: expected{level: 0}},

		// 特殊文字 (Special Chars)
		{testName: "emoji role returns 0", args: args{role: "👑"}, expected: expected{level: 0}},
		{testName: "SQL injection returns 0", args: args{role: "'; DROP TABLE --"}, expected: expected{level: 0}},

		// 境界値 (Boundary Values)
		{testName: "owner is highest", args: args{role: RoleOwner}, expected: expected{level: 4}},
		{testName: "viewer is lowest valid", args: args{role: RoleViewer}, expected: expected{level: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := RoleLevel(tt.args.role)
			if diff := cmp.Diff(tt.expected.level, got); diff != "" {
				t.Errorf("level mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetMemberRole(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	type expected struct {
		role string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 (Happy Path)
		{
			testName: "returns role from context",
			args:     args{ctx: context.WithValue(context.Background(), memberRoleKey, "admin")},
			expected: expected{role: "admin"},
		},
		// 異常系 (Error Cases)
		{
			testName: "returns empty string when context has wrong type",
			args:     args{ctx: context.WithValue(context.Background(), memberRoleKey, 123)},
			expected: expected{role: ""},
		},
		// 空文字 (Empty/whitespace)
		{
			testName: "returns empty string when no role in context",
			args:     args{ctx: context.Background()},
			expected: expected{role: ""},
		},
		// Null/Nil
		{
			testName: "returns empty string for nil value in context",
			args:     args{ctx: context.WithValue(context.Background(), memberRoleKey, nil)},
			expected: expected{role: ""},
		},
		// 特殊文字 (Special Chars)
		{
			testName: "returns role with special characters",
			args:     args{ctx: context.WithValue(context.Background(), memberRoleKey, "ロール🔐")},
			expected: expected{role: "ロール🔐"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := GetMemberRole(tt.args.ctx)
			if diff := cmp.Diff(tt.expected.role, got); diff != "" {
				t.Errorf("role mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	validID := uuid.New()

	type args struct {
		ctx context.Context
	}
	type expected struct {
		userID uuid.UUID
		ok     bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 (Happy Path)
		{
			testName: "returns user ID from context",
			args:     args{ctx: context.WithValue(context.Background(), userIDKey, validID)},
			expected: expected{userID: validID, ok: true},
		},
		// 異常系 (Error Cases)
		{
			testName: "returns false when context has wrong type",
			args:     args{ctx: context.WithValue(context.Background(), userIDKey, "not-uuid")},
			expected: expected{userID: uuid.UUID{}, ok: false},
		},
		// 空文字 (Empty/whitespace)
		{
			testName: "returns false when no user ID in context",
			args:     args{ctx: context.Background()},
			expected: expected{userID: uuid.UUID{}, ok: false},
		},
		// Null/Nil
		{
			testName: "returns false for nil value in context",
			args:     args{ctx: context.WithValue(context.Background(), userIDKey, nil)},
			expected: expected{userID: uuid.UUID{}, ok: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, ok := GetUserID(tt.args.ctx)
			if diff := cmp.Diff(tt.expected.ok, ok); diff != "" {
				t.Errorf("ok mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected.userID, got); diff != "" {
				t.Errorf("userID mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
