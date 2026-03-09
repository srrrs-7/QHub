// Package industry defines domain value objects and entities for industry configurations.
//
// Industry configurations store domain-specific knowledge bases and compliance
// rules for vertical industries (e.g. healthcare, finance, legal). They are
// used by the consulting feature to provide industry-aware responses.
package industry

import (
	"fmt"
	"regexp"

	"api/src/domain/apperror"
	"api/src/domain/valobj"

	"github.com/google/uuid"
)

// IndustryID is the unique identifier for an industry config (UUID).
type IndustryID uuid.UUID

// NewIndustryID parses a string UUID into an IndustryID.
func NewIndustryID(id string) (IndustryID, error) {
	parsed, err := valobj.ParseUUID(id, "IndustryID")
	if err != nil {
		return IndustryID{}, err
	}
	return IndustryID(parsed), nil
}

// IndustryIDFromUUID converts a uuid.UUID directly (for DB results).
func IndustryIDFromUUID(id uuid.UUID) IndustryID { return IndustryID(id) }

// String returns the string representation.
func (i IndustryID) String() string { return uuid.UUID(i).String() }

// UUID returns the underlying uuid.UUID.
func (i IndustryID) UUID() uuid.UUID { return uuid.UUID(i) }

// --- IndustrySlug ---

// slugRegex validates slug format: lowercase alphanumeric with hyphens,
// must start and end with alphanumeric. Minimum 2 chars required by regex.
var slugRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)

// IndustrySlug is a URL-safe identifier (2-50 chars, lowercase + hyphens + digits).
type IndustrySlug string

// NewIndustrySlug validates and creates an IndustrySlug.
func NewIndustrySlug(slug string) (IndustrySlug, error) {
	if len(slug) < 2 {
		return "", apperror.NewValidationError(fmt.Errorf("slug must be at least 2 characters"), "IndustrySlug")
	}
	if len(slug) > 50 {
		return "", apperror.NewValidationError(fmt.Errorf("slug must be at most 50 characters"), "IndustrySlug")
	}
	if !slugRegex.MatchString(slug) {
		return "", apperror.NewValidationError(fmt.Errorf("slug must contain only lowercase letters, numbers, and hyphens, and must not start or end with a hyphen"), "IndustrySlug")
	}
	return IndustrySlug(slug), nil
}

// String returns the slug as a plain string.
func (s IndustrySlug) String() string { return string(s) }

// --- IndustryName ---

// IndustryName is a validated name (1-100 characters, non-blank).
type IndustryName string

// NewIndustryName validates and creates an IndustryName.
func NewIndustryName(name string) (IndustryName, error) {
	if err := valobj.ValidateName(name, 1, 100, "IndustryName"); err != nil {
		return "", err
	}
	return IndustryName(name), nil
}

// String returns the name as a plain string.
func (n IndustryName) String() string { return string(n) }

// --- IndustryDescription ---

// IndustryDescription is an optional description (0-1000 characters).
// A zero-value (empty string) is valid.
type IndustryDescription string

// NewIndustryDescription validates and creates an IndustryDescription.
func NewIndustryDescription(desc string) (IndustryDescription, error) {
	if err := valobj.ValidateMaxLength(desc, 1000, "IndustryDescription"); err != nil {
		return "", err
	}
	return IndustryDescription(desc), nil
}

// String returns the description as a plain string.
func (d IndustryDescription) String() string { return string(d) }
