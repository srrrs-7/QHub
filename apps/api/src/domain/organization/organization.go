package organization

import (
	"fmt"
	"regexp"
	"strings"

	"api/src/domain/apperror"

	"github.com/google/uuid"
)

// --- OrganizationID ---

type OrganizationID uuid.UUID

func NewOrganizationID(id string) (OrganizationID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return OrganizationID{}, apperror.NewValidationError(fmt.Errorf("invalid organization ID: %w", err), "OrganizationID")
	}
	return OrganizationID(parsed), nil
}

func OrganizationIDFromUUID(id uuid.UUID) OrganizationID {
	return OrganizationID(id)
}

func (o OrganizationID) String() string {
	return uuid.UUID(o).String()
}

func (o OrganizationID) UUID() uuid.UUID {
	return uuid.UUID(o)
}

// --- OrganizationName ---

type OrganizationName string

func NewOrganizationName(name string) (OrganizationName, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", apperror.NewValidationError(fmt.Errorf("organization name must not be empty"), "OrganizationName")
	}
	if len(name) < 2 {
		return "", apperror.NewValidationError(fmt.Errorf("organization name must be at least 2 characters"), "OrganizationName")
	}
	if len(name) > 100 {
		return "", apperror.NewValidationError(fmt.Errorf("organization name must be at most 100 characters"), "OrganizationName")
	}
	return OrganizationName(name), nil
}

func (o OrganizationName) String() string {
	return string(o)
}

// --- OrganizationSlug ---

type OrganizationSlug string

var slugRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)

func NewOrganizationSlug(slug string) (OrganizationSlug, error) {
	if len(slug) < 2 {
		return "", apperror.NewValidationError(fmt.Errorf("slug must be at least 2 characters"), "OrganizationSlug")
	}
	if len(slug) > 50 {
		return "", apperror.NewValidationError(fmt.Errorf("slug must be at most 50 characters"), "OrganizationSlug")
	}
	if !slugRegex.MatchString(slug) {
		return "", apperror.NewValidationError(fmt.Errorf("slug must contain only lowercase letters, numbers, and hyphens, and must not start or end with a hyphen"), "OrganizationSlug")
	}
	return OrganizationSlug(slug), nil
}

func (o OrganizationSlug) String() string {
	return string(o)
}

// --- Plan ---

type Plan string

const (
	PlanFree       Plan = "free"
	PlanPro        Plan = "pro"
	PlanTeam       Plan = "team"
	PlanEnterprise Plan = "enterprise"
)

func NewPlan(plan string) (Plan, error) {
	switch Plan(plan) {
	case PlanFree, PlanPro, PlanTeam, PlanEnterprise:
		return Plan(plan), nil
	default:
		return "", apperror.NewValidationError(fmt.Errorf("invalid plan: %s (must be 'free', 'pro', 'team', or 'enterprise')", plan), "Plan")
	}
}

func (p Plan) String() string {
	return string(p)
}

// --- MemberRole ---

type MemberRole string

const (
	RoleOwner  MemberRole = "owner"
	RoleAdmin  MemberRole = "admin"
	RoleMember MemberRole = "member"
	RoleViewer MemberRole = "viewer"
)

func NewMemberRole(role string) (MemberRole, error) {
	switch MemberRole(role) {
	case RoleOwner, RoleAdmin, RoleMember, RoleViewer:
		return MemberRole(role), nil
	default:
		return "", apperror.NewValidationError(fmt.Errorf("invalid role: %s (must be 'owner', 'admin', 'member', or 'viewer')", role), "MemberRole")
	}
}

func (r MemberRole) String() string {
	return string(r)
}

// --- Organization (Aggregate) ---

type Organization struct {
	ID   OrganizationID
	Name OrganizationName
	Slug OrganizationSlug
	Plan Plan
}

func NewOrganization(id OrganizationID, name OrganizationName, slug OrganizationSlug, plan Plan) Organization {
	return Organization{
		ID:   id,
		Name: name,
		Slug: slug,
		Plan: plan,
	}
}

// --- OrganizationCmd ---

type OrganizationCmd struct {
	Name OrganizationName
	Slug OrganizationSlug
	Plan Plan
}

func NewOrganizationCmd(name OrganizationName, slug OrganizationSlug, plan Plan) OrganizationCmd {
	return OrganizationCmd{
		Name: name,
		Slug: slug,
		Plan: plan,
	}
}

// --- OrganizationMember ---

type OrganizationMember struct {
	OrganizationID OrganizationID
	UserID         UserID
	Role           MemberRole
}

// --- UserID (referenced from user domain, minimal definition here) ---

type UserID uuid.UUID

func NewUserID(id string) (UserID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return UserID{}, apperror.NewValidationError(fmt.Errorf("invalid user ID: %w", err), "UserID")
	}
	return UserID(parsed), nil
}

func UserIDFromUUID(id uuid.UUID) UserID {
	return UserID(id)
}

func (u UserID) String() string {
	return uuid.UUID(u).String()
}

func (u UserID) UUID() uuid.UUID {
	return uuid.UUID(u)
}
