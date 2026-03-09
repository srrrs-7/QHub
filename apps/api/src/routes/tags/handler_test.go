package tags

import (
	"api/src/infra/rds/tag_repository"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
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

// seedOrg creates an organization and returns its ID.
func seedOrg(t *testing.T, q db.Querier) uuid.UUID {
	t.Helper()
	org, err := q.CreateOrganization(context.Background(), db.CreateOrganizationParams{
		Name: "Test Org",
		Slug: "test-org-" + uuid.New().String()[:8],
		Plan: "free",
	})
	if err != nil {
		t.Fatalf("failed to create org: %v", err)
	}
	return org.ID
}

// seedProject creates a project under the given org and returns its ID.
func seedProject(t *testing.T, q db.Querier, orgID uuid.UUID) uuid.UUID {
	t.Helper()
	proj, err := q.CreateProject(context.Background(), db.CreateProjectParams{
		OrganizationID: orgID,
		Name:           "Test Project",
		Slug:           "test-proj-" + uuid.New().String()[:8],
		Description:    sql.NullString{String: "desc", Valid: true},
	})
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}
	return proj.ID
}

// seedPrompt creates a prompt under the given project and returns its ID.
func seedPrompt(t *testing.T, q db.Querier, projectID uuid.UUID) uuid.UUID {
	t.Helper()
	prompt, err := q.CreatePrompt(context.Background(), db.CreatePromptParams{
		ProjectID:   projectID,
		Name:        "Test Prompt",
		Slug:        "test-prompt-" + uuid.New().String()[:8],
		PromptType:  "system",
		Description: sql.NullString{String: "desc", Valid: true},
	})
	if err != nil {
		t.Fatalf("failed to create prompt: %v", err)
	}
	return prompt.ID
}

// seedTag creates a tag under the given org and returns the db.Tag.
func seedTag(t *testing.T, q db.Querier, orgID uuid.UUID, name, color string) db.Tag {
	t.Helper()
	tag, err := q.CreateTag(context.Background(), db.CreateTagParams{
		OrganizationID: orgID,
		Name:           name,
		Color:          color,
	})
	if err != nil {
		t.Fatalf("failed to create tag: %v", err)
	}
	return tag
}

// --- Post Handler Tests ---

func TestPostHandler(t *testing.T) {
	t.Run("201 Created", func(t *testing.T) {
		type args struct {
			body map[string]string
		}
		type expected struct {
			statusCode int
			name       string
			color      string
		}

		tests := []struct {
			testName string
			args     args
			expected expected
		}{
			// 正常系
			{
				testName: "create tag with valid name and color",
				args:     args{body: map[string]string{"name": "bugfix", "color": "#ff0000"}},
				expected: expected{statusCode: http.StatusCreated, name: "bugfix", color: "#ff0000"},
			},
			// 特殊文字
			{
				testName: "create tag with emoji name",
				args:     args{body: map[string]string{"name": "bug 🐛", "color": "#00ff00"}},
				expected: expected{statusCode: http.StatusCreated, name: "bug 🐛", color: "#00ff00"},
			},
			{
				testName: "create tag with Japanese name",
				args:     args{body: map[string]string{"name": "バグ修正", "color": "#0000ff"}},
				expected: expected{statusCode: http.StatusCreated, name: "バグ修正", color: "#0000ff"},
			},
			// 境界値
			{
				testName: "create tag with single char name (min length)",
				args:     args{body: map[string]string{"name": "a", "color": "r"}},
				expected: expected{statusCode: http.StatusCreated, name: "a", color: "r"},
			},
			{
				testName: "create tag with 50 char name (DB max length)",
				args:     args{body: map[string]string{"name": strings.Repeat("a", 50), "color": "red"}},
				expected: expected{statusCode: http.StatusCreated, name: strings.Repeat("a", 50), color: "red"},
			},
			{
				testName: "create tag with 7 char color (DB max length)",
				args:     args{body: map[string]string{"name": "test", "color": "#abcdef"}},
				expected: expected{statusCode: http.StatusCreated, name: "test", color: "#abcdef"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)
				orgID := seedOrg(t, q)

				body := map[string]string{
					"org_id": orgID.String(),
					"name":   tt.args.body["name"],
					"color":  tt.args.body["color"],
				}
				jsonBody, err := json.Marshal(body)
				if err != nil {
					t.Fatalf("failed to marshal body: %v", err)
				}

				req := httptest.NewRequest(http.MethodPost, "/tags", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				repo := tag_repository.NewTagRepository(q)
				handler := NewTagHandler(repo).Post()
				handler.ServeHTTP(w, req)

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
				if diff := cmp.Diff(tt.expected.name, result["name"]); diff != "" {
					t.Errorf("name mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expected.color, result["color"]); diff != "" {
					t.Errorf("color mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(orgID.String(), result["org_id"]); diff != "" {
					t.Errorf("org_id mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		tests := []struct {
			testName       string
			body           string
			expectedStatus int
		}{
			// 異常系
			{
				testName:       "invalid JSON body",
				body:           `{invalid`,
				expectedStatus: http.StatusBadRequest,
			},
			// 空文字
			{
				testName:       "empty name field",
				body:           `{"org_id":"` + uuid.New().String() + `","name":"","color":"red"}`,
				expectedStatus: http.StatusBadRequest,
			},
			// Note: whitespace-only name "   " passes validation (length 3, required passes)
			// since bluemonday does not strip spaces and go-playground/validator treats it as non-empty.
			// Null/Nil
			{
				testName:       "missing name field",
				body:           `{"org_id":"` + uuid.New().String() + `","color":"red"}`,
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName:       "missing color field",
				body:           `{"org_id":"` + uuid.New().String() + `","name":"tag"}`,
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName:       "missing org_id field",
				body:           `{"name":"tag","color":"red"}`,
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName:       "empty body",
				body:           `{}`,
				expectedStatus: http.StatusBadRequest,
			},
			// 異常系 - invalid org_id
			{
				testName:       "non-uuid org_id",
				body:           `{"org_id":"not-a-uuid","name":"tag","color":"red"}`,
				expectedStatus: http.StatusBadRequest,
			},
			// 境界値 - over max length
			{
				testName:       "name over 100 chars",
				body:           `{"org_id":"` + uuid.New().String() + `","name":"` + strings.Repeat("a", 101) + `","color":"red"}`,
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName:       "color over 20 chars",
				body:           `{"org_id":"` + uuid.New().String() + `","name":"tag","color":"` + strings.Repeat("c", 21) + `"}`,
				expectedStatus: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				req := httptest.NewRequest(http.MethodPost, "/tags", strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				repo := tag_repository.NewTagRepository(q)
				handler := NewTagHandler(repo).Post()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expectedStatus, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})
}

// --- List Handler Tests ---

func TestListHandler(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		type expected struct {
			statusCode int
			tagCount   int
		}

		tests := []struct {
			testName string
			setup    func(t *testing.T, q db.Querier) uuid.UUID
			expected expected
		}{
			// 正常系
			{
				testName: "list tags for org with no tags",
				setup: func(t *testing.T, q db.Querier) uuid.UUID {
					return seedOrg(t, q)
				},
				expected: expected{statusCode: http.StatusOK, tagCount: 0},
			},
			{
				testName: "list tags for org with multiple tags",
				setup: func(t *testing.T, q db.Querier) uuid.UUID {
					orgID := seedOrg(t, q)
					seedTag(t, q, orgID, "alpha", "red")
					seedTag(t, q, orgID, "beta", "blue")
					seedTag(t, q, orgID, "gamma", "green")
					return orgID
				},
				expected: expected{statusCode: http.StatusOK, tagCount: 3},
			},
			// 境界値
			{
				testName: "list tags returns only tags for specific org",
				setup: func(t *testing.T, q db.Querier) uuid.UUID {
					orgID1 := seedOrg(t, q)
					orgID2 := seedOrg(t, q)
					seedTag(t, q, orgID1, "org1-tag", "red")
					seedTag(t, q, orgID2, "org2-tag", "blue")
					return orgID1
				},
				expected: expected{statusCode: http.StatusOK, tagCount: 1},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)
				orgID := tt.setup(t, q)

				req := httptest.NewRequest(http.MethodGet, "/tags?org_id="+orgID.String(), nil)
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				repo := tag_repository.NewTagRepository(q)
				handler := NewTagHandler(repo).List()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var result []map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(tt.expected.tagCount, len(result)); diff != "" {
					t.Errorf("tag count mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("error cases", func(t *testing.T) {
		tests := []struct {
			testName       string
			orgIDParam     string
			expectedStatus int
		}{
			// 異常系
			{
				testName:       "invalid org_id query param",
				orgIDParam:     "not-a-uuid",
				expectedStatus: http.StatusInternalServerError,
			},
			// 空文字
			{
				testName:       "empty org_id query param",
				orgIDParam:     "",
				expectedStatus: http.StatusInternalServerError,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				req := httptest.NewRequest(http.MethodGet, "/tags?org_id="+tt.orgIDParam, nil)
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				repo := tag_repository.NewTagRepository(q)
				handler := NewTagHandler(repo).List()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expectedStatus, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("data integrity", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		orgID := seedOrg(t, q)
		created := seedTag(t, q, orgID, "integrity-tag", "#abcdef")

		req := httptest.NewRequest(http.MethodGet, "/tags?org_id="+orgID.String(), nil)
		testutil.SetAuthHeader(req)
		w := httptest.NewRecorder()

		repo := tag_repository.NewTagRepository(q)
		handler := NewTagHandler(repo).List()
		handler.ServeHTTP(w, req)

		var result []map[string]any
		if err := json.NewDecoder(w.Result().Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(result) != 1 {
			t.Fatalf("expected 1 tag, got %d", len(result))
		}

		if diff := cmp.Diff(created.ID.String(), result[0]["id"]); diff != "" {
			t.Errorf("id mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff("integrity-tag", result[0]["name"]); diff != "" {
			t.Errorf("name mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff("#abcdef", result[0]["color"]); diff != "" {
			t.Errorf("color mismatch (-want +got):\n%s", diff)
		}
	})
}

// --- Delete Handler Tests ---

func TestDeleteHandler(t *testing.T) {
	t.Run("204 No Content", func(t *testing.T) {
		tests := []struct {
			testName string
			setup    func(t *testing.T, q db.Querier) string
		}{
			// 正常系
			{
				testName: "delete existing tag",
				setup: func(t *testing.T, q db.Querier) string {
					orgID := seedOrg(t, q)
					tag := seedTag(t, q, orgID, "to-delete", "red")
					return tag.ID.String()
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)
				tagID := tt.setup(t, q)

				req := httptest.NewRequest(http.MethodDelete, "/tags/"+tagID, nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", tagID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				repo := tag_repository.NewTagRepository(q)
				handler := NewTagHandler(repo).Delete()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(http.StatusNoContent, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("error cases", func(t *testing.T) {
		tests := []struct {
			testName       string
			tagID          string
			expectedStatus int
		}{
			// 異常系
			{
				testName:       "invalid UUID",
				tagID:          "not-a-uuid",
				expectedStatus: http.StatusBadRequest,
			},
			// 空文字
			{
				testName:       "empty id",
				tagID:          "",
				expectedStatus: http.StatusBadRequest,
			},
			// 特殊文字
			{
				testName:       "special chars in id",
				tagID:          "x-drop-table-tags",
				expectedStatus: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				req := httptest.NewRequest(http.MethodDelete, "/tags/"+tt.tagID, nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", tt.tagID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				repo := tag_repository.NewTagRepository(q)
				handler := NewTagHandler(repo).Delete()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expectedStatus, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("delete then verify gone", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		orgID := seedOrg(t, q)
		tag := seedTag(t, q, orgID, "delete-verify", "red")
		tagID := tag.ID.String()

		// Delete
		req := httptest.NewRequest(http.MethodDelete, "/tags/"+tagID, nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", tagID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		testutil.SetAuthHeader(req)
		w := httptest.NewRecorder()

		repo := tag_repository.NewTagRepository(q)
		handler := NewTagHandler(repo).Delete()
		handler.ServeHTTP(w, req)

		if diff := cmp.Diff(http.StatusNoContent, w.Result().StatusCode); diff != "" {
			t.Errorf("delete status mismatch (-want +got):\n%s", diff)
		}

		// Verify list returns empty
		req2 := httptest.NewRequest(http.MethodGet, "/tags?org_id="+orgID.String(), nil)
		testutil.SetAuthHeader(req2)
		w2 := httptest.NewRecorder()

		listHandler := NewTagHandler(repo).List()
		listHandler.ServeHTTP(w2, req2)

		var result []map[string]any
		if err := json.NewDecoder(w2.Result().Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if diff := cmp.Diff(0, len(result)); diff != "" {
			t.Errorf("expected empty list after delete (-want +got):\n%s", diff)
		}
	})
}

// --- AddToPrompt Handler Tests ---

func TestAddToPromptHandler(t *testing.T) {
	t.Run("204 No Content", func(t *testing.T) {
		tests := []struct {
			testName string
		}{
			// 正常系
			{testName: "add tag to prompt"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)
				orgID := seedOrg(t, q)
				projID := seedProject(t, q, orgID)
				promptID := seedPrompt(t, q, projID)
				tag := seedTag(t, q, orgID, "feature", "blue")

				body := map[string]string{"tag_id": tag.ID.String()}
				jsonBody, err := json.Marshal(body)
				if err != nil {
					t.Fatalf("failed to marshal body: %v", err)
				}

				req := httptest.NewRequest(http.MethodPost, "/prompts/"+promptID.String()+"/tags", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", promptID.String())
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				repo := tag_repository.NewTagRepository(q)
				handler := NewTagHandler(repo).AddToPrompt()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(http.StatusNoContent, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("idempotent add", func(t *testing.T) {
		// Adding same tag twice should not error (ON CONFLICT DO NOTHING)
		q := testutil.SetupTestTx(t)
		orgID := seedOrg(t, q)
		projID := seedProject(t, q, orgID)
		promptID := seedPrompt(t, q, projID)
		tag := seedTag(t, q, orgID, "idempotent", "green")

		for i := 0; i < 2; i++ {
			body := map[string]string{"tag_id": tag.ID.String()}
			jsonBody, _ := json.Marshal(body)

			req := httptest.NewRequest(http.MethodPost, "/prompts/"+promptID.String()+"/tags", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("prompt_id", promptID.String())
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			testutil.SetAuthHeader(req)
			w := httptest.NewRecorder()

			repo := tag_repository.NewTagRepository(q)
			handler := NewTagHandler(repo).AddToPrompt()
			handler.ServeHTTP(w, req)

			if diff := cmp.Diff(http.StatusNoContent, w.Result().StatusCode); diff != "" {
				t.Errorf("attempt %d: status code mismatch (-want +got):\n%s", i+1, diff)
			}
		}
	})

	t.Run("error cases", func(t *testing.T) {
		tests := []struct {
			testName       string
			promptID       string
			body           string
			expectedStatus int
		}{
			// 異常系
			{
				testName:       "invalid prompt_id",
				promptID:       "not-a-uuid",
				body:           `{"tag_id":"` + uuid.New().String() + `"}`,
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName:       "invalid tag_id in body",
				promptID:       uuid.New().String(),
				body:           `{"tag_id":"not-a-uuid"}`,
				expectedStatus: http.StatusBadRequest,
			},
			// 空文字
			{
				testName:       "empty prompt_id",
				promptID:       "",
				body:           `{"tag_id":"` + uuid.New().String() + `"}`,
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName:       "empty body",
				promptID:       uuid.New().String(),
				body:           `{}`,
				expectedStatus: http.StatusBadRequest,
			},
			// Null/Nil
			{
				testName:       "missing tag_id in body",
				promptID:       uuid.New().String(),
				body:           `{"tag_id":""}`,
				expectedStatus: http.StatusBadRequest,
			},
			// 異常系 - invalid JSON
			{
				testName:       "malformed JSON body",
				promptID:       uuid.New().String(),
				body:           `{invalid`,
				expectedStatus: http.StatusBadRequest,
			},
			// 特殊文字
			{
				testName:       "SQL injection in prompt_id",
				promptID:       "x-drop-prompt-tags",
				body:           `{"tag_id":"` + uuid.New().String() + `"}`,
				expectedStatus: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				req := httptest.NewRequest(http.MethodPost, "/prompts/"+tt.promptID+"/tags", strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", tt.promptID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				repo := tag_repository.NewTagRepository(q)
				handler := NewTagHandler(repo).AddToPrompt()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expectedStatus, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})
}

// --- ListByPrompt Handler Tests ---

func TestListByPromptHandler(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		type expected struct {
			statusCode int
			tagCount   int
		}

		tests := []struct {
			testName string
			setup    func(t *testing.T, q db.Querier) uuid.UUID
			expected expected
		}{
			// 正常系
			{
				testName: "prompt with no tags",
				setup: func(t *testing.T, q db.Querier) uuid.UUID {
					orgID := seedOrg(t, q)
					projID := seedProject(t, q, orgID)
					return seedPrompt(t, q, projID)
				},
				expected: expected{statusCode: http.StatusOK, tagCount: 0},
			},
			{
				testName: "prompt with multiple tags",
				setup: func(t *testing.T, q db.Querier) uuid.UUID {
					orgID := seedOrg(t, q)
					projID := seedProject(t, q, orgID)
					promptID := seedPrompt(t, q, projID)
					tag1 := seedTag(t, q, orgID, "alpha", "red")
					tag2 := seedTag(t, q, orgID, "beta", "blue")
					if err := q.AddPromptTag(context.Background(), db.AddPromptTagParams{PromptID: promptID, TagID: tag1.ID}); err != nil {
						t.Fatalf("failed to add tag: %v", err)
					}
					if err := q.AddPromptTag(context.Background(), db.AddPromptTagParams{PromptID: promptID, TagID: tag2.ID}); err != nil {
						t.Fatalf("failed to add tag: %v", err)
					}
					return promptID
				},
				expected: expected{statusCode: http.StatusOK, tagCount: 2},
			},
			// 境界値
			{
				testName: "tags only for specific prompt",
				setup: func(t *testing.T, q db.Querier) uuid.UUID {
					orgID := seedOrg(t, q)
					projID := seedProject(t, q, orgID)
					promptID1 := seedPrompt(t, q, projID)
					promptID2 := seedPrompt(t, q, projID)
					tag1 := seedTag(t, q, orgID, "for-prompt1", "red")
					tag2 := seedTag(t, q, orgID, "for-prompt2", "blue")
					if err := q.AddPromptTag(context.Background(), db.AddPromptTagParams{PromptID: promptID1, TagID: tag1.ID}); err != nil {
						t.Fatalf("failed to add tag: %v", err)
					}
					if err := q.AddPromptTag(context.Background(), db.AddPromptTagParams{PromptID: promptID2, TagID: tag2.ID}); err != nil {
						t.Fatalf("failed to add tag: %v", err)
					}
					return promptID1
				},
				expected: expected{statusCode: http.StatusOK, tagCount: 1},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)
				promptID := tt.setup(t, q)

				req := httptest.NewRequest(http.MethodGet, "/prompts/"+promptID.String()+"/tags", nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", promptID.String())
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				repo := tag_repository.NewTagRepository(q)
				handler := NewTagHandler(repo).ListByPrompt()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var result []map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(tt.expected.tagCount, len(result)); diff != "" {
					t.Errorf("tag count mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("error cases", func(t *testing.T) {
		tests := []struct {
			testName       string
			promptID       string
			expectedStatus int
		}{
			// 異常系
			{
				testName:       "invalid prompt_id",
				promptID:       "not-a-uuid",
				expectedStatus: http.StatusBadRequest,
			},
			// 空文字
			{
				testName:       "empty prompt_id",
				promptID:       "",
				expectedStatus: http.StatusBadRequest,
			},
			// 特殊文字
			{
				testName:       "SQL injection in prompt_id",
				promptID:       "x-drop-table-tags",
				expectedStatus: http.StatusBadRequest,
			},
			// Null/Nil - non-existent prompt returns empty list, not error
			{
				testName:       "non-existent prompt returns empty",
				promptID:       uuid.New().String(),
				expectedStatus: http.StatusOK,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				req := httptest.NewRequest(http.MethodGet, "/prompts/"+tt.promptID+"/tags", nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", tt.promptID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				repo := tag_repository.NewTagRepository(q)
				handler := NewTagHandler(repo).ListByPrompt()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expectedStatus, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("data integrity", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		orgID := seedOrg(t, q)
		projID := seedProject(t, q, orgID)
		promptID := seedPrompt(t, q, projID)
		tag := seedTag(t, q, orgID, "integrity-check", "#123456")
		if err := q.AddPromptTag(context.Background(), db.AddPromptTagParams{PromptID: promptID, TagID: tag.ID}); err != nil {
			t.Fatalf("failed to add tag: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/prompts/"+promptID.String()+"/tags", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("prompt_id", promptID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		testutil.SetAuthHeader(req)
		w := httptest.NewRecorder()

		repo := tag_repository.NewTagRepository(q)
		handler := NewTagHandler(repo).ListByPrompt()
		handler.ServeHTTP(w, req)

		var result []map[string]any
		if err := json.NewDecoder(w.Result().Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(result) != 1 {
			t.Fatalf("expected 1 tag, got %d", len(result))
		}

		if diff := cmp.Diff(tag.ID.String(), result[0]["id"]); diff != "" {
			t.Errorf("id mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff("integrity-check", result[0]["name"]); diff != "" {
			t.Errorf("name mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff("#123456", result[0]["color"]); diff != "" {
			t.Errorf("color mismatch (-want +got):\n%s", diff)
		}
	})
}

// --- RemoveFromPrompt Handler Tests ---

func TestRemoveFromPromptHandler(t *testing.T) {
	t.Run("204 No Content", func(t *testing.T) {
		tests := []struct {
			testName string
		}{
			// 正常系
			{testName: "remove existing tag from prompt"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)
				orgID := seedOrg(t, q)
				projID := seedProject(t, q, orgID)
				promptID := seedPrompt(t, q, projID)
				tag := seedTag(t, q, orgID, "removable", "red")
				if err := q.AddPromptTag(context.Background(), db.AddPromptTagParams{PromptID: promptID, TagID: tag.ID}); err != nil {
					t.Fatalf("failed to add tag: %v", err)
				}

				req := httptest.NewRequest(http.MethodDelete, "/prompts/"+promptID.String()+"/tags/"+tag.ID.String(), nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", promptID.String())
				rctx.URLParams.Add("tag_id", tag.ID.String())
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				repo := tag_repository.NewTagRepository(q)
				handler := NewTagHandler(repo).RemoveFromPrompt()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(http.StatusNoContent, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("remove then verify gone", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		orgID := seedOrg(t, q)
		projID := seedProject(t, q, orgID)
		promptID := seedPrompt(t, q, projID)
		tag := seedTag(t, q, orgID, "remove-verify", "blue")
		if err := q.AddPromptTag(context.Background(), db.AddPromptTagParams{PromptID: promptID, TagID: tag.ID}); err != nil {
			t.Fatalf("failed to add tag: %v", err)
		}

		// Remove
		req := httptest.NewRequest(http.MethodDelete, "/prompts/"+promptID.String()+"/tags/"+tag.ID.String(), nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("prompt_id", promptID.String())
		rctx.URLParams.Add("tag_id", tag.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		testutil.SetAuthHeader(req)
		w := httptest.NewRecorder()

		repo := tag_repository.NewTagRepository(q)
		handler := NewTagHandler(repo).RemoveFromPrompt()
		handler.ServeHTTP(w, req)

		if diff := cmp.Diff(http.StatusNoContent, w.Result().StatusCode); diff != "" {
			t.Errorf("remove status mismatch (-want +got):\n%s", diff)
		}

		// Verify list is empty
		req2 := httptest.NewRequest(http.MethodGet, "/prompts/"+promptID.String()+"/tags", nil)
		rctx2 := chi.NewRouteContext()
		rctx2.URLParams.Add("prompt_id", promptID.String())
		req2 = req2.WithContext(context.WithValue(req2.Context(), chi.RouteCtxKey, rctx2))
		testutil.SetAuthHeader(req2)
		w2 := httptest.NewRecorder()

		listHandler := NewTagHandler(repo).ListByPrompt()
		listHandler.ServeHTTP(w2, req2)

		var result []map[string]any
		if err := json.NewDecoder(w2.Result().Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if diff := cmp.Diff(0, len(result)); diff != "" {
			t.Errorf("expected empty list after remove (-want +got):\n%s", diff)
		}
	})

	t.Run("error cases", func(t *testing.T) {
		tests := []struct {
			testName       string
			promptID       string
			tagID          string
			expectedStatus int
		}{
			// 異常系
			{
				testName:       "invalid prompt_id",
				promptID:       "not-a-uuid",
				tagID:          uuid.New().String(),
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName:       "invalid tag_id",
				promptID:       uuid.New().String(),
				tagID:          "not-a-uuid",
				expectedStatus: http.StatusBadRequest,
			},
			// 空文字
			{
				testName:       "empty prompt_id",
				promptID:       "",
				tagID:          uuid.New().String(),
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName:       "empty tag_id",
				promptID:       uuid.New().String(),
				tagID:          "",
				expectedStatus: http.StatusBadRequest,
			},
			// 特殊文字
			{
				testName:       "SQL injection in prompt_id",
				promptID:       "x-drop-prompt-tags",
				tagID:          uuid.New().String(),
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName:       "SQL injection in tag_id",
				promptID:       uuid.New().String(),
				tagID:          "x-drop-table-tags",
				expectedStatus: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				req := httptest.NewRequest(http.MethodDelete, "/prompts/"+tt.promptID+"/tags/"+tt.tagID, nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", tt.promptID)
				rctx.URLParams.Add("tag_id", tt.tagID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				repo := tag_repository.NewTagRepository(q)
				handler := NewTagHandler(repo).RemoveFromPrompt()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expectedStatus, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("remove non-existent association is idempotent", func(t *testing.T) {
		// Removing a tag that isn't associated should not error (DELETE WHERE returns no rows)
		q := testutil.SetupTestTx(t)
		orgID := seedOrg(t, q)
		projID := seedProject(t, q, orgID)
		promptID := seedPrompt(t, q, projID)
		tag := seedTag(t, q, orgID, "not-associated", "gray")

		req := httptest.NewRequest(http.MethodDelete, "/prompts/"+promptID.String()+"/tags/"+tag.ID.String(), nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("prompt_id", promptID.String())
		rctx.URLParams.Add("tag_id", tag.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		testutil.SetAuthHeader(req)
		w := httptest.NewRecorder()

		repo := tag_repository.NewTagRepository(q)
		handler := NewTagHandler(repo).RemoveFromPrompt()
		handler.ServeHTTP(w, req)

		if diff := cmp.Diff(http.StatusNoContent, w.Result().StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}
	})
}
