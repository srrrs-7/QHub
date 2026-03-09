package members

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"utils/db/db"
	"utils/testutil"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

// setupTest creates a test transaction, an organization, and a user for seeding.
func setupTest(t *testing.T) (db.Querier, db.Organization, db.User) {
	t.Helper()
	q := testutil.SetupTestTx(t)
	ctx := context.Background()

	org, err := q.CreateOrganization(ctx, db.CreateOrganizationParams{
		Name: "Test Org",
		Slug: "test-org",
		Plan: "free",
	})
	if err != nil {
		t.Fatalf("failed to create organization: %v", err)
	}

	user, err := q.CreateUser(ctx, db.CreateUserParams{
		Email: "test@example.com",
		Name:  "Test User",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	return q, org, user
}

// createUser creates an additional user with a unique email.
func createUser(t *testing.T, q db.Querier, suffix string) db.User {
	t.Helper()
	user, err := q.CreateUser(context.Background(), db.CreateUserParams{
		Email: fmt.Sprintf("user-%s@example.com", suffix),
		Name:  fmt.Sprintf("User %s", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	return user
}

// addMember seeds a member into the organization directly via the querier.
func addMember(t *testing.T, q db.Querier, orgID, userID uuid.UUID, role string) db.OrganizationMember {
	t.Helper()
	member, err := q.AddOrganizationMember(context.Background(), db.AddOrganizationMemberParams{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           role,
	})
	if err != nil {
		t.Fatalf("failed to add member: %v", err)
	}
	return member
}

// setChiURLParams creates an http.Request with chi URL params set.
func setChiURLParams(req *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

// --- Post Tests ---

func TestPostHandler(t *testing.T) {
	t.Run("201 Created", func(t *testing.T) {
		type expected struct {
			statusCode int
			role       string
		}

		tests := []struct {
			testName string
			role     string
			expected expected
		}{
			// 正常系
			{
				testName: "add member with role member",
				role:     "member",
				expected: expected{statusCode: http.StatusCreated, role: "member"},
			},
			// 境界値 - each valid role
			{
				testName: "add member with role owner",
				role:     "owner",
				expected: expected{statusCode: http.StatusCreated, role: "owner"},
			},
			{
				testName: "add member with role admin",
				role:     "admin",
				expected: expected{statusCode: http.StatusCreated, role: "admin"},
			},
			{
				testName: "add member with role viewer",
				role:     "viewer",
				expected: expected{statusCode: http.StatusCreated, role: "viewer"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q, org, user := setupTest(t)

				body := map[string]string{
					"user_id": user.ID.String(),
					"role":    tt.role,
				}
				jsonBody, err := json.Marshal(body)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}

				req := httptest.NewRequest(http.MethodPost, "/organizations/"+org.ID.String()+"/members", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)
				req = setChiURLParams(req, map[string]string{"org_id": org.ID.String()})

				w := httptest.NewRecorder()
				handler := NewMemberHandler(q).Post()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var respBody memberResponse
				if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(org.ID.String(), respBody.OrganizationID); diff != "" {
					t.Errorf("organization_id mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(user.ID.String(), respBody.UserID); diff != "" {
					t.Errorf("user_id mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expected.role, respBody.Role); diff != "" {
					t.Errorf("role mismatch (-want +got):\n%s", diff)
				}
				if respBody.JoinedAt == "" {
					t.Error("expected non-empty joined_at")
				}
			})
		}
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		tests := []struct {
			testName string
			body     string
		}{
			// 異常系 - invalid role
			{
				testName: "invalid role value",
				body:     `{"user_id":"` + uuid.New().String() + `","role":"superadmin"}`,
			},
			// 空文字 - empty role
			{
				testName: "empty role",
				body:     `{"user_id":"` + uuid.New().String() + `","role":""}`,
			},
			// 空文字 - empty user_id
			{
				testName: "empty user_id",
				body:     `{"user_id":"","role":"member"}`,
			},
			// Null/Nil - missing fields
			{
				testName: "missing role field",
				body:     `{"user_id":"` + uuid.New().String() + `"}`,
			},
			{
				testName: "missing user_id field",
				body:     `{"role":"member"}`,
			},
			{
				testName: "empty JSON object",
				body:     `{}`,
			},
			// 異常系 - invalid JSON
			{
				testName: "invalid JSON",
				body:     `not valid json`,
			},
			// 特殊文字 - invalid UUID with special chars
			{
				testName: "user_id with special characters",
				body:     `{"user_id":"not-a-uuid-<script>","role":"member"}`,
			},
			// 異常系 - malformed UUID
			{
				testName: "malformed user_id UUID",
				body:     `{"user_id":"12345","role":"member"}`,
			},
			// 特殊文字 - role with unicode
			{
				testName: "role with Japanese characters",
				body:     `{"user_id":"` + uuid.New().String() + `","role":"メンバー"}`,
			},
			// 空文字 - whitespace only role
			{
				testName: "whitespace only role",
				body:     `{"user_id":"` + uuid.New().String() + `","role":"   "}`,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q, org, _ := setupTest(t)

				req := httptest.NewRequest(http.MethodPost, "/organizations/"+org.ID.String()+"/members", strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)
				req = setChiURLParams(req, map[string]string{"org_id": org.ID.String()})

				w := httptest.NewRecorder()
				handler := NewMemberHandler(q).Post()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
					t.Errorf("expected error status, got %d", resp.StatusCode)
				}
			})
		}
	})

	t.Run("invalid org_id", func(t *testing.T) {
		tests := []struct {
			testName string
			orgID    string
		}{
			// 異常系 - non-UUID org_id
			{testName: "non-uuid org_id", orgID: "not-a-uuid"},
			// 空文字
			{testName: "empty org_id", orgID: ""},
			// 特殊文字
			{testName: "special chars org_id", orgID: "<script>alert('xss')</script>"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q, _, user := setupTest(t)

				body := fmt.Sprintf(`{"user_id":"%s","role":"member"}`, user.ID.String())
				req := httptest.NewRequest(http.MethodPost, "/organizations/"+tt.orgID+"/members", strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)
				req = setChiURLParams(req, map[string]string{"org_id": tt.orgID})

				w := httptest.NewRecorder()
				handler := NewMemberHandler(q).Post()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode == http.StatusCreated {
					t.Errorf("expected error status, got %d", resp.StatusCode)
				}
			})
		}
	})

	t.Run("database error - non-existent user FK violation", func(t *testing.T) {
		q, org, _ := setupTest(t)

		nonExistentUserID := uuid.New()
		body := fmt.Sprintf(`{"user_id":"%s","role":"member"}`, nonExistentUserID.String())
		req := httptest.NewRequest(http.MethodPost, "/organizations/"+org.ID.String()+"/members", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)
		req = setChiURLParams(req, map[string]string{"org_id": org.ID.String()})

		w := httptest.NewRecorder()
		handler := NewMemberHandler(q).Post()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode == http.StatusCreated {
			t.Errorf("expected error status for non-existent user, got %d", resp.StatusCode)
		}
	})

	t.Run("database error - duplicate member", func(t *testing.T) {
		q, org, user := setupTest(t)

		// Add the member first
		addMember(t, q, org.ID, user.ID, "member")

		// Try to add the same member again
		body := fmt.Sprintf(`{"user_id":"%s","role":"admin"}`, user.ID.String())
		req := httptest.NewRequest(http.MethodPost, "/organizations/"+org.ID.String()+"/members", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)
		req = setChiURLParams(req, map[string]string{"org_id": org.ID.String()})

		w := httptest.NewRecorder()
		handler := NewMemberHandler(q).Post()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode == http.StatusCreated {
			t.Errorf("expected error status for duplicate member, got %d", resp.StatusCode)
		}
	})
}

// --- List Tests ---

func TestListHandler(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		tests := []struct {
			testName      string
			memberCount   int
			expectedCount int
		}{
			// 正常系
			{testName: "list members with one member", memberCount: 1, expectedCount: 1},
			// 境界値
			{testName: "list members with multiple members", memberCount: 3, expectedCount: 3},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q, org, user := setupTest(t)

				// Seed members
				addMember(t, q, org.ID, user.ID, "owner")
				for i := 1; i < tt.memberCount; i++ {
					u := createUser(t, q, fmt.Sprintf("list-%d-%d", i, org.ID.ClockSequence()))
					addMember(t, q, org.ID, u.ID, "member")
				}

				req := httptest.NewRequest(http.MethodGet, "/organizations/"+org.ID.String()+"/members", nil)
				testutil.SetAuthHeader(req)
				req = setChiURLParams(req, map[string]string{"org_id": org.ID.String()})

				w := httptest.NewRecorder()
				handler := NewMemberHandler(q).List()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(http.StatusOK, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var body []memberResponse
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(tt.expectedCount, len(body)); diff != "" {
					t.Errorf("member count mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("200 OK empty list", func(t *testing.T) {
		// 境界値 - no members
		q, org, _ := setupTest(t)

		req := httptest.NewRequest(http.MethodGet, "/organizations/"+org.ID.String()+"/members", nil)
		testutil.SetAuthHeader(req)
		req = setChiURLParams(req, map[string]string{"org_id": org.ID.String()})

		w := httptest.NewRecorder()
		handler := NewMemberHandler(q).List()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if diff := cmp.Diff(http.StatusOK, resp.StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var body []memberResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if body != nil && len(body) != 0 {
			t.Errorf("expected empty list, got %d members", len(body))
		}
	})

	t.Run("invalid org_id", func(t *testing.T) {
		tests := []struct {
			testName string
			orgID    string
		}{
			// 異常系
			{testName: "non-uuid org_id", orgID: "not-a-uuid"},
			// 空文字
			{testName: "empty org_id", orgID: ""},
			// 特殊文字
			{testName: "sql injection org_id", orgID: "drop-table-members"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				req := httptest.NewRequest(http.MethodGet, "/organizations/"+tt.orgID+"/members", nil)
				testutil.SetAuthHeader(req)
				req = setChiURLParams(req, map[string]string{"org_id": tt.orgID})

				w := httptest.NewRecorder()
				handler := NewMemberHandler(q).List()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode == http.StatusOK {
					t.Errorf("expected error status for invalid org_id, got %d", resp.StatusCode)
				}
			})
		}
	})

	t.Run("200 OK with non-existent org returns empty", func(t *testing.T) {
		// 異常系 - valid UUID but org does not exist (no FK on list query)
		q := testutil.SetupTestTx(t)
		nonExistentOrgID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/organizations/"+nonExistentOrgID.String()+"/members", nil)
		testutil.SetAuthHeader(req)
		req = setChiURLParams(req, map[string]string{"org_id": nonExistentOrgID.String()})

		w := httptest.NewRecorder()
		handler := NewMemberHandler(q).List()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if diff := cmp.Diff(http.StatusOK, resp.StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("200 OK verify response fields", func(t *testing.T) {
		// 正常系 - verify all response fields are populated
		q, org, user := setupTest(t)
		addMember(t, q, org.ID, user.ID, "admin")

		req := httptest.NewRequest(http.MethodGet, "/organizations/"+org.ID.String()+"/members", nil)
		testutil.SetAuthHeader(req)
		req = setChiURLParams(req, map[string]string{"org_id": org.ID.String()})

		w := httptest.NewRecorder()
		handler := NewMemberHandler(q).List()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if diff := cmp.Diff(http.StatusOK, resp.StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var body []memberResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(body) != 1 {
			t.Fatalf("expected 1 member, got %d", len(body))
		}

		m := body[0]
		if diff := cmp.Diff(org.ID.String(), m.OrganizationID); diff != "" {
			t.Errorf("organization_id mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff(user.ID.String(), m.UserID); diff != "" {
			t.Errorf("user_id mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff("admin", m.Role); diff != "" {
			t.Errorf("role mismatch (-want +got):\n%s", diff)
		}
		if m.JoinedAt == "" {
			t.Error("expected non-empty joined_at")
		}
	})
}

// --- Put Tests ---

func TestPutHandler(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		type expected struct {
			statusCode int
			role       string
		}

		tests := []struct {
			testName    string
			initialRole string
			newRole     string
			expected    expected
		}{
			// 正常系
			{
				testName:    "update role from member to admin",
				initialRole: "member",
				newRole:     "admin",
				expected:    expected{statusCode: http.StatusOK, role: "admin"},
			},
			// 境界値 - all valid role transitions
			{
				testName:    "update role to owner",
				initialRole: "member",
				newRole:     "owner",
				expected:    expected{statusCode: http.StatusOK, role: "owner"},
			},
			{
				testName:    "update role to viewer",
				initialRole: "admin",
				newRole:     "viewer",
				expected:    expected{statusCode: http.StatusOK, role: "viewer"},
			},
			{
				testName:    "update role to member",
				initialRole: "owner",
				newRole:     "member",
				expected:    expected{statusCode: http.StatusOK, role: "member"},
			},
			// 境界値 - same role
			{
				testName:    "update role to same role",
				initialRole: "admin",
				newRole:     "admin",
				expected:    expected{statusCode: http.StatusOK, role: "admin"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q, org, user := setupTest(t)
				addMember(t, q, org.ID, user.ID, tt.initialRole)

				body := fmt.Sprintf(`{"role":"%s"}`, tt.newRole)
				req := httptest.NewRequest(http.MethodPut, "/organizations/"+org.ID.String()+"/members/"+user.ID.String(), strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)
				req = setChiURLParams(req, map[string]string{
					"org_id":  org.ID.String(),
					"user_id": user.ID.String(),
				})

				w := httptest.NewRecorder()
				handler := NewMemberHandler(q).Put()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var respBody memberResponse
				if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(tt.expected.role, respBody.Role); diff != "" {
					t.Errorf("role mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(org.ID.String(), respBody.OrganizationID); diff != "" {
					t.Errorf("organization_id mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(user.ID.String(), respBody.UserID); diff != "" {
					t.Errorf("user_id mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		tests := []struct {
			testName string
			body     string
		}{
			// 異常系 - invalid role
			{testName: "invalid role", body: `{"role":"superadmin"}`},
			// 空文字
			{testName: "empty role", body: `{"role":""}`},
			// Null/Nil - missing role
			{testName: "missing role field", body: `{}`},
			// 異常系 - invalid JSON
			{testName: "invalid JSON", body: `not json`},
			// 特殊文字
			{testName: "role with emoji", body: `{"role":"admin🔥"}`},
			// 空文字 - whitespace
			{testName: "whitespace role", body: `{"role":"  "}`},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q, org, user := setupTest(t)
				addMember(t, q, org.ID, user.ID, "member")

				req := httptest.NewRequest(http.MethodPut, "/organizations/"+org.ID.String()+"/members/"+user.ID.String(), strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)
				req = setChiURLParams(req, map[string]string{
					"org_id":  org.ID.String(),
					"user_id": user.ID.String(),
				})

				w := httptest.NewRecorder()
				handler := NewMemberHandler(q).Put()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode == http.StatusOK {
					t.Errorf("expected error status, got %d", resp.StatusCode)
				}
			})
		}
	})

	t.Run("invalid URL params", func(t *testing.T) {
		tests := []struct {
			testName string
			orgID    string
			userID   string
		}{
			// 異常系 - invalid org_id
			{testName: "non-uuid org_id", orgID: "bad-org", userID: uuid.New().String()},
			// 異常系 - invalid user_id
			{testName: "non-uuid user_id", orgID: uuid.New().String(), userID: "bad-user"},
			// 空文字
			{testName: "empty org_id", orgID: "", userID: uuid.New().String()},
			{testName: "empty user_id", orgID: uuid.New().String(), userID: ""},
			// 特殊文字
			{testName: "sql injection in org_id", orgID: "'; DROP TABLE organizations;--", userID: uuid.New().String()},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				body := `{"role":"admin"}`
				req := httptest.NewRequest(http.MethodPut, "/organizations/x/members/y", strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)
				req = setChiURLParams(req, map[string]string{
					"org_id":  tt.orgID,
					"user_id": tt.userID,
				})

				w := httptest.NewRecorder()
				handler := NewMemberHandler(q).Put()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode == http.StatusOK {
					t.Errorf("expected error status for invalid params, got %d", resp.StatusCode)
				}
			})
		}
	})

	t.Run("database error - non-existent member", func(t *testing.T) {
		// 異常系 - updating a member that does not exist
		q, org, _ := setupTest(t)
		nonExistentUserID := uuid.New()

		body := `{"role":"admin"}`
		req := httptest.NewRequest(http.MethodPut, "/organizations/"+org.ID.String()+"/members/"+nonExistentUserID.String(), strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)
		req = setChiURLParams(req, map[string]string{
			"org_id":  org.ID.String(),
			"user_id": nonExistentUserID.String(),
		})

		w := httptest.NewRecorder()
		handler := NewMemberHandler(q).Put()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode == http.StatusOK {
			t.Errorf("expected error status for non-existent member, got %d", resp.StatusCode)
		}
	})
}

// --- Delete Tests ---

func TestDeleteHandler(t *testing.T) {
	t.Run("204 No Content", func(t *testing.T) {
		tests := []struct {
			testName string
			role     string
		}{
			// 正常系
			{testName: "delete member with role member", role: "member"},
			// 境界値 - delete members of different roles
			{testName: "delete member with role admin", role: "admin"},
			{testName: "delete member with role viewer", role: "viewer"},
			{testName: "delete member with role owner", role: "owner"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q, org, user := setupTest(t)
				addMember(t, q, org.ID, user.ID, tt.role)

				req := httptest.NewRequest(http.MethodDelete, "/organizations/"+org.ID.String()+"/members/"+user.ID.String(), nil)
				testutil.SetAuthHeader(req)
				req = setChiURLParams(req, map[string]string{
					"org_id":  org.ID.String(),
					"user_id": user.ID.String(),
				})

				w := httptest.NewRecorder()
				handler := NewMemberHandler(q).Delete()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(http.StatusNoContent, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				// Verify member was actually removed
				members, err := q.ListOrganizationMembers(context.Background(), org.ID)
				if err != nil {
					t.Fatalf("failed to list members: %v", err)
				}
				if len(members) != 0 {
					t.Errorf("expected 0 members after delete, got %d", len(members))
				}
			})
		}
	})

	t.Run("204 No Content - delete non-existent member is idempotent", func(t *testing.T) {
		// 境界値 - deleting a member that does not exist should still return 204
		q, org, _ := setupTest(t)
		nonExistentUserID := uuid.New()

		req := httptest.NewRequest(http.MethodDelete, "/organizations/"+org.ID.String()+"/members/"+nonExistentUserID.String(), nil)
		testutil.SetAuthHeader(req)
		req = setChiURLParams(req, map[string]string{
			"org_id":  org.ID.String(),
			"user_id": nonExistentUserID.String(),
		})

		w := httptest.NewRecorder()
		handler := NewMemberHandler(q).Delete()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		// RemoveOrganizationMember uses :exec which doesn't error on no rows
		if diff := cmp.Diff(http.StatusNoContent, resp.StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("invalid URL params", func(t *testing.T) {
		tests := []struct {
			testName string
			orgID    string
			userID   string
		}{
			// 異常系
			{testName: "non-uuid org_id", orgID: "invalid", userID: uuid.New().String()},
			{testName: "non-uuid user_id", orgID: uuid.New().String(), userID: "invalid"},
			// 空文字
			{testName: "empty org_id", orgID: "", userID: uuid.New().String()},
			{testName: "empty user_id", orgID: uuid.New().String(), userID: ""},
			// 特殊文字
			{testName: "special chars in org_id", orgID: "abc<>def", userID: uuid.New().String()},
			{testName: "emoji in user_id", orgID: uuid.New().String(), userID: "🔥🔥🔥"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				req := httptest.NewRequest(http.MethodDelete, "/organizations/x/members/y", nil)
				testutil.SetAuthHeader(req)
				req = setChiURLParams(req, map[string]string{
					"org_id":  tt.orgID,
					"user_id": tt.userID,
				})

				w := httptest.NewRecorder()
				handler := NewMemberHandler(q).Delete()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode == http.StatusNoContent {
					t.Errorf("expected error status for invalid params, got %d", resp.StatusCode)
				}
			})
		}
	})

	t.Run("verify delete then list shows member removed", func(t *testing.T) {
		// 正常系 - integration: add, delete, then list
		q, org, user := setupTest(t)
		user2 := createUser(t, q, "delete-integration")
		addMember(t, q, org.ID, user.ID, "owner")
		addMember(t, q, org.ID, user2.ID, "member")

		// Delete user2
		req := httptest.NewRequest(http.MethodDelete, "/organizations/"+org.ID.String()+"/members/"+user2.ID.String(), nil)
		testutil.SetAuthHeader(req)
		req = setChiURLParams(req, map[string]string{
			"org_id":  org.ID.String(),
			"user_id": user2.ID.String(),
		})

		w := httptest.NewRecorder()
		handler := NewMemberHandler(q).Delete()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if diff := cmp.Diff(http.StatusNoContent, resp.StatusCode); diff != "" {
			t.Errorf("delete status code mismatch (-want +got):\n%s", diff)
		}

		// List to verify
		listReq := httptest.NewRequest(http.MethodGet, "/organizations/"+org.ID.String()+"/members", nil)
		testutil.SetAuthHeader(listReq)
		listReq = setChiURLParams(listReq, map[string]string{"org_id": org.ID.String()})

		listW := httptest.NewRecorder()
		listHandler := NewMemberHandler(q).List()
		listHandler.ServeHTTP(listW, listReq)

		listResp := listW.Result()
		if diff := cmp.Diff(http.StatusOK, listResp.StatusCode); diff != "" {
			t.Errorf("list status code mismatch (-want +got):\n%s", diff)
		}

		var body []memberResponse
		if err := json.NewDecoder(listResp.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if diff := cmp.Diff(1, len(body)); diff != "" {
			t.Errorf("member count mismatch after delete (-want +got):\n%s", diff)
		}

		if len(body) > 0 {
			if diff := cmp.Diff(user.ID.String(), body[0].UserID); diff != "" {
				t.Errorf("remaining member should be the owner (-want +got):\n%s", diff)
			}
		}
	})
}
