package task_test

import (
	"api/src/domain/task"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTaskCmd(t *testing.T) {
	type args struct {
		title       task.TaskTitle
		description task.TaskDescription
		status      task.TaskStatus
	}
	type expected struct {
		title       task.TaskTitle
		description task.TaskDescription
		status      task.TaskStatus
	}

	tests := []struct {
		testName string
		args     args
		expected expected
	}{
		// 正常系
		{
			testName: "valid task command with all fields",
			args: args{
				title:       task.TaskTitle("Test Task"),
				description: task.TaskDescription("Test Description"),
				status:      task.TaskStatusPending,
			},
			expected: expected{
				title:       task.TaskTitle("Test Task"),
				description: task.TaskDescription("Test Description"),
				status:      task.TaskStatusPending,
			},
		},
		{
			testName: "valid task command with completed status",
			args: args{
				title:       task.TaskTitle("Completed Task"),
				description: task.TaskDescription("Completed Description"),
				status:      task.TaskStatusCompleted,
			},
			expected: expected{
				title:       task.TaskTitle("Completed Task"),
				description: task.TaskDescription("Completed Description"),
				status:      task.TaskStatusCompleted,
			},
		},
		// 境界値
		{
			testName: "task command with long title",
			args: args{
				title:       task.TaskTitle("Very long title that contains many characters to test boundary conditions"),
				description: task.TaskDescription("Description"),
				status:      task.TaskStatusPending,
			},
			expected: expected{
				title:       task.TaskTitle("Very long title that contains many characters to test boundary conditions"),
				description: task.TaskDescription("Description"),
				status:      task.TaskStatusPending,
			},
		},
		// 特殊文字
		{
			testName: "task command with emoji in title",
			args: args{
				title:       task.TaskTitle("Task with emoji 📋✅"),
				description: task.TaskDescription("Description with emoji 📝"),
				status:      task.TaskStatusPending,
			},
			expected: expected{
				title:       task.TaskTitle("Task with emoji 📋✅"),
				description: task.TaskDescription("Description with emoji 📝"),
				status:      task.TaskStatusPending,
			},
		},
		{
			testName: "task command with Japanese characters",
			args: args{
				title:       task.TaskTitle("タスクのタイトル"),
				description: task.TaskDescription("タスクの説明"),
				status:      task.TaskStatusPending,
			},
			expected: expected{
				title:       task.TaskTitle("タスクのタイトル"),
				description: task.TaskDescription("タスクの説明"),
				status:      task.TaskStatusPending,
			},
		},
		// 空文字
		{
			testName: "task command with empty strings",
			args: args{
				title:       task.TaskTitle(""),
				description: task.TaskDescription(""),
				status:      task.TaskStatusPending,
			},
			expected: expected{
				title:       task.TaskTitle(""),
				description: task.TaskDescription(""),
				status:      task.TaskStatusPending,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			cmd := task.NewTaskCmd(tt.args.title, tt.args.description, tt.args.status)

			assert.Equal(t, tt.expected.title, cmd.Title)
			assert.Equal(t, tt.expected.description, cmd.Description)
			assert.Equal(t, tt.expected.status, cmd.Status)
		})
	}
}
