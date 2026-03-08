package project

import (
	"fmt"
	"regexp"
	"strings"

	"api/src/domain/apperror"

	"github.com/google/uuid"
)

// --- ProjectID ---

type ProjectID uuid.UUID

func NewProjectID(id string) (ProjectID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return ProjectID{}, apperror.NewValidationError(fmt.Errorf("invalid project ID: %w", err), "ProjectID")
	}
	return ProjectID(parsed), nil
}

func ProjectIDFromUUID(id uuid.UUID) ProjectID {
	return ProjectID(id)
}

func (p ProjectID) String() string { return uuid.UUID(p).String() }
func (p ProjectID) UUID() uuid.UUID { return uuid.UUID(p) }

// --- ProjectName ---

type ProjectName string

func NewProjectName(name string) (ProjectName, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", apperror.NewValidationError(fmt.Errorf("project name must not be empty"), "ProjectName")
	}
	if len(name) < 2 {
		return "", apperror.NewValidationError(fmt.Errorf("project name must be at least 2 characters"), "ProjectName")
	}
	if len(name) > 100 {
		return "", apperror.NewValidationError(fmt.Errorf("project name must be at most 100 characters"), "ProjectName")
	}
	return ProjectName(name), nil
}

func (p ProjectName) String() string { return string(p) }

// --- ProjectSlug ---

type ProjectSlug string

var slugRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)

func NewProjectSlug(slug string) (ProjectSlug, error) {
	if len(slug) < 2 {
		return "", apperror.NewValidationError(fmt.Errorf("slug must be at least 2 characters"), "ProjectSlug")
	}
	if len(slug) > 50 {
		return "", apperror.NewValidationError(fmt.Errorf("slug must be at most 50 characters"), "ProjectSlug")
	}
	if !slugRegex.MatchString(slug) {
		return "", apperror.NewValidationError(fmt.Errorf("slug must contain only lowercase letters, numbers, and hyphens"), "ProjectSlug")
	}
	return ProjectSlug(slug), nil
}

func (p ProjectSlug) String() string { return string(p) }

// --- ProjectDescription ---

type ProjectDescription string

func NewProjectDescription(desc string) (ProjectDescription, error) {
	if len(desc) > 500 {
		return "", apperror.NewValidationError(fmt.Errorf("description must be at most 500 characters"), "ProjectDescription")
	}
	return ProjectDescription(desc), nil
}

func (p ProjectDescription) String() string { return string(p) }

// --- Project (Aggregate) ---

type Project struct {
	ID             ProjectID
	OrganizationID uuid.UUID
	Name           ProjectName
	Slug           ProjectSlug
	Description    ProjectDescription
}

func NewProject(id ProjectID, orgID uuid.UUID, name ProjectName, slug ProjectSlug, desc ProjectDescription) Project {
	return Project{
		ID:             id,
		OrganizationID: orgID,
		Name:           name,
		Slug:           slug,
		Description:    desc,
	}
}

// --- ProjectCmd ---

type ProjectCmd struct {
	OrganizationID uuid.UUID
	Name           ProjectName
	Slug           ProjectSlug
	Description    ProjectDescription
}

func NewProjectCmd(orgID uuid.UUID, name ProjectName, slug ProjectSlug, desc ProjectDescription) ProjectCmd {
	return ProjectCmd{
		OrganizationID: orgID,
		Name:           name,
		Slug:           slug,
		Description:    desc,
	}
}
