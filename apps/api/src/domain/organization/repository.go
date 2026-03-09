package organization

import "context"

// OrganizationRepository defines persistence operations for organizations.
// Implementations live in the infra layer (dependency inversion).
type OrganizationRepository interface {
	FindAll(ctx context.Context) ([]Organization, error)
	FindByID(ctx context.Context, id OrganizationID) (Organization, error)
	FindBySlug(ctx context.Context, slug OrganizationSlug) (Organization, error)
	FindAllByUserID(ctx context.Context, userID UserID) ([]Organization, error)
	Create(ctx context.Context, cmd OrganizationCmd) (Organization, error)
	Update(ctx context.Context, id OrganizationID, cmd OrganizationCmd) (Organization, error)
}

// MemberRepository defines persistence operations for organization members.
type MemberRepository interface {
	FindByOrgAndUser(ctx context.Context, orgID OrganizationID, userID UserID) (OrganizationMember, error)
	FindAllByOrg(ctx context.Context, orgID OrganizationID) ([]OrganizationMember, error)
	Add(ctx context.Context, orgID OrganizationID, userID UserID, role MemberRole) (OrganizationMember, error)
	UpdateRole(ctx context.Context, orgID OrganizationID, userID UserID, role MemberRole) (OrganizationMember, error)
	Remove(ctx context.Context, orgID OrganizationID, userID UserID) error
}
