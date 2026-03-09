package prompts

import (
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

func TestListVersionsHandler(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		tests := []struct {
			testName      string
			versionCount  int
			expectedCount int
		}{
			// 正常系
			{testName: "list versions returns all versions", versionCount: 3, expectedCount: 3},
			// 境界値
			{testName: "list versions returns single version", versionCount: 1, expectedCount: 1},
			// Null/Nil - empty list
			{testName: "list versions returns empty for no versions", versionCount: 0, expectedCount: 0},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				projectID := createTestProject(t, q)
				promptID := createTestPromptRecord(t, q, projectID)
				authorID := createTestUser(t, q)

				createVersions(t, q, promptID, authorID, tt.versionCount)

				req := httptest.NewRequest(http.MethodGet, "/prompts/"+promptID+"/versions", nil)
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", promptID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				pRepo := prompt_repository.NewPromptRepository(q)
				vRepo := prompt_repository.NewVersionRepository(q)
				handler := NewPromptHandler(pRepo, vRepo, nil, nil, nil).ListVersions()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(http.StatusOK, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var body []versionResponse
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
			testName string
			promptID string
		}{
			// 異常系
			{testName: "invalid prompt_id", promptID: "not-a-uuid"},
			// 空文字
			{testName: "empty prompt_id", promptID: ""},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)

				req := httptest.NewRequest(http.MethodGet, "/prompts/"+tt.promptID+"/versions", nil)
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", tt.promptID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				pRepo := prompt_repository.NewPromptRepository(q)
				vRepo := prompt_repository.NewVersionRepository(q)
				handler := NewPromptHandler(pRepo, vRepo, nil, nil, nil).ListVersions()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode != http.StatusBadRequest {
					t.Errorf("expected status 400, got %d", resp.StatusCode)
				}
			})
		}
	})
}

// createVersions creates the specified number of prompt versions for testing.
func createVersions(t *testing.T, q db.Querier, promptID, authorID string, count int) {
	t.Helper()

	if count == 0 {
		return
	}

	parsedPromptID, err := uuid.Parse(promptID)
	if err != nil {
		t.Fatalf("failed to parse prompt ID: %v", err)
	}

	parsedAuthorID, err := uuid.Parse(authorID)
	if err != nil {
		t.Fatalf("failed to parse author ID: %v", err)
	}

	pRepo := prompt_repository.NewPromptRepository(q)
	vRepo := prompt_repository.NewVersionRepository(q)

	for i := 1; i <= count; i++ {
		_, err := vRepo.Create(t.Context(), prompt.VersionCmd{
			PromptID:          prompt.PromptIDFromUUID(parsedPromptID),
			Content:           json.RawMessage(`{"system":"Version ` + string(rune('0'+i)) + ` content"}`),
			Variables:         nil,
			ChangeDescription: prompt.ChangeDescription("Version " + string(rune('0'+i))),
			AuthorID:          parsedAuthorID,
		}, i)
		if err != nil {
			t.Fatalf("failed to create version %d: %v", i, err)
		}

		_, err = pRepo.UpdateLatestVersion(t.Context(), prompt.PromptIDFromUUID(parsedPromptID), i)
		if err != nil {
			t.Fatalf("failed to update latest version: %v", err)
		}
	}
}
