package organization_repository

import (
	"api/src/domain/organization"
	"api/src/infra/rds/repoerr"
	"context"
	"database/sql"
	"utils/db/db"

	"github.com/google/uuid"
)

func (r *OrganizationRepository) Create(ctx context.Context, cmd organization.OrganizationCmd) (organization.Organization, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	o, err := r.q.CreateOrganization(ctx, db.CreateOrganizationParams{
		Name: cmd.Name.String(),
		Slug: cmd.Slug.String(),
		Plan: cmd.Plan.String(),
	})
	if err != nil {
		return organization.Organization{}, repoerr.Handle(err, "OrganizationRepository", "")
	}

	return organization.NewOrganization(
		organization.OrganizationIDFromUUID(o.ID),
		organization.OrganizationName(o.Name),
		organization.OrganizationSlug(o.Slug),
		organization.Plan(o.Plan),
	), nil
}

func (r *OrganizationRepository) Update(ctx context.Context, id organization.OrganizationID, cmd organization.OrganizationCmd) (organization.Organization, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	name := cmd.Name.String()
	slug := cmd.Slug.String()
	plan := cmd.Plan.String()

	o, err := r.q.UpdateOrganization(ctx, db.UpdateOrganizationParams{
		ID:   uuid.UUID(id),
		Name: sql.NullString{String: name, Valid: name != ""},
		Slug: sql.NullString{String: slug, Valid: slug != ""},
		Plan: sql.NullString{String: plan, Valid: plan != ""},
	})
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
