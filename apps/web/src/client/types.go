package client

import "encoding/json"

// --- Organization ---

type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Plan string `json:"plan"`
}

// --- Project ---

type Project struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	Description    string `json:"description"`
}

// --- Prompt ---

type Prompt struct {
	ID                string `json:"id"`
	ProjectID         string `json:"project_id"`
	Name              string `json:"name"`
	Slug              string `json:"slug"`
	PromptType        string `json:"prompt_type"`
	Description       string `json:"description"`
	LatestVersion     int    `json:"latest_version"`
	ProductionVersion *int   `json:"production_version"`
}

func (p Prompt) HasProduction() bool {
	return p.ProductionVersion != nil
}

func (p Prompt) StatusLabel() string {
	if p.ProductionVersion != nil {
		return "production"
	}
	return "draft"
}

// --- PromptVersion ---

type PromptVersion struct {
	ID                string          `json:"id"`
	PromptID          string          `json:"prompt_id"`
	VersionNumber     int             `json:"version_number"`
	Status            string          `json:"status"`
	Content           json.RawMessage `json:"content"`
	Variables         json.RawMessage `json:"variables"`
	ChangeDescription string          `json:"change_description"`
	AuthorID          string          `json:"author_id"`
}

func (v PromptVersion) IsProduction() bool { return v.Status == "production" }
func (v PromptVersion) IsDraft() bool      { return v.Status == "draft" }
func (v PromptVersion) IsReview() bool     { return v.Status == "review" }
func (v PromptVersion) IsArchived() bool   { return v.Status == "archived" }

func (v PromptVersion) ContentString() string {
	if v.Content == nil {
		return ""
	}
	var s string
	if err := json.Unmarshal(v.Content, &s); err != nil {
		return string(v.Content)
	}
	return s
}

func (v PromptVersion) VariablesList() []string {
	if v.Variables == nil {
		return nil
	}
	var vars []string
	if err := json.Unmarshal(v.Variables, &vars); err != nil {
		return nil
	}
	return vars
}

// --- Legacy Task types (kept for backward compatibility) ---

type Task struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func (t Task) IsCompleted() bool {
	return t.Status == "completed"
}

type TasksResponse struct {
	Tasks []Task `json:"tasks"`
}

type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}
