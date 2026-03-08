// Package task_repository implements the task.TaskRepository interface
// using PostgreSQL via sqlc-generated queries.
package task_repository

import (
	"api/src/domain/task"
	"time"
	"utils/db/db"
)

// dbTimeout is the maximum duration for a single DB operation.
const dbTimeout = 5 * time.Second

// TaskRepository is the PostgreSQL implementation of task.TaskRepository.
type TaskRepository struct {
	q db.Querier
}

// NewTaskRepository creates a new TaskRepository backed by the given querier.
func NewTaskRepository(q db.Querier) *TaskRepository {
	return &TaskRepository{q: q}
}

// Compile-time check that TaskRepository implements the domain interface.
var _ task.TaskRepository = (*TaskRepository)(nil)
