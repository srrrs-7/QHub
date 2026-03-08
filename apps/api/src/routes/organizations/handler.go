package organizations

import (
	"api/src/domain/organization"
)

type OrganizationHandler struct {
	repo organization.OrganizationRepository
}

func NewOrganizationHandler(repo organization.OrganizationRepository) *OrganizationHandler {
	return &OrganizationHandler{repo: repo}
}
