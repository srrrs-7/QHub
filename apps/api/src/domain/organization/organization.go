// Package organization defines the Organization aggregate and its value objects.
//
// An Organization is the top-level tenant in the system. It has members
// (users with roles) and contains projects. Plans control feature access.
package organization

import (
	"fmt"

	"api/src/domain/apperror"
	"api/src/domain/valobj"

	"github.com/google/uuid"
)

// --- OrganizationID ---

// OrganizationID is the unique identifier for an organization (UUID).
type OrganizationID uuid.UUID

// NewOrganizationID parses a string UUID into an OrganizationID.
func NewOrganizationID(id string) (OrganizationID, error) {
	parsed, err := valobj.ParseUUID(id, "OrganizationID")
	if err != nil {
		return OrganizationID{}, err
	}
	return OrganizationID(parsed), nil
}

// OrganizationIDFromUUID converts a uuid.UUID directly (for DB results).
func OrganizationIDFromUUID(id uuid.UUID) OrganizationID { return OrganizationID(id) }

// String returns the string representation.
func (o OrganizationID) String() string { return uuid.UUID(o).String() }

// UUID returns the underlying uuid.UUID.
func (o OrganizationID) UUID() uuid.UUID { return uuid.UUID(o) }

// --- OrganizationName ---

// OrganizationName is a validated name (2–100 characters, non-blank).
type OrganizationName string

// NewOrganizationName validates and creates an OrganizationName.
func NewOrganizationName(name string) (OrganizationName, error) {
	if err := valobj.ValidateName(name, 2, 100, "OrganizationName"); err != nil {
		return "", err
	}
	return OrganizationName(name), nil
}

// String returns the name as a plain string.
func (o OrganizationName) String() string { return string(o) }

// --- OrganizationSlug ---

// OrganizationSlug is a URL-safe identifier (2–50 chars, lowercase+hyphens).
type OrganizationSlug string

// NewOrganizationSlug validates and creates an OrganizationSlug.
func NewOrganizationSlug(slug string) (OrganizationSlug, error) {
	if err := valobj.ValidateSlug(slug, 2, 50, "OrganizationSlug"); err != nil {
		return "", err
	}
	return OrganizationSlug(slug), nil
}

// String returns the slug as a plain string.
func (o OrganizationSlug) String() string { return string(o) }

// --- Plan ---

// Plan represents the subscription tier of an organization.
type Plan string

const (
	PlanFree       Plan = "free"
	PlanPro        Plan = "pro"
	PlanTeam       Plan = "team"
	PlanEnterprise Plan = "enterprise"
)

// NewPlan validates a plan string against known tiers.
func NewPlan(plan string) (Plan, error) {
	switch Plan(plan) {
	case PlanFree, PlanPro, PlanTeam, PlanEnterprise:
		return Plan(plan), nil
	default:
		return "", apperror.NewValidationError(fmt.Errorf("invalid plan: %s (must be 'free', 'pro', 'team', or 'enterprise')", plan), "Plan")
	}
}

// String returns the plan as a plain string.
func (p Plan) String() string { return string(p) }

// --- MemberRole ---

// MemberRole represents a user's role within an organization.
type MemberRole string

const (
	RoleOwner  MemberRole = "owner"
	RoleAdmin  MemberRole = "admin"
	RoleMember MemberRole = "member"
	RoleViewer MemberRole = "viewer"
)

// NewMemberRole validates a role string against known roles.
func NewMemberRole(role string) (MemberRole, error) {
	switch MemberRole(role) {
	case RoleOwner, RoleAdmin, RoleMember, RoleViewer:
		return MemberRole(role), nil
	default:
		return "", apperror.NewValidationError(fmt.Errorf("invalid role: %s (must be 'owner', 'admin', 'member', or 'viewer')", role), "MemberRole")
	}
}

// String returns the role as a plain string.
func (r MemberRole) String() string { return string(r) }

// --- Organization (Aggregate) ---

// Organization is the root aggregate representing a tenant.
type Organization struct {
	ID   OrganizationID
	Name OrganizationName
	Slug OrganizationSlug
	Plan Plan
}

// NewOrganization constructs an Organization from validated value objects.
func NewOrganization(id OrganizationID, name OrganizationName, slug OrganizationSlug, plan Plan) Organization {
	return Organization{ID: id, Name: name, Slug: slug, Plan: plan}
}

// --- OrganizationCmd ---

// OrganizationCmd is a command object for creating or updating an organization.
type OrganizationCmd struct {
	Name OrganizationName
	Slug OrganizationSlug
	Plan Plan
}

// NewOrganizationCmd constructs an OrganizationCmd.
func NewOrganizationCmd(name OrganizationName, slug OrganizationSlug, plan Plan) OrganizationCmd {
	return OrganizationCmd{Name: name, Slug: slug, Plan: plan}
}

// --- OrganizationMember ---

// OrganizationMember represents the membership of a user in an organization.
type OrganizationMember struct {
	OrganizationID OrganizationID
	UserID         UserID
	Role           MemberRole
}

// --- UserID ---

// UserID is the unique identifier for a user (UUID).
type UserID uuid.UUID

// NewUserID parses a string UUID into a UserID.
func NewUserID(id string) (UserID, error) {
	parsed, err := valobj.ParseUUID(id, "UserID")
	if err != nil {
		return UserID{}, err
	}
	return UserID(parsed), nil
}

// UserIDFromUUID converts a uuid.UUID directly (for DB results).
func UserIDFromUUID(id uuid.UUID) UserID { return UserID(id) }

// String returns the string representation.
func (u UserID) String() string { return uuid.UUID(u).String() }

// UUID returns the underlying uuid.UUID.
func (u UserID) UUID() uuid.UUID { return uuid.UUID(u) }
