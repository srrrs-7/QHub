package prompts

import (
	"api/src/domain/prompt"
	"api/src/services/diffservice"
	"api/src/services/embeddingservice"
	"api/src/services/lintservice"
)

type PromptHandler struct {
	promptRepo   prompt.PromptRepository
	versionRepo  prompt.VersionRepository
	diffService  *diffservice.DiffService
	lintService  *lintservice.LintService
	embeddingSvc *embeddingservice.EmbeddingService
}

func NewPromptHandler(
	promptRepo prompt.PromptRepository,
	versionRepo prompt.VersionRepository,
	diffService *diffservice.DiffService,
	lintService *lintservice.LintService,
	embeddingSvc *embeddingservice.EmbeddingService,
) *PromptHandler {
	return &PromptHandler{
		promptRepo:   promptRepo,
		versionRepo:  versionRepo,
		diffService:  diffService,
		lintService:  lintService,
		embeddingSvc: embeddingSvc,
	}
}
