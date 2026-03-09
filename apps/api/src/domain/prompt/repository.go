package prompt

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

// PromptRepository defines persistence operations for prompts.
// Implementations live in the infra layer (dependency inversion).
type PromptRepository interface {
	FindByID(ctx context.Context, id PromptID) (Prompt, error)
	FindByProjectAndSlug(ctx context.Context, projectID uuid.UUID, slug PromptSlug) (Prompt, error)
	FindAllByProject(ctx context.Context, projectID uuid.UUID) ([]Prompt, error)
	Create(ctx context.Context, cmd PromptCmd) (Prompt, error)
	Update(ctx context.Context, id PromptID, cmd PromptCmd) (Prompt, error)
	UpdateLatestVersion(ctx context.Context, id PromptID, version int) (Prompt, error)
	UpdateProductionVersion(ctx context.Context, id PromptID, version *int) (Prompt, error)
}

// VersionRepository defines persistence operations for prompt versions.
// It is also used by service-layer packages (diffservice, lintservice)
// to keep services decoupled from the DB layer.
type VersionRepository interface {
	FindByPromptAndNumber(ctx context.Context, promptID PromptID, number int) (PromptVersion, error)
	FindAllByPrompt(ctx context.Context, promptID PromptID) ([]PromptVersion, error)
	FindLatest(ctx context.Context, promptID PromptID) (PromptVersion, error)
	FindProduction(ctx context.Context, promptID PromptID) (PromptVersion, error)
	Create(ctx context.Context, cmd VersionCmd, versionNumber int) (PromptVersion, error)
	UpdateStatus(ctx context.Context, id PromptVersionID, status VersionStatus) (PromptVersion, error)
	ArchiveProduction(ctx context.Context, promptID PromptID) error
	UpdateLintResult(ctx context.Context, id PromptVersionID, result json.RawMessage) error
	UpdateSemanticDiff(ctx context.Context, id PromptVersionID, diff json.RawMessage) error
}
