package consulting

import (
	domain "api/src/domain/consulting"
	"api/src/infra/rds/consulting_repository"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"utils/db/db"
	"utils/testutil"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

// setupHandler creates all three repositories and the handler from a test transaction.
func setupHandler(t *testing.T) (db.Querier, *ConsultingHandler) {
	t.Helper()
	q := testutil.SetupTestTx(t)
	sessionRepo := consulting_repository.NewSessionRepository(q)
	messageRepo := consulting_repository.NewMessageRepository(q)
	industryRepo := consulting_repository.NewIndustryConfigRepository(q)
	handler := NewConsultingHandler(sessionRepo, messageRepo, industryRepo, nil)
	return q, handler
}

// seedOrg creates an organization and returns its ID.
func seedOrg(t *testing.T, q db.Querier) uuid.UUID {
	t.Helper()
	org, err := q.CreateOrganization(context.Background(), db.CreateOrganizationParams{
		Name: "Test Org",
		Slug: "test-org-" + uuid.New().String()[:8],
		Plan: "free",
	})
	if err != nil {
		t.Fatalf("failed to seed organization: %v", err)
	}
	return org.ID
}

// seedSession creates a consulting session and returns its ID.
func seedSession(t *testing.T, q db.Querier, orgID uuid.UUID, title string) uuid.UUID {
	t.Helper()
	session, err := q.CreateConsultingSession(context.Background(), db.CreateConsultingSessionParams{
		OrganizationID: orgID,
		Title:          title,
	})
	if err != nil {
		t.Fatalf("failed to seed session: %v", err)
	}
	return session.ID
}

// seedIndustryConfig creates an industry config and returns its ID.
func seedIndustryConfig(t *testing.T, q db.Querier) uuid.UUID {
	t.Helper()
	ic, err := q.CreateIndustryConfig(context.Background(), db.CreateIndustryConfigParams{
		Slug:            "test-industry-" + uuid.New().String()[:8],
		Name:            "Test Industry",
		Description:     sql.NullString{String: "A test industry", Valid: true},
		KnowledgeBase:   json.RawMessage(`{}`),
		ComplianceRules: json.RawMessage(`{}`),
	})
	if err != nil {
		t.Fatalf("failed to seed industry config: %v", err)
	}
	return ic.ID
}

// --- PostSession Tests ---

func TestPostSession(t *testing.T) {
	t.Run("201 Created", func(t *testing.T) {
		type expected struct {
			statusCode int
			title      string
			status     string
		}

		tests := []struct {
			testName string
			setup    func(t *testing.T, q db.Querier) map[string]any
			expected expected
		}{
			// 正常系
			{
				testName: "create session with valid org_id and title",
				setup: func(t *testing.T, q db.Querier) map[string]any {
					orgID := seedOrg(t, q)
					return map[string]any{"org_id": orgID.String(), "title": "My Session"}
				},
				expected: expected{statusCode: http.StatusCreated, title: "My Session", status: "active"},
			},
			{
				testName: "create session with optional industry_config_id as empty",
				setup: func(t *testing.T, q db.Querier) map[string]any {
					orgID := seedOrg(t, q)
					return map[string]any{"org_id": orgID.String(), "title": "Session No Industry"}
				},
				expected: expected{statusCode: http.StatusCreated, title: "Session No Industry", status: "active"},
			},
			// 正常系 - with valid industry_config_id
			{
				testName: "create session with valid industry_config_id",
				setup: func(t *testing.T, q db.Querier) map[string]any {
					orgID := seedOrg(t, q)
					icID := seedIndustryConfig(t, q)
					return map[string]any{"org_id": orgID.String(), "title": "Industry Session", "industry_config_id": icID.String()}
				},
				expected: expected{statusCode: http.StatusCreated, title: "Industry Session", status: "active"},
			},
			// 特殊文字
			{
				testName: "create session with Japanese title",
				setup: func(t *testing.T, q db.Querier) map[string]any {
					orgID := seedOrg(t, q)
					return map[string]any{"org_id": orgID.String(), "title": "コンサルティングセッション"}
				},
				expected: expected{statusCode: http.StatusCreated, title: "コンサルティングセッション", status: "active"},
			},
			{
				testName: "create session with emoji title",
				setup: func(t *testing.T, q db.Querier) map[string]any {
					orgID := seedOrg(t, q)
					return map[string]any{"org_id": orgID.String(), "title": "Session 🚀💼"}
				},
				expected: expected{statusCode: http.StatusCreated, title: "Session 🚀💼", status: "active"},
			},
			// 境界値
			{
				testName: "create session with min length title (1 char)",
				setup: func(t *testing.T, q db.Querier) map[string]any {
					orgID := seedOrg(t, q)
					return map[string]any{"org_id": orgID.String(), "title": "A"}
				},
				expected: expected{statusCode: http.StatusCreated, title: "A", status: "active"},
			},
			{
				testName: "create session with max length title (200 chars)",
				setup: func(t *testing.T, q db.Querier) map[string]any {
					orgID := seedOrg(t, q)
					return map[string]any{"org_id": orgID.String(), "title": strings.Repeat("a", 200)}
				},
				expected: expected{statusCode: http.StatusCreated, title: strings.Repeat("a", 200), status: "active"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q, handler := setupHandler(t)
				body := tt.setup(t, q)

				jsonBody, err := json.Marshal(body)
				if err != nil {
					t.Fatalf("failed to marshal request: %v", err)
				}

				req := httptest.NewRequest(http.MethodPost, "/consulting/sessions", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				handler.PostSession().ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var result map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if _, ok := result["id"]; !ok {
					t.Error("response should contain 'id' field")
				}
				if diff := cmp.Diff(tt.expected.title, result["title"]); diff != "" {
					t.Errorf("title mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expected.status, result["status"]); diff != "" {
					t.Errorf("status mismatch (-want +got):\n%s", diff)
				}
				if _, ok := result["org_id"]; !ok {
					t.Error("response should contain 'org_id' field")
				}
			})
		}
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		tests := []struct {
			testName string
			setup    func(t *testing.T, q db.Querier) map[string]any
		}{
			// 異常系
			{
				testName: "missing org_id",
				setup: func(t *testing.T, q db.Querier) map[string]any {
					return map[string]any{"title": "My Session"}
				},
			},
			{
				testName: "missing title",
				setup: func(t *testing.T, q db.Querier) map[string]any {
					orgID := seedOrg(t, q)
					return map[string]any{"org_id": orgID.String()}
				},
			},
			{
				testName: "invalid org_id format",
				setup: func(t *testing.T, q db.Querier) map[string]any {
					return map[string]any{"org_id": "not-a-uuid", "title": "My Session"}
				},
			},
			{
				testName: "invalid industry_config_id format",
				setup: func(t *testing.T, q db.Querier) map[string]any {
					orgID := seedOrg(t, q)
					return map[string]any{"org_id": orgID.String(), "title": "My Session", "industry_config_id": "not-a-uuid"}
				},
			},
			// 空文字
			{
				testName: "empty title",
				setup: func(t *testing.T, q db.Querier) map[string]any {
					orgID := seedOrg(t, q)
					return map[string]any{"org_id": orgID.String(), "title": ""}
				},
			},
			{
				testName: "empty org_id",
				setup: func(t *testing.T, q db.Querier) map[string]any {
					return map[string]any{"org_id": "", "title": "My Session"}
				},
			},
			// Note: whitespace-only title "   " passes validation (min=1 checks length, not trimmed content).
			// This is expected behavior given the current request validation rules.
			// 境界値
			{
				testName: "title exceeds max length (201 chars)",
				setup: func(t *testing.T, q db.Querier) map[string]any {
					orgID := seedOrg(t, q)
					return map[string]any{"org_id": orgID.String(), "title": strings.Repeat("a", 201)}
				},
			},
			// Null/Nil
			{
				testName: "null org_id",
				setup: func(t *testing.T, q db.Querier) map[string]any {
					return map[string]any{"org_id": nil, "title": "My Session"}
				},
			},
			{
				testName: "null title",
				setup: func(t *testing.T, q db.Querier) map[string]any {
					orgID := seedOrg(t, q)
					return map[string]any{"org_id": orgID.String(), "title": nil}
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q, handler := setupHandler(t)
				body := tt.setup(t, q)

				jsonBody, err := json.Marshal(body)
				if err != nil {
					t.Fatalf("failed to marshal request: %v", err)
				}

				req := httptest.NewRequest(http.MethodPost, "/consulting/sessions", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				handler.PostSession().ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode == http.StatusCreated {
					t.Errorf("expected non-201 status, got %d", resp.StatusCode)
				}
			})
		}
	})

	t.Run("500 or error - non-existent org_id FK violation", func(t *testing.T) {
		tests := []struct {
			testName string
			setup    func(t *testing.T, q db.Querier) map[string]any
		}{
			// 異常系 - valid UUID but org doesn't exist (FK constraint)
			{
				testName: "non-existent org_id causes create error",
				setup: func(t *testing.T, q db.Querier) map[string]any {
					return map[string]any{
						"org_id": uuid.New().String(),
						"title":  "Orphan Session",
					}
				},
			},
			// 異常系 - valid UUID for industry_config_id but doesn't exist (FK constraint)
			{
				testName: "non-existent industry_config_id causes create error",
				setup: func(t *testing.T, q db.Querier) map[string]any {
					orgID := seedOrg(t, q)
					return map[string]any{
						"org_id":             orgID.String(),
						"title":              "Bad Industry Session",
						"industry_config_id": uuid.New().String(),
					}
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q, handler := setupHandler(t)
				body := tt.setup(t, q)

				jsonBody, err := json.Marshal(body)
				if err != nil {
					t.Fatalf("failed to marshal request: %v", err)
				}

				req := httptest.NewRequest(http.MethodPost, "/consulting/sessions", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				handler.PostSession().ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode == http.StatusCreated {
					t.Errorf("expected non-201 status, got %d", resp.StatusCode)
				}
			})
		}
	})

	t.Run("400 Invalid JSON", func(t *testing.T) {
		_, handler := setupHandler(t)

		req := httptest.NewRequest(http.MethodPost, "/consulting/sessions", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)
		w := httptest.NewRecorder()

		handler.PostSession().ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", resp.StatusCode)
		}
	})
}

// --- GetSession Tests ---

func TestGetSession(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		type expected struct {
			statusCode int
			title      string
			status     string
		}

		tests := []struct {
			testName string
			setup    func(t *testing.T, q db.Querier) string
			expected expected
		}{
			// 正常系
			{
				testName: "get existing session",
				setup: func(t *testing.T, q db.Querier) string {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Test Session")
					return sessionID.String()
				},
				expected: expected{statusCode: http.StatusOK, title: "Test Session", status: "active"},
			},
			// 特殊文字
			{
				testName: "get session with Japanese title",
				setup: func(t *testing.T, q db.Querier) string {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "日本語セッション")
					return sessionID.String()
				},
				expected: expected{statusCode: http.StatusOK, title: "日本語セッション", status: "active"},
			},
			// 特殊文字
			{
				testName: "get session with emoji title",
				setup: func(t *testing.T, q db.Querier) string {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Session 📊🤖")
					return sessionID.String()
				},
				expected: expected{statusCode: http.StatusOK, title: "Session 📊🤖", status: "active"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q, handler := setupHandler(t)
				sessionID := tt.setup(t, q)

				req := httptest.NewRequest(http.MethodGet, "/consulting/sessions/"+sessionID, nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", sessionID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				handler.GetSession().ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var result map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(sessionID, result["id"]); diff != "" {
					t.Errorf("id mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expected.title, result["title"]); diff != "" {
					t.Errorf("title mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expected.status, result["status"]); diff != "" {
					t.Errorf("status mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("404 Not Found", func(t *testing.T) {
		tests := []struct {
			testName  string
			sessionID string
		}{
			// 異常系
			{
				testName:  "non-existent session",
				sessionID: "00000000-0000-0000-0000-000000000000",
			},
			{
				testName:  "random non-existent UUID",
				sessionID: uuid.New().String(),
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				_, handler := setupHandler(t)

				req := httptest.NewRequest(http.MethodGet, "/consulting/sessions/"+tt.sessionID, nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", tt.sessionID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				handler.GetSession().ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(http.StatusNotFound, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		tests := []struct {
			testName  string
			sessionID string
		}{
			// 異常系
			{
				testName:  "invalid UUID format",
				sessionID: "not-a-uuid",
			},
			// 空文字
			{
				testName:  "empty ID",
				sessionID: "",
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				_, handler := setupHandler(t)

				req := httptest.NewRequest(http.MethodGet, "/consulting/sessions/"+tt.sessionID, nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", tt.sessionID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				handler.GetSession().ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
					t.Errorf("expected error status, got %d", resp.StatusCode)
				}
			})
		}
	})
}

// --- ListSessions Tests ---

func TestListSessions(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		type expected struct {
			statusCode int
			count      int
		}

		tests := []struct {
			testName    string
			setup       func(t *testing.T, q db.Querier) uuid.UUID
			queryParams map[string]string
			expected    expected
		}{
			// 正常系
			{
				testName: "list sessions for org with sessions",
				setup: func(t *testing.T, q db.Querier) uuid.UUID {
					orgID := seedOrg(t, q)
					seedSession(t, q, orgID, "Session 1")
					seedSession(t, q, orgID, "Session 2")
					seedSession(t, q, orgID, "Session 3")
					return orgID
				},
				queryParams: map[string]string{},
				expected:    expected{statusCode: http.StatusOK, count: 3},
			},
			// 境界値
			{
				testName: "list sessions for org with no sessions",
				setup: func(t *testing.T, q db.Querier) uuid.UUID {
					return seedOrg(t, q)
				},
				queryParams: map[string]string{},
				expected:    expected{statusCode: http.StatusOK, count: 0},
			},
			{
				testName: "list with limit=1",
				setup: func(t *testing.T, q db.Querier) uuid.UUID {
					orgID := seedOrg(t, q)
					seedSession(t, q, orgID, "Session 1")
					seedSession(t, q, orgID, "Session 2")
					return orgID
				},
				queryParams: map[string]string{"limit": "1"},
				expected:    expected{statusCode: http.StatusOK, count: 1},
			},
			{
				testName: "list with offset=1",
				setup: func(t *testing.T, q db.Querier) uuid.UUID {
					orgID := seedOrg(t, q)
					seedSession(t, q, orgID, "Session 1")
					seedSession(t, q, orgID, "Session 2")
					seedSession(t, q, orgID, "Session 3")
					return orgID
				},
				queryParams: map[string]string{"offset": "1"},
				expected:    expected{statusCode: http.StatusOK, count: 2},
			},
			{
				testName: "list with limit=1 offset=1",
				setup: func(t *testing.T, q db.Querier) uuid.UUID {
					orgID := seedOrg(t, q)
					seedSession(t, q, orgID, "Session 1")
					seedSession(t, q, orgID, "Session 2")
					seedSession(t, q, orgID, "Session 3")
					return orgID
				},
				queryParams: map[string]string{"limit": "1", "offset": "1"},
				expected:    expected{statusCode: http.StatusOK, count: 1},
			},
			// 境界値 - offset beyond data
			{
				testName: "list with offset beyond data",
				setup: func(t *testing.T, q db.Querier) uuid.UUID {
					orgID := seedOrg(t, q)
					seedSession(t, q, orgID, "Session 1")
					return orgID
				},
				queryParams: map[string]string{"offset": "100"},
				expected:    expected{statusCode: http.StatusOK, count: 0},
			},
			// 境界値 - invalid limit/offset are ignored (defaults used)
			{
				testName: "list with negative limit uses default",
				setup: func(t *testing.T, q db.Querier) uuid.UUID {
					orgID := seedOrg(t, q)
					seedSession(t, q, orgID, "Session 1")
					return orgID
				},
				queryParams: map[string]string{"limit": "-1"},
				expected:    expected{statusCode: http.StatusOK, count: 1},
			},
			{
				testName: "list with non-numeric limit uses default",
				setup: func(t *testing.T, q db.Querier) uuid.UUID {
					orgID := seedOrg(t, q)
					seedSession(t, q, orgID, "Session 1")
					return orgID
				},
				queryParams: map[string]string{"limit": "abc"},
				expected:    expected{statusCode: http.StatusOK, count: 1},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q, handler := setupHandler(t)
				orgID := tt.setup(t, q)

				target := "/consulting/sessions?org_id=" + orgID.String()
				for k, v := range tt.queryParams {
					target += "&" + k + "=" + v
				}

				req := httptest.NewRequest(http.MethodGet, target, nil)
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				handler.ListSessions().ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var result []map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(tt.expected.count, len(result)); diff != "" {
					t.Errorf("result count mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("error cases", func(t *testing.T) {
		tests := []struct {
			testName string
			orgID    string
		}{
			// 異常系
			{
				testName: "invalid org_id format",
				orgID:    "not-a-uuid",
			},
			// 空文字
			{
				testName: "empty org_id",
				orgID:    "",
			},
			// Null/Nil - missing org_id query param
			{
				testName: "missing org_id",
				orgID:    "",
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				_, handler := setupHandler(t)

				target := "/consulting/sessions"
				if tt.orgID != "" {
					target += "?org_id=" + tt.orgID
				}

				req := httptest.NewRequest(http.MethodGet, target, nil)
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				handler.ListSessions().ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode == http.StatusOK {
					t.Errorf("expected error status, got %d", resp.StatusCode)
				}
			})
		}
	})
}

// --- PostMessage Tests ---

func TestPostMessage(t *testing.T) {
	t.Run("201 Created", func(t *testing.T) {
		type expected struct {
			statusCode int
			role       string
			content    string
		}

		tests := []struct {
			testName string
			setup    func(t *testing.T, q db.Querier) (string, map[string]any)
			expected expected
		}{
			// 正常系
			{
				testName: "create user message",
				setup: func(t *testing.T, q db.Querier) (string, map[string]any) {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Test Session")
					return sessionID.String(), map[string]any{
						"role":    "user",
						"content": "Hello, I need consulting help",
					}
				},
				expected: expected{statusCode: http.StatusCreated, role: "user", content: "Hello, I need consulting help"},
			},
			{
				testName: "create assistant message",
				setup: func(t *testing.T, q db.Querier) (string, map[string]any) {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Test Session")
					return sessionID.String(), map[string]any{
						"role":    "assistant",
						"content": "I can help with that",
					}
				},
				expected: expected{statusCode: http.StatusCreated, role: "assistant", content: "I can help with that"},
			},
			{
				testName: "create system message",
				setup: func(t *testing.T, q db.Querier) (string, map[string]any) {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Test Session")
					return sessionID.String(), map[string]any{
						"role":    "system",
						"content": "You are a helpful assistant.",
					}
				},
				expected: expected{statusCode: http.StatusCreated, role: "system", content: "You are a helpful assistant."},
			},
			{
				testName: "create message with citations",
				setup: func(t *testing.T, q db.Querier) (string, map[string]any) {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Test Session")
					return sessionID.String(), map[string]any{
						"role":      "assistant",
						"content":   "Based on the research...",
						"citations": []map[string]string{{"source": "doc1.pdf", "page": "5"}},
					}
				},
				expected: expected{statusCode: http.StatusCreated, role: "assistant", content: "Based on the research..."},
			},
			{
				testName: "create message with actions_taken",
				setup: func(t *testing.T, q db.Querier) (string, map[string]any) {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Test Session")
					return sessionID.String(), map[string]any{
						"role":          "assistant",
						"content":       "I performed the analysis",
						"actions_taken": []string{"analyzed_data", "generated_report"},
					}
				},
				expected: expected{statusCode: http.StatusCreated, role: "assistant", content: "I performed the analysis"},
			},
			// 特殊文字
			{
				testName: "create message with Japanese content",
				setup: func(t *testing.T, q db.Querier) (string, map[string]any) {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Test Session")
					return sessionID.String(), map[string]any{
						"role":    "user",
						"content": "コンサルティングについて質問があります",
					}
				},
				expected: expected{statusCode: http.StatusCreated, role: "user", content: "コンサルティングについて質問があります"},
			},
			{
				testName: "create message with emoji content",
				setup: func(t *testing.T, q db.Querier) (string, map[string]any) {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Test Session")
					return sessionID.String(), map[string]any{
						"role":    "user",
						"content": "Great work! 🎉👍",
					}
				},
				expected: expected{statusCode: http.StatusCreated, role: "user", content: "Great work! 🎉👍"},
			},
			{
				testName: "create message with SQL injection attempt in content",
				setup: func(t *testing.T, q db.Querier) (string, map[string]any) {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Test Session")
					return sessionID.String(), map[string]any{
						"role":    "user",
						"content": "Robert'); DROP TABLE consulting_messages;--",
					}
				},
				expected: expected{statusCode: http.StatusCreated, role: "user", content: "Robert&#39;); DROP TABLE consulting_messages;--"},
			},
			// 境界値
			{
				testName: "create message with min length content (1 char)",
				setup: func(t *testing.T, q db.Querier) (string, map[string]any) {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Test Session")
					return sessionID.String(), map[string]any{
						"role":    "user",
						"content": "X",
					}
				},
				expected: expected{statusCode: http.StatusCreated, role: "user", content: "X"},
			},
			// Null/Nil - null optional fields
			{
				testName: "create message with null citations and actions_taken",
				setup: func(t *testing.T, q db.Querier) (string, map[string]any) {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Test Session")
					return sessionID.String(), map[string]any{
						"role":          "user",
						"content":       "Hello",
						"citations":     nil,
						"actions_taken": nil,
					}
				},
				expected: expected{statusCode: http.StatusCreated, role: "user", content: "Hello"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q, handler := setupHandler(t)
				sessionID, body := tt.setup(t, q)

				jsonBody, err := json.Marshal(body)
				if err != nil {
					t.Fatalf("failed to marshal request: %v", err)
				}

				req := httptest.NewRequest(http.MethodPost, "/consulting/sessions/"+sessionID+"/messages", bytes.NewReader(jsonBody))
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("session_id", sessionID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				handler.PostMessage().ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var result map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if _, ok := result["id"]; !ok {
					t.Error("response should contain 'id' field")
				}
				if diff := cmp.Diff(sessionID, result["session_id"]); diff != "" {
					t.Errorf("session_id mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expected.role, result["role"]); diff != "" {
					t.Errorf("role mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expected.content, result["content"]); diff != "" {
					t.Errorf("content mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		tests := []struct {
			testName string
			setup    func(t *testing.T, q db.Querier) (string, map[string]any)
		}{
			// 異常系
			{
				testName: "invalid role",
				setup: func(t *testing.T, q db.Querier) (string, map[string]any) {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Test Session")
					return sessionID.String(), map[string]any{
						"role":    "invalid_role",
						"content": "Hello",
					}
				},
			},
			{
				testName: "missing role",
				setup: func(t *testing.T, q db.Querier) (string, map[string]any) {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Test Session")
					return sessionID.String(), map[string]any{
						"content": "Hello",
					}
				},
			},
			{
				testName: "missing content",
				setup: func(t *testing.T, q db.Querier) (string, map[string]any) {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Test Session")
					return sessionID.String(), map[string]any{
						"role": "user",
					}
				},
			},
			// 空文字
			{
				testName: "empty role",
				setup: func(t *testing.T, q db.Querier) (string, map[string]any) {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Test Session")
					return sessionID.String(), map[string]any{
						"role":    "",
						"content": "Hello",
					}
				},
			},
			{
				testName: "empty content",
				setup: func(t *testing.T, q db.Querier) (string, map[string]any) {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Test Session")
					return sessionID.String(), map[string]any{
						"role":    "user",
						"content": "",
					}
				},
			},
			// Null/Nil
			{
				testName: "null role",
				setup: func(t *testing.T, q db.Querier) (string, map[string]any) {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Test Session")
					return sessionID.String(), map[string]any{
						"role":    nil,
						"content": "Hello",
					}
				},
			},
			{
				testName: "null content",
				setup: func(t *testing.T, q db.Querier) (string, map[string]any) {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Test Session")
					return sessionID.String(), map[string]any{
						"role":    "user",
						"content": nil,
					}
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q, handler := setupHandler(t)
				sessionID, body := tt.setup(t, q)

				jsonBody, err := json.Marshal(body)
				if err != nil {
					t.Fatalf("failed to marshal request: %v", err)
				}

				req := httptest.NewRequest(http.MethodPost, "/consulting/sessions/"+sessionID+"/messages", bytes.NewReader(jsonBody))
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("session_id", sessionID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				handler.PostMessage().ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode == http.StatusCreated {
					t.Errorf("expected non-201 status, got %d", resp.StatusCode)
				}
			})
		}
	})

	t.Run("invalid session_id", func(t *testing.T) {
		tests := []struct {
			testName  string
			sessionID string
		}{
			// 異常系
			{
				testName:  "invalid UUID format for session_id",
				sessionID: "not-a-uuid",
			},
			// 空文字
			{
				testName:  "empty session_id",
				sessionID: "",
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				_, handler := setupHandler(t)

				body := map[string]any{"role": "user", "content": "Hello"}
				jsonBody, _ := json.Marshal(body)

				req := httptest.NewRequest(http.MethodPost, "/consulting/sessions/"+tt.sessionID+"/messages", bytes.NewReader(jsonBody))
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("session_id", tt.sessionID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				handler.PostMessage().ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode == http.StatusCreated {
					t.Errorf("expected error status, got %d", resp.StatusCode)
				}
			})
		}
	})

	t.Run("error - non-existent session FK violation", func(t *testing.T) {
		// 異常系 - valid UUID but session doesn't exist (FK constraint)
		_, handler := setupHandler(t)
		nonExistentSessionID := uuid.New().String()

		body := map[string]any{"role": "user", "content": "Hello"}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/consulting/sessions/"+nonExistentSessionID+"/messages", bytes.NewReader(jsonBody))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("session_id", nonExistentSessionID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)
		w := httptest.NewRecorder()

		handler.PostMessage().ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode == http.StatusCreated {
			t.Errorf("expected error status for non-existent session, got %d", resp.StatusCode)
		}
	})

	t.Run("400 Invalid JSON", func(t *testing.T) {
		q, handler := setupHandler(t)
		orgID := seedOrg(t, q)
		sessionID := seedSession(t, q, orgID, "Test Session")

		req := httptest.NewRequest(http.MethodPost, "/consulting/sessions/"+sessionID.String()+"/messages", bytes.NewReader([]byte("invalid json")))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("session_id", sessionID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)
		w := httptest.NewRecorder()

		handler.PostMessage().ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", resp.StatusCode)
		}
	})
}

// --- ListMessages Tests ---

func TestListMessages(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		type expected struct {
			statusCode int
			count      int
		}

		tests := []struct {
			testName string
			setup    func(t *testing.T, q db.Querier) string
			expected expected
		}{
			// 正常系
			{
				testName: "list messages for session with messages",
				setup: func(t *testing.T, q db.Querier) string {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Test Session")
					// Seed messages via the handler's repository path
					q.CreateConsultingMessage(context.Background(), db.CreateConsultingMessageParams{
						SessionID: sessionID,
						Role:      "user",
						Content:   "Hello",
					})
					q.CreateConsultingMessage(context.Background(), db.CreateConsultingMessageParams{
						SessionID: sessionID,
						Role:      "assistant",
						Content:   "Hi there!",
					})
					return sessionID.String()
				},
				expected: expected{statusCode: http.StatusOK, count: 2},
			},
			// 境界値
			{
				testName: "list messages for session with no messages",
				setup: func(t *testing.T, q db.Querier) string {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Empty Session")
					return sessionID.String()
				},
				expected: expected{statusCode: http.StatusOK, count: 0},
			},
			// 正常系 - single message
			{
				testName: "list messages for session with one message",
				setup: func(t *testing.T, q db.Querier) string {
					orgID := seedOrg(t, q)
					sessionID := seedSession(t, q, orgID, "Test Session")
					q.CreateConsultingMessage(context.Background(), db.CreateConsultingMessageParams{
						SessionID: sessionID,
						Role:      "user",
						Content:   "Only message",
					})
					return sessionID.String()
				},
				expected: expected{statusCode: http.StatusOK, count: 1},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q, handler := setupHandler(t)
				sessionID := tt.setup(t, q)

				req := httptest.NewRequest(http.MethodGet, "/consulting/sessions/"+sessionID+"/messages", nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("session_id", sessionID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				handler.ListMessages().ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var result []map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(tt.expected.count, len(result)); diff != "" {
					t.Errorf("result count mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("invalid session_id", func(t *testing.T) {
		tests := []struct {
			testName  string
			sessionID string
		}{
			// 異常系
			{
				testName:  "invalid UUID format",
				sessionID: "not-a-uuid",
			},
			// 空文字
			{
				testName:  "empty session_id",
				sessionID: "",
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				_, handler := setupHandler(t)

				req := httptest.NewRequest(http.MethodGet, "/consulting/sessions/"+tt.sessionID+"/messages", nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("session_id", tt.sessionID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				handler.ListMessages().ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode == http.StatusOK {
					t.Errorf("expected error status, got %d", resp.StatusCode)
				}
			})
		}
	})
}

// --- Integration: PostSession + GetSession round-trip ---

func TestPostSessionAndGetSession(t *testing.T) {
	t.Run("create and then retrieve session", func(t *testing.T) {
		q, handler := setupHandler(t)
		orgID := seedOrg(t, q)

		// Create session
		body := map[string]any{"org_id": orgID.String(), "title": "Round-trip Test"}
		jsonBody, _ := json.Marshal(body)

		createReq := httptest.NewRequest(http.MethodPost, "/consulting/sessions", bytes.NewReader(jsonBody))
		createReq.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(createReq)
		createW := httptest.NewRecorder()

		handler.PostSession().ServeHTTP(createW, createReq)

		if diff := cmp.Diff(http.StatusCreated, createW.Result().StatusCode); diff != "" {
			t.Fatalf("create status mismatch (-want +got):\n%s", diff)
		}

		var createResult map[string]any
		json.NewDecoder(createW.Result().Body).Decode(&createResult)
		sessionID := createResult["id"].(string)

		// Get session
		getReq := httptest.NewRequest(http.MethodGet, "/consulting/sessions/"+sessionID, nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", sessionID)
		getReq = getReq.WithContext(context.WithValue(getReq.Context(), chi.RouteCtxKey, rctx))
		testutil.SetAuthHeader(getReq)
		getW := httptest.NewRecorder()

		handler.GetSession().ServeHTTP(getW, getReq)

		if diff := cmp.Diff(http.StatusOK, getW.Result().StatusCode); diff != "" {
			t.Fatalf("get status mismatch (-want +got):\n%s", diff)
		}

		var getResult map[string]any
		json.NewDecoder(getW.Result().Body).Decode(&getResult)

		if diff := cmp.Diff(sessionID, getResult["id"]); diff != "" {
			t.Errorf("id mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff("Round-trip Test", getResult["title"]); diff != "" {
			t.Errorf("title mismatch (-want +got):\n%s", diff)
		}
	})
}

// --- toSessionResponse Tests ---

func TestToSessionResponse(t *testing.T) {
	type args struct {
		session domain.Session
	}
	type expected struct {
		hasIndustryConfigID bool
		title               string
		status              string
	}

	industryID := uuid.New()

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 - session without industry config
		{
			testName: "session without industry config has nil industry_config_id",
			args: args{
				session: domain.Session{
					ID:               uuid.New(),
					OrgID:            uuid.New(),
					Title:            "Test Session",
					IndustryConfigID: nil,
					Status:           "active",
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				},
			},
			expected: expected{hasIndustryConfigID: false, title: "Test Session", status: "active"},
		},
		// 正常系 - session with industry config
		{
			testName: "session with industry config has non-nil industry_config_id",
			args: args{
				session: domain.Session{
					ID:               uuid.New(),
					OrgID:            uuid.New(),
					Title:            "Industry Session",
					IndustryConfigID: &industryID,
					Status:           "active",
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				},
			},
			expected: expected{hasIndustryConfigID: true, title: "Industry Session", status: "active"},
		},
		// 境界値 - closed session
		{
			testName: "closed session has correct status",
			args: args{
				session: domain.Session{
					ID:               uuid.New(),
					OrgID:            uuid.New(),
					Title:            "Closed Session",
					IndustryConfigID: nil,
					Status:           "closed",
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				},
			},
			expected: expected{hasIndustryConfigID: false, title: "Closed Session", status: "closed"},
		},
		// 特殊文字 - Japanese title
		{
			testName: "session with Japanese title",
			args: args{
				session: domain.Session{
					ID:               uuid.New(),
					OrgID:            uuid.New(),
					Title:            "日本語セッション",
					IndustryConfigID: nil,
					Status:           "active",
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				},
			},
			expected: expected{hasIndustryConfigID: false, title: "日本語セッション", status: "active"},
		},
		// 空文字 - empty title
		{
			testName: "session with empty title",
			args: args{
				session: domain.Session{
					ID:               uuid.New(),
					OrgID:            uuid.New(),
					Title:            "",
					IndustryConfigID: nil,
					Status:           "active",
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				},
			},
			expected: expected{hasIndustryConfigID: false, title: "", status: "active"},
		},
		// Null/Nil - zero value UUID session
		{
			testName: "session with zero UUID",
			args: args{
				session: domain.Session{
					ID:               uuid.UUID{},
					OrgID:            uuid.UUID{},
					Title:            "Zero UUID",
					IndustryConfigID: nil,
					Status:           "active",
					CreatedAt:        time.Time{},
					UpdatedAt:        time.Time{},
				},
			},
			expected: expected{hasIndustryConfigID: false, title: "Zero UUID", status: "active"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := toSessionResponse(tt.args.session)

			if diff := cmp.Diff(tt.expected.title, got.Title); diff != "" {
				t.Errorf("title mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected.status, got.Status); diff != "" {
				t.Errorf("status mismatch (-want +got):\n%s", diff)
			}
			if tt.expected.hasIndustryConfigID {
				if got.IndustryConfigID == nil {
					t.Error("expected non-nil industry_config_id")
				} else if diff := cmp.Diff(industryID.String(), *got.IndustryConfigID); diff != "" {
					t.Errorf("industry_config_id mismatch (-want +got):\n%s", diff)
				}
			} else {
				if got.IndustryConfigID != nil {
					t.Errorf("expected nil industry_config_id, got %v", *got.IndustryConfigID)
				}
			}
			// Verify ID and OrgID are serialized
			if got.ID == "" {
				t.Error("expected non-empty ID")
			}
			if got.OrgID == "" {
				t.Error("expected non-empty OrgID")
			}
		})
	}
}

// --- toMessageResponse Tests ---

func TestToMessageResponse(t *testing.T) {
	type args struct {
		msg domain.Message
	}
	type expected struct {
		role    string
		content string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 - user message
		{
			testName: "user message response",
			args: args{
				msg: domain.Message{
					ID:        uuid.New(),
					SessionID: uuid.New(),
					Role:      "user",
					Content:   "Hello world",
					CreatedAt: time.Now(),
				},
			},
			expected: expected{role: "user", content: "Hello world"},
		},
		// 正常系 - assistant message with citations
		{
			testName: "assistant message with citations",
			args: args{
				msg: domain.Message{
					ID:           uuid.New(),
					SessionID:    uuid.New(),
					Role:         "assistant",
					Content:      "Based on research",
					Citations:    json.RawMessage(`[{"source":"doc.pdf"}]`),
					ActionsTaken: json.RawMessage(`["analyzed"]`),
					CreatedAt:    time.Now(),
				},
			},
			expected: expected{role: "assistant", content: "Based on research"},
		},
		// 特殊文字 - Japanese content
		{
			testName: "message with Japanese content",
			args: args{
				msg: domain.Message{
					ID:        uuid.New(),
					SessionID: uuid.New(),
					Role:      "user",
					Content:   "こんにちは世界",
					CreatedAt: time.Now(),
				},
			},
			expected: expected{role: "user", content: "こんにちは世界"},
		},
		// Null/Nil - nil citations and actions
		{
			testName: "message with nil citations and actions",
			args: args{
				msg: domain.Message{
					ID:           uuid.New(),
					SessionID:    uuid.New(),
					Role:         "assistant",
					Content:      "Response",
					Citations:    nil,
					ActionsTaken: nil,
					CreatedAt:    time.Now(),
				},
			},
			expected: expected{role: "assistant", content: "Response"},
		},
		// 空文字 - empty content
		{
			testName: "message with empty content",
			args: args{
				msg: domain.Message{
					ID:        uuid.New(),
					SessionID: uuid.New(),
					Role:      "system",
					Content:   "",
					CreatedAt: time.Now(),
				},
			},
			expected: expected{role: "system", content: ""},
		},
		// 境界値 - zero UUID
		{
			testName: "message with zero UUIDs",
			args: args{
				msg: domain.Message{
					ID:        uuid.UUID{},
					SessionID: uuid.UUID{},
					Role:      "user",
					Content:   "Zero IDs",
					CreatedAt: time.Time{},
				},
			},
			expected: expected{role: "user", content: "Zero IDs"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := toMessageResponse(tt.args.msg)

			if diff := cmp.Diff(tt.expected.role, got.Role); diff != "" {
				t.Errorf("role mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expected.content, got.Content); diff != "" {
				t.Errorf("content mismatch (-want +got):\n%s", diff)
			}
			if got.ID == "" {
				t.Error("expected non-empty ID")
			}
			if got.SessionID == "" {
				t.Error("expected non-empty SessionID")
			}
		})
	}
}

// --- Integration: PostMessage + ListMessages round-trip ---

func TestPostMessageAndListMessages(t *testing.T) {
	t.Run("create messages and then list them", func(t *testing.T) {
		q, handler := setupHandler(t)
		orgID := seedOrg(t, q)
		sessionID := seedSession(t, q, orgID, "Message Round-trip")

		// Post first message
		msg1 := map[string]any{"role": "user", "content": "First message"}
		jsonMsg1, _ := json.Marshal(msg1)
		req1 := httptest.NewRequest(http.MethodPost, "/consulting/sessions/"+sessionID.String()+"/messages", bytes.NewReader(jsonMsg1))
		rctx1 := chi.NewRouteContext()
		rctx1.URLParams.Add("session_id", sessionID.String())
		req1 = req1.WithContext(context.WithValue(req1.Context(), chi.RouteCtxKey, rctx1))
		req1.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req1)
		w1 := httptest.NewRecorder()
		handler.PostMessage().ServeHTTP(w1, req1)

		if diff := cmp.Diff(http.StatusCreated, w1.Result().StatusCode); diff != "" {
			t.Fatalf("post message 1 status mismatch (-want +got):\n%s", diff)
		}

		// Post second message
		msg2 := map[string]any{"role": "assistant", "content": "Second message"}
		jsonMsg2, _ := json.Marshal(msg2)
		req2 := httptest.NewRequest(http.MethodPost, "/consulting/sessions/"+sessionID.String()+"/messages", bytes.NewReader(jsonMsg2))
		rctx2 := chi.NewRouteContext()
		rctx2.URLParams.Add("session_id", sessionID.String())
		req2 = req2.WithContext(context.WithValue(req2.Context(), chi.RouteCtxKey, rctx2))
		req2.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req2)
		w2 := httptest.NewRecorder()
		handler.PostMessage().ServeHTTP(w2, req2)

		if diff := cmp.Diff(http.StatusCreated, w2.Result().StatusCode); diff != "" {
			t.Fatalf("post message 2 status mismatch (-want +got):\n%s", diff)
		}

		// List messages
		listReq := httptest.NewRequest(http.MethodGet, "/consulting/sessions/"+sessionID.String()+"/messages", nil)
		listCtx := chi.NewRouteContext()
		listCtx.URLParams.Add("session_id", sessionID.String())
		listReq = listReq.WithContext(context.WithValue(listReq.Context(), chi.RouteCtxKey, listCtx))
		testutil.SetAuthHeader(listReq)
		listW := httptest.NewRecorder()
		handler.ListMessages().ServeHTTP(listW, listReq)

		if diff := cmp.Diff(http.StatusOK, listW.Result().StatusCode); diff != "" {
			t.Fatalf("list status mismatch (-want +got):\n%s", diff)
		}

		var messages []map[string]any
		json.NewDecoder(listW.Result().Body).Decode(&messages)

		if diff := cmp.Diff(2, len(messages)); diff != "" {
			t.Errorf("message count mismatch (-want +got):\n%s", diff)
		}

		// Messages should be ordered by created_at ASC
		if len(messages) == 2 {
			if diff := cmp.Diff("user", messages[0]["role"]); diff != "" {
				t.Errorf("first message role mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff("assistant", messages[1]["role"]); diff != "" {
				t.Errorf("second message role mismatch (-want +got):\n%s", diff)
			}
		}
	})
}
