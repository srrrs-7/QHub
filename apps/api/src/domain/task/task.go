package task

import (
	"api/src/domain/apperror"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// TaskID represents a unique identifier for a task.
type TaskID uuid.UUID

// NewTaskID creates a new TaskID from a string representation of a UUID.
func NewTaskID(id string) (TaskID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return TaskID{}, apperror.NewValidationError(fmt.Errorf("invalid task ID: %w", err), "TaskID")
	}
	return TaskID(parsed), nil
}

// TaskIDFromUUID creates a TaskID directly from a uuid.UUID.
// Use this when you already have a validated UUID (e.g., from the database).
func TaskIDFromUUID(id uuid.UUID) TaskID {
	return TaskID(id)
}

// String returns the string representation of the TaskID.
func (t TaskID) String() string {
	return uuid.UUID(t).String()
}

// TaskTitle represents the title of a task.
type TaskTitle string

// NewTaskTitle creates a new TaskTitle with validation.
// Title must be between 3 and 100 characters and not whitespace-only.
func NewTaskTitle(title string) (TaskTitle, error) {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return "", apperror.NewValidationError(fmt.Errorf("title must not be empty"), "TaskTitle")
	}
	if len(title) < 3 {
		return "", apperror.NewValidationError(fmt.Errorf("title must be at least 3 characters"), "TaskTitle")
	}
	if len(title) > 100 {
		return "", apperror.NewValidationError(fmt.Errorf("title must be at most 100 characters"), "TaskTitle")
	}
	return TaskTitle(title), nil
}

// String returns the string representation of the TaskTitle.
func (t TaskTitle) String() string {
	return string(t)
}

// TaskDescription represents the detailed description of a task.
type TaskDescription string

// NewTaskDescription creates a new TaskDescription with validation.
// Description must be at most 500 characters. Empty is allowed.
func NewTaskDescription(description string) (TaskDescription, error) {
	if len(description) > 500 {
		return "", apperror.NewValidationError(fmt.Errorf("description must be at most 500 characters"), "TaskDescription")
	}
	return TaskDescription(description), nil
}

// String returns the string representation of the TaskDescription.
func (t TaskDescription) String() string {
	return string(t)
}

// TaskStatus represents the status of a task.
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusCompleted TaskStatus = "completed"
)

// NewTaskStatus creates a new TaskStatus with validation.
func NewTaskStatus(status string) (TaskStatus, error) {
	switch TaskStatus(status) {
	case TaskStatusPending, TaskStatusCompleted:
		return TaskStatus(status), nil
	default:
		return "", apperror.NewValidationError(fmt.Errorf("invalid status: %s (must be 'pending' or 'completed')", status), "TaskStatus")
	}
}

// String returns the string representation of the TaskStatus.
func (t TaskStatus) String() string {
	return string(t)
}

// Task represents a task entity in the domain model.
type Task struct {
	ID          TaskID
	Title       TaskTitle
	Description TaskDescription
	Status      TaskStatus
}

// NewTask creates a new Task with the given properties.
func NewTask(id TaskID, title TaskTitle, description TaskDescription, status TaskStatus) Task {
	return Task{
		ID:          id,
		Title:       title,
		Description: description,
		Status:      status,
	}
}

func (t Task) IsCompleted() bool {
	return t.Status == TaskStatusCompleted
}

func (t Task) IsPending() bool {
	return t.Status == TaskStatusPending
}
