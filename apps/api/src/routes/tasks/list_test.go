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

	"github.com/google/go-cmp/cmp"
)

func TestListHandler(t *testing.T) {

	t.Run("200 OK", func(t *testing.T) {
		type expected struct {
			statusCode int
			taskCount  int
		}

		tests := []struct {
			name     string
			setup    func(t *testing.T, q db.Querier)
			expected expected
		}{
			{
				name: "list empty tasks",
				setup: func(t *testing.T, q db.Querier) {
					// No setup needed - empty database
				},
				expected: expected{
					statusCode: http.StatusOK,
					taskCount:  0,
				},
			},
			{
				name: "list multiple tasks",
				setup: func(t *testing.T, q db.Querier) {
					titles := []string{"Task 1", "Task 2", "Task 3"}
					for _, title := range titles {
						_, err := q.CreateTask(context.Background(), db.CreateTaskParams{
							Title:    title,
							Status:   "pending",
							Priority: "medium",
						})
						if err != nil {
							t.Fatalf("failed to create test task: %v", err)
						}
					}
				},
				expected: expected{
					statusCode: http.StatusOK,
					taskCount:  3,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				q := testutil.SetupTestTx(t)
				tt.setup(t, q)

				req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
				testutil.SetAuthHeader(req)
				w := httptest.NewRecorder()

				repo := task_repository.NewTaskRepository(q)
				handler := NewTaskHandler(repo).List()
				handler.ServeHTTP(w, req)

				resp := w.Result()
				if diff := cmp.Diff(tt.expected.statusCode, resp.StatusCode); diff != "" {
					t.Errorf("status code mismatch (-want +got):\n%s", diff)
				}

				var result struct {
					Tasks []map[string]any `json:"tasks"`
				}
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if diff := cmp.Diff(tt.expected.taskCount, len(result.Tasks)); diff != "" {
					t.Errorf("task count mismatch (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("data integrity", func(t *testing.T) {
		q := testutil.SetupTestTx(t)

		// Create a task
		task, err := q.CreateTask(context.Background(), db.CreateTaskParams{
			Title:    "Data Integrity Test",
			Status:   "pending",
			Priority: "medium",
		})
		if err != nil {
			t.Fatalf("failed to create test task: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
		testutil.SetAuthHeader(req)
		w := httptest.NewRecorder()

		repo := task_repository.NewTaskRepository(q)
		handler := NewTaskHandler(repo).List()
		handler.ServeHTTP(w, req)

		var result struct {
			Tasks []map[string]any `json:"tasks"`
		}
		if err := json.NewDecoder(w.Result().Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(result.Tasks) != 1 {
			t.Fatalf("expected 1 task, got %d", len(result.Tasks))
		}

		if diff := cmp.Diff(task.ID.String(), result.Tasks[0]["id"]); diff != "" {
			t.Errorf("id mismatch (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff("Data Integrity Test", result.Tasks[0]["title"]); diff != "" {
			t.Errorf("title mismatch (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff("pending", result.Tasks[0]["status"]); diff != "" {
			t.Errorf("status mismatch (-want +got):\n%s", diff)
		}
	})
}
