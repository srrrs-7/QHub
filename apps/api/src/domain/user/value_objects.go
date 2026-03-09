// Package user defines the User entity and its value objects.
//
// A User represents an authenticated person in the system. Users can be
// members of organizations and interact with projects and prompts.
package user

import (
	"fmt"
	"net/mail"
	"strings"

	"api/src/domain/apperror"
	"api/src/domain/valobj"

	"github.com/google/uuid"
)

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

// --- UserEmail ---

// UserEmail is a validated email address (max 255 characters, valid email format).
type UserEmail string

// NewUserEmail validates and creates a UserEmail.
func NewUserEmail(email string) (UserEmail, error) {
	trimmed := strings.TrimSpace(email)
	if trimmed == "" {
		return "", apperror.NewValidationError(fmt.Errorf("UserEmail must not be empty"), "UserEmail")
	}
	if len(trimmed) > 255 {
		return "", apperror.NewValidationError(fmt.Errorf("UserEmail must be at most 255 characters"), "UserEmail")
	}
	if _, err := mail.ParseAddress(trimmed); err != nil {
		return "", apperror.NewValidationError(fmt.Errorf("UserEmail is not a valid email address: %w", err), "UserEmail")
	}
	return UserEmail(trimmed), nil
}

// String returns the email as a plain string.
func (e UserEmail) String() string { return string(e) }

// --- UserName ---

// UserName is a validated user name (1–100 characters, non-blank).
type UserName string

// NewUserName validates and creates a UserName.
func NewUserName(name string) (UserName, error) {
	if err := valobj.ValidateName(name, 1, 100, "UserName"); err != nil {
		return "", err
	}
	return UserName(name), nil
}

// String returns the name as a plain string.
func (n UserName) String() string { return string(n) }
