package projects

import (
	"api/src/domain/project"
)

type ProjectHandler struct {
	repo project.ProjectRepository
}

func NewProjectHandler(repo project.ProjectRepository) *ProjectHandler {
	return &ProjectHandler{repo: repo}
}
