// Package prompt defines the Prompt / PromptVersion aggregates, their
// value objects, and the status lifecycle for version management.
//
// A Prompt belongs to a Project. Each Prompt has multiple PromptVersions
// that move through the status lifecycle: draft → review → production → archived.
package prompt

import (
	"encoding/json"
	"fmt"
	"slices"

	"api/src/domain/apperror"
	"api/src/domain/valobj"

	"github.com/google/uuid"
)

// --- PromptID ---

// PromptID is the unique identifier for a prompt (UUID).
type PromptID uuid.UUID

// NewPromptID parses a string UUID into a PromptID.
func NewPromptID(id string) (PromptID, error) {
	parsed, err := valobj.ParseUUID(id, "PromptID")
	if err != nil {
		return PromptID{}, err
	}
	return PromptID(parsed), nil
}

// PromptIDFromUUID converts a uuid.UUID directly (for DB results).
func PromptIDFromUUID(id uuid.UUID) PromptID { return PromptID(id) }

// String returns the string representation.
func (p PromptID) String() string { return uuid.UUID(p).String() }

// UUID returns the underlying uuid.UUID.
func (p PromptID) UUID() uuid.UUID { return uuid.UUID(p) }

// --- PromptName ---

// PromptName is a validated name (2–200 characters, non-blank).
type PromptName string

// NewPromptName validates and creates a PromptName.
func NewPromptName(name string) (PromptName, error) {
	if err := valobj.ValidateName(name, 2, 200, "PromptName"); err != nil {
		return "", err
	}
	return PromptName(name), nil
}

// String returns the name as a plain string.
func (p PromptName) String() string { return string(p) }

// --- PromptSlug ---

// PromptSlug is a URL-safe identifier (2–80 chars, lowercase+hyphens).
type PromptSlug string

// NewPromptSlug validates and creates a PromptSlug.
func NewPromptSlug(slug string) (PromptSlug, error) {
	if err := valobj.ValidateSlug(slug, 2, 80, "PromptSlug"); err != nil {
		return "", err
	}
	return PromptSlug(slug), nil
}

// String returns the slug as a plain string.
func (p PromptSlug) String() string { return string(p) }

// --- PromptType ---

// PromptType categorises how the prompt is used: system, user, or combined.
type PromptType string

const (
	PromptTypeSystem   PromptType = "system"
	PromptTypeUser     PromptType = "user"
	PromptTypeCombined PromptType = "combined"
)

// NewPromptType validates a prompt-type string.
func NewPromptType(pt string) (PromptType, error) {
	switch PromptType(pt) {
	case PromptTypeSystem, PromptTypeUser, PromptTypeCombined:
		return PromptType(pt), nil
	default:
		return "", apperror.NewValidationError(fmt.Errorf("invalid prompt type: %s", pt), "PromptType")
	}
}

// String returns the prompt type as a plain string.
func (p PromptType) String() string { return string(p) }

// --- PromptDescription ---

// PromptDescription is an optional description (max 1000 characters).
type PromptDescription string

// NewPromptDescription validates and creates a PromptDescription.
func NewPromptDescription(desc string) (PromptDescription, error) {
	if err := valobj.ValidateMaxLength(desc, 1000, "PromptDescription"); err != nil {
		return "", err
	}
	return PromptDescription(desc), nil
}

// String returns the description as a plain string.
func (p PromptDescription) String() string { return string(p) }

// --- VersionStatus ---

// VersionStatus represents the lifecycle stage of a prompt version.
type VersionStatus string

const (
	StatusDraft      VersionStatus = "draft"
	StatusReview     VersionStatus = "review"
	StatusProduction VersionStatus = "production"
	StatusArchived   VersionStatus = "archived"
)

// NewVersionStatus validates a version-status string.
func NewVersionStatus(status string) (VersionStatus, error) {
	switch VersionStatus(status) {
	case StatusDraft, StatusReview, StatusProduction, StatusArchived:
		return VersionStatus(status), nil
	default:
		return "", apperror.NewValidationError(fmt.Errorf("invalid version status: %s", status), "VersionStatus")
	}
}

// String returns the status as a plain string.
func (v VersionStatus) String() string { return string(v) }

// ValidateStatusTransition checks if moving from one status to another is
// allowed by the lifecycle rules:
//
//	draft → review → production → archived
//	draft → archived (discard without review)
func ValidateStatusTransition(from, to VersionStatus) error {
	allowed := map[VersionStatus][]VersionStatus{
		StatusDraft:      {StatusReview, StatusArchived},
		StatusReview:     {StatusProduction},
		StatusProduction: {StatusArchived},
		StatusArchived:   {},
	}

	targets, ok := allowed[from]
	if !ok {
		return apperror.NewValidationError(fmt.Errorf("unknown status: %s", from), "VersionStatus")
	}
	if slices.Contains(targets, to) {
		return nil
	}
	return apperror.NewValidationError(fmt.Errorf("invalid status transition: %s → %s", from, to), "VersionStatus")
}

// --- ChangeDescription ---

// ChangeDescription summarises what changed in a version (max 500 characters).
type ChangeDescription string

// NewChangeDescription validates and creates a ChangeDescription.
func NewChangeDescription(desc string) (ChangeDescription, error) {
	if err := valobj.ValidateMaxLength(desc, 500, "ChangeDescription"); err != nil {
		return "", err
	}
	return ChangeDescription(desc), nil
}

// String returns the description as a plain string.
func (c ChangeDescription) String() string { return string(c) }

// --- Prompt (Aggregate) ---

// Prompt is the aggregate root representing a managed prompt template.
type Prompt struct {
	ID                PromptID
	ProjectID         uuid.UUID
	Name              PromptName
	Slug              PromptSlug
	PromptType        PromptType
	Description       PromptDescription
	LatestVersion     int
	ProductionVersion *int // nil when no version is in production
}

// NewPrompt constructs a Prompt from validated value objects.
func NewPrompt(id PromptID, projectID uuid.UUID, name PromptName, slug PromptSlug, promptType PromptType, desc PromptDescription, latestVersion int, productionVersion *int) Prompt {
	return Prompt{
		ID: id, ProjectID: projectID, Name: name, Slug: slug,
		PromptType: promptType, Description: desc,
		LatestVersion: latestVersion, ProductionVersion: productionVersion,
	}
}

// --- PromptVersion ---

// PromptVersion represents a single revision of a prompt's content.
type PromptVersion struct {
	ID                PromptVersionID
	PromptID          PromptID
	VersionNumber     int
	Status            VersionStatus
	Content           json.RawMessage // JSONB content
	Variables         json.RawMessage // JSONB variable definitions
	ChangeDescription ChangeDescription
	AuthorID          uuid.UUID
}

// --- PromptVersionID ---

// PromptVersionID is the unique identifier for a prompt version (UUID).
type PromptVersionID uuid.UUID

// NewPromptVersionID parses a string UUID into a PromptVersionID.
func NewPromptVersionID(id string) (PromptVersionID, error) {
	parsed, err := valobj.ParseUUID(id, "PromptVersionID")
	if err != nil {
		return PromptVersionID{}, err
	}
	return PromptVersionID(parsed), nil
}

// PromptVersionIDFromUUID converts a uuid.UUID directly (for DB results).
func PromptVersionIDFromUUID(id uuid.UUID) PromptVersionID { return PromptVersionID(id) }

// String returns the string representation.
func (p PromptVersionID) String() string { return uuid.UUID(p).String() }

// UUID returns the underlying uuid.UUID.
func (p PromptVersionID) UUID() uuid.UUID { return uuid.UUID(p) }

// --- PromptCmd ---

// PromptCmd is a command object for creating or updating a prompt.
type PromptCmd struct {
	ProjectID   uuid.UUID
	Name        PromptName
	Slug        PromptSlug
	PromptType  PromptType
	Description PromptDescription
}

// --- VersionCmd ---

// VersionCmd is a command object for creating a new prompt version.
type VersionCmd struct {
	PromptID          PromptID
	Content           json.RawMessage
	Variables         json.RawMessage
	ChangeDescription ChangeDescription
	AuthorID          uuid.UUID
}
