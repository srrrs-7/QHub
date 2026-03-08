package project_repository

import (
	"api/src/domain/project"
	"time"
	"utils/db/db"
)

const dbTimeout = 5 * time.Second

type ProjectRepository struct {
	q db.Querier
}

func NewProjectRepository(q db.Querier) *ProjectRepository {
	return &ProjectRepository{q: q}
}

var _ project.ProjectRepository = (*ProjectRepository)(nil)
