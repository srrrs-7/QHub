package prompt_repository

import (
	"api/src/domain/prompt"
	"api/src/infra/rds/repoerr"
	"context"
	"database/sql"
	"encoding/json"
	"utils/db/db"

	"github.com/sqlc-dev/pqtype"
)

// --- PromptRepository ---

func (r *PromptRepository) Create(ctx context.Context, cmd prompt.PromptCmd) (prompt.Prompt, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	desc := string(cmd.Description)
	p, err := r.q.CreatePrompt(ctx, db.CreatePromptParams{
		ProjectID:   cmd.ProjectID,
		Name:        cmd.Name.String(),
		Slug:        cmd.Slug.String(),
		PromptType:  cmd.PromptType.String(),
		Description: sql.NullString{String: desc, Valid: desc != ""},
	})
	if err != nil {
		return prompt.Prompt{}, repoerr.Handle(err, "PromptRepository", "")
	}
	return toPrompt(p), nil
}

func (r *PromptRepository) UpdateLatestVersion(ctx context.Context, id prompt.PromptID, version int) (prompt.Prompt, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	p, err := r.q.UpdatePromptLatestVersion(ctx, db.UpdatePromptLatestVersionParams{
		ID:            id.UUID(),
		LatestVersion: int32(version),
	})
	if err != nil {
		return prompt.Prompt{}, repoerr.Handle(err, "PromptRepository", "Prompt")
	}
	return toPrompt(p), nil
}

func (r *PromptRepository) UpdateProductionVersion(ctx context.Context, id prompt.PromptID, version *int) (prompt.Prompt, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	var v sql.NullInt32
	if version != nil {
		v = sql.NullInt32{Int32: int32(*version), Valid: true}
	}

	p, err := r.q.UpdatePromptProductionVersion(ctx, db.UpdatePromptProductionVersionParams{
		ID:                id.UUID(),
		ProductionVersion: v,
	})
	if err != nil {
		return prompt.Prompt{}, repoerr.Handle(err, "PromptRepository", "Prompt")
	}
	return toPrompt(p), nil
}

func (r *PromptRepository) Update(ctx context.Context, id prompt.PromptID, cmd prompt.PromptCmd) (prompt.Prompt, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	name := cmd.Name.String()
	slug := cmd.Slug.String()
	desc := string(cmd.Description)

	p, err := r.q.UpdatePrompt(ctx, db.UpdatePromptParams{
		ID:          id.UUID(),
		Name:        sql.NullString{String: name, Valid: name != ""},
		Slug:        sql.NullString{String: slug, Valid: slug != ""},
		Description: sql.NullString{String: desc, Valid: desc != ""},
	})
	if err != nil {
		return prompt.Prompt{}, repoerr.Handle(err, "PromptRepository", "Prompt")
	}
	return toPrompt(p), nil
}

// --- VersionRepository ---

func (r *VersionRepository) Create(ctx context.Context, cmd prompt.VersionCmd, versionNumber int) (prompt.PromptVersion, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	changeDesc := cmd.ChangeDescription.String()

	v, err := r.q.CreatePromptVersion(ctx, db.CreatePromptVersionParams{
		PromptID:          cmd.PromptID.UUID(),
		VersionNumber:     int32(versionNumber),
		Status:            string(prompt.StatusDraft),
		Content:           cmd.Content,
		Variables:         pqtype.NullRawMessage{RawMessage: cmd.Variables, Valid: cmd.Variables != nil},
		ChangeDescription: sql.NullString{String: changeDesc, Valid: changeDesc != ""},
		AuthorID:          cmd.AuthorID,
	})
	if err != nil {
		return prompt.PromptVersion{}, repoerr.Handle(err, "VersionRepository", "")
	}
	return toVersion(v), nil
}

func (r *VersionRepository) UpdateStatus(ctx context.Context, id prompt.PromptVersionID, status prompt.VersionStatus) (prompt.PromptVersion, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	v, err := r.q.UpdatePromptVersionStatus(ctx, db.UpdatePromptVersionStatusParams{
		ID:     id.UUID(),
		Status: status.String(),
	})
	if err != nil {
		return prompt.PromptVersion{}, repoerr.Handle(err, "VersionRepository", "PromptVersion")
	}
	return toVersion(v), nil
}

func (r *VersionRepository) ArchiveProduction(ctx context.Context, promptID prompt.PromptID) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := r.q.ArchiveProductionVersion(ctx, promptID.UUID())
	if err != nil {
		return repoerr.Handle(err, "VersionRepository", "")
	}
	return nil
}

func (r *VersionRepository) UpdateLintResult(ctx context.Context, id prompt.PromptVersionID, result json.RawMessage) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	return r.q.UpdatePromptVersionLintResult(ctx, db.UpdatePromptVersionLintResultParams{
		ID: id.UUID(),
		LintResult: pqtype.NullRawMessage{
			RawMessage: result,
			Valid:      true,
		},
	})
}

func (r *VersionRepository) UpdateSemanticDiff(ctx context.Context, id prompt.PromptVersionID, diff json.RawMessage) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	return r.q.UpdatePromptVersionSemanticDiff(ctx, db.UpdatePromptVersionSemanticDiffParams{
		ID: id.UUID(),
		SemanticDiff: pqtype.NullRawMessage{
			RawMessage: diff,
			Valid:      true,
		},
	})
}
