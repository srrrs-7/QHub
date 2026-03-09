package user

import "context"

// UserRepository defines persistence operations for users.
// Implementations live in the infra layer (dependency inversion).
type UserRepository interface {
	FindByID(ctx context.Context, id UserID) (User, error)
	FindByEmail(ctx context.Context, email UserEmail) (User, error)
	Create(ctx context.Context, email UserEmail, name UserName) (User, error)
	Update(ctx context.Context, id UserID, email *UserEmail, name *UserName) (User, error)
}
