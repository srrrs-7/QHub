package industries

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"api/src/infra/rds/consulting_repository"

	"utils/db/db"
	"utils/testutil"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func setupTestHandler(t *testing.T) db.Querier {
	t.Helper()
	return testutil.SetupTestTx(t)
}

func createIndustryViaHandler(t *testing.T, q db.Querier, slug, name, description string, knowledgeBase, complianceRules json.RawMessage) industryConfigResponse {
	t.Helper()

	repo := consulting_repository.NewIndustryConfigRepository(q)
	handler := NewIndustryHandler(repo, q)

	body := map[string]any{
		"slug":        slug,
		"name":        name,
		"description": description,
	}
	if knowledgeBase != nil {
		body["knowledge_base"] = knowledgeBase
	}
	if complianceRules != nil {
		body["compliance_rules"] = complianceRules
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/industries", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	testutil.SetAuthHeader(req)

	w := httptest.NewRecorder()
	handler.Post().ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusCreated {
		t.Fatalf("setup: expected 201, got %d: %s", w.Result().StatusCode, w.Body.String())
	}

	var resp industryConfigResponse
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode setup response: %v", err)
	}
	return resp
}

// ---------- Post() ----------

func TestPostHandler(t *testing.T) {
	t.Run("201 Created", func(t *testing.T) {
		type expected struct {
			statusCode  int
			slug        string
			name        string
			description string
		}

		tests := []struct {
			testName        string
			reqBody         map[string]any
			expected        expected
		}{
			// 正常系
			{
				testName: "create industry config with all fields",
				reqBody: map[string]any{
					"slug":             "healthcare",
					"name":             "Healthcare",
					"description":      "Healthcare industry config",
					"knowledge_base":   map[string]string{"topic": "health"},
					"compliance_rules": map[string]any{"rules": []any{}},
				},
				expected: expected{statusCode: http.StatusCreated, slug: "healthcare", name: "Healthcare", description: "Healthcare industry config"},
			},
			{
				testName: "create industry config without optional fields",
				reqBody: map[string]any{
					"slug": "finance",
					"name": "Finance",
				},
				expected: expected{statusCode: http.StatusCreated, slug: "finance", name: "Finance", description: ""},
			},
			// 特殊文字
			{
				testName: "create with Japanese name",
				reqBody: map[string]any{
					"slug": "jp-healthcare",
					"name": "医療業界",
					"description": "医療業界の設定",
				},
				expected: expected{statusCode: http.StatusCreated, slug: "jp-healthcare", name: "医療業界", description: "医療業界の設定"},
			},
			{
				testName: "create with emoji in name",
				reqBody: map[string]any{
					"slug": "emoji-industry",
					"name": "Health Care",
				},
				expected: expected{statusCode: http.StatusCreated, slug: "emoji-industry", name: "Health Care"},
			},
			// 境界値
			{
				testName: "create with min length slug (2 chars)",
				reqBody: map[string]any{
					"slug": "ab",
					"name": "Min Slug",
				},
				expected: expected{statusCode: http.StatusCreated, slug: "ab", name: "Min Slug"},
			},
			{
				testName: "create with max length slug (50 chars, DB limit)",
				reqBody: map[string]any{
					"slug": strings.Repeat("a", 50),
					"name": "Max Slug",
				},
				expected: expected{statusCode: http.StatusCreated, slug: strings.Repeat("a", 50), name: "Max Slug"},
			},
			{
				testName: "create with min length name (1 char)",
				reqBody: map[string]any{
					"slug": "min-name",
					"name": "A",
				},
				expected: expected{statusCode: http.StatusCreated, slug: "min-name", name: "A"},
			},
			{
				testName: "create with max length name (100 chars, DB limit)",
				reqBody: map[string]any{
					"slug": "max-name",
					"name": strings.Repeat("N", 100),
				},
				expected: expected{statusCode: http.StatusCreated, slug: "max-name", name: strings.Repeat("N", 100)},
			},
			// Null/Nil - null knowledge_base and compliance_rules
			{
				testName: "create with null knowledge_base and compliance_rules",
				reqBody: map[string]any{
					"slug": "null-fields",
					"name": "Null Fields",
				},
				expected: expected{statusCode: http.StatusCreated, slug: "null-fields", name: "Null Fields"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)

				jsonBody, err := json.Marshal(tt.reqBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}

				req := httptest.NewRequest(http.MethodPost, "/industries", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()

				repo := consulting_repository.NewIndustryConfigRepository(q)
				handler := NewIndustryHandler(repo, q).Post()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var body industryConfigResponse
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(tt.expected.slug, body.Slug); diff != "" {
					t.Errorf("slug mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expected.name, body.Name); diff != "" {
					t.Errorf("name mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(tt.expected.description, body.Description); diff != "" {
					t.Errorf("description mismatch (-want +got):\n%s", diff)
				}
				if body.ID == "" {
					t.Error("expected non-empty ID")
				}
				if body.CreatedAt == "" {
					t.Error("expected non-empty created_at")
				}
				if body.UpdatedAt == "" {
					t.Error("expected non-empty updated_at")
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
			{testName: "missing slug", reqBody: map[string]any{"name": "Healthcare"}},
			{testName: "missing name", reqBody: map[string]any{"slug": "healthcare"}},
			{testName: "missing both slug and name", reqBody: map[string]any{}},
			// 空文字
			{testName: "empty slug", reqBody: map[string]any{"slug": "", "name": "Healthcare"}},
			{testName: "empty name", reqBody: map[string]any{"slug": "healthcare", "name": ""}},
			// 境界値
			{testName: "slug too short (1 char)", reqBody: map[string]any{"slug": "a", "name": "Healthcare"}},
			{testName: "slug too long (81 chars)", reqBody: map[string]any{"slug": strings.Repeat("a", 81), "name": "Healthcare"}},
			{testName: "name too long (201 chars)", reqBody: map[string]any{"slug": "healthcare", "name": strings.Repeat("N", 201)}},
			{testName: "description too long (1001 chars)", reqBody: map[string]any{"slug": "healthcare", "name": "Healthcare", "description": strings.Repeat("D", 1001)}},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := setupTestHandler(t)

				jsonBody, err := json.Marshal(tt.reqBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}

				req := httptest.NewRequest(http.MethodPost, "/industries", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()

				repo := consulting_repository.NewIndustryConfigRepository(q)
				handler := NewIndustryHandler(repo, q).Post()
				handler.ServeHTTP(w, req)

				if w.Result().StatusCode != http.StatusBadRequest {
					t.Errorf("expected status 400, got %d: %s", w.Result().StatusCode, w.Body.String())
				}
			})
		}
	})

	t.Run("400 Invalid JSON", func(t *testing.T) {
		q := setupTestHandler(t)

		req := httptest.NewRequest(http.MethodPost, "/industries", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		w := httptest.NewRecorder()

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).Post()
		handler.ServeHTTP(w, req)

		if w.Result().StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Result().StatusCode)
		}
	})
}

// ---------- List() ----------

func TestListHandler(t *testing.T) {
	// 正常系 - empty list
	t.Run("200 OK empty list", func(t *testing.T) {
		q := setupTestHandler(t)

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).List()

		req := httptest.NewRequest(http.MethodGet, "/industries", nil)
		testutil.SetAuthHeader(req)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if diff := cmp.Diff(http.StatusOK, w.Result().StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var body []industryConfigResponse
		if err := json.NewDecoder(w.Result().Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(body) != 0 {
			t.Errorf("expected empty list, got %d items", len(body))
		}
	})

	// 正常系 - list with items
	t.Run("200 OK with items", func(t *testing.T) {
		q := setupTestHandler(t)

		createIndustryViaHandler(t, q, "healthcare", "Healthcare", "Health desc", nil, nil)
		createIndustryViaHandler(t, q, "finance", "Finance", "Finance desc", nil, nil)

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).List()

		req := httptest.NewRequest(http.MethodGet, "/industries", nil)
		testutil.SetAuthHeader(req)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if diff := cmp.Diff(http.StatusOK, w.Result().StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var body []industryConfigResponse
		if err := json.NewDecoder(w.Result().Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(body) != 2 {
			t.Errorf("expected 2 items, got %d", len(body))
		}
	})

	// 特殊文字 - items with unicode names
	t.Run("200 OK with unicode names", func(t *testing.T) {
		q := setupTestHandler(t)

		createIndustryViaHandler(t, q, "jp-test", "医療業界テスト", "テスト説明", nil, nil)

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).List()

		req := httptest.NewRequest(http.MethodGet, "/industries", nil)
		testutil.SetAuthHeader(req)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		var body []industryConfigResponse
		if err := json.NewDecoder(w.Result().Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(body) != 1 {
			t.Fatalf("expected 1 item, got %d", len(body))
		}
		if diff := cmp.Diff("医療業界テスト", body[0].Name); diff != "" {
			t.Errorf("name mismatch (-want +got):\n%s", diff)
		}
	})
}

// ---------- GetBySlug() ----------

func TestGetBySlugHandler(t *testing.T) {
	// 正常系
	t.Run("200 OK", func(t *testing.T) {
		q := setupTestHandler(t)

		created := createIndustryViaHandler(t, q, "healthcare", "Healthcare", "Health desc", json.RawMessage(`{"topic":"health"}`), json.RawMessage(`{"rules":[]}`))

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).GetBySlug()

		req := httptest.NewRequest(http.MethodGet, "/industries/healthcare", nil)
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("slug", "healthcare")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if diff := cmp.Diff(http.StatusOK, w.Result().StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var body industryConfigResponse
		if err := json.NewDecoder(w.Result().Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if diff := cmp.Diff(created.ID, body.ID); diff != "" {
			t.Errorf("id mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff("healthcare", body.Slug); diff != "" {
			t.Errorf("slug mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff("Healthcare", body.Name); diff != "" {
			t.Errorf("name mismatch (-want +got):\n%s", diff)
		}
	})

	// 異常系 - not found
	t.Run("404 Not Found", func(t *testing.T) {
		q := setupTestHandler(t)

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).GetBySlug()

		req := httptest.NewRequest(http.MethodGet, "/industries/nonexistent", nil)
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("slug", "nonexistent")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if diff := cmp.Diff(http.StatusNotFound, w.Result().StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}
	})

	// 空文字 - empty slug
	t.Run("404 empty slug", func(t *testing.T) {
		q := setupTestHandler(t)

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).GetBySlug()

		req := httptest.NewRequest(http.MethodGet, "/industries/", nil)
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("slug", "")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		// Empty slug should result in not found
		if w.Result().StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Result().StatusCode)
		}
	})

	// 特殊文字 - SQL injection attempt in slug
	t.Run("404 SQL injection slug", func(t *testing.T) {
		q := setupTestHandler(t)

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).GetBySlug()

		req := httptest.NewRequest(http.MethodGet, "/industries/test", nil)
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("slug", "'; DROP TABLE industry_configs;--")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Result().StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Result().StatusCode)
		}
	})
}

// ---------- PutBySlug() ----------

func TestPutBySlugHandler(t *testing.T) {
	// 正常系 - update name and description
	t.Run("200 OK update name and description", func(t *testing.T) {
		q := setupTestHandler(t)

		createIndustryViaHandler(t, q, "healthcare", "Healthcare", "Old desc", nil, nil)

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).PutBySlug()

		reqBody := map[string]any{
			"name":        "Healthcare Updated",
			"description": "New description",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/industries/healthcare", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("slug", "healthcare")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if diff := cmp.Diff(http.StatusOK, w.Result().StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var body industryConfigResponse
		if err := json.NewDecoder(w.Result().Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if diff := cmp.Diff("Healthcare Updated", body.Name); diff != "" {
			t.Errorf("name mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff("New description", body.Description); diff != "" {
			t.Errorf("description mismatch (-want +got):\n%s", diff)
		}
	})

	// 正常系 - update compliance_rules
	t.Run("200 OK update compliance rules", func(t *testing.T) {
		q := setupTestHandler(t)

		createIndustryViaHandler(t, q, "finance", "Finance", "", nil, json.RawMessage(`{"rules":[]}`))

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).PutBySlug()

		newRules := json.RawMessage(`{"rules":[{"keyword":"password","message":"Do not include passwords"}]}`)
		reqBody := map[string]any{
			"name":             "Finance",
			"compliance_rules": newRules,
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/industries/finance", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("slug", "finance")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if diff := cmp.Diff(http.StatusOK, w.Result().StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var body industryConfigResponse
		if err := json.NewDecoder(w.Result().Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		// Verify compliance_rules was updated
		var parsed struct {
			Rules []struct {
				Keyword string `json:"keyword"`
				Message string `json:"message"`
			} `json:"rules"`
		}
		if err := json.Unmarshal(body.ComplianceRules, &parsed); err != nil {
			t.Fatalf("failed to unmarshal compliance_rules: %v", err)
		}
		if len(parsed.Rules) != 1 {
			t.Fatalf("expected 1 rule, got %d", len(parsed.Rules))
		}
		if diff := cmp.Diff("password", parsed.Rules[0].Keyword); diff != "" {
			t.Errorf("keyword mismatch (-want +got):\n%s", diff)
		}
	})

	// 特殊文字 - unicode update
	t.Run("200 OK update with Japanese name", func(t *testing.T) {
		q := setupTestHandler(t)

		createIndustryViaHandler(t, q, "jp-update", "Original", "", nil, nil)

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).PutBySlug()

		reqBody := map[string]any{
			"name":        "更新された名前",
			"description": "日本語の説明文",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/industries/jp-update", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("slug", "jp-update")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if diff := cmp.Diff(http.StatusOK, w.Result().StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var body industryConfigResponse
		if err := json.NewDecoder(w.Result().Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if diff := cmp.Diff("更新された名前", body.Name); diff != "" {
			t.Errorf("name mismatch (-want +got):\n%s", diff)
		}
	})

	// 異常系 - not found
	t.Run("404 Not Found", func(t *testing.T) {
		q := setupTestHandler(t)

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).PutBySlug()

		reqBody := map[string]any{"name": "Updated"}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/industries/nonexistent", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("slug", "nonexistent")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if diff := cmp.Diff(http.StatusNotFound, w.Result().StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}
	})

	// 異常系 - invalid JSON body
	t.Run("400 Invalid JSON", func(t *testing.T) {
		q := setupTestHandler(t)

		createIndustryViaHandler(t, q, "bad-json", "Bad JSON Test", "", nil, nil)

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).PutBySlug()

		req := httptest.NewRequest(http.MethodPut, "/industries/bad-json", bytes.NewReader([]byte("not json")))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("slug", "bad-json")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Result().StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Result().StatusCode)
		}
	})

	// 境界値 - name too long
	t.Run("400 name too long", func(t *testing.T) {
		q := setupTestHandler(t)

		createIndustryViaHandler(t, q, "long-name", "Original Name", "", nil, nil)

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).PutBySlug()

		reqBody := map[string]any{"name": strings.Repeat("X", 201)}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/industries/long-name", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("slug", "long-name")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Result().StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d: %s", w.Result().StatusCode, w.Body.String())
		}
	})

	// 境界値 - description too long
	t.Run("400 description too long", func(t *testing.T) {
		q := setupTestHandler(t)

		createIndustryViaHandler(t, q, "long-desc", "Original", "", nil, nil)

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).PutBySlug()

		reqBody := map[string]any{"name": "Valid", "description": strings.Repeat("D", 1001)}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/industries/long-desc", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("slug", "long-desc")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Result().StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d: %s", w.Result().StatusCode, w.Body.String())
		}
	})
}

// ---------- ListBenchmarks() ----------

func TestListBenchmarksHandler(t *testing.T) {
	// 正常系 - empty benchmarks
	t.Run("200 OK empty benchmarks", func(t *testing.T) {
		q := setupTestHandler(t)

		createIndustryViaHandler(t, q, "healthcare", "Healthcare", "", nil, nil)

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).ListBenchmarks()

		req := httptest.NewRequest(http.MethodGet, "/industries/healthcare/benchmarks", nil)
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("slug", "healthcare")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if diff := cmp.Diff(http.StatusOK, w.Result().StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var body []benchmarkResponse
		if err := json.NewDecoder(w.Result().Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(body) != 0 {
			t.Errorf("expected empty benchmarks, got %d", len(body))
		}
	})

	// 正常系 - with benchmarks
	t.Run("200 OK with benchmarks", func(t *testing.T) {
		q := setupTestHandler(t)

		created := createIndustryViaHandler(t, q, "finance", "Finance", "", nil, nil)

		// Insert a benchmark directly via Querier
		_, err := q.CreatePlatformBenchmark(context.Background(), db.CreatePlatformBenchmarkParams{
			IndustryConfigID:  mustParseUUID(t, created.ID),
			Period:            "2026-03",
			AvgQualityScore:   sql.NullString{String: "8.55", Valid: true},
			AvgLatencyMs:      sql.NullInt32{Int32: 150, Valid: true},
			AvgCostPerRequest: sql.NullString{String: "0.050000", Valid: true},
			TotalExecutions:   1000,
			P50Quality:        sql.NullString{String: "8.30", Valid: true},
			P90Quality:        sql.NullString{String: "9.20", Valid: true},
			OptInCount:        5,
		})
		if err != nil {
			t.Fatalf("failed to create benchmark: %v", err)
		}

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).ListBenchmarks()

		req := httptest.NewRequest(http.MethodGet, "/industries/finance/benchmarks", nil)
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("slug", "finance")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if diff := cmp.Diff(http.StatusOK, w.Result().StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var body []benchmarkResponse
		if err := json.NewDecoder(w.Result().Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(body) != 1 {
			t.Fatalf("expected 1 benchmark, got %d", len(body))
		}

		if diff := cmp.Diff("2026-03", body[0].Period); diff != "" {
			t.Errorf("period mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff("8.55", body[0].AvgQualityScore); diff != "" {
			t.Errorf("avg_quality_score mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff(150, body[0].AvgLatencyMs); diff != "" {
			t.Errorf("avg_latency_ms mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff(int64(1000), body[0].TotalExecutions); diff != "" {
			t.Errorf("total_executions mismatch (-want +got):\n%s", diff)
		}
	})

	// 異常系 - not found slug
	t.Run("404 Not Found", func(t *testing.T) {
		q := setupTestHandler(t)

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).ListBenchmarks()

		req := httptest.NewRequest(http.MethodGet, "/industries/nonexistent/benchmarks", nil)
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("slug", "nonexistent")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if diff := cmp.Diff(http.StatusNotFound, w.Result().StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}
	})

	// Null/Nil - benchmark with null optional fields
	t.Run("200 OK benchmark with null fields", func(t *testing.T) {
		q := setupTestHandler(t)

		created := createIndustryViaHandler(t, q, "null-bench", "Null Bench", "", nil, nil)

		_, err := q.CreatePlatformBenchmark(context.Background(), db.CreatePlatformBenchmarkParams{
			IndustryConfigID: mustParseUUID(t, created.ID),
			Period:           "2026-01",
			TotalExecutions:  0,
			OptInCount:       0,
			// All nullable fields left as zero values (invalid)
		})
		if err != nil {
			t.Fatalf("failed to create benchmark: %v", err)
		}

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).ListBenchmarks()

		req := httptest.NewRequest(http.MethodGet, "/industries/null-bench/benchmarks", nil)
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("slug", "null-bench")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if diff := cmp.Diff(http.StatusOK, w.Result().StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var body []benchmarkResponse
		if err := json.NewDecoder(w.Result().Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(body) != 1 {
			t.Fatalf("expected 1 benchmark, got %d", len(body))
		}

		// Null fields should be empty/zero
		if body[0].AvgQualityScore != "" {
			t.Errorf("expected empty avg_quality_score, got %q", body[0].AvgQualityScore)
		}
		if body[0].AvgLatencyMs != 0 {
			t.Errorf("expected 0 avg_latency_ms, got %d", body[0].AvgLatencyMs)
		}
		if body[0].AvgCostPerRequest != "" {
			t.Errorf("expected empty avg_cost_per_request, got %q", body[0].AvgCostPerRequest)
		}
	})
}

// ---------- ComplianceCheck() ----------

func TestComplianceCheckHandler(t *testing.T) {
	// 正常系 - no violations
	t.Run("200 OK compliant", func(t *testing.T) {
		q := setupTestHandler(t)

		rules := json.RawMessage(`{"rules":[{"keyword":"password","message":"Do not include passwords"},{"keyword":"ssn","message":"Do not include SSN"}]}`)
		createIndustryViaHandler(t, q, "healthcare", "Healthcare", "", nil, rules)

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).ComplianceCheck()

		reqBody := map[string]string{"content": "This is a safe healthcare document about patient wellness."}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/industries/healthcare/compliance", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("slug", "healthcare")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if diff := cmp.Diff(http.StatusOK, w.Result().StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var body complianceCheckResponse
		if err := json.NewDecoder(w.Result().Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !body.Compliant {
			t.Error("expected compliant to be true")
		}
		if len(body.Violations) != 0 {
			t.Errorf("expected 0 violations, got %d", len(body.Violations))
		}
	})

	// 正常系 - with violations
	t.Run("200 OK with violations", func(t *testing.T) {
		q := setupTestHandler(t)

		rules := json.RawMessage(`{"rules":[{"keyword":"password","message":"Do not include passwords"},{"keyword":"ssn","message":"Do not include SSN"}]}`)
		createIndustryViaHandler(t, q, "finance", "Finance", "", nil, rules)

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).ComplianceCheck()

		reqBody := map[string]string{"content": "Please enter your password and SSN for verification."}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/industries/finance/compliance", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("slug", "finance")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if diff := cmp.Diff(http.StatusOK, w.Result().StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var body complianceCheckResponse
		if err := json.NewDecoder(w.Result().Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if body.Compliant {
			t.Error("expected compliant to be false")
		}
		if len(body.Violations) != 2 {
			t.Fatalf("expected 2 violations, got %d", len(body.Violations))
		}
	})

	// 正常系 - case insensitive matching
	t.Run("200 OK case insensitive violation", func(t *testing.T) {
		q := setupTestHandler(t)

		rules := json.RawMessage(`{"rules":[{"keyword":"SECRET","message":"Do not include secrets"}]}`)
		createIndustryViaHandler(t, q, "case-test", "Case Test", "", nil, rules)

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).ComplianceCheck()

		reqBody := map[string]string{"content": "This contains a secret value."}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/industries/case-test/compliance", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("slug", "case-test")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		var body complianceCheckResponse
		if err := json.NewDecoder(w.Result().Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if body.Compliant {
			t.Error("expected compliant to be false (case insensitive)")
		}
		if len(body.Violations) != 1 {
			t.Fatalf("expected 1 violation, got %d", len(body.Violations))
		}
		if diff := cmp.Diff("SECRET", body.Violations[0].Rule); diff != "" {
			t.Errorf("rule mismatch (-want +got):\n%s", diff)
		}
	})

	// 異常系 - not found slug
	t.Run("404 Not Found", func(t *testing.T) {
		q := setupTestHandler(t)

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).ComplianceCheck()

		reqBody := map[string]string{"content": "test content"}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/industries/nonexistent/compliance", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("slug", "nonexistent")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if diff := cmp.Diff(http.StatusNotFound, w.Result().StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}
	})

	// 異常系 - missing content
	t.Run("400 missing content", func(t *testing.T) {
		q := setupTestHandler(t)

		createIndustryViaHandler(t, q, "missing-content", "Missing Content", "", nil, nil)

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).ComplianceCheck()

		reqBody := map[string]string{}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/industries/missing-content/compliance", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("slug", "missing-content")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Result().StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Result().StatusCode)
		}
	})

	// 異常系 - invalid JSON
	t.Run("400 invalid JSON", func(t *testing.T) {
		q := setupTestHandler(t)

		createIndustryViaHandler(t, q, "invalid-json", "Invalid JSON", "", nil, nil)

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).ComplianceCheck()

		req := httptest.NewRequest(http.MethodPost, "/industries/invalid-json/compliance", bytes.NewReader([]byte("not json")))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("slug", "invalid-json")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Result().StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Result().StatusCode)
		}
	})

	// Null/Nil - null compliance_rules (should be compliant)
	t.Run("200 OK null compliance rules is compliant", func(t *testing.T) {
		q := setupTestHandler(t)

		createIndustryViaHandler(t, q, "no-rules", "No Rules", "", nil, nil)

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).ComplianceCheck()

		reqBody := map[string]string{"content": "password and ssn are here"}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/industries/no-rules/compliance", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("slug", "no-rules")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if diff := cmp.Diff(http.StatusOK, w.Result().StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var body complianceCheckResponse
		if err := json.NewDecoder(w.Result().Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !body.Compliant {
			t.Error("expected compliant to be true when no rules defined")
		}
	})

	// 特殊文字 - unicode content
	t.Run("200 OK unicode content compliance check", func(t *testing.T) {
		q := setupTestHandler(t)

		rules := json.RawMessage(`{"rules":[{"keyword":"パスワード","message":"パスワードを含めないでください"}]}`)
		createIndustryViaHandler(t, q, "jp-compliance", "JP Compliance", "", nil, rules)

		repo := consulting_repository.NewIndustryConfigRepository(q)
		handler := NewIndustryHandler(repo, q).ComplianceCheck()

		reqBody := map[string]string{"content": "このドキュメントにはパスワードが含まれています。"}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/industries/jp-compliance/compliance", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("slug", "jp-compliance")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		var body complianceCheckResponse
		if err := json.NewDecoder(w.Result().Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if body.Compliant {
			t.Error("expected compliant to be false for Japanese keyword match")
		}
		if len(body.Violations) != 1 {
			t.Fatalf("expected 1 violation, got %d", len(body.Violations))
		}
		if diff := cmp.Diff("パスワードを含めないでください", body.Violations[0].Message); diff != "" {
			t.Errorf("message mismatch (-want +got):\n%s", diff)
		}
	})
}

// ---------- checkCompliance() unit tests ----------

func TestCheckCompliance(t *testing.T) {
	tests := []struct {
		testName          string
		rulesJSON         json.RawMessage
		content           string
		expectedCompliant bool
		expectedCount     int
	}{
		// 正常系
		{
			testName:          "no violations",
			rulesJSON:         json.RawMessage(`{"rules":[{"keyword":"password","message":"No passwords"}]}`),
			content:           "This is a safe document.",
			expectedCompliant: true,
			expectedCount:     0,
		},
		{
			testName:          "single violation",
			rulesJSON:         json.RawMessage(`{"rules":[{"keyword":"password","message":"No passwords"}]}`),
			content:           "Your password is 1234.",
			expectedCompliant: false,
			expectedCount:     1,
		},
		{
			testName:          "multiple violations",
			rulesJSON:         json.RawMessage(`{"rules":[{"keyword":"password","message":"No passwords"},{"keyword":"secret","message":"No secrets"}]}`),
			content:           "password and secret inside",
			expectedCompliant: false,
			expectedCount:     2,
		},
		// 正常系 - case insensitive
		{
			testName:          "case insensitive keyword in content",
			rulesJSON:         json.RawMessage(`{"rules":[{"keyword":"PASSWORD","message":"No passwords"}]}`),
			content:           "your password is here",
			expectedCompliant: false,
			expectedCount:     1,
		},
		{
			testName:          "case insensitive uppercase content",
			rulesJSON:         json.RawMessage(`{"rules":[{"keyword":"password","message":"No passwords"}]}`),
			content:           "YOUR PASSWORD IS HERE",
			expectedCompliant: false,
			expectedCount:     1,
		},
		// Null/Nil
		{
			testName:          "nil rules JSON",
			rulesJSON:         nil,
			content:           "anything",
			expectedCompliant: true,
			expectedCount:     0,
		},
		// 異常系 - invalid JSON rules
		{
			testName:          "invalid rules JSON",
			rulesJSON:         json.RawMessage(`not valid json`),
			content:           "password",
			expectedCompliant: true,
			expectedCount:     0,
		},
		// 空文字 - empty rules array
		{
			testName:          "empty rules array",
			rulesJSON:         json.RawMessage(`{"rules":[]}`),
			content:           "password",
			expectedCompliant: true,
			expectedCount:     0,
		},
		// 空文字 - empty content
		{
			testName:          "empty content",
			rulesJSON:         json.RawMessage(`{"rules":[{"keyword":"password","message":"No passwords"}]}`),
			content:           "",
			expectedCompliant: true,
			expectedCount:     0,
		},
		// 空文字 - empty keyword in rule
		{
			testName:          "empty keyword skipped",
			rulesJSON:         json.RawMessage(`{"rules":[{"keyword":"","message":"Empty keyword should be skipped"}]}`),
			content:           "anything",
			expectedCompliant: true,
			expectedCount:     0,
		},
		// 境界値 - keyword at start of content
		{
			testName:          "keyword at start",
			rulesJSON:         json.RawMessage(`{"rules":[{"keyword":"danger","message":"Dangerous"}]}`),
			content:           "danger zone ahead",
			expectedCompliant: false,
			expectedCount:     1,
		},
		// 境界値 - keyword at end of content
		{
			testName:          "keyword at end",
			rulesJSON:         json.RawMessage(`{"rules":[{"keyword":"danger","message":"Dangerous"}]}`),
			content:           "beware of danger",
			expectedCompliant: false,
			expectedCount:     1,
		},
		// 特殊文字 - Japanese keywords
		{
			testName:          "Japanese keyword match",
			rulesJSON:         json.RawMessage(`{"rules":[{"keyword":"機密","message":"機密情報を含まないでください"}]}`),
			content:           "この文書には機密情報が含まれています",
			expectedCompliant: false,
			expectedCount:     1,
		},
		// 特殊文字 - keyword with special characters
		{
			testName:          "special chars in keyword",
			rulesJSON:         json.RawMessage(`{"rules":[{"keyword":"$ecret","message":"No dollar secrets"}]}`),
			content:           "The $ecret is here",
			expectedCompliant: false,
			expectedCount:     1,
		},
		// Null/Nil - rules without keyword and message fields
		{
			testName:          "malformed rule missing keyword",
			rulesJSON:         json.RawMessage(`{"rules":[{"message":"No keyword defined"}]}`),
			content:           "password",
			expectedCompliant: true,
			expectedCount:     0,
		},
		// 正常系 - partial match
		{
			testName:          "partial word match",
			rulesJSON:         json.RawMessage(`{"rules":[{"keyword":"pass","message":"Contains pass"}]}`),
			content:           "password",
			expectedCompliant: false,
			expectedCount:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			violations := checkCompliance(tt.rulesJSON, tt.content)

			compliant := len(violations) == 0
			if diff := cmp.Diff(tt.expectedCompliant, compliant); diff != "" {
				t.Errorf("compliant mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expectedCount, len(violations)); diff != "" {
				t.Errorf("violation count mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// ---------- helpers ----------

func mustParseUUID(t *testing.T, s string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("failed to parse UUID %q: %v", s, err)
	}
	return id
}
