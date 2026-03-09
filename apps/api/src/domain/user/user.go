package user

import (
	"time"

	"github.com/google/uuid"
)

// User is the aggregate root representing a person in the system.
type User struct {
	id        UserID
	email     UserEmail
	name      UserName
	createdAt time.Time
	updatedAt time.Time
}

// NewUser constructs a User from validated value objects.
func NewUser(id UserID, email UserEmail, name UserName, createdAt, updatedAt time.Time) User {
	return User{
		id:        id,
		email:     email,
		name:      name,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

// NewUserFromDB constructs a User from raw DB values (trusted source).
func NewUserFromDB(id uuid.UUID, email string, name string, createdAt, updatedAt time.Time) User {
	return User{
		id:        UserIDFromUUID(id),
		email:     UserEmail(email),
		name:      UserName(name),
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

// ID returns the user's unique identifier.
func (u User) ID() UserID { return u.id }

// Email returns the user's email address.
func (u User) Email() UserEmail { return u.email }

// Name returns the user's display name.
func (u User) Name() UserName { return u.name }

// CreatedAt returns the time the user was created.
func (u User) CreatedAt() time.Time { return u.createdAt }

// UpdatedAt returns the time the user was last updated.
func (u User) UpdatedAt() time.Time { return u.updatedAt }
