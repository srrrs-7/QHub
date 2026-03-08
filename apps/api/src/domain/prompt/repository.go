package prompt

import (
	"context"

	"github.com/google/uuid"
)

type PromptRepository interface {
	FindByID(ctx context.Context, id PromptID) (Prompt, error)
	FindByProjectAndSlug(ctx context.Context, projectID uuid.UUID, slug PromptSlug) (Prompt, error)
	FindAllByProject(ctx context.Context, projectID uuid.UUID) ([]Prompt, error)
	Create(ctx context.Context, cmd PromptCmd) (Prompt, error)
	Update(ctx context.Context, id PromptID, cmd PromptCmd) (Prompt, error)
	UpdateLatestVersion(ctx context.Context, id PromptID, version int) (Prompt, error)
	UpdateProductionVersion(ctx context.Context, id PromptID, version *int) (Prompt, error)
}

type VersionRepository interface {
	FindByPromptAndNumber(ctx context.Context, promptID PromptID, number int) (PromptVersion, error)
	FindAllByPrompt(ctx context.Context, promptID PromptID) ([]PromptVersion, error)
	FindLatest(ctx context.Context, promptID PromptID) (PromptVersion, error)
	FindProduction(ctx context.Context, promptID PromptID) (PromptVersion, error)
	Create(ctx context.Context, cmd VersionCmd, versionNumber int) (PromptVersion, error)
	UpdateStatus(ctx context.Context, id PromptVersionID, status VersionStatus) (PromptVersion, error)
	ArchiveProduction(ctx context.Context, promptID PromptID) error
}
