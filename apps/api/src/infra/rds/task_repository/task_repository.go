package task_repository

import (
	"api/src/domain/task"
	"time"
	"utils/db/db"
)

const dbTimeout = 5 * time.Second

// TaskRepository is the PostgreSQL implementation of task.TaskRepository.
type TaskRepository struct {
	q db.Querier
}

// NewTaskRepository creates a new TaskRepository.
func NewTaskRepository(q db.Querier) *TaskRepository {
	return &TaskRepository{q: q}
}

// compile-time check that TaskRepository implements task.TaskRepository
var _ task.TaskRepository = (*TaskRepository)(nil)
