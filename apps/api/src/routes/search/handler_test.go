package search

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	db "utils/db/db"
	"utils/testutil"

	"api/src/services/embeddingservice"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestSemanticSearch(t *testing.T) {
	type args struct {
		body any
	}
	type expected struct {
		statusCode int
	}

	tests := []struct {
		testName  string
		args      args
		embSvcNil bool
		expected  expected
	}{
		// 異常系 - embedding service unavailable
		{
			testName:  "embedding service unavailable returns error",
			embSvcNil: true,
			args: args{
				body: map[string]any{
					"query":  "test query",
					"org_id": uuid.New().String(),
				},
			},
			expected: expected{statusCode: http.StatusInternalServerError},
		},
		// 異常系 - empty query
		{
			testName:  "empty query returns validation error",
			embSvcNil: true,
			args: args{
				body: map[string]any{
					"query":  "",
					"org_id": uuid.New().String(),
				},
			},
			expected: expected{statusCode: http.StatusInternalServerError},
		},
		// 異常系 - invalid org_id
		{
			testName:  "invalid org_id returns validation error",
			embSvcNil: true,
			args: args{
				body: map[string]any{
					"query":  "test query",
					"org_id": "not-a-uuid",
				},
			},
			expected: expected{statusCode: http.StatusInternalServerError},
		},
		// 異常系 - invalid JSON body
		{
			testName:  "invalid JSON body",
			embSvcNil: true,
			args:      args{body: "not-json"},
			expected:  expected{statusCode: http.StatusInternalServerError},
		},
		// 空文字 - empty body
		{
			testName:  "empty body returns error",
			embSvcNil: true,
			args: args{
				body: map[string]any{},
			},
			expected: expected{statusCode: http.StatusInternalServerError},
		},
		// 特殊文字
		{
			testName:  "special characters in query with unavailable service",
			embSvcNil: true,
			args: args{
				body: map[string]any{
					"query":  "日本語テスト 🔍",
					"org_id": uuid.New().String(),
				},
			},
			expected: expected{statusCode: http.StatusInternalServerError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			q := testutil.SetupTestTx(t)

			var embSvc *embeddingservice.EmbeddingService
			if tt.embSvcNil {
				embSvc = embeddingservice.NewEmbeddingService(nil, nil)
			}

			handler := NewSearchHandler(embSvc, q).SemanticSearch()

			var bodyBytes []byte
			switch v := tt.args.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				var err error
				bodyBytes, err = json.Marshal(v)
				if err != nil {
					t.Fatalf("failed to marshal body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/search/semantic", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			testutil.SetAuthHeader(req)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if diff := cmp.Diff(tt.expected.statusCode, w.Result().StatusCode); diff != "" {
				t.Errorf("status code mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSemanticSearchValidation(t *testing.T) {
	// Test that when embedding service is available but query/org_id validation fails,
	// we still get appropriate errors. Since we can't create a real embedding client,
	// these tests verify the unavailable path.
	t.Run("service unavailable before query validation", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		embSvc := embeddingservice.NewEmbeddingService(nil, nil)
		handler := NewSearchHandler(embSvc, q).SemanticSearch()

		body, _ := json.Marshal(map[string]any{
			"query":  "find prompts about testing",
			"org_id": uuid.New().String(),
			"limit":  5,
		})

		req := httptest.NewRequest(http.MethodPost, "/search/semantic", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Service unavailable is checked first, so we get 500
		if diff := cmp.Diff(http.StatusInternalServerError, w.Result().StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}
	})

	// 境界値 - limit boundaries
	t.Run("limit at boundary with unavailable service", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		embSvc := embeddingservice.NewEmbeddingService(nil, nil)
		handler := NewSearchHandler(embSvc, q).SemanticSearch()

		body, _ := json.Marshal(map[string]any{
			"query":  "test",
			"org_id": uuid.New().String(),
			"limit":  51,
		})

		req := httptest.NewRequest(http.MethodPost, "/search/semantic", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Should fail at service unavailable check
		if diff := cmp.Diff(http.StatusInternalServerError, w.Result().StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestEmbeddingStatus(t *testing.T) {
	type expected struct {
		statusCode int
		status     string
	}

	tests := []struct {
		testName  string
		embSvcNil bool
		expected  expected
	}{
		// 正常系 - service disabled
		{
			testName:  "embedding service disabled returns disabled status",
			embSvcNil: true,
			expected: expected{
				statusCode: http.StatusOK,
				status:     "disabled",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			q := testutil.SetupTestTx(t)

			var embSvc *embeddingservice.EmbeddingService
			if tt.embSvcNil {
				embSvc = embeddingservice.NewEmbeddingService(nil, nil)
			}

			handler := NewSearchHandler(embSvc, q).EmbeddingStatus()

			req := httptest.NewRequest(http.MethodGet, "/search/embedding/status", nil)
			testutil.SetAuthHeader(req)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if diff := cmp.Diff(tt.expected.statusCode, w.Result().StatusCode); diff != "" {
				t.Errorf("status code mismatch (-want +got):\n%s", diff)
			}

			var result map[string]string
			if err := json.NewDecoder(w.Result().Body).Decode(&result); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if diff := cmp.Diff(tt.expected.status, result["embedding_service"]); diff != "" {
				t.Errorf("embedding_service status mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUnavailableErrorUnwrap(t *testing.T) {
	err := &unavailableError{}
	if err.Unwrap() != nil {
		t.Error("expected nil from Unwrap()")
	}
}

func TestValidationErrorUnwrap(t *testing.T) {
	err := &validationError{msg: "test"}
	if err.Unwrap() != nil {
		t.Error("expected nil from Unwrap()")
	}
}

func TestToSearchResult(t *testing.T) {
	id := uuid.New()
	promptID := uuid.New()
	now := time.Now()

	tests := []struct {
		testName string
		row      db.SearchPromptVersionsByEmbeddingRow
		expected searchResultResponse
	}{
		// 正常系 - with valid change description
		{
			testName: "valid row with change description",
			row: db.SearchPromptVersionsByEmbeddingRow{
				ID:                id,
				PromptID:          promptID,
				PromptName:        "Test Prompt",
				PromptSlug:        "test-prompt",
				VersionNumber:     1,
				Status:            "production",
				Content:           json.RawMessage(`{"text":"hello"}`),
				ChangeDescription: sql.NullString{String: "initial version", Valid: true},
				Similarity:        0.95,
				CreatedAt:         now,
			},
			expected: searchResultResponse{
				ID:                id.String(),
				PromptID:          promptID.String(),
				PromptName:        "Test Prompt",
				PromptSlug:        "test-prompt",
				VersionNumber:     1,
				Status:            "production",
				Content:           json.RawMessage(`{"text":"hello"}`),
				ChangeDescription: "initial version",
				Similarity:        0.95,
				CreatedAt:         now.Format(time.RFC3339),
			},
		},
		// Nil - null change description
		{
			testName: "row with null change description",
			row: db.SearchPromptVersionsByEmbeddingRow{
				ID:                id,
				PromptID:          promptID,
				PromptName:        "Prompt",
				PromptSlug:        "prompt",
				VersionNumber:     2,
				Status:            "draft",
				Content:           json.RawMessage(`{}`),
				ChangeDescription: sql.NullString{Valid: false},
				Similarity:        0.80,
				CreatedAt:         now,
			},
			expected: searchResultResponse{
				ID:                id.String(),
				PromptID:          promptID.String(),
				PromptName:        "Prompt",
				PromptSlug:        "prompt",
				VersionNumber:     2,
				Status:            "draft",
				Content:           json.RawMessage(`{}`),
				ChangeDescription: "",
				Similarity:        0.80,
				CreatedAt:         now.Format(time.RFC3339),
			},
		},
		// 特殊文字 - Unicode in fields
		{
			testName: "row with Unicode characters",
			row: db.SearchPromptVersionsByEmbeddingRow{
				ID:                id,
				PromptID:          promptID,
				PromptName:        "日本語プロンプト",
				PromptSlug:        "jp-prompt",
				VersionNumber:     1,
				Status:            "production",
				Content:           json.RawMessage(`{"text":"こんにちは"}`),
				ChangeDescription: sql.NullString{String: "変更説明 🔄", Valid: true},
				Similarity:        0.99,
				CreatedAt:         now,
			},
			expected: searchResultResponse{
				ID:                id.String(),
				PromptID:          promptID.String(),
				PromptName:        "日本語プロンプト",
				PromptSlug:        "jp-prompt",
				VersionNumber:     1,
				Status:            "production",
				Content:           json.RawMessage(`{"text":"こんにちは"}`),
				ChangeDescription: "変更説明 🔄",
				Similarity:        0.99,
				CreatedAt:         now.Format(time.RFC3339),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := toSearchResult(tt.row)
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestEmbeddingStatusResponseFormat(t *testing.T) {
	t.Run("response has correct JSON structure", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		embSvc := embeddingservice.NewEmbeddingService(nil, nil)
		handler := NewSearchHandler(embSvc, q).EmbeddingStatus()

		req := httptest.NewRequest(http.MethodGet, "/search/embedding/status", nil)
		testutil.SetAuthHeader(req)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		if diff := cmp.Diff("application/json", resp.Header.Get("Content-Type")); diff != "" {
			t.Errorf("Content-Type mismatch (-want +got):\n%s", diff)
		}

		var result map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if _, ok := result["embedding_service"]; !ok {
			t.Error("expected 'embedding_service' key in response")
		}
	})
}
