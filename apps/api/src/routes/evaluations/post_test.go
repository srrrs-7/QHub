package evaluations

import (
	"api/src/infra/rds/executionlog_repository"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"utils/db/db"
	"utils/testutil"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

// seedExecutionLog creates the full FK chain: org → project → prompt → prompt_version → execution_log
// and returns the execution log ID.
func seedExecutionLog(t *testing.T, q db.Querier) string {
	t.Helper()
	ctx := context.Background()

	org, err := q.CreateOrganization(ctx, db.CreateOrganizationParams{
		Name: "Test Org",
		Slug: "test-org-" + uuid.New().String()[:8],
		Plan: "free",
	})
	if err != nil {
		t.Fatalf("failed to create org: %v", err)
	}

	proj, err := q.CreateProject(ctx, db.CreateProjectParams{
		OrganizationID: org.ID,
		Name:           "Test Project",
		Slug:           "test-proj-" + uuid.New().String()[:8],
	})
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	prompt, err := q.CreatePrompt(ctx, db.CreatePromptParams{
		ProjectID:  proj.ID,
		Name:       "Test Prompt",
		Slug:       "test-prompt-" + uuid.New().String()[:8],
		PromptType: "system",
	})
	if err != nil {
		t.Fatalf("failed to create prompt: %v", err)
	}

	user, err := q.CreateUser(ctx, db.CreateUserParams{
		Email: "test-" + uuid.New().String()[:8] + "@example.com",
		Name:  "Test User",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	_, err = q.CreatePromptVersion(ctx, db.CreatePromptVersionParams{
		PromptID:      prompt.ID,
		VersionNumber: 1,
		Status:        "draft",
		Content:       json.RawMessage(`{"text":"hello"}`),
		AuthorID:      user.ID,
	})
	if err != nil {
		t.Fatalf("failed to create prompt version: %v", err)
	}

	log, err := q.CreateExecutionLog(ctx, db.CreateExecutionLogParams{
		OrganizationID: org.ID,
		PromptID:       prompt.ID,
		VersionNumber:  1,
		RequestBody:    json.RawMessage(`{"prompt":"hello"}`),
		ResponseBody:   pqtype.NullRawMessage{RawMessage: json.RawMessage(`{"text":"world"}`), Valid: true},
		Model:          "claude-sonnet-4-20250514",
		Provider:       "anthropic",
		InputTokens:    10,
		OutputTokens:   20,
		TotalTokens:    30,
		LatencyMs:      150,
		EstimatedCost:  "0.001",
		Status:         "success",
		Environment:    "development",
		ExecutedAt:     time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create execution log: %v", err)
	}

	return log.ID.String()
}

// seedEvaluation creates an evaluation for the given execution log.
func seedEvaluation(t *testing.T, q db.Querier, logID string) string {
	t.Helper()
	ctx := context.Background()

	overall := sql.NullString{String: "4.50", Valid: true}
	accuracy := sql.NullString{String: "3.75", Valid: true}

	eval, err := q.CreateEvaluation(ctx, db.CreateEvaluationParams{
		ExecutionLogID: uuid.MustParse(logID),
		OverallScore:   overall,
		AccuracyScore:  accuracy,
		EvaluatorType:  "human",
		EvaluatorID:    sql.NullString{String: "user-123", Valid: true},
		Feedback:       sql.NullString{String: "Good response", Valid: true},
	})
	if err != nil {
		t.Fatalf("failed to create evaluation: %v", err)
	}

	return eval.ID.String()
}

func TestPostHandler(t *testing.T) {
	t.Run("201 Created", func(t *testing.T) {
		type expected struct {
			statusCode    int
			evaluatorType string
			hasFeedback   bool
			hasScores     bool
		}

		tests := []struct {
			testName string
			reqBody  map[string]any
			expected expected
		}{
			{
				testName: "正常系: create evaluation with all scores",
				reqBody: map[string]any{
					"overall_score":   "4.50",
					"accuracy_score":  "3.75",
					"relevance_score": "4.00",
					"fluency_score":   "3.50",
					"safety_score":    "4.80",
					"feedback":        "Great response overall",
					"evaluator_type":  "human",
					"evaluator_id":    "reviewer-001",
					"metadata":        map[string]string{"source": "manual"},
				},
				expected: expected{
					statusCode:    http.StatusCreated,
					evaluatorType: "human",
					hasFeedback:   true,
					hasScores:     true,
				},
			},
			{
				testName: "正常系: create evaluation with auto evaluator",
				reqBody: map[string]any{
					"overall_score":  "3.50",
					"evaluator_type": "auto",
					"evaluator_id":   "model-eval-v1",
				},
				expected: expected{
					statusCode:    http.StatusCreated,
					evaluatorType: "auto",
					hasFeedback:   false,
					hasScores:     true,
				},
			},
			{
				testName: "正常系: create evaluation with no scores (feedback only)",
				reqBody: map[string]any{
					"feedback":       "Needs improvement",
					"evaluator_type": "human",
				},
				expected: expected{
					statusCode:    http.StatusCreated,
					evaluatorType: "human",
					hasFeedback:   true,
					hasScores:     false,
				},
			},
			{
				testName: "正常系: create evaluation with minimal fields",
				reqBody: map[string]any{
					"evaluator_type": "human",
				},
				expected: expected{
					statusCode:    http.StatusCreated,
					evaluatorType: "human",
					hasFeedback:   false,
					hasScores:     false,
				},
			},
			{
				testName: "特殊文字: unicode in feedback",
				reqBody: map[string]any{
					"overall_score":  "4.00",
					"feedback":       "素晴らしい結果です 🎉",
					"evaluator_type": "human",
					"evaluator_id":   "レビュアー太郎",
				},
				expected: expected{
					statusCode:    http.StatusCreated,
					evaluatorType: "human",
					hasFeedback:   true,
					hasScores:     true,
				},
			},
			{
				testName: "特殊文字: emoji in evaluator_id",
				reqBody: map[string]any{
					"evaluator_type": "auto",
					"evaluator_id":   "bot-🤖-v2",
				},
				expected: expected{
					statusCode:    http.StatusCreated,
					evaluatorType: "auto",
					hasFeedback:   false,
					hasScores:     false,
				},
			},
			{
				testName: "境界値: score at 0",
				reqBody: map[string]any{
					"overall_score":  "0.00",
					"evaluator_type": "human",
				},
				expected: expected{
					statusCode:    http.StatusCreated,
					evaluatorType: "human",
					hasFeedback:   false,
					hasScores:     true,
				},
			},
			{
				testName: "境界値: score at max (5.00)",
				reqBody: map[string]any{
					"overall_score":  "5.00",
					"evaluator_type": "auto",
				},
				expected: expected{
					statusCode:    http.StatusCreated,
					evaluatorType: "auto",
					hasFeedback:   false,
					hasScores:     true,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)
				logID := seedExecutionLog(t, q)

				// Inject the execution_log_id
				tt.reqBody["execution_log_id"] = logID

				jsonBody, err := json.Marshal(tt.reqBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}

				req := httptest.NewRequest(http.MethodPost, "/evaluations", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()

				evalRepo := executionlog_repository.NewEvaluationRepository(q)
				handler := NewEvaluationHandler(evalRepo).Post()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var result map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if _, ok := result["id"]; !ok {
					t.Error("response should contain 'id' field")
				}

				if diff := cmp.Diff(logID, result["execution_log_id"]); diff != "" {
					t.Errorf("execution_log_id mismatch (-want +got):\n%s", diff)
				}

				if diff := cmp.Diff(tt.expected.evaluatorType, result["evaluator_type"]); diff != "" {
					t.Errorf("evaluator_type mismatch (-want +got):\n%s", diff)
				}

				if _, ok := result["created_at"]; !ok {
					t.Error("response should contain 'created_at' field")
				}
			})
		}
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		tests := []struct {
			testName       string
			reqBody        string
			expectedStatus int
		}{
			{
				testName:       "異常系: invalid JSON body",
				reqBody:        `{invalid json`,
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName:       "異常系: empty body",
				reqBody:        ``,
				expectedStatus: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				req := httptest.NewRequest(http.MethodPost, "/evaluations", strings.NewReader(tt.reqBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()

				evalRepo := executionlog_repository.NewEvaluationRepository(q)
				handler := NewEvaluationHandler(evalRepo).Post()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expectedStatus, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("400 Validation Error", func(t *testing.T) {
		tests := []struct {
			testName       string
			reqBody        map[string]any
			expectedStatus int
		}{
			{
				testName: "異常系: missing execution_log_id",
				reqBody: map[string]any{
					"evaluator_type": "human",
				},
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName: "異常系: invalid execution_log_id (not UUID)",
				reqBody: map[string]any{
					"execution_log_id": "not-a-uuid",
					"evaluator_type":   "human",
				},
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName: "異常系: missing evaluator_type",
				reqBody: map[string]any{
					"execution_log_id": uuid.New().String(),
				},
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName: "異常系: invalid evaluator_type value",
				reqBody: map[string]any{
					"execution_log_id": uuid.New().String(),
					"evaluator_type":   "unknown",
				},
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName: "空文字: empty evaluator_type",
				reqBody: map[string]any{
					"execution_log_id": uuid.New().String(),
					"evaluator_type":   "",
				},
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName: "境界値: feedback exceeds max length (2000)",
				reqBody: map[string]any{
					"execution_log_id": uuid.New().String(),
					"evaluator_type":   "human",
					"feedback":         strings.Repeat("a", 2001),
				},
				expectedStatus: http.StatusBadRequest,
			},
			{
				testName: "境界値: evaluator_id exceeds max length (200)",
				reqBody: map[string]any{
					"execution_log_id": uuid.New().String(),
					"evaluator_type":   "human",
					"evaluator_id":     strings.Repeat("x", 201),
				},
				expectedStatus: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				jsonBody, err := json.Marshal(tt.reqBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}

				req := httptest.NewRequest(http.MethodPost, "/evaluations", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				testutil.SetAuthHeader(req)

				w := httptest.NewRecorder()

				evalRepo := executionlog_repository.NewEvaluationRepository(q)
				handler := NewEvaluationHandler(evalRepo).Post()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expectedStatus, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("境界値: feedback at max length (2000)", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		logID := seedExecutionLog(t, q)

		reqBody := map[string]any{
			"execution_log_id": logID,
			"evaluator_type":   "human",
			"feedback":         strings.Repeat("b", 2000),
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/evaluations", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		w := httptest.NewRecorder()

		evalRepo := executionlog_repository.NewEvaluationRepository(q)
		handler := NewEvaluationHandler(evalRepo).Post()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if diff := cmp.Diff(http.StatusCreated, resp.StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("境界値: evaluator_id at max length (200)", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		logID := seedExecutionLog(t, q)

		reqBody := map[string]any{
			"execution_log_id": logID,
			"evaluator_type":   "human",
			"evaluator_id":     strings.Repeat("z", 200),
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/evaluations", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		w := httptest.NewRecorder()

		evalRepo := executionlog_repository.NewEvaluationRepository(q)
		handler := NewEvaluationHandler(evalRepo).Post()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if diff := cmp.Diff(http.StatusCreated, resp.StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("Null/Nil: null scores in request", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		logID := seedExecutionLog(t, q)

		reqBody := map[string]any{
			"execution_log_id": logID,
			"evaluator_type":   "human",
			"overall_score":    nil,
			"accuracy_score":   nil,
			"relevance_score":  nil,
			"fluency_score":    nil,
			"safety_score":     nil,
			"metadata":         nil,
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/evaluations", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		w := httptest.NewRecorder()

		evalRepo := executionlog_repository.NewEvaluationRepository(q)
		handler := NewEvaluationHandler(evalRepo).Post()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if diff := cmp.Diff(http.StatusCreated, resp.StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}

		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		// Null scores should remain null in the response
		if result["overall_score"] != nil {
			t.Errorf("expected overall_score to be nil, got %v", result["overall_score"])
		}
		if result["accuracy_score"] != nil {
			t.Errorf("expected accuracy_score to be nil, got %v", result["accuracy_score"])
		}
	})

	t.Run("異常系: non-existent execution_log_id (FK violation)", func(t *testing.T) {
		q := testutil.SetupTestTx(t)

		reqBody := map[string]any{
			"execution_log_id": uuid.New().String(),
			"evaluator_type":   "human",
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/evaluations", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		w := httptest.NewRecorder()

		evalRepo := executionlog_repository.NewEvaluationRepository(q)
		handler := NewEvaluationHandler(evalRepo).Post()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		// FK violation results in a 500 or appropriate error
		if resp.StatusCode == http.StatusCreated {
			t.Error("expected error for non-existent execution_log_id, but got 201")
		}
	})

	t.Run("空文字: empty feedback and evaluator_id", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		logID := seedExecutionLog(t, q)

		reqBody := map[string]any{
			"execution_log_id": logID,
			"evaluator_type":   "human",
			"feedback":         "",
			"evaluator_id":     "",
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/evaluations", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		w := httptest.NewRecorder()

		evalRepo := executionlog_repository.NewEvaluationRepository(q)
		handler := NewEvaluationHandler(evalRepo).Post()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if diff := cmp.Diff(http.StatusCreated, resp.StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("特殊文字: SQL injection in feedback", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		logID := seedExecutionLog(t, q)

		reqBody := map[string]any{
			"execution_log_id": logID,
			"evaluator_type":   "human",
			"feedback":         "'; DROP TABLE evaluations; --",
			"evaluator_id":     "\" OR 1=1; --",
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/evaluations", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		testutil.SetAuthHeader(req)

		w := httptest.NewRecorder()

		evalRepo := executionlog_repository.NewEvaluationRepository(q)
		handler := NewEvaluationHandler(evalRepo).Post()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		// Should succeed since parameterized queries handle SQL injection safely
		if diff := cmp.Diff(http.StatusCreated, resp.StatusCode); diff != "" {
			t.Errorf("status code mismatch (-want +got):\n%s", diff)
		}
	})
}

