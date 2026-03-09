package tasks

import (
	"api/src/infra/rds/task_repository"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"utils/db/db"
	"utils/testutil"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
)

func TestGetHandler(t *testing.T) {

	t.Run("200 OK", func(t *testing.T) {
		type expected struct {
			statusCode int
			title      string
			status     string
		}

		tests := []struct {
			name     string
			setup    func(t *testing.T, q db.Querier) string
			expected expected
		}{
			{
				name: "get existing task",
				setup: func(t *testing.T, q db.Querier) string {
					task, err := q.CreateTask(context.Background(), db.CreateTaskParams{
						Title:    "Integration Test Task",
						Status:   "pending",
						Priority: "medium",
					})
					if err != nil {
						t.Fatalf("failed to create test task: %v", err)
					}
					return task.ID.String()
				},
				expected: expected{
					statusCode: http.StatusOK,
					title:      "Integration Test Task",
					status:     "pending",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				q := testutil.SetupTestTx(t)
				taskID := tt.setup(t, q)

				req := httptest.NewRequest(http.MethodGet, "/tasks/"+taskID, nil)

				// Set URL params for chi router
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", taskID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				repo := task_repository.NewTaskRepository(q)
				handler := NewTaskHandler(repo).Get()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var result map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(taskID, result["id"]); diff != "" {
					t.Errorf("id mismatch (-want +got):\n%s", diff)
				}

				if diff := cmp.Diff(tt.expected.title, result["title"]); diff != "" {
					t.Errorf("title mismatch (-want +got):\n%s", diff)
				}

				if diff := cmp.Diff(tt.expected.status, result["status"]); diff != "" {
					t.Errorf("status mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("400 Bad Request - invalid UUID", func(t *testing.T) {
		tests := []struct {
			name           string
			taskID         string
			expectedStatus int
		}{
			{
				name:           "malformed uuid",
				taskID:         "not-a-uuid",
				expectedStatus: http.StatusBadRequest,
			},
			{
				name:           "empty uuid",
				taskID:         "",
				expectedStatus: http.StatusBadRequest,
			},
			{
				name:           "special characters",
				taskID:         "<script>alert('xss')</script>",
				expectedStatus: http.StatusBadRequest,
			},
			{
				name:           "Japanese characters",
				taskID:         "タスク識別子",
				expectedStatus: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				req := httptest.NewRequest(http.MethodGet, "/tasks/"+tt.taskID, nil)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", tt.taskID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				repo := task_repository.NewTaskRepository(q)
				handler := NewTaskHandler(repo).Get()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expectedStatus, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("404 Not Found", func(t *testing.T) {
		tests := []struct {
			name           string
			taskID         string
			expectedStatus int
		}{
			{
				name:           "non-existent task",
				taskID:         "00000000-0000-0000-0000-000000000000",
				expectedStatus: http.StatusNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				q := testutil.SetupTestTx(t)

				req := httptest.NewRequest(http.MethodGet, "/tasks/"+tt.taskID, nil)

				// Set URL params for chi router
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", tt.taskID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				repo := task_repository.NewTaskRepository(q)
				handler := NewTaskHandler(repo).Get()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expectedStatus, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})
}
