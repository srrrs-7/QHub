package organization

import "context"

type OrganizationRepository interface {
	FindByID(ctx context.Context, id OrganizationID) (Organization, error)
	FindBySlug(ctx context.Context, slug OrganizationSlug) (Organization, error)
	FindAllByUserID(ctx context.Context, userID UserID) ([]Organization, error)
	Create(ctx context.Context, cmd OrganizationCmd) (Organization, error)
	Update(ctx context.Context, id OrganizationID, cmd OrganizationCmd) (Organization, error)
}

type MemberRepository interface {
	FindByOrgAndUser(ctx context.Context, orgID OrganizationID, userID UserID) (OrganizationMember, error)
	FindAllByOrg(ctx context.Context, orgID OrganizationID) ([]OrganizationMember, error)
	Add(ctx context.Context, orgID OrganizationID, userID UserID, role MemberRole) (OrganizationMember, error)
	UpdateRole(ctx context.Context, orgID OrganizationID, userID UserID, role MemberRole) (OrganizationMember, error)
	Remove(ctx context.Context, orgID OrganizationID, userID UserID) error
}
