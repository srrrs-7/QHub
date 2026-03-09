package project_repository

import (
	"api/src/domain/project"
	"api/src/infra/rds/repoerr"
	"context"

	"github.com/google/uuid"

	"utils/db/db"
)

func toProject(p db.Project) project.Project {
	return project.NewProject(
		project.ProjectIDFromUUID(p.ID),
		p.OrganizationID,
		project.ProjectName(p.Name),
		project.ProjectSlug(p.Slug),
		project.ProjectDescription(p.Description.String),
	)
}

func (r *ProjectRepository) FindByID(ctx context.Context, id project.ProjectID) (project.Project, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	p, err := r.q.GetProject(ctx, id.UUID())
	if err != nil {
		return project.Project{}, repoerr.Handle(err, "ProjectRepository", "Project")
	}
	return toProject(p), nil
}

func (r *ProjectRepository) FindByOrgAndSlug(ctx context.Context, orgID uuid.UUID, slug project.ProjectSlug) (project.Project, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	p, err := r.q.GetProjectByOrgAndSlug(ctx, db.GetProjectByOrgAndSlugParams{
		OrganizationID: orgID,
		Slug:           slug.String(),
	})
	if err != nil {
		return project.Project{}, repoerr.Handle(err, "ProjectRepository", "Project")
	}
	return toProject(p), nil
}

func (r *ProjectRepository) FindAllByOrg(ctx context.Context, orgID uuid.UUID) ([]project.Project, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	projects, err := r.q.ListProjectsByOrganization(ctx, orgID)
	if err != nil {
		return nil, repoerr.Handle(err, "ProjectRepository", "")
	}

	result := make([]project.Project, 0, len(projects))
	for _, p := range projects {
		result = append(result, toProject(p))
	}
	return result, nil
}
