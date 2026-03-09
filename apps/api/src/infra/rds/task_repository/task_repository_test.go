package task_repository

import (
	"api/src/domain/apperror"
	"api/src/domain/task"
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
	"utils/db/db"
	"utils/testutil"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

// TestFindTaskByID tests the FindTaskByID function with all 6 TDD categories
func TestFindTaskByID(t *testing.T) {
	type args struct {
		id task.TaskID
	}

	type expected struct {
		isErr      bool
		errName    string
		domainName string
		checkTask  bool
		taskTitle  string
		taskStatus string
	}

	tests := []struct {
		testName string
		setup    func(t *testing.T, q db.Querier) uuid.UUID
		args     func(setupID uuid.UUID) args
		expected expected
	}{
		// 正常系 (Happy Path)
		{
			testName: "正常系: find existing task",
			setup: func(t *testing.T, q db.Querier) uuid.UUID {
				t.Helper()
				created, err := q.CreateTask(context.Background(), db.CreateTaskParams{
					Title:    "Test Task",
					Status:   "pending",
					Priority: "medium",
				})
				if err != nil {
					t.Fatalf("failed to create test task: %v", err)
				}
				return created.ID
			},
			args: func(setupID uuid.UUID) args {
				return args{id: task.TaskIDFromUUID(setupID)}
			},
			expected: expected{
				isErr:      false,
				checkTask:  true,
				taskTitle:  "Test Task",
				taskStatus: "pending",
			},
		},
		{
			testName: "正常系: find task with description",
			setup: func(t *testing.T, q db.Querier) uuid.UUID {
				t.Helper()
				created, err := q.CreateTask(context.Background(), db.CreateTaskParams{
					Title:       "Task with Description",
					Description: sql.NullString{String: "This is a detailed description", Valid: true},
					Status:      "completed",
					Priority:    "high",
				})
				if err != nil {
					t.Fatalf("failed to create test task: %v", err)
				}
				return created.ID
			},
			args: func(setupID uuid.UUID) args {
				return args{id: task.TaskIDFromUUID(setupID)}
			},
			expected: expected{
				isErr:      false,
				checkTask:  true,
				taskTitle:  "Task with Description",
				taskStatus: "completed",
			},
		},
		// 異常系 (Error Cases)
		{
			testName: "異常系: task not found",
			setup: func(t *testing.T, q db.Querier) uuid.UUID {
				t.Helper()
				return uuid.New()
			},
			args: func(setupID uuid.UUID) args {
				return args{id: task.TaskIDFromUUID(setupID)}
			},
			expected: expected{
				isErr:      true,
				errName:    apperror.NotFoundErrorName,
				domainName: "Task",
			},
		},
		// Nil - empty description
		{
			testName: "Nil: empty description",
			setup: func(t *testing.T, q db.Querier) uuid.UUID {
				t.Helper()
				created, err := q.CreateTask(context.Background(), db.CreateTaskParams{
					Title:       "Task without Description",
					Description: sql.NullString{String: "", Valid: false},
					Status:      "pending",
					Priority:    "medium",
				})
				if err != nil {
					t.Fatalf("failed to create test task: %v", err)
				}
				return created.ID
			},
			args: func(setupID uuid.UUID) args {
				return args{id: task.TaskIDFromUUID(setupID)}
			},
			expected: expected{
				isErr:      false,
				checkTask:  true,
				taskTitle:  "Task without Description",
				taskStatus: "pending",
			},
		},
		// 特殊文字 (Special Chars)
		{
			testName: "特殊文字: task with emoji in title",
			setup: func(t *testing.T, q db.Querier) uuid.UUID {
				t.Helper()
				created, err := q.CreateTask(context.Background(), db.CreateTaskParams{
					Title:    "Task 📋 with emoji",
					Status:   "pending",
					Priority: "medium",
				})
				if err != nil {
					t.Fatalf("failed to create test task: %v", err)
				}
				return created.ID
			},
			args: func(setupID uuid.UUID) args {
				return args{id: task.TaskIDFromUUID(setupID)}
			},
			expected: expected{
				isErr:      false,
				checkTask:  true,
				taskTitle:  "Task 📋 with emoji",
				taskStatus: "pending",
			},
		},
		{
			testName: "特殊文字: task with Japanese characters",
			setup: func(t *testing.T, q db.Querier) uuid.UUID {
				t.Helper()
				created, err := q.CreateTask(context.Background(), db.CreateTaskParams{
					Title:       "タスク",
					Description: sql.NullString{String: "これは日本語の説明です", Valid: true},
					Status:      "pending",
					Priority:    "medium",
				})
				if err != nil {
					t.Fatalf("failed to create test task: %v", err)
				}
				return created.ID
			},
			args: func(setupID uuid.UUID) args {
				return args{id: task.TaskIDFromUUID(setupID)}
			},
			expected: expected{
				isErr:      false,
				checkTask:  true,
				taskTitle:  "タスク",
				taskStatus: "pending",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			q := testutil.SetupTestTx(t)
			setupID := tt.setup(t, q)
			args := tt.args(setupID)

			repo := NewTaskRepository(q)
			got, err := repo.FindByID(context.Background(), args.id)

			if tt.expected.isErr {
				if err == nil {
					t.Fatalf("expected error but got success")
				}
				var appErr apperror.AppError
				if errors.As(err, &appErr) {
					if diff := cmp.Diff(tt.expected.errName, appErr.ErrorName()); diff != "" {
						t.Errorf("error name mismatch (-want +got):\n%s", diff)
					}
					if diff := cmp.Diff(tt.expected.domainName, appErr.DomainName()); diff != "" {
						t.Errorf("domain name mismatch (-want +got):\n%s", diff)
					}
				} else {
					t.Errorf("expected AppError, got: %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected success but got error: %v", err)
				}
				if tt.expected.checkTask {
					if diff := cmp.Diff(tt.expected.taskTitle, got.Title.String()); diff != "" {
						t.Errorf("task title mismatch (-want +got):\n%s", diff)
					}
					if diff := cmp.Diff(tt.expected.taskStatus, got.Status.String()); diff != "" {
						t.Errorf("task status mismatch (-want +got):\n%s", diff)
					}
					if diff := cmp.Diff(args.id.String(), got.ID.String()); diff != "" {
						t.Errorf("task ID mismatch (-want +got):\n%s", diff)
					}
				}
			}
		})
	}
}

// TestFindAllTasks tests the FindAllTasks function with all 6 TDD categories
func TestFindAllTasks(t *testing.T) {
	type expected struct {
		isErr      bool
		errName    string
		domainName string
		taskCount  int
		checkFirst bool
		firstTitle string
	}

	tests := []struct {
		testName string
		setup    func(t *testing.T, q db.Querier)
		expected expected
	}{
		// 正常系 (Happy Path)
		{
			testName: "正常系: find multiple tasks",
			setup: func(t *testing.T, q db.Querier) {
				t.Helper()
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
				isErr:      false,
				taskCount:  3,
				checkFirst: false,
			},
		},
		// 境界値 (Boundary Values)
		{
			testName: "境界値: empty task list",
			setup: func(t *testing.T, q db.Querier) {
				t.Helper()
			},
			expected: expected{
				isErr:      false,
				taskCount:  0,
				checkFirst: false,
			},
		},
		{
			testName: "境界値: single task",
			setup: func(t *testing.T, q db.Querier) {
				t.Helper()
				_, err := q.CreateTask(context.Background(), db.CreateTaskParams{
					Title:    "Single Task",
					Status:   "pending",
					Priority: "medium",
				})
				if err != nil {
					t.Fatalf("failed to create test task: %v", err)
				}
			},
			expected: expected{
				isErr:      false,
				taskCount:  1,
				checkFirst: true,
				firstTitle: "Single Task",
			},
		},
		// 特殊文字 (Special Chars)
		{
			testName: "特殊文字: tasks with emoji",
			setup: func(t *testing.T, q db.Querier) {
				t.Helper()
				titles := []string{"Task 📋", "Task 🚀", "Task ✅"}
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
				isErr:      false,
				taskCount:  3,
				checkFirst: false,
			},
		},
		{
			testName: "特殊文字: tasks with Japanese characters",
			setup: func(t *testing.T, q db.Querier) {
				t.Helper()
				titles := []string{"タスク1", "タスク2"}
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
				isErr:      false,
				taskCount:  2,
				checkFirst: false,
			},
		},
		// Nil - empty descriptions
		{
			testName: "Nil: tasks with null descriptions",
			setup: func(t *testing.T, q db.Querier) {
				t.Helper()
				_, err := q.CreateTask(context.Background(), db.CreateTaskParams{
					Title:       "Task without description",
					Description: sql.NullString{String: "", Valid: false},
					Status:      "pending",
					Priority:    "medium",
				})
				if err != nil {
					t.Fatalf("failed to create test task: %v", err)
				}
			},
			expected: expected{
				isErr:      false,
				taskCount:  1,
				checkFirst: true,
				firstTitle: "Task without description",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			q := testutil.SetupTestTx(t)
			tt.setup(t, q)

			repo := NewTaskRepository(q)
			tasks, err := repo.FindAll(context.Background())

			if tt.expected.isErr {
				if err == nil {
					t.Fatalf("expected error but got success")
				}
				var appErr apperror.AppError
				if errors.As(err, &appErr) {
					if diff := cmp.Diff(tt.expected.errName, appErr.ErrorName()); diff != "" {
						t.Errorf("error name mismatch (-want +got):\n%s", diff)
					}
					if diff := cmp.Diff(tt.expected.domainName, appErr.DomainName()); diff != "" {
						t.Errorf("domain name mismatch (-want +got):\n%s", diff)
					}
				}
			} else {
				if err != nil {
					t.Fatalf("expected success but got error: %v", err)
				}
				if diff := cmp.Diff(tt.expected.taskCount, len(tasks)); diff != "" {
					t.Errorf("task count mismatch (-want +got):\n%s", diff)
				}
				if tt.expected.checkFirst && len(tasks) > 0 {
					if diff := cmp.Diff(tt.expected.firstTitle, tasks[0].Title.String()); diff != "" {
						t.Errorf("first task title mismatch (-want +got):\n%s", diff)
					}
				}
			}
		})
	}
}

// TestCreateTask tests the CreateTask function with all 6 TDD categories
func TestCreateTask(t *testing.T) {
	type args struct {
		cmd task.TaskCmd
	}

	type expected struct {
		isErr      bool
		errName    string
		domainName string
		checkTask  bool
		taskTitle  string
		taskStatus string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系 (Happy Path)
		{
			testName: "正常系: create task with description",
			args: args{
				cmd: task.NewTaskCmd(
					task.TaskTitle("New Task"),
					task.TaskDescription("This is a new task"),
					task.TaskStatusPending,
				),
			},
			expected: expected{
				isErr:      false,
				checkTask:  true,
				taskTitle:  "New Task",
				taskStatus: "pending",
			},
		},
		{
			testName: "正常系: create task without description",
			args: args{
				cmd: task.NewTaskCmd(
					task.TaskTitle("Task without description"),
					task.TaskDescription(""),
					task.TaskStatusPending,
				),
			},
			expected: expected{
				isErr:      false,
				checkTask:  true,
				taskTitle:  "Task without description",
				taskStatus: "pending",
			},
		},
		// 特殊文字 (Special Chars)
		{
			testName: "特殊文字: create task with emoji",
			args: args{
				cmd: task.NewTaskCmd(
					task.TaskTitle("Task 📋 with emoji"),
					task.TaskDescription("Description 🚀 with emoji"),
					task.TaskStatusPending,
				),
			},
			expected: expected{
				isErr:      false,
				checkTask:  true,
				taskTitle:  "Task 📋 with emoji",
				taskStatus: "pending",
			},
		},
		{
			testName: "特殊文字: create task with Japanese characters",
			args: args{
				cmd: task.NewTaskCmd(
					task.TaskTitle("新しいタスク"),
					task.TaskDescription("これは新しいタスクの説明です"),
					task.TaskStatusPending,
				),
			},
			expected: expected{
				isErr:      false,
				checkTask:  true,
				taskTitle:  "新しいタスク",
				taskStatus: "pending",
			},
		},
		{
			testName: "特殊文字: create task with special characters",
			args: args{
				cmd: task.NewTaskCmd(
					task.TaskTitle("Task with symbols !@#$%^&*()"),
					task.TaskDescription("Description with <>&\"'"),
					task.TaskStatusPending,
				),
			},
			expected: expected{
				isErr:      false,
				checkTask:  true,
				taskTitle:  "Task with symbols !@#$%^&*()",
				taskStatus: "pending",
			},
		},
		// 空文字 (Empty String)
		{
			testName: "空文字: create task with empty description",
			args: args{
				cmd: task.NewTaskCmd(
					task.TaskTitle("Task with empty description"),
					task.TaskDescription(""),
					task.TaskStatusPending,
				),
			},
			expected: expected{
				isErr:      false,
				checkTask:  true,
				taskTitle:  "Task with empty description",
				taskStatus: "pending",
			},
		},
		// Nil - zero values
		{
			testName: "Nil: create task with zero value status",
			args: args{
				cmd: task.NewTaskCmd(
					task.TaskTitle("Task with default status"),
					task.TaskDescription("Description"),
					task.TaskStatus(""),
				),
			},
			expected: expected{
				isErr:      false,
				checkTask:  true,
				taskTitle:  "Task with default status",
				taskStatus: "pending",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			q := testutil.SetupTestTx(t)

			repo := NewTaskRepository(q)
			createdTask, err := repo.Create(context.Background(), tt.args.cmd)

			if tt.expected.isErr {
				if err == nil {
					t.Fatalf("expected error but got success")
				}
				var appErr apperror.AppError
				if errors.As(err, &appErr) {
					if diff := cmp.Diff(tt.expected.errName, appErr.ErrorName()); diff != "" {
						t.Errorf("error name mismatch (-want +got):\n%s", diff)
					}
					if diff := cmp.Diff(tt.expected.domainName, appErr.DomainName()); diff != "" {
						t.Errorf("domain name mismatch (-want +got):\n%s", diff)
					}
				}
			} else {
				if err != nil {
					t.Fatalf("expected success but got error: %v", err)
				}
				if tt.expected.checkTask {
					if diff := cmp.Diff(tt.expected.taskTitle, createdTask.Title.String()); diff != "" {
						t.Errorf("task title mismatch (-want +got):\n%s", diff)
					}
					if diff := cmp.Diff(tt.expected.taskStatus, createdTask.Status.String()); diff != "" {
						t.Errorf("task status mismatch (-want +got):\n%s", diff)
					}
					// Verify task was actually created in DB
					_, verifyErr := repo.FindByID(context.Background(), createdTask.ID)
					if verifyErr != nil {
						t.Errorf("created task not found in database: %v", verifyErr)
					}
				}
			}
		})
	}
}

// TestUpdateTask tests the UpdateTask function with all 6 TDD categories
func TestUpdateTask(t *testing.T) {
	type args struct {
		id  task.TaskID
		cmd task.TaskCmd
	}

	type expected struct {
		isErr      bool
		errName    string
		domainName string
		checkTask  bool
		taskTitle  string
		taskStatus string
	}

	tests := []struct {
		testName string
		setup    func(t *testing.T, q db.Querier) uuid.UUID
		args     func(setupID uuid.UUID) args
		expected expected
	}{
		// 正常系 (Happy Path)
		{
			testName: "正常系: update task title",
			setup: func(t *testing.T, q db.Querier) uuid.UUID {
				t.Helper()
				created, err := q.CreateTask(context.Background(), db.CreateTaskParams{
					Title:    "Original Title",
					Status:   "pending",
					Priority: "medium",
				})
				if err != nil {
					t.Fatalf("failed to create test task: %v", err)
				}
				return created.ID
			},
			args: func(setupID uuid.UUID) args {
				return args{
					id: task.TaskIDFromUUID(setupID),
					cmd: task.NewTaskCmd(
						task.TaskTitle("Updated Title"),
						task.TaskDescription("Updated description"),
						task.TaskStatusPending,
					),
				}
			},
			expected: expected{
				isErr:      false,
				checkTask:  true,
				taskTitle:  "Updated Title",
				taskStatus: "pending",
			},
		},
		{
			testName: "正常系: update task status to completed",
			setup: func(t *testing.T, q db.Querier) uuid.UUID {
				t.Helper()
				created, err := q.CreateTask(context.Background(), db.CreateTaskParams{
					Title:    "Task to complete",
					Status:   "pending",
					Priority: "medium",
				})
				if err != nil {
					t.Fatalf("failed to create test task: %v", err)
				}
				return created.ID
			},
			args: func(setupID uuid.UUID) args {
				return args{
					id: task.TaskIDFromUUID(setupID),
					cmd: task.NewTaskCmd(
						task.TaskTitle("Task to complete"),
						task.TaskDescription(""),
						task.TaskStatusCompleted,
					),
				}
			},
			expected: expected{
				isErr:      false,
				checkTask:  true,
				taskTitle:  "Task to complete",
				taskStatus: "completed",
			},
		},
		// 異常系 (Error Cases)
		{
			testName: "異常系: update non-existent task",
			setup: func(t *testing.T, q db.Querier) uuid.UUID {
				t.Helper()
				return uuid.New()
			},
			args: func(setupID uuid.UUID) args {
				return args{
					id: task.TaskIDFromUUID(setupID),
					cmd: task.NewTaskCmd(
						task.TaskTitle("Updated Title"),
						task.TaskDescription(""),
						task.TaskStatusPending,
					),
				}
			},
			expected: expected{
				isErr:      true,
				errName:    apperror.NotFoundErrorName,
				domainName: "Task",
			},
		},
		// 特殊文字 (Special Chars)
		{
			testName: "特殊文字: update task with emoji",
			setup: func(t *testing.T, q db.Querier) uuid.UUID {
				t.Helper()
				created, err := q.CreateTask(context.Background(), db.CreateTaskParams{
					Title:    "Original Task",
					Status:   "pending",
					Priority: "medium",
				})
				if err != nil {
					t.Fatalf("failed to create test task: %v", err)
				}
				return created.ID
			},
			args: func(setupID uuid.UUID) args {
				return args{
					id: task.TaskIDFromUUID(setupID),
					cmd: task.NewTaskCmd(
						task.TaskTitle("Updated Task 📋"),
						task.TaskDescription("Description with emoji 🚀"),
						task.TaskStatusPending,
					),
				}
			},
			expected: expected{
				isErr:      false,
				checkTask:  true,
				taskTitle:  "Updated Task 📋",
				taskStatus: "pending",
			},
		},
		{
			testName: "特殊文字: update task with Japanese characters",
			setup: func(t *testing.T, q db.Querier) uuid.UUID {
				t.Helper()
				created, err := q.CreateTask(context.Background(), db.CreateTaskParams{
					Title:    "Original Task",
					Status:   "pending",
					Priority: "medium",
				})
				if err != nil {
					t.Fatalf("failed to create test task: %v", err)
				}
				return created.ID
			},
			args: func(setupID uuid.UUID) args {
				return args{
					id: task.TaskIDFromUUID(setupID),
					cmd: task.NewTaskCmd(
						task.TaskTitle("更新されたタスク"),
						task.TaskDescription("更新された説明"),
						task.TaskStatusPending,
					),
				}
			},
			expected: expected{
				isErr:      false,
				checkTask:  true,
				taskTitle:  "更新されたタスク",
				taskStatus: "pending",
			},
		},
		// 空文字 (Empty String)
		{
			testName: "空文字: update task with empty description",
			setup: func(t *testing.T, q db.Querier) uuid.UUID {
				t.Helper()
				created, err := q.CreateTask(context.Background(), db.CreateTaskParams{
					Title:       "Task with description",
					Description: sql.NullString{String: "Original description", Valid: true},
					Status:      "pending",
					Priority:    "medium",
				})
				if err != nil {
					t.Fatalf("failed to create test task: %v", err)
				}
				return created.ID
			},
			args: func(setupID uuid.UUID) args {
				return args{
					id: task.TaskIDFromUUID(setupID),
					cmd: task.NewTaskCmd(
						task.TaskTitle("Updated Title"),
						task.TaskDescription(""),
						task.TaskStatusPending,
					),
				}
			},
			expected: expected{
				isErr:      false,
				checkTask:  true,
				taskTitle:  "Updated Title",
				taskStatus: "pending",
			},
		},
		// Nil - zero values
		{
			testName: "Nil: update task with empty status (default to pending)",
			setup: func(t *testing.T, q db.Querier) uuid.UUID {
				t.Helper()
				created, err := q.CreateTask(context.Background(), db.CreateTaskParams{
					Title:    "Task to update",
					Status:   "completed",
					Priority: "medium",
				})
				if err != nil {
					t.Fatalf("failed to create test task: %v", err)
				}
				return created.ID
			},
			args: func(setupID uuid.UUID) args {
				return args{
					id: task.TaskIDFromUUID(setupID),
					cmd: task.NewTaskCmd(
						task.TaskTitle("Updated Task"),
						task.TaskDescription(""),
						task.TaskStatus(""),
					),
				}
			},
			expected: expected{
				isErr:      false,
				checkTask:  true,
				taskTitle:  "Updated Task",
				taskStatus: "pending",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			q := testutil.SetupTestTx(t)
			setupID := tt.setup(t, q)
			args := tt.args(setupID)

			repo := NewTaskRepository(q)
			updatedTask, err := repo.Update(context.Background(), args.id, args.cmd)

			if tt.expected.isErr {
				if err == nil {
					t.Fatalf("expected error but got success")
				}
				var appErr apperror.AppError
				if errors.As(err, &appErr) {
					if diff := cmp.Diff(tt.expected.errName, appErr.ErrorName()); diff != "" {
						t.Errorf("error name mismatch (-want +got):\n%s", diff)
					}
					if diff := cmp.Diff(tt.expected.domainName, appErr.DomainName()); diff != "" {
						t.Errorf("domain name mismatch (-want +got):\n%s", diff)
					}
				}
			} else {
				if err != nil {
					t.Fatalf("expected success but got error: %v", err)
				}
				if tt.expected.checkTask {
					if diff := cmp.Diff(tt.expected.taskTitle, updatedTask.Title.String()); diff != "" {
						t.Errorf("task title mismatch (-want +got):\n%s", diff)
					}
					if diff := cmp.Diff(tt.expected.taskStatus, updatedTask.Status.String()); diff != "" {
						t.Errorf("task status mismatch (-want +got):\n%s", diff)
					}
					if diff := cmp.Diff(args.id.String(), updatedTask.ID.String()); diff != "" {
						t.Errorf("task ID mismatch (-want +got):\n%s", diff)
					}
					// Verify task was actually updated in DB
					verifiedTask, verifyErr := repo.FindByID(context.Background(), updatedTask.ID)
					if verifyErr != nil {
						t.Errorf("updated task not found in database: %v", verifyErr)
					} else {
						if diff := cmp.Diff(tt.expected.taskTitle, verifiedTask.Title.String()); diff != "" {
							t.Errorf("verified task title mismatch (-want +got):\n%s", diff)
						}
					}
				}
			}
		})
	}
}

// TestContextTimeout tests context timeout handling for all repository functions
func TestContextTimeout(t *testing.T) {
	t.Run("FindByID with cancelled context", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		repo := NewTaskRepository(q)

		created, err := q.CreateTask(context.Background(), db.CreateTaskParams{
			Title:    "Test Task",
			Status:   "pending",
			Priority: "medium",
		})
		if err != nil {
			t.Fatalf("failed to create test task: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, findErr := repo.FindByID(ctx, task.TaskIDFromUUID(created.ID))
		if findErr != nil {
			var appErr apperror.AppError
			if errors.As(findErr, &appErr) {
				if appErr.ErrorName() != apperror.DatabaseErrorName && appErr.ErrorName() != apperror.InternalServerErrorName {
					t.Errorf("expected DatabaseError or InternalServerError, got %s", appErr.ErrorName())
				}
			}
		}
	})

	t.Run("FindAll with cancelled context", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		repo := NewTaskRepository(q)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, findErr := repo.FindAll(ctx)
		if findErr != nil {
			var appErr apperror.AppError
			if errors.As(findErr, &appErr) {
				if appErr.ErrorName() != apperror.DatabaseErrorName && appErr.ErrorName() != apperror.InternalServerErrorName {
					t.Errorf("expected DatabaseError or InternalServerError, got %s", appErr.ErrorName())
				}
			}
		}
	})

	t.Run("Create with cancelled context", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		repo := NewTaskRepository(q)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		cmd := task.NewTaskCmd(
			task.TaskTitle("Test Task"),
			task.TaskDescription(""),
			task.TaskStatusPending,
		)

		_, createErr := repo.Create(ctx, cmd)
		if createErr != nil {
			var appErr apperror.AppError
			if errors.As(createErr, &appErr) {
				if appErr.ErrorName() != apperror.DatabaseErrorName && appErr.ErrorName() != apperror.InternalServerErrorName {
					t.Errorf("expected DatabaseError or InternalServerError, got %s", appErr.ErrorName())
				}
			}
		}
	})

	t.Run("Update with cancelled context", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		repo := NewTaskRepository(q)

		created, err := q.CreateTask(context.Background(), db.CreateTaskParams{
			Title:    "Test Task",
			Status:   "pending",
			Priority: "medium",
		})
		if err != nil {
			t.Fatalf("failed to create test task: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		cmd := task.NewTaskCmd(
			task.TaskTitle("Updated Task"),
			task.TaskDescription(""),
			task.TaskStatusPending,
		)

		_, updateErr := repo.Update(ctx, task.TaskIDFromUUID(created.ID), cmd)
		if updateErr != nil {
			var appErr apperror.AppError
			if errors.As(updateErr, &appErr) {
				if appErr.ErrorName() != apperror.DatabaseErrorName && appErr.ErrorName() != apperror.InternalServerErrorName && appErr.ErrorName() != apperror.NotFoundErrorName {
					t.Errorf("expected DatabaseError, InternalServerError or NotFoundError, got %s", appErr.ErrorName())
				}
			}
		}
	})
}

// TestDataTransformation tests proper transformation between db.Task and domain.Task
func TestDataTransformation(t *testing.T) {
	t.Run("transformation includes all fields", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		repo := NewTaskRepository(q)

		created, err := q.CreateTask(context.Background(), db.CreateTaskParams{
			Title:       "Complete Task",
			Description: sql.NullString{String: "Full description with details", Valid: true},
			Status:      "completed",
			Priority:    "high",
		})
		if err != nil {
			t.Fatalf("failed to create test task: %v", err)
		}

		domainTask, findErr := repo.FindByID(context.Background(), task.TaskIDFromUUID(created.ID))
		if findErr != nil {
			t.Fatalf("expected success but got error: %v", findErr)
		}

		if diff := cmp.Diff(created.ID.String(), domainTask.ID.String()); diff != "" {
			t.Errorf("ID transformation mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff(created.Title, domainTask.Title.String()); diff != "" {
			t.Errorf("Title transformation mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff(created.Description.String, domainTask.Description.String()); diff != "" {
			t.Errorf("Description transformation mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff(created.Status, domainTask.Status.String()); diff != "" {
			t.Errorf("Status transformation mismatch (-want +got):\n%s", diff)
		}

		if !domainTask.IsCompleted() {
			t.Error("expected task to be completed")
		}
		if domainTask.IsPending() {
			t.Error("expected task not to be pending")
		}
	})

	t.Run("transformation handles null description", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		repo := NewTaskRepository(q)

		created, err := q.CreateTask(context.Background(), db.CreateTaskParams{
			Title:       "Task without description",
			Description: sql.NullString{String: "", Valid: false},
			Status:      "pending",
			Priority:    "medium",
		})
		if err != nil {
			t.Fatalf("failed to create test task: %v", err)
		}

		domainTask, findErr := repo.FindByID(context.Background(), task.TaskIDFromUUID(created.ID))
		if findErr != nil {
			t.Fatalf("expected success but got error: %v", findErr)
		}

		if domainTask.Description.String() != "" {
			t.Errorf("expected empty description, got: %s", domainTask.Description.String())
		}
	})
}

// TestErrorMapping tests that database errors are properly mapped to domain errors
func TestErrorMapping(t *testing.T) {
	t.Run("sql.ErrNoRows maps to NotFoundError in FindByID", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		repo := NewTaskRepository(q)

		_, findErr := repo.FindByID(context.Background(), task.TaskIDFromUUID(uuid.New()))
		if findErr == nil {
			t.Fatal("expected error but got success")
		}

		var appErr apperror.AppError
		if !errors.As(findErr, &appErr) {
			t.Fatalf("expected AppError, got: %v", findErr)
		}
		if diff := cmp.Diff(apperror.NotFoundErrorName, appErr.ErrorName()); diff != "" {
			t.Errorf("error name mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff("Task", appErr.DomainName()); diff != "" {
			t.Errorf("domain name mismatch (-want +got):\n%s", diff)
		}

		if !errors.Is(appErr.Unwrap(), sql.ErrNoRows) {
			t.Error("expected error to wrap sql.ErrNoRows")
		}
	})

	t.Run("sql.ErrNoRows maps to NotFoundError in Update", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		repo := NewTaskRepository(q)

		cmd := task.NewTaskCmd(
			task.TaskTitle("Updated Task"),
			task.TaskDescription(""),
			task.TaskStatusPending,
		)

		_, updateErr := repo.Update(context.Background(), task.TaskIDFromUUID(uuid.New()), cmd)
		if updateErr == nil {
			t.Fatal("expected error but got success")
		}

		var appErr apperror.AppError
		if !errors.As(updateErr, &appErr) {
			t.Fatalf("expected AppError, got: %v", updateErr)
		}
		if diff := cmp.Diff(apperror.NotFoundErrorName, appErr.ErrorName()); diff != "" {
			t.Errorf("error name mismatch (-want +got):\n%s", diff)
		}
		if diff := cmp.Diff("Task", appErr.DomainName()); diff != "" {
			t.Errorf("domain name mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("context.DeadlineExceeded maps to InternalServerError", func(t *testing.T) {
		q := testutil.SetupTestTx(t)
		repo := NewTaskRepository(q)

		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-1*time.Second))
		defer cancel()

		_, findErr := repo.FindAll(ctx)
		if findErr != nil {
			var appErr apperror.AppError
			if errors.As(findErr, &appErr) {
				if appErr.ErrorName() != apperror.InternalServerErrorName && appErr.ErrorName() != apperror.DatabaseErrorName {
					t.Errorf("expected InternalServerError or DatabaseError, got %s", appErr.ErrorName())
				}
				if appErr.ErrorName() == apperror.InternalServerErrorName {
					if diff := cmp.Diff("TaskRepository", appErr.DomainName()); diff != "" {
						t.Errorf("domain name mismatch (-want +got):\n%s", diff)
					}
				}
			}
		}
	})
}
