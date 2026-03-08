package task_test

import (
	"api/src/domain/task"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestNewTaskID tests the NewTaskID constructor
func TestNewTaskID(t *testing.T) {
	type args struct {
		id string
	}
	type expected struct {
		shouldPanic bool
		value       string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "valid UUID lowercase",
			args:     args{id: "123e4567-e89b-12d3-a456-426614174000"},
			expected: expected{shouldPanic: false, value: "123e4567-e89b-12d3-a456-426614174000"},
		},
		{
			testName: "valid UUID uppercase",
			args:     args{id: "123E4567-E89B-12D3-A456-426614174000"},
			expected: expected{shouldPanic: false, value: "123e4567-e89b-12d3-a456-426614174000"},
		},
		{
			testName: "valid UUID mixed case",
			args:     args{id: "a1b2c3d4-E5f6-7890-AbCd-Ef1234567890"},
			expected: expected{shouldPanic: false, value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"},
		},

		// ❌ 異常系 - CRITICAL: これらはPANICする！
		{
			testName: "invalid UUID - too short",
			args:     args{id: "123"},
			expected: expected{shouldPanic: true},
		},
		{
			testName: "invalid UUID - malformed",
			args:     args{id: "not-a-uuid"},
			expected: expected{shouldPanic: true},
		},
		{
			testName: "UUID without dashes is valid",
			args:     args{id: "123e4567e89b12d3a456426614174000"},
			expected: expected{shouldPanic: false, value: "123e4567-e89b-12d3-a456-426614174000"},
		},
		{
			testName: "invalid UUID - invalid characters",
			args:     args{id: "gggggggg-gggg-gggg-gggg-gggggggggggg"},
			expected: expected{shouldPanic: true},
		},

		// 📭 空文字 - CRITICAL: PANIC
		{
			testName: "empty string",
			args:     args{id: ""},
			expected: expected{shouldPanic: true},
		},
		{
			testName: "whitespace only",
			args:     args{id: "   "},
			expected: expected{shouldPanic: true},
		},

		// 🔤 特殊文字
		{
			testName: "SQL injection attempt",
			args:     args{id: "'; DROP TABLE tasks; --"},
			expected: expected{shouldPanic: true},
		},
		{
			testName: "emoji in UUID",
			args:     args{id: "123e4567-e89b-12d3-a456-42661417400🔥"},
			expected: expected{shouldPanic: true},
		},
		{
			testName: "Japanese characters",
			args:     args{id: "タスク-ID-です-よね-はいそうです"},
			expected: expected{shouldPanic: true},
		},

		// 📏 境界値
		{
			testName: "UUID with all zeros",
			args:     args{id: "00000000-0000-0000-0000-000000000000"},
			expected: expected{shouldPanic: false, value: "00000000-0000-0000-0000-000000000000"},
		},
		{
			testName: "UUID with all f's",
			args:     args{id: "ffffffff-ffff-ffff-ffff-ffffffffffff"},
			expected: expected{shouldPanic: false, value: "ffffffff-ffff-ffff-ffff-ffffffffffff"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if tt.expected.shouldPanic {
				assert.Panics(t, func() {
					_ = task.NewTaskID(tt.args.id)
				}, "NewTaskID should panic for invalid UUID")
			} else {
				assert.NotPanics(t, func() {
					result := task.NewTaskID(tt.args.id)
					assert.Equal(t, tt.expected.value, result.String())
				}, "NewTaskID should not panic for valid UUID")
			}
		})
	}
}

// TestTaskID_String tests the TaskID String method
func TestTaskID_String(t *testing.T) {
	type args struct {
		id string
	}
	type expected struct {
		value string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "convert TaskID to string",
			args:     args{id: "123e4567-e89b-12d3-a456-426614174000"},
			expected: expected{value: "123e4567-e89b-12d3-a456-426614174000"},
		},
		{
			testName: "uppercase UUID converts to lowercase",
			args:     args{id: "123E4567-E89B-12D3-A456-426614174000"},
			expected: expected{value: "123e4567-e89b-12d3-a456-426614174000"},
		},

		// 📏 境界値
		{
			testName: "all zeros UUID",
			args:     args{id: "00000000-0000-0000-0000-000000000000"},
			expected: expected{value: "00000000-0000-0000-0000-000000000000"},
		},
		{
			testName: "all f's UUID",
			args:     args{id: "ffffffff-ffff-ffff-ffff-ffffffffffff"},
			expected: expected{value: "ffffffff-ffff-ffff-ffff-ffffffffffff"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			taskID := task.NewTaskID(tt.args.id)
			assert.Equal(t, tt.expected.value, taskID.String())
		})
	}
}

// TestTaskTitle_String tests the TaskTitle String method
func TestTaskTitle_String(t *testing.T) {
	type args struct {
		title string
	}
	type expected struct {
		value string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "normal task title",
			args:     args{title: "Buy groceries"},
			expected: expected{value: "Buy groceries"},
		},
		{
			testName: "task title with numbers",
			args:     args{title: "Task 123"},
			expected: expected{value: "Task 123"},
		},

		// 🔤 特殊文字
		{
			testName: "title with emoji",
			args:     args{title: "Task 📋 with emoji ✅"},
			expected: expected{value: "Task 📋 with emoji ✅"},
		},
		{
			testName: "title with Japanese",
			args:     args{title: "タスクのタイトル"},
			expected: expected{value: "タスクのタイトル"},
		},
		{
			testName: "title with special characters",
			args:     args{title: "Task: Do this & that!"},
			expected: expected{value: "Task: Do this & that!"},
		},

		// 📭 空文字
		{
			testName: "empty title",
			args:     args{title: ""},
			expected: expected{value: ""},
		},
		{
			testName: "whitespace only title",
			args:     args{title: "   "},
			expected: expected{value: "   "},
		},

		// 📏 境界値
		{
			testName: "very long title",
			args:     args{title: "This is a very long task title that contains many characters to test boundary conditions and ensure proper handling of long strings"},
			expected: expected{value: "This is a very long task title that contains many characters to test boundary conditions and ensure proper handling of long strings"},
		},
		{
			testName: "single character",
			args:     args{title: "A"},
			expected: expected{value: "A"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			title := task.TaskTitle(tt.args.title)
			assert.Equal(t, tt.expected.value, title.String())
		})
	}
}

// TestTaskDescription_String tests the TaskDescription String method
func TestTaskDescription_String(t *testing.T) {
	type args struct {
		description string
	}
	type expected struct {
		value string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "normal description",
			args:     args{description: "This is a task description"},
			expected: expected{value: "This is a task description"},
		},
		{
			testName: "multiline description",
			args:     args{description: "Line 1\nLine 2\nLine 3"},
			expected: expected{value: "Line 1\nLine 2\nLine 3"},
		},

		// 🔤 特殊文字
		{
			testName: "description with emoji",
			args:     args{description: "Description with 📝 emoji"},
			expected: expected{value: "Description with 📝 emoji"},
		},
		{
			testName: "description with Japanese",
			args:     args{description: "タスクの説明です"},
			expected: expected{value: "タスクの説明です"},
		},

		// 📭 空文字
		{
			testName: "empty description",
			args:     args{description: ""},
			expected: expected{value: ""},
		},

		// 📏 境界値
		{
			testName: "very long description",
			args:     args{description: "This is a very long description that contains many characters and multiple sentences to test boundary conditions and ensure proper handling of long text content in task descriptions."},
			expected: expected{value: "This is a very long description that contains many characters and multiple sentences to test boundary conditions and ensure proper handling of long text content in task descriptions."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			description := task.TaskDescription(tt.args.description)
			assert.Equal(t, tt.expected.value, description.String())
		})
	}
}

// TestTaskStatus_String tests the TaskStatus String method
func TestTaskStatus_String(t *testing.T) {
	type args struct {
		status task.TaskStatus
	}
	type expected struct {
		value string
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "pending status",
			args:     args{status: task.TaskStatusPending},
			expected: expected{value: "pending"},
		},
		{
			testName: "completed status",
			args:     args{status: task.TaskStatusCompleted},
			expected: expected{value: "completed"},
		},

		// ❌ 異常系 - custom status (not recommended but possible)
		{
			testName: "custom status",
			args:     args{status: task.TaskStatus("in-progress")},
			expected: expected{value: "in-progress"},
		},

		// 📭 空文字
		{
			testName: "empty status",
			args:     args{status: task.TaskStatus("")},
			expected: expected{value: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			assert.Equal(t, tt.expected.value, tt.args.status.String())
		})
	}
}

// TestNewTask tests the NewTask constructor
func TestNewTask(t *testing.T) {
	type args struct {
		id          string
		title       string
		description string
		status      task.TaskStatus
	}
	type expected struct {
		title       string
		description string
		status      task.TaskStatus
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "create task with all fields",
			args: args{
				id:          "123e4567-e89b-12d3-a456-426614174000",
				title:       "Buy groceries",
				description: "Buy milk, eggs, and bread",
				status:      task.TaskStatusPending,
			},
			expected: expected{
				title:       "Buy groceries",
				description: "Buy milk, eggs, and bread",
				status:      task.TaskStatusPending,
			},
		},
		{
			testName: "create completed task",
			args: args{
				id:          "00000000-0000-0000-0000-000000000000",
				title:       "Completed task",
				description: "This is done",
				status:      task.TaskStatusCompleted,
			},
			expected: expected{
				title:       "Completed task",
				description: "This is done",
				status:      task.TaskStatusCompleted,
			},
		},

		// 🔤 特殊文字
		{
			testName: "task with emoji",
			args: args{
				id:          "ffffffff-ffff-ffff-ffff-ffffffffffff",
				title:       "Task 📋",
				description: "Description 📝",
				status:      task.TaskStatusPending,
			},
			expected: expected{
				title:       "Task 📋",
				description: "Description 📝",
				status:      task.TaskStatusPending,
			},
		},
		{
			testName: "task with Japanese",
			args: args{
				id:          "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
				title:       "タスク",
				description: "説明",
				status:      task.TaskStatusPending,
			},
			expected: expected{
				title:       "タスク",
				description: "説明",
				status:      task.TaskStatusPending,
			},
		},

		// 📭 空文字
		{
			testName: "task with empty title and description",
			args: args{
				id:          "123e4567-e89b-12d3-a456-426614174000",
				title:       "",
				description: "",
				status:      task.TaskStatusPending,
			},
			expected: expected{
				title:       "",
				description: "",
				status:      task.TaskStatusPending,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			taskID := task.NewTaskID(tt.args.id)
			result := task.NewTask(
				taskID,
				task.TaskTitle(tt.args.title),
				task.TaskDescription(tt.args.description),
				tt.args.status,
			)

			assert.Equal(t, tt.args.id, result.ID.String())
			assert.Equal(t, tt.expected.title, result.Title.String())
			assert.Equal(t, tt.expected.description, result.Description.String())
			assert.Equal(t, tt.expected.status, result.Status)
		})
	}
}

// TestTask_IsCompleted tests the IsCompleted method
func TestTask_IsCompleted(t *testing.T) {
	validUUID := uuid.New().String()

	type args struct {
		status task.TaskStatus
	}
	type expected struct {
		isCompleted bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "completed task returns true",
			args:     args{status: task.TaskStatusCompleted},
			expected: expected{isCompleted: true},
		},

		// ❌ 異常系
		{
			testName: "pending task returns false",
			args:     args{status: task.TaskStatusPending},
			expected: expected{isCompleted: false},
		},
		{
			testName: "custom status returns false",
			args:     args{status: task.TaskStatus("in-progress")},
			expected: expected{isCompleted: false},
		},

		// 📭 空文字
		{
			testName: "empty status returns false",
			args:     args{status: task.TaskStatus("")},
			expected: expected{isCompleted: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			testTask := task.NewTask(
				task.NewTaskID(validUUID),
				task.TaskTitle("Test"),
				task.TaskDescription("Test"),
				tt.args.status,
			)

			assert.Equal(t, tt.expected.isCompleted, testTask.IsCompleted())
		})
	}
}

// TestTask_IsPending tests the IsPending method
func TestTask_IsPending(t *testing.T) {
	validUUID := uuid.New().String()

	type args struct {
		status task.TaskStatus
	}
	type expected struct {
		isPending bool
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// ✅ 正常系
		{
			testName: "pending task returns true",
			args:     args{status: task.TaskStatusPending},
			expected: expected{isPending: true},
		},

		// ❌ 異常系
		{
			testName: "completed task returns false",
			args:     args{status: task.TaskStatusCompleted},
			expected: expected{isPending: false},
		},
		{
			testName: "custom status returns false",
			args:     args{status: task.TaskStatus("in-progress")},
			expected: expected{isPending: false},
		},

		// 📭 空文字
		{
			testName: "empty status returns false",
			args:     args{status: task.TaskStatus("")},
			expected: expected{isPending: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			testTask := task.NewTask(
				task.NewTaskID(validUUID),
				task.TaskTitle("Test"),
				task.TaskDescription("Test"),
				tt.args.status,
			)

			assert.Equal(t, tt.expected.isPending, testTask.IsPending())
		})
	}
}

// TestTaskStatus_Constants tests that status constants have correct values
func TestTaskStatus_Constants(t *testing.T) {
	tests := []struct {
		testName string
		status   task.TaskStatus
		expected string
	}{
		{
			testName: "TaskStatusPending is 'pending'",
			status:   task.TaskStatusPending,
			expected: "pending",
		},
		{
			testName: "TaskStatusCompleted is 'completed'",
			status:   task.TaskStatusCompleted,
			expected: "completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.status))
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}
