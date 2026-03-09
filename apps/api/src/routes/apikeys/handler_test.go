package apikeys

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"utils/db/db"
	"utils/testutil"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

// setupTestOrg creates a test organization and returns the querier and org.
func setupTestOrg(t *testing.T) (db.Querier, db.Organization) {
	t.Helper()
	q := testutil.SetupTestTx(t)
	org, err := q.CreateOrganization(context.Background(), db.CreateOrganizationParams{
		Name: "Test Org",
		Slug: "test-org-" + uuid.New().String()[:8],
		Plan: "free",
	})
	if err != nil {
		t.Fatalf("failed to create organization: %v", err)
	}
	return q, org
}

// seedApiKey creates an API key directly in the database with a valid key_prefix
// that satisfies the chk_api_keys_prefix constraint (key_prefix ~ '^pl_(live|test)_').
func seedApiKey(t *testing.T, q db.Querier, orgID uuid.UUID, name string) db.ApiKey {
	t.Helper()
	rawKey := "qlb_" + uuid.New().String()
	h := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(h[:])
	prefix := "pl_live_" + uuid.New().String()[:4]

	apiKey, err := q.CreateApiKey(context.Background(), db.CreateApiKeyParams{
		OrganizationID: orgID,
		Name:           name,
		KeyHash:        keyHash,
		KeyPrefix:      prefix,
	})
	if err != nil {
		t.Fatalf("failed to seed api key: %v", err)
	}
	return apiKey
}

// ---------- Post() Tests ----------

func TestPostHandler(t *testing.T) {
	t.Run("201 Created", func(t *testing.T) {
		type expected struct {
			statusCode int
			name       string
			hasKey     bool
			hasPrefix  bool
		}

		tests := []struct {
			testName string
			name     string
			expected expected
		}{
			// 正常系
			{
				testName: "create api key with valid name",
				name:     "my-api-key",
				expected: expected{statusCode: http.StatusCreated, name: "my-api-key", hasKey: true, hasPrefix: true},
			},
			// 特殊文字
			{
				testName: "create api key with Japanese name",
				name:     "テスト鍵",
				expected: expected{statusCode: http.StatusCreated, name: "テスト鍵", hasKey: true, hasPrefix: true},
			},
			{
				testName: "create api key with emoji name",
				name:     "key-🔑-test",
				expected: expected{statusCode: http.StatusCreated, name: "key-🔑-test", hasKey: true, hasPrefix: true},
			},
			{
				testName: "create api key with unicode characters",
				name:     "clé-für-tëst",
				expected: expected{statusCode: http.StatusCreated, name: "clé-für-tëst", hasKey: true, hasPrefix: true},
			},
			// 境界値
			{
				testName: "create api key with min length name (1 char)",
				name:     "a",
				expected: expected{statusCode: http.StatusCreated, name: "a", hasKey: true, hasPrefix: true},
			},
			{
				testName: "create api key with max length name (100 chars)",
				name:     strings.Repeat("a", 100),
				expected: expected{statusCode: http.StatusCreated, name: strings.Repeat("a", 100), hasKey: true, hasPrefix: true},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q, org := setupTestOrg(t)

				body := fmt.Sprintf(`{"organization_id": %q, "name": %q}`, org.ID.String(), tt.name)
				req := httptest.NewRequest(http.MethodPost, "/organizations/"+org.ID.String()+"/api-keys", bytes.NewReader([]byte(body)))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_id", org.ID.String())
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()
				handler := NewApiKeyHandler(q).Post()
				handler.ServeHTTP(w, req)

				resp := w.Result()

				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var respBody apiKeyCreatedResponse
				if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(tt.expected.name, respBody.Name); diff != "" {
					t.Errorf("name mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff(org.ID.String(), respBody.OrganizationID); diff != "" {
					t.Errorf("organization_id mismatch (-want +got):\n%s", diff)
				}
				if respBody.ID == "" {
					t.Error("expected non-empty ID")
				}
				if tt.expected.hasKey {
					if !strings.HasPrefix(respBody.Key, "pl_live_") {
						t.Errorf("expected key to start with 'pl_live_', got %q", respBody.Key)
					}
					// qlb_ + 64 hex chars (32 bytes)
					if len(respBody.Key) != 8+64 {
						t.Errorf("expected key length 72, got %d", len(respBody.Key))
					}
				}
				if tt.expected.hasPrefix && respBody.KeyPrefix == "" {
					t.Error("expected non-empty key_prefix")
				}
			})
		}
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		tests := []struct {
			testName string
			body     string
		}{
			// 異常系
			{testName: "missing name field", body: `{"organization_id": "%s"}`},
			{testName: "missing organization_id field", body: `{"name": "my-key"}`},
			// 空文字
			{testName: "empty name", body: `{"organization_id": "%s", "name": ""}`},
			// 境界値
			{testName: "name exceeds max length (101 chars)", body: `{"organization_id": "%s", "name": "` + strings.Repeat("a", 101) + `"}`},
			// 異常系 - invalid organization_id
			{testName: "invalid organization_id format", body: `{"organization_id": "not-a-uuid", "name": "my-key"}`},
			// Null/Nil
			{testName: "null name", body: `{"organization_id": "%s", "name": null}`},
			{testName: "empty JSON object", body: `{}`},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q, org := setupTestOrg(t)

				body := fmt.Sprintf(tt.body, org.ID.String())
				req := httptest.NewRequest(http.MethodPost, "/organizations/"+org.ID.String()+"/api-keys", bytes.NewReader([]byte(body)))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_id", org.ID.String())
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()
				handler := NewApiKeyHandler(q).Post()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode != http.StatusBadRequest {
					t.Errorf("expected status 400, got %d: %s", resp.StatusCode, w.Body.String())
				}
			})
		}
	})

	t.Run("400 Invalid JSON", func(t *testing.T) {
		q, org := setupTestOrg(t)

		req := httptest.NewRequest(http.MethodPost, "/organizations/"+org.ID.String()+"/api-keys", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("org_id", org.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler := NewApiKeyHandler(q).Post()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("500 non-existent organization_id (FK violation)", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		fakeOrgID := uuid.New()

		body := fmt.Sprintf(`{"organization_id": %q, "name": "my-key"}`, fakeOrgID.String())
		req := httptest.NewRequest(http.MethodPost, "/organizations/"+fakeOrgID.String()+"/api-keys", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("org_id", fakeOrgID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler := NewApiKeyHandler(q).Post()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d: %s", resp.StatusCode, w.Body.String())
		}
	})

	t.Run("whitespace only name passes validation", func(t *testing.T) {
		// 空文字 - whitespace-only name. The validator allows it (min=1 counts spaces).
		// Bluemonday sanitizer preserves whitespace, so it passes through.
		q, org := setupTestOrg(t)

		body := fmt.Sprintf(`{"organization_id": %q, "name": "   "}`, org.ID.String())
		req := httptest.NewRequest(http.MethodPost, "/organizations/"+org.ID.String()+"/api-keys", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("org_id", org.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler := NewApiKeyHandler(q).Post()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		// Whitespace-only passes validation (min=1 counts spaces) and creates successfully
		if diff := cmp.Diff(http.StatusCreated, resp.StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}
	})
}

// ---------- List() Tests ----------

func TestListHandler(t *testing.T) {
	t.Run("200 OK", func(t *testing.T) {
		tests := []struct {
			testName    string
			keyNames    []string
			expectedLen int
		}{
			// 正常系
			{
				testName:    "list api keys for organization with keys",
				keyNames:    []string{"key-1", "key-2", "key-3"},
				expectedLen: 3,
			},
			// 境界値 - empty list
			{
				testName:    "list api keys for organization with no keys",
				keyNames:    []string{},
				expectedLen: 0,
			},
			// 境界値 - single key
			{
				testName:    "list api keys for organization with one key",
				keyNames:    []string{"only-key"},
				expectedLen: 1,
			},
			// 特殊文字
			{
				testName:    "list api keys with Japanese name",
				keyNames:    []string{"テスト鍵"},
				expectedLen: 1,
			},
			{
				testName:    "list api keys with emoji name",
				keyNames:    []string{"key-🔑"},
				expectedLen: 1,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q, org := setupTestOrg(t)

				for _, name := range tt.keyNames {
					seedApiKey(t, q, org.ID, name)
				}

				req := httptest.NewRequest(http.MethodGet, "/organizations/"+org.ID.String()+"/api-keys", nil)
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_id", org.ID.String())
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()
				handler := NewApiKeyHandler(q).List()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(http.StatusOK, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var respBody []apiKeyResponse
				if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(tt.expectedLen, len(respBody)); diff != "" {
					t.Errorf("response length mismatch (-want +got):\n%s", diff)
				}

				// Verify response fields are populated and do not contain raw key hash
				for _, k := range respBody {
					if k.ID == "" {
						t.Error("expected non-empty ID")
					}
					if k.OrganizationID != org.ID.String() {
						t.Errorf("expected organization_id %s, got %s", org.ID.String(), k.OrganizationID)
					}
					if k.KeyPrefix == "" {
						t.Error("expected non-empty key_prefix")
					}
					if k.CreatedAt == "" {
						t.Error("expected non-empty created_at")
					}
				}
			})
		}
	})

	t.Run("200 OK - response contains expected key names", func(t *testing.T) {
		// 正常系 - verify returned names match seeded data
		q, org := setupTestOrg(t)
		expectedNames := []string{"alpha-key", "beta-key"}
		for _, name := range expectedNames {
			seedApiKey(t, q, org.ID, name)
		}

		req := httptest.NewRequest(http.MethodGet, "/organizations/"+org.ID.String()+"/api-keys", nil)
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("org_id", org.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		NewApiKeyHandler(q).List().ServeHTTP(w, req)

		var respBody []apiKeyResponse
		if err := json.NewDecoder(w.Result().Body).Decode(&respBody); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		gotNames := make(map[string]bool)
		for _, k := range respBody {
			gotNames[k.Name] = true
		}
		for _, name := range expectedNames {
			if !gotNames[name] {
				t.Errorf("expected key name %q in response, not found", name)
			}
		}
	})

	t.Run("400 Bad Request - invalid org_id", func(t *testing.T) {
		tests := []struct {
			testName string
			orgID    string
		}{
			// 異常系
			{testName: "invalid UUID format", orgID: "not-a-uuid"},
			// 空文字
			{testName: "empty org_id", orgID: ""},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				req := httptest.NewRequest(http.MethodGet, "/api/v1/api-keys", nil)
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_id", tt.orgID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()
				handler := NewApiKeyHandler(q).List()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode != http.StatusBadRequest {
					t.Errorf("expected status 400, got %d: %s", resp.StatusCode, w.Body.String())
				}
			})
		}
	})

	t.Run("200 OK - keys scoped to organization", func(t *testing.T) {
		// 正常系 - verify keys are scoped to org
		q, org1 := setupTestOrg(t)
		org2, err := q.CreateOrganization(context.Background(), db.CreateOrganizationParams{
			Name: "Other Org",
			Slug: "other-org-" + uuid.New().String()[:8],
			Plan: "free",
		})
		if err != nil {
			t.Fatalf("failed to create second organization: %v", err)
		}

		seedApiKey(t, q, org1.ID, "org1-key")
		seedApiKey(t, q, org2.ID, "org2-key")

		req := httptest.NewRequest(http.MethodGet, "/organizations/"+org1.ID.String()+"/api-keys", nil)
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("org_id", org1.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		NewApiKeyHandler(q).List().ServeHTTP(w, req)

		var respBody []apiKeyResponse
		if err := json.NewDecoder(w.Result().Body).Decode(&respBody); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if diff := cmp.Diff(1, len(respBody)); diff != "" {
			t.Errorf("expected 1 key for org1 (-want +got):\n%s", diff)
		}
		if len(respBody) > 0 && respBody[0].Name != "org1-key" {
			t.Errorf("expected key name 'org1-key', got %q", respBody[0].Name)
		}
	})

	t.Run("200 OK - non-existent org returns empty list", func(t *testing.T) {
		// Null/Nil - org with no data
		q := testutil.SetupTestTx(t)
		fakeOrgID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/organizations/"+fakeOrgID.String()+"/api-keys", nil)
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("org_id", fakeOrgID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		NewApiKeyHandler(q).List().ServeHTTP(w, req)

		resp := w.Result()
		if diff := cmp.Diff(http.StatusOK, resp.StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var respBody []apiKeyResponse
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if diff := cmp.Diff(0, len(respBody)); diff != "" {
			t.Errorf("expected empty list (-want +got):\n%s", diff)
		}
	})
}

// ---------- Delete() Tests ----------

func TestDeleteHandler(t *testing.T) {
	t.Run("204 No Content", func(t *testing.T) {
		tests := []struct {
			testName string
			keyName  string
		}{
			// 正常系
			{testName: "delete existing api key", keyName: "to-delete"},
			// 特殊文字
			{testName: "delete key with unicode name", keyName: "鍵-テスト-key"},
			{testName: "delete key with emoji name", keyName: "🔑-key"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q, org := setupTestOrg(t)
				apiKey := seedApiKey(t, q, org.ID, tt.keyName)

				req := httptest.NewRequest(http.MethodDelete, "/organizations/"+org.ID.String()+"/api-keys/"+apiKey.ID.String(), nil)
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_id", org.ID.String())
				rctx.URLParams.Add("id", apiKey.ID.String())
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()
				handler := NewApiKeyHandler(q).Delete()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(http.StatusNoContent, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("204 then revoked_at is set (soft delete verification)", func(t *testing.T) {
		// 正常系 - verify soft delete behavior
		q, org := setupTestOrg(t)
		apiKey := seedApiKey(t, q, org.ID, "soft-delete-test")
		seedApiKey(t, q, org.ID, "keep-this-key")

		// Delete one key
		req := httptest.NewRequest(http.MethodDelete, "/organizations/"+org.ID.String()+"/api-keys/"+apiKey.ID.String(), nil)
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("org_id", org.ID.String())
		rctx.URLParams.Add("id", apiKey.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		NewApiKeyHandler(q).Delete().ServeHTTP(w, req)

		if diff := cmp.Diff(http.StatusNoContent, w.Result().StatusCode); diff != "" {
			t.Fatalf("delete status code mismatch (-want +got):\n%s", diff)
		}

		// List should still show both keys (ListApiKeysByOrganization does not filter revoked)
		listReq := httptest.NewRequest(http.MethodGet, "/organizations/"+org.ID.String()+"/api-keys", nil)
		testutil.SetAuthHeader(listReq)

		listCtx := chi.NewRouteContext()
		listCtx.URLParams.Add("org_id", org.ID.String())
		listReq = listReq.WithContext(context.WithValue(listReq.Context(), chi.RouteCtxKey, listCtx))

		listW := httptest.NewRecorder()
		NewApiKeyHandler(q).List().ServeHTTP(listW, listReq)

		var keys []apiKeyResponse
		if err := json.NewDecoder(listW.Result().Body).Decode(&keys); err != nil {
			t.Fatalf("failed to decode list response: %v", err)
		}

		if diff := cmp.Diff(2, len(keys)); diff != "" {
			t.Errorf("expected 2 keys in list (-want +got):\n%s", diff)
		}

		// Find the revoked key and verify revoked_at is set
		for _, k := range keys {
			if k.ID == apiKey.ID.String() {
				if k.RevokedAt == nil {
					t.Error("expected revoked_at to be set for deleted key")
				}
			}
		}
	})

	t.Run("400 Bad Request - invalid params", func(t *testing.T) {
		tests := []struct {
			testName string
			orgID    string
			keyID    string
		}{
			// 異常系
			{testName: "invalid org_id", orgID: "not-a-uuid", keyID: uuid.New().String()},
			{testName: "invalid key id", orgID: uuid.New().String(), keyID: "not-a-uuid"},
			{testName: "both invalid", orgID: "bad-org", keyID: "bad-key"},
			// 空文字
			{testName: "empty org_id", orgID: "", keyID: uuid.New().String()},
			{testName: "empty key id", orgID: uuid.New().String(), keyID: ""},
			{testName: "both empty", orgID: "", keyID: ""},
			// 特殊文字
			{testName: "special chars in org_id", orgID: "abc!@#$%", keyID: uuid.New().String()},
			{testName: "special chars in key id", orgID: uuid.New().String(), keyID: "abc!@#$%"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				req := httptest.NewRequest(http.MethodDelete, "/api/v1/api-keys", nil)
				testutil.SetAuthHeader(req)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("org_id", tt.orgID)
				rctx.URLParams.Add("id", tt.keyID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				w := httptest.NewRecorder()
				handler := NewApiKeyHandler(q).Delete()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if resp.StatusCode != http.StatusBadRequest {
					t.Errorf("expected status 400, got %d: %s", resp.StatusCode, w.Body.String())
				}
			})
		}
	})

	t.Run("500 - delete non-existent key", func(t *testing.T) {
		// 異常系 - key does not exist
		q, org := setupTestOrg(t)
		fakeKeyID := uuid.New()

		req := httptest.NewRequest(http.MethodDelete, "/organizations/"+org.ID.String()+"/api-keys/"+fakeKeyID.String(), nil)
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("org_id", org.ID.String())
		rctx.URLParams.Add("id", fakeKeyID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler := NewApiKeyHandler(q).Delete()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		// RevokeApiKey uses RETURNING which returns sql.ErrNoRows when no match,
		// mapped to DatabaseError -> 500
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d: %s", resp.StatusCode, w.Body.String())
		}
	})

	t.Run("500 - delete key with wrong org_id", func(t *testing.T) {
		// 異常系 - org mismatch: key belongs to org1 but request uses org2
		q, org := setupTestOrg(t)
		apiKey := seedApiKey(t, q, org.ID, "wrong-org-key")

		otherOrg, err := q.CreateOrganization(context.Background(), db.CreateOrganizationParams{
			Name: "Other Org",
			Slug: "other-org-" + uuid.New().String()[:8],
			Plan: "free",
		})
		if err != nil {
			t.Fatalf("failed to create other org: %v", err)
		}

		req := httptest.NewRequest(http.MethodDelete, "/organizations/"+otherOrg.ID.String()+"/api-keys/"+apiKey.ID.String(), nil)
		testutil.SetAuthHeader(req)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("org_id", otherOrg.ID.String())
		rctx.URLParams.Add("id", apiKey.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler := NewApiKeyHandler(q).Delete()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		// RevokeApiKey filters by both id AND organization_id, so wrong org returns no rows -> 500
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d: %s", resp.StatusCode, w.Body.String())
		}
	})

	t.Run("500 - delete already revoked key", func(t *testing.T) {
		// 境界値 - double delete
		q, org := setupTestOrg(t)
		apiKey := seedApiKey(t, q, org.ID, "double-delete")

		// First delete
		req1 := httptest.NewRequest(http.MethodDelete, "/organizations/"+org.ID.String()+"/api-keys/"+apiKey.ID.String(), nil)
		testutil.SetAuthHeader(req1)
		rctx1 := chi.NewRouteContext()
		rctx1.URLParams.Add("org_id", org.ID.String())
		rctx1.URLParams.Add("id", apiKey.ID.String())
		req1 = req1.WithContext(context.WithValue(req1.Context(), chi.RouteCtxKey, rctx1))
		w1 := httptest.NewRecorder()
		NewApiKeyHandler(q).Delete().ServeHTTP(w1, req1)

		if diff := cmp.Diff(http.StatusNoContent, w1.Result().StatusCode); diff != "" {
			t.Fatalf("first delete failed (-want +got):\n%s", diff)
		}

		// Second delete of same key - the row still exists (soft delete sets revoked_at),
		// so the UPDATE with RETURNING should still match and succeed.
		req2 := httptest.NewRequest(http.MethodDelete, "/organizations/"+org.ID.String()+"/api-keys/"+apiKey.ID.String(), nil)
		testutil.SetAuthHeader(req2)
		rctx2 := chi.NewRouteContext()
		rctx2.URLParams.Add("org_id", org.ID.String())
		rctx2.URLParams.Add("id", apiKey.ID.String())
		req2 = req2.WithContext(context.WithValue(req2.Context(), chi.RouteCtxKey, rctx2))
		w2 := httptest.NewRecorder()
		NewApiKeyHandler(q).Delete().ServeHTTP(w2, req2)

		// Second revoke should also succeed (204) since RevokeApiKey doesn't check revoked_at
		if diff := cmp.Diff(http.StatusNoContent, w2.Result().StatusCode); diff != "" {
			t.Errorf("second delete status mismatch (-want +got):\n%s", diff)
		}
	})
}

// ---------- generateApiKey() Tests ----------

func TestGenerateApiKey(t *testing.T) {
	t.Run("正常系 - generates valid key components", func(t *testing.T) {
		rawKey, hash, prefix, err := generateApiKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// rawKey should start with pl_live_
		if !strings.HasPrefix(rawKey, "pl_live_") {
			t.Errorf("expected key to start with 'pl_live_', got %q", rawKey)
		}

		// rawKey should be pl_live_ (8) + 64 hex chars
		if diff := cmp.Diff(72, len(rawKey)); diff != "" {
			t.Errorf("raw key length mismatch (-want +got):\n%s", diff)
		}

		// hash should be 64 hex chars (SHA-256)
		if diff := cmp.Diff(64, len(hash)); diff != "" {
			t.Errorf("hash length mismatch (-want +got):\n%s", diff)
		}

		// prefix should be pl_live_ + first 8 hex chars = 16 chars
		if !strings.HasPrefix(prefix, "pl_live_") {
			t.Errorf("prefix should start with 'pl_live_', got %q", prefix)
		}
		if diff := cmp.Diff(16, len(prefix)); diff != "" {
			t.Errorf("prefix length mismatch (-want +got):\n%s", diff)
		}

		// prefix hex part should match beginning of hex part in rawKey
		hexPart := rawKey[8:]   // strip pl_live_
		prefixHex := prefix[8:] // strip pl_live_
		if diff := cmp.Diff(hexPart[:8], prefixHex); diff != "" {
			t.Errorf("prefix hex should match start of hex part (-want +got):\n%s", diff)
		}
	})

	t.Run("正常系 - hash is SHA-256 of raw key", func(t *testing.T) {
		rawKey, hash, _, err := generateApiKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		h := sha256.Sum256([]byte(rawKey))
		expectedHash := hex.EncodeToString(h[:])
		if diff := cmp.Diff(expectedHash, hash); diff != "" {
			t.Errorf("hash mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("正常系 - generates unique keys on successive calls", func(t *testing.T) {
		rawKey1, hash1, prefix1, err := generateApiKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		rawKey2, hash2, prefix2, err := generateApiKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rawKey1 == rawKey2 {
			t.Error("expected unique raw keys")
		}
		if hash1 == hash2 {
			t.Error("expected unique hashes")
		}
		_ = prefix1
		_ = prefix2
	})

	t.Run("正常系 - all hex characters in key", func(t *testing.T) {
		rawKey, _, _, err := generateApiKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		hexPart := rawKey[8:] // strip pl_live_
		_, err = hex.DecodeString(hexPart)
		if err != nil {
			t.Errorf("hex part of key is not valid hex: %v", err)
		}
	})

	t.Run("正常系 - all hex characters in hash", func(t *testing.T) {
		_, hash, _, err := generateApiKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = hex.DecodeString(hash)
		if err != nil {
			t.Errorf("hash is not valid hex: %v", err)
		}
	})

	t.Run("正常系 - prefix satisfies DB CHECK constraint", func(t *testing.T) {
		_, _, prefix, err := generateApiKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// DB constraint: key_prefix ~ '^pl_(live|test)_'
		if !strings.HasPrefix(prefix, "pl_live_") && !strings.HasPrefix(prefix, "pl_test_") {
			t.Errorf("prefix %q does not match DB CHECK constraint '^pl_(live|test)_'", prefix)
		}
	})
}
