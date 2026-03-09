package prompts

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"utils/testutil"

	"api/src/infra/rds/prompt_repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestGetHandler(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		type expected struct {
			statusCode  int
			name        string
			slug        string
			promptType  string
			description string
		}

		tests := []struct {
			testName string
			slug     string
			expected expected
		}{
			// 正常系
			{
				testName: "get existing prompt by slug",
				slug:     "test-prompt",
				expected: expected{
					statusCode:  http.StatusOK,
					name:        "Test Prompt",
					slug:        "test-prompt",
					promptType:  "system",
					description: "",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				projectID := createTestProject(t, q)
				_ = createTestPromptRecord(t, q, projectID)

				req := httptest.NewRequest(http.MethodGet, "/projects/"+projectID+"/prompts/"+tt.slug, nil)
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("project_id", projectID)
				rctx.URLParams.Add("prompt_slug", tt.slug)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				pRepo := prompt_repository.NewPromptRepository(q)
				vRepo := prompt_repository.NewVersionRepository(q)
				handler := NewPromptHandler(pRepo, vRepo, nil, nil, nil).Get()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var body promptResponse
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(tt.expected.name, body.Name); diff != "" {
					t.Errorf("name mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expected.slug, body.Slug); diff != "" {
					t.Errorf("slug mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expected.promptType, body.PromptType); diff != "" {
					t.Errorf("prompt_type mismatch (-want +got):\n%s", diff)
				}
				if body.ProjectID != projectID {
					t.Errorf("expected project_id %s, got %s", projectID, body.ProjectID)
				}
			})
		}
	})

	t.Run("404 Not Found", func(t *testing.T) {
		tests := []struct {
			testName string
			slug     string
		}{
			// 異常系
			{testName: "non-existent prompt slug", slug: "non-existent"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				projectID := createTestProject(t, q)

				req := httptest.NewRequest(http.MethodGet, "/projects/"+projectID+"/prompts/"+tt.slug, nil)
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("project_id", projectID)
				rctx.URLParams.Add("prompt_slug", tt.slug)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				pRepo := prompt_repository.NewPromptRepository(q)
				vRepo := prompt_repository.NewVersionRepository(q)
				handler := NewPromptHandler(pRepo, vRepo, nil, nil, nil).Get()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode != http.StatusNotFound {
					t.Errorf("expected status 404, got %d", resp.StatusCode)
				}
			})
		}
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		tests := []struct {
			testName  string
			projectID string
			slug      string
		}{
			// 異常系
			{testName: "invalid project_id", projectID: "not-a-uuid", slug: "test-prompt"},
			// 境界値
			{testName: "slug too short", projectID: uuid.New().String(), slug: "a"},
			// 空文字
			{testName: "empty slug", projectID: uuid.New().String(), slug: ""},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)

				req := httptest.NewRequest(http.MethodGet, "/projects/"+tt.projectID+"/prompts/"+tt.slug, nil)
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("project_id", tt.projectID)
				rctx.URLParams.Add("prompt_slug", tt.slug)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				pRepo := prompt_repository.NewPromptRepository(q)
				vRepo := prompt_repository.NewVersionRepository(q)
				handler := NewPromptHandler(pRepo, vRepo, nil, nil, nil).Get()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode != http.StatusBadRequest {
					t.Errorf("expected status 400, got %d", resp.StatusCode)
				}
			})
		}
	})
}
