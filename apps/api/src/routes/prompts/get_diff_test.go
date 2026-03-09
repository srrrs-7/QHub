package prompts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"utils/db/db"
	"utils/testutil"

	"api/src/domain/prompt"
	"api/src/infra/rds/prompt_repository"
	"api/src/services/diffservice"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestGetDiffHandler(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		tests := []struct {
			testName string
			v1       string
			v2       string
		}{
			// 正常系
			{testName: "diff between version 1 and 2", v1: "1", v2: "2"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				projectID := createTestProject(t, q)
				promptID := createTestPromptRecord(t, q, projectID)
				authorID := createTestUser(t, q)
				createDiffVersions(t, q, promptID, authorID)

				vRepo := prompt_repository.NewVersionRepository(q)
				pRepo := prompt_repository.NewPromptRepository(q)
				diffSvc := diffservice.NewDiffService(vRepo)
				handler := NewPromptHandler(pRepo, vRepo, diffSvc, nil, nil).GetDiff()

				req := httptest.NewRequest(http.MethodGet, "/prompts/"+promptID+"/semantic-diff/"+tt.v1+"/"+tt.v2, nil)
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", promptID)
				rctx.URLParams.Add("v1", tt.v1)
				rctx.URLParams.Add("v2", tt.v2)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(http.StatusOK, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var body map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if _, ok := body["summary"]; !ok {
					t.Error("expected summary field in diff response")
				}
				if _, ok := body["changes"]; !ok {
					t.Error("expected changes field in diff response")
				}
			})
		}
	})

	t.Run("404 Not Found", func(t *testing.T) {
		tests := []struct {
			testName string
			v1       string
			v2       string
		}{
			// 異常系
			{testName: "non-existent from version", v1: "999", v2: "1"},
			{testName: "non-existent to version", v1: "1", v2: "999"},
			// 境界値
			{testName: "version zero", v1: "0", v2: "1"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				projectID := createTestProject(t, q)
				promptID := createTestPromptRecord(t, q, projectID)
				authorID := createTestUser(t, q)
				createDiffVersions(t, q, promptID, authorID)

				vRepo := prompt_repository.NewVersionRepository(q)
				pRepo := prompt_repository.NewPromptRepository(q)
				diffSvc := diffservice.NewDiffService(vRepo)
				handler := NewPromptHandler(pRepo, vRepo, diffSvc, nil, nil).GetDiff()

				req := httptest.NewRequest(http.MethodGet, "/prompts/"+promptID+"/semantic-diff/"+tt.v1+"/"+tt.v2, nil)
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", promptID)
				rctx.URLParams.Add("v1", tt.v1)
				rctx.URLParams.Add("v2", tt.v2)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()
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
			testName string
			promptID string
			v1       string
			v2       string
		}{
			// 異常系
			{testName: "invalid prompt_id", promptID: "not-a-uuid", v1: "1", v2: "2"},
			// 特殊文字 — strconv.Atoi errors are not AppError so handler returns 500
			{testName: "non-numeric v1", promptID: uuid.New().String(), v1: "abc", v2: "2"},
			{testName: "non-numeric v2", promptID: uuid.New().String(), v1: "1", v2: "xyz"},
			// 空文字
			{testName: "empty prompt_id", promptID: "", v1: "1", v2: "2"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)

				vRepo := prompt_repository.NewVersionRepository(q)
				pRepo := prompt_repository.NewPromptRepository(q)
				diffSvc := diffservice.NewDiffService(vRepo)
				handler := NewPromptHandler(pRepo, vRepo, diffSvc, nil, nil).GetDiff()

				req := httptest.NewRequest(http.MethodGet, "/prompts/"+tt.promptID+"/semantic-diff/"+tt.v1+"/"+tt.v2, nil)
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("prompt_id", tt.promptID)
				rctx.URLParams.Add("v1", tt.v1)
				rctx.URLParams.Add("v2", tt.v2)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode == http.StatusOK {
					t.Errorf("expected error status, got %d", resp.StatusCode)
				}
			})
		}
	})
}

// createDiffVersions creates two versions with different content suitable for diff testing.
func createDiffVersions(t *testing.T, q db.Querier, promptID, authorID string) {
	t.Helper()

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
	pid := prompt.PromptIDFromUUID(parsedPromptID)

	contents := []string{
		`{"system":"You are a helpful assistant. Please respond in a formal tone."}`,
		`{"system":"You are a friendly assistant. Just respond casually and include {{topic}} details."}`,
	}

	for i, content := range contents {
		vNum := i + 1
		_, err := vRepo.Create(t.Context(), prompt.VersionCmd{
			PromptID:          pid,
			Content:           json.RawMessage(content),
			Variables:         nil,
			ChangeDescription: prompt.ChangeDescription(fmt.Sprintf("Version %d", vNum)),
			AuthorID:          parsedAuthorID,
		}, vNum)
		if err != nil {
			t.Fatalf("failed to create version %d: %v", vNum, err)
		}

		_, err = pRepo.UpdateLatestVersion(t.Context(), pid, vNum)
		if err != nil {
			t.Fatalf("failed to update latest version: %v", err)
		}
	}
}
