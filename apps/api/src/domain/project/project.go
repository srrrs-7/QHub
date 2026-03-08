// Package project defines the Project aggregate and its value objects.
//
// A Project belongs to an Organization and groups related prompts.
package project

import (
	"api/src/domain/valobj"

	"github.com/google/uuid"
)

// --- ProjectID ---

// ProjectID is the unique identifier for a project (UUID).
type ProjectID uuid.UUID

// NewProjectID parses a string UUID into a ProjectID.
func NewProjectID(id string) (ProjectID, error) {
	parsed, err := valobj.ParseUUID(id, "ProjectID")
	if err != nil {
		return ProjectID{}, err
	}
	return ProjectID(parsed), nil
}

// ProjectIDFromUUID converts a uuid.UUID directly (for DB results).
func ProjectIDFromUUID(id uuid.UUID) ProjectID { return ProjectID(id) }

// String returns the string representation.
func (p ProjectID) String() string { return uuid.UUID(p).String() }

// UUID returns the underlying uuid.UUID.
func (p ProjectID) UUID() uuid.UUID { return uuid.UUID(p) }

// --- ProjectName ---

// ProjectName is a validated name (2–100 characters, non-blank).
type ProjectName string

// NewProjectName validates and creates a ProjectName.
func NewProjectName(name string) (ProjectName, error) {
	if err := valobj.ValidateName(name, 2, 100, "ProjectName"); err != nil {
		return "", err
	}
	return ProjectName(name), nil
}

// String returns the name as a plain string.
func (p ProjectName) String() string { return string(p) }

// --- ProjectSlug ---

// ProjectSlug is a URL-safe identifier (2–50 chars, lowercase+hyphens).
type ProjectSlug string

// NewProjectSlug validates and creates a ProjectSlug.
func NewProjectSlug(slug string) (ProjectSlug, error) {
	if err := valobj.ValidateSlug(slug, 2, 50, "ProjectSlug"); err != nil {
		return "", err
	}
	return ProjectSlug(slug), nil
}

// String returns the slug as a plain string.
func (p ProjectSlug) String() string { return string(p) }

// --- ProjectDescription ---

// ProjectDescription is an optional description (max 500 characters).
type ProjectDescription string

// NewProjectDescription validates and creates a ProjectDescription.
func NewProjectDescription(desc string) (ProjectDescription, error) {
	if err := valobj.ValidateMaxLength(desc, 500, "ProjectDescription"); err != nil {
		return "", err
	}
	return ProjectDescription(desc), nil
}

// String returns the description as a plain string.
func (p ProjectDescription) String() string { return string(p) }

// --- Project (Aggregate) ---

// Project is the aggregate root representing a collection of prompts.
type Project struct {
	ID             ProjectID
	OrganizationID uuid.UUID
	Name           ProjectName
	Slug           ProjectSlug
	Description    ProjectDescription
}

// NewProject constructs a Project from validated value objects.
func NewProject(id ProjectID, orgID uuid.UUID, name ProjectName, slug ProjectSlug, desc ProjectDescription) Project {
	return Project{ID: id, OrganizationID: orgID, Name: name, Slug: slug, Description: desc}
}

// --- ProjectCmd ---

// ProjectCmd is a command object for creating or updating a project.
type ProjectCmd struct {
	OrganizationID uuid.UUID
	Name           ProjectName
	Slug           ProjectSlug
	Description    ProjectDescription
}

// NewProjectCmd constructs a ProjectCmd.
func NewProjectCmd(orgID uuid.UUID, name ProjectName, slug ProjectSlug, desc ProjectDescription) ProjectCmd {
	return ProjectCmd{OrganizationID: orgID, Name: name, Slug: slug, Description: desc}
}
