package organizations

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"utils/db/db"
	"utils/testutil"

	"api/src/infra/rds/organization_repository"

	"github.com/google/go-cmp/cmp"
)

func TestPostHandler(t *testing.T) {
	t.Run("201 Created", func(t *testing.T) {
		type expected struct {
			statusCode int
			name       string
			slug       string
			plan       string
		}

		tests := []struct {
			testName string
			reqBody  map[string]string
			expected expected
		}{
			// 正常系
			{
				testName: "create organization with name and slug",
				reqBody:  map[string]string{"name": "My Organization", "slug": "my-org"},
				expected: expected{statusCode: http.StatusCreated, name: "My Organization", slug: "my-org", plan: "free"},
			},
			// 特殊文字
			{
				testName: "create organization with Japanese name",
				reqBody:  map[string]string{"name": "株式会社テスト", "slug": "test-corp"},
				expected: expected{statusCode: http.StatusCreated, name: "株式会社テスト", slug: "test-corp", plan: "free"},
			},
			// 境界値
			{
				testName: "create with min length name and slug",
				reqBody:  map[string]string{"name": "AB", "slug": "ab"},
				expected: expected{statusCode: http.StatusCreated, name: "AB", slug: "ab", plan: "free"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)

				jsonBody, err := json.Marshal(tt.reqBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}
				req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()

				repo := organization_repository.NewOrganizationRepository(q)
				handler := NewOrganizationHandler(repo).Post()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var body organizationResponse
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(tt.expected.name, body.Name); diff != "" {
					t.Errorf("name mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expected.slug, body.Slug); diff != "" {
					t.Errorf("slug mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expected.plan, body.Plan); diff != "" {
					t.Errorf("plan mismatch (-want +got):\n%s", diff)
				}
				if body.ID == "" {
					t.Error("expected non-empty ID")
				}
			})
		}
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		tests := []struct {
			testName string
			reqBody  map[string]string
		}{
			// 異常系
			{testName: "missing name", reqBody: map[string]string{"slug": "my-org"}},
			{testName: "missing slug", reqBody: map[string]string{"name": "My Org"}},
			// 空文字
			{testName: "empty name", reqBody: map[string]string{"name": "", "slug": "my-org"}},
			{testName: "empty slug", reqBody: map[string]string{"name": "My Org", "slug": ""}},
			// 境界値
			{testName: "name too short", reqBody: map[string]string{"name": "A", "slug": "my-org"}},
			{testName: "slug too short", reqBody: map[string]string{"name": "My Org", "slug": "a"}},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)

				jsonBody, err := json.Marshal(tt.reqBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}
				req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()

				repo := organization_repository.NewOrganizationRepository(q)
				handler := NewOrganizationHandler(repo).Post()
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

		req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		w := httptest.NewRecorder()

		repo := organization_repository.NewOrganizationRepository(q)
		handler := NewOrganizationHandler(repo).Post()
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
