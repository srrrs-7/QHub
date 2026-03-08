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

	"api/src/domain/organization"
	"api/src/domain/project"
	"api/src/infra/rds/organization_repository"
	"api/src/infra/rds/project_repository"
	"api/src/infra/rds/prompt_repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
)

func TestPostPromptHandler(t *testing.T) {
	t.Run("201 Created", func(t *testing.T) {
		type expected struct {
			statusCode  int
			name        string
			slug        string
			promptType  string
			description string
		}

		tests := []struct {
			testName string
			reqBody  map[string]any
			expected expected
		}{
			// 正常系
			{
				testName: "create system prompt",
				reqBody:  map[string]any{"name": "My Prompt", "slug": "my-prompt", "prompt_type": "system", "description": "A test prompt"},
				expected: expected{statusCode: http.StatusCreated, name: "My Prompt", slug: "my-prompt", promptType: "system", description: "A test prompt"},
			},
			// 特殊文字
			{
				testName: "create prompt with Japanese name",
				reqBody:  map[string]any{"name": "テストプロンプト", "slug": "test-prompt", "prompt_type": "user"},
				expected: expected{statusCode: http.StatusCreated, name: "テストプロンプト", slug: "test-prompt", promptType: "user", description: ""},
			},
			// 境界値
			{
				testName: "create with min length name and slug",
				reqBody:  map[string]any{"name": "AB", "slug": "ab", "prompt_type": "combined"},
				expected: expected{statusCode: http.StatusCreated, name: "AB", slug: "ab", promptType: "combined", description: ""},
			},
			// 空文字 (description optional)
			{
				testName: "create without description",
				reqBody:  map[string]any{"name": "No Desc", "slug": "no-desc", "prompt_type": "system"},
				expected: expected{statusCode: http.StatusCreated, name: "No Desc", slug: "no-desc", promptType: "system", description: ""},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				projectID := createTestProject(t, q)

				jsonBody, err := json.Marshal(tt.reqBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}

				req := httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/prompts", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				// Set chi URL params
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("project_id", projectID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				pRepo := prompt_repository.NewPromptRepository(q)
				vRepo := prompt_repository.NewVersionRepository(q)
				handler := NewPromptHandler(pRepo, vRepo).Post()
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
				if body.ID == "" {
					t.Error("expected non-empty ID")
				}
				if body.ProjectID != projectID {
					t.Errorf("expected project_id %s, got %s", projectID, body.ProjectID)
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
			{testName: "missing name", reqBody: map[string]any{"slug": "my-prompt", "prompt_type": "system"}},
			{testName: "missing slug", reqBody: map[string]any{"name": "My Prompt", "prompt_type": "system"}},
			{testName: "missing prompt_type", reqBody: map[string]any{"name": "My Prompt", "slug": "my-prompt"}},
			{testName: "invalid prompt_type", reqBody: map[string]any{"name": "My Prompt", "slug": "my-prompt", "prompt_type": "invalid"}},
			// 空文字
			{testName: "empty name", reqBody: map[string]any{"name": "", "slug": "my-prompt", "prompt_type": "system"}},
			{testName: "empty slug", reqBody: map[string]any{"name": "My Prompt", "slug": "", "prompt_type": "system"}},
			// 境界値
			{testName: "name too short", reqBody: map[string]any{"name": "A", "slug": "my-prompt", "prompt_type": "system"}},
			{testName: "slug too short", reqBody: map[string]any{"name": "My Prompt", "slug": "a", "prompt_type": "system"}},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				projectID := createTestProject(t, q)

				jsonBody, err := json.Marshal(tt.reqBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}

				req := httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/prompts", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("project_id", projectID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				pRepo := prompt_repository.NewPromptRepository(q)
				vRepo := prompt_repository.NewVersionRepository(q)
				handler := NewPromptHandler(pRepo, vRepo).Post()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode != http.StatusBadRequest {
					t.Errorf("expected status 400, got %d", resp.StatusCode)
				}
			})
		}
	})

	t.Run("400 Invalid JSON", func(t *testing.T) {
		q := setupTestHandler(t)
		projectID := createTestProject(t, q)

		req := httptest.NewRequest(http.MethodPost, "/projects/"+projectID+"/prompts", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("project_id", projectID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		pRepo := prompt_repository.NewPromptRepository(q)
		vRepo := prompt_repository.NewVersionRepository(q)
		handler := NewPromptHandler(pRepo, vRepo).Post()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", resp.StatusCode)
		}
	})
}

func setupTestHandler(t *testing.T) db.Querier {
	t.Helper()
	return testutil.SetupTestTx(t)
}

func createTestProject(t *testing.T, q db.Querier) string {
	t.Helper()

	orgRepo := organization_repository.NewOrganizationRepository(q)
	orgCmd := organization.NewOrganizationCmd(
		organization.OrganizationName("Test Org"),
		organization.OrganizationSlug("test-org"),
		organization.PlanFree,
	)
	org, err := orgRepo.Create(t.Context(), orgCmd)
	if err != nil {
		t.Fatalf("failed to create test org: %v", err)
	}

	projRepo := project_repository.NewProjectRepository(q)
	projCmd := project.NewProjectCmd(
		org.ID.UUID(),
		project.ProjectName("Test Project"),
		project.ProjectSlug("test-project"),
		project.ProjectDescription(""),
	)
	proj, err := projRepo.Create(t.Context(), projCmd)
	if err != nil {
		t.Fatalf("failed to create test project: %v", err)
	}
	return proj.ID.String()
}
