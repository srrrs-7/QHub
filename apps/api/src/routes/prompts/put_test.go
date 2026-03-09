package prompts

import (
	"bytes"
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

func TestPutHandler(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		type expected struct {
			statusCode  int
			name        string
			slug        string
			description string
		}

		tests := []struct {
			testName string
			reqBody  map[string]any
			expected expected
		}{
			// 正常系
			{
				testName: "update name only",
				reqBody:  map[string]any{"name": "Updated Prompt"},
				expected: expected{statusCode: http.StatusOK, name: "Updated Prompt", slug: "test-prompt", description: ""},
			},
			{
				testName: "update slug only",
				reqBody:  map[string]any{"slug": "updated-slug"},
				expected: expected{statusCode: http.StatusOK, name: "Test Prompt", slug: "updated-slug", description: ""},
			},
			{
				testName: "update description only",
				reqBody:  map[string]any{"description": "New description"},
				expected: expected{statusCode: http.StatusOK, name: "Test Prompt", slug: "test-prompt", description: "New description"},
			},
			{
				testName: "update all fields",
				reqBody:  map[string]any{"name": "New Name", "slug": "new-slug", "description": "New desc"},
				expected: expected{statusCode: http.StatusOK, name: "New Name", slug: "new-slug", description: "New desc"},
			},
			// 特殊文字
			{
				testName: "update with Japanese name",
				reqBody:  map[string]any{"name": "日本語プロンプト名"},
				expected: expected{statusCode: http.StatusOK, name: "日本語プロンプト名", slug: "test-prompt", description: ""},
			},
			{
				testName: "update with emoji in description",
				reqBody:  map[string]any{"description": "Updated with emoji 🚀"},
				expected: expected{statusCode: http.StatusOK, name: "Test Prompt", slug: "test-prompt", description: "Updated with emoji 🚀"},
			},
			// 空文字 - empty body keeps existing values
			{
				testName: "empty body keeps existing values",
				reqBody:  map[string]any{},
				expected: expected{statusCode: http.StatusOK, name: "Test Prompt", slug: "test-prompt", description: ""},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				projectID := createTestProject(t, q)
				_ = createTestPromptRecord(t, q, projectID)

				jsonBody, err := json.Marshal(tt.reqBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}

				req := httptest.NewRequest(http.MethodPut, "/projects/"+projectID+"/prompts/test-prompt", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("project_id", projectID)
				rctx.URLParams.Add("prompt_slug", "test-prompt")
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				pRepo := prompt_repository.NewPromptRepository(q)
				vRepo := prompt_repository.NewVersionRepository(q)
				handler := NewPromptHandler(pRepo, vRepo, nil, nil, nil).Put()
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
				if diff := cmp.Diff(tt.expected.description, body.Description); diff != "" {
					t.Errorf("description mismatch (-want +got):\n%s", diff)
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
			{testName: "non-existent prompt", slug: "non-existent"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				projectID := createTestProject(t, q)

				jsonBody, _ := json.Marshal(map[string]any{"name": "Updated"})

				req := httptest.NewRequest(http.MethodPut, "/projects/"+projectID+"/prompts/"+tt.slug, bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("project_id", projectID)
				rctx.URLParams.Add("prompt_slug", tt.slug)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				pRepo := prompt_repository.NewPromptRepository(q)
				vRepo := prompt_repository.NewVersionRepository(q)
				handler := NewPromptHandler(pRepo, vRepo, nil, nil, nil).Put()
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
			reqBody   map[string]any
		}{
			// 異常系
			{testName: "invalid project_id", projectID: "not-a-uuid", slug: "test-prompt", reqBody: map[string]any{"name": "X"}},
			// 境界値
			{testName: "name too short", projectID: uuid.New().String(), slug: "test-prompt", reqBody: map[string]any{"name": "A"}},
			{testName: "slug too short", projectID: uuid.New().String(), slug: "test-prompt", reqBody: map[string]any{"slug": "a"}},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)

				// For the boundary value tests, we need a real prompt
				if tt.testName != "invalid project_id" {
					projID := createTestProject(t, q)
					_ = createTestPromptRecord(t, q, projID)
					tt.projectID = projID
				}

				jsonBody, _ := json.Marshal(tt.reqBody)

				req := httptest.NewRequest(http.MethodPut, "/projects/"+tt.projectID+"/prompts/"+tt.slug, bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("project_id", tt.projectID)
				rctx.URLParams.Add("prompt_slug", tt.slug)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				pRepo := prompt_repository.NewPromptRepository(q)
				vRepo := prompt_repository.NewVersionRepository(q)
				handler := NewPromptHandler(pRepo, vRepo, nil, nil, nil).Put()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode != http.StatusBadRequest {
					t.Errorf("expected status 400, got %d", resp.StatusCode)
				}
			})
		}
	})
}
