package organizations

import (
	"bytes"
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

func TestPutHandler(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
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
			// 正常系 - update all fields
			{
				testName: "update all fields",
				reqBody:  map[string]string{"name": "Updated Org", "slug": "updated-org", "plan": "pro"},
				expected: expected{statusCode: http.StatusOK, name: "Updated Org", slug: "updated-org", plan: "pro"},
			},
			// 正常系 - update name only (partial update)
			{
				testName: "update name only",
				reqBody:  map[string]string{"name": "New Name Only"},
				expected: expected{statusCode: http.StatusOK, name: "New Name Only", slug: "original-org", plan: "free"},
			},
			// 正常系 - update slug only
			{
				testName: "update slug only",
				reqBody:  map[string]string{"slug": "new-slug"},
				expected: expected{statusCode: http.StatusOK, name: "Original Org", slug: "new-slug", plan: "free"},
			},
			// 正常系 - update plan only
			{
				testName: "update plan only",
				reqBody:  map[string]string{"plan": "enterprise"},
				expected: expected{statusCode: http.StatusOK, name: "Original Org", slug: "original-org", plan: "enterprise"},
			},
			// 特殊文字
			{
				testName: "update name with Japanese",
				reqBody:  map[string]string{"name": "更新された組織"},
				expected: expected{statusCode: http.StatusOK, name: "更新された組織", slug: "original-org", plan: "free"},
			},
			// 特殊文字 (emoji)
			{
				testName: "update name with emoji",
				reqBody:  map[string]string{"name": "Updated 🎉"},
				expected: expected{statusCode: http.StatusOK, name: "Updated 🎉", slug: "original-org", plan: "free"},
			},
			// 境界値
			{
				testName: "update name to min length",
				reqBody:  map[string]string{"name": "AB"},
				expected: expected{statusCode: http.StatusOK, name: "AB", slug: "original-org", plan: "free"},
			},
			// Null/Nil - empty body keeps existing values
			{
				testName: "empty body keeps existing values",
				reqBody:  map[string]string{},
				expected: expected{statusCode: http.StatusOK, name: "Original Org", slug: "original-org", plan: "free"},
			},
			// 正常系 - all plan values
			{
				testName: "update plan to team",
				reqBody:  map[string]string{"plan": "team"},
				expected: expected{statusCode: http.StatusOK, name: "Original Org", slug: "original-org", plan: "team"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				repo := organization_repository.NewOrganizationRepository(q)

				cmd := organization.NewOrganizationCmd(
					organization.OrganizationName("Original Org"),
					organization.OrganizationSlug("original-org"),
					organization.PlanFree,
				)
				_, err := repo.Create(t.Context(), cmd)
				if err != nil {
					t.Fatalf("failed to create test org: %v", err)
				}

				jsonBody, err := json.Marshal(tt.reqBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}

				req := httptest.NewRequest(http.MethodPut, "/organizations/original-org", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_slug", "original-org")
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()
				handler := NewOrganizationHandler(repo).Put()
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
			// 異常系 - invalid field values
			{testName: "name too short", reqBody: map[string]string{"name": "A"}},
			{testName: "slug too short", reqBody: map[string]string{"slug": "a"}},
			{testName: "invalid plan", reqBody: map[string]string{"plan": "invalid"}},
			// 境界値 - single char slug
			{testName: "slug single char", reqBody: map[string]string{"slug": "x"}},
			// 特殊文字 - invalid slug characters
			{testName: "slug with uppercase", reqBody: map[string]string{"slug": "Invalid-Slug"}},
			{testName: "slug with spaces", reqBody: map[string]string{"slug": "bad slug"}},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				repo := organization_repository.NewOrganizationRepository(q)

				cmd := organization.NewOrganizationCmd(
					organization.OrganizationName("Test Org"),
					organization.OrganizationSlug("test-org"),
					organization.PlanFree,
				)
				_, err := repo.Create(t.Context(), cmd)
				if err != nil {
					t.Fatalf("failed to create test org: %v", err)
				}

				jsonBody, err := json.Marshal(tt.reqBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}

				req := httptest.NewRequest(http.MethodPut, "/organizations/test-org", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_slug", "test-org")
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()
				handler := NewOrganizationHandler(repo).Put()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode != http.StatusBadRequest {
					t.Errorf("expected status 400, got %d", resp.StatusCode)
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
			{testName: "non-existent organization", slug: "non-existent-org"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				repo := organization_repository.NewOrganizationRepository(q)

				body := map[string]string{"name": "Updated"}
				jsonBody, err := json.Marshal(body)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}

				req := httptest.NewRequest(http.MethodPut, "/organizations/"+tt.slug, bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_slug", tt.slug)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()
				handler := NewOrganizationHandler(repo).Put()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(http.StatusNotFound, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("400 Invalid JSON", func(t *testing.T) {
		q := setupTestHandler(t)
		repo := organization_repository.NewOrganizationRepository(q)

		cmd := organization.NewOrganizationCmd(
			organization.OrganizationName("Test Org"),
			organization.OrganizationSlug("test-org"),
			organization.PlanFree,
		)
		_, err := repo.Create(t.Context(), cmd)
		if err != nil {
			t.Fatalf("failed to create test org: %v", err)
		}

		req := httptest.NewRequest(http.MethodPut, "/organizations/test-org", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("org_slug", "test-org")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		testutil.SetAuthHeader(req)

		w := httptest.NewRecorder()
		handler := NewOrganizationHandler(repo).Put()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("400 Invalid Slug Param", func(t *testing.T) {
		tests := []struct {
			testName string
			slug     string
		}{
			// 空文字
			{testName: "empty slug param", slug: ""},
			// 境界値
			{testName: "slug param too short", slug: "a"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)
				repo := organization_repository.NewOrganizationRepository(q)

				body := map[string]string{"name": "Updated"}
				jsonBody, err := json.Marshal(body)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}

				req := httptest.NewRequest(http.MethodPut, "/organizations/"+tt.slug, bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_slug", tt.slug)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()
				handler := NewOrganizationHandler(repo).Put()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode != http.StatusBadRequest {
					t.Errorf("expected status 400, got %d", resp.StatusCode)
				}
			})
		}
	})
}
