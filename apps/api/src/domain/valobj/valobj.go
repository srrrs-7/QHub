// Package valobj provides shared validation helpers for domain value objects.
//
// Domain entities (task, organization, project, prompt) define type-safe
// value objects whose constructors delegate to these helpers for common
// validation rules such as UUID parsing, name length, slug format, and
// max-length constraints. This eliminates duplicate validation logic
// while keeping each domain type independently typed.
package valobj

import (
	"fmt"
	"regexp"
	"strings"

	"api/src/domain/apperror"

	"github.com/google/uuid"
)

// SlugRegex validates slug format: lowercase alphanumeric with hyphens,
// must not start or end with a hyphen. Example: "my-project-01".
var SlugRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)

// ParseUUID parses a UUID string and returns the parsed UUID.
// Returns a ValidationError if the string is not a valid UUID.
func ParseUUID(id string, domainName string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, apperror.NewValidationError(fmt.Errorf("invalid %s: %w", domainName, err), domainName)
	}
	return parsed, nil
}

// ValidateName checks that name is non-empty (after trimming whitespace)
// and within [minLen, maxLen] characters. Returns a ValidationError on failure.
func ValidateName(name string, minLen, maxLen int, domainName string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return apperror.NewValidationError(fmt.Errorf("%s must not be empty", domainName), domainName)
	}
	if len(name) < minLen {
		return apperror.NewValidationError(fmt.Errorf("%s must be at least %d characters", domainName, minLen), domainName)
	}
	if len(name) > maxLen {
		return apperror.NewValidationError(fmt.Errorf("%s must be at most %d characters", domainName, maxLen), domainName)
	}
	return nil
}

// ValidateSlug checks that slug is within [minLen, maxLen] characters
// and matches the SlugRegex pattern (lowercase alphanumeric with hyphens).
// Returns a ValidationError on failure.
func ValidateSlug(slug string, minLen, maxLen int, domainName string) error {
	if len(slug) < minLen {
		return apperror.NewValidationError(fmt.Errorf("slug must be at least %d characters", minLen), domainName)
	}
	if len(slug) > maxLen {
		return apperror.NewValidationError(fmt.Errorf("slug must be at most %d characters", maxLen), domainName)
	}
	if !SlugRegex.MatchString(slug) {
		return apperror.NewValidationError(fmt.Errorf("slug must contain only lowercase letters, numbers, and hyphens, and must not start or end with a hyphen"), domainName)
	}
	return nil
}

// ValidateMaxLength checks that s does not exceed maxLen characters.
// An empty string is considered valid. Returns a ValidationError on failure.
func ValidateMaxLength(s string, maxLen int, domainName string) error {
	if len(s) > maxLen {
		return apperror.NewValidationError(fmt.Errorf("%s must be at most %d characters", domainName, maxLen), domainName)
	}
	return nil
}
