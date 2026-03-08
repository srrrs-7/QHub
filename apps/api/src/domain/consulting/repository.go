package consulting

import (
	"context"

	"github.com/google/uuid"
)

// SessionRepository defines persistence operations for consulting sessions.
type SessionRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (Session, error)
	FindAllByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]Session, error)
	Create(ctx context.Context, session Session) (Session, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) (Session, error)
}

// MessageRepository defines persistence operations for session messages.
type MessageRepository interface {
	FindAllBySession(ctx context.Context, sessionID uuid.UUID) ([]Message, error)
	Create(ctx context.Context, msg Message) (Message, error)
}

// IndustryConfigRepository defines persistence operations for industry configs.
type IndustryConfigRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (IndustryConfig, error)
	FindBySlug(ctx context.Context, slug string) (IndustryConfig, error)
	FindAll(ctx context.Context) ([]IndustryConfig, error)
	Create(ctx context.Context, cfg IndustryConfig) (IndustryConfig, error)
	Update(ctx context.Context, cfg IndustryConfig) (IndustryConfig, error)
}
