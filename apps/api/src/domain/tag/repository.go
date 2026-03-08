package tag

import (
	"context"

	"github.com/google/uuid"
)

type TagRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (Tag, error)
	FindAllByOrg(ctx context.Context, orgID uuid.UUID) ([]Tag, error)
	Create(ctx context.Context, tag Tag) (Tag, error)
	Delete(ctx context.Context, id uuid.UUID) error
	AddToPrompt(ctx context.Context, promptID, tagID uuid.UUID) error
	RemoveFromPrompt(ctx context.Context, promptID, tagID uuid.UUID) error
	FindByPrompt(ctx context.Context, promptID uuid.UUID) ([]Tag, error)
}
