package organization_repository

import (
	"api/src/domain/organization"
	"api/src/infra/rds/repoerr"
	"context"

	"github.com/google/uuid"
)

func (r *OrganizationRepository) FindByID(ctx context.Context, id organization.OrganizationID) (organization.Organization, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	o, err := r.q.GetOrganization(ctx, uuid.UUID(id))
	if err != nil {
		return organization.Organization{}, repoerr.Handle(err, "OrganizationRepository", "Organization")
	}

	return organization.NewOrganization(
		organization.OrganizationIDFromUUID(o.ID),
		organization.OrganizationName(o.Name),
		organization.OrganizationSlug(o.Slug),
		organization.Plan(o.Plan),
	), nil
}

func (r *OrganizationRepository) FindBySlug(ctx context.Context, slug organization.OrganizationSlug) (organization.Organization, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	o, err := r.q.GetOrganizationBySlug(ctx, slug.String())
	if err != nil {
		return organization.Organization{}, repoerr.Handle(err, "OrganizationRepository", "Organization")
	}

	return organization.NewOrganization(
		organization.OrganizationIDFromUUID(o.ID),
		organization.OrganizationName(o.Name),
		organization.OrganizationSlug(o.Slug),
		organization.Plan(o.Plan),
	), nil
}

func (r *OrganizationRepository) FindAllByUserID(ctx context.Context, userID organization.UserID) ([]organization.Organization, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	orgs, err := r.q.ListOrganizationsByUser(ctx, userID.UUID())
	if err != nil {
		return nil, repoerr.Handle(err, "OrganizationRepository", "")
	}

	result := make([]organization.Organization, 0, len(orgs))
	for _, o := range orgs {
		result = append(result, organization.NewOrganization(
			organization.OrganizationIDFromUUID(o.ID),
			organization.OrganizationName(o.Name),
			organization.OrganizationSlug(o.Slug),
			organization.Plan(o.Plan),
		))
	}
	return result, nil
}
