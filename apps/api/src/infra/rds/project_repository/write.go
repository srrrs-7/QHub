package project_repository

import (
	"api/src/domain/project"
	"api/src/infra/rds/repoerr"
	"context"
	"database/sql"
	"utils/db/db"

	"github.com/google/uuid"
)

func (r *ProjectRepository) Create(ctx context.Context, cmd project.ProjectCmd) (project.Project, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	desc := string(cmd.Description)
	p, err := r.q.CreateProject(ctx, db.CreateProjectParams{
		OrganizationID: cmd.OrganizationID,
		Name:           cmd.Name.String(),
		Slug:           cmd.Slug.String(),
		Description:    sql.NullString{String: desc, Valid: desc != ""},
	})
	if err != nil {
		return project.Project{}, repoerr.Handle(err, "ProjectRepository", "")
	}
	return toProject(p), nil
}

func (r *ProjectRepository) Update(ctx context.Context, id project.ProjectID, cmd project.ProjectCmd) (project.Project, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	name := cmd.Name.String()
	slug := cmd.Slug.String()
	desc := string(cmd.Description)

	p, err := r.q.UpdateProject(ctx, db.UpdateProjectParams{
		ID:          id.UUID(),
		Name:        sql.NullString{String: name, Valid: name != ""},
		Slug:        sql.NullString{String: slug, Valid: slug != ""},
		Description: sql.NullString{String: desc, Valid: desc != ""},
	})
	if err != nil {
		return project.Project{}, repoerr.Handle(err, "ProjectRepository", "Project")
	}
	return toProject(p), nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id project.ProjectID) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := r.q.DeleteProject(ctx, uuid.UUID(id))
	if err != nil {
		return repoerr.Handle(err, "ProjectRepository", "")
	}
	return nil
}
