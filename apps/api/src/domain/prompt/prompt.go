package prompt

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"api/src/domain/apperror"

	"github.com/google/uuid"
)

// --- PromptID ---

type PromptID uuid.UUID

func NewPromptID(id string) (PromptID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return PromptID{}, apperror.NewValidationError(fmt.Errorf("invalid prompt ID: %w", err), "PromptID")
	}
	return PromptID(parsed), nil
}

func PromptIDFromUUID(id uuid.UUID) PromptID { return PromptID(id) }
func (p PromptID) String() string             { return uuid.UUID(p).String() }
func (p PromptID) UUID() uuid.UUID            { return uuid.UUID(p) }

// --- PromptName ---

type PromptName string

func NewPromptName(name string) (PromptName, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", apperror.NewValidationError(fmt.Errorf("prompt name must not be empty"), "PromptName")
	}
	if len(name) < 2 {
		return "", apperror.NewValidationError(fmt.Errorf("prompt name must be at least 2 characters"), "PromptName")
	}
	if len(name) > 200 {
		return "", apperror.NewValidationError(fmt.Errorf("prompt name must be at most 200 characters"), "PromptName")
	}
	return PromptName(name), nil
}

func (p PromptName) String() string { return string(p) }

// --- PromptSlug ---

type PromptSlug string

var slugRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)

func NewPromptSlug(slug string) (PromptSlug, error) {
	if len(slug) < 2 {
		return "", apperror.NewValidationError(fmt.Errorf("slug must be at least 2 characters"), "PromptSlug")
	}
	if len(slug) > 80 {
		return "", apperror.NewValidationError(fmt.Errorf("slug must be at most 80 characters"), "PromptSlug")
	}
	if !slugRegex.MatchString(slug) {
		return "", apperror.NewValidationError(fmt.Errorf("slug must contain only lowercase letters, numbers, and hyphens"), "PromptSlug")
	}
	return PromptSlug(slug), nil
}

func (p PromptSlug) String() string { return string(p) }

// --- PromptType ---

type PromptType string

const (
	PromptTypeSystem   PromptType = "system"
	PromptTypeUser     PromptType = "user"
	PromptTypeCombined PromptType = "combined"
)

func NewPromptType(pt string) (PromptType, error) {
	switch PromptType(pt) {
	case PromptTypeSystem, PromptTypeUser, PromptTypeCombined:
		return PromptType(pt), nil
	default:
		return "", apperror.NewValidationError(fmt.Errorf("invalid prompt type: %s", pt), "PromptType")
	}
}

func (p PromptType) String() string { return string(p) }

// --- PromptDescription ---

type PromptDescription string

func NewPromptDescription(desc string) (PromptDescription, error) {
	if len(desc) > 1000 {
		return "", apperror.NewValidationError(fmt.Errorf("description must be at most 1000 characters"), "PromptDescription")
	}
	return PromptDescription(desc), nil
}

func (p PromptDescription) String() string { return string(p) }

// --- VersionStatus ---

type VersionStatus string

const (
	StatusDraft      VersionStatus = "draft"
	StatusReview     VersionStatus = "review"
	StatusProduction VersionStatus = "production"
	StatusArchived   VersionStatus = "archived"
)

func NewVersionStatus(status string) (VersionStatus, error) {
	switch VersionStatus(status) {
	case StatusDraft, StatusReview, StatusProduction, StatusArchived:
		return VersionStatus(status), nil
	default:
		return "", apperror.NewValidationError(fmt.Errorf("invalid version status: %s", status), "VersionStatus")
	}
}

func (v VersionStatus) String() string { return string(v) }

// ValidateStatusTransition checks if a status transition is allowed.
//
//	draft → review → production → archived
//	draft → archived (discard)
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

	for _, t := range targets {
		if t == to {
			return nil
		}
	}

	return apperror.NewValidationError(
		fmt.Errorf("invalid status transition: %s → %s", from, to),
		"VersionStatus",
	)
}

// --- ChangeDescription ---

type ChangeDescription string

func NewChangeDescription(desc string) (ChangeDescription, error) {
	if len(desc) > 500 {
		return "", apperror.NewValidationError(fmt.Errorf("change description must be at most 500 characters"), "ChangeDescription")
	}
	return ChangeDescription(desc), nil
}

func (c ChangeDescription) String() string { return string(c) }

// --- Prompt (Aggregate) ---

type Prompt struct {
	ID                PromptID
	ProjectID         uuid.UUID
	Name              PromptName
	Slug              PromptSlug
	PromptType        PromptType
	Description       PromptDescription
	LatestVersion     int
	ProductionVersion *int
}

func NewPrompt(id PromptID, projectID uuid.UUID, name PromptName, slug PromptSlug, promptType PromptType, desc PromptDescription, latestVersion int, productionVersion *int) Prompt {
	return Prompt{
		ID:                id,
		ProjectID:         projectID,
		Name:              name,
		Slug:              slug,
		PromptType:        promptType,
		Description:       desc,
		LatestVersion:     latestVersion,
		ProductionVersion: productionVersion,
	}
}

// --- PromptVersion ---

type PromptVersion struct {
	ID                PromptVersionID
	PromptID          PromptID
	VersionNumber     int
	Status            VersionStatus
	Content           json.RawMessage
	Variables         json.RawMessage
	ChangeDescription ChangeDescription
	AuthorID          uuid.UUID
}

type PromptVersionID uuid.UUID

func NewPromptVersionID(id string) (PromptVersionID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return PromptVersionID{}, apperror.NewValidationError(fmt.Errorf("invalid prompt version ID: %w", err), "PromptVersionID")
	}
	return PromptVersionID(parsed), nil
}

func PromptVersionIDFromUUID(id uuid.UUID) PromptVersionID { return PromptVersionID(id) }
func (p PromptVersionID) String() string                    { return uuid.UUID(p).String() }
func (p PromptVersionID) UUID() uuid.UUID                   { return uuid.UUID(p) }

// --- PromptCmd ---

type PromptCmd struct {
	ProjectID  uuid.UUID
	Name       PromptName
	Slug       PromptSlug
	PromptType PromptType
	Description PromptDescription
}

// --- VersionCmd ---

type VersionCmd struct {
	PromptID          PromptID
	Content           json.RawMessage
	Variables         json.RawMessage
	ChangeDescription ChangeDescription
	AuthorID          uuid.UUID
}
