package prompts

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"utils/testutil"

	"api/src/domain/prompt"
	"api/src/infra/rds/prompt_repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestListHandler(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		tests := []struct {
			testName      string
			promptCount   int
			expectedCount int
		}{
			// 正常系
			{testName: "list prompts returns all prompts", promptCount: 3, expectedCount: 3},
			// 境界値
			{testName: "list prompts returns single prompt", promptCount: 1, expectedCount: 1},
			// Null/Nil - empty list
			{testName: "list prompts returns empty for no prompts", promptCount: 0, expectedCount: 0},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				projectID := createTestProject(t, q)

				parsedProjectID, err := uuid.Parse(projectID)
				if err != nil {
					t.Fatalf("failed to parse project ID: %v", err)
				}

				pRepo := prompt_repository.NewPromptRepository(q)
				for i := range tt.promptCount {
					_, err := pRepo.Create(t.Context(), prompt.PromptCmd{
						ProjectID:   parsedProjectID,
						Name:        prompt.PromptName("Prompt " + string(rune('A'+i))),
						Slug:        prompt.PromptSlug("prompt-" + string(rune('a'+i))),
						PromptType:  prompt.PromptTypeSystem,
						Description: prompt.PromptDescription(""),
					})
					if err != nil {
						t.Fatalf("failed to create prompt %d: %v", i, err)
					}
				}

				req := httptest.NewRequest(http.MethodGet, "/projects/"+projectID+"/prompts", nil)
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("project_id", projectID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				vRepo := prompt_repository.NewVersionRepository(q)
				handler := NewPromptHandler(pRepo, vRepo, nil, nil, nil).List()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(http.StatusOK, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var body []promptResponse
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(tt.expectedCount, len(body)); diff != "" {
					t.Errorf("count mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		tests := []struct {
			testName  string
			projectID string
		}{
			// 異常系
			{testName: "invalid project_id", projectID: "not-a-uuid"},
			// 空文字
			{testName: "empty project_id", projectID: ""},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)

				req := httptest.NewRequest(http.MethodGet, "/projects/"+tt.projectID+"/prompts", nil)
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("project_id", tt.projectID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				pRepo := prompt_repository.NewPromptRepository(q)
				vRepo := prompt_repository.NewVersionRepository(q)
				handler := NewPromptHandler(pRepo, vRepo, nil, nil, nil).List()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode != http.StatusBadRequest {
					t.Errorf("expected status 400, got %d", resp.StatusCode)
				}
			})
		}
	})
}
