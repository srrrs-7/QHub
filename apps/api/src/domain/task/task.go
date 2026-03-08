package task

import "github.com/google/uuid"

// TaskID represents a unique identifier for a task.
// It wraps a UUID to ensure type safety.
type TaskID uuid.UUID

// NewTaskID creates a new TaskID from a string representation of a UUID.
// It panics if the provided string is not a valid UUID.
func NewTaskID(id string) TaskID {
	return TaskID(uuid.MustParse(id))
}

// String returns the string representation of the TaskID.
func (t TaskID) String() string {
	return uuid.UUID(t).String()
}

// TaskTitle represents the title of a task.
// It should be a descriptive name for the task.
type TaskTitle string

// String returns the string representation of the TaskTitle.
func (t TaskTitle) String() string {
	return string(t)
}

// TaskDescription represents the detailed description of a task.
// It provides additional context and information about the task.
type TaskDescription string

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

// String returns the string representation of the TaskStatus.
func (t TaskStatus) String() string {
	return string(t)
}

// Task represents a task entity in the domain model.
// It contains all the properties that define a task.
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
