package prompts

import (
	"api/src/domain/prompt"
)

type PromptHandler struct {
	promptRepo  prompt.PromptRepository
	versionRepo prompt.VersionRepository
}

func NewPromptHandler(promptRepo prompt.PromptRepository, versionRepo prompt.VersionRepository) *PromptHandler {
	return &PromptHandler{promptRepo: promptRepo, versionRepo: versionRepo}
}
