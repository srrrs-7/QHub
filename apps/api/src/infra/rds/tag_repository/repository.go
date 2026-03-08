package tag_repository

import (
	"api/src/domain/tag"
	"api/src/infra/rds/repoerr"
	"context"
	"time"
	"utils/db/db"

	"github.com/google/uuid"
)

const dbTimeout = 5 * time.Second

type TagRepository struct {
	q db.Querier
}

func NewTagRepository(q db.Querier) *TagRepository {
	return &TagRepository{q: q}
}

var _ tag.TagRepository = (*TagRepository)(nil)

func toTag(t db.Tag) tag.Tag {
	return tag.Tag{
		ID:        t.ID,
		OrgID:     t.OrganizationID,
		Name:      t.Name,
		Color:     t.Color,
		CreatedAt: t.CreatedAt,
	}
}

func (r *TagRepository) FindByID(ctx context.Context, id uuid.UUID) (tag.Tag, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	t, err := r.q.GetTag(ctx, id)
	if err != nil {
		return tag.Tag{}, repoerr.Handle(err, "TagRepository", "Tag")
	}
	return toTag(t), nil
}

func (r *TagRepository) FindAllByOrg(ctx context.Context, orgID uuid.UUID) ([]tag.Tag, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	tags, err := r.q.ListTagsByOrg(ctx, orgID)
	if err != nil {
		return nil, repoerr.Handle(err, "TagRepository", "")
	}

	result := make([]tag.Tag, 0, len(tags))
	for _, t := range tags {
		result = append(result, toTag(t))
	}
	return result, nil
}

func (r *TagRepository) Create(ctx context.Context, t tag.Tag) (tag.Tag, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	created, err := r.q.CreateTag(ctx, db.CreateTagParams{
		OrganizationID: t.OrgID,
		Name:           t.Name,
		Color:          t.Color,
	})
	if err != nil {
		return tag.Tag{}, repoerr.Handle(err, "TagRepository", "")
	}
	return toTag(created), nil
}

func (r *TagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := r.q.DeleteTag(ctx, id)
	if err != nil {
		return repoerr.Handle(err, "TagRepository", "")
	}
	return nil
}

func (r *TagRepository) AddToPrompt(ctx context.Context, promptID, tagID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := r.q.AddPromptTag(ctx, db.AddPromptTagParams{
		PromptID: promptID,
		TagID:    tagID,
	})
	if err != nil {
		return repoerr.Handle(err, "TagRepository", "")
	}
	return nil
}

func (r *TagRepository) RemoveFromPrompt(ctx context.Context, promptID, tagID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := r.q.RemovePromptTag(ctx, db.RemovePromptTagParams{
		PromptID: promptID,
		TagID:    tagID,
	})
	if err != nil {
		return repoerr.Handle(err, "TagRepository", "")
	}
	return nil
}

func (r *TagRepository) FindByPrompt(ctx context.Context, promptID uuid.UUID) ([]tag.Tag, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	tags, err := r.q.ListTagsByPrompt(ctx, promptID)
	if err != nil {
		return nil, repoerr.Handle(err, "TagRepository", "")
	}

	result := make([]tag.Tag, 0, len(tags))
	for _, t := range tags {
		result = append(result, toTag(t))
	}
	return result, nil
}
