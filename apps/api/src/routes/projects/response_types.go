package projects

import "api/src/domain/project"

type projectResponse struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	Description    string `json:"description"`
}

func toProjectResponse(p project.Project) projectResponse {
	return projectResponse{
		ID:             p.ID.String(),
		OrganizationID: p.OrganizationID.String(),
		Name:           p.Name.String(),
		Slug:           p.Slug.String(),
		Description:    p.Description.String(),
	}
}
