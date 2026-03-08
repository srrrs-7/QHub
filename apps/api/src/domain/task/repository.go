package task

import "context"

// TaskRepository defines the interface for task persistence operations.
// This interface is defined in the domain layer to invert the dependency:
// handlers depend on this interface, infra layer implements it.
type TaskRepository interface {
	FindByID(ctx context.Context, id TaskID) (Task, error)
	FindAll(ctx context.Context) ([]Task, error)
	Create(ctx context.Context, cmd TaskCmd) (Task, error)
	Update(ctx context.Context, id TaskID, cmd TaskCmd) (Task, error)
}
