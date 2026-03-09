package organizations

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"utils/testutil"

	"api/src/domain/organization"
	"api/src/infra/rds/organization_repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
)

func TestGetHandler(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		type expected struct {
			statusCode int
			name       string
			slug       string
			plan       string
		}

		tests := []struct {
			testName string
			orgName  string
			orgSlug  string
			expected expected
		}{
			// 正常系
			{
				testName: "get organization by slug",
				orgName:  "My Organization",
				orgSlug:  "my-org",
				expected: expected{statusCode: http.StatusOK, name: "My Organization", slug: "my-org", plan: "free"},
			},
			// 特殊文字
			{
				testName: "get organization with Japanese name",
				orgName:  "株式会社テスト",
				orgSlug:  "test-corp",
				expected: expected{statusCode: http.StatusOK, name: "株式会社テスト", slug: "test-corp", plan: "free"},
			},
			// 境界値
			{
				testName: "get organization with min length slug",
				orgName:  "AB",
				orgSlug:  "ab",
				expected: expected{statusCode: http.StatusOK, name: "AB", slug: "ab", plan: "free"},
			},
			// 特殊文字 (emoji in name)
			{
				testName: "get organization with emoji in name",
				orgName:  "Org 🚀",
				orgSlug:  "org-rocket",
				expected: expected{statusCode: http.StatusOK, name: "Org 🚀", slug: "org-rocket", plan: "free"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				repo := organization_repository.NewOrganizationRepository(q)

				cmd := organization.NewOrganizationCmd(
					organization.OrganizationName(tt.orgName),
					organization.OrganizationSlug(tt.orgSlug),
					organization.PlanFree,
				)
				_, err := repo.Create(t.Context(), cmd)
				if err != nil {
					t.Fatalf("failed to create test org: %v", err)
				}

				req := httptest.NewRequest(http.MethodGet, "/organizations/"+tt.orgSlug, nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_slug", tt.orgSlug)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()
				handler := NewOrganizationHandler(repo).Get()
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

	t.Run("404 Not Found", func(t *testing.T) {
		tests := []struct {
			testName string
			slug     string
		}{
			// 異常系
			{testName: "non-existent slug", slug: "non-existent-org"},
			// 境界値
			{testName: "min length non-existent slug", slug: "zz"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				repo := organization_repository.NewOrganizationRepository(q)

				req := httptest.NewRequest(http.MethodGet, "/organizations/"+tt.slug, nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_slug", tt.slug)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()
				handler := NewOrganizationHandler(repo).Get()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(http.StatusNotFound, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		tests := []struct {
			testName string
			slug     string
		}{
			// 異常系 - invalid slug format
			{testName: "slug too short", slug: "a"},
			// 空文字
			{testName: "empty slug", slug: ""},
			// 特殊文字
			{testName: "slug with special chars", slug: "org@slug"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				repo := organization_repository.NewOrganizationRepository(q)

				req := httptest.NewRequest(http.MethodGet, "/organizations/test", nil)
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_slug", tt.slug)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()
				handler := NewOrganizationHandler(repo).Get()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode != http.StatusBadRequest {
					t.Errorf("expected status 400, got %d", resp.StatusCode)
				}
			})
		}
	})
}
