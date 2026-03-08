package prompts

import (
	"api/src/domain/prompt"
	"encoding/json"
)

type promptResponse struct {
	ID                string `json:"id"`
	ProjectID         string `json:"project_id"`
	Name              string `json:"name"`
	Slug              string `json:"slug"`
	PromptType        string `json:"prompt_type"`
	Description       string `json:"description"`
	LatestVersion     int    `json:"latest_version"`
	ProductionVersion *int   `json:"production_version"`
}

func toPromptResponse(p prompt.Prompt) promptResponse {
	return promptResponse{
		ID:                p.ID.String(),
		ProjectID:         p.ProjectID.String(),
		Name:              p.Name.String(),
		Slug:              p.Slug.String(),
		PromptType:        p.PromptType.String(),
		Description:       p.Description.String(),
		LatestVersion:     p.LatestVersion,
		ProductionVersion: p.ProductionVersion,
	}
}

type versionResponse struct {
	ID                string          `json:"id"`
	PromptID          string          `json:"prompt_id"`
	VersionNumber     int             `json:"version_number"`
	Status            string          `json:"status"`
	Content           json.RawMessage `json:"content"`
	Variables         json.RawMessage `json:"variables"`
	ChangeDescription string          `json:"change_description"`
	AuthorID          string          `json:"author_id"`
}

func toVersionResponse(v prompt.PromptVersion) versionResponse {
	return versionResponse{
		ID:                v.ID.String(),
		PromptID:          v.PromptID.String(),
		VersionNumber:     v.VersionNumber,
		Status:            v.Status.String(),
		Content:           v.Content,
		Variables:         v.Variables,
		ChangeDescription: v.ChangeDescription.String(),
		AuthorID:          v.AuthorID.String(),
	}
}
