// Package actionservice detects and executes actionable suggestions from
// consulting chat responses. When the AI suggests new prompt text during a
// consulting session, this service can extract those suggestions and
// materialise them as new prompt versions.
package actionservice

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"api/src/domain/apperror"
	"api/src/domain/prompt"
	"api/src/services/intentservice"

	"github.com/google/uuid"
)

// ActionType enumerates the kinds of actions that can be extracted from an
// AI consulting response.
type ActionType string

const (
	// ActionCreateVersion indicates the AI suggested new prompt content that
	// can be saved as a new version.
	ActionCreateVersion ActionType = "create_version"
)

// Action represents an actionable suggestion extracted from a consulting
// chat response.
type Action struct {
	Type             ActionType      `json:"type"`
	PromptID         uuid.UUID       `json:"prompt_id"`
	SuggestedContent json.RawMessage `json:"suggested_content"`
	Description      string          `json:"description"`
}

// ExecuteResult holds the outcome of executing an action.
type ExecuteResult struct {
	Action        Action              `json:"action"`
	VersionID     prompt.PromptVersionID `json:"version_id"`
	VersionNumber int                 `json:"version_number"`
}

// ActionService detects actionable suggestions in AI responses and executes
// them against the prompt/version repositories.
type ActionService struct {
	promptRepo  prompt.PromptRepository
	versionRepo prompt.VersionRepository
}

// NewActionService creates an ActionService with the given repositories.
func NewActionService(promptRepo prompt.PromptRepository, versionRepo prompt.VersionRepository) *ActionService {
	return &ActionService{promptRepo: promptRepo, versionRepo: versionRepo}
}

// codeBlockMarkers used to detect fenced code blocks in AI responses.
var codeBlockMarkers = []string{"```"}

// ExtractActions parses the AI response text to detect actionable suggestions.
// Currently it looks for fenced code blocks when the intent is "create" or
// "improve", treating each block as a suggested new prompt version.
func ExtractActions(intent intentservice.Intent, promptID uuid.UUID, responseText string) []Action {
	if responseText == "" {
		return nil
	}

	// Only extract actions for intents that imply content creation.
	if intent.Type != intentservice.IntentCreate && intent.Type != intentservice.IntentImprove {
		return nil
	}

	blocks := extractCodeBlocks(responseText)
	if len(blocks) == 0 {
		return nil
	}

	actions := make([]Action, 0, len(blocks))
	for i, block := range blocks {
		content, err := json.Marshal(map[string]string{"text": block})
		if err != nil {
			continue
		}

		desc := "Suggested prompt from consulting chat"
		if len(blocks) > 1 {
			desc = fmt.Sprintf("Suggested prompt (variant %d) from consulting chat", i+1)
		}

		actions = append(actions, Action{
			Type:             ActionCreateVersion,
			PromptID:         promptID,
			SuggestedContent: json.RawMessage(content),
			Description:      desc,
		})
	}

	return actions
}

// extractCodeBlocks finds fenced code blocks (``` ... ```) in text and
// returns their contents. Only non-empty blocks are returned.
func extractCodeBlocks(text string) []string {
	var blocks []string
	remaining := text

	for {
		startIdx := strings.Index(remaining, codeBlockMarkers[0])
		if startIdx == -1 {
			break
		}

		// Skip past the opening marker and any language tag on the same line.
		afterOpen := remaining[startIdx+len(codeBlockMarkers[0]):]
		newlineIdx := strings.Index(afterOpen, "\n")
		if newlineIdx == -1 {
			break
		}
		contentStart := afterOpen[newlineIdx+1:]

		endIdx := strings.Index(contentStart, codeBlockMarkers[0])
		if endIdx == -1 {
			break
		}

		block := strings.TrimSpace(contentStart[:endIdx])
		if block != "" {
			blocks = append(blocks, block)
		}

		// Move past the closing marker.
		remaining = contentStart[endIdx+len(codeBlockMarkers[0]):]
	}

	return blocks
}

// ExecuteAction executes a single action, creating a new prompt version.
// It returns an error if the prompt is not found or the version cannot be
// created.
func (s *ActionService) ExecuteAction(ctx context.Context, action Action, authorID uuid.UUID) (ExecuteResult, error) {
	if action.Type != ActionCreateVersion {
		return ExecuteResult{}, apperror.NewValidationError(
			fmt.Errorf("unsupported action type: %s", action.Type),
			"Action",
		)
	}

	if action.PromptID == uuid.Nil {
		return ExecuteResult{}, apperror.NewValidationError(
			fmt.Errorf("prompt ID is required"),
			"Action",
		)
	}

	if len(action.SuggestedContent) == 0 {
		return ExecuteResult{}, apperror.NewValidationError(
			fmt.Errorf("suggested content is required"),
			"Action",
		)
	}

	pid := prompt.PromptIDFromUUID(action.PromptID)

	p, err := s.promptRepo.FindByID(ctx, pid)
	if err != nil {
		return ExecuteResult{}, err
	}

	changeDesc, err := prompt.NewChangeDescription(action.Description)
	if err != nil {
		return ExecuteResult{}, err
	}

	nextVersion := p.LatestVersion + 1

	v, err := s.versionRepo.Create(ctx, prompt.VersionCmd{
		PromptID:          pid,
		Content:           action.SuggestedContent,
		Variables:         json.RawMessage(`{}`),
		ChangeDescription: changeDesc,
		AuthorID:          authorID,
	}, nextVersion)
	if err != nil {
		return ExecuteResult{}, err
	}

	if _, err = s.promptRepo.UpdateLatestVersion(ctx, p.ID, nextVersion); err != nil {
		return ExecuteResult{}, err
	}

	return ExecuteResult{
		Action:        action,
		VersionID:     v.ID,
		VersionNumber: nextVersion,
	}, nil
}
