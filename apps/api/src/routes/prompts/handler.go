package prompts

import (
	"api/src/domain/prompt"
	"api/src/services/diffservice"
	"api/src/services/lintservice"
)

type PromptHandler struct {
	promptRepo  prompt.PromptRepository
	versionRepo prompt.VersionRepository
	diffService *diffservice.DiffService
	lintService *lintservice.LintService
}

func NewPromptHandler(
	promptRepo prompt.PromptRepository,
	versionRepo prompt.VersionRepository,
	diffService *diffservice.DiffService,
	lintService *lintservice.LintService,
) *PromptHandler {
	return &PromptHandler{
		promptRepo:  promptRepo,
		versionRepo: versionRepo,
		diffService: diffService,
		lintService: lintService,
	}
}
