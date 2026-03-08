package organization_repository

import (
	"api/src/domain/organization"
	"time"
	"utils/db/db"
)

const dbTimeout = 5 * time.Second

type OrganizationRepository struct {
	q db.Querier
}

func NewOrganizationRepository(q db.Querier) *OrganizationRepository {
	return &OrganizationRepository{q: q}
}

var _ organization.OrganizationRepository = (*OrganizationRepository)(nil)
