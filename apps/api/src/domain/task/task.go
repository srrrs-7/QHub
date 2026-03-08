// Package task defines the Task aggregate and its value objects.
//
// A Task is a simple work item with a title, description, and status.
// It is used as the reference example for the clean-architecture pattern
// in this codebase.
package task

import (
	"fmt"

	"api/src/domain/apperror"
	"api/src/domain/valobj"

	"github.com/google/uuid"
)

// --- TaskID ---

// TaskID is the unique identifier for a task (UUID).
type TaskID uuid.UUID

// NewTaskID parses a string UUID into a TaskID.
func NewTaskID(id string) (TaskID, error) {
	parsed, err := valobj.ParseUUID(id, "TaskID")
	if err != nil {
		return TaskID{}, err
	}
	return TaskID(parsed), nil
}

// TaskIDFromUUID converts a uuid.UUID directly (for DB results).
func TaskIDFromUUID(id uuid.UUID) TaskID { return TaskID(id) }

// String returns the string representation.
func (t TaskID) String() string { return uuid.UUID(t).String() }

// --- TaskTitle ---

// TaskTitle is a validated title (3–100 characters, non-blank).
type TaskTitle string

// NewTaskTitle validates and creates a TaskTitle.
func NewTaskTitle(title string) (TaskTitle, error) {
	if err := valobj.ValidateName(title, 3, 100, "TaskTitle"); err != nil {
		return "", err
	}
	return TaskTitle(title), nil
}

// String returns the title as a plain string.
func (t TaskTitle) String() string { return string(t) }

// --- TaskDescription ---

// TaskDescription is an optional description (max 500 characters).
type TaskDescription string

// NewTaskDescription validates and creates a TaskDescription.
func NewTaskDescription(description string) (TaskDescription, error) {
	if err := valobj.ValidateMaxLength(description, 500, "TaskDescription"); err != nil {
		return "", err
	}
	return TaskDescription(description), nil
}

// String returns the description as a plain string.
func (t TaskDescription) String() string { return string(t) }

// --- TaskStatus ---

// TaskStatus represents the completion state of a task.
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusCompleted TaskStatus = "completed"
)

// NewTaskStatus validates a status string.
func NewTaskStatus(status string) (TaskStatus, error) {
	switch TaskStatus(status) {
	case TaskStatusPending, TaskStatusCompleted:
		return TaskStatus(status), nil
	default:
		return "", apperror.NewValidationError(fmt.Errorf("invalid status: %s (must be 'pending' or 'completed')", status), "TaskStatus")
	}
}

// String returns the status as a plain string.
func (t TaskStatus) String() string { return string(t) }

// --- Task (Aggregate) ---

// Task is the aggregate root representing a work item.
type Task struct {
	ID          TaskID
	Title       TaskTitle
	Description TaskDescription
	Status      TaskStatus
}

// NewTask constructs a Task from validated value objects.
func NewTask(id TaskID, title TaskTitle, description TaskDescription, status TaskStatus) Task {
	return Task{ID: id, Title: title, Description: description, Status: status}
}

// IsCompleted returns true if the task is marked as completed.
func (t Task) IsCompleted() bool { return t.Status == TaskStatusCompleted }

// IsPending returns true if the task is still pending.
func (t Task) IsPending() bool { return t.Status == TaskStatusPending }
