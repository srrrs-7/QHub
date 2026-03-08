package task

// TaskCmd represents a command to create or update a task.
// It contains only the mutable properties of a task (excluding ID).
type TaskCmd struct {
	Title       TaskTitle
	Description TaskDescription
	Status      TaskStatus
}

// NewTaskCmd creates a new TaskCmd with the given properties.
func NewTaskCmd(title TaskTitle, description TaskDescription, status TaskStatus) TaskCmd {
	return TaskCmd{
		Title:       title,
		Description: description,
		Status:      status,
	}
}
