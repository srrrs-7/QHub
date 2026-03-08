package prompt_repository

import (
	"api/src/domain/prompt"
	"time"
	"utils/db/db"
)

const dbTimeout = 5 * time.Second

type PromptRepository struct {
	q db.Querier
}

func NewPromptRepository(q db.Querier) *PromptRepository {
	return &PromptRepository{q: q}
}

var _ prompt.PromptRepository = (*PromptRepository)(nil)

type VersionRepository struct {
	q db.Querier
}

func NewVersionRepository(q db.Querier) *VersionRepository {
	return &VersionRepository{q: q}
}

var _ prompt.VersionRepository = (*VersionRepository)(nil)
