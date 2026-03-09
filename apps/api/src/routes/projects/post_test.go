package projects

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
	"api/src/infra/rds/organization_repository"
	"api/src/infra/rds/project_repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
)

func TestPostHandler(t *testing.T) {
	t.Run("201 Created", func(t *testing.T) {
		type expected struct {
			statusCode  int
			name        string
			slug        string
			description string
		}

		tests := []struct {
			testName string
			reqBody  map[string]string
			expected expected
		}{
			// 正常系
			{
				testName: "create project with all fields",
				reqBody:  map[string]string{"name": "My Project", "slug": "my-project", "description": "A test project"},
				expected: expected{statusCode: http.StatusCreated, name: "My Project", slug: "my-project", description: "A test project"},
			},
			// 特殊文字
			{
				testName: "create project with Japanese name",
				reqBody:  map[string]string{"name": "テストプロジェクト", "slug": "test-project", "description": "テスト説明"},
				expected: expected{statusCode: http.StatusCreated, name: "テストプロジェクト", slug: "test-project", description: "テスト説明"},
			},
			// 境界値
			{
				testName: "create with min length name and slug",
				reqBody:  map[string]string{"name": "AB", "slug": "ab"},
				expected: expected{statusCode: http.StatusCreated, name: "AB", slug: "ab", description: ""},
			},
			// 空文字 (description optional)
			{
				testName: "create without description",
				reqBody:  map[string]string{"name": "No Desc Project", "slug": "no-desc"},
				expected: expected{statusCode: http.StatusCreated, name: "No Desc Project", slug: "no-desc", description: ""},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				orgID := createTestOrg(t, q)

				// organization_id comes from the URL param, not the body.
				jsonBody, err := json.Marshal(tt.reqBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}
				req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				// Inject org_id as chi URL param.
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_id", orgID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()

				repo := project_repository.NewProjectRepository(q)
				handler := NewProjectHandler(repo).Post()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var body projectResponse
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
				if body.ID == "" {
					t.Error("expected non-empty ID")
				}
				if body.OrganizationID != orgID {
					t.Errorf("expected organization_id %s, got %s", orgID, body.OrganizationID)
				}
			})
		}
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		type args struct {
			reqBody map[string]string
			orgID   string // chi URL param; "" means omit (simulate missing param)
		}

		tests := []struct {
			testName string
			args     args
		}{
			// 異常系 — invalid / missing body fields
			{testName: "missing name", args: args{orgID: "valid", reqBody: map[string]string{"slug": "my-project"}}},
			{testName: "missing slug", args: args{orgID: "valid", reqBody: map[string]string{"name": "My Project"}}},
			// 空文字
			{testName: "empty name", args: args{orgID: "valid", reqBody: map[string]string{"name": "", "slug": "my-project"}}},
			{testName: "empty slug", args: args{orgID: "valid", reqBody: map[string]string{"name": "My Project", "slug": ""}}},
			// 境界値
			{testName: "name too short", args: args{orgID: "valid", reqBody: map[string]string{"name": "A", "slug": "my-project"}}},
			{testName: "slug too short", args: args{orgID: "valid", reqBody: map[string]string{"name": "My Project", "slug": "a"}}},
			// 異常系 — invalid / missing org_id URL param
			{testName: "missing org_id URL param", args: args{orgID: "", reqBody: map[string]string{"name": "My Project", "slug": "my-project"}}},
			{testName: "invalid org_id URL param", args: args{orgID: "not-a-uuid", reqBody: map[string]string{"name": "My Project", "slug": "my-project"}}},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				realOrgID := createTestOrg(t, q)

				jsonBody, err := json.Marshal(tt.args.reqBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}
				req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				// Inject org_id URL param. Use real orgID when args.orgID == "valid".
				orgIDParam := tt.args.orgID
				if orgIDParam == "valid" {
					orgIDParam = realOrgID
				}
				if orgIDParam != "" {
					rctx := chi.NewRouteContext()
					rctx.URLParams.Add("org_id", orgIDParam)
					req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				}

				w := httptest.NewRecorder()

				repo := project_repository.NewProjectRepository(q)
				handler := NewProjectHandler(repo).Post()
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
		orgID := createTestOrg(t, q)

		req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("org_id", orgID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		repo := project_repository.NewProjectRepository(q)
		handler := NewProjectHandler(repo).Post()
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

func createTestOrg(t *testing.T, q db.Querier) string {
	t.Helper()
	orgRepo := organization_repository.NewOrganizationRepository(q)

	cmd := organization.NewOrganizationCmd(
		organization.OrganizationName("Test Org"),
		organization.OrganizationSlug("test-org"),
		organization.PlanFree,
	)

	org, err := orgRepo.Create(t.Context(), cmd)
	if err != nil {
		t.Fatalf("failed to create test org: %v", err)
	}
	return org.ID.String()
}
