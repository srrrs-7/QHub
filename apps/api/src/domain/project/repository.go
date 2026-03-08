package project

import (
	"context"

	"github.com/google/uuid"
)

type ProjectRepository interface {
	FindByID(ctx context.Context, id ProjectID) (Project, error)
	FindByOrgAndSlug(ctx context.Context, orgID uuid.UUID, slug ProjectSlug) (Project, error)
	FindAllByOrg(ctx context.Context, orgID uuid.UUID) ([]Project, error)
	Create(ctx context.Context, cmd ProjectCmd) (Project, error)
	Update(ctx context.Context, id ProjectID, cmd ProjectCmd) (Project, error)
	Delete(ctx context.Context, id ProjectID) error
}
