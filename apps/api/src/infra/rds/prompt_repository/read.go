package prompt_repository

import (
	"api/src/domain/prompt"
	"api/src/infra/rds/repoerr"
	"context"
	"utils/db/db"

	"github.com/google/uuid"
)

func toPrompt(p db.Prompt) prompt.Prompt {
	var prodVersion *int
	if p.ProductionVersion.Valid {
		v := int(p.ProductionVersion.Int32)
		prodVersion = &v
	}
	return prompt.NewPrompt(
		prompt.PromptIDFromUUID(p.ID),
		p.ProjectID,
		prompt.PromptName(p.Name),
		prompt.PromptSlug(p.Slug),
		prompt.PromptType(p.PromptType),
		prompt.PromptDescription(p.Description.String),
		int(p.LatestVersion),
		prodVersion,
	)
}

func toVersion(v db.PromptVersion) prompt.PromptVersion {
	return prompt.PromptVersion{
		ID:                prompt.PromptVersionIDFromUUID(v.ID),
		PromptID:          prompt.PromptIDFromUUID(v.PromptID),
		VersionNumber:     int(v.VersionNumber),
		Status:            prompt.VersionStatus(v.Status),
		Content:           v.Content,
		Variables:         v.Variables.RawMessage,
		ChangeDescription: prompt.ChangeDescription(v.ChangeDescription.String),
		AuthorID:          v.AuthorID,
	}
}

// --- PromptRepository ---

func (r *PromptRepository) FindByID(ctx context.Context, id prompt.PromptID) (prompt.Prompt, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	p, err := r.q.GetPrompt(ctx, id.UUID())
	if err != nil {
		return prompt.Prompt{}, repoerr.Handle(err, "PromptRepository", "Prompt")
	}
	return toPrompt(p), nil
}

func (r *PromptRepository) FindByProjectAndSlug(ctx context.Context, projectID uuid.UUID, slug prompt.PromptSlug) (prompt.Prompt, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	p, err := r.q.GetPromptByProjectAndSlug(ctx, db.GetPromptByProjectAndSlugParams{
		ProjectID: projectID,
		Slug:      slug.String(),
	})
	if err != nil {
		return prompt.Prompt{}, repoerr.Handle(err, "PromptRepository", "Prompt")
	}
	return toPrompt(p), nil
}

func (r *PromptRepository) FindAllByProject(ctx context.Context, projectID uuid.UUID) ([]prompt.Prompt, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	prompts, err := r.q.ListPromptsByProject(ctx, projectID)
	if err != nil {
		return nil, repoerr.Handle(err, "PromptRepository", "")
	}

	result := make([]prompt.Prompt, 0, len(prompts))
	for _, p := range prompts {
		result = append(result, toPrompt(p))
	}
	return result, nil
}

// --- VersionRepository ---

func (r *VersionRepository) FindByPromptAndNumber(ctx context.Context, promptID prompt.PromptID, number int) (prompt.PromptVersion, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	v, err := r.q.GetPromptVersion(ctx, db.GetPromptVersionParams{
		PromptID:      promptID.UUID(),
		VersionNumber: int32(number),
	})
	if err != nil {
		return prompt.PromptVersion{}, repoerr.Handle(err, "VersionRepository", "PromptVersion")
	}
	return toVersion(v), nil
}

func (r *VersionRepository) FindAllByPrompt(ctx context.Context, promptID prompt.PromptID) ([]prompt.PromptVersion, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	versions, err := r.q.ListPromptVersions(ctx, promptID.UUID())
	if err != nil {
		return nil, repoerr.Handle(err, "VersionRepository", "")
	}

	result := make([]prompt.PromptVersion, 0, len(versions))
	for _, v := range versions {
		result = append(result, toVersion(v))
	}
	return result, nil
}

func (r *VersionRepository) FindLatest(ctx context.Context, promptID prompt.PromptID) (prompt.PromptVersion, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	v, err := r.q.GetLatestPromptVersion(ctx, promptID.UUID())
	if err != nil {
		return prompt.PromptVersion{}, repoerr.Handle(err, "VersionRepository", "PromptVersion")
	}
	return toVersion(v), nil
}

func (r *VersionRepository) FindProduction(ctx context.Context, promptID prompt.PromptID) (prompt.PromptVersion, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	v, err := r.q.GetProductionPromptVersion(ctx, promptID.UUID())
	if err != nil {
		return prompt.PromptVersion{}, repoerr.Handle(err, "VersionRepository", "PromptVersion")
	}
	return toVersion(v), nil
}
