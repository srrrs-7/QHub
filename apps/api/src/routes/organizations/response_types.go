package organizations

import "api/src/domain/organization"

type organizationResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Plan string `json:"plan"`
}

func toOrganizationResponse(o organization.Organization) organizationResponse {
	return organizationResponse{
		ID:   o.ID.String(),
		Name: o.Name.String(),
		Slug: o.Slug.String(),
		Plan: o.Plan.String(),
	}
}
