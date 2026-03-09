package prompts

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"utils/db/db"
	"utils/testutil"

	"api/src/domain/prompt"
	"api/src/infra/rds/prompt_repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestPostVersionHandler(t *testing.T) {
	t.Run("201 Created", func(t *testing.T) {
		type expected struct {
			statusCode        int
			status            string
			changeDescription string
		}

		tests := []struct {
			testName string
			reqBody  map[string]any
			expected expected
		}{
			// 正常系
			{
				testName: "create version with content and variables",
				reqBody: map[string]any{
					"content":            map[string]any{"system": "You are a helpful assistant"},
					"variables":          []map[string]any{{"name": "topic", "type": "string"}},
					"change_description": "Initial version",
				},
				expected: expected{statusCode: http.StatusCreated, status: "draft", changeDescription: "Initial version"},
			},
			// 特殊文字
			{
				testName: "create version with Japanese content",
				reqBody: map[string]any{
					"content":            map[string]any{"system": "あなたは親切なアシスタントです"},
					"change_description": "日本語テスト",
				},
				expected: expected{statusCode: http.StatusCreated, status: "draft", changeDescription: "日本語テスト"},
			},
			// 空文字 (change_description optional)
			{
				testName: "create version without change description",
				reqBody: map[string]any{
					"content": map[string]any{"system": "You are a helpful assistant"},
				},
				expected: expected{statusCode: http.StatusCreated, status: "draft", changeDescription: ""},
			},
			// Null/Nil (variables optional)
			{
				testName: "create version without variables",
				reqBody: map[string]any{
					"content": map[string]any{"system": "Simple prompt"},
				},
				expected: expected{statusCode: http.StatusCreated, status: "draft", changeDescription: ""},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				projectID := createTestProject(t, q)
				promptID := createTestPromptRecord(t, q, projectID)
				authorID := createTestUser(t, q)

				tt.reqBody["author_id"] = authorID

				jsonBody, err := json.Marshal(tt.reqBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}

				req := httptest.NewRequest(http.MethodPost, "/prompts/"+promptID+"/versions", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", promptID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				pRepo := prompt_repository.NewPromptRepository(q)
				vRepo := prompt_repository.NewVersionRepository(q)
				handler := NewPromptHandler(pRepo, vRepo, nil, nil).PostVersion()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var body versionResponse
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(tt.expected.status, body.Status); diff != "" {
					t.Errorf("status mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expected.changeDescription, body.ChangeDescription); diff != "" {
					t.Errorf("change_description mismatch (-want +got):\n%s", diff)
				}
				if body.ID == "" {
					t.Error("expected non-empty ID")
				}
				if body.PromptID != promptID {
					t.Errorf("expected prompt_id %s, got %s", promptID, body.PromptID)
				}
			})
		}
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		tests := []struct {
			testName string
			reqBody  map[string]any
		}{
			// 異常系
			{testName: "missing content", reqBody: map[string]any{"author_id": uuid.New().String()}},
			{testName: "missing author_id", reqBody: map[string]any{"content": map[string]any{"system": "test"}}},
			{testName: "invalid author_id", reqBody: map[string]any{"content": map[string]any{"system": "test"}, "author_id": "not-uuid"}},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				projectID := createTestProject(t, q)
				promptID := createTestPromptRecord(t, q, projectID)

				jsonBody, err := json.Marshal(tt.reqBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}

				req := httptest.NewRequest(http.MethodPost, "/prompts/"+promptID+"/versions", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", promptID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				pRepo := prompt_repository.NewPromptRepository(q)
				vRepo := prompt_repository.NewVersionRepository(q)
				handler := NewPromptHandler(pRepo, vRepo, nil, nil).PostVersion()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode != http.StatusBadRequest {
					t.Errorf("expected status 400, got %d", resp.StatusCode)
				}
			})
		}
	})
}

func createTestPromptRecord(t *testing.T, q db.Querier, projectID string) string {
	t.Helper()

	parsedProjectID, err := uuid.Parse(projectID)
	if err != nil {
		t.Fatalf("failed to parse project ID: %v", err)
	}

	pRepo := prompt_repository.NewPromptRepository(q)
	cmd := prompt.PromptCmd{
		ProjectID:   parsedProjectID,
		Name:        prompt.PromptName("Test Prompt"),
		Slug:        prompt.PromptSlug("test-prompt"),
		PromptType:  prompt.PromptTypeSystem,
		Description: prompt.PromptDescription(""),
	}
	p, err := pRepo.Create(t.Context(), cmd)
	if err != nil {
		t.Fatalf("failed to create test prompt: %v", err)
	}
	return p.ID.String()
}

func createTestUser(t *testing.T, q db.Querier) string {
	t.Helper()

	user, err := q.CreateUser(t.Context(), db.CreateUserParams{
		Email: "test-" + uuid.New().String()[:8] + "@example.com",
		Name:  "Test User",
	})
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user.ID.String()
}
